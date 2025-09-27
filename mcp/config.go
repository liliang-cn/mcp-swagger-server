package mcp

import (
	"context"
	"io"

	"github.com/go-openapi/spec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config holds the configuration for the MCP server
type Config struct {
	// API configuration
	APIBaseURL string
	APIKey     string
	
	// Swagger specification
	SwaggerSpec *spec.Swagger
	SwaggerData []byte // Raw swagger data for lazy loading
	
	// Server configuration
	Name        string
	Version     string
	Description string
	
	// Transport configuration
	Transport Transport
}

// Transport interface for different transport methods
type Transport interface {
	Connect(ctx context.Context, server *mcp.Server) (*mcp.ServerSession, error)
}

// StdioTransport implements stdio transport
type StdioTransport struct{}

func (t *StdioTransport) Connect(ctx context.Context, server *mcp.Server) (*mcp.ServerSession, error) {
	transport := &mcp.StdioTransport{}
	return server.Connect(ctx, transport, nil)
}

// HTTPTransport implements HTTP transport
type HTTPTransport struct {
	Port   int
	Host   string
	Path   string
	Writer io.Writer // For response output
}

func (t *HTTPTransport) Connect(ctx context.Context, server *mcp.Server) (*mcp.ServerSession, error) {
	// HTTP transport doesn't use the standard MCP session model
	// Instead, it runs as an HTTP server
	// For now, fallback to stdio for compatibility
	transport := &mcp.StdioTransport{}
	return server.Connect(ctx, transport, nil)
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Name:        "swagger-mcp-server",
		Version:     "v0.2.0",
		Description: "MCP server generated from Swagger/OpenAPI specification",
		Transport:   &StdioTransport{},
	}
}

// WithSwaggerSpec sets the swagger specification
func (c *Config) WithSwaggerSpec(swagger *spec.Swagger) *Config {
	c.SwaggerSpec = swagger
	return c
}

// WithSwaggerData sets raw swagger data for lazy loading
func (c *Config) WithSwaggerData(data []byte) *Config {
	c.SwaggerData = data
	return c
}

// WithAPIConfig sets API configuration
func (c *Config) WithAPIConfig(baseURL, apiKey string) *Config {
	c.APIBaseURL = baseURL
	c.APIKey = apiKey
	return c
}

// WithTransport sets the transport method
func (c *Config) WithTransport(transport Transport) *Config {
	c.Transport = transport
	return c
}

// WithServerInfo sets server information
func (c *Config) WithServerInfo(name, version, description string) *Config {
	c.Name = name
	c.Version = version
	c.Description = description
	return c
}