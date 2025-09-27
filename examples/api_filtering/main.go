package main

import (
	"context"
	"fmt"
	"log"

	"github.com/go-openapi/spec"
	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

// This example demonstrates how to use API filtering to exclude certain endpoints
// from being converted to MCP tools
func main() {
	fmt.Println("=== API Filtering Example ===")

	// Example 1: Exclude specific paths
	fmt.Println("\n1. Excluding specific paths...")
	config1 := mcp.DefaultConfig().
		WithSwaggerData(createSampleSwagger()).
		WithAPIConfig("https://api.example.com", "").
		WithExcludePaths("/users/{id}", "/admin/settings")

	server1, err := mcp.New(config1)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	fmt.Printf("Server created with path exclusion filtering\n")

	// Example 2: Exclude by HTTP methods
	fmt.Println("\n2. Excluding DELETE and PATCH methods...")
	config2 := mcp.DefaultConfig().
		WithSwaggerData(createSampleSwagger()).
		WithAPIConfig("https://api.example.com", "").
		WithExcludeMethods("DELETE", "PATCH")

	server2, err := mcp.New(config2)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	fmt.Printf("Server created with method exclusion filtering\n")

	// Example 3: Include only specific paths
	fmt.Println("\n3. Including only user-related endpoints...")
	config3 := mcp.DefaultConfig().
		WithSwaggerData(createSampleSwagger()).
		WithAPIConfig("https://api.example.com", "").
		WithIncludeOnlyPaths("/users", "/users/{id}")

	server3, err := mcp.New(config3)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	fmt.Printf("Server created with path inclusion filtering\n")

	// Example 4: Exclude by path patterns (wildcards)
	fmt.Println("\n4. Excluding admin endpoints with wildcards...")
	config4 := mcp.DefaultConfig().
		WithSwaggerData(createSampleSwagger()).
		WithAPIConfig("https://api.example.com", "").
		WithExcludePathPatterns("/admin/*")

	server4, err := mcp.New(config4)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	fmt.Printf("Server created with pattern exclusion filtering\n")

	// Example 5: Complex filtering - exclude admin paths and DELETE methods
	fmt.Println("\n5. Complex filtering (exclude admin + DELETE methods)...")
	filter := &mcp.APIFilter{
		ExcludePathPatterns: []string{"/admin/*"},
		ExcludeMethods:      []string{"DELETE"},
	}
	
	config5 := mcp.DefaultConfig().
		WithSwaggerData(createSampleSwagger()).
		WithAPIConfig("https://api.example.com", "").
		WithAPIFilter(filter)

	server5, err := mcp.New(config5)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	fmt.Printf("Server created with complex filtering\n")

	// Example 6: Testing the filter logic directly
	fmt.Println("\n6. Testing filter logic directly...")
	testFilter := &mcp.APIFilter{
		ExcludePaths:        []string{"/users/{id}"},
		ExcludePathPatterns: []string{"/admin/*"},
		ExcludeMethods:      []string{"DELETE"},
	}

	// Test cases
	testCases := []struct {
		method string
		path   string
		opID   string
		expectedExcluded bool
	}{
		{"GET", "/users", "", false},          // Should be included
		{"GET", "/users/{id}", "", true},      // Should be excluded (exact path match)
		{"GET", "/admin/settings", "", true},  // Should be excluded (pattern match)
		{"DELETE", "/posts", "", true},        // Should be excluded (method match)
		{"POST", "/posts", "", false},         // Should be included
	}

	for _, tc := range testCases {
		// Create a minimal operation for testing
		op := &spec.Operation{}
		if tc.opID != "" {
			op.ID = tc.opID
		}
		excluded := testFilter.ShouldExcludeOperation(tc.method, tc.path, op)
		status := "✓"
		if excluded != tc.expectedExcluded {
			status = "✗"
		}
		fmt.Printf("  %s %s %s -> excluded: %v (expected: %v)\n", 
			status, tc.method, tc.path, excluded, tc.expectedExcluded)
	}

	fmt.Println("\n=== Filtering example completed ===")
	
	// Note: We're not actually running the servers since this is just a demo
	// In real usage, you would call server.RunStdio(context.Background()) or similar
	_ = server1
	_ = server2
	_ = server3
	_ = server4
	_ = server5
	_ = context.Background()
}

// createSampleSwagger creates a minimal swagger spec for testing
func createSampleSwagger() []byte {
	return []byte(`{
		"swagger": "2.0",
		"info": {
			"title": "Sample API",
			"version": "1.0.0"
		},
		"host": "api.example.com",
		"basePath": "/v1",
		"paths": {
			"/users": {
				"get": {
					"operationId": "getUsers",
					"summary": "Get all users"
				},
				"post": {
					"operationId": "createUser",
					"summary": "Create a user"
				}
			},
			"/users/{id}": {
				"get": {
					"operationId": "getUser",
					"summary": "Get a user by ID"
				},
				"delete": {
					"operationId": "deleteUser",
					"summary": "Delete a user"
				}
			},
			"/admin/settings": {
				"get": {
					"operationId": "getAdminSettings",
					"summary": "Get admin settings"
				},
				"patch": {
					"operationId": "updateAdminSettings",
					"summary": "Update admin settings"
				}
			},
			"/posts": {
				"get": {
					"operationId": "getPosts",
					"summary": "Get all posts"
				},
				"post": {
					"operationId": "createPost",
					"summary": "Create a post"
				},
				"delete": {
					"operationId": "deleteAllPosts",
					"summary": "Delete all posts"
				}
			}
		}
	}`)
}