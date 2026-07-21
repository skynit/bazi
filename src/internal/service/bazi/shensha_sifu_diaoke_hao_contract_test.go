package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var diaoKeTargets = map[string]string{
	"子": "戌", "丑": "亥", "寅": "子", "卯": "丑", "辰": "寅", "巳": "卯",
	"午": "辰", "未": "巳", "申": "午", "酉": "未", "戌": "申", "亥": "酉",
}

func TestSiFuAndDiaoKeExactTablesWhileHaoFailsClosed(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		bingFu := ruleTargetsByName(yearZhiShenShaRules[yearBranch], "病符")
		siFu := ruleTargetsByName(yearZhiShenShaRules[yearBranch], "死符")
		if len(bingFu) != 1 || len(siFu) != 1 || zhiLiuChong[bingFu[0]] != siFu[0] {
			t.Errorf("year branch %s 病符=%v 死符=%v, want exact clash", yearBranch, bingFu, siFu)
		}
		diaoKe := ruleTargetsByName(yearZhiShenShaRules[yearBranch], "吊客")
		if len(diaoKe) != 1 || diaoKe[0] != diaoKeTargets[yearBranch] {
			t.Errorf("year branch %s 吊客=%v, want [%s]", yearBranch, diaoKe, diaoKeTargets[yearBranch])
		}
		for _, name := range []string{"大耗", "小耗"} {
			if got := ruleTargetsByName(yearZhiShenShaRules[yearBranch], name); len(got) != 0 {
				t.Errorf("year branch %s still publishes %s: %v", yearBranch, name, got)
			}
		}
	}
}

func TestDiaoKeFormalEntryAssignsTargetToMonthDayAndHour(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		target := diaoKeTargets[yearBranch]
		for targetIndex := 1; targetIndex < 4; targetIndex++ {
			got := calcDiaoKeFixture(t, yearBranch, target, targetIndex)
			assertOnlyPillarBucketHas(t, got, targetIndex, "吊客："+target)
			assertShenShaNameAbsentEverywhere(t, got, "死符")
			assertShenShaNameAbsentEverywhere(t, got, "大耗")
			assertShenShaNameAbsentEverywhere(t, got, "小耗")
		}
	}
}

func TestHaoNamesRemainAbsentAcrossFormerYearBranchSearchSpace(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		for _, branch := range data.Zhis {
			got := calcPoZhaiFixture(t, yearBranch, branch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "大耗")
			assertShenShaNameAbsentEverywhere(t, got, "小耗")
		}
	}
}

func TestSiFuDiaoKeAndHaoMetadata(t *testing.T) {
	siFu := LookupShenShaMeta("死符")
	if siFu.Status != "observed" || siFu.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("死符 metadata = %+v", siFu)
	}
	for _, fragment := range []string{"生年支", "病符目标的对冲支", "《三命通会》PDF第122页", "正式输出层屏蔽"} {
		if !strings.Contains(siFu.Basis, fragment) {
			t.Errorf("死符 basis = %q, want %q", siFu.Basis, fragment)
		}
	}
	diaoKe := LookupShenShaMeta("吊客")
	if diaoKe.Status != "observed" || diaoKe.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("吊客 metadata = %+v", diaoKe)
	}
	for _, fragment := range []string{"生年支", "命后二辰", "逐柱落位", "《三命通会》PDF第122页"} {
		if !strings.Contains(diaoKe.Basis, fragment) {
			t.Errorf("吊客 basis = %q, want %q", diaoKe.Basis, fragment)
		}
	}
	for _, name := range []string{"大耗", "小耗"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v, want unregistered/not_available", name, meta)
		}
	}
}

func calcDiaoKeFixture(t testing.TB, yearBranch, placedBranch string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	target := diaoKeTargets[yearBranch]
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
