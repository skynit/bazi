package bazi

import (
	"reflect"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

// ValidPillarDerivedEvidence requires all public pillar-derived facts to match
// a fresh calculation from the four pillars and normalized gender.
func ValidPillarDerivedEvidence(result *BaziResult, gender string) bool {
	if result == nil {
		return false
	}
	if _, err := toTymeGender(gender); err != nil {
		return false
	}
	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}
	cycles := make([]*tyme.SixtyCycle, 0, len(pillars))
	for _, pillar := range pillars {
		cycle, err := tyme.SixtyCycle{}.FromName(pillar.Gan + pillar.Zhi)
		if err != nil {
			return false
		}
		cycles = append(cycles, cycle)
	}
	ec := tyme.EightChar{}.FromSixtyCycle(*cycles[0], *cycles[1], *cycles[2], *cycles[3])

	mingGongGanZhi, err := calcMingGongGanZhi(result.YearPillar.Gan, result.MonthPillar.Zhi, result.HourPillar.Zhi)
	if err != nil {
		return false
	}
	wantProportions := calcTenGodProportion(&ec, result.DayPillar.Gan)
	wantShenSha, err := CalcShenShaByPillars(ShenShaPillars{
		Year: result.YearPillar, Month: result.MonthPillar,
		Day: result.DayPillar, Hour: result.HourPillar, Gender: gender,
	})
	if err != nil {
		return false
	}
	wantShell := &BaziResult{
		YearPillar: result.YearPillar, MonthPillar: result.MonthPillar,
		DayPillar: result.DayPillar, HourPillar: result.HourPillar,
	}

	return reflect.DeepEqual(result.TenGods, calcTenGods(&ec)) &&
		reflect.DeepEqual(result.HiddenStems, calcHiddenStems(&ec)) &&
		reflect.DeepEqual(result.PillarDetails, buildPillarDetails(pillars)) &&
		reflect.DeepEqual(result.MingGong, buildMingGongDetail(mingGongGanZhi)) &&
		reflect.DeepEqual(result.TenGodProportion, wantProportions) &&
		reflect.DeepEqual(result.TenGodAnalysis, ObserveTenGodDistribution(wantProportions)) &&
		reflect.DeepEqual(result.DayShenSha, wantShenSha.Day) &&
		reflect.DeepEqual(result.DayShenShaDetails, BuildShenShaDetails(wantShenSha.Day)) &&
		reflect.DeepEqual(result.GlobalShenSha, wantShenSha.Global) &&
		reflect.DeepEqual(result.GlobalShenShaDetails, BuildShenShaDetails(wantShenSha.Global)) &&
		reflect.DeepEqual(result.ShenShaByPillar, buildPillarShenSha(wantShell, wantShenSha))
}
