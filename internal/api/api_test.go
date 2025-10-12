package api

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
	"github.com/user/truckllm/internal/balancer"
	"github.com/user/truckllm/internal/config"
	"github.com/user/truckllm/internal/log"
	"github.com/user/truckllm/internal/provider"
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
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "list", resp["object"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)
}

func TestChatCompletionsEndpoint(t *testing.T) {
	// Mock config and components
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

	reg := provider.NewRegistry()
	reg.Register(&mockProvider{})

	runtimeStore := &mockStore{}
	selector := balancer.NewSelector(cfg, runtimeStore)
	logger := log.NewLogger(&config.Logging{})

	r := chi.NewRouter()
	SetupRoutes(r, selector, logger, reg, cfg)

	reqBody := map[string]interface{}{
		"model": "gpt-4o",
		"messages": []map[string]string{
			{"role": "user", "content": "Hello"},
		},
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

type mockProvider struct{}

func (m *mockProvider) Name() string { return "openai" }
func (m *mockProvider) Generate(ctx context.Context, req *provider.Request) (*provider.Response, error) {
	return &provider.Response{
		RawResponse: []byte(`{"choices":[{"message":{"content":"Hello back"}}]}`),
		HTTPCode:    200,
		Latency:     100,
	}, nil
}
func (m *mockProvider) ListModels(ctx context.Context) ([]string, error) {
	return []string{"gpt-4o"}, nil
}

type mockStore struct{}

func (m *mockStore) GetUsage(provider, keyID, metric string) (float64, error)           { return 0, nil }
func (m *mockStore) SetUsage(provider, keyID, metric string, value float64) error       { return nil }
func (m *mockStore) IncrementUsage(provider, keyID, metric string, delta float64) error { return nil }
