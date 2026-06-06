package mcp

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"

    "github.com/go-openapi/spec"
    "gopkg.in/yaml.v3"
)

// ParseSwaggerSpec parses a Swagger 2.0 or OpenAPI 3.x spec from JSON or
// YAML. OpenAPI 3.x documents are converted to Swagger 2.0 internally, and
// all $refs are expanded so downstream schema generation sees full schemas.
func ParseSwaggerSpec(data []byte) (*spec.Swagger, error) {
    jsonData := data

    // Normalize YAML input to JSON first
    if !json.Valid(data) {
        var yamlData map[string]interface{}
        if err := yaml.Unmarshal(data, &yamlData); err != nil {
            return nil, fmt.Errorf("failed to parse spec as JSON or YAML")
        }
        converted, err := json.Marshal(yamlData)
        if err != nil {
            return nil, fmt.Errorf("failed to convert YAML to JSON: %w", err)
        }
        jsonData = converted
    }

    // Convert OpenAPI 3.x documents to Swagger 2.0
    if isOpenAPI3(jsonData) {
        converted, err := convertOpenAPI3ToSwagger2(jsonData)
        if err != nil {
            return nil, err
        }
        jsonData = converted
    }

    var swagger spec.Swagger
    if err := json.Unmarshal(jsonData, &swagger); err != nil {
        return nil, fmt.Errorf("failed to parse spec: %w", err)
    }

    // Expand $refs (e.g. body schemas referencing #/definitions) so tool
    // input schemas include the full field list instead of a bare object.
    if err := spec.ExpandSpec(&swagger, nil); err != nil {
        return nil, fmt.Errorf("failed to expand spec refs: %w", err)
    }

    return &swagger, nil
}

// FetchSwaggerFromURL downloads a Swagger/OpenAPI spec from a URL
func FetchSwaggerFromURL(url string) ([]byte, error) {
    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("failed to fetch spec from URL: %w", err)
    }
    defer func() { _ = resp.Body.Close() }()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("failed to fetch spec, status code: %d", resp.StatusCode)
    }

    data, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    return data, nil
}

// readFile reads a file from disk
func readFile(filepath string) ([]byte, error) {
    return os.ReadFile(filepath)
}

// GenerateToolName generates a consistent tool name from method, path, and operation
func GenerateToolName(method, path string, op *spec.Operation) string {
    if op.ID != "" {
        toolName := strings.ReplaceAll(op.ID, " ", "_")
        return strings.ToLower(toolName)
    }
    
    // Create tool name from method and path
    toolName := strings.ToLower(method) + "_"
    pathName := strings.ReplaceAll(path, "/", "_")
    pathName = strings.ReplaceAll(pathName, "{", "")
    pathName = strings.ReplaceAll(pathName, "}", "")
    pathName = strings.TrimPrefix(pathName, "_")
    return toolName + pathName
}

// GenerateToolDescription generates a consistent tool description
func GenerateToolDescription(method, path string, op *spec.Operation) string {
    description := op.Summary
    if description == "" {
        description = op.Description
    }
    if description == "" {
        description = fmt.Sprintf("%s %s", method, path)
    }
    return description
}