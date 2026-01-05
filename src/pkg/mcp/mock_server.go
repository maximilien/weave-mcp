// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/maximilien/weave-mcp/src/pkg/config"
	"github.com/maximilien/weave-mcp/src/pkg/mock"
	"go.uber.org/zap"
)

// MockServer represents a mock MCP server implementation for testing
type MockServer struct {
	config *config.Config
	logger *zap.Logger
	mockDB *mock.Client
	mu     sync.RWMutex
	Tools  map[string]Tool
}

// NewMockServer creates a new mock MCP server for testing
func NewMockServer(cfg *config.Config, logger *zap.Logger) (*MockServer, error) {
	server := &MockServer{
		config: cfg,
		logger: logger,
		mockDB: mock.NewClient(),
		Tools:  make(map[string]Tool),
	}

	// Initialize mock database
	if err := server.mockDB.Connect(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to initialize mock database: %w", err)
	}

	// Register tools
	server.registerTools()

	return server, nil
}

// registerTools registers all available MCP tools for mock server
func (s *MockServer) registerTools() {
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

	// Phase 4 tools
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
		Description: "Delete all documents from a collection or all collections",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection (optional)",
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
					"description": "Filename to search for",
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
					"description": "Filename to search for",
				},
			},
			"required": []string{"collection", "filename"},
		},
		Handler: s.handleDeleteDocumentByName,
	})

	s.registerTool(Tool{
		Name:        "execute_query",
		Description: "Execute a natural language query against documents",
		InputSchema: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"query": map[string]interface{}{
					"type":        "string",
					"description": "Natural language query",
				},
				"collection": map[string]interface{}{
					"type":        "string",
					"description": "Name of the collection (optional)",
				},
				"limit": map[string]interface{}{
					"type":        "number",
					"description": "Maximum number of results (default: 5)",
				},
			},
			"required": []string{"query"},
		},
		Handler: s.handleExecuteQuery,
	})
}

// registerTool registers a tool with the mock server
func (s *MockServer) registerTool(tool Tool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tools[tool.Name] = tool
	s.logger.Debug("Registered tool", zap.String("name", tool.Name))
}

// Mock tool handlers

func (s *MockServer) handleListCollections(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collections, err := s.mockDB.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return map[string]interface{}{
		"collections": collections,
		"count":       len(collections),
	}, nil
}

func (s *MockServer) handleCreateCollection(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	collectionType, ok := args["type"].(string)
	if !ok {
		return nil, fmt.Errorf("collection type is required")
	}

	description, _ := args["description"].(string)

	// Create collection using mock client
	err := s.mockDB.CreateCollection(ctx, name, collectionType, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return map[string]interface{}{
		"name":        name,
		"type":        collectionType,
		"description": description,
		"status":      "created",
	}, nil
}

func (s *MockServer) handleDeleteCollection(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	// Delete collection using mock client
	err := s.mockDB.DeleteCollection(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to delete collection: %w", err)
	}

	return map[string]interface{}{
		"name":   name,
		"status": "deleted",
	}, nil
}

func (s *MockServer) handleListDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	limit := 10
	if limitStr, ok := args["limit"].(string); ok {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	} else if limitInt, ok := args["limit"].(int); ok {
		limit = limitInt
	}

	documents, err := s.mockDB.ListDocuments(ctx, collection, limit, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	// Convert documents to a more MCP-friendly format
	var result []map[string]interface{}
	for _, doc := range documents {
		result = append(result, map[string]interface{}{
			"id":       doc.ID,
			"url":      doc.URL,
			"text":     doc.Text,
			"content":  doc.Content,
			"metadata": doc.Metadata,
		})
	}

	return map[string]interface{}{
		"documents":  result,
		"count":      len(result),
		"collection": collection,
	}, nil
}

func (s *MockServer) handleCreateDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	url, ok := args["url"].(string)
	if !ok {
		return nil, fmt.Errorf("document URL is required")
	}

	text, ok := args["text"].(string)
	if !ok {
		return nil, fmt.Errorf("document text is required")
	}

	metadata, _ := args["metadata"].(map[string]interface{})
	if metadata == nil {
		metadata = make(map[string]interface{})
	}

	// Create document using mock client
	doc := mock.Document{
		URL:      url,
		Text:     text,
		Content:  text, // Use text as content
		Metadata: metadata,
	}

	err := s.mockDB.Insert(ctx, collection, []mock.Document{doc})
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return map[string]interface{}{
		"collection": collection,
		"url":        url,
		"text":       text,
		"metadata":   metadata,
		"status":     "created",
	}, nil
}

func (s *MockServer) handleGetDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	documentID, ok := args["document_id"].(string)
	if !ok {
		return nil, fmt.Errorf("document ID is required")
	}

	// Get document using mock client
	doc, err := s.mockDB.GetDocument(ctx, collection, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	return map[string]interface{}{
		"id":         doc.ID,
		"url":        doc.URL,
		"text":       doc.Text,
		"content":    doc.Content,
		"metadata":   doc.Metadata,
		"collection": collection,
	}, nil
}

func (s *MockServer) handleDeleteDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	documentID, ok := args["document_id"].(string)
	if !ok {
		return nil, fmt.Errorf("document ID is required")
	}

	// Delete document using mock client
	err := s.mockDB.DeleteDocument(ctx, collection, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	return map[string]interface{}{
		"document_id": documentID,
		"collection":  collection,
		"status":      "deleted",
	}, nil
}

func (s *MockServer) handleCountDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	// Count documents using mock client
	count, err := s.mockDB.CountDocuments(ctx, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	return map[string]interface{}{
		"collection": collection,
		"count":      count,
	}, nil
}

func (s *MockServer) handleQueryDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	query, ok := args["query"].(string)
	if !ok {
		return nil, fmt.Errorf("query is required")
	}

	limit := 5
	if limitStr, ok := args["limit"].(string); ok {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil {
			limit = parsedLimit
		}
	} else if limitInt, ok := args["limit"].(int); ok {
		limit = limitInt
	}

	// Query documents using mock client
	results, err := s.mockDB.Search(ctx, collection, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}

	// Convert results to a more MCP-friendly format
	var result []map[string]interface{}
	for _, res := range results {
		result = append(result, map[string]interface{}{
			"id":       res.Document.ID,
			"content":  res.Document.Text,
			"metadata": res.Document.Metadata,
			"score":    res.Score,
		})
	}

	return map[string]interface{}{
		"results":    result,
		"count":      len(result),
		"collection": collection,
		"query":      query,
	}, nil
}

// New tool handlers (Phase 1)

func (s *MockServer) handleHealthCheck(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Mock health check - always return healthy for mock server
	return map[string]interface{}{
		"status":   "healthy",
		"database": "mock",
		"url":      "mock://localhost",
	}, nil
}

func (s *MockServer) handleCountCollections(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collections, err := s.mockDB.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return map[string]interface{}{
		"count":       len(collections),
		"collections": collections,
	}, nil
}

func (s *MockServer) handleShowCollection(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	// Get collection info from mock DB
	_, err := s.mockDB.GetCollectionInfo(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	count, err := s.mockDB.CountDocuments(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	// Return mock schema
	return map[string]interface{}{
		"name":  name,
		"count": count,
		"schema": map[string]interface{}{
			"class": name,
			"properties": []map[string]interface{}{
				{
					"name":     "text",
					"dataType": []string{"text"},
				},
				{
					"name":     "url",
					"dataType": []string{"string"},
				},
			},
		},
		"vectorizer": "text2vec-openai",
		"properties": []map[string]interface{}{
			{"name": "text", "type": "text"},
			{"name": "url", "type": "string"},
		},
	}, nil
}

func (s *MockServer) handleListEmbeddingModels(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Return same models as real server
	models := []map[string]interface{}{
		{
			"name":        "text2vec-openai",
			"type":        "openai",
			"description": "OpenAI text embedding model (legacy, uses text-embedding-ada-002)",
			"dimensions":  1536,
			"provider":    "openai",
		},
		{
			"name":        "text-embedding-3-small",
			"type":        "openai",
			"description": "OpenAI's latest small embedding model - faster and cheaper",
			"dimensions":  1536,
			"provider":    "openai",
		},
		{
			"name":        "text-embedding-3-large",
			"type":        "openai",
			"description": "OpenAI's latest large embedding model - better quality",
			"dimensions":  3072,
			"provider":    "openai",
		},
		{
			"name":        "text-embedding-ada-002",
			"type":        "openai",
			"description": "OpenAI's Ada model (legacy)",
			"dimensions":  1536,
			"provider":    "openai",
		},
	}

	return map[string]interface{}{
		"models": models,
		"count":  len(models),
	}, nil
}

func (s *MockServer) handleShowCollectionEmbeddings(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	// Verify collection exists
	_, err := s.mockDB.GetCollectionInfo(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	// Return mock embedding config
	return map[string]interface{}{
		"collection": name,
		"vectorizer": "text-embedding-3-small",
		"model":      "text-embedding-3-small",
		"dimensions": 1536,
		"provider":   "openai",
	}, nil
}

// Phase 4 tool handlers

func (s *MockServer) handleGetCollectionStats(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	// Verify collection exists
	_, err := s.mockDB.GetCollectionInfo(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get collection info: %w", err)
	}

	// Get document count
	count, err := s.mockDB.CountDocuments(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	return map[string]interface{}{
		"collection":     name,
		"document_count": count,
		"schema": map[string]interface{}{
			"vectorizer": "text-embedding-3-small",
			"properties": 3,
		},
	}, nil
}

func (s *MockServer) handleDeleteAllDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collectionName, _ := args["collection"].(string)

	if collectionName == "" {
		// Delete all documents from all collections
		collections, err := s.mockDB.ListCollections(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list collections: %w", err)
		}

		totalDeleted := 0
		for _, coll := range collections {
			docs, err := s.mockDB.ListDocuments(ctx, coll, 1000, 0)
			if err != nil {
				continue
			}
			for _, doc := range docs {
				err := s.mockDB.DeleteDocument(ctx, coll, doc.ID)
				if err != nil {
					continue
				}
				totalDeleted++
			}
		}

		return map[string]interface{}{
			"deleted_count":       totalDeleted,
			"collections_cleaned": len(collections),
		}, nil
	}

	// Delete all documents from specific collection
	docs, err := s.mockDB.ListDocuments(ctx, collectionName, 1000, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	deletedCount := 0
	for _, doc := range docs {
		err := s.mockDB.DeleteDocument(ctx, collectionName, doc.ID)
		if err != nil {
			continue
		}
		deletedCount++
	}

	return map[string]interface{}{
		"collection":    collectionName,
		"deleted_count": deletedCount,
	}, nil
}

func (s *MockServer) handleShowDocumentByName(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collectionName, ok := args["collection"].(string)
	if !ok || collectionName == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	filename, ok := args["filename"].(string)
	if !ok || filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	// List documents and search by filename
	docs, err := s.mockDB.ListDocuments(ctx, collectionName, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	// Search for document with matching filename
	for _, doc := range docs {
		if doc.URL != "" && doc.URL == filename {
			return map[string]interface{}{
				"document_id": doc.ID,
				"collection":  collectionName,
				"url":         doc.URL,
				"text":        doc.Text,
				"metadata":    doc.Metadata,
			}, nil
		}
		if doc.Metadata != nil {
			if filenameVal, ok := doc.Metadata["filename"].(string); ok && filenameVal == filename {
				return map[string]interface{}{
					"document_id": doc.ID,
					"collection":  collectionName,
					"url":         doc.URL,
					"text":        doc.Text,
					"metadata":    doc.Metadata,
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("document with filename '%s' not found in collection '%s'", filename, collectionName)
}

func (s *MockServer) handleDeleteDocumentByName(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collectionName, ok := args["collection"].(string)
	if !ok || collectionName == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	filename, ok := args["filename"].(string)
	if !ok || filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	// List documents and search by filename
	docs, err := s.mockDB.ListDocuments(ctx, collectionName, 100, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to list documents: %w", err)
	}

	// Search for document with matching filename and delete it
	for _, doc := range docs {
		if doc.URL != "" && doc.URL == filename {
			err := s.mockDB.DeleteDocument(ctx, collectionName, doc.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to delete document: %w", err)
			}
			return map[string]interface{}{
				"document_id": doc.ID,
				"collection":  collectionName,
				"filename":    filename,
				"status":      "deleted",
			}, nil
		}
		if doc.Metadata != nil {
			if filenameVal, ok := doc.Metadata["filename"].(string); ok && filenameVal == filename {
				err := s.mockDB.DeleteDocument(ctx, collectionName, doc.ID)
				if err != nil {
					return nil, fmt.Errorf("failed to delete document: %w", err)
				}
				return map[string]interface{}{
					"document_id": doc.ID,
					"collection":  collectionName,
					"filename":    filename,
					"status":      "deleted",
				}, nil
			}
		}
	}

	return nil, fmt.Errorf("document with filename '%s' not found in collection '%s'", filename, collectionName)
}

func (s *MockServer) handleExecuteQuery(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	collectionName, _ := args["collection"].(string)
	limit := 5
	if limitArg, ok := args["limit"].(float64); ok {
		limit = int(limitArg)
	}

	// If no collection specified, search all collections
	if collectionName == "" {
		collections, err := s.mockDB.ListCollections(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list collections: %w", err)
		}

		allResults := []interface{}{}
		for _, coll := range collections {
			results, err := s.mockDB.Search(ctx, coll, query, limit)
			if err != nil {
				continue
			}

			for _, result := range results {
				resultMap := map[string]interface{}{
					"collection":  coll,
					"document_id": result.Document.ID,
					"text":        result.Document.Text,
					"url":         result.Document.URL,
					"metadata":    result.Document.Metadata,
					"score":       result.Score,
				}
				allResults = append(allResults, resultMap)
			}
		}

		if len(allResults) > limit {
			allResults = allResults[:limit]
		}

		return map[string]interface{}{
			"query":   query,
			"results": allResults,
			"count":   len(allResults),
		}, nil
	}

	// Query specific collection
	results, err := s.mockDB.Search(ctx, collectionName, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	formattedResults := make([]interface{}, len(results))
	for i, result := range results {
		formattedResults[i] = map[string]interface{}{
			"document_id": result.Document.ID,
			"text":        result.Document.Text,
			"url":         result.Document.URL,
			"metadata":    result.Document.Metadata,
			"score":       result.Score,
		}
	}

	return map[string]interface{}{
		"collection": collectionName,
		"query":      query,
		"results":    formattedResults,
		"count":      len(formattedResults),
	}, nil
}

// Cleanup cleans up resources
func (s *MockServer) Cleanup() error {
	return s.mockDB.Close()
}
