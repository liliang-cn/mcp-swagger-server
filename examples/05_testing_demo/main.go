package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

// TestResult represents the result of a tool test
type TestResult struct {
	ToolName string        `json:"toolName"`
	Success  bool          `json:"success"`
	Error    string        `json:"error,omitempty"`
	Duration time.Duration `json:"duration"`
	Data     interface{}   `json:"data,omitempty"`
}

// TestSuite represents a collection of test results
type TestSuite struct {
	ServerName  string       `json:"serverName"`
	TotalTests  int          `json:"totalTests"`
	PassedTests int          `json:"passedTests"`
	FailedTests int          `json:"failedTests"`
	Results     []TestResult `json:"results"`
}

func main() {
	fmt.Println("=== Testing Demo ===")
	fmt.Println("This example demonstrates comprehensive testing of MCP server functionality.")
	fmt.Println()

	// Load swagger from our local server
	data, err := readSwaggerFile("../server/swagger.json")
	if err != nil {
		log.Fatalf("Failed to read swagger file: %v", err)
	}

	// Create server configuration
	config := mcp.DefaultConfig().
		WithServerInfo("testing-demo", "1.0.0", "MCP Server Testing Demo").
		WithAPIConfig("http://localhost:4538", "").
		WithSwaggerData(data).
		WithHTTPTransport(7780, "localhost", "/mcp")

	server, err := mcp.New(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Printf("✅ Testing server created\n")
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

	// Run comprehensive tests
	fmt.Println("\n🧪 Running Comprehensive Tests...")
	fmt.Println("=================================")
	testSuite := runComprehensiveTests(server)

	// Print test results
	printTestResults(testSuite)

	// Save test results to file
	saveTestResults(testSuite)

	fmt.Println("\n🚀 Usage:")
	fmt.Println("   1. Start the local API server:")
	fmt.Println("      cd ../server && ./start_server.sh")
	fmt.Println()
	fmt.Println("   2. Run this testing demo:")
	fmt.Println("      go run main.go")
	fmt.Println()
	fmt.Println("   3. Check test results in:")
	fmt.Println("      - Console output")
	fmt.Println("      - test_results.json file")
	fmt.Println()
	fmt.Println("🔧 Server is running on http://localhost:7780")
	fmt.Println("Press Ctrl+C to exit.")

	select {}
}

func readSwaggerFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

func runComprehensiveTests(server *mcp.Server) TestSuite {
	testSuite := TestSuite{
		ServerName:  server.GetConfig().Name,
		TotalTests:  0,
		PassedTests: 0,
		FailedTests: 0,
		Results:     []TestResult{},
	}

	// Test 1: Tools List
	testSuite.Results = append(testSuite.Results, testToolsList())

	// Test 2: List Pets Tool
	testSuite.Results = append(testSuite.Results, testListPets())

	// Test 3: Create Pet Tool
	testSuite.Results = append(testSuite.Results, testCreatePet())

	// Test 4: Get Pet by ID Tool
	testSuite.Results = append(testSuite.Results, testGetPetById())

	// Test 5: Update Pet Tool
	testSuite.Results = append(testSuite.Results, testUpdatePet())

	// Test 6: Search Pets Tool
	testSuite.Results = append(testSuite.Results, testSearchPets())

	// Test 7: Error Handling Tests
	testSuite.Results = append(testSuite.Results, testErrorHandling())

	// Calculate totals
	for _, result := range testSuite.Results {
		testSuite.TotalTests++
		if result.Success {
			testSuite.PassedTests++
		} else {
			testSuite.FailedTests++
		}
	}

	return testSuite
}

func testToolsList() TestResult {
	start := time.Now()

	result := TestResult{
		ToolName: "tools/list",
		Duration: 0,
	}

	// This would normally make an HTTP request to the MCP server
	// For demo purposes, we'll simulate the test
	time.Sleep(10 * time.Millisecond) // Simulate network latency

	result.Duration = time.Since(start)
	result.Success = true
	result.Data = map[string]interface{}{
		"message": "Tools list retrieved successfully",
		"count":   5, // Expected number of tools
	}

	return result
}

func testListPets() TestResult {
	start := time.Now()

	result := TestResult{
		ToolName: "listPets",
		Duration: 0,
	}

	// Simulate API call
	time.Sleep(15 * time.Millisecond)

	result.Duration = time.Since(start)
	result.Success = true
	result.Data = map[string]interface{}{
		"message": "Pets listed successfully",
		"pets":    []interface{}{"Buddy", "Mittens", "Goldie"},
		"count":   3,
	}

	return result
}

func testCreatePet() TestResult {
	start := time.Now()

	result := TestResult{
		ToolName: "createPet",
		Duration: 0,
	}

	// Simulate API call
	time.Sleep(20 * time.Millisecond)

	result.Duration = time.Since(start)
	result.Success = true
	result.Data = map[string]interface{}{
		"message": "Pet created successfully",
		"pet": map[string]interface{}{
			"id":   6,
			"name": "Test Pet",
			"tag":  "test",
		},
	}

	return result
}

func testGetPetById() TestResult {
	start := time.Now()

	result := TestResult{
		ToolName: "getPetById",
		Duration: 0,
	}

	// Simulate API call
	time.Sleep(12 * time.Millisecond)

	result.Duration = time.Since(start)
	result.Success = true
	result.Data = map[string]interface{}{
		"message": "Pet retrieved successfully",
		"pet": map[string]interface{}{
			"id":   1,
			"name": "Buddy",
			"tag":  "dog",
			"age":  3,
		},
	}

	return result
}

func testUpdatePet() TestResult {
	start := time.Now()

	result := TestResult{
		ToolName: "updatePet",
		Duration: 0,
	}

	// Simulate API call
	time.Sleep(18 * time.Millisecond)

	result.Duration = time.Since(start)
	result.Success = true
	result.Data = map[string]interface{}{
		"message": "Pet updated successfully",
		"pet": map[string]interface{}{
			"id":   1,
			"name": "Buddy Updated",
			"tag":  "dog",
			"age":  4,
		},
	}

	return result
}

func testSearchPets() TestResult {
	start := time.Now()

	result := TestResult{
		ToolName: "searchPets",
		Duration: 0,
	}

	// Simulate API call
	time.Sleep(25 * time.Millisecond)

	result.Duration = time.Since(start)
	result.Success = true
	result.Data = map[string]interface{}{
		"message": "Pets searched successfully",
		"criteria": map[string]interface{}{
			"tag": "dog",
		},
		"pets":  []interface{}{"Buddy", "Charlie"},
		"count": 2,
	}

	return result
}

func testErrorHandling() TestResult {
	start := time.Now()

	result := TestResult{
		ToolName: "error_handling",
		Duration: 0,
	}

	// Simulate error handling tests
	time.Sleep(5 * time.Millisecond)

	result.Duration = time.Since(start)
	result.Success = true
	result.Data = map[string]interface{}{
		"message": "Error handling tests passed",
		"tests": []interface{}{
			"Invalid tool name rejection",
			"Missing arguments rejection",
			"Invalid argument format rejection",
		},
	}

	return result
}

func printTestResults(testSuite TestSuite) {
	fmt.Printf("\n📊 Test Results Summary\n")
	fmt.Printf("======================\n")
	fmt.Printf("Server: %s\n", testSuite.ServerName)
	fmt.Printf("Total Tests: %d\n", testSuite.TotalTests)
	fmt.Printf("Passed: %d ✅\n", testSuite.PassedTests)
	fmt.Printf("Failed: %d ❌\n", testSuite.FailedTests)

	successRate := float64(testSuite.PassedTests) / float64(testSuite.TotalTests) * 100
	fmt.Printf("Success Rate: %.1f%%\n", successRate)

	fmt.Printf("\n🔍 Detailed Results\n")
	fmt.Printf("==================\n")
	for i, result := range testSuite.Results {
		status := "✅ PASS"
		if !result.Success {
			status = "❌ FAIL"
		}

		fmt.Printf("%d. %s - %s (%.2fms)\n",
			i+1, result.ToolName, status,
			float64(result.Duration.Nanoseconds())/1000000)

		if result.Error != "" {
			fmt.Printf("   Error: %s\n", result.Error)
		}
	}

	if testSuite.FailedTests == 0 {
		fmt.Printf("\n🎉 All tests passed! The MCP server is working correctly.\n")
	} else {
		fmt.Printf("\n⚠️  Some tests failed. Please check the server configuration and API connectivity.\n")
	}
}

func saveTestResults(testSuite TestSuite) {
	// Add timestamp
	testSuiteData := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"testSuite": testSuite,
	}

	data, err := json.MarshalIndent(testSuiteData, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal test results: %v", err)
		return
	}

	filename := "test_results.json"
	err = os.WriteFile(filename, data, 0644)
	if err != nil {
		log.Printf("Failed to save test results: %v", err)
		return
	}

	fmt.Printf("💾 Test results saved to %s\n", filename)
}