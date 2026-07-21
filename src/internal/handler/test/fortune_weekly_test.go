package handler_test

import (
	. "bazi/internal/handler"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service/fortune"

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
func (m *mockWeeklyChartStore) FindByIDForUser(id uint, userID uint) (*model.BirthChart, error) {
	if m.chart != nil && m.chart.ID == id && (m.chart.UserID == 0 || m.chart.UserID == userID) {
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
		Engine: fortune.NewFortuneEngine(),
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
	responseBody := w.Body.String()
	for _, required := range []string{
		`"structural_relation_index"`, `"highest_index_day"`, `"lowest_index_day"`,
		`"average_index"`, `"index_standard_deviation"`, `"season_element"`, `"ten_god"`,
		`"seasonal_state"`, `"fortune_layers"`,
	} {
		if !strings.Contains(responseBody, required) {
			t.Fatalf("weekly response missing %s: %s", required, responseBody)
		}
	}
	for _, forbidden := range []string{
		`"weekly_score"`, `"best_day"`, `"worst_day"`, `"peak_days"`,
		`"low_days"`, `"good_streak"`, `"bad_streak"`, `"key_advice"`,
		`"flow_impact"`, `"season_element_advice"`, `"today_ten_god"`,
		`"ten_god_favorable"`, `"ten_god_desc"`,
		`"dayun_influence"`, `"liunian_influence"`, `"advance_retreat"`,
		`"yongshen_impact"`, `"pattern_favorable"`, `"pattern_unfavorable"`,
		`"tiao_hou"`,
	} {
		if strings.Contains(responseBody, forbidden) {
			t.Fatalf("weekly response leaked legacy outcome field %s", forbidden)
		}
	}

	if len(resp.DailyFortunes) != 7 {
		t.Errorf("expected 7 daily fortunes, got %d", len(resp.DailyFortunes))
	}
	for i, day := range resp.DailyFortunes {
		if day.BaziEngineVersion == "" || day.BaziResolutionSource != "normalized_raw_birth" {
			t.Fatalf("day %d missing bazi trace metadata: engine=%q source=%q", i, day.BaziEngineVersion, day.BaziResolutionSource)
		}
		if day.ScoreBreakdown.PipelineVersion == "" || day.Score != day.ScoreBreakdown.FinalScore {
			t.Fatalf("day %d does not use unified score pipeline: score=%d breakdown=%+v", i, day.Score, day.ScoreBreakdown)
		}
		if day.SupportingEvidence == nil || day.CounterEvidence == nil {
			t.Fatalf("day %d evidence arrays must be present: %+v", i, day)
		}
		if day.SeasonalState.Status != "observed" || day.SeasonalState.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("day %d seasonal-state evidence is invalid: %+v", i, day.SeasonalState)
		}
		if day.FortuneLayers.LiuNian.RuleID == "" || day.FortuneLayers.LiuNian.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("day %d fortune-layer evidence is invalid: %+v", i, day.FortuneLayers)
		}
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
