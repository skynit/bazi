package bazi

import (
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestValidFiveElementEvidenceRequiresCompletePillarRecomputation(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("甲子", "丙寅", "戊辰", "庚申", model.GenderMale)
	if err != nil {
		t.Fatal(err)
	}
	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}
	valid := func(
		scores map[string]int,
		details []ElementStrength,
		missing data.MissingElementAnalysis,
	) bool {
		return ValidFiveElementEvidence(scores, details, missing, pillars)
	}
	if !valid(result.FiveElements, result.ElementDetail, result.MissingElements) {
		t.Fatal("freshly calculated five-element evidence must validate")
	}

	tamperedScores := make(map[string]int, len(result.FiveElements))
	for element, score := range result.FiveElements {
		tamperedScores[element] = score
	}
	tamperedScores["木"]++
	if valid(
		tamperedScores,
		result.ElementDetail,
		data.FindMissingElements(tamperedScores),
	) {
		t.Fatal("consistently tampered scores and derived observations must be rejected")
	}

	tamperedDetails := append([]ElementStrength(nil), result.ElementDetail...)
	tamperedDetails[0].Total++
	if valid(result.FiveElements, tamperedDetails, result.MissingElements) {
		t.Fatal("tampered element breakdown must be rejected")
	}

	tamperedMissing := result.MissingElements
	tamperedMissing.Note = "tampered"
	if valid(result.FiveElements, result.ElementDetail, tamperedMissing) {
		t.Fatal("tampered derived observation must be rejected")
	}
}
