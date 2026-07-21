package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestGongLuGongGuiUsesOnlyClassicalDayHourPairs(t *testing.T) {
	wants := []struct {
		day    string
		hour   string
		name   string
		target string
	}{
		{day: "癸亥", hour: "癸丑", name: "拱禄", target: "子"},
		{day: "癸丑", hour: "癸亥", name: "拱禄", target: "子"},
		{day: "丁巳", hour: "丁未", name: "拱禄", target: "午"},
		{day: "己未", hour: "己巳", name: "拱禄", target: "午"},
		{day: "戊辰", hour: "戊午", name: "拱禄", target: "巳"},
		{day: "甲申", hour: "甲戌", name: "拱贵", target: "酉"},
		{day: "乙未", hour: "乙酉", name: "拱贵", target: "申"},
		{day: "甲寅", hour: "甲子", name: "拱贵", target: "丑"},
		{day: "戊申", hour: "戊午", name: "拱贵", target: "未"},
		{day: "辛丑", hour: "辛卯", name: "拱贵", target: "寅"},
	}
	wantKeys := make(map[string]struct{}, len(wants))
	for _, want := range wants {
		wantKeys[want.day+"/"+want.hour] = struct{}{}
		rule, ok := gongLuGongGuiRule(want.day, want.hour)
		if !ok || rule.Name != want.name || rule.Target != want.target {
			t.Errorf("%s/%s rule = %+v, %v, want %s:%s", want.day, want.hour, rule, ok, want.name, want.target)
		}
		pillars := gongTestPillars(t, want.day, want.hour)
		got, err := CalcShenShaByPillars(pillars)
		if err != nil {
			t.Fatal(err)
		}
		assertExactShenShaInBucket(t, "global", got.Global, want.name+"："+want.target)
	}

	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		day := data.Gans[dayIndex%10] + data.Zhis[dayIndex%12]
		for hourIndex := 0; hourIndex < 60; hourIndex++ {
			hour := data.Gans[hourIndex%10] + data.Zhis[hourIndex%12]
			_, want := wantKeys[day+"/"+hour]
			_, got := gongLuGongGuiRule(day, hour)
			if got != want {
				t.Errorf("%s/%s registered = %v, want %v", day, hour, got, want)
			}
		}
	}
}

func TestGongLuGongGuiRejectsYearOrMonthFilling(t *testing.T) {
	for _, tc := range []struct {
		day    string
		hour   string
		name   string
		target string
	}{
		{day: "癸亥", hour: "癸丑", name: "拱禄", target: "子"},
		{day: "癸丑", hour: "癸亥", name: "拱禄", target: "子"},
		{day: "丁巳", hour: "丁未", name: "拱禄", target: "午"},
		{day: "己未", hour: "己巳", name: "拱禄", target: "午"},
		{day: "戊辰", hour: "戊午", name: "拱禄", target: "巳"},
		{day: "甲申", hour: "甲戌", name: "拱贵", target: "酉"},
		{day: "乙未", hour: "乙酉", name: "拱贵", target: "申"},
		{day: "甲寅", hour: "甲子", name: "拱贵", target: "丑"},
		{day: "戊申", hour: "戊午", name: "拱贵", target: "未"},
		{day: "辛丑", hour: "辛卯", name: "拱贵", target: "寅"},
	} {
		for _, filled := range []string{"year", "month"} {
			t.Run(tc.day+"-"+tc.hour+"-"+filled, func(t *testing.T) {
				pillars := gongTestPillars(t, tc.day, tc.hour)
				filler := gongTestPillarForBranch(t, tc.target)
				if filled == "year" {
					pillars.Year = filler
				} else {
					pillars.Month = filler
				}
				got, err := CalcShenShaByPillars(pillars)
				if err != nil {
					t.Fatal(err)
				}
				if hasShenShaName(got.Global, tc.name) {
					t.Fatalf("filled %s target %s still produced %s: %+v", filled, tc.target, tc.name, got)
				}
			})
		}
	}
}

func TestGenericForwardTwoBranchPairsDoNotCreateExtraGongRules(t *testing.T) {
	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		dayGan := data.Gans[dayIndex%10]
		dayZhiIndex := dayIndex % 12
		day := dayGan + data.Zhis[dayZhiIndex]
		hour := dayGan + data.Zhis[(dayZhiIndex+2)%12]
		if _, ok := gongLuGongGuiRule(day, hour); ok {
			continue
		}
		pillars := gongTestPillars(t, day, hour)
		got, err := CalcShenShaByPillars(pillars)
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"拱禄", "拱贵"} {
			if hasShenShaName(got.Global, name) {
				t.Errorf("non-classical pair %s/%s produced %s: %+v", day, hour, name, got)
			}
		}
	}
}

func TestGongLuGongGuiMetadataHasLocatedEvidence(t *testing.T) {
	for _, name := range []string{"拱禄", "拱贵"} {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata = %+v", name, meta)
		}
		for _, citation := range []string{"《三命通会》第181-182页", "《渊海子平》第546页", "不填实"} {
			if !strings.Contains(meta.Basis, citation) {
				t.Errorf("%s basis = %q, want %q", name, meta.Basis, citation)
			}
		}
	}
}

func gongTestPillars(t testing.TB, day, hour string) ShenShaPillars {
	t.Helper()
	return ShenShaPillars{
		Year: model.Pillar{Gan: "甲", Zhi: "辰"}, Month: model.Pillar{Gan: "甲", Zhi: "辰"},
		Day: parseGongTestPillar(t, day), Hour: parseGongTestPillar(t, hour), Gender: "MALE",
	}
}

func parseGongTestPillar(t testing.TB, value string) model.Pillar {
	t.Helper()
	runes := []rune(value)
	if len(runes) != 2 {
		t.Fatalf("invalid test pillar %q", value)
	}
	return model.Pillar{Gan: string(runes[0]), Zhi: string(runes[1])}
}

func gongTestPillarForBranch(t testing.TB, zhi string) model.Pillar {
	t.Helper()
	for i := 0; i < 60; i++ {
		if data.Zhis[i%12] == zhi {
			return model.Pillar{Gan: data.Gans[i%10], Zhi: zhi}
		}
	}
	t.Fatalf("no sixty-cycle pillar for branch %q", zhi)
	return model.Pillar{}
}
