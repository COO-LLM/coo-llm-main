package store

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfluxDBStore(t *testing.T) {
	// Skip if InfluxDB not available
	t.Skip("InfluxDB tests require running InfluxDB instance")

	// Mock InfluxDB connection for testing
	logger := zerolog.New(nil)
	store := NewInfluxDBStore("http://localhost:8086", "token", "org", "bucket", logger)

	// Test StoreMetric
	err := store.StoreMetric("test_metric", 42.0, map[string]string{"provider": "openai"}, 1234567890)
	require.NoError(t, err)

	// Test GetMetrics
	points, err := store.GetMetrics("test_metric", map[string]string{"provider": "openai"}, 1234567800, 1234567900)
	require.NoError(t, err)
	assert.Len(t, points, 1)
	assert.Equal(t, 42.0, points[0].Value)
	assert.Equal(t, int64(1234567890), points[0].Timestamp)
}
