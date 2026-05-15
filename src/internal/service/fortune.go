package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/6tail/tyme4go/tyme"

	"bazi/internal/model"
)

// FortuneEngine computes daily, weekly, and monthly fortunes by comparing
// a user's BaZi chart pillars against the query date's pillars.
type FortuneEngine struct {
	bazi       *BaziService
	auspicious *AuspiciousData
}

// NewFortuneEngine creates a ready-to-use FortuneEngine.
func NewFortuneEngine() *FortuneEngine {
	return &FortuneEngine{
		bazi:       &BaziService{},
		auspicious: NewAuspiciousData(),
	}
}

// DailyFortune is a single-day fortune result.
type DailyFortune struct {
	Date            string               `json:"date"`
	DayPillar       model.Pillar         `json:"day_pillar"`
	Score           int                  `json:"score"`
	LuckyColor      string               `json:"lucky_color"`
	LuckyNumbers    []int                `json:"lucky_numbers"`
	WealthDir       string               `json:"wealth_dir"`
	ClashZodiac     string               `json:"clash_zodiac"`
	AuspiciousHours []string             `json:"auspicious_hours"`
	Yi              []model.YiJiItem     `json:"yi"`
	Ji              []model.YiJiItem     `json:"ji"`
	ShengKe         ShengKeAnalysis      `json:"sheng_ke"`
	ElementImages   []model.ElementImage `json:"element_images"`
}

// WeeklyFortune aggregates seven daily fortunes.
type WeeklyFortune struct {
	WeekStart      string              `json:"week_start"`
	DailyFortunes  []DailyFortune      `json:"daily_fortunes"`
	OverallSummary string              `json:"overall_summary"`
	WeeklyScore    int                 `json:"weekly_score"`
	ElementTrend   []ElementTrendPoint `json:"element_trend"`
}

// MonthlyFortune aggregates a full calendar month of daily fortunes.
type MonthlyFortune struct {
	Year           int                 `json:"year"`
	Month          int                 `json:"month"`
	DailyFortunes  []DailyFortune      `json:"daily_fortunes"`
	OverallSummary string              `json:"overall_summary"`
	MonthlyScore   int                 `json:"monthly_score"`
	ElementTrend   []ElementTrendPoint `json:"element_trend"`
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

// ShengKeAnalysis describes the generating/overcoming (生克) relationship
// between the user's day pillar and the query date's day pillar.
type ShengKeAnalysis struct {
	DayStemRelation   string `json:"day_stem_relation"`
	DayBranchRelation string `json:"day_branch_relation"`
	Summary           string `json:"summary"`
}

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

// CalculateDaily computes the fortune for a single date.
func (e *FortuneEngine) CalculateDaily(userChart *BaziResult, queryDate time.Time) *DailyFortune {
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

	score := calcScore(stemRel, branchRel)
	luckyColor := e.auspicious.GetLuckyColor(dayPillar.Gan)
	luckyNumbers := e.auspicious.GetLuckyNumbers(dayPillar.Gan)
	wealthDir := e.auspicious.GetWealthDirection(dayPillar.Gan)
	clashZodiac := e.auspicious.GetClashZodiac(dayPillar.Zhi)
	auspHours := e.auspicious.GetAuspiciousHours(dayPillar.Zhi)

	yi, ji := pickYiJi(score, stemRel)

	shengKe := ShengKeAnalysis{
		DayStemRelation:   stemRelLabel(stemRel, userDayStem, dayPillar.Gan),
		DayBranchRelation: branchRelLabel(branchRel),
		Summary:           shengKeSummary(stemRel, branchRel, score),
	}

	return &DailyFortune{
		Date:            queryDate.Format("2006-01-02"),
		DayPillar:       dayPillar,
		Score:           score,
		LuckyColor:      luckyColor,
		LuckyNumbers:    luckyNumbers,
		WealthDir:       wealthDir,
		ClashZodiac:     clashZodiac,
		AuspiciousHours: auspHours,
		Yi:              yi,
		Ji:              ji,
		ShengKe:         shengKe,
		ElementImages:   fixedElementImages(),
	}
}

// CalculateWeekly computes fortunes for 7 consecutive days starting from weekStart.
func (e *FortuneEngine) CalculateWeekly(userChart *BaziResult, weekStart time.Time) *WeeklyFortune {
	weekStart = toDateStart(weekStart)
	fortunes := make([]DailyFortune, 7)
	trends := make([]ElementTrendPoint, 7)
	totalScore := 0

	for i := 0; i < 7; i++ {
		day := weekStart.AddDate(0, 0, i)
		df := e.CalculateDaily(userChart, day)
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
	}
}

// CalculateMonthly computes fortunes for every day in the given year/month.
func (e *FortuneEngine) CalculateMonthly(userChart *BaziResult, year, month int) *MonthlyFortune {
	days := daysInMonth(year, month)
	fortunes := make([]DailyFortune, 0, days)
	trends := make([]ElementTrendPoint, 0, days)
	totalScore := 0

	for d := 1; d <= days; d++ {
		date := time.Date(year, time.Month(month), d, 12, 0, 0, 0, time.UTC)
		df := e.CalculateDaily(userChart, date)
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

func stemRelation(userStem, queryStem string) string {
	ue := stemToElement[userStem]
	qe := stemToElement[queryStem]
	if ue == "" || qe == "" {
		return "unknown"
	}
	if ue == qe {
		return "same"
	}
	if elementGenerates[ue] == qe || elementGenerates[qe] == ue {
		return "gen"
	}
	if elementOvercomes[ue] == qe || elementOvercomes[qe] == ue {
		return "overcome"
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
	return "neutral"
}

func calcScore(stemRel, branchRel string) int {
	score := 50

	switch stemRel {
	case "same":
		score += 10
	case "gen":
		score += 20
	case "overcome":
		score -= 20
	}

	switch branchRel {
	case "clash", "harm":
		score -= 30
	case "combine":
		score += 5
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
	ue := stemToElement[userStem]
	qe := stemToElement[queryStem]
	switch rel {
	case "same":
		return "比和"
	case "gen":
		if elementGenerates[ue] == qe {
			return "我生"
		}
		return "生我"
	case "overcome":
		if elementOvercomes[ue] == qe {
			return "我克"
		}
		return "克我"
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
	default:
		return "平和"
	}
}

func shengKeSummary(stemRel, branchRel string, score int) string {
	var parts []string
	if stemRel != "unknown" {
		parts = append(parts, fmt.Sprintf("日干关系: %s", stemRel))
	}
	if branchRel != "neutral" && branchRel != "unknown" {
		parts = append(parts, fmt.Sprintf("日支关系: %s", branchRel))
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
	yiActs := []string{"出行", "会友", "嫁娶", "祭祀", "入学", "开市"}
	jiActs := []string{"动土", "安葬", "行丧", "开渠", "伐木"}

	yi = make([]model.YiJiItem, 0, 3)
	for _, act := range yiActs[:min(3, len(yiActs))] {
		yi = append(yi, model.YiJiItem{
			Activity: act,
			Reason:   yiReason(act, score, stemRel),
		})
	}

	ji = make([]model.YiJiItem, 0, 2)
	for _, act := range jiActs[:min(2, len(jiActs))] {
		ji = append(ji, model.YiJiItem{
			Activity: act,
			Reason:   jiReason(act, score, stemRel),
		})
	}
	return
}

func yiReason(activity string, _ int, _ string) string {
	reasons := map[string]string{
		"出行": "天时地利，出行顺利",
		"会友": "人际关系和谐，适合社交",
		"嫁娶": "阴阳调和，适合婚嫁",
		"祭祀": "吉星高照，适合祭祀祈福",
		"入学": "文昌显耀，学业有成",
		"开市": "财星显现，开市大吉",
	}
	if r, ok := reasons[activity]; ok {
		return r
	}
	return "今日宜此"
}

func jiReason(activity string, _ int, _ string) string {
	reasons := map[string]string{
		"动土": "冲煞较重，不宜动土",
		"安葬": "阴气较重，不宜安葬",
		"行丧": "凶星当值，不宜行丧",
		"开渠": "水气不利，不宜开渠",
		"伐木": "木气受损，不宜伐木",
	}
	if r, ok := reasons[activity]; ok {
		return r
	}
	return "今日忌此"
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
		{Element: "金", ImageURL: "/images/elements/metal.svg", Description: "金"},
		{Element: "木", ImageURL: "/images/elements/wood.svg", Description: "木"},
		{Element: "水", ImageURL: "/images/elements/water.svg", Description: "水"},
		{Element: "火", ImageURL: "/images/elements/fire.svg", Description: "火"},
		{Element: "土", ImageURL: "/images/elements/earth.svg", Description: "土"},
	}
}

func (e *FortuneEngine) fallbackDaily(queryDate time.Time) *DailyFortune {
	return &DailyFortune{
		Date:          queryDate.Format("2006-01-02"),
		DayPillar:     model.Pillar{Gan: "?", Zhi: "?"},
		Score:         50,
		ElementImages: fixedElementImages(),
		ShengKe: ShengKeAnalysis{
			Summary: "无法计算日柱，使用默认运势。",
		},
	}
}

func toDateStart(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func daysInMonth(year, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}
