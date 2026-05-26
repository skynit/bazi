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
	Colors           string   `json:"colors"`
	Numbers          string   `json:"numbers"`
	Actions          string   `json:"actions"`
	AvoidDir         string   `json:"avoid_dir"`
	FaceDir          string   `json:"face_dir"`
	Outfit           string   `json:"outfit"`
	FavorableElems   []string `json:"favorable_elems"`   // 喜用五行
	UnfavorableElems []string `json:"unfavorable_elems"` // 忌五行
}

// GanElements and ZhiElements are defined in data_gans.go.
// Use GanElements[GanIndex(gan)] for stem elements and ZhiElements[ZhiIndex(zhi)] for branch elements.

// generates/overcomes relationships for wuxing
var generates = map[string]string{"木": "火", "火": "土", "土": "金", "金": "水", "水": "木"}
var overcomes = map[string]string{"木": "土", "土": "水", "水": "火", "火": "金", "金": "木"}

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

type catResult struct {
	score int
	cs    CategoryScore
}

// ── 主分析入口 ──────────────────────────────────────────────

func AnalyzeDailyFortune(userBazi *BaziResult, todayDayGan, todayDayZhi string) *FortuneAnalysis {
	userDayGan := userBazi.DayPillar.Gan
	userElem := GanElement[userDayGan]
	todayElem := GanElement[todayDayGan]
	todayZhiElem := ZhiElement[todayDayZhi]

	// 判断身旺程度
	isStrong := userBazi.BodyStrength.Verdict == "身旺"

	// 十神关系
	rel := tenGodRelation(userDayGan, todayDayGan, isStrong)

	// 各维度独立评分（起分60），保留原始分数用于加权平均
	results := []catResult{
		scoreCareer(rel, todayDayZhi, isStrong),
		scoreWealth(rel, todayDayGan, todayZhiElem),
		scoreLove(todayDayZhi, userBazi),
		scoreHealth(todayElem, userElem, isStrong),
		scoreNoble(todayDayZhi, userBazi),
		scoreStudy(rel, todayDayGan),
		scoreInvest(rel, todayDayGan),
		scoreTravel(todayDayZhi, userBazi),
		scoreLawsuit(rel, todayDayZhi, userBazi),
	}

	// 综合评分 = 九维度加权平均（权重：事业15 财12 感情10 健康8 贵人10 学业8 投资8 出行5 是非5）
	weights := []int{15, 12, 10, 8, 10, 8, 8, 5, 5}
	total, weightTotal := 0, 0
	categories := make([]CategoryScore, len(results))
	for i, r := range results {
		total += r.score * weights[i]
		weightTotal += weights[i]
		categories[i] = r.cs
	}
	baseScore := total / weightTotal

	stars := scoreToStars(baseScore)
	userBaziStr := fmt.Sprintf("%s%s %s%s %s%s %s%s",
		userBazi.YearPillar.Gan, userBazi.YearPillar.Zhi,
		userBazi.MonthPillar.Gan, userBazi.MonthPillar.Zhi,
		userBazi.DayPillar.Gan, userBazi.DayPillar.Zhi,
		userBazi.HourPillar.Gan, userBazi.HourPillar.Zhi,
	)

	summary, keyTip := makeSummary(userDayGan, userElem, todayDayGan, todayElem, todayDayZhi, rel, userBazi)

	return &FortuneAnalysis{
		SolarDate:   time.Now().Format("2006-01-02"),
		LunarDate:   "农历（通过tyme4go获取）",
		UserBazi:    userBaziStr,
		TodayGanZhi: todayDayGan + todayDayZhi,
		TodayElem:   todayElem + todayZhiElem,
		Overall:     OverallAnalysis{baseScore, stars, summary, keyTip},
		Categories:  categories,
		Hourly:      makeHourly(todayDayGan, todayDayZhi, isFavorableRel(rel)),
		LuckyGuide:  makeLuckyGuide(todayDayGan, todayDayZhi, userElem, userBazi.BodyStrength.Like, userBazi.BodyStrength.Dislike),
	}
}

// ── 十神关系 ─────────────────────────────────────────────────

func tenGodRelation(userGan, dayGan string, isStrong bool) string {
	ue := GanElement[userGan]
	de := GanElement[dayGan]
	if ue == "" || de == "" {
		return "中性"
	}

	userGanIdx := ganIndex(userGan)
	dayGanIdx := ganIndex(dayGan)
	sameYinYang := (userGanIdx%2 == dayGanIdx%2) // 同阴阳

	// 同五行 → 比劫
	if ue == de {
		if sameYinYang {
			return "比肩"
		}
		return "劫财"
	}

	// 我生 → 食伤
	if generates[ue] == de {
		if sameYinYang {
			return "食神"
		}
		return "伤官"
	}

	// 生我 → 印星
	if generates[de] == ue {
		if sameYinYang {
			return "偏印"
		}
		return "正印"
	}

	// 我克 → 财星
	if overcomes[ue] == de {
		if sameYinYang {
			return "偏财"
		}
		return "正财"
	}

	// 克我 → 官杀
	if overcomes[de] == ue {
		if sameYinYang {
			return "七杀"
		}
		return "正官"
	}

	return "中性"
}

func ganIndex(gan string) int {
	gans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	for i, g := range gans {
		if g == gan {
			return i
		}
	}
	return 0
}

// ── 地支关系计算 ─────────────────────────────────────────────

func checkBranchHarmony(bazi *BaziResult, todayZhi string) int {
	branches := []string{bazi.YearPillar.Zhi, bazi.MonthPillar.Zhi, bazi.DayPillar.Zhi, bazi.HourPillar.Zhi}
	score := 0
	for _, b := range branches {
		if b == todayZhi {
			score += 5
		}
		if isLiuHe(b, todayZhi) {
			score += 15
		}
		if isLiuChong(b, todayZhi) {
			score -= 15
		}
		if isLiuHai(b, todayZhi) {
			score -= 10
		}
		if isSanHe(b, todayZhi) {
			score += 10
		}
	}
	// 自刑
	if isSelfPunish(todayZhi) && hasSelfPunish(bazi, todayZhi) {
		score -= 10
	}
	return score
}

func isLiuHe(a, b string) bool {
	he := map[string]string{"子": "丑", "丑": "子", "寅": "亥", "亥": "寅", "卯": "戌", "戌": "卯", "辰": "酉", "酉": "辰", "巳": "申", "申": "巳", "午": "未", "未": "午"}
	return he[a] == b
}

func isLiuChong(a, b string) bool {
	chong := map[string]string{"子": "午", "午": "子", "丑": "未", "未": "丑", "寅": "申", "申": "寅", "卯": "酉", "酉": "卯", "辰": "戌", "戌": "辰", "巳": "亥", "亥": "巳"}
	return chong[a] == b
}

func isLiuHai(a, b string) bool {
	hai := map[string]string{"子": "未", "未": "子", "丑": "午", "午": "丑", "寅": "巳", "巳": "寅", "卯": "辰", "辰": "卯", "申": "亥", "亥": "申", "酉": "戌", "戌": "酉"}
	return hai[a] == b
}

func isSanHe(a, b string) bool {
	sets := [][]string{{"申", "子", "辰"}, {"寅", "午", "戌"}, {"亥", "卯", "未"}, {"巳", "酉", "丑"}}
	for _, set := range sets {
		for _, x := range set {
			if x == b {
				for _, y := range set {
					if y == a && y != b {
						return true
					}
				}
			}
		}
	}
	return false
}

func isSelfPunish(zhi string) bool {
	return zhi == "午" || zhi == "酉" || zhi == "辰" || zhi == "亥"
}

func hasSelfPunish(bazi *BaziResult, zhi string) bool {
	count := 0
	for _, b := range []string{bazi.YearPillar.Zhi, bazi.MonthPillar.Zhi, bazi.DayPillar.Zhi, bazi.HourPillar.Zhi} {
		if b == zhi {
			count++
		}
	}
	return count >= 2
}

func buildBranchDesc(bazi *BaziResult, todayZhi string) string {
	branches := []string{bazi.YearPillar.Zhi, bazi.MonthPillar.Zhi, bazi.DayPillar.Zhi, bazi.HourPillar.Zhi}
	var parts []string
	for _, b := range branches {
		if isLiuHe(b, todayZhi) {
			parts = append(parts, fmt.Sprintf("与%s六合", b))
		}
		if isLiuChong(b, todayZhi) {
			parts = append(parts, fmt.Sprintf("与%s六冲", b))
		}
	}
	if hasSelfPunish(bazi, todayZhi) && isSelfPunish(todayZhi) {
		parts = append(parts, "自刑")
	}
	if len(parts) == 0 {
		return fmt.Sprintf("地支%s平和", todayZhi)
	}
	return fmt.Sprintf("地支%s（%s）", todayZhi, strings.Join(parts, "、"))
}

// ── 综合概述 ─────────────────────────────────────────────────

func makeSummary(userGan, userElem, todayGan, todayElem, todayZhi, rel string, bazi *BaziResult) (string, string) {
	branchDesc := buildBranchDesc(bazi, todayZhi)

	switch rel {
	case "喜用":
		s := fmt.Sprintf("今日天干%s（%s），与日主%s（%s）相生，为喜用神。%s。整体运势较佳。",
			todayGan, todayElem, userGan, userElem, branchDesc)
		return s, "今日运势上扬，宜积极行动，把握贵人运。"

	case "忌神":
		s := fmt.Sprintf("今日天干%s（%s），与日主%s（%s）相克，为忌神。%s。上午宜谨慎，下午好转。",
			todayGan, todayElem, userGan, userElem, branchDesc)
		return s, "上午耐心应对，下午运势好转，宜主动出击。"

	case "比肩", "劫财":
		s := fmt.Sprintf("今日%s（%s）为比劫，竞争增强。%s。",
			todayGan, todayElem, branchDesc)
		return s, "竞争激烈，注意合作而非对抗。"

	case "食神", "伤官":
		s := fmt.Sprintf("今日食伤（%s）透出，利于创意表达。%s。",
			todayGan, branchDesc)
		return s, "创意迸发，适合写作、设计、技术攻关。"

	case "正财", "偏财":
		s := fmt.Sprintf("今日财星（%s）透出，利于求财。%s。",
			todayGan, branchDesc)
		return s, "适合理财规划，把握商机。"

	case "正官":
		s := fmt.Sprintf("今日正官（%s）透出，事业机会增多。%s。",
			todayGan, branchDesc)
		return s, "适合展示专业能力，抓住晋升机会。"

	case "七杀":
		s := fmt.Sprintf("今日七杀（%s）透出，有挑战也有机遇。%s。有压力亦有破局之机。",
			todayGan, branchDesc)
		return s, "有挑战亦有破局之机。下午申时后运势转好。"

	case "正印", "偏印":
		s := fmt.Sprintf("今日印星（%s）透出，利于学习思考。%s。",
			todayGan, branchDesc)
		return s, "适合学习充电，安静思考。"

	default:
		s := fmt.Sprintf("今日%s（%s），与日主%s（%s）平和。%s。",
			todayGan, todayElem, userGan, userElem, branchDesc)
		return s, "保持平常心，顺势而为。"
	}
}

// ── 九维度评分 ───────────────────────────────────────────────

func scoreCareer(rel string, todayZhi string, isStrong bool) catResult {
	score := 60
	var analysis, advice string

	switch rel {
	case "七杀":
		score += 8
		analysis = "七杀透出，利于攻坚克难。适合处理竞争性事务，容易得到上级认可。"
		advice = "宜主动承担挑战性任务，下午申时后效率最高。"
	case "正官":
		score += 10
		analysis = "正官显耀，事业机会增多。适合展示专业能力，争取晋升。"
		advice = "主动汇报工作成果，把握贵人提携机会。"
	case "食神", "伤官":
		score += 6
		analysis = "食伤泄秀，利于创意和技术类工作。思绪活跃，适合攻坚。"
		advice = "专注技术难题，下午3-5点为高效时段。"
	case "正财", "偏财":
		score += 4
		analysis = "财星透出，工作中可能有额外收益机会。适合商务洽谈。"
		advice = "把握商机，但人际关系上不宜强求。"
	case "比肩", "劫财":
		if isStrong {
			score -= 5
			analysis = "比劫旺，竞争激烈。容易与同事产生摩擦，宜低调行事。"
			advice = "注意合作而非对抗，避免正面冲突。"
		} else {
			score += 5
			analysis = "比劫帮扶，工作中容易得到同事支持。适合团队协作。"
			advice = "依靠团队力量，分担工作压力。"
		}
	case "正印", "偏印":
		analysis = "印星透出，利于文书、学习类工作。适合处理细致事务。"
		advice = "上午处理文案工作，效率最高。"
	default:
		analysis = "今日事业运平稳，按部就班即可。"
		advice = "处理日常工作，不宜做重大决策。"
	}

	return catResult{score, finalize(score, "事业运", analysis, advice)}
}

func scoreWealth(rel string, todayGan, todayZhiElem string) catResult {
	score := 60
	var analysis, advice string

	switch rel {
	case "正财", "偏财":
		score += 12
		analysis = fmt.Sprintf("财星（%s）透出，利于求财合作。可能有额外收入或理财收益。", GanElement[todayGan])
		advice = "适合理财规划，下午5-7点检查账户。"
	case "食神", "伤官":
		score += 6
		analysis = "食伤生财，创意可转化为收入。适合商务洽谈和签约。"
		advice = "发挥专业优势，将想法变现。"
	case "七杀", "正官":
		score += 2
		analysis = "官杀透出，财运平平。宜保守理财，避免风险投资。"
		advice = "守住钱袋子，大额支出需三思。"
	case "正印", "偏印":
		score -= 3
		analysis = "印星耗财，不宜大额消费。注意避免冲动购物。"
		advice = "保守理财，适合存款和稳健投资。"
	default:
		analysis = "今日财运平稳，无大起大落。"
		advice = "按计划消费，不必过虑。"
	}

	return catResult{score, finalize(score, "财运", analysis, advice)}
}

func scoreLove(todayZhi string, bazi *BaziResult) catResult {
	score := 60
	var analysis, advice string

	dayBranch := bazi.DayPillar.Zhi

	if isLiuHe(dayBranch, todayZhi) {
		score += 10
		analysis = "今日地支六合，人际关系和谐。单身者有机会认识新朋友。"
		advice = "傍晚适合社交或约会，把握缘分。"
	} else if isLiuChong(dayBranch, todayZhi) {
		score -= 10
		analysis = "今日地支六冲，容易因小事产生摩擦。注意控制情绪。"
		advice = "避免在上午讨论敏感话题，多包容对方。"
	} else if hasSelfPunish(bazi, todayZhi) && isSelfPunish(todayZhi) {
		score -= 8
		analysis = "今日自刑，情绪易烦躁。单身者不宜表白，有伴者多忍让。"
		advice = "务必午休，避免午时讨论敏感话题。多喝水降火。"
	} else if isLiuHe(bazi.YearPillar.Zhi, todayZhi) || isLiuHe(bazi.HourPillar.Zhi, todayZhi) {
		score += 5
		analysis = "今日与人缘宫位相合，人际关系顺畅。适合社交放松。"
		advice = "傍晚可约朋友小聚，放松心情。"
	} else {
		analysis = "今日感情运平淡，顺其自然即可。"
		advice = "用心经营，细水长流。"
	}

	return catResult{score, finalize(score, "感情运", analysis, advice)}
}

func scoreHealth(todayElem, userElem string, isStrong bool) catResult {
	score := 60
	var analysis, advice string

	if todayElem == userElem {
		score += 5
		analysis = "今日元素与日主相同，身体状态良好，精力充沛。"
		advice = "适合运动锻炼，保持作息规律。"
	} else if generates[todayElem] == userElem {
		score += 3
		analysis = "今日元素生身，精力补充。适合调养身体。"
		advice = "多喝水，饮食清淡，适度运动。"
	} else if overcomes[todayElem] == userElem {
		score -= 5
		if isStrong {
			analysis = fmt.Sprintf("今日%s克土，虽利于制衡旺土，但需注意消化系统。", todayElem)
			advice = "饮食规律，少吃辛辣油腻。"
		} else {
			analysis = fmt.Sprintf("今日%s克身，需注意%s对应的身体部位。容易疲惫。", todayElem, todayElem)
			advice = "注意休息，多补充水分，避免劳累。"
		}
	} else {
		analysis = "今日身体状态一般，保持作息规律即可。"
		advice = "多喝水，饮食清淡，适当运动。"
	}

	return catResult{score, finalize(score, "健康运", analysis, advice)}
}

func scoreNoble(todayZhi string, bazi *BaziResult) catResult {
	score := 60
	var analysis, advice string

	// 贵人运看地支六合
	hasHe := false
	for _, b := range []string{bazi.YearPillar.Zhi, bazi.MonthPillar.Zhi, bazi.DayPillar.Zhi, bazi.HourPillar.Zhi} {
		if isLiuHe(b, todayZhi) {
			hasHe = true
			break
		}
	}

	if hasHe {
		score += 10
		analysis = "今日贵人运较旺，容易得到他人帮助。可以主动联系旧友或前辈。"
		advice = "贵人方位在西北方向，下午主动联系。"
	} else if isLiuChong(bazi.DayPillar.Zhi, todayZhi) {
		score -= 5
		analysis = "今日冲日柱，贵人运平平。遇事多靠自己。"
		advice = "独立解决问题，积累经验。"
	} else {
		analysis = "今日贵人运平稳，不强求他人帮助。"
		advice = "靠西北坐下，身边放一杯水。"
	}

	return catResult{score, finalize(score, "贵人运", analysis, advice)}
}

func scoreStudy(rel string, todayGan string) catResult {
	score := 60
	var analysis, advice string

	switch rel {
	case "正印", "偏印":
		score += 8
		analysis = "印星透出，利于学习思考、备考复习。文昌显耀。"
		advice = "上午记忆力最佳，适合攻克理论难题。"
	case "食神", "伤官":
		score += 6
		analysis = "食伤泄秀，思绪活跃。适合学习新技术、做创意训练。"
		advice = "下午申时金水旺，利于突破难题。"
	case "七杀":
		score += 4
		analysis = "七杀透出，有压力才有动力。适合复习备考、攻克难关。"
		advice = "上午宜静心学习，将压力转化为效率。"
	default:
		analysis = fmt.Sprintf("今日%s气平稳，适合常规学习。按计划推进即可。", GanElement[todayGan])
		advice = "保持学习节奏，厚积薄发。"
	}

	return catResult{score, finalize(score, "学业运", analysis, advice)}
}

func scoreInvest(rel string, todayGan string) catResult {
	score := 60
	var analysis, advice string

	switch rel {
	case "正财", "偏财":
		score += 10
		analysis = "财星透出，利于投资理财。可适当关注市场机会。"
		advice = "理性分析，适度参与，见好就收。"
	case "七杀", "正官":
		score -= 5
		analysis = "官杀透出，投资风险增大。不宜冲动决策。"
		advice = "保守观望，避免高风险投资。"
	default:
		score -= 3
		analysis = "今日偏财星无气，不宜激进投资。适合存款和稳健理财。"
		advice = "守住钱袋子，不宜盲目跟风。"
	}

	return catResult{score, finalize(score, "投资运", analysis, advice)}
}

func scoreTravel(todayZhi string, bazi *BaziResult) catResult {
	score := 60
	var analysis, advice string

	hasChong := false
	for _, b := range []string{bazi.YearPillar.Zhi, bazi.MonthPillar.Zhi, bazi.DayPillar.Zhi, bazi.HourPillar.Zhi} {
		if isLiuChong(b, todayZhi) {
			hasChong = true
			break
		}
	}

	if hasChong {
		score -= 10
		analysis = "今日地支六冲，出行易有阻滞。不宜长途旅行或冒险行动。"
		advice = "如需出行，务必注意安全，避开午时(11-13点)。"
	} else if isLiuHe(bazi.DayPillar.Zhi, todayZhi) {
		score += 5
		analysis = "今日六合，出行顺利。适合短途旅行或户外活动。"
		advice = "午后出行最佳，注意防晒补水。"
	} else {
		analysis = "今日出行运平稳，常规出行无碍。"
		advice = "如需出行，避开午时(11-13点)，可选申时后。"
	}

	return catResult{score, finalize(score, "出行运", analysis, advice)}
}

func scoreLawsuit(rel string, todayZhi string, bazi *BaziResult) catResult {
	score := 60
	var analysis, advice string

	if rel == "七杀" || rel == "正官" {
		score -= 5
		analysis = "官杀透出，易有口舌是非。注意谨言慎行。"
		advice = "低调行事，遇争执先停顿3秒再回应。"
	} else if hasSelfPunish(bazi, todayZhi) && isSelfPunish(todayZhi) {
		score -= 3
		analysis = "今日自刑，内心易纠结。注意避免与人争执。"
		advice = "多喝水降心火，保持冷静。"
	} else {
		analysis = "今日是非运平稳，无特别需要注意的事项。"
		advice = "谨言慎行，与人为善。"
	}

	return catResult{score, finalize(score, "是非运", analysis, advice)}
}

// ── 评分工具 ─────────────────────────────────────────────────

func finalize(score int, name, analysis, advice string) CategoryScore {
	if score > 100 {
		score = 100
	}
	if score < 20 {
		score = 20
	}
	return CategoryScore{
		Name:     name,
		Stars:    scoreToStars(score),
		Analysis: analysis,
		Advice:   advice,
	}
}

func scoreToStars(score int) string {
	switch {
	case score >= 85:
		return "★★★★★"
	case score >= 70:
		return "★★★★☆"
	case score >= 55:
		return "★★★☆☆"
	case score >= 40:
		return "★★☆☆☆"
	default:
		return "★☆☆☆☆"
	}
}

// ── 时辰运势 ─────────────────────────────────────────────────

func isFavorableRel(rel string) bool {
	switch rel {
	case "食神", "正财", "正官", "正印":
		return true
	default:
		return false
	}
}

func makeHourly(dayGan, dayZhi string, favorable bool) []HourlyFortune {
	result := make([]HourlyFortune, 12)
	for i, si := range shichenInfo {
		mood := si.desc
		sug := "宜正常作息"
		if i >= 3 && i <= 5 && favorable {
			sug = "宜处理重要事务"
		}
		if i >= 6 && i <= 8 {
			sug = "宜做决策或学习"
		}
		if i == 4 {
			sug = "午休充电，避免争论"
		}
		result[i] = HourlyFortune{si.name, si.timeRange, mood, sug}
	}
	return result
}

// ── 开运指南 ─────────────────────────────────────────────────

func makeLuckyGuide(dayGan, dayZhi, userElem string, like, dislike []string) LuckyGuide {
	elem := GanElement[dayGan]
	colors := map[string]string{"木": "绿色、青色", "火": "红色、紫色", "土": "黄色、棕色", "金": "白色、金色", "水": "黑色、蓝色"}
	numbers := map[string]string{"木": "3、8", "火": "2、7", "土": "5、0", "金": "4、9", "水": "1、6"}
	avoidDirs := map[string]string{"木": "西方", "火": "北方", "土": "东方", "金": "南方", "水": "中央"}
	faceDirs := map[string]string{"木": "东方", "火": "南方", "土": "中央", "金": "西方", "水": "北方"}

	// 以首选用神(like[0])作为当日开运元素
	primaryElem := userElem
	if len(like) > 0 {
		primaryElem = like[0]
	}

	return LuckyGuide{
		Colors:           fmt.Sprintf("%s（%s）、%s（%s）", colors[primaryElem], primaryElem, colors[userElem], userElem),
		Numbers:          fmt.Sprintf("%s；%s", numbers[primaryElem], numbers[userElem]),
		Actions:          fmt.Sprintf("随身携带%s属性物品；面向%s方工作", primaryElem, faceDirs[primaryElem]),
		AvoidDir:         avoidDirs[elem],
		FaceDir:          faceDirs[primaryElem],
		Outfit:           fmt.Sprintf("%s色上衣+%s色裤子", colors[primaryElem], colors[userElem]),
		FavorableElems:   like,
		UnfavorableElems: dislike,
	}
}

// ── 格式化输出 ───────────────────────────────────────────────

func (fa *FortuneAnalysis) FormatAnalysis() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s 运势详解\n", fa.SolarDate))
	sb.WriteString(fmt.Sprintf("您的八字：%s\n", fa.UserBazi))
	sb.WriteString(fmt.Sprintf("今日干支：%s（%s）\n\n", fa.TodayGanZhi, fa.TodayElem))
	sb.WriteString(fmt.Sprintf("一、整体运势：%s %d分\n%s\n核心提示：%s\n\n", fa.Overall.Stars, fa.Overall.Score, fa.Overall.Summary, fa.Overall.KeyTip))
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
