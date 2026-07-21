package store

import (
	"testing"

	"bazi/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestFeedbackSummaryPreservesTargetVersionAndResearchConsent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.FortuneFeedback{}); err != nil {
		t.Fatal(err)
	}
	rows := []model.FortuneFeedback{
		{UserID: 1, ChartID: 7, TargetType: "interpretation_section", TargetID: "pattern:格局", Rating: model.FeedbackRatingAccurate, EngineVersion: "e1", RuleVersion: "r1", ConsentResearch: true},
		{UserID: 1, ChartID: 7, TargetType: "interpretation_section", TargetID: "pattern:格局", Rating: model.FeedbackRatingAccurate, EngineVersion: "e1", RuleVersion: "r1", ConsentResearch: true},
		{UserID: 1, ChartID: 7, TargetType: "interpretation_section", TargetID: "tiaohou:调候", Rating: model.FeedbackRatingInaccurate, EngineVersion: "e2", RuleVersion: "r2"},
		{UserID: 2, ChartID: 7, TargetType: "interpretation_section", TargetID: "pattern:格局", Rating: model.FeedbackRatingHelpful, EngineVersion: "e1", RuleVersion: "r1", ConsentResearch: true},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}

	items, total, researchEligible, err := NewDBFeedbackStore(db).SummaryByChartID(1, 7)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || researchEligible != 2 || len(items) != 2 {
		t.Fatalf("summary totals = total:%d research:%d items:%+v", total, researchEligible, items)
	}
	for _, item := range items {
		if item.TargetType == "interpretation_section" && item.TargetID == "pattern:格局" &&
			item.Rating == model.FeedbackRatingAccurate && item.EngineVersion == "e1" &&
			item.RuleVersion == "r1" && item.Count == 2 {
			return
		}
	}
	t.Fatalf("versioned pattern feedback group missing: %+v", items)
}
