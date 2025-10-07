#!/bin/bash

# Weave MCP Server Setup Script
# Installs MCP Inspector and other dependencies

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[SETUP]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_header() {
    echo -e "${BLUE}================================${NC}"
    echo -e "${BLUE}  Weave MCP Server Setup${NC}"
    echo -e "${BLUE}================================${NC}"
    echo ""
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Setup MCP Inspector
setup_mcp_inspector() {
    print_status "Setting up MCP Inspector..."
    
    # Check if Node.js is installed
    if ! command_exists node; then
        print_error "Node.js is required for MCP Inspector but not installed."
        print_status "Please install Node.js from https://nodejs.org/"
        print_status "Or install via package manager:"
        print_status "  macOS: brew install node"
        print_status "  Ubuntu: sudo apt install nodejs npm"
        print_status "  CentOS: sudo yum install nodejs npm"
        exit 1
    fi
    
    # Check Node.js version (require 22.7.5+ for MCP Inspector)
    NODE_VERSION=$(node --version | cut -d'v' -f2)
    NODE_MAJOR=$(echo "$NODE_VERSION" | cut -d'.' -f1)
    NODE_MINOR=$(echo "$NODE_VERSION" | cut -d'.' -f2)
    NODE_PATCH=$(echo "$NODE_VERSION" | cut -d'.' -f3)
    
    # Check if version is 22.7.5 or higher
    if [ "$NODE_MAJOR" -lt 22 ] || { [ "$NODE_MAJOR" -eq 22 ] && [ "$NODE_MINOR" -lt 7 ]; } || { [ "$NODE_MAJOR" -eq 22 ] && [ "$NODE_MINOR" -eq 7 ] && [ "$NODE_PATCH" -lt 5 ]; }; then
        print_error "Node.js version 22.7.5+ is required for MCP Inspector. Current version: $(node --version)"
        print_status ""
        print_status "Your Node.js version ($(node --version)) is actually quite recent and stable!"
        print_status "However, MCP Inspector requires Node.js 22.7.5+ (latest is v24.9.0)"
        print_status ""
        print_status "To upgrade Node.js:"
        print_status "  • macOS: brew install node@22 (or brew install node for latest)"
        print_status "  • Ubuntu: curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash - && sudo apt-get install -y nodejs"
        print_status "  • Or download from: https://nodejs.org/"
        print_status ""
        print_status "Note: The MCP server works perfectly with your current Node.js version!"
        print_warning "MCP Inspector setup will be skipped due to Node.js version requirement"
        return 1
    fi
    
    print_success "Node.js $(node --version) detected"
    
    # Check if npm is available
    if ! command_exists npm; then
        print_error "npm is required but not found. Please install npm."
        exit 1
    fi
    
    print_success "npm $(npm --version) detected"
    
    print_status "Using official npx method - no local installation needed"
    
    # Create inspector configuration
    print_status "Creating MCP Inspector configuration..."
    cat > tools/mcp-inspector-config.json << 'EOF'
{
  "mcpServers": {
    "weave-mcp": {
      "command": "node",
      "args": ["../bin/weave-mcp"],
      "env": {
        "WEAVIATE_URL": "${WEAVIATE_URL}",
        "WEAVIATE_API_KEY": "${WEAVIATE_API_KEY}",
        "WEAVIATE_COLLECTION": "${WEAVIATE_COLLECTION}",
        "WEAVIATE_COLLECTION_IMAGES": "${WEAVIATE_COLLECTION_IMAGES}"
      }
    }
  }
}
EOF
    
    print_success "MCP Inspector configuration created at tools/mcp-inspector-config.json"
}

# Setup Python dependencies (for MCP Inspector if needed)
setup_python_deps() {
    print_status "Checking Python dependencies..."
    
    if ! command_exists python3; then
        print_warning "Python 3 not found. MCP Inspector may require Python."
        print_status "Please install Python 3 from https://python.org/"
        return 0
    fi
    
    print_success "Python $(python3 --version) detected"
    
    # Check if pip is available
    if command_exists pip3; then
        print_success "pip3 $(pip3 --version) detected"
    elif command_exists pip; then
        print_success "pip $(pip --version) detected"
    else
        print_warning "pip not found. Python packages may not install correctly."
    fi
}

# Main setup function
main() {
    print_header
    
    print_status "Starting Weave MCP Server setup..."
    echo ""
    
    # Check if we're in the right directory
    if [ ! -f "src/main.go" ] || [ ! -f "config.yaml" ]; then
        print_error "This doesn't appear to be the Weave MCP Server directory."
        print_status "Please run this script from the project root directory."
        exit 1
    fi
    
    # Setup MCP Inspector
    if setup_mcp_inspector; then
        print_success "MCP Inspector setup completed"
    else
        print_warning "MCP Inspector setup failed or skipped"
    fi
    echo ""
    
    # Setup Python dependencies
    setup_python_deps
    echo ""
    
    # Check if Go is available for building
    if command_exists go; then
        print_success "Go $(go version) detected"
        print_status "Building Weave MCP Server..."
        if ./build.sh; then
            print_success "Weave MCP Server built successfully"
        else
            print_warning "Failed to build Weave MCP Server"
        fi
    else
        print_warning "Go not found. Please install Go to build the server."
        print_status "Download from: https://golang.org/dl/"
    fi
    
    echo ""
    print_success "Setup completed!"
    echo ""
    print_status "Next steps:"
    print_status "1. Configure your .env file with Weaviate credentials"
    print_status "2. Start the MCP server: ./start.sh"
    print_status "3. Start the MCP Inspector: ./start.sh inspector"
    print_status "4. Or run both together: ./start.sh both"
    echo ""
    print_status "For more information, see README.md"
}

# Run main function
main "$@"