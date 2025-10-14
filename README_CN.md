# MCP Swagger Server 使用指南

> **📖 文档**: [中文文档](README_CN.md) | [English Documentation](README.md)

## 功能特性概览

MCP Swagger Server 提供了以下核心功能：

### 核心特性
- **双模式支持**: 既可作为独立CLI工具，也可作为Go库集成
- **多种传输方式**: 支持标准输入输出和HTTP传输，HTTP传输包含跨域资源共享支持
- **完整的接口规范支持**: 支持Swagger 2.0和OpenAPI规范（JSON和YAML格式）
- **自动转换**: 自动将应用程序接口端点转换为MCP工具，包含完整的模式生成
- **高级接口过滤**: 全面的过滤系统，精确控制哪些接口成为工具
- **完整超文本传输协议支持**: 支持所有HTTP方法（GET、POST、PUT、DELETE、PATCH）
- **智能参数处理**: 智能处理路径参数、查询参数和请求体
- **认证支持**: 自动应用程序接口密钥认证，支持多种头格式
- **网络集成**: 轻松集成到现有Go网络应用程序
- **超文本传输协议接口端点**: 内置HTTP端点用于工具列表和健康检查
- **错误处理**: 完善的错误处理和适当的HTTP状态码
- **JSON格式化**: 自动JSON响应格式化和验证
- **灵活配置**: 灵活的配置系统和流畅的应用程序接口

## 问题解答：MCP端点没有响应的原因

你遇到的问题是因为 **MCP 服务器有两种不同的传输方式 (transport)**：

1. **stdio transport** - 用于命令行工具和MCP客户端通信
2. **HTTP transport** - 用于HTTP API和web集成

## 两种传输方式的区别

### 1. stdio transport (默认)
- 用于 Claude Desktop、其他MCP客户端
- 通过标准输入/输出进行通信
- 不能通过HTTP访问

### 2. HTTP transport
- 提供HTTP API端点
- 可以用curl、浏览器、Postman等测试
- 适合web应用集成

## 如何使用HTTP transport

### 方法1: 使用配置创建

```go
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
```

### 方法2: 直接使用RunHTTP

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
```

### 方法3: 命令行启动HTTP服务器

现在命令行工具原生支持HTTP transport：

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

当使用HTTP transport时，服务器提供以下端点：

```bash
GET  /health          # 健康检查，返回状态信息
GET  /tools           # 获取可用工具列表，包含详细信息
POST /mcp             # 执行MCP请求（支持tools/list和tools/call）
OPTIONS /mcp          # CORS预检支持
```

所有HTTP端点都包含CORS头，支持跨域请求。

### 示例请求

#### 1. 健康检查
```bash
curl http://localhost:8127/health
# 响应: {"status":"ok"}
```

#### 2. 获取工具列表
```bash
curl http://localhost:8127/tools
# 响应: {"tools":[{工具信息}]}
```

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
```

## 完整示例

运行HTTP transport示例：

```bash
go run examples/http_server/main.go
```

这个示例会：
1. 启动两个HTTP服务器 (端口7777和8888)
2. 使用 JSONPlaceholder API 作为后端
3. 测试所有HTTP端点
4. 展示正确的使用方式

## API过滤在HTTP transport中的使用

HTTP transport也支持API过滤：

```go
config := mcp.DefaultConfig().
    WithSwaggerData(data).
    WithAPIConfig("https://api.example.com", "").
    WithHTTPTransport(6724, "", "").
    WithExcludePaths("/admin/*").
    WithExcludeMethods("DELETE", "PATCH")

server, _ := mcp.New(config)
```

过滤的API不会出现在 `/tools` 端点中，也无法通过 `/mcp` 调用。

## 选择合适的传输方式

### 使用 stdio transport 当：
- 与Claude Desktop集成
- 与其他MCP客户端集成
- 作为命令行工具使用

### 使用 HTTP transport 当：
- 构建web应用
- 需要HTTP API
- 与现有HTTP服务集成
- 进行开发和测试

## 故障排除

### 问题1: "MCP端点没有响应"
**原因**: 使用了stdio transport但试图通过HTTP访问  
**解决**: 使用HTTP transport或通过MCP客户端访问

### 问题2: "404 Not Found"
**原因**: 端点路径错误  
**解决**: 确保使用正确的端点 (`/health`, `/tools`, `/mcp`)

### 问题3: "Connection refused"
**原因**: 服务器未启动或端口错误  
**解决**: 确认服务器正在运行并使用正确端口

### 问题4: "工具调用失败"
**原因**: API过滤、认证问题或后端API不可达  
**解决**: 检查过滤配置、API密钥和网络连接

## 开发建议

1. **开发时使用HTTP transport** - 便于测试和调试
2. **生产时根据需求选择** - MCP客户端用stdio，web应用用HTTP
3. **使用API过滤增强安全性** - 避免暴露敏感端点
4. **监控健康检查端点** - 用于负载均衡和监控

## 下一步

1. 查看 `examples/http_server/main.go` 了解完整示例
2. 查看 `examples/api_filtering/main.go` 了解过滤功能
3. 阅读 README.md 了解所有功能
4. 根据你的需求选择合适的传输方式