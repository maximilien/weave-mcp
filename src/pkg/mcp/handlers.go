// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/maximilien/weave-cli/src/pkg/agents"
	"github.com/maximilien/weave-cli/src/pkg/logging"
	"github.com/maximilien/weave-cli/src/pkg/metrics"
	"github.com/maximilien/weave-cli/src/pkg/vectordb"
)

// generateCorrelationID generates a unique correlation ID for request tracking
func generateCorrelationID() string {
	return "mcp-" + uuid.New().String()[:8]
}

// withMetrics wraps a handler to record metrics and add correlation IDs
func (s *Server) withMetrics(operation string, handler func(ctx context.Context, args map[string]interface{}) (interface{}, error)) func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	return func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
		start := time.Now()
		correlationID := generateCorrelationID()

		// Get VDB type for metrics
		dbConfig, _ := s.config.GetDefaultDatabase()
		vdbType := "unknown"
		if dbConfig != nil {
			vdbType = string(dbConfig.Type)
		}

		// Create logger with context
		logger := logging.WithFields(map[string]interface{}{
			"correlation_id": correlationID,
			"operation":      operation,
			"vdb_type":       vdbType,
		})

		logger.Debug("Starting operation")

		// Execute handler
		result, err := handler(ctx, args)
		duration := time.Since(start)

		// Record metrics
		metrics.RecordRequest(vdbType, operation, duration, err)

		if err != nil {
			logger.Error("Operation failed: %v", err)
			metrics.RecordError(vdbType, operation, categorizeError(err))
		} else {
			logger.Debug("Operation completed successfully in %v", duration)
		}

		// Add metadata to result if it's a map
		if err == nil {
			if resultMap, ok := result.(map[string]interface{}); ok {
				resultMap["_metadata"] = map[string]interface{}{
					"correlation_id": correlationID,
					"duration_ms":    duration.Milliseconds(),
					"operation":      operation,
				}
			}
		}

		return result, err
	}
}

// categorizeError categorizes errors for metrics
func categorizeError(err error) string {
	if err == nil {
		return "none"
	}
	errStr := err.Error()
	switch {
	case strings.Contains(errStr, "connection refused"), strings.Contains(errStr, "dial tcp"):
		return "connection"
	case strings.Contains(errStr, "timeout"), strings.Contains(errStr, "deadline exceeded"):
		return "timeout"
	case strings.Contains(errStr, "authentication"), strings.Contains(errStr, "unauthorized"):
		return "auth"
	case strings.Contains(errStr, "not found"):
		return "not_found"
	default:
		return "unknown"
	}
}

// enhanceError adds helpful context to database errors with VDB type prefix
func (s *Server) enhanceError(operation string, err error) error {
	if err == nil {
		return nil
	}

	// Get database type from config
	dbConfig, configErr := s.config.GetDefaultDatabase()
	if configErr != nil {
		// Fallback if we can't get config
		return fmt.Errorf("%s: %w", operation, err)
	}

	vdbType := string(dbConfig.Type)
	errStr := err.Error()

	// Check for common connection/network errors
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "dial tcp") ||
		strings.Contains(errStr, "no such host") {
		return fmt.Errorf("%s: %s: database connection failed - please check if the database is running and accessible. Original error: %w", vdbType, operation, err)
	}

	// Check for timeout errors
	if strings.Contains(errStr, "context deadline exceeded") ||
		strings.Contains(errStr, "timeout") {
		return fmt.Errorf("%s: %s: operation timed out - database may be slow or unreachable. Original error: %w", vdbType, operation, err)
	}

	// Check for authentication errors
	if strings.Contains(errStr, "authentication failed") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "invalid credentials") {
		return fmt.Errorf("%s: %s: authentication failed - please check your API key or credentials. Original error: %w", vdbType, operation, err)
	}

	// Default: include VDB type in error
	return fmt.Errorf("%s: %s: %w", vdbType, operation, err)
}

// createContextWithTimeout creates a context with operation-specific timeout
func (s *Server) createContextWithTimeout(ctx context.Context, opType vectordb.OperationType) (context.Context, context.CancelFunc) {
	// Get database config to determine if cloud or local
	dbConfig, err := s.config.GetDefaultDatabase()
	if err != nil {
		// Fallback to default timeout
		return context.WithTimeout(ctx, 30*time.Second)
	}

	// Determine if cloud deployment (heuristic based on type)
	isCloud := strings.Contains(string(dbConfig.Type), "cloud") ||
		dbConfig.Type == "supabase" ||
		dbConfig.Type == "mongodb" ||
		dbConfig.Type == "pinecone"

	// Get timeout for this operation type
	timeout := vectordb.GetTimeoutForOperation(opType, isCloud, dbConfig.Timeout)

	return context.WithTimeout(ctx, timeout)
}

// handleListCollections handles the list_collections tool
func (s *Server) handleListCollections(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Create context with collection operation timeout
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeCollection)
	defer cancel()

	collections, err := s.dbClient.ListCollections(timeoutCtx)
	if err != nil {
		return nil, s.enhanceError("failed to list collections", err)
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
			{
				Name:     "metadata",
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

	// Create context with collection operation timeout
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeCollection)
	defer cancel()

	err := s.dbClient.CreateCollection(timeoutCtx, name, schema)
	if err != nil {
		return nil, s.enhanceError("failed to create collection", err)
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

	// Create context with collection operation timeout
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeCollection)
	defer cancel()

	err := s.dbClient.DeleteCollection(timeoutCtx, name)
	if err != nil {
		return nil, s.enhanceError("failed to delete collection", err)
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

	// Create context with query operation timeout (listing is a query)
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeQuery)
	defer cancel()

	documents, err := s.dbClient.ListDocuments(timeoutCtx, collection, limit, 0)
	if err != nil {
		return nil, s.enhanceError("failed to list documents", err)
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

	// Create context with document operation timeout
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeDocument)
	defer cancel()

	err := s.dbClient.CreateDocument(timeoutCtx, collection, doc)
	if err != nil {
		return nil, s.enhanceError("failed to create document", err)
	}

	return map[string]interface{}{
		"collection": collection,
		"url":        url,
		"text":       text,
		"metadata":   metadata,
		"status":     "created",
	}, nil
}

// handleBatchCreateDocuments handles the batch_create_documents tool
func (s *Server) handleBatchCreateDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collection, ok := args["collection"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	documentsArg, ok := args["documents"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("documents array is required")
	}

	if len(documentsArg) == 0 {
		return nil, fmt.Errorf("documents array cannot be empty")
	}

	// Parse and validate all documents first
	documents := make([]*vectordb.Document, 0, len(documentsArg))
	for i, docArg := range documentsArg {
		docMap, ok := docArg.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("document at index %d is not a valid object", i)
		}

		url, ok := docMap["url"].(string)
		if !ok {
			return nil, fmt.Errorf("document at index %d: URL is required", i)
		}

		text, ok := docMap["text"].(string)
		if !ok {
			return nil, fmt.Errorf("document at index %d: text is required", i)
		}

		metadata, _ := docMap["metadata"].(map[string]interface{})
		if metadata == nil {
			metadata = make(map[string]interface{})
		}

		doc := &vectordb.Document{
			URL:      url,
			Text:     text,
			Content:  text, // Use text as content
			Metadata: metadata,
		}
		documents = append(documents, doc)
	}

	// Create context with bulk operation timeout
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeBulk)
	defer cancel()

	// Create all documents in batch
	err := s.dbClient.CreateDocuments(timeoutCtx, collection, documents)
	if err != nil {
		return nil, s.enhanceError("failed to create documents in batch", err)
	}

	return map[string]interface{}{
		"collection": collection,
		"count":      len(documents),
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

	// Create context with document operation timeout
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeDocument)
	defer cancel()

	// Get document using vectordb client
	doc, err := s.dbClient.GetDocument(timeoutCtx, collection, documentID)
	if err != nil {
		return nil, s.enhanceError("failed to get document", err)
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

	// Create context with document operation timeout
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeDocument)
	defer cancel()

	// Delete document using vectordb client
	err := s.dbClient.DeleteDocument(timeoutCtx, collection, documentID)
	if err != nil {
		return nil, s.enhanceError("failed to delete document", err)
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

	// Create context with query operation timeout (counting is a query)
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeQuery)
	defer cancel()

	// Count documents using vectordb client
	count, err := s.dbClient.GetCollectionCount(timeoutCtx, collection)
	if err != nil {
		return nil, s.enhanceError("failed to count documents", err)
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

	// Create context with query operation timeout
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeQuery)
	defer cancel()

	// Query documents using vectordb client
	queryOptions := &vectordb.QueryOptions{
		TopK: limit,
	}

	results, err := s.dbClient.SearchSemantic(timeoutCtx, collection, query, queryOptions)
	if err != nil {
		return nil, s.enhanceError("failed to query documents", err)
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

	// Create context with document operation timeout
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeDocument)
	defer cancel()

	// Get the existing document first
	doc, err := s.dbClient.GetDocument(timeoutCtx, collection, documentID)
	if err != nil {
		return nil, s.enhanceError("failed to get existing document", err)
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

	// Update document using vectordb client (reuse same timeout context)
	err = s.dbClient.UpdateDocument(timeoutCtx, collection, doc)
	if err != nil {
		return nil, s.enhanceError("failed to update document", err)
	}

	return map[string]interface{}{
		"document_id": documentID,
		"collection":  collection,
		"status":      "updated",
	}, nil
}

// handleSuggestSchema handles the suggest_schema tool
func (s *Server) handleSuggestSchema(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	sourcePath, ok := args["source_path"].(string)
	if !ok {
		return nil, fmt.Errorf("source_path is required")
	}

	collectionName, ok := args["collection_name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection_name is required")
	}

	// Optional parameters
	requirements := ""
	if req, ok := args["requirements"].(string); ok {
		requirements = req
	}

	vdbType := "weaviate"
	if vdb, ok := args["vdb_type"].(string); ok {
		vdbType = vdb
	}

	maxSamples := "50"
	if max, ok := args["max_samples"].(float64); ok {
		maxSamples = fmt.Sprintf("%d", int(max))
	}

	// Build CLI command
	cmdParts := []string{"weave", "schema", "suggest", sourcePath, "--collection", collectionName, "--vdb", vdbType, "--max-samples", maxSamples, "--output", "json"}
	if requirements != "" {
		cmdParts = append(cmdParts, "--requirements", fmt.Sprintf("\"%s\"", requirements))
	}

	cmd := strings.Join(cmdParts, " ")

	// Create timeout context (60 seconds for AI operations)
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, 60)
	defer cancel()

	// Execute command
	result, err := s.executeCommand(timeoutCtx, cmd)
	if err != nil {
		return nil, s.enhanceError("failed to suggest schema", err)
	}

	return result, nil
}

// handleSuggestChunking handles the suggest_chunking tool
func (s *Server) handleSuggestChunking(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	sourcePath, ok := args["source_path"].(string)
	if !ok {
		return nil, fmt.Errorf("source_path is required")
	}

	collectionName, ok := args["collection_name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection_name is required")
	}

	// Optional parameters
	requirements := ""
	if req, ok := args["requirements"].(string); ok {
		requirements = req
	}

	vdbType := "weaviate"
	if vdb, ok := args["vdb_type"].(string); ok {
		vdbType = vdb
	}

	maxSamples := "50"
	if max, ok := args["max_samples"].(float64); ok {
		maxSamples = fmt.Sprintf("%d", int(max))
	}

	// Build CLI command
	cmdParts := []string{"weave", "chunking", "suggest", sourcePath, "--collection", collectionName, "--vdb", vdbType, "--max-samples", maxSamples, "--output", "json"}
	if requirements != "" {
		cmdParts = append(cmdParts, "--requirements", fmt.Sprintf("\"%s\"", requirements))
	}

	cmd := strings.Join(cmdParts, " ")

	// Create timeout context (60 seconds for AI operations)
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, 60)
	defer cancel()

	// Execute command
	result, err := s.executeCommand(timeoutCtx, cmd)
	if err != nil {
		return nil, s.enhanceError("failed to suggest chunking", err)
	}

	return result, nil
}

// executeCommand executes a shell command and returns the result
func (s *Server) executeCommand(ctx context.Context, cmdStr string) (interface{}, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("command failed: %s - %s", err, stderr.String())
	}

	// Try to parse JSON output
	var result interface{}
	if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
		// If not JSON, return raw output
		return map[string]interface{}{
			"output": stdout.String(),
		}, nil
	}

	return result, nil
}

// handleHealthCheck checks the health of the vector database
func (s *Server) handleHealthCheck(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Create timeout context (10 seconds for health check)
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, 10)
	defer cancel()

	// Get database config
	dbConfig, err := s.config.GetDefaultDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database config: %w", err)
	}

	// Check database health
	err = s.dbClient.Health(timeoutCtx)
	if err != nil {
		return map[string]interface{}{
			"status":   "unhealthy",
			"database": string(dbConfig.Type),
			"error":    err.Error(),
		}, nil
	}

	return map[string]interface{}{
		"status":   "healthy",
		"database": string(dbConfig.Type),
		"url":      dbConfig.URL,
	}, nil
}

// handleCountCollections counts the total number of collections
func (s *Server) handleCountCollections(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Create timeout context (20 seconds for collection operation)
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, 20)
	defer cancel()

	// List all collections
	collections, err := s.dbClient.ListCollections(timeoutCtx)
	if err != nil {
		return nil, s.enhanceError("failed to list collections", err)
	}

	// Extract collection names
	names := make([]string, len(collections))
	for i, col := range collections {
		names[i] = col.Name
	}

	return map[string]interface{}{
		"count":       len(collections),
		"collections": names,
	}, nil
}

// handleShowCollection shows detailed information about a collection
func (s *Server) handleShowCollection(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collectionName, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	// Create timeout context (20 seconds for collection operation)
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, 20)
	defer cancel()

	// Get collection schema
	schema, err := s.dbClient.GetSchema(timeoutCtx, collectionName)
	if err != nil {
		return nil, s.enhanceError("failed to get collection schema", err)
	}

	// Get collection count
	count, err := s.dbClient.GetCollectionCount(timeoutCtx, collectionName)
	if err != nil {
		return nil, s.enhanceError("failed to get collection count", err)
	}

	return map[string]interface{}{
		"name":       collectionName,
		"schema":     schema,
		"count":      count,
		"vectorizer": schema.Vectorizer,
		"properties": schema.Properties,
	}, nil
}

// handleListEmbeddingModels lists all available embedding models
func (s *Server) handleListEmbeddingModels(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Return list of supported embedding models
	// These are the models supported across all vector databases
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

// handleShowCollectionEmbeddings shows embedding configuration for a collection
func (s *Server) handleShowCollectionEmbeddings(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collectionName, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("collection name is required")
	}

	// Create timeout context (20 seconds for collection operation)
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, 20)
	defer cancel()

	// Get collection schema which contains vectorizer info
	schema, err := s.dbClient.GetSchema(timeoutCtx, collectionName)
	if err != nil {
		return nil, s.enhanceError("failed to get collection schema", err)
	}

	// Determine dimensions based on vectorizer
	dimensions := 1536 // default for most OpenAI models
	if schema.Vectorizer == "text-embedding-3-large" {
		dimensions = 3072
	}

	return map[string]interface{}{
		"collection": collectionName,
		"vectorizer": schema.Vectorizer,
		"model":      schema.Vectorizer,
		"dimensions": dimensions,
		"provider":   "openai",
	}, nil
}

// handleGetCollectionStats returns statistics for a collection
func (s *Server) handleGetCollectionStats(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collectionName, ok := args["name"].(string)
	if !ok || collectionName == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	// Create timeout context for collection operations
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeCollection)
	defer cancel()

	// Get collection schema/info
	schema, err := s.dbClient.GetSchema(timeoutCtx, collectionName)
	if err != nil {
		return nil, s.enhanceError("failed to get collection schema", err)
	}

	// Get document count
	count, err := s.dbClient.GetCollectionCount(timeoutCtx, collectionName)
	if err != nil {
		return nil, s.enhanceError("failed to count documents", err)
	}

	// Build stats response
	stats := map[string]interface{}{
		"collection":     collectionName,
		"document_count": count,
		"schema": map[string]interface{}{
			"vectorizer": schema.Vectorizer,
			"properties": len(schema.Properties),
		},
	}

	return stats, nil
}

// handleDeleteAllDocuments deletes all documents from a collection or all collections
func (s *Server) handleDeleteAllDocuments(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collectionName, _ := args["collection"].(string)

	// Create timeout context for delete operations
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeDocument)
	defer cancel()

	if collectionName == "" {
		// Delete all documents from all collections
		collections, err := s.dbClient.ListCollections(timeoutCtx)
		if err != nil {
			return nil, s.enhanceError("failed to list collections", err)
		}

		totalDeleted := 0
		for _, coll := range collections {
			// Get all documents in collection
			docs, err := s.dbClient.ListDocuments(timeoutCtx, coll.Name, 10000, 0) // Large limit, offset 0
			if err != nil {
				s.logger.Warn(fmt.Sprintf("Failed to list documents in %s: %v", coll.Name, err))
				continue
			}

			// Delete each document
			for _, doc := range docs {
				err := s.dbClient.DeleteDocument(timeoutCtx, coll.Name, doc.ID)
				if err != nil {
					s.logger.Warn(fmt.Sprintf("Failed to delete document %s: %v", doc.ID, err))
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
	docs, err := s.dbClient.ListDocuments(timeoutCtx, collectionName, 10000, 0) // Large limit, offset 0
	if err != nil {
		return nil, s.enhanceError("failed to list documents", err)
	}

	deletedCount := 0
	for _, doc := range docs {
		err := s.dbClient.DeleteDocument(timeoutCtx, collectionName, doc.ID)
		if err != nil {
			s.logger.Warn(fmt.Sprintf("Failed to delete document %s: %v", doc.ID, err))
			continue
		}
		deletedCount++
	}

	return map[string]interface{}{
		"collection":    collectionName,
		"deleted_count": deletedCount,
	}, nil
}

// handleShowDocumentByName shows a document by filename instead of ID
func (s *Server) handleShowDocumentByName(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collectionName, ok := args["collection"].(string)
	if !ok || collectionName == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	filename, ok := args["filename"].(string)
	if !ok || filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	// Create timeout context
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeDocument)
	defer cancel()

	// List documents and find by filename
	// We'll use a reasonable limit and search through documents
	docs, err := s.dbClient.ListDocuments(timeoutCtx, collectionName, 1000, 0)
	if err != nil {
		return nil, s.enhanceError("failed to list documents", err)
	}

	// Search for document with matching filename in metadata or URL
	for _, doc := range docs {
		// Check URL field
		if doc.URL != "" && strings.Contains(doc.URL, filename) {
			return map[string]interface{}{
				"document_id": doc.ID,
				"collection":  collectionName,
				"url":         doc.URL,
				"text":        doc.Text,
				"metadata":    doc.Metadata,
			}, nil
		}

		// Check metadata for filename field
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

// handleDeleteDocumentByName deletes a document by filename instead of ID
func (s *Server) handleDeleteDocumentByName(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	collectionName, ok := args["collection"].(string)
	if !ok || collectionName == "" {
		return nil, fmt.Errorf("collection name is required")
	}

	filename, ok := args["filename"].(string)
	if !ok || filename == "" {
		return nil, fmt.Errorf("filename is required")
	}

	// Create timeout context
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeDocument)
	defer cancel()

	// List documents and find by filename
	docs, err := s.dbClient.ListDocuments(timeoutCtx, collectionName, 1000, 0)
	if err != nil {
		return nil, s.enhanceError("failed to list documents", err)
	}

	// Search for document with matching filename
	for _, doc := range docs {
		// Check URL field
		if doc.URL != "" && strings.Contains(doc.URL, filename) {
			err := s.dbClient.DeleteDocument(timeoutCtx, collectionName, doc.ID)
			if err != nil {
				return nil, s.enhanceError("failed to delete document", err)
			}
			return map[string]interface{}{
				"document_id": doc.ID,
				"collection":  collectionName,
				"filename":    filename,
				"status":      "deleted",
			}, nil
		}

		// Check metadata for filename field
		if doc.Metadata != nil {
			if filenameVal, ok := doc.Metadata["filename"].(string); ok && filenameVal == filename {
				err := s.dbClient.DeleteDocument(timeoutCtx, collectionName, doc.ID)
				if err != nil {
					return nil, s.enhanceError("failed to delete document", err)
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

// handleExecuteQuery executes a natural language query against documents
func (s *Server) handleExecuteQuery(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return nil, fmt.Errorf("query is required")
	}

	collectionName, _ := args["collection"].(string)
	limit := 5 // default limit
	if limitArg, ok := args["limit"].(float64); ok {
		limit = int(limitArg)
	}

	// Create timeout context for query operations
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, vectordb.OperationTypeQuery)
	defer cancel()

	// If no collection specified, search across all collections
	if collectionName == "" {
		collections, err := s.dbClient.ListCollections(timeoutCtx)
		if err != nil {
			return nil, s.enhanceError("failed to list collections", err)
		}

		// Query each collection and aggregate results
		allResults := []interface{}{}
		for _, coll := range collections {
			queryOptions := &vectordb.QueryOptions{
				TopK: limit,
			}
			results, err := s.dbClient.SearchSemantic(timeoutCtx, coll.Name, query, queryOptions)
			if err != nil {
				s.logger.Warn(fmt.Sprintf("Failed to query collection %s: %v", coll.Name, err))
				continue
			}

			// Add collection name to each result
			for _, result := range results {
				resultMap := map[string]interface{}{
					"collection":  coll.Name,
					"document_id": result.Document.ID,
					"text":        result.Document.Text,
					"url":         result.Document.URL,
					"metadata":    result.Document.Metadata,
					"score":       result.Score,
				}
				allResults = append(allResults, resultMap)
			}
		}

		// Sort by score and limit
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
	queryOptions := &vectordb.QueryOptions{
		TopK: limit,
	}
	results, err := s.dbClient.SearchSemantic(timeoutCtx, collectionName, query, queryOptions)
	if err != nil {
		return nil, s.enhanceError("failed to execute query", err)
	}

	// Format results
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

// Phase 1: Observability & Monitoring tool handlers

// handleConfigureLogging configures structured logging
func (s *Server) handleConfigureLogging(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	// Parse log level
	logLevelStr, _ := args["log_level"].(string)
	if logLevelStr == "" {
		logLevelStr = "info"
	}

	logLevel, err := logging.ParseLevel(logLevelStr)
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}

	// Parse log format
	logFormatStr, _ := args["log_format"].(string)
	if logFormatStr == "" {
		logFormatStr = "text"
	}
	logFormat := logging.Format(logFormatStr)

	// Parse log file
	logFile, _ := args["log_file"].(string)

	// Initialize logging
	if err := logging.InitWithFormat(logLevel, logFormat, logFile, false); err != nil {
		return nil, fmt.Errorf("failed to configure logging: %w", err)
	}

	result := map[string]interface{}{
		"status":     "configured",
		"log_level":  logLevelStr,
		"log_format": string(logFormat),
	}

	if logFile != "" {
		result["log_file"] = logFile
	}

	return result, nil
}

// handleGetMetrics retrieves current Prometheus metrics
func (s *Server) handleGetMetrics(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	format, _ := args["format"].(string)
	if format == "" {
		format = "json"
	}

	// For now, return a summary of metrics in JSON format
	// In the future, we can parse the actual Prometheus metrics
	if format == "json" {
		return map[string]interface{}{
			"metrics_endpoint": "/metrics",
			"description":      "Prometheus metrics available at /metrics endpoint",
			"available_metrics": []string{
				"weave_request_duration_seconds",
				"weave_documents_total",
				"weave_errors_total",
				"weave_active_connections",
			},
			"labels": map[string]interface{}{
				"weave_request_duration_seconds": []string{"vdb_type", "operation", "status"},
				"weave_documents_total":          []string{"vdb_type", "operation"},
				"weave_errors_total":             []string{"vdb_type", "operation", "error_type"},
				"weave_active_connections":       []string{"vdb_type"},
			},
		}, nil
	}

	// For prometheus format, direct users to the /metrics endpoint
	return map[string]interface{}{
		"format":  "prometheus",
		"message": "Prometheus metrics available at /metrics endpoint",
		"url":     "/metrics",
	}, nil
}

// handleCheckHealth performs detailed health check
func (s *Server) handleCheckHealth(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	detailed, _ := args["detailed"].(bool)

	// Create timeout context for health check
	timeoutCtx, cancel := s.createContextWithTimeout(ctx, 10)
	defer cancel()

	// Get database config
	dbConfig, err := s.config.GetDefaultDatabase()
	if err != nil {
		return nil, fmt.Errorf("failed to get database config: %w", err)
	}

	// Perform health check
	dbErr := s.dbClient.Health(timeoutCtx)

	result := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"server":    "weave-mcp",
	}

	// Database health
	dbStatus := map[string]interface{}{
		"vdb_type": string(dbConfig.Type),
		"name":     dbConfig.Name,
		"status":   "healthy",
	}

	if dbErr != nil {
		dbStatus["status"] = "unhealthy"
		dbStatus["error"] = dbErr.Error()
		result["status"] = "degraded"
	}

	if detailed {
		dbStatus["url"] = dbConfig.URL
		dbStatus["enabled"] = dbConfig.Enabled
		if dbConfig.Timeout > 0 {
			dbStatus["timeout_seconds"] = dbConfig.Timeout
		}

		// Try to get collection count as a connectivity test
		if dbErr == nil {
			collections, err := s.dbClient.ListCollections(timeoutCtx)
			if err == nil {
				dbStatus["collections_count"] = len(collections)
			}
		}
	}

	result["database"] = dbStatus

	return result, nil
}

// Phase 2: Agent framework tool handlers

// handleListAgents lists all available agents
func (s *Server) handleListAgents(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	agentType, _ := args["agent_type"].(string)

	// Get agent registry
	registry := agents.GetDefaultAgentRegistry()

	// List all agents
	var agentList []agents.AgentInfo
	var err error

	if agentType != "" {
		// Filter by type
		agentList, err = registry.GetAgentsByType(agentType)
	} else {
		// Get all agents
		agentList, err = registry.ListAgents()
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list agents: %w", err)
	}

	// Convert to response format
	agentsResponse := make([]map[string]interface{}, 0, len(agentList))
	for _, agent := range agentList {
		agentsResponse = append(agentsResponse, map[string]interface{}{
			"name":        agent.Name,
			"type":        agent.Type,
			"description": agent.Description,
			"version":     agent.Version,
		})
	}

	return map[string]interface{}{
		"agents": agentsResponse,
		"count":  len(agentsResponse),
	}, nil
}

// handleGetAgentInfo gets detailed info about a specific agent
func (s *Server) handleGetAgentInfo(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	agentName, ok := args["agent_name"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_name is required")
	}

	// Get agent registry
	registry := agents.GetDefaultAgentRegistry()

	// Get agent info
	agentInfo, err := registry.GetAgentInfo(agentName)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent info: %w", err)
	}

	return map[string]interface{}{
		"name":        agentInfo.Name,
		"type":        agentInfo.Type,
		"description": agentInfo.Description,
		"version":     agentInfo.Version,
		"file_path":   agentInfo.FilePath,
	}, nil
}

// handleRunAgent executes an agent
func (s *Server) handleRunAgent(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	agentName, ok := args["agent_name"].(string)
	if !ok {
		return nil, fmt.Errorf("agent_name is required")
	}

	task, ok := args["task"].(string)
	if !ok {
		return nil, fmt.Errorf("task is required")
	}

	parameters, _ := args["parameters"].(map[string]interface{})
	if parameters == nil {
		parameters = make(map[string]interface{})
	}

	// Load the agent
	agentLoader := agents.GetDefaultAgentLoader()
	agentConfig, err := agentLoader.LoadAgent(agentName)
	if err != nil {
		return nil, fmt.Errorf("failed to load agent '%s': %w", agentName, err)
	}

	// For now, return agent configuration and task details
	// In the future, this could actually execute the agent
	return map[string]interface{}{
		"agent":      agentConfig.Name,
		"type":       agentConfig.Type,
		"task":       task,
		"parameters": parameters,
		"status":     "agent_loaded",
		"message":    fmt.Sprintf("Agent '%s' loaded successfully. Full agent execution will be implemented in a future update.", agentName),
		"config": map[string]interface{}{
			"description": agentConfig.Description,
			"version":     agentConfig.Version,
		},
	}, nil
}
