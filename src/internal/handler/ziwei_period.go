package handler

import (
	"fmt"
	"net/http"
	"time"

	"bazi/internal/model"
	"bazi/internal/service"

	"github.com/gin-gonic/gin"
)

// ZiWeiPeriodHandler handles period (dayun/liunian/liuyue/liuri) and overlay calculations.
type ZiWeiPeriodHandler struct {
	Charts  ChartStore
	Service interface{}
}

// getChart looks up the birth chart and calculates the ZiWeiChart.
func (h *ZiWeiPeriodHandler) getChart(chartID uint) (*service.ZiWeiChart, *model.BirthChart, error) {
	svc, ok := h.Service.(*service.ZiWeiService)
	if !ok || svc == nil {
		return nil, nil, fmt.Errorf("service not available")
	}
	birthChart, err := h.Charts.FindByID(chartID)
	if err != nil {
		return nil, nil, fmt.Errorf("chart lookup failed: %w", err)
	}
	if birthChart == nil {
		return nil, nil, fmt.Errorf("chart not found")
	}
	chart, err := svc.CalculateChart(birthChart.BirthYear, birthChart.BirthMonth, birthChart.BirthDay, birthChart.BirthHour, birthChart.BirthMin, birthChart.Gender)
	if err != nil {
		return nil, nil, fmt.Errorf("chart calculation failed: %w", err)
	}
	return chart, birthChart, nil
}

// Period handles dayun, liunian, liuyue, liuri, and sihua_feixing period calculations.
func (h *ZiWeiPeriodHandler) Period(c *gin.Context) {
	var req struct {
		ChartID    uint   `json:"chart_id"`
		PeriodType string `json:"period_type"`
		Year       int    `json:"year"`
		Month      int    `json:"month"`
		Day        int    `json:"day"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}

	chart, birthChart, err := h.getChart(req.ChartID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	svc, _ := h.Service.(*service.ZiWeiService)

	switch req.PeriodType {
	case "dayun":
		dayun := svc.CalculateDayun(chart)
		c.JSON(http.StatusOK, gin.H{"periods": dayun})

	case "liunian":
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liunian := svc.CalculateLiunian(chart, year)
		c.JSON(http.StatusOK, gin.H{"periods": []gin.H{mapChartToResponse(liunian, birthChart.BirthMonth, birthChart.BirthHour)}})

	case "liuyue":
		month := req.Month
		if month == 0 {
			month = int(time.Now().Month())
		}
		liuyue := svc.CalculateLiuyue(chart, month)
		c.JSON(http.StatusOK, gin.H{"periods": []gin.H{mapChartToResponse(liuyue, birthChart.BirthMonth, birthChart.BirthHour)}})

	case "liuri":
		day := req.Day
		if day == 0 {
			day = time.Now().Day()
		}
		liuri := svc.CalculateLiuri(chart, day)
		c.JSON(http.StatusOK, gin.H{"periods": []gin.H{mapChartToResponse(liuri, birthChart.BirthMonth, birthChart.BirthHour)}})

	case "sihua_feixing":
		flying := svc.AnalyzeFlyingStars(chart)
		c.JSON(http.StatusOK, gin.H{"periods": flying})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown period_type"})
	}
}

// Overlay handles the liunian overlay calculation.
func (h *ZiWeiPeriodHandler) Overlay(c *gin.Context) {
	var req struct {
		ChartID uint `json:"chart_id"`
		Year    int  `json:"year"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid"})
		return
	}

	chart, birthChart, err := h.getChart(req.ChartID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	svc, _ := h.Service.(*service.ZiWeiService)
	liunian := svc.CalculateLiunian(chart, req.Year)
	c.JSON(http.StatusOK, mapChartToResponse(liunian, birthChart.BirthMonth, birthChart.BirthHour))
}

// RegisterZiWeiPeriodRoutes registers ZiWei period and overlay routes.
func RegisterZiWeiPeriodRoutes(r *gin.Engine, svc *service.ZiWeiService, store ChartStore) {
	h := &ZiWeiPeriodHandler{Service: svc, Charts: store}
	r.POST("/api/ziwei/period", h.Period)
	r.POST("/api/ziwei/overlay", h.Overlay)
}
