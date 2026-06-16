package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"bazi/internal/model"

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
		ID:           chart.ID,
		Name:         chart.Name,
		Gender:       chart.Gender,
		BirthYear:    chart.BirthYear,
		BirthMonth:   chart.BirthMonth,
		BirthDay:     chart.BirthDay,
		BirthHour:    chart.BirthHour,
		BirthMin:     chart.BirthMin,
		CalendarType: chart.CalendarType,
		CreatedAt:    formatAPITime(chart.CreatedAt),
		UpdatedAt:    formatAPITime(chart.UpdatedAt),
	}
}

func chartDetailResponse(chart model.BirthChart) model.ChartDetailResponse {
	return model.ChartDetailResponse{
		ChartSummaryResponse: chartSummaryResponse(chart),
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

func jsonFromDB(raw datatypes.JSON) json.RawMessage {
	if len(raw) == 0 || !json.Valid(raw) {
		return json.RawMessage("null")
	}
	out := make([]byte, len(raw))
	copy(out, raw)
	return json.RawMessage(out)
}

func formatAPITime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
