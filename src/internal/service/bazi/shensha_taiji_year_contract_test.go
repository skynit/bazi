package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var taiJiTargets = map[string]string{
	"甲": "子午", "乙": "子午", "丙": "卯酉", "丁": "卯酉",
	"戊": "辰戌丑未", "己": "辰戌丑未", "庚": "寅亥", "辛": "寅亥",
	"壬": "巳申", "癸": "巳申",
}

func TestTaiJiExactTableUsesYearStemOnly(t *testing.T) {
	for _, yearGan := range data.Gans {
		want := taiJiTargets[yearGan]
		got := ruleTargetsByName(yearGanShenShaRules[yearGan], "太极贵人")
		if len(got) != 1 || got[0] != want {
			t.Errorf("year stem %s 太极贵人 targets = %v, want [%s]", yearGan, got, want)
		}
		if got := ruleTargetsByName(dayGanShenShaRules[yearGan], "太极贵人"); len(got) != 0 {
			t.Errorf("day stem %s still publishes 太极贵人 shortcut: %v", yearGan, got)
		}
		for _, branch := range data.Zhis {
			if got := targetContainsZhi(want, branch); got != strings.Contains(want, branch) {
				t.Errorf("year stem %s branch %s match = %v", yearGan, branch, got)
			}
		}
	}
}

func TestTaiJiFormalEntryAssignsEveryTargetToActualPillar(t *testing.T) {
	for _, yearGan := range data.Gans {
		for _, branch := range data.Zhis {
			if !strings.Contains(taiJiTargets[yearGan], branch) {
				continue
			}
			for targetIndex := 0; targetIndex < 4; targetIndex++ {
				if targetIndex == 0 && sixtyCycleIndex(yearGan, branch) < 0 {
					continue
				}
				got := calcTaiJiFixture(t, yearGan, branch, targetIndex)
				assertOnlyPillarBucketHas(t, got, targetIndex, "太极贵人："+taiJiTargets[yearGan])
				if hasShenShaName(got.Global, "太极贵人") {
					t.Errorf("year stem %s target %s leaked 太极贵人 to global: %+v", yearGan, branch, got)
				}
			}
		}
	}
}

func TestTaiJiFormalEntryRejectsEveryNonTargetBranch(t *testing.T) {
	for _, yearGan := range data.Gans {
		for _, branch := range data.Zhis {
			if strings.Contains(taiJiTargets[yearGan], branch) {
				continue
			}
			got := calcTaiJiFixture(t, yearGan, branch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "太极贵人")
		}
	}
}

func TestTaiJiDoesNotUseDayStemAsRuleKey(t *testing.T) {
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year:   model.Pillar{Gan: "甲", Zhi: "辰"},
		Month:  model.Pillar{Gan: "乙", Zhi: "丑"},
		Day:    model.Pillar{Gan: "庚", Zhi: "寅"},
		Hour:   model.Pillar{Gan: "丙", Zhi: "戌"},
		Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertShenShaNameAbsentEverywhere(t, got, "太极贵人")
}

func TestTaiJiFormerWrongAndMissingTargets(t *testing.T) {
	for _, tc := range []struct {
		yearGan string
		branch  string
		want    bool
	}{
		{yearGan: "戊", branch: "丑", want: true},
		{yearGan: "己", branch: "未", want: true},
		{yearGan: "庚", branch: "亥", want: true},
		{yearGan: "辛", branch: "申", want: false},
		{yearGan: "壬", branch: "申", want: true},
		{yearGan: "癸", branch: "亥", want: false},
	} {
		got := calcTaiJiFixture(t, tc.yearGan, tc.branch, 1)
		if tc.want {
			assertOnlyPillarBucketHas(t, got, 1, "太极贵人："+taiJiTargets[tc.yearGan])
		} else {
			assertShenShaNameAbsentEverywhere(t, got, "太极贵人")
		}
	}
}

func TestTaiJiMetadataRecordsProfileAndSourceDifference(t *testing.T) {
	meta := LookupShenShaMeta("太极贵人")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("太极贵人 metadata status = %+v", meta)
	}
	for _, fragment := range []string{
		"只以生年干", "逐柱落位", "甲乙子午、丙丁卯酉、戊己辰戌丑未、庚辛寅亥、壬癸巳申",
		"《渊海子平》PDF第67页", "取别干则非也", "《三命通会》PDF第104-105页", "当前Profile不混入申",
	} {
		if !strings.Contains(meta.Basis, fragment) {
			t.Errorf("太极贵人 basis = %q, want %q", meta.Basis, fragment)
		}
	}
}

func calcTaiJiFixture(t testing.TB, yearGan, placedBranch string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	targets := taiJiTargets[yearGan]
	pillars := []model.Pillar{
		taiJiYearPillar(t, yearGan, targets),
		taiJiNeutralPillar(t, targets, 10),
		taiJiNeutralPillar(t, targets, 20),
		taiJiNeutralPillar(t, targets, 30),
	}
	if targetIndex == 0 {
		pillars[0] = model.Pillar{Gan: yearGan, Zhi: placedBranch}
	} else {
		pillars[targetIndex] = taiJiPillarForBranch(t, placedBranch)
	}
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func taiJiYearPillar(t testing.TB, yearGan, avoidBranches string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Gan == yearGan && !strings.Contains(avoidBranches, pillar.Zhi) {
			return pillar
		}
	}
	t.Fatalf("no year pillar for stem %s avoiding %s", yearGan, avoidBranches)
	return model.Pillar{}
}

func taiJiNeutralPillar(t testing.TB, avoidBranches string, start int) model.Pillar {
	t.Helper()
	for offset := 0; offset < 60; offset++ {
		i := (start + offset) % 60
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if !strings.Contains(avoidBranches, pillar.Zhi) {
			return pillar
		}
	}
	t.Fatalf("no neutral pillar avoiding %s", avoidBranches)
	return model.Pillar{}
}

func taiJiPillarForBranch(t testing.TB, branch string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Zhi == branch {
			return pillar
		}
	}
	t.Fatalf("no sixty-cycle pillar for branch %s", branch)
	return model.Pillar{}
}
