package bazi

import (
	"strings"
	"testing"

	"bazi/internal/service/data"
)

var jiangXingTargets = map[string]string{
	"子": "子", "丑": "酉", "寅": "午", "卯": "卯", "辰": "子", "巳": "酉",
	"午": "午", "未": "卯", "申": "子", "酉": "酉", "戌": "午", "亥": "卯",
}

var yiMaTargets = map[string]string{
	"子": "寅", "丑": "亥", "寅": "申", "卯": "巳", "辰": "寅", "巳": "亥",
	"午": "申", "未": "巳", "申": "寅", "酉": "亥", "戌": "申", "亥": "巳",
}

func TestJiangXingAndYiMaExactTables(t *testing.T) {
	for _, branch := range data.Zhis {
		rule := sanHeShenShaRules[branch]
		if rule.Jiang != jiangXingTargets[branch] {
			t.Errorf("branch %s 将星=%s, want %s", branch, rule.Jiang, jiangXingTargets[branch])
		}
		if rule.YiMa != yiMaTargets[branch] {
			t.Errorf("branch %s 驿马=%s, want %s", branch, rule.YiMa, yiMaTargets[branch])
		}
		if rule.YiMa != zhiLiuChong[sanHeFirstBranch(branch)] {
			t.Errorf("branch %s 驿马=%s, want clash of trine first branch %s", branch, rule.YiMa, sanHeFirstBranch(branch))
		}
	}
}

func TestJiangXingAndYiMaUseYearAndDayKeys(t *testing.T) {
	for name, getTarget := range map[string]func(sanHeShenSha) string{
		"将星": func(rule sanHeShenSha) string { return rule.Jiang },
		"驿马": func(rule sanHeShenSha) string { return rule.YiMa },
	} {
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

func TestJiangXingAndYiMaMetadata(t *testing.T) {
	want := map[string][]string{
		"将星": {"年支或日支", "三合旺位", "逐柱落位", "《三命通会》PDF第80页", "不生成职位、权力或现实事件结论"},
		"驿马": {"年支或日支", "首支对冲位", "逐柱落位", "《三命通会》PDF第92页", "《渊海子平》PDF第665页", "不生成迁移、升职或现实事件结论"},
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

func sanHeFirstBranch(branch string) string {
	switch branch {
	case "寅", "午", "戌":
		return "寅"
	case "巳", "酉", "丑":
		return "巳"
	case "申", "子", "辰":
		return "申"
	case "亥", "卯", "未":
		return "亥"
	default:
		return ""
	}
}
