package mcp

import (
	"testing"

	"github.com/go-openapi/spec"
)

// Basic tests that don't cause server startup issues

func TestDefaultConfig_Basic(t *testing.T) {
	config := DefaultConfig()
	if config == nil {
		t.Error("DefaultConfig should not return nil")
		return
	}
	if config.Name == "" {
		t.Error("DefaultConfig should have a name")
	}
	if config.Version == "" {
		t.Error("DefaultConfig should have a version")
	}
	if config.Transport == nil {
		t.Error("DefaultConfig should have a transport")
	}
}

func TestConfig_Chaining(t *testing.T) {
	swagger := &spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
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
		WithServerInfo("test-server", "v1.0.0", "Test description")

	if config.SwaggerSpec != swagger {
		t.Error("WithSwaggerSpec should set the swagger spec")
	}
	if config.APIBaseURL != "https://api.example.com" {
		t.Error("WithAPIConfig should set the base URL")
	}
	if config.APIKey != "test-key" {
		t.Error("WithAPIConfig should set the API key")
	}
	if config.Name != "test-server" {
		t.Error("WithServerInfo should set the name")
	}
}

func TestValidateConfig_Basic(t *testing.T) {
	// Test nil config
	err := validateConfig(nil)
	if err == nil {
		t.Error("validateConfig should fail with nil config")
	}

	// Test empty config
	err = validateConfig(&Config{})
	if err == nil {
		t.Error("validateConfig should fail with empty config")
	}

	// Test valid config
	config := &Config{
		Name:        "test",
		Version:     "1.0.0",
		SwaggerSpec: &spec.Swagger{},
		Transport:   &StdioTransport{},
	}
	err = validateConfig(config)
	if err != nil {
		t.Errorf("validateConfig should pass with valid config: %v", err)
	}
}

func TestInferBaseURL_Basic(t *testing.T) {
	tests := []struct {
		name    string
		swagger *spec.Swagger
		want    string
	}{
		{
			name: "with host and https scheme",
			swagger: &spec.Swagger{
				SwaggerProps: spec.SwaggerProps{
					Host:     "api.example.com",
					BasePath: "/v1",
					Schemes:  []string{"https", "http"},
				},
			},
			want: "https://api.example.com/v1",
		},
		{
			name: "with host and no schemes",
			swagger: &spec.Swagger{
				SwaggerProps: spec.SwaggerProps{
					Host:     "api.example.com",
					BasePath: "/v1",
				},
			},
			want: "https://api.example.com/v1",
		},
		{
			name: "no host",
			swagger: &spec.Swagger{
				SwaggerProps: spec.SwaggerProps{
					BasePath: "/v1",
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := inferBaseURL(tt.swagger)
			if got != tt.want {
				t.Errorf("inferBaseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseSwaggerSpec_Basic(t *testing.T) {
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

func TestGetJSONType_Basic(t *testing.T) {
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