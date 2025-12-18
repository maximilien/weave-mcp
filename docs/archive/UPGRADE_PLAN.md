# Weave MCP Upgrade Plan

## Overview

This document outlines the plan to upgrade weave-mcp from weave-cli v0.2.14 to v0.3.11.

## Current State

- **Current weave-cli version**: v0.2.14
- **Target weave-cli version**: v0.3.11
- **Last weave-mcp release**: v0.1.3 (2025-11-05)

## Major Changes in weave-cli (v0.2.14 → v0.3.11)

### 1. Supabase Vector Database Support (v0.3.9 - v0.3.11)

**Features:**
- New Supabase adapter with OpenAI embedding support
- Vector database abstraction layer
- Collection query support via vectordb abstraction
- Comprehensive integration tests

**Impact on weave-mcp:**
- Add Supabase as a supported vector database option
- Update MCP tools to support Supabase operations
- Add configuration options for Supabase connection
- Update documentation with Supabase setup

### 2. Enhanced Configuration Management (v0.3.5 - v0.3.7)

**Features:**
- Global configuration directory support (`~/.weave-cli`)
- Config file precedence: local → global
- Interactive config create/update commands
- Improved `.env` preservation during updates
- weave-mcp installer integration

**Impact on weave-mcp:**
- Adopt global config directory support
- Update config loading to check both local and global paths
- Add documentation for global vs local configuration
- Potentially simplify installation with config auto-detection

### 3. JSON Output Support (v0.3.8 - v0.3.11)

**Features:**
- `--json` flag for health, document show, and collection query commands
- JSON output for `weave docs ls` and virtual summaries
- Machine-readable output formats

**Impact on weave-mcp:**
- MCP tools already return structured JSON
- May inform future CLI wrapper or debugging tools
- Consider exposing health check endpoint

### 4. Embedding Enhancements (v0.3.8 - v0.3.11)

**Features:**
- `weave embeddings list` command
- `--embedding` flag for document and collection creation
- Enhanced embedding documentation
- Chunk-aware document creation for text files

**Impact on weave-mcp:**
- Review embedding configuration in MCP tools
- Add embedding provider selection to create_collection
- Add embedding options to create_document
- Update documentation with embedding best practices

### 5. REPL and Agent Features (v0.3.0 - v0.3.4)

**Features:**
- REPL mode with batch query support
- AI agents with multi-agent architecture
- Interactive query capabilities
- Version display in REPL banner

**Impact on weave-mcp:**
- Potentially low impact (CLI-specific feature)
- Consider if REPL mode could benefit MCP testing
- May inform future interactive debugging tools

### 6. PDF and Document Processing (v0.2.14 - v0.3.0)

**Features:**
- PDF conversion (CMYK to RGB)
- Text-only PDF processing
- Batch document creation
- Directory-based PDF conversion (`--directory` flag)
- Global `--no-tips` flag

**Impact on weave-mcp:**
- Batch document creation support through MCP
- Consider adding batch_create_documents MCP tool
- Update document creation to support PDF conversion flags

### 7. Database Selection and UX (v0.3.9 - v0.3.11)

**Features:**
- Streamlined database selection
- Smarter defaults for collection/document commands
- Richer collection listing display
- Better offline guidance

**Impact on weave-mcp:**
- Review database selection logic
- Improve error messages when database unavailable
- Add database health checks

### 8. Global Timeout Support (v0.3.8)

**Features:**
- `--timeout` flag with duration format
- Configurable operation timeouts

**Impact on weave-mcp:**
- Add timeout support to MCP tool calls
- Configure per-operation timeouts
- Add timeout parameters to tools

## Proposed Implementation Plan

### Phase 1: Core Dependency Upgrade
**Priority**: High
**Estimated Effort**: Small

1. Update `go.mod` to use weave-cli v0.3.11
2. Run `go mod tidy` to update dependencies
3. Run existing test suite to identify breaking changes
4. Fix any compilation or test failures

### Phase 2: Supabase Support
**Priority**: High
**Estimated Effort**: Medium

1. Add Supabase configuration to `config.yaml` schema
2. Update config loading to support Supabase settings
3. Add Supabase database initialization
4. Update all MCP tools to work with Supabase
5. Add Supabase integration tests
6. Update documentation with Supabase setup instructions

### Phase 3: Global Configuration Support
**Priority**: Medium
**Estimated Effort**: Small

1. Update config path resolution to check `~/.weave-cli`
2. Implement local → global precedence logic
3. Update documentation for global config setup
4. Add config path information to version/debug output

### Phase 4: Enhanced Embedding Support
**Priority**: Medium
**Estimated Effort**: Small-Medium

1. Add embedding provider parameter to `create_collection`
2. Add embedding options to `create_document`
3. Update tool schemas with embedding parameters
4. Add embedding configuration examples
5. Document embedding provider options

### Phase 5: Batch Operations
**Priority**: Low
**Estimated Effort**: Medium

1. Design `batch_create_documents` MCP tool schema
2. Implement batch document creation
3. Add batch operation tests
4. Update documentation with batch examples

### Phase 6: Timeout and Health Checks
**Priority**: Low
**Estimated Effort**: Small

1. Add timeout configuration to tool calls
2. Implement health check endpoint (HTTP mode)
3. Add database connectivity checks
4. Improve error messages for offline databases

### Phase 7: Testing and Documentation
**Priority**: High
**Estimated Effort**: Medium

1. Comprehensive testing with all databases (Mock, Weaviate, Milvus, Supabase)
2. Update README with new features
3. Update CHANGELOG with all changes
4. Create migration guide for existing users
5. Add troubleshooting entries for new features

## Breaking Changes

### Minimal Expected
- Configuration schema additions (backward compatible)
- New optional parameters (backward compatible)
- Dependency version bump (internal change)

### Potential Issues
- Supabase requires new dependencies (may increase binary size)
- Global config precedence may surprise users
- New embedding parameters may need default values

## Testing Strategy

1. **Unit Tests**: Ensure all existing tests pass
2. **Integration Tests**: Test with all supported databases
3. **Compatibility Tests**: Verify backward compatibility with existing configs
4. **New Feature Tests**: Add tests for Supabase and new features
5. **Regression Tests**: Ensure no functionality lost

## Release Strategy

### Version Bump
- **Suggested**: v0.2.0 (minor version bump)
- **Rationale**: Major feature additions (Supabase), minor breaking changes

### Release Checklist
- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG updated
- [ ] Migration guide created
- [ ] CI/CD pipeline successful
- [ ] Binary size validated
- [ ] Manual testing complete

## Timeline Estimate

- **Phase 1**: 1-2 days
- **Phase 2**: 3-4 days
- **Phase 3**: 1 day
- **Phase 4**: 1-2 days
- **Phase 5**: 2-3 days (optional)
- **Phase 6**: 1-2 days (optional)
- **Phase 7**: 2-3 days

**Total**: 11-17 days (core features: 8-10 days)

## Dependencies

### Required
- weave-cli v0.3.11
- Supabase client libraries (via weave-cli)
- Updated MCP SDK (already at v1.1.0)

### Optional
- Additional testing tools
- Performance profiling tools

## Risks and Mitigations

### Risk 1: Binary Size Increase
**Impact**: Medium
**Mitigation**: Monitor binary size, consider build optimization flags

### Risk 2: Supabase Compatibility
**Impact**: Medium
**Mitigation**: Comprehensive integration tests, fallback to existing databases

### Risk 3: Configuration Migration
**Impact**: Low
**Mitigation**: Maintain backward compatibility, provide migration guide

### Risk 4: Performance Regression
**Impact**: Low
**Mitigation**: Benchmark before/after, optimize hot paths

## Success Criteria

1. All existing functionality works with weave-cli v0.3.11
2. Supabase support fully operational
3. All tests passing (unit, integration, e2e)
4. Documentation complete and accurate
5. Binary size increase < 20%
6. No performance regression > 10%
7. Successful release to GitHub

## Questions for Review

1. Should we prioritize Supabase support (Phase 2) or defer to a later release?
2. Is batch operations (Phase 5) important for the initial upgrade?
3. Should we target v0.2.0 or v0.1.4 for version number?
4. Do we want to add any additional MCP tools beyond batch operations?
5. Should we maintain compatibility with older weave-cli versions?

## Next Steps

1. Review and approve this plan
2. Create GitHub issues for each phase
3. Set up project board for tracking
4. Begin Phase 1 implementation
5. Schedule regular progress reviews
