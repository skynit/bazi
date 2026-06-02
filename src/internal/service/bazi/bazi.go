package bazi

import (
	"fmt"
	"math"
	"strings"

	"bazi/internal/model"
	"bazi/internal/service/data"
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

// PillarTenGods tracks which ten gods appear in each pillar position.
type PillarTenGods struct {
	Year  []string `json:"year"`
	Month []string `json:"month"`
	Day   []string `json:"day"`
	Hour  []string `json:"hour"`
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
	MingGong         data.MingGongDetail      `json:"ming_gong"`
	RiZhuDesc        string               `json:"ri_zhu_desc"`
	PillarDetails    []PillarDetail       `json:"pillar_details"`
	DayStemTiaoHou   string               `json:"tiao_hou"`
	DayStemJinBuHuan string               `json:"jin_bu_huan"`
	DayShenSha       []string             `json:"day_shen_sha"`
	ShenShaByPillar  []PillarShenSha      `json:"shen_sha_by_pillar"`
	GlobalShenSha    []string             `json:"global_shen_sha"`
	ShenShaSummary   *ShenShaSummary      `json:"shen_sha_summary"`
	SeasonText       string               `json:"season_text"`
	SeasonTextMonth  string               `json:"season_text_month"` // month-specific data.YueTexts when available
	TenGodProportion []TenGodRatio        `json:"ten_god_proportion"`
	TenGodAnalysis   *TenGodAnalysis      `json:"ten_god_analysis"`
	RiZhuPoem        string               `json:"ri_zhu_poem"`
	RiZhuSource      string               `json:"ri_zhu_source"`
	RiZhuComment     string               `json:"ri_zhu_comment"`
	RiZhuHourDetail  string               `json:"ri_zhu_hour_detail"`
	JiaZiDetail      *data.JiaZiEntry            `json:"jia_zi_detail"`       // 《三命通会》60甲子性质
	WuxingSeasonNote string                     `json:"wuxing_season_note"`  // 五行四时论分析
	HealthNote       string                     `json:"health_note"`         // 五行疾病提示
	Tiaohou          *TiaohouResult             `json:"tiaohou"`             // 《穷通宝鉴》调候用神分析
	WuXingFlow       data.WuXingFlowAnalysis    `json:"wuxing_flow"`         // 五行流通分析
	TongGuan         data.TongGuanAnalysis      `json:"tong_guan"`           // 通关用神
	MissingElements  data.MissingElementAnalysis `json:"missing_elements"`    // 缺失五行
	FlowPatternDesc  string                     `json:"flow_pattern_desc"`   // 流通格局描述
	DaYunFlow        []data.DaYunFlowItem        `json:"dayun_flow"`          // 大运流通影响
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
	mingGongGanZhi, err := data.CalcMingGong(result.YearPillar.Gan, result.MonthPillar.Zhi, result.HourPillar.Zhi)
	if err != nil {
		return nil, fmt.Errorf("计算命宫失败: %w", err)
	}
	result.MingGong = data.BuildMingGongDetail(mingGongGanZhi)
	// RiZhuDesc from day pillar (key format: dayGan + "日" + dayZhi e.g. "甲日甲子")
	riZhuKey := result.DayPillar.Gan + "日" + result.DayPillar.Zhi
	result.RiZhuDesc = data.SiZiSummaries[riZhuKey]

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
	result.PatternAnalysis = AnalyzePatternExtended(
		[]model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar},
		result.MonthPillar.Zhi,
		result.FiveElements,
		result.BodyStrength,
	)

	// --- enrich pillar details ---
	dayElem := data.GanElement[result.DayPillar.Gan]
	result.TenGodProportion = calcTenGodProportion(&ec, result.DayPillar.Gan)
	pillarTenGods := calcPillarTenGods(&ec, result.DayPillar.Gan)
	analyzer := &TenGodAnalyzer{}
	result.TenGodAnalysis = analyzer.AnalyzeTenGod(result.TenGodProportion, dayElem, result.BodyStrength, pillarTenGods, gender)
	s.enrichPillarDetails(result, month, gender)

	enrichRiZhuText(result)

	// --- enrich from 《三命通会》knowledge ---
	enrichWuxingSeason(result, month)
	enrichJiaZiDetail(result)
	enrichHealthNote(result)

	// --- enrich from 《穷通宝鉴》tiaohou analysis ---
	tiaohouResult, _ := AnalyzeTiaohou(result.DayPillar.Gan, result.MonthPillar.Zhi)
	result.Tiaohou = tiaohouResult

	// --- enrich from 《滴天髓》流通分析 ---
	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}
	result.WuXingFlow = data.AnalyzeWuXingFlowV2(result.FiveElements, dayElem)
	result.TongGuan = data.FindTongGuan(pillars, dayElem, result.MonthPillar.Zhi)
	result.MissingElements = data.FindMissingElements(result.FiveElements)
	result.FlowPatternDesc = data.BuildFlowPatternDesc(result.WuXingFlow, result.TongGuan, result.MissingElements)

	// --- 大运流通分析 ---
	if len(result.DaYunInfo.Pillars) > 0 {
		result.DaYunFlow = data.CalcDaYunFlow(result.DayPillar.Gan, result.FiveElements, result.DaYunInfo.Pillars, result.DaYunInfo.StartAge)
	}

	return result, nil
}

func (s *BaziService) enrichPillarDetails(result *BaziResult, birthMonth int, gender string) {
	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}

	// PillarDetails: one per pillar
	for _, p := range pillars {
		gIdx := data.GanIndex(p.Gan)
		zIdx := data.ZhiIndex(p.Zhi)
		nayinName := data.Nayin[gIdx][zIdx]
		entry := data.NaYinMap[nayinName]
		detail := PillarDetail{
			Stem:      p.Gan,
			Branch:    p.Zhi,
			ShengXiao: data.ShengXiao[zIdx],
			Empties:   data.Empties[gIdx][zIdx],
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

	shenSha := CalcShenShaByPillars(ShenShaPillars{
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

	// data.TiaoHou & data.JinBuHuan
	// 调候用神按日干+月支查（非日干+日支）
	tiaoKey := result.DayPillar.Gan + result.MonthPillar.Zhi
	result.DayStemTiaoHou = data.TiaoHou[tiaoKey]
	result.DayStemJinBuHuan = data.JinBuHuan[result.DayPillar.Gan]

	// SeasonText via data.YueTexts[dayGan][season]
	season := monthKey(birthMonth)
	if texts, ok := data.YueTexts[result.DayPillar.Gan]; ok {
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

// monthKey returns the best-matching data.YueTexts key for a given birth month.
// It prefers specific month anchors (正二月/五月/六月) where the data has them,
// and falls back to the generic season (春/夏/秋/冬).
func buildPillarShenSha(result *BaziResult, calc ShenShaCalcResult) []PillarShenSha {
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

func monthKey(m int) string {
	switch m {
	case 1, 2:
		return "正二月"
	case 5, 6:
		return fmt.Sprintf("%d月", m)
	default:
		return data.SeasonFromMonth(m)
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
// 经典依据：日主强弱判断"囚：日主克他力量较弱；死：日主被克力量最弱"
// 旺(3) 同我, 相(2) 我生, 休(1) 生我, 囚(0.5) 我克, 死(0) 克我
var yueLingMatrix = [5][5]float64{
	// 木   火   土   金   水   ← 月支五行
	{3, 2, 0, 0.5, 1}, // 木日主: 旺(木) 相(火) 死(土) 囚(金) 休(水)
	{1, 3, 2, 0, 0.5}, // 火日主: 休(木) 旺(火) 相(土) 死(金) 囚(水)
	{0.5, 1, 3, 2, 0}, // 土日主: 囚(木) 休(火) 旺(土) 相(金) 死(水)
	{0, 0.5, 1, 3, 2}, // 金日主: 死(木) 囚(火) 休(土) 旺(金) 相(水)
	{2, 0, 0.5, 1, 3}, // 水日主: 相(木) 死(火) 囚(土) 休(金) 旺(水)
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

// isSameElement returns true if gan's element is the same as day master (比劫 only).
// 经典依据：滴天髓"通根者如甲木见寅卯"，通根仅指同五行
func isSameElement(gan string, dayElem string) bool {
	tg, ok := tianGanMap[gan]
	if !ok {
		return false
	}
	return tg.WuXing == dayElem
}

// isYinStar returns true if gan's element is印星 (生我者).
func isYinStar(gan string, dayElem string) bool {
	tg, ok := tianGanMap[gan]
	if !ok {
		return false
	}
	supporter := map[string]string{
		"木": "火", "火": "土", "土": "金", "金": "水", "水": "木",
	}
	return supporter[tg.WuXing] == dayElem
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

// changShengMap 十二长生阶段查询表（阳干顺行）
// 经典依据：滴天髓"长生帝旺，得气最厚；冠带临官，得气次之"
var changShengMap = map[string]map[string]string{
	"木": {"亥": "长生", "子": "沐浴", "丑": "冠带", "寅": "临官", "卯": "帝旺", "辰": "衰", "巳": "病", "午": "死", "未": "墓", "申": "绝", "酉": "胎", "戌": "养"},
	"火": {"寅": "长生", "卯": "沐浴", "辰": "冠带", "巳": "临官", "午": "帝旺", "未": "衰", "申": "病", "酉": "死", "戌": "墓", "亥": "绝", "子": "胎", "丑": "养"},
	"土": {"寅": "长生", "卯": "沐浴", "辰": "冠带", "巳": "临官", "午": "帝旺", "未": "衰", "申": "病", "酉": "死", "戌": "墓", "亥": "绝", "子": "胎", "丑": "养"},
	"金": {"巳": "长生", "午": "沐浴", "未": "冠带", "申": "临官", "酉": "帝旺", "戌": "衰", "亥": "病", "子": "死", "丑": "墓", "寅": "绝", "卯": "胎", "辰": "养"},
	"水": {"申": "长生", "酉": "沐浴", "戌": "冠带", "亥": "临官", "子": "帝旺", "丑": "衰", "寅": "病", "卯": "死", "辰": "墓", "巳": "绝", "午": "胎", "未": "养"},
}

// changShengWeight returns the twelve-stage weight for a given day element and branch.
// 长生/帝旺/临官 → 1.5, 沐浴/冠带/衰/墓 → 1.0, 胎/养/病/死 → 0.5, 绝 → 0
func changShengWeight(dayElem, branch string) float64 {
	stages, ok := changShengMap[dayElem]
	if !ok {
		return 1.0
	}
	stage, ok := stages[branch]
	if !ok {
		return 1.0
	}
	switch stage {
	case "长生", "帝旺", "临官":
		return 1.5
	case "沐浴", "冠带", "衰", "墓":
		return 1.0
	case "胎", "养", "病", "死":
		return 0.5
	case "绝":
		return 0.0
	}
	return 1.0
}

func calcBodyStrengthV2(ec *tyme.EightChar) BodyStrengthResult {
	dayStem := ec.GetDay().GetHeavenStem()
	dayElem := dayStem.GetElement().GetName()

	monthBranch := ec.GetMonth().GetEarthBranch()
	monthElem := monthBranch.GetElement().GetName()

	// 1. 得令
	lingScore := getYueLingScore(dayElem, monthElem)

	// 2. 得地：四柱地支藏干中仅统计同五行（通根），印星归入得势
	// 经典依据：滴天髓"通根者如甲木见寅卯"，通根仅指同五行，印星归入得势
	diScore := 0.0
	pillars := [](func() tyme.SixtyCycle){ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour}
	for _, fn := range pillars {
		branch := fn().GetEarthBranch()
		branchName := branch.GetName()
		csW := changShengWeight(dayElem, branchName)
		for _, hhs := range branch.GetHideHeavenStems() {
			if isSameElement(hhs.GetHeavenStem().GetName(), dayElem) {
				diScore += zangGanWeight(hhs.GetType()) * 1.5 * csW
			}
		}
	}

	// 3. 得势：天干（年月时三干）+ 日支藏干（×1.5，坐下有根最贴身）
	// 经典依据：渊海子平"天干有比劫帮身有印绶生身"，力量有层级差异
	// 经典依据：滴天髓"坐下有根最贴身"，日支藏干给予更高权重
	supportWeight := 0.0
	restrictWeight := 0.0
	for i, fn := range pillars {
		gan := fn().GetHeavenStem().GetName()
		if i == 2 { // 跳过日干本身
			continue
		}
		tg, ok := tianGanMap[gan]
		if !ok {
			continue
		}
		if tg.WuXing == dayElem {
			// 比肩 1.0，劫财 0.8（阴阳异，助力稍弱）
			if GanInfoOf(gan).yang == GanInfoOf(dayStem.GetName()).yang {
				supportWeight += 1.0 // 比肩
			} else {
				supportWeight += 0.8 // 劫财
			}
		} else if isRestrict(gan, dayElem) {
			// 官杀 -1.2，食伤 -0.8，财 -0.6
			godName := ClassifyTenGod(gan, dayStem.GetName(), false)
			switch godName {
			case "正官", "七杀":
				restrictWeight += 1.2
			case "食神", "伤官":
				restrictWeight += 0.8
			case "正财", "偏财":
				restrictWeight += 0.6
			default:
				restrictWeight += 1.0
			}
		}
	}
	// 日支藏干：坐下有根最贴身，权重×1.5
	dayBranch := ec.GetDay().GetEarthBranch()
	for _, hhs := range dayBranch.GetHideHeavenStems() {
		hiddenGan := hhs.GetHeavenStem().GetName()
		w := zangGanWeight(hhs.GetType()) * 1.5
		tg, ok := tianGanMap[hiddenGan]
		if !ok {
			continue
		}
		if tg.WuXing == dayElem {
			if GanInfoOf(hiddenGan).yang == GanInfoOf(dayStem.GetName()).yang {
				supportWeight += w // 比肩
			} else {
				supportWeight += w * 0.8 // 劫财
			}
		} else if isRestrict(hiddenGan, dayElem) {
			godName := ClassifyTenGod(hiddenGan, dayStem.GetName(), false)
			switch godName {
			case "正官", "七杀":
				restrictWeight += w * 1.2
			case "食神", "伤官":
				restrictWeight += w * 0.8
			case "正财", "偏财":
				restrictWeight += w * 0.6
			default:
				restrictWeight += w * 1.0
			}
		}
	}
	shiScore := supportWeight - restrictWeight

	// 4. 得生：地支藏干中印星归入此处（与通根区分）
	shengScore := 0.0
	// 天干印星
	for i, fn := range pillars {
		if i == 2 {
			continue
		}
		tg := fn().GetHeavenStem()
		if isYinStar(tg.GetName(), dayElem) {
			shengScore += 1.0
		}
	}
	// 地支藏干印星
	for _, fn := range pillars {
		for _, hhs := range fn().GetEarthBranch().GetHideHeavenStems() {
			if isYinStar(hhs.GetHeavenStem().GetName(), dayElem) {
				shengScore += zangGanWeight(hhs.GetType())
			}
		}
	}

	// 总分：归一化后按经典权重加权
	// 经典依据：日主强弱判断"得令50%得地25%得势15%得气10%"
	normLing := lingScore / 3.0                            // lingScore最大3
	normDi := diScore / 7.0                                // diScore理论最大约7
	normShi := 1.0 / (1.0 + math.Exp(-shiScore/2.0))      // sigmoid归一化，避免极端值
	normSheng := 1.0 / (1.0 + math.Exp(-shengScore/2.0))  // sigmoid归一化
	totalScore := normLing*0.50 + normDi*0.25 + normShi*0.15 + normSheng*0.10

	var verdict string
	var like, dislike []string
	// 经典依据：滴天髓"身旺者日主得令得地得势三者居其二以上"
	switch {
	case totalScore > 0.60:
		verdict = "身旺"
	case totalScore > 0.55:
		verdict = "偏旺"
	case totalScore > 0.45:
		verdict = "中和"
	case totalScore > 0.40:
		verdict = "偏弱"
	default:
		verdict = "身弱"
	}

	// 5. 后验修正："得令不旺失令不衰"
	// 经典依据：滴天髓"春木虽强金太重而木亦危"
	// 如果得令五行被克它的五行总力超过自身力量的60%，则降低其旺度20%
	restrainingMap := map[string]string{
		"木": "金", "火": "水", "土": "木", "金": "火", "水": "土",
	}
	supportingMap := map[string]string{
		"木": "水", "火": "木", "土": "火", "金": "土", "水": "金",
	}
	if lingScore >= 2.0 { // 得令
		reElem := restrainingMap[dayElem]
		restrainingForce := 0.0
		selfForce := lingScore + diScore + shengScore
		for _, fn := range pillars {
			sc := fn()
			stElem := sc.GetHeavenStem().GetElement().GetName()
			if stElem == reElem {
				restrainingForce += 5.0
			}
			for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
				if hhs.GetHeavenStem().GetElement().GetName() == reElem {
					restrainingForce += zangGanWeight(hhs.GetType()) * 3.0
				}
			}
		}
		if selfForce > 0 && restrainingForce/selfForce > 0.6 {
			totalScore *= 0.8
			verdict = "偏旺" // 降级：得令但被严重克制
		}
	} else if lingScore <= 0.5 { // 失令
		spElem := supportingMap[dayElem]
		supportingForce := 0.0
		for _, fn := range pillars {
			sc := fn()
			stElem := sc.GetHeavenStem().GetElement().GetName()
			if stElem == spElem || stElem == dayElem {
				supportingForce += 5.0
			}
			for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
				hElem := hhs.GetHeavenStem().GetElement().GetName()
				if hElem == spElem || hElem == dayElem {
					supportingForce += zangGanWeight(hhs.GetType()) * 3.0
				}
			}
		}
		if supportingForce >= 8.0 { // 有足够生扶
			totalScore = totalScore*0.8 + 0.5*0.2 // 适度提升，不完全逆转
			if verdict == "身弱" {
				verdict = "偏弱"
			} else if verdict == "偏弱" {
				verdict = "中和"
			}
		}
	}

	// 根据日主五行动态计算喜忌
	// 身旺: 喜克泄耗(克我+我生+我克), 忌生扶(生我+同我)
	// 身弱: 喜生扶(生我+同我), 忌克泄耗(克我+我生+我克)
	elemRelation := map[string]map[string]string{
		"木": {"同我": "木", "生我": "水", "我生": "火", "克我": "金", "我克": "土"},
		"火": {"同我": "火", "生我": "木", "我生": "土", "克我": "水", "我克": "金"},
		"土": {"同我": "土", "生我": "火", "我生": "金", "克我": "木", "我克": "水"},
		"金": {"同我": "金", "生我": "土", "我生": "水", "克我": "火", "我克": "木"},
		"水": {"同我": "水", "生我": "金", "我生": "木", "克我": "土", "我克": "火"},
	}
	rel := elemRelation[dayElem]
	sameElem := rel["同我"]
	supportElem := rel["生我"]
	drainElem := rel["我生"]
	controlElem := rel["克我"]
	wealthElem := rel["我克"]
	if verdict == "身旺" || verdict == "偏旺" {
		like = []string{controlElem, drainElem, wealthElem}
		dislike = []string{supportElem, sameElem}
	} else if verdict == "身弱" || verdict == "偏弱" {
		like = []string{supportElem, sameElem}
		dislike = []string{controlElem, drainElem, wealthElem}
	} else {
		// 中和：五行相对平衡，喜通关调候
		// 经典依据：穷通宝鉴调候用神，在中和格局中优先考虑
		tiaoHouRules := data.GetTiaohou(dayStem.GetName(), monthBranch.GetName())
		tiaoHouElem := ""
		if len(tiaoHouRules) > 0 {
			tiaoHouElem = data.GanElement[tiaoHouRules[0].XiShen]
		}
		like = []string{drainElem, wealthElem}
		if tiaoHouElem != "" && tiaoHouElem != drainElem && tiaoHouElem != wealthElem {
			like = append(like, tiaoHouElem)
		}
		dislike = []string{controlElem}
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
		nayinName := data.Nayin[data.GanIndex(p.fn().GetHeavenStem().GetName())][data.ZhiIndex(p.fn().GetEarthBranch().GetName())]
		entry := data.NaYinMap[nayinName]
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

	// 三合: check all 4 branches for 3-branch sanhe groups
	relations = append(relations, detectTripleCombinations(pairs)...)

	// 三会: check all 4 branches for 3-branch groups
	relations = append(relations, detectTripleMeetings(pairs)...)

	return relations
}

// tortureType checks if two branches form a 三刑 relation.
func tortureType(a, b tyme.EarthBranch) string {
	aName := a.GetName()
	bName := b.GetName()

	// 无礼之刑: 子-卯
	if (aName == "子" && bName == "卯") || (aName == "卯" && bName == "子") {
		return "无礼之刑"
	}
	// 恃势之刑: 丑-戌, 戌-未, 未-丑 (三命通会第38章)
	shiShi := [][]string{{"丑", "戌"}, {"戌", "未"}, {"未", "丑"}}
	for _, pair := range shiShi {
		if (aName == pair[0] && bName == pair[1]) || (aName == pair[1] && bName == pair[0]) {
			return "恃势之刑"
		}
	}
	// 无恩之刑: 寅-巳, 巳-申, 申-寅 (三命通会第38章)
	wuEn := [][]string{{"寅", "巳"}, {"巳", "申"}, {"申", "寅"}}
	for _, pair := range wuEn {
		if (aName == pair[0] && bName == pair[1]) || (aName == pair[1] && bName == pair[0]) {
			return "无恩之刑"
		}
	}
	// 自刑: 辰-辰, 午-午, 酉-酉, 亥-亥
	selfTorture := map[string]bool{"辰": true, "午": true, "酉": true, "亥": true}
	if aName == bName && selfTorture[aName] {
		return "自刑"
	}
	return ""
}

// detectTripleCombinations detects 三合 (three-branch combination for element generation).
func detectTripleCombinations(pairs []pillarPair) []ClashRelation {
	// 三合局: 申子辰(水), 亥卯未(木), 寅午戌(火), 巳酉丑(金)
	tripleGroups := [][]string{
		{"申", "子", "辰"}, // 水局
		{"亥", "卯", "未"}, // 木局
		{"寅", "午", "戌"}, // 火局
		{"巳", "酉", "丑"}, // 金局
	}

	var relations []ClashRelation
	branchIndex := make(map[string]int)
	for i, p := range pairs {
		n := p.branch.GetName()
		branchIndex[n] = i
	}

	seen := make(map[string]bool)

	for _, group := range tripleGroups {
		var matched []int
		for _, b := range group {
			if idx, ok := branchIndex[b]; ok {
				matched = append(matched, idx)
			}
		}
		if len(matched) >= 3 {
			for i := 0; i < len(matched); i++ {
				for j := i + 1; j < len(matched); j++ {
					pi, pj := pairs[matched[i]], pairs[matched[j]]
					key := pi.name + "<>" + pj.name + "<>三合"
					if key2 := pj.name + "<>" + pi.name + "<>三合"; seen[key2] {
						continue
					}
					if !seen[key] {
						seen[key] = true
						relations = append(relations, ClashRelation{pi.name, pj.name, "三合"})
					}
				}
			}
		}
	}
	return relations
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

// GanInfoOf returns element and yang flag for a stem name.
func GanInfoOf(name string) ganInfo {
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
	return GanInfoOf(dayGan).elem
}

// ClassifyTenGod classifies a stem (stemName) against the day stem (dayGan),
// returning the ten-god name string.
// Only the VISIBLE day pillar stem itself is treated as 日主 (self) and excluded.
// Hidden stems that happen to share the same name as the day pillar stem
// are classified normally (比肩/劫财 when same element).
func ClassifyTenGod(stemName string, dayGan string, isDayPillarStem bool) string {
	stem := GanInfoOf(stemName)
	day := GanInfoOf(dayGan)

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
		god := ClassifyTenGod(stemName, dayGan, fn.day)
		if god != "" && god != "日主" {
			counts[god]++
		}
	}

	// Hidden stems from 4 branches — all are from branch (not visible day stem) so isDayPillarStem=false
	for _, fn := range [](func() tyme.SixtyCycle){ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour} {
		sc := fn()
		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			stemName := hhs.GetHeavenStem().GetName()
			god := ClassifyTenGod(stemName, dayGan, false)
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

// calcPillarTenGods computes which ten gods appear in each pillar position.
func calcPillarTenGods(ec *tyme.EightChar, dayGan string) PillarTenGods {
	classify := func(stemName string) string {
		god := ClassifyTenGod(stemName, dayGan, false)
		if god == "日主" {
			return ""
		}
		return god
	}

	ptg := PillarTenGods{}

	// Visible stems (year, month, hour)
	if g := classify(ec.GetYear().GetHeavenStem().GetName()); g != "" {
		ptg.Year = append(ptg.Year, g)
	}
	if g := classify(ec.GetMonth().GetHeavenStem().GetName()); g != "" {
		ptg.Month = append(ptg.Month, g)
	}
	if g := classify(ec.GetHour().GetHeavenStem().GetName()); g != "" {
		ptg.Hour = append(ptg.Hour, g)
	}

	// Hidden stems from all 4 branches
	type pillarTarget struct {
		fn     func() tyme.SixtyCycle
		target *[]string
	}
	for _, pair := range []pillarTarget{
		{ec.GetYear, &ptg.Year},
		{ec.GetMonth, &ptg.Month},
		{ec.GetDay, &ptg.Day},
		{ec.GetHour, &ptg.Hour},
	} {
		for _, hhs := range pair.fn().GetEarthBranch().GetHideHeavenStems() {
			if g := classify(hhs.GetHeavenStem().GetName()); g != "" {
				*pair.target = append(*pair.target, g)
			}
		}
	}

	return ptg
}

// enrichRiZhuText splits the data.SiZiSummaries text into poem/source/comment/hourDetail.
func enrichRiZhuText(result *BaziResult) {
	key := result.DayPillar.Gan + "日" + result.DayPillar.Zhi
	text := data.SiZiSummaries[key]

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

// --- 《三命通会》knowledge enrichment functions ---

// enrichWuxingSeason adds seasonal five-element analysis from 《三命通会》chapters 65-69.
func enrichWuxingSeason(result *BaziResult, birthMonth int) {
	dayElem := data.GanElement[result.DayPillar.Gan]
	season := data.SeasonFromMonth(birthMonth)
	entries, ok := data.WuxingSeasonKnowledge[dayElem]
	if !ok {
		return
	}
	for _, e := range entries {
		if e.Season == season {
			result.WuxingSeasonNote = fmt.Sprintf("【%s%s论】%s：%s 喜：%s 忌：%s",
				dayElem, season, e.State, e.Judgment, e.Favor, e.Taboo)
			return
		}
	}
}

// enrichJiaZiDetail adds 60-Jiazi properties from 《三命通会》chapter 7.
func enrichJiaZiDetail(result *BaziResult) {
	dayGZ := result.DayPillar.Gan + result.DayPillar.Zhi
	if entry, ok := data.JiaZiKnowledge[dayGZ]; ok {
		result.JiaZiDetail = &entry
	}
}

// enrichHealthNote adds health analysis based on five-element imbalance.
func enrichHealthNote(result *BaziResult) {
	// Find the most excessive element
	var maxElem string
	maxScore := 0
	for elem, score := range result.FiveElements {
		if score > maxScore {
			maxScore = score
			maxElem = elem
		}
	}
	if health, ok := data.WuxingHealthMap[maxElem]; ok && maxScore > 10 {
		result.HealthNote = fmt.Sprintf("【%s过旺】注意%s：%s", maxElem,
			strings.Join(health.Organs, "、"), health.Excess)
	}

	// Also check the weakest element
	var minElem string
	minScore := int(^uint(0) >> 1) // max int
	for elem, score := range result.FiveElements {
		if score < minScore {
			minScore = score
			minElem = elem
		}
	}
	if health, ok := data.WuxingHealthMap[minElem]; ok && minScore < 3 {
		if result.HealthNote != "" {
			result.HealthNote += "；"
		}
		result.HealthNote += fmt.Sprintf("【%s过弱】注意%s：%s", minElem,
			strings.Join(health.Organs, "、"), health.Deficit)
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
