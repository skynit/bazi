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
	ming := svc.GetPalaceReading(chart, mingIdx)
	if ming == nil {
		t.Fatal("命宫 reading is nil")
	}
	if !strings.Contains(ming.Summary, "廉贞") || !strings.Contains(ming.Summary, "破军") {
		t.Fatalf("命宫 summary should include 廉贞、破军, got %q", ming.Summary)
	}
	if readingHasPattern(ming, "廉贞破军同宫") {
		t.Fatalf("命宫不得把无固定来源的主星对发布成格局: %#v", ming.PatternDetails)
	}
	if readingHasPattern(ming, "紫府同宫") {
		t.Fatalf("命宫 should not include false 紫府同宫 detail: %#v", ming.PatternDetails)
	}
	if !readingEvidenceContains(ming, "four_hua", "破军化禄") {
		t.Fatalf("命宫 evidence should include 破军化禄, got %#v", ming.Evidence)
	}

	caibo := svc.GetPalaceReading(chart, palaceIndexByNameForReadingTest(chart, "财帛"))
	for _, want := range []string{"紫微", "七杀", "天马", "铃星"} {
		if !readingEvidenceContains(caibo, "", want) {
			t.Fatalf("财帛 evidence should include %s, got %#v", want, caibo.Evidence)
		}
	}

	shiye := svc.GetPalaceReading(chart, palaceIndexByNameForReadingTest(chart, "事业"))
	for _, want := range []string{"武曲", "贪狼", "擎羊", "贪狼化忌"} {
		if !readingEvidenceContains(shiye, "", want) {
			t.Fatalf("事业 evidence should include %s, got %#v", want, shiye.Evidence)
		}
	}
	if !strings.Contains(shiye.SihuaInfluence, "贪狼化忌") || !strings.Contains(shiye.SihuaInfluence, "仅记录本命四化标签") ||
		strings.Contains(shiye.SihuaInfluence, "阻滞") {
		t.Fatalf("事业 sihua influence should remain structural, got %q", shiye.SihuaInfluence)
	}

	fude := svc.GetPalaceReading(chart, palaceIndexByNameForReadingTest(chart, "福德"))
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

	reading := svc.GetPalaceReading(chart, palaceIndexByNameForReadingTest(chart, "夫妻"))
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
	if len(reading.ReviewNotes) == 0 || len(reading.Limitations) == 0 {
		t.Fatalf("empty palace should include review notes and limitations, notes=%#v limitations=%#v", reading.ReviewNotes, reading.Limitations)
	}
	if reading.EvidenceBasis != "mixed_deterministic_projection_and_unadjudicated_traditional_labels" ||
		reading.PlacementBasis != "deterministic_rule_projection" ||
		reading.InterpretationBasis != "traditional_rule_labels" ||
		reading.InterpretationStatus != "not_adjudicated" ||
		reading.ValidationStatus != "not_adjudicated" || reading.IsOutcomeConclusion {
		t.Fatalf("reading collapsed evidence boundary: %+v", reading)
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

	reading := svc.GetPalaceReading(&cached, palaceIndexByNameForReadingTest(&cached, "命宫"))
	if reading == nil {
		t.Fatal("cached reading is nil")
	}
	if !strings.Contains(reading.MainStarAnalysis, "廉贞") || !strings.Contains(reading.MainStarAnalysis, "破军") {
		t.Fatalf("cached reading should recover main stars from public Stars field, got %q", reading.MainStarAnalysis)
	}
	if readingHasPattern(reading, "廉贞破军同宫") {
		t.Fatalf("cached reading不得恢复已撤下的无来源格局: %#v", reading.PatternDetails)
	}
}

func TestPalaceReadingRequiresAuthenticatedNatalChart(t *testing.T) {
	svc := NewZiWeiService()
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	if svc.GetPalaceReading(chart, -1) != nil || svc.GetPalaceReading(chart, len(chart.Palaces)) != nil {
		t.Fatal("palace reading accepted an invalid palace index")
	}
	derived := svc.CalculateLiunian(chart, 2026)
	if derived == nil {
		t.Fatal("failed to build valid derived chart fixture")
	}
	if svc.GetPalaceReading(derived, 0) != nil {
		t.Fatal("palace reading accepted a derived chart instead of an authenticated natal chart")
	}
}

func TestPatternDetailsKeepLifeSanfangPatternsVisibleFromLifePalace(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[4].Stars = []StarOutput{{Name: "七杀", Type: "major"}}
	chart.Palaces[8].Stars = []StarOutput{{Name: "破军", Type: "major"}}
	chart.Palaces[6].Stars = []StarOutput{{Name: "贪狼", Type: "major"}}
	chart.Patterns = DetectLocalPatterns(chart)

	details := buildPatternDetailsForPalace(chart, 0)
	if !patternDetailsContain(details, "杀破狼格") {
		t.Fatalf("空命宫未保留命宫三方四正的杀破狼格依据: %#v", details)
	}
}

func TestPatternDetailsUseSameWuTanAndLingTanRelationsAsDetectors(t *testing.T) {
	chart := chartWithLifePalaceAtZi()
	chart.Palaces[0].Branch = "丑"
	chart.Palaces[0].Stars = []StarOutput{
		{Name: "武曲", Type: "major"},
		{Name: "贪狼", Type: "major"},
		{Name: "铃星", Type: "tough"},
	}
	chart.Patterns = DetectLocalPatterns(chart)

	details := buildPatternDetailsForPalace(chart, 0)
	for _, want := range []string{"武贪格", "铃贪格"} {
		if !patternDetailsContain(details, want) {
			t.Fatalf("%s检测成功但命宫解读缺少同口径依据: patterns=%v details=%#v", want, chart.Patterns, details)
		}
	}
}

func TestPublishedPatternsInOrdinarySampleHavePalaceEvidence(t *testing.T) {
	svc := NewZiWeiService()
	missing := map[string]int{}
	for year := 1984; year <= 1993; year++ {
		for month := 1; month <= 12; month++ {
			day := 5 + (year+month)%20
			hour := (year*3 + month*2) % 24
			gender := "男"
			if (year+month)%2 != 0 {
				gender = "女"
			}
			chart, err := svc.CalculateChart(year, month, day, hour, 0, gender)
			if err != nil {
				t.Fatalf("CalculateChart(%d-%02d-%02d %02d:00) failed: %v", year, month, day, hour, err)
			}

			explained := map[string]bool{}
			for palaceIdx := range chart.Palaces {
				for _, detail := range buildPatternDetailsForPalace(chart, palaceIdx) {
					explained[detail.Name] = detail.Basis != ""
				}
			}
			for _, pattern := range chart.Patterns {
				if !explained[pattern] {
					missing[pattern]++
				}
			}
		}
	}
	if len(missing) != 0 {
		t.Fatalf("普通命盘中存在已发布但无宫位结构依据的格局: %v", missing)
	}
}

func patternDetailsContain(details []ReadingPatternDetail, name string) bool {
	for _, detail := range details {
		if detail.Name == name {
			return true
		}
	}
	return false
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
		if value == "" || strings.Contains(item.Value, value) || strings.Contains(item.Basis, value) {
			return true
		}
	}
	return false
}
