package mcp_test

import (
	"fmt"
	"log"

	"github.com/go-openapi/spec"
	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

// Example demonstrates how to create an MCP server from a Swagger specification
func Example() {
	// Create a simple Swagger specification
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Version: "1.0.0",
					Title:   "Example API",
				},
			},
			Host:     "api.example.com",
			BasePath: "/v1",
			Schemes:  []string{"https"},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
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

	// Create the MCP server
	server := mcp.NewSwaggerMCPServer("https://api.example.com", swagger, "")
	
	// The server is now ready to handle MCP requests
	if server != nil {
		fmt.Println("MCP server created successfully")
	}
	
	// In production, you would run the server:
	// ctx := context.Background()
	// if err := server.Run(ctx); err != nil {
	//     log.Fatal(err)
	// }
	
	// Output: MCP server created successfully
}

// ExampleParseSwaggerSpec demonstrates parsing a Swagger specification from JSON
func ExampleParseSwaggerSpec() {
	jsonSpec := `{
		"swagger": "2.0",
		"info": {
			"version": "1.0.0",
			"title": "Pet Store API"
		},
		"paths": {
			"/pets": {
				"get": {
					"summary": "List all pets"
				}
			}
		}
	}`

	swagger, err := mcp.ParseSwaggerSpec([]byte(jsonSpec))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("API Title: %s\n", swagger.Info.Title)
	fmt.Printf("API Version: %s\n", swagger.Info.Version)
	
	// Output:
	// API Title: Pet Store API
	// API Version: 1.0.0
}

// ExampleFetchSwaggerFromURL demonstrates fetching a Swagger spec from a URL
func ExampleFetchSwaggerFromURL() {
	// This is a demonstration - in real use, provide a valid URL
	// url := "https://petstore.swagger.io/v2/swagger.json"
	// 
	// data, err := mcp.FetchSwaggerFromURL(url)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// 
	// swagger, err := mcp.ParseSwaggerSpec(data)
	// if err != nil {
	//     log.Fatal(err)
	// }
	// 
	// fmt.Printf("Fetched API: %s\n", swagger.Info.Title)
	
	fmt.Println("FetchSwaggerFromURL fetches specs from remote URLs")
	
	// Output: FetchSwaggerFromURL fetches specs from remote URLs
}

// ExampleNewSwaggerMCPServer demonstrates creating a new MCP server with authentication
func ExampleNewSwaggerMCPServer() {
	// Load your Swagger specification (simplified for example)
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Version: "1.0.0",
					Title:   "Authenticated API",
				},
			},
			Host:     "api.secure.com",
			BasePath: "/api",
			Schemes:  []string{"https"},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{
					"/secure": {
						PathItemProps: spec.PathItemProps{
							Get: &spec.Operation{
								OperationProps: spec.OperationProps{
									ID:      "getSecure",
									Summary: "Get secure data",
								},
							},
						},
					},
				},
			},
		},
	}

	// Create server with API key authentication
	apiKey := "your-secret-api-key"
	server := mcp.NewSwaggerMCPServer("https://api.secure.com/api", swagger, apiKey)

	// The server will automatically include the API key in all requests
	_ = server
	fmt.Println("Server created with authentication")
	
	// Output: Server created with authentication
}

// ExampleSwaggerMCPServer_Run demonstrates running the MCP server
func ExampleSwaggerMCPServer_Run() {
	// This example shows how to run the server
	// In practice, this would be in your main function
	
	/*
		swagger := loadYourSwaggerSpec()
		server := mcp.NewSwaggerMCPServer("https://api.example.com", swagger, "")
		
		ctx := context.Background()
		
		// The server will run until the context is cancelled
		if err := server.Run(ctx); err != nil {
			log.Fatal(err)
		}
	*/
	
	fmt.Println("Server runs with stdio transport")
	
	// Output: Server runs with stdio transport
}

// Example_fullWorkflow demonstrates the complete workflow from spec to MCP server
func Example_fullWorkflow() {
	// Step 1: Load specification (from file or URL)
	jsonSpec := `{
		"swagger": "2.0",
		"info": {
			"version": "1.0.0",
			"title": "Todo API"
		},
		"host": "api.todos.com",
		"paths": {
			"/todos": {
				"get": {
					"operationId": "listTodos",
					"summary": "List all todos"
				},
				"post": {
					"operationId": "createTodo",
					"summary": "Create a new todo",
					"parameters": [{
						"name": "body",
						"in": "body",
						"schema": {
							"type": "object",
							"properties": {
								"title": {"type": "string"},
								"completed": {"type": "boolean"}
							}
						}
					}]
				}
			}
		}
	}`

	// Step 2: Parse the specification
	swagger, err := mcp.ParseSwaggerSpec([]byte(jsonSpec))
	if err != nil {
		log.Fatal(err)
	}

	// Step 3: Create MCP server
	baseURL := fmt.Sprintf("https://%s", swagger.Host)
	server := mcp.NewSwaggerMCPServer(baseURL, swagger, "")

	// Step 4: Server is ready to handle MCP protocol
	_ = server
	
	fmt.Printf("Created MCP server for %s with %d paths\n", 
		swagger.Info.Title, 
		len(swagger.Paths.Paths))
	
	// Step 5: In production, run the server:
	// ctx := context.Background()
	// server.Run(ctx)
	
	// Output: Created MCP server for Todo API with 1 paths
}

// Example_yamlParsing demonstrates parsing YAML specifications
func Example_yamlParsing() {
	yamlSpec := `swagger: "2.0"
info:
  version: "1.0.0"
  title: "YAML API"
paths:
  /health:
    get:
      summary: "Health check"`

	swagger, err := mcp.ParseSwaggerSpec([]byte(yamlSpec))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Parsed YAML API: %s\n", swagger.Info.Title)
	
	// Output: Parsed YAML API: YAML API
}