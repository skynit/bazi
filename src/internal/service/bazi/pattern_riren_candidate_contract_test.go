package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestRiRenExactThreeDaysAcrossSixtyCycle(t *testing.T) {
	wants := map[string]bool{"丙午": true, "戊午": true, "壬子": true}
	matched := 0
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		gan := data.Gans[dayIndex%10]
		zhi := data.Zhis[dayIndex%12]
		got := checkRiRenGe(gan, zhi)
		want := wants[gan+zhi]
		if (got != nil) != want {
			t.Errorf("day %s%s ri-ren = %+v, want %v", gan, zhi, got, want)
		}
		if got != nil {
			matched++
			if got.PatternName != "日刃格" {
				t.Errorf("day %s%s ri-ren metadata = %+v", gan, zhi, got)
			}
		}
	}
	if matched != 3 {
		t.Fatalf("ri-ren matched %d days, want 3", matched)
	}
}

func TestRiRenAndMonthBladeRemainIndependentCandidates(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "甲", Zhi: "子"}, {Gan: "戊", Zhi: "午"},
		{Gan: "丙", Zhi: "午"}, {Gan: "庚", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "午")
	for _, ruleID := range []string{"pattern.lu.yueren", "pattern.lu.riren"} {
		if !hasRiRenPatternRuleID(analysis.Candidates, ruleID) {
			t.Errorf("candidate %s missing when 月刃 and 日刃 coexist: %+v", ruleID, analysis.Candidates)
		}
	}
	candidate, ok := riRenPatternCandidateByID(analysis.Candidates, "pattern.lu.riren")
	if !ok || candidate.PatternName != "日刃格" || !strings.Contains(candidate.Source, "《渊海子平》PDF第217页") {
		t.Errorf("日刃 candidate = %+v", candidate)
	}
}

func TestFormerYangRenPatternShortcutIsAbsent(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{"checkYangRenGe", "pattern.lu.yangren", "羊刃格"} {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("production pattern source %s still contains former shortcut %q", path, forbidden)
			}
		}
	}
}

func TestRiRenPatternManifestRecordsProfile(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{"丙午、戊午、壬子三个日刃格结构", "日刃与月刃独立并可同时命中", "刑冲破害、会合和官杀制化只作为未裁决条件"} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}

func riRenPatternScores() map[string]int {
	return map[string]int{"木": 20, "火": 20, "土": 20, "金": 20, "水": 20}
}

func hasRiRenPatternRuleID(candidates []PatternCandidate, ruleID string) bool {
	_, ok := riRenPatternCandidateByID(candidates, ruleID)
	return ok
}

func riRenPatternCandidateByID(candidates []PatternCandidate, ruleID string) (PatternCandidate, bool) {
	for _, candidate := range candidates {
		if candidate.RuleID == ruleID {
			return candidate, true
		}
	}
	return PatternCandidate{}, false
}
