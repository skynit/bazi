package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var yearBranchLocatedTargets = map[string]map[string]string{
	"官符": {
		"子": "辰", "丑": "巳", "寅": "午", "卯": "未", "辰": "申", "巳": "酉",
		"午": "戌", "未": "亥", "申": "子", "酉": "丑", "戌": "寅", "亥": "卯",
	},
	"病符": {
		"子": "亥", "丑": "子", "寅": "丑", "卯": "寅", "辰": "卯", "巳": "辰",
		"午": "巳", "未": "午", "申": "未", "酉": "申", "戌": "酉", "亥": "戌",
	},
	"丧门": {
		"子": "寅", "丑": "卯", "寅": "辰", "卯": "巳", "辰": "午", "巳": "未",
		"午": "申", "未": "酉", "申": "戌", "酉": "亥", "戌": "子", "亥": "丑",
	},
}

func TestLocatedYearBranchRulesAndBaiHuFailClosed(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		for name, targets := range yearBranchLocatedTargets {
			got := ruleTargetsByName(yearZhiShenShaRules[yearBranch], name)
			if len(got) != 1 || got[0] != targets[yearBranch] {
				t.Errorf("year branch %s %s=%v, want [%s]", yearBranch, name, got, targets[yearBranch])
			}
		}
		if got := ruleTargetsByName(yearZhiShenShaRules[yearBranch], "白虎"); len(got) != 0 {
			t.Errorf("year branch %s still publishes 白虎: %v", yearBranch, got)
		}
	}
}

func TestLocatedYearBranchRulesAssignOnlyTheTargetPillar(t *testing.T) {
	for name, targets := range yearBranchLocatedTargets {
		for _, yearBranch := range data.Zhis {
			target := targets[yearBranch]
			for targetIndex := 1; targetIndex < 4; targetIndex++ {
				got := calcLocatedYearBranchFixture(t, yearBranch, target, targetIndex)
				assertOnlyPillarBucketHas(t, got, targetIndex, name+"："+target)
				assertShenShaNameAbsentEverywhere(t, got, "白虎")
			}
		}
	}
}

func TestBaiHuRemainsAbsentAcrossFormerYearBranchSearchSpace(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		for _, branch := range data.Zhis {
			got := calcPoZhaiFixture(t, yearBranch, branch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "白虎")
		}
	}
}

func TestLocatedYearBranchRuleMetadataAndBaiHuMetadata(t *testing.T) {
	wantFragments := map[string][]string{
		"官符": {"生年支", "太岁前五辰", "逐柱落位", "十二宫含本位计数", "《三命通会》PDF第122页"},
		"病符": {"生年支", "太岁后一辰", "逐柱落位", "《三命通会》PDF第122页", "不生成疾病或健康结论"},
		"丧门": {"生年支", "命前二辰", "逐柱落位", "《三命通会》PDF第122页", "《渊海子平》PDF第631-633页"},
	}
	for name, fragments := range wantFragments {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata = %+v", name, meta)
		}
		for _, fragment := range fragments {
			if !strings.Contains(meta.Basis, fragment) {
				t.Errorf("%s basis = %q, want %q", name, meta.Basis, fragment)
			}
		}
	}

	meta := LookupShenShaMeta("白虎")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
		t.Fatalf("白虎 metadata = %+v, want unregistered/not_available", meta)
	}
}

func calcLocatedYearBranchFixture(t testing.TB, yearBranch, target string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	pillars := []model.Pillar{
		poZhaiPillarForBranch(t, yearBranch),
		poZhaiNeutralPillar(t, target, 10),
		poZhaiNeutralPillar(t, target, 20),
		poZhaiNeutralPillar(t, target, 30),
	}
	pillars[targetIndex] = poZhaiPillarForBranch(t, target)
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}
