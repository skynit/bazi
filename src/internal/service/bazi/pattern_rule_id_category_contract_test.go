package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternRuleIDPrefixesMatchEveryRegisteredCategory(t *testing.T) {
	tests := []struct {
		name, ruleID, category, month string
		pillars                       []model.Pillar
	}{
		{name: "zhuan-wang", ruleID: "pattern.special.zhuanwang", category: patternCategoryStructural, month: "卯", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"}, {Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"},
		}},
		{name: "jian-lu", ruleID: "pattern.lu.jianlu", category: patternCategoryStructural, month: "寅", pillars: []model.Pillar{
			{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"}, {Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
		}},
		{name: "yue-ren", ruleID: "pattern.lu.yueren", category: patternCategoryStructural, month: "午", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "午"}, {Gan: "丙", Zhi: "午"}, {Gan: "庚", Zhi: "辰"},
		}},
		{name: "zhuan-lu", ruleID: "pattern.lu.zhuanlu", category: patternCategoryStructural, month: "寅", pillars: []model.Pillar{
			{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"}, {Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
		}},
		{name: "ri-ren", ruleID: "pattern.lu.riren", category: patternCategoryStructural, month: "午", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "午"}, {Gan: "丙", Zhi: "午"}, {Gan: "庚", Zhi: "辰"},
		}},
		{name: "liang-qi", ruleID: "pattern.special.liangqi", category: patternCategoryStructural, month: "卯", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"}, {Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
		}},
		{name: "kui-gang", ruleID: "pattern.aux.kuigang", category: patternCategoryAuxiliary, month: "寅", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "子"}, {Gan: "丙", Zhi: "寅"}, {Gan: "庚", Zhi: "辰"}, {Gan: "庚", Zhi: "午"},
		}},
		{name: "jin-shen", ruleID: "pattern.aux.jinshen", category: patternCategoryAuxiliary, month: "寅", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"}, {Gan: "甲", Zhi: "子"}, {Gan: "癸", Zhi: "酉"},
		}},
		{name: "san-qi", ruleID: "pattern.aux.sanqi", category: patternCategoryAuxiliary, month: "辰", pillars: []model.Pillar{
			{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "辰"}, {Gan: "庚", Zhi: "午"}, {Gan: "甲", Zhi: "戌"},
		}},
		{name: "ri-de", ruleID: "pattern.aux.ride", category: patternCategoryAuxiliary, month: "辰", pillars: []model.Pillar{
			{Gan: "丙", Zhi: "子"}, {Gan: "戊", Zhi: "辰"}, {Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"},
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			analysis := AnalyzePatternExtended(tc.pillars, tc.month)
			candidate, ok := luPatternCandidateByID(analysis.Candidates, tc.ruleID)
			if !ok || candidate.Category != tc.category {
				t.Fatalf("candidate %s = %+v, want category %s", tc.ruleID, candidate, tc.category)
			}
			if tc.category == patternCategoryAuxiliary && !strings.HasPrefix(candidate.RuleID, "pattern.aux.") {
				t.Errorf("auxiliary candidate has non-auxiliary rule ID: %+v", candidate)
			}
			if tc.category == patternCategoryStructural &&
				!strings.HasPrefix(candidate.RuleID, "pattern.lu.") &&
				!strings.HasPrefix(candidate.RuleID, "pattern.special.") {
				t.Errorf("structural candidate has non-structural rule ID: %+v", candidate)
			}
		})
	}
}

func TestFormerJinShenStructuralRuleIDIsAbsent(t *testing.T) {
	for _, path := range []string{"pattern_candidates.go", "shensha_sanqi_jinshen_contract_test.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(source), "pattern.special.jinshen") {
			t.Errorf("%s still contains former structural JinShen rule ID", path)
		}
	}
}

func TestPatternRuleIDCategoryManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧金神规则ID pattern.special.jinshen",
			"category却是辅助特征",
			"更名为pattern.aux.jinshen",
			"10个注册规则ID前缀与category一致",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
