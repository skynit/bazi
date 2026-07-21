package bazi

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	ragNCLOCRRoot                = "../../../../research/rag/ocr/sanming-ncl-06589-rapidocr-v1"
	ragNCLOCRManifestSHA256      = "e072f6a467ea83f55f63687b1b936756071dbc8a87862937be4f2478c87dce27"
	ragNCLOCRArtifactSHA256      = "c7fd70b653015a99949dd9926c07660efa6744560a0dd44c4c9a54d8f6d8c164"
	ragNCLPageCandidatesPath     = "../../../../research/rag/sanming-ncl-06589-chapter-page-candidates-v1.json"
	ragNCLPageCandidatesSHA256   = "0f915acaea01df0fd66738097105ba42786850b7fd9e323d9db3529693d0ce65"
	ragNCLVolumeComparisonSHA256 = "ce7ee55bd4c6b68c488f3326b9b3725860b5ebea0c679ee12d156f330a180b7f"
	ragNCLMarkdownManifestSHA256 = "ddef8e89beac4d106336d17e47899e8220b139c36117524b3ea71518dbb91bb7"
)

type ragNCLOCRManifest struct {
	Schema      string `json:"schema"`
	Version     string `json:"version"`
	Status      string `json:"status"`
	GeneratedAt string `json:"generated_at"`
	CandidateID string `json:"candidate_id"`
	Purpose     string `json:"purpose"`
	Inputs      struct {
		ScanSnapshotManifest   ragCommonsArtifactReference `json:"scan_snapshot_manifest"`
		VolumeLabelObservation ragCommonsArtifactReference `json:"volume_label_observation"`
	} `json:"inputs"`
	Rendering struct {
		Tool         string `json:"tool"`
		Arguments    string `json:"arguments"`
		ColorMode    string `json:"color_mode"`
		PageIdentity string `json:"page_identity"`
	} `json:"rendering"`
	OCR struct {
		Package              string  `json:"package"`
		PackageVersion       string  `json:"package_version"`
		PackageLicense       string  `json:"package_license"`
		ONNXRuntimeVersion   string  `json:"onnxruntime_version"`
		OpenCVPythonVersion  string  `json:"opencv_python_version"`
		UseDetection         bool    `json:"use_detection"`
		UseClassification    bool    `json:"use_classification"`
		UseRecognition       bool    `json:"use_recognition"`
		WorkerProcesses      int     `json:"worker_processes"`
		IntraOpThreads       int     `json:"intra_op_threads_per_session"`
		InterOpThreads       int     `json:"inter_op_threads_per_session"`
		RecognitionBatchSize int     `json:"recognition_batch_size"`
		TextScoreThreshold   float64 `json:"text_score_threshold"`
		ConfigSHA256         string  `json:"config_sha256"`
		Models               struct {
			Detection            ragNCLOCRModel `json:"detection"`
			Recognition          ragNCLOCRModel `json:"recognition"`
			ClassificationUnused ragNCLOCRModel `json:"classification_unused"`
		} `json:"models"`
	} `json:"ocr"`
	Artifact   ragNCLOCRArtifact   `json:"artifact"`
	Summary    ragNCLOCRSummary    `json:"summary"`
	Boundaries ragNCLOCRBoundaries `json:"boundaries"`
}

type ragNCLOCRModel struct {
	Name   string `json:"name"`
	SHA256 string `json:"sha256"`
}

type ragNCLOCRArtifact struct {
	Path         string `json:"path"`
	SizeBytes    int64  `json:"size_bytes"`
	SHA256       string `json:"sha256"`
	RecordScheme string `json:"record_scheme"`
}

type ragNCLOCRSummary struct {
	PageCount           int   `json:"page_count"`
	Part1Pages          int   `json:"part_1_pages"`
	Part2Pages          int   `json:"part_2_pages"`
	EmptyPages          int   `json:"empty_pages"`
	RecognizedLines     int   `json:"recognized_lines"`
	ScoreBasisPointsSum int64 `json:"score_basis_points_sum"`
	MinimumLineScore    int   `json:"minimum_line_score_basis_points"`
	MaximumLineScore    int   `json:"maximum_line_score_basis_points"`
}

type ragNCLOCRBoundaries struct {
	CompletePageSetObserved     bool `json:"complete_page_set_observed"`
	MachineOCROnly              bool `json:"machine_ocr_only"`
	OCRTextNotDiplomatic        bool `json:"ocr_text_is_not_diplomatic_transcription"`
	OCRScoreNotMapping          bool `json:"ocr_score_is_not_mapping_confidence"`
	IndependentReviewComplete   bool `json:"independent_review_complete"`
	CompletePrimaryTextVerified bool `json:"complete_primary_text_verified"`
	VolumeMappingVerified       bool `json:"volume_mapping_verified"`
	ChapterPageMappingVerified  bool `json:"chapter_page_mapping_verified"`
	ClaimSupportReviewed        bool `json:"claim_support_reviewed"`
	RuntimeIngestionAllowed     bool `json:"runtime_ingestion_allowed"`
	ClaimEligible               bool `json:"claim_eligible"`
	PublishableAccuracy         bool `json:"publishable_accuracy"`
}

type ragNCLOCRPage struct {
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

type ragNCLPageCandidateReport struct {
	Schema      string                        `json:"schema"`
	Version     string                        `json:"version"`
	Status      string                        `json:"status"`
	ObservedAt  string                        `json:"observed_at"`
	CandidateID string                        `json:"candidate_id"`
	Purpose     string                        `json:"purpose"`
	Sources     ragNCLPageCandidateSources    `json:"sources"`
	Method      ragNCLPageCandidateMethod     `json:"method"`
	OCRSummary  ragNCLOCRSummary              `json:"ocr_summary"`
	Summary     ragNCLPageCandidateSummary    `json:"summary"`
	Chapters    []ragNCLPageCandidateChapter  `json:"chapters"`
	Boundaries  ragNCLPageCandidateBoundaries `json:"boundaries"`
}

type ragNCLArtifactReference struct {
	Path      string `json:"path"`
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes,omitempty"`
}

type ragNCLPageCandidateSources struct {
	OCRManifest            ragNCLArtifactReference `json:"ocr_manifest"`
	OCRArtifact            ragNCLArtifactReference `json:"ocr_artifact"`
	VolumeLabelObservation ragNCLArtifactReference `json:"volume_label_observation"`
	VolumeComparison       ragNCLArtifactReference `json:"volume_comparison"`
	MarkdownRootLabel      string                  `json:"markdown_root_label"`
	MarkdownManifestScheme string                  `json:"markdown_manifest_scheme"`
	MarkdownManifestSHA256 string                  `json:"markdown_manifest_sha256"`
	MarkdownFileCount      int                     `json:"markdown_file_count"`
	NumberedChapterCount   int                     `json:"numbered_chapter_count"`
}

type ragNCLPageCandidateMethod struct {
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
}

type ragNCLPageKey struct {
	Part         int `json:"part"`
	PhysicalPage int `json:"physical_page"`
}

type ragNCLScoredPage struct {
	Part         int `json:"part"`
	PhysicalPage int `json:"physical_page"`
	OverlapCount int `json:"overlap_count"`
}

type ragNCLPageCandidateChapter struct {
	Chapter                    int                `json:"chapter"`
	File                       string             `json:"file"`
	Title                      string             `json:"title"`
	OriginalSHA256             string             `json:"original_sha256"`
	VolumeCandidate            int                `json:"volume_candidate"`
	SearchPageCount            int                `json:"search_page_count"`
	TitleLocatorCandidates     []ragNCLPageKey    `json:"title_locator_candidates"`
	TitleLocatorCandidateCount int                `json:"title_locator_candidate_count"`
	ContentNGramCount          int                `json:"content_ngram_count"`
	BestContentOverlap         int                `json:"best_content_overlap"`
	SecondDistinctOverlap      int                `json:"second_distinct_content_overlap"`
	ContentOverlapMargin       int                `json:"content_overlap_margin"`
	BestContentCandidateCount  int                `json:"best_content_candidate_count"`
	BestContentCandidates      []ragNCLPageKey    `json:"best_content_candidates"`
	ContentTopCandidates       []ragNCLScoredPage `json:"content_top_candidates"`
	ZeroContentOverlap         bool               `json:"zero_content_overlap"`
}

type ragNCLPageCandidateSummary struct {
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

type ragNCLPageCandidateBoundaries struct {
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
}

func TestRAGNCLOCRSnapshotContract(t *testing.T) {
	manifest := ragCommonsReadStrictJSON[ragNCLOCRManifest](t, filepath.Join(ragNCLOCRRoot, "ocr-manifest.json"), ragNCLOCRManifestSHA256)
	if manifest.Schema != "sanming_ncl_page_ocr_snapshot_v1" || manifest.Version != "2026-07-17.1" ||
		manifest.Status != "machine_ocr_silver_not_page_mapping_gold" || manifest.GeneratedAt != "2026-07-17" ||
		manifest.CandidateID != ragCommonsCandidateID || strings.TrimSpace(manifest.Purpose) == "" {
		t.Fatalf("unexpected NCL OCR manifest identity: %+v", manifest)
	}
	if manifest.Inputs.ScanSnapshotManifest != (ragCommonsArtifactReference{Path: "../../snapshots/sanming-ncl-06589-1578-v1/snapshot-manifest.json", SHA256: ragCommonsSnapshotManifestSHA256}) ||
		manifest.Inputs.VolumeLabelObservation != (ragCommonsArtifactReference{Path: "../../snapshots/sanming-ncl-06589-1578-v1/volume-label-observation.json", SHA256: ragNCLVolumeLabelObservationSHA256}) {
		t.Fatalf("unexpected NCL OCR inputs: %+v", manifest.Inputs)
	}
	if manifest.Rendering.Tool != "pdftoppm version 26.05.0" || manifest.Rendering.Arguments != "-jpeg -jpegopt quality=90 -r 72" ||
		manifest.Rendering.ColorMode != "source RGB JPEG" || manifest.Rendering.PageIdentity != "part plus one-based physical PDF page" {
		t.Fatalf("unexpected OCR rendering profile: %+v", manifest.Rendering)
	}
	o := manifest.OCR
	if o.Package != "rapidocr-onnxruntime" || o.PackageVersion != "1.4.4" || o.PackageLicense != "Apache-2.0" ||
		o.ONNXRuntimeVersion != "1.27.0" || o.OpenCVPythonVersion != "5.0.0.93" || !o.UseDetection || o.UseClassification || !o.UseRecognition ||
		o.WorkerProcesses != 4 || o.IntraOpThreads != 2 || o.InterOpThreads != 1 || o.RecognitionBatchSize != 6 || o.TextScoreThreshold != 0.5 ||
		o.ConfigSHA256 != "bf94a1da4cba828e67b1d61e27cee14d9e7da27c9f272e04048a17e41ae97332" ||
		o.Models.Detection != (ragNCLOCRModel{Name: "ch_PP-OCRv4_det_infer.onnx", SHA256: "d2a7720d45a54257208b1e13e36a8479894cb74155a5efe29462512d42f49da9"}) ||
		o.Models.Recognition != (ragNCLOCRModel{Name: "ch_PP-OCRv4_rec_infer.onnx", SHA256: "48fc40f24f6d2a207a2b1091d3437eb3cc3eb6b676dc3ef9c37384005483683b"}) ||
		o.Models.ClassificationUnused != (ragNCLOCRModel{Name: "ch_ppocr_mobile_v2.0_cls_infer.onnx", SHA256: "e47acedf663230f8863ff1ab0e64dd2d82b838fceb5957146dab185a89d6215c"}) {
		t.Fatalf("unexpected OCR engine profile: %+v", o)
	}
	if manifest.Artifact != (ragNCLOCRArtifact{Path: "page-ocr.jsonl", SizeBytes: 3232151, SHA256: ragNCLOCRArtifactSHA256, RecordScheme: "part_physical_page_render_identity_and_ordered_line_boxes_text_score_v1"}) {
		t.Fatalf("unexpected OCR artifact: %+v", manifest.Artifact)
	}
	wantSummary := ragNCLOCRSummary{PageCount: 1187, Part1Pages: 1000, Part2Pages: 187, EmptyPages: 14, RecognizedLines: 25247, ScoreBasisPointsSum: 203304267, MinimumLineScore: 5000, MaximumLineScore: 10000}
	if manifest.Summary != wantSummary {
		t.Fatalf("unexpected OCR summary: %+v", manifest.Summary)
	}
	b := manifest.Boundaries
	if !b.CompletePageSetObserved || !b.MachineOCROnly || !b.OCRTextNotDiplomatic || !b.OCRScoreNotMapping ||
		b.IndependentReviewComplete || b.CompletePrimaryTextVerified || b.VolumeMappingVerified || b.ChapterPageMappingVerified ||
		b.ClaimSupportReviewed || b.RuntimeIngestionAllowed || b.ClaimEligible || b.PublishableAccuracy {
		t.Fatalf("OCR boundaries must remain fail-closed: %+v", b)
	}
	ragNCLValidateOCRArtifact(t, manifest.Artifact, wantSummary)
}

func ragNCLValidateOCRArtifact(t *testing.T, artifact ragNCLOCRArtifact, want ragNCLOCRSummary) {
	t.Helper()
	path := filepath.Join(ragNCLOCRRoot, artifact.Path)
	info, err := os.Stat(path)
	if err != nil || info.Size() != artifact.SizeBytes || ragCommonsHashFile(t, path) != ragNCLOCRArtifactSHA256 {
		t.Fatalf("OCR artifact identity mismatch: info=%v err=%v", info, err)
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 2*1024*1024)
	got := ragNCLOCRSummary{MinimumLineScore: 10001}
	for scanner.Scan() {
		var page ragNCLOCRPage
		decoder := json.NewDecoder(bytes.NewReader(scanner.Bytes()))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&page); err != nil {
			t.Fatalf("OCR record %d: %v", got.PageCount+1, err)
		}
		got.PageCount++
		wantPart, wantPage := 1, got.PageCount
		if wantPage > 1000 {
			wantPart, wantPage = 2, wantPage-1000
		}
		if page.Part != wantPart || page.PhysicalPage != wantPage || page.Render.WidthPixels <= 0 || page.Render.HeightPixels <= 0 || len(page.Render.SHA256) != 64 {
			t.Fatalf("invalid or non-contiguous OCR page at record %d: %+v", got.PageCount, page)
		}
		if len(page.Lines) == 0 {
			got.EmptyPages++
		}
		for _, line := range page.Lines {
			if len(line.Box) != 4 || strings.ContainsAny(line.Text, "\r\n") || line.ScoreBasisPoints < 5000 || line.ScoreBasisPoints > 10000 {
				t.Fatalf("invalid OCR line at part %d page %d: %+v", page.Part, page.PhysicalPage, line)
			}
			got.RecognizedLines++
			got.ScoreBasisPointsSum += int64(line.ScoreBasisPoints)
			if line.ScoreBasisPoints < got.MinimumLineScore {
				got.MinimumLineScore = line.ScoreBasisPoints
			}
			if line.ScoreBasisPoints > got.MaximumLineScore {
				got.MaximumLineScore = line.ScoreBasisPoints
			}
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
	got.Part1Pages, got.Part2Pages = 1000, got.PageCount-1000
	if got != want {
		t.Fatalf("OCR JSONL observed summary = %+v, want %+v", got, want)
	}
}

func TestRAGNCLChapterPageCandidatesContract(t *testing.T) {
	report := ragCommonsReadStrictJSON[ragNCLPageCandidateReport](t, ragNCLPageCandidatesPath, ragNCLPageCandidatesSHA256)
	if report.Schema != "sanming_ncl_chapter_page_candidate_audit_v1" || report.Version != "2026-07-17.1" ||
		report.Status != "chapter_page_candidates_machine_only" || report.ObservedAt != "2026-07-17" ||
		report.CandidateID != ragCommonsCandidateID || strings.TrimSpace(report.Purpose) == "" {
		t.Fatalf("unexpected chapter-page report identity: %+v", report)
	}
	ragNCLValidatePageCandidateSourcesAndMethod(t, report)
	ragNCLValidatePageCandidateChapters(t, report)
	b := report.Boundaries
	if !b.MachineCandidatesOnly || !b.TitleLocatorCandidateNotStartPageGold || !b.ContentOverlapNotSupportingCitation || !b.OCRTextNotDiplomatic ||
		b.IndependentReviewComplete || b.VolumeMappingVerified || b.ChapterPageMappingVerified || b.ClaimSupportReviewed ||
		b.RuntimeIngestionAllowed || b.ClaimEligible || b.PublishableAccuracy {
		t.Fatalf("chapter-page candidate boundaries must remain fail-closed: %+v", b)
	}
}

func ragNCLValidatePageCandidateSourcesAndMethod(t *testing.T, report ragNCLPageCandidateReport) {
	t.Helper()
	s := report.Sources
	if s.OCRManifest != (ragNCLArtifactReference{Path: "research/rag/ocr/sanming-ncl-06589-rapidocr-v1/ocr-manifest.json", SHA256: ragNCLOCRManifestSHA256}) ||
		s.OCRArtifact != (ragNCLArtifactReference{Path: "research/rag/ocr/sanming-ncl-06589-rapidocr-v1/page-ocr.jsonl", SHA256: ragNCLOCRArtifactSHA256, SizeBytes: 3232151}) ||
		s.VolumeLabelObservation != (ragNCLArtifactReference{Path: "research/rag/snapshots/sanming-ncl-06589-1578-v1/volume-label-observation.json", SHA256: ragNCLVolumeLabelObservationSHA256}) ||
		s.VolumeComparison != (ragNCLArtifactReference{Path: "research/rag/sanming-wikisource-markdown-comparison-v1.json", SHA256: ragNCLVolumeComparisonSHA256}) ||
		s.MarkdownRootLabel != "external_mingli_db/md/bazi/三命通会" || s.MarkdownManifestScheme != "sorted_filename_tab_sha256_lf_v1" ||
		s.MarkdownManifestSHA256 != ragNCLMarkdownManifestSHA256 || s.MarkdownFileCount != 382 || s.NumberedChapterCount != 381 {
		t.Fatalf("unexpected chapter-page sources: %+v", s)
	}
	m := report.Method
	expectedAssets := map[string]string{
		"t2s.json":                          "96fe5cc374a80ccc49e3370006cce3aefe4af955868ae0b14fb3079ec695be4f",
		"CJK_Compatibility_Ideographs.ocd2": "4b1faa6649012f524068ec18c0fb520ead343c11cbe0a8e4c8853ca61369d666",
		"TSPhrases.ocd2":                    "e7f9d419d54f71a66d7f0283b29910913f08defdb6d4322e00c459c7ebe3f991",
		"TSCharactersExt.ocd2":              "2ee61f852d05a3241326ae8d7eeae00818a80c0a0f4e03704050312b9561bf33",
		"TSCharacters.ocd2":                 "014a1c9615f2a0800a56f0e6ce132c01ec233b089cd6160da66df9c346c0274b",
	}
	if m.ScriptConversion != "opencc_t2s_fixed_assets" || m.OpenCCVersion != "1.3.2.dirty" || fmt.Sprint(m.OpenCCAssets) != fmt.Sprint(expectedAssets) ||
		m.CharacterFilter != "unicode_letters_and_numbers_without_compatibility_fold" ||
		m.OCRConversionUnit != "each_ocr_line_independently_addressable_no_cross_line_concatenation" || m.ContentNGramWidth != 4 ||
		m.ContentNGramRule != "unique_normalized_4_rune_grams_per_ocr_line_then_page_set_intersection_with_unique_chapter_original_grams" ||
		m.VolumeRestriction != "one_unique_best_volume_candidate_from_sanming_cross_source_text_comparison_v1_then_physical_book_segments" ||
		m.SecondScoreRule != "highest_distinct_overlap_strictly_below_best_or_zero_if_none" ||
		m.CandidateOrdering != "overlap_descending_then_part_and_physical_page_ascending" || m.MaximumContentTopCandidates != 5 {
		t.Fatalf("unexpected chapter-page method: %+v", m)
	}
	wantOCR := ragNCLOCRSummary{PageCount: 1187, Part1Pages: 1000, Part2Pages: 187, EmptyPages: 14, RecognizedLines: 25247, ScoreBasisPointsSum: 203304267, MinimumLineScore: 5000, MaximumLineScore: 10000}
	if report.OCRSummary != wantOCR {
		t.Fatalf("report OCR summary = %+v, want %+v", report.OCRSummary, wantOCR)
	}
}

func ragNCLValidatePageCandidateChapters(t *testing.T, report ragNCLPageCandidateReport) {
	t.Helper()
	observation := ragCommonsReadStrictJSON[ragNCLVolumeLabelObservation](t, filepath.Join(ragCommonsSnapshotRoot, "volume-label-observation.json"), ragNCLVolumeLabelObservationSHA256)
	if len(report.Chapters) != 381 {
		t.Fatalf("chapter candidates = %d, want 381", len(report.Chapters))
	}
	var got ragNCLPageCandidateSummary
	got.ChapterCount = len(report.Chapters)
	volumeCounts := make([]int, 12)
	for index, chapter := range report.Chapters {
		if chapter.Chapter != index+1 || chapter.File != fmt.Sprintf("%03d.md", index+1) || chapter.VolumeCandidate < 1 || chapter.VolumeCandidate > 12 ||
			len(chapter.OriginalSHA256) != 64 || chapter.ContentNGramCount <= 0 || chapter.TitleLocatorCandidateCount != len(chapter.TitleLocatorCandidates) ||
			chapter.BestContentCandidateCount != len(chapter.BestContentCandidates) || len(chapter.ContentTopCandidates) == 0 || len(chapter.ContentTopCandidates) > 5 ||
			chapter.ContentTopCandidates[0].OverlapCount != chapter.BestContentOverlap || chapter.ContentOverlapMargin != chapter.BestContentOverlap-chapter.SecondDistinctOverlap ||
			chapter.ZeroContentOverlap != (chapter.BestContentOverlap == 0) {
			t.Fatalf("invalid chapter-page candidate %d: %+v", index+1, chapter)
		}
		segments := observation.Observations[chapter.VolumeCandidate-1].Source.PhysicalBookSegments
		searchPages := 0
		for _, segment := range segments {
			searchPages += segment.LastPage - segment.FirstPage + 1
		}
		if chapter.SearchPageCount != searchPages {
			t.Fatalf("chapter %d search pages = %d, want %d", chapter.Chapter, chapter.SearchPageCount, searchPages)
		}
		for _, page := range append(append([]ragNCLPageKey{}, chapter.TitleLocatorCandidates...), chapter.BestContentCandidates...) {
			if !ragNCLPageWithinSegments(page.Part, page.PhysicalPage, segments) {
				t.Fatalf("chapter %d page %+v is outside volume %d", chapter.Chapter, page, chapter.VolumeCandidate)
			}
		}
		for candidateIndex, page := range chapter.ContentTopCandidates {
			if !ragNCLPageWithinSegments(page.Part, page.PhysicalPage, segments) {
				t.Fatalf("chapter %d top page %+v is outside volume", chapter.Chapter, page)
			}
			if candidateIndex > 0 {
				previous := chapter.ContentTopCandidates[candidateIndex-1]
				if page.OverlapCount > previous.OverlapCount || (page.OverlapCount == previous.OverlapCount && !ragNCLPageCandidatePageLess(previous.Part, previous.PhysicalPage, page.Part, page.PhysicalPage)) {
					t.Fatalf("chapter %d top candidates are not canonical: %+v", chapter.Chapter, chapter.ContentTopCandidates)
				}
			}
		}
		if chapter.BestContentOverlap == 0 {
			if chapter.SecondDistinctOverlap != 0 || chapter.BestContentCandidateCount != chapter.SearchPageCount {
				t.Fatalf("zero-overlap chapter %d must leave every searched page tied", chapter.Chapter)
			}
		} else if chapter.SecondDistinctOverlap >= chapter.BestContentOverlap {
			t.Fatalf("chapter %d second distinct score is not below best", chapter.Chapter)
		}
		volumeCounts[chapter.VolumeCandidate-1]++
		ragNCLAccumulatePageSummary(&got, chapter)
	}
	wantSummary := ragNCLPageCandidateSummary{ChapterCount: 381, ExactTitleLocatedChapters: 58, ZeroContentOverlapChapters: 9, BestContentPageTieChapters: 25, ZeroMarginChapters: 9, MarginAtMost2Chapters: 56, BestContentOverlapAtLeast1: 372, BestContentOverlapAtLeast3: 346, BestContentOverlapAtLeast5: 334, BestContentOverlapAtLeast10: 317, BestContentOverlapAtLeast20: 279, BestContentOverlapAtLeast40: 234}
	if got != wantSummary || report.Summary != wantSummary {
		t.Fatalf("chapter-page summary mismatch: got=%+v report=%+v want=%+v", got, report.Summary, wantSummary)
	}
	if fmt.Sprint(volumeCounts) != fmt.Sprint([]int{13, 26, 23, 7, 47, 45, 22, 60, 60, 41, 21, 16}) {
		t.Fatalf("chapter volume counts = %v", volumeCounts)
	}
	ragNCLValidatePageCandidateSentinels(t, report.Chapters)
}

func ragNCLAccumulatePageSummary(summary *ragNCLPageCandidateSummary, chapter ragNCLPageCandidateChapter) {
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

func ragNCLValidatePageCandidateSentinels(t *testing.T, chapters []ragNCLPageCandidateChapter) {
	t.Helper()
	first := chapters[0]
	if first.Title != "原造化之始" || fmt.Sprint(first.TitleLocatorCandidates) != fmt.Sprint([]ragNCLPageKey{{Part: 1, PhysicalPage: 10}}) ||
		fmt.Sprint(first.BestContentCandidates) != fmt.Sprint([]ragNCLPageKey{{Part: 1, PhysicalPage: 13}}) || first.BestContentOverlap != 52 || first.SecondDistinctOverlap != 42 || first.ContentOverlapMargin != 10 {
		t.Fatalf("chapter 1 title and content locators must remain distinct: %+v", first)
	}
	second := chapters[1]
	if second.Title != "论五行生成" || len(second.TitleLocatorCandidates) != 0 || fmt.Sprint(second.BestContentCandidates) != fmt.Sprint([]ragNCLPageKey{{Part: 1, PhysicalPage: 19}}) || second.BestContentOverlap != 11 || second.SecondDistinctOverlap != 1 {
		t.Fatalf("unexpected chapter 2 sentinel: %+v", second)
	}
	chapter358 := chapters[357]
	if chapter358.Title != "通玄子撰集" || fmt.Sprint(chapter358.TitleLocatorCandidates) != fmt.Sprint([]ragNCLPageKey{{Part: 1, PhysicalPage: 988}}) || !chapter358.ZeroContentOverlap || chapter358.BestContentCandidateCount != 111 {
		t.Fatalf("chapter 358 must preserve title-only locator uncertainty: %+v", chapter358)
	}
}

func ragNCLPageCandidatePageLess(partA, pageA, partB, pageB int) bool {
	return partA < partB || (partA == partB && pageA < pageB)
}

func TestRAGNCLPageCandidateGeneratorContract(t *testing.T) {
	raw, err := os.ReadFile("../../../../scripts/map-sanmintonghui-ncl-pages.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{
		"DisallowUnknownFields", ragNCLOCRManifestSHA256, ragNCLOCRArtifactSHA256, ragNCLVolumeLabelObservationSHA256,
		ragNCLVolumeComparisonSHA256, ragNCLMarkdownManifestSHA256, "contentNGramWidth",
		"each_ocr_line_independently_addressable_no_cross_line_concatenation", "highest_distinct_overlap_strictly_below_best_or_zero_if_none",
		"chapter_page_candidates_machine_only", "title_locator_candidate_not_start_page_gold", "content_overlap_candidate_not_supporting_citation",
	} {
		if !bytes.Contains(raw, []byte(marker)) {
			t.Fatalf("page-candidate generator missing %q", marker)
		}
	}
}

func TestRAGNCLPageCandidatesAreNotRuntimeRegistered(t *testing.T) {
	for _, sourcePath := range []string{
		"../localrag/index.go", "../localrag/retriever.go", "../interpretation/bazi.go", "../../model/dto.go",
		"../../../cmd/bazi-rag-index/main.go", "../../../../scripts/build-local-bazi-rag-index.sh", "../../../../scripts/build-ragflow-bazi-manifest.sh",
	} {
		raw, err := os.ReadFile(sourcePath)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{
			"sanming_ncl_page_ocr_snapshot_v1", "sanming_ncl_chapter_page_candidate_audit_v1", "sanming-ncl-06589-rapidocr-v1",
			"sanming-ncl-06589-chapter-page-candidates-v1.json", "page-ocr.jsonl", "title_locator_candidates", "content_overlap_candidates",
		} {
			if bytes.Contains(raw, []byte(forbidden)) {
				t.Fatalf("research-only NCL page marker %q leaked into %s", forbidden, sourcePath)
			}
		}
	}
}

func TestRAGNCLPageCandidateResearchDocumentsContract(t *testing.T) {
	marker := "第一百五十项完成 NCL 1578 年全页 OCR Silver 与381章机器页候选治理"
	for _, documentPath := range []string{
		"../../../../docs/fortune-accuracy-research-plan.md", "../../../../docs/fortune-accuracy-roadmap.md", "../../../../docs/precision-test-plan.md",
	} {
		raw, err := os.ReadFile(documentPath)
		if err != nil {
			t.Fatal(err)
		}
		if count := bytes.Count(raw, []byte(marker)); count != 1 {
			t.Fatalf("phase 150 marker count in %s = %d, want 1", documentPath, count)
		}
	}
}
