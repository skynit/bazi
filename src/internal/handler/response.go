package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bazi/internal/model"
	"bazi/internal/service/bazi"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

const (
	ErrCodeInvalidRequest  = "INVALID_REQUEST"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeServiceError    = "SERVICE_ERROR"
	ErrCodeServiceDisabled = "SERVICE_DISABLED"
)

func respondError(c *gin.Context, status int, code string, message string) {
	if strings.TrimSpace(code) == "" {
		code = codeFromStatus(status)
	}
	c.JSON(status, model.ErrorResponse{
		Error:   message,
		Code:    code,
		Message: message,
	})
}

func respondJSON(c *gin.Context, status int, payload interface{}) {
	c.JSON(status, payload)
}

func codeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return ErrCodeInvalidRequest
	case http.StatusUnauthorized:
		return ErrCodeUnauthorized
	case http.StatusForbidden:
		return ErrCodeForbidden
	case http.StatusNotFound:
		return ErrCodeNotFound
	case http.StatusConflict:
		return ErrCodeConflict
	default:
		return ErrCodeServiceError
	}
}

func chartSummaryResponse(chart model.BirthChart) model.ChartSummaryResponse {
	return model.ChartSummaryResponse{
		ID:                    chart.ID,
		Name:                  chart.Name,
		Gender:                chart.Gender,
		ZiHourPolicy:          chart.ZiHourPolicy,
		BirthYear:             chart.BirthYear,
		BirthMonth:            chart.BirthMonth,
		BirthDay:              chart.BirthDay,
		BirthHour:             chart.BirthHour,
		BirthMin:              chart.BirthMin,
		BirthSec:              chart.BirthSec,
		CalendarType:          chart.CalendarType,
		LunarLeapMonth:        chart.LunarLeapMonth,
		BirthPlace:            chart.BirthPlace,
		Timezone:              chart.Timezone,
		BirthUTCOffsetSeconds: cloneInt(chart.BirthUTCOffsetSeconds),
		Longitude:             cloneFloat64(chart.Longitude),
		UseTrueSolarTime:      chart.UseTrueSolarTime,
		TimeUncertain:         chart.TimeUncertain,
		UncertaintySeconds:    chart.UncertaintySeconds,
		SelectedCandidateID:   chart.SelectedCandidateID,
		EngineVersion:         chart.EngineVersion,
		StoredRuleVersion:     chart.RuleVersion,
		CreatedAt:             formatAPITime(chart.CreatedAt),
		UpdatedAt:             formatAPITime(chart.UpdatedAt),
	}
}

func chartDetailResponse(chart model.BirthChart) model.ChartDetailResponse {
	ruleVersion := chart.RuleVersion
	if ruleVersion == "" {
		ruleVersion = bazi.RuleVersion
	}
	return model.ChartDetailResponse{
		ChartSummaryResponse: chartSummaryResponse(chart),
		BirthValidation:      birthValidationFromDB(chart.NormalizedBirth),
		RuleVersion:          ruleVersion,
		School:               bazi.RuleSchool,
		RuleMeta:             bazi.DefaultRuleMeta(),
		YearPillar:           jsonFromDB(chart.YearPillar),
		MonthPillar:          jsonFromDB(chart.MonthPillar),
		DayPillar:            jsonFromDB(chart.DayPillar),
		HourPillar:           jsonFromDB(chart.HourPillar),
		FiveElements:         jsonFromDB(chart.FiveElements),
		ElementDetail:        jsonFromDB(chart.ElementDetail),
		BodyStrength:         jsonFromDB(chart.BodyStrength),
		TenGods:              jsonFromDB(chart.TenGods),
		NaYin:                jsonFromDB(chart.NaYin),
		DaYunStart:           jsonFromDB(chart.DaYunStart),
		DaYun:                jsonFromDB(chart.DaYunStart),
		ZiWeiResult:          jsonFromDB(chart.ZiWeiResult),
		ZiWeiComputed:        chart.ZiWeiComputed,
	}
}

func birthValidationFromDB(raw datatypes.JSON) json.RawMessage {
	if len(raw) == 0 || !json.Valid(raw) {
		return json.RawMessage("null")
	}
	var normalized struct {
		Validation json.RawMessage `json:"validation"`
	}
	if err := json.Unmarshal(raw, &normalized); err != nil || len(normalized.Validation) == 0 || !json.Valid(normalized.Validation) {
		return json.RawMessage("null")
	}
	return normalized.Validation
}

func chartDetailResponseWithBazi(chart model.BirthChart, result *bazi.BaziResult) model.ChartDetailResponse {
	resp := chartDetailResponse(chart)
	if result == nil {
		return resp
	}

	resp.RuleVersion = result.RuleVersion
	resp.School = result.School
	resp.RuleMeta = result.RuleMeta
	resp.YearPillar = jsonFromValue(result.YearPillar)
	resp.MonthPillar = jsonFromValue(result.MonthPillar)
	resp.DayPillar = jsonFromValue(result.DayPillar)
	resp.HourPillar = jsonFromValue(result.HourPillar)
	resp.FiveElements = jsonFromValue(result.FiveElements)
	resp.ElementDetail = jsonFromValue(result.ElementDetail)
	resp.BodyStrength = jsonFromValue(result.BodyStrength)
	resp.TenGods = jsonFromValue(result.TenGods)
	resp.NaYin = jsonFromValue(result.NaYin)
	resp.HiddenStems = jsonFromValue(result.HiddenStems)
	resp.DaYunStart = jsonFromValue(result.DaYunInfo)
	resp.DaYun = jsonFromValue(result.DaYunInfo)
	resp.GanZhiAnalysis = jsonFromValue(result.GanZhiAnalysis)
	resp.PatternAnalysis = jsonFromValue(result.PatternAnalysis)
	resp.MingGong = jsonFromValue(result.MingGong)
	resp.PillarDetails = jsonFromValue(result.PillarDetails)
	resp.Tiaohou = jsonFromValue(result.Tiaohou)
	resp.GlobalShenSha = jsonFromValue(result.GlobalShenSha)
	resp.GlobalShenShaDetails = jsonFromValue(result.GlobalShenShaDetails)
	resp.DayShenSha = jsonFromValue(result.DayShenSha)
	resp.DayShenShaDetails = jsonFromValue(result.DayShenShaDetails)
	resp.MonthSeason = jsonFromValue(result.MonthSeason)
	resp.ShenShaByPillar = jsonFromValue(result.ShenShaByPillar)
	resp.TenGodProportion = jsonFromValue(result.TenGodProportion)
	resp.TenGodAnalysis = jsonFromValue(result.TenGodAnalysis)
	resp.MissingElements = jsonFromValue(result.MissingElements)
	return resp
}

func jsonFromDB(raw datatypes.JSON) json.RawMessage {
	if len(raw) == 0 || !json.Valid(raw) {
		return json.RawMessage("null")
	}
	out := make([]byte, len(raw))
	copy(out, raw)
	return json.RawMessage(out)
}

func jsonFromValue(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil || !json.Valid(b) {
		return json.RawMessage("null")
	}
	return json.RawMessage(b)
}

func formatAPITime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
