package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bazi/internal/model"
	"bazi/internal/service/ziwei"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

// ZiWeiPeriodHandler handles period and overlay requests.
type ZiWeiPeriodHandler struct {
	Charts  ChartStore
	Service *ziwei.ZiWeiService
}

type ziWeiPeriodRequest struct {
	ChartID    uint   `json:"chart_id"`
	PeriodType string `json:"period_type"`
	Year       int    `json:"year"`
	Month      int    `json:"month"`
	Day        int    `json:"day"`
	PalaceIdx  int    `json:"palace_idx"`
	ChartID2   uint   `json:"chart_id2"`
}

// getChart loads a user's birth chart and resolves the cached ZiWei result.
func (h *ZiWeiPeriodHandler) getChart(chartID, userID uint) (*ziwei.ZiWeiChart, *model.BirthChart, error) {
	if h.Service == nil {
		return nil, nil, fmt.Errorf("service not available")
	}
	birthChart, err := h.Charts.FindByIDForUser(chartID, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("chart lookup failed: %w", err)
	}
	if birthChart == nil {
		return nil, nil, fmt.Errorf("chart not found")
	}
	birth, err := resolveStoredZiWeiBirth(birthChart)
	if err != nil {
		return nil, nil, err
	}
	if birthChart.ZiWeiComputed && len(birthChart.ZiWeiResult) > 0 {
		var cached ziwei.ZiWeiChart
		if err := json.Unmarshal(birthChart.ZiWeiResult, &cached); err == nil && h.Service.ChartMatchesInputProfile(
			&cached, "",
			birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender,
		) {
			if err := h.Service.AttachBirthData(&cached, birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender); err != nil {
				return nil, nil, fmt.Errorf("attach birth data: %w", err)
			}
			return &cached, birthChart, nil
		}
	}

	chart, err := h.Service.CalculateChart(birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender)
	if err != nil {
		return nil, nil, fmt.Errorf("chart calculation failed: %w", err)
	}
	if data, err := json.Marshal(chart); err == nil {
		birthChart.ZiWeiResult = datatypes.JSON(data)
		birthChart.ZiWeiComputed = true
		_ = h.Charts.Update(birthChart)
	}
	return chart, birthChart, nil
}

// Period validates the HTTP contract and delegates each period type to a
// focused response assembler.
func (h *ZiWeiPeriodHandler) Period(c *gin.Context) {
	identity, exists := c.Get("userID")
	userID, validIdentity := identity.(uint)
	if !exists || !validIdentity {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	var req ziWeiPeriodRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}
	chart, _, err := h.getChart(req.ChartID, userID)
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, err.Error())
		return
	}
	h.dispatchPeriod(c, chart, userID, req)
}

// Overlay handles the liunian overlay calculation.
func (h *ZiWeiPeriodHandler) Overlay(c *gin.Context) {
	identity, exists := c.Get("userID")
	userID, validIdentity := identity.(uint)
	if !exists || !validIdentity {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ChartID uint `json:"chart_id"`
		Year    int  `json:"year"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}
	chart, _, err := h.getChart(req.ChartID, userID)
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, err.Error())
		return
	}

	year := req.Year
	if year == 0 {
		year = ziwei.CurrentLunarYearLabel(time.Now())
	}
	liunian := h.Service.CalculateLiunian(chart, year)
	if liunian == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian year")
		return
	}
	result := mapChartToResponse(liunian, h.Service)
	result["year"] = year
	result["liu_nian_stars"] = liunian.LiuNianStars
	result["overlay_analysis"] = h.Service.AnalyzeLiunianOverlay(chart, liunian, year)
	respondJSON(c, http.StatusOK, result)
}

// RegisterZiWeiPeriodRoutes registers ZiWei period and overlay routes.
func RegisterZiWeiPeriodRoutes(r gin.IRouter, svc *ziwei.ZiWeiService, store ChartStore) {
	h := &ZiWeiPeriodHandler{Service: svc, Charts: store}
	r.POST("/ziwei/period", h.Period)
	r.POST("/ziwei/overlay", h.Overlay)
}
