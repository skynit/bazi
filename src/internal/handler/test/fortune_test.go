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
	"bazi/internal/service/fortune"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// mockChartStore implements ChartStore for testing.
type mockChartStore struct {
	chart *model.BirthChart
}

func (m *mockChartStore) FindByID(id uint) (*model.BirthChart, error) {
	if m.chart != nil && m.chart.ID == id {
		return m.chart, nil
	}
	return nil, nil
}
func (m *mockChartStore) FindByIDForUser(id uint, userID uint) (*model.BirthChart, error) {
	if m.chart != nil && m.chart.ID == id && (m.chart.UserID == 0 || m.chart.UserID == userID) {
		return m.chart, nil
	}
	return nil, nil
}
func (m *mockChartStore) Update(chart *model.BirthChart) error {
	m.chart = chart
	return nil
}

func setupFortuneRouter(store ChartStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	h := &FortuneHandler{
		Engine:     fortune.NewFortuneEngine(),
		ChartStore: store,
	}

	r.POST("/api/fortune", middleware.AuthMiddleware(), h.CalculateDaily)
	return r
}

func fortuneJSONBody(t *testing.T, v interface{}) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	return strings.NewReader(string(b))
}

func TestCalculateDailyValid(t *testing.T) {
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
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.FortuneResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.SolarDate == "" {
		t.Fatal("expected non-empty solar_date in response")
	}
	if resp.DayGanZhi == "" {
		t.Fatal("expected non-empty day_gan_zhi in response")
	}
	if resp.EngineVersion == "" || resp.RuleVersion == "" {
		t.Fatalf("expected engine and rule versions, got engine=%q rule=%q", resp.EngineVersion, resp.RuleVersion)
	}
	if resp.BaziEngineVersion != bazi.EngineVersion || resp.BaziResolutionSource != "normalized_raw_birth" {
		t.Fatalf("unexpected bazi trace metadata: engine=%q source=%q", resp.BaziEngineVersion, resp.BaziResolutionSource)
	}
	if resp.ScoreBreakdown.PipelineVersion == "" || resp.Score != resp.ScoreBreakdown.FinalScore {
		t.Fatalf("score contract mismatch: score=%d breakdown=%+v", resp.Score, resp.ScoreBreakdown)
	}
	if resp.ScoreBreakdown.ScoreKind != "structural_relation_index" || resp.ScoreBreakdown.ValidationStatus != "not_validated" || resp.ScoreBreakdown.InterpretationStatus != "not_adjudicated" || resp.ScoreBreakdown.IsOutcomeProbability {
		t.Fatalf("score governance boundary missing: %+v", resp.ScoreBreakdown)
	}
	if resp.EvidenceCompleteness != resp.ScoreBreakdown.EvidenceCompleteness {
		t.Fatalf("evidence completeness mismatch: response=%d breakdown=%d", resp.EvidenceCompleteness, resp.ScoreBreakdown.EvidenceCompleteness)
	}

	var contract map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &contract); err != nil {
		t.Fatalf("failed to parse raw contract: %v", err)
	}
	for _, field := range []string{"score_breakdown", "evidence_completeness", "supporting_evidence", "counter_evidence", "engine_version", "bazi_engine_version", "bazi_resolution_source", "rule_version", "season_element", "ten_god", "twelve_stage", "jian_chu", "huang_dao", "seasonal_state", "fortune_layers"} {
		if _, ok := contract[field]; !ok {
			t.Fatalf("fortune contract missing field %q: %s", field, w.Body.String())
		}
	}
	if resp.TwelveStage.Status != "observed" || resp.TwelveStage.InterpretationStatus != "not_adjudicated" || resp.TwelveStage.Name == "" {
		t.Fatalf("invalid twelve-stage evidence: %+v", resp.TwelveStage)
	}
	if resp.JianChu.Status != "observed" || resp.JianChu.InterpretationStatus != "not_adjudicated" || resp.JianChu.Name == "" {
		t.Fatalf("invalid JianChu evidence: %+v", resp.JianChu)
	}
	if resp.HuangDao.Status != "observed" || resp.HuangDao.InterpretationStatus != "not_adjudicated" || resp.HuangDao.Name == "" {
		t.Fatalf("invalid HuangDao evidence: %+v", resp.HuangDao)
	}
	if resp.SeasonElement.Status != "observed" || resp.SeasonElement.InterpretationStatus != "not_adjudicated" || resp.SeasonElement.Season == "" {
		t.Fatalf("invalid season-element evidence: %+v", resp.SeasonElement)
	}
	if resp.TenGod.Status != "observed" || resp.TenGod.InterpretationStatus != "not_adjudicated" || resp.TenGod.Name == "" {
		t.Fatalf("invalid ten-god evidence: %+v", resp.TenGod)
	}
	if resp.SeasonalState.Status != "observed" || resp.SeasonalState.InterpretationStatus != "not_adjudicated" || resp.SeasonalState.State == "" {
		t.Fatalf("invalid seasonal-state evidence: %+v", resp.SeasonalState)
	}
	for _, layer := range []model.FortuneLayer{resp.FortuneLayers.DaYun, resp.FortuneLayers.LiuNian, resp.FortuneLayers.LiuYue, resp.FortuneLayers.XiaoYun} {
		if layer.RuleID == "" || layer.Status != "observed" || layer.Basis == "" || layer.InterpretationStatus != "not_adjudicated" {
			t.Fatalf("invalid fortune-layer evidence: %+v", layer)
		}
	}
	if strings.Contains(w.Body.String(), `"confidence"`) {
		t.Fatalf("fortune contract must use evidence completeness instead of confidence: %s", w.Body.String())
	}
	for _, field := range []string{"stage_favorable", "stage_desc", "stage_flexible", "overall_verdict", "favor_score"} {
		if _, ok := contract[field]; ok {
			t.Fatalf("fortune contract must not expose legacy field %q: %s", field, w.Body.String())
		}
	}
	for _, field := range []string{
		"analysis", "lucky_color", "lucky_number", "wealth_direction", "auspicious_hours",
		"guide", "blessing_assets", "yi", "ji", "yi_ji", "flow_impact",
		"season_element_advice", "today_ten_god", "ten_god_favorable", "ten_god_desc",
		"dayun_influence", "liunian_influence", "advance_retreat", "yongshen_impact",
		"pattern_name", "pattern_type", "pattern_favorable", "pattern_unfavorable",
		"tiao_hou",
	} {
		if _, ok := contract[field]; ok {
			t.Fatalf("fortune contract must not expose unadjudicated event field %q: %s", field, w.Body.String())
		}
	}
	layerJSON, err := json.Marshal(resp.FortuneLayers)
	if err != nil {
		t.Fatalf("failed to marshal fortune layers: %v", err)
	}
	for _, forbidden := range []string{`"favorable"`, `"is_favorable"`, `"score"`, `"description"`, `"evidence"`, `"element_change"`, `"activated_shen_sha"`} {
		if strings.Contains(string(layerJSON), forbidden) {
			t.Fatalf("fortune layers leaked judgment field %s: %s", forbidden, layerJSON)
		}
	}
	var breakdown map[string]json.RawMessage
	if err := json.Unmarshal(contract["score_breakdown"], &breakdown); err != nil {
		t.Fatalf("failed to parse score breakdown: %v", err)
	}
	if _, ok := breakdown["detail_score"]; ok {
		t.Fatalf("fortune score breakdown must not expose legacy nine-dimension score: %s", w.Body.String())
	}
}

func TestCalculateDailyUsesVerifiedStoredSnapshot(t *testing.T) {
	normalized, err := bazi.NormalizeBirthInput(bazi.BirthInput{
		Year: 1990, Month: 6, Day: 15, Hour: 8, Minute: 30,
		CalendarType: model.CalendarSolar, Gender: model.GenderMale, Timezone: bazi.DefaultBirthTimezone,
	})
	if err != nil {
		t.Fatalf("failed to normalize snapshot fixture: %v", err)
	}
	snapshot, err := (&bazi.BaziService{}).CalculateNormalizedBirth(normalized)
	if err != nil {
		t.Fatalf("failed to calculate snapshot fixture: %v", err)
	}
	snapshot.RuleVersion = "stored-rule-v1"
	snapshot.School = "stored-school-v1"
	snapshot.RuleMeta.RuleVersion = snapshot.RuleVersion
	snapshot.RuleMeta.School = snapshot.School
	snapshot.BodyStrength.RuleVersion = snapshot.RuleVersion
	snapshot.BodyStrength.School = snapshot.School

	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("failed to marshal snapshot: %v", err)
	}
	normalizedJSON, err := json.Marshal(normalized)
	if err != nil {
		t.Fatalf("failed to marshal normalized birth: %v", err)
	}

	chart := &model.BirthChart{
		BirthYear:       2000,
		BirthMonth:      99,
		BirthDay:        99,
		BirthHour:       99,
		Gender:          model.GenderFemale,
		EngineVersion:   "stored-engine-v1",
		RuleVersion:     snapshot.RuleVersion,
		NormalizedBirth: datatypes.JSON(normalizedJSON),
		BaziSnapshot:    datatypes.JSON(snapshotJSON),
	}
	chart.ID = 1
	router := setupFortuneRouter(&mockChartStore{chart: chart})

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/fortune", fortuneJSONBody(t, model.FortuneRequest{
		ChartID:   1,
		QueryDate: "2025-01-15",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp model.FortuneResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.BaziResolutionSource != "bazi_snapshot" || resp.BaziEngineVersion != chart.EngineVersion {
		t.Fatalf("unexpected snapshot trace metadata: engine=%q source=%q", resp.BaziEngineVersion, resp.BaziResolutionSource)
	}
	if resp.RuleVersion != snapshot.RuleVersion || resp.School != snapshot.School {
		t.Fatalf("stored snapshot rules were not preserved: rule=%q school=%q", resp.RuleVersion, resp.School)
	}
}

func TestCalculateDailyNoJWT(t *testing.T) {
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

	body := fortuneJSONBody(t, model.FortuneRequest{
		ChartID:   1,
		QueryDate: "2025-01-15",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/fortune", body)
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
