package bazi

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestZhuanWangProfilesMatchClassicalDirectionAndCombinationTables(t *testing.T) {
	wants := map[string]zhuanWangProfile{
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
	if got := zhuanWangProfileRegistry(); !reflect.DeepEqual(got, wants) {
		t.Fatalf("zhuan-wang profiles = %+v, want %+v", got, wants)
	}
}

func TestZhuanWangProfileRegistryReturnsIndependentNestedValues(t *testing.T) {
	mutated := zhuanWangProfileRegistry()
	mutated["木"] = zhuanWangProfile{name: "变更", breakingElement: "水", structures: []zhuanWangStructure{{branches: []string{"子"}}}}
	metal := mutated["金"]
	metal.structures[0].branches[0] = "子"
	mutated["金"] = metal

	fresh := zhuanWangProfileRegistry()
	if fresh["木"].name != "曲直格" || fresh["金"].structures[0].branches[0] != "申" {
		t.Fatalf("fresh zhuan-wang profile inherited mutation: %+v", fresh)
	}
	pillars := []model.Pillar{
		{Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"},
		{Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"},
	}
	if got := checkZhuanWangGe(pillars); got == nil || got.PatternName != "曲直格" {
		t.Fatalf("zhuan-wang detection inherited profile mutation: %+v", got)
	}
}

func TestZhuanWangDirectProfilesMatchFiveStructures(t *testing.T) {
	tests := []struct {
		name, patternName string
		pillars           []model.Pillar
	}{
		{name: "QuZhi", patternName: "曲直格", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"}, {Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"},
		}},
		{name: "YanShang", patternName: "炎上格", pillars: []model.Pillar{
			{Gan: "丙", Zhi: "寅"}, {Gan: "甲", Zhi: "午"}, {Gan: "丙", Zhi: "戌"}, {Gan: "乙", Zhi: "未"},
		}},
		{name: "JiaSe", patternName: "稼穑格", pillars: []model.Pillar{
			{Gan: "戊", Zhi: "辰"}, {Gan: "己", Zhi: "丑"}, {Gan: "戊", Zhi: "戌"}, {Gan: "己", Zhi: "未"},
		}},
		{name: "CongGe", patternName: "从革格", pillars: []model.Pillar{
			{Gan: "庚", Zhi: "申"}, {Gan: "乙", Zhi: "酉"}, {Gan: "庚", Zhi: "戌"}, {Gan: "庚", Zhi: "辰"},
		}},
		{name: "RunXia", patternName: "润下格", pillars: []model.Pillar{
			{Gan: "壬", Zhi: "子"}, {Gan: "辛", Zhi: "亥"}, {Gan: "癸", Zhi: "丑"}, {Gan: "壬", Zhi: "子"},
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checkZhuanWangGe(tc.pillars)
			if got == nil || got.PatternName != tc.patternName {
				t.Fatalf("direct zhuan-wang structure = %+v", got)
			}
		})
	}
}

func TestZhuanWangRejectsMissingTrioAndExternalBreakingBranch(t *testing.T) {
	missingTrio := []model.Pillar{
		{Gan: "甲", Zhi: "寅"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
	}
	externalMetal := []model.Pillar{
		{Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"},
		{Gan: "甲", Zhi: "辰"}, {Gan: "乙", Zhi: "酉"},
	}
	if got := checkZhuanWangGe(missingTrio); got != nil {
		t.Fatalf("score-dominant chart without complete direction/combination matched: %+v", got)
	}
	if got := checkZhuanWangGe(externalMetal); got != nil {
		t.Fatalf("complete eastern direction with external metal branch matched: %+v", got)
	}
}

func TestZhuanWangFormalCandidateIgnoresAggregateScoreThresholds(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"},
		{Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"},
	}
	for _, scores := range []map[string]int{
		{"木": 70, "火": 15, "土": 5, "金": 0, "水": 10},
		{"木": 5, "火": 5, "土": 30, "金": 30, "水": 30},
	} {
		analysis := AnalyzePatternExtended(pillars, "卯")
		candidate, ok := luPatternCandidateByID(analysis.Candidates, "pattern.special.zhuanwang")
		if !ok || candidate.PatternName != "曲直格" || !strings.Contains(candidate.Source, "《滴天髓阐微》PDF第44-45页") {
			t.Fatalf("formal zhuan-wang candidate = %+v for scores %+v", candidate, scores)
		}
	}
}

func TestFormerCongQiangThresholdPathIsAbsent(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go", "testexport.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{
			"checkCongQiangGe", "ExportCheckCongQiangGe", "稼穑格（从强格）",
			"《三命通会》专旺格规则（本地条件化）",
		} {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("production source %s still contains former cong-qiang path %q", path, forbidden)
			}
		}
	}
}

func TestZhuanWangManifestRecordsClassicalDirectStructureProfile(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"《滴天髓阐微》PDF第44-45页",
			"曲直、炎上、从革、润下要求地支完整成方或三合局",
			"稼穑要求辰戌丑未四库皆全",
			"删除旧60%生扶、30%日主和10%克神分数阈值",
			"转化成局、藏干杂气、得时旺衰、引通取用和行运破局仍未裁决",
		} {
			if !strings.Contains(table.Description+table.Source, fragment) {
				t.Errorf("pattern description/source missing %q: %+v", fragment, table)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
