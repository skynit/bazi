package bazi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternCandidateSchemaOmitsUnreachableDisputeFields(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	if len(analysis.Candidates) < 2 {
		t.Fatalf("fixture must exercise overlapping non-exclusive candidates: %+v", analysis)
	}
	payload, err := json.Marshal(analysis)
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(payload, &top); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"has_dispute", "dispute_reasons"} {
		if _, ok := top[forbidden]; ok {
			t.Errorf("pattern analysis still exposes unreachable field %q: %s", forbidden, payload)
		}
	}
	var candidates []map[string]json.RawMessage
	if err := json.Unmarshal(top["candidates"], &candidates); err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		for _, forbidden := range []string{"status", "dispute_reason"} {
			if _, ok := candidate[forbidden]; ok {
				t.Errorf("pattern candidate still exposes unreachable field %q: %s", forbidden, payload)
			}
		}
	}
}

func TestPatternDisputeStateMachineIsAbsentFromProduction(t *testing.T) {
	for _, path := range []string{"pattern.go", "pattern_candidates.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{
			"exclusiveGroup", "patternStatusDisputed", "patternStatusCandidate",
			"patternCategoryCompound", "HasDispute", "DisputeReason",
		} {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("pattern production source %s still contains dead dispute path %q", path, forbidden)
			}
		}
	}
}

func TestPatternDisputeConsumersAreRemoved(t *testing.T) {
	checks := map[string][]string{
		"../interpretation/bazi.go": {
			"p.HasDispute", "p.DisputeReasons", "candidate.DisputeReason", "candidate.Status",
			"Inputs.ElementScores", "Inputs.BodyStrengthScoreBandCandidate",
		},
		"../../../../vue/src/api/chart.ts": {
			"has_dispute", "dispute_reason?:", "element_scores", "body_strength_score_band_candidate",
		},
		"../../../../vue/src/components/BaziChart.vue": {
			"candidate.dispute_reason", "pattern_analysis.has_dispute", "patternStatusLabel", "pattern-dispute",
		},
	}
	for path, forbiddenValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range forbiddenValues {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("consumer %s still contains retired pattern field %q", path, forbidden)
			}
		}
	}
}

func TestPatternRetirementPreservesRelationGraphDisputes(t *testing.T) {
	relations := buildGanRelationGraph([]ganRelationPillar{
		{key: "year", label: labelYear, stem: "甲", branch: "子"},
		{key: "month", label: labelMonth, stem: "己", branch: "辰"},
		{key: "day", label: labelDay, stem: "甲", branch: "午"},
		{key: "hour", label: labelHour, stem: "丙", branch: "酉"},
	})
	count := 0
	for _, relation := range relations {
		if relation.Type != "五合" {
			continue
		}
		count++
		if relation.Status != "disputed" || len(relation.DisputeReasons) == 0 || len(relation.ConflictsWith) == 0 {
			t.Errorf("real five-combine dispute was lost: %+v", relation)
		}
	}
	if count != 2 {
		t.Fatalf("competing five-combine count = %d, want 2", count)
	}
}

func TestPatternDisputeRetirementManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧 pattern-candidate-set-v4 保留互斥争议状态机",
			"10个注册器的exclusiveGroup全部为空",
			"has_dispute、dispute_reasons、候选status和dispute_reason",
			"pattern-candidate-set-v5删除死注册参数、汇总算法和四个公开字段",
			"关系图的真实disputed、ConflictsWith和DisputeReasons合同不受影响",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
