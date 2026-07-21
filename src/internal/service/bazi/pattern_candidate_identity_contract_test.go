package bazi

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternRuleIDIsTheSingleCandidateIdentity(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	if len(analysis.Candidates) < 2 {
		t.Fatalf("fixture must retain overlapping candidates: %+v", analysis)
	}
	seen := make(map[string]struct{}, len(analysis.Candidates))
	for _, candidate := range analysis.Candidates {
		if candidate.RuleID == "" {
			t.Fatal("pattern candidate has empty rule ID")
		}
		if _, exists := seen[candidate.RuleID]; exists {
			t.Fatalf("duplicate pattern rule ID %q: %+v", candidate.RuleID, analysis.Candidates)
		}
		seen[candidate.RuleID] = struct{}{}
	}

	payload, err := json.Marshal(analysis)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(payload), `"candidate_id"`) {
		t.Fatalf("pattern response still exposes duplicate candidate identity: %s", payload)
	}
}

func TestPatternCandidateIDConsumersAreRemoved(t *testing.T) {
	checks := map[string][]string{
		"pattern_candidates.go":                        {"CandidateID", `json:"candidate_id"`},
		"../../../../vue/src/components/BaziChart.vue": {"candidate.candidate_id"},
	}
	for path, forbiddenValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range forbiddenValues {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("consumer %s still contains duplicate pattern identity %q", path, forbidden)
			}
		}
	}

	assertSectionOmitsPatternCandidateID(t, "../../../../vue/src/api/chart.ts", "export interface PatternCandidate", "export interface PatternInputSnapshot")
	assertSectionOmitsPatternCandidateID(t, "../../../../API.md", `"candidates": [`, `"status": "observed"`)
}

func TestPatternFinalizationUsesRuleIDAsOnlyInternalIdentity(t *testing.T) {
	source, err := os.ReadFile("pattern_candidates.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		`match.ruleID + "|" + match.analysis.PatternName`,
		"matches[i].analysis.PatternName < matches[j].analysis.PatternName",
		"deduped := make([]patternMatch",
	} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("pattern finalization retains second identity path %q", forbidden)
		}
	}
	for _, required := range []string{
		"return matches[i].ruleID < matches[j].ruleID",
		"make([]PatternCandidate, 0, len(matches))",
		"for _, match := range matches",
	} {
		if !strings.Contains(string(source), required) {
			t.Errorf("pattern finalization missing unique rule-ID path %q", required)
		}
	}
}

func assertSectionOmitsPatternCandidateID(t *testing.T, path, startMarker, endMarker string) {
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
	if strings.Contains(section, "candidate_id") {
		t.Errorf("pattern candidate section in %s still contains candidate_id: %s", path, section)
	}
}

func TestPatternCandidateIdentityRetirementManifest(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"旧 pattern-candidate-set-v7 的candidate_id始终等于rule_id",
			"没有独立生成、选择或引用语义",
			"pattern-candidate-set-v8删除重复candidate_id",
			"rule_id作为唯一候选身份",
			"旧最终汇总器以rule_id与PatternName拼接复合去重键",
			"pattern-candidate-set-v19删除不可达的静默去重和名称次级排序",
			"候选只按唯一rule_id确定序列化顺序并直接映射",
			"正式候选集合、检测器清单与detector_manifest_sha256均不变",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
