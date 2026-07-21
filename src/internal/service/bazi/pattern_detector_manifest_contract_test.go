package bazi

import (
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
)

const expectedPatternDetectorManifestSHA256 = "6334f79633183924f9daf4d1a695bd84281b1bb3126e853657a436068fff57d8"

func TestPatternDetectorManifestHashIsCanonicalAndMutationSensitive(t *testing.T) {
	detectors := patternDetectorRegistry()
	want := expectedPatternDetectorManifestSHA256
	if got := patternDetectorManifestSHA256(detectors); got != want {
		t.Fatalf("pattern detector manifest SHA-256 = %s, want %s", got, want)
	}

	reversed := append([]patternDetectorDefinition(nil), detectors...)
	for left, right := 0, len(reversed)-1; left < right; left, right = left+1, right-1 {
		reversed[left], reversed[right] = reversed[right], reversed[left]
	}
	if got := patternDetectorManifestSHA256(reversed); got != want {
		t.Fatalf("registry order changed canonical manifest SHA-256 to %s", got)
	}

	mutations := []func(*patternDetectorDefinition){
		func(detector *patternDetectorDefinition) { detector.ruleID += ".mutated" },
		func(detector *patternDetectorDefinition) { detector.source += "（变更）" },
		func(detector *patternDetectorDefinition) { detector.category = "变更" },
		func(detector *patternDetectorDefinition) { detector.algorithmSHA256 = strings.Repeat("1", 64) },
		func(detector *patternDetectorDefinition) { detector.behaviorSHA256 = strings.Repeat("2", 64) },
		func(detector *patternDetectorDefinition) { detector.outputNames[0] += "变更" },
		func(detector *patternDetectorDefinition) { detector.profileSHA256 = strings.Repeat("0", 64) },
	}
	for index, mutate := range mutations {
		mutated := patternDetectorRegistry()
		mutate(&mutated[0])
		if got := patternDetectorManifestSHA256(mutated); got == want {
			t.Errorf("manifest mutation %d did not change SHA-256", index)
		}
	}
}

func TestPatternDetectorManifestBindsEveryAlgorithm(t *testing.T) {
	tests := []struct {
		ruleID, rootFunction, category, patternName, month string
		pillars                                            []model.Pillar
	}{
		{ruleID: "pattern.special.zhuanwang", rootFunction: "checkZhuanWangGe", category: patternCategoryStructural, patternName: "曲直格", month: "卯", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"}, {Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"},
		}},
		{ruleID: "pattern.lu.jianlu", rootFunction: "checkJianLuGe", category: patternCategoryStructural, patternName: "建禄格", month: "寅", pillars: []model.Pillar{
			{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"}, {Gan: "甲", Zhi: "辰"}, {Gan: "戊", Zhi: "辰"},
		}},
		{ruleID: "pattern.lu.yueren", rootFunction: "checkYueRenGe", category: patternCategoryStructural, patternName: "月刃格", month: "午", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "午"}, {Gan: "丙", Zhi: "辰"}, {Gan: "庚", Zhi: "辰"},
		}},
		{ruleID: "pattern.lu.zhuanlu", rootFunction: "checkZhuanLuGe", category: patternCategoryStructural, patternName: "专禄格", month: "辰", pillars: []model.Pillar{
			{Gan: "丙", Zhi: "子"}, {Gan: "戊", Zhi: "辰"}, {Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
		}},
		{ruleID: "pattern.lu.riren", rootFunction: "checkRiRenGe", category: patternCategoryStructural, patternName: "日刃格", month: "辰", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "辰"}, {Gan: "丙", Zhi: "午"}, {Gan: "庚", Zhi: "辰"},
		}},
		{ruleID: "pattern.special.liangqi", rootFunction: "checkLiangQiChengXiang", category: patternCategoryStructural, patternName: "两气成象格", month: "卯", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"}, {Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
		}},
		{ruleID: "pattern.aux.kuigang", rootFunction: "checkKuiGangGe", category: patternCategoryAuxiliary, patternName: "魁罡格", month: "寅", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "子"}, {Gan: "丙", Zhi: "寅"}, {Gan: "庚", Zhi: "辰"}, {Gan: "庚", Zhi: "午"},
		}},
		{ruleID: "pattern.aux.jinshen", rootFunction: "checkJinShenHour", category: patternCategoryAuxiliary, patternName: "金神", month: "寅", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"}, {Gan: "甲", Zhi: "子"}, {Gan: "癸", Zhi: "酉"},
		}},
		{ruleID: "pattern.aux.sanqi", rootFunction: "checkSanQi", category: patternCategoryAuxiliary, patternName: "三奇", month: "辰", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "辰"}, {Gan: "庚", Zhi: "午"}, {Gan: "甲", Zhi: "戌"},
		}},
		{ruleID: "pattern.aux.ride", rootFunction: "checkRiDeGe", category: patternCategoryAuxiliary, patternName: "日德格", month: "辰", pillars: []model.Pillar{
			{Gan: "丙", Zhi: "子"}, {Gan: "戊", Zhi: "辰"}, {Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"},
		}},
	}

	detectors := make(map[string]patternDetectorDefinition, patternDetectorCount())
	for _, detector := range patternDetectorRegistry() {
		detectors[detector.ruleID] = detector
	}
	for _, tc := range tests {
		t.Run(tc.ruleID, func(t *testing.T) {
			detector, ok := detectors[tc.ruleID]
			if !ok {
				t.Fatalf("detector %s missing from registry", tc.ruleID)
			}
			algorithm, algorithmOK := patternDetectorAlgorithmProfileForRule(tc.ruleID)
			if !algorithmOK || algorithm.RootFunction != tc.rootFunction || detector.algorithmSHA256 != algorithm.ASTSHA256 ||
				detector.behaviorSHA256 != patternDetectorBehaviorSHA256(tc.ruleID) ||
				detector.category != tc.category || detector.profileSHA256 != patternDetectorProfileSHA256(tc.ruleID) ||
				!patternStringProfileContains(detector.outputNames, tc.patternName) {
				t.Fatalf("detector %s metadata = %s/%s/%s", tc.ruleID, detector.algorithmSHA256, detector.category, detector.profileSHA256)
			}
			got := detector.detect(patternDetectorContext{
				pillars: tc.pillars, monthBranch: tc.month,
				dayGan: tc.pillars[2].Gan, dayZhi: tc.pillars[2].Zhi,
			})
			if got == nil || got.PatternName != tc.patternName {
				t.Fatalf("detector %s returned %+v, want %s", tc.ruleID, got, tc.patternName)
			}
			if !validPatternDetectorOutput(detector, got) {
				t.Fatalf("detector %s rejected registered output %+v", tc.ruleID, got)
			}
			tampered := *got
			tampered.PatternName += "变更"
			if validPatternDetectorOutput(detector, &tampered) {
				t.Fatalf("detector %s accepted unregistered output %+v", tc.ruleID, tampered)
			}
		})
	}
}

func TestPatternAnalysisPublishesAndValidatesDetectorManifestHash(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	want := expectedPatternDetectorManifestSHA256
	valid := AnalyzePatternExtended(pillars, "寅")
	invalid := AnalyzePatternExtended(pillars[:3], "寅")
	for _, analysis := range []PatternAnalysis{valid, invalid} {
		if analysis.DetectorManifestSHA256 != want || len(analysis.DetectorManifestSHA256) != 64 {
			t.Errorf("pattern manifest SHA-256 = %q, want %q", analysis.DetectorManifestSHA256, want)
		}
	}

	tampered := valid
	tampered.DetectorManifestSHA256 = strings.Repeat("0", 64)
	if reflect.DeepEqual(tampered, valid) || ValidPatternAnalysis(tampered, pillars, "寅") {
		t.Fatal("tampered detector manifest SHA-256 passed strict validation")
	}
}

func TestPatternDetectorManifestMetadataContract(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧detector_profile只是人工版本标签",
			"规则身份、来源、分类和实现映射的机器摘要",
			"pattern-candidate-set-v15新增detector_manifest_sha256",
			"按rule_id规范排序",
			"合法与非法结果都从单次执行快照派生同一摘要",
			"持久化重算逐字段验证",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
