package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestTianSheAndSiFeiExactSeasonDayTables(t *testing.T) {
	tianSheByMonth := map[string]string{
		"寅": "戊寅", "卯": "戊寅", "辰": "戊寅",
		"巳": "甲午", "午": "甲午", "未": "甲午",
		"申": "戊申", "酉": "戊申", "戌": "戊申",
		"亥": "甲子", "子": "甲子", "丑": "甲子",
	}
	siFeiByMonth := map[string]string{
		"寅": "庚申", "卯": "庚申", "辰": "庚申",
		"巳": "壬子", "午": "壬子", "未": "壬子",
		"申": "甲寅", "酉": "甲寅", "戌": "甲寅",
		"亥": "丙午", "子": "丙午", "丑": "丙午",
	}
	for _, month := range data.Zhis {
		for dayIndex := 0; dayIndex < 60; dayIndex++ {
			day := data.Gans[dayIndex%10] + data.Zhis[dayIndex%12]
			if got, want := isTianShe(month, day), day == tianSheByMonth[month]; got != want {
				t.Errorf("isTianShe(%s, %s) = %v, want %v", month, day, got, want)
			}
			if got, want := isSiFei(month, day), day == siFeiByMonth[month]; got != want {
				t.Errorf("isSiFei(%s, %s) = %v, want %v", month, day, got, want)
			}
		}
	}
}

func TestSiFeiRejectsTheFormerDoubleDayExpansion(t *testing.T) {
	for _, tc := range []struct {
		month string
		day   string
	}{
		{month: "寅", day: "辛酉"},
		{month: "巳", day: "癸亥"},
		{month: "申", day: "乙卯"},
		{month: "亥", day: "丁巳"},
	} {
		if isSiFei(tc.month, tc.day) {
			t.Errorf("former expanded 四废 case %s月/%s日 still matches", tc.month, tc.day)
		}
		got := calcGlobalRuleFixture(t, model.Pillar{Gan: "甲", Zhi: "子"}, tc.month, tc.day, "庚午")
		if hasShenShaName(got.Global, "四废") {
			t.Errorf("former expanded 四废 case %s月/%s日 still published: %+v", tc.month, tc.day, got)
		}
	}
}

func TestTianLuoDiWangRequireNayinEligibilityAcrossSixtyYears(t *testing.T) {
	for yearIndex := 0; yearIndex < 60; yearIndex++ {
		year := model.Pillar{Gan: data.Gans[yearIndex%10], Zhi: data.Zhis[yearIndex%12]}
		nayin := data.Nayin[data.GanIndex(year.Gan)][data.ZhiIndex(year.Zhi)]
		element := data.NaYinMap[nayin].Element

		tianLuo := calcGlobalRuleFixture(t, year, "戌", "丁亥", "庚午")
		if got, want := hasShenShaName(tianLuo.Global, "天罗"), element == "火"; got != want {
			t.Errorf("year %s%s (%s/%s) 天罗 = %v, want %v: %+v", year.Gan, year.Zhi, nayin, element, got, want, tianLuo)
		}
		diWang := calcGlobalRuleFixture(t, year, "辰", "己巳", "庚午")
		if got, want := hasShenShaName(diWang.Global, "地网"), element == "水" || element == "土"; got != want {
			t.Errorf("year %s%s (%s/%s) 地网 = %v, want %v: %+v", year.Gan, year.Zhi, nayin, element, got, want, diWang)
		}
	}
}

func TestTianLuoDiWangRejectSingleBranches(t *testing.T) {
	for _, tc := range []struct {
		name  string
		year  model.Pillar
		month string
		day   string
		hour  string
	}{
		{name: "tian-luo-only-xu", year: model.Pillar{Gan: "丙", Zhi: "寅"}, month: "戌", day: "甲子", hour: "庚午"},
		{name: "tian-luo-only-hai", year: model.Pillar{Gan: "丙", Zhi: "寅"}, month: "子", day: "丁亥", hour: "庚午"},
		{name: "di-wang-only-chen", year: model.Pillar{Gan: "戊", Zhi: "申"}, month: "辰", day: "甲子", hour: "庚午"},
		{name: "di-wang-only-si", year: model.Pillar{Gan: "戊", Zhi: "申"}, month: "子", day: "己巳", hour: "庚午"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := calcGlobalRuleFixture(t, tc.year, tc.month, tc.day, tc.hour)
			for _, name := range []string{"天罗", "地网"} {
				if hasShenShaName(got.Global, name) {
					t.Errorf("single branch fixture produced %s: %+v", name, got)
				}
			}
		})
	}
}

func TestTianLuoDiWangPairCanSpanAnyPillars(t *testing.T) {
	tianLuo := calcGlobalRuleFixture(t, model.Pillar{Gan: "甲", Zhi: "戌"}, "亥", "甲子", "庚午")
	assertExactShenShaInBucket(t, "global", tianLuo.Global, "天罗：戌亥")
	diWang := calcGlobalRuleFixture(t, model.Pillar{Gan: "壬", Zhi: "辰"}, "巳", "甲子", "庚午")
	assertExactShenShaInBucket(t, "global", diWang.Global, "地网：辰巳")
}

func TestTianSheSiFeiAndLuoWangMetadataHasLocatedEvidence(t *testing.T) {
	for _, tc := range []struct {
		name      string
		citations []string
	}{
		{name: "天赦", citations: []string{"PDF第103-104页", "春戊寅", "冬甲子"}},
		{name: "四废", citations: []string{"PDF第113页", "春庚申", "每季只取一日"}},
		{name: "天罗", citations: []string{"PDF第117页", "PDF第119-120页", "纳音火", "戌亥成对"}},
		{name: "地网", citations: []string{"PDF第117页", "PDF第119-120页", "纳音水或土", "辰巳成对"}},
	} {
		meta := LookupShenShaMeta(tc.name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata = %+v", tc.name, meta)
		}
		for _, citation := range tc.citations {
			if !strings.Contains(meta.Basis, citation) {
				t.Errorf("%s basis = %q, want %q", tc.name, meta.Basis, citation)
			}
		}
	}
}

func calcGlobalRuleFixture(t testing.TB, year model.Pillar, monthBranch, dayPillar, hourPillar string) ShenShaCalcResult {
	t.Helper()
	got, err := CalcShenShaByPillars(ShenShaPillars{
		Year: year, Month: shenShaTestPillarForBranch(t, monthBranch),
		Day: parseGongTestPillar(t, dayPillar), Hour: parseGongTestPillar(t, hourPillar), Gender: "MALE",
	})
	if err != nil {
		t.Fatal(err)
	}
	return got
}
