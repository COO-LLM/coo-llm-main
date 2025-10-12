package store

import "github.com/user/truckllm/internal/config"

type RuntimeStore interface {
	GetUsage(provider, keyID, metric string) (float64, error)
	SetUsage(provider, keyID, metric string, value float64) error
	IncrementUsage(provider, keyID, metric string, delta float64) error
}

type ConfigStore interface {
	LoadConfig() (*config.Config, error)
	SaveConfig(cfg *config.Config) error
}

type StoreProvider interface {
	RuntimeStore
	ConfigStore
}
