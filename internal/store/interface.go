package store

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/user/coo-llm/internal/config"
)

type MetricPoint struct {
	Value     float64
	Timestamp int64
	Tags      map[string]string
}

type RuntimeStore interface {
	GetUsage(provider, keyID, metric string) (float64, error)
	SetUsage(provider, keyID, metric string, value float64) error
	IncrementUsage(provider, keyID, metric string, delta float64) error
	GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error)
	SetCache(key, value string, ttlSeconds int64) error
	GetCache(key string) (string, error)
	StoreMetric(name string, value float64, tags map[string]string, timestamp int64) error
	GetMetrics(name string, tags map[string]string, start, end int64) ([]MetricPoint, error)
}

type ConfigStore interface {
	LoadConfig() (*config.Config, error)
	SaveConfig(cfg *config.Config) error
}

type ClientStore interface {
	CreateClient(clientID, apiKey, description string, allowedProviders []string) error
	UpdateClient(clientID, description string, allowedProviders []string) error
	DeleteClient(clientID string) error
	GetClient(clientID string) (*ClientInfo, error)
	ListClients() ([]*ClientInfo, error)
	ValidateClient(apiKey string) (*ClientInfo, error)
}

type ClientInfo struct {
	ID               string   `json:"id"`
	APIKey           string   `json:"api_key"`
	Description      string   `json:"description"`
	AllowedProviders []string `json:"allowed_providers"`
	CreatedAt        int64    `json:"created_at"`
	LastUsed         int64    `json:"last_used"`
}

type MetricsStore interface {
	// Enhanced metrics queries
	GetClientMetrics(clientID string, start, end int64) (*ClientMetrics, error)
	GetProviderMetrics(providerID string, start, end int64) (*ProviderMetrics, error)
	GetKeyMetrics(providerID, keyID string, start, end int64) (*KeyMetrics, error)
	GetGlobalMetrics(start, end int64) (*GlobalMetrics, error)

	// Time-series metrics
	GetClientTimeSeries(clientID string, start, end int64, interval string) ([]TimeSeriesPoint, error)
	GetProviderTimeSeries(providerID string, start, end int64, interval string) ([]TimeSeriesPoint, error)
	GetKeyTimeSeries(providerID, keyID string, start, end int64, interval string) ([]TimeSeriesPoint, error)
}

type ClientMetrics struct {
	ClientID        string  `json:"client_id"`
	TotalRequests   int64   `json:"total_requests"`
	TotalTokens     int64   `json:"total_tokens"`
	TotalCost       float64 `json:"total_cost"`
	SuccessRate     float64 `json:"success_rate"`
	AvgLatency      float64 `json:"avg_latency"`
	LastRequestTime int64   `json:"last_request_time"`
}

type ProviderMetrics struct {
	ProviderID    string  `json:"provider_id"`
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
	SuccessRate   float64 `json:"success_rate"`
	AvgLatency    float64 `json:"avg_latency"`
	ErrorCount    int64   `json:"error_count"`
}

type KeyMetrics struct {
	ProviderID    string  `json:"provider_id"`
	KeyID         string  `json:"key_id"`
	TotalRequests int64   `json:"total_requests"`
	TotalTokens   int64   `json:"total_tokens"`
	TotalCost     float64 `json:"total_cost"`
	SuccessRate   float64 `json:"success_rate"`
	AvgLatency    float64 `json:"avg_latency"`
	ErrorCount    int64   `json:"error_count"`
	LastUsed      int64   `json:"last_used"`
}

type GlobalMetrics struct {
	TotalClients       int     `json:"total_clients"`
	TotalProviders     int     `json:"total_providers"`
	TotalRequests      int64   `json:"total_requests"`
	TotalTokens        int64   `json:"total_tokens"`
	TotalCost          float64 `json:"total_cost"`
	OverallSuccessRate float64 `json:"overall_success_rate"`
	AvgLatency         float64 `json:"avg_latency"`
}

type TimeSeriesPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
	Metric    string  `json:"metric"`
}

type AlgorithmStore interface {
	SaveAlgorithmConfig(algorithm string, config map[string]interface{}) error
	LoadAlgorithmConfig(algorithm string) (map[string]interface{}, error)
	ListAlgorithms() ([]string, error)
}

// StoreProviderWrapper wraps existing stores to implement full StoreProvider interface
type StoreProviderWrapper struct {
	RuntimeStore
	ConfigStore
	ClientStore
	MetricsStore
	AlgorithmStore
}

// NewStoreProviderWrapper creates a wrapper with default implementations
func NewStoreProviderWrapper(runtimeStore RuntimeStore, configStore ConfigStore) StoreProvider {
	return &StoreProviderWrapper{
		RuntimeStore:   runtimeStore,
		ConfigStore:    configStore,
		ClientStore:    &DefaultClientStore{runtimeStore: runtimeStore},
		MetricsStore:   &DefaultMetricsStore{runtimeStore: runtimeStore},
		AlgorithmStore: &DefaultAlgorithmStore{runtimeStore: runtimeStore},
	}
}

// SimpleConfigStore provides basic config storage using RuntimeStore
type SimpleConfigStore struct {
	runtimeStore RuntimeStore
}

func NewSimpleConfigStore(runtimeStore RuntimeStore) ConfigStore {
	return &SimpleConfigStore{runtimeStore: runtimeStore}
}

func (s *SimpleConfigStore) LoadConfig() (*config.Config, error) {
	data, err := s.runtimeStore.GetCache("config")
	if err != nil {
		return nil, err
	}
	var cfg config.Config
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (s *SimpleConfigStore) SaveConfig(cfg *config.Config) error {
	data, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	return s.runtimeStore.SetCache("config", string(data), 0)
}

// Default implementations for missing interfaces
type DefaultClientStore struct {
	runtimeStore RuntimeStore
}

func (d *DefaultClientStore) CreateClient(clientID, apiKey, description string, allowedProviders []string) error {
	// Store client info in cache with TTL
	key := "client:" + clientID
	value := fmt.Sprintf("%s|%s|%s", apiKey, description, strings.Join(allowedProviders, ","))
	return d.runtimeStore.SetCache(key, value, 0) // No TTL for persistent clients
}

func (d *DefaultClientStore) UpdateClient(clientID, description string, allowedProviders []string) error {
	key := "client:" + clientID
	// Get existing
	existing, err := d.runtimeStore.GetCache(key)
	if err != nil {
		return err
	}
	parts := strings.Split(existing, "|")
	if len(parts) < 3 {
		return fmt.Errorf("invalid client data")
	}
	apiKey := parts[0]
	value := fmt.Sprintf("%s|%s|%s", apiKey, description, strings.Join(allowedProviders, ","))
	return d.runtimeStore.SetCache(key, value, 0)
}

func (d *DefaultClientStore) DeleteClient(clientID string) error {
	key := "client:" + clientID
	return d.runtimeStore.SetCache(key, "", 0) // Effectively delete
}

func (d *DefaultClientStore) GetClient(clientID string) (*ClientInfo, error) {
	key := "client:" + clientID
	value, err := d.runtimeStore.GetCache(key)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(value, "|")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid client data")
	}
	return &ClientInfo{
		ID:               clientID,
		APIKey:           parts[0],
		Description:      parts[1],
		AllowedProviders: strings.Split(parts[2], ","),
	}, nil
}

func (d *DefaultClientStore) ListClients() ([]*ClientInfo, error) {
	// This is a simplified implementation for cache-based storage
	// In a real database, we'd query all clients
	// For cache-based storage, we'd need key scanning which most caches don't support efficiently
	// For now, return empty list - clients are managed through config
	return []*ClientInfo{}, nil
}

func (d *DefaultClientStore) ValidateClient(apiKey string) (*ClientInfo, error) {
	// In a cache-based system, we need to check if this API key corresponds to any stored client
	// Since we don't have an index, we'll try to find it by checking common patterns
	// This is a simplified implementation - in production, you'd want a proper index

	// For now, return nil to indicate client validation is not implemented
	// The system currently uses config-based API key validation
	return nil, fmt.Errorf("client validation not implemented - use config-based keys")
}

type DefaultMetricsStore struct {
	runtimeStore RuntimeStore
}

func (d *DefaultMetricsStore) GetClientMetrics(clientID string, start, end int64) (*ClientMetrics, error) {
	metrics := &ClientMetrics{ClientID: clientID}

	// Get latency metrics for request count and avg latency
	latencyPoints, err := d.runtimeStore.GetMetrics("latency", map[string]string{"client_key": clientID}, start, end)
	if err == nil {
		metrics.TotalRequests = int64(len(latencyPoints))
		var totalLatency float64
		var successCount int64
		for _, p := range latencyPoints {
			totalLatency += p.Value
			if p.Value > 0 { // Assume positive latency means success
				successCount++
			}
		}
		if metrics.TotalRequests > 0 {
			metrics.AvgLatency = totalLatency / float64(metrics.TotalRequests)
			metrics.SuccessRate = float64(successCount) / float64(metrics.TotalRequests)
		}
	}

	// Get token metrics
	tokenPoints, err := d.runtimeStore.GetMetrics("tokens", map[string]string{"client_key": clientID}, start, end)
	if err == nil {
		for _, p := range tokenPoints {
			metrics.TotalTokens += int64(p.Value)
		}
	}

	// Get cost metrics
	costPoints, err := d.runtimeStore.GetMetrics("cost", map[string]string{"client_key": clientID}, start, end)
	if err == nil {
		for _, p := range costPoints {
			metrics.TotalCost += p.Value
		}
	}

	// Get last request time
	if len(latencyPoints) > 0 {
		metrics.LastRequestTime = latencyPoints[len(latencyPoints)-1].Timestamp
	}

	return metrics, nil
}

func (d *DefaultMetricsStore) GetProviderMetrics(providerID string, start, end int64) (*ProviderMetrics, error) {
	metrics := &ProviderMetrics{ProviderID: providerID}

	// Get latency metrics for request count and avg latency
	latencyPoints, err := d.runtimeStore.GetMetrics("latency", map[string]string{"provider": providerID}, start, end)
	if err == nil {
		metrics.TotalRequests = int64(len(latencyPoints))
		var totalLatency float64
		var successCount int64
		for _, p := range latencyPoints {
			totalLatency += p.Value
			if p.Value > 0 { // Assume positive latency means success
				successCount++
			}
		}
		if metrics.TotalRequests > 0 {
			metrics.AvgLatency = totalLatency / float64(metrics.TotalRequests)
			metrics.SuccessRate = float64(successCount) / float64(metrics.TotalRequests)
		}
	}

	// Get token metrics
	tokenPoints, err := d.runtimeStore.GetMetrics("tokens", map[string]string{"provider": providerID}, start, end)
	if err == nil {
		for _, p := range tokenPoints {
			metrics.TotalTokens += int64(p.Value)
		}
	}

	// Get cost metrics
	costPoints, err := d.runtimeStore.GetMetrics("cost", map[string]string{"provider": providerID}, start, end)
	if err == nil {
		for _, p := range costPoints {
			metrics.TotalCost += p.Value
		}
	}

	// Get error count (negative latency could indicate errors)
	errorPoints, err := d.runtimeStore.GetMetrics("latency", map[string]string{"provider": providerID, "error": "true"}, start, end)
	if err == nil {
		metrics.ErrorCount = int64(len(errorPoints))
	}

	return metrics, nil
}

func (d *DefaultMetricsStore) GetKeyMetrics(providerID, keyID string, start, end int64) (*KeyMetrics, error) {
	metrics := &KeyMetrics{ProviderID: providerID, KeyID: keyID}

	// Get latency metrics for request count and avg latency
	latencyPoints, err := d.runtimeStore.GetMetrics("latency", map[string]string{"provider": providerID, "key": keyID}, start, end)
	if err == nil {
		metrics.TotalRequests = int64(len(latencyPoints))
		var totalLatency float64
		var successCount int64
		for _, p := range latencyPoints {
			totalLatency += p.Value
			if p.Value > 0 { // Assume positive latency means success
				successCount++
			}
		}
		if metrics.TotalRequests > 0 {
			metrics.AvgLatency = totalLatency / float64(metrics.TotalRequests)
			metrics.SuccessRate = float64(successCount) / float64(metrics.TotalRequests)
		}
		if len(latencyPoints) > 0 {
			metrics.LastUsed = latencyPoints[len(latencyPoints)-1].Timestamp
		}
	}

	// Get token metrics
	tokenPoints, err := d.runtimeStore.GetMetrics("tokens", map[string]string{"provider": providerID, "key": keyID}, start, end)
	if err == nil {
		for _, p := range tokenPoints {
			metrics.TotalTokens += int64(p.Value)
		}
	}

	// Get cost metrics
	costPoints, err := d.runtimeStore.GetMetrics("cost", map[string]string{"provider": providerID, "key": keyID}, start, end)
	if err == nil {
		for _, p := range costPoints {
			metrics.TotalCost += p.Value
		}
	}

	// Get error count
	errorPoints, err := d.runtimeStore.GetMetrics("latency", map[string]string{"provider": providerID, "key": keyID, "error": "true"}, start, end)
	if err == nil {
		metrics.ErrorCount = int64(len(errorPoints))
	}

	return metrics, nil
}

func (d *DefaultMetricsStore) GetGlobalMetrics(start, end int64) (*GlobalMetrics, error) {
	metrics := &GlobalMetrics{}

	// Get all latency points to count total requests
	latencyPoints, err := d.runtimeStore.GetMetrics("latency", map[string]string{}, start, end)
	if err == nil {
		metrics.TotalRequests = int64(len(latencyPoints))
		var totalLatency float64
		var successCount int64
		providerSet := make(map[string]bool)
		clientSet := make(map[string]bool)

		for _, p := range latencyPoints {
			totalLatency += p.Value
			if p.Value > 0 {
				successCount++
			}
			if p.Tags["provider"] != "" {
				providerSet[p.Tags["provider"]] = true
			}
			if p.Tags["client_key"] != "" {
				clientSet[p.Tags["client_key"]] = true
			}
		}

		if metrics.TotalRequests > 0 {
			metrics.AvgLatency = totalLatency / float64(metrics.TotalRequests)
			metrics.OverallSuccessRate = float64(successCount) / float64(metrics.TotalRequests)
		}

		metrics.TotalProviders = len(providerSet)
		metrics.TotalClients = len(clientSet)
	}

	// Get all token points
	tokenPoints, err := d.runtimeStore.GetMetrics("tokens", map[string]string{}, start, end)
	if err == nil {
		for _, p := range tokenPoints {
			metrics.TotalTokens += int64(p.Value)
		}
	}

	// Get all cost points
	costPoints, err := d.runtimeStore.GetMetrics("cost", map[string]string{}, start, end)
	if err == nil {
		for _, p := range costPoints {
			metrics.TotalCost += p.Value
		}
	}

	return metrics, nil
}

func (d *DefaultMetricsStore) GetClientTimeSeries(clientID string, start, end int64, interval string) ([]TimeSeriesPoint, error) {
	return []TimeSeriesPoint{}, nil
}

func (d *DefaultMetricsStore) GetProviderTimeSeries(providerID string, start, end int64, interval string) ([]TimeSeriesPoint, error) {
	return []TimeSeriesPoint{}, nil
}

func (d *DefaultMetricsStore) GetKeyTimeSeries(providerID, keyID string, start, end int64, interval string) ([]TimeSeriesPoint, error) {
	return []TimeSeriesPoint{}, nil
}

type DefaultAlgorithmStore struct {
	runtimeStore RuntimeStore
}

func (d *DefaultAlgorithmStore) SaveAlgorithmConfig(algorithm string, config map[string]interface{}) error {
	key := "algorithm:" + algorithm
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	return d.runtimeStore.SetCache(key, string(data), 0)
}

func (d *DefaultAlgorithmStore) LoadAlgorithmConfig(algorithm string) (map[string]interface{}, error) {
	key := "algorithm:" + algorithm
	data, err := d.runtimeStore.GetCache(key)
	if err != nil {
		return nil, err
	}
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(data), &config); err != nil {
		return nil, err
	}
	return config, nil
}

func (d *DefaultAlgorithmStore) ListAlgorithms() ([]string, error) {
	// Simplified - would need to scan keys in real implementation
	return []string{"hybrid", "round_robin", "least_loaded"}, nil
}

type StoreProvider interface {
	RuntimeStore
	ConfigStore
	ClientStore
	MetricsStore
	AlgorithmStore
}
