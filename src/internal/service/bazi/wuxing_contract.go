package bazi

import (
	"reflect"

	"bazi/internal/model"
	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

// ValidFiveElementEvidence requires the raw score, its breakdown, and every
// public derived observation to be exactly reproducible from the four pillars.
func ValidFiveElementEvidence(
	scores map[string]int,
	details []ElementStrength,
	missing data.MissingElementAnalysis,
	pillars []model.Pillar,
) bool {
	if len(pillars) != 4 {
		return false
	}
	cycles := make([]*tyme.SixtyCycle, 0, len(pillars))
	for _, pillar := range pillars {
		cycle, err := tyme.SixtyCycle{}.FromName(pillar.Gan + pillar.Zhi)
		if err != nil {
			return false
		}
		cycles = append(cycles, cycle)
	}
	ec := tyme.EightChar{}.FromSixtyCycle(*cycles[0], *cycles[1], *cycles[2], *cycles[3])
	wantScores := calcFiveElements(&ec)
	wantDetails := calcElementDetail(&ec)
	return reflect.DeepEqual(scores, wantScores) &&
		reflect.DeepEqual(details, wantDetails) &&
		reflect.DeepEqual(missing, data.FindMissingElements(wantScores))
}
