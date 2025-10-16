package store

import (
	"sync"

	"github.com/rs/zerolog"
)

type MemoryStore struct {
	data   map[string]float64
	cache  map[string]cacheEntry
	logger zerolog.Logger
	mu     sync.RWMutex
}

type cacheEntry struct {
	value  string
	expiry int64
}

func NewMemoryStore(logger zerolog.Logger) *MemoryStore {
	return &MemoryStore{
		data:   make(map[string]float64),
		cache:  make(map[string]cacheEntry),
		logger: logger,
	}
}

func (m *MemoryStore) key(provider, keyID, metric string) string {
	return provider + ":" + keyID + ":" + metric
}

func (m *MemoryStore) GetUsage(provider, keyID, metric string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	value := m.data[m.key(provider, keyID, metric)]
	m.logger.Debug().Str("operation", "GetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation")
	return value, nil
}

func (m *MemoryStore) SetUsage(provider, keyID, metric string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[m.key(provider, keyID, metric)] = value
	m.logger.Debug().Str("operation", "SetUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("value", value).Msg("store operation")
	return nil
}

func (m *MemoryStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.key(provider, keyID, metric)
	oldValue := m.data[key]
	m.data[key] += delta
	m.logger.Debug().Str("operation", "IncrementUsage").Str("provider", provider).Str("keyID", keyID).Str("metric", metric).Float64("delta", delta).Float64("old_value", oldValue).Float64("new_value", m.data[key]).Msg("store operation")
	return nil
}

func (m *MemoryStore) GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error) {
	// Memory store doesn't support time windows, return total
	return m.GetUsage(provider, keyID, metric)
}

func (m *MemoryStore) SetCache(key, value string, ttlSeconds int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Simple cache without TTL for memory
	m.cache[key] = cacheEntry{value: value, expiry: 0}
	return nil
}

func (m *MemoryStore) GetCache(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if entry, ok := m.cache[key]; ok {
		return entry.value, nil
	}
	return "", nil // Not found
}
