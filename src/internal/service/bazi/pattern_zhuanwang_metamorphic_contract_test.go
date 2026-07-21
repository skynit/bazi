package bazi

import (
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

func zhuanWangPillarsForBranches(t *testing.T, element string, branches []string) []model.Pillar {
	t.Helper()
	stemElements := patternStemElementProfile()
	pillars := make([]model.Pillar, 0, len(branches))
	for _, branch := range branches {
		stem := ""
		for _, target := range stemElements {
			if target.Element != element {
				continue
			}
			if _, err := (tyme.SixtyCycle{}).FromName(target.Symbol + branch); err == nil {
				stem = target.Symbol
				break
			}
		}
		if stem == "" {
			t.Fatalf("no valid %s stem for branch %s", element, branch)
		}
		pillars = append(pillars, model.Pillar{Gan: stem, Zhi: branch})
	}
	return pillars
}

func assertZhuanWangMetamorphicResult(t *testing.T, entry zhuanWangSemanticEntry, branches []string, wantMatch bool) {
	t.Helper()
	pillars := zhuanWangPillarsForBranches(t, entry.Element, branches)
	result := checkZhuanWangGe(pillars)
	if wantMatch {
		if result == nil || result.PatternName != entry.PatternName {
			t.Errorf("%s branches %v returned %+v", entry.PatternName, branches, result)
		}
		return
	}
	if result != nil {
		t.Errorf("%s rejecting branches %v returned %+v", entry.PatternName, branches, result)
	}
}

func TestZhuanWangBranchMultisetMetamorphicRelations(t *testing.T) {
	profile := zhuanWangDetectorSemanticProfile()
	structureCount, preserveCount, rejectCount := 0, 0, 0
	for _, entry := range profile.Entries {
		for _, structure := range entry.Structures {
			structureCount++
			if len(structure) == 3 {
				for _, repeated := range structure {
					branches := append(append([]string(nil), structure...), repeated)
					for _, permutation := range uniqueBranchPermutations(branches) {
						assertZhuanWangMetamorphicResult(t, entry, permutation, true)
						preserveCount++
					}
				}
				for _, target := range profile.BranchElements {
					if patternStringProfileContains(structure, target.Symbol) {
						continue
					}
					branches := append(append([]string(nil), structure...), target.Symbol)
					wantMatch := target.Element != entry.BreakingElement
					for _, permutation := range uniqueBranchPermutations(branches) {
						assertZhuanWangMetamorphicResult(t, entry, permutation, wantMatch)
						if wantMatch {
							preserveCount++
						} else {
							rejectCount++
						}
					}
				}
			} else {
				for _, permutation := range uniqueBranchPermutations(append([]string(nil), structure...)) {
					assertZhuanWangMetamorphicResult(t, entry, permutation, true)
					preserveCount++
				}
			}

			for omitted := range structure {
				branches := make([]string, 0, 4)
				for index, branch := range structure {
					if index != omitted {
						branches = append(branches, branch)
					}
				}
				for len(branches) < 4 {
					branches = append(branches, branches[0])
				}
				for _, permutation := range uniqueBranchPermutations(branches) {
					assertZhuanWangMetamorphicResult(t, entry, permutation, false)
					rejectCount++
				}
			}
		}
	}
	if structureCount != 9 || preserveCount != 1632 || rejectCount != 552 {
		t.Fatalf("zhuan-wang metamorphic cases = %d structures/%d preserve/%d reject", structureCount, preserveCount, rejectCount)
	}
}

func TestZhuanWangMetamorphicProfileIsIndependentAndBound(t *testing.T) {
	want := []patternDetectorMetamorphicPolicy{
		{ID: "zhuanwang.meta.branch_permutation", Transform: "permute_four_branch_positions_keep_day_stem_element", Relation: "preserve_output"},
		{ID: "zhuanwang.meta.repeat_required_branch", Transform: "repeat_any_required_branch_after_all_required_present", Relation: "preserve_output"},
		{ID: "zhuanwang.meta.external_nonbreaking_branch", Transform: "append_branch_outside_structure_with_nonbreaking_main_element", Relation: "preserve_output"},
		{ID: "zhuanwang.meta.external_breaking_branch", Transform: "append_branch_outside_structure_with_breaking_main_element", Relation: "reject"},
		{ID: "zhuanwang.meta.missing_required_branch", Transform: "remove_one_required_branch_and_duplicate_remaining_branches", Relation: "reject"},
	}
	got := patternDetectorMetamorphicProfile("pattern.special.zhuanwang")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("zhuan-wang metamorphic Profile = %+v, want %+v", got, want)
	}
	got[0].ID = "mutated"
	if fresh := patternDetectorMetamorphicProfile("pattern.special.zhuanwang"); !reflect.DeepEqual(fresh, want) {
		t.Fatalf("fresh metamorphic Profile inherited mutation: %+v", fresh)
	}
	profile, ok := patternDetectorSemanticProfile("pattern.special.zhuanwang")
	envelope, envelopeOK := profile.(patternDetectorSemanticEnvelope)
	if !ok || !envelopeOK || !reflect.DeepEqual(envelope.MetamorphicPolicies, want) {
		t.Fatalf("zhuan-wang semantic envelope omits metamorphic policies: %#v", profile)
	}
	for _, policy := range patternDetectorMetamorphicProfile("pattern.special.liangqi") {
		if strings.HasPrefix(policy.ID, "zhuanwang.meta.") {
			t.Fatalf("liang-qi inherited zhuan-wang metamorphic policy: %+v", policy)
		}
	}
}

func TestZhuanWangMetamorphicMetadataContract(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"专旺检测仍缺少地支多重集与柱位排列的形变合同",
			"metamorphic_policies",
			"9组方局、三合局与四库结构",
			"1632个保持命中与552个拒绝盘",
			"重复必需支、缺失必需支和全部结构外地支",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
