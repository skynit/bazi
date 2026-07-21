package fortune

import (
	"testing"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
)

func TestBranchRelationPartialGroups(t *testing.T) {
	tests := []struct {
		name string
		a    string
		b    string
		want string
	}{
		{"三合含中神为半合", "申", "子", "banHe"},
		{"三合首尾为拱合", "申", "辰", "gongHe"},
		{"三会两支为半会", "寅", "卯", "banHui"},
		{"六合优先", "辰", "酉", "combine"},
		{"六冲优先", "子", "午", "clash"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := branchRelation(tc.a, tc.b); got != tc.want {
				t.Fatalf("branchRelation(%s,%s)=%s, want %s", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

func TestBranchRelationPunishmentIsSymmetric(t *testing.T) {
	for _, pair := range [][2]string{
		{"丑", "戌"}, {"戌", "未"}, {"子", "卯"},
	} {
		if forward, reverse := branchRelation(pair[0], pair[1]), branchRelation(pair[1], pair[0]); forward != "punish" || reverse != "punish" {
			t.Errorf("%s%s关系不对称: forward=%s reverse=%s", pair[0], pair[1], forward, reverse)
		}
	}
}

func TestBranchRelationDoesNotTreatSameBranchAsHalfMeeting(t *testing.T) {
	for _, branch := range []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"} {
		if got := branchRelation(branch, branch); got != "same" {
			t.Errorf("branchRelation(%s, %s) = %s, want same", branch, branch, got)
		}
	}
	if got := calcScore("unknown", "same", "", ""); got != 50 {
		t.Fatalf("同支不应被误作半会加分: score=%d, want 50", got)
	}
	if label := branchRelLabel("same"); label != "同支" {
		t.Fatalf("same label = %s, want 同支", label)
	}
}

func TestPartialBranchRelationsHaveDistinctStructuralIndexes(t *testing.T) {
	if got := calcScore("unknown", "banHe", "", ""); got != 56 {
		t.Fatalf("半合 score=%d, want 56", got)
	}
	if got := calcScore("unknown", "gongHe", "", ""); got != 54 {
		t.Fatalf("拱合 score=%d, want 54", got)
	}
	if got := calcScore("unknown", "banHui", "", ""); got != 55 {
		t.Fatalf("半会 score=%d, want 55", got)
	}
}

func TestRikuyoBranchRelationsPreserveCompoundStructures(t *testing.T) {
	chart := &bazipkg.BaziResult{DayPillar: model.Pillar{Zhi: "巳"}}
	relations := calcBranchRelations("申", chart)
	assertRikuyoRelationTypes(t, relations, "日支", []string{"punish", "break", "combine"})

	chart.DayPillar.Zhi = "辰"
	relations = calcBranchRelations("辰", chart)
	assertRikuyoRelationTypes(t, relations, "日支", []string{"same", "punish"})
}

func assertRikuyoRelationTypes(t *testing.T, relations []BranchRelation, target string, want []string) {
	t.Helper()
	got := map[string]bool{}
	for _, relation := range relations {
		if relation.TargetPillar != target {
			continue
		}
		got[relation.Type] = true
		if relation.RuleID != "rikuyo.branch-relation-v3."+relation.Type ||
			relation.Basis != "query_day_branch_and_natal_pillar_branch_all_structures" {
			t.Fatalf("复合关系缺少可复核依据: %+v", relation)
		}
	}
	for _, relationType := range want {
		if !got[relationType] {
			t.Errorf("%s关系缺少%s: %+v", target, relationType, relations)
		}
	}
}
