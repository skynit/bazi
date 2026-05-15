package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service"

	"github.com/gin-gonic/gin"
)

func setupZiWeiRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	svc := service.NewZiWeiService()
	RegisterZiWeiRoutes(r, svc)
	return r
}

func getValidJWT(t *testing.T) string {
	t.Helper()
	middleware.InitJWT("test-secret")
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return token
}

func TestZiWeiChart_Success(t *testing.T) {
	router := setupZiWeiRouter()
	token := getValidJWT(t)

	body := jsonBody(t, model.ChartRequest{
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var chart service.ZiWeiChart
	if err := json.Unmarshal(w.Body.Bytes(), &chart); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(chart.Palaces) != 12 {
		t.Errorf("expected 12 palaces, got %d", len(chart.Palaces))
	}

	for i, p := range chart.Palaces {
		if p.Name == "" {
			t.Errorf("palace %d has empty name", i)
		}
	}

	if chart.Palaces[0].Name != "命宮" {
		t.Errorf("expected first palace to be 命宮, got %s", chart.Palaces[0].Name)
	}

	if chart.FiveBureau == "" {
		t.Error("FiveBureau is empty")
	}
}

func TestZiWeiChart_Unauthorized(t *testing.T) {
	router := setupZiWeiRouter()

	body := jsonBody(t, model.ChartRequest{
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
