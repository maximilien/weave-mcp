#!/usr/bin/env bash

# Load environment variables from .env file
set -a
if [ -f .env ]; then
    # shellcheck disable=SC1091
    . .env
fi
set +a

# Weave MCP Server Test Suite

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_help() {
    echo -e "${BLUE}Weave MCP Server Test Suite${NC}"
    echo ""
    echo "Usage: ./test.sh [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  unit        Run only unit tests (tests/unit/)"
    echo "  integration Run only integration tests (tests/integration/)"
    echo "  e2e         Run only end-to-end tests (tests/e2e/)"
    echo "  fast        Run fast tests (unit + integration, skip E2E)"
    echo "  all         Run all tests (unit + integration + e2e)"
    echo "  coverage    Run tests with coverage report"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./test.sh unit         # Run only unit tests"
    echo "  ./test.sh integration  # Run only integration tests"
    echo "  ./test.sh e2e          # Run only E2E tests"
    echo "  ./test.sh fast         # Run unit + integration (skip E2E)"
    echo "  ./test.sh all          # Run all tests"
    echo "  ./test.sh coverage     # Run tests with coverage report"
    echo "  ./test.sh              # Run unit tests (default)"
    echo ""
    echo "Test Categories:"
    echo "  Unit Tests (tests/unit/):"
    echo "    - Configuration management"
    echo "    - MCP server initialization"
    echo "    - Mock database functionality"
    echo "    - Weaviate client unit tests"
    echo ""
    echo "  Integration Tests (tests/integration/):"
    echo "    - MCP + Weaviate Cloud integration"
    echo "    - MCP binary HTTP API tests"
    echo "    - weave-cli integration tests"
    echo ""
    echo "  E2E Tests (tests/e2e/):"
    echo "    - Full weave-cli + weave-mcp integration"
    echo "    - AI features (suggest_schema, suggest_chunking)"
    echo "    - Requires Docker and OPENAI_API_KEY"
}

# Initialize variables
RUN_UNIT_TESTS=false
RUN_INTEGRATION_TESTS=false
RUN_E2E_TESTS=false
RUN_COVERAGE=false

# Check command line arguments
case "${1:-unit}" in
    "unit")
        RUN_UNIT_TESTS=true
        RUN_INTEGRATION_TESTS=false
        RUN_E2E_TESTS=false
        RUN_COVERAGE=false
        ;;
    "integration")
        RUN_UNIT_TESTS=false
        RUN_INTEGRATION_TESTS=true
        RUN_E2E_TESTS=false
        RUN_COVERAGE=false
        ;;
    "e2e")
        RUN_UNIT_TESTS=false
        RUN_INTEGRATION_TESTS=false
        RUN_E2E_TESTS=true
        RUN_COVERAGE=false
        ;;
    "fast")
        RUN_UNIT_TESTS=true
        RUN_INTEGRATION_TESTS=true
        RUN_E2E_TESTS=false
        RUN_COVERAGE=false
        ;;
    "all")
        RUN_UNIT_TESTS=true
        RUN_INTEGRATION_TESTS=true
        RUN_E2E_TESTS=true
        RUN_COVERAGE=false
        ;;
    "coverage")
        RUN_UNIT_TESTS=false
        RUN_INTEGRATION_TESTS=false
        RUN_E2E_TESTS=false
        RUN_COVERAGE=true
        ;;
    "help"|"-h"|"--help")
        print_help
        exit 0
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        print_help
        exit 1
        ;;
esac

# Function to run unit tests
run_unit_tests() {
    print_header "Running Unit Tests (tests/unit/)..."

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Run unit tests with coverage
    print_status "Running unit tests..."
    if go test -v -cover -timeout=30s ./tests/unit/...; then
        print_success "Unit tests passed!"
    else
        print_error "Unit tests failed!"
        exit 1
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_header "Running Integration Tests (tests/integration/)..."

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Run integration tests with short flag to skip slow tests
    print_status "Running integration tests..."
    if go test -v -short -timeout=120s ./tests/integration/...; then
        print_success "Integration tests passed!"
    else
        print_warning "Integration tests failed or skipped"
        print_status "Some tests may require WEAVIATE_URL and WEAVIATE_API_KEY"
    fi
}

# Function to run E2E tests
run_e2e_tests() {
    print_header "Running E2E Tests (tests/e2e/)..."

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Check for Docker
    if ! command -v docker >/dev/null 2>&1; then
        print_warning "Docker not found - E2E tests may be skipped"
    fi

    # Check for OPENAI_API_KEY
    if [ -z "$OPENAI_API_KEY" ]; then
        print_warning "OPENAI_API_KEY not set - AI feature tests may be skipped"
    fi

    # Run E2E tests (they will skip if dependencies are missing)
    print_status "Running E2E tests..."
    if go test -v -timeout=10m ./tests/e2e/...; then
        print_success "E2E tests passed!"
    else
        print_warning "E2E tests failed or skipped"
        print_status "E2E tests require Docker and OPENAI_API_KEY"
    fi
}


# Function to run coverage tests
run_coverage_tests() {
    print_header "Running Coverage Analysis..."

    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    # Create coverage directory
    mkdir -p coverage

    # Run tests with coverage across all packages
    print_status "Running tests with coverage..."
    if go test -coverprofile=coverage/coverage.out -covermode=atomic -coverpkg=./src/pkg/... ./tests/unit/... ./tests/integration/... ./src/pkg/mcp/...; then
        print_status "Generating coverage report..."

        # Generate HTML coverage report
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html

        # Generate text coverage report
        go tool cover -func=coverage/coverage.out > coverage/coverage.txt

        # Show coverage summary
        print_status "Coverage Summary:"
        go tool cover -func=coverage/coverage.out | grep total | awk '{print "  Total Coverage: " $3}'

        print_success "Coverage analysis completed!"
        print_status "Coverage files available in:"
        echo "  - coverage/coverage.html (HTML report)"
        echo "  - coverage/coverage.txt (Text report)"
        echo "  - coverage/coverage.out (Raw coverage data)"
    else
        print_error "Coverage analysis failed!"
        exit 1
    fi
}

# Run tests based on command
if [ "$RUN_UNIT_TESTS" = true ]; then
    run_unit_tests
fi

if [ "$RUN_INTEGRATION_TESTS" = true ]; then
    run_integration_tests
fi

if [ "$RUN_E2E_TESTS" = true ]; then
    run_e2e_tests
fi

# Run coverage tests if requested
if [ "$RUN_COVERAGE" = true ]; then
    run_coverage_tests
fi

print_status "All requested tests completed!"
exit 0