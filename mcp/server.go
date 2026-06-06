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

// Skill represents an Agent Skill metadata
type Skill struct {
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Tag         string   `json:"tag"`
    Operations  []string `json:"operations"`
}

// SkillsMetadata holds all skills metadata
type SkillsMetadata struct {
    Skills []Skill `json:"skills"`
}

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

// GetSkillsMetadata returns skills metadata for all API operations
func (s *SwaggerMCPServer) GetSkillsMetadata() *SkillsMetadata {
    tagGroups := s.groupOperationsByTag()

    skills := make([]Skill, 0, len(tagGroups))
    for tag, operations := range tagGroups {
        operationNames := make([]string, len(operations))

        for i, op := range operations {
            if op.Spec.ID != "" {
                operationNames[i] = op.Spec.ID
            } else {
                operationNames[i] = GenerateToolName(op.Method, op.Path, op.Spec)
            }
        }

        description := s.generateSkillDescription(tag, operations)

        skills = append(skills, Skill{
            Name:        toTitleCase(tag),
            Description: description,
            Tag:         tag,
            Operations:  operationNames,
        })
    }

    return &SkillsMetadata{Skills: skills}
}

// GetSkillsMetadataJSON returns skills metadata as JSON
func (s *SwaggerMCPServer) GetSkillsMetadataJSON() (string, error) {
    metadata := s.GetSkillsMetadata()
    data, err := json.MarshalIndent(metadata, "", "  ")
    if err != nil {
        return "", err
    }
    return string(data), nil
}

// groupOperationsByTag groups operations by their tags
func (s *SwaggerMCPServer) groupOperationsByTag() map[string][]Operation {
    groups := make(map[string][]Operation)

    for path, pathItem := range s.swagger.Paths.Paths {
        operations := []struct {
            method string
            op     *spec.Operation
        }{
            {"GET", pathItem.Get},
            {"POST", pathItem.Post},
            {"PUT", pathItem.Put},
            {"DELETE", pathItem.Delete},
            {"PATCH", pathItem.Patch},
        }

        for _, op := range operations {
            if op.op == nil {
                continue
            }

            // Get tags - default to "default" if no tags
            tags := op.op.Tags
            if len(tags) == 0 {
                tags = []string{"default"}
            }

            for _, tag := range tags {
                // Clean tag name
                cleanTag := sanitizeName(tag)
                groups[cleanTag] = append(groups[cleanTag], Operation{
                    Method: op.method,
                    Path:   path,
                    Spec:   op.op,
                    Tag:    tag,
                })
            }
        }
    }

    return groups
}

// generateSkillDescription creates a description for a skill
func (s *SwaggerMCPServer) generateSkillDescription(tag string, operations []Operation) string {
    if len(operations) == 0 {
        return fmt.Sprintf("API operations for %s", tag)
    }

    var summaries []string
    for _, op := range operations {
        if op.Spec.Summary != "" {
            summaries = append(summaries, op.Spec.Summary)
        } else if op.Spec.ID != "" {
            summaries = append(summaries, op.Spec.ID)
        }
    }

    if len(summaries) > 0 && len(summaries) <= 3 {
        return fmt.Sprintf("Provides %s", joinWithComma(summaries))
    }

    return fmt.Sprintf("API operations for %s (%d endpoints)", tag, len(operations))
}

// GenerateSkills generates Agent Skills files to the specified directory
func (s *SwaggerMCPServer) GenerateSkills(outputDir string) error {
    generator := NewSkillsGenerator(s.swagger, s.apiBaseURL, outputDir)
    return generator.Generate()
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
            // Body parameters carry a full JSON schema (already $ref-expanded
            // by ParseSwaggerSpec). Serialize it as-is so nested objects,
            // arrays, required lists and descriptions all survive.
            if full := schemaToMap(param.Schema); full != nil {
                paramSchema = full
            } else {
                paramSchema["type"] = "object"
            }
            if _, ok := paramSchema["type"]; !ok {
                paramSchema["type"] = "object"
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

// schemaToMap serializes a spec.Schema into a generic JSON-schema map.
// Returns nil if the schema cannot be serialized.
func schemaToMap(schema *spec.Schema) map[string]interface{} {
    data, err := json.Marshal(schema)
    if err != nil {
        return nil
    }
    var m map[string]interface{}
    if err := json.Unmarshal(data, &m); err != nil {
        return nil
    }
    return m
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

// Helper functions for Skills generation (shared with skills_generator.go)

// sanitizeName cleans a tag name into a valid Agent Skills name per
// https://agentskills.io/specification: 1-64 chars, lowercase a-z0-9 and
// hyphens only, no leading/trailing/consecutive hyphens.
func sanitizeName(name string) string {
    name = strings.ToLower(name)
    name = strings.Map(func(r rune) rune {
        if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
            return r
        }
        return '-'
    }, name)

    // Collapse consecutive hyphens and trim leading/trailing ones
    for strings.Contains(name, "--") {
        name = strings.ReplaceAll(name, "--", "-")
    }
    name = strings.Trim(name, "-")

    if name == "" {
        return "default"
    }
    if len(name) > 64 {
        name = strings.Trim(name[:64], "-")
    }
    return name
}

// toTitleCase converts a string to title case
func toTitleCase(s string) string {
    if s == "" {
        return "API"
    }
    words := strings.Split(s, "-_-/")
    for i, w := range words {
        if w != "" {
            if len(w) > 0 {
               	words[i] = strings.ToUpper(w[:1]) + strings.ToLower(w[1:])
            }
        }
    }
    return strings.Join(words, " ")
}

// joinWithComma joins strings with commas
func joinWithComma(items []string) string {
    switch len(items) {
    case 0:
        return ""
    case 1:
        return items[0]
    case 2:
        return items[0] + " and " + items[1]
    default:
        return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
    }
}