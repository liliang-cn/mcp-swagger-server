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

	// Load Swagger spec
	var swaggerSpec []byte
	var err error

	if *swaggerFile != "" {
		swaggerSpec, err = os.ReadFile(*swaggerFile)
		if err != nil {
			log.Fatalf("Failed to read swagger file: %v", err)
		}
	} else if *swaggerURL != "" {
		swaggerSpec, err = mcp.FetchSwaggerFromURL(*swaggerURL)
		if err != nil {
			log.Fatalf("Failed to fetch swagger from URL: %v", err)
		}
	}

	// Parse Swagger spec
	swagger, err := mcp.ParseSwaggerSpec(swaggerSpec)
	if err != nil {
		log.Fatalf("Failed to parse swagger spec: %v", err)
	}

	// Determine base URL
	baseURL := *apiBaseURL
	if baseURL == "" {
		// Try to extract from swagger spec
		if swagger.Host != "" {
			scheme := "https"
			if len(swagger.Schemes) > 0 {
				scheme = swagger.Schemes[0]
			}
			baseURL = fmt.Sprintf("%s://%s%s", scheme, swagger.Host, swagger.BasePath)
		} else {
			log.Fatal("API base URL not specified and cannot be determined from spec")
		}
	}

	// Create and run MCP server
	server := mcp.NewSwaggerMCPServer(baseURL, swagger, *apiKey)
	
	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}