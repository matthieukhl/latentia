package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	LLM    LLMConfig    `mapstructure:"llm"`
	Ingest IngestConfig `mapstructure:"ingest"`
	Safety SafetyConfig `mapstructure:"safety"`
	Vector VectorConfig `mapstructure:"vector"`
}

type ServerConfig struct {
	Addr string `mapstructure:"addr"`
}

type DBConfig struct {
	DSN          string `mapstructure:"dsn"`
	MaxOpenConns int    `mapstructure:"maxOpenConns"`
}

type LLMConfig struct {
	Embedder  ProviderConfig `mapstructure:"embedder"`
	Generator ProviderConfig `mapstructure:"generator"`
}

type ProviderConfig struct {
	Provider  string `mapstructure:"provider"`
	Model     string `mapstructure:"model"`
	APIKeyEnv string `mapstructure:"api_key_env"`
	APIKey    string `mapstructure:"api_key"`
}

type IngestConfig struct {
	SlowQueryInterval time.Duration `mapstructure:"slowquery_interval"`
	Docs             DocsConfig    `mapstructure:"docs"`
}

type DocsConfig struct {
	Sources    []SourceConfig `mapstructure:"sources"`
	OCREnabled bool          `mapstructure:"ocr_enabled"`
}

type SourceConfig struct {
	Type string `mapstructure:"type"`
	URL  string `mapstructure:"url"`
}

type SafetyConfig struct {
	MaxStmtSeconds   int      `mapstructure:"max_stmt_seconds"`
	ForbidPatterns   []string `mapstructure:"forbid_patterns"`
}

type VectorConfig struct {
	Dim  int `mapstructure:"dim"`
	TopK int `mapstructure:"top_k"`
}

// LoadConfig loads configuration from config.yaml and environment variables
func LoadConfig() (*Config, error) {
	v := viper.New()
	
	// Set config file locations
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./deploy/")
	v.AddConfigPath("./")
	v.AddConfigPath("$HOME/.latentia/")
	v.AddConfigPath("/etc/latentia/")
	
	// Enable environment variable override with LATENTIA_ prefix
	v.SetEnvPrefix("LATENTIA")
	v.AutomaticEnv()
	
	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	return &config, nil
}