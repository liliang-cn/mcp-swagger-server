package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	fmt.Println("=== HTTP Transport Example ===")

	// 创建一个示例 Swagger 规范
	swaggerData := []byte(`{
		"swagger": "2.0",
		"info": {
			"title": "Example API",
			"version": "1.0.0"
		},
		"host": "jsonplaceholder.typicode.com",
		"schemes": ["https"],
		"basePath": "",
		"paths": {
			"/posts": {
				"get": {
					"operationId": "getPosts",
					"summary": "获取所有帖子",
					"parameters": [
						{
							"name": "userId",
							"in": "query",
							"type": "integer",
							"description": "按用户ID过滤"
						}
					]
				}
			},
			"/posts/{id}": {
				"get": {
					"operationId": "getPost",
					"summary": "获取单个帖子",
					"parameters": [
						{
							"name": "id",
							"in": "path",
							"required": true,
							"type": "integer",
							"description": "帖子ID"
						}
					]
				}
			},
			"/users": {
				"get": {
					"operationId": "getUsers",
					"summary": "获取所有用户"
				}
			}
		}
	}`)

	// 方法1: 使用配置创建HTTP服务器
	fmt.Println("\n方法1: 使用配置创建HTTP服务器")
	config := mcp.DefaultConfig().
		WithSwaggerData(swaggerData).
		WithAPIConfig("https://jsonplaceholder.typicode.com", ""). // 使用 JSONPlaceholder 作为测试API
		WithHTTPTransport(7777, "localhost", "/mcp")

	server, err := mcp.New(config)
	if err != nil {
		log.Fatalf("创建服务器失败: %v", err)
	}

	// 在后台启动HTTP服务器
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		fmt.Printf("正在启动HTTP服务器，地址: http://localhost:7777\n")
		if err := server.Run(ctx); err != nil {
			log.Printf("HTTP服务器错误: %v", err)
		}
	}()

	// 方法2: 直接使用RunHTTP
	fmt.Println("\n方法2: 直接使用RunHTTP (在端口8888)")
	server2, err := mcp.NewFromSwaggerData(swaggerData, "https://jsonplaceholder.typicode.com", "")
	if err != nil {
		log.Fatalf("创建服务器2失败: %v", err)
	}

	go func() {
		fmt.Printf("正在启动HTTP服务器2，地址: http://localhost:8888\n")
		if err := server2.RunHTTP(ctx, 8888); err != nil {
			log.Printf("HTTP服务器2错误: %v", err)
		}
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试HTTP端点
	fmt.Println("\n=== 测试HTTP端点 ===")
	
	// 测试健康检查
	testEndpoint("GET", "http://localhost:7777/health", "健康检查")
	testEndpoint("GET", "http://localhost:8888/health", "健康检查 (服务器2)")

	// 测试工具列表
	testEndpoint("GET", "http://localhost:7777/tools", "工具列表")

	// 测试MCP工具调用
	testMCPCall("http://localhost:7777/mcp", "getPosts", map[string]interface{}{
		"userId": 1,
	})

	testMCPCall("http://localhost:7777/mcp", "getPost", map[string]interface{}{
		"id": 1,
	})

	testMCPCall("http://localhost:7777/mcp", "getUsers", map[string]interface{}{})

	fmt.Println("\n=== HTTP Transport 说明 ===")
	fmt.Println("1. 使用 WithHTTPTransport() 配置HTTP传输")
	fmt.Println("2. 或者直接使用 server.RunHTTP(ctx, port)")
	fmt.Println("3. 访问端点:")
	fmt.Println("   - GET /health - 健康检查")
	fmt.Println("   - GET /tools - 获取可用工具列表")
	fmt.Println("   - POST /mcp - 执行MCP请求")
	fmt.Println("\n4. MCP请求格式:")
	fmt.Println(`   {
     "method": "tools/call",
     "params": {
       "name": "工具名称",
       "arguments": {参数}
     }
   }`)
	
	// 保持服务器运行一段时间以便测试
	fmt.Println("\n服务器将继续运行30秒，你可以在浏览器中测试端点...")
	time.Sleep(30 * time.Second)
	
	fmt.Println("示例结束")
}

// testEndpoint 测试HTTP端点
func testEndpoint(method, url, description string) {
	fmt.Printf("\n测试 %s: %s %s\n", description, method, url)
	
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("  ❌ 错误: %v\n", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("  ⚠️ 关闭响应体失败: %v\n", err)
		}
	}()
	
	fmt.Printf("  ✅ 状态码: %d\n", resp.StatusCode)
}

// testMCPCall 测试MCP工具调用
func testMCPCall(url, toolName string, arguments map[string]interface{}) {
	fmt.Printf("\n测试MCP调用: %s\n", toolName)
	
	payload := map[string]interface{}{
		"method": "tools/call",
		"params": map[string]interface{}{
			"name":      toolName,
			"arguments": arguments,
		},
	}
	
	client := &http.Client{Timeout: 10 * time.Second}
	
	// 这里只是测试请求格式，实际调用需要JSON body
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Printf("  ❌ 创建请求失败: %v\n", err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("  ❌ 请求失败: %v\n", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Printf("  ⚠️ 关闭响应体失败: %v\n", err)
		}
	}()
	
	fmt.Printf("  ✅ MCP端点响应状态码: %d\n", resp.StatusCode)
	fmt.Printf("  📝 请求格式: %+v\n", payload)
}