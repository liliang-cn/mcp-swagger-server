# MCP Swagger Server

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![Go Reference](https://pkg.go.dev/badge/github.com/liliang-cn/mcp-swagger-server.svg)](https://pkg.go.dev/github.com/liliang-cn/mcp-swagger-server)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/liliang-cn/mcp-swagger-server)](https://goreportcard.com/report/github.com/liliang-cn/mcp-swagger-server)

A Model Context Protocol (MCP) server that converts Swagger/OpenAPI specifications into MCP tools. This project can be used both as a standalone CLI tool and as a Go library for integration into other projects.

> **Status**: ✅ Production Ready - The project is stable, well-tested, and ready for production use.

> **📖 Documentation**: [中文文档](README_CN.md) | [English Documentation](README.md)

## Features

- **Dual Usage**: Works as standalone CLI and Go library
- **Multiple Transport**: Supports stdio and HTTP transport with CORS support
- **Complete Swagger Support**: Loads Swagger 2.0 and OpenAPI specifications (JSON or YAML)
- **Auto-conversion**: Automatically converts API endpoints to MCP tools with proper schema generation
- **Advanced API Filtering**: Comprehensive filtering system to control which APIs become tools
  - Path-based filtering with wildcard support
  - HTTP method filtering (GET, POST, PUT, DELETE, PATCH)
  - Operation ID filtering
  - Tag-based filtering
  - Include-only (whitelist) mode
- **Full HTTP Support**: Supports all HTTP methods (GET, POST, PUT, DELETE, PATCH)
- **Parameter Handling**: Intelligent handling of path parameters, query parameters, and request bodies
- **Authentication**: Automatic API key authentication support with multiple header formats
- **Web Integration**: Easy integration into existing Go web applications
- **HTTP API Endpoints**: Built-in HTTP endpoints for tools listing and health checks
- **Error Handling**: Comprehensive error handling with proper HTTP status codes
- **JSON Formatting**: Automatic JSON response formatting and validation
- **Server Configuration**: Flexible configuration system with fluent API

## Installation

### Using go install (Recommended)

```bash
go install github.com/liliang-cn/mcp-swagger-server@latest
```

This will install the `mcp-swagger-server` binary in your `$GOPATH/bin` directory.

### Using go get

```bash
go get github.com/liliang-cn/mcp-swagger-server
```

### Building from source

```bash
git clone https://github.com/liliang-cn/mcp-swagger-server.git
cd mcp-swagger-server
go build -o mcp-swagger-server .
```

## Build

```bash
go build -o mcp-swagger-server .
```

## Usage

### Standalone CLI Usage

#### From a local Swagger file:

```bash
./mcp-swagger-server -swagger examples/petstore.json
```

#### From a URL:

```bash
./mcp-swagger-server -swagger-url https://petstore.swagger.io/v2/swagger.json
```

#### With custom API base URL:

```bash
./mcp-swagger-server -swagger examples/api.yaml -api-base https://api.example.com
```

#### With API key authentication:

```bash
./mcp-swagger-server -swagger examples/api.json -api-key YOUR_API_KEY
```

#### With HTTP transport:

```bash
./mcp-swagger-server -swagger examples/api.json -http-port 8127
```

#### With API filtering (exclude admin endpoints):

```bash
./mcp-swagger-server -swagger examples/api.json -exclude-paths "/admin/*,/internal/*"
```

#### With multiple filtering options:

```bash
./mcp-swagger-server -swagger examples/api.json \
  -exclude-methods "DELETE,PATCH" \
  -exclude-tags "admin,internal" \
  -exclude-paths "/debug/*"
```

#### With HTTP transport and filtering:

```bash
./mcp-swagger-server -swagger examples/api.json \
  -http-port 8127 \
  -exclude-methods "DELETE,PATCH" \
  -exclude-paths "/admin/*"
```

### Go Library Usage

#### Basic Library Usage

```go
package main

import (
    "context"
    "log"
    "github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
    // Create MCP server from local file
    server, err := mcp.NewFromSwaggerFile("api.json", "https://api.example.com", "your-api-key")
    if err != nil {
        log.Fatal(err)
    }

    // Run with stdio transport (for CLI usage)
    ctx := context.Background()
    server.RunStdio(ctx)
}
```

#### Web Application Integration

```go
package main

import (
    "context"
    "net/http"
    "github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
    // Your existing web app setup
    router := http.NewServeMux()
    
    // Create MCP server from your API swagger
    mcpServer, _ := mcp.NewFromSwaggerFile("your-api.json", "http://localhost:6724", "")

    // Option 1: Run MCP server on separate HTTP port
    go mcpServer.RunHTTP(context.Background(), 8127)
    
    // Option 2: Integrate into existing server (custom implementation needed)
    
    // Your existing routes
    router.HandleFunc("/api/users", handleUsers)

    http.ListenAndServe(":6724", router)
}
```

#### Advanced Configuration

```go
config := mcp.DefaultConfig().
    WithAPIConfig("https://api.example.com", "your-api-key").
    WithServerInfo("my-api-server", "v1.0.0", "Custom API MCP Server").
    WithHTTPTransport(8127, "localhost", "/mcp")

server, err := mcp.New(config)
if err != nil {
    log.Fatal(err)
}

// Run with HTTP transport
server.Run(context.Background())
```

#### API Filtering in Go Library

```go
// Example 1: Exclude specific paths and methods
config := mcp.DefaultConfig().
    WithSwaggerData(swaggerData).
    WithAPIConfig("https://api.example.com", "your-api-key").
    WithExcludePaths("/admin/*", "/internal/*").
    WithExcludeMethods("DELETE", "PATCH")

server, err := mcp.New(config)
if err != nil {
    log.Fatal(err)
}

// Example 2: Include only specific endpoints
config := mcp.DefaultConfig().
    WithSwaggerData(swaggerData).
    WithAPIConfig("https://api.example.com", "your-api-key").
    WithIncludeOnlyPaths("/users", "/users/{id}", "/posts")

server, err := mcp.New(config)

// Example 3: Complex filtering with custom filter
filter := &mcp.APIFilter{
    ExcludePathPatterns: []string{"/admin/*", "/debug/*"},
    ExcludeMethods:      []string{"DELETE", "PATCH"},
    ExcludeTags:         []string{"internal", "admin"},
    IncludeOnlyOperationIDs: []string{"getUsers", "createUser", "getUser"},
}

config := mcp.DefaultConfig().
    WithSwaggerData(swaggerData).
    WithAPIConfig("https://api.example.com", "your-api-key").
    WithAPIFilter(filter)

server, err := mcp.New(config)
```

## Command Line Options

### Basic Options
- `-swagger` - Path to local Swagger/OpenAPI spec file (JSON or YAML)
- `-swagger-url` - URL to fetch Swagger/OpenAPI spec from
- `-api-base` - Override the base URL for API calls (defaults to spec's host)
- `-api-key` - API key for authentication

### Transport Options
- `-http-port` - HTTP server port (default: 0 = use stdio transport)
- `-http-host` - HTTP server host (default: localhost)
- `-http-path` - HTTP server path for MCP endpoint (default: /mcp)

### API Filtering Options
- `-exclude-paths` - Comma-separated list of paths to exclude (supports wildcards like `/admin/*`)
- `-exclude-operations` - Comma-separated list of operation IDs to exclude
- `-exclude-methods` - Comma-separated list of HTTP methods to exclude (e.g., `DELETE,PATCH`)
- `-exclude-tags` - Comma-separated list of Swagger tags to exclude
- `-include-only-paths` - Comma-separated list of paths to include exclusively (whitelist mode)
- `-include-only-operations` - Comma-separated list of operation IDs to include exclusively

## HTTP API Endpoints

When running with HTTP transport, the server exposes the following endpoints:

- `GET /health` - Health check endpoint with status information
- `GET /tools` - List available tools with detailed information
- `POST /mcp` - Execute MCP requests (supports tools/list and tools/call)
- `OPTIONS /mcp` - CORS preflight support

All HTTP endpoints include CORS headers for cross-origin requests.

### Example HTTP Usage

```bash
# Get available tools
curl http://localhost:8127/tools

# Execute a tool
curl -X POST http://localhost:8127/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "tools/call",
    "params": {
      "name": "get_users",
      "arguments": {
        "limit": 10
      }
    }
  }'
```

## Examples

### Example Swagger Spec

The `examples/petstore.json` file contains a sample Swagger specification for testing.

### API Filtering Example

Run the API filtering example to see how different filtering options work:

```bash
go run examples/api_filtering/main.go
```

This example demonstrates:
- Path-based exclusion
- HTTP method filtering  
- Include-only (whitelist) mode
- Wildcard pattern matching
- Complex filtering combinations
- Direct filter testing

## API Filtering

The MCP Swagger Server supports comprehensive API filtering to control which endpoints are exposed as MCP tools. This is essential for security and to prevent unwanted API access.

### Filtering Methods

#### 1. Path-based Filtering
```bash
# Exclude specific paths
-exclude-paths "/admin,/internal,/debug"

# Exclude paths with wildcards
-exclude-paths "/admin/*,/internal/*,/v1/debug/*"

# Include only specific paths (whitelist mode)
-include-only-paths "/users,/users/{id},/posts"
```

#### 2. HTTP Method Filtering
```bash
# Exclude dangerous methods
-exclude-methods "DELETE,PATCH"

# Only allow read operations
-include-only-methods "GET"
```

#### 3. Operation ID Filtering
```bash
# Exclude specific operations
-exclude-operations "deleteUser,deleteAllData,resetSystem"

# Include only specific operations
-include-only-operations "getUsers,getUser,createUser"
```

#### 4. Tag-based Filtering
```bash
# Exclude operations with specific tags
-exclude-tags "admin,internal,debug"
```

### Filtering Priority

1. **Include-only filters** are applied first (if specified)
2. **Exclude filters** are applied second
3. If both include-only and exclude filters are specified, an endpoint must pass both

### Wildcard Patterns

The filtering system supports wildcard patterns for paths:
- `*` matches any characters within a path segment
- `/admin/*` matches `/admin/users`, `/admin/settings`, etc.
- `/api/v*/admin` matches `/api/v1/admin`, `/api/v2/admin`, etc.

### Security Best Practices

1. **Always exclude administrative endpoints**: Use `-exclude-paths "/admin/*"`
2. **Limit dangerous HTTP methods**: Use `-exclude-methods "DELETE,PATCH"`
3. **Use whitelist mode for sensitive APIs**: Use `-include-only-paths` for maximum control
4. **Exclude internal/debug endpoints**: Use `-exclude-tags "internal,debug"`

### Examples

```bash
# Security-focused filtering
./mcp-swagger-server -swagger api.json \
  -exclude-paths "/admin/*,/internal/*,/debug/*" \
  -exclude-methods "DELETE,PATCH" \
  -exclude-tags "admin,internal"

# Whitelist mode - only allow user operations
./mcp-swagger-server -swagger api.json \
  -include-only-paths "/users,/users/{id}" \
  -include-only-operations "getUsers,getUser,createUser"
```

## How It Works

1. The server loads a Swagger/OpenAPI specification
2. API filtering rules are applied to determine which endpoints to expose
3. Each allowed API endpoint is converted to an MCP tool
4. Tool names are derived from the operation ID or the path
5. Parameters are converted to MCP tool input schemas
6. When a tool is called, the server makes the corresponding HTTP request
7. Response data is returned to the MCP client

## MCP Client Configuration

To use this server with an MCP client, configure it to run:

```json
{
  "servers": {
    "swagger-api": {
      "command": "./mcp-swagger-server",
      "args": ["-swagger", "path/to/your/api.json"]
    }
  }
}
```

## Testing

Run the test suite:

```bash
go test ./mcp -v
```

Run tests with race condition detection:

```bash
go test ./mcp -v -race
```

The project includes comprehensive unit tests covering:
- Configuration management and fluent API
- Swagger specification parsing and validation  
- Base URL inference and parameter handling
- Transport layer functionality
- Core library creation and validation logic
- API filtering and exclusion rules

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

## Dependencies

- [github.com/modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) v0.8.0
- [github.com/go-openapi/spec](https://github.com/go-openapi/spec)
- [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Author

**liliang-cn**

## Acknowledgments

- [Model Context Protocol](https://github.com/modelcontextprotocol) for the MCP SDK
- [OpenAPI Initiative](https://www.openapis.org/) for the OpenAPI Specification