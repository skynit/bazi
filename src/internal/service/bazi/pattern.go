package bazi

import (
	"bazi/internal/model"
)

type PatternAnalysis struct {
	RuleID                 string                                       `json:"rule_id"`
	SchemaVersion          string                                       `json:"schema_version"`
	DetectorProfile        string                                       `json:"detector_profile"`
	DetectorCount          int                                          `json:"detector_count"`
	DetectorManifestSHA256 string                                       `json:"detector_manifest_sha256"`
	DetectorProfiles       []PatternDetectorProfileDigest               `json:"detector_profiles"`
	DetectorChangeContract PatternDetectorProfileChangeContract         `json:"detector_profile_change_contract"`
	DetectorMigration      PatternDetectorProfileMigrationReference     `json:"detector_profile_migration"`
	DetectorReleaseAnchor  PatternDetectorProfileReleaseAnchorReference `json:"detector_profile_release_anchor"`
	Inputs                 PatternInputSnapshot                         `json:"inputs"`
	Candidates             []PatternCandidate                           `json:"candidates"`
	MonthCommandEvidence   []MonthCommandPatternEvidence                `json:"month_command_evidence"`
	Status                 string                                       `json:"status"`
	ValidationStatus       string                                       `json:"validation_status"`
	InterpretationStatus   string                                       `json:"interpretation_status"`
	Limitations            []string                                     `json:"limitations"`
}

// MonthCommandPatternEvidence records the classical "月支之神，透于天干"
// condition without promoting it to an established ordinary pattern. Exact
// hidden-stem exposure requires the same stem to appear visibly; an
// opposite-polarity stem of the same element is not the month command exposed.
type MonthCommandPatternEvidence struct {
	RuleID                string                     `json:"rule_id"`
	MonthBranch           string                     `json:"month_branch"`
	HiddenStem            string                     `json:"hidden_stem"`
	HiddenStemType        string                     `json:"hidden_stem_type"`
	HiddenTenGod          string                     `json:"hidden_ten_god"`
	Exposures             []MonthCommandStemExposure `json:"exposures"`
	CandidateNames        []string                   `json:"candidate_names"`
	ExposureStatus        string                     `json:"exposure_status"`
	MonthSpecialStructure string                     `json:"month_special_structure,omitempty"`
	Source                string                     `json:"source"`
	Status                string                     `json:"status"`
	InterpretationStatus  string                     `json:"interpretation_status"`
	IsEstablishedPattern  bool                       `json:"is_established_pattern"`
}

type MonthCommandStemExposure struct {
	Pillar          string `json:"pillar"`
	Stem            string `json:"stem"`
	TenGod          string `json:"ten_god"`
	ExactHiddenStem bool   `json:"exact_hidden_stem"`
}

type PatternDetectorProfileDigest struct {
	RuleID          string `json:"rule_id"`
	AlgorithmSHA256 string `json:"algorithm_sha256"`
	BehaviorSHA256  string `json:"behavior_sha256"`
	ProfileSHA256   string `json:"profile_sha256"`
}

type patternDetection struct {
	PatternName string
}

func fixedPatternDetection(ruleID string) *patternDetection {
	name, ok := patternDetectorSingleOutputName(ruleID)
	if !ok {
		return nil
	}
	return &patternDetection{PatternName: name}
}

// AnalyzePatternExtended 八字格局检测主入口。所有成立规则都会进入候选集：
//  1. 专旺格（曲直/炎上/稼穑/从革/润下）
//  2. 月令建禄 / 月刃 / 专禄 / 日刃
//  3. 两气成象格
//  4. 三奇格 / 魁罡格 / 金神格 / 日德格
func AnalyzePatternExtended(pillars []model.Pillar, monthZhi string) PatternAnalysis {
	return analyzePatternCandidates(pillars, monthZhi)
}

func checkZhuanWangGe(pillars []model.Pillar) *patternDetection {
	semanticProfile := zhuanWangDetectorSemanticProfile()
	contextProfile := patternPillarContextSemanticProfile()
	if len(pillars) != semanticProfile.PillarCount {
		return nil
	}
	if contextProfile.DayPillarIndex < 0 || contextProfile.DayPillarIndex >= len(pillars) {
		return nil
	}
	dayElement, ok := patternElementForSymbol(semanticProfile.StemElements, pillars[contextProfile.DayPillarIndex].Gan)
	if !ok {
		return nil
	}
	profile, ok := zhuanWangProfileForElement(dayElement)
	if !ok {
		return nil
	}
	for _, pillar := range pillars {
		stemElement, ok := patternElementForSymbol(semanticProfile.StemElements, pillar.Gan)
		if !ok {
			return nil
		}
		if stemElement == profile.breakingElement {
			return nil
		}
	}

	matchedStructure := false
	for _, structure := range profile.structures {
		if !containsAllBranches(pillars, structure.branches...) {
			continue
		}
		broken := false
		for _, pillar := range pillars {
			if inStrings(pillar.Zhi, structure.branches...) {
				continue
			}
			branchElement, ok := patternElementForSymbol(semanticProfile.BranchElements, pillar.Zhi)
			if !ok {
				return nil
			}
			if branchElement == profile.breakingElement {
				broken = true
				break
			}
		}
		if !broken {
			matchedStructure = true
			break
		}
	}
	if !matchedStructure {
		return nil
	}
	return &patternDetection{
		PatternName: profile.name,
	}
}

type zhuanWangStructure struct {
	branches []string
}

type zhuanWangProfile struct {
	name            string
	breakingElement string
	structures      []zhuanWangStructure
}

func zhuanWangProfileRegistry() map[string]zhuanWangProfile {
	return map[string]zhuanWangProfile{
		"木": {name: "曲直格", breakingElement: "金", structures: []zhuanWangStructure{
			{branches: []string{"寅", "卯", "辰"}},
			{branches: []string{"亥", "卯", "未"}},
		}},
		"火": {name: "炎上格", breakingElement: "水", structures: []zhuanWangStructure{
			{branches: []string{"巳", "午", "未"}},
			{branches: []string{"寅", "午", "戌"}},
		}},
		"土": {name: "稼穑格", breakingElement: "木", structures: []zhuanWangStructure{
			{branches: []string{"辰", "戌", "丑", "未"}},
		}},
		"金": {name: "从革格", breakingElement: "火", structures: []zhuanWangStructure{
			{branches: []string{"申", "酉", "戌"}},
			{branches: []string{"巳", "酉", "丑"}},
		}},
		"水": {name: "润下格", breakingElement: "土", structures: []zhuanWangStructure{
			{branches: []string{"亥", "子", "丑"}},
			{branches: []string{"申", "子", "辰"}},
		}},
	}
}

func zhuanWangProfileForElement(element string) (zhuanWangProfile, bool) {
	profile, ok := zhuanWangProfileRegistry()[element]
	return profile, ok
}

func containsAllBranches(pillars []model.Pillar, branches ...string) bool {
	for _, branch := range branches {
		found := false
		for _, pillar := range pillars {
			if pillar.Zhi == branch {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func checkLiangQiChengXiang(pillars []model.Pillar) *patternDetection {
	profile := liangQiSemanticProfile()
	if len(pillars) != profile.PillarCount {
		return nil
	}
	counts := make(map[string]int, profile.DistinctElements)
	for _, pillar := range pillars {
		ganElement, ganOK := patternElementForSymbol(profile.StemElements, pillar.Gan)
		zhiElement, zhiOK := patternElementForSymbol(profile.BranchElements, pillar.Zhi)
		if !ganOK || !zhiOK {
			return nil
		}
		counts[ganElement]++
		counts[zhiElement]++
	}
	if len(counts) != profile.DistinctElements {
		return nil
	}
	elements := make([]string, 0, profile.DistinctElements)
	for _, element := range profile.ElementOrder {
		if count := counts[element]; count != 0 {
			if count != profile.OccurrencesPerElement {
				return nil
			}
			elements = append(elements, element)
		}
	}
	if len(elements) != profile.DistinctElements {
		return nil
	}
	return fixedPatternDetection("pattern.special.liangqi")
}

func checkKuiGangGe(gan, zhi string) *patternDetection {
	dayCol := gan + zhi
	if !patternStringProfileContains(kuiGangDayProfile(), dayCol) {
		return nil
	}
	return fixedPatternDetection("pattern.aux.kuigang")
}

func checkRiDeGe(gan, zhi string) *patternDetection {
	if patternStringProfileContains(riDeDayProfile(), gan+zhi) {
		return fixedPatternDetection("pattern.aux.ride")
	}
	return nil
}

// checkJianLuGe records the month-branch lu structure. Whether exposed
// finance, office, seal, food, or other conditions complete the pattern is
// intentionally left unadjudicated.
func checkJianLuGe(dayGan, monthZhi string) *patternDetection {
	luBranch, ok := luBranchForStem(dayGan)
	if !ok || monthZhi == "" || monthZhi != luBranch {
		return nil
	}
	return fixedPatternDetection("pattern.lu.jianlu")
}

// checkYueRenGe records a five-yang-stem blade in the month branch. Officer,
// killer, clash, combination, finance, seal, and food conditions remain
// separate evidence and are not inferred from this match.
func checkYueRenGe(dayGan, monthZhi string) *patternDetection {
	if monthZhi == "" || monthZhi != yangRenZhi(dayGan) {
		return nil
	}
	return fixedPatternDetection("pattern.lu.yueren")
}

// checkZhuanLuGe records the four valid self-lu day structures. Month lu is
// an independent fact, so a chart may carry both 专禄 and 建禄 candidates.
func checkZhuanLuGe(dayGan, dayZhi string) *patternDetection {
	luBranch, ok := luBranchForStem(dayGan)
	if !ok || dayZhi != luBranch {
		return nil
	}
	return fixedPatternDetection("pattern.lu.zhuanlu")
}

// checkRiRenGe records the three valid day-pillar blade structures. Month
// blade is independent, so both candidates may be present in one chart.
func checkRiRenGe(dayGan, dayZhi string) *patternDetection {
	if dayZhi != yangRenZhi(dayGan) {
		return nil
	}
	return fixedPatternDetection("pattern.lu.riren")
}

// 《渊海子平》金神专章明确只取癸酉、己巳、乙丑三种时柱。
// 火局、行运与制伏条件未在此候选中裁决，因此只作为辅助结构。
func checkJinShenHour(pillars []model.Pillar) *patternDetection {
	jinShenProfile := jinShenSemanticProfile()
	if len(pillars) != jinShenProfile.PillarCount || jinShenProfile.PillarIndex < 0 ||
		jinShenProfile.PillarIndex >= len(pillars) || !isJinShenHourPillar(pillars[jinShenProfile.PillarIndex]) {
		return nil
	}
	return fixedPatternDetection("pattern.aux.jinshen")
}

func checkSanQi(pillars []model.Pillar) *patternDetection {
	gans := make([]string, 0, len(pillars))
	for _, pillar := range pillars {
		gans = append(gans, pillar.Gan)
	}
	if classicalSanQiSequence(gans) != "" {
		return fixedPatternDetection("pattern.aux.sanqi")
	}
	return nil
}

func yangRenZhi(gan string) string {
	branch, _ := patternBranchForStem(yangRenProfile(), gan)
	return branch
}
