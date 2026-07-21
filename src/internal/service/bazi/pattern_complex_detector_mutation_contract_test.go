package bazi

import (
	"reflect"
	"strings"
	"testing"

	"bazi/internal/model"
	"github.com/6tail/tyme4go/tyme"
)

func patternWitnessPillars(witness patternDetectorBehaviorWitness) []model.Pillar {
	pillars := make([]model.Pillar, 0, len(witness.Pillars))
	for _, pillar := range witness.Pillars {
		pillars = append(pillars, model.Pillar{Gan: pillar.Stem, Zhi: pillar.Branch})
	}
	return pillars
}

func patternWitnessMutationDistance(baseline, mutant patternDetectorBehaviorWitness) int {
	distance := 0
	shared := len(baseline.Pillars)
	if len(mutant.Pillars) < shared {
		shared = len(mutant.Pillars)
	}
	distance += len(baseline.Pillars) - shared
	distance += len(mutant.Pillars) - shared
	for index := 0; index < shared; index++ {
		if baseline.Pillars[index].Stem != mutant.Pillars[index].Stem {
			distance++
		}
		if baseline.Pillars[index].Branch != mutant.Pillars[index].Branch {
			distance++
		}
	}
	return distance
}

func evaluateComplexPatternWitness(ruleID string, witness patternDetectorBehaviorWitness) string {
	pillars := patternWitnessPillars(witness)
	var result *patternDetection
	switch ruleID {
	case "pattern.special.zhuanwang":
		result = checkZhuanWangGe(pillars)
	case "pattern.special.liangqi":
		result = checkLiangQiChengXiang(pillars)
	}
	if result == nil {
		return ""
	}
	return result.PatternName
}

func TestComplexPatternBehaviorWitnessesKillSingleInputMutations(t *testing.T) {
	for _, ruleID := range []string{"pattern.special.zhuanwang", "pattern.special.liangqi"} {
		witnesses := patternDetectorBehaviorWitnessProfile(ruleID)
		byID := make(map[string]patternDetectorBehaviorWitness, len(witnesses))
		positiveCount, mutationCount := 0, 0
		for _, witness := range witnesses {
			if witness.ID == "" {
				t.Errorf("%s has empty witness ID", ruleID)
				continue
			}
			if _, duplicate := byID[witness.ID]; duplicate {
				t.Errorf("%s has duplicate witness ID %q", ruleID, witness.ID)
			}
			byID[witness.ID] = witness
			for _, pillar := range witness.Pillars {
				if pillar.Stem == "?" || pillar.Branch == "?" {
					continue
				}
				if _, err := (tyme.SixtyCycle{}).FromName(pillar.Stem + pillar.Branch); err != nil {
					t.Errorf("%s/%s has invalid non-mutated pillar %s%s: %v", ruleID, witness.ID, pillar.Stem, pillar.Branch, err)
				}
			}
			if got := evaluateComplexPatternWitness(ruleID, witness); got != witness.ExpectedOutputName {
				t.Errorf("%s/%s output = %q, want %q", ruleID, witness.ID, got, witness.ExpectedOutputName)
			}
			if witness.BaselineID == "" {
				positiveCount++
			} else {
				mutationCount++
			}
		}
		for _, witness := range witnesses {
			if witness.BaselineID == "" {
				if witness.ExpectedOutputName == "" || witness.Mutation != "" {
					t.Errorf("%s baseline %s is incomplete: %+v", ruleID, witness.ID, witness)
				}
				continue
			}
			baseline, ok := byID[witness.BaselineID]
			if !ok || baseline.ExpectedOutputName == "" {
				t.Errorf("%s mutant %s has invalid baseline %q", ruleID, witness.ID, witness.BaselineID)
				continue
			}
			if witness.Mutation == "" || witness.ExpectedOutputName != "" || patternWitnessMutationDistance(baseline, witness) != 1 {
				t.Errorf("%s mutant %s is not a single killed mutation: %+v", ruleID, witness.ID, witness)
			}
		}
		wantPositive, wantMutations := 1, 5
		if ruleID == "pattern.special.zhuanwang" {
			wantPositive, wantMutations = 5, 18
		}
		if positiveCount != wantPositive || mutationCount != wantMutations {
			t.Errorf("%s witness counts = %d positive/%d mutations", ruleID, positiveCount, mutationCount)
		}
	}
}

func TestComplexPatternBehaviorWitnessProfilesAreIndependentAndBound(t *testing.T) {
	for _, ruleID := range []string{"pattern.special.zhuanwang", "pattern.special.liangqi"} {
		witnesses := patternDetectorBehaviorWitnessProfile(ruleID)
		witnesses[0].ID = "mutated"
		witnesses[0].Pillars[0] = patternWitnessPillar{Stem: "?", Branch: "?"}
		fresh := patternDetectorBehaviorWitnessProfile(ruleID)
		if fresh[0].ID == "mutated" || fresh[0].Pillars[0] == (patternWitnessPillar{Stem: "?", Branch: "?"}) {
			t.Fatalf("%s fresh behavior witnesses inherited mutation: %+v", ruleID, fresh[0])
		}

		profile, ok := patternDetectorSemanticProfile(ruleID)
		envelope, envelopeOK := profile.(patternDetectorSemanticEnvelope)
		if !ok || !envelopeOK || !reflect.DeepEqual(envelope.BehaviorWitnesses, fresh) {
			t.Errorf("%s semantic envelope omits behavior witnesses: %#v", ruleID, profile)
		}
	}
	if witnesses := patternDetectorBehaviorWitnessProfile("pattern.lu.jianlu"); witnesses != nil {
		t.Errorf("simple exhaustive detector unexpectedly has mutation witnesses: %+v", witnesses)
	}
}

func TestComplexPatternBehaviorWitnessMetadataContract(t *testing.T) {
	if PatternSchemaVersion != "pattern-candidates-2026-07-17.27" ||
		PatternDetectorProfile != "classical_structural_detectors_v45" || patternDetectorCount() != 10 {
		t.Fatalf("pattern detector contract = %s/%s/%d", PatternSchemaVersion, PatternDetectorProfile, patternDetectorCount())
	}
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "pattern_candidates" {
			continue
		}
		for _, fragment := range []string{
			"专旺与两气复杂检测器缺少可序列化的单原子输入变异见证",
			"behavior_witnesses",
			"基线成立盘与单字段或单柱删除变异",
			"结构缺失、天干克神、结构外克神、未知干支、柱数、第三气及非四四均分",
		} {
			if !strings.Contains(table.Description, fragment) {
				t.Errorf("pattern description missing %q: %s", fragment, table.Description)
			}
		}
		return
	}
	t.Fatal("pattern-candidate rule table not found")
}
