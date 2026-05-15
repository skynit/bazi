package service

import (
	"fmt"
	"strings"
	"time"
)

// FortuneAnalysis is the detailed daily fortune reading.
type FortuneAnalysis struct {
	SolarDate   string          `json:"solar_date"`
	LunarDate   string          `json:"lunar_date"`
	UserBazi    string          `json:"user_bazi"`
	TodayGanZhi string          `json:"today_gan_zhi"`
	TodayElem   string          `json:"today_element"`
	Overall     OverallAnalysis `json:"overall"`
	Categories  []CategoryScore `json:"categories"`
	Hourly      []HourlyFortune `json:"hourly"`
	LuckyGuide  LuckyGuide      `json:"lucky_guide"`
}

type OverallAnalysis struct {
	Score   int    `json:"score"`
	Stars   string `json:"stars"`
	Summary string `json:"summary"`
	KeyTip  string `json:"key_tip"`
}

type CategoryScore struct {
	Name     string `json:"name"`
	Stars    string `json:"stars"`
	Analysis string `json:"analysis"`
	Advice   string `json:"advice"`
}

type HourlyFortune struct {
	Shichen    string `json:"shichen"`
	TimeRange  string `json:"time_range"`
	Mood       string `json:"mood"`
	Suggestion string `json:"suggestion"`
}

type LuckyGuide struct {
	Colors   string `json:"colors"`
	Numbers  string `json:"numbers"`
	Actions  string `json:"actions"`
	AvoidDir string `json:"avoid_dir"`
	FaceDir  string `json:"face_dir"`
	Outfit   string `json:"outfit"`
}

// elementInfo maps gan to element name
var ganElement = map[string]string{
	"甲": "木", "乙": "木", "丙": "火", "丁": "火", "戊": "土",
	"己": "土", "庚": "金", "辛": "金", "壬": "水", "癸": "水",
}

var zhiElement = map[string]string{
	"寅": "木", "卯": "木", "巳": "火", "午": "火",
	"辰": "土", "戌": "土", "丑": "土", "未": "土",
	"申": "金", "酉": "金", "亥": "水", "子": "水",
}

// generating and overcoming cycles
var generates = map[string]string{"木": "火", "火": "土", "土": "金", "金": "水", "水": "木"}
var overcomes = map[string]string{"木": "土", "土": "水", "水": "火", "火": "金", "金": "木"}
var generatedBy = map[string]string{"木": "水", "火": "木", "土": "火", "金": "土", "水": "金"}

// tenGods maps relationship to ten-god name
var tenGodNames = map[string]string{
	"same": "比肩", "diff_yang": "劫财",
	"gen_me": "食神", "i_gen": "伤官",
	"i_overcome": "偏财", "overcome_me": "正财",
	"overcome_me2": "七杀", "overcome_by": "正官",
	"gen_by": "偏印", "gen_by2": "正印",
}

// shichen info
var shichenInfo = []struct {
	name, timeRange, desc string
}{
	{"子时", "23-01点", "水旺，宜静思"},
	{"丑时", "01-03点", "土旺，宜安睡"},
	{"寅时", "03-05点", "木旺，宜早起"},
	{"卯时", "05-07点", "木旺，宜晨练"},
	{"辰时", "07-09点", "土旺，宜早餐"},
	{"巳时", "09-11点", "火旺，宜工作"},
	{"午时", "11-13点", "火旺极，宜午休"},
	{"未时", "13-15点", "土旺，宜学习"},
	{"申时", "15-17点", "金旺，宜决策"},
	{"酉时", "17-19点", "金旺，宜社交"},
	{"戌时", "19-21点", "土旺，宜放松"},
	{"亥时", "21-23点", "水旺，宜静养"},
}

// AnalyzeDailyFortune generates a detailed fortune reading
func AnalyzeDailyFortune(userBazi *BaziResult, todayDayGan, todayDayZhi string) *FortuneAnalysis {
	userDayGan := userBazi.DayPillar.Gan
	userElem := ganElement[userDayGan]
	todayElem := ganElement[todayDayGan]

	// Determine if today's stem is favorable (喜用神) or unfavorable (忌神)
	isFavorable := isFavorableElement(userElem, todayElem)
	isUnfavorable := isOvercomingElement(userElem, todayElem)

	// Score based on favorability
	baseScore := 50
	if isFavorable {
		baseScore += 25
	} else if isUnfavorable {
		baseScore -= 20
	}
	// Adjust based on branch harmony
	harmonyScore := checkBranchHarmony(userBazi, todayDayZhi)
	baseScore += harmonyScore
	if baseScore > 100 { baseScore = 100 }
	if baseScore < 10 { baseScore = 10 }

	stars := scoreToStars(baseScore)

	// Generate the user bazi string
	userBaziStr := fmt.Sprintf("%s%s %s%s %s%s %s%s",
		userBazi.YearPillar.Gan, userBazi.YearPillar.Zhi,
		userBazi.MonthPillar.Gan, userBazi.MonthPillar.Zhi,
		userBazi.DayPillar.Gan, userBazi.DayPillar.Zhi,
		userBazi.HourPillar.Gan, userBazi.HourPillar.Zhi,
	)

	todayGZ := todayDayGan + todayDayZhi
	todayElemStr := ganElement[todayDayGan] + zhiElement[todayDayZhi]

	// Generate overall summary
	var summary string
	var keyTip string
	favStr := "喜用神"
	unfavStr := "忌神"
	if isFavorable {
		summary = fmt.Sprintf("今日天干%s（%s），与您的日主%s（%s）相生，为%s。",
			todayDayGan, todayElem, userDayGan, userElem, favStr)
		summary += fmt.Sprintf("地支%s为%s。整体运势较佳，宜主动把握机会。", todayDayZhi, zhiElement[todayDayZhi])
		keyTip = "今日运势上扬，宜积极行动，把握贵人运。"
	} else if isUnfavorable {
		summary = fmt.Sprintf("今日天干%s（%s），与您的日主%s（%s）相克，为%s。",
			todayDayGan, todayElem, userDayGan, userElem, unfavStr)
		summary += fmt.Sprintf("但地支%s为%s，可缓解天干之克。上午宜谨慎，下午好转。", todayDayZhi, zhiElement[todayDayZhi])
		keyTip = "上午耐心应对，下午运势好转，宜主动出击。"
	} else {
		summary = fmt.Sprintf("今日天干%s（%s），与您的日主%s（%s）同五行，运势平稳。",
			todayDayGan, todayElem, userDayGan, userElem)
		keyTip = "今日运势平稳，适合处理日常事务，不宜做重大决定。"
	}

	// Category scores
	categories := []CategoryScore{
		makeCategory("事业运", baseScore+5, todayDayGan, todayElem, userElem, isFavorable),
		makeCategory("财运", baseScore+10, todayDayGan, todayElem, userElem, isFavorable),
		makeCategory("感情运", baseScore-5, todayDayGan, todayElem, userElem, isFavorable),
		makeCategory("健康运", baseScore, todayDayGan, todayElem, userElem, isFavorable),
		makeCategory("贵人运", baseScore+15, todayDayGan, todayElem, userElem, isFavorable),
	}

	return &FortuneAnalysis{
		SolarDate:   time.Now().Format("2006-01-02"),
		LunarDate:   "农历（通过tyme4go获取）",
		UserBazi:    userBaziStr,
		TodayGanZhi: todayGZ,
		TodayElem:   todayElemStr,
		Overall:     OverallAnalysis{baseScore, stars, summary, keyTip},
		Categories:  categories,
		Hourly:      makeHourly(todayDayGan, todayDayZhi, isFavorable),
		LuckyGuide:  makeLuckyGuide(todayDayGan, todayDayZhi, userElem),
	}
}

func isFavorableElement(userElem, todayElem string) bool {
	return generates[todayElem] == userElem || generatedBy[userElem] == todayElem
}

func isOvercomingElement(userElem, todayElem string) bool {
	return overcomes[todayElem] == userElem
}

func checkBranchHarmony(bazi *BaziResult, todayZhi string) int {
	branches := []string{bazi.YearPillar.Zhi, bazi.MonthPillar.Zhi, bazi.DayPillar.Zhi, bazi.HourPillar.Zhi}
	score := 0
	for _, b := range branches {
		if b == todayZhi { score += 5 }              // same branch
		if isLiuHe(b, todayZhi) { score += 15 }       // harmony
		if isLiuChong(b, todayZhi) { score -= 15 }    // clash
	}
	return score
}

func isLiuHe(a, b string) bool {
	he := map[string]string{"子":"丑","丑":"子","寅":"亥","亥":"寅","卯":"戌","戌":"卯","辰":"酉","酉":"辰","巳":"申","申":"巳","午":"未","未":"午"}
	return he[a] == b
}

func isLiuChong(a, b string) bool {
	chong := map[string]string{"子":"午","午":"子","丑":"未","未":"丑","寅":"申","申":"寅","卯":"酉","酉":"卯","辰":"戌","戌":"辰","巳":"亥","亥":"巳"}
	return chong[a] == b
}

func scoreToStars(score int) string {
	switch {
	case score >= 85: return "★★★★★"
	case score >= 70: return "★★★★☆"
	case score >= 55: return "★★★☆☆"
	case score >= 40: return "★★☆☆☆"
	default: return "★☆☆☆☆"
	}
}

func makeCategory(name string, baseScore int, gan, ganElem, userElem string, favorable bool) CategoryScore {
	adj := 0
	if name == "财运" && favorable { adj = 5 }
	if name == "贵人运" { adj = 10 }
	if name == "感情运" && !favorable { adj = -5 }
	score := baseScore + adj
	if score > 100 { score = 100 }
	if score < 10 { score = 10 }

	stars := scoreToStars(score)
	var analysis, advice string
	elemName := ganElement[gan]

	switch name {
	case "事业运":
		if favorable {
			analysis = fmt.Sprintf("今日%s气旺，与日主相生，有利于工作进展。适合处理重要项目，容易得到上级认可。", elemName)
			advice = "宜主动承担任务，展现能力。下午是高效时段。"
		} else {
			analysis = fmt.Sprintf("今日%s气较重，与日主相克，工作中可能遇到阻力。宜低调行事，避免正面冲突。", elemName)
			advice = "上午做机械性工作，重要决策留到下午。"
		}
	case "财运":
		if favorable {
			analysis = fmt.Sprintf("今日财星得力，%s元素旺相。可能有额外收入或理财收益。适合谈合作、签约。", elemName)
			advice = "下午5-7点检查账户，理性消费。"
		} else {
			analysis = fmt.Sprintf("今日财星受制，不宜大额消费或投资。%s气过重，注意避免冲动购物。", elemName)
			advice = "保守理财，避免高风险投资。"
		}
	case "感情运":
		if favorable {
			analysis = fmt.Sprintf("今日%s气柔，人际关系和谐。单身者有机会认识新朋友。有伴侣者适合一起放松。", elemName)
			advice = "傍晚适合社交或约会。"
		} else {
			analysis = fmt.Sprintf("今日%s气燥，容易因小事产生摩擦。注意控制情绪，多包容对方。", elemName)
			advice = "避免在上午讨论敏感话题。"
		}
	case "健康运":
		analysis = fmt.Sprintf("今日需注意%s对应的身体部位。保持作息规律，适当运动。", elemName)
		advice = "多喝水，饮食清淡。"
	case "贵人运":
		if favorable {
			analysis = "今日贵人运较旺，容易得到他人帮助。可以主动联系旧友或前辈。"
			advice = "贵人方位在西北方，下午主动联系。"
		} else {
			analysis = "今日贵人运平平，遇事多靠自己。不必强求他人帮助。"
			advice = "独立解决问题，积累经验。"
		}
	}

	return CategoryScore{name, stars, analysis, advice}
}

func makeHourly(dayGan, dayZhi string, favorable bool) []HourlyFortune {
	result := make([]HourlyFortune, 12)
	for i, si := range shichenInfo {
		mood := si.desc
		sug := "宜正常作息"
		// Adjust based on time of day
		if i >= 3 && i <= 5 && favorable { sug = "宜处理重要事务" }
		if i >= 6 && i <= 8 && favorable { sug = "宜做决策" }
		if i == 4 && favorable { sug = "午休充电" }
		result[i] = HourlyFortune{si.name, si.timeRange, mood, sug}
	}
	return result
}

func makeLuckyGuide(dayGan, dayZhi, userElem string) LuckyGuide {
	elem := ganElement[dayGan]
	colors := map[string]string{"木":"绿色、青色", "火":"红色、紫色", "土":"黄色、棕色", "金":"白色、金色", "水":"黑色、蓝色"}
	numbers := map[string]string{"木":"3、8", "火":"2、7", "土":"5、0", "金":"4、9", "水":"1、6"}
	avoidDirs := map[string]string{"木":"西方", "火":"北方", "土":"东方", "金":"南方", "水":"中央"}
	faceDirs := map[string]string{"木":"东方", "火":"南方", "土":"中央", "金":"西方", "水":"北方"}

	// Favorable colors are the ones that generate the user element
	genElem := generatedBy[userElem]

	return LuckyGuide{
		Colors:   fmt.Sprintf("%s（%s）、%s（%s）", colors[genElem], genElem, colors[userElem], userElem),
		Numbers:  fmt.Sprintf("%s；%s", numbers[genElem], numbers[userElem]),
		Actions:  fmt.Sprintf("随身携带%s属性物品；面向%s方工作", genElem, faceDirs[genElem]),
		AvoidDir: avoidDirs[elem],
		FaceDir:  faceDirs[genElem],
		Outfit:   fmt.Sprintf("%s色上衣+%s色裤子", colors[genElem], colors[userElem]),
	}
}

// FormatAnalysis returns the analysis as a formatted string (for testing/display)
func (fa *FortuneAnalysis) FormatAnalysis() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s 运势详解\n", fa.SolarDate))
	sb.WriteString(fmt.Sprintf("您的八字：%s\n", fa.UserBazi))
	sb.WriteString(fmt.Sprintf("今日干支：%s（%s）\n\n", fa.TodayGanZhi, fa.TodayElem))
	sb.WriteString(fmt.Sprintf("一、整体运势：%s\n%s\n核心提示：%s\n\n", fa.Overall.Stars, fa.Overall.Summary, fa.Overall.KeyTip))
	sb.WriteString("二、分维度详解\n")
	for _, c := range fa.Categories {
		sb.WriteString(fmt.Sprintf("%s %s  %s\n  建议：%s\n", c.Name, c.Stars, c.Analysis, c.Advice))
	}
	sb.WriteString("\n三、今日开运指南\n")
	sb.WriteString(fmt.Sprintf("幸运色：%s\n", fa.LuckyGuide.Colors))
	sb.WriteString(fmt.Sprintf("幸运数字：%s\n", fa.LuckyGuide.Numbers))
	sb.WriteString(fmt.Sprintf("开运动作：%s\n", fa.LuckyGuide.Actions))
	sb.WriteString(fmt.Sprintf("避开方位：%s\n", fa.LuckyGuide.AvoidDir))
	sb.WriteString(fmt.Sprintf("面向方位：%s\n", fa.LuckyGuide.FaceDir))
	sb.WriteString(fmt.Sprintf("穿搭建议：%s\n", fa.LuckyGuide.Outfit))
	return sb.String()
}
