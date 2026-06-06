package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const httpTestSwagger = `{
  "swagger": "2.0",
  "info": {"title": "Test API", "version": "1.0.0"},
  "paths": {
    "/pets": {
      "get": {
        "operationId": "listPets",
        "summary": "List all pets",
        "responses": {"200": {"description": "OK"}}
      }
    }
  }
}`

// startHTTPServer starts Server.RunHTTP on a free port and waits until it accepts connections.
func startHTTPServer(t *testing.T, server *Server) (endpoint string, cancel context.CancelFunc) {
	t.Helper()

	// Find a free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := l.Addr().(*net.TCPAddr).Port
	_ = l.Close()

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		_ = server.RunHTTP(ctx, port)
	}()

	endpoint = fmt.Sprintf("http://localhost:%d/mcp", port)

	// Wait for the server to come up
	for i := 0; i < 50; i++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("localhost:%d", port), 100*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return endpoint, cancel
		}
		time.Sleep(50 * time.Millisecond)
	}
	cancel()
	t.Fatal("HTTP server did not start in time")
	return "", nil
}

// TestRunHTTP_StandardMCPClient verifies that a standard MCP client using the
// Streamable HTTP transport can initialize, list tools, and call a tool
// against Server.RunHTTP.
func TestRunHTTP_StandardMCPClient(t *testing.T) {
	// Backend API the tool will call
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]any{{"id": 1, "name": "Buddy"}})
	}))
	defer backend.Close()

	config := DefaultConfig().
		WithSwaggerData([]byte(httpTestSwagger)).
		WithAPIConfig(backend.URL, "")

	server, err := New(config)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}

	endpoint, cancel := startHTTPServer(t, server)
	defer cancel()

	// Connect with the official MCP client over Streamable HTTP
	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "1.0"}, nil)
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()

	session, err := client.Connect(ctx, &sdk.StreamableClientTransport{Endpoint: endpoint}, nil)
	if err != nil {
		t.Fatalf("standard MCP client failed to initialize over HTTP: %v", err)
	}
	defer func() { _ = session.Close() }()

	// tools/list
	tools, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("tools/list failed: %v", err)
	}
	if len(tools.Tools) != 1 || tools.Tools[0].Name != "listpets" {
		t.Fatalf("expected tool 'listpets', got %+v", tools.Tools)
	}

	// tools/call
	result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "listpets"})
	if err != nil {
		t.Fatalf("tools/call failed: %v", err)
	}
	if result.IsError {
		t.Fatalf("tool call returned error: %+v", result.Content)
	}
	text, ok := result.Content[0].(*sdk.TextContent)
	if !ok || text.Text == "" {
		t.Fatalf("expected non-empty text content, got %+v", result.Content[0])
	}
}
