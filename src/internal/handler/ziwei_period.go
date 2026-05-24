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
	if _, exists := c.Get("userID"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
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
		c.JSON(http.StatusOK, gin.H{"periods": enriched})

	case "liunian":
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liunian := svc.CalculateLiunian(chart, year)
		resp := mapChartToResponse(liunian, birthChart.BirthMonth, birthChart.BirthHour)
		resp["year"] = year
		resp["description"] = fmt.Sprintf("%d年流年星曜分布，各宫依次更换", year)
		c.JSON(http.StatusOK, gin.H{"periods": []gin.H{resp}})

	case "liuyue":
		month := req.Month
		if month == 0 {
			month = int(time.Now().Month())
		}
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liuyue := svc.CalculateLiuyue(chart, month)
		resp := mapChartToResponse(liuyue, birthChart.BirthMonth, birthChart.BirthHour)
		resp["year"] = year
		resp["month"] = month
		resp["description"] = fmt.Sprintf("%d年%d月流月星曜分布", year, month)
		c.JSON(http.StatusOK, gin.H{"periods": []gin.H{resp}})

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
		liuri := svc.CalculateLiuri(chart, day)
		resp := mapChartToResponse(liuri, birthChart.BirthMonth, birthChart.BirthHour)
		resp["year"] = year
		resp["month"] = month
		resp["day"] = day
		resp["description"] = fmt.Sprintf("%d年%d月%d日流日星曜分布", year, month, day)
		c.JSON(http.StatusOK, gin.H{"periods": []gin.H{resp}})

	case "sihua_feixing":
		flying := svc.AnalyzeFlyingStars(chart)
		// Remove unused fields, return clean struct
		c.JSON(http.StatusOK, gin.H{
			"periods": flying,
			"description": "四化飞星：化禄/化权/化科/化忌在各宫的分布",
		})

	case "sihua_chain":
		chain := svc.AnalyzeSihuaChain(chart)
		c.JSON(http.StatusOK, gin.H{
			"chain": chain,
			"description": "四化飞星链式分析：追踪每颗四化星的来源宫位与链式影响",
		})

	case "self_mutagen":
		result := svc.AnalyzeSelfMutagen(chart)
		c.JSON(http.StatusOK, gin.H{
			"self_mutagens": result,
			"description": "自化检测：分析星曜在同宫的自化现象（化禄/化权/化科/化忌留本宫）",
		})

	case "palace_reading":
		reading := svc.GetPalaceReading(chart, req.PalaceIdx)
		if reading == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid palace index"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"reading": reading,
		})

	case "heming":
		chart2, _, err := h.getChart(req.ChartID2)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		result := svc.AnalyzeHeming(chart, chart2)
		c.JSON(http.StatusOK, gin.H{
			"heming": result,
		})

	case "liunian_interpretation":
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liunian := svc.CalculateLiunian(chart, year)
		interp := service.NewPeriodInterpreter(chart.BirthInfo())
		result := interp.AnalyzeLiunian(liunian, year)
		c.JSON(http.StatusOK, gin.H{"periods": []gin.H{
			{
				"year":             result.Year,
				"gan_zhi":          result.GanZhi,
				"gan_zhi_desc":     result.GanZhiDesc,
				"shi_shen":        result.ShiShen,
				"relation_to_ming": result.RelationToMing,
				"overall_tone":     result.OverallTone,
				"key_tips":        result.KeyTips,
				"score":           result.Score,
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
		liuyue := svc.CalculateLiuyue(chart, month)
		interp := service.NewPeriodInterpreter(chart.BirthInfo())
		result := interp.AnalyzeLiuyue(liuyue, year, month)
		c.JSON(http.StatusOK, gin.H{"periods": []gin.H{
			{
				"year":             result.Year,
				"month":            result.Month,
				"gan_zhi":          result.GanZhi,
				"gan_zhi_desc":     result.GanZhiDesc,
				"shi_shen":        result.ShiShen,
				"relation_to_ming": result.RelationToMing,
				"effect":          result.Effect,
				"health":          result.Health,
				"score":           result.Score,
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
		liuri := svc.CalculateLiuri(chart, day)
		interp := service.NewPeriodInterpreter(chart.BirthInfo())
		result := interp.AnalyzeLiuri(liuri, year, month, day)
		// Convert HourlyAnalysis to []gin.H for JSON
		hourly := make([]gin.H, len(result.HourlyAnalysis))
		for i, h := range result.HourlyAnalysis {
			hourly[i] = gin.H{
				"hour":        h.Hour,
				"stem_branch": h.StemBranch,
				"effect":      h.Effect,
				"score":       h.Score,
			}
		}
		c.JSON(http.StatusOK, gin.H{"periods": []gin.H{
			{
				"year":             result.Year,
				"month":            result.Month,
				"day":              result.Day,
				"gan_zhi":          result.GanZhi,
				"gan_zhi_desc":     result.GanZhiDesc,
				"shi_shen":        result.ShiShen,
				"relation_to_ming": result.RelationToMing,
				"qi_zi_effect":    result.QiZiEffect,
				"emotional_state": result.EmotionalState,
				"health":          result.Health,
				"score":           result.Score,
				"hourly_analysis":  hourly,
				"summary":         result.Summary,
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
		liuyue := svc.CalculateLiuyue(chart, month)
		liuri := svc.CalculateLiuri(chart, day)
		interp := service.NewPeriodInterpreter(chart.BirthInfo())
		summary := interp.SummarizeAll(liunian, liuyue, liuri, year, month, day)
		c.JSON(http.StatusOK, gin.H{"summary": summary})

	case "liu_nian_stars":
		year := req.Year
		if year == 0 {
			year = time.Now().Year()
		}
		liunian := svc.CalculateLiunian(chart, year)
		c.JSON(http.StatusOK, gin.H{
			"palaces": liunian.LiuNianStars,
			"year":    year,
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown period_type"})
	}
}

// Overlay handles the liunian overlay calculation.
func (h *ZiWeiPeriodHandler) Overlay(c *gin.Context) {
	if _, exists := c.Get("userID"); !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

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
	result := mapChartToResponse(liunian, birthChart.BirthMonth, birthChart.BirthHour)
	result["year"] = req.Year
	result["liu_nian_stars"] = liunian.LiuNianStars
	c.JSON(http.StatusOK, result)
}

func dayunDesc(palace string, startAge int) string {
	descs := map[string]string{
		"命宮": "个人运势与性格转变的关键十年",
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

// RegisterZiWeiPeriodRoutes registers ZiWei period and overlay routes.
func RegisterZiWeiPeriodRoutes(r gin.IRouter, svc *service.ZiWeiService, store ChartStore) {
	h := &ZiWeiPeriodHandler{Service: svc, Charts: store}
	r.POST("/ziwei/period", h.Period)
	r.POST("/ziwei/overlay", h.Overlay)
}
