package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DashboardHandler serves aggregate stats for the front-end overview page.
type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/dashboard/overview", h.Overview)
}

// Overview returns high-level counts for the main dashboard.
func (h *DashboardHandler) Overview(c *gin.Context) {
	var domainCount, nodeCount, certCount, firingCount int64

	h.db.Table("domains").Where("deleted_at IS NULL").Count(&domainCount)
	h.db.Table("cdn_nodes").Where("deleted_at IS NULL").Count(&nodeCount)
	h.db.Table("certificates").Where("deleted_at IS NULL").Count(&certCount)
	h.db.Table("alerts").Where("status = ?", "firing").Count(&firingCount)

	c.JSON(http.StatusOK, gin.H{
		"domain_count":        domainCount,
		"node_count":          nodeCount,
		"certificate_count":   certCount,
		"firing_alert_count":  firingCount,
	})
}
