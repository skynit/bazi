package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service/bazi"
	"bazi/internal/service/fortune"

	"github.com/gin-gonic/gin"
)

// ChartStore defines an interface for looking up birth charts.
type ChartStore interface {
	FindByID(id uint) (*model.BirthChart, error)
	FindByIDForUser(id uint, userID uint) (*model.BirthChart, error)
	Update(chart *model.BirthChart) error
}

// WeeklyFortuneHandler handles weekly fortune endpoints.
type WeeklyFortuneHandler struct {
	Engine *fortune.FortuneEngine
	Charts ChartStore
}

// Weekly handles POST /api/fortune/weekly.
func (h *WeeklyFortuneHandler) Weekly(c *gin.Context) {
	var req model.WeeklyFortuneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	chart, err := h.Charts.FindByIDForUser(req.ChartID, userID.(uint))
	if err != nil || chart == nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, "chart not found")
		return
	}

	gender := normalizeGender(chart.Gender)

	baziSvc := bazi.BaziService{}
	baziResult, err := baziSvc.Calculate(
		chart.BirthYear, chart.BirthMonth, chart.BirthDay,
		chart.BirthHour, chart.BirthMin, gender,
	)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to compute chart: "+err.Error())
		return
	}

	startDate, err := time.ParseInLocation("2006-01-02", req.StartDate, time.FixedZone("CST", 8*3600))
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid start_date format, use YYYY-MM-DD")
		return
	}

	result := h.Engine.CalculateWeekly(baziResult, startDate, chart.BirthYear)

	dailyFortunes := make([]model.FortuneResponse, len(result.DailyFortunes))
	for i, df := range result.DailyFortunes {
		dailyFortunes[i] = dailyFortuneToResponse(df)
	}

	trendJSON, _ := json.Marshal(result.ElementTrend)

	respondJSON(c, http.StatusOK, model.WeeklyFortuneResponse{
		DailyFortunes: dailyFortunes,
		WeeklyScore:   result.WeeklyScore,
		ElementTrend:  string(trendJSON),
		Summary:       result.Summary,
	})
}

func normalizeGender(g string) string {
	s := strings.TrimSpace(g)
	switch {
	case s == "男" || strings.EqualFold(s, "male") || strings.EqualFold(s, "m"):
		return model.GenderMale
	case s == "女" || strings.EqualFold(s, "female") || strings.EqualFold(s, "f"):
		return model.GenderFemale
	default:
		return model.GenderMale
	}
}

func dailyFortuneToResponse(df fortune.DailyFortune) model.FortuneResponse {
	yiJi := buildYiJiString(df.Yi, df.Ji)

	yiItems := make([]string, len(df.Yi))
	for i, item := range df.Yi {
		yiItems[i] = item.Activity
	}
	jiItems := make([]string, len(df.Ji))
	for i, item := range df.Ji {
		jiItems[i] = item.Activity
	}

	luckyNum := 0
	if len(df.LuckyNumbers) > 0 {
		luckyNum = df.LuckyNumbers[0]
	}

	resp := model.FortuneResponse{
		RuleVersion:         df.Layers.RuleVersion,
		School:              df.Layers.School,
		SolarDate:           df.Date,
		DayGanZhi:           df.DayPillar.Gan + df.DayPillar.Zhi,
		YiJi:                yiJi,
		ElementImages:       df.ElementImages,
		Score:               df.Score,
		LuckyColor:          df.LuckyColor,
		LuckyNumber:         luckyNum,
		WealthDir:           df.WealthDir,
		Guide:               df.Guide,
		ClashZodiac:         df.ClashZodiac,
		AuspiciousHours:     df.AuspiciousHours,
		YiItems:             yiItems,
		JiItems:             jiItems,
		TodayElements:       df.TodayElements,
		SeasonElementAdvice: df.SeasonElementAdvice,
		FlowImpact:          df.FlowImpact,
		FortuneLayers:       df.Layers,
	}
	if rikuyo := df.Rikuyo; rikuyo != nil {
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
		resp.FortuneLayers = df.Layers
		resp.OverallVerdict = rikuyo.OverallVerdict
		resp.FavorScore = rikuyo.FavorScore
		resp.PatternName = rikuyo.PatternName
		resp.PatternType = rikuyo.PatternType
		resp.PatternFavorable = rikuyo.PatternFavorable
		resp.PatternUnfavorable = rikuyo.PatternUnfavorable
	}
	return resp
}

func buildYiJiString(yi, ji []model.YiJiItem) string {
	parts := make([]string, 0, len(yi)+len(ji))
	for _, item := range yi {
		parts = append(parts, "宜"+item.Activity)
	}
	for _, item := range ji {
		parts = append(parts, "忌"+item.Activity)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

// RegisterFortuneRoutes registers fortune routes requiring JWT.
func RegisterFortuneRoutes(router *gin.Engine, engine *fortune.FortuneEngine, charts ChartStore) {
	h := &WeeklyFortuneHandler{Engine: engine, Charts: charts}

	fortune := router.Group("/api/fortune")
	fortune.Use(middleware.AuthMiddleware())
	{
		fortune.POST("/weekly", h.Weekly)
	}
}
