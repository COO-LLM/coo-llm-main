package provider

import (
	"context"
	"fmt"
	"sync"
	"time"

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
)

// KeyUsage tracks usage for each API key
type KeyUsage struct {
	ReqCount   int64
	TokenCount int64
	LastUsed   time.Time
}

// LLMConfig holds configuration for LLM providers
type LLMConfig struct {
	Type    ProviderType   `yaml:"type" mapstructure:"type"`
	APIKeys []string       `yaml:"api_keys" mapstructure:"api_keys"`
	BaseURL string         `yaml:"base_url,omitempty" mapstructure:"base_url,omitempty"`
	Model   string         `yaml:"model" mapstructure:"model"`
	Pricing config.Pricing `yaml:"pricing" mapstructure:"pricing"`
	Limits  config.Limits  `yaml:"limits" mapstructure:"limits"`
	mu      sync.Mutex     // For thread-safe key rotation
	usages  []KeyUsage     // Usage tracking for each key
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

// InitUsages initializes usage tracking for keys
func (c *LLMConfig) InitUsages() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.usages) == 0 {
		c.usages = make([]KeyUsage, len(c.APIKeys))
		for i := range c.usages {
			c.usages[i] = KeyUsage{LastUsed: time.Now()}
		}
	}
}

// NextAPIKey returns the next API key in round-robin fashion (thread-safe)
func (c *LLMConfig) NextAPIKey() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.APIKeys) == 0 {
		return ""
	}
	if c.APIKeys[0] == "" {
		return ""
	}
	if len(c.usages) != len(c.APIKeys) {
		c.InitUsages()
	}
	key := c.APIKeys[0]
	c.APIKeys = append(c.APIKeys[1:], key)
	c.usages = append(c.usages[1:], c.usages[0])
	return key
}

// SelectLeastLoadedKey returns the key with least usage (req + token ratio)
func (c *LLMConfig) SelectLeastLoadedKey() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.APIKeys) == 0 {
		return ""
	}
	if len(c.usages) != len(c.APIKeys) {
		c.InitUsages()
	}

	minIndex := 0
	minScore := float64(c.usages[0].ReqCount) + float64(c.usages[0].TokenCount)*0.01 // Weight tokens less

	for i := 1; i < len(c.usages); i++ {
		score := float64(c.usages[i].ReqCount) + float64(c.usages[i].TokenCount)*0.01
		if score < minScore {
			minScore = score
			minIndex = i
		}
	}

	// Rotate to put selected key first
	if minIndex > 0 {
		c.APIKeys = append(c.APIKeys[minIndex:], c.APIKeys[:minIndex]...)
		c.usages = append(c.usages[minIndex:], c.usages[:minIndex]...)
	}

	return c.APIKeys[0]
}

// UpdateUsage updates usage for the current key
func (c *LLMConfig) UpdateUsage(reqCount, tokenCount int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.usages) > 0 {
		c.usages[0].ReqCount += int64(reqCount)
		c.usages[0].TokenCount += int64(tokenCount)
		c.usages[0].LastUsed = time.Now()
	}
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
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", config.Type)
	}
}

// Legacy types for backward compatibility
type Provider = LLMProvider
type Request = LLMRequest
type Response = LLMResponse
