package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestDeXiuDedicatedChapterTableIsExactForAllMonthsAndStems(t *testing.T) {
	wantRules := map[string]deXiuRule{
		"寅": {De: "丙丁", Xiu: "戊癸"}, "午": {De: "丙丁", Xiu: "戊癸"}, "戌": {De: "丙丁", Xiu: "戊癸"},
		"申": {De: "壬癸戊己", Xiu: "丙辛甲己"}, "子": {De: "壬癸戊己", Xiu: "丙辛甲己"}, "辰": {De: "壬癸戊己", Xiu: "丙辛甲己"},
		"巳": {De: "庚辛", Xiu: "乙庚"}, "酉": {De: "庚辛", Xiu: "乙庚"}, "丑": {De: "庚辛", Xiu: "乙庚"},
		"亥": {De: "甲乙", Xiu: "丁壬"}, "卯": {De: "甲乙", Xiu: "丁壬"}, "未": {De: "甲乙", Xiu: "丁壬"},
	}
	for _, monthZhi := range data.Zhis {
		got, ok := deXiuRuleForMonth(monthZhi)
		if !ok || got != wantRules[monthZhi] {
			t.Fatalf("month %s 德秀 rule = %+v, %v, want %+v", monthZhi, got, ok, wantRules[monthZhi])
		}
		for _, gan := range data.Gans {
			wantRole := expectedDeXiuRole(wantRules[monthZhi], gan)
			if gotRole := deXiuRole(monthZhi, gan); gotRole != wantRole {
				t.Errorf("month %s stem %s role = %q, want %q", monthZhi, gan, gotRole, wantRole)
			}
		}
	}
	if _, ok := deXiuRuleForMonth("unknown"); ok || deXiuRole("unknown", "甲") != "" || deXiuRole("寅", "unknown") != "" {
		t.Fatal("invalid month or stem must not produce 德秀 evidence")
	}
}

func TestDeXiuCorrectsLegacyMissingAndFalsePositiveStems(t *testing.T) {
	for _, tc := range []struct {
		monthZhi string
		gan      string
		wantRole string
	}{
		{monthZhi: "申", gan: "甲", wantRole: "秀"},
		{monthZhi: "申", gan: "辛", wantRole: "秀"},
		{monthZhi: "申", gan: "己", wantRole: "德秀"},
		{monthZhi: "亥", gan: "丁", wantRole: "秀"},
		{monthZhi: "亥", gan: "壬", wantRole: "秀"},
		{monthZhi: "亥", gan: "己", wantRole: ""},
		{monthZhi: "亥", gan: "庚", wantRole: ""},
		{monthZhi: "巳", gan: "丙", wantRole: ""},
		{monthZhi: "巳", gan: "乙", wantRole: "秀"},
	} {
		if got := deXiuRole(tc.monthZhi, tc.gan); got != tc.wantRole {
			t.Errorf("month %s stem %s role = %q, want %q", tc.monthZhi, tc.gan, got, tc.wantRole)
		}
	}
}

func TestDeXiuRoleEvidenceFlowsThroughAuthoritativeEntrypoint(t *testing.T) {
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year:   model.Pillar{Gan: "甲", Zhi: "子"},
		Month:  model.Pillar{Gan: "戊", Zhi: "申"},
		Day:    model.Pillar{Gan: "辛", Zhi: "巳"},
		Hour:   model.Pillar{Gan: "己", Zhi: "丑"},
		Gender: "MALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertExactShenShaInBucket(t, "year", got.Year, "德秀贵人：秀/甲")
	assertExactShenShaInBucket(t, "month", got.Month, "德秀贵人：德/戊")
	assertExactShenShaInBucket(t, "day", got.Day, "德秀贵人：秀/辛")
	assertExactShenShaInBucket(t, "hour", got.Hour, "德秀贵人：德秀/己")

	got, err = CalcShenShaByPillars(ShenShaPillars{
		Year:   model.Pillar{Gan: "庚", Zhi: "午"},
		Month:  model.Pillar{Gan: "乙", Zhi: "亥"},
		Day:    model.Pillar{Gan: "丁", Zhi: "卯"},
		Hour:   model.Pillar{Gan: "壬", Zhi: "子"},
		Gender: "FEMALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertNoPillarBucketHas(t, got, "德秀贵人：德/庚")
	assertNoPillarBucketHas(t, got, "德秀贵人：秀/庚")
	assertExactShenShaInBucket(t, "month", got.Month, "德秀贵人：德/乙")
	assertExactShenShaInBucket(t, "day", got.Day, "德秀贵人：秀/丁")
	assertExactShenShaInBucket(t, "hour", got.Hour, "德秀贵人：秀/壬")

	got, err = CalcShenShaByPillars(ShenShaPillars{
		Year:   model.Pillar{Gan: "己", Zhi: "丑"},
		Month:  model.Pillar{Gan: "乙", Zhi: "亥"},
		Day:    model.Pillar{Gan: "甲", Zhi: "子"},
		Hour:   model.Pillar{Gan: "庚", Zhi: "午"},
		Gender: "MALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertNoPillarBucketHas(t, got, "德秀贵人：德/己")
	assertNoPillarBucketHas(t, got, "德秀贵人：秀/己")
	assertNoPillarBucketHas(t, got, "德秀贵人：德/庚")
	assertNoPillarBucketHas(t, got, "德秀贵人：秀/庚")

	got, err = CalcShenShaByPillars(ShenShaPillars{
		Year:   model.Pillar{Gan: "丙", Zhi: "寅"},
		Month:  model.Pillar{Gan: "辛", Zhi: "巳"},
		Day:    model.Pillar{Gan: "乙", Zhi: "丑"},
		Hour:   model.Pillar{Gan: "庚", Zhi: "午"},
		Gender: "FEMALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertNoPillarBucketHas(t, got, "德秀贵人：德/丙")
	assertNoPillarBucketHas(t, got, "德秀贵人：秀/丙")
	assertExactShenShaInBucket(t, "month", got.Month, "德秀贵人：德/辛")
	assertExactShenShaInBucket(t, "day", got.Day, "德秀贵人：秀/乙")
	assertExactShenShaInBucket(t, "hour", got.Hour, "德秀贵人：德秀/庚")

	got, err = CalcShenShaByPillars(ShenShaPillars{
		Year:   model.Pillar{Gan: "甲", Zhi: "子"},
		Month:  model.Pillar{Gan: "戊", Zhi: "申"},
		Day:    model.Pillar{Gan: "丁", Zhi: "卯"},
		Hour:   model.Pillar{Gan: "甲", Zhi: "子"},
		Gender: "MALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	assertExactShenShaInBucket(t, "year", got.Year, "德秀贵人：秀/甲")
	assertExactShenShaInBucket(t, "hour", got.Hour, "德秀贵人：秀/甲")
}

func TestDeXiuMetadataPublishesProfileAndTextualDispute(t *testing.T) {
	meta := LookupShenShaMeta("德秀贵人")
	for _, want := range []string{"《三命通会》第108页", "《渊海子平》第666页", "差异待裁决", "德/秀角色"} {
		if !strings.Contains(meta.Basis, want) {
			t.Errorf("德秀贵人 basis = %q, want %q", meta.Basis, want)
		}
	}
	if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
		t.Fatalf("德秀贵人 metadata = %+v", meta)
	}
}

func expectedDeXiuRole(rule deXiuRule, gan string) string {
	isDe := strings.Contains(rule.De, gan)
	isXiu := strings.Contains(rule.Xiu, gan)
	switch {
	case isDe && isXiu:
		return "德秀"
	case isDe:
		return "德"
	case isXiu:
		return "秀"
	default:
		return ""
	}
}
