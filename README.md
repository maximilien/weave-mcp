# Weave MCP Server

A Model Context Protocol (MCP) server for vector database operations, built with
Go and designed to work seamlessly with the
[weave-cli](https://github.com/maximilien/weave-cli) tool.

## Features

- **Vector Database Support**: Weaviate, Milvus, and Mock databases
- **MCP Tools**: Complete set of tools for collection and document management
- **Configuration**: YAML + Environment Variables
- **Testing**: Comprehensive unit and integration tests with mocks
- **Scripts**: Build, start, stop, lint, and test automation
- **Embedding Support**: OpenAI and custom local embeddings

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
│       └── mock/              # Mock client for testing
├── tests/                     # Test files
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

### v0.1.0 (Initial Release)

- Initial MCP server implementation
- Support for Weaviate, Milvus, and Mock databases
- Complete set of collection and document management tools
- Comprehensive testing suite
- Build and deployment scripts
- Configuration management with YAML and environment variables
