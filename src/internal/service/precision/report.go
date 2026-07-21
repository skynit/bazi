package precision

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	bazipkg "bazi/internal/service/bazi"
	fortunePkg "bazi/internal/service/fortune"
	ziweipkg "bazi/internal/service/ziwei"
)

type Options struct {
	RootDir string
}

type fieldCheck struct {
	ruleID   checkRuleID
	field    string
	required bool
	want     string
	got      string
	wantSet  []string
	gotSet   []string
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
		Version:           "4.0",
		ComparatorVersion: comparatorVersion,
		ComparatorHash:    checkRegistryHash(),
		BaselineKind:      "data_quality",
		PublicationStatus: "blocked",
		GeneratedAt:       time.Now().UTC().Format(time.RFC3339),
	}

	testdata := filepath.Join(root, "internal", "service", "testdata")
	modules := []ModuleReport{
		evaluateBaziScoped("bazi_pillar_candidates", filepath.Join(testdata, "classical_cases.json"), baziScopePillarDerived),
		evaluateBaziScoped("bazi_date_candidates", filepath.Join(testdata, "bazi_date_gold_candidates.json"), baziScopeDateChart),
		evaluateBaziScoped("bazi_extended_pillar_candidates", filepath.Join(testdata, "classical_cases_extended.json"), baziScopePillarDerived),
		evaluateZiwei(filepath.Join(testdata, "ziwei_cases.json")),
		evaluateZiweiGold(filepath.Join(testdata, "ziwei_full_chart_gold.json")),
		evaluateRikuyo(filepath.Join(testdata, "rikuyo_cases.json")),
	}
	report.Modules = modules
	for _, module := range modules {
		report.TotalCases += module.Cases
		report.EvaluatedCases += module.EvaluatedCases
		report.NonAssertiveCases += module.NonAssertiveCases
		report.SkippedCases += module.SkippedCases
		report.QuarantinedCases += module.QuarantinedCases
		report.UnsupportedChecks += module.UnsupportedChecks
		report.DiagnosticChecks += module.DiagnosticChecks
		report.DiagnosticPassed += module.DiagnosticPassed
		report.DiagnosticFailed += module.DiagnosticFailed
		report.PublishableCases += module.PublishableCases
		report.PublishableChecks += module.PublishableChecks
		report.PublishablePassed += module.PublishablePassed
		report.PublishableFailed += module.PublishableFailed
		report.Warnings = append(report.Warnings, module.Warnings...)
		if module.Name == "ziwei_full_chart_gold" && module.PublishableCases == 0 {
			report.ReleaseBlockers = append(report.ReleaseBlockers, "no frozen, independently reviewed Ziwei full-chart Gold cases are available")
		}
	}
	if report.PublishableChecks == 0 {
		report.ReleaseBlockers = append(report.ReleaseBlockers, "no reviewed Gold checks are eligible for a publishable accuracy metric")
	} else if report.PublishableFailed > 0 {
		report.BaselineKind = "gold_precision"
		report.ReleaseBlockers = append(report.ReleaseBlockers, fmt.Sprintf("%d reviewed Gold checks failed", report.PublishableFailed))
	} else {
		report.BaselineKind = "gold_precision"
		if len(report.ReleaseBlockers) == 0 {
			report.PublicationStatus = "eligible"
		}
	}
	report.External = probeExternal(root)
	return report, nil
}

type baziEvaluationScope uint8

const (
	baziScopeAuto baziEvaluationScope = iota
	baziScopeDateChart
	baziScopePillarDerived
)

func evaluateBazi(path string) ModuleReport {
	return evaluateBaziScoped("bazi", path, baziScopeAuto)
}

func evaluateBaziScoped(name, path string, scope baziEvaluationScope) ModuleReport {
	module := newModuleReport(name, path)
	module.EngineVersion = bazipkg.EngineVersion
	module.RuleVersion = bazipkg.RuleVersion
	module.RuleSchool = bazipkg.RuleSchool
	file, err := loadFixture(path)
	if err != nil {
		module.Warnings = append(module.Warnings, err.Error())
		return module
	}
	module.Cases = len(file.Cases)
	module.DuplicateCaseIDs = duplicateCaseIDs(file.Cases)
	duplicates := stringSet(module.DuplicateCaseIDs)
	svc := &bazipkg.BaziService{}
	pillarOnlyCases := 0
	patternCandidateChecks := 0
	bodyStrengthCandidateChecks := 0
	for _, tc := range file.Cases {
		metadata := effectiveMetadata(file.Metadata, tc.Metadata)
		tier, publishable := registerCase(&module, metadata, duplicates[tc.ID])
		if tc.Expected == nil {
			markNonAssertive(&module, tier, "missing_expected")
			continue
		}
		var result *bazipkg.BaziResult
		var calcErr error
		hasBirthDate := tc.Year > 0 && tc.Month > 0 && tc.Day > 0
		hasPillarInput := hasExpectedPillars(tc.Expected)
		pillarOnly := scope == baziScopePillarDerived || scope == baziScopeAuto && !hasBirthDate && hasPillarInput
		switch scope {
		case baziScopeDateChart:
			module.UnsupportedChecks += countUnsupportedExpected(tc.Expected,
				"year_pillar", "month_pillar", "day_pillar", "hour_pillar", "day_master")
			if !hasBirthDate {
				markSkippedCase(&module, tier, "missing_birth_input")
				continue
			}
		case baziScopePillarDerived:
			module.UnsupportedChecks += countUnsupportedExpected(tc.Expected,
				"year_pillar", "month_pillar", "day_pillar", "hour_pillar", "day_master",
				"body_strength", "pattern", "description")
			if !hasPillarInput {
				markSkippedCase(&module, tier, "missing_pillar_input")
				continue
			}
		default:
			module.UnsupportedChecks += countUnsupportedExpected(tc.Expected,
				"year_pillar", "month_pillar", "day_pillar", "hour_pillar", "day_master",
				"body_strength", "pattern", "description")
			if !hasBirthDate && !hasPillarInput {
				markSkippedCase(&module, tier, "missing_birth_input")
				continue
			}
		}
		if !hasSupportedBaziAssertionsForScope(tc.Expected, scope, pillarOnly) {
			markNonAssertive(&module, tier, "no_supported_assertions")
			continue
		}
		if !pillarOnly {
			// Date-based fixtures must exercise the calendar engine. Expected
			// pillars are outputs to verify, never inputs to the calculation.
			result, calcErr = svc.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, nonEmpty(tc.Gender, "MALE"))
		} else {
			result, calcErr = svc.CalculateFromPillars(
				stringValue(tc.Expected["year_pillar"]),
				stringValue(tc.Expected["month_pillar"]),
				stringValue(tc.Expected["day_pillar"]),
				stringValue(tc.Expected["hour_pillar"]),
				nonEmpty(tc.Gender, "MALE"),
			)
			pillarOnlyCases++
		}
		if calcErr != nil {
			recordDiagnosticFailure(&module, tier, publishable, CheckResult{CaseID: tc.ID, Field: "calculate", Status: "failed", Note: calcErr.Error()})
			continue
		}
		markEvaluated(&module, tier)
		if scope != baziScopeDateChart && hasExpectedValue(tc.Expected["pattern"]) {
			patternCandidateChecks++
		}
		if scope != baziScopeDateChart && hasExpectedValue(tc.Expected["body_strength"]) {
			bodyStrengthCandidateChecks++
		}
		checks := []fieldCheck{}
		if scope != baziScopeDateChart {
			checks = append(checks,
				fieldCheck{ruleID: checkBaziBodyStrength, field: "body_strength_score_band_candidate", want: stringValue(tc.Expected["body_strength"]), gotSet: []string{result.BodyStrength.ScoreBandCandidate}},
				fieldCheck{ruleID: checkBaziPatternCandidate, field: "pattern_candidate", want: stringValue(tc.Expected["pattern"]), gotSet: baziPatternCandidateNames(result.PatternAnalysis)},
			)
		}
		if !pillarOnly {
			checks = append([]fieldCheck{
				{ruleID: checkBaziYearPillar, field: "year_pillar", want: stringValue(tc.Expected["year_pillar"]), got: pillar(result.YearPillar.Gan, result.YearPillar.Zhi)},
				{ruleID: checkBaziMonthPillar, field: "month_pillar", want: stringValue(tc.Expected["month_pillar"]), got: pillar(result.MonthPillar.Gan, result.MonthPillar.Zhi)},
				{ruleID: checkBaziDayPillar, field: "day_pillar", want: stringValue(tc.Expected["day_pillar"]), got: pillar(result.DayPillar.Gan, result.DayPillar.Zhi)},
				{ruleID: checkBaziHourPillar, field: "hour_pillar", want: stringValue(tc.Expected["hour_pillar"]), got: pillar(result.HourPillar.Gan, result.HourPillar.Zhi)},
				{ruleID: checkBaziDayMaster, field: "day_master", want: stringValue(tc.Expected["day_master"]), got: result.DayPillar.Gan},
			}, checks...)
		}
		counts := applyChecks(&module, tc.ID, checks)
		addTierChecks(&module, tier, counts)
		if publishable {
			addPublishableChecks(&module, tier, counts)
		}
	}
	if pillarOnlyCases > 0 {
		module.Warnings = append(module.Warnings, fmt.Sprintf("%d pillar-only cases validate derived rules only; they do not validate calendar pillar accuracy", pillarOnlyCases))
	}
	if patternCandidateChecks > 0 {
		module.Warnings = append(module.Warnings, fmt.Sprintf("%d quarantined pattern labels are compared with detector candidate membership only; they do not adjudicate a unique pattern or validate detector thresholds", patternCandidateChecks))
	}
	if bodyStrengthCandidateChecks > 0 {
		module.Warnings = append(module.Warnings, fmt.Sprintf("%d quarantined body-strength labels are compared with an unvalidated local score-band candidate only; they do not validate weights, thresholds, adjustments, or a strength conclusion", bodyStrengthCandidateChecks))
	}
	appendFixtureQualityWarnings(&module)
	module.BoundaryStatus = boundaryStatus(path)
	return module
}

func baziPatternCandidateNames(analysis bazipkg.PatternAnalysis) []string {
	names := make([]string, 0, len(analysis.Candidates))
	for _, candidate := range analysis.Candidates {
		if candidate.PatternName != "" {
			names = append(names, candidate.PatternName)
		}
	}
	return names
}

func evaluateZiwei(path string) ModuleReport {
	module := newModuleReport("ziwei", path)
	module.ProfileID = ziweipkg.DefaultProfileID
	module.EngineVersion = ziweipkg.ZiWeiEngineVersion
	module.RuleVersion = ziweipkg.ZiWeiRuleVersion
	module.RuleSchool = ziweipkg.ZiWeiRuleSchool
	file, err := loadFixture(path)
	if err != nil {
		module.Warnings = append(module.Warnings, err.Error())
		return module
	}
	module.Cases = len(file.Cases)
	module.DuplicateCaseIDs = duplicateCaseIDs(file.Cases)
	duplicates := stringSet(module.DuplicateCaseIDs)
	svc := ziweipkg.NewZiWeiService()
	for _, tc := range file.Cases {
		metadata := effectiveMetadata(file.Metadata, tc.Metadata)
		tier, publishable := registerCase(&module, metadata, duplicates[tc.ID])
		module.UnsupportedChecks += countUnsupportedExpected(tc.Expected, "pattern", "five_bureau")
		if tc.Expected == nil || !hasAnyExpected(tc.Expected, "pattern", "five_bureau") {
			markNonAssertive(&module, tier, "no_supported_assertions")
			continue
		}
		if tc.Year <= 0 || tc.Month <= 0 || tc.Day <= 0 {
			markSkippedCase(&module, tier, "missing_birth_input")
			continue
		}
		result, err := svc.CalculateChart(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, nonEmpty(tc.Gender, "MALE"))
		if err != nil {
			recordDiagnosticFailure(&module, tier, publishable, CheckResult{CaseID: tc.ID, Field: "calculate", Status: "failed", Note: err.Error()})
			continue
		}
		markEvaluated(&module, tier)
		checks := []fieldCheck{
			{ruleID: checkZiweiLegacyPattern, field: "pattern", want: stringValue(tc.Expected["pattern"]), gotSet: result.Patterns},
			{ruleID: checkZiweiLegacyFiveBureau, field: "five_bureau", want: stringValue(tc.Expected["five_bureau"]), got: result.FiveBureau},
		}
		counts := applyChecks(&module, tc.ID, checks)
		addTierChecks(&module, tier, counts)
		if publishable {
			addPublishableChecks(&module, tier, counts)
		}
	}
	appendFixtureQualityWarnings(&module)
	return module
}

func evaluateRikuyo(path string) ModuleReport {
	module := newModuleReport("rikuyo", path)
	module.EngineVersion = fortunePkg.FortuneEngineVersion
	module.RuleVersion = bazipkg.RuleVersion
	file, err := loadFixture(path)
	if err != nil {
		module.Warnings = append(module.Warnings, err.Error())
		return module
	}
	module.Cases = len(file.Cases)
	module.DuplicateCaseIDs = duplicateCaseIDs(file.Cases)
	duplicates := stringSet(module.DuplicateCaseIDs)
	baziSvc := &bazipkg.BaziService{}
	for _, tc := range file.Cases {
		metadata := effectiveMetadata(file.Metadata, tc.Metadata)
		tier, publishable := registerCase(&module, metadata, duplicates[tc.ID])
		supportedFields := []string{"twelve_stage_name", "jian_chu_name", "huang_dao_name", "month_branch", "query_branch", "query_stem"}
		if tc.Expected == nil || !hasAnyExpected(tc.Expected, supportedFields...) {
			markNonAssertive(&module, tier, "no_supported_assertions")
			module.UnsupportedChecks += countUnsupportedExpected(tc.Expected)
			continue
		}
		module.UnsupportedChecks += countUnsupportedExpected(tc.Expected, supportedFields...)
		if tc.BirthYear <= 0 || tc.BirthMonth <= 0 || tc.BirthDay <= 0 || tc.QueryDate == "" {
			markSkippedCase(&module, tier, "missing_birth_input")
			continue
		}
		bazi, err := baziSvc.Calculate(tc.BirthYear, tc.BirthMonth, tc.BirthDay, tc.BirthHour, 0, "MALE")
		if err != nil {
			recordDiagnosticFailure(&module, tier, publishable, CheckResult{CaseID: tc.ID, Field: "bazi", Status: "failed", Note: err.Error()})
			continue
		}
		queryDate, err := time.Parse("2006-01-02", tc.QueryDate)
		if err != nil {
			recordDiagnosticFailure(&module, tier, publishable, CheckResult{CaseID: tc.ID, Field: "query_date", Status: "failed", Note: err.Error()})
			continue
		}
		result := fortunePkg.CalcRikuyo(bazi, queryDate)
		markEvaluated(&module, tier)
		checks := []fieldCheck{
			{ruleID: checkRikuyoTwelveStage, field: "twelve_stage_name", want: stringValue(tc.Expected["twelve_stage_name"]), got: result.TwelveStage.Name},
			{ruleID: checkRikuyoJianChu, field: "jian_chu_name", want: stringValue(tc.Expected["jian_chu_name"]), got: result.JianChu.Name},
			{ruleID: checkRikuyoHuangDao, field: "huang_dao_name", want: stringValue(tc.Expected["huang_dao_name"]), got: result.HuangDao.Name},
			{ruleID: checkRikuyoMonthBranch, field: "month_branch", want: stringValue(tc.Expected["month_branch"]), got: result.JianChu.MonthBranch},
			{ruleID: checkRikuyoQueryBranch, field: "query_branch", want: stringValue(tc.Expected["query_branch"]), got: result.JianChu.QueryBranch},
			{ruleID: checkRikuyoQueryStem, field: "query_stem", want: stringValue(tc.Expected["query_stem"]), got: result.SeasonalState.QueryStem},
		}
		counts := applyChecks(&module, tc.ID, checks)
		addTierChecks(&module, tier, counts)
		if publishable {
			addPublishableChecks(&module, tier, counts)
		}
	}
	appendFixtureQualityWarnings(&module)
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

type checkCounts struct {
	Checks            int
	Passed            int
	Failed            int
	PublishableChecks int
	PublishablePassed int
	PublishableFailed int
}

func applyChecks(module *ModuleReport, caseID string, checks []fieldCheck) checkCounts {
	counts := checkCounts{}
	for _, check := range checks {
		if !check.required && check.want == "" && len(check.wantSet) == 0 {
			continue
		}
		rule, exists := checkRegistry[check.ruleID]
		if !exists {
			module.UnsupportedChecks++
			module.Warnings = append(module.Warnings, fmt.Sprintf("%s.%s uses unregistered check rule %q and is excluded from every denominator", caseID, check.field, check.ruleID))
			continue
		}
		matched, evaluable := compareFieldCheck(rule, check)
		if !evaluable {
			module.UnsupportedChecks++
			module.Warnings = append(module.Warnings, fmt.Sprintf("%s.%s uses non-automatic %s comparison and is excluded from every denominator", caseID, check.field, rule.Mode))
			continue
		}
		counts.Checks++
		module.DiagnosticChecks++
		if rule.Publishable {
			counts.PublishableChecks++
		}
		if matched {
			counts.Passed++
			module.DiagnosticPassed++
			if rule.Publishable {
				counts.PublishablePassed++
			}
			continue
		}
		counts.Failed++
		module.DiagnosticFailed++
		if rule.Publishable {
			counts.PublishableFailed++
		}
		module.Failures = append(module.Failures, CheckResult{
			CaseID: caseID,
			Field:  check.field,
			Want:   fieldCheckWant(check),
			Got:    fieldCheckGot(check),
			Status: "failed",
			Note:   "comparison=" + string(rule.Mode) + "; rule_id=" + string(check.ruleID),
		})
	}
	return counts
}

func fieldCheckWant(check fieldCheck) string {
	if len(check.wantSet) > 0 {
		return strings.Join(sortedStringList(check.wantSet), ",")
	}
	return check.want
}

func fieldCheckGot(check fieldCheck) string {
	if len(check.gotSet) > 0 {
		return strings.Join(sortedStringList(check.gotSet), ",")
	}
	return check.got
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

func newModuleReport(name, path string) ModuleReport {
	return ModuleReport{
		Name:          name,
		Path:          path,
		TierBreakdown: make(map[CaseTier]TierReport),
		SkipReasons:   make(map[string]int),
	}
}

func effectiveMetadata(fileMeta, caseMeta *CaseMetadata) *CaseMetadata {
	if fileMeta == nil && caseMeta == nil {
		return nil
	}
	merged := CaseMetadata{}
	if fileMeta != nil {
		merged = *fileMeta
		merged.Reviewers = append([]string(nil), fileMeta.Reviewers...)
	}
	if caseMeta == nil {
		return &merged
	}
	if caseMeta.Tier != "" {
		merged.Tier = caseMeta.Tier
	}
	if caseMeta.SourceName != "" {
		merged.SourceName = caseMeta.SourceName
	}
	if caseMeta.SourceURL != "" {
		merged.SourceURL = caseMeta.SourceURL
	}
	if caseMeta.License != "" {
		merged.License = caseMeta.License
	}
	if caseMeta.SourceHash != "" {
		merged.SourceHash = caseMeta.SourceHash
	}
	if caseMeta.Confidence != 0 {
		merged.Confidence = caseMeta.Confidence
	}
	if caseMeta.ReviewStatus != "" {
		merged.ReviewStatus = caseMeta.ReviewStatus
	}
	if caseMeta.Reviewers != nil {
		merged.Reviewers = append([]string(nil), caseMeta.Reviewers...)
	}
	if caseMeta.Purpose != "" {
		merged.Purpose = caseMeta.Purpose
	}
	if caseMeta.QuarantineReason != "" {
		merged.QuarantineReason = caseMeta.QuarantineReason
	}
	return &merged
}

func registerCase(module *ModuleReport, metadata *CaseMetadata, duplicateID bool) (CaseTier, bool) {
	tier := TierUnclassified
	if metadata != nil && metadata.Tier != "" {
		tier = metadata.Tier
	}
	stats := module.TierBreakdown[tier]
	stats.Cases++

	missing := !metadataComplete(metadata)
	if missing {
		module.MissingMetadata++
	}
	quarantined := missing || duplicateID || metadataQuarantined(metadata)
	if quarantined {
		module.QuarantinedCases++
		stats.QuarantinedCases++
	}
	module.TierBreakdown[tier] = stats
	return tier, !quarantined && metadataPublishable(metadata)
}

func metadataComplete(metadata *CaseMetadata) bool {
	return metadata != nil && metadata.Tier != "" && strings.TrimSpace(metadata.SourceName) != "" &&
		strings.TrimSpace(metadata.License) != "" && strings.TrimSpace(metadata.ReviewStatus) != "" &&
		strings.TrimSpace(metadata.Purpose) != ""
}

func metadataQuarantined(metadata *CaseMetadata) bool {
	if metadata == nil {
		return true
	}
	status := strings.ToLower(strings.TrimSpace(metadata.ReviewStatus))
	return status == "quarantined" || status == "rejected" || strings.TrimSpace(metadata.QuarantineReason) != ""
}

func metadataPublishable(metadata *CaseMetadata) bool {
	if metadata == nil || metadata.Tier != TierGold || strings.ToLower(strings.TrimSpace(metadata.ReviewStatus)) != "approved" {
		return false
	}
	license := strings.ToLower(strings.TrimSpace(metadata.License))
	if license == "" || license == "unknown" || license == "unverified" || license == "unclear" {
		return false
	}
	return strings.TrimSpace(metadata.SourceHash) != "" && len(metadata.Reviewers) >= 2
}

func markEvaluated(module *ModuleReport, tier CaseTier) {
	module.EvaluatedCases++
	stats := module.TierBreakdown[tier]
	stats.EvaluatedCases++
	module.TierBreakdown[tier] = stats
}

func markNonAssertive(module *ModuleReport, tier CaseTier, reason string) {
	module.NonAssertiveCases++
	module.SkipReasons[reason]++
	stats := module.TierBreakdown[tier]
	stats.NonAssertiveCases++
	module.TierBreakdown[tier] = stats
}

func markSkippedCase(module *ModuleReport, tier CaseTier, reason string) {
	module.SkippedCases++
	module.SkipReasons[reason]++
	stats := module.TierBreakdown[tier]
	stats.SkippedCases++
	module.TierBreakdown[tier] = stats
}

func addTierChecks(module *ModuleReport, tier CaseTier, counts checkCounts) {
	stats := module.TierBreakdown[tier]
	stats.DiagnosticChecks += counts.Checks
	stats.DiagnosticPassed += counts.Passed
	stats.DiagnosticFailed += counts.Failed
	module.TierBreakdown[tier] = stats
}

func addPublishableChecks(module *ModuleReport, tier CaseTier, counts checkCounts) {
	if counts.PublishableChecks == 0 {
		return
	}
	module.PublishableCases++
	module.PublishableChecks += counts.PublishableChecks
	module.PublishablePassed += counts.PublishablePassed
	module.PublishableFailed += counts.PublishableFailed
	stats := module.TierBreakdown[tier]
	stats.PublishableCases++
	stats.PublishableChecks += counts.PublishableChecks
	stats.PublishablePassed += counts.PublishablePassed
	stats.PublishableFailed += counts.PublishableFailed
	module.TierBreakdown[tier] = stats
}

func recordDiagnosticFailure(module *ModuleReport, tier CaseTier, publishable bool, failure CheckResult) {
	module.DiagnosticChecks++
	module.DiagnosticFailed++
	module.Failures = append(module.Failures, failure)
	stats := module.TierBreakdown[tier]
	stats.DiagnosticChecks++
	stats.DiagnosticFailed++
	if publishable {
		module.PublishableCases++
		module.PublishableChecks++
		module.PublishableFailed++
		stats.PublishableCases++
		stats.PublishableChecks++
		stats.PublishableFailed++
	}
	module.TierBreakdown[tier] = stats
}

func duplicateCaseIDs(cases []genericCase) []string {
	counts := make(map[string]int)
	for _, tc := range cases {
		if strings.TrimSpace(tc.ID) != "" {
			counts[tc.ID]++
		}
	}
	duplicates := make([]string, 0)
	for id, count := range counts {
		if count > 1 {
			duplicates = append(duplicates, id)
		}
	}
	sort.Strings(duplicates)
	return duplicates
}

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func appendFixtureQualityWarnings(module *ModuleReport) {
	if module.MissingMetadata > 0 {
		module.Warnings = append(module.Warnings, fmt.Sprintf("%d cases have incomplete fixture metadata", module.MissingMetadata))
	}
	if len(module.DuplicateCaseIDs) > 0 {
		module.Warnings = append(module.Warnings, fmt.Sprintf("duplicate case IDs: %s", strings.Join(module.DuplicateCaseIDs, ", ")))
	}
	if module.NonAssertiveCases > 0 {
		module.Warnings = append(module.Warnings, fmt.Sprintf("%d cases contain no assertions supported by this report", module.NonAssertiveCases))
	}
	if module.UnsupportedChecks > 0 {
		module.Warnings = append(module.Warnings, fmt.Sprintf("%d expected fields are not evaluated and are excluded from every denominator", module.UnsupportedChecks))
	}
	if module.QuarantinedCases > 0 {
		module.Warnings = append(module.Warnings, fmt.Sprintf("%d cases are quarantined and excluded from publishable metrics", module.QuarantinedCases))
	}
}

func hasSupportedBaziAssertionsForScope(expected map[string]interface{}, scope baziEvaluationScope, pillarOnly bool) bool {
	if scope == baziScopeDateChart {
		return hasAnyExpected(expected, "year_pillar", "month_pillar", "day_pillar", "hour_pillar", "day_master")
	}
	if pillarOnly {
		return hasAnyExpected(expected, "body_strength", "pattern")
	}
	return hasAnyExpected(expected, "year_pillar", "month_pillar", "day_pillar", "hour_pillar", "day_master", "body_strength", "pattern")
}

func hasAnyExpected(expected map[string]interface{}, fields ...string) bool {
	for _, field := range fields {
		if hasExpectedValue(expected[field]) {
			return true
		}
	}
	return false
}

func countUnsupportedExpected(expected map[string]interface{}, supported ...string) int {
	supportedSet := stringSet(supported)
	count := 0
	for field, value := range expected {
		if !supportedSet[field] && hasExpectedValue(value) {
			count++
		}
	}
	return count
}

func hasExpectedValue(value interface{}) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	case []interface{}:
		return len(typed) > 0
	case map[string]interface{}:
		return len(typed) > 0
	default:
		return true
	}
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
		return "covered_by_go_tests: src/internal/service/bazi/test/pillar_boundary_test.go, src/internal/service/bazi/dayun_test.go, src/internal/service/fortune/dayun_date_test.go"
	}
	return ""
}
