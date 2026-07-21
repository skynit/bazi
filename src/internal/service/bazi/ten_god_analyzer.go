package bazi

import (
	"math"
	"reflect"
	"sort"
)

const tenGodOccurrenceBasis = "three_visible_stems_and_all_hidden_stems_counted_equally"

// TenGodAnalysis records a reproducible ordering of ten-god occurrences. It
// does not infer personality, relationships, career, wealth, or future events.
type TenGodAnalysis struct {
	RuleID               string       `json:"rule_id"`
	CalculationMethod    string       `json:"calculation_method"`
	TotalOccurrences     int          `json:"total_occurrences"`
	DominantGods         []string     `json:"dominant_gods"`
	DominantPercent      float64      `json:"dominant_percent"`
	RankedGods           []TenGodRank `json:"ranked_gods"`
	Status               string       `json:"status"`
	ValidationStatus     string       `json:"validation_status"`
	InterpretationStatus string       `json:"interpretation_status"`
	Limitations          []string     `json:"limitations"`
}

// TenGodRank is one stable, dense-ranked occurrence record. Equal counts share
// the same rank and remain ordered by the canonical ten-god table.
type TenGodRank struct {
	Rank                 int     `json:"rank"`
	God                  string  `json:"god"`
	Count                int     `json:"count"`
	Percent              float64 `json:"percent"`
	Basis                string  `json:"basis"`
	Status               string  `json:"status"`
	InterpretationStatus string  `json:"interpretation_status"`
}

// ObserveTenGodDistribution converts the raw occurrence distribution into an
// auditable ranking without assigning favorability or real-world meaning.
func ObserveTenGodDistribution(proportions []TenGodRatio) *TenGodAnalysis {
	analysis := &TenGodAnalysis{
		RuleID:               "bazi.ten-god-occurrence-ranking-v1",
		CalculationMethod:    tenGodOccurrenceBasis,
		DominantGods:         []string{},
		RankedGods:           []TenGodRank{},
		Status:               "unavailable",
		ValidationStatus:     "not_validated",
		InterpretationStatus: "not_adjudicated",
		Limitations: []string{
			"visible stems and hidden stems are counted equally",
			"hidden-stem depth and seasonal strength are not weighted",
			"occurrence share is not influence strength or outcome probability",
		},
	}
	if len(proportions) != len(tenGodNames) {
		return analysis
	}

	canonicalOrder := make(map[string]int, len(tenGodNames))
	for index, name := range tenGodNames {
		canonicalOrder[name] = index
	}
	items := make([]TenGodRatio, 0, len(proportions))
	seen := make(map[string]bool, len(proportions))
	for _, item := range proportions {
		if _, ok := canonicalOrder[item.Name]; !ok || seen[item.Name] || item.Count < 0 {
			return analysis
		}
		seen[item.Name] = true
		items = append(items, item)
		analysis.TotalOccurrences += item.Count
	}
	if analysis.TotalOccurrences == 0 {
		return analysis
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].Count != items[j].Count {
			return items[i].Count > items[j].Count
		}
		return canonicalOrder[items[i].Name] < canonicalOrder[items[j].Name]
	})

	rank := 0
	previousCount := -1
	for _, item := range items {
		if item.Count != previousCount {
			rank++
			previousCount = item.Count
		}
		percent := math.Round(float64(item.Count)*10000/float64(analysis.TotalOccurrences)) / 100
		analysis.RankedGods = append(analysis.RankedGods, TenGodRank{
			Rank:                 rank,
			God:                  item.Name,
			Count:                item.Count,
			Percent:              percent,
			Basis:                tenGodOccurrenceBasis,
			Status:               "observed",
			InterpretationStatus: "not_adjudicated",
		})
	}

	maxCount := items[0].Count
	analysis.DominantPercent = analysis.RankedGods[0].Percent
	for _, item := range items {
		if item.Count != maxCount {
			break
		}
		analysis.DominantGods = append(analysis.DominantGods, item.Name)
	}
	analysis.Status = "observed"
	return analysis
}

// ValidTenGodAnalysis rejects stale or partially migrated snapshots instead of
// silently serving an incomplete occurrence contract.
func ValidTenGodAnalysis(analysis *TenGodAnalysis, proportions []TenGodRatio) bool {
	expected := ObserveTenGodDistribution(proportions)
	return expected.Status == "observed" && reflect.DeepEqual(analysis, expected)
}
