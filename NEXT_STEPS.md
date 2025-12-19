# Next Steps for CI/Release Fix

## Current Status

We successfully fixed the Build workflow but are still working on the Release workflow. The issue is **Chroma's CGO dependency** which prevents cross-compilation and even native macOS builds in GitHub Actions.

## What We've Tried

1. ✅ **Exclude darwin from cross-compile** - Worked for Build workflow
2. ✅ **Keep Windows native builds** - Worked
3. ❌ **Delete Chroma source files** - Permission denied (Go module cache read-only)
4. ❌ **Make Go module cache writable with chmod** - Files deleted but imports still fail
5. ❌ **Remove build tag from Chroma stub** - sed command may have failed

## Current Approach (Just Tested - FAILED)

Modified Release workflow to:
- Make Go module cache writable: `chmod -R u+w $(go env GOMODCACHE)/github.com/maximilien/weave-cli@v0.8.2`
- Remove build tag from `stub_unsupported.go`: `sed -i '1d' stub_unsupported.go`
- Set `CGO_ENABLED=0` for all builds

This should allow the stub to compile on all platforms including darwin.

## Alternative Solutions

### Option 1: Remove Chroma Import from Source ⭐ RECOMMENDED

**Simplest and cleanest solution:**

1. Remove Chroma blank imports from source:
   ```bash
   # Edit src/main.go - remove line 25
   # Edit src/cmd/stdio/main.go - remove line 22
   _ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
   ```

2. Update documentation:
   - README: Change "11 databases" to "10 databases"
   - CHANGELOG v0.4.0: Note Chroma exclusion
   - Release notes: Remove Chroma from supported list

3. Benefits:
   - No CI complexity
   - Clean codebase
   - Clear messaging to users
   - Chroma has too many issues (CGO, macOS-only, large binaries)
   - 10 other VDBs are well-supported

4. Users who need Chroma can use weave-cli directly

### Option 2: Use Build Tags in weave-mcp Source

Create conditional imports using build tags:

1. Create `src/init_chroma_darwin.go`:
   ```go
   //go:build darwin && cgo

   package main

   import _ "github.com/maximilien/weave-cli/src/pkg/vectordb/chroma"
   ```

2. Create `src/init_chroma_stub.go`:
   ```go
   //go:build !darwin || !cgo

   package main

   // Chroma not available on this platform
   ```

3. Do NOT import Chroma in main.go/stdio/main.go
4. Use `-tags` flag in release builds if needed

### Option 3: Create Custom Chroma Stub Package

1. Create `src/pkg/vectordb/chroma/` in weave-mcp
2. Implement minimal factory that returns "not supported" errors
3. Import local stub instead of weave-cli Chroma
4. No build tag restrictions

## Recommendation

**Go with Option 1** (Remove Chroma Import):

### Pros:
- Simple, clean solution
- No build complexity
- Clear user expectations
- Chroma support is problematic anyway
- 10 VDBs is still excellent coverage

### Cons:
- Reduces VDB count from 11 to 10
- macOS users lose Chroma option (but can use weave-cli)

## Implementation Steps for Option 1

1. Edit source files:
   ```bash
   # Remove Chroma import from both files
   vim src/main.go        # Remove line 25
   vim src/cmd/stdio/main.go  # Remove line 22
   ```

2. Update documentation:
   ```bash
   # README.md
   - Change "11 databases supported" to "10 databases supported"
   - Remove Chroma from list
   - Add note: "Chroma excluded due to CGO requirements. Use weave-cli for Chroma support."
   
   # CHANGELOG.md v0.4.0
   - Update summary to mention 10 VDBs (not 11)
   - Add note about Chroma exclusion
   
   # .github/workflows/release.yml
   - Remove "Prepare Chroma stub" steps
   - Remove notes about Chroma in release notes
   ```

3. Test locally:
   ```bash
   ./build.sh
   ./bin/weave-mcp --version
   ./bin/weave-mcp-stdio --version
   ```

4. Commit and push:
   ```bash
   git add src/main.go src/cmd/stdio/main.go README.md CHANGELOG.md .github/workflows/release.yml
   git commit -m "feat: exclude Chroma from weave-mcp due to CGO requirements"
   git push origin main
   git tag -d v0.4.0
   git tag -a v0.4.0 -m "Release v0.4.0"
   git push -f origin v0.4.0
   ```

## Test Commands

```bash
# Test local build without Chroma
go build -o bin/weave-mcp ./src/main.go
go build -o bin/weave-mcp-stdio ./src/cmd/stdio/main.go

# Test with mock VDB
VECTOR_DB_TYPE=mock ./bin/weave-mcp &
curl http://localhost:8030/health
```

## Files to Update

1. `src/main.go` - Remove line 25 (Chroma import)
2. `src/cmd/stdio/main.go` - Remove line 22 (Chroma import)
3. `README.md`:
   - Line 7: Change "6 new..." to "5 new..."  
   - Line 17-19: Change "11 databases" to "10 databases", remove Chroma
   - Add note about Chroma exclusion
4. `CHANGELOG.md`:
   - v0.4.0 summary: Update VDB count
   - Add Chroma exclusion note
5. `.github/workflows/release.yml`:
   - Remove "Prepare Chroma stub" steps (lines 59-71, 138-150)
   - Update release notes template

## Timeline

- **Tomorrow Morning**: Implement Option 1
  - Remove imports (5 min)
  - Update docs (10 min)
  - Test locally (5 min)
  - Commit and push (5 min)
  - **Total: 25 minutes**

## Related Issues

- Chroma CGO dependency: `github.com/amikos-tech/chroma-go/pkg/tokenizers/libtokenizers`
- Build tag in stub: `//go:build !(darwin && (amd64 || arm64))`
- Go module cache is read-only in GitHub Actions
- sed command syntax differs between Linux and macOS

## Why This Happened

weave-cli successfully releases because they:
1. Remove Chroma source files during release builds
2. Use `git checkout stub_unsupported.go` to restore only the stub
3. Their stub has build tag that prevents compilation on darwin

weave-mcp tried to follow same pattern but failed because:
1. Our code imports Chroma directly (weave-cli doesn't)
2. Deleting files doesn't affect import resolution  
3. The stub's build tag prevents it from compiling on darwin
4. GitHub Actions runners don't have CGO tools even on macOS

## Success Criteria

Release should:
- ✅ Build successfully on all platforms
- ✅ Support 10 VDBs (all except Chroma)
- ✅ Generate binaries for Linux, macOS, Windows (amd64/arm64)
- ✅ Create GitHub release with checksums
- ✅ Pass all CI checks (build, test, lint)
