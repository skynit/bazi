package handler

import (
	"net/http"
	"time"

	"bazi/internal/model"
	"bazi/internal/service/bazi"
	"bazi/internal/service/data"
	fortunePkg "bazi/internal/service/fortune"

	"github.com/gin-gonic/gin"
)

// FortuneHandler handles fortune-telling endpoints.
type FortuneHandler struct {
	Engine     *fortunePkg.FortuneEngine
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
	baziSvc := bazi.BaziService{}
	baziResult, err := baziSvc.Calculate(
		chart.BirthYear, chart.BirthMonth, chart.BirthDay,
		chart.BirthHour, chart.BirthMin, gender,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to compute chart: " + err.Error()})
		return
	}

	queryDate, err := time.ParseInLocation("2006-01-02", req.QueryDate, time.FixedZone("CST", 8*3600))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query_date format, expected YYYY-MM-DD"})
		return
	}

	fortune := h.Engine.CalculateDaily(baziResult, queryDate, chart.BirthYear)

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
		TiaoHou:             data.TiaoHou[baziResult.DayPillar.Gan+baziResult.MonthPillar.Zhi],
		SeasonElementAdvice: fortune.SeasonElementAdvice,
		FlowImpact:          fortune.FlowImpact,
	}
	// 日课推算结果
	if rikuyo := fortune.Rikuyo; rikuyo != nil {
		resp.TodayTenGod = rikuyo.TodayTenGod
		resp.TenGodFavorable = rikuyo.TenGodFavorable
		resp.TenGodDesc = rikuyo.TenGodDesc
		resp.TwelveStage = rikuyo.TwelveStage
		resp.StageFavorable = rikuyo.StageFavorable
		resp.StageDesc = rikuyo.StageDesc
		resp.StageFlexible = rikuyo.StageFlexible
		resp.HiddenStems = rikuyo.HiddenStems
		resp.StemRelations = rikuyo.StemRelations
		resp.BranchRelations = rikuyo.BranchRelations
		resp.ActivatedShenSha = rikuyo.ActivatedShenSha
		resp.DaYunInfluence = rikuyo.DaYunInfluence
		resp.LiuNianInfluence = rikuyo.LiuNianInfluence
		resp.AdvanceRetreat = rikuyo.AdvanceRetreat
		resp.YongShenImpact = rikuyo.YongShenImpact
		resp.OverallVerdict = rikuyo.OverallVerdict
		resp.FavorScore = rikuyo.FavorScore
		resp.PatternName = rikuyo.PatternName
		resp.PatternType = rikuyo.PatternType
		resp.PatternFavorable = rikuyo.PatternFavorable
		resp.PatternUnfavorable = rikuyo.PatternUnfavorable
	}
	// Generate detailed analysis
	analysis := fortunePkg.AnalyzeDailyFortune(baziResult, fortune.DayPillar.Gan, fortune.DayPillar.Zhi)
	resp.Analysis = analysis
	resp.Score = analysis.Overall.Score   // use AI score, not basic calcScore

	c.JSON(http.StatusOK, resp)
}
