package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

const (
	expectedPageCandidatesSHA256   = "0f915acaea01df0fd66738097105ba42786850b7fd9e323d9db3529693d0ce65"
	expectedVolumeLabelsSHA256     = "8dfe0d08345b2ded9b58b394e44576e72a8609d46208ac8d4908c05bb873ca31"
	expectedSubmissionSchemaSHA256 = "e2a84f7ff40d15d3a6404ad11901178a3ada1ce10df0f29aa603d36cb899dcf4"
	agreementAuditCount            = 39
)

type artifactReference struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type pageKey struct {
	Part         int `json:"part"`
	PhysicalPage int `json:"physical_page"`
}

type bookSegment struct {
	Part      int `json:"part"`
	FirstPage int `json:"first_page"`
	LastPage  int `json:"last_page"`
}

type pageCandidateReport struct {
	Schema      string `json:"schema"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	CandidateID string `json:"candidate_id"`
	Chapters    []struct {
		Chapter                    int       `json:"chapter"`
		File                       string    `json:"file"`
		Title                      string    `json:"title"`
		OriginalSHA256             string    `json:"original_sha256"`
		VolumeCandidate            int       `json:"volume_candidate"`
		TitleLocatorCandidates     []pageKey `json:"title_locator_candidates"`
		TitleLocatorCandidateCount int       `json:"title_locator_candidate_count"`
		BestContentOverlap         int       `json:"best_content_overlap"`
		ContentOverlapMargin       int       `json:"content_overlap_margin"`
		BestContentCandidateCount  int       `json:"best_content_candidate_count"`
		BestContentCandidates      []pageKey `json:"best_content_candidates"`
		ZeroContentOverlap         bool      `json:"zero_content_overlap"`
	} `json:"chapters"`
	Boundaries struct {
		MachineCandidatesOnly      bool `json:"machine_candidates_only"`
		ChapterPageMappingVerified bool `json:"chapter_page_mapping_verified"`
		RuntimeIngestionAllowed    bool `json:"runtime_ingestion_allowed"`
		ClaimEligible              bool `json:"claim_eligible"`
		PublishableAccuracy        bool `json:"publishable_accuracy"`
	} `json:"boundaries"`
}

type volumeLabelObservation struct {
	Schema       string `json:"schema"`
	Version      string `json:"version"`
	Status       string `json:"status"`
	CandidateID  string `json:"candidate_id"`
	Observations []struct {
		BookCandidate      int    `json:"book_candidate"`
		VolumeCandidate    int    `json:"volume_candidate"`
		TranscribedHeading string `json:"transcribed_heading"`
		Source             struct {
			Part                 int           `json:"part"`
			PhysicalPage         int           `json:"physical_page"`
			PhysicalBookSegments []bookSegment `json:"physical_book_segments"`
			CommonsPageLocator   string        `json:"commons_page_locator"`
		} `json:"source"`
		Evidence struct {
			Path      string `json:"path"`
			SizeBytes int64  `json:"size_bytes"`
			SHA256    string `json:"sha256"`
		} `json:"evidence"`
	} `json:"observations"`
	Boundaries struct {
		IndependentReviewComplete  bool `json:"independent_review_complete"`
		VolumeMappingVerified      bool `json:"volume_mapping_verified"`
		ChapterPageMappingVerified bool `json:"chapter_page_mapping_verified"`
		RuntimeIngestionAllowed    bool `json:"runtime_ingestion_allowed"`
		ClaimEligible              bool `json:"claim_eligible"`
		PublishableAccuracy        bool `json:"publishable_accuracy"`
	} `json:"boundaries"`
}

type reviewProtocol struct {
	Schema      string `json:"schema"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	ObservedAt  string `json:"observed_at"`
	ProtocolID  string `json:"protocol_id"`
	CandidateID string `json:"candidate_id"`
	Purpose     string `json:"purpose"`
	Sources     struct {
		PageCandidates   artifactReference `json:"page_candidates"`
		VolumeLabels     artifactReference `json:"volume_labels"`
		SubmissionSchema artifactReference `json:"submission_schema"`
	} `json:"sources"`
	Blinding     blindingProtocol     `json:"blinding"`
	Stages       reviewStages         `json:"stages"`
	Adjudication adjudicationProtocol `json:"adjudication"`
	Quality      qualityProtocol      `json:"quality_control"`
	Gates        reviewGates          `json:"gates"`
	ReviewState  reviewState          `json:"review_state"`
	Summary      reviewSummary        `json:"summary"`
	VolumeItems  []volumeReviewItem   `json:"volume_items"`
	ChapterItems []chapterReviewItem  `json:"chapter_items"`
	Boundaries   reviewBoundaries     `json:"boundaries"`
}

type blindingProtocol struct {
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
	RealIdentityStoredInArtifact     bool     `json:"real_identity_stored_in_research_artifact"`
}

type reviewStage struct {
	Status              string   `json:"status"`
	ItemCount           int      `json:"item_count"`
	ReviewUnit          string   `json:"review_unit"`
	AllowedDecisions    []string `json:"allowed_decisions"`
	RequiredFields      []string `json:"required_fields"`
	ReleaseCondition    string   `json:"release_condition"`
	GoldCondition       string   `json:"gold_condition"`
	ClaimSupportInScope bool     `json:"claim_support_in_scope"`
}

type reviewStages struct {
	VolumeLabelReview     reviewStage `json:"volume_label_review"`
	ChapterBoundaryReview reviewStage `json:"chapter_boundary_review"`
}

type adjudicationProtocol struct {
	Trigger                           string   `json:"trigger"`
	ComparisonFields                  []string `json:"comparison_fields"`
	MajorityVoteAllowed               bool     `json:"majority_vote_allowed"`
	AdjudicatorSeesReviewerPseudonyms bool     `json:"adjudicator_sees_reviewer_pseudonyms"`
	AdjudicatorSeesMachineCandidates  bool     `json:"adjudicator_sees_machine_candidates"`
	RequiredResolutionFields          []string `json:"required_resolution_fields"`
	UnresolvedDecision                string   `json:"unresolved_decision"`
}

type qualityProtocol struct {
	AgreementAuditCount   int      `json:"agreement_audit_count"`
	AgreementAuditRateBP  int      `json:"agreement_audit_rate_basis_points"`
	SelectionScheme       string   `json:"agreement_audit_selection_scheme"`
	SelectionSeed         string   `json:"agreement_audit_selection_seed"`
	SelectedItemIDs       []string `json:"agreement_audit_selected_item_ids"`
	AuditReviewerCount    int      `json:"audit_reviewer_count"`
	AgreementCanSkipAudit bool     `json:"agreement_can_skip_prespecified_audit"`
	FailedAuditAction     string   `json:"failed_audit_action"`
}

type reviewGates struct {
	VolumeReviewReleased           bool `json:"volume_review_released"`
	VolumeMappingAdjudicated       bool `json:"volume_mapping_adjudicated"`
	ChapterReviewReleased          bool `json:"chapter_review_released"`
	AllChapterReviewsComplete      bool `json:"all_chapter_reviews_complete"`
	AgreementAuditComplete         bool `json:"agreement_audit_complete"`
	ChapterPageMappingGoldReleased bool `json:"chapter_page_mapping_gold_released"`
}

type reviewState struct {
	ReviewerAssignmentsCreated   int `json:"reviewer_assignments_created"`
	IndependentReviewersAssigned int `json:"independent_reviewers_assigned"`
	ReviewerSubmissionsReceived  int `json:"reviewer_submissions_received"`
	AdjudicatorsAssigned         int `json:"adjudicators_assigned"`
	AdjudicationsComplete        int `json:"adjudications_complete"`
	AgreementAuditsComplete      int `json:"agreement_audits_complete"`
	GoldItems                    int `json:"gold_items"`
}

type reviewSummary struct {
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

type volumeReviewItem struct {
	ItemID                      string        `json:"item_id"`
	VolumeCandidate             int           `json:"volume_candidate"`
	CoordinatorCandidateHeading string        `json:"coordinator_candidate_heading"`
	HeadingPage                 pageKey       `json:"heading_page"`
	PhysicalBookSegments        []bookSegment `json:"physical_book_segments"`
	CommonsPageLocator          string        `json:"commons_page_locator"`
	Evidence                    struct {
		Path      string `json:"path"`
		SizeBytes int64  `json:"size_bytes"`
		SHA256    string `json:"sha256"`
	} `json:"evidence"`
	MachineCandidateVisibleToReviewer bool   `json:"machine_candidate_visible_to_reviewer"`
	ReviewStatus                      string `json:"review_status"`
}

type chapterMachineSignal struct {
	TitleLocatorCandidateCount int       `json:"title_locator_candidate_count"`
	TitleLocatorCandidates     []pageKey `json:"title_locator_candidates"`
	BestContentOverlap         int       `json:"best_content_overlap"`
	ContentOverlapMargin       int       `json:"content_overlap_margin"`
	BestContentCandidateCount  int       `json:"best_content_candidate_count"`
	BestContentCandidates      []pageKey `json:"best_content_candidates"`
	ZeroContentOverlap         bool      `json:"zero_content_overlap"`
}

type chapterReviewItem struct {
	ItemID                  string               `json:"item_id"`
	Chapter                 int                  `json:"chapter"`
	File                    string               `json:"file"`
	Title                   string               `json:"title"`
	OriginalSHA256          string               `json:"original_sha256"`
	VolumeCandidate         int                  `json:"volume_candidate"`
	PhysicalBookSegments    []bookSegment        `json:"physical_book_segments"`
	PageLocatorPatterns     []string             `json:"page_locator_patterns"`
	CoordinatorPriorityRank int                  `json:"coordinator_priority_rank"`
	CoordinatorRiskStratum  string               `json:"coordinator_risk_stratum"`
	CoordinatorRiskReasons  []string             `json:"coordinator_risk_reasons"`
	AgreementAuditSelected  bool                 `json:"agreement_audit_selected"`
	MachineSignalVisibility string               `json:"machine_signal_visibility"`
	MachineSignal           chapterMachineSignal `json:"machine_signal"`
	ReviewStatus            string               `json:"review_status"`
}

type reviewBoundaries struct {
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

type priorityRecord struct {
	Chapter int
	Stratum string
	Reasons []string
}

type auditRecord struct {
	ItemID string
	Hash   string
}

func main() {
	var candidatesPath, volumeLabelsPath, submissionSchemaPath, output string
	flag.StringVar(&candidatesPath, "page-candidates", "", "path to the frozen chapter page-candidate report")
	flag.StringVar(&volumeLabelsPath, "volume-labels", "", "path to the frozen volume-label observation")
	flag.StringVar(&submissionSchemaPath, "submission-schema", "", "path to the independent-review submission schema")
	flag.StringVar(&output, "output", "", "path for the coordinator protocol and queue")
	flag.Parse()
	if candidatesPath == "" || volumeLabelsPath == "" || submissionSchemaPath == "" || output == "" {
		fail(errors.New("-page-candidates, -volume-labels, -submission-schema, and -output are required"))
	}
	if err := run(candidatesPath, volumeLabelsPath, submissionSchemaPath, output); err != nil {
		fail(err)
	}
}

func run(candidatesPath, volumeLabelsPath, submissionSchemaPath, output string) error {
	candidateRaw, candidates, err := readJSON[pageCandidateReport](candidatesPath)
	if err != nil {
		return err
	}
	if hashBytes(candidateRaw) != expectedPageCandidatesSHA256 {
		return errors.New("page-candidate report SHA-256 mismatch")
	}
	volumeRaw, volumes, err := readJSON[volumeLabelObservation](volumeLabelsPath)
	if err != nil {
		return err
	}
	if hashBytes(volumeRaw) != expectedVolumeLabelsSHA256 {
		return errors.New("volume-label observation SHA-256 mismatch")
	}
	schemaRaw, err := os.ReadFile(submissionSchemaPath)
	if err != nil {
		return err
	}
	if hashBytes(schemaRaw) != expectedSubmissionSchemaSHA256 {
		return errors.New("submission schema SHA-256 mismatch")
	}
	var schema map[string]any
	if err := decodeOne(schemaRaw, &schema); err != nil {
		return err
	}
	if schema["$schema"] != "https://json-schema.org/draft/2020-12/schema" {
		return errors.New("unexpected submission schema dialect")
	}
	if err := validateInputs(candidates, volumes); err != nil {
		return err
	}
	protocol, err := buildProtocol(candidates, volumes)
	if err != nil {
		return err
	}
	return writeJSONAtomic(output, protocol)
}

func validateInputs(candidates pageCandidateReport, volumes volumeLabelObservation) error {
	if candidates.Schema != "sanming_ncl_chapter_page_candidate_audit_v1" || candidates.Version != "2026-07-17.1" ||
		candidates.Status != "chapter_page_candidates_machine_only" || candidates.CandidateID != "sanming-ncl-06589-1578-12vol-scan-v1" ||
		len(candidates.Chapters) != 381 || !candidates.Boundaries.MachineCandidatesOnly || candidates.Boundaries.ChapterPageMappingVerified ||
		candidates.Boundaries.RuntimeIngestionAllowed || candidates.Boundaries.ClaimEligible || candidates.Boundaries.PublishableAccuracy {
		return errors.New("unexpected page-candidate report identity or boundary")
	}
	if volumes.Schema != "sanming_ncl_volume_label_observation_v1" || volumes.Version != "2026-07-17.1" ||
		volumes.Status != "single_operator_volume_mapping_candidates_not_gold" || volumes.CandidateID != candidates.CandidateID || len(volumes.Observations) != 12 ||
		volumes.Boundaries.IndependentReviewComplete || volumes.Boundaries.VolumeMappingVerified || volumes.Boundaries.ChapterPageMappingVerified ||
		volumes.Boundaries.RuntimeIngestionAllowed || volumes.Boundaries.ClaimEligible || volumes.Boundaries.PublishableAccuracy {
		return errors.New("unexpected volume-label observation identity or boundary")
	}
	for index, volume := range volumes.Observations {
		if volume.BookCandidate != index+1 || volume.VolumeCandidate != index+1 || len(volume.Source.PhysicalBookSegments) == 0 {
			return fmt.Errorf("invalid volume item %d", index+1)
		}
	}
	for index, chapter := range candidates.Chapters {
		if chapter.Chapter != index+1 || chapter.File != fmt.Sprintf("%03d.md", index+1) || chapter.VolumeCandidate < 1 || chapter.VolumeCandidate > 12 ||
			chapter.TitleLocatorCandidateCount != len(chapter.TitleLocatorCandidates) || chapter.BestContentCandidateCount != len(chapter.BestContentCandidates) ||
			chapter.ZeroContentOverlap != (chapter.BestContentOverlap == 0) {
			return fmt.Errorf("invalid chapter item %d", index+1)
		}
	}
	return nil
}

func buildProtocol(candidates pageCandidateReport, volumes volumeLabelObservation) (reviewProtocol, error) {
	priorities := buildPriorities(candidates)
	auditIDs := selectAgreementAudit(candidates)
	auditSet := map[string]bool{}
	for _, id := range auditIDs {
		auditSet[id] = true
	}
	priorityRank := map[int]int{}
	for index, priority := range priorities {
		priorityRank[priority.Chapter] = index + 1
	}

	var protocol reviewProtocol
	protocol.Schema = "sanming_ncl_page_mapping_review_protocol_v1"
	protocol.Version = "2026-07-17.1"
	protocol.Status = "protocol_frozen_reviews_not_started"
	protocol.ObservedAt = "2026-07-17"
	protocol.ProtocolID = "sanming-ncl-page-mapping-review-v1"
	protocol.CandidateID = candidates.CandidateID
	protocol.Purpose = "Pre-register two independent volume-label and chapter-boundary reviews, gated adjudication, and agreement auditing without creating or implying any human decision."
	protocol.Sources.PageCandidates = artifactReference{Path: "research/rag/sanming-ncl-06589-chapter-page-candidates-v1.json", SHA256: expectedPageCandidatesSHA256}
	protocol.Sources.VolumeLabels = artifactReference{Path: "research/rag/snapshots/sanming-ncl-06589-1578-v1/volume-label-observation.json", SHA256: expectedVolumeLabelsSHA256}
	protocol.Sources.SubmissionSchema = artifactReference{Path: "research/schemas/sanming-ncl-page-review-submission-v1.schema.json", SHA256: expectedSubmissionSchemaSHA256}
	protocol.Blinding = blindingProtocol{
		ReviewerSlots: []string{"reviewer_a", "reviewer_b"}, RequiredIndependentReviewers: 2, RequiredAdjudicators: 1,
		ReviewerIdentityAssignment:  "external_coordinator_pseudonymization_not_stored_in_research_artifact",
		SeparateSubmissionArtifacts: true, CrossReviewerVisibility: "sealed_until_both_submissions_pass_schema_and_completeness_checks",
		MachineSignalVisibility:       "coordinator_only_until_both_independent_submissions_are_sealed",
		ReviewerPacketVisibleFields:   []string{"item_id", "chapter", "title", "original_sha256", "physical_book_segments", "page_locator_patterns", "scan_and_fixed_evidence"},
		ReviewerPacketForbiddenFields: []string{"coordinator_priority_rank", "coordinator_risk_stratum", "coordinator_risk_reasons", "machine_signal", "other_reviewer_submission"},
		PriorVolumeOperatorMayReview:  false, ReviewerAndAdjudicatorMayOverlap: false, RealIdentityStoredInArtifact: false,
	}
	protocol.Stages.VolumeLabelReview = reviewStage{
		Status: "ready_unassigned", ItemCount: 12, ReviewUnit: "one physical-book candidate and its printed volume heading",
		AllowedDecisions: []string{"volume_heading_confirmed", "volume_heading_differs", "wrong_book_segment", "scan_incomplete_or_illegible"},
		RequiredFields:   []string{"item_id", "decision", "transcribed_heading", "evidence_pages", "attestations"},
		ReleaseCondition: "protocol_frozen_and_two_eligible_reviewers_assigned_separately",
		GoldCondition:    "two exact independent decisions agree or a separate adjudicator resolves every disagreement", ClaimSupportInScope: false,
	}
	protocol.Stages.ChapterBoundaryReview = reviewStage{
		Status: "blocked_by_volume_mapping_adjudication", ItemCount: 381, ReviewUnit: "one Markdown chapter mapped to exact inclusive scan start and end physical pages",
		AllowedDecisions: []string{"chapter_boundary_located", "chapter_not_located", "scan_incomplete_or_illegible"},
		RequiredFields:   []string{"item_id", "decision", "chapter_start", "chapter_end", "evidence_pages", "attestations"},
		ReleaseCondition: "all_12_volume_items_have_two_valid_reviews_and_all_disagreements_are_adjudicated",
		GoldCondition:    "all chapter items have two exact independent decisions or adjudication plus prespecified agreement audit completion", ClaimSupportInScope: false,
	}
	protocol.Adjudication = adjudicationProtocol{
		Trigger:             "any decision, transcription, start page, end page, or evidence sufficiency disagreement",
		ComparisonFields:    []string{"decision", "transcribed_heading", "chapter_start", "chapter_end", "evidence_pages"},
		MajorityVoteAllowed: false, AdjudicatorSeesReviewerPseudonyms: false, AdjudicatorSeesMachineCandidates: false,
		RequiredResolutionFields: []string{"resolved_decision", "resolved_pages", "evidence_pages", "reason", "source_submission_sha256"},
		UnresolvedDecision:       "unresolved_not_gold_eligible",
	}
	protocol.Quality = qualityProtocol{
		AgreementAuditCount: agreementAuditCount, AgreementAuditRateBP: 1024,
		SelectionScheme: "lowest_sha256_of_page_candidate_report_sha256_tab_item_id_lf_first_39",
		SelectionSeed:   expectedPageCandidatesSHA256, SelectedItemIDs: auditIDs, AuditReviewerCount: 1,
		AgreementCanSkipAudit: false, FailedAuditAction: "freeze_gold_release_and_expand_independent_re_review_to_all_items_in_affected_volume",
	}
	protocol.Gates = reviewGates{VolumeReviewReleased: true}
	protocol.ReviewState = reviewState{}
	protocol.Boundaries.ProtocolOnly = true

	for index, observation := range volumes.Observations {
		var item volumeReviewItem
		item.ItemID = fmt.Sprintf("volume-%02d", index+1)
		item.VolumeCandidate = index + 1
		item.CoordinatorCandidateHeading = observation.TranscribedHeading
		item.HeadingPage = pageKey{Part: observation.Source.Part, PhysicalPage: observation.Source.PhysicalPage}
		item.PhysicalBookSegments = observation.Source.PhysicalBookSegments
		item.CommonsPageLocator = observation.Source.CommonsPageLocator
		item.Evidence.Path = "research/rag/snapshots/sanming-ncl-06589-1578-v1/" + observation.Evidence.Path
		item.Evidence.SizeBytes = observation.Evidence.SizeBytes
		item.Evidence.SHA256 = observation.Evidence.SHA256
		item.ReviewStatus = "ready_unassigned"
		protocol.VolumeItems = append(protocol.VolumeItems, item)
	}
	strata := map[string]int{}
	priorityByChapter := map[int]priorityRecord{}
	for _, priority := range priorities {
		priorityByChapter[priority.Chapter] = priority
		strata[priority.Stratum]++
	}
	for _, chapter := range candidates.Chapters {
		priority := priorityByChapter[chapter.Chapter]
		segments := append([]bookSegment(nil), volumes.Observations[chapter.VolumeCandidate-1].Source.PhysicalBookSegments...)
		itemID := fmt.Sprintf("chapter-%03d", chapter.Chapter)
		item := chapterReviewItem{
			ItemID: itemID, Chapter: chapter.Chapter, File: chapter.File, Title: chapter.Title, OriginalSHA256: chapter.OriginalSHA256,
			VolumeCandidate: chapter.VolumeCandidate, PhysicalBookSegments: segments, PageLocatorPatterns: locatorPatterns(segments),
			CoordinatorPriorityRank: priorityRank[chapter.Chapter], CoordinatorRiskStratum: priority.Stratum,
			CoordinatorRiskReasons: priority.Reasons, AgreementAuditSelected: auditSet[itemID],
			MachineSignalVisibility: "coordinator_only", ReviewStatus: "blocked_by_volume_mapping_adjudication",
			MachineSignal: chapterMachineSignal{
				TitleLocatorCandidateCount: chapter.TitleLocatorCandidateCount, TitleLocatorCandidates: nonNilPages(chapter.TitleLocatorCandidates),
				BestContentOverlap: chapter.BestContentOverlap, ContentOverlapMargin: chapter.ContentOverlapMargin,
				BestContentCandidateCount: chapter.BestContentCandidateCount, BestContentCandidates: nonNilPages(chapter.BestContentCandidates),
				ZeroContentOverlap: chapter.ZeroContentOverlap,
			},
		}
		protocol.ChapterItems = append(protocol.ChapterItems, item)
	}
	protocol.Summary = reviewSummary{
		VolumeItems: 12, ChapterItems: 381, CriticalZeroOverlap: strata["critical_zero_overlap"],
		HighTieOrLowMargin: strata["high_tie_or_low_margin"], HighTitleContentDisagreement: strata["high_title_content_disagreement"],
		Standard: strata["standard"], ReadyItems: 12, BlockedItems: 381,
	}
	return protocol, nil
}

func buildPriorities(report pageCandidateReport) []priorityRecord {
	priorities := make([]priorityRecord, 0, len(report.Chapters))
	for _, chapter := range report.Chapters {
		priority := priorityRecord{Chapter: chapter.Chapter, Stratum: "standard", Reasons: []string{"no_prespecified_machine_uncertainty_trigger"}}
		switch {
		case chapter.ZeroContentOverlap:
			priority.Stratum = "critical_zero_overlap"
			priority.Reasons = []string{"zero_content_overlap"}
		case chapter.BestContentCandidateCount > 1 || chapter.ContentOverlapMargin <= 2:
			priority.Stratum = "high_tie_or_low_margin"
			priority.Reasons = nil
			if chapter.BestContentCandidateCount > 1 {
				priority.Reasons = append(priority.Reasons, "best_content_page_tie")
			}
			if chapter.ContentOverlapMargin <= 2 {
				priority.Reasons = append(priority.Reasons, "content_overlap_margin_at_most_2")
			}
		case titleAndContentDisagree(chapter.TitleLocatorCandidates, chapter.BestContentCandidates):
			priority.Stratum = "high_title_content_disagreement"
			priority.Reasons = []string{"title_locator_and_best_content_pages_disjoint"}
		}
		priorities = append(priorities, priority)
	}
	rank := map[string]int{"critical_zero_overlap": 0, "high_tie_or_low_margin": 1, "high_title_content_disagreement": 2, "standard": 3}
	sort.Slice(priorities, func(i, j int) bool {
		if rank[priorities[i].Stratum] != rank[priorities[j].Stratum] {
			return rank[priorities[i].Stratum] < rank[priorities[j].Stratum]
		}
		return priorities[i].Chapter < priorities[j].Chapter
	})
	return priorities
}

func titleAndContentDisagree(titles, content []pageKey) bool {
	if len(titles) == 0 {
		return false
	}
	contentSet := map[pageKey]bool{}
	for _, page := range content {
		contentSet[page] = true
	}
	for _, page := range titles {
		if contentSet[page] {
			return false
		}
	}
	return true
}

func selectAgreementAudit(report pageCandidateReport) []string {
	records := make([]auditRecord, 0, len(report.Chapters))
	for _, chapter := range report.Chapters {
		itemID := fmt.Sprintf("chapter-%03d", chapter.Chapter)
		value := expectedPageCandidatesSHA256 + "\t" + itemID + "\n"
		records = append(records, auditRecord{ItemID: itemID, Hash: hashBytes([]byte(value))})
	}
	sort.Slice(records, func(i, j int) bool {
		if records[i].Hash != records[j].Hash {
			return records[i].Hash < records[j].Hash
		}
		return records[i].ItemID < records[j].ItemID
	})
	selected := make([]string, 0, agreementAuditCount)
	for _, record := range records[:agreementAuditCount] {
		selected = append(selected, record.ItemID)
	}
	sort.Strings(selected)
	return selected
}

func locatorPatterns(segments []bookSegment) []string {
	seen := map[int]bool{}
	var patterns []string
	for _, segment := range segments {
		if seen[segment.Part] {
			continue
		}
		seen[segment.Part] = true
		pageID := 138125281
		if segment.Part == 2 {
			pageID = 138043642
		}
		patterns = append(patterns, fmt.Sprintf("https://commons.wikimedia.org/wiki/Special:Redirect/page/%d?page={physical_page}", pageID))
	}
	return patterns
}

func nonNilPages(pages []pageKey) []pageKey {
	if pages == nil {
		return []pageKey{}
	}
	return pages
}

func readJSON[T any](path string) ([]byte, T, error) {
	var value T
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, value, err
	}
	if err := decodeOne(raw, &value); err != nil {
		return nil, value, err
	}
	return raw, value, nil
}

func decodeOne(raw []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(value); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return errors.New("input must contain one JSON document")
	}
	return nil
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".sanming-review-queue-*.json")
	if err != nil {
		return err
	}
	temporaryName := temporary.Name()
	defer os.Remove(temporaryName)
	encoder := json.NewEncoder(temporary)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(value); err != nil {
		temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return os.Rename(temporaryName, path)
}

func hashBytes(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
