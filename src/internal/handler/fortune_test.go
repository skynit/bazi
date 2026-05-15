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

// mockChartStore implements ChartStore for testing.
type mockChartStore struct {
	chart *model.BirthChart
}

func (m *mockChartStore) FindByID(id uint) (*model.BirthChart, error) {
	if m.chart != nil && m.chart.ID == id {
		return m.chart, nil
	}
	return nil, nil
}

func setupFortuneRouter(store ChartStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	h := &FortuneHandler{
		Engine:     service.NewFortuneEngine(),
		ChartStore: store,
	}

	r.POST("/api/fortune", middleware.AuthMiddleware(), h.CalculateDaily)
	return r
}

func fortuneJSONBody(t *testing.T, v interface{}) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	return strings.NewReader(string(b))
}

func TestCalculateDailyValid(t *testing.T) {
	chart := &model.BirthChart{}
	chart.ID = 1
	chart.DayPillar = json.RawMessage(`{"gan":"甲","zhi":"子"}`)

	store := &mockChartStore{chart: chart}
	router := setupFortuneRouter(store)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	body := fortuneJSONBody(t, model.FortuneRequest{
		ChartID:   1,
		QueryDate: "2025-01-15",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fortune", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.FortuneResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.SolarDate == "" {
		t.Fatal("expected non-empty solar_date in response")
	}
	if resp.DayGanZhi == "" {
		t.Fatal("expected non-empty day_gan_zhi in response")
	}
}

func TestCalculateDailyNoJWT(t *testing.T) {
	chart := &model.BirthChart{}
	chart.ID = 1
	chart.DayPillar = json.RawMessage(`{"gan":"甲","zhi":"子"}`)

	store := &mockChartStore{chart: chart}
	router := setupFortuneRouter(store)

	body := fortuneJSONBody(t, model.FortuneRequest{
		ChartID:   1,
		QueryDate: "2025-01-15",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fortune", body)
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
