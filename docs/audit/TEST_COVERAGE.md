# Test Coverage Analysis

**Audit Date:** 2026-01-02
**Issue:** [#5](https://github.com/maximilien/weave-mcp/issues/5)

## Executive Summary

- **Overall Test Coverage:** 36.6% (as of v0.4.0)
- **MCP Package Coverage:** 0.0% ⚠️
- **Critical Gap:** No unit tests for MCP server and handlers
- **Integration Tests:** Present but limited coverage

---

## Current Test Coverage

### Package-Level Coverage

| Package | Coverage | Status | Files |
|---------|----------|--------|-------|
| `src/pkg/mcp` | **0.0%** | ❌ Critical | server.go, handlers.go |
| `src/pkg/config` | 53-100% | ✅ Good | config.go |
| `src/pkg/mock` | 75-100% | ✅ Good | mock implementations |
| `tests/` | N/A | ⚠️ Integration only | fast_integration_test.go, weave_cli_integration_test.go |
| **OVERALL** | **36.6%** | ⚠️ **Needs Work** | - |

### Test Files Inventory

```
./tests/fast_integration_test.go          ✅ Integration tests (mock VDB)
./tests/weave_cli_integration_test.go     ✅ Integration tests (weave-cli)
```

**Missing:**
- ❌ `src/pkg/mcp/server_test.go` - Unit tests for MCP server
- ❌ `src/pkg/mcp/handlers_test.go` - Unit tests for all 13 handlers
- ❌ `src/pkg/mcp/integration_test.go` - MCP protocol integration tests

---

## Coverage by Functionality

### MCP Tools Coverage

| MCP Tool | Handler Function | Unit Tests | Integration Tests | Coverage |
|----------|-----------------|------------|-------------------|----------|
| `list_collections` | `handleListCollections` | ❌ None | ✅ Yes | 50% |
| `create_collection` | `handleCreateCollection` | ❌ None | ✅ Yes | 50% |
| `delete_collection` | `handleDeleteCollection` | ❌ None | ✅ Yes | 50% |
| `list_documents` | `handleListDocuments` | ❌ None | ✅ Yes | 50% |
| `create_document` | `handleCreateDocument` | ❌ None | ✅ Yes | 50% |
| `batch_create_documents` | `handleBatchCreateDocuments` | ❌ None | ⚠️ Partial | 25% |
| `get_document` | `handleGetDocument` | ❌ None | ✅ Yes | 50% |
| `update_document` | `handleUpdateDocument` | ❌ None | ⚠️ Partial | 25% |
| `delete_document` | `handleDeleteDocument` | ❌ None | ✅ Yes | 50% |
| `count_documents` | `handleCountDocuments` | ❌ None | ✅ Yes | 50% |
| `query_documents` | `handleQueryDocuments` | ❌ None | ✅ Yes | 50% |
| `suggest_schema` | `handleSuggestSchema` | ❌ None | ❌ None | **0%** |
| `suggest_chunking` | `handleSuggestChunking` | ❌ None | ❌ None | **0%** |

**Average Coverage:** ~37% (integration only)
**Target Coverage:** 80%+ (unit + integration)

---

## Missing Test Coverage

### 🔴 Critical (0% Coverage)

#### 1. MCP Server Core (`server.go`)

**Missing Tests:**
- Server initialization and configuration
- Tool registration and discovery
- Request routing to handlers
- Error handling and recovery
- Timeout management
- Context propagation
- VDB client lifecycle

**Estimated Test Count:** 15-20 unit tests
**Estimated Effort:** 2 days

#### 2. AI-Powered Handlers (`handlers.go:547-667`)

**Missing Tests:**
- `handleSuggestSchema` - All scenarios
- `handleSuggestChunking` - All scenarios
- `executeCommand` - Command execution, error handling
- Shell command failures
- Invalid weave-cli installation
- JSON parsing errors

**Estimated Test Count:** 8-10 unit tests
**Estimated Effort:** 1 day

### 🟡 Medium Priority (Partial Coverage)

#### 3. Document Handlers

**Missing Test Scenarios:**
- Invalid parameters (malformed input)
- Database connection failures
- Timeout scenarios
- VDB-specific error handling
- Edge cases (empty collections, missing documents)
- Concurrent operations
- Large batch operations (stress testing)

**Estimated Test Count:** 30-40 unit tests
**Estimated Effort:** 2-3 days

#### 4. Collection Handlers

**Missing Test Scenarios:**
- Invalid collection names
- Duplicate collection creation
- Delete non-existent collection
- Vectorizer validation
- Schema validation errors
- Timeout handling

**Estimated Test Count:** 20-25 unit tests
**Estimated Effort:** 1-2 days

### 🟢 Low Priority (Future)

#### 5. End-to-End Tests

**Missing:**
- Full MCP protocol compliance tests
- Multi-tool workflow tests
- Performance/load tests
- Cross-VDB compatibility tests
- Error recovery tests

**Estimated Test Count:** 10-15 e2e tests
**Estimated Effort:** 2-3 days

---

## Test Quality Issues

### Integration Tests

**`tests/fast_integration_test.go`:**
- ✅ Good: Tests basic CRUD operations
- ✅ Good: Uses mock VDB for fast execution
- ⚠️ Limited: Only covers happy path scenarios
- ❌ Missing: Error handling tests
- ❌ Missing: Timeout tests
- ❌ Missing: AI-powered tools

**`tests/weave_cli_integration_test.go`:**
- ✅ Good: Tests against real Weaviate Cloud
- ✅ Good: Validates weave-cli integration
- ⚠️ Slow: Requires cloud connection
- ❌ Missing: Tests for other 10 VDBs
- ❌ Missing: Comprehensive error scenarios

---

## Test Infrastructure Gaps

### Missing Test Utilities

1. **Mock MCP Client**
   - No helper for simulating MCP requests
   - Manual JSON construction in tests
   - No request/response validation helpers

2. **Test Fixtures**
   - No standardized test documents
   - No test collection templates
   - No error scenario fixtures

3. **VDB Test Helpers**
   - No helpers for setting up test databases
   - No cleanup utilities
   - No test data generation

4. **Coverage Tools**
   - ✅ Coverage reporting added in v0.4.0
   - ⚠️ No coverage enforcement in CI
   - ❌ No coverage differential reporting

---

## Recommendations

### Phase 1: Core MCP Coverage (High Priority)

**Goal:** Achieve 80%+ coverage for MCP package

1. **Create `src/pkg/mcp/server_test.go`**
   - Test server initialization
   - Test tool registration
   - Test request routing
   - Test error handling

2. **Create `src/pkg/mcp/handlers_test.go`**
   - Unit test all 13 handlers
   - Cover happy path + error scenarios
   - Mock VDB client for isolation

3. **Test AI-Powered Tools**
   - Unit test `handleSuggestSchema`
   - Unit test `handleSuggestChunking`
   - Unit test `executeCommand`
   - Mock shell execution

**Estimated Effort:** 4-5 days
**Target Coverage:** 80%+

### Phase 2: Error Handling & Edge Cases (Medium Priority)

**Goal:** Comprehensive error scenario coverage

1. **Add error scenario tests** for all handlers
   - Invalid parameters
   - Database failures
   - Timeouts
   - VDB-specific errors

2. **Add edge case tests**
   - Empty collections
   - Large documents
   - Concurrent operations
   - Resource limits

3. **Enhance integration tests**
   - Add error recovery tests
   - Add timeout tests
   - Add multi-VDB tests

**Estimated Effort:** 3-4 days
**Target Coverage:** 90%+

### Phase 3: Test Infrastructure (Low Priority)

**Goal:** Improve test quality and maintainability

1. **Create test utilities package**
   - Mock MCP client helpers
   - Test fixture generators
   - VDB test helpers
   - Assertion utilities

2. **Add coverage enforcement**
   - CI coverage checks (minimum 80%)
   - Coverage differential reporting
   - Per-package coverage requirements

3. **Add performance tests**
   - Load testing
   - Concurrent operation tests
   - Memory profiling

**Estimated Effort:** 2-3 days

---

## Coverage Targets

### By Package

| Package | Current | Target | Priority |
|---------|---------|--------|----------|
| `src/pkg/mcp` | 0% | 80%+ | **CRITICAL** |
| `src/pkg/config` | 53-100% | 90%+ | **MEDIUM** |
| `src/pkg/mock` | 75-100% | 90%+ | **LOW** |
| `tests/integration` | N/A | 80%+ | **HIGH** |
| **OVERALL** | **36.6%** | **80%+** | **HIGH** |

### By Test Type

| Test Type | Current | Target | Priority |
|-----------|---------|--------|----------|
| Unit Tests | ~10% | 80%+ | **CRITICAL** |
| Integration Tests | ~50% | 80%+ | **HIGH** |
| E2E Tests | 0% | 60%+ | **MEDIUM** |
| Performance Tests | 0% | 40%+ | **LOW** |

---

## Test Execution

### Current Commands

```bash
# Run all tests
./test.sh

# Run with coverage
./test.sh coverage

# View coverage report
open coverage/index.html
```

### Recommended Additions

```bash
# Run only unit tests
./test.sh unit

# Run only integration tests
./test.sh integration

# Run specific package tests
./test.sh pkg/mcp

# Run with coverage enforcement (fail if < 80%)
./test.sh coverage --enforce
```

---

## References

- **Test Files:** `tests/fast_integration_test.go`, `tests/weave_cli_integration_test.go`
- **Coverage Report:** `./test.sh coverage`
- **CHANGELOG:** Current coverage 36.6% (v0.4.0)
- **Issue #5:** <https://github.com/maximilien/weave-mcp/issues/5>
