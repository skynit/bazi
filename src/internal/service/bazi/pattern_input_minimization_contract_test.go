package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternCandidatesDependOnlyOnPillarsAndMonthBranch(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "辰"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	if !hasPatternRuleID(analysis.Candidates, "pattern.lu.jianlu") {
		t.Fatalf("fixture did not exercise a structural candidate: %+v", analysis)
	}
	if !ValidPatternAnalysis(analysis, pillars, "寅") {
		t.Fatal("persisted validation rejected authoritative pillar/month inputs")
	}
}

func TestPatternInputSnapshotContainsOnlyAuthoritativeInputs(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "辰"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	want := PatternInputSnapshot{Pillars: []string{"丙子", "丙寅", "甲辰", "戊辰"}, MonthBranch: "寅"}
	if !reflect.DeepEqual(analysis.Inputs, want) {
		t.Fatalf("pattern inputs = %+v, want %+v", analysis.Inputs, want)
	}
	payload, err := json.Marshal(analysis.Inputs)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"element_scores", "body_strength_score_band_candidate"} {
		if strings.Contains(string(payload), forbidden) {
			t.Errorf("pattern input snapshot still exposes %q: %s", forbidden, payload)
		}
	}
}

func TestPatternCandidateCoreHasNoScoringGate(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go", "bazi.go", "calculate_from_pillars.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{
			"ElementScores", "BodyStrengthScoreBandCandidate", "len(scores)",
			"scoreBandCandidate", "five-element scores", "body-strength score band",
		} {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("pattern production path %s still contains unconsumed scoring dependency %q", path, forbidden)
			}
		}
	}
}

func TestPatternInputMinimizationManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧 pattern-candidate-set-v3",
			"五行分数和身强分段",
			"10个存量检测器均不消费",
			"移除分数与分段门禁及快照字段",
			"pattern-candidate-set-v4",
			"权威输入只保留四柱与月支",
			"函数签名与持久化验证只接收四柱和月支",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
