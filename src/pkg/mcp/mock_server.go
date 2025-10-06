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

// Cleanup cleans up resources
func (s *MockServer) Cleanup() error {
	return s.mockDB.Close()
}
