package ziwei

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/6tail/tyme4go/tyme"
)

type iztroSilverFixture struct {
	Version     string                  `json:"version"`
	ProfileID   string                  `json:"profile_id"`
	Metadata    iztroSilverMetadata     `json:"metadata"`
	Sources     []iztroSilverSource     `json:"sources"`
	Cases       []iztroSilverCase       `json:"cases"`
	PeriodCases []iztroPeriodSilverCase `json:"period_cases"`
}

type iztroSilverMetadata struct {
	Tier                string `json:"tier"`
	Purpose             string `json:"purpose"`
	ReviewStatus        string `json:"review_status"`
	PublishableAccuracy bool   `json:"publishable_accuracy"`
	Generator           string `json:"generator"`
}

type iztroSilverSource struct {
	RuleID     string `json:"rule_id"`
	Repository string `json:"repository"`
	Commit     string `json:"commit"`
	Path       string `json:"path"`
	SHA256     string `json:"sha256"`
	License    string `json:"license"`
}

type iztroSilverCase struct {
	ID       string              `json:"id"`
	Input    iztroSilverInput    `json:"input"`
	Expected iztroSilverExpected `json:"expected"`
}

type iztroSilverInput struct {
	ID     string `json:"id"`
	Date   string `json:"date"`
	Year   int    `json:"year"`
	Month  int    `json:"month"`
	Day    int    `json:"day"`
	Hour   int    `json:"hour"`
	Gender string `json:"gender"`
}

type iztroSilverExpected struct {
	SoulPalace string                       `json:"soul_palace"`
	BodyPalace string                       `json:"body_palace"`
	LifeMaster string                       `json:"life_master"`
	BodyMaster string                       `json:"body_master"`
	FiveBureau string                       `json:"five_bureau"`
	Palaces    map[string]iztroSilverPalace `json:"palaces"`
}

type iztroSilverPalace struct {
	Name           string            `json:"name"`
	HeavenlyStem   string            `json:"heavenly_stem"`
	MainStars      []iztroSilverStar `json:"main_stars"`
	AuxStars       []iztroSilverStar `json:"aux_stars"`
	AdjectiveStars []string          `json:"adjective_stars"`
	Changsheng12   string            `json:"changsheng_12"`
	Boshi12        string            `json:"boshi_12"`
	JiangQian12    string            `json:"jiang_qian_12"`
	SuiQian12      string            `json:"sui_qian_12"`
	Dayun          iztroSilverDayun  `json:"dayun"`
}

type iztroSilverStar struct {
	Name       string `json:"name"`
	Brightness string `json:"brightness"`
	Mutagen    string `json:"mutagen"`
}

type iztroSilverDayun struct {
	StartAge int `json:"start_age"`
	EndAge   int `json:"end_age"`
}

type iztroPeriodSilverCase struct {
	ID       string                    `json:"id"`
	Input    iztroPeriodSilverInput    `json:"input"`
	Expected iztroPeriodSilverExpected `json:"expected"`
}

type iztroPeriodSilverInput struct {
	ID        string `json:"id"`
	BirthDate string `json:"birth_date"`
	BirthHour int    `json:"birth_hour"`
	Gender    string `json:"gender"`
	Target    string `json:"target"`
}

type iztroPeriodGanZhi struct {
	Index               int                 `json:"index"`
	HeavenlyStem        string              `json:"heavenly_stem"`
	EarthlyBranch       string              `json:"earthly_branch"`
	Mutagen             []string            `json:"mutagen"`
	PalaceNamesByBranch map[string]string   `json:"palace_names_by_branch"`
	StarsByBranch       map[string][]string `json:"stars_by_branch"`
}

type iztroPeriodSilverExpected struct {
	SolarDate string            `json:"solar_date"`
	LunarDate string            `json:"lunar_date"`
	Yearly    iztroPeriodGanZhi `json:"yearly"`
	Monthly   iztroPeriodGanZhi `json:"monthly"`
	Daily     iztroPeriodGanZhi `json:"daily"`
}

func TestIztroSilverFullChartDifferential(t *testing.T) {
	fixture := loadIztroSilverFixture(t)
	assertIztroSilverMetadata(t, fixture)

	svc := NewZiWeiService()
	seenIDs := make(map[string]bool, len(fixture.Cases))
	coveredStems := make(map[int]bool)
	coveredLunarMonths := make(map[int]bool)
	coveredLunarDays := make(map[int]bool)
	coveredYearBranches := make(map[int]bool)
	coveredHourBranches := make(map[int]bool)
	coveredSoulBranches := make(map[string]bool)
	coveredBodyBranches := make(map[string]bool)
	coveredBureaus := make(map[string]bool)
	coveredGenders := make(map[string]bool)
	leapBoundaryDays := map[string]struct {
		solarDay int
		lunarDay int
	}{
		"leap_day_15": {solarDay: 6, lunarDay: 15},
		"leap_day_16": {solarDay: 7, lunarDay: 16},
	}
	coveredLeapBoundary := make(map[string]bool, len(leapBoundaryDays))
	for index, tc := range fixture.Cases {
		if tc.ID == "" || seenIDs[tc.ID] || tc.Input.ID != tc.ID {
			t.Fatalf("case %d has invalid or duplicate id: case=%q input=%q", index, tc.ID, tc.Input.ID)
		}
		if tc.Input.Date != fmt.Sprintf("%d-%d-%d", tc.Input.Year, tc.Input.Month, tc.Input.Day) {
			t.Fatalf("case %q date and numeric fields diverge: %+v", tc.ID, tc.Input)
		}
		seenIDs[tc.ID] = true
		t.Run(tc.ID, func(t *testing.T) {
			assertIztroSilverAdjectiveCoverage(t, tc.Expected)
			chart, err := svc.CalculateChart(tc.Input.Year, tc.Input.Month, tc.Input.Day, tc.Input.Hour, 0, tc.Input.Gender)
			if err != nil {
				t.Fatal(err)
			}
			birth := mustPublishedBirthData(t, chart)
			if index < 10 {
				coveredStems[birth.YearStem] = true
			}
			coveredLunarMonths[birth.LunarMonth] = true
			coveredLunarDays[birth.LunarDay] = true
			coveredYearBranches[birth.YearBranch] = true
			coveredHourBranches[birth.HourBranch] = true
			coveredSoulBranches[chart.EarthlyBranchOfSoulPalace] = true
			coveredBodyBranches[chart.EarthlyBranchOfBodyPalace] = true
			coveredBureaus[chart.FiveBureau] = true
			coveredGenders[tc.Input.Gender] = true
			if want, ok := leapBoundaryDays[tc.ID]; ok {
				if tc.Input.Year != 2020 || tc.Input.Month != 6 || tc.Input.Day != want.solarDay ||
					birth.LunarMonth != 4 || birth.LunarDay != want.lunarDay || !birth.IsLeapMonth {
					t.Fatalf("leap-month boundary input/result mismatch: input=%+v birth=%+v", tc.Input, birth)
				}
				coveredLeapBoundary[tc.ID] = true
			}
			got := projectChartForIztroSilver(chart, svc.CalculateDayun(chart))
			if !reflect.DeepEqual(got, tc.Expected) {
				t.Fatalf("fixture mismatch:\n%s", strings.Join(explainIztroSilverDiff(got, tc.Expected), "\n"))
			}
		})
	}
	if len(fixture.Cases) != 37 {
		t.Fatalf("silver cases = %d, want 37", len(fixture.Cases))
	}
	if len(coveredStems) != 10 {
		t.Fatalf("first ten cases cover %d heavenly stems, want 10: %#v", len(coveredStems), coveredStems)
	}
	if len(coveredLeapBoundary) != len(leapBoundaryDays) {
		t.Fatalf("full-chart Silver is missing leap-month day 15/16 boundary cases: %+v", coveredLeapBoundary)
	}
	if len(coveredLunarMonths) != 12 || !coveredLunarDays[1] || !coveredLunarDays[30] {
		t.Fatalf("full-chart Silver lunar coverage is incomplete: months=%+v days=%+v", coveredLunarMonths, coveredLunarDays)
	}
	if len(coveredYearBranches) != 12 || len(coveredHourBranches) != 12 {
		t.Fatalf("full-chart Silver branch coverage is incomplete: years=%+v hours=%+v", coveredYearBranches, coveredHourBranches)
	}
	if len(coveredSoulBranches) != 12 || len(coveredBodyBranches) != 12 {
		t.Fatalf("full-chart Silver palace coverage is incomplete: soul=%+v body=%+v", coveredSoulBranches, coveredBodyBranches)
	}
	if len(coveredBureaus) != 5 || len(coveredGenders) != 2 {
		t.Fatalf("full-chart Silver bureau/gender coverage is incomplete: bureaus=%+v genders=%+v", coveredBureaus, coveredGenders)
	}
}

func TestIztroSilverPeriodChronologyDifferential(t *testing.T) {
	fixture := loadIztroSilverFixture(t)
	assertIztroSilverMetadata(t, fixture)
	if len(fixture.PeriodCases) != 18 {
		t.Fatalf("period silver cases = %d, want 18", len(fixture.PeriodCases))
	}

	svc := NewZiWeiService()
	seen := map[string]bool{}
	coveredDailyStems := map[string]bool{}
	for _, tc := range fixture.PeriodCases {
		if tc.ID == "" || seen[tc.ID] || tc.Input.ID != tc.ID {
			t.Fatalf("invalid or duplicate period case: case=%q input=%q", tc.ID, tc.Input.ID)
		}
		seen[tc.ID] = true
		coveredDailyStems[tc.Expected.Daily.HeavenlyStem] = true
		t.Run(tc.ID, func(t *testing.T) {
			var birthYear, birthMonth, birthDay int
			if _, err := fmt.Sscanf(tc.Input.BirthDate, "%d-%d-%d", &birthYear, &birthMonth, &birthDay); err != nil {
				t.Fatal(err)
			}
			var year, month, day int
			if _, err := fmt.Sscanf(tc.Input.Target, "%d-%d-%d", &year, &month, &day); err != nil {
				t.Fatal(err)
			}

			base, err := svc.CalculateChart(birthYear, birthMonth, birthDay, tc.Input.BirthHour, 0, tc.Input.Gender)
			if err != nil {
				t.Fatal(err)
			}
			liuyue := svc.CalculateLiuyueForDate(base, year, month, day)
			liuri := svc.CalculateLiuriForDate(base, year, month, day)
			if liuyue == nil || liuri == nil || liuyue.DerivationInput == nil || liuri.DerivationInput == nil {
				t.Fatal("valid Silver target did not produce complete period inputs")
			}

			wantMonthly := tc.Expected.Monthly.HeavenlyStem + tc.Expected.Monthly.EarthlyBranch
			wantDaily := tc.Expected.Daily.HeavenlyStem + tc.Expected.Daily.EarthlyBranch
			if liuyue.DerivationInput.PeriodGanZhi != wantMonthly {
				t.Fatalf("monthly gan-zhi = %q, want iztro %q", liuyue.DerivationInput.PeriodGanZhi, wantMonthly)
			}
			if liuri.DerivationInput.PeriodGanZhi != wantDaily {
				t.Fatalf("daily gan-zhi = %q, want iztro %q", liuri.DerivationInput.PeriodGanZhi, wantDaily)
			}
			lunarYear, err := tyme.LunarYear{}.FromYear(liuyue.DerivationInput.ResolvedLunarDate.Year)
			if err != nil {
				t.Fatal(err)
			}
			wantYearly := tc.Expected.Yearly.HeavenlyStem + tc.Expected.Yearly.EarthlyBranch
			if lunarYear.GetSixtyCycle().GetName() != wantYearly {
				t.Fatalf("yearly gan-zhi = %q, want iztro %q", lunarYear.GetSixtyCycle().GetName(), wantYearly)
			}
			liunian := svc.CalculateLiunian(base, liuyue.DerivationInput.ResolvedLunarDate.Year)
			if liunian == nil {
				t.Fatal("valid Silver lunar year did not produce a liunian chart")
			}
			assertIztroTransitLayer(t, liunian, liunian.LiuNianPalaces, liunian.LiuNianStars, liunian.LiuNianFourHua, tc.Expected.Yearly)
			assertIztroTransitLayer(t, liuyue, liuyue.LiuYuePalaces, liuyue.LiuYueStars, liuyue.LiuYueFourHua, tc.Expected.Monthly)
			assertIztroTransitLayer(t, liuri, liuri.LiuRiPalaces, liuri.LiuRiStars, liuri.LiuRiFourHua, tc.Expected.Daily)
			if strings.Contains(tc.Expected.LunarDate, "闰") != liuyue.DerivationInput.ResolvedLunarDate.IsLeapMonth {
				t.Fatalf("leap-month flag does not match iztro lunar date %q: %+v", tc.Expected.LunarDate, liuyue.DerivationInput.ResolvedLunarDate)
			}
			if !ValidDerivedChartContract(liuyue) || !ValidDerivedChartContract(liuri) {
				t.Fatal("Silver period result failed its derivation integrity contract")
			}
		})
	}
	if len(coveredDailyStems) != 10 {
		t.Fatalf("period Silver cases cover %d daily stems, want 10: %+v", len(coveredDailyStems), coveredDailyStems)
	}
}

func assertIztroTransitLayer(t *testing.T, chart *ZiWeiChart, palaces [12]string, stars, fourHua [12][]string, want iztroPeriodGanZhi) {
	t.Helper()
	gotByBranch := make(map[string][]string, len(chart.Palaces))
	for i, palace := range chart.Palaces {
		gotByBranch[palace.Branch] = append([]string{}, stars[i]...)
	}
	if !reflect.DeepEqual(gotByBranch, want.StarsByBranch) {
		gotJSON, _ := json.Marshal(gotByBranch)
		wantJSON, _ := json.Marshal(want.StarsByBranch)
		t.Fatalf("transit stars mismatch: got=%s want=%s", gotJSON, wantJSON)
	}
	gotPalaces := make(map[string]string, len(chart.Palaces))
	gotIndex := -1
	for i, palace := range chart.Palaces {
		gotPalaces[palace.Branch] = palaces[i]
		if palaces[i] == "命宫" {
			branchIndex := BranchIndex[palace.Branch]
			gotIndex = fixIndex(branchIndex - 2)
		}
	}
	if gotIndex != want.Index || !reflect.DeepEqual(gotPalaces, want.PalaceNamesByBranch) {
		gotJSON, _ := json.Marshal(gotPalaces)
		wantJSON, _ := json.Marshal(want.PalaceNamesByBranch)
		t.Fatalf("transit palaces mismatch: index=%d/%d got=%s want=%s", gotIndex, want.Index, gotJSON, wantJSON)
	}

	wantFourHua := make([]string, 0, len(want.Mutagen))
	for i, star := range want.Mutagen {
		wantFourHua = append(wantFourHua, star+SiHuaLabels[i])
	}
	gotFourHua := flattenPeriodStars(fourHua)
	sort.Strings(gotFourHua)
	sort.Strings(wantFourHua)
	if !reflect.DeepEqual(gotFourHua, wantFourHua) {
		t.Fatalf("transit four-hua = %v, want iztro %v", gotFourHua, wantFourHua)
	}
}

func loadIztroSilverFixture(t *testing.T) iztroSilverFixture {
	t.Helper()
	raw, err := os.ReadFile("../testdata/ziwei_iztro_silver.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixture iztroSilverFixture
	if err := json.Unmarshal(raw, &fixture); err != nil {
		t.Fatal(err)
	}
	return fixture
}

func assertIztroSilverMetadata(t *testing.T, fixture iztroSilverFixture) {
	t.Helper()
	if fixture.Version != "1.1" || fixture.ProfileID != DefaultProfileID {
		t.Fatalf("fixture version/profile = %q/%q", fixture.Version, fixture.ProfileID)
	}
	if fixture.Metadata.Tier != "silver" || fixture.Metadata.Purpose != "external_chart_and_period_differential" ||
		fixture.Metadata.ReviewStatus != "cross_checked_not_gold" ||
		fixture.Metadata.PublishableAccuracy || fixture.Metadata.Generator != "scripts/generate-ziwei-iztro-silver.sh" {
		t.Fatalf("fixture metadata can be mistaken for Gold: %+v", fixture.Metadata)
	}
	wantSources := map[string]RuleSourceRef{
		SiHuaRuleID: {
			RuleID: SiHuaRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: SiHuaSourcePath, SHA256: SiHuaSourceSHA256, License: "MIT",
		},
		StarBrightnessRuleID: {
			RuleID: StarBrightnessRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: StarBrightnessSourcePath, SHA256: StarBrightnessSHA256, License: "MIT",
		},
		LeapMonthRuleID: {
			RuleID: LeapMonthRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: LeapMonthSourcePath, SHA256: LeapMonthSHA256, License: "MIT",
		},
		MonthlyStarsRuleID: {
			RuleID: MonthlyStarsRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: MonthlyStarsSourcePath, SHA256: MonthlyStarsSHA256, License: "MIT",
		},
		AdjectiveStarsRuleID: {
			RuleID: AdjectiveStarsRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: AdjectiveStarsSourcePath, SHA256: AdjectiveStarsSHA256, License: "MIT",
		},
		PeriodChronologyRuleID: {
			RuleID: PeriodChronologyRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: PeriodChronologySourcePath, SHA256: PeriodChronologySHA256, License: "MIT",
		},
		TransitStarsRuleID: {
			RuleID: TransitStarsRuleID, Repository: SiHuaSourceRepo, Commit: SiHuaSourceCommit,
			Path: TransitStarsSourcePath, SHA256: TransitStarsSHA256, License: "MIT",
		},
	}
	if len(fixture.Sources) != len(wantSources) {
		t.Fatalf("fixture sources = %+v", fixture.Sources)
	}
	for _, source := range fixture.Sources {
		want, ok := wantSources[source.RuleID]
		if !ok || source.Repository != want.Repository || source.Commit != want.Commit ||
			source.Path != want.Path || source.SHA256 != want.SHA256 || source.License != want.License {
			t.Fatalf("fixture source is not pinned to profile: %+v", source)
		}
	}
}

func assertIztroSilverAdjectiveCoverage(t testing.TB, expected iztroSilverExpected) {
	t.Helper()
	seen := make(map[string]bool, 38)
	count := 0
	for _, palace := range expected.Palaces {
		for _, star := range palace.AdjectiveStars {
			if star == "" || seen[star] {
				t.Fatalf("Silver adjective stars contain an empty or duplicate name %q", star)
			}
			seen[star] = true
			count++
		}
	}
	if count != 38 {
		t.Fatalf("Silver adjective-star coverage = %d unique stars, want 38", count)
	}
}

func projectChartForIztroSilver(chart *ZiWeiChart, dayun Dayun) iztroSilverExpected {
	result := iztroSilverExpected{
		SoulPalace: chart.EarthlyBranchOfSoulPalace,
		BodyPalace: chart.EarthlyBranchOfBodyPalace,
		LifeMaster: chart.LifeMaster,
		BodyMaster: chart.BodyMaster,
		FiveBureau: chart.FiveBureau,
		Palaces:    make(map[string]iztroSilverPalace, len(chart.Palaces)),
	}
	ranges := make(map[string]iztroSilverDayun, len(dayun))
	for _, stage := range dayun {
		ranges[stage.Palace] = iztroSilverDayun{StartAge: stage.StartAge, EndAge: stage.EndAge}
	}
	for _, palace := range chart.Palaces {
		item := iztroSilverPalace{
			Name: palace.Name, HeavenlyStem: palace.HeavenlyStem,
			Changsheng12: palace.Changsheng12, Boshi12: palace.Boshi12,
			JiangQian12: palace.JiangQian12, SuiQian12: palace.SuiQian12,
			Dayun: ranges[palace.Name], MainStars: []iztroSilverStar{}, AuxStars: []iztroSilverStar{},
			AdjectiveStars: append([]string{}, palace.AdjectiveStars...),
		}
		for _, star := range palace.Stars {
			mutagen := ""
			for _, hua := range palace.FourHua {
				if strings.HasPrefix(hua, star.Name+"化") {
					mutagen = strings.TrimPrefix(hua, star.Name+"化")
					break
				}
			}
			value := iztroSilverStar{Name: star.Name, Brightness: star.Brightness, Mutagen: mutagen}
			if star.Type == "major" {
				item.MainStars = append(item.MainStars, value)
			} else {
				item.AuxStars = append(item.AuxStars, value)
			}
		}
		sort.Slice(item.MainStars, func(i, j int) bool { return item.MainStars[i].Name < item.MainStars[j].Name })
		sort.Slice(item.AuxStars, func(i, j int) bool { return item.AuxStars[i].Name < item.AuxStars[j].Name })
		sort.Strings(item.AdjectiveStars)
		result.Palaces[palace.Branch] = item
	}
	return result
}

func explainIztroSilverDiff(got, want iztroSilverExpected) []string {
	var diffs []string
	for _, item := range []struct{ label, got, want string }{
		{"soul_palace", got.SoulPalace, want.SoulPalace},
		{"body_palace", got.BodyPalace, want.BodyPalace},
		{"life_master", got.LifeMaster, want.LifeMaster},
		{"body_master", got.BodyMaster, want.BodyMaster},
		{"five_bureau", got.FiveBureau, want.FiveBureau},
	} {
		if item.got != item.want {
			diffs = append(diffs, fmt.Sprintf("%s got=%q want=%q", item.label, item.got, item.want))
		}
	}
	branches := append([]string(nil), BranchNames...)
	for _, branch := range branches {
		if !reflect.DeepEqual(got.Palaces[branch], want.Palaces[branch]) {
			gotJSON, _ := json.Marshal(got.Palaces[branch])
			wantJSON, _ := json.Marshal(want.Palaces[branch])
			diffs = append(diffs, fmt.Sprintf("palace %s got=%s want=%s", branch, gotJSON, wantJSON))
		}
	}
	if len(diffs) == 0 {
		diffs = append(diffs, "unknown structural mismatch")
	}
	return diffs
}
