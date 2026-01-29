// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	weaveCLIVersion = "v0.9.12"
	weaveCLIRepo    = "maximilien/weave-cli"
)

// TestE2EWeaveWithMCP tests end-to-end weave-cli + weave-mcp integration
// This test:
// 1. Downloads weave-cli binary from GitHub release
// 2. Sets up local Weaviate using Docker
// 3. Builds weave-mcp binary
// 4. Tests weave CLI commands work with MCP
// 5. Tests AI features (suggest_schema, suggest_chunking)
func TestE2EWeaveWithMCP(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Load environment variables from both projects
	loadEnvFiles(t)

	// Check for required environment variables
	if os.Getenv("OPENAI_API_KEY") == "" {
		t.Skip("Skipping E2E test - OPENAI_API_KEY not set")
	}

	// Create temporary directory for test artifacts
	tmpDir, err := os.MkdirTemp("", "weave-e2e-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Logf("Test directory: %s", tmpDir)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Step 1: Download weave-cli binary
	t.Run("DownloadWeaveCLI", func(t *testing.T) {
		weaveBinary := downloadWeaveCLI(t, ctx, tmpDir)
		require.FileExists(t, weaveBinary)
		t.Logf("Downloaded weave-cli to: %s", weaveBinary)
	})

	// Step 2: Build weave-mcp binary
	t.Run("BuildWeaveMCP", func(t *testing.T) {
		buildWeaveMCP(t, ctx)
		mcpBinary := filepath.Join("..", "..", "bin", "weave-mcp")
		require.FileExists(t, mcpBinary)
		t.Logf("Built weave-mcp binary: %s", mcpBinary)
	})

	// Step 3: Set up local Weaviate
	t.Run("SetupWeaviate", func(t *testing.T) {
		setupWeaviate(t, ctx)
		t.Logf("Weaviate is running at http://localhost:8080")
	})

	// Step 4: Start weave-mcp server
	var mcpProcess *os.Process
	t.Run("StartMCPServer", func(t *testing.T) {
		mcpProcess = startMCPServer(t, ctx, tmpDir)
		require.NotNil(t, mcpProcess)
		t.Logf("Started weave-mcp server (PID: %d)", mcpProcess.Pid)

		// Wait for MCP server to be ready
		require.NoError(t, waitForMCPServer(ctx))
	})
	defer func() {
		if mcpProcess != nil {
			mcpProcess.Kill()
		}
	}()

	// Step 5: Test weave CLI basic commands
	t.Run("WeaveBasicCommands", func(t *testing.T) {
		weaveBinary := filepath.Join(tmpDir, "weave")

		// Test version
		output := runWeaveCommand(t, ctx, weaveBinary, "version")
		assert.Contains(t, output, weaveCLIVersion)
		t.Logf("weave version: %s", strings.TrimSpace(output))

		// Test config show
		output = runWeaveCommand(t, ctx, weaveBinary, "config", "show")
		assert.NotEmpty(t, output)
		t.Logf("weave config show succeeded")
	})

	// Step 6: Test weave config update --weave-mcp
	t.Run("WeaveConfigUpdateMCP", func(t *testing.T) {
		weaveBinary := filepath.Join(tmpDir, "weave")

		// Test config update with MCP
		output := runWeaveCommand(t, ctx, weaveBinary, "config", "update", "--weave-mcp")
		t.Logf("weave config update --weave-mcp output: %s", output)

		// Verify MCP configuration was updated
		output = runWeaveCommand(t, ctx, weaveBinary, "config", "show")
		assert.Contains(t, output, "weave-mcp")
	})

	// Step 7: Test weave collection commands via MCP
	t.Run("WeaveCollectionCommands", func(t *testing.T) {
		weaveBinary := filepath.Join(tmpDir, "weave")

		// List collections
		output := runWeaveCommand(t, ctx, weaveBinary, "collections", "list")
		t.Logf("Collections: %s", output)

		// Create test collection
		testCollection := fmt.Sprintf("e2e-test-%d", time.Now().Unix())
		output = runWeaveCommand(t, ctx, weaveBinary, "collections", "create",
			"--name", testCollection,
			"--type", "text",
			"--description", "E2E test collection")
		t.Logf("Created collection: %s", output)

		// List collections again
		output = runWeaveCommand(t, ctx, weaveBinary, "collections", "list")
		assert.Contains(t, output, testCollection)

		// Show collection details
		output = runWeaveCommand(t, ctx, weaveBinary, "collections", "show", "--name", testCollection)
		assert.Contains(t, output, testCollection)

		// Clean up - delete collection
		defer func() {
			runWeaveCommand(t, ctx, weaveBinary, "collections", "delete", "--name", testCollection)
		}()
	})

	// Step 8: Test AI features (suggest_schema, suggest_chunking)
	t.Run("WeaveAIFeatures", func(t *testing.T) {
		weaveBinary := filepath.Join(tmpDir, "weave")

		// Test suggest_schema
		testDoc := "This is a test document about machine learning and artificial intelligence. It discusses neural networks, deep learning, and natural language processing."

		t.Run("SuggestSchema", func(t *testing.T) {
			// Create temp file with test content
			tmpFile := filepath.Join(tmpDir, "test-doc.txt")
			err := os.WriteFile(tmpFile, []byte(testDoc), 0644)
			require.NoError(t, err)

			output := runWeaveCommand(t, ctx, weaveBinary, "schema", "suggest",
				"--file", tmpFile,
				"--type", "text")
			t.Logf("suggest_schema output: %s", output)
			assert.NotEmpty(t, output)
			// Should contain schema-related keywords
			assert.True(t, strings.Contains(output, "schema") || strings.Contains(output, "properties") || strings.Contains(output, "field"))
		})

		t.Run("SuggestChunking", func(t *testing.T) {
			// Create temp file with test content
			tmpFile := filepath.Join(tmpDir, "test-doc-chunking.txt")
			testContent := strings.Repeat(testDoc+" ", 50) // Make it longer for chunking
			err := os.WriteFile(tmpFile, []byte(testContent), 0644)
			require.NoError(t, err)

			output := runWeaveCommand(t, ctx, weaveBinary, "chunking", "suggest",
				"--file", tmpFile,
				"--type", "text")
			t.Logf("suggest_chunking output: %s", output)
			assert.NotEmpty(t, output)
			// Should contain chunking-related keywords
			assert.True(t, strings.Contains(output, "chunk") || strings.Contains(output, "size") || strings.Contains(output, "strategy"))
		})
	})

	// Step 9: Test document operations via weave CLI
	t.Run("WeaveDocumentOperations", func(t *testing.T) {
		weaveBinary := filepath.Join(tmpDir, "weave")

		// Create test collection
		testCollection := fmt.Sprintf("e2e-docs-%d", time.Now().Unix())
		runWeaveCommand(t, ctx, weaveBinary, "collections", "create",
			"--name", testCollection,
			"--type", "text",
			"--description", "E2E document test collection")

		defer func() {
			runWeaveCommand(t, ctx, weaveBinary, "collections", "delete", "--name", testCollection)
		}()

		// Create test document file
		tmpFile := filepath.Join(tmpDir, "test-document.txt")
		testContent := "This is a test document for E2E integration testing."
		err := os.WriteFile(tmpFile, []byte(testContent), 0644)
		require.NoError(t, err)

		// Add document
		output := runWeaveCommand(t, ctx, weaveBinary, "documents", "add",
			"--collection", testCollection,
			"--file", tmpFile,
			"--url", "https://example.com/test-doc")
		t.Logf("Added document: %s", output)

		// List documents
		output = runWeaveCommand(t, ctx, weaveBinary, "documents", "list",
			"--collection", testCollection)
		t.Logf("Documents: %s", output)

		// Query documents
		output = runWeaveCommand(t, ctx, weaveBinary, "documents", "query",
			"--collection", testCollection,
			"--query", "test document",
			"--limit", "5")
		t.Logf("Query results: %s", output)
		assert.Contains(t, output, "test")

		// Count documents
		output = runWeaveCommand(t, ctx, weaveBinary, "documents", "count",
			"--collection", testCollection)
		t.Logf("Document count: %s", output)
	})

	// Step 10: Clean up Weaviate
	t.Run("CleanupWeaviate", func(t *testing.T) {
		stopWeaviate(t, ctx)
		t.Logf("Stopped Weaviate")
	})
}

// loadEnvFiles loads environment variables from both weave-mcp and weave-cli
func loadEnvFiles(t *testing.T) {
	// Load weave-mcp .env
	if err := loadEnvFile("../../.env"); err != nil {
		t.Logf("Could not load weave-mcp .env: %v", err)
	}

	// Load weave-cli .env if it exists
	weaveCliEnv := filepath.Join("..", "..", "weave-cli", ".env")
	if err := loadEnvFile(weaveCliEnv); err != nil {
		t.Logf("Could not load weave-cli .env: %v", err)
	}
}

// downloadWeaveCLI downloads the weave-cli binary from GitHub release
func downloadWeaveCLI(t *testing.T, ctx context.Context, tmpDir string) string {
	// Determine platform and architecture
	platform := runtime.GOOS
	arch := runtime.GOARCH

	// Map Go arch names to release names
	archMap := map[string]string{
		"amd64": "x86_64",
		"arm64": "arm64",
	}
	releaseArch := archMap[arch]
	if releaseArch == "" {
		releaseArch = arch
	}

	// Construct download URL
	var downloadURL string
	var archiveName string
	if platform == "windows" {
		archiveName = fmt.Sprintf("weave-cli_%s_%s_%s.zip", weaveCLIVersion, platform, releaseArch)
		downloadURL = fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", weaveCLIRepo, weaveCLIVersion, archiveName)
	} else {
		archiveName = fmt.Sprintf("weave-cli_%s_%s_%s.tar.gz", weaveCLIVersion, platform, releaseArch)
		downloadURL = fmt.Sprintf("https://github.com/%s/releases/download/%s/%s", weaveCLIRepo, weaveCLIVersion, archiveName)
	}

	t.Logf("Downloading weave-cli from: %s", downloadURL)

	// Download archive
	archivePath := filepath.Join(tmpDir, archiveName)
	err := downloadFile(ctx, downloadURL, archivePath)
	require.NoError(t, err)

	// Extract archive
	if platform == "windows" {
		err = extractZip(archivePath, tmpDir)
	} else {
		err = extractTarGz(archivePath, tmpDir)
	}
	require.NoError(t, err)

	// Find weave binary
	weaveBinary := filepath.Join(tmpDir, "weave")
	if platform == "windows" {
		weaveBinary += ".exe"
	}

	// Make binary executable
	if platform != "windows" {
		err = os.Chmod(weaveBinary, 0755)
		require.NoError(t, err)
	}

	return weaveBinary
}

// downloadFile downloads a file from URL
func downloadFile(ctx context.Context, url, filepath string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// extractTarGz extracts a .tar.gz archive
func extractTarGz(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

// extractZip extracts a .zip archive
func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(destDir, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// buildWeaveMCP builds the weave-mcp binary

// setupWeaviate starts Weaviate using Docker
func setupWeaviate(t *testing.T, ctx context.Context) {
	// Check if Docker is available
	if !isDockerAvailable() {
		t.Skip("Docker not available, skipping Weaviate setup")
	}

	// Stop any existing Weaviate container
	stopWeaviate(t, ctx)

	// Start Weaviate
	cmd := exec.CommandContext(ctx, "docker", "run", "-d",
		"--name", "weave-e2e-weaviate",
		"-p", "8080:8080",
		"-p", "50051:50051",
		"-e", "AUTHENTICATION_ANONYMOUS_ACCESS_ENABLED=true",
		"-e", "PERSISTENCE_DATA_PATH=/var/lib/weaviate",
		"-e", "DEFAULT_VECTORIZER_MODULE=none",
		"-e", "ENABLE_MODULES=",
		"-e", "CLUSTER_HOSTNAME=node1",
		"semitechnologies/weaviate:latest")

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Docker output: %s", output)
	}
	require.NoError(t, err, "Failed to start Weaviate")

	// Wait for Weaviate to be ready
	require.NoError(t, waitForWeaviate(ctx), "Weaviate failed to start")
}

// stopWeaviate stops the Weaviate Docker container
func stopWeaviate(t *testing.T, ctx context.Context) {
	cmd := exec.CommandContext(ctx, "docker", "rm", "-f", "weave-e2e-weaviate")
	cmd.Run() // Ignore error if container doesn't exist
}

// isDockerAvailable checks if Docker is available
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}

// waitForWeaviate waits for Weaviate to be ready
func waitForWeaviate(ctx context.Context) error {
	timeout := time.After(60 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for Weaviate")
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/v1/.well-known/ready", nil)
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

// startMCPServer starts the weave-mcp server
func startMCPServer(t *testing.T, ctx context.Context, tmpDir string) *os.Process {
	mcpBinary := filepath.Join("..", "..", "bin", "weave-mcp")

	// Create MCP config for local Weaviate
	mcpConfig := filepath.Join(tmpDir, "mcp-config.yaml")
	configContent := `databases:
  default: "local-weaviate"
  vector_databases:
    - name: "local-weaviate"
      type: "weaviate"
      enabled: true
      url: "http://localhost:8080"
      api_key: ""
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
  create_collection: 60
  delete_collection: 60
  create_document: 120
  query_documents: 60
  list_documents: 30
`
	err := os.WriteFile(mcpConfig, []byte(configContent), 0644)
	require.NoError(t, err)

	cmd := exec.Command(mcpBinary, "--config", mcpConfig)
	cmd.Dir = filepath.Dir(mcpBinary)
	cmd.Env = os.Environ()

	// Create log file for MCP server output
	logFile := filepath.Join(tmpDir, "mcp-server.log")
	log, err := os.Create(logFile)
	require.NoError(t, err)

	cmd.Stdout = log
	cmd.Stderr = log

	err = cmd.Start()
	require.NoError(t, err)

	t.Logf("MCP server log: %s", logFile)

	return cmd.Process
}

// waitForMCPServer waits for MCP server to be ready
// runWeaveCommand runs a weave CLI command and returns output
func runWeaveCommand(t *testing.T, ctx context.Context, weaveBinary string, args ...string) string {
	cmd := exec.CommandContext(ctx, weaveBinary, args...)
	cmd.Env = os.Environ()

	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Command failed: %s %v", weaveBinary, args)
		t.Logf("Output: %s", output)
		t.Logf("Error: %v", err)
	}
	require.NoError(t, err, "weave command failed: %s", args)

	return string(output)
}
