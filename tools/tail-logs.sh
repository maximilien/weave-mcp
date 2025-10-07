#!/bin/bash

# Weave MCP Log Tailing Script
# Usage: ./tools/tail-logs.sh [all|mcp|status|recent]

# Colors for different services
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to check if a service is running
check_service() {
    local port=$1
    local service_name=$2
    if lsof -Pi :$port -sTCP:LISTEN -t >/dev/null 2>&1; then
        echo -e "${GREEN}✅ $service_name is running on port $port${NC}"
        return 0
    else
        echo -e "${RED}❌ $service_name is not running on port $port${NC}"
        return 1
    fi
}

# Function to get the MCP server PID
get_mcp_pid() {
    pgrep -f "weave-mcp" | head -1
}

# Function to show all service status
show_status() {
    echo -e "${BLUE}=== Weave MCP Service Status ===${NC}"
    echo ""
    check_service 8030 "Weave MCP Server"
    
    local mcp_pid=$(get_mcp_pid)
    if [ ! -z "$mcp_pid" ]; then
        echo -e "${GREEN}✅ Weave MCP Server is running (PID: $mcp_pid)${NC}"
    else
        echo -e "${RED}❌ Weave MCP Server is not running${NC}"
    fi
    echo ""
    
    # Check if logs directory exists
    if [ -d "logs" ]; then
        echo -e "${BLUE}📁 Log files:${NC}"
        ls -la logs/ 2>/dev/null | grep -E "\.(log|txt)$" || echo "No log files found"
    else
        echo -e "${YELLOW}⚠️ Logs directory not found${NC}"
    fi
    echo ""
}

# Function to show recent logs
show_recent_logs() {
    echo -e "${BLUE}=== Recent Weave MCP Logs (Last 50 lines) ===${NC}"
    echo ""
    
    if [ -f "logs/weave-mcp.log" ]; then
        echo -e "${YELLOW}Recent MCP server logs:${NC}"
        tail -50 logs/weave-mcp.log | while read line; do
            echo "[MCP] $line"
        done
    else
        echo -e "${RED}❌ MCP log file not found (logs/weave-mcp.log)${NC}"
    fi
    
    echo ""
    echo -e "${YELLOW}Recent system logs containing 'weave-mcp':${NC}"
    log show --predicate 'eventMessage CONTAINS "weave-mcp"' --last 5m 2>/dev/null | tail -20 | while read line; do
        echo "[SYS] $line"
    done
}

# Function to monitor system logs for MCP process
monitor_system_logs() {
    local service_name=$1
    local service_tag=$2
    local pid=$3
    
    echo -e "${YELLOW}Monitoring system logs for $service_name (PID: $pid)...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop monitoring${NC}"
    echo ""
    
    # Monitor system logs for the specific process
    log stream --predicate "process == '$pid'" 2>/dev/null | while read line; do
        echo "[$service_tag] $line"
    done
}

# Function to tail MCP logs
tail_mcp_logs() {
    echo -e "${BLUE}📡 Tailing Weave MCP logs (port 8030)...${NC}"
    if check_service 8030 "Weave MCP Server"; then
        if [ -f "logs/weave-mcp.log" ]; then
            echo -e "${BLUE}Following MCP log file...${NC}"
            echo -e "${YELLOW}Press Ctrl+C to stop monitoring${NC}"
            echo ""
            tail -f logs/weave-mcp.log
        else
            echo -e "${RED}❌ MCP log file not found (logs/weave-mcp.log)${NC}"
            echo -e "${YELLOW}MCP server may not be running or log file not created${NC}"
        fi
    fi
}

# Function to tail all logs
tail_all_logs() {
    echo -e "${BLUE}📡 Tailing all Weave MCP logs...${NC}"
    echo -e "${YELLOW}Press Ctrl+C to stop monitoring${NC}"
    echo ""
    
    # Use tail -f with MCP log file
    if [ -f "logs/weave-mcp.log" ]; then
        tail -f logs/weave-mcp.log
    else
        echo -e "${YELLOW}⚠️ MCP log file not found. Server may not be running.${NC}"
        echo -e "${YELLOW}Available log files:${NC}"
        ls -la logs/ 2>/dev/null || echo "No logs directory found"
    fi
}

# Function to show help
show_help() {
    echo -e "${BLUE}Weave MCP Log Tailing Script${NC}"
    echo ""
    echo "Usage: $0 [all|mcp|status|recent]"
    echo ""
    echo "Commands:"
    echo -e "  ${GREEN}all${NC}        - Tail logs from all running services"
    echo -e "  ${GREEN}mcp${NC}        - Tail MCP server logs (port 8030)"
    echo -e "  ${GREEN}status${NC}     - Show status of all services"
    echo -e "  ${GREEN}recent${NC}     - Show recent logs"
    echo ""
    echo "Examples:"
    echo "  $0 all           # Tail all service logs"
    echo "  $0 mcp           # Tail only MCP server logs"
    echo "  $0 status        # Show service status"
    echo "  $0 recent        # Show recent logs"
    echo ""
    echo "Note: This script tails real-time logs from Weave MCP services."
    echo "Logs are written to ./logs/weave-mcp.log"
}

# Main script logic
case "${1:-all}" in
    "all")
        tail_all_logs
        ;;
    "mcp")
        tail_mcp_logs
        ;;
    "status")
        show_status
        ;;
    "recent")
        show_recent_logs
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        echo -e "${RED}Unknown command: $1${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac