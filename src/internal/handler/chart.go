package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"bazi/internal/model"
	"bazi/internal/service/bazi"
	"gorm.io/datatypes"

	"github.com/gin-gonic/gin"
)

type ChartSaver interface {
	Create(chart *model.BirthChart) error
}

type ChartHandler struct {
	Parser *bazi.InputParser
	Bazi   *bazi.BaziService
	Store  ChartSaver
}

type preparedChart struct {
	normalized *bazi.NormalizedBirth
	result     *bazi.BaziResult
}

// Preview validates and calculates a BaZi chart without persisting it. The
// response is used as the confirmation step before a user saves the chart.
func (h *ChartHandler) Preview(c *gin.Context) {
	if _, ok := authUserID(c); !ok {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}
	var req model.ChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}
	prepared, err := h.prepare(req)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
		return
	}
	respondJSON(c, http.StatusOK, chartPayload(0, req, prepared))
}

func (h *ChartHandler) Chart(c *gin.Context) {
	var req model.ChartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	userID, ok := authUserID(c)
	if !ok {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	prepared, err := h.prepare(req)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
		return
	}

	chart, err := buildBirthChart(userID, req, prepared)
	if err != nil {
		respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to serialize chart result")
		return
	}
	if h.Store != nil {
		if err := h.Store.Create(chart); err != nil {
			respondError(c, http.StatusInternalServerError, ErrCodeServiceError, "failed to save chart")
			return
		}
	}

	respondJSON(c, http.StatusOK, chartPayload(chart.ID, req, prepared))
}

func (h *ChartHandler) prepare(req model.ChartRequest) (*preparedChart, error) {
	if h == nil || h.Bazi == nil {
		return nil, fmt.Errorf("chart service is not available")
	}
	normalized, err := bazi.NormalizeBirthInput(bazi.BirthInput{
		Year:             req.BirthYear,
		Month:            req.BirthMonth,
		Day:              req.BirthDay,
		Hour:             req.BirthHour,
		Minute:           req.BirthMin,
		CalendarType:     req.CalendarType,
		LunarLeapMonth:   req.LunarLeapMonth,
		Gender:           req.Gender,
		BirthPlace:       req.BirthPlace,
		Timezone:         req.Timezone,
		Longitude:        req.Longitude,
		UseTrueSolarTime: req.UseTrueSolarTime,
		TimeUncertain:    req.TimeUncertain,
	})
	if err != nil {
		return nil, err
	}
	result, err := h.Bazi.Calculate(
		normalized.Year,
		normalized.Month,
		normalized.Day,
		normalized.Hour,
		normalized.Minute,
		normalized.Gender,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate chart: %w", err)
	}
	return &preparedChart{normalized: normalized, result: result}, nil
}

func buildBirthChart(userID uint, req model.ChartRequest, prepared *preparedChart) (*model.BirthChart, error) {
	result := prepared.result
	name := normalizedChartName(req)

	marshal := func(value interface{}) (datatypes.JSON, error) {
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		return datatypes.JSON(encoded), nil
	}

	yearPillar, err := marshal(result.YearPillar)
	if err != nil {
		return nil, err
	}
	monthPillar, err := marshal(result.MonthPillar)
	if err != nil {
		return nil, err
	}
	dayPillar, err := marshal(result.DayPillar)
	if err != nil {
		return nil, err
	}
	hourPillar, err := marshal(result.HourPillar)
	if err != nil {
		return nil, err
	}
	fiveElements, err := marshal(result.FiveElements)
	if err != nil {
		return nil, err
	}
	elementDetail, err := marshal(result.ElementDetail)
	if err != nil {
		return nil, err
	}
	bodyStrength, err := marshal(result.BodyStrength)
	if err != nil {
		return nil, err
	}
	tenGods, err := marshal(result.TenGods)
	if err != nil {
		return nil, err
	}
	naYin, err := marshal(result.NaYin)
	if err != nil {
		return nil, err
	}
	daYun, err := marshal(result.DaYunInfo)
	if err != nil {
		return nil, err
	}
	normalizedBirth, err := marshal(prepared.normalized)
	if err != nil {
		return nil, err
	}
	snapshot, err := marshal(result)
	if err != nil {
		return nil, err
	}

	return &model.BirthChart{
		UserID:           userID,
		Name:             name,
		Gender:           prepared.normalized.Gender,
		BirthYear:        req.BirthYear,
		BirthMonth:       req.BirthMonth,
		BirthDay:         req.BirthDay,
		BirthHour:        req.BirthHour,
		BirthMin:         req.BirthMin,
		CalendarType:     prepared.normalized.Validation.InputCalendar,
		LunarLeapMonth:   req.LunarLeapMonth,
		BirthPlace:       strings.TrimSpace(req.BirthPlace),
		Timezone:         prepared.normalized.Validation.Timezone,
		Longitude:        cloneFloat64(req.Longitude),
		UseTrueSolarTime: req.UseTrueSolarTime,
		TimeUncertain:    req.TimeUncertain,
		EngineVersion:    bazi.EngineVersion,
		RuleVersion:      result.RuleVersion,
		NormalizedBirth:  normalizedBirth,
		BaziSnapshot:     snapshot,
		YearPillar:       yearPillar,
		MonthPillar:      monthPillar,
		DayPillar:        dayPillar,
		HourPillar:       hourPillar,
		FiveElements:     fiveElements,
		ElementDetail:    elementDetail,
		BodyStrength:     bodyStrength,
		TenGods:          tenGods,
		NaYin:            naYin,
		DaYunStart:       daYun,
	}, nil
}

func chartPayload(id uint, req model.ChartRequest, prepared *preparedChart) gin.H {
	result := prepared.result
	return gin.H{
		"id":                      id,
		"name":                    normalizedChartName(req),
		"gender":                  prepared.normalized.Gender,
		"birth_year":              req.BirthYear,
		"birth_month":             req.BirthMonth,
		"birth_day":               req.BirthDay,
		"birth_hour":              req.BirthHour,
		"birth_min":               req.BirthMin,
		"calendar_type":           prepared.normalized.Validation.InputCalendar,
		"lunar_leap_month":        req.LunarLeapMonth,
		"birth_place":             strings.TrimSpace(req.BirthPlace),
		"timezone":                prepared.normalized.Validation.Timezone,
		"longitude":               req.Longitude,
		"use_true_solar_time":     req.UseTrueSolarTime,
		"time_uncertain":          req.TimeUncertain,
		"birth_validation":        prepared.normalized.Validation,
		"engine_version":          bazi.EngineVersion,
		"rule_version":            result.RuleVersion,
		"school":                  result.School,
		"rule_meta":               result.RuleMeta,
		"year_pillar":             result.YearPillar,
		"month_pillar":            result.MonthPillar,
		"day_pillar":              result.DayPillar,
		"hour_pillar":             result.HourPillar,
		"five_elements":           result.FiveElements,
		"element_detail":          result.ElementDetail,
		"body_strength":           result.BodyStrength,
		"na_yin":                  result.NaYin,
		"ten_gods":                result.TenGods,
		"hidden_stems":            result.HiddenStems,
		"da_yun":                  result.DaYunInfo,
		"clash_harmony":           result.ClashHarmony,
		"gan_zhi_analysis":        result.GanZhiAnalysis,
		"pattern_analysis":        result.PatternAnalysis,
		"ming_gong":               result.MingGong,
		"ri_zhu_desc":             result.RiZhuDesc,
		"pillar_details":          result.PillarDetails,
		"tiao_hou":                result.DayStemTiaoHou,
		"tiaohou":                 result.Tiaohou,
		"global_shen_sha":         result.GlobalShenSha,
		"global_shen_sha_details": result.GlobalShenShaDetails,
		"jin_bu_huan":             result.DayStemJinBuHuan,
		"day_shen_sha":            result.DayShenSha,
		"day_shen_sha_details":    result.DayShenShaDetails,
		"season_text":             result.SeasonText,
		"season_text_month":       result.SeasonTextMonth,
		"ri_zhu_poem":             result.RiZhuPoem,
		"ri_zhu_source":           result.RiZhuSource,
		"ri_zhu_comment":          result.RiZhuComment,
		"ri_zhu_hour_detail":      result.RiZhuHourDetail,
		"shen_sha_by_pillar":      result.ShenShaByPillar,
		"shen_sha_summary":        result.ShenShaSummary,
		"ten_god_proportion":      result.TenGodProportion,
		"ten_god_analysis":        result.TenGodAnalysis,
		"wuxing_season_note":      result.WuxingSeasonNote,
		"wuxing_flow":             result.WuXingFlow,
		"tong_guan":               result.TongGuan,
		"missing_elements":        result.MissingElements,
		"flow_pattern_desc":       result.FlowPatternDesc,
		"dayun_flow":              result.DaYunFlow,
	}
}

func normalizedChartName(req model.ChartRequest) string {
	if name := strings.TrimSpace(req.Name); name != "" {
		return name
	}
	return fmt.Sprintf("%04d-%02d-%02d 命盘", req.BirthYear, req.BirthMonth, req.BirthDay)
}

func cloneFloat64(value *float64) *float64 {
	if value == nil {
		return nil
	}
	copyValue := *value
	return &copyValue
}
