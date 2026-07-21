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
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	expectedSnapshotManifestSHA256 = "38a03d0ee048a097620de3765f2793e9f0a5383a3238748fcd583ac79bb26974"
	expectedAttributionSHA256      = "84aad48ff04e5220b0dc999ffda2de4fd5e86157896a7f14b7f2f9511461c6b2"
	expectedMarkdownManifestSHA256 = "ddef8e89beac4d106336d17e47899e8220b139c36117524b3ea71518dbb91bb7"
	anchorWidthRunes               = 20
	maximumAnchorsPerChapter       = 48
)

var numberedMarkdown = regexp.MustCompile(`^[0-9]{3}\.md$`)

type snapshotManifest struct {
	Schema         string           `json:"schema"`
	Version        string           `json:"version"`
	Status         string           `json:"status"`
	RegistryPath   string           `json:"registry_path"`
	RegistrySHA256 string           `json:"registry_sha256"`
	CandidateID    string           `json:"candidate_id"`
	Provider       string           `json:"provider"`
	License        snapshotLicense  `json:"license"`
	Boundaries     snapshotBoundary `json:"boundaries"`
	Volumes        []snapshotVolume `json:"volumes"`
}

type snapshotLicense struct {
	UnderlyingWork       string `json:"underlying_work"`
	DigitalContributions string `json:"digital_contributions"`
	LicenseURL           string `json:"license_url"`
	AttributionFile      string `json:"attribution_file"`
}

type snapshotBoundary struct {
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

type snapshotVolume struct {
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

type auditReport struct {
	Schema        string              `json:"schema"`
	Version       string              `json:"version"`
	Status        string              `json:"status"`
	ObservedAt    string              `json:"observed_at"`
	Purpose       string              `json:"purpose"`
	Sources       auditSources        `json:"sources"`
	Normalization auditNormalization  `json:"normalization"`
	Summary       auditSummary        `json:"summary"`
	Volumes       []normalizedVolume  `json:"volumes"`
	Chapters      []chapterComparison `json:"chapters"`
	Boundaries    auditBoundaries     `json:"boundaries"`
}

type auditSources struct {
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

type auditNormalization struct {
	WikitextExtraction       string            `json:"wikitext_extraction"`
	MarkdownExtraction       string            `json:"markdown_extraction"`
	ScriptConversion         string            `json:"script_conversion"`
	CharacterFilter          string            `json:"character_filter"`
	AnchorSampling           string            `json:"anchor_sampling"`
	AnchorWidthRunes         int               `json:"anchor_width_runes"`
	MaximumAnchorsPerChapter int               `json:"maximum_anchors_per_chapter"`
	OpenCCVersion            string            `json:"opencc_version"`
	OpenCCAssets             map[string]string `json:"opencc_asset_sha256"`
	LocalCorpusSHA256        string            `json:"local_normalized_chapter_manifest_sha256"`
	RemoteCorpusSHA256       string            `json:"remote_normalized_volume_manifest_sha256"`
}

type auditSummary struct {
	ComparedChapters       int           `json:"compared_chapters"`
	ZeroHitChapters        int           `json:"zero_hit_chapters"`
	ScoreBelow010          int           `json:"score_below_0_10_chapters"`
	Score010ToBelow025     int           `json:"score_0_10_to_below_0_25_chapters"`
	Score025ToBelow050     int           `json:"score_0_25_to_below_0_50_chapters"`
	Score050ToBelow080     int           `json:"score_0_50_to_below_0_80_chapters"`
	ScoreAtLeast080        int           `json:"score_at_least_0_80_chapters"`
	TitleLocatedChapters   int           `json:"title_located_chapters"`
	AmbiguousBestVolume    int           `json:"ambiguous_best_volume_chapters"`
	BestVolumeCounts       []volumeCount `json:"best_volume_counts"`
	MachineOverlapObserved bool          `json:"machine_textual_overlap_observed"`
}

type volumeCount struct {
	Volume int `json:"volume"`
	Count  int `json:"count"`
}

type normalizedVolume struct {
	Volume           int    `json:"volume"`
	PageID           string `json:"page_id"`
	RevisionID       int    `json:"revision_id"`
	SourceURL        string `json:"source_url"`
	ArtifactPath     string `json:"artifact_path"`
	LocalSHA256      string `json:"local_sha256"`
	NormalizedRunes  int    `json:"normalized_runes"`
	NormalizedSHA256 string `json:"normalized_sha256"`
}

type chapterComparison struct {
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

type auditBoundaries struct {
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

type localChapter struct {
	comparison chapterComparison
	titleNorm  string
	textNorm   string
	anchors    []string
	hits       []int
	titles     []int
}

func main() {
	var snapshotRoot, markdownRoot, output, openCC, openCCConfig string
	flag.StringVar(&snapshotRoot, "snapshot", "", "path to the frozen Wikisource snapshot")
	flag.StringVar(&markdownRoot, "markdown-root", "", "path to the external 三命通会 Markdown root")
	flag.StringVar(&output, "output", "", "path for the JSON audit artifact")
	flag.StringVar(&openCC, "opencc", "/usr/sbin/opencc", "path to the OpenCC binary")
	flag.StringVar(&openCCConfig, "opencc-config", "/usr/share/opencc/t2s.json", "path to the fixed t2s config")
	flag.Parse()
	if snapshotRoot == "" || markdownRoot == "" || output == "" {
		fail(errors.New("-snapshot, -markdown-root, and -output are required"))
	}
	if err := run(snapshotRoot, markdownRoot, output, openCC, openCCConfig); err != nil {
		fail(err)
	}
}

func run(snapshotRoot, markdownRoot, output, openCC, openCCConfig string) error {
	manifestPath := filepath.Join(snapshotRoot, "snapshot-manifest.json")
	manifestRaw, manifest, err := readSnapshot(manifestPath)
	if err != nil {
		return err
	}
	if hashBytes(manifestRaw) != expectedSnapshotManifestSHA256 {
		return fmt.Errorf("snapshot manifest SHA-256 mismatch")
	}
	attributionPath := filepath.Join(snapshotRoot, manifest.License.AttributionFile)
	attributionRaw, err := os.ReadFile(attributionPath)
	if err != nil || hashBytes(attributionRaw) != expectedAttributionSHA256 {
		return fmt.Errorf("snapshot attribution identity mismatch: %w", err)
	}
	if err := validateSnapshot(snapshotRoot, manifest); err != nil {
		return err
	}

	chapters, markdownFiles, markdownManifest, localCorpusManifest, err := readLocalChapters(markdownRoot)
	if err != nil {
		return err
	}
	if hashBytes(markdownManifest) != expectedMarkdownManifestSHA256 {
		return fmt.Errorf("Markdown manifest SHA-256 mismatch")
	}
	openCCVersion, assets, err := validateOpenCC(openCC, openCCConfig)
	if err != nil {
		return err
	}

	volumes := make([]normalizedVolume, 0, len(manifest.Volumes))
	remoteCorpusManifest := bytes.Buffer{}
	for volumeIndex, volume := range manifest.Volumes {
		raw, err := os.ReadFile(filepath.Join(snapshotRoot, filepath.FromSlash(volume.ArtifactPath)))
		if err != nil {
			return err
		}
		rendered := renderWikitext(string(raw))
		simplified, err := convertOpenCC(openCC, openCCConfig, rendered)
		if err != nil {
			return fmt.Errorf("OpenCC volume %d: %w", volume.Volume, err)
		}
		normalized := normalizeText(simplified)
		normalizedHash := hashBytes([]byte(normalized))
		fmt.Fprintf(&remoteCorpusManifest, "volume-%02d\t%s\n", volume.Volume, normalizedHash)
		volumes = append(volumes, normalizedVolume{
			Volume: volume.Volume, PageID: volume.PageID, RevisionID: volume.RevisionID,
			SourceURL: volume.SourceURL, ArtifactPath: volume.ArtifactPath, LocalSHA256: volume.LocalSHA256,
			NormalizedRunes: utf8.RuneCountInString(normalized), NormalizedSHA256: normalizedHash,
		})
		index := runeNGramIndex(normalized, anchorWidthRunes)
		for _, chapter := range chapters {
			hits := 0
			for _, anchor := range chapter.anchors {
				if _, ok := index[anchor]; ok {
					hits++
				}
			}
			chapter.hits[volumeIndex] = hits
			if chapter.titleNorm != "" && strings.Contains(normalized, chapter.titleNorm) {
				chapter.titles = append(chapter.titles, volume.Volume)
			}
		}
	}

	comparisons, summary := finalizeComparisons(chapters, len(manifest.Volumes))
	report := auditReport{
		Schema: "sanming_cross_source_text_comparison_v1", Version: "2026-07-17.1",
		Status: "machine_comparison_not_independence_or_claim_adjudication", ObservedAt: "2026-07-17",
		Purpose: "Freeze a reproducible, chapter-level textual-overlap audit between the complete Wikisource transcription and the external Markdown corpus without promoting either source to runtime evidence.",
		Sources: auditSources{
			SnapshotManifestPath:   "research/rag/snapshots/sanming-siku-wikisource-12vol-v1/snapshot-manifest.json",
			SnapshotManifestSHA256: expectedSnapshotManifestSHA256,
			AttributionPath:        "research/rag/snapshots/sanming-siku-wikisource-12vol-v1/ATTRIBUTION.md",
			AttributionSHA256:      expectedAttributionSHA256, CandidateID: manifest.CandidateID, Provider: manifest.Provider,
			RevisionVolumeCount: len(manifest.Volumes), MarkdownRootLabel: "external_mingli_db/md/bazi/三命通会",
			MarkdownManifestScheme: "sorted_filename_tab_sha256_lf_v1", MarkdownManifestSHA256: expectedMarkdownManifestSHA256,
			MarkdownFileCount: markdownFiles, NumberedChapterCount: len(chapters),
		},
		Normalization: auditNormalization{
			WikitextExtraction: "balanced_template_render_v1_preserve_SK_anchor_SK_notes_YL_payload_drop_SKchar_and_layout_templates",
			MarkdownExtraction: "numbered_markdown_exact_section_from_原文_to_next_h2_or_eof_v1",
			ScriptConversion:   "opencc_t2s_fixed_assets", CharacterFilter: "unicode_letters_and_numbers_without_compatibility_fold",
			AnchorSampling: "unique_evenly_spaced_fixed_width_v1", AnchorWidthRunes: anchorWidthRunes,
			MaximumAnchorsPerChapter: maximumAnchorsPerChapter, OpenCCVersion: openCCVersion, OpenCCAssets: assets,
			LocalCorpusSHA256: hashBytes(localCorpusManifest), RemoteCorpusSHA256: hashBytes(remoteCorpusManifest.Bytes()),
		},
		Summary: summary, Volumes: volumes, Chapters: comparisons,
		Boundaries: auditBoundaries{
			MachineComparisonOnly: true, TextualOverlapIsNotIdentity: true, TextualOverlapIsNotIndependence: true,
			BibliographyAdjudicated: false, IndependenceVerified: false, PageMappingVerified: false,
			ClaimSupportReviewed: false, RuntimeIngestionAllowed: false, ClaimEligible: false, PublishableAccuracy: false,
		},
	}
	return writeJSONAtomic(output, report)
}

func readSnapshot(path string) ([]byte, snapshotManifest, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, snapshotManifest{}, err
	}
	var manifest snapshotManifest
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&manifest); err != nil {
		return nil, manifest, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, manifest, errors.New("snapshot manifest must contain one JSON document")
	}
	return raw, manifest, nil
}

func validateSnapshot(root string, manifest snapshotManifest) error {
	if manifest.Schema != "wikisource_revision_snapshot_v1" || manifest.Version != "2026-07-17.1" ||
		manifest.Status != "research_snapshot_not_runtime_eligible" || manifest.CandidateID != "sanming-siku-wikisource-12vol-v1" ||
		manifest.Provider != "zh.wikisource.org" || len(manifest.Volumes) != 12 {
		return errors.New("unexpected snapshot identity")
	}
	if manifest.License.UnderlyingWork != "PD-old" || manifest.License.DigitalContributions != "CC BY-SA 4.0 and GFDL" ||
		manifest.License.LicenseURL == "" || manifest.License.AttributionFile != "ATTRIBUTION.md" {
		return errors.New("snapshot license metadata incomplete")
	}
	b := manifest.Boundaries
	if !b.RawWikitextUnmodified || !b.LocalArtifactFrozen || b.IndependenceVerified || b.BibliographyAdjudicated ||
		b.PageMappingVerified || b.ClaimSupportReviewed || b.RuntimeIngestionAllowed || b.ClaimEligible || b.PublishableAccuracy {
		return errors.New("snapshot boundary is not fail-closed")
	}
	seenURL := map[string]bool{}
	for index, volume := range manifest.Volumes {
		if volume.Volume != index+1 || volume.PageID == "" || volume.RevisionID <= 0 || volume.RemoteSize <= 0 ||
			volume.LocalSize != volume.RemoteSize || !validHex(volume.RemoteSHA1, 40) || !validHex(volume.LocalSHA256, 64) ||
			seenURL[volume.SourceURL] || !strings.HasPrefix(volume.SourceURL, "https://zh.wikisource.org/w/index.php?curid=") {
			return fmt.Errorf("invalid snapshot volume %d", index+1)
		}
		seenURL[volume.SourceURL] = true
		raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(volume.ArtifactPath)))
		if err != nil || int64(len(raw)) != volume.LocalSize || hashBytes(raw) != volume.LocalSHA256 {
			return fmt.Errorf("snapshot volume %d local identity mismatch: %w", volume.Volume, err)
		}
	}
	return nil
}

func readLocalChapters(root string) ([]*localChapter, int, []byte, []byte, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	manifest := bytes.Buffer{}
	localCorpus := bytes.Buffer{}
	chapters := make([]*localChapter, 0, 381)
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return nil, 0, nil, nil, err
		}
		fmt.Fprintf(&manifest, "%s\t%s\n", name, hashBytes(raw))
		if !numberedMarkdown.MatchString(name) {
			continue
		}
		number, _ := strconv.Atoi(strings.TrimSuffix(name, ".md"))
		if number == 0 {
			continue
		}
		title, original, err := extractMarkdownOriginal(string(raw))
		if err != nil {
			return nil, 0, nil, nil, fmt.Errorf("%s: %w", name, err)
		}
		normalized := normalizeText(original)
		anchors := sampleAnchors(normalized, anchorWidthRunes, maximumAnchorsPerChapter)
		if len(anchors) == 0 {
			return nil, 0, nil, nil, fmt.Errorf("%s has no comparable original text", name)
		}
		normalizedHash := hashBytes([]byte(normalized))
		fmt.Fprintf(&localCorpus, "%s\t%s\n", name, normalizedHash)
		chapters = append(chapters, &localChapter{
			comparison: chapterComparison{Chapter: number, File: name, Title: title, OriginalSHA256: hashBytes([]byte(original)), NormalizedRunes: utf8.RuneCountInString(normalized), AnchorCount: len(anchors)},
			titleNorm:  normalizeText(title), textNorm: normalized, anchors: anchors, hits: make([]int, 12),
		})
	}
	if len(names) != 382 || len(chapters) != 381 {
		return nil, 0, nil, nil, fmt.Errorf("Markdown inputs = %d files/%d chapters, want 382/381", len(names), len(chapters))
	}
	for index, chapter := range chapters {
		if chapter.comparison.Chapter != index+1 {
			return nil, 0, nil, nil, errors.New("numbered Markdown chapters must be contiguous")
		}
	}
	return chapters, len(names), manifest.Bytes(), localCorpus.Bytes(), nil
}

func extractMarkdownOriginal(raw string) (string, string, error) {
	firstEnd := strings.IndexByte(raw, '\n')
	if firstEnd < 0 || !strings.HasPrefix(raw[:firstEnd], "# ") {
		return "", "", errors.New("missing H1 title")
	}
	title := strings.TrimSpace(strings.TrimPrefix(raw[:firstEnd], "# "))
	marker := "\n## 原文\n"
	start := strings.Index(raw, marker)
	if start < 0 {
		return "", "", errors.New("missing 原文 section")
	}
	start += len(marker)
	remainder := raw[start:]
	end := strings.Index(remainder, "\n## ")
	if end < 0 {
		end = len(remainder)
	}
	return title, strings.TrimSpace(remainder[:end]), nil
}

func validateOpenCC(binary, config string) (string, map[string]string, error) {
	output, err := exec.Command(binary, "--version").CombinedOutput()
	if err != nil {
		return "", nil, err
	}
	version := ""
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Version:") {
			version = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "Version:"))
		}
	}
	if version != "1.3.2.dirty" {
		return "", nil, fmt.Errorf("OpenCC version = %q, want 1.3.2.dirty", version)
	}
	dir := filepath.Dir(config)
	expected := map[string]string{
		"t2s.json":                          "96fe5cc374a80ccc49e3370006cce3aefe4af955868ae0b14fb3079ec695be4f",
		"CJK_Compatibility_Ideographs.ocd2": "4b1faa6649012f524068ec18c0fb520ead343c11cbe0a8e4c8853ca61369d666",
		"TSPhrases.ocd2":                    "e7f9d419d54f71a66d7f0283b29910913f08defdb6d4322e00c459c7ebe3f991",
		"TSCharactersExt.ocd2":              "2ee61f852d05a3241326ae8d7eeae00818a80c0a0f4e03704050312b9561bf33",
		"TSCharacters.ocd2":                 "014a1c9615f2a0800a56f0e6ce132c01ec233b089cd6160da66df9c346c0274b",
	}
	for name, want := range expected {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil || hashBytes(raw) != want {
			return "", nil, fmt.Errorf("OpenCC asset %s identity mismatch: %w", name, err)
		}
	}
	return version, expected, nil
}

func convertOpenCC(binary, config, input string) (string, error) {
	command := exec.Command(binary, "-c", config)
	command.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	if err := command.Run(); err != nil {
		return "", fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return stdout.String(), nil
}

func renderWikitext(raw string) string {
	raw = removeDelimited(raw, "<!--", "-->")
	raw = renderTemplates(raw)
	raw = removeDelimited(raw, "<", ">")
	raw = renderWikiLinks(raw)
	return raw
}

func renderTemplates(raw string) string {
	var out strings.Builder
	for {
		start := strings.Index(raw, "{{")
		if start < 0 {
			out.WriteString(raw)
			return out.String()
		}
		out.WriteString(raw[:start])
		end := matchingTemplateEnd(raw, start)
		if end < 0 {
			out.WriteString(raw[start:])
			return out.String()
		}
		inner := raw[start+2 : end-2]
		name, payload := splitTemplate(inner)
		switch strings.TrimSpace(name) {
		case "SK anchor", "SK notes", "YL":
			out.WriteString(renderTemplates(payload))
		}
		raw = raw[end:]
	}
}

func matchingTemplateEnd(raw string, start int) int {
	depth := 0
	for index := start; index+1 < len(raw); index++ {
		switch raw[index : index+2] {
		case "{{":
			depth++
			index++
		case "}}":
			depth--
			index++
			if depth == 0 {
				return index + 1
			}
		}
	}
	return -1
}

func splitTemplate(inner string) (string, string) {
	depth := 0
	for index := 0; index < len(inner); index++ {
		if index+1 < len(inner) && inner[index:index+2] == "{{" {
			depth++
			index++
			continue
		}
		if index+1 < len(inner) && inner[index:index+2] == "}}" {
			depth--
			index++
			continue
		}
		if inner[index] == '|' && depth == 0 {
			return inner[:index], inner[index+1:]
		}
	}
	return inner, ""
}

func removeDelimited(raw, opening, closing string) string {
	var out strings.Builder
	for {
		start := strings.Index(raw, opening)
		if start < 0 {
			out.WriteString(raw)
			return out.String()
		}
		out.WriteString(raw[:start])
		end := strings.Index(raw[start+len(opening):], closing)
		if end < 0 {
			return out.String()
		}
		raw = raw[start+len(opening)+end+len(closing):]
	}
}

func renderWikiLinks(raw string) string {
	for {
		start := strings.Index(raw, "[[")
		if start < 0 {
			return raw
		}
		end := strings.Index(raw[start+2:], "]]")
		if end < 0 {
			return raw
		}
		end += start + 2
		inner := raw[start+2 : end]
		if pipe := strings.LastIndexByte(inner, '|'); pipe >= 0 {
			inner = inner[pipe+1:]
		}
		raw = raw[:start] + inner + raw[end+2:]
	}
}

func normalizeText(raw string) string {
	var out strings.Builder
	for _, r := range raw {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			out.WriteRune(r)
		}
	}
	return out.String()
}

func sampleAnchors(normalized string, width, maximum int) []string {
	runes := []rune(normalized)
	if len(runes) < width {
		return nil
	}
	positions := len(runes) - width + 1
	count := maximum
	if positions < count {
		count = positions
	}
	seen := map[string]bool{}
	anchors := make([]string, 0, count)
	for index := 0; index < count; index++ {
		position := 0
		if count > 1 {
			position = index * (positions - 1) / (count - 1)
		}
		anchor := string(runes[position : position+width])
		if !seen[anchor] {
			seen[anchor] = true
			anchors = append(anchors, anchor)
		}
	}
	return anchors
}

func runeNGramIndex(normalized string, width int) map[string]struct{} {
	runes := []rune(normalized)
	index := make(map[string]struct{}, len(runes))
	for position := 0; position+width <= len(runes); position++ {
		index[string(runes[position:position+width])] = struct{}{}
	}
	return index
}

func finalizeComparisons(chapters []*localChapter, volumeTotal int) ([]chapterComparison, auditSummary) {
	summary := auditSummary{ComparedChapters: len(chapters), BestVolumeCounts: make([]volumeCount, volumeTotal)}
	for index := range summary.BestVolumeCounts {
		summary.BestVolumeCounts[index].Volume = index + 1
	}
	comparisons := make([]chapterComparison, 0, len(chapters))
	for _, chapter := range chapters {
		best := 0
		for _, hits := range chapter.hits {
			if hits > best {
				best = hits
			}
		}
		for index, hits := range chapter.hits {
			if best > 0 && hits == best {
				chapter.comparison.BestCandidateVolumes = append(chapter.comparison.BestCandidateVolumes, index+1)
			}
		}
		chapter.comparison.BestAnchorHits = best
		chapter.comparison.ScoreBasisPoints = best * 10000 / len(chapter.anchors)
		chapter.comparison.TitleCandidateVolumes = chapter.titles
		if chapter.comparison.BestCandidateVolumes == nil {
			chapter.comparison.BestCandidateVolumes = []int{}
		}
		if chapter.comparison.TitleCandidateVolumes == nil {
			chapter.comparison.TitleCandidateVolumes = []int{}
		}
		if len(chapter.titles) > 0 {
			summary.TitleLocatedChapters++
		}
		if len(chapter.comparison.BestCandidateVolumes) > 1 {
			summary.AmbiguousBestVolume++
		}
		if len(chapter.comparison.BestCandidateVolumes) == 1 {
			summary.BestVolumeCounts[chapter.comparison.BestCandidateVolumes[0]-1].Count++
		}
		score := chapter.comparison.ScoreBasisPoints
		switch {
		case best == 0:
			summary.ZeroHitChapters++
		case score < 1000:
			summary.ScoreBelow010++
		case score < 2500:
			summary.Score010ToBelow025++
		case score < 5000:
			summary.Score025ToBelow050++
		case score < 8000:
			summary.Score050ToBelow080++
		default:
			summary.ScoreAtLeast080++
		}
		comparisons = append(comparisons, chapter.comparison)
	}
	summary.MachineOverlapObserved = summary.ScoreAtLeast080 > 0
	return comparisons, summary
}

func writeJSONAtomic(path string, value any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(filepath.Dir(path), ".sanming-comparison-*.json")
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

func validHex(value string, length int) bool {
	if len(value) != length {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil && value == strings.ToLower(value)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
