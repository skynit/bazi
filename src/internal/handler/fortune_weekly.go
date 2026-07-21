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

	baziSvc := bazi.BaziService{}
	resolved, err := resolveChartBazi(&baziSvc, chart)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to compute chart: "+err.Error())
		return
	}
	baziResult := resolved.Result

	startDate, err := time.ParseInLocation("2006-01-02", req.StartDate, time.FixedZone("CST", 8*3600))
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid start_date format, use YYYY-MM-DD")
		return
	}

	result := h.Engine.CalculateWeekly(baziResult, startDate, resolved.BirthYear)

	dailyFortunes := make([]model.FortuneResponse, len(result.DailyFortunes))
	for i, df := range result.DailyFortunes {
		dailyFortunes[i] = dailyFortuneToResponse(df, resolved)
	}

	trendJSON, _ := json.Marshal(result.ElementTrend)

	respondJSON(c, http.StatusOK, model.WeeklyFortuneResponse{
		DailyFortunes:           dailyFortunes,
		StructuralRelationIndex: result.StructuralRelationIndex,
		ElementTrend:            string(trendJSON),
		Summary:                 result.Summary,
	})
}

func normalizeGender(g string) string {
	gender, ok := parseGender(g)
	if ok {
		return gender
	}
	return model.GenderMale
}

func parseGender(g string) (string, bool) {
	s := strings.TrimSpace(g)
	switch {
	case s == "男" || strings.EqualFold(s, "male") || strings.EqualFold(s, "m"):
		return model.GenderMale, true
	case s == "女" || strings.EqualFold(s, "female") || strings.EqualFold(s, "f"):
		return model.GenderFemale, true
	default:
		return "", false
	}
}

func dailyFortuneToResponse(df fortune.DailyFortune, resolved *resolvedChartBazi) model.FortuneResponse {
	resp := model.FortuneResponse{
		EngineVersion:        fortune.FortuneEngineVersion,
		BaziEngineVersion:    resolved.EngineVersion,
		BaziResolutionSource: resolved.Source,
		RuleVersion:          df.Layers.RuleVersion,
		School:               df.Layers.School,
		RuleMeta:             bazi.DefaultRuleMeta(),
		SolarDate:            df.Date,
		DayGanZhi:            df.DayPillar.Gan + df.DayPillar.Zhi,
		ElementImages:        df.ElementImages,
		Score:                df.Score,
		ScoreBreakdown:       df.ScoreBreakdown,
		EvidenceCompleteness: df.ScoreBreakdown.EvidenceCompleteness,
		SupportingEvidence:   df.ScoreBreakdown.SupportingEvidence,
		CounterEvidence:      df.ScoreBreakdown.CounterEvidence,
		ClashZodiac:          df.ClashZodiac,
		TodayElements:        df.TodayElements,
		SeasonElement:        df.SeasonElement,
		ShengKeAnalysis:      model.ShengKeAnalysis(df.ShengKe),
		FortuneLayers:        df.Layers,
	}
	if rikuyo := df.Rikuyo; rikuyo != nil {
		resp.TenGod = rikuyo.TenGod
		resp.TwelveStage = rikuyo.TwelveStage
		resp.JianChu = rikuyo.JianChu
		resp.HuangDao = rikuyo.HuangDao
		resp.HiddenStems = rikuyo.HiddenStems
		resp.StemRelations = rikuyo.StemRelations
		resp.BranchRelations = rikuyo.BranchRelations
		resp.ActivatedShenSha = rikuyo.ActivatedShenSha
		resp.SeasonalState = rikuyo.SeasonalState
		resp.FortuneLayers = df.Layers
	}
	return resp
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
