package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"

	"github.com/wuekevin/smartcdn/internal/domain/entity"
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
	g.POST("", h.Create)
	g.GET("", h.List)
	g.DELETE("/:id", h.Delete)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=64"`
		Password string `json:"password" binding:"required,min=6"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "password hash failed"})
		return
	}
	u := &entity.User{Username: req.Username, Password: string(hash), Email: req.Email, IsActive: true}
	if err := h.users.Create(c.Request.Context(), u); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"id": u.ID, "username": u.Username, "email": u.Email,
		"is_active": u.IsActive, "created_at": u.CreatedAt,
	})
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
