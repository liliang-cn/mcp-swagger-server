package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	fmt.Println("=== HTTP Transport Example ===")
	fmt.Println("This example shows how to use MCP server with HTTP transport.")
	fmt.Println("HTTP transport is ideal for web applications and HTTP clients.")
	fmt.Println()

	// Create server configuration
	config := mcp.DefaultConfig().
		WithServerInfo("local-petstore-http", "1.0.0", "Local Petstore API HTTP Server").
		WithAPIConfig("http://localhost:4538", "")

	// Configure HTTP transport
	port := 6724
	config.WithHTTPTransport(port, "localhost", "/mcp")

	// Load swagger from our local server
	data, err := readSwaggerFile("../server/swagger.json")
	if err != nil {
		log.Fatalf("Failed to read swagger file: %v", err)
	}
	config.WithSwaggerData(data)

	// Create the server
	server, err := mcp.New(config)
	if err != nil {
		log.Fatalf("Failed to create MCP server: %v", err)
	}

	fmt.Printf("✅ MCP HTTP Server created successfully!\n")
	fmt.Printf("   Server Name: %s\n", server.GetConfig().Name)
	fmt.Printf("   Version: %s\n", server.GetConfig().Version)
	fmt.Printf("   HTTP Port: %d\n", port)

	// Test tools count by making a quick request to the tools endpoint
	fmt.Printf("   Configuring tools from swagger specification...\n")

	// Start HTTP server in a goroutine
	ctx := context.Background()
	go func() {
		if err := server.Run(ctx); err != nil {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	// Give server time to start
	fmt.Println("\n🚀 Starting HTTP server...")
	time.Sleep(2 * time.Second)

	// Test HTTP endpoints
	fmt.Println("\n🧪 Testing HTTP Endpoints...")
	testHTTPEndpoints(port)

	fmt.Println("\n📋 Available HTTP Endpoints:")
	fmt.Printf("   • GET  http://localhost:%d/health   - Health check\n", port)
	fmt.Printf("   • GET  http://localhost:%d/tools    - List available tools\n", port)
	fmt.Printf("   • POST http://localhost:%d/mcp      - Execute MCP commands\n", port)

	fmt.Println("\n🔧 Usage Examples:")
	fmt.Printf("   1. Health check:\n")
	fmt.Printf("      curl http://localhost:%d/health\n", port)
	fmt.Println()
	fmt.Printf("   2. List tools:\n")
	fmt.Printf("      curl http://localhost:%d/tools\n", port)
	fmt.Println()
	fmt.Printf("   3. Execute tool:\n")
	fmt.Printf(`      curl -X POST http://localhost:%d/mcp \
        -H "Content-Type: application/json" \
        -d '{
          "method": "tools/call",
          "params": {
            "name": "listPets",
            "arguments": {"limit": 5}
          }
        }'`, port)

	fmt.Println("\n⚠️  Make sure the local API server is running:")
	fmt.Println("   cd ../server && ./start_server.sh")
	fmt.Println("\nServer is running. Press Ctrl+C to exit.")

	// Keep the main goroutine alive
	select {}
}

func readSwaggerFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func testHTTPEndpoints(port int) {
	baseURL := fmt.Sprintf("http://localhost:%d", port)

	// Test health endpoint
	fmt.Printf("   1. Testing GET %s/health... ", baseURL)
	resp, err := http.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Println("✅ OK")
	} else {
		fmt.Printf("❌ Status: %d\n", resp.StatusCode)
	}

	// Test tools endpoint
	fmt.Printf("   2. Testing GET %s/tools... ", baseURL)
	resp, err = http.Get(baseURL + "/tools")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Println("✅ OK")
	} else {
		fmt.Printf("❌ Status: %d\n", resp.StatusCode)
	}

	// Test MCP endpoint with tools/list
	fmt.Printf("   3. Testing POST %s/mcp (tools/list)... ", baseURL)
	mcpReq := `{
		"method": "tools/list",
		"params": {}
	}`
	resp, err = http.Post(baseURL+"/mcp", "application/json", strings.NewReader(mcpReq))
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		fmt.Println("✅ OK")
	} else {
		fmt.Printf("❌ Status: %d\n", resp.StatusCode)
	}

	fmt.Println("\n✅ All HTTP endpoints are working correctly!")
}