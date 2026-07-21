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

type periodRequest struct {
	ChartID    uint   `json:"chart_id"`
	PeriodType string `json:"period_type"`
	Year       int    `json:"year"`
	Month      int    `json:"month,omitempty"`
	Day        int    `json:"day,omitempty"`
}

func setupZiWeiPeriodRouter(store ChartStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	svc := ziwei.NewZiWeiService()
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

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	periods, ok := resp["periods"]
	if !ok {
		t.Fatal("response missing 'periods' key")
	}
	periodSlice, ok := periods.([]interface{})
	if !ok || len(periodSlice) == 0 {
		t.Fatal("expected non-empty dayun stages")
	}
}

func TestZiWeiPeriodDayunUsesLunarYearNominalAge(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear: 2000, BirthMonth: 8, BirthDay: 16, BirthHour: 3, Gender: "女",
	}
	chart.ID = 1
	router := setupZiWeiPeriodRouter(&mockWeeklyChartStore{chart: chart})
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(periodRequest{
		ChartID: 1, PeriodType: "dayun", Year: 2022, Month: 8, Day: 19,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		TargetDate     string `json:"target_date"`
		NominalAge     int    `json:"nominal_age"`
		AgeBasis       string `json:"age_basis"`
		BoundaryPolicy string `json:"boundary_policy"`
		Analysis       struct {
			Stages []struct {
				StartAge int  `json:"start_age"`
				EndAge   int  `json:"end_age"`
				Current  bool `json:"current"`
			} `json:"dayun_stages"`
		} `json:"analysis"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.TargetDate != "2022-08-19" || payload.NominalAge != 23 ||
		payload.AgeBasis != "target_lunar_year_minus_birth_lunar_year_plus_one" ||
		payload.BoundaryPolicy != ziwei.ZiWeiHoroscopeBoundaryNormal {
		t.Fatalf("unexpected dayun age context: %+v", payload)
	}
	for _, stage := range payload.Analysis.Stages {
		if !stage.Current {
			continue
		}
		if stage.StartAge != 23 || stage.EndAge != 32 {
			t.Fatalf("current dayun = %d-%d, want 23-32", stage.StartAge, stage.EndAge)
		}
		return
	}
	t.Fatal("dayun analysis has no current stage")
}

func TestZiWeiPeriodCachesComputedChart(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  2003,
		BirthMonth: 4,
		BirthDay:   15,
		BirthHour:  14,
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
	if !store.chart.ZiWeiComputed {
		t.Fatal("expected ziwei chart to be marked computed")
	}
	if len(store.chart.ZiWeiResult) == 0 {
		t.Fatal("expected ziwei result to be cached")
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

func TestZiWeiPeriodLiunianReturnsDerivedIntegrityContract(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
		Gender:     "男",
	}
	chart.ID = 1
	router := setupZiWeiPeriodRouter(&mockWeeklyChartStore{chart: chart})
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(periodRequest{ChartID: 1, PeriodType: "liunian", Year: 2025})
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Periods []map[string]json.RawMessage `json:"periods"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Periods) != 1 {
		t.Fatalf("periods length = %d, want 1", len(payload.Periods))
	}
	period := payload.Periods[0]
	if _, exists := period["content_hash"]; exists {
		t.Fatalf("derived response leaked natal content_hash: %s", period["content_hash"])
	}
	for _, field := range []string{"derivation_type", "derivation_input", "derivation_fingerprint", "base_content_hash", "derived_content_hash"} {
		if len(period[field]) == 0 {
			t.Fatalf("derived response missing %s: %s", field, w.Body.String())
		}
	}
	var derivationType string
	if err := json.Unmarshal(period["derivation_type"], &derivationType); err != nil || derivationType != "liunian" {
		t.Fatalf("derivation_type = %q, err=%v", derivationType, err)
	}
	var input ziwei.ZiWeiDerivationInput
	if err := json.Unmarshal(period["derivation_input"], &input); err != nil {
		t.Fatal(err)
	}
	if input.CalendarType != "LUNAR_YEAR" || input.Year != 2025 || input.Month != 0 || input.Day != 0 ||
		input.Basis != "target_lunar_year_label" || input.PeriodGanZhi != "乙巳" {
		t.Fatalf("unexpected derivation input: %+v", input)
	}
}

func TestZiWeiPeriodRejectsInvalidSolarDate(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
		Gender:     "男",
	}
	chart.ID = 1
	router := setupZiWeiPeriodRouter(&mockWeeklyChartStore{chart: chart})
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(periodRequest{
		ChartID: 1, PeriodType: "liuri", Year: 2026, Month: 2, Day: 29,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestZiWeiPeriodLiuyueUsesExactSolarDateContext(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear: 2003, BirthMonth: 4, BirthDay: 15, BirthHour: 14, Gender: "男",
	}
	chart.ID = 1
	router := setupZiWeiPeriodRouter(&mockWeeklyChartStore{chart: chart})
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(periodRequest{
		ChartID: 1, PeriodType: "liuyue", Year: 2020, Month: 6, Day: 7,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Periods []struct {
			Input   ziwei.ZiWeiDerivationInput `json:"derivation_input"`
			Stars   [][]string                 `json:"liu_yue_stars"`
			FourHua [][]string                 `json:"liu_yue_four_hua"`
			Palaces []string                   `json:"liu_yue_palaces"`
		} `json:"periods"`
		Analysis struct {
			GanZhi       string `json:"gan_zhi"`
			FocusPalaces []struct {
				PeriodPalace string `json:"period_palace"`
			} `json:"focus_palaces"`
		} `json:"analysis"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if len(payload.Periods) != 1 {
		t.Fatalf("periods length = %d, want 1", len(payload.Periods))
	}
	input := payload.Periods[0].Input
	if input.Year != 2020 || input.Month != 6 || input.Day != 7 ||
		input.ResolvedLunarDate != (ziwei.ZiWeiResolvedLunarDate{Year: 2020, Month: 4, Day: 16, IsLeapMonth: true}) ||
		input.PeriodGanZhi != "壬午" || payload.Analysis.GanZhi != input.PeriodGanZhi {
		t.Fatalf("liuyue context diverged: input=%+v analysis=%+v", input, payload.Analysis)
	}
	starCount, fourHuaCount := 0, 0
	for _, group := range payload.Periods[0].Stars {
		starCount += len(group)
		for _, star := range group {
			if strings.Contains(star, "化") {
				t.Fatalf("four-hua label leaked into liu_yue_stars: %q", star)
			}
		}
	}
	for _, group := range payload.Periods[0].FourHua {
		fourHuaCount += len(group)
	}
	if starCount != 10 || fourHuaCount != 4 {
		t.Fatalf("liuyue layer counts = stars:%d four_hua:%d, want 10/4", starCount, fourHuaCount)
	}
	palaceCount := map[string]int{}
	for _, palace := range payload.Periods[0].Palaces {
		palaceCount[palace]++
	}
	if len(payload.Periods[0].Palaces) != 12 || len(palaceCount) != 12 || palaceCount["命宫"] != 1 {
		t.Fatalf("liuyue palace layer is incomplete: %v", payload.Periods[0].Palaces)
	}
	for _, focus := range payload.Analysis.FocusPalaces {
		if focus.PeriodPalace == "" {
			t.Fatal("focus palace omitted the liuyue palace name")
		}
	}
}

func TestZiWeiPeriodSummaryUsesLunarYearBoundary(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear: 2000, BirthMonth: 8, BirthDay: 16, BirthHour: 3, Gender: "男",
	}
	chart.ID = 1
	router := setupZiWeiPeriodRouter(&mockWeeklyChartStore{chart: chart})
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(periodRequest{
		ChartID: 1, PeriodType: "period_summary", Year: 2024, Month: 2, Day: 9,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var payload struct {
		Summary struct {
			Liunian struct {
				GanZhi string `json:"gan_zhi"`
			} `json:"liunian"`
			Liuyue struct {
				GanZhi string `json:"gan_zhi"`
			} `json:"liuyue"`
			Liuri struct {
				GanZhi string `json:"gan_zhi"`
			} `json:"liuri"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload.Summary.Liunian.GanZhi != "癸卯" {
		t.Fatalf("summary liunian gan-zhi = %q, want lunar year 癸卯", payload.Summary.Liunian.GanZhi)
	}
	if payload.Summary.Liuyue.GanZhi == "" || payload.Summary.Liuri.GanZhi == "" {
		t.Fatalf("summary period layers are incomplete: %+v", payload.Summary)
	}
}

func TestZiWeiPeriodRejectsPartialSolarDate(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear: 2003, BirthMonth: 4, BirthDay: 15, BirthHour: 14, Gender: "男",
	}
	chart.ID = 1
	router := setupZiWeiPeriodRouter(&mockWeeklyChartStore{chart: chart})
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := json.Marshal(periodRequest{
		ChartID: 1, PeriodType: "liuyue", Year: 2026, Month: 7,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/period", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestZiWeiPeriodAnalysisPresent(t *testing.T) {
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
		PeriodType: "liuyue",
		Year:       2023,
		Month:      3,
		Day:        15,
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
		Analysis struct {
			RuleVersion         string   `json:"rule_version"`
			School              string   `json:"school"`
			EvidenceBasis       string   `json:"evidence_basis"`
			PlacementBasis      string   `json:"placement_basis"`
			InterpretationBasis string   `json:"interpretation_basis"`
			ValidationStatus    string   `json:"validation_status"`
			IsOutcomeConclusion bool     `json:"is_outcome_conclusion"`
			ReviewNotes         []string `json:"review_notes"`
			Limitations         []string `json:"limitations"`
			CrossLayerRelations []struct {
				SourceLayer string `json:"source_layer"`
				TargetLayer string `json:"target_layer"`
				Relation    string `json:"relation"`
				RuleID      string `json:"rule_id"`
			} `json:"cross_layer_relations"`
		} `json:"analysis"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Analysis.RuleVersion == "" || resp.Analysis.School == "" {
		t.Fatal("expected analysis rule metadata")
	}
	if resp.Analysis.EvidenceBasis != "mixed_deterministic_projection_and_unadjudicated_traditional_labels" ||
		resp.Analysis.PlacementBasis != "deterministic_rule_projection" ||
		resp.Analysis.InterpretationBasis != "traditional_rule_labels" ||
		resp.Analysis.ValidationStatus != "not_adjudicated" || resp.Analysis.IsOutcomeConclusion {
		t.Fatalf("analysis validation boundary missing: %+v", resp.Analysis)
	}
	if len(resp.Analysis.CrossLayerRelations) == 0 ||
		resp.Analysis.CrossLayerRelations[0].SourceLayer != "liuyue" ||
		resp.Analysis.CrossLayerRelations[0].TargetLayer != "liunian" ||
		resp.Analysis.CrossLayerRelations[0].Relation != "伏吟" ||
		resp.Analysis.CrossLayerRelations[0].RuleID == "" {
		t.Fatalf("expected reproducible liuyue-to-liunian structure: %+v", resp.Analysis.CrossLayerRelations)
	}
	if len(resp.Analysis.ReviewNotes) == 0 || len(resp.Analysis.Limitations) == 0 {
		t.Fatalf("expected review notes and limitations: %+v", resp.Analysis)
	}
	if strings.Contains(w.Body.String(), `"score"`) || strings.Contains(w.Body.String(), `"recommendations"`) || strings.Contains(w.Body.String(), `"risks"`) {
		t.Fatalf("legacy prediction fields leaked: %s", w.Body.String())
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
