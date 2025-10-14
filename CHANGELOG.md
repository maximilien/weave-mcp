<!-- markdownlint-disable MD024 -->
# Changelog

All notable changes to the Weave MCP Server project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.1.2] - 2025-10-14

### Changed

- Updated weave-cli dependency from v0.2.8 to v0.2.10

### Features from weave-cli v0.2.10

- Enhanced PDF text extraction with better fallback handling
- Human-friendly error messages with helpful suggestions
- Optional config.yaml with environment variable fallback
- Document update functionality improvements

## [v0.1.1] - 2025-10-07

### Added

- Comprehensive troubleshooting guide in `docs/TROUBLESHOOTING.md`
- Consolidated CI/CD workflows for better maintainability

### Changed

- Separated changelog from README.md into dedicated CHANGELOG.md file

## [v0.1.0] - 2025-10-07

### Added

- Enhanced process management with better inspector process detection and
  cleanup
- Pattern-based process detection for inspector cleanup
- `-env .env` argument to stdio server for proper environment variable loading
- `*.log` and `*.pid` files to gitignore to prevent accidental commits

### Changed

- Updated MCP Inspector config to use relative paths and environment
  variables
- Changed absolute paths to relative paths for better portability
- Improved stop script with more robust process stopping and graceful
  shutdown
- Force kill fallback for stubborn processes

### Fixed

- Fixed Markdown linting issues (line length and emphasis formatting)
- Removed log and pid files from git tracking (now properly ignored)

## [v0.0.7] - Enhanced Process Management and Configuration

### Added

- **Improved Process Management**: Enhanced stop script with better inspector
  process detection and cleanup
- **Configuration Updates**: Updated MCP Inspector config to use relative paths
  and environment variables
- **Better Error Handling**: More robust process stopping with graceful shutdown
  and force kill fallback
- **Path Standardization**: Changed absolute paths to relative paths for better
  portability
- **Environment Integration**: Added `-env .env` argument to stdio server for
  proper environment variable loading
- **Process Detection**: Added pattern-based process detection for inspector
  cleanup
- **Gitignore Updates**: Added `*.log` and `*.pid` files to gitignore to
  prevent accidental commits

## [v0.0.6] - Comprehensive Logging and Monitoring

### Added

- **Logging System**: Added file logging to `./logs/weave-mcp.log`
- **Log Monitoring**: Created `./tools/tail-logs.sh` script for real-time log
  monitoring
- **Dual Output**: Logs written to both console and file simultaneously
- **Service Status**: Added service status checking and PID detection
- **System Integration**: Integrated with system logs for comprehensive
  monitoring
- **Color-coded Output**: Enhanced readability with syntax highlighting
- **Multiple Monitoring Modes**: all, mcp, status, recent, help commands

## [v0.0.5] - Code Reuse and Integration

### Added

- **Weave-cli Integration**: Import Weaviate client directly from weave-cli
- **Code Reuse**: Eliminated code duplication between projects
- **Consistent Behavior**: Same Weaviate client behavior across both projects
- **Easier Maintenance**: Updates to weave-cli automatically benefit MCP
  project
- **Integration Tests**: Added comprehensive integration test suite
- **Test Coverage**: Mock, MCP, and Weave-cli integration tests

## [v0.0.4] - Fast Integration Tests

### Added

- **Fast Integration Tests**: Added comprehensive MCP server integration tests
- **Weaviate Cloud Testing**: Direct testing with Weaviate Cloud client
- **Collection Management**: Test collection creation, listing, and operations
- **Document Operations**: Test document CRUD operations via MCP server
- **Query Testing**: Test semantic search and query functionality
- **Error Handling**: Improved error handling and test reliability

## [v0.0.3] - CI/CD and Release Automation

### Added

- **GitHub Actions**: Complete CI/CD pipeline with build, test, lint, and
  release workflows
- **Multi-platform Builds**: Support for Linux, macOS, and Windows
- **Automated Releases**: Automatic release creation on tag pushes
- **Security Scanning**: Integrated vulnerability scanning and secret detection
- **Code Quality**: Automated linting for Go, YAML, and Markdown files
- **Version Management**: Automated version injection and build information

## [v0.0.2] - Initial Release

### Added

- **MCP Server**: Initial MCP server implementation
- **Vector Database Support**: Weaviate, Milvus, and Mock databases
- **MCP Tools**: Complete set of collection and document management tools
- **Testing Suite**: Comprehensive unit and integration tests with mocks
- **Scripts**: Build, start, stop, lint, and test automation
- **Configuration**: YAML + Environment Variables support
- **Embedding Support**: OpenAI and custom local embeddings

---

## Version History

- **v0.1.2**: Dependency Update - weave-cli v0.2.10
- **v0.1.1**: Documentation and CI/CD Improvements
- **v0.1.0**: Enhanced Process Management and Configuration
- **v0.0.7**: Enhanced Process Management and Configuration (legacy format)
- **v0.0.6**: Comprehensive Logging and Monitoring
- **v0.0.5**: Code Reuse and Integration
- **v0.0.4**: Fast Integration Tests
- **v0.0.3**: CI/CD and Release Automation
- **v0.0.2**: Initial Release

## Contributing

When adding new features or fixing bugs, please update this changelog by:

1. Adding a new entry under `[Unreleased]` for the change
2. Moving `[Unreleased]` to a new version number when releasing
3. Following the format: `### Added`, `### Changed`, `### Deprecated`,
   `### Removed`, `### Fixed`, `### Security`

## Links

- [README.md](README.md) - Main project documentation
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md) - Common issues and
  solutions
- [GitHub Releases](https://github.com/maximilien/weave-mcp/releases) - Release
  downloads
