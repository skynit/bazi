package ziwei

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFlyingStarAnalysisSurvivesJSONRoundTrip(t *testing.T) {
	chart, err := NewZiWeiService().CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart: %v", err)
	}
	want := buildFlyingStarAnalysisFromChart(chart)
	assertFlyingStarAnalysisMatchesPublishedFourHua(t, chart, want)

	data, err := json.Marshal(chart)
	if err != nil {
		t.Fatalf("marshal chart: %v", err)
	}
	var replayed ZiWeiChart
	if err := json.Unmarshal(data, &replayed); err != nil {
		t.Fatalf("unmarshal chart: %v", err)
	}

	got := buildFlyingStarAnalysisFromChart(&replayed)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("JSON-replayed flying-star analysis = %+v, want %+v", got, want)
	}
	assertFlyingStarAnalysisMatchesPublishedFourHua(t, &replayed, got)
}

func TestFlyingStarAnalysisIgnoresMalformedPublishedLabels(t *testing.T) {
	chart := &ZiWeiChart{}
	chart.Palaces[0] = PalaceInfo{
		Name: "命宫",
		FourHua: []string{
			"化禄",
			"廉贞",
			"廉贞化禄附注",
			"廉贞化禄",
		},
	}

	got := buildFlyingStarAnalysisFromChart(chart)
	if len(got.HuaLu) != 1 || got.HuaLu[0].TransformedStar != "廉贞" || got.HuaLu[0].TargetPalace != "命宫" {
		t.Fatalf("flying-star analysis accepted malformed labels: %+v", got)
	}
	if len(got.HuaQuan) != 0 || len(got.HuaKe) != 0 || len(got.HuaJi) != 0 {
		t.Fatalf("malformed labels created other transformations: %+v", got)
	}
}

func assertFlyingStarAnalysisMatchesPublishedFourHua(t *testing.T, chart *ZiWeiChart, analysis *FlyingStarAnalysis) {
	t.Helper()
	if analysis == nil {
		t.Fatal("flying-star analysis is nil")
	}
	assertSihuaProjectionSemantics(t, analysis.SihuaProjectionSemantics)
	if analysis.AnalysisKind != "natal_year_stem_four_hua_projection" {
		t.Errorf("analysis_kind = %q, want natal year-stem projection", analysis.AnalysisKind)
	}

	groups := map[string][]FlyTarget{
		"化禄": analysis.HuaLu,
		"化权": analysis.HuaQuan,
		"化科": analysis.HuaKe,
		"化忌": analysis.HuaJi,
	}
	for huaType, items := range groups {
		if len(items) != 1 {
			t.Fatalf("%s targets = %+v, want exactly one natal transformation", huaType, items)
		}
	}

	for _, palace := range chart.Palaces {
		for _, transformed := range palace.FourHua {
			star, huaType, ok := parsePublishedFourHuaLabelForTest(transformed)
			if !ok {
				t.Fatalf("invalid published four_hua label %q", transformed)
			}
			items := groups[huaType]
			if items[0].TransformedStar != star || items[0].HuaType != huaType || items[0].TargetPalace != palace.Name {
				t.Errorf("%s target = %+v, want star=%s palace=%s", huaType, items[0], star, palace.Name)
			}
			assertSihuaProjectionSemantics(t, items[0].SihuaProjectionSemantics)
		}
	}
}

func parsePublishedFourHuaLabelForTest(label string) (string, string, bool) {
	for _, huaType := range SiHuaLabels {
		if len(label) > len(huaType) && label[len(label)-len(huaType):] == huaType {
			return label[:len(label)-len(huaType)], huaType, true
		}
	}
	return "", "", false
}
