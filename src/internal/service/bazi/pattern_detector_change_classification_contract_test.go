package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternDetectorProfileChangeContractIsCanonicalAndPublished(t *testing.T) {
	want := PatternDetectorProfileChangeContract{
		Scheme:       "layered_detector_digest_delta_v1",
		AlignmentKey: "rule_id",
		Classes: []PatternDetectorProfileChangeClass{
			"detector_added",
			"detector_removed",
			"algorithm_digest_changed",
			"behavior_evidence_digest_changed",
			"semantic_profile_digest_changed",
			"layered_digests_unchanged",
		},
		BehaviorEvidenceScope: "simple_full_truth_table_complex_partial_contract",
		InferenceBoundary:     "digest_evidence_only",
	}
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	for _, analysis := range []PatternAnalysis{
		AnalyzePatternExtended(pillars, "寅"),
		AnalyzePatternExtended(pillars[:3], "寅"),
	} {
		if !reflect.DeepEqual(analysis.DetectorChangeContract, want) {
			t.Errorf("detector change contract = %+v, want %+v", analysis.DetectorChangeContract, want)
		}
		payload, err := json.Marshal(analysis)
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range []string{
			`"detector_profile_change_contract"`, `"layered_detector_digest_delta_v1"`,
			`"behavior_evidence_digest_changed"`, `"digest_evidence_only"`,
		} {
			if !strings.Contains(string(payload), fragment) {
				t.Errorf("pattern JSON missing %s: %s", fragment, payload)
			}
		}
	}

	mutated := patternDetectorProfileChangeContract()
	mutated.Classes[0] = "mutated"
	if fresh := patternDetectorProfileChangeContract(); reflect.DeepEqual(fresh, mutated) || fresh.Classes[0] != PatternDetectorAdded {
		t.Fatalf("fresh detector change contract inherited mutation: %+v", fresh)
	}
}

func TestPatternDetectorProfileChangeContractTamperingIsRejected(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	for _, mutate := range []func(*PatternDetectorProfileChangeContract){
		func(contract *PatternDetectorProfileChangeContract) { contract.Scheme = "mutated" },
		func(contract *PatternDetectorProfileChangeContract) { contract.AlignmentKey = "mutated" },
		func(contract *PatternDetectorProfileChangeContract) { contract.Classes[0] = "mutated" },
		func(contract *PatternDetectorProfileChangeContract) { contract.BehaviorEvidenceScope = "mutated" },
		func(contract *PatternDetectorProfileChangeContract) { contract.InferenceBoundary = "mutated" },
	} {
		tampered := analysis
		tampered.DetectorChangeContract.Classes = append([]PatternDetectorProfileChangeClass(nil), analysis.DetectorChangeContract.Classes...)
		mutate(&tampered.DetectorChangeContract)
		if reflect.DeepEqual(tampered, analysis) || ValidPatternAnalysis(tampered, pillars, "寅") {
			t.Fatal("tampered detector change contract passed strict validation")
		}
	}
}

func TestComparePatternDetectorProfilesClassifiesOnlyObservedDigestDeltas(t *testing.T) {
	before := []PatternDetectorProfileDigest{
		profileDigestFixture("pattern.b", '1', '2', '3'),
		profileDigestFixture("pattern.a", '4', '5', '6'),
		profileDigestFixture("pattern.removed", '7', '8', '9'),
	}
	after := []PatternDetectorProfileDigest{
		profileDigestFixture("pattern.added", 'a', 'b', 'c'),
		profileDigestFixture("pattern.a", 'd', 'e', 'f'),
		profileDigestFixture("pattern.b", '1', '2', '3'),
	}
	got := ComparePatternDetectorProfiles(before, after)
	want := PatternDetectorProfileChangeSet{
		Scheme: "layered_detector_digest_delta_v1",
		Status: PatternDetectorProfilesCompared,
		Changes: []PatternDetectorProfileChange{
			{RuleID: "pattern.a", Classes: []PatternDetectorProfileChangeClass{
				PatternDetectorAlgorithmDigestChanged,
				PatternDetectorBehaviorDigestChanged,
				PatternDetectorSemanticDigestChanged,
			}},
			{RuleID: "pattern.added", Classes: []PatternDetectorProfileChangeClass{PatternDetectorAdded}},
			{RuleID: "pattern.b", Classes: []PatternDetectorProfileChangeClass{PatternDetectorLayeredDigestsUnchanged}},
			{RuleID: "pattern.removed", Classes: []PatternDetectorProfileChangeClass{PatternDetectorRemoved}},
		},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("profile change set = %+v, want %+v", got, want)
	}

	algorithmOnly := append([]PatternDetectorProfileDigest(nil), before...)
	algorithmOnly[0].AlgorithmSHA256 = strings.Repeat("a", 64)
	got = ComparePatternDetectorProfiles(before, algorithmOnly)
	for _, change := range got.Changes {
		if change.RuleID == "pattern.b" && !reflect.DeepEqual(change.Classes, []PatternDetectorProfileChangeClass{PatternDetectorAlgorithmDigestChanged}) {
			t.Fatalf("algorithm-only delta inferred other changes: %+v", change)
		}
	}
}

func TestComparePatternDetectorProfilesRejectsInvalidSnapshots(t *testing.T) {
	valid := []PatternDetectorProfileDigest{profileDigestFixture("pattern.a", '1', '2', '3')}
	invalidCases := [][]PatternDetectorProfileDigest{
		{{RuleID: "", AlgorithmSHA256: strings.Repeat("1", 64), BehaviorSHA256: strings.Repeat("2", 64), ProfileSHA256: strings.Repeat("3", 64)}},
		{{RuleID: "pattern.a", AlgorithmSHA256: "short", BehaviorSHA256: strings.Repeat("2", 64), ProfileSHA256: strings.Repeat("3", 64)}},
		{{RuleID: "pattern.a", AlgorithmSHA256: strings.Repeat("A", 64), BehaviorSHA256: strings.Repeat("2", 64), ProfileSHA256: strings.Repeat("3", 64)}},
		{valid[0], valid[0]},
	}
	for _, invalid := range invalidCases {
		for _, pair := range [][2][]PatternDetectorProfileDigest{{invalid, valid}, {valid, invalid}} {
			got := ComparePatternDetectorProfiles(pair[0], pair[1])
			if got.Status != PatternDetectorProfilesInvalidInput || got.Scheme != PatternDetectorProfileChangeScheme || len(got.Changes) != 0 || got.Changes == nil {
				t.Errorf("invalid comparison did not fail closed: %+v", got)
			}
		}
	}

	got := ComparePatternDetectorProfiles(nil, nil)
	if got.Status != PatternDetectorProfilesCompared || len(got.Changes) != 0 || got.Changes == nil {
		t.Errorf("empty valid snapshots = %+v", got)
	}
}

func TestPatternDetectorChangeClassificationConsumersAreSynchronized(t *testing.T) {
	checks := map[string][]string{
		"../../../../API.md": {
			`"detector_profile_change_contract"`, `"layered_detector_digest_delta_v1"`,
			`"behavior_evidence_digest_changed"`, `"digest_evidence_only"`,
		},
		"../../../../vue/src/api/chart.ts": {
			"export type PatternDetectorProfileChangeClass", "detector_profile_change_contract: PatternDetectorProfileChangeContract",
			"simple_full_truth_table_complex_partial_contract", "digest_evidence_only",
		},
	}
	for path, requiredValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, required := range requiredValues {
			if !strings.Contains(string(source), required) {
				t.Errorf("consumer %s missing detector change contract %q", path, required)
			}
		}
	}
}

func TestPatternDetectorChangeClassificationMetadataContract(t *testing.T) {
	if EngineVersion != "bazi-engine-2026-07-17.27" || RuleVersion != "bazi-rules-2026-07-17.27" ||
		PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("version contract = %s/%s/%s/%s/%d", EngineVersion, RuleVersion, PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		if table.Version != "2026-07-17.27" {
			t.Errorf("pattern metadata version = %q", table.Version)
		}
		if count := strings.Count(table.Description, "三层摘要虽已公开，但仍只提供原始散列"); count != 1 {
			t.Errorf("change-classification metadata statement count = %d, want 1", count)
		}
		for _, fragment := range []string{
			"pattern-candidate-set-v31新增layered_detector_digest_delta_v1变更分类合同",
			"按rule_id对齐",
			"behavior_evidence_digest_changed",
			"空ID、重复ID或非64位小写十六进制摘要统一失败关闭",
			"simple_full_truth_table_complex_partial_contract",
			"推断边界固定为digest_evidence_only",
			"不证明传统正确性、完整四柱行为或现实预测等价",
			"6334f79633183924f9daf4d1a695bd84281b1bb3126e853657a436068fff57d8保持不变",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}

func profileDigestFixture(ruleID string, algorithm, behavior, profile byte) PatternDetectorProfileDigest {
	return PatternDetectorProfileDigest{
		RuleID:          ruleID,
		AlgorithmSHA256: strings.Repeat(string(algorithm), 64),
		BehaviorSHA256:  strings.Repeat(string(behavior), 64),
		ProfileSHA256:   strings.Repeat(string(profile), 64),
	}
}
