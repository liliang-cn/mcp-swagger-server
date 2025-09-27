package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-openapi/spec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestIntegration_FullSwaggerToMCP(t *testing.T) {
	// Create a mock API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users":
			switch r.Method {
			case "GET":
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode([]map[string]interface{}{
					{"id": 1, "name": "Alice"},
					{"id": 2, "name": "Bob"},
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			case "POST":
				var body map[string]interface{}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					body = make(map[string]interface{})
				}
				body["id"] = 3
				w.WriteHeader(http.StatusCreated)
				if err := json.NewEncoder(w).Encode(body); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		case "/users/1":
			if r.Method == "GET" {
				w.Header().Set("Content-Type", "application/json")
				if err := json.NewEncoder(w).Encode(map[string]interface{}{
					"id": 1, "name": "Alice", "email": "alice@example.com",
				}); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer apiServer.Close()

	// Create swagger spec
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Version: "1.0.0",
					Title:   "User API",
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/users": {
						PathItemProps: spec.PathItemProps{
							Get: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID:      "listUsers",
									Summary: "List all users",
								},
							},
							Post: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID:      "createUser",
									Summary: "Create a new user",
									Parameters: []spec.Parameter{
										{
											ParamProps: spec.ParamProps{
												Name:     "body",
												In:       "body",
												Required: true,
												Schema: &spec.Schema{
													SchemaProps: spec.SchemaProps{
														Type: []string{"object"},
														Properties: map[string]spec.Schema{
															"name": {
																SchemaProps: spec.SchemaProps{
																	Type: []string{"string"},
																},
															},
															"email": {
																SchemaProps: spec.SchemaProps{
																	Type: []string{"string"},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
					"/users/{id}": {
						PathItemProps: spec.PathItemProps{
							Get: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID:      "getUser",
									Summary: "Get a user by ID",
									Parameters: []spec.Parameter{
										{
											SimpleSchema: spec.SimpleSchema{
												Type: "string",
											},
											ParamProps: spec.ParamProps{
												Name:     "id",
												In:       "path",
												Required: true,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create MCP server
	mcpServer := NewSwaggerMCPServer(apiServer.URL, swagger, "")

	// Test that tools are registered correctly
	if mcpServer.server == nil {
		t.Fatal("MCP server not initialized")
	}

	// Test listUsers tool
	t.Run("ListUsers", func(t *testing.T) {
		handler := mcpServer.createHandler("GET", "/users", swagger.Paths.Paths["/users"].Get)
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{},
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Handler error: %v", err)
		}

		if result.IsError {
			t.Error("Expected successful result")
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		var users []map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &users); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if len(users) != 2 {
			t.Errorf("Expected 2 users, got %d", len(users))
		}
	})

	// Test createUser tool
	t.Run("CreateUser", func(t *testing.T) {
		handler := mcpServer.createHandler("POST", "/users", swagger.Paths.Paths["/users"].Post)
		
		args := map[string]interface{}{
			"body": map[string]interface{}{
				"name":  "Charlie",
				"email": "charlie@example.com",
			},
		}
		argBytes, _ := json.Marshal(args)
		
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Arguments: json.RawMessage(argBytes),
			},
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Handler error: %v", err)
		}

		if result.IsError {
			t.Error("Expected successful result")
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		var user map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &user); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if user["name"] != "Charlie" {
			t.Errorf("Expected name 'Charlie', got '%v'", user["name"])
		}
	})

	// Test getUser tool with path parameter
	t.Run("GetUser", func(t *testing.T) {
		handler := mcpServer.createHandler("GET", "/users/{id}", swagger.Paths.Paths["/users/{id}"].Get)
		
		args := map[string]interface{}{
			"id": "1",
		}
		argBytes, _ := json.Marshal(args)
		
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{
				Arguments: json.RawMessage(argBytes),
			},
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Handler error: %v", err)
		}

		if result.IsError {
			t.Error("Expected successful result")
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		var user map[string]interface{}
		if err := json.Unmarshal([]byte(textContent.Text), &user); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if user["name"] != "Alice" {
			t.Errorf("Expected name 'Alice', got '%v'", user["name"])
		}

		if user["email"] != "alice@example.com" {
			t.Errorf("Expected email 'alice@example.com', got '%v'", user["email"])
		}
	})
}

func TestIntegration_ErrorHandling(t *testing.T) {
	// Create a mock API server that returns errors
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/timeout":
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		case "/error":
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(map[string]string{
				"error": "Internal server error",
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case "/unauthorized":
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte("Unauthorized"))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Not found"))
		}
	}))
	defer apiServer.Close()

	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Version: "1.0.0",
					Title:   "Error Test API",
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/error": {
						PathItemProps: spec.PathItemProps{
							Get: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID: "getError",
								},
							},
						},
					},
					"/unauthorized": {
						PathItemProps: spec.PathItemProps{
							Get: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID: "getUnauthorized",
								},
							},
						},
					},
				},
			},
		},
	}

	mcpServer := NewSwaggerMCPServer(apiServer.URL, swagger, "")

	t.Run("500Error", func(t *testing.T) {
		handler := mcpServer.createHandler("GET", "/error", swagger.Paths.Paths["/error"].Get)
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{},
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Handler error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result")
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		if !contains(textContent.Text, "API error 500") {
			t.Errorf("Expected 'API error 500' in error message, got: %s", textContent.Text)
		}
	})

	t.Run("401Unauthorized", func(t *testing.T) {
		handler := mcpServer.createHandler("GET", "/unauthorized", swagger.Paths.Paths["/unauthorized"].Get)
		req := &mcp.CallToolRequest{
			Params: &mcp.CallToolParamsRaw{},
		}

		result, err := handler(context.Background(), req)
		if err != nil {
			t.Fatalf("Handler error: %v", err)
		}

		if !result.IsError {
			t.Error("Expected error result")
		}

		textContent, ok := result.Content[0].(*mcp.TextContent)
		if !ok {
			t.Fatal("Expected TextContent")
		}

		if !contains(textContent.Text, "API error 401") {
			t.Errorf("Expected 'API error 401' in error message, got: %s", textContent.Text)
		}
	})
}

func TestIntegration_ComplexParameters(t *testing.T) {
	// Create a mock API server
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Echo back the received parameters for testing
		response := map[string]interface{}{
			"method":  r.Method,
			"path":    r.URL.Path,
			"query":   r.URL.Query(),
			"headers": r.Header,
		}
		
		if r.Method == "POST" || r.Method == "PUT" {
			var body interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
				response["body"] = body
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer apiServer.Close()

	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Version: "1.0.0",
					Title:   "Complex Parameters API",
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/complex/{id}": {
						PathItemProps: spec.PathItemProps{
							Post: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID: "complexOperation",
									Parameters: []spec.Parameter{
										{
											SimpleSchema: spec.SimpleSchema{
												Type: "string",
											},
											ParamProps: spec.ParamProps{
												Name:     "id",
												In:       "path",
												Required: true,
											},
										},
										{
											SimpleSchema: spec.SimpleSchema{
												Type: "array",
												Items: &spec.Items{
													SimpleSchema: spec.SimpleSchema{
														Type: "string",
													},
												},
											},
											ParamProps: spec.ParamProps{
												Name: "tags",
												In:   "query",
											},
										},
										{
											SimpleSchema: spec.SimpleSchema{
												Type: "integer",
											},
											ParamProps: spec.ParamProps{
												Name: "limit",
												In:   "query",
											},
										},
										{
											ParamProps: spec.ParamProps{
												Name:     "body",
												In:       "body",
												Required: true,
												Schema: &spec.Schema{
													SchemaProps: spec.SchemaProps{
														Type: []string{"object"},
														Properties: map[string]spec.Schema{
															"data": {
																SchemaProps: spec.SchemaProps{
																	Type: []string{"object"},
																},
															},
															"metadata": {
																SchemaProps: spec.SchemaProps{
																	Type: []string{"object"},
																	Properties: map[string]spec.Schema{
																		"timestamp": {
																			SchemaProps: spec.SchemaProps{
																				Type: []string{"integer"},
																			},
																		},
																	},
																},
															},
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	mcpServer := NewSwaggerMCPServer(apiServer.URL, swagger, "")

	handler := mcpServer.createHandler("POST", "/complex/{id}", swagger.Paths.Paths["/complex/{id}"].Post)
	
	args := map[string]interface{}{
		"id":    "test-123",
		"tags":  []string{"tag1", "tag2"},
		"limit": 10,
		"body": map[string]interface{}{
			"data": map[string]interface{}{
				"key": "value",
			},
			"metadata": map[string]interface{}{
				"timestamp": 1234567890,
			},
		},
	}
	argBytes, _ := json.Marshal(args)
	
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Arguments: json.RawMessage(argBytes),
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler error: %v", err)
	}

	if result.IsError {
		t.Error("Expected successful result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Verify the path parameter was replaced
	if response["path"] != "/complex/test-123" {
		t.Errorf("Expected path '/complex/test-123', got '%v'", response["path"])
	}

	// Verify query parameters
	query, ok := response["query"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected query to be a map")
	}

	// Note: query parameters are returned as arrays by Go's URL.Query()
	if tags, ok := query["tags"].([]interface{}); ok {
		if len(tags) == 0 {
			t.Error("Expected tags to be present")
		}
	}

	// Verify body was sent
	if body, ok := response["body"].(map[string]interface{}); ok {
		if data, ok := body["data"].(map[string]interface{}); ok {
			if data["key"] != "value" {
				t.Errorf("Expected body.data.key to be 'value', got '%v'", data["key"])
			}
		} else {
			t.Error("Expected body.data to be present")
		}
	} else {
		t.Error("Expected body to be present")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}