package bazi

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPatternDetectorLayeredDigestValuesAreComplete(t *testing.T) {
	wantBehavior := map[string]string{
		"pattern.special.zhuanwang": "6193bc4d839aead50015b4ce6f155151e528437eb9fd269b7d49cb3657c235e7",
		"pattern.special.liangqi":   "454a7144a3ec3f6fd6a5cf2a7e2f2af5ec668fcf02c513f5ecd833a09763b55e",
		"pattern.lu.jianlu":         "d51707e2f5627ade0d053f10c9885eb7406e0b55defb6ea3b7ed6ad25c91a97b",
		"pattern.lu.yueren":         "cb4d3b6db008d758dbefa4730bf876c8ec6edf82a116733ed3bcd8131b0a6df3",
		"pattern.lu.zhuanlu":        "3e5fa70c2fb6805ecb5e4c0d57318242840866b2ee3b3bb7057fd6521deb4581",
		"pattern.lu.riren":          "d3258cb7609e23df271a827467f048799339f31623a47949de1c55e1930cbc1b",
		"pattern.aux.kuigang":       "40f31c0dd264242bb03b8e13a1befab913faeddb75e40a4a11154e5c72f8bfe9",
		"pattern.aux.ride":          "7be77b6771f51945042d3fe7ddb0a305c8840147f89c97b84754d2dc67c65a1c",
		"pattern.aux.jinshen":       "be71d3d52d03a913a3e2f8010c92480c2da401397c68d82fcdd4492881b2fe09",
		"pattern.aux.sanqi":         "7a9e9fb9f8a63f9bfd908c4b27187c8329e132cb63d410d371adc9b51de2348a",
	}
	for _, detector := range patternDetectorRegistry() {
		if len(detector.algorithmSHA256) != 64 || len(detector.behaviorSHA256) != 64 || len(detector.profileSHA256) != 64 {
			t.Errorf("detector %s has incomplete layered digests: %+v", detector.ruleID, detector)
		}
		if detector.behaviorSHA256 != wantBehavior[detector.ruleID] {
			t.Errorf("detector %s behavior SHA-256 = %s, want %s", detector.ruleID, detector.behaviorSHA256, wantBehavior[detector.ruleID])
		}
		if detector.algorithmSHA256 == detector.behaviorSHA256 || detector.behaviorSHA256 == detector.profileSHA256 {
			t.Errorf("detector %s layered digests are not distinct: %+v", detector.ruleID, detector)
		}
	}
}

func TestPatternAnalysisPublishesCanonicalDetectorProfileDigests(t *testing.T) {
	registry := patternDetectorRegistry()
	want := patternDetectorProfileDigests(registry)
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	for _, analysis := range []PatternAnalysis{
		AnalyzePatternExtended(pillars, "寅"),
		AnalyzePatternExtended(pillars[:3], "寅"),
	} {
		if !reflect.DeepEqual(analysis.DetectorProfiles, want) {
			t.Errorf("detector profiles = %+v, want %+v", analysis.DetectorProfiles, want)
		}
		if len(analysis.DetectorProfiles) != analysis.DetectorCount {
			t.Errorf("detector profile/count = %d/%d", len(analysis.DetectorProfiles), analysis.DetectorCount)
		}
		seen := make(map[string]struct{}, len(analysis.DetectorProfiles))
		for index, digest := range analysis.DetectorProfiles {
			if digest.RuleID == "" || len(digest.AlgorithmSHA256) != 64 || len(digest.BehaviorSHA256) != 64 || len(digest.ProfileSHA256) != 64 ||
				digest.AlgorithmSHA256 != patternDetectorAlgorithmSHA256(digest.RuleID) ||
				digest.BehaviorSHA256 != patternDetectorBehaviorSHA256(digest.RuleID) ||
				digest.ProfileSHA256 != patternDetectorProfileSHA256(digest.RuleID) {
				t.Errorf("invalid detector digest %d: %+v", index, digest)
			}
			if index > 0 && analysis.DetectorProfiles[index-1].RuleID >= digest.RuleID {
				t.Errorf("detector profiles are not strictly sorted: %+v", analysis.DetectorProfiles)
			}
			if _, duplicate := seen[digest.RuleID]; duplicate {
				t.Errorf("duplicate detector profile rule ID %q", digest.RuleID)
			}
			seen[digest.RuleID] = struct{}{}
		}
		payload, err := json.Marshal(analysis)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(payload), `"detector_profiles"`) || !strings.Contains(string(payload), `"algorithm_sha256"`) ||
			!strings.Contains(string(payload), `"behavior_sha256"`) || !strings.Contains(string(payload), `"profile_sha256"`) {
			t.Errorf("pattern JSON omits detector profile evidence: %s", payload)
		}
	}
}

func TestPatternDetectorProfileDigestsAreIndependentAndTamperChecked(t *testing.T) {
	digests := patternDetectorProfileDigests(patternDetectorRegistry())
	digests[0].RuleID = "mutated"
	digests[0].AlgorithmSHA256 = strings.Repeat("1", 64)
	digests[0].BehaviorSHA256 = strings.Repeat("2", 64)
	digests[0].ProfileSHA256 = strings.Repeat("0", 64)
	fresh := patternDetectorProfileDigests(patternDetectorRegistry())
	if fresh[0].RuleID == "mutated" || fresh[0].AlgorithmSHA256 == strings.Repeat("1", 64) ||
		fresh[0].BehaviorSHA256 == strings.Repeat("2", 64) || fresh[0].ProfileSHA256 == strings.Repeat("0", 64) {
		t.Fatalf("fresh detector profile digests inherited mutation: %+v", fresh[0])
	}

	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	analysis := AnalyzePatternExtended(pillars, "寅")
	for _, mutate := range []func(*PatternDetectorProfileDigest){
		func(digest *PatternDetectorProfileDigest) { digest.AlgorithmSHA256 = strings.Repeat("1", 64) },
		func(digest *PatternDetectorProfileDigest) { digest.BehaviorSHA256 = strings.Repeat("2", 64) },
		func(digest *PatternDetectorProfileDigest) { digest.ProfileSHA256 = strings.Repeat("0", 64) },
	} {
		tampered := analysis
		tampered.DetectorProfiles = append([]PatternDetectorProfileDigest(nil), analysis.DetectorProfiles...)
		mutate(&tampered.DetectorProfiles[0])
		if reflect.DeepEqual(tampered, analysis) || ValidPatternAnalysis(tampered, pillars, "寅") {
			t.Fatal("tampered layered detector digest passed strict validation")
		}
	}
}

func TestPatternDetectorProfileDigestConsumersAreSynchronized(t *testing.T) {
	checks := map[string][]string{
		"../../../../API.md": {
			`"detector_profiles"`, `"rule_id": "pattern.aux.jinshen"`, `"algorithm_sha256"`, `"behavior_sha256"`, `"profile_sha256"`,
		},
		"../../../../vue/src/api/chart.ts": {
			"export interface PatternDetectorProfileDigest", "algorithm_sha256: string", "behavior_sha256: string", "detector_profiles: PatternDetectorProfileDigest[]",
		},
	}
	for path, requiredValues := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, required := range requiredValues {
			if !strings.Contains(string(source), required) {
				t.Errorf("consumer %s missing detector profile digest contract %q", path, required)
			}
		}
	}
}

func TestPatternDetectorProfilePublicationMetadataContract(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		if count := strings.Count(table.Description, "旧响应只公开不透明的detector_manifest_sha256总摘要"); count != 1 {
			t.Errorf("detector-profile publication metadata statement count = %d, want 1", count)
		}
		if count := strings.Count(table.Description, "公开detector_profiles仍只有profile_sha256"); count != 1 {
			t.Errorf("layered detector digest metadata statement count = %d, want 1", count)
		}
		for _, fragment := range []string{
			"pattern-candidate-set-v24新增规范排序的detector_profiles",
			"只发布rule_id与profile_sha256",
			"合法与非法结果都从同一次注册表快照生成10项清单",
			"持久化重算逐字段校验并拒绝篡改",
			"总detector_manifest_sha256保持acd631f529e51ead2c50fa1c7832149ad7d994d7137be348133ecd70de2cff1a",
			"pattern-candidate-set-v30为每条规则同时公开algorithm_sha256、behavior_sha256与profile_sha256",
			"简单行为层来自canonical_truth_table_v1",
			"专旺与两气行为层来自behavior_contract_v1",
			"三层任一篡改都由持久化重算拒绝",
			"6334f79633183924f9daf4d1a695bd84281b1bb3126e853657a436068fff57d8",
			"10条profile_sha256保持不变",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
