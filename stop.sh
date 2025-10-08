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

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_header() {
    echo -e "${BLUE}[STOP]${NC} $1"
}

print_help() {
    echo -e "${BLUE}Weave MCP Server Stop Script${NC}"
    echo ""
    echo "Usage: ./stop.sh [COMMAND] [TARGET]"
    echo ""
    echo "Commands:"
    echo "  stop        Stop services (default)"
    echo "  status      Check service status"
    echo "  restart     Stop and start services"
    echo "  --help, -h  Show this help message"
    echo ""
    echo "Targets:"
    echo "  http        Stop/check HTTP MCP server only (default)"
    echo "  stdio       Stop/check stdio MCP server only"
    echo "  inspector   Stop/check MCP Inspector only"
    echo "  all         Stop/check all services"
    echo ""
    echo "This script will:"
    echo "  • Check if services are running"
    echo "  • Stop services gracefully"
    echo "  • Clean up PID files"
    echo ""
    echo "Examples:"
    echo "  ./stop.sh                    # Stop the HTTP server"
    echo "  ./stop.sh stop http          # Stop the HTTP server"
    echo "  ./stop.sh stop stdio         # Stop the stdio server"
    echo "  ./stop.sh stop inspector     # Stop the inspector"
    echo "  ./stop.sh stop all           # Stop all services"
    echo "  ./stop.sh status             # Check HTTP server status"
    echo "  ./stop.sh status all         # Check all services status (enhanced)"
    echo "  ./stop.sh restart http       # Restart the HTTP server"
    echo "  ./stop.sh restart all        # Restart all services"
    echo ""
    echo "Enhanced Status Features:"
    echo "  • Health checks for all services"
    echo "  • Process information and resource usage"
    echo "  • Service URLs and accessibility"
    echo "  • Comprehensive summary with status indicators"
}

# Function to check inspector status
check_inspector_status() {
    print_header "MCP Inspector Status"
    echo ""
    
    # Check if inspector is running (look for npx processes)
    INSPECTOR_PID=$(pgrep -f "npx.*inspector" | head -1)
    if [ -n "$INSPECTOR_PID" ]; then
        print_success "✅ MCP Inspector is running with PID $INSPECTOR_PID"
        
        # Check if inspector is responding
        if curl -s http://localhost:6274 > /dev/null 2>&1; then
            print_success "   • Web interface: OK"
            print_status "   • URL: http://localhost:6274"
        else
            print_warning "   • Web interface: Not responding (may be starting up)"
        fi
        
        # Get process info
        if command -v ps >/dev/null 2>&1; then
            print_status "   • Process info:"
            ps -p "$INSPECTOR_PID" -o pid,ppid,cmd,etime,pcpu,pmem 2>/dev/null | tail -1 | sed 's/^/     /'
        fi
        
        return 0
    else
        print_error "❌ MCP Inspector is not running"
        return 1
    fi
}

# Function to check stdio server status
check_stdio_server_status() {
    print_header "stdio MCP Server Status"
    echo ""

    if [ -f "weave-mcp-stdio.pid" ]; then
        PID=$(cat weave-mcp-stdio.pid)
        if ps -p "$PID" > /dev/null 2>&1; then
            print_success "✅ stdio server is running with PID $PID"

            # Get process info
            if command -v ps >/dev/null 2>&1; then
                print_status "   • Process info:"
                ps -p "$PID" -o pid,ppid,cmd,etime,pcpu,pmem 2>/dev/null | tail -1 | sed 's/^/     /'
            fi

            return 0
        else
            print_error "❌ stdio server is not running (stale PID file)"
            rm -f weave-mcp-stdio.pid
            return 1
        fi
    else
        print_error "❌ stdio server is not running (no PID file)"
        return 1
    fi
}

# Function to check all services status
check_all_status() {
    print_header "Weave MCP Services Status"
    echo ""

    # Check HTTP MCP Server
    print_status "🔧 HTTP MCP Server Status:"
    if [ -f "weave-mcp.pid" ]; then
        PID=$(cat weave-mcp.pid)
        if ps -p "$PID" > /dev/null 2>&1; then
            print_success "✅ Running (PID: $PID)"

            # Check if server is responding
            if curl -s http://localhost:8030/health > /dev/null 2>&1; then
                print_success "   • Health check: OK"
                print_status "   • URL: http://localhost:8030"
            else
                print_warning "   • Health check: Failed (server may be starting up)"
            fi

            # Get process info
            if command -v ps >/dev/null 2>&1; then
                print_status "   • Process info:"
                ps -p "$PID" -o pid,ppid,cmd,etime,pcpu,pmem 2>/dev/null | tail -1 | sed 's/^/     /'
            fi
        else
            print_error "❌ Not running (stale PID file)"
            rm -f weave-mcp.pid
        fi
    else
        print_error "❌ Not running (no PID file)"
    fi
    echo ""

    # Check stdio MCP Server
    print_status "🔧 stdio MCP Server Status:"
    if [ -f "weave-mcp-stdio.pid" ]; then
        PID=$(cat weave-mcp-stdio.pid)
        if ps -p "$PID" > /dev/null 2>&1; then
            print_success "✅ Running (PID: $PID)"

            # Get process info
            if command -v ps >/dev/null 2>&1; then
                print_status "   • Process info:"
                ps -p "$PID" -o pid,ppid,cmd,etime,pcpu,pmem 2>/dev/null | tail -1 | sed 's/^/     /'
            fi
        else
            print_error "❌ Not running (stale PID file)"
            rm -f weave-mcp-stdio.pid
        fi
    else
        print_error "❌ Not running (no PID file)"
    fi
    echo ""

    # Check MCP Inspector
    print_status "🔍 MCP Inspector Status:"
    INSPECTOR_PID=$(pgrep -f "npx.*inspector" | head -1)
    if [ -n "$INSPECTOR_PID" ]; then
        print_success "✅ Running (PID: $INSPECTOR_PID)"

        # Check if inspector is responding
        if curl -s http://localhost:6274 > /dev/null 2>&1; then
            print_success "   • Web interface: OK"
            print_status "   • URL: http://localhost:6274"
        else
            print_warning "   • Web interface: Not responding (may be starting up)"
        fi

        # Get process info
        if command -v ps >/dev/null 2>&1; then
            print_status "   • Process info:"
            ps -p "$INSPECTOR_PID" -o pid,ppid,cmd,etime,pcpu,pmem 2>/dev/null | tail -1 | sed 's/^/     /'
        fi
    else
        print_error "❌ Not running"
    fi
    echo ""

    # Summary
    print_status "📊 Summary:"
    HTTP_SERVER_RUNNING=false
    STDIO_SERVER_RUNNING=false
    INSPECTOR_RUNNING=false

    if [ -f "weave-mcp.pid" ] && ps -p "$(cat weave-mcp.pid)" > /dev/null 2>&1; then
        HTTP_SERVER_RUNNING=true
    fi

    if [ -f "weave-mcp-stdio.pid" ] && ps -p "$(cat weave-mcp-stdio.pid)" > /dev/null 2>&1; then
        STDIO_SERVER_RUNNING=true
    fi

    if [ -f "mcp-inspector.pid" ] && ps -p "$(cat mcp-inspector.pid)" > /dev/null 2>&1; then
        INSPECTOR_RUNNING=true
    fi

    if [ "$HTTP_SERVER_RUNNING" = true ] && [ "$STDIO_SERVER_RUNNING" = true ] && [ "$INSPECTOR_RUNNING" = true ]; then
        print_success "🎉 All services are running!"
    elif [ "$HTTP_SERVER_RUNNING" = true ] || [ "$STDIO_SERVER_RUNNING" = true ] || [ "$INSPECTOR_RUNNING" = true ]; then
        print_status "⚠️  Some services are running"
    else
        print_error "❌ No services are running"
    fi
}

# Function to check HTTP server status
check_http_server_status() {
    if [ -f "weave-mcp.pid" ]; then
        PID=$(cat weave-mcp.pid)
        if ps -p "$PID" > /dev/null 2>&1; then
            print_status "HTTP server is running with PID $PID"

            # Get process info
            if command -v ps >/dev/null 2>&1; then
                print_status "Process info:"
                ps -p "$PID" -o pid,ppid,cmd,etime,pcpu,pmem 2>/dev/null || true
            fi

            # Check if server is responding
            HOST=${MCP_SERVER_HOST:-localhost}
            PORT=${MCP_SERVER_PORT:-8030}

            if command -v curl >/dev/null 2>&1; then
                print_status "Checking HTTP server health..."
                if curl -s "http://$HOST:$PORT/health" >/dev/null 2>&1; then
                    print_status "✅ HTTP server is responding to health checks"
                else
                    print_warning "⚠️  HTTP server is running but not responding to health checks"
                fi
            fi

            return 0
        else
            print_warning "PID file exists but process is not running"
            rm -f weave-mcp.pid
            return 1
        fi
    else
        print_status "HTTP server is not running (no PID file found)"
        return 1
    fi
}

# Function to stop the HTTP server
stop_http_server() {
    print_header "Stopping Weave HTTP MCP Server..."

    if [ -f "weave-mcp.pid" ]; then
        PID=$(cat weave-mcp.pid)

        if ps -p "$PID" > /dev/null 2>&1; then
            print_status "Stopping HTTP server with PID $PID..."

            # Try graceful shutdown first
            if kill -TERM "$PID" 2>/dev/null; then
                print_status "Sent SIGTERM signal to HTTP server..."

                # Wait for graceful shutdown
                for _ in {1..30}; do
                    if ! ps -p "$PID" > /dev/null 2>&1; then
                        print_status "HTTP server stopped gracefully"
                        rm -f weave-mcp.pid
                        return 0
                    fi
                    sleep 1
                done

                # Force kill if still running
                print_warning "HTTP server did not stop gracefully, forcing shutdown..."
                if kill -KILL "$PID" 2>/dev/null; then
                    print_status "Sent SIGKILL signal to HTTP server..."
                    sleep 2
                    if ! ps -p "$PID" > /dev/null 2>&1; then
                        print_status "HTTP server stopped forcefully"
                        rm -f weave-mcp.pid
                        return 0
                    else
                        print_error "Failed to stop HTTP server"
                        return 1
                    fi
                else
                    print_error "Failed to send kill signal to HTTP server"
                    return 1
                fi
            else
                print_error "Failed to send TERM signal to HTTP server"
                return 1
            fi
        else
            print_warning "PID file exists but process is not running"
            rm -f weave-mcp.pid
            return 0
        fi
    else
        print_status "No PID file found - HTTP server may not be running"
        return 0
    fi
}

# Function to stop the stdio server
stop_stdio_server() {
    print_header "Stopping Weave stdio MCP Server..."

    if [ -f "weave-mcp-stdio.pid" ]; then
        PID=$(cat weave-mcp-stdio.pid)

        if ps -p "$PID" > /dev/null 2>&1; then
            print_status "Stopping stdio server with PID $PID..."

            # Try graceful shutdown first
            if kill -TERM "$PID" 2>/dev/null; then
                print_status "Sent SIGTERM signal to stdio server..."

                # Wait for graceful shutdown
                for _ in {1..30}; do
                    if ! ps -p "$PID" > /dev/null 2>&1; then
                        print_status "stdio server stopped gracefully"
                        rm -f weave-mcp-stdio.pid
                        return 0
                    fi
                    sleep 1
                done

                # Force kill if still running
                print_warning "stdio server did not stop gracefully, forcing shutdown..."
                if kill -KILL "$PID" 2>/dev/null; then
                    print_status "Sent SIGKILL signal to stdio server..."
                    sleep 2
                    if ! ps -p "$PID" > /dev/null 2>&1; then
                        print_status "stdio server stopped forcefully"
                        rm -f weave-mcp-stdio.pid
                        return 0
                    else
                        print_error "Failed to stop stdio server"
                        return 1
                    fi
                else
                    print_error "Failed to send kill signal to stdio server"
                    return 1
                fi
            else
                print_error "Failed to send TERM signal to stdio server"
                return 1
            fi
        else
            print_warning "PID file exists but process is not running"
            rm -f weave-mcp-stdio.pid
            return 0
        fi
    else
        print_status "No PID file found - stdio server may not be running"
        return 0
    fi
}

# Function to stop the inspector
stop_inspector() {
    print_header "Stopping MCP Inspector..."
    
    # First try to stop using PID file if it exists
    if [ -f "mcp-inspector.pid" ]; then
        PID=$(cat mcp-inspector.pid)
        
        if ps -p "$PID" > /dev/null 2>&1; then
            print_status "Stopping inspector with PID $PID..."
            
            # Try graceful shutdown first
            if kill -TERM "$PID" 2>/dev/null; then
                print_status "Sent SIGTERM signal to inspector..."
                
                # Wait for graceful shutdown
                for _ in {1..30}; do
                    if ! ps -p "$PID" > /dev/null 2>&1; then
                        print_status "Inspector stopped gracefully"
                        rm -f mcp-inspector.pid
                        return 0
                    fi
                    sleep 1
                done
                
                # Force kill if still running
                print_warning "Inspector did not stop gracefully, forcing shutdown..."
                if kill -KILL "$PID" 2>/dev/null; then
                    print_status "Sent SIGKILL signal to inspector..."
                    sleep 2
                    if ! ps -p "$PID" > /dev/null 2>&1; then
                        print_status "Inspector stopped forcefully"
                        rm -f mcp-inspector.pid
                        return 0
                    else
                        print_error "Failed to stop inspector"
                        return 1
                    fi
                else
                    print_error "Failed to send kill signal to inspector"
                    return 1
                fi
            else
                print_error "Failed to send TERM signal to inspector"
                return 1
            fi
        else
            print_warning "Inspector PID file exists but process is not running"
            rm -f mcp-inspector.pid
        fi
    fi
    
    # Also kill any inspector processes by name pattern (in case PID file doesn't exist)
    print_status "Checking for inspector processes by name pattern..."
    
    # Find all inspector-related processes
    INSPECTOR_PIDS=$(pgrep -f "inspector|@modelcontextprotocol/inspector" 2>/dev/null || true)
    
    if [ -n "$INSPECTOR_PIDS" ]; then
        print_status "Found inspector processes: $INSPECTOR_PIDS"
        
        # Try graceful shutdown first
        for PID in $INSPECTOR_PIDS; do
            if ps -p "$PID" > /dev/null 2>&1; then
                print_status "Stopping inspector process with PID $PID..."
                if kill -TERM "$PID" 2>/dev/null; then
                    print_status "Sent SIGTERM signal to process $PID..."
                fi
            fi
        done
        
        # Wait for graceful shutdown
        sleep 3
        
        # Check if any processes are still running and force kill them
        REMAINING_PIDS=$(pgrep -f "inspector|@modelcontextprotocol/inspector" 2>/dev/null || true)
        if [ -n "$REMAINING_PIDS" ]; then
            print_warning "Some inspector processes did not stop gracefully, forcing shutdown..."
            for PID in $REMAINING_PIDS; do
                if ps -p "$PID" > /dev/null 2>&1; then
                    print_status "Force killing inspector process with PID $PID..."
                    kill -KILL "$PID" 2>/dev/null || true
                fi
            done
            sleep 2
        fi
        
        # Final check
        FINAL_PIDS=$(pgrep -f "inspector|@modelcontextprotocol/inspector" 2>/dev/null || true)
        if [ -n "$FINAL_PIDS" ]; then
            print_error "Failed to stop some inspector processes: $FINAL_PIDS"
            return 1
        else
            print_success "All inspector processes stopped successfully"
            return 0
        fi
    else
        print_status "No inspector processes found running"
        return 0
    fi
}

# Function to stop all services
stop_all() {
    print_header "Stopping all services..."
    echo ""

    print_status "Stopping MCP Inspector..."
    stop_inspector
    echo ""

    print_status "Stopping HTTP MCP Server..."
    stop_http_server
    echo ""

    print_status "Stopping stdio MCP Server..."
    stop_stdio_server
    echo ""

    print_status "All services stopped"
}

# Function to restart all services
restart_all() {
    print_header "Restarting all services..."
    
    # Stop all services
    stop_all
    
    # Wait a moment
    sleep 2
    
    # Start all services
    print_status "Starting all services..."
    if ./start.sh both; then
        print_status "All services restarted successfully"
    else
        print_error "Failed to restart all services"
        return 1
    fi
}

# Function to restart the HTTP server
restart_http_server() {
    print_header "Restarting Weave HTTP MCP Server..."

    # Stop the server
    if stop_http_server; then
        print_status "HTTP server stopped successfully"
    else
        print_warning "HTTP server stop had issues, but continuing with restart"
    fi

    # Wait a moment
    sleep 2

    # Start the server
    print_status "Starting HTTP server..."
    if ./start.sh http --daemon; then
        print_status "HTTP server restarted successfully"
    else
        print_error "Failed to restart HTTP server"
        return 1
    fi
}

# Function to restart the stdio server
restart_stdio_server() {
    print_header "Restarting Weave stdio MCP Server..."

    # Stop the server
    if stop_stdio_server; then
        print_status "stdio server stopped successfully"
    else
        print_warning "stdio server stop had issues, but continuing with restart"
    fi

    # Wait a moment
    sleep 2

    # Start the server
    print_status "Starting stdio server..."
    if ./start.sh stdio --daemon; then
        print_status "stdio server restarted successfully"
    else
        print_error "Failed to restart stdio server"
        return 1
    fi
}

# Main script logic
main() {
    local command="${1:-stop}"
    local target="${2:-http}"

    case "$command" in
        "stop")
            case "$target" in
                "http")
                    stop_http_server
                    ;;
                "stdio")
                    stop_stdio_server
                    ;;
                "inspector")
                    stop_inspector
                    ;;
                "all")
                    stop_all
                    ;;
                *)
                    print_error "Unknown target for stop: $target"
                    print_help
                    exit 1
                    ;;
            esac
            ;;
        "status")
            case "$target" in
                "http")
                    check_http_server_status
                    ;;
                "stdio")
                    check_stdio_server_status
                    ;;
                "inspector")
                    check_inspector_status
                    ;;
                "all")
                    check_all_status
                    ;;
                *)
                    print_error "Unknown target for status: $target"
                    print_help
                    exit 1
                    ;;
            esac
            ;;
        "restart")
            case "$target" in
                "http")
                    restart_http_server
                    ;;
                "stdio")
                    restart_stdio_server
                    ;;
                "inspector")
                    stop_inspector
                    sleep 2
                    ./start.sh inspector
                    ;;
                "all")
                    restart_all
                    ;;
                *)
                    print_error "Unknown target for restart: $target"
                    print_help
                    exit 1
                    ;;
            esac
            ;;
        "--help"|"-h")
            print_help
            ;;
        *)
            print_error "Unknown command: $command"
            print_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"