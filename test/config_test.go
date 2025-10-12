package test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/truckllm/internal/config"
)

func TestLoadConfigWithPricing(t *testing.T) {
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
        limit_req_per_min: 200
        limit_tokens_per_min: 100000
        pricing:
          input_token_cost: 0.002
          output_token_cost: 0.01
          currency: "USD"
model_aliases:
  gpt-4o: openai:gpt-4o
policy:
  strategy: "hybrid"
  cost_first: true
  hybrid_weights:
    token_ratio: 0.3
    req_ratio: 0.2
    error_score: 0.2
    latency: 0.1
    cost_ratio: 0.2
`
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	cfg, err := config.LoadConfig(tmpFile.Name())
	require.NoError(t, err)

	assert.Equal(t, "1.0", cfg.Version)
	assert.Len(t, cfg.Providers, 1)
	assert.Equal(t, 0.002, cfg.Providers[0].Keys[0].Pricing.InputTokenCost)
	assert.True(t, cfg.Policy.CostFirst)
}
