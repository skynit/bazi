package handler_test

import (
	. "bazi/internal/handler"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service/bazi"
	"bazi/internal/service/ziwei"

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
func (s *ziweiCachingStore) FindByIDForUser(id uint, userID uint) (*model.BirthChart, error) {
	if c, ok := s.charts[id]; ok && c.UserID == userID {
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
	svc := ziwei.NewZiWeiService()
	RegisterZiWeiRoutesWithStore(api, svc, store)
	return r
}

func TestZiWeiChart_CachedResult(t *testing.T) {
	store := newZiWeiCachingStore()
	chart := &model.BirthChart{
		UserID:     1,
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
	var staleCached ziwei.ZiWeiChart
	if err := json.Unmarshal(updated.ZiWeiResult, &staleCached); err != nil {
		t.Fatalf("unmarshal stored chart: %v", err)
	}
	foundLucun := false
	for i := range staleCached.Palaces {
		for j := range staleCached.Palaces[i].Stars {
			star := &staleCached.Palaces[i].Stars[j]
			if star.Name == "禄存" {
				star.Brightness = "陷"
				foundLucun = true
			}
		}
	}
	if !foundLucun {
		t.Fatal("stored chart does not contain 禄存")
	}
	staleData, err := json.Marshal(&staleCached)
	if err != nil {
		t.Fatalf("marshal stale chart: %v", err)
	}
	updated.ZiWeiResult = staleData

	// Second call: structural content hash mismatch must force recomputation.
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
			Name  string             `json:"name"`
			Stars []ziwei.StarOutput `json:"stars"`
		} `json:"palaces"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal cached response: %v", err)
	}
	if len(resp.Palaces) != 12 {
		t.Errorf("cached: expected 12 palaces, got %d", len(resp.Palaces))
	}
	foundLucun = false
	for _, palace := range resp.Palaces {
		for _, star := range palace.Stars {
			if star.Name == "禄存" {
				foundLucun = true
				if star.Brightness != "" {
					t.Errorf("cached 禄存 brightness = %q, want empty because source has no table", star.Brightness)
				}
			}
		}
	}
	if !foundLucun {
		t.Error("cached response does not contain 禄存")
	}
	recomputed, _ := store.FindByID(chart.ID)
	var refreshed ziwei.ZiWeiChart
	if err := json.Unmarshal(recomputed.ZiWeiResult, &refreshed); err != nil {
		t.Fatalf("unmarshal refreshed chart: %v", err)
	}
	if refreshed.RuleVersion != ziwei.ZiWeiRuleVersion {
		t.Fatalf("stale cache was not replaced: rule_version=%q", refreshed.RuleVersion)
	}
	profile, err := ziwei.ResolveProfile(ziwei.DefaultProfileID)
	if err != nil {
		t.Fatal(err)
	}
	if refreshed.PluginManifestHash != profile.PluginManifestHash {
		t.Fatalf("stale plugin cache was not replaced: hash=%q", refreshed.PluginManifestHash)
	}
	if !reflect.DeepEqual(refreshed.RuleSources, profile.RuleSources) {
		t.Fatalf("stale rule source cache was not replaced: %+v", refreshed.RuleSources)
	}
}

func TestZiWeiChart_CacheInvalidatesWhenBirthInputChanges(t *testing.T) {
	store := newZiWeiCachingStore()
	chart := &model.BirthChart{
		UserID: 1, Name: "input-change", Gender: "男",
		BirthYear: 1984, BirthMonth: 2, BirthDay: 15, BirthHour: 8,
	}
	if err := store.Create(chart); err != nil {
		t.Fatal(err)
	}
	router := setupZiWeiCachingRouter(store)
	token := getValidJWT(t)
	requestChart := func() *httptest.ResponseRecorder {
		body := jsonBody(t, map[string]interface{}{"chart_id": chart.ID})
		req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		router.ServeHTTP(response, req)
		if response.Code != http.StatusOK {
			t.Fatalf("chart request failed: %d %s", response.Code, response.Body.String())
		}
		return response
	}

	requestChart()
	var first ziwei.ZiWeiChart
	if err := json.Unmarshal(chart.ZiWeiResult, &first); err != nil {
		t.Fatal(err)
	}
	if len(first.InputFingerprint) != 64 || len(first.ContentHash) != 64 {
		t.Fatalf("initial cache contract is incomplete: %+v", first)
	}

	chart.BirthHour = 10
	requestChart()
	var refreshed ziwei.ZiWeiChart
	if err := json.Unmarshal(chart.ZiWeiResult, &refreshed); err != nil {
		t.Fatal(err)
	}
	if refreshed.InputFingerprint == first.InputFingerprint {
		t.Fatal("birth input change did not invalidate the cached input fingerprint")
	}
	if !ziwei.NewZiWeiService().ChartMatchesInputProfile(
		&refreshed, ziwei.DefaultProfileID,
		chart.BirthYear, chart.BirthMonth, chart.BirthDay, chart.BirthHour, chart.BirthMin, chart.Gender,
	) {
		t.Fatal("refreshed cache does not match the updated birth input")
	}
}

func TestZiWeiChart_UsesPersistedNormalizedBirthInsteadOfRawLunarDate(t *testing.T) {
	normalized, err := bazi.NormalizeBirthInput(bazi.BirthInput{
		Year: 2024, Month: 1, Day: 1, Hour: 8, Minute: 30,
		CalendarType: model.CalendarLunar, Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	if normalized.Month == 1 && normalized.Day == 1 {
		t.Fatal("test lunar date unexpectedly equals its converted solar date")
	}
	normalizedJSON, err := json.Marshal(normalized)
	if err != nil {
		t.Fatal(err)
	}

	store := newZiWeiCachingStore()
	chart := &model.BirthChart{
		UserID: 1, Name: "lunar-normalized", Gender: model.GenderMale,
		BirthYear: 2024, BirthMonth: 1, BirthDay: 1, BirthHour: 8, BirthMin: 30,
		CalendarType: model.CalendarLunar, NormalizedBirth: normalizedJSON,
	}
	if err := store.Create(chart); err != nil {
		t.Fatal(err)
	}
	router := setupZiWeiCachingRouter(store)
	token := getValidJWT(t)
	body := jsonBody(t, map[string]interface{}{"chart_id": chart.ID})
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("chart request failed: %d %s", response.Code, response.Body.String())
	}

	var payload struct {
		CalculationInput ziwei.ZiWeiCalculationInput `json:"calculation_input"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	got := payload.CalculationInput
	if got.Year != normalized.Year || got.Month != normalized.Month || got.Day != normalized.Day ||
		got.Hour != normalized.Hour || got.Minute != normalized.Minute || got.Gender != "男" {
		t.Fatalf("ziwei calculation input = %+v, normalized birth = %+v", got, normalized)
	}
	if got.Month == chart.BirthMonth && got.Day == chart.BirthDay {
		t.Fatalf("ziwei calculation silently reused raw lunar fields: %+v", got)
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
