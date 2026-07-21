package bazi

import (
	"reflect"
	"testing"

	"bazi/internal/model"
)

func TestTiaohouReviewedChartConditionsUseCompleteFourPillars(t *testing.T) {
	tests := []struct {
		name       string
		pillars    []model.Pillar
		candidates []string
		wantRuleID string
	}{
		{
			name:       "乙酉月金局取丁",
			pillars:    []model.Pillar{{Gan: "己", Zhi: "巳"}, {Gan: "乙", Zhi: "酉"}, {Gan: "乙", Zhi: "丑"}, {Gan: "丁", Zhi: "亥"}},
			candidates: []string{"丁"}, wantRuleID: "bazi.tiaohou.chart.yi-you-metal-frame",
		},
		{
			name:       "丙巳月水局取戊",
			pillars:    []model.Pillar{{Gan: "甲", Zhi: "申"}, {Gan: "己", Zhi: "巳"}, {Gan: "丙", Zhi: "子"}, {Gan: "壬", Zhi: "辰"}},
			candidates: []string{"戊"}, wantRuleID: "bazi.tiaohou.chart.bing-si-water-frame",
		},
		{
			name:       "丁辰月水局壬透取戊己",
			pillars:    []model.Pillar{{Gan: "壬", Zhi: "申"}, {Gan: "甲", Zhi: "辰"}, {Gan: "丁", Zhi: "酉"}, {Gan: "戊", Zhi: "子"}},
			candidates: []string{"戊", "己"}, wantRuleID: "bazi.tiaohou.chart.ding-chen-water-frame-ren-visible",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AnalyzeTiaohouForPillars(tc.pillars[0], tc.pillars[1], tc.pillars[2], tc.pillars[3])
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.ChartCandidates, tc.candidates) || len(got.MatchedConditions) != 1 ||
				got.MatchedConditions[0].RuleID != tc.wantRuleID ||
				got.ChartSelectionBasis != "reviewed_four_pillar_condition_match" {
				t.Fatalf("chart condition result = %+v, want candidates %v rule %s", got, tc.candidates, tc.wantRuleID)
			}
			if !ValidTiaohouEvidenceForPillars(got, tc.pillars) {
				t.Fatalf("chart condition evidence is not reproducible: %+v", got)
			}
		})
	}
}

func TestTiaohouChartConditionRequiresEveryExplicitFact(t *testing.T) {
	tests := []struct {
		name    string
		pillars []model.Pillar
	}{
		{
			name: "丁辰水局无壬透",
			pillars: []model.Pillar{
				{Gan: "庚", Zhi: "申"}, {Gan: "甲", Zhi: "辰"},
				{Gan: "丁", Zhi: "酉"}, {Gan: "戊", Zhi: "子"},
			},
		},
		{
			name: "丙巳水局无壬透",
			pillars: []model.Pillar{
				{Gan: "甲", Zhi: "申"}, {Gan: "己", Zhi: "巳"},
				{Gan: "丙", Zhi: "子"}, {Gan: "戊", Zhi: "辰"},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := AnalyzeTiaohouForPillars(tc.pillars[0], tc.pillars[1], tc.pillars[2], tc.pillars[3])
			if err != nil {
				t.Fatal(err)
			}
			if len(got.ChartEvidence.CompleteBranchStructures) != 1 ||
				got.ChartEvidence.CompleteBranchStructures[0].RuleID != "branch.sanhe.water" {
				t.Fatalf("complete water structure was not observed: %+v", got.ChartEvidence)
			}
			if len(got.MatchedConditions) != 0 || len(got.ChartCandidates) != 0 ||
				got.ChartSelectionBasis != "no_reviewed_chart_condition_match" {
				t.Fatalf("condition must not match without visible 壬: %+v", got)
			}
		})
	}
}

func TestCalculateFromPillarsPublishesChartConditionCandidates(t *testing.T) {
	result, err := (&BaziService{}).CalculateSyntheticPillars("己巳", "乙酉", "乙丑", "丁亥", "MALE")
	if err != nil {
		t.Fatal(err)
	}
	if result.Tiaohou == nil || !reflect.DeepEqual(result.Tiaohou.ChartCandidates, []string{"丁"}) {
		t.Fatalf("calculation pipeline did not use full-chart Tiaohou conditions: %+v", result.Tiaohou)
	}
	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}
	if !ValidTiaohouEvidenceForPillars(result.Tiaohou, pillars) {
		t.Fatalf("calculated chart Tiaohou is not reproducible: %+v", result.Tiaohou)
	}
	tampered := *result.Tiaohou
	tampered.ChartCandidates = []string{"癸"}
	if ValidTiaohouEvidenceForPillars(&tampered, pillars) {
		t.Fatal("four-pillar validation accepted tampered chart candidates")
	}
}
