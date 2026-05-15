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

func setupZiWeiPeriodRouter(store ChartStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	svc := service.NewZiWeiService()
	RegisterZiWeiPeriodRoutes(r, svc, store)
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

	var dayun service.Dayun
	if err := json.Unmarshal(w.Body.Bytes(), &dayun); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

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

	var result service.ZiWeiChart
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if len(result.Palaces) != 12 {
		t.Errorf("expected 12 palaces, got %d", len(result.Palaces))
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
