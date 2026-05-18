package handler

import (
	"fmt"
	"net/http"
	"bazi/internal/model"
	"bazi/internal/service"
	"github.com/gin-gonic/gin"
)

type ZiWeiChartHandler struct {
	Service interface{}
}

// StarInfoResponse maps a star name + brightness for frontend consumption.
type StarInfoResponse struct {
	Name       string `json:"name"`
	Brightness string `json:"brightness"`
}

// ZiWeiPalaceResponse is the frontend-compatible palace format.
type ZiWeiPalaceResponse struct {
	Name      string            `json:"name"`
	Branch    string            `json:"branch"`
	MainStars []StarInfoResponse `json:"main_stars"`
	AuxStars  []StarInfoResponse `json:"aux_stars"`
	Sihua     []string          `json:"sihua"`
}

func mapPalaceToResponse(p *service.PalaceInfo, branch string) ZiWeiPalaceResponse {
	resp := ZiWeiPalaceResponse{
		Name:   p.Name,
		Branch: branch,
		Sihua:  p.FourHua,
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
		palaces[i] = mapPalaceToResponse(&chart.Palaces[i], branch)
	}

	return gin.H{
		"palaces":    palaces,
		"mingZhu":    chart.LifeMaster,
		"shenZhu":    chart.BodyMaster,
		"bodyPalace": chart.BodyPalace,
		"wuxingJu":   chart.FiveBureau,
		"patterns":   chart.Patterns,
	}
}

func (h *ZiWeiChartHandler) Calculate(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req model.ChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}

	svc, ok := h.Service.(*service.ZiWeiService)
	if !ok || svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		return
	}

	chart, err := svc.CalculateChart(req.BirthYear, req.BirthMonth, req.BirthDay, req.BirthHour, req.BirthMin, req.Gender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("chart calculation failed: %v", err)})
		return
	}

	// Use lunar month from chart's internal birth info, fallback to solar month
	lunarMonth := req.BirthMonth
	c.JSON(http.StatusOK, mapChartToResponse(chart, lunarMonth, req.BirthHour))
}

// RegisterZiWeiRoutes registers the ZiWei chart calculation route.
func RegisterZiWeiRoutes(r gin.IRouter, svc *service.ZiWeiService) {
	h := &ZiWeiChartHandler{Service: svc}
	r.POST("/ziwei/chart", h.Calculate)
}
