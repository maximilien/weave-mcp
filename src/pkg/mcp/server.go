// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-mcp/src/pkg/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

// Server represents the MCP server implementation
type Server struct {
	config     *config.Config
	logger     *zap.Logger
	dbClient   vectordb.VectorDBClient
	corsConfig *CORSConfig
	mu         sync.RWMutex
	Tools      map[string]Tool
}

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
	Handler     func(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// NewServer creates a new MCP server
func NewServer(cfg *config.Config, logger *zap.Logger) (*Server, error) {
	server := &Server{
		config:     cfg,
		logger:     logger,
		corsConfig: DefaultCORSConfig(),
		Tools:      make(map[string]Tool),
	}

	// Initialize vector database client
	if err := server.initializeVectorDB(); err != nil {
		return nil, fmt.Errorf("failed to initialize vector database client: %w", err)
	}

	// Register tools
	server.registerTools()

	return server, nil
}

// SetCORSConfig sets the CORS configuration for the server
func (s *Server) SetCORSConfig(config *CORSConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.corsConfig = config
}

// initializeVectorDB initializes the vector database client
func (s *Server) initializeVectorDB() error {
	// Get the default database configuration
	dbConfig, err := s.config.GetDefaultDatabase()
	if err != nil {
		return fmt.Errorf("failed to get default database: %w", err)
	}

	// Convert to vectordb.Config
	vdbConfig := &vectordb.Config{
		Type:               vectordb.VectorDBType(dbConfig.Type),
		URL:                dbConfig.URL,
		APIKey:             dbConfig.APIKey,
		OpenAIAPIKey:       dbConfig.OpenAIAPIKey,
		DatabaseURL:        dbConfig.DatabaseURL,
		DatabaseKey:        dbConfig.DatabaseKey,
		Timeout:            dbConfig.Timeout,
		Enabled:            dbConfig.Enabled,
		SimulateEmbeddings: dbConfig.SimulateEmbeddings,
		EmbeddingDimension: dbConfig.EmbeddingDimension,
	}

	// Create vector database client using factory pattern
	client, err := vectordb.CreateClient(vdbConfig)
	if err != nil {
		return fmt.Errorf("failed to create vector database client: %w", err)
	}

	// Test the connection with a health check
	ctx := context.Background()
	if err := client.Health(ctx); err != nil {
		s.logger.Warn("Vector database health check failed (non-fatal)",
			zap.String("type", string(dbConfig.Type)),
			zap.String("name", dbConfig.Name),
			zap.Error(err))
		// Don't fail - some databases may not support health checks
	}

	s.dbClient = client
	s.logger.Info("Vector database initialized",
		zap.String("type", string(dbConfig.Type)),
		zap.String("name", dbConfig.Name))
	return nil
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge         int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Content-Type", "Authorization", "X-Requested-With"},
		MaxAge:         86400, // 24 hours
	}
}

// corsMiddleware creates a CORS middleware
func (s *Server) corsMiddleware(config *CORSConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			origin := r.Header.Get("Origin")
			if s.isOriginAllowed(origin, config.AllowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			} else if len(config.AllowedOrigins) > 0 && config.AllowedOrigins[0] == "*" {
				w.Header().Set("Access-Control-Allow-Origin", "*")
			}

			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if an origin is allowed
func (s *Server) isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		// Support wildcard subdomains like *.example.com
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}
	return false
}

// Handler returns the HTTP handler for the MCP server
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	// Health check endpoint
	mux.HandleFunc("/health", s.handleHealth)

	// Metrics endpoint (Prometheus format)
	mux.Handle("/metrics", promhttp.Handler())

	// MCP endpoints
	mux.HandleFunc("/mcp/tools/list", s.handleToolsList)
	mux.HandleFunc("/mcp/tools/call", s.handleToolCall)

	// Apply CORS middleware with configured settings
	s.mu.RLock()
	corsConfig := s.corsConfig
	s.mu.RUnlock()

	return s.corsMiddleware(corsConfig)(mux)
}

// MetricsHandler returns a standalone metrics HTTP handler for :9091
func (s *Server) MetricsHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/healthz", s.handleMetricsHealth)
	return mux
}

// handleMetricsHealth handles health check for metrics endpoint
func (s *Server) handleMetricsHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"service":   "weave-mcp-metrics",
	}
	json.NewEncoder(w).Encode(response)
}

// registerTools registers all available MCP tools
func (s *Server) registerTools() {
	// Collection management tools
	s.registerTool(Tool{
		Name:        "list_collections",
		Description: "List all collections in the vector database",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: s.withMetrics("list_collections", s.handleListCollections),
	})

	s.registerTool(Tool{
		Name:        "create_collection",
		Description: "Create a new collection in the vector database",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection to create",
				},
				"type": map[string]interface{}{
					"type":        "string",
					"description": "Type of collection (text or image)",
					"enum":        []string{"text", "image"},
				},
				"description": map[string]interface{}{
					"type":        "string",
					"description": "Description of the collection",
				},
				"vectorizer": map[string]interface{}{
					"type":        "string",
					"description": "Embedding model/vectorizer to use (e.g., text2vec-openai, text-embedding-3-small, text-embedding-ada-002)",
					"default":     "text2vec-openai",
				},
			},
			"required": []string{"name", "type"},
		},
		Handler: s.withMetrics("create_collection", s.handleCreateCollection),
	})

	s.registerTool(Tool{
		Name:        "delete_collection",
		Description: "Delete a collection from the vector database",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection to delete",
				},
			},
			"required": []string{"name"},
		},
		Handler: s.withMetrics("delete_collection", s.handleDeleteCollection),
	})

	// Document management tools
	s.registerTool(Tool{
		Name:        "list_documents",
		Description: "List documents in a collection",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of documents to return",
					"default":     10,
				},
			},
			"required": []string{"collection"},
		},
		Handler: s.handleListDocuments,
	})

	s.registerTool(Tool{
		Name:        "create_document",
		Description: "Create a new document in a collection",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
				"url": map[string]interface{}{
					"type":        "string",
					"description": "URL of the document",
				},
				"text": map[string]interface{}{
					"type":        "string",
					"description": "Text content of the document",
				},
				"metadata": map[string]interface{}{
					"type":        "object",
					"description": "Additional metadata for the document",
					"default":     map[string]interface{}{},
				},
			},
			"required": []string{"collection", "url", "text"},
		},
		Handler: s.handleCreateDocument,
	})

	s.registerTool(Tool{
		Name:        "batch_create_documents",
		Description: "Create multiple documents in a collection in a single batch operation",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
				"documents": map[string]interface{}{
					"type":        "array",
					"description": "Array of documents to create",
					"items": map[string]interface{}{
						"type": "object",
						"properties": map[string]interface{}{
							"url": map[string]interface{}{
								"type":        "string",
								"description": "URL of the document",
							},
							"text": map[string]interface{}{
								"type":        "string",
								"description": "Text content of the document",
							},
							"metadata": map[string]interface{}{
								"type":        "object",
								"description": "Additional metadata for the document",
								"default":     map[string]interface{}{},
							},
						},
						"required": []string{"url", "text"},
					},
				},
			},
			"required": []string{"collection", "documents"},
		},
		Handler: s.handleBatchCreateDocuments,
	})

	s.registerTool(Tool{
		Name:        "get_document",
		Description: "Get a specific document by ID",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
				"document_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the document to retrieve",
				},
			},
			"required": []string{"collection", "document_id"},
		},
		Handler: s.handleGetDocument,
	})

	s.registerTool(Tool{
		Name:        "delete_document",
		Description: "Delete a document from a collection",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
				"document_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the document to delete",
				},
			},
			"required": []string{"collection", "document_id"},
		},
		Handler: s.handleDeleteDocument,
	})

	s.registerTool(Tool{
		Name:        "count_documents",
		Description: "Count documents in a collection",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
			},
			"required": []string{"collection"},
		},
		Handler: s.handleCountDocuments,
	})

	s.registerTool(Tool{
		Name:        "update_document",
		Description: "Update a document's content or metadata",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
				"document_id": map[string]interface{}{
					"type":        "string",
					"description": "ID of the document to update",
				},
				"content": map[string]interface{}{
					"type":        "string",
					"description": "New content for the document (optional)",
				},
				"metadata": map[string]interface{}{
					"type":        "object",
					"description": "Metadata fields to update (optional)",
					"default":     map[string]interface{}{},
				},
			},
			"required": []string{"collection", "document_id"},
		},
		Handler: s.handleUpdateDocument,
	})

	// Query tools
	s.registerTool(Tool{
		Name:        "query_documents",
		Description: "Query documents using semantic search",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Search query",
				},
				"limit": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of results to return",
					"default":     5,
				},
			},
			"required": []string{"collection", "query"},
		},
		Handler: s.handleQueryDocuments,
	})

	// AI tools
	s.registerTool(Tool{
		Name:        "suggest_schema",
		Description: "Analyze documents and suggest an optimal collection schema using AI",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to directory or file(s) to analyze",
				},
				"collection_name": map[string]interface{}{
					"type":        "string",
					"description": "Name for the target collection",
				},
				"requirements": map[string]interface{}{
					"type":        "string",
					"description": "Optional user requirements for the schema",
					"default":     "",
				},
				"vdb_type": map[string]interface{}{
					"type":        "string",
					"description": "Target vector database type",
					"default":     "weaviate",
				},
				"max_samples": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of files to sample",
					"default":     50,
				},
			},
			"required": []string{"source_path", "collection_name"},
		},
		Handler: s.handleSuggestSchema,
	})

	s.registerTool(Tool{
		Name:        "suggest_chunking",
		Description: "Analyze documents and suggest optimal chunking configuration using AI",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"source_path": map[string]interface{}{
					"type":        "string",
					"description": "Path to directory or file(s) to analyze",
				},
				"collection_name": map[string]interface{}{
					"type":        "string",
					"description": "Name for the target collection",
				},
				"requirements": map[string]interface{}{
					"type":        "string",
					"description": "Optional user requirements for chunking",
					"default":     "",
				},
				"vdb_type": map[string]interface{}{
					"type":        "string",
					"description": "Target vector database type",
					"default":     "weaviate",
				},
				"max_samples": map[string]interface{}{
					"type":        "integer",
					"description": "Maximum number of files to sample",
					"default":     50,
				},
			},
			"required": []string{"source_path", "collection_name"},
		},
		Handler: s.handleSuggestChunking,
	})

	// Health and monitoring tools
	s.registerTool(Tool{
		Name:        "health_check",
		Description: "Check the health and connectivity of the vector database",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: s.handleHealthCheck,
	})

	s.registerTool(Tool{
		Name:        "count_collections",
		Description: "Count the total number of collections in the database",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: s.handleCountCollections,
	})

	s.registerTool(Tool{
		Name:        "show_collection",
		Description: "Show detailed information about a collection including schema, count, and properties",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection to show",
				},
			},
			"required": []string{"name"},
		},
		Handler: s.handleShowCollection,
	})

	// Embedding tools
	s.registerTool(Tool{
		Name:        "list_embedding_models",
		Description: "List all available embedding models and their properties",
		InputSchema: map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		},
		Handler: s.handleListEmbeddingModels,
	})

	s.registerTool(Tool{
		Name:        "show_collection_embeddings",
		Description: "Show embedding configuration for a specific collection",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
			},
			"required": []string{"name"},
		},
		Handler: s.handleShowCollectionEmbeddings,
	})

	// Phase 4 tools - Medium priority operations
	s.registerTool(Tool{
		Name:        "get_collection_stats",
		Description: "Get statistics for a collection including document count and schema info",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
			},
			"required": []string{"name"},
		},
		Handler: s.handleGetCollectionStats,
	})

	s.registerTool(Tool{
		Name:        "delete_all_documents",
		Description: "Delete all documents from a collection or all collections (use with caution)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection (optional - if not provided, deletes from all collections)",
				},
			},
		},
		Handler: s.handleDeleteAllDocuments,
	})

	s.registerTool(Tool{
		Name:        "show_document_by_name",
		Description: "Show a document by filename instead of ID",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
				"filename": map[string]interface{}{
					"type":        "string",
					"description": "Filename to search for in URL or metadata",
				},
			},
			"required": []string{"collection", "filename"},
		},
		Handler: s.handleShowDocumentByName,
	})

	s.registerTool(Tool{
		Name:        "delete_document_by_name",
		Description: "Delete a document by filename instead of ID",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection",
				},
				"filename": map[string]interface{}{
					"type":        "string",
					"description": "Filename to search for in URL or metadata",
				},
			},
			"required": []string{"collection", "filename"},
		},
		Handler: s.handleDeleteDocumentByName,
	})

	s.registerTool(Tool{
		Name:        "execute_query",
		Description: "Execute a natural language query against documents in one or all collections",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Natural language query to search for",
				},
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection (optional - if not provided, searches all collections)",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results to return (default: 5)",
				},
			},
			"required": []string{"query"},
		},
		Handler: s.handleExecuteQuery,
	})

	// Phase 1: Observability & Monitoring tools
	s.registerTool(Tool{
		Name:        "configure_logging",
		Description: "Configure structured logging for MCP server (log level, format, file output)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"log_level": map[string]interface{}{
					"type":        "string",
					"description": "Log level (debug, info, warn, error)",
					"enum":        []string{"debug", "info", "warn", "error"},
					"default":     "info",
				},
				"log_format": map[string]interface{}{
					"type":        "string",
					"description": "Log format (text or json)",
					"enum":        []string{"text", "json"},
					"default":     "text",
				},
				"log_file": map[string]interface{}{
					"type":        "string",
					"description": "Path to log file (optional, logs to stderr if not specified)",
				},
			},
		},
		Handler: s.withMetrics("configure_logging", s.handleConfigureLogging),
	})

	s.registerTool(Tool{
		Name:        "get_metrics",
		Description: "Retrieve current Prometheus metrics for MCP operations",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"metric_name": map[string]interface{}{
					"type":        "string",
					"description": "Filter by specific metric name (optional)",
				},
				"format": map[string]interface{}{
					"type":        "string",
					"description": "Output format (prometheus or json)",
					"enum":        []string{"prometheus", "json"},
					"default":     "json",
				},
			},
		},
		Handler: s.withMetrics("get_metrics", s.handleGetMetrics),
	})

	s.registerTool(Tool{
		Name:        "check_health",
		Description: "Check detailed health status of VDB connections and MCP server",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"detailed": map[string]interface{}{
					"type":        "boolean",
					"description": "Include detailed connection information",
					"default":     true,
				},
			},
		},
		Handler: s.withMetrics("check_health", s.handleCheckHealth),
	})

	// Phase 2: Agent framework tools
	s.registerTool(Tool{
		Name:        "list_agents",
		Description: "List all available specialized agents for complex tasks",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agent_type": map[string]interface{}{
					"type":        "string",
					"description": "Filter by agent type (optional)",
				},
			},
		},
		Handler: s.withMetrics("list_agents", s.handleListAgents),
	})

	s.registerTool(Tool{
		Name:        "get_agent_info",
		Description: "Get detailed information about a specific agent",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agent_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the agent to get info about",
				},
			},
			"required": []string{"agent_name"},
		},
		Handler: s.withMetrics("get_agent_info", s.handleGetAgentInfo),
	})

	s.registerTool(Tool{
		Name:        "run_agent",
		Description: "Execute a specialized agent for complex tasks (e.g., schema design, query optimization, RAG)",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"agent_name": map[string]interface{}{
					"type":        "string",
					"description": "Name of the agent to run (e.g., 'schema', 'chunking', 'rag')",
				},
				"task": map[string]interface{}{
					"type":        "string",
					"description": "Description of the task for the agent to perform",
				},
				"parameters": map[string]interface{}{
					"type":        "object",
					"description": "Agent-specific parameters (varies by agent type)",
					"default":     map[string]interface{}{},
				},
			},
			"required": []string{"agent_name", "task"},
		},
		Handler: s.withMetrics("run_agent", s.handleRunAgent),
	})
}

// registerTool registers a tool with the server
func (s *Server) registerTool(tool Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tools[tool.Name] = tool
	s.logger.Debug("Registered tool", zap.String("name", tool.Name))
}

// handleHealth handles health check requests
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check database health
	dbStatus := "healthy"
	dbError := ""
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.dbClient.Health(ctx); err != nil {
		dbStatus = "unhealthy"
		dbError = err.Error()
		s.logger.Warn("Database health check failed", zap.Error(err))
	}

	// Get database type from config
	dbConfig, _ := s.config.GetDefaultDatabase()
	dbType := "unknown"
	dbName := "unknown"
	if dbConfig != nil {
		dbType = string(dbConfig.Type)
		dbName = dbConfig.Name
	}

	// Overall status is healthy only if database is healthy
	overallStatus := "healthy"
	httpStatus := http.StatusOK
	if dbStatus != "healthy" {
		overallStatus = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	response := map[string]interface{}{
		"status":    overallStatus,
		"timestamp": time.Now().UTC(),
		"version":   "dev",
		"database": map[string]interface{}{
			"status": dbStatus,
			"type":   dbType,
			"name":   dbName,
		},
	}

	if dbError != "" {
		response["database"].(map[string]interface{})["error"] = dbError
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode health response", zap.Error(err))
	}
}

// handleToolsList handles tool listing requests
func (s *Server) handleToolsList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.mu.RLock()
	tools := make([]map[string]interface{}, 0, len(s.Tools))
	for _, tool := range s.Tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": tool.InputSchema,
		})
	}
	s.mu.RUnlock()

	response := map[string]interface{}{
		"tools": tools,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode tools list response", zap.Error(err))
	}
}

// handleToolCall handles tool execution requests
func (s *Server) handleToolCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	tool, exists := s.Tools[request.Name]
	s.mu.RUnlock()

	if !exists {
		http.Error(w, fmt.Sprintf("Tool '%s' not found", request.Name), http.StatusNotFound)
		return
	}

	// Execute tool with timeout
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := tool.Handler(ctx, request.Arguments)
	if err != nil {
		s.logger.Error("Tool execution failed",
			zap.String("tool", request.Name),
			zap.Error(err))

		response := map[string]interface{}{
			"error": err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		if encodeErr := json.NewEncoder(w).Encode(response); encodeErr != nil {
			s.logger.Error("Failed to encode error response", zap.Error(encodeErr))
		}
		return
	}

	response := map[string]interface{}{
		"result": result,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		s.logger.Error("Failed to encode tool call response", zap.Error(err))
	}
}

// Cleanup cleans up resources
func (s *Server) Cleanup() error {
	// Close Weaviate client if needed
	// (Weaviate client doesn't have a Close method, so nothing to do here)
	return nil
}
