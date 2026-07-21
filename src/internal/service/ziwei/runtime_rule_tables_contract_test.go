package ziwei

import "testing"

func TestRuntimeRuleTablesHashIsPinned(t *testing.T) {
	got, err := runtimeRuleTablesHash()
	if err != nil {
		t.Fatal(err)
	}
	if got != ZiWeiRuntimeRuleTablesSHA256 {
		t.Fatalf("runtime rule-table hash = %q, want pinned %q", got, ZiWeiRuntimeRuleTablesSHA256)
	}
}

func TestRuntimeRuleTableMutationFailsClosed(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	original := SiHuaTable[0][0]
	SiHuaTable[0][0] = "篡改四化"
	t.Cleanup(func() { SiHuaTable[0][0] = original })

	if forged, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男"); err == nil || forged != nil {
		t.Fatalf("mutated runtime table produced chart=%+v err=%v", forged, err)
	}
	if svc.ChartMatchesProfile(chart, DefaultProfileID) {
		t.Fatal("cached chart remained valid after runtime rule-table mutation")
	}
}
