package handler

import (
	"bazi/internal/model"
	"bazi/internal/service"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"net/http"
)

type ZiWeiChartHandler struct {
	Service interface{}
	Charts  ChartStore
}

// StarInfoResponse maps a star name + brightness for frontend consumption.
type StarInfoResponse struct {
	Name       string `json:"name"`
	Brightness string `json:"brightness"`
}

// ZiWeiPalaceResponse is the frontend-compatible palace format.
type ZiWeiPalaceResponse struct {
	Name            string             `json:"name"`
	Branch          string             `json:"branch"`
	MainStars       []StarInfoResponse `json:"mainStars"`
	AuxStars        []StarInfoResponse `json:"auxStars"`
	Sihua           []string           `json:"sihua"`
	SanfangSizheng  *SanfangResponse   `json:"sanfang_sizheng,omitempty"`
	AdjectiveStars  []StarInfoResponse `json:"adjective_stars,omitempty"`
	Changsheng12    string             `json:"changsheng_12,omitempty"`
	Boshi12         string             `json:"boshi_12,omitempty"`
	JiangQian12     string             `json:"jiang_qian_12,omitempty"`
	SuiQian12       string             `json:"sui_qian_12,omitempty"`
}

// SanfangResponse holds the sanfang sizheng data for a palace.
type SanfangResponse struct {
	Opposite string `json:"opposite"`
	Trine1   string `json:"trine1"`
	Trine2   string `json:"trine2"`
}

func mapPalaceToResponse(p *service.PalaceInfo, branch string, sf *service.SanfangSizhengResult) ZiWeiPalaceResponse {
	resp := ZiWeiPalaceResponse{
		Name:            p.Name,
		Branch:          branch,
		Sihua:           p.FourHua,
		MainStars:       []StarInfoResponse{},
		AuxStars:        []StarInfoResponse{},
		AdjectiveStars:  []StarInfoResponse{},
	}
	if sf != nil {
		resp.SanfangSizheng = &SanfangResponse{
			Opposite: sf.Opposite,
			Trine1:   sf.Trine1,
			Trine2:   sf.Trine2,
		}
	}
	for _, star := range p.MainStars {
		b := ""
		if p.Brightness != nil {
			b = p.Brightness[star]
		}
		resp.MainStars = append(resp.MainStars, StarInfoResponse{Name: star, Brightness: b})
	}
	for _, star := range p.AuxStars {
		b := ""
		if p.Brightness != nil {
			b = p.Brightness[star]
		}
		resp.AuxStars = append(resp.AuxStars, StarInfoResponse{Name: star, Brightness: b})
	}
	for _, star := range p.AdjectiveStars {
		resp.AdjectiveStars = append(resp.AdjectiveStars, StarInfoResponse{Name: star, Brightness: ""})
	}
	resp.Changsheng12 = p.Changsheng12
	resp.Boshi12 = p.Boshi12
	resp.JiangQian12 = p.JiangQian12
	resp.SuiQian12 = p.SuiQian12
	return resp
}

// computeMingGongBranch calculates which branch 命宫 occupies using the standard ZiWei formula.
// branchOrder: ["寅","卯","辰","巳","午","未","申","酉","戌","亥","子","丑"]
func computeMingGongBranch(lunarMonth, hour int) string {
	branchOrder := []string{"寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑"}
	// hourBranchIndex: 子(23-1)=0, 丑(1-3)=1, ..., 亥(21-23)=11
	hourBranchIndex := ((hour + 1) / 2) % 12
	// Standard 安命宫: 寅宫起正月顺数至生月, 再从该宫起子时逆数至生时
	monthOffset := (lunarMonth - 1) % 12
	mingGongIdx := (monthOffset - hourBranchIndex + 12) % 12
	return branchOrder[mingGongIdx]
}

func mapChartToResponse(chart *service.ZiWeiChart, lunarMonth, hour int) gin.H {
	branchOrder := []string{"寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑"}
	mingGongBranch := computeMingGongBranch(lunarMonth, hour)
	mingGongIdx := 0
	for i, b := range branchOrder {
		if b == mingGongBranch {
			mingGongIdx = i
			break
		}
	}

	palaces := make([]ZiWeiPalaceResponse, 12)
	for i := 0; i < 12; i++ {
		branch := branchOrder[(mingGongIdx+i)%12]
		sf := &service.SanfangSizhengResult{
			Opposite: chart.SanfangSizheng[i].Opposite,
			Trine1:   chart.SanfangSizheng[i].Trine1,
			Trine2:   chart.SanfangSizheng[i].Trine2,
		}
		palaces[i] = mapPalaceToResponse(&chart.Palaces[i], branch, sf)
	}

	return gin.H{
		"palaces":         palaces,
		"mingZhu":         chart.LifeMaster,
		"shenZhu":         chart.BodyMaster,
		"bodyPalace":      chart.BodyPalace,
		"wuxingJu":        chart.FiveBureau,
		"patterns":        chart.Patterns,
		"liu_nian_stars":  chart.LiuNianStars,
		"liu_yue_stars":   chart.LiuYueStars,
		"liu_ri_stars":    chart.LiuRiStars,
	}
}

func (h *ZiWeiChartHandler) Calculate(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req struct {
		model.ChartRequest
		Algorithm string `json:"algorithm"` // "default" or "zhongzhou"
		ChartID   uint   `json:"chart_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	svc, ok := h.Service.(*service.ZiWeiService)
	if !ok || svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		return
	}

	switch req.Algorithm {
	case "zhongzhou":
		svc.SetAlgorithm(service.AlgorithmZhongZhou)
	default:
		svc.SetAlgorithm(service.AlgorithmFullBook)
	}

	// chart_id provided: check cache or compute and store
	if req.ChartID > 0 && h.Charts != nil {
		birthChart, err := h.Charts.FindByID(req.ChartID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "chart lookup failed"})
			return
		}
		if birthChart == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
			return
		}

		if birthChart.ZiWeiComputed && len(birthChart.ZiWeiResult) > 0 {
			// Serve cached result
			var cached service.ZiWeiChart
			if err := json.Unmarshal(birthChart.ZiWeiResult, &cached); err == nil {
				lunarMonth := cached.LunarMonth
				if lunarMonth == 0 {
					lunarMonth = birthChart.BirthMonth
				}
				c.JSON(http.StatusOK, mapChartToResponse(&cached, lunarMonth, birthChart.BirthHour))
				return
			}
			// If unmarshal fails, fall through to recompute
		}

		chart, err := svc.CalculateChart(birthChart.BirthYear, birthChart.BirthMonth, birthChart.BirthDay, birthChart.BirthHour, birthChart.BirthMin, birthChart.Gender)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("chart calculation failed: %v", err)})
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

		lunarMonth := chart.LunarMonth
		if lunarMonth == 0 {
			lunarMonth = birthChart.BirthMonth
		}
		c.JSON(http.StatusOK, mapChartToResponse(chart, lunarMonth, birthChart.BirthHour))
		return
	}

	// No chart_id: compute from raw birth data (original behavior)
	chart, err := svc.CalculateChart(req.BirthYear, req.BirthMonth, req.BirthDay, req.BirthHour, req.BirthMin, req.Gender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("chart calculation failed: %v", err)})
		return
	}

	// Use lunar month from the calculated chart
	lunarMonth := chart.LunarMonth
	if lunarMonth == 0 {
		lunarMonth = req.BirthMonth
	}
	c.JSON(http.StatusOK, mapChartToResponse(chart, lunarMonth, req.BirthHour))
}

// RegisterZiWeiRoutes registers the ZiWei chart calculation route.
func RegisterZiWeiRoutes(r gin.IRouter, svc *service.ZiWeiService) {
	h := &ZiWeiChartHandler{Service: svc}
	r.POST("/ziwei/chart", h.Calculate)
}

// RegisterZiWeiRoutesWithStore registers the ZiWei route with a ChartStore for caching.
func RegisterZiWeiRoutesWithStore(r gin.IRouter, svc *service.ZiWeiService, store ChartStore) {
	h := &ZiWeiChartHandler{Service: svc, Charts: store}
	r.POST("/ziwei/chart", h.Calculate)
}
