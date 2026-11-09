package dnscheck

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/wuekevin/smartcdn/internal/infra/httpclient"
)

// HTTPSResult is the outcome of a reachability probe.
type HTTPSResult struct {
	URL        string
	StatusCode int
	TLSValid   bool
	Latency    time.Duration
}

// CheckHTTPS probes https://host/ via a direct IP dial, with SSRF protection.
// ip is the resolved address to dial; host is the SNI/Host header.
func CheckHTTPS(ctx context.Context, ip, host string) (*HTTPSResult, error) {
	if err := httpclient.CheckIP(ip); err != nil {
		return nil, fmt.Errorf("SSRF protection: %w", err)
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, ip+":443")
		},
		TLSClientConfig: &tls.Config{ServerName: host},
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse // don't follow redirects
		},
	}

	url := "https://" + host + "/"
	start := time.Now()
	req, _ := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	req.Host = host

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("https probe %q: %w", url, err)
	}
	defer resp.Body.Close()

	return &HTTPSResult{
		URL:        url,
		StatusCode: resp.StatusCode,
		TLSValid:   resp.TLS != nil,
		Latency:    time.Since(start),
	}, nil
}
