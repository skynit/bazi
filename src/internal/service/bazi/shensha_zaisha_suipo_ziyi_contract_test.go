package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestZaiShaUsesOnlySanHeRuleAndClashesWithJiangXing(t *testing.T) {
	for _, branch := range data.Zhis {
		if got := ruleTargetsByName(yearZhiShenShaRules[branch], "灾煞"); len(got) != 0 {
			t.Errorf("year branch table still duplicates 灾煞 for %s: %v", branch, got)
		}
		rule := sanHeShenShaRules[branch]
		if rule.ZaiSha == "" || rule.ZaiSha != zhiLiuChong[rule.Jiang] {
			t.Errorf("branch %s 将星=%s 灾煞=%s, want exact clash", branch, rule.Jiang, rule.ZaiSha)
		}
	}
}

func TestZaiShaYearAndDayKeysAssignExactlyOnce(t *testing.T) {
	for _, branch := range data.Zhis {
		target := sanHeShenShaRules[branch].ZaiSha
		other := zaiShaBranchWithDifferentTarget(branch, target)

		yearDerived := zaiShaFixture(t, branch, other, target, 1)
		assertExactShenShaHitCount(t, yearDerived.Month, "灾煞："+target, 1)

		dayDerived := zaiShaFixture(t, other, branch, target, 3)
		assertExactShenShaHitCount(t, dayDerived.Hour, "灾煞："+target, 1)
	}
}

func TestRetiredSuiPoAndIncorrectZiYiPathFailClosed(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"yearZhiChong", "appendShenShaByTargetBranch(res, branches, \"自缢煞\""} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains retired path %q", forbidden)
		}
	}

	for _, yearBranch := range data.Zhis {
		for _, branch := range data.Zhis {
			got := calcPoZhaiFixture(t, yearBranch, branch, 1)
			assertShenShaNameAbsentEverywhere(t, got, "岁破")
			assertShenShaNameAbsentEverywhere(t, got, "自缢煞")
		}
	}
	for _, name := range []string{"岁破", "自缢煞"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("%s metadata = %+v", name, meta)
		}
	}
}

func TestZaiShaMetadataRecordsFormulaWithoutRiskClaim(t *testing.T) {
	meta := LookupShenShaMeta("灾煞")
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("灾煞 metadata = %+v", meta)
	}
	for _, fragment := range []string{"年支或日支", "将星对冲支", "逐柱落位", "《三命通会》PDF第116页", "不生成伤亡、疾病或现实事件结论"} {
		if !strings.Contains(meta.Basis, fragment) {
			t.Errorf("灾煞 basis = %q, want %q", meta.Basis, fragment)
		}
	}
}

func zaiShaBranchWithDifferentTarget(branch, target string) string {
	for _, candidate := range data.Zhis {
		if candidate != branch && sanHeShenShaRules[candidate].ZaiSha != target {
			return candidate
		}
	}
	return ""
}

func zaiShaFixture(t testing.TB, yearBranch, dayBranch, target string, targetIndex int) ShenShaCalcResult {
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

func assertExactShenShaHitCount(t testing.TB, items []string, want string, count int) {
	t.Helper()
	got := 0
	for _, item := range items {
		if item == want {
			got++
		}
	}
	if got != count {
		t.Errorf("items=%v contain %q %d times, want %d", items, want, got, count)
	}
}
