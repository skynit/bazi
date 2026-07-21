package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var jinYuTargets = map[string]string{
	"甲": "辰", "乙": "巳", "丙": "未", "丁": "申", "戊": "未",
	"己": "申", "庚": "戌", "辛": "亥", "壬": "丑", "癸": "寅",
}

var guoYinTargets = map[string]string{
	"甲": "戌", "乙": "亥", "丙": "丑", "丁": "寅", "戊": "丑",
	"己": "寅", "庚": "辰", "辛": "巳", "壬": "未", "癸": "申",
}

func TestJinYuTableEqualsDayStemLuPlusTwoBranches(t *testing.T) {
	for _, dayGan := range data.Gans {
		lu, _ := luBranchForStem(dayGan)
		want := data.Zhis[(data.ZhiIndex(lu)+2)%len(data.Zhis)]
		if want != jinYuTargets[dayGan] {
			t.Fatalf("day stem %s derived 金舆 = %s, fixture wants %s", dayGan, want, jinYuTargets[dayGan])
		}
		got := ruleTargetsByName(dayGanShenShaRules[dayGan], "金舆")
		if len(got) != 1 || got[0] != want {
			t.Errorf("day stem %s 金舆 targets = %v, want [%s]", dayGan, got, want)
		}
	}
}

func TestJinYuFormalEntryCoversEveryTargetAndRejectsEveryOtherBranch(t *testing.T) {
	for _, dayGan := range data.Gans {
		target := jinYuTargets[dayGan]
		for targetIndex := 0; targetIndex < 4; targetIndex++ {
			if targetIndex == 2 && sixtyCycleIndex(dayGan, target) < 0 {
				continue
			}
			got := calcHongYanFixture(t, dayGan, target, target, targetIndex)
			assertOnlyPillarBucketHas(t, got, targetIndex, "金舆："+target)
			if hasShenShaName(got.Global, "金舆") {
				t.Errorf("day stem %s target %s leaked 金舆 to global: %+v", dayGan, target, got)
			}
		}
		for _, branch := range data.Zhis {
			if branch == target {
				continue
			}
			got := calcHongYanFixture(t, dayGan, target, branch, 0)
			assertShenShaNameAbsentEverywhere(t, got, "金舆")
		}
	}
}

func TestGuoYinTableUsesYearStemLuPlusEightBranches(t *testing.T) {
	for _, yearGan := range data.Gans {
		lu, _ := luBranchForStem(yearGan)
		want := data.Zhis[(data.ZhiIndex(lu)+8)%len(data.Zhis)]
		if want != guoYinTargets[yearGan] {
			t.Fatalf("year stem %s derived 国印贵人 = %s, fixture wants %s", yearGan, want, guoYinTargets[yearGan])
		}
		got := ruleTargetsByName(yearGanShenShaRules[yearGan], "国印贵人")
		if len(got) != 1 || got[0] != want {
			t.Errorf("year stem %s 国印贵人 targets = %v, want [%s]", yearGan, got, want)
		}
		if got := ruleTargetsByName(dayGanShenShaRules[yearGan], "国印贵人"); len(got) != 0 {
			t.Errorf("day stem %s still publishes 国印贵人 shortcut: %v", yearGan, got)
		}
	}
}

func TestGuoYinFormalEntryUsesYearStemAndAssignsActualPillar(t *testing.T) {
	for _, yearGan := range data.Gans {
		target := guoYinTargets[yearGan]
		for targetIndex := 0; targetIndex < 4; targetIndex++ {
			if targetIndex == 0 && sixtyCycleIndex(yearGan, target) < 0 {
				continue
			}
			got := calcGuoYinFixture(t, yearGan, target, target, targetIndex)
			assertOnlyPillarBucketHas(t, got, targetIndex, "国印贵人："+target)
			if hasShenShaName(got.Global, "国印贵人") {
				t.Errorf("year stem %s target %s leaked 国印贵人 to global: %+v", yearGan, target, got)
			}
		}
		for _, branch := range data.Zhis {
			if branch == target {
				continue
			}
			got := calcGuoYinFixture(t, yearGan, target, branch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "国印贵人")
		}
	}
}

func TestRetiredFuXingSingleBranchProfileIsUnavailable(t *testing.T) {
	retiredTargets := map[string]string{
		"甲": "寅", "乙": "丑", "丙": "寅", "丁": "亥", "戊": "申",
		"己": "未", "庚": "午", "辛": "巳", "壬": "辰", "癸": "卯",
	}
	for _, dayGan := range data.Gans {
		if got := ruleTargetsByName(dayGanShenShaRules[dayGan], "福星贵人"); len(got) != 0 {
			t.Errorf("day stem %s still publishes retired 福星贵人 shortcut: %v", dayGan, got)
		}
		for _, branch := range data.Zhis {
			got := calcHongYanFixture(t, dayGan, retiredTargets[dayGan], branch, 0)
			assertShenShaNameAbsentEverywhere(t, got, "福星贵人")
		}
	}
	meta := LookupShenShaMeta("福星贵人")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" {
		t.Fatalf("retired 福星贵人 metadata = %+v", meta)
	}
}

func TestJinYuAndGuoYinMetadataIsLocatedButNotAdjudicated(t *testing.T) {
	wants := map[string][]string{
		"金舆":   {"日干禄位顺数第二支", "甲辰、乙巳、丙戊未、丁己申、庚戌、辛亥、壬丑、癸寅", "《三命通会》PDF第91页", "《渊海子平》PDF第93页"},
		"国印贵人": {"年干禄宫顺数第九位", "甲戌、乙亥、丙丑、丁寅、戊丑、己寅、庚辰、辛巳、壬未、癸申", "《渊海子平》PDF第733页"},
	}
	for name, fragments := range wants {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata status = %+v", name, meta)
		}
		for _, fragment := range append(fragments, "逐柱落位") {
			if !strings.Contains(meta.Basis, fragment) {
				t.Errorf("%s basis = %q, want %q", name, meta.Basis, fragment)
			}
		}
	}
}

func ruleTargetsByName(rules []shenShaRule, name string) []string {
	targets := make([]string, 0, 1)
	for _, rule := range rules {
		if rule.Name == name {
			targets = append(targets, rule.Target)
		}
	}
	return targets
}

func calcGuoYinFixture(t testing.TB, yearGan, target, placedBranch string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	pillars := []model.Pillar{
		guoYinYearPillar(t, yearGan, target),
		hongYanNeutralPillar(t, target, 10),
		hongYanNeutralPillar(t, target, 20),
		hongYanNeutralPillar(t, target, 30),
	}
	if targetIndex == 0 {
		pillars[0] = model.Pillar{Gan: yearGan, Zhi: placedBranch}
		if sixtyCycleIndex(pillars[0].Gan, pillars[0].Zhi) < 0 {
			t.Fatalf("invalid year pillar fixture %s%s", yearGan, placedBranch)
		}
	} else {
		pillars[targetIndex] = hongYanPillarForBranch(t, placedBranch)
	}
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func guoYinYearPillar(t testing.TB, yearGan, avoidBranch string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Gan == yearGan && pillar.Zhi != avoidBranch {
			return pillar
		}
	}
	t.Fatalf("no year pillar for stem %s avoiding branch %s", yearGan, avoidBranch)
	return model.Pillar{}
}
