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
	Type         string `yaml:"type" mapstructure:"type"`
	Addr         string `yaml:"addr" mapstructure:"addr"`
	Password     string `yaml:"password" mapstructure:"password"`
	APIKey       string `yaml:"api_key" mapstructure:"api_key"`
	Database     string `yaml:"database" mapstructure:"database"`
	TableUsage   string `yaml:"table_usage" mapstructure:"table_usage"`
	TableCache   string `yaml:"table_cache" mapstructure:"table_cache"`
	TableHistory string `yaml:"table_history" mapstructure:"table_history"`
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
	Enabled             bool    `yaml:"enabled" mapstructure:"enabled"`
	TTLSeconds          int64   `yaml:"ttl_seconds" mapstructure:"ttl_seconds"` // Cache TTL
	SemanticEnabled     bool    `yaml:"semantic_enabled" mapstructure:"semantic_enabled"`
	EmbeddingModel      string  `yaml:"embedding_model" mapstructure:"embedding_model"`
	SimilarityThreshold float64 `yaml:"similarity_threshold" mapstructure:"similarity_threshold"`
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

// expandEnvVars expands ${VAR} placeholders in string values
func expandEnvVars(data map[string]interface{}) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			data[key] = os.ExpandEnv(v)
		case map[string]interface{}:
			expandEnvVars(v)
		case []interface{}:
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					expandEnvVars(itemMap)
				} else if itemStr, ok := item.(string); ok {
					v[i] = os.ExpandEnv(itemStr)
				}
			}
		}
	}
}

func LoadConfig(path string) (*Config, error) {
	var cfg Config

	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file does not exist: %s", path)
		}

		viper.SetConfigFile(path)
		viper.SetConfigType("yaml")

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}

		// Get raw config data
		var rawData map[string]interface{}
		if err := viper.Unmarshal(&rawData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw config: %w", err)
		}

		// Expand environment variables
		expandEnvVars(rawData)

		// Marshal back to YAML and unmarshal to struct
		yamlData, err := yaml.Marshal(rawData)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal expanded config: %w", err)
		}

		if err := yaml.Unmarshal(yamlData, &cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal expanded config: %w", err)
		}
	} else {
		// Load from environment variables only
		viper.AutomaticEnv()

		// Set defaults
		cfg.Version = "1.0"
		cfg.Server.Listen = ":2906"
		cfg.Server.AdminAPIKey = os.Getenv("COO__ADMIN_API_KEY")
		cfg.Logging.File.Enabled = true
		cfg.Logging.File.Path = "./logs/llm.log"
		cfg.Logging.File.MaxSizeMB = 100
		cfg.Logging.File.MaxBackups = 5
		cfg.Logging.Prometheus.Enabled = true
		cfg.Logging.Prometheus.Endpoint = "/metrics"
		cfg.Storage.Config.Type = "file"
		cfg.Storage.Config.Path = "./data/config.json"
		cfg.Storage.Runtime.Type = "memory"
		cfg.Storage.Runtime.Database = "coo_llm"
		cfg.Storage.Runtime.TableUsage = "coo_llm_usage"
		cfg.Storage.Runtime.TableCache = "coo_llm_cache"
		cfg.Storage.Runtime.TableHistory = "coo_llm_history"
		cfg.Policy.Strategy = "hybrid"
		cfg.Policy.Algorithm = "hybrid"
		cfg.Policy.Priority = "balanced"
		cfg.Policy.Retry.MaxAttempts = 3
		cfg.Policy.Retry.Timeout = 30 * time.Second
		cfg.Policy.Retry.Interval = 1 * time.Second
		cfg.Policy.Cache.Enabled = true
		cfg.Policy.Cache.TTLSeconds = 10
		cfg.Policy.Cache.SemanticEnabled = false
		cfg.Policy.Cache.EmbeddingModel = "text-embedding-ada-002"
		cfg.Policy.Cache.SimilarityThreshold = 0.9

		if err := viper.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal from env: %w", err)
		}

		// Build providers from env vars with COO__ prefix
		if openaiKey := os.Getenv("COO__OPENAI_API_KEY"); openaiKey != "" {
			cfg.LLMProviders = append(cfg.LLMProviders, LLMProvider{
				ID:      "openai",
				Type:    "openai",
				APIKeys: []string{openaiKey},
				Model:   os.Getenv("COO_LLM_OPENAI_MODEL"),
			})
			if cfg.LLMProviders[len(cfg.LLMProviders)-1].Model == "" {
				cfg.LLMProviders[len(cfg.LLMProviders)-1].Model = "gpt-4o"
			}
		}
		if geminiKey := os.Getenv("COO__GEMINI_API_KEY"); geminiKey != "" {
			cfg.LLMProviders = append(cfg.LLMProviders, LLMProvider{
				ID:      "gemini",
				Type:    "gemini",
				APIKeys: []string{geminiKey},
				Model:   os.Getenv("COO__GEMINI_MODEL"),
			})
			if cfg.LLMProviders[len(cfg.LLMProviders)-1].Model == "" {
				cfg.LLMProviders[len(cfg.LLMProviders)-1].Model = "gemini-1.5-pro"
			}
		}
		if claudeKey := os.Getenv("COO__CLAUDE_API_KEY"); claudeKey != "" {
			cfg.LLMProviders = append(cfg.LLMProviders, LLMProvider{
				ID:      "claude",
				Type:    "claude",
				APIKeys: []string{claudeKey},
				BaseURL: os.Getenv("COO__CLAUDE_BASE_URL"),
				Model:   os.Getenv("COO__CLAUDE_MODEL"),
			})
			if cfg.LLMProviders[len(cfg.LLMProviders)-1].Model == "" {
				cfg.LLMProviders[len(cfg.LLMProviders)-1].Model = "claude-3-opus-20240229"
			}
			if cfg.LLMProviders[len(cfg.LLMProviders)-1].BaseURL == "" {
				cfg.LLMProviders[len(cfg.LLMProviders)-1].BaseURL = "https://api.anthropic.com"
			}
		}

		// For testing, add dummy if no providers
		if len(cfg.LLMProviders) == 0 {
			cfg.LLMProviders = append(cfg.LLMProviders, LLMProvider{
				ID:      "dummy",
				Type:    "openai",
				APIKeys: []string{"dummy"},
				Model:   "gpt-4o",
			})
		}
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
