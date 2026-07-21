package ziwei

import "github.com/6tail/tyme4go/tyme"

var transitPalaceOrder = [12]string{
	"命宫", "父母", "福德", "田宅", "事业", "交友",
	"迁移", "疾厄", "财帛", "子女", "夫妻", "兄弟",
}

func buildDayunPalaceNames(chart *ZiWeiChart, stageBranch int) ([12]string, bool) {
	var result [12]string
	if chart == nil || stageBranch < 0 || stageBranch >= len(BranchNames) {
		return result, false
	}
	for i, palace := range chart.Palaces {
		branch, ok := BranchIndex[palace.Branch]
		if !ok {
			return [12]string{}, false
		}
		result[i] = transitPalaceOrder[fixIndex(branch-stageBranch)]
	}
	return result, true
}

func buildTransitPalaceNames(chart *ZiWeiChart, input ZiWeiDerivationInput, scope string) ([12]string, bool) {
	var result [12]string
	if chart == nil || !validZiWeiDerivationInput(scope, input) {
		return result, false
	}
	birth, ok := birthDataFromPublishedChart(chart)
	if !ok {
		return result, false
	}
	_, periodBranch, ok := derivationStemBranch(input)
	if !ok {
		return result, false
	}

	fromIndex := fixIndex(periodBranch - 2)
	if scope == "liuyue" || scope == "liuri" {
		lunarYear, err := tyme.LunarYear{}.FromYear(input.ResolvedLunarDate.Year)
		if err != nil {
			return result, false
		}
		yearGanZhi := []rune(lunarYear.GetSixtyCycle().GetName())
		if len(yearGanZhi) != 2 {
			return result, false
		}
		yearBranch, exists := BranchIndex[string(yearGanZhi[1])]
		if !exists {
			return result, false
		}
		fromIndex = fixIndex(yearBranch - 2)

		birthInput, err := buildZiWeiDerivationInput("liuri", birth.SolarYear, birth.SolarMonth, birth.SolarDay)
		if err != nil {
			return result, false
		}
		birthLeapAddition := 0
		if birthInput.ResolvedLunarDate.IsLeapMonth && birthInput.ResolvedLunarDate.Day > 15 {
			birthLeapAddition = 1
		}
		targetLeapAddition := 0
		if input.ResolvedLunarDate.IsLeapMonth && input.ResolvedLunarDate.Day > 15 {
			targetLeapAddition = 1
		}
		fromIndex = fixIndex(
			fromIndex -
				(birthInput.ResolvedLunarDate.Month + birthLeapAddition) +
				birth.HourBranch +
				(input.ResolvedLunarDate.Month + targetLeapAddition),
		)
		if scope == "liuri" {
			fromIndex = fixIndex(fromIndex + input.ResolvedLunarDate.Day - 1)
		}
	}

	var byBranch [12]string
	for palaceIndex := 0; palaceIndex < 12; palaceIndex++ {
		branchIndex := fixIndex(palaceIndex + 2)
		byBranch[branchIndex] = transitPalaceOrder[fixIndex(palaceIndex-fromIndex)]
	}
	for i, palace := range chart.Palaces {
		branchIndex, exists := BranchIndex[palace.Branch]
		if !exists {
			return [12]string{}, false
		}
		result[i] = byBranch[branchIndex]
	}
	return result, true
}
