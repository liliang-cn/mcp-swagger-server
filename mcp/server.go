package mcp

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "strings"

    "github.com/go-openapi/spec"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)

type SwaggerMCPServer struct {
    server      *mcp.Server
    apiBaseURL  string
    swagger     *spec.Swagger
    apiKey      string
    filter      *APIFilter
    apiExecutor *APIExecutor
}

// NewSwaggerMCPServer creates a new MCP server from Swagger spec
func NewSwaggerMCPServer(apiBaseURL string, swaggerSpec *spec.Swagger, apiKey string) *SwaggerMCPServer {
    return NewSwaggerMCPServerWithFilter(apiBaseURL, swaggerSpec, apiKey, nil)
}

// NewSwaggerMCPServerWithFilter creates a new MCP server from Swagger spec with filtering
func NewSwaggerMCPServerWithFilter(apiBaseURL string, swaggerSpec *spec.Swagger, apiKey string, filter *APIFilter) *SwaggerMCPServer {
    // Create MCP server with Implementation
    implementation := &mcp.Implementation{
        Name:    "swagger-mcp-server",
        Version: "v1.0.0",
    }

    server := mcp.NewServer(implementation, nil)

    // Create converter
    converter := &SwaggerMCPServer{
        server:      server,
        apiBaseURL:  apiBaseURL,
        swagger:     swaggerSpec,
        apiKey:      apiKey,
        filter:      filter,
        apiExecutor: NewAPIExecutor(apiBaseURL, apiKey),
    }

    // Register tools from Swagger
    converter.RegisterTools()

    return converter
}

// Run starts the MCP server with stdio transport (backward compatibility)
func (s *SwaggerMCPServer) Run(ctx context.Context) error {
    return s.RunStdio(ctx)
}

// RunStdio starts the MCP server with stdio transport
func (s *SwaggerMCPServer) RunStdio(ctx context.Context) error {
    // Create stdio transport
    transport := &mcp.StdioTransport{}

    log.Println("Starting MCP server from Swagger with stdio transport...")

    // Run the server directly using the new v1.0.0 API
    return s.server.Run(ctx, transport)
}

// RunHTTP starts the MCP server with HTTP transport
func (s *SwaggerMCPServer) RunHTTP(addr string) error {
    log.Printf("Starting MCP server from Swagger with HTTP transport on %s...", addr)
    
    // Create the streamable HTTP handler
    handler := mcp.NewStreamableHTTPHandler(func(req *http.Request) *mcp.Server {
        return s.server
    }, nil)

    // Start the HTTP server
    return http.ListenAndServe(addr, handler)
}

// GetServer returns the underlying MCP server (useful for custom transport implementations)
func (s *SwaggerMCPServer) GetServer() *mcp.Server {
    return s.server
}

// RegisterTools creates MCP tools from Swagger endpoints
func (s *SwaggerMCPServer) RegisterTools() {
    for path, pathItem := range s.swagger.Paths.Paths {
        s.registerPathTools(path, pathItem)
    }
}

func (s *SwaggerMCPServer) registerPathTools(path string, pathItem spec.PathItem) {
    // Register GET endpoints
    if pathItem.Get != nil {
        s.registerOperation("GET", path, pathItem.Get)
    }

    // Register POST endpoints
    if pathItem.Post != nil {
        s.registerOperation("POST", path, pathItem.Post)
    }

    // Register PUT endpoints
    if pathItem.Put != nil {
        s.registerOperation("PUT", path, pathItem.Put)
    }

    // Register DELETE endpoints
    if pathItem.Delete != nil {
        s.registerOperation("DELETE", path, pathItem.Delete)
    }

    // Register PATCH endpoints
    if pathItem.Patch != nil {
        s.registerOperation("PATCH", path, pathItem.Patch)
    }
}

func (s *SwaggerMCPServer) registerOperation(method, path string, op *spec.Operation) {
    // Check if this operation should be excluded
    if s.filter != nil && s.filter.ShouldExcludeOperation(method, path, op) {
        return // Skip this operation
    }

    // Generate tool name using shared utility
    toolName := GenerateToolName(method, path, op)

    // Build description using shared utility
    description := GenerateToolDescription(method, path, op)

    // Create tool with basic info (input schema will be auto-generated)
    tool := &mcp.Tool{
        Name:        toolName,
        Description: description,
        InputSchema: s.buildParametersSchema(op.Parameters), // Keep manual schema for now
    }

    // Register the tool using the new generic AddTool function
    // This provides automatic type validation and schema generation
    mcp.AddTool(s.server, tool, s.createTypedHandler(method, path, op))
}

func (s *SwaggerMCPServer) buildParametersSchema(params []spec.Parameter) interface{} {
    properties := make(map[string]interface{})
    required := []string{}

    for _, param := range params {
        // Skip header and cookie params
        if param.In == "header" && !strings.EqualFold(param.Name, "content-type") {
            continue
        }
        if param.In == "cookie" {
            continue
        }

        // Create parameter schema based on type
        paramSchema := make(map[string]interface{})
        
        if param.Type != "" {
            paramSchema["type"] = getJSONType(param.Type)
        } else if param.Schema != nil {
            // Handle body parameters with schema
            if len(param.Schema.Type) > 0 {
                paramSchema["type"] = param.Schema.Type[0]
            } else {
                paramSchema["type"] = "object"
            }
            
            // Add properties if available
            if param.Schema.Properties != nil {
                props := make(map[string]interface{})
                for name, prop := range param.Schema.Properties {
                    propSchema := make(map[string]interface{})
                    if len(prop.Type) > 0 {
                        propSchema["type"] = prop.Type[0]
                    }
                    if prop.Description != "" {
                        propSchema["description"] = prop.Description
                    }
                    props[name] = propSchema
                }
                paramSchema["properties"] = props
            }
        }

        if param.Description != "" {
            paramSchema["description"] = param.Description
        }

        // Add format if specified
        if param.Format != "" {
            paramSchema["format"] = param.Format
        }

        // Handle array items
        if param.Type == "array" && param.Items != nil {
            itemSchema := make(map[string]interface{})
            if param.Items.Type != "" {
                itemSchema["type"] = getJSONType(param.Items.Type)
            }
            paramSchema["items"] = itemSchema
        }

        // Add to properties
        paramName := param.Name
        if param.In == "body" {
            // For body parameters, use "body" as the key
            paramName = "body"
        }
        properties[paramName] = paramSchema

        // Add to required if necessary
        if param.Required {
            required = append(required, paramName)
        }
    }

    schema := map[string]interface{}{
        "type":       "object",
        "properties": properties,
    }

    if len(required) > 0 {
        schema["required"] = required
    }

    return schema
}

// APIRequest represents the input structure for API calls
type APIRequest struct {
    // Dynamic parameters based on the Swagger spec
    // We use map[string]interface{} to handle various parameter types
    Params map[string]interface{} `json:"_params,omitempty"`
}

// APIResponse represents the output structure for API calls
type APIResponse struct {
    Content string `json:"content" jsonschema:"The response content from the API call"`
    Status  int    `json:"status,omitempty" jsonschema:"HTTP status code"`
}

// Create a typed handler function that works with the generic AddTool
func (s *SwaggerMCPServer) createTypedHandler(method, path string, op *spec.Operation) mcp.ToolHandlerFor[map[string]interface{}, APIResponse] {
    return func(ctx context.Context, req *mcp.CallToolRequest, args map[string]interface{}) (*mcp.CallToolResult, APIResponse, error) {
        // Use the shared API executor
        content, statusCode, err := s.apiExecutor.BuildAndExecuteRequest(ctx, method, path, args)
        if err != nil {
            return nil, APIResponse{}, err
        }

        // Create response
        apiResponse := APIResponse{
            Content: content,
            Status:  statusCode,
        }

        // Check status code and create appropriate MCP result
        if statusCode >= 400 {
            return &mcp.CallToolResult{
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Text: fmt.Sprintf("API error %d: %s", statusCode, content),
                    },
                },
                IsError: true,
            }, apiResponse, nil
        }

        return &mcp.CallToolResult{
            Content: []mcp.Content{
                &mcp.TextContent{
                    Text: content,
                },
            },
        }, apiResponse, nil
    }
}

// Create a handler function that works as a basic ToolHandler (legacy)
func (s *SwaggerMCPServer) createHandler(method, path string, op *spec.Operation) mcp.ToolHandler {
    return func(ctx context.Context, req *mcp.CallToolRequest) (*mcp.CallToolResult, error) {
        // Extract parameters from the request arguments
        var params map[string]interface{}
        if req.Params.Arguments != nil {
            if err := json.Unmarshal(req.Params.Arguments, &params); err != nil {
                return nil, fmt.Errorf("failed to parse arguments: %w", err)
            }
        }
        if params == nil {
            params = make(map[string]interface{})
        }

        // Use the shared API executor
        content, statusCode, err := s.apiExecutor.BuildAndExecuteRequest(ctx, method, path, params)
        if err != nil {
            return nil, err
        }

        // Check status code
        if statusCode >= 400 {
            return &mcp.CallToolResult{
                Content: []mcp.Content{
                    &mcp.TextContent{
                        Text: fmt.Sprintf("API error %d: %s", statusCode, content),
                    },
                },
                IsError: true,
            }, nil
        }

        return &mcp.CallToolResult{
            Content: []mcp.Content{
                &mcp.TextContent{
                    Text: content,
                },
            },
        }, nil
    }
}

func getJSONType(swaggerType string) string {
    switch swaggerType {
    case "integer":
        return "number"
    case "number":
        return "number"
    case "boolean":
        return "boolean"
    case "array":
        return "array"
    case "object":
        return "object"
    default:
        return "string"
    }
}