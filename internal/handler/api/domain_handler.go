package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	domainsvc "github.com/wuekevin/smartcdn/internal/application/domain"
)

type DomainHandler struct {
	svc *domainsvc.Service
}

func NewDomainHandler(svc *domainsvc.Service) *DomainHandler {
	return &DomainHandler{svc: svc}
}

func (h *DomainHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/domains")
	g.POST("", h.CreateDomain)
	g.GET("", h.ListDomains)
	g.GET("/:id", h.GetDomain)
	g.DELETE("/:id", h.DeleteDomain)

	// DNS records
	g.POST("/:id/records", h.CreateRecord)
	g.GET("/:id/records", h.ListRecords)
	g.PUT("/:id/records/:rid", h.UpdateRecord)
	g.DELETE("/:id/records/:rid", h.DeleteRecord)

	// Snapshots
	g.POST("/:id/snapshots", h.CreateSnapshot)
	g.GET("/:id/snapshots", h.ListSnapshots)
}

// POST /api/v1/domains
func (h *DomainHandler) CreateDomain(c *gin.Context) {
	var input domainsvc.CreateDomainInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.CreateDomain(c.Request.Context(), input)
	if err != nil {
		if err == domainsvc.ErrDomainExists {
			c.JSON(http.StatusConflict, gin.H{"error": "domain already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

// GET /api/v1/domains
func (h *DomainHandler) ListDomains(c *gin.Context) {
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	views, err := h.svc.ListDomains(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

// GET /api/v1/domains/:id
func (h *DomainHandler) GetDomain(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	v, err := h.svc.GetDomain(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}
	c.JSON(http.StatusOK, v)
}

// DELETE /api/v1/domains/:id
func (h *DomainHandler) DeleteDomain(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteDomain(c.Request.Context(), id); err != nil {
		if err == domainsvc.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// POST /api/v1/domains/:id/records
func (h *DomainHandler) CreateRecord(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var input domainsvc.CreateRecordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.CreateRecord(c.Request.Context(), domainID, input)
	if err != nil {
		if err == domainsvc.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

// GET /api/v1/domains/:id/records
func (h *DomainHandler) ListRecords(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 50)
	views, err := h.svc.ListRecords(c.Request.Context(), domainID, afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

// PUT /api/v1/domains/:id/records/:rid
func (h *DomainHandler) UpdateRecord(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	recordID, err := parseUint64Param(c, "rid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid record id"})
		return
	}
	var input domainsvc.UpdateRecordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.UpdateRecord(c.Request.Context(), domainID, recordID, input)
	if err != nil {
		if err == domainsvc.ErrRecordNotFound || err == domainsvc.ErrDomainNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

// DELETE /api/v1/domains/:id/records/:rid
func (h *DomainHandler) DeleteRecord(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	recordID, err := parseUint64Param(c, "rid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid record id"})
		return
	}
	if err := h.svc.DeleteRecord(c.Request.Context(), domainID, recordID); err != nil {
		if err == domainsvc.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "record not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// POST /api/v1/domains/:id/snapshots
func (h *DomainHandler) CreateSnapshot(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	var body struct {
		TriggerEvent string `json:"trigger_event"`
	}
	_ = c.ShouldBindJSON(&body)
	if body.TriggerEvent == "" {
		body.TriggerEvent = "manual"
	}
	v, err := h.svc.CreateSnapshot(c.Request.Context(), domainID, body.TriggerEvent)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

// GET /api/v1/domains/:id/snapshots
func (h *DomainHandler) ListSnapshots(c *gin.Context) {
	domainID, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid domain id"})
		return
	}
	afterID, _ := parseUint64Query(c, "after_id")
	limit, _ := parseIntQuery(c, "limit", 20)
	views, err := h.svc.ListSnapshots(c.Request.Context(), domainID, afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

// --- helpers ---

func parseUint64Query(c *gin.Context, key string) (uint64, error) {
	s := c.Query(key)
	if s == "" {
		return 0, nil
	}
	return strconv.ParseUint(s, 10, 64)
}

func parseIntQuery(c *gin.Context, key string, def int) (int, error) {
	s := c.Query(key)
	if s == "" {
		return def, nil
	}
	var v int
	_, err := fmt.Sscan(s, &v)
	return v, err
}
