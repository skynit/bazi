package bazi

import (
	"sync"
	"testing"

	"github.com/6tail/tyme4go/tyme"
)

func TestCalculateAtWithPolicySeparatesLateZiDayConventions(t *testing.T) {
	service := &BaziService{}
	defaultLateZi, err := service.CalculateAtWithPolicy(2024, 6, 10, 23, 30, 0, "MALE", ZiHourLateZiNextDay)
	if err != nil {
		t.Fatalf("CalculateAtWithPolicy default late Zi: %v", err)
	}
	sameDayLateZi, err := service.CalculateAtWithPolicy(2024, 6, 10, 23, 30, 0, "MALE", ZiHourLateZiSameDay)
	if err != nil {
		t.Fatalf("CalculateAtWithPolicy same-day late Zi: %v", err)
	}
	beforeZi, err := service.CalculateAtWithPolicy(2024, 6, 10, 22, 30, 0, "MALE", ZiHourLateZiSameDay)
	if err != nil {
		t.Fatal(err)
	}
	afterMidnightDefault, err := service.CalculateAtWithPolicy(2024, 6, 11, 0, 30, 0, "MALE", ZiHourLateZiNextDay)
	if err != nil {
		t.Fatal(err)
	}
	afterMidnightSameDay, err := service.CalculateAtWithPolicy(2024, 6, 11, 0, 30, 0, "MALE", ZiHourLateZiSameDay)
	if err != nil {
		t.Fatal(err)
	}

	if sameDayLateZi.DayPillar != beforeZi.DayPillar {
		t.Fatalf("same-day late Zi changed day pillar: 22:30=%+v 23:30=%+v", beforeZi.DayPillar, sameDayLateZi.DayPillar)
	}
	if defaultLateZi.DayPillar != afterMidnightDefault.DayPillar {
		t.Fatalf("next-day late Zi did not use next day pillar: 23:30=%+v 00:30=%+v", defaultLateZi.DayPillar, afterMidnightDefault.DayPillar)
	}
	if defaultLateZi.DayPillar == sameDayLateZi.DayPillar {
		t.Fatalf("policies did not separate late-Zi day pillars: next=%+v/%+v same=%+v/%+v",
			defaultLateZi.DayPillar, defaultLateZi.HourPillar, sameDayLateZi.DayPillar, sameDayLateZi.HourPillar)
	}
	if defaultLateZi.HourPillar != sameDayLateZi.HourPillar {
		t.Fatalf("tyme sect-2 convention must retain the same Zi-hour pillar: next=%+v same=%+v", defaultLateZi.HourPillar, sameDayLateZi.HourPillar)
	}
	if err := validatePillarLinkage(
		sixtyCycleForTest(t, sameDayLateZi.YearPillar.Gan+sameDayLateZi.YearPillar.Zhi),
		sixtyCycleForTest(t, sameDayLateZi.MonthPillar.Gan+sameDayLateZi.MonthPillar.Zhi),
		sixtyCycleForTest(t, sameDayLateZi.DayPillar.Gan+sameDayLateZi.DayPillar.Zhi),
		sixtyCycleForTest(t, sameDayLateZi.HourPillar.Gan+sameDayLateZi.HourPillar.Zhi),
	); err != nil {
		t.Fatalf("same-day late-Zi chart must round-trip through pillar validation: %v", err)
	}
	if afterMidnightDefault.DayPillar != afterMidnightSameDay.DayPillar || afterMidnightDefault.HourPillar != afterMidnightSameDay.HourPillar {
		t.Fatalf("policies must agree after midnight: default=%+v/%+v same=%+v/%+v",
			afterMidnightDefault.DayPillar, afterMidnightDefault.HourPillar,
			afterMidnightSameDay.DayPillar, afterMidnightSameDay.HourPillar)
	}
	if defaultLateZi.ZiHourPolicy != ZiHourLateZiNextDay || sameDayLateZi.ZiHourPolicy != ZiHourLateZiSameDay {
		t.Fatalf("result policy evidence missing: %q %q", defaultLateZi.ZiHourPolicy, sameDayLateZi.ZiHourPolicy)
	}
}

func sixtyCycleForTest(t *testing.T, name string) tyme.SixtyCycle {
	t.Helper()
	cycle, err := tyme.SixtyCycle{}.FromName(name)
	if err != nil {
		t.Fatal(err)
	}
	return *cycle
}

func TestCalculateAtWithPolicyRejectsUnknownPolicy(t *testing.T) {
	if _, err := (&BaziService{}).CalculateAtWithPolicy(2024, 6, 10, 23, 30, 0, "MALE", "silent-fallback"); err == nil {
		t.Fatal("unknown zi-hour policy must be rejected")
	}
}

func TestNormalizeBirthInputRecordsDefaultZiHourPolicy(t *testing.T) {
	normalized, err := NormalizeBirthInput(BirthInput{
		Year: 2024, Month: 6, Day: 10, Hour: 23, Minute: 30,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatal(err)
	}
	if normalized.ZiHourPolicy != DefaultZiHourPolicy || normalized.Validation.ZiHourPolicy != DefaultZiHourPolicy {
		t.Fatalf("default policy evidence = normalized %q validation %q", normalized.ZiHourPolicy, normalized.Validation.ZiHourPolicy)
	}
}

func TestZiHourPolicyIsBoundToCandidateID(t *testing.T) {
	service := &BaziService{}
	input := BirthInput{
		Year: 2024, Month: 6, Day: 10, Hour: 12,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
	}
	defaultSet, err := CalculateBirthCandidates(service, input)
	if err != nil {
		t.Fatal(err)
	}
	input.ZiHourPolicy = ZiHourLateZiSameDay
	sameDaySet, err := CalculateBirthCandidates(service, input)
	if err != nil {
		t.Fatal(err)
	}
	if defaultSet.Candidates[0].CandidateID == sameDaySet.Candidates[0].CandidateID {
		t.Fatal("candidate ID must bind the selected zi-hour policy even when current pillars agree")
	}
}

func TestCalculateAtWithPolicyDoesNotLeakAcrossConcurrentRequests(t *testing.T) {
	service := &BaziService{}
	expectedNext, err := service.CalculateAtWithPolicy(2024, 6, 10, 23, 30, 0, "MALE", ZiHourLateZiNextDay)
	if err != nil {
		t.Fatal(err)
	}
	expectedSame, err := service.CalculateAtWithPolicy(2024, 6, 10, 23, 30, 0, "MALE", ZiHourLateZiSameDay)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errors := make(chan string, 64)
	for i := 0; i < 64; i++ {
		policy, want := ZiHourLateZiNextDay, expectedNext.DayPillar
		if i%2 == 1 {
			policy, want = ZiHourLateZiSameDay, expectedSame.DayPillar
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, calculateErr := service.CalculateAtWithPolicy(2024, 6, 10, 23, 30, 0, "MALE", policy)
			if calculateErr != nil {
				errors <- calculateErr.Error()
				return
			}
			if got.DayPillar != want || got.ZiHourPolicy != policy {
				errors <- policy
			}
		}()
	}
	wg.Wait()
	close(errors)
	for failure := range errors {
		t.Fatalf("concurrent policy calculation leaked: %s", failure)
	}
}
