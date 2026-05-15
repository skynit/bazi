package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service"

	"github.com/gin-gonic/gin"
)

func setupChartRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	h := &ChartHandler{
		Parser: &service.InputParser{},
		Bazi:   &service.BaziService{},
	}
	r.POST("/api/chart", middleware.AuthMiddleware(), h.Chart)
	return r
}

func chartJSONBody(t *testing.T, v interface{}) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return strings.NewReader(string(b))
}

func TestChartValidSolar(t *testing.T) {
	router := setupChartRouter()
	token, _ := middleware.GenerateToken(1, "testuser")

	body := chartJSONBody(t, model.ChartRequest{
		BirthYear:    1990,
		BirthMonth:   5,
		BirthDay:     15,
		BirthHour:    8,
		BirthMin:     0,
		CalendarType: "SOLAR",
		Gender:       "MALE",
		Name:         "Test",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.ChartResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.YearPillar.Gan == "" || resp.YearPillar.Zhi == "" {
		t.Error("expected non-empty year pillar")
	}
	if resp.DayPillar.Gan == "" || resp.DayPillar.Zhi == "" {
		t.Error("expected non-empty day pillar")
	}
}

func TestChartMissingJWT(t *testing.T) {
	router := setupChartRouter()

	body := chartJSONBody(t, model.ChartRequest{
		BirthYear:    1990,
		BirthMonth:   5,
		BirthDay:     15,
		BirthHour:    8,
		CalendarType: "SOLAR",
		Gender:       "MALE",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chart", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChartInvalidInput(t *testing.T) {
	router := setupChartRouter()
	token, _ := middleware.GenerateToken(1, "testuser")

	body := chartJSONBody(t, model.ChartRequest{
		BirthYear:    1990,
		BirthMonth:   13,
		BirthDay:     15,
		BirthHour:    8,
		CalendarType: "SOLAR",
		Gender:       "MALE",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
