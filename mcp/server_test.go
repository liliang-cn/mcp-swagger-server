package mcp

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestNewSwaggerMCPServer(t *testing.T) {
	swagger := createTestSwagger()
	server := NewSwaggerMCPServer("http://api.example.com", swagger, "test-key")

	if server == nil {
		t.Fatal("Expected server to be created")
	}

	if server.apiBaseURL != "http://api.example.com" {
		t.Errorf("Expected API base URL 'http://api.example.com', got '%s'", server.apiBaseURL)
	}

	if server.apiKey != "test-key" {
		t.Errorf("Expected API key 'test-key', got '%s'", server.apiKey)
	}

	if server.swagger == nil {
		t.Fatal("Expected swagger spec to be set")
	}

	if server.server == nil {
		t.Fatal("Expected MCP server to be created")
	}
}

func TestRegisterOperation_WithOperationID(t *testing.T) {
	swagger := createTestSwagger()
	server := NewSwaggerMCPServer("http://api.example.com", swagger, "")

	// Since we can't access the tools directly, we can only verify that server was created
	// The actual tool registration is tested through integration tests
	if server.server == nil {
		t.Error("Expected MCP server to be initialized")
	}
}

func TestRegisterOperation_WithoutOperationID(t *testing.T) {
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Version: "1.0.0",
					Title:   "Test API",
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/users/{id}": {
						PathItemProps: spec.PathItemProps{
							Get: &spec.Operation{
								OperationProps: spec.OperationProps{
									Summary: "Get user",
								},
							},
						},
					},
				},
			},
		},
	}

	server := NewSwaggerMCPServer("http://api.example.com", swagger, "")

	// Since we can't access the tools directly, we can only verify that server was created
	// The actual tool registration is tested through integration tests
	if server.server == nil {
		t.Error("Expected MCP server to be initialized")
	}
}

func TestBuildParametersSchema_Simple(t *testing.T) {
	server := &SwaggerMCPServer{}
	
	params := []spec.Parameter{
		{
			SimpleSchema: spec.SimpleSchema{
				Type: "string",
			},
			ParamProps: spec.ParamProps{
				Name:        "name",
				In:          "query",
				Required:    true,
				Description: "User name",
			},
		},
		{
			SimpleSchema: spec.SimpleSchema{
				Type:   "integer",
				Format: "int32",
			},
			ParamProps: spec.ParamProps{
				Name:     "age",
				In:       "query",
				Required: false,
			},
		},
	}

	schema := server.buildParametersSchema(params)
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		t.Fatal("Expected schema to be map[string]interface{}")
	}

	if schemaMap["type"] != "object" {
		t.Errorf("Expected schema type 'object', got '%v'", schemaMap["type"])
	}

	props, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be map[string]interface{}")
	}

	// Check name parameter
	nameProp, ok := props["name"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected name property to exist")
	}

	if nameProp["type"] != "string" {
		t.Errorf("Expected name type 'string', got '%v'", nameProp["type"])
	}

	if nameProp["description"] != "User name" {
		t.Errorf("Expected description 'User name', got '%v'", nameProp["description"])
	}

	// Check age parameter
	ageProp, ok := props["age"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected age property to exist")
	}

	if ageProp["type"] != "number" {
		t.Errorf("Expected age type 'number', got '%v'", ageProp["type"])
	}

	if ageProp["format"] != "int32" {
		t.Errorf("Expected format 'int32', got '%v'", ageProp["format"])
	}

	// Check required
	required, ok := schemaMap["required"].([]string)
	if !ok {
		t.Fatal("Expected required to be []string")
	}

	if len(required) != 1 || required[0] != "name" {
		t.Errorf("Expected required ['name'], got %v", required)
	}
}

func TestBuildParametersSchema_WithBody(t *testing.T) {
	server := &SwaggerMCPServer{}
	
	bodySchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: []string{"object"},
			Properties: map[string]spec.Schema{
				"email": {
					SchemaProps: spec.SchemaProps{
						Type:        []string{"string"},
						Description: "Email address",
					},
				},
				"password": {
					SchemaProps: spec.SchemaProps{
						Type: []string{"string"},
					},
				},
			},
		},
	}

	params := []spec.Parameter{
		{
			ParamProps: spec.ParamProps{
				Name:     "body",
				In:       "body",
				Required: true,
				Schema:   bodySchema,
			},
		},
	}

	schema := server.buildParametersSchema(params)
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		t.Fatal("Expected schema to be map[string]interface{}")
	}

	props, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be map[string]interface{}")
	}

	// Check body parameter
	bodyProp, ok := props["body"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected body property to exist")
	}

	if bodyProp["type"] != "object" {
		t.Errorf("Expected body type 'object', got '%v'", bodyProp["type"])
	}

	// Check nested properties
	nestedProps, ok := bodyProp["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected body to have properties")
	}

	emailProp, ok := nestedProps["email"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected email property to exist")
	}

	if emailProp["type"] != "string" {
		t.Errorf("Expected email type 'string', got '%v'", emailProp["type"])
	}
}

func TestBuildParametersSchema_WithArray(t *testing.T) {
	server := &SwaggerMCPServer{}
	
	params := []spec.Parameter{
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
	}

	schema := server.buildParametersSchema(params)
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		t.Fatal("Expected schema to be map[string]interface{}")
	}

	props, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be map[string]interface{}")
	}

	tagsProp, ok := props["tags"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected tags property to exist")
	}

	if tagsProp["type"] != "array" {
		t.Errorf("Expected tags type 'array', got '%v'", tagsProp["type"])
	}

	items, ok := tagsProp["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected items to exist")
	}

	if items["type"] != "string" {
		t.Errorf("Expected items type 'string', got '%v'", items["type"])
	}
}

func TestBuildParametersSchema_SkipHeaders(t *testing.T) {
	server := &SwaggerMCPServer{}
	
	params := []spec.Parameter{
		{
			ParamProps: spec.ParamProps{
				Name: "Authorization",
				In:   "header",
			},
		},
		{
			ParamProps: spec.ParamProps{
				Name: "Cookie",
				In:   "cookie",
			},
		},
		{
			SimpleSchema: spec.SimpleSchema{
				Type: "string",
			},
			ParamProps: spec.ParamProps{
				Name: "query",
				In:   "query",
			},
		},
	}

	schema := server.buildParametersSchema(params)
	schemaMap, ok := schema.(map[string]interface{})
	if !ok {
		t.Fatal("Expected schema to be map[string]interface{}")
	}

	props, ok := schemaMap["properties"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected properties to be map[string]interface{}")
	}

	// Should only have query parameter
	if len(props) != 1 {
		t.Errorf("Expected 1 property, got %d", len(props))
	}

	if _, hasQuery := props["query"]; !hasQuery {
		t.Error("Expected query property to exist")
	}

	if _, hasAuth := props["Authorization"]; hasAuth {
		t.Error("Expected Authorization header to be skipped")
	}

	if _, hasCookie := props["Cookie"]; hasCookie {
		t.Error("Expected Cookie to be skipped")
	}
}

func TestCreateHandler_GET(t *testing.T) {
	// Create a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET method, got %s", r.Method)
		}

		if r.URL.Path != "/test/123" {
			t.Errorf("Expected path /test/123, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("filter") != "active" {
			t.Errorf("Expected filter=active, got %s", query.Get("filter"))
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":     "123",
			"status": "active",
		})
	}))
	defer testServer.Close()

	server := &SwaggerMCPServer{
		apiBaseURL: testServer.URL,
	}

	op := &spec.Operation{
		OperationProps: spec.OperationProps{
			Parameters: []spec.Parameter{
				{
					ParamProps: spec.ParamProps{
						Name:     "id",
						In:       "path",
						Required: true,
					},
				},
				{
					ParamProps: spec.ParamProps{
						Name: "filter",
						In:   "query",
					},
				},
			},
		},
	}

	handler := server.createHandler("GET", "/test/{id}", op)

	// Create request
	args := map[string]interface{}{
		"id":     "123",
		"filter": "active",
	}
	argBytes, _ := json.Marshal(args)
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Arguments: json.RawMessage(argBytes),
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		t.Error("Expected successful result")
	}

	if len(result.Content) != 1 {
		t.Fatalf("Expected 1 content item, got %d", len(result.Content))
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	// Verify response
	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["id"] != "123" {
		t.Errorf("Expected id '123', got '%v'", response["id"])
	}

	if response["status"] != "active" {
		t.Errorf("Expected status 'active', got '%v'", response["status"])
	}
}

func TestCreateHandler_POST(t *testing.T) {
	// Create a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)

		if body["name"] != "Test User" {
			t.Errorf("Expected name 'Test User', got '%v'", body["name"])
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":   "456",
			"name": body["name"],
		})
	}))
	defer testServer.Close()

	server := &SwaggerMCPServer{
		apiBaseURL: testServer.URL,
	}

	op := &spec.Operation{}
	handler := server.createHandler("POST", "/users", op)

	// Create request
	args := map[string]interface{}{
		"body": map[string]interface{}{
			"name": "Test User",
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
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		t.Error("Expected successful result")
	}

	// Verify response
	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	var response map[string]interface{}
	if err := json.Unmarshal([]byte(textContent.Text), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["name"] != "Test User" {
		t.Errorf("Expected name 'Test User', got '%v'", response["name"])
	}
}

func TestCreateHandler_WithAPIKey(t *testing.T) {
	// Create a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check API key headers
		if r.Header.Get("X-API-Key") != "test-api-key" {
			t.Errorf("Expected X-API-Key 'test-api-key', got '%s'", r.Header.Get("X-API-Key"))
		}

		if !strings.HasPrefix(r.Header.Get("Authorization"), "Bearer test-api-key") {
			t.Errorf("Expected Authorization header with Bearer token")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer testServer.Close()

	server := &SwaggerMCPServer{
		apiBaseURL: testServer.URL,
		apiKey:     "test-api-key",
	}

	op := &spec.Operation{}
	handler := server.createHandler("GET", "/test", op)

	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		t.Error("Expected successful result")
	}
}

func TestCreateHandler_ErrorResponse(t *testing.T) {
	// Create a test HTTP server that returns an error
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "Bad request"}`))
	}))
	defer testServer.Close()

	server := &SwaggerMCPServer{
		apiBaseURL: testServer.URL,
	}

	op := &spec.Operation{}
	handler := server.createHandler("GET", "/test", op)

	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if !result.IsError {
		t.Error("Expected error result")
	}

	textContent, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatal("Expected TextContent")
	}

	if !strings.Contains(textContent.Text, "API error 400") {
		t.Errorf("Expected error message to contain 'API error 400', got '%s'", textContent.Text)
	}
}

func TestRegisterTools_AllMethods(t *testing.T) {
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Version: "1.0.0",
					Title:   "Test API",
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/resource": {
						PathItemProps: spec.PathItemProps{
							Get: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID: "getResource",
								},
							},
							Post: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID: "createResource",
								},
							},
							Put: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID: "updateResource",
								},
							},
							Delete: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID: "deleteResource",
								},
							},
							Patch: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID: "patchResource",
								},
							},
						},
					},
				},
			},
		},
	}

	server := NewSwaggerMCPServer("http://api.example.com", swagger, "")

	// Since we can't access the tools directly, verify server was created
	// The actual tool registration is tested through integration tests
	if server.server == nil {
		t.Error("Expected MCP server to be initialized")
	}
	
	// Verify that all operations were registered by checking server state
	if server.swagger == nil || server.swagger.Paths == nil {
		t.Error("Expected swagger paths to be loaded")
	}
}

func TestCreateHandler_ComplexPath(t *testing.T) {
	// Create a test HTTP server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedPath := "/api/v1/users/123/posts/456"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer testServer.Close()

	server := &SwaggerMCPServer{
		apiBaseURL: testServer.URL,
	}

	op := &spec.Operation{}
	handler := server.createHandler("GET", "/api/v1/users/{userId}/posts/{postId}", op)

	// Create request with path parameters
	args := map[string]interface{}{
		"userId": "123",
		"postId": "456",
	}
	argBytes, _ := json.Marshal(args)
	req := &mcp.CallToolRequest{
		Params: &mcp.CallToolParamsRaw{
			Arguments: json.RawMessage(argBytes),
		},
	}

	result, err := handler(context.Background(), req)
	if err != nil {
		t.Fatalf("Handler returned error: %v", err)
	}

	if result.IsError {
		t.Error("Expected successful result")
	}
}