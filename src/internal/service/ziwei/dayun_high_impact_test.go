package ziwei

import (
	"reflect"
	"sort"
	"testing"
)

func TestDayunUsesDecadalGanZhiForStarsAndFourHua(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2005, 6, 18, 10, 0, "女")
	if err != nil {
		t.Fatal(err)
	}

	stages := svc.CalculateDayun(chart)
	var target *DayunStage
	for i := range stages {
		if stages[i].GanZhi == "庚辰" {
			target = &stages[i]
			break
		}
	}
	if target == nil {
		t.Fatal("chart has no 庚辰 dayun stage")
	}
	if target.HeavenlyStem != "庚" || target.EarthlyBranch != "辰" {
		t.Fatalf("dayun stem-branch = %s/%s", target.HeavenlyStem, target.EarthlyBranch)
	}

	analysis := BuildDayunAnalysis(chart, stages, target.StartAge)
	if analysis == nil || analysis.GanZhi != "庚辰" {
		t.Fatalf("current dayun analysis = %+v", analysis)
	}
	var current *DayunStageAnalysis
	for i := range analysis.DayunStages {
		if analysis.DayunStages[i].Current {
			current = &analysis.DayunStages[i]
			break
		}
	}
	if current == nil {
		t.Fatal("analysis has no current dayun stage")
	}
	wantFourHua := []string{"太阳化禄", "武曲化权", "太阴化科", "天同化忌"}
	sort.Strings(wantFourHua)
	if !reflect.DeepEqual(current.FourHua, wantFourHua) {
		t.Fatalf("庚辰大限四化 = %v, want %v", current.FourHua, wantFourHua)
	}

	// Pinned iztro getHoroscopeStar("庚", "辰", "decadal") distribution.
	wantStars := map[string][]string{
		"寅": {"运马"},
		"卯": {"运曲"},
		"巳": {"运喜"},
		"未": {"运钺", "运陀"},
		"申": {"运禄"},
		"酉": {"运羊"},
		"亥": {"运昌", "运鸾"},
		"丑": {"运魁"},
	}
	gotStars := map[string][]string{}
	for _, focus := range analysis.FocusPalaces {
		if len(focus.PeriodStars) > 0 {
			gotStars[focus.Branch] = focus.PeriodStars
		}
	}
	if !reflect.DeepEqual(gotStars, wantStars) {
		t.Fatalf("庚辰大限运曜落宫 = %v, want %v", gotStars, wantStars)
	}
}

func TestTransitAnalysisUsesTargetYearDayunContext(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2000, 8, 16, 3, 0, "女")
	if err != nil {
		t.Fatal(err)
	}

	analysis2022 := BuildLiunianAnalysis(chart, svc.CalculateLiunian(chart, 2022), 2022)
	analysis2032 := BuildLiunianAnalysis(chart, svc.CalculateLiunian(chart, 2032), 2032)
	if analysis2022 == nil || analysis2022.DayunContext == nil ||
		analysis2032 == nil || analysis2032.DayunContext == nil {
		t.Fatal("target-year analysis omitted dayun context")
	}
	if got := analysis2022.DayunContext; got.NominalAge != 23 || got.StartAge != 23 || got.EndAge != 32 {
		t.Fatalf("2022 dayun context = %+v, want nominal age 23 in 23-32 stage", got)
	}
	if got := analysis2032.DayunContext; got.NominalAge != 33 || got.StartAge != 33 || got.EndAge != 42 {
		t.Fatalf("2032 dayun context = %+v, want nominal age 33 in 33-42 stage", got)
	}
	if analysis2022.DayunContext.GanZhi == analysis2032.DayunContext.GanZhi {
		t.Fatalf("target-year dayun did not advance: 2022=%s 2032=%s", analysis2022.DayunContext.GanZhi, analysis2032.DayunContext.GanZhi)
	}

	month := svc.CalculateLiuyueForDate(chart, 2022, 8, 19)
	monthAnalysis := BuildLiuyueAnalysis(chart, month, 2022, 8, 19)
	if monthAnalysis == nil || monthAnalysis.DayunContext == nil ||
		monthAnalysis.DayunContext.GanZhi != analysis2022.DayunContext.GanZhi {
		t.Fatalf("month and year layers use different dayun context: month=%+v year=%+v", monthAnalysis, analysis2022.DayunContext)
	}

	overlay := svc.AnalyzeLiunianOverlay(chart, svc.CalculateLiunian(chart, 2022), 2022)
	if overlay == nil || overlay.DayunContext == nil ||
		overlay.DayunContext.GanZhi != analysis2022.DayunContext.GanZhi {
		t.Fatalf("overlay omitted target-year dayun context: %+v", overlay)
	}
}
