package bazi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	calculationProfileManifestSchema = "fortune_calculation_profile_manifest_v1"
	calculationProfileRepoRoot       = "../../../.."
)

type calculationProfileManifest struct {
	Schema            string                             `json:"schema"`
	ManifestVersion   string                             `json:"manifest_version"`
	ProfileID         string                             `json:"profile_id"`
	System            string                             `json:"system"`
	Title             string                             `json:"title"`
	Status            string                             `json:"status"`
	RuntimeSelectable bool                               `json:"runtime_selectable"`
	ActivationPolicy  string                             `json:"activation_policy"`
	RuntimeBinding    calculationProfileRuntimeBinding   `json:"runtime_binding"`
	CalculationScope  []string                           `json:"calculation_scope"`
	RuleModules       []calculationProfileRuleModule     `json:"rule_modules"`
	SourceEvidence    []calculationProfileSourceEvidence `json:"source_evidence"`
	UnsupportedRules  []calculationProfileUnsupported    `json:"unsupported_rules"`
	Disputed          []calculationProfileDispute        `json:"disputed_conventions"`
	GoldGates         []calculationProfileGoldGate       `json:"gold_gates"`
	MixingPolicy      calculationProfileMixingPolicy     `json:"mixing_policy"`
	ClaimBoundary     calculationProfileClaimBoundary    `json:"claim_boundary"`
}

type calculationProfileRuntimeBinding struct {
	Kind              string `json:"kind"`
	RegisteredID      string `json:"registered_profile_id"`
	ReferenceID       string `json:"reference_profile_id"`
	EngineVersion     string `json:"engine_version"`
	RuleVersion       string `json:"rule_version"`
	School            string `json:"school"`
	EquivalenceStatus string `json:"equivalence_status"`
}

type calculationProfileRuleModule struct {
	RuleID       string   `json:"rule_id"`
	Status       string   `json:"status"`
	Scope        string   `json:"scope"`
	EvidenceTier string   `json:"evidence_tier"`
	EvidenceRefs []string `json:"evidence_refs"`
	Limitations  []string `json:"limitations"`
}

type calculationProfileSourceEvidence struct {
	ID           string   `json:"id"`
	Tier         string   `json:"tier"`
	Kind         string   `json:"kind"`
	ArtifactPath string   `json:"artifact_path"`
	Repository   string   `json:"repository"`
	Commit       string   `json:"commit"`
	SourcePath   string   `json:"source_path"`
	SHA256       string   `json:"sha256"`
	LicenseScope string   `json:"license_scope"`
	ReviewStatus string   `json:"review_status"`
	CaseCount    int      `json:"case_count"`
	Limitations  []string `json:"limitations"`
}

type calculationProfileUnsupported struct {
	RuleID           string `json:"rule_id"`
	Reason           string `json:"reason"`
	RequiredEvidence string `json:"required_evidence"`
}

type calculationProfileDispute struct {
	ID             string   `json:"id"`
	Question       string   `json:"question"`
	Options        []string `json:"options"`
	Selected       string   `json:"selected"`
	Status         string   `json:"status"`
	ResolutionGate string   `json:"resolution_gate"`
}

type calculationProfileGoldGate struct {
	GateID                           string `json:"gate_id"`
	MinimumCases                     int    `json:"minimum_cases"`
	CurrentQualifiedCases            int    `json:"current_qualified_cases"`
	IndependentReviewers             int    `json:"independent_reviewers"`
	RequiresThirdReviewerForDisputes bool   `json:"requires_third_reviewer_for_disputes"`
	FrozenTestRequired               bool   `json:"frozen_test_required"`
	Status                           string `json:"status"`
	AdmissionScope                   string `json:"admission_scope"`
}

type calculationProfileMixingPolicy struct {
	Fallback             string `json:"fallback"`
	CrossProfileImport   string `json:"cross_profile_rule_import"`
	SilentDefault        string `json:"silent_default"`
	MissingRuleBehavior  string `json:"missing_rule_behavior"`
	InterpretationMixing string `json:"interpretation_mixing"`
}

type calculationProfileClaimBoundary struct {
	PublishableAccuracy          bool     `json:"publishable_accuracy"`
	TraditionalCorrectnessProven bool     `json:"traditional_correctness_proven"`
	PredictiveValidityProven     bool     `json:"predictive_validity_proven"`
	EventProbabilityAllowed      bool     `json:"event_probability_allowed"`
	AllowedClaims                []string `json:"allowed_claims"`
	ProhibitedClaims             []string `json:"prohibited_claims"`
}

func TestCalculationProfileManifestsAreStrictNonRuntimeResearchArtifacts(t *testing.T) {
	expected := map[string]struct {
		path   string
		system string
		status string
		hash   string
	}{
		"ziping-fuyi-v2": {
			path: "research/profiles/ziping-fuyi-v2.json", system: "bazi",
			status: "draft_partial_implementation_not_selectable",
			hash:   "6f041e00b350be80ebdbc2c06d19c57ea7d9013f6fbfeb4295bd4e83eb73a96e",
		},
		"ziwei-sanhe-v1": {
			path: "research/profiles/ziwei-sanhe-v1.json", system: "ziwei",
			status: "draft_not_selectable",
			hash:   "4fd9e55d67b40b0a1214933151088694f1e61385e11145c5003441d6f07297c6",
		},
		"ziwei-nihaixia-v1": {
			path: "research/profiles/ziwei-nihaixia-v1.json", system: "ziwei",
			status: "draft_not_selectable_source_missing",
			hash:   "c4c080b2626d42be76e0f75b3254e1eab8e9e36a107df3f4ff7e2bd0c4d31562",
		},
	}

	seen := make(map[string]bool, len(expected))
	for profileID, want := range expected {
		manifest, raw := loadCalculationProfileManifest(t, want.path)
		if manifest.ProfileID != profileID || manifest.System != want.system || manifest.Status != want.status {
			t.Errorf("manifest %s identity = %s/%s/%s", want.path, manifest.ProfileID, manifest.System, manifest.Status)
		}
		if got := sha256Hex(raw); got != want.hash {
			t.Errorf("manifest %s SHA-256 = %s, want %s", profileID, got, want.hash)
		}
		if seen[manifest.ProfileID] {
			t.Errorf("duplicate profile ID %q", manifest.ProfileID)
		}
		seen[manifest.ProfileID] = true
		validateCalculationProfileManifest(t, manifest)
	}
}

func TestCalculationProfileManifestEvidenceMatchesFrozenArtifacts(t *testing.T) {
	for _, path := range []string{
		"research/profiles/ziping-fuyi-v2.json",
		"research/profiles/ziwei-sanhe-v1.json",
		"research/profiles/ziwei-nihaixia-v1.json",
	} {
		manifest, _ := loadCalculationProfileManifest(t, path)
		for _, evidence := range manifest.SourceEvidence {
			if evidence.ArtifactPath == "" {
				continue
			}
			raw, err := os.ReadFile(filepath.Join(calculationProfileRepoRoot, evidence.ArtifactPath))
			if err != nil {
				t.Errorf("%s evidence %s: %v", manifest.ProfileID, evidence.ID, err)
				continue
			}
			got := sha256Hex(raw)
			if evidence.Kind == "local_release_artifact" {
				var compact bytes.Buffer
				if err := json.Compact(&compact, raw); err != nil {
					t.Errorf("%s evidence %s is not JSON: %v", manifest.ProfileID, evidence.ID, err)
					continue
				}
				got = sha256Hex(compact.Bytes())
			}
			if got != evidence.SHA256 {
				t.Errorf("%s evidence %s SHA-256 = %s, want %s", manifest.ProfileID, evidence.ID, got, evidence.SHA256)
			}
		}
	}
}

func TestCalculationProfileManifestEvidenceTiersDoNotPromoteSilverToGold(t *testing.T) {
	type fixture struct {
		Version  string `json:"version"`
		Frozen   bool   `json:"frozen"`
		Metadata struct {
			Tier                string `json:"tier"`
			ReviewStatus        string `json:"review_status"`
			PublishableAccuracy bool   `json:"publishable_accuracy"`
		} `json:"metadata"`
		Cases []json.RawMessage `json:"cases"`
	}

	checks := []struct {
		path, version, tier, review string
		cases                       int
	}{
		{"src/internal/service/testdata/bazi_external_silver.json", "1.2", "silver", "cross_checked_not_gold", 60},
		{"src/internal/service/testdata/ziwei_iztro_silver.json", "1.1", "silver", "cross_checked_not_gold", 37},
	}
	for _, check := range checks {
		raw, err := os.ReadFile(filepath.Join(calculationProfileRepoRoot, check.path))
		if err != nil {
			t.Fatal(err)
		}
		var value fixture
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatal(err)
		}
		if value.Version != check.version || value.Metadata.Tier != check.tier ||
			value.Metadata.ReviewStatus != check.review || value.Metadata.PublishableAccuracy || len(value.Cases) != check.cases {
			t.Errorf("fixture %s evidence contract changed: version=%s metadata=%+v cases=%d", check.path, value.Version, value.Metadata, len(value.Cases))
		}
	}

	raw, err := os.ReadFile(filepath.Join(calculationProfileRepoRoot, "src/internal/service/testdata/ziwei_full_chart_gold.json"))
	if err != nil {
		t.Fatal(err)
	}
	var gold fixture
	if err := json.Unmarshal(raw, &gold); err != nil {
		t.Fatal(err)
	}
	if gold.Version != "0.1-draft" || gold.Frozen || len(gold.Cases) != 0 {
		t.Fatalf("ZiWei Gold registry must remain empty and unfrozen until expert admission: version=%s frozen=%v cases=%d", gold.Version, gold.Frozen, len(gold.Cases))
	}
}

func TestDraftCalculationProfilesAreNotRegisteredInProduction(t *testing.T) {
	profileIDs := []string{"ziping-fuyi-v2", "ziwei-sanhe-v1", "ziwei-nihaixia-v1"}
	for _, directory := range []string{".", "../ziwei"} {
		paths, err := filepath.Glob(filepath.Join(directory, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			if strings.HasSuffix(path, "_test.go") {
				continue
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, profileID := range profileIDs {
				if strings.Contains(string(raw), profileID) {
					t.Errorf("draft profile %q leaked into production Go file %s", profileID, path)
				}
			}
		}
	}

	ziweiProfile, err := os.ReadFile("../ziwei/profile.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, fragment := range []string{
		`DefaultProfileID   = "ziwei-local-composite-v2"`,
		`ZiWeiEngineVersion = "ziwei-local-go-2026-07-16.47"`,
		`ZiWeiRuleVersion   = "ziwei-rules-2026-07-16.47"`,
	} {
		if !strings.Contains(string(ziweiProfile), fragment) {
			t.Errorf("ZiWei reference runtime changed without manifest adjudication: missing %q", fragment)
		}
	}
}

func TestCalculationProfileManifestSpecificBoundaries(t *testing.T) {
	ziping, _ := loadCalculationProfileManifest(t, "research/profiles/ziping-fuyi-v2.json")
	if ziping.RuntimeBinding.EngineVersion != EngineVersion || ziping.RuntimeBinding.RuleVersion != RuleVersion ||
		ziping.RuntimeBinding.School != RuleSchool || ziping.RuntimeBinding.Kind != "partial_current_runtime_reference" {
		t.Errorf("ziping runtime reference = %+v, current = %s/%s/%s", ziping.RuntimeBinding, EngineVersion, RuleVersion, RuleSchool)
	}
	if !manifestHasModuleStatus(ziping, "implemented_current_engine") ||
		!manifestHasModuleStatus(ziping, "implemented_candidate_only") || !manifestHasModuleStatus(ziping, "draft_unimplemented") {
		t.Error("ziping manifest must distinguish implemented facts, candidates, and unimplemented FuYi rules")
	}

	sanhe, _ := loadCalculationProfileManifest(t, "research/profiles/ziwei-sanhe-v1.json")
	if sanhe.RuntimeBinding.ReferenceID != "ziwei-local-composite-v2" ||
		sanhe.RuntimeBinding.EquivalenceStatus != "reference_runtime_is_not_sanhe_profile" ||
		!manifestHasModuleStatus(sanhe, "reference_implementation_only_not_profile_adjudicated") ||
		!manifestHasModuleStatus(sanhe, "draft_unimplemented") {
		t.Errorf("sanhe manifest does not preserve reference/non-equivalence boundary: %+v", sanhe.RuntimeBinding)
	}

	nihaixia, _ := loadCalculationProfileManifest(t, "research/profiles/ziwei-nihaixia-v1.json")
	if nihaixia.RuntimeBinding.EquivalenceStatus != "reference_runtime_is_not_nihaixia_profile" ||
		!manifestHasEvidence(nihaixia, "nihaixia_primary_source_gap", "missing_primary_source", 0) ||
		!manifestHasEvidence(nihaixia, "nihaixia_dataset_gap", "missing_dataset_and_provenance", 0) {
		t.Errorf("nihaixia missing-source boundary is incomplete: %+v", nihaixia.RuntimeBinding)
	}
	for _, module := range nihaixia.RuleModules {
		if module.Status != "draft_unimplemented" {
			t.Errorf("nihaixia module %s status = %s, want draft_unimplemented", module.RuleID, module.Status)
		}
	}

	entries, err := os.ReadDir(filepath.Join(calculationProfileRepoRoot, "library"))
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(strings.ToLower(name), ".pdf") &&
			(strings.Contains(name, "紫微") || strings.Contains(name, "天纪") || strings.Contains(name, "倪海")) {
			t.Errorf("new ZiWei source %q requires manifest evidence review before source_missing can remain", name)
		}
	}
}

func TestCalculationProfileManifestResearchPlansAreSynchronized(t *testing.T) {
	marker := "第一百三十九项完成三份非运行时流派 Profile 清单治理"
	for _, path := range []string{
		"docs/fortune-accuracy-research-plan.md",
		"docs/fortune-accuracy-roadmap.md",
		"docs/precision-test-plan.md",
	} {
		raw, err := os.ReadFile(filepath.Join(calculationProfileRepoRoot, path))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(raw), marker) != 1 {
			t.Errorf("research document %s must contain exactly one phase-139 marker", path)
		}
		for _, fragment := range []string{
			"ziping-fuyi-v2", "ziwei-sanhe-v1", "ziwei-nihaixia-v1",
			"fortune_calculation_profile_manifest_v1", "不得发布准确率",
		} {
			if !strings.Contains(string(raw), fragment) {
				t.Errorf("research document %s missing %q", path, fragment)
			}
		}
	}
}

func validateCalculationProfileManifest(t testing.TB, manifest calculationProfileManifest) {
	t.Helper()
	if manifest.Schema != calculationProfileManifestSchema || manifest.ManifestVersion == "" || manifest.Title == "" ||
		manifest.RuntimeSelectable || manifest.ActivationPolicy != "explicit_runtime_registration_after_gold_gate_only" ||
		manifest.RuntimeBinding.RegisteredID != "" || manifest.RuntimeBinding.Kind == "" || manifest.RuntimeBinding.EngineVersion == "" ||
		manifest.RuntimeBinding.RuleVersion == "" || manifest.RuntimeBinding.School == "" || manifest.RuntimeBinding.EquivalenceStatus == "" {
		t.Errorf("manifest %s has invalid identity or runtime boundary: %+v", manifest.ProfileID, manifest.RuntimeBinding)
	}
	if len(manifest.CalculationScope) < 8 || hasEmptyOrDuplicate(manifest.CalculationScope) {
		t.Errorf("manifest %s has incomplete or duplicate calculation scope: %v", manifest.ProfileID, manifest.CalculationScope)
	}

	evidenceByID := make(map[string]calculationProfileSourceEvidence, len(manifest.SourceEvidence))
	for _, evidence := range manifest.SourceEvidence {
		if evidence.ID == "" || evidenceByID[evidence.ID].ID != "" || evidence.Tier == "" || evidence.Kind == "" ||
			evidence.LicenseScope == "" || evidence.ReviewStatus == "" || evidence.CaseCount < 0 || len(evidence.Limitations) == 0 {
			t.Errorf("manifest %s has invalid evidence: %+v", manifest.ProfileID, evidence)
			continue
		}
		if evidence.Kind == "evidence_gap" {
			if evidence.SHA256 != "" || evidence.ArtifactPath != "" || evidence.CaseCount != 0 || evidence.Tier != "not_available" {
				t.Errorf("manifest %s evidence gap %s must not fabricate an artifact", manifest.ProfileID, evidence.ID)
			}
		} else if !validLowerSHA256(evidence.SHA256) {
			t.Errorf("manifest %s evidence %s has invalid SHA-256 %q", manifest.ProfileID, evidence.ID, evidence.SHA256)
		}
		evidenceByID[evidence.ID] = evidence
	}
	if len(evidenceByID) == 0 {
		t.Errorf("manifest %s has no evidence registry", manifest.ProfileID)
	}

	allowedStatuses := map[string]bool{
		"implemented_current_engine":                            true,
		"implemented_candidate_only":                            true,
		"reference_implementation_only_not_profile_adjudicated": true,
		"draft_unimplemented":                                   true,
	}
	seenRules := make(map[string]bool, len(manifest.RuleModules))
	for _, module := range manifest.RuleModules {
		if module.RuleID == "" || seenRules[module.RuleID] || !allowedStatuses[module.Status] || module.Scope == "" ||
			module.EvidenceTier == "" || len(module.EvidenceRefs) == 0 || hasEmptyOrDuplicate(module.EvidenceRefs) || len(module.Limitations) == 0 {
			t.Errorf("manifest %s has invalid rule module: %+v", manifest.ProfileID, module)
		}
		seenRules[module.RuleID] = true
		for _, evidenceID := range module.EvidenceRefs {
			if evidenceByID[evidenceID].ID == "" {
				t.Errorf("manifest %s module %s references unknown evidence %s", manifest.ProfileID, module.RuleID, evidenceID)
			}
		}
		if module.Status != "draft_unimplemented" && module.EvidenceTier == "not_available" {
			t.Errorf("manifest %s module %s claims implementation without evidence", manifest.ProfileID, module.RuleID)
		}
	}
	if len(seenRules) < 8 || len(manifest.UnsupportedRules) == 0 || len(manifest.Disputed) == 0 {
		t.Errorf("manifest %s lacks rule, unsupported, or dispute coverage", manifest.ProfileID)
	}
	for _, unsupported := range manifest.UnsupportedRules {
		if unsupported.RuleID == "" || unsupported.Reason == "" || unsupported.RequiredEvidence == "" {
			t.Errorf("manifest %s has invalid unsupported rule: %+v", manifest.ProfileID, unsupported)
		}
	}
	for _, disputed := range manifest.Disputed {
		if disputed.ID == "" || disputed.Question == "" || len(disputed.Options) == 0 || disputed.Selected == "" ||
			disputed.Status == "" || disputed.ResolutionGate == "" || hasEmptyOrDuplicate(disputed.Options) {
			t.Errorf("manifest %s has invalid disputed convention: %+v", manifest.ProfileID, disputed)
		}
	}

	if len(manifest.GoldGates) == 0 {
		t.Errorf("manifest %s has no Gold gate", manifest.ProfileID)
	}
	for _, gate := range manifest.GoldGates {
		if gate.GateID == "" || gate.MinimumCases <= 0 || gate.CurrentQualifiedCases != 0 || gate.IndependentReviewers != 2 ||
			!gate.RequiresThirdReviewerForDisputes || !gate.FrozenTestRequired || !strings.HasPrefix(gate.Status, "blocked_missing_") || gate.AdmissionScope == "" {
			t.Errorf("manifest %s has invalid Gold gate: %+v", manifest.ProfileID, gate)
		}
	}

	policy := manifest.MixingPolicy
	if policy.Fallback != "prohibited" || policy.CrossProfileImport != "prohibited_without_new_profile_id" ||
		policy.SilentDefault != "prohibited" || policy.MissingRuleBehavior == "" || policy.InterpretationMixing == "" {
		t.Errorf("manifest %s has unsafe mixing policy: %+v", manifest.ProfileID, policy)
	}
	boundary := manifest.ClaimBoundary
	if boundary.PublishableAccuracy || boundary.TraditionalCorrectnessProven || boundary.PredictiveValidityProven ||
		boundary.EventProbabilityAllowed || len(boundary.AllowedClaims) == 0 || len(boundary.ProhibitedClaims) == 0 {
		t.Errorf("manifest %s overstates its claim boundary: %+v", manifest.ProfileID, boundary)
	}
}

func loadCalculationProfileManifest(t testing.TB, path string) (calculationProfileManifest, []byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(calculationProfileRepoRoot, path))
	if err != nil {
		t.Fatal(err)
	}
	var manifest calculationProfileManifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&manifest); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("%s contains trailing JSON content: %v", path, err)
	}
	return manifest, raw
}

func manifestHasModuleStatus(manifest calculationProfileManifest, status string) bool {
	for _, module := range manifest.RuleModules {
		if module.Status == status {
			return true
		}
	}
	return false
}

func manifestHasEvidence(manifest calculationProfileManifest, id, reviewStatus string, cases int) bool {
	for _, evidence := range manifest.SourceEvidence {
		if evidence.ID == id {
			return evidence.ReviewStatus == reviewStatus && evidence.CaseCount == cases
		}
	}
	return false
}

func hasEmptyOrDuplicate(values []string) bool {
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) == "" || seen[value] {
			return true
		}
		seen[value] = true
	}
	return false
}

func validLowerSHA256(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func sha256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
