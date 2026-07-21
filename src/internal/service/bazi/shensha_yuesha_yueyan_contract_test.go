package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

var classicalMonthBranchTargets = map[string]map[string]string{
	"月厌": {
		"寅": "戌", "卯": "酉", "辰": "申", "巳": "未", "午": "午", "未": "巳",
		"申": "辰", "酉": "卯", "戌": "寅", "亥": "丑", "子": "子", "丑": "亥",
	},
	"月煞": {
		"寅": "丑", "午": "丑", "戌": "丑",
		"亥": "戌", "卯": "戌", "未": "戌",
		"申": "未", "子": "未", "辰": "未",
		"巳": "辰", "酉": "辰", "丑": "辰",
	},
}

func TestYueShaAndYueYanUseLocatedTwelveMonthTables(t *testing.T) {
	for name, wants := range classicalMonthBranchTargets {
		if len(wants) != 12 {
			t.Fatalf("%s expected table size = %d, want 12", name, len(wants))
		}
		for _, monthZhi := range data.Zhis {
			got := monthRuleTarget(t, monthZhi, name)
			if got != wants[monthZhi] {
				t.Errorf("%s month %s target = %s, want %s", name, monthZhi, got, wants[monthZhi])
			}
			for _, candidate := range data.Zhis {
				if gotHit := candidate == got; gotHit != (candidate == wants[monthZhi]) {
					t.Errorf("%s month %s candidate %s truth mismatch", name, monthZhi, candidate)
				}
			}
		}
	}
}

func TestYueShaAndYueYanFormalEntryAssignsEveryTargetPillar(t *testing.T) {
	for name, wants := range classicalMonthBranchTargets {
		for _, monthZhi := range data.Zhis {
			target := wants[monthZhi]
			for _, targetIndex := range []int{0, 2, 3} {
				pillars := []model.Pillar{
					monthRuleNeutralPillar(t, target, 1),
					poZhaiPillarForBranch(t, monthZhi),
					monthRuleNeutralPillar(t, target, 11),
					monthRuleNeutralPillar(t, target, 21),
				}
				pillars[targetIndex] = poZhaiPillarForBranch(t, target)
				got, err := CalcShenShaByPillars(ShenShaPillars{
					Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
				})
				if err != nil {
					t.Fatal(err)
				}
				for i, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour} {
					wantHit := i == targetIndex || (i == 1 && monthZhi == target)
					if gotHit := hasShenShaName(bucket, name); gotHit != wantHit {
						t.Errorf("%s month %s target %s bucket %d hit = %v, want %v: %+v", name, monthZhi, target, i, gotHit, wantHit, got)
					}
				}
				if hasShenShaName(got.Global, name) {
					t.Errorf("%s leaked into global bucket: %+v", name, got)
				}
			}
		}
	}
}

func TestYueShaAndYueYanMetadataUsesLocatedPage(t *testing.T) {
	for _, name := range []string{"月厌", "月煞"} {
		meta := LookupShenShaMeta(name)
		for _, want := range []string{"《三命通会》PDF第103页", "书内第100页", "逐柱匹配目标支"} {
			if !strings.Contains(meta.Basis, want) {
				t.Errorf("%s metadata basis %q missing %q", name, meta.Basis, want)
			}
		}
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata = %+v, want observed/not_adjudicated", name, meta)
		}
	}
}

func monthRuleTarget(t testing.TB, monthZhi, name string) string {
	t.Helper()
	target := ""
	for _, rule := range monthZhiShenShaRules[monthZhi] {
		if rule.Name != name {
			continue
		}
		if target != "" {
			t.Fatalf("month %s has duplicate %s rules", monthZhi, name)
		}
		target = rule.Target
	}
	if target == "" {
		t.Fatalf("month %s missing %s rule", monthZhi, name)
	}
	return target
}

func monthRuleNeutralPillar(t testing.TB, avoidBranch string, start int) model.Pillar {
	t.Helper()
	for offset := range 60 {
		i := (start + offset) % 60
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Zhi != avoidBranch {
			return pillar
		}
	}
	t.Fatalf("no neutral pillar avoiding branch %s", avoidBranch)
	return model.Pillar{}
}
