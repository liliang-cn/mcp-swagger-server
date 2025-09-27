package mcp

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/spec"
)

func TestParseSwaggerSpec_JSON(t *testing.T) {
	jsonSpec := `{
		"swagger": "2.0",
		"info": {
			"version": "1.0.0",
			"title": "Test API"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint"
				}
			}
		}
	}`

	swagger, err := ParseSwaggerSpec([]byte(jsonSpec))
	if err != nil {
		t.Fatalf("Failed to parse JSON spec: %v", err)
	}

	if swagger.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", swagger.Info.Title)
	}

	if swagger.Info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", swagger.Info.Version)
	}

	if _, exists := swagger.Paths.Paths["/test"]; !exists {
		t.Error("Expected /test path to exist")
	}
}

func TestParseSwaggerSpec_YAML(t *testing.T) {
	yamlSpec := `swagger: "2.0"
info:
  version: "1.0.0"
  title: "Test API"
paths:
  /test:
    get:
      summary: "Test endpoint"`

	swagger, err := ParseSwaggerSpec([]byte(yamlSpec))
	if err != nil {
		t.Fatalf("Failed to parse YAML spec: %v", err)
	}

	if swagger == nil || swagger.Info == nil {
		t.Fatal("Failed to parse YAML spec properly")
	}
	
	if swagger.Info.Title != "Test API" {
		t.Errorf("Expected title 'Test API', got '%s'", swagger.Info.Title)
	}

	if swagger.Info.Version != "1.0.0" {
		t.Errorf("Expected version '1.0.0', got '%s'", swagger.Info.Version)
	}

	if _, exists := swagger.Paths.Paths["/test"]; !exists {
		t.Error("Expected /test path to exist")
	}
}

func TestParseSwaggerSpec_InvalidFormat(t *testing.T) {
	invalidSpec := `This is not valid JSON or YAML`

	_, err := ParseSwaggerSpec([]byte(invalidSpec))
	if err == nil {
		t.Error("Expected error for invalid spec format")
	}
}

func TestParseSwaggerSpec_ComplexSpec(t *testing.T) {
	complexSpec := `{
		"swagger": "2.0",
		"info": {
			"version": "1.0.0",
			"title": "Complex API",
			"description": "A complex API for testing"
		},
		"host": "api.example.com",
		"basePath": "/v1",
		"schemes": ["https"],
		"paths": {
			"/users/{userId}": {
				"get": {
					"operationId": "getUser",
					"summary": "Get user by ID",
					"parameters": [
						{
							"name": "userId",
							"in": "path",
							"required": true,
							"type": "string"
						}
					],
					"responses": {
						"200": {
							"description": "Success"
						}
					}
				},
				"put": {
					"operationId": "updateUser",
					"summary": "Update user",
					"parameters": [
						{
							"name": "userId",
							"in": "path",
							"required": true,
							"type": "string"
						},
						{
							"name": "body",
							"in": "body",
							"required": true,
							"schema": {
								"type": "object",
								"properties": {
									"name": {
										"type": "string"
									},
									"email": {
										"type": "string"
									}
								}
							}
						}
					]
				}
			}
		}
	}`

	swagger, err := ParseSwaggerSpec([]byte(complexSpec))
	if err != nil {
		t.Fatalf("Failed to parse complex spec: %v", err)
	}

	if swagger.Host != "api.example.com" {
		t.Errorf("Expected host 'api.example.com', got '%s'", swagger.Host)
	}

	if swagger.BasePath != "/v1" {
		t.Errorf("Expected base path '/v1', got '%s'", swagger.BasePath)
	}

	pathItem, exists := swagger.Paths.Paths["/users/{userId}"]
	if !exists {
		t.Fatal("Expected /users/{userId} path to exist")
	}

	if pathItem.Get == nil {
		t.Error("Expected GET operation to exist")
	}

	if pathItem.Put == nil {
		t.Error("Expected PUT operation to exist")
	}

	if pathItem.Get != nil && pathItem.Get.ID != "getUser" {
		t.Errorf("Expected operation ID 'getUser', got '%s'", pathItem.Get.ID)
	}
}

func TestFetchSwaggerFromURL(t *testing.T) {
	// Create a test server
	testSpec := `{
		"swagger": "2.0",
		"info": {
			"version": "1.0.0",
			"title": "Remote API"
		}
	}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(testSpec))
	}))
	defer server.Close()

	// Test fetching from URL
	data, err := FetchSwaggerFromURL(server.URL)
	if err != nil {
		t.Fatalf("Failed to fetch swagger from URL: %v", err)
	}

	// Verify the fetched data
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("Failed to unmarshal fetched data: %v", err)
	}

	if result["swagger"] != "2.0" {
		t.Errorf("Expected swagger version '2.0', got '%v'", result["swagger"])
	}

	info, ok := result["info"].(map[string]interface{})
	if !ok {
		t.Fatal("Failed to get info from result")
	}

	if info["title"] != "Remote API" {
		t.Errorf("Expected title 'Remote API', got '%v'", info["title"])
	}
}

func TestFetchSwaggerFromURL_404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not found"))
	}))
	defer server.Close()

	_, err := FetchSwaggerFromURL(server.URL)
	if err == nil {
		t.Error("Expected error for 404 response")
	}
}

func TestFetchSwaggerFromURL_InvalidURL(t *testing.T) {
	_, err := FetchSwaggerFromURL("not-a-valid-url")
	if err == nil {
		t.Error("Expected error for invalid URL")
	}
}

func TestGetJSONType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"integer", "number"},
		{"number", "number"},
		{"boolean", "boolean"},
		{"array", "array"},
		{"object", "object"},
		{"string", "string"},
		{"unknown", "string"},
		{"", "string"},
	}

	for _, test := range tests {
		result := getJSONType(test.input)
		if result != test.expected {
			t.Errorf("getJSONType(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestParseSwaggerSpec_WithDefinitions(t *testing.T) {
	specWithDefs := `{
		"swagger": "2.0",
		"info": {
			"version": "1.0.0",
			"title": "API with Definitions"
		},
		"paths": {
			"/pets": {
				"post": {
					"parameters": [{
						"name": "body",
						"in": "body",
						"schema": {
							"$ref": "#/definitions/Pet"
						}
					}]
				}
			}
		},
		"definitions": {
			"Pet": {
				"type": "object",
				"properties": {
					"name": {
						"type": "string"
					},
					"age": {
						"type": "integer"
					}
				}
			}
		}
	}`

	swagger, err := ParseSwaggerSpec([]byte(specWithDefs))
	if err != nil {
		t.Fatalf("Failed to parse spec with definitions: %v", err)
	}

	if swagger.Definitions == nil {
		t.Fatal("Expected definitions to exist")
	}

	petDef, exists := swagger.Definitions["Pet"]
	if !exists {
		t.Fatal("Expected Pet definition to exist")
	}

	if petDef.Type == nil || len(petDef.Type) == 0 || petDef.Type[0] != "object" {
		t.Error("Expected Pet to be of type object")
	}

	if petDef.Properties == nil {
		t.Fatal("Expected Pet to have properties")
	}

	if _, hasName := petDef.Properties["name"]; !hasName {
		t.Error("Expected Pet to have 'name' property")
	}
}

func BenchmarkParseSwaggerSpec_JSON(b *testing.B) {
	jsonSpec := `{
		"swagger": "2.0",
		"info": {
			"version": "1.0.0",
			"title": "Test API"
		},
		"paths": {
			"/test": {
				"get": {
					"summary": "Test endpoint"
				}
			}
		}
	}`

	data := []byte(jsonSpec)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, _ = ParseSwaggerSpec(data)
	}
}

func BenchmarkParseSwaggerSpec_YAML(b *testing.B) {
	yamlSpec := `swagger: "2.0"
info:
  version: "1.0.0"
  title: "Test API"
paths:
  /test:
    get:
      summary: "Test endpoint"`

	data := []byte(yamlSpec)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, _ = ParseSwaggerSpec(data)
	}
}

// Helper function to create a test swagger spec
func createTestSwagger() *spec.Swagger {
	return &spec.Swagger{
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
					"/test": {
						PathItemProps: spec.PathItemProps{
							Get: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID:      "testOperation",
									Summary: "Test operation",
									Parameters: []spec.Parameter{
										{
											SimpleSchema: spec.SimpleSchema{
												Type: "string",
											},
											ParamProps: spec.ParamProps{
												Name:     "query",
												In:       "query",
												Required: false,
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
}