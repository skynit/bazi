package ziwei

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestBuildHemingStemRelation_FiveCombines(t *testing.T) {
	tests := []struct {
		stemA, stemB int
		target       string
		ruleID       string
	}{
		{0, 5, "土", "heming.year-stem.five-combine.jia-ji"},
		{1, 6, "金", "heming.year-stem.five-combine.yi-geng"},
		{2, 7, "水", "heming.year-stem.five-combine.bing-xin"},
		{3, 8, "木", "heming.year-stem.five-combine.ding-ren"},
		{4, 9, "火", "heming.year-stem.five-combine.wu-gui"},
	}
	for _, tc := range tests {
		for _, pair := range [][2]int{{tc.stemA, tc.stemB}, {tc.stemB, tc.stemA}} {
			relation, ok := buildHemingStemRelation(pair[0], pair[1])
			if !ok {
				t.Fatalf("five-combine %d,%d was rejected", pair[0], pair[1])
			}
			if relation.RelationType != "five_combine" || relation.RelationLabel != "天干五合" ||
				relation.Direction != "mutual" || relation.FiveCombineTarget != tc.target ||
				relation.StructureStatus != "complete_structure" || relation.TransformationStatus != "unadjudicated" ||
				relation.RuleID != tc.ruleID || relation.IsOutcomeConclusion {
				t.Fatalf("unexpected five-combine relation for %d,%d: %+v", pair[0], pair[1], relation)
			}
		}
	}
}

func TestBuildHemingStemRelation_ElementRelations(t *testing.T) {
	tests := []struct {
		name         string
		stemA, stemB int
		relationType string
		direction    string
	}{
		{name: "same element", stemA: 0, stemB: 1, relationType: "same_element", direction: "mutual"},
		{name: "wood generates fire", stemA: 0, stemB: 2, relationType: "generates", direction: "a_to_b"},
		{name: "water generates wood", stemA: 8, stemB: 0, relationType: "generates", direction: "a_to_b"},
		{name: "reverse generation", stemA: 2, stemB: 0, relationType: "generates", direction: "b_to_a"},
		{name: "wood controls earth", stemA: 0, stemB: 4, relationType: "controls", direction: "a_to_b"},
		{name: "metal controls wood", stemA: 6, stemB: 0, relationType: "controls", direction: "a_to_b"},
		{name: "reverse control", stemA: 4, stemB: 0, relationType: "controls", direction: "b_to_a"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			relation, ok := buildHemingStemRelation(tc.stemA, tc.stemB)
			if !ok {
				t.Fatal("valid stems were rejected")
			}
			if relation.RelationType != tc.relationType || relation.Direction != tc.direction {
				t.Fatalf("relation = %+v, want type=%s direction=%s", relation, tc.relationType, tc.direction)
			}
			if relation.FiveCombineTarget != "" || relation.TransformationStatus != "not_applicable" ||
				relation.StructureStatus != "observed_relation" || relation.RuleID == "" || relation.IsOutcomeConclusion {
				t.Fatalf("invalid non-combine contract: %+v", relation)
			}
		})
	}
}

func TestBuildHemingStemRelation_AllOrderedPairsDeterministic(t *testing.T) {
	validTypes := map[string]bool{
		"five_combine": true,
		"same_element": true,
		"generates":    true,
		"controls":     true,
	}
	for stemA := range StemNames {
		for stemB := range StemNames {
			first, ok := buildHemingStemRelation(stemA, stemB)
			if !ok {
				t.Fatalf("ordered pair %d,%d was rejected", stemA, stemB)
			}
			second, ok := buildHemingStemRelation(stemA, stemB)
			if !ok || !reflect.DeepEqual(second, first) {
				t.Fatalf("ordered pair %d,%d is not deterministic: first=%+v second=%+v", stemA, stemB, first, second)
			}
			if first.StemA == "" || first.StemB == "" || first.ElementA == "" || first.ElementB == "" ||
				!validTypes[first.RelationType] || first.RelationLabel == "" || first.Direction == "" ||
				first.RuleID == "" || first.EvidenceBasis != "deterministic_traditional_rule" ||
				first.ValidationStatus != "not_adjudicated" || first.Notes == "" {
				t.Fatalf("incomplete contract for ordered pair %d,%d: %+v", stemA, stemB, first)
			}
		}
	}

	for _, pair := range [][2]int{{-1, 0}, {0, -1}, {10, 0}, {0, 10}} {
		if got, ok := buildHemingStemRelation(pair[0], pair[1]); ok || got != (HemingStemRelation{}) {
			t.Fatalf("invalid pair %d,%d must be rejected: ok=%v got=%+v", pair[0], pair[1], ok, got)
		}
	}
}

func TestAnalyzeHeming_StemRelationReplayAndLegacyLabels(t *testing.T) {
	chartA, chartB := calculateKnowledgeFixtures(t)
	fresh := analyzeHeming(chartA, chartB)
	replayed := analyzeHeming(roundTripProjectionFixture(t, chartA), roundTripProjectionFixture(t, chartB))
	if fresh == nil || replayed == nil {
		t.Fatalf("valid published charts must produce heming results: fresh=%+v replayed=%+v", fresh, replayed)
	}
	if !reflect.DeepEqual(replayed.StemRelation, fresh.StemRelation) {
		t.Fatalf("stem relation changed after JSON replay: got=%+v want=%+v", replayed.StemRelation, fresh.StemRelation)
	}

	payload, err := json.Marshal(fresh)
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, forbidden := range []string{
		"stem_structure", "帝旺格", "长生格", "墓绝格", "普通格",
		"rule_score", "shuang_gong_lian_can", "wu_bu_xing_yun",
		"chart_a_score", "chart_b_score", "difference_band", "score_band",
		"夫妻宫对冲", "福德宫共鸣", "官禄宫合作", "财帛宫流通",
	} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("heming output retains misleading legacy label %q: %s", forbidden, text)
		}
	}
}
