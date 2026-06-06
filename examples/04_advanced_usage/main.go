package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	fmt.Println("=== Advanced Usage Example ===")
	fmt.Println("This example shows advanced MCP server usage patterns.")
	fmt.Println()

	// Load swagger from our local server
	data, err := readSwaggerFile("../server/swagger.json")
	if err != nil {
		log.Fatalf("Failed to read swagger file: %v", err)
	}

	// Example 1: Custom server with detailed configuration
	fmt.Println("🔧 Example 1: Custom Server Configuration")
	fmt.Println("-----------------------------------------")
	config := mcp.DefaultConfig().
		WithServerInfo("advanced-petstore", "2.0.0", "Advanced Petstore MCP Server with custom settings").
		WithAPIConfig("http://localhost:4538/v2", "").
		WithSwaggerData(data).
		WithExcludeMethods("DELETE"). // No destructive operations
		WithHTTPTransport(8127, "localhost", "/mcp")

	server, err := mcp.New(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("✅ Advanced server created\n")
	fmt.Printf("   Server: %s v%s\n", server.GetConfig().Name, server.GetConfig().Version)
	fmt.Printf("   Transport: %T\n", server.GetConfig().Transport)
	fmt.Printf("   Tools configured from swagger specification\n")

	// Start the server
	ctx := context.Background()
	go func() {
		if err := server.Run(ctx); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// Give server time to start
	time.Sleep(2 * time.Second)

	// Example 2: Testing MCP tools with a standard MCP client
	fmt.Println("\n🧪 Example 2: Testing MCP Tools")
	fmt.Println("-------------------------------")
	testMCPTools()

	// Example 3: Demonstrate tool information
	fmt.Println("\n📋 Example 3: Tool Information")
	fmt.Println("-----------------------------")
	fmt.Println("Tools are dynamically generated from swagger specification:")
	fmt.Println("   • listpets - List all pets with optional filtering")
	fmt.Println("   • createpet - Create a new pet")
	fmt.Println("   • getpetbyid - Get a specific pet by ID")
	fmt.Println("   • updatepet - Update an existing pet (excluded: PUT allowed, DELETE filtered)")
	fmt.Println("   • searchpets - Search pets by criteria")
	fmt.Println()
	fmt.Println("Each tool's input schema is automatically generated from swagger parameters")

	fmt.Println("\n🚀 Advanced Features Demonstrated:")
	fmt.Println("   • Custom server configuration")
	fmt.Println("   • HTTP transport setup (standard MCP Streamable HTTP)")
	fmt.Println("   • API filtering (DELETE operations excluded)")
	fmt.Println("   • Error handling patterns")
	fmt.Println()
	fmt.Println("⚠️  Make sure the local API server is running:")
	fmt.Println("   cd ../server && ./start_server.sh")
	fmt.Println("\n🔧 Server is running on http://localhost:8127/mcp")
	fmt.Println("Press Ctrl+C to exit.")

	select {}
}

func readSwaggerFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func testMCPTools() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect with the official MCP client over Streamable HTTP
	client := sdk.NewClient(&sdk.Implementation{Name: "advanced-example-client", Version: "1.0"}, nil)
	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{
		Endpoint: "http://localhost:8127/mcp",
	}, nil)
	if err != nil {
		fmt.Printf("   ❌ Failed to connect: %v\n", err)
		return
	}
	defer func() { _ = session.Close() }()

	// Test tools list
	fmt.Println("   Testing tools/list...")
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		fmt.Printf("   ❌ tools/list failed: %v\n", err)
		return
	}
	fmt.Printf("   ✅ tools/list successful (%d tools, DELETE operations excluded)\n", len(tools.Tools))

	// Test listpets tool
	fmt.Println("   Testing listpets tool...")
	result, err := session.CallTool(ctx, &sdk.CallToolParams{
		Name:      "listpets",
		Arguments: map[string]any{"limit": 3},
	})
	if err != nil || result.IsError {
		fmt.Printf("   ❌ listpets failed: err=%v isError=%v\n", err, result != nil && result.IsError)
	} else {
		fmt.Println("   ✅ listpets tool called successfully")
	}

	// Test getpetbyid tool
	fmt.Println("   Testing getpetbyid tool...")
	result, err = session.CallTool(ctx, &sdk.CallToolParams{
		Name:      "getpetbyid",
		Arguments: map[string]any{"petId": 1},
	})
	if err != nil || result.IsError {
		fmt.Printf("   ❌ getpetbyid failed: err=%v isError=%v\n", err, result != nil && result.IsError)
	} else {
		fmt.Println("   ✅ getpetbyid tool called successfully")
	}

	// Test searchpets tool
	fmt.Println("   Testing searchpets tool...")
	result, err = session.CallTool(ctx, &sdk.CallToolParams{
		Name:      "searchpets",
		Arguments: map[string]any{"body": map[string]any{"name": "Bud", "tag": "dog"}},
	})
	if err != nil || result.IsError {
		fmt.Printf("   ❌ searchpets failed: err=%v isError=%v\n", err, result != nil && result.IsError)
	} else {
		fmt.Println("   ✅ searchpets tool called successfully")
	}

	// Error handling: invalid tool name
	fmt.Println("   Testing invalid tool name...")
	_, err = session.CallTool(ctx, &sdk.CallToolParams{Name: "invalidTool"})
	if err != nil {
		fmt.Printf("   ✅ Invalid tool correctly rejected: %v\n", err)
	} else {
		fmt.Println("   ❌ Invalid tool should have been rejected")
	}

	fmt.Println("   ✅ All tool tests completed")
}
