package store

import (
	"sync"
)

type MemoryStore struct {
	data  map[string]float64
	cache map[string]cacheEntry
	mu    sync.RWMutex
}

type cacheEntry struct {
	value  string
	expiry int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data:  make(map[string]float64),
		cache: make(map[string]cacheEntry),
	}
}

func (m *MemoryStore) key(provider, keyID, metric string) string {
	return provider + ":" + keyID + ":" + metric
}

func (m *MemoryStore) GetUsage(provider, keyID, metric string) (float64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.data[m.key(provider, keyID, metric)], nil
}

func (m *MemoryStore) SetUsage(provider, keyID, metric string, value float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[m.key(provider, keyID, metric)] = value
	return nil
}

func (m *MemoryStore) IncrementUsage(provider, keyID, metric string, delta float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := m.key(provider, keyID, metric)
	m.data[key] += delta
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
