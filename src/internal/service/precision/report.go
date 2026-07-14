package precision

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	bazipkg "bazi/internal/service/bazi"
	fortunePkg "bazi/internal/service/fortune"
	ziweipkg "bazi/internal/service/ziwei"
)

type Options struct {
	RootDir string
}

type fixtureFile struct {
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Metadata    *CaseMetadata          `json:"metadata"`
	Cases       []genericCase          `json:"cases"`
	Extra       map[string]interface{} `json:"-"`
}

type genericCase struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Source     string                 `json:"source"`
	Metadata   *CaseMetadata          `json:"metadata"`
	Year       int                    `json:"year"`
	Month      int                    `json:"month"`
	Day        int                    `json:"day"`
	Hour       int                    `json:"hour"`
	Minute     int                    `json:"minute"`
	Gender     string                 `json:"gender"`
	BirthYear  int                    `json:"birth_year"`
	BirthMonth int                    `json:"birth_month"`
	BirthDay   int                    `json:"birth_day"`
	BirthHour  int                    `json:"birth_hour"`
	QueryDate  string                 `json:"query_date"`
	Expected   map[string]interface{} `json:"expected"`
}

func BuildReport(opts Options) (Report, error) {
	root := opts.RootDir
	if root == "" {
		root = "."
	}
	report := Report{
		Version:     "1.0",
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
	}

	testdata := filepath.Join(root, "internal", "service", "testdata")
	modules := []ModuleReport{
		evaluateBazi(filepath.Join(testdata, "classical_cases.json")),
		evaluateBazi(filepath.Join(testdata, "classical_cases_extended.json")),
		evaluateZiwei(filepath.Join(testdata, "ziwei_cases.json")),
		evaluateRikuyo(filepath.Join(testdata, "rikuyo_cases.json")),
	}
	report.Modules = modules
	for _, module := range modules {
		report.TotalCases += module.Cases
		report.TotalChecks += module.Checks
		report.PassedChecks += module.Passed
		report.FailedChecks += module.Failed
		report.SkippedChecks += module.Skipped
		report.Warnings = append(report.Warnings, module.Warnings...)
	}
	report.External = probeExternal(root)
	return report, nil
}

func evaluateBazi(path string) ModuleReport {
	module := ModuleReport{Name: "bazi", Path: path}
	file, err := loadFixture(path)
	if err != nil {
		module.Warnings = append(module.Warnings, err.Error())
		return module
	}
	module.Cases = len(file.Cases)
	svc := &bazipkg.BaziService{}
	for _, tc := range file.Cases {
		module.MissingMetadata += missingMetadata(file.Metadata, tc.Metadata)
		if tc.Expected == nil {
			module.Skipped++
			continue
		}
		var result *bazipkg.BaziResult
		var calcErr error
		hasBirthDate := tc.Year > 0 && tc.Month > 0 && tc.Day > 0
		pillarOnly := !hasBirthDate && hasExpectedPillars(tc.Expected)
		if hasBirthDate {
			// Date-based fixtures must exercise the calendar engine. Expected
			// pillars are outputs to verify, never inputs to the calculation.
			result, calcErr = svc.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, nonEmpty(tc.Gender, "MALE"))
		} else if pillarOnly {
			result, calcErr = svc.CalculateFromPillars(
				stringValue(tc.Expected["year_pillar"]),
				stringValue(tc.Expected["month_pillar"]),
				stringValue(tc.Expected["day_pillar"]),
				stringValue(tc.Expected["hour_pillar"]),
				nonEmpty(tc.Gender, "MALE"),
			)
			module.Warnings = append(module.Warnings, fmt.Sprintf("[%s] pillar-only fixture validates derived rules only; it does not validate calendar pillar accuracy", tc.ID))
		} else {
			module.Skipped++
			continue
		}
		if calcErr != nil {
			module.Failed++
			module.Failures = append(module.Failures, CheckResult{CaseID: tc.ID, Field: "calculate", Status: "failed", Note: calcErr.Error()})
			continue
		}
		checks := []struct {
			field string
			want  string
			got   string
		}{
			{"body_strength", stringValue(tc.Expected["body_strength"]), result.BodyStrength.Verdict},
			{"pattern", stringValue(tc.Expected["pattern"]), result.PatternAnalysis.PatternName},
		}
		if !pillarOnly {
			checks = append([]struct {
				field string
				want  string
				got   string
			}{
				{"year_pillar", stringValue(tc.Expected["year_pillar"]), pillar(result.YearPillar.Gan, result.YearPillar.Zhi)},
				{"month_pillar", stringValue(tc.Expected["month_pillar"]), pillar(result.MonthPillar.Gan, result.MonthPillar.Zhi)},
				{"day_pillar", stringValue(tc.Expected["day_pillar"]), pillar(result.DayPillar.Gan, result.DayPillar.Zhi)},
				{"hour_pillar", stringValue(tc.Expected["hour_pillar"]), pillar(result.HourPillar.Gan, result.HourPillar.Zhi)},
				{"day_master", stringValue(tc.Expected["day_master"]), result.DayPillar.Gan},
			}, checks...)
		}
		applyChecks(&module, tc.ID, checks)
	}
	module.BoundaryStatus = boundaryStatus(path)
	return module
}

func evaluateZiwei(path string) ModuleReport {
	module := ModuleReport{Name: "ziwei", Path: path}
	file, err := loadFixture(path)
	if err != nil {
		module.Warnings = append(module.Warnings, err.Error())
		return module
	}
	module.Cases = len(file.Cases)
	svc := ziweipkg.NewZiWeiService()
	for _, tc := range file.Cases {
		module.MissingMetadata += missingMetadata(file.Metadata, tc.Metadata)
		if tc.Year <= 0 || tc.Month <= 0 || tc.Day <= 0 {
			module.Skipped++
			continue
		}
		result, err := svc.CalculateChart(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, nonEmpty(tc.Gender, "MALE"))
		if err != nil {
			module.Failed++
			module.Failures = append(module.Failures, CheckResult{CaseID: tc.ID, Field: "calculate", Status: "failed", Note: err.Error()})
			continue
		}
		checks := []struct {
			field string
			want  string
			got   string
		}{
			{"pattern", stringValue(tc.Expected["pattern"]), strings.Join(result.Patterns, ",")},
			{"five_bureau", stringValue(tc.Expected["five_bureau"]), result.FiveBureau},
		}
		applyChecks(&module, tc.ID, checks)
	}
	return module
}

func evaluateRikuyo(path string) ModuleReport {
	module := ModuleReport{Name: "rikuyo", Path: path}
	file, err := loadFixture(path)
	if err != nil {
		module.Warnings = append(module.Warnings, err.Error())
		return module
	}
	module.Cases = len(file.Cases)
	baziSvc := &bazipkg.BaziService{}
	for _, tc := range file.Cases {
		module.MissingMetadata += missingMetadata(file.Metadata, tc.Metadata)
		if tc.BirthYear <= 0 || tc.BirthMonth <= 0 || tc.BirthDay <= 0 || tc.QueryDate == "" {
			module.Skipped++
			continue
		}
		bazi, err := baziSvc.Calculate(tc.BirthYear, tc.BirthMonth, tc.BirthDay, tc.BirthHour, 0, "MALE")
		if err != nil {
			module.Failed++
			module.Failures = append(module.Failures, CheckResult{CaseID: tc.ID, Field: "bazi", Status: "failed", Note: err.Error()})
			continue
		}
		queryDate, err := time.Parse("2006-01-02", tc.QueryDate)
		if err != nil {
			module.Failed++
			module.Failures = append(module.Failures, CheckResult{CaseID: tc.ID, Field: "query_date", Status: "failed", Note: err.Error()})
			continue
		}
		result := fortunePkg.CalcRikuyo(bazi, queryDate, tc.BirthYear)
		if wantMonthZhi := stringValue(tc.Expected["month_zhi"]); wantMonthZhi != "" {
			module.Skipped++
			module.Warnings = append(module.Warnings, fmt.Sprintf("[%s] month_zhi=%s is fixture context; precision report does not expose calculated month_zhi yet", tc.ID, wantMonthZhi))
		}
		checks := []struct {
			field string
			want  string
			got   string
		}{
			{"jian_chu_should_be", stringValue(tc.Expected["jian_chu_should_be"]), result.JianChuName},
			{"huang_dao_or_hei_dao", stringValue(tc.Expected["huang_dao_or_hei_dao"]), huangDaoStatus(result.HuangDaoName, result.HuangDaoFavorable)},
		}
		applyChecks(&module, tc.ID, checks)
	}
	return module
}

func loadFixture(path string) (fixtureFile, error) {
	var file fixtureFile
	data, err := os.ReadFile(path)
	if err != nil {
		return file, fmt.Errorf("load %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return file, fmt.Errorf("parse %s: %w", path, err)
	}
	return file, nil
}

func applyChecks(module *ModuleReport, caseID string, checks []struct {
	field string
	want  string
	got   string
}) {
	for _, check := range checks {
		if check.want == "" {
			continue
		}
		module.Checks++
		if check.got == check.want || strings.Contains(check.got, check.want) || strings.Contains(check.want, check.got) {
			module.Passed++
			continue
		}
		module.Failed++
		module.Failures = append(module.Failures, CheckResult{
			CaseID: caseID,
			Field:  check.field,
			Want:   check.want,
			Got:    check.got,
			Status: "failed",
		})
	}
}

func probeExternal(root string) []ExternalProbe {
	probes := []ExternalProbe{}
	ziweiSamples := filepath.Join(root, "..", "data", "external", "ziwei-doushu-v3", "samples")
	if entries, err := os.ReadDir(ziweiSamples); err == nil {
		count := 0
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				count++
			}
		}
		probes = append(probes, ExternalProbe{Name: "Renhuai123/ziwei-doushu", Status: "available", Path: ziweiSamples, Note: fmt.Sprintf("%d json samples found", count)})
	} else {
		probes = append(probes, ExternalProbe{Name: "Renhuai123/ziwei-doushu", Status: "skipped", Path: ziweiSamples, Note: "sample directory not found; no large data downloaded"})
	}
	if _, err := exec.LookPath("node"); err != nil {
		probes = append(probes, ExternalProbe{Name: "mystilight-8char", Status: "skipped", Note: "node not found"})
	} else {
		probes = append(probes, ExternalProbe{Name: "mystilight-8char", Status: "skipped", Note: "optional differential runner not installed in repo"})
	}
	mingliPath := filepath.Join(root, "..", "data", "external", "MingLi-Bench")
	if _, err := os.Stat(mingliPath); err == nil {
		probes = append(probes, ExternalProbe{Name: "DestinyLinker/MingLi-Bench", Status: "available", Path: mingliPath, Note: "local checkout found; run stats without API key"})
	} else {
		probes = append(probes, ExternalProbe{Name: "DestinyLinker/MingLi-Bench", Status: "skipped", Path: mingliPath, Note: "local checkout not found"})
	}
	return probes
}

func missingMetadata(fileMeta, caseMeta *CaseMetadata) int {
	if caseMeta != nil && caseMeta.Tier != "" && caseMeta.SourceName != "" && caseMeta.License != "" && caseMeta.ReviewStatus != "" {
		return 0
	}
	if fileMeta != nil && fileMeta.Tier != "" && fileMeta.SourceName != "" && fileMeta.License != "" && fileMeta.ReviewStatus != "" {
		return 0
	}
	return 1
}

func hasExpectedPillars(expected map[string]interface{}) bool {
	return stringValue(expected["year_pillar"]) != "" &&
		stringValue(expected["month_pillar"]) != "" &&
		stringValue(expected["day_pillar"]) != "" &&
		stringValue(expected["hour_pillar"]) != ""
}

func stringValue(v interface{}) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return ""
	}
}

func nonEmpty(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func pillar(gan, zhi string) string {
	if gan == "" && zhi == "" {
		return ""
	}
	return gan + zhi
}

func boundaryStatus(path string) string {
	if strings.Contains(path, "classical_cases") {
		return "covered_by_go_tests: src/internal/service/bazi/test/pillar_boundary_test.go"
	}
	return ""
}

func huangDaoStatus(name string, favorable bool) string {
	if favorable {
		return name + " 黄道"
	}
	return name + " 黑道"
}
