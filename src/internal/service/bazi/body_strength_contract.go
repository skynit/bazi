package bazi

import (
	"reflect"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

func ValidBodyStrengthEvidence(analysis BodyStrengthResult, pillars []model.Pillar) bool {
	if len(pillars) != 4 || analysis.RuleVersion == "" || analysis.School == "" {
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
	want := calcBodyStrength(&ec)
	want.RuleVersion = analysis.RuleVersion
	want.School = analysis.School
	return reflect.DeepEqual(analysis, want)
}
