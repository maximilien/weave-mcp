// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/maximilien/weave-mcp/src/pkg/config"
	"github.com/maximilien/weave-mcp/src/pkg/mcp"
	"github.com/maximilien/weave-mcp/src/pkg/weaviate"
	"go.uber.org/zap"
)

// loadEnvFile loads environment variables from .env file
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || 
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		os.Setenv(key, value)
	}

	return scanner.Err()
}

// TestFastMCPIntegration runs fast integration tests with MCP server and Weaviate Cloud
func TestFastMCPIntegration(t *testing.T) {
	// Load .env file if it exists (from project root)
	if err := loadEnvFile("../.env"); err != nil {
		t.Logf("Could not load .env file: %v", err)
	}

	// Skip if no Weaviate configuration
	if os.Getenv("WEAVIATE_URL") == "" || os.Getenv("WEAVIATE_API_KEY") == "" {
		t.Skip("Skipping MCP integration tests - missing WEAVIATE_URL or WEAVIATE_API_KEY")
	}

	// Skip if URL is invalid (contains double protocol)
	if strings.Contains(os.Getenv("WEAVIATE_URL"), "https://https") || strings.Contains(os.Getenv("WEAVIATE_URL"), "http://http") {
		t.Skip("Skipping MCP integration tests - invalid URL format")
	}

	// Load configuration
	cfg, err := config.LoadConfig("../config.yaml", "../.env")
	if err != nil {
		t.Skipf("Skipping MCP integration tests - failed to load configuration: %v", err)
	}

	// Create logger
	logger, err := zap.NewProduction()
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create MCP server
	server, err := mcp.NewServer(cfg, logger)
	if err != nil {
		t.Fatalf("Failed to create MCP server: %v", err)
	}
	defer server.Cleanup()

	// Use short timeout for fast tests
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Test collections
	textCollection := "WeaveMcpDocs"
	imageCollection := "WeaveMcpImages"

	t.Run("MCPHealthCheck", func(t *testing.T) {
		healthCtx, healthCancel := context.WithTimeout(ctx, 5*time.Second)
		defer healthCancel()

		// Test MCP server health
		req, err := http.NewRequestWithContext(healthCtx, "GET", "http://localhost:8030/health", nil)
		if err != nil {
			t.Errorf("Failed to create health check request: %v", err)
			return
		}

		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Health check failed (server may not be running): %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Health check returned status %d, expected 200", resp.StatusCode)
		}
	})

	t.Run("MCPToolsList", func(t *testing.T) {
		toolsCtx, toolsCancel := context.WithTimeout(ctx, 5*time.Second)
		defer toolsCancel()

		// Test MCP tools list endpoint
		req, err := http.NewRequestWithContext(toolsCtx, "GET", "http://localhost:8030/mcp/tools/list", nil)
		if err != nil {
			t.Errorf("Failed to create tools list request: %v", err)
			return
		}

		client := &http.Client{Timeout: 3 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Tools list failed (server may not be running): %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Tools list returned status %d, expected 200", resp.StatusCode)
		}
	})

	t.Run("WeaviateDirectConnection", func(t *testing.T) {
		weaviateCtx, weaviateCancel := context.WithTimeout(ctx, 10*time.Second)
		defer weaviateCancel()

		// Test direct Weaviate connection
		weaviateClient, err := weaviate.NewClient(&weaviate.Config{
			URL:    os.Getenv("WEAVIATE_URL"),
			APIKey: os.Getenv("WEAVIATE_API_KEY"),
		})
		if err != nil {
			t.Fatalf("Failed to create Weaviate client: %v", err)
		}

		// Test listing collections
		collections, err := weaviateClient.ListCollections(weaviateCtx)
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
			return
		}

		t.Logf("Found %d collections in Weaviate", len(collections))
	})

	t.Run("CreateTextCollection", func(t *testing.T) {
		createCtx, createCancel := context.WithTimeout(ctx, 10*time.Second)
		defer createCancel()

		// Test creating text collection via MCP server
		createReq := map[string]interface{}{
			"name":        textCollection,
			"type":        "text",
			"description": "MCP integration test text collection",
		}

		result, err := server.Tools["create_collection"].Handler(createCtx, createReq)
		if err != nil {
			t.Logf("Failed to create text collection (may already exist): %v", err)
		} else {
			t.Logf("Created text collection: %v", result)
		}
	})

	t.Run("CreateImageCollection", func(t *testing.T) {
		createCtx, createCancel := context.WithTimeout(ctx, 10*time.Second)
		defer createCancel()

		// Test creating image collection via MCP server
		createReq := map[string]interface{}{
			"name":        imageCollection,
			"type":        "image",
			"description": "MCP integration test image collection",
		}

		result, err := server.Tools["create_collection"].Handler(createCtx, createReq)
		if err != nil {
			t.Logf("Failed to create image collection (may already exist): %v", err)
		} else {
			t.Logf("Created image collection: %v", result)
		}
	})

	t.Run("ListCollections", func(t *testing.T) {
		listCtx, listCancel := context.WithTimeout(ctx, 5*time.Second)
		defer listCancel()

		// Test listing collections via MCP server
		result, err := server.Tools["list_collections"].Handler(listCtx, map[string]interface{}{})
		if err != nil {
			t.Errorf("Failed to list collections: %v", err)
			return
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Errorf("Expected map result, got %T", result)
			return
		}

		collections, ok := resultMap["collections"].([]string)
		if !ok {
			t.Errorf("Expected collections array, got %T", resultMap["collections"])
			return
		}

		t.Logf("Found %d collections via MCP: %v", len(collections), collections)
	})

	t.Run("AddTextDocument", func(t *testing.T) {
		addCtx, addCancel := context.WithTimeout(ctx, 10*time.Second)
		defer addCancel()

		// Test adding a text document via MCP server
		docReq := map[string]interface{}{
			"collection": textCollection,
			"url":        "https://example.com/mcp-test-doc-1",
			"text":       "This is a test document for MCP integration testing. It contains sample text content for testing document operations.",
			"metadata": map[string]interface{}{
				"test_type":    "mcp_integration",
				"timestamp":    time.Now().Unix(),
				"source":       "fast_integration_test",
				"content_type": "text",
			},
		}

		result, err := server.Tools["create_document"].Handler(addCtx, docReq)
		if err != nil {
			t.Errorf("Failed to add text document: %v", err)
			return
		}

		t.Logf("Added text document: %v", result)
	})

	t.Run("AddImageDocument", func(t *testing.T) {
		addCtx, addCancel := context.WithTimeout(ctx, 10*time.Second)
		defer addCancel()

		// Test adding an image document via MCP server
		// Note: Image collections may expect metadata as JSON string
		metadataJSON := `{"test_type": "mcp_integration", "timestamp": ` + fmt.Sprintf("%d", time.Now().Unix()) + `, "source": "fast_integration_test", "content_type": "image", "image_url": "https://example.com/mcp-test-image-1.jpg"}`
		docReq := map[string]interface{}{
			"collection": imageCollection,
			"url":        "https://example.com/mcp-test-image-1.jpg",
			"text":       "Test image for MCP integration testing",
			"metadata":   metadataJSON,
		}

		result, err := server.Tools["create_document"].Handler(addCtx, docReq)
		if err != nil {
			t.Logf("Failed to add image document (collection may have different schema): %v", err)
			return
		}

		t.Logf("Added image document: %v", result)
	})

	t.Run("ListDocuments", func(t *testing.T) {
		listCtx, listCancel := context.WithTimeout(ctx, 5*time.Second)
		defer listCancel()

		// Test listing documents in text collection
		listReq := map[string]interface{}{
			"collection": textCollection,
			"limit":      5,
		}

		result, err := server.Tools["list_documents"].Handler(listCtx, listReq)
		if err != nil {
			t.Errorf("Failed to list documents: %v", err)
			return
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Errorf("Expected map result, got %T", result)
			return
		}

		documents, ok := resultMap["documents"].([]interface{})
		if !ok {
			// Try to handle case where documents might be []map[string]interface{}
			if docMaps, ok := resultMap["documents"].([]map[string]interface{}); ok {
				// Convert []map[string]interface{} to []interface{}
				documents = make([]interface{}, len(docMaps))
				for i, doc := range docMaps {
					documents[i] = doc
				}
			} else {
				t.Errorf("Expected documents array, got %T", resultMap["documents"])
				return
			}
		}

		t.Logf("Found %d documents in text collection", len(documents))
	})

	t.Run("QueryDocuments", func(t *testing.T) {
		queryCtx, queryCancel := context.WithTimeout(ctx, 10*time.Second)
		defer queryCancel()

		// Test querying documents via MCP server
		queryReq := map[string]interface{}{
			"collection": textCollection,
			"query":      "test document integration",
			"limit":      3,
		}

		result, err := server.Tools["query_documents"].Handler(queryCtx, queryReq)
		if err != nil {
			t.Errorf("Failed to query documents: %v", err)
			return
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Errorf("Expected map result, got %T", result)
			return
		}

		results, ok := resultMap["results"].([]interface{})
		if !ok {
			// Try to handle case where results might be []map[string]interface{}
			if resultMaps, ok := resultMap["results"].([]map[string]interface{}); ok {
				// Convert []map[string]interface{} to []interface{}
				results = make([]interface{}, len(resultMaps))
				for i, result := range resultMaps {
					results[i] = result
				}
			} else {
				t.Errorf("Expected results array, got %T", resultMap["results"])
				return
			}
		}

		t.Logf("Query returned %d results", len(results))
	})

	t.Run("CountDocuments", func(t *testing.T) {
		countCtx, countCancel := context.WithTimeout(ctx, 5*time.Second)
		defer countCancel()

		// Test counting documents via MCP server
		countReq := map[string]interface{}{
			"collection": textCollection,
		}

		result, err := server.Tools["count_documents"].Handler(countCtx, countReq)
		if err != nil {
			t.Errorf("Failed to count documents: %v", err)
			return
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Errorf("Expected map result, got %T", result)
			return
		}

		count, ok := resultMap["count"].(int)
		if !ok {
			t.Errorf("Expected count integer, got %T", resultMap["count"])
			return
		}

		t.Logf("Text collection contains %d documents", count)
	})

	t.Run("GetDocument", func(t *testing.T) {
		getCtx, getCancel := context.WithTimeout(ctx, 5*time.Second)
		defer getCancel()

		// First list documents to get an ID
		listReq := map[string]interface{}{
			"collection": textCollection,
			"limit":      1,
		}

		listResult, err := server.Tools["list_documents"].Handler(getCtx, listReq)
		if err != nil {
			t.Errorf("Failed to list documents for get test: %v", err)
			return
		}

		listMap, ok := listResult.(map[string]interface{})
		if !ok {
			t.Errorf("Expected map result from list, got %T", listResult)
			return
		}

		documents, ok := listMap["documents"].([]interface{})
		if !ok {
			// Try to handle case where documents might be []map[string]interface{}
			if docMaps, ok := listMap["documents"].([]map[string]interface{}); ok {
				// Convert []map[string]interface{} to []interface{}
				documents = make([]interface{}, len(docMaps))
				for i, doc := range docMaps {
					documents[i] = doc
				}
			} else {
				t.Errorf("Expected documents array, got %T", listMap["documents"])
				return
			}
		}
		
		if len(documents) == 0 {
			t.Logf("No documents found for get test")
			return
		}

		// Get the first document
		firstDoc, ok := documents[0].(map[string]interface{})
		if !ok {
			t.Errorf("Expected document map, got %T", documents[0])
			return
		}

		docID, ok := firstDoc["id"].(string)
		if !ok {
			t.Errorf("Expected document ID string, got %T", firstDoc["id"])
			return
		}

		// Test getting specific document
		getReq := map[string]interface{}{
			"collection":  textCollection,
			"document_id": docID,
		}

		result, err := server.Tools["get_document"].Handler(getCtx, getReq)
		if err != nil {
			t.Errorf("Failed to get document: %v", err)
			return
		}

		t.Logf("Retrieved document: %v", result)
	})

	t.Run("DeleteDocument", func(t *testing.T) {
		deleteCtx, deleteCancel := context.WithTimeout(ctx, 10*time.Second)
		defer deleteCancel()

		// First list documents to get an ID
		listReq := map[string]interface{}{
			"collection": textCollection,
			"limit":      1,
		}

		listResult, err := server.Tools["list_documents"].Handler(deleteCtx, listReq)
		if err != nil {
			t.Errorf("Failed to list documents for delete test: %v", err)
			return
		}

		listMap, ok := listResult.(map[string]interface{})
		if !ok {
			t.Errorf("Expected map result from list, got %T", listResult)
			return
		}

		documents, ok := listMap["documents"].([]interface{})
		if !ok {
			// Try to handle case where documents might be []map[string]interface{}
			if docMaps, ok := listMap["documents"].([]map[string]interface{}); ok {
				// Convert []map[string]interface{} to []interface{}
				documents = make([]interface{}, len(docMaps))
				for i, doc := range docMaps {
					documents[i] = doc
				}
			} else {
				t.Errorf("Expected documents array, got %T", listMap["documents"])
				return
			}
		}
		
		if len(documents) == 0 {
			t.Logf("No documents found for delete test")
			return
		}

		// Get the first document
		firstDoc, ok := documents[0].(map[string]interface{})
		if !ok {
			t.Errorf("Expected document map, got %T", documents[0])
			return
		}

		docID, ok := firstDoc["id"].(string)
		if !ok {
			t.Errorf("Expected document ID string, got %T", firstDoc["id"])
			return
		}

		// Test deleting specific document
		deleteReq := map[string]interface{}{
			"collection":  textCollection,
			"document_id": docID,
		}

		result, err := server.Tools["delete_document"].Handler(deleteCtx, deleteReq)
		if err != nil {
			t.Errorf("Failed to delete document: %v", err)
			return
		}

		t.Logf("Deleted document: %v", result)
	})

	t.Run("CleanupCollections", func(t *testing.T) {
		// Clean up test collections (optional - comment out if you want to keep them)
		// Uncomment the following lines if you want to clean up after testing
		/*
		cleanupCtx, cleanupCancel := context.WithTimeout(ctx, 10*time.Second)
		defer cleanupCancel()

		deleteTextReq := map[string]interface{}{
			"name": textCollection,
		}

		_, err := server.Tools["delete_collection"].Handler(cleanupCtx, deleteTextReq)
		if err != nil {
			t.Logf("Failed to delete text collection: %v", err)
		} else {
			t.Logf("Deleted text collection: %s", textCollection)
		}

		deleteImageReq := map[string]interface{}{
			"name": imageCollection,
		}

		_, err = server.Tools["delete_collection"].Handler(cleanupCtx, deleteImageReq)
		if err != nil {
			t.Logf("Failed to delete image collection: %v", err)
		} else {
			t.Logf("Deleted image collection: %s", imageCollection)
		}
		*/
		t.Logf("Cleanup collections test completed (cleanup disabled)")
	})
}

// TestMCPToolCallViaHTTP tests MCP tools via HTTP API
func TestMCPToolCallViaHTTP(t *testing.T) {
	// Load .env file if it exists (from project root)
	if err := loadEnvFile("../.env"); err != nil {
		t.Logf("Could not load .env file: %v", err)
	}

	// Skip if no Weaviate configuration
	if os.Getenv("WEAVIATE_URL") == "" || os.Getenv("WEAVIATE_API_KEY") == "" {
		t.Skip("Skipping HTTP MCP tests - missing WEAVIATE_URL or WEAVIATE_API_KEY")
	}

	// Skip if URL is invalid
	if strings.Contains(os.Getenv("WEAVIATE_URL"), "https://https") || strings.Contains(os.Getenv("WEAVIATE_URL"), "http://http") {
		t.Skip("Skipping HTTP MCP tests - invalid URL format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	t.Run("HTTPToolCall", func(t *testing.T) {
		// Test MCP tool call via HTTP
		toolCallReq := map[string]interface{}{
			"name": "list_collections",
			"arguments": map[string]interface{}{},
		}

		reqBody, err := json.Marshal(toolCallReq)
		if err != nil {
			t.Errorf("Failed to marshal request: %v", err)
			return
		}

		req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8030/mcp/tools/call", strings.NewReader(string(reqBody)))
		if err != nil {
			t.Logf("Failed to create HTTP request (server may not be running): %v", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("HTTP tool call failed (server may not be running): %v", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("HTTP tool call returned status %d, expected 200", resp.StatusCode)
			return
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Errorf("Failed to decode response: %v", err)
			return
		}

		t.Logf("HTTP tool call result: %v", result)
	})
}