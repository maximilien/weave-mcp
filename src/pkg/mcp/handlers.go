// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"fmt"
	"strconv"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// handleListCollections handles the list_collections tool
func (s *Server) handleListCollections(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collections, err := s.dbClient.ListCollections(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list collections: %w", err)
	}

	// Convert to string array for consistent output
	collectionNames := make([]string, 0, len(collections))
	for _, coll := range collections {
		collectionNames = append(collectionNames, coll.Name)
	}

	return map[string]interface{}{
		"collections": collectionNames,
		"count":       len(collectionNames),
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

	// Get vectorizer/embedding model (default to text2vec-openai)
	vectorizer := "text2vec-openai"
	if v, ok := args["vectorizer"].(string); ok && v != "" {
		vectorizer = v
	}

	// Create basic schema based on type
	schema := &vectordb.CollectionSchema{
		Class:      name,
		Vectorizer: vectorizer,
		Properties: []vectordb.SchemaProperty{
			{
				Name:     "text",
				DataType: []string{"text"},
			},
			{
				Name:     "url",
				DataType: []string{"text"},
			},
		},
	}

	// Add image-specific properties for image collections
	if collectionType == "image" {
		schema.Properties = append(schema.Properties, vectordb.SchemaProperty{
			Name:     "image",
			DataType: []string{"text"},
		})
	}

	err := s.dbClient.CreateCollection(ctx, name, schema)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	return map[string]interface{}{
		"name":        name,
		"type":        collectionType,
		"description": description,
		"vectorizer":  vectorizer,
		"status":      "created",
	}, nil
}

// handleDeleteCollection handles the delete_collection tool
func (s *Server) handleDeleteCollection(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	err := s.dbClient.DeleteCollection(ctx, name)
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

	documents, err := s.dbClient.ListDocuments(ctx, collection, limit, 0)
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

	// Create document using vectordb client
	doc := &vectordb.Document{
		URL:      url,
		Text:     text,
		Content:  text, // Use text as content
		Metadata: metadata,
	}

	err := s.dbClient.CreateDocument(ctx, collection, doc)
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

	// Get document using vectordb client
	doc, err := s.dbClient.GetDocument(ctx, collection, documentID)
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

	// Delete document using vectordb client
	err := s.dbClient.DeleteDocument(ctx, collection, documentID)
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

	// Count documents using vectordb client
	count, err := s.dbClient.GetCollectionCount(ctx, collection)
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

	// Query documents using vectordb client
	queryOptions := &vectordb.QueryOptions{
		TopK: limit,
	}

	results, err := s.dbClient.SearchSemantic(ctx, collection, query, queryOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to query documents: %w", err)
	}

	// Convert results to a more MCP-friendly format
	var result []map[string]interface{}
	for _, res := range results {
		result = append(result, map[string]interface{}{
			"id":       res.Document.ID,
			"content":  res.Document.Content,
			"text":     res.Document.Text,
			"url":      res.Document.URL,
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

	// Get the existing document first
	doc, err := s.dbClient.GetDocument(ctx, collection, documentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing document: %w", err)
	}

	// Update the fields
	if content != "" {
		doc.Content = content
		doc.Text = content
	}
	if len(metadata) > 0 {
		if doc.Metadata == nil {
			doc.Metadata = make(map[string]interface{})
		}
		for k, v := range metadata {
			doc.Metadata[k] = v
		}
	}

	// Update document using vectordb client
	err = s.dbClient.UpdateDocument(ctx, collection, doc)
	if err != nil {
		return nil, fmt.Errorf("failed to update document: %w", err)
	}

	return map[string]interface{}{
		"document_id": documentID,
		"collection":  collection,
		"status":      "updated",
	}, nil
}
