package handler_test

import (
	handler "bazi/internal/handler"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"
)

// ── Chart Handler 额外测试 ──────────────────────────────────

func TestChartLunarCalendar(t *testing.T) {
	router := setupChartRouter()
	token, _ := middleware.GenerateToken(1, "testuser")

	body := chartJSONBody(t, model.ChartRequest{
		BirthYear:    2024,
		BirthMonth:   1,
		BirthDay:     1,
		BirthHour:    8,
		BirthMin:     30,
		CalendarType: "LUNAR",
		Gender:       "MALE",
		Name:         "LunarTest",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for a valid lunar date, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		CalendarType string `json:"calendar_type"`
		Validation   struct {
			Converted string `json:"converted_solar_date_time"`
		} `json:"birth_validation"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.CalendarType != "LUNAR" {
		t.Fatalf("calendar_type = %q, want LUNAR", resp.CalendarType)
	}
	if resp.Validation.Converted != "2024-02-10 08:30" {
		t.Fatalf("converted solar time = %q, want 2024-02-10 08:30", resp.Validation.Converted)
	}
}

func TestChartPreviewDoesNotPersist(t *testing.T) {
	store := &mockChartSaver{}
	router := setupChartRouterWithStore(store)
	token, _ := middleware.GenerateToken(1, "testuser")
	body := chartJSONBody(t, model.ChartRequest{
		BirthYear: 1990, BirthMonth: 5, BirthDay: 15, BirthHour: 8, BirthMin: 12,
		CalendarType: "", Gender: "MALE",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/chart/preview", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if store.chart != nil {
		t.Fatal("preview must not persist a chart")
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if resp["id"].(float64) != 0 {
		t.Fatalf("preview id = %v, want 0", resp["id"])
	}
	if resp["calendar_type"] != "SOLAR" {
		t.Fatalf("default calendar_type = %v, want SOLAR", resp["calendar_type"])
	}
	if resp["name"] != "1990-05-15 命盘" {
		t.Fatalf("default name = %v", resp["name"])
	}
	if _, ok := resp["birth_validation"]; !ok {
		t.Fatal("preview response missing birth_validation")
	}
	for _, key := range []string{"year_pillar", "month_pillar", "day_pillar", "hour_pillar"} {
		if _, ok := resp[key]; !ok {
			t.Fatalf("preview response missing %s", key)
		}
	}
}

func TestChartBaziCalendar(t *testing.T) {
	router := setupChartRouter()
	token, _ := middleware.GenerateToken(1, "testuser")

	body := chartJSONBody(t, model.ChartRequest{
		BirthYear:    1990,
		BirthMonth:   5,
		BirthDay:     15,
		BirthHour:    8,
		BirthMin:     0,
		CalendarType: "BAZI",
		Gender:       "MALE",
		Name:         "BaziTest",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// BAZI format requires 4 pillar pairs like "庚午 辛巳 甲子 壬申",
	// formatChartInput always produces solar format so this should return 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for BAZI with solar-formatted input, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChartEmptyName(t *testing.T) {
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
		Name:         "",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Name is optional, should still return 200
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for empty name, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChartFemale(t *testing.T) {
	router := setupChartRouter()
	token, _ := middleware.GenerateToken(1, "testuser")

	body := chartJSONBody(t, model.ChartRequest{
		BirthYear:    1995,
		BirthMonth:   8,
		BirthDay:     20,
		BirthHour:    14,
		BirthMin:     30,
		CalendarType: "SOLAR",
		Gender:       "FEMALE",
		Name:         "FemaleTest",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for FEMALE, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify five elements present
	if fe, ok := resp["five_elements"].(map[string]interface{}); ok {
		if len(fe) == 0 {
			t.Error("five_elements is empty")
		}
	} else {
		t.Log("five_elements not present (may be optional)")
	}
}

// ── Fortune Handler 额外测试 ──────────────────────────────────

func TestCalculateDaily_Components(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1990,
		BirthMonth: 6,
		BirthDay:   15,
		BirthHour:  8,
		Gender:     "男",
	}
	chart.ID = 1

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
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.FortuneResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	// Verify score is in valid range
	if resp.Score < 0 || resp.Score > 100 {
		t.Errorf("Score = %d, expected [0,100]", resp.Score)
	}

	// Verify lucky number is non-negative
	if resp.LuckyNumber < 0 {
		t.Errorf("LuckyNumber = %d, expected >= 0", resp.LuckyNumber)
	}

	// Verify color is non-empty
	if resp.LuckyColor == "" {
		t.Error("LuckyColor is empty")
	}

	// Verify wealth direction
	if resp.WealthDir == "" {
		t.Error("WealthDir is empty")
	}

	// Verify clash zodiac
	if resp.ClashZodiac == "" {
		t.Error("ClashZodiac is empty")
	}

	// Verify auspicious hours
	if len(resp.AuspiciousHours) == 0 {
		t.Error("AuspiciousHours is empty")
	}
}

func TestCalculateDaily_MissingQueryDate(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1990,
		BirthMonth: 6,
		BirthDay:   15,
		BirthHour:  8,
		Gender:     "男",
	}
	chart.ID = 1

	store := &mockChartStore{chart: chart}
	router := setupFortuneRouter(store)

	token, _ := middleware.GenerateToken(1, "testuser")

	body := fortuneJSONBody(t, map[string]interface{}{
		"chart_id": 1,
		// query_date missing
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fortune", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing query_date, got %d", w.Code)
	}
}

func TestCalculateDaily_ChartNotFound(t *testing.T) {
	// Empty store — no chart set up
	store := &mockChartStore{}
	router := setupFortuneRouter(store)

	token, _ := middleware.GenerateToken(1, "testuser")

	body := fortuneJSONBody(t, model.FortuneRequest{
		ChartID:   999,
		QueryDate: "2025-01-15",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fortune", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for chart not found, got %d: %s", w.Code, w.Body.String())
	}
}

// ── Weekly Fortune Handler 额外测试 ──────────────────────────

func TestWeeklyFortune_ScoreRange(t *testing.T) {
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
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.WeeklyFortuneResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.WeeklyScore < 0 || resp.WeeklyScore > 100 {
		t.Errorf("WeeklyScore = %d, expected [0,100]", resp.WeeklyScore)
	}

	if len(resp.DailyFortunes) != 7 {
		t.Errorf("expected 7 daily fortunes, got %d", len(resp.DailyFortunes))
	}
}

func TestWeeklyFortune_InvalidDate(t *testing.T) {
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

	token, _ := middleware.GenerateToken(1, "testuser")

	reqBody := model.WeeklyFortuneRequest{
		ChartID:   1,
		StartDate: "invalid-date",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/weekly", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid date, got %d", w.Code)
	}
}

func TestWeeklyFortune_ChartNotFound(t *testing.T) {
	store := &mockWeeklyChartStore{}
	router := setupWeeklyRouter(store)

	token, _ := middleware.GenerateToken(1, "testuser")

	reqBody := model.WeeklyFortuneRequest{
		ChartID:   999,
		StartDate: "2025-01-06",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/weekly", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for chart not found, got %d", w.Code)
	}
}

// ── Monthly Fortune Handler 额外测试 ─────────────────────────

func TestMonthlyFortune_ScoreRange(t *testing.T) {
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
		Month:   3,
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
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.MonthlyScore < 0 || resp.MonthlyScore > 100 {
		t.Errorf("MonthlyScore = %d, expected [0,100]", resp.MonthlyScore)
	}
}

func TestMonthlyFortune_InvalidYear(t *testing.T) {
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

	token, _ := middleware.GenerateToken(1, "testuser")

	reqBody := model.MonthlyFortuneRequest{
		ChartID: 1,
		Year:    1800, // before 1900
		Month:   1,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/monthly", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid year, got %d: %s", w.Code, w.Body.String())
	}
}

func TestMonthlyFortune_InvalidMonth(t *testing.T) {
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

	token, _ := middleware.GenerateToken(1, "testuser")

	// Month 0 is invalid
	reqBody := model.MonthlyFortuneRequest{
		ChartID: 1,
		Year:    2024,
		Month:   0,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/monthly", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid month, got %d: %s", w.Code, w.Body.String())
	}

	// Month 13 is also invalid
	reqBody.Month = 13
	body2, _ := json.Marshal(reqBody)
	req2 := httptest.NewRequest(http.MethodPost, "/api/fortune/monthly", strings.NewReader(string(body2)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid month 13, got %d: %s", w2.Code, w2.Body.String())
	}
}

func TestMonthlyFortune_ChartNotFound(t *testing.T) {
	store := &mockMonthlyChartStore{}
	router := setupMonthlyTestRouter(store)

	token, _ := middleware.GenerateToken(1, "testuser")

	reqBody := model.MonthlyFortuneRequest{
		ChartID: 999,
		Year:    2024,
		Month:   6,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/fortune/monthly", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for chart not found, got %d: %s", w.Code, w.Body.String())
	}
}

// ── 辅助验证 ──────────────────────────────────────────

func TestSetupChartRouterValidates(t *testing.T) {
	router := setupChartRouter()
	if router == nil {
		t.Fatal("setupChartRouter returned nil")
	}
}

func TestNormalizeGender(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"男", "MALE"},
		{"女", "FEMALE"},
		{"male", "MALE"},
		{"MALE", "MALE"},
		{"m", "MALE"},
		{"female", "FEMALE"},
		{"FEMALE", "FEMALE"},
		{"f", "FEMALE"},
		{"unknown", "MALE"}, // default
		{"", "MALE"},
	}

	for _, tt := range tests {
		got := handler.ExportNormalizeGender(tt.input)
		if got != tt.want {
			t.Errorf("ExportNormalizeGender(%q) = %s, want %s", tt.input, got, tt.want)
		}
	}
}
