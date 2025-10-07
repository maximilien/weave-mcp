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
    echo "  unit        Run only unit tests"
    echo "  integration Run only integration tests"
    echo "  fast        Run fast tests (unit + mock + MCP integration)"
    echo "  all         Run all tests (unit + integration)"
    echo "  coverage    Run tests with coverage report"
    echo "  help        Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./test.sh unit         # Run only unit tests"
    echo "  ./test.sh integration  # Run only integration tests"
    echo "  ./test.sh fast         # Run fast tests (unit + mock + MCP integration)"
    echo "  ./test.sh all          # Run all tests"
    echo "  ./test.sh coverage     # Run tests with coverage report"
    echo "  ./test.sh              # Run unit tests (default)"
    echo ""
    echo "Test Categories:"
    echo "  Unit Tests:"
    echo "    - Configuration management testing"
    echo "    - Mock client testing"
    echo "    - Utility function testing"
    echo ""
    echo "  Integration Tests:"
    echo "    - Mock client testing"
    echo "    - MCP integration with Weaviate Cloud"
    echo "    - Vector database client testing"
    echo "    - End-to-end workflow testing"
}

# Initialize variables
RUN_UNIT_TESTS=false
RUN_INTEGRATION_TESTS=false
RUN_COVERAGE=false

# Check command line arguments
case "${1:-unit}" in
    "unit")
        RUN_UNIT_TESTS=true
        RUN_INTEGRATION_TESTS=false
        RUN_COVERAGE=false
        ;;
    "integration")
        RUN_UNIT_TESTS=false
        RUN_INTEGRATION_TESTS=true
        RUN_COVERAGE=false
        ;;
    "fast")
        RUN_UNIT_TESTS=true
        RUN_INTEGRATION_TESTS=true
        RUN_COVERAGE=false
        ;;
    "all")
        RUN_UNIT_TESTS=true
        RUN_INTEGRATION_TESTS=true
        RUN_COVERAGE=false
        ;;
    "coverage")
        RUN_UNIT_TESTS=false
        RUN_INTEGRATION_TESTS=false
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
    print_header "Running Unit Tests..."
    
    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Run basic unit tests
    print_status "Running basic unit tests..."
    if go test -v -timeout=30s ./tests/... -run="TestConfig|TestMock|TestVectorDB"; then
        print_success "Basic unit tests passed!"
    else
        print_error "Basic unit tests failed!"
        exit 1
    fi
    
    # Run extended unit tests if available
    print_status "Running extended unit tests..."
    if go test -v -timeout=30s ./tests/... -run="TestConfigExtended|TestMockExtended"; then
        print_success "Extended unit tests passed!"
    else
        print_warning "Extended unit tests failed or not found"
    fi
}

# Function to run integration tests
run_integration_tests() {
    print_header "Running Integration Tests..."
    
    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Run fast integration tests (mock only)
    print_status "Running fast integration tests (mock)..."
    if go test -v -timeout=10s ./tests/... -run="TestMock"; then
        print_success "Fast integration tests passed!"
    else
        print_warning "Fast integration tests failed"
    fi
    
    # Run MCP integration tests (if Weaviate is configured)
    print_status "Running MCP integration tests (Weaviate Cloud)..."
    if go test -v -timeout=60s ./tests/... -run="TestFastMCPIntegration|TestMCPToolCallViaHTTP"; then
        print_success "MCP integration tests passed!"
    else
        print_warning "MCP integration tests failed or skipped (check Weaviate configuration)"
    fi
    
    # Run Weaviate integration tests if configured
    if [ -n "$WEAVIATE_URL" ] && [ "$WEAVIATE_URL" != "http://localhost:8080" ]; then
        print_status "Running Weaviate integration tests..."
        if go test -v -timeout=30s ./tests/... -run="TestWeaviateIntegration|TestWeaviateConnectionSpeed"; then
            print_success "Weaviate integration tests passed!"
        else
            print_warning "Weaviate integration tests failed"
        fi
    else
        print_warning "Skipping Weaviate integration tests - no credentials provided"
        print_status "Set WEAVIATE_URL to run Weaviate tests"
    fi
}

# Function to run fast tests
run_fast_tests() {
    print_header "Running Fast Tests..."
    
    # Check if Go is installed
    if ! command -v go >/dev/null 2>&1; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi
    
    # Run unit tests
    print_status "Running unit tests..."
    if go test -v -timeout=30s ./tests/... -run="TestConfig|TestMock|TestVectorDB"; then
        print_success "Unit tests passed!"
    else
        print_error "Unit tests failed!"
        exit 1
    fi
    
    # Run fast integration tests (mock only)
    print_status "Running fast integration tests (mock)..."
    if go test -v -timeout=10s ./tests/... -run="TestMock"; then
        print_success "Fast integration tests passed!"
    else
        print_warning "Fast integration tests failed"
    fi
    
    # Run MCP integration tests (if Weaviate is configured)
    print_status "Running MCP integration tests (Weaviate Cloud)..."
    if go test -v -timeout=60s ./tests/... -run="TestFastMCPIntegration|TestMCPToolCallViaHTTP"; then
        print_success "MCP integration tests passed!"
    else
        print_warning "MCP integration tests failed or skipped (check Weaviate configuration)"
    fi
    
    print_success "Fast tests completed!"
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
    
    # Run tests with coverage (only unit tests and mock integration tests)
    print_status "Running tests with coverage..."
    if go test -coverprofile=coverage/coverage.out -covermode=atomic ./tests/... -run="TestConfig|TestMock|TestMCP|TestFastMock|TestFastConfig"; then
        print_status "Generating coverage report..."
        
        # Generate HTML coverage report
        go tool cover -html=coverage/coverage.out -o coverage/coverage.html
        
        # Generate text coverage report
        go tool cover -func=coverage/coverage.out > coverage/coverage.txt
        
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
if [ "$RUN_UNIT_TESTS" = true ] && [ "$RUN_INTEGRATION_TESTS" = true ]; then
    # Check if this is a fast test run
    if [ "${1:-unit}" = "fast" ]; then
        run_fast_tests
    else
        run_unit_tests
        run_integration_tests
    fi
elif [ "$RUN_UNIT_TESTS" = true ]; then
    run_unit_tests
elif [ "$RUN_INTEGRATION_TESTS" = true ]; then
    run_integration_tests
fi

# Run coverage tests if requested
if [ "$RUN_COVERAGE" = true ]; then
    run_coverage_tests
fi

print_status "All requested tests completed!"
exit 0