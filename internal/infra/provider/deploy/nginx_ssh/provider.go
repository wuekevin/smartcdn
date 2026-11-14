package nginxssh

import (
	"context"
	"encoding/json"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"

	"github.com/wuekevin/smartcdn/internal/domain/port"
)

type config struct {
	Host       string `json:"host"`
	Port       string `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	PrivateKey string `json:"private_key"`
	CertPath   string `json:"cert_path"`
	KeyPath    string `json:"key_path"`
	ReloadCmd  string `json:"reload_cmd"`
}

// Provider 通过 SSH 将证书部署到远程 Nginx 服务器。
type Provider struct{}

func New(_ []byte) (port.SSLDeployProvider, error) { return &Provider{}, nil }

func (p *Provider) DeployType() string { return "nginx_ssh" }

func (p *Provider) Deploy(_ context.Context, input port.SSLDeployInput) error {
	var cfg config
	if err := json.Unmarshal([]byte(input.ConfigJSON), &cfg); err != nil {
		return fmt.Errorf("nginx_ssh: 解析配置失败: %w", err)
	}
	if cfg.Host == "" {
		return fmt.Errorf("nginx_ssh: host 为必填项")
	}
	if cfg.CertPath == "" || cfg.KeyPath == "" {
		return fmt.Errorf("nginx_ssh: cert_path 和 key_path 为必填项")
	}

	sshPort := cfg.Port
	if sshPort == "" {
		sshPort = "22"
	}

	var authMethods []ssh.AuthMethod
	if cfg.PrivateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(cfg.PrivateKey))
		if err != nil {
			return fmt.Errorf("nginx_ssh: 解析私钥失败: %w", err)
		}
		authMethods = append(authMethods, ssh.PublicKeys(signer))
	}
	if cfg.Password != "" {
		authMethods = append(authMethods, ssh.Password(cfg.Password))
	}
	if len(authMethods) == 0 {
		return fmt.Errorf("nginx_ssh: 需要提供 password 或 private_key")
	}

	sshCfg := &ssh.ClientConfig{
		User:            cfg.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), //nolint:gosec
	}
	addr := net.JoinHostPort(cfg.Host, sshPort)
	client, err := ssh.Dial("tcp", addr, sshCfg)
	if err != nil {
		return fmt.Errorf("nginx_ssh: SSH 连接 %s 失败: %w", addr, err)
	}
	defer client.Close()

	run := func(cmd string) error {
		sess, err := client.NewSession()
		if err != nil {
			return err
		}
		defer sess.Close()
		return sess.Run(cmd)
	}

	writeRemote := func(remotePath, content string, mode string) error {
		mkdirCmd := fmt.Sprintf("mkdir -p $(dirname %q)", remotePath)
		if err := run(mkdirCmd); err != nil {
			return fmt.Errorf("mkdir failed: %w", err)
		}
		// 通过 stdin 写入内容以避免 shell 转义问题
		sess, err := client.NewSession()
		if err != nil {
			return err
		}
		defer sess.Close()
		sess.Stdin = newStringReader(content)
		return sess.Run(fmt.Sprintf("cat > %q && chmod %s %q", remotePath, mode, remotePath))
	}

	if err := writeRemote(cfg.CertPath, input.CertPEM, "644"); err != nil {
		return fmt.Errorf("nginx_ssh: 写入证书失败: %w", err)
	}
	if err := writeRemote(cfg.KeyPath, input.PrivKeyPEM, "600"); err != nil {
		return fmt.Errorf("nginx_ssh: 写入私钥失败: %w", err)
	}

	if cfg.ReloadCmd != "" {
		if err := run(cfg.ReloadCmd); err != nil {
			return fmt.Errorf("nginx_ssh: 重载命令执行失败: %w", err)
		}
	}
	return nil
}

type stringReader struct {
	data []byte
	pos  int
}

func newStringReader(s string) *stringReader { return &stringReader{data: []byte(s)} }

func (r *stringReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, fmt.Errorf("EOF")
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
