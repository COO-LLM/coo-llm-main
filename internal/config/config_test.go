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
llm_providers:
  - type: "openai"
    api_keys: ["sk-test"]
    base_url: "https://api.openai.com"
    model: "gpt-4o"
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
	assert.Len(t, cfg.LLMProviders, 1)
	assert.Equal(t, "openai", cfg.LLMProviders[0].Type)
	// Should populate Providers from LLMProviders
	assert.Len(t, cfg.Providers, 1)
	assert.Equal(t, "openai", cfg.Providers[0].ID)
}

func TestValidateConfig(t *testing.T) {
	cfg := &Config{
		Version:      "1.0",
		Server:       Server{Listen: ":8080"},
		LLMProviders: []LLMProvider{{Type: "test"}},
	}

	err := ValidateConfig(cfg)
	assert.NoError(t, err)

	cfg.Version = ""
	err = ValidateConfig(cfg)
	assert.Error(t, err)

	cfg.Version = "1.0"
	cfg.LLMProviders = nil
	err = ValidateConfig(cfg)
	assert.Error(t, err)
}
