package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/truckllm/internal/config"
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

func TestLoadFromConfig(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{ID: "openai", BaseURL: "https://api.openai.com/v1"},
			{ID: "gemini", BaseURL: "https://generativelanguage.googleapis.com"},
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

	// Test unsupported
	cfg.Providers = append(cfg.Providers, config.Provider{ID: "unknown"})
	err = reg.LoadFromConfig(cfg)
	assert.Error(t, err)
}

func TestOpenAIProvider_Name(t *testing.T) {
	cfg := &config.Provider{ID: "openai"}
	p := NewOpenAIProvider(cfg)
	assert.Equal(t, "openai", p.Name())
}

func TestOpenAIProvider_ListModels(t *testing.T) {
	cfg := &config.Provider{ID: "openai"}
	p := NewOpenAIProvider(cfg)
	models, err := p.ListModels(context.Background())
	require.NoError(t, err)
	assert.Contains(t, models, "gpt-4o")
}

// Mock provider for testing
type mockProvider struct {
	name string
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Generate(ctx context.Context, req *Request) (*Response, error) {
	return &Response{RawResponse: []byte(`{"response":"ok"}`), HTTPCode: 200}, nil
}
func (m *mockProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"model1"}, nil
}
