package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SetupLockMiddleware blocks all /api/v1/setup/* routes when install.lock exists.
func SetupLockMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, err := os.Stat("install.lock"); err == nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "SETUP_LOCKED"})
			return
		}
		c.Next()
	}
}

// RequestIDMiddleware injects a unique request_id into the context.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = newRequestID()
		}
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

// SecurityHeadersMiddleware adds standard security response headers.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	}
}

func newRequestID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
