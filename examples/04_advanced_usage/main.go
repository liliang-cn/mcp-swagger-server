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
		WithAPIConfig("http://localhost:4538", "").
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

	// Example 2: Testing MCP tools via HTTP
	fmt.Println("\n🧪 Example 2: Testing MCP Tools")
	fmt.Println("-------------------------------")
	testMCPTools()

	// Example 3: Demonstrate tool information
	fmt.Println("\n📋 Example 3: Tool Information")
	fmt.Println("-----------------------------")
	fmt.Println("Tools are dynamically generated from swagger specification:")
	fmt.Println("   • listPets - List all pets with optional filtering")
	fmt.Println("   • createPet - Create a new pet")
	fmt.Println("   • getPetById - Get a specific pet by ID")
	fmt.Println("   • updatePet - Update an existing pet")
	fmt.Println("   • deletePet - Delete a pet")
	fmt.Println("   • searchPets - Search pets by criteria")
	fmt.Println()
	fmt.Println("Each tool's input schema is automatically generated from swagger parameters")

	// Example 4: Error handling demonstration
	fmt.Println("⚠️  Example 4: Error Handling")
	fmt.Println("-----------------------------")
	demonstrateErrorHandling()

	fmt.Println("\n🚀 Advanced Features Demonstrated:")
	fmt.Println("   • Custom server configuration")
	fmt.Println("   • HTTP transport setup")
	fmt.Println("   • Tool schema inspection")
	fmt.Println("   • Error handling patterns")
	fmt.Println("   • MCP protocol via HTTP")
	fmt.Println()
	fmt.Println("⚠️  Make sure the local API server is running:")
	fmt.Println("   cd ../server && ./start_server.sh")
	fmt.Println("\n🔧 Server is running on http://localhost:8127")
	fmt.Println("Press Ctrl+C to exit.")

	select {}
}

func readSwaggerFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func testMCPTools() {
	baseURL := "http://localhost:8127"

	// Test tools list
	fmt.Println("   Testing tools/list...")
	resp, err := http.Post(baseURL+"/mcp", "application/json",
		strings.NewReader(`{"method": "tools/list", "params": {}}`))
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("   ✅ tools/list successful")
	} else {
		fmt.Printf("   ❌ tools/list failed: %d\n", resp.StatusCode)
	}

	// Test listPets tool
	fmt.Println("   Testing listPets tool...")
	listPetsReq := `{
		"method": "tools/call",
		"params": {
			"name": "listPets",
			"arguments": {"limit": 3}
		}
	}`

	resp, err = http.Post(baseURL+"/mcp", "application/json",
		strings.NewReader(listPetsReq))
	if err != nil {
		fmt.Printf("   ❌ Error calling listPets: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("   ✅ listPets tool called successfully")
	} else {
		fmt.Printf("   ❌ listPets tool failed: %d\n", resp.StatusCode)
	}

	// Test getPetById tool
	fmt.Println("   Testing getPetById tool...")
	getPetReq := `{
		"method": "tools/call",
		"params": {
			"name": "getPetById",
			"arguments": {"petId": 1}
		}
	}`

	resp, err = http.Post(baseURL+"/mcp", "application/json",
		strings.NewReader(getPetReq))
	if err != nil {
		fmt.Printf("   ❌ Error calling getPetById: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("   ✅ getPetById tool called successfully")
	} else {
		fmt.Printf("   ❌ getPetById tool failed: %d\n", resp.StatusCode)
	}

	// Test searchPets tool
	fmt.Println("   Testing searchPets tool...")
	searchReq := `{
		"method": "tools/call",
		"params": {
			"name": "searchPets",
			"arguments": {"name": "Bud", "tag": "dog"}
		}
	}`

	resp, err = http.Post(baseURL+"/mcp", "application/json",
		strings.NewReader(searchReq))
	if err != nil {
		fmt.Printf("   ❌ Error calling searchPets: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		fmt.Println("   ✅ searchPets tool called successfully")
	} else {
		fmt.Printf("   ❌ searchPets tool failed: %d\n", resp.StatusCode)
	}

	fmt.Println("   ✅ All tool tests completed")
}

func demonstrateErrorHandling() {
	baseURL := "http://localhost:8127"

	// Test invalid tool name
	fmt.Println("   Testing invalid tool name...")
	invalidReq := `{
		"method": "tools/call",
		"params": {
			"name": "invalidTool",
			"arguments": {}
		}
	}`

	resp, err := http.Post(baseURL+"/mcp", "application/json",
		strings.NewReader(invalidReq))
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("   ✅ Invalid tool correctly rejected: %d\n", resp.StatusCode)
	} else {
		fmt.Println("   ❌ Invalid tool should have been rejected")
	}

	// Test missing required arguments
	fmt.Println("   Testing missing required arguments...")
	missingArgsReq := `{
		"method": "tools/call",
		"params": {
			"name": "createPet",
			"arguments": {}
		}
	}`

	resp, err = http.Post(baseURL+"/mcp", "application/json",
		strings.NewReader(missingArgsReq))
	if err != nil {
		fmt.Printf("   ❌ Error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("   ✅ Missing arguments correctly rejected: %d\n", resp.StatusCode)
	} else {
		fmt.Println("   ❌ Missing arguments should have been rejected")
	}

	fmt.Println("   ✅ Error handling tests completed")
}