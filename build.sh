#!/usr/bin/env bash

# Load environment variables from .env file
set -a
if [ -f .env ]; then
    # shellcheck disable=SC1091
    . .env
fi
set +a

# Weave MCP Server Build Script
# Usage: ./build.sh [clean|all|help]

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
    echo -e "${BLUE}[BUILD]${NC} $1"
}

print_help() {
    echo -e "${BLUE}Weave MCP Server Build Script${NC}"
    echo ""
    echo "Usage: ./build.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  all          Build the MCP server binary (default)"
    echo "  clean        Clean all build artifacts"
    echo "  --help, -h   Show this help message"
    echo ""
    echo "This script will:"
    echo "  • Build Golang MCP server binary"
    echo "  • Generate optimized production builds"
    echo "  • Create build information"
    echo ""
    echo "Examples:"
    echo "  ./build.sh                    # Build everything"
    echo "  ./build.sh clean              # Clean build artifacts"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to build the MCP server
build_server() {
    print_header "Building Weave MCP Server..."
    
    # Check if Go is installed
    if ! command_exists go; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        print_status "Visit: https://golang.org/dl/"
        exit 1
    fi
    
    # Check Go version
    go_version=$(go version | awk '{print $3}' | sed 's/go//')
    required_version="1.21"
    
    if ! printf '%s\n%s\n' "$required_version" "$go_version" | sort -V -C; then
        print_error "Go version $go_version is too old. Please install Go $required_version or later."
        exit 1
    fi
    
    # Download dependencies
    print_status "Downloading Go dependencies..."
    if ! go mod tidy; then
        print_error "Failed to download Go dependencies"
        exit 1
    fi
    
    # Run tests first
    print_status "Running Go tests..."
    if ! go test ./src/...; then
        print_warning "Some tests failed, but continuing with build..."
    fi
    
    # Create bin directory if it doesn't exist
    mkdir -p bin
    
    # Get version information from git
    print_status "Getting version information from git..."
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date +"%Y-%m-%dT%H:%M:%SZ")
    
    # Clean version string (remove 'v' prefix if present)
    if [[ $VERSION == v* ]]; then
        VERSION=${VERSION#v}
    fi
    
    print_status "Version: $VERSION"
    print_status "Git Commit: ${GIT_COMMIT:0:7}"
    print_status "Build Time: $BUILD_TIME"
    
    # Build the MCP server binary with version information
    print_status "Building MCP server binary..."
    LDFLAGS="-X main.version=$VERSION"
    LDFLAGS="$LDFLAGS -X main.gitCommit=$GIT_COMMIT"
    LDFLAGS="$LDFLAGS -X main.buildTime=$BUILD_TIME"
    
    if ! go build -ldflags "$LDFLAGS" -o bin/weave-mcp ./src/main.go; then
        print_error "MCP server build failed"
        exit 1
    fi
    
    # Make binary executable
    chmod +x bin/weave-mcp
    
    print_status "✅ MCP server built successfully!"
    print_status "   • Binary: bin/weave-mcp"
    print_status "   • Size: $(stat -f%z bin/weave-mcp 2>/dev/null | numfmt --to=iec || echo "unknown")"
}

# Function to clean build artifacts
clean_build() {
    print_header "Cleaning build artifacts..."
    
    # Clean binary
    if [ -d "bin" ]; then
        print_status "Removing bin directory..."
        rm -rf bin
    fi
    
    if [ -d "build" ]; then
        print_status "Removing build directory..."
        rm -rf build
    fi
    
    print_status "Cleaning Go build cache..."
    go clean -cache
    go clean -modcache
    
    print_status "✅ Build artifacts cleaned!"
}

# Function to create build info
create_build_info() {
    print_status "Creating build information..."
    
    # Get version information
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || date +"%Y-%m-%dT%H:%M:%SZ")
    
    # Clean version string (remove 'v' prefix if present)
    if [[ $VERSION == v* ]]; then
        VERSION=${VERSION#v}
    fi
    
    BUILD_INFO_FILE="bin/build-info.txt"
    cat > "$BUILD_INFO_FILE" << EOF
Weave MCP Server Build Information
===================================

Version: $VERSION
Build Date: $(date)
Build Time: $BUILD_TIME
Build Host: $(hostname)
Git Branch: $(git branch --show-current 2>/dev/null || echo "unknown")
Git Commit: $GIT_COMMIT
Git Status: $(git status --porcelain 2>/dev/null | wc -l | xargs echo) files changed

MCP Server Build:
$(if [ -f "bin/weave-mcp" ]; then
    echo "  Status: Built"
    echo "  Version: $VERSION"
    echo "  Go Version: $(go version 2>/dev/null || echo "unknown")"
    echo "  Binary: bin/weave-mcp ($(stat -f%z bin/weave-mcp 2>/dev/null | numfmt --to=iec || echo "unknown"))"
else
    echo "  Status: Not built"
fi)

Environment Variables:
$(if [ -f .env ]; then
    echo "  .env file: Present"
    echo "  Variables loaded: $(grep -c '^[A-Z]' .env 2>/dev/null || echo "0")"
else
    echo "  .env file: Not found"
fi)

Configuration:
$(if [ -f config.yaml ]; then
    echo "  config.yaml: Present"
else
    echo "  config.yaml: Not found"
fi)

Features:
  Vector Database Support: Weaviate, Milvus, Mock
  MCP Tools: list_collections, create_collection, delete_collection, list_documents, create_document, get_document, delete_document, count_documents, query_documents
  Configuration: YAML + Environment Variables
  Mock Database: For testing and development
  Embedding Support: OpenAI, Custom Local

EOF

    print_status "Build information saved to: $BUILD_INFO_FILE"
}

# Function to build everything
build_all() {
    print_header "Building Weave MCP Server..."
    echo "====================================="
    
    # Build MCP server
    build_server
    
    # Create build information
    create_build_info
    
    print_header "Build Complete! 🎉"
    echo "====================================="
    print_status "Weave MCP Server has been built successfully!"
    echo ""
    print_status "Built components:"
    print_status "  • MCP Server Binary: bin/weave-mcp"
    print_status "  • Build Info: bin/build-info.txt"
    echo ""
    print_status "Next steps:"
    print_status "  1. Run './bin/weave-mcp' to start the server"
    print_status "  2. Check 'bin/build-info.txt' for detailed build information"
    print_status "  3. Run './test.sh' to run tests"
    print_status "  4. Run './lint.sh' to check code quality"
}

# Main script logic
case "${1:-all}" in
    "all"|"")
        build_all
        ;;
    "clean")
        clean_build
        ;;
    "--help"|"-h")
        print_help
        ;;
    *)
        print_error "Unknown option: $1"
        print_help
        exit 1
        ;;
esac