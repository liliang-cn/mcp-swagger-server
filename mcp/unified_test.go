package mcp

import (
    "testing"

    "github.com/go-openapi/spec"
)

// TestGenerateToolName verifies consistent tool name generation
func TestGenerateToolName(t *testing.T) {
    tests := []struct {
        name     string
        method   string
        path     string
        opID     string
        expected string
    }{
        {
            name:     "with operation ID",
            method:   "GET",
            path:     "/users/{id}",
            opID:     "getUserById",
            expected: "getuserbyid",
        },
        {
            name:     "without operation ID",
            method:   "POST",
            path:     "/users",
            opID:     "",
            expected: "post_users",
        },
        {
            name:     "with path parameters",
            method:   "GET",
            path:     "/users/{id}/posts/{postId}",
            opID:     "",
            expected: "get_users_id_posts_postId",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            op := spec.NewOperation(tt.opID)
            result := GenerateToolName(tt.method, tt.path, op)
            if result != tt.expected {
                t.Errorf("GenerateToolName() = %v, want %v", result, tt.expected)
            }
        })
    }
}

// TestGenerateToolDescription verifies consistent description generation
func TestGenerateToolDescription(t *testing.T) {
    tests := []struct {
        name        string
        method      string
        path        string
        summary     string
        description string
        expected    string
    }{
        {
            name:     "with summary",
            method:   "GET",
            path:     "/users",
            summary:  "Get all users",
            expected: "Get all users",
        },
        {
            name:        "with description only",
            method:      "POST",
            path:        "/users",
            description: "Create a new user",
            expected:    "Create a new user",
        },
        {
            name:     "without summary or description",
            method:   "DELETE",
            path:     "/users/{id}",
            expected: "DELETE /users/{id}",
        },
        {
            name:        "summary takes precedence",
            method:      "PUT",
            path:        "/users/{id}",
            summary:     "Update user",
            description: "This is a longer description",
            expected:    "Update user",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            op := spec.NewOperation("")
            if tt.summary != "" {
                op.WithSummary(tt.summary)
            }
            if tt.description != "" {
                op.WithDescription(tt.description)
            }
            result := GenerateToolDescription(tt.method, tt.path, op)
            if result != tt.expected {
                t.Errorf("GenerateToolDescription() = %v, want %v", result, tt.expected)
            }
        })
    }
}

// TestAPIExecutor verifies the API executor can be created
func TestAPIExecutor(t *testing.T) {
    executor := NewAPIExecutor("https://api.example.com", "test-key")
    
    if executor == nil {
        t.Fatal("Expected non-nil executor")
    }
    
    if executor.APIBaseURL != "https://api.example.com" {
        t.Errorf("Expected base URL to be https://api.example.com, got %s", executor.APIBaseURL)
    }
    
    if executor.APIKey != "test-key" {
        t.Errorf("Expected API key to be test-key, got %s", executor.APIKey)
    }
}
