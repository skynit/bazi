package handler_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"bazi/internal/handler"
	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service/buyi"

	"github.com/gin-gonic/gin"
)

type buyiMockStore struct {
	record               *model.BuyiRecord
	createErr            error
	recordAfterCreateErr *model.BuyiRecord
	createCount          int
}

func (s *buyiMockStore) Create(record *model.BuyiRecord) error {
	s.createCount++
	if s.createErr != nil {
		if s.recordAfterCreateErr != nil {
			s.record = s.recordAfterCreateErr
		}
		return s.createErr
	}
	if record.ID == 0 {
		record.ID = uint(s.createCount)
	}
	if record.CreatedAt.IsZero() {
		record.CreatedAt = time.Date(2026, 7, 5, 9, 0, 0, 0, time.FixedZone("CST", 8*3600))
	}
	copied := *record
	s.record = &copied
	return nil
}

func (s *buyiMockStore) FindByUserDate(userID uint, date time.Time) (*model.BuyiRecord, error) {
	if s.record == nil {
		return nil, nil
	}
	if s.record.UserID == userID && sameDate(s.record.DivinationDate, date) {
		return s.record, nil
	}
	return nil, nil
}

func sameDate(a, b time.Time) bool {
	return a.Format("2006-01-02") == b.Format("2006-01-02")
}

func setupBuyiRouter(store *buyiMockStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	fixedNow := time.Date(2026, 7, 5, 15, 30, 0, 0, time.FixedZone("CST", 8*3600))
	h := &handler.BuyiHandler{
		Service: buyi.NewService(),
		Store:   store,
		Now:     func() time.Time { return fixedNow },
	}

	r := gin.New()
	r.GET("/api/buyi/today", middleware.AuthMiddleware(), h.Today)
	r.POST("/api/buyi/today", middleware.AuthMiddleware(), h.DrawToday)
	return r
}

func buyiToken(t *testing.T) string {
	t.Helper()
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func decodeJSON(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func TestBuyiTodayNoJWT(t *testing.T) {
	router := setupBuyiRouter(&buyiMockStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/buyi/today", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBuyiTodayEmpty(t *testing.T) {
	router := setupBuyiRouter(&buyiMockStore{})
	req := httptest.NewRequest(http.MethodGet, "/api/buyi/today", nil)
	req.Header.Set("Authorization", "Bearer "+buyiToken(t))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.BuyiTodayResponse
	if err := decodeJSON(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Date != "2026-07-05" || resp.HasRecord || resp.AlreadyDrawn || resp.Record != nil {
		t.Fatalf("unexpected empty response: %+v", resp)
	}
}

func TestBuyiDrawCreatesAndRepeats(t *testing.T) {
	store := &buyiMockStore{}
	router := setupBuyiRouter(store)
	token := buyiToken(t)

	req := httptest.NewRequest(http.MethodPost, "/api/buyi/today", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("first draw expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var first model.BuyiTodayResponse
	if err := decodeJSON(w.Body.Bytes(), &first); err != nil {
		t.Fatalf("decode first response: %v", err)
	}
	if !first.HasRecord || first.AlreadyDrawn || first.Record == nil {
		t.Fatalf("unexpected first draw response: %+v", first)
	}
	if first.Record.HexagramNumber < 1 || first.Record.HexagramNumber > 64 || first.Record.HexagramName == "" {
		t.Fatalf("unexpected hexagram record: %+v", first.Record)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/api/buyi/today", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("repeat draw expected 200, got %d: %s", w2.Code, w2.Body.String())
	}

	var second model.BuyiTodayResponse
	if err := decodeJSON(w2.Body.Bytes(), &second); err != nil {
		t.Fatalf("decode second response: %v", err)
	}
	if !second.HasRecord || !second.AlreadyDrawn || second.Record == nil {
		t.Fatalf("unexpected repeat response: %+v", second)
	}
	if second.Record.ID != first.Record.ID || second.Record.HexagramName != first.Record.HexagramName {
		t.Fatalf("repeat draw returned different record: first=%+v second=%+v", first.Record, second.Record)
	}
	if store.createCount != 1 {
		t.Fatalf("expected one create, got %d", store.createCount)
	}
}

func TestBuyiDrawReadsExistingAfterCreateConflict(t *testing.T) {
	date := time.Date(2026, 7, 5, 0, 0, 0, 0, time.FixedZone("CST", 8*3600))
	store := &buyiMockStore{
		createErr: errors.New("duplicate key"),
		recordAfterCreateErr: &model.BuyiRecord{
			UserID:         1,
			DivinationDate: date,
			HexagramNumber: 42,
			HexagramName:   "风雷益",
			Score:          86,
			Level:          "大吉",
			Summary:        "今日得风雷益",
			HumanWay:       "损上益下，普施恩泽。",
			ImageReading:   "风雷相激、双重增益、获利、分红。",
			Advice:         "今日气势较顺。",
		},
	}
	store.recordAfterCreateErr.ID = 99
	router := setupBuyiRouter(store)
	req := httptest.NewRequest(http.MethodPost, "/api/buyi/today", nil)
	req.Header.Set("Authorization", "Bearer "+buyiToken(t))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 after conflict readback, got %d: %s", w.Code, w.Body.String())
	}

	var resp model.BuyiTodayResponse
	if err := decodeJSON(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !resp.HasRecord || !resp.AlreadyDrawn || resp.Record == nil || resp.Record.ID != 99 {
		t.Fatalf("unexpected conflict response: %+v", resp)
	}
}
