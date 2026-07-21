package bazi

import (
	"strings"
	"testing"

	"bazi/internal/service/data"
)

func TestClassicalSpecialDayTablesExactAcrossSixtyCycle(t *testing.T) {
	wants := map[string]map[string]bool{
		"九丑日": {
			"壬子": true, "壬午": true, "戊子": true, "戊午": true, "己酉": true,
			"己卯": true, "乙卯": true, "辛酉": true, "辛卯": true,
		},
		"八专": {
			"甲寅": true, "乙卯": true, "己未": true, "丁未": true,
			"庚申": true, "辛酉": true, "戊戌": true, "癸丑": true,
		},
		"孤鸾煞": {
			"乙巳": true, "丁巳": true, "辛亥": true, "戊申": true,
			"甲寅": true, "丙午": true, "戊午": true, "壬子": true,
		},
		"阴差阳错": {
			"丙子": true, "丁丑": true, "戊寅": true, "辛卯": true,
			"壬辰": true, "癸巳": true, "丙午": true, "丁未": true,
			"戊申": true, "辛酉": true, "壬戌": true, "癸亥": true,
		},
	}
	wantCounts := map[string]int{"九丑日": 9, "八专": 8, "孤鸾煞": 8, "阴差阳错": 12}
	matched := map[string]int{}

	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		day := data.Gans[dayIndex%10] + data.Zhis[dayIndex%12]
		result := calcSpecialDayFixture(t, day)
		for name, days := range wants {
			want := days[day]
			if got := specialDayContainsName(day, name); got != want {
				t.Errorf("special day %s %s = %v, want %v", day, name, got, want)
			}
			if got := hasShenShaName(result.Day, name); got != want {
				t.Errorf("formal day %s %s = %v, want %v: %+v", day, name, got, want, result)
			}
			for _, bucket := range [][]string{result.Year, result.Month, result.Hour, result.Global} {
				if hasShenShaName(bucket, name) {
					t.Errorf("day %s leaked %s outside day bucket: %+v", day, name, result)
				}
			}
			if want {
				matched[name]++
			}
		}
	}

	for name, want := range wantCounts {
		if got := matched[name]; got != want {
			t.Errorf("%s matched days = %d, want %d", name, got, want)
		}
	}
}

func TestRemovedClassicalSpecialDayFalsePositivesPreserveIndependentRules(t *testing.T) {
	for _, day := range []string{"戊子", "戊午"} {
		if specialDayContainsName(day, "阴差阳错") {
			t.Errorf("special day table still marks %s as 阴差阳错", day)
		}
		result := calcSpecialDayFixture(t, day)
		assertShenShaNameAbsentEverywhere(t, result, "阴差阳错")
		if !hasShenShaName(result.Day, "九丑日") {
			t.Errorf("day %s lost independent 九丑日: %+v", day, result)
		}
		if day == "戊午" && !hasShenShaName(result.Day, "孤鸾煞") {
			t.Errorf("day 戊午 lost independent 孤鸾煞: %+v", result)
		}
	}

	gengXu := calcSpecialDayFixture(t, "庚戌")
	if hasShenShaName(gengXu.Day, "八专") {
		t.Errorf("day 庚戌 still has 八专: %+v", gengXu)
	}
	if !hasShenShaName(gengXu.Day, "魁罡") {
		t.Errorf("day 庚戌 lost independent 魁罡: %+v", gengXu)
	}
}

func TestClassicalSpecialDayMetadataIsLocatedButNotAdjudicated(t *testing.T) {
	for _, name := range []string{"九丑日", "八专", "孤鸾煞", "阴差阳错"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata status = %+v", name, meta)
		}
		for _, citation := range []string{"《三命通会》", "PDF第124页", "书内第121页"} {
			if !strings.Contains(meta.Basis, citation) {
				t.Errorf("%s basis = %q, want %q", name, meta.Basis, citation)
			}
		}
	}
}
