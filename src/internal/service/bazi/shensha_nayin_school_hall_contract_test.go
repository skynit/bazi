package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestNayinSchoolHallRuleTableUsesSameElementBirthAndTargetPillars(t *testing.T) {
	wants := map[string]nayinSchoolHallRule{
		"金": {XueTang: "辛巳", CiGuan: "壬申"},
		"木": {XueTang: "己亥", CiGuan: "庚寅"},
		"水": {XueTang: "甲申", CiGuan: "癸亥"},
		"火": {XueTang: "丙寅", CiGuan: "乙巳"},
		"土": {XueTang: "戊申", CiGuan: "丁亥"},
	}
	if len(nayinSchoolHallRules) != len(wants) {
		t.Fatalf("school-hall element rules = %d, want %d", len(nayinSchoolHallRules), len(wants))
	}
	for element, want := range wants {
		got, ok := nayinSchoolHallRules[element]
		if !ok || got != want {
			t.Errorf("element %s school-hall rule = %+v, want %+v", element, got, want)
		}
		for _, target := range []string{got.XueTang, got.CiGuan} {
			pillar := parseGongTestPillar(t, target)
			nayin := data.Nayin[data.GanIndex(pillar.Gan)][data.ZhiIndex(pillar.Zhi)]
			if gotElement := data.NaYinMap[nayin].Element; gotElement != element {
				t.Errorf("target %s nayin element = %s, want %s", target, gotElement, element)
			}
			for i := 0; i < 60; i++ {
				candidate := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
				wantMatch := candidate.Gan+candidate.Zhi == target
				if gotMatch := matchGanZhiTarget(target, candidate); gotMatch != wantMatch {
					t.Errorf("target %s against %s%s = %v, want %v", target, candidate.Gan, candidate.Zhi, gotMatch, wantMatch)
				}
			}
		}
	}
}

func TestNayinSchoolHallFormalEntryAssignsExactTargetsToActualPillars(t *testing.T) {
	for element, rule := range nayinSchoolHallRules {
		for _, tc := range []struct {
			name   string
			target string
		}{
			{name: "学堂", target: rule.XueTang},
			{name: "词馆", target: rule.CiGuan},
		} {
			for targetIndex := 0; targetIndex < 4; targetIndex++ {
				got := calcNayinSchoolHallFixture(t, element, rule, parseGongTestPillar(t, tc.target), targetIndex)
				assertOnlyPillarBucketHas(t, got, targetIndex, tc.name+"："+tc.target)
				if hasShenShaName(got.Global, tc.name) {
					t.Errorf("element %s %s leaked to global: %+v", element, tc.name, got)
				}
			}
		}
	}
}

func TestNayinSchoolHallRejectsSameBranchWithWrongStem(t *testing.T) {
	for element, rule := range nayinSchoolHallRules {
		for _, tc := range []struct {
			name   string
			target string
		}{
			{name: "学堂", target: rule.XueTang},
			{name: "词馆", target: rule.CiGuan},
		} {
			wrong := schoolHallPillarWithSameBranch(t, tc.target)
			got := calcNayinSchoolHallFixture(t, element, rule, wrong, 1)
			assertShenShaNameAbsentEverywhere(t, got, tc.name)
		}
	}
}

func TestNayinSchoolHallRejectsCorrectTargetUnderWrongBirthElement(t *testing.T) {
	for element, rule := range nayinSchoolHallRules {
		wrongElement := "金"
		if element == wrongElement {
			wrongElement = "木"
		}
		wrongRule := nayinSchoolHallRules[wrongElement]
		for _, tc := range []struct {
			name   string
			target string
		}{
			{name: "学堂", target: rule.XueTang},
			{name: "词馆", target: rule.CiGuan},
		} {
			got := calcNayinSchoolHallFixture(t, wrongElement, wrongRule, parseGongTestPillar(t, tc.target), 1)
			assertShenShaNameAbsentEverywhere(t, got, tc.name)
		}
	}
}

func TestDayStemRulesDoNotPublishSchoolHallShortcuts(t *testing.T) {
	for dayGan, rules := range dayGanShenShaRules {
		for _, rule := range rules {
			if rule.Name == "学堂" || rule.Name == "词馆" {
				t.Errorf("day stem %s still contains %s branch shortcut %+v", dayGan, rule.Name, rule)
			}
		}
	}
}

func TestNayinSchoolHallMetadataIsLocatedButNotAdjudicated(t *testing.T) {
	wants := map[string][]string{
		"学堂": {"生年纳音五行", "长生正位", "完整干支", "金辛巳、木己亥、水甲申、火丙寅、土戊申"},
		"词馆": {"生年纳音五行", "临官正位", "完整干支", "金壬申、木庚寅、水癸亥、火乙巳、土丁亥"},
	}
	for name, fragments := range wants {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata status = %+v", name, meta)
		}
		for _, fragment := range append(fragments, "《三命通会》", "PDF第105-106页", "书内第102-103页") {
			if !strings.Contains(meta.Basis, fragment) {
				t.Errorf("%s basis = %q, want %q", name, meta.Basis, fragment)
			}
		}
	}
}

func calcNayinSchoolHallFixture(t testing.TB, element string, rule nayinSchoolHallRule, placed model.Pillar, targetIndex int) ShenShaCalcResult {
	t.Helper()
	pillars := []model.Pillar{
		schoolHallYearPillar(t, element, rule),
		schoolHallNeutralPillar(t, rule, 0),
		schoolHallNeutralPillar(t, rule, 10),
		schoolHallNeutralPillar(t, rule, 20),
	}
	pillars[targetIndex] = placed
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: model.GenderMale,
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func schoolHallYearPillar(t testing.TB, element string, rule nayinSchoolHallRule) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		name := pillar.Gan + pillar.Zhi
		nayin := data.Nayin[data.GanIndex(pillar.Gan)][data.ZhiIndex(pillar.Zhi)]
		if data.NaYinMap[nayin].Element == element && name != rule.XueTang && name != rule.CiGuan {
			return pillar
		}
	}
	t.Fatalf("no neutral year pillar for element %s", element)
	return model.Pillar{}
}

func schoolHallNeutralPillar(t testing.TB, rule nayinSchoolHallRule, start int) model.Pillar {
	t.Helper()
	for offset := 0; offset < 60; offset++ {
		i := (start + offset) % 60
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		name := pillar.Gan + pillar.Zhi
		if name != rule.XueTang && name != rule.CiGuan {
			return pillar
		}
	}
	t.Fatal("no neutral school-hall pillar")
	return model.Pillar{}
}

func schoolHallPillarWithSameBranch(t testing.TB, target string) model.Pillar {
	t.Helper()
	targetPillar := parseGongTestPillar(t, target)
	for i := 0; i < 60; i++ {
		pillar := model.Pillar{Gan: data.Gans[i%10], Zhi: data.Zhis[i%12]}
		if pillar.Zhi == targetPillar.Zhi && pillar.Gan != targetPillar.Gan {
			return pillar
		}
	}
	t.Fatalf("no wrong-stem pillar for target %s", target)
	return model.Pillar{}
}
