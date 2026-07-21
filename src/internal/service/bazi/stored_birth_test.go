package bazi

import (
	"testing"

	"bazi/internal/model"
)

func TestResolveStoredBirthNormalizesRawLunarDate(t *testing.T) {
	service := &BaziService{}
	chart := &model.BirthChart{
		BirthYear:    2020,
		BirthMonth:   12,
		BirthDay:     25,
		BirthHour:    8,
		CalendarType: model.CalendarLunar,
		Gender:       model.GenderFemale,
		Timezone:     DefaultBirthTimezone,
	}

	resolved, err := ResolveStoredBirth(service, chart)
	if err != nil {
		t.Fatalf("resolve stored lunar birth: %v", err)
	}
	expectedBirth, err := NormalizeBirthInput(BirthInput{
		Year: 2020, Month: 12, Day: 25, Hour: 8,
		CalendarType: model.CalendarLunar, Gender: model.GenderFemale, Timezone: DefaultBirthTimezone,
	})
	if err != nil {
		t.Fatalf("normalize expected lunar birth: %v", err)
	}
	expected, err := service.CalculateNormalizedBirth(expectedBirth)
	if err != nil {
		t.Fatalf("calculate expected lunar birth: %v", err)
	}
	rawSolar, err := service.Calculate(2020, 12, 25, 8, 0, model.GenderFemale)
	if err != nil {
		t.Fatalf("calculate same-number solar birth: %v", err)
	}

	if resolved.Source != StoredBirthSourceRaw || !sameStoredBirthPillars(resolved.Result, expected) {
		t.Fatalf("resolved lunar pillars do not match normalized birth: got=%+v want=%+v", resolved.Result, expected)
	}
	if sameStoredBirthPillars(resolved.Result, rawSolar) {
		t.Fatal("raw lunar fields were silently treated as the same-number solar date")
	}
}

func sameStoredBirthPillars(left, right *BaziResult) bool {
	return left != nil && right != nil &&
		left.YearPillar == right.YearPillar && left.MonthPillar == right.MonthPillar &&
		left.DayPillar == right.DayPillar && left.HourPillar == right.HourPillar
}
