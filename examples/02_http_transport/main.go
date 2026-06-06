package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	fmt.Println("=== HTTP Transport Example ===")
	fmt.Println("This example shows how to use MCP server with HTTP transport.")
	fmt.Println("The /mcp endpoint speaks the standard MCP Streamable HTTP protocol,")
	fmt.Println("so any standard MCP client can connect to it.")
	fmt.Println()

	// Create server configuration
	config := mcp.DefaultConfig().
		WithServerInfo("local-petstore-http", "1.0.0", "Local Petstore API HTTP Server").
		WithAPIConfig("http://localhost:4538/v2", "")

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
	fmt.Printf("   • POST http://localhost:%d/mcp        - MCP Streamable HTTP endpoint\n", port)
	fmt.Printf("   • GET  http://localhost:%d/mcp/health - Health check\n", port)
	fmt.Printf("   • GET  http://localhost:%d/mcp/tools  - List available tools (REST)\n", port)

	fmt.Println("\n🔧 Usage Examples:")
	fmt.Printf("   1. Health check:\n")
	fmt.Printf("      curl http://localhost:%d/mcp/health\n", port)
	fmt.Println()
	fmt.Printf("   2. List tools (REST convenience endpoint):\n")
	fmt.Printf("      curl http://localhost:%d/mcp/tools\n", port)
	fmt.Println()
	fmt.Printf("   3. Connect with a standard MCP client:\n")
	fmt.Printf("      claude mcp add petstore --transport http http://localhost:%d/mcp\n", port)

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
	fmt.Printf("   1. Testing GET %s/mcp/health... ", baseURL)
	resp, err := http.Get(baseURL + "/mcp/health")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == 200 {
		fmt.Println("✅ OK")
	} else {
		fmt.Printf("❌ Status: %d\n", resp.StatusCode)
	}

	// Test tools endpoint
	fmt.Printf("   2. Testing GET %s/mcp/tools... ", baseURL)
	resp, err = http.Get(baseURL + "/mcp/tools")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == 200 {
		fmt.Println("✅ OK")
	} else {
		fmt.Printf("❌ Status: %d\n", resp.StatusCode)
	}

	// Test the MCP endpoint with a standard MCP client over Streamable HTTP
	fmt.Printf("   3. Testing MCP Streamable HTTP at %s/mcp... ", baseURL)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client := sdk.NewClient(&sdk.Implementation{Name: "example-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: baseURL + "/mcp"}, nil)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	defer func() { _ = session.Close() }()

	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		fmt.Printf("❌ tools/list error: %v\n", err)
		return
	}
	fmt.Printf("✅ OK (%d tools)\n", len(tools.Tools))

	fmt.Println("\n✅ All HTTP endpoints are working correctly!")
}
