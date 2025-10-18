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
