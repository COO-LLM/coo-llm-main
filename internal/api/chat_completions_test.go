package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user/coo-llm/internal/config"
)

func TestGetProviderFromModel(t *testing.T) {
	cfg := &config.Config{
		LLMProviders: []config.LLMProvider{
			{ID: "openai", Type: "openai"},
			{ID: "gemini", Type: "gemini"},
			{ID: "claude", Type: "claude"},
			{ID: "custom", Type: "openai"},
		},
		ModelAliases: map[string]string{
			"gpt-4o":   "openai:gpt-4o",
			"my-model": "custom:my-model",
		},
	}

	handler := &ChatCompletionsHandler{cfg: cfg}

	// Test 1: Direct provider:model syntax
	assert.Equal(t, "openai", handler.GetProviderFromModel("openai:gpt-4o"))
	assert.Equal(t, "custom", handler.GetProviderFromModel("custom:my-model"))

	// Test 2: Model aliases
	assert.Equal(t, "openai", handler.GetProviderFromModel("gpt-4o"))
	assert.Equal(t, "custom", handler.GetProviderFromModel("my-model"))

	// Test 3: Pattern matching fallback
	assert.Equal(t, "openai", handler.GetProviderFromModel("gpt-4-turbo"))
	assert.Equal(t, "gemini", handler.GetProviderFromModel("gemini-1.5-pro"))
	assert.Equal(t, "claude", handler.GetProviderFromModel("claude-3-opus"))

	// Test 4: No match
	assert.Equal(t, "", handler.GetProviderFromModel("unknown-model"))
}
