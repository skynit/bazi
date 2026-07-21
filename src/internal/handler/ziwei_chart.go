package handler

import (
	"bazi/internal/model"
	"bazi/internal/service/ziwei"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"net/http"
	"strings"
)

type ZiWeiChartHandler struct {
	Service *ziwei.ZiWeiService
	Charts  ChartStore
}

type ZiWeiPalaceResponse struct {
	Name           string             `json:"name"`
	Branch         string             `json:"branch"`
	HeavenlyStem   string             `json:"heavenly_stem"`
	IsBodyPalace   bool               `json:"is_body_palace"`
	Stars          []ziwei.StarOutput `json:"stars"`
	FourHua        []string           `json:"four_hua"`
	AdjectiveStars []string           `json:"adjective_stars,omitempty"`
	Changsheng12   string             `json:"changsheng_12,omitempty"`
	Boshi12        string             `json:"boshi_12,omitempty"`
	JiangQian12    string             `json:"jiang_qian_12,omitempty"`
	SuiQian12      string             `json:"sui_qian_12,omitempty"`
	SanfangSizheng *SanfangResponse   `json:"sanfang_sizheng,omitempty"`
}

// SanfangResponse holds the sanfang sizheng data for a palace.
type SanfangResponse struct {
	Opposite string `json:"opposite"`
	Trine1   string `json:"trine1"`
	Trine2   string `json:"trine2"`
}

func mapPalaceToResponse(p *ziwei.PalaceInfo, sf *ziwei.SanfangSizhengResult) ZiWeiPalaceResponse {
	resp := ZiWeiPalaceResponse{
		Name:           p.Name,
		Branch:         p.Branch,
		HeavenlyStem:   p.HeavenlyStem,
		IsBodyPalace:   p.IsBodyPalace,
		Stars:          p.Stars,
		FourHua:        p.FourHua,
		AdjectiveStars: p.AdjectiveStars,
		Changsheng12:   p.Changsheng12,
		Boshi12:        p.Boshi12,
		JiangQian12:    p.JiangQian12,
		SuiQian12:      p.SuiQian12,
	}
	if sf != nil {
		resp.SanfangSizheng = &SanfangResponse{
			Opposite: sf.Opposite,
			Trine1:   sf.Trine1,
			Trine2:   sf.Trine2,
		}
	}
	return resp
}

func mapChartToResponse(chart *ziwei.ZiWeiChart, svc *ziwei.ZiWeiService) gin.H {
	palaces := make([]ZiWeiPalaceResponse, 12)
	for i := 0; i < 12; i++ {
		sf := &ziwei.SanfangSizhengResult{
			Opposite: chart.SanfangSizheng[i].Opposite,
			Trine1:   chart.SanfangSizheng[i].Trine1,
			Trine2:   chart.SanfangSizheng[i].Trine2,
		}
		palaces[i] = mapPalaceToResponse(&chart.Palaces[i], sf)
	}

	response := gin.H{
		"profile_id":                    chart.ProfileID,
		"engine_version":                chart.EngineVersion,
		"rule_version":                  chart.RuleVersion,
		"rule_school":                   chart.RuleSchool,
		"rule_sources":                  chart.RuleSources,
		"runtime_rule_tables_schema":    chart.RuntimeRuleTablesSchema,
		"runtime_rule_tables_hash":      chart.RuntimeRuleTablesHash,
		"plugin_manifest":               chart.PluginManifest,
		"plugin_manifest_hash":          chart.PluginManifestHash,
		"calculation_input":             chart.CalculationInput,
		"input_fingerprint":             chart.InputFingerprint,
		"palaces":                       palaces,
		"life_master":                   chart.LifeMaster,
		"body_master":                   chart.BodyMaster,
		"five_bureau":                   chart.FiveBureau,
		"body_palace":                   chart.BodyPalace,
		"earthly_branch_of_soul_palace": chart.EarthlyBranchOfSoulPalace,
		"earthly_branch_of_body_palace": chart.EarthlyBranchOfBodyPalace,
		"patterns":                      chart.Patterns,
		"liu_nian_stars":                chart.LiuNianStars,
		"liu_yue_stars":                 chart.LiuYueStars,
		"liu_ri_stars":                  chart.LiuRiStars,
		"liu_nian_four_hua":             chart.LiuNianFourHua,
		"liu_yue_four_hua":              chart.LiuYueFourHua,
		"liu_ri_four_hua":               chart.LiuRiFourHua,
		"liu_nian_palaces":              chart.LiuNianPalaces,
		"liu_yue_palaces":               chart.LiuYuePalaces,
		"liu_ri_palaces":                chart.LiuRiPalaces,
		"query_view":                    svc.BuildQueryView(chart),
	}
	if chart.ContentHash != "" {
		response["content_hash"] = chart.ContentHash
	}
	if chart.DerivedContentHash != "" {
		response["derivation_type"] = chart.DerivationType
		response["derivation_input"] = chart.DerivationInput
		response["derivation_fingerprint"] = chart.DerivationFingerprint
		response["base_content_hash"] = chart.BaseContentHash
		response["derived_content_hash"] = chart.DerivedContentHash
	}
	return response
}

func (h *ZiWeiChartHandler) Calculate(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	var req struct {
		model.ChartRequest
		Algorithm string `json:"algorithm"`
		Profile   string `json:"profile"`
		ChartID   uint   `json:"chart_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	svc := h.Service
	if svc == nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceDisabled, "service not available")
		return
	}
	profile, err := ziwei.ResolveProfile(req.Profile)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Algorithm) != "" {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "algorithm is not a calculation contract; use a registered profile")
		return
	}

	// chart_id provided: check cache or compute and store
	if req.ChartID > 0 && h.Charts != nil {
		userID, _ := c.Get("userID")
		birthChart, err := h.Charts.FindByIDForUser(req.ChartID, userID.(uint))
		if err != nil {
			respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "chart lookup failed")
			return
		}
		if birthChart == nil {
			respondError(c, http.StatusNotFound, ErrCodeNotFound, "chart not found")
			return
		}
		birth, err := resolveStoredZiWeiBirth(birthChart)
		if err != nil {
			respondError(c, http.StatusInternalServerError, ErrCodeServiceError, err.Error())
			return
		}

		if birthChart.ZiWeiComputed && len(birthChart.ZiWeiResult) > 0 {
			// Serve cached result
			var cached ziwei.ZiWeiChart
			if err := json.Unmarshal(birthChart.ZiWeiResult, &cached); err == nil && svc.ChartMatchesInputProfile(
				&cached, profile.ID,
				birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender,
			) {
				if err := svc.AttachBirthData(&cached, birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender); err != nil {
					respondError(c, http.StatusInternalServerError, ErrCodeServiceError, fmt.Sprintf("restore cached chart failed: %v", err))
					return
				}
				respondJSON(c, http.StatusOK, mapChartToResponse(&cached, svc))
				return
			}
			// If unmarshal fails, fall through to recompute
		}

		chart, err := svc.CalculateChartWithProfile(profile.ID, birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender)
		if err != nil {
			respondError(c, http.StatusInternalServerError, ErrCodeServiceError, fmt.Sprintf("chart calculation failed: %v", err))
			return
		}

		// Store result
		if data, err := json.Marshal(chart); err == nil {
			birthChart.ZiWeiResult = datatypes.JSON(data)
			birthChart.ZiWeiComputed = true
			if err := h.Charts.Update(birthChart); err != nil {
				// Log but don't fail the request
				_ = err
			}
		}

		respondJSON(c, http.StatusOK, mapChartToResponse(chart, svc))
		return
	}

	// No chart_id: normalize the complete request before calculating Zi Wei.
	birth, requiresSelection, err := resolveZiWeiRequestBirth(req.ChartRequest)
	if err != nil {
		status := http.StatusBadRequest
		if requiresSelection && strings.TrimSpace(req.CandidateID) == "" {
			status = http.StatusConflict
		}
		respondError(c, status, codeFromStatus(status), err.Error())
		return
	}
	chart, err := svc.CalculateChartWithProfile(profile.ID, birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, fmt.Sprintf("chart calculation failed: %v", err))
		return
	}

	respondJSON(c, http.StatusOK, mapChartToResponse(chart, svc))
}

// RegisterZiWeiRoutes registers the ZiWei chart calculation route.
func RegisterZiWeiRoutes(r gin.IRouter, svc *ziwei.ZiWeiService) {
	h := &ZiWeiChartHandler{Service: svc}
	r.POST("/ziwei/chart", h.Calculate)
}

// RegisterZiWeiRoutesWithStore registers the ZiWei route with a ChartStore for caching.
func RegisterZiWeiRoutesWithStore(r gin.IRouter, svc *ziwei.ZiWeiService, store ChartStore) {
	h := &ZiWeiChartHandler{Service: svc, Charts: store}
	r.POST("/ziwei/chart", h.Calculate)
}
