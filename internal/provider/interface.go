package provider

import (
	"context"
	"fmt"

	"github.com/user/coo-llm/internal/config"
)

// LLMProvider defines the interface for LLM providers
type LLMProvider interface {
	Name() string
	Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error)
	ListModels(ctx context.Context) ([]string, error)
}

// ProviderType represents the type of LLM provider
type ProviderType string

const (
	ProviderOpenAI ProviderType = "openai"
	ProviderGemini ProviderType = "gemini"
	ProviderClaude ProviderType = "claude"
	ProviderCustom ProviderType = "custom"
)

// LLMConfig holds configuration for LLM providers
type LLMConfig struct {
	Type    ProviderType   `yaml:"type" mapstructure:"type"`
	APIKeys []string       `yaml:"api_keys" mapstructure:"api_keys"`
	BaseURL string         `yaml:"base_url,omitempty" mapstructure:"base_url,omitempty"`
	Model   string         `yaml:"model" mapstructure:"model"`
	Pricing config.Pricing `yaml:"pricing" mapstructure:"pricing"`
	Limits  config.Limits  `yaml:"limits" mapstructure:"limits"`
}

type Pricing struct {
	InputTokenCost  float64 `yaml:"input_token_cost" mapstructure:"input_token_cost"`
	OutputTokenCost float64 `yaml:"output_token_cost" mapstructure:"output_token_cost"`
	Currency        string  `yaml:"currency" mapstructure:"currency"`
}

type Limits struct {
	ReqPerMin    int `yaml:"req_per_min" mapstructure:"req_per_min"`
	TokensPerMin int `yaml:"tokens_per_min" mapstructure:"tokens_per_min"`
}

// APIKey returns the first API key for backward compatibility
func (c LLMConfig) APIKey() string {
	if len(c.APIKeys) > 0 {
		return c.APIKeys[0]
	}
	return ""
}

// NextAPIKey returns the next API key in round-robin fashion
func (c *LLMConfig) NextAPIKey() string {
	if len(c.APIKeys) == 0 {
		return ""
	}
	// Simple round-robin, can be improved with mutex for concurrency
	if c.APIKeys[0] == "" {
		return ""
	}
	key := c.APIKeys[0]
	c.APIKeys = append(c.APIKeys[1:], key)
	return key
}

// LLMRequest represents a request to generate text
type LLMRequest struct {
	Prompt    string                   `json:"prompt"`
	Messages  []map[string]interface{} `json:"messages,omitempty"`
	Model     string                   `json:"model,omitempty"`
	MaxTokens int                      `json:"max_tokens,omitempty"`
	Params    map[string]any           `json:"params,omitempty"`
}

// LLMResponse represents the response from LLM
type LLMResponse struct {
	Text         string `json:"text"`
	InputTokens  int    `json:"input_tokens"`
	OutputTokens int    `json:"output_tokens"`
	TokensUsed   int    `json:"tokens_used"` // Total
	FinishReason string `json:"finish_reason"`
}

// NewLLMProvider creates a new LLM provider based on config
func NewLLMProvider(config LLMConfig) (LLMProvider, error) {
	switch config.Type {
	case ProviderOpenAI:
		return NewOpenAIProvider(config), nil
	case ProviderGemini:
		return NewGeminiProvider(config), nil
	case ProviderClaude:
		return NewClaudeProvider(config), nil
	case ProviderCustom:
		return NewCustomProvider(config), nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}

// Legacy types for backward compatibility
type Provider = LLMProvider
type Request = LLMRequest
type Response = LLMResponse
