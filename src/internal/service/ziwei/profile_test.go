package ziwei

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestResolveProfileRejectsUnknownProfile(t *testing.T) {
	profile, err := ResolveProfile("")
	if err != nil {
		t.Fatal(err)
	}
	if profile.ID != DefaultProfileID {
		t.Fatalf("default profile = %q, want %q", profile.ID, DefaultProfileID)
	}
	if profile.RuntimeRuleTablesSchema != ZiWeiRuntimeRuleTablesSchema ||
		profile.RuntimeRuleTablesHash != ZiWeiRuntimeRuleTablesSHA256 || len(profile.RuntimeRuleTablesHash) != 64 {
		t.Fatalf("runtime rule-table contract = schema:%q hash:%q", profile.RuntimeRuleTablesSchema, profile.RuntimeRuleTablesHash)
	}
	if len(profile.PluginManifest) != 0 || len(profile.PluginManifestHash) != 64 {
		t.Fatalf("default plugin contract = %+v hash=%q", profile.PluginManifest, profile.PluginManifestHash)
	}
	if len(profile.RuleSources) != 9 || profile.RuleSources[0].RuleID != SiHuaRuleID ||
		profile.RuleSources[0].Commit != SiHuaSourceCommit || profile.RuleSources[0].SHA256 != SiHuaSourceSHA256 ||
		profile.RuleSources[0].ValidationStatus != "cross_checked_not_gold" {
		t.Fatalf("default rule sources = %+v", profile.RuleSources)
	}
	if profile.RuleSources[1].RuleID != StarBrightnessRuleID || profile.RuleSources[1].Path != StarBrightnessSourcePath ||
		profile.RuleSources[1].SHA256 != StarBrightnessSHA256 {
		t.Fatalf("brightness source = %+v", profile.RuleSources[1])
	}
	if profile.RuleSources[2].RuleID != LeapMonthRuleID || profile.RuleSources[2].Path != LeapMonthSourcePath ||
		profile.RuleSources[2].SHA256 != LeapMonthSHA256 {
		t.Fatalf("leap-month source = %+v", profile.RuleSources[2])
	}
	if profile.RuleSources[3].RuleID != MonthlyStarsRuleID || profile.RuleSources[3].Path != MonthlyStarsSourcePath ||
		profile.RuleSources[3].SHA256 != MonthlyStarsSHA256 {
		t.Fatalf("monthly-stars source = %+v", profile.RuleSources[3])
	}
	if profile.RuleSources[4].RuleID != AdjectiveStarsRuleID || profile.RuleSources[4].Path != AdjectiveStarsSourcePath ||
		profile.RuleSources[4].SHA256 != AdjectiveStarsSHA256 {
		t.Fatalf("adjective-stars source = %+v", profile.RuleSources[4])
	}
	if profile.RuleSources[5].RuleID != PeriodChronologyRuleID || profile.RuleSources[5].Path != PeriodChronologySourcePath ||
		profile.RuleSources[5].SHA256 != PeriodChronologySHA256 {
		t.Fatalf("period chronology source = %+v", profile.RuleSources[5])
	}
	if profile.RuleSources[6].RuleID != TransitStarsRuleID || profile.RuleSources[6].Path != TransitStarsSourcePath ||
		profile.RuleSources[6].SHA256 != TransitStarsSHA256 {
		t.Fatalf("transit stars source = %+v", profile.RuleSources[6])
	}
	if profile.RuleSources[7].RuleID != CorePatternRuleID || profile.RuleSources[7].Repository != CorePatternSourceRepo ||
		profile.RuleSources[7].Commit != CorePatternSourceCommit || profile.RuleSources[7].Path != CorePatternSourcePath ||
		profile.RuleSources[7].SHA256 != CorePatternSourceSHA256 || profile.RuleSources[7].License != "MIT" {
		t.Fatalf("core-pattern source = %+v", profile.RuleSources[7])
	}
	if profile.RuleSources[8].RuleID != QiShaChaoDouRuleID || profile.RuleSources[8].Repository != SiHuaSourceRepo ||
		profile.RuleSources[8].Commit != SiHuaSourceCommit || profile.RuleSources[8].Path != QiShaChaoDouSourcePath ||
		profile.RuleSources[8].SHA256 != QiShaChaoDouSourceSHA256 || profile.RuleSources[8].License != "MIT" {
		t.Fatalf("qisha-chaodou source = %+v", profile.RuleSources[8])
	}

	if _, err := ResolveProfile("ziwei-unknown-v1"); err == nil || !strings.Contains(err.Error(), "unknown ziwei profile") {
		t.Fatalf("unknown profile must fail explicitly, got %v", err)
	}
}

func TestCalculateChartStampsReproducibilityMetadata(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	if chart.ProfileID != DefaultProfileID || chart.EngineVersion != ZiWeiEngineVersion ||
		chart.RuleVersion != ZiWeiRuleVersion || chart.RuleSchool != ZiWeiRuleSchool ||
		chart.RuntimeRuleTablesSchema != ZiWeiRuntimeRuleTablesSchema ||
		chart.RuntimeRuleTablesHash != ZiWeiRuntimeRuleTablesSHA256 {
		t.Fatalf("chart metadata is incomplete: %+v", chart)
	}
	if !svc.ChartMatchesProfile(chart, DefaultProfileID) {
		t.Fatal("fresh chart must match the requested profile")
	}

	data, err := json.Marshal(chart)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{`"profile_id"`, `"engine_version"`, `"rule_version"`, `"rule_school"`, `"rule_sources"`, `"runtime_rule_tables_schema"`, `"runtime_rule_tables_hash"`, `"plugin_manifest"`, `"plugin_manifest_hash"`, `"calculation_input"`, `"input_fingerprint"`, `"content_hash"`} {
		if !strings.Contains(string(data), field) {
			t.Fatalf("serialized chart missing %s: %s", field, data)
		}
	}
	if !strings.Contains(string(data), `"plugin_manifest":[]`) {
		t.Fatalf("empty plugin manifest must serialize as [], got %s", data)
	}
	if chart.CalculationInput != (ZiWeiCalculationInput{
		CalendarType: "SOLAR", Year: 2003, Month: 4, Day: 15,
		Hour: 14, Minute: 0, Gender: "男", Basis: "normalized_solar_minute",
	}) {
		t.Fatalf("calculation input = %+v", chart.CalculationInput)
	}
	if !strings.Contains(string(data), `"rule_id":"`+SiHuaRuleID+`"`) ||
		!strings.Contains(string(data), `"rule_id":"`+StarBrightnessRuleID+`"`) ||
		!strings.Contains(string(data), `"rule_id":"`+LeapMonthRuleID+`"`) ||
		!strings.Contains(string(data), `"rule_id":"`+MonthlyStarsRuleID+`"`) ||
		!strings.Contains(string(data), `"rule_id":"`+AdjectiveStarsRuleID+`"`) ||
		!strings.Contains(string(data), `"rule_id":"`+PeriodChronologyRuleID+`"`) ||
		!strings.Contains(string(data), `"rule_id":"`+TransitStarsRuleID+`"`) ||
		!strings.Contains(string(data), `"rule_id":"`+CorePatternRuleID+`"`) ||
		!strings.Contains(string(data), `"rule_id":"`+QiShaChaoDouRuleID+`"`) ||
		!strings.Contains(string(data), `"commit":"`+SiHuaSourceCommit+`"`) ||
		!strings.Contains(string(data), `"commit":"`+CorePatternSourceCommit+`"`) ||
		!strings.Contains(string(data), `"sha256":"`+SiHuaSourceSHA256+`"`) ||
		!strings.Contains(string(data), `"sha256":"`+StarBrightnessSHA256+`"`) ||
		!strings.Contains(string(data), `"sha256":"`+LeapMonthSHA256+`"`) ||
		!strings.Contains(string(data), `"sha256":"`+MonthlyStarsSHA256+`"`) ||
		!strings.Contains(string(data), `"sha256":"`+AdjectiveStarsSHA256+`"`) ||
		!strings.Contains(string(data), `"sha256":"`+PeriodChronologySHA256+`"`) ||
		!strings.Contains(string(data), `"sha256":"`+TransitStarsSHA256+`"`) ||
		!strings.Contains(string(data), `"sha256":"`+CorePatternSourceSHA256+`"`) ||
		!strings.Contains(string(data), `"sha256":"`+QiShaChaoDouSourceSHA256+`"`) {
		t.Fatalf("pinned rule sources missing from chart JSON: %s", data)
	}
	second, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	secondData, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, secondData) {
		t.Fatal("same profile and birth input did not produce byte-identical chart JSON")
	}
}

func TestChartMatchesInputProfileRejectsBirthOrContentMismatch(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	if len(chart.InputFingerprint) != 64 || len(chart.ContentHash) != 64 {
		t.Fatalf("cache contract is incomplete: input=%q content=%q", chart.InputFingerprint, chart.ContentHash)
	}
	if !svc.ChartMatchesInputProfile(chart, DefaultProfileID, 2003, 4, 15, 14, 0, "MALE") {
		t.Fatal("canonical gender aliases must produce the same input fingerprint")
	}
	if svc.ChartMatchesInputProfile(chart, DefaultProfileID, 2003, 4, 15, 14, 1, "男") {
		t.Fatal("different birth minute must invalidate the cached chart")
	}

	payload, err := json.Marshal(chart)
	if err != nil {
		t.Fatal(err)
	}
	var tampered ZiWeiChart
	if err := json.Unmarshal(payload, &tampered); err != nil {
		t.Fatal(err)
	}
	tampered.Palaces[0].Name = "篡改宫"
	if svc.ChartMatchesProfile(&tampered, DefaultProfileID) ||
		svc.ChartMatchesInputProfile(&tampered, DefaultProfileID, 2003, 4, 15, 14, 0, "男") {
		t.Fatal("tampered public chart content must invalidate profile and input cache checks")
	}

	var restored ZiWeiChart
	if err := json.Unmarshal(payload, &restored); err != nil {
		t.Fatal(err)
	}
	if err := svc.AttachBirthData(&restored, 2003, 4, 15, 14, 0, "男"); err != nil {
		t.Fatal(err)
	}
	if !svc.ChartMatchesInputProfile(&restored, DefaultProfileID, 2003, 4, 15, 14, 0, "男") {
		t.Fatal("cache birth authentication must preserve the public content hash")
	}
}

func TestChartMatchesProfileRejectsStaleCache(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}

	chart.RuleVersion = "ziwei-rules-stale"
	if svc.ChartMatchesProfile(chart, "") {
		t.Fatal("stale cached chart must not match the current default profile")
	}
	chart.RuleVersion = profileRuleVersion(t)
	chart.RuleSources[0].Commit = "stale-source-commit"
	if svc.ChartMatchesProfile(chart, "") {
		t.Fatal("stale rule source must not match the current profile")
	}
	profile, err := ResolveProfile(DefaultProfileID)
	if err != nil {
		t.Fatal(err)
	}
	chart.RuleSources = cloneRuleSources(profile.RuleSources)
	chart.RuleSources[1].SHA256 = "stale-source-sha256"
	if svc.ChartMatchesProfile(chart, "") {
		t.Fatal("stale rule source hash must not match the current profile")
	}
	chart.RuleSources = cloneRuleSources(profile.RuleSources)
	chart.PluginManifestHash = "stale-plugin-manifest"
	if svc.ChartMatchesProfile(chart, "") {
		t.Fatal("stale plugin manifest must not match the current profile")
	}
	if svc.ChartMatchesProfile(chart, "ziwei-unknown-v1") {
		t.Fatal("unknown requested profile must never match a cached chart")
	}
}

func profileRuleVersion(t *testing.T) string {
	t.Helper()
	profile, err := ResolveProfile(DefaultProfileID)
	if err != nil {
		t.Fatal(err)
	}
	return profile.RuleVersion
}

func TestCalculateChartWithProfileFailsBeforeCalculation(t *testing.T) {
	svc := NewZiWeiService()
	if _, err := svc.CalculateChartWithProfile("ziwei-unknown-v1", 0, 0, 0, 0, 0, ""); err == nil || !strings.Contains(err.Error(), "unknown ziwei profile") {
		t.Fatalf("unknown profile error = %v", err)
	}
}
