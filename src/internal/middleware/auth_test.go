package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func init() {
	gin.SetMode(gin.TestMode)
	InitJWT("test-secret-key")
}

func TestGenerateToken(t *testing.T) {
	token, err := GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateToken(t *testing.T) {
	userID := uint(42)
	username := "alice"
	tokenStr, err := GenerateToken(userID, username)
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	customClaims := &claims{}
	parsed, err := jwt.ParseWithClaims(tokenStr, customClaims, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		t.Fatalf("ParseWithClaims failed: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("expected valid token")
	}
	if customClaims.UserID != userID {
		t.Fatalf("expected user_id %d, got %d", userID, customClaims.UserID)
	}
	if customClaims.Username != username {
		t.Fatalf("expected username %q, got %q", username, customClaims.Username)
	}
}

func TestExpiredToken(t *testing.T) {
	now := time.Now()
	c := claims{
		UserID:   1,
		Username: "expired",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
		},
	}
	tokenStr, err := jwt.NewWithClaims(jwt.SigningMethodHS256, c).SignedString(jwtKey)
	if err != nil {
		t.Fatalf("failed to create expired token: %v", err)
	}

	_, err = jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err == nil {
		t.Fatal("expected error for expired token")
	}
}

func TestMissingToken(t *testing.T) {
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}
