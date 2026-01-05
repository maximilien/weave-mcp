# Test Organization

This directory contains all tests for weave-mcp, organized by test type.

## Directory Structure

```text
tests/
├── unit/           # Unit tests - fast, isolated, no external dependencies
├── integration/    # Integration tests - test with mock or real databases
├── e2e/           # End-to-end tests - full system tests with real binaries
└── README.md      # This file
```

## Test Categories

### Unit Tests (`unit/`)

Fast, isolated tests that don't require external services:

- `config_test.go` - Configuration loading and validation
- `mcp_test.go` - MCP server initialization and tool registration
- `mock_test.go` - Mock database functionality
- `weaviate_test.go` - Weaviate client unit tests

**Run unit tests:**

```bash
cd tests/unit && go test -v
```

### Integration Tests (`integration/`)

Tests that interact with databases or external services:

- `fast_integration_test.go` - Fast integration tests with Weaviate Cloud
- `weave_cli_integration_test.go` - Integration tests with weave-cli
- `mcp_binary_integration_test.go` - MCP binary HTTP API tests

**Run integration tests:**

```bash
cd tests/integration && go test -v
```

**Skip slow tests:**

```bash
cd tests/integration && go test -v -short
```

### E2E Tests (`e2e/`)

End-to-end tests that test the complete system:

- `e2e_weave_cli_integration_test.go` - Full weave-cli + weave-mcp integration
  - Downloads weave-cli binary from GitHub release
  - Sets up local Weaviate in Docker
  - Tests `weave config update --weave-mcp`
  - Tests AI features (suggest_schema, suggest_chunking)

**Run E2E tests:**

```bash
cd tests/e2e && go test -v
```

**Requirements:**

- Docker installed and running
- OPENAI_API_KEY environment variable set
- Internet connection to download binaries

## Running All Tests

From project root:

```bash
# Run all tests
./test.sh

# Run specific test category
cd tests/unit && go test -v
cd tests/integration && go test -v
cd tests/e2e && go test -v

# Run all tests with coverage
go test -v -cover ./tests/...

# Run only fast tests
go test -v -short ./tests/...
```

## Environment Variables

Tests may require these environment variables:

- `WEAVIATE_URL` - Weaviate instance URL (for integration tests)
- `WEAVIATE_API_KEY` - Weaviate API key (for integration tests)
- `OPENAI_API_KEY` - OpenAI API key (for AI feature tests)

These are loaded from `.env` files in:

- `/Users/maximilien/github/maximilien/weave-mcp/.env`
- `/Users/maximilien/github/maximilien/weave-cli/.env`

## Writing New Tests

### Unit Tests

Place in `unit/` if the test:

- Runs in < 1 second
- Doesn't require external services
- Uses only mock databases
- Tests isolated components

### Integration Tests

Place in `integration/` if the test:

- Requires database connection
- Tests multiple components together
- May use real external services
- Runs in seconds to minutes

### E2E Tests

Place in `e2e/` if the test:

- Tests complete workflows
- Requires building binaries
- Uses Docker containers
- Tests real-world scenarios
- Runs in minutes

## Test Naming Conventions

- Test files: `*_test.go`
- Test functions: `TestXxx(t *testing.T)`
- Benchmark functions: `BenchmarkXxx(b *testing.B)`
- Example functions: `ExampleXxx()`

## CI/CD

GitHub Actions runs:

- Unit tests on every commit
- Integration tests on pull requests
- E2E tests on release branches (if configured)

Use `t.Skip()` to skip tests when required dependencies are unavailable.
