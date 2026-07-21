package ziwei

import (
	"encoding/json"
	"strings"
	"testing"
)

var ziweiMedicalTerms = []string{
	"肝胆", "脾胃", "心脑血管", "肺呼吸道", "肾膀胱", "泌尿系统",
	"疾病倾向", "心理健康", "身心健康", "主动体检", "手术", "寿命较长", "主长寿",
}

var ziweiOutcomeClaimTerms = []string{
	"大富大贵", "大富大貴", "衣食无忧", "衣食無憂", "财运", "財運",
	"破财", "破財", "适合投资", "適合投資", "不宜投资", "不宜投資",
	"借贷", "借貸", "事业有成", "事業有成", "步步高升", "婚期", "早婚之命",
	"非常匹配", "较好匹配", "較好匹配", "适合共同创业", "適合共同創業",
	"适合", "適合", "不宜", "可把握", "宜在",
}

func TestPeriodInterpretationOmitsMedicalInference(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	liunian := svc.CalculateLiunian(base, 2026)
	liuyue := svc.CalculateLiuyueForDate(base, 2026, 7, 15)
	liuri := svc.CalculateLiuriForDate(base, 2026, 7, 15)
	interpreter := NewPeriodInterpreterFromChart(base)
	payload, err := json.Marshal(map[string]any{
		"liunian": interpreter.AnalyzeLiunian(liunian, 2026),
		"liuyue":  interpreter.AnalyzeLiuyue(liuyue, 2026, 7, 15),
		"liuri":   interpreter.AnalyzeLiuri(liuri, 2026, 7, 15),
		"summary": interpreter.SummarizeAll(liunian, liuyue, liuri, 2026, 7, 15),
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	if strings.Contains(text, `"health"`) {
		t.Fatalf("period interpretation leaked a health field: %s", payload)
	}
	if strings.Contains(text, `"emotional_state"`) {
		t.Fatalf("period interpretation leaked an inferred emotional state: %s", payload)
	}
	if strings.Contains(text, `"advice"`) || !strings.Contains(text, `"review_notes"`) {
		t.Fatalf("period summary exposed advice instead of review notes: %s", payload)
	}
	if strings.Count(text, `"evidence_basis":"deterministic_rule_projection"`) < 4 || strings.Count(text, `"validation_status":"not_adjudicated"`) < 4 {
		t.Fatalf("period interpretation omitted deterministic validation metadata: %s", payload)
	}
	assertNoZiWeiMedicalTerms(t, text)
	assertNoZiWeiOutcomeClaims(t, text)
}

func TestZiWeiTemplatesOmitMedicalAndLifespanClaims(t *testing.T) {
	payload, err := json.Marshal([]any{
		buildMainStarTemplates(),
		buildAuxStarTemplates(),
		buildFourHuaTemplates(),
		buildPatternTemplates(),
		buildPalaceContext(),
	})
	if err != nil {
		t.Fatal(err)
	}
	assertNoZiWeiMedicalTerms(t, string(payload))
	assertNoZiWeiOutcomeClaims(t, string(payload))
}

func TestRuntimeInterpretationsOmitOutcomeClaimsAndConfidence(t *testing.T) {
	svc := NewZiWeiService()
	chartA, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	chartB, err := svc.CalculateChart(1992, 9, 8, 9, 0, "女")
	if err != nil {
		t.Fatal(err)
	}
	liunian := svc.CalculateLiunian(chartA, 2026)
	readings := make([]*PalaceReading, 0, len(chartA.Palaces))
	for i := range chartA.Palaces {
		readings = append(readings, svc.GetPalaceReading(chartA, i))
	}
	payload, err := json.Marshal(map[string]any{
		"readings":         readings,
		"self_mutagens":    svc.AnalyzeSelfMutagen(chartA),
		"overlay":          svc.AnalyzeLiunianOverlay(chartA, liunian, 2026),
		"heming":           svc.AnalyzeHeming(chartA, chartB),
		"dayun_analysis":   BuildDayunAnalysis(chartA, nil, 23),
		"liunian_analysis": BuildLiunianAnalysis(chartA, liunian, 2026),
	})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	assertNoZiWeiMedicalTerms(t, text)
	assertNoZiWeiOutcomeClaims(t, text)
	if strings.Contains(text, `"confidence"`) || strings.Contains(text, `"marriage_timing"`) ||
		strings.Contains(text, `"evidence_completeness"`) || strings.Contains(text, `"impact"`) ||
		strings.Contains(text, `"evidence_basis":"empirical"`) {
		t.Fatalf("runtime interpretation exposed prediction-like fields: %s", payload)
	}
	if !strings.Contains(text, `"evidence_basis":"mixed_deterministic_projection_and_unadjudicated_traditional_labels"`) ||
		!strings.Contains(text, `"placement_basis":"deterministic_rule_projection"`) ||
		!strings.Contains(text, `"interpretation_basis":"traditional_rule_labels"`) ||
		!strings.Contains(text, `"validation_status":"not_adjudicated"`) ||
		!strings.Contains(text, `"is_outcome_conclusion":false`) {
		t.Fatalf("runtime interpretation omitted validation boundary: %s", payload)
	}
}

func TestTianLiangBrightnessIsNotPublishedAsStandalonePattern(t *testing.T) {
	chart := &ZiWeiChart{}
	chart.Palaces[0] = PalaceInfo{
		Name: "命宫",
		Stars: []StarOutput{{Name: "天梁", Type: "major", Scope: "origin", Brightness: "庙"}},
	}
	for _, pattern := range DetectLocalPatterns(chart) {
		if pattern == "天梁庙旺" {
			t.Fatal("天梁亮度属性仍被误发布为整盘独立格局")
		}
	}
	if _, exists := buildPatternTemplates()["天梁庙旺"]; exists {
		t.Fatal("已撤下的天梁庙旺独立格局仍保留解释模板")
	}
}

func assertNoZiWeiMedicalTerms(t *testing.T, text string) {
	t.Helper()
	for _, term := range ziweiMedicalTerms {
		if strings.Contains(text, term) {
			t.Fatalf("ziwei interpretation inferred %q: %s", term, text)
		}
	}
}

func assertNoZiWeiOutcomeClaims(t *testing.T, text string) {
	t.Helper()
	for _, term := range ziweiOutcomeClaimTerms {
		if strings.Contains(text, term) {
			t.Fatalf("ziwei interpretation inferred outcome %q: %s", term, text)
		}
	}
}
