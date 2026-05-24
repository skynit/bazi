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

// mockWeeklyChartStore implements ChartStore for weekly fortune tests.
type mockWeeklyChartStore struct {
	chart *model.BirthChart
}

func (m *mockWeeklyChartStore) FindByID(id uint) (*model.BirthChart, error) {
	if m.chart != nil && m.chart.ID == id {
		return m.chart, nil
	}
	return nil, nil
}
func (m *mockWeeklyChartStore) Update(chart *model.BirthChart) error {
	m.chart = chart
	return nil
}

func setupWeeklyRouter(store ChartStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	h := &WeeklyFortuneHandler{
		Engine: service.NewFortuneEngine(),
		Charts: store,
	}

	r.POST("/api/fortune/weekly", middleware.AuthMiddleware(), h.Weekly)
	return r
}

func TestWeeklyFortune(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1990,
		BirthMonth: 1,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	}
	chart.ID = 1

	store := &mockWeeklyChartStore{chart: chart}
	router := setupWeeklyRouter(store)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	reqBody := model.WeeklyFortuneRequest{
		ChartID:   1,
		StartDate: "2025-01-06",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/weekly", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.WeeklyFortuneResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.DailyFortunes) != 7 {
		t.Errorf("expected 7 daily fortunes, got %d", len(resp.DailyFortunes))
	}
}

func TestWeeklyFortuneNoJWT(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1990,
		BirthMonth: 1,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	}
	chart.ID = 1

	store := &mockWeeklyChartStore{chart: chart}
	router := setupWeeklyRouter(store)

	reqBody := model.WeeklyFortuneRequest{
		ChartID:   1,
		StartDate: "2025-01-06",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/weekly", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
