package handler

import (
	"bazi/internal/model"
	"bazi/internal/service/ziwei"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"net/http"
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

func mapChartToResponse(chart *ziwei.ZiWeiChart) gin.H {
	palaces := make([]ZiWeiPalaceResponse, 12)
	for i := 0; i < 12; i++ {
		sf := &ziwei.SanfangSizhengResult{
			Opposite: chart.SanfangSizheng[i].Opposite,
			Trine1:   chart.SanfangSizheng[i].Trine1,
			Trine2:   chart.SanfangSizheng[i].Trine2,
		}
		palaces[i] = mapPalaceToResponse(&chart.Palaces[i], sf)
	}

	return gin.H{
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
		"query_view":                    ziwei.BuildQueryView(chart),
	}
}

func (h *ZiWeiChartHandler) Calculate(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	var req struct {
		model.ChartRequest
		Algorithm string `json:"algorithm"` // "default" or "zhongzhou"
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

	switch req.Algorithm {
	case "zhongzhou":
		svc.SetAlgorithm(ziwei.AlgorithmZhongZhou)
	default:
		svc.SetAlgorithm(ziwei.AlgorithmFullBook)
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

		if birthChart.ZiWeiComputed && len(birthChart.ZiWeiResult) > 0 {
			// Serve cached result
			var cached ziwei.ZiWeiChart
			if err := json.Unmarshal(birthChart.ZiWeiResult, &cached); err == nil {
				if err := svc.AttachBirthData(&cached, birthChart.BirthYear, birthChart.BirthMonth, birthChart.BirthDay, birthChart.BirthHour, birthChart.BirthMin, birthChart.Gender); err != nil {
					respondError(c, http.StatusInternalServerError, ErrCodeServiceError, fmt.Sprintf("restore cached chart failed: %v", err))
					return
				}
				respondJSON(c, http.StatusOK, mapChartToResponse(&cached))
				return
			}
			// If unmarshal fails, fall through to recompute
		}

		chart, err := svc.CalculateChart(birthChart.BirthYear, birthChart.BirthMonth, birthChart.BirthDay, birthChart.BirthHour, birthChart.BirthMin, birthChart.Gender)
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

		respondJSON(c, http.StatusOK, mapChartToResponse(chart))
		return
	}

	// No chart_id: compute from raw birth data (original behavior)
	chart, err := svc.CalculateChart(req.BirthYear, req.BirthMonth, req.BirthDay, req.BirthHour, req.BirthMin, req.Gender)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, fmt.Sprintf("chart calculation failed: %v", err))
		return
	}

	respondJSON(c, http.StatusOK, mapChartToResponse(chart))
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
