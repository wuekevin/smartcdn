package api

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"fmt"
)

const (
	csrfCookieName = "csrf_token"
	csrfHeaderName = "X-CSRF-Token"
	csrfTokenTTL   = 2 * time.Hour
)

func csrfRedisKey(token string) string {
	return fmt.Sprintf("csrf:%s", token)
}

// CSRFMiddleware issues and validates CSRF tokens using Redis storage.
// GET/HEAD/OPTIONS 请求豁免，其余请求必须携带有效 CSRF Token。
func CSRFMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			// 对只读请求发放或刷新 CSRF Token
			ensureCSRFToken(c, rdb)
			c.Next()
			return
		}

		token := c.GetHeader(csrfHeaderName)
		if token == "" {
			token, _ = c.Cookie(csrfCookieName)
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "CSRF_MISSING"})
			return
		}

		exists, _ := rdb.Exists(c.Request.Context(), csrfRedisKey(token)).Result()
		if exists == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": "CSRF_INVALID"})
			return
		}

		c.Next()
	}
}

func ensureCSRFToken(c *gin.Context, rdb *redis.Client) {
	existing, err := c.Cookie(csrfCookieName)
	if err == nil && existing != "" {
		return
	}

	token := generateToken(32)
	rdb.Set(c.Request.Context(), csrfRedisKey(token), "1", csrfTokenTTL)
	c.SetCookie(csrfCookieName, token, int(csrfTokenTTL.Seconds()), "/", "", false, false)
	c.Header("X-CSRF-Token", token)
}

func generateToken(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
