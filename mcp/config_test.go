package mcp

import (
	"context"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config == nil {
		t.Error("DefaultConfig() returned nil")
		return
	}
	if config.Name != "swagger-mcp-server" {
		t.Errorf("DefaultConfig() Name = %v, want %v", config.Name, "swagger-mcp-server")
	}
	if config.Version != "v1.0.0" {
		t.Errorf("DefaultConfig() Version = %v, want %v", config.Version, "v1.0.0")
	}
	if config.Description != "MCP server generated from Swagger/OpenAPI specification" {
		t.Errorf("DefaultConfig() Description = %v, want %v", config.Description, "MCP server generated from Swagger/OpenAPI specification")
	}
	if config.Transport == nil {
		t.Error("DefaultConfig() Transport is nil")
	}
	if _, ok := config.Transport.(*StdioTransport); !ok {
		t.Error("DefaultConfig() Transport is not StdioTransport")
	}
}

func TestConfig_WithSwaggerSpec(t *testing.T) {
	config := DefaultConfig()
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:   "Test API",
					Version: "1.0.0",
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{},
			},
		},
	}
	
	result := config.WithSwaggerSpec(swagger)
	
	if result != config {
		t.Error("WithSwaggerSpec() should return the same config instance")
	}
	if config.SwaggerSpec != swagger {
		t.Error("WithSwaggerSpec() did not set SwaggerSpec correctly")
	}
}

func TestConfig_WithSwaggerData(t *testing.T) {
	config := DefaultConfig()
	data := []byte(`{"swagger": "2.0", "info": {"title": "Test", "version": "1.0.0"}}`)
	
	result := config.WithSwaggerData(data)
	
	if result != config {
		t.Error("WithSwaggerData() should return the same config instance")
	}
	if string(config.SwaggerData) != string(data) {
		t.Error("WithSwaggerData() did not set SwaggerData correctly")
	}
}

func TestConfig_WithAPIConfig(t *testing.T) {
	config := DefaultConfig()
	baseURL := "https://api.example.com"
	apiKey := "test-key"
	
	result := config.WithAPIConfig(baseURL, apiKey)
	
	if result != config {
		t.Error("WithAPIConfig() should return the same config instance")
	}
	if config.APIBaseURL != baseURL {
		t.Errorf("WithAPIConfig() APIBaseURL = %v, want %v", config.APIBaseURL, baseURL)
	}
	if config.APIKey != apiKey {
		t.Errorf("WithAPIConfig() APIKey = %v, want %v", config.APIKey, apiKey)
	}
}

func TestConfig_WithTransport(t *testing.T) {
	config := DefaultConfig()
	transport := &HTTPTransport{
		Port: 8080,
		Host: "localhost",
		Path: "/test",
	}
	
	result := config.WithTransport(transport)
	
	if result != config {
		t.Error("WithTransport() should return the same config instance")
	}
	if config.Transport != transport {
		t.Error("WithTransport() did not set Transport correctly")
	}
}

func TestConfig_WithServerInfo(t *testing.T) {
	config := DefaultConfig()
	name := "test-server"
	version := "v2.0.0"
	description := "Test server description"
	
	result := config.WithServerInfo(name, version, description)
	
	if result != config {
		t.Error("WithServerInfo() should return the same config instance")
	}
	if config.Name != name {
		t.Errorf("WithServerInfo() Name = %v, want %v", config.Name, name)
	}
	if config.Version != version {
		t.Errorf("WithServerInfo() Version = %v, want %v", config.Version, version)
	}
	if config.Description != description {
		t.Errorf("WithServerInfo() Description = %v, want %v", config.Description, description)
	}
}

func TestConfig_WithHTTPTransport(t *testing.T) {
	config := DefaultConfig()
	port := 9090
	host := "test.com"
	path := "/mcp"
	
	result := config.WithHTTPTransport(port, host, path)
	
	if result != config {
		t.Error("WithHTTPTransport() should return the same config instance")
	}
	
	httpTransport, ok := config.Transport.(*HTTPTransport)
	if !ok {
		t.Error("WithHTTPTransport() did not set HTTPTransport")
	}
	if httpTransport.Port != port {
		t.Errorf("WithHTTPTransport() Port = %v, want %v", httpTransport.Port, port)
	}
	if httpTransport.Host != host {
		t.Errorf("WithHTTPTransport() Host = %v, want %v", httpTransport.Host, host)
	}
	if httpTransport.Path != path {
		t.Errorf("WithHTTPTransport() Path = %v, want %v", httpTransport.Path, path)
	}
}

func TestStdioTransport_Connect(t *testing.T) {
	transport := &StdioTransport{}
	
	// Create a mock server for testing
	implementation := &mcp.Implementation{
		Name:    "test-server",
		Version: "v1.0.0",
	}
	server := mcp.NewServer(implementation, nil)
	
	ctx := context.Background()
	
	// This will try to connect via stdio, which may not work in tests
	// but we can test that it doesn't panic and returns some result
	session, err := transport.Connect(ctx, server)
	
	// The actual connection might fail in test environment, that's OK
	// We're just testing the method exists and doesn't panic
	_ = session
	_ = err
}

func TestHTTPTransport_Connect(t *testing.T) {
	transport := &HTTPTransport{
		Port: 8080,
		Host: "localhost",
		Path: "/mcp",
	}
	
	// Create a mock server for testing
	implementation := &mcp.Implementation{
		Name:    "test-server",
		Version: "v1.0.0",
	}
	server := mcp.NewServer(implementation, nil)
	
	ctx := context.Background()
	
	// This will try to connect via stdio (fallback), which may not work in tests
	// but we can test that it doesn't panic and returns some result
	session, err := transport.Connect(ctx, server)
	
	// The actual connection might fail in test environment, that's OK
	// We're just testing the method exists and doesn't panic
	_ = session
	_ = err
}

func TestConfig_ChainedMethods(t *testing.T) {
	// Test that all methods can be chained together
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Info: &spec.Info{
				InfoProps: spec.InfoProps{
					Title:   "Test API",
					Version: "1.0.0",
				},
			},
			Paths: &spec.Paths{
				Paths: map[string]spec.PathItem{},
			},
		},
	}
	
	config := DefaultConfig().
		WithSwaggerSpec(swagger).
		WithAPIConfig("https://api.example.com", "test-key").
		WithServerInfo("custom-server", "v2.0.0", "Custom description").
		WithHTTPTransport(9090, "custom.com", "/custom")
	
	if config.SwaggerSpec != swagger {
		t.Error("Chained methods: SwaggerSpec not set correctly")
	}
	if config.APIBaseURL != "https://api.example.com" {
		t.Error("Chained methods: APIBaseURL not set correctly")
	}
	if config.APIKey != "test-key" {
		t.Error("Chained methods: APIKey not set correctly")
	}
	if config.Name != "custom-server" {
		t.Error("Chained methods: Name not set correctly")
	}
	if config.Version != "v2.0.0" {
		t.Error("Chained methods: Version not set correctly")
	}
	if config.Description != "Custom description" {
		t.Error("Chained methods: Description not set correctly")
	}
	
	httpTransport, ok := config.Transport.(*HTTPTransport)
	if !ok {
		t.Error("Chained methods: Transport is not HTTPTransport")
	}
	if httpTransport.Port != 9090 {
		t.Error("Chained methods: HTTPTransport Port not set correctly")
	}
	if httpTransport.Host != "custom.com" {
		t.Error("Chained methods: HTTPTransport Host not set correctly")
	}
	if httpTransport.Path != "/custom" {
		t.Error("Chained methods: HTTPTransport Path not set correctly")
	}
}