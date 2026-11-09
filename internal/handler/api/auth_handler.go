package api

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/wuekevin/smartcdn/internal/application/auth"
)

type AuthHandler struct {
	svc *auth.Service
}

func NewAuthHandler(svc *auth.Service) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) RegisterAuth(rg *gin.RouterGroup) {
	g := rg.Group("/auth")
	g.POST("/login", h.Login)
}

// Login authenticates a user and returns a JWT.
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"code": "VALIDATION_ERROR", "message": err.Error()})
		return
	}

	result, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrAccountLocked):
			c.JSON(http.StatusTooManyRequests, gin.H{"code": "ACCOUNT_LOCKED", "message": err.Error()})
		case errors.Is(err, auth.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"code": "INVALID_CREDENTIALS"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"code": "INTERNAL_ERROR"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":      result.Token,
		"expires_at": result.ExpiresAt,
	})
}
