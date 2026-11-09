package api

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterSPA registers the React SPA catch-all route.
// frontendFS must be an embed.FS sub-rooted at the dist/ directory.
func RegisterSPA(r *gin.Engine, frontendFS fs.FS) {
	fileServer := http.FileServer(http.FS(frontendFS))

	// Serve static assets; fall through to index.html for SPA routing.
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// API routes should 404, not serve index.html
		if strings.HasPrefix(path, "/api/") {
			c.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND"})
			return
		}

		// Try to serve as a static file
		if _, err := fs.Stat(frontendFS, strings.TrimPrefix(path, "/")); err == nil {
			fileServer.ServeHTTP(c.Writer, c.Request)
			return
		}

		// Fall back to index.html for client-side routing
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
