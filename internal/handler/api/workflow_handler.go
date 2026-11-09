package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	wfapp "github.com/wuekevin/smartcdn/internal/application/workflow"
)

// WorkflowHandler exposes DAG workflow management APIs.
type WorkflowHandler struct {
	engine *wfapp.Engine
}

func NewWorkflowHandler(engine *wfapp.Engine) *WorkflowHandler {
	return &WorkflowHandler{engine: engine}
}

func (h *WorkflowHandler) RegisterRoutes(r *gin.RouterGroup) {
	t := r.Group("/workflow-templates")
	t.POST("", h.CreateTemplate)
	t.GET("", h.ListTemplates)
	t.GET("/:id", h.GetTemplate)
	t.PUT("/:id", h.UpdateTemplate)
	t.DELETE("/:id", h.DeleteTemplate)
	t.POST("/:id/trigger", h.TriggerRun)
	t.GET("/:id/runs", h.ListRuns)

	r2 := r.Group("/workflow-runs")
	r2.GET("/:id", h.GetRun)
	r2.POST("/:id/cancel", h.CancelRun)
	r2.GET("/:id/nodes", h.GetNodeRuns)
}

// --- Templates ---

type createTemplateReq struct {
	Name        string `json:"name" binding:"required,max=128"`
	Description string `json:"description" binding:"max=512"`
	GraphJSON   string `json:"graph_json" binding:"required"`
}

func (h *WorkflowHandler) CreateTemplate(c *gin.Context) {
	var req createTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.engine.CreateTemplate(c.Request.Context(), req.Name, req.Description, req.GraphJSON)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *WorkflowHandler) ListTemplates(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	if limit <= 0 {
		limit = 20
	}
	ts, err := h.engine.ListTemplates(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": ts})
}

func (h *WorkflowHandler) GetTemplate(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	v, err := h.engine.GetTemplate(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, wfapp.ErrTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

type updateTemplateReq struct {
	Description string `json:"description"`
	GraphJSON   string `json:"graph_json"`
}

func (h *WorkflowHandler) UpdateTemplate(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateTemplateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.engine.UpdateTemplate(c.Request.Context(), id, req.Description, req.GraphJSON)
	if err != nil {
		if errors.Is(err, wfapp.ErrTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *WorkflowHandler) DeleteTemplate(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.engine.DeleteTemplate(c.Request.Context(), id); err != nil {
		if errors.Is(err, wfapp.ErrTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// --- Run trigger ---

type triggerRunReq struct {
	TriggerType string                 `json:"trigger_type" binding:"required,max=64"`
	TriggerRef  string                 `json:"trigger_ref" binding:"max=255"`
	ContextVars map[string]interface{} `json:"context_vars"`
}

func (h *WorkflowHandler) TriggerRun(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req triggerRunReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.engine.TriggerRun(c.Request.Context(), id, req.TriggerType, req.TriggerRef, req.ContextVars)
	if err != nil {
		if errors.Is(err, wfapp.ErrTemplateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

// --- Runs ---

func (h *WorkflowHandler) ListRuns(c *gin.Context) {
	templateID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	if limit <= 0 {
		limit = 20
	}
	rs, err := h.engine.ListRuns(c.Request.Context(), templateID, afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": rs})
}

func (h *WorkflowHandler) GetRun(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	v, err := h.engine.GetRun(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, wfapp.ErrRunNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *WorkflowHandler) CancelRun(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.engine.CancelRun(c.Request.Context(), id); err != nil {
		if errors.Is(err, wfapp.ErrRunNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "cancelled"})
}

func (h *WorkflowHandler) GetNodeRuns(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	ns, err := h.engine.GetNodeRuns(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": ns})
}
