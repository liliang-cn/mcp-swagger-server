# MCP Swagger Server Examples

This directory contains comprehensive examples demonstrating how to use the MCP Swagger Server with various transport modes and configurations.

## 📁 Directory Structure

```
examples/
├── server/                    # Local API server for testing
│   ├── main.go               # Simple petstore API server
│   ├── swagger.json          # Swagger definition
│   └── start_server.sh       # Server startup script
├── 01_basic_stdio/           # Basic stdio transport example
│   └── main.go
├── 02_http_transport/        # HTTP transport example
│   └── main.go
├── 03_api_filtering/         # API filtering example
│   └── main.go
├── 04_advanced_usage/        # Advanced usage patterns
│   └── main.go
├── 05_testing_demo/          # Comprehensive testing demo
│   └── main.go
└── README.md                 # This file
```

## 🚀 Quick Start

### 1. Start the Local API Server

First, start the local petstore API server that all examples will connect to:

```bash
cd examples/server
./start_server.sh
```

The server will start on `http://localhost:8080` with the following endpoints:
- `GET /health` - Health check
- `GET /v2/pets` - List pets
- `POST /v2/pets` - Create pet
- `GET /v2/pets/{id}` - Get pet by ID
- `PUT /v2/pets/{id}` - Update pet
- `DELETE /v2/pets/{id}` - Delete pet
- `POST /v2/pets/search` - Search pets

### 2. Run Examples

Each example demonstrates different aspects of the MCP Swagger Server:

#### Example 1: Basic Stdio Transport
```bash
cd examples/01_basic_stdio
go run main.go
```

Shows basic usage with stdio transport, ideal for CLI tools and MCP clients.

#### Example 2: HTTP Transport
```bash
cd examples/02_http_transport
go run main.go
```

Demonstrates HTTP transport with REST API endpoints, ideal for web applications.

#### Example 3: API Filtering
```bash
cd examples/03_api_filtering
go run main.go
```

Shows how to filter which API operations are exposed as MCP tools.

#### Example 4: Advanced Usage
```bash
cd examples/04_advanced_usage
go run main.go
```

Demonstrates advanced configuration, error handling, and tool inspection.

#### Example 5: Testing Demo
```bash
cd examples/05_testing_demo
go run main.go
```

Comprehensive testing of all MCP server functionality with detailed results.

## 📋 Examples Overview

### 1️⃣ Basic Stdio Transport (`01_basic_stdio`)

**Purpose**: Learn the fundamentals of MCP server setup with stdio transport

**Features**:
- Basic server configuration
- Loading Swagger from local file
- Stdio transport setup
- Tool listing and information

**Use Cases**:
- CLI tool integration
- MCP client integration (Claude Desktop, etc.)
- Session-based communication

### 2️⃣ HTTP Transport (`02_http_transport`)

**Purpose**: Learn how to use HTTP transport for web applications

**Features**:
- HTTP transport configuration
- REST API endpoints (`/health`, `/tools`, `/mcp`)
- Endpoint testing
- HTTP client examples

**Use Cases**:
- Web applications
- HTTP-based clients
- Multi-client scenarios
- Stateless request/response

### 3️⃣ API Filtering (`03_api_filtering`)

**Purpose**: Learn how to control which API operations are exposed

**Features**:
- Include/exclude HTTP methods
- Tag-based filtering
- Combined filtering strategies
- Security considerations

**Use Cases**:
- Security hardening
- Client-specific toolsets
- Read-only access patterns
- Simplified interfaces

### 4️⃣ Advanced Usage (`04_advanced_usage`)

**Purpose**: Learn advanced patterns and best practices

**Features**:
- Custom server configuration
- Tool schema inspection
- Error handling patterns
- MCP protocol testing
- Performance considerations

**Use Cases**:
- Production deployments
- Custom integrations
- Debugging and monitoring
- Complex workflows

### 5️⃣ Testing Demo (`05_testing_demo`)

**Purpose**: Learn how to test MCP server functionality

**Features**:
- Comprehensive test suite
- Performance testing
- Error scenario testing
- Test result reporting
- Automated validation

**Use Cases**:
- CI/CD integration
- Quality assurance
- Regression testing
- Performance monitoring

## 🔧 Configuration Options

All examples support common configuration patterns:

### Basic Configuration
```go
config := mcp.DefaultConfig().
    WithServerInfo("name", "version", "description").
    WithAPIConfig("baseURL", "apiKey").
    WithSwaggerData(swaggerBytes)
```

### Transport Configuration
```go
// Stdio (default)
server.RunStdio(ctx)

// HTTP transport
config.WithHTTPTransport(port, host, path)
server.Run(ctx)
```

### Filtering Configuration
```go
// Exclude methods
config.WithExcludeMethods("DELETE", "PUT")

// Include only specific tags
config.WithIncludeTags("pets", "search")

// Exclude specific tags
config.WithExcludeTags("admin", "internal")
```

## 🛠️ Requirements

- Go 1.19 or higher
- Access to `github.com/liliang-cn/mcp-swagger-server`
- For HTTP examples: ability to make HTTP requests
- Port availability (examples use ports 7777-7780)

## 📝 Notes

- All examples use the local petstore API server for consistent testing
- Examples are numbered to suggest learning progression
- Each example can be run independently
- The local API server must be running before executing examples
- No external API keys or authentication required for local testing

## 🐛 Troubleshooting

**Server won't start**:
- Check if ports 8080 or 7777-7780 are already in use
- Ensure Go dependencies are installed
- Verify you're in the correct directory

**API calls fail**:
- Make sure the local API server is running
- Check server startup logs for errors
- Verify the API base URL is correct

**MCP tools not working**:
- Check Swagger file is correctly formatted
- Verify API endpoints are accessible
- Review server configuration logs

For more troubleshooting tips, see individual example files and their comments.