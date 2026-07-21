package bazi

import (
	"strings"
	"testing"

	"github.com/6tail/tyme4go/tyme"
)

func TestMingGongExhaustiveAgainstTymeOwnSign(t *testing.T) {
	day := tyme.SixtyCycle{}.FromIndex(0)
	for yearIndex := 0; yearIndex < 60; yearIndex++ {
		year := tyme.SixtyCycle{}.FromIndex(yearIndex)
		for monthBranchIndex := 0; monthBranchIndex < 12; monthBranchIndex++ {
			month := tyme.SixtyCycle{}.FromIndex(monthBranchIndex)
			for hourBranchIndex := 0; hourBranchIndex < 12; hourBranchIndex++ {
				hour := tyme.SixtyCycle{}.FromIndex(hourBranchIndex)
				eightChar := tyme.EightChar{}.FromSixtyCycle(year, month, day, hour)
				want := eightChar.GetOwnSign()
				got, err := calcMingGongGanZhi(
					year.GetHeavenStem().GetName(),
					month.GetEarthBranch().GetName(),
					hour.GetEarthBranch().GetName(),
				)
				if err != nil {
					t.Fatalf("CalcMingGong(%s,%s,%s): %v",
						year.GetHeavenStem().GetName(), month.GetEarthBranch().GetName(), hour.GetEarthBranch().GetName(), err)
				}
				if got != want.GetName() {
					t.Fatalf("CalcMingGong(%s,%s,%s) = %s, tyme GetOwnSign = %s",
						year.GetHeavenStem().GetName(), month.GetEarthBranch().GetName(), hour.GetEarthBranch().GetName(), got, want.GetName())
				}
			}
		}
	}
}

func TestMingGongRejectsInvalidInputs(t *testing.T) {
	for _, tc := range []struct {
		name, yearGan, monthZhi, hourZhi, wantError string
	}{
		{name: "year stem", yearGan: "无", monthZhi: "寅", hourZhi: "子", wantError: "year stem"},
		{name: "month branch", yearGan: "甲", monthZhi: "无", hourZhi: "子", wantError: "month branch"},
		{name: "hour branch", yearGan: "甲", monthZhi: "寅", hourZhi: "子丑", wantError: "hour branch"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := calcMingGongGanZhi(tc.yearGan, tc.monthZhi, tc.hourZhi)
			if err == nil || got != "" || !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("calcMingGongGanZhi(%q,%q,%q) = %q, %v", tc.yearGan, tc.monthZhi, tc.hourZhi, got, err)
			}
		})
	}
}

func TestMingGongShenShaAllBranches(t *testing.T) {
	branches := []rune("子丑寅卯辰巳午未申酉戌亥")
	wants := []string{
		"天贵", "天厄", "天权", "天赦", "天如", "天文",
		"天福", "天驿", "天孤", "天秘", "天艺", "天寿",
	}
	for index, branch := range branches {
		if got := mingGongShenSha(string(branch)); got != wants[index] {
			t.Errorf("ming-gong shen-sha for %s = %q, want %q", string(branch), got, wants[index])
		}
	}
	for _, invalid := range []string{"", "无", "子丑"} {
		if got := mingGongShenSha(invalid); got != "" {
			t.Errorf("invalid branch %q produced ming-gong shen-sha %q", invalid, got)
		}
	}
}

func TestMingGongDetailExhaustsNaYinTable(t *testing.T) {
	for index := 0; index < 60; index++ {
		cycle := tyme.SixtyCycle{}.FromIndex(index)
		detail := buildMingGongDetail(cycle.GetName())
		wantName := canonicalNaYinAlias(cycle.GetSound().GetName())
		gotName := canonicalNaYinAlias(detail.Nayin)
		if detail.GanZhi != cycle.GetName() || detail.Gan != cycle.GetHeavenStem().GetName() ||
			detail.Zhi != cycle.GetEarthBranch().GetName() || gotName != wantName {
			t.Errorf("ming-gong detail %s = %+v, want na-yin %s", cycle.GetName(), detail, cycle.GetSound().GetName())
		}
		if detail.Nayin == "" || detail.ShenSha == "" {
			t.Errorf("ming-gong detail %s has empty fixed data: %+v", cycle.GetName(), detail)
			continue
		}
		wantElement := string([]rune(wantName)[len([]rune(wantName))-1])
		_, gotElement, ok := naYinNameAndElement(detail.Gan, detail.Zhi)
		if !ok || gotElement != wantElement {
			t.Errorf("ming-gong %s na-yin element = %s, want %s from %s", cycle.GetName(), gotElement, wantElement, detail.Nayin)
		}
	}
}

func canonicalNaYinAlias(value string) string {
	value = strings.ReplaceAll(value, "砂中", "沙中")
	return strings.ReplaceAll(value, "井泉水", "泉中水")
}
