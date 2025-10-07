# Weave MCP Server

A Model Context Protocol (MCP) server for vector database operations, built with
Go and designed to work seamlessly with the
[weave-cli](https://github.com/maximilien/weave-cli) tool.

> **Recent Updates**: The server now includes comprehensive logging and monitoring
> capabilities, direct integration with weave-cli for code reuse, and a complete
> CI/CD pipeline with automated testing and releases.

## Features

- **Vector Database Support**: Weaviate, Milvus, and Mock databases
- **MCP Tools**: Complete set of tools for collection and document management
- **Configuration**: YAML + Environment Variables
- **Testing**: Comprehensive unit and integration tests with mocks
- **Scripts**: Build, start, stop, lint, and test automation
- **Embedding Support**: OpenAI and custom local embeddings
- **Logging**: Comprehensive file logging with monitoring tools
- **Code Reuse**: Direct integration with weave-cli for consistency

## MCP Tools

The server exposes the following MCP tools:

### Collection Management

- `list_collections` - List all collections in the vector database
- `create_collection` - Create a new collection with specified schema
- `delete_collection` - Delete a collection and all its documents

### Document Management

- `list_documents` - List documents in a collection with pagination
- `create_document` - Create a new document in a collection
- `get_document` - Retrieve a specific document by ID
- `delete_document` - Delete a document from a collection
- `count_documents` - Count documents in a collection

### Query Operations

- `query_documents` - Perform semantic search on documents

## Quick Start

### Prerequisites

- Go 1.21 or later
- Vector database (Weaviate, Milvus, or use mock for testing)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/maximilien/weave-mcp.git
cd weave-mcp
```

1. Install dependencies:

```bash
go mod tidy
```

1. Configure the server:

```bash
cp config.yaml.example config.yaml
cp .env.example .env
# Edit config.yaml and .env with your settings
```

### Building

```bash
./build.sh
```

This will:

- Download Go dependencies
- Run tests
- Build the MCP server binary
- Create build information

### Running

#### Start the server

```bash
./start.sh
```

#### Start as daemon

```bash
./start.sh --daemon
```

#### Stop the server

```bash
./stop.sh
```

#### Check server status

```bash
./stop.sh status
```

#### Monitor logs

```bash
./tools/tail-logs.sh
```

### Testing

Run all tests:

```bash
./test.sh
```

Run specific test types:

```bash
./test.sh unit        # Unit tests only
./test.sh integration # Integration tests only
./test.sh fast        # Fast tests (unit + mock integration)
./test.sh coverage    # Tests with coverage report
```

### Linting

Check code quality:

```bash
./lint.sh
```

## Configuration

### Environment Variables

Create a `.env` file with the following variables:

```bash
# Vector Database Configuration
VECTOR_DB_TYPE=weaviate-cloud

# Weaviate Cloud Configuration
WEAVIATE_URL=https://your-cluster-url.weaviate.network
WEAVIATE_API_KEY=your-weaviate-api-key
OPENAI_API_KEY=your-openai-api-key

# Collection Names
WEAVIATE_COLLECTION=WeaveDocs
WEAVIATE_COLLECTION_IMAGES=WeaveImages

# MCP Server Configuration
MCP_SERVER_HOST=localhost
MCP_SERVER_PORT=8030
```

### Configuration File

The `config.yaml` file supports multiple database configurations:

```yaml
databases:
  default: weaviate-cloud
  vector_databases:
    - name: weaviate-cloud
      type: weaviate-cloud
      url: ${WEAVIATE_URL}
      api_key: ${WEAVIATE_API_KEY}
      openai_api_key: ${OPENAI_API_KEY}
      collections:
        - name: WeaveDocs
          type: text
          description: Main text documents collection
        - name: WeaveImages
          type: image
          description: Image documents collection
    - name: mock
      type: mock
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections:
        - name: WeaveDocs
          type: text
          description: Mock text documents collection
```

## API Endpoints

The MCP server exposes the following HTTP endpoints:

- `GET /health` - Health check
- `GET /mcp/tools/list` - List available MCP tools
- `POST /mcp/tools/call` - Execute an MCP tool

### Example API Usage

#### List available tools

```bash
curl http://localhost:8030/mcp/tools/list
```

#### Execute a tool

```bash
curl -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "list_collections",
    "arguments": {}
  }'
```

#### Create a collection

```bash
curl -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "create_collection",
    "arguments": {
      "name": "MyCollection",
      "type": "text",
      "description": "My test collection"
    }
  }'
```

#### Create a document

```bash
curl -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "create_document",
    "arguments": {
      "collection": "MyCollection",
      "url": "https://example.com/doc1",
      "text": "This is a test document",
      "metadata": {
        "type": "test",
        "author": "user"
      }
    }
  }'
```

#### Query documents

```bash
curl -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "query_documents",
    "arguments": {
      "collection": "MyCollection",
      "query": "test document",
      "limit": 5
    }
  }'
```

## Logging and Monitoring

The Weave MCP Server includes comprehensive logging and monitoring capabilities.

### Log Files

- **Location**: `./logs/weave-mcp.log`
- **Format**: Structured JSON with timestamps and caller information
- **Output**: Both console and file logging simultaneously
- **Auto-creation**: Logs directory is created automatically on startup

### Log Monitoring Tools

Use the `./tools/tail-logs.sh` script to monitor logs in real-time:

#### Monitor all logs

```bash
./tools/tail-logs.sh all
```

#### Monitor MCP server logs only

```bash
./tools/tail-logs.sh mcp
```

#### Check service status

```bash
./tools/tail-logs.sh status
```

#### Show recent logs

```bash
./tools/tail-logs.sh recent
```

#### Show help

```bash
./tools/tail-logs.sh help
```

### Log Features

- **Real-time monitoring**: Tail logs as they're written
- **Service status**: Check if MCP server is running
- **Color-coded output**: Easy to read with syntax highlighting
- **System integration**: Includes system logs for comprehensive monitoring
- **Multiple modes**: Different monitoring options for different needs

### Example Log Output

```json
{
  "level": "info",
  "ts": 1759848440.371797,
  "caller": "src/main.go:87",
  "msg": "Starting Weave MCP Server",
  "address": "localhost:8030",
  "version": "0.0.6",
  "git_commit": "da2f207"
}
{
  "level": "info",
  "ts": 1759848441.123456,
  "caller": "src/main.go:95",
  "msg": "MCP server started successfully"
}
{
  "level": "info",
  "ts": 1759848442.789012,
  "caller": "src/pkg/mcp/handlers.go:45",
  "msg": "Tool called",
  "tool": "list_collections",
  "collection": "WeaveDocs"
}
```

## Development

### Project Structure

```text
weave-mcp/
├── src/
│   ├── main.go                 # Main server entry point
│   └── pkg/
│       ├── config/            # Configuration management
│       ├── mcp/               # MCP server implementation
│       ├── weaviate/          # Weaviate client (from weave-cli)
│       ├── milvus/            # Milvus client
│       ├── mock/              # Mock client for testing
│       └── version/           # Version information
├── tests/                     # Test files
├── tools/                     # Utility scripts
│   └── tail-logs.sh          # Log monitoring script
├── logs/                      # Log files directory
│   └── .gitkeep              # Preserve directory in git
├── schemas/                   # Collection schemas
├── bin/                       # Built binaries
├── config.yaml               # Configuration file
├── .env                      # Environment variables
├── build.sh                  # Build script
├── start.sh                  # Start script
├── stop.sh                   # Stop script
├── test.sh                   # Test script
└── lint.sh                   # Lint script
```

### Adding New Tools

To add a new MCP tool:

1. Add the tool definition in `src/pkg/mcp/server.go`:

```go
s.registerTool(Tool{
    Name:        "my_new_tool",
    Description: "Description of my new tool",
    InputSchema: map[string]interface{}{
        "type": "object",
        "properties": map[string]interface{}{
            "param1": map[string]interface{}{
                "type":        "string",
                "description": "Parameter description",
            },
        },
        "required": []string{"param1"},
    },
    Handler: s.handleMyNewTool,
})
```

1. Implement the handler in `src/pkg/mcp/handlers.go`:

```go
func (s *Server) handleMyNewTool(ctx context.Context, 
    args map[string]interface{}) (interface{}, error) {
    // Implementation here
    return result, nil
}
```

1. Add tests in `tests/mcp_test.go`

### Adding New Vector Database Support

To add support for a new vector database:

1. Create a new client package in `src/pkg/`
2. Implement the database interface
3. Update the MCP server to support the new database type
4. Add configuration support
5. Add tests

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite: `./test.sh`
6. Run the linter: `./lint.sh`
7. Submit a pull request

## License

This project is licensed under the MIT License - see the
[LICENSE](LICENSE) file for details.

## Related Projects

- [weave-cli](https://github.com/maximilien/weave-cli) - Command-line tool for vector
  database operations
- [Model Context Protocol](https://github.com/modelcontextprotocol) - MCP
  specification and SDKs

## Support

For issues and questions:

1. Check the [Issues](https://github.com/maximilien/weave-mcp/issues) page
2. Create a new issue with detailed information
3. Include logs and configuration details

## Changelog

### v0.0.6 (Latest) - Comprehensive Logging and Monitoring

- **Logging System**: Added file logging to `./logs/weave-mcp.log`
- **Log Monitoring**: Created `./tools/tail-logs.sh` script for real-time log monitoring
- **Dual Output**: Logs written to both console and file simultaneously
- **Service Status**: Added service status checking and PID detection
- **System Integration**: Integrated with system logs for comprehensive monitoring
- **Color-coded Output**: Enhanced readability with syntax highlighting
- **Multiple Monitoring Modes**: all, mcp, status, recent, help commands

### v0.0.5 - Code Reuse and Integration

- **Weave-cli Integration**: Import Weaviate client directly from weave-cli
- **Code Reuse**: Eliminated code duplication between projects
- **Consistent Behavior**: Same Weaviate client behavior across both projects
- **Easier Maintenance**: Updates to weave-cli automatically benefit MCP project
- **Integration Tests**: Added comprehensive integration test suite
- **Test Coverage**: Mock, MCP, and Weave-cli integration tests

### v0.0.4 - Fast Integration Tests

- **Fast Integration Tests**: Added comprehensive MCP server integration tests
- **Weaviate Cloud Testing**: Direct testing with Weaviate Cloud client
- **Collection Management**: Test collection creation, listing, and operations
- **Document Operations**: Test document CRUD operations via MCP server
- **Query Testing**: Test semantic search and query functionality
- **Error Handling**: Improved error handling and test reliability

### v0.0.3 - CI/CD and Release Automation

- **GitHub Actions**: Complete CI/CD pipeline with build, test, lint, and release
  workflows
- **Multi-platform Builds**: Support for Linux, macOS, and Windows
- **Automated Releases**: Automatic release creation on tag pushes
- **Security Scanning**: Integrated vulnerability scanning and secret detection
- **Code Quality**: Automated linting for Go, YAML, and Markdown files
- **Version Management**: Automated version injection and build information

### v0.0.2 - Initial Release

- **MCP Server**: Initial MCP server implementation
- **Vector Database Support**: Weaviate, Milvus, and Mock databases
- **MCP Tools**: Complete set of collection and document management tools
- **Testing Suite**: Comprehensive unit and integration tests with mocks
- **Scripts**: Build, start, stop, lint, and test automation
- **Configuration**: YAML + Environment Variables support
- **Embedding Support**: OpenAI and custom local embeddings
