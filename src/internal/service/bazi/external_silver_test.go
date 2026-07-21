package bazi

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"
)

type externalSilverFixture struct {
	Version  string                 `json:"version"`
	Metadata externalSilverMetadata `json:"metadata"`
	Sources  []externalSilverSource `json:"sources"`
	Cases    []externalSilverCase   `json:"cases"`
}

type externalSilverMetadata struct {
	Tier                string   `json:"tier"`
	Purpose             string   `json:"purpose"`
	ReviewStatus        string   `json:"review_status"`
	PublishableAccuracy bool     `json:"publishable_accuracy"`
	Generator           string   `json:"generator"`
	DualConsensusCases  int      `json:"dual_consensus_cases"`
	BoundaryCases       int      `json:"boundary_cases"`
	BoundaryGroups      int      `json:"boundary_groups"`
	ZiPolicyCases       int      `json:"zi_policy_cases"`
	DisputedCases       int      `json:"disputed_cases"`
	Limitations         []string `json:"limitations"`
}

type externalSilverSource struct {
	ID         string            `json:"id"`
	Repository string            `json:"repository"`
	Commit     string            `json:"commit"`
	License    string            `json:"license"`
	Files      map[string]string `json:"files"`
}

type externalSilverCase struct {
	ID              string               `json:"id"`
	Scope           string               `json:"scope"`
	Term            string               `json:"term"`
	Input           externalSilverInput  `json:"input"`
	Consensus       externalConsensus    `json:"consensus"`
	LunarJavascript externalSilverResult `json:"lunar_javascript"`
	CNLunar         externalSilverResult `json:"cnlunar"`
}

type externalSilverInput struct {
	Year         int    `json:"year"`
	Month        int    `json:"month"`
	Day          int    `json:"day"`
	Hour         int    `json:"hour"`
	Minute       int    `json:"minute"`
	Second       int    `json:"second"`
	Gender       string `json:"gender"`
	ZiHourPolicy string `json:"zi_hour_policy"`
}

type externalConsensus struct {
	PillarsAgree bool `json:"pillars_agree"`
	Admitted     bool `json:"admitted"`
}

type externalSilverResult struct {
	Precision    string                `json:"precision"`
	ZiHourPolicy string                `json:"zi_hour_policy"`
	Pillars      externalSilverPillars `json:"pillars"`
	MingGong     string                `json:"ming_gong"`
	NaYin        externalSilverPillars `json:"nayin"`
	Dayun        externalSilverDayun   `json:"dayun"`
}

type externalSilverPillars struct {
	Year  string `json:"year"`
	Month string `json:"month"`
	Day   string `json:"day"`
	Hour  string `json:"hour"`
}

type externalSilverDayun struct {
	Direction string   `json:"direction"`
	Pillars   []string `json:"pillars"`
}

func TestExternalSilverStructuralDifferential(t *testing.T) {
	fixture := loadExternalSilverFixture(t)
	assertExternalSilverMetadata(t, fixture)

	service := &BaziService{}
	seen := make(map[string]bool, len(fixture.Cases))
	termCounts := make(map[string]int)
	boundaryGroups := make(map[string][]externalSilverCase)
	counts := make(map[string]int)
	ziPolicyCounts := make(map[string]int)
	ziPolicyResults := make(map[string]externalSilverResult)
	for _, tc := range fixture.Cases {
		if tc.ID == "" || seen[tc.ID] {
			t.Fatalf("empty or duplicate case id %q", tc.ID)
		}
		seen[tc.ID] = true
		counts[tc.Scope]++

		switch tc.Scope {
		case "dual_consensus":
			if !tc.Consensus.PillarsAgree || !tc.Consensus.Admitted || tc.LunarJavascript.Pillars != tc.CNLunar.Pillars {
				t.Fatalf("%s is not a valid dual-source consensus case: %+v", tc.ID, tc.Consensus)
			}
			if normalizeExternalNaYin(tc.LunarJavascript.NaYin) != tc.CNLunar.NaYin {
				t.Fatalf("%s sources disagree on normalized nayin: lunar=%+v cnlunar=%+v", tc.ID, tc.LunarJavascript.NaYin, tc.CNLunar.NaYin)
			}
		case "lunar_js_jie_boundary":
			if tc.Term == "" || tc.Consensus.Admitted {
				t.Fatalf("%s has invalid boundary metadata", tc.ID)
			}
			termCounts[tc.Term]++
			group := fmt.Sprintf("%s/%d", tc.Term, tc.Input.Year)
			boundaryGroups[group] = append(boundaryGroups[group], tc)
		case "lunar_js_zi_boundary":
			if tc.Term != "" || tc.Consensus.Admitted || tc.Input.ZiHourPolicy != tc.LunarJavascript.ZiHourPolicy {
				t.Fatalf("%s has invalid late-Zi boundary metadata", tc.ID)
			}
			ziPolicyCounts[tc.Input.ZiHourPolicy]++
			ziPolicyResults[tc.ID] = tc.LunarJavascript
		case "upstream_disputed":
			if tc.Consensus.PillarsAgree || tc.Consensus.Admitted || tc.LunarJavascript.Pillars == tc.CNLunar.Pillars {
				t.Fatalf("%s no longer records an upstream dispute", tc.ID)
			}
			continue
		default:
			t.Fatalf("%s has unknown scope %q", tc.ID, tc.Scope)
		}

		t.Run(tc.ID, func(t *testing.T) {
			got, err := service.CalculateAtWithPolicy(
				tc.Input.Year, tc.Input.Month, tc.Input.Day,
				tc.Input.Hour, tc.Input.Minute, tc.Input.Second,
				tc.Input.Gender, tc.Input.ZiHourPolicy,
			)
			if err != nil {
				t.Fatal(err)
			}
			want := tc.LunarJavascript
			gotPillars := externalSilverPillars{
				Year:  pillarName(got.YearPillar.Gan, got.YearPillar.Zhi),
				Month: pillarName(got.MonthPillar.Gan, got.MonthPillar.Zhi),
				Day:   pillarName(got.DayPillar.Gan, got.DayPillar.Zhi),
				Hour:  pillarName(got.HourPillar.Gan, got.HourPillar.Zhi),
			}
			if gotPillars != want.Pillars {
				t.Fatalf("pillars = %+v, want %+v", gotPillars, want.Pillars)
			}
			if got.MingGong.GanZhi != want.MingGong {
				t.Fatalf("ming_gong = %q, want %q", got.MingGong.GanZhi, want.MingGong)
			}
			gotNaYin := externalSilverPillars{
				Year: got.NaYin["year"].Name, Month: got.NaYin["month"].Name,
				Day: got.NaYin["day"].Name, Hour: got.NaYin["hour"].Name,
			}
			wantNaYin := normalizeExternalNaYin(want.NaYin)
			if gotNaYin != wantNaYin {
				t.Fatalf("nayin = %+v, want %+v (raw lunar-javascript=%+v)", gotNaYin, wantNaYin, want.NaYin)
			}
			gotDirection := "reverse"
			if got.DaYunInfo.Direction == "顺行" {
				gotDirection = "forward"
			}
			if gotDirection != want.Dayun.Direction {
				t.Fatalf("dayun direction = %q, want %q", gotDirection, want.Dayun.Direction)
			}
			gotDayun := make([]string, 0, len(got.DaYunInfo.Pillars))
			for _, pillar := range got.DaYunInfo.Pillars {
				gotDayun = append(gotDayun, pillarName(pillar.Gan, pillar.Zhi))
			}
			if !reflect.DeepEqual(gotDayun, want.Dayun.Pillars) {
				t.Fatalf("dayun pillars = %v, want %v", gotDayun, want.Dayun.Pillars)
			}
		})
	}

	if counts["dual_consensus"] != fixture.Metadata.DualConsensusCases ||
		counts["lunar_js_jie_boundary"] != fixture.Metadata.BoundaryCases ||
		counts["lunar_js_zi_boundary"] != fixture.Metadata.ZiPolicyCases ||
		counts["upstream_disputed"] != fixture.Metadata.DisputedCases {
		t.Fatalf("scope counts = %+v, metadata = %+v", counts, fixture.Metadata)
	}
	if len(termCounts) != 12 {
		t.Fatalf("boundary terms = %d, want 12: %+v", len(termCounts), termCounts)
	}
	for term, count := range termCounts {
		want := 2
		if term == "立春" {
			want = 8
		}
		if count != want {
			t.Fatalf("boundary term %s has %d cases, want %d", term, count, want)
		}
	}
	if len(boundaryGroups) != fixture.Metadata.BoundaryGroups || len(boundaryGroups) != 15 {
		t.Fatalf("boundary groups = %d, metadata=%d, want 15", len(boundaryGroups), fixture.Metadata.BoundaryGroups)
	}
	assertExternalJieBoundaryContract(t, boundaryGroups)
	if ziPolicyCounts[ZiHourLateZiNextDay] != 4 || ziPolicyCounts[ZiHourLateZiSameDay] != 4 {
		t.Fatalf("late-Zi Silver policy counts = %+v, want four cases per policy", ziPolicyCounts)
	}
	assertExternalZiBoundaryContract(t, ziPolicyResults)
}

func assertExternalJieBoundaryContract(t testing.TB, casesByGroup map[string][]externalSilverCase) {
	t.Helper()
	for group, cases := range casesByGroup {
		if len(cases) != 2 {
			t.Fatalf("boundary group %s has %d cases, want 2", group, len(cases))
		}
		term := cases[0].Term
		before, at := cases[0], cases[1]
		beforeTime := externalSilverInputTime(before.Input)
		atTime := externalSilverInputTime(at.Input)
		if atTime.Before(beforeTime) {
			before, at = at, before
			beforeTime, atTime = atTime, beforeTime
		}
		if atTime.Sub(beforeTime) != time.Second {
			t.Fatalf("boundary group %s inputs are not adjacent seconds: before=%+v at=%+v", group, before.Input, at.Input)
		}
		if before.Input.Gender != at.Input.Gender || before.Input.ZiHourPolicy != at.Input.ZiHourPolicy ||
			before.LunarJavascript.Precision != "second_level_solar_terms" ||
			at.LunarJavascript.Precision != "second_level_solar_terms" {
			t.Fatalf("boundary group %s metadata diverges: before=%+v at=%+v", group, before, at)
		}

		beforePillars, atPillars := before.LunarJavascript.Pillars, at.LunarJavascript.Pillars
		if beforePillars.Month == atPillars.Month || before.LunarJavascript.MingGong == at.LunarJavascript.MingGong {
			t.Fatalf("boundary group %s does not cross a month/MingGong boundary: before=%+v at=%+v", group, beforePillars, atPillars)
		}
		if beforePillars.Day != atPillars.Day || beforePillars.Hour != atPillars.Hour {
			t.Fatalf("boundary group %s changed day/hour pillars across one second: before=%+v at=%+v", group, beforePillars, atPillars)
		}
		if (term == "立春") != (beforePillars.Year != atPillars.Year) {
			t.Fatalf("boundary group %s has invalid year-pillar boundary: before=%+v at=%+v", group, beforePillars, atPillars)
		}
	}
}

func externalSilverInputTime(input externalSilverInput) time.Time {
	return time.Date(input.Year, time.Month(input.Month), input.Day, input.Hour, input.Minute, input.Second, 0, time.UTC)
}

func assertExternalZiBoundaryContract(t testing.TB, cases map[string]externalSilverResult) {
	t.Helper()
	if len(cases) != 8 {
		t.Fatalf("late-Zi Silver cases = %d, want 8", len(cases))
	}
	result := func(id string) externalSilverResult {
		item, ok := cases[id]
		if !ok {
			t.Fatalf("missing late-Zi Silver case %q", id)
		}
		return item
	}

	nextBefore, nextAt := result("zi_next_225959"), result("zi_next_230000")
	sameBefore, sameAt := result("zi_same_225959"), result("zi_same_230000")
	if nextBefore.Pillars.Day == nextAt.Pillars.Day {
		t.Fatal("late_zi_next_day did not change the day pillar at 23:00")
	}
	if sameBefore.Pillars.Day != sameAt.Pillars.Day {
		t.Fatal("late_zi_same_day changed the day pillar at 23:00")
	}
	if nextAt.Pillars.Day == sameAt.Pillars.Day || nextAt.Pillars.Hour != sameAt.Pillars.Hour {
		t.Fatalf("late-Zi policies did not separate only the day pillar: next=%+v same=%+v", nextAt.Pillars, sameAt.Pillars)
	}

	nextLate, nextMidnight := result("zi_next_235959"), result("zi_next_000000")
	sameLate, sameMidnight := result("zi_same_235959"), result("zi_same_000000")
	if nextLate.Pillars != nextMidnight.Pillars {
		t.Fatalf("late_zi_next_day did not converge with civil midnight: late=%+v midnight=%+v", nextLate.Pillars, nextMidnight.Pillars)
	}
	if sameLate.Pillars.Day == sameMidnight.Pillars.Day || sameLate.Pillars.Hour != sameMidnight.Pillars.Hour {
		t.Fatalf("late_zi_same_day civil-midnight boundary is invalid: late=%+v midnight=%+v", sameLate.Pillars, sameMidnight.Pillars)
	}
	if nextMidnight.Pillars != sameMidnight.Pillars {
		t.Fatalf("late-Zi policies did not agree after midnight: next=%+v same=%+v", nextMidnight.Pillars, sameMidnight.Pillars)
	}
}

func loadExternalSilverFixture(t *testing.T) externalSilverFixture {
	t.Helper()
	raw, err := os.ReadFile("../testdata/bazi_external_silver.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture externalSilverFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func assertExternalSilverMetadata(t *testing.T, fixture externalSilverFixture) {
	t.Helper()
	if fixture.Version != "1.2" || fixture.Metadata.Tier != "silver" ||
		fixture.Metadata.Purpose != "bazi_external_structural_differential" ||
		fixture.Metadata.ReviewStatus != "cross_checked_not_gold" || fixture.Metadata.PublishableAccuracy ||
		fixture.Metadata.Generator != "scripts/generate-bazi-external-silver.sh" || len(fixture.Metadata.Limitations) != 5 {
		t.Fatalf("fixture can be mistaken for Gold: %+v", fixture.Metadata)
	}
	if len(fixture.Cases) != 60 || fixture.Metadata.DualConsensusCases != 21 ||
		fixture.Metadata.BoundaryCases != 30 || fixture.Metadata.BoundaryGroups != 15 ||
		fixture.Metadata.ZiPolicyCases != 8 || fixture.Metadata.DisputedCases != 1 {
		t.Fatalf("unexpected fixture coverage: cases=%d metadata=%+v", len(fixture.Cases), fixture.Metadata)
	}
	wantSources := make(map[string]RuleSourceMeta)
	for _, source := range baziExternalSilverSources() {
		wantSources[source.ID] = source
	}
	if len(fixture.Sources) != len(wantSources) {
		t.Fatalf("sources = %+v", fixture.Sources)
	}
	for _, source := range fixture.Sources {
		want, ok := wantSources[source.ID]
		if !ok || source.Repository != want.Repository || source.Commit != want.Commit ||
			source.License != want.License || !reflect.DeepEqual(source.Files, want.Files) {
			t.Fatalf("unpinned external source: %+v", source)
		}
	}
}

func pillarName(stem, branch string) string {
	return stem + branch
}

func normalizeExternalNaYin(value externalSilverPillars) externalSilverPillars {
	normalize := func(name string) string {
		switch name {
		case "沙中金":
			return "砂中金"
		case "沙中土":
			return "砂中土"
		case "泉中水":
			return "井泉水"
		default:
			return name
		}
	}
	return externalSilverPillars{
		Year: normalize(value.Year), Month: normalize(value.Month),
		Day: normalize(value.Day), Hour: normalize(value.Hour),
	}
}
