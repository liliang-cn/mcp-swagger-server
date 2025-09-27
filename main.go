package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	var (
		swaggerFile = flag.String("swagger", "", "Path to Swagger/OpenAPI spec file (JSON or YAML)")
		swaggerURL  = flag.String("swagger-url", "", "URL to fetch Swagger/OpenAPI spec")
		apiBaseURL  = flag.String("api-base", "", "Base URL for API calls (overrides spec)")
		apiKey      = flag.String("api-key", "", "API key for authentication")
	)

	flag.Parse()

	// Validate inputs
	if *swaggerFile == "" && *swaggerURL == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -swagger <file> | -swagger-url <url> [-api-base <url>] [-api-key <key>]\n", os.Args[0])
		os.Exit(1)
	}

	// Create MCP server using the new library interface
	var server *mcp.Server
	var err error

	if *swaggerFile != "" {
		server, err = mcp.NewFromSwaggerFile(*swaggerFile, *apiBaseURL, *apiKey)
		if err != nil {
			log.Fatalf("Failed to create server from swagger file: %v", err)
		}
	} else if *swaggerURL != "" {
		server, err = mcp.NewFromSwaggerURL(*swaggerURL, *apiBaseURL, *apiKey)
		if err != nil {
			log.Fatalf("Failed to create server from swagger URL: %v", err)
		}
	}

	// Run the server with stdio transport (for CLI usage)
	ctx := context.Background()
	if err := server.RunStdio(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}