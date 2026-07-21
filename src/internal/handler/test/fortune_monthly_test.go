package handler_test

import (
	. "bazi/internal/handler"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service/fortune"

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
func (m *mockMonthlyChartStore) FindByIDForUser(id uint, userID uint) (*model.BirthChart, error) {
	if m.chart != nil && id == m.chart.ID && (m.chart.UserID == 0 || m.chart.UserID == userID) {
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
		Engine:     fortune.NewFortuneEngine(),
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
	responseBody := w.Body.String()
	for _, required := range []string{
		`"structural_relation_index"`, `"highest_index_day"`, `"lowest_index_day"`,
		`"average_index"`, `"index_standard_deviation"`, `"season_element"`, `"ten_god"`,
		`"seasonal_state"`, `"fortune_layers"`,
	} {
		if !strings.Contains(responseBody, required) {
			t.Fatalf("monthly response missing %s: %s", required, responseBody)
		}
	}
	for _, forbidden := range []string{
		`"monthly_score"`, `"best_day"`, `"worst_day"`, `"peak_days"`,
		`"low_days"`, `"good_streak"`, `"bad_streak"`, `"key_advice"`,
		`"flow_impact"`, `"season_element_advice"`, `"today_ten_god"`,
		`"ten_god_favorable"`, `"ten_god_desc"`,
		`"dayun_influence"`, `"liunian_influence"`, `"advance_retreat"`,
		`"yongshen_impact"`, `"pattern_favorable"`, `"pattern_unfavorable"`,
		`"tiao_hou"`,
	} {
		if strings.Contains(responseBody, forbidden) {
			t.Fatalf("monthly response leaked legacy outcome field %s", forbidden)
		}
	}

	// June has 30 days
	if len(resp.DailyFortunes) != 30 {
		t.Errorf("DailyFortunes has %d items, want 30", len(resp.DailyFortunes))
	}

	if resp.StructuralRelationIndex < 0 || resp.StructuralRelationIndex > 100 {
		t.Errorf("StructuralRelationIndex = %d, want in [0, 100]", resp.StructuralRelationIndex)
	}

	for i, df := range resp.DailyFortunes {
		if df.BaziEngineVersion == "" || df.BaziResolutionSource != "normalized_raw_birth" {
			t.Fatalf("DailyFortunes[%d] missing bazi trace metadata: engine=%q source=%q", i, df.BaziEngineVersion, df.BaziResolutionSource)
		}
		if df.SolarDate == "" {
			t.Errorf("DailyFortunes[%d].SolarDate is empty", i)
		}
		if df.DayGanZhi == "" {
			t.Errorf("DailyFortunes[%d].DayGanZhi is empty", i)
		}
		if df.SeasonalState.Status != "observed" || df.SeasonalState.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("DailyFortunes[%d] seasonal-state evidence is invalid: %+v", i, df.SeasonalState)
		}
		if df.FortuneLayers.LiuNian.RuleID == "" || df.FortuneLayers.LiuNian.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("DailyFortunes[%d] fortune-layer evidence is invalid: %+v", i, df.FortuneLayers)
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
