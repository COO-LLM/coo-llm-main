package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/coo-llm/internal/api"
	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/provider"
)

// mockStore implements store.RuntimeStore for testing
type mockStore struct {
	data map[string]map[string]map[string]float64
}

func newMockStore() *mockStore {
	return &mockStore{data: make(map[string]map[string]map[string]float64)}
}

func (m *mockStore) GetUsage(provider, keyID, metric string) (float64, error) {
	if m.data[provider] != nil && m.data[provider][keyID] != nil {
		return m.data[provider][keyID][metric], nil
	}
	return 0, nil
}

func (m *mockStore) SetUsage(provider, keyID, metric string, value float64) error {
	if m.data == nil {
		m.data = make(map[string]map[string]map[string]float64)
	}
	if m.data[provider] == nil {
		m.data[provider] = make(map[string]map[string]float64)
	}
	if m.data[provider][keyID] == nil {
		m.data[provider][keyID] = make(map[string]float64)
	}
	m.data[provider][keyID][metric] = value
	return nil
}

func (m *mockStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	val, _ := m.GetUsage(provider, keyID, metric)
	return m.SetUsage(provider, keyID, metric, val+delta)
}

func (m *mockStore) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	return m.GetUsage(provider, keyID, metric)
}

func (m *mockStore) SetCache(key, value string, ttlSeconds int64) error {
	return nil
}

func (m *mockStore) GetCache(key string) (string, error) {
	return "", nil
}

// Mock provider that simulates external API
type e2eMockProvider struct{}

func (m *e2eMockProvider) Name() string { return "gemini" }
func (m *e2eMockProvider) Generate(ctx context.Context, req *provider.LLMRequest) (*provider.LLMResponse, error) {
	// Simulate API delay and response
	return &provider.LLMResponse{
		Text:         "Hello! How can I help you today?",
		TokensUsed:   21,
		FinishReason: "stop",
	}, nil
}
func (m *e2eMockProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"gemini-1.5-pro"}, nil
}

func TestE2E_ChatCompletionsFlow(t *testing.T) {
	// Setup config
	geminiKey := os.Getenv("GEMINI_KEY_1")
	if geminiKey == "" {
		t.Skip("GEMINI_KEY_1 not set, skipping real API test")
	}
	cfg := &config.Config{
		LLMProviders: []config.LLMProvider{
			{
				ID:      "gemini",
				Type:    "gemini",
				APIKeys: []string{geminiKey},
				BaseURL: "https://generativelanguage.googleapis.com",
				Model:   "gemini-2.0-flash-exp",
			},
		},
		Providers: []config.Provider{
			{
				ID:      "gemini",
				BaseURL: "https://generativelanguage.googleapis.com",
				Keys: []config.Key{
					{ID: "key1", Secret: geminiKey},
				},
			},
		},
		ModelAliases: map[string]string{
			"gemini-1.5-flash": "gemini:gemini-1.5-flash",
		},
		Policy: config.Policy{Strategy: "round_robin"},
	}

	// Setup components
	reg := provider.NewRegistry()
	reg.Register(&e2eMockProvider{})

	runtimeStore := &mockStore{}
	selector := balancer.NewSelector(cfg, runtimeStore)
	logger := log.NewLogger(&config.Logging{})

	// Setup router
	r := chi.NewRouter()
	api.SetupRoutes(r, selector, logger, reg, cfg)

	// Create test server
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test chat completions request
	reqBody := map[string]any{
		"model": "gemini-1.5-flash",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	}
	body, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", ts.URL+"/v1/chat/completions", bytes.NewReader(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "chat.completion", response["object"])
	choices := response["choices"].([]any)
	assert.Len(t, choices, 1)
	choice := choices[0].(map[string]any)
	message := choice["message"].(map[string]any)
	assert.Equal(t, "assistant", message["role"])
	assert.Contains(t, message["content"], "Hello")
}

func TestE2E_ModelsEndpoint(t *testing.T) {
	cfg := &config.Config{
		ModelAliases: map[string]string{
			"gpt-4o": "openai:gpt-4o",
		},
	}

	r := chi.NewRouter()
	api.SetupModelsRoute(r, cfg)

	ts := httptest.NewServer(r)
	defer ts.Close()

	req, err := http.NewRequest("GET", ts.URL+"/v1/models", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer test-key")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]any
	err = json.NewDecoder(resp.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "list", response["object"])
	data := response["data"].([]any)
	assert.Len(t, data, 1)
	model := data[0].(map[string]interface{})
	assert.Equal(t, "gpt-4o", model["id"])
}
