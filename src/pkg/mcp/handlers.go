// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"fmt"
	"strconv"

	"github.com/maximilien/weave-mcp/src/pkg/weaviate"
)

// handleListCollections handles the list_collections tool
func (s *Server) handleListCollections(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collections, err := s.weaviate.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	return map[string]interface{}{
		"collections": collections,
		"count":       len(collections),
	}, nil
}

// handleCreateCollection handles the create_collection tool
func (s *Server) handleCreateCollection(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	collectionType, ok := args["type"].(string)
	if !ok {
		return nil, fmt.Errorf("collection type is required")
	}

	description, _ := args["description"].(string)

	// Create collection using Weaviate client with schema type
	err := s.weaviate.CreateCollectionWithSchema(ctx, name, "text-embedding-3-small", nil, collectionType)
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

// handleDeleteCollection handles the delete_collection tool
func (s *Server) handleDeleteCollection(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	// Delete collection schema completely using Weaviate client
	err := s.weaviate.DeleteCollectionSchema(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to delete collection: %w", err)
	}

	return map[string]interface{}{
		"name":   name,
		"status": "deleted",
	}, nil
}

// handleListDocuments handles the list_documents tool
func (s *Server) handleListDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
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

	documents, err := s.weaviate.ListDocuments(ctx, collection, limit)
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

// handleCreateDocument handles the create_document tool
func (s *Server) handleCreateDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
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

	// Create document using Weaviate client
	doc := weaviate.Document{
		URL:      url,
		Text:     text,
		Content:  text, // Use text as content
		Metadata: metadata,
	}

	err := s.weaviate.CreateDocument(ctx, collection, doc)
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

// handleGetDocument handles the get_document tool
func (s *Server) handleGetDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	documentID, ok := args["document_id"].(string)
	if !ok {
		return nil, fmt.Errorf("document ID is required")
	}

	// Get document using Weaviate client
	doc, err := s.weaviate.GetDocument(ctx, collection, documentID)
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

// handleDeleteDocument handles the delete_document tool
func (s *Server) handleDeleteDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	documentID, ok := args["document_id"].(string)
	if !ok {
		return nil, fmt.Errorf("document ID is required")
	}

	// Delete document using Weaviate client
	err := s.weaviate.DeleteDocument(ctx, collection, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete document: %w", err)
	}

	return map[string]interface{}{
		"document_id": documentID,
		"collection":  collection,
		"status":      "deleted",
	}, nil
}

// handleCountDocuments handles the count_documents tool
func (s *Server) handleCountDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	// Count documents using Weaviate client
	count, err := s.weaviate.CountDocuments(ctx, collection)
	if err != nil {
		return nil, fmt.Errorf("failed to count documents: %w", err)
	}

	return map[string]interface{}{
		"collection": collection,
		"count":      count,
	}, nil
}

// handleQueryDocuments handles the query_documents tool
func (s *Server) handleQueryDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
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

	// Query documents using Weaviate client
	queryOptions := weaviate.QueryOptions{
		TopK: limit,
	}

	results, err := s.weaviate.Query(ctx, collection, query, queryOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}

	// Convert results to a more MCP-friendly format
	var result []map[string]interface{}
	for _, res := range results {
		result = append(result, map[string]interface{}{
			"id":       res.ID,
			"content":  res.Content,
			"metadata": res.Metadata,
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

// handleUpdateDocument handles the update_document tool
func (s *Server) handleUpdateDocument(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	documentID, ok := args["document_id"].(string)
	if !ok {
		return nil, fmt.Errorf("document ID is required")
	}

	content, _ := args["content"].(string)

	var metadata map[string]interface{}
	if metadataArg, ok := args["metadata"].(map[string]interface{}); ok {
		metadata = metadataArg
	}

	// Validate that at least one field is being updated
	if content == "" && len(metadata) == 0 {
		return nil, fmt.Errorf("must provide at least one of: content or metadata")
	}

	// Update document using Weaviate client
	err := s.weaviate.UpdateDocument(ctx, collection, documentID, content, metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return map[string]interface{}{
		"document_id": documentID,
		"collection":  collection,
		"status":      "updated",
	}, nil
}
