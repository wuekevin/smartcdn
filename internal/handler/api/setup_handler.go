package api

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wuekevin/smartcdn/internal/application/setup"
	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

type SetupHandler struct {
	svc *setup.Service
}

func NewSetupHandler(svc *setup.Service) *SetupHandler {
	return &SetupHandler{svc: svc}
}

// RegisterSetup wires /api/v1/setup/* routes.
// Must be called before the install.lock middleware is attached.
func (h *SetupHandler) RegisterSetup(rg *gin.RouterGroup) {
	g := rg.Group("/setup")
	g.GET("/check", h.Check)
	g.POST("/complete", h.Complete)
}

// Check runs environment detection and returns per-item results.
func (h *SetupHandler) Check(c *gin.Context) {
	if h.svc.IsInstalled() {
		c.JSON(http.StatusForbidden, gin.H{"code": "SETUP_LOCKED"})
		return
	}

	var req struct {
		DBDSN     string `form:"db_dsn"`
		RedisAddr string `form:"redis_addr"`
	}
	_ = c.ShouldBindQuery(&req)

	result := h.svc.Check(c.Request.Context(), req.DBDSN, req.RedisAddr)
	c.JSON(http.StatusOK, result)
}

// Complete finalizes installation and writes install.lock.
func (h *SetupHandler) Complete(c *gin.Context) {
	if h.svc.IsInstalled() {
		c.JSON(http.StatusForbidden, gin.H{"code": "SETUP_LOCKED"})
		return
	}

	var input entity.SetupCompleteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	if err := h.svc.Complete(c.Request.Context(), input); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": "SETUP_FAILED", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"installed": true})
}
