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
	"github.com/user/coo-llm/internal/store"
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
func (m *mockStore) StoreMetric(name string, value float64, tags map[string]string, timestamp int64) error {
	return nil
}
func (m *mockStore) GetMetrics(name string, tags map[string]string, start, end int64) ([]store.MetricPoint, error) {
	return []store.MetricPoint{}, nil
}

func (m *mockStore) LoadConfig() (*config.Config, error) {
	return &config.Config{Policy: config.Policy{Algorithm: "round_robin"}}, nil
}

func (m *mockStore) SaveConfig(cfg *config.Config) error {
	return nil
}

func (m *mockStore) CreateClient(clientID, apiKey, description string, allowedProviders []string) error {
	return nil
}

func (m *mockStore) UpdateClient(clientID, description string, allowedProviders []string) error {
	return nil
}

func (m *mockStore) DeleteClient(clientID string) error {
	return nil
}

func (m *mockStore) GetClient(clientID string) (*store.ClientInfo, error) {
	return nil, nil
}

func (m *mockStore) ListClients() ([]*store.ClientInfo, error) {
	return nil, nil
}

func (m *mockStore) ValidateClient(apiKey string) (*store.ClientInfo, error) {
	return nil, nil
}

func (m *mockStore) GetClientMetrics(clientID string, start, end int64) (*store.ClientMetrics, error) {
	return nil, nil
}

func (m *mockStore) GetProviderMetrics(providerID string, start, end int64) (*store.ProviderMetrics, error) {
	return nil, nil
}

func (m *mockStore) GetKeyMetrics(providerID, keyID string, start, end int64) (*store.KeyMetrics, error) {
	return nil, nil
}

func (m *mockStore) GetGlobalMetrics(start, end int64) (*store.GlobalMetrics, error) {
	return nil, nil
}

func (m *mockStore) GetClientTimeSeries(clientID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return nil, nil
}

func (m *mockStore) GetProviderTimeSeries(providerID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return nil, nil
}

func (m *mockStore) GetKeyTimeSeries(providerID, keyID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return nil, nil
}

func (m *mockStore) SaveAlgorithmConfig(algorithm string, config map[string]interface{}) error {
	return nil
}

func (m *mockStore) LoadAlgorithmConfig(algorithm string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockStore) ListAlgorithms() ([]string, error) {
	return []string{"round_robin", "least_loaded", "hybrid"}, nil
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
func (m *e2eMockProvider) GenerateStream(ctx context.Context, req *provider.LLMRequest) (<-chan *provider.LLMStreamResponse, error) {
	streamChan := make(chan *provider.LLMStreamResponse, 1)
	go func() {
		defer close(streamChan)
		streamChan <- &provider.LLMStreamResponse{Text: "Hello! How can I help you today?", Done: true}
	}()
	return streamChan, nil
}
func (m *e2eMockProvider) CreateEmbeddings(ctx context.Context, req *provider.EmbeddingsRequest) (*provider.EmbeddingsResponse, error) {
	return &provider.EmbeddingsResponse{
		Embeddings: []provider.Embedding{[]float64{0.1, 0.2, 0.3}},
		Usage: provider.TokenUsage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
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

	logger := log.NewLogger(&config.Logging{})
	runtimeStore := &mockStore{}
	selector := balancer.NewSelector(cfg, runtimeStore, logger)

	// Setup router
	r := chi.NewRouter()
	api.SetupRoutes(r, selector, logger, reg, cfg, runtimeStore)

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
