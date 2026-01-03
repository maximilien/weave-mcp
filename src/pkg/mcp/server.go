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

	// MCP endpoints
	mux.HandleFunc("/mcp/tools/list", s.handleToolsList)
	mux.HandleFunc("/mcp/tools/call", s.handleToolCall)

	// Apply CORS middleware with configured settings
	s.mu.RLock()
	corsConfig := s.corsConfig
	s.mu.RUnlock()

	return s.corsMiddleware(corsConfig)(mux)
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
		Handler: s.handleListCollections,
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
		Handler: s.handleCreateCollection,
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
		Handler: s.handleDeleteCollection,
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
