package ziwei

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPeriodAnalysisPublishesCrossLayerRelations(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2000, 8, 16, 3, 0, "女")
	if err != nil {
		t.Fatal(err)
	}

	liuyue := svc.CalculateLiuyueForDate(base, 2023, 3, 15)
	monthAnalysis := BuildLiuyueAnalysis(base, liuyue, 2023, 3, 15)
	if monthAnalysis == nil {
		t.Fatal("valid liuyue analysis returned nil")
	}
	assertPeriodLayerRelation(t, monthAnalysis.CrossLayerRelations, "liuyue", "liunian", "伏吟")

	liuri := svc.CalculateLiuriForDate(base, 2023, 8, 19)
	dayAnalysis := BuildLiuriAnalysis(base, liuri, 2023, 8, 19)
	if dayAnalysis == nil {
		t.Fatal("valid liuri analysis returned nil")
	}
	assertPeriodLayerRelation(t, dayAnalysis.CrossLayerRelations, "liuri", "liunian", "六冲")
	for _, relation := range dayAnalysis.CrossLayerRelations {
		if relation.RuleID == "" || relation.EvidenceBasis != periodPlacementBasis ||
			relation.InterpretationStatus != periodInterpretationStatus || relation.IsOutcomeConclusion {
			t.Fatalf("cross-layer relation semantics are incomplete: %+v", relation)
		}
	}
	payload, err := json.Marshal(dayAnalysis)
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	if !strings.Contains(text, `"cross_layer_relations"`) ||
		!strings.Contains(text, `"rule_id":"ziwei.period.branch.clash.卯酉-v1"`) {
		t.Fatalf("published analysis omitted cross-layer evidence: %s", text)
	}
	if strings.Contains(text, `"score"`) || strings.Contains(text, `"favorable"`) {
		t.Fatalf("cross-layer evidence leaked judgment fields: %s", text)
	}
}

func assertPeriodLayerRelation(t *testing.T, relations []PeriodLayerRelation, source, target, relationName string) {
	t.Helper()
	for _, relation := range relations {
		if relation.SourceLayer == source && relation.TargetLayer == target && relation.Relation == relationName {
			return
		}
	}
	t.Fatalf("missing %s -> %s %s relation: %+v", source, target, relationName, relations)
}
