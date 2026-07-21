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
	birth, err := resolveStoredZiWeiBirth(birthChart)
	if err != nil {
		return nil, nil, err
	}
	// Try cached result first
	if birthChart.ZiWeiComputed && len(birthChart.ZiWeiResult) > 0 {
		var cached ziwei.ZiWeiChart
		if err := json.Unmarshal(birthChart.ZiWeiResult, &cached); err == nil && h.Service.ChartMatchesInputProfile(
			&cached, "",
			birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender,
		) {
			if err := h.Service.AttachBirthData(&cached, birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender); err != nil {
				return nil, nil, fmt.Errorf("attach birth data: %w", err)
			}
			return &cached, birthChart, nil
		}
	}
	chart, err := h.Service.CalculateChart(birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Gender)
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

	chart, _, err := h.getChart(req.ChartID, userID.(uint))
	if err != nil {
		respondError(c, http.StatusNotFound, ErrCodeNotFound, err.Error())
		return
	}

	svc := h.Service

	switch req.PeriodType {
	case "dayun":
		year, month, day, dateErr := resolveZiWeiPeriodSolarDate(req.Year, req.Month, req.Day)
		if dateErr != nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, dateErr.Error())
			return
		}
		targetDate := time.Date(year, time.Month(month), day, 12, 0, 0, 0, time.Local)
		nominalAge := ziWeiNominalAgeAt(chart, targetDate)
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
		resp := gin.H{
			"periods":         enriched,
			"target_date":     fmt.Sprintf("%04d-%02d-%02d", year, month, day),
			"nominal_age":     nominalAge,
			"age_basis":       "target_lunar_year_minus_birth_lunar_year_plus_one",
			"boundary_policy": ziwei.ZiWeiHoroscopeBoundaryNormal,
		}
		resp["analysis"] = ziwei.BuildDayunAnalysis(chart, dayun, nominalAge)
		respondJSON(c, http.StatusOK, resp)

	case "liunian":
		year := req.Year
		if year == 0 {
			year = currentZiWeiLunarYearLabel()
		}
		liunian := svc.CalculateLiunian(chart, year)
		if liunian == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian year")
			return
		}
		resp := mapChartToResponse(liunian, svc)
		resp["year"] = year
		resp["description"] = fmt.Sprintf("%d年流年星曜分布，各宫依次更换", year)
		respondJSON(c, http.StatusOK, gin.H{
			"periods":    []gin.H{resp},
			"analysis":   ziwei.BuildLiunianAnalysis(chart, liunian, year),
			"year":       year,
			"period_key": "liunian",
		})

	case "liuyue":
		year, month, day, dateErr := resolveZiWeiPeriodSolarDate(req.Year, req.Month, req.Day)
		if dateErr != nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, dateErr.Error())
			return
		}
		liuyue := svc.CalculateLiuyueForDate(chart, year, month, day)
		if liuyue == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuyue year or month")
			return
		}
		resp := mapChartToResponse(liuyue, svc)
		resp["year"] = year
		resp["month"] = month
		resp["day"] = day
		resp["description"] = fmt.Sprintf("%d年%d月%d日所在农历月的流月星曜分布", year, month, day)
		respondJSON(c, http.StatusOK, gin.H{
			"periods":    []gin.H{resp},
			"analysis":   ziwei.BuildLiuyueAnalysis(chart, liuyue, year, month, day),
			"year":       year,
			"month":      month,
			"day":        day,
			"period_key": "liuyue",
		})

	case "liuri":
		year, month, day, dateErr := resolveZiWeiPeriodSolarDate(req.Year, req.Month, req.Day)
		if dateErr != nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, dateErr.Error())
			return
		}
		liuri := svc.CalculateLiuriForDate(chart, year, month, day)
		if liuri == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuri solar date")
			return
		}
		resp := mapChartToResponse(liuri, svc)
		resp["year"] = year
		resp["month"] = month
		resp["day"] = day
		resp["description"] = fmt.Sprintf("%d年%d月%d日流日星曜分布", year, month, day)
		respondJSON(c, http.StatusOK, gin.H{
			"periods":    []gin.H{resp},
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
			"description": "宫干四化直接飞行：展示化曜的来源宫干、目标宫位与同宫/跨宫结构",
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
			year = currentZiWeiLunarYearLabel()
		}
		liunian := svc.CalculateLiunian(chart, year)
		if liunian == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian year")
			return
		}
		interp := ziwei.NewPeriodInterpreterFromChart(chart)
		if interp == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid natal chart contract")
			return
		}
		result := interp.AnalyzeLiunian(liunian, year)
		if result == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian derivation contract")
			return
		}
		respondJSON(c, http.StatusOK, gin.H{"periods": []*ziwei.LiunianResult{result}})

	case "liuyue_interpretation":
		year, month, day, dateErr := resolveZiWeiPeriodSolarDate(req.Year, req.Month, req.Day)
		if dateErr != nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, dateErr.Error())
			return
		}
		liuyue := svc.CalculateLiuyueForDate(chart, year, month, day)
		if liuyue == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuyue year or month")
			return
		}
		interp := ziwei.NewPeriodInterpreterFromChart(chart)
		if interp == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid natal chart contract")
			return
		}
		result := interp.AnalyzeLiuyue(liuyue, year, month, day)
		if result == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuyue derivation contract")
			return
		}
		respondJSON(c, http.StatusOK, gin.H{"periods": []*ziwei.LiuyueResult{result}})

	case "liuri_interpretation":
		year, month, day, dateErr := resolveZiWeiPeriodSolarDate(req.Year, req.Month, req.Day)
		if dateErr != nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, dateErr.Error())
			return
		}
		liuri := svc.CalculateLiuriForDate(chart, year, month, day)
		if liuri == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuri solar date")
			return
		}
		interp := ziwei.NewPeriodInterpreterFromChart(chart)
		if interp == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid natal chart contract")
			return
		}
		result := interp.AnalyzeLiuri(liuri, year, month, day)
		if result == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liuri derivation contract")
			return
		}
		respondJSON(c, http.StatusOK, gin.H{"periods": []*ziwei.LiuriResult{result}})

	case "period_summary":
		year, month, day, dateErr := resolveZiWeiPeriodSolarDate(req.Year, req.Month, req.Day)
		if dateErr != nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, dateErr.Error())
			return
		}
		lunarYearLabel, dateErr := ziwei.LunarYearLabelForSolarDate(year, month, day)
		if dateErr != nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, dateErr.Error())
			return
		}
		liunian := svc.CalculateLiunian(chart, lunarYearLabel)
		liuyue := svc.CalculateLiuyueForDate(chart, year, month, day)
		liuri := svc.CalculateLiuriForDate(chart, year, month, day)
		if liunian == nil || liuyue == nil || liuri == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid period summary solar date")
			return
		}
		interp := ziwei.NewPeriodInterpreterFromChart(chart)
		if interp == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid natal chart contract")
			return
		}
		summary := interp.SummarizeAll(liunian, liuyue, liuri, year, month, day)
		if summary == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "inconsistent period summary context")
			return
		}
		respondJSON(c, http.StatusOK, gin.H{"summary": summary})

	case "liu_nian_stars":
		year := req.Year
		if year == 0 {
			year = currentZiWeiLunarYearLabel()
		}
		liunian := svc.CalculateLiunian(chart, year)
		if liunian == nil {
			respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian year")
			return
		}
		respondJSON(c, http.StatusOK, gin.H{
			"palaces": liunian.LiuNianStars,
			"year":    year,
		})

	default:
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "unknown period_type")
	}
}

func resolveZiWeiPeriodSolarDate(year, month, day int) (int, int, int, error) {
	if year == 0 && month == 0 && day == 0 {
		now := time.Now()
		return now.Year(), int(now.Month()), now.Day(), nil
	}
	if year == 0 || month == 0 || day == 0 {
		return 0, 0, 0, fmt.Errorf("year, month, and day must be provided together")
	}
	if year < 1 || year > 9999 || month < 1 || month > 12 || day < 1 || day > 31 {
		return 0, 0, 0, fmt.Errorf("invalid solar date")
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if date.Year() != year || int(date.Month()) != month || date.Day() != day {
		return 0, 0, 0, fmt.Errorf("invalid solar date")
	}
	return year, month, day, nil
}

func currentZiWeiLunarYearLabel() int {
	now := time.Now()
	year, err := ziwei.LunarYearLabelForSolarDate(now.Year(), int(now.Month()), now.Day())
	if err != nil {
		return now.Year()
	}
	return year
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
		year = currentZiWeiLunarYearLabel()
	}
	liunian := h.Service.CalculateLiunian(chart, year)
	if liunian == nil {
		respondError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid liunian year")
		return
	}
	result := mapChartToResponse(liunian, h.Service)
	result["year"] = year
	result["liu_nian_stars"] = liunian.LiuNianStars
	result["overlay_analysis"] = h.Service.AnalyzeLiunianOverlay(chart, liunian, year)
	respondJSON(c, http.StatusOK, result)
}

func dayunDesc(palace string, startAge int) string {
	descs := map[string]string{
		"命宮":  "命宫主题在此大限的结构位置",
		"兄弟宮": "同辈、协作与资源分配主题的结构位置",
		"夫妻宮": "亲密关系、承诺与协商主题的结构位置",
		"子女宮": "子女、下属与创造输出主题的结构位置",
		"財帛宮": "现金流与资源配置主题的结构位置；不构成财务建议",
		"疾厄宮": "传统疾厄宫主题被触发；仅展示宫位与星曜结构，不作个体身体状态推断",
		"遷移宮": "外部环境、出行与社会形象主题的结构位置",
		"僕役宮": "朋友、团队与合作对象主题的结构位置",
		"官祿宮": "职业、责任与组织角色主题的结构位置；不构成职业建议",
		"田宅宮": "家庭、居住与不动产主题的结构位置；不构成交易建议",
		"福德宮": "精神生活、兴趣与内在节奏主题的结构位置",
		"父母宮": "长辈、制度与支持来源主题的结构位置",
	}
	if d, ok := descs[palace]; ok {
		return d
	}
	return fmt.Sprintf("%s%s-%d岁大限", palace, palace, startAge)
}

func ziWeiNominalAgeAt(chart *ziwei.ZiWeiChart, target time.Time) int {
	if chart == nil {
		return 0
	}
	birthLunarYear, err := ziwei.LunarYearLabelForSolarDate(
		chart.CalculationInput.Year,
		chart.CalculationInput.Month,
		chart.CalculationInput.Day,
	)
	if err != nil {
		return 0
	}
	targetLunarYear, err := ziwei.LunarYearLabelForSolarDate(
		target.Year(), int(target.Month()), target.Day(),
	)
	if err != nil || targetLunarYear < birthLunarYear {
		return 0
	}
	return targetLunarYear - birthLunarYear + 1
}

// RegisterZiWeiPeriodRoutes registers ZiWei period and overlay routes.
func RegisterZiWeiPeriodRoutes(r gin.IRouter, svc *ziwei.ZiWeiService, store ChartStore) {
	h := &ZiWeiPeriodHandler{Service: svc, Charts: store}
	r.POST("/ziwei/period", h.Period)
	r.POST("/ziwei/overlay", h.Overlay)
}
