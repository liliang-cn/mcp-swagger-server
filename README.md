# MCP Swagger Server

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat&logo=go)](https://go.dev)
[![Go Reference](https://pkg.go.dev/badge/github.com/liliang-cn/mcp-swagger-server.svg)](https://pkg.go.dev/github.com/liliang-cn/mcp-swagger-server)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/liliang-cn/mcp-swagger-server)](https://goreportcard.com/report/github.com/liliang-cn/mcp-swagger-server)

A Model Context Protocol (MCP) server that converts Swagger/OpenAPI specifications into MCP tools. This project can be used both as a standalone CLI tool and as a Go library for integration into other projects.

> **Status**: ✅ Production Ready - The project is stable, well-tested, and ready for production use.

## Features

- **Dual Usage**: Works as standalone CLI and Go library
- **Multiple Transport**: Supports stdio and HTTP transport
- **Complete Swagger Support**: Loads Swagger 2.0 and OpenAPI specifications (JSON or YAML)
- **Auto-conversion**: Automatically converts API endpoints to MCP tools
- **Full HTTP Support**: Supports all HTTP methods (GET, POST, PUT, DELETE, PATCH)
- **Parameter Handling**: Handles path parameters, query parameters, and request bodies
- **Authentication**: Automatic API key authentication support
- **Web Integration**: Easy integration into existing Go web applications

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
    mcpServer, _ := mcp.NewFromSwaggerFile("your-api.json", "http://localhost:6666", "")
    
    // Option 1: Run MCP server on separate HTTP port
    go mcpServer.RunHTTP(context.Background(), 7777)
    
    // Option 2: Integrate into existing server (custom implementation needed)
    
    // Your existing routes
    router.HandleFunc("/api/users", handleUsers)
    
    http.ListenAndServe(":6666", router)
}
```

#### Advanced Configuration

```go
config := mcp.DefaultConfig().
    WithAPIConfig("https://api.example.com", "your-api-key").
    WithServerInfo("my-api-server", "v1.0.0", "Custom API MCP Server").
    WithHTTPTransport(7777, "localhost", "/mcp")

server, err := mcp.New(config)
if err != nil {
    log.Fatal(err)
}

// Run with HTTP transport
server.Run(context.Background())
```

## Command Line Options

- `-swagger` - Path to local Swagger/OpenAPI spec file (JSON or YAML)
- `-swagger-url` - URL to fetch Swagger/OpenAPI spec from
- `-api-base` - Override the base URL for API calls (defaults to spec's host)
- `-api-key` - API key for authentication

## HTTP API Endpoints

When running with HTTP transport, the server exposes the following endpoints:

- `GET /health` - Health check endpoint
- `GET /tools` - List available tools
- `POST /mcp` - Execute MCP requests

### Example HTTP Usage

```bash
# Get available tools
curl http://localhost:7777/tools

# Execute a tool
curl -X POST http://localhost:7777/mcp \
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

## Example Swagger Spec

The `examples/petstore.json` file contains a sample Swagger specification for testing.

## How It Works

1. The server loads a Swagger/OpenAPI specification
2. Each API endpoint is converted to an MCP tool
3. Tool names are derived from the operation ID or the path
4. Parameters are converted to MCP tool input schemas
5. When a tool is called, the server makes the corresponding HTTP request
6. Response data is returned to the MCP client

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