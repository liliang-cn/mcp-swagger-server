// Package main provides a Model Context Protocol (MCP) server that converts
// Swagger/OpenAPI specifications into MCP tools, enabling AI assistants to
// interact with REST APIs dynamically.
//
// # Overview
//
// MCP Swagger Server bridges the gap between OpenAPI/Swagger specifications
// and the Model Context Protocol, automatically converting REST API endpoints
// into tools that AI assistants can use. This allows AI models to interact
// with any REST API that has a Swagger or OpenAPI specification.
//
// # Features
//
//   - Automatic conversion of Swagger/OpenAPI endpoints to MCP tools with proper schema generation
//   - Support for Swagger 2.0 and OpenAPI 3.0 specifications (JSON and YAML formats)
//   - Full support for HTTP methods: GET, POST, PUT, DELETE, PATCH
//   - Intelligent parameter handling (path, query, body)
//   - API key authentication support with multiple header formats
//   - Real-time API request execution
//   - Advanced API filtering system with path, method, tag, and operation ID filtering
//   - Dual transport support: stdio and HTTP with CORS
//   - Built-in HTTP API endpoints for health checks and tool listing
//   - Comprehensive error handling with proper HTTP status codes
//   - JSON response formatting and validation
//   - Flexible configuration system with fluent API
//
// # Installation
//
// Install the binary using go install:
//
//	go install github.com/liliang-cn/mcp-swagger-server@latest
//
// # Usage
//
// Basic usage with a local Swagger file:
//
//	mcp-swagger-server -swagger api.json
//
// Load specification from URL:
//
//	mcp-swagger-server -swagger-url https://api.example.com/swagger.json
//
// With API key authentication:
//
//	mcp-swagger-server -swagger api.json -api-key YOUR_API_KEY
//
// Override base URL (include basePath if defined in Swagger):
//
//	mcp-swagger-server -swagger api.yaml -api-base https://api.example.com/v2
//
// With HTTP transport:
//
//	mcp-swagger-server -swagger api.json -http-port 8127
//
// With filtering:
//
//	mcp-swagger-server -swagger api.json -exclude-paths "/admin/*" -exclude-methods "DELETE"
//
// # Command Line Flags
//
//	-swagger string
//	    Path to local Swagger/OpenAPI specification file (JSON or YAML)
//	-swagger-url string
//	    URL to fetch Swagger/OpenAPI specification from
//	-api-base string
//	    Override base URL for API requests (must include basePath from Swagger if defined)
//	-api-key string
//	    API key for authentication (optional)
//	-http-port int
//	    HTTP server port (default: 0 = use stdio transport)
//	    Example: -http-port 4539
//	-http-host string
//	    HTTP server host (default: localhost)
//	-http-path string
//	    HTTP server path for MCP endpoint (default: /mcp)
//	    All endpoints will be under this path: /mcp/health, /mcp/tools, /mcp
//	-exclude-paths string
//	    Comma-separated paths to exclude (supports wildcards)
//	-exclude-operations string
//	    Comma-separated operation IDs to exclude
//	-exclude-methods string
//	    Comma-separated HTTP methods to exclude
//	-exclude-tags string
//	    Comma-separated tags to exclude
//	-include-only-paths string
//	    Include only these paths (exclusive mode)
//	-include-only-operations string
//	    Include only these operation IDs (exclusive mode)
//
// # MCP Integration
//
// The server communicates via stdio using the Model Context Protocol.
// Configure it in your MCP client settings:
//
//	{
//	  "servers": {
//	    "my-api": {
//	      "command": "mcp-swagger-server",
//	      "args": ["-swagger", "/path/to/api.json"]
//	    }
//	  }
//	}
//
// # Tool Naming Convention
//
// Tools are named based on the OpenAPI operation ID or generated from
// the HTTP method and path:
//
//   - Operation ID "getUser" becomes tool "getuser"
//   - GET /users/{id} becomes tool "get_users_id"
//   - POST /users becomes tool "post_users"
//
// # Parameter Handling
//
// The server automatically handles different parameter types:
//
//   - Path parameters: Replaced in the URL path
//   - Query parameters: Added to the query string
//   - Body parameters: Sent as JSON request body
//   - Header parameters: Added to request headers (excluding auth)
//
// # Authentication
//
// When an API key is provided via -api-key flag, it's automatically
// included in requests using both common header formats:
//
//   - X-API-Key: YOUR_API_KEY
//   - Authorization: Bearer YOUR_API_KEY
//
// # Example
//
// Given a Swagger specification with a GET /users/{id} endpoint:
//
//	{
//	  "swagger": "2.0",
//	  "paths": {
//	    "/users/{id}": {
//	      "get": {
//	        "operationId": "getUser",
//	        "parameters": [
//	          {
//	            "name": "id",
//	            "in": "path",
//	            "required": true,
//	            "type": "string"
//	          }
//	        ]
//	      }
//	    }
//	  }
//	}
//
// This creates an MCP tool "getuser" that accepts an "id" parameter
// and makes a GET request to /users/{id}.
//
// # Architecture
//
// The server consists of three main components:
//
//   - Swagger Parser: Loads and validates OpenAPI/Swagger specifications
//   - Tool Generator: Converts API endpoints to MCP tool definitions
//   - Request Handler: Executes HTTP requests when tools are called
//
// # Error Handling
//
// The server provides detailed error messages for:
//
//   - Invalid Swagger/OpenAPI specifications
//   - Network errors during API requests
//   - Authentication failures (401/403 responses)
//   - Server errors (5xx responses)
//   - Malformed request parameters
//
// All errors are returned via the MCP protocol with descriptive messages.
//
// # License
//
// MIT License - see LICENSE file for details.
//
// # Repository
//
// https://github.com/liliang-cn/mcp-swagger-server
package main
