package bazi

import (
	"strings"
	"testing"
)

func TestNormalizeBirthInputSolar(t *testing.T) {
	got, err := NormalizeBirthInput(BirthInput{
		Year: 1990, Month: 5, Day: 15, Hour: 8, Minute: 26,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput: %v", err)
	}
	if got.Year != 1990 || got.Month != 5 || got.Day != 15 || got.Hour != 8 || got.Minute != 26 {
		t.Fatalf("normalized time = %+v", got)
	}
	if got.Validation.LunarDate == "" || got.Validation.CurrentSolarTerm == "" {
		t.Fatalf("missing validation facts: %+v", got.Validation)
	}
}

func TestNormalizeBirthInputLunar(t *testing.T) {
	got, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 1, Day: 1, Hour: 9, Minute: 30,
		CalendarType: "LUNAR", Gender: "FEMALE", Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput: %v", err)
	}
	if got.Year != 2024 || got.Month != 2 || got.Day != 10 || got.Hour != 9 || got.Minute != 30 {
		t.Fatalf("农历 2024-01-01 conversion = %04d-%02d-%02d %02d:%02d", got.Year, got.Month, got.Day, got.Hour, got.Minute)
	}
	if got.Validation.InputCalendar != "LUNAR" {
		t.Fatalf("input calendar = %q", got.Validation.InputCalendar)
	}
}

func TestNormalizeBirthInputTrueSolarTime(t *testing.T) {
	longitude := 87.62
	got, err := NormalizeBirthInput(BirthInput{
		Year: 1990, Month: 5, Day: 15, Hour: 8, Minute: 0,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
		Longitude: &longitude, UseTrueSolarTime: true,
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput: %v", err)
	}
	if !got.Validation.TrueSolarTimeApplied {
		t.Fatal("expected true solar time to be applied")
	}
	if got.Validation.TrueSolarAdjustmentMinutes >= -60 {
		t.Fatalf("expected a material westward correction, got %d minutes", got.Validation.TrueSolarAdjustmentMinutes)
	}
}

func TestNormalizeBirthInputRejectsDSTGap(t *testing.T) {
	_, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 3, Day: 10, Hour: 2, Minute: 30,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "America/New_York",
	})
	if err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected DST gap error, got %v", err)
	}
}

func TestNormalizeBirthInputRequiresLongitudeForTrueSolarTime(t *testing.T) {
	_, err := NormalizeBirthInput(BirthInput{
		Year: 1990, Month: 5, Day: 15, Hour: 8, Minute: 0,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
		UseTrueSolarTime: true,
	})
	if err == nil {
		t.Fatal("expected longitude validation error")
	}
}
