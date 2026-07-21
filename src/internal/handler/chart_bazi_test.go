package handler

import (
	"encoding/json"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/bazi"
	"github.com/6tail/tyme4go/tyme"
	"gorm.io/datatypes"
)

func TestResolveChartBaziPrefersVerifiedSnapshot(t *testing.T) {
	service := &bazi.BaziService{}
	normalized, err := bazi.NormalizeBirthInput(bazi.BirthInput{
		Year: 1990, Month: 6, Day: 15, Hour: 8, Minute: 30,
		CalendarType: model.CalendarSolar, Gender: model.GenderMale, Timezone: bazi.DefaultBirthTimezone,
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput: %v", err)
	}
	calculated, err := service.CalculateNormalizedBirth(normalized)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}

	snapshot := *calculated
	snapshot.RuleVersion = "stored-rule-v1"
	snapshot.School = "stored-school-v1"
	snapshot.RuleMeta.RuleVersion = snapshot.RuleVersion
	snapshot.RuleMeta.School = snapshot.School
	snapshot.BodyStrength.RuleVersion = snapshot.RuleVersion
	snapshot.BodyStrength.School = snapshot.School

	chart := &model.BirthChart{
		BirthYear:       2000,
		BirthMonth:      99,
		BirthDay:        99,
		BirthHour:       99,
		Gender:          model.GenderFemale,
		EngineVersion:   "stored-engine-v1",
		RuleVersion:     snapshot.RuleVersion,
		NormalizedBirth: mustJSON(t, normalized),
		BaziSnapshot:    mustJSON(t, snapshot),
	}

	resolved, err := resolveChartBazi(service, chart)
	if err != nil {
		t.Fatalf("resolveChartBazi: %v", err)
	}
	if resolved.Source != chartBaziSourceSnapshot {
		t.Fatalf("source = %q, want %q", resolved.Source, chartBaziSourceSnapshot)
	}
	if resolved.BirthYear != normalized.Year {
		t.Fatalf("birth year = %d, want normalized year %d", resolved.BirthYear, normalized.Year)
	}
	if resolved.EngineVersion != chart.EngineVersion {
		t.Fatalf("engine version = %q, want stored %q", resolved.EngineVersion, chart.EngineVersion)
	}
	if resolved.Result.RuleVersion != snapshot.RuleVersion || resolved.Result.School != snapshot.School {
		t.Fatalf("snapshot metadata was not preserved: rule=%q school=%q", resolved.Result.RuleVersion, resolved.Result.School)
	}
}

func TestCompleteBaziSnapshotRequiresMonthSeasonEvidence(t *testing.T) {
	service := &bazi.BaziService{}
	result, err := service.Calculate(1990, 6, 15, 8, 30, model.GenderMale)
	if err != nil {
		t.Fatal(err)
	}
	chart := &model.BirthChart{
		EngineVersion: bazi.EngineVersion,
		RuleVersion:   result.RuleVersion,
	}
	complete := func(candidate *bazi.BaziResult) bool {
		return isCompleteBaziSnapshot(chart, candidate, model.GenderMale)
	}
	if !complete(result) {
		t.Fatal("current result should be accepted as a complete snapshot")
	}

	legacy := *result
	legacy.MonthSeason = bazi.MonthSeasonEvidence{}
	if complete(&legacy) {
		t.Fatal("snapshot without month-season evidence must be rejected")
	}

	tampered := *result
	tampered.MonthSeason.Season = "夏"
	if result.MonthSeason.Season == "夏" {
		tampered.MonthSeason.Season = "春"
	}
	if complete(&tampered) {
		t.Fatal("snapshot with inconsistent month-season evidence must be rejected")
	}

	legacyRelations := *result
	legacyRelations.GanZhiAnalysis = bazi.GanZhiAnalysis{}
	if complete(&legacyRelations) {
		t.Fatal("snapshot without the complete canonical relation graph must be rejected")
	}

	tamperedRelations := *result
	tamperedRelations.GanZhiAnalysis.ZhiRelations = append(
		[]bazi.ZhiRelation(nil),
		result.GanZhiAnalysis.ZhiRelations...,
	)
	if len(tamperedRelations.GanZhiAnalysis.ZhiRelations) == 0 {
		t.Fatal("test chart must contain at least one branch relation")
	}
	tamperedRelations.GanZhiAnalysis.ZhiRelations[0].RuleID = "tampered"
	if complete(&tamperedRelations) {
		t.Fatal("snapshot with a tampered branch relation must be rejected")
	}

	tamperedWuXing := *result
	tamperedWuXing.MissingElements.Note = "tampered"
	if complete(&tamperedWuXing) {
		t.Fatal("snapshot with tampered five-element evidence must be rejected")
	}

	legacyNaYin := *result
	legacyNaYin.NaYin = make(map[string]bazi.NaYinInfo, len(result.NaYin))
	for key, value := range result.NaYin {
		legacyNaYin.NaYin[key] = value
	}
	yearNaYin := legacyNaYin.NaYin["year"]
	yearNaYin.RuleID = ""
	legacyNaYin.NaYin["year"] = yearNaYin
	if complete(&legacyNaYin) {
		t.Fatal("snapshot without complete na-yin evidence must be rejected")
	}

	legacyTenGod := *result
	legacyTenGodAnalysis := *result.TenGodAnalysis
	legacyTenGodAnalysis.RuleID = ""
	legacyTenGod.TenGodAnalysis = &legacyTenGodAnalysis
	if complete(&legacyTenGod) {
		t.Fatal("snapshot without complete ten-god evidence must be rejected")
	}

	legacyTiaohou := *result
	legacyTiaohouEvidence := *result.Tiaohou
	legacyTiaohouEvidence.RuleID = ""
	legacyTiaohou.Tiaohou = &legacyTiaohouEvidence
	if complete(&legacyTiaohou) {
		t.Fatal("snapshot without complete Tiaohou evidence must be rejected")
	}

	tamperedBodyStrength := *result
	tamperedBodyStrength.BodyStrength.TotalScore += 0.01
	if complete(&tamperedBodyStrength) {
		t.Fatal("snapshot with tampered body-strength evidence must be rejected")
	}

	tamperedPattern := *result
	tamperedPattern.PatternAnalysis.Candidates = append([]bazi.PatternCandidate(nil), result.PatternAnalysis.Candidates...)
	tamperedPattern.PatternAnalysis.Candidates[0].PatternName = "篡改格"
	if complete(&tamperedPattern) {
		t.Fatal("snapshot with tampered pattern candidate must be rejected")
	}

	misorderedShenSha := *result
	misorderedShenSha.ShenShaByPillar = append([]bazi.PillarShenSha(nil), result.ShenShaByPillar...)
	misorderedShenSha.ShenShaByPillar[0], misorderedShenSha.ShenShaByPillar[2] =
		misorderedShenSha.ShenShaByPillar[2], misorderedShenSha.ShenShaByPillar[0]
	if complete(&misorderedShenSha) {
		t.Fatal("snapshot with legacy shen-sha pillar ordering must be rejected")
	}

	legacyShenSha := *result
	legacyShenSha.ShenShaByPillar = append([]bazi.PillarShenSha(nil), result.ShenShaByPillar...)
	tamperedDetail := false
	for i := range legacyShenSha.ShenShaByPillar {
		if len(legacyShenSha.ShenShaByPillar[i].Details) == 0 {
			continue
		}
		legacyShenSha.ShenShaByPillar[i].Details = append([]bazi.ShenShaMeta(nil), legacyShenSha.ShenShaByPillar[i].Details...)
		legacyShenSha.ShenShaByPillar[i].Details[0].RuleID = ""
		tamperedDetail = true
		break
	}
	if !tamperedDetail {
		t.Fatal("test fixture produced no shen-sha detail to tamper")
	}
	if complete(&legacyShenSha) {
		t.Fatal("snapshot without complete shen-sha evidence must be rejected")
	}
}

func TestChartDetailResponseMapsMonthSeasonEvidence(t *testing.T) {
	result, err := (&bazi.BaziService{}).Calculate(1990, 6, 15, 8, 30, model.GenderMale)
	if err != nil {
		t.Fatal(err)
	}
	response := chartDetailResponseWithBazi(model.BirthChart{}, result)
	var got bazi.MonthSeasonEvidence
	if err := json.Unmarshal(response.MonthSeason, &got); err != nil {
		t.Fatalf("decode month season response: %v", err)
	}
	if got != result.MonthSeason {
		t.Fatalf("month season response = %+v, want %+v", got, result.MonthSeason)
	}
	var tenGod bazi.TenGodAnalysis
	if err := json.Unmarshal(response.TenGodAnalysis, &tenGod); err != nil {
		t.Fatalf("decode ten-god response: %v", err)
	}
	if !bazi.ValidTenGodAnalysis(&tenGod, result.TenGodProportion) {
		t.Fatalf("ten-god response is not auditable: %+v", tenGod)
	}
	var tiaohou bazi.TiaohouResult
	if err := json.Unmarshal(response.Tiaohou, &tiaohou); err != nil {
		t.Fatalf("decode Tiaohou response: %v", err)
	}
	if !bazi.ValidTiaohouEvidence(&tiaohou, result.DayPillar.Gan, result.MonthPillar.Zhi) {
		t.Fatalf("Tiaohou response is not auditable: %+v", tiaohou)
	}
	var bodyStrength bazi.BodyStrengthResult
	if err := json.Unmarshal(response.BodyStrength, &bodyStrength); err != nil {
		t.Fatalf("decode body-strength response: %v", err)
	}
	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}
	if !bazi.ValidBodyStrengthEvidence(bodyStrength, pillars) {
		t.Fatalf("body-strength response is not auditable: %+v", bodyStrength)
	}
	var pattern bazi.PatternAnalysis
	if err := json.Unmarshal(response.PatternAnalysis, &pattern); err != nil {
		t.Fatalf("decode pattern response: %v", err)
	}
	if !bazi.ValidPatternAnalysis(
		pattern,
		pillars,
		result.MonthPillar.Zhi,
	) {
		t.Fatalf("pattern response is not auditable: %+v", pattern)
	}
	var patternTopLevel map[string]json.RawMessage
	if err := json.Unmarshal(response.PatternAnalysis, &patternTopLevel); err != nil {
		t.Fatalf("decode top-level pattern response: %v", err)
	}
	for _, forbidden := range []string{"pattern_name", "pattern_type", "description", "favorable_elements", "unfavorable_elements", "sub_type"} {
		if _, ok := patternTopLevel[forbidden]; ok {
			t.Fatalf("pattern response leaked top-level legacy field %q: %s", forbidden, response.PatternAnalysis)
		}
	}
	payload, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{
		"season_text", "season_text_month", "wuxing_season_note",
		"ri_zhu_desc", "ri_zhu_poem", "ri_zhu_source", "ri_zhu_comment",
		"ri_zhu_hour_detail", "jia_zi_detail", "image_desc", "energy_stage", "modern_ext", "judgments",
		"shen_sha_summary",
		"personality", "interpersonal", "career_fortune", "emotion_relation",
		"taboo", "meaning", "advice",
		"tiao_hou", "primary_god", "depth_affects_primary", "depth_hint",
		"verdict", "like", "dislike",
		"wuxing_flow", "flow_pattern_desc", "dayun_flow", "flow_change", "clash_harmony",
	} {
		if strings.Contains(string(payload), forbidden) {
			t.Fatalf("chart detail response leaked legacy field %q: %s", forbidden, payload)
		}
	}
}

func TestResolveChartBaziRejectsSnapshotWithDifferentPillars(t *testing.T) {
	service := &bazi.BaziService{}
	normalized := bazi.NormalizedBirth{
		Year: 1990, Month: 6, Day: 15, Hour: 8, Minute: 30, Gender: model.GenderMale,
	}
	calculated, err := service.Calculate(
		normalized.Year,
		normalized.Month,
		normalized.Day,
		normalized.Hour,
		normalized.Minute,
		normalized.Gender,
	)
	if err != nil {
		t.Fatalf("Calculate: %v", err)
	}

	snapshot := *calculated
	snapshot.RuleVersion = "untrusted-rule"
	snapshot.RuleMeta.RuleVersion = snapshot.RuleVersion
	snapshot.BodyStrength.RuleVersion = snapshot.RuleVersion
	if snapshot.DayPillar.Gan == "甲" {
		snapshot.DayPillar.Gan = "乙"
	} else {
		snapshot.DayPillar.Gan = "甲"
	}

	chart := &model.BirthChart{
		EngineVersion:   "stored-engine-v1",
		RuleVersion:     snapshot.RuleVersion,
		NormalizedBirth: mustJSON(t, normalized),
		BaziSnapshot:    mustJSON(t, snapshot),
	}
	resolved, err := resolveChartBazi(service, chart)
	if err != nil {
		t.Fatalf("resolveChartBazi: %v", err)
	}
	if resolved.Source != chartBaziSourceNormalized {
		t.Fatalf("source = %q, want %q", resolved.Source, chartBaziSourceNormalized)
	}
	if resolved.Result.RuleVersion != bazi.RuleVersion {
		t.Fatalf("rule version = %q, want recalculated %q", resolved.Result.RuleVersion, bazi.RuleVersion)
	}
	if !sameFourPillars(resolved.Result, calculated) {
		t.Fatalf("expected recalculated pillars after rejecting inconsistent snapshot")
	}
	if resolved.EngineVersion != bazi.EngineVersion {
		t.Fatalf("engine version = %q, want current %q", resolved.EngineVersion, bazi.EngineVersion)
	}
}

func TestResolveChartBaziUsesNormalizedBirthBeforeRawFields(t *testing.T) {
	service := &bazi.BaziService{}
	normalized, err := bazi.NormalizeBirthInput(bazi.BirthInput{
		Year: 1990, Month: 6, Day: 15, Hour: 8, Minute: 30,
		CalendarType: model.CalendarSolar, Gender: model.GenderMale, Timezone: bazi.DefaultBirthTimezone,
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput: %v", err)
	}
	chart := &model.BirthChart{
		BirthYear:       2001,
		BirthMonth:      1,
		BirthDay:        1,
		BirthHour:       0,
		BirthMin:        0,
		CalendarType:    model.CalendarSolar,
		Gender:          model.GenderFemale,
		NormalizedBirth: mustJSON(t, normalized),
	}

	resolved, err := resolveChartBazi(service, chart)
	if err != nil {
		t.Fatalf("resolveChartBazi: %v", err)
	}
	expected, err := service.Calculate(1990, 6, 15, 8, 30, model.GenderMale)
	if err != nil {
		t.Fatalf("Calculate expected: %v", err)
	}
	if resolved.Source != chartBaziSourceNormalized || resolved.BirthYear != 1990 {
		t.Fatalf("resolution = source %q year %d, want normalized birth", resolved.Source, resolved.BirthYear)
	}
	if !sameFourPillars(resolved.Result, expected) {
		t.Fatalf("resolved pillars do not match normalized birth")
	}
}

func TestResolveChartBaziPreservesNormalizedBirthSecond(t *testing.T) {
	service := &bazi.BaziService{}
	jie, err := tyme.SolarTerm{}.FromName(2022, "惊蛰")
	if err != nil {
		t.Fatalf("create solar term: %v", err)
	}
	at := jie.GetJulianDay().GetSolarTime()
	normalized := bazi.NormalizedBirth{
		Year: at.GetYear(), Month: at.GetMonth(), Day: at.GetDay(),
		Hour: at.GetHour(), Minute: at.GetMinute(), Second: at.GetSecond(), Gender: model.GenderFemale,
	}
	chart := &model.BirthChart{NormalizedBirth: mustJSON(t, normalized)}

	resolved, err := resolveChartBazi(service, chart)
	if err != nil {
		t.Fatalf("resolveChartBazi: %v", err)
	}
	if resolved.Result.DaYunInfo.ReferenceDeltaSeconds != 0 {
		t.Fatalf("normalized second was lost: reference delta = %d, want 0", resolved.Result.DaYunInfo.ReferenceDeltaSeconds)
	}
}

func TestResolveChartBaziNormalizesRawLunarBirthAndUsesSolarYear(t *testing.T) {
	service := &bazi.BaziService{}
	chart := &model.BirthChart{
		BirthYear:    2020,
		BirthMonth:   12,
		BirthDay:     25,
		BirthHour:    8,
		BirthMin:     0,
		CalendarType: model.CalendarLunar,
		Gender:       model.GenderFemale,
		Timezone:     bazi.DefaultBirthTimezone,
	}

	resolved, err := resolveChartBazi(service, chart)
	if err != nil {
		t.Fatalf("resolveChartBazi: %v", err)
	}
	if resolved.Source != chartBaziSourceRaw {
		t.Fatalf("source = %q, want %q", resolved.Source, chartBaziSourceRaw)
	}
	if resolved.BirthYear != 2021 {
		t.Fatalf("birth year = %d, want converted solar year 2021", resolved.BirthYear)
	}

	expectedBirth, err := bazi.NormalizeBirthInput(bazi.BirthInput{
		Year:         chart.BirthYear,
		Month:        chart.BirthMonth,
		Day:          chart.BirthDay,
		Hour:         chart.BirthHour,
		Minute:       chart.BirthMin,
		CalendarType: chart.CalendarType,
		Gender:       chart.Gender,
		Timezone:     chart.Timezone,
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput expected: %v", err)
	}
	expected, err := service.Calculate(
		expectedBirth.Year,
		expectedBirth.Month,
		expectedBirth.Day,
		expectedBirth.Hour,
		expectedBirth.Minute,
		expectedBirth.Gender,
	)
	if err != nil {
		t.Fatalf("Calculate expected: %v", err)
	}
	if !sameFourPillars(resolved.Result, expected) {
		t.Fatalf("resolved pillars do not match normalized lunar birth")
	}
}

func TestResolveChartBaziRejectsUnknownRawGender(t *testing.T) {
	chart := &model.BirthChart{
		BirthYear:    1990,
		BirthMonth:   6,
		BirthDay:     15,
		BirthHour:    8,
		CalendarType: model.CalendarSolar,
		Gender:       "unknown",
	}

	if _, err := resolveChartBazi(&bazi.BaziService{}, chart); err == nil {
		t.Fatal("expected invalid raw gender to be rejected instead of silently using male")
	}
}

func TestResolveChartBaziRebuildsStoredCandidateFromRawInput(t *testing.T) {
	service := &bazi.BaziService{}
	input := bazi.BirthInput{
		Year: 2024, Month: 6, Day: 10, Hour: 0, Minute: 59, Second: 59,
		CalendarType: model.CalendarSolar, Gender: model.GenderMale,
		Timezone: bazi.DefaultBirthTimezone, TimeUncertain: true, UncertaintySeconds: 1,
	}
	set, err := bazi.CalculateBirthCandidates(service, input)
	if err != nil || len(set.Candidates) != 2 {
		t.Fatalf("CalculateBirthCandidates: %v candidates=%d", err, len(set.Candidates))
	}
	selected := set.Candidates[1]
	chart := &model.BirthChart{
		BirthYear: input.Year, BirthMonth: input.Month, BirthDay: input.Day,
		BirthHour: input.Hour, BirthMin: input.Minute, BirthSec: input.Second,
		CalendarType: input.CalendarType, Gender: input.Gender, Timezone: input.Timezone,
		TimeUncertain: true, UncertaintySeconds: 1, SelectedCandidateID: selected.CandidateID,
		NormalizedBirth: datatypes.JSON(`{"invalid":true}`),
	}

	resolved, err := resolveChartBazi(service, chart)
	if err != nil {
		t.Fatalf("resolveChartBazi: %v", err)
	}
	if resolved.Source != chartBaziSourceRaw || !sameFourPillars(resolved.Result, selected.Result) {
		t.Fatalf("resolved source=%q pillars=%+v, want selected candidate %+v", resolved.Source, resolved.Result, selected.Result)
	}

	chart.SelectedCandidateID = "tampered-candidate-id"
	if _, err := resolveChartBazi(service, chart); err == nil {
		t.Fatal("expected mismatched stored candidate ID to be rejected")
	}
}

func TestResolveChartBaziReplaysStoredZiHourPolicy(t *testing.T) {
	service := &bazi.BaziService{}
	chart := &model.BirthChart{
		BirthYear: 2024, BirthMonth: 6, BirthDay: 10, BirthHour: 23, BirthMin: 30,
		CalendarType: model.CalendarSolar, Gender: model.GenderMale,
		Timezone: bazi.DefaultBirthTimezone, ZiHourPolicy: bazi.ZiHourLateZiSameDay,
	}
	resolved, err := resolveChartBazi(service, chart)
	if err != nil {
		t.Fatalf("resolveChartBazi: %v", err)
	}
	expected, err := service.CalculateAtWithPolicy(2024, 6, 10, 23, 30, 0, model.GenderMale, bazi.ZiHourLateZiSameDay)
	if err != nil {
		t.Fatal(err)
	}
	defaultResult, err := service.CalculateAtWithPolicy(2024, 6, 10, 23, 30, 0, model.GenderMale, bazi.ZiHourLateZiNextDay)
	if err != nil {
		t.Fatal(err)
	}
	if !sameFourPillars(resolved.Result, expected) || resolved.Result.ZiHourPolicy != bazi.ZiHourLateZiSameDay {
		t.Fatalf("resolved policy chart = %+v, want %+v", resolved.Result, expected)
	}
	if resolved.Result.DayPillar == defaultResult.DayPillar {
		t.Fatal("stored same-day late-Zi policy silently fell back to default")
	}
}

func mustJSON(t *testing.T, value interface{}) datatypes.JSON {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	return datatypes.JSON(encoded)
}
