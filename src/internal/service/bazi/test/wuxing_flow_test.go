package bazi_test

import (
	"strings"
	"testing"

	. "bazi/internal/service/bazi"
)

type wuxingFlowCase struct {
	ID, Desc, Source                string
	YearP, MonP, DayP, HouP, Gender string
}

var wuxingFlowCases = []wuxingFlowCase{
	{ID: "WX-001", Desc: "甲日申月", YearP: "壬辰", MonP: "戊申", DayP: "甲午", HouP: "庚午", Gender: "MALE"},
	{ID: "WX-002", Desc: "丙午日午月", YearP: "丁未", MonP: "丙午", DayP: "丙午", HouP: "甲午", Gender: "MALE"},
	{ID: "WX-003", Desc: "壬子日子月", YearP: "壬子", MonP: "壬子", DayP: "壬子", HouP: "庚子", Gender: "MALE"},
	{ID: "WX-004", Desc: "五行较全", YearP: "甲午", MonP: "己巳", DayP: "戊戌", HouP: "庚申", Gender: "MALE"},
	{ID: "WX-005", Desc: "庚申日申月", YearP: "庚申", MonP: "甲申", DayP: "庚申", HouP: "庚辰", Gender: "MALE"},
	{ID: "WX-006", Desc: "甲寅日寅月", YearP: "甲寅", MonP: "丙寅", DayP: "甲寅", HouP: "丙寅", Gender: "MALE"},
	{ID: "WX-007", Desc: "己未日酉月", YearP: "己未", MonP: "癸酉", DayP: "己未", HouP: "甲子", Gender: "MALE"},
}

func TestFiveElementDerivedEvidenceStructuralInvariants(t *testing.T) {
	service := &BaziService{}
	for _, tc := range wuxingFlowCases {
		t.Run(tc.ID, func(t *testing.T) {
			result, err := service.CalculateSyntheticPillars(tc.YearP, tc.MonP, tc.DayP, tc.HouP, tc.Gender)
			if err != nil {
				t.Fatal(err)
			}
			if result.MissingElements.Status != "observed" || result.MissingElements.RuleID == "" {
				t.Fatalf("incomplete missing-element evidence: %+v", result.MissingElements)
			}
			if result.MissingElements.IsYongshenConclusion || result.MissingElements.RemedyStatus != "not_adjudicated" {
				t.Fatalf("raw distribution became a remedy conclusion: %+v", result.MissingElements)
			}
			if len(result.MissingElements.MissingElements)+len(result.MissingElements.WeakElements) > 0 {
				if !strings.Contains(result.MissingElements.Note, "不等于喜用神") || strings.Contains(result.MissingElements.Note, "需注意补足") || strings.Contains(result.MissingElements.Note, "宜加强") {
					t.Fatalf("missing-element note promotes raw distribution to a remedy: %q", result.MissingElements.Note)
				}
			}
		})
	}
}

func TestFiveElementEvidenceCoversAllDayElements(t *testing.T) {
	service := &BaziService{}
	variations := []struct{ year, month, day, hour, gender string }{
		{"甲子", "丙寅", "甲子", "庚午", "MALE"},
		{"乙丑", "丁卯", "乙丑", "辛未", "FEMALE"},
		{"丙寅", "戊辰", "丙寅", "壬申", "MALE"},
		{"丁卯", "己巳", "丁卯", "癸酉", "FEMALE"},
		{"戊辰", "庚午", "戊辰", "甲戌", "MALE"},
		{"己巳", "辛未", "己巳", "乙亥", "FEMALE"},
		{"庚午", "壬申", "庚午", "丙子", "MALE"},
		{"辛未", "癸酉", "辛未", "丁丑", "FEMALE"},
		{"壬申", "甲戌", "壬申", "戊寅", "MALE"},
		{"癸酉", "乙亥", "癸酉", "己卯", "FEMALE"},
	}
	for _, tc := range variations {
		result, err := service.CalculateSyntheticPillars(tc.year, tc.month, tc.day, tc.hour, tc.gender)
		if err != nil {
			t.Fatal(err)
		}
		if len(result.FiveElements) != 5 || len(result.ElementDetail) != 5 || result.MissingElements.Status != "observed" {
			t.Fatalf("incomplete five-element evidence for %v: %+v", tc, result)
		}
	}
}
