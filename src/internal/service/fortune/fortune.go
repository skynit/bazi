package fortune

import (
	"fmt"
	"strings"
	"time"

	"github.com/6tail/tyme4go/tyme"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/data"
)

// FortuneEngine computes daily, weekly, and monthly fortunes by comparing
// a user's BaZi chart pillars against the query date's pillars.
type FortuneEngine struct {
	bazi *bazipkg.BaziService
}

// NewFortuneEngine creates a ready-to-use FortuneEngine.
func NewFortuneEngine() *FortuneEngine {
	return &FortuneEngine{
		bazi: &bazipkg.BaziService{},
	}
}

// DailyFortune is a single-day fortune result.
type DailyFortune struct {
	Date           string                      `json:"date"`
	DayPillar      model.Pillar                `json:"day_pillar"`
	Score          int                         `json:"score"`
	ScoreBreakdown model.FortuneScoreBreakdown `json:"score_breakdown"`
	ClashZodiac    string                      `json:"clash_zodiac"`
	ShengKe        ShengKeAnalysis             `json:"sheng_ke"`
	ElementImages  []model.ElementImage        `json:"element_images"`
	TodayElements  map[string]int              `json:"today_elements"`
	SeasonElement  model.SeasonElementEvidence `json:"season_element"`
	Rikuyo         *RikuyoResult               `json:"rikuyo"`
	Layers         model.FortuneLayerSet       `json:"fortune_layers"`
	LunarDate      string                      `json:"lunar_date"`
	WeekDay        string                      `json:"week_day"`
	ShengXiao      string                      `json:"sheng_xiao"`
	JiShen         string                      `json:"ji_shen"`
	XiongShen      string                      `json:"xiong_shen"`
	TaiShen        string                      `json:"tai_shen"`
	WuXing         string                      `json:"wu_xing"`
	PengZu         string                      `json:"peng_zu"`
	Gua            string                      `json:"gua"`
	JieQi          string                      `json:"jie_qi"`
}

// WeeklyFortune aggregates seven daily fortunes.
type WeeklyFortune struct {
	WeekStart               string               `json:"week_start"`
	DailyFortunes           []DailyFortune       `json:"daily_fortunes"`
	StructuralRelationIndex int                  `json:"structural_relation_index"`
	ElementTrend            []ElementTrendPoint  `json:"element_trend"`
	Summary                 model.FortuneSummary `json:"summary"`
}

// MonthlyFortune aggregates a full calendar month of daily fortunes.
type MonthlyFortune struct {
	Year                    int                  `json:"year"`
	Month                   int                  `json:"month"`
	DailyFortunes           []DailyFortune       `json:"daily_fortunes"`
	StructuralRelationIndex int                  `json:"structural_relation_index"`
	ElementTrend            []ElementTrendPoint  `json:"element_trend"`
	Summary                 model.FortuneSummary `json:"summary"`
}

// ElementTrendPoint is a single data point for element-trend charts.
type ElementTrendPoint struct {
	Date  string  `json:"date"`
	Score int     `json:"score"`
	Metal float64 `json:"metal"`
	Wood  float64 `json:"wood"`
	Water float64 `json:"water"`
	Fire  float64 `json:"fire"`
	Earth float64 `json:"earth"`
}

type ShengKeAnalysis = model.ShengKeAnalysis

var stemToElement = map[string]string{
	"甲": "木", "乙": "木",
	"丙": "火", "丁": "火",
	"戊": "土", "己": "土",
	"庚": "金", "辛": "金",
	"壬": "水", "癸": "水",
}

var elementGenerates = map[string]string{
	"木": "火", "火": "土", "土": "金", "金": "水", "水": "木",
}

var elementOvercomes = map[string]string{
	"木": "土", "土": "水", "水": "火", "火": "金", "金": "木",
}

var clashPairs = map[string]string{
	"子": "午", "丑": "未", "寅": "申", "卯": "酉",
	"辰": "戌", "巳": "亥",
	"午": "子", "未": "丑", "申": "寅", "酉": "卯",
	"戌": "辰", "亥": "巳",
}

var harmPairs = map[string]string{
	"子": "未", "丑": "午", "寅": "巳",
	"卯": "辰", "申": "亥", "酉": "戌",
	"未": "子", "午": "丑", "巳": "寅",
	"辰": "卯", "亥": "申", "戌": "酉",
}

var combinePairs = map[string]string{
	"子": "丑", "丑": "子",
	"寅": "亥", "亥": "寅",
	"卯": "戌", "戌": "卯",
	"辰": "酉", "酉": "辰",
	"巳": "申", "申": "巳",
	"午": "未", "未": "午",
}

// 经典依据：《三命通会》论地支相刑
// 三刑：寅巳申（无恩之刑）、丑戌未（恃势之刑）、子卯（无礼之刑）
// 注意：此处punishPairs仅做两两配对判断，完整三刑名称见rikuyo.go
var punishPairs = map[string]bool{
	"寅巳": true, "巳寅": true, "巳申": true, "申巳": true, "申寅": true, "寅申": true,
	"丑戌": true, "戌丑": true, "戌未": true, "未戌": true, "未丑": true, "丑未": true,
	"子卯": true, "卯子": true,
}

// 经典依据：《协纪辨方书》论地支相破
var breakPairs = map[string]string{
	"子": "酉", "酉": "子",
	"丑": "辰", "辰": "丑",
	"寅": "亥", "亥": "寅",
	"卯": "午", "午": "卯",
	"巳": "申", "申": "巳",
	"未": "戌", "戌": "未",
}

// 三合局：申子辰合水、寅午戌合火、亥卯未合木、巳酉丑合金
var sanHeGroups = [][]string{
	{"申", "子", "辰"},
	{"寅", "午", "戌"},
	{"亥", "卯", "未"},
	{"巳", "酉", "丑"},
}

// 三会局：寅卯辰会木、巳午未会火、申酉戌会金、亥子丑会水
var sanHuiGroups = [][]string{
	{"寅", "卯", "辰"},
	{"巳", "午", "未"},
	{"申", "酉", "戌"},
	{"亥", "子", "丑"},
}

// CalculateDaily computes the fortune for a single date.
func (e *FortuneEngine) CalculateDaily(userChart *bazipkg.BaziResult, queryDate time.Time, birthYear int) *DailyFortune {
	qYear := queryDate.Year()
	qMonth := int(queryDate.Month())
	qDay := queryDate.Day()

	dayPillar, err := getDayPillar(qYear, qMonth, qDay)
	if err != nil {
		return e.fallbackDaily(queryDate)
	}

	userDayStem := userChart.DayPillar.Gan
	userDayBranch := userChart.DayPillar.Zhi

	stemRel := stemRelation(userDayStem, dayPillar.Gan)
	branchRel := branchRelation(userDayBranch, dayPillar.Zhi)

	clashZodiac := data.NewAuspiciousData().GetClashZodiac(dayPillar.Zhi)

	// compute today's five-element distribution
	ec, _ := getDayEightChar(qYear, qMonth, qDay)
	todayElements := calcFiveElements(ec)
	// Filter to only 金木水火土
	filtered := map[string]int{
		"金": todayElements["金"], "木": todayElements["木"],
		"水": todayElements["水"], "火": todayElements["火"], "土": todayElements["土"],
	}

	// 日课推算
	rikuyo := CalcRikuyo(userChart, queryDate)
	scoreBreakdown := buildScorePipeline(userChart, stemRel, branchRel, userDayStem, dayPillar.Gan)
	score := scoreBreakdown.FinalScore

	shengKe := ShengKeAnalysis{
		DayStemRelation:   stemRelLabel(stemRel, userDayStem, dayPillar.Gan),
		DayBranchRelation: branchRelLabel(branchRel),
		Summary:           shengKeSummary(stemRel, branchRel),
	}

	// 黄历数据（通过 tyme4go 获取）
	almanac := getAlmanacData(qYear, qMonth, qDay)

	return &DailyFortune{
		Date:           queryDate.Format("2006-01-02"),
		DayPillar:      dayPillar,
		Score:          score,
		ScoreBreakdown: scoreBreakdown,
		ClashZodiac:    clashZodiac,
		ShengKe:        shengKe,
		ElementImages:  fixedElementImages(),
		TodayElements:  filtered,
		SeasonElement:  observeSeasonElement(userChart.DayPillar.Gan, ec.GetMonth().GetEarthBranch().GetName()),
		Rikuyo:         rikuyo,
		Layers:         BuildFortuneLayers(userChart, queryDate, birthYear),
		LunarDate:      almanac.LunarDate,
		WeekDay:        almanac.WeekDay,
		ShengXiao:      almanac.ShengXiao,
		JiShen:         almanac.JiShen,
		XiongShen:      almanac.XiongShen,
		TaiShen:        almanac.TaiShen,
		WuXing:         almanac.WuXing,
		PengZu:         almanac.PengZu,
		Gua:            almanac.Gua,
		JieQi:          almanac.JieQi,
	}
}

func observeSeasonElement(dayGan, monthZhi string) model.SeasonElementEvidence {
	evidence := model.SeasonElementEvidence{
		RuleID:               "fortune.season-element.month-branch-v1",
		ReferenceStem:        dayGan,
		ReferenceElement:     stemToElement[dayGan],
		QueryMonthBranch:     monthZhi,
		Basis:                "reference_day_stem_element_and_query_month_branch",
		Status:               "unavailable",
		InterpretationStatus: "not_adjudicated",
	}
	if evidence.ReferenceElement == "" || data.ZhiIndex(monthZhi) < 0 {
		return evidence
	}
	evidence.Season = monthZhiToSeason(monthZhi)
	if evidence.Season == "" {
		return evidence
	}
	evidence.Status = "observed"
	return evidence
}

// CalculateWeekly computes fortunes for 7 consecutive days starting from weekStart.
func (e *FortuneEngine) CalculateWeekly(userChart *bazipkg.BaziResult, weekStart time.Time, birthYear int) *WeeklyFortune {
	weekStart = toDateStart(weekStart)
	fortunes := make([]DailyFortune, 7)
	trends := make([]ElementTrendPoint, 7)
	totalScore := 0

	for i := 0; i < 7; i++ {
		day := weekStart.AddDate(0, 0, i)
		df := e.CalculateDaily(userChart, day, birthYear)
		fortunes[i] = *df
		totalScore += df.Score
		trends[i] = e.elementTrend(day, df.Score)
	}

	avg := totalScore / 7

	return &WeeklyFortune{
		WeekStart:               weekStart.Format("2006-01-02"),
		DailyFortunes:           fortunes,
		StructuralRelationIndex: avg,
		ElementTrend:            trends,
		Summary:                 computeSummary(fortunes),
	}
}

// CalculateMonthly computes fortunes for every day in the given year/month.
func (e *FortuneEngine) CalculateMonthly(userChart *bazipkg.BaziResult, year, month, birthYear int) *MonthlyFortune {
	days := daysInMonth(year, month)
	fortunes := make([]DailyFortune, 0, days)
	trends := make([]ElementTrendPoint, 0, days)
	totalScore := 0

	for d := 1; d <= days; d++ {
		date := time.Date(year, time.Month(month), d, 12, 0, 0, 0, time.UTC)
		df := e.CalculateDaily(userChart, date, birthYear)
		fortunes = append(fortunes, *df)
		totalScore += df.Score
		trends = append(trends, e.elementTrend(date, df.Score))
	}

	avg := totalScore / days

	return &MonthlyFortune{
		Year:                    year,
		Month:                   month,
		DailyFortunes:           fortunes,
		StructuralRelationIndex: avg,
		ElementTrend:            trends,
		Summary:                 computeSummary(fortunes),
	}
}

func getDayPillar(year, month, day int) (model.Pillar, error) {
	ec, err := getDayEightChar(year, month, day)
	if err != nil {
		return model.Pillar{}, err
	}
	return pillarFromSixtyCycle(ec.GetDay()), nil
}

func getDayEightChar(year, month, day int) (*tyme.EightChar, error) {
	st, err := tyme.SolarTime{}.FromYmdHms(year, month, day, 12, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid date %d-%02d-%02d: %w", year, month, day, err)
	}
	ec := st.GetLunarHour().GetEightChar()
	return &ec, nil
}

// stemRelation returns the ten-god-style relation between user stem and query stem.
// 经典依据：《滴天髓》"印绶生身喜其有气；食伤泄秀须看身强弱"
// 区分"生我/我生"和"克我/我克"方向，评分意义完全不同。
func stemRelation(userStem, queryStem string) string {
	ue := stemToElement[userStem]
	qe := stemToElement[queryStem]
	if ue == "" || qe == "" {
		return "unknown"
	}
	if ue == qe {
		return "same"
	}
	if elementGenerates[qe] == ue {
		return "shengWo" // 生我（印星）
	}
	if elementGenerates[ue] == qe {
		return "woSheng" // 我生（食伤）
	}
	if elementOvercomes[qe] == ue {
		return "keWo" // 克我（官杀）
	}
	if elementOvercomes[ue] == qe {
		return "woKe" // 我克（财星）
	}
	return "unknown"
}

func branchRelation(userBranch, queryBranch string) string {
	if data.ZhiIndex(userBranch) < 0 || data.ZhiIndex(queryBranch) < 0 {
		return "unknown"
	}
	if userBranch == queryBranch {
		return "same"
	}
	if clashPairs[userBranch] == queryBranch {
		return "clash"
	}
	if harmPairs[userBranch] == queryBranch {
		return "harm"
	}
	if combinePairs[userBranch] == queryBranch {
		return "combine"
	}
	// 经典依据：《三命通会》论地支相刑
	if punishPairs[userBranch+queryBranch] {
		return "punish"
	}
	// 经典依据：《协纪辨方书》论地支相破
	if breakPairs[userBranch] == queryBranch {
		return "break"
	}
	// 日支与流日支只有两支，不能宣称形成完整三合或三会。
	// 三合两支：含中神为半合，首尾两支为拱合。
	if relation := partialSanHeRelation(userBranch, queryBranch); relation != "" {
		return relation
	}
	// 三会任意两支只记半会。
	if isInSameGroup(sanHuiGroups, userBranch, queryBranch) {
		return "banHui"
	}
	return "neutral"
}

func partialSanHeRelation(a, b string) string {
	for _, group := range sanHeGroups {
		indexA, indexB := -1, -1
		for i, branch := range group {
			if branch == a {
				indexA = i
			}
			if branch == b {
				indexB = i
			}
		}
		if indexA < 0 || indexB < 0 || indexA == indexB {
			continue
		}
		if indexA == 1 || indexB == 1 {
			return "banHe"
		}
		return "gongHe"
	}
	return ""
}

// isInSameGroup checks if two branches belong to the same group.
func isInSameGroup(groups [][]string, a, b string) bool {
	if a == b {
		return false
	}
	for _, g := range groups {
		hasA, hasB := false, false
		for _, z := range g {
			if z == a {
				hasA = true
			}
			if z == b {
				hasB = true
			}
		}
		if hasA && hasB {
			return true
		}
	}
	return false
}

// calcScore computes a 0-100 fortune score based on stem and branch relations.
// 经典依据：《三命通会》论十神关系，区分生我/我生/克我/我克方向。
// 六冲远比六害严重（《协纪辨方书》"冲者冲散之义；害者暗害之义"）。
// 如果日主天干与流日天干形成五合（甲己合、乙庚合、丙辛合、丁壬合、戊癸合），额外加分。
func calcScore(stemRel, branchRel string, userGan, dayGan string) int {
	score, _ := relationScoreStage(stemRel, branchRel, userGan, dayGan)
	return score
}

func stemRelLabel(rel, userStem, queryStem string) string {
	switch rel {
	case "same":
		return "比和"
	case "shengWo":
		return "生我"
	case "woSheng":
		return "我生"
	case "keWo":
		return "克我"
	case "woKe":
		return "我克"
	default:
		return "无特殊关系"
	}
}

func branchRelLabel(rel string) string {
	switch rel {
	case "same":
		return "同支"
	case "clash":
		return "六冲"
	case "harm":
		return "六害"
	case "combine":
		return "六合"
	case "punish":
		return "相刑"
	case "break":
		return "相破"
	case "banHe":
		return "半合"
	case "gongHe":
		return "拱合"
	case "banHui":
		return "半会"
	case "sanHe":
		return "三合"
	case "sanHui":
		return "三会"
	default:
		return "平和"
	}
}

func shengKeSummary(stemRel, branchRel string) string {
	var parts []string
	if stemRel != "unknown" {
		parts = append(parts, fmt.Sprintf("日干关系: %s", stemRelLabel(stemRel, "", "")))
	}
	if branchRel != "neutral" && branchRel != "unknown" {
		parts = append(parts, fmt.Sprintf("日支关系: %s", branchRelLabel(branchRel)))
	}

	base := strings.Join(parts, "；")
	if base == "" {
		base = "干支平和"
	}

	return base + "。仅记录干支结构，现实结果未裁决。"
}

func (e *FortuneEngine) elementTrend(date time.Time, score int) ElementTrendPoint {
	ec, err := getDayEightChar(date.Year(), int(date.Month()), date.Day())
	if err != nil {
		return ElementTrendPoint{Date: date.Format("2006-01-02"), Score: score}
	}

	elements := calcFiveElements(ec)
	total := 0
	for _, v := range elements {
		total += v
	}

	pt := ElementTrendPoint{
		Date:  date.Format("2006-01-02"),
		Score: score,
	}
	if total > 0 {
		pt.Metal = float64(elements["金"]) / float64(total) * 100
		pt.Wood = float64(elements["木"]) / float64(total) * 100
		pt.Water = float64(elements["水"]) / float64(total) * 100
		pt.Fire = float64(elements["火"]) / float64(total) * 100
		pt.Earth = float64(elements["土"]) / float64(total) * 100
	}
	return pt
}

func fixedElementImages() []model.ElementImage {
	return []model.ElementImage{
		{Element: "金", ImageURL: "/images/elements/metal.png", Description: "金"},
		{Element: "木", ImageURL: "/images/elements/wood.png", Description: "木"},
		{Element: "水", ImageURL: "/images/elements/water.png", Description: "水"},
		{Element: "火", ImageURL: "/images/elements/fire.png", Description: "火"},
		{Element: "土", ImageURL: "/images/elements/earth.png", Description: "土"},
	}
}

func (e *FortuneEngine) fallbackDaily(queryDate time.Time) *DailyFortune {
	breakdown := model.FortuneScoreBreakdown{
		PipelineVersion:      FortuneScorePipelineVersion,
		ScoreKind:            "structural_relation_index",
		EvidenceBasis:        "empirical",
		ValidationStatus:     "not_validated",
		InterpretationStatus: "not_adjudicated",
		IsOutcomeProbability: false,
		BaseScore:            50,
		RelationScore:        50,
		FinalScore:           50,
		EvidenceCompleteness: 20,
		SupportingEvidence:   []model.ScoreEvidence{},
		CounterEvidence:      []model.ScoreEvidence{},
	}
	return &DailyFortune{
		Date:           queryDate.Format("2006-01-02"),
		DayPillar:      model.Pillar{Gan: "?", Zhi: "?"},
		Score:          50,
		ScoreBreakdown: breakdown,
		ElementImages:  fixedElementImages(),
		SeasonElement:  observeSeasonElement("", ""),
		ShengKe: ShengKeAnalysis{
			Summary: "无法计算日柱，使用默认运势。",
		},
	}
}

func toDateStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// 天干五合映射（经典：《三命通会》天干五合论）
// 甲己合、乙庚合、丙辛合、丁壬合、戊癸合
var ganHePairs = map[string]string{
	"甲": "己", "己": "甲",
	"乙": "庚", "庚": "乙",
	"丙": "辛", "辛": "丙",
	"丁": "壬", "壬": "丁",
	"戊": "癸", "癸": "戊",
}

// 五合化气五行
var ganHeHuaMap = map[string]string{
	"甲己": "土",
	"乙庚": "金",
	"丙辛": "水",
	"丁壬": "木",
	"戊癸": "火",
}

// isGanHe 判断两个天干是否形成五合关系
func isGanHe(gan1, gan2 string) bool {
	return ganHePairs[gan1] == gan2
}

// ganHeHuaElem 返回五合化气五行
func ganHeHuaElem(gan1, gan2 string) string {
	// 标准化顺序：按甲乙丙丁戊在前
	pair := gan1 + gan2
	if hua, ok := ganHeHuaMap[pair]; ok {
		return hua
	}
	// 尝试反序
	pair = gan2 + gan1
	return ganHeHuaMap[pair]
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// pillarFromSixtyCycle converts a tyme.SixtyCycle to a model.Pillar.
func pillarFromSixtyCycle(sc tyme.SixtyCycle) model.Pillar {
	return model.Pillar{
		Gan: sc.GetHeavenStem().GetName(),
		Zhi: sc.GetEarthBranch().GetName(),
	}
}

// calcFiveElements computes the five-element distribution from an EightChar.
func calcFiveElements(ec *tyme.EightChar) map[string]int {
	scores := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}

	pillars := [](func() tyme.SixtyCycle){
		ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour,
	}
	for _, fn := range pillars {
		sc := fn()
		// heavenly stem: 5 points
		elem := sc.GetHeavenStem().GetElement().GetName()
		scores[elem] += 5

		// earthly branch hidden stems: main=3, middle=2, residual=1
		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			weight := 1
			switch hhs.GetType() {
			case tyme.MAIN:
				weight = 3
			case tyme.MIDDLE:
				weight = 2
			}
			elem := hhs.GetHeavenStem().GetElement().GetName()
			scores[elem] += weight
		}
	}
	return scores
}

type AlmanacData struct {
	LunarDate string
	WeekDay   string
	ShengXiao string
	JiShen    string
	XiongShen string
	TaiShen   string
	WuXing    string
	PengZu    string
	Gua       string
	JieQi     string
}

func getAlmanacData(year, month, day int) AlmanacData {
	solar, err := tyme.SolarDay{}.FromYmd(year, month, day)
	if err != nil {
		return AlmanacData{}
	}

	lunar := solar.GetLunarDay()
	scd := solar.GetSixtyCycleDay()

	var lunarStr string
	lMonth := lunar.GetLunarMonth()
	lYear := lMonth.GetLunarYear()
	lunarStr = lYear.GetSixtyCycle().GetHeavenStem().GetName() +
		lYear.GetSixtyCycle().GetEarthBranch().GetName() + "年" +
		lMonth.GetName() + lunar.GetName()

	weekName := "星期" + solar.GetWeek().GetName()

	zodiac := tyme.Zodiac{}.FromIndex(scd.GetYear().GetEarthBranch().GetIndex())
	shengXiao := zodiac.GetName()

	var jiShens, xiongShens []string
	gods, err := scd.GetGods()
	if err == nil {
		for _, g := range gods {
			if g.GetLuck().GetIndex() == 0 {
				jiShens = append(jiShens, g.GetName())
			} else {
				xiongShens = append(xiongShens, g.GetName())
			}
		}
	}

	taiShen := scd.GetFetusDay().GetName()

	pengZu := tyme.PengZu{}.FromSixtyCycle(scd.GetSixtyCycle()).GetName()

	wuXing := scd.GetSixtyCycle().GetHeavenStem().GetElement().GetName() +
		scd.GetSixtyCycle().GetEarthBranch().GetElement().GetName()

	nineStar := scd.GetNineStar()
	gua := nineStar.GetName() + nineStar.GetColor() + nineStar.GetElement().GetName()

	jieQi := solar.GetTerm().GetName()

	return AlmanacData{
		LunarDate: lunarStr,
		WeekDay:   weekName,
		ShengXiao: shengXiao,
		JiShen:    strings.Join(jiShens, " "),
		XiongShen: strings.Join(xiongShens, " "),
		TaiShen:   taiShen,
		WuXing:    wuXing,
		PengZu:    pengZu,
		Gua:       gua,
		JieQi:     jieQi,
	}
}
