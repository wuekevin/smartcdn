package baotapanel

import (
	"context"
	"crypto/md5" //nolint:gosec
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/wuekevin/smartcdn/internal/domain/port"
)

// Config 是宝塔面板部署 Provider 的配置，存入 SSLDeployJob.ConfigJSON。
type Config struct {
	PanelURL string `json:"panel_url"` // 如 http://1.2.3.4:8888
	APIKey   string `json:"api_key"`   // 宝塔面板的 API Key
	SiteName string `json:"site_name"` // 宝塔网站名称（域名）
}

// Provider 通过宝塔面板 API 将证书部署到指定网站。
type Provider struct {
	cfg    Config
	client *http.Client
}

// New 是注册到 SSLDeployRegistry 的工厂函数。
// credentials 参数暂不使用（配置存于 config_json）。
func New(_ []byte) (port.SSLDeployProvider, error) {
	return &Provider{
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (p *Provider) DeployType() string { return "baota_panel" }

func (p *Provider) Deploy(ctx context.Context, input port.SSLDeployInput) error {
	var cfg Config
	if err := json.Unmarshal([]byte(input.ConfigJSON), &cfg); err != nil {
		return fmt.Errorf("baota_panel: 解析配置失败: %w", err)
	}
	if cfg.PanelURL == "" || cfg.APIKey == "" || cfg.SiteName == "" {
		return fmt.Errorf("baota_panel: panel_url / api_key / site_name 均为必填")
	}
	p.cfg = cfg

	token, ts := p.makeToken()

	endpoint := strings.TrimRight(cfg.PanelURL, "/") + "/site?action=SetSsl"

	form := url.Values{}
	form.Set("request_time", ts)
	form.Set("request_token", token)
	form.Set("type", "1")          // 1 = PEM 格式
	form.Set("siteName", cfg.SiteName)
	form.Set("key", input.PrivKeyPEM)
	form.Set("csr", input.CertPEM)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("baota_panel: 创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := p.client.Do(req)
	if err != nil {
		return fmt.Errorf("baota_panel: 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Status  bool   `json:"status"`
		Message string `json:"msg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("baota_panel: 解析响应失败 (HTTP %d): %s", resp.StatusCode, body)
	}
	if !result.Status {
		return fmt.Errorf("baota_panel: 部署失败: %s", result.Message)
	}
	return nil
}

// makeToken 生成宝塔面板 API 鉴权 Token。
// 算法：token = MD5( unix_timestamp + MD5(api_key) )
func (p *Provider) makeToken() (token, ts string) {
	now := time.Now().Unix()
	ts = fmt.Sprintf("%d", now)

	//nolint:gosec
	keyHash := fmt.Sprintf("%x", md5.Sum([]byte(p.cfg.APIKey)))
	//nolint:gosec
	full := fmt.Sprintf("%x", md5.Sum([]byte(ts+keyHash)))
	return full, ts
}

var _ port.SSLDeployProvider = (*Provider)(nil)
