// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package tests

import (
	"os"
	"testing"

	"github.com/maximilien/weave-mcp/src/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Run("LoadConfig with valid config", func(t *testing.T) {
		// Create a temporary config file
		configContent := `
databases:
  default: mock
  vector_databases:
    - name: mock
      type: mock
      enabled: true
      simulate_embeddings: true
      embedding_dimension: 384
      collections:
        - name: TestCollection
          type: text
          description: Test collection
`

		// Write to temporary file
		tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(configContent)
		require.NoError(t, err)
		tmpFile.Close()

		// Load config
		cfg, err := config.LoadConfig(tmpFile.Name(), "")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Verify config
		assert.Equal(t, "mock", cfg.Databases.Default)
		assert.Len(t, cfg.Databases.VectorDatabases, 1)
		assert.Equal(t, "mock", cfg.Databases.VectorDatabases[0].Name)
		assert.Equal(t, config.VectorDBTypeMock, cfg.Databases.VectorDatabases[0].Type)
		assert.True(t, cfg.Databases.VectorDatabases[0].Enabled)
		assert.True(t, cfg.Databases.VectorDatabases[0].SimulateEmbeddings)
		assert.Equal(t, 384, cfg.Databases.VectorDatabases[0].EmbeddingDimension)
		assert.Len(t, cfg.Databases.VectorDatabases[0].Collections, 1)
		assert.Equal(t, "TestCollection", cfg.Databases.VectorDatabases[0].Collections[0].Name)
		assert.Equal(t, "text", cfg.Databases.VectorDatabases[0].Collections[0].Type)
	})

	t.Run("GetDefaultDatabase", func(t *testing.T) {
		cfg := &config.Config{
			Databases: config.DatabasesConfig{
				Default: "test-db",
				VectorDatabases: []config.VectorDBConfig{
					{
						Name: "test-db",
						Type: config.VectorDBTypeMock,
					},
					{
						Name: "other-db",
						Type: config.VectorDBTypeLocal,
					},
				},
			},
		}

		db, err := cfg.GetDefaultDatabase()
		require.NoError(t, err)
		assert.Equal(t, "test-db", db.Name)
		assert.Equal(t, config.VectorDBTypeMock, db.Type)
	})

	t.Run("GetDatabase", func(t *testing.T) {
		cfg := &config.Config{
			Databases: config.DatabasesConfig{
				VectorDatabases: []config.VectorDBConfig{
					{
						Name: "test-db",
						Type: config.VectorDBTypeMock,
					},
					{
						Name: "other-db",
						Type: config.VectorDBTypeLocal,
					},
				},
			},
		}

		// Test existing database
		db, err := cfg.GetDatabase("test-db")
		require.NoError(t, err)
		assert.Equal(t, "test-db", db.Name)
		assert.Equal(t, config.VectorDBTypeMock, db.Type)

		// Test non-existing database
		_, err = cfg.GetDatabase("non-existing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	t.Run("ListDatabases", func(t *testing.T) {
		cfg := &config.Config{
			Databases: config.DatabasesConfig{
				VectorDatabases: []config.VectorDBConfig{
					{
						Name: "test-db",
						Type: config.VectorDBTypeMock,
					},
					{
						Name: "other-db",
						Type: config.VectorDBTypeLocal,
					},
				},
			},
		}

		names := cfg.ListDatabases()
		assert.Len(t, names, 2)
		assert.Contains(t, names, "test-db")
		assert.Contains(t, names, "other-db")
	})

	t.Run("GetDatabaseNames", func(t *testing.T) {
		cfg := &config.Config{
			Databases: config.DatabasesConfig{
				VectorDatabases: []config.VectorDBConfig{
					{
						Name: "test-db",
						Type: config.VectorDBTypeMock,
					},
					{
						Name: "other-db",
						Type: config.VectorDBTypeLocal,
					},
				},
			},
		}

		names := cfg.GetDatabaseNames()
		assert.Len(t, names, 2)
		assert.Equal(t, config.VectorDBTypeMock, names["test-db"])
		assert.Equal(t, config.VectorDBTypeLocal, names["other-db"])
	})
}

func TestConfigExtended(t *testing.T) {
	t.Run("LoadConfig with environment variables", func(t *testing.T) {
		// Set environment variables
		os.Setenv("VECTOR_DB_TYPE", "weaviate-cloud")
		os.Setenv("WEAVIATE_URL", "https://test.weaviate.network")
		os.Setenv("WEAVIATE_API_KEY", "test-key")
		os.Setenv("OPENAI_API_KEY", "test-openai-key")

		// Create a temporary config file with environment variable interpolation
		configContent := `
databases:
  default: ${VECTOR_DB_TYPE:-mock}
  vector_databases:
    - name: weaviate-cloud
      type: weaviate-cloud
      url: ${WEAVIATE_URL}
      api_key: ${WEAVIATE_API_KEY}
      openai_api_key: ${OPENAI_API_KEY}
      collections:
        - name: TestCollection
          type: text
          description: Test collection
`

		// Write to temporary file
		tmpFile, err := os.CreateTemp("", "test-config-env-*.yaml")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.WriteString(configContent)
		require.NoError(t, err)
		tmpFile.Close()

		// Load config
		cfg, err := config.LoadConfig(tmpFile.Name(), "")
		require.NoError(t, err)
		require.NotNil(t, cfg)

		// Verify config with environment variables
		assert.Equal(t, "weaviate-cloud", cfg.Databases.Default)
		assert.Len(t, cfg.Databases.VectorDatabases, 1)
		assert.Equal(t, "weaviate-cloud", cfg.Databases.VectorDatabases[0].Name)
		assert.Equal(t, config.VectorDBTypeCloud, cfg.Databases.VectorDatabases[0].Type)
		assert.Equal(t, "https://test.weaviate.network", cfg.Databases.VectorDatabases[0].URL)
		assert.Equal(t, "test-key", cfg.Databases.VectorDatabases[0].APIKey)
		assert.Equal(t, "test-openai-key", cfg.Databases.VectorDatabases[0].OpenAIAPIKey)
	})
}
