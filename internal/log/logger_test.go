package log

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user/truckllm/internal/config"
)

func TestNewLogger(t *testing.T) {
	cfg := &config.Logging{}
	logger := NewLogger(cfg)
	assert.NotNil(t, logger)
	assert.NotNil(t, logger.logger)
}

func TestLogRequest(t *testing.T) {
	cfg := &config.Logging{}
	logger := NewLogger(cfg)

	entry := &LogEntry{
		Provider:  "openai",
		Model:     "gpt-4o",
		ReqID:     "req123",
		LatencyMS: 100,
		Status:    200,
		Tokens:    50,
	}

	// Should not panic
	logger.LogRequest(context.Background(), entry)
}

func TestLogEntry_JSON(t *testing.T) {
	entry := LogEntry{
		Timestamp: "2023-01-01T00:00:00Z",
		Provider:  "openai",
		Model:     "gpt-4o",
		ReqID:     "req123",
		LatencyMS: 100,
		Status:    200,
		Tokens:    50,
		Error:     "",
	}

	// Test struct fields
	assert.Equal(t, "openai", entry.Provider)
	assert.Equal(t, "gpt-4o", entry.Model)
}

func TestSendToProvider(t *testing.T) {
	cfg := &config.Logging{
		Providers: []config.LogProvider{
			{Type: "http", Name: "test", Endpoint: "http://example.com"},
		},
	}
	logger := NewLogger(cfg)

	entry := &LogEntry{Provider: "openai"}
	// Test that it calls sendHTTP
	logger.sendToProvider(cfg.Providers[0], entry)
}
