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
	"strconv"
	"strings"
	"testing"
)

const (
	ragSnapshotRoot                 = "../../../../research/rag/snapshots/sanming-siku-wikisource-12vol-v1"
	ragSnapshotManifestPath         = ragSnapshotRoot + "/snapshot-manifest.json"
	ragSnapshotManifestSHA256       = "38a03d0ee048a097620de3765f2793e9f0a5383a3238748fcd583ac79bb26974"
	ragSnapshotAttributionPath      = ragSnapshotRoot + "/ATTRIBUTION.md"
	ragSnapshotAttributionSHA256    = "84aad48ff04e5220b0dc999ffda2de4fd5e86157896a7f14b7f2f9511461c6b2"
	ragSnapshotComparisonPath       = "../../../../research/rag/sanming-wikisource-markdown-comparison-v1.json"
	ragSnapshotComparisonSHA256     = "ce7ee55bd4c6b68c488f3326b9b3725860b5ebea0c679ee12d156f330a180b7f"
	ragSnapshotFetchScriptSHA256    = "e10160dea4c879a3df6d3843b3d6863f88f3771f1cd17a95f3dc25f054e40283"
	ragSnapshotAuditGeneratorSHA256 = "4f6fdaec5e9f6b4a2e76e2322c0b1135c76071ca834d92fb9814d5ac693b757f"
)

type ragSnapshotManifest struct {
	Schema         string                `json:"schema"`
	Version        string                `json:"version"`
	Status         string                `json:"status"`
	RetrievedAt    string                `json:"retrieved_at"`
	RegistryPath   string                `json:"registry_path"`
	RegistrySHA256 string                `json:"registry_sha256"`
	CandidateID    string                `json:"candidate_id"`
	Provider       string                `json:"provider"`
	Work           string                `json:"work"`
	Edition        string                `json:"edition"`
	License        ragSnapshotLicense    `json:"license"`
	Boundaries     ragSnapshotBoundaries `json:"boundaries"`
	Volumes        []ragSnapshotVolume   `json:"volumes"`
}

type ragSnapshotLicense struct {
	UnderlyingWork       string `json:"underlying_work"`
	DigitalContributions string `json:"digital_contributions"`
	LicenseURL           string `json:"license_url"`
	AttributionFile      string `json:"attribution_file"`
}

type ragSnapshotBoundaries struct {
	RawWikitextUnmodified   bool `json:"raw_wikitext_unmodified"`
	LocalArtifactFrozen     bool `json:"local_artifact_frozen"`
	IndependenceVerified    bool `json:"independence_verified"`
	BibliographyAdjudicated bool `json:"bibliography_adjudicated"`
	PageMappingVerified     bool `json:"page_mapping_verified"`
	ClaimSupportReviewed    bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed bool `json:"runtime_ingestion_allowed"`
	ClaimEligible           bool `json:"claim_eligible"`
	PublishableAccuracy     bool `json:"publishable_accuracy"`
}

type ragSnapshotVolume struct {
	Volume       int    `json:"volume"`
	Title        string `json:"title"`
	PageID       string `json:"page_id"`
	RevisionID   int    `json:"revision_id"`
	Timestamp    string `json:"timestamp"`
	RemoteSize   int64  `json:"remote_size"`
	RemoteSHA1   string `json:"remote_sha1"`
	SourceURL    string `json:"source_url"`
	ArtifactPath string `json:"artifact_path"`
	LocalSize    int64  `json:"local_size"`
	LocalSHA256  string `json:"local_sha256"`
}

type ragSnapshotComparison struct {
	Schema        string                         `json:"schema"`
	Version       string                         `json:"version"`
	Status        string                         `json:"status"`
	ObservedAt    string                         `json:"observed_at"`
	Purpose       string                         `json:"purpose"`
	Sources       ragSnapshotComparisonSources   `json:"sources"`
	Normalization ragSnapshotNormalization       `json:"normalization"`
	Summary       ragSnapshotComparisonSummary   `json:"summary"`
	Volumes       []ragSnapshotNormalizedVolume  `json:"volumes"`
	Chapters      []ragSnapshotChapterComparison `json:"chapters"`
	Boundaries    ragSnapshotAuditBoundaries     `json:"boundaries"`
}

type ragSnapshotComparisonSources struct {
	SnapshotManifestPath   string `json:"snapshot_manifest_path"`
	SnapshotManifestSHA256 string `json:"snapshot_manifest_sha256"`
	AttributionPath        string `json:"attribution_path"`
	AttributionSHA256      string `json:"attribution_sha256"`
	CandidateID            string `json:"candidate_id"`
	Provider               string `json:"provider"`
	RevisionVolumeCount    int    `json:"revision_volume_count"`
	MarkdownRootLabel      string `json:"markdown_root_label"`
	MarkdownManifestScheme string `json:"markdown_manifest_scheme"`
	MarkdownManifestSHA256 string `json:"markdown_manifest_sha256"`
	MarkdownFileCount      int    `json:"markdown_file_count"`
	NumberedChapterCount   int    `json:"numbered_chapter_count"`
}

type ragSnapshotNormalization struct {
	WikitextExtraction       string            `json:"wikitext_extraction"`
	MarkdownExtraction       string            `json:"markdown_extraction"`
	ScriptConversion         string            `json:"script_conversion"`
	CharacterFilter          string            `json:"character_filter"`
	AnchorSampling           string            `json:"anchor_sampling"`
	AnchorWidthRunes         int               `json:"anchor_width_runes"`
	MaximumAnchorsPerChapter int               `json:"maximum_anchors_per_chapter"`
	OpenCCVersion            string            `json:"opencc_version"`
	OpenCCAssetSHA256        map[string]string `json:"opencc_asset_sha256"`
	LocalCorpusSHA256        string            `json:"local_normalized_chapter_manifest_sha256"`
	RemoteCorpusSHA256       string            `json:"remote_normalized_volume_manifest_sha256"`
}

type ragSnapshotComparisonSummary struct {
	ComparedChapters       int                      `json:"compared_chapters"`
	ZeroHitChapters        int                      `json:"zero_hit_chapters"`
	ScoreBelow010          int                      `json:"score_below_0_10_chapters"`
	Score010ToBelow025     int                      `json:"score_0_10_to_below_0_25_chapters"`
	Score025ToBelow050     int                      `json:"score_0_25_to_below_0_50_chapters"`
	Score050ToBelow080     int                      `json:"score_0_50_to_below_0_80_chapters"`
	ScoreAtLeast080        int                      `json:"score_at_least_0_80_chapters"`
	TitleLocatedChapters   int                      `json:"title_located_chapters"`
	AmbiguousBestVolume    int                      `json:"ambiguous_best_volume_chapters"`
	BestVolumeCounts       []ragSnapshotVolumeCount `json:"best_volume_counts"`
	MachineOverlapObserved bool                     `json:"machine_textual_overlap_observed"`
}

type ragSnapshotVolumeCount struct {
	Volume int `json:"volume"`
	Count  int `json:"count"`
}

type ragSnapshotNormalizedVolume struct {
	Volume           int    `json:"volume"`
	PageID           string `json:"page_id"`
	RevisionID       int    `json:"revision_id"`
	SourceURL        string `json:"source_url"`
	ArtifactPath     string `json:"artifact_path"`
	LocalSHA256      string `json:"local_sha256"`
	NormalizedRunes  int    `json:"normalized_runes"`
	NormalizedSHA256 string `json:"normalized_sha256"`
}

type ragSnapshotChapterComparison struct {
	Chapter               int    `json:"chapter"`
	File                  string `json:"file"`
	Title                 string `json:"title"`
	OriginalSHA256        string `json:"original_sha256"`
	NormalizedRunes       int    `json:"normalized_runes"`
	AnchorCount           int    `json:"anchor_count"`
	BestAnchorHits        int    `json:"best_anchor_hits"`
	ScoreBasisPoints      int    `json:"score_basis_points"`
	BestCandidateVolumes  []int  `json:"best_candidate_volumes"`
	TitleCandidateVolumes []int  `json:"title_candidate_volumes"`
}

type ragSnapshotAuditBoundaries struct {
	MachineComparisonOnly           bool `json:"machine_comparison_only"`
	TextualOverlapIsNotIdentity     bool `json:"textual_overlap_is_not_artifact_identity"`
	TextualOverlapIsNotIndependence bool `json:"textual_overlap_is_not_independence_verification"`
	BibliographyAdjudicated         bool `json:"bibliography_adjudicated"`
	IndependenceVerified            bool `json:"independence_verified"`
	PageMappingVerified             bool `json:"page_mapping_verified"`
	ClaimSupportReviewed            bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed         bool `json:"runtime_ingestion_allowed"`
	ClaimEligible                   bool `json:"claim_eligible"`
	PublishableAccuracy             bool `json:"publishable_accuracy"`
}

func TestRAGWikisourceRevisionSnapshotContract(t *testing.T) {
	raw := ragSnapshotMustRead(t, ragSnapshotManifestPath)
	if got := ragSnapshotSHA256(raw); got != ragSnapshotManifestSHA256 {
		t.Fatalf("snapshot manifest SHA-256 = %s, want %s", got, ragSnapshotManifestSHA256)
	}
	manifest := ragSnapshotDecodeStrict[ragSnapshotManifest](t, raw)
	if manifest.Schema != "wikisource_revision_snapshot_v1" || manifest.Version != "2026-07-17.1" ||
		manifest.Status != "research_snapshot_not_runtime_eligible" || manifest.RetrievedAt != "2026-07-17" ||
		manifest.RegistryPath != "research/rag/bazi-external-source-candidates-v1.json" ||
		manifest.RegistrySHA256 != ragExternalCandidateRegistrySHA256 || manifest.CandidateID != "sanming-siku-wikisource-12vol-v1" ||
		manifest.Provider != "zh.wikisource.org" || manifest.Work != "三命通會" || manifest.Edition != "欽定四庫全書本" {
		t.Fatalf("unexpected snapshot identity: %+v", manifest)
	}
	if manifest.License.UnderlyingWork != "PD-old" || manifest.License.DigitalContributions != "CC BY-SA 4.0 and GFDL" ||
		manifest.License.LicenseURL != "https://zh.wikisource.org/wiki/Wikisource:版权信息" ||
		manifest.License.AttributionFile != "ATTRIBUTION.md" || !ragSnapshotBoundaryFailsClosed(manifest.Boundaries) {
		t.Fatalf("snapshot license or boundary invalid: %+v %+v", manifest.License, manifest.Boundaries)
	}

	registryRaw := ragSnapshotMustRead(t, ragExternalCandidateRegistryPath)
	registry := ragSnapshotDecodeStrict[ragExternalCandidateRegistry](t, registryRaw)
	var candidate *ragExternalCandidate
	for index := range registry.Candidates {
		if registry.Candidates[index].CandidateID == manifest.CandidateID {
			candidate = &registry.Candidates[index]
		}
	}
	if candidate == nil || len(candidate.Volumes) != 12 || len(manifest.Volumes) != 12 {
		t.Fatalf("snapshot candidate or volume count missing")
	}
	for index, volume := range manifest.Volumes {
		expected := candidate.Volumes[index]
		wantFile := fmt.Sprintf("volumes/volume-%02d.wikitext", index+1)
		if volume.Volume != index+1 || volume.Title != expected.Title || volume.PageID != expected.StableID ||
			volume.RevisionID != expected.StableRevision || volume.Timestamp != expected.Timestamp ||
			volume.RemoteSize != expected.SizeBytes || volume.RemoteSHA1 != expected.RemoteSHA1 ||
			volume.SourceURL != expected.SourceURL || volume.ArtifactPath != wantFile || volume.LocalSize != volume.RemoteSize ||
			!ragExternalValidSHA256(volume.LocalSHA256) || expected.Retrieved || expected.LocalSHA256 != "" {
			t.Fatalf("snapshot volume %d diverges from frozen remote identity: %+v / %+v", index+1, volume, expected)
		}
		parsed, err := url.Parse(volume.SourceURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host != "zh.wikisource.org" ||
			parsed.Query().Get("curid") != volume.PageID || parsed.Query().Get("oldid") != strconv.Itoa(volume.RevisionID) {
			t.Fatalf("snapshot volume %d source URL is not revision-addressable: %q", volume.Volume, volume.SourceURL)
		}
		artifact := ragSnapshotMustRead(t, filepath.Join(ragSnapshotRoot, filepath.FromSlash(volume.ArtifactPath)))
		if int64(len(artifact)) != volume.LocalSize || ragSnapshotSHA256(artifact) != volume.LocalSHA256 {
			t.Fatalf("snapshot volume %d local artifact identity mismatch", volume.Volume)
		}
	}

	attribution := ragSnapshotMustRead(t, ragSnapshotAttributionPath)
	if got := ragSnapshotSHA256(attribution); got != ragSnapshotAttributionSHA256 {
		t.Fatalf("snapshot attribution SHA-256 = %s, want %s", got, ragSnapshotAttributionSHA256)
	}
	for _, required := range []string{"三命通會 (四庫全書本)", "PD-old", "CC BY-SA 4.0 and GFDL", "snapshot-manifest.json", "not registered in the runtime RAG index", "not claim-eligible"} {
		if !bytes.Contains(attribution, []byte(required)) {
			t.Fatalf("snapshot attribution missing %q", required)
		}
	}
	if got := ragSnapshotSHA256(ragSnapshotMustRead(t, "../../../../scripts/fetch-wikisource-sanmintonghui-snapshot.sh")); got != ragSnapshotFetchScriptSHA256 {
		t.Fatalf("snapshot fetch script SHA-256 = %s, want %s", got, ragSnapshotFetchScriptSHA256)
	}
}

func TestRAGWikisourceMarkdownComparisonContract(t *testing.T) {
	raw := ragSnapshotMustRead(t, ragSnapshotComparisonPath)
	if got := ragSnapshotSHA256(raw); got != ragSnapshotComparisonSHA256 {
		t.Fatalf("snapshot comparison SHA-256 = %s, want %s", got, ragSnapshotComparisonSHA256)
	}
	report := ragSnapshotDecodeStrict[ragSnapshotComparison](t, raw)
	if report.Schema != "sanming_cross_source_text_comparison_v1" || report.Version != "2026-07-17.1" ||
		report.Status != "machine_comparison_not_independence_or_claim_adjudication" || report.ObservedAt != "2026-07-17" ||
		strings.TrimSpace(report.Purpose) == "" {
		t.Fatalf("unexpected comparison identity: %+v", report)
	}
	sources := report.Sources
	if sources.SnapshotManifestPath != "research/rag/snapshots/sanming-siku-wikisource-12vol-v1/snapshot-manifest.json" ||
		sources.SnapshotManifestSHA256 != ragSnapshotManifestSHA256 ||
		sources.AttributionPath != "research/rag/snapshots/sanming-siku-wikisource-12vol-v1/ATTRIBUTION.md" ||
		sources.AttributionSHA256 != ragSnapshotAttributionSHA256 || sources.CandidateID != "sanming-siku-wikisource-12vol-v1" ||
		sources.Provider != "zh.wikisource.org" || sources.RevisionVolumeCount != 12 ||
		sources.MarkdownRootLabel != "external_mingli_db/md/bazi/三命通会" ||
		sources.MarkdownManifestScheme != "sorted_filename_tab_sha256_lf_v1" ||
		sources.MarkdownManifestSHA256 != "ddef8e89beac4d106336d17e47899e8220b139c36117524b3ea71518dbb91bb7" ||
		sources.MarkdownFileCount != 382 || sources.NumberedChapterCount != 381 {
		t.Fatalf("comparison source identity invalid: %+v", sources)
	}
	ragSnapshotValidateNormalization(t, report.Normalization)
	ragSnapshotValidateSummary(t, report.Summary)
	ragSnapshotValidateComparisonVolumes(t, report.Volumes)
	ragSnapshotValidateChapters(t, report.Chapters, report.Summary)
	b := report.Boundaries
	if !b.MachineComparisonOnly || !b.TextualOverlapIsNotIdentity || !b.TextualOverlapIsNotIndependence ||
		b.BibliographyAdjudicated || b.IndependenceVerified || b.PageMappingVerified || b.ClaimSupportReviewed ||
		b.RuntimeIngestionAllowed || b.ClaimEligible || b.PublishableAccuracy {
		t.Fatalf("machine comparison boundary must fail closed: %+v", b)
	}
	if got := ragSnapshotSHA256(ragSnapshotMustRead(t, "../../../../scripts/audit-sanmintonghui-wikisource-markdown.go")); got != ragSnapshotAuditGeneratorSHA256 {
		t.Fatalf("comparison generator SHA-256 = %s, want %s", got, ragSnapshotAuditGeneratorSHA256)
	}
}

func TestRAGWikisourceSnapshotIsNotRuntimeRegistered(t *testing.T) {
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
		for _, forbidden := range []string{"sanming-siku-wikisource-12vol-v1", "sanming_cross_source_text_comparison_v1", "research/rag/snapshots"} {
			if bytes.Contains(raw, []byte(forbidden)) {
				t.Fatalf("research-only snapshot marker %q leaked into runtime ingestion source %s", forbidden, sourcePath)
			}
		}
	}
}

func TestRAGWikisourceSnapshotResearchDocumentsContract(t *testing.T) {
	marker := "第一百四十六项完成维基文库固定快照与跨来源文本差分治理"
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
			ragSnapshotManifestSHA256,
			ragSnapshotComparisonSHA256,
			"sanming_cross_source_text_comparison_v1",
			"381",
			"226",
			"卷十",
			"卷十一",
			"卷十二",
			"claim_eligible=false",
		} {
			if !strings.Contains(content, required) {
				t.Fatalf("%s is missing phase 146 evidence %q", path, required)
			}
		}
	}
}

func ragSnapshotValidateNormalization(t *testing.T, normalization ragSnapshotNormalization) {
	t.Helper()
	if normalization.WikitextExtraction != "balanced_template_render_v1_preserve_SK_anchor_SK_notes_YL_payload_drop_SKchar_and_layout_templates" ||
		normalization.MarkdownExtraction != "numbered_markdown_exact_section_from_原文_to_next_h2_or_eof_v1" ||
		normalization.ScriptConversion != "opencc_t2s_fixed_assets" ||
		normalization.CharacterFilter != "unicode_letters_and_numbers_without_compatibility_fold" ||
		normalization.AnchorSampling != "unique_evenly_spaced_fixed_width_v1" || normalization.AnchorWidthRunes != 20 ||
		normalization.MaximumAnchorsPerChapter != 48 || normalization.OpenCCVersion != "1.3.2.dirty" ||
		normalization.LocalCorpusSHA256 != "e2bb2a0498b0a765a8b635b2a484ff27bfd159d04252e18c53e9a8b9d2023915" ||
		normalization.RemoteCorpusSHA256 != "485261c008439e6fc9ff83e6ccb4e359e4794e27a9b69f89bc5ffa827bd7cf8c" {
		t.Fatalf("unexpected comparison normalization: %+v", normalization)
	}
	expectedAssets := map[string]string{
		"t2s.json":                          "96fe5cc374a80ccc49e3370006cce3aefe4af955868ae0b14fb3079ec695be4f",
		"CJK_Compatibility_Ideographs.ocd2": "4b1faa6649012f524068ec18c0fb520ead343c11cbe0a8e4c8853ca61369d666",
		"TSPhrases.ocd2":                    "e7f9d419d54f71a66d7f0283b29910913f08defdb6d4322e00c459c7ebe3f991",
		"TSCharactersExt.ocd2":              "2ee61f852d05a3241326ae8d7eeae00818a80c0a0f4e03704050312b9561bf33",
		"TSCharacters.ocd2":                 "014a1c9615f2a0800a56f0e6ce132c01ec233b089cd6160da66df9c346c0274b",
	}
	if len(normalization.OpenCCAssetSHA256) != len(expectedAssets) {
		t.Fatalf("OpenCC asset count = %d, want %d", len(normalization.OpenCCAssetSHA256), len(expectedAssets))
	}
	for name, want := range expectedAssets {
		if normalization.OpenCCAssetSHA256[name] != want {
			t.Fatalf("OpenCC asset %s SHA-256 = %s, want %s", name, normalization.OpenCCAssetSHA256[name], want)
		}
	}
}

func ragSnapshotValidateSummary(t *testing.T, summary ragSnapshotComparisonSummary) {
	t.Helper()
	wantCounts := []int{13, 26, 23, 7, 47, 45, 22, 60, 60, 41, 21, 16}
	if summary.ComparedChapters != 381 || summary.ZeroHitChapters != 0 || summary.ScoreBelow010 != 5 ||
		summary.Score010ToBelow025 != 35 || summary.Score025ToBelow050 != 206 || summary.Score050ToBelow080 != 125 ||
		summary.ScoreAtLeast080 != 10 || summary.TitleLocatedChapters != 226 || summary.AmbiguousBestVolume != 0 ||
		!summary.MachineOverlapObserved || len(summary.BestVolumeCounts) != 12 {
		t.Fatalf("unexpected comparison summary: %+v", summary)
	}
	for index, count := range summary.BestVolumeCounts {
		if count.Volume != index+1 || count.Count != wantCounts[index] {
			t.Fatalf("best volume count %d = %+v, want %d", index+1, count, wantCounts[index])
		}
	}
}

func ragSnapshotValidateComparisonVolumes(t *testing.T, volumes []ragSnapshotNormalizedVolume) {
	t.Helper()
	wantRunes := []int{36731, 36266, 34598, 35862, 36104, 36862, 32414, 40567, 39970, 33022, 42997, 27378}
	wantHashes := []string{
		"91ed904f472c2fd0b9089a0c2e040a0e451649395051ea860219046b80b2d523",
		"930b079a3e273d7d2f9d7e0b0e69c344e2f42cd3150347b65cc0a9c7721792b9",
		"df0dc184db3138477443a2551422f234053a7a61745dc94b08225bc3169641c3",
		"3ba5dde1ebd887bdb847dd7f67a96c14a1ed234eb84f9da805757b2f5d45c50d",
		"d6ad65f04652258ecb9fa59849cc875c23606d9faba0ea164cf383d83bc386be",
		"c0301ba9da9a91736d53ba539b7a9ac1a965af8499e37a2ba515d6bb4cace8bc",
		"b962652d532c204af55491bffc5ea2de09ac5efd89a822e5da073f5f5dd5c51f",
		"9e1ab1bb08d2cd1353fbb7d5e5c1b88d764beb6d2116ef4aa90a888648cc263f",
		"240fedf30951decd632ac990949fd6a0576c53950d7c264e13ea5d2da989cae6",
		"4dbb8d6b74c056b25d74729523951947a115a360ffaff82cae18e1679c1d8e1e",
		"470ec0871077aa6296446cbd5dcb3cc13dfb46c160f06b5e58765b9a54bfea9c",
		"565bfeb3d751358830aaf9926947fe772806ba2c7ce016ad782fde80682ae72a",
	}
	manifest := ragSnapshotDecodeStrict[ragSnapshotManifest](t, ragSnapshotMustRead(t, ragSnapshotManifestPath))
	if len(volumes) != 12 {
		t.Fatalf("normalized volume count = %d, want 12", len(volumes))
	}
	for index, volume := range volumes {
		source := manifest.Volumes[index]
		if volume.Volume != index+1 || volume.PageID != source.PageID || volume.RevisionID != source.RevisionID ||
			volume.SourceURL != source.SourceURL || volume.ArtifactPath != source.ArtifactPath || volume.LocalSHA256 != source.LocalSHA256 ||
			volume.NormalizedRunes != wantRunes[index] || volume.NormalizedSHA256 != wantHashes[index] {
			t.Fatalf("normalized volume %d identity mismatch: %+v", index+1, volume)
		}
	}
}

func ragSnapshotValidateChapters(t *testing.T, chapters []ragSnapshotChapterComparison, summary ragSnapshotComparisonSummary) {
	t.Helper()
	if len(chapters) != 381 {
		t.Fatalf("comparison chapter count = %d, want 381", len(chapters))
	}
	buckets := make([]int, 6)
	volumeCounts := make([]int, 12)
	titleLocated := 0
	for index, chapter := range chapters {
		wantChapter := index + 1
		if chapter.Chapter != wantChapter || chapter.File != fmt.Sprintf("%03d.md", wantChapter) || strings.TrimSpace(chapter.Title) == "" ||
			!ragExternalValidSHA256(chapter.OriginalSHA256) || chapter.NormalizedRunes < 20 || chapter.AnchorCount <= 0 ||
			chapter.AnchorCount > 48 || chapter.BestAnchorHits <= 0 || chapter.BestAnchorHits > chapter.AnchorCount ||
			chapter.ScoreBasisPoints != chapter.BestAnchorHits*10000/chapter.AnchorCount || len(chapter.BestCandidateVolumes) != 1 {
			t.Fatalf("invalid chapter comparison %d: %+v", wantChapter, chapter)
		}
		bestVolume := chapter.BestCandidateVolumes[0]
		if bestVolume < 1 || bestVolume > 12 {
			t.Fatalf("chapter %d invalid best volume %d", wantChapter, bestVolume)
		}
		volumeCounts[bestVolume-1]++
		if wantChapter <= 303 && bestVolume > 9 {
			t.Fatalf("chapter %d unexpectedly maps beyond volume 9", wantChapter)
		}
		if wantChapter >= 304 {
			wantVolume := 10
			if wantChapter >= 345 {
				wantVolume = 11
			}
			if wantChapter >= 366 {
				wantVolume = 12
			}
			if bestVolume != wantVolume {
				t.Fatalf("tail chapter %d best volume = %d, want %d", wantChapter, bestVolume, wantVolume)
			}
		}
		seenTitles := map[int]bool{}
		for _, volume := range chapter.TitleCandidateVolumes {
			if volume < 1 || volume > 12 || seenTitles[volume] {
				t.Fatalf("chapter %d invalid title candidate volumes: %v", wantChapter, chapter.TitleCandidateVolumes)
			}
			seenTitles[volume] = true
		}
		if len(chapter.TitleCandidateVolumes) > 0 {
			titleLocated++
		}
		switch {
		case chapter.BestAnchorHits == 0:
			buckets[0]++
		case chapter.ScoreBasisPoints < 1000:
			buckets[1]++
		case chapter.ScoreBasisPoints < 2500:
			buckets[2]++
		case chapter.ScoreBasisPoints < 5000:
			buckets[3]++
		case chapter.ScoreBasisPoints < 8000:
			buckets[4]++
		default:
			buckets[5]++
		}
	}
	wantBuckets := []int{summary.ZeroHitChapters, summary.ScoreBelow010, summary.Score010ToBelow025, summary.Score025ToBelow050, summary.Score050ToBelow080, summary.ScoreAtLeast080}
	if fmt.Sprint(buckets) != fmt.Sprint(wantBuckets) || titleLocated != summary.TitleLocatedChapters {
		t.Fatalf("recomputed comparison summary mismatch: buckets=%v titles=%d", buckets, titleLocated)
	}
	for index, count := range volumeCounts {
		if count != summary.BestVolumeCounts[index].Count {
			t.Fatalf("recomputed volume %d count = %d, want %d", index+1, count, summary.BestVolumeCounts[index].Count)
		}
	}
}

func ragSnapshotBoundaryFailsClosed(boundary ragSnapshotBoundaries) bool {
	return boundary.RawWikitextUnmodified && boundary.LocalArtifactFrozen && !boundary.IndependenceVerified &&
		!boundary.BibliographyAdjudicated && !boundary.PageMappingVerified && !boundary.ClaimSupportReviewed &&
		!boundary.RuntimeIngestionAllowed && !boundary.ClaimEligible && !boundary.PublishableAccuracy
}

func ragSnapshotDecodeStrict[T any](t *testing.T, raw []byte) T {
	t.Helper()
	var value T
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("strict JSON decode: %v", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		t.Fatalf("JSON artifact must contain exactly one document: %v", err)
	}
	return value
}

func ragSnapshotMustRead(t *testing.T, path string) []byte {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return raw
}

func ragSnapshotSHA256(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}
