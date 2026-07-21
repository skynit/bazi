package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var tianYiTargets = map[string]string{
	"甲": "丑未", "乙": "子申", "丙": "亥酉", "丁": "亥酉", "戊": "丑未",
	"己": "子申", "庚": "丑未", "辛": "寅午", "壬": "卯巳", "癸": "卯巳",
}

func TestTianYiExactTableUsesDayStemOnly(t *testing.T) {
	for _, dayGan := range data.Gans {
		want := tianYiTargets[dayGan]
		got := ruleTargetsByName(dayGanShenShaRules[dayGan], "天乙贵人")
		if len(got) != 1 || got[0] != want {
			t.Errorf("day stem %s 天乙贵人 targets = %v, want [%s]", dayGan, got, want)
		}
		if got := ruleTargetsByName(yearGanShenShaRules[dayGan], "天乙贵人"); len(got) != 0 {
			t.Errorf("year stem %s still publishes 天乙贵人 shortcut: %v", dayGan, got)
		}
		for _, branch := range data.Zhis {
			wantMatch := strings.Contains(want, branch)
			if got := targetContainsZhi(want, branch); got != wantMatch {
				t.Errorf("day stem %s branch %s match = %v, want %v", dayGan, branch, got, wantMatch)
			}
		}
	}
}

func TestTianYiFormalEntryAssignsEveryTargetToActualPillar(t *testing.T) {
	for _, dayGan := range data.Gans {
		for _, branch := range data.Zhis {
			if !strings.Contains(tianYiTargets[dayGan], branch) {
				continue
			}
			for targetIndex := 0; targetIndex < 4; targetIndex++ {
				if targetIndex == 2 && sixtyCycleIndex(dayGan, branch) < 0 {
					continue
				}
				got := calcTianYiFixture(t, dayGan, branch, targetIndex)
				assertOnlyPillarBucketHas(t, got, targetIndex, "天乙贵人："+tianYiTargets[dayGan])
				assertShenShaNameAbsentEverywhere(t, got, "时贵")
			}
		}
	}
}

func TestTianYiFormalEntryRejectsEveryNonTargetBranch(t *testing.T) {
	for _, dayGan := range data.Gans {
		for _, branch := range data.Zhis {
			if strings.Contains(tianYiTargets[dayGan], branch) {
				continue
			}
			got := calcTianYiFixture(t, dayGan, branch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "天乙贵人")
			assertShenShaNameAbsentEverywhere(t, got, "时贵")
		}
	}
}

func TestTianYiDoesNotUseYearStemAsRuleKey(t *testing.T) {
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year:   model.Pillar{Gan: "甲", Zhi: "辰"},
		Month:  model.Pillar{Gan: "乙", Zhi: "丑"},
		Day:    model.Pillar{Gan: "乙", Zhi: "亥"},
		Hour:   model.Pillar{Gan: "丙", Zhi: "戌"},
		Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	assertShenShaNameAbsentEverywhere(t, got, "天乙贵人")
}

func TestTianYiMetadataRecordsSelectedAndDeferredProfiles(t *testing.T) {
	meta := LookupShenShaMeta("天乙贵人")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("天乙贵人 metadata status = %+v", meta)
	}
	for _, fragment := range []string{
		"当前Profile只按日干", "逐柱落位", "甲戊庚丑未、乙己子申、丙丁亥酉、辛寅午、壬癸卯巳",
		"《三命通会》PDF第97-98页", "《渊海子平》PDF第63页", "庚与辛同列寅午",
		"昼夜贵", "冬夏至分治", "年时互换贵均未实现",
	} {
		if !strings.Contains(meta.Basis, fragment) {
			t.Errorf("天乙贵人 basis = %q, want %q", meta.Basis, fragment)
		}
	}
	alias := LookupShenShaMeta("时贵")
	if alias.Status != "unregistered" || alias.InterpretationStatus != "not_available" {
		t.Errorf("obsolete 时贵 metadata = %+v", alias)
	}
}

func calcTianYiFixture(t testing.TB, dayGan, placedBranch string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	targets := tianYiTargets[dayGan]
	pillars := []model.Pillar{
		tianYiNeutralPillar(t, targets, 0),
		tianYiNeutralPillar(t, targets, 10),
		tianYiDayPillar(t, dayGan, targets),
		tianYiNeutralPillar(t, targets, 30),
	}
	if targetIndex == 2 {
		pillars[2] = model.Pillar{Gan: dayGan, Zhi: placedBranch}
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

func tianYiDayPillar(t testing.TB, dayGan, avoidBranches string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Gan == dayGan && !strings.Contains(avoidBranches, pillar.Zhi) {
			return pillar
		}
	}
	t.Fatalf("no day pillar for stem %s avoiding %s", dayGan, avoidBranches)
	return model.Pillar{}
}

func tianYiNeutralPillar(t testing.TB, avoidBranches string, start int) model.Pillar {
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
