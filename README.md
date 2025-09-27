# MCP Swagger Server

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Test Coverage](https://img.shields.io/badge/Coverage-88.8%25-brightgreen)](https://github.com/liliang-cn/mcp-swagger-server)
[![Go Report Card](https://goreportcard.com/badge/github.com/liliang-cn/mcp-swagger-server)](https://goreportcard.com/report/github.com/liliang-cn/mcp-swagger-server)

A Model Context Protocol (MCP) server that converts Swagger/OpenAPI specifications into MCP tools.

## Features

- Loads Swagger 2.0 and OpenAPI specifications (JSON or YAML)
- Automatically converts API endpoints to MCP tools
- Supports all HTTP methods (GET, POST, PUT, DELETE, PATCH)
- Handles path parameters, query parameters, and request bodies
- Automatic API key authentication support

## Installation

```bash
go get github.com/liliang-cn/mcp-swagger-server
```

Or clone and build from source:

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

### From a local Swagger file:

```bash
./mcp-swagger-server -swagger examples/petstore.json
```

### From a URL:

```bash
./mcp-swagger-server -swagger-url https://petstore.swagger.io/v2/swagger.json
```

### With custom API base URL:

```bash
./mcp-swagger-server -swagger examples/api.yaml -api-base https://api.example.com
```

### With API key authentication:

```bash
./mcp-swagger-server -swagger examples/api.json -api-key YOUR_API_KEY
```

## Command Line Options

- `-swagger` - Path to local Swagger/OpenAPI spec file (JSON or YAML)
- `-swagger-url` - URL to fetch Swagger/OpenAPI spec from
- `-api-base` - Override the base URL for API calls (defaults to spec's host)
- `-api-key` - API key for authentication

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

Run tests with coverage:

```bash
go test ./... -v -cover
```

Current test coverage: **88.8%**

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