package service

import (
	"math"
	"testing"
	"time"
)

func testUserChart(t *testing.T) *BaziResult {
	t.Helper()
	svc := &BaziService{}
	chart, err := svc.Calculate(1990, 1, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatalf("failed to create test user chart: %v", err)
	}
	return chart
}

func TestDailyFortune(t *testing.T) {
	engine := NewFortuneEngine()
	chart := testUserChart(t)

	queryDate := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	df := engine.CalculateDaily(chart, queryDate)

	if df.Date != "2024-06-01" {
		t.Errorf("Date = %q, want %q", df.Date, "2024-06-01")
	}
	if df.DayPillar.Gan == "" || df.DayPillar.Zhi == "" {
		t.Errorf("DayPillar is empty: %+v", df.DayPillar)
	}
	if df.Score < 0 || df.Score > 100 {
		t.Errorf("Score = %d, want in [0, 100]", df.Score)
	}
	if df.LuckyColor == "" {
		t.Error("LuckyColor is empty")
	}
	if len(df.LuckyNumbers) < 2 {
		t.Errorf("LuckyNumbers has %d items, want >= 2: %v", len(df.LuckyNumbers), df.LuckyNumbers)
	}
	if df.WealthDir == "" {
		t.Error("WealthDir is empty")
	}
	if df.ClashZodiac == "" {
		t.Error("ClashZodiac is empty")
	}
	if len(df.AuspiciousHours) == 0 {
		t.Error("AuspiciousHours is empty")
	}
	if len(df.Yi) == 0 {
		t.Error("Yi is empty")
	}
	if len(df.Ji) == 0 {
		t.Error("Ji is empty")
	}
	if df.ShengKe.Summary == "" {
		t.Error("ShengKe.Summary is empty")
	}
	if len(df.ElementImages) != 5 {
		t.Errorf("ElementImages has %d items, want 5", len(df.ElementImages))
	}

	expected := []string{"金", "木", "水", "火", "土"}
	for i, img := range df.ElementImages {
		if img.Element != expected[i] {
			t.Errorf("ElementImages[%d].Element = %q, want %q", i, img.Element, expected[i])
		}
		if img.ImageURL == "" {
			t.Errorf("ElementImages[%d].ImageURL is empty", i)
		}
	}

	t.Logf("Daily: date=%s pillar=%s%s score=%d color=%s nums=%v dir=%s clash=%s hours=%v",
		df.Date, df.DayPillar.Gan, df.DayPillar.Zhi, df.Score,
		df.LuckyColor, df.LuckyNumbers, df.WealthDir, df.ClashZodiac, df.AuspiciousHours)
	t.Logf("ShengKe: stem=%s branch=%s summary=%s",
		df.ShengKe.DayStemRelation, df.ShengKe.DayBranchRelation, df.ShengKe.Summary)
	t.Logf("Yi: %+v", df.Yi)
	t.Logf("Ji: %+v", df.Ji)
}

func TestWeeklyFortuneHasSevenItems(t *testing.T) {
	engine := NewFortuneEngine()
	chart := testUserChart(t)

	weekStart := time.Date(2024, 6, 3, 0, 0, 0, 0, time.UTC)
	wf := engine.CalculateWeekly(chart, weekStart)

	if len(wf.DailyFortunes) != 7 {
		t.Fatalf("DailyFortunes has %d items, want 7", len(wf.DailyFortunes))
	}
	if len(wf.ElementTrend) != 7 {
		t.Fatalf("ElementTrend has %d items, want 7", len(wf.ElementTrend))
	}

	for i, df := range wf.DailyFortunes {
		if df.Score < 0 || df.Score > 100 {
			t.Errorf("DailyFortunes[%d].Score = %d, want in [0, 100]", i, df.Score)
		}
	}

	if wf.WeeklyScore < 0 || wf.WeeklyScore > 100 {
		t.Errorf("WeeklyScore = %d, want in [0, 100]", wf.WeeklyScore)
	}
	if wf.OverallSummary == "" {
		t.Error("OverallSummary is empty")
	}
	if wf.WeekStart != "2024-06-03" {
		t.Errorf("WeekStart = %q, want %q", wf.WeekStart, "2024-06-03")
	}

	t.Logf("Weekly: start=%s score=%d days=%d summary=%s",
		wf.WeekStart, wf.WeeklyScore, len(wf.DailyFortunes), wf.OverallSummary)
	t.Logf("First day pillar: %s%s", wf.DailyFortunes[0].DayPillar.Gan, wf.DailyFortunes[0].DayPillar.Zhi)
}

func TestMonthlyFortuneCorrectDayCount(t *testing.T) {
	engine := NewFortuneEngine()
	chart := testUserChart(t)

	tests := []struct {
		year, month, wantDays int
	}{
		{2024, 1, 31},
		{2024, 2, 29},
		{2024, 4, 30},
		{2023, 2, 28},
	}

	for _, tt := range tests {
		mf := engine.CalculateMonthly(chart, tt.year, tt.month)

		if len(mf.DailyFortunes) != tt.wantDays {
			t.Errorf("%d-%02d: DailyFortunes has %d items, want %d",
				tt.year, tt.month, len(mf.DailyFortunes), tt.wantDays)
		}
		if len(mf.ElementTrend) != tt.wantDays {
			t.Errorf("%d-%02d: ElementTrend has %d items, want %d",
				tt.year, tt.month, len(mf.ElementTrend), tt.wantDays)
		}
		if mf.Year != tt.year {
			t.Errorf("%d-%02d: Year = %d, want %d", tt.year, tt.month, mf.Year, tt.year)
		}
		if mf.Month != tt.month {
			t.Errorf("%d-%02d: Month = %d, want %d", tt.year, tt.month, mf.Month, tt.month)
		}
		if mf.MonthlyScore < 0 || mf.MonthlyScore > 100 {
			t.Errorf("%d-%02d: MonthlyScore = %d, want in [0, 100]", tt.year, tt.month, mf.MonthlyScore)
		}
		if mf.OverallSummary == "" {
			t.Errorf("%d-%02d: OverallSummary is empty", tt.year, tt.month)
		}
	}
}

func TestScoreAlwaysInRange(t *testing.T) {
	engine := NewFortuneEngine()
	chart := testUserChart(t)

	start := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	for i := 0; i < 365; i++ {
		day := start.AddDate(0, 0, i)
		df := engine.CalculateDaily(chart, day)
		if df.Score < 0 {
			t.Errorf("%s: Score = %d, want >= 0", df.Date, df.Score)
		}
		if df.Score > 100 {
			t.Errorf("%s: Score = %d, want <= 100", df.Date, df.Score)
		}
	}
}

func TestElementPercentagesSumToNear100(t *testing.T) {
	engine := NewFortuneEngine()
	chart := testUserChart(t)

	date := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	df := engine.CalculateDaily(chart, date)

	trend := engine.elementTrend(date, df.Score)
	total := trend.Metal + trend.Wood + trend.Water + trend.Fire + trend.Earth

	if math.Abs(total-100) > 0.01 {
		t.Errorf("Element percentages sum to %.2f, want ~100: Metal=%.2f Wood=%.2f Water=%.2f Fire=%.2f Earth=%.2f",
			total, trend.Metal, trend.Wood, trend.Water, trend.Fire, trend.Earth)
	}

	for name, pct := range map[string]float64{
		"Metal": trend.Metal, "Wood": trend.Wood, "Water": trend.Water,
		"Fire": trend.Fire, "Earth": trend.Earth,
	} {
		if pct < 0 {
			t.Errorf("%s percentage = %.2f, want >= 0", name, pct)
		}
	}

	t.Logf("Elements: Metal=%.1f%% Wood=%.1f%% Water=%.1f%% Fire=%.1f%% Earth=%.1f%% Sum=%.1f%%",
		trend.Metal, trend.Wood, trend.Water, trend.Fire, trend.Earth, total)
}
