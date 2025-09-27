package main

import (
	"context"
	"log"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	// Example 1: Create MCP server from a local Swagger file
	server1, err := mcp.NewFromSwaggerFile("../petstore.json", "", "")
	if err != nil {
		log.Printf("Failed to create server from file: %v", err)
	} else {
		log.Println("Created server from local file successfully")
		_ = server1 // Use the server
	}

	// Example 2: Create MCP server from a Swagger URL
	server2, err := mcp.NewFromSwaggerURL("https://petstore.swagger.io/v2/swagger.json", "https://petstore.swagger.io/v2", "")
	if err != nil {
		log.Printf("Failed to create server from URL: %v", err)
	} else {
		log.Println("Created server from URL successfully")
		
		// Run with stdio transport (for CLI usage)
		ctx := context.Background()
		log.Println("Starting MCP server with stdio transport...")
		if err := server2.RunStdio(ctx); err != nil {
			log.Printf("Server error: %v", err)
		}
	}

	// Example 3: Create server with custom configuration
	config := mcp.DefaultConfig().
		WithAPIConfig("https://api.example.com", "your-api-key").
		WithServerInfo("my-api-server", "v1.0.0", "Custom API MCP Server")

	// You would set swagger data here
	// config.WithSwaggerData(swaggerData)

	server3, err := mcp.New(config)
	if err != nil {
		log.Printf("Failed to create custom server: %v", err)
	} else {
		log.Println("Created custom server successfully")
		_ = server3 // Use the server
	}
}