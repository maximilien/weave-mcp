# Troubleshooting Guide

This guide covers common issues when setting up and running the Weave MCP
Server and MCP Inspector.

## Table of Contents

- [Prerequisites Issues](#prerequisites-issues)
- [Setup Issues](#setup-issues)
- [MCP Server Issues](#mcp-server-issues)
- [MCP Inspector Issues](#mcp-inspector-issues)
- [Configuration Issues](#configuration-issues)
- [Network and Port Issues](#network-and-port-issues)
- [Logging and Debugging](#logging-and-debugging)
- [Performance Issues](#performance-issues)
- [Getting Help](#getting-help)

## Prerequisites Issues

### Go Version Issues

**Problem**: `go: requires go >= 1.24.0`

**Solution**:
```bash
# Check current Go version
go version

# Install Go 1.24+ if needed
# Using Homebrew (macOS)
brew install go@1.24

# Or download from https://golang.org/dl/
```

### Node.js Version Issues

**Problem**: MCP Inspector requires Node.js 22.7.5+

**Solution**:
```bash
# Check current Node.js version
node --version

# Install Node.js 22.7.5+ if needed
# Using Homebrew (macOS)
brew install node@22

# Or download from https://nodejs.org/
```

### Missing Dependencies

**Problem**: `command not found` errors

**Solution**:
```bash
# Install required tools
# Go tools
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Python tools
pip install yamllint

# Node.js tools
npm install -g markdownlint-cli
```

## Setup Issues

### Setup Script Fails

**Problem**: `./setup.sh` fails with errors

**Solution**:
```bash
# Make scripts executable
chmod +x *.sh

# Run setup with verbose output
bash -x ./setup.sh

# Check for specific error messages
./setup.sh 2>&1 | tee setup.log
```

### Permission Denied Errors

**Problem**: Permission denied when running scripts

**Solution**:
```bash
# Fix script permissions
chmod +x *.sh
chmod +x tools/*.sh

# Run with proper permissions
sudo ./setup.sh  # Only if necessary
```

### Build Failures

**Problem**: `./build.sh` fails to compile

**Solution**:
```bash
# Clean and rebuild
rm -rf bin/
go clean -cache
go mod tidy
./build.sh

# Check Go environment
go env
go version
```

## MCP Server Issues

### Server Won't Start

**Problem**: MCP server fails to start

**Solution**:
```bash
# Check if port is already in use
lsof -i :8030

# Kill existing processes
./stop.sh all

# Check configuration
cat config.yaml
cat .env

# Start with verbose logging
./start.sh http --daemon
./tools/tail-logs.sh
```

### Configuration Errors

**Problem**: Invalid configuration errors

**Solution**:
```bash
# Validate configuration
./bin/weave-mcp --config-check

# Check YAML syntax
yamllint config.yaml

# Verify environment variables
env | grep -E "(WEAVIATE|MILVUS|MCP)"
```

### Database Connection Issues

**Problem**: Cannot connect to vector database

**Solution**:
```bash
# Test Weaviate connection
curl -H "Authorization: Bearer $WEAVIATE_API_KEY" \
  "$WEAVIATE_URL/v1/meta"

# Test Milvus connection
telnet $MILVUS_HOST $MILVUS_PORT

# Use mock database for testing
export VECTOR_DB_TYPE=mock
./start.sh http
```

### stdio Server Issues

**Problem**: stdio server not working with MCP clients

**Solution**:
```bash
# Test stdio server directly
echo '{"jsonrpc": "2.0", "id": 1, "method": "initialize", "params": {}}' | \
  ./bin/weave-mcp-stdio

# Check MCP client configuration
cat tools/mcp-inspector-config.json

# Verify environment variables are loaded
./bin/weave-mcp-stdio -env .env --help
```

## MCP Inspector Issues

### Inspector Won't Start

**Problem**: MCP Inspector fails to start

**Solution**:
```bash
# Check Node.js version
node --version  # Should be 22.7.5+

# Install inspector manually
npm install -g @modelcontextprotocol/inspector

# Start inspector manually
npx @modelcontextprotocol/inspector

# Check for port conflicts
lsof -i :6274
lsof -i :6277
```

### Inspector Can't Connect to Server

**Problem**: Inspector shows connection errors

**Solution**:
```bash
# Verify MCP server is running
curl http://localhost:8030/health

# Check inspector configuration
cat tools/mcp-inspector-config.json

# Test server connection
curl http://localhost:8030/mcp/tools/list
```

### Browser Access Issues

**Problem**: Cannot access inspector in browser

**Solution**:
```bash
# Check if inspector is running
ps aux | grep inspector

# Check port accessibility
curl http://localhost:6274

# Try different browser or incognito mode
# Clear browser cache and cookies
```

## Configuration Issues

### Environment Variables Not Loading

**Problem**: Environment variables not being read

**Solution**:
```bash
# Check .env file exists and is readable
ls -la .env
cat .env

# Verify variable format (no spaces around =)
# Correct: WEAVIATE_URL=https://example.com
# Wrong:   WEAVIATE_URL = https://example.com

# Test environment loading
source .env
echo $WEAVIATE_URL
```

### YAML Configuration Errors

**Problem**: YAML parsing errors

**Solution**:
```bash
# Validate YAML syntax
yamllint config.yaml

# Check for common YAML issues:
# - Proper indentation (2 spaces, not tabs)
# - Correct quotes around strings
# - No trailing commas
# - Proper list formatting

# Test configuration loading
./bin/weave-mcp --config-check
```

### Database Configuration Issues

**Problem**: Database connection configuration errors

**Solution**:
```bash
# Weaviate Cloud configuration
export VECTOR_DB_TYPE=weaviate-cloud
export WEAVIATE_URL=https://your-cluster.weaviate.network
export WEAVIATE_API_KEY=your-api-key
export OPENAI_API_KEY=your-openai-key

# Milvus configuration
export VECTOR_DB_TYPE=milvus
export MILVUS_HOST=localhost
export MILVUS_PORT=19530

# Mock database (for testing)
export VECTOR_DB_TYPE=mock
```

## Network and Port Issues

### Port Already in Use

**Problem**: Port 8030 or 6274 already in use

**Solution**:
```bash
# Find process using port
lsof -i :8030
lsof -i :6274

# Kill the process
kill -9 <PID>

# Or use different ports
export MCP_SERVER_PORT=8031
./start.sh http
```

### Firewall Issues

**Problem**: Cannot access server from external machines

**Solution**:
```bash
# Check firewall status
sudo ufw status  # Ubuntu/Debian
sudo firewall-cmd --list-all  # CentOS/RHEL

# Allow port through firewall
sudo ufw allow 8030
sudo firewall-cmd --permanent --add-port=8030/tcp
sudo firewall-cmd --reload
```

### Localhost vs External Access

**Problem**: Server only accessible via localhost

**Solution**:
```bash
# Update configuration to bind to all interfaces
# In config.yaml or environment:
MCP_SERVER_HOST=0.0.0.0

# Or start with specific host
./start.sh http --host 0.0.0.0
```

## Logging and Debugging

### Enable Verbose Logging

**Solution**:
```bash
# Set debug log level
export LOG_LEVEL=debug

# Start with verbose output
./start.sh http --verbose

# Monitor logs in real-time
./tools/tail-logs.sh
```

### Log File Issues

**Problem**: Log files not being created or empty

**Solution**:
```bash
# Check log directory permissions
ls -la logs/
mkdir -p logs
chmod 755 logs

# Check disk space
df -h

# Verify logging configuration
grep -i log config.yaml
```

### Debug MCP Protocol

**Solution**:
```bash
# Enable MCP debug logging
export MCP_DEBUG=true

# Test MCP protocol manually
echo '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}' | \
  curl -X POST http://localhost:8030/mcp/tools/call \
  -H "Content-Type: application/json" \
  -d @-
```

## Performance Issues

### Slow Response Times

**Problem**: MCP server responds slowly

**Solution**:
```bash
# Check system resources
top
htop
free -h
df -h

# Monitor network latency
ping your-database-host

# Check database performance
# For Weaviate: Check cluster status in console
# For Milvus: Check server logs and metrics
```

### Memory Issues

**Problem**: High memory usage or out of memory errors

**Solution**:
```bash
# Monitor memory usage
ps aux --sort=-%mem | head

# Check for memory leaks
valgrind --tool=memcheck ./bin/weave-mcp

# Adjust Go garbage collection
export GOGC=50  # More aggressive GC
```

### Database Performance

**Problem**: Database queries are slow

**Solution**:
```bash
# Check database connection pool settings
# In config.yaml:
database:
  max_connections: 10
  connection_timeout: 30s

# Monitor database metrics
# Weaviate: Check cluster metrics in console
# Milvus: Check server metrics and logs
```

## Getting Help

### Collect Debug Information

Before asking for help, collect this information:

```bash
# System information
uname -a
go version
node --version
python3 --version

# Project information
git log --oneline -5
git status

# Configuration
cat config.yaml
cat .env

# Logs
./tools/tail-logs.sh recent

# Process information
ps aux | grep -E "(weave-mcp|inspector)"
lsof -i :8030
lsof -i :6274
```

### Common Commands for Debugging

```bash
# Full system check
./lint.sh && ./test.sh && ./build.sh

# Check all services
./stop.sh status

# Restart everything
./stop.sh all
./start.sh both

# Monitor everything
./tools/tail-logs.sh all
```

### Where to Get Help

1. **Check the logs first**: `./tools/tail-logs.sh`
2. **Run the linter**: `./lint.sh`
3. **Run tests**: `./test.sh`
4. **Check GitHub Issues**: [Issues](https://github.com/maximilien/weave-mcp/issues)
5. **Create a new issue** with debug information

### Creating a Good Issue Report

Include this information in your issue:

```markdown
## Environment
- OS: [e.g., macOS 14.0, Ubuntu 22.04]
- Go version: [e.g., 1.24.1]
- Node.js version: [e.g., 22.7.5]
- Project version: [e.g., v0.1.0]

## Problem
Brief description of the issue

## Steps to Reproduce
1. Run `./setup.sh`
2. Run `./start.sh http`
3. See error: [paste error message]

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Debug Information
[Paste output from debug commands above]
```

## Quick Reference

### Essential Commands

```bash
# Setup
./setup.sh

# Build
./build.sh

# Start services
./start.sh http          # HTTP server only
./start.sh stdio         # stdio server only
./start.sh inspector     # Inspector only
./start.sh both          # HTTP server + Inspector

# Stop services
./stop.sh http
./stop.sh stdio
./stop.sh inspector
./stop.sh all

# Monitor
./tools/tail-logs.sh

# Test
./test.sh

# Lint
./lint.sh
```

### Important Files

- `config.yaml` - Main configuration
- `.env` - Environment variables
- `logs/` - Log files directory
- `bin/` - Compiled binaries
- `tools/mcp-inspector-config.json` - Inspector configuration

### Default Ports

- **MCP HTTP Server**: 8030
- **MCP Inspector Web**: 6274
- **MCP Inspector Proxy**: 6277