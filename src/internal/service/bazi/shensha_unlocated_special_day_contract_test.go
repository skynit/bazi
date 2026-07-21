package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/service/data"
)

func TestUnlocatedTenLingYangShaYinShaAreAbsentAcrossSixtyDays(t *testing.T) {
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		day := data.Gans[dayIndex%10] + data.Zhis[dayIndex%12]
		result := calcSpecialDayFixture(t, day)
		for _, name := range []string{"十灵日", "阳煞", "阴煞", "阴阳煞"} {
			if specialDayContainsName(day, name) {
				t.Errorf("special day table still contains %s for %s", name, day)
			}
			assertShenShaNameAbsentEverywhere(t, result, name)
		}
	}
}

func TestFormerTenLingYangShaYinShaMembersRemainNegative(t *testing.T) {
	former := map[string][]string{
		"十灵日": {"戊午", "甲辰", "庚戌", "辛亥", "丙辰", "乙亥", "丁酉", "庚寅", "壬寅", "癸未"},
		"阳煞":  {"戊子", "戊午", "壬子", "壬午", "丙申", "壬申", "甲寅", "甲子", "甲午", "甲申", "丙寅", "戊申", "戊寅", "庚申", "庚寅", "壬寅"},
		"阴煞":  {"乙卯", "乙酉", "辛卯", "辛酉", "己卯", "己酉", "乙未", "丁卯", "癸卯", "乙丑", "癸酉", "丁酉"},
	}
	wantCounts := map[string]int{"十灵日": 10, "阳煞": 16, "阴煞": 12}
	for name, days := range former {
		if len(days) != wantCounts[name] {
			t.Fatalf("former %s fixture count = %d, want %d", name, len(days), wantCounts[name])
		}
		for _, day := range days {
			result := calcSpecialDayFixture(t, day)
			assertShenShaNameAbsentEverywhere(t, result, name)
		}
	}
}

func TestRemovingUnlocatedNamesPreservesIndependentSpecialDayRules(t *testing.T) {
	for _, tc := range []struct {
		day   string
		wants []string
	}{
		{day: "甲辰", wants: []string{"十恶大败"}},
		{day: "戊午", wants: []string{"九丑日", "孤鸾煞"}},
		{day: "乙卯", wants: []string{"九丑日", "八专"}},
		{day: "庚申", wants: []string{"八专"}},
	} {
		result := calcSpecialDayFixture(t, tc.day)
		for _, want := range tc.wants {
			if !hasShenShaName(result.Day, want) {
				t.Errorf("day %s lost independent rule %s: %+v", tc.day, want, result)
			}
		}
	}
}

func TestUnlocatedNamesAreAbsentFromProductionSourceAndMetadata(t *testing.T) {
	source, err := os.ReadFile("shensha.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"十灵日", "阳煞", "阴煞", "阴阳煞"} {
		if strings.Contains(string(source), `"`+name+`"`) {
			t.Errorf("production shen-sha source still publishes %s", name)
		}
		meta := LookupShenShaMeta(name)
		if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" || meta.Basis != "未登记可审计查法依据" {
			t.Errorf("removed %s metadata = %+v", name, meta)
		}
	}
}
