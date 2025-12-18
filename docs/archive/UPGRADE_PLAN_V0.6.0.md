# Weave-MCP Upgrade Plan: weave-cli v0.6.0

## Executive Summary

Upgrade weave-mcp from weave-cli **v0.3.14** to **v0.6.0** to incorporate major new features including support for three additional vector databases (MongoDB, Milvus, Chroma).

**Current Version**: weave-cli v0.3.14 (weave-mcp v0.2.1)
**Target Version**: weave-cli v0.6.0
**Proposed weave-mcp Version**: v0.3.0 (minor bump due to new features)

---

## Changes in weave-cli v0.3.14 → v0.6.0

### Major Features Added

#### 1. **MongoDB Vector Database Support** (v0.4.0)
- Full MongoDB Atlas Vector Search support
- Automatic embedding generation for MongoDB documents
- MongoDB-specific configuration and commands
- `--mongodb` flag support
- Comprehensive integration tests

**Impact on weave-mcp**:
- ✅ Available via vectordb interface (auto-inherited)
- ⚠️ Needs configuration examples
- ⚠️ May need MongoDB-specific error handling

#### 2. **Milvus Vector Database Support** (v0.5.0)
- Complete Milvus VDB support (local and cloud)
- Milvus integration tests
- Collection operations for Milvus
- Enhanced error messages for collection operations
- Improved test coverage across all VDBs

**Impact on weave-mcp**:
- ✅ Available via vectordb interface (auto-inherited)
- ⚠️ Needs configuration examples
- ⚠️ Should add Milvus to config.yaml

#### 3. **Chroma Vector Database Support** (v0.6.0)
- **Production Ready** Chroma Cloud support
- Chroma VDB support (complete implementation)
- Chroma integration tests
- Windows ARM64 support improvements
- CGO dependency handling for cross-platform builds

**Impact on weave-mcp**:
- ✅ Available via vectordb interface (auto-inherited)
- ⚠️ **CGO dependency may affect build process**
- ⚠️ Windows builds may be affected
- ⚠️ Needs configuration examples

### Supporting Changes

#### Configuration & Testing
- Moved config example files to `configs/` directory
- Enhanced integration test coverage across all VDBs
- Fixed flaky tests in short mode
- Improved help output with grouped flags

#### Bug Fixes
- MongoDB connection fixes
- Supabase integration test fixes
- Weaviate nil pointer dereference prevention
- Collection deletion fixes for Milvus
- Error message improvements

#### Documentation
- Added presentations directory
- Updated VDB support matrix
- Added Neo4j and OpenSearch to roadmap
- Comprehensive planning for v0.4.0-v0.8.0

---

## Impact Assessment for weave-mcp

### Code Changes Required

#### 1. **Low Impact** - Automatic Integration ✅
These work automatically through the vectordb interface:
- MongoDB support
- Milvus support
- Chroma support (if CGO builds work)
- Improved error messages
- Bug fixes

**Action**: None required - inherited via library

#### 2. **Medium Impact** - Configuration Updates ⚠️
- Add MongoDB configuration to `config.yaml.example`
- Add Milvus configuration to `config.yaml.example`
- Add Chroma configuration to `config.yaml.example`
- Update README with new VDB options
- Update CHANGELOG

**Estimated Time**: 2-3 hours

#### 3. **High Impact** - Build System Changes ⚠️⚠️
- **Chroma CGO Dependencies**: May affect cross-compilation
- Windows builds may need special handling
- Binary sizes may increase further
- CI/CD workflows may need updates

**Estimated Time**: 4-6 hours (including testing)

### Binary Size Impact

**Current** (v0.3.14):
- HTTP server: 11M
- stdio server: 36M

**Expected** (v0.6.0):
- HTTP server: ~12-13M (MongoDB, Milvus, Chroma clients)
- stdio server: ~40-45M (Chroma with CGO dependencies)

### Testing Requirements

1. **Unit Tests**: Should pass without changes ✅
2. **Integration Tests**:
   - Mock tests: ✅ Should pass
   - Weaviate tests: ✅ Should pass
   - Supabase tests: ✅ Should pass
   - **New**: MongoDB tests (if credentials available)
   - **New**: Milvus tests (if credentials available)
   - **New**: Chroma tests (if credentials available)

3. **Build Tests**:
   - Linux (amd64, arm64): ✅ Expected to work
   - macOS (amd64, arm64): ⚠️ May need CGO setup
   - Windows (amd64): ⚠️⚠️ **Likely problematic due to Chroma CGO**

---

## Upgrade Strategy

### Phase 1: Dependency Update (1-2 hours)
1. Update `go.mod`: `v0.3.14` → `v0.6.0`
2. Run `go mod tidy`
3. Run `go mod verify`
4. Run `./build.sh` to test basic compilation

**Success Criteria**:
- Dependencies download successfully
- No import errors
- Basic build completes

### Phase 2: Build Verification (2-3 hours)
1. Test Linux builds (native)
2. Test macOS builds (if on macOS)
3. Test Windows builds (cross-compile)
4. Investigate binary size increases
5. Document any CGO-related issues

**Success Criteria**:
- HTTP server builds successfully
- stdio server builds successfully
- All platforms build (or document known limitations)

### Phase 3: Configuration Updates (2-3 hours)
1. Update `config.yaml.example`:
   - Add MongoDB configuration section
   - Add Milvus configuration section
   - Add Chroma configuration section
2. Update `.env.example` with new environment variables
3. Update README.md:
   - Add MongoDB to features list
   - Add Milvus to features list
   - Add Chroma to features list
   - Update vector database support matrix

**Success Criteria**:
- All example configs pass YAML linting
- Documentation is clear and complete

### Phase 4: Testing (3-4 hours)
1. Run `./test.sh unit` - Should pass
2. Run `./test.sh fast` - Should pass
3. Run `./test.sh integration` - Test with available VDBs
4. Test each new VDB if credentials available:
   - MongoDB Atlas
   - Milvus Cloud
   - Chroma Cloud
5. Run `./lint.sh` - Should pass

**Success Criteria**:
- All existing tests pass
- New VDBs work via MCP tools
- No regressions

### Phase 5: CI/CD Updates (2-3 hours)
1. Review build workflow for CGO handling
2. Update cross-compilation matrix if needed
3. Add conditional Windows builds (may skip Chroma)
4. Test CI/CD pipeline

**Success Criteria**:
- CI builds pass on all platforms
- Release artifacts are generated correctly

### Phase 6: Documentation & Release (1-2 hours)
1. Update CHANGELOG.md
2. Update version to v0.3.0
3. Create release notes
4. Tag and release

**Success Criteria**:
- Documentation is complete
- Release is published successfully

---

## Risk Assessment

### High Risk ⚠️⚠️⚠️
**Chroma CGO Dependencies on Windows**
- **Risk**: Windows builds may fail due to Chroma's CGO requirements
- **Mitigation**:
  - Exclude Chroma on Windows (like weave-cli does)
  - Document Windows limitation
  - Consider build tags to make Chroma optional

### Medium Risk ⚠️⚠️
**Binary Size Increase**
- **Risk**: stdio binary may exceed 40MB
- **Impact**: Slower downloads, more disk space
- **Mitigation**: Document size increase, consider optimization

### Low Risk ⚠️
**Configuration Complexity**
- **Risk**: Too many VDB options may confuse users
- **Mitigation**: Clear documentation, good examples

---

## Timeline Estimate

| Phase | Duration | Dependencies |
|-------|----------|--------------|
| Phase 1: Dependency Update | 1-2 hours | None |
| Phase 2: Build Verification | 2-3 hours | Phase 1 |
| Phase 3: Configuration | 2-3 hours | Phase 2 |
| Phase 4: Testing | 3-4 hours | Phase 3 |
| Phase 5: CI/CD | 2-3 hours | Phase 4 |
| Phase 6: Documentation | 1-2 hours | Phase 5 |
| **Total** | **11-17 hours** | Sequential |

---

## Decision Points

### 1. Windows Support for Chroma
**Options**:
- A) Exclude Chroma on Windows (recommended - matches weave-cli)
- B) Attempt CGO builds on Windows (risky, time-consuming)
- C) Make Chroma completely optional with build tags

**Recommendation**: Option A - Exclude Chroma on Windows

### 2. Version Number
**Options**:
- A) v0.3.0 - Minor bump (new features, backward compatible)
- B) v0.2.2 - Patch bump (dependency update only)
- C) v1.0.0 - Major bump (production ready)

**Recommendation**: Option A - v0.3.0 (new VDB support is a feature)

### 3. Testing Strategy
**Options**:
- A) Test only with Mock VDB (fast, limited coverage)
- B) Test with Mock + Weaviate (current approach)
- C) Test with all VDBs (comprehensive, requires credentials)

**Recommendation**: Option B + document new VDB setup

---

## Success Metrics

1. ✅ All existing tests pass
2. ✅ Build succeeds on Linux, macOS
3. ✅ Windows builds (with documented Chroma limitation)
4. ✅ All linting passes
5. ✅ Binary sizes documented
6. ✅ Documentation updated
7. ✅ CI/CD pipeline green
8. ✅ Release published

---

## Rollback Plan

If critical issues are discovered:

1. **Revert go.mod** to v0.3.14
2. **Revert configuration files**
3. **Rebuild and test**
4. **Document issues** for future attempt
5. **Consider partial upgrade** (e.g., skip Chroma)

---

## Next Steps

1. **Review this plan** with stakeholders
2. **Approve/modify** upgrade strategy
3. **Schedule upgrade** work
4. **Execute phases** sequentially
5. **Monitor and iterate** as needed

---

## Questions for Review

1. Are we comfortable with Windows Chroma limitation?
2. Should we test with MongoDB/Milvus/Chroma or just document?
3. Is v0.3.0 the right version number?
4. Do we need to support all VDBs or keep some optional?
5. What's the target timeline for completion?

---

**Document Version**: 1.0
**Created**: 2025-11-15
**Author**: Claude Code
**Status**: Draft - Awaiting Review
