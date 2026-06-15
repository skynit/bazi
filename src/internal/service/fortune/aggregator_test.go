package fortune

import (
	"math"
	"testing"
)

// makeDay constructs a DailyFortune with the minimum fields used by computeSummary.
func makeDay(date string, score int, elements map[string]int, tenGod string) DailyFortune {
	df := DailyFortune{
		Date:          date,
		Score:         score,
		TodayElements: elements,
	}
	if tenGod != "" {
		df.Rikuyo = &RikuyoResult{TodayTenGod: tenGod}
	}
	return df
}

func TestComputeSummary_BestWorst(t *testing.T) {
	days := []DailyFortune{
		makeDay("2026-06-10", 70, map[string]int{"木": 2}, "正官"),
		makeDay("2026-06-11", 92, map[string]int{"火": 3}, "正印"),
		makeDay("2026-06-12", 55, map[string]int{"土": 1}, "正印"),
	}
	s := computeSummary(days)
	if s.BestDay != "2026-06-11" {
		t.Fatalf("best_day: want 2026-06-11, got %s", s.BestDay)
	}
	if s.BestScore != 92 {
		t.Fatalf("best_score: want 92, got %d", s.BestScore)
	}
	if s.WorstDay != "2026-06-12" {
		t.Fatalf("worst_day: want 2026-06-12, got %s", s.WorstDay)
	}
	if s.WorstScore != 55 {
		t.Fatalf("worst_score: want 55, got %d", s.WorstScore)
	}
}

func TestComputeSummary_PeakLowAndStreak(t *testing.T) {
	// 5 days: 82,85,30,35,90 -> peak [d1,d2,d5], low [d3,d4]
	// good streak (>=70): d1-d2 = 2, d5 = 1 -> max 2
	// bad streak (<=40): d3-d4 = 2 -> max 2
	days := []DailyFortune{
		makeDay("d1", 82, nil, ""),
		makeDay("d2", 85, nil, ""),
		makeDay("d3", 30, nil, ""),
		makeDay("d4", 35, nil, ""),
		makeDay("d5", 90, nil, ""),
	}
	s := computeSummary(days)
	if len(s.PeakDays) != 3 {
		t.Fatalf("peak_days len: want 3, got %d (%v)", len(s.PeakDays), s.PeakDays)
	}
	if len(s.LowDays) != 2 {
		t.Fatalf("low_days len: want 2, got %d (%v)", len(s.LowDays), s.LowDays)
	}
	if s.GoodStreak != 2 {
		t.Fatalf("good_streak: want 2, got %d", s.GoodStreak)
	}
	if s.BadStreak != 2 {
		t.Fatalf("bad_streak: want 2, got %d", s.BadStreak)
	}
}

func TestComputeSummary_ElementDistributionNormalized(t *testing.T) {
	days := []DailyFortune{
		makeDay("d1", 70, map[string]int{"木": 4, "火": 1}, ""),
		makeDay("d2", 65, map[string]int{"木": 2, "土": 3}, ""),
	}
	s := computeSummary(days)
	var sum float64
	for _, v := range s.ElementDistribution {
		sum += v
	}
	if math.Abs(sum-1.0) > 0.001 {
		t.Fatalf("element_distribution sum: want ≈1.0, got %f", sum)
	}
	if s.DominantElement != "木" {
		t.Fatalf("dominant_element: want 木, got %s", s.DominantElement)
	}
}

func TestComputeSummary_DominantTenGod(t *testing.T) {
	days := []DailyFortune{
		makeDay("d1", 70, nil, "正官"),
		makeDay("d2", 70, nil, "正印"),
		makeDay("d3", 70, nil, "正印"),
	}
	s := computeSummary(days)
	if s.DominantTenGod != "正印" {
		t.Fatalf("dominant_ten_god: want 正印, got %s", s.DominantTenGod)
	}
}

func TestComputeSummary_AverageVolatility(t *testing.T) {
	days := []DailyFortune{
		makeDay("d1", 60, nil, ""),
		makeDay("d2", 70, nil, ""),
		makeDay("d3", 80, nil, ""),
	}
	s := computeSummary(days)
	if math.Abs(s.AverageScore-70.0) > 0.01 {
		t.Fatalf("average: want 70, got %f", s.AverageScore)
	}
	// population stddev of [60,70,80] = sqrt((100+0+100)/3) ≈ 8.16
	if math.Abs(s.Volatility-8.16) > 0.05 {
		t.Fatalf("volatility: want ≈8.16, got %f", s.Volatility)
	}
}

func TestComputeSummary_Empty(t *testing.T) {
	s := computeSummary(nil)
	if len(s.ElementDistribution) != 5 {
		t.Fatalf("empty: element_distribution should have 5 zero keys, got %d", len(s.ElementDistribution))
	}
	if s.BestDay != "" || s.WorstDay != "" {
		t.Fatalf("empty: best/worst should be blank, got %q/%q", s.BestDay, s.WorstDay)
	}
}

func TestComputeSummary_KeyAdviceNotEmpty(t *testing.T) {
	days := []DailyFortune{
		makeDay("2026-06-10", 80, map[string]int{"木": 2}, "正官"),
	}
	s := computeSummary(days)
	if s.KeyAdvice == "" {
		t.Fatal("key_advice should not be empty")
	}
}
