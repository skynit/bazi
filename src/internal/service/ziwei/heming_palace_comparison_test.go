package ziwei

import (
	"reflect"
	"testing"
)

func TestAnalyzeHeming_PalaceComparisonsProjectPublishedStructure(t *testing.T) {
	chartA, chartB := calculateKnowledgeFixtures(t)
	result := analyzeHeming(chartA, chartB)
	if result == nil {
		t.Fatal("valid charts must produce a structural comparison")
	}
	if result.EvidenceBasis != "deterministic_published_chart_projection" ||
		result.ValidationStatus != "not_adjudicated" || result.IsOutcomeConclusion {
		t.Fatalf("invalid heming boundary metadata: %+v", result)
	}

	wantPalaces := []string{"命宫", "夫妻", "福德", "事业", "财帛"}
	if len(result.PalaceComparisons) != len(wantPalaces) {
		t.Fatalf("palace comparisons = %d, want %d", len(result.PalaceComparisons), len(wantPalaces))
	}
	for i, wantName := range wantPalaces {
		comparison := result.PalaceComparisons[i]
		if comparison.Palace != wantName ||
			comparison.ComparisonBasis != "same_named_palace_exact_public_projection" ||
			comparison.InterpretationStatus != "not_adjudicated" || comparison.IsCompatibilityResult {
			t.Fatalf("invalid comparison metadata at %d: %+v", i, comparison)
		}
		palaceA := mustPublishedPalace(t, chartA, wantName)
		palaceB := mustPublishedPalace(t, chartB, wantName)
		if want := expectedHemingProjection(palaceA); !reflect.DeepEqual(comparison.ChartA, want) {
			t.Fatalf("chart A %s projection differs from public palace:\ngot=%+v\nwant=%+v", wantName, comparison.ChartA, want)
		}
		if want := expectedHemingProjection(palaceB); !reflect.DeepEqual(comparison.ChartB, want) {
			t.Fatalf("chart B %s projection differs from public palace:\ngot=%+v\nwant=%+v", wantName, comparison.ChartB, want)
		}
		if want := testOrderedIntersection(testStarNames(palaceA.Stars), testStarNames(palaceB.Stars)); !reflect.DeepEqual(comparison.SharedStars, want) {
			t.Fatalf("%s shared stars = %v, want %v", wantName, comparison.SharedStars, want)
		}
		if want := testOrderedIntersection(palaceA.FourHua, palaceB.FourHua); !reflect.DeepEqual(comparison.SharedFourHua, want) {
			t.Fatalf("%s shared four-hua = %v, want %v", wantName, comparison.SharedFourHua, want)
		}
		if want := testOrderedIntersection(palaceA.AdjectiveStars, palaceB.AdjectiveStars); !reflect.DeepEqual(comparison.SharedAdjectiveStars, want) {
			t.Fatalf("%s shared adjective stars = %v, want %v", wantName, comparison.SharedAdjectiveStars, want)
		}
	}
}

func TestBuildHemingPalaceComparison_UsesOnlyPublicFields(t *testing.T) {
	chartA := &ZiWeiChart{}
	chartB := &ZiWeiChart{}
	chartA.Palaces[0] = PalaceInfo{
		Name:    "命宫",
		Stars:   []StarOutput{},
		FourHua: []string{},
	}
	chartB.Palaces[0] = PalaceInfo{
		Name:    "命宫",
		Stars:   []StarOutput{},
		FourHua: []string{},
	}
	comparison, ok := buildHemingPalaceComparison(chartA, chartB, "命宫")
	if !ok {
		t.Fatal("public palace comparison was rejected")
	}
	if len(comparison.ChartA.Stars) != 0 || len(comparison.ChartB.Stars) != 0 || len(comparison.SharedStars) != 0 {
		t.Fatalf("empty public stars produced a non-empty comparison: %+v", comparison)
	}

	chartA.Palaces[0].Stars = []StarOutput{{Name: "A"}, {Name: "B"}, {Name: "A"}}
	chartB.Palaces[0].Stars = []StarOutput{{Name: "B"}, {Name: "A"}}
	chartA.Palaces[0].FourHua = []string{"甲化禄", "乙化权", "甲化禄"}
	chartB.Palaces[0].FourHua = []string{"乙化权", "甲化禄"}
	comparison, ok = buildHemingPalaceComparison(chartA, chartB, "命宫")
	if !ok {
		t.Fatal("public palace comparison was rejected")
	}
	if !reflect.DeepEqual(comparison.SharedStars, []string{"A", "B"}) ||
		!reflect.DeepEqual(comparison.SharedFourHua, []string{"甲化禄", "乙化权"}) {
		t.Fatalf("intersections are not stable A-order sets: %+v", comparison)
	}
}

func TestAnalyzeHeming_RejectsProfileAndPalaceLayoutMismatch(t *testing.T) {
	chartA, chartB := calculateKnowledgeFixtures(t)

	wrongVersion := *chartB
	wrongVersion.EngineVersion = "different-engine"
	restampZiWeiChartContentHash(t, &wrongVersion)
	if got := analyzeHeming(chartA, &wrongVersion); got != nil {
		t.Fatalf("cross-profile chart must be rejected: %+v", got)
	}

	duplicateName := *chartB
	duplicateName.Palaces[1].Name = duplicateName.Palaces[0].Name
	restampZiWeiChartContentHash(t, &duplicateName)
	if got := analyzeHeming(chartA, &duplicateName); got != nil {
		t.Fatalf("duplicate palace name must be rejected: %+v", got)
	}

	duplicateBranch := *chartB
	duplicateBranch.Palaces[1].Branch = duplicateBranch.Palaces[0].Branch
	restampZiWeiChartContentHash(t, &duplicateBranch)
	if got := analyzeHeming(chartA, &duplicateBranch); got != nil {
		t.Fatalf("duplicate palace branch must be rejected: %+v", got)
	}

	missingBodyPalace := *chartB
	for i := range missingBodyPalace.Palaces {
		missingBodyPalace.Palaces[i].IsBodyPalace = false
	}
	restampZiWeiChartContentHash(t, &missingBodyPalace)
	if got := analyzeHeming(chartA, &missingBodyPalace); got != nil {
		t.Fatalf("chart without one body-palace marker must be rejected: %+v", got)
	}

	inconsistentBodyBranch := *chartB
	inconsistentBodyBranch.BodyPalace = "子"
	if inconsistentBodyBranch.BodyPalace == chartB.BodyPalace {
		inconsistentBodyBranch.BodyPalace = "丑"
	}
	restampZiWeiChartContentHash(t, &inconsistentBodyBranch)
	if got := analyzeHeming(chartA, &inconsistentBodyBranch); got != nil {
		t.Fatalf("inconsistent body-palace branch must be rejected: %+v", got)
	}
}

func mustPublishedPalace(t *testing.T, chart *ZiWeiChart, name string) PalaceInfo {
	t.Helper()
	palace, ok := uniquePublishedPalace(chart, name)
	if !ok {
		t.Fatalf("missing unique palace %q", name)
	}
	return palace
}

func expectedHemingProjection(palace PalaceInfo) HemingPalaceProjection {
	var stars []StarOutput
	if palace.Stars != nil {
		stars = append([]StarOutput{}, palace.Stars...)
	}
	var fourHua, adjectiveStars []string
	if palace.FourHua != nil {
		fourHua = append([]string{}, palace.FourHua...)
	}
	if palace.AdjectiveStars != nil {
		adjectiveStars = append([]string{}, palace.AdjectiveStars...)
	}
	return HemingPalaceProjection{
		Branch: palace.Branch, HeavenlyStem: palace.HeavenlyStem, IsBodyPalace: palace.IsBodyPalace,
		Stars: stars, FourHua: fourHua, AdjectiveStars: adjectiveStars,
		Changsheng12: palace.Changsheng12, Boshi12: palace.Boshi12,
		JiangQian12: palace.JiangQian12, SuiQian12: palace.SuiQian12,
	}
}

func testStarNames(stars []StarOutput) []string {
	result := make([]string, 0, len(stars))
	for _, star := range stars {
		result = append(result, star.Name)
	}
	return result
}

func testOrderedIntersection(a, b []string) []string {
	present := make(map[string]bool, len(b))
	for _, value := range b {
		present[value] = true
	}
	result := make([]string, 0)
	seen := make(map[string]bool)
	for _, value := range a {
		if present[value] && !seen[value] {
			result = append(result, value)
			seen[value] = true
		}
	}
	return result
}
