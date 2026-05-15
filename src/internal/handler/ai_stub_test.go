package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"

	"github.com/gin-gonic/gin"
)

func setupAIStubRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	h := &AIStubHandler{}

	fortune := r.Group("/api/fortune")
	fortune.Use(middleware.AuthMiddleware())
	{
		fortune.POST("/ai", h.AnalyzeFortune)
	}
	return r
}

func TestAIStubAnalyzeFortune(t *testing.T) {
	router := setupAIStubRouter()

	// Generate valid JWT
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/ai", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.AIFortuneStubResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Status != "coming_soon" {
		t.Fatalf("expected status coming_soon, got %s", resp.Status)
	}
	if resp.Message != "AI分析功能即将上线" {
		t.Fatalf("expected message AI分析功能即将上线, got %s", resp.Message)
	}
}
