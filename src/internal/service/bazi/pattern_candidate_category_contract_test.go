package bazi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternCategoryIsTheSingleCandidateClassification(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "癸", Zhi: "亥"}, {Gan: "甲", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "丁", Zhi: "卯"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	wants := map[string]string{
		"pattern.lu.jianlu":  patternCategoryStructural,
		"pattern.lu.zhuanlu": patternCategoryStructural,
		"pattern.aux.ride":   patternCategoryAuxiliary,
	}
	for ruleID, wantCategory := range wants {
		candidate, ok := luPatternCandidateByID(analysis.Candidates, ruleID)
		if !ok || candidate.Category != wantCategory {
			t.Errorf("candidate %s = %+v, want category %s", ruleID, candidate, wantCategory)
		}
	}
	for _, candidate := range analysis.Candidates {
		if candidate.Category != patternCategoryStructural && candidate.Category != patternCategoryAuxiliary {
			t.Errorf("candidate has unknown category: %+v", candidate)
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
		if _, exists := candidate["pattern_type"]; exists {
			t.Errorf("candidate exposes conflicting second classification: %s", payload)
		}
	}
}

func TestPatternTypeConsumersAreRemoved(t *testing.T) {
	checks := map[string][]string{
		"pattern.go":                                   {"PatternType"},
		"pattern_candidates.go":                        {"PatternType", `json:"pattern_type"`},
		"../interpretation/bazi.go":                    {"candidate.PatternType"},
		"../../../../vue/src/components/BaziChart.vue": {"candidate.pattern_type"},
	}
	for path, forbiddenValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range forbiddenValues {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("consumer %s still contains conflicting pattern classification %q", path, forbidden)
			}
		}

	}

	assertPatternSectionOmitsPatternType(t, "../../../../vue/src/api/chart.ts", "export interface PatternCandidate", "export interface PatternInputSnapshot")
	assertPatternSectionOmitsPatternType(t, "../../../../API.md", `"candidates": [`, `"status": "observed"`)
}

func assertPatternSectionOmitsPatternType(t *testing.T, path, startMarker, endMarker string) {
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
	if strings.Contains(section, "pattern_type") {
		t.Errorf("pattern candidate section in %s still contains pattern_type: %s", path, section)
	}
}

func TestPatternCategoryManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧 pattern-candidate-set-v10 同时发布pattern_type与category",
			"魁罡和日德被注册为辅助特征却标成特殊格局",
			"pattern-candidate-set-v11删除冲突pattern_type",
			"category作为唯一候选分类",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
