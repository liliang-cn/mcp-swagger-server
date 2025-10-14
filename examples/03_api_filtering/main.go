package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	fmt.Println("=== API Filtering Example ===")
	fmt.Println("This example demonstrates how to filter which API operations are exposed as MCP tools.")
	fmt.Println()

	// Load swagger from our local server
	data, err := readSwaggerFile("../server/swagger.json")
	if err != nil {
		log.Fatalf("Failed to read swagger file: %v", err)
	}

	// Example 1: All operations (default)
	fmt.Println("🔍 Example 1: All Operations")
	fmt.Println("----------------------------")
	config1 := mcp.DefaultConfig().
		WithServerInfo("all-operations", "1.0.0", "All API operations available").
		WithAPIConfig("http://localhost:4538", "").
		WithSwaggerData(data)

	_, err = mcp.New(config1)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("All operations will be converted to MCP tools\n")
	fmt.Printf("Expected tools: listPets, createPet, getPetById, updatePet, deletePet, searchPets\n")

	// Example 2: Read-only operations
	fmt.Println("\n🔍 Example 2: Read-Only Operations")
	fmt.Println("---------------------------------")
	config2 := mcp.DefaultConfig().
		WithServerInfo("read-only", "1.0.0", "Read-only API operations").
		WithAPIConfig("http://localhost:4538", "").
		WithSwaggerData(data).
		WithExcludeMethods("POST", "PUT", "DELETE") // Exclude write operations

	server2, err := mcp.New(config2)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Read-only tools will be available\n")
	fmt.Printf("Expected tools: listPets, getPetById, searchPets\n")

	// Example 3: Exclude search operations
	fmt.Println("\n🔍 Example 3: Basic Operations Only")
	fmt.Println("----------------------------------")
	config3 := mcp.DefaultConfig().
		WithServerInfo("basic-only", "1.0.0", "Basic CRUD operations only").
		WithAPIConfig("http://localhost:4538", "").
		WithSwaggerData(data).
		WithExcludeTags("search")

	_, err = mcp.New(config3)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Basic CRUD tools will be available\n")
	fmt.Printf("Expected tools: listPets, createPet, getPetById, updatePet\n")

	// Example 4: Include specific tags only (using exclude all but search)
	fmt.Println("\n🔍 Example 4: Search Operations Only")
	fmt.Println("------------------------------------")
	config4 := mcp.DefaultConfig().
		WithServerInfo("search-only", "1.0.0", "Search operations only").
		WithAPIConfig("http://localhost:4538", "").
		WithSwaggerData(data)

	// Filter to include only search operations by excluding common CRUD methods
	filter := &mcp.APIFilter{
		ExcludeMethods: []string{"POST", "PUT", "DELETE", "GET"},
		IncludeOnlyOperationIDs: []string{"searchPets"},
	}
	config4.WithAPIFilter(filter)

	_, err = mcp.New(config4)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Search-only tools will be available\n")
	fmt.Printf("Expected tools: searchPets\n")

	// Example 5: Combined filtering
	fmt.Println("\n🔍 Example 5: Safe Read-Only with Basic Operations")
	fmt.Println("-------------------------------------------------")
	config5 := mcp.DefaultConfig().
		WithServerInfo("safe-basic", "1.0.0", "Safe read-only basic operations").
		WithAPIConfig("http://localhost:4538", "").
		WithSwaggerData(data).
		WithExcludeMethods("POST", "PUT", "DELETE").
		WithExcludeTags("search")

	_, err = mcp.New(config5)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("Safe read-only basic tools will be available\n")
	fmt.Printf("Expected tools: listPets, getPetById\n")

	fmt.Println("\n🚀 Usage:")
	fmt.Println("   1. Start the local API server:")
	fmt.Println("      cd ../server && ./start_server.sh")
	fmt.Println()
	fmt.Println("   2. Run this example:")
	fmt.Println("      go run main.go")
	fmt.Println()
	fmt.Println("   3. Choose the filtering strategy that fits your needs:")
	fmt.Println("      - Read-only for clients that should only view data")
	fmt.Println("      - Basic operations for simple CRUD workflows")
	fmt.Println("      - Search-only for specialized search applications")
	fmt.Println("      - Custom combinations for specific use cases")

	ctx := context.Background()

	// Start one of the filtered servers (example: read-only)
	fmt.Println("\n🔧 Starting read-only server with HTTP transport...")
	config2.WithHTTPTransport(7778, "localhost", "/mcp")
	go func() {
		if err := server2.Run(ctx); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	fmt.Println("   Server running on http://localhost:7778")
	fmt.Println("   Available tools: listPets, getPetById")
	fmt.Println("\nPress Ctrl+C to exit.")

	select {}
}

func readSwaggerFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}