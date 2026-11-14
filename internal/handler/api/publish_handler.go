package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	publishsvc "github.com/wuekevin/smartcdn/internal/application/publish"
	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

type PublishHandler struct {
	engine *publishsvc.Engine
}

func NewPublishHandler(engine *publishsvc.Engine) *PublishHandler {
	return &PublishHandler{engine: engine}
}

func (h *PublishHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/domains/:id/publish")
	g.POST("", h.Publish)
	g.POST("/dry-run", h.DryRun)

	r.GET("/publish-jobs", h.ListJobs)
}

// GET /api/v1/publish-jobs
func (h *PublishHandler) ListJobs(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	jobs, err := h.engine.ListJobs(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": jobs})
}

// POST /api/v1/domains/:id/publish
func (h *PublishHandler) Publish(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var body struct {
		Records []entity.PublishRecord `json:"records" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.engine.Run(c.Request.Context(), publishsvc.PublishRequest{
		DomainID: domainID,
		Records:  body.Records,
		DryRun:   false,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	status := http.StatusOK
	if result.Status == string(entity.PublishStatusFailed) {
		status = http.StatusUnprocessableEntity
	}
	c.JSON(status, result)
}

// POST /api/v1/domains/:id/publish/dry-run
func (h *PublishHandler) DryRun(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var body struct {
		Records []entity.PublishRecord `json:"records" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.engine.Run(c.Request.Context(), publishsvc.PublishRequest{
		DomainID: domainID,
		Records:  body.Records,
		DryRun:   true,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
