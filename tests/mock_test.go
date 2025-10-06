// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"context"
	"testing"

	"github.com/maximilien/weave-mcp/src/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMock(t *testing.T) {
	t.Run("NewClient", func(t *testing.T) {
		client := mock.NewClient()
		require.NotNil(t, client)
	})
	
	t.Run("Connect", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		err := client.Connect(ctx)
		assert.NoError(t, err)
	})
	
	t.Run("Close", func(t *testing.T) {
		client := mock.NewClient()
		
		err := client.Close()
		assert.NoError(t, err)
	})
	
	t.Run("CreateCollection", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Test creating a new collection
		err := client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		assert.NoError(t, err)
		
		// Test creating a collection with the same name (should fail)
		err = client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
	})
	
	t.Run("ListCollections", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Initially no collections
		collections, err := client.ListCollections(ctx)
		assert.NoError(t, err)
		assert.Empty(t, collections)
		
		// Create a collection
		err = client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		require.NoError(t, err)
		
		// Now should have one collection
		collections, err = client.ListCollections(ctx)
		assert.NoError(t, err)
		assert.Len(t, collections, 1)
		assert.Contains(t, collections, "test-collection")
	})
	
	t.Run("DeleteCollection", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Create a collection
		err := client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		require.NoError(t, err)
		
		// Delete the collection
		err = client.DeleteCollection(ctx, "test-collection")
		assert.NoError(t, err)
		
		// Try to delete non-existing collection
		err = client.DeleteCollection(ctx, "non-existing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
	
	t.Run("Insert and ListDocuments", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Create a collection
		err := client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		require.NoError(t, err)
		
		// Insert documents
		documents := []mock.Document{
			{
				ID:       "doc1",
				URL:      "https://example.com/doc1",
				Text:     "This is document 1",
				Content:  "This is document 1",
				Metadata: map[string]interface{}{"type": "test"},
			},
			{
				ID:       "doc2",
				URL:      "https://example.com/doc2",
				Text:     "This is document 2",
				Content:  "This is document 2",
				Metadata: map[string]interface{}{"type": "test"},
			},
		}
		
		err = client.Insert(ctx, "test-collection", documents)
		assert.NoError(t, err)
		
		// List documents
		docs, err := client.ListDocuments(ctx, "test-collection", 10, 0)
		assert.NoError(t, err)
		assert.Len(t, docs, 2)
		
		// Check document content (order may vary due to map iteration)
		docIDs := make([]string, len(docs))
		docURLs := make([]string, len(docs))
		docTexts := make([]string, len(docs))
		for i, doc := range docs {
			docIDs[i] = doc.ID
			docURLs[i] = doc.URL
			docTexts[i] = doc.Text
		}
		assert.Contains(t, docIDs, "doc1")
		assert.Contains(t, docIDs, "doc2")
		assert.Contains(t, docURLs, "https://example.com/doc1")
		assert.Contains(t, docURLs, "https://example.com/doc2")
		assert.Contains(t, docTexts, "This is document 1")
		assert.Contains(t, docTexts, "This is document 2")
	})
	
	t.Run("CountDocuments", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Create a collection
		err := client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		require.NoError(t, err)
		
		// Initially no documents
		count, err := client.CountDocuments(ctx, "test-collection")
		assert.NoError(t, err)
		assert.Equal(t, 0, count)
		
		// Insert documents
		documents := []mock.Document{
			{ID: "doc1", Text: "Document 1"},
			{ID: "doc2", Text: "Document 2"},
		}
		
		err = client.Insert(ctx, "test-collection", documents)
		require.NoError(t, err)
		
		// Now should have 2 documents
		count, err = client.CountDocuments(ctx, "test-collection")
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})
	
	t.Run("GetDocument", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Create a collection
		err := client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		require.NoError(t, err)
		
		// Insert a document
		documents := []mock.Document{
			{
				ID:       "doc1",
				URL:      "https://example.com/doc1",
				Text:     "This is document 1",
				Content:  "This is document 1",
				Metadata: map[string]interface{}{"type": "test"},
			},
		}
		
		err = client.Insert(ctx, "test-collection", documents)
		require.NoError(t, err)
		
		// Get the document
		doc, err := client.GetDocument(ctx, "test-collection", "doc1")
		assert.NoError(t, err)
		assert.Equal(t, "doc1", doc.ID)
		assert.Equal(t, "https://example.com/doc1", doc.URL)
		assert.Equal(t, "This is document 1", doc.Text)
		
		// Try to get non-existing document
		_, err = client.GetDocument(ctx, "test-collection", "non-existing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
	
	t.Run("DeleteDocument", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Create a collection
		err := client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		require.NoError(t, err)
		
		// Insert a document
		documents := []mock.Document{
			{ID: "doc1", Text: "Document 1"},
		}
		
		err = client.Insert(ctx, "test-collection", documents)
		require.NoError(t, err)
		
		// Delete the document
		err = client.DeleteDocument(ctx, "test-collection", "doc1")
		assert.NoError(t, err)
		
		// Try to get the deleted document
		_, err = client.GetDocument(ctx, "test-collection", "doc1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
	
	t.Run("Search", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Create a collection
		err := client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		require.NoError(t, err)
		
		// Insert documents
		documents := []mock.Document{
			{ID: "doc1", Text: "This is about machine learning"},
			{ID: "doc2", Text: "This is about artificial intelligence"},
		}
		
		err = client.Insert(ctx, "test-collection", documents)
		require.NoError(t, err)
		
		// Search for documents
		results, err := client.Search(ctx, "test-collection", "machine learning", 10)
		assert.NoError(t, err)
		assert.Len(t, results, 2) // Should return both documents
		
		// Check that results have scores
		for _, result := range results {
			assert.GreaterOrEqual(t, result.Score, 0.0)
			assert.LessOrEqual(t, result.Score, 1.0)
		}
	})
	
	t.Run("Query", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Create a collection
		err := client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		require.NoError(t, err)
		
		// Insert documents
		documents := []mock.Document{
			{ID: "doc1", Text: "This is about machine learning"},
		}
		
		err = client.Insert(ctx, "test-collection", documents)
		require.NoError(t, err)
		
		// Query documents
		result, err := client.Query(ctx, "test-collection", "machine learning", 10)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		
		// Check result structure
		resultMap, ok := result.(map[string]interface{})
		assert.True(t, ok)
		assert.Contains(t, resultMap, "results")
		assert.Contains(t, resultMap, "query")
		assert.Contains(t, resultMap, "count")
	})
	
	t.Run("GetCollectionInfo", func(t *testing.T) {
		client := mock.NewClient()
		ctx := context.Background()
		
		// Create a collection
		err := client.CreateCollection(ctx, "test-collection", "text", "Test collection")
		require.NoError(t, err)
		
		// Get collection info
		info, err := client.GetCollectionInfo(ctx, "test-collection")
		assert.NoError(t, err)
		assert.NotNil(t, info)
		assert.Equal(t, "test-collection", info["name"])
		assert.Equal(t, "text", info["type"])
		assert.Equal(t, "Test collection", info["description"])
		assert.Equal(t, 0, info["document_count"])
	})
}