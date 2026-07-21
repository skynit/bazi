package bazi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	ragCommonsSnapshotRoot           = "../../../../research/rag/snapshots/sanming-ncl-06589-1578-v1"
	ragCommonsSnapshotManifestSHA256 = "a0db6189460aa495122db8809d70a9837a7e92d30f209a11f7839859b3f6c2b3"
	ragCommonsDiscoveryAuditSHA256   = "1958f96c18c58a9665414f147c17169d5f6fdd3575c2d289ef01d48551f25c7e"
	ragCommonsBoundaryAuditSHA256    = "bab12b9be839ce8205cc22705529cc6c9fcaab2168241338bba3aff48642f8e2"
	ragCommonsAttributionSHA256      = "6ba961757cea68b64a81a7c114cbf8d1872c07b0e450c05682e756febbccc941"
	ragCommonsSearchEvidenceSHA256   = "74d454268323b4fc1f34573add6a701467a7b20edea7c36f9d61e5ce12139dd6"
	ragCommonsCandidateID            = "sanming-ncl-06589-1578-12vol-scan-v1"
)

type ragCommonsSnapshotManifest struct {
	Schema                string                          `json:"schema"`
	Version               string                          `json:"version"`
	Status                string                          `json:"status"`
	RetrievedAt           string                          `json:"retrieved_at"`
	CandidateID           string                          `json:"candidate_id"`
	Provider              string                          `json:"provider"`
	SourceInstitution     string                          `json:"source_institution"`
	Work                  string                          `json:"work"`
	Author                string                          `json:"author"`
	Edition               string                          `json:"edition"`
	BibliographicIdentity ragCommonsBibliographicIdentity `json:"bibliographic_identity"`
	License               ragCommonsLicense               `json:"license"`
	DiscoveryAudit        ragCommonsArtifactReference     `json:"discovery_audit"`
	Aggregate             ragCommonsSnapshotAggregate     `json:"aggregate"`
	Boundaries            ragCommonsSnapshotBoundaries    `json:"boundaries"`
	Files                 []ragCommonsSnapshotFile        `json:"files"`
}

type ragCommonsBibliographicIdentity struct {
	CallNumber     string `json:"call_number"`
	AlternateTitle string `json:"alternate_title"`
	Extent         string `json:"extent"`
	Classification string `json:"classification"`
	Collation      string `json:"collation"`
	SourceURL      string `json:"source_url"`
}

type ragCommonsLicense struct {
	Status              string   `json:"status"`
	CommonsTemplate     string   `json:"commons_template"`
	UnderlyingWork      string   `json:"underlying_work"`
	DigitalReproduction string   `json:"digital_reproduction"`
	UsageTerms          string   `json:"usage_terms"`
	AttributionRequired bool     `json:"attribution_required"`
	EvidencePaths       []string `json:"evidence_paths"`
	EvidenceURLs        []string `json:"evidence_urls"`
}

type ragCommonsArtifactReference struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type ragCommonsSnapshotAggregate struct {
	FileCount              int    `json:"file_count"`
	TotalSizeBytes         int64  `json:"total_size_bytes"`
	TotalPhysicalPages     int    `json:"total_physical_pages"`
	ArtifactManifestScheme string `json:"artifact_manifest_scheme"`
	ArtifactManifestSHA256 string `json:"artifact_manifest_sha256"`
}

type ragCommonsSnapshotBoundaries struct {
	ScanArtifactVerified               bool `json:"scan_artifact_verified"`
	StableRemoteIdentityVerified       bool `json:"stable_remote_identity_verified"`
	LicenseTermsResolved               bool `json:"license_terms_resolved"`
	BibliographicMetadataFrozen        bool `json:"bibliographic_metadata_frozen"`
	CompleteStructureObserved          bool `json:"complete_structure_observed"`
	LocalArtifactFrozen                bool `json:"local_artifact_frozen"`
	BibliographyAdjudicated            bool `json:"bibliography_adjudicated"`
	IndependentPrimaryArtifactVerified bool `json:"independent_primary_artifact_verified"`
	CompletePrimaryTextVerified        bool `json:"complete_primary_text_verified"`
	PageMappingVerified                bool `json:"page_mapping_verified"`
	ClaimSupportReviewed               bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed            bool `json:"runtime_ingestion_allowed"`
	ClaimEligible                      bool `json:"claim_eligible"`
	PublishableAccuracy                bool `json:"publishable_accuracy"`
}

type ragCommonsSnapshotFile struct {
	Part                  int                         `json:"part"`
	CommonsPageID         int                         `json:"commons_page_id"`
	Title                 string                      `json:"title"`
	FixedFilePageRevision ragCommonsFixedPageRevision `json:"fixed_file_page_revision"`
	RemoteFile            ragCommonsRemoteFile        `json:"remote_file"`
	LocalArtifact         ragCommonsLocalArtifact     `json:"local_artifact"`
	LicenseEvidence       ragCommonsLicenseEvidence   `json:"license_evidence"`
}

type ragCommonsFixedPageRevision struct {
	RevisionID int    `json:"revision_id"`
	ParentID   int    `json:"parent_id"`
	Timestamp  string `json:"timestamp"`
	SizeBytes  int64  `json:"size_bytes"`
	SHA1       string `json:"sha1"`
	SourceURL  string `json:"source_url"`
}

type ragCommonsRemoteFile struct {
	Timestamp          string                      `json:"timestamp"`
	SizeBytes          int64                       `json:"size_bytes"`
	PageCount          int                         `json:"page_count"`
	SHA1               string                      `json:"sha1"`
	MIME               string                      `json:"mime"`
	DownloadURL        string                      `json:"download_url"`
	DescriptionURL     string                      `json:"description_url"`
	PhysicalPageRange  ragCommonsPhysicalPageRange `json:"physical_page_range"`
	PageLocatorPattern string                      `json:"page_locator_pattern"`
}

type ragCommonsPhysicalPageRange struct {
	First int `json:"first"`
	Last  int `json:"last"`
}

type ragCommonsLocalArtifact struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	PageCount int    `json:"page_count"`
	SHA1      string `json:"sha1"`
	SHA256    string `json:"sha256"`
}

type ragCommonsLicenseEvidence struct {
	Path      string `json:"path"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type ragCommonsDiscoveryAudit struct {
	Schema          string                         `json:"schema"`
	Version         string                         `json:"version"`
	Status          string                         `json:"status"`
	ObservedAt      string                         `json:"observed_at"`
	Work            string                         `json:"work"`
	Purpose         string                         `json:"purpose"`
	Query           ragCommonsDiscoveryQuery       `json:"query"`
	Observation     ragCommonsDiscoveryObservation `json:"observation"`
	CandidateMatrix []ragCommonsCandidateDecision  `json:"candidate_matrix"`
	Selection       ragCommonsSelection            `json:"selection"`
	Boundaries      ragCommonsDiscoveryBoundaries  `json:"boundaries"`
}

type ragCommonsDiscoveryQuery struct {
	Provider          string `json:"provider"`
	API               string `json:"api"`
	Action            string `json:"action"`
	List              string `json:"list"`
	Search            string `json:"search"`
	Namespaces        []int  `json:"namespaces"`
	Limit             int    `json:"limit"`
	ResponsePath      string `json:"response_path"`
	ResponseSizeBytes int64  `json:"response_size_bytes"`
	ResponseSHA256    string `json:"response_sha256"`
}

type ragCommonsDiscoveryObservation struct {
	TotalHits      int      `json:"total_hits"`
	ReturnedHits   int      `json:"returned_hits"`
	FileHits       int      `json:"file_hits"`
	CategoryHits   int      `json:"category_hits"`
	SelectedTitles []string `json:"selected_titles"`
}

type ragCommonsCandidateDecision struct {
	Candidate      string `json:"candidate"`
	TitleHits      int    `json:"title_hits"`
	Edition        string `json:"edition"`
	Structure      string `json:"structure"`
	License        string `json:"license"`
	SourceIdentity string `json:"source_identity"`
	Decision       string `json:"decision"`
}

type ragCommonsSelection struct {
	CandidateID string   `json:"candidate_id"`
	Reasons     []string `json:"reasons"`
}

type ragCommonsDiscoveryBoundaries struct {
	ProviderMetadataIsNotFinalBibliographicAdjudication bool `json:"provider_metadata_is_not_final_bibliographic_adjudication"`
	TitleLevelCompletenessIsNotPageLevelCompleteness    bool `json:"title_level_completeness_is_not_page_level_completeness"`
	LocalSnapshotDoesNotImplyRuntimeIngestion           bool `json:"local_snapshot_does_not_imply_runtime_ingestion"`
	BibliographyAdjudicated                             bool `json:"bibliography_adjudicated"`
	CompletePrimaryTextVerified                         bool `json:"complete_primary_text_verified"`
	PageMappingVerified                                 bool `json:"page_mapping_verified"`
	ClaimSupportReviewed                                bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed                             bool `json:"runtime_ingestion_allowed"`
	ClaimEligible                                       bool `json:"claim_eligible"`
	PublishableAccuracy                                 bool `json:"publishable_accuracy"`
}

type ragCommonsBoundaryAudit struct {
	Schema                      string                         `json:"schema"`
	Version                     string                         `json:"version"`
	Status                      string                         `json:"status"`
	ObservedAt                  string                         `json:"observed_at"`
	CandidateID                 string                         `json:"candidate_id"`
	SnapshotManifest            ragCommonsArtifactReference    `json:"snapshot_manifest"`
	Method                      ragCommonsBoundaryMethod       `json:"method"`
	ObservedCoverCandidates     []ragCommonsCoverCandidate     `json:"observed_cover_candidates"`
	PhysicalBookCandidateRanges []ragCommonsBookCandidateRange `json:"physical_book_candidate_ranges"`
	VisualSpotCheck             ragCommonsVisualSpotCheck      `json:"visual_spot_check"`
	Result                      ragCommonsBoundaryResult       `json:"result"`
	Boundaries                  ragCommonsBoundaryLimitations  `json:"boundaries"`
}

type ragCommonsBoundaryMethod struct {
	Renderer        string  `json:"renderer"`
	RenderArguments string  `json:"render_arguments"`
	MeasurementTool string  `json:"measurement_tool"`
	Measurement     string  `json:"measurement"`
	CandidateRule   string  `json:"candidate_rule"`
	Threshold       float64 `json:"threshold"`
}

type ragCommonsCoverCandidate struct {
	Part           int     `json:"part"`
	PhysicalPage   int     `json:"physical_page"`
	MeanLuminance  float64 `json:"mean_luminance"`
	Classification string  `json:"classification"`
}

type ragCommonsBookCandidateRange struct {
	BookCandidate int                     `json:"book_candidate"`
	Segments      []ragCommonsBookSegment `json:"segments"`
}

type ragCommonsBookSegment struct {
	Part      int `json:"part"`
	FirstPage int `json:"first_page"`
	LastPage  int `json:"last_page"`
}

type ragCommonsVisualSpotCheck struct {
	Status            string                        `json:"status"`
	SampledRanges     []ragCommonsVisualSampleRange `json:"sampled_ranges"`
	IndependentReview bool                          `json:"independent_review"`
}

type ragCommonsVisualSampleRange struct {
	Part        int    `json:"part"`
	FirstPage   int    `json:"first_page"`
	LastPage    int    `json:"last_page"`
	Observation string `json:"observation"`
}

type ragCommonsBoundaryResult struct {
	DarkCoverCandidateCount    int                `json:"dark_cover_candidate_count"`
	PhysicalBookCandidateCount int                `json:"physical_book_candidate_count"`
	CrossPDFBookCandidate      int                `json:"cross_pdf_book_candidate"`
	FinalBookCandidateStart    ragCommonsPartPage `json:"final_book_candidate_start"`
	ProviderExtentConsistent   bool               `json:"provider_extent_consistent_with_observation"`
}

type ragCommonsPartPage struct {
	Part         int `json:"part"`
	PhysicalPage int `json:"physical_page"`
}

type ragCommonsBoundaryLimitations struct {
	CoverDetectionIsNotVolumeLabelReading            bool `json:"cover_detection_is_not_volume_label_reading"`
	SequenceNumberIsCandidateNotVerifiedVolumeNumber bool `json:"sequence_number_is_a_candidate_not_verified_volume_number"`
	CompleteStructureObserved                        bool `json:"complete_structure_observed"`
	CompletePrimaryTextVerified                      bool `json:"complete_primary_text_verified"`
	VolumeMappingVerified                            bool `json:"volume_mapping_verified"`
	ChapterPageMappingVerified                       bool `json:"chapter_page_mapping_verified"`
	ClaimSupportReviewed                             bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed                          bool `json:"runtime_ingestion_allowed"`
	ClaimEligible                                    bool `json:"claim_eligible"`
	PublishableAccuracy                              bool `json:"publishable_accuracy"`
}

func TestRAGCommonsNCLScanSnapshotContract(t *testing.T) {
	manifest := ragCommonsReadStrictJSON[ragCommonsSnapshotManifest](t, filepath.Join(ragCommonsSnapshotRoot, "snapshot-manifest.json"), ragCommonsSnapshotManifestSHA256)
	if manifest.Schema != "commons_scan_snapshot_v1" || manifest.Version != "2026-07-17.1" ||
		manifest.Status != "research_snapshot_not_runtime_eligible" || manifest.RetrievedAt != "2026-07-17" ||
		manifest.CandidateID != ragCommonsCandidateID || manifest.Provider != "commons.wikimedia.org" ||
		manifest.SourceInstitution != "National Central Library, Republic of China (Taiwan)" ||
		manifest.Work != "三命通會" || manifest.Author != "萬民英" || manifest.Edition != "明萬曆戊寅(六年, 1578)刊本" {
		t.Fatalf("unexpected Commons snapshot identity: %+v", manifest)
	}
	ragCommonsValidateBibliographyAndLicense(t, manifest)
	ragCommonsValidateSnapshotFiles(t, manifest)
	ragCommonsValidateDiscovery(t, manifest)
	ragCommonsValidateBookBoundaries(t)
	ragCommonsValidateScripts(t)

	if got := ragCommonsHashFile(t, filepath.Join(ragCommonsSnapshotRoot, "ATTRIBUTION.md")); got != ragCommonsAttributionSHA256 {
		t.Fatalf("source notice SHA-256 = %s, want %s", got, ragCommonsAttributionSHA256)
	}
}

func TestRAGCommonsNCLScanIsNotRuntimeRegistered(t *testing.T) {
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
			ragCommonsCandidateID,
			"commons_scan_snapshot_v1",
			"sanming-ncl-06589-1578-v1",
		} {
			if bytes.Contains(raw, []byte(forbidden)) {
				t.Fatalf("research-only Commons snapshot marker %q leaked into %s", forbidden, sourcePath)
			}
		}
	}
}

func TestRAGCommonsNCLScanResearchDocumentsContract(t *testing.T) {
	marker := "第一百四十八项完成 Commons/NCL 1578 年完整扫描研究快照治理"
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
			t.Fatalf("phase 148 marker count in %s = %d, want 1", documentPath, count)
		}
	}
}

func ragCommonsValidateBibliographyAndLicense(t *testing.T, manifest ragCommonsSnapshotManifest) {
	t.Helper()
	bibliography := manifest.BibliographicIdentity
	if bibliography.CallNumber != "306.5 06589" || bibliography.AlternateTitle != "三命會通" ||
		bibliography.Extent != "十二卷，12冊，線裝" || bibliography.Classification != "子部-術數類-命相之屬" ||
		bibliography.Collation != "10行，行22字，雙欄，版心白口，單魚尾" ||
		bibliography.SourceURL != "https://rbook.ncl.edu.tw/NCLSearch" {
		t.Fatalf("unexpected NCL bibliography: %+v", bibliography)
	}
	license := manifest.License
	if license.Status != "terms_resolved_public_domain_scan" || license.CommonsTemplate != "PD-scan|PD-old" ||
		license.UnderlyingWork != "PD-old" || license.DigitalReproduction != "PD-scan" ||
		license.UsageTerms != "Public domain" || license.AttributionRequired || len(license.EvidencePaths) != 2 ||
		len(license.EvidenceURLs) != 4 || license.EvidenceURLs[2] != "https://commons.wikimedia.org/wiki/Template:PD-scan" ||
		license.EvidenceURLs[3] != "https://commons.wikimedia.org/wiki/Template:PD-old" {
		t.Fatalf("unexpected Commons license evidence: %+v", license)
	}
	boundaries := manifest.Boundaries
	if !boundaries.ScanArtifactVerified || !boundaries.StableRemoteIdentityVerified || !boundaries.LicenseTermsResolved ||
		!boundaries.BibliographicMetadataFrozen || !boundaries.CompleteStructureObserved || !boundaries.LocalArtifactFrozen ||
		boundaries.BibliographyAdjudicated || boundaries.IndependentPrimaryArtifactVerified ||
		boundaries.CompletePrimaryTextVerified || boundaries.PageMappingVerified || boundaries.ClaimSupportReviewed ||
		boundaries.RuntimeIngestionAllowed || boundaries.ClaimEligible || boundaries.PublishableAccuracy {
		t.Fatalf("Commons snapshot boundaries must fail closed: %+v", boundaries)
	}
}

func ragCommonsValidateSnapshotFiles(t *testing.T, manifest ragCommonsSnapshotManifest) {
	t.Helper()
	expected := []struct {
		pageID, revisionID, parentID, revisionSize, pages                  int
		title, revisionTime, revisionSHA1, fileTime, fileSHA1, localSHA256 string
		size                                                               int64
	}{
		{138125281, 1207461339, 1138624022, 5414, 1000, "File:NCL-06589 1 三命通會.pdf", "2026-05-03T01:58:40Z", "a009fba1a2909653d5cc1f55a2e6d022f76d316f", "2023-09-26T02:32:46Z", "c222bc54815d8e5cef15338c03b9fc11d540f41a", "3de5c45efb1919965afae00a3c97121054aaabd0f38f9fd5ab8f0a28bb8e36dd", 101956385},
		{138043642, 1207461333, 804277976, 2208, 187, "File:NCL-06589 2 三命通會.pdf", "2026-05-03T01:58:39Z", "2c97f0f2ea3cf51ec861585e23b18698bc9e0304", "2023-09-24T12:54:27Z", "090ce9d53d9aa37bd7d2680b3b69c16385edbe2d", "3a309831d5b0f0396ac8c89c0176734c560ceaafb03e71b1457f45662de16eb8", 18967335},
	}
	if len(manifest.Files) != 2 {
		t.Fatalf("snapshot file count = %d, want 2", len(manifest.Files))
	}
	var artifactRows strings.Builder
	var totalSize int64
	var totalPages int
	for index, file := range manifest.Files {
		want := expected[index]
		if file.Part != index+1 || file.CommonsPageID != want.pageID || file.Title != want.title ||
			file.FixedFilePageRevision.RevisionID != want.revisionID || file.FixedFilePageRevision.ParentID != want.parentID ||
			file.FixedFilePageRevision.Timestamp != want.revisionTime || file.FixedFilePageRevision.SizeBytes != int64(want.revisionSize) ||
			file.FixedFilePageRevision.SHA1 != want.revisionSHA1 || file.RemoteFile.Timestamp != want.fileTime ||
			file.RemoteFile.SizeBytes != want.size || file.RemoteFile.PageCount != want.pages ||
			file.RemoteFile.SHA1 != want.fileSHA1 || file.RemoteFile.MIME != "application/pdf" ||
			file.LocalArtifact.SizeBytes != want.size || file.LocalArtifact.PageCount != want.pages ||
			file.LocalArtifact.SHA1 != want.fileSHA1 || file.LocalArtifact.SHA256 != want.localSHA256 ||
			file.RemoteFile.PhysicalPageRange != (ragCommonsPhysicalPageRange{First: 1, Last: want.pages}) ||
			file.LicenseEvidence.SHA256 != "65435ab54769d4dd0e05405578a555ec6b4b76208af82474e8fe1acf24890ce1" {
			t.Fatalf("unexpected Commons snapshot file %d: %+v", index+1, file)
		}
		ragCommonsValidateHTTPS(t, file.FixedFilePageRevision.SourceURL, "commons.wikimedia.org")
		ragCommonsValidateHTTPS(t, file.RemoteFile.DownloadURL, "upload.wikimedia.org")
		ragCommonsValidateHTTPS(t, file.RemoteFile.DescriptionURL, "commons.wikimedia.org")
		if file.RemoteFile.PageLocatorPattern != fmt.Sprintf("https://commons.wikimedia.org/wiki/Special:Redirect/page/%d?page={physical_page}", want.pageID) {
			t.Fatalf("unexpected page locator pattern: %s", file.RemoteFile.PageLocatorPattern)
		}

		artifactPath := ragCommonsJoinSnapshotPath(t, file.LocalArtifact.Path)
		info, err := os.Stat(artifactPath)
		if err != nil || info.Size() != want.size {
			t.Fatalf("local artifact identity for part %d: info=%v err=%v", index+1, info, err)
		}
		if got := ragCommonsHashFile(t, artifactPath); got != want.localSHA256 {
			t.Fatalf("local artifact SHA-256 for part %d = %s, want %s", index+1, got, want.localSHA256)
		}
		rawPrefix := make([]byte, 5)
		artifact, err := os.Open(artifactPath)
		if err != nil {
			t.Fatal(err)
		}
		_, readErr := io.ReadFull(artifact, rawPrefix)
		closeErr := artifact.Close()
		if readErr != nil || closeErr != nil || string(rawPrefix) != "%PDF-" {
			t.Fatalf("artifact part %d is not an unambiguous PDF: prefix=%q read=%v close=%v", index+1, rawPrefix, readErr, closeErr)
		}

		evidencePath := ragCommonsJoinSnapshotPath(t, file.LicenseEvidence.Path)
		evidence, err := os.ReadFile(evidencePath)
		if err != nil {
			t.Fatal(err)
		}
		if int64(len(evidence)) != file.LicenseEvidence.SizeBytes || ragCommonsSHA256(evidence) != file.LicenseEvidence.SHA256 {
			t.Fatalf("license evidence identity mismatch for part %d", index+1)
		}
		for _, marker := range []string{"明萬曆戊寅(六年, 1578)刊本", "卷數：十二卷", "數量：12冊", "索書號：306.5 06589", "{{PD-scan|PD-old}}"} {
			if !bytes.Contains(evidence, []byte(marker)) {
				t.Fatalf("license evidence part %d missing %q", index+1, marker)
			}
		}

		fmt.Fprintf(&artifactRows, "%d\t%d\t%d\t%s\t%d\t%d\t%s\t%s\n", file.Part, file.CommonsPageID,
			file.FixedFilePageRevision.RevisionID, file.RemoteFile.Timestamp, file.RemoteFile.SizeBytes,
			file.RemoteFile.PageCount, file.RemoteFile.SHA1, file.LocalArtifact.SHA256)
		totalSize += want.size
		totalPages += want.pages
	}
	aggregate := manifest.Aggregate
	if aggregate.FileCount != 2 || aggregate.TotalSizeBytes != totalSize || aggregate.TotalPhysicalPages != totalPages ||
		aggregate.ArtifactManifestScheme != "part_tab_commons_page_id_tab_revision_id_tab_file_timestamp_tab_size_tab_page_count_tab_remote_sha1_tab_local_sha256_lf_v1" ||
		aggregate.ArtifactManifestSHA256 != ragCommonsSHA256([]byte(artifactRows.String())) ||
		aggregate.ArtifactManifestSHA256 != "996974d6fc6d6cee9815292ffa733ffc61b87fa7ae147b7ed8e4d0c2039e8c3d" {
		t.Fatalf("unexpected Commons snapshot aggregate: %+v", aggregate)
	}
}

func ragCommonsValidateDiscovery(t *testing.T, manifest ragCommonsSnapshotManifest) {
	t.Helper()
	if manifest.DiscoveryAudit.Path != "discovery-audit.json" || manifest.DiscoveryAudit.SHA256 != ragCommonsDiscoveryAuditSHA256 {
		t.Fatalf("unexpected discovery reference: %+v", manifest.DiscoveryAudit)
	}
	audit := ragCommonsReadStrictJSON[ragCommonsDiscoveryAudit](t, filepath.Join(ragCommonsSnapshotRoot, manifest.DiscoveryAudit.Path), ragCommonsDiscoveryAuditSHA256)
	if audit.Schema != "sanming_complete_scan_discovery_audit_v1" || audit.Version != "2026-07-17.1" ||
		audit.Status != "selected_for_research_snapshot_not_runtime_ingestion" || audit.ObservedAt != "2026-07-17" ||
		audit.Work != "三命通會" || strings.TrimSpace(audit.Purpose) == "" || audit.Selection.CandidateID != ragCommonsCandidateID ||
		len(audit.Selection.Reasons) != 5 || len(audit.CandidateMatrix) != 4 {
		t.Fatalf("unexpected discovery audit identity: %+v", audit)
	}
	query := audit.Query
	if query.Provider != "commons.wikimedia.org" || query.API != "https://commons.wikimedia.org/w/api.php" ||
		query.Action != "query" || query.List != "search" || query.Search != "intitle:\"三命通會\"" ||
		fmt.Sprint(query.Namespaces) != "[6 14]" || query.Limit != 500 || query.ResponsePath != "evidence/commons-title-search.json" ||
		query.ResponseSizeBytes != 61069 || query.ResponseSHA256 != ragCommonsSearchEvidenceSHA256 {
		t.Fatalf("unexpected Commons discovery query: %+v", query)
	}
	observation := audit.Observation
	if observation.TotalHits != 91 || observation.ReturnedHits != 91 || observation.FileHits != 89 || observation.CategoryHits != 2 ||
		fmt.Sprint(observation.SelectedTitles) != "[File:NCL-06589 1 三命通會.pdf File:NCL-06589 2 三命通會.pdf]" {
		t.Fatalf("unexpected Commons discovery observation: %+v", observation)
	}
	boundaries := audit.Boundaries
	if !boundaries.ProviderMetadataIsNotFinalBibliographicAdjudication || !boundaries.TitleLevelCompletenessIsNotPageLevelCompleteness ||
		!boundaries.LocalSnapshotDoesNotImplyRuntimeIngestion || boundaries.BibliographyAdjudicated ||
		boundaries.CompletePrimaryTextVerified || boundaries.PageMappingVerified || boundaries.ClaimSupportReviewed ||
		boundaries.RuntimeIngestionAllowed || boundaries.ClaimEligible || boundaries.PublishableAccuracy {
		t.Fatalf("discovery audit boundaries must fail closed: %+v", boundaries)
	}

	searchPath := ragCommonsJoinSnapshotPath(t, query.ResponsePath)
	if got := ragCommonsHashFile(t, searchPath); got != ragCommonsSearchEvidenceSHA256 {
		t.Fatalf("Commons search evidence SHA-256 = %s, want %s", got, ragCommonsSearchEvidenceSHA256)
	}
	var searchEvidence struct {
		Query struct {
			SearchInfo struct {
				TotalHits int `json:"totalhits"`
			} `json:"searchinfo"`
			Search []struct {
				NS     int    `json:"ns"`
				Title  string `json:"title"`
				PageID int    `json:"pageid"`
			} `json:"search"`
		} `json:"query"`
	}
	raw, err := os.ReadFile(searchPath)
	if err != nil || json.Unmarshal(raw, &searchEvidence) != nil {
		t.Fatalf("decode Commons search evidence: %v", err)
	}
	if searchEvidence.Query.SearchInfo.TotalHits != 91 || len(searchEvidence.Query.Search) != 91 {
		t.Fatalf("unexpected frozen Commons search result size")
	}
	wantedTitles := map[string]int{"File:NCL-06589 1 三命通會.pdf": 138125281, "File:NCL-06589 2 三命通會.pdf": 138043642}
	var fileHits, categoryHits int
	for _, hit := range searchEvidence.Query.Search {
		if hit.NS == 6 {
			fileHits++
		} else if hit.NS == 14 {
			categoryHits++
		}
		if wantPageID, ok := wantedTitles[hit.Title]; ok {
			if hit.PageID != wantPageID {
				t.Fatalf("Commons selected title %q page ID = %d, want %d", hit.Title, hit.PageID, wantPageID)
			}
			delete(wantedTitles, hit.Title)
		}
	}
	if fileHits != 89 || categoryHits != 2 || len(wantedTitles) != 0 {
		t.Fatalf("frozen Commons search classification mismatch: files=%d categories=%d missing=%v", fileHits, categoryHits, wantedTitles)
	}
}

func ragCommonsValidateBookBoundaries(t *testing.T) {
	t.Helper()
	audit := ragCommonsReadStrictJSON[ragCommonsBoundaryAudit](t, filepath.Join(ragCommonsSnapshotRoot, "book-boundary-audit.json"), ragCommonsBoundaryAuditSHA256)
	if audit.Schema != "sanming_ncl_physical_book_boundary_audit_v1" || audit.Version != "2026-07-17.1" ||
		audit.Status != "physical_book_candidates_not_volume_mapping" || audit.ObservedAt != "2026-07-17" ||
		audit.CandidateID != ragCommonsCandidateID || audit.SnapshotManifest.Path != "snapshot-manifest.json" ||
		audit.SnapshotManifest.SHA256 != ragCommonsSnapshotManifestSHA256 || audit.Method.Threshold != 0.30 ||
		audit.Method.CandidateRule != "mean_luminance < threshold" || !strings.Contains(audit.Method.Renderer, "pdftoppm version") ||
		!strings.Contains(audit.Method.MeasurementTool, "ImageMagick") {
		t.Fatalf("unexpected book-boundary audit identity: %+v", audit)
	}
	expectedCoverPages := map[int][]int{
		1: {1, 98, 99, 190, 191, 257, 258, 334, 335, 425, 426, 523, 524, 639, 640, 759, 760, 876, 877, 959, 960},
		2: {70, 71, 187},
	}
	actualCoverPages := map[int][]int{1: {}, 2: {}}
	for _, candidate := range audit.ObservedCoverCandidates {
		if candidate.Classification != "dark_outer_cover_candidate" || candidate.MeanLuminance >= audit.Method.Threshold ||
			(candidate.Part != 1 && candidate.Part != 2) || candidate.PhysicalPage <= 0 {
			t.Fatalf("invalid dark-cover candidate: %+v", candidate)
		}
		actualCoverPages[candidate.Part] = append(actualCoverPages[candidate.Part], candidate.PhysicalPage)
	}
	if fmt.Sprint(actualCoverPages[1]) != fmt.Sprint(expectedCoverPages[1]) || fmt.Sprint(actualCoverPages[2]) != fmt.Sprint(expectedCoverPages[2]) {
		t.Fatalf("cover candidates mismatch: %+v", actualCoverPages)
	}
	if len(audit.PhysicalBookCandidateRanges) != 12 {
		t.Fatalf("physical book candidate count = %d, want 12", len(audit.PhysicalBookCandidateRanges))
	}
	wantStarts := []ragCommonsBookSegment{{1, 1, 98}, {1, 99, 190}, {1, 191, 257}, {1, 258, 334}, {1, 335, 425}, {1, 426, 523}, {1, 524, 639}, {1, 640, 759}, {1, 760, 876}, {1, 877, 959}}
	for index := range wantStarts {
		rangeCandidate := audit.PhysicalBookCandidateRanges[index]
		if rangeCandidate.BookCandidate != index+1 || len(rangeCandidate.Segments) != 1 || rangeCandidate.Segments[0] != wantStarts[index] {
			t.Fatalf("unexpected physical book range %d: %+v", index+1, rangeCandidate)
		}
	}
	if fmt.Sprint(audit.PhysicalBookCandidateRanges[10].Segments) != "[{1 960 1000} {2 1 70}]" ||
		fmt.Sprint(audit.PhysicalBookCandidateRanges[11].Segments) != "[{2 71 187}]" {
		t.Fatalf("unexpected cross-PDF/final physical book ranges: %+v", audit.PhysicalBookCandidateRanges[10:])
	}
	if audit.VisualSpotCheck.Status != "single_operator_non_independent" || audit.VisualSpotCheck.IndependentReview || len(audit.VisualSpotCheck.SampledRanges) != 2 ||
		audit.Result.DarkCoverCandidateCount != 24 || audit.Result.PhysicalBookCandidateCount != 12 ||
		audit.Result.CrossPDFBookCandidate != 11 || audit.Result.FinalBookCandidateStart != (ragCommonsPartPage{Part: 2, PhysicalPage: 71}) ||
		!audit.Result.ProviderExtentConsistent {
		t.Fatalf("unexpected physical-book audit result: spot=%+v result=%+v", audit.VisualSpotCheck, audit.Result)
	}
	boundaries := audit.Boundaries
	if !boundaries.CoverDetectionIsNotVolumeLabelReading || !boundaries.SequenceNumberIsCandidateNotVerifiedVolumeNumber ||
		!boundaries.CompleteStructureObserved || boundaries.CompletePrimaryTextVerified || boundaries.VolumeMappingVerified ||
		boundaries.ChapterPageMappingVerified || boundaries.ClaimSupportReviewed || boundaries.RuntimeIngestionAllowed ||
		boundaries.ClaimEligible || boundaries.PublishableAccuracy {
		t.Fatalf("physical-book audit boundaries must fail closed: %+v", boundaries)
	}
}

func ragCommonsValidateScripts(t *testing.T) {
	t.Helper()
	checks := map[string][]string{
		"../../../../scripts/fetch-commons-ncl-sanmintonghui-snapshot.sh": {
			"set -euo pipefail", "revids=$revision_id", "{{PD-scan|PD-old}}", "sha1sum", "sha256sum", "pdfinfo",
			"c222bc54815d8e5cef15338c03b9fc11d540f41a", "090ce9d53d9aa37bd7d2680b3b69c16385edbe2d",
		},
		"../../../../scripts/audit-ncl-sanmintonghui-book-boundaries.sh": {
			"set -euo pipefail", ragCommonsSnapshotManifestSHA256, "mean_luminance < threshold", "physical_book_candidates_not_volume_mapping",
		},
	}
	for scriptPath, markers := range checks {
		raw, err := os.ReadFile(scriptPath)
		if err != nil {
			t.Fatal(err)
		}
		for _, marker := range markers {
			if !bytes.Contains(raw, []byte(marker)) {
				t.Fatalf("script %s missing %q", scriptPath, marker)
			}
		}
	}
}

func ragCommonsReadStrictJSON[T any](t *testing.T, filePath, expectedSHA256 string) T {
	t.Helper()
	raw, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if got := ragCommonsSHA256(raw); got != expectedSHA256 {
		t.Fatalf("%s SHA-256 = %s, want %s", filePath, got, expectedSHA256)
	}
	var value T
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decode %s: %v", filePath, err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("%s must contain exactly one JSON document: %v", filePath, err)
	}
	return value
}

func ragCommonsJoinSnapshotPath(t *testing.T, relativePath string) string {
	t.Helper()
	if relativePath == "" || filepath.IsAbs(relativePath) || filepath.Clean(relativePath) != relativePath ||
		relativePath == "." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) {
		t.Fatalf("unsafe snapshot-relative path: %q", relativePath)
	}
	return filepath.Join(ragCommonsSnapshotRoot, relativePath)
}

func ragCommonsValidateHTTPS(t *testing.T, rawURL, host string) {
	t.Helper()
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host != host {
		t.Fatalf("invalid HTTPS URL %q for host %q", rawURL, host)
	}
}

func ragCommonsHashFile(t *testing.T, filePath string) string {
	t.Helper()
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatal(err)
	}
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		file.Close()
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func ragCommonsSHA256(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
