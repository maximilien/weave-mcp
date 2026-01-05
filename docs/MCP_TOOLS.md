# MCP Tools Reference

Complete reference for all 23 MCP tools provided by the Weave MCP Server.

**Version:** 0.6.0 (in development)
**Last Updated:** 2026-01-05

---

## Quick Reference

| Tool | Category | Parameters | Description |
|------|----------|------------|-------------|
| `list_collections` | Collections | none | List all collections |
| `create_collection` | Collections | name, type | Create new collection |
| `delete_collection` | Collections | name | Delete collection |
| `count_collections` | Collections | none | Count collections |
| `show_collection` | Collections | name | Show collection details |
| `get_collection_stats` | Collections | name | Get collection statistics |
| `list_documents` | Documents | collection, limit | List documents |
| `create_document` | Documents | collection, url, text, metadata | Create document |
| `batch_create_documents` | Documents | collection, documents | Batch create documents |
| `get_document` | Documents | collection, id | Get document by ID |
| `update_document` | Documents | collection, id, text, metadata | Update document |
| `delete_document` | Documents | collection, id | Delete document |
| `count_documents` | Documents | collection | Count documents |
| `show_document_by_name` | Documents | collection, filename | Show document by name |
| `delete_document_by_name` | Documents | collection, filename | Delete document by name |
| `delete_all_documents` | Documents | collection (optional) | Delete all documents |
| `query_documents` | Query | collection, query, top_k | Semantic search |
| `execute_query` | Query | query, collection, limit | Execute semantic query |
| `suggest_schema` | AI | source_path, collection_name | AI schema suggestions |
| `suggest_chunking` | AI | source_path, collection_name | AI chunking suggestions |
| `health_check` | Monitoring | none | Database health check |
| `list_embedding_models` | Embeddings | none | List embedding models |
| `show_collection_embeddings` | Embeddings | name | Show collection embeddings |

---

## Collection Management Tools

### list_collections

List all collections in the vector database.

**Parameters:** None

**Response:**
```json
{
  "collections": [
    {
      "name": "articles",
      "count": 150,
      "vectorizer": "text-embedding-3-small"
    }
  ]
}
```

**Example Use Cases:**
- Discover available collections
- Monitor collection count
- Verify collection creation

---

### create_collection

Create a new collection with specified schema.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Collection name (alphanumeric + underscores) |
| `type` | string | Yes | Collection type: "text" or "image" |
| `description` | string | No | Collection description |
| `vectorizer` | string | No | Embedding model (default: text2vec-openai) |

**Response:**
```json
{
  "name": "articles",
  "type": "text",
  "vectorizer": "text-embedding-3-small",
  "success": true
}
```

**Supported Vectorizers:**
- `text2vec-openai` (default, uses text-embedding-ada-002)
- `text-embedding-3-small` (recommended - faster, cheaper)
- `text-embedding-3-large` (best quality)
- `text-embedding-ada-002` (legacy)

**Errors:**
- **Collection already exists:** Returns error with existing collection details
- **Invalid vectorizer:** Returns list of supported vectorizers

---

### delete_collection

Delete a collection and all its documents.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Collection name to delete |

**Response:**
```json
{
  "deleted": true,
  "collection": "articles"
}
```

**Warning:** This operation is destructive and cannot be undone.

---

### count_collections

Count the total number of collections in the database.

**Parameters:** None

**Response:**
```json
{
  "count": 5,
  "collections": ["articles", "docs", "images", "notes", "papers"]
}
```

**Example Use Cases:**
- Monitor database size
- Verify multi-collection setups
- Quick inventory check

---

### show_collection

Show detailed information about a collection including schema, count, and properties.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Collection name |

**Response:**
```json
{
  "name": "articles",
  "count": 150,
  "vectorizer": "text-embedding-3-small",
  "schema": {
    "class": "articles",
    "vectorizer": "text-embedding-3-small",
    "properties": [
      {
        "name": "text",
        "dataType": ["text"]
      },
      {
        "name": "url",
        "dataType": ["string"]
      }
    ]
  },
  "properties": [...]
}
```

**Example Use Cases:**
- Inspect collection schema before adding documents
- Debug collection configuration
- Verify vectorizer settings

---

### get_collection_stats

Get statistics for a collection including document count and schema information.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Collection name |

**Response:**
```json
{
  "collection": "articles",
  "document_count": 42,
  "schema": {
    "vectorizer": "text-embedding-3-small",
    "properties": 3
  }
}
```

**Example Use Cases:**
- Monitor collection size and growth
- Quick overview of collection statistics
- Verify collection configuration

---

## Document Management Tools

### list_documents

List documents in a collection with pagination.

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `collection` | string | Yes | - | Collection name |
| `limit` | integer | No | 10 | Max documents to return |
| `offset` | integer | No | 0 | Pagination offset |

**Response:**
```json
{
  "documents": [
    {
      "id": "doc123",
      "url": "https://example.com/article",
      "text": "Document content...",
      "metadata": {
        "title": "Example Article",
        "author": "John Doe"
      }
    }
  ],
  "count": 1
}
```

---

### create_document

Create a new document in a collection.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `collection` | string | Yes | Collection name |
| `url` | string | No | Document URL/identifier |
| `text` | string | No | Document text content |
| `metadata` | object | No | Document metadata (key-value pairs) |

**Response:**
```json
{
  "id": "doc123",
  "created": true
}
```

**Notes:**
- Either `url` or `text` must be provided
- Embeddings are automatically generated
- Metadata is indexed for search

---

### batch_create_documents

Create multiple documents in a single batch operation.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `collection` | string | Yes | Collection name |
| `documents` | array | Yes | Array of document objects |

**Document Object:**
```json
{
  "url": "https://example.com/article1",
  "text": "Content...",
  "metadata": {"key": "value"}
}
```

**Response:**
```json
{
  "created": 10,
  "failed": 0,
  "ids": ["doc1", "doc2", ...]
}
```

**Performance:**
- Much faster than individual creates
- Recommended for bulk imports
- Supports up to 1000 documents per batch

---

### get_document

Retrieve a specific document by ID.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `collection` | string | Yes | Collection name |
| `id` | string | Yes | Document ID |

**Response:**
```json
{
  "id": "doc123",
  "url": "https://example.com/article",
  "text": "Document content...",
  "metadata": {
    "title": "Example"
  }
}
```

---

### update_document

Update an existing document's content or metadata.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `collection` | string | Yes | Collection name |
| `id` | string | Yes | Document ID |
| `text` | string | No | New text content |
| `metadata` | object | No | New metadata |

**Response:**
```json
{
  "id": "doc123",
  "updated": true
}
```

**Notes:**
- Embeddings are regenerated if text is updated
- Metadata is merged with existing values

---

### delete_document

Delete a document from a collection.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `collection` | string | Yes | Collection name |
| `id` | string | Yes | Document ID |

**Response:**
```json
{
  "deleted": true,
  "id": "doc123"
}
```

---

### count_documents

Count documents in a collection.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `collection` | string | Yes | Collection name |

**Response:**
```json
{
  "count": 150,
  "collection": "articles"
}
```

---

### show_document_by_name

Show a document by its filename (from URL or metadata).

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `collection` | string | Yes | Collection name |
| `filename` | string | Yes | Filename to search for |

**Response:**
```json
{
  "document_id": "doc123",
  "collection": "articles",
  "url": "file.txt",
  "text": "Document content...",
  "metadata": {
    "filename": "file.txt",
    "author": "John Doe"
  }
}
```

**Example Use Cases:**
- Find document by filename without knowing ID
- Search documents by metadata filename field
- Lookup documents imported from files

---

### delete_document_by_name

Delete a document by its filename (from URL or metadata).

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `collection` | string | Yes | Collection name |
| `filename` | string | Yes | Filename to search for |

**Response:**
```json
{
  "document_id": "doc123",
  "collection": "articles",
  "filename": "file.txt",
  "status": "deleted"
}
```

**Example Use Cases:**
- Remove documents by filename
- Clean up specific imported files
- Delete documents without knowing their ID

---

### delete_all_documents

Delete all documents from a collection or all collections.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `collection` | string | No | Collection name (if omitted, deletes from all collections) |

**Response (specific collection):**
```json
{
  "collection": "articles",
  "deleted_count": 150
}
```

**Response (all collections):**
```json
{
  "deleted_count": 500,
  "collections_cleaned": 3
}
```

**Warning:** This operation is destructive and cannot be undone. Use with caution.

**Example Use Cases:**
- Reset collection to empty state
- Clean up test data
- Bulk delete before re-import

---

## Query Operations

### query_documents

Perform semantic search on documents using natural language queries.

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `collection` | string | Yes | - | Collection name |
| `query` | string | Yes | - | Search query (natural language) |
| `top_k` | integer | No | 5 | Number of results to return |
| `distance` | number | No | 0.0 | Minimum similarity threshold |

**Response:**
```json
{
  "results": [
    {
      "document": {
        "id": "doc123",
        "text": "Matching content...",
        "metadata": {"title": "Article"}
      },
      "score": 0.95
    }
  ],
  "count": 5
}
```

**Notes:**
- Results are sorted by relevance (score descending)
- Score ranges from 0.0 (no match) to 1.0 (perfect match)
- Uses semantic similarity, not keyword matching

---

### execute_query

Execute a semantic search query across one or all collections.

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `query` | string | Yes | - | Search query (natural language) |
| `collection` | string | No | - | Collection name (if omitted, searches all collections) |
| `limit` | integer | No | 5 | Number of results to return |

**Response (specific collection):**
```json
{
  "collection": "articles",
  "query": "machine learning basics",
  "results": [
    {
      "document_id": "doc123",
      "text": "Introduction to machine learning...",
      "url": "ml-intro.txt",
      "metadata": {"category": "ai"},
      "score": 0.95
    }
  ],
  "count": 1
}
```

**Response (all collections):**
```json
{
  "query": "machine learning basics",
  "results": [
    {
      "collection": "articles",
      "document_id": "doc123",
      "text": "Introduction to machine learning...",
      "url": "ml-intro.txt",
      "metadata": {"category": "ai"},
      "score": 0.95
    },
    {
      "collection": "docs",
      "document_id": "doc456",
      "text": "ML fundamentals...",
      "url": "ml-fund.txt",
      "metadata": {"category": "tutorial"},
      "score": 0.87
    }
  ],
  "count": 2
}
```

**Example Use Cases:**
- Search across multiple collections simultaneously
- Find relevant documents without knowing which collection they're in
- Cross-collection semantic search

---

## AI-Powered Tools

### suggest_schema

Analyze documents and suggest an optimal collection schema using AI.

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `source_path` | string | Yes | - | Path to directory or file(s) to analyze |
| `collection_name` | string | Yes | - | Name for target collection |
| `requirements` | string | No | "" | Optional user requirements for schema |
| `vdb_type` | string | No | "weaviate" | Target vector database type |
| `max_samples` | integer | No | 50 | Max samples to analyze |

**Requirements:**
- `weave` CLI must be installed and in PATH
- OpenAI API key must be configured

**Response:**
```json
{
  "schema": {
    "suggested_properties": [...],
    "vectorizer": "text-embedding-3-small",
    "reasoning": "Analysis explanation..."
  }
}
```

**Example Use Cases:**
- Design schema for new document collections
- Optimize existing schemas
- Understand document structure

---

### suggest_chunking

Analyze documents and suggest optimal chunking configuration using AI.

**Parameters:**

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `source_path` | string | Yes | - | Path to directory or file(s) to analyze |
| `collection_name` | string | Yes | - | Name for target collection |
| `requirements` | string | No | "" | Optional user requirements |
| `vdb_type` | string | No | "weaviate" | Target vector database type |
| `max_samples` | integer | No | 50 | Max samples to analyze |

**Requirements:**
- `weave` CLI must be installed and in PATH
- OpenAI API key must be configured

**Response:**
```json
{
  "chunking_strategy": {
    "chunk_size": 500,
    "chunk_overlap": 50,
    "reasoning": "Explanation..."
  }
}
```

---

## Health & Monitoring

### health_check

Check the health and connectivity of the vector database.

**Parameters:** None

**Response (Healthy):**
```json
{
  "status": "healthy",
  "database": "weaviate-cloud",
  "url": "https://cluster.weaviate.network"
}
```

**Response (Unhealthy):**
```json
{
  "status": "unhealthy",
  "database": "weaviate-cloud",
  "error": "connection refused"
}
```

**Example Use Cases:**
- Verify database connectivity before operations
- Monitor database status
- Debug connection issues
- Health check endpoints

---

## Embedding Management

### list_embedding_models

List all available embedding models and their properties.

**Parameters:** None

**Response:**
```json
{
  "models": [
    {
      "name": "text-embedding-3-small",
      "type": "openai",
      "description": "OpenAI's latest small embedding model - faster and cheaper",
      "dimensions": 1536,
      "provider": "openai"
    },
    {
      "name": "text-embedding-3-large",
      "type": "openai",
      "description": "OpenAI's latest large embedding model - better quality",
      "dimensions": 3072,
      "provider": "openai"
    }
  ],
  "count": 4
}
```

**Example Use Cases:**
- Choose embedding model for new collections
- Compare model capabilities
- Verify model availability

---

### show_collection_embeddings

Show embedding configuration for a specific collection.

**Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | string | Yes | Collection name |

**Response:**
```json
{
  "collection": "articles",
  "vectorizer": "text-embedding-3-small",
  "model": "text-embedding-3-small",
  "dimensions": 1536,
  "provider": "openai"
}
```

**Example Use Cases:**
- Verify collection embedding configuration
- Debug search quality issues
- Compare embeddings across collections

---

## Error Handling

All tools return errors in this format:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {...}
  }
}
```

### Common Errors

| Error | Description | Solution |
|-------|-------------|----------|
| `connection_refused` | Database not accessible | Check if database is running and config is correct |
| `timeout` | Operation timed out | Database may be slow or unreachable |
| `authentication_failed` | Invalid credentials | Check API key configuration |
| `collection_not_found` | Collection doesn't exist | Create collection first |
| `document_not_found` | Document ID not found | Verify document ID |
| `invalid_parameters` | Missing or invalid parameters | Check parameter types and requirements |

---

## Best Practices

### Performance

1. **Use batch operations** for multiple documents (`batch_create_documents`)
2. **Limit query results** with `top_k` to improve performance
3. **Monitor health** before large operations
4. **Use appropriate timeouts** for cloud vs local databases

### Schema Design

1. **Use `suggest_schema`** to design optimal schemas
2. **Choose appropriate vectorizers** based on use case:
   - `text-embedding-3-small`: Fast, cost-effective, good quality
   - `text-embedding-3-large`: Best quality, higher cost
   - `text2vec-openai`: Legacy, uses ada-002

### Error Recovery

1. **Always check `health_check`** before critical operations
2. **Handle timeouts gracefully** with retries
3. **Validate parameters** before sending requests
4. **Use try-catch** for all tool invocations

---

## Examples

See [EXAMPLES.md](EXAMPLES.md) for complete end-to-end usage examples.

---

## References

- [Weave MCP Server GitHub](https://github.com/maximilien/weave-mcp)
- [Model Context Protocol](https://modelcontextprotocol.io)
- [weave-cli Tool](https://github.com/maximilien/weave-cli)
