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

func setupZiWeiRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT("test-secret")

	r := gin.New()
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())
	svc := ziwei.NewZiWeiService()
	RegisterZiWeiRoutes(api, svc)
	return r
}

func getValidJWT(t *testing.T) string {
	t.Helper()
	middleware.InitJWT("test-secret")
	token, err := middleware.GenerateToken(1, "testuser")
	if err != nil {
		t.Fatalf("failed to generate test token: %v", err)
	}
	return token
}

func TestZiWeiChart_Success(t *testing.T) {
	router := setupZiWeiRouter()
	token := getValidJWT(t)

	body := jsonBody(t, model.ChartRequest{
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		ProfileID          string                      `json:"profile_id"`
		EngineVersion      string                      `json:"engine_version"`
		RuleVersion        string                      `json:"rule_version"`
		RuleSchool         string                      `json:"rule_school"`
		RuleSources        []ziwei.RuleSourceRef       `json:"rule_sources"`
		PluginManifest     []ziwei.PluginRequirement   `json:"plugin_manifest"`
		PluginManifestHash string                      `json:"plugin_manifest_hash"`
		CalculationInput   ziwei.ZiWeiCalculationInput `json:"calculation_input"`
		InputFingerprint   string                      `json:"input_fingerprint"`
		ContentHash        string                      `json:"content_hash"`
		Palaces            []struct {
			Name   string `json:"name"`
			Branch string `json:"branch"`
		} `json:"palaces"`
		FiveBureau                string `json:"five_bureau"`
		EarthlyBranchOfSoulPalace string `json:"earthly_branch_of_soul_palace"`
		EarthlyBranchOfBodyPalace string `json:"earthly_branch_of_body_palace"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(resp.Palaces) != 12 {
		t.Errorf("expected 12 palaces, got %d", len(resp.Palaces))
	}
	profile, err := ziwei.ResolveProfile(ziwei.DefaultProfileID)
	if err != nil {
		t.Fatalf("resolve default profile: %v", err)
	}
	if !reflect.DeepEqual(resp.RuleSources, profile.RuleSources) {
		t.Fatalf("rule sources = %+v, want profile sources %+v", resp.RuleSources, profile.RuleSources)
	}

	for i, p := range resp.Palaces {
		if p.Name == "" {
			t.Errorf("palace %d has empty name", i)
		}
	}

	if len(resp.Palaces) > 0 && resp.Palaces[0].Name != "命宫" {
		t.Errorf("expected first palace to be 命宫, got %s", resp.Palaces[0].Name)
	}
	if len(resp.Palaces) > 0 && resp.Palaces[0].Branch != resp.EarthlyBranchOfSoulPalace {
		t.Errorf("命宫 branch = %s, want soul palace branch %s", resp.Palaces[0].Branch, resp.EarthlyBranchOfSoulPalace)
	}

	if resp.FiveBureau == "" {
		t.Error("FiveBureau is empty")
	}
	if resp.EarthlyBranchOfBodyPalace == "" {
		t.Error("EarthlyBranchOfBodyPalace is empty")
	}
	if resp.ProfileID != ziwei.DefaultProfileID || resp.EngineVersion != ziwei.ZiWeiEngineVersion ||
		resp.RuleVersion != ziwei.ZiWeiRuleVersion || resp.RuleSchool != ziwei.ZiWeiRuleSchool {
		t.Fatalf("unexpected reproducibility metadata: %+v", resp)
	}
	if len(resp.PluginManifest) != 0 || len(resp.PluginManifestHash) != 64 {
		t.Fatalf("unexpected plugin contract: %+v hash=%q", resp.PluginManifest, resp.PluginManifestHash)
	}
	if len(resp.InputFingerprint) != 64 || len(resp.ContentHash) != 64 {
		t.Fatalf("unexpected cache contract: input=%q content=%q", resp.InputFingerprint, resp.ContentHash)
	}
	if resp.CalculationInput.CalendarType != "SOLAR" || resp.CalculationInput.Basis != "normalized_solar_minute" {
		t.Fatalf("unexpected calculation input: %+v", resp.CalculationInput)
	}
}

func TestZiWeiChart_DirectLunarRequestUsesNormalizedSolarInput(t *testing.T) {
	expected, err := bazi.NormalizeBirthInput(bazi.BirthInput{
		Year: 2024, Month: 1, Day: 1, Hour: 8, Minute: 30,
		CalendarType: model.CalendarLunar, Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	router := setupZiWeiRouter()
	token := getValidJWT(t)
	body := jsonBody(t, map[string]interface{}{
		"birth_year": 2024, "birth_month": 1, "birth_day": 1,
		"birth_hour": 8, "birth_min": 30, "calendar_type": model.CalendarLunar,
		"gender": "男",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("lunar chart request failed: %d %s", response.Code, response.Body.String())
	}
	var payload struct {
		CalculationInput ziwei.ZiWeiCalculationInput `json:"calculation_input"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	got := payload.CalculationInput
	if got.Year != expected.Year || got.Month != expected.Month || got.Day != expected.Day ||
		got.Hour != expected.Hour || got.Minute != expected.Minute || got.Gender != "男" {
		t.Fatalf("calculation input = %+v, normalized request = %+v", got, expected)
	}
}

func TestZiWeiChart_RejectsIgnoredAlgorithmParameter(t *testing.T) {
	router := setupZiWeiRouter()
	token := getValidJWT(t)
	body := jsonBody(t, map[string]interface{}{
		"birth_year": 1984, "birth_month": 2, "birth_day": 15, "birth_hour": 8,
		"gender": "男", "algorithm": "zhongzhou",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestZiWeiChart_RejectsUnknownProfile(t *testing.T) {
	router := setupZiWeiRouter()
	token := getValidJWT(t)
	body := jsonBody(t, map[string]interface{}{
		"birth_year": 1984, "birth_month": 2, "birth_day": 15, "birth_hour": 8,
		"gender": "男", "profile": "ziwei-unknown-v1",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestZiWeiChart_Unauthorized(t *testing.T) {
	router := setupZiWeiRouter()

	body := jsonBody(t, model.ChartRequest{
		BirthYear:  1984,
		BirthMonth: 2,
		BirthDay:   15,
		BirthHour:  8,
		BirthMin:   0,
		Gender:     "男",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/ziwei/chart", body)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d: %s", w.Code, w.Body.String())
	}
}
