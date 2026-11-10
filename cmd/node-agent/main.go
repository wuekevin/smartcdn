// node-agent runs on each CDN node and communicates with the SmartCDN server
// via mTLS. It polls for tasks, executes them (reload_nginx, clear_cache, etc.),
// and reports node metrics.
package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func main() {
	serverURL := requireEnv("SMARTCDN_SERVER_URL")  // e.g. https://cdn.example.com
	certFile := requireEnv("SMARTCDN_AGENT_CERT")   // client cert PEM
	keyFile := requireEnv("SMARTCDN_AGENT_KEY")     // client key PEM
	caFile := requireEnv("SMARTCDN_CA_CERT")        // CA cert PEM
	nodeID := requireEnv("SMARTCDN_NODE_ID")        // assigned node UUID

	client := buildMTLSClient(caFile, certFile, keyFile)

	// Send initial heartbeat
	heartbeat(client, serverURL, nodeID)

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		heartbeat(client, serverURL, nodeID)
		pollAndExecTasks(client, serverURL, nodeID)
	}
}

// heartbeat POSTs to /agent/node/heartbeat.
func heartbeat(c *http.Client, server, nodeID string) {
	payload := map[string]interface{}{
		"node_id": nodeID,
		"ts":      time.Now().Unix(),
	}
	postJSON(c, server+"/agent/node/heartbeat", payload)
}

type task struct {
	ID       uint64            `json:"id"`
	Type     string            `json:"type"`
	Params   map[string]string `json:"params"`
}

// pollAndExecTasks fetches pending tasks and executes them.
func pollAndExecTasks(c *http.Client, server, nodeID string) {
	resp, err := c.Get(server + "/agent/node/tasks?node_id=" + nodeID)
	if err != nil || resp.StatusCode != 200 {
		return
	}
	defer resp.Body.Close()

	var body struct {
		Items []task `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return
	}

	for _, t := range body.Items {
		success, out := execTask(t)
		ackTask(c, server, t.ID, success, out)
	}
}

// execTask runs a single task and returns (success, output).
func execTask(t task) (bool, string) {
	switch t.Type {
	case "reload_nginx":
		return runCmd("nginx", "-s", "reload")
	case "restart_nginx":
		return runCmd("systemctl", "restart", "nginx")
	case "clear_cache":
		cacheDir := t.Params["cache_dir"]
		if cacheDir == "" {
			cacheDir = "/var/cache/nginx"
		}
		return runCmd("find", cacheDir, "-type", "f", "-delete")
	case "check_https":
		domain := t.Params["domain"]
		if domain == "" {
			return false, "missing domain param"
		}
		return runCmd("curl", "-sSo", "/dev/null", "-w", "%{http_code}", "https://"+domain)
	case "collect_metrics":
		return true, collectMetricsJSON()
	default:
		return false, fmt.Sprintf("unknown task type: %s", t.Type)
	}
}

func runCmd(name string, args ...string) (bool, string) {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return false, string(out) + " err: " + err.Error()
	}
	return true, strings.TrimSpace(string(out))
}

func collectMetricsJSON() string {
	// Read basic /proc metrics for the report.
	load, _ := os.ReadFile("/proc/loadavg")
	mem, _ := os.ReadFile("/proc/meminfo")
	m := map[string]string{
		"loadavg": strings.TrimSpace(string(load)),
		"meminfo": string(mem),
	}
	b, _ := json.Marshal(m)
	return string(b)
}

func ackTask(c *http.Client, server string, taskID uint64, success bool, output string) {
	status := "succeeded"
	if !success {
		status = "failed"
	}
	postJSON(c, fmt.Sprintf("%s/agent/node/tasks/%d/ack", server, taskID), map[string]interface{}{
		"status": status,
		"output": output,
	})
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
