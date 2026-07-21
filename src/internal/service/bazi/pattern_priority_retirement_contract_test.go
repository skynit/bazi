package bazi

import (
	"encoding/json"
	"os"
	"sort"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternCandidateSetHasNoUnadjudicatedRanking(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	if len(analysis.Candidates) < 2 {
		t.Fatalf("fixture must retain overlapping candidates: %+v", analysis)
	}

	payload, err := json.Marshal(analysis)
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(payload, &top); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"selection_basis", "primary_candidate_id"} {
		if _, ok := top[forbidden]; ok {
			t.Errorf("pattern analysis still exposes unadjudicated ranking field %q: %s", forbidden, payload)
		}
	}

	var candidates []map[string]json.RawMessage
	if err := json.Unmarshal(top["candidates"], &candidates); err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		for _, forbidden := range []string{"role", "priority"} {
			if _, ok := candidate[forbidden]; ok {
				t.Errorf("pattern candidate still exposes unadjudicated ranking field %q: %s", forbidden, payload)
			}
		}
	}

	gotRuleIDs := make([]string, 0, len(analysis.Candidates))
	for _, candidate := range analysis.Candidates {
		gotRuleIDs = append(gotRuleIDs, candidate.RuleID)
	}
	wantRuleIDs := append([]string(nil), gotRuleIDs...)
	sort.Strings(wantRuleIDs)
	for i := range gotRuleIDs {
		if gotRuleIDs[i] != wantRuleIDs[i] {
			t.Fatalf("candidate order = %v, want deterministic rule-id order %v", gotRuleIDs, wantRuleIDs)
		}
	}
}

func TestPatternRankingConsumersAreRemoved(t *testing.T) {
	checks := map[string][]string{
		"pattern.go": {
			"SelectionBasis", "PrimaryCandidateID",
		},
		"pattern_candidates.go": {
			"patternRolePrimary", "patternRoleSecondary", "patternRoleAuxiliary",
			"Priority", "priority_descending_then_rule_id", "PrimaryPatternCandidate",
		},
		"../interpretation/bazi.go": {
			"PrimaryPatternCandidate", "candidate.Role", "candidate.Priority", "p.SelectionBasis",
		},
		"../../../../vue/src/api/chart.ts": {
			"primary_candidate_id", "priority_descending_then_rule_id", "role: '主格'",
		},
		"../../../../vue/src/components/BaziChart.vue": {
			"primaryPatternCandidate", "candidate.role", "candidate.priority", "显示顺序首项",
		},
	}
	for path, forbiddenValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range forbiddenValues {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("consumer %s still contains retired ranking path %q", path, forbidden)
			}
		}
	}
}

func TestPatternPriorityRetirementManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧 pattern-candidate-set-v5 用未裁决的本地整数优先级",
			"primary_candidate_id、selection_basis、候选role和priority",
			"pattern-candidate-set-v6删除伪主格排序",
			"按规则ID与名称稳定排序只用于确定性序列化",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
