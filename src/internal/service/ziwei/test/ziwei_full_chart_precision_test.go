package ziwei_test

import (
	"bazi/internal/service/ziwei"
	"testing"
)

// TestLegacyBronzeChartsHaveCompleteInternalStructure uses quarantined cases
// only as broad input coverage. It verifies invariants of one engine output and
// deliberately does not compare the unadjudicated expected text.
func TestLegacyBronzeChartsHaveCompleteInternalStructure(t *testing.T) {
	data, err := loadZiWeiTestData("../../testdata/ziwei_cases.json")
	if err != nil {
		t.Fatalf("load legacy Ziwei fixtures: %v", err)
	}
	service := ziwei.NewZiWeiService()
	majorStars := map[string]struct{}{
		"紫微": {}, "天机": {}, "太阳": {}, "武曲": {}, "天同": {}, "廉贞": {}, "天府": {},
		"太阴": {}, "贪狼": {}, "巨门": {}, "天相": {}, "天梁": {}, "七杀": {}, "破军": {},
	}
	for _, tc := range data.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			chart, err := service.CalculateChart(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, tc.Gender)
			if err != nil {
				t.Fatal(err)
			}
			if chart.ProfileID != ziwei.DefaultProfileID || len(chart.PluginManifestHash) != 64 {
				t.Fatalf("missing reproducibility contract: %+v", chart)
			}
			palaceNames, branches := make(map[string]struct{}), make(map[string]struct{})
			majorCounts := make(map[string]int)
			bodyPalaces, fourHua := 0, 0
			for _, palace := range chart.Palaces {
				if palace.Name == "" || palace.Branch == "" || palace.HeavenlyStem == "" {
					t.Fatalf("incomplete palace: %+v", palace)
				}
				palaceNames[palace.Name] = struct{}{}
				branches[palace.Branch] = struct{}{}
				if palace.IsBodyPalace {
					bodyPalaces++
					if palace.Branch != chart.BodyPalace {
						t.Fatalf("body palace branch mismatch: palace=%s chart=%s", palace.Branch, chart.BodyPalace)
					}
				}
				for _, star := range publishedMainStarNames(palace) {
					if _, ok := majorStars[star]; !ok {
						t.Fatalf("unknown major star %q", star)
					}
					majorCounts[star]++
				}
				fourHua += len(palace.FourHua)
			}
			if len(palaceNames) != 12 || len(branches) != 12 || bodyPalaces != 1 {
				t.Fatalf("chart structure: palace names=%d branches=%d body markers=%d", len(palaceNames), len(branches), bodyPalaces)
			}
			for star := range majorStars {
				if majorCounts[star] != 1 {
					t.Fatalf("major star %s count=%d, want 1", star, majorCounts[star])
				}
			}
			if fourHua != 4 {
				t.Fatalf("four-hua count=%d, want 4", fourHua)
			}
			juValue, ok := ziwei.FiveBureauValue[chart.FiveBureau]
			if !ok || juValue < 2 || juValue > 6 || len(service.CalculateDayun(chart)) == 0 {
				t.Fatalf("invalid bureau/dayun: bureau=%q value=%d/%t", chart.FiveBureau, juValue, ok)
			}
		})
	}
}
