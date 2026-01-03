# Issue #5 Implementation Plan

**Issue:** [#5 - Complete audit of weave-mcp code](https://github.com/maximilien/weave-mcp/issues/5)
**Created:** 2026-01-02
**Status:** Planning Phase

## Audit Summary

### Key Findings

1. **Missing MCP Tools:** 27 tools not yet implemented (32% coverage)
2. **Documentation Gaps:** Critical user-facing documentation missing (25% coverage)
3. **Test Coverage:** MCP package has 0% unit test coverage (36.6% overall)

See detailed audit reports:
- `docs/audit/TOOLS_COMPARISON.md` - MCP tools vs weave-cli capabilities
- `docs/audit/DOCUMENTATION_GAPS.md` - Missing documentation analysis
- `docs/audit/TEST_COVERAGE.md` - Test coverage analysis

---

## Implementation Phases

### Phase 1: High-Priority MCP Tools (Week 1-2)

**Goal:** Implement 5 most critical missing tools

#### Tasks

1. **`show_collection`** - Show collection details
   - Handler: `handleShowCollection`
   - Parameters: `collection_name`
   - Returns: schema, count, properties, vectorizer
   - Estimated: 4 hours

2. **`health_check`** - Database health check
   - Handler: `handleHealthCheck`
   - Parameters: none
   - Returns: status, database type, connection info
   - Estimated: 3 hours

3. **`count_collections`** - Count collections
   - Handler: `handleCountCollections`
   - Parameters: none
   - Returns: count, list of names
   - Estimated: 2 hours

4. **`list_embedding_models`** - List available embedding models
   - Handler: `handleListEmbeddingModels`
   - Parameters: none
   - Returns: array of model names and descriptions
   - Estimated: 3 hours

5. **`show_collection_embeddings`** - Show collection embeddings
   - Handler: `handleShowCollectionEmbeddings`
   - Parameters: `collection_name`
   - Returns: vectorizer, model, dimensions
   - Estimated: 3 hours

**Total Estimated Effort:** 15 hours (2 days)

---

### Phase 2: Critical Documentation (Week 2)

**Goal:** Enable users to discover and use all MCP tools

#### Tasks

1. **Create `docs/MCP_TOOLS.md`**
   - Document all 13 current tools
   - Add parameters, examples, errors for each
   - Include quick reference table
   - Add best practices section
   - Estimated: 12 hours (1.5 days)

2. **Create `docs/EXAMPLES.md`**
   - 5-10 complete end-to-end examples
   - RAG workflow
   - Semantic search
   - Document management
   - Batch operations
   - Error handling
   - Estimated: 8 hours (1 day)

3. **Enhance README.md**
   - Add "Available MCP Tools" section with descriptions
   - Add VDB selection guide
   - Add embedding model comparison table
   - Estimated: 4 hours (0.5 day)

**Total Estimated Effort:** 24 hours (3 days)

---

### Phase 3: MCP Unit Tests (Week 3-4)

**Goal:** Achieve 80%+ test coverage for MCP package

#### Tasks

1. **Create `src/pkg/mcp/server_test.go`**
   - Server initialization tests
   - Tool registration tests
   - Request routing tests
   - Error handling tests
   - Estimated: 8 hours (1 day)

2. **Create `src/pkg/mcp/handlers_test.go`**
   - Unit tests for all 13 handlers
   - Happy path scenarios
   - Error scenarios
   - Mock VDB client
   - Estimated: 16 hours (2 days)

3. **Test AI-Powered Tools**
   - `handleSuggestSchema` tests
   - `handleSuggestChunking` tests
   - `executeCommand` tests with mocks
   - Estimated: 4 hours (0.5 day)

4. **Add Phase 1 Tool Tests**
   - Tests for 5 new handlers
   - Estimated: 6 hours (0.75 day)

**Total Estimated Effort:** 34 hours (4.25 days)

---

### Phase 4: Medium-Priority Tools (Week 5)

**Goal:** Implement enhanced operations

#### Tasks

1. **`get_collection_stats`** - Collection statistics
   - Handler: `handleGetCollectionStats`
   - Parameters: `collection_name`
   - Returns: document count, size, last updated
   - Estimated: 4 hours

2. **`delete_all_documents`** - Bulk delete
   - Handler: `handleDeleteAllDocuments`
   - Parameters: none or `collection_name`
   - Returns: deleted count
   - Estimated: 3 hours

3. **`show_document_by_name`** - Show by filename
   - Handler: `handleShowDocumentByName`
   - Parameters: `collection_name`, `filename`
   - Returns: document details
   - Estimated: 3 hours

4. **`delete_document_by_name`** - Delete by filename
   - Handler: `handleDeleteDocumentByName`
   - Parameters: `collection_name`, `filename`
   - Returns: success/failure
   - Estimated: 3 hours

5. **`execute_query`** - Natural language queries
   - Handler: `handleExecuteQuery`
   - Parameters: `query`, `collection_name` (optional)
   - Returns: query results
   - Estimated: 6 hours

**Total Estimated Effort:** 19 hours (2.5 days)

---

### Phase 5: Developer Documentation (Week 6)

**Goal:** Enable contributors to extend the codebase

#### Tasks

1. **Create `docs/ARCHITECTURE.md`**
   - System architecture diagram
   - MCP server implementation overview
   - Handler flow
   - Database client abstraction
   - Error handling strategy
   - Estimated: 8 hours (1 day)

2. **Create `docs/DEVELOPMENT.md`**
   - Development environment setup
   - Building from source
   - Running tests
   - Adding new MCP tools (step-by-step)
   - Adding new VDB support
   - CI/CD workflow
   - Estimated: 8 hours (1 day)

3. **Create `docs/TESTING.md`**
   - Test structure explanation
   - Running specific test suites
   - Coverage requirements
   - Mock database usage
   - Writing integration tests
   - Estimated: 4 hours (0.5 day)

**Total Estimated Effort:** 20 hours (2.5 days)

---

## Timeline Summary

| Phase | Tasks | Estimated Effort | Week |
|-------|-------|-----------------|------|
| Phase 1 | 5 High-Priority Tools | 2 days | Week 1 |
| Phase 2 | Critical Documentation | 3 days | Week 2 |
| Phase 3 | MCP Unit Tests | 4.25 days | Week 3-4 |
| Phase 4 | 5 Medium-Priority Tools | 2.5 days | Week 5 |
| Phase 5 | Developer Documentation | 2.5 days | Week 6 |
| **TOTAL** | **All Tasks** | **~14 days** | **6 weeks** |

---

## Success Metrics

### Tools

- ✅ Implement 10 new high/medium-priority MCP tools
- ✅ Achieve 60%+ coverage of weave-cli functionality
- ✅ All tools have comprehensive error handling
- ✅ All tools tested (unit + integration)

### Documentation

- ✅ All MCP tools documented with examples
- ✅ 5+ end-to-end usage examples
- ✅ Architecture and development guides complete
- ✅ Documentation coverage: 80%+

### Testing

- ✅ MCP package coverage: 80%+
- ✅ Overall project coverage: 60%+
- ✅ All new tools have unit tests
- ✅ Integration tests for all major workflows

---

## Optional Future Enhancements

### Low-Priority Tools (Phase 6+)

1. **`pipeline_ingest`** - Batch directory ingestion
2. **Configuration tools** - `config_create`, `config_update`, `config_sync`, `config_show`, `config_list`
3. **Schema tools** - `config_list_schemas`, `config_show_schema`, `delete_collection_schema`

**Estimated Effort:** 5-7 days

### Advanced Testing (Phase 7+)

1. **Performance tests** - Load testing, stress testing
2. **E2E tests** - Full MCP protocol compliance
3. **Cross-VDB tests** - Test all 11 databases
4. **Concurrent operation tests** - Thread safety

**Estimated Effort:** 3-4 days

### Documentation Polish (Phase 8+)

1. **Create `docs/FAQ.md`**
2. **Create migration guides** (v0.3→v0.4, v0.4→v0.5)
3. **Auto-generate API reference**
4. **Add documentation CI checks**

**Estimated Effort:** 2-3 days

---

## Dependencies

### Required

- weave-cli v0.8.2+ (already present)
- Go 1.24+ (already present)
- MCP SDK v1.1.0+ (already present)

### Optional

- weave-cli binary in PATH (for AI-powered tools)
- Vector database access (for integration tests)

---

## Risks & Mitigation

### Risk 1: API Changes in weave-cli

**Impact:** Medium
**Probability:** Low
**Mitigation:** Pin weave-cli version, monitor for breaking changes

### Risk 2: Test Coverage Too Time-Consuming

**Impact:** Medium
**Probability:** Medium
**Mitigation:** Prioritize critical paths, use test generation tools

### Risk 3: Documentation Becomes Outdated

**Impact:** High
**Probability:** High
**Mitigation:** Add version tags, automate API reference, CI doc checks

---

## Next Steps

### Immediate (This Session)

1. ✅ Complete audit (DONE)
2. ✅ Create planning documents (DONE)
3. ⏳ User review and approval
4. ⏳ Commit and push audit results

### Short-Term (Week 1)

1. Create feature branch: `feature/issue-5-audit-fixes`
2. Start Phase 1: Implement `show_collection` tool
3. Update MCP server with new tool registration
4. Write handler tests
5. Create PR for review

### Medium-Term (Week 2-4)

1. Complete Phase 1 tools
2. Complete Phase 2 documentation
3. Start Phase 3 unit tests

---

## References

- **Issue #5:** <https://github.com/maximilien/weave-mcp/issues/5>
- **Audit Reports:** `docs/audit/`
- **weave-cli:** `/Users/maximilien/github/maximilien/weave-cli`
- **MCP Server:** `src/pkg/mcp/server.go`
- **MCP Handlers:** `src/pkg/mcp/handlers.go`
