package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/service/data"
)

var liuETargets = map[string]string{
	"子": "卯", "丑": "子", "寅": "酉", "卯": "午", "辰": "卯", "巳": "子",
	"午": "酉", "未": "午", "申": "卯", "酉": "子", "戌": "酉", "亥": "午",
}

func TestLiuEAndHuaGaiExactTables(t *testing.T) {
	for _, branch := range data.Zhis {
		if got := sanHeLiuE[branch]; got != liuETargets[branch] {
			t.Errorf("year branch %s 六厄=%s, want %s", branch, got, liuETargets[branch])
		}
		rule := sanHeShenShaRules[branch]
		if rule.HuaGai == "" || rule.HuaGai != sanHeGraveTarget(branch) {
			t.Errorf("branch %s 华盖=%s, want trine grave %s", branch, rule.HuaGai, sanHeGraveTarget(branch))
		}
	}
}

func TestLiuEAssignsEveryMatchingPillarFromYearKey(t *testing.T) {
	for _, yearBranch := range data.Zhis {
		target := liuETargets[yearBranch]
		for targetIndex := 1; targetIndex < 4; targetIndex++ {
			got := calcLocatedYearBranchFixture(t, yearBranch, target, targetIndex)
			assertOnlyPillarBucketHas(t, got, targetIndex, "六厄："+target)
			assertShenShaNameAbsentEverywhere(t, got, "墓煞")
		}
	}
}

func TestIncorrectMuShaAliasFailsClosed(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"sanHeMuSha", "\"墓煞\""} {
		if strings.Contains(string(source), forbidden) {
			t.Errorf("production source still contains incorrect 墓煞 path %q", forbidden)
		}
	}
	meta := LookupShenShaMeta("墓煞")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
		t.Fatalf("墓煞 metadata = %+v", meta)
	}
}

func TestLiuEAndHuaGaiMetadata(t *testing.T) {
	want := map[string][]string{
		"六厄": {"生年支", "五行死位", "逐柱落位", "《三命通会》PDF第117页", "不生成仕途、困境或现实事件结论"},
		"华盖": {"年支或日支", "三合墓库位", "逐柱落位", "《三命通会》PDF第80-81页"},
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

func sanHeGraveTarget(branch string) string {
	switch branch {
	case "寅", "午", "戌":
		return "戌"
	case "巳", "酉", "丑":
		return "丑"
	case "申", "子", "辰":
		return "辰"
	case "亥", "卯", "未":
		return "未"
	default:
		return ""
	}
}
