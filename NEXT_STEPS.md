# Next Steps for weave-mcp

**Last Updated:** 2026-01-04
**Current Version:** v0.8.2
**Status:** ✅ Shipped and production-ready

---

## ✅ Completed (v0.8.2 - Released Today!)

- [x] Issue #5 audit - Complete analysis
- [x] Issue #6 - Version alignment with weave-cli
- [x] Phase 1 - 5 new MCP tools (health_check, count_collections, show_collection, list_embedding_models, show_collection_embeddings)
- [x] Phase 2 - Comprehensive documentation (MCP_TOOLS.md, EXAMPLES.md)
- [x] Phase 3 - Unit tests (100% coverage for new handlers)
- [x] Release v0.8.2 - Aligned with weave-cli v0.8.2

**Current State:**
- 18 MCP tools (45% weave-cli coverage)
- 60% documentation coverage
- 9.8% test coverage (new handlers 100%)
- Production-ready ✨

---

## 🐛 Open Issues

### Issue #4: Update to latest weave-cli
**Status:** ✅ DONE (Can be closed)
**Completed:** v0.8.2 now aligned with weave-cli v0.8.2
**Action:** Close this issue tomorrow

### Issue #1: Feature - Support HTTPS
**Status:** OPEN
**Priority:** Medium
**Effort:** 2-4 hours
**Description:** Add HTTPS support to HTTP MCP server

**Implementation Plan:**
1. Add TLS configuration to config.yaml
2. Add cert/key file paths to config
3. Update HTTP server to support TLS
4. Add --tls flag to enable HTTPS
5. Document HTTPS setup in README
6. Add example certs for development

**Benefits:**
- Secure MCP server for production use
- Required for some enterprise deployments
- Better security for API key transmission

**Could combine with:** Phase 4 or do standalone

---

## 🎯 Tomorrow's Priority Options

### Option A: Phase 4 - Medium-Priority Tools (⭐ Recommended)

**Goal:** Add 5 useful operation tools → 23 total tools, 60% coverage

**Time:** 2-3 hours

**Tasks:**

1. **`get_collection_stats`** (1 hour)
   - Handler: `handleGetCollectionStats`
   - Returns: document count, size estimate, last updated
   - Most requested tool
   - Tests: 3-4 scenarios

2. **`delete_all_documents`** (45 mins)
   - Handler: `handleDeleteAllDocuments`
   - Parameters: collection_name (optional - all if not provided)
   - Returns: deleted count
   - Useful for testing/cleanup
   - Tests: 2-3 scenarios

3. **`show_document_by_name`** (45 mins)
   - Handler: `handleShowDocumentByName`
   - Parameters: collection_name, filename
   - More user-friendly than ID-based lookup
   - Tests: 3-4 scenarios

4. **`delete_document_by_name`** (30 mins)
   - Handler: `handleDeleteDocumentByName`
   - Parameters: collection_name, filename
   - Returns: success/failure
   - Tests: 2-3 scenarios

5. **`execute_query`** (1 hour)
   - Handler: `handleExecuteQuery`
   - Parameters: query (natural language), collection_name (optional)
   - Natural language search capability
   - Tests: 3-4 scenarios

**Deliverables:**
- 5 new handlers in handlers.go
- 5 tool registrations in server.go
- ~150 lines of tests in handlers_test.go
- Update mock_server.go with mock handlers
- Update MCP_TOOLS.md with new tools
- Update README.md tool counts
- Commit and release v0.8.3

---

### Option B: Issue #1 - Add HTTPS Support (Good Alternative)

**Goal:** Enable secure MCP server for production

**Time:** 2-4 hours

**Tasks:**

1. **Add TLS configuration** (30 mins)
   - Update config.yaml.example
   - Add TLS cert/key paths
   - Add enable_tls flag

2. **Implement TLS in HTTP server** (1 hour)
   - Update src/main.go to support TLS
   - Add cert loading logic
   - Add --tls flag
   - Fallback to HTTP if no certs

3. **Testing** (1 hour)
   - Generate self-signed certs for testing
   - Test HTTPS mode
   - Test HTTP fallback
   - Update tests

4. **Documentation** (30 mins)
   - Document HTTPS setup in README
   - Add cert generation instructions
   - Update MCP Inspector config example

**Deliverables:**
- TLS support in HTTP server
- Configuration examples
- Documentation
- Close issue #1
- Release v0.8.3

---

### Option C: Phase 5 - Developer Documentation

**Goal:** Enable contributors → 80% documentation coverage

**Time:** 2-3 hours

**Tasks:**

1. **`docs/ARCHITECTURE.md`** (1 hour)
   - System architecture diagram
   - MCP server implementation overview
   - Handler flow and lifecycle
   - Database client abstraction
   - Error handling strategy
   - Timeout management

2. **`docs/DEVELOPMENT.md`** (1 hour)
   - Development environment setup
   - Building from source
   - Running tests (unit, integration, coverage)
   - Adding new MCP tools (step-by-step guide)
   - Adding new VDB support
   - CI/CD workflow explanation
   - Release process

3. **`docs/TESTING.md`** (30 mins)
   - Test structure (unit vs integration)
   - Running specific test suites
   - Coverage requirements and goals
   - Mock database usage
   - Writing integration tests
   - CI test execution

**Deliverables:**
- 3 comprehensive developer docs
- Update README.md with links
- Enable community contributions
- Commit and push

---

### Option D: Quick Wins (1 hour)

**Pick 1-2 easy tasks:**

1. **Close issue #4** (5 mins)
   - Already done, just close it
   - Add comment about v0.8.2

2. **Add `get_collection_stats`** (1 hour)
   - Most requested single tool
   - Simple implementation
   - High value

3. **Create `docs/FAQ.md`** (45 mins)
   - Answer common questions
   - VDB selection guide
   - Embedding model comparison
   - Performance tips

4. **Update existing handler tests** (1 hour)
   - Add tests for original 13 tools
   - Increase overall coverage
   - Establish consistent pattern

---

## 💡 Recommended Plan for Tomorrow

**Start with:** Option A (Phase 4) - Best ROI!

**Why:**
- ✅ Quick (2-3 hours)
- ✅ High user value (5 useful tools)
- ✅ Hits 60% coverage milestone
- ✅ Natural continuation of today's work
- ✅ Can do Issue #1 (HTTPS) after if time

**Order:**
1. `get_collection_stats` - Most requested
2. `delete_all_documents` - Testing utility
3. Document-by-name tools - UX improvement
4. `execute_query` - Advanced feature
5. Update docs, test, commit
6. **Then** tackle Issue #1 (HTTPS) if time remains

---

## 📋 Backlog (Future)

### Phase 6: Low-Priority Tools (~5-7 days)

**Configuration Tools (9 tools):**
- config_create, config_update, config_sync
- config_show, config_list
- config_list_schemas, config_show_schema
- config_update_weave_mcp

**Schema Management (3 tools):**
- list_schemas
- show_schema
- delete_collection_schema

**Pipeline (1 tool):**
- pipeline_ingest (batch directory ingestion)

### Advanced Testing (~3-4 days)

- Integration tests for all 18 tools
- E2E MCP protocol compliance tests
- Performance/load tests
- Cross-VDB compatibility tests (11 databases)

### Documentation Polish (~2-3 days)

- FAQ.md (or include in Phase 5)
- Migration guides (v0.4→v0.8, v0.8→v0.9)
- Auto-generated API reference
- Documentation CI checks (link checking, spell check)

### Other Features

- WebSocket support for MCP
- Rate limiting
- Authentication/authorization
- Metrics/monitoring endpoints
- Admin API

---

## 📊 Progress Tracking

| Metric | v0.4.0 | v0.8.2 (Today) | After Phase 4 | After Issue #1 | Goal |
|--------|--------|----------------|---------------|----------------|------|
| **MCP Tools** | 13 | 18 | 23 | 23 | 30+ |
| **weave-cli Coverage** | 32% | 45% | 60% | 60% | 80% |
| **Documentation** | 25% | 60% | 65% | 70% | 80% |
| **Test Coverage** | 36.6% | 9.8% | 15%+ | 15%+ | 60%+ |
| **Features** | Basic | +Health | +Stats | +HTTPS | Complete |

---

## 🚀 Quick Start Tomorrow

```bash
# Pull latest (v0.8.2 tag)
git pull
git pull --tags

# Verify version
git describe --tags
# Should show: v0.8.2

# Check build
./build.sh
cat bin/build-info.txt | head -5

# Verify tests pass
./test.sh

# Ready to code! Start with Phase 4
```

**First task:** Implement `get_collection_stats` handler

---

## 📝 Development Notes

### Patterns to Follow

**Adding a new MCP tool:**

1. **Handler** (handlers.go):
   ```go
   func (s *Server) handleNewTool(ctx context.Context, args map[string]interface{}) (interface{}, error) {
       // 1. Extract and validate parameters
       // 2. Create timeout context
       // 3. Call VDB client method
       // 4. Return formatted response
       // 5. Use s.enhanceError() for errors
   }
   ```

2. **Registration** (server.go):
   ```go
   s.registerTool(Tool{
       Name:        "new_tool",
       Description: "Tool description",
       InputSchema: map[string]interface{}{
           "type": "object",
           "properties": map[string]interface{}{
               "param": map[string]interface{}{
                   "type": "string",
                   "description": "Parameter description",
               },
           },
           "required": []string{"param"},
       },
       Handler: s.handleNewTool,
   })
   ```

3. **Tests** (handlers_test.go):
   ```go
   func TestHandleNewTool(t *testing.T) {
       t.Run("success", func(t *testing.T) { /* ... */ })
       t.Run("error - missing param", func(t *testing.T) { /* ... */ })
       t.Run("error - database failure", func(t *testing.T) { /* ... */ })
   }
   ```

4. **Mock** (mock_server.go):
   - Add tool registration
   - Add mock handler implementation

5. **Documentation** (MCP_TOOLS.md):
   - Add tool to quick reference table
   - Add detailed section with params, response, examples

### Commit Message Format

```
<type>: <short description>

<detailed description>

- Bullet points for changes
- Link to issues if applicable

Related to #<issue_number>
Closes #<issue_number>
```

Types: feat, fix, docs, test, chore, refactor

---

## 🎯 Success Criteria

**After Phase 4:**
- [ ] 23 MCP tools implemented
- [ ] 60% weave-cli coverage achieved
- [ ] All new tools have 100% test coverage
- [ ] MCP_TOOLS.md updated
- [ ] README.md updated
- [ ] All tests passing
- [ ] v0.8.3 tagged and released

**After Issue #1 (HTTPS):**
- [ ] HTTPS support working
- [ ] Configuration documented
- [ ] Tests passing
- [ ] Issue #1 closed
- [ ] Could release as v0.8.4 or v0.9.0

---

## 🎉 Celebrate Today's Wins!

**What we accomplished:**
- ✅ Complete audit of codebase
- ✅ 5 new MCP tools
- ✅ Comprehensive documentation (900+ lines!)
- ✅ 100% test coverage for new features
- ✅ v0.8.2 released (aligned with weave-cli)
- ✅ Closed issues #5 and #6
- ✅ ~3 hours of focused work

**Time well spent!** 🚀

---

## 📞 Questions?

If anything is unclear tomorrow:
1. Check this file for context
2. Review MCP_TOOLS.md for examples
3. Look at existing handlers for patterns
4. Check EXAMPLES.md for usage patterns

**Have fun coding!** 🎨
