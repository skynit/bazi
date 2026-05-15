package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bazi/internal/middleware"
	"bazi/internal/model"
	"bazi/internal/service"

	"github.com/gin-gonic/gin"
)

// ChartStore defines an interface for looking up birth charts.
type ChartStore interface {
	FindByID(id uint) (*model.BirthChart, error)
}

// WeeklyFortuneHandler handles weekly fortune endpoints.
type WeeklyFortuneHandler struct {
	Engine *service.FortuneEngine
	Charts ChartStore
}

// Weekly handles POST /api/fortune/weekly.
func (h *WeeklyFortuneHandler) Weekly(c *gin.Context) {
	var req model.WeeklyFortuneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	chart, err := h.Charts.FindByID(req.ChartID)
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

	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start_date format, use YYYY-MM-DD"})
		return
	}

	result := h.Engine.CalculateWeekly(baziResult, startDate)

	dailyFortunes := make([]model.FortuneResponse, len(result.DailyFortunes))
	for i, df := range result.DailyFortunes {
		dailyFortunes[i] = dailyFortuneToResponse(df)
	}

	trendJSON, _ := json.Marshal(result.ElementTrend)

	c.JSON(http.StatusOK, model.WeeklyFortuneResponse{
		DailyFortunes: dailyFortunes,
		WeeklyScore:   result.WeeklyScore,
		ElementTrend:  string(trendJSON),
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

func dailyFortuneToResponse(df service.DailyFortune) model.FortuneResponse {
	yiJi := buildYiJiString(df.Yi, df.Ji)
	return model.FortuneResponse{
		SolarDate:     df.Date,
		DayGanZhi:     df.DayPillar.Gan + df.DayPillar.Zhi,
		YiJi:          yiJi,
		ElementImages: df.ElementImages,
	}
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
func RegisterFortuneRoutes(router *gin.Engine, engine *service.FortuneEngine, charts ChartStore) {
	h := &WeeklyFortuneHandler{Engine: engine, Charts: charts}

	fortune := router.Group("/api/fortune")
	fortune.Use(middleware.AuthMiddleware())
	{
		fortune.POST("/weekly", h.Weekly)
	}
}
