package bazi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const (
	claimEvidenceCitationSchemaPath   = "../../../../research/schemas/claim-evidence-citation-v1.schema.json"
	claimSupportReviewQueuePath       = "../../../../research/annotations/claim-support-review-candidates-v1.json"
	claimEvidenceCitationSchemaSHA256 = "9c95bcb39dc518d2fcc26c9be8bc75111621d47fead79997da521fe45a16bbd5"
	claimSupportReviewQueueSHA256     = "2b6d353cc10f4111c9ac4ba2cad6eda2f87579af96ac8c3c560ec56f19ab012e"
)

type claimSupportReviewQueue struct {
	Schema            string                            `json:"schema"`
	Version           string                            `json:"version"`
	ClaimSchemaPath   string                            `json:"claim_schema_path"`
	ClaimSchemaSHA256 string                            `json:"claim_schema_sha256"`
	Description       string                            `json:"description"`
	SelectionPolicy   claimSupportReviewSelectionPolicy `json:"selection_policy"`
	ReviewProtocol    claimSupportReviewProtocol        `json:"review_protocol"`
	SourceFiles       []claimSupportReviewSourceFile    `json:"source_files"`
	Items             []claimSupportReviewCandidate     `json:"items"`
}

type claimSupportReviewSelectionPolicy struct {
	FrozenCount        int                         `json:"frozen_count"`
	SelectionMethod    string                      `json:"selection_method"`
	PrefilledDecisions bool                        `json:"prefilled_decisions"`
	Systems            map[string]int              `json:"systems"`
	Strata             []claimSupportReviewStratum `json:"strata"`
	Inclusion          []string                    `json:"inclusion"`
	Exclusion          []string                    `json:"exclusion"`
}

type claimSupportReviewStratum struct {
	ID    string `json:"id"`
	Count int    `json:"count"`
}

type claimSupportReviewProtocol struct {
	Decisions                               []string `json:"decisions"`
	RequiredIndependentReviewers            int      `json:"required_independent_reviewers"`
	BlindedToOtherReviews                   bool     `json:"blinded_to_other_reviews"`
	ThirdReviewerForDisagreement            bool     `json:"third_reviewer_for_disagreement"`
	PrimarySourceCitationRequiredForSupport bool     `json:"primary_source_citation_required_for_support"`
	SourceCodeIsNotSupportingEvidence       bool     `json:"source_code_is_not_supporting_evidence"`
	GoldPromotionRequiresFrozenAdjudication bool     `json:"gold_promotion_requires_frozen_adjudication"`
	PublishableAccuracy                     bool     `json:"publishable_accuracy"`
}

type claimSupportReviewSourceFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Role   string `json:"role"`
}

type claimSupportReviewCandidate struct {
	CandidateID          string                          `json:"candidate_id"`
	Stratum              string                          `json:"stratum"`
	System               string                          `json:"system"`
	ProposedProfileID    string                          `json:"proposed_profile_id"`
	ProfileBindingStatus string                          `json:"profile_binding_status"`
	ClaimType            string                          `json:"claim_type"`
	ClaimText            string                          `json:"claim_text"`
	SourceLocator        claimSupportReviewSourceLocator `json:"source_locator"`
	EvidenceStatus       string                          `json:"evidence_status"`
	SensitiveDomains     []string                        `json:"sensitive_domains"`
	RequiredChecks       []string                        `json:"required_checks"`
	Review               claimSupportReviewState         `json:"review"`
}

type claimSupportReviewSourceLocator struct {
	Path       string `json:"path"`
	FileSHA256 string `json:"file_sha256"`
	Symbol     string `json:"symbol"`
	Key        string `json:"key"`
	Field      string `json:"field"`
	Line       int    `json:"line"`
}

type claimSupportReviewState struct {
	Status       string            `json:"status"`
	Reviews      []json.RawMessage `json:"reviews"`
	Adjudication json.RawMessage   `json:"adjudication"`
	GoldEligible bool              `json:"gold_eligible"`
}

func TestClaimEvidenceCitationSchemaContract(t *testing.T) {
	raw, err := os.ReadFile(claimEvidenceCitationSchemaPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := claimContractSHA256(raw); got != claimEvidenceCitationSchemaSHA256 {
		t.Fatalf("Claim/Evidence/Citation schema SHA-256 = %s, want %s", got, claimEvidenceCitationSchemaSHA256)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	if schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" ||
		schema["$id"] != "urn:bazi:schema:claim-evidence-citation:v1" || schema["type"] != "object" || schema["additionalProperties"] != false {
		t.Fatalf("invalid schema identity: $schema=%v $id=%v type=%v additional=%v", schema["$schema"], schema["$id"], schema["type"], schema["additionalProperties"])
	}

	required := claimStringSet(t, schema["required"])
	for _, field := range []string{
		"schema", "claim_id", "system", "profile_id", "engine_version", "rule_version", "input_fingerprint",
		"claim_text", "claim_type", "claim_status", "confidence", "evidence", "counter_evidence", "citations",
		"sensitive_domains", "provenance", "review", "limitations",
	} {
		if !required[field] {
			t.Errorf("claim schema does not require %q", field)
		}
	}

	defs := claimObject(t, schema["$defs"])
	for _, name := range []string{"sha256", "confidence", "confidence_component", "evidence", "citation", "provenance", "review", "reviewer_decision", "adjudication"} {
		if _, ok := defs[name]; !ok {
			t.Errorf("claim schema missing $defs.%s", name)
		}
	}
	confidence := claimProperties(t, defs["confidence"])
	if claimObject(t, confidence["event_probability"])["const"] != false {
		t.Error("confidence must explicitly prohibit event-probability semantics")
	}
	reviewer := claimProperties(t, defs["reviewer_decision"])
	if claimObject(t, reviewer["blinded_to_other_reviews"])["const"] != true {
		t.Error("reviewer decisions must be blinded to other reviews")
	}
	decisionEnums := claimStringSet(t, claimObject(t, reviewer["decision"])["enum"])
	for _, decision := range []string{"supported", "refuted", "no_evidence", "over_inference", "sensitive_risk"} {
		if !decisionEnums[decision] {
			t.Errorf("review schema missing decision %q", decision)
		}
	}

	citationRequired := claimStringSet(t, claimObject(t, defs["citation"])["required"])
	for _, field := range []string{"artifact_path", "artifact_sha256", "quote", "quote_sha256", "page", "locator", "verification_status"} {
		if !citationRequired[field] {
			t.Errorf("citation schema does not require %q", field)
		}
	}
	for _, fragment := range []string{
		`"const": "outcome_claim"`, `"const": "adjudicated"`, `"const": "insufficient_evidence"`,
		`"const": "not_scored"`, `"minItems": 2`, `"gold_eligible"`,
	} {
		if !bytes.Contains(raw, []byte(fragment)) {
			t.Errorf("claim schema missing safety condition %s", fragment)
		}
	}
}

func TestClaimSupportReviewQueueFreezesFiftyUnreviewedCandidates(t *testing.T) {
	queue, raw := loadClaimSupportReviewQueue(t)
	if got := claimContractSHA256(raw); got != claimSupportReviewQueueSHA256 {
		t.Fatalf("claim review queue SHA-256 = %s, want %s", got, claimSupportReviewQueueSHA256)
	}
	if queue.Schema != "claim_support_review_queue_v1" || queue.Version != "2026-07-17.1" ||
		queue.ClaimSchemaPath != "research/schemas/claim-evidence-citation-v1.schema.json" ||
		queue.ClaimSchemaSHA256 != claimEvidenceCitationSchemaSHA256 || queue.SelectionPolicy.FrozenCount != 50 || len(queue.Items) != 50 {
		t.Fatalf("claim review queue identity/count changed: schema=%s version=%s schema_ref=%s/%s frozen=%d items=%d",
			queue.Schema, queue.Version, queue.ClaimSchemaPath, queue.ClaimSchemaSHA256, queue.SelectionPolicy.FrozenCount, len(queue.Items))
	}
	if queue.SelectionPolicy.SelectionMethod != "deterministic_stratified_source_order_v1" || queue.SelectionPolicy.PrefilledDecisions ||
		!reflect.DeepEqual(queue.SelectionPolicy.Systems, map[string]int{"bazi": 30, "ziwei": 20}) ||
		len(queue.SelectionPolicy.Inclusion) == 0 || len(queue.SelectionPolicy.Exclusion) == 0 {
		t.Errorf("claim review selection policy is incomplete: %+v", queue.SelectionPolicy)
	}

	wantStrata := map[string]int{
		"bazi_ten_god_positive": 10, "bazi_ten_god_negative": 10, "bazi_ten_god_advice": 10,
		"ziwei_palace_context": 12, "ziwei_four_hua_theme": 4, "ziwei_period_boundary": 4,
	}
	declaredStrata := make(map[string]int, len(queue.SelectionPolicy.Strata))
	for _, stratum := range queue.SelectionPolicy.Strata {
		if stratum.ID == "" || stratum.Count <= 0 || declaredStrata[stratum.ID] != 0 {
			t.Errorf("invalid declared claim-review stratum: %+v", stratum)
		}
		declaredStrata[stratum.ID] = stratum.Count
	}
	if !reflect.DeepEqual(declaredStrata, wantStrata) {
		t.Errorf("declared claim-review strata = %+v, want %+v", declaredStrata, wantStrata)
	}

	protocol := queue.ReviewProtocol
	if !reflect.DeepEqual(protocol.Decisions, []string{"supported", "refuted", "no_evidence", "over_inference", "sensitive_risk"}) ||
		protocol.RequiredIndependentReviewers != 2 || !protocol.BlindedToOtherReviews || !protocol.ThirdReviewerForDisagreement ||
		!protocol.PrimarySourceCitationRequiredForSupport || !protocol.SourceCodeIsNotSupportingEvidence ||
		!protocol.GoldPromotionRequiresFrozenAdjudication || protocol.PublishableAccuracy {
		t.Errorf("claim-review protocol permits self-validation or premature publication: %+v", protocol)
	}
}

func TestClaimSupportReviewCandidatesTraceExactlyToCurrentSource(t *testing.T) {
	queue, _ := loadClaimSupportReviewQueue(t)
	sourceFiles := make(map[string]claimSupportReviewSourceFile, len(queue.SourceFiles))
	sourceLines := make(map[string][]string, len(queue.SourceFiles))
	for _, source := range queue.SourceFiles {
		if source.Path == "" || source.Role == "" || !claimContractValidSHA256(source.SHA256) || sourceFiles[source.Path].Path != "" {
			t.Errorf("invalid or duplicate claim-review source file: %+v", source)
			continue
		}
		raw, err := os.ReadFile(filepath.Join("../../../..", source.Path))
		if err != nil {
			t.Errorf("source %s: %v", source.Path, err)
			continue
		}
		if got := claimContractSHA256(raw); got != source.SHA256 {
			t.Errorf("source %s SHA-256 = %s, want %s", source.Path, got, source.SHA256)
		}
		sourceFiles[source.Path] = source
		sourceLines[source.Path] = strings.Split(string(raw), "\n")
	}
	if len(sourceFiles) != 3 {
		t.Fatalf("claim-review source files = %d, want 3", len(sourceFiles))
	}

	wantChecks := []string{"source_support", "claim_scope", "over_inference", "sensitive_risk", "citation_locatability"}
	allowedTypes := map[string]bool{"school_dependent_interpretation": true, "advisory": true, "structural_rule": true}
	allowedDomains := map[string]bool{
		"personality": true, "relationship": true, "health": true, "lifespan": true, "finance": true,
		"career": true, "legal": true, "violence_or_self_harm": true, "minor_or_parenting": true,
		"housing_or_property": true, "none": true,
	}
	seenIDs := make(map[string]bool, len(queue.Items))
	seenClaims := make(map[string]bool, len(queue.Items))
	actualStrata := make(map[string]int)
	actualSystems := make(map[string]int)
	for index, item := range queue.Items {
		wantID := fmt.Sprintf("support-review-%03d", index+1)
		if item.CandidateID != wantID || seenIDs[item.CandidateID] || strings.TrimSpace(item.ClaimText) == "" || seenClaims[item.ClaimText] {
			t.Errorf("invalid candidate identity/text at index %d: id=%q text=%q", index, item.CandidateID, item.ClaimText)
		}
		seenIDs[item.CandidateID] = true
		seenClaims[item.ClaimText] = true
		actualStrata[item.Stratum]++
		actualSystems[item.System]++

		if !allowedTypes[item.ClaimType] || !reflect.DeepEqual(item.RequiredChecks, wantChecks) || len(item.SensitiveDomains) == 0 {
			t.Errorf("candidate %s has incomplete claim/review fields: type=%s domains=%v checks=%v", item.CandidateID, item.ClaimType, item.SensitiveDomains, item.RequiredChecks)
		}
		seenDomains := map[string]bool{}
		for _, domain := range item.SensitiveDomains {
			if !allowedDomains[domain] || seenDomains[domain] || (domain == "none" && len(item.SensitiveDomains) != 1) {
				t.Errorf("candidate %s has invalid sensitive domain %q in %v", item.CandidateID, domain, item.SensitiveDomains)
			}
			seenDomains[domain] = true
		}
		if item.Review.Status != "unreviewed" || len(item.Review.Reviews) != 0 || string(item.Review.Adjudication) != "null" || item.Review.GoldEligible {
			t.Errorf("candidate %s contains a prefilled or Gold review: %+v", item.CandidateID, item.Review)
		}
		if item.System == "bazi" {
			if item.ProposedProfileID != "ziping-fuyi-v2" || item.ProfileBindingStatus != "proposed_not_adjudicated" ||
				item.EvidenceStatus != "source_code_only_no_primary_citation" {
				t.Errorf("candidate %s overstates BaZi profile/evidence binding", item.CandidateID)
			}
		} else if item.System == "ziwei" {
			if item.ProposedProfileID != "ziwei-local-composite-v2" || item.ProfileBindingStatus != "existing_runtime_not_expert_adjudicated" ||
				(item.EvidenceStatus != "runtime_template_only_no_primary_citation" && item.EvidenceStatus != "runtime_source_contract_not_gold") {
				t.Errorf("candidate %s overstates ZiWei profile/evidence binding", item.CandidateID)
			}
		} else {
			t.Errorf("candidate %s has unknown system %q", item.CandidateID, item.System)
		}

		locator := item.SourceLocator
		source := sourceFiles[locator.Path]
		lines := sourceLines[locator.Path]
		if source.Path == "" || locator.FileSHA256 != source.SHA256 || locator.Symbol == "" || locator.Key == "" || locator.Field == "" ||
			locator.Line <= 0 || locator.Line > len(lines) {
			t.Errorf("candidate %s has invalid source locator: %+v", item.CandidateID, locator)
			continue
		}
		if !strings.Contains(lines[locator.Line-1], item.ClaimText) {
			t.Errorf("candidate %s text %q is not present at %s:%d: %q", item.CandidateID, item.ClaimText, locator.Path, locator.Line, lines[locator.Line-1])
		}
	}
	if !reflect.DeepEqual(actualStrata, map[string]int{
		"bazi_ten_god_positive": 10, "bazi_ten_god_negative": 10, "bazi_ten_god_advice": 10,
		"ziwei_palace_context": 12, "ziwei_four_hua_theme": 4, "ziwei_period_boundary": 4,
	}) || !reflect.DeepEqual(actualSystems, map[string]int{"bazi": 30, "ziwei": 20}) {
		t.Errorf("actual queue distribution changed: systems=%+v strata=%+v", actualSystems, actualStrata)
	}
}

func TestClaimEvidenceCitationResearchPlansAreSynchronized(t *testing.T) {
	marker := "第一百四十项完成 Claim/Evidence/Citation Schema 与 50 条人工审阅队列治理"
	for _, path := range []string{
		"../../../../docs/fortune-accuracy-research-plan.md",
		"../../../../docs/fortune-accuracy-roadmap.md",
		"../../../../docs/precision-test-plan.md",
	} {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Count(string(raw), marker) != 1 {
			t.Errorf("research document %s must contain exactly one phase-140 marker", path)
		}
		for _, fragment := range []string{
			"claim_evidence_citation_v1", "claim_support_review_queue_v1", "30 条八字", "20 条紫微",
			"unreviewed", "source code", "不得发布准确率",
		} {
			if !strings.Contains(string(raw), fragment) {
				t.Errorf("research document %s missing %q", path, fragment)
			}
		}
	}
}

func loadClaimSupportReviewQueue(t testing.TB) (claimSupportReviewQueue, []byte) {
	t.Helper()
	raw, err := os.ReadFile(claimSupportReviewQueuePath)
	if err != nil {
		t.Fatal(err)
	}
	var queue claimSupportReviewQueue
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&queue); err != nil {
		t.Fatal(err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("claim support review queue contains trailing JSON: %v", err)
	}
	return queue, raw
}

func claimObject(t testing.TB, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("schema value is %T, want object", value)
	}
	return object
}

func claimProperties(t testing.TB, value any) map[string]any {
	t.Helper()
	return claimObject(t, claimObject(t, value)["properties"])
}

func claimStringSet(t testing.TB, value any) map[string]bool {
	t.Helper()
	values, ok := value.([]any)
	if !ok {
		t.Fatalf("schema value is %T, want array", value)
	}
	result := make(map[string]bool, len(values))
	for _, item := range values {
		text, ok := item.(string)
		if !ok || text == "" || result[text] {
			t.Fatalf("invalid schema string array item %#v", item)
		}
		result[text] = true
	}
	return result
}

func claimContractValidSHA256(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func claimContractSHA256(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
