// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/maximilien/weave-mcp/src/pkg/config"
	internalmcp "github.com/maximilien/weave-mcp/src/pkg/mcp"
	"github.com/maximilien/weave-mcp/src/pkg/version"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"

	// Import vectordb implementations to register their factories
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/mock"
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/supabase"
	_ "github.com/maximilien/weave-cli/src/pkg/vectordb/weaviate"
)

func main() {
	var (
		configFile  = flag.String("config", "", "Path to configuration file (default: auto-detect from local or ~/.weave-cli)")
		envFile     = flag.String("env", "", "Path to environment file (default: auto-detect from local or ~/.weave-cli)")
		showVersion = flag.Bool("version", false, "Show version information")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println(version.String())
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.LoadConfig(*configFile, *envFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create logs directory
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		log.Fatalf("Failed to create logs directory: %v", err)
	}

	// Create file logger
	logFile, err := os.OpenFile(filepath.Join(logsDir, "weave-mcp-stdio.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Create logger with stderr and file output (NOT stdout - that's for JSON-RPC!)
	zapConfig := zap.NewProductionConfig()
	zapConfig.OutputPaths = []string{"stderr", logFile.Name()}
	zapConfig.ErrorOutputPaths = []string{"stderr", logFile.Name()}

	logger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	// Create MCP server using our existing implementation
	internalServer, err := internalmcp.NewServer(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create internal MCP server", zap.Error(err))
	}

	// Create stdio MCP server
	stdioServer := mcp.NewServer(&mcp.Implementation{
		Name:    "weave-mcp",
		Version: version.Version,
	}, nil)

	// Register tools from our MCP server
	registerToolsFromMCP(stdioServer, internalServer, logger)

	// Run stdio server
	if err := stdioServer.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		logger.Fatal("Failed to run stdio server", zap.Error(err))
	}
}

// registerToolsFromMCP registers tools from our existing MCP server to the stdio server
func registerToolsFromMCP(stdioServer *mcp.Server, internalServer *internalmcp.Server, logger *zap.Logger) {
	// Get all tools from our MCP server
	for toolName, tool := range internalServer.Tools {
		// Create a wrapper function that calls our MCP server's tool handler
		handler := func(toolName string, tool internalmcp.Tool) func(ctx context.Context, req *mcp.CallToolRequest, args map[string]interface{}) (*mcp.CallToolResult, any, error) {
			return func(ctx context.Context, req *mcp.CallToolRequest, args map[string]interface{}) (*mcp.CallToolResult, any, error) {
				// Call the tool handler from our MCP server
				result, err := tool.Handler(ctx, args)
				if err != nil {
					logger.Error("Tool execution failed",
						zap.String("tool", toolName),
						zap.Error(err))
					return &mcp.CallToolResult{
						Content: []mcp.Content{
							&mcp.TextContent{Text: fmt.Sprintf("Error: %v", err)},
						},
					}, nil, err
				}

				// Convert result to JSON string for content
				resultJSON, err := json.MarshalIndent(result, "", "  ")
				if err != nil {
					logger.Error("Failed to marshal result",
						zap.String("tool", toolName),
						zap.Error(err))
					resultJSON = []byte(fmt.Sprintf("%+v", result))
				}

				return &mcp.CallToolResult{
					Content: []mcp.Content{
						&mcp.TextContent{Text: string(resultJSON)},
					},
				}, result, nil
			}
		}(toolName, tool)

		// Register the tool with the stdio server
		mcp.AddTool(stdioServer, &mcp.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			InputSchema: tool.InputSchema,
		}, handler)

		logger.Debug("Registered stdio tool", zap.String("name", toolName))
	}
}
