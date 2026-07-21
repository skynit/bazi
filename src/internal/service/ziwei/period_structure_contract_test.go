package ziwei

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestPeriodPublicContractOmitsHeuristicScores(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	liunian := svc.CalculateLiunian(base, 2026)
	liuyue := svc.CalculateLiuyueForDate(base, 2026, 3, 15)
	liuri := svc.CalculateLiuriForDate(base, 2026, 3, 15)
	interpreter := NewPeriodInterpreterFromChart(base)
	if liunian == nil || liuyue == nil || liuri == nil || interpreter == nil {
		t.Fatal("valid period input returned nil")
	}

	payload, err := json.Marshal(map[string]any{
		"liunian_interpretation": interpreter.AnalyzeLiunian(liunian, 2026),
		"liuyue_interpretation":  interpreter.AnalyzeLiuyue(liuyue, 2026, 3, 15),
		"liuri_interpretation":   interpreter.AnalyzeLiuri(liuri, 2026, 3, 15),
		"summary":                interpreter.SummarizeAll(liunian, liuyue, liuri, 2026, 3, 15),
		"dayun_analysis":         BuildDayunAnalysis(base, nil, 23),
		"liunian_analysis":       BuildLiunianAnalysis(base, liunian, 2026),
		"liuyue_analysis":        BuildLiuyueAnalysis(base, liuyue, 2026, 3, 15),
		"liuri_analysis":         BuildLiuriAnalysis(base, liuri, 2026, 3, 15),
		"liunian_overlay":        svc.AnalyzeLiunianOverlay(base, liunian, 2026),
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, field := range []string{
		`"score"`, `"tone"`, `"level"`, `"overall_tone"`, `"key_tips"`,
		`"effect"`, `"qi_zi_effect"`,
	} {
		if strings.Contains(text, field) {
			t.Fatalf("period contract retained removed field %s: %s", field, text)
		}
	}
	for _, phrase := range []string{
		"经验规则评分", "经验规则分", "经验评分", "时辰分数", "时段评分",
		"规则高区间", "规则中高区间", "规则中区间", "规则低区间",
	} {
		if strings.Contains(text, phrase) {
			t.Fatalf("period contract retained heuristic score phrase %q: %s", phrase, text)
		}
	}
	if strings.Contains(text, `"rule_id":"ziwei.period.branch-relation-v1"`) {
		t.Fatalf("period contract retained obsolete generic branch-relation rule ID: %s", text)
	}
	if !strings.Contains(text, `"rule_id":"ziwei.period.branch.`) ||
		!strings.Contains(text, `"structural_status":`) ||
		!strings.Contains(text, `"transformation_status":`) ||
		!strings.Contains(text, `"evidence_basis":"deterministic_rule_projection"`) ||
		!strings.Contains(text, `"interpretation_status":"not_adjudicated"`) ||
		!strings.Contains(text, `"is_outcome_conclusion":false`) {
		t.Fatalf("period contract omitted structured branch-relation evidence: %s", text)
	}
}

func TestPeriodMixedEvidenceContractSeparatesPlacementFromInterpretation(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	liunian := svc.CalculateLiunian(base, 2026)
	liuyue := svc.CalculateLiuyueForDate(base, 2026, 3, 15)
	liuri := svc.CalculateLiuriForDate(base, 2026, 3, 15)
	analyses := []*PeriodAnalysis{
		BuildDayunAnalysis(base, nil, 23),
		BuildLiunianAnalysis(base, liunian, 2026),
		BuildLiuyueAnalysis(base, liuyue, 2026, 3, 15),
		BuildLiuriAnalysis(base, liuri, 2026, 3, 15),
	}
	for _, analysis := range analyses {
		if analysis == nil {
			t.Fatal("valid period input returned nil analysis")
		}
		if analysis.EvidenceBasis != periodMixedEvidenceBasis ||
			analysis.PlacementBasis != periodPlacementBasis ||
			analysis.InterpretationBasis != periodInterpretationBasis ||
			analysis.InterpretationStatus != periodInterpretationStatus ||
			analysis.ValidationStatus != periodInterpretationStatus || analysis.IsOutcomeConclusion {
			t.Fatalf("period analysis collapsed mixed evidence semantics: %+v", analysis)
		}
		for _, item := range analysis.Highlights {
			assertPeriodEvidenceSemantics(t, item.PeriodEvidenceSemantics)
		}
		for _, item := range analysis.FocusPalaces {
			assertPeriodEvidenceSemantics(t, item.PeriodEvidenceSemantics)
		}
		for _, item := range analysis.Evidence {
			assertPeriodEvidenceSemantics(t, item.PeriodEvidenceSemantics)
		}
		for _, item := range analysis.DayunStages {
			assertPeriodEvidenceSemantics(t, item.PeriodEvidenceSemantics)
		}
	}

	overlay := svc.AnalyzeLiunianOverlay(base, liunian, 2026)
	if overlay == nil {
		t.Fatal("valid period input returned nil overlay")
	}
	if overlay.EvidenceBasis != periodMixedEvidenceBasis ||
		overlay.PlacementBasis != periodPlacementBasis ||
		overlay.InterpretationBasis != periodInterpretationBasis ||
		overlay.InterpretationStatus != periodInterpretationStatus ||
		overlay.ValidationStatus != periodInterpretationStatus || overlay.IsOutcomeConclusion {
		t.Fatalf("overlay collapsed mixed evidence semantics: %+v", overlay)
	}
	for _, item := range overlay.Method {
		assertPeriodEvidenceSemantics(t, item.PeriodEvidenceSemantics)
	}
	for _, item := range append(append([]OverlayTrigger{}, overlay.FourHua...), overlay.AnnualStars...) {
		assertPeriodEvidenceSemantics(t, item.PeriodEvidenceSemantics)
	}
	for _, item := range overlay.FocusPalaces {
		assertPeriodEvidenceSemantics(t, item.PeriodEvidenceSemantics)
	}

	payload, err := json.Marshal(map[string]any{"analyses": analyses, "overlay": overlay})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	if strings.Contains(text, `"evidence_basis":"empirical"`) || strings.Contains(text, `"impact"`) {
		t.Fatalf("period output retained ambiguous evidence semantics: %s", text)
	}
	for _, fragment := range []string{
		`"evidence_basis":"mixed_deterministic_projection_and_unadjudicated_traditional_labels"`,
		`"placement_basis":"deterministic_rule_projection"`,
		`"interpretation_basis":"traditional_rule_labels"`,
		`"interpretation_status":"not_adjudicated"`,
		`"is_outcome_conclusion":false`,
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("period output omitted mixed evidence boundary %s: %s", fragment, text)
		}
	}
}

func assertPeriodEvidenceSemantics(t *testing.T, got PeriodEvidenceSemantics) {
	t.Helper()
	if got.PlacementBasis != periodPlacementBasis ||
		got.InterpretationBasis != periodInterpretationBasis ||
		got.InterpretationStatus != periodInterpretationStatus || got.IsOutcomeConclusion {
		t.Fatalf("period item omitted evidence semantics: %+v", got)
	}
}

func TestPeriodBranchElementsMatchCanonicalMapping(t *testing.T) {
	want := []string{"水", "土", "木", "木", "土", "火", "火", "土", "金", "金", "土", "水"}
	for branch := range BranchNames {
		if got := wuXingBranch(branch); got != want[branch] {
			t.Fatalf("%s element = %q, want %q", BranchNames[branch], got, want[branch])
		}
	}
	if wuXingBranch(-1) != "" || wuXingBranch(len(BranchNames)) != "" {
		t.Fatal("invalid branch index produced an element")
	}
}

func TestPeriodBranchRelationsPreserveOverlappingStructures(t *testing.T) {
	for left := range BranchNames {
		for right := range BranchNames {
			got := relationTypes(left, right)
			reverse := relationTypes(right, left)
			if !reflect.DeepEqual(got, reverse) {
				t.Fatalf("relation is not symmetric for %s/%s: %v vs %v", BranchNames[left], BranchNames[right], got, reverse)
			}
		}
	}

	cases := []struct {
		left  int
		right int
		want  []string
	}{
		{left: 0, right: 0, want: []string{"伏吟"}},
		{left: 4, right: 4, want: []string{"相刑（自刑）", "伏吟"}},
		{left: 6, right: 6, want: []string{"相刑（自刑）", "伏吟"}},
		{left: 9, right: 9, want: []string{"相刑（自刑）", "伏吟"}},
		{left: 11, right: 11, want: []string{"相刑（自刑）", "伏吟"}},
		{left: 0, right: 3, want: []string{"相刑（无礼之刑）"}},
		{left: 2, right: 8, want: []string{"六冲", "相刑（无恩之刑）"}},
		{left: 5, right: 8, want: []string{"相刑（无恩之刑）", "六破", "六合"}},
		{left: 1, right: 7, want: []string{"六冲", "相刑（恃势之刑）"}},
		{left: 2, right: 5, want: []string{"相刑（无恩之刑）", "六害"}},
		{left: 2, right: 11, want: []string{"六破", "六合"}},
	}
	for _, tt := range cases {
		if got := relationTypes(tt.left, tt.right); !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("relation for %s/%s = %v, want %v", BranchNames[tt.left], BranchNames[tt.right], got, tt.want)
		}
	}
	if len(relationTypes(-1, 0)) != 0 || len(relationTypes(0, len(BranchNames))) != 0 {
		t.Fatal("invalid branch index produced a relation")
	}
}

func TestPeriodBranchRelationsCoverHarmAndBreakTables(t *testing.T) {
	tables := []struct {
		relation string
		pairs    [][2]int
	}{
		{relation: "六害", pairs: [][2]int{{0, 7}, {1, 6}, {2, 5}, {3, 4}, {8, 11}, {9, 10}}},
		{relation: "六破", pairs: [][2]int{{0, 9}, {1, 4}, {2, 11}, {3, 6}, {5, 8}, {7, 10}}},
	}
	for _, table := range tables {
		for _, pair := range table.pairs {
			for _, direction := range [][2]int{pair, {pair[1], pair[0]}} {
				rule, ok := findPeriodRelationRule(direction[0], direction[1], table.relation)
				if !ok {
					t.Fatalf("%s missing for %s/%s", table.relation, BranchNames[direction[0]], BranchNames[direction[1]])
				}
				wantID := "ziwei.period.branch."
				if table.relation == "六害" {
					wantID += "harm."
				} else {
					wantID += "break."
				}
				wantID += canonicalPeriodBranchPair(pair[0], pair[1]) + "-v1"
				if rule.ruleID != wantID || rule.structuralStatus != "observed" ||
					rule.transformationStatus != "not_applicable" || rule.targetElement != "" {
					t.Fatalf("%s contract for %s/%s = %+v, want rule_id=%s observed/not_applicable", table.relation, BranchNames[direction[0]], BranchNames[direction[1]], rule, wantID)
				}
			}
		}
	}
}

func TestPeriodLiuHeContractKeepsTransformationUnadjudicated(t *testing.T) {
	cases := []struct {
		left, right int
		element     string
	}{
		{left: 0, right: 1, element: "土"},
		{left: 2, right: 11, element: "木"},
		{left: 3, right: 10, element: "火"},
		{left: 4, right: 9, element: "金"},
		{left: 5, right: 8, element: "水"},
		{left: 6, right: 7, element: "土"},
	}
	for _, tt := range cases {
		for _, pair := range [][2]int{{tt.left, tt.right}, {tt.right, tt.left}} {
			rule, ok := findPeriodRelationRule(pair[0], pair[1], "六合")
			if !ok {
				t.Fatalf("六合 missing for %s/%s", BranchNames[pair[0]], BranchNames[pair[1]])
			}
			wantID := "ziwei.period.branch.liuhe." + canonicalPeriodBranchPair(tt.left, tt.right) + "-v1"
			if rule.ruleID != wantID || rule.structuralStatus != "complete" ||
				rule.transformationStatus != "unadjudicated" || rule.targetElement != tt.element {
				t.Fatalf("六合 contract for %s/%s = %+v, want rule_id=%s complete/unadjudicated target=%s", BranchNames[pair[0]], BranchNames[pair[1]], rule, wantID, tt.element)
			}
		}
	}
}

func TestPeriodBranchRelationRuleMetadataIsComplete(t *testing.T) {
	wantPriority := map[string]int{
		"六冲": 100,
		"相刑": 90,
		"六害": 80,
		"六破": 70,
		"六合": 60,
		"伏吟": 50,
	}
	for left := range BranchNames {
		for right := range BranchNames {
			for _, rule := range periodPairRelationRules(left, right) {
				if rule.ruleID == "ziwei.period.branch-relation-v1" || !strings.HasPrefix(rule.ruleID, "ziwei.period.branch.") || !strings.HasSuffix(rule.ruleID, "-v1") {
					t.Fatalf("relation for %s/%s has unstable or obsolete rule ID: %+v", BranchNames[left], BranchNames[right], rule)
				}
				if rule.relation == "三刑" {
					t.Fatalf("two-branch relation for %s/%s was mislabeled as complete 三刑: %+v", BranchNames[left], BranchNames[right], rule)
				}
				if rule.priority != wantPriority[rule.relation] {
					t.Fatalf("priority for %s/%s %s = %d, want %d", BranchNames[left], BranchNames[right], rule.relation, rule.priority, wantPriority[rule.relation])
				}
				if rule.relation == "六合" {
					if rule.structuralStatus != "complete" || rule.transformationStatus != "unadjudicated" || rule.targetElement == "" {
						t.Fatalf("六合 metadata for %s/%s is incomplete: %+v", BranchNames[left], BranchNames[right], rule)
					}
				} else if rule.structuralStatus != "observed" || rule.transformationStatus != "not_applicable" || rule.targetElement != "" {
					t.Fatalf("non-combination metadata for %s/%s is invalid: %+v", BranchNames[left], BranchNames[right], rule)
				}
			}
		}
	}
}

func TestPeriodBranchRelationEvidenceSerializesSpecificRuleSemantics(t *testing.T) {
	interpreter := &PeriodInterpreter{birthData: &BirthData{
		YearBranch: 8, MonthPillarBranch: 8, DayBranch: 8, HourBranch: 8,
	}}
	evidence := interpreter.relationEvidence(5)
	if len(evidence) != 12 {
		t.Fatalf("巳/申 should retain three relations for all four pillars, got %d: %+v", len(evidence), evidence)
	}
	payload, err := json.Marshal(evidence)
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, fragment := range []string{
		`"relation":"相刑","subtype":"无恩之刑","rule_id":"ziwei.period.branch.punish.巳申-v1"`,
		`"relation":"六破","rule_id":"ziwei.period.branch.break.巳申-v1"`,
		`"relation":"六合","rule_id":"ziwei.period.branch.liuhe.巳申-v1","structural_status":"complete","transformation_status":"unadjudicated","target_element":"水"`,
		`"evidence_basis":"deterministic_rule_projection"`,
		`"interpretation_status":"not_adjudicated"`,
		`"is_outcome_conclusion":false`,
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("serialized relation evidence omitted %s: %s", fragment, text)
		}
	}
	if strings.Contains(text, "ziwei.period.branch-relation-v1") || strings.Contains(text, `"relation":"三刑"`) {
		t.Fatalf("serialized relation evidence retained obsolete generic ID or two-branch 三刑 label: %s", text)
	}
}

func findPeriodRelationRule(left, right int, relation string) (periodPairRelationRule, bool) {
	for _, rule := range periodPairRelationRules(left, right) {
		if rule.relation == relation {
			return rule, true
		}
	}
	return periodPairRelationRule{}, false
}
