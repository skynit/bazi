package ziwei

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFiveRatHourStemCoversAllDayStemsAndBranches(t *testing.T) {
	ziHourStem := [10]int{0, 2, 4, 6, 8, 0, 2, 4, 6, 8}
	for dayStem := 0; dayStem < len(StemNames); dayStem++ {
		for branch := 0; branch < len(BranchNames); branch++ {
			got, ok := fiveRatHourStem(dayStem, branch)
			want := (ziHourStem[dayStem] + branch) % len(StemNames)
			if !ok || got != want {
				t.Fatalf("day stem %s branch %s hour stem = %d/%t, want %d", StemNames[dayStem], BranchNames[branch], got, ok, want)
			}
		}
	}
	for _, input := range [][2]int{{-1, 0}, {10, 0}, {0, -1}, {0, 12}} {
		if _, ok := fiveRatHourStem(input[0], input[1]); ok {
			t.Fatalf("invalid five-rat input accepted: day_stem=%d branch=%d", input[0], input[1])
		}
	}
}

func TestTraditionalHourIntervalsAreExplicitAndStable(t *testing.T) {
	wantStart := [12]int{23, 1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21}
	wantEnd := [12]int{1, 3, 5, 7, 9, 11, 13, 15, 17, 19, 21, 23}
	wantLabel := [12]string{
		"23:00-00:59", "01:00-02:59", "03:00-04:59", "05:00-06:59",
		"07:00-08:59", "09:00-10:59", "11:00-12:59", "13:00-14:59",
		"15:00-16:59", "17:00-18:59", "19:00-20:59", "21:00-22:59",
	}
	for branch := range BranchNames {
		start, end, label, crosses, ok := traditionalHourInterval(branch)
		if !ok || start != wantStart[branch] || end != wantEnd[branch] || label != wantLabel[branch] || crosses != (branch == 0) {
			t.Fatalf("%s interval = %d/%d/%q crosses=%t ok=%t", BranchNames[branch], start, end, label, crosses, ok)
		}
	}
	for _, branch := range []int{-1, 12} {
		if _, _, _, _, ok := traditionalHourInterval(branch); ok {
			t.Fatalf("invalid hour branch accepted: %d", branch)
		}
	}
}

func TestLiuriHourBlocksExposeBoundaryAndEvidenceContract(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	liuri := svc.CalculateLiuriForDate(base, 2026, 7, 15)
	result := NewPeriodInterpreterFromChart(base).AnalyzeLiuri(liuri, 2026, 7, 15)
	if result == nil || len(result.HourlyAnalysis) != len(BranchNames) {
		t.Fatalf("hourly analysis = %+v", result)
	}
	for branch, block := range result.HourlyAnalysis {
		if block.Branch != BranchNames[branch] || block.Stem == "" || block.StemBranch != block.Stem+block.Branch ||
			block.DayStemBasis != "period_derivation_day_stem" || block.BoundaryPolicy != periodHourBoundaryPolicy ||
			block.RuleID != periodHourStemRuleID || block.EvidenceBasis != "deterministic_rule_projection" ||
			block.ValidationStatus != "not_adjudicated" || block.IsOutcomeConclusion {
			t.Fatalf("hour block %s contract = %+v", BranchNames[branch], block)
		}
	}

	payload, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	serialized := string(payload)
	if strings.Contains(serialized, `"hour":`) {
		t.Fatalf("ambiguous legacy hour field leaked into JSON: %s", serialized)
	}
	for _, required := range []string{
		`"interval_start_hour":23`,
		`"interval_end_hour_exclusive":1`,
		`"interval_label":"23:00-00:59"`,
		`"crosses_midnight":true`,
		`"boundary_policy":"traditional_two_hour_branch_slots_no_civil_date_assignment"`,
	} {
		if !strings.Contains(serialized, required) {
			t.Fatalf("hour boundary contract missing %s", required)
		}
	}
}
