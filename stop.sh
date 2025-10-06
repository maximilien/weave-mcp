#!/usr/bin/env bash

# Weave MCP Server Stop Script

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
    echo -e "${BLUE}[STOP]${NC} $1"
}

print_help() {
    echo -e "${BLUE}Weave MCP Server Stop Script${NC}"
    echo ""
    echo "Usage: ./stop.sh [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  stop        Stop the server (default)"
    echo "  status      Check server status"
    echo "  restart     Stop and start the server"
    echo "  --help, -h  Show this help message"
    echo ""
    echo "This script will:"
    echo "  • Check if the server is running"
    echo "  • Stop the server gracefully"
    echo "  • Clean up PID files"
    echo ""
    echo "Examples:"
    echo "  ./stop.sh              # Stop the server"
    echo "  ./stop.sh status       # Check server status"
    echo "  ./stop.sh restart      # Restart the server"
}

# Function to check server status
check_status() {
    if [ -f "weave-mcp.pid" ]; then
        PID=$(cat weave-mcp.pid)
        if ps -p "$PID" > /dev/null 2>&1; then
            print_status "Server is running with PID $PID"
            
            # Get process info
            if command -v ps >/dev/null 2>&1; then
                print_status "Process info:"
                ps -p "$PID" -o pid,ppid,cmd,etime,pcpu,pmem 2>/dev/null || true
            fi
            
            # Check if server is responding
            HOST=${MCP_SERVER_HOST:-localhost}
            PORT=${MCP_SERVER_PORT:-8030}
            
            if command -v curl >/dev/null 2>&1; then
                print_status "Checking server health..."
                if curl -s "http://$HOST:$PORT/health" >/dev/null 2>&1; then
                    print_status "✅ Server is responding to health checks"
                else
                    print_warning "⚠️  Server is running but not responding to health checks"
                fi
            fi
            
            return 0
        else
            print_warning "PID file exists but process is not running"
            rm -f weave-mcp.pid
            return 1
        fi
    else
        print_status "Server is not running (no PID file found)"
        return 1
    fi
}

# Function to stop the server
stop_server() {
    print_header "Stopping Weave MCP Server..."
    
    if [ -f "weave-mcp.pid" ]; then
        PID=$(cat weave-mcp.pid)
        
        if ps -p "$PID" > /dev/null 2>&1; then
            print_status "Stopping server with PID $PID..."
            
            # Try graceful shutdown first
            if kill -TERM "$PID" 2>/dev/null; then
                print_status "Sent SIGTERM signal to server..."
                
                # Wait for graceful shutdown
                for _ in {1..30}; do
                    if ! ps -p "$PID" > /dev/null 2>&1; then
                        print_status "Server stopped gracefully"
                        rm -f weave-mcp.pid
                        return 0
                    fi
                    sleep 1
                done
                
                # Force kill if still running
                print_warning "Server did not stop gracefully, forcing shutdown..."
                if kill -KILL "$PID" 2>/dev/null; then
                    print_status "Sent SIGKILL signal to server..."
                    sleep 2
                    if ! ps -p "$PID" > /dev/null 2>&1; then
                        print_status "Server stopped forcefully"
                        rm -f weave-mcp.pid
                        return 0
                    else
                        print_error "Failed to stop server"
                        return 1
                    fi
                else
                    print_error "Failed to send kill signal to server"
                    return 1
                fi
            else
                print_error "Failed to send TERM signal to server"
                return 1
            fi
        else
            print_warning "PID file exists but process is not running"
            rm -f weave-mcp.pid
            return 0
        fi
    else
        print_status "No PID file found - server may not be running"
        return 0
    fi
}

# Function to restart the server
restart_server() {
    print_header "Restarting Weave MCP Server..."
    
    # Stop the server
    if stop_server; then
        print_status "Server stopped successfully"
    else
        print_warning "Server stop had issues, but continuing with restart"
    fi
    
    # Wait a moment
    sleep 2
    
    # Start the server
    print_status "Starting server..."
    if ./start.sh --daemon; then
        print_status "Server restarted successfully"
    else
        print_error "Failed to restart server"
        return 1
    fi
}

# Main script logic
case "${1:-stop}" in
    "stop")
        stop_server
        ;;
    "status")
        check_status
        ;;
    "restart")
        restart_server
        ;;
    "--help"|"-h")
        print_help
        ;;
    *)
        print_error "Unknown command: $1"
        print_help
        exit 1
        ;;
esac