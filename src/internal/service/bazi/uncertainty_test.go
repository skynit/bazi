package bazi

import (
	"testing"
	"time"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

func TestCalculateBirthCandidatesExactTimeHasOneCandidate(t *testing.T) {
	set, err := CalculateBirthCandidates(&BaziService{}, BirthInput{
		Year: 1990, Month: 5, Day: 15, Hour: 8, Minute: 12, Second: 30,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
	})
	if err != nil {
		t.Fatalf("CalculateBirthCandidates: %v", err)
	}
	if len(set.Candidates) != 1 || set.RequiresCandidateSelection {
		t.Fatalf("exact candidate set = %d candidates, requires=%v", len(set.Candidates), set.RequiresCandidateSelection)
	}
	if len(set.StableFields) != 4 || len(set.UnstableFields) != 0 {
		t.Fatalf("stability = stable %v unstable %v", set.StableFields, set.UnstableFields)
	}
}

func TestCalculateBirthCandidatesStableUncertainRangeStaysSingle(t *testing.T) {
	set, err := CalculateBirthCandidates(&BaziService{}, BirthInput{
		Year: 1990, Month: 5, Day: 15, Hour: 8, Minute: 12, Second: 30,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
		TimeUncertain: true, UncertaintySeconds: 5 * 60,
	})
	if err != nil {
		t.Fatalf("CalculateBirthCandidates: %v", err)
	}
	if len(set.Candidates) != 1 || len(set.StableFields) != 4 {
		t.Fatalf("stable range = %d candidates, stable=%v", len(set.Candidates), set.StableFields)
	}
	if set.Candidates[0].DaYunStartAtMin == set.Candidates[0].DaYunStartAtMax {
		t.Fatal("stable four-pillar range must still expose changing DaYun start bounds")
	}
}

func TestNormalizeBirthInputRejectsUnquantifiedUncertainty(t *testing.T) {
	_, err := NormalizeBirthInput(BirthInput{
		Year: 1990, Month: 5, Day: 15, Hour: 8,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
		TimeUncertain: true,
	})
	if err == nil {
		t.Fatal("expected unquantified uncertainty to be rejected")
	}
}

func TestCalculateBirthCandidatesCrossesHourBoundary(t *testing.T) {
	set, err := CalculateBirthCandidates(&BaziService{}, BirthInput{
		Year: 2024, Month: 6, Day: 10, Hour: 0, Minute: 59, Second: 59,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
		TimeUncertain: true, UncertaintySeconds: 1,
	})
	if err != nil {
		t.Fatalf("CalculateBirthCandidates: %v", err)
	}
	if len(set.Candidates) != 2 {
		t.Fatalf("candidate count = %d, want 2: %+v", len(set.Candidates), set.Candidates)
	}
	if !containsUncertaintyField(set.UnstableFields, "hour_pillar") || len(set.UnstableFields) != 1 {
		t.Fatalf("unstable fields = %v, want only hour_pillar", set.UnstableFields)
	}
	if got := set.Uncertainty.CrossedBoundaries[0].Type; got != "hour_branch" {
		t.Fatalf("boundary type = %q, want hour_branch", got)
	}
	if set.Candidates[0].DaYunStartAtMin == set.Candidates[0].DaYunStartAtMax {
		t.Fatal("DaYun range collapsed even though candidate spans multiple seconds")
	}
}

func TestCalculateBirthCandidatesCrossesLateZiDayBoundary(t *testing.T) {
	set, err := CalculateBirthCandidates(&BaziService{}, BirthInput{
		Year: 2024, Month: 6, Day: 10, Hour: 22, Minute: 59, Second: 59,
		CalendarType: "SOLAR", Gender: "FEMALE", Timezone: "Asia/Shanghai",
		TimeUncertain: true, UncertaintySeconds: 1,
	})
	if err != nil {
		t.Fatalf("CalculateBirthCandidates: %v", err)
	}
	if len(set.Candidates) != 2 {
		t.Fatalf("candidate count = %d, want 2", len(set.Candidates))
	}
	if !containsUncertaintyField(set.UnstableFields, "day_pillar") || !containsUncertaintyField(set.UnstableFields, "hour_pillar") {
		t.Fatalf("unstable fields = %v, want day and hour", set.UnstableFields)
	}
	if got := set.Uncertainty.CrossedBoundaries[0].Type; got != "zi_hour_day_boundary" {
		t.Fatalf("boundary type = %q, want zi_hour_day_boundary", got)
	}
}

func TestCalculateBirthCandidatesSameDayLateZiMovesDayBoundaryToMidnight(t *testing.T) {
	service := &BaziService{}
	lateZi, err := CalculateBirthCandidates(service, BirthInput{
		Year: 2024, Month: 6, Day: 10, Hour: 22, Minute: 59, Second: 59,
		CalendarType: "SOLAR", Gender: "FEMALE", Timezone: "Asia/Shanghai",
		ZiHourPolicy: ZiHourLateZiSameDay, TimeUncertain: true, UncertaintySeconds: 1,
	})
	if err != nil {
		t.Fatalf("late-Zi candidates: %v", err)
	}
	if len(lateZi.UnstableFields) != 1 || lateZi.UnstableFields[0] != "hour_pillar" || lateZi.Uncertainty.CrossedBoundaries[0].Type != "hour_branch" {
		t.Fatalf("23:00 same-day policy boundary = unstable %v boundaries %+v", lateZi.UnstableFields, lateZi.Uncertainty.CrossedBoundaries)
	}

	midnight, err := CalculateBirthCandidates(service, BirthInput{
		Year: 2024, Month: 6, Day: 10, Hour: 23, Minute: 59, Second: 59,
		CalendarType: "SOLAR", Gender: "FEMALE", Timezone: "Asia/Shanghai",
		ZiHourPolicy: ZiHourLateZiSameDay, TimeUncertain: true, UncertaintySeconds: 1,
	})
	if err != nil {
		t.Fatalf("midnight candidates: %v", err)
	}
	if len(midnight.UnstableFields) != 1 || midnight.UnstableFields[0] != "day_pillar" {
		t.Fatalf("midnight unstable fields = %v, want only day_pillar", midnight.UnstableFields)
	}
	if got := midnight.Uncertainty.CrossedBoundaries[0].Type; got != "civil_day" {
		t.Fatalf("midnight boundary type = %q, want civil_day", got)
	}
}

func TestCalculateBirthCandidatesCrossesJieAtExactSecond(t *testing.T) {
	term, err := tyme.SolarTerm{}.FromName(2022, "惊蛰")
	if err != nil {
		t.Fatalf("create solar term: %v", err)
	}
	at := term.GetJulianDay().GetSolarTime()
	set, err := CalculateBirthCandidates(&BaziService{}, BirthInput{
		Year: at.GetYear(), Month: at.GetMonth(), Day: at.GetDay(),
		Hour: at.GetHour(), Minute: at.GetMinute(), Second: at.GetSecond(),
		CalendarType: "SOLAR", Gender: "FEMALE", Timezone: "Asia/Shanghai",
		TimeUncertain: true, UncertaintySeconds: 1,
	})
	if err != nil {
		t.Fatalf("CalculateBirthCandidates: %v", err)
	}
	if len(set.Candidates) != 2 || !containsUncertaintyField(set.UnstableFields, "month_pillar") {
		t.Fatalf("jie candidate set = %d candidates, unstable=%v", len(set.Candidates), set.UnstableFields)
	}
	if got := set.Uncertainty.CrossedBoundaries[0].Type; got != "solar_term" {
		t.Fatalf("boundary type = %q, want solar_term", got)
	}
}

func TestCalculateBirthCandidatesCrossesLiChunYearAndMonth(t *testing.T) {
	term, err := tyme.SolarTerm{}.FromName(2022, "立春")
	if err != nil {
		t.Fatalf("create solar term: %v", err)
	}
	at := term.GetJulianDay().GetSolarTime()
	set, err := CalculateBirthCandidates(&BaziService{}, BirthInput{
		Year: at.GetYear(), Month: at.GetMonth(), Day: at.GetDay(),
		Hour: at.GetHour(), Minute: at.GetMinute(), Second: at.GetSecond(),
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
		TimeUncertain: true, UncertaintySeconds: 1,
	})
	if err != nil {
		t.Fatalf("CalculateBirthCandidates: %v", err)
	}
	if len(set.Candidates) != 2 || !containsUncertaintyField(set.UnstableFields, "year_pillar") || !containsUncertaintyField(set.UnstableFields, "month_pillar") {
		t.Fatalf("LiChun candidates = %d, unstable=%v", len(set.Candidates), set.UnstableFields)
	}
}

func TestCalculateBirthCandidatesRecalculatesTrueSolarEndpoints(t *testing.T) {
	longitude := 120.0
	set, err := CalculateBirthCandidates(&BaziService{}, BirthInput{
		Year: 2024, Month: 2, Day: 11, Hour: 12,
		CalendarType: "SOLAR", Gender: "MALE", Timezone: "Asia/Shanghai",
		Longitude: &longitude, UseTrueSolarTime: true,
		TimeUncertain: true, UncertaintySeconds: 24 * 60 * 60,
	})
	if err != nil {
		t.Fatalf("CalculateBirthCandidates: %v", err)
	}
	first := set.Candidates[0].Normalized.Validation
	last := set.Candidates[len(set.Candidates)-1].Normalized.Validation
	if first.EquationOfTimeSeconds == last.EquationOfTimeSeconds {
		t.Fatalf("true-solar endpoints reused center equation of time: both=%d", first.EquationOfTimeSeconds)
	}
	if set.Uncertainty.AlgorithmUncertaintySeconds != trueSolarUncertainty || set.Uncertainty.EffectiveSeconds != 24*60*60+trueSolarUncertainty {
		t.Fatalf("algorithm uncertainty evidence = %+v", set.Uncertainty)
	}
	if len(set.Candidates) >= 100 {
		t.Fatalf("candidate grouping exploded to %d charts", len(set.Candidates))
	}
}

func TestPartitionCandidateRangesDoesNotSkipStateDuringUTCOffsetRollback(t *testing.T) {
	chartA := &BaziResult{
		YearPillar: model.Pillar{Gan: "甲", Zhi: "子"}, MonthPillar: model.Pillar{Gan: "丙", Zhi: "寅"},
		DayPillar: model.Pillar{Gan: "戊", Zhi: "辰"}, HourPillar: model.Pillar{Gan: "庚", Zhi: "申"},
	}
	chartB := &BaziResult{
		YearPillar: model.Pillar{Gan: "乙", Zhi: "丑"}, MonthPillar: model.Pillar{Gan: "丁", Zhi: "卯"},
		DayPillar: model.Pillar{Gan: "己", Zhi: "巳"}, HourPillar: model.Pillar{Gan: "辛", Zhi: "酉"},
	}
	beforeRollback := time.FixedZone("before", 2*60*60)
	afterRollback := time.FixedZone("after", 0)
	pointAt := func(offset int) (*candidatePoint, error) {
		point := &candidatePoint{offset: offset, result: chartA, inputTime: time.Unix(int64(offset), 0).In(afterRollback)}
		switch {
		case offset < 1200:
			point.inputTime = time.Unix(int64(offset), 0).In(beforeRollback)
		case offset < 2400:
			point.result = chartB
		}
		return point, nil
	}

	ranges, err := partitionCandidateRanges(pointAt, 0, uncertaintyProbeSpanSeconds)
	if err != nil {
		t.Fatal(err)
	}
	wants := []candidateRange{
		{start: 0, end: 1199, key: fourPillarKey(chartA)},
		{start: 1200, end: 2399, key: fourPillarKey(chartB)},
		{start: 2400, end: uncertaintyProbeSpanSeconds, key: fourPillarKey(chartA)},
	}
	if len(ranges) != len(wants) {
		t.Fatalf("rollback ranges = %+v, want %+v", ranges, wants)
	}
	for index := range wants {
		if ranges[index] != wants[index] {
			t.Errorf("rollback range %d = %+v, want %+v", index, ranges[index], wants[index])
		}
	}
}

func containsUncertaintyField(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
