// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// loadEnvFile loads environment variables from .env file
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') ||
			(value[0] == '\'' && value[len(value)-1] == '\'')) {
			value = value[1 : len(value)-1]
		}

		os.Setenv(key, value)
	}

	return scanner.Err()
}

// buildWeaveMCP builds the weave-mcp binary
func buildWeaveMCP(t *testing.T, ctx context.Context) {
	buildScript := filepath.Join("..", "build.sh")
	if !fileExists(buildScript) {
		// Try from different directory levels
		buildScript = "./build.sh"
		if !fileExists(buildScript) {
			buildScript = "../../build.sh"
		}
	}

	cmd := exec.CommandContext(ctx, "bash", buildScript)
	cmd.Dir = filepath.Dir(buildScript)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Build output: %s", output)
	}
	require.NoError(t, err, "Failed to build weave-mcp")
	t.Logf("Built weave-mcp successfully")
}

// waitForMCPServer waits for MCP server to be ready
func waitForMCPServer(ctx context.Context) error {
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for MCP server")
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8030/health", nil)
			if err != nil {
				continue
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				continue
			}
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
	}
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
