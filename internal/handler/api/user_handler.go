package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/wuekevin/smartcdn/internal/infra/repo"
)

type UserHandler struct {
	users *repo.UserRepo
}

func NewUserHandler(users *repo.UserRepo) *UserHandler {
	return &UserHandler{users: users}
}

func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	g := r.Group("/users")
	g.GET("", h.List)
	g.DELETE("/:id", h.Delete)
}

func (h *UserHandler) List(c *gin.Context) {
	afterID, _ := parseIntQuery(c, "after_id", 0)
	limit, _ := parseIntQuery(c, "limit", 20)

	users, err := h.users.List(c.Request.Context(), uint64(afterID), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	type userView struct {
		ID        uint64    `json:"id"`
		Username  string    `json:"username"`
		Email     string    `json:"email"`
		IsActive  bool      `json:"is_active"`
		CreatedAt time.Time `json:"created_at"`
	}
	items := make([]userView, 0, len(users))
	for _, u := range users {
		items = append(items, userView{
			ID:        u.ID,
			Username:  u.Username,
			Email:     u.Email,
			IsActive:  u.IsActive,
			CreatedAt: u.CreatedAt,
		})
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "invalid id"})
		return
	}
	callerID, _ := GetUserID(c)
	if id == callerID {
		c.JSON(http.StatusBadRequest, gin.H{"message": "cannot delete yourself"})
		return
	}
	if err := h.users.SoftDelete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
