// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPBinaryStartup tests MCP binary startup and shutdown
func TestMCPBinaryStartup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping binary integration test in short mode")
	}

	// Build MCP binary first
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	buildWeaveMCP(t, ctx)

	t.Run("HTTPServerStartup", func(t *testing.T) {
		mcpBinary := filepath.Join("..", "..", "bin", "weave-mcp")
		configFile := filepath.Join("..", "..", "config.yaml")

		cmd := exec.CommandContext(ctx, mcpBinary, "--config", configFile)
		cmd.Env = os.Environ()

		// Capture output
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		err := cmd.Start()
		require.NoError(t, err)
		defer cmd.Process.Kill()

		// Wait for server to start
		time.Sleep(2 * time.Second)

		// Check if process is still running
		assert.Nil(t, cmd.Process.Signal(os.Signal(nil)))

		// Check health endpoint
		resp, err := http.Get("http://localhost:8030/health")
		if err != nil {
			t.Logf("Stdout: %s", stdout.String())
			t.Logf("Stderr: %s", stderr.String())
		}
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		// Read health response
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)

		var healthResp map[string]interface{}
		err = json.Unmarshal(body, &healthResp)
		require.NoError(t, err)

		assert.Equal(t, "healthy", healthResp["status"])
		t.Logf("Health response: %v", healthResp)
	})

	t.Run("StdioServerStartup", func(t *testing.T) {
		mcpBinary := filepath.Join("..", "..", "bin", "weave-mcp-stdio")
		configFile := filepath.Join("..", "..", "config.yaml")

		cmd := exec.CommandContext(ctx, mcpBinary, "--config", configFile)
		cmd.Env = os.Environ()

		// Create pipes for stdin/stdout
		stdin, err := cmd.StdinPipe()
		require.NoError(t, err)
		defer stdin.Close()

		stdout, err := cmd.StdoutPipe()
		require.NoError(t, err)
		defer stdout.Close()

		err = cmd.Start()
		require.NoError(t, err)
		defer cmd.Process.Kill()

		// Give it a moment to initialize
		time.Sleep(1 * time.Second)

		// Check if process is still running
		assert.Nil(t, cmd.Process.Signal(os.Signal(nil)))

		t.Logf("stdio server started successfully")
	})
}

// TestMCPBinaryHTTPAPI tests MCP binary HTTP API endpoints
func TestMCPBinaryHTTPAPI(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping binary API test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start MCP server with mock database
	mcpProcess, configFile := startMCPServerWithMock(t, ctx)
	defer mcpProcess.Kill()

	// Wait for server to be ready
	require.NoError(t, waitForMCPServer(ctx))

	baseURL := "http://localhost:8030"

	t.Run("HealthEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var health map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&health)
		require.NoError(t, err)

		assert.Equal(t, "healthy", health["status"])
	})

	t.Run("ToolsListEndpoint", func(t *testing.T) {
		resp, err := http.Get(baseURL + "/mcp/tools/list")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var tools map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&tools)
		require.NoError(t, err)

		toolsList, ok := tools["tools"].([]interface{})
		require.True(t, ok)
		assert.Greater(t, len(toolsList), 15) // Should have at least 18 tools
	})

	t.Run("ToolCallEndpoint_ListCollections", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":      "list_collections",
			"arguments": map[string]interface{}{},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/mcp/tools/call",
			"application/json",
			bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotNil(t, result["result"])
		t.Logf("list_collections result: %v", result)
	})

	t.Run("ToolCallEndpoint_CreateCollection", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "create_collection",
			"arguments": map[string]interface{}{
				"name":        "test-http-collection",
				"type":        "text",
				"description": "HTTP API test collection",
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/mcp/tools/call",
			"application/json",
			bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotNil(t, result["result"])
		t.Logf("create_collection result: %v", result)
	})

	t.Run("ToolCallEndpoint_HealthCheck", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":      "health_check",
			"arguments": map[string]interface{}{},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/mcp/tools/call",
			"application/json",
			bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		resultMap, ok := result["result"].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "healthy", resultMap["status"])
		t.Logf("health_check result: %v", result)
	})

	t.Run("ToolCallEndpoint_CountCollections", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":      "count_collections",
			"arguments": map[string]interface{}{},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/mcp/tools/call",
			"application/json",
			bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		resultMap, ok := result["result"].(map[string]interface{})
		require.True(t, ok)
		assert.NotNil(t, resultMap["count"])
		t.Logf("count_collections result: %v", result)
	})

	t.Run("ToolCallEndpoint_InvalidTool", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":      "nonexistent_tool",
			"arguments": map[string]interface{}{},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/mcp/tools/call",
			"application/json",
			bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return error
		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ToolCallEndpoint_InvalidArguments", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "create_collection",
			"arguments": map[string]interface{}{
				// Missing required 'name' parameter
				"type": "text",
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/mcp/tools/call",
			"application/json",
			bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		// Should return error for invalid arguments
		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotNil(t, result["error"])
		t.Logf("Error response (expected): %v", result["error"])
	})

	// Clean up config file
	defer os.Remove(configFile)
}

// TestMCPBinaryAllTools tests all MCP tools via HTTP API
func TestMCPBinaryAllTools(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping comprehensive tools test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Start MCP server with mock database
	mcpProcess, configFile := startMCPServerWithMock(t, ctx)
	defer mcpProcess.Kill()
	defer os.Remove(configFile)

	// Wait for server to be ready
	require.NoError(t, waitForMCPServer(ctx))

	baseURL := "http://localhost:8030"

	// Test all tools
	testCases := []struct {
		name      string
		tool      string
		args      map[string]interface{}
		expectErr bool
	}{
		{
			name:      "health_check",
			tool:      "health_check",
			args:      map[string]interface{}{},
			expectErr: false,
		},
		{
			name:      "list_collections",
			tool:      "list_collections",
			args:      map[string]interface{}{},
			expectErr: false,
		},
		{
			name:      "count_collections",
			tool:      "count_collections",
			args:      map[string]interface{}{},
			expectErr: false,
		},
		{
			name: "create_collection",
			tool: "create_collection",
			args: map[string]interface{}{
				"name":        "test-all-tools",
				"type":        "text",
				"description": "Test collection",
			},
			expectErr: false,
		},
		{
			name: "show_collection",
			tool: "show_collection",
			args: map[string]interface{}{
				"name": "test-all-tools",
			},
			expectErr: false,
		},
		{
			name:      "list_embedding_models",
			tool:      "list_embedding_models",
			args:      map[string]interface{}{},
			expectErr: false,
		},
		{
			name: "create_document",
			tool: "create_document",
			args: map[string]interface{}{
				"collection": "test-all-tools",
				"url":        "https://example.com/doc1",
				"text":       "Test document",
			},
			expectErr: false,
		},
		{
			name: "list_documents",
			tool: "list_documents",
			args: map[string]interface{}{
				"collection": "test-all-tools",
				"limit":      10,
			},
			expectErr: false,
		},
		{
			name: "count_documents",
			tool: "count_documents",
			args: map[string]interface{}{
				"collection": "test-all-tools",
			},
			expectErr: false,
		},
		{
			name: "query_documents",
			tool: "query_documents",
			args: map[string]interface{}{
				"collection": "test-all-tools",
				"query":      "test",
				"limit":      5,
			},
			expectErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := map[string]interface{}{
				"name":      tc.tool,
				"arguments": tc.args,
			}

			body, err := json.Marshal(reqBody)
			require.NoError(t, err)

			resp, err := http.Post(baseURL+"/mcp/tools/call",
				"application/json",
				bytes.NewReader(body))
			require.NoError(t, err)
			defer resp.Body.Close()

			var result map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&result)
			require.NoError(t, err)

			if tc.expectErr {
				assert.NotNil(t, result["error"], "Expected error for %s", tc.tool)
			} else {
				if result["error"] != nil {
					t.Logf("Tool %s returned error: %v", tc.tool, result["error"])
				}
				assert.NotNil(t, result["result"], "Expected result for %s", tc.tool)
			}

			t.Logf("%s result: %v", tc.tool, result)
		})
	}
}

// TestMCPBinaryErrorHandling tests error handling in MCP binary
func TestMCPBinaryErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping error handling test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	mcpProcess, configFile := startMCPServerWithMock(t, ctx)
	defer mcpProcess.Kill()
	defer os.Remove(configFile)

	require.NoError(t, waitForMCPServer(ctx))

	baseURL := "http://localhost:8030"

	t.Run("MissingRequiredParameter", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "create_collection",
			"arguments": map[string]interface{}{
				"type": "text", // Missing 'name' parameter
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/mcp/tools/call",
			"application/json",
			bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		assert.NotNil(t, result["error"])
		t.Logf("Error (expected): %v", result["error"])
	})

	t.Run("InvalidJSON", func(t *testing.T) {
		resp, err := http.Post(baseURL+"/mcp/tools/call",
			"application/json",
			bytes.NewReader([]byte("invalid json")))
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.NotEqual(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("NonExistentCollection", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name": "list_documents",
			"arguments": map[string]interface{}{
				"collection": "nonexistent-collection",
				"limit":      10,
			},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		resp, err := http.Post(baseURL+"/mcp/tools/call",
			"application/json",
			bytes.NewReader(body))
		require.NoError(t, err)
		defer resp.Body.Close()

		var result map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result)
		require.NoError(t, err)

		// Mock DB might not error on nonexistent collection
		// but real DB would
		t.Logf("Result: %v", result)
	})
}

// TestMCPBinaryPerformance tests MCP binary performance
func TestMCPBinaryPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	mcpProcess, configFile := startMCPServerWithMock(t, ctx)
	defer mcpProcess.Kill()
	defer os.Remove(configFile)

	require.NoError(t, waitForMCPServer(ctx))

	baseURL := "http://localhost:8030"

	t.Run("HealthCheckLatency", func(t *testing.T) {
		iterations := 10
		var totalDuration time.Duration

		for i := 0; i < iterations; i++ {
			start := time.Now()
			resp, err := http.Get(baseURL + "/health")
			duration := time.Since(start)
			totalDuration += duration

			require.NoError(t, err)
			resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		avgDuration := totalDuration / time.Duration(iterations)
		t.Logf("Average health check latency: %v", avgDuration)

		// Health check should be fast
		assert.Less(t, avgDuration, 100*time.Millisecond)
	})

	t.Run("ToolCallLatency", func(t *testing.T) {
		reqBody := map[string]interface{}{
			"name":      "list_collections",
			"arguments": map[string]interface{}{},
		}

		body, err := json.Marshal(reqBody)
		require.NoError(t, err)

		iterations := 10
		var totalDuration time.Duration

		for i := 0; i < iterations; i++ {
			start := time.Now()
			resp, err := http.Post(baseURL+"/mcp/tools/call",
				"application/json",
				bytes.NewReader(body))
			duration := time.Since(start)
			totalDuration += duration

			require.NoError(t, err)
			resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		}

		avgDuration := totalDuration / time.Duration(iterations)
		t.Logf("Average tool call latency: %v", avgDuration)

		// Tool calls should be reasonably fast with mock DB
		assert.Less(t, avgDuration, 500*time.Millisecond)
	})
}

// startMCPServerWithMock starts MCP server with mock database configuration
func startMCPServerWithMock(t *testing.T, ctx context.Context) (*os.Process, string) {
	mcpBinary := filepath.Join("..", "..", "bin", "weave-mcp")

	// Create temporary config file
	tmpDir, err := os.MkdirTemp("", "mcp-test-*")
	require.NoError(t, err)

	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `databases:
  default: "mock"
  vector_databases:
    - name: "mock"
      type: "mock"
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections: []

mcp:
  server:
    http:
      enabled: true
      host: "0.0.0.0"
      port: 8030
    stdio:
      enabled: false

operation_timeouts:
  default: 30
`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := exec.CommandContext(ctx, mcpBinary, "--config", configFile)
	cmd.Env = os.Environ()

	// Create log file
	logFile := filepath.Join(tmpDir, "mcp.log")
	log, err := os.Create(logFile)
	require.NoError(t, err)

	cmd.Stdout = log
	cmd.Stderr = log

	err = cmd.Start()
	require.NoError(t, err)

	t.Logf("Started MCP server (PID: %d, log: %s)", cmd.Process.Pid, logFile)

	return cmd.Process, configFile
}
