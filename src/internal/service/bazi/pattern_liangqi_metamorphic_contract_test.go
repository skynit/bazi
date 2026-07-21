package bazi

import (
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

func fourOfEightCombinations() [][]int {
	result := make([][]int, 0, 70)
	for first := 0; first < 5; first++ {
		for second := first + 1; second < 6; second++ {
			for third := second + 1; third < 7; third++ {
				for fourth := third + 1; fourth < 8; fourth++ {
					result = append(result, []int{first, second, third, fourth})
				}
			}
		}
	}
	return result
}

func liangQiCanonicalStem(element string) string {
	for _, target := range patternStemElementProfile() {
		if target.Element == element {
			return target.Symbol
		}
	}
	return ""
}

func liangQiCanonicalBranch(element string) string {
	for _, target := range patternBranchElementProfile() {
		if target.Element != element {
			continue
		}
		if _, err := (tyme.SixtyCycle{}).FromName(liangQiCanonicalStem(element) + target.Symbol); err == nil {
			return target.Symbol
		}
	}
	return ""
}

func liangQiPillarsForElementSlots(t *testing.T, slots []string) []model.Pillar {
	t.Helper()
	if len(slots) != 8 {
		t.Fatalf("liang-qi slot count = %d, want 8", len(slots))
	}
	pillars := make([]model.Pillar, 4)
	for index := range pillars {
		pillars[index] = model.Pillar{
			Gan: liangQiCanonicalStem(slots[index*2]),
			Zhi: liangQiCanonicalBranch(slots[index*2+1]),
		}
		if _, err := (tyme.SixtyCycle{}).FromName(pillars[index].Gan + pillars[index].Zhi); err != nil {
			t.Fatalf("invalid generated pillar %s%s for slots %v: %v", pillars[index].Gan, pillars[index].Zhi, slots, err)
		}
	}
	return pillars
}

func assertLiangQiMetamorphicResult(t *testing.T, slots []string, wantMatch bool) {
	t.Helper()
	result := checkLiangQiChengXiang(liangQiPillarsForElementSlots(t, slots))
	if wantMatch {
		if result == nil || result.PatternName != "两气成象格" {
			t.Errorf("liang-qi slots %v returned %+v", slots, result)
		}
		return
	}
	if result != nil {
		t.Errorf("liang-qi rejecting slots %v returned %+v", slots, result)
	}
}

func TestLiangQiElementPairAndPositionMetamorphicRelations(t *testing.T) {
	elements := liangQiSemanticProfile().ElementOrder
	combinations := fourOfEightCombinations()
	pairCount, preserveCount, rejectCount := 0, 0, 0
	for first := 0; first < len(elements); first++ {
		for second := first + 1; second < len(elements); second++ {
			pairCount++
			third := ""
			for _, element := range elements {
				if element != elements[first] && element != elements[second] {
					third = element
					break
				}
			}
			for _, positions := range combinations {
				slots := make([]string, 8)
				for index := range slots {
					slots[index] = elements[second]
				}
				for _, position := range positions {
					slots[position] = elements[first]
				}
				assertLiangQiMetamorphicResult(t, slots, true)
				preserveCount++

				thirdElement := append([]string(nil), slots...)
				thirdElement[positions[0]] = third
				assertLiangQiMetamorphicResult(t, thirdElement, false)
				rejectCount++

				unequal := append([]string(nil), slots...)
				unequal[positions[0]] = elements[second]
				assertLiangQiMetamorphicResult(t, unequal, false)
				rejectCount++
			}
		}
	}
	if len(combinations) != 70 || pairCount != 10 || preserveCount != 700 || rejectCount != 1400 {
		t.Fatalf("liang-qi metamorphic cases = %d combinations/%d pairs/%d preserve/%d reject", len(combinations), pairCount, preserveCount, rejectCount)
	}
}

func TestLiangQiMetamorphicProfileIsIndependentAndBound(t *testing.T) {
	want := []patternDetectorMetamorphicPolicy{
		{ID: "liangqi.meta.unordered_element_pair", Transform: "choose_each_unordered_pair_from_five_elements", Relation: "preserve_output"},
		{ID: "liangqi.meta.four_of_eight_positions", Transform: "assign_first_element_to_any_four_of_eight_stem_branch_positions", Relation: "preserve_output"},
		{ID: "liangqi.meta.introduce_third_element", Transform: "replace_one_first_element_position_with_element_outside_pair", Relation: "reject"},
		{ID: "liangqi.meta.unequal_five_three_split", Transform: "replace_one_first_element_position_with_second_element", Relation: "reject"},
	}
	got := patternDetectorMetamorphicProfile("pattern.special.liangqi")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("liang-qi metamorphic Profile = %+v, want %+v", got, want)
	}
	got[0].ID = "mutated"
	if fresh := patternDetectorMetamorphicProfile("pattern.special.liangqi"); !reflect.DeepEqual(fresh, want) {
		t.Fatalf("fresh liang-qi metamorphic Profile inherited mutation: %+v", fresh)
	}
	profile, ok := patternDetectorSemanticProfile("pattern.special.liangqi")
	envelope, envelopeOK := profile.(patternDetectorSemanticEnvelope)
	if !ok || !envelopeOK || !reflect.DeepEqual(envelope.MetamorphicPolicies, want) {
		t.Fatalf("liang-qi semantic envelope omits metamorphic policies: %#v", profile)
	}
	for _, policy := range patternDetectorMetamorphicProfile("pattern.special.zhuanwang") {
		if strings.HasPrefix(policy.ID, "liangqi.meta.") {
			t.Fatalf("zhuan-wang inherited liang-qi metamorphic policy: %+v", policy)
		}
	}
}

func TestLiangQiMetamorphicMetadataContract(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"两气成象仍缺少十种无序五行对与八个位点的对称形变合同",
			"10种无序五行对",
			"每对70种四选四位置组合",
			"700个保持命中与1400个拒绝盘",
			"第三气与5:3非均分",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
