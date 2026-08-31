package precision

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	ziweipkg "bazi/internal/service/ziwei"
)

func TestZiweiGoldStarGroupsDerivesLegacyViewsFromStars(t *testing.T) {
	palace := ziweipkg.PalaceInfo{Stars: []ziweipkg.StarOutput{
		{Name: "紫微", Type: "major"},
		{Name: "左辅", Type: "soft"},
		{Name: "天马", Type: "tianma"},
	}}

	mainStars, auxStars := ziweiGoldStarGroups(palace)
	if !reflect.DeepEqual(mainStars, []string{"紫微"}) {
		t.Fatalf("main stars = %v, want [紫微]", mainStars)
	}
	if !reflect.DeepEqual(auxStars, []string{"左辅", "天马"}) {
		t.Fatalf("aux stars = %v, want [左辅 天马]", auxStars)
	}
}

func TestEvaluateZiweiGoldRequiresFrozenCompleteAdjudicatedCases(t *testing.T) {
	file := ziweiGoldFile{
		Version: "test", ProfileID: ziweipkg.DefaultProfileID,
		EngineVersion: ziweipkg.ZiWeiEngineVersion, RuleVersion: ziweipkg.ZiWeiRuleVersion,
		Frozen: false,
		Cases: []ziweiGoldCase{{
			ID: "incomplete", Year: 2003, Month: 4, Day: 15, Hour: 14, Gender: "MALE",
			Metadata: approvedZiweiGoldMetadata(),
		}},
	}
	file.DatasetHash = hashZiweiGoldCases(t, file.Cases)
	path := writeZiweiGoldFile(t, file)
	module := evaluateZiweiGold(path)
	if module.Cases != 1 || module.QuarantinedCases != 1 || module.SkippedCases != 1 || module.PublishableChecks != 0 {
		t.Fatalf("incomplete Gold case was not quarantined: %+v", module)
	}
	if module.BoundaryStatus != "collecting" || module.SkipReasons["gold_admission_failed"] != 1 {
		t.Fatalf("unexpected Gold admission evidence: %+v", module)
	}
}

func TestEvaluateZiweiGoldComparesCompleteChartExactly(t *testing.T) {
	service := ziweipkg.NewZiWeiService()
	chart, err := service.CalculateChart(2003, 4, 15, 14, 0, "MALE")
	if err != nil {
		t.Fatal(err)
	}
	goldCase := ziweiGoldCase{
		ID: "adjudicated-test", Name: "evaluator fixture",
		Year: 2003, Month: 4, Day: 15, Hour: 14, Gender: "MALE",
		Metadata: approvedZiweiGoldMetadata(),
		Expected: ziweiGoldExpectedFromChart(chart, service.CalculateDayun(chart)),
	}
	file := ziweiGoldFile{
		Version: "test", ProfileID: ziweipkg.DefaultProfileID,
		EngineVersion: ziweipkg.ZiWeiEngineVersion, RuleVersion: ziweipkg.ZiWeiRuleVersion,
		Frozen: true, Cases: []ziweiGoldCase{goldCase},
	}
	file.DatasetHash = hashZiweiGoldCases(t, file.Cases)
	module := evaluateZiweiGold(writeZiweiGoldFile(t, file))
	if module.PublishableCases != 1 || module.PublishableChecks != 127 || module.PublishableFailed != 0 || module.QuarantinedCases != 0 {
		t.Fatalf("complete Gold evaluation = %+v", module)
	}

	file.Cases[0].Expected.FiveBureau = "错误五行局"
	file.DatasetHash = hashZiweiGoldCases(t, file.Cases)
	failed := evaluateZiweiGold(writeZiweiGoldFile(t, file))
	if failed.PublishableChecks != 127 || failed.PublishableFailed != 1 {
		t.Fatalf("exact field mismatch was not isolated: %+v", failed)
	}
	if len(failed.Failures) != 1 || failed.Failures[0].Field != "five_bureau" {
		t.Fatalf("unexpected failure evidence: %+v", failed.Failures)
	}
}

func TestLoadZiweiGoldFileRejectsTrailingJSON(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ziwei_full_chart_gold.json")
	if err := os.WriteFile(path, []byte(`{"version":"test","cases":[]} {}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadZiweiGoldFile(path); err == nil || !strings.Contains(err.Error(), "trailing content") {
		t.Fatalf("trailing JSON must be rejected, got %v", err)
	}
}

func TestValidateZiweiGoldExpectedRejectsPartialFullChart(t *testing.T) {
	service := ziweipkg.NewZiWeiService()
	chart, err := service.CalculateChart(2003, 4, 15, 14, 0, "MALE")
	if err != nil {
		t.Fatal(err)
	}
	expected := ziweiGoldExpectedFromChart(chart, service.CalculateDayun(chart))

	expected.Dayun = expected.Dayun[:1]
	expected.Palaces[0].IsBodyPalace = false
	expected.Palaces[0].FourHua = nil
	reasons := strings.Join(validateZiweiGoldExpected(expected), "; ")
	for _, want := range []string{"十二个十年阶段", "explicit star/four-hua arrays"} {
		if !strings.Contains(reasons, want) {
			t.Fatalf("missing admission reason %q in %q", want, reasons)
		}
	}
}

func approvedZiweiGoldMetadata() *CaseMetadata {
	return &CaseMetadata{
		Tier: TierGold, SourceName: "independently adjudicated chart",
		SourceURL: "https://example.invalid/source-page", License: "CC-BY-4.0",
		SourceHash: "sha256:" + strings.Repeat("a", 64), Confidence: 1,
		ReviewStatus: "approved", Reviewers: []string{"reviewer-a", "reviewer-b"},
		Purpose: ziweiGoldPurpose,
	}
}

func ziweiGoldExpectedFromChart(chart *ziweipkg.ZiWeiChart, dayun ziweipkg.Dayun) ziweiGoldExpected {
	expected := ziweiGoldExpected{
		SoulPalaceBranch: chart.EarthlyBranchOfSoulPalace,
		BodyPalaceBranch: chart.EarthlyBranchOfBodyPalace,
		BodyPalace:       chart.BodyPalace, LifeMaster: chart.LifeMaster,
		BodyMaster: chart.BodyMaster, FiveBureau: chart.FiveBureau,
		Palaces: make([]ziweiGoldPalace, 0, 12),
		Dayun:   append([]ziweipkg.DayunStage(nil), dayun...),
	}
	for i := range expected.Dayun {
		expected.Dayun[i].Stars = append([]string{}, expected.Dayun[i].Stars...)
		expected.Dayun[i].LiuNianStars = append([]string{}, expected.Dayun[i].LiuNianStars...)
		expected.Dayun[i].LiuYueStars = append([]string{}, expected.Dayun[i].LiuYueStars...)
	}
	for _, palace := range chart.Palaces {
		mainStars, auxStars := ziweiGoldStarGroups(palace)
		expected.Palaces = append(expected.Palaces, ziweiGoldPalace{
			Name: palace.Name, Branch: palace.Branch, HeavenlyStem: palace.HeavenlyStem,
			IsBodyPalace: palace.IsBodyPalace,
			MainStars:    append([]string{}, mainStars...),
			AuxStars:     append([]string{}, auxStars...),
			FourHua:      append([]string{}, palace.FourHua...),
			Changsheng12: palace.Changsheng12, Boshi12: palace.Boshi12,
			JiangQian12: palace.JiangQian12, SuiQian12: palace.SuiQian12,
		})
	}
	return expected
}

func hashZiweiGoldCases(t *testing.T, cases []ziweiGoldCase) string {
	t.Helper()
	encoded, err := json.Marshal(cases)
	if err != nil {
		t.Fatal(err)
	}
	return "sha256:" + sha256HexBytes(encoded)
}

func writeZiweiGoldFile(t *testing.T, file ziweiGoldFile) string {
	t.Helper()
	encoded, err := json.Marshal(file)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ziwei_full_chart_gold.json")
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
