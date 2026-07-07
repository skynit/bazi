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
	bazi       *bazipkg.BaziService
	auspicious *data.AuspiciousData
}

// NewFortuneEngine creates a ready-to-use FortuneEngine.
func NewFortuneEngine() *FortuneEngine {
	return &FortuneEngine{
		bazi:       &bazipkg.BaziService{},
		auspicious: data.NewAuspiciousData(),
	}
}

// DailyFortune is a single-day fortune result.
type DailyFortune struct {
	Date                string                `json:"date"`
	DayPillar           model.Pillar          `json:"day_pillar"`
	Score               int                   `json:"score"`
	LuckyColor          string                `json:"lucky_color"`
	LuckyNumbers        []int                 `json:"lucky_numbers"`
	WealthDir           string                `json:"wealth_dir"`
	Guide               *model.FortuneGuide   `json:"guide,omitempty"`
	ClashZodiac         string                `json:"clash_zodiac"`
	AuspiciousHours     []string              `json:"auspicious_hours"`
	Yi                  []model.YiJiItem      `json:"yi"`
	Ji                  []model.YiJiItem      `json:"ji"`
	ShengKe             ShengKeAnalysis       `json:"sheng_ke"`
	ElementImages       []model.ElementImage  `json:"element_images"`
	TodayElements       map[string]int        `json:"today_elements"`
	SeasonElementAdvice string                `json:"season_element_advice"`
	FlowImpact          string                `json:"flow_impact"`
	Rikuyo              *RikuyoResult         `json:"rikuyo"`
	Layers              model.FortuneLayerSet `json:"fortune_layers"`
	LunarDate           string                `json:"lunar_date"`
	WeekDay             string                `json:"week_day"`
	ShengXiao           string                `json:"sheng_xiao"`
	JiShen              string                `json:"ji_shen"`
	XiongShen           string                `json:"xiong_shen"`
	TaiShen             string                `json:"tai_shen"`
	WuXing              string                `json:"wu_xing"`
	PengZu              string                `json:"peng_zu"`
	Gua                 string                `json:"gua"`
	JieQi               string                `json:"jie_qi"`
}

// WeeklyFortune aggregates seven daily fortunes.
type WeeklyFortune struct {
	WeekStart      string               `json:"week_start"`
	DailyFortunes  []DailyFortune       `json:"daily_fortunes"`
	OverallSummary string               `json:"overall_summary"`
	WeeklyScore    int                  `json:"weekly_score"`
	ElementTrend   []ElementTrendPoint  `json:"element_trend"`
	Summary        model.FortuneSummary `json:"summary"`
}

// MonthlyFortune aggregates a full calendar month of daily fortunes.
type MonthlyFortune struct {
	Year           int                  `json:"year"`
	Month          int                  `json:"month"`
	DailyFortunes  []DailyFortune       `json:"daily_fortunes"`
	OverallSummary string               `json:"overall_summary"`
	MonthlyScore   int                  `json:"monthly_score"`
	ElementTrend   []ElementTrendPoint  `json:"element_trend"`
	Summary        model.FortuneSummary `json:"summary"`
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
	"寅巳": true, "巳申": true, "申寅": true,
	"丑戌": true, "戌未": true, "未丑": true,
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

// 协纪辨方书·宜事（扩展版）
var allYiActs = []string{
	"嫁娶", "出行", "入学", "开市", "会友", "祭祀", "求医", "纳采",
	"裁衣", "竖柱", "祈福", "求嗣", "牧养", "宴饮", "修造", "移徙",
	"安床", "开光", "针灸", "解除", "栽种", "纳财", "交易", "创作",
}

// 协纪辨方书·忌事（扩展版）
var allJiActs = []string{
	"动土", "安葬", "行丧", "开渠", "伐木", "破土", "词讼", "远行",
	"乘船", "渡水", "造屋", "上梁", "掘井", "安门", "作灶",
}

// 十神维度宜事推荐：根据日干与流日天干的十神关系，推荐对应活动
var tenGodYiActs = map[string][]string{
	"same":    {"会友", "结盟", "合作"},
	"shengWo": {"祈福", "入学", "修造"},
	"woSheng": {"宴饮", "创作", "开光"},
	"keWo":    {"祭祀", "求医", "解除"},
	"woKe":    {"开市", "交易", "纳财"},
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

	// 获取用户喜用神（格局优先）
	like, _, _ := getEffectiveFavor(userChart)

	score := calcScore(stemRel, branchRel, userDayStem, dayPillar.Gan)
	luckyColor := e.auspicious.GetLuckyColor(like)
	luckyNumbers := e.auspicious.GetLuckyNumbers(like)
	wealthDir := e.auspicious.GetWealthDirection(dayPillar.Gan)
	clashZodiac := e.auspicious.GetClashZodiac(dayPillar.Zhi)
	auspHours := e.auspicious.GetAuspiciousHours(dayPillar.Zhi)

	yi, ji := pickYiJi(score, stemRel)

	shengKe := ShengKeAnalysis{
		DayStemRelation:   stemRelLabel(stemRel, userDayStem, dayPillar.Gan),
		DayBranchRelation: branchRelLabel(branchRel),
		Summary:           shengKeSummary(stemRel, branchRel, score),
	}

	// compute today's five-element distribution
	ec, _ := getDayEightChar(qYear, qMonth, qDay)
	todayElements := calcFiveElements(ec)
	// Filter to only 金木水火土
	filtered := map[string]int{
		"金": todayElements["金"], "木": todayElements["木"],
		"水": todayElements["水"], "火": todayElements["火"], "土": todayElements["土"],
	}

	// 日课推算
	rikuyo := CalcRikuyo(userChart, queryDate, birthYear)
	guide := BuildFortuneGuide(userChart, dayPillar, score, stemRel, branchRel, luckyColor, luckyNumbers, wealthDir, auspHours, yi, ji, rikuyo)

	// 黄历数据（通过 tyme4go 获取）
	almanac := getAlmanacData(qYear, qMonth, qDay)

	return &DailyFortune{
		Date:                queryDate.Format("2006-01-02"),
		DayPillar:           dayPillar,
		Score:               score,
		LuckyColor:          luckyColor,
		LuckyNumbers:        luckyNumbers,
		WealthDir:           wealthDir,
		Guide:               guide,
		ClashZodiac:         clashZodiac,
		AuspiciousHours:     auspHours,
		Yi:                  yi,
		Ji:                  ji,
		ShengKe:             shengKe,
		ElementImages:       fixedElementImages(),
		TodayElements:       filtered,
		SeasonElementAdvice: getSeasonElementAdvice(userChart.DayPillar.Gan, ec.GetMonth().GetEarthBranch().GetName()),
		FlowImpact:          analyzeDayFlowImpact(userChart, dayPillar),
		Rikuyo:              rikuyo,
		Layers:              BuildFortuneLayers(userChart, queryDate, birthYear),
		LunarDate:           almanac.LunarDate,
		WeekDay:             almanac.WeekDay,
		ShengXiao:           almanac.ShengXiao,
		JiShen:              almanac.JiShen,
		XiongShen:           almanac.XiongShen,
		TaiShen:             almanac.TaiShen,
		WuXing:              almanac.WuXing,
		PengZu:              almanac.PengZu,
		Gua:                 almanac.Gua,
		JieQi:               almanac.JieQi,
	}
}

func getSeasonElementAdvice(dayGan string, monthZhi string) string {
	dayElem := stemToElement[dayGan]
	season := monthZhiToSeason(monthZhi)
	key := dayElem
	entries, ok := data.WuxingSeasonKnowledge[key]
	if !ok {
		return ""
	}
	for _, e := range entries {
		if e.Season == season {
			advice := e.Favor
			if e.Taboo != "" {
				advice += "。忌" + e.Taboo
			}
			if e.Judgment != "" {
				advice += "。断语：" + e.Judgment
			}
			return advice
		}
	}
	return ""
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
		WeekStart:      weekStart.Format("2006-01-02"),
		DailyFortunes:  fortunes,
		OverallSummary: periodSummary(avg, "本周"),
		WeeklyScore:    avg,
		ElementTrend:   trends,
		Summary:        computeSummary(fortunes),
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
		Year:           year,
		Month:          month,
		DailyFortunes:  fortunes,
		OverallSummary: periodSummary(avg, "本月"),
		MonthlyScore:   avg,
		ElementTrend:   trends,
		Summary:        computeSummary(fortunes),
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
	// 三合局检测
	if isInSameGroup(sanHeGroups, userBranch, queryBranch) {
		return "sanHe"
	}
	// 三会局检测
	if isInSameGroup(sanHuiGroups, userBranch, queryBranch) {
		return "sanHui"
	}
	return "neutral"
}

// isInSameGroup checks if two branches belong to the same group.
func isInSameGroup(groups [][]string, a, b string) bool {
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
	score := 50

	switch stemRel {
	case "same":
		score += 10 // 比和，助力适中
	case "shengWo":
		score += 18 // 生我（印星），得助有力
	case "woSheng":
		score += 5 // 我生（食伤），泄气轻微正面
	case "keWo":
		score -= 18 // 克我（官杀），压力较大
	case "woKe":
		score += 8 // 我克（财星），可得之象
	}

	switch branchRel {
	case "clash":
		score -= 30 // 六冲，冲散之义，影响最大
	case "harm":
		score -= 15 // 六害，暗害之义，影响次之
	case "punish":
		score -= 20 // 三刑，刑罚之义
	case "break":
		score -= 10 // 六破，破败之义
	case "combine":
		score += 8 // 六合，和合之情
	case "sanHe":
		score += 15 // 三合，气势专旺
	case "sanHui":
		score += 20 // 三会，方局之力更强
	}

	// 天干五合检测（经典：《三命通会》"天干五合论"）
	// 甲己合化土、乙庚合化金、丙辛合化水、丁壬合化木、戊癸合化火
	// 五合主气机交融、缘分契合，评分应有正面加成
	if userGan != "" && dayGan != "" && isGanHe(userGan, dayGan) {
		score += 12 // 五合加分，气机交融
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
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
	case "sanHe":
		return "三合"
	case "sanHui":
		return "三会"
	default:
		return "平和"
	}
}

func shengKeSummary(stemRel, branchRel string, score int) string {
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

	switch {
	case score >= 80:
		return base + "。运势大吉，诸事顺遂。"
	case score >= 60:
		return base + "。运势良好，宜积极进取。"
	case score >= 40:
		return base + "。运势平平，宜守不宜攻。"
	default:
		return base + "。运势欠佳，凡事小心为宜。"
	}
}

func periodSummary(avgScore int, period string) string {
	switch {
	case avgScore >= 70:
		return fmt.Sprintf("%s整体运势良好，适合积极行动，把握机会。", period)
	case avgScore >= 50:
		return fmt.Sprintf("%s整体运势平稳，按部就班即可，不必过于强求。", period)
	default:
		return fmt.Sprintf("%s整体运势偏低，建议多注意细节，避免冲动决策。", period)
	}
}

func pickYiJi(score int, stemRel string) (yi, ji []model.YiJiItem) {
	// 根据评分动态决定宜忌数量：分高多宜少忌，分低少宜多忌
	var yiCount, jiCount int
	switch {
	case score >= 80:
		yiCount, jiCount = 5, 2
	case score >= 60:
		yiCount, jiCount = 4, 3
	case score >= 40:
		yiCount, jiCount = 3, 4
	default:
		yiCount, jiCount = 2, 5
	}

	// 宜：优先插入十神维度推荐（最多2项），再用协纪辨方书列表补齐
	yi = make([]model.YiJiItem, 0, yiCount)
	added := make(map[string]bool)
	if godActs, ok := tenGodYiActs[stemRel]; ok {
		for i := 0; i < len(godActs) && len(yi) < yiCount && len(yi) < 2; i++ {
			act := godActs[i]
			yi = append(yi, model.YiJiItem{Activity: act, Reason: yiReason(act, score, stemRel)})
			added[act] = true
		}
	}
	for _, act := range allYiActs {
		if len(yi) >= yiCount {
			break
		}
		if !added[act] {
			yi = append(yi, model.YiJiItem{Activity: act, Reason: yiReason(act, score, stemRel)})
		}
	}

	// 忌：从协纪辨方书列表依次取
	ji = make([]model.YiJiItem, 0, jiCount)
	for i := 0; i < jiCount && i < len(allJiActs); i++ {
		ji = append(ji, model.YiJiItem{Activity: allJiActs[i], Reason: jiReason(allJiActs[i], score, stemRel)})
	}
	return
}

func yiReason(activity string, score int, stemRel string) string {
	reasons := map[string]string{
		// 协纪辨方书通用宜事
		"出行": "天时地利，出行顺利",
		"会友": "人际关系和谐，适合社交",
		"嫁娶": "阴阳调和，适合婚嫁",
		"祭祀": "吉星高照，适合祭祀祈福",
		"入学": "文昌显耀，学业有成",
		"开市": "财星显现，开市大吉",
		"求医": "贵人相助，求医有门",
		"纳采": "和合之气，纳采吉",
		"裁衣": "气运通达，裁衣顺遂",
		"竖柱": "根基稳固，竖柱大吉",
		"祈福": "神明庇佑，祈福灵验",
		"求嗣": "子星旺盛，求嗣有望",
		"牧养": "生机蓬勃，牧养得利",
		"宴饮": "和气生财，宴饮尽欢",
		"修造": "气运亨通，修造顺利",
		"移徙": "方位吉利，移徙安康",
		"安床": "阴阳调和，安床吉祥",
		"开光": "灵光显现，开光有应",
		"针灸": "气血调和，针灸见效",
		"解除": "煞气消散，解除有效",
		"栽种": "木气旺盛，栽种易活",
		"纳财": "财源广进，纳财有道",
		"交易": "诚信互利，交易顺遂",
		"创作": "灵感涌现，创作有成",
		// 十神维度
		"结盟": "比肩助力，结盟互利",
		"合作": "同心协力，合作有成",
	}
	r, ok := reasons[activity]
	if !ok {
		r = "今日宜" + activity
	}

	// 根据十神关系补充说明
	var godNote string
	switch stemRel {
	case "shengWo":
		godNote = "印星生扶"
	case "woSheng":
		godNote = "食伤泄秀"
	case "keWo":
		godNote = "官杀当值"
	case "woKe":
		godNote = "财星照临"
	case "same":
		godNote = "比和助力"
	}

	// 根据评分调整语气
	var mood string
	switch {
	case score >= 80:
		mood = "，大吉"
	case score >= 60:
		mood = "，吉利"
	case score >= 40:
		mood = "，尚可"
	default:
		mood = "，宜慎行"
	}

	if godNote != "" {
		return r + "（" + godNote + mood + "）"
	}
	return r + mood
}

func jiReason(activity string, score int, stemRel string) string {
	reasons := map[string]string{
		// 协纪辨方书通用忌事
		"动土": "冲煞较重，不宜动土",
		"安葬": "阴气较重，不宜安葬",
		"行丧": "凶星当值，不宜行丧",
		"开渠": "水气不利，不宜开渠",
		"伐木": "木气受损，不宜伐木",
		"破土": "地气不稳，不宜破土",
		"词讼": "口舌是非，不宜词讼",
		"远行": "路途多阻，不宜远行",
		"乘船": "水厄之象，不宜乘船",
		"渡水": "水势不顺，不宜渡水",
		"造屋": "根基不稳，不宜造屋",
		"上梁": "气运不济，不宜上梁",
		"掘井": "水脉受损，不宜掘井",
		"安门": "方位不利，不宜安门",
		"作灶": "火气不顺，不宜作灶",
	}
	r, ok := reasons[activity]
	if !ok {
		r = "今日忌" + activity
	}

	// 根据十神关系补充说明
	var godNote string
	switch stemRel {
	case "shengWo":
		godNote = "虽有印星生扶"
	case "woSheng":
		godNote = "食伤泄气"
	case "keWo":
		godNote = "官杀压身"
	case "woKe":
		godNote = "财星耗气"
	case "same":
		godNote = "比劫争财"
	}

	// 根据评分调整严重程度
	var severity string
	switch {
	case score >= 80:
		severity = "，小忌可解"
	case score >= 60:
		severity = "，宜慎行"
	case score >= 40:
		severity = "，切忌强为"
	default:
		severity = "，万不可为"
	}

	if godNote != "" {
		return r + "（" + godNote + severity + "）"
	}
	return r + severity
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
	return &DailyFortune{
		Date:          queryDate.Format("2006-01-02"),
		DayPillar:     model.Pillar{Gan: "?", Zhi: "?"},
		Score:         50,
		ElementImages: fixedElementImages(),
		Guide: &model.FortuneGuide{
			PrecisionLevel: "fallback",
			Confidence:     20,
			Analysis:       "未能取得有效日柱，开运指南暂按平稳守成为准。",
			Strategy:       "今日先保持作息和节奏稳定，重大事项等日课恢复后再定。",
		},
		ShengKe: ShengKeAnalysis{
			Summary: "无法计算日柱，使用默认运势。",
		},
		FlowImpact: "无法计算",
	}
}

// analyzeDayFlowImpact 分析流日对日主的流通影响
// 经典依据：滴天髓"地支藏干亦当论"——增加藏干分析
func analyzeDayFlowImpact(userChart *bazipkg.BaziResult, dayPillar model.Pillar) string {
	dayElem := data.GanElement[dayPillar.Gan]
	dayZhiElem := data.ZhiElement[dayPillar.Zhi]
	userDayElem := data.GanElement[userChart.DayPillar.Gan]

	var parts []string

	// 天干关系
	if dayElem == bazipkg.ShengWo(userDayElem) {
		parts = append(parts, fmt.Sprintf("天干%s（印星）生扶日主，得长辈或贵人助力", dayPillar.Gan))
	} else if dayElem == bazipkg.KeWo(userDayElem) {
		parts = append(parts, fmt.Sprintf("天干%s（官杀）克制日主，有压力或管制之象", dayPillar.Gan))
	} else if dayElem == bazipkg.WoKe(userDayElem) {
		parts = append(parts, fmt.Sprintf("天干%s（财星）为日主所克，有求财之机", dayPillar.Gan))
	} else if dayElem == bazipkg.WoSheng(userDayElem) {
		parts = append(parts, fmt.Sprintf("天干%s（食伤）泄日主之气，利创意表达但耗精力", dayPillar.Gan))
	} else if dayElem == userDayElem {
		parts = append(parts, fmt.Sprintf("天干%s（比劫）与日主同类，竞争或合作之象", dayPillar.Gan))
	}

	// 地支关系
	if dayZhiElem == bazipkg.ShengWo(userDayElem) {
		parts = append(parts, fmt.Sprintf("地支%s（印星）根基生扶，底层有贵人暗助", dayPillar.Zhi))
	} else if dayZhiElem == bazipkg.KeWo(userDayElem) {
		parts = append(parts, fmt.Sprintf("地支%s（官杀）暗中克身，需注意潜在压力", dayPillar.Zhi))
	} else if dayZhiElem == bazipkg.WoKe(userDayElem) {
		parts = append(parts, fmt.Sprintf("地支%s（财星）暗藏财机，有暗财或投资机会", dayPillar.Zhi))
	} else if dayZhiElem == bazipkg.WoSheng(userDayElem) {
		parts = append(parts, fmt.Sprintf("地支%s（食伤）才华暗泄，灵感丰富但精力分散", dayPillar.Zhi))
	} else if dayZhiElem == userDayElem {
		parts = append(parts, fmt.Sprintf("地支%s（比劫）同类帮身，根基稳固有依靠", dayPillar.Zhi))
	}

	// 藏干分析（经典：滴天髓"地支藏干亦当论"）
	zhiIdx := data.ZhiIndex(dayPillar.Zhi)
	if zhiIdx >= 0 {
		for _, s := range hiddenStemMap[zhiIdx] {
			if s == "" {
				continue
			}
			sElem := data.GanElement[s]
			if sElem == bazipkg.ShengWo(userDayElem) {
				parts = append(parts, fmt.Sprintf("藏干%s（印星）暗中助力，有隐性贵人", s))
			} else if sElem == bazipkg.KeWo(userDayElem) {
				parts = append(parts, fmt.Sprintf("藏干%s（官杀）暗中施压，需防暗中阻力", s))
			} else if sElem == bazipkg.WoKe(userDayElem) {
				parts = append(parts, fmt.Sprintf("藏干%s（财星）暗藏财机，有意外之财", s))
			} else if sElem == bazipkg.WoSheng(userDayElem) {
				parts = append(parts, fmt.Sprintf("藏干%s（食伤）暗中泄秀，潜藏才华可发掘", s))
			} else if sElem == userDayElem {
				parts = append(parts, fmt.Sprintf("藏干%s（比劫）暗中帮身，有隐形助力", s))
			}
		}
	}

	// 天干五合检测（经典：《三命通会》"天干五合论"）
	userDayGan := userChart.DayPillar.Gan
	if isGanHe(userDayGan, dayPillar.Gan) {
		huaElem := ganHeHuaElem(userDayGan, dayPillar.Gan)
		if huaElem != "" {
			parts = append(parts, fmt.Sprintf("日主%s与流日%s天干五合化%s，气机交融，运势有特殊变化", userDayGan, dayPillar.Gan, huaElem))
		}
	}

	if len(parts) == 0 {
		return "今日流通平稳，无明显增减，宜按部就班"
	}
	return strings.Join(parts, "；")
}

// 五行生克辅助函数已统一至 bazi/wuxing.go（ShengWo/WoSheng/WoKe/KeWo），
// 本文件通过 bazipkg.ShengWo 等导出函数调用，不再重复定义。

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
