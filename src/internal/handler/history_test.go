package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"

	"github.com/gin-gonic/gin"
)

type mockChartListStore struct {
	mu     sync.RWMutex
	charts map[uint]*model.BirthChart
	idSeq  uint
}

func newMockChartListStore() *mockChartListStore {
	return &mockChartListStore{charts: make(map[uint]*model.BirthChart)}
}

func (m *mockChartListStore) FindByID(id uint) (*model.BirthChart, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.charts[id]
	if !ok {
		return nil, nil
	}
	return c, nil
}

func (m *mockChartListStore) ListByUser(userID uint, page, pageSize int) ([]model.BirthChart, int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var all []model.BirthChart
	for _, c := range m.charts {
		if c.UserID == userID {
			all = append(all, *c)
		}
	}

	total := int64(len(all))
	start := (page - 1) * pageSize
	if start >= len(all) {
		return []model.BirthChart{}, total, nil
	}
	end := start + pageSize
	if end > len(all) {
		end = len(all)
	}
	return all[start:end], total, nil
}

func (m *mockChartListStore) AddChart(c *model.BirthChart) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.idSeq++
	c.ID = m.idSeq
	m.charts[c.ID] = c
}

type mockFortuneHistoryStore struct {
	history []model.HistoryResponse
}

func (m *mockFortuneHistoryStore) ListByChartID(chartID uint, page, pageSize int) ([]model.HistoryResponse, int64, error) {
	var filtered []model.HistoryResponse
	for _, h := range m.history {
		if h.ChartID == chartID {
			filtered = append(filtered, h)
		}
	}

	total := int64(len(filtered))
	start := (page - 1) * pageSize
	if start >= len(filtered) {
		return []model.HistoryResponse{}, total, nil
	}
	end := start + pageSize
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[start:end], total, nil
}

func setupHistoryRouter(charts *mockChartListStore, fortune *mockFortuneHistoryStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	RegisterHistoryRoutes(r, charts, fortune)
	return r
}

func TestListChartsPaginated(t *testing.T) {
	charts := newMockChartListStore()
	charts.AddChart(&model.BirthChart{UserID: 1, Name: "Chart A", BirthYear: 1990, BirthMonth: 1, BirthDay: 1, BirthHour: 0, Gender: "男"})
	charts.AddChart(&model.BirthChart{UserID: 1, Name: "Chart B", BirthYear: 1995, BirthMonth: 6, BirthDay: 15, BirthHour: 12, Gender: "女"})

	fortune := &mockFortuneHistoryStore{}
	router := setupHistoryRouter(charts, fortune)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/charts?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Charts   []model.BirthChart `json:"charts"`
		Total    int64              `json:"total"`
		Page     int                `json:"page"`
		PageSize int                `json:"page_size"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
	if len(resp.Charts) != 2 {
		t.Errorf("expected 2 charts, got %d", len(resp.Charts))
	}
}

func TestGetChartByID(t *testing.T) {
	charts := newMockChartListStore()
	charts.AddChart(&model.BirthChart{UserID: 1, Name: "Test Chart", BirthYear: 2000, BirthMonth: 3, BirthDay: 20, BirthHour: 10, Gender: "男"})

	fortune := &mockFortuneHistoryStore{}
	router := setupHistoryRouter(charts, fortune)

	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/charts/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var chart model.BirthChart
	if err := json.Unmarshal(w.Body.Bytes(), &chart); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if chart.Name != "Test Chart" {
		t.Errorf("expected chart name 'Test Chart', got '%s'", chart.Name)
	}
}

func TestHistoryNoJWT(t *testing.T) {
	charts := newMockChartListStore()
	fortune := &mockFortuneHistoryStore{}
	router := setupHistoryRouter(charts, fortune)

	req := httptest.NewRequest(http.MethodGet, "/api/charts", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
