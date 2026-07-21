package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestCanonicalSanHeTablesAndYearDayKeys(t *testing.T) {
	targetFor := map[string]func(sanHeShenSha) string{
		"咸池": func(rule sanHeShenSha) string { return rule.XianChi },
		"劫煞": func(rule sanHeShenSha) string { return rule.JieSha },
		"亡神": func(rule sanHeShenSha) string { return rule.WangShen },
	}
	for name, getTarget := range targetFor {
		for _, branch := range data.Zhis {
			target := getTarget(sanHeShenShaRules[branch])
			other := sanHeBranchWithDifferentTarget(branch, target, getTarget)

			yearDerived := canonicalSanHeFixture(t, branch, other, target, 1)
			assertExactShenShaHitCount(t, yearDerived.Month, name+"："+target, 1)

			dayDerived := canonicalSanHeFixture(t, other, branch, target, 3)
			assertExactShenShaHitCount(t, dayDerived.Hour, name+"："+target, 1)
		}
	}
}

func TestCanonicalNamesReplaceUnregisteredHourAliases(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, alias := range []string{"桃花", "时桃花", "时煞", "时马", "时刃", "时禄"} {
		if strings.Contains(string(source), "\""+alias+"\"") {
			t.Errorf("production source still publishes alias %q", alias)
		}
		meta := LookupShenShaMeta(alias)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" {
			t.Errorf("alias %s metadata = %+v", alias, meta)
		}
	}
}

func TestCanonicalSanHeMetadataHasLocatedFormula(t *testing.T) {
	want := map[string][]string{
		"咸池": {"年支或日支", "五行沐浴位", "逐柱落位", "《三命通会》PDF第81页", "正式输出统一使用原名咸池"},
		"劫煞": {"年支或日支", "五行绝处", "逐柱落位", "《三命通会》PDF第108页", "《渊海子平》PDF第665页"},
		"亡神": {"年支或日支", "五行临官/泄气位", "逐柱落位", "《三命通会》PDF第108-109页", "《渊海子平》PDF第668页"},
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

func sanHeBranchWithDifferentTarget(branch, target string, getTarget func(sanHeShenSha) string) string {
	for _, candidate := range data.Zhis {
		if candidate != branch && getTarget(sanHeShenShaRules[candidate]) != target {
			return candidate
		}
	}
	return ""
}

func canonicalSanHeFixture(t testing.TB, yearBranch, dayBranch, target string, targetIndex int) ShenShaCalcResult {
	t.Helper()
	pillars := []model.Pillar{
		poZhaiPillarForBranch(t, yearBranch),
		poZhaiNeutralPillar(t, target, 10),
		poZhaiPillarForBranch(t, dayBranch),
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
