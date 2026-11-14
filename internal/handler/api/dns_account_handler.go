package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/wuekevin/smartcdn/internal/application/provider"
	"github.com/wuekevin/smartcdn/internal/domain/port"
)

type DNSAccountHandler struct {
	svc *provider.Service
}

func NewDNSAccountHandler(svc *provider.Service) *DNSAccountHandler {
	return &DNSAccountHandler{svc: svc}
}

func (h *DNSAccountHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/dns-accounts")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
	g.PUT("/:id/credentials", h.RotateCredentials)
	g.PUT("/:id/refresh-capabilities", h.RefreshCapabilities)
	g.DELETE("/:id", h.Delete)
}

type createAccountReq struct {
	Name         string            `json:"name" binding:"required,max=128"`
	ProviderType string            `json:"provider_type" binding:"required"`
	Credentials  map[string]string `json:"credentials" binding:"required"`
}

func (h *DNSAccountHandler) Create(c *gin.Context) {
	var req createAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	acc, err := h.svc.CreateAccount(c.Request.Context(), req.Name, port.ProviderType(req.ProviderType), req.Credentials)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	view, _ := h.svc.GetAccount(c.Request.Context(), acc.ID)
	c.JSON(http.StatusCreated, view)
}

func (h *DNSAccountHandler) List(c *gin.Context) {
	afterID, _ := strconv.ParseUint(c.Query("after_id"), 10, 64)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	views, err := h.svc.ListAccounts(c.Request.Context(), afterID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": views})
}

func (h *DNSAccountHandler) Get(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	view, err := h.svc.GetAccount(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, view)
}

type updateAccountReq struct {
	Name     string `json:"name" binding:"required,max=128"`
	IsActive bool   `json:"is_active"`
}

func (h *DNSAccountHandler) Update(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateAccountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	view, err := h.svc.UpdateAccount(c.Request.Context(), id, req.Name, req.IsActive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, view)
}

type rotateCredsReq struct {
	Credentials map[string]string `json:"credentials" binding:"required"`
}

func (h *DNSAccountHandler) RotateCredentials(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req rotateCredsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.RotateCredentials(c.Request.Context(), id, req.Credentials); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "credentials updated"})
}

func (h *DNSAccountHandler) RefreshCapabilities(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	caps, err := h.svc.RefreshCapabilities(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, caps)
}

func (h *DNSAccountHandler) Delete(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteAccount(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseUint64Param(c *gin.Context, name string) (uint64, error) {
	return strconv.ParseUint(c.Param(name), 10, 64)
}
