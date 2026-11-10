package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

type AuditLogsHandler struct {
	audit *repo.AuditRepo
}

func NewAuditLogsHandler(audit *repo.AuditRepo) *AuditLogsHandler {
	return &AuditLogsHandler{audit: audit}
}

func (h *AuditLogsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/audit-logs", h.List)
}

func (h *AuditLogsHandler) List(c *gin.Context) {
	afterID, _ := parseIntQuery(c, "after_id", 0)
	limit, _ := parseIntQuery(c, "limit", 20)
	operatorID, _ := parseIntQuery(c, "operator_id", 0)

	filter := repo.AuditFilter{
		OperatorID: uint64(operatorID),
		AfterID:    uint64(afterID),
		Limit:      limit,
	}

	if s := c.Query("start"); s != "" {
		if t, err := time.Parse(time.RFC3339, s); err == nil {
			filter.StartTime = t
		}
	}
	if e := c.Query("end"); e != "" {
		if t, err := time.Parse(time.RFC3339, e); err == nil {
			filter.EndTime = t
		}
	}

	logs, err := h.audit.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	type logView struct {
		ID           uint64     `json:"id"`
		OperatorID   *uint64    `json:"operator_id"`
		OperatorIP   string     `json:"operator_ip"`
		Action       string     `json:"action"`
		ResourceType string     `json:"resource_type"`
		ResourceID   string     `json:"resource_id"`
		BeforeState  *string    `json:"before_state,omitempty"`
		AfterState   *string    `json:"after_state,omitempty"`
		CreatedAt    time.Time  `json:"created_at"`
	}
	items := make([]logView, 0, len(logs))
	for _, l := range logs {
		items = append(items, logView{
			ID:           l.ID,
			OperatorID:   l.OperatorID,
			OperatorIP:   l.OperatorIP,
			Action:       l.Action,
			ResourceType: l.ResourceType,
			ResourceID:   l.ResourceID,
			BeforeState:  l.BeforeState,
			AfterState:   l.AfterState,
			CreatedAt:    l.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
