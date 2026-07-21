package fortune

import (
	"math"
	"sort"

	"bazi/internal/model"
)

// computeSummary produces a FortuneSummary aggregated over the given daily fortunes.
// Empty input yields a zero-valued summary; element_distribution is always
// initialized with the five wuxing keys so callers can index without nil checks.
func computeSummary(days []DailyFortune) model.FortuneSummary {
	dist := map[string]float64{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}
	if len(days) == 0 {
		return model.FortuneSummary{ElementDistribution: dist}
	}

	var (
		highestIdx, lowestIdx int
		sumScore              float64
		elementTotal          float64
		tenGodCount           = map[string]int{}
	)

	for i, d := range days {
		sumScore += float64(d.Score)
		if d.Score > days[highestIdx].Score {
			highestIdx = i
		}
		if d.Score < days[lowestIdx].Score {
			lowestIdx = i
		}
		for elem, n := range d.TodayElements {
			if _, ok := dist[elem]; !ok {
				continue
			}
			dist[elem] += float64(n)
			elementTotal += float64(n)
		}
		if d.Rikuyo != nil && d.Rikuyo.TodayTenGod != "" {
			tenGodCount[d.Rikuyo.TodayTenGod]++
		}
	}

	avg := sumScore / float64(len(days))
	var sqDiff float64
	for _, d := range days {
		diff := float64(d.Score) - avg
		sqDiff += diff * diff
	}
	volatility := math.Sqrt(sqDiff / float64(len(days)))

	if elementTotal > 0 {
		for k := range dist {
			dist[k] = round4(dist[k] / elementTotal)
		}
	}

	dominantElement := dominantKeyFloat(dist)
	dominantTenGod := dominantKeyInt(tenGodCount)

	return model.FortuneSummary{
		HighestIndexDay:        days[highestIdx].Date,
		LowestIndexDay:         days[lowestIdx].Date,
		HighestIndex:           days[highestIdx].Score,
		LowestIndex:            days[lowestIdx].Score,
		ElementDistribution:    dist,
		DominantElement:        dominantElement,
		DominantTenGod:         dominantTenGod,
		AverageIndex:           round2(avg),
		IndexStandardDeviation: round2(volatility),
	}
}

func round2(v float64) float64 { return math.Round(v*100) / 100 }
func round4(v float64) float64 { return math.Round(v*10000) / 10000 }

func dominantKeyFloat(m map[string]float64) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	best := keys[0]
	for _, k := range keys {
		if m[k] > m[best] {
			best = k
		}
	}
	if m[best] == 0 {
		return ""
	}
	return best
}

func dominantKeyInt(m map[string]int) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	best := keys[0]
	for _, k := range keys {
		if m[k] > m[best] {
			best = k
		}
	}
	if m[best] == 0 {
		return ""
	}
	return best
}
