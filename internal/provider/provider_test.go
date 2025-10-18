package provider

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/coo-llm/internal/config"
)

func TestRegistry(t *testing.T) {
	reg := NewRegistry()

	mockP := &mockProvider{name: "test"}
	reg.Register(mockP)

	p, err := reg.Get("test")
	require.NoError(t, err)
	assert.Equal(t, "test", p.Name())

	_, err = reg.Get("unknown")
	assert.Error(t, err)

	list := reg.List()
	assert.Contains(t, list, "test")
}

func TestNewLLMProvider(t *testing.T) {
	cfg := LLMConfig{Type: ProviderOpenAI, APIKeys: []string{"test"}, Model: "gpt-4"}
	p, err := NewLLMProvider(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "openai", p.Name())

	cfg.Type = ProviderGemini
	p, err = NewLLMProvider(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "gemini", p.Name())

	cfg.Type = ProviderGrok
	p, err = NewLLMProvider(&cfg)
	require.NoError(t, err)
	assert.Equal(t, "grok", p.Name())

	cfg.Type = "unknown"
	_, err = NewLLMProvider(&cfg)
	assert.Error(t, err)
}

func TestLoadFromConfig(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{ID: "openai", BaseURL: "https://api.openai.com/v1", Keys: []config.Key{{Secret: "key1"}}},
			{ID: "gemini", BaseURL: "https://generativelanguage.googleapis.com", Keys: []config.Key{{Secret: "key2"}}},
		},
	}
	reg := NewRegistry()
	err := reg.LoadFromConfig(cfg)
	require.NoError(t, err)

	p, err := reg.Get("openai")
	require.NoError(t, err)
	assert.Equal(t, "openai", p.Name())

	p, err = reg.Get("gemini")
	require.NoError(t, err)
	assert.Equal(t, "gemini", p.Name())
}

func TestOpenAIProvider_Name(t *testing.T) {
	cfg := LLMConfig{Type: ProviderOpenAI, APIKeys: []string{"test"}}
	p := NewOpenAIProvider(&cfg)
	assert.Equal(t, "openai", p.Name())
}

func TestOpenAIProvider_ListModels(t *testing.T) {
	cfg := LLMConfig{Type: ProviderOpenAI, APIKeys: []string{"test"}}
	p := NewOpenAIProvider(&cfg)
	models, err := p.ListModels(context.Background())
	require.NoError(t, err)
	assert.Contains(t, models, "gpt-4o")
}

func TestGrokProvider_Name(t *testing.T) {
	cfg := LLMConfig{Type: ProviderGrok, APIKeys: []string{"test"}}
	p := NewGrokProvider(&cfg)
	assert.Equal(t, "grok", p.Name())
}

func TestGrokProvider_ListModels(t *testing.T) {
	cfg := LLMConfig{Type: ProviderGrok, APIKeys: []string{"test"}}
	p := NewGrokProvider(&cfg)
	models, err := p.ListModels(context.Background())
	require.NoError(t, err)
	assert.Contains(t, models, "grok-beta")
}

func TestLLMConfig_APIKey(t *testing.T) {
	cfg := LLMConfig{APIKeys: []string{"key1", "key2"}}
	assert.Equal(t, "key1", cfg.APIKey())
}

func TestLLMConfig_NextAPIKey(t *testing.T) {
	cfg := LLMConfig{APIKeys: []string{"key1", "key2", "key3"}}
	assert.Equal(t, "key1", cfg.NextAPIKey())
	assert.Equal(t, "key2", cfg.NextAPIKey())
	assert.Equal(t, "key3", cfg.NextAPIKey())
	assert.Equal(t, "key1", cfg.NextAPIKey()) // Rotate back
}

func TestLLMConfig_NextAPIKeyWithEnv(t *testing.T) {
	os.Setenv("TEST_KEY1", "resolved_key1")
	os.Setenv("TEST_KEY2", "resolved_key2")
	defer os.Unsetenv("TEST_KEY1")
	defer os.Unsetenv("TEST_KEY2")

	cfg := LLMConfig{APIKeys: []string{"${TEST_KEY1}", "${TEST_KEY2}"}}
	assert.Equal(t, "resolved_key1", cfg.NextAPIKey())
	assert.Equal(t, "resolved_key2", cfg.NextAPIKey())
	assert.Equal(t, "resolved_key1", cfg.NextAPIKey()) // Rotate back
}

func TestLLMConfig_SelectLeastLoadedKey(t *testing.T) {
	cfg := LLMConfig{APIKeys: []string{"key1", "key2", "key3"}}
	cfg.InitUsages()

	// Initially all have 0 usage, should return first
	assert.Equal(t, "key1", cfg.SelectLeastLoadedKey())

	// Update usage for key1
	cfg.UpdateUsage(10, 1000)

	// Now key2 should be selected
	assert.Equal(t, "key2", cfg.SelectLeastLoadedKey())

	// Update key2 with more usage
	cfg.UpdateUsage(20, 2000)

	// key3 should be selected
	assert.Equal(t, "key3", cfg.SelectLeastLoadedKey())
}

func TestLLMConfig_SelectLeastLoadedKeyWithEnv(t *testing.T) {
	os.Setenv("TEST_KEY1", "resolved_key1")
	os.Setenv("TEST_KEY2", "resolved_key2")
	defer os.Unsetenv("TEST_KEY1")
	defer os.Unsetenv("TEST_KEY2")

	cfg := LLMConfig{APIKeys: []string{"${TEST_KEY1}", "${TEST_KEY2}"}}
	cfg.InitUsages()

	assert.Equal(t, "resolved_key1", cfg.SelectLeastLoadedKey())

	// Update usage for first key to make second key selected
	cfg.UpdateUsage(10, 1000)

	assert.Equal(t, "resolved_key2", cfg.SelectLeastLoadedKey())
}

func TestLLMConfig_UpdateUsage(t *testing.T) {
	cfg := LLMConfig{APIKeys: []string{"key1", "key2"}}
	cfg.InitUsages()

	cfg.UpdateUsage(5, 500)
	assert.Equal(t, int64(5), cfg.usages[0].ReqCount)
	assert.Equal(t, int64(500), cfg.usages[0].TokenCount)

	cfg.NextAPIKey() // Switch to key2
	cfg.UpdateUsage(3, 300)
	assert.Equal(t, int64(3), cfg.usages[0].ReqCount) // Now key2 is first
	assert.Equal(t, int64(300), cfg.usages[0].TokenCount)
}

// Mock provider for testing
type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Generate(ctx context.Context, req *LLMRequest) (*LLMResponse, error) {
	return &LLMResponse{Text: "ok", TokensUsed: 10, FinishReason: "stop"}, nil
}
func (m *mockProvider) GenerateStream(ctx context.Context, req *LLMRequest) (<-chan *LLMStreamResponse, error) {
	streamChan := make(chan *LLMStreamResponse, 1)
	go func() {
		defer close(streamChan)
		streamChan <- &LLMStreamResponse{Text: "ok", Done: true}
	}()
	return streamChan, nil
}
func (m *mockProvider) CreateEmbeddings(ctx context.Context, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	return &EmbeddingsResponse{
		Embeddings: []Embedding{[]float64{0.1, 0.2, 0.3}},
		Usage: TokenUsage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}, nil
}
func (m *mockProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"model1"}, nil
}
