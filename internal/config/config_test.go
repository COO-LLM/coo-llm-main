package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	configContent := `
version: "1.0"
server:
  listen: ":8080"
  admin_api_key: "test"
providers:
  - id: "openai"
    name: "OpenAI"
    base_url: "https://api.openai.com/v1"
    keys:
      - id: "key1"
        secret: "sk-test"
model_aliases:
  gpt-4o: openai:gpt-4o
`
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "1.0", cfg.Version)
	assert.Equal(t, ":8080", cfg.Server.Listen)
	assert.Len(t, cfg.Providers, 1)
	assert.Equal(t, "openai", cfg.Providers[0].ID)
}

func TestValidateConfig(t *testing.T) {
	cfg := &Config{
		Version:   "1.0",
		Server:    Server{Listen: ":8080"},
		Providers: []Provider{{ID: "test"}},
	}

	err := ValidateConfig(cfg)
	assert.NoError(t, err)

	cfg.Version = ""
	err = ValidateConfig(cfg)
	assert.Error(t, err)
}
