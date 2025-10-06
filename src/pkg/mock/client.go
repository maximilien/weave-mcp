// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package mock

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Client represents a mock vector database client
type Client struct {
	collections map[string]*Collection
	mu          sync.RWMutex
}

// Collection represents a mock collection
type Collection struct {
	Name        string
	Type        string
	Description string
	Documents   map[string]*Document
	mu          sync.RWMutex
}

// Document represents a mock document
type Document struct {
	ID       string                 `json:"id"`
	URL      string                 `json:"url"`
	Text     string                 `json:"text"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
	Vector   []float64              `json:"vector,omitempty"`
}

// SearchResult represents a mock search result
type SearchResult struct {
	Document Document `json:"document"`
	Score    float64  `json:"score"`
}

// NewClient creates a new mock client
func NewClient() *Client {
	return &Client{
		collections: make(map[string]*Collection),
	}
}

// Connect simulates connecting to the mock database
func (c *Client) Connect(ctx context.Context) error {
	// Simulate connection time
	time.Sleep(50 * time.Millisecond)
	return nil
}

// Close simulates closing the mock database connection
func (c *Client) Close() error {
	return nil
}

// ListCollections returns all collections
func (c *Client) ListCollections(ctx context.Context) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var collections []string
	for name := range c.collections {
		collections = append(collections, name)
	}
	return collections, nil
}

// CreateCollection creates a new collection
func (c *Client) CreateCollection(ctx context.Context, name, collectionType, description string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.collections[name]; exists {
		return fmt.Errorf("collection '%s' already exists", name)
	}

	c.collections[name] = &Collection{
		Name:        name,
		Type:        collectionType,
		Description: description,
		Documents:   make(map[string]*Document),
	}

	return nil
}

// DeleteCollection deletes a collection
func (c *Client) DeleteCollection(ctx context.Context, name string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.collections[name]; !exists {
		return fmt.Errorf("collection '%s' not found", name)
	}

	delete(c.collections, name)
	return nil
}

// Insert inserts documents into a collection
func (c *Client) Insert(ctx context.Context, collectionName string, documents []Document) error {
	c.mu.RLock()
	collection, exists := c.collections[collectionName]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("collection '%s' not found", collectionName)
	}

	collection.mu.Lock()
	defer collection.mu.Unlock()

	for _, doc := range documents {
		if doc.ID == "" {
			doc.ID = generateID()
		}
		collection.Documents[doc.ID] = &doc
	}

	return nil
}

// ListDocuments lists documents from a collection
func (c *Client) ListDocuments(ctx context.Context, collectionName string, limit, offset int) ([]Document, error) {
	c.mu.RLock()
	collection, exists := c.collections[collectionName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("collection '%s' not found", collectionName)
	}

	collection.mu.RLock()
	defer collection.mu.RUnlock()

	var documents []Document
	count := 0
	for _, doc := range collection.Documents {
		if count >= offset && count < offset+limit {
			documents = append(documents, *doc)
		}
		count++
	}

	return documents, nil
}

// CountDocuments counts documents in a collection
func (c *Client) CountDocuments(ctx context.Context, collectionName string) (int, error) {
	c.mu.RLock()
	collection, exists := c.collections[collectionName]
	c.mu.RUnlock()

	if !exists {
		return 0, fmt.Errorf("collection '%s' not found", collectionName)
	}

	collection.mu.RLock()
	defer collection.mu.RUnlock()

	return len(collection.Documents), nil
}

// GetDocument retrieves a specific document by ID
func (c *Client) GetDocument(ctx context.Context, collectionName, documentID string) (*Document, error) {
	c.mu.RLock()
	collection, exists := c.collections[collectionName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("collection '%s' not found", collectionName)
	}

	collection.mu.RLock()
	defer collection.mu.RUnlock()

	doc, exists := collection.Documents[documentID]
	if !exists {
		return nil, fmt.Errorf("document '%s' not found in collection '%s'", documentID, collectionName)
	}

	return doc, nil
}

// DeleteDocument deletes a document by ID
func (c *Client) DeleteDocument(ctx context.Context, collectionName, documentID string) error {
	c.mu.RLock()
	collection, exists := c.collections[collectionName]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("collection '%s' not found", collectionName)
	}

	collection.mu.Lock()
	defer collection.mu.Unlock()

	if _, exists := collection.Documents[documentID]; !exists {
		return fmt.Errorf("document '%s' not found in collection '%s'", documentID, collectionName)
	}

	delete(collection.Documents, documentID)
	return nil
}

// Search performs a mock search
func (c *Client) Search(ctx context.Context, collectionName, query string, limit int) ([]SearchResult, error) {
	c.mu.RLock()
	collection, exists := c.collections[collectionName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("collection '%s' not found", collectionName)
	}

	collection.mu.RLock()
	defer collection.mu.RUnlock()

	var results []SearchResult
	count := 0
	for _, doc := range collection.Documents {
		if count >= limit {
			break
		}

		// Simple mock scoring based on query length and document text length
		score := 0.5 + rand.Float64()*0.5 // Random score between 0.5 and 1.0
		if len(query) > 0 && len(doc.Text) > 0 {
			// Boost score if query appears in text
			if containsIgnoreCase(doc.Text, query) {
				score = 0.8 + rand.Float64()*0.2 // Higher score for matches
			}
		}

		results = append(results, SearchResult{
			Document: *doc,
			Score:    score,
		})
		count++
	}

	return results, nil
}

// Query performs a mock query
func (c *Client) Query(ctx context.Context, collectionName, query string, limit int) (interface{}, error) {
	results, err := c.Search(ctx, collectionName, query, limit)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"results": results,
		"query":   query,
		"count":   len(results),
	}, nil
}

// GetCollectionInfo returns information about a collection
func (c *Client) GetCollectionInfo(ctx context.Context, collectionName string) (map[string]interface{}, error) {
	c.mu.RLock()
	collection, exists := c.collections[collectionName]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("collection '%s' not found", collectionName)
	}

	collection.mu.RLock()
	defer collection.mu.RUnlock()

	return map[string]interface{}{
		"name":           collection.Name,
		"type":           collection.Type,
		"description":    collection.Description,
		"document_count": len(collection.Documents),
	}, nil
}

// Helper functions

func generateID() string {
	return fmt.Sprintf("doc_%d_%d", time.Now().UnixNano(), rand.Intn(10000))
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
