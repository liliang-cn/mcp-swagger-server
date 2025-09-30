package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/liliang-cn/mcp-swagger-server/mcp"
)

func main() {
	var (
		swaggerFile         = flag.String("swagger", "", "Path to Swagger/OpenAPI spec file (JSON or YAML)")
		swaggerURL          = flag.String("swagger-url", "", "URL to fetch Swagger/OpenAPI spec")
		apiBaseURL          = flag.String("api-base", "", "Base URL for API calls (overrides spec)")
		apiKey              = flag.String("api-key", "", "API key for authentication")
		excludePaths        = flag.String("exclude-paths", "", "Comma-separated list of paths to exclude (e.g., '/users,/admin/*')")
		excludeOperationIDs = flag.String("exclude-operations", "", "Comma-separated list of operation IDs to exclude")
		excludeMethods      = flag.String("exclude-methods", "", "Comma-separated list of HTTP methods to exclude (e.g., 'DELETE,PATCH')")
		excludeTags         = flag.String("exclude-tags", "", "Comma-separated list of tags to exclude")
		includeOnlyPaths    = flag.String("include-only-paths", "", "Comma-separated list of paths to include exclusively")
		includeOnlyOps      = flag.String("include-only-operations", "", "Comma-separated list of operation IDs to include exclusively")
		httpPort            = flag.Int("http-port", 0, "HTTP server port (0 = disabled, use stdio transport)")
		httpHost            = flag.String("http-host", "localhost", "HTTP server host")
		httpPath            = flag.String("http-path", "/mcp", "HTTP server path for MCP endpoint")
	)

	flag.Parse()

	// Validate inputs
	if *swaggerFile == "" && *swaggerURL == "" {
		fmt.Fprintf(os.Stderr, "Usage: %s -swagger <file> | -swagger-url <url> [-api-base <url>] [-api-key <key>] [transport options] [filtering options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nTransport options:\n")
		fmt.Fprintf(os.Stderr, "  -http-port: HTTP server port (default: 0 = use stdio)\n")
		fmt.Fprintf(os.Stderr, "  -http-host: HTTP server host (default: localhost)\n")
		fmt.Fprintf(os.Stderr, "  -http-path: HTTP server path (default: /mcp)\n")
		fmt.Fprintf(os.Stderr, "\nFiltering options:\n")
		fmt.Fprintf(os.Stderr, "  -exclude-paths: Comma-separated paths to exclude (supports wildcards)\n")
		fmt.Fprintf(os.Stderr, "  -exclude-operations: Comma-separated operation IDs to exclude\n")
		fmt.Fprintf(os.Stderr, "  -exclude-methods: Comma-separated HTTP methods to exclude\n")
		fmt.Fprintf(os.Stderr, "  -exclude-tags: Comma-separated tags to exclude\n")
		fmt.Fprintf(os.Stderr, "  -include-only-paths: Include only these paths (exclusive)\n")
		fmt.Fprintf(os.Stderr, "  -include-only-operations: Include only these operation IDs (exclusive)\n")
		os.Exit(1)
	}

	// Build API filter configuration
	var filter *mcp.APIFilter
	if *excludePaths != "" || *excludeOperationIDs != "" || *excludeMethods != "" || *excludeTags != "" || 
	   *includeOnlyPaths != "" || *includeOnlyOps != "" {
		filter = &mcp.APIFilter{}
		
		if *excludePaths != "" {
			// Split exclude paths and handle patterns
			paths := strings.Split(*excludePaths, ",")
			for i, path := range paths {
				paths[i] = strings.TrimSpace(path)
			}
			// Separate exact paths from patterns
			for _, path := range paths {
				if strings.Contains(path, "*") {
					filter.ExcludePathPatterns = append(filter.ExcludePathPatterns, path)
				} else {
					filter.ExcludePaths = append(filter.ExcludePaths, path)
				}
			}
		}
		
		if *excludeOperationIDs != "" {
			ops := strings.Split(*excludeOperationIDs, ",")
			for i, op := range ops {
				ops[i] = strings.TrimSpace(op)
			}
			filter.ExcludeOperationIDs = ops
		}
		
		if *excludeMethods != "" {
			methods := strings.Split(*excludeMethods, ",")
			for i, method := range methods {
				methods[i] = strings.TrimSpace(strings.ToUpper(method))
			}
			filter.ExcludeMethods = methods
		}
		
		if *excludeTags != "" {
			tags := strings.Split(*excludeTags, ",")
			for i, tag := range tags {
				tags[i] = strings.TrimSpace(tag)
			}
			filter.ExcludeTags = tags
		}
		
		if *includeOnlyPaths != "" {
			paths := strings.Split(*includeOnlyPaths, ",")
			for i, path := range paths {
				paths[i] = strings.TrimSpace(path)
			}
			filter.IncludeOnlyPaths = paths
		}
		
		if *includeOnlyOps != "" {
			ops := strings.Split(*includeOnlyOps, ",")
			for i, op := range ops {
				ops[i] = strings.TrimSpace(op)
			}
			filter.IncludeOnlyOperationIDs = ops
		}
	}

	// Create MCP server using the new library interface with filtering
	var server *mcp.Server

	if *swaggerFile != "" {
		// Create with config to support filtering
		config := mcp.DefaultConfig().
			WithAPIConfig(*apiBaseURL, *apiKey).
			WithAPIFilter(filter)
		
		data, err := readSwaggerFile(*swaggerFile)
		if err != nil {
			log.Fatalf("Failed to read swagger file: %v", err)
		}
		config.WithSwaggerData(data)
		
		server, err = mcp.New(config)
		if err != nil {
			log.Fatalf("Failed to create server from swagger file: %v", err)
		}
	} else if *swaggerURL != "" {
		// Create with config to support filtering
		config := mcp.DefaultConfig().
			WithAPIConfig(*apiBaseURL, *apiKey).
			WithAPIFilter(filter)
		
		data, err := mcp.FetchSwaggerFromURL(*swaggerURL)
		if err != nil {
			log.Fatalf("Failed to fetch swagger from URL: %v", err)
		}
		config.WithSwaggerData(data)
		
		server, err = mcp.New(config)
		if err != nil {
			log.Fatalf("Failed to create server from swagger URL: %v", err)
		}
	}

	// Run the server with appropriate transport
	ctx := context.Background()
	
	if *httpPort > 0 {
		// Use HTTP transport
		config := server.GetConfig()
		config.WithHTTPTransport(*httpPort, *httpHost, *httpPath)
		log.Printf("Starting MCP server with HTTP transport on %s:%d%s", *httpHost, *httpPort, *httpPath)
		if err := server.RunHTTP(ctx, *httpPort); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		// Use stdio transport (default for CLI usage)
		log.Println("Starting MCP server with stdio transport")
		if err := server.RunStdio(ctx); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

// readSwaggerFile reads a swagger file from disk
func readSwaggerFile(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}