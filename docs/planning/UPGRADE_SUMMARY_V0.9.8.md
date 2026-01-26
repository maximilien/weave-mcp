# Weave-MCP Upgrade Summary: weave-cli v0.9.8

**Date**: 2026-01-21  
**Current Version**: weave-cli v0.9.4 → **Target Version**: weave-cli v0.9.8  
**Proposed weave-mcp Version**: v0.9.8

---

## Quick Overview

This upgrade incorporates **critical bug fixes** for multi-modal RAG operations, particularly for Milvus vector database and multi-collection queries. The fixes are **automatic** - no code changes needed in weave-mcp handlers.

### Key Fixes

1. ✅ **Milvus Image Collection Ingestion** (v0.9.7)
   - Fixed: 0% → 100% success rate (253/253 images)
   - Image VARCHAR field now stores URL reference when > 2048 chars
   - Full base64 data preserved in ImageData JSON field

2. ✅ **Multi-Collection Query Panic** (v0.9.8)
   - Fixed: Panic with 3+ collections
   - Now supports unlimited collections (1, 2, 3, 10+)
   - Type safety improvements for Milvus metadata columns

3. ✅ **Metadata Truncation Support** (v0.9.5/v0.9.6)
   - Added `--max-metadata-length` flag support
   - Storage reduction: ~87% (3.8MB → 500KB)
   - Embedding cost reduction: ~87% ($18.98 → $2.53)

---

## Impact Assessment

### ✅ No Code Changes Required

- All fixes are internal to weave-cli client
- MCP tools automatically benefit from improvements
- Existing handlers continue to work unchanged
- Backward compatible

### ⚠️ Recommended Updates

1. **Documentation** (30-45 mins)
   - Update README.md with new version info
   - Update CHANGELOG.md with upgrade details
   - Optional: Document `max_metadata_length` parameter in MCP_TOOLS.md

2. **Testing** (30-45 mins)
   - Run full test suite
   - Test multi-collection queries with 3+ collections
   - Test image document creation with Milvus

3. **Configuration** (15 mins)
   - Optional: Add `max_metadata_length` to config.yaml.example

---

## Upgrade Steps

### 1. Update Dependencies (5 mins)

```bash
go get github.com/maximilien/weave-cli@v0.9.8
go mod tidy
```

### 2. Update Test Version (2 mins)

**File**: `tests/e2e/e2e_weave_cli_integration_test.go`
```go
const (
    weaveCLIVersion = "v0.9.8"  // Update from v0.9.4
    weaveCLIRepo    = "maximilien/weave-cli"
)
```

### 3. Update Documentation (30-45 mins)

- README.md: Update "Recent Updates" section
- CHANGELOG.md: Add v0.9.8 entry
- MCP_TOOLS.md: Optional - document `max_metadata_length` parameter

### 4. Build and Test (15 mins)

```bash
./build.sh
./test.sh
```

---

## Benefits

### Immediate Benefits

- ✅ **100% success rate** for Milvus image collections (was 0%)
- ✅ **No collection limit** for multi-collection queries (was panicking at 3+)
- ✅ **87% cost reduction** for image embeddings
- ✅ **87% storage reduction** for image metadata

### Long-term Benefits

- ✅ **Production-ready** multi-modal RAG workflows
- ✅ **Stable** multi-collection query operations
- ✅ **Type-safe** Milvus operations
- ✅ **Better error handling** for edge cases

---

## Risk Level: **LOW** ✅

- No breaking changes
- No API changes
- Automatic benefits
- Backward compatible
- Well-tested in weave-cli

---

## Timeline

**Estimated Time**: 1-2 hours total

- Dependency update: 5 mins
- Documentation: 30-45 mins
- Testing: 30-45 mins
- Build & verify: 15 mins

---

## Next Steps

1. ✅ Review this summary
2. ✅ Review detailed plan: `docs/planning/UPGRADE_PLAN_V0.9.8.md`
3. ✅ Execute upgrade steps
4. ✅ Run test suite
5. ✅ Update documentation
6. ✅ Tag and release v0.9.8

---

## Questions?

- See detailed plan: `docs/planning/UPGRADE_PLAN_V0.9.8.md`
- Check weave-cli changelog: https://github.com/maximilien/weave-cli/blob/main/CHANGELOG.md
- Review weave-cli release: https://github.com/maximilien/weave-cli/releases/tag/v0.9.8

---

**Status**: Ready for implementation  
**Priority**: High (critical bug fixes)  
**Complexity**: Low (dependency update only)
