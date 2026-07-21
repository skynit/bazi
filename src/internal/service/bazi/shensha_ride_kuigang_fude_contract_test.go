package bazi

import (
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestRiDeKuiGangFuDeXiuQiExactTablesAcrossSixtyCycle(t *testing.T) {
	wants := map[string]map[string]bool{
		"日德": {
			"甲寅": true, "丙辰": true, "戊辰": true, "庚辰": true, "壬戌": true,
		},
		"魁罡": {
			"庚辰": true, "壬辰": true, "戊戌": true, "庚戌": true,
		},
		"福德秀气": {
			"乙巳": true, "乙酉": true, "乙丑": true,
			"丁巳": true, "丁酉": true, "丁丑": true,
			"己巳": true, "己酉": true, "己丑": true,
			"辛巳": true, "辛酉": true, "辛丑": true,
			"癸巳": true, "癸酉": true, "癸丑": true,
		},
	}
	wantCounts := map[string]int{"日德": 5, "魁罡": 4, "福德秀气": 15}
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

func TestYiMaoIsNotFuDeXiuQiAndKeepsIndependentRules(t *testing.T) {
	result := calcSpecialDayFixture(t, "乙卯")
	assertShenShaNameAbsentEverywhere(t, result, "福德秀气")
	for _, want := range []string{"九丑日", "八专"} {
		if !hasShenShaName(result.Day, want) {
			t.Errorf("day 乙卯 lost independent %s: %+v", want, result)
		}
	}
}

func TestRiDeAndKuiGangPatternCandidatesMatchClassicalDayTables(t *testing.T) {
	ride := map[string]bool{"甲寅": true, "丙辰": true, "戊辰": true, "庚辰": true, "壬戌": true}
	kuigang := map[string]bool{"庚辰": true, "壬辰": true, "戊戌": true, "庚戌": true}

	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		day := data.Gans[dayIndex%10] + data.Zhis[dayIndex%12]
		dayPillar := parseGongTestPillar(t, day)
		pillars := []model.Pillar{
			{Gan: "甲", Zhi: "子"}, {Gan: "丙", Zhi: "寅"}, dayPillar, {Gan: "庚", Zhi: "午"},
		}
		analysis := AnalyzePatternExtended(pillars, "寅")

		if got := hasPatternRuleID(analysis.Candidates, "pattern.aux.ride"); got != ride[day] {
			t.Errorf("pattern day %s 日德 = %v, want %v: %+v", day, got, ride[day], analysis.Candidates)
		}
		if got := hasPatternRuleID(analysis.Candidates, "pattern.aux.kuigang"); got != kuigang[day] {
			t.Errorf("pattern day %s 魁罡 = %v, want %v: %+v", day, got, kuigang[day], analysis.Candidates)
		}
		for _, candidate := range analysis.Candidates {
			if candidate.RuleID == "pattern.aux.kuigang" && candidate.PatternName != "魁罡格" {
				t.Errorf("pattern day %s 魁罡 name = %q, want neutral 魁罡格", day, candidate.PatternName)
			}
		}
	}
}

func TestRiDeKuiGangFuDeMetadataIsLocatedButNotAdjudicated(t *testing.T) {
	wants := map[string][]string{
		"日德":   {"甲寅、丙辰、戊辰、庚辰、壬戌", "PDF第185-186页", "书内第182-183页", "《渊海子平》PDF第516页"},
		"魁罡":   {"庚辰、壬辰、戊戌、庚戌", "PDF第186-187页", "书内第183-184页"},
		"福德秀气": {"十五日", "辛组三日", "《渊海子平》PDF第209页", "PDF第188-189页", "书内第185-186页"},
	}
	for name, citations := range wants {
		meta := LookupShenShaMeta(name)
		if meta.Status != "observed" || meta.InterpretationStatus != "not_adjudicated" {
			t.Errorf("%s metadata status = %+v", name, meta)
		}
		for _, citation := range citations {
			if !strings.Contains(meta.Basis, citation) {
				t.Errorf("%s basis = %q, want %q", name, meta.Basis, citation)
			}
		}
	}
}

func hasPatternRuleID(candidates []PatternCandidate, ruleID string) bool {
	for _, candidate := range candidates {
		if candidate.RuleID == ruleID {
			return true
		}
	}
	return false
}
