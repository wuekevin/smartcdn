// probe-agent is deployed in real ISP networks to measure CDN latency/TTFB
// from actual end-user network positions. It registers itself with the
// SmartCDN server via mTLS and submits periodic probe results.
package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	serverURL := requireEnv("SMARTCDN_SERVER_URL")
	certFile := requireEnv("SMARTCDN_AGENT_CERT")
	keyFile := requireEnv("SMARTCDN_AGENT_KEY")
	caFile := requireEnv("SMARTCDN_CA_CERT")
	isp := requireEnv("SMARTCDN_ISP")        // e.g. ChinaTelecom
	region := requireEnv("SMARTCDN_REGION")  // e.g. Beijing

	client := buildMTLSClient(caFile, certFile, keyFile)

	// Register this probe agent.
	registerProbe(client, serverURL, isp, region)

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		heartbeat(client, serverURL)
		submitProbeResults(client, serverURL, isp, region)
	}
}

func registerProbe(c *http.Client, server, isp, region string) {
	postJSON(c, server+"/agent/probe/register", map[string]interface{}{
		"isp":    isp,
		"region": region,
	})
}

func heartbeat(c *http.Client, server string) {
	postJSON(c, server+"/agent/probe/heartbeat", map[string]interface{}{
		"ts": time.Now().Unix(),
	})
}

type probeTarget struct {
	IP        string `json:"ip"`
	CandidateID uint64 `json:"candidate_id"`
}

type probeResult struct {
	CandidateID uint64  `json:"candidate_id"`
	LatencyMs   int64   `json:"latency_ms"`
	TTFBMs      int64   `json:"ttfb_ms"`
	StatusCode  int     `json:"status_code"`
	TLSValid    bool    `json:"tls_valid"`
	Success     bool    `json:"success"`
}

func submitProbeResults(c *http.Client, server, isp, region string) {
	// Fetch IP candidates to probe from server.
	resp, err := c.Get(server + "/agent/probe/targets")
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	var body struct {
		Items []probeTarget `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil || len(body.Items) == 0 {
		return
	}

	results := make([]probeResult, 0, len(body.Items))
	for _, target := range body.Items {
		r := measureTarget(target.IP)
		r.CandidateID = target.CandidateID
		results = append(results, r)
	}

	postJSON(c, server+"/agent/probe/results", map[string]interface{}{
		"results": results,
	})
}

func measureTarget(ip string) probeResult {
	url := "https://" + ip + "/"
	start := time.Now()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return probeResult{Success: false}
	}
	// Skip TLS hostname verification; we're probing by IP.
	probeClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
		},
		Timeout: 10 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := probeClient.Do(req)
	latency := time.Since(start)
	if err != nil {
		return probeResult{LatencyMs: latency.Milliseconds(), Success: false}
	}
	defer resp.Body.Close()

	tlsValid := resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0

	return probeResult{
		LatencyMs:  latency.Milliseconds(),
		TTFBMs:     latency.Milliseconds(), // approximation; first byte = response received
		StatusCode: resp.StatusCode,
		TLSValid:   tlsValid,
		Success:    resp.StatusCode < 500,
	}
}

// --- HTTP helpers ---

func postJSON(c *http.Client, url string, body interface{}) {
	b, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.Do(req)
	if err != nil {
		return
	}
	resp.Body.Close()
}

func buildMTLSClient(caFile, certFile, keyFile string) *http.Client {
	caCert, err := os.ReadFile(caFile)
	must(err, "read CA cert")
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	must(err, "load client cert/key")

	tlsCfg := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
	}
	return &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsCfg},
		Timeout:   15 * time.Second,
	}
}

func requireEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		fmt.Fprintf(os.Stderr, "FATAL: %s is required\n", key)
		os.Exit(1)
	}
	return v
}

func must(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "FATAL: %s: %v\n", msg, err)
		os.Exit(1)
	}
}
