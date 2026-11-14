package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	sslapp "github.com/wuekevin/smartcdn/internal/application/ssl"
	"github.com/wuekevin/smartcdn/internal/infra/provider/schema"
)

// SSLHandler 提供 SSL 凭据中心、部署任务和 Provider 目录 API。
type SSLHandler struct{ svc *sslapp.Service }

func NewSSLHandler(svc *sslapp.Service) *SSLHandler { return &SSLHandler{svc: svc} }

func (h *SSLHandler) RegisterRoutes(r gin.IRouter) {
	g := r.Group("/ssl")

	// Provider 目录（schema 驱动前端表单）
	g.GET("/providers", h.listProviders)

	// 凭据中心
	g.GET("/credentials", h.listCredentials)
	g.POST("/credentials", h.createCredential)
	g.PUT("/credentials/:id", h.updateCredential)
	g.DELETE("/credentials/:id", h.deleteCredential)

	// 部署任务
	g.GET("/certs/:certID/deploy-jobs", h.listDeployJobs)
	g.POST("/certs/:certID/deploy-jobs", h.createDeployJob)
	g.POST("/deploy-jobs/:id/execute", h.executeDeployJob)
	g.DELETE("/deploy-jobs/:id", h.deleteDeployJob)
}

func (h *SSLHandler) listProviders(c *gin.Context) {
	category := c.Query("category")
	var items interface{}
	if category != "" {
		items = schema.CatalogByCategory(category)
	} else {
		items = schema.Catalog()
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *SSLHandler) listCredentials(c *gin.Context) {
	category := c.Query("category")
	items, err := h.svc.ListCredentials(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *SSLHandler) createCredential(c *gin.Context) {
	var body struct {
		Name         string            `json:"name" binding:"required"`
		ProviderType string            `json:"provider_type" binding:"required"`
		Category     string            `json:"category" binding:"required"`
		Credentials  map[string]string `json:"credentials" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.CreateCredential(c.Request.Context(), sslapp.CreateCredentialInput{
		Name:         body.Name,
		ProviderType: body.ProviderType,
		Category:     body.Category,
		Credentials:  body.Credentials,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *SSLHandler) updateCredential(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var body struct {
		Name        string            `json:"name"`
		Credentials map[string]string `json:"credentials"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.UpdateCredential(c.Request.Context(), id, body.Name, body.Credentials)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *SSLHandler) deleteCredential(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteCredential(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *SSLHandler) listDeployJobs(c *gin.Context) {
	certID, _ := strconv.ParseUint(c.Param("certID"), 10, 64)
	items, err := h.svc.ListDeployJobs(c.Request.Context(), certID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *SSLHandler) createDeployJob(c *gin.Context) {
	certID, _ := strconv.ParseUint(c.Param("certID"), 10, 64)
	var body struct {
		Name         string  `json:"name" binding:"required"`
		DeployType   string  `json:"deploy_type" binding:"required"`
		ConfigJSON   string  `json:"config_json" binding:"required"`
		CredentialID *uint64 `json:"credential_id"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.CreateDeployJob(c.Request.Context(), sslapp.CreateDeployJobInput{
		CertID:       certID,
		Name:         body.Name,
		DeployType:   body.DeployType,
		ConfigJSON:   body.ConfigJSON,
		CredentialID: body.CredentialID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *SSLHandler) executeDeployJob(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.ExecuteDeployJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (h *SSLHandler) deleteDeployJob(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.svc.DeleteDeployJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}
