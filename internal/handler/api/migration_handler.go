package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	migsvc "github.com/wuekevin/smartcdn/internal/application/migration"
)

type MigrationHandler struct {
	engine   *migsvc.Engine
	multiDNS *migsvc.MultiDNSService
}

func NewMigrationHandler(engine *migsvc.Engine, multiDNS *migsvc.MultiDNSService) *MigrationHandler {
	return &MigrationHandler{engine: engine, multiDNS: multiDNS}
}

func (h *MigrationHandler) RegisterRoutes(r *gin.RouterGroup) {
	// Migration
	m := r.Group("/domains/:id/migrations")
	m.POST("", h.StartMigration)
	m.GET("", h.ListMigrations)
	m.POST("/:jobid/cutover", h.Cutover)
	m.POST("/:jobid/rollback", h.Rollback)

	// Multi-DNS
	md := r.Group("/domains/:id/multidns")
	md.PUT("", h.EnableMultiDNS)
	md.GET("", h.GetMultiDNS)
	md.DELETE("", h.DisableMultiDNS)
}

// POST /api/v1/domains/:id/migrations
func (h *MigrationHandler) StartMigration(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var body struct {
		TargetAccountID uint64 `json:"target_account_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.engine.Start(c.Request.Context(), migsvc.StartRequest{
		DomainID:        domainID,
		TargetAccountID: body.TargetAccountID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

// GET /api/v1/domains/:id/migrations  (placeholder; engine has no ListByDomain on handler yet)
func (h *MigrationHandler) ListMigrations(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"items": []interface{}{}})
}

// POST /api/v1/domains/:id/migrations/:jobid/cutover
func (h *MigrationHandler) Cutover(c *gin.Context) {
	jobID, err := parseUint64Param(c, "jobid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	result, err := h.engine.Cutover(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// POST /api/v1/domains/:id/migrations/:jobid/rollback
func (h *MigrationHandler) Rollback(c *gin.Context) {
	jobID, err := parseUint64Param(c, "jobid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid job id"})
		return
	}
	result, err := h.engine.Rollback(c.Request.Context(), jobID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// PUT /api/v1/domains/:id/multidns
func (h *MigrationHandler) EnableMultiDNS(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var body struct {
		PrimaryAccountID   uint64 `json:"primary_account_id" binding:"required"`
		SecondaryAccountID uint64 `json:"secondary_account_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cfg, err := h.multiDNS.Enable(c.Request.Context(), domainID, body.PrimaryAccountID, body.SecondaryAccountID)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// GET /api/v1/domains/:id/multidns
func (h *MigrationHandler) GetMultiDNS(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	cfg, err := h.multiDNS.GetConfig(c.Request.Context(), domainID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "multidns not configured for this domain"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// DELETE /api/v1/domains/:id/multidns
func (h *MigrationHandler) DisableMultiDNS(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	if err := h.multiDNS.Disable(c.Request.Context(), domainID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}
