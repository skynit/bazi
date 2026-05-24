package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service"

	"github.com/gin-gonic/gin"
)

// ziweiCachingStore implements ChartStore for testing ZiWei caching.
type ziweiCachingStore struct {
	charts map[uint]*model.BirthChart
	nextID uint
}

func newZiWeiCachingStore() *ziweiCachingStore {
	return &ziweiCachingStore{charts: make(map[uint]*model.BirthChart), nextID: 1}
}

func (s *ziweiCachingStore) Create(c *model.BirthChart) error {
	c.ID = s.nextID
	s.nextID++
	s.charts[c.ID] = c
	return nil
}

func (s *ziweiCachingStore) FindByID(id uint) (*model.BirthChart, error) {
	if c, ok := s.charts[id]; ok {
		return c, nil
	}
	return nil, nil
}

func (s *ziweiCachingStore) Update(c *model.BirthChart) error {
	if _, ok := s.charts[c.ID]; !ok {
		return nil
	}
	s.charts[c.ID] = c
	return nil
}

func setupZiWeiCachingRouter(store *ziweiCachingStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	svc := service.NewZiWeiService()
	RegisterZiWeiRoutesWithStore(api, svc, store)
	return r
}

func TestZiWeiChart_CachedResult(t *testing.T) {
	store := newZiWeiCachingStore()
	chart := &model.BirthChart{
		Name:       "test",
		Gender:     "男",
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
	}
	if err := store.Create(chart); err != nil {
		t.Fatalf("create chart: %v", err)
	}

	if chart.ZiWeiComputed {
		t.Fatal("chart should not be computed yet")
	}

	router := setupZiWeiCachingRouter(store)
	token := getValidJWT(t)

	// First call: should compute and store
	reqBody := map[string]interface{}{"chart_id": chart.ID}
	body1 := jsonBody(t, reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body1)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("first call: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify chart was cached
	updated, _ := store.FindByID(chart.ID)
	if updated == nil {
		t.Fatal("chart not found after caching")
	}
	if !updated.ZiWeiComputed {
		t.Error("chart should be marked as computed after first call")
	}
	if len(updated.ZiWeiResult) == 0 {
		t.Error("ziwei_result should not be empty after caching")
	}

	// Second call: should return cached result
	body2 := jsonBody(t, reqBody)
	req2 := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body2)
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("second call: expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var resp struct {
		Palaces []struct {
			Name string `json:"name"`
		} `json:"palaces"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal cached response: %v", err)
	}
	if len(resp.Palaces) != 12 {
		t.Errorf("cached: expected 12 palaces, got %d", len(resp.Palaces))
	}
}

func TestZiWeiChart_InvalidChartID(t *testing.T) {
	store := newZiWeiCachingStore()
	router := setupZiWeiCachingRouter(store)
	token := getValidJWT(t)

	reqBody := map[string]interface{}{"chart_id": 999}
	body := jsonBody(t, reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unknown chart_id, got %d: %s", w.Code, w.Body.String())
	}
}
