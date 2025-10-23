package mcp

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"

    "github.com/go-openapi/spec"
)

// APIExecutor handles API request building and execution.
// This provides unified logic for both HTTP and stdio transports to execute API calls
// against the target API based on Swagger/OpenAPI specifications.
type APIExecutor struct {
    APIBaseURL string
    APIKey     string
}

// NewAPIExecutor creates a new API executor
func NewAPIExecutor(apiBaseURL, apiKey string) *APIExecutor {
    return &APIExecutor{
        APIBaseURL: apiBaseURL,
        APIKey:     apiKey,
    }
}

// BuildAndExecuteRequest builds and executes an API request
func (e *APIExecutor) BuildAndExecuteRequest(ctx context.Context, method, path string, args map[string]interface{}) (string, int, error) {
    // Build URL with path parameters
    url := e.APIBaseURL + path

    // Extract body parameter if present
    var bodyData interface{}
    if body, exists := args["body"]; exists {
        bodyData = body
        delete(args, "body")
    }

    // Replace path parameters
    for key, value := range args {
        placeholder := "{" + key + "}"
        if strings.Contains(url, placeholder) {
            url = strings.ReplaceAll(url, placeholder, fmt.Sprintf("%v", value))
            delete(args, key)
        }
    }

    // Prepare request body
    var body io.Reader
    if method == "POST" || method == "PUT" || method == "PATCH" {
        var dataToSend interface{}
        if bodyData != nil {
            dataToSend = bodyData
        } else if len(args) > 0 {
            dataToSend = args
        }

        if dataToSend != nil {
            jsonData, err := json.Marshal(dataToSend)
            if err != nil {
                return "", 0, fmt.Errorf("failed to marshal request body: %w", err)
            }
            body = bytes.NewReader(jsonData)
        }
    } else {
        // Add remaining args as query parameters
        if len(args) > 0 {
            queryParams := []string{}
            for key, value := range args {
                queryParams = append(queryParams, fmt.Sprintf("%s=%v", key, value))
            }
            if strings.Contains(url, "?") {
                url += "&" + strings.Join(queryParams, "&")
            } else {
                url += "?" + strings.Join(queryParams, "&")
            }
        }
    }

    // Create HTTP request
    httpReq, err := http.NewRequestWithContext(ctx, method, url, body)
    if err != nil {
        return "", 0, fmt.Errorf("failed to create request: %w", err)
    }

    // Set headers
    if body != nil {
        httpReq.Header.Set("Content-Type", "application/json")
    }
    httpReq.Header.Set("Accept", "application/json")

    // Add API key if configured
    if e.APIKey != "" {
        httpReq.Header.Set("X-API-Key", e.APIKey)
        httpReq.Header.Set("Authorization", "Bearer "+e.APIKey)
    }

    // Execute request
    client := &http.Client{}
    resp, err := client.Do(httpReq)
    if err != nil {
        return "", 0, fmt.Errorf("request failed: %w", err)
    }
    defer func() { _ = resp.Body.Close() }()

    // Read response
    responseBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return "", resp.StatusCode, fmt.Errorf("failed to read response: %w", err)
    }

    // Try to format JSON response
    var jsonResponse interface{}
    var content string
    if err := json.Unmarshal(responseBody, &jsonResponse); err == nil {
        formattedJSON, _ := json.MarshalIndent(jsonResponse, "", "  ")
        content = string(formattedJSON)
    } else {
        content = string(responseBody)
    }

    return content, resp.StatusCode, nil
}

// FindOperationByToolName finds the operation that matches a tool name
func FindOperationByToolName(toolName string, swagger *spec.Swagger, filter *APIFilter) (string, string, *spec.Operation) {
    for path, pathItem := range swagger.Paths.Paths {
        operations := map[string]*spec.Operation{
            "GET":    pathItem.Get,
            "POST":   pathItem.Post,
            "PUT":    pathItem.Put,
            "DELETE": pathItem.Delete,
            "PATCH":  pathItem.Patch,
        }

        for method, op := range operations {
            if op == nil {
                continue
            }
            
            // Check if operation should be excluded
            if filter != nil && filter.ShouldExcludeOperation(method, path, op) {
                continue
            }
            
            // Check if tool name matches
            if GenerateToolName(method, path, op) == toolName {
                return method, path, op
            }
        }
    }
    return "", "", nil
}
