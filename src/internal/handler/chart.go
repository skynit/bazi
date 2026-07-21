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

type preparedCandidateSet struct {
	center     *preparedChart
	candidates *bazi.BirthCandidateSet
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
	preparedSet, err := h.prepareCandidates(req)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
		return
	}
	respondJSON(c, http.StatusOK, chartPayload(0, req, preparedSet.center, preparedSet.candidates, ""))
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

	preparedSet, err := h.prepareCandidates(req)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
		return
	}
	prepared, selectedCandidateID, err := selectPreparedCandidate(preparedSet, req.CandidateID)
	if err != nil {
		status := http.StatusBadRequest
		if preparedSet.candidates.RequiresCandidateSelection && strings.TrimSpace(req.CandidateID) == "" {
			status = http.StatusConflict
		}
		respondError(c, status, codeFromStatus(status), err.Error())
		return
	}
	req.CandidateID = selectedCandidateID

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

	respondJSON(c, http.StatusOK, chartPayload(chart.ID, req, prepared, preparedSet.candidates, selectedCandidateID))
}

func (h *ChartHandler) prepareCandidates(req model.ChartRequest) (*preparedCandidateSet, error) {
	if h == nil || h.Bazi == nil {
		return nil, fmt.Errorf("chart service is not available")
	}
	set, err := bazi.CalculateBirthCandidates(h.Bazi, birthInputFromChartRequest(req))
	if err != nil {
		return nil, err
	}
	return &preparedCandidateSet{
		center:     &preparedChart{normalized: set.Center, result: set.CenterResult},
		candidates: set,
	}, nil
}

func birthInputFromChartRequest(req model.ChartRequest) bazi.BirthInput {
	return bazi.BirthInput{
		Year:               req.BirthYear,
		Month:              req.BirthMonth,
		Day:                req.BirthDay,
		Hour:               req.BirthHour,
		Minute:             req.BirthMin,
		Second:             req.BirthSec,
		CalendarType:       req.CalendarType,
		LunarLeapMonth:     req.LunarLeapMonth,
		Gender:             req.Gender,
		ZiHourPolicy:       req.ZiHourPolicy,
		BirthPlace:         req.BirthPlace,
		Timezone:           req.Timezone,
		UTCOffsetSeconds:   req.BirthUTCOffsetSeconds,
		Longitude:          req.Longitude,
		UseTrueSolarTime:   req.UseTrueSolarTime,
		TimeUncertain:      req.TimeUncertain,
		UncertaintySeconds: req.UncertaintySeconds,
	}
}

func selectPreparedCandidate(set *preparedCandidateSet, requestedID string) (*preparedChart, string, error) {
	if set == nil || set.candidates == nil || len(set.candidates.Candidates) == 0 {
		return nil, "", fmt.Errorf("no calculated chart candidate is available")
	}
	requestedID = strings.TrimSpace(requestedID)
	if requestedID == "" {
		if set.candidates.RequiresCandidateSelection {
			return nil, "", fmt.Errorf("birth-time interval crosses a four-pillar boundary; submit a candidate_id returned by preview")
		}
		candidate := set.candidates.Candidates[0]
		return &preparedChart{normalized: candidate.Normalized, result: candidate.Result}, candidate.CandidateID, nil
	}
	for _, candidate := range set.candidates.Candidates {
		if candidate.CandidateID == requestedID {
			return &preparedChart{normalized: candidate.Normalized, result: candidate.Result}, candidate.CandidateID, nil
		}
	}
	return nil, "", fmt.Errorf("candidate_id does not match the current birth input and uncertainty interval")
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
		UserID:                userID,
		Name:                  name,
		Gender:                prepared.normalized.Gender,
		ZiHourPolicy:          prepared.normalized.ZiHourPolicy,
		BirthYear:             req.BirthYear,
		BirthMonth:            req.BirthMonth,
		BirthDay:              req.BirthDay,
		BirthHour:             req.BirthHour,
		BirthMin:              req.BirthMin,
		BirthSec:              req.BirthSec,
		CalendarType:          prepared.normalized.Validation.InputCalendar,
		LunarLeapMonth:        req.LunarLeapMonth,
		BirthPlace:            strings.TrimSpace(req.BirthPlace),
		Timezone:              prepared.normalized.Validation.Timezone,
		BirthUTCOffsetSeconds: cloneInt(req.BirthUTCOffsetSeconds),
		Longitude:             cloneFloat64(req.Longitude),
		UseTrueSolarTime:      req.UseTrueSolarTime,
		TimeUncertain:         req.TimeUncertain,
		UncertaintySeconds:    req.UncertaintySeconds,
		SelectedCandidateID:   strings.TrimSpace(req.CandidateID),
		EngineVersion:         bazi.EngineVersion,
		RuleVersion:           result.RuleVersion,
		NormalizedBirth:       normalizedBirth,
		BaziSnapshot:          snapshot,
		YearPillar:            yearPillar,
		MonthPillar:           monthPillar,
		DayPillar:             dayPillar,
		HourPillar:            hourPillar,
		FiveElements:          fiveElements,
		ElementDetail:         elementDetail,
		BodyStrength:          bodyStrength,
		TenGods:               tenGods,
		NaYin:                 naYin,
		DaYunStart:            daYun,
	}, nil
}

func chartPayload(id uint, req model.ChartRequest, prepared *preparedChart, set *bazi.BirthCandidateSet, selectedCandidateID string) gin.H {
	result := prepared.result
	return gin.H{
		"id":                           id,
		"name":                         normalizedChartName(req),
		"gender":                       prepared.normalized.Gender,
		"zi_hour_policy":               prepared.normalized.ZiHourPolicy,
		"birth_year":                   req.BirthYear,
		"birth_month":                  req.BirthMonth,
		"birth_day":                    req.BirthDay,
		"birth_hour":                   req.BirthHour,
		"birth_min":                    req.BirthMin,
		"birth_sec":                    req.BirthSec,
		"calendar_type":                prepared.normalized.Validation.InputCalendar,
		"lunar_leap_month":             req.LunarLeapMonth,
		"birth_place":                  strings.TrimSpace(req.BirthPlace),
		"timezone":                     prepared.normalized.Validation.Timezone,
		"birth_utc_offset_seconds":     req.BirthUTCOffsetSeconds,
		"longitude":                    req.Longitude,
		"use_true_solar_time":          req.UseTrueSolarTime,
		"time_uncertain":               req.TimeUncertain,
		"uncertainty_seconds":          req.UncertaintySeconds,
		"selected_candidate_id":        selectedCandidateID,
		"uncertainty":                  set.Uncertainty,
		"candidate_charts":             candidatePayloads(set.Candidates),
		"stable_fields":                set.StableFields,
		"unstable_fields":              set.UnstableFields,
		"requires_candidate_selection": set.RequiresCandidateSelection,
		"birth_validation":             prepared.normalized.Validation,
		"engine_version":               bazi.EngineVersion,
		"rule_version":                 result.RuleVersion,
		"school":                       result.School,
		"rule_meta":                    result.RuleMeta,
		"year_pillar":                  result.YearPillar,
		"month_pillar":                 result.MonthPillar,
		"day_pillar":                   result.DayPillar,
		"hour_pillar":                  result.HourPillar,
		"five_elements":                result.FiveElements,
		"element_detail":               result.ElementDetail,
		"body_strength":                result.BodyStrength,
		"na_yin":                       result.NaYin,
		"ten_gods":                     result.TenGods,
		"hidden_stems":                 result.HiddenStems,
		"da_yun":                       result.DaYunInfo,
		"gan_zhi_analysis":             result.GanZhiAnalysis,
		"pattern_analysis":             result.PatternAnalysis,
		"ming_gong":                    result.MingGong,
		"pillar_details":               result.PillarDetails,
		"tiaohou":                      result.Tiaohou,
		"global_shen_sha":              result.GlobalShenSha,
		"global_shen_sha_details":      result.GlobalShenShaDetails,
		"day_shen_sha":                 result.DayShenSha,
		"day_shen_sha_details":         result.DayShenShaDetails,
		"month_season":                 result.MonthSeason,
		"shen_sha_by_pillar":           result.ShenShaByPillar,
		"ten_god_proportion":           result.TenGodProportion,
		"ten_god_analysis":             result.TenGodAnalysis,
		"missing_elements":             result.MissingElements,
	}
}

func candidatePayloads(candidates []bazi.BirthChartCandidate) []gin.H {
	payloads := make([]gin.H, 0, len(candidates))
	for _, candidate := range candidates {
		payloads = append(payloads, gin.H{
			"candidate_id":            candidate.CandidateID,
			"input_range_start":       candidate.InputRangeStart,
			"input_range_end":         candidate.InputRangeEnd,
			"calculation_range_start": candidate.CalculationRangeStart,
			"calculation_range_end":   candidate.CalculationRangeEnd,
			"representative_time":     candidate.RepresentativeTime,
			"birth_validation":        candidate.Normalized.Validation,
			"year_pillar":             candidate.Result.YearPillar,
			"month_pillar":            candidate.Result.MonthPillar,
			"day_pillar":              candidate.Result.DayPillar,
			"hour_pillar":             candidate.Result.HourPillar,
			"da_yun_start_at_min":     candidate.DaYunStartAtMin,
			"da_yun_start_at_max":     candidate.DaYunStartAtMax,
		})
	}
	return payloads
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

func cloneInt(value *int) *int {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
