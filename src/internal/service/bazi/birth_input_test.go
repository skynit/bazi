package bazi

import (
	"math"
	"strings"
	"testing"
	"time"
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

func TestNormalizeBirthInputRequiresOffsetForRepeatedDSTTime(t *testing.T) {
	input := BirthInput{
		Year: 2024, Month: 11, Day: 3, Hour: 1, Minute: 30,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "America/New_York",
	}
	if _, err := NormalizeBirthInput(input); err == nil || !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("expected repeated clock ambiguity, got %v", err)
	}

	daylightOffset, standardOffset := -4*60*60, -5*60*60
	input.UTCOffsetSeconds = &daylightOffset
	daylight, err := NormalizeBirthInput(input)
	if err != nil {
		t.Fatalf("NormalizeBirthInput daylight occurrence: %v", err)
	}
	input.UTCOffsetSeconds = &standardOffset
	standard, err := NormalizeBirthInput(input)
	if err != nil {
		t.Fatalf("NormalizeBirthInput standard occurrence: %v", err)
	}
	if daylight.Validation.UTCDateTime != "2024-11-03T05:30:00Z" || standard.Validation.UTCDateTime != "2024-11-03T06:30:00Z" {
		t.Fatalf("repeated clock UTC values = %q and %q", daylight.Validation.UTCDateTime, standard.Validation.UTCDateTime)
	}
	if !daylight.Validation.LocalTimeAmbiguous || len(daylight.Validation.PossibleUTCOffsetSeconds) != 2 {
		t.Fatalf("missing ambiguity evidence: %+v", daylight.Validation)
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

func TestEquationOfTimeSecondsUSNORegression(t *testing.T) {
	tests := []struct {
		name string
		at   time.Time
		want int
	}{
		{name: "february minimum", at: time.Date(2024, 2, 11, 12, 0, 0, 0, time.UTC), want: -852},
		{name: "november maximum", at: time.Date(2024, 11, 3, 12, 0, 0, 0, time.UTC), want: 987},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := equationOfTimeSeconds(tt.at); got != tt.want {
				t.Fatalf("equationOfTimeSeconds(%s) = %d, want %d", tt.at.Format(time.RFC3339), got, tt.want)
			}
		})
	}
}

func TestEquationOfTimeSecondsDoesNotJumpAtRightAscensionWrap(t *testing.T) {
	before := equationOfTimeSeconds(time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC))
	after := equationOfTimeSeconds(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC))
	if math.Abs(float64(before)) > 20*60 || math.Abs(float64(after)) > 20*60 {
		t.Fatalf("equation of time jumped outside physical range: before=%d after=%d", before, after)
	}
	if delta := after - before; delta < -2 || delta > 2 {
		t.Fatalf("equation of time discontinuity at year boundary: before=%d after=%d", before, after)
	}
}

func TestNormalizeBirthInputTrueSolarEvidenceUsesZoneOffset(t *testing.T) {
	shanghaiLongitude := 120.0
	shanghai, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 2, Day: 11, Hour: 20, Minute: 0, Second: 45,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
		Longitude: &shanghaiLongitude, UseTrueSolarTime: true,
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput Shanghai: %v", err)
	}
	validation := shanghai.Validation
	if validation.TimezoneOffsetSeconds != 8*60*60 {
		t.Errorf("timezone offset = %d, want 28800", validation.TimezoneOffsetSeconds)
	}
	if validation.MeanSolarAdjustmentSeconds != 0 {
		t.Errorf("mean solar adjustment = %d, want 0 at 120E UTC+8", validation.MeanSolarAdjustmentSeconds)
	}
	if validation.TrueSolarAdjustmentSeconds != validation.EquationOfTimeSeconds {
		t.Errorf("total adjustment = %d, want EOT %d", validation.TrueSolarAdjustmentSeconds, validation.EquationOfTimeSeconds)
	}
	if shanghai.Second != 33 {
		t.Errorf("normalized second = %d, want 33 after -852 second EOT", shanghai.Second)
	}

	newYorkLongitude := -75.0
	summer, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 7, Day: 1, Hour: 12, Minute: 0,
		CalendarType: "SOLAR", Gender: "FEMALE", Timezone: "America/New_York",
		Longitude: &newYorkLongitude, UseTrueSolarTime: true,
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput New York summer: %v", err)
	}
	if summer.Validation.TimezoneOffsetSeconds != -4*60*60 {
		t.Errorf("summer timezone offset = %d, want -14400", summer.Validation.TimezoneOffsetSeconds)
	}
	if summer.Validation.MeanSolarAdjustmentSeconds != -60*60 {
		t.Errorf("summer mean solar adjustment = %d, want -3600", summer.Validation.MeanSolarAdjustmentSeconds)
	}

	winter, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 1, Day: 1, Hour: 12, Minute: 0,
		CalendarType: "SOLAR", Gender: "FEMALE", Timezone: "America/New_York",
		Longitude: &newYorkLongitude, UseTrueSolarTime: true,
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput New York winter: %v", err)
	}
	if winter.Validation.MeanSolarAdjustmentSeconds != 0 {
		t.Errorf("winter mean solar adjustment = %d, want 0", winter.Validation.MeanSolarAdjustmentSeconds)
	}
}

func TestNormalizeBirthInputPreservesSecondsWithoutTrueSolarTime(t *testing.T) {
	got, err := NormalizeBirthInput(BirthInput{
		Year: 1990, Month: 5, Day: 15, Hour: 8, Minute: 26, Second: 47,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput: %v", err)
	}
	if got.Second != 47 {
		t.Fatalf("normalized second = %d, want 47", got.Second)
	}
	if got.Validation.OriginalDateTime != "1990-05-15 08:26:47" ||
		got.Validation.CalculationDateTime != "1990-05-15 08:26:47" {
		t.Fatalf("second-level timestamps missing: %+v", got.Validation)
	}
}

func TestNormalizeBirthInputRejectsInvalidSecond(t *testing.T) {
	_, err := NormalizeBirthInput(BirthInput{
		Year: 1990, Month: 5, Day: 15, Hour: 8, Minute: 26, Second: 60,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
	})
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("expected second range error, got %v", err)
	}
}

func TestNormalizeBirthInputTrueSolarValidatedRange(t *testing.T) {
	longitude := 120.0
	for _, year := range []int{1800, 2200} {
		got, err := NormalizeBirthInput(BirthInput{
			Year: year, Month: 6, Day: 1, Hour: 12,
			CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
			Longitude: &longitude, UseTrueSolarTime: true,
		})
		if err != nil {
			t.Fatalf("NormalizeBirthInput(%d): %v", year, err)
		}
		if !got.Validation.TrueSolarWithinValidatedRange || got.Validation.TrueSolarUncertaintySeconds != 6 {
			t.Errorf("year %d validation = %+v, want validated with 6-second engineering uncertainty", year, got.Validation)
		}
		if notices := strings.Join(got.Validation.Notices, " "); !strings.Contains(notices, "USNO约1角分") ||
			!strings.Contains(notices, "UTC/UT1") || !strings.Contains(notices, "6秒") || !strings.Contains(notices, "不是严格天文误差上界") {
			t.Errorf("year %d missing true-solar uncertainty basis: %v", year, got.Validation.Notices)
		}
	}

	outside, err := NormalizeBirthInput(BirthInput{
		Year: 1799, Month: 6, Day: 1, Hour: 12,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
		Longitude: &longitude, UseTrueSolarTime: true,
	})
	if err != nil {
		t.Fatalf("NormalizeBirthInput outside range: %v", err)
	}
	if outside.Validation.TrueSolarWithinValidatedRange || outside.Validation.TrueSolarUncertaintySeconds != 0 {
		t.Errorf("outside validation = %+v, want unvalidated with unknown uncertainty", outside.Validation)
	}
	if !strings.Contains(strings.Join(outside.Validation.Notices, " "), "1800–2200") {
		t.Fatalf("missing outside-range notice: %v", outside.Validation.Notices)
	}
}
