package handler

import (
	"encoding/json"
	"net/http"

	"bazi/internal/model"
	"bazi/internal/service/bazi"
	"bazi/internal/service/fortune"

	"github.com/gin-gonic/gin"
)

// MonthlyChartStore is the interface for looking up birth charts
// needed for monthly fortune calculations.
type MonthlyChartStore interface {
	FindByID(id uint) (*model.BirthChart, error)
	FindByIDForUser(id uint, userID uint) (*model.BirthChart, error)
}

// MonthlyFortuneHandler handles monthly fortune endpoints.
type MonthlyFortuneHandler struct {
	ChartStore MonthlyChartStore
	Engine     *fortune.FortuneEngine
}

// HandleMonthly processes POST /api/fortune/monthly.
// Requires JWT authentication.
func (h *MonthlyFortuneHandler) HandleMonthly(c *gin.Context) {
	var req model.MonthlyFortuneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	if req.Year < 1900 || req.Year > 2100 || req.Month < 1 || req.Month > 12 {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "year must be 1900-2100, month must be 1-12")
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

	baziSvc := &bazi.BaziService{}
	resolved, err := resolveChartBazi(baziSvc, chart)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to calculate birth chart")
		return
	}
	baziResult := resolved.Result

	monthlyFortune := h.Engine.CalculateMonthly(baziResult, req.Year, req.Month, resolved.BirthYear)
	resp := mapMonthlyFortuneToResponse(monthlyFortune, resolved)

	respondJSON(c, http.StatusOK, resp)
}

// mapMonthlyFortuneToResponse converts a fortune.MonthlyFortune
// to the API DTO model.MonthlyFortuneResponse.
func mapMonthlyFortuneToResponse(mf *fortune.MonthlyFortune, resolved *resolvedChartBazi) model.MonthlyFortuneResponse {
	dailyFortunes := make([]model.FortuneResponse, len(mf.DailyFortunes))
	for i, df := range mf.DailyFortunes {
		dailyFortunes[i] = dailyFortuneToResponse(df, resolved)
	}

	trendJSON, _ := json.Marshal(mf.ElementTrend)

	return model.MonthlyFortuneResponse{
		DailyFortunes:           dailyFortunes,
		StructuralRelationIndex: mf.StructuralRelationIndex,
		ElementTrend:            string(trendJSON),
		Summary:                 mf.Summary,
	}
}

// dailyFortuneToResponse is defined in fortune_weekly.go — shared helper across weekly/monthly handlers.
