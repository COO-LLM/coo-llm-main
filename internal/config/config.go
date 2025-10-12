package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Version      string            `yaml:"version" mapstructure:"version"`
	Server       Server            `yaml:"server" mapstructure:"server"`
	Logging      Logging           `yaml:"logging" mapstructure:"logging"`
	Storage      Storage           `yaml:"storage" mapstructure:"storage"`
	Providers    []Provider        `yaml:"providers" mapstructure:"providers"`
	ModelAliases map[string]string `yaml:"model_aliases" mapstructure:"model_aliases"`
	Policy       Policy            `yaml:"policy" mapstructure:"policy"`
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
	InputTokenCost  float64 `yaml:"input_token_cost" mapstructure:"input_token_cost"`   // per 1K tokens
	OutputTokenCost float64 `yaml:"output_token_cost" mapstructure:"output_token_cost"` // per 1K tokens
	Currency        string  `yaml:"currency" mapstructure:"currency"`
}

type Policy struct {
	Strategy      string        `yaml:"strategy" mapstructure:"strategy"`
	HybridWeights HybridWeights `yaml:"hybrid_weights" mapstructure:"hybrid_weights"`
	CostFirst     bool          `yaml:"cost_first" mapstructure:"cost_first"`
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

	return &cfg, nil
}

func ValidateConfig(cfg *Config) error {
	if cfg.Version == "" {
		return fmt.Errorf("version is required")
	}
	if cfg.Server.Listen == "" {
		return fmt.Errorf("server.listen is required")
	}
	if len(cfg.Providers) == 0 {
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
