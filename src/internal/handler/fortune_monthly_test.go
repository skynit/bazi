package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service"

	"github.com/gin-gonic/gin"
)

// mockMonthlyChartStore implements MonthlyChartStore for testing.
type mockMonthlyChartStore struct {
	chart *model.BirthChart
}

func (m *mockMonthlyChartStore) FindByID(id uint) (*model.BirthChart, error) {
	if m.chart != nil && id == m.chart.ID {
		return m.chart, nil
	}
	return nil, fmt.Errorf("chart not found")
}

func setupMonthlyTestRouter(store MonthlyChartStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	h := &MonthlyFortuneHandler{
		ChartStore: store,
		Engine:     service.NewFortuneEngine(),
	}

	fortune := r.Group("/api/fortune")
	fortune.Use(middleware.AuthMiddleware())
	{
		fortune.POST("/monthly", h.HandleMonthly)
	}
	return r
}

func TestMonthlyFortuneReturnsDailyFortunesForCorrectMonth(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1990,
		BirthMonth: 1,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	}
	chart.ID = 1

	store := &mockMonthlyChartStore{chart: chart}
	router := setupMonthlyTestRouter(store)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	reqBody := model.MonthlyFortuneRequest{
		ChartID: 1,
		Year:    2024,
		Month:   6,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/monthly", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.MonthlyFortuneResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// June has 30 days
	if len(resp.DailyFortunes) != 30 {
		t.Errorf("DailyFortunes has %d items, want 30", len(resp.DailyFortunes))
	}

	if resp.WeeklyScore < 0 || resp.WeeklyScore > 100 {
		t.Errorf("WeeklyScore = %d, want in [0, 100]", resp.WeeklyScore)
	}

	for i, df := range resp.DailyFortunes {
		if df.SolarDate == "" {
			t.Errorf("DailyFortunes[%d].SolarDate is empty", i)
		}
		if df.DayGanZhi == "" {
			t.Errorf("DailyFortunes[%d].DayGanZhi is empty", i)
		}
	}
}

func TestMonthlyFortuneNoJWTReturns401(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1990,
		BirthMonth: 1,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	}
	chart.ID = 1

	store := &mockMonthlyChartStore{chart: chart}
	router := setupMonthlyTestRouter(store)

	reqBody := model.MonthlyFortuneRequest{
		ChartID: 1,
		Year:    2024,
		Month:   6,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/monthly", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
