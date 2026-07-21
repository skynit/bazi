package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestZhuanLuExactFourDaysAcrossSixtyCycle(t *testing.T) {
	wants := map[string]bool{"甲寅": true, "乙卯": true, "庚申": true, "辛酉": true}
	matched := 0
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		gan := data.Gans[dayIndex%10]
		zhi := data.Zhis[dayIndex%12]
		got := checkZhuanLuGe(gan, zhi)
		want := wants[gan+zhi]
		if (got != nil) != want {
			t.Errorf("day %s%s zhuan-lu = %+v, want %v", gan, zhi, got, want)
		}
		if got != nil {
			matched++
			if got.PatternName != "专禄格" {
				t.Errorf("day %s%s zhuan-lu metadata = %+v", gan, zhi, got)
			}
		}
	}
	if matched != 4 {
		t.Fatalf("zhuan-lu matched %d days, want 4", matched)
	}
}

func TestZhuanLuAndJianLuRemainIndependentCandidates(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	for _, ruleID := range []string{"pattern.lu.jianlu", "pattern.lu.zhuanlu"} {
		if !hasLuPatternRuleID(analysis.Candidates, ruleID) {
			t.Errorf("candidate %s missing when 建禄 and 专禄 coexist: %+v", ruleID, analysis.Candidates)
		}
	}
	if candidate, ok := luPatternCandidateByID(analysis.Candidates, "pattern.lu.zhuanlu"); !ok ||
		!strings.Contains(candidate.Source, "《三命通会》PDF第190页") {
		t.Errorf("专禄 source = %+v", candidate)
	}
}

func TestIncompleteRiLuGuiShiDetectorFailsClosed(t *testing.T) {
	source, err := os.ReadFile("pattern.go")
	if err != nil {
		t.Fatal(err)
	}
	collector, err := os.ReadFile("pattern_candidates.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"checkRiLuGuiShiGe", "日禄归时格", "pattern.lu.riluguishi"} {
		if strings.Contains(string(source), forbidden) || strings.Contains(string(collector), forbidden) {
			t.Errorf("production pattern source still contains incomplete path %q", forbidden)
		}
	}

	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "戊", Zhi: "辰"},
		{Gan: "甲", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
	}
	analysis := AnalyzePatternExtended(pillars, "辰")
	if hasLuPatternRuleID(analysis.Candidates, "pattern.lu.riluguishi") {
		t.Fatalf("incomplete 日禄归时 candidate survived formal entry: %+v", analysis.Candidates)
	}
}

func TestLuPatternManifestRecordsRetirementAndDetectorCount(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{"专禄固定甲寅、乙卯、庚申、辛酉", "与月令建禄独立并可同时命中", "六忌的大部分条件", "现失败关闭"} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}

func luCandidateBalancedScores() map[string]int {
	return map[string]int{"木": 20, "火": 20, "土": 20, "金": 20, "水": 20}
}

func hasLuPatternRuleID(candidates []PatternCandidate, ruleID string) bool {
	_, ok := luPatternCandidateByID(candidates, ruleID)
	return ok
}

func luPatternCandidateByID(candidates []PatternCandidate, ruleID string) (PatternCandidate, bool) {
	for _, candidate := range candidates {
		if candidate.RuleID == ruleID {
			return candidate, true
		}
	}
	return PatternCandidate{}, false
}
