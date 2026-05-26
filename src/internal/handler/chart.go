package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"bazi/internal/model"
	"bazi/internal/service"
	"gorm.io/datatypes"

	"github.com/gin-gonic/gin"
)

type ChartSaver interface {
	Create(chart *model.BirthChart) error
}

type ChartHandler struct {
	Parser *service.InputParser
	Bazi   *service.BaziService
	Store  ChartSaver
}

func mustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func (h *ChartHandler) Chart(c *gin.Context) {
	var req model.ChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	input := formatChartInput(req)
	parsed, err := h.Parser.Parse(service.ParseRequest{
		Input: input, CalendarType: req.CalendarType, Gender: req.Gender,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.Bazi.Calculate(parsed.Year, parsed.Month, parsed.Day, parsed.Hour, parsed.Minute, parsed.Gender)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	chart := &model.BirthChart{
		UserID:       userID.(uint),
		Name:         req.Name,
		Gender:       req.Gender,
		BirthYear:    req.BirthYear,
		BirthMonth:   req.BirthMonth,
		BirthDay:     req.BirthDay,
		BirthHour:    req.BirthHour,
		BirthMin:     req.BirthMin,
		CalendarType: req.CalendarType,
		YearPillar:    datatypes.JSON(mustJSON(result.YearPillar)),
		MonthPillar:   datatypes.JSON(mustJSON(result.MonthPillar)),
		DayPillar:     datatypes.JSON(mustJSON(result.DayPillar)),
		HourPillar:    datatypes.JSON(mustJSON(result.HourPillar)),
		FiveElements:  datatypes.JSON(mustJSON(result.FiveElements)),
		ElementDetail: datatypes.JSON(mustJSON(result.ElementDetail)),
		BodyStrength:  datatypes.JSON(mustJSON(result.BodyStrength)),
		TenGods:       datatypes.JSON(mustJSON(result.TenGods)),
		NaYin:         datatypes.JSON(mustJSON(result.NaYin)),
	}
	if h.Store != nil {
		if err := h.Store.Create(chart); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "save failed: " + err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":             chart.ID,
		"year_pillar":    result.YearPillar,
		"month_pillar":   result.MonthPillar,
		"day_pillar":     result.DayPillar,
		"hour_pillar":    result.HourPillar,
		"five_elements":  result.FiveElements,
		"element_detail": result.ElementDetail,
		"body_strength":  result.BodyStrength,
		"na_yin":         result.NaYin,
		"ten_gods":       result.TenGods,
		"hidden_stems":   result.HiddenStems,
		"da_yun":         result.DaYunInfo,
		"clash_harmony":     result.ClashHarmony,
		"gan_zhi_analysis":  result.GanZhiAnalysis,
		"pattern_analysis":  result.PatternAnalysis,
		"ming_gong":         result.MingGong,
		"ri_zhu_desc":    result.RiZhuDesc,
		"pillar_details": result.PillarDetails,
		"tiao_hou":       result.DayStemTiaoHou,
		"jin_bu_huan":    result.DayStemJinBuHuan,
		"day_shen_sha":   result.DayShenSha,
		"season_text":        result.SeasonText,
		"season_text_month":  result.SeasonTextMonth,
		"ri_zhu_poem":        result.RiZhuPoem,
		"ri_zhu_source":      result.RiZhuSource,
		"ri_zhu_comment":     result.RiZhuComment,
		"ri_zhu_hour_detail":   result.RiZhuHourDetail,
		"shen_sha_by_pillar":   result.ShenShaByPillar,
		"shen_sha_summary":     result.ShenShaSummary,
		"ten_god_proportion":   result.TenGodProportion,
		"ten_god_analysis":    result.TenGodAnalysis,
		"birth_month":          req.BirthMonth,
	})
}

func formatChartInput(req model.ChartRequest) string {
	return fmt.Sprintf("%04d-%02d-%02d %02d:%02d",
		req.BirthYear, req.BirthMonth, req.BirthDay,
		req.BirthHour, req.BirthMin,
	)
}
