package fortune

import (
	"testing"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
)

func TestAnalyzeDailyFortuneWithBaseScoreIncludesWeightedCategories(t *testing.T) {
	chart := &bazipkg.BaziResult{
		YearPillar:  model.Pillar{Gan: "甲", Zhi: "子"},
		MonthPillar: model.Pillar{Gan: "丙", Zhi: "寅"},
		DayPillar:   model.Pillar{Gan: "戊", Zhi: "午"},
		HourPillar:  model.Pillar{Gan: "庚", Zhi: "申"},
		BodyStrength: bazipkg.BodyStrengthResult{
			Verdict: "身旺",
			Like:    []string{"金", "水"},
			Dislike: []string{"火", "土"},
		},
	}

	analysis := AnalyzeDailyFortuneWithBaseScore(chart, "甲", "子", 50)
	if analysis == nil {
		t.Fatal("analysis should not be nil")
	}
	if got, want := len(analysis.Categories), len(dailyCategoryWeights); got != want {
		t.Fatalf("category count = %d, want %d", got, want)
	}
	if analysis.Overall.BaseScore != 50 {
		t.Fatalf("base score = %d, want 50", analysis.Overall.BaseScore)
	}
	if analysis.Overall.DetailScore < 20 || analysis.Overall.DetailScore > 100 {
		t.Fatalf("detail score out of range: %d", analysis.Overall.DetailScore)
	}
	if got, want := analysis.Overall.Score, blendScores(50, analysis.Overall.DetailScore); got != want {
		t.Fatalf("overall score = %d, want %d", got, want)
	}
	if analysis.Overall.Level == "" || analysis.Overall.Stars == "" || analysis.Overall.Summary == "" || analysis.Overall.KeyTip == "" {
		t.Fatalf("overall analysis should expose score metadata: %+v", analysis.Overall)
	}

	weightTotal := 0
	for i, category := range analysis.Categories {
		if category.Name != dailyCategoryWeights[i].name {
			t.Fatalf("category[%d] name = %s, want %s", i, category.Name, dailyCategoryWeights[i].name)
		}
		if category.Score < 20 || category.Score > 100 {
			t.Fatalf("%s score out of range: %d", category.Name, category.Score)
		}
		if category.Weight <= 0 {
			t.Fatalf("%s weight should be positive", category.Name)
		}
		if category.Level == "" || category.Trend == "" || category.Stars == "" {
			t.Fatalf("%s should expose level/trend/stars: %+v", category.Name, category)
		}
		if len(category.Keywords) < 4 {
			t.Fatalf("%s should expose actionable keywords: %+v", category.Name, category.Keywords)
		}
		if category.Analysis == "" || category.Advice == "" {
			t.Fatalf("%s should include analysis and advice: %+v", category.Name, category)
		}
		weightTotal += category.Weight
	}
	if weightTotal != 100 {
		t.Fatalf("category weights should sum to 100, got %d", weightTotal)
	}
}

func TestBlendScoresClampsAndWeightsDetailScore(t *testing.T) {
	if got, want := blendScores(50, 80), 71; got != want {
		t.Fatalf("blendScores(50, 80) = %d, want %d", got, want)
	}
	if got := blendScores(-20, 140); got != 70 {
		t.Fatalf("blendScores should clamp inputs, got %d", got)
	}
}
