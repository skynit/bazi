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

// ZiWeiPeriodHandler handles period (dayun/liunian/liuyue/liuri) and overlay calculations.
type ZiWeiPeriodHandler struct {
	Charts  ChartStore
	Service *ziwei.ZiWeiService
}

// getChart looks up the birth chart and calculates the ZiWeiChart.
// Uses cached result when available (same as ziwei_chart.go).
func (h *ZiWeiPeriodHandler) getChart(chartID uint, userID uint) (*ziwei.ZiWeiChart, *model.BirthChart, error) {
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
	// Try cached result first
	if birthChart.ZiWeiComputed && len(birthChart.ZiWeiResult) > 0 {
		var cached ziwei.ZiWeiChart
		if err := json.Unmarshal(birthChart.ZiWeiResult, &cached); err == nil {
			if err := h.Service.AttachBirthData(&cached, birthChart.BirthYear, birthChart.BirthMonth, birthChart.BirthDay, birthChart.BirthHour, birthChart.BirthMin, birthChart.Gender); err != nil {
				return nil, nil, fmt.Errorf("attach birth data: %w", err)
			}
			return &cached, birthChart, nil
		}
	}
	chart, err := h.Service.CalculateChart(birthChart.BirthYear, birthChart.BirthMonth, birthChart.BirthDay, birthChart.BirthHour, birthChart.BirthMin, birthChart.Gender)
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

// Period handles dayun, liunian, liuyue, liuri, and sihua_feixing period calculations.
func (h *ZiWeiPeriodHandler) Period(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		respondError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
		return
	}

	var req struct {
		ChartID    uint   `json:"chart_id"`
		PeriodType string `json:"period_type"`
		Year       int    `json:"year"`
		Month      int    `json:"month"`
		Day        int    `json:"day"`
		PalaceIdx  int    `json:"palace_idx"`
		ChartID2   uint   `json:"chart_id2"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid request body")
		return
	}

	chart, birthChart, err := h.getChart(req.ChartID, userID.(uint))
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, err.Error())
		return
	}

	svc := h.Service

	switch req.PeriodType {
	case "dayun":
		dayun := svc.CalculateDayun(chart)
		// Enrich each dayun stage with description and palace stars
		type enrichedDayun struct {
			StartAge    int      `json:"start_age"`
			EndAge      int      `json:"end_age"`
			Palace      string   `json:"palace"`
			Stars       []string `json:"stars"`
			Description string   `json:"description"`
		}
		enriched := make([]enrichedDayun, len(dayun))
		for i, d := range dayun {
			enriched[i] = enrichedDayun{
				StartAge:    d.StartAge,
				EndAge:      d.EndAge,
				Palace:      d.Palace,
				Stars:       d.Stars,
				Description: dayunDesc(d.Palace, d.StartAge),
			}
		}
		resp := gin.H{"periods": enriched}
		resp["analysis"] = ziwei.BuildDayunAnalysis(chart, dayun, currentAgeFromBirthChart(birthChart))
		respondJSON(c, http.StatusOK, resp)

	case "liunian":
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liunian := svc.CalculateLiunian(chart, year)
		resp := mapChartToResponse(liunian)
		resp["year"] = year
		resp["description"] = fmt.Sprintf("%d年流年星曜分布，各宫依次更换", year)
		respondJSON(c, http.StatusOK, gin.H{
			"periods":   []gin.H{resp},
			"analysis":   ziwei.BuildLiunianAnalysis(chart, liunian, year),
			"year":       year,
			"period_key": "liunian",
		})

	case "liuyue":
		month := req.Month
		if month == 0 {
			month = int(time.Now().Month())
		}
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liuyue := svc.CalculateLiuyueForYear(chart, year, month)
		resp := mapChartToResponse(liuyue)
		resp["year"] = year
		resp["month"] = month
		resp["description"] = fmt.Sprintf("%d年%d月流月星曜分布", year, month)
		respondJSON(c, http.StatusOK, gin.H{
			"periods":   []gin.H{resp},
			"analysis":   ziwei.BuildLiuyueAnalysis(chart, liuyue, year, month),
			"year":       year,
			"month":      month,
			"period_key": "liuyue",
		})

	case "liuri":
		day := req.Day
		if day == 0 {
			day = time.Now().Day()
		}
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		month := req.Month
		if month == 0 {
			month = int(time.Now().Month())
		}
		liuri := svc.CalculateLiuriForDate(chart, year, month, day)
		resp := mapChartToResponse(liuri)
		resp["year"] = year
		resp["month"] = month
		resp["day"] = day
		resp["description"] = fmt.Sprintf("%d年%d月%d日流日星曜分布", year, month, day)
		respondJSON(c, http.StatusOK, gin.H{
			"periods":   []gin.H{resp},
			"analysis":   ziwei.BuildLiuriAnalysis(chart, liuri, year, month, day),
			"year":       year,
			"month":      month,
			"day":        day,
			"period_key": "liuri",
		})

	case "sihua_feixing":
		flying := svc.AnalyzeFlyingStars(chart)
		// Remove unused fields, return clean struct
		respondJSON(c, http.StatusOK, gin.H{
			"periods":     flying,
			"description": "四化飞星：化禄/化权/化科/化忌在各宫的分布",
		})

	case "sihua_chain":
		chain := svc.AnalyzeSihuaChain(chart)
		respondJSON(c, http.StatusOK, gin.H{
			"chain":       chain,
			"description": "四化飞星链式分析：追踪每颗四化星的来源宫位与链式影响",
		})

	case "self_mutagen":
		result := svc.AnalyzeSelfMutagen(chart)
		respondJSON(c, http.StatusOK, gin.H{
			"self_mutagens": result,
			"description":   "自化检测：分析星曜在同宫的自化现象（化禄/化权/化科/化忌留本宫）",
		})

	case "palace_reading":
		reading := svc.GetPalaceReading(chart, req.PalaceIdx)
		if reading == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid palace index")
			return
		}
		respondJSON(c, http.StatusOK, gin.H{
			"reading": reading,
		})

	case "query_view":
		respondJSON(c, http.StatusOK, gin.H{
			"query_view":  svc.BuildQueryView(chart),
			"description": "紫微查询视图：预计算宫位 has_star、star_index 与三方四正星曜索引。",
		})

	case "heming":
		chart2, _, err := h.getChart(req.ChartID2, userID.(uint))
		if err != nil {
			respondError(c, http.StatusNotFound, ErrCodeNotFound, err.Error())
			return
		}
		result := svc.AnalyzeHeming(chart, chart2)
		respondJSON(c, http.StatusOK, gin.H{
			"heming": result,
		})

	case "liunian_interpretation":
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liunian := svc.CalculateLiunian(chart, year)
		interp := ziwei.NewPeriodInterpreter(chart.GetBirthData())
		result := interp.AnalyzeLiunian(liunian, year)
		respondJSON(c, http.StatusOK, gin.H{"periods": []gin.H{
			{
				"year":             result.Year,
				"gan_zhi":          result.GanZhi,
				"gan_zhi_desc":     result.GanZhiDesc,
				"shi_shen":         result.ShiShen,
				"relation_to_ming": result.RelationToMing,
				"overall_tone":     result.OverallTone,
				"key_tips":         result.KeyTips,
				"score":            result.Score,
			},
		}})

	case "liuyue_interpretation":
		month := req.Month
		if month == 0 {
			month = int(time.Now().Month())
		}
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liuyue := svc.CalculateLiuyueForYear(chart, year, month)
		interp := ziwei.NewPeriodInterpreter(chart.GetBirthData())
		result := interp.AnalyzeLiuyue(liuyue, year, month)
		respondJSON(c, http.StatusOK, gin.H{"periods": []gin.H{
			{
				"year":             result.Year,
				"month":            result.Month,
				"gan_zhi":          result.GanZhi,
				"gan_zhi_desc":     result.GanZhiDesc,
				"shi_shen":         result.ShiShen,
				"relation_to_ming": result.RelationToMing,
				"effect":           result.Effect,
				"health":           result.Health,
				"score":            result.Score,
			},
		}})

	case "liuri_interpretation":
		day := req.Day
		if day == 0 {
			day = time.Now().Day()
		}
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		month := req.Month
		if month == 0 {
			month = int(time.Now().Month())
		}
		liuri := svc.CalculateLiuriForDate(chart, year, month, day)
		interp := ziwei.NewPeriodInterpreter(chart.GetBirthData())
		result := interp.AnalyzeLiuri(liuri, year, month, day)
		// Convert HourlyAnalysis to []gin.H for JSON
		hourly := make([]gin.H, len(result.HourlyAnalysis))
		for i, ha := range result.HourlyAnalysis {
			hourly[i] = gin.H{
				"hour":        ha.Hour,
				"stem_branch": ha.StemBranch,
				"effect":      ha.Effect,
				"score":       ha.Score,
			}
		}
		respondJSON(c, http.StatusOK, gin.H{"periods": []gin.H{
			{
				"year":             result.Year,
				"month":            result.Month,
				"day":              result.Day,
				"gan_zhi":          result.GanZhi,
				"gan_zhi_desc":     result.GanZhiDesc,
				"shi_shen":         result.ShiShen,
				"relation_to_ming": result.RelationToMing,
				"qi_zi_effect":     result.QiZiEffect,
				"emotional_state":  result.EmotionalState,
				"health":           result.Health,
				"score":            result.Score,
				"hourly_analysis":  hourly,
				"summary":          result.Summary,
			},
		}})

	case "period_summary":
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		month := req.Month
		if month == 0 {
			month = int(time.Now().Month())
		}
		day := req.Day
		if day == 0 {
			day = time.Now().Day()
		}
		liunian := svc.CalculateLiunian(chart, year)
		liuyue := svc.CalculateLiuyueForYear(chart, year, month)
		liuri := svc.CalculateLiuriForDate(chart, year, month, day)
		interp := ziwei.NewPeriodInterpreter(chart.GetBirthData())
		summary := interp.SummarizeAll(liunian, liuyue, liuri, year, month, day)
		respondJSON(c, http.StatusOK, gin.H{"summary": summary})

	case "liu_nian_stars":
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liunian := svc.CalculateLiunian(chart, year)
		respondJSON(c, http.StatusOK, gin.H{
			"palaces": liunian.LiuNianStars,
			"year":    year,
		})

	default:
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "unknown period_type")
	}
}

// Overlay handles the liunian overlay calculation.
func (h *ZiWeiPeriodHandler) Overlay(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
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

	chart, _, err := h.getChart(req.ChartID, userID.(uint))
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, err.Error())
		return
	}

	year := req.Year
	if year == 0 {
		year = time.Now().Year()
	}
	liunian := h.Service.CalculateLiunian(chart, year)
	result := mapChartToResponse(liunian)
	result["year"] = year
	result["liu_nian_stars"] = liunian.LiuNianStars
	result["overlay_analysis"] = h.Service.AnalyzeLiunianOverlay(chart, liunian, year)
	respondJSON(c, http.StatusOK, result)
}

func dayunDesc(palace string, startAge int) string {
	descs := map[string]string{
		"命宮":  "个人运势与性格转变的关键十年",
		"兄弟宮": "兄弟姐妹关系与助力变化",
		"夫妻宮": "婚姻感情与配偶关系的关键时期",
		"子女宮": "子女缘分与下属关系的变化",
		"財帛宮": "财运金钱进出的关键阶段",
		"疾厄宮": "身体健康状况的重要周期",
		"遷移宮": "外出运程与社会形象的转变",
		"僕役宮": "朋友与部属关系的十年变化",
		"官祿宮": "事业运程与工作成就的关键期",
		"田宅宮": "房产运程与家庭环境的变化",
		"福德宮": "精神享受与内心世界的重要阶段",
		"父母宮": "父母缘分与长辈助力的变化",
	}
	if d, ok := descs[palace]; ok {
		return d
	}
	return fmt.Sprintf("%s%s-%d岁大限", palace, palace, startAge)
}

func currentAgeFromBirthChart(chart *model.BirthChart) int {
	if chart == nil {
		return 0
	}
	nowYear := time.Now().Year()
	age := nowYear - chart.BirthYear
	if age < 0 {
		return 0
	}
	return age
}

// RegisterZiWeiPeriodRoutes registers ZiWei period and overlay routes.
func RegisterZiWeiPeriodRoutes(r gin.IRouter, svc *ziwei.ZiWeiService, store ChartStore) {
	h := &ZiWeiPeriodHandler{Service: svc, Charts: store}
	r.POST("/ziwei/period", h.Period)
	r.POST("/ziwei/overlay", h.Overlay)
}
