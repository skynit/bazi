package bazi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	ragCitationRepositoryRoot = "../../../.."
	ragCitationCatalogPath    = "../../../../research/rag/bazi-source-catalog-v1.json"
	ragCitationCatalogSHA256  = "5a012c68833eaa1163a175f579833cca2337de2d7b22eeb2ec9ba396038059d5"
	ragProvenanceAuditPath    = "../../../../research/rag/bazi-source-provenance-audit-v1.json"
	ragProvenanceAuditSHA256  = "ad314ffa101c9bea021bb3d2feffb5d6d22ff8389986d08a955dc6db1562a834"
)

type ragCitationCatalog struct {
	Schema        string                          `json:"schema"`
	Version       string                          `json:"version"`
	Description   string                          `json:"description"`
	DefaultPolicy ragCitationCatalogDefaultPolicy `json:"default_policy"`
	Sources       []ragCitationCatalogSource      `json:"sources"`
}

type ragCitationCatalogDefaultPolicy struct {
	SourceTier         string `json:"source_tier"`
	VerificationStatus string `json:"verification_status"`
	ArtifactKind       string `json:"artifact_kind"`
	ProvenanceStatus   string `json:"provenance_status"`
	IndependenceStatus string `json:"independence_status"`
	CoverageStatus     string `json:"coverage_status"`
	ClaimEligible      bool   `json:"claim_eligible"`
}

type ragCitationCatalogSource struct {
	Book               string   `json:"book"`
	Author             string   `json:"author"`
	Edition            string   `json:"edition"`
	ArtifactPath       string   `json:"artifact_path"`
	ArtifactSHA256     string   `json:"artifact_sha256"`
	MarkdownRoot       string   `json:"markdown_root"`
	SourceTier         string   `json:"source_tier"`
	VerificationStatus string   `json:"verification_status"`
	ArtifactKind       string   `json:"artifact_kind"`
	ProvenanceStatus   string   `json:"provenance_status"`
	IndependenceStatus string   `json:"independence_status"`
	CoverageStatus     string   `json:"coverage_status"`
	PageMappingStatus  string   `json:"page_mapping_status"`
	LicenseScope       string   `json:"license_scope"`
	ClaimEligible      bool     `json:"claim_eligible"`
	Limitations        []string `json:"limitations"`
}

type ragProvenanceAudit struct {
	Schema        string                       `json:"schema"`
	Version       string                       `json:"version"`
	Status        string                       `json:"status"`
	CatalogPath   string                       `json:"catalog_path"`
	CatalogSHA256 string                       `json:"catalog_sha256"`
	ObservedWith  ragProvenanceObservedWith    `json:"observed_with"`
	ClaimBoundary ragProvenanceClaimBoundary   `json:"claim_boundary"`
	Sources       []ragProvenanceAuditedSource `json:"sources"`
}

type ragProvenanceObservedWith struct {
	PDFInfo   string `json:"pdfinfo"`
	PDFToText string `json:"pdftotext"`
	TextMode  string `json:"text_mode"`
}

type ragProvenanceClaimBoundary struct {
	BibliographicMetadataAdjudicated bool `json:"bibliographic_metadata_adjudicated"`
	IndependentPrimarySourceVerified bool `json:"independent_primary_source_verified"`
	PageMappingVerified              bool `json:"page_mapping_verified"`
	ClaimSupportReviewed             bool `json:"claim_support_reviewed"`
	PublishableAccuracy              bool `json:"publishable_accuracy"`
}

type ragProvenanceAuditedSource struct {
	Book                   string                           `json:"book"`
	ArtifactPath           string                           `json:"artifact_path"`
	ArtifactSHA256         string                           `json:"artifact_sha256"`
	PDFPages               int                              `json:"pdf_pages"`
	PDFTitle               string                           `json:"pdf_title"`
	PDFAuthor              string                           `json:"pdf_author"`
	PDFCreator             string                           `json:"pdf_creator"`
	PDFProducer            string                           `json:"pdf_producer"`
	PDFCreationDate        string                           `json:"pdf_creation_date"`
	PDFModificationDate    string                           `json:"pdf_modification_date"`
	MarkdownFiles          int                              `json:"markdown_files"`
	MarkdownManifestScheme string                           `json:"markdown_manifest_scheme"`
	MarkdownManifestSHA256 string                           `json:"markdown_manifest_sha256"`
	MachineMarkers         ragProvenanceMachineMarkers      `json:"machine_markers"`
	ArtifactKind           string                           `json:"artifact_kind"`
	ProvenanceStatus       string                           `json:"provenance_status"`
	IndependenceStatus     string                           `json:"independence_status"`
	CoverageStatus         string                           `json:"coverage_status"`
	CoverageObservation    ragProvenanceCoverageObservation `json:"coverage_observation"`
	DecisionBasis          []string                         `json:"decision_basis"`
}

type ragProvenanceCoverageObservation struct {
	Method                       string   `json:"method"`
	Normalization                string   `json:"normalization"`
	PageSplit                    string   `json:"page_split"`
	AnchorSampling               string   `json:"anchor_sampling"`
	BestPageOrder                string   `json:"best_page_order"`
	AmbiguityDefinition          string   `json:"ambiguity_definition"`
	NonMonotonicDefinition       string   `json:"non_monotonic_definition"`
	CandidateFirstChapter        int      `json:"candidate_first_markdown_chapter"`
	CandidateLastChapter         int      `json:"candidate_last_markdown_chapter"`
	UnmappedPrefixChapters       []int    `json:"unmapped_prefix_chapters"`
	UnmappedTailStart            int      `json:"unmapped_tail_start"`
	UnmappedTailEnd              int      `json:"unmapped_tail_end"`
	UnmappedTailCount            int      `json:"unmapped_tail_count"`
	PDFTerminalMissingVolumes    []string `json:"pdf_terminal_missing_volumes"`
	AnchorWidthRunes             int      `json:"anchor_width_runes"`
	MaximumAnchorsPerChapter     int      `json:"maximum_anchors_per_chapter"`
	ZeroHitChapters              int      `json:"zero_hit_chapters"`
	ScoreBelow010Chapters        int      `json:"score_below_0_10_chapters"`
	Score010ToBelow025Chapters   int      `json:"score_0_10_to_below_0_25_chapters"`
	Score025ToBelow050Chapters   int      `json:"score_0_25_to_below_0_50_chapters"`
	ScoreAtLeast050Chapters      int      `json:"score_at_least_0_50_chapters"`
	ExactTitleMatchChapters      int      `json:"exact_title_match_chapters"`
	AmbiguousBestPageChapters    int      `json:"ambiguous_best_page_chapters"`
	NonMonotonicBestPageChapters int      `json:"non_monotonic_best_page_chapters"`
}

type ragProvenanceMachineMarkers struct {
	NumberedChapterHeaders     int `json:"numbered_chapter_headers"`
	CircleHeadingMarkers       int `json:"circle_heading_markers"`
	LuckclubFooterOccurrences  int `json:"luckclub_footer_occurrences"`
	MarkdownModernInsightFiles int `json:"markdown_modern_insight_files"`
}

func TestRAGCitationSourceCatalogContract(t *testing.T) {
	raw := ragCitationMustRead(t, ragCitationCatalogPath)
	if got := ragCitationSHA256(raw); got != ragCitationCatalogSHA256 {
		t.Fatalf("RAG source catalog SHA-256 = %s, want %s", got, ragCitationCatalogSHA256)
	}

	var catalog ragCitationCatalog
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		t.Fatalf("decode RAG source catalog: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("RAG source catalog must contain exactly one JSON document: %v", err)
	}
	if catalog.Schema != "bazi_rag_source_catalog_v1" || catalog.Version != "2026-07-17.3" ||
		strings.TrimSpace(catalog.Description) == "" {
		t.Fatalf("unexpected catalog identity: schema=%q version=%q", catalog.Schema, catalog.Version)
	}
	if catalog.DefaultPolicy.SourceTier != "bronze_unverified" ||
		catalog.DefaultPolicy.VerificationStatus != "source_catalog_missing" ||
		catalog.DefaultPolicy.ArtifactKind != "unregistered" ||
		catalog.DefaultPolicy.ProvenanceStatus != "source_catalog_missing" ||
		catalog.DefaultPolicy.IndependenceStatus != "unknown" ||
		catalog.DefaultPolicy.CoverageStatus != "unknown" ||
		catalog.DefaultPolicy.ClaimEligible {
		t.Fatalf("default source policy must fail closed: %+v", catalog.DefaultPolicy)
	}

	expectedArtifacts := map[string]string{
		"三命通会":  "63eb2a85036ebbd360b815a58780edd242bda0fd1a5faaac7413e00c5f726d47",
		"渊海子平":  "57a130f26a4d45abd0f706405c7f9de00a8e90b6d4630676370f504ebbe2a0f5",
		"滴天髓阐微": "65c67d88421319fccbba23bce88d61d4ace288a7913edd6a10ebf3143e72a48b",
		"穷通宝鉴":  "1c7ef872809bd17ce55607343680489832d58a39b8e2e3584429484d2fd02219",
	}
	if len(catalog.Sources) != len(expectedArtifacts) {
		t.Fatalf("catalog source count = %d, want %d", len(catalog.Sources), len(expectedArtifacts))
	}
	seen := make(map[string]bool, len(catalog.Sources))
	for _, source := range catalog.Sources {
		expectedSHA, ok := expectedArtifacts[source.Book]
		if !ok || seen[source.Book] {
			t.Fatalf("unexpected or duplicate catalog source %q", source.Book)
		}
		seen[source.Book] = true
		if source.MarkdownRoot != source.Book || source.ArtifactPath != "library/"+source.Book+".pdf" ||
			source.ArtifactSHA256 != expectedSHA {
			t.Fatalf("source identity mismatch for %q: %+v", source.Book, source)
		}
		if source.Author != "unrecorded" || source.PageMappingStatus != "not_available" || source.ClaimEligible ||
			source.SourceTier != "classical_text_local" ||
			source.LicenseScope != "internal_research_review_required" || len(source.Limitations) < 4 {
			t.Fatalf("incomplete source must remain ineligible for %q: %+v", source.Book, source)
		}
		if source.Book == "三命通会" {
			if source.Edition != "unrecorded_local_pdf" || source.ArtifactKind != "legacy_text_pdf" ||
				source.ProvenanceStatus != "bibliographic_provenance_unverified" ||
				source.IndependenceStatus != "independence_from_markdown_unverified" ||
				source.CoverageStatus != "partial_pdf_volumes_10_12_missing" ||
				source.VerificationStatus != "artifact_hash_verified_bibliography_and_page_mapping_unavailable" {
				t.Fatalf("legacy PDF provenance mismatch: %+v", source)
			}
		} else if source.Edition != "same_corpus_web_export_2026_unverified" ||
			source.ArtifactKind != "chromium_web_export" ||
			source.ProvenanceStatus != "same_corpus_web_export_detected" ||
			source.IndependenceStatus != "not_independent_from_markdown" ||
			source.CoverageStatus != "same_corpus_export_chapter_structure_observed" ||
			source.VerificationStatus != "artifact_hash_verified_same_corpus_export_not_independent" {
			t.Fatalf("same-corpus export provenance mismatch for %q: %+v", source.Book, source)
		}
		cleanPath := filepath.Clean(source.ArtifactPath)
		if filepath.IsAbs(cleanPath) || cleanPath == "." || cleanPath == ".." ||
			strings.HasPrefix(cleanPath, ".."+string(filepath.Separator)) {
			t.Fatalf("artifact path escapes repository: %q", source.ArtifactPath)
		}
		artifact := ragCitationMustRead(t, filepath.Join(ragCitationRepositoryRoot, cleanPath))
		if got := ragCitationSHA256(artifact); got != expectedSHA {
			t.Fatalf("artifact SHA-256 for %q = %s, want %s", source.Book, got, expectedSHA)
		}
	}
}

func TestRAGSourceProvenanceAuditContract(t *testing.T) {
	raw := ragCitationMustRead(t, ragProvenanceAuditPath)
	if got := ragCitationSHA256(raw); got != ragProvenanceAuditSHA256 {
		t.Fatalf("RAG provenance audit SHA-256 = %s, want %s", got, ragProvenanceAuditSHA256)
	}
	var audit ragProvenanceAudit
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&audit); err != nil {
		t.Fatalf("decode provenance audit: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("provenance audit must contain exactly one JSON document: %v", err)
	}
	if audit.Schema != "bazi_rag_source_provenance_audit_v1" || audit.Version != "2026-07-17.2" ||
		audit.Status != "machine_observed_not_bibliographically_adjudicated" ||
		audit.CatalogPath != "research/rag/bazi-source-catalog-v1.json" ||
		audit.CatalogSHA256 != ragCitationCatalogSHA256 {
		t.Fatalf("unexpected provenance audit identity: %+v", audit)
	}
	if audit.ObservedWith.PDFInfo != "26.05.0" || audit.ObservedWith.PDFToText != "26.07.0" ||
		audit.ObservedWith.TextMode != "layout" {
		t.Fatalf("unexpected provenance observation tools: %+v", audit.ObservedWith)
	}
	if audit.ClaimBoundary.BibliographicMetadataAdjudicated ||
		audit.ClaimBoundary.IndependentPrimarySourceVerified || audit.ClaimBoundary.PageMappingVerified ||
		audit.ClaimBoundary.ClaimSupportReviewed || audit.ClaimBoundary.PublishableAccuracy {
		t.Fatalf("provenance audit must not publish evidence claims: %+v", audit.ClaimBoundary)
	}

	expectedPages := map[string]int{"三命通会": 390, "渊海子平": 777, "滴天髓阐微": 245, "穷通宝鉴": 362}
	expectedMarkdown := map[string]int{"三命通会": 382, "渊海子平": 304, "滴天髓阐微": 66, "穷通宝鉴": 114}
	expectedChapterHeaders := map[string]int{"渊海子平": 305, "滴天髓阐微": 66, "穷通宝鉴": 114}
	expectedLastChapter := map[string]int{"渊海子平": 304, "滴天髓阐微": 65, "穷通宝鉴": 113}
	expectedArtifactSHA := map[string]string{
		"三命通会":  "63eb2a85036ebbd360b815a58780edd242bda0fd1a5faaac7413e00c5f726d47",
		"渊海子平":  "57a130f26a4d45abd0f706405c7f9de00a8e90b6d4630676370f504ebbe2a0f5",
		"滴天髓阐微": "65c67d88421319fccbba23bce88d61d4ace288a7913edd6a10ebf3143e72a48b",
		"穷通宝鉴":  "1c7ef872809bd17ce55607343680489832d58a39b8e2e3584429484d2fd02219",
	}
	expectedMarkdownManifestSHA := map[string]string{
		"三命通会":  "ddef8e89beac4d106336d17e47899e8220b139c36117524b3ea71518dbb91bb7",
		"渊海子平":  "2276be59ae7a18f9f458bac758643ef899871b38cdf7f1c4514189fd18e80f21",
		"滴天髓阐微": "de6c5205abd352b0fcc8a8abbe6666b6bbba16ead340c6052a9e793c75e7e92d",
		"穷通宝鉴":  "ba0c82fc1e625616795ea879dc0a3a3833f4e0e1496eed03b6bf259ea5804512",
	}
	if len(audit.Sources) != len(expectedPages) {
		t.Fatalf("provenance source count = %d, want %d", len(audit.Sources), len(expectedPages))
	}
	for _, source := range audit.Sources {
		if source.ArtifactPath != "library/"+source.Book+".pdf" ||
			source.ArtifactSHA256 != expectedArtifactSHA[source.Book] ||
			source.PDFPages != expectedPages[source.Book] || source.MarkdownFiles != expectedMarkdown[source.Book] ||
			source.MarkdownManifestScheme != "sorted_filename_tab_sha256_lf_v1" ||
			source.MarkdownManifestSHA256 != expectedMarkdownManifestSHA[source.Book] ||
			strings.TrimSpace(source.PDFTitle) == "" || strings.TrimSpace(source.PDFCreator) == "" ||
			strings.TrimSpace(source.PDFProducer) == "" || strings.TrimSpace(source.PDFCreationDate) == "" ||
			strings.TrimSpace(source.PDFModificationDate) == "" || len(source.DecisionBasis) < 3 {
			t.Fatalf("incomplete provenance observation for %q: %+v", source.Book, source)
		}
		artifact := ragCitationMustRead(t, filepath.Join(ragCitationRepositoryRoot, source.ArtifactPath))
		if got := ragCitationSHA256(artifact); got != source.ArtifactSHA256 {
			t.Fatalf("provenance artifact SHA-256 for %q = %s, want %s", source.Book, got, source.ArtifactSHA256)
		}
		if source.Book == "三命通会" {
			coverage := source.CoverageObservation
			bucketTotal := coverage.ZeroHitChapters + coverage.ScoreBelow010Chapters +
				coverage.Score010ToBelow025Chapters + coverage.Score025ToBelow050Chapters +
				coverage.ScoreAtLeast050Chapters
			if source.PDFAuthor != "Administrator" || source.ArtifactKind != "legacy_text_pdf" ||
				source.ProvenanceStatus != "bibliographic_provenance_unverified" ||
				source.IndependenceStatus != "independence_from_markdown_unverified" ||
				source.CoverageStatus != "partial_pdf_volumes_10_12_missing" ||
				source.MachineMarkers.CircleHeadingMarkers != 298 ||
				coverage.Method != "terminal_missing_volume_notice_plus_20_rune_anchor_candidates_v1" ||
				coverage.Normalization != "unicode_nfkc_keep_letters_and_numbers" ||
				coverage.PageSplit != "form_feed_discard_single_trailing_empty_page" ||
				coverage.AnchorSampling != "unique_evenly_spaced_fixed_width" ||
				coverage.BestPageOrder != "anchor_hits_desc_title_presence_desc_physical_page_asc" ||
				coverage.AmbiguityDefinition != "best_and_second_anchor_hit_counts_equal" ||
				coverage.NonMonotonicDefinition != "best_physical_page_below_prior_running_max" ||
				coverage.CandidateFirstChapter != 2 || coverage.CandidateLastChapter != 303 ||
				len(coverage.UnmappedPrefixChapters) != 1 || coverage.UnmappedPrefixChapters[0] != 1 ||
				coverage.UnmappedTailStart != 304 || coverage.UnmappedTailEnd != 381 || coverage.UnmappedTailCount != 78 ||
				strings.Join(coverage.PDFTerminalMissingVolumes, ",") != "卷十,卷十一,卷十二" ||
				coverage.AnchorWidthRunes != 20 || coverage.MaximumAnchorsPerChapter != 48 ||
				coverage.ZeroHitChapters != 76 || coverage.ScoreBelow010Chapters != 3 ||
				coverage.Score010ToBelow025Chapters != 9 || coverage.Score025ToBelow050Chapters != 34 ||
				coverage.ScoreAtLeast050Chapters != 259 || bucketTotal != 381 ||
				coverage.ExactTitleMatchChapters != 198 || coverage.AmbiguousBestPageChapters != 97 ||
				coverage.NonMonotonicBestPageChapters != 78 ||
				bytes.Contains(artifact, []byte("/Creator (Chromium)")) {
				t.Fatalf("legacy PDF provenance evidence mismatch: %+v", source)
			}
			continue
		}
		if source.PDFAuthor != "" || source.PDFCreator != "Chromium" || source.PDFProducer != "Skia/PDF m145" ||
			source.ArtifactKind != "chromium_web_export" ||
			source.ProvenanceStatus != "same_corpus_web_export_detected" ||
			source.IndependenceStatus != "not_independent_from_markdown" ||
			source.CoverageStatus != "same_corpus_export_chapter_structure_observed" ||
			source.CoverageObservation.Method != "same_corpus_numbered_chapter_headers" ||
			source.CoverageObservation.CandidateFirstChapter != 0 ||
			source.CoverageObservation.CandidateLastChapter != expectedLastChapter[source.Book] ||
			source.CoverageObservation.UnmappedTailCount != 0 ||
			source.MachineMarkers.NumberedChapterHeaders != expectedChapterHeaders[source.Book] ||
			source.MachineMarkers.LuckclubFooterOccurrences == 0 ||
			!bytes.Contains(artifact, []byte("/Creator (Chromium)")) {
			t.Fatalf("same-corpus web export evidence mismatch for %q: %+v", source.Book, source)
		}
	}
}

func TestRAGCitationCrossPackageGoSyntaxContract(t *testing.T) {
	paths := []string{
		"../../model/dto.go",
		"../../handler/test/interpretation_test.go",
		"../interpretation/bazi.go",
		"../interpretation/bazi_test.go",
		"../localrag/index.go",
		"../localrag/retriever.go",
		"../localrag/retriever_test.go",
		"../../../cmd/bazi-rag-index/main.go",
	}
	files := token.NewFileSet()
	for _, path := range paths {
		raw := ragCitationMustRead(t, path)
		if _, err := parser.ParseFile(files, path, raw, parser.AllErrors); err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
	}
}

func TestRAGCitationMetadataPropagationContract(t *testing.T) {
	indexSource := string(ragCitationMustRead(t, "../localrag/index.go"))
	retrieverSource := string(ragCitationMustRead(t, "../localrag/retriever.go"))
	manifestSource := string(ragCitationMustRead(t, "../../../../scripts/build-ragflow-bazi-manifest.sh"))
	interpretationSource := string(ragCitationMustRead(t, "../interpretation/bazi.go"))
	dtoSource := string(ragCitationMustRead(t, "../../model/dto.go"))
	vueAPISource := string(ragCitationMustRead(t, "../../../../vue/src/api/interpretation.ts"))
	vuePanelSource := string(ragCitationMustRead(t, "../../../../vue/src/components/ClassicalInterpretationPanel.vue"))
	apiSource := string(ragCitationMustRead(t, "../../../../API.md"))

	metadataFields := []string{
		"author", "edition", "volume", "chapter", "page", "locator", "artifact_path",
		"artifact_sha256", "document_sha256", "source_tier", "verification_status",
		"artifact_kind", "provenance_status", "independence_status", "coverage_status",
		"catalog_claim_eligible", "catalog_schema", "catalog_version", "catalog_sha256",
	}
	for _, field := range metadataFields {
		if !strings.Contains(indexSource, field) {
			t.Fatalf("local index is missing metadata field %q", field)
		}
		if !strings.Contains(retrieverSource, "c."+field) || !strings.Contains(retrieverSource, `"`+field+`"`) {
			t.Fatalf("local retriever does not select and publish metadata field %q", field)
		}
		if !strings.Contains(manifestSource, field) {
			t.Fatalf("RAGFlow manifest is missing metadata field %q", field)
		}
	}

	outputFields := []string{
		"book", "author", "edition", "volume", "chapter", "page", "locator", "path",
		"artifact_path", "artifact_sha256", "document_sha256", "quote", "quote_sha256",
		"source_tier", "verification_status", "artifact_kind", "provenance_status", "independence_status", "coverage_status",
		"catalog_schema", "catalog_version", "catalog_sha256",
		"claim_eligible", "score",
	}
	for _, field := range outputFields {
		if !strings.Contains(dtoSource, `json:"`+field+`"`) {
			t.Fatalf("Go interpretation DTO is missing JSON field %q", field)
		}
		if !strings.Contains(vueAPISource, field+":") {
			t.Fatalf("Vue interpretation DTO is missing field %q", field)
		}
		if !strings.Contains(apiSource, `"`+field+`"`) {
			t.Fatalf("API citation example is missing field %q", field)
		}
	}
	for _, field := range []string{"author", "edition", "locator", "verification_status", "artifact_kind", "provenance_status", "independence_status", "coverage_status", "claim_eligible", "quote_sha256", "catalog_schema", "catalog_version", "catalog_sha256"} {
		if !strings.Contains(vuePanelSource, "citation."+field) {
			t.Fatalf("citation panel does not expose audit field %q", field)
		}
	}

	requiredInterpretationFragments := []string{
		`ReasonCitationMetadataIncomplete = "citation_metadata_incomplete"`,
		`ReasonCitationNotSupporting      = "citation_not_supporting_claim"`,
		`meta["catalog_claim_eligible"]`,
		`citation.VerificationStatus == "bibliography_page_mapping_and_support_verified"`,
		`validClaimArtifactKind(citation.ArtifactKind)`,
		`citation.ProvenanceStatus == "bibliographic_provenance_verified"`,
		`citation.IndependenceStatus == "independent_primary_artifact_verified"`,
		`citation.CoverageStatus == "complete_primary_text_verified"`,
		`citation.CatalogSchema == "bazi_rag_source_catalog_v1"`,
		`citation.CatalogVersion != ""`,
		`validInterpretationSHA256(citation.CatalogSHA256)`,
		`if citation.ClaimEligible && citation.Path != ""`,
		`if citationID == 0`,
		`if !sentenceMatches(sentence, keywords)`,
	}
	for _, fragment := range requiredInterpretationFragments {
		if !strings.Contains(interpretationSource, fragment) {
			t.Fatalf("interpretation fail-closed contract is missing %q", fragment)
		}
	}
	for _, forbidden := range []string{"firstCitationIDs", "fallbackCitationIDs", "defaultCitationIDs"} {
		if strings.Contains(interpretationSource, forbidden) {
			t.Fatalf("unmatched evidence must not receive automatic citations: found %q", forbidden)
		}
	}
}

func TestRAGCitationResearchDocumentsContract(t *testing.T) {
	markers := []string{
		"第一百四十二项完成 RAG 来源目录与可定位引用失败关闭治理",
		"第一百四十三项完成 RAG 来源 provenance 与独立性失败关闭治理",
		"第一百四十四项完成 RAG 来源覆盖范围与缺卷失败关闭治理",
		"第一百四十五项完成完整独立来源候选登记与许可失败关闭治理",
	}
	for _, path := range []string{
		"../../../../docs/fortune-accuracy-research-plan.md",
		"../../../../docs/fortune-accuracy-roadmap.md",
		"../../../../docs/precision-test-plan.md",
	} {
		content := string(ragCitationMustRead(t, path))
		for _, marker := range markers {
			if count := strings.Count(content, marker); count != 1 {
				t.Fatalf("%s marker %q count = %d, want 1", path, marker, count)
			}
		}
		for _, required := range []string{
			"bazi_rag_source_catalog_v1",
			ragCitationCatalogSHA256,
			ragProvenanceAuditSHA256,
			ragExternalCandidateRegistrySHA256,
			"not_independent_from_markdown",
			"citation_metadata_incomplete",
			"citation_not_supporting_claim",
			"claim_eligible",
		} {
			if !strings.Contains(content, required) {
				t.Fatalf("%s is missing phase 142 evidence %q", path, required)
			}
		}
	}
}

func ragCitationMustRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}

func ragCitationSHA256(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
