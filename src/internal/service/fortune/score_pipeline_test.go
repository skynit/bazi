package fortune

import (
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
		BodyStrength: bazipkg.BodyStrengthResult{
			Verdict: "身弱",
			Like:    []string{"水", "木"},
			Dislike: []string{"金", "土"},
		},
	}
	analysis := &FortuneAnalysis{
		Overall: OverallAnalysis{DetailScore: 70},
		Categories: []CategoryScore{
			{Name: "事业", Score: 75, Weight: 18, Analysis: "推进条件较好。"},
			{Name: "健康", Score: 45, Weight: 14, Analysis: "需要控制消耗。"},
		},
	}

	got := buildScorePipeline(chart, "shengWo", "clash", "甲", "壬", analysis, &RikuyoResult{})
	wantRelation := 38 // 50 + 18 - 30
	if got.RelationScore != wantRelation {
		t.Fatalf("relation score = %d, want %d", got.RelationScore, wantRelation)
	}
	if want := blendScores(wantRelation, 70); got.FinalScore != want {
		t.Fatalf("final score = %d, want %d", got.FinalScore, want)
	}
	if analysis.Overall.Score != got.FinalScore || analysis.Overall.BaseScore != got.RelationScore {
		t.Fatalf("analysis and pipeline scores diverged: analysis=%+v pipeline=%+v", analysis.Overall, got)
	}
	if len(got.SupportingEvidence) == 0 || len(got.CounterEvidence) == 0 {
		t.Fatalf("expected both evidence directions, got supporting=%+v counter=%+v", got.SupportingEvidence, got.CounterEvidence)
	}
	if got.EvidenceCompleteness < 80 {
		t.Fatalf("evidence completeness = %d, want >= 80", got.EvidenceCompleteness)
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

func TestScorePipelineEvidenceSlicesAreNeverNil(t *testing.T) {
	got := buildScorePipeline(nil, "unknown", "neutral", "", "", nil, nil)
	if got.SupportingEvidence == nil || got.CounterEvidence == nil {
		t.Fatalf("evidence arrays must serialize as [] instead of null: %+v", got)
	}
}
