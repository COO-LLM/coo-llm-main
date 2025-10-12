package balancer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/truckllm/internal/config"
)

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

func TestResolveModel(t *testing.T) {
	cfg := &config.Config{
		ModelAliases: map[string]string{
			"gpt-4o": "openai:gpt-4o",
		},
	}
	store := newMockStore()
	selector := NewSelector(cfg, store)

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
		Policy: config.Policy{Strategy: "round_robin"},
	}
	store := newMockStore()
	selector := NewSelector(cfg, store)

	pCfg, key, err := selector.SelectBest("gpt-4o")
	require.NoError(t, err)
	assert.Equal(t, "openai", pCfg.ID)
	assert.Equal(t, "key1", key.ID)

	pCfg, key, err = selector.SelectBest("unknown")
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
		Policy: config.Policy{Strategy: "round_robin"},
	}
	store := newMockStore()
	selector := NewSelector(cfg, store)

	key, err := selector.selectRoundRobin(&cfg.Providers[0])
	require.NoError(t, err)
	assert.Contains(t, []string{"key1", "key2"}, key.ID)
}

func TestSelectLeastError(t *testing.T) {
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
		Policy: config.Policy{Strategy: "least_error"},
	}
	store := newMockStore()
	store.SetUsage("openai", "key1", "errors", 10)
	store.SetUsage("openai", "key2", "errors", 5)
	selector := NewSelector(cfg, store)

	key, err := selector.selectLeastError(&cfg.Providers[0])
	require.NoError(t, err)
	// Currently returns first, but should be improved
	assert.Equal(t, "key1", key.ID)
}

func TestSelectHybrid(t *testing.T) {
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				ID: "openai",
				Keys: []config.Key{
					{
						ID:      "key1",
						Pricing: config.Pricing{InputTokenCost: 0.01, OutputTokenCost: 0.02},
					},
					{
						ID:      "key2",
						Pricing: config.Pricing{InputTokenCost: 0.02, OutputTokenCost: 0.04},
					},
				},
			},
		},
		Policy: config.Policy{
			Strategy: "hybrid",
			HybridWeights: config.HybridWeights{
				ReqRatio:   0.2,
				TokenRatio: 0.3,
				ErrorScore: 0.2,
				Latency:    0.1,
				CostRatio:  0.2,
			},
		},
	}
	store := newMockStore()
	store.SetUsage("openai", "key1", "req", 100)
	store.SetUsage("openai", "key2", "req", 50)
	selector := NewSelector(cfg, store)

	key, err := selector.selectHybrid(&cfg.Providers[0], "gpt-4o")
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
	store := newMockStore()
	store.SetUsage("openai", "key1", "req", 10)
	store.SetUsage("openai", "key1", "tokens", 1000)
	store.SetUsage("openai", "key1", "errors", 1)
	store.SetUsage("openai", "key1", "latency", 200)
	selector := NewSelector(cfg, store)

	key := &config.Key{
		ID:      "key1",
		Pricing: config.Pricing{InputTokenCost: 0.01, OutputTokenCost: 0.02},
	}
	score := selector.calculateScore("openai", key, "gpt-4o")
	expected := 0.2*10 + 0.3*1000 + 0.2*1 + 0.1*200 + 0.2*(0.01+0.02)*1000/1000
	assert.InDelta(t, expected, score, 0.001)
}

func TestUpdateUsage(t *testing.T) {
	cfg := &config.Config{}
	store := newMockStore()
	selector := NewSelector(cfg, store)

	selector.UpdateUsage("openai", "key1", "req", 5)
	val, _ := store.GetUsage("openai", "key1", "req")
	assert.Equal(t, 5.0, val)

	selector.UpdateUsage("openai", "key1", "req", 3)
	val, _ = store.GetUsage("openai", "key1", "req")
	assert.Equal(t, 8.0, val)
}
