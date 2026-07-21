package main

import (
	"bufio"
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
)

const (
	expectedOCRManifestSHA256      = "e072f6a467ea83f55f63687b1b936756071dbc8a87862937be4f2478c87dce27"
	expectedOCRArtifactSHA256      = "c7fd70b653015a99949dd9926c07660efa6744560a0dd44c4c9a54d8f6d8c164"
	expectedVolumeLabelsSHA256     = "8dfe0d08345b2ded9b58b394e44576e72a8609d46208ac8d4908c05bb873ca31"
	expectedVolumeMappingSHA256    = "6a8fb60ee791c5b01dddbc5068cb1a58f1694626c686d5a82fd4ce2af25da1a5"
	expectedVolumeComparisonSHA256 = "ce7ee55bd4c6b68c488f3326b9b3725860b5ebea0c679ee12d156f330a180b7f"
	expectedMarkdownManifestSHA256 = "ddef8e89beac4d106336d17e47899e8220b139c36117524b3ea71518dbb91bb7"
	contentNGramWidth              = 4
	maximumContentTopCandidates    = 5
)

var numberedMarkdown = regexp.MustCompile(`^[0-9]{3}\.md$`)

type artifactReference struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

type ocrManifest struct {
	Schema      string `json:"schema"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	CandidateID string `json:"candidate_id"`
	OCR         struct {
		Package             string  `json:"package"`
		PackageVersion      string  `json:"package_version"`
		ONNXRuntimeVersion  string  `json:"onnxruntime_version"`
		OpenCVPythonVersion string  `json:"opencv_python_version"`
		UseClassification   bool    `json:"use_classification"`
		WorkerProcesses     int     `json:"worker_processes"`
		IntraOpThreads      int     `json:"intra_op_threads_per_session"`
		TextScoreThreshold  float64 `json:"text_score_threshold"`
		ConfigSHA256        string  `json:"config_sha256"`
		Models              struct {
			Detection            modelIdentity `json:"detection"`
			Recognition          modelIdentity `json:"recognition"`
			ClassificationUnused modelIdentity `json:"classification_unused"`
		} `json:"models"`
	} `json:"ocr"`
	Artifact   artifactReference `json:"artifact"`
	Summary    ocrSummary        `json:"summary"`
	Boundaries struct {
		CompletePageSetObserved    bool `json:"complete_page_set_observed"`
		MachineOCROnly             bool `json:"machine_ocr_only"`
		OCRTextNotDiplomatic       bool `json:"ocr_text_is_not_diplomatic_transcription"`
		IndependentReviewComplete  bool `json:"independent_review_complete"`
		VolumeMappingVerified      bool `json:"volume_mapping_verified"`
		ChapterPageMappingVerified bool `json:"chapter_page_mapping_verified"`
		RuntimeIngestionAllowed    bool `json:"runtime_ingestion_allowed"`
		ClaimEligible              bool `json:"claim_eligible"`
		PublishableAccuracy        bool `json:"publishable_accuracy"`
	} `json:"boundaries"`
}

type modelIdentity struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

type ocrSummary struct {
	PageCount           int   `json:"page_count"`
	Part1Pages          int   `json:"part_1_pages"`
	Part2Pages          int   `json:"part_2_pages"`
	EmptyPages          int   `json:"empty_pages"`
	RecognizedLines     int   `json:"recognized_lines"`
	ScoreBasisPointsSum int64 `json:"score_basis_points_sum"`
	MinimumLineScore    int   `json:"minimum_line_score_basis_points"`
	MaximumLineScore    int   `json:"maximum_line_score_basis_points"`
}

type pageKey struct {
	Part int `json:"part"`
	Page int `json:"physical_page"`
}

type ocrRecord struct {
	Part         int `json:"part"`
	PhysicalPage int `json:"physical_page"`
	Render       struct {
		WidthPixels  int    `json:"width_pixels"`
		HeightPixels int    `json:"height_pixels"`
		SHA256       string `json:"sha256"`
	} `json:"render"`
	Lines []struct {
		Box              [][]int `json:"box"`
		Text             string  `json:"text"`
		ScoreBasisPoints int     `json:"score_basis_points"`
	} `json:"lines"`
}

type bookSegment struct {
	Part      int `json:"part"`
	FirstPage int `json:"first_page"`
	LastPage  int `json:"last_page"`
}

type volumeLabelObservation struct {
	Schema      string `json:"schema"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	CandidateID string `json:"candidate_id"`
	Method      struct {
		MappingManifestSHA256 string `json:"mapping_manifest_sha256"`
	} `json:"method"`
	Observations []struct {
		BookCandidate   int `json:"book_candidate"`
		VolumeCandidate int `json:"volume_candidate"`
		Source          struct {
			Part                 int           `json:"part"`
			PhysicalPage         int           `json:"physical_page"`
			PhysicalBookSegments []bookSegment `json:"physical_book_segments"`
		} `json:"source"`
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

type volumeComparison struct {
	Schema  string `json:"schema"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Sources struct {
		MarkdownManifestSHA256 string `json:"markdown_manifest_sha256"`
		MarkdownFileCount      int    `json:"markdown_file_count"`
		NumberedChapterCount   int    `json:"numbered_chapter_count"`
	} `json:"sources"`
	Chapters []struct {
		Chapter              int    `json:"chapter"`
		File                 string `json:"file"`
		Title                string `json:"title"`
		OriginalSHA256       string `json:"original_sha256"`
		BestCandidateVolumes []int  `json:"best_candidate_volumes"`
	} `json:"chapters"`
	Boundaries struct {
		PageMappingVerified     bool `json:"page_mapping_verified"`
		RuntimeIngestionAllowed bool `json:"runtime_ingestion_allowed"`
		ClaimEligible           bool `json:"claim_eligible"`
		PublishableAccuracy     bool `json:"publishable_accuracy"`
	} `json:"boundaries"`
}

type localChapter struct {
	Chapter      int
	File         string
	Title        string
	Original     string
	TitleNorm    string
	OriginalNorm string
}

type pageIndex struct {
	Key   pageKey
	Lines []string
	Grams map[string]struct{}
}

type scoredPage struct {
	Part         int `json:"part"`
	PhysicalPage int `json:"physical_page"`
	OverlapCount int `json:"overlap_count"`
}

type chapterCandidate struct {
	Chapter                    int          `json:"chapter"`
	File                       string       `json:"file"`
	Title                      string       `json:"title"`
	OriginalSHA256             string       `json:"original_sha256"`
	VolumeCandidate            int          `json:"volume_candidate"`
	SearchPageCount            int          `json:"search_page_count"`
	TitleLocatorCandidates     []pageKey    `json:"title_locator_candidates"`
	TitleLocatorCandidateCount int          `json:"title_locator_candidate_count"`
	ContentNGramCount          int          `json:"content_ngram_count"`
	BestContentOverlap         int          `json:"best_content_overlap"`
	SecondDistinctOverlap      int          `json:"second_distinct_content_overlap"`
	ContentOverlapMargin       int          `json:"content_overlap_margin"`
	BestContentCandidateCount  int          `json:"best_content_candidate_count"`
	BestContentCandidates      []pageKey    `json:"best_content_candidates"`
	ContentTopCandidates       []scoredPage `json:"content_top_candidates"`
	ZeroContentOverlap         bool         `json:"zero_content_overlap"`
}

type reportSummary struct {
	ChapterCount                int `json:"chapter_count"`
	ExactTitleLocatedChapters   int `json:"exact_title_located_chapters"`
	ZeroContentOverlapChapters  int `json:"zero_content_overlap_chapters"`
	BestContentPageTieChapters  int `json:"best_content_page_tie_chapters"`
	ZeroMarginChapters          int `json:"zero_margin_chapters"`
	MarginAtMost2Chapters       int `json:"margin_at_most_2_chapters"`
	BestContentOverlapAtLeast1  int `json:"best_content_overlap_at_least_1_chapters"`
	BestContentOverlapAtLeast3  int `json:"best_content_overlap_at_least_3_chapters"`
	BestContentOverlapAtLeast5  int `json:"best_content_overlap_at_least_5_chapters"`
	BestContentOverlapAtLeast10 int `json:"best_content_overlap_at_least_10_chapters"`
	BestContentOverlapAtLeast20 int `json:"best_content_overlap_at_least_20_chapters"`
	BestContentOverlapAtLeast40 int `json:"best_content_overlap_at_least_40_chapters"`
}

type auditReport struct {
	Schema      string `json:"schema"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	ObservedAt  string `json:"observed_at"`
	CandidateID string `json:"candidate_id"`
	Purpose     string `json:"purpose"`
	Sources     struct {
		OCRManifest            artifactReference `json:"ocr_manifest"`
		OCRArtifact            artifactReference `json:"ocr_artifact"`
		VolumeLabelObservation artifactReference `json:"volume_label_observation"`
		VolumeComparison       artifactReference `json:"volume_comparison"`
		MarkdownRootLabel      string            `json:"markdown_root_label"`
		MarkdownManifestScheme string            `json:"markdown_manifest_scheme"`
		MarkdownManifestSHA256 string            `json:"markdown_manifest_sha256"`
		MarkdownFileCount      int               `json:"markdown_file_count"`
		NumberedChapterCount   int               `json:"numbered_chapter_count"`
	} `json:"sources"`
	Method struct {
		ScriptConversion            string            `json:"script_conversion"`
		OpenCCVersion               string            `json:"opencc_version"`
		OpenCCAssets                map[string]string `json:"opencc_asset_sha256"`
		CharacterFilter             string            `json:"character_filter"`
		OCRConversionUnit           string            `json:"ocr_conversion_unit"`
		ContentNGramWidth           int               `json:"content_ngram_width"`
		ContentNGramRule            string            `json:"content_ngram_rule"`
		VolumeRestriction           string            `json:"volume_restriction"`
		SecondScoreRule             string            `json:"second_score_rule"`
		CandidateOrdering           string            `json:"candidate_ordering"`
		MaximumContentTopCandidates int               `json:"maximum_content_top_candidates"`
	} `json:"method"`
	OCRSummary ocrSummary         `json:"ocr_summary"`
	Summary    reportSummary      `json:"summary"`
	Chapters   []chapterCandidate `json:"chapters"`
	Boundaries struct {
		MachineCandidatesOnly                 bool `json:"machine_candidates_only"`
		TitleLocatorCandidateNotStartPageGold bool `json:"title_locator_candidate_not_start_page_gold"`
		ContentOverlapNotSupportingCitation   bool `json:"content_overlap_candidate_not_supporting_citation"`
		OCRTextNotDiplomatic                  bool `json:"ocr_text_is_not_diplomatic_transcription"`
		IndependentReviewComplete             bool `json:"independent_review_complete"`
		VolumeMappingVerified                 bool `json:"volume_mapping_verified"`
		ChapterPageMappingVerified            bool `json:"chapter_page_mapping_verified"`
		ClaimSupportReviewed                  bool `json:"claim_support_reviewed"`
		RuntimeIngestionAllowed               bool `json:"runtime_ingestion_allowed"`
		ClaimEligible                         bool `json:"claim_eligible"`
		PublishableAccuracy                   bool `json:"publishable_accuracy"`
	} `json:"boundaries"`
}

func main() {
	var ocrRoot, volumeLabelsPath, comparisonPath, markdownRoot, output, openCC, openCCConfig string
	flag.StringVar(&ocrRoot, "ocr-root", "", "path to the frozen NCL OCR snapshot")
	flag.StringVar(&volumeLabelsPath, "volume-labels", "", "path to the frozen volume-label observation")
	flag.StringVar(&comparisonPath, "volume-comparison", "", "path to the frozen chapter-to-volume comparison")
	flag.StringVar(&markdownRoot, "markdown-root", "", "path to the external 三命通会 Markdown root")
	flag.StringVar(&output, "output", "", "path for the JSON audit artifact")
	flag.StringVar(&openCC, "opencc", "/usr/sbin/opencc", "path to the OpenCC binary")
	flag.StringVar(&openCCConfig, "opencc-config", "/usr/share/opencc/t2s.json", "path to the fixed t2s config")
	flag.Parse()
	if ocrRoot == "" || volumeLabelsPath == "" || comparisonPath == "" || markdownRoot == "" || output == "" {
		fail(errors.New("-ocr-root, -volume-labels, -volume-comparison, -markdown-root, and -output are required"))
	}
	if err := run(ocrRoot, volumeLabelsPath, comparisonPath, markdownRoot, output, openCC, openCCConfig); err != nil {
		fail(err)
	}
}

func run(ocrRoot, volumeLabelsPath, comparisonPath, markdownRoot, output, openCC, openCCConfig string) error {
	manifestPath := filepath.Join(ocrRoot, "ocr-manifest.json")
	manifestRaw, manifest, err := readStrictJSON[ocrManifest](manifestPath)
	if err != nil {
		return err
	}
	if hashBytes(manifestRaw) != expectedOCRManifestSHA256 {
		return errors.New("OCR manifest SHA-256 mismatch")
	}
	if err := validateOCRManifest(manifest); err != nil {
		return err
	}
	artifactPath := filepath.Join(ocrRoot, filepath.FromSlash(manifest.Artifact.Path))
	artifactRaw, err := os.ReadFile(artifactPath)
	if err != nil {
		return err
	}
	if int64(len(artifactRaw)) != manifest.Artifact.SizeBytes || hashBytes(artifactRaw) != expectedOCRArtifactSHA256 || manifest.Artifact.SHA256 != expectedOCRArtifactSHA256 {
		return errors.New("OCR artifact identity mismatch")
	}
	pages, observedOCRSummary, err := readOCRRecords(artifactRaw)
	if err != nil {
		return err
	}
	if observedOCRSummary != manifest.Summary {
		return fmt.Errorf("OCR artifact summary mismatch: got %+v want %+v", observedOCRSummary, manifest.Summary)
	}

	volumeRaw, volumeLabels, err := readStrictJSON[volumeLabelObservation](volumeLabelsPath)
	if err != nil {
		return err
	}
	if hashBytes(volumeRaw) != expectedVolumeLabelsSHA256 {
		return errors.New("volume-label observation SHA-256 mismatch")
	}
	if err := validateVolumeLabels(volumeLabels, pages); err != nil {
		return err
	}

	comparisonRaw, comparison, err := readStrictJSON[volumeComparison](comparisonPath)
	if err != nil {
		return err
	}
	if hashBytes(comparisonRaw) != expectedVolumeComparisonSHA256 {
		return errors.New("volume comparison SHA-256 mismatch")
	}
	if err := validateVolumeComparison(comparison); err != nil {
		return err
	}

	chapters, markdownFileCount, markdownManifest, err := readLocalChapters(markdownRoot, comparison)
	if err != nil {
		return err
	}
	if hashBytes(markdownManifest) != expectedMarkdownManifestSHA256 {
		return errors.New("Markdown manifest SHA-256 mismatch")
	}
	openCCVersion, openCCAssets, err := validateOpenCC(openCC, openCCConfig)
	if err != nil {
		return err
	}
	if err := normalizeInputsWithOpenCC(openCC, openCCConfig, chapters, pages); err != nil {
		return err
	}

	pageIndexes := buildPageIndexes(pages)
	chapterCandidates, summary, err := mapChapters(chapters, comparison, volumeLabels, pageIndexes)
	if err != nil {
		return err
	}
	report := buildReport(manifest, markdownFileCount, openCCVersion, openCCAssets, observedOCRSummary, summary, chapterCandidates)
	return writeJSONAtomic(output, report)
}

func validateOCRManifest(manifest ocrManifest) error {
	if manifest.Schema != "sanming_ncl_page_ocr_snapshot_v1" || manifest.Version != "2026-07-17.1" ||
		manifest.Status != "machine_ocr_silver_not_page_mapping_gold" || manifest.CandidateID != "sanming-ncl-06589-1578-12vol-scan-v1" ||
		manifest.Artifact.Path != "page-ocr.jsonl" || manifest.Artifact.SizeBytes != 3232151 ||
		manifest.OCR.Package != "rapidocr-onnxruntime" || manifest.OCR.PackageVersion != "1.4.4" ||
		manifest.OCR.ONNXRuntimeVersion != "1.27.0" || manifest.OCR.OpenCVPythonVersion != "5.0.0.93" ||
		manifest.OCR.UseClassification || manifest.OCR.WorkerProcesses != 4 || manifest.OCR.IntraOpThreads != 2 ||
		manifest.OCR.TextScoreThreshold != 0.5 || manifest.OCR.ConfigSHA256 != "bf94a1da4cba828e67b1d61e27cee14d9e7da27c9f272e04048a17e41ae97332" ||
		manifest.OCR.Models.Detection.SHA256 != "d2a7720d45a54257208b1e13e36a8479894cb74155a5efe29462512d42f49da9" ||
		manifest.OCR.Models.Recognition.SHA256 != "48fc40f24f6d2a207a2b1091d3437eb3cc3eb6b676dc3ef9c37384005483683b" ||
		manifest.OCR.Models.ClassificationUnused.SHA256 != "e47acedf663230f8863ff1ab0e64dd2d82b838fceb5957146dab185a89d6215c" {
		return errors.New("unexpected OCR manifest identity")
	}
	b := manifest.Boundaries
	if !b.CompletePageSetObserved || !b.MachineOCROnly || !b.OCRTextNotDiplomatic || b.IndependentReviewComplete ||
		b.VolumeMappingVerified || b.ChapterPageMappingVerified || b.RuntimeIngestionAllowed || b.ClaimEligible || b.PublishableAccuracy {
		return errors.New("OCR manifest boundary is not fail-closed")
	}
	return nil
}

func readOCRRecords(raw []byte) ([]*ocrRecord, ocrSummary, error) {
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	pages := make([]*ocrRecord, 0, 1187)
	summary := ocrSummary{MinimumLineScore: 10001}
	for scanner.Scan() {
		var record ocrRecord
		if err := decodeStrict(scanner.Bytes(), &record); err != nil {
			return nil, summary, fmt.Errorf("OCR record %d: %w", len(pages)+1, err)
		}
		wantPart, wantPage := 1, len(pages)+1
		if wantPage > 1000 {
			wantPart, wantPage = 2, wantPage-1000
		}
		if record.Part != wantPart || record.PhysicalPage != wantPage || record.Render.WidthPixels <= 0 || record.Render.HeightPixels <= 0 || !validHex(record.Render.SHA256, 64) {
			return nil, summary, fmt.Errorf("non-contiguous or invalid OCR page at record %d", len(pages)+1)
		}
		if len(record.Lines) == 0 {
			summary.EmptyPages++
		}
		for _, line := range record.Lines {
			if strings.ContainsAny(line.Text, "\r\n") || len(line.Box) != 4 || line.ScoreBasisPoints < 5000 || line.ScoreBasisPoints > 10000 {
				return nil, summary, fmt.Errorf("invalid OCR line at part %d page %d", record.Part, record.PhysicalPage)
			}
			summary.RecognizedLines++
			summary.ScoreBasisPointsSum += int64(line.ScoreBasisPoints)
			if line.ScoreBasisPoints < summary.MinimumLineScore {
				summary.MinimumLineScore = line.ScoreBasisPoints
			}
			if line.ScoreBasisPoints > summary.MaximumLineScore {
				summary.MaximumLineScore = line.ScoreBasisPoints
			}
		}
		pages = append(pages, &record)
	}
	if err := scanner.Err(); err != nil {
		return nil, summary, err
	}
	summary.PageCount = len(pages)
	if len(pages) >= 1000 {
		summary.Part1Pages = 1000
		summary.Part2Pages = len(pages) - 1000
	}
	return pages, summary, nil
}

func validateVolumeLabels(observation volumeLabelObservation, pages []*ocrRecord) error {
	if observation.Schema != "sanming_ncl_volume_label_observation_v1" || observation.Version != "2026-07-17.1" ||
		observation.Status != "single_operator_volume_mapping_candidates_not_gold" || observation.CandidateID != "sanming-ncl-06589-1578-12vol-scan-v1" ||
		observation.Method.MappingManifestSHA256 != expectedVolumeMappingSHA256 || len(observation.Observations) != 12 {
		return errors.New("unexpected volume-label observation identity")
	}
	b := observation.Boundaries
	if b.IndependentReviewComplete || b.VolumeMappingVerified || b.ChapterPageMappingVerified || b.RuntimeIngestionAllowed || b.ClaimEligible || b.PublishableAccuracy {
		return errors.New("volume-label observation boundary is not fail-closed")
	}
	seenPages := make(map[pageKey]bool, len(pages))
	for _, page := range pages {
		seenPages[pageKey{Part: page.Part, Page: page.PhysicalPage}] = true
	}
	seenCandidatePages := map[pageKey]bool{}
	for index, entry := range observation.Observations {
		volume := index + 1
		if entry.BookCandidate != volume || entry.VolumeCandidate != volume || len(entry.Source.PhysicalBookSegments) == 0 {
			return fmt.Errorf("invalid volume-label entry %d", volume)
		}
		for _, segment := range entry.Source.PhysicalBookSegments {
			if segment.Part < 1 || segment.Part > 2 || segment.FirstPage < 1 || segment.LastPage < segment.FirstPage {
				return fmt.Errorf("invalid volume %d segment", volume)
			}
			for page := segment.FirstPage; page <= segment.LastPage; page++ {
				key := pageKey{Part: segment.Part, Page: page}
				if !seenPages[key] || seenCandidatePages[key] {
					return fmt.Errorf("missing or duplicate volume candidate page %+v", key)
				}
				seenCandidatePages[key] = true
			}
		}
	}
	if len(seenCandidatePages) != len(pages) {
		return fmt.Errorf("volume segments cover %d OCR pages, want %d", len(seenCandidatePages), len(pages))
	}
	return nil
}

func validateVolumeComparison(comparison volumeComparison) error {
	if comparison.Schema != "sanming_cross_source_text_comparison_v1" || comparison.Version != "2026-07-17.1" ||
		comparison.Status != "machine_comparison_not_independence_or_claim_adjudication" || len(comparison.Chapters) != 381 ||
		comparison.Sources.MarkdownManifestSHA256 != expectedMarkdownManifestSHA256 || comparison.Sources.MarkdownFileCount != 382 ||
		comparison.Sources.NumberedChapterCount != 381 || comparison.Boundaries.PageMappingVerified ||
		comparison.Boundaries.RuntimeIngestionAllowed || comparison.Boundaries.ClaimEligible || comparison.Boundaries.PublishableAccuracy {
		return errors.New("unexpected volume comparison identity or boundary")
	}
	for index, chapter := range comparison.Chapters {
		if chapter.Chapter != index+1 || chapter.File != fmt.Sprintf("%03d.md", index+1) || len(chapter.BestCandidateVolumes) != 1 ||
			chapter.BestCandidateVolumes[0] < 1 || chapter.BestCandidateVolumes[0] > 12 {
			return fmt.Errorf("chapter %d does not have one valid best volume candidate", index+1)
		}
	}
	return nil
}

func readLocalChapters(root string, comparison volumeComparison) ([]*localChapter, int, []byte, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, 0, nil, err
	}
	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	manifest := bytes.Buffer{}
	chapters := make([]*localChapter, 0, 381)
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(root, name))
		if err != nil {
			return nil, 0, nil, err
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
			return nil, 0, nil, fmt.Errorf("%s: %w", name, err)
		}
		want := comparison.Chapters[number-1]
		if want.File != name || want.Title != title || want.OriginalSHA256 != hashBytes([]byte(original)) {
			return nil, 0, nil, fmt.Errorf("%s differs from frozen volume comparison", name)
		}
		chapters = append(chapters, &localChapter{Chapter: number, File: name, Title: title, Original: original})
	}
	if len(names) != 382 || len(chapters) != 381 {
		return nil, 0, nil, fmt.Errorf("Markdown inputs = %d files/%d chapters, want 382/381", len(names), len(chapters))
	}
	for index, chapter := range chapters {
		if chapter.Chapter != index+1 {
			return nil, 0, nil, errors.New("numbered Markdown chapters must be contiguous")
		}
	}
	return chapters, len(names), manifest.Bytes(), nil
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

func normalizeInputsWithOpenCC(binary, config string, chapters []*localChapter, pages []*ocrRecord) error {
	chapterLines := make([]string, 0, len(chapters)*2)
	for _, chapter := range chapters {
		chapterLines = append(chapterLines, normalizeText(chapter.Title), normalizeText(chapter.Original))
	}
	convertedChapters, err := convertLinesOpenCC(binary, config, chapterLines)
	if err != nil {
		return fmt.Errorf("OpenCC Markdown: %w", err)
	}
	for index, chapter := range chapters {
		chapter.TitleNorm = normalizeText(convertedChapters[index*2])
		chapter.OriginalNorm = normalizeText(convertedChapters[index*2+1])
		if chapter.TitleNorm == "" || len([]rune(chapter.OriginalNorm)) < contentNGramWidth {
			return fmt.Errorf("chapter %d has no comparable normalized content", chapter.Chapter)
		}
	}
	ocrLines := make([]string, 0, 25247)
	for _, page := range pages {
		for _, line := range page.Lines {
			ocrLines = append(ocrLines, normalizeText(line.Text))
		}
	}
	convertedOCR, err := convertLinesOpenCC(binary, config, ocrLines)
	if err != nil {
		return fmt.Errorf("OpenCC OCR: %w", err)
	}
	lineIndex := 0
	for _, page := range pages {
		for index := range page.Lines {
			page.Lines[index].Text = normalizeText(convertedOCR[lineIndex])
			lineIndex++
		}
	}
	return nil
}

func buildPageIndexes(pages []*ocrRecord) map[pageKey]pageIndex {
	indexes := make(map[pageKey]pageIndex, len(pages))
	for _, page := range pages {
		index := pageIndex{Key: pageKey{Part: page.Part, Page: page.PhysicalPage}, Grams: map[string]struct{}{}}
		for _, line := range page.Lines {
			index.Lines = append(index.Lines, line.Text)
			for gram := range runeNGrams(line.Text, contentNGramWidth) {
				index.Grams[gram] = struct{}{}
			}
		}
		indexes[index.Key] = index
	}
	return indexes
}

func mapChapters(chapters []*localChapter, comparison volumeComparison, labels volumeLabelObservation, pages map[pageKey]pageIndex) ([]chapterCandidate, reportSummary, error) {
	result := make([]chapterCandidate, 0, len(chapters))
	summary := reportSummary{ChapterCount: len(chapters)}
	for index, chapter := range chapters {
		volume := comparison.Chapters[index].BestCandidateVolumes[0]
		searchPages := pagesForSegments(labels.Observations[volume-1].Source.PhysicalBookSegments)
		chapterGrams := runeNGrams(chapter.OriginalNorm, contentNGramWidth)
		scores := make([]scoredPage, 0, len(searchPages))
		var titlePages []pageKey
		for _, key := range searchPages {
			page, ok := pages[key]
			if !ok {
				return nil, summary, fmt.Errorf("volume %d references missing OCR page %+v", volume, key)
			}
			for _, line := range page.Lines {
				if strings.Contains(line, chapter.TitleNorm) {
					titlePages = append(titlePages, key)
					break
				}
			}
			overlap := 0
			for gram := range chapterGrams {
				if _, ok := page.Grams[gram]; ok {
					overlap++
				}
			}
			scores = append(scores, scoredPage{Part: key.Part, PhysicalPage: key.Page, OverlapCount: overlap})
		}
		sort.Slice(scores, func(i, j int) bool {
			if scores[i].OverlapCount != scores[j].OverlapCount {
				return scores[i].OverlapCount > scores[j].OverlapCount
			}
			return pageLess(pageKey{Part: scores[i].Part, Page: scores[i].PhysicalPage}, pageKey{Part: scores[j].Part, Page: scores[j].PhysicalPage})
		})
		best := scores[0].OverlapCount
		secondDistinct := 0
		for _, score := range scores {
			if score.OverlapCount < best {
				secondDistinct = score.OverlapCount
				break
			}
		}
		var bestPages []pageKey
		for _, score := range scores {
			if score.OverlapCount != best {
				break
			}
			bestPages = append(bestPages, pageKey{Part: score.Part, Page: score.PhysicalPage})
		}
		topLimit := maximumContentTopCandidates
		if len(scores) < topLimit {
			topLimit = len(scores)
		}
		candidate := chapterCandidate{
			Chapter: chapter.Chapter, File: chapter.File, Title: chapter.Title, OriginalSHA256: hashBytes([]byte(chapter.Original)),
			VolumeCandidate: volume, SearchPageCount: len(searchPages), TitleLocatorCandidates: nonNilPageKeys(titlePages),
			TitleLocatorCandidateCount: len(titlePages), ContentNGramCount: len(chapterGrams), BestContentOverlap: best,
			SecondDistinctOverlap: secondDistinct, ContentOverlapMargin: best - secondDistinct,
			BestContentCandidateCount: len(bestPages), BestContentCandidates: nonNilPageKeys(bestPages),
			ContentTopCandidates: append([]scoredPage(nil), scores[:topLimit]...), ZeroContentOverlap: best == 0,
		}
		updateSummary(&summary, candidate)
		result = append(result, candidate)
	}
	return result, summary, nil
}

func updateSummary(summary *reportSummary, chapter chapterCandidate) {
	if chapter.TitleLocatorCandidateCount > 0 {
		summary.ExactTitleLocatedChapters++
	}
	if chapter.ZeroContentOverlap {
		summary.ZeroContentOverlapChapters++
	}
	if chapter.BestContentCandidateCount > 1 {
		summary.BestContentPageTieChapters++
	}
	if chapter.ContentOverlapMargin == 0 {
		summary.ZeroMarginChapters++
	}
	if chapter.ContentOverlapMargin <= 2 {
		summary.MarginAtMost2Chapters++
	}
	if chapter.BestContentOverlap >= 1 {
		summary.BestContentOverlapAtLeast1++
	}
	if chapter.BestContentOverlap >= 3 {
		summary.BestContentOverlapAtLeast3++
	}
	if chapter.BestContentOverlap >= 5 {
		summary.BestContentOverlapAtLeast5++
	}
	if chapter.BestContentOverlap >= 10 {
		summary.BestContentOverlapAtLeast10++
	}
	if chapter.BestContentOverlap >= 20 {
		summary.BestContentOverlapAtLeast20++
	}
	if chapter.BestContentOverlap >= 40 {
		summary.BestContentOverlapAtLeast40++
	}
}

func buildReport(manifest ocrManifest, markdownFiles int, openCCVersion string, openCCAssets map[string]string, ocr ocrSummary, summary reportSummary, chapters []chapterCandidate) auditReport {
	var report auditReport
	report.Schema = "sanming_ncl_chapter_page_candidate_audit_v1"
	report.Version = "2026-07-17.1"
	report.Status = "chapter_page_candidates_machine_only"
	report.ObservedAt = "2026-07-17"
	report.CandidateID = manifest.CandidateID
	report.Purpose = "Freeze reproducible title-locator and content-overlap page candidates for manual review without treating either machine signal as a verified chapter start page or supporting citation."
	report.Sources.OCRManifest = artifactReference{Path: "research/rag/ocr/sanming-ncl-06589-rapidocr-v1/ocr-manifest.json", SHA256: expectedOCRManifestSHA256}
	report.Sources.OCRArtifact = artifactReference{Path: "research/rag/ocr/sanming-ncl-06589-rapidocr-v1/page-ocr.jsonl", SHA256: expectedOCRArtifactSHA256, SizeBytes: manifest.Artifact.SizeBytes}
	report.Sources.VolumeLabelObservation = artifactReference{Path: "research/rag/snapshots/sanming-ncl-06589-1578-v1/volume-label-observation.json", SHA256: expectedVolumeLabelsSHA256}
	report.Sources.VolumeComparison = artifactReference{Path: "research/rag/sanming-wikisource-markdown-comparison-v1.json", SHA256: expectedVolumeComparisonSHA256}
	report.Sources.MarkdownRootLabel = "external_mingli_db/md/bazi/三命通会"
	report.Sources.MarkdownManifestScheme = "sorted_filename_tab_sha256_lf_v1"
	report.Sources.MarkdownManifestSHA256 = expectedMarkdownManifestSHA256
	report.Sources.MarkdownFileCount = markdownFiles
	report.Sources.NumberedChapterCount = len(chapters)
	report.Method.ScriptConversion = "opencc_t2s_fixed_assets"
	report.Method.OpenCCVersion = openCCVersion
	report.Method.OpenCCAssets = openCCAssets
	report.Method.CharacterFilter = "unicode_letters_and_numbers_without_compatibility_fold"
	report.Method.OCRConversionUnit = "each_ocr_line_independently_addressable_no_cross_line_concatenation"
	report.Method.ContentNGramWidth = contentNGramWidth
	report.Method.ContentNGramRule = "unique_normalized_4_rune_grams_per_ocr_line_then_page_set_intersection_with_unique_chapter_original_grams"
	report.Method.VolumeRestriction = "one_unique_best_volume_candidate_from_sanming_cross_source_text_comparison_v1_then_physical_book_segments"
	report.Method.SecondScoreRule = "highest_distinct_overlap_strictly_below_best_or_zero_if_none"
	report.Method.CandidateOrdering = "overlap_descending_then_part_and_physical_page_ascending"
	report.Method.MaximumContentTopCandidates = maximumContentTopCandidates
	report.OCRSummary = ocr
	report.Summary = summary
	report.Chapters = chapters
	report.Boundaries.MachineCandidatesOnly = true
	report.Boundaries.TitleLocatorCandidateNotStartPageGold = true
	report.Boundaries.ContentOverlapNotSupportingCitation = true
	report.Boundaries.OCRTextNotDiplomatic = true
	return report
}

func pagesForSegments(segments []bookSegment) []pageKey {
	var pages []pageKey
	for _, segment := range segments {
		for page := segment.FirstPage; page <= segment.LastPage; page++ {
			pages = append(pages, pageKey{Part: segment.Part, Page: page})
		}
	}
	return pages
}

func pageLess(a, b pageKey) bool {
	return a.Part < b.Part || (a.Part == b.Part && a.Page < b.Page)
}

func nonNilPageKeys(values []pageKey) []pageKey {
	if values == nil {
		return []pageKey{}
	}
	return values
}

func runeNGrams(text string, width int) map[string]struct{} {
	runes := []rune(text)
	grams := make(map[string]struct{}, len(runes))
	for index := 0; index+width <= len(runes); index++ {
		grams[string(runes[index:index+width])] = struct{}{}
	}
	return grams
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

func convertLinesOpenCC(binary, config string, lines []string) ([]string, error) {
	input := strings.Join(lines, "\n") + "\n"
	command := exec.Command(binary, "-c", config)
	command.Stdin = strings.NewReader(input)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	if err := command.Run(); err != nil {
		return nil, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
	}
	output := strings.TrimSuffix(stdout.String(), "\n")
	converted := strings.Split(output, "\n")
	if len(converted) != len(lines) {
		return nil, fmt.Errorf("OpenCC line count = %d, want %d", len(converted), len(lines))
	}
	return converted, nil
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

func readStrictJSON[T any](path string) ([]byte, T, error) {
	var value T
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, value, err
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if err := decoder.Decode(&value); err != nil {
		return nil, value, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return nil, value, errors.New("input must contain one JSON document")
	}
	return raw, value, nil
}

func decodeStrict(raw []byte, value any) error {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
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
	temporary, err := os.CreateTemp(filepath.Dir(path), ".sanming-page-map-*.json")
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
	if len(value) != length || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
