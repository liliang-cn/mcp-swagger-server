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
//   - Automatic conversion of Swagger/OpenAPI endpoints to MCP tools
//   - Support for Swagger 2.0 and OpenAPI 3.0 specifications
//   - Handles JSON and YAML specification formats
//   - Full support for HTTP methods: GET, POST, PUT, DELETE, PATCH
//   - Automatic parameter handling (path, query, body)
//   - API key authentication support
//   - Real-time API request execution
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
// Override base URL:
//
//	mcp-swagger-server -swagger api.yaml -api-base https://api.production.com
//
// # Command Line Flags
//
//	-swagger string
//	    Path to local Swagger/OpenAPI specification file
//	-swagger-url string
//	    URL to fetch Swagger/OpenAPI specification from
//	-api-base string
//	    Override base URL for API requests (optional)
//	-api-key string
//	    API key for authentication (optional)
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