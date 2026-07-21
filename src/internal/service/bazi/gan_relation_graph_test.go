package bazi

import (
	"strings"
	"testing"

	"bazi/internal/service/data"
)

func TestGanRelationGraphRecordsAllFiveCombinesWithoutClaimingTransformation(t *testing.T) {
	tests := []struct {
		a, b, target string
	}{
		{"甲", "己", "土"}, {"乙", "庚", "金"}, {"丙", "辛", "水"}, {"丁", "壬", "木"}, {"戊", "癸", "火"},
	}
	for _, tc := range tests {
		t.Run(tc.a+tc.b, func(t *testing.T) {
			pillars := testGanRelationPillars([]string{tc.a, tc.b}, []string{"子", "辰"})
			relations := buildGanRelationGraph(pillars)
			got := findGanRelation(relations, "五合", "stem.combine."+canonicalGanPair(tc.a, tc.b))
			if got == nil {
				t.Fatalf("missing combine in %+v", relations)
			}
			if got.Status != "observed" || got.StructureStatus != "complete_structure" || got.TransformationStatus != "unadjudicated" || got.TargetElement != tc.target || got.TransformationEvidence == nil {
				t.Fatalf("combine = %+v", *got)
			}
			if strings.Contains(got.Detail, "合化为") || strings.Contains(got.Detail, "已经成化") {
				t.Fatalf("detail overclaims transformation: %q", got.Detail)
			}
		})
	}
}

func TestGanElementRelationsCoverBothYinYangStems(t *testing.T) {
	tests := []struct {
		a, b, relation, direction string
	}{
		{"甲", "己", "相克", "year_to_month"},
		{"乙", "戊", "相克", "year_to_month"},
		{"丁", "庚", "相克", "year_to_month"},
		{"癸", "丙", "相克", "year_to_month"},
		{"甲", "丁", "相生", "year_to_month"},
		{"丁", "甲", "相生", "month_to_year"},
		{"甲", "乙", "比和", "mutual"},
	}
	for _, tc := range tests {
		t.Run(tc.a+tc.b, func(t *testing.T) {
			got, ok := newGanElementRelation(
				ganRelationPillar{key: "year", label: labelYear, stem: tc.a, branch: "子"},
				ganRelationPillar{key: "month", label: labelMonth, stem: tc.b, branch: "寅"},
			)
			if !ok || got.Type != tc.relation || got.Direction != tc.direction || got.TransformationStatus != "not_applicable" {
				t.Fatalf("element relation = %+v, ok=%v", got, ok)
			}
		})
	}
}

func TestGanRelationGraphRecordsAllFourClashesAlongsideElementControl(t *testing.T) {
	for _, tc := range []struct{ a, b string }{
		{"甲", "庚"}, {"乙", "辛"}, {"丙", "壬"}, {"丁", "癸"},
	} {
		t.Run(tc.a+tc.b, func(t *testing.T) {
			for _, stems := range [][]string{{tc.a, tc.b}, {tc.b, tc.a}} {
				relations := buildGanRelationGraph(testGanRelationPillars(stems, []string{"子", "寅"}))
				clash := findGanRelation(relations, "天干相冲", "stem.clash."+canonicalGanPair(tc.a, tc.b))
				if clash == nil || clash.Subtype != "四冲" || clash.Status != "observed" ||
					clash.StructureStatus != "complete_structure" || clash.TransformationStatus != "not_applicable" ||
					clash.Direction != "mutual" || len(clash.Evidence) != 2 {
					t.Fatalf("%s%s clash = %+v", stems[0], stems[1], clash)
				}
				if !hasGanRelationType(relations, "相克") {
					t.Fatalf("%s%s lost directional element control relation: %+v", stems[0], stems[1], relations)
				}
			}
		})
	}

	for _, stems := range [][]string{{"甲", "戊"}, {"乙", "己"}, {"丙", "庚"}, {"丁", "辛"}, {"戊", "壬"}, {"己", "癸"}} {
		relations := buildGanRelationGraph(testGanRelationPillars(stems, []string{"子", "寅"}))
		if hasGanRelationType(relations, "天干相冲") || !hasGanRelationType(relations, "相克") {
			t.Errorf("ordinary control %s%s misclassified as four-clash: %+v", stems[0], stems[1], relations)
		}
	}
}

func TestGanRelationManifestPublishesFourClashScope(t *testing.T) {
	for _, table := range DefaultRuleMeta().Tables {
		if table.Key != "stem_relations" {
			continue
		}
		if table.Version != "2026-07-18.1" || table.School != "天干五合四冲生克" ||
			!strings.Contains(table.Source, "0194eb4574f33ab056fe7cac62a9d8bf24272478") ||
			!strings.Contains(table.Description, "甲庚、乙辛、丙壬、丁癸四冲") ||
			!strings.Contains(table.Description, "四冲与五行相克方向分别记录") {
			t.Fatalf("stem relation manifest does not publish runtime four-clash scope: %+v", table)
		}
		return
	}
	t.Fatal("stem_relations table missing from runtime rule manifest")
}

func TestGanRelationGraphMarksCompetingCombinesDisputed(t *testing.T) {
	relations := buildGanRelationGraph(testGanRelationPillars(
		[]string{"甲", "己", "甲", "丙"}, []string{"子", "辰", "午", "酉"},
	))
	combines := make([]GanRelation, 0, 2)
	for _, relation := range relations {
		if relation.Type == "五合" {
			combines = append(combines, relation)
		}
	}
	if len(combines) != 2 {
		t.Fatalf("competing combine count = %d: %+v", len(combines), relations)
	}
	for _, relation := range combines {
		if relation.Status != "disputed" || relation.TransformationStatus != "disputed" || len(relation.ConflictsWith) != 1 || !strings.Contains(strings.Join(relation.DisputeReasons, ";"), "争合/妒合") {
			t.Fatalf("competing combine = %+v", relation)
		}
	}
}

func TestGanCombineEvidenceIsExplicitButNonDecisive(t *testing.T) {
	pillars := testGanRelationPillars(
		[]string{"甲", "戊", "己", "庚"}, []string{"子", "辰", "午", "申"},
	)
	relations := buildGanRelationGraph(pillars)
	got := findGanRelation(relations, "五合", "stem.combine.甲己")
	if got == nil || got.Proximity != "remote" || got.TransformationEvidence == nil {
		t.Fatalf("remote combine = %+v", got)
	}
	evidence := got.TransformationEvidence
	if evidence.MonthBranch != "辰" || evidence.MonthElement != "土" || !evidence.MonthSupportsTarget || !evidence.TargetStemExposed || len(evidence.TargetRootBranches) == 0 {
		t.Fatalf("transformation evidence = %+v", evidence)
	}
	if got.TransformationStatus != "unadjudicated" {
		t.Fatalf("evidence must not auto-transform: %+v", got)
	}
}

func TestGanRelationGraphDoesNotDependOnMutableExportedElementMap(t *testing.T) {
	originalJia := data.GanElement["甲"]
	originalBing := data.GanElement["丙"]
	originalWu := data.GanElement["戊"]
	defer func() {
		data.GanElement["甲"] = originalJia
		data.GanElement["丙"] = originalBing
		data.GanElement["戊"] = originalWu
	}()
	data.GanElement["甲"] = "金"
	data.GanElement["丙"] = "水"
	data.GanElement["戊"] = "木"

	relation, ok := newGanElementRelation(
		ganRelationPillar{key: "year", label: labelYear, stem: "甲", branch: "子"},
		ganRelationPillar{key: "month", label: labelMonth, stem: "丙", branch: "寅"},
	)
	if !ok || relation.Type != "相生" || relation.Direction != "year_to_month" {
		t.Fatalf("mutable exported map changed fixed stem relation: %+v, ok=%v", relation, ok)
	}

	combine := findGanRelation(buildGanRelationGraph(testGanRelationPillars(
		[]string{"甲", "戊", "己", "庚"}, []string{"子", "辰", "午", "申"},
	)), "五合", "stem.combine.甲己")
	if combine == nil || combine.TransformationEvidence == nil || !combine.TransformationEvidence.TargetStemExposed {
		t.Fatalf("mutable exported map changed transformation evidence: %+v", combine)
	}
}

func testGanRelationPillars(stems, branches []string) []ganRelationPillar {
	keys := []string{"year", "month", "day", "hour"}
	labels := []string{labelYear, labelMonth, labelDay, labelHour}
	result := make([]ganRelationPillar, len(stems))
	for i := range stems {
		result[i] = ganRelationPillar{key: keys[i], label: labels[i], stem: stems[i], branch: branches[i]}
	}
	return result
}

func findGanRelation(relations []GanRelation, relationType, ruleID string) *GanRelation {
	for i := range relations {
		if relations[i].Type == relationType && relations[i].RuleID == ruleID {
			return &relations[i]
		}
	}
	return nil
}

func hasGanRelationType(relations []GanRelation, relationType string) bool {
	for _, relation := range relations {
		if relation.Type == relationType {
			return true
		}
	}
	return false
}
