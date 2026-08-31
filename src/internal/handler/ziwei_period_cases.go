package handler

import (
	"fmt"
	"net/http"
	"time"

	"bazi/internal/service/ziwei"

	"github.com/gin-gonic/gin"
)

type ziWeiDayunResponse struct {
	StartAge    int      `json:"start_age"`
	EndAge      int      `json:"end_age"`
	Palace      string   `json:"palace"`
	Stars       []string `json:"stars"`
	Description string   `json:"description"`
}

func (h *ZiWeiPeriodHandler) dispatchPeriod(c *gin.Context, chart *ziwei.ZiWeiChart, userID uint, req ziWeiPeriodRequest) {
	switch req.PeriodType {
	case "dayun":
		h.respondDayun(c, chart, req)
	case "liunian":
		h.respondLiunian(c, chart, req.Year)
	case "liuyue":
		h.respondLiuyue(c, chart, req)
	case "liuri":
		h.respondLiuri(c, chart, req)
	case "sihua_feixing":
		respondJSON(c, http.StatusOK, gin.H{
			"periods":     h.Service.AnalyzeFlyingStars(chart),
			"description": "四化飞星：化禄/化权/化科/化忌在各宫的分布",
		})
	case "sihua_chain":
		respondJSON(c, http.StatusOK, gin.H{
			"chain":       h.Service.AnalyzeSihuaChain(chart),
			"description": "宫干四化直接飞行：展示化曜的来源宫干、目标宫位与同宫/跨宫结构",
		})
	case "self_mutagen":
		respondJSON(c, http.StatusOK, gin.H{
			"self_mutagens": h.Service.AnalyzeSelfMutagen(chart),
			"description":   "自化检测：分析星曜在同宫的自化现象（化禄/化权/化科/化忌留本宫）",
		})
	case "palace_reading":
		h.respondPalaceReading(c, chart, req.PalaceIdx)
	case "query_view":
		respondJSON(c, http.StatusOK, gin.H{
			"query_view":  h.Service.BuildQueryView(chart),
			"description": "紫微查询视图：预计算宫位 has_star、star_index 与三方四正星曜索引。",
		})
	case "heming":
		h.respondHeming(c, chart, userID, req.ChartID2)
	case "liunian_interpretation":
		h.respondLiunianInterpretation(c, chart, req.Year)
	case "liuyue_interpretation":
		h.respondLiuyueInterpretation(c, chart, req)
	case "liuri_interpretation":
		h.respondLiuriInterpretation(c, chart, req)
	case "period_summary":
		h.respondPeriodSummary(c, chart, req)
	case "liu_nian_stars":
		h.respondLiunianStars(c, chart, req.Year)
	default:
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "unknown period_type")
	}
}

func (h *ZiWeiPeriodHandler) respondDayun(c *gin.Context, chart *ziwei.ZiWeiChart, req ziWeiPeriodRequest) {
	year, month, day, ok := resolveZiWeiRequestDate(c, req)
	if !ok {
		return
	}
	targetDate := time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.Local)
	nominalAge := ziwei.NominalAgeAt(chart, targetDate)
	dayun := h.Service.CalculateDayun(chart)
	enriched := make([]ziWeiDayunResponse, len(dayun))
	for i, stage := range dayun {
		enriched[i] = ziWeiDayunResponse{
			StartAge:    stage.StartAge,
			EndAge:      stage.EndAge,
			Palace:      stage.Palace,
			Stars:       stage.Stars,
			Description: ziwei.DayunDescription(stage.Palace, stage.StartAge),
		}
	}
	respondJSON(c, http.StatusOK, gin.H{
		"periods":         enriched,
		"target_date":     fmt.Sprintf("%04d-%02d-%02d", year, month, day),
		"nominal_age":     nominalAge,
		"age_basis":       "target_lunar_year_minus_birth_lunar_year_plus_one",
		"boundary_policy": ziwei.ZiWeiHoroscopeBoundaryNormal,
		"analysis":        ziwei.BuildDayunAnalysis(chart, dayun, nominalAge),
	})
}

func (h *ZiWeiPeriodHandler) respondLiunian(c *gin.Context, chart *ziwei.ZiWeiChart, year int) {
	year = defaultZiWeiLunarYear(year)
	liunian := h.Service.CalculateLiunian(chart, year)
	if liunian == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian year")
		return
	}
	period := mapChartToResponse(liunian, h.Service)
	period["year"] = year
	period["description"] = fmt.Sprintf("%d年流年星曜分布，各宫依次更换", year)
	respondJSON(c, http.StatusOK, gin.H{
		"periods":    []gin.H{period},
		"analysis":   ziwei.BuildLiunianAnalysis(chart, liunian, year),
		"year":       year,
		"period_key": "liunian",
	})
}

func (h *ZiWeiPeriodHandler) respondLiuyue(c *gin.Context, chart *ziwei.ZiWeiChart, req ziWeiPeriodRequest) {
	year, month, day, ok := resolveZiWeiRequestDate(c, req)
	if !ok {
		return
	}
	liuyue := h.Service.CalculateLiuyueForDate(chart, year, month, day)
	if liuyue == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuyue year or month")
		return
	}
	period := mapChartToResponse(liuyue, h.Service)
	period["year"], period["month"], period["day"] = year, month, day
	period["description"] = fmt.Sprintf("%d年%d月%d日所在农历月的流月星曜分布", year, month, day)
	respondJSON(c, http.StatusOK, gin.H{
		"periods":    []gin.H{period},
		"analysis":   ziwei.BuildLiuyueAnalysis(chart, liuyue, year, month, day),
		"year":       year,
		"month":      month,
		"day":        day,
		"period_key": "liuyue",
	})
}

func (h *ZiWeiPeriodHandler) respondLiuri(c *gin.Context, chart *ziwei.ZiWeiChart, req ziWeiPeriodRequest) {
	year, month, day, ok := resolveZiWeiRequestDate(c, req)
	if !ok {
		return
	}
	liuri := h.Service.CalculateLiuriForDate(chart, year, month, day)
	if liuri == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuri solar date")
		return
	}
	period := mapChartToResponse(liuri, h.Service)
	period["year"], period["month"], period["day"] = year, month, day
	period["description"] = fmt.Sprintf("%d年%d月%d日流日星曜分布", year, month, day)
	respondJSON(c, http.StatusOK, gin.H{
		"periods":    []gin.H{period},
		"analysis":   ziwei.BuildLiuriAnalysis(chart, liuri, year, month, day),
		"year":       year,
		"month":      month,
		"day":        day,
		"period_key": "liuri",
	})
}

func (h *ZiWeiPeriodHandler) respondPalaceReading(c *gin.Context, chart *ziwei.ZiWeiChart, palaceIndex int) {
	reading := h.Service.GetPalaceReading(chart, palaceIndex)
	if reading == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid palace index")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"reading": reading})
}

func (h *ZiWeiPeriodHandler) respondHeming(c *gin.Context, chart *ziwei.ZiWeiChart, userID, chartID uint) {
	otherChart, _, err := h.getChart(chartID, userID)
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, err.Error())
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"heming": h.Service.AnalyzeHeming(chart, otherChart)})
}

func (h *ZiWeiPeriodHandler) respondLiunianInterpretation(c *gin.Context, chart *ziwei.ZiWeiChart, year int) {
	year = defaultZiWeiLunarYear(year)
	liunian := h.Service.CalculateLiunian(chart, year)
	if liunian == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian year")
		return
	}
	interpreter, ok := periodInterpreter(c, chart)
	if !ok {
		return
	}
	result := interpreter.AnalyzeLiunian(liunian, year)
	if result == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian derivation contract")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"periods": []*ziwei.LiunianResult{result}})
}

func (h *ZiWeiPeriodHandler) respondLiuyueInterpretation(c *gin.Context, chart *ziwei.ZiWeiChart, req ziWeiPeriodRequest) {
	year, month, day, ok := resolveZiWeiRequestDate(c, req)
	if !ok {
		return
	}
	liuyue := h.Service.CalculateLiuyueForDate(chart, year, month, day)
	if liuyue == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuyue year or month")
		return
	}
	interpreter, ok := periodInterpreter(c, chart)
	if !ok {
		return
	}
	result := interpreter.AnalyzeLiuyue(liuyue, year, month, day)
	if result == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuyue derivation contract")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"periods": []*ziwei.LiuyueResult{result}})
}

func (h *ZiWeiPeriodHandler) respondLiuriInterpretation(c *gin.Context, chart *ziwei.ZiWeiChart, req ziWeiPeriodRequest) {
	year, month, day, ok := resolveZiWeiRequestDate(c, req)
	if !ok {
		return
	}
	liuri := h.Service.CalculateLiuriForDate(chart, year, month, day)
	if liuri == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuri solar date")
		return
	}
	interpreter, ok := periodInterpreter(c, chart)
	if !ok {
		return
	}
	result := interpreter.AnalyzeLiuri(liuri, year, month, day)
	if result == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuri derivation contract")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"periods": []*ziwei.LiuriResult{result}})
}

func (h *ZiWeiPeriodHandler) respondPeriodSummary(c *gin.Context, chart *ziwei.ZiWeiChart, req ziWeiPeriodRequest) {
	year, month, day, ok := resolveZiWeiRequestDate(c, req)
	if !ok {
		return
	}
	lunarYear, err := ziwei.LunarYearLabelForSolarDate(year, month, day)
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
		return
	}
	liunian := h.Service.CalculateLiunian(chart, lunarYear)
	liuyue := h.Service.CalculateLiuyueForDate(chart, year, month, day)
	liuri := h.Service.CalculateLiuriForDate(chart, year, month, day)
	if liunian == nil || liuyue == nil || liuri == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid period summary solar date")
		return
	}
	interpreter, ok := periodInterpreter(c, chart)
	if !ok {
		return
	}
	summary := interpreter.SummarizeAll(liunian, liuyue, liuri, year, month, day)
	if summary == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "inconsistent period summary context")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"summary": summary})
}

func (h *ZiWeiPeriodHandler) respondLiunianStars(c *gin.Context, chart *ziwei.ZiWeiChart, year int) {
	year = defaultZiWeiLunarYear(year)
	liunian := h.Service.CalculateLiunian(chart, year)
	if liunian == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian year")
		return
	}
	respondJSON(c, http.StatusOK, gin.H{"palaces": liunian.LiuNianStars, "year": year})
}

func resolveZiWeiRequestDate(c *gin.Context, req ziWeiPeriodRequest) (int, int, int, bool) {
	year, month, day, err := ziwei.ResolvePeriodSolarDate(req.Year, req.Month, req.Day, time.Now())
	if err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, err.Error())
		return 0, 0, 0, false
	}
	return year, month, day, true
}

func defaultZiWeiLunarYear(year int) int {
	if year != 0 {
		return year
	}
	return ziwei.CurrentLunarYearLabel(time.Now())
}

func periodInterpreter(c *gin.Context, chart *ziwei.ZiWeiChart) (*ziwei.PeriodInterpreter, bool) {
	interpreter := ziwei.NewPeriodInterpreterFromChart(chart)
	if interpreter == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid natal chart contract")
		return nil, false
	}
	return interpreter, true
}
