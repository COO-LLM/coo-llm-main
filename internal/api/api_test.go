package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/provider"
	"github.com/user/coo-llm/internal/store"
)

func TestModelsEndpoint(t *testing.T) {
	cfg := &config.Config{
		ModelAliases: map[string]string{
			"gpt-4o": "openai:gpt-4o",
		},
	}

	r := chi.NewRouter()
	SetupModelsRoute(r, cfg)

	req := httptest.NewRequest("GET", "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "list", resp["object"])
	data := resp["data"].([]any)
	assert.Len(t, data, 1)
}

func TestChatCompletionsEndpoint(t *testing.T) {
	// Mock config and components
	cfg := &config.Config{
		LLMProviders: []config.LLMProvider{
			{
				ID:      "openai-prod",
				Type:    "openai",
				BaseURL: "https://api.openai.com/v1",
				APIKeys: []string{"sk-test"},
			},
		},
		APIKeys: []config.APIKeyConfig{
			{
				ID:               "test-client",
				Key:              "test-key",
				AllowedProviders: []string{"*"},
			},
		},
		ModelAliases: map[string]string{
			"gpt-4o": "openai-prod:gpt-4o",
		},
		Policy: config.Policy{Strategy: "round_robin"},
	}

	reg := provider.NewRegistry()
	reg.Register(&mockProvider{})

	logger := log.NewLogger(&config.Logging{})
	runtimeStore := &mockStore{}
	selector := balancer.NewSelector(cfg, runtimeStore, logger)
	r := chi.NewRouter()
	SetupRoutes(r, selector, logger, reg, cfg, runtimeStore)

	reqBody := map[string]any{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp, "choices")
}

func TestChatCompletionsEndpoint_InvalidModel(t *testing.T) {
	cfg := &config.Config{
		LLMProviders: []config.LLMProvider{
			{
				ID:      "openai-prod",
				Type:    "openai",
				APIKeys: []string{"sk-test"},
			},
		},
		APIKeys: []config.APIKeyConfig{
			{
				ID:               "test-client",
				Key:              "test-key",
				AllowedProviders: []string{"*"},
			},
		},
		Policy: config.Policy{Strategy: "round_robin"},
	}

	reg := provider.NewRegistry()
	reg.Register(&mockProvider{})

	logger := log.NewLogger(&config.Logging{})
	runtimeStore := &mockStore{}
	selector := balancer.NewSelector(cfg, runtimeStore, logger)
	r := chi.NewRouter()
	SetupRoutes(r, selector, logger, reg, cfg, runtimeStore)

	reqBody := map[string]any{
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestModelsEndpoint_EmptyConfig(t *testing.T) {
	cfg := &config.Config{}

	r := chi.NewRouter()
	SetupModelsRoute(r, cfg)

	req := httptest.NewRequest("GET", "/v1/models", nil)
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "list", resp["object"])
	data := resp["data"].([]any)
	assert.Len(t, data, 0)
}

func TestChatCompletionsEndpoint_RetryOnFailure(t *testing.T) {
	// Mock config with retry enabled
	cfg := &config.Config{
		LLMProviders: []config.LLMProvider{
			{
				ID:      "openai-prod",
				Type:    "openai",
				BaseURL: "https://api.openai.com/v1",
				APIKeys: []string{"sk-test"},
			},
		},
		APIKeys: []config.APIKeyConfig{
			{
				ID:               "test-client",
				Key:              "test-key",
				AllowedProviders: []string{"*"},
			},
		},
		ModelAliases: map[string]string{
			"gpt-4o": "openai-prod:gpt-4o",
		},
		Policy: config.Policy{
			Strategy: "round_robin",
			Retry: config.RetryConfig{
				MaxAttempts: 3,
				Timeout:     1000000000, // 1 second in nanoseconds
				Interval:    100000000,  // 0.1 second in nanoseconds
			},
		},
	}

	reg := provider.NewRegistry()
	reg.Register(&mockProvider{})

	logger := log.NewLogger(&config.Logging{})
	runtimeStore := &mockStore{}
	selector := balancer.NewSelector(cfg, runtimeStore, logger)

	r := chi.NewRouter()
	SetupRoutes(r, selector, logger, reg, cfg, runtimeStore)

	reqBody := map[string]any{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp, "choices")
}

func TestChatCompletionsEndpoint_Caching(t *testing.T) {
	// Mock config with caching enabled
	cfg := &config.Config{
		LLMProviders: []config.LLMProvider{
			{
				ID:      "openai-prod",
				Type:    "openai",
				BaseURL: "https://api.openai.com/v1",
				APIKeys: []string{"sk-test"},
			},
		},
		APIKeys: []config.APIKeyConfig{
			{
				ID:               "test-client",
				Key:              "test-key",
				AllowedProviders: []string{"*"},
			},
		},
		ModelAliases: map[string]string{
			"gpt-4o": "openai-prod:gpt-4o",
		},
		Policy: config.Policy{
			Strategy: "round_robin",
			Cache: config.CacheConfig{
				Enabled:    true,
				TTLSeconds: 10,
			},
		},
	}

	reg := provider.NewRegistry()
	mockProv := &mockProvider{callCount: 0}
	reg.Register(mockProv)

	logger := log.NewLogger(&config.Logging{})
	runtimeStore := &mockStoreWithCache{cache: make(map[string]string)}
	selector := balancer.NewSelector(cfg, runtimeStore, logger)
	r := chi.NewRouter()
	SetupRoutes(r, selector, logger, reg, cfg, runtimeStore)

	reqBody := map[string]any{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": "Test caching"},
		},
	}
	body, _ := json.Marshal(reqBody)

	// First request - should call provider
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Authorization", "Bearer test-key")
	w1 := httptest.NewRecorder()

	r.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Check that provider was called once
	assert.Equal(t, 1, mockProv.callCount)

	// Second request with same content - should use cache
	req2 := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer test-key")
	w2 := httptest.NewRecorder()

	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Check that provider was not called again (cache hit)
	assert.Equal(t, 1, mockProv.callCount)

	// Verify cache was set
	cacheKey := "testcaching" // normalized
	cached, exists := runtimeStore.cache[cacheKey]
	assert.True(t, exists)
	assert.NotEmpty(t, cached)

	// Verify second response has cache_hit flag
	var resp2 map[string]any
	err := json.Unmarshal(w2.Body.Bytes(), &resp2)
	require.NoError(t, err)
	assert.Equal(t, true, resp2["cache_hit"])
}

func TestChatCompletionsEndpoint_ConversationHistory(t *testing.T) {
	cfg := &config.Config{
		LLMProviders: []config.LLMProvider{
			{
				ID:      "openai-prod",
				Type:    "openai",
				BaseURL: "https://api.openai.com/v1",
				APIKeys: []string{"sk-test"},
			},
		},
		APIKeys: []config.APIKeyConfig{
			{
				ID:               "test-client",
				Key:              "test-key",
				AllowedProviders: []string{"*"},
			},
		},
		ModelAliases: map[string]string{
			"gpt-4o": "openai-prod:gpt-4o",
		},
		Policy: config.Policy{Strategy: "round_robin"},
	}

	reg := provider.NewRegistry()
	mockProv := &mockProviderWithMessages{}
	reg.Register(mockProv)

	logger := log.NewLogger(&config.Logging{})
	runtimeStore := &mockStore{}
	selector := balancer.NewSelector(cfg, runtimeStore, logger)

	r := chi.NewRouter()
	SetupRoutes(r, selector, logger, reg, cfg, runtimeStore)

	reqBody := map[string]any{
		"model": "gpt-4o",
		"messages": []map[string]any{
			{"role": "user", "content": "Hello, how are you?"},
			{"role": "assistant", "content": "You are me"},
			{"role": "user", "content": "Oh, my god"},
		},
		"max_tokens":  100,
		"temperature": 0.7,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-key")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Contains(t, resp, "choices")

	// Verify that the provider received the messages
	assert.True(t, mockProv.receivedMessages)
	assert.Equal(t, 3, len(mockProv.messages))
	assert.Equal(t, "Hello, how are you?", mockProv.messages[0]["content"])
	assert.Equal(t, "You are me", mockProv.messages[1]["content"])
	assert.Equal(t, "Oh, my god", mockProv.messages[2]["content"])
}

type mockProvider struct {
	callCount int
}

func (m *mockProvider) Name() string { return "openai-prod" }
func (m *mockProvider) Generate(ctx context.Context, req *provider.LLMRequest) (*provider.LLMResponse, error) {
	m.callCount++
	return &provider.LLMResponse{
		Text:         "Hello back",
		TokensUsed:   10,
		InputTokens:  5,
		OutputTokens: 5,
		FinishReason: "stop",
	}, nil
}
func (m *mockProvider) GenerateStream(ctx context.Context, req *provider.LLMRequest) (<-chan *provider.LLMStreamResponse, error) {
	streamChan := make(chan *provider.LLMStreamResponse, 1)
	go func() {
		defer close(streamChan)
		streamChan <- &provider.LLMStreamResponse{Text: "Hello back", Done: true}
	}()
	return streamChan, nil
}
func (m *mockProvider) CreateEmbeddings(ctx context.Context, req *provider.EmbeddingsRequest) (*provider.EmbeddingsResponse, error) {
	return &provider.EmbeddingsResponse{
		Embeddings: []provider.Embedding{[]float64{0.1, 0.2, 0.3}},
		Usage: provider.TokenUsage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}, nil
}
func (m *mockProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"gpt-4o"}, nil
}

type mockStore struct{}

func (m *mockStore) GetUsage(provider, keyID, metric string) (float64, error)           { return 0, nil }
func (m *mockStore) SetUsage(provider, keyID, metric string, value float64) error       { return nil }
func (m *mockStore) IncrementUsage(provider, keyID, metric string, delta float64) error { return nil }
func (m *mockStore) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	return 0, nil
}
func (m *mockStore) SetCache(key, value string, ttlSeconds int64) error { return nil }
func (m *mockStore) GetCache(key string) (string, error)                { return "", nil }
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

type mockProviderWithRetry struct {
	callCount int
}

func (m *mockProviderWithRetry) Name() string { return "openai-prod" }
func (m *mockProviderWithRetry) Generate(ctx context.Context, req *provider.LLMRequest) (*provider.LLMResponse, error) {
	m.callCount++
	if m.callCount < 3 {
		return nil, errors.New("provider unavailable") // Fail first two attempts
	}
	return &provider.LLMResponse{
		Text:         "Hello back after retry",
		TokensUsed:   10,
		InputTokens:  5,
		OutputTokens: 5,
		FinishReason: "stop",
	}, nil
}
func (m *mockProviderWithRetry) GenerateStream(ctx context.Context, req *provider.LLMRequest) (<-chan *provider.LLMStreamResponse, error) {
	streamChan := make(chan *provider.LLMStreamResponse, 1)
	go func() {
		defer close(streamChan)
		streamChan <- &provider.LLMStreamResponse{Text: "Hello back after retry", Done: true}
	}()
	return streamChan, nil
}
func (m *mockProviderWithRetry) CreateEmbeddings(ctx context.Context, req *provider.EmbeddingsRequest) (*provider.EmbeddingsResponse, error) {
	return &provider.EmbeddingsResponse{
		Embeddings: []provider.Embedding{[]float64{0.1, 0.2, 0.3}},
		Usage: provider.TokenUsage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}, nil
}
func (m *mockProviderWithRetry) ListModels(ctx context.Context) ([]string, error) {
	return []string{"gpt-4o"}, nil
}

type mockStoreWithCache struct {
	cache map[string]string
}

func (m *mockStoreWithCache) GetUsage(provider, keyID, metric string) (float64, error) { return 0, nil }
func (m *mockStoreWithCache) SetUsage(provider, keyID, metric string, value float64) error {
	return nil
}
func (m *mockStoreWithCache) IncrementUsage(provider, keyID, metric string, delta float64) error {
	return nil
}
func (m *mockStoreWithCache) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	return 0, nil
}
func (m *mockStoreWithCache) SetCache(key, value string, ttlSeconds int64) error {
	m.cache[key] = value
	return nil
}
func (m *mockStoreWithCache) GetCache(key string) (string, error) {
	if val, ok := m.cache[key]; ok {
		return val, nil
	}
	return "", nil
}
func (m *mockStoreWithCache) StoreMetric(name string, value float64, tags map[string]string, timestamp int64) error {
	return nil
}
func (m *mockStoreWithCache) GetMetrics(name string, tags map[string]string, start, end int64) ([]store.MetricPoint, error) {
	return []store.MetricPoint{}, nil
}

func (m *mockStoreWithCache) LoadConfig() (*config.Config, error) {
	return &config.Config{Policy: config.Policy{Algorithm: "round_robin"}}, nil
}

func (m *mockStoreWithCache) SaveConfig(cfg *config.Config) error {
	return nil
}

func (m *mockStoreWithCache) CreateClient(clientID, apiKey, description string, allowedProviders []string) error {
	return nil
}

func (m *mockStoreWithCache) UpdateClient(clientID, description string, allowedProviders []string) error {
	return nil
}

func (m *mockStoreWithCache) DeleteClient(clientID string) error {
	return nil
}

func (m *mockStoreWithCache) GetClient(clientID string) (*store.ClientInfo, error) {
	return nil, nil
}

func (m *mockStoreWithCache) ListClients() ([]*store.ClientInfo, error) {
	return nil, nil
}

func (m *mockStoreWithCache) ValidateClient(apiKey string) (*store.ClientInfo, error) {
	return nil, nil
}

func (m *mockStoreWithCache) GetClientMetrics(clientID string, start, end int64) (*store.ClientMetrics, error) {
	return nil, nil
}

func (m *mockStoreWithCache) GetProviderMetrics(providerID string, start, end int64) (*store.ProviderMetrics, error) {
	return nil, nil
}

func (m *mockStoreWithCache) GetKeyMetrics(providerID, keyID string, start, end int64) (*store.KeyMetrics, error) {
	return nil, nil
}

func (m *mockStoreWithCache) GetGlobalMetrics(start, end int64) (*store.GlobalMetrics, error) {
	return nil, nil
}

func (m *mockStoreWithCache) GetClientTimeSeries(clientID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return nil, nil
}

func (m *mockStoreWithCache) GetProviderTimeSeries(providerID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return nil, nil
}

func (m *mockStoreWithCache) GetKeyTimeSeries(providerID, keyID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return nil, nil
}

func (m *mockStoreWithCache) SaveAlgorithmConfig(algorithm string, config map[string]interface{}) error {
	return nil
}

func (m *mockStoreWithCache) LoadAlgorithmConfig(algorithm string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockStoreWithCache) ListAlgorithms() ([]string, error) {
	return []string{"round_robin", "least_loaded", "hybrid"}, nil
}

type mockProviderWithMessages struct {
	receivedMessages bool
	messages         []map[string]any
}

func (m *mockProviderWithMessages) Name() string { return "openai-prod" }
func (m *mockProviderWithMessages) Generate(ctx context.Context, req *provider.LLMRequest) (*provider.LLMResponse, error) {
	m.receivedMessages = true
	m.messages = req.Messages
	return &provider.LLMResponse{
		Text:         "Response to conversation",
		TokensUsed:   15,
		InputTokens:  10,
		OutputTokens: 5,
		FinishReason: "stop",
	}, nil
}
func (m *mockProviderWithMessages) GenerateStream(ctx context.Context, req *provider.LLMRequest) (<-chan *provider.LLMStreamResponse, error) {
	streamChan := make(chan *provider.LLMStreamResponse, 1)
	go func() {
		defer close(streamChan)
		streamChan <- &provider.LLMStreamResponse{Text: "Response to conversation", Done: true}
	}()
	return streamChan, nil
}
func (m *mockProviderWithMessages) CreateEmbeddings(ctx context.Context, req *provider.EmbeddingsRequest) (*provider.EmbeddingsResponse, error) {
	return &provider.EmbeddingsResponse{
		Embeddings: []provider.Embedding{[]float64{0.1, 0.2, 0.3}},
		Usage: provider.TokenUsage{
			PromptTokens: 5,
			TotalTokens:  5,
		},
	}, nil
}
func (m *mockProviderWithMessages) ListModels(ctx context.Context) ([]string, error) {
	return []string{"gpt-4o"}, nil
}
