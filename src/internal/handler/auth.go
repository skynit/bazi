package handler

import (
	"net/http"

	"bazi/internal/middleware"
	"bazi/internal/model"

	"github.com/gin-gonic/gin"
)

// UserStore defines the interface for user persistence.
// A mock implementation is used in tests; a GORM implementation
// will be provided later for production.
type UserStore interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByID(id uint) (*model.User, error)
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	Store UserStore
}

// Register handles POST /api/auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username, email, and password are required"})
		return
	}

	// Check for duplicate username.
	if existing, _ := h.Store.FindByUsername(req.Username); existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
		return
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
	}
	if err := user.SetPassword(req.Password); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.Store.Create(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user": model.RegisterResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
		"token": token,
	})
}

// Login handles POST /api/auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.Store.FindByUsername(req.Username)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid username or password"})
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, model.LoginResponse{Token: token})
}

// Me handles GET /api/auth/me.
// Requires AuthMiddleware to be applied on the route.
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id in context"})
		return
	}

	user, err := h.Store.FindByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, model.RegisterResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}
