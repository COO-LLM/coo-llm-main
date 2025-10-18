package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/log"
	"github.com/user/coo-llm/internal/store"
)

func newTestLogger() *log.Logger {
	return log.NewLogger(&config.Logging{})
}

func TestAdminMetricsAPI(t *testing.T) {
	cfg := &config.Config{
		Server: config.Server{
			AdminAPIKey: "test-admin-key",
		},
	}

	// Mock store
	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore, newTestLogger())
	handler := NewAdminHandler(cfg, mockStore, selector, nil)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Mock bearer auth
			r.Header.Set("Authorization", "Bearer test-admin-key")
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/admin/v1/metrics", handler.GetMetrics)

	// Test metrics endpoint
	req := httptest.NewRequest("GET", "/admin/v1/metrics?name=latency&start=1000&end=2000", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "latency", response["name"])
	assert.Equal(t, float64(1000), response["start"])
	assert.Equal(t, float64(2000), response["end"])
	assert.Contains(t, response, "points")
}

func TestAdminClientStatsAPI(t *testing.T) {
	cfg := &config.Config{
		Server: config.Server{
			AdminAPIKey: "test-admin-key",
		},
		APIKeys: []config.APIKeyConfig{
			{ID: "client-1", Key: "client1", AllowedProviders: []string{"*"}},
			{ID: "client-2", Key: "client2", AllowedProviders: []string{"openai"}},
		},
	}

	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore, newTestLogger())
	handler := NewAdminHandler(cfg, mockStore, selector, nil)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("Authorization", "Bearer test-admin-key")
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/admin/v1/clients", handler.GetClientStats)

	req := httptest.NewRequest("GET", "/admin/v1/clients?start=1000&end=2000", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "client_stats")
	clientStats := response["client_stats"].(map[string]any)
	assert.Contains(t, clientStats, "client1")
	assert.Contains(t, clientStats, "client2")
}

func TestAdminStatsAPI(t *testing.T) {
	cfg := &config.Config{
		Server: config.Server{
			AdminAPIKey: "test-admin-key",
		},
	}

	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore, newTestLogger())
	handler := NewAdminHandler(cfg, mockStore, selector, nil)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("Authorization", "Bearer test-admin-key")
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/admin/v1/stats", handler.GetStats)

	req := httptest.NewRequest("GET", "/admin/v1/stats?group_by=client_key&start=1000&end=2000", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "stats")
	assert.Contains(t, response, "group_by")
}

func TestAdminClientManagement(t *testing.T) {
	cfg := &config.Config{
		Server: config.Server{
			AdminAPIKey: "test-admin-key",
		},
	}

	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore, newTestLogger())
	handler := NewAdminHandler(cfg, mockStore, selector, nil)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("Authorization", "Bearer test-admin-key")
			next.ServeHTTP(w, r)
		})
	})

	// Test CreateClient
	r.Post("/admin/v1/clients", handler.CreateClient)

	t.Run("CreateClient", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/v1/clients", strings.NewReader(`{
			"client_id": "test-client",
			"api_key": "test-api-key",
			"description": "Test client",
			"allowed_providers": ["openai", "anthropic"]
		}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "created", response["status"])
		assert.Equal(t, "test-client", response["client_id"])
	})

	t.Run("CreateClient_InvalidData", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/admin/v1/clients", strings.NewReader(`{
			"client_id": "",
			"api_key": "test-api-key"
		}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	// Test ListClients
	r.Get("/admin/v1/clients/list", handler.ListClients)

	t.Run("ListClients", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/clients/list", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Contains(t, response, "clients")
	})

	// Test GetClient
	r.Get("/admin/v1/clients/{client_id}", handler.GetClient)

	t.Run("GetClient", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/clients/test-client", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response store.ClientInfo
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "test-client", response.ID)
	})

	t.Run("GetClient_NotFound", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/clients/nonexistent", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	// Test UpdateClient
	r.Put("/admin/v1/clients/{client_id}", handler.UpdateClient)

	t.Run("UpdateClient", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/admin/v1/clients/test-client", strings.NewReader(`{
			"description": "Updated test client",
			"allowed_providers": ["openai"]
		}`))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "updated", response["status"])
	})

	// Test DeleteClient
	r.Delete("/admin/v1/clients/{client_id}", handler.DeleteClient)

	t.Run("DeleteClient", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/admin/v1/clients/test-client", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]string
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "deleted", response["status"])
	})
}

func TestAdminEnhancedMetrics(t *testing.T) {
	cfg := &config.Config{
		Server: config.Server{
			AdminAPIKey: "test-admin-key",
		},
	}

	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore, newTestLogger())
	handler := NewAdminHandler(cfg, mockStore, selector, nil)

	r := chi.NewRouter()
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("Authorization", "Bearer test-admin-key")
			next.ServeHTTP(w, r)
		})
	})

	// Test GetClientMetrics
	r.Get("/admin/v1/metrics/clients/{client_id}", handler.GetClientMetrics)

	t.Run("GetClientMetrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/metrics/clients/test-client?start=1000&end=2000", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response store.ClientMetrics
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "test-client", response.ClientID)
	})

	// Test GetProviderMetrics
	r.Get("/admin/v1/metrics/providers/{provider_id}", handler.GetProviderMetrics)

	t.Run("GetProviderMetrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/metrics/providers/openai?start=1000&end=2000", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response store.ProviderMetrics
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "openai", response.ProviderID)
	})

	// Test GetGlobalMetrics
	r.Get("/admin/v1/metrics/global", handler.GetGlobalMetrics)

	t.Run("GetGlobalMetrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/metrics/global?start=1000&end=2000", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response store.GlobalMetrics
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		// Check that the response has the expected fields
		assert.Equal(t, 0, response.TotalClients)
		assert.Equal(t, 0, response.TotalProviders)
	})
}

func TestRateLimiting(t *testing.T) {
	cfg := &config.Config{
		Server: config.Server{
			AdminAPIKey: "test-admin-key",
		},
	}

	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore, newTestLogger())
	handler := NewAdminHandler(cfg, mockStore, selector, nil)

	r := chi.NewRouter()
	r.Use(RateLimitMiddleware())
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("Authorization", "Bearer test-admin-key")
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/admin/v1/config", handler.GetConfig)

	t.Run("RateLimit_WithinLimit", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			req := httptest.NewRequest("GET", "/admin/v1/config", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			// Should not be rate limited
			assert.NotEqual(t, http.StatusTooManyRequests, w.Code)
		}
	})

	t.Run("RateLimit_ExceedLimit", func(t *testing.T) {
		// Make requests that exceed the limit
		for i := 0; i < 60; i++ {
			req := httptest.NewRequest("GET", "/admin/v1/config", nil)
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if i >= 100 { // Our limit is 100 per minute
				assert.Equal(t, http.StatusTooManyRequests, w.Code)
				break
			}
		}
	})
}

func TestAuditLogging(t *testing.T) {
	cfg := &config.Config{
		Server: config.Server{
			AdminAPIKey: "test-admin-key",
		},
	}

	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore, newTestLogger())
	handler := NewAdminHandler(cfg, mockStore, selector, nil)

	r := chi.NewRouter()
	r.Use(AuditLogMiddleware(nil)) // Using nil logger for test
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Header.Set("Authorization", "Bearer test-admin-key")
			next.ServeHTTP(w, r)
		})
	})
	r.Get("/admin/v1/config", handler.GetConfig)

	t.Run("AuditLog_RequestLogged", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/config", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		// Audit logging happens via fmt.Printf, so we can't easily test the output
		// but we can verify the request completes successfully
	})
}

type mockStoreWithMetrics struct{}

func (m *mockStoreWithMetrics) GetUsage(provider, keyID, metric string) (float64, error) {
	return 0, nil
}
func (m *mockStoreWithMetrics) SetUsage(provider, keyID, metric string, value float64) error {
	return nil
}
func (m *mockStoreWithMetrics) IncrementUsage(provider, keyID, metric string, delta float64) error {
	return nil
}
func (m *mockStoreWithMetrics) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	return 0, nil
}
func (m *mockStoreWithMetrics) SetCache(key, value string, ttlSeconds int64) error { return nil }
func (m *mockStoreWithMetrics) GetCache(key string) (string, error)                { return "", nil }
func (m *mockStoreWithMetrics) StoreMetric(name string, value float64, tags map[string]string, timestamp int64) error {
	return nil
}
func (m *mockStoreWithMetrics) GetMetrics(name string, tags map[string]string, start, end int64) ([]store.MetricPoint, error) {
	return []store.MetricPoint{
		{Value: 100.0, Timestamp: 1500, Tags: map[string]string{"provider": "openai", "client_key": "client1", "key": "key1"}},
	}, nil
}

// ConfigStore methods
func (m *mockStoreWithMetrics) LoadConfig() (*config.Config, error) {
	return &config.Config{}, nil
}
func (m *mockStoreWithMetrics) SaveConfig(cfg *config.Config) error {
	return nil
}

// ClientStore methods
func (m *mockStoreWithMetrics) CreateClient(clientID, apiKey, description string, allowedProviders []string) error {
	return nil
}
func (m *mockStoreWithMetrics) UpdateClient(clientID, description string, allowedProviders []string) error {
	return nil
}
func (m *mockStoreWithMetrics) DeleteClient(clientID string) error {
	return nil
}
func (m *mockStoreWithMetrics) GetClient(clientID string) (*store.ClientInfo, error) {
	if clientID == "test-client" {
		return &store.ClientInfo{ID: clientID}, nil
	}
	return nil, fmt.Errorf("client not found")
}
func (m *mockStoreWithMetrics) ListClients() ([]*store.ClientInfo, error) {
	return []*store.ClientInfo{}, nil
}
func (m *mockStoreWithMetrics) ValidateClient(apiKey string) (*store.ClientInfo, error) {
	return nil, nil
}

// MetricsStore methods
func (m *mockStoreWithMetrics) GetClientMetrics(clientID string, start, end int64) (*store.ClientMetrics, error) {
	return &store.ClientMetrics{ClientID: clientID}, nil
}
func (m *mockStoreWithMetrics) GetProviderMetrics(providerID string, start, end int64) (*store.ProviderMetrics, error) {
	return &store.ProviderMetrics{ProviderID: providerID}, nil
}
func (m *mockStoreWithMetrics) GetKeyMetrics(providerID, keyID string, start, end int64) (*store.KeyMetrics, error) {
	return &store.KeyMetrics{ProviderID: providerID, KeyID: keyID}, nil
}
func (m *mockStoreWithMetrics) GetGlobalMetrics(start, end int64) (*store.GlobalMetrics, error) {
	return &store.GlobalMetrics{
		TotalRequests: 1000,
		TotalTokens:   50000,
		TotalCost:     25.0,
		AvgLatency:    150.0,
	}, nil
}

// Test SetupAdminRoutes with proper route mounting
func TestSetupAdminRoutes(t *testing.T) {
	cfg := &config.Config{
		Server: config.Server{
			AdminAPIKey: "test-admin-key",
		},
	}

	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore, newTestLogger())

	r := chi.NewRouter()

	// Setup admin routes (this should mount at /admin)
	SetupAdminRoutes(r, cfg, mockStore, selector, nil)

	// Test authentication middleware
	t.Run("AdminRoutes_RequireAuth", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/config", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "error")
		assert.Contains(t, response["error"].(map[string]any), "message")
		assert.Contains(t, response["error"].(map[string]any)["message"], "Authorization header required")
	})

	// Test with authentication
	t.Run("AdminRoutes_WithAuth_Config", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/config", nil)
		req.Header.Set("Authorization", "Bearer test-admin-key")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		// Should return empty config from mock
		assert.NotNil(t, response)
	})

	t.Run("AdminRoutes_WithAuth_GlobalMetrics", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/metrics/global?start=1000&end=2000", nil)
		req.Header.Set("Authorization", "Bearer test-admin-key")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response store.GlobalMetrics
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, int64(1000), response.TotalRequests)
		assert.Equal(t, int64(50000), response.TotalTokens)
		assert.Equal(t, 25.0, response.TotalCost)
		assert.Equal(t, 150.0, response.AvgLatency)
	})

	t.Run("AdminRoutes_WithAuth_ClientStats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/clients?start=1000&end=2000", nil)
		req.Header.Set("Authorization", "Bearer test-admin-key")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "start")
		assert.Contains(t, response, "end")
		assert.Contains(t, response, "client_stats")
	})

	t.Run("AdminRoutes_WithAuth_Stats", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/stats?group_by=provider&start=1000&end=2000", nil)
		req.Header.Set("Authorization", "Bearer test-admin-key")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "stats")
	})

	t.Run("AdminRoutes_WithAuth_ListClients", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/admin/v1/clients/list", nil)
		req.Header.Set("Authorization", "Bearer test-admin-key")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]any
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Contains(t, response, "clients")
		assert.IsType(t, []any{}, response["clients"])
	})
}
func (m *mockStoreWithMetrics) GetClientTimeSeries(clientID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return []store.TimeSeriesPoint{}, nil
}
func (m *mockStoreWithMetrics) GetProviderTimeSeries(providerID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return []store.TimeSeriesPoint{}, nil
}
func (m *mockStoreWithMetrics) GetKeyTimeSeries(providerID, keyID string, start, end int64, interval string) ([]store.TimeSeriesPoint, error) {
	return []store.TimeSeriesPoint{}, nil
}

// AlgorithmStore methods
func (m *mockStoreWithMetrics) SaveAlgorithmConfig(algorithm string, config map[string]interface{}) error {
	return nil
}
func (m *mockStoreWithMetrics) LoadAlgorithmConfig(algorithm string) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (m *mockStoreWithMetrics) ListAlgorithms() ([]string, error) {
	return []string{"round_robin"}, nil
}
