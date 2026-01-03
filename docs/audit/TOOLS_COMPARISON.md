# MCP Tools vs weave-cli Capabilities Comparison

**Audit Date:** 2026-01-02
**Issue:** [#5](https://github.com/maximilien/weave-mcp/issues/5)

## Executive Summary

- **Current MCP Tools:** 13 tools implemented
- **weave-cli Commands:** ~40+ distinct operations
- **Coverage:** ~32% of weave-cli functionality exposed via MCP
- **Critical Gaps:** Configuration management, health checks, statistics, pipeline operations

---

## Current MCP Tools (13)

### ✅ Implemented

| MCP Tool | weave-cli Equivalent | Status |
|----------|---------------------|--------|
| `list_collections` | `weave cols ls` | ✅ Implemented |
| `create_collection` | `weave cols create` | ✅ Implemented |
| `delete_collection` | `weave cols del` | ✅ Implemented |
| `list_documents` | `weave docs ls` | ✅ Implemented |
| `create_document` | `weave docs create` | ✅ Implemented |
| `batch_create_documents` | `weave docs batch create` | ✅ Implemented |
| `get_document` | `weave docs show` | ✅ Implemented |
| `update_document` | `weave docs update` | ✅ Implemented |
| `delete_document` | `weave docs del` | ✅ Implemented |
| `count_documents` | `weave docs count` | ✅ Implemented |
| `query_documents` | `weave cols query` | ✅ Implemented |
| `suggest_schema` | `weave schema suggest` | ✅ Implemented (AI-powered) |
| `suggest_chunking` | `weave chunking suggest` | ✅ Implemented (AI-powered) |

---

## Missing MCP Tools (27)

### 🔴 High Priority (Core Functionality)

| Missing Tool | weave-cli Command | Description | Priority |
|--------------|------------------|-------------|----------|
| `show_collection` | `weave cols show COLLECTION` | Show collection details (schema, count, properties) | **HIGH** |
| `count_collections` | `weave cols count` | Count total collections across database | **HIGH** |
| `health_check` | `weave health check` | Check database connectivity and health | **HIGH** |
| `list_embedding_models` | `weave embeddings list` | List all available embedding models | **HIGH** |
| `show_collection_embeddings` | `weave embeddings list COLLECTION` | Show embeddings for specific collection | **HIGH** |

### 🟡 Medium Priority (Enhanced Operations)

| Missing Tool | weave-cli Command | Description | Priority |
|--------------|------------------|-------------|----------|
| `delete_all_documents` | `weave cols da` | Delete all documents in all collections | **MEDIUM** |
| `delete_collection_schema` | `weave cols ds COLLECTION` | Delete collection schema (destructive) | **MEDIUM** |
| `show_document_by_name` | `weave docs show --name FILE` | Show document by filename instead of ID | **MEDIUM** |
| `delete_document_by_name` | `weave docs del --name FILE` | Delete document by filename | **MEDIUM** |
| `get_collection_stats` | `weave stats` | Show collection statistics and analytics | **MEDIUM** |
| `execute_query` | `weave query` | Execute natural language queries | **MEDIUM** |

### 🟢 Low Priority (Advanced Features)

| Missing Tool | weave-cli Command | Description | Priority |
|--------------|------------------|-------------|----------|
| `pipeline_ingest` | `weave pipeline ingest` | Batch document ingestion from directory | **LOW** |
| `config_create` | `weave config create` | Create new configuration interactively | **LOW** |
| `config_update` | `weave config update` | Update existing configuration | **LOW** |
| `config_sync` | `weave config sync` | Sync local config to ~/.weave-cli | **LOW** |
| `config_show` | `weave config show` | Show current configuration | **LOW** |
| `config_list` | `weave config list` | List all configured databases | **LOW** |
| `config_list_schemas` | `weave config list-schemas` | List all collection schemas | **LOW** |
| `config_show_schema` | `weave config show-schema` | Show specific schema details | **LOW** |
| `config_update_weave_mcp` | `weave config update --weave-mcp` | Install weave-mcp binary for REPL | **LOW** |

---

## Implementation Notes

### AI-Powered Tools

The MCP server implements two AI-powered tools by shelling out to weave-cli:

- **`suggest_schema`**: Calls `weave schema suggest` to analyze text and recommend optimal schema
- **`suggest_chunking`**: Calls `weave chunking suggest` to analyze documents and recommend chunking strategy

**Implementation:** `src/pkg/mcp/handlers.go:547-667`

```go
func executeCommand(name string, args ...string) (string, error) {
    cmd := exec.Command(name, args...)
    output, err := cmd.CombinedOutput()
    // ...
}
```

### Database Selection

weave-cli supports extensive database selection flags (`--weaviate`, `--mongodb`, etc.) that are **NOT exposed** in MCP tools. All MCP operations use the database configured in `config.yaml`.

**Gap:** No way to dynamically select database or operate on multiple databases via MCP.

---

## Recommendations

### Phase 1: Core Tools (High Priority)

1. **`show_collection`** - Essential for understanding collection structure before operations
2. **`health_check`** - Critical for debugging connection issues
3. **`count_collections`** - Basic collection management capability
4. **`list_embedding_models`** - Helps users choose correct vectorizer
5. **`show_collection_embeddings`** - Debug embedding configurations

**Estimated Effort:** 2-3 days

### Phase 2: Enhanced Operations (Medium Priority)

1. **`get_collection_stats`** - Analytics and monitoring
2. **`delete_all_documents`** - Bulk cleanup operations
3. **Document-by-name operations** - More user-friendly than ID-based
4. **`execute_query`** - Natural language search capability

**Estimated Effort:** 3-4 days

### Phase 3: Advanced Features (Low Priority)

1. **Configuration tools** - Advanced configuration management
2. **`pipeline_ingest`** - Batch processing workflows
3. **Schema management tools** - Advanced schema operations

**Estimated Effort:** 5-7 days

---

## Coverage Analysis

### By Category

| Category | weave-cli Commands | MCP Tools | Coverage |
|----------|-------------------|-----------|----------|
| Collections | 7 | 3 | 43% |
| Documents | 10 | 7 | 70% |
| Configuration | 9 | 0 | 0% |
| Health | 1 | 0 | 0% |
| Embeddings | 2 | 0 | 0% |
| AI/Search | 4 | 2 | 50% |
| Pipeline | 1 | 0 | 0% |
| Stats | 1 | 0 | 0% |
| **TOTAL** | **35** | **12** | **34%** |

### Feature Parity

- ✅ **Document CRUD:** Excellent coverage (70%)
- ⚠️ **Collection Management:** Partial coverage (43%)
- ❌ **Configuration:** No coverage (0%)
- ❌ **Monitoring/Health:** No coverage (0%)
- ❌ **Analytics:** No coverage (0%)

---

## References

- **weave-cli Help:** `/Users/maximilien/github/maximilien/weave-cli/bin/weave --help`
- **MCP Server Code:** `src/pkg/mcp/server.go:50-166`
- **MCP Handlers:** `src/pkg/mcp/handlers.go`
- **Issue #5:** <https://github.com/maximilien/weave-mcp/issues/5>
