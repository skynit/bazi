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
	"bazi/internal/service/bazi"

	"github.com/gin-gonic/gin"
)

type mockChartSaver struct {
	chart *model.BirthChart
}

func (m *mockChartSaver) Create(chart *model.BirthChart) error {
	chart.ID = 1
	m.chart = chart
	return nil
}

func setupChartRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	h := &ChartHandler{
		Parser: &bazi.InputParser{},
		Bazi:   &bazi.BaziService{},
	}
	r.POST("/api/chart/preview", middleware.AuthMiddleware(), h.Preview)
	r.POST("/api/chart", middleware.AuthMiddleware(), h.Chart)
	return r
}

func setupChartRouterWithStore(store ChartSaver) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	h := &ChartHandler{
		Parser: &bazi.InputParser{},
		Bazi:   &bazi.BaziService{},
		Store:  store,
	}
	r.POST("/api/chart/preview", middleware.AuthMiddleware(), h.Preview)
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

func TestChartCreatePersistsDaYun(t *testing.T) {
	store := &mockChartSaver{}
	router := setupChartRouterWithStore(store)
	token, _ := middleware.GenerateToken(1, "testuser")

	body := chartJSONBody(t, model.ChartRequest{
		BirthYear:    2003,
		BirthMonth:   4,
		BirthDay:     15,
		BirthHour:    14,
		BirthMin:     0,
		CalendarType: "SOLAR",
		Gender:       "MALE",
		Name:         "Dayun",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if store.chart == nil {
		t.Fatal("chart was not saved")
	}
	if len(store.chart.DaYunStart) == 0 {
		t.Fatal("expected DaYunStart to be persisted")
	}

	var dayun struct {
		Calculated     bool                 `json:"calculated"`
		StartAge       int                  `json:"start_age"`
		StartAgeDetail bazi.DaYunStartAge   `json:"start_age_detail"`
		StartAt        string               `json:"start_at"`
		Direction      string               `json:"direction"`
		DirectionBasis string               `json:"direction_basis"`
		PreviousJie    *bazi.DaYunSolarTerm `json:"previous_jie"`
		NextJie        *bazi.DaYunSolarTerm `json:"next_jie"`
		ReferenceJie   *bazi.DaYunSolarTerm `json:"reference_jie"`
		Pillars        []model.Pillar       `json:"pillars"`
	}
	if err := json.Unmarshal(store.chart.DaYunStart, &dayun); err != nil {
		t.Fatalf("unmarshal DaYunStart: %v", err)
	}
	if dayun.Direction == "" {
		t.Error("expected dayun direction")
	}
	if !dayun.Calculated || dayun.StartAt == "" {
		t.Errorf("expected date-level dayun start, got %+v", dayun)
	}
	if dayun.StartAge != dayun.StartAgeDetail.Years {
		t.Errorf("start_age = %d, detail years = %d", dayun.StartAge, dayun.StartAgeDetail.Years)
	}
	if dayun.DirectionBasis == "" || dayun.PreviousJie == nil || dayun.NextJie == nil || dayun.ReferenceJie == nil {
		t.Errorf("expected dayun calculation evidence, got %+v", dayun)
	}
	if len(dayun.Pillars) == 0 {
		t.Error("expected dayun pillars")
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
