package bazi

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestMonthSeasonChangesAtExactLiChunSecond(t *testing.T) {
	service := &BaziService{}
	before, err := service.CalculateAtWithPolicy(2024, 2, 4, 16, 27, 6, "FEMALE", DefaultZiHourPolicy)
	if err != nil {
		t.Fatal(err)
	}
	at, err := service.CalculateAtWithPolicy(2024, 2, 4, 16, 27, 7, "FEMALE", DefaultZiHourPolicy)
	if err != nil {
		t.Fatal(err)
	}

	wantBefore := observeMonthSeason("丑")
	wantAt := observeMonthSeason("寅")
	if before.MonthPillar.Zhi != "丑" || before.MonthSeason != wantBefore {
		t.Fatalf("before LiChun month evidence = pillar=%+v evidence=%+v, want 丑/%+v", before.MonthPillar, before.MonthSeason, wantBefore)
	}
	if at.MonthPillar.Zhi != "寅" || at.MonthSeason != wantAt {
		t.Fatalf("at LiChun month evidence = pillar=%+v evidence=%+v, want 寅/%+v", at.MonthPillar, at.MonthSeason, wantAt)
	}
	if before.MonthSeason.Season != "冬" || before.MonthSeason.TraditionalMonth != 12 {
		t.Fatalf("before LiChun classification = %+v, want 冬/12", before.MonthSeason)
	}
	if at.MonthSeason.Season != "春" || at.MonthSeason.TraditionalMonth != 1 {
		t.Fatalf("at LiChun classification = %+v, want 春/1", at.MonthSeason)
	}
}

func TestCalculateFromPillarsUsesMonthBranchForSeason(t *testing.T) {
	tests := []struct {
		monthPillar      string
		traditionalMonth int
		season           string
	}{
		{monthPillar: "丙寅", traditionalMonth: 1, season: "春"},
		{monthPillar: "己巳", traditionalMonth: 4, season: "夏"},
		{monthPillar: "壬申", traditionalMonth: 7, season: "秋"},
		{monthPillar: "乙亥", traditionalMonth: 10, season: "冬"},
	}
	service := &BaziService{}
	for _, tc := range tests {
		t.Run(tc.monthPillar, func(t *testing.T) {
			result, err := service.CalculateFromPillars("甲子", tc.monthPillar, "戊辰", "庚申", "MALE")
			if err != nil {
				t.Fatal(err)
			}
			if result.MonthSeason.TraditionalMonth != tc.traditionalMonth || result.MonthSeason.Season != tc.season {
				t.Fatalf("month season = %+v, want month=%d season=%s", result.MonthSeason, tc.traditionalMonth, tc.season)
			}
			if !ValidMonthSeasonEvidence(result.MonthSeason, result.MonthPillar.Zhi) {
				t.Fatalf("invalid month evidence: %+v", result.MonthSeason)
			}
		})
	}
}

func TestDateAndPillarPathsReturnSameMonthSeasonEvidence(t *testing.T) {
	service := &BaziService{}
	dateResult, err := service.CalculateAtWithPolicy(2024, 2, 4, 16, 27, 7, "FEMALE", DefaultZiHourPolicy)
	if err != nil {
		t.Fatal(err)
	}
	pillarResult, err := service.CalculateFromPillars(
		dateResult.YearPillar.Gan+dateResult.YearPillar.Zhi,
		dateResult.MonthPillar.Gan+dateResult.MonthPillar.Zhi,
		dateResult.DayPillar.Gan+dateResult.DayPillar.Zhi,
		dateResult.HourPillar.Gan+dateResult.HourPillar.Zhi,
		"FEMALE",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(dateResult.MonthSeason, pillarResult.MonthSeason) {
		t.Fatalf("date evidence = %+v, pillar evidence = %+v", dateResult.MonthSeason, pillarResult.MonthSeason)
	}
}

func TestMonthSeasonJSONOmitsLegacyInterpretationFields(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("甲子", "丙寅", "戊辰", "庚申", "MALE")
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, forbidden := range []string{`"season_text"`, `"season_text_month"`, `"wuxing_season_note"`} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("legacy interpretation field %s leaked into chart JSON: %s", forbidden, payload)
		}
	}
	if !strings.Contains(text, `"month_season"`) {
		t.Fatalf("month season evidence missing from chart JSON: %s", payload)
	}
}
