package bazi

import (
	"reflect"
	"strings"
	"testing"

	"github.com/6tail/tyme4go/tyme"
)

func TestCalcDaYunIncludesDateLevelStartEvidence(t *testing.T) {
	birth, err := tyme.SolarTime{}.FromYmdHms(2022, 3, 9, 20, 51, 0)
	if err != nil {
		t.Fatalf("create birth time: %v", err)
	}

	got := calcDaYun(birth, tyme.MAN)

	if !got.Calculated {
		t.Fatal("calculated = false, want true")
	}
	if got.Direction != "顺行" {
		t.Fatalf("direction = %q, want 顺行", got.Direction)
	}
	if got.StartAge != 8 {
		t.Errorf("start_age = %d, want 8", got.StartAge)
	}
	wantAge := DaYunStartAge{Years: 8, Months: 9, Days: 2, Hours: 10, Minutes: 28}
	if got.StartAgeDetail != wantAge {
		t.Errorf("start_age_detail = %+v, want %+v", got.StartAgeDetail, wantAge)
	}
	if got.StartAt != "2030-12-12T07:19:00" {
		t.Errorf("start_at = %q, want %q", got.StartAt, "2030-12-12T07:19:00")
	}
	if got.PreviousJie == nil || got.PreviousJie.Name != "惊蛰" || got.PreviousJie.DeltaSeconds >= 0 {
		t.Errorf("previous_jie = %+v, want past 惊蛰", got.PreviousJie)
	}
	if got.NextJie == nil || got.NextJie.Name != "清明" || got.NextJie.DeltaSeconds <= 0 {
		t.Errorf("next_jie = %+v, want future 清明", got.NextJie)
	}
	if got.ReferenceJie == nil || got.ReferenceJie.Name != "清明" {
		t.Errorf("reference_jie = %+v, want 清明", got.ReferenceJie)
	} else if got.ReferenceDeltaSeconds != got.ReferenceJie.DeltaSeconds {
		t.Errorf("reference_delta_seconds = %d, want %d", got.ReferenceDeltaSeconds, got.ReferenceJie.DeltaSeconds)
	}
	if !strings.Contains(got.DirectionBasis, "阳男顺行") {
		t.Errorf("direction_basis = %q, want 阳男顺行 evidence", got.DirectionBasis)
	}
	if got.CalculationProfile == "" || got.Provider == "" || got.AgeConversionRule == "" || got.BoundaryRule == "" {
		t.Errorf("missing calculation metadata: %+v", got)
	}
}

func TestCalcDaYunIgnoresMutableGlobalChildLimitProvider(t *testing.T) {
	birth, err := tyme.SolarTime{}.FromYmdHms(2022, 3, 9, 20, 51, 0)
	if err != nil {
		t.Fatal(err)
	}
	baseline := calcDaYun(birth, tyme.MAN)

	original := tyme.ChildLimitProvider
	t.Cleanup(func() { tyme.ChildLimitProvider = original })
	if baseline.Provider != "tyme.DefaultChildLimitProvider" {
		t.Fatalf("baseline provider = %q, want fixed default", baseline.Provider)
	}
	tyme.ChildLimitProvider = tyme.LunarSect2ChildLimitProvider{}

	got := calcDaYun(birth, tyme.MAN)
	if !reflect.DeepEqual(got, baseline) {
		t.Fatalf("mutable global ChildLimitProvider changed service output:\nbaseline=%+v\ngot=%+v", baseline, got)
	}
	if got.Provider != "tyme.DefaultChildLimitProvider" {
		t.Fatalf("provider evidence = %q, want fixed default", got.Provider)
	}
}

func TestCalcDaYunJieBoundaryBelongsToNewJie(t *testing.T) {
	jie, err := tyme.SolarTerm{}.FromName(2022, "惊蛰")
	if err != nil {
		t.Fatalf("create solar term: %v", err)
	}
	birth := jie.GetJulianDay().GetSolarTime()

	// 壬寅为阳年，女命逆行。出生时刻恰等于惊蛰时，逆行应取当前惊蛰，
	// 而不是退回上一个节令，故节令时间差和起运年龄均为零。
	got := calcDaYun(&birth, tyme.WOMAN)

	if !got.Calculated {
		t.Fatal("calculated = false, want true")
	}
	if got.Direction != "逆行" {
		t.Fatalf("direction = %q, want 逆行", got.Direction)
	}
	if got.ReferenceJie == nil || got.ReferenceJie.Name != "惊蛰" || got.ReferenceJie.DeltaSeconds != 0 {
		t.Errorf("reference_jie = %+v, want current 惊蛰 with zero delta", got.ReferenceJie)
	}
	if got.PreviousJie == nil || got.PreviousJie.Name != "惊蛰" || got.PreviousJie.DeltaSeconds != 0 {
		t.Errorf("previous_jie = %+v, want boundary 惊蛰 with zero delta", got.PreviousJie)
	}
	if got.NextJie == nil || got.NextJie.Name != "清明" || got.NextJie.DeltaSeconds <= 0 {
		t.Errorf("next_jie = %+v, want future 清明", got.NextJie)
	}
	if got.StartAge != 0 || got.StartAgeDetail != (DaYunStartAge{}) {
		t.Errorf("start age = %d %+v, want all zero at exact reverse boundary", got.StartAge, got.StartAgeDetail)
	}
	if got.StartAt != formatSolarTime(birth) {
		t.Errorf("start_at = %q, want birth time %q", got.StartAt, formatSolarTime(birth))
	}
	if !strings.Contains(got.BoundaryRule, "等于") {
		t.Errorf("boundary_rule = %q, want equality convention", got.BoundaryRule)
	}
}
