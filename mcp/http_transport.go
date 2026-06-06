package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-openapi/spec"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// HTTPServer wraps the MCP server for HTTP transport
type HTTPServer struct {
	server     *Server
	port       int
	host       string
	path       string
	httpServer *http.Server
}

// NewHTTPServer creates a new HTTP server wrapper
func NewHTTPServer(server *Server, port int, host, path string) *HTTPServer {
	if host == "" {
		host = "localhost"
	}
	if path == "" {
		path = "/mcp"
	}

	return &HTTPServer{
		server: server,
		port:   port,
		host:   host,
		path:   path,
	}
}

// Start starts the HTTP server
func (h *HTTPServer) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Add CORS middleware
	corsHandler := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next(w, r)
		}
	}

	// Ensure path ends with / for proper prefix matching in ServeMux
	basePath := h.path
	if basePath[len(basePath)-1] != '/' {
		basePath += "/"
	}

	// MCP endpoint - official MCP Streamable HTTP handler so that standard
	// MCP clients (e.g. `claude mcp add --transport http`) can connect.
	streamableHandler := sdk.NewStreamableHTTPHandler(func(req *http.Request) *sdk.Server {
		return h.server.GetMCPServer().GetServer()
	}, nil)
	mux.HandleFunc(h.path, corsHandler(streamableHandler.ServeHTTP))

	// Health check endpoint
	mux.HandleFunc(basePath+"health", corsHandler(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]string{
			"status":  "ok",
			"server":  h.server.config.Name,
			"version": h.server.config.Version,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Failed to encode health response: %v", err)
		}
	}))

	// Tools list endpoint
	mux.HandleFunc(basePath+"tools", corsHandler(h.handleToolsList))

	addr := fmt.Sprintf("%s:%d", h.host, h.port)
	h.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	log.Printf("Starting HTTP MCP server on %s%s", addr, h.path)

	go func() {
		<-ctx.Done()
		if err := h.httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("Failed to shutdown HTTP server: %v", err)
		}
	}()

	if err := h.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("HTTP server error: %w", err)
	}

	return nil
}

// handleToolsList handles GET /tools endpoint
func (h *HTTPServer) handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	tools := h.getAvailableTools()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"tools": tools,
	}); err != nil {
		log.Printf("Failed to encode tools response: %v", err)
	}
}

// getToolName generates tool name using shared utility
func (h *HTTPServer) getToolName(method, path string, op *spec.Operation) string {
	return GenerateToolName(method, path, op)
}

// getAvailableTools returns a list of available tools (applying filters)
func (h *HTTPServer) getAvailableTools() []map[string]interface{} {
	config := h.server.GetConfig()
	if config.SwaggerSpec == nil {
		return []map[string]interface{}{}
	}

	tools := []map[string]interface{}{}
	for path, pathItem := range config.SwaggerSpec.Paths.Paths {
		if pathItem.Get != nil && !h.shouldExcludeOperation("GET", path, pathItem.Get, config.Filter) {
			tools = append(tools, h.createToolInfo("GET", path, pathItem.Get))
		}
		if pathItem.Post != nil && !h.shouldExcludeOperation("POST", path, pathItem.Post, config.Filter) {
			tools = append(tools, h.createToolInfo("POST", path, pathItem.Post))
		}
		if pathItem.Put != nil && !h.shouldExcludeOperation("PUT", path, pathItem.Put, config.Filter) {
			tools = append(tools, h.createToolInfo("PUT", path, pathItem.Put))
		}
		if pathItem.Delete != nil && !h.shouldExcludeOperation("DELETE", path, pathItem.Delete, config.Filter) {
			tools = append(tools, h.createToolInfo("DELETE", path, pathItem.Delete))
		}
		if pathItem.Patch != nil && !h.shouldExcludeOperation("PATCH", path, pathItem.Patch, config.Filter) {
			tools = append(tools, h.createToolInfo("PATCH", path, pathItem.Patch))
		}
	}

	return tools
}

// shouldExcludeOperation checks if an operation should be excluded based on filters
func (h *HTTPServer) shouldExcludeOperation(method, path string, operation *spec.Operation, filter *APIFilter) bool {
	if filter == nil {
		return false
	}
	return filter.ShouldExcludeOperation(method, path, operation)
}

// createToolInfo creates tool information from swagger operation
func (h *HTTPServer) createToolInfo(method, path string, op *spec.Operation) map[string]interface{} {
	toolName := h.getToolName(method, path, op)
	description := GenerateToolDescription(method, path, op)

	// Build parameter schema
	parameters := []map[string]interface{}{}
	for _, param := range op.Parameters {
		paramInfo := map[string]interface{}{
			"name":        param.Name,
			"in":          param.In,
			"required":    param.Required,
			"description": param.Description,
		}

		if param.Type != "" {
			paramInfo["type"] = param.Type
		}
		if param.Format != "" {
			paramInfo["format"] = param.Format
		}

		parameters = append(parameters, paramInfo)
	}

	return map[string]interface{}{
		"name":        toolName,
		"description": description,
		"method":      method,
		"path":        path,
		"parameters":  parameters,
		"operationId": op.ID,
	}
}

// RunHTTP runs the server with HTTP transport
func (s *Server) RunHTTP(ctx context.Context, port int) error {
	httpServer := NewHTTPServer(s, port, "", "")
	return httpServer.Start(ctx)
}

// WithHTTPTransport configures the server to use HTTP transport
func (c *Config) WithHTTPTransport(port int, host, path string) *Config {
	c.Transport = &HTTPTransport{
		Port: port,
		Host: host,
		Path: path,
	}
	return c
}
