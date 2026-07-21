package bazi

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"sort"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

const (
	PatternRuleID          = "bazi.pattern-candidate-set-v34"
	PatternSchemaVersion   = "pattern-candidates-2026-07-17.27"
	PatternDetectorProfile = "classical_structural_detectors_v45"
)

const (
	patternCategoryStructural = "结构格局"
	patternCategoryAuxiliary  = "辅助特征"
	monthCommandPatternRuleID = "bazi.pattern.month-command-exposure.v1"
)

const monthCommandPatternSource = "《滴天髓阐微》PDF第54-55页八格：先观月令、再看天干透出何神、再究司令真假；月支之神透于天干方为格候选"

// PatternCandidate records every detector that matched. RuleID is the single
// candidate identity; Source makes the theoretical convention auditable.
type PatternCandidate struct {
	RuleID      string `json:"rule_id"`
	PatternName string `json:"pattern_name"`
	Category    string `json:"category"`
	Source      string `json:"source"`
}

type PatternInputSnapshot struct {
	Pillars     []string `json:"pillars"`
	MonthBranch string   `json:"month_branch"`
}

type patternMatch struct {
	analysis patternDetection
	ruleID   string
	source   string
	category string
}

type patternDetectorContext struct {
	pillars     []model.Pillar
	monthBranch string
	dayGan      string
	dayZhi      string
}

type patternDetectorDefinition struct {
	ruleID          string
	source          string
	category        string
	outputNames     []string
	algorithmSHA256 string
	behaviorSHA256  string
	profileSHA256   string
	detect          func(patternDetectorContext) *patternDetection
}

type patternDetectorManifestEntry struct {
	RuleID          string   `json:"rule_id"`
	Source          string   `json:"source"`
	Category        string   `json:"category"`
	OutputNames     []string `json:"output_names"`
	AlgorithmSHA256 string   `json:"algorithm_sha256"`
	BehaviorSHA256  string   `json:"behavior_sha256"`
	ProfileSHA256   string   `json:"profile_sha256"`
}

func patternDetectorRegistry() []patternDetectorDefinition {
	detectors := []patternDetectorDefinition{
		{ruleID: "pattern.special.zhuanwang", source: "《滴天髓阐微》PDF第44-45页独象：方局全、四库全且不杂克神", category: patternCategoryStructural,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkZhuanWangGe(ctx.pillars) }},
		{ruleID: "pattern.lu.jianlu", source: "《三命通会》PDF第230-232页月令建禄十干表及取用条件", category: patternCategoryStructural,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkJianLuGe(ctx.dayGan, ctx.monthBranch) }},
		{ruleID: "pattern.lu.yueren", source: "《三命通会》PDF第226、228-230页五阳干阳刃及月柱条件", category: patternCategoryStructural,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkYueRenGe(ctx.dayGan, ctx.monthBranch) }},
		{ruleID: "pattern.lu.zhuanlu", source: "《三命通会》PDF第190页甲寅乙卯庚申辛酉专禄结构", category: patternCategoryStructural,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkZhuanLuGe(ctx.dayGan, ctx.dayZhi) }},
		{ruleID: "pattern.lu.riren", source: "《渊海子平》PDF第217页丙午戊午壬子日刃三日表；《三命通会》PDF第230页日刃同羊刃", category: patternCategoryStructural,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkRiRenGe(ctx.dayGan, ctx.dayZhi) }},
		{ruleID: "pattern.special.liangqi", source: "《滴天髓阐微》PDF第43页两气双清、生克十局各半且不可夹杂", category: patternCategoryStructural,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkLiangQiChengXiang(ctx.pillars) }},
		{ruleID: "pattern.aux.kuigang", source: "《三命通会》PDF第186-187页魁罡四日规则", category: patternCategoryAuxiliary,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkKuiGangGe(ctx.dayGan, ctx.dayZhi) }},
		{ruleID: "pattern.aux.jinshen", source: "《渊海子平》PDF第221页金神时柱规则", category: patternCategoryAuxiliary,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkJinShenHour(ctx.pillars) }},
		{ruleID: "pattern.aux.sanqi", source: "《三命通会》PDF第100-102页三奇顺布规则", category: patternCategoryAuxiliary,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkSanQi(ctx.pillars) }},
		{ruleID: "pattern.aux.ride", source: "《三命通会》PDF第185-186页日德五日规则", category: patternCategoryAuxiliary,
			detect: func(ctx patternDetectorContext) *patternDetection { return checkRiDeGe(ctx.dayGan, ctx.dayZhi) }},
	}
	for index := range detectors {
		detectors[index].outputNames = patternDetectorOutputNames(detectors[index].ruleID)
		if algorithm, ok := patternDetectorAlgorithmProfileForRule(detectors[index].ruleID); ok {
			detectors[index].algorithmSHA256 = algorithm.ASTSHA256
		}
		detectors[index].behaviorSHA256 = patternDetectorBehaviorSHA256(detectors[index].ruleID)
		detectors[index].profileSHA256 = patternDetectorProfileSHA256(detectors[index].ruleID)
	}
	return detectors
}

func patternDetectorManifest(detectors []patternDetectorDefinition) []patternDetectorManifestEntry {
	manifest := make([]patternDetectorManifestEntry, 0, len(detectors))
	for _, detector := range detectors {
		manifest = append(manifest, patternDetectorManifestEntry{
			RuleID:          detector.ruleID,
			Source:          detector.source,
			Category:        detector.category,
			OutputNames:     append([]string(nil), detector.outputNames...),
			AlgorithmSHA256: detector.algorithmSHA256,
			BehaviorSHA256:  detector.behaviorSHA256,
			ProfileSHA256:   detector.profileSHA256,
		})
	}
	return manifest
}

func patternDetectorManifestSHA256(detectors []patternDetectorDefinition) string {
	manifest := patternDetectorManifest(detectors)
	sort.Slice(manifest, func(i, j int) bool { return manifest[i].RuleID < manifest[j].RuleID })
	payload, err := json.Marshal(manifest)
	if err != nil {
		panic(err)
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

func patternDetectorProfileDigests(detectors []patternDetectorDefinition) []PatternDetectorProfileDigest {
	digests := make([]PatternDetectorProfileDigest, 0, len(detectors))
	for _, detector := range detectors {
		digests = append(digests, PatternDetectorProfileDigest{
			RuleID:          detector.ruleID,
			AlgorithmSHA256: detector.algorithmSHA256,
			BehaviorSHA256:  detector.behaviorSHA256,
			ProfileSHA256:   detector.profileSHA256,
		})
	}
	sort.Slice(digests, func(i, j int) bool { return digests[i].RuleID < digests[j].RuleID })
	return digests
}

func patternDetectorCount() int {
	return len(patternDetectorRegistry())
}

func analyzePatternCandidates(pillars []model.Pillar, monthZhi string) PatternAnalysis {
	detectors := patternDetectorRegistry()
	detectorCount := len(detectors)
	detectorManifestSHA256 := patternDetectorManifestSHA256(detectors)
	detectorProfiles := patternDetectorProfileDigests(detectors)
	if !validPatternInputs(pillars, monthZhi) {
		return invalidPatternAnalysis(pillars, monthZhi, detectorCount, detectorManifestSHA256, detectorProfiles)
	}
	contextProfile := patternPillarContextSemanticProfile()
	ctx := patternDetectorContext{
		pillars:     pillars,
		monthBranch: pillars[contextProfile.MonthPillarIndex].Zhi,
		dayGan:      pillars[contextProfile.DayPillarIndex].Gan,
		dayZhi:      pillars[contextProfile.DayPillarIndex].Zhi,
	}
	matches := make([]patternMatch, 0, detectorCount)
	for _, detector := range detectors {
		pattern := detector.detect(ctx)
		if !validPatternDetectorOutput(detector, pattern) {
			continue
		}
		matches = append(matches, patternMatch{
			analysis: *pattern,
			ruleID:   detector.ruleID,
			source:   detector.source,
			category: detector.category,
		})
	}

	return finalizePatternMatches(matches, pillars, monthZhi, detectorCount, detectorManifestSHA256, detectorProfiles)
}

func validPatternDetectorOutput(detector patternDetectorDefinition, pattern *patternDetection) bool {
	return pattern != nil && pattern.PatternName != "" && patternStringProfileContains(detector.outputNames, pattern.PatternName)
}

func validPatternInputs(pillars []model.Pillar, monthZhi string) bool {
	contextProfile := patternPillarContextSemanticProfile()
	if !validPatternPillarContextProfile(contextProfile) || len(pillars) != contextProfile.PillarCount ||
		pillars[contextProfile.MonthPillarIndex].Zhi != monthZhi {
		return false
	}
	for _, pillar := range pillars {
		if _, err := (tyme.SixtyCycle{}).FromName(pillar.Gan + pillar.Zhi); err != nil {
			return false
		}
	}
	return true
}

func validPatternPillarContextProfile(profile patternPillarContextProfile) bool {
	if profile.PillarCount <= 0 || profile.DeclaredMonthBranchPolicy != "must_equal_month_pillar_branch" {
		return false
	}
	seen := make(map[int]struct{}, profile.PillarCount)
	for _, index := range []int{
		profile.YearPillarIndex,
		profile.MonthPillarIndex,
		profile.DayPillarIndex,
		profile.HourPillarIndex,
	} {
		if index < 0 || index >= profile.PillarCount {
			return false
		}
		if _, exists := seen[index]; exists {
			return false
		}
		seen[index] = struct{}{}
	}
	return len(seen) == profile.PillarCount
}

func invalidPatternAnalysis(pillars []model.Pillar, monthZhi string, detectorCount int, detectorManifestSHA256 string, detectorProfiles []PatternDetectorProfileDigest) PatternAnalysis {
	return PatternAnalysis{
		RuleID:                 PatternRuleID,
		SchemaVersion:          PatternSchemaVersion,
		DetectorProfile:        PatternDetectorProfile,
		DetectorCount:          detectorCount,
		DetectorManifestSHA256: detectorManifestSHA256,
		DetectorProfiles:       append([]PatternDetectorProfileDigest(nil), detectorProfiles...),
		DetectorChangeContract: patternDetectorProfileChangeContract(),
		DetectorMigration:      patternDetectorProfileMigrationReference(),
		DetectorReleaseAnchor:  patternDetectorProfileReleaseAnchorReference(),
		Inputs:                 patternInputSnapshot(pillars, monthZhi),
		Candidates:             []PatternCandidate{},
		MonthCommandEvidence:   []MonthCommandPatternEvidence{},
		Status:                 "invalid_input",
		ValidationStatus:       "invalid_input",
		InterpretationStatus:   "not_available",
		Limitations: []string{
			"pattern detection requires exactly four valid sixty-cycle pillars",
			"declared month branch must match the month pillar",
			"invalid or incomplete input never creates a fallback pattern candidate",
		},
	}
}

func finalizePatternMatches(matches []patternMatch, pillars []model.Pillar, monthZhi string, detectorCount int, detectorManifestSHA256 string, detectorProfiles []PatternDetectorProfileDigest) PatternAnalysis {
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ruleID < matches[j].ruleID
	})

	candidates := make([]PatternCandidate, 0, len(matches))
	hasStructuralCandidate := false
	for _, match := range matches {
		if match.category == patternCategoryStructural {
			hasStructuralCandidate = true
		}
		candidates = append(candidates, PatternCandidate{
			RuleID:      match.ruleID,
			PatternName: match.analysis.PatternName,
			Category:    match.category,
			Source:      match.source,
		})
	}

	result := PatternAnalysis{
		RuleID:                 PatternRuleID,
		SchemaVersion:          PatternSchemaVersion,
		DetectorProfile:        PatternDetectorProfile,
		DetectorCount:          detectorCount,
		DetectorManifestSHA256: detectorManifestSHA256,
		DetectorProfiles:       append([]PatternDetectorProfileDigest(nil), detectorProfiles...),
		DetectorChangeContract: patternDetectorProfileChangeContract(),
		DetectorMigration:      patternDetectorProfileMigrationReference(),
		DetectorReleaseAnchor:  patternDetectorProfileReleaseAnchorReference(),
		Inputs:                 patternInputSnapshot(pillars, monthZhi),
		Candidates:             candidates,
		MonthCommandEvidence:   observeMonthCommandPatternEvidence(pillars),
		Status:                 "observed",
		ValidationStatus:       "not_validated",
		InterpretationStatus:   "not_adjudicated",
		Limitations: []string{
			"detector conditions are local classical-text Profiles without expert Gold adjudication",
			"candidate list order is deterministic serialization only and does not rank or adjudicate patterns",
			"month-command exposure records ordinary-pattern evidence but does not adjudicate commanding depth, pattern purity, support, control, or breakage",
			"candidates do not determine favorable elements or real-world outcomes",
		},
	}
	if !hasStructuralCandidate {
		result.Status = "observed_without_structural_candidate"
		result.Limitations = append(result.Limitations, "absence of a structural detector match is not a normal-pattern conclusion")
	}
	return result
}

func observeMonthCommandPatternEvidence(pillars []model.Pillar) []MonthCommandPatternEvidence {
	result := make([]MonthCommandPatternEvidence, 0, 3)
	context := patternPillarContextSemanticProfile()
	if !validPatternPillarContextProfile(context) || len(pillars) != context.PillarCount {
		return result
	}
	monthPillar := pillars[context.MonthPillarIndex]
	dayStem := pillars[context.DayPillarIndex].Gan
	monthBranch, err := (tyme.EarthBranch{}).FromName(monthPillar.Zhi)
	if err != nil {
		return result
	}
	specialStructure := monthCommandSpecialStructure(dayStem, monthPillar.Zhi)
	visibleStems := []struct {
		index int
		label string
	}{
		{context.YearPillarIndex, "年干"},
		{context.MonthPillarIndex, "月干"},
		{context.HourPillarIndex, "时干"},
	}

	for _, hidden := range monthBranch.GetHideHeavenStems() {
		hiddenStem := hidden.GetHeavenStem().GetName()
		hiddenTenGod := ClassifyTenGod(hiddenStem, dayStem, false)
		if hiddenTenGod == "" || hiddenTenGod == "比肩" || hiddenTenGod == "劫财" {
			continue
		}
		exposures := make([]MonthCommandStemExposure, 0, 3)
		candidateNames := []string{ordinaryPatternName(hiddenTenGod)}
		for _, visible := range visibleStems {
			visibleStem := pillars[visible.index].Gan
			if visibleStem != hiddenStem {
				continue
			}
			exposures = append(exposures, MonthCommandStemExposure{
				Pillar: visible.label, Stem: visibleStem, TenGod: hiddenTenGod, ExactHiddenStem: true,
			})
		}
		if len(exposures) == 0 {
			continue
		}
		result = append(result, MonthCommandPatternEvidence{
			RuleID:                monthCommandPatternRuleID,
			MonthBranch:           monthPillar.Zhi,
			HiddenStem:            hiddenStem,
			HiddenStemType:        hideStemTypeLabel(hidden.GetType()),
			HiddenTenGod:          hiddenTenGod,
			Exposures:             exposures,
			CandidateNames:        candidateNames,
			ExposureStatus:        "exact_hidden_stem_exposed",
			MonthSpecialStructure: specialStructure,
			Source:                monthCommandPatternSource,
			Status:                "observed",
			InterpretationStatus:  "pattern_candidate_not_adjudicated",
			IsEstablishedPattern:  false,
		})
	}
	return result
}

func monthCommandSpecialStructure(dayStem, monthBranch string) string {
	if luBranch, ok := luBranchForStem(dayStem); ok && monthBranch == luBranch {
		return "建禄月"
	}
	if monthBranch == yangRenZhi(dayStem) {
		return "月刃"
	}
	return ""
}

func ordinaryPatternName(tenGod string) string {
	return map[string]string{
		"正官": "正官格", "七杀": "七杀格",
		"正财": "正财格", "偏财": "偏财格",
		"正印": "正印格", "偏印": "偏印格",
		"食神": "食神格", "伤官": "伤官格",
	}[tenGod]
}

func patternInputSnapshot(pillars []model.Pillar, monthZhi string) PatternInputSnapshot {
	inputs := PatternInputSnapshot{
		Pillars:     make([]string, 0, len(pillars)),
		MonthBranch: monthZhi,
	}
	for _, pillar := range pillars {
		inputs.Pillars = append(inputs.Pillars, pillar.Gan+pillar.Zhi)
	}
	return inputs
}

func ValidPatternAnalysis(analysis PatternAnalysis, pillars []model.Pillar, monthZhi string) bool {
	if !validPatternInputs(pillars, monthZhi) {
		return false
	}
	want := analyzePatternCandidates(pillars, monthZhi)
	return reflect.DeepEqual(analysis, want)
}
