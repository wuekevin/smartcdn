package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

const ConfigPath = "config/config.yaml"

// AppConfig holds all runtime configuration for SmartCDN.
// Loaded from config/config.yaml (written by the setup wizard).
// Falls back to environment variables when the file is absent.
type AppConfig struct {
	Database   DatabaseConfig `yaml:"database"`
	Redis      RedisConfig    `yaml:"redis"`
	EncryptKey string         `yaml:"encrypt_key"`
	SiteURL    string         `yaml:"site_url"`
	Port       string         `yaml:"port"`
}

type DatabaseConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// DSN builds a MySQL DSN string from the DatabaseConfig.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

// Load reads config/config.yaml if it exists, otherwise falls back to
// environment variables (DB_DSN, REDIS_URL, SMARTCDN_ENCRYPT_KEY, PORT).
func Load() (*AppConfig, error) {
	if _, err := os.Stat(ConfigPath); err == nil {
		return loadFromFile(ConfigPath)
	}
	return loadFromEnv(), nil
}

func loadFromFile(path string) (*AppConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return &cfg, nil
}

func loadFromEnv() *AppConfig {
	cfg := &AppConfig{
		EncryptKey: os.Getenv("SMARTCDN_ENCRYPT_KEY"),
		SiteURL:    os.Getenv("SITE_URL"),
		Port:       os.Getenv("PORT"),
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	// DB_DSN is a full DSN; parse into struct fields for uniform access.
	// When using env vars the caller can also read DB_DSN directly via DSNFromEnv().
	return cfg
}

// DSNFromEnv returns the raw DB_DSN environment variable.
// Used when config.yaml is absent and Database struct fields are not populated.
func DSNFromEnv() string {
	return os.Getenv("DB_DSN")
}

// RedisURLFromEnv returns the raw REDIS_URL environment variable.
func RedisURLFromEnv() string {
	return os.Getenv("REDIS_URL")
}

// Write serialises cfg to config/config.yaml, creating the directory if needed.
func Write(cfg *AppConfig) error {
	if err := os.MkdirAll("config", 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}
	return os.WriteFile(ConfigPath, data, 0600)
}
