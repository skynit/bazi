package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
)

func TestPanAnUsesBranchBeforeYiMaForAllYearGroups(t *testing.T) {
	wantByYearBranch := map[string]string{
		"寅": "未", "午": "未", "戌": "未",
		"巳": "戌", "酉": "戌", "丑": "戌",
		"申": "丑", "子": "丑", "辰": "丑",
		"亥": "辰", "卯": "辰", "未": "辰",
	}
	for yearBranch, want := range wantByYearBranch {
		if got := panAnByYearZhi(yearBranch); got != want {
			t.Errorf("year branch %s pan-an = %s, want %s", yearBranch, got, want)
		}
	}
	if got := panAnByYearZhi("unknown"); got != "" {
		t.Fatalf("unknown year branch pan-an = %q, want empty", got)
	}
}

func TestYearBranchExtraKeepsPanAnAndDropsUnsupportedAliases(t *testing.T) {
	pillars := ShenShaPillars{
		Year:   model.Pillar{Gan: "壬", Zhi: "申"},
		Month:  model.Pillar{Gan: "癸", Zhi: "丑"},
		Day:    model.Pillar{Gan: "甲", Zhi: "辰"},
		Hour:   model.Pillar{Gan: "甲", Zhi: "子"},
		Gender: "MALE",
	}
	var got ShenShaCalcResult
	addYearZhiExtra(pillars, []string{"申", "丑", "辰", "子"}, &got)

	if !containsExactShenSha(got.Month, "攀鞍：丑") {
		t.Errorf("month shen-sha = %v, want 攀鞍：丑", got.Month)
	}
	if containsExactShenSha(got.Day, "攀鞍：辰") {
		t.Fatalf("pan-an used the trine grave instead of the branch before yi-ma: %v", got.Day)
	}
	if containsExactShenSha(got.Year, "攀鞍：丑") {
		t.Fatalf("target-branch shen-sha leaked into year pillar: %v", got.Year)
	}
	for _, bucket := range [][]string{got.Year, got.Month, got.Day, got.Hour, got.Global} {
		if hasShenShaName(bucket, "墓煞") {
			t.Fatalf("incorrect 墓煞 alias leaked into result: %+v", got)
		}
	}
	for _, item := range got.Year {
		if strings.HasPrefix(item, "功曹") {
			t.Fatalf("unsupported gong-cao rule still emitted: %v", got.Year)
		}
	}
	meta := LookupShenShaMeta("攀鞍")
	if !strings.Contains(meta.Basis, "马前一辰") || !strings.Contains(meta.Basis, "635") {
		t.Fatalf("pan-an source metadata = %+v", meta)
	}
}

func containsExactShenSha(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
