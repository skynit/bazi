package precision

import (
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/ziwei"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSplitBaziFixturesIsolateCalendarAndDerivedAssertions(t *testing.T) {
	testdata := filepath.Join("..", "testdata")
	datePath := filepath.Join(testdata, "bazi_date_gold_candidates.json")
	pillarPath := filepath.Join(testdata, "classical_cases.json")
	dateModule := evaluateBaziScoped("bazi_date_candidates", datePath, baziScopeDateChart)
	pillarModule := evaluateBaziScoped("bazi_pillar_candidates", pillarPath, baziScopePillarDerived)
	if dateModule.Cases != 32 || dateModule.EvaluatedCases != 32 || dateModule.DiagnosticChecks != 128 || dateModule.UnsupportedChecks != 0 {
		t.Fatalf("date candidate scope is not isolated: %+v", dateModule)
	}
	if pillarModule.Cases != 32 || pillarModule.EvaluatedCases != 32 || pillarModule.DiagnosticChecks != 51 || pillarModule.UnsupportedChecks != 0 {
		t.Fatalf("pillar candidate scope is not isolated: %+v", pillarModule)
	}
	for _, failure := range dateModule.Failures {
		if !containsString([]string{"year_pillar", "month_pillar", "day_pillar", "hour_pillar"}, failure.Field) {
			t.Fatalf("derived field leaked into date diagnostics: %+v", failure)
		}
	}
	for _, failure := range pillarModule.Failures {
		if !containsString([]string{"body_strength_score_band_candidate", "pattern_candidate"}, failure.Field) {
			t.Fatalf("calendar field leaked into pillar-derived diagnostics: %+v", failure)
		}
	}

	type splitCase struct {
		ID       string                 `json:"id"`
		LegacyID string                 `json:"legacy_id"`
		Year     int                    `json:"year"`
		Expected map[string]interface{} `json:"expected"`
	}
	type splitFile struct {
		Cases []splitCase `json:"cases"`
	}
	load := func(path string) splitFile {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var file splitFile
		if err := json.Unmarshal(data, &file); err != nil {
			t.Fatal(err)
		}
		return file
	}
	dateFile, pillarFile := load(datePath), load(pillarPath)
	dateByLegacy := make(map[string]splitCase, len(dateFile.Cases))
	ids := map[string]bool{}
	for _, tc := range dateFile.Cases {
		if tc.LegacyID == "" || tc.Year <= 0 || ids[tc.ID] {
			t.Fatalf("invalid date candidate identity: %+v", tc)
		}
		ids[tc.ID] = true
		dateByLegacy[tc.LegacyID] = tc
	}
	for _, tc := range pillarFile.Cases {
		dateCase, exists := dateByLegacy[tc.LegacyID]
		if !exists || tc.Year != 0 || ids[tc.ID] {
			t.Fatalf("invalid pillar candidate identity or linkage: %+v", tc)
		}
		ids[tc.ID] = true
		for _, field := range []string{"year_pillar", "month_pillar", "day_pillar", "hour_pillar"} {
			if stringValue(tc.Expected[field]) != stringValue(dateCase.Expected[field]) {
				t.Fatalf("linked pillar changed for %s.%s: date=%q pillar=%q", tc.LegacyID, field, dateCase.Expected[field], tc.Expected[field])
			}
		}
	}
	if len(dateByLegacy) != 32 || len(ids) != 64 {
		t.Fatalf("split fixture linkage is incomplete: legacy=%d ids=%d", len(dateByLegacy), len(ids))
	}
}

func TestLegacyUnsupportedLabelsArePreservedAsNonAssertiveAnnotations(t *testing.T) {
	type governedCase struct {
		ID       string                 `json:"id"`
		Expected map[string]interface{} `json:"expected"`
		Legacy   struct {
			OriginalExpected map[string]interface{} `json:"original_expected"`
		} `json:"legacy_annotations"`
	}
	type governedFile struct {
		Cases []governedCase `json:"cases"`
	}
	for _, fixture := range []struct {
		path    string
		allowed map[string]bool
	}{
		{filepath.Join("..", "testdata", "ziwei_cases.json"), map[string]bool{"pattern": true, "five_bureau": true}},
		{filepath.Join("..", "testdata", "rikuyo_cases.json"), map[string]bool{"month_branch": true, "query_stem": true}},
	} {
		data, err := os.ReadFile(fixture.path)
		if err != nil {
			t.Fatal(err)
		}
		var file governedFile
		if err := json.Unmarshal(data, &file); err != nil {
			t.Fatal(err)
		}
		if len(file.Cases) == 0 {
			t.Fatalf("governed fixture is empty: %s", fixture.path)
		}
		for _, tc := range file.Cases {
			if tc.ID == "" || len(tc.Legacy.OriginalExpected) == 0 {
				t.Fatalf("legacy expectation was not preserved in %s: %+v", fixture.path, tc)
			}
			for field := range tc.Expected {
				if !fixture.allowed[field] {
					t.Fatalf("unsupported field %s remained assertive in %s case %s", field, fixture.path, tc.ID)
				}
			}
		}
	}
}

func TestBaziPatternCandidateNamesKeepsAllDetectorLabels(t *testing.T) {
	got := baziPatternCandidateNames(bazipkg.PatternAnalysis{Candidates: []bazipkg.PatternCandidate{
		{PatternName: "正官格"},
		{PatternName: "正官佩印格"},
		{},
	}})
	if len(got) != 2 || got[0] != "正官格" || got[1] != "正官佩印格" {
		t.Fatalf("candidate names = %v", got)
	}
}

func TestEvaluateBaziPillarOnlyDoesNotSelfValidatePillars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "pillar-only.json")
	fixture := `{
  "version":"1",
  "cases":[{
    "id":"pillar-only",
    "gender":"MALE",
    "expected":{
      "year_pillar":"庚午",
      "month_pillar":"壬午",
      "day_pillar":"癸卯",
      "hour_pillar":"丙辰",
      "day_master":"癸"
    }
  }]
}`
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	module := evaluateBazi(path)
	if module.DiagnosticChecks != 0 || module.DiagnosticPassed != 0 || module.NonAssertiveCases != 1 {
		t.Fatalf("pillar-only fixture must not count pillar/day-master self-checks: %+v", module)
	}
	if module.SkipReasons["no_supported_assertions"] != 1 {
		t.Fatalf("expected explicit non-assertive reason, got %+v", module.SkipReasons)
	}
}

func TestEvaluateBaziDateFixtureChecksCalculatedPillars(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "date.json")
	fixture := `{
  "version":"1",
  "cases":[{
    "id":"date-case",
    "year":1990,"month":6,"day":15,"hour":8,"minute":0,"gender":"MALE",
    "expected":{"year_pillar":"错误"}
  }]
}`
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}
	module := evaluateBazi(path)
	if module.DiagnosticChecks != 1 || module.DiagnosticFailed != 1 {
		t.Fatalf("date fixture should verify calculated pillar, got %+v", module)
	}
	if module.PublishableChecks != 0 || module.QuarantinedCases != 1 {
		t.Fatalf("fixture without metadata must be diagnostic-only: %+v", module)
	}
}

func TestApplyChecksUsesExactComparisonForScalarFields(t *testing.T) {
	module := ModuleReport{}
	applyChecks(&module, "body-strength", []fieldCheck{
		{ruleID: checkBaziYearPillar, field: "year_pillar_a", want: "身旺", got: "偏旺"},
		{ruleID: checkBaziYearPillar, field: "year_pillar_b", want: "正官格", got: "正官格（近似）"},
		{ruleID: checkBaziYearPillar, field: "year_pillar_c", want: "庚午", got: "庚午"},
	})

	if module.DiagnosticChecks != 3 || module.DiagnosticPassed != 1 || module.DiagnosticFailed != 2 {
		t.Fatalf("exact comparison counts = checks:%d passed:%d failed:%d", module.DiagnosticChecks, module.DiagnosticPassed, module.DiagnosticFailed)
	}
}

func TestApplyChecksMatchesCompleteSetMembersOnly(t *testing.T) {
	module := ModuleReport{}
	applyChecks(&module, "ziwei-pattern", []fieldCheck{
		{ruleID: checkZiweiLegacyPattern, field: "pattern", want: "杀破狼格", gotSet: []string{"紫府同宫", "杀破狼格"}},
		{ruleID: checkZiweiLegacyPattern, field: "pattern", want: "紫府同宫", gotSet: []string{"紫府同宫格"}},
	})

	if module.DiagnosticChecks != 2 || module.DiagnosticPassed != 1 || module.DiagnosticFailed != 1 {
		t.Fatalf("set comparison counts = checks:%d passed:%d failed:%d", module.DiagnosticChecks, module.DiagnosticPassed, module.DiagnosticFailed)
	}
	if len(module.Failures) != 1 || module.Failures[0].Got != "紫府同宫格" {
		t.Fatalf("unexpected set comparison failure: %+v", module.Failures)
	}
}

func TestApplyChecksExcludesUnregisteredRulesFromDenominators(t *testing.T) {
	module := ModuleReport{}
	counts := applyChecks(&module, "unknown", []fieldCheck{
		{ruleID: "unregistered.rule", field: "field", want: "甲", got: "甲"},
	})
	if counts.Checks != 0 || module.DiagnosticChecks != 0 || module.UnsupportedChecks != 1 {
		t.Fatalf("unregistered rule entered a denominator: counts=%+v module=%+v", counts, module)
	}
	if len(module.Warnings) != 1 {
		t.Fatalf("unregistered rule lacks explicit evidence: %+v", module.Warnings)
	}
}

func TestCandidateMembershipCannotBecomePublishable(t *testing.T) {
	module := newModuleReport("test", "")
	counts := applyChecks(&module, "candidate", []fieldCheck{
		{ruleID: checkBaziPatternCandidate, field: "pattern_candidate", want: "正官格", gotSet: []string{"正官格"}},
	})
	addTierChecks(&module, TierGold, counts)
	addPublishableChecks(&module, TierGold, counts)
	if module.DiagnosticPassed != 1 || module.PublishableChecks != 0 || counts.PublishableChecks != 0 {
		t.Fatalf("candidate membership leaked into publishable metrics: counts=%+v module=%+v", counts, module)
	}
}

func TestExtendedBaziFixtureReportsNonAssertiveAndDuplicateCases(t *testing.T) {
	path := filepath.Join("..", "testdata", "classical_cases_extended.json")
	module := evaluateBazi(path)

	if module.Cases != 207 || module.EvaluatedCases != 1 || module.NonAssertiveCases != 206 {
		t.Fatalf("extended fixture case classification is wrong: %+v", module)
	}
	if module.DiagnosticChecks != 1 || module.PublishableChecks != 0 || module.QuarantinedCases != 207 {
		t.Fatalf("extended fixture must remain diagnostic-only: %+v", module)
	}
	if len(module.DuplicateCaseIDs) != 0 {
		t.Fatalf("extended fixture IDs are not unique: %v", module.DuplicateCaseIDs)
	}
	if module.MissingMetadata != 0 || module.TierBreakdown[TierBronze].Cases != 207 {
		t.Fatalf("file metadata must classify every case as bronze: %+v", module)
	}
}

func TestReviewedGoldFixtureIsolatedAsPublishable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "gold.json")
	fixture := `{
  "version":"1",
  "metadata":{
    "tier":"gold",
    "source_name":"adjudicated source",
    "license":"CC-BY-4.0",
    "source_hash":"sha256:test",
    "confidence":1,
    "review_status":"approved",
    "reviewers":["reviewer-a","reviewer-b"],
    "purpose":"calendar_chart"
  },
  "cases":[{
    "id":"gold-date",
    "year":1990,"month":6,"day":15,"hour":8,"minute":0,"gender":"MALE",
    "expected":{"year_pillar":"错误"}
  }]
}`
	if err := os.WriteFile(path, []byte(fixture), 0o600); err != nil {
		t.Fatal(err)
	}

	module := evaluateBazi(path)
	if module.PublishableCases != 1 || module.PublishableChecks != 1 || module.PublishableFailed != 1 {
		t.Fatalf("reviewed Gold check must be isolated in publishable metrics: %+v", module)
	}
	if module.QuarantinedCases != 0 || module.TierBreakdown[TierGold].PublishableChecks != 1 {
		t.Fatalf("unexpected Gold classification: %+v", module)
	}
}

func TestBuildReportBlocksAccuracyWithoutReviewedGold(t *testing.T) {
	report, err := BuildReport(Options{RootDir: filepath.Join("..", "..", "..")})
	if err != nil {
		t.Fatal(err)
	}

	if report.BaselineKind != "data_quality" || report.PublicationStatus != "blocked" {
		t.Fatalf("legacy fixtures must produce a blocked data-quality report: %+v", report)
	}
	if report.Version != "4.0" || report.ComparatorVersion != comparatorVersion || report.ComparatorHash != checkRegistryHash() {
		t.Fatalf("precision comparator identity is not reproducible: %+v", report)
	}
	if report.TotalCases != 298 || report.DiagnosticChecks != 200 || report.PublishableChecks != 0 {
		t.Fatalf("unexpected report denominators: %+v", report)
	}
	if report.NonAssertiveCases != 214 || report.QuarantinedCases != 298 || report.UnsupportedChecks != 0 || len(report.ReleaseBlockers) == 0 {
		t.Fatalf("expected explicit fixture governance blockers: %+v", report)
	}
	ziweiModule := report.Modules[3]
	if ziweiModule.ProfileID != ziwei.DefaultProfileID || ziweiModule.EngineVersion != ziwei.ZiWeiEngineVersion || ziweiModule.RuleVersion != ziwei.ZiWeiRuleVersion {
		t.Fatalf("precision report must pin the Ziwei calculation profile: %+v", ziweiModule)
	}
	ziweiGoldModule := report.Modules[4]
	if ziweiGoldModule.Name != "ziwei_full_chart_gold" || ziweiGoldModule.Cases != 0 || ziweiGoldModule.BoundaryStatus != "collecting" {
		t.Fatalf("empty Ziwei Gold registry must remain explicit: %+v", ziweiGoldModule)
	}
	if !containsString(report.ReleaseBlockers, "no frozen, independently reviewed Ziwei full-chart Gold cases are available") {
		t.Fatalf("missing Ziwei Gold release blocker: %+v", report.ReleaseBlockers)
	}
	rikuyoModule := report.Modules[5]
	if rikuyoModule.Name != "rikuyo" || rikuyoModule.EvaluatedCases != 2 || rikuyoModule.NonAssertiveCases != 3 ||
		rikuyoModule.UnsupportedChecks != 0 || rikuyoModule.DiagnosticChecks != 2 {
		t.Fatalf("legacy Rikuyo interpretations must stay outside diagnostic denominators: %+v", rikuyoModule)
	}
}
