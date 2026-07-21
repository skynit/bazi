package bazi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

const (
	ragNCLPageReviewProtocolPath    = "../../../../research/annotations/sanming-ncl-page-mapping-review-v1.json"
	ragNCLPageReviewProtocolSHA256  = "d0bafac669f4ca92e498f3a087738452b7e1417b9ac0be3307cebade935ca977"
	ragNCLPageReviewSchemaPath      = "../../../../research/schemas/sanming-ncl-page-review-submission-v1.schema.json"
	ragNCLPageReviewSchemaSHA256    = "e2a84f7ff40d15d3a6404ad11901178a3ada1ce10df0f29aa603d36cb899dcf4"
	ragNCLPageReviewGeneratorSHA256 = "02a9da7d62cc26dcddb11fdcb96f14f2e1db3946196e0fd7d00e1f7ff6e6c69f"
)

type ragNCLPageReviewProtocol struct {
	Schema      string `json:"schema"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	ObservedAt  string `json:"observed_at"`
	ProtocolID  string `json:"protocol_id"`
	CandidateID string `json:"candidate_id"`
	Purpose     string `json:"purpose"`
	Sources     struct {
		PageCandidates   ragCommonsArtifactReference `json:"page_candidates"`
		VolumeLabels     ragCommonsArtifactReference `json:"volume_labels"`
		SubmissionSchema ragCommonsArtifactReference `json:"submission_schema"`
	} `json:"sources"`
	Blinding     ragNCLPageReviewBlinding     `json:"blinding"`
	Stages       ragNCLPageReviewStages       `json:"stages"`
	Adjudication ragNCLPageReviewAdjudication `json:"adjudication"`
	Quality      ragNCLPageReviewQuality      `json:"quality_control"`
	Gates        ragNCLPageReviewGates        `json:"gates"`
	ReviewState  ragNCLPageReviewState        `json:"review_state"`
	Summary      ragNCLPageReviewSummary      `json:"summary"`
	VolumeItems  []ragNCLVolumeReviewItem     `json:"volume_items"`
	ChapterItems []ragNCLChapterReviewItem    `json:"chapter_items"`
	Boundaries   ragNCLPageReviewBoundaries   `json:"boundaries"`
}

type ragNCLPageReviewBlinding struct {
	ReviewerSlots                    []string `json:"reviewer_slots"`
	RequiredIndependentReviewers     int      `json:"required_independent_reviewers"`
	RequiredAdjudicators             int      `json:"required_adjudicators"`
	ReviewerIdentityAssignment       string   `json:"reviewer_identity_assignment"`
	SeparateSubmissionArtifacts      bool     `json:"separate_submission_artifacts_required"`
	CrossReviewerVisibility          string   `json:"cross_reviewer_visibility"`
	MachineSignalVisibility          string   `json:"machine_signal_visibility"`
	ReviewerPacketVisibleFields      []string `json:"reviewer_packet_visible_fields"`
	ReviewerPacketForbiddenFields    []string `json:"reviewer_packet_forbidden_fields"`
	PriorVolumeOperatorMayReview     bool     `json:"prior_volume_label_operator_may_review"`
	ReviewerAndAdjudicatorMayOverlap bool     `json:"reviewer_and_adjudicator_roles_may_overlap"`
	RealIdentityStored               bool     `json:"real_identity_stored_in_research_artifact"`
}

type ragNCLPageReviewStage struct {
	Status              string   `json:"status"`
	ItemCount           int      `json:"item_count"`
	ReviewUnit          string   `json:"review_unit"`
	AllowedDecisions    []string `json:"allowed_decisions"`
	RequiredFields      []string `json:"required_fields"`
	ReleaseCondition    string   `json:"release_condition"`
	GoldCondition       string   `json:"gold_condition"`
	ClaimSupportInScope bool     `json:"claim_support_in_scope"`
}

type ragNCLPageReviewStages struct {
	VolumeLabelReview     ragNCLPageReviewStage `json:"volume_label_review"`
	ChapterBoundaryReview ragNCLPageReviewStage `json:"chapter_boundary_review"`
}

type ragNCLPageReviewAdjudication struct {
	Trigger                          string   `json:"trigger"`
	ComparisonFields                 []string `json:"comparison_fields"`
	MajorityVoteAllowed              bool     `json:"majority_vote_allowed"`
	AdjudicatorSeesPseudonyms        bool     `json:"adjudicator_sees_reviewer_pseudonyms"`
	AdjudicatorSeesMachineCandidates bool     `json:"adjudicator_sees_machine_candidates"`
	RequiredResolutionFields         []string `json:"required_resolution_fields"`
	UnresolvedDecision               string   `json:"unresolved_decision"`
}

type ragNCLPageReviewQuality struct {
	AgreementAuditCount   int      `json:"agreement_audit_count"`
	AgreementAuditRateBP  int      `json:"agreement_audit_rate_basis_points"`
	SelectionScheme       string   `json:"agreement_audit_selection_scheme"`
	SelectionSeed         string   `json:"agreement_audit_selection_seed"`
	SelectedItemIDs       []string `json:"agreement_audit_selected_item_ids"`
	AuditReviewerCount    int      `json:"audit_reviewer_count"`
	AgreementCanSkipAudit bool     `json:"agreement_can_skip_prespecified_audit"`
	FailedAuditAction     string   `json:"failed_audit_action"`
}

type ragNCLPageReviewGates struct {
	VolumeReviewReleased           bool `json:"volume_review_released"`
	VolumeMappingAdjudicated       bool `json:"volume_mapping_adjudicated"`
	ChapterReviewReleased          bool `json:"chapter_review_released"`
	AllChapterReviewsComplete      bool `json:"all_chapter_reviews_complete"`
	AgreementAuditComplete         bool `json:"agreement_audit_complete"`
	ChapterPageMappingGoldReleased bool `json:"chapter_page_mapping_gold_released"`
}

type ragNCLPageReviewState struct {
	ReviewerAssignmentsCreated   int `json:"reviewer_assignments_created"`
	IndependentReviewersAssigned int `json:"independent_reviewers_assigned"`
	ReviewerSubmissionsReceived  int `json:"reviewer_submissions_received"`
	AdjudicatorsAssigned         int `json:"adjudicators_assigned"`
	AdjudicationsComplete        int `json:"adjudications_complete"`
	AgreementAuditsComplete      int `json:"agreement_audits_complete"`
	GoldItems                    int `json:"gold_items"`
}

type ragNCLPageReviewSummary struct {
	VolumeItems                  int `json:"volume_items"`
	ChapterItems                 int `json:"chapter_items"`
	CriticalZeroOverlap          int `json:"critical_zero_overlap_chapters"`
	HighTieOrLowMargin           int `json:"high_tie_or_low_margin_chapters"`
	HighTitleContentDisagreement int `json:"high_title_content_disagreement_chapters"`
	Standard                     int `json:"standard_chapters"`
	ReadyItems                   int `json:"ready_items"`
	BlockedItems                 int `json:"blocked_items"`
	ReviewDecisionsPresent       int `json:"review_decisions_present"`
}

type ragNCLVolumeReviewItem struct {
	ItemID                      string                  `json:"item_id"`
	VolumeCandidate             int                     `json:"volume_candidate"`
	CoordinatorCandidateHeading string                  `json:"coordinator_candidate_heading"`
	HeadingPage                 ragNCLPageKey           `json:"heading_page"`
	PhysicalBookSegments        []ragCommonsBookSegment `json:"physical_book_segments"`
	CommonsPageLocator          string                  `json:"commons_page_locator"`
	Evidence                    struct {
		Path      string `json:"path"`
		SizeBytes int64  `json:"size_bytes"`
		SHA256    string `json:"sha256"`
	} `json:"evidence"`
	MachineCandidateVisible bool   `json:"machine_candidate_visible_to_reviewer"`
	ReviewStatus            string `json:"review_status"`
}

type ragNCLReviewMachineSignal struct {
	TitleLocatorCandidateCount int             `json:"title_locator_candidate_count"`
	TitleLocatorCandidates     []ragNCLPageKey `json:"title_locator_candidates"`
	BestContentOverlap         int             `json:"best_content_overlap"`
	ContentOverlapMargin       int             `json:"content_overlap_margin"`
	BestContentCandidateCount  int             `json:"best_content_candidate_count"`
	BestContentCandidates      []ragNCLPageKey `json:"best_content_candidates"`
	ZeroContentOverlap         bool            `json:"zero_content_overlap"`
}

type ragNCLChapterReviewItem struct {
	ItemID                  string                    `json:"item_id"`
	Chapter                 int                       `json:"chapter"`
	File                    string                    `json:"file"`
	Title                   string                    `json:"title"`
	OriginalSHA256          string                    `json:"original_sha256"`
	VolumeCandidate         int                       `json:"volume_candidate"`
	PhysicalBookSegments    []ragCommonsBookSegment   `json:"physical_book_segments"`
	PageLocatorPatterns     []string                  `json:"page_locator_patterns"`
	CoordinatorPriorityRank int                       `json:"coordinator_priority_rank"`
	RiskStratum             string                    `json:"coordinator_risk_stratum"`
	RiskReasons             []string                  `json:"coordinator_risk_reasons"`
	AgreementAuditSelected  bool                      `json:"agreement_audit_selected"`
	MachineSignalVisibility string                    `json:"machine_signal_visibility"`
	MachineSignal           ragNCLReviewMachineSignal `json:"machine_signal"`
	ReviewStatus            string                    `json:"review_status"`
}

type ragNCLPageReviewBoundaries struct {
	ProtocolOnly                bool `json:"protocol_only"`
	HumanReviewStarted          bool `json:"human_review_started"`
	IndependentReviewComplete   bool `json:"independent_review_complete"`
	BibliographyAdjudicated     bool `json:"bibliography_adjudicated"`
	CompletePrimaryTextVerified bool `json:"complete_primary_text_verified"`
	VolumeMappingVerified       bool `json:"volume_mapping_verified"`
	ChapterPageMappingVerified  bool `json:"chapter_page_mapping_verified"`
	ClaimSupportReviewed        bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed     bool `json:"runtime_ingestion_allowed"`
	ClaimEligible               bool `json:"claim_eligible"`
	PublishableAccuracy         bool `json:"publishable_accuracy"`
}

func TestRAGNCLPageReviewSubmissionSchemaContract(t *testing.T) {
	raw, err := os.ReadFile(ragNCLPageReviewSchemaPath)
	if err != nil {
		t.Fatal(err)
	}
	if got := ragCommonsSHA256(raw); got != ragNCLPageReviewSchemaSHA256 {
		t.Fatalf("page-review submission schema SHA-256 = %s, want %s", got, ragNCLPageReviewSchemaSHA256)
	}
	var schema map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&schema); err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{
		`"additionalProperties": false`, `"sanming_ncl_page_mapping_review_submission_v1"`,
		`"worked_independently"`, `"other_reviewer_decisions_unseen"`, `"machine_page_candidates_unseen"`,
		`"no_prior_volume_label_operator_role"`, `"chapter_boundary_located"`, `"chapter_not_located"`,
		`"volume_heading_confirmed"`, `"volume_heading_differs"`, `"scan_incomplete_or_illegible"`,
	} {
		if !bytes.Contains(raw, []byte(marker)) {
			t.Fatalf("page-review submission schema missing %q", marker)
		}
	}
	for _, forbidden := range []string{"reviewer_real_name", "claim_supported", "prediction_accurate", "fortune_outcome"} {
		if bytes.Contains(raw, []byte(forbidden)) {
			t.Fatalf("page-review submission schema contains forbidden field %q", forbidden)
		}
	}
}

func TestRAGNCLPageReviewProtocolContract(t *testing.T) {
	protocol := ragCommonsReadStrictJSON[ragNCLPageReviewProtocol](t, ragNCLPageReviewProtocolPath, ragNCLPageReviewProtocolSHA256)
	if protocol.Schema != "sanming_ncl_page_mapping_review_protocol_v1" || protocol.Version != "2026-07-17.1" ||
		protocol.Status != "protocol_frozen_reviews_not_started" || protocol.ObservedAt != "2026-07-17" ||
		protocol.ProtocolID != "sanming-ncl-page-mapping-review-v1" || protocol.CandidateID != ragCommonsCandidateID || strings.TrimSpace(protocol.Purpose) == "" {
		t.Fatalf("unexpected page-review protocol identity: %+v", protocol)
	}
	if protocol.Sources.PageCandidates != (ragCommonsArtifactReference{Path: "research/rag/sanming-ncl-06589-chapter-page-candidates-v1.json", SHA256: ragNCLPageCandidatesSHA256}) ||
		protocol.Sources.VolumeLabels != (ragCommonsArtifactReference{Path: "research/rag/snapshots/sanming-ncl-06589-1578-v1/volume-label-observation.json", SHA256: ragNCLVolumeLabelObservationSHA256}) ||
		protocol.Sources.SubmissionSchema != (ragCommonsArtifactReference{Path: "research/schemas/sanming-ncl-page-review-submission-v1.schema.json", SHA256: ragNCLPageReviewSchemaSHA256}) {
		t.Fatalf("unexpected page-review sources: %+v", protocol.Sources)
	}
	ragNCLValidateReviewGovernance(t, protocol)
	ragNCLValidateVolumeReviewItems(t, protocol.VolumeItems)
	ragNCLValidateChapterReviewItems(t, protocol)
	b := protocol.Boundaries
	if !b.ProtocolOnly || b.HumanReviewStarted || b.IndependentReviewComplete || b.BibliographyAdjudicated || b.CompletePrimaryTextVerified ||
		b.VolumeMappingVerified || b.ChapterPageMappingVerified || b.ClaimSupportReviewed || b.RuntimeIngestionAllowed || b.ClaimEligible || b.PublishableAccuracy {
		t.Fatalf("page-review protocol boundaries must remain fail-closed: %+v", b)
	}
}

func ragNCLValidateReviewGovernance(t *testing.T, protocol ragNCLPageReviewProtocol) {
	t.Helper()
	b := protocol.Blinding
	if fmt.Sprint(b.ReviewerSlots) != fmt.Sprint([]string{"reviewer_a", "reviewer_b"}) || b.RequiredIndependentReviewers != 2 || b.RequiredAdjudicators != 1 ||
		b.ReviewerIdentityAssignment != "external_coordinator_pseudonymization_not_stored_in_research_artifact" || !b.SeparateSubmissionArtifacts ||
		b.CrossReviewerVisibility != "sealed_until_both_submissions_pass_schema_and_completeness_checks" ||
		b.MachineSignalVisibility != "coordinator_only_until_both_independent_submissions_are_sealed" || b.PriorVolumeOperatorMayReview ||
		b.ReviewerAndAdjudicatorMayOverlap || b.RealIdentityStored || !containsAll(b.ReviewerPacketForbiddenFields, []string{"coordinator_priority_rank", "coordinator_risk_stratum", "coordinator_risk_reasons", "machine_signal", "other_reviewer_submission"}) {
		t.Fatalf("unexpected review blinding: %+v", b)
	}
	volume, chapter := protocol.Stages.VolumeLabelReview, protocol.Stages.ChapterBoundaryReview
	if volume.Status != "ready_unassigned" || volume.ItemCount != 12 || volume.ClaimSupportInScope ||
		chapter.Status != "blocked_by_volume_mapping_adjudication" || chapter.ItemCount != 381 || chapter.ClaimSupportInScope ||
		!strings.Contains(chapter.ReleaseCondition, "all_12_volume_items") || !strings.Contains(chapter.GoldCondition, "agreement audit") {
		t.Fatalf("unexpected gated review stages: %+v", protocol.Stages)
	}
	a := protocol.Adjudication
	if a.MajorityVoteAllowed || a.AdjudicatorSeesPseudonyms || a.AdjudicatorSeesMachineCandidates ||
		a.UnresolvedDecision != "unresolved_not_gold_eligible" || !strings.Contains(a.Trigger, "any decision") {
		t.Fatalf("unexpected adjudication protocol: %+v", a)
	}
	if protocol.Gates != (ragNCLPageReviewGates{VolumeReviewReleased: true}) || protocol.ReviewState != (ragNCLPageReviewState{}) {
		t.Fatalf("reviews must be released only at volume stage and remain unstarted: gates=%+v state=%+v", protocol.Gates, protocol.ReviewState)
	}
	wantSummary := ragNCLPageReviewSummary{VolumeItems: 12, ChapterItems: 381, CriticalZeroOverlap: 9, HighTieOrLowMargin: 53, HighTitleContentDisagreement: 21, Standard: 298, ReadyItems: 12, BlockedItems: 381}
	if protocol.Summary != wantSummary {
		t.Fatalf("unexpected review summary: got %+v want %+v", protocol.Summary, wantSummary)
	}
}

func ragNCLValidateVolumeReviewItems(t *testing.T, items []ragNCLVolumeReviewItem) {
	t.Helper()
	observation := ragCommonsReadStrictJSON[ragNCLVolumeLabelObservation](t, filepath.Join(ragCommonsSnapshotRoot, "volume-label-observation.json"), ragNCLVolumeLabelObservationSHA256)
	if len(items) != 12 {
		t.Fatalf("volume review items = %d, want 12", len(items))
	}
	for index, item := range items {
		source := observation.Observations[index]
		if item.ItemID != fmt.Sprintf("volume-%02d", index+1) || item.VolumeCandidate != index+1 ||
			item.CoordinatorCandidateHeading != source.TranscribedHeading ||
			item.HeadingPage != (ragNCLPageKey{Part: source.Source.Part, PhysicalPage: source.Source.PhysicalPage}) ||
			fmt.Sprint(item.PhysicalBookSegments) != fmt.Sprint(source.Source.PhysicalBookSegments) || item.CommonsPageLocator != source.Source.CommonsPageLocator ||
			item.MachineCandidateVisible || item.ReviewStatus != "ready_unassigned" || item.Evidence.SizeBytes != source.Evidence.SizeBytes || item.Evidence.SHA256 != source.Evidence.SHA256 {
			t.Fatalf("unexpected volume review item %d: %+v", index+1, item)
		}
		evidencePath := "../../../../" + item.Evidence.Path
		info, err := os.Stat(evidencePath)
		if err != nil || info.Size() != item.Evidence.SizeBytes || ragCommonsHashFile(t, evidencePath) != item.Evidence.SHA256 {
			t.Fatalf("volume review evidence %d mismatch: info=%v err=%v", index+1, info, err)
		}
	}
}

func ragNCLValidateChapterReviewItems(t *testing.T, protocol ragNCLPageReviewProtocol) {
	t.Helper()
	candidates := ragCommonsReadStrictJSON[ragNCLPageCandidateReport](t, ragNCLPageCandidatesPath, ragNCLPageCandidatesSHA256)
	if len(protocol.ChapterItems) != 381 {
		t.Fatalf("chapter review items = %d, want 381", len(protocol.ChapterItems))
	}
	priorityOrder := ragNCLExpectedReviewPriority(candidates.Chapters)
	priorityRank := map[int]int{}
	for index, chapter := range priorityOrder {
		priorityRank[chapter] = index + 1
	}
	seenRanks := map[int]bool{}
	seenAudit := map[string]bool{}
	strata := map[string]int{}
	for index, item := range protocol.ChapterItems {
		source := candidates.Chapters[index]
		wantStratum, wantReasons := ragNCLExpectedRisk(source)
		if item.ItemID != fmt.Sprintf("chapter-%03d", index+1) || item.Chapter != index+1 || item.File != source.File || item.Title != source.Title ||
			item.OriginalSHA256 != source.OriginalSHA256 || item.VolumeCandidate != source.VolumeCandidate || item.CoordinatorPriorityRank != priorityRank[index+1] ||
			seenRanks[item.CoordinatorPriorityRank] || item.RiskStratum != wantStratum || fmt.Sprint(item.RiskReasons) != fmt.Sprint(wantReasons) ||
			item.MachineSignalVisibility != "coordinator_only" || item.ReviewStatus != "blocked_by_volume_mapping_adjudication" {
			t.Fatalf("unexpected chapter review item %d: %+v", index+1, item)
		}
		seenRanks[item.CoordinatorPriorityRank] = true
		strata[item.RiskStratum]++
		if item.AgreementAuditSelected {
			seenAudit[item.ItemID] = true
		}
		wantSegments := ragNCLVolumeSegmentsForChapter(t, item.VolumeCandidate)
		if fmt.Sprint(item.PhysicalBookSegments) != fmt.Sprint(wantSegments) || fmt.Sprint(item.PageLocatorPatterns) != fmt.Sprint(ragNCLExpectedLocatorPatterns(wantSegments)) {
			t.Fatalf("chapter %d scan scope mismatch", item.Chapter)
		}
		m := item.MachineSignal
		if m.TitleLocatorCandidateCount != source.TitleLocatorCandidateCount || fmt.Sprint(m.TitleLocatorCandidates) != fmt.Sprint(source.TitleLocatorCandidates) ||
			m.BestContentOverlap != source.BestContentOverlap || m.ContentOverlapMargin != source.ContentOverlapMargin ||
			m.BestContentCandidateCount != source.BestContentCandidateCount || fmt.Sprint(m.BestContentCandidates) != fmt.Sprint(source.BestContentCandidates) ||
			m.ZeroContentOverlap != source.ZeroContentOverlap {
			t.Fatalf("chapter %d coordinator machine signal drifted", item.Chapter)
		}
	}
	if len(seenRanks) != 381 || fmt.Sprint(strata) != fmt.Sprint(map[string]int{"critical_zero_overlap": 9, "high_tie_or_low_margin": 53, "high_title_content_disagreement": 21, "standard": 298}) {
		t.Fatalf("priority ranks or strata mismatch: ranks=%d strata=%v", len(seenRanks), strata)
	}
	wantAudit := ragNCLExpectedAgreementAuditIDs()
	if protocol.Quality.AgreementAuditCount != 39 || protocol.Quality.AgreementAuditRateBP != 1024 ||
		protocol.Quality.SelectionScheme != "lowest_sha256_of_page_candidate_report_sha256_tab_item_id_lf_first_39" ||
		protocol.Quality.SelectionSeed != ragNCLPageCandidatesSHA256 || protocol.Quality.AuditReviewerCount != 1 || protocol.Quality.AgreementCanSkipAudit ||
		fmt.Sprint(protocol.Quality.SelectedItemIDs) != fmt.Sprint(wantAudit) || len(seenAudit) != 39 {
		t.Fatalf("unexpected agreement audit protocol: %+v", protocol.Quality)
	}
	for _, itemID := range wantAudit {
		if !seenAudit[itemID] {
			t.Fatalf("agreement audit item %s not marked in queue", itemID)
		}
	}
}

func ragNCLExpectedReviewPriority(chapters []ragNCLPageCandidateChapter) []int {
	values := make([]int, len(chapters))
	for index := range values {
		values[index] = index + 1
	}
	rank := map[string]int{"critical_zero_overlap": 0, "high_tie_or_low_margin": 1, "high_title_content_disagreement": 2, "standard": 3}
	sort.Slice(values, func(i, j int) bool {
		a, _ := ragNCLExpectedRisk(chapters[values[i]-1])
		b, _ := ragNCLExpectedRisk(chapters[values[j]-1])
		if rank[a] != rank[b] {
			return rank[a] < rank[b]
		}
		return values[i] < values[j]
	})
	return values
}

func ragNCLExpectedRisk(chapter ragNCLPageCandidateChapter) (string, []string) {
	if chapter.ZeroContentOverlap {
		return "critical_zero_overlap", []string{"zero_content_overlap"}
	}
	if chapter.BestContentCandidateCount > 1 || chapter.ContentOverlapMargin <= 2 {
		var reasons []string
		if chapter.BestContentCandidateCount > 1 {
			reasons = append(reasons, "best_content_page_tie")
		}
		if chapter.ContentOverlapMargin <= 2 {
			reasons = append(reasons, "content_overlap_margin_at_most_2")
		}
		return "high_tie_or_low_margin", reasons
	}
	content := map[ragNCLPageKey]bool{}
	for _, page := range chapter.BestContentCandidates {
		content[page] = true
	}
	disjoint := len(chapter.TitleLocatorCandidates) > 0
	for _, page := range chapter.TitleLocatorCandidates {
		if content[page] {
			disjoint = false
		}
	}
	if disjoint {
		return "high_title_content_disagreement", []string{"title_locator_and_best_content_pages_disjoint"}
	}
	return "standard", []string{"no_prespecified_machine_uncertainty_trigger"}
}

func ragNCLExpectedAgreementAuditIDs() []string {
	type record struct{ id, hash string }
	records := make([]record, 0, 381)
	for chapter := 1; chapter <= 381; chapter++ {
		id := fmt.Sprintf("chapter-%03d", chapter)
		hash := ragCommonsSHA256([]byte(ragNCLPageCandidatesSHA256 + "\t" + id + "\n"))
		records = append(records, record{id: id, hash: hash})
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].hash != records[j].hash {
			return records[i].hash < records[j].hash
		}
		return records[i].id < records[j].id
	})
	selected := make([]string, 0, 39)
	for _, item := range records[:39] {
		selected = append(selected, item.id)
	}
	sort.Strings(selected)
	return selected
}

func ragNCLVolumeSegmentsForChapter(t *testing.T, volume int) []ragCommonsBookSegment {
	t.Helper()
	observation := ragCommonsReadStrictJSON[ragNCLVolumeLabelObservation](t, filepath.Join(ragCommonsSnapshotRoot, "volume-label-observation.json"), ragNCLVolumeLabelObservationSHA256)
	return observation.Observations[volume-1].Source.PhysicalBookSegments
}

func ragNCLExpectedLocatorPatterns(segments []ragCommonsBookSegment) []string {
	seen := map[int]bool{}
	var result []string
	for _, segment := range segments {
		if seen[segment.Part] {
			continue
		}
		seen[segment.Part] = true
		pageID := 138125281
		if segment.Part == 2 {
			pageID = 138043642
		}
		result = append(result, fmt.Sprintf("https://commons.wikimedia.org/wiki/Special:Redirect/page/%d?page={physical_page}", pageID))
	}
	return result
}

func TestRAGNCLPageReviewGeneratorContract(t *testing.T) {
	path := "../../../../scripts/generate-sanmintonghui-ncl-review-queue.go"
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := ragCommonsHashFile(t, path); got != ragNCLPageReviewGeneratorSHA256 {
		t.Fatalf("page-review generator SHA-256 = %s, want %s", got, ragNCLPageReviewGeneratorSHA256)
	}
	for _, marker := range []string{
		ragNCLPageCandidatesSHA256, ragNCLVolumeLabelObservationSHA256, ragNCLPageReviewSchemaSHA256,
		"protocol_frozen_reviews_not_started", "blocked_by_volume_mapping_adjudication", "coordinator_only_until_both_independent_submissions_are_sealed",
		"lowest_sha256_of_page_candidate_report_sha256_tab_item_id_lf_first_39", "HumanReviewStarted", "ChapterPageMappingVerified",
	} {
		if !bytes.Contains(raw, []byte(marker)) {
			t.Fatalf("page-review generator missing %q", marker)
		}
	}
}

func TestRAGNCLPageReviewProtocolIsNotRuntimeRegistered(t *testing.T) {
	for _, sourcePath := range []string{
		"../localrag/index.go", "../localrag/retriever.go", "../interpretation/bazi.go", "../../model/dto.go",
		"../../../cmd/bazi-rag-index/main.go", "../../../../scripts/build-local-bazi-rag-index.sh", "../../../../scripts/build-ragflow-bazi-manifest.sh",
	} {
		raw, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{
			"sanming_ncl_page_mapping_review_protocol_v1", "sanming_ncl_page_mapping_review_submission_v1",
			"sanming-ncl-page-mapping-review-v1.json", "coordinator_priority_rank", "agreement_audit_selected_item_ids",
		} {
			if bytes.Contains(raw, []byte(forbidden)) {
				t.Fatalf("research-only page-review marker %q leaked into %s", forbidden, sourcePath)
			}
		}
	}
}

func TestRAGNCLPageReviewResearchDocumentsContract(t *testing.T) {
	marker := "第一百五十一项完成 NCL 章节页码双人独立审阅协议与失败关闭队列治理"
	for _, documentPath := range []string{
		"../../../../docs/fortune-accuracy-research-plan.md", "../../../../docs/fortune-accuracy-roadmap.md", "../../../../docs/precision-test-plan.md",
	} {
		raw, err := os.ReadFile(documentPath)
		if err != nil {
			t.Fatal(err)
		}
		if count := bytes.Count(raw, []byte(marker)); count != 1 {
			t.Fatalf("phase 151 marker count in %s = %d, want 1", documentPath, count)
		}
	}
}

func containsAll(values, expected []string) bool {
	seen := map[string]bool{}
	for _, value := range values {
		seen[value] = true
	}
	for _, value := range expected {
		if !seen[value] {
			return false
		}
	}
	return true
}
