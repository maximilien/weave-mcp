# Weave MCP Server

A Model Context Protocol (MCP) server for vector database operations, built with
Go and designed to work seamlessly with the
[weave-cli](https://github.com/maximilien/weave-cli) tool.

> **Recent Updates (v0.4.0)**: Major upgrade to weave-cli v0.8.2 adds support for
> 6 new vector databases (Qdrant, Neo4j, Pinecone, OpenSearch, Elasticsearch,
> enhanced Supabase), bringing total to 11 databases. New features include
> operation-specific timeouts (20-300s based on operation type and deployment),
> VDB-specific error messages for better debugging, enhanced test coverage (36.6%),
> and comprehensive timeout system for reliable operations across all databases.

## Features

- **Dual Transport Support**: HTTP and stdio MCP transports
- **Vector Database Support**: 11 databases supported - Weaviate, Supabase,
  MongoDB, Milvus, Chroma, Qdrant, Neo4j, Pinecone, OpenSearch, Elasticsearch,
  and Mock
- **MCP Tools**: Complete set of tools for collection and document management
- **MCP Inspector**: Web-based debugging and testing interface
- **Configuration**: YAML + Environment Variables
- **Testing**: Comprehensive unit and integration tests with mocks
- **Scripts**: Build, start, stop, lint, and test automation
- **Embedding Support**: OpenAI and custom local embeddings
- **Logging**: Comprehensive file logging with monitoring tools
- **Code Reuse**: Direct integration with weave-cli for consistency

## Transport Modes

The Weave MCP Server supports two transport modes:

### HTTP Transport

- **Binary**: `bin/weave-mcp`
- **URL**: `http://localhost:8030`
- **Use Case**: Web applications, API integrations, testing
- **Features**: RESTful API endpoints, health checks, easy debugging

### stdio Transport

- **Binary**: `bin/weave-mcp-stdio`
- **Transport**: stdin/stdout communication
- **Use Case**: MCP clients like Claude Desktop, direct integration
- **Features**: Native MCP protocol, efficient communication, client integration

## MCP Tools

The server exposes the following MCP tools:

### Collection Management

- `list_collections` - List all collections in the vector database
- `create_collection` - Create a new collection with specified schema
- `delete_collection` - Delete a collection and all its documents

### Document Management

- `list_documents` - List documents in a collection with pagination
- `create_document` - Create a new document in a collection
- `batch_create_documents` - Create multiple documents in a single batch
  operation
- `get_document` - Retrieve a specific document by ID
- `update_document` - Update a document's content or metadata
- `delete_document` - Delete a document from a collection
- `count_documents` - Count documents in a collection

### Query Operations

- `query_documents` - Perform semantic search on documents

## MCP Inspector

The MCP Inspector is a web-based debugging tool that provides a graphical
interface for testing and exploring MCP tools. It's particularly useful for:

- **Testing MCP Tools**: Execute tools with custom parameters
- **Debugging**: See detailed request/response information
- **Exploration**: Discover available tools and their schemas
- **Development**: Rapid prototyping and testing

### Inspector Features

- **Interactive Tool Testing**: Execute any MCP tool with custom parameters
- **Real-time Logging**: See tool execution logs in real-time
- **Schema Exploration**: Browse available tools and their schemas
- **Request/Response Inspection**: Detailed view of MCP protocol messages
- **Web Interface**: Easy-to-use browser-based interface

### Access

Once started, the MCP Inspector is typically available at:

- **Web Interface**: <http://localhost:6274> (with auth token)
- **MCP Server**: <http://localhost:8030>

### Inspector Configuration

The inspector uses the official `npx @modelcontextprotocol/inspector` method
and is configured to connect to the Weave MCP Server using the configuration
in `tools/mcp-inspector-config.json`. This file is created during setup and
includes:

- Server connection details
- Environment variable mapping
- Tool configuration

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

1. Run the setup script (installs MCP Inspector and builds servers):

```bash
./setup.sh
```

1. Configure the server:

```bash
cp config.yaml.example config.yaml
cp .env.example .env
# Edit config.yaml and .env with your settings
```

**Manual Installation** (if setup script fails):

```bash
# Install Go dependencies
go mod tidy

# Build servers
./build.sh
```

### MCP Inspector Setup (Optional)

The MCP Inspector provides a web-based interface for testing and debugging MCP
tools. To set it up:

```bash
# Install MCP Inspector and dependencies
./setup.sh
```

This will:

- Install Node.js dependencies (if not already installed)
- Clone the MCP Inspector repository
- Install inspector dependencies
- Create inspector configuration
- Build the MCP server

**Prerequisites for MCP Inspector:**

- Node.js 22.7.5+ (required for MCP Inspector)
- npm (comes with Node.js)

> **Note**: MCP Inspector requires Node.js 22.7.5 or later. If you have an older
> version, the setup script will skip the inspector installation but continue
> with the server setup.

### Building

```bash
# Build both HTTP and stdio servers
./build.sh

# Build only HTTP server
./build.sh http

# Build only stdio server
./build.sh stdio
```

This will:

- Download Go dependencies
- Run tests
- Build the MCP server binaries (HTTP and/or stdio)
- Create build information

### Running

#### Start HTTP Server

```bash
# Start HTTP server in foreground
./start.sh http

# Start HTTP server as daemon
./start.sh http --daemon
```

#### Start stdio Server

```bash
# Show stdio server configuration for MCP clients
./start.sh stdio
```

#### Start MCP Inspector

```bash
# Start inspector (will start HTTP server if not running)
./start.sh inspector

# Or start both HTTP server and inspector together
./start.sh both
```

#### Stop Services

```bash
# Stop HTTP server
./stop.sh http

# Stop stdio server (if running)
./stop.sh stdio

# Stop inspector
./stop.sh inspector

# Stop all services
./stop.sh all

# Check status
./stop.sh status
```

#### MCP Client Integration

For stdio server integration with MCP clients like Cursor 2.0,
Claude Desktop, and other MCP-compatible clients:

```bash
# Show stdio server configuration
./start.sh stdio
```

**For Cursor 2.0:**

Add this configuration to your Cursor MCP settings (typically in
`~/.cursor/mcp.json` or via Cursor Settings):

```json
{
  "mcpServers": {
    "weave-mcp": {
      "command": "/path/to/weave-mcp/bin/weave-mcp-stdio",
      "args": []
    }
  }
}
```

**For Claude Desktop:**

Add this configuration to your Claude Desktop MCP settings:

```json
{
  "mcpServers": {
    "weave-mcp": {
      "command": "/path/to/weave-mcp/bin/weave-mcp-stdio",
      "args": []
    }
  }
}
```

> **Note**: The server is now compatible with Cursor 2.0's enhanced MCP
> interface and uses MCP SDK v1.1.0 for optimal compatibility.

#### Monitor logs

```bash
# Monitor all logs
./tools/tail-logs.sh

# Monitor HTTP server logs only
./tools/tail-logs.sh mcp

# Monitor stdio server logs only
./tools/tail-logs.sh stdio

# Monitor MCP Inspector logs only
./tools/tail-logs.sh inspector
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

- `GET /health` - Health check (includes database status)
- `GET /mcp/tools/list` - List available MCP tools
- `POST /mcp/tools/call` - Execute an MCP tool

### Example API Usage

#### Health check

```bash
curl http://localhost:8030/health
```

Returns:

```json
{
  "status": "healthy",
  "timestamp": "2025-11-14T19:22:26Z",
  "database": {
    "status": "healthy",
    "type": "weaviate-cloud",
    "name": "weaviate-cloud"
  }
}
```

Returns HTTP 503 if database is unhealthy.

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

#### Update a document

```bash
curl -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d '{
    "name": "update_document",
    "arguments": {
      "collection": "MyCollection",
      "document_id": "document-id-123",
      "content": "Updated document content",
      "metadata": {
        "updated_by": "user",
        "last_modified": "2025-01-09"
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
│   ├── main.go                 # HTTP server entry point
│   ├── cmd/
│   │   └── stdio/
│   │       └── main.go         # stdio server entry point
│   └── pkg/
│       ├── config/            # Configuration management
│       ├── mcp/               # MCP server implementation
│       ├── weaviate/          # Weaviate client (from weave-cli)
│       ├── milvus/            # Milvus client
│       ├── mock/              # Mock client for testing
│       └── version/           # Version information
├── tests/                     # Test files
├── tools/                     # Utility scripts
│   ├── tail-logs.sh          # Log monitoring script
│   └── mcp-inspector-config.json # MCP Inspector configuration
├── logs/                      # Log files directory
│   └── .gitkeep              # Preserve directory in git
├── schemas/                   # Collection schemas
├── bin/                       # Built binaries
│   ├── weave-mcp             # HTTP server binary
│   └── weave-mcp-stdio       # stdio server binary
├── config.yaml               # Configuration file
├── .env                      # Environment variables
├── setup.sh                  # Setup script (MCP Inspector + build)
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

## Documentation

- **[Changelog](CHANGELOG.md)** - Complete version history and release notes
- **[Troubleshooting Guide](docs/TROUBLESHOOTING.md)** - Common issues and
  solutions
- **[GitHub Releases](https://github.com/maximilien/weave-mcp/releases)** -
  Download releases
