package handler

import (
	"net/http"
	"time"

	"bazi/internal/model"
	"bazi/internal/service/bazi"
	fortunePkg "bazi/internal/service/fortune"

	"github.com/gin-gonic/gin"
)

type FortuneHandler struct {
	Engine     *fortunePkg.FortuneEngine
	ChartStore ChartStore
}

// CalculateDaily handles POST /api/fortune.
// It requires JWT authentication via AuthMiddleware.
func (h *FortuneHandler) CalculateDaily(c *gin.Context) {
	var req model.FortuneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	if req.ChartID == 0 || req.QueryDate == "" {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "chart_id and query_date are required")
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	chart, err := h.ChartStore.FindByIDForUser(req.ChartID, userID.(uint))
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

	queryDate, err := time.ParseInLocation("2006-01-02", req.QueryDate, time.FixedZone("CST", 8*3600))
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid query_date format, expected YYYY-MM-DD")
		return
	}

	fortune := h.Engine.CalculateDaily(baziResult, queryDate, resolved.BirthYear)

	resp := model.FortuneResponse{
		EngineVersion:        fortunePkg.FortuneEngineVersion,
		BaziEngineVersion:    resolved.EngineVersion,
		BaziResolutionSource: resolved.Source,
		RuleVersion:          baziResult.RuleVersion,
		School:               baziResult.School,
		RuleMeta:             baziResult.RuleMeta,
		SolarDate:            fortune.Date,
		LunarDate:            fortune.LunarDate,
		DayGanZhi:            fortune.DayPillar.Gan + fortune.DayPillar.Zhi,
		WeekDay:              fortune.WeekDay,
		ShengXiao:            fortune.ShengXiao,
		JiShen:               fortune.JiShen,
		XiongShen:            fortune.XiongShen,
		TaiShen:              fortune.TaiShen,
		WuXing:               fortune.WuXing,
		PengZu:               fortune.PengZu,
		Gua:                  fortune.Gua,
		JieQi:                fortune.JieQi,
		ElementImages:        fortune.ElementImages,
		Score:                fortune.Score,
		ScoreBreakdown:       fortune.ScoreBreakdown,
		EvidenceCompleteness: fortune.ScoreBreakdown.EvidenceCompleteness,
		SupportingEvidence:   fortune.ScoreBreakdown.SupportingEvidence,
		CounterEvidence:      fortune.ScoreBreakdown.CounterEvidence,
		ClashZodiac:          fortune.ClashZodiac,
		TodayElements:        fortune.TodayElements,
		SeasonElement:        fortune.SeasonElement,
		ShengKeAnalysis:      model.ShengKeAnalysis(fortune.ShengKe),
		FortuneLayers:        fortune.Layers,
	}
	// 日课推算结果
	if rikuyo := fortune.Rikuyo; rikuyo != nil {
		resp.TenGod = rikuyo.TenGod
		resp.TwelveStage = rikuyo.TwelveStage
		resp.JianChu = rikuyo.JianChu
		resp.HuangDao = rikuyo.HuangDao
		resp.HiddenStems = rikuyo.HiddenStems
		resp.StemRelations = rikuyo.StemRelations
		resp.BranchRelations = rikuyo.BranchRelations
		resp.ActivatedShenSha = rikuyo.ActivatedShenSha
		resp.SeasonalState = rikuyo.SeasonalState
		resp.FortuneLayers = fortune.Layers
	}
	respondJSON(c, http.StatusOK, resp)
}
