package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"bazi/internal/handler"
	"bazi/internal/middleware"
	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/interpretation"
	"bazi/internal/service/rag"

	"github.com/gin-gonic/gin"
)

type mockInterpretationChartStore struct {
	chart *model.BirthChart
}

func (m *mockInterpretationChartStore) FindByIDForUser(id uint, userID uint) (*model.BirthChart, error) {
	if m.chart != nil && m.chart.ID == id && m.chart.UserID == userID {
		return m.chart, nil
	}
	return nil, nil
}

type mockRetriever struct {
	chunks []rag.RetrievedChunk
	err    error
}

func (m mockRetriever) Retrieve(ctx context.Context, req rag.RetrieveRequest) ([]rag.RetrievedChunk, error) {
	return m.chunks, m.err
}

func setupInterpretationRouter(store interpretation.ChartStore, retriever rag.Retriever) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	svc := &interpretation.Service{
		Charts:    store,
		Bazi:      &bazipkg.BaziService{},
		Retriever: retriever,
		MinScore:  0.35,
		TopK:      8,
	}
	h := &handler.InterpretationHandler{Service: svc}
	r.POST("/api/interpretation/bazi", middleware.AuthMiddleware(), h.Bazi)
	return r
}

func interpretationBody(t *testing.T, v interface{}) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("failed to marshal request: %v", err)
	}
	return strings.NewReader(string(b))
}

func testChart(userID uint) *model.BirthChart {
	chart := &model.BirthChart{
		UserID:       userID,
		BirthYear:    1990,
		BirthMonth:   6,
		BirthDay:     15,
		BirthHour:    8,
		BirthMin:     0,
		CalendarType: model.CalendarSolar,
		Gender:       model.GenderMale,
	}
	chart.ID = 1
	return chart
}

func TestBaziInterpretationNoJWT(t *testing.T) {
	router := setupInterpretationRouter(&mockInterpretationChartStore{chart: testChart(1)}, mockRetriever{})

	req := httptest.NewRequest(http.MethodPost, "/api/interpretation/bazi", interpretationBody(t, model.BaziInterpretationRequest{ChartID: 1}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBaziInterpretationMissingChartID(t *testing.T) {
	router := setupInterpretationRouter(&mockInterpretationChartStore{chart: testChart(1)}, mockRetriever{})
	token, _ := middleware.GenerateToken(1, "testuser")

	req := httptest.NewRequest(http.MethodPost, "/api/interpretation/bazi", interpretationBody(t, map[string]interface{}{"focus": "overview"}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBaziInterpretationOtherUsersChartNotFound(t *testing.T) {
	router := setupInterpretationRouter(&mockInterpretationChartStore{chart: testChart(2)}, mockRetriever{})
	token, _ := middleware.GenerateToken(1, "testuser")

	req := httptest.NewRequest(http.MethodPost, "/api/interpretation/bazi", interpretationBody(t, model.BaziInterpretationRequest{ChartID: 1}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBaziInterpretationOK(t *testing.T) {
	router := setupInterpretationRouter(&mockInterpretationChartStore{chart: testChart(1)}, mockRetriever{
		chunks: []rag.RetrievedChunk{{
			ID:      "chunk-1",
			Content: "月令为提纲，格局当以月令为主。",
			Score:   0.82,
			Metadata: map[string]string{
				"domain":      "bazi",
				"book":        "子平真诠",
				"chapter":     "001",
				"source_path": "bazi/子平真诠/001.md",
			},
		}},
	})
	token, _ := middleware.GenerateToken(1, "testuser")

	req := httptest.NewRequest(http.MethodPost, "/api/interpretation/bazi", interpretationBody(t, model.BaziInterpretationRequest{
		ChartID: 1,
		Focus:   "pattern",
	}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.BaziInterpretationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Status != "ok" {
		t.Fatalf("expected status ok, got %q: %+v", resp.Status, resp)
	}
	if len(resp.Citations) != 1 || resp.Citations[0].Book != "子平真诠" {
		t.Fatalf("unexpected citations: %+v", resp.Citations)
	}
}

func TestBaziInterpretationFallbackWhenDisabled(t *testing.T) {
	router := setupInterpretationRouter(&mockInterpretationChartStore{chart: testChart(1)}, nil)
	token, _ := middleware.GenerateToken(1, "testuser")

	req := httptest.NewRequest(http.MethodPost, "/api/interpretation/bazi", interpretationBody(t, model.BaziInterpretationRequest{ChartID: 1}))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 fallback, got %d: %s", w.Code, w.Body.String())
	}
	var resp model.BaziInterpretationResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Status != "fallback" || resp.Reason != "disabled" {
		t.Fatalf("expected disabled fallback, got %+v", resp)
	}
}
