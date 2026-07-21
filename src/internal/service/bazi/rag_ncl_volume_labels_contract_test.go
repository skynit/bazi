package bazi

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	ragNCLVolumeLabelObservationSHA256 = "8dfe0d08345b2ded9b58b394e44576e72a8609d46208ac8d4908c05bb873ca31"
	ragNCLVolumeLabelManifestSHA256    = "6a8fb60ee791c5b01dddbc5068cb1a58f1694626c686d5a82fd4ce2af25da1a5"
)

type ragNCLVolumeLabelObservation struct {
	Schema       string                      `json:"schema"`
	Version      string                      `json:"version"`
	Status       string                      `json:"status"`
	ObservedAt   string                      `json:"observed_at"`
	CandidateID  string                      `json:"candidate_id"`
	Purpose      string                      `json:"purpose"`
	Inputs       ragNCLVolumeLabelInputs     `json:"inputs"`
	Method       ragNCLVolumeLabelMethod     `json:"method"`
	Review       ragNCLVolumeLabelReview     `json:"review"`
	Observations []ragNCLVolumeLabelEntry    `json:"observations"`
	Result       ragNCLVolumeLabelResult     `json:"result"`
	Boundaries   ragNCLVolumeLabelBoundaries `json:"boundaries"`
}

type ragNCLVolumeLabelInputs struct {
	SnapshotManifest          ragCommonsArtifactReference `json:"snapshot_manifest"`
	PhysicalBookBoundaryAudit ragCommonsArtifactReference `json:"physical_book_boundary_audit"`
}

type ragNCLVolumeLabelMethod struct {
	EvidenceRenderer      string `json:"evidence_renderer"`
	RenderArguments       string `json:"render_arguments"`
	ReadingOrder          string `json:"reading_order"`
	SelectionRule         string `json:"selection_rule"`
	TranscriptionScope    string `json:"transcription_scope"`
	MappingManifestScheme string `json:"mapping_manifest_scheme"`
	MappingManifestSHA256 string `json:"mapping_manifest_sha256"`
}

type ragNCLVolumeLabelReview struct {
	OperatorCount            int      `json:"operator_count"`
	IndependentReviewerCount int      `json:"independent_reviewer_count"`
	AdjudicatorCount         int      `json:"adjudicator_count"`
	Status                   string   `json:"status"`
	Disagreements            []string `json:"disagreements"`
	GoldEligible             bool     `json:"gold_eligible"`
}

type ragNCLVolumeLabelEntry struct {
	BookCandidate        int                       `json:"book_candidate"`
	VolumeCandidate      int                       `json:"volume_candidate"`
	PrintedVolumeNumeral string                    `json:"printed_volume_numeral"`
	TranscribedHeading   string                    `json:"transcribed_heading"`
	TranscriptionStatus  string                    `json:"transcription_status"`
	Source               ragNCLVolumeLabelSource   `json:"source"`
	Evidence             ragNCLVolumeLabelEvidence `json:"evidence"`
}

type ragNCLVolumeLabelSource struct {
	Part                 int                     `json:"part"`
	PhysicalPage         int                     `json:"physical_page"`
	PhysicalBookSegments []ragCommonsBookSegment `json:"physical_book_segments"`
	CommonsPageLocator   string                  `json:"commons_page_locator"`
}

type ragNCLVolumeLabelEvidence struct {
	Path         string `json:"path"`
	SizeBytes    int64  `json:"size_bytes"`
	SHA256       string `json:"sha256"`
	WidthPixels  int    `json:"width_pixels"`
	HeightPixels int    `json:"height_pixels"`
	MIME         string `json:"mime"`
}

type ragNCLVolumeLabelResult struct {
	PhysicalBookCandidateCount           int             `json:"physical_book_candidate_count"`
	PrintedVolumeHeadingCount            int             `json:"printed_volume_heading_count"`
	UniqueVolumeCandidateCount           int             `json:"unique_volume_candidate_count"`
	SequentialVolumeCandidates           bool            `json:"sequential_volume_candidates"`
	BookToVolumeCandidateAlignment       bool            `json:"book_to_volume_candidate_alignment"`
	VolumeOneFrontMatterBeforeHeading    ragNCLPageRange `json:"volume_one_front_matter_before_heading"`
	DirectPrintedHeadingObservedAllBooks bool            `json:"direct_printed_heading_observed_for_all_books"`
}

type ragNCLPageRange struct {
	Part      int `json:"part"`
	FirstPage int `json:"first_page"`
	LastPage  int `json:"last_page"`
}

type ragNCLVolumeLabelBoundaries struct {
	PrintedVolumeLabelsObserved        bool `json:"printed_volume_labels_observed"`
	SingleOperatorMappingComplete      bool `json:"single_operator_mapping_complete"`
	ProviderExtentConsistent           bool `json:"provider_extent_consistent"`
	IndependentReviewComplete          bool `json:"independent_review_complete"`
	BibliographyAdjudicated            bool `json:"bibliography_adjudicated"`
	IndependentPrimaryArtifactVerified bool `json:"independent_primary_artifact_verified"`
	CompletePrimaryTextVerified        bool `json:"complete_primary_text_verified"`
	VolumeMappingVerified              bool `json:"volume_mapping_verified"`
	ChapterPageMappingVerified         bool `json:"chapter_page_mapping_verified"`
	ClaimSupportReviewed               bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed            bool `json:"runtime_ingestion_allowed"`
	ClaimEligible                      bool `json:"claim_eligible"`
	PublishableAccuracy                bool `json:"publishable_accuracy"`
}

func TestRAGNCLVolumeLabelObservationContract(t *testing.T) {
	observationPath := filepath.Join(ragCommonsSnapshotRoot, "volume-label-observation.json")
	observation := ragCommonsReadStrictJSON[ragNCLVolumeLabelObservation](t, observationPath, ragNCLVolumeLabelObservationSHA256)
	if observation.Schema != "sanming_ncl_volume_label_observation_v1" || observation.Version != "2026-07-17.1" ||
		observation.Status != "single_operator_volume_mapping_candidates_not_gold" || observation.ObservedAt != "2026-07-17" ||
		observation.CandidateID != ragCommonsCandidateID || strings.TrimSpace(observation.Purpose) == "" {
		t.Fatalf("unexpected NCL volume-label observation identity: %+v", observation)
	}
	if observation.Inputs.SnapshotManifest != (ragCommonsArtifactReference{Path: "snapshot-manifest.json", SHA256: ragCommonsSnapshotManifestSHA256}) ||
		observation.Inputs.PhysicalBookBoundaryAudit != (ragCommonsArtifactReference{Path: "book-boundary-audit.json", SHA256: ragCommonsBoundaryAuditSHA256}) {
		t.Fatalf("unexpected NCL volume-label inputs: %+v", observation.Inputs)
	}
	method := observation.Method
	if method.EvidenceRenderer != "pdftoppm version 26.05.0" ||
		method.RenderArguments != "-f physical_page -singlefile -jpeg -jpegopt quality=90 -r 144" ||
		method.ReadingOrder != "vertical_columns_right_to_left" ||
		method.SelectionRule != "first observed printed heading matching 三命通會卷之{Chinese volume numeral} inside each physical-book candidate" ||
		method.TranscriptionScope != "printed_volume_heading_only" ||
		method.MappingManifestScheme != "volume_tab_book_candidate_tab_part_tab_physical_page_tab_transcribed_heading_tab_evidence_sha256_lf_v1" ||
		method.MappingManifestSHA256 != ragNCLVolumeLabelManifestSHA256 {
		t.Fatalf("unexpected NCL volume-label method: %+v", method)
	}
	if observation.Review.OperatorCount != 1 || observation.Review.IndependentReviewerCount != 0 ||
		observation.Review.AdjudicatorCount != 0 || observation.Review.Status != "single_operator_non_independent" ||
		len(observation.Review.Disagreements) != 0 || observation.Review.GoldEligible {
		t.Fatalf("volume-label review must remain non-independent and non-Gold: %+v", observation.Review)
	}

	ragNCLValidateVolumeLabelEntries(t, observation)
	ragNCLValidateVolumeLabelResultAndBoundaries(t, observation)
	ragNCLValidateVolumeLabelScript(t)
}

func TestRAGNCLVolumeLabelsAreNotRuntimeRegistered(t *testing.T) {
	for _, sourcePath := range []string{
		"../localrag/index.go",
		"../localrag/retriever.go",
		"../interpretation/bazi.go",
		"../../model/dto.go",
		"../../../cmd/bazi-rag-index/main.go",
		"../../../../scripts/build-local-bazi-rag-index.sh",
		"../../../../scripts/build-ragflow-bazi-manifest.sh",
	} {
		raw, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{
			"sanming_ncl_volume_label_observation_v1",
			"volume-label-observation.json",
			"evidence/volume-labels",
		} {
			if bytes.Contains(raw, []byte(forbidden)) {
				t.Fatalf("research-only volume-label marker %q leaked into %s", forbidden, sourcePath)
			}
		}
	}
}

func TestRAGNCLVolumeLabelResearchDocumentsContract(t *testing.T) {
	marker := "第一百四十九项完成 NCL 1578 年十二卷印刷卷标候选治理"
	for _, documentPath := range []string{
		"../../../../docs/fortune-accuracy-research-plan.md",
		"../../../../docs/fortune-accuracy-roadmap.md",
		"../../../../docs/precision-test-plan.md",
	} {
		raw, err := os.ReadFile(documentPath)
		if err != nil {
			t.Fatal(err)
		}
		if count := bytes.Count(raw, []byte(marker)); count != 1 {
			t.Fatalf("phase 149 marker count in %s = %d, want 1", documentPath, count)
		}
	}
}

func ragNCLValidateVolumeLabelEntries(t *testing.T, observation ragNCLVolumeLabelObservation) {
	t.Helper()
	expectedNumerals := []string{"一", "二", "三", "四", "五", "六", "七", "八", "九", "十", "十一", "十二"}
	expectedParts := []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 2}
	expectedPages := []int{7, 101, 193, 260, 337, 428, 526, 642, 762, 879, 962, 73}
	expectedSizes := []int64{295792, 274667, 268936, 279400, 267516, 266166, 261128, 261444, 264348, 263655, 261111, 260712}
	expectedHeights := []int{1192, 1188, 1198, 1190, 1186, 1174, 1174, 1164, 1192, 1188, 1172, 1156}
	expectedHashes := []string{
		"87c5e5e321b1e3e2f21724ced9078ec484b4c2d468a1574799a5e121cab47666",
		"24ed98506e6ddb943f222c3815f7aa8a66e552bcfeb961d16c076f011e013964",
		"9e78d54651fef2d4d2c66dd9c93c806e4fededf2e4461da0205735f19c2d3465",
		"cec1c461f08def66044691621a737a20ce4f8c0c909d38dd055076a79ce8af78",
		"7c67b7e6706d3581f0129454ee183181c7502834e44c24b80ea8b61f85fe05be",
		"6423f52d732d8886cf865da91350ba99b11260c4989c4fb670568677a7bb2937",
		"92a4145507738768986c0d0f8e61f633fe3190f66e1d68b7d4452053968e28b9",
		"868bb46a2f6103b43ea8d795823cebf37cfcf34048a2b7ba1435795e6f4e8fb2",
		"e4b209fb789c9fde41d33538f5a4f1164ee85d5aa089840173c83b8dda14c49c",
		"9b92bfe074ed47b57d8a4cad601b256a11ca4dbaec402762ec5428f56546860b",
		"71a6706d04f62ebbfc9a9f5f3287eae0e7a3a8d4cdd152945440d90a3e5fec7d",
		"4e89f2778c622dacbc1b9c0dfd54eca4cd81f589764b0a17d48c74d04e6a8dae",
	}
	boundaryAudit := ragCommonsReadStrictJSON[ragCommonsBoundaryAudit](t, filepath.Join(ragCommonsSnapshotRoot, "book-boundary-audit.json"), ragCommonsBoundaryAuditSHA256)
	if len(observation.Observations) != 12 || len(boundaryAudit.PhysicalBookCandidateRanges) != 12 {
		t.Fatalf("volume-label or physical-book count is not 12")
	}
	var mappingRows strings.Builder
	seenVolumes := map[int]bool{}
	seenEvidence := map[string]bool{}
	for index, entry := range observation.Observations {
		volume := index + 1
		wantHeading := "三命通會卷之" + expectedNumerals[index]
		if entry.BookCandidate != volume || entry.VolumeCandidate != volume || seenVolumes[volume] ||
			entry.PrintedVolumeNumeral != expectedNumerals[index] || entry.TranscribedHeading != wantHeading ||
			entry.TranscriptionStatus != "single_operator_visual_reading" || entry.Source.Part != expectedParts[index] ||
			entry.Source.PhysicalPage != expectedPages[index] || entry.Evidence.SizeBytes != expectedSizes[index] ||
			entry.Evidence.SHA256 != expectedHashes[index] || entry.Evidence.WidthPixels != 1684 ||
			entry.Evidence.HeightPixels != expectedHeights[index] || entry.Evidence.MIME != "image/jpeg" {
			t.Fatalf("unexpected volume-label entry %d: %+v", volume, entry)
		}
		seenVolumes[volume] = true
		if seenEvidence[entry.Evidence.Path] {
			t.Fatalf("duplicate volume-label evidence path %q", entry.Evidence.Path)
		}
		seenEvidence[entry.Evidence.Path] = true

		wantSegments := boundaryAudit.PhysicalBookCandidateRanges[index].Segments
		if fmt.Sprint(entry.Source.PhysicalBookSegments) != fmt.Sprint(wantSegments) ||
			!ragNCLPageWithinSegments(entry.Source.Part, entry.Source.PhysicalPage, wantSegments) {
			t.Fatalf("volume %d label page is not inside its physical-book candidate: source=%+v want=%+v", volume, entry.Source, wantSegments)
		}
		pageID := 138125281
		if entry.Source.Part == 2 {
			pageID = 138043642
		}
		wantLocator := fmt.Sprintf("https://commons.wikimedia.org/wiki/Special:Redirect/page/%d?page=%d", pageID, entry.Source.PhysicalPage)
		if entry.Source.CommonsPageLocator != wantLocator {
			t.Fatalf("volume %d locator = %q, want %q", volume, entry.Source.CommonsPageLocator, wantLocator)
		}

		evidencePath := ragCommonsJoinSnapshotPath(t, entry.Evidence.Path)
		info, err := os.Stat(evidencePath)
		if err != nil || info.Size() != entry.Evidence.SizeBytes {
			t.Fatalf("volume %d evidence identity: info=%v err=%v", volume, info, err)
		}
		if got := ragCommonsHashFile(t, evidencePath); got != entry.Evidence.SHA256 {
			t.Fatalf("volume %d evidence SHA-256 = %s, want %s", volume, got, entry.Evidence.SHA256)
		}
		raw, err := os.ReadFile(evidencePath)
		if err != nil || len(raw) < 3 || !bytes.Equal(raw[:3], []byte{0xff, 0xd8, 0xff}) {
			t.Fatalf("volume %d evidence is not an unambiguous JPEG: %v", volume, err)
		}

		fmt.Fprintf(&mappingRows, "%d\t%d\t%d\t%d\t%s\t%s\n", entry.VolumeCandidate, entry.BookCandidate,
			entry.Source.Part, entry.Source.PhysicalPage, entry.TranscribedHeading, entry.Evidence.SHA256)
	}
	if got := ragCommonsSHA256([]byte(mappingRows.String())); got != ragNCLVolumeLabelManifestSHA256 ||
		got != observation.Method.MappingManifestSHA256 {
		t.Fatalf("volume-label mapping manifest SHA-256 = %s, want %s", got, ragNCLVolumeLabelManifestSHA256)
	}
}

func ragNCLValidateVolumeLabelResultAndBoundaries(t *testing.T, observation ragNCLVolumeLabelObservation) {
	t.Helper()
	result := observation.Result
	if result.PhysicalBookCandidateCount != 12 || result.PrintedVolumeHeadingCount != 12 ||
		result.UniqueVolumeCandidateCount != 12 || !result.SequentialVolumeCandidates ||
		!result.BookToVolumeCandidateAlignment || result.VolumeOneFrontMatterBeforeHeading != (ragNCLPageRange{Part: 1, FirstPage: 3, LastPage: 6}) ||
		!result.DirectPrintedHeadingObservedAllBooks {
		t.Fatalf("unexpected NCL volume-label result: %+v", result)
	}
	boundaries := observation.Boundaries
	if !boundaries.PrintedVolumeLabelsObserved || !boundaries.SingleOperatorMappingComplete || !boundaries.ProviderExtentConsistent ||
		boundaries.IndependentReviewComplete || boundaries.BibliographyAdjudicated || boundaries.IndependentPrimaryArtifactVerified ||
		boundaries.CompletePrimaryTextVerified || boundaries.VolumeMappingVerified || boundaries.ChapterPageMappingVerified ||
		boundaries.ClaimSupportReviewed || boundaries.RuntimeIngestionAllowed || boundaries.ClaimEligible || boundaries.PublishableAccuracy {
		t.Fatalf("NCL volume-label boundaries must fail closed: %+v", boundaries)
	}
}

func ragNCLValidateVolumeLabelScript(t *testing.T) {
	t.Helper()
	raw, err := os.ReadFile("../../../../scripts/audit-ncl-sanmintonghui-volume-labels.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{
		"set -euo pipefail",
		ragCommonsSnapshotManifestSHA256,
		ragCommonsBoundaryAuditSHA256,
		"-singlefile -jpeg -jpegopt quality=90 -r 144",
		"single_operator_volume_mapping_candidates_not_gold",
		"independent_review_complete:false",
		"volume_mapping_verified:false",
	} {
		if !bytes.Contains(raw, []byte(marker)) {
			t.Fatalf("volume-label audit script missing %q", marker)
		}
	}
}

func ragNCLPageWithinSegments(part, page int, segments []ragCommonsBookSegment) bool {
	for _, segment := range segments {
		if segment.Part == part && segment.FirstPage <= page && page <= segment.LastPage {
			return true
		}
	}
	return false
}
