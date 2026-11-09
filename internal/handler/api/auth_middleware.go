package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/wuekevin/smartcdn/internal/application/auth"
)

const (
	ctxUserID   = "user_id"
	ctxUsername = "username"
)

// JWTMiddleware validates Bearer token and injects user_id + username into the context.
func JWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED"})
			return
		}

		claims, err := auth.ParseToken(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "token invalid or expired"})
			return
		}

		c.Set(ctxUserID, claims.UserID)
		c.Set(ctxUsername, claims.Username)
		c.Next()
	}
}

// RequirePermission returns a middleware that enforces resource:action access (AD-9).
// Unauthorized → 403 (never 404).
func RequirePermission(perms *auth.PermissionService, resource, action string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDRaw, exists := c.Get(ctxUserID)
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN"})
			return
		}
		userID, ok := userIDRaw.(uint64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN"})
			return
		}

		allowed, err := perms.HasPermission(c.Request.Context(), userID, resource, action)
		if err != nil || !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "FORBIDDEN"})
			return
		}

		c.Next()
	}
}

// GetUserID extracts user_id from the Gin context (set by JWTMiddleware).
func GetUserID(c *gin.Context) (uint64, bool) {
	v, ok := c.Get(ctxUserID)
	if !ok {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}
