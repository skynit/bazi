package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var guChenTargets = map[string]string{
	"亥": "寅", "子": "寅", "丑": "寅",
	"寅": "巳", "卯": "巳", "辰": "巳",
	"巳": "申", "午": "申", "未": "申",
	"申": "亥", "酉": "亥", "戌": "亥",
}

var guaSuTargets = map[string]string{
	"亥": "戌", "子": "戌", "丑": "戌",
	"寅": "丑", "卯": "丑", "辰": "丑",
	"巳": "辰", "午": "辰", "未": "辰",
	"申": "未", "酉": "未", "戌": "未",
}

func TestGuChenAndGuaSuExactBirthYearTables(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		if got := guChenBySanHui[yearBranch]; got != guChenTargets[yearBranch] {
			t.Errorf("year branch %s 孤辰=%s, want %s", yearBranch, got, guChenTargets[yearBranch])
		}
		if got := guSuBySanHui[yearBranch]; got != guaSuTargets[yearBranch] {
			t.Errorf("year branch %s 寡宿=%s, want %s", yearBranch, got, guaSuTargets[yearBranch])
		}
	}
}

func TestGuChenAndGuaSuAttachToEveryMatchingPillarFromBirthYear(t *testing.T) {
	for name, targets := range map[string]map[string]string{"孤辰": guChenTargets, "寡宿": guaSuTargets} {
		for _, yearBranch := range data.Zhis {
			target := targets[yearBranch]
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year:   poZhaiPillarForBranch(t, yearBranch),
				Month:  poZhaiPillarForBranch(t, target),
				Day:    poZhaiPillarForBranch(t, target),
				Hour:   poZhaiPillarForBranch(t, target),
				Gender: model.GenderMale,
			})
			if err != nil {
				t.Fatal(err)
			}
			for _, bucket := range [][]string{got.Month, got.Day, got.Hour} {
				assertExactShenShaHitCount(t, bucket, name+"："+target, 1)
			}
			assertExactShenShaHitCount(t, got.Year, name+"："+target, 0)
		}
	}
}

func TestGuChenAndGuaSuDoNotUseDayBranchAsLookupKey(t *testing.T) {
	for name, targets := range map[string]map[string]string{"孤辰": guChenTargets, "寡宿": guaSuTargets} {
		for _, dayBranch := range data.Zhis {
			target := targets[dayBranch]
			yearBranch := guChenGuaSuYearWithDifferentTarget(targets, target)
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year:   poZhaiPillarForBranch(t, yearBranch),
				Month:  poZhaiNeutralPillar(t, target, 10),
				Day:    poZhaiPillarForBranch(t, dayBranch),
				Hour:   poZhaiPillarForBranch(t, target),
				Gender: model.GenderMale,
			})
			if err != nil {
				t.Fatal(err)
			}
			assertExactShenShaHitCount(t, got.Hour, name+"："+target, 0)
		}
	}
}

func TestGuGuaDoubleHitDoesNotPublishSyntheticAlias(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(source), "\"孤寡煞\"") {
		t.Fatal("production source still publishes synthetic 孤寡煞 alias")
	}

	for _, yearBranch := range data.Zhis {
		got, err := CalcShenShaByPillars(ShenShaPillars{
			Year:   poZhaiPillarForBranch(t, yearBranch),
			Month:  poZhaiPillarForBranch(t, guChenTargets[yearBranch]),
			Day:    poZhaiPillarForBranch(t, guaSuTargets[yearBranch]),
			Hour:   poZhaiNeutralPillar(t, guChenTargets[yearBranch], 30),
			Gender: model.GenderMale,
		})
		if err != nil {
			t.Fatal(err)
		}
		assertShenShaNameAbsentEverywhere(t, got, "孤寡煞")
	}

	meta := LookupShenShaMeta("孤寡煞")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
		t.Fatalf("孤寡煞 metadata = %+v, want unregistered/not_available", meta)
	}
}

func TestGuChenAndGuaSuMetadata(t *testing.T) {
	want := map[string][]string{
		"孤辰": {"只以生年支", "进前一辰", "逐柱落位", "《三命通会》PDF第118页", "《渊海子平》PDF第632、744页", "不生成亲属、婚姻或现实事件结论"},
		"寡宿": {"只以生年支", "退后一辰", "逐柱落位", "《三命通会》PDF第118页", "《渊海子平》PDF第632、744页", "不生成亲属、婚姻或现实事件结论"},
	}
	for name, fragments := range want {
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
}

func guChenGuaSuYearWithDifferentTarget(targets map[string]string, target string) string {
	for _, branch := range data.Zhis {
		if targets[branch] != target {
			return branch
		}
	}
	return ""
}
