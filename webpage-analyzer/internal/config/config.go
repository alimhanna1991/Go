package config

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config contains runtime configuration loaded from YAML.
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	HTTPClient    HTTPClientConfig    `yaml:"http_client"`
	Browser       BrowserConfig       `yaml:"browser"`
	Logging       LoggingConfig       `yaml:"logging"`
	Cache         CacheConfig         `yaml:"cache"`
	TemplatePaths TemplatePathsConfig `yaml:"template_paths"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type HTTPClientConfig struct {
	TimeoutSeconds     int  `yaml:"timeout_seconds"`
	MaxRedirects       int  `yaml:"max_redirects"`
	InsecureSkipVerify bool `yaml:"insecure_skip_verify"`
}

type BrowserConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Command        string `yaml:"command"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}

type LoggingConfig struct {
	Enabled       bool                   `yaml:"enabled"`
	Backends      []string               `yaml:"backends"`
	File          FileLoggerConfig       `yaml:"file"`
	Database      DatabaseLoggerConfig   `yaml:"database"`
	Elasticsearch ElasticsearchLogConfig `yaml:"elasticsearch"`
}

type FileLoggerConfig struct {
	Path string `yaml:"path"`
}

type DatabaseLoggerConfig struct {
	DSN string `yaml:"dsn"`
}

type ElasticsearchLogConfig struct {
	URL   string `yaml:"url"`
	Index string `yaml:"index"`
}

type CacheConfig struct {
	Enabled    bool             `yaml:"enabled"`
	TTLSeconds int              `yaml:"ttl_seconds"`
	Redis      RedisCacheConfig `yaml:"redis"`
}

type RedisCacheConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type TemplatePathsConfig struct {
	Index string `yaml:"index"`
}

// Load reads configuration from a YAML file.
func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := Default()
	if err := yaml.Unmarshal(content, cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Default returns the baseline runtime configuration.
func Default() *Config {
	return &Config{
		Server: ServerConfig{
			Port: "8080",
		},
		HTTPClient: HTTPClientConfig{
			TimeoutSeconds:     10,
			MaxRedirects:       10,
			InsecureSkipVerify: false,
		},
		Browser: BrowserConfig{
			Enabled:        true,
			Command:        "google-chrome",
			TimeoutSeconds: 15,
		},
		Logging: LoggingConfig{
			Enabled:  true,
			Backends: []string{"file"},
			File: FileLoggerConfig{
				Path: "logs/errors.jsonl",
			},
			Database: DatabaseLoggerConfig{
				DSN: "file:logs/errors.db?_pragma=busy_timeout(5000)",
			},
			Elasticsearch: ElasticsearchLogConfig{
				URL:   "http://localhost:9200",
				Index: "webpage-analyzer-errors",
			},
		},
		Cache: CacheConfig{
			Enabled:    false,
			TTLSeconds: 300,
			Redis: RedisCacheConfig{
				Addr: "localhost:6379",
				DB:   0,
			},
		},
		TemplatePaths: TemplatePathsConfig{
			Index: "web/templates/index.html",
		},
	}
}

// Validate checks required values.
func (c *Config) Validate() error {
	if c.Server.Port == "" {
		return errors.New("server.port is required")
	}

	if c.HTTPClient.TimeoutSeconds <= 0 {
		return errors.New("http_client.timeout_seconds must be greater than zero")
	}

	if c.HTTPClient.MaxRedirects < 0 {
		return errors.New("http_client.max_redirects cannot be negative")
	}

	if c.Browser.TimeoutSeconds <= 0 {
		return errors.New("browser.timeout_seconds must be greater than zero")
	}

	if c.Cache.TTLSeconds <= 0 {
		return errors.New("cache.ttl_seconds must be greater than zero")
	}

	if c.TemplatePaths.Index == "" {
		return errors.New("template_paths.index is required")
	}

	return nil
}

func (c *Config) HTTPTimeout() time.Duration {
	return time.Duration(c.HTTPClient.TimeoutSeconds) * time.Second
}

func (c *Config) BrowserTimeout() time.Duration {
	return time.Duration(c.Browser.TimeoutSeconds) * time.Second
}

func (c *Config) CacheTTL() time.Duration {
	return time.Duration(c.Cache.TTLSeconds) * time.Second
}
