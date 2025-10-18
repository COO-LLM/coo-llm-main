package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type APIKeyConfig struct {
	ID               string   `yaml:"id,omitempty" mapstructure:"id,omitempty"`
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
	Name    string   `yaml:"name,omitempty" mapstructure:"name,omitempty"`
	Type    string   `yaml:"type" mapstructure:"type"`
	APIKeys []string `yaml:"api_keys" mapstructure:"api_keys"`
	BaseURL string   `yaml:"base_url,omitempty" mapstructure:"base_url,omitempty"`
	Model   string   `yaml:"model" mapstructure:"model"`
	Pricing Pricing  `yaml:"pricing" mapstructure:"pricing"`
	Limits  Limits   `yaml:"limits" mapstructure:"limits"`
}

type Limits struct {
	ReqPerMin    int    `yaml:"req_per_min" mapstructure:"req_per_min"`
	TokensPerMin int    `yaml:"tokens_per_min" mapstructure:"tokens_per_min"`
	MaxTokens    int    `yaml:"max_tokens" mapstructure:"max_tokens"`
	SessionLimit int    `yaml:"session_limit" mapstructure:"session_limit"`
	SessionType  string `yaml:"session_type" mapstructure:"session_type"`
}

type Server struct {
	Listen      string `yaml:"listen" mapstructure:"listen"`
	AdminAPIKey string `yaml:"admin_api_key" mapstructure:"admin_api_key"`
	WebUI       WebUI  `yaml:"webui" mapstructure:"webui"`
	CORS        CORS   `yaml:"cors" mapstructure:"cors"`
}

type WebUI struct {
	Enabled       bool   `yaml:"enabled" mapstructure:"enabled"`
	AdminID       string `yaml:"admin_id" mapstructure:"admin_id"`
	AdminPassword string `yaml:"admin_password" mapstructure:"admin_password"`
	WebUIPath     string `yaml:"web_ui_path,omitempty" mapstructure:"web_ui_path,omitempty"`
}

type CORS struct {
	Enabled          bool     `yaml:"enabled" mapstructure:"enabled"`
	AllowedOrigins   []string `yaml:"allowed_origins" mapstructure:"allowed_origins"`
	AllowedMethods   []string `yaml:"allowed_methods" mapstructure:"allowed_methods"`
	AllowedHeaders   []string `yaml:"allowed_headers" mapstructure:"allowed_headers"`
	AllowCredentials bool     `yaml:"allow_credentials" mapstructure:"allow_credentials"`
	MaxAge           int      `yaml:"max_age" mapstructure:"max_age"`
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
	ID      string  `yaml:"id" mapstructure:"id"`
	Name    string  `yaml:"name,omitempty" mapstructure:"name,omitempty"`
	BaseURL string  `yaml:"base_url" mapstructure:"base_url"`
	Keys    []Key   `yaml:"keys" mapstructure:"keys"`
	Limits  Limits  `yaml:"limits" mapstructure:"limits"`
	Pricing Pricing `yaml:"pricing" mapstructure:"pricing"`
}

type Key struct {
	ID                string `yaml:"id" mapstructure:"id"`
	Secret            string `yaml:"secret" mapstructure:"secret"`
	LimitReqPerMin    int    `yaml:"limit_req_per_min" mapstructure:"limit_req_per_min"`
	LimitTokensPerMin int    `yaml:"limit_tokens_per_min" mapstructure:"limit_tokens_per_min"`
	SessionLimit      int    `yaml:"session_limit" mapstructure:"session_limit"`
	SessionType       string `yaml:"session_type" mapstructure:"session_type"`
}

type Pricing struct {
	InputTokenCost  float64 `yaml:"input_token_cost" mapstructure:"input_token_cost"`
	OutputTokenCost float64 `yaml:"output_token_cost" mapstructure:"output_token_cost"`
}

type Policy struct {
	Strategy      string         `yaml:"strategy" mapstructure:"strategy"`
	Algorithm     string         `yaml:"algorithm" mapstructure:"algorithm"` // "round_robin", "least_loaded", "hybrid"
	Priority      string         `yaml:"priority" mapstructure:"priority"`   // "balanced", "cost", "req", "token"
	HybridWeights HybridWeights  `yaml:"hybrid_weights" mapstructure:"hybrid_weights"`
	Retry         RetryConfig    `yaml:"retry" mapstructure:"retry"`
	Fallback      FallbackConfig `yaml:"fallback" mapstructure:"fallback"`
	Cache         CacheConfig    `yaml:"cache" mapstructure:"cache"`
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

type FallbackConfig struct {
	Enabled      bool     `yaml:"enabled" mapstructure:"enabled"`             // Enable fallback to other providers
	MaxProviders int      `yaml:"max_providers" mapstructure:"max_providers"` // Max fallback providers to try
	Providers    []string `yaml:"providers" mapstructure:"providers"`         // List of fallback provider IDs
}

type HybridWeights struct {
	TokenRatio float64 `yaml:"token_ratio" mapstructure:"token_ratio"`
	ReqRatio   float64 `yaml:"req_ratio" mapstructure:"req_ratio"`
	ErrorScore float64 `yaml:"error_score" mapstructure:"error_score"`
	Latency    float64 `yaml:"latency" mapstructure:"latency"`
	CostRatio  float64 `yaml:"cost_ratio" mapstructure:"cost_ratio"`
}

// expandEnvVars expands ${VAR} placeholders in string values
func expandEnvVars(data map[string]any) {
	for key, value := range data {
		switch v := value.(type) {
		case string:
			data[key] = os.ExpandEnv(v)
		case map[string]any:
			expandEnvVars(v)
		case []any:
			for _, item := range v {
				if itemMap, ok := item.(map[string]any); ok {
					expandEnvVars(itemMap)
				}
			}
		}
	}
}

// LoadConfig loads config from file or environment variables
func LoadConfig(path string) (*Config, error) {
	var cfg Config

	if path != "" {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, fmt.Errorf("config file does not exist: %s", path)
		}

		viper.SetConfigFile(path)
		viper.SetConfigType("yaml")
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

		if err := viper.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config: %w", err)
		}

		// Unmarshal directly to struct (keep env var placeholders)
		if err := viper.Unmarshal(&cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
	} else {
		// Load from environment variables only
		loadFromEnv(&cfg)
	}

	return &cfg, nil
}

// loadFromEnv loads config from environment variables only
func loadFromEnv(cfg *Config) {
	// Set defaults
	cfg.Version = "1.0"
	cfg.Server.Listen = ":2906"
	cfg.Server.AdminAPIKey = os.Getenv("COO__ADMIN_API_KEY")
	if cfg.Server.AdminAPIKey == "" {
		cfg.Server.AdminAPIKey = os.Getenv("ADMIN_API_KEY")
	}
	cfg.Server.CORS.Enabled = true
	cfg.Server.CORS.AllowedOrigins = []string{"*"}
	cfg.Server.CORS.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	cfg.Server.CORS.AllowedHeaders = []string{"*"}
	cfg.Server.CORS.AllowCredentials = true
	cfg.Server.CORS.MaxAge = 86400
	cfg.Logging.File.Enabled = true
	cfg.Logging.File.Path = "./logs/llm.log"
	cfg.Logging.File.MaxSizeMB = 100
	cfg.Logging.File.MaxBackups = 5
	cfg.Logging.Prometheus.Enabled = true
	cfg.Logging.Prometheus.Endpoint = "/metrics"
	cfg.Storage.Runtime.Type = "sql"
	cfg.Storage.Runtime.Addr = "./data/coo-llm.db"
	cfg.Storage.Runtime.Database = "coo_llm"
	cfg.Storage.Runtime.TableUsage = "coo_llm_usage"
	cfg.Storage.Runtime.TableCache = "coo_llm_cache"
	cfg.Storage.Runtime.TableHistory = "coo_llm_history"
	cfg.Policy.Strategy = "hybrid"
	cfg.Policy.Algorithm = "hybrid"
	cfg.Policy.Priority = "balanced"
	cfg.Policy.HybridWeights.TokenRatio = 0.2
	cfg.Policy.HybridWeights.ReqRatio = 0.2
	cfg.Policy.HybridWeights.ErrorScore = 0.2
	cfg.Policy.HybridWeights.Latency = 0.2
	cfg.Policy.HybridWeights.CostRatio = 0.2
	cfg.Policy.Retry.MaxAttempts = 3
	cfg.Policy.Retry.Timeout = 30 * time.Second
	cfg.Policy.Retry.Interval = 1 * time.Second
	cfg.Policy.Fallback.Enabled = true
	cfg.Policy.Fallback.MaxProviders = 2
	cfg.Policy.Cache.Enabled = true
	cfg.Policy.Cache.TTLSeconds = 10
}

// SaveConfigToFile saves config to a file (with sensitive data sanitized)
func SaveConfigToFile(cfg *Config, path string) error {
	// Sanitize config before saving
	safeCfg := sanitizeConfigForFile(cfg)

	// Marshal to YAML
	yamlData, err := yaml.Marshal(safeCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, yamlData, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// sanitizeConfigForFile removes sensitive data before saving to file
func sanitizeConfigForFile(cfg *Config) *Config {
	safeCfg := *cfg // Shallow copy

	// Replace sensitive server config with env var placeholders
	if cfg.Server.AdminAPIKey != "" && !strings.HasPrefix(cfg.Server.AdminAPIKey, "${") {
		safeCfg.Server.AdminAPIKey = "${COO__ADMIN_API_KEY}"
	}

	// Replace API keys with env var placeholders
	for idx := range safeCfg.APIKeys {
		if safeCfg.APIKeys[idx].Key != "" && !strings.HasPrefix(safeCfg.APIKeys[idx].Key, "${") {
			safeCfg.APIKeys[idx].Key = "${API_KEY_" + fmt.Sprintf("%d", idx) + "}"
		}
	}

	// Replace provider API keys with env var placeholders
	for i := range safeCfg.LLMProviders {
		for j := range safeCfg.LLMProviders[i].APIKeys {
			if safeCfg.LLMProviders[i].APIKeys[j] != "" && !strings.HasPrefix(safeCfg.LLMProviders[i].APIKeys[j], "${") {
				envName := strings.ToUpper(safeCfg.LLMProviders[i].ID) + "_API_KEY"
				if j > 0 {
					envName += fmt.Sprintf("_%d", j)
				}
				safeCfg.LLMProviders[i].APIKeys[j] = "${" + envName + "}"
			}
		}
	}

	return &safeCfg
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

// MaskSensitiveConfig creates a copy of config with sensitive data removed
func MaskSensitiveConfig(cfg *Config) *Config {
	safeCfg := *cfg

	// Remove sensitive providers entirely (they contain API keys)
	safeCfg.LLMProviders = nil

	// Mask admin API key
	if len(safeCfg.Server.AdminAPIKey) > 4 {
		safeCfg.Server.AdminAPIKey = safeCfg.Server.AdminAPIKey[:4] + "****"
	}

	// Mask storage passwords/api keys
	if safeCfg.Storage.Runtime.Password != "" && len(safeCfg.Storage.Runtime.Password) > 4 {
		safeCfg.Storage.Runtime.Password = safeCfg.Storage.Runtime.Password[:4] + "****"
	}
	if safeCfg.Storage.Runtime.APIKey != "" && len(safeCfg.Storage.Runtime.APIKey) > 4 {
		safeCfg.Storage.Runtime.APIKey = safeCfg.Storage.Runtime.APIKey[:4] + "****"
	}

	return &safeCfg
}

func SaveConfig(cfg *Config, path string) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return os.WriteFile(path, data, 0644)
}
