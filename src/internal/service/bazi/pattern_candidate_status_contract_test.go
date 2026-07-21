package bazi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternCandidateSetOwnsValidationAndInterpretationStatus(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	if analysis.ValidationStatus != "not_validated" || analysis.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("candidate-set statuses = %s/%s", analysis.ValidationStatus, analysis.InterpretationStatus)
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
		for _, forbidden := range []string{"validation_status", "interpretation_status"} {
			if _, exists := candidate[forbidden]; exists {
				t.Errorf("candidate repeats set-level status %q: %s", forbidden, payload)
			}
		}
	}
}

func TestPatternCandidateStatusConsumersAreRemoved(t *testing.T) {
	checks := map[string][]string{
		"pattern_candidates.go":                        {"candidate.ValidationStatus", "candidate.InterpretationStatus"},
		"../interpretation/bazi.go":                    {"candidate.ValidationStatus", "candidate.InterpretationStatus"},
		"../../../../vue/src/components/BaziChart.vue": {"candidate.validation_status", "candidate.interpretation_status"},
	}
	for path, forbiddenValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range forbiddenValues {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("consumer %s still contains duplicate candidate status %q", path, forbidden)
			}
		}
	}

	assertPatternSectionOmitsStatuses(t, "../../../../vue/src/api/chart.ts", "export interface PatternCandidate", "export interface PatternInputSnapshot")
	assertPatternSectionOmitsStatuses(t, "../../../../API.md", `"candidates": [`, `"status": "observed"`)
}

func assertPatternSectionOmitsStatuses(t *testing.T, path, startMarker, endMarker string) {
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
	for _, forbidden := range []string{"validation_status", "interpretation_status"} {
		if strings.Contains(section, forbidden) {
			t.Errorf("pattern candidate section in %s still contains %s: %s", path, forbidden, section)
		}
	}
}

func TestPatternCandidateStatusRetirementManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧 pattern-candidate-set-v8 在每个候选重复固定状态",
			"与集合级validation_status和interpretation_status完全相同",
			"pattern-candidate-set-v9删除候选级重复状态",
			"集合级状态作为唯一裁决边界",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
