package fortune

import "testing"

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

func TestPartialBranchScoresAreBelowCompleteGroups(t *testing.T) {
	if got := calcScore("unknown", "banHe", "", ""); got != 56 {
		t.Fatalf("半合 score=%d, want 56", got)
	}
	if got := calcScore("unknown", "gongHe", "", ""); got != 54 {
		t.Fatalf("拱合 score=%d, want 54", got)
	}
	if got := calcScore("unknown", "banHui", "", ""); got != 55 {
		t.Fatalf("半会 score=%d, want 55", got)
	}
	if !(relationScore("banHe") < relationScore("sanHe")) {
		t.Fatalf("分层评分中半合必须低于完整三合")
	}
	if !(relationScore("banHui") < relationScore("sanHui")) {
		t.Fatalf("分层评分中半会必须低于完整三会")
	}
}
