package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	certapp "github.com/wuekevin/smartcdn/internal/application/certificate"
	"github.com/wuekevin/smartcdn/internal/domain/entity"
)

// CertificateHandler exposes SSL certificate management APIs.
type CertificateHandler struct {
	svc *certapp.Service
}

func NewCertificateHandler(svc *certapp.Service) *CertificateHandler {
	return &CertificateHandler{svc: svc}
}

func (h *CertificateHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/certificates")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.DELETE("/:id", h.Delete)
	g.POST("/:id/retrigger", h.Retrigger)
	g.GET("/:id/deploy-targets", h.ListDeployTargets)
	g.POST("/:id/deploy-targets", h.AddDeployTarget)
	g.DELETE("/:id/deploy-targets/:tid", h.DeleteDeployTarget)
}

type createCertReq struct {
	DomainID     uint64   `json:"domain_id" binding:"required"`
	DomainName   string   `json:"domain_name" binding:"required,max=253"`
	SANs         []string `json:"sans"`
	ACMEProvider string   `json:"acme_provider"` // letsencrypt | zerossl
	Challenge    string   `json:"challenge"`     // dns-01 | http-01
	AccountEmail string   `json:"account_email" binding:"required,email"`
	AutoRenew    bool     `json:"auto_renew"`
	DNSAccountID *uint64  `json:"dns_account_id"`
	Staging      bool     `json:"staging"`
}

func (h *CertificateHandler) Create(c *gin.Context) {
	var req createCertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	v, err := h.svc.CreateCertificate(c.Request.Context(), certapp.IssueCertInput{
		DomainID:     req.DomainID,
		DomainName:   req.DomainName,
		SANs:         req.SANs,
		ACMEProvider: req.ACMEProvider,
		Challenge:    entity.ACMEChallenge(req.Challenge),
		AccountEmail: req.AccountEmail,
		AutoRenew:    req.AutoRenew,
		DNSAccountID: req.DNSAccountID,
		Staging:      req.Staging,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *CertificateHandler) List(c *gin.Context) {
	domainID, err := parseUint64Query(c, "domain_id")
	if err != nil || domainID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain_id required"})
		return
	}
	views, err := h.svc.ListCertificates(c.Request.Context(), domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": views})
}

func (h *CertificateHandler) Get(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	v, err := h.svc.GetCertificate(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, certapp.ErrCertNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, v)
}

func (h *CertificateHandler) Delete(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := h.svc.DeleteCertificate(c.Request.Context(), id); err != nil {
		if errors.Is(err, certapp.ErrCertNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type retriggerReq struct {
	Staging bool `json:"staging"`
}

func (h *CertificateHandler) Retrigger(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req retriggerReq
	_ = c.ShouldBindJSON(&req)
	if err := h.svc.RetriggerIssuance(c.Request.Context(), id, req.Staging); err != nil {
		if errors.Is(err, certapp.ErrCertNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"message": "certificate issuance triggered"})
}

func (h *CertificateHandler) ListDeployTargets(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	targets, err := h.svc.ListDeployTargets(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": targets})
}

type addDeployTargetReq struct {
	TargetType string `json:"target_type" binding:"required"` // cdn_node | cdn_provider | nginx_local
	TargetID   string `json:"target_id" binding:"required"`
}

func (h *CertificateHandler) AddDeployTarget(c *gin.Context) {
	id, err := parseUint64Param(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req addDeployTargetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	t, err := h.svc.AddDeployTarget(c.Request.Context(), certapp.AddDeployTargetInput{
		CertID:     id,
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, t)
}

func (h *CertificateHandler) DeleteDeployTarget(c *gin.Context) {
	tid, err := parseUint64Param(c, "tid")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid deploy target id"})
		return
	}
	if err := h.svc.DeleteDeployTarget(c.Request.Context(), tid); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
