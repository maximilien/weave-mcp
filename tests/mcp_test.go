// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"testing"

	"github.com/maximilien/weave-mcp/src/pkg/config"
	"github.com/maximilien/weave-mcp/src/pkg/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMCP(t *testing.T) {
	t.Run("NewServer", func(t *testing.T) {
		// Create a test config
		cfg := &config.Config{
			Databases: config.DatabasesConfig{
				Default: "mock",
				VectorDatabases: []config.VectorDBConfig{
					{
						Name:               "mock",
						Type:               config.VectorDBTypeMock,
						Enabled:            true,
						SimulateEmbeddings: true,
						EmbeddingDimension: 384,
						Collections: []config.Collection{
							{
								Name:        "TestCollection",
								Type:        "text",
								Description: "Test collection",
							},
						},
					},
				},
			},
		}
		
		// Create logger
		logger, err := zap.NewProduction()
		require.NoError(t, err)
		
		// Create mock server for testing
		server, err := mcp.NewMockServer(cfg, logger)
		require.NoError(t, err)
		require.NotNil(t, server)
		
		// Check that tools are registered
		assert.NotEmpty(t, server.Tools)
		assert.Contains(t, server.Tools, "list_collections")
		assert.Contains(t, server.Tools, "create_collection")
		assert.Contains(t, server.Tools, "delete_collection")
		assert.Contains(t, server.Tools, "list_documents")
		assert.Contains(t, server.Tools, "create_document")
		assert.Contains(t, server.Tools, "get_document")
		assert.Contains(t, server.Tools, "delete_document")
		assert.Contains(t, server.Tools, "count_documents")
		assert.Contains(t, server.Tools, "query_documents")
	})
	
	t.Run("Tool Registration", func(t *testing.T) {
		cfg := &config.Config{
			Databases: config.DatabasesConfig{
				Default: "mock",
				VectorDatabases: []config.VectorDBConfig{
					{
						Name:               "mock",
						Type:               config.VectorDBTypeMock,
						Enabled:            true,
						SimulateEmbeddings: true,
						EmbeddingDimension: 384,
					},
				},
			},
		}
		
		logger, err := zap.NewProduction()
		require.NoError(t, err)
		
		server, err := mcp.NewMockServer(cfg, logger)
		require.NoError(t, err)
		
		// Test tool properties
		tool := server.Tools["list_collections"]
		assert.Equal(t, "list_collections", tool.Name)
		assert.NotEmpty(t, tool.Description)
		assert.NotNil(t, tool.InputSchema)
		assert.NotNil(t, tool.Handler)
	})
	
	t.Run("Handler Execution", func(t *testing.T) {
		cfg := &config.Config{
			Databases: config.DatabasesConfig{
				Default: "mock",
				VectorDatabases: []config.VectorDBConfig{
					{
						Name:               "mock",
						Type:               config.VectorDBTypeMock,
						Enabled:            true,
						SimulateEmbeddings: true,
						EmbeddingDimension: 384,
					},
				},
			},
		}
		
		logger, err := zap.NewProduction()
		require.NoError(t, err)
		
		server, err := mcp.NewMockServer(cfg, logger)
		require.NoError(t, err)
		
		ctx := context.Background()
		
		// Test list_collections tool
		tool := server.Tools["list_collections"]
		result, err := tool.Handler(ctx, map[string]interface{}{})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		
		// Test create_collection tool
		tool = server.Tools["create_collection"]
		result, err = tool.Handler(ctx, map[string]interface{}{
			"name":        "test-collection",
			"type":        "text",
			"description": "Test collection",
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		
		// Test list_documents tool
		tool = server.Tools["list_documents"]
		result, err = tool.Handler(ctx, map[string]interface{}{
			"collection": "test-collection",
			"limit":      10,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		
		// Test create_document tool
		tool = server.Tools["create_document"]
		result, err = tool.Handler(ctx, map[string]interface{}{
			"collection": "test-collection",
			"url":        "https://example.com/doc1",
			"text":       "This is a test document",
			"metadata": map[string]interface{}{
				"type": "test",
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		
		// Test count_documents tool
		tool = server.Tools["count_documents"]
		result, err = tool.Handler(ctx, map[string]interface{}{
			"collection": "test-collection",
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
		
		// Test query_documents tool
		tool = server.Tools["query_documents"]
		result, err = tool.Handler(ctx, map[string]interface{}{
			"collection": "test-collection",
			"query":      "test document",
			"limit":      5,
		})
		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
	
	t.Run("Error Handling", func(t *testing.T) {
		cfg := &config.Config{
			Databases: config.DatabasesConfig{
				Default: "mock",
				VectorDatabases: []config.VectorDBConfig{
					{
						Name:               "mock",
						Type:               config.VectorDBTypeMock,
						Enabled:            true,
						SimulateEmbeddings: true,
						EmbeddingDimension: 384,
					},
				},
			},
		}
		
		logger, err := zap.NewProduction()
		require.NoError(t, err)
		
		server, err := mcp.NewMockServer(cfg, logger)
		require.NoError(t, err)
		
		ctx := context.Background()
		
		// Test create_collection with missing name
		tool := server.Tools["create_collection"]
		_, err = tool.Handler(ctx, map[string]interface{}{
			"type": "text",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "collection name is required")
		
		// Test create_document with missing collection
		tool = server.Tools["create_document"]
		_, err = tool.Handler(ctx, map[string]interface{}{
			"url":  "https://example.com/doc1",
			"text": "This is a test document",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "collection name is required")
		
		// Test get_document with missing document_id
		tool = server.Tools["get_document"]
		_, err = tool.Handler(ctx, map[string]interface{}{
			"collection": "test-collection",
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "document ID is required")
	})
	
	t.Run("Cleanup", func(t *testing.T) {
		cfg := &config.Config{
			Databases: config.DatabasesConfig{
				Default: "mock",
				VectorDatabases: []config.VectorDBConfig{
					{
						Name:               "mock",
						Type:               config.VectorDBTypeMock,
						Enabled:            true,
						SimulateEmbeddings: true,
						EmbeddingDimension: 384,
					},
				},
			},
		}
		
		logger, err := zap.NewProduction()
		require.NoError(t, err)
		
		server, err := mcp.NewMockServer(cfg, logger)
		require.NoError(t, err)
		
		// Test cleanup
		err = server.Cleanup()
		assert.NoError(t, err)
	})
}