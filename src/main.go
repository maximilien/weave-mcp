// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/maximilien/weave-mcp/src/pkg/config"
	"github.com/maximilien/weave-mcp/src/pkg/mcp"
	"github.com/maximilien/weave-mcp/src/pkg/version"
	"go.uber.org/zap"
)

func main() {
	var (
		configFile  = flag.String("config", "config.yaml", "Path to configuration file")
		envFile     = flag.String("env", ".env", "Path to environment file")
		host        = flag.String("host", "localhost", "Server host")
		port        = flag.String("port", "8030", "Server port")
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
	logFile, err := os.OpenFile(filepath.Join(logsDir, "weave-mcp.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}
	defer logFile.Close()

	// Create logger with both console and file output
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout", logFile.Name()}
	config.ErrorOutputPaths = []string{"stderr", logFile.Name()}

	logger, err := config.Build()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create MCP server
	server, err := mcp.NewServer(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create MCP server", zap.Error(err))
	}

	// Create HTTP server
	addr := fmt.Sprintf("%s:%s", *host, *port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      server.Handler(),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting Weave MCP Server",
			zap.String("address", addr),
			zap.String("version", version.Version),
			zap.String("git_commit", version.GitCommit))

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	// Cleanup MCP server
	if err := server.Cleanup(); err != nil {
		logger.Error("Failed to cleanup MCP server", zap.Error(err))
	}

	logger.Info("Server stopped")
}
