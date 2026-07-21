package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bazi/internal/handler"
	"bazi/internal/middleware"
	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

type mockFeedbackChartStore struct {
	chart *model.BirthChart
}

func (m *mockFeedbackChartStore) FindByIDForUser(id uint, userID uint) (*model.BirthChart, error) {
	if m.chart != nil && m.chart.ID == id && m.chart.UserID == userID {
		return m.chart, nil
	}
	return nil, nil
}

type mockFeedbackStore struct {
	items []model.FortuneFeedback
}

func (m *mockFeedbackStore) Create(feedback *model.FortuneFeedback) error {
	feedback.ID = uint(len(m.items) + 1)
	m.items = append(m.items, *feedback)
	return nil
}

func (m *mockFeedbackStore) SummaryByChartID(userID, chartID uint) ([]model.FeedbackSummaryItem, int64, int64, error) {
	counts := map[string]model.FeedbackSummaryItem{}
	var total, researchEligible int64
	for _, item := range m.items {
		if item.UserID == userID && item.ChartID == chartID {
			key := item.TargetType + "\x00" + item.TargetID + "\x00" + item.Rating + "\x00" + item.EngineVersion + "\x00" + item.RuleVersion
			row := counts[key]
			row.TargetType, row.TargetID, row.Rating = item.TargetType, item.TargetID, item.Rating
			row.EngineVersion, row.RuleVersion, row.Count = item.EngineVersion, item.RuleVersion, row.Count+1
			counts[key] = row
			total++
			if item.ConsentResearch {
				researchEligible++
			}
		}
	}
	out := make([]model.FeedbackSummaryItem, 0, len(counts))
	for _, item := range counts {
		out = append(out, item)
	}
	return out, total, researchEligible, nil
}

func setupFeedbackRouter(chart *model.BirthChart, store *mockFeedbackStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")
	r := gin.New()
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	handler.RegisterFeedbackRoutes(api, &mockFeedbackChartStore{chart: chart}, store)
	return r
}

func feedbackBody(t *testing.T, v interface{}) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	return strings.NewReader(string(b))
}

func feedbackChart(userID uint) *model.BirthChart {
	chart := &model.BirthChart{
		UserID: userID, EngineVersion: "engine-stored", RuleVersion: "rule-stored",
		BirthYear: 1990, BirthMonth: 6, BirthDay: 15, BirthHour: 8, BirthMin: 30,
		Gender: "男", CalendarType: model.CalendarSolar, Timezone: "Asia/Shanghai",
	}
	chart.ID = 1
	return chart
}

func TestFeedbackNoJWT(t *testing.T) {
	router := setupFeedbackRouter(feedbackChart(1), &mockFeedbackStore{})
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", feedbackBody(t, model.FeedbackRequest{
		ChartID: 1,
		Rating:  model.FeedbackRatingAccurate,
	}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFeedbackOtherUsersChartNotFound(t *testing.T) {
	router := setupFeedbackRouter(feedbackChart(2), &mockFeedbackStore{})
	token, _ := middleware.GenerateToken(1, "testuser")
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", feedbackBody(t, model.FeedbackRequest{
		ChartID: 1,
		Rating:  model.FeedbackRatingAccurate,
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFeedbackInvalidRating(t *testing.T) {
	router := setupFeedbackRouter(feedbackChart(1), &mockFeedbackStore{})
	token, _ := middleware.GenerateToken(1, "testuser")
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", feedbackBody(t, model.FeedbackRequest{
		ChartID: 1,
		Rating:  "great",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestFeedbackCreateOK(t *testing.T) {
	store := &mockFeedbackStore{}
	router := setupFeedbackRouter(feedbackChart(1), store)
	token, _ := middleware.GenerateToken(1, "testuser")
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", feedbackBody(t, model.FeedbackRequest{
		ChartID:         1,
		TargetType:      "interpretation_section",
		TargetID:        "pattern",
		Rating:          model.FeedbackRatingHelpful,
		Tags:            []string{"格局", "格局"},
		Comment:         "很有帮助",
		ConsentResearch: true,
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if len(store.items) != 1 {
		t.Fatalf("expected 1 saved item, got %d", len(store.items))
	}
	if store.items[0].ConsentTraining {
		t.Fatal("consent_training should default to false")
	}
	if store.items[0].Tags != `["格局"]` {
		t.Fatalf("expected deduplicated tags, got %s", store.items[0].Tags)
	}
	if store.items[0].EngineVersion != bazipkg.EngineVersion || store.items[0].RuleVersion != bazipkg.RuleVersion {
		t.Fatalf("feedback versions = %q/%q", store.items[0].EngineVersion, store.items[0].RuleVersion)
	}
}

func TestFeedbackCreateIgnoresClientSuppliedVersions(t *testing.T) {
	store := &mockFeedbackStore{}
	router := setupFeedbackRouter(feedbackChart(1), store)
	token, _ := middleware.GenerateToken(1, "testuser")
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", feedbackBody(t, map[string]interface{}{
		"chart_id": 1, "rating": model.FeedbackRatingAccurate,
		"engine_version": "spoofed-engine", "rule_version": "spoofed-rule",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if store.items[0].EngineVersion != bazipkg.EngineVersion || store.items[0].RuleVersion != bazipkg.RuleVersion {
		t.Fatalf("client-supplied versions overrode authoritative chart versions: %+v", store.items[0])
	}
}

func TestFeedbackCreateBindsCurrentInterpretationVersionWhenChartHasHistoricalSnapshot(t *testing.T) {
	chart := feedbackChart(1)
	normalized, err := bazipkg.NormalizeBirthInput(bazipkg.BirthInput{
		Year: chart.BirthYear, Month: chart.BirthMonth, Day: chart.BirthDay,
		Hour: chart.BirthHour, Minute: chart.BirthMin,
		CalendarType: chart.CalendarType, Gender: model.GenderMale, Timezone: chart.Timezone,
	})
	if err != nil {
		t.Fatalf("normalize historical chart: %v", err)
	}
	snapshot, err := (&bazipkg.BaziService{}).CalculateNormalizedBirth(normalized)
	if err != nil {
		t.Fatalf("calculate historical snapshot: %v", err)
	}
	snapshot.RuleVersion = "historical-rule"
	snapshot.School = "historical-school"
	snapshot.RuleMeta.RuleVersion = snapshot.RuleVersion
	snapshot.RuleMeta.School = snapshot.School
	snapshot.BodyStrength.RuleVersion = snapshot.RuleVersion
	snapshot.BodyStrength.School = snapshot.School

	normalizedJSON, err := json.Marshal(normalized)
	if err != nil {
		t.Fatalf("marshal normalized birth: %v", err)
	}
	snapshotJSON, err := json.Marshal(snapshot)
	if err != nil {
		t.Fatalf("marshal historical snapshot: %v", err)
	}
	chart.EngineVersion = "historical-engine"
	chart.RuleVersion = snapshot.RuleVersion
	chart.NormalizedBirth = datatypes.JSON(normalizedJSON)
	chart.BaziSnapshot = datatypes.JSON(snapshotJSON)

	store := &mockFeedbackStore{}
	router := setupFeedbackRouter(chart, store)
	token, _ := middleware.GenerateToken(1, "testuser")
	req := httptest.NewRequest(http.MethodPost, "/api/feedback", feedbackBody(t, model.FeedbackRequest{
		ChartID: 1, TargetType: "interpretation_section", TargetID: "pattern", Rating: model.FeedbackRatingAccurate,
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if len(store.items) != 1 {
		t.Fatalf("expected 1 saved item, got %d", len(store.items))
	}
	if store.items[0].EngineVersion != bazipkg.EngineVersion || store.items[0].RuleVersion != bazipkg.RuleVersion {
		t.Fatalf("feedback was attributed to historical snapshot instead of displayed interpretation: %+v", store.items[0])
	}
}

func TestFeedbackSummary(t *testing.T) {
	store := &mockFeedbackStore{
		items: []model.FortuneFeedback{
			{UserID: 1, ChartID: 1, TargetType: "interpretation_section", TargetID: "pattern", Rating: model.FeedbackRatingAccurate, EngineVersion: "e1", RuleVersion: "r1", ConsentResearch: true},
			{UserID: 1, ChartID: 1, TargetType: "interpretation_section", TargetID: "pattern", Rating: model.FeedbackRatingAccurate, EngineVersion: "e1", RuleVersion: "r1", ConsentResearch: true},
			{UserID: 1, ChartID: 1, TargetType: "interpretation_section", TargetID: "tiaohou", Rating: model.FeedbackRatingConfusing, EngineVersion: "e2", RuleVersion: "r2"},
			{UserID: 2, ChartID: 1, Rating: model.FeedbackRatingHelpful},
		},
	}
	router := setupFeedbackRouter(feedbackChart(1), store)
	token, _ := middleware.GenerateToken(1, "testuser")
	req := httptest.NewRequest(http.MethodGet, "/api/feedback/summary?chart_id=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp model.FeedbackSummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Total != 3 {
		t.Fatalf("expected total 3, got %d", resp.Total)
	}
	if resp.ResearchEligible != 2 || resp.Scope != model.FeedbackSummaryScopeInterpretationQuality || len(resp.Items) != 2 {
		t.Fatalf("feedback quality summary lost consent, target, or version dimensions: %+v", resp)
	}
}
