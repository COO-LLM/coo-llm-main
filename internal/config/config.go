package config

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type APIKeyConfig struct {
	Key              string   `yaml:"key" mapstructure:"key"`
	AllowedProviders []string `yaml:"allowed_providers" mapstructure:"allowed_providers"` // ["openai", "gemini", "*"] - "*" means all
	Description      string   `yaml:"description,omitempty" mapstructure:"description,omitempty"`
}

type Config struct {
	Version      string            `yaml:"version" mapstructure:"version"`
	Server       Server            `yaml:"server" mapstructure:"server"`
	Logging      Logging           `yaml:"logging" mapstructure:"logging"`
	Storage      Storage           `yaml:"storage" mapstructure:"storage"`
	LLMProviders []LLMProvider     `yaml:"llm_providers" mapstructure:"llm_providers"`
	Providers    []Provider        `yaml:"providers" mapstructure:"providers"` // Legacy
	APIKeys      []APIKeyConfig    `yaml:"api_keys" mapstructure:"api_keys"`
	ModelAliases map[string]string `yaml:"model_aliases" mapstructure:"model_aliases"`
	Policy       Policy            `yaml:"policy" mapstructure:"policy"`
}

type LLMProvider struct {
	ID      string   `yaml:"id" mapstructure:"id"`
	Type    string   `yaml:"type" mapstructure:"type"`
	APIKeys []string `yaml:"api_keys" mapstructure:"api_keys"`
	BaseURL string   `yaml:"base_url,omitempty" mapstructure:"base_url,omitempty"`
	Model   string   `yaml:"model" mapstructure:"model"`
	Pricing Pricing  `yaml:"pricing" mapstructure:"pricing"`
	Limits  Limits   `yaml:"limits" mapstructure:"limits"`
}

type Limits struct {
	ReqPerMin    int `yaml:"req_per_min" mapstructure:"req_per_min"`
	TokensPerMin int `yaml:"tokens_per_min" mapstructure:"tokens_per_min"`
}

type Server struct {
	Listen      string `yaml:"listen" mapstructure:"listen"`
	AdminAPIKey string `yaml:"admin_api_key" mapstructure:"admin_api_key"`
}

type Logging struct {
	File       FileLog       `yaml:"file" mapstructure:"file"`
	Prometheus PrometheusLog `yaml:"prometheus" mapstructure:"prometheus"`
	Providers  []LogProvider `yaml:"providers" mapstructure:"providers"`
}

type FileLog struct {
	Enabled    bool   `yaml:"enabled" mapstructure:"enabled"`
	Path       string `yaml:"path" mapstructure:"path"`
	MaxSizeMB  int    `yaml:"max_size_mb" mapstructure:"max_size_mb"`
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"`
}

type PrometheusLog struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	Endpoint string `yaml:"endpoint" mapstructure:"endpoint"`
}

type LogProvider struct {
	Name     string            `yaml:"name" mapstructure:"name"`
	Type     string            `yaml:"type" mapstructure:"type"`
	Endpoint string            `yaml:"endpoint" mapstructure:"endpoint"`
	Batch    BatchConfig       `yaml:"batch" mapstructure:"batch"`
	Headers  map[string]string `yaml:"headers" mapstructure:"headers"`
}

type BatchConfig struct {
	Enabled         bool `yaml:"enabled" mapstructure:"enabled"`
	Size            int  `yaml:"size" mapstructure:"size"`
	IntervalSeconds int  `yaml:"interval_seconds" mapstructure:"interval_seconds"`
}

type Storage struct {
	Config  ConfigStore  `yaml:"config" mapstructure:"config"`
	Runtime RuntimeStore `yaml:"runtime" mapstructure:"runtime"`
}

type ConfigStore struct {
	Type string `yaml:"type" mapstructure:"type"`
	Path string `yaml:"path" mapstructure:"path"`
}

type RuntimeStore struct {
	Type     string `yaml:"type" mapstructure:"type"`
	Addr     string `yaml:"addr" mapstructure:"addr"`
	Password string `yaml:"password" mapstructure:"password"`
	APIKey   string `yaml:"api_key" mapstructure:"api_key"`
}

type Provider struct {
	ID      string `yaml:"id" mapstructure:"id"`
	Name    string `yaml:"name" mapstructure:"name"`
	BaseURL string `yaml:"base_url" mapstructure:"base_url"`
	Keys    []Key  `yaml:"keys" mapstructure:"keys"`
}

type Key struct {
	ID                string  `yaml:"id" mapstructure:"id"`
	Secret            string  `yaml:"secret" mapstructure:"secret"`
	LimitReqPerMin    int     `yaml:"limit_req_per_min" mapstructure:"limit_req_per_min"`
	LimitTokensPerMin int     `yaml:"limit_tokens_per_min" mapstructure:"limit_tokens_per_min"`
	Pricing           Pricing `yaml:"pricing" mapstructure:"pricing"`
}

type Pricing struct {
	InputTokenCost  float64 `yaml:"input_token_cost" mapstructure:"input_token_cost"`
	OutputTokenCost float64 `yaml:"output_token_cost" mapstructure:"output_token_cost"`
}

type Policy struct {
	Strategy      string        `yaml:"strategy" mapstructure:"strategy"`
	Algorithm     string        `yaml:"algorithm" mapstructure:"algorithm"` // "round_robin", "least_loaded", "hybrid"
	Priority      string        `yaml:"priority" mapstructure:"priority"`   // "balanced", "cost", "req", "token"
	HybridWeights HybridWeights `yaml:"hybrid_weights" mapstructure:"hybrid_weights"`
	Retry         RetryConfig   `yaml:"retry" mapstructure:"retry"`
	Cache         CacheConfig   `yaml:"cache" mapstructure:"cache"`
}

type CacheConfig struct {
	Enabled    bool  `yaml:"enabled" mapstructure:"enabled"`
	TTLSeconds int64 `yaml:"ttl_seconds" mapstructure:"ttl_seconds"` // Cache TTL
}

type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts" mapstructure:"max_attempts"` // Max retry attempts
	Timeout     time.Duration `yaml:"timeout" mapstructure:"timeout"`           // Timeout per attempt
	Interval    time.Duration `yaml:"interval" mapstructure:"interval"`         // Interval between retries
}

type HybridWeights struct {
	TokenRatio float64 `yaml:"token_ratio" mapstructure:"token_ratio"`
	ReqRatio   float64 `yaml:"req_ratio" mapstructure:"req_ratio"`
	ErrorScore float64 `yaml:"error_score" mapstructure:"error_score"`
	Latency    float64 `yaml:"latency" mapstructure:"latency"`
	CostRatio  float64 `yaml:"cost_ratio" mapstructure:"cost_ratio"`
}

func LoadConfig(path string) (*Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", path)
	}

	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Set weights based on priority
	if cfg.Policy.Priority != "" {
		switch cfg.Policy.Priority {
		case "cost":
			cfg.Policy.HybridWeights = HybridWeights{
				TokenRatio: 0.1, ReqRatio: 0.1, ErrorScore: 0.2, Latency: 0.1, CostRatio: 0.5,
			}
		case "req":
			cfg.Policy.HybridWeights = HybridWeights{
				TokenRatio: 0.1, ReqRatio: 0.5, ErrorScore: 0.2, Latency: 0.1, CostRatio: 0.1,
			}
		case "token":
			cfg.Policy.HybridWeights = HybridWeights{
				TokenRatio: 0.5, ReqRatio: 0.1, ErrorScore: 0.2, Latency: 0.1, CostRatio: 0.1,
			}
		case "balanced":
			fallthrough
		default:
			cfg.Policy.HybridWeights = HybridWeights{
				TokenRatio: 0.2, ReqRatio: 0.2, ErrorScore: 0.2, Latency: 0.2, CostRatio: 0.2,
			}
		}
	}

	// Convert LLMProviders to legacy Providers for backward compatibility
	if len(cfg.Providers) == 0 && len(cfg.LLMProviders) > 0 {
		for _, lp := range cfg.LLMProviders {
			keys := make([]Key, len(lp.APIKeys))
			for i, apiKey := range lp.APIKeys {
				// Use hash of API key as stable ID for consistency
				h := sha256.Sum256([]byte(apiKey))
				keyID := fmt.Sprintf("%s-%x", lp.Type, h[:8])
				keys[i] = Key{
					ID:                keyID,
					Secret:            apiKey,
					LimitReqPerMin:    lp.Limits.ReqPerMin,
					LimitTokensPerMin: lp.Limits.TokensPerMin,
					Pricing:           lp.Pricing,
				}
			}
			cfg.Providers = append(cfg.Providers, Provider{
				ID:      lp.Type,
				Name:    lp.Type,
				BaseURL: lp.BaseURL,
				Keys:    keys,
			})
		}
	}

	return &cfg, nil
}

func ValidateConfig(cfg *Config) error {
	if cfg.Version == "" {
		return fmt.Errorf("version is required")
	}
	if cfg.Server.Listen == "" {
		return fmt.Errorf("server.listen is required")
	}
	if len(cfg.LLMProviders) == 0 && len(cfg.Providers) == 0 {
		return fmt.Errorf("at least one provider is required")
	}
	// Add more validations as needed
	return nil
}

func SaveConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
