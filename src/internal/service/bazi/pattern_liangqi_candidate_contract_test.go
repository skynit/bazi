package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestLiangQiChengXiangRequiresExactlyFourByFourVisibleMainQi(t *testing.T) {
	tests := []struct {
		name    string
		pillars []model.Pillar
	}{
		{
			name: "wood generates fire classical structure",
			pillars: []model.Pillar{
				{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
				{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
			},
		},
		{
			name: "earth controls water classical structure",
			pillars: []model.Pillar{
				{Gan: "癸", Zhi: "亥"}, {Gan: "己", Zhi: "未"},
				{Gan: "癸", Zhi: "亥"}, {Gan: "己", Zhi: "未"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := checkLiangQiChengXiang(tc.pillars)
			if got == nil || got.PatternName != "两气成象格" {
				t.Fatalf("two-qi structure = %+v", got)
			}
		})
	}
}

func TestLiangQiChengXiangRejectsThirdQiAndUnequalSplit(t *testing.T) {
	thirdQi := []model.Pillar{
		{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
		{Gan: "甲", Zhi: "午"}, {Gan: "戊", Zhi: "辰"},
	}
	unequal := []model.Pillar{
		{Gan: "甲", Zhi: "寅"}, {Gan: "甲", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "丙", Zhi: "午"},
	}
	if got := checkLiangQiChengXiang(thirdQi); got != nil {
		t.Fatalf("three visible qi matched two-qi structure: %+v", got)
	}
	if got := checkLiangQiChengXiang(unequal); got != nil {
		t.Fatalf("6:2 visible split matched two-qi structure: %+v", got)
	}
}

func TestLiangQiFormalCandidateUsesClassicalProfile(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
		{Gan: "甲", Zhi: "午"}, {Gan: "丁", Zhi: "卯"},
	}
	analysis := AnalyzePatternExtended(pillars, "卯")
	candidate, ok := luPatternCandidateByID(analysis.Candidates, "pattern.special.liangqi")
	if !ok || candidate.PatternName != "两气成象格" || !strings.Contains(candidate.Source, "《滴天髓阐微》PDF第43页") {
		t.Fatalf("two-qi formal candidate = %+v", candidate)
	}
	if hasLuPatternRuleID(analysis.Candidates, "pattern.special.liangshen") {
		t.Fatalf("former two-spirit candidate survived: %+v", analysis.Candidates)
	}
}

func TestFormerLiangShenThresholdPathIsAbsent(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go", "wuxing.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{
			"checkLiangShenChengXiang", "pattern.special.liangshen", "两神成像格",
			"majorElems", "tongGuan",
		} {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("production source %s still contains former two-spirit path %q", path, forbidden)
			}
		}
	}
}

func TestLiangQiManifestRecordsClassicalVisibleMainQiProfile(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"《滴天髓阐微》PDF第43页",
			"两气双清",
			"四干与四支本气八个位点",
			"恰好两种五行且各占四位",
			"删除旧15%聚合分数阈值",
			"藏干杂气、月令旺衰、顺逆取用和行运破局仍未裁决",
		} {
			if !strings.Contains(table.Description+table.Source, fragment) {
				t.Errorf("pattern description/source missing %q: %+v", fragment, table)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}

func liangQiContractScores() map[string]int {
	return map[string]int{"木": 43, "火": 37, "土": 7, "金": 7, "水": 6}
}
