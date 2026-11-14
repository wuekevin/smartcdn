package api

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	cdnsvc "github.com/wuekevin/smartcdn/internal/application/cdn"
)

type CDNHandler struct {
	svc *cdnsvc.Service
}

func NewCDNHandler(svc *cdnsvc.Service) *CDNHandler {
	return &CDNHandler{svc: svc}
}

func (h *CDNHandler) RegisterRoutes(r *gin.RouterGroup) {
	// CDN accounts
	acc := r.Group("/cdn/accounts")
	acc.POST("", h.CreateAccount)
	acc.GET("", h.ListAccounts)
	acc.GET("/:aid", h.GetAccount)
	acc.POST("/:aid/validate", h.ValidateAccount)

	// CDN domains
	dom := r.Group("/cdn/domains")
	dom.POST("", h.AddDomain)
	dom.GET("", h.ListDomains)
	dom.GET("/:did", h.GetDomain)
	dom.DELETE("/:did", h.DeleteDomain)
	dom.POST("/:did/sync", h.SyncDomainStatus)
	dom.POST("/:did/purge-urls", h.PurgeURLs)
	dom.POST("/:did/purge-all", h.PurgeAll)

	// Custom hostnames (SaaS)
	dom.POST("/:did/hostnames", h.CreateCustomHostname)
	dom.GET("/:did/hostnames", h.ListCustomHostnames)
	dom.GET("/:did/hostnames/:hid", h.GetCustomHostname)
	dom.POST("/:did/hostnames/:hid/sync", h.SyncCustomHostname)
	dom.DELETE("/:did/hostnames/:hid", h.DeleteCustomHostname)
}

// POST /api/v1/cdn/accounts
func (h *CDNHandler) CreateAccount(c *gin.Context) {
	var req struct {
		Name         string            `json:"name" binding:"required"`
		ProviderType string            `json:"provider_type" binding:"required"`
		Credentials  map[string]string `json:"credentials" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	credBytes, _ := json.Marshal(req.Credentials)
	v, err := h.svc.CreateAccount(c.Request.Context(), cdnsvc.CreateAccountInput{
		Name:         req.Name,
		ProviderType: req.ProviderType,
		Credentials:  credBytes,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

// GET /api/v1/cdn/accounts
func (h *CDNHandler) ListAccounts(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	views, err := h.svc.ListAccounts(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

// GET /api/v1/cdn/accounts/:aid
func (h *CDNHandler) GetAccount(c *gin.Context) {
	id, err := parseUint64Param(c, "aid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	v, err := h.svc.GetAccount(c.Request.Context(), id)
	if err != nil {
		if err == cdnsvc.ErrAccountNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cdn account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

// POST /api/v1/cdn/accounts/:aid/validate
func (h *CDNHandler) ValidateAccount(c *gin.Context) {
	id, err := parseUint64Param(c, "aid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid account id"})
		return
	}
	if err := h.svc.ValidateAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/v1/cdn/domains
func (h *CDNHandler) AddDomain(c *gin.Context) {
	var req struct {
		AccountID  uint64 `json:"account_id" binding:"required"`
		DomainName string `json:"domain_name" binding:"required"`
		OriginAddr string `json:"origin_addr"`
		OriginPort int    `json:"origin_port"`
		Protocol   string `json:"protocol"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.AddDomain(c.Request.Context(), cdnsvc.AddDomainInput{
		AccountID:  req.AccountID,
		DomainName: req.DomainName,
		OriginAddr: req.OriginAddr,
		OriginPort: req.OriginPort,
		Protocol:   req.Protocol,
	})
	if err != nil {
		if err == cdnsvc.ErrAccountNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cdn account not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

// GET /api/v1/cdn/domains
func (h *CDNHandler) ListDomains(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	views, err := h.svc.ListDomains(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

// GET /api/v1/cdn/domains/:did
func (h *CDNHandler) GetDomain(c *gin.Context) {
	id, err := parseUint64Param(c, "did")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	v, err := h.svc.GetDomain(c.Request.Context(), id)
	if err != nil {
		if err == cdnsvc.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cdn domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

// DELETE /api/v1/cdn/domains/:did
func (h *CDNHandler) DeleteDomain(c *gin.Context) {
	id, err := parseUint64Param(c, "did")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	if err := h.svc.DeleteDomain(c.Request.Context(), id); err != nil {
		if err == cdnsvc.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cdn domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// POST /api/v1/cdn/domains/:did/sync
func (h *CDNHandler) SyncDomainStatus(c *gin.Context) {
	id, err := parseUint64Param(c, "did")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	v, err := h.svc.SyncDomainStatus(c.Request.Context(), id)
	if err != nil {
		if err == cdnsvc.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cdn domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

// POST /api/v1/cdn/domains/:did/purge-urls
func (h *CDNHandler) PurgeURLs(c *gin.Context) {
	id, err := parseUint64Param(c, "did")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var body struct {
		URLs []string `json:"urls" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.PurgeURLs(c.Request.Context(), id, body.URLs); err != nil {
		if err == cdnsvc.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cdn domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/v1/cdn/domains/:did/purge-all
func (h *CDNHandler) PurgeAll(c *gin.Context) {
	id, err := parseUint64Param(c, "did")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	if err := h.svc.PurgeAll(c.Request.Context(), id); err != nil {
		if err == cdnsvc.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cdn domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// POST /api/v1/cdn/domains/:did/hostnames
func (h *CDNHandler) CreateCustomHostname(c *gin.Context) {
	domainID, err := parseUint64Param(c, "did")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var body struct {
		Hostname string `json:"hostname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.CreateCustomHostname(c.Request.Context(), cdnsvc.CreateHostnameInput{
		CDNDomainID: domainID,
		Hostname:    body.Hostname,
	})
	if err != nil {
		if err == cdnsvc.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "cdn domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

// GET /api/v1/cdn/domains/:did/hostnames
func (h *CDNHandler) ListCustomHostnames(c *gin.Context) {
	domainID, err := parseUint64Param(c, "did")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	views, err := h.svc.ListCustomHostnames(c.Request.Context(), domainID, afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

// GET /api/v1/cdn/domains/:did/hostnames/:hid
func (h *CDNHandler) GetCustomHostname(c *gin.Context) {
	id, err := parseUint64Param(c, "hid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hostname id"})
		return
	}
	v, err := h.svc.GetCustomHostname(c.Request.Context(), id)
	if err != nil {
		if err == cdnsvc.ErrHostnameNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "hostname not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

// POST /api/v1/cdn/domains/:did/hostnames/:hid/sync
func (h *CDNHandler) SyncCustomHostname(c *gin.Context) {
	id, err := parseUint64Param(c, "hid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hostname id"})
		return
	}
	v, err := h.svc.SyncCustomHostname(c.Request.Context(), id)
	if err != nil {
		if err == cdnsvc.ErrHostnameNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "hostname not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

// DELETE /api/v1/cdn/domains/:did/hostnames/:hid
func (h *CDNHandler) DeleteCustomHostname(c *gin.Context) {
	id, err := parseUint64Param(c, "hid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid hostname id"})
		return
	}
	if err := h.svc.DeleteCustomHostname(c.Request.Context(), id); err != nil {
		if err == cdnsvc.ErrHostnameNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "hostname not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

