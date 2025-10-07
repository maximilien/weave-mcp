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

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[START]${NC} $1"
}

print_help() {
    echo -e "${BLUE}Weave MCP Server Start Script${NC}"
    echo ""
    echo "Usage: ./start.sh [MODE] [OPTIONS]"
    echo ""
    echo "Modes:"
    echo "  http        Start the HTTP MCP server (default)"
    echo "  stdio       Start the stdio MCP server (for MCP clients)"
    echo "  inspector   Start the MCP Inspector"
    echo "  both        Start both HTTP server and inspector"
    echo "  all         Start HTTP server only (stdio needs MCP client)"
    echo ""
    echo "Options:"
    echo "  --build     Build the server before starting"
    echo "  --daemon    Start as daemon process (HTTP mode only)"
    echo "  --help, -h  Show this help message"
    echo ""
    echo "This script will:"
    echo "  • Check if the server binaries exist"
    echo "  • Optionally build the servers"
    echo "  • Start the MCP server(s) and/or inspector"
    echo ""
    echo "Examples:"
    echo "  ./start.sh                # Start the HTTP server"
    echo "  ./start.sh http           # Start the HTTP server"
    echo "  ./start.sh stdio          # Show stdio server config for MCP clients"
    echo "  ./start.sh inspector      # Start the MCP Inspector"
    echo "  ./start.sh both           # Start both HTTP server and inspector"
    echo "  ./start.sh all            # Start HTTP server (see stdio config)"
    echo "  ./start.sh --build        # Build and start the HTTP server"
    echo "  ./start.sh http --daemon  # Start HTTP server as daemon"
    echo ""
    echo "Note: stdio servers are designed to be launched by MCP clients like"
    echo "Claude Desktop, not as standalone daemons. They communicate via stdin/stdout."
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check if HTTP server is already running
check_http_server_running() {
    if [ -f "weave-mcp.pid" ]; then
        PID=$(cat weave-mcp.pid)
        if ps -p "$PID" > /dev/null 2>&1; then
            print_warning "HTTP server is already running with PID $PID"
            return 0
        else
            print_status "Removing stale HTTP server PID file"
            rm -f weave-mcp.pid
        fi
    fi
    return 1
}

# Function to check if stdio server is already running
check_stdio_server_running() {
    if [ -f "weave-mcp-stdio.pid" ]; then
        PID=$(cat weave-mcp-stdio.pid)
        if ps -p "$PID" > /dev/null 2>&1; then
            print_warning "stdio server is already running with PID $PID"
            return 0
        else
            print_status "Removing stale stdio server PID file"
            rm -f weave-mcp-stdio.pid
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

# Function to check if inspector is already running
check_inspector_running() {
    if [ -f "mcp-inspector.pid" ]; then
        PID=$(cat mcp-inspector.pid)
        if ps -p "$PID" > /dev/null 2>&1; then
            print_warning "MCP Inspector is already running with PID $PID"
            return 0
        else
            print_status "Removing stale inspector PID file"
            rm -f mcp-inspector.pid
        fi
    fi
    return 1
}

# Function to start the MCP Inspector
start_inspector() {
    print_header "Starting MCP Inspector..."
    
    # Check if Node.js is available for npx
    if ! command -v node >/dev/null 2>&1; then
        print_error "Node.js is required for MCP Inspector but not installed."
        print_status "Please install Node.js from https://nodejs.org/"
        exit 1
    fi
    
    # Check if inspector config exists
    if [ ! -f "tools/mcp-inspector-config.json" ]; then
        print_error "MCP Inspector configuration not found."
        print_status "Run './setup.sh' first to create the configuration."
        exit 1
    fi
    
    # Check if inspector is already running
    if check_inspector_running; then
        print_status "MCP Inspector is already running. Use './stop.sh inspector' to stop it first."
        exit 0
    fi
    
    # Check if HTTP server is running (inspector needs server)
    if ! check_http_server_running; then
        print_warning "HTTP MCP Server is not running. Starting server first..."
        start_http_server --daemon
        sleep 3
    fi
    
    print_status "Starting MCP Inspector..."
    print_status "  • Using official npx method"
    print_status "  • Server: weave-mcp-http (HTTP transport)"
    print_status "  • Server URL: http://localhost:8030"
    print_status "  • Note: You can switch to weave-mcp-stdio in the inspector UI"
    print_status ""

    # Start inspector using npx in background
    # Note: MCP inspector requires both --config and --server parameters
    # Set BROWSER=none to prevent auto-opening browser which can cause errors
    # Use stdio server by default as it works better with the inspector
    (BROWSER=none npx --cache /tmp/npm-cache-new @modelcontextprotocol/inspector --config tools/mcp-inspector-config.json --server weave-mcp-stdio > mcp-inspector.log 2>&1 &)
    INSPECTOR_PID=$!

    # Save PID
    echo $INSPECTOR_PID > mcp-inspector.pid

    # Wait a moment to check if inspector started successfully
    # Inspector needs time to download dependencies and start up
    print_status "Waiting for inspector to start..."
    sleep 8

    # Check if inspector is responding
    if curl -s http://localhost:6274 > /dev/null 2>&1; then
        print_success "MCP Inspector started successfully!"
        print_status ""

        # Extract auth token from log
        local auth_token
        auth_token=$(grep -o 'MCP_PROXY_AUTH_TOKEN=[a-f0-9]*' mcp-inspector.log | head -1 | cut -d= -f2)

        if [ -n "$auth_token" ]; then
            print_success "🚀 Open in your browser:"
            print_status "   http://localhost:6274/?MCP_PROXY_AUTH_TOKEN=$auth_token"
        else
            print_status "  • Inspector URL: http://localhost:6274"
            print_status "  • Auth token: Check mcp-inspector.log for session token"
        fi

        print_status ""
        print_status "  • Log file: mcp-inspector.log"
        print_status "  • Use './stop.sh inspector' to stop the inspector"
        print_status "  • Use './tools/tail-logs.sh inspector' to monitor logs"
    else
        print_error "Failed to start MCP Inspector"
        print_status "Check mcp-inspector.log for error details"
        exit 1
    fi
}

# Function to start both HTTP server and inspector
start_both() {
    print_header "Starting both HTTP MCP Server and Inspector..."

    # Start HTTP server first
    start_http_server --daemon

    # Wait a moment for server to start
    sleep 3

    # Start inspector
    start_inspector
}

# Function to start all servers (HTTP and stdio)
start_all_servers() {
    print_header "Starting HTTP MCP Server..."
    print_warning "Note: stdio servers cannot run as daemons - they are meant to be launched by MCP clients"
    print_status "Only starting HTTP server in daemon mode"
    print_status ""

    # Start HTTP server in daemon mode
    start_http_server --daemon

    print_status ""
    print_status "To use the stdio server, configure it in your MCP client (e.g., Claude Desktop)"
    print_status "See './start.sh stdio' for configuration instructions"
}

# Function to start the HTTP server
start_http_server() {
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
        print_error "HTTP server binary not found. Run './build.sh' first or use '--build' option"
        exit 1
    fi

    # Check if server is already running
    if check_http_server_running; then
        print_status "HTTP server is already running. Use './stop.sh http' to stop it first."
        exit 0
    fi

    print_header "Starting Weave HTTP MCP Server..."

    # Get server configuration
    HOST=${MCP_SERVER_HOST:-localhost}
    PORT=${MCP_SERVER_PORT:-8030}

    print_status "HTTP server configuration:"
    print_status "  • Host: $HOST"
    print_status "  • Port: $PORT"
    print_status "  • Binary: bin/weave-mcp"
    print_status "  • Logs: logs/weave-mcp.log"

    if [ "$daemon_mode" = true ]; then
        print_status "Starting HTTP server in daemon mode..."

        # Start server in background
        nohup ./bin/weave-mcp > weave-mcp.log 2>&1 &
        PID=$!

        # Save PID
        echo $PID > weave-mcp.pid

        # Wait a moment to check if server started successfully
        sleep 2

        if ps -p "$PID" > /dev/null 2>&1; then
        print_status "HTTP server started successfully in daemon mode!"
        print_status "  • PID: $PID"
        print_status "  • Log file: logs/weave-mcp.log"
        print_status "  • PID file: weave-mcp.pid"
        print_status "  • Server URL: http://$HOST:$PORT"
        print_status ""
        print_status "Use './stop.sh http' to stop the server"
        print_status "Use './tools/tail-logs.sh' to monitor logs"
        else
            print_error "Failed to start HTTP server in daemon mode"
            print_status "Check weave-mcp.log for error details"
            rm -f weave-mcp.pid
            exit 1
        fi
    else
        print_status "Starting HTTP server in foreground mode..."
        print_status "Press Ctrl+C to stop the server"
        print_status ""
        print_status "Server URL: http://$HOST:$PORT"
        print_status "Health check: http://$HOST:$PORT/health"
        print_status "MCP tools: http://$HOST:$PORT/mcp/tools/list"
        print_status "Logs: logs/weave-mcp.log"
        print_status "Monitor logs: ./tools/tail-logs.sh"
        print_status ""

        # Start server in foreground
        ./bin/weave-mcp
    fi
}

# Function to start the stdio server
start_stdio_server() {
    local daemon_mode=false

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --build)
                build_server
                shift
                ;;
            --daemon)
                print_warning "stdio servers communicate via stdin/stdout and cannot run as daemons"
                print_status "The stdio server will run in foreground mode"
                print_status "Use MCP clients like Claude Desktop to connect to it"
                daemon_mode=false
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
    if [ ! -f "bin/weave-mcp-stdio" ]; then
        print_error "stdio server binary not found. Run './build.sh' first or use '--build' option"
        exit 1
    fi

    print_header "Starting Weave stdio MCP Server..."

    print_status "stdio server configuration:"
    print_status "  • Binary: bin/weave-mcp-stdio"
    print_status "  • Logs: logs/weave-mcp-stdio.log"
    print_status "  • Transport: stdio (stdin/stdout)"
    print_status ""
    print_warning "Note: stdio servers are designed to be launched by MCP clients"
    print_warning "They communicate via stdin/stdout and will wait for input"
    print_status ""
    print_status "To use with Claude Desktop, add this to your MCP settings:"
    print_status "{"
    print_status "  \"mcpServers\": {"
    print_status "    \"weave-mcp\": {"
    print_status "      \"command\": \"$(pwd)/bin/weave-mcp-stdio\","
    print_status "      \"args\": []"
    print_status "    }"
    print_status "  }"
    print_status "}"
    print_status ""
    print_status "Starting stdio server in foreground mode..."
    print_status "Press Ctrl+C to stop the server"
    print_status ""

    # Start server in foreground
    ./bin/weave-mcp-stdio
}

# Main execution
main() {
    local mode="http"
    local args=()

    # Parse mode and arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            http|stdio|inspector|both|all)
                mode="$1"
                shift
                ;;
            --help|-h)
                print_help
                exit 0
                ;;
            *)
                args+=("$1")
                shift
                ;;
        esac
    done

    # Execute based on mode
    case $mode in
        http)
            start_http_server "${args[@]}"
            ;;
        stdio)
            start_stdio_server "${args[@]}"
            ;;
        inspector)
            start_inspector
            ;;
        both)
            start_both
            ;;
        all)
            start_all_servers
            ;;
        *)
            print_error "Unknown mode: $mode"
            print_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"