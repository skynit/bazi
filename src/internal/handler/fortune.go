package handler

import (
	"net/http"
	"time"

	"bazi/internal/model"
	"bazi/internal/service"

	"github.com/gin-gonic/gin"
)

// FortuneHandler handles fortune-telling endpoints.
type FortuneHandler struct {
	Engine     *service.FortuneEngine
	ChartStore ChartStore
}

// CalculateDaily handles POST /api/fortune.
// It requires JWT authentication via AuthMiddleware.
func (h *FortuneHandler) CalculateDaily(c *gin.Context) {
	var req model.FortuneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if req.ChartID == 0 || req.QueryDate == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "chart_id and query_date are required"})
		return
	}

	chart, err := h.ChartStore.FindByID(req.ChartID)
	if err != nil || chart == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chart not found"})
		return
	}

	gender := normalizeGender(chart.Gender)
	baziSvc := service.BaziService{}
	baziResult, err := baziSvc.Calculate(
		chart.BirthYear, chart.BirthMonth, chart.BirthDay,
		chart.BirthHour, chart.BirthMin, gender,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute chart: " + err.Error()})
		return
	}

	queryDate, err := time.Parse("2006-01-02", req.QueryDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query_date format, expected YYYY-MM-DD"})
		return
	}

	fortune := h.Engine.CalculateDaily(baziResult, queryDate)

	// Map daily fortune to response
	yiItems := make([]string, len(fortune.Yi))
	for i, item := range fortune.Yi {
		yiItems[i] = item.Activity
	}
	jiItems := make([]string, len(fortune.Ji))
	for i, item := range fortune.Ji {
		jiItems[i] = item.Activity
	}

	luckyNum := 0
	if len(fortune.LuckyNumbers) > 0 {
		luckyNum = fortune.LuckyNumbers[0]
	}
	resp := model.FortuneResponse{
		SolarDate:       fortune.Date,
		DayGanZhi:       fortune.DayPillar.Gan + fortune.DayPillar.Zhi,
		ElementImages:   fortune.ElementImages,
		Score:           fortune.Score,
		LuckyColor:      fortune.LuckyColor,
		LuckyNumber:     luckyNum,
		WealthDir:       fortune.WealthDir,
		ClashZodiac:     fortune.ClashZodiac,
		AuspiciousHours: fortune.AuspiciousHours,
		YiItems:         yiItems,
		JiItems:         jiItems,
		TodayElements:   fortune.TodayElements,
		TiaoHou:         service.TiaoHou[fortune.DayPillar.Gan+fortune.DayPillar.Zhi],
	}
	// Generate detailed analysis
	analysis := service.AnalyzeDailyFortune(baziResult, fortune.DayPillar.Gan, fortune.DayPillar.Zhi)
	resp.Analysis = analysis
	resp.Score = analysis.Overall.Score   // use AI score, not basic calcScore

	c.JSON(http.StatusOK, resp)
}
