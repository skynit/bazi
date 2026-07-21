package bazi

import (
	"os"
	"strings"
	"testing"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

func TestClassicalSanQiSequenceExhaustiveFourStemWindows(t *testing.T) {
	matched := 0
	for _, yearGan := range data.Gans {
		for _, monthGan := range data.Gans {
			for _, dayGan := range data.Gans {
				for _, hourGan := range data.Gans {
					gans := []string{yearGan, monthGan, dayGan, hourGan}
					want := ""
					left := yearGan + monthGan + dayGan
					right := monthGan + dayGan + hourGan
					if left == "乙丙丁" || left == "甲戊庚" {
						want = left
					} else if right == "乙丙丁" || right == "甲戊庚" {
						want = right
					}
					if got := classicalSanQiSequence(gans); got != want {
						t.Errorf("classicalSanQiSequence(%v) = %q, want %q", gans, got, want)
					}
					if want != "" {
						matched++
					}
				}
			}
		}
	}
	if matched != 40 {
		t.Fatalf("four-stem 三奇 window count = %d, want 40", matched)
	}
}

func TestClassicalSanQiRejectsNonFourPillarWindows(t *testing.T) {
	for _, gans := range [][]string{
		{"乙", "丙", "丁"},
		{"甲", "甲", "乙", "丙", "丁"},
		{"甲", "戊", "庚", "甲", "甲"},
	} {
		if got := classicalSanQiSequence(gans); got != "" {
			t.Errorf("non-four-pillar 三奇 window %v matched %q", gans, got)
		}
	}
}

func TestSanQiShenShaAndPatternConditionsAgree(t *testing.T) {
	for _, tc := range []struct {
		name string
		gans []string
		want bool
	}{
		{name: "left-jia-wu-geng", gans: []string{"甲", "戊", "庚", "甲"}, want: true},
		{name: "right-yi-bing-ding", gans: []string{"甲", "乙", "丙", "丁"}, want: true},
		{name: "reversed-jia-wu-geng", gans: []string{"庚", "戊", "甲", "甲"}},
		{name: "reversed-yi-bing-ding", gans: []string{"丁", "丙", "乙", "甲"}},
		{name: "old-ren-gui-xin", gans: []string{"壬", "癸", "辛", "甲"}},
		{name: "disputed-xin-ren-gui", gans: []string{"辛", "壬", "癸", "甲"}},
		{name: "disputed-gui-ren-xin", gans: []string{"癸", "壬", "辛", "甲"}},
		{name: "non-adjacent", gans: []string{"甲", "乙", "戊", "庚"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			pillars := testPillarsForGans(t, tc.gans)
			got, err := CalcShenShaByPillars(ShenShaPillars{
				Year: pillars[0], Month: pillars[1], Day: pillars[2], Hour: pillars[3], Gender: "MALE",
			})
			if err != nil {
				t.Fatal(err)
			}
			if gotHit := hasShenShaName(got.Global, "三奇"); gotHit != tc.want {
				t.Errorf("三奇 shen-sha hit = %v, want %v: %+v", gotHit, tc.want, got)
			}
			if patternHit := checkSanQi(pillars) != nil; patternHit != tc.want {
				t.Errorf("三奇 pattern hit = %v, want %v for %v", patternHit, tc.want, pillars)
			}
			for _, oldName := range []string{"天三奇", "地三奇", "人三奇"} {
				if hasShenShaName(got.Global, oldName) {
					t.Errorf("legacy taxonomy %s still published: %+v", oldName, got)
				}
			}
		})
	}
}

func TestSanQiPatternCandidateIsAuxiliary(t *testing.T) {
	pillars := testPillarsForGans(t, []string{"甲", "乙", "丙", "丁"})
	analysis := AnalyzePatternExtended(pillars, pillars[1].Zhi)
	candidate := findPatternCandidate(t, analysis.Candidates, "三奇")
	if candidate.RuleID != "pattern.aux.sanqi" || candidate.Category != patternCategoryAuxiliary {
		t.Fatalf("三奇 candidate = %+v", candidate)
	}
}

func TestJinShenOnlyMatchesTheThreeLocatedHourPillars(t *testing.T) {
	wants := map[string]bool{"癸酉": true, "己巳": true, "乙丑": true}
	for index := 0; index < 60; index++ {
		hour := model.Pillar{Gan: data.Gans[index%10], Zhi: data.Zhis[index%12]}
		hourName := hour.Gan + hour.Zhi
		if got, want := isJinShenHourPillar(hour), wants[hourName]; got != want {
			t.Errorf("isJinShenHourPillar(%s) = %v, want %v", hourName, got, want)
		}
		result, err := CalcShenShaByPillars(ShenShaPillars{
			Year: model.Pillar{Gan: "甲", Zhi: "辰"}, Month: model.Pillar{Gan: "丙", Zhi: "寅"},
			Day: model.Pillar{Gan: "甲", Zhi: "子"}, Hour: hour, Gender: "FEMALE",
		})
		if err != nil {
			t.Fatal(err)
		}
		if got, want := hasShenShaName(result.Hour, "金神"), wants[hourName]; got != want {
			t.Errorf("hour %s 金神 = %v, want %v: %+v", hourName, got, want, result)
		}
		for _, bucket := range [][]string{result.Year, result.Month, result.Day, result.Global} {
			if hasShenShaName(bucket, "金神") || hasShenShaName(bucket, "金神成格") {
				t.Errorf("hour %s leaked 金神 outside hour bucket: %+v", hourName, result)
			}
		}
	}
}

func TestJinShenDayWithoutJinShenHourDoesNotMatch(t *testing.T) {
	for _, day := range []string{"癸酉", "己巳", "乙丑"} {
		pillars := ShenShaPillars{
			Year: model.Pillar{Gan: "甲", Zhi: "辰"}, Month: model.Pillar{Gan: "丙", Zhi: "寅"},
			Day: parseGongTestPillar(t, day), Hour: model.Pillar{Gan: "庚", Zhi: "午"}, Gender: "MALE",
		}
		got, err := CalcShenShaByPillars(pillars)
		if err != nil {
			t.Fatal(err)
		}
		assertShenShaNameAbsentEverywhere(t, got, "金神")
		if checkJinShenHour([]model.Pillar{pillars.Year, pillars.Month, pillars.Day, pillars.Hour}) != nil {
			t.Errorf("day-only 金神 fixture %s created pattern candidate", day)
		}
	}
}

func TestJinShenPatternRequiresCompleteFourPillarHourPosition(t *testing.T) {
	valid := []model.Pillar{
		{Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"},
		{Gan: "甲", Zhi: "子"}, {Gan: "癸", Zhi: "酉"},
	}
	if got := checkJinShenHour(valid); got == nil || got.PatternName != "金神" {
		t.Fatalf("complete hour-position 金神 = %+v", got)
	}
	for _, pillars := range [][]model.Pillar{
		valid[:3],
		append(append([]model.Pillar(nil), valid...), model.Pillar{Gan: "甲", Zhi: "子"}),
	} {
		if got := checkJinShenHour(pillars); got != nil {
			t.Errorf("non-four-pillar 金神 matched: %+v for %+v", got, pillars)
		}
	}
}

func TestJinShenPatternCandidateIsAuxiliaryForAllThreeHours(t *testing.T) {
	for _, hour := range []model.Pillar{
		{Gan: "癸", Zhi: "酉"}, {Gan: "己", Zhi: "巳"}, {Gan: "乙", Zhi: "丑"},
	} {
		pillars := []model.Pillar{
			{Gan: "甲", Zhi: "辰"}, {Gan: "丙", Zhi: "寅"}, {Gan: "甲", Zhi: "子"}, hour,
		}
		analysis := AnalyzePatternExtended(pillars, "寅")
		candidate := findPatternCandidate(t, analysis.Candidates, "金神")
		if candidate.RuleID != "pattern.aux.jinshen" || candidate.Category != patternCategoryAuxiliary {
			t.Errorf("hour %s%s 金神 candidate = %+v", hour.Gan, hour.Zhi, candidate)
		}
	}
}

func TestRemovedSanQiTaxonomyAndJinShenChengGeShortcutsAreAbsent(t *testing.T) {
	for _, path := range []string{"shensha.go", "pattern.go"} {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{`"天三奇"`, `"地三奇"`, `"人三奇"`, `"金神成格"`, "checkJinShenGe", "checkSanQiGe"} {
			if strings.Contains(string(source), forbidden) {
				t.Errorf("%s still contains removed shortcut %q", path, forbidden)
			}
		}
	}
	meta := LookupShenShaMeta("金神成格")
	if meta.Status != "unregistered" || meta.InterpretationStatus != "not_available" {
		t.Fatalf("removed 金神成格 metadata = %+v", meta)
	}
}

func TestSanQiAndJinShenMetadataHasLocatedEvidence(t *testing.T) {
	for _, tc := range []struct {
		name      string
		citations []string
	}{
		{name: "三奇", citations: []string{"PDF第100-102页", "乙丙丁", "甲戊庚", "倒序", "人间三奇"}},
		{name: "金神", citations: []string{"PDF第221页", "时柱", "癸酉", "己巳", "乙丑", "未裁决"}},
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

func testPillarsForGans(t testing.TB, gans []string) []model.Pillar {
	t.Helper()
	if len(gans) != 4 {
		t.Fatalf("gan fixture length = %d, want 4", len(gans))
	}
	pillars := make([]model.Pillar, 4)
	for i, gan := range gans {
		ganIndex := data.GanIndex(gan)
		if ganIndex < 0 {
			t.Fatalf("unknown test gan %q", gan)
		}
		pillars[i] = model.Pillar{Gan: gan, Zhi: data.Zhis[ganIndex%2]}
	}
	return pillars
}

func balancedPatternScores() map[string]int {
	return map[string]int{"木": 20, "火": 20, "土": 20, "金": 20, "水": 20}
}
