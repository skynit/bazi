package precision

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const defaultMingLiSourceDir = "/home/skynit/mingli_db/md"

var (
	ganZhiTokenPattern      = `[甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥]`
	labelledPillarsPattern  = regexp.MustCompile(`(` + ganZhiTokenPattern + `)\s*年[、,，\s]+(` + ganZhiTokenPattern + `)\s*月[、,，\s]+(` + ganZhiTokenPattern + `)\s*日[、,，\s]+(` + ganZhiTokenPattern + `)\s*时`)
	plainPillarsPattern     = regexp.MustCompile(`(?:^|[^\p{Han}])(` + ganZhiTokenPattern + `)[、,，\s]+(` + ganZhiTokenPattern + `)[、,，\s]+(` + ganZhiTokenPattern + `)[、,，\s]+(` + ganZhiTokenPattern + `)(?:[、,，\s。；;]|$)`)
	ganZhiTokenRE           = regexp.MustCompile(ganZhiTokenPattern)
	rawBirthDatePattern     = regexp.MustCompile(`(?:(?:公历|农历)?\d{3,4}年[^，。；;\n]{0,24}(?:日|时)|(?:乾隆|嘉庆|道光|咸丰|同治|光绪|宣统|民国|康熙|雍正|成化)[^，。；;\n]{0,30}(?:日|时))`)
	mingLiCaseContextTokens = []string{"命造", "命：", "命，", "造：", "造：", "某命", "命", "位至", "官至", "贵为", "都宪", "状元", "尚书", "巡抚", "总督", "总统", "都督", "富贵", "贫贱", "大贵"}
)

type MingLiExtractOptions struct {
	SourceDir     string
	System        string
	Limit         int
	MinConfidence float64
}

type MingLiCandidateFixture struct {
	Version     string                 `json:"version"`
	GeneratedAt string                 `json:"generated_at"`
	SourceRoot  string                 `json:"source_root"`
	System      string                 `json:"system"`
	Cases       []MingLiCandidateCase  `json:"cases"`
	Warnings    []string               `json:"warnings,omitempty"`
	Stats       MingLiExtractionStats  `json:"stats"`
	Extra       map[string]interface{} `json:"-"`
}

type MingLiExtractionStats struct {
	FilesScanned int `json:"files_scanned"`
	MatchesFound int `json:"matches_found"`
	Deduped      int `json:"deduped"`
}

type MingLiCandidateCase struct {
	CaseID   string                  `json:"case_id"`
	Tier     CaseTier                `json:"tier"`
	System   string                  `json:"system"`
	Source   MingLiCandidateSource   `json:"source"`
	Input    MingLiCandidateInput    `json:"input"`
	Expected MingLiCandidateExpected `json:"expected"`
	Review   MingLiCandidateReview   `json:"review"`
	Evidence MingLiCandidateEvidence `json:"evidence"`
}

type MingLiCandidateSource struct {
	Name       string `json:"name"`
	URL        string `json:"url,omitempty"`
	License    string `json:"license"`
	SourceHash string `json:"source_hash"`
}

type MingLiCandidateInput struct {
	CalendarType  string                 `json:"calendar_type"`
	RawBirthDate  string                 `json:"raw_birth_date,omitempty"`
	TimePrecision string                 `json:"time_precision"`
	Gender        string                 `json:"gender,omitempty"`
	Options       map[string]interface{} `json:"options,omitempty"`
}

type MingLiCandidateExpected struct {
	YearPillar  string `json:"year_pillar"`
	MonthPillar string `json:"month_pillar"`
	DayPillar   string `json:"day_pillar"`
	HourPillar  string `json:"hour_pillar"`
}

type MingLiCandidateReview struct {
	Status     string   `json:"status"`
	Reviewers  []string `json:"reviewers,omitempty"`
	Confidence float64  `json:"confidence"`
	Notes      string   `json:"notes,omitempty"`
}

type MingLiCandidateEvidence struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Kind    string `json:"kind"`
	Excerpt string `json:"excerpt"`
}

type pillarMatch struct {
	year       string
	month      string
	day        string
	hour       string
	kind       string
	confidence float64
}

func ExtractMingLiCandidates(opts MingLiExtractOptions) (MingLiCandidateFixture, error) {
	sourceDir := strings.TrimSpace(opts.SourceDir)
	if sourceDir == "" {
		sourceDir = defaultMingLiSourceDir
	}
	system := strings.TrimSpace(opts.System)
	if system == "" {
		system = "bazi"
	}
	if system != "bazi" {
		return MingLiCandidateFixture{}, fmt.Errorf("mingli extractor currently supports bazi samples only, got %q", system)
	}
	if opts.MinConfidence <= 0 {
		opts.MinConfidence = 0.50
	}

	scanRoot := resolveMingLiScanRoot(sourceDir, system)
	fixture := MingLiCandidateFixture{
		Version:     "1.0",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		SourceRoot:  scanRoot,
		System:      system,
	}

	seen := make(map[string]struct{})
	err := filepath.WalkDir(scanRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			fixture.Warnings = append(fixture.Warnings, fmt.Sprintf("skip %s: %v", path, err))
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		fixture.Stats.FilesScanned++
		cases, err := extractMingLiCasesFromFile(scanRoot, path, opts.MinConfidence)
		if err != nil {
			fixture.Warnings = append(fixture.Warnings, err.Error())
			return nil
		}
		for _, candidate := range cases {
			key := strings.Join([]string{
				candidate.Source.Name,
				candidate.Expected.YearPillar,
				candidate.Expected.MonthPillar,
				candidate.Expected.DayPillar,
				candidate.Expected.HourPillar,
			}, "|")
			if _, ok := seen[key]; ok {
				fixture.Stats.Deduped++
				continue
			}
			seen[key] = struct{}{}
			fixture.Cases = append(fixture.Cases, candidate)
			fixture.Stats.MatchesFound++
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return fixture, err
	}
	sort.SliceStable(fixture.Cases, func(i, j int) bool {
		if fixture.Cases[i].Review.Confidence == fixture.Cases[j].Review.Confidence {
			if fixture.Cases[i].Source.Name == fixture.Cases[j].Source.Name {
				return fixture.Cases[i].Evidence.Line < fixture.Cases[j].Evidence.Line
			}
			return fixture.Cases[i].Source.Name < fixture.Cases[j].Source.Name
		}
		return fixture.Cases[i].Review.Confidence > fixture.Cases[j].Review.Confidence
	})
	if opts.Limit > 0 && len(fixture.Cases) > opts.Limit {
		fixture.Cases = fixture.Cases[:opts.Limit]
	}
	return fixture, nil
}

func extractMingLiCasesFromFile(root, path string, minConfidence float64) ([]MingLiCandidateCase, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		rel = path
	}
	book := mingLiBookName(rel)
	lines := strings.Split(string(data), "\n")
	var cases []MingLiCandidateCase
	for i, line := range lines {
		lineMatches := findPillarMatches(line)
		if len(lineMatches) == 0 {
			continue
		}
		window := mingLiWindow(lines, i)
		if !looksLikeMingLiCase(window) {
			continue
		}
		if isLikelyGanziEnumeration(line) {
			continue
		}
		rawBirthDate := firstRegexpMatch(rawBirthDatePattern, window)
		for _, match := range lineMatches {
			confidence := scoreMingLiCandidate(match, window, rawBirthDate, book)
			if confidence < minConfidence {
				continue
			}
			evidence := compactText(window)
			sourceHash := sha256Hex(evidence)
			caseHash := sha256Hex(fmt.Sprintf("%s:%d:%s:%s:%s:%s:%s", rel, i+1, match.kind, match.year, match.month, match.day, match.hour))
			cases = append(cases, MingLiCandidateCase{
				CaseID: "mingli-bazi-" + caseHash[:12],
				Tier:   mingLiTier(confidence),
				System: "bazi",
				Source: MingLiCandidateSource{
					Name:       book,
					License:    "local_research_review_required",
					SourceHash: "sha256:" + sourceHash,
				},
				Input: MingLiCandidateInput{
					CalendarType:  "pillars",
					RawBirthDate:  rawBirthDate,
					TimePrecision: "pillar",
					Options: map[string]interface{}{
						"source_file":       path,
						"requires_review":   true,
						"extraction_kind":   match.kind,
						"use_as_l1_fixture": match.kind == "labelled_pillars",
					},
				},
				Expected: MingLiCandidateExpected{
					YearPillar:  match.year,
					MonthPillar: match.month,
					DayPillar:   match.day,
					HourPillar:  match.hour,
				},
				Review: MingLiCandidateReview{
					Status:     "pending",
					Confidence: confidence,
					Notes:      "Auto-extracted from local mingli markdown. Human review required before promoting to gold.",
				},
				Evidence: MingLiCandidateEvidence{
					File:    path,
					Line:    i + 1,
					Kind:    match.kind,
					Excerpt: truncateRunes(evidence, 240),
				},
			})
		}
	}
	return cases, nil
}

func resolveMingLiScanRoot(sourceDir, system string) string {
	if filepath.Base(sourceDir) == system {
		return sourceDir
	}
	nested := filepath.Join(sourceDir, system)
	if info, err := os.Stat(nested); err == nil && info.IsDir() {
		return nested
	}
	return sourceDir
}

func findPillarMatches(line string) []pillarMatch {
	var matches []pillarMatch
	for _, groups := range labelledPillarsPattern.FindAllStringSubmatch(line, -1) {
		if len(groups) == 5 {
			matches = append(matches, pillarMatch{
				year:       groups[1],
				month:      groups[2],
				day:        groups[3],
				hour:       groups[4],
				kind:       "labelled_pillars",
				confidence: 0.72,
			})
		}
	}
	for _, groups := range plainPillarsPattern.FindAllStringSubmatch(line, -1) {
		if len(groups) == 5 {
			matches = append(matches, pillarMatch{
				year:       groups[1],
				month:      groups[2],
				day:        groups[3],
				hour:       groups[4],
				kind:       "plain_pillars",
				confidence: 0.55,
			})
		}
	}
	return matches
}

func scoreMingLiCandidate(match pillarMatch, context, rawBirthDate, book string) float64 {
	score := match.confidence
	if containsAny(context, []string{"命造", "命：", "命，", "命,", "某命", "人命"}) {
		score += 0.14
	} else if containsAny(context, []string{"命", "造"}) {
		score += 0.08
	}
	if containsAny(context, []string{"位至", "官至", "贵为", "状元", "尚书", "总督", "巡抚", "总统", "都督", "富贵", "贫贱"}) {
		score += 0.06
	}
	if rawBirthDate != "" {
		score += 0.06
	}
	if containsAny(book, []string{"穷通宝鉴", "三命通会", "五行精纪", "滴天髓阐微"}) {
		score += 0.04
	}
	if score > 0.98 {
		return 0.98
	}
	return roundConfidence(score)
}

func mingLiTier(confidence float64) CaseTier {
	if confidence >= 0.86 {
		return TierSilver
	}
	return TierBronze
}

func mingLiWindow(lines []string, idx int) string {
	start := idx - 1
	if start < 0 {
		start = 0
	}
	end := idx + 2
	if end > len(lines) {
		end = len(lines)
	}
	return strings.Join(lines[start:end], "\n")
}

func looksLikeMingLiCase(context string) bool {
	return containsAny(context, mingLiCaseContextTokens)
}

func isLikelyGanziEnumeration(line string) bool {
	count := len(ganZhiTokenRE.FindAllString(line, -1))
	if count > 4 {
		return true
	}
	return containsAny(line, []string{"六甲日", "六乙日", "六丙日", "六丁日", "六戊日", "六己日", "六庚日", "六辛日", "六壬日", "六癸日", "各日", "日（"})
}

func mingLiBookName(rel string) string {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) > 0 && parts[0] != "" {
		return parts[0]
	}
	return "mingli_db"
}

func firstRegexpMatch(re *regexp.Regexp, text string) string {
	match := re.FindString(text)
	return strings.TrimSpace(match)
}

func containsAny(text string, needles []string) bool {
	for _, needle := range needles {
		if strings.Contains(text, needle) {
			return true
		}
	}
	return false
}

func compactText(text string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(text)), " ")
}

func truncateRunes(text string, max int) string {
	if max <= 0 || utf8.RuneCountInString(text) <= max {
		return text
	}
	runes := []rune(text)
	return string(runes[:max]) + "..."
}

func sha256Hex(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func roundConfidence(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
