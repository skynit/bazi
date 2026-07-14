package handler

import (
	"net/http"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

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
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	if username == "" || req.Password == "" || email == "" {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "username, email, and password are required")
		return
	}
	if !validUsername(username) {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "username must be 3-32 characters and contain only letters, numbers, _, -, or .")
		return
	}
	if !validEmail(email) {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid email address")
		return
	}
	if utf8.RuneCountInString(req.Password) < 8 {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "password must be at least 8 characters")
		return
	}
	if h == nil || h.Store == nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceDisabled, "service not available")
		return
	}

	// Check for duplicate username.
	if existing, _ := h.Store.FindByUsername(username); existing != nil {
		respondError(c, http.StatusConflict, ErrCodeConflict, "username already exists")
		return
	}

	user := &model.User{
		Username: username,
		Email:    email,
	}
	if err := user.SetPassword(req.Password); err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to hash password")
		return
	}

	if err := h.Store.Create(user); err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to create user")
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to generate token")
		return
	}

	respondJSON(c, http.StatusCreated, gin.H{
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
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	if h == nil || h.Store == nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceDisabled, "service not available")
		return
	}
	user, err := h.Store.FindByUsername(strings.TrimSpace(req.Username))
	if err != nil || user == nil {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "invalid username or password")
		return
	}

	if !user.CheckPassword(req.Password) {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "invalid username or password")
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Username)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to generate token")
		return
	}

	respondJSON(c, http.StatusOK, model.LoginResponse{Token: token})
}

// Me handles GET /api/auth/me.
// Requires AuthMiddleware to be applied on the route.
func (h *AuthHandler) Me(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	userID, ok := userIDVal.(uint)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "invalid user id in context")
		return
	}

	user, err := h.Store.FindByID(userID)
	if err != nil || user == nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "user not found")
		return
	}

	respondJSON(c, http.StatusOK, gin.H{
		"user": model.RegisterResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}

func validUsername(username string) bool {
	count := utf8.RuneCountInString(username)
	if count < 3 || count > 32 {
		return false
	}
	for _, r := range username {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '-' || r == '.' {
			continue
		}
		return false
	}
	return true
}

func validEmail(value string) bool {
	address, err := mail.ParseAddress(value)
	return err == nil && address.Address == value && strings.Contains(value, "@")
}
