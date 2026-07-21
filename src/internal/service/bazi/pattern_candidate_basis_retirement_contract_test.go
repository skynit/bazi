package bazi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternCandidatePresenceIsTheDetectorMatchFact(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	if len(analysis.Candidates) < 2 {
		t.Fatalf("fixture must retain overlapping candidates: %+v", analysis)
	}
	for _, candidate := range analysis.Candidates {
		if candidate.RuleID == "" || candidate.Source == "" {
			t.Errorf("candidate lacks auditable identity or source: %+v", candidate)
		}
	}

	payload, err := json.Marshal(analysis)
	if err != nil {
		t.Fatal(err)
	}
	var top map[string]json.RawMessage
	if err := json.Unmarshal(payload, &top); err != nil {
		t.Fatal(err)
	}
	var candidates []map[string]json.RawMessage
	if err := json.Unmarshal(top["candidates"], &candidates); err != nil {
		t.Fatal(err)
	}
	for _, candidate := range candidates {
		if _, exists := candidate["basis"]; exists {
			t.Errorf("candidate repeats its existence as a fixed basis: %s", payload)
		}
	}
}

func TestPatternCandidateBasisConsumersAreRemoved(t *testing.T) {
	checks := map[string][]string{
		"pattern_candidates.go":                        {"candidate.Basis", "local_detector_conditions_matched"},
		"../interpretation/bazi.go":                    {"candidate.Basis"},
		"../../../../vue/src/components/BaziChart.vue": {"candidate.basis"},
	}
	for path, forbiddenValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range forbiddenValues {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("consumer %s still contains tautological pattern basis %q", path, forbidden)
			}
		}
	}

	assertPatternSectionOmitsBasis(t, "../../../../vue/src/api/chart.ts", "export interface PatternCandidate", "export interface PatternInputSnapshot")
	assertPatternSectionOmitsBasis(t, "../../../../API.md", `"candidates": [`, `"status": "observed"`)
}

func assertPatternSectionOmitsBasis(t *testing.T, path, startMarker, endMarker string) {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	start := strings.Index(string(source), startMarker)
	if start < 0 {
		t.Fatalf("start marker %q missing from %s", startMarker, path)
	}
	endOffset := strings.Index(string(source)[start+len(startMarker):], endMarker)
	if endOffset < 0 {
		t.Fatalf("end marker %q missing from %s", endMarker, path)
	}
	section := string(source)[start : start+len(startMarker)+endOffset]
	if strings.Contains(section, "basis") {
		t.Errorf("pattern candidate section in %s still contains basis: %s", path, section)
	}
}

func TestPatternCandidateBasisRetirementManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧 pattern-candidate-set-v9 的basis固定为local_detector_conditions_matched",
			"候选进入集合本身已经证明检测条件命中",
			"pattern-candidate-set-v10删除同义basis字段",
			"rule_id和source继续提供可审计身份与来源",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
