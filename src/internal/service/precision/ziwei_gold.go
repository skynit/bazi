package precision

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	ziweipkg "bazi/internal/service/ziwei"
)

const ziweiGoldPurpose = "ziwei_full_chart_gold"

type ziweiGoldFile struct {
	Version       string          `json:"version"`
	Description   string          `json:"description"`
	ProfileID     string          `json:"profile_id"`
	EngineVersion string          `json:"engine_version"`
	RuleVersion   string          `json:"rule_version"`
	Frozen        bool            `json:"frozen"`
	DatasetHash   string          `json:"dataset_hash"`
	Cases         []ziweiGoldCase `json:"cases"`
}

type ziweiGoldCase struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Metadata *CaseMetadata     `json:"metadata"`
	Year     int               `json:"year"`
	Month    int               `json:"month"`
	Day      int               `json:"day"`
	Hour     int               `json:"hour"`
	Minute   int               `json:"minute"`
	Gender   string            `json:"gender"`
	Expected ziweiGoldExpected `json:"expected"`
}

type ziweiGoldExpected struct {
	SoulPalaceBranch string                `json:"soul_palace_branch"`
	BodyPalaceBranch string                `json:"body_palace_branch"`
	BodyPalace       string                `json:"body_palace"`
	LifeMaster       string                `json:"life_master"`
	BodyMaster       string                `json:"body_master"`
	FiveBureau       string                `json:"five_bureau"`
	Palaces          []ziweiGoldPalace     `json:"palaces"`
	Dayun            []ziweipkg.DayunStage `json:"dayun"`
}

type ziweiGoldPalace struct {
	Name         string   `json:"name"`
	Branch       string   `json:"branch"`
	HeavenlyStem string   `json:"heavenly_stem"`
	IsBodyPalace bool     `json:"is_body_palace"`
	MainStars    []string `json:"main_stars"`
	AuxStars     []string `json:"aux_stars"`
	FourHua      []string `json:"four_hua"`
	Changsheng12 string   `json:"changsheng_12"`
	Boshi12      string   `json:"boshi_12"`
	JiangQian12  string   `json:"jiang_qian_12"`
	SuiQian12    string   `json:"sui_qian_12"`
}

func evaluateZiweiGold(path string) ModuleReport {
	module := newModuleReport("ziwei_full_chart_gold", path)
	module.ProfileID = ziweipkg.DefaultProfileID
	module.EngineVersion = ziweipkg.ZiWeiEngineVersion
	module.RuleVersion = ziweipkg.ZiWeiRuleVersion
	module.RuleSchool = ziweipkg.ZiWeiRuleSchool

	file, err := loadZiweiGoldFile(path)
	if err != nil {
		module.Warnings = append(module.Warnings, err.Error())
		return module
	}
	module.Cases = len(file.Cases)
	if file.Frozen {
		module.BoundaryStatus = "frozen"
	} else {
		module.BoundaryStatus = "collecting"
	}
	if len(file.Cases) == 0 {
		module.Warnings = append(module.Warnings, "Ziwei full-chart Gold registry contains 0 cases; no accuracy claim is eligible")
		return module
	}

	fileReasons := validateZiweiGoldFile(file)
	duplicates := duplicateZiweiGoldIDs(file.Cases)
	duplicateSet := stringSet(duplicates)
	module.DuplicateCaseIDs = duplicates
	service := ziweipkg.NewZiWeiService()
	for _, tc := range file.Cases {
		reasons := append([]string(nil), fileReasons...)
		reasons = append(reasons, validateZiweiGoldCase(tc)...)
		metadata := tc.Metadata
		if len(reasons) > 0 {
			copyMetadata := CaseMetadata{Tier: TierGold, ReviewStatus: "quarantined", Purpose: ziweiGoldPurpose, QuarantineReason: strings.Join(reasons, "; ")}
			if metadata != nil {
				copyMetadata = *metadata
				copyMetadata.Reviewers = append([]string(nil), metadata.Reviewers...)
				copyMetadata.ReviewStatus = "quarantined"
				copyMetadata.QuarantineReason = strings.Join(reasons, "; ")
			}
			metadata = &copyMetadata
		}
		tier, publishable := registerCase(&module, metadata, duplicateSet[tc.ID])
		if !publishable {
			markSkippedCase(&module, tier, "gold_admission_failed")
			if len(reasons) > 0 {
				module.Warnings = append(module.Warnings, fmt.Sprintf("%s Gold admission failed: %s", tc.ID, strings.Join(reasons, "; ")))
			}
			continue
		}

		chart, calculateErr := service.CalculateChartWithProfile(file.ProfileID, tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, tc.Gender)
		if calculateErr != nil {
			recordDiagnosticFailure(&module, tier, true, CheckResult{CaseID: tc.ID, Field: "calculate", Status: "failed", Note: calculateErr.Error()})
			continue
		}
		markEvaluated(&module, tier)
		checks := ziweiGoldChecks(chart, service.CalculateDayun(chart), tc.Expected)
		counts := applyChecks(&module, tc.ID, checks)
		addTierChecks(&module, tier, counts)
		addPublishableChecks(&module, tier, counts)
	}
	return module
}

func loadZiweiGoldFile(path string) (*ziweiGoldFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("load Ziwei full-chart Gold registry: %w", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()
	var data ziweiGoldFile
	if err := decoder.Decode(&data); err != nil {
		return nil, fmt.Errorf("decode Ziwei full-chart Gold registry: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return nil, fmt.Errorf("decode Ziwei full-chart Gold registry trailing content: %w", err)
	}
	return &data, nil
}

func validateZiweiGoldFile(file *ziweiGoldFile) []string {
	reasons := make([]string, 0, 5)
	if !file.Frozen {
		reasons = append(reasons, "dataset is not frozen")
	}
	if file.ProfileID != ziweipkg.DefaultProfileID || file.EngineVersion != ziweipkg.ZiWeiEngineVersion || file.RuleVersion != ziweipkg.ZiWeiRuleVersion {
		reasons = append(reasons, "dataset calculation versions do not match the current Ziwei profile")
	}
	encoded, err := json.Marshal(file.Cases)
	if err != nil {
		reasons = append(reasons, "cases cannot be canonically encoded")
	} else if file.DatasetHash != "sha256:"+sha256HexBytes(encoded) {
		reasons = append(reasons, "dataset_hash does not match canonical cases JSON")
	}
	return reasons
}

func validateZiweiGoldCase(tc ziweiGoldCase) []string {
	reasons := make([]string, 0, 8)
	if strings.TrimSpace(tc.ID) == "" || tc.Year <= 0 || tc.Month <= 0 || tc.Day <= 0 || tc.Hour < 0 || tc.Hour > 23 || tc.Minute < 0 || tc.Minute > 59 || strings.TrimSpace(tc.Gender) == "" {
		reasons = append(reasons, "complete factual birth input is required")
	}
	metadata := tc.Metadata
	if !metadataPublishable(metadata) || metadata == nil || metadata.Purpose != ziweiGoldPurpose {
		reasons = append(reasons, "case metadata must be approved Gold with two reviewers, license, source hash, and ziwei_full_chart_gold purpose")
	} else {
		if strings.TrimSpace(metadata.SourceURL) == "" {
			reasons = append(reasons, "source_url is required")
		}
		if !validSHA256(metadata.SourceHash) {
			reasons = append(reasons, "source_hash must be sha256:<64 lowercase hex>")
		}
		if uniqueNonEmptyStrings(metadata.Reviewers) < 2 {
			reasons = append(reasons, "two distinct non-empty reviewers are required")
		}
	}
	reasons = append(reasons, validateZiweiGoldExpected(tc.Expected)...)
	return reasons
}

func validateZiweiGoldExpected(expected ziweiGoldExpected) []string {
	reasons := make([]string, 0, 12)
	if expected.SoulPalaceBranch == "" || expected.BodyPalaceBranch == "" || expected.BodyPalace == "" || expected.LifeMaster == "" || expected.BodyMaster == "" || expected.FiveBureau == "" {
		reasons = append(reasons, "命宫、身宫、命主、身主和五行局必须完整")
	}
	if !containsString(ziweipkg.BranchNames, expected.SoulPalaceBranch) || !containsString(ziweipkg.BranchNames, expected.BodyPalaceBranch) || expected.BodyPalace != expected.BodyPalaceBranch {
		reasons = append(reasons, "命宫与身宫地支必须合法，且 body_palace 必须与 body_palace_branch 一致")
	}
	if len(expected.Palaces) != 12 {
		reasons = append(reasons, "exactly 12 palaces are required")
	}
	palaceNames, palaceBranches := make(map[string]struct{}), make(map[string]struct{})
	majorCounts := make(map[string]int)
	bodyPalaces, fourHuaCount := 0, 0
	fourHuaSuffixes := make(map[string]int, 4)
	for _, palace := range expected.Palaces {
		if !containsString(ziweipkg.ZIWEI_PALACE_NAMES, palace.Name) || !containsString(ziweipkg.BranchNames, palace.Branch) || !containsString(ziweipkg.StemNames, palace.HeavenlyStem) || palace.MainStars == nil || palace.AuxStars == nil || palace.FourHua == nil {
			reasons = append(reasons, "each palace requires a canonical name, branch, stem, and explicit star/four-hua arrays")
			continue
		}
		if palace.Changsheng12 == "" || palace.Boshi12 == "" || palace.JiangQian12 == "" || palace.SuiQian12 == "" {
			reasons = append(reasons, "each palace requires all four twelve-stage cycle labels")
		}
		palaceNames[palace.Name] = struct{}{}
		palaceBranches[palace.Branch] = struct{}{}
		if palace.IsBodyPalace {
			bodyPalaces++
			if palace.Branch != expected.BodyPalaceBranch {
				reasons = append(reasons, "the body-palace marker must be on body_palace_branch")
			}
		}
		for _, star := range palace.MainStars {
			majorCounts[star]++
		}
		for _, hua := range palace.FourHua {
			fourHuaCount++
			for _, suffix := range []string{"化禄", "化权", "化科", "化忌"} {
				if strings.HasSuffix(hua, suffix) {
					fourHuaSuffixes[suffix]++
				}
			}
		}
	}
	if !sameStringSet(palaceNames, ziweipkg.ZIWEI_PALACE_NAMES) || !sameStringSet(palaceBranches, ziweipkg.BranchNames) {
		reasons = append(reasons, "palaces must contain each canonical palace name and earthly branch exactly once")
	}
	if bodyPalaces != 1 {
		reasons = append(reasons, "exactly one palace must carry the body-palace marker")
	}
	for _, star := range ziweiMajorStars {
		if majorCounts[star] != 1 {
			reasons = append(reasons, fmt.Sprintf("major star %s must appear exactly once", star))
		}
	}
	if len(majorCounts) != len(ziweiMajorStars) {
		reasons = append(reasons, "unknown or duplicate major-star labels are present")
	}
	if fourHuaCount != 4 || fourHuaSuffixes["化禄"] != 1 || fourHuaSuffixes["化权"] != 1 || fourHuaSuffixes["化科"] != 1 || fourHuaSuffixes["化忌"] != 1 {
		reasons = append(reasons, "四化全盘必须恰有化禄、化权、化科、化忌各一项")
	}
	reasons = append(reasons, validateZiweiGoldDayun(expected.Dayun)...)
	return uniqueStrings(reasons)
}

func validateZiweiGoldDayun(dayun []ziweipkg.DayunStage) []string {
	if len(dayun) != 12 {
		return []string{"大限全表必须恰有十二个十年阶段"}
	}
	reasons := make([]string, 0, 3)
	palaces := make(map[string]struct{}, 12)
	for i, stage := range dayun {
		if !containsString(ziweipkg.ZIWEI_PALACE_NAMES, stage.Palace) || stage.Stars == nil {
			reasons = append(reasons, "each Dayun stage requires a canonical palace and an explicit stars array")
		}
		palaces[stage.Palace] = struct{}{}
		if stage.StartAge <= 0 || stage.EndAge != stage.StartAge+9 || i > 0 && stage.StartAge != dayun[i-1].StartAge+10 {
			reasons = append(reasons, "Dayun stages must be contiguous ten-year ranges")
		}
	}
	if !sameStringSet(palaces, ziweipkg.ZIWEI_PALACE_NAMES) {
		reasons = append(reasons, "Dayun must cover each canonical palace exactly once")
	}
	return uniqueStrings(reasons)
}

func ziweiGoldChecks(chart *ziweipkg.ZiWeiChart, dayun ziweipkg.Dayun, expected ziweiGoldExpected) []fieldCheck {
	checks := []fieldCheck{
		{ruleID: checkZiweiGoldScalar, field: "soul_palace_branch", required: true, want: expected.SoulPalaceBranch, got: chart.EarthlyBranchOfSoulPalace},
		{ruleID: checkZiweiGoldScalar, field: "body_palace_branch", required: true, want: expected.BodyPalaceBranch, got: chart.EarthlyBranchOfBodyPalace},
		{ruleID: checkZiweiGoldScalar, field: "body_palace", required: true, want: expected.BodyPalace, got: chart.BodyPalace},
		{ruleID: checkZiweiGoldScalar, field: "life_master", required: true, want: expected.LifeMaster, got: chart.LifeMaster},
		{ruleID: checkZiweiGoldScalar, field: "body_master", required: true, want: expected.BodyMaster, got: chart.BodyMaster},
		{ruleID: checkZiweiGoldScalar, field: "five_bureau", required: true, want: expected.FiveBureau, got: chart.FiveBureau},
	}
	actualByName := make(map[string]ziweipkg.PalaceInfo, len(chart.Palaces))
	for _, palace := range chart.Palaces {
		actualByName[palace.Name] = palace
	}
	for _, expectedPalace := range expected.Palaces {
		actual := actualByName[expectedPalace.Name]
		actualMainStars, actualAuxStars := ziweiGoldStarGroups(actual)
		prefix := "palaces." + expectedPalace.Name + "."
		checks = append(checks,
			fieldCheck{ruleID: checkZiweiGoldPalaceScalar, field: prefix + "branch", required: true, want: expectedPalace.Branch, got: actual.Branch},
			fieldCheck{ruleID: checkZiweiGoldPalaceScalar, field: prefix + "heavenly_stem", required: true, want: expectedPalace.HeavenlyStem, got: actual.HeavenlyStem},
			fieldCheck{ruleID: checkZiweiGoldPalaceScalar, field: prefix + "is_body_palace", required: true, want: fmt.Sprint(expectedPalace.IsBodyPalace), got: fmt.Sprint(actual.IsBodyPalace)},
			fieldCheck{ruleID: checkZiweiGoldPalaceSet, field: prefix + "main_stars", required: true, wantSet: expectedPalace.MainStars, gotSet: actualMainStars},
			fieldCheck{ruleID: checkZiweiGoldPalaceSet, field: prefix + "aux_stars", required: true, wantSet: expectedPalace.AuxStars, gotSet: actualAuxStars},
			fieldCheck{ruleID: checkZiweiGoldPalaceSet, field: prefix + "four_hua", required: true, wantSet: expectedPalace.FourHua, gotSet: actual.FourHua},
			fieldCheck{ruleID: checkZiweiGoldPalaceScalar, field: prefix + "changsheng_12", required: true, want: expectedPalace.Changsheng12, got: actual.Changsheng12},
			fieldCheck{ruleID: checkZiweiGoldPalaceScalar, field: prefix + "boshi_12", required: true, want: expectedPalace.Boshi12, got: actual.Boshi12},
			fieldCheck{ruleID: checkZiweiGoldPalaceScalar, field: prefix + "jiang_qian_12", required: true, want: expectedPalace.JiangQian12, got: actual.JiangQian12},
			fieldCheck{ruleID: checkZiweiGoldPalaceScalar, field: prefix + "sui_qian_12", required: true, want: expectedPalace.SuiQian12, got: actual.SuiQian12},
		)
	}
	checks = append(checks, fieldCheck{ruleID: checkZiweiGoldDayun, field: "dayun", required: true, want: canonicalDayun(expected.Dayun), got: canonicalDayun(dayun)})
	return checks
}

func ziweiGoldStarGroups(palace ziweipkg.PalaceInfo) (mainStars, auxStars []string) {
	for _, star := range palace.Stars {
		if star.Type == "major" {
			mainStars = append(mainStars, star.Name)
			continue
		}
		auxStars = append(auxStars, star.Name)
	}
	return mainStars, auxStars
}

func canonicalDayun(value []ziweipkg.DayunStage) string {
	cloned := append([]ziweipkg.DayunStage(nil), value...)
	for i := range cloned {
		cloned[i].Stars = sortedStrings(cloned[i].Stars)
		cloned[i].LiuNianStars = sortedStrings(cloned[i].LiuNianStars)
		cloned[i].LiuYueStars = sortedStrings(cloned[i].LiuYueStars)
	}
	encoded, _ := json.Marshal(cloned)
	return string(encoded)
}

func sortedStrings(values []string) []string {
	cloned := append([]string{}, values...)
	sort.Strings(cloned)
	return cloned
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sameStringSet(actual map[string]struct{}, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	for _, value := range expected {
		if _, ok := actual[value]; !ok {
			return false
		}
	}
	return true
}

func validSHA256(value string) bool {
	if !strings.HasPrefix(value, "sha256:") || len(value) != len("sha256:")+64 {
		return false
	}
	_, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && value == strings.ToLower(value)
}

func sha256HexBytes(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}

func uniqueNonEmptyStrings(values []string) int {
	set := make(map[string]struct{})
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			set[value] = struct{}{}
		}
	}
	return len(set)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func duplicateZiweiGoldIDs(cases []ziweiGoldCase) []string {
	counts := make(map[string]int)
	for _, tc := range cases {
		counts[tc.ID]++
	}
	duplicates := make([]string, 0)
	for id, count := range counts {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}
	sort.Strings(duplicates)
	return duplicates
}

var ziweiMajorStars = []string{
	"紫微", "天机", "太阳", "武曲", "天同", "廉贞", "天府",
	"太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "破军",
}
