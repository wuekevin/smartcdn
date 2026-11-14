package localfile

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wuekevin/smartcdn/internal/domain/port"
)

type deployConfig struct {
	CertPath string `json:"cert_path"`
	KeyPath  string `json:"key_path"`
}

// Provider 将证书写入本地文件系统。
type Provider struct{}

func New(_ []byte) (port.SSLDeployProvider, error) { return &Provider{}, nil }

func (p *Provider) DeployType() string { return "local_file" }

func (p *Provider) Deploy(_ context.Context, input port.SSLDeployInput) error {
	var cfg deployConfig
	if err := json.Unmarshal([]byte(input.ConfigJSON), &cfg); err != nil {
		return fmt.Errorf("local_file: parse config: %w", err)
	}
	if cfg.CertPath == "" || cfg.KeyPath == "" {
		return fmt.Errorf("local_file: cert_path 和 key_path 为必填项")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.CertPath), 0755); err != nil {
		return fmt.Errorf("local_file: 创建证书目录: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cfg.KeyPath), 0700); err != nil {
		return fmt.Errorf("local_file: 创建私钥目录: %w", err)
	}
	if err := os.WriteFile(cfg.CertPath, []byte(input.CertPEM), 0644); err != nil {
		return fmt.Errorf("local_file: 写入证书: %w", err)
	}
	if err := os.WriteFile(cfg.KeyPath, []byte(input.PrivKeyPEM), 0600); err != nil {
		return fmt.Errorf("local_file: 写入私钥: %w", err)
	}
	return nil
}
