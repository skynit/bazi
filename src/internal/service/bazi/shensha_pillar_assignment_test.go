package bazi

import (
	"testing"

	"bazi/internal/model"
)

func TestYearBranchExtrasAttachToEveryMatchingTargetPillar(t *testing.T) {
	pillars := ShenShaPillars{
		Year:   model.Pillar{Gan: "甲", Zhi: "子"},
		Month:  model.Pillar{Gan: "乙", Zhi: "丑"},
		Day:    model.Pillar{Gan: "丁", Zhi: "卯"},
		Hour:   model.Pillar{Gan: "己", Zhi: "卯"},
		Gender: "MALE",
	}
	got, err := CalcShenShaByPillars(pillars)
	if err != nil {
		t.Fatal(err)
	}

	for _, bucket := range []struct {
		name  string
		items []string
	}{
		{name: "day", items: got.Day},
		{name: "hour", items: got.Hour},
	} {
		for _, want := range []string{"六厄：卯"} {
			if !containsExactShenSha(bucket.items, want) {
				t.Errorf("%s shen-sha = %v, want %s", bucket.name, bucket.items, want)
			}
		}
	}
	for _, bucket := range []struct {
		name  string
		items []string
	}{
		{name: "year", items: got.Year},
		{name: "month", items: got.Month},
	} {
		for _, unwanted := range []string{"六厄：卯"} {
			if containsExactShenSha(bucket.items, unwanted) {
				t.Errorf("%s shen-sha = %v, must not contain %s", bucket.name, bucket.items, unwanted)
			}
		}
	}
	assertShenShaNameAbsentEverywhere(t, got, "红鸾")
	assertShenShaNameAbsentEverywhere(t, got, "天喜")
	if !containsExactShenSha(got.Month, "攀鞍：丑") {
		t.Errorf("month shen-sha = %v, want 攀鞍：丑", got.Month)
	}
	if containsExactShenSha(got.Year, "攀鞍：丑") {
		t.Errorf("year shen-sha = %v, must not contain 攀鞍：丑", got.Year)
	}
}

func TestGenderBasedShenShaAssignsDistinctClassicalTargets(t *testing.T) {
	pillars := ShenShaPillars{
		Year:   model.Pillar{Gan: "甲", Zhi: "子"},
		Month:  model.Pillar{Gan: "乙", Zhi: "未"},
		Day:    model.Pillar{Gan: "丁", Zhi: "卯"},
		Hour:   model.Pillar{Gan: "辛", Zhi: "酉"},
		Gender: "MALE",
	}
	var got ShenShaCalcResult
	branches := []string{pillars.Year.Zhi, pillars.Month.Zhi, pillars.Day.Zhi, pillars.Hour.Zhi}
	addGenderBasedShenSha(pillars, branches, &got)

	for _, tc := range []struct {
		name  string
		items []string
		want  string
	}{
		{name: "month", items: got.Month, want: "元辰：未"},
		{name: "day", items: got.Day, want: "勾煞：卯"},
		{name: "hour", items: got.Hour, want: "绞煞：酉"},
	} {
		if !containsExactShenSha(tc.items, tc.want) {
			t.Errorf("%s shen-sha = %v, want %s", tc.name, tc.items, tc.want)
		}
	}
	if len(got.Year) != 0 {
		t.Errorf("year shen-sha = %v, want no gender-derived target hit", got.Year)
	}
	assertShenShaNameAbsentEverywhere(t, got, "勾绞煞")
	assertShenShaNameAbsentEverywhere(t, got, "暴败煞")
}
