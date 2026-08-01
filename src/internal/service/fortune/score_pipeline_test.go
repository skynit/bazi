package fortune

import (
	"strings"
	"testing"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
)

func TestBuildScorePipelineUsesSingleFinalScore(t *testing.T) {
	chart := &bazipkg.BaziResult{
		YearPillar:  model.Pillar{Gan: "庚", Zhi: "午"},
		MonthPillar: model.Pillar{Gan: "壬", Zhi: "午"},
		DayPillar:   model.Pillar{Gan: "甲", Zhi: "子"},
		HourPillar:  model.Pillar{Gan: "戊", Zhi: "辰"},
	}
	got := buildScorePipeline(chart, "shengWo", "clash", "甲", "壬")
	wantRelation := 38 // 50 + 18 - 30
	if got.RelationScore != wantRelation {
		t.Fatalf("relation score = %d, want %d", got.RelationScore, wantRelation)
	}
	if got.FinalScore != wantRelation {
		t.Fatalf("final score = %d, want structural relation score %d", got.FinalScore, wantRelation)
	}
	if got.ScoreKind != "structural_relation_index" || got.EvidenceBasis != "empirical" || got.ValidationStatus != "not_validated" || got.InterpretationStatus != "not_adjudicated" || got.IsOutcomeProbability {
		t.Fatalf("score governance metadata is incomplete: %+v", got)
	}
	if len(got.SupportingEvidence) == 0 || len(got.CounterEvidence) == 0 {
		t.Fatalf("expected both evidence directions, got supporting=%+v counter=%+v", got.SupportingEvidence, got.CounterEvidence)
	}
	for _, item := range append(got.SupportingEvidence, got.CounterEvidence...) {
		if item.EvidenceBasis != "empirical" || item.ValidationStatus != "not_validated" || item.InterpretationStatus != "not_adjudicated" || item.IsOutcomeConclusion {
			t.Fatalf("score evidence governance metadata is incomplete: %+v", item)
		}
	}
	if got.EvidenceCompleteness != 100 {
		t.Fatalf("evidence completeness = %d, want 100", got.EvidenceCompleteness)
	}
}

func TestRelationScoreStageExplainsPartialBranchCombinations(t *testing.T) {
	cases := []struct {
		relation string
		impact   int
		code     string
	}{
		{"banHe", 6, "relation.branch.banHe"},
		{"gongHe", 4, "relation.branch.gongHe"},
		{"banHui", 5, "relation.branch.banHui"},
	}
	for _, tc := range cases {
		t.Run(tc.relation, func(t *testing.T) {
			score, evidence := relationScoreStage("unknown", tc.relation, "", "")
			if score != 50+tc.impact {
				t.Fatalf("score = %d, want %d", score, 50+tc.impact)
			}
			if len(evidence) != 1 || evidence[0].Code != tc.code || evidence[0].Impact != tc.impact {
				t.Fatalf("unexpected evidence: %+v", evidence)
			}
		})
	}
}

func TestStemScoreEvidenceUsesReadableDirectionLabels(t *testing.T) {
	for relation, rule := range stemScoreRules {
		for _, forbidden := range []string{"生我", "我生", "克我", "我克"} {
			if strings.Contains(rule.detail, forbidden) {
				t.Fatalf("stem relation %s exposes internal direction label %q: %s", relation, forbidden, rule.detail)
			}
		}
	}

	_, evidence := relationScoreStage("unknown", "neutral", "甲", "己")
	if len(evidence) != 1 || strings.Contains(evidence[0].Description, "未裁决") {
		t.Fatalf("five-combine evidence must explain its boundary in user language: %+v", evidence)
	}
}

func TestScorePipelineEvidenceSlicesAreNeverNil(t *testing.T) {
	got := buildScorePipeline(nil, "unknown", "neutral", "", "")
	if got.SupportingEvidence == nil || got.CounterEvidence == nil {
		t.Fatalf("evidence arrays must serialize as [] instead of null: %+v", got)
	}
}
