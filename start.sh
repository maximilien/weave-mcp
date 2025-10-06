#!/usr/bin/env bash

# Load environment variables from .env file
set -a
if [ -f .env ]; then
    # shellcheck disable=SC1091
    . .env
fi
set +a

# Weave MCP Server Start Script

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
    echo -e "${BLUE}[START]${NC} $1"
}

print_help() {
    echo -e "${BLUE}Weave MCP Server Start Script${NC}"
    echo ""
    echo "Usage: ./start.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --build     Build the server before starting"
    echo "  --daemon    Start as daemon process"
    echo "  --help, -h  Show this help message"
    echo ""
    echo "This script will:"
    echo "  • Check if the server binary exists"
    echo "  • Optionally build the server"
    echo "  • Start the MCP server"
    echo ""
    echo "Examples:"
    echo "  ./start.sh                # Start the server"
    echo "  ./start.sh --build        # Build and start the server"
    echo "  ./start.sh --daemon       # Start as daemon"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if server is already running
check_server_running() {
    if [ -f "weave-mcp.pid" ]; then
        PID=$(cat weave-mcp.pid)
        if ps -p "$PID" > /dev/null 2>&1; then
            print_warning "Server is already running with PID $PID"
            return 0
        else
            print_status "Removing stale PID file"
            rm -f weave-mcp.pid
        fi
    fi
    return 1
}

# Function to build the server
build_server() {
    print_status "Building server..."
    if ! ./build.sh; then
        print_error "Failed to build server"
        exit 1
    fi
    print_status "Server built successfully"
}

# Function to start the server
start_server() {
    local daemon_mode=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --build)
                build_server
                shift
                ;;
            --daemon)
                daemon_mode=true
                shift
                ;;
            --help|-h)
                print_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                print_help
                exit 1
                ;;
        esac
    done
    
    # Check if server binary exists
    if [ ! -f "bin/weave-mcp" ]; then
        print_error "Server binary not found. Run './build.sh' first or use '--build' option"
        exit 1
    fi
    
    # Check if server is already running
    if check_server_running; then
        print_status "Server is already running. Use './stop.sh' to stop it first."
        exit 0
    fi
    
    print_header "Starting Weave MCP Server..."
    
    # Get server configuration
    HOST=${MCP_SERVER_HOST:-localhost}
    PORT=${MCP_SERVER_PORT:-8030}
    
    print_status "Server configuration:"
    print_status "  • Host: $HOST"
    print_status "  • Port: $PORT"
    print_status "  • Binary: bin/weave-mcp"
    
    if [ "$daemon_mode" = true ]; then
        print_status "Starting server in daemon mode..."
        
        # Start server in background
        nohup ./bin/weave-mcp > weave-mcp.log 2>&1 &
        PID=$!
        
        # Save PID
        echo $PID > weave-mcp.pid
        
        # Wait a moment to check if server started successfully
        sleep 2
        
        if ps -p "$PID" > /dev/null 2>&1; then
            print_status "Server started successfully in daemon mode!"
            print_status "  • PID: $PID"
            print_status "  • Log file: weave-mcp.log"
            print_status "  • PID file: weave-mcp.pid"
            print_status "  • Server URL: http://$HOST:$PORT"
            print_status ""
            print_status "Use './stop.sh' to stop the server"
        else
            print_error "Failed to start server in daemon mode"
            print_status "Check weave-mcp.log for error details"
            rm -f weave-mcp.pid
            exit 1
        fi
    else
        print_status "Starting server in foreground mode..."
        print_status "Press Ctrl+C to stop the server"
        print_status ""
        print_status "Server URL: http://$HOST:$PORT"
        print_status "Health check: http://$HOST:$PORT/health"
        print_status "MCP tools: http://$HOST:$PORT/mcp/tools/list"
        print_status ""
        
        # Start server in foreground
        ./bin/weave-mcp
    fi
}

# Main execution
start_server "$@"