package service

import (
	"fmt"
	"math"
	"strings"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

// TenGodRatio holds one ten-god's count and percentage in a birth chart.
type TenGodRatio struct {
	Name    string  `json:"name"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

// tenGodNames is the fixed ordering for proportion output.
var tenGodNames = [10]string{
	"比肩", "劫财", "食神", "伤官",
	"正财", "偏财", "正官", "七杀",
	"正印", "偏印",
}

// BaziService calculates BaZi (八字) birth charts using tyme4go.
type BaziService struct{}

// BaziResult holds the complete BaZi calculation output.
type BaziResult struct {
	YearPillar       model.Pillar         `json:"year_pillar"`
	MonthPillar      model.Pillar         `json:"month_pillar"`
	DayPillar        model.Pillar         `json:"day_pillar"`
	HourPillar       model.Pillar         `json:"hour_pillar"`
	FiveElements     map[string]int       `json:"five_elements"`
	ElementDetail    []ElementStrength    `json:"element_detail"`
	BodyStrength     BodyStrengthResult   `json:"body_strength"`
	TenGods          map[string]string    `json:"ten_gods"`
	NaYin            map[string]NaYinInfo `json:"na_yin"`
	HiddenStems      map[string][]string  `json:"hidden_stems"`
	DaYunInfo        DaYunInfo            `json:"da_yun_info"`
	ClashHarmony     []ClashRelation      `json:"clash_harmony"`
	GanZhiAnalysis   GanZhiAnalysis       `json:"gan_zhi_analysis"`
	PatternAnalysis PatternAnalysis      `json:"pattern_analysis"`
	MingGong         MingGongDetail      `json:"ming_gong"`
	RiZhuDesc        string               `json:"ri_zhu_desc"`
	PillarDetails    []PillarDetail       `json:"pillar_details"`
	DayStemTiaoHou   string               `json:"tiao_hou"`
	DayStemJinBuHuan string               `json:"jin_bu_huan"`
	DayShenSha       []string             `json:"day_shen_sha"`
	ShenShaByPillar  []PillarShenSha      `json:"shen_sha_by_pillar"`
	GlobalShenSha    []string             `json:"global_shen_sha"`
	ShenShaSummary   *ShenShaSummary      `json:"shen_sha_summary"`
	SeasonText       string               `json:"season_text"`
	SeasonTextMonth  string               `json:"season_text_month"` // month-specific YueTexts when available
	TenGodProportion []TenGodRatio        `json:"ten_god_proportion"`
	TenGodAnalysis   *TenGodAnalysis      `json:"ten_god_analysis"`
	RiZhuPoem        string               `json:"ri_zhu_poem"`
	RiZhuSource      string               `json:"ri_zhu_source"`
	RiZhuComment     string               `json:"ri_zhu_comment"`
	RiZhuHourDetail  string               `json:"ri_zhu_hour_detail"`
}

// ElementStrength holds the strength breakdown for one element.
type ElementStrength struct {
	Element     string   `json:"element"`
	TianGan     int      `json:"tian_gan"`
	ZhiCangGan  int      `json:"zhi_cang_gan"`
	Total       int      `json:"total"`
	CangGanList []string `json:"cang_gan_list"` // e.g. ["未中乙", "辰中乙"]
}

// BodyStrengthResult holds the body strength conclusion.
type BodyStrengthResult struct {
	Verdict    string   `json:"verdict"`
	Like       []string `json:"like"`
	Dislike    []string `json:"dislike"`
	TotalScore float64  `json:"total_score"`
	LingScore  float64  `json:"ling_score"`
	DiScore    float64  `json:"di_score"`
	ShiScore   float64  `json:"shi_score"`
	ShengScore float64  `json:"sheng_score"`
}

// DaYunInfo describes the major fortune cycle (大运).
type DaYunInfo struct {
	StartAge  int            `json:"start_age"`
	Direction string         `json:"direction"`
	Pillars   []model.Pillar `json:"pillars"`
}

// ClashRelation describes a clash/harmony relation between two pillars.
type ClashRelation struct {
	Pillar1 string `json:"pillar1"`
	Pillar2 string `json:"pillar2"`
	Type    string `json:"type"`
}

// PillarShenSha groups shen-sha items for a single pillar with metadata.
type PillarShenSha struct {
	Pillar   string   `json:"pillar"`
	Label    string   `json:"label"`
	Gan      string   `json:"gan"`
	Zhi      string   `json:"zhi"`
	Priority int      `json:"priority"`
	Role     string   `json:"role"`
	Items    []string `json:"items"`
}

// ShenShaSummary provides a high-level explanation of the shen-sha ordering.
type ShenShaSummary struct {
	Title       string   `json:"title"`
	Description []string `json:"description"`
}

// NaYinInfo is the JSON-serializable na-yin detail for API responses.
// StemBranches is omitted to keep response compact; it is available in the knowledge base.
type NaYinInfo struct {
	Name        string   `json:"name"`
	Element     string   `json:"element"`
	ImageDesc   string   `json:"image_desc"`
	Personality string   `json:"personality"`
	EnergyStage string   `json:"energy_stage"`
	ModernExt   string   `json:"modern_ext"`
	Judgments   []string `json:"judgments"`
}

// PillarDetail holds enriched per-pillar data.
type PillarDetail struct {
	Stem      string    `json:"stem"`
	Branch    string    `json:"branch"`
	ShengXiao string    `json:"sheng_xiao"`
	Empties   [2]string `json:"empties"`
	Nayin     NaYinInfo `json:"nayin"`
}

// Calculate computes a full BaZi chart.
func (s *BaziService) Calculate(year, month, day, hour, minute int, gender string) (*BaziResult, error) {
	tymeGender, err := toTymeGender(gender)
	if err != nil {
		return nil, err
	}

	st, err := tyme.SolarTime{}.FromYmdHms(year, month, day, hour, minute, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid birth time: %w", err)
	}

	ec := st.GetLunarHour().GetEightChar()

	result := &BaziResult{}

	// --- four pillars ---
	result.YearPillar = pillarFromSixtyCycle(ec.GetYear())
	result.MonthPillar = pillarFromSixtyCycle(ec.GetMonth())
	result.DayPillar = pillarFromSixtyCycle(ec.GetDay())
	result.HourPillar = pillarFromSixtyCycle(ec.GetHour())

	// MingGong from year gan, month zhi, hour zhi (《渊海子平》古法)
	mingGongGanZhi, err := CalcMingGong(result.YearPillar.Gan, result.MonthPillar.Zhi, result.HourPillar.Zhi)
	if err != nil {
		return nil, fmt.Errorf("计算命宫失败: %w", err)
	}
	result.MingGong = BuildMingGongDetail(mingGongGanZhi)
	// RiZhuDesc from day pillar (key format: dayGan + "日" + dayZhi e.g. "甲日甲子")
	riZhuKey := result.DayPillar.Gan + "日" + result.DayPillar.Zhi
	result.RiZhuDesc = SiZiSummaries[riZhuKey]

	// --- five elements scores ---
	result.FiveElements = calcFiveElements(&ec)
	result.ElementDetail = calcElementDetail(&ec)
	result.BodyStrength = calcBodyStrength(&ec)

	// --- ten gods ---
	result.TenGods = calcTenGods(&ec)

	// --- na yin ---
	result.NaYin = calcNaYin(&ec)

	// --- hidden stems ---
	result.HiddenStems = calcHiddenStems(&ec)

	// --- da yun ---
	result.DaYunInfo = calcDaYun(st, tymeGender)

	// --- clash / harmony ---
	result.ClashHarmony = calcClashHarmony(&ec)

	// --- gan/zhi analysis ---
	result.GanZhiAnalysis = CalcGanZhiAnalysis(
		result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar,
	)
	result.PatternAnalysis = analyzePatternExtended(
		[]model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar},
		result.MonthPillar.Zhi,
		result.FiveElements,
		result.BodyStrength,
	)

	// --- enrich pillar details ---
	result.TenGodProportion = calcTenGodProportion(&ec, result.DayPillar.Gan)
	analyzer := &TenGodAnalyzer{}
	result.TenGodAnalysis = analyzer.AnalyzeTenGod(result.TenGodProportion)
	s.enrichPillarDetails(result, month, gender)

	enrichRiZhuText(result)

	return result, nil
}

func (s *BaziService) enrichPillarDetails(result *BaziResult, birthMonth int, gender string) {
	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}

	// PillarDetails: one per pillar
	for _, p := range pillars {
		gIdx := GanIndex(p.Gan)
		zIdx := ZhiIndex(p.Zhi)
		nayinName := Nayin[gIdx][zIdx]
		entry := NaYinMap[nayinName]
		detail := PillarDetail{
			Stem:      p.Gan,
			Branch:    p.Zhi,
			ShengXiao: ShengXiao[zIdx],
			Empties:   Empties[gIdx][zIdx],
			Nayin: NaYinInfo{
				Name:        entry.Name,
				Element:     entry.Element,
				ImageDesc:   entry.ImageDesc,
				Personality: entry.Personality,
				EnergyStage: entry.EnergyStage,
				ModernExt:   entry.ModernExt,
				Judgments:   entry.Judgments,
			},
		}
		result.PillarDetails = append(result.PillarDetails, detail)
	}

	shenSha := calcShenShaByPillars(shenShaPillars{
		Year:   result.YearPillar,
		Month:  result.MonthPillar,
		Day:    result.DayPillar,
		Hour:   result.HourPillar,
		Gender: gender,
	})
	result.DayShenSha = shenSha.Day
	result.ShenShaByPillar = buildPillarShenSha(result, shenSha)
	result.GlobalShenSha = shenSha.Global
	result.ShenShaSummary = &ShenShaSummary{
		Title: "神煞排序说明",
		Description: []string{
			"日柱最重要：代表自身、配偶、中年运势，影响最直接、最持久。多数神煞以日干、日支为直接查法。",
			"年柱次之：代表祖上、童年、大环境，决定先天福荫。病符、官符、丧门、吊客等依年支而定。",
			"月柱辅助：代表父母、兄弟、青年期。天德、月德、天赦等特定神煞需参考月令。",
			"时柱辅助：代表子女、晚年、事业成果。童子煞、桃花等特定神煞有时需看时支。",
			"注意：不同神煞有不同查法依据，本模块按柱位优先级展示，个别神煞仍有其他特定查法。",
		},
	}

	// TiaoHou & JinBuHuan
	tiaoKey := result.DayPillar.Gan + result.DayPillar.Zhi
	result.DayStemTiaoHou = TiaoHou[tiaoKey]
	result.DayStemJinBuHuan = JinBuHuan[result.DayPillar.Gan]

	// SeasonText via YueTexts[dayGan][season]
	season := monthKey(birthMonth)
	if texts, ok := YueTexts[result.DayPillar.Gan]; ok {
		if txt, ok := texts[season]; ok {
			result.SeasonText = txt
		}
		// Also set month-specific text when available (5月 or 6月)
		if m := birthMonth; m == 5 || m == 6 {
			if txt, ok := texts[fmt.Sprintf("%d月", m)]; ok {
				result.SeasonTextMonth = txt
			}
		}
	}
}

// monthKey returns the best-matching YueTexts key for a given birth month.
// It prefers specific month anchors (正二月/五月/六月) where the data has them,
// and falls back to the generic season (春/夏/秋/冬).
func calcDayShenSha(dayGan, dayZhi string) []string {
	return calcDayShenShaOnly(dayGan, dayZhi)
}

func calcAllPillarShenSha(result *BaziResult) []PillarShenSha {
	calc := calcShenShaByPillars(shenShaPillars{
		Year:  result.YearPillar,
		Month: result.MonthPillar,
		Day:   result.DayPillar,
		Hour:  result.HourPillar,
	})
	return buildPillarShenSha(result, calc)
}

func buildPillarShenSha(result *BaziResult, calc shenShaCalcResult) []PillarShenSha {
	pillars := []struct {
		pillar   string
		label    string
		gan      string
		zhi      string
		priority int
		role     string
		items    []string
	}{
		{"day", "日柱", result.DayPillar.Gan, result.DayPillar.Zhi, 1, "自身·配偶·中年", calc.Day},
		{"year", "年柱", result.YearPillar.Gan, result.YearPillar.Zhi, 2, "祖上·童年·大环境", calc.Year},
		{"month", "月柱", result.MonthPillar.Gan, result.MonthPillar.Zhi, 3, "父母·兄弟·青年", calc.Month},
		{"hour", "时柱", result.HourPillar.Gan, result.HourPillar.Zhi, 4, "子女·晚年·事业成果", calc.Hour},
	}

	output := make([]PillarShenSha, 0, len(pillars))
	for _, p := range pillars {
		output = append(output, PillarShenSha{
			Pillar:   p.pillar,
			Label:    p.label,
			Gan:      p.gan,
			Zhi:      p.zhi,
			Priority: p.priority,
			Role:     p.role,
			Items:    p.items,
		})
	}
	return output
}

func calcGlobalShenSha(result *BaziResult) []string {
	calc := calcShenShaByPillars(shenShaPillars{
		Year:  result.YearPillar,
		Month: result.MonthPillar,
		Day:   result.DayPillar,
		Hour:  result.HourPillar,
	})
	return calc.Global
}

func monthKey(m int) string {
	switch m {
	case 1, 2:
		return "正二月"
	case 5, 6:
		return fmt.Sprintf("%d月", m)
	default:
		return seasonFromMonth(m)
	}
}

// seasonFromMonth maps birth month (1-12) to YueTexts season key.
func seasonFromMonth(m int) string {
	switch m {
	case 1, 2:
		return "春"
	case 3, 4:
		return "夏"
	case 5, 6, 7, 8:
		return "秋"
	default:
		return "冬"
	}
}

// --- helpers ---------------------------------------------------------------

func pillarFromSixtyCycle(sc tyme.SixtyCycle) model.Pillar {
	return model.Pillar{
		Gan: sc.GetHeavenStem().GetName(),
		Zhi: sc.GetEarthBranch().GetName(),
	}
}

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

func calcElementDetail(ec *tyme.EightChar) []ElementStrength {
	elements := []string{"木", "火", "土", "金", "水"}
	tianGan := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}
	zhiCangGan := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}
	cangGanMap := map[string][]string{"木": {}, "火": {}, "土": {}, "金": {}, "水": {}}

	pillars := [](func() tyme.SixtyCycle){
		ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour,
	}
	for _, fn := range pillars {
		sc := fn()
		elem := sc.GetHeavenStem().GetElement().GetName()
		tianGan[elem] += 5

		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			weight := 1
			label := sc.GetEarthBranch().GetName() + hhs.GetHeavenStem().GetName()
			if hhs.GetType() == tyme.MAIN {
				weight = 3
			} else if hhs.GetType() == tyme.MIDDLE {
				weight = 2
				label += "(中)"
			} else if hhs.GetType() == tyme.RESIDUAL {
				label += "(余)"
			}
			elem := hhs.GetHeavenStem().GetElement().GetName()
			zhiCangGan[elem] += weight
			cangGanMap[elem] = append(cangGanMap[elem], label)
		}
	}

	var result []ElementStrength
	for _, e := range elements {
		result = append(result, ElementStrength{
			Element:     e,
			TianGan:     tianGan[e],
			ZhiCangGan:  zhiCangGan[e],
			Total:       tianGan[e] + zhiCangGan[e],
			CangGanList: cangGanMap[e],
		})
	}
	return result
}

func calcBodyStrength(ec *tyme.EightChar) BodyStrengthResult {
	return calcBodyStrengthV2(ec)
}

var elementIdx = map[string]int{"木": 0, "火": 1, "土": 2, "金": 3, "水": 4}

var tianGanMap = map[string]struct {
	WuXing string
}{
	"甲": {"木"}, "乙": {"木"},
	"丙": {"火"}, "丁": {"火"},
	"戊": {"土"}, "己": {"土"},
	"庚": {"金"}, "辛": {"金"},
	"壬": {"水"}, "癸": {"水"},
}

// yueLingMatrix: rows = day element (木火土金水), cols = month branch element
// 旺(3) 同我, 相(2) 我生, 休(1) 生我, 囚(0) 克我, 死(0) 我克
var yueLingMatrix = [5][5]float64{
	// 木   火   土   金   水   ← 月支五行
	{3, 2, 0, 0, 1}, // 木日主: 旺(木) 相(火) 死(土) 囚(金) 休(水)
	{1, 3, 2, 0, 0}, // 火日主: 休(木) 旺(火) 相(土) 死(金) 囚(水)
	{0, 1, 3, 2, 0}, // 土日主: 囚(木) 休(火) 旺(土) 相(金) 死(水)
	{0, 0, 1, 3, 2}, // 金日主: 死(木) 囚(火) 休(土) 旺(金) 相(水)
	{2, 0, 0, 1, 3}, // 水日主: 相(木) 死(火) 囚(土) 休(金) 旺(水)
}

func getYueLingScore(dayElem string, monthBranchElem string) float64 {
	di := elementIdx[dayElem]
	mi := elementIdx[monthBranchElem]
	return yueLingMatrix[di][mi]
}

// isSupport returns true if gan's element supports (比劫/印星) the day master.
func isSupport(gan string, dayElem string) bool {
	tg, ok := tianGanMap[gan]
	if !ok {
		return false
	}
	// 比劫：同五行
	if tg.WuXing == dayElem {
		return true
	}
	// 印星：生我者（木生火→火日主，壬癸生木→木日主...）
	// 生我者：木生火、火生土、土生金、金生水、水生木
	supporter := map[string]string{
		"木": "火", "火": "土", "土": "金", "金": "水", "水": "木",
	}
	if supporter[tg.WuXing] == dayElem {
		return true
	}
	return false
}

// isRestrict returns true if gan's element restricts (克泄耗) the day master.
func isRestrict(gan string, dayElem string) bool {
	tg, ok := tianGanMap[gan]
	if !ok {
		return false
	}
	if tg.WuXing == dayElem {
		return false // 同五行已由 isSupport 处理
	}
	// 克我: gan克day
	// 我生(泄): day生gan
	// 我克(耗): day克gan
	ke := map[string]string{"木": "土", "火": "金", "土": "水", "金": "木", "水": "火"}
	sheng := map[string]string{"木": "火", "火": "土", "土": "金", "金": "水", "水": "木"}
	return ke[tg.WuXing] == dayElem ||        // 克我
		sheng[dayElem] == tg.WuXing ||        // 我生(泄)
		ke[dayElem] == tg.WuXing              // 我克(耗)
}

// zangGanWeight returns the藏干 weight for a given earth branch position.
func zangGanWeight(hsType tyme.HideHeavenStemType) float64 {
	switch hsType {
	case tyme.MAIN:
		return 0.6
	case tyme.MIDDLE:
		return 0.3
	case tyme.RESIDUAL:
		return 0.1
	}
	return 0.0
}

func calcBodyStrengthV2(ec *tyme.EightChar) BodyStrengthResult {
	dayStem := ec.GetDay().GetHeavenStem()
	dayElem := dayStem.GetElement().GetName()

	monthBranch := ec.GetMonth().GetEarthBranch()
	monthElem := monthBranch.GetElement().GetName()

	// 1. 得令
	lingScore := getYueLingScore(dayElem, monthElem)

	// 2. 得地：四柱地支藏干中比劫/印星 × weight × 1.5
	diScore := 0.0
	pillars := [](func() tyme.SixtyCycle){ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour}
	for _, fn := range pillars {
		for _, hhs := range fn().GetEarthBranch().GetHideHeavenStems() {
			if isSupport(hhs.GetHeavenStem().GetName(), dayElem) {
				diScore += zangGanWeight(hhs.GetType()) * 1.5
			}
		}
	}

	// 3. 得势：生扶 - 克泄耗（天干非日干 + 藏干）
	supportWeight := 0.0
	restrictWeight := 0.0
	for i, fn := range pillars {
		gan := fn().GetHeavenStem().GetName()
		if i == 2 { // 跳过日干本身
			continue
		}
		if isSupport(gan, dayElem) {
			supportWeight += 1.0
		} else if isRestrict(gan, dayElem) {
			restrictWeight += 1.0
		}
	}
	for _, fn := range pillars {
		for _, hhs := range fn().GetEarthBranch().GetHideHeavenStems() {
			gan := hhs.GetHeavenStem().GetName()
			if isSupport(gan, dayElem) {
				supportWeight += zangGanWeight(hhs.GetType())
			} else if isRestrict(gan, dayElem) {
				restrictWeight += zangGanWeight(hhs.GetType())
			}
		}
	}
	shiScore := supportWeight - restrictWeight

	// 4. 得生
	shengScore := 0.0
	// 天干印星
	for i, fn := range pillars {
		if i == 2 {
			continue
		}
		tg := fn().GetHeavenStem()
		tgElem := tg.GetElement().GetName()
		// 生我者
		if (tgElem == "木" && dayElem == "火") ||
			(tgElem == "火" && dayElem == "土") ||
			(tgElem == "土" && dayElem == "金") ||
			(tgElem == "金" && dayElem == "水") ||
			(tgElem == "水" && dayElem == "木") {
			shengScore += 1.0
		}
	}
	// 地支藏干印星
	for _, fn := range pillars {
		for _, hhs := range fn().GetEarthBranch().GetHideHeavenStems() {
			tgElem := hhs.GetHeavenStem().GetElement().GetName()
			if (tgElem == "木" && dayElem == "火") ||
				(tgElem == "火" && dayElem == "土") ||
				(tgElem == "土" && dayElem == "金") ||
				(tgElem == "金" && dayElem == "水") ||
				(tgElem == "水" && dayElem == "木") {
				shengScore += zangGanWeight(hhs.GetType())
			}
		}
	}
	if shengScore > 0 {
		shengScore = 1.0
	}

	// 总分
	totalScore := lingScore*3 + diScore*2 + shiScore*1 + shengScore*1
	if totalScore < 0 {
		totalScore = 0
	}

	var verdict string
	var like, dislike []string
	if totalScore > 5.0 {
		verdict = "身旺"
	} else {
		verdict = "身弱"
	}

	// 根据日主五行动态计算喜忌
	// 身旺: 喜克泄耗(克我+我生+我克), 忌生扶(生我+同我)
	// 身弱: 喜生扶(生我+同我), 忌克泄耗(克我+我生+我克)
	allElems := []string{"木", "火", "土", "金", "水"}
	idx := elementIdx[dayElem]
	sameElem := allElems[idx]             // 同我(比劫)
	supportElem := allElems[(idx+4)%5]    // 生我(印星)
	drainElem := allElems[(idx+1)%5]      // 我生(食伤)
	controlElem := allElems[(idx+3)%5]    // 克我(官杀)
	wealthElem := allElems[(idx+2)%5]     // 我克(财)
	if verdict == "身旺" {
		like = []string{controlElem, drainElem, wealthElem}
		dislike = []string{supportElem, sameElem}
	} else {
		like = []string{supportElem, sameElem}
		dislike = []string{controlElem, drainElem, wealthElem}
	}

	return BodyStrengthResult{
		Verdict:    verdict,
		Like:       like,
		Dislike:    dislike,
		TotalScore: totalScore,
		LingScore:  lingScore,
		DiScore:    diScore,
		ShiScore:   shiScore,
		ShengScore: shengScore,
	}
}

func calcTenGods(ec *tyme.EightChar) map[string]string {
	dayStem := ec.GetDay().GetHeavenStem()
	return map[string]string{
		"year":  dayStem.GetTenStar(ec.GetYear().GetHeavenStem()).GetName(),
		"month": dayStem.GetTenStar(ec.GetMonth().GetHeavenStem()).GetName(),
		"day":   "日主",
		"hour":  dayStem.GetTenStar(ec.GetHour().GetHeavenStem()).GetName(),
	}
}

func calcNaYin(ec *tyme.EightChar) map[string]NaYinInfo {
	pillars := []struct {
		key string
		fn  func() tyme.SixtyCycle
	}{
		{"year", ec.GetYear}, {"month", ec.GetMonth},
		{"day", ec.GetDay}, {"hour", ec.GetHour},
	}
	result := make(map[string]NaYinInfo, 4)
	for _, p := range pillars {
		nayinName := Nayin[GanIndex(p.fn().GetHeavenStem().GetName())][ZhiIndex(p.fn().GetEarthBranch().GetName())]
		entry := NaYinMap[nayinName]
		result[p.key] = NaYinInfo{
			Name:        entry.Name,
			Element:     entry.Element,
			ImageDesc:   entry.ImageDesc,
			Personality: entry.Personality,
			EnergyStage: entry.EnergyStage,
			ModernExt:   entry.ModernExt,
			Judgments:   entry.Judgments,
		}
	}
	return result
}

func calcHiddenStems(ec *tyme.EightChar) map[string][]string {
	result := make(map[string][]string, 4)
	pillars := map[string]func() tyme.SixtyCycle{
		"year": ec.GetYear, "month": ec.GetMonth,
		"day": ec.GetDay, "hour": ec.GetHour,
	}
	for name, fn := range pillars {
		sc := fn()
		var stems []string
		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			label := sc.GetEarthBranch().GetName() + hhs.GetHeavenStem().GetName()
			switch hhs.GetType() {
			case tyme.MAIN:
				// keep as-is
			case tyme.MIDDLE:
				label += "(中)"
			case tyme.RESIDUAL:
				label += "(余)"
			}
			stems = append(stems, label)
		}
		result[name] = stems
	}
	return result
}

func calcDaYun(st *tyme.SolarTime, gender tyme.Gender) DaYunInfo {
	cl := tyme.ChildLimit{}.FromSolarTime(*st, gender)

	dir := "逆排"
	if cl.IsForward() {
		dir = "顺排"
	}

	daYun := DaYunInfo{
		StartAge:  cl.GetYearCount(),
		Direction: dir,
	}

	df := cl.GetStartDecadeFortune()
	for i := 0; i < 8; i++ {
		cur := df.Next(i)
		sx := cur.GetSixtyCycle()
		daYun.Pillars = append(daYun.Pillars, model.Pillar{
			Gan: sx.GetHeavenStem().GetName(),
			Zhi: sx.GetEarthBranch().GetName(),
		})
	}
	return daYun
}

// --- clash / harmony detection ---------------------------------------------

type pillarPair struct {
	name   string
	branch tyme.EarthBranch
}

func calcClashHarmony(ec *tyme.EightChar) []ClashRelation {
	pairs := []pillarPair{
		{"年柱", ec.GetYear().GetEarthBranch()},
		{"月柱", ec.GetMonth().GetEarthBranch()},
		{"日柱", ec.GetDay().GetEarthBranch()},
		{"时柱", ec.GetHour().GetEarthBranch()},
	}

	var relations []ClashRelation

	// pairwise
	for i := 0; i < len(pairs); i++ {
		for j := i + 1; j < len(pairs); j++ {
			a, b := pairs[i], pairs[j]

			if a.branch.GetOpposite().Equals(b.branch) {
				relations = append(relations, ClashRelation{a.name, b.name, "六冲"})
			}
			if a.branch.GetCombine().Equals(b.branch) {
				relations = append(relations, ClashRelation{a.name, b.name, "六合"})
			}
			if a.branch.GetHarm().Equals(b.branch) {
				relations = append(relations, ClashRelation{a.name, b.name, "六害"})
			}

			// 三刑
			if t := tortureType(a.branch, b.branch); t != "" {
				relations = append(relations, ClashRelation{a.name, b.name, t})
			}
		}
	}

	// 三会: check all 4 branches for 3-branch groups
	relations = append(relations, detectTripleMeetings(pairs)...)

	return relations
}

// tortureType checks if two branches form a 三刑 relation.
func tortureType(a, b tyme.EarthBranch) string {
	defer func() { recover() }() // guard invalid index panics

	aName := a.GetName()
	bName := b.GetName()

	// 无礼之刑: 子-卯
	if (aName == "子" && bName == "卯") || (aName == "卯" && bName == "子") {
		return "无礼之刑"
	}
	// 无恩之刑: 丑-戌, 戌-未, 未-丑
	wuEn := [][]string{{"丑", "戌"}, {"戌", "未"}, {"未", "丑"}}
	for _, pair := range wuEn {
		if (aName == pair[0] && bName == pair[1]) || (aName == pair[1] && bName == pair[0]) {
			return "无恩之刑"
		}
	}
	// 恃势之刑: 寅-巳, 巳-申, 申-寅
	shiShi := [][]string{{"寅", "巳"}, {"巳", "申"}, {"申", "寅"}}
	for _, pair := range shiShi {
		if (aName == pair[0] && bName == pair[1]) || (aName == pair[1] && bName == pair[0]) {
			return "恃势之刑"
		}
	}
	// 自刑: 辰-辰, 午-午, 酉-酉, 亥-亥
	selfTorture := map[string]bool{"辰": true, "午": true, "酉": true, "亥": true}
	if aName == bName && selfTorture[aName] {
		return "自刑"
	}
	return ""
}

// detectTripleMeetings detects 三会 (three-branch meeting of same direction element).
func detectTripleMeetings(pairs []pillarPair) []ClashRelation {
	// 三会局: 寅卯辰(木), 巳午未(火), 申酉戌(金), 亥子丑(水)
	tripleGroups := [][]string{
		{"寅", "卯", "辰"}, // 东方木
		{"巳", "午", "未"}, // 南方火
		{"申", "酉", "戌"}, // 西方金
		{"亥", "子", "丑"}, // 北方水
	}

	var relations []ClashRelation
	branchIndex := make(map[string]int)
	for i, p := range pairs {
		n := p.branch.GetName()
		branchIndex[n] = i
	}

	seen := make(map[string]bool) // deduplicate

	for _, group := range tripleGroups {
		var matched []int
		for _, b := range group {
			if idx, ok := branchIndex[b]; ok {
				matched = append(matched, idx)
			}
		}
		if len(matched) >= 3 {
			// generate pairwise relations among the three
			for i := 0; i < len(matched); i++ {
				for j := i + 1; j < len(matched); j++ {
					pi, pj := pairs[matched[i]], pairs[matched[j]]
					key := pi.name + "<>" + pj.name + "<>三会"
					if key2 := pj.name + "<>" + pi.name + "<>三会"; seen[key2] {
						continue
					}
					if !seen[key] {
						seen[key] = true
						relations = append(relations, ClashRelation{pi.name, pj.name, "三会"})
					}
				}
			}
		}
	}
	return relations
}

// ganInfo holds element and yang flag for a stem name.
type ganInfo struct {
	elem string
	yang bool
}

// ganInfoOf returns element and yang flag for a stem name.
func ganInfoOf(name string) ganInfo {
	gans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	elems := []string{"木", "木", "火", "火", "土", "土", "金", "金", "水", "水"}
	yangs := []bool{true, false, true, false, true, false, true, false, true, false}
	for i, g := range gans {
		if g == name {
			return ganInfo{elems[i], yangs[i]}
		}
	}
	return ganInfo{"", false}
}

// dayGanElement returns the element name for the day stem.
func dayGanElement(dayGan string) string {
	return ganInfoOf(dayGan).elem
}

// classifyTenGod classifies a stem (stemName) against the day stem (dayGan),
// returning the ten-god name string.
// Only the VISIBLE day pillar stem itself is treated as 日主 (self) and excluded.
// Hidden stems that happen to share the same name as the day pillar stem
// are classified normally (比肩/劫财 when same element).
func classifyTenGod(stemName string, dayGan string, isDayPillarStem bool) string {
	stem := ganInfoOf(stemName)
	day := ganInfoOf(dayGan)

	// Only the visible day pillar stem is 日主 — hidden stems use normal classification
	if isDayPillarStem && stemName == dayGan {
		return "日主"
	}

	if stem.elem == day.elem {
		// Same element → 比肩 (same yang) or 劫财 (diff yang)
		if stem.yang == day.yang {
			return "比肩"
		}
		return "劫财"
	}
	// Stem element is "child" of day element (泄) → target: 食神/伤官
	// Element cycle: 木→火→土→金→水→木 (each is "child" of previous)
	child := map[string]string{
		"木": "火", "火": "土", "土": "金", "金": "水", "水": "木",
	}
	if child[day.elem] == stem.elem {
		if stem.yang == day.yang {
			return "食神"
		}
		return "伤官"
	}
	// Stem element is "parent" of day element (生) → 正印/偏印
	parent := map[string]string{
		"火": "木", "土": "火", "金": "土", "水": "金", "木": "水",
	}
	if parent[day.elem] == stem.elem {
		if stem.yang == day.yang {
			return "偏印"
		}
		return "正印"
	}
	// Stem element "controls" day element (克) → 正官/七杀
	ke := map[string]string{
		"木": "金", "火": "水", "土": "木", "金": "火", "水": "土",
	}
	if ke[day.elem] == stem.elem {
		if stem.yang == day.yang {
			return "七杀"
		}
		return "正官"
	}
	// Stem element is "controlled" by day element (耗) → 正财/偏财
	hao := map[string]string{
		"金": "木", "水": "火", "木": "土", "火": "金", "土": "水",
	}
	if hao[day.elem] == stem.elem {
		if stem.yang == day.yang {
			return "偏财"
		}
		return "正财"
	}
	return ""
}

// calcTenGodProportion computes the ten-god proportion using simple counting.
// It counts: year/month/hour stems (3) + all hidden stems from 4 branches (via tyme library).
// Returns 10 TenGodRatio in fixed order.
func calcTenGodProportion(ec *tyme.EightChar, dayGan string) []TenGodRatio {
	counts := make(map[string]int)

	// 3 visible stems (not day) — day stem is the visible one so isDayPillarStem=true only for day
	for _, fn := range [](struct {
		fn  func() tyme.SixtyCycle
		day bool
	}){
		{ec.GetYear, false},
		{ec.GetMonth, false},
		{ec.GetHour, false},
	} {
		sc := fn.fn()
		stemName := sc.GetHeavenStem().GetName()
		god := classifyTenGod(stemName, dayGan, fn.day)
		if god != "" && god != "日主" {
			counts[god]++
		}
	}

	// Hidden stems from 4 branches — all are from branch (not visible day stem) so isDayPillarStem=false
	for _, fn := range [](func() tyme.SixtyCycle){ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour} {
		sc := fn()
		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			stemName := hhs.GetHeavenStem().GetName()
			god := classifyTenGod(stemName, dayGan, false)
			if god != "" && god != "日主" {
				counts[god]++
			}
		}
	}

	total := 0
	for _, c := range counts {
		total += c
	}

	var result []TenGodRatio
	for _, name := range tenGodNames {
		c := counts[name]
		pct := 0.0
		if total > 0 {
			pct = math.Round(float64(c)*10000/float64(total)) / 100 // 2 decimal places
		}
		result = append(result, TenGodRatio{Name: name, Count: c, Percent: pct})
	}
	return result
}

// enrichRiZhuText splits the SiZiSummaries text into poem/source/comment/hourDetail.
func enrichRiZhuText(result *BaziResult) {
	key := result.DayPillar.Gan + "日" + result.DayPillar.Zhi
	text := SiZiSummaries[key]

	// Split by " # " to get all segments
	allParts := strings.Split(text, " # ")
	if len(allParts) == 1 {
		// No " # " found, whole text is the poem
		result.RiZhuPoem = text
		return
	}

	result.RiZhuPoem = allParts[0]
	if len(allParts) >= 2 {
		result.RiZhuSource = allParts[1]
	}
	if len(allParts) >= 3 {
		result.RiZhuComment = allParts[2]
	}
	if len(allParts) >= 4 {
		result.RiZhuHourDetail = strings.Join(allParts[3:], " # ")
	}
}

func toTymeGender(gender string) (tyme.Gender, error) {
	switch gender {
	case "MALE":
		return tyme.MAN, nil
	case "FEMALE":
		return tyme.WOMAN, nil
	default:
		return 0, fmt.Errorf("invalid gender %q: must be MALE or FEMALE", gender)
	}
}
