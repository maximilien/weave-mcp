# MCP Tools Usage Examples

Practical end-to-end examples for using Weave MCP Server tools.

**Version:** 0.5.0 (in development)
**Last Updated:** 2026-01-03

---

## Table of Contents

1. [Getting Started](#getting-started)
2. [Basic Workflows](#basic-workflows)
3. [RAG (Retrieval Augmented Generation)](#rag-workflow)
4. [Document Management](#document-management)
5. [Monitoring & Health](#monitoring--health)
6. [AI-Powered Features](#ai-powered-features)
7. [Error Handling](#error-handling)

---

## Getting Started

### Prerequisites

1. **Start the MCP Server:**
   ```bash
   ./start.sh
   ```

2. **Verify Health:**
   ```bash
   # Using MCP Inspector or your MCP client
   health_check()
   ```

---

## Basic Workflows

### Example 1: Create Collection and Add Documents

**Step 1: Check Health**
```json
// Tool: health_check
{}
```

**Response:**
```json
{
  "status": "healthy",
  "database": "weaviate-cloud",
  "url": "https://cluster.weaviate.network"
}
```

**Step 2: Create Collection**
```json
// Tool: create_collection
{
  "name": "articles",
  "type": "text",
  "description": "News articles collection",
  "vectorizer": "text-embedding-3-small"
}
```

**Step 3: Add Documents**
```json
// Tool: batch_create_documents
{
  "collection": "articles",
  "documents": [
    {
      "url": "https://example.com/article1",
      "text": "First article content about AI...",
      "metadata": {
        "title": "AI Breakthroughs",
        "author": "Jane Doe",
        "date": "2026-01-01"
      }
    },
    {
      "url": "https://example.com/article2",
      "text": "Second article about machine learning...",
      "metadata": {
        "title": "ML Trends",
        "author": "John Smith",
        "date": "2026-01-02"
      }
    }
  ]
}
```

**Step 4: Verify**
```json
// Tool: count_documents
{
  "collection": "articles"
}
```

**Response:**
```json
{
  "count": 2,
  "collection": "articles"
}
```

---

### Example 2: Semantic Search

**Search for Documents:**
```json
// Tool: query_documents
{
  "collection": "articles",
  "query": "What are the latest AI developments?",
  "top_k": 3
}
```

**Response:**
```json
{
  "results": [
    {
      "document": {
        "id": "doc1",
        "text": "First article content about AI...",
        "metadata": {
          "title": "AI Breakthroughs",
          "author": "Jane Doe"
        }
      },
      "score": 0.92
    },
    {
      "document": {
        "id": "doc2",
        "text": "Second article about machine learning...",
        "metadata": {
          "title": "ML Trends"
        }
      },
      "score": 0.85
    }
  ],
  "count": 2
}
```

---

## RAG Workflow

Complete workflow for Retrieval Augmented Generation.

### Step 1: Setup Knowledge Base

```json
// Tool: create_collection
{
  "name": "knowledge_base",
  "type": "text",
  "vectorizer": "text-embedding-3-small"
}
```

### Step 2: Import Documents

```json
// Tool: batch_create_documents
{
  "collection": "knowledge_base",
  "documents": [
    {
      "text": "Python is a high-level programming language...",
      "metadata": {"topic": "programming", "language": "python"}
    },
    {
      "text": "Machine learning is a subset of AI...",
      "metadata": {"topic": "ai", "subtopic": "ml"}
    },
    {
      "text": "Vector databases store embeddings...",
      "metadata": {"topic": "databases", "type": "vector"}
    }
  ]
}
```

### Step 3: Query Knowledge Base

```json
// Tool: query_documents
{
  "collection": "knowledge_base",
  "query": "How do vector databases work?",
  "top_k": 3
}
```

### Step 4: Use Results in RAG

```python
# Pseudo-code for RAG integration
results = mcp_client.query_documents(
    collection="knowledge_base",
    query=user_question,
    top_k=3
)

context = "\n".join([r["document"]["text"] for r in results["results"]])

prompt = f"""Answer the question using this context:

Context:
{context}

Question: {user_question}

Answer:"""

response = llm.generate(prompt)
```

---

## Document Management

### Update Document Content

```json
// Tool: update_document
{
  "collection": "articles",
  "id": "doc123",
  "text": "Updated content with new information...",
  "metadata": {
    "updated_at": "2026-01-03",
    "version": 2
  }
}
```

### Get Specific Document

```json
// Tool: get_document
{
  "collection": "articles",
  "id": "doc123"
}
```

### List All Documents (Paginated)

```json
// Tool: list_documents
{
  "collection": "articles",
  "limit": 10,
  "offset": 0
}
```

**Next Page:**
```json
{
  "collection": "articles",
  "limit": 10,
  "offset": 10
}
```

### Delete Document

```json
// Tool: delete_document
{
  "collection": "articles",
  "id": "doc123"
}
```

---

## Monitoring & Health

### Check Database Health

```json
// Tool: health_check
{}
```

### Count All Collections

```json
// Tool: count_collections
{}
```

**Response:**
```json
{
  "count": 3,
  "collections": ["articles", "knowledge_base", "images"]
}
```

### Inspect Collection

```json
// Tool: show_collection
{
  "name": "articles"
}
```

**Response:**
```json
{
  "name": "articles",
  "count": 150,
  "vectorizer": "text-embedding-3-small",
  "schema": {
    "class": "articles",
    "properties": [...]
  }
}
```

### Check Embedding Configuration

```json
// Tool: show_collection_embeddings
{
  "name": "articles"
}
```

**Response:**
```json
{
  "collection": "articles",
  "vectorizer": "text-embedding-3-small",
  "dimensions": 1536,
  "provider": "openai"
}
```

---

## AI-Powered Features

### AI Schema Suggestion

**Analyze documents and get schema recommendations:**

```json
// Tool: suggest_schema
{
  "source_path": "./data/articles",
  "collection_name": "news_articles",
  "requirements": "Schema should support multi-language articles with categories",
  "vdb_type": "weaviate",
  "max_samples": 50
}
```

**Response:**
```json
{
  "schema": {
    "suggested_properties": [
      {
        "name": "title",
        "type": "string",
        "description": "Article title"
      },
      {
        "name": "content",
        "type": "text",
        "description": "Article body"
      },
      {
        "name": "language",
        "type": "string",
        "description": "Article language (en, es, fr, etc.)"
      },
      {
        "name": "category",
        "type": "string",
        "description": "Article category"
      }
    ],
    "vectorizer": "text-embedding-3-small",
    "reasoning": "Analysis shows articles are multi-language with distinct categories..."
  }
}
```

### AI Chunking Suggestion

**Get optimal chunking recommendations:**

```json
// Tool: suggest_chunking
{
  "source_path": "./data/technical_docs",
  "collection_name": "docs",
  "requirements": "Preserve code blocks and maintain context",
  "max_samples": 30
}
```

**Response:**
```json
{
  "chunking_strategy": {
    "chunk_size": 800,
    "chunk_overlap": 100,
    "separator": "\n\n",
    "reasoning": "Technical documents with code blocks require larger chunks to preserve context..."
  }
}
```

---

## Error Handling

### Handle Connection Errors

```python
# Pseudo-code
try:
    result = mcp_client.health_check()
    if result["status"] != "healthy":
        print(f"Database unhealthy: {result.get('error')}")
        # Wait and retry
except ConnectionError as e:
    print(f"Connection failed: {e}")
    # Check database configuration
```

### Handle Collection Not Found

```python
try:
    result = mcp_client.query_documents(
        collection="non_existent",
        query="test"
    )
except MCPError as e:
    if "collection_not_found" in str(e):
        print("Creating collection first...")
        mcp_client.create_collection(
            name="non_existent",
            type="text"
        )
```

### Handle Timeout

```python
try:
    result = mcp_client.batch_create_documents(
        collection="large_collection",
        documents=large_batch
    )
except TimeoutError:
    # Split into smaller batches
    batch_size = 100
    for i in range(0, len(large_batch), batch_size):
        batch = large_batch[i:i+batch_size]
        mcp_client.batch_create_documents(
            collection="large_collection",
            documents=batch
        )
```

---

## Advanced Examples

### Multi-Step Workflow: Build Knowledge Base from Files

```python
# Pseudo-code for complete workflow

# 1. Check health
health = mcp_client.health_check()
assert health["status"] == "healthy"

# 2. Get schema suggestions
schema_suggestion = mcp_client.suggest_schema(
    source_path="./documents",
    collection_name="knowledge_base"
)

# 3. Create collection with suggested schema
mcp_client.create_collection(
    name="knowledge_base",
    type="text",
    vectorizer=schema_suggestion["schema"]["vectorizer"]
)

# 4. Get chunking suggestions
chunking = mcp_client.suggest_chunking(
    source_path="./documents",
    collection_name="knowledge_base"
)

# 5. Process and chunk documents
documents = process_files(
    "./documents",
    chunk_size=chunking["chunk_size"],
    overlap=chunking["chunk_overlap"]
)

# 6. Batch import
mcp_client.batch_create_documents(
    collection="knowledge_base",
    documents=documents
)

# 7. Verify
count = mcp_client.count_documents(collection="knowledge_base")
print(f"Imported {count['count']} documents")
```

### Compare Embedding Models

```python
# Get available models
models = mcp_client.list_embedding_models()

# Create test collections with different models
for model in models["models"]:
    collection_name = f"test_{model['name'].replace('-', '_')}"

    mcp_client.create_collection(
        name=collection_name,
        type="text",
        vectorizer=model["name"]
    )

    # Add same documents to each
    mcp_client.batch_create_documents(
        collection=collection_name,
        documents=test_documents
    )

    # Compare search results
    results = mcp_client.query_documents(
        collection=collection_name,
        query="test query",
        top_k=5
    )

    print(f"Model: {model['name']}")
    print(f"Top score: {results['results'][0]['score']}")
```

---

## Integration Examples

### Claude Desktop Integration

```json
// claude_desktop_config.json
{
  "mcpServers": {
    "weave": {
      "command": "/path/to/weave-mcp/bin/weave-mcp-stdio",
      "args": ["-env", "/path/to/weave-mcp/.env"]
    }
  }
}
```

**Usage in Claude:**
```
Human: Search my knowledge base for information about vector databases