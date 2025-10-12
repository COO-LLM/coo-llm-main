package test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/truckllm/internal/api"
	"github.com/user/truckllm/internal/balancer"
	"github.com/user/truckllm/internal/config"
	"github.com/user/truckllm/internal/log"
	"github.com/user/truckllm/internal/provider"
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
	if m.data == nil {
		m.data = make(map[string]map[string]map[string]float64)
	}
	if m.data[provider] == nil {
		m.data[provider] = make(map[string]map[string]float64)
	}
	if m.data[provider][keyID] == nil {
		m.data[provider][keyID] = make(map[string]float64)
	}
	val := m.data[provider][keyID][metric]
	m.data[provider][keyID][metric] = val + delta
	return nil
}

// Mock provider that simulates external API
type e2eMockProvider struct{}

func (m *e2eMockProvider) Name() string { return "openai" }
func (m *e2eMockProvider) Generate(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	// Simulate API delay and response
	return &provider.Response{
		RawResponse: []byte(`{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1677652288,
			"model": "gpt-4o",
			"choices": [{
				"index": 0,
				"message": {
					"role": "assistant",
					"content": "Hello! How can I help you today?"
				},
				"finish_reason": "stop"
			}],
			"usage": {
				"prompt_tokens": 9,
				"completion_tokens": 12,
				"total_tokens": 21
			}
		}`),
		HTTPCode:   200,
		Latency:    150,
		TokensUsed: 21,
	}, nil
}
func (m *e2eMockProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"gpt-4o", "gpt-4"}, nil
}

func TestE2E_ChatCompletionsFlow(t *testing.T) {
	// Setup config
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				ID:      "openai",
				BaseURL: "https://api.openai.com/v1",
				Keys: []config.Key{
					{ID: "key1", Secret: "sk-test"},
				},
			},
		},
		ModelAliases: map[string]string{
			"gpt-4o": "openai:gpt-4o",
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
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
		"max_tokens": 100,
	}
	body, _ := json.Marshal(reqBody)

	resp, err := http.Post(ts.URL+"/v1/chat/completions", "application/json", bytes.NewReader(body))
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

	resp, err := http.Get(ts.URL + "/v1/models")
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
