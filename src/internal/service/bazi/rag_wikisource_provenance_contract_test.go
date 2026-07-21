package bazi

import (
	"bytes"
	"fmt"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const (
	ragWikisourceProvenanceRoot         = "../../../../research/rag/provenance/sanming-siku-wikisource-provenance-v1"
	ragWikisourceProvenanceAuditPath    = ragWikisourceProvenanceRoot + "/provenance-audit.json"
	ragWikisourceProvenanceAuditSHA256  = "e1fe3bd056bfa6c6c662f2a0ed04cadc01547feca2944076fc4b6fcf3ffa5215"
	ragWikisourceRevisionHistoryPath    = ragWikisourceProvenanceRoot + "/revision-history.json"
	ragWikisourceRevisionHistorySHA256  = "3516ebd04d36c2b8d896868bada8921963092fe48ecf68135317326c11e10f54"
	ragWikisourceProvenanceScriptSHA256 = "4667c6ce07defec692ee6bc920c513faeb0b78ae1278ccd27c271fe08c233898"
)

type ragWikisourceProvenanceAudit struct {
	Schema                      string                            `json:"schema"`
	Version                     string                            `json:"version"`
	Status                      string                            `json:"status"`
	ObservedAt                  string                            `json:"observed_at"`
	CandidateID                 string                            `json:"candidate_id"`
	Purpose                     string                            `json:"purpose"`
	Inputs                      ragWikisourceProvenanceInputs     `json:"inputs"`
	Method                      ragWikisourceProvenanceMethod     `json:"method"`
	License                     ragWikisourceProvenanceLicense    `json:"license"`
	FixedEvidence               []ragWikisourceFixedEvidence      `json:"fixed_evidence"`
	HistorySummary              ragWikisourceHistorySummary       `json:"history_summary"`
	ImportProvenanceObservation ragWikisourceImportObservation    `json:"import_provenance_observation"`
	FrozenVolumeObservations    []ragWikisourceVolumeObservation  `json:"frozen_volume_observations"`
	AggregateContentMarkers     ragWikisourceAggregateMarkers     `json:"aggregate_content_markers"`
	Decision                    ragWikisourceProvenanceDecision   `json:"decision"`
	Boundaries                  ragWikisourceProvenanceBoundaries `json:"boundaries"`
	Limitations                 []string                          `json:"limitations"`
}

type ragWikisourceProvenanceInputs struct {
	SnapshotManifestPath        string `json:"snapshot_manifest_path"`
	SnapshotManifestSHA256      string `json:"snapshot_manifest_sha256"`
	CrossSourceComparisonPath   string `json:"cross_source_comparison_path"`
	CrossSourceComparisonSHA256 string `json:"cross_source_comparison_sha256"`
	RevisionHistoryPath         string `json:"revision_history_path"`
	RevisionHistorySHA256       string `json:"revision_history_sha256"`
}

type ragWikisourceProvenanceMethod struct {
	FixedRevisionAPI              string `json:"fixed_revision_api"`
	HistoryAPI                    string `json:"history_api"`
	LocalMarkerScan               string `json:"local_marker_scan"`
	ExcludesCurrentTemplateRender bool   `json:"excludes_current_template_render"`
	ExcludesUnlicensedScans       bool   `json:"excludes_unlicensed_scan_downloads"`
}

type ragWikisourceProvenanceLicense struct {
	UnderlyingWork       string `json:"underlying_work"`
	DigitalContributions string `json:"digital_contributions"`
	LicenseURL           string `json:"license_url"`
	AttributionRequired  bool   `json:"attribution_required"`
}

type ragWikisourceFixedEvidence struct {
	Role         string `json:"role"`
	Title        string `json:"title"`
	PageID       string `json:"page_id"`
	RevisionID   int    `json:"revision_id"`
	ParentID     int    `json:"parent_id"`
	Timestamp    string `json:"timestamp"`
	User         string `json:"user"`
	UserID       int    `json:"user_id"`
	Comment      string `json:"comment"`
	RemoteSize   int64  `json:"remote_size"`
	RemoteSHA1   string `json:"remote_sha1"`
	SourceURL    string `json:"source_url"`
	ArtifactPath string `json:"artifact_path"`
	LocalSize    int64  `json:"local_size"`
	LocalSHA256  string `json:"local_sha256"`
}

type ragWikisourceHistorySummary struct {
	Pages                                 int      `json:"pages"`
	Revisions                             int      `json:"revisions"`
	FirstRevisionImporter                 string   `json:"first_revision_importer"`
	FirstRevisionImporterUserID           int      `json:"first_revision_importer_user_id"`
	PagesWithContentHashChangedSinceFirst int      `json:"pages_with_content_hash_changed_since_first"`
	FixedRevisionByImporter               int      `json:"fixed_revision_by_importer"`
	FixedRevisionByWmrBot                 int      `json:"fixed_revision_by_wmr_bot"`
	FixedRevisionByCrowleyBot             int      `json:"fixed_revision_by_crowley_bot"`
	FixedCommentClasses                   []string `json:"fixed_comment_classes"`
	FixedCommentExamples                  []string `json:"fixed_comment_examples"`
}

type ragWikisourceImportObservation struct {
	UpstreamDescription                              string   `json:"upstream_description"`
	UpstreamURL                                      string   `json:"upstream_url"`
	UpstreamIdentifier                               string   `json:"upstream_identifier"`
	UpstreamHash                                     string   `json:"upstream_hash"`
	ScanEditionIdentifier                            string   `json:"scan_edition_identifier"`
	LibraryCatalogIdentifier                         string   `json:"library_catalog_identifier"`
	PlanDirectExternalURLs                           []string `json:"plan_direct_external_urls"`
	PlanUpstreamArchiveURLCount                      int      `json:"plan_upstream_archive_url_count"`
	PlanStatesOnlyTextWithoutImagesOrTables          bool     `json:"plan_states_only_text_without_images_or_tables"`
	PlanStatesIAScansAreFutureManualInputs           bool     `json:"plan_states_ia_scans_are_future_manual_inputs"`
	GuidelineStatesImagesAndTablesRequireScanCompare bool     `json:"guideline_states_images_and_tables_require_future_scan_comparison"`
	GuidelineStates3951RareCharactersRequireManual   bool     `json:"guideline_states_3951_rare_characters_require_manual_identification"`
}

type ragWikisourceVolumeObservation struct {
	Volume             int    `json:"volume"`
	ArtifactPath       string `json:"artifact_path"`
	ArtifactSHA256     string `json:"artifact_sha256"`
	ScanNoticeCount    int    `json:"scan_notice_count"`
	SKCharCount        int    `json:"skchar_count"`
	DirectFileLinks    int    `json:"direct_file_links"`
	DirectPageLinks    int    `json:"direct_page_links"`
	DirectExternalURLs int    `json:"direct_external_urls"`
}

type ragWikisourceAggregateMarkers struct {
	VolumesWithScanNotice    int `json:"volumes_with_scan_comparison_notice"`
	UnresolvedSKChar         int `json:"unresolved_skchar_placeholders"`
	DirectFileLinks          int `json:"direct_file_links"`
	DirectPageNamespaceLinks int `json:"direct_page_namespace_links"`
	DirectExternalURLs       int `json:"direct_external_urls"`
}

type ragWikisourceProvenanceDecision struct {
	ArtifactKind           string `json:"artifact_kind"`
	ProvenanceStatus       string `json:"provenance_status"`
	IndependenceStatus     string `json:"independence_status"`
	ScanIdentityStatus     string `json:"scan_identity_status"`
	ScanProofreadingStatus string `json:"scan_proofreading_status"`
	CoverageStatus         string `json:"coverage_status"`
	PromotionResult        string `json:"promotion_result"`
}

type ragWikisourceProvenanceBoundaries struct {
	FixedRevisionIdentityVerified   bool `json:"fixed_revision_identity_verified"`
	CompleteTextStructureObserved   bool `json:"complete_text_structure_observed"`
	BibliographicProvenanceVerified bool `json:"bibliographic_provenance_verified"`
	IndependentPrimaryArtifact      bool `json:"independent_primary_artifact_verified"`
	ScanArtifactVerified            bool `json:"scan_artifact_verified"`
	PageMappingVerified             bool `json:"page_mapping_verified"`
	ClaimSupportReviewed            bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed         bool `json:"runtime_ingestion_allowed"`
	ClaimEligible                   bool `json:"claim_eligible"`
	PublishableAccuracy             bool `json:"publishable_accuracy"`
}

type ragWikisourceRevisionHistory struct {
	Schema  string                             `json:"schema"`
	Version string                             `json:"version"`
	Status  string                             `json:"status"`
	Pages   []ragWikisourceRevisionHistoryPage `json:"pages"`
}

type ragWikisourceRevisionHistoryPage struct {
	Role                         string                       `json:"role"`
	Volume                       int                          `json:"volume"`
	PageID                       string                       `json:"page_id"`
	Title                        string                       `json:"title"`
	FixedRevisionID              int                          `json:"fixed_revision_id"`
	RevisionCount                int                          `json:"revision_count"`
	Contributors                 []string                     `json:"contributors"`
	Comments                     []string                     `json:"comments"`
	ContentHashChangedSinceFirst bool                         `json:"content_hash_changed_since_first"`
	Revisions                    []ragWikisourceRevisionEntry `json:"revisions"`
}

type ragWikisourceRevisionEntry struct {
	RevisionID int      `json:"revision_id"`
	ParentID   int      `json:"parent_id"`
	Minor      bool     `json:"minor"`
	User       string   `json:"user"`
	UserID     int      `json:"user_id"`
	Timestamp  string   `json:"timestamp"`
	Size       int64    `json:"size"`
	SHA1       string   `json:"sha1"`
	Comment    string   `json:"comment"`
	Tags       []string `json:"tags"`
}

func TestRAGWikisourceTranscriptionProvenanceContract(t *testing.T) {
	raw := ragSnapshotMustRead(t, ragWikisourceProvenanceAuditPath)
	if got := ragSnapshotSHA256(raw); got != ragWikisourceProvenanceAuditSHA256 {
		t.Fatalf("Wikisource provenance audit SHA-256 = %s, want %s", got, ragWikisourceProvenanceAuditSHA256)
	}
	audit := ragSnapshotDecodeStrict[ragWikisourceProvenanceAudit](t, raw)
	if audit.Schema != "sanming_wikisource_transcription_provenance_audit_v1" || audit.Version != "2026-07-17.1" ||
		audit.Status != "bulk_imported_transcription_upstream_archive_unidentified" || audit.ObservedAt != "2026-07-17" ||
		audit.CandidateID != "sanming-siku-wikisource-12vol-v1" || strings.TrimSpace(audit.Purpose) == "" {
		t.Fatalf("unexpected Wikisource provenance identity: %+v", audit)
	}
	inputs := audit.Inputs
	if inputs.SnapshotManifestPath != "research/rag/snapshots/sanming-siku-wikisource-12vol-v1/snapshot-manifest.json" ||
		inputs.SnapshotManifestSHA256 != ragSnapshotManifestSHA256 ||
		inputs.CrossSourceComparisonPath != "research/rag/sanming-wikisource-markdown-comparison-v1.json" ||
		inputs.CrossSourceComparisonSHA256 != ragSnapshotComparisonSHA256 ||
		inputs.RevisionHistoryPath != "revision-history.json" || inputs.RevisionHistorySHA256 != ragWikisourceRevisionHistorySHA256 {
		t.Fatalf("unexpected Wikisource provenance inputs: %+v", inputs)
	}
	if audit.Method.FixedRevisionAPI != "MediaWiki action=query revids plus revisions content" ||
		audit.Method.HistoryAPI != "MediaWiki action=query single page rvdir=newer rvendid=frozen_revision" ||
		audit.Method.LocalMarkerScan != "ripgrep exact patterns over frozen raw wikitext" ||
		!audit.Method.ExcludesCurrentTemplateRender || !audit.Method.ExcludesUnlicensedScans {
		t.Fatalf("unexpected provenance method: %+v", audit.Method)
	}
	if audit.License.UnderlyingWork != "PD-old" || audit.License.DigitalContributions != "CC BY-SA 4.0 and GFDL" ||
		audit.License.LicenseURL != "https://zh.wikisource.org/wiki/Wikisource:版权信息" || !audit.License.AttributionRequired {
		t.Fatalf("unexpected provenance license: %+v", audit.License)
	}
	ragWikisourceValidateFixedEvidence(t, audit.FixedEvidence)
	ragWikisourceValidateImportObservation(t, audit.ImportProvenanceObservation)
	ragWikisourceValidateVolumeObservations(t, audit.FrozenVolumeObservations, audit.AggregateContentMarkers)
	ragWikisourceValidateDecision(t, audit.Decision, audit.Boundaries, audit.Limitations)

	historyRaw := ragSnapshotMustRead(t, ragWikisourceRevisionHistoryPath)
	if got := ragSnapshotSHA256(historyRaw); got != ragWikisourceRevisionHistorySHA256 {
		t.Fatalf("Wikisource revision history SHA-256 = %s, want %s", got, ragWikisourceRevisionHistorySHA256)
	}
	history := ragSnapshotDecodeStrict[ragWikisourceRevisionHistory](t, historyRaw)
	ragWikisourceValidateHistory(t, history, audit.HistorySummary)

	if got := ragSnapshotSHA256(ragSnapshotMustRead(t, "../../../../scripts/audit-wikisource-sanmintonghui-provenance.sh")); got != ragWikisourceProvenanceScriptSHA256 {
		t.Fatalf("Wikisource provenance script SHA-256 = %s, want %s", got, ragWikisourceProvenanceScriptSHA256)
	}
}

func TestRAGWikisourceProvenanceIsNotRuntimeRegistered(t *testing.T) {
	for _, sourcePath := range []string{
		"../localrag/index.go",
		"../localrag/retriever.go",
		"../interpretation/bazi.go",
		"../../model/dto.go",
		"../../../cmd/bazi-rag-index/main.go",
		"../../../../scripts/build-local-bazi-rag-index.sh",
		"../../../../scripts/build-ragflow-bazi-manifest.sh",
	} {
		raw := ragSnapshotMustRead(t, sourcePath)
		for _, forbidden := range []string{
			"sanming_wikisource_transcription_provenance_audit_v1",
			"upstream_text_compressed_archive_unidentified",
			"research/rag/provenance/sanming-siku-wikisource-provenance-v1",
		} {
			if bytes.Contains(raw, []byte(forbidden)) {
				t.Fatalf("research provenance marker %q leaked into runtime source %s", forbidden, sourcePath)
			}
		}
	}
}

func TestRAGWikisourceProvenanceResearchDocumentsContract(t *testing.T) {
	marker := "第一百四十七项完成维基文库转录 provenance 与扫描身份失败关闭治理"
	for _, path := range []string{
		"../../../../docs/fortune-accuracy-research-plan.md",
		"../../../../docs/fortune-accuracy-roadmap.md",
		"../../../../docs/precision-test-plan.md",
	} {
		content := string(ragSnapshotMustRead(t, path))
		if count := strings.Count(content, marker); count != 1 {
			t.Fatalf("%s marker count = %d, want 1", path, count)
		}
		for _, required := range []string{
			ragWikisourceProvenanceAuditSHA256,
			ragWikisourceRevisionHistorySHA256,
			"sanming_wikisource_transcription_provenance_audit_v1",
			"61",
			"438",
			"文本压缩包",
			"not_scan_proofread",
			"claim_eligible=false",
		} {
			if !strings.Contains(content, required) {
				t.Fatalf("%s is missing phase 147 evidence %q", path, required)
			}
		}
	}
}

func ragWikisourceValidateFixedEvidence(t *testing.T, evidence []ragWikisourceFixedEvidence) {
	t.Helper()
	expected := []struct {
		role, title, pageID, timestamp, user, sha1, artifact, localSHA string
		revisionID, parentID, userID                                   int
		size                                                           int64
	}{
		{"work_page", "三命通會 (四庫全書本)", "260853", "2016-10-15T11:19:12Z", "維基小霸王", "0fd68ff5e57220d21be67716a6294fdc6b0ddc63", "evidence/work-page-rev657391.wikitext", "4c9923df0ab644fe7b675796f41195746cc2a07273826116ab8d02ed442bb650", 657391, 631739, 20996, 2492},
		{"import_guideline_active_at_bulk_import", "Wikisource:四庫全書原文", "205610", "2016-09-21T12:28:20Z", "維基小霸王", "27341a89cb4b7235381862c09317be4694ef4e1b", "evidence/import-guideline-rev513112.wikitext", "5c04db3a9967f07bbc0f3cbb30219036698d781f60ac01e20e8748f274af328c", 513112, 0, 20996, 1522},
		{"import_plan_active_at_bulk_import", "User:維基小霸王/录入四库全书计划", "204146", "2016-10-03T13:37:37Z", "維基小霸王", "b6c7b33884b7f09db30d367334e8707aa574743b", "evidence/import-plan-rev534259.wikitext", "e799578008b9b8963c17a1f17281ee301a163b56dd79e4f5a6f4a44dfdad8f6f", 534259, 521165, 20996, 4564},
	}
	if len(evidence) != len(expected) {
		t.Fatalf("fixed evidence count = %d, want %d", len(evidence), len(expected))
	}
	contents := make(map[string][]byte, len(evidence))
	for index, item := range evidence {
		want := expected[index]
		if item.Role != want.role || item.Title != want.title || item.PageID != want.pageID ||
			item.RevisionID != want.revisionID || item.ParentID != want.parentID || item.Timestamp != want.timestamp ||
			item.User != want.user || item.UserID != want.userID || item.RemoteSize != want.size || item.LocalSize != want.size ||
			item.RemoteSHA1 != want.sha1 || item.ArtifactPath != want.artifact || item.LocalSHA256 != want.localSHA {
			t.Fatalf("fixed evidence %d identity mismatch: %+v", index, item)
		}
		parsed, err := url.Parse(item.SourceURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host != "zh.wikisource.org" ||
			parsed.Query().Get("curid") != item.PageID || parsed.Query().Get("oldid") != strconv.Itoa(item.RevisionID) {
			t.Fatalf("fixed evidence %d source URL invalid: %q", index, item.SourceURL)
		}
		artifact := ragSnapshotMustRead(t, filepath.Join(ragWikisourceProvenanceRoot, filepath.FromSlash(item.ArtifactPath)))
		if int64(len(artifact)) != item.LocalSize || ragSnapshotSHA256(artifact) != item.LocalSHA256 {
			t.Fatalf("fixed evidence %d local artifact mismatch", index)
		}
		contents[item.Role] = artifact
	}
	work := string(contents["work_page"])
	if strings.Count(work, "請根據四庫全書掃描版校對本頁") != 1 || strings.Contains(work, "http://") || strings.Contains(work, "https://") {
		t.Fatalf("work page scan notice or direct URL identity changed")
	}
	for volume := 1; volume <= 12; volume++ {
		if !strings.Contains(work, fmt.Sprintf("/卷%02d|", volume)) {
			t.Fatalf("work page missing volume %02d link", volume)
		}
	}
	guideline := string(contents["import_guideline_active_at_bulk_import"])
	for _, required := range []string{
		"導入的內容只包含文本，請根據掃描版本加入圖片",
		"導入的內容只包含文本，請根據掃描版本加入表格",
		"總共有3951個字符",
	} {
		if strings.Count(guideline, required) != 1 {
			t.Fatalf("fixed import guideline missing %q", required)
		}
	}
	plan := string(contents["import_plan_active_at_bulk_import"])
	for _, required := range []string{
		"网上看到了一个包含四库全书书籍的文本压缩包",
		"只有文本，沒有圖片和表格",
		"图片、表格可以用IA的扫描版添加",
	} {
		if strings.Count(plan, required) != 1 {
			t.Fatalf("fixed import plan missing %q", required)
		}
	}
	if strings.Count(plan, "https://") != 2 || strings.Contains(plan, "archive.org") {
		t.Fatalf("fixed import plan unexpectedly identifies an upstream archive")
	}
}

func ragWikisourceValidateImportObservation(t *testing.T, observation ragWikisourceImportObservation) {
	t.Helper()
	wantURLs := []string{
		"https://www.w3.org/International/articles/css3-text/",
		"https://www.w3.org/TR/jlreq/#inline_cutting_note",
	}
	if observation.UpstreamDescription != "网上看到的一个包含四库全书书籍的文本压缩包" ||
		observation.UpstreamURL != "" || observation.UpstreamIdentifier != "" || observation.UpstreamHash != "" ||
		observation.ScanEditionIdentifier != "" || observation.LibraryCatalogIdentifier != "" ||
		fmt.Sprint(observation.PlanDirectExternalURLs) != fmt.Sprint(wantURLs) || observation.PlanUpstreamArchiveURLCount != 0 ||
		!observation.PlanStatesOnlyTextWithoutImagesOrTables || !observation.PlanStatesIAScansAreFutureManualInputs ||
		!observation.GuidelineStatesImagesAndTablesRequireScanCompare || !observation.GuidelineStates3951RareCharactersRequireManual {
		t.Fatalf("unexpected import provenance observation: %+v", observation)
	}
}

func ragWikisourceValidateVolumeObservations(t *testing.T, observations []ragWikisourceVolumeObservation, aggregate ragWikisourceAggregateMarkers) {
	t.Helper()
	wantSKChar := []int{2, 65, 21, 5, 18, 15, 49, 77, 14, 58, 52, 62}
	manifest := ragSnapshotDecodeStrict[ragSnapshotManifest](t, ragSnapshotMustRead(t, ragSnapshotManifestPath))
	if len(observations) != 12 {
		t.Fatalf("volume observation count = %d, want 12", len(observations))
	}
	totalSKChar := 0
	for index, observation := range observations {
		volume := manifest.Volumes[index]
		wantPath := fmt.Sprintf("research/rag/snapshots/sanming-siku-wikisource-12vol-v1/volumes/volume-%02d.wikitext", index+1)
		if observation.Volume != index+1 || observation.ArtifactPath != wantPath || observation.ArtifactSHA256 != volume.LocalSHA256 ||
			observation.ScanNoticeCount != 1 || observation.SKCharCount != wantSKChar[index] ||
			observation.DirectFileLinks != 0 || observation.DirectPageLinks != 0 || observation.DirectExternalURLs != 0 {
			t.Fatalf("volume observation %d invalid: %+v", index+1, observation)
		}
		artifact := ragSnapshotMustRead(t, "../../../../"+observation.ArtifactPath)
		if ragSnapshotSHA256(artifact) != observation.ArtifactSHA256 {
			t.Fatalf("volume observation %d artifact SHA-256 mismatch", index+1)
		}
		totalSKChar += observation.SKCharCount
	}
	if aggregate.VolumesWithScanNotice != 12 || aggregate.UnresolvedSKChar != totalSKChar || totalSKChar != 438 ||
		aggregate.DirectFileLinks != 0 || aggregate.DirectPageNamespaceLinks != 0 || aggregate.DirectExternalURLs != 0 {
		t.Fatalf("unexpected aggregate content markers: %+v", aggregate)
	}
}

func ragWikisourceValidateDecision(t *testing.T, decision ragWikisourceProvenanceDecision, boundaries ragWikisourceProvenanceBoundaries, limitations []string) {
	t.Helper()
	if decision.ArtifactKind != "bulk_imported_plain_text_transcription" ||
		decision.ProvenanceStatus != "upstream_text_compressed_archive_unidentified" ||
		decision.IndependenceStatus != "not_adjudicated_against_local_markdown" ||
		decision.ScanIdentityStatus != "no_scan_identity_or_page_links_in_frozen_pages" ||
		decision.ScanProofreadingStatus != "not_scan_proofread" ||
		decision.CoverageStatus != "complete_twelve_volume_text_structure_observed" || decision.PromotionResult != "blocked" {
		t.Fatalf("unexpected Wikisource provenance decision: %+v", decision)
	}
	if !boundaries.FixedRevisionIdentityVerified || !boundaries.CompleteTextStructureObserved ||
		boundaries.BibliographicProvenanceVerified || boundaries.IndependentPrimaryArtifact || boundaries.ScanArtifactVerified ||
		boundaries.PageMappingVerified || boundaries.ClaimSupportReviewed || boundaries.RuntimeIngestionAllowed ||
		boundaries.ClaimEligible || boundaries.PublishableAccuracy || len(limitations) < 5 {
		t.Fatalf("Wikisource provenance boundary must fail closed: %+v limitations=%v", boundaries, limitations)
	}
}

func ragWikisourceValidateHistory(t *testing.T, history ragWikisourceRevisionHistory, summary ragWikisourceHistorySummary) {
	t.Helper()
	if history.Schema != "wikisource_fixed_revision_history_v1" || history.Version != "2026-07-17.1" ||
		history.Status != "history_through_frozen_revision" || len(history.Pages) != 13 {
		t.Fatalf("unexpected revision history identity: %+v", history)
	}
	manifest := ragSnapshotDecodeStrict[ragSnapshotManifest](t, ragSnapshotMustRead(t, ragSnapshotManifestPath))
	wantCounts := []int{4, 4, 5, 4, 4, 5, 4, 6, 5, 5, 5, 5, 5}
	totalRevisions := 0
	changed := 0
	fixedUsers := map[string]int{}
	for index, page := range history.Pages {
		wantRole := "volume"
		wantVolume := index
		var wantPageID, wantTitle string
		var wantRevision int
		var wantSize int64
		var wantSHA1 string
		if index == 0 {
			wantRole, wantVolume = "work_page", 0
			wantPageID, wantTitle, wantRevision, wantSize, wantSHA1 = "260853", "三命通會 (四庫全書本)", 657391, 2492, "0fd68ff5e57220d21be67716a6294fdc6b0ddc63"
		} else {
			volume := manifest.Volumes[index-1]
			wantPageID, wantTitle, wantRevision, wantSize, wantSHA1 = volume.PageID, volume.Title, volume.RevisionID, volume.RemoteSize, volume.RemoteSHA1
		}
		if page.Role != wantRole || page.Volume != wantVolume || page.PageID != wantPageID || page.Title != wantTitle ||
			page.FixedRevisionID != wantRevision || page.RevisionCount != wantCounts[index] || len(page.Revisions) != page.RevisionCount {
			t.Fatalf("revision history page %d identity invalid: %+v", index, page)
		}
		seenRevisions := map[int]bool{}
		contributors := map[string]bool{}
		comments := map[string]bool{}
		for revisionIndex, revision := range page.Revisions {
			if revision.RevisionID <= 0 || seenRevisions[revision.RevisionID] || revision.User == "" || revision.UserID <= 0 ||
				revision.Timestamp == "" || revision.Size <= 0 || !ragExternalValidHex(revision.SHA1, 40) || revision.Tags == nil {
				t.Fatalf("page %d revision %d invalid: %+v", index, revisionIndex, revision)
			}
			seenRevisions[revision.RevisionID] = true
			contributors[revision.User] = true
			comments[revision.Comment] = true
		}
		first := page.Revisions[0]
		fixed := page.Revisions[len(page.Revisions)-1]
		if first.ParentID != 0 || first.User != "維基小霸王" || first.UserID != 20996 ||
			fixed.RevisionID != wantRevision || fixed.Size != wantSize || fixed.SHA1 != wantSHA1 ||
			page.ContentHashChangedSinceFirst != (first.SHA1 != fixed.SHA1) {
			t.Fatalf("page %d first/fixed revision mismatch", index)
		}
		if page.ContentHashChangedSinceFirst {
			changed++
		}
		fixedUsers[fixed.User]++
		if fmt.Sprint(page.Contributors) != fmt.Sprint(ragWikisourceSortedKeys(contributors)) ||
			fmt.Sprint(page.Comments) != fmt.Sprint(ragWikisourceSortedKeys(comments)) {
			t.Fatalf("page %d contributor/comment summary mismatch", index)
		}
		totalRevisions += page.RevisionCount
	}
	if totalRevisions != 61 || changed != 8 || fixedUsers["維基小霸王"] != 5 || fixedUsers["Wmr-bot"] != 6 || fixedUsers["CrowleyBot"] != 2 {
		t.Fatalf("recomputed revision summary invalid: revisions=%d changed=%d fixed=%v", totalRevisions, changed, fixedUsers)
	}
	wantClasses := []string{"bulk_import", "template_repair", "single_text_correction"}
	wantExamples := []string{"导入1个版本", "repair templates", "辛西→辛酉（[[Special:permalink/2082153|本批详情]]）"}
	if summary.Pages != 13 || summary.Revisions != totalRevisions || summary.FirstRevisionImporter != "維基小霸王" ||
		summary.FirstRevisionImporterUserID != 20996 || summary.PagesWithContentHashChangedSinceFirst != changed ||
		summary.FixedRevisionByImporter != fixedUsers["維基小霸王"] || summary.FixedRevisionByWmrBot != fixedUsers["Wmr-bot"] ||
		summary.FixedRevisionByCrowleyBot != fixedUsers["CrowleyBot"] ||
		fmt.Sprint(summary.FixedCommentClasses) != fmt.Sprint(wantClasses) || fmt.Sprint(summary.FixedCommentExamples) != fmt.Sprint(wantExamples) {
		t.Fatalf("stored revision history summary mismatch: %+v", summary)
	}
}

func ragWikisourceSortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
