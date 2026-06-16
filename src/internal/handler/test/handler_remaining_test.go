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
	"bazi/internal/service/ziwei"

	"github.com/gin-gonic/gin"
)

// =============================================================================
// /api/charts GET pagination supplementary tests
// =============================================================================

func TestListCharts_PaginationEdgeCases(t *testing.T) {
	charts := newMockChartListStore()
	// Add 5 charts for user 1
	for i := 0; i < 5; i++ {
		charts.AddChart(&model.BirthChart{
			UserID: 1, Name: "Chart", BirthYear: 1990 + i,
			BirthMonth: 1, BirthDay: 1, BirthHour: 0, Gender: "男",
		})
	}

	fortune := &mockFortuneHistoryStore{}
	router := setupHistoryRouter(charts, fortune)
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	tests := []struct {
		name     string
		url      string
		wantPage int
		wantSize int
		wantLen  int
		wantCode int
	}{
		{"default pagination", "/api/charts", 1, 10, 5, http.StatusOK},
		{"page 2 with size 2", "/api/charts?page=2&page_size=2", 2, 2, 2, http.StatusOK},
		{"page 3 with size 2 (last page 1 item)", "/api/charts?page=3&page_size=2", 3, 2, 1, http.StatusOK},
		{"page beyond range", "/api/charts?page=10&page_size=10", 10, 10, 0, http.StatusOK},
		{"page_size capped at 100", "/api/charts?page=1&page_size=200", 1, 10, 5, http.StatusOK},
		{"page_size 0 uses default", "/api/charts?page=1&page_size=0", 1, 10, 5, http.StatusOK},
		{"page 0 uses default", "/api/charts?page=0&page_size=10", 1, 10, 5, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.url, nil)
			req.Header.Set("Authorization", "Bearer "+token)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantCode {
				t.Errorf("expected status %d, got %d", tt.wantCode, w.Code)
				return
			}

			var resp model.ChartListResponse
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to parse: %v", err)
			}

			if resp.Page != tt.wantPage {
				t.Errorf("page = %d, want %d", resp.Page, tt.wantPage)
			}
			if resp.PageSize != tt.wantSize {
				t.Errorf("page_size = %d, want %d", resp.PageSize, tt.wantSize)
			}
			if len(resp.Charts) != tt.wantLen {
				t.Errorf("charts count = %d, want %d", len(resp.Charts), tt.wantLen)
			}
			if resp.Total != 5 {
				t.Errorf("total = %d, want 5", resp.Total)
			}
		})
	}
}

// =============================================================================
// /api/charts/:id  chart retrieval supplementary tests
// =============================================================================

func TestGetChartByID_NotFound(t *testing.T) {
	charts := newMockChartListStore()
	charts.AddChart(&model.BirthChart{
		UserID: 1, Name: "Existing", BirthYear: 2000,
		BirthMonth: 1, BirthDay: 1, BirthHour: 0, Gender: "男",
	})

	fortune := &mockFortuneHistoryStore{}
	router := setupHistoryRouter(charts, fortune)
	token, _ := middleware.GenerateToken(1, "testuser")

	// Non-existent ID
	req := httptest.NewRequest(http.MethodGet, "/api/charts/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent chart, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetChartByID_InvalidID(t *testing.T) {
	charts := newMockChartListStore()
	fortune := &mockFortuneHistoryStore{}
	router := setupHistoryRouter(charts, fortune)
	token, _ := middleware.GenerateToken(1, "testuser")

	// Invalid ID (not a number)
	req := httptest.NewRequest(http.MethodGet, "/api/charts/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid ID, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetCharts_Unauthorized(t *testing.T) {
	charts := newMockChartListStore()
	fortune := &mockFortuneHistoryStore{}
	router := setupHistoryRouter(charts, fortune)

	req := httptest.NewRequest(http.MethodGet, "/api/charts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestGetChartByID_Unauthorized(t *testing.T) {
	charts := newMockChartListStore()
	fortune := &mockFortuneHistoryStore{}
	router := setupHistoryRouter(charts, fortune)

	req := httptest.NewRequest(http.MethodGet, "/api/charts/1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

// =============================================================================
// /api/ziwei/period — supplementary tests
// =============================================================================

func TestZiWeiPeriod_UnknownType(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear: 1984, BirthMonth: 2, BirthDay: 15,
		BirthHour: 8, BirthMin: 0, Gender: "男",
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
		PeriodType: "unknown_type_xyz",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown period_type, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if errMsg, ok := resp["error"]; !ok || errMsg == "" {
		t.Error("expected error message for unknown period_type")
	}
}

func TestZiWeiPeriod_InvalidJSON(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear: 1984, BirthMonth: 2, BirthDay: 15,
		BirthHour: 8, BirthMin: 0, Gender: "男",
	}
	chart.ID = 1

	store := &mockWeeklyChartStore{chart: chart}
	router := setupZiWeiPeriodRouter(store)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	// Malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d: %s", w.Code, w.Body.String())
	}
}

// =============================================================================
// /api/ziwei/overlay tests
// =============================================================================

func setupZiWeiOverlayRouter(store ChartStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	svc := ziwei.NewZiWeiService()
	RegisterZiWeiPeriodRoutes(api, svc, store)
	return r
}

func TestZiWeiOverlay_Success(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear: 1984, BirthMonth: 2, BirthDay: 15,
		BirthHour: 8, BirthMin: 0, Gender: "男",
	}
	chart.ID = 1

	store := &mockWeeklyChartStore{chart: chart}
	router := setupZiWeiOverlayRouter(store)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"chart_id": 1,
		"year":     2025,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/overlay", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if year, ok := resp["year"]; !ok || year.(float64) != 2025 {
		t.Errorf("expected year=2025, got %v", resp["year"])
	}

	// Should have liu_nian_stars
	if _, ok := resp["liu_nian_stars"]; !ok {
		t.Error("response missing 'liu_nian_stars'")
	}
}

func TestZiWeiOverlay_ChartNotFound(t *testing.T) {
	store := &mockWeeklyChartStore{} // no chart set up
	router := setupZiWeiOverlayRouter(store)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	body, _ := json.Marshal(map[string]interface{}{
		"chart_id": 999,
		"year":     2025,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/overlay", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for chart not found, got %d: %s", w.Code, w.Body.String())
	}
}

func TestZiWeiOverlay_Unauthorized(t *testing.T) {
	store := &mockWeeklyChartStore{}
	router := setupZiWeiOverlayRouter(store)

	body, _ := json.Marshal(map[string]interface{}{
		"chart_id": 1,
		"year":     2025,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/overlay", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestZiWeiOverlay_InvalidJSON(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear: 1984, BirthMonth: 2, BirthDay: 15,
		BirthHour: 8, BirthMin: 0, Gender: "男",
	}
	chart.ID = 1

	store := &mockWeeklyChartStore{chart: chart}
	router := setupZiWeiOverlayRouter(store)

	token, _ := middleware.GenerateToken(1, "testuser")

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/overlay", strings.NewReader("bad-json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid JSON, got %d: %s", w.Code, w.Body.String())
	}
}

func TestZiWeiOverlay_DefaultYear(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear: 1984, BirthMonth: 2, BirthDay: 15,
		BirthHour: 8, BirthMin: 0, Gender: "男",
	}
	chart.ID = 1

	store := &mockWeeklyChartStore{chart: chart}
	router := setupZiWeiOverlayRouter(store)

	token, _ := middleware.GenerateToken(1, "testuser")

	// No year provided — should use default (current year)
	body, _ := json.Marshal(map[string]interface{}{
		"chart_id": 1,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/overlay", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with default year, got %d: %s", w.Code, w.Body.String())
	}
}
