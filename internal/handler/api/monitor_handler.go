package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	monapp "github.com/wuekevin/smartcdn/internal/application/monitor"
	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// MonitorHandler exposes metrics, alert rules, channels, and fired alerts APIs.
type MonitorHandler struct {
	svc *monapp.Service
}

func NewMonitorHandler(svc *monapp.Service) *MonitorHandler {
	return &MonitorHandler{svc: svc}
}

func (h *MonitorHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Metrics
	m := r.Group("/metrics")
	m.POST("/ingest", h.IngestMetrics)
	m.GET("", h.QueryMetrics)

	// Alert rules
	ar := r.Group("/alert-rules")
	ar.POST("", h.CreateRule)
	ar.GET("", h.ListRules)
	ar.GET("/:id", h.GetRule)
	ar.PUT("/:id", h.UpdateRule)
	ar.DELETE("/:id", h.DeleteRule)

	// Alert channels
	ac := r.Group("/alert-channels")
	ac.POST("", h.CreateChannel)
	ac.GET("", h.ListChannels)
	ac.GET("/:id", h.GetChannel)
	ac.DELETE("/:id", h.DeleteChannel)

	// Fired alerts
	al := r.Group("/alerts")
	al.GET("", h.ListAlerts)
	al.POST("/:id/resolve", h.ResolveAlert)
}

// --- Metrics ---

type ingestMetricItem struct {
	SourceType  string            `json:"source_type" binding:"required"`
	SourceID    uint64            `json:"source_id" binding:"required"`
	MetricType  string            `json:"metric_type" binding:"required"`
	Value       float64           `json:"value"`
	Tags        map[string]string `json:"tags"`
}

type ingestMetricsReq struct {
	Items []ingestMetricItem `json:"items" binding:"required,min=1"`
}

func (h *MonitorHandler) IngestMetrics(c *gin.Context) {
	var req ingestMetricsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	inputs := make([]monapp.IngestInput, len(req.Items))
	for i, it := range req.Items {
		inputs[i] = monapp.IngestInput{
			SourceType: it.SourceType,
			SourceID:   it.SourceID,
			MetricType: entity.MetricType(it.MetricType),
			Value:      it.Value,
			Tags:       it.Tags,
		}
	}
	if err := h.svc.IngestMetrics(c.Request.Context(), inputs); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

func (h *MonitorHandler) QueryMetrics(c *gin.Context) {
	sourceType := c.Query("source_type")
	sourceID, _ := parseUint64Query(c, "source_id")
	metricType := entity.MetricType(c.Query("metric_type"))
	limit, _ := parseIntQuery(c, "limit", 100)
	if limit <= 0 {
		limit = 100
	}
	ms, err := h.svc.QueryMetrics(c.Request.Context(), sourceType, sourceID, metricType, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": ms})
}

// --- Alert rules ---

type ruleReq struct {
	Name          string   `json:"name" binding:"required,max=128"`
	Description   string   `json:"description"`
	SourceType    string   `json:"source_type" binding:"required"`
	SourceID      *uint64  `json:"source_id"`
	MetricType    string   `json:"metric_type" binding:"required"`
	Condition     string   `json:"condition" binding:"required"`
	Threshold     float64  `json:"threshold"`
	WindowMinutes int      `json:"window_minutes"`
	ForMinutes    int      `json:"for_minutes"`
	Severity      string   `json:"severity"`
	ChannelIDs    []uint64 `json:"channel_ids"`
}

func toCreateRuleInput(req ruleReq) monapp.CreateRuleInput {
	return monapp.CreateRuleInput{
		Name:          req.Name,
		Description:   req.Description,
		SourceType:    req.SourceType,
		SourceID:      req.SourceID,
		MetricType:    entity.MetricType(req.MetricType),
		Condition:     entity.AlertCondition(req.Condition),
		Threshold:     req.Threshold,
		WindowMinutes: req.WindowMinutes,
		ForMinutes:    req.ForMinutes,
		Severity:      entity.AlertSeverity(req.Severity),
		ChannelIDs:    req.ChannelIDs,
	}
}

func (h *MonitorHandler) CreateRule(c *gin.Context) {
	var req ruleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.svc.CreateRule(c.Request.Context(), toCreateRuleInput(req))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, rule)
}

func (h *MonitorHandler) ListRules(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	rules, err := h.svc.ListRules(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rules})
}

func (h *MonitorHandler) GetRule(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	rule, err := h.svc.GetRule(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, monapp.ErrRuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *MonitorHandler) UpdateRule(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req ruleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rule, err := h.svc.UpdateRule(c.Request.Context(), id, toCreateRuleInput(req))
	if err != nil {
		if errors.Is(err, monapp.ErrRuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rule)
}

func (h *MonitorHandler) DeleteRule(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteRule(c.Request.Context(), id); err != nil {
		if errors.Is(err, monapp.ErrRuleNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// --- Alert channels ---

type channelReq struct {
	Name   string                 `json:"name" binding:"required,max=128"`
	Type   string                 `json:"type" binding:"required"`
	Config map[string]interface{} `json:"config"`
}

func (h *MonitorHandler) CreateChannel(c *gin.Context) {
	var req channelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ch, err := h.svc.CreateChannel(c.Request.Context(), monapp.CreateChannelInput{
		Name:   req.Name,
		Type:   entity.AlertChannelType(req.Type),
		Config: req.Config,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, ch)
}

func (h *MonitorHandler) ListChannels(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	chs, err := h.svc.ListChannels(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": chs})
}

func (h *MonitorHandler) GetChannel(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ch, err := h.svc.GetChannel(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, monapp.ErrChannelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ch)
}

func (h *MonitorHandler) DeleteChannel(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteChannel(c.Request.Context(), id); err != nil {
		if errors.Is(err, monapp.ErrChannelNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// --- Alerts ---

func (h *MonitorHandler) ListAlerts(c *gin.Context) {
	status := entity.AlertStatus(c.Query("status"))
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	alerts, err := h.svc.ListAlerts(c.Request.Context(), status, afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": alerts})
}

func (h *MonitorHandler) ResolveAlert(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.ResolveAlert(c.Request.Context(), id); err != nil {
		if errors.Is(err, monapp.ErrAlertNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "resolved"})
}
