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

type periodRequest struct {
	ChartID    uint   `json:"chart_id"`
	PeriodType string `json:"period_type"`
	Year       int    `json:"year"`
}

func setupZiWeiPeriodRouter(store ChartStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	svc := service.NewZiWeiService()
	RegisterZiWeiPeriodRoutes(api, svc, store)
	return r
}

func TestZiWeiPeriodDayun(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	}
	chart.ID = 1

	store := &mockWeeklyChartStore{chart: chart}
	router := setupZiWeiPeriodRouter(store)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	body, _ := json.Marshal(periodRequest{
		ChartID:    1,
		PeriodType: "dayun",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Periods service.Dayun `json:"periods"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	dayun := resp.Periods

	if len(dayun) == 0 {
		t.Fatal("expected non-empty dayun stages")
	}

	for i, stage := range dayun {
		if stage.Palace == "" {
			t.Errorf("dayun stage %d has empty palace", i)
		}
	}
}

func TestZiWeiPeriodLiunian(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	}
	chart.ID = 1

	store := &mockWeeklyChartStore{chart: chart}
	router := setupZiWeiPeriodRouter(store)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	body, _ := json.Marshal(periodRequest{
		ChartID:    1,
		PeriodType: "liunian",
		Year:       2025,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Periods []interface{} `json:"periods"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(resp.Periods) == 0 {
		t.Error("expected non-empty periods for liunian")
	}
}

func TestZiWeiPeriodNoJWT(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	}
	chart.ID = 1

	store := &mockWeeklyChartStore{chart: chart}
	router := setupZiWeiPeriodRouter(store)

	body, _ := json.Marshal(periodRequest{
		ChartID:    1,
		PeriodType: "dayun",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
