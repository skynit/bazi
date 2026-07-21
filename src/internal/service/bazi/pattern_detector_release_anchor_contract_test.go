package bazi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
)

const patternDetectorProfileReleaseAnchorTestPath = "../../../../release/pattern-detector-profile-anchor.json"

func TestPatternDetectorProfileReleaseAnchorArtifact(t *testing.T) {
	raw, err := os.ReadFile(patternDetectorProfileReleaseAnchorTestPath)
	if err != nil {
		t.Fatal(err)
	}
	var anchor PatternDetectorProfileReleaseAnchor
	if err := json.Unmarshal(raw, &anchor); err != nil {
		t.Fatal(err)
	}
	if !ValidPatternDetectorProfileReleaseAnchor(anchor) {
		t.Fatalf("release anchor does not match runtime evidence: %+v", anchor)
	}
	var compact bytes.Buffer
	if err := json.Compact(&compact, raw); err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(compact.Bytes())
	if got := hex.EncodeToString(sum[:]); got != PatternDetectorProfileReleaseAnchorSHA256 {
		t.Fatalf("release anchor SHA-256 = %s, want %s", got, PatternDetectorProfileReleaseAnchorSHA256)
	}
	if got := patternDetectorProfilesSHA256(patternDetectorProfileDigests(patternDetectorRegistry())); got != "b00a44f1659cc578d1f6fcb321a19aa479e06c5c8ddcfb62682506a7fde0d1e1" {
		t.Fatalf("detector profile-set SHA-256 = %s", got)
	}
}

func TestPatternDetectorProfileReleaseAnchorRejectsEveryEvidenceMutation(t *testing.T) {
	anchor := loadPatternDetectorProfileReleaseAnchor(t)
	mutations := []func(*PatternDetectorProfileReleaseAnchor){
		func(value *PatternDetectorProfileReleaseAnchor) { value.Schema = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.AnchorID = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.EngineVersion = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.RuleVersion = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.RuleID = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.SchemaVersion = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.DetectorProfile = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) {
			value.DetectorManifestSHA256 = strings.Repeat("0", 64)
		},
		func(value *PatternDetectorProfileReleaseAnchor) {
			value.DetectorProfilesSHA256 = strings.Repeat("0", 64)
		},
		func(value *PatternDetectorProfileReleaseAnchor) { value.LedgerID = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.LedgerSchema = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.LedgerSHA256 = strings.Repeat("0", 64) },
		func(value *PatternDetectorProfileReleaseAnchor) { value.MigrationCount++ },
		func(value *PatternDetectorProfileReleaseAnchor) { value.ChainScheme = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.ChainHeadSHA256 = strings.Repeat("0", 64) },
		func(value *PatternDetectorProfileReleaseAnchor) { value.VerificationProfile = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.TrustBoundary = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchor) { value.ClaimBoundary = "mutated" },
	}
	for index, mutate := range mutations {
		mutated := anchor
		mutate(&mutated)
		if ValidPatternDetectorProfileReleaseAnchor(mutated) {
			t.Errorf("mutated release anchor field %d passed validation", index)
		}
	}
}

func TestPatternAnalysisPublishesAndValidatesReleaseAnchorReference(t *testing.T) {
	want := PatternDetectorProfileReleaseAnchorReference{
		Schema:              "pattern_detector_profile_release_anchor_v1",
		AnchorID:            "bazi.pattern-detector-profile-release-anchor-v34",
		ArtifactPath:        "release/pattern-detector-profile-anchor.json",
		SHA256:              "ebd6323f28715695aa3c4ee9038e74d261c9fa34b422037266c4b097e3086a2e",
		VerificationProfile: "repository_ci_cross_check_v1",
		TrustBoundary:       "unsigned_repository_ci_artifact",
		ClaimBoundary:       "digest_evidence_only",
	}
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "子"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "寅"}, {Gan: "戊", Zhi: "辰"},
	}
	for _, analysis := range []PatternAnalysis{
		AnalyzePatternExtended(pillars, "寅"),
		AnalyzePatternExtended(pillars[:3], "寅"),
	} {
		if !reflect.DeepEqual(analysis.DetectorReleaseAnchor, want) {
			t.Errorf("release anchor reference = %+v, want %+v", analysis.DetectorReleaseAnchor, want)
		}
		payload, err := json.Marshal(analysis)
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range []string{`"detector_profile_release_anchor"`, want.AnchorID, want.SHA256, want.TrustBoundary} {
			if !strings.Contains(string(payload), fragment) {
				t.Errorf("pattern JSON missing release anchor evidence %q: %s", fragment, payload)
			}
		}
	}

	analysis := AnalyzePatternExtended(pillars, "寅")
	for _, mutate := range []func(*PatternDetectorProfileReleaseAnchorReference){
		func(value *PatternDetectorProfileReleaseAnchorReference) { value.Schema = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchorReference) { value.AnchorID = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchorReference) { value.ArtifactPath = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchorReference) { value.SHA256 = strings.Repeat("0", 64) },
		func(value *PatternDetectorProfileReleaseAnchorReference) { value.VerificationProfile = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchorReference) { value.TrustBoundary = "mutated" },
		func(value *PatternDetectorProfileReleaseAnchorReference) { value.ClaimBoundary = "mutated" },
	} {
		tampered := analysis
		mutate(&tampered.DetectorReleaseAnchor)
		if reflect.DeepEqual(tampered, analysis) || ValidPatternAnalysis(tampered, pillars, "寅") {
			t.Fatal("tampered release anchor reference passed strict validation")
		}
	}
}

func TestPatternDetectorProfileReleaseAnchorCIContract(t *testing.T) {
	workflow, err := os.ReadFile("../../../../.github/workflows/pattern-profile-release-anchor.yml")
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		"go test -run '^TestPatternDetectorProfileReleaseAnchorArtifact$' -count=1 ./internal/service/bazi",
		"actions/upload-artifact@v4",
		"release/pattern-detector-profile-anchor.json",
		"pattern-detector-profile-anchor-v34",
	} {
		if !strings.Contains(string(workflow), fragment) {
			t.Errorf("release anchor workflow missing %q", fragment)
		}
	}
	production, err := os.ReadFile("pattern_profile_release_anchor.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(production), "//go:embed") {
		t.Fatal("release anchor must remain a non-embedded CI artifact")
	}
}

func TestPatternDetectorProfileReleaseAnchorConsumersAndMetadataAreSynchronized(t *testing.T) {
	checks := map[string][]string{
		"../../../../API.md": {
			`"detector_profile_release_anchor"`, PatternDetectorProfileReleaseAnchorID,
			PatternDetectorProfileReleaseAnchorSHA256, PatternDetectorProfileReleaseAnchorTrustBoundary,
		},
		"../../../../vue/src/api/chart.ts": {
			"export interface PatternDetectorProfileReleaseAnchorReference",
			"detector_profile_release_anchor: PatternDetectorProfileReleaseAnchorReference",
			PatternDetectorProfileReleaseAnchorSHA256, PatternDetectorProfileReleaseAnchorTrustBoundary,
		},
	}
	for path, fragments := range checks {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, fragment := range fragments {
			if !strings.Contains(string(source), fragment) {
				t.Errorf("consumer %s missing release anchor contract %q", path, fragment)
			}
		}
	}

	if EngineVersion != "bazi-engine-2026-07-17.27" || RuleVersion != "bazi-rules-2026-07-17.27" ||
		PatternRuleID != "bazi.pattern-candidate-set-v34" || PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" {
		t.Fatalf("release anchor version contract = %s/%s/%s/%s/%s", EngineVersion, RuleVersion, PatternRuleID, PatternSchemaVersion, PatternDetectorProfile)
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		if table.Version != "2026-07-17.27" || strings.Count(table.Description, "本地摘要链仍可与账本和常量在同一次修改中共同重写") != 1 {
			t.Errorf("release anchor metadata contract is incomplete: %+v", table)
		}
		for _, fragment := range []string{
			"pattern-candidate-set-v34新增pattern_detector_profile_release_anchor_v1",
			"release/pattern-detector-profile-anchor.json不嵌入生产二进制",
			"b00a44f1659cc578d1f6fcb321a19aa479e06c5c8ddcfb62682506a7fde0d1e1",
			"a72422e12e07adae349c147b3581f8c4829368f134f00a4f229c9a1c29d21825",
			"07dc296ad9e5dd0f834e40256c1e0f6033eb0ded435d0c76be6a0602ae0113bd",
			"repository_ci_cross_check_v1只运行internal/service/bazi锚合同",
			PatternDetectorProfileReleaseAnchorSHA256,
			"trust_boundary固定为unsigned_repository_ci_artifact",
			"不是签名tag、透明日志或外部时间戳",
			"只提高工程发布交叉验证",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q", fragment)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}

func loadPatternDetectorProfileReleaseAnchor(t *testing.T) PatternDetectorProfileReleaseAnchor {
	t.Helper()
	raw, err := os.ReadFile(patternDetectorProfileReleaseAnchorTestPath)
	if err != nil {
		t.Fatal(err)
	}
	var anchor PatternDetectorProfileReleaseAnchor
	if err := json.Unmarshal(raw, &anchor); err != nil {
		t.Fatal(err)
	}
	return anchor
}
