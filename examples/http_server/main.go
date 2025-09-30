package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	var (
		swaggerFile = flag.String("swagger", "../../test/petstore.json", "Path to Swagger/OpenAPI spec file")
		apiBaseURL  = flag.String("api-base", "https://petstore.swagger.io/v2", "Base URL for API calls")
		httpPort    = flag.Int("port", 3218, "HTTP server port")
		httpHost    = flag.String("host", "localhost", "HTTP server host")
		demo        = flag.Bool("demo", false, "Run demo mode with test calls")
	)

	flag.Parse()

	// Read swagger file
	data, err := os.ReadFile(*swaggerFile)
	if err != nil {
		log.Fatalf("Failed to read swagger file: %v", err)
	}

	// Create server configuration with HTTP transport
	config := mcp.DefaultConfig().
		WithSwaggerData(data).
		WithAPIConfig(*apiBaseURL, "").
		WithHTTPTransport(*httpPort, *httpHost, "/mcp")

	// Create server
	server, err := mcp.New(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\nShutting down...")
		cancel()
	}()

	// Start server in background
	serverErrChan := make(chan error, 1)
	go func() {
		serverErrChan <- server.RunHTTP(ctx, *httpPort)
	}()

	// Wait for server to start
	time.Sleep(1 * time.Second)

	// Print server info
	fmt.Println("=== MCP Swagger Server (HTTP Mode) ===")
	fmt.Printf("Server running on: http://%s:%d\n", *httpHost, *httpPort)
	fmt.Println("\nAvailable endpoints:")
	fmt.Printf("  - Health check: http://%s:%d/health\n", *httpHost, *httpPort)
	fmt.Printf("  - Tools list:   http://%s:%d/tools\n", *httpHost, *httpPort)
	fmt.Printf("  - MCP endpoint: http://%s:%d/mcp\n", *httpHost, *httpPort)
	fmt.Println("\nPress Ctrl+C to stop the server")

	// If demo mode, run test calls
	if *demo {
		fmt.Println("\n=== Running Demo Mode ===")
		runDemoTests(*httpHost, *httpPort)
	}

	// Wait for server error or shutdown
	select {
	case err := <-serverErrChan:
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	case <-ctx.Done():
		fmt.Println("Server stopped")
	}
}

// runDemoTests runs demo test calls
func runDemoTests(host string, port int) {
	baseURL := fmt.Sprintf("http://%s:%d", host, port)
	
	// Test health endpoint
	fmt.Println("\n1. Testing health endpoint...")
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		log.Printf("Health check failed: %v", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("   Response: %s\n", string(body))
	
	// Test tools list
	fmt.Println("\n2. Testing tools list endpoint...")
	resp, err = http.Get(baseURL + "/tools")
	if err != nil {
		log.Printf("Tools list failed: %v", err)
		return
	}
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	
	var toolsResponse map[string]interface{}
	if err := json.Unmarshal(body, &toolsResponse); err == nil {
		if tools, ok := toolsResponse["tools"].([]interface{}); ok {
			fmt.Printf("   Found %d tools:\n", len(tools))
			for i, tool := range tools {
				if toolMap, ok := tool.(map[string]interface{}); ok {
					fmt.Printf("   %d. %s - %s\n", i+1, toolMap["name"], toolMap["description"])
				}
			}
		}
	}
	
	// Test MCP call
	fmt.Println("\n3. Testing MCP tool call...")
	mcpRequest := map[string]interface{}{
		"method": "tools/call",
		"params": map[string]interface{}{
			"name": "getpet",
			"arguments": map[string]interface{}{
				"petId": 1,
			},
		},
	}
	
	jsonData, _ := json.Marshal(mcpRequest)
	resp, err = http.Post(baseURL+"/mcp", "application/json", bytes.NewReader(jsonData))
	if err != nil {
		log.Printf("MCP call failed: %v", err)
		return
	}
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {
		fmt.Printf("   Response status: %d\n", resp.StatusCode)
		if resp.StatusCode == 200 {
			fmt.Println("   Tool call successful!")
		} else {
			fmt.Printf("   Error: %v\n", result)
		}
	}
	
	fmt.Println("\nDemo complete! Server continues running...")
}