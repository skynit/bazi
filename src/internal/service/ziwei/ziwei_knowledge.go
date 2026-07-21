package ziwei

// ──────────────────── ZiWei Knowledge Enrichment Layer ────────────────────
//
// All constants and algorithms are derived from classical ZiWei Dou Shu texts
// and the reference implementation at https://github.com/Renhuai123/ziwei-doushu.
// No AI/LLM generation. No external runtime dependencies.
//
// This layer enriches the output of ziwei-zenith with local computation:
//   - 四化飞星 chain analysis
//   - 40+ local pattern detectors
//   - 三方四正 computation
//   - 合盘 (heming) analysis
//   - Template-based palace readings
// ─────────────────────────────────────────────────────────────────────────

import (
	"fmt"
	"strings"
)

// ──────────────────── Constants ────────────────────

var brightnessLevels = []string{"庙", "旺", "得", "利", "平", "陷", "不"}

// STAR_BRIGHTNESS records star and brightness structure without converting it
// into personality, career, financial, or life-outcome claims.
var STAR_BRIGHTNESS = buildStructuralBrightnessDescriptions([]string{
	"紫微", "天府", "天机", "太阳", "武曲", "天同", "廉贞",
	"贪狼", "巨门", "天相", "天梁", "七杀", "破军", "太阴",
})

// STAR_BRIGHTNESS_AUX follows the same boundary for auxiliary stars.
var STAR_BRIGHTNESS_AUX = buildStructuralBrightnessDescriptions([]string{
	"文昌", "文曲", "擎羊", "陀罗", "火星", "铃星",
})

func buildStructuralBrightnessDescriptions(stars []string) map[string]map[string]string {
	result := make(map[string]map[string]string, len(stars))
	for _, star := range stars {
		levels := make(map[string]string, len(brightnessLevels))
		for _, level := range brightnessLevels {
			levels[level] = fmt.Sprintf("%s亮度为%s；仅记录传统星曜强弱结构，具体个体结果未裁决。", star, level)
		}
		result[star] = levels
	}
	return result
}

// LUCUN_TABLE maps year stem index to the branch where 禄存 is located.
// Kept as a compatibility alias for legacy helpers and tests.
var LUCUN_TABLE = map[int]int{
	0: LucunBranchIdx[0],
	1: LucunBranchIdx[1],
	2: LucunBranchIdx[2],
	3: LucunBranchIdx[3],
	4: LucunBranchIdx[4],
	5: LucunBranchIdx[5],
	6: LucunBranchIdx[6],
	7: LucunBranchIdx[7],
	8: LucunBranchIdx[8],
	9: LucunBranchIdx[9],
}

// TIANMA_TABLE maps the three-combination group to the branch where 天马 is.
// 经典口诀：申子辰马在寅，寅午戌马在申，巳酉丑马在亥，亥卯未马在巳
var TIANMA_TABLE = map[string]int{
	"寅午戌": 8,  // 马在申(8)
	"申子辰": 2,  // 马在寅(2)
	"巳酉丑": 11, // 马在亥(11)
	"亥卯未": 5,  // 马在巳(5)
}

// ──────────────────── Palace order (clockwise from 命宫) ────────────────────
var PALACE_NAMES = []string{
	"命宫", "兄弟", "夫妻", "子女",
	"财帛", "疾厄", "迁移", "交友",
	"事业", "田宅", "福德", "父母",
}

// ──────────────────── Palace-stem Sihua Flight Projection ────────────────────

const (
	sihuaProjectionBasis          = "deterministic_rule_projection"
	sihuaProjectionValidation     = "cross_checked_not_gold"
	sihuaProjectionSourceTier     = "silver_external"
	sihuaDirectFlightAnalysisKind = "direct_palace_stem_four_hua_flights"
)

type SihuaProjectionSemantics struct {
	RuleID              string `json:"rule_id"`
	SourceTier          string `json:"source_tier"`
	PlacementBasis      string `json:"placement_basis"`
	ValidationStatus    string `json:"validation_status"`
	IsOutcomeConclusion bool   `json:"is_outcome_conclusion"`
}

// SihuaChainResult retains the historical API type name, but contains only
// direct palace-stem transformation edges. It does not claim recursive chains,
// aggregate depth, affinity, or real-world effects.
type SihuaChainResult struct {
	SihuaProjectionSemantics
	AnalysisKind string           `json:"analysis_kind"`
	HuaLu        []SihuaChainItem `json:"hua_lu"`
	HuaQuan      []SihuaChainItem `json:"hua_quan"`
	HuaKe        []SihuaChainItem `json:"hua_ke"`
	HuaJi        []SihuaChainItem `json:"hua_ji"`
}

// SelfMutagenResult holds a self-mutagen occurrence.
type SelfMutagenResult struct {
	SihuaProjectionSemantics
	Palace          string `json:"palace"`
	PalaceStem      string `json:"palace_stem"`
	TransformedStar string `json:"transformed_star"`
	HuaType         string `json:"hua_type"`
	StructureStatus string `json:"structure_status"`
	IsSelfMutagen   bool   `json:"is_self_mutagen"`
}

// SihuaChainItem represents one direct palace-stem transformation edge.
type SihuaChainItem struct {
	SihuaProjectionSemantics
	SourcePalace     string `json:"source_palace"`
	SourcePalaceStem string `json:"source_palace_stem"`
	TransformedStar  string `json:"transformed_star"`
	HuaType          string `json:"hua_type"`
	TargetPalace     string `json:"target_palace"`
	FlightScope      string `json:"flight_scope"` // "same_palace" | "cross_palace"
	IsSelfMutagen    bool   `json:"is_self_mutagen"`
}

// analyzeSihuaChain projects every palace stem's four transformations onto the
// actual star positions in the chart. The palace whose heavenly stem produces
// the transformation is the source; the palace containing the transformed star
// is the target. A transformation is self-mutating only when both are the same
// palace. PalaceInfo.FourHua is deliberately not used here because it records
// natal year-stem transformations rather than palace-stem flights.
func analyzeSihuaChain(chart *ZiWeiChart) *SihuaChainResult {
	if chart == nil {
		return nil
	}

	starToPalaceIdx := buildStarPalaceIndex(chart)

	result := &SihuaChainResult{
		SihuaProjectionSemantics: sihuaProjectionSemantics(),
		AnalysisKind:             sihuaDirectFlightAnalysisKind,
	}
	for fromPalaceIdx, palace := range chart.Palaces {
		stemIdx, ok := StemIndex[strings.TrimSpace(palace.HeavenlyStem)]
		if !ok {
			continue
		}

		for huaIdx, starName := range SiHuaTable[stemIdx] {
			toPalaceIdx, found := starToPalaceIdx[starName]
			if !found {
				continue
			}

			huaType := SiHuaLabels[huaIdx]
			fromPalace := chartPalaceName(chart, fromPalaceIdx)
			toPalace := chartPalaceName(chart, toPalaceIdx)
			flightScope, isSelf := classifySihuaFlight(fromPalaceIdx, toPalaceIdx)

			item := SihuaChainItem{
				SihuaProjectionSemantics: sihuaProjectionSemantics(),
				SourcePalace:             fromPalace,
				SourcePalaceStem:         strings.TrimSpace(palace.HeavenlyStem),
				TransformedStar:          starName,
				HuaType:                  huaType,
				TargetPalace:             toPalace,
				FlightScope:              flightScope,
				IsSelfMutagen:            isSelf,
			}
			appendSihuaChainItem(result, huaIdx, item)
		}
	}

	return result
}

func appendSihuaChainItem(result *SihuaChainResult, huaIdx int, item SihuaChainItem) {
	switch huaIdx {
	case 0:
		result.HuaLu = append(result.HuaLu, item)
	case 1:
		result.HuaQuan = append(result.HuaQuan, item)
	case 2:
		result.HuaKe = append(result.HuaKe, item)
	case 3:
		result.HuaJi = append(result.HuaJi, item)
	}
}

func buildStarPalaceIndex(chart *ZiWeiChart) map[string]int {
	index := make(map[string]int)
	for palaceIdx, palace := range chart.Palaces {
		add := func(starName string) {
			starName = strings.TrimSpace(starName)
			if starName == "" {
				return
			}
			if _, exists := index[starName]; !exists {
				index[starName] = palaceIdx
			}
		}
		for _, starName := range palaceMainStars(palace) {
			add(starName)
		}
		for _, starName := range palaceAuxStars(palace) {
			add(starName)
		}
	}
	return index
}

func chartPalaceName(chart *ZiWeiChart, palaceIdx int) string {
	if chart != nil && palaceIdx >= 0 && palaceIdx < len(chart.Palaces) {
		if name := strings.TrimSpace(chart.Palaces[palaceIdx].Name); name != "" {
			return name
		}
	}
	if palaceIdx >= 0 && palaceIdx < len(PALACE_NAMES) {
		return PALACE_NAMES[palaceIdx]
	}
	return ""
}

func sihuaProjectionSemantics() SihuaProjectionSemantics {
	return SihuaProjectionSemantics{
		RuleID:              SiHuaRuleID,
		SourceTier:          sihuaProjectionSourceTier,
		PlacementBasis:      sihuaProjectionBasis,
		ValidationStatus:    sihuaProjectionValidation,
		IsOutcomeConclusion: false,
	}
}

// classifySihuaFlight checks whether a palace-stem transformation lands back
// in its source palace or crosses into a different palace.
func classifySihuaFlight(fromPalaceIdx, toPalaceIdx int) (flightScope string, isSelf bool) {
	if fromPalaceIdx == toPalaceIdx {
		return "same_palace", true
	}
	return "cross_palace", false
}

// detectSelfMutagens returns only palace-stem transformations whose target star
// is located in the source palace. Natal year-stem FourHua labels do not imply
// self-mutation and are intentionally ignored.
func detectSelfMutagens(chart *ZiWeiChart) []SelfMutagenResult {
	if chart == nil {
		return nil
	}
	chain := analyzeSihuaChain(chart)
	if chain == nil {
		return nil
	}

	var results []SelfMutagenResult
	groups := []struct {
		huaType string
		items   []SihuaChainItem
	}{
		{"化禄", chain.HuaLu},
		{"化权", chain.HuaQuan},
		{"化科", chain.HuaKe},
		{"化忌", chain.HuaJi},
	}
	for _, group := range groups {
		for _, item := range group.items {
			if !item.IsSelfMutagen {
				continue
			}
			results = append(results, SelfMutagenResult{
				SihuaProjectionSemantics: sihuaProjectionSemantics(),
				Palace:                   item.SourcePalace,
				PalaceStem:               item.SourcePalaceStem,
				TransformedStar:          item.TransformedStar,
				HuaType:                  group.huaType,
				StructureStatus:          "same_palace_transformation",
				IsSelfMutagen:            true,
			})
		}
	}
	return results
}

// ──────────────────── Pattern Detection ────────────────────

// patternChecker is a function that checks for a specific pattern.
type patternChecker func(chart *ZiWeiChart) (bool, string)

// DetectLocalPatterns detects fortune patterns using local rules (not engine).
func DetectLocalPatterns(chart *ZiWeiChart) []string {
	if chart == nil {
		return nil
	}

	var detected []string

	for _, pc := range patternCheckers {
		if present, _ := pc.checker(chart); present {
			detected = append(detected, pc.name)
		}
	}

	return detected
}

// patternCheckers is a list of pattern checker functions with their names.
var patternCheckers = []struct {
	name    string
	checker patternChecker
}{
	{"紫府同宫", checkZiFuTongGong},
	{"杀破狼格", checkShaPoLang},
	{"机月同梁格", checkJiYueTongLiang},
	{"紫武廉府", checkZiWuLianFu},
	{"府相朝垣", checkFuXiangChaoYuan},
	{"日月拱照", checkRiYueGongZhao},
	{"日月反背", checkRiYueFanBei},
	{"禄马交驰", checkLuMaJiaoChi},
	{"火贪格", checkHuoTanGe},
	{"铃贪格", checkLingTanGe},
	{"空宫", checkKongGong},
	{"日月并明", checkRiYueBingMing},
	{"极向离明", checkJiXiangLiMing},
	{"石中隐玉", checkShiZhongYinYu},
	{"昌曲同会", checkChangQuTongHui},
	{"马头带箭", checkMaTouDaiJian},
	{"巨日同宫", checkJuRiTongGong},
	{"禄马佩印", checkLuMaPeiYin},
	{"三奇加会", checkSanQiJiaHui},
	{"七杀朝斗", checkQiShaChaoDou},
	{"武贪格", checkWuTanGe},
	{"天同天梁格", checkTianTongTianLiang},
	{"日月夹命", checkRiYueJiaMing},
	{"辅弼夹命", checkFuBiJiaMing},
	{"科权双会", checkKeQuanShuangHui},
	{"权禄生逢", checkQuanLuShengFeng},
	{"魁钺夹命", checkKuiYueJiaMing},
	{"羊陀夹忌", checkYangTuoJiaJi},
	{"紫府夹命", checkZiFuJiaMing},
	{"日月夹财", checkRiYueJiaCai},
	{"火铃夹命", checkHuoLingJiaMing},
	{"空劫夹命", checkKongJieJiaMing},
	{"月朗天门", checkYueLangTianMen},
	{"日照雷门", checkRiZhaoLeiMen},
}

func starInPalace(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return false
	}
	for _, main := range palaceMainStars(chart.Palaces[palaceIdx]) {
		for _, s := range starNames {
			if main == s {
				return true
			}
		}
	}
	return false
}

func starInPalaceByName(chart *ZiWeiChart, palaceName string, starNames []string) bool {
	if chart == nil {
		return false
	}
	for i, p := range chart.Palaces {
		if p.Name == palaceName {
			return starInPalace(chart, i, starNames)
		}
	}
	return false
}

func auxStarInPalace(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return false
	}
	for _, aux := range palaceAuxStars(chart.Palaces[palaceIdx]) {
		for _, s := range starNames {
			if aux == s {
				return true
			}
		}
	}
	return false
}

func anyStarInPalace(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	return starInPalace(chart, palaceIdx, starNames) || auxStarInPalace(chart, palaceIdx, starNames)
}

func starPairInPalace(chart *ZiWeiChart, palaceIdx int, star1, star2 string) bool {
	return anyStarInPalace(chart, palaceIdx, []string{star1}) && anyStarInPalace(chart, palaceIdx, []string{star2})
}

func starInSamePalace(chart *ZiWeiChart, palaceIdx int, star1, star2 string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return false
	}
	p := chart.Palaces[palaceIdx]
	has1 := false
	has2 := false
	for _, s := range palaceMainStars(p) {
		if s == star1 {
			has1 = true
		}
		if s == star2 {
			has2 = true
		}
	}
	return has1 && has2
}

func hasBrightness(chart *ZiWeiChart, palaceIdx int, starName string, brightness []string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return false
	}
	starBright := palaceStarBrightness(chart.Palaces[palaceIdx], starName)
	if starBright == "" {
		return false
	}
	for _, br := range brightness {
		if starBright == br {
			return true
		}
	}
	return false
}

func trineIndexes(chart *ZiWeiChart, palaceIdx int) (int, int) {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return (palaceIdx + 4) % 12, (palaceIdx + 8) % 12
	}
	_, trine1, trine2 := chartSanfangIndexes(chart, palaceIdx)
	return trine1, trine2
}

func sanfangIndexesWithSelf(chart *ZiWeiChart, palaceIdx int) []int {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		sf := computeSanfangSizheng(palaceIdx)
		return []int{palaceIdx, sf[0], sf[1], sf[2]}
	}
	oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, palaceIdx)
	return []int{palaceIdx, oppositeIdx, trine1Idx, trine2Idx}
}

func countMainStarsInIndexes(chart *ZiWeiChart, indexes []int, starNames []string) int {
	if chart == nil {
		return 0
	}
	seen := map[string]bool{}
	for _, idx := range indexes {
		if idx < 0 || idx >= 12 {
			continue
		}
		for _, main := range palaceMainStars(chart.Palaces[idx]) {
			for _, wanted := range starNames {
				if main == wanted {
					seen[wanted] = true
				}
			}
		}
	}
	return len(seen)
}

func countAnyStarsInIndexes(chart *ZiWeiChart, indexes []int, starNames []string) int {
	if chart == nil {
		return 0
	}
	seen := map[string]bool{}
	for _, idx := range indexes {
		if idx < 0 || idx >= 12 {
			continue
		}
		for _, wanted := range starNames {
			if anyStarInPalace(chart, idx, []string{wanted}) {
				seen[wanted] = true
			}
		}
	}
	return len(seen)
}

func hasTrine(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil {
		return false
	}
	trine1, trine2 := trineIndexes(chart, palaceIdx)
	return starInPalace(chart, trine1, starNames) || starInPalace(chart, trine2, starNames)
}

func hasOpposition(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil {
		return false
	}
	opposite, _, _ := chartSanfangIndexes(chart, palaceIdx)
	return starInPalace(chart, opposite, starNames)
}

func hasAnyStarTrine(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil {
		return false
	}
	trine1, trine2 := trineIndexes(chart, palaceIdx)
	return anyStarInPalace(chart, trine1, starNames) || anyStarInPalace(chart, trine2, starNames)
}

func hasAnyStarOpposition(chart *ZiWeiChart, palaceIdx int, starNames []string) bool {
	if chart == nil {
		return false
	}
	opposite, _, _ := chartSanfangIndexes(chart, palaceIdx)
	return anyStarInPalace(chart, opposite, starNames)
}

// Pattern checkers

func checkZiFuTongGong(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 || (chart.Palaces[mingIdx].Branch != "寅" && chart.Palaces[mingIdx].Branch != "申") {
		return false, ""
	}
	if starInSamePalace(chart, mingIdx, "紫微", "天府") {
		return true, "紫府同宫"
	}
	return false, ""
}

func checkShaPoLang(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	targets := []string{"七杀", "破军", "贪狼"}
	if countMainStarsInIndexes(chart, sanfangIndexesWithSelf(chart, mingIdx), targets) == len(targets) {
		return true, "杀破狼格"
	}
	return false, ""
}

func checkJiYueTongLiang(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 || (chart.Palaces[mingIdx].Branch != "寅" && chart.Palaces[mingIdx].Branch != "申") {
		return false, ""
	}
	targets := []string{"天机", "太阴", "天同", "天梁"}
	if countMainStarsInIndexes(chart, sanfangIndexesWithSelf(chart, mingIdx), targets) == len(targets) {
		return true, "机月同梁格"
	}
	return false, ""
}

func checkZiWuLianFu(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		stars := palaceMainStars(chart.Palaces[i])
		count := 0
		for _, s := range stars {
			if s == "紫微" || s == "武曲" || s == "廉贞" || s == "天府" {
				count++
			}
		}
		if count >= 3 {
			return true, "紫武廉府"
		}
	}
	return false, ""
}

func checkFuXiangChaoYuan(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	if starInPalaceByName(chart, "事业", []string{"天府"}) &&
		starInPalaceByName(chart, "财帛", []string{"天相"}) {
		return true, "府相朝垣"
	}
	return false, ""
}

func checkRiYueGongZhao(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	sunIdx := findMainStarPalaceIndex(chart, "太阳")
	moonIdx := findMainStarPalaceIndex(chart, "太阴")
	if sunIdx < 0 || moonIdx < 0 {
		return false, ""
	}

	lifeBranch := chart.Palaces[mingIdx].Branch
	sunBranch := chart.Palaces[sunIdx].Branch
	moonBranch := chart.Palaces[moonIdx].Branch
	if (lifeBranch == "丑" && sunBranch == "巳" && moonBranch == "酉") ||
		(lifeBranch == "未" && sunBranch == "卯" && moonBranch == "亥") ||
		(lifeBranch == "丑" && sunBranch == "未" && moonBranch == "未") ||
		((lifeBranch == "辰" || lifeBranch == "戌") && sunBranch == "辰" && moonBranch == "戌") {
		return true, "日月拱照"
	}
	return false, ""
}

func checkRiYueFanBei(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 日月反背：太阳、太阴均落陷，且同在命宫三方四正。
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	hasDimSun, hasDimMoon := false, false
	for _, idx := range sanfangIndexesWithSelf(chart, mingIdx) {
		hasDimSun = hasDimSun || (starInPalace(chart, idx, []string{"太阳"}) && hasBrightness(chart, idx, "太阳", []string{"陷", "不"}))
		hasDimMoon = hasDimMoon || (starInPalace(chart, idx, []string{"太阴"}) && hasBrightness(chart, idx, "太阴", []string{"陷", "不"}))
	}
	if hasDimSun && hasDimMoon {
		return true, "日月反背"
	}
	return false, ""
}

// checkZiFuJiaMing 紫府夹命：寅申命宫坐天机太阴，紫微、天府分居两侧。
func checkZiFuJiaMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 || (chart.Palaces[mingIdx].Branch != "寅" && chart.Palaces[mingIdx].Branch != "申") ||
		!starInSamePalace(chart, mingIdx, "天机", "太阴") {
		return false, ""
	}
	leftIdx := (mingIdx + 11) % 12
	rightIdx := (mingIdx + 1) % 12
	hasZiweiLeft := starInPalace(chart, leftIdx, []string{"紫微"})
	hasZiweiRight := starInPalace(chart, rightIdx, []string{"紫微"})
	hasTianfuLeft := starInPalace(chart, leftIdx, []string{"天府"})
	hasTianfuRight := starInPalace(chart, rightIdx, []string{"天府"})
	if (hasZiweiLeft && hasTianfuRight) || (hasTianfuLeft && hasZiweiRight) {
		return true, "紫府夹命"
	}
	return false, ""
}

// checkRiYueJiaCai 日月夹财：武曲守命受日月夹，或财帛宫受日月夹。
func checkRiYueJiaCai(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	wealthIdx := findPalaceIndex(chart, "财帛")
	if mingIdx >= 0 && starInPalace(chart, mingIdx, []string{"武曲"}) && sunMoonClampPalace(chart, mingIdx) {
		return true, "日月夹财"
	}
	if wealthIdx >= 0 && sunMoonClampPalace(chart, wealthIdx) {
		return true, "日月夹财"
	}
	return false, ""
}

func sunMoonClampPalace(chart *ZiWeiChart, palaceIdx int) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= len(chart.Palaces) {
		return false
	}
	leftIdx := (palaceIdx + 11) % 12
	rightIdx := (palaceIdx + 1) % 12
	hasSunL := starInPalace(chart, leftIdx, []string{"太阳"})
	hasSunR := starInPalace(chart, rightIdx, []string{"太阳"})
	hasMoonL := starInPalace(chart, leftIdx, []string{"太阴"})
	hasMoonR := starInPalace(chart, rightIdx, []string{"太阴"})
	if (hasSunL && hasMoonR) || (hasMoonL && hasSunR) {
		return true
	}
	return false
}

// checkHuoLingJiaMing 火铃夹命：火星、铃星在命宫两侧
func checkHuoLingJiaMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	leftIdx := (mingIdx + 11) % 12
	rightIdx := (mingIdx + 1) % 12
	hasHuoL := auxStarInPalace(chart, leftIdx, []string{"火星"})
	hasHuoR := auxStarInPalace(chart, rightIdx, []string{"火星"})
	hasLingL := auxStarInPalace(chart, leftIdx, []string{"铃星"})
	hasLingR := auxStarInPalace(chart, rightIdx, []string{"铃星"})
	if (hasHuoL && hasLingR) || (hasLingL && hasHuoR) {
		return true, "火铃夹命"
	}
	return false, ""
}

// checkKongJieJiaMing 空劫夹命：地空、地劫在命宫两侧
func checkKongJieJiaMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	leftIdx := (mingIdx + 11) % 12
	rightIdx := (mingIdx + 1) % 12
	hasKongL := auxStarInPalace(chart, leftIdx, []string{"地空"})
	hasKongR := auxStarInPalace(chart, rightIdx, []string{"地空"})
	hasJieL := auxStarInPalace(chart, leftIdx, []string{"地劫"})
	hasJieR := auxStarInPalace(chart, rightIdx, []string{"地劫"})
	if (hasKongL && hasJieR) || (hasJieL && hasKongR) {
		return true, "空劫夹命"
	}
	return false, ""
}

// checkYueLangTianMen 月朗天门：命宫在亥，太阴入庙
func checkYueLangTianMen(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	if chart.Palaces[mingIdx].Branch != "亥" {
		return false, ""
	}
	if starInPalace(chart, mingIdx, []string{"太阴"}) {
		return true, "月朗天门"
	}
	return false, ""
}

// checkRiZhaoLeiMen 日照雷门：命宫在卯，太阳入庙
func checkRiZhaoLeiMen(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	if chart.Palaces[mingIdx].Branch != "卯" {
		return false, ""
	}
	if starInPalace(chart, mingIdx, []string{"太阳"}) {
		return true, "日照雷门"
	}
	return false, ""
}

func checkLuMaJiaoChi(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		hasLu := auxStarInPalace(chart, i, []string{"禄存"})
		hasMa := auxStarInPalace(chart, i, []string{"天马"})
		if hasLu && hasMa {
			return true, "禄马交驰"
		}
	}
	return false, ""
}

func checkHuoTanGe(chart *ZiWeiChart) (bool, string) {
	return checkTanLangAuxPattern(chart, "火星", "火贪格")
}

func checkLingTanGe(chart *ZiWeiChart) (bool, string) {
	return checkTanLangAuxPattern(chart, "铃星", "铃贪格")
}

func checkTanLangAuxPattern(chart *ZiWeiChart, auxName, patternName string) (bool, string) {
	if chart == nil {
		return false, ""
	}
	tanIdx, mingIdx := -1, findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"贪狼"}) {
			tanIdx = i
			break
		}
	}
	if tanIdx < 0 {
		return false, ""
	}
	tanInLifeSanfang := false
	for _, idx := range sanfangIndexesWithSelf(chart, mingIdx) {
		if idx == tanIdx {
			tanInLifeSanfang = true
			break
		}
	}
	if !tanInLifeSanfang {
		return false, ""
	}
	for _, idx := range sanfangIndexesWithSelf(chart, tanIdx) {
		if auxStarInPalace(chart, idx, []string{auxName}) &&
			tanLangAuxRelationAllowed(auxName, chart.Palaces[tanIdx].Branch, idx == tanIdx) {
			return true, patternName
		}
	}
	return false, ""
}

func tanLangAuxRelationAllowed(auxName, tanBranch string, samePalace bool) bool {
	switch auxName {
	case "火星":
		switch tanBranch {
		case "子", "午", "卯", "酉", "辰", "戌", "丑", "未":
			return true
		}
	case "铃星":
		return samePalace || tanBranch == "辰" || tanBranch == "戌" || tanBranch == "丑" || tanBranch == "未"
	}
	return false
}

func checkKongGong(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := 0; i < 12; i++ {
		p := chart.Palaces[i]
		if p.Name == "命宫" {
			if len(palaceMainStars(p)) == 0 {
				return true, "空宫"
			}
		}
	}
	return false, ""
}

func checkRiYueBingMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 日月并明：太阳、太阴均庙旺，且同在命宫三方四正。
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	hasBrightSun, hasBrightMoon := false, false
	for _, idx := range sanfangIndexesWithSelf(chart, mingIdx) {
		hasBrightSun = hasBrightSun || (starInPalace(chart, idx, []string{"太阳"}) && hasBrightness(chart, idx, "太阳", []string{"庙", "旺"}))
		hasBrightMoon = hasBrightMoon || (starInPalace(chart, idx, []string{"太阴"}) && hasBrightness(chart, idx, "太阴", []string{"庙", "旺"}))
	}
	if hasBrightSun && hasBrightMoon {
		return true, "日月并明"
	}
	return false, ""
}

func checkJiXiangLiMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 紫微在午宫坐命，且命宫三方四正不见六煞。
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	if chart.Palaces[mingIdx].Branch != "午" ||
		!starInPalace(chart, mingIdx, []string{"紫微"}) ||
		!hasBrightness(chart, mingIdx, "紫微", []string{"庙", "旺"}) {
		return false, ""
	}
	shaStars := toughStarSet()
	for _, idx := range sanfangIndexesWithSelf(chart, mingIdx) {
		if len(filterStars(palaceAuxStars(chart.Palaces[idx]), shaStars)) > 0 {
			return false, ""
		}
	}
	return true, "极向离明"
}

func checkShiZhongYinYu(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	ming := chart.Palaces[mingIdx]
	if (ming.Branch == "子" || ming.Branch == "午") && starInPalace(chart, mingIdx, []string{"巨门"}) {
		return true, "石中隐玉"
	}
	return false, ""
}

func checkChangQuTongHui(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	targets := []string{"文昌", "文曲"}
	if countAnyStarsInIndexes(chart, sanfangIndexesWithSelf(chart, mingIdx), targets) == len(targets) {
		return true, "昌曲同会"
	}
	return false, ""
}

func checkMaTouDaiJian(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx >= 0 && chart.Palaces[mingIdx].Branch == "午" &&
		auxStarInPalace(chart, mingIdx, []string{"擎羊"}) {
		return true, "马头带箭"
	}
	return false, ""
}

func checkJuRiTongGong(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 || (chart.Palaces[mingIdx].Branch != "寅" && chart.Palaces[mingIdx].Branch != "申") {
		return false, ""
	}
	if starInSamePalace(chart, mingIdx, "巨门", "太阳") {
		return true, "巨日同宫"
	}
	return false, ""
}

func checkLuMaPeiYin(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{"天相"}) &&
			auxStarInPalace(chart, i, []string{"禄存"}) &&
			auxStarInPalace(chart, i, []string{"天马"}) {
			return true, "禄马佩印"
		}
	}
	return false, ""
}

func checkSanQiJiaHui(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	// 三奇: 化禄、化权、化科同时出现在命宫三方四正
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	huaTypes := make(map[string]int)
	oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, mingIdx)
	for _, i := range []int{mingIdx, oppositeIdx, trine1Idx, trine2Idx} {
		for _, t := range chart.Palaces[i].FourHua {
			if strings.Contains(t, "化禄") {
				huaTypes["化禄"]++
			}
			if strings.Contains(t, "化权") {
				huaTypes["化权"]++
			}
			if strings.Contains(t, "化科") {
				huaTypes["化科"]++
			}
		}
	}
	if huaTypes["化禄"] > 0 && huaTypes["化权"] > 0 && huaTypes["化科"] > 0 {
		return true, "三奇加会"
	}
	return false, ""
}

func checkQiShaChaoDou(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 || !isQiShaChaoDouBranch(chart.Palaces[mingIdx].Branch) ||
		!starInPalace(chart, mingIdx, []string{"七杀"}) {
		return false, ""
	}
	oppositeIdx, _, _ := chartSanfangIndexes(chart, mingIdx)
	switch chart.Palaces[mingIdx].Branch {
	case "寅", "申":
		if starInSamePalace(chart, oppositeIdx, "紫微", "天府") {
			return true, "七杀朝斗"
		}
	case "子", "午":
		if starInSamePalace(chart, oppositeIdx, "武曲", "天府") {
			return true, "七杀朝斗"
		}
	}
	return false, ""
}

func isQiShaChaoDouBranch(branch string) bool {
	switch branch {
	case "子", "午", "寅", "申":
		return true
	default:
		return false
	}
}

func checkWuTanGe(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	matches := func(idx int) bool {
		if idx < 0 || idx >= len(chart.Palaces) {
			return false
		}
		branch := chart.Palaces[idx].Branch
		return (branch == "丑" || branch == "未") && starInSamePalace(chart, idx, "武曲", "贪狼")
	}
	if matches(mingIdx) {
		return true, "武贪格"
	}
	for idx := range chart.Palaces {
		if chart.Palaces[idx].IsBodyPalace && matches(idx) {
			return true, "武贪格"
		}
	}
	return false, ""
}

func checkTianTongTianLiang(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 || (chart.Palaces[mingIdx].Branch != "寅" && chart.Palaces[mingIdx].Branch != "申") {
		return false, ""
	}
	if starInSamePalace(chart, mingIdx, "天同", "天梁") {
		return true, "天同天梁格"
	}
	return false, ""
}

func checkRiYueJiaMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	leftIdx, rightIdx := (mingIdx+11)%12, (mingIdx+1)%12
	if (starInPalace(chart, leftIdx, []string{"太阳"}) && starInPalace(chart, rightIdx, []string{"太阴"})) ||
		(starInPalace(chart, leftIdx, []string{"太阴"}) && starInPalace(chart, rightIdx, []string{"太阳"})) {
		return true, "日月夹命"
	}
	return false, ""
}

func checkFuBiJiaMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	leftIdx, rightIdx := (mingIdx+11)%12, (mingIdx+1)%12
	if (auxStarInPalace(chart, leftIdx, []string{"左辅"}) && auxStarInPalace(chart, rightIdx, []string{"右弼"})) ||
		(auxStarInPalace(chart, leftIdx, []string{"右弼"}) && auxStarInPalace(chart, rightIdx, []string{"左辅"})) {
		return true, "辅弼夹命"
	}
	return false, ""
}

// ──────────────────── Sanfang Sizheng Computation ────────────────────

// computeSanfangSizheng computes the 三方四正 for a valid palace index.
// Returns [3]int: {oppositeIndex, trine1Index, trine2Index}
func computeSanfangSizheng(palaceIdx int) [3]int {
	opposite := (palaceIdx + 6) % 12
	trine1 := (palaceIdx + 4) % 12
	trine2 := (palaceIdx + 8) % 12
	return [3]int{opposite, trine1, trine2}
}

// SanfangSizhengResult holds the sanfang analysis for one palace.
type SanfangSizhengResult struct {
	Opposite string `json:"opposite"` // 对宫 palace name
	Trine1   string `json:"trine1"`   // 三合宫 1
	Trine2   string `json:"trine2"`   // 三合宫 2
}

// getPalaceSanfang returns the sanfang sizheng result for a valid palace index.
func getPalaceSanfang(palaceIdx int) *SanfangSizhengResult {
	if palaceIdx < 0 || palaceIdx >= len(PALACE_NAMES) {
		return nil
	}
	sf := computeSanfangSizheng(palaceIdx)
	return &SanfangSizhengResult{
		Opposite: PALACE_NAMES[sf[0]],
		Trine1:   PALACE_NAMES[sf[1]],
		Trine2:   PALACE_NAMES[sf[2]],
	}
}

// getChartPalaceSanfang returns 三方四正 using the chart's actual palace
// branch positions instead of assuming array order equals branch order.
func getChartPalaceSanfang(chart *ZiWeiChart, palaceIdx int) *SanfangSizhengResult {
	if chart == nil || palaceIdx < 0 || palaceIdx >= len(chart.Palaces) {
		return nil
	}

	oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, palaceIdx)
	return &SanfangSizhengResult{
		Opposite: chart.Palaces[oppositeIdx].Name,
		Trine1:   chart.Palaces[trine1Idx].Name,
		Trine2:   chart.Palaces[trine2Idx].Name,
	}
}

func chartSanfangIndexes(chart *ZiWeiChart, palaceIdx int) (oppositeIdx, trine1Idx, trine2Idx int) {
	if chart == nil || palaceIdx < 0 || palaceIdx >= len(chart.Palaces) {
		sf := computeSanfangSizheng(palaceIdx)
		return sf[0], sf[1], sf[2]
	}

	branch := BranchIndex[chart.Palaces[palaceIdx].Branch]
	oppositeBranch := BranchNames[fixIndex(branch+6)]
	trine1Branch := BranchNames[fixIndex(branch+8)]
	trine2Branch := BranchNames[fixIndex(branch+4)]

	findByBranch := func(branchName string) int {
		for i, p := range chart.Palaces {
			if p.Branch == branchName {
				return i
			}
		}
		return palaceIdx
	}

	return findByBranch(oppositeBranch), findByBranch(trine1Branch), findByBranch(trine2Branch)
}

// EnhancedSanfangResult holds detailed sanfang analysis including star energy from SiHua.
type EnhancedSanfangResult struct {
	Opposite      string `json:"opposite"`       // 对宫 palace name
	OppositeIdx   int    `json:"opposite_idx"`   // 对宫 palace index
	Trine1        string `json:"trine1"`         // 三合宫 1
	Trine1Idx     int    `json:"trine1_idx"`     // 三合宫 1 index
	Trine2        string `json:"trine2"`         // 三合宫 2
	Trine2Idx     int    `json:"trine2_idx"`     // 三合宫 2 index
	OppositeSihua string `json:"opposite_sihua"` // 四化能量冲宫描述
	TrineSihua    string `json:"trine_sihua"`    // 三合拱照四化描述
}

// getEnhancedSanfang returns detailed sanfang analysis with SiHua interaction descriptions.
func getEnhancedSanfang(chart *ZiWeiChart, palaceIdx int) *EnhancedSanfangResult {
	if chart == nil || palaceIdx < 0 || palaceIdx >= len(chart.Palaces) {
		return nil
	}
	oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, palaceIdx)
	result := &EnhancedSanfangResult{
		Opposite:    chart.Palaces[oppositeIdx].Name,
		OppositeIdx: oppositeIdx,
		Trine1:      chart.Palaces[trine1Idx].Name,
		Trine1Idx:   trine1Idx,
		Trine2:      chart.Palaces[trine2Idx].Name,
		Trine2Idx:   trine2Idx,
	}

	var oppSihua, trineSihua []string

	// Every natal four-hua placement in the opposite palace participates in the
	// opposite-palace projection. Requiring the transformed star to also be a
	// main star in the current palace makes the condition impossible for the
	// chart's unique major-star placements and silently drops the influence.
	oppPalace := chart.Palaces[oppositeIdx]
	curPalace := chart.Palaces[palaceIdx]

	for _, t := range oppPalace.FourHua {
		oppSihua = append(oppSihua, t+"照"+curPalace.Name)
	}

	// The same projection rule applies to both trine palaces.
	for _, triIdx := range []int{trine1Idx, trine2Idx} {
		triPalace := chart.Palaces[triIdx]
		for _, t := range triPalace.FourHua {
			trineSihua = append(trineSihua, t+"拱"+curPalace.Name)
		}
	}

	if len(oppSihua) > 0 {
		result.OppositeSihua = strings.Join(oppSihua, "、")
	}
	if len(trineSihua) > 0 {
		result.TrineSihua = strings.Join(trineSihua, "、")
	}

	return result
}

// ──────────────────── Star Brightness Description ────────────────────

// GetStarBrightness returns the brightness description for a star+brightness combination.
func GetStarBrightness(star, brightness string) (string, bool) {
	if starMap, ok := STAR_BRIGHTNESS[star]; ok {
		if desc, ok := starMap[brightness]; ok {
			return desc, true
		}
	}
	if auxMap, ok := STAR_BRIGHTNESS_AUX[star]; ok {
		if desc, ok := auxMap[brightness]; ok {
			return desc, true
		}
	}
	return "", false
}

// ──────────────────── Heming (合盘) Analysis ────────────────────

// HemingResult holds a deterministic structural comparison of two published
// charts. It deliberately contains no compatibility or outcome score.
type HemingResult struct {
	StemRelation        HemingStemRelation       `json:"stem_relation"`
	PalaceComparisons   []HemingPalaceComparison `json:"palace_comparisons"`
	EvidenceBasis       string                   `json:"evidence_basis"`
	ValidationStatus    string                   `json:"validation_status"`
	IsOutcomeConclusion bool                     `json:"is_outcome_conclusion"`
	Notes               string                   `json:"notes"`
}

// HemingStemRelation records the deterministic relation between the two
// published birth-year stems. It describes only a traditional rule structure;
// it never decides whether a five-combination transforms or predicts a
// relationship outcome.
type HemingStemRelation struct {
	StemA                string `json:"stem_a"`
	StemB                string `json:"stem_b"`
	ElementA             string `json:"element_a"`
	ElementB             string `json:"element_b"`
	RelationType         string `json:"relation_type"`
	RelationLabel        string `json:"relation_label"`
	Direction            string `json:"direction"`
	FiveCombineTarget    string `json:"five_combine_target,omitempty"`
	StructureStatus      string `json:"structure_status"`
	TransformationStatus string `json:"transformation_status"`
	RuleID               string `json:"rule_id"`
	EvidenceBasis        string `json:"evidence_basis"`
	ValidationStatus     string `json:"validation_status"`
	IsOutcomeConclusion  bool   `json:"is_outcome_conclusion"`
	Notes                string `json:"notes"`
}

// HemingPalaceComparison projects the complete public structure of the same
// named palace from both charts and records exact set intersections only.
type HemingPalaceComparison struct {
	Palace                string                 `json:"palace"`
	ChartA                HemingPalaceProjection `json:"chart_a"`
	ChartB                HemingPalaceProjection `json:"chart_b"`
	SharedStars           []string               `json:"shared_stars"`
	SharedFourHua         []string               `json:"shared_four_hua"`
	SharedAdjectiveStars  []string               `json:"shared_adjective_stars"`
	ComparisonBasis       string                 `json:"comparison_basis"`
	InterpretationStatus  string                 `json:"interpretation_status"`
	IsCompatibilityResult bool                   `json:"is_compatibility_result"`
}

// HemingPalaceProjection preserves every public PalaceInfo field except Name,
// which is represented once by HemingPalaceComparison.Palace.
type HemingPalaceProjection struct {
	Branch         string       `json:"branch"`
	HeavenlyStem   string       `json:"heavenly_stem"`
	IsBodyPalace   bool         `json:"is_body_palace"`
	Stars          []StarOutput `json:"stars"`
	FourHua        []string     `json:"four_hua"`
	AdjectiveStars []string     `json:"adjective_stars"`
	Changsheng12   string       `json:"changsheng_12"`
	Boshi12        string       `json:"boshi_12"`
	JiangQian12    string       `json:"jiang_qian_12"`
	SuiQian12      string       `json:"sui_qian_12"`
}

// analyzeHeming compares authenticated charts under one exact calculation
// profile. It returns structural facts only, not a compatibility score.
func analyzeHeming(chartA, chartB *ZiWeiChart) *HemingResult {
	if chartA == nil || chartB == nil {
		return nil
	}
	profile, err := ResolveProfile(chartA.ProfileID)
	if err != nil || !chartMatchesProfile(chartA, profile) || !chartMatchesProfile(chartB, profile) {
		return nil
	}
	birthA, ok := birthDataFromPublishedChart(chartA)
	if !ok {
		return nil
	}
	birthB, ok := birthDataFromPublishedChart(chartB)
	if !ok {
		return nil
	}

	result := &HemingResult{
		PalaceComparisons:   make([]HemingPalaceComparison, 0, 5),
		EvidenceBasis:       "deterministic_published_chart_projection",
		ValidationStatus:    "not_adjudicated",
		IsOutcomeConclusion: false,
		Notes:               "仅并列双方公开命盘结构；共同项不表示契合，差异项不表示冲突，不推断关系质量、婚姻结果或事件时点。",
	}

	for _, palaceName := range []string{"命宫", "夫妻", "福德", "事业", "财帛"} {
		comparison, ok := buildHemingPalaceComparison(chartA, chartB, palaceName)
		if !ok {
			return nil
		}
		result.PalaceComparisons = append(result.PalaceComparisons, comparison)
	}

	stemRelation, ok := buildHemingStemRelation(birthA.YearStem, birthB.YearStem)
	if !ok {
		return nil
	}
	result.StemRelation = stemRelation

	return result
}

func buildHemingPalaceComparison(chartA, chartB *ZiWeiChart, palaceName string) (HemingPalaceComparison, bool) {
	palaceA, ok := uniquePublishedPalace(chartA, palaceName)
	if !ok {
		return HemingPalaceComparison{}, false
	}
	palaceB, ok := uniquePublishedPalace(chartB, palaceName)
	if !ok {
		return HemingPalaceComparison{}, false
	}
	return HemingPalaceComparison{
		Palace:                palaceName,
		ChartA:                projectHemingPalace(palaceA),
		ChartB:                projectHemingPalace(palaceB),
		SharedStars:           orderedStringIntersection(starOutputNames(palaceA.Stars), starOutputNames(palaceB.Stars)),
		SharedFourHua:         orderedStringIntersection(palaceA.FourHua, palaceB.FourHua),
		SharedAdjectiveStars:  orderedStringIntersection(palaceA.AdjectiveStars, palaceB.AdjectiveStars),
		ComparisonBasis:       "same_named_palace_exact_public_projection",
		InterpretationStatus:  "not_adjudicated",
		IsCompatibilityResult: false,
	}, true
}

func uniquePublishedPalace(chart *ZiWeiChart, palaceName string) (PalaceInfo, bool) {
	if chart == nil {
		return PalaceInfo{}, false
	}
	var result PalaceInfo
	found := false
	for _, palace := range chart.Palaces {
		if palace.Name != palaceName {
			continue
		}
		if found {
			return PalaceInfo{}, false
		}
		result = palace
		found = true
	}
	return result, found
}

func projectHemingPalace(palace PalaceInfo) HemingPalaceProjection {
	return HemingPalaceProjection{
		Branch:         palace.Branch,
		HeavenlyStem:   palace.HeavenlyStem,
		IsBodyPalace:   palace.IsBodyPalace,
		Stars:          cloneStarOutputs(palace.Stars),
		FourHua:        cloneStringsPreserveNil(palace.FourHua),
		AdjectiveStars: cloneStringsPreserveNil(palace.AdjectiveStars),
		Changsheng12:   palace.Changsheng12,
		Boshi12:        palace.Boshi12,
		JiangQian12:    palace.JiangQian12,
		SuiQian12:      palace.SuiQian12,
	}
}

func cloneStarOutputs(stars []StarOutput) []StarOutput {
	if stars == nil {
		return nil
	}
	result := make([]StarOutput, len(stars))
	copy(result, stars)
	return result
}

func cloneStringsPreserveNil(values []string) []string {
	if values == nil {
		return nil
	}
	return append([]string{}, values...)
}

func starOutputNames(stars []StarOutput) []string {
	names := make([]string, 0, len(stars))
	for _, star := range stars {
		names = append(names, star.Name)
	}
	return names
}

func orderedStringIntersection(valuesA, valuesB []string) []string {
	presentB := make(map[string]bool, len(valuesB))
	for _, value := range valuesB {
		presentB[value] = true
	}
	result := make([]string, 0)
	seen := make(map[string]bool)
	for _, value := range valuesA {
		if presentB[value] && !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result
}

func buildHemingStemRelation(stemA, stemB int) (HemingStemRelation, bool) {
	if stemA < 0 || stemA >= len(StemNames) || stemB < 0 || stemB >= len(StemNames) {
		return HemingStemRelation{}, false
	}

	elements := [...]string{"木", "火", "土", "金", "水"}
	nameA, nameB := StemNames[stemA], StemNames[stemB]
	elementA, elementB := elements[stemWuXingIdx(stemA)], elements[stemWuXingIdx(stemB)]
	relation := HemingStemRelation{
		StemA:               nameA,
		StemB:               nameB,
		ElementA:            elementA,
		ElementB:            elementB,
		EvidenceBasis:       "deterministic_traditional_rule",
		ValidationStatus:    "not_adjudicated",
		IsOutcomeConclusion: false,
	}

	if target, pairID, ok := hemingFiveCombine(stemA, stemB); ok {
		relation.RelationType = "five_combine"
		relation.RelationLabel = "天干五合"
		relation.Direction = "mutual"
		relation.FiveCombineTarget = target
		relation.StructureStatus = "complete_structure"
		relation.TransformationStatus = "unadjudicated"
		relation.RuleID = "heming.year-stem.five-combine." + pairID
		relation.Notes = "双方生年天干构成五合结构；目标五行为" + target + "。仅确认配对，不裁决合化成功，也不据此推断关系质量或现实结果。"
		return relation, true
	}

	relation.StructureStatus = "observed_relation"
	relation.TransformationStatus = "not_applicable"
	switch {
	case elementA == elementB:
		relation.RelationType = "same_element"
		relation.RelationLabel = "同五行"
		relation.Direction = "mutual"
		relation.RuleID = "heming.year-stem.same-element." + elementA
	case (stemWuXingIdx(stemA)+1)%5 == stemWuXingIdx(stemB):
		relation.RelationType = "generates"
		relation.RelationLabel = "五行相生"
		relation.Direction = "a_to_b"
		relation.RuleID = "heming.year-stem.generate." + elementA + "-" + elementB
	case (stemWuXingIdx(stemB)+1)%5 == stemWuXingIdx(stemA):
		relation.RelationType = "generates"
		relation.RelationLabel = "五行相生"
		relation.Direction = "b_to_a"
		relation.RuleID = "heming.year-stem.generate." + elementB + "-" + elementA
	case (stemWuXingIdx(stemA)+2)%5 == stemWuXingIdx(stemB):
		relation.RelationType = "controls"
		relation.RelationLabel = "五行相克"
		relation.Direction = "a_to_b"
		relation.RuleID = "heming.year-stem.control." + elementA + "-" + elementB
	default:
		relation.RelationType = "controls"
		relation.RelationLabel = "五行相克"
		relation.Direction = "b_to_a"
		relation.RuleID = "heming.year-stem.control." + elementB + "-" + elementA
	}
	relation.Notes = "仅记录双方生年天干的确定性五行结构与方向，不据此推断关系质量、婚姻结果或事件时点。"
	return relation, true
}

func hemingFiveCombine(stemA, stemB int) (target, pairID string, ok bool) {
	pair := [2]int{stemA, stemB}
	if pair[0] > pair[1] {
		pair[0], pair[1] = pair[1], pair[0]
	}
	switch pair {
	case [2]int{0, 5}:
		return "土", "jia-ji", true
	case [2]int{1, 6}:
		return "金", "yi-geng", true
	case [2]int{2, 7}:
		return "水", "bing-xin", true
	case [2]int{3, 8}:
		return "木", "ding-ren", true
	case [2]int{4, 9}:
		return "火", "wu-gui", true
	default:
		return "", "", false
	}
}

// ──────────────────── Palace Interpretation ────────────────────

// PalaceReading holds the full interpretation for one palace.
type PalaceReading struct {
	PalaceName       string `json:"palace_name"`
	PalaceFocus      string `json:"palace_focus"`
	MainStarAnalysis string `json:"main_star_analysis"`
	AuxStarInfluence string `json:"aux_star_influence"`
	SihuaInfluence   string `json:"sihua_influence"`
	SanfangAnalysis  string `json:"sanfang_analysis"`
	PatternNotes     string `json:"pattern_notes"`
	Brightness       string `json:"brightness"`

	Summary              string                 `json:"summary"`
	KeyPoints            []string               `json:"key_points"`
	Evidence             []ReadingEvidence      `json:"evidence"`
	SanfangContext       *ReadingSanfangContext `json:"sanfang_context"`
	PatternDetails       []ReadingPatternDetail `json:"pattern_details"`
	ReviewNotes          []string               `json:"review_notes"`
	Limitations          []string               `json:"limitations"`
	EvidenceBasis        string                 `json:"evidence_basis"`
	PlacementBasis       string                 `json:"placement_basis"`
	InterpretationBasis  string                 `json:"interpretation_basis"`
	InterpretationStatus string                 `json:"interpretation_status"`
	ValidationStatus     string                 `json:"validation_status"`
	IsOutcomeConclusion  bool                   `json:"is_outcome_conclusion"`
}

// ReadingEvidence is a single computable fact used by the front-end explanation panel.
type ReadingEvidence struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
	Basis string `json:"basis"`
}

// ReadingSanfangContext describes the opposite and trine palace context.
type ReadingSanfangContext struct {
	Opposite      string   `json:"opposite"`
	Trine1        string   `json:"trine1"`
	Trine2        string   `json:"trine2"`
	OppositeStars []string `json:"opposite_stars"`
	Trine1Stars   []string `json:"trine1_stars"`
	Trine2Stars   []string `json:"trine2_stars"`
	Notes         []string `json:"notes"`
}

// ReadingPatternDetail explains why a pattern is considered related to this palace.
type ReadingPatternDetail struct {
	Name             string   `json:"name"`
	Palace           string   `json:"palace"`
	Stars            []string `json:"stars"`
	Basis            string   `json:"basis"`
	StructureStatus  string   `json:"structure_status"`
	ValidationStatus string   `json:"validation_status"`
}

// buildPalaceReading generates a reading after the service entry point has
// authenticated the complete published natal-chart contract.
func buildPalaceReading(chart *ZiWeiChart, palaceIdx int) *PalaceReading {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return nil
	}

	p := chart.Palaces[palaceIdx]
	mainStars := palaceMainStars(p)
	auxStars := palaceAuxStars(p)
	reading := &PalaceReading{
		PalaceName:           p.Name,
		PalaceFocus:          palaceFocus(p.Name),
		Evidence:             buildReadingEvidence(chart, palaceIdx),
		SanfangContext:       buildReadingSanfangContext(chart, palaceIdx),
		PatternDetails:       buildPatternDetailsForPalace(chart, palaceIdx),
		EvidenceBasis:        "mixed_deterministic_projection_and_unadjudicated_traditional_labels",
		PlacementBasis:       "deterministic_rule_projection",
		InterpretationBasis:  "traditional_rule_labels",
		InterpretationStatus: "not_adjudicated",
		ValidationStatus:     "not_adjudicated",
		IsOutcomeConclusion:  false,
	}

	// Main star analysis
	if len(mainStars) > 0 {
		reading.MainStarAnalysis = buildMainStarAnalysis(p, mainStars)
		reading.Brightness = buildBrightnessSummary(p, mainStars)
	} else {
		oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, palaceIdx)
		borrowed := mergeUniqueStrings(
			palaceMainStars(chart.Palaces[oppositeIdx]),
			palaceMainStars(chart.Palaces[trine1Idx]),
			palaceMainStars(chart.Palaces[trine2Idx]),
		)
		if len(borrowed) > 0 {
			reading.MainStarAnalysis = fmt.Sprintf("%s无主星，解读以对宫%s及三合%s、%s借星为主，借看星曜为%s。",
				p.Name, chart.Palaces[oppositeIdx].Name, chart.Palaces[trine1Idx].Name, chart.Palaces[trine2Idx].Name, strings.Join(borrowed, "、"))
		} else {
			reading.MainStarAnalysis = fmt.Sprintf("%s无主星，且对宫与三合宫均无主星可列；本宫只保留辅星、四化和十二神等结构字段。", p.Name)
		}
		reading.Brightness = "空宫"
	}

	// Auxiliary star influence
	if len(auxStars) > 0 {
		influence := buildAuxStarInfluence(auxStars)
		reading.AuxStarInfluence = influence
	} else {
		reading.AuxStarInfluence = "本宫未见辅曜或煞曜直接落入；仅记录这一结构事实。"
	}

	// Sihua influence
	if len(p.FourHua) > 0 {
		influence := buildSihuaInfluence(p.FourHua)
		reading.SihuaInfluence = influence
	} else {
		reading.SihuaInfluence = "本宫未见本命四化；对宫与三合宫的四化另在三方四正结构中列示。"
	}

	// Sanfang analysis - enhanced with SiHua interaction
	sf := getChartPalaceSanfang(chart, palaceIdx)
	enhanced := getEnhancedSanfang(chart, palaceIdx)

	base := fmt.Sprintf("对宫%s，三合%s与%s，形成三方四正格局。",
		sf.Opposite, sf.Trine1, sf.Trine2)

	// Add SiHua interaction details
	var extra []string
	if enhanced.OppositeSihua != "" {
		extra = append(extra, enhanced.OppositeSihua)
	}
	if enhanced.TrineSihua != "" {
		extra = append(extra, enhanced.TrineSihua)
	}
	if len(extra) > 0 {
		reading.SanfangAnalysis = base + " " + strings.Join(extra, "，") + "。"
	} else {
		reading.SanfangAnalysis = base + buildSanfangStarSummary(reading.SanfangContext)
	}

	reading.PatternNotes = buildPatternNotes(reading.PatternDetails)
	reading.Summary = buildPalaceSummary(chart, palaceIdx, reading)
	reading.KeyPoints = buildKeyPoints(chart, palaceIdx, reading)
	reading.ReviewNotes = buildReadingReviewNotes(chart, palaceIdx)
	reading.Limitations = buildReadingLimitations(chart, palaceIdx)

	return reading
}

func palaceMainStars(p PalaceInfo) []string {
	names := make([]string, 0, len(p.Stars))
	for _, s := range p.Stars {
		if s.Type == "major" {
			names = append(names, s.Name)
		}
	}
	return names
}

func palaceAuxStars(p PalaceInfo) []string {
	names := make([]string, 0, len(p.Stars))
	for _, s := range p.Stars {
		if s.Type != "major" {
			names = append(names, s.Name)
		}
	}
	return names
}

func palaceAllStarNames(p PalaceInfo) []string {
	return mergeUniqueStrings(palaceMainStars(p), palaceAuxStars(p), p.AdjectiveStars)
}

func palaceStarBrightness(p PalaceInfo, star string) string {
	for _, s := range p.Stars {
		if s.Name == star {
			return s.Brightness
		}
	}
	return ""
}

func buildMainStarAnalysis(p PalaceInfo, mainStars []string) string {
	var parts []string
	var descs []string
	for _, star := range mainStars {
		brightness := palaceStarBrightness(p, star)
		parts = append(parts, fmt.Sprintf("%s%s", star, formatBrightness(brightness)))
		if desc, ok := GetStarBrightness(star, brightness); ok {
			descs = append(descs, desc)
		}
	}
	prefix := fmt.Sprintf("%s主星为%s。", p.Name, strings.Join(parts, "、"))
	if len(descs) == 0 {
		return prefix
	}
	return prefix + strings.Join(descs, "；")
}

func buildBrightnessSummary(p PalaceInfo, mainStars []string) string {
	var parts []string
	for _, star := range mainStars {
		brightness := palaceStarBrightness(p, star)
		parts = append(parts, fmt.Sprintf("%s%s", star, formatBrightness(brightness)))
	}
	return strings.Join(parts, "、")
}

func formatBrightness(brightness string) string {
	if brightness == "" {
		return ""
	}
	return fmt.Sprintf("(%s)", brightness)
}

func buildReadingEvidence(chart *ZiWeiChart, palaceIdx int) []ReadingEvidence {
	p := chart.Palaces[palaceIdx]
	evidence := []ReadingEvidence{
		{Type: "palace", Label: "宫位", Value: fmt.Sprintf("%s%s", p.Name, formatBranch(p.Branch)), Basis: "宫名与地支来自本命盘公开宫位"},
	}
	if p.IsBodyPalace || chart.BodyPalace == p.Name || chart.BodyPalace == p.Branch {
		evidence = append(evidence, ReadingEvidence{Type: "body_palace", Label: "身宫", Value: p.Name, Basis: "本命盘公开字段标记本宫为身宫"})
	}
	for _, star := range palaceMainStars(p) {
		brightness := palaceStarBrightness(p, star)
		evidence = append(evidence, ReadingEvidence{Type: "main_star", Label: "主星", Value: star + formatBrightness(brightness), Basis: mainStarBasis(star, brightness)})
	}
	if len(palaceMainStars(p)) == 0 {
		oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, palaceIdx)
		borrowed := mergeUniqueStrings(
			palaceMainStars(chart.Palaces[oppositeIdx]),
			palaceMainStars(chart.Palaces[trine1Idx]),
			palaceMainStars(chart.Palaces[trine2Idx]),
		)
		evidence = append(evidence, ReadingEvidence{Type: "borrowed_star", Label: "空宫借星", Value: joinOrNone(borrowed), Basis: "本宫无主星；列出对宫与三合宫主星作为传统借星上下文"})
	}
	for _, star := range palaceAuxStars(p) {
		evidence = append(evidence, ReadingEvidence{Type: auxEvidenceType(star), Label: auxEvidenceLabel(star), Value: star, Basis: auxStarBasis(star)})
	}
	for _, hua := range p.FourHua {
		evidence = append(evidence, ReadingEvidence{Type: "four_hua", Label: "四化", Value: hua, Basis: fourHuaBasis(hua)})
	}
	for _, star := range p.AdjectiveStars {
		evidence = append(evidence, ReadingEvidence{Type: "adjective_star", Label: "杂曜", Value: star, Basis: "记录本命盘公开杂曜位置；具体个体结果未裁决"})
	}
	for _, item := range []struct {
		label string
		value string
	}{
		{"长生十二神", p.Changsheng12},
		{"博士十二神", p.Boshi12},
		{"将前十二神", p.JiangQian12},
		{"岁前十二神", p.SuiQian12},
	} {
		if item.value != "" {
			evidence = append(evidence, ReadingEvidence{Type: "twelve_shen", Label: item.label, Value: item.value, Basis: "记录本宫所属十二神序列标签；不推导阶段或事件结果"})
		}
	}
	if ctx := buildReadingSanfangContext(chart, palaceIdx); ctx != nil {
		evidence = append(evidence, ReadingEvidence{Type: "sanfang", Label: "三方四正", Value: fmt.Sprintf("对%s，三合%s、%s", ctx.Opposite, ctx.Trine1, ctx.Trine2), Basis: "由本宫索引按对宫与三合索引关系计算"})
	}
	return evidence
}

func buildReadingSanfangContext(chart *ZiWeiChart, palaceIdx int) *ReadingSanfangContext {
	if chart == nil || palaceIdx < 0 || palaceIdx >= len(chart.Palaces) {
		return nil
	}
	oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, palaceIdx)
	ctx := &ReadingSanfangContext{
		Opposite:      chart.Palaces[oppositeIdx].Name,
		Trine1:        chart.Palaces[trine1Idx].Name,
		Trine2:        chart.Palaces[trine2Idx].Name,
		OppositeStars: palaceAllStarNames(chart.Palaces[oppositeIdx]),
		Trine1Stars:   palaceAllStarNames(chart.Palaces[trine1Idx]),
		Trine2Stars:   palaceAllStarNames(chart.Palaces[trine2Idx]),
	}
	if len(ctx.OppositeStars) > 0 {
		ctx.Notes = append(ctx.Notes, fmt.Sprintf("对宫%s见%s；列为本宫对照宫位的星曜结构。", ctx.Opposite, strings.Join(ctx.OppositeStars, "、")))
	}
	if len(ctx.Trine1Stars) > 0 || len(ctx.Trine2Stars) > 0 {
		ctx.Notes = append(ctx.Notes, fmt.Sprintf("三合%s、%s分别见%s、%s；列为本宫三合宫位的星曜结构。",
			ctx.Trine1, ctx.Trine2, joinOrNone(ctx.Trine1Stars), joinOrNone(ctx.Trine2Stars)))
	}
	return ctx
}

func buildSanfangStarSummary(ctx *ReadingSanfangContext) string {
	if ctx == nil {
		return ""
	}
	return fmt.Sprintf(" 对宫星曜：%s；三合星曜：%s、%s。",
		joinOrNone(ctx.OppositeStars), joinOrNone(ctx.Trine1Stars), joinOrNone(ctx.Trine2Stars))
}

func buildPatternDetailsForPalace(chart *ZiWeiChart, palaceIdx int) []ReadingPatternDetail {
	if chart == nil || palaceIdx < 0 || palaceIdx >= len(chart.Palaces) {
		return nil
	}
	p := chart.Palaces[palaceIdx]
	mingIdx := findPalaceIndex(chart, "命宫")
	var details []ReadingPatternDetail
	add := func(name string, stars []string, basis string) {
		if patternDetected(chart, name) && basis != "" {
			details = append(details, ReadingPatternDetail{
				Name:             name,
				Palace:           p.Name,
				Stars:            stars,
				Basis:            basis,
				StructureStatus:  "matched",
				ValidationStatus: "not_adjudicated",
			})
		}
	}
	addLife := func(name string, stars []string, basis string) {
		if palaceIdx == mingIdx {
			add(name, stars, basis)
		}
	}

	pairPatterns := []struct {
		name  string
		star1 string
		star2 string
	}{
		{"紫府同宫", "紫微", "天府"},
		{"天同天梁格", "天同", "天梁"},
		{"巨日同宫", "巨门", "太阳"},
	}
	for _, pp := range pairPatterns {
		if starInSamePalace(chart, palaceIdx, pp.star1, pp.star2) {
			add(pp.name, []string{pp.star1, pp.star2}, fmt.Sprintf("%s、%s同在%s", pp.star1, pp.star2, p.Name))
		}
	}
	ziWuLianFu := starsInIndexes(chart, []int{palaceIdx}, []string{"紫微", "武曲", "廉贞", "天府"})
	if len(ziWuLianFu) >= 3 {
		add("紫武廉府", ziWuLianFu, fmt.Sprintf("%s同在%s", strings.Join(ziWuLianFu, "、"), p.Name))
	}
	if auxStarInPalace(chart, palaceIdx, []string{"禄存"}) && auxStarInPalace(chart, palaceIdx, []string{"天马"}) {
		add("禄马交驰", []string{"禄存", "天马"}, fmt.Sprintf("禄存、天马同在%s", p.Name))
	}
	if starInPalace(chart, palaceIdx, []string{"天相"}) &&
		auxStarInPalace(chart, palaceIdx, []string{"禄存"}) && auxStarInPalace(chart, palaceIdx, []string{"天马"}) {
		add("禄马佩印", []string{"天相", "禄存", "天马"}, fmt.Sprintf("天相、禄存、天马同在%s", p.Name))
	}
	if palaceIdx == mingIdx && patternDetected(chart, "日月拱照") {
		sunIdx := findMainStarPalaceIndex(chart, "太阳")
		moonIdx := findMainStarPalaceIndex(chart, "太阴")
		if sunIdx >= 0 && moonIdx >= 0 {
			basis := fmt.Sprintf("命宫在%s，太阳在%s、太阴在%s，符合固定日月拱照结构",
				p.Branch, chart.Palaces[sunIdx].Branch, chart.Palaces[moonIdx].Branch)
			add("日月拱照", []string{"太阳", "太阴"}, basis)
		}
	}
	addLife("日月反背", []string{"太阳", "太阴"}, "落陷的太阳、太阴同在命宫三方四正")
	addLife("空宫", nil, "命宫未见十四主星直接落入")
	addLife("日月并明", []string{"太阳", "太阴"}, "庙旺的太阳、太阴同在命宫三方四正")
	addLife("极向离明", []string{"紫微"}, "紫微庙旺坐午宫命宫，且命宫三方四正未见六煞")
	addLife("石中隐玉", []string{"巨门"}, fmt.Sprintf("巨门坐%s命宫", p.Branch))
	addLife("马头带箭", []string{"擎羊"}, "擎羊坐午宫命宫")
	addLife("七杀朝斗", []string{"七杀"}, fmt.Sprintf("七杀坐%s命宫，对宫主星组符合朝斗条件", p.Branch))
	addLife("日月夹命", []string{"太阳", "太阴"}, "太阳、太阴分居命宫相邻两侧")
	addLife("辅弼夹命", []string{"左辅", "右弼"}, "左辅、右弼分居命宫相邻两侧")
	addLife("科权双会", []string{"化科", "化权"}, "主星化科、化权同在命宫三方四正")
	addLife("权禄生逢", []string{"化权", "化禄"}, "庙旺主星化权、化禄同在命宫")
	addLife("魁钺夹命", []string{"天魁", "天钺"}, "天魁、天钺分居命宫相邻两侧")
	addLife("羊陀夹忌", []string{"擎羊", "陀罗", "化忌"}, "化忌坐命，擎羊、陀罗分居命宫相邻两侧")
	addLife("紫府夹命", []string{"天机", "太阴", "紫微", "天府"}, "天机、太阴坐寅申命宫，紫微、天府分居命宫相邻两侧")
	if patternDetected(chart, "日月夹财") {
		if palaceIdx == mingIdx && starInPalace(chart, mingIdx, []string{"武曲"}) && sunMoonClampPalace(chart, mingIdx) {
			add("日月夹财", []string{"武曲", "太阳", "太阴"}, "武曲坐命，太阳、太阴分居命宫相邻两侧")
		}
		wealthIdx := findPalaceIndex(chart, "财帛")
		if palaceIdx == wealthIdx && sunMoonClampPalace(chart, wealthIdx) {
			add("日月夹财", []string{"太阳", "太阴"}, "太阳、太阴分居财帛宫相邻两侧")
		}
	}
	addLife("火铃夹命", []string{"火星", "铃星"}, "火星、铃星分居命宫相邻两侧")
	addLife("空劫夹命", []string{"地空", "地劫"}, "地空、地劫分居命宫相邻两侧")
	addLife("月朗天门", []string{"太阴"}, "太阴坐亥宫命宫")
	addLife("日照雷门", []string{"太阳"}, "太阳坐卯宫命宫")
	addTanLangAux := func(name, auxName string) {
		if !patternDetected(chart, name) {
			return
		}
		tanIdx := findMainStarPalaceIndex(chart, "贪狼")
		if tanIdx < 0 {
			return
		}
		for _, auxIdx := range sanfangIndexesWithSelf(chart, tanIdx) {
			if !auxStarInPalace(chart, auxIdx, []string{auxName}) {
				continue
			}
			if !tanLangAuxRelationAllowed(auxName, chart.Palaces[tanIdx].Branch, tanIdx == auxIdx) {
				continue
			}
			if palaceIdx != mingIdx && palaceIdx != tanIdx && palaceIdx != auxIdx {
				continue
			}
			basis := fmt.Sprintf("贪狼在%s，与%s在%s会照命宫", palaceDisplayName(chart, tanIdx), auxName, palaceDisplayName(chart, auxIdx))
			if tanIdx == auxIdx {
				basis = fmt.Sprintf("贪狼、%s同在%s并会照命宫", auxName, palaceDisplayName(chart, tanIdx))
			}
			if name == "火贪格" {
				basis += "；贪狼宫位符合庙旺、卯酉见火或四墓会火条件"
			} else {
				basis += "；符合贪铃同守或四墓宫守照条件"
			}
			add(name, []string{"贪狼", auxName}, basis)
			return
		}
	}
	addTanLangAux("火贪格", "火星")
	addTanLangAux("铃贪格", "铃星")

	if patternDetected(chart, "武贪格") {
		wuIdx := findMainStarPalaceIndex(chart, "武曲")
		tanIdx := findMainStarPalaceIndex(chart, "贪狼")
		if wuIdx >= 0 && tanIdx >= 0 && (palaceIdx == mingIdx || palaceIdx == wuIdx || palaceIdx == tanIdx) {
			basis := fmt.Sprintf("武曲、贪狼同在%s的丑未身命宫结构", palaceDisplayName(chart, wuIdx))
			add("武贪格", []string{"武曲", "贪狼"}, basis)
		}
	}
	if patternDetected(chart, "杀破狼格") {
		targets := []string{"七杀", "破军", "贪狼"}
		stars := starsInIndexes(chart, sanfangIndexesWithSelf(chart, palaceIdx), targets)
		if len(stars) == len(targets) && (palaceIdx == mingIdx || anyStarInPalace(chart, palaceIdx, targets)) {
			add("杀破狼格", stars, fmt.Sprintf("%s在%s三方四正会照", strings.Join(stars, "、"), p.Name))
		}
	}
	if patternDetected(chart, "机月同梁格") {
		targets := []string{"天机", "太阴", "天同", "天梁"}
		stars := starsInIndexes(chart, sanfangIndexesWithSelf(chart, palaceIdx), targets)
		if len(stars) == len(targets) && (palaceIdx == mingIdx || anyStarInPalace(chart, palaceIdx, targets)) {
			add("机月同梁格", stars, fmt.Sprintf("%s在%s三方四正会照", strings.Join(stars, "、"), p.Name))
		}
	}
	if patternDetected(chart, "府相朝垣") {
		fuIdx := findPalaceIndex(chart, "事业")
		xiangIdx := findPalaceIndex(chart, "财帛")
		if fuIdx >= 0 && xiangIdx >= 0 && (palaceIdx == mingIdx || palaceIdx == fuIdx || palaceIdx == xiangIdx) {
			add("府相朝垣", []string{"天府", "天相"}, "天府守事业宫、天相守财帛宫，会照命宫")
		}
	}
	if patternDetected(chart, "昌曲同会") {
		targets := []string{"文昌", "文曲"}
		stars := starsInIndexes(chart, sanfangIndexesWithSelf(chart, palaceIdx), targets)
		if len(stars) == len(targets) && (palaceIdx == mingIdx || anyStarInPalace(chart, palaceIdx, targets)) {
			add("昌曲同会", stars, fmt.Sprintf("%s在%s三方四正会照", strings.Join(stars, "、"), p.Name))
		}
	}
	if patternDetected(chart, "三奇加会") {
		if basis, stars := sanqiBasisForPalace(chart, palaceIdx); basis != "" {
			add("三奇加会", stars, basis)
		}
	}
	return details
}

func findMainStarPalaceIndex(chart *ZiWeiChart, starName string) int {
	if chart == nil {
		return -1
	}
	for i := range chart.Palaces {
		if starInPalace(chart, i, []string{starName}) {
			return i
		}
	}
	return -1
}

func palaceDisplayName(chart *ZiWeiChart, palaceIdx int) string {
	if chart == nil || palaceIdx < 0 || palaceIdx >= len(chart.Palaces) {
		return "未知宫位"
	}
	if chart.Palaces[palaceIdx].Name != "" {
		return chart.Palaces[palaceIdx].Name
	}
	return chart.Palaces[palaceIdx].Branch + "宫"
}

func patternDetected(chart *ZiWeiChart, name string) bool {
	for _, p := range chart.Patterns {
		if p == name {
			return true
		}
	}
	for _, pc := range patternCheckers {
		if pc.name == name {
			ok, _ := pc.checker(chart)
			return ok
		}
	}
	return false
}

func sanqiBasisForPalace(chart *ZiWeiChart, palaceIdx int) (string, []string) {
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return "", nil
	}
	indexes := sanfangIndexesWithSelf(chart, mingIdx)
	if !containsInt(indexes, palaceIdx) {
		return "", nil
	}
	var huaTypes []string
	var huaStars []string
	for _, hua := range []string{"化禄", "化权", "化科"} {
		for _, idx := range indexes {
			for _, item := range chart.Palaces[idx].FourHua {
				if strings.Contains(item, hua) {
					huaTypes = append(huaTypes, fmt.Sprintf("%s在%s", item, chart.Palaces[idx].Name))
					huaStars = append(huaStars, item)
				}
			}
		}
	}
	if len(huaTypes) < 3 {
		return "", nil
	}
	return "化禄、化权、化科落在命宫三方四正：" + strings.Join(huaTypes, "、"), uniqueStrings(huaStars)
}

func buildPatternNotes(details []ReadingPatternDetail) string {
	if len(details) == 0 {
		return "本宫未见可由当前规则直接验证的格局标签，解读以星曜、四化、三方四正为准。"
	}
	notes := make([]string, 0, len(details))
	for _, d := range details {
		notes = append(notes, fmt.Sprintf("%s：%s", d.Name, d.Basis))
	}
	return strings.Join(notes, "；")
}

func buildPalaceSummary(chart *ZiWeiChart, palaceIdx int, reading *PalaceReading) string {
	p := chart.Palaces[palaceIdx]
	mainStars := palaceMainStars(p)
	if len(mainStars) == 0 {
		return fmt.Sprintf("%s为空宫；传统借星上下文列出对宫%s与三合%s、%s，并同时保留本宫四化和辅杂曜结构。",
			p.Name, reading.SanfangContext.Opposite, reading.SanfangContext.Trine1, reading.SanfangContext.Trine2)
	}
	summary := fmt.Sprintf("%s以%s为核心，亮度为%s", p.Name, strings.Join(mainStars, "、"), reading.Brightness)
	if len(p.FourHua) > 0 {
		summary += "，并见" + strings.Join(p.FourHua, "、")
	}
	if p.IsBodyPalace || chart.BodyPalace == p.Name || chart.BodyPalace == p.Branch {
		summary += "；公开命盘同时标记此宫为身宫"
	}
	summary += "。"
	return summary
}

func buildKeyPoints(chart *ZiWeiChart, palaceIdx int, reading *PalaceReading) []string {
	p := chart.Palaces[palaceIdx]
	var points []string
	mainStars := palaceMainStars(p)
	if len(mainStars) > 0 {
		points = append(points, fmt.Sprintf("主星组合：%s，亮度：%s。", strings.Join(mainStars, "、"), reading.Brightness))
	} else if reading.SanfangContext != nil {
		points = append(points, fmt.Sprintf("空宫借星：对宫%s，三合%s、%s。", reading.SanfangContext.Opposite, reading.SanfangContext.Trine1, reading.SanfangContext.Trine2))
	}
	if len(p.FourHua) > 0 {
		points = append(points, "本宫四化："+strings.Join(p.FourHua, "、")+"。")
	}
	if tough := filterStars(palaceAuxStars(p), toughStarSet()); len(tough) > 0 {
		points = append(points, "传统煞曜标签："+strings.Join(tough, "、")+"；只记录落宫位置。")
	}
	if p.IsBodyPalace || chart.BodyPalace == p.Name || chart.BodyPalace == p.Branch {
		points = append(points, "身宫落此宫，按规则提高本宫结构权重；不推导现实行为结果。")
	}
	if len(reading.PatternDetails) > 0 {
		points = append(points, "相关格局："+strings.Join(patternDetailNames(reading.PatternDetails), "、")+"。")
	}
	return points
}

func buildReadingReviewNotes(chart *ZiWeiChart, palaceIdx int) []string {
	p := chart.Palaces[palaceIdx]
	focus := palaceFocus(p.Name)
	notes := []string{fmt.Sprintf("本段只解释%s的可计算命盘结构，具体个体结果尚未独立裁决。", focus)}
	if len(palaceMainStars(p)) == 0 {
		notes = append(notes, fmt.Sprintf("%s为空宫，解释需同时列出对宫与三合宫的借星依据。", p.Name))
	}
	if len(p.FourHua) > 0 {
		notes = append(notes, fmt.Sprintf("%s见%s；四化只作为主题标签，不代表事件必然发生。", p.Name, strings.Join(p.FourHua, "、")))
	}
	if tough := filterStars(palaceAuxStars(p), toughStarSet()); len(tough) > 0 {
		notes = append(notes, fmt.Sprintf("%s见%s；煞曜是传统分类标签，不作为现实风险概率。", p.Name, strings.Join(tough, "、")))
	}
	if soft := filterStars(palaceAuxStars(p), softStarSet()); len(soft) > 0 {
		notes = append(notes, fmt.Sprintf("%s见%s；辅曜是传统分类标签，不作为现实助力概率。", p.Name, strings.Join(soft, "、")))
	}
	return notes
}

func buildReadingLimitations(chart *ZiWeiChart, palaceIdx int) []string {
	p := chart.Palaces[palaceIdx]
	limitations := []string{"经验规则和模板文本未进入独立 Gold 裁决，不能解释为预测准确率。"}
	for _, star := range filterStars(palaceAuxStars(p), toughStarSet()) {
		limitations = append(limitations, fmt.Sprintf("%s入%s只证明星曜位置，不证明现实损失或冲突。", star, p.Name))
	}
	for _, hua := range p.FourHua {
		if strings.Contains(hua, "化忌") {
			limitations = append(limitations, fmt.Sprintf("%s在%s只证明四化结构，不证明现实阻滞或损失。", hua, p.Name))
		}
	}
	for _, star := range palaceMainStars(p) {
		br := palaceStarBrightness(p, star)
		if br == "陷" || br == "不" {
			limitations = append(limitations, fmt.Sprintf("%s%s在%s只记录亮度等级，不证明个体能力或结果。", star, formatBrightness(br), p.Name))
		}
	}
	if len(palaceMainStars(p)) == 0 {
		limitations = append(limitations, fmt.Sprintf("%s为空宫，单宫信息不完整，必须连同对宫与三合宫展示。", p.Name))
	}
	return limitations
}

func mainStarBasis(star, brightness string) string {
	if brightness == "" {
		return fmt.Sprintf("%s为本宫主星，亮度未提供；具体个体结果未裁决", star)
	}
	return fmt.Sprintf("%s为本宫主星，亮度等级为%s；具体个体结果未裁决", star, brightness)
}

func auxEvidenceType(star string) string {
	if toughStarSet()[star] {
		return "tough_star"
	}
	if softStarSet()[star] {
		return "soft_star"
	}
	return "aux_star"
}

func auxEvidenceLabel(star string) string {
	if toughStarSet()[star] {
		return "煞曜"
	}
	if softStarSet()[star] {
		return "辅曜"
	}
	return "辅杂曜"
}

func auxStarBasis(star string) string {
	if toughStarSet()[star] {
		return "记录传统煞曜位置；不推导现实冲突、损失或个体状态"
	}
	if softStarSet()[star] {
		return "记录传统辅曜位置；不推导现实助力或个体能力"
	}
	if star == "天马" {
		return "记录传统移动类星曜位置；不推导现实变动或行动方案"
	}
	if star == "禄存" {
		return "记录传统资源类星曜位置；不推导现实收益或财务结果"
	}
	return "记录辅杂曜位置；具体个体结果未裁决"
}

func fourHuaBasis(hua string) string {
	switch {
	case strings.Contains(hua, "化禄"):
		return "记录本宫化禄标签；不推导现实结果"
	case strings.Contains(hua, "化权"):
		return "记录本宫化权标签；不推导现实结果"
	case strings.Contains(hua, "化科"):
		return "记录本宫化科标签；不推导现实结果"
	case strings.Contains(hua, "化忌"):
		return "记录本宫化忌标签；不推导现实结果"
	default:
		return "记录本命盘公开四化标签；不推导现实结果"
	}
}

func palaceFocus(name string) string {
	focus := map[string]string{
		"命宫": "性格、选择与自我定位",
		"兄弟": "同辈、协作与资源分配",
		"夫妻": "亲密关系、承诺与协商",
		"子女": "子女、下属与创造输出",
		"财帛": "现金流与资源配置",
		"疾厄": "传统疾厄宫、压力主题与节奏观察，不作个体身体状态推断",
		"迁移": "外部环境、出行与社会形象",
		"交友": "朋友、团队与合作对象",
		"事业": "职业、责任与组织角色",
		"田宅": "家庭、不动产与安全感",
		"福德": "精神状态、享受与内在稳定",
		"父母": "长辈、制度与支持来源",
	}
	if v, ok := focus[name]; ok {
		return v
	}
	return name
}

func softStarSet() map[string]bool {
	return map[string]bool{"左辅": true, "右弼": true, "文昌": true, "文曲": true, "天魁": true, "天钺": true}
}

func toughStarSet() map[string]bool {
	return map[string]bool{"擎羊": true, "陀罗": true, "火星": true, "铃星": true, "地空": true, "地劫": true}
}

func filterStars(stars []string, allowed map[string]bool) []string {
	var out []string
	for _, s := range stars {
		if allowed[s] {
			out = append(out, s)
		}
	}
	return out
}

func starsInIndexes(chart *ZiWeiChart, indexes []int, wanted []string) []string {
	var stars []string
	for _, idx := range indexes {
		if idx < 0 || idx >= len(chart.Palaces) {
			continue
		}
		for _, star := range wanted {
			if anyStarInPalace(chart, idx, []string{star}) {
				stars = append(stars, star)
			}
		}
	}
	return uniqueStrings(stars)
}

func mergeUniqueStrings(groups ...[]string) []string {
	var merged []string
	for _, group := range groups {
		merged = append(merged, group...)
	}
	return uniqueStrings(merged)
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func joinOrNone(in []string) string {
	if len(in) == 0 {
		return "无"
	}
	return strings.Join(in, "、")
}

func formatBranch(branch string) string {
	if branch == "" {
		return ""
	}
	return fmt.Sprintf("(%s)", branch)
}

func containsInt(in []int, target int) bool {
	for _, v := range in {
		if v == target {
			return true
		}
	}
	return false
}

func patternDetailNames(details []ReadingPatternDetail) []string {
	names := make([]string, 0, len(details))
	for _, detail := range details {
		names = append(names, detail.Name)
	}
	return uniqueStrings(names)
}

func buildAuxStarInfluence(auxStars []string) string {
	if len(auxStars) == 0 {
		return ""
	}
	var parts []string
	for _, star := range auxStars {
		if desc := getAuxStarDesc(star); desc != "" {
			parts = append(parts, desc)
		}
	}
	return strings.Join(parts, "；")
}

func getAuxStarDesc(star string) string {
	known := map[string]bool{
		"左辅": true, "右弼": true, "文昌": true, "文曲": true,
		"天魁": true, "天钺": true, "擎羊": true, "陀罗": true,
		"火星": true, "铃星": true, "地空": true, "地劫": true,
		"禄存": true, "天马": true,
	}
	if known[star] {
		return fmt.Sprintf("%s入此宫；仅记录星曜位置，具体个体结果未裁决。", star)
	}
	return ""
}

func buildSihuaInfluence(fourHua []string) string {
	if len(fourHua) == 0 {
		return ""
	}
	var parts []string
	for _, t := range fourHua {
		if desc := getFourHuaDesc(t); desc != "" {
			parts = append(parts, desc)
		}
	}
	return strings.Join(parts, "；")
}

func getFourHuaDesc(star string) string {
	huaTypeMap := map[string]int{"化禄": 0, "化权": 1, "化科": 2, "化忌": 3}
	for huaStr, huaIdx := range huaTypeMap {
		if strings.Contains(star, huaStr) {
			starName := strings.ReplaceAll(star, huaStr, "")
			return getHuaDesc(starName, huaIdx)
		}
	}
	return star // fallback
}

func getHuaDesc(starName string, huaType int) string {
	huaNames := []string{"化禄", "化权", "化科", "化忌"}
	if huaType >= 0 && huaType < 4 {
		return fmt.Sprintf("%s%s；仅记录本命四化标签，具体个体结果未裁决", starName, huaNames[huaType])
	}
	return starName
}

// ──────────────────── Additional Pattern Checkers ────────────────────

// findPalaceIndex returns the index of a palace by its name, or -1 if not found.
func findPalaceIndex(chart *ZiWeiChart, name string) int {
	if chart == nil {
		return -1
	}
	for i, p := range chart.Palaces {
		if p.Name == name {
			return i
		}
	}
	return -1
}

// palaceHasHua returns true if the given palace has a transformation matching huaType
// (e.g., "化禄", "化权", "化科", "化忌").
func palaceHasHua(chart *ZiWeiChart, palaceIdx int, huaType string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= 12 {
		return false
	}
	for _, h := range chart.Palaces[palaceIdx].FourHua {
		if strings.Contains(h, huaType) {
			return true
		}
	}
	return false
}

// checkKeQuanShuangHui detects 化科 and 化权 among the major stars in the
// life palace's sanfang-sizheng.
func checkKeQuanShuangHui(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	hasKe, hasQuan := false, false
	for _, idx := range sanfangIndexesWithSelf(chart, mingIdx) {
		hasKe = hasKe || palaceHasMajorHua(chart, idx, "化科")
		hasQuan = hasQuan || palaceHasMajorHua(chart, idx, "化权")
	}
	if hasKe && hasQuan {
		return true, "科权双会"
	}
	return false, ""
}

func palaceHasMajorHua(chart *ZiWeiChart, palaceIdx int, huaType string) bool {
	if chart == nil || palaceIdx < 0 || palaceIdx >= len(chart.Palaces) {
		return false
	}
	palace := chart.Palaces[palaceIdx]
	for _, label := range palace.FourHua {
		starName, gotHuaType, ok := parseFourHuaLabel(label)
		if !ok || gotHuaType != huaType {
			continue
		}
		for _, star := range palace.Stars {
			if star.Name == starName && star.Type == "major" {
				return true
			}
		}
	}
	return false
}

// checkQuanLuShengFeng detects bright 化权 and 化禄 stars together in 命宫.
func checkQuanLuShengFeng(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	hasBrightQuan, hasBrightLu := false, false
	for _, label := range chart.Palaces[mingIdx].FourHua {
		starName, huaType, ok := parseFourHuaLabel(label)
		if !ok || !hasBrightness(chart, mingIdx, starName, []string{"庙", "旺"}) {
			continue
		}
		hasBrightQuan = hasBrightQuan || huaType == "化权"
		hasBrightLu = hasBrightLu || huaType == "化禄"
	}
	if hasBrightQuan && hasBrightLu {
		return true, "权禄生逢"
	}
	return false, ""
}

// checkKuiYueJiaMing 魁钺夹命：
// 天魁、天钺分别落在命宫两侧的宫位
func checkKuiYueJiaMing(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 {
		return false, ""
	}
	left := (mingIdx + 11) % 12
	right := (mingIdx + 1) % 12
	hasKuiLeft := auxStarInPalace(chart, left, []string{"天魁"})
	hasYueRight := auxStarInPalace(chart, right, []string{"天钺"})
	hasYueLeft := auxStarInPalace(chart, left, []string{"天钺"})
	hasKuiRight := auxStarInPalace(chart, right, []string{"天魁"})
	if (hasKuiLeft && hasYueRight) || (hasYueLeft && hasKuiRight) {
		return true, "魁钺夹命"
	}
	return false, ""
}

// checkYangTuoJiaJi 羊陀夹忌：化忌坐命，擎羊、陀罗分居命宫两侧。
func checkYangTuoJiaJi(chart *ZiWeiChart) (bool, string) {
	if chart == nil {
		return false, ""
	}
	mingIdx := findPalaceIndex(chart, "命宫")
	if mingIdx < 0 || !palaceHasHua(chart, mingIdx, "化忌") {
		return false, ""
	}
	left := (mingIdx + 11) % 12
	right := (mingIdx + 1) % 12
	hasYangLeft := auxStarInPalace(chart, left, []string{"擎羊"})
	hasTuoRight := auxStarInPalace(chart, right, []string{"陀罗"})
	hasTuoLeft := auxStarInPalace(chart, left, []string{"陀罗"})
	hasYangRight := auxStarInPalace(chart, right, []string{"擎羊"})
	if (hasYangLeft && hasTuoRight) || (hasTuoLeft && hasYangRight) {
		return true, "羊陀夹忌"
	}
	return false, ""
}
