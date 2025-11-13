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
	"strings"
	"syscall"
	"time"

	"github.com/maximilien/weave-mcp/src/pkg/config"
	"github.com/maximilien/weave-mcp/src/pkg/mcp"
	"github.com/maximilien/weave-mcp/src/pkg/version"
	"go.uber.org/zap"
)

func main() {
	var (
		configFile  = flag.String("config", "", "Path to configuration file (default: auto-detect from local or ~/.weave-cli)")
		envFile     = flag.String("env", "", "Path to environment file (default: auto-detect from local or ~/.weave-cli)")
		host        = flag.String("host", "localhost", "Server host")
		port        = flag.String("port", "8030", "Server port")
		corsOrigins = flag.String("cors-origins", "*", "Comma-separated list of allowed CORS origins")
		corsMethods = flag.String("cors-methods", "GET,POST,PUT,DELETE,OPTIONS", "Comma-separated list of allowed CORS methods")
		corsHeaders = flag.String("cors-headers", "Content-Type,Authorization,X-Requested-With", "Comma-separated list of allowed CORS headers")
		corsMaxAge  = flag.Int("cors-max-age", 86400, "CORS preflight cache max age in seconds")
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

	// Parse CORS configuration with environment variable fallbacks
	corsOriginsEnv := os.Getenv("CORS_ORIGINS")
	if corsOriginsEnv != "" {
		*corsOrigins = corsOriginsEnv
	}

	corsMethodsEnv := os.Getenv("CORS_METHODS")
	if corsMethodsEnv != "" {
		*corsMethods = corsMethodsEnv
	}

	corsHeadersEnv := os.Getenv("CORS_HEADERS")
	if corsHeadersEnv != "" {
		*corsHeaders = corsHeadersEnv
	}

	corsMaxAgeEnv := os.Getenv("CORS_MAX_AGE")
	if corsMaxAgeEnv != "" {
		if maxAge, err := fmt.Sscanf(corsMaxAgeEnv, "%d", corsMaxAge); err == nil && maxAge == 1 {
			// Successfully parsed
		} else {
			logger.Warn("Invalid CORS_MAX_AGE value, using default", zap.String("value", corsMaxAgeEnv))
		}
	}

	corsConfig := &mcp.CORSConfig{
		AllowedOrigins: strings.Split(*corsOrigins, ","),
		AllowedMethods: strings.Split(*corsMethods, ","),
		AllowedHeaders: strings.Split(*corsHeaders, ","),
		MaxAge:         *corsMaxAge,
	}

	// Create MCP server
	server, err := mcp.NewServer(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create MCP server", zap.Error(err))
	}

	// Set CORS configuration
	server.SetCORSConfig(corsConfig)

	// Log CORS configuration
	logger.Info("CORS configuration",
		zap.Strings("origins", corsConfig.AllowedOrigins),
		zap.Strings("methods", corsConfig.AllowedMethods),
		zap.Strings("headers", corsConfig.AllowedHeaders),
		zap.Int("max_age", corsConfig.MaxAge))

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
