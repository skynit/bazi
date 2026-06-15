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

	"github.com/gin-gonic/gin"
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

func (m *mockFeedbackStore) SummaryByChartID(userID, chartID uint) ([]model.FeedbackSummaryItem, int64, error) {
	counts := map[string]int64{}
	var total int64
	for _, item := range m.items {
		if item.UserID == userID && item.ChartID == chartID {
			counts[item.Rating]++
			total++
		}
	}
	out := make([]model.FeedbackSummaryItem, 0, len(counts))
	for rating, count := range counts {
		out = append(out, model.FeedbackSummaryItem{Rating: rating, Count: count})
	}
	return out, total, nil
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
	chart := &model.BirthChart{UserID: userID}
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
}

func TestFeedbackSummary(t *testing.T) {
	store := &mockFeedbackStore{
		items: []model.FortuneFeedback{
			{UserID: 1, ChartID: 1, Rating: model.FeedbackRatingAccurate},
			{UserID: 1, ChartID: 1, Rating: model.FeedbackRatingAccurate},
			{UserID: 1, ChartID: 1, Rating: model.FeedbackRatingConfusing},
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
}
