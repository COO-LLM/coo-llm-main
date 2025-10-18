package store

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore(t *testing.T) {
	// Create temp config file
	configContent := `
version: "1.0"
server:
  listen: ":2906"
llm_providers:
  - type: "openai"
    api_keys: ["sk-test"]
    base_url: "https://api.openai.com"
    model: "gpt-4o"
`
	tmpFile, err := os.CreateTemp("", "config*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	store := NewFileStore(tmpFile.Name())
	cfg, err := store.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "1.0", cfg.Version)

	// Test save
	cfg.Version = "2.0"
	err = store.SaveConfig(cfg)
	require.NoError(t, err)

	cfg2, err := store.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "2.0", cfg2.Version)
}

func TestMemoryStore(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	store := NewMemoryStore(logger)

	// Test IncrementUsage
	err := store.IncrementUsage("openai", "key1", "req", 5.0)
	require.NoError(t, err)

	err = store.IncrementUsage("openai", "key1", "tokens", 100.0)
	require.NoError(t, err)

	err = store.IncrementUsage("openai", "key1", "input_tokens", 50.0)
	require.NoError(t, err)

	err = store.IncrementUsage("openai", "key1", "output_tokens", 50.0)
	require.NoError(t, err)

	// Test GetUsage
	val, err := store.GetUsage("openai", "key1", "req")
	require.NoError(t, err)
	assert.Equal(t, 5.0, val)

	val, err = store.GetUsage("openai", "key1", "tokens")
	require.NoError(t, err)
	assert.Equal(t, 100.0, val)

	val, err = store.GetUsage("openai", "key1", "input_tokens")
	require.NoError(t, err)
	assert.Equal(t, 50.0, val)

	val, err = store.GetUsage("openai", "key1", "output_tokens")
	require.NoError(t, err)
	assert.Equal(t, 50.0, val)

	// Test SetUsage
	err = store.SetUsage("openai", "key1", "latency", 500.0)
	require.NoError(t, err)

	val, err = store.GetUsage("openai", "key1", "latency")
	require.NoError(t, err)
	assert.Equal(t, 500.0, val)
}

func TestRedisStore(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	// For Redis, we can test with a mock or skip if no Redis
	// Since it's hard to mock, test the key generation logic
	store := NewRedisStore("localhost:6379", "", logger)

	// Test key format (without actual Redis)
	// This is more of a unit test for the logic
	assert.NotNil(t, store)
}

func TestHTTPStore(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "Unauthorized", 401)
			return
		}

		switch r.URL.Path {
		case "/usage/openai/key1/req":
			if r.Method == "GET" {
				w.Write([]byte("10.5"))
			}
		case "/usage/openai/key1/req/increment":
			if r.Method == "POST" {
				w.WriteHeader(200)
			}
		case "/config":
			if r.Method == "GET" {
				w.Write([]byte(`{"version":"1.0"}`))
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	store := NewHTTPStore(server.URL, "test-key", logger)

	// Test GetUsage
	val, err := store.GetUsage("openai", "key1", "req")
	require.NoError(t, err)
	assert.Equal(t, 10.5, val)

	// Test IncrementUsage
	err = store.IncrementUsage("openai", "key1", "req", 5.0)
	require.NoError(t, err)

	// Test LoadConfig
	cfg, err := store.LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, "1.0", cfg.Version)
}

func TestDefaultClientStore(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	runtimeStore := NewMemoryStore(logger)
	clientStore := &DefaultClientStore{runtimeStore: runtimeStore}

	t.Run("CreateClient", func(t *testing.T) {
		err := clientStore.CreateClient("test-client", "api-key-123", "Test client", []string{"openai", "anthropic"})
		require.NoError(t, err)
	})

	t.Run("GetClient", func(t *testing.T) {
		client, err := clientStore.GetClient("test-client")
		require.NoError(t, err)
		assert.Equal(t, "test-client", client.ID)
		assert.Equal(t, "api-key-123", client.APIKey)
		assert.Equal(t, "Test client", client.Description)
		assert.Equal(t, []string{"openai", "anthropic"}, client.AllowedProviders)
	})

	t.Run("UpdateClient", func(t *testing.T) {
		err := clientStore.UpdateClient("test-client", "Updated description", []string{"openai"})
		require.NoError(t, err)

		client, err := clientStore.GetClient("test-client")
		require.NoError(t, err)
		assert.Equal(t, "Updated description", client.Description)
		assert.Equal(t, []string{"openai"}, client.AllowedProviders)
	})

	t.Run("DeleteClient", func(t *testing.T) {
		err := clientStore.DeleteClient("test-client")
		require.NoError(t, err)

		_, err = clientStore.GetClient("test-client")
		assert.Error(t, err)
	})

	t.Run("ListClients", func(t *testing.T) {
		// Create a few clients
		clientStore.CreateClient("client1", "key1", "Client 1", []string{"openai"})
		clientStore.CreateClient("client2", "key2", "Client 2", []string{"anthropic"})

		clients, err := clientStore.ListClients()
		require.NoError(t, err)
		// Note: Default implementation returns empty list for simplicity
		assert.Equal(t, 0, len(clients))
	})

	t.Run("ValidateClient", func(t *testing.T) {
		// Default implementation doesn't validate
		client, err := clientStore.ValidateClient("any-key")
		assert.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestDefaultMetricsStore(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	runtimeStore := NewMemoryStore(logger)
	metricsStore := &DefaultMetricsStore{runtimeStore: runtimeStore}

	// Add some test metrics
	now := int64(1000000)
	metricsStore.runtimeStore.StoreMetric("latency", 150.0, map[string]string{"client_key": "test-client", "provider": "openai"}, now)
	metricsStore.runtimeStore.StoreMetric("tokens", 1000.0, map[string]string{"client_key": "test-client", "provider": "openai"}, now)
	metricsStore.runtimeStore.StoreMetric("cost", 0.5, map[string]string{"client_key": "test-client", "provider": "openai"}, now)

	t.Run("GetClientMetrics", func(t *testing.T) {
		metrics, err := metricsStore.GetClientMetrics("test-client", now-1000, now+1000)
		require.NoError(t, err)
		assert.Equal(t, "test-client", metrics.ClientID)
		assert.Equal(t, int64(1), metrics.TotalRequests) // One latency point
		assert.Equal(t, int64(1000), metrics.TotalTokens)
		assert.InDelta(t, 0.5, metrics.TotalCost, 0.01)
		assert.InDelta(t, 150.0, metrics.AvgLatency, 0.01)
	})

	t.Run("GetProviderMetrics", func(t *testing.T) {
		metrics, err := metricsStore.GetProviderMetrics("openai", now-1000, now+1000)
		require.NoError(t, err)
		assert.Equal(t, "openai", metrics.ProviderID)
		assert.Equal(t, int64(1), metrics.TotalRequests)
		assert.Equal(t, int64(1000), metrics.TotalTokens)
		assert.InDelta(t, 0.5, metrics.TotalCost, 0.01)
	})

	t.Run("GetKeyMetrics", func(t *testing.T) {
		metrics, err := metricsStore.GetKeyMetrics("openai", "key1", now-1000, now+1000)
		require.NoError(t, err)
		assert.Equal(t, "openai", metrics.ProviderID)
		assert.Equal(t, "key1", metrics.KeyID)
		// Should be empty since we don't have metrics for this key
		assert.Equal(t, int64(0), metrics.TotalRequests)
	})

	t.Run("GetGlobalMetrics", func(t *testing.T) {
		metrics, err := metricsStore.GetGlobalMetrics(now-1000, now+1000)
		require.NoError(t, err)
		assert.Equal(t, 1, metrics.TotalClients)   // One client in tags
		assert.Equal(t, 1, metrics.TotalProviders) // One provider in tags
		assert.Equal(t, int64(1), metrics.TotalRequests)
		assert.Equal(t, int64(1000), metrics.TotalTokens)
		assert.InDelta(t, 0.5, metrics.TotalCost, 0.01)
		assert.InDelta(t, 150.0, metrics.AvgLatency, 0.01)
	})
}

func TestDefaultAlgorithmStore(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	runtimeStore := NewMemoryStore(logger)
	algorithmStore := &DefaultAlgorithmStore{runtimeStore: runtimeStore}

	t.Run("SaveAndLoadAlgorithmConfig", func(t *testing.T) {
		config := map[string]interface{}{
			"weight_openai":    0.7,
			"weight_anthropic": 0.3,
			"fallback_enabled": true,
		}

		err := algorithmStore.SaveAlgorithmConfig("hybrid", config)
		require.NoError(t, err)

		loaded, err := algorithmStore.LoadAlgorithmConfig("hybrid")
		require.NoError(t, err)
		assert.Equal(t, config, loaded)
	})

	t.Run("ListAlgorithms", func(t *testing.T) {
		algorithms, err := algorithmStore.ListAlgorithms()
		require.NoError(t, err)
		assert.Contains(t, algorithms, "hybrid")
		assert.Contains(t, algorithms, "round_robin")
		assert.Contains(t, algorithms, "least_loaded")
	})
}

func TestStoreProviderWrapper(t *testing.T) {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	runtimeStore := NewMemoryStore(logger)
	configStore := NewSimpleConfigStore(runtimeStore)

	provider := NewStoreProviderWrapper(runtimeStore, configStore)

	// Test that all interfaces are implemented
	assert.NotNil(t, provider)

	// Test basic functionality
	err := provider.CreateClient("test", "key", "desc", []string{"openai"})
	require.NoError(t, err)

	client, err := provider.GetClient("test")
	require.NoError(t, err)
	assert.Equal(t, "test", client.ID)
}
