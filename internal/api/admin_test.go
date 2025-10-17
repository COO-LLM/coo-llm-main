package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/coo-llm/internal/balancer"
	"github.com/user/coo-llm/internal/config"
	"github.com/user/coo-llm/internal/store"
)

func TestAdminMetricsAPI(t *testing.T) {
	cfg := &config.Config{
		Server: config.Server{
			AdminAPIKey: "test-admin-key",
		},
	}

	// Mock store
	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore)
	handler := NewAdminHandler(cfg, mockStore, selector)

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

	var response map[string]interface{}
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
			{Key: "client1", AllowedProviders: []string{"*"}},
			{Key: "client2", AllowedProviders: []string{"openai"}},
		},
	}

	mockStore := &mockStoreWithMetrics{}
	selector := balancer.NewSelector(cfg, mockStore)
	handler := NewAdminHandler(cfg, mockStore, selector)

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

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "client_stats")
	clientStats := response["client_stats"].(map[string]interface{})
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
	selector := balancer.NewSelector(cfg, mockStore)
	handler := NewAdminHandler(cfg, mockStore, selector)

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

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Contains(t, response, "stats")
	assert.Contains(t, response, "group_by")
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
