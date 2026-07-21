package bazi

import (
	"reflect"
	"testing"
)

func TestNormalizeBirthInputSeparatesLocalClockAndSolarTermInstant(t *testing.T) {
	got, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 2, Day: 4, Hour: 16,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "America/New_York",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Year != 2024 || got.Month != 2 || got.Day != 4 || got.Hour != 16 {
		t.Fatalf("local calculation clock changed: %+v", got)
	}
	wantReference := (SolarTermReference{Year: 2024, Month: 2, Day: 5, Hour: 5})
	if got.SolarTermReference != wantReference {
		t.Fatalf("solar-term reference = %+v, want %+v", got.SolarTermReference, wantReference)
	}
	validation := got.Validation
	if validation.NormalizationVersion != BirthNormalizationVersion || validation.CalendarEngineVersion != CalendarEngineVersion ||
		validation.UTCDateTime != "2024-02-04T21:00:00Z" || validation.SolarTermReferenceDateTime != "2024-02-05 05:00:00" ||
		validation.SolarTermReferenceTimezone != "UTC+08:00" || validation.CurrentSolarTerm != "立春" {
		t.Fatalf("unexpected split time-basis evidence: %+v", validation)
	}
}

func TestCalculateNormalizedBirthUsesInstantForYearMonthAndLocalClockForDayHour(t *testing.T) {
	normalized, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 2, Day: 4, Hour: 16,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "America/New_York",
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := (&BaziService{}).CalculateNormalizedBirth(normalized)
	if err != nil {
		t.Fatal(err)
	}
	if pillarName(got.YearPillar.Gan, got.YearPillar.Zhi) != "甲辰" ||
		pillarName(got.MonthPillar.Gan, got.MonthPillar.Zhi) != "丙寅" ||
		pillarName(got.DayPillar.Gan, got.DayPillar.Zhi) != "戊戌" ||
		pillarName(got.HourPillar.Gan, got.HourPillar.Zhi) != "庚申" {
		t.Fatalf("split-basis pillars = %s %s %s %s, want 甲辰 丙寅 戊戌 庚申",
			pillarName(got.YearPillar.Gan, got.YearPillar.Zhi),
			pillarName(got.MonthPillar.Gan, got.MonthPillar.Zhi),
			pillarName(got.DayPillar.Gan, got.DayPillar.Zhi),
			pillarName(got.HourPillar.Gan, got.HourPillar.Zhi))
	}
	if got.CalendarVersion != CalendarEngineVersion || got.DaYunInfo.Direction != "顺行" ||
		got.DaYunInfo.SolarTermReferenceAt != "2024-02-05T05:00:00" || got.DaYunInfo.SolarTermTimezone != "UTC+08:00" {
		t.Fatalf("calendar/dayun evidence = version %q dayun %+v", got.CalendarVersion, got.DaYunInfo)
	}

	legacyLocalClock, err := (&BaziService{}).CalculateAt(2024, 2, 4, 16, 0, 0, "MALE")
	if err != nil {
		t.Fatal(err)
	}
	if got.DayPillar != legacyLocalClock.DayPillar || got.HourPillar != legacyLocalClock.HourPillar ||
		got.YearPillar == legacyLocalClock.YearPillar || got.MonthPillar == legacyLocalClock.MonthPillar {
		t.Fatalf("year/month were not separated from local day/hour: split=%+v/%+v/%+v/%+v local=%+v/%+v/%+v/%+v",
			got.YearPillar, got.MonthPillar, got.DayPillar, got.HourPillar,
			legacyLocalClock.YearPillar, legacyLocalClock.MonthPillar, legacyLocalClock.DayPillar, legacyLocalClock.HourPillar)
	}
}

func TestTrueSolarTimeCannotMoveGlobalSolarTermBoundary(t *testing.T) {
	longitude := -75.0
	base := BirthInput{
		Year: 2024, Month: 2, Day: 4, Hour: 15, Minute: 5,
		CalendarType: "SOLAR", Gender: "FEMALE", Timezone: "America/New_York",
		Longitude: &longitude,
	}
	civil, err := NormalizeBirthInput(base)
	if err != nil {
		t.Fatal(err)
	}
	base.UseTrueSolarTime = true
	apparent, err := NormalizeBirthInput(base)
	if err != nil {
		t.Fatal(err)
	}
	civilResult, err := (&BaziService{}).CalculateNormalizedBirth(civil)
	if err != nil {
		t.Fatal(err)
	}
	apparentResult, err := (&BaziService{}).CalculateNormalizedBirth(apparent)
	if err != nil {
		t.Fatal(err)
	}
	if civil.SolarTermReference != apparent.SolarTermReference ||
		civilResult.YearPillar != apparentResult.YearPillar || civilResult.MonthPillar != apparentResult.MonthPillar {
		t.Fatalf("true-solar correction moved a global Jie boundary: civil=%+v %+v/%+v apparent=%+v %+v/%+v",
			civil.SolarTermReference, civilResult.YearPillar, civilResult.MonthPillar,
			apparent.SolarTermReference, apparentResult.YearPillar, apparentResult.MonthPillar)
	}
	if civilResult.HourPillar == apparentResult.HourPillar {
		t.Fatalf("test setup did not cross the local 15:00 hour boundary: civil=%s apparent=%s",
			civil.Validation.CalculationDateTime, apparent.Validation.CalculationDateTime)
	}
}

func TestDaYunSolarTermDeltaUsesPhysicalInstantAcrossTimezones(t *testing.T) {
	normalized, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 2, Day: 4, Hour: 16,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "America/New_York",
	})
	if err != nil {
		t.Fatal(err)
	}
	localResult, err := (&BaziService{}).CalculateNormalizedBirth(normalized)
	if err != nil {
		t.Fatal(err)
	}
	referenceResult, err := (&BaziService{}).CalculateAt(2024, 2, 5, 5, 0, 0, "MALE")
	if err != nil {
		t.Fatal(err)
	}
	if localResult.DaYunInfo.ReferenceDeltaSeconds != referenceResult.DaYunInfo.ReferenceDeltaSeconds ||
		localResult.DaYunInfo.StartAgeDetail != referenceResult.DaYunInfo.StartAgeDetail ||
		localResult.DaYunInfo.ReferenceJie == nil || referenceResult.DaYunInfo.ReferenceJie == nil ||
		localResult.DaYunInfo.ReferenceJie.Name != referenceResult.DaYunInfo.ReferenceJie.Name {
		t.Fatalf("timezone changed physical Jie distance: local=%+v reference=%+v", localResult.DaYunInfo, referenceResult.DaYunInfo)
	}
	if localResult.DaYunInfo.StartAt == referenceResult.DaYunInfo.StartAt {
		t.Fatal("local and China-standard start_at representations unexpectedly collapsed")
	}
}

func TestSameInstantPreservesSolarTermPillarsAndDaYunAcrossTimezones(t *testing.T) {
	inputs := []BirthInput{
		{Year: 2024, Month: 2, Day: 5, Hour: 5, CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai"},
		{Year: 2024, Month: 2, Day: 4, Hour: 16, CalendarType: "SOLAR", Gender: "MALE", Timezone: "America/New_York"},
		{Year: 2024, Month: 2, Day: 5, Hour: 2, Minute: 45, CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Kathmandu"},
		{Year: 2024, Month: 2, Day: 4, Hour: 11, CalendarType: "SOLAR", Gender: "MALE", Timezone: "Pacific/Honolulu"},
	}

	var baseline *BaziResult
	for _, input := range inputs {
		normalized, err := NormalizeBirthInput(input)
		if err != nil {
			t.Fatalf("normalize %s: %v", input.Timezone, err)
		}
		if normalized.Validation.UTCDateTime != "2024-02-04T21:00:00Z" ||
			normalized.SolarTermReference != (SolarTermReference{Year: 2024, Month: 2, Day: 5, Hour: 5}) {
			t.Fatalf("%s does not represent the expected instant: %+v", input.Timezone, normalized)
		}
		result, err := (&BaziService{}).CalculateNormalizedBirth(normalized)
		if err != nil {
			t.Fatalf("calculate %s: %v", input.Timezone, err)
		}
		if baseline == nil {
			baseline = result
			continue
		}
		if result.YearPillar != baseline.YearPillar || result.MonthPillar != baseline.MonthPillar ||
			result.DaYunInfo.Direction != baseline.DaYunInfo.Direction ||
			result.DaYunInfo.StartAgeDetail != baseline.DaYunInfo.StartAgeDetail ||
			result.DaYunInfo.ReferenceDeltaSeconds != baseline.DaYunInfo.ReferenceDeltaSeconds ||
			!reflect.DeepEqual(result.DaYunInfo.PreviousJie, baseline.DaYunInfo.PreviousJie) ||
			!reflect.DeepEqual(result.DaYunInfo.NextJie, baseline.DaYunInfo.NextJie) ||
			!reflect.DeepEqual(result.DaYunInfo.ReferenceJie, baseline.DaYunInfo.ReferenceJie) ||
			!reflect.DeepEqual(result.DaYunInfo.Pillars, baseline.DaYunInfo.Pillars) {
			t.Fatalf("timezone changed instant-based result: baseline=%+v/%+v/%+v %s=%+v/%+v/%+v",
				baseline.YearPillar, baseline.MonthPillar, baseline.DaYunInfo,
				input.Timezone, result.YearPillar, result.MonthPillar, result.DaYunInfo)
		}
	}

	newYork, err := NormalizeBirthInput(inputs[1])
	if err != nil {
		t.Fatal(err)
	}
	newYorkResult, err := (&BaziService{}).CalculateNormalizedBirth(newYork)
	if err != nil {
		t.Fatal(err)
	}
	if newYorkResult.DayPillar == baseline.DayPillar && newYorkResult.HourPillar == baseline.HourPillar {
		t.Fatalf("local day/hour unexpectedly collapsed to China clock: baseline=%+v/%+v NewYork=%+v/%+v",
			baseline.DayPillar, baseline.HourPillar, newYorkResult.DayPillar, newYorkResult.HourPillar)
	}
}

func TestCalculateNormalizedBirthRejectsStaleCalendarVersion(t *testing.T) {
	normalized, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 6, Day: 1, Hour: 12,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatal(err)
	}
	normalized.Validation.CalendarEngineVersion = "stale"
	if _, err := (&BaziService{}).CalculateNormalizedBirth(normalized); err == nil {
		t.Fatal("stale normalized calendar version must not be silently reused")
	}
}
