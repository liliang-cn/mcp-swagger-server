package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	fmt.Println("=== Basic Stdio Transport Example ===")
	fmt.Println("This example shows how to use MCP server with stdio transport.")
	fmt.Println("Stdio transport is ideal for CLI tools and MCP client integration.")
	fmt.Println()

	// Create server configuration
	config := mcp.DefaultConfig().
		WithServerInfo("local-petstore-client", "1.0.0", "Local Petstore API MCP Client").
		WithAPIConfig("http://localhost:4538", "")

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

	fmt.Printf("✅ MCP Server created successfully!\n")
	fmt.Printf("   Server Name: %s\n", server.GetConfig().Name)
	fmt.Printf("   Version: %s\n", server.GetConfig().Version)
	fmt.Printf("   API Base URL: %s\n", server.GetConfig().APIBaseURL)

	fmt.Println()
	fmt.Println("📋 Available Tools:")
	tools := server.ListTools()
	if len(tools) > 0 {
		for i, tool := range tools {
			fmt.Printf("   %d. %s\n", i+1, tool)
		}
	} else {
		fmt.Printf("   Tools will be available when the server connects to an MCP client\n")
		fmt.Printf("   Expected tools based on swagger: listPets, createPet, getPetById, updatePet, searchPets\n")
	}

	fmt.Println()
	fmt.Println("🚀 Usage:")
	fmt.Println("   1. Start the local API server:")
	fmt.Println("      cd ../server && ./start_server.sh")
	fmt.Println()
	fmt.Println("   2. Run this MCP server:")
	fmt.Println("      go run main.go")
	fmt.Println()
	fmt.Println("   3. Use with MCP client (like Claude Desktop):")
	fmt.Println(`      {
        "servers": {
          "local-petstore": {
            "command": "go",
            "args": ["run", "examples/01_basic_stdio/main.go"],
            "cwd": "."
          }
        }
      }`)

	fmt.Println()
	fmt.Println("⚠️  Note: The server will block waiting for MCP protocol input on stdin.")
	fmt.Println("   Use this with MCP clients that support stdio communication.")

	// Run the server with stdio transport
	ctx := context.Background()
	if err := server.RunStdio(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func readSwaggerFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}