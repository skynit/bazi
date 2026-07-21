package fortune

import (
	"encoding/json"
	"math"
	"strings"
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

func TestComputeSummary_HighestLowestIndex(t *testing.T) {
	days := []DailyFortune{
		makeDay("2026-06-10", 70, map[string]int{"木": 2}, "正官"),
		makeDay("2026-06-11", 92, map[string]int{"火": 3}, "正印"),
		makeDay("2026-06-12", 55, map[string]int{"土": 1}, "正印"),
	}
	s := computeSummary(days)
	if s.HighestIndexDay != "2026-06-11" {
		t.Fatalf("highest_index_day: want 2026-06-11, got %s", s.HighestIndexDay)
	}
	if s.HighestIndex != 92 {
		t.Fatalf("highest_index: want 92, got %d", s.HighestIndex)
	}
	if s.LowestIndexDay != "2026-06-12" {
		t.Fatalf("lowest_index_day: want 2026-06-12, got %s", s.LowestIndexDay)
	}
	if s.LowestIndex != 55 {
		t.Fatalf("lowest_index: want 55, got %d", s.LowestIndex)
	}
}

func TestComputeSummary_JSONContractOmitsOutcomeLabels(t *testing.T) {
	days := []DailyFortune{
		makeDay("d1", 82, nil, ""),
		makeDay("d2", 85, nil, ""),
		makeDay("d3", 30, nil, ""),
		makeDay("d4", 35, nil, ""),
		makeDay("d5", 90, nil, ""),
	}
	payload, err := json.Marshal(computeSummary(days))
	if err != nil {
		t.Fatalf("marshal summary: %v", err)
	}
	text := string(payload)
	for _, forbidden := range []string{
		`"best_day"`, `"worst_day"`, `"peak_days"`, `"low_days"`,
		`"good_streak"`, `"bad_streak"`, `"key_advice"`,
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("summary must omit outcome label %s: %s", forbidden, text)
		}
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

func TestComputeSummary_AverageAndStandardDeviation(t *testing.T) {
	days := []DailyFortune{
		makeDay("d1", 60, nil, ""),
		makeDay("d2", 70, nil, ""),
		makeDay("d3", 80, nil, ""),
	}
	s := computeSummary(days)
	if math.Abs(s.AverageIndex-70.0) > 0.01 {
		t.Fatalf("average: want 70, got %f", s.AverageIndex)
	}
	// population stddev of [60,70,80] = sqrt((100+0+100)/3) ≈ 8.16
	if math.Abs(s.IndexStandardDeviation-8.16) > 0.05 {
		t.Fatalf("standard deviation: want approximately 8.16, got %f", s.IndexStandardDeviation)
	}
}

func TestComputeSummary_Empty(t *testing.T) {
	s := computeSummary(nil)
	if len(s.ElementDistribution) != 5 {
		t.Fatalf("empty: element_distribution should have 5 zero keys, got %d", len(s.ElementDistribution))
	}
	if s.HighestIndexDay != "" || s.LowestIndexDay != "" {
		t.Fatalf("empty: index extrema should be blank, got %q/%q", s.HighestIndexDay, s.LowestIndexDay)
	}
}
