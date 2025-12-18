<!-- markdownlint-disable MD024 -->
# Changelog

All notable changes to the Weave MCP Server project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [v0.4.0] - 2025-12-18

### Added

- **6 New Vector Database Implementations**: Added support for all remaining
  vector databases from weave-cli v0.8.2
  - **Qdrant**: Local and cloud deployments with OpenAI embeddings
  - **Neo4j**: Local and Aura cloud with graph database vector search
  - **Pinecone**: Cloud-only managed vector database service
  - **OpenSearch**: Local and AWS managed deployments (Beta status)
  - **Elasticsearch**: Local and Elastic Cloud deployments (Beta status)
  - **Supabase**: Enhanced configuration (already had basic support)

- **Operation-Specific Timeouts**: Smart timeout handling based on operation type
  - Collection operations: 20s local / 40s cloud (create, delete, list)
  - Document operations: 15s local / 30s cloud (CRUD operations)
  - Query operations: 20s local / 40s cloud (search, list, count)
  - Bulk operations: 120s local / 300s cloud (batch create/delete)
  - Automatically detects cloud vs local deployment for timeout selection

- **Enhanced Error Messages**: VDB-specific error prefixes for better debugging
  - Error format: `{VDB_TYPE}: {operation}: {error_details}`
  - Examples: `weaviate-cloud: failed to list collections: connection refused`
  - Automatic VDB type detection from configuration
  - Helpful context for connection, timeout, and authentication errors

- **Test Coverage Reporting**: Added coverage analysis to test suite
  - New `./test.sh coverage` command generates HTML and text reports
  - Coverage metrics shown in default test runs with `-cover` flag
  - Coverage directory with detailed per-function breakdown
  - Current coverage: 36.6% (config: 53-100%, mock: 75-100%)

### Changed

- **Major Dependency Upgrade**: Updated weave-cli from v0.6.0 to v0.8.2
  - Adds 6 new VDB implementations (Qdrant, Neo4j, Pinecone, OpenSearch,
    Elasticsearch, Supabase enhancements)
  - Comprehensive timeout system with operation-specific durations
  - Enhanced error messages with VDB-specific prefixes
  - 100% batch test coverage across all 10 VDBs
  - Troubleshooting hints in Health() checks
  - Improved connection handling and Close() methods

- **Binary Size**: Increased to accommodate all 11 VDB clients
  - HTTP server: 11M → 108M (882% increase)
  - stdio server: 37M → 108M (192% increase)
  - Reason: All 11 VDB SDK dependencies now statically linked
  - New dependencies: chroma-go, go-elasticsearch/v9, milvus-sdk-go/v2,
    neo4j-go-driver/v5, opensearch-go/v4, go-pinecone, go-client (qdrant)

- **Configuration Files**: Comprehensive updates for 6 new VDBs
  - `config.yaml.example`: Added detailed configs for all 11 VDBs
  - `.env.example`: Added environment variables for 6 new databases
  - Updated VECTOR_DB_TYPE options list (11 total databases)
  - Added deployment-specific guidance (local vs cloud)
  - Beta status markers for OpenSearch and Elasticsearch

- **VDB Factory Registration**: All 11 VDB packages now imported at runtime
  - Updated `src/main.go` (HTTP server) with all VDB imports
  - Updated `src/cmd/stdio/main.go` (stdio server) with all VDB imports
  - Pattern follows weave-cli v0.8.2 factory registration approach
  - Each VDB self-registers via init() functions

- **MCP Handler Improvements**: All 11 handlers updated with modern error handling
  - `handleListCollections`, `handleCreateCollection`, `handleDeleteCollection`
  - `handleListDocuments`, `handleCreateDocument`, `handleBatchCreateDocuments`
  - `handleGetDocument`, `handleDeleteDocument`, `handleCountDocuments`
  - `handleQueryDocuments`, `handleUpdateDocument`
  - Each handler now uses operation-specific timeout contexts
  - All handlers include VDB-specific error prefixes

- **Linting Configuration**: Excluded NEXT_STEPS.md from markdown linting
  - Added NEXT_STEPS.md to `.gitignore` (working project management file)
  - Updated `lint.sh` to skip NEXT_STEPS.md in markdownlint checks
  - Working documents tracked separately from committed documentation

### Fixed

- **Nil Pointer Dereference**: Fixed critical bug in `enhanceError()` method
  - Variable shadowing bug where error parameter was overwritten
  - Renamed config error variable from `err` to `configErr`
  - Prevented nil pointer dereference on `err.Error()` call
  - Bug caught by integration tests (TestFastMCPIntegration)

### Summary

This is a **major feature release** that brings weave-mcp to full feature parity
with weave-cli v0.8.2, supporting all 10 production vector databases plus the
mock testing database.

**Total Vector Databases Supported**: 11
- Weaviate (Cloud + Local)
- Supabase (Cloud + Local PostgreSQL)
- MongoDB (Atlas Cloud)
- Milvus (Local + Cloud)
- Chroma (Local + Cloud, macOS CGO)
- Qdrant (Local + Cloud) - **NEW**
- Neo4j (Local + Aura Cloud) - **NEW**
- Pinecone (Cloud only) - **NEW**
- OpenSearch (Local + Cloud, Beta) - **NEW**
- Elasticsearch (Local + Cloud, Beta) - **NEW**
- Mock (Testing)

**Breaking Changes**: None - all changes are backwards compatible

**Migration Guide**: No migration required from v0.3.0

## [v0.3.0] - 2025-11-27

### Added

- **MongoDB Atlas Vector Search Support**: Full support for MongoDB as a
  vector database backend
  - Automatic embedding generation for MongoDB documents
  - MongoDB-specific configuration in `config.yaml.example`
  - Environment variables: MONGODB_URI, MONGODB_DATABASE

- **Milvus Vector Database Support**: Complete support for Milvus (local
  and cloud)
  - Local and cloud Milvus deployments
  - Milvus-specific configuration examples
  - Environment variables: MILVUS_HOST, MILVUS_PORT, MILVUS_API_KEY

- **Chroma Vector Database Support**: Production-ready Chroma Cloud support
  - Chroma local and cloud deployments
  - Chroma-specific configuration examples
  - Environment variables: CHROMA_URL, CHROMA_API_KEY
  - Note: Windows not supported due to CGO dependencies

### Changed

- **Dependency Upgrade**: Updated weave-cli from v0.3.14 to v0.6.0
  - MongoDB Atlas Vector Search support (v0.4.0)
  - Milvus vector database support (v0.5.0)
  - Chroma vector database support (v0.6.0)
  - Enhanced error messages for collection operations
  - Improved integration test coverage across all VDBs
  - Bug fixes for MongoDB, Supabase, and Weaviate

- **Configuration Files**: Updated config.yaml.example and .env.example
  - Added MongoDB configuration section
  - Added Milvus configuration section
  - Added Chroma configuration section
  - Updated VECTOR_DB_TYPE options list

- **Binary Size**: stdio server increased to ~37M (due to new VDB clients)

### Fixed

- Collection deletion fixes for all VDB types
- Improved error handling for collection operations
- Fixed nil pointer dereferences in integration tests

## [v0.2.1] - 2025-11-15

### Changed

- **Dependency Upgrade**: Updated weave-cli from v0.3.12 to v0.3.14
  - Improved REPL functionality with real-time status updates
  - Enhanced embedding tests for Weaviate and Supabase
  - Documentation improvements and demo updates
  - Note: Binary size increased due to OpenTelemetry dependencies
    (stdio: 9.6M → 36M)

### Fixed

- Removed local replace directive from go.mod that was causing CI failures
- Fixed logger output in stdio mode (stdout → stderr for JSON-RPC
  compatibility)
- Registered all vectordb implementations (mock, supabase, weaviate) for
  runtime availability

## [v0.2.0] - 2025-11-14

### Added

- **Supabase Vector Database Support**: Full support for Supabase as a
  vector database backend
  - Added Supabase configuration to `config.yaml`
  - Supabase-specific fields: `database_url` and `database_key`
  - Works seamlessly with OpenAI embeddings via Supabase adapter

- **Multi-Database Architecture**: Refactored to support multiple vector
  databases through abstraction layer
  - Replaced direct Weaviate client with generic `vectordb.VectorDBClient`
    interface
  - Support for Weaviate (Cloud & Local), Supabase, Mock databases
  - Factory pattern for database client creation

- **Enhanced Embedding Support**: Configurable embedding models and
  vectorizers
  - Added `vectorizer` parameter to `create_collection` MCP tool
  - Support for multiple embedding models (text2vec-openai,
    text-embedding-3-small, text-embedding-3-large, text-embedding-ada-002)
  - Default vectorizer: text2vec-openai
  - Comprehensive embedding documentation in config examples
  - Updated `.env.example` with Supabase configuration and embedding
    model info
  - Updated `config.yaml.example` with detailed embedding model
    descriptions

- **Global Configuration Support**: Support for `~/.weave-cli` global
  configuration directory
  - Configuration precedence: local `.env`/`config.yaml` → global
    `~/.weave-cli`
  - Shared configuration with weave-cli tool

- **Batch Operations**: Efficient bulk document creation
  - Added `batch_create_documents` MCP tool for creating multiple documents
    in a single operation
  - Accepts array of documents with url, text, and metadata
  - Significantly faster than individual document creation for large datasets
  - Full validation and error reporting for batch operations

- **Enhanced Health Checks**: Comprehensive database health monitoring
  - Health endpoint now includes database status (healthy/unhealthy)
  - Returns database type and name information
  - HTTP 503 status when database is unhealthy
  - 5-second timeout for database health checks

- **Improved Error Messages**: User-friendly error reporting
  - Enhanced error messages for common failure scenarios
  - Connection errors: "database connection failed - please check if the
    database is running and accessible"
  - Timeout errors: "operation timed out - database may be slow or
    unreachable"
  - Authentication errors: "authentication failed - please check your API
    key or credentials"
  - Applied to all MCP tool handlers for consistent error reporting

- **CI/CD Improvements**: Refactored GitHub Actions workflows
  - Split monolithic ci.yml into separate workflow files (build, lint,
    test, security)
  - Aligned workflow structure with weave-cli project for consistency
  - Added cross-compilation verification for all platforms
  - Added CodeQL security analysis for main/develop branches
  - Improved matrix builds for Ubuntu, macOS, and Windows
  - Daily security scans with cron schedule

### Changed

- **Dependency Upgrade**: Updated weave-cli from v0.2.14 to v0.3.12
  - Includes all features from weave-cli releases v0.2.14 through v0.3.12
  - New vectordb abstraction layer
  - Supabase adapter
  - Enhanced configuration management

- **MCP Tool Handlers**: Refactored all handlers to use vectordb interfaces
  - `list_collections`, `create_collection`, `delete_collection`
  - `list_documents`, `create_document`, `get_document`, `update_document`, `delete_document`
  - `count_documents`, `query_documents`

- **Binary Size Reduction**: Reduced from ~23MB to ~11MB (HTTP) / ~9.6MB (stdio)
  - Code reuse through weave-cli shared library
  - More efficient builds

- **Configuration Schema**: Extended `VectorDBConfig` with new fields
  - `database_url`: Supabase PostgreSQL connection URL
  - `database_key`: Supabase service role key or anon key
  - `timeout`: Connection timeout in seconds

### Fixed

- Import path updates for weave-cli v0.3.11 compatibility
  - Changed `src/pkg/weaviate` → `src/pkg/vectordb/weaviate`
  - Updated test imports across 3 test files

## [v0.1.3] - 2025-11-05

### Changed

- Upgraded MCP SDK from v1.0.0 to v1.1.0 for better compatibility
- Updated README with Cursor 2.0 configuration instructions
- Fixed default embedding configuration
- Fixed various linting issues

### Added

- Full compatibility with Cursor 2.0's enhanced MCP interface
- Compatibility note for Cursor 2.0 in documentation

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
