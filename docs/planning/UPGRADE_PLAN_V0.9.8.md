# Weave-MCP Upgrade Plan: weave-cli v0.9.8

## Executive Summary

Upgrade weave-mcp from weave-cli **v0.9.4** to **v0.9.8** to incorporate critical bug fixes for multi-modal RAG operations, particularly for Milvus vector database and multi-collection queries.

**Current Version**: weave-cli v0.9.4 (weave-mcp v0.9.4)  
**Target Version**: weave-cli v0.9.8  
**Proposed weave-mcp Version**: v0.9.8 (version alignment)

---

## Changes in weave-cli v0.9.4 → v0.9.8

### Critical Bug Fixes (v0.9.5 - v0.9.8)

#### 1. **Image Metadata Truncation Support (v0.9.5, Fixed in v0.9.6)**

**Problem:**
- Image collections failed to ingest in Milvus due to VARCHAR limit (2048 characters)
- 89% of images failed (0/253 in Milvus, 28/253 in Weaviate)
- Metadata fields (`surrounding_text`, `ocr_content`, `section_heading`) exceeded limits

**Solution:**
- Added `--max-metadata-length` flag (default: 2000 chars)
- Truncates at word boundaries to preserve readability
- v0.9.6 fixed the issue where truncation wasn't applied to Milvus metadata map

**Impact on weave-mcp:**
- ✅ **No code changes needed** - MCP tools use underlying weave-cli client
- ⚠️ **Documentation update needed** - Document `max_metadata_length` parameter for `create_document` and `batch_create_documents` tools
- ⚠️ **Configuration update** - Consider adding `max_metadata_length` to config.yaml.example

**Benefits:**
- Storage reduction: ~87% (3.8MB → 500KB for 253 images)
- Embedding cost reduction: ~87% ($18.98 → $2.53 for 253 images)
- Success rate improvement: +80-91 percentage points

#### 2. **Image VARCHAR Field Fix (v0.9.7)**

**Problem:**
- `Image` VARCHAR field stores base64 data URLs (15KB-96KB) exceeding Milvus 2048 char limit
- v0.9.5/v0.9.6 fixed metadata but not the Image field itself
- 0/253 images succeeded in Milvus

**Solution:**
- When `Image` field > 2048 chars: store URL reference instead of full base64 data URL
- Full base64 data remains in `ImageData` JSON field (64KB limit)
- Code changes in `src/pkg/vectordb/milvus/document.go`

**Impact on weave-mcp:**
- ✅ **No code changes needed** - Automatic fix in underlying client
- ✅ **All existing MCP tools benefit automatically**
- ⚠️ **Test update recommended** - Verify image document creation works with Milvus

**Benefits:**
- Success rate: 0% → 100% (253/253 images)
- No data loss: Full base64 stored in ImageData JSON field
- Backward compatible: Weaviate and other VDBs continue to work

#### 3. **Multi-Collection Query Panic Fix (v0.9.8)**

**Problem:**
- Multi-collection queries with 3+ collections crashed with panic:
  ```
  panic: interface conversion: entity.Column is *entity.ColumnVarChar, 
  not *entity.ColumnJSONBytes
  ```
- Root cause: `ListDocuments()` assumed metadata/imageData columns were always `*entity.ColumnJSONBytes`, but after v0.9.7 they can be `*entity.ColumnVarChar`

**Solution:**
- Added type switches to handle both VARCHAR and JSONBytes
- Same pattern as v0.9.3 fix for `parseSearchResults`/`parseQueryResults`
- Code changes in `src/pkg/vectordb/milvus/document.go` lines 430-460

**Impact on weave-mcp:**
- ✅ **No code changes needed** - Fix in underlying client
- ✅ **Critical for `execute_query` tool** - Now supports unlimited collections
- ⚠️ **Test update recommended** - Test multi-collection queries with 3+ collections

**Benefits:**
- No collection number limit (supports 1, 2, 3, 10+ collections)
- Multi-modal RAG queries work across any number of collections
- Type safety consistent across all Milvus operations

---

## Upgrade Steps

### Step 1: Update Dependencies

**File**: `go.mod`

```go
require (
    github.com/maximilien/weave-cli v0.9.8  // Update from v0.9.4
    // ... other dependencies
)
```

**Commands:**
```bash
go get github.com/maximilien/weave-cli@v0.9.8
go mod tidy
```

### Step 2: Update Test Version Reference

**File**: `tests/e2e/e2e_weave_cli_integration_test.go`

```go
const (
    weaveCLIVersion = "v0.9.8"  // Update from v0.9.4
    weaveCLIRepo    = "maximilien/weave-cli"
)
```

### Step 3: Update Documentation

#### 3.1 Update README.md

**File**: `README.md`

- Update version reference in "Recent Updates" section:
  ```markdown
  > **Recent Updates (v0.9.8)**: Critical bug fixes for multi-modal RAG operations
  > - Fixed Milvus image collection ingestion (0% → 100% success rate)
  > - Fixed multi-collection query panic with 3+ collections
  > - Added metadata truncation support for large image metadata
  > - All 23 existing MCP tools automatically benefit from improved stability
  ```

#### 3.2 Update MCP_TOOLS.md

**File**: `docs/MCP_TOOLS.md`

Add documentation for `max_metadata_length` parameter:

- **`create_document` tool**: Add optional `max_metadata_length` parameter
  - Description: Maximum length for metadata fields (default: 2000)
  - Recommended values: Milvus: 2000, Weaviate: 8000, Other: 2000
  - Applies to: `surrounding_text`, `ocr_content`, `section_heading` fields

- **`batch_create_documents` tool**: Same parameter documentation

#### 3.3 Update CHANGELOG.md

**File**: `CHANGELOG.md`

Add new entry:

```markdown
## [v0.9.8] - 2026-01-21

### Changed

- **Updated weave-cli dependency to v0.9.8** - Critical bug fixes for multi-modal RAG
  - Fixed Milvus image collection ingestion (0% → 100% success rate)
    - Image VARCHAR field now stores URL reference when > 2048 chars
    - Full base64 data preserved in ImageData JSON field
  - Fixed multi-collection query panic with 3+ collections
    - Added type switches for VARCHAR and JSONBytes columns
    - No collection number limit (supports 1, 2, 3, 10+ collections)
  - Added metadata truncation support via `max_metadata_length` parameter
    - Storage reduction: ~87% (3.8MB → 500KB for 253 images)
    - Embedding cost reduction: ~87% ($18.98 → $2.53 for 253 images)
  - All 23 existing MCP tools automatically benefit from improved stability
```

### Step 4: Update Configuration Example

**File**: `config.yaml.example`

Consider adding optional `max_metadata_length` setting:

```yaml
# Optional: Maximum length for metadata fields in image collections
# Recommended: 2000 for Milvus (VARCHAR limit 2048), 8000 for Weaviate
# Default: 2000 if not specified
max_metadata_length: 2000
```

### Step 5: Testing

#### 5.1 Unit Tests

**Files**: `tests/unit/`, `src/pkg/mcp/handlers_test.go`

- ✅ No changes needed - existing tests should pass
- ⚠️ Consider adding test for `max_metadata_length` parameter if we expose it

#### 5.2 Integration Tests

**Files**: `tests/integration/`

- ✅ Run existing integration tests
- ⚠️ **Recommended**: Add test for multi-collection queries with 3+ collections
- ⚠️ **Recommended**: Test image document creation with Milvus

#### 5.3 E2E Tests

**File**: `tests/e2e/e2e_weave_cli_integration_test.go`

- Update `weaveCLIVersion` constant to `v0.9.8`
- ✅ Existing E2E tests should pass
- ⚠️ **Recommended**: Add test for multi-collection query with 3+ collections

### Step 6: Build and Verify

```bash
# Clean build
./build.sh

# Run all tests
./test.sh

# Verify version
go list -m github.com/maximilien/weave-cli
# Should show: github.com/maximilien/weave-cli v0.9.8
```

---

## Risk Assessment

### Low Risk ✅

- **Dependency Update**: Simple version bump, no API changes
- **Backward Compatibility**: All fixes are internal to weave-cli client
- **No Breaking Changes**: Existing MCP tools continue to work
- **Automatic Benefits**: All 23 MCP tools benefit from fixes automatically

### Medium Risk ⚠️

- **Test Coverage**: Need to verify multi-collection queries work correctly
- **Documentation**: Need to document new `max_metadata_length` parameter
- **Configuration**: Optional config update for better UX

### Mitigation

1. **Comprehensive Testing**: Run full test suite before release
2. **Gradual Rollout**: Test with real workloads before production
3. **Documentation**: Clear migration notes (though no migration needed)

---

## Timeline

**Estimated Time**: 1-2 hours

- **Step 1-2** (Dependency Update): 15 minutes
- **Step 3** (Documentation): 30-45 minutes
- **Step 4** (Configuration): 15 minutes
- **Step 5** (Testing): 30-45 minutes
- **Step 6** (Build & Verify): 15 minutes

---

## Success Criteria

- [ ] `go.mod` updated to weave-cli v0.9.8
- [ ] All tests passing (unit, integration, E2E)
- [ ] Documentation updated (README, CHANGELOG, MCP_TOOLS.md)
- [ ] Configuration example updated (optional)
- [ ] Build successful
- [ ] Version alignment: weave-mcp v0.9.8 matches weave-cli v0.9.8

---

## Post-Upgrade Verification

### Manual Testing Checklist

- [ ] `list_collections` works with Milvus
- [ ] `create_document` with image collection works with Milvus
- [ ] `execute_query` works with 3+ collections
- [ ] `batch_create_documents` with images works
- [ ] Multi-modal RAG queries return results from all collections

### Automated Testing

```bash
# Run full test suite
./test.sh

# Run integration tests
./test.sh integration

# Run E2E tests (requires weave-cli binary)
./test.sh e2e
```

---

## Notes

### Why This Upgrade is Important

1. **Critical Bug Fixes**: Fixes blocking issues for production deployments
2. **Multi-Modal RAG**: Essential for image collection workflows
3. **Milvus Support**: Makes Milvus fully functional for image collections
4. **Stability**: Fixes panic in multi-collection queries

### What Doesn't Need to Change

- ✅ MCP tool implementations (no code changes needed)
- ✅ Tool schemas (no API changes)
- ✅ Handler logic (benefits automatically)
- ✅ Mock implementations (no changes needed)

### Future Considerations

- Consider exposing `max_metadata_length` as an MCP tool parameter
- Consider adding configuration option for default metadata length
- Monitor for any edge cases in production deployments

---

## References

- [weave-cli v0.9.8 Release](https://github.com/maximilien/weave-cli/releases/tag/v0.9.8)
- [weave-cli CHANGELOG](https://github.com/maximilien/weave-cli/blob/main/CHANGELOG.md)
- [Previous Upgrade Plan (v0.8.2)](./UPGRADE_PLAN_V0.6.0.md)

---

**Last Updated**: 2026-01-21  
**Status**: Ready for Implementation
