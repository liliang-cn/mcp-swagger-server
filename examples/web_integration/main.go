package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

// WebApp represents your existing web application
type WebApp struct {
	router    *http.ServeMux
	mcpServer *mcp.Server
	port      string
}

// NewWebApp creates a new web application instance
func NewWebApp() *WebApp {
	return &WebApp{
		router: http.NewServeMux(),
		port:   "5555", // Using non-standard port as per instructions
	}
}

func (app *WebApp) setupRoutes() {
	// Your existing API routes
	app.router.HandleFunc("/api/users", app.handleUsers)
	app.router.HandleFunc("/api/orders", app.handleOrders)
	
	// Health check
	app.router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
			log.Printf("Failed to encode health response: %v", err)
		}
	})
}

func (app *WebApp) handleUsers(w http.ResponseWriter, r *http.Request) {
	// Your existing user API logic
	users := []map[string]interface{}{
		{"id": 1, "name": "John Doe", "email": "john@example.com"},
		{"id": 2, "name": "Jane Smith", "email": "jane@example.com"},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(users); err != nil {
		log.Printf("Failed to encode users response: %v", err)
	}
}

func (app *WebApp) handleOrders(w http.ResponseWriter, r *http.Request) {
	// Your existing order API logic
	orders := []map[string]interface{}{
		{"id": 1, "user_id": 1, "amount": 100.50, "status": "completed"},
		{"id": 2, "user_id": 2, "amount": 75.25, "status": "pending"},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		log.Printf("Failed to encode orders response: %v", err)
	}
}

// setupMCPServer integrates MCP server with the web application
func (app *WebApp) setupMCPServer() error {
	// Option 1: Generate swagger from your existing routes (pseudo-code)
	// swaggerData := app.generateSwaggerFromRoutes()
	
	// Option 2: Use existing swagger file
	baseURL := "http://localhost:" + app.port
	server, err := mcp.NewFromSwaggerFile("../petstore.json", baseURL, "")
	if err != nil {
		return err
	}
	
	app.mcpServer = server
	
	// Start MCP server in background with HTTP transport
	go func() {
		ctx := context.Background()
		if err := server.RunHTTP(ctx, 7777); err != nil { // MCP HTTP server on different port
			log.Printf("MCP HTTP server error: %v", err)
		}
	}()
	
	return nil
}

// embedMCPServer embeds MCP server in the same HTTP server (alternative integration approach)
func (app *WebApp) embedMCPServer() error { //nolint:unused // Keep as alternative example
	baseURL := "http://localhost:" + app.port
	server, err := mcp.NewFromSwaggerFile("../petstore.json", baseURL, "")
	if err != nil {
		return err
	}
	
	app.mcpServer = server
	
	// Add MCP endpoints to your existing router
	app.router.HandleFunc("/mcp/tools", func(w http.ResponseWriter, r *http.Request) {
		// Proxy to MCP server's tools endpoint
		// This would require exposing more methods from the MCP server
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(map[string]string{
			"message": "MCP tools endpoint - implementation depends on your needs",
		}); err != nil {
			log.Printf("Failed to encode MCP tools response: %v", err)
		}
	})
	
	return nil
}

// Start starts the web application
func (app *WebApp) Start() error {
	// Setup your routes
	app.setupRoutes()
	
	// Setup MCP server integration
	if err := app.setupMCPServer(); err != nil {
		return err
	}
	
	// Alternative: embed MCP in same server
	// if err := app.embedMCPServer(); err != nil {
	//     return err
	// }
	
	log.Printf("Starting web server on port %s", app.port)
	log.Printf("MCP server running on port 7777")
	log.Printf("Visit http://localhost:%s/health for health check", app.port)
	
	return http.ListenAndServe(":"+app.port, app.router)
}

func main() {
	app := NewWebApp()
	
	if err := app.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}