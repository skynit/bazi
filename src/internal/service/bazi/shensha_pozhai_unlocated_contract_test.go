package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var poZhaiTargets = map[string]string{
	"子": "丑", "丑": "寅", "寅": "卯", "卯": "辰", "辰": "巳", "巳": "午",
	"午": "未", "未": "申", "申": "酉", "酉": "戌", "戌": "亥", "亥": "子",
}

func TestPoZhaiExactYearBranchTableAndRetiredNames(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		got := ruleTargetsByName(yearZhiShenShaRules[yearBranch], "破宅煞")
		if len(got) != 1 || got[0] != poZhaiTargets[yearBranch] {
			t.Errorf("year branch %s 破宅煞 targets = %v, want [%s]", yearBranch, got, poZhaiTargets[yearBranch])
		}
		for _, name := range []string{"宅煞", "飞廉", "的煞", "暗金的煞"} {
			if got := ruleTargetsByName(yearZhiShenShaRules[yearBranch], name); len(got) != 0 {
				t.Errorf("year branch %s still publishes %s: %v", yearBranch, name, got)
			}
			if got := ruleTargetsByName(monthZhiShenShaRules[yearBranch], name); len(got) != 0 {
				t.Errorf("month branch %s still publishes %s: %v", yearBranch, name, got)
			}
		}
	}
}

func TestPoZhaiFormalEntryAssignsTargetToMonthDayAndHour(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		target := poZhaiTargets[yearBranch]
		for targetIndex := 1; targetIndex < 4; targetIndex++ {
			got := calcPoZhaiFixture(t, yearBranch, target, targetIndex)
			assertOnlyPillarBucketHas(t, got, targetIndex, "破宅煞："+target)
			for _, name := range []string{"宅煞", "飞廉", "的煞", "暗金的煞"} {
				assertShenShaNameAbsentEverywhere(t, got, name)
			}
		}
	}
}

func TestPoZhaiFormalEntryRejectsEveryNonTargetBranch(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		for _, branch := range data.Zhis {
			if branch == poZhaiTargets[yearBranch] {
				continue
			}
			got := calcPoZhaiFixture(t, yearBranch, branch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "破宅煞")
		}
	}
}

func TestPoZhaiAndRetiredNameMetadata(t *testing.T) {
	meta := LookupShenShaMeta("破宅煞")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("破宅煞 metadata status = %+v", meta)
	}
	for _, fragment := range []string{"生年支", "命后一辰", "逐柱落位", "子丑、丑寅", "《渊海子平》PDF第688页"} {
		if !strings.Contains(meta.Basis, fragment) {
			t.Errorf("破宅煞 basis = %q, want %q", meta.Basis, fragment)
		}
	}
	for _, name := range []string{"宅煞", "飞廉", "的煞", "暗金的煞"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
}

func calcPoZhaiFixture(t testing.TB, yearBranch, placedBranch string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	target := poZhaiTargets[yearBranch]
	pillars := []model.Pillar{
		poZhaiPillarForBranch(t, yearBranch),
		poZhaiNeutralPillar(t, target, 10),
		poZhaiNeutralPillar(t, target, 20),
		poZhaiNeutralPillar(t, target, 30),
	}
	pillars[targetIndex] = poZhaiPillarForBranch(t, placedBranch)
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func poZhaiPillarForBranch(t testing.TB, branch string) model.Pillar {
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

func poZhaiNeutralPillar(t testing.TB, avoidBranch string, start int) model.Pillar {
	t.Helper()
	for offset := 0; offset < 60; offset++ {
		i := (start + offset) % 60
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Zhi != avoidBranch {
			return pillar
		}
	}
	t.Fatalf("no neutral pillar avoiding %s", avoidBranch)
	return model.Pillar{}
}
