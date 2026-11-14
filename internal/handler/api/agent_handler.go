package api

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	nodeapp "github.com/wuekevin/smartcdn/internal/application/node"
	probeapp "github.com/wuekevin/smartcdn/internal/application/probe"
	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// AgentHandler serves the internal API consumed by Node Agents and Probe Agents.
// All routes under /agent/* require mTLS client certificate authentication.
type AgentHandler struct {
	nodeSvc  *nodeapp.Service
	probeSvc *probeapp.Service
}

func NewAgentHandler(nodeSvc *nodeapp.Service, probeSvc *probeapp.Service) *AgentHandler {
	return &AgentHandler{nodeSvc: nodeSvc, probeSvc: probeSvc}
}

// RegisterManagementRoutes registers operator-facing management endpoints for probe agents
// and IP candidates. These are mounted on the JWT-protected router.
func (h *AgentHandler) RegisterManagementRoutes(r *gin.RouterGroup) {
	r.GET("/probe-agents", h.ListProbeAgents)
	r.GET("/ip-candidates", h.ListIPCandidates)
}

// GET /api/v1/probe-agents
func (h *AgentHandler) ListProbeAgents(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	views, err := h.probeSvc.ListAgents(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

// GET /api/v1/ip-candidates
func (h *AgentHandler) ListIPCandidates(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 50)
	views, err := h.probeSvc.ListAllCandidates(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

// RegisterRoutes registers agent API routes. The caller is responsible for
// mounting this on a *gin.RouterGroup that is served over a tls.Config with
// tls.RequireAndVerifyClientCert, so the cert SN is available via tls.ConnectionState.
func (h *AgentHandler) RegisterRoutes(r *gin.RouterGroup) {
	agent := r.Group("/agent")
	agent.Use(h.mtlsAuthMiddleware())

	// Node Agent routes
	node := agent.Group("/node")
	node.POST("/heartbeat", h.NodeHeartbeat)
	node.GET("/tasks", h.NodePullTasks)
	node.POST("/tasks/:tid/ack", h.NodeAckTask)

	// Probe Agent routes
	probe := agent.Group("/probe")
	probe.POST("/register", h.ProbeRegister)
	probe.POST("/heartbeat", h.ProbeHeartbeat)
	probe.POST("/results", h.ProbeSubmitResult)
}

// mtlsAuthMiddleware extracts the client cert serial from the TLS connection state
// and injects it into the Gin context for downstream handlers.
func (h *AgentHandler) mtlsAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tlsState := tlsConnectionState(c.Request)
		if tlsState == nil || len(tlsState.PeerCertificates) == 0 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "mTLS client certificate required"})
			return
		}
		certSN := tlsState.PeerCertificates[0].SerialNumber.Text(16)
		c.Set("client_cert_sn", certSN)
		c.Next()
	}
}

// --- Node Agent handlers ---

type nodeHeartbeatReq struct {
	Token       string  `json:"token" binding:"required"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemPercent  float64 `json:"mem_percent"`
	HTTPSStatus string  `json:"https_status"`
}

func (h *AgentHandler) NodeHeartbeat(c *gin.Context) {
	var req nodeHeartbeatReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	certSN, _ := c.Get("client_cert_sn")
	err := h.nodeSvc.RecordHeartbeat(c.Request.Context(), nodeapp.HeartbeatInput{
		Token:       req.Token,
		CPUPercent:  req.CPUPercent,
		MemPercent:  req.MemPercent,
		HTTPSStatus: req.HTTPSStatus,
		CertSN:      fmt.Sprintf("%v", certSN),
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *AgentHandler) NodePullTasks(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
		return
	}
	tasks, err := h.nodeSvc.PullTasks(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

type ackTaskReq struct {
	Token     string `json:"token" binding:"required"`
	Status    string `json:"status" binding:"required"` // succeeded | failed
	ResultMsg string `json:"result_msg"`
}

func (h *AgentHandler) NodeAckTask(c *gin.Context) {
	tid, err := parseUint64Param(c, "tid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid task id"})
		return
	}
	var req ackTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	status := entity.NodeTaskStatus(req.Status)
	if status != entity.NodeTaskSucceeded && status != entity.NodeTaskFailed {
		c.JSON(http.StatusBadRequest, gin.H{"error": "status must be 'succeeded' or 'failed'"})
		return
	}
	if err := h.nodeSvc.AckTask(c.Request.Context(), req.Token, tid, status, req.ResultMsg); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// --- Probe Agent handlers ---

type probeRegisterReq struct {
	Name   string `json:"name" binding:"required,max=128"`
	ISP    string `json:"isp" binding:"required"`
	Region string `json:"region"`
}

func (h *AgentHandler) ProbeRegister(c *gin.Context) {
	var req probeRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	certSN, _ := c.Get("client_cert_sn")
	v, err := h.probeSvc.RegisterAgent(c.Request.Context(), probeapp.RegisterAgentInput{
		Name:   req.Name,
		ISP:    req.ISP,
		Region: req.Region,
		CertSN: fmt.Sprintf("%v", certSN),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *AgentHandler) ProbeHeartbeat(c *gin.Context) {
	certSN, _ := c.Get("client_cert_sn")
	if err := h.probeSvc.RecordHeartbeat(c.Request.Context(), fmt.Sprintf("%v", certSN)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type probeResultReq struct {
	ProbeAgentID uint64 `json:"probe_agent_id" binding:"required"`
	TargetIP     string `json:"target_ip" binding:"required"`
	ISP          string `json:"isp" binding:"required"`
	LatencyMs    int64  `json:"latency_ms"`
	TTFBMs       int64  `json:"ttfb_ms"`
	HTTPSStatus  int    `json:"https_status"`
	Success      bool   `json:"success"`
}

func (h *AgentHandler) ProbeSubmitResult(c *gin.Context) {
	var req probeResultReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.probeSvc.SubmitResult(c.Request.Context(), probeapp.SubmitResultInput{
		ProbeAgentID: req.ProbeAgentID,
		TargetIP:     req.TargetIP,
		ISP:          req.ISP,
		LatencyMs:    req.LatencyMs,
		TTFBMs:       req.TTFBMs,
		HTTPSStatus:  req.HTTPSStatus,
		Success:      req.Success,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// tlsConnectionState extracts TLS state from an HTTP request (compatible with
// both native *tls.Conn and http.Server's tls field).
func tlsConnectionState(r *http.Request) *tls.ConnectionState {
	if r.TLS != nil {
		return r.TLS
	}
	return nil
}
