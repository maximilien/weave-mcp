// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mcp

import (
	"context"
	"errors"
	"testing"

	"github.com/maximilien/weave-cli/src/pkg/vectordb"
	"github.com/maximilien/weave-mcp/src/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockVectorDBClient is a mock implementation of vectordb.VectorDBClient for testing
type mockVectorDBClient struct {
	// Health mock
	healthError error

	// Collection mocks
	collections      []vectordb.CollectionInfo
	listCollError    error
	collectionSchema *vectordb.CollectionSchema
	getSchemaError   error
	collectionCount  int64
	getCountError    error
}

func (m *mockVectorDBClient) Health(ctx context.Context) error {
	return m.healthError
}

func (m *mockVectorDBClient) ListCollections(ctx context.Context) ([]vectordb.CollectionInfo, error) {
	return m.collections, m.listCollError
}

func (m *mockVectorDBClient) GetSchema(ctx context.Context, collectionName string) (*vectordb.CollectionSchema, error) {
	return m.collectionSchema, m.getSchemaError
}

func (m *mockVectorDBClient) GetCollectionCount(ctx context.Context, name string) (int64, error) {
	return m.collectionCount, m.getCountError
}

// Implement remaining interface methods (not used in tests, but required for interface compliance)
func (m *mockVectorDBClient) CreateCollection(ctx context.Context, name string, schema *vectordb.CollectionSchema) error {
	return nil
}

func (m *mockVectorDBClient) DeleteCollection(ctx context.Context, name string) error {
	return nil
}

func (m *mockVectorDBClient) CollectionExists(ctx context.Context, name string) (bool, error) {
	return false, nil
}

func (m *mockVectorDBClient) CreateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	return nil
}

func (m *mockVectorDBClient) CreateDocuments(ctx context.Context, collectionName string, documents []*vectordb.Document) error {
	return nil
}

func (m *mockVectorDBClient) GetDocument(ctx context.Context, collectionName, documentID string) (*vectordb.Document, error) {
	return nil, nil
}

func (m *mockVectorDBClient) UpdateDocument(ctx context.Context, collectionName string, document *vectordb.Document) error {
	return nil
}

func (m *mockVectorDBClient) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	return nil
}

func (m *mockVectorDBClient) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	return nil
}

func (m *mockVectorDBClient) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	return nil
}

func (m *mockVectorDBClient) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	return nil, nil
}

func (m *mockVectorDBClient) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, nil
}

func (m *mockVectorDBClient) SearchBM25(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, nil
}

func (m *mockVectorDBClient) SearchHybrid(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, nil
}

func (m *mockVectorDBClient) SearchByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	return nil, nil
}

func (m *mockVectorDBClient) UpdateSchema(ctx context.Context, collectionName string, schema *vectordb.CollectionSchema) error {
	return nil
}

func (m *mockVectorDBClient) GetDefaultSchema(schemaType vectordb.SchemaType, collectionName string) *vectordb.CollectionSchema {
	return nil
}

func (m *mockVectorDBClient) ValidateSchema(schema *vectordb.CollectionSchema) error {
	return nil
}

// createTestServer creates a test server with a mock client
func createTestServer(mockClient vectordb.VectorDBClient) *Server {
	cfg := &config.Config{
		Databases: config.DatabasesConfig{
			Default: "mock",
			VectorDatabases: []config.VectorDBConfig{
				{
					Name:    "mock",
					Type:    config.VectorDBTypeMock,
					URL:     "http://localhost:8080",
					Enabled: true,
				},
			},
		},
	}

	logger, _ := zap.NewDevelopment()

	server := &Server{
		config:   cfg,
		logger:   logger,
		dbClient: mockClient,
		Tools:    make(map[string]Tool),
	}

	return server
}

// TestHandleHealthCheck tests the health_check handler
func TestHandleHealthCheck(t *testing.T) {
	t.Run("healthy database", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			healthError: nil,
		}
		server := createTestServer(mockClient)

		result, err := server.handleHealthCheck(context.Background(), map[string]interface{}{})

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "healthy", response["status"])
		assert.Equal(t, "mock", response["database"])
		assert.Equal(t, "http://localhost:8080", response["url"])
	})

	t.Run("unhealthy database", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			healthError: errors.New("connection refused"),
		}
		server := createTestServer(mockClient)

		result, err := server.handleHealthCheck(context.Background(), map[string]interface{}{})

		require.NoError(t, err) // Handler returns error in response, not as error
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "unhealthy", response["status"])
		assert.Equal(t, "mock", response["database"])
		assert.Contains(t, response["error"], "connection refused")
	})
}

// TestHandleCountCollections tests the count_collections handler
func TestHandleCountCollections(t *testing.T) {
	t.Run("success with collections", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collections: []vectordb.CollectionInfo{
				{Name: "articles", Count: 10},
				{Name: "documents", Count: 25},
				{Name: "images", Count: 5},
			},
			listCollError: nil,
		}
		server := createTestServer(mockClient)

		result, err := server.handleCountCollections(context.Background(), map[string]interface{}{})

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, 3, response["count"])

		collections, ok := response["collections"].([]string)
		require.True(t, ok)
		assert.Len(t, collections, 3)
		assert.Contains(t, collections, "articles")
		assert.Contains(t, collections, "documents")
		assert.Contains(t, collections, "images")
	})

	t.Run("success with no collections", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collections:   []vectordb.CollectionInfo{},
			listCollError: nil,
		}
		server := createTestServer(mockClient)

		result, err := server.handleCountCollections(context.Background(), map[string]interface{}{})

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, 0, response["count"])
		collections, ok := response["collections"].([]string)
		require.True(t, ok)
		assert.Len(t, collections, 0)
	})

	t.Run("database error", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collections:   nil,
			listCollError: errors.New("database connection failed"),
		}
		server := createTestServer(mockClient)

		result, err := server.handleCountCollections(context.Background(), map[string]interface{}{})

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to list collections")
	})
}

// TestHandleShowCollection tests the show_collection handler
func TestHandleShowCollection(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collectionSchema: &vectordb.CollectionSchema{
				Class:      "articles",
				Vectorizer: "text-embedding-3-small",
				Properties: []vectordb.SchemaProperty{
					{
						Name:     "text",
						DataType: []string{"text"},
					},
					{
						Name:     "url",
						DataType: []string{"string"},
					},
				},
			},
			getSchemaError:  nil,
			collectionCount: 150,
			getCountError:   nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"name": "articles",
		}

		result, err := server.handleShowCollection(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "articles", response["name"])
		assert.Equal(t, int64(150), response["count"])
		assert.Equal(t, "text-embedding-3-small", response["vectorizer"])

		schema, ok := response["schema"].(*vectordb.CollectionSchema)
		require.True(t, ok)
		assert.Equal(t, "articles", schema.Class)
		assert.Len(t, schema.Properties, 2)
	})

	t.Run("missing collection name", func(t *testing.T) {
		mockClient := &mockVectorDBClient{}
		server := createTestServer(mockClient)

		args := map[string]interface{}{}

		result, err := server.handleShowCollection(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "collection name is required")
	})

	t.Run("schema error", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collectionSchema: nil,
			getSchemaError:   errors.New("collection not found"),
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"name": "nonexistent",
		}

		result, err := server.handleShowCollection(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get collection schema")
	})

	t.Run("count error", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collectionSchema: &vectordb.CollectionSchema{
				Class: "articles",
			},
			getSchemaError:  nil,
			collectionCount: 0,
			getCountError:   errors.New("count failed"),
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"name": "articles",
		}

		result, err := server.handleShowCollection(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get collection count")
	})
}

// TestHandleListEmbeddingModels tests the list_embedding_models handler
func TestHandleListEmbeddingModels(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := &mockVectorDBClient{}
		server := createTestServer(mockClient)

		result, err := server.handleListEmbeddingModels(context.Background(), map[string]interface{}{})

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, 4, response["count"])

		models, ok := response["models"].([]map[string]interface{})
		require.True(t, ok)
		assert.Len(t, models, 4)

		// Verify model details
		foundSmall := false
		foundLarge := false
		for _, model := range models {
			name := model["name"].(string)
			if name == "text-embedding-3-small" {
				foundSmall = true
				assert.Equal(t, "openai", model["type"])
				assert.Equal(t, 1536, model["dimensions"])
				assert.Equal(t, "openai", model["provider"])
				assert.Contains(t, model["description"], "faster and cheaper")
			}
			if name == "text-embedding-3-large" {
				foundLarge = true
				assert.Equal(t, "openai", model["type"])
				assert.Equal(t, 3072, model["dimensions"])
				assert.Equal(t, "openai", model["provider"])
				assert.Contains(t, model["description"], "better quality")
			}
		}

		assert.True(t, foundSmall, "Should include text-embedding-3-small")
		assert.True(t, foundLarge, "Should include text-embedding-3-large")
	})
}

// TestHandleShowCollectionEmbeddings tests the show_collection_embeddings handler
func TestHandleShowCollectionEmbeddings(t *testing.T) {
	t.Run("success with small model", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collectionSchema: &vectordb.CollectionSchema{
				Class:      "articles",
				Vectorizer: "text-embedding-3-small",
			},
			getSchemaError: nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"name": "articles",
		}

		result, err := server.handleShowCollectionEmbeddings(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "articles", response["collection"])
		assert.Equal(t, "text-embedding-3-small", response["vectorizer"])
		assert.Equal(t, "text-embedding-3-small", response["model"])
		assert.Equal(t, 1536, response["dimensions"])
		assert.Equal(t, "openai", response["provider"])
	})

	t.Run("success with large model", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collectionSchema: &vectordb.CollectionSchema{
				Class:      "documents",
				Vectorizer: "text-embedding-3-large",
			},
			getSchemaError: nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"name": "documents",
		}

		result, err := server.handleShowCollectionEmbeddings(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "documents", response["collection"])
		assert.Equal(t, "text-embedding-3-large", response["vectorizer"])
		assert.Equal(t, 3072, response["dimensions"]) // Large model has 3072 dimensions
	})

	t.Run("missing collection name", func(t *testing.T) {
		mockClient := &mockVectorDBClient{}
		server := createTestServer(mockClient)

		args := map[string]interface{}{}

		result, err := server.handleShowCollectionEmbeddings(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "collection name is required")
	})

	t.Run("schema error", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collectionSchema: nil,
			getSchemaError:   errors.New("collection not found"),
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"name": "nonexistent",
		}

		result, err := server.handleShowCollectionEmbeddings(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get collection schema")
	})
}
