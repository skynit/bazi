package ziwei

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestPalaceReading_UserSampleStructuredDetails(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	mingIdx := palaceIndexByNameForReadingTest(chart, "命宫")
	ming := GetPalaceReading(chart, mingIdx)
	if ming == nil {
		t.Fatal("命宫 reading is nil")
	}
	if !strings.Contains(ming.Summary, "廉贞") || !strings.Contains(ming.Summary, "破军") {
		t.Fatalf("命宫 summary should include 廉贞、破军, got %q", ming.Summary)
	}
	if !readingHasPattern(ming, "廉贞破军同宫") {
		t.Fatalf("命宫 pattern_details should include 廉贞破军同宫, got %#v", ming.PatternDetails)
	}
	if readingHasPattern(ming, "紫府同宫") {
		t.Fatalf("命宫 should not include false 紫府同宫 detail: %#v", ming.PatternDetails)
	}
	if !readingEvidenceContains(ming, "four_hua", "破军化禄") {
		t.Fatalf("命宫 evidence should include 破军化禄, got %#v", ming.Evidence)
	}

	caibo := GetPalaceReading(chart, palaceIndexByNameForReadingTest(chart, "财帛"))
	for _, want := range []string{"紫微", "七杀", "天马", "铃星"} {
		if !readingEvidenceContains(caibo, "", want) {
			t.Fatalf("财帛 evidence should include %s, got %#v", want, caibo.Evidence)
		}
	}

	shiye := GetPalaceReading(chart, palaceIndexByNameForReadingTest(chart, "事业"))
	for _, want := range []string{"武曲", "贪狼", "擎羊", "贪狼化忌"} {
		if !readingEvidenceContains(shiye, "", want) {
			t.Fatalf("事业 evidence should include %s, got %#v", want, shiye.Evidence)
		}
	}
	if !strings.Contains(shiye.SihuaInfluence, "贪狼化忌") || !strings.Contains(shiye.SihuaInfluence, "阻滞") {
		t.Fatalf("事业 sihua influence should explain 贪狼化忌 as 化忌, got %q", shiye.SihuaInfluence)
	}

	fude := GetPalaceReading(chart, palaceIndexByNameForReadingTest(chart, "福德"))
	for _, want := range []string{"天府", "文曲", "陀罗"} {
		if !readingEvidenceContains(fude, "", want) {
			t.Fatalf("福德 evidence should include %s, got %#v", want, fude.Evidence)
		}
	}
	if !readingEvidenceContains(fude, "body_palace", "福德") || !strings.Contains(fude.Summary, "身宫") {
		t.Fatalf("福德 should be marked as body palace, summary=%q evidence=%#v", fude.Summary, fude.Evidence)
	}
}

func TestPalaceReading_EmptyPalaceBorrowsOppositeAndSanfang(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	reading := GetPalaceReading(chart, palaceIndexByNameForReadingTest(chart, "夫妻"))
	if reading == nil {
		t.Fatal("夫妻 reading is nil")
	}
	if reading.Summary == "" || !strings.Contains(reading.Summary, "空宫") {
		t.Fatalf("empty palace summary should explain 空宫, got %q", reading.Summary)
	}
	if reading.SanfangContext == nil || len(reading.SanfangContext.Notes) == 0 {
		t.Fatalf("empty palace should include sanfang context notes, got %#v", reading.SanfangContext)
	}
	if !readingEvidenceContains(reading, "borrowed_star", "") {
		t.Fatalf("empty palace should include borrowed_star evidence, got %#v", reading.Evidence)
	}
	if len(reading.Advice) == 0 || len(reading.RiskFlags) == 0 {
		t.Fatalf("empty palace should include advice and risk flags, advice=%#v risks=%#v", reading.Advice, reading.RiskFlags)
	}
}

func TestPalaceReading_CachedJSONStillHasStructuredDetails(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}
	raw, err := json.Marshal(chart)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	var cached ZiWeiChart
	if err := json.Unmarshal(raw, &cached); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	reading := GetPalaceReading(&cached, palaceIndexByNameForReadingTest(&cached, "命宫"))
	if reading == nil {
		t.Fatal("cached reading is nil")
	}
	if !strings.Contains(reading.MainStarAnalysis, "廉贞") || !strings.Contains(reading.MainStarAnalysis, "破军") {
		t.Fatalf("cached reading should recover main stars from public Stars field, got %q", reading.MainStarAnalysis)
	}
	if !readingHasPattern(reading, "廉贞破军同宫") {
		t.Fatalf("cached reading should keep pattern detail, got %#v", reading.PatternDetails)
	}
}

func palaceIndexByNameForReadingTest(chart *ZiWeiChart, name string) int {
	for i, p := range chart.Palaces {
		if p.Name == name {
			return i
		}
	}
	return -1
}

func readingHasPattern(reading *PalaceReading, name string) bool {
	if reading == nil {
		return false
	}
	for _, detail := range reading.PatternDetails {
		if detail.Name == name && detail.Basis != "" {
			return true
		}
	}
	return false
}

func readingEvidenceContains(reading *PalaceReading, evidenceType, value string) bool {
	if reading == nil {
		return false
	}
	for _, item := range reading.Evidence {
		if evidenceType != "" && item.Type != evidenceType {
			continue
		}
		if value == "" || strings.Contains(item.Value, value) || strings.Contains(item.Impact, value) {
			return true
		}
	}
	return false
}
