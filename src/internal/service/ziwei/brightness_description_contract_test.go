package ziwei

import (
	"strings"
	"testing"
)

func TestGetStarBrightnessAcceptsOnlyKnownStarLevelPairs(t *testing.T) {
	for _, table := range []map[string]map[string]string{STAR_BRIGHTNESS, STAR_BRIGHTNESS_AUX} {
		for star, levels := range table {
			for _, level := range brightnessLevels {
				want := levels[level]
				got, ok := GetStarBrightness(star, level)
				if !ok || got != want || got == "" {
					t.Fatalf("brightness description %s/%s = %q/%t, want %q", star, level, got, ok, want)
				}
			}
		}
	}

	for _, input := range [][2]string{
		{"", "平"}, {"未知星", "平"}, {"紫微", ""}, {"紫微", "未知亮度"}, {"左辅", "平"},
	} {
		if got, ok := GetStarBrightness(input[0], input[1]); ok || got != "" {
			t.Fatalf("unknown brightness pair %q/%q returned %q/%t", input[0], input[1], got, ok)
		}
	}
}

func TestPalaceReadingBuildersDoNotInventNeutralBrightness(t *testing.T) {
	palace := PalaceInfo{
		Name:  "命宫",
		Stars: []StarOutput{{Name: "紫微", Type: "major"}},
	}
	mainStars := palaceMainStars(palace)
	analysis := buildMainStarAnalysis(palace, mainStars)
	summary := buildBrightnessSummary(palace, mainStars)
	basis := mainStarBasis("紫微", "")

	for label, text := range map[string]string{
		"analysis": analysis,
		"summary":  summary,
		"basis":    basis,
	} {
		if strings.Contains(text, "(平)") || strings.Contains(text, "按平级") || strings.Contains(text, "亮度等级为平") {
			t.Fatalf("%s invented neutral brightness: %q", label, text)
		}
	}
	if analysis != "命宫主星为紫微。" || summary != "紫微" || !strings.Contains(basis, "亮度未提供") {
		t.Fatalf("missing-brightness semantics = analysis:%q summary:%q basis:%q", analysis, summary, basis)
	}
}

func TestPublishedChartMainStarBrightnessRemainsFullyDescribed(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}

	mainStarCount := 0
	for palaceIdx, palace := range chart.Palaces {
		for _, star := range palace.Stars {
			if star.Type != "major" {
				continue
			}
			mainStarCount++
			if desc, ok := GetStarBrightness(star.Name, star.Brightness); !ok || desc == "" {
				t.Fatalf("published main star %s/%s has no strict description", star.Name, star.Brightness)
			}
		}
		reading := svc.GetPalaceReading(chart, palaceIdx)
		if reading == nil || strings.Contains(reading.MainStarAnalysis, "亮度未提供") ||
			strings.Contains(reading.Brightness, "亮度未提供") {
			t.Fatalf("published palace %s lost brightness semantics: %+v", palace.Name, reading)
		}
	}
	if mainStarCount != len(StarBrightnessMap) {
		t.Fatalf("published major-star count = %d, want %d", mainStarCount, len(StarBrightnessMap))
	}
}
