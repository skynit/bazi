package service

import (
	"fmt"
	"strings"

	"github.com/kaecer68/ziwei-zenith/pkg/basis"
)

// ──────────────────── Period Interpretation Service ────────────────────

// PeriodInterpreter analyzes 流年/流月/流日 data and produces
// human-readable interpretations with ShiShen, relations, scores, and advice.
type PeriodInterpreter struct {
	birthInfo basis.BirthInfo
}

// NewPeriodInterpreter creates a PeriodInterpreter with the given birth info.
func NewPeriodInterpreter(birthInfo basis.BirthInfo) *PeriodInterpreter {
	return &PeriodInterpreter{birthInfo: birthInfo}
}

// ──────────────────── Result types ────────────────────

type LiunianResult struct {
	Year           int    `json:"year"`
	GanZhi         string `json:"gan_zhi"`
	GanZhiDesc     string `json:"gan_zhi_desc"`
	ShiShen        string `json:"shi_shen"`
	RelationToMing string `json:"relation_to_ming"`
	OverallTone    string `json:"overall_tone"`
	KeyTips        string `json:"key_tips"`
	Score          int    `json:"score"`
}

type LiuyueResult struct {
	Year           int    `json:"year"`
	Month          int    `json:"month"`
	GanZhi         string `json:"gan_zhi"`
	GanZhiDesc     string `json:"gan_zhi_desc"`
	ShiShen        string `json:"shi_shen"`
	RelationToMing string `json:"relation_to_ming"`
	Effect         string `json:"effect"`
	Health         string `json:"health"`
	Score          int    `json:"score"`
}

type LiuriResult struct {
	Year           int         `json:"year"`
	Month          int         `json:"month"`
	Day            int         `json:"day"`
	GanZhi         string      `json:"gan_zhi"`
	GanZhiDesc     string      `json:"gan_zhi_desc"`
	ShiShen        string      `json:"shi_shen"`
	RelationToMing string      `json:"relation_to_ming"`
	QiZiEffect     string      `json:"qi_zi_effect"`
	EmotionalState string      `json:"emotional_state"`
	Health         string      `json:"health"`
	Score          int         `json:"score"`
	HourlyAnalysis []HourBlock `json:"hourly_analysis"`
	Summary        string      `json:"summary"`
}

type HourBlock struct {
	Hour       int    `json:"hour"`
	StemBranch string  `json:"stem_branch"`
	Effect     string `json:"effect"`
	Score      int    `json:"score"`
}

// PeriodSummary holds the summary of all three layers.
type PeriodSummary struct {
	Liunian LiunianSummaryItem `json:"liunian"`
	Liuyue  LiuyueSummaryItem  `json:"liuyue"`
	Liuri  LiuriSummaryItem    `json:"liuri"`
	Advice PeriodAdvice        `json:"advice"`
}

type LiunianSummaryItem struct {
	GanZhi      string `json:"gan_zhi"`
	ShiShen     string `json:"shi_shen"`
	Relation    string `json:"relation"`
	Score       int    `json:"score"`
	Description string `json:"description"`
}

type LiuyueSummaryItem struct {
	GanZhi      string `json:"gan_zhi"`
	ShiShen     string `json:"shi_shen"`
	Relation    string `json:"relation"`
	Score       int    `json:"score"`
	Description string `json:"description"`
}

type LiuriSummaryItem struct {
	GanZhi      string `json:"gan_zhi"`
	ShiShen     string `json:"shi_shen"`
	Relation    string `json:"relation"`
	Score       int    `json:"score"`
	Description string `json:"description"`
}

type PeriodAdvice struct {
	Liunian []string `json:"liunian"`
	Liuyue  []string `json:"liuyue"`
	Liuri   []string `json:"liuri"`
}


// stemName returns the Chinese name of a Stem.
func stemName(s basis.Stem) string {
	if s >= 0 && int(s) < len(stemNames) {
		return stemNames[s]
	}
	return ""
}

// branchName returns the Chinese name of a Branch.
func branchName(b basis.Branch) string {
	if b >= 0 && int(b) < len(branchNames) {
		return branchNames[b]
	}
	return ""
}

// wuXingStem returns the five-element name for a stem.
func wuXingStem(s basis.Stem) string {
	switch s {
	case 0, 1:
		return "木"
	case 2, 3:
		return "火"
	case 4, 5:
		return "土"
	case 6, 7:
		return "金"
	case 8, 9:
		return "水"
	}
	return ""
}

// wuXingBranch returns the five-element name for a branch.
func wuXingBranch(b basis.Branch) string {
	switch b {
	case 0, 1:
		return "水"
	case 2, 3:
		return "木"
	case 4, 5:
		return "土"
	case 6, 7:
		return "火"
	case 8, 9:
		return "金"
	case 10, 11:
		return "土"
	}
	return ""
}

// isYangStem returns true if the stem is yang (甲丙戊庚壬).
func isYangStem(s basis.Stem) bool {
	return s == 0 || s == 2 || s == 4 || s == 6 || s == 8
}

// stemWuXingIdx returns the 5-element index for stem (0=木, 1=火, 2=土, 3=金, 4=水).
func stemWuXingIdx(s basis.Stem) int {
	switch s {
	case 0, 1:
		return 0 // 木
	case 2, 3:
		return 1 // 火
	case 4, 5:
		return 2 // 土
	case 6, 7:
		return 3 // 金
	case 8, 9:
		return 4 // 水
	}
	return 0
}

// getShiShen returns the ShiShen (十神) name for a given stem vs the day stem.
func getShiShen(stem, dayStem basis.Stem) string {
	stemIdx := int(stem)
	dayIdx := int(dayStem)

	if stemIdx == dayIdx {
		if isYangStem(dayStem) {
			return "比肩"
		}
		return "比劫"
	}

	stemEle := stemWuXingIdx(stem)
	dayEle := stemWuXingIdx(dayStem)

	// (dayEle+1)%5 == stemEle → stem 生 day → 印星
	// (dayEle+3)%5 == stemEle → stem 克 day → 官星
	// (stemEle+1)%5 == dayEle → day 生 stem → 食神/伤官
	// (stemEle+3)%5 == dayEle → day 克 stem → 财星
	isBornBy := (stemEle == (dayEle+1)%5)
	isOvercome := (stemEle == (dayEle+3)%5)
	isBirth := (dayEle == (stemEle+1)%5)
	isConquer := (dayEle == (stemEle+3)%5)

	switch {
	case isBornBy:
		if isYangStem(stem) {
			return "正印"
		}
		return "偏印"
	case isBirth:
		if isYangStem(dayStem) {
			return "食神"
		}
		return "伤官"
	case isOvercome:
		if isYangStem(stem) {
			return "正官"
		}
		return "七杀"
	case isConquer:
		if isYangStem(stem) {
			return "正财"
		}
		return "偏财"
	}
	return "无关系"
}

// relationPair returns the relationship description between two branches.
func relationPair(b1, b2 basis.Branch) string {
	if b1 == b2 {
		switch b1 {
		case 4, 7, 9, 11:
			return "自刑"
		}
		return "伏吟"
	}

	pair := int(b1)*100 + int(b2)
	switch {
	// 六冲: 子-午, 丑-未, 寅-申, 卯-酉, 辰-戌, 巳-亥
	case pair == 0*100+6 || pair == 6*100+0 ||
		pair == 1*100+7 || pair == 7*100+1 ||
		pair == 2*100+8 || pair == 8*100+2 ||
		pair == 3*100+9 || pair == 9*100+3 ||
		pair == 4*100+10 || pair == 10*100+4 ||
		pair == 5*100+11 || pair == 11*100+5:
		return "六冲"
	// 六合: 子-丑, 寅-亥, 卯-戌, 辰-酉, 巳-申, 午-未
	case pair == 0*100+1 || pair == 1*100+0 ||
		pair == 2*100+11 || pair == 11*100+2 ||
		pair == 3*100+10 || pair == 10*100+3 ||
		pair == 4*100+9 || pair == 9*100+4 ||
		pair == 5*100+8 || pair == 8*100+5 ||
		pair == 6*100+7 || pair == 7*100+6:
		return "六合"
	// 三刑: 寅巳申三刑, 丑戌未三刑, 子卯刑
	case pair == 2*100+5 || pair == 5*100+2 || // 寅巳
		pair == 5*100+8 || pair == 8*100+5 || // 巳申
		pair == 2*100+8 || pair == 8*100+2 || // 寅申
		pair == 1*100+10 || pair == 10*100+1 || // 丑戌
		pair == 10*100+7 || pair == 7*100+10 || // 戌未
		pair == 1*100+7 || pair == 7*100+1 || // 丑未
		pair == 0*100+3 || pair == 3*100+0: // 子卯
		return "三刑"
	}
	return ""
}

// describeRelation creates a Chinese description of how period branch relates to birth branches.
func (s *PeriodInterpreter) describeRelation(periodBranch basis.Branch) string {
	var parts []string

	keyBranches := []basis.Branch{
		s.birthInfo.YearPillar.Branch,
		s.birthInfo.MonthPillar.Branch,
		s.birthInfo.DayPillar.Branch,
		s.birthInfo.HourPillar.Branch,
	}

	for _, bc := range keyBranches {
		rel := relationPair(periodBranch, bc)
		if rel != "" {
			parts = append(parts, rel)
		}
	}

	if len(parts) == 0 {
		return "与命局无特殊刑冲合关系"
	}

	seen := map[string]bool{}
	unique := []string{}
	for _, p := range parts {
		if !seen[p] {
			seen[p] = true
			unique = append(unique, p)
		}
	}
	return strings.Join(unique, "、")
}

// evaluateScore returns a 0-100 score based on the stem/branch and birth info.
func (s *PeriodInterpreter) evaluateScore(stem basis.Stem, periodBranch basis.Branch) int {
	score := 60

	dayStem := s.birthInfo.DayPillar.Stem
	shiShen := getShiShen(stem, dayStem)

	switch shiShen {
	case "正印", "偏印", "比肩", "比劫", "食神", "伤官":
		score += 15
	case "七杀":
		score -= 15
	case "偏财":
		score -= 5
	}

	rel := relationPair(periodBranch, s.birthInfo.DayPillar.Branch)
	if rel == "六冲" || rel == "自刑" || rel == "伏吟" || rel == "三刑" {
		score -= 10
	} else if rel == "六合" {
		score += 10
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	return score
}

// healthAdvice returns health advice based on branch wuxing.
func healthAdvice(branch basis.Branch) string {
	switch branch {
	case 2, 3: // 寅卯 木
		return "注意肝胆、筋骨、手足"
	case 4, 5: // 辰巳 土
		return "注意脾胃、皮肤、代谢"
	case 6, 7: // 午未 火
		return "注意心脑血管、眼睛、咽喉"
	case 8, 9: // 申酉 金
		return "注意肺呼吸道、肠道、筋骨"
	case 0, 1, 10, 11: // 亥子水 + 戌丑
		return "注意肾膀胱、泌尿系统、水液代谢"
	}
	return "注意整体健康"
}

// ──────────────────── Analysis methods ────────────────────

// AnalyzeLiunian produces a full interpretation for the given 流年.
func (s *PeriodInterpreter) AnalyzeLiunian(chart *ZiWeiChart, year int) *LiunianResult {
	if chart == nil || chart.engineChart == nil {
		return nil
	}
	ec := chart.engineChart
	stem := ec.LiuNian.Stem
	branch := ec.LiuNian.Branch

	ganZhi := stemName(stem) + branchName(branch)
	shiShen := getShiShen(stem, s.birthInfo.DayPillar.Stem)
	rel := s.describeRelation(branch)
	score := s.evaluateScore(stem, branch)

	var tone string
	if score >= 70 {
		tone = fmt.Sprintf("流年%s对命局有情，运势较佳", rel)
	} else if score >= 45 {
		tone = fmt.Sprintf("流年%s对命局影响中性，运势平稳", rel)
	} else {
		tone = fmt.Sprintf("流年%s克命局，运势偏弱", rel)
	}

	var tips string
	if strings.Contains(rel, "伏吟") || strings.Contains(rel, "自刑") || strings.Contains(rel, "六冲") {
		tips = fmt.Sprintf("今年%s，运势反复，注意健康、情绪、财务三大方面。", rel)
	} else if score >= 70 {
		tips = "今年运势较佳，宜把握机会，推进事业学业。"
	} else if score >= 45 {
		tips = "今年运势平稳，稳扎稳打，不宜冒进。"
	} else {
		tips = "今年运势偏弱，宜守不宜攻，注重健康管理。"
	}

	return &LiunianResult{
		Year:           year,
		GanZhi:         ganZhi,
		GanZhiDesc:     fmt.Sprintf("%s（%s）+ %s（%s）", stemName(stem), shiShen, branchName(branch), wuXingBranch(branch)),
		ShiShen:        shiShen,
		RelationToMing: rel,
		OverallTone:    tone,
		KeyTips:        tips,
		Score:          score,
	}
}

// AnalyzeLiuyue produces a full interpretation for the given 流月.
func (s *PeriodInterpreter) AnalyzeLiuyue(chart *ZiWeiChart, year, month int) *LiuyueResult {
	if chart == nil || chart.engineChart == nil {
		return nil
	}
	ec := chart.engineChart
	branch := ec.LiuYue
	// LiuYue stem = (dayStem*2 + month - 1) % 10
	stem := basis.Stem((int(s.birthInfo.DayPillar.Stem)*2 + (month - 1)) % 10)

	ganZhi := stemName(stem) + branchName(branch)
	shiShen := getShiShen(stem, s.birthInfo.DayPillar.Stem)
	rel := s.describeRelation(branch)
	score := s.evaluateScore(stem, branch)

	var effect string
	if score >= 65 {
		effect = fmt.Sprintf("本月运势较佳，%s透干，有进展", shiShen)
	} else if score >= 40 {
		effect = fmt.Sprintf("本月运势平稳，%s透干，需注意调节", shiShen)
	} else {
		effect = fmt.Sprintf("本月忌神发力，%s透干，防破财争执", shiShen)
	}

	return &LiuyueResult{
		Year:           year,
		Month:          month,
		GanZhi:         ganZhi,
		GanZhiDesc:     fmt.Sprintf("%s（%s）+ %s（%s）", stemName(stem), shiShen, branchName(branch), wuXingBranch(branch)),
		ShiShen:        shiShen,
		RelationToMing: rel,
		Effect:         effect,
		Health:         healthAdvice(branch),
		Score:          score,
	}
}

// AnalyzeLiuri produces a full interpretation for the given 流日.
func (s *PeriodInterpreter) AnalyzeLiuri(chart *ZiWeiChart, year, month, day int) *LiuriResult {
	if chart == nil || chart.engineChart == nil {
		return nil
	}
	ec := chart.engineChart
	branch := ec.LiuRi
	// LiuRi stem = (dayStem * 2 + dayOfMonth - 1) % 10
	stem := basis.Stem((int(s.birthInfo.DayPillar.Stem)*2 + (day - 1)) % 10)

	ganZhi := stemName(stem) + branchName(branch)
	shiShen := getShiShen(stem, s.birthInfo.DayPillar.Stem)
	rel := s.describeRelation(branch)
	score := s.evaluateScore(stem, branch)

	// 12 时辰 analysis: 0=子时(23-01), 1=丑时(01-03), ..., 11=亥时(21-23)
	hourly := make([]HourBlock, 12)

	for i := 0; i < 12; i++ {
		hourStemIdx := (int(stem)*2 + i) % 10
		hourStem := basis.Stem(hourStemIdx)
		hourBranch := basis.Branch(i)
		hourGanZhi := stemName(hourStem) + branchName(hourBranch)
		hourScore := s.evaluateScore(hourStem, hourBranch)
		hourShiShen := getShiShen(hourStem, s.birthInfo.DayPillar.Stem)

		var effect string
		if hourScore >= 65 {
			effect = fmt.Sprintf("%s时，%s透出%s，吉利", branchName(hourBranch), shiShen, hourShiShen)
		} else if hourScore >= 45 {
			effect = fmt.Sprintf("%s时，%s透出%s，平稳", branchName(hourBranch), shiShen, hourShiShen)
		} else {
			effect = fmt.Sprintf("%s时，%s透出%s，需谨慎", branchName(hourBranch), shiShen, hourShiShen)
		}

		hourly[i] = HourBlock{
			Hour:       i*2 + 1,
			StemBranch: hourGanZhi,
			Effect:     effect,
			Score:      hourScore,
		}
	}

	var emotion string
	if strings.Contains(rel, "自刑") || strings.Contains(rel, "伏吟") {
		emotion = "内心烦躁、易怒、钻牛角尖，注意情绪管理"
	} else if score >= 65 {
		emotion = "心情舒畅，思路清晰，适合处理重要事务"
	} else if score >= 45 {
		emotion = "情绪平稳，但小有波动，注意调节"
	} else {
		emotion = "情绪低落或焦躁，宜静心，勿做重大决策"
	}

	summary := fmt.Sprintf("%s日评分%d分，%s透干。%s", ganZhi, score, shiShen, emotion)

	return &LiuriResult{
		Year:           year,
		Month:          month,
		Day:            day,
		GanZhi:         ganZhi,
		GanZhiDesc:     fmt.Sprintf("%s（%s）+ %s（%s）", stemName(stem), shiShen, branchName(branch), wuXingBranch(branch)),
		ShiShen:        shiShen,
		RelationToMing: rel,
		QiZiEffect:     fmt.Sprintf("%s透出%s，对事业财运有影响", shiShen, ganZhi),
		EmotionalState: emotion,
		Health:         healthAdvice(branch),
		Score:          score,
		HourlyAnalysis: hourly,
		Summary:        summary,
	}
}

// SummarizeAll produces a summary combining all three period layers.
func (s *PeriodInterpreter) SummarizeAll(liunian, liuyue, liuri *ZiWeiChart, year, month, day int) *PeriodSummary {
	ln := s.AnalyzeLiunian(liunian, year)
	ly := s.AnalyzeLiuyue(liuyue, year, month)
	lr := s.AnalyzeLiuri(liuri, year, month, day)

	var adv PeriodAdvice
	if ln != nil {
		if ln.Score < 50 {
			adv.Liunian = []string{"禁止投资借贷创业", "夏季主动体检", "技术深耕但不跳槽"}
		} else if ln.Score >= 70 {
			adv.Liunian = []string{"把握机会拓展事业", "财运较佳可适当投资"}
		} else {
			adv.Liunian = []string{"稳扎稳打，专注主业", "注意健康管理"}
		}
	}
	if ly != nil {
		if ly.Score < 50 {
			adv.Liuyue = []string{"守住钱包，避免大额消费", "每天2升温水", "减少无效社交"}
		} else {
			adv.Liuyue = []string{"财运平稳，注意把握机会", "注意脾胃健康"}
		}
	}
	if lr != nil {
		if lr.Score < 50 {
			adv.Liuri = []string{"午时闭目养神，不处理重要事务", "下午申时把握机会"}
		} else {
			adv.Liuri = []string{"思路清晰，适合处理技术难点"}
		}
	}

	return &PeriodSummary{
		Liunian: LiunianSummaryItem{
			GanZhi:      ln.GanZhi,
			ShiShen:     ln.ShiShen,
			Relation:    ln.RelationToMing,
			Score:       ln.Score,
			Description: ln.OverallTone,
		},
		Liuyue: LiuyueSummaryItem{
			GanZhi:      ly.GanZhi,
			ShiShen:     ly.ShiShen,
			Relation:    ly.RelationToMing,
			Score:       ly.Score,
			Description: ly.Effect,
		},
		Liuri: LiuriSummaryItem{
			GanZhi:      lr.GanZhi,
			ShiShen:     lr.ShiShen,
			Relation:    lr.RelationToMing,
			Score:       lr.Score,
			Description: lr.Summary,
		},
		Advice: adv,
	}
}