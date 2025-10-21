package mcp

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/go-openapi/spec"
)

// Server represents the MCP server that can run in different modes
type Server struct {
	config *Config
	mcp    *SwaggerMCPServer
}

// New creates a new MCP server with the given configuration
func New(config *Config) (*Server, error) {
	if config == nil {
		config = DefaultConfig()
	}
	
	// Validate configuration
	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}
	
	// Parse swagger spec if not already parsed
	if config.SwaggerSpec == nil && len(config.SwaggerData) > 0 {
		swagger, err := ParseSwaggerSpec(config.SwaggerData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse swagger spec: %w", err)
		}
		config.SwaggerSpec = swagger
	}
	
	// Determine base URL if not set
	if config.APIBaseURL == "" && config.SwaggerSpec != nil {
		config.APIBaseURL = inferBaseURL(config.SwaggerSpec)
	}
	
	// Create the underlying MCP server with filtering support
	mcpServer := NewSwaggerMCPServerWithFilter(config.APIBaseURL, config.SwaggerSpec, config.APIKey, config.Filter)
	
	return &Server{
		config: config,
		mcp:    mcpServer,
	}, nil
}

// NewFromSwaggerFile creates a server from a swagger file
func NewFromSwaggerFile(filePath, apiBaseURL, apiKey string) (*Server, error) {
	data, err := readFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read swagger file: %w", err)
	}
	
	config := DefaultConfig().
		WithSwaggerData(data).
		WithAPIConfig(apiBaseURL, apiKey)
	
	return New(config)
}

// NewFromSwaggerURL creates a server from a swagger URL
func NewFromSwaggerURL(url, apiBaseURL, apiKey string) (*Server, error) {
	data, err := FetchSwaggerFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch swagger from URL: %w", err)
	}
	
	config := DefaultConfig().
		WithSwaggerData(data).
		WithAPIConfig(apiBaseURL, apiKey)
	
	return New(config)
}

// NewFromSwaggerSpec creates a server from a parsed swagger spec
func NewFromSwaggerSpec(swagger *spec.Swagger, apiBaseURL, apiKey string) (*Server, error) {
	config := DefaultConfig().
		WithSwaggerSpec(swagger).
		WithAPIConfig(apiBaseURL, apiKey)
	
	return New(config)
}

// NewFromSwaggerData creates a server from raw swagger data
func NewFromSwaggerData(data []byte, apiBaseURL, apiKey string) (*Server, error) {
	config := DefaultConfig().
		WithSwaggerData(data).
		WithAPIConfig(apiBaseURL, apiKey)
	
	return New(config)
}

// Run starts the MCP server with the configured transport
func (s *Server) Run(ctx context.Context) error {
	log.Printf("Starting MCP server %s %s...", s.config.Name, s.config.Version)
	
	// Check if this is HTTP transport
	if httpTransport, ok := s.config.Transport.(*HTTPTransport); ok {
		// Use HTTP transport
		return s.RunHTTP(ctx, httpTransport.Port)
	}
	
	// Connect using the configured transport (stdio)
	session, err := s.config.Transport.Connect(ctx, s.mcp.server)
	if err != nil {
		return fmt.Errorf("failed to connect MCP server: %w", err)
	}
	
	// Wait for the session to end
	_ = session.Wait()
	return nil
}

// RunStdio runs the server with stdio transport (for CLI usage)
func (s *Server) RunStdio(ctx context.Context) error {
	// Temporarily override transport
	originalTransport := s.config.Transport
	s.config.Transport = &StdioTransport{}
	
	defer func() {
		s.config.Transport = originalTransport
	}()
	
	return s.Run(ctx)
}

// GetMCPServer returns the underlying MCP server for advanced usage
func (s *Server) GetMCPServer() *SwaggerMCPServer {
	return s.mcp
}

// GetConfig returns the server configuration
func (s *Server) GetConfig() *Config {
	return s.config
}

// ListTools returns a list of available tool names derived from the swagger spec and filters
func (s *Server) ListTools() []string {
	tools := []string{}
	cfg := s.GetConfig()
	if cfg == nil || cfg.SwaggerSpec == nil {
		return tools
	}

	for path, pathItem := range cfg.SwaggerSpec.Paths.Paths {
		if pathItem.Get != nil && !cfg.Filter.ShouldExcludeOperation("GET", path, pathItem.Get) {
			tools = append(tools, toolNameFor("GET", path, pathItem.Get))
		}
		if pathItem.Post != nil && !cfg.Filter.ShouldExcludeOperation("POST", path, pathItem.Post) {
			tools = append(tools, toolNameFor("POST", path, pathItem.Post))
		}
		if pathItem.Put != nil && !cfg.Filter.ShouldExcludeOperation("PUT", path, pathItem.Put) {
			tools = append(tools, toolNameFor("PUT", path, pathItem.Put))
		}
		if pathItem.Delete != nil && !cfg.Filter.ShouldExcludeOperation("DELETE", path, pathItem.Delete) {
			tools = append(tools, toolNameFor("DELETE", path, pathItem.Delete))
		}
		if pathItem.Patch != nil && !cfg.Filter.ShouldExcludeOperation("PATCH", path, pathItem.Patch) {
			tools = append(tools, toolNameFor("PATCH", path, pathItem.Patch))
		}
	}

	sort.Strings(tools)
	return tools
}

// validateConfig validates the server configuration
func validateConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("config cannot be nil")
	}
	
	if config.SwaggerSpec == nil && len(config.SwaggerData) == 0 {
		return fmt.Errorf("either SwaggerSpec or SwaggerData must be provided")
	}
	
	if config.Name == "" {
		return fmt.Errorf("server name cannot be empty")
	}
	
	if config.Version == "" {
		return fmt.Errorf("server version cannot be empty")
	}
	
	if config.Transport == nil {
		return fmt.Errorf("transport cannot be nil")
	}
	
	return nil
}

// inferBaseURL attempts to determine the base URL from swagger spec
func inferBaseURL(swagger *spec.Swagger) string {
	if swagger.Host != "" {
		scheme := "https"
		if len(swagger.Schemes) > 0 {
			scheme = swagger.Schemes[0]
		}
		return fmt.Sprintf("%s://%s%s", scheme, swagger.Host, swagger.BasePath)
	}
	return ""
}

// toolNameFor generates a tool name from method/path/operation using the same logic as HTTP transport
func toolNameFor(method, path string, op *spec.Operation) string {
	if op != nil && op.ID != "" {
		toolName := strings.ReplaceAll(op.ID, " ", "_")
		return strings.ToLower(toolName)
	}
	toolName := strings.ToLower(method) + "_"
	pathName := strings.ReplaceAll(path, "/", "_")
	pathName = strings.ReplaceAll(pathName, "{", "")
	pathName = strings.ReplaceAll(pathName, "}", "")
	pathName = strings.TrimPrefix(pathName, "_")
	return toolName + pathName
}
