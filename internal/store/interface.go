package store

import "github.com/user/coo-llm/internal/config"

type RuntimeStore interface {
	GetUsage(provider, keyID, metric string) (float64, error)
	SetUsage(provider, keyID, metric string, value float64) error
	IncrementUsage(provider, keyID, metric string, delta float64) error
	GetUsageInWindow(provider, keyID, metric string, windowSeconds int64) (float64, error)
	SetCache(key, value string, ttlSeconds int64) error
	GetCache(key string) (string, error)
}

type ConfigStore interface {
	LoadConfig() (*config.Config, error)
	SaveConfig(cfg *config.Config) error
}

type StoreProvider interface {
	RuntimeStore
	ConfigStore
}
