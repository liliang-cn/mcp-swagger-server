# MCP Swagger Server

将 Swagger/OpenAPI 规范转换为 MCP 工具，支持 CLI 和 Go 库两种使用方式。

> **📖 文档**: [中文文档](README_CN.md) | [English Documentation](README.md)

## 🚀 快速开始

### 1. 启动测试 API（终端 1）

```bash
cd examples/server && go run main.go
# 运行在 http://localhost:4538
```

### 2. 启动 MCP 服务器（终端 2）

```bash
# HTTP 模式（测试用）
./mcp-swagger-server \
  -swagger examples/server/swagger.json \
  -api-base http://localhost:4538/v2 \
  -http-port 4539

# 测试
curl http://localhost:4539/mcp/health
```

### 3. Claude Desktop 配置

```json
{
  "mcpServers": {
    "local-petstore": {
      "command": "/path/to/mcp-swagger-server",
      "args": [
        "-swagger",
        "/path/to/swagger.json",
        "-api-base",
        "http://localhost:4538/v2"
      ]
    }
  }
}
```

**重要**: API base URL 必须包含 Swagger 的 `basePath`（如 `/v2`）

## 功能特性

- **双模式**: CLI 工具 + Go 库
- **双传输**: stdio（Claude Desktop）+ HTTP（测试/Web）
- **完整支持**: Swagger 2.0/OpenAPI 规范
- **API 过滤**: 路径/方法/标签过滤
- **认证**: API Key 支持

## 命令行使用

```bash
# Stdio 模式（Claude Desktop）
./mcp-swagger-server -swagger api.json -api-base https://api.example.com

# HTTP 模式（测试）
./mcp-swagger-server -swagger api.json -api-base https://api.example.com -http-port 4539

# API 过滤
./mcp-swagger-server -swagger api.json \
  -exclude-methods "DELETE,PATCH" \
  -exclude-paths "/admin/*"
```

## HTTP 端点

所有端点在 `/mcp` 路径下：

- `GET /mcp/health` - 健康检查
- `GET /mcp/tools` - 工具列表
- `POST /mcp` - MCP 协议

## Go 库使用

```go
server, _ := mcp.NewFromSwaggerFile("api.json", "https://api.example.com", "api-key")
server.RunStdio(context.Background())
```

## 常见问题

**Q: 404 错误**  
A: 检查 API base URL 是否包含 Swagger 的 `basePath`

**Q: 连接失败**  
A: 确保 API 服务器运行中

**Q: 工具列表为空**  
A: 检查 Swagger 文件路径和格式
package main

import (
"context"
"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
// 创建配置
config := mcp.DefaultConfig().
WithSwaggerFile("api.json").
WithAPIConfig("https://api.example.com", "your-api-key").
WithHTTPTransport(8127, "localhost", "/mcp")

    // 创建服务器
    server, err := mcp.New(config)
    if err != nil {
        panic(err)
    }

    // 启动服务器 (自动使用HTTP transport)
    ctx := context.Background()
    server.Run(ctx)

}

````

### 方法 2: 直接使用 RunHTTP

```go
package main

import (
    "context"
    "github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
    // 创建服务器
    server, err := mcp.NewFromSwaggerFile("api.json", "https://api.example.com", "")
    if err != nil {
        panic(err)
    }

    // 直接启动HTTP服务器
    ctx := context.Background()
    server.RunHTTP(ctx, 6724)
}
````

### 方法 3: 命令行启动 HTTP 服务器

现在命令行工具原生支持 HTTP transport：

```bash
# 启动HTTP服务器
./mcp-swagger-server -swagger api.json -http-port 8127

# 带过滤的HTTP服务器
./mcp-swagger-server -swagger api.json \
  -http-port 8127 \
  -exclude-methods "DELETE,PATCH" \
  -exclude-paths "/admin/*"

# 自定义主机和路径
./mcp-swagger-server -swagger api.json \
  -http-port 8127 \
  -http-host 0.0.0.0 \
  -http-path /api/mcp
```

## HTTP API 端点

当使用 HTTP transport 时，服务器提供以下端点（所有端点都在配置的路径下，默认 `/mcp`）：

```bash
GET  /mcp/health      # 健康检查，返回服务器状态信息
GET  /mcp/tools       # 获取可用工具列表，包含详细信息
POST /mcp             # 执行MCP请求（支持tools/list和tools/call）
OPTIONS /mcp          # CORS预检支持
```

所有 HTTP 端点都包含 CORS 头，支持跨域请求。

### 示例请求

#### 1. 健康检查

```bash
curl http://localhost:4539/mcp/health
# 响应: {"status":"ok","server":"swagger-mcp-server","version":"v1.0.0"}
```

#### 2. 获取工具列表

```bash
curl http://localhost:4539/mcp/tools
# 响应: {"tools":[{工具信息}]}
```

#### 3. 调用工具

```bash
curl -X POST http://localhost:4539/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "method": "tools/call",
    "params": {
      "name": "listpets",
      "arguments": {"limit": 5}
    },
    "id": 1
  }'
```

**重要提示：**

- 如果 Swagger 文件中定义了 `basePath`（如 `/v2`），API base URL 必须包含它
- 示例：`-api-base http://localhost:4538/v2` 而不是 `http://localhost:4538`
  curl http://localhost:8127/tools

# 响应: {"tools":[{工具信息}]}

````

#### 3. 调用工具
```bash
curl -X POST http://localhost:8127/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "method": "tools/call",
    "params": {
      "name": "getPosts",
      "arguments": {
        "userId": 1
      }
    }
  }'
````

## 完整示例

运行 HTTP transport 示例：

```bash
go run examples/http_server/main.go
```

这个示例会：

1. 启动两个 HTTP 服务器 (端口 7777 和 8888)
2. 使用 JSONPlaceholder API 作为后端
3. 测试所有 HTTP 端点
4. 展示正确的使用方式

## API 过滤在 HTTP transport 中的使用

HTTP transport 也支持 API 过滤：

```go
config := mcp.DefaultConfig().
    WithSwaggerData(data).
    WithAPIConfig("https://api.example.com", "").
    WithHTTPTransport(6724, "", "").
    WithExcludePaths("/admin/*").
    WithExcludeMethods("DELETE", "PATCH")

server, _ := mcp.New(config)
```

过滤的 API 不会出现在 `/tools` 端点中，也无法通过 `/mcp` 调用。

## 选择合适的传输方式

### 使用 stdio transport 当：

- 与 Claude Desktop 集成
- 与其他 MCP 客户端集成
- 作为命令行工具使用

### 使用 HTTP transport 当：

- 构建 web 应用
- 需要 HTTP API
- 与现有 HTTP 服务集成
- 进行开发和测试

## 故障排除

### 问题 1: "MCP 端点没有响应"

**原因**: 使用了 stdio transport 但试图通过 HTTP 访问  
**解决**: 使用 HTTP transport 或通过 MCP 客户端访问

### 问题 2: "404 Not Found"

**原因**: 端点路径错误  
**解决**: 确保使用正确的端点 (`/health`, `/tools`, `/mcp`)

### 问题 3: "Connection refused"

**原因**: 服务器未启动或端口错误  
**解决**: 确认服务器正在运行并使用正确端口

### 问题 4: "工具调用失败"

**原因**: API 过滤、认证问题或后端 API 不可达  
**解决**: 检查过滤配置、API 密钥和网络连接

## 开发建议

1. **开发时使用 HTTP transport** - 便于测试和调试
2. **生产时根据需求选择** - MCP 客户端用 stdio，web 应用用 HTTP
3. **使用 API 过滤增强安全性** - 避免暴露敏感端点
4. **监控健康检查端点** - 用于负载均衡和监控

## 下一步

1. 查看 `examples/http_server/main.go` 了解完整示例
2. 查看 `examples/api_filtering/main.go` 了解过滤功能
3. 阅读 README.md 了解所有功能
4. 根据你的需求选择合适的传输方式
