package config

import (
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

const ConfigPath = "config/config.yaml"

// AppConfig holds all runtime configuration written by the setup wizard.
// It is only read when install.lock exists; if the file is missing that is
// a hard error — the user must re-run the setup wizard.
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

// DSN builds a MySQL DSN string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

// Load reads config/config.yaml. Called only after install.lock is confirmed
// present; returns an error if the file is missing or malformed.
func Load() (*AppConfig, error) {
	data, err := os.ReadFile(ConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config/config.yaml not found — please re-run the setup wizard")
		}
		return nil, fmt.Errorf("read %s: %w", ConfigPath, err)
	}
	var cfg AppConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", ConfigPath, err)
	}
	if cfg.Port == "" {
		cfg.Port = "8080"
	}
	return &cfg, nil
}

// Write serialises cfg to config/config.yaml (mode 0600), creating the
// config/ directory if needed. Called by the setup wizard on completion.
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
