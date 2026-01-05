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

	// Document mocks
	documents     []*vectordb.Document
	listDocsError error
	deleteError   error
	deletedDocs   []string // Track deleted document IDs

	// Search mocks
	searchResults []*vectordb.QueryResult
	searchError   error
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
	if m.deleteError != nil {
		return m.deleteError
	}
	m.deletedDocs = append(m.deletedDocs, documentID)
	return nil
}

func (m *mockVectorDBClient) DeleteDocuments(ctx context.Context, collectionName string, documentIDs []string) error {
	return nil
}

func (m *mockVectorDBClient) DeleteDocumentsByMetadata(ctx context.Context, collectionName string, metadata map[string]interface{}) error {
	return nil
}

func (m *mockVectorDBClient) ListDocuments(ctx context.Context, collectionName string, limit int, offset int) ([]*vectordb.Document, error) {
	if m.listDocsError != nil {
		return nil, m.listDocsError
	}
	return m.documents, nil
}

func (m *mockVectorDBClient) SearchSemantic(ctx context.Context, collectionName, query string, options *vectordb.QueryOptions) ([]*vectordb.QueryResult, error) {
	if m.searchError != nil {
		return nil, m.searchError
	}
	return m.searchResults, nil
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

// TestHandleGetCollectionStats tests the get_collection_stats handler
func TestHandleGetCollectionStats(t *testing.T) {
	t.Run("success with valid collection", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collectionSchema: &vectordb.CollectionSchema{
				Class:      "articles",
				Vectorizer: "text-embedding-3-small",
				Properties: []vectordb.SchemaProperty{
					{Name: "text", DataType: []string{"text"}},
					{Name: "url", DataType: []string{"text"}},
				},
			},
			getSchemaError:  nil,
			collectionCount: 42,
			getCountError:   nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"name": "articles",
		}

		result, err := server.handleGetCollectionStats(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "articles", response["collection"])
		assert.Equal(t, int64(42), response["document_count"])

		schema, ok := response["schema"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "text-embedding-3-small", schema["vectorizer"])
		assert.Equal(t, 2, schema["properties"])
	})

	t.Run("missing collection name", func(t *testing.T) {
		mockClient := &mockVectorDBClient{}
		server := createTestServer(mockClient)

		args := map[string]interface{}{}

		result, err := server.handleGetCollectionStats(context.Background(), args)

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

		result, err := server.handleGetCollectionStats(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to get collection schema")
	})

	t.Run("count error", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collectionSchema: &vectordb.CollectionSchema{
				Class:      "articles",
				Vectorizer: "text-embedding-3-small",
			},
			getSchemaError:  nil,
			collectionCount: 0,
			getCountError:   errors.New("count failed"),
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"name": "articles",
		}

		result, err := server.handleGetCollectionStats(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to count documents")
	})
}

// TestHandleDeleteAllDocuments tests the delete_all_documents handler
func TestHandleDeleteAllDocuments(t *testing.T) {
	t.Run("delete from specific collection", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			documents: []*vectordb.Document{
				{ID: "doc1", Text: "test1"},
				{ID: "doc2", Text: "test2"},
				{ID: "doc3", Text: "test3"},
			},
			listDocsError: nil,
			deleteError:   nil,
			deletedDocs:   []string{},
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
		}

		result, err := server.handleDeleteAllDocuments(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "articles", response["collection"])
		assert.Equal(t, 3, response["deleted_count"])
		assert.Len(t, mockClient.deletedDocs, 3)
	})

	t.Run("delete from all collections", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collections: []vectordb.CollectionInfo{
				{Name: "articles", Count: 2},
				{Name: "docs", Count: 1},
			},
			listCollError: nil,
			documents: []*vectordb.Document{
				{ID: "doc1", Text: "test1"},
				{ID: "doc2", Text: "test2"},
			},
			listDocsError: nil,
			deleteError:   nil,
			deletedDocs:   []string{},
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{}

		result, err := server.handleDeleteAllDocuments(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, 2, response["collections_cleaned"])
		assert.Equal(t, 4, response["deleted_count"]) // 2 collections × 2 docs each
	})

	t.Run("list documents error", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			documents:     nil,
			listDocsError: errors.New("failed to list documents"),
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
		}

		result, err := server.handleDeleteAllDocuments(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to list documents")
	})
}

// TestHandleShowDocumentByName tests the show_document_by_name handler
func TestHandleShowDocumentByName(t *testing.T) {
	t.Run("find document by URL", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			documents: []*vectordb.Document{
				{ID: "doc1", URL: "file1.txt", Text: "content1"},
				{ID: "doc2", URL: "file2.txt", Text: "content2"},
			},
			listDocsError: nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
			"filename":   "file1.txt",
		}

		result, err := server.handleShowDocumentByName(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "doc1", response["document_id"])
		assert.Equal(t, "articles", response["collection"])
		assert.Equal(t, "file1.txt", response["url"])
		assert.Equal(t, "content1", response["text"])
	})

	t.Run("find document by metadata filename", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			documents: []*vectordb.Document{
				{
					ID:       "doc1",
					Text:     "content1",
					Metadata: map[string]interface{}{"filename": "file1.txt"},
				},
			},
			listDocsError: nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
			"filename":   "file1.txt",
		}

		result, err := server.handleShowDocumentByName(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "doc1", response["document_id"])
		assert.Equal(t, "articles", response["collection"])
	})

	t.Run("document not found", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			documents: []*vectordb.Document{
				{ID: "doc1", URL: "file1.txt", Text: "content1"},
			},
			listDocsError: nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
			"filename":   "nonexistent.txt",
		}

		result, err := server.handleShowDocumentByName(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "document with filename 'nonexistent.txt' not found")
	})

	t.Run("missing collection name", func(t *testing.T) {
		mockClient := &mockVectorDBClient{}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"filename": "file1.txt",
		}

		result, err := server.handleShowDocumentByName(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "collection name is required")
	})

	t.Run("missing filename", func(t *testing.T) {
		mockClient := &mockVectorDBClient{}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
		}

		result, err := server.handleShowDocumentByName(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "filename is required")
	})
}

// TestHandleDeleteDocumentByName tests the delete_document_by_name handler
func TestHandleDeleteDocumentByName(t *testing.T) {
	t.Run("delete document by URL", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			documents: []*vectordb.Document{
				{ID: "doc1", URL: "file1.txt", Text: "content1"},
				{ID: "doc2", URL: "file2.txt", Text: "content2"},
			},
			listDocsError: nil,
			deleteError:   nil,
			deletedDocs:   []string{},
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
			"filename":   "file1.txt",
		}

		result, err := server.handleDeleteDocumentByName(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "doc1", response["document_id"])
		assert.Equal(t, "articles", response["collection"])
		assert.Equal(t, "file1.txt", response["filename"])
		assert.Equal(t, "deleted", response["status"])
		assert.Contains(t, mockClient.deletedDocs, "doc1")
	})

	t.Run("delete document by metadata filename", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			documents: []*vectordb.Document{
				{
					ID:       "doc1",
					Text:     "content1",
					Metadata: map[string]interface{}{"filename": "file1.txt"},
				},
			},
			listDocsError: nil,
			deleteError:   nil,
			deletedDocs:   []string{},
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
			"filename":   "file1.txt",
		}

		result, err := server.handleDeleteDocumentByName(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "doc1", response["document_id"])
		assert.Equal(t, "deleted", response["status"])
		assert.Contains(t, mockClient.deletedDocs, "doc1")
	})

	t.Run("document not found", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			documents: []*vectordb.Document{
				{ID: "doc1", URL: "file1.txt", Text: "content1"},
			},
			listDocsError: nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
			"filename":   "nonexistent.txt",
		}

		result, err := server.handleDeleteDocumentByName(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "document with filename 'nonexistent.txt' not found")
	})

	t.Run("delete error", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			documents: []*vectordb.Document{
				{ID: "doc1", URL: "file1.txt", Text: "content1"},
			},
			listDocsError: nil,
			deleteError:   errors.New("delete failed"),
			deletedDocs:   []string{},
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
			"filename":   "file1.txt",
		}

		result, err := server.handleDeleteDocumentByName(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to delete document")
	})
}

// TestHandleExecuteQuery tests the execute_query handler
func TestHandleExecuteQuery(t *testing.T) {
	t.Run("query specific collection", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			searchResults: []*vectordb.QueryResult{
				{
					Document: vectordb.Document{
						ID:       "doc1",
						Text:     "machine learning basics",
						URL:      "ml.txt",
						Metadata: map[string]interface{}{"category": "ai"},
					},
					Score: 0.95,
				},
				{
					Document: vectordb.Document{
						ID:       "doc2",
						Text:     "deep learning tutorial",
						URL:      "dl.txt",
						Metadata: map[string]interface{}{"category": "ai"},
					},
					Score: 0.87,
				},
			},
			searchError: nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"query":      "machine learning",
			"collection": "articles",
			"limit":      float64(5),
		}

		result, err := server.handleExecuteQuery(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "articles", response["collection"])
		assert.Equal(t, "machine learning", response["query"])
		assert.Equal(t, 2, response["count"])

		results, ok := response["results"].([]interface{})
		require.True(t, ok)
		assert.Len(t, results, 2)

		result1, ok := results[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "doc1", result1["document_id"])
		assert.Equal(t, "machine learning basics", result1["text"])
		assert.Equal(t, 0.95, result1["score"])
	})

	t.Run("query all collections", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			collections: []vectordb.CollectionInfo{
				{Name: "articles", Count: 1},
				{Name: "docs", Count: 1},
			},
			listCollError: nil,
			searchResults: []*vectordb.QueryResult{
				{
					Document: vectordb.Document{
						ID:   "doc1",
						Text: "test document",
					},
					Score: 0.9,
				},
			},
			searchError: nil,
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"query": "test query",
			"limit": float64(5),
		}

		result, err := server.handleExecuteQuery(context.Background(), args)

		require.NoError(t, err)
		require.NotNil(t, result)

		response, ok := result.(map[string]interface{})
		require.True(t, ok)

		assert.Equal(t, "test query", response["query"])
		results, ok := response["results"].([]interface{})
		require.True(t, ok)
		assert.Greater(t, len(results), 0)
	})

	t.Run("missing query", func(t *testing.T) {
		mockClient := &mockVectorDBClient{}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"collection": "articles",
		}

		result, err := server.handleExecuteQuery(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "query is required")
	})

	t.Run("search error", func(t *testing.T) {
		mockClient := &mockVectorDBClient{
			searchResults: nil,
			searchError:   errors.New("search failed"),
		}
		server := createTestServer(mockClient)

		args := map[string]interface{}{
			"query":      "test",
			"collection": "articles",
		}

		result, err := server.handleExecuteQuery(context.Background(), args)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to execute query")
	})
}
