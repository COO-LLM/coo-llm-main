package balancer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/store"
)

type mockStore struct {
	data map[string]map[string]map[string]float64
}

func newMockStore() *mockStore {
	return &mockStore{data: make(map[string]map[string]map[string]float64)}
}

type mockStoreProvider struct {
	*mockStore
}

func newMockStoreProvider() *mockStoreProvider {
	return &mockStoreProvider{mockStore: newMockStore()}
}

func newTestLogger() *log.Logger {
	return log.NewLogger(&config.Logging{})
}

func (m *mockStoreProvider) LoadConfig() (*config.Config, error) {
	return &config.Config{Policy: config.Policy{Algorithm: "round_robin"}}, nil
}

func (m *mockStoreProvider) SaveConfig(cfg *config.Config) error {
	return nil
}

func (m *mockStoreProvider) CreateClient(clientID, apiKey, description string, allowedProviders []string) error {
	return nil
}

func (m *mockStoreProvider) UpdateClient(clientID, description string, allowedProviders []string) error {
	return nil
}

func (m *mockStoreProvider) DeleteClient(clientID string) error {
	return nil
}

func (m *mockStoreProvider) GetClient(clientID string) (*store.ClientInfo, error) {
	return nil, nil
}

func (m *mockStoreProvider) ListClients() ([]*store.ClientInfo, error) {
	return nil, nil
}

func (m *mockStoreProvider) ValidateClient(apiKey string) (*store.ClientInfo, error) {
	return nil, nil
}

func (m *mockStoreProvider) GetClientMetrics(clientID string, start, end int64) (*store.ClientMetrics, error) {
	return nil, nil
}

func (m *mockStoreProvider) GetProviderMetrics(providerID string, start, end int64) (*store.ProviderMetrics, error) {
	return nil, nil
}

func (m *mockStoreProvider) GetKeyMetrics(providerID, keyID string, start, end int64) (*store.KeyMetrics, error) {
	return nil, nil
}

func (m *mockStoreProvider) GetGlobalMetrics(start, end int64) (*store.GlobalMetrics, error) {
	return nil, nil
}

func (m *mockStoreProvider) GetClientTimeSeries(clientID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return nil, nil
}

func (m *mockStoreProvider) GetProviderTimeSeries(providerID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return nil, nil
}

func (m *mockStoreProvider) GetKeyTimeSeries(providerID, keyID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return nil, nil
}

func (m *mockStoreProvider) SaveAlgorithmConfig(algorithm string, config map[string]interface{}) error {
	return nil
}

func (m *mockStoreProvider) LoadAlgorithmConfig(algorithm string) (map[string]interface{}, error) {
	return nil, nil
}

func (m *mockStoreProvider) ListAlgorithms() ([]string, error) {
	return []string{"round_robin", "least_loaded", "hybrid"}, nil
}

func (m *mockStore) GetUsage(provider, keyID, metric string) (float64, error) {
	if m.data[provider] != nil && m.data[provider][keyID] != nil {
		return m.data[provider][keyID][metric], nil
	}
	return 0, nil
}

func (m *mockStore) SetUsage(provider, keyID, metric string, value float64) error {
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

func TestRateLimiting(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				ID: "openai",
				Keys: []config.Key{
					{ID: "key1", LimitReqPerMin: 10, LimitTokensPerMin: 1000},
					{ID: "key2", LimitReqPerMin: 10, LimitTokensPerMin: 1000},
				},
			},
		},
		Policy: config.Policy{Algorithm: "round_robin"},
	}

	store := newMockStoreProvider()
	// Set key1 as rate limited (exceeded req limit)
	store.SetUsage("openai", "key1", "req", 15) // 15 > 10 limit

	selector := NewSelector(cfg, store, newTestLogger())

	// Should select key2 since key1 is rate limited
	key, err := selector.selectKey(&cfg.Providers[0], "gpt-4o")
	require.NoError(t, err)
	assert.Equal(t, "key2", key.ID)

	// Test with both keys rate limited - should still select one
	store.SetUsage("openai", "key2", "req", 15)
	key, err = selector.selectKey(&cfg.Providers[0], "gpt-4o")
	require.NoError(t, err)
	// Should select one of the keys (allows bursting)
	assert.True(t, key.ID == "key1" || key.ID == "key2")
}

func TestResolveModel(t *testing.T) {
	cfg := &config.Config{
		ModelAliases: map[string]string{
			"gpt-4o": "openai:gpt-4o",
		},
	}
	store := newMockStoreProvider()

	selector := NewSelector(cfg, store, newTestLogger())

	providerID, modelName := selector.resolveModel("gpt-4o")
	assert.Equal(t, "openai", providerID)
	assert.Equal(t, "gpt-4o", modelName)

	providerID, modelName = selector.resolveModel("unknown")
	assert.Equal(t, "openai", providerID)
	assert.Equal(t, "unknown", modelName)
}

func TestSelectBest(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				ID: "openai",
				Keys: []config.Key{
					{ID: "key1", Secret: "sk-test"},
				},
			},
		},
		ModelAliases: map[string]string{
			"gpt-4o": "openai:gpt-4o",
		},
		Policy: config.Policy{Algorithm: "round_robin"},
	}
	store := newMockStoreProvider()
	selector := NewSelector(cfg, store, newTestLogger())

	pCfg, key, _, err := selector.SelectBest("gpt-4o")
	require.NoError(t, err)
	assert.Equal(t, "openai", pCfg.ID)
	assert.Equal(t, "key1", key.ID)

	pCfg, key, _, err = selector.SelectBest("unknown")
	require.NoError(t, err)
	assert.Equal(t, "openai", pCfg.ID)
	assert.Equal(t, "key1", key.ID)
}

func TestSelectRoundRobin(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				ID: "openai",
				Keys: []config.Key{
					{ID: "key1"},
					{ID: "key2"},
				},
			},
		},
		Policy: config.Policy{Algorithm: "round_robin"},
	}
	store := newMockStoreProvider()
	selector := NewSelector(cfg, store, newTestLogger())

	key, err := selector.selectRoundRobin(&cfg.Providers[0])
	require.NoError(t, err)
	assert.Contains(t, []string{"key1", "key2"}, key.ID)
}

func TestSelectHybrid(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				ID:      "openai",
				Pricing: config.Pricing{InputTokenCost: 0.01, OutputTokenCost: 0.02},
				Keys: []config.Key{
					{
						ID:           "key1",
						SessionLimit: 10000,
						SessionType:  "1h",
					},
					{
						ID:           "key2",
						SessionLimit: 20000,
						SessionType:  "1h",
					},
				},
			},
		},
		Policy: config.Policy{
			Algorithm: "hybrid",
			HybridWeights: config.HybridWeights{
				ReqRatio:   0.2,
				TokenRatio: 0.3,
				ErrorScore: 0.2,
				Latency:    0.1,
				CostRatio:  0.2,
			},
		},
	}
	store := newMockStoreProvider()
	store.SetUsage("openai", "key1", "req", 100)
	store.SetUsage("openai", "key2", "req", 50)
	selector := NewSelector(cfg, store, newTestLogger())

	key, err := selector.selectHybrid(&cfg.Providers[0], "gpt-4o", cfg.Policy)
	require.NoError(t, err)
	assert.Equal(t, "key2", key.ID) // key2 has lower req usage
}

func TestCalculateScore(t *testing.T) {
	cfg := &config.Config{
		Policy: config.Policy{
			HybridWeights: config.HybridWeights{
				ReqRatio:   0.2,
				TokenRatio: 0.3,
				ErrorScore: 0.2,
				Latency:    0.1,
				CostRatio:  0.2,
			},
		},
	}
	store := newMockStoreProvider()
	store.SetUsage("openai", "key1", "req", 10)
	store.SetUsage("openai", "key1", "tokens", 1000)
	store.SetUsage("openai", "key1", "errors", 1)
	store.SetUsage("openai", "key1", "latency", 200)
	selector := NewSelector(cfg, store, newTestLogger())

	pCfg := &config.Provider{
		ID:      "openai",
		Pricing: config.Pricing{InputTokenCost: 0.01, OutputTokenCost: 0.02},
		Limits:  config.Limits{MaxTokens: 4000},
	}
	key := &config.Key{
		ID:           "key1",
		SessionLimit: 10000,
		SessionType:  "1h",
	}
	score := selector.calculateScore(pCfg, key, "gpt-4o", cfg.Policy)
	expected := 0.2*10 + 0.3*1000 + 0.2*1 + 0.1*200 + 0.2*(0.01+0.02)*1000/1000 - 4000.0/1000.0*0.1
	assert.InDelta(t, expected, score, 0.001)
}

func TestUpdateUsage(t *testing.T) {
	cfg := &config.Config{}
	store := newMockStoreProvider()
	selector := NewSelector(cfg, store, newTestLogger())

	selector.UpdateUsage("openai", "key1", "req", 5)
	val, _ := store.GetUsage("openai", "key1", "req")
	assert.Equal(t, 5.0, val)

	selector.UpdateUsage("openai", "key1", "req", 3)
	val, _ = store.GetUsage("openai", "key1", "req")
	assert.Equal(t, 8.0, val)
}
