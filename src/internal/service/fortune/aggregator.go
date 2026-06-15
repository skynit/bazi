package fortune

import (
	"fmt"
	"math"
	"sort"

	"bazi/internal/model"
)

const (
	peakScoreThreshold = 80
	lowScoreThreshold  = 40
	goodScoreThreshold = 70
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
		bestIdx, worstIdx     int
		peak, low             []string
		goodStreak, badStreak int
		curGood, curBad       int
		sumScore              float64
		elementTotal          float64
		tenGodCount           = map[string]int{}
	)

	for i, d := range days {
		sumScore += float64(d.Score)
		if d.Score > days[bestIdx].Score {
			bestIdx = i
		}
		if d.Score < days[worstIdx].Score {
			worstIdx = i
		}
		if d.Score >= peakScoreThreshold {
			peak = append(peak, d.Date)
		}
		if d.Score <= lowScoreThreshold {
			low = append(low, d.Date)
		}
		if d.Score >= goodScoreThreshold {
			curGood++
			if curGood > goodStreak {
				goodStreak = curGood
			}
		} else {
			curGood = 0
		}
		if d.Score <= lowScoreThreshold {
			curBad++
			if curBad > badStreak {
				badStreak = curBad
			}
		} else {
			curBad = 0
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

	advice := buildKeyAdvice(days[bestIdx], days[worstIdx], dominantElement, dominantTenGod, avg)

	return model.FortuneSummary{
		BestDay:             days[bestIdx].Date,
		WorstDay:            days[worstIdx].Date,
		BestScore:           days[bestIdx].Score,
		WorstScore:          days[worstIdx].Score,
		PeakDays:            peak,
		LowDays:             low,
		ElementDistribution: dist,
		DominantElement:     dominantElement,
		DominantTenGod:      dominantTenGod,
		GoodStreak:          goodStreak,
		BadStreak:           badStreak,
		AverageScore:        round2(avg),
		Volatility:          round2(volatility),
		KeyAdvice:           advice,
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

func buildKeyAdvice(best, worst DailyFortune, dominantElem, dominantTenGod string, avg float64) string {
	tone := "平稳"
	switch {
	case avg >= 75:
		tone = "整体偏吉，势头积极"
	case avg >= 60:
		tone = "稳中有进，可循序推进"
	case avg >= 45:
		tone = "起伏中性，宜守不宜攻"
	default:
		tone = "宜静养待时，避免冒进"
	}
	parts := tone
	if dominantElem != "" {
		parts += fmt.Sprintf("；五行以%s为主导", dominantElem)
	}
	if dominantTenGod != "" {
		parts += fmt.Sprintf("，主气为%s", dominantTenGod)
	}
	if best.Date != "" {
		parts += fmt.Sprintf("。最佳%s（%d 分）", best.Date, best.Score)
	}
	if worst.Date != "" && worst.Date != best.Date {
		parts += fmt.Sprintf("，低谷%s（%d 分）", worst.Date, worst.Score)
	}
	return parts + "。"
}

