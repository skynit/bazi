package fortune

import (
	"strings"
	"testing"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
)

func TestBuildFortuneGuideUsesEffectiveFavor(t *testing.T) {
	chart := &bazipkg.BaziResult{
		DayPillar: model.Pillar{Gan: "戊", Zhi: "午"},
		BodyStrength: bazipkg.BodyStrengthResult{
			Like:    []string{"木", "金"},
			Dislike: []string{"火", "土"},
			Verdict: "身旺",
		},
	}

	guide := BuildFortuneGuide(
		chart,
		model.Pillar{Gan: "甲", Zhi: "子"},
		72,
		"keWo",
		"clash",
		"绿色系",
		[]int{3, 8},
		"东北",
		[]string{"辰时", "申时"},
		[]model.YiJiItem{{Activity: "汇报", Reason: "官杀当值，宜按流程担责"}},
		[]model.YiJiItem{{Activity: "动土", Reason: "冲煞较重，不宜动土"}},
		&RikuyoResult{OverallVerdict: "日课偏吉，宜稳中推进。", FavorScore: 72},
		85,
	)

	if guide == nil {
		t.Fatal("guide should not be nil")
	}
	if guide.PrimaryElement != "木" || guide.SecondaryElement != "金" || guide.AvoidElement != "火" {
		t.Fatalf("unexpected elements: %+v", guide)
	}
	if guide.EvidenceCompleteness < 70 {
		t.Fatalf("expected guide evidence completeness >= 70, got %d", guide.EvidenceCompleteness)
	}
	if !strings.Contains(guide.Analysis, "流日甲子") || !strings.Contains(guide.Analysis, "六冲") {
		t.Fatalf("analysis should mention flow day and clash: %s", guide.Analysis)
	}
	if len(guide.Cautions) < 2 || !containsGuideValue(guide.Cautions, "避免硬碰硬") {
		t.Fatalf("expected clash caution, got %+v", guide.Cautions)
	}
	if len(guide.RecommendedActions) < 5 {
		t.Fatalf("expected richer action list, got %+v", guide.RecommendedActions)
	}
	if guide.RecommendedActions[0].Priority == 0 || guide.RecommendedActions[0].Method == "" || guide.RecommendedActions[0].Timing == "" {
		t.Fatalf("expected detailed action metadata, got %+v", guide.RecommendedActions[0])
	}
	if clash := findGuideValue(guide.Cautions, "避免硬碰硬"); clash == nil || clash.Intensity != "高" || clash.Source == "" || clash.Method == "" {
		t.Fatalf("expected detailed clash caution, got %+v", clash)
	}
}

func TestBuildFortuneGuidePrimaryChangesWithFlowDay(t *testing.T) {
	chart := &bazipkg.BaziResult{
		DayPillar: model.Pillar{Gan: "戊", Zhi: "午"},
		BodyStrength: bazipkg.BodyStrengthResult{
			Like:    []string{"水", "木"},
			Dislike: []string{"金", "火"},
			Verdict: "身旺",
		},
	}

	woodDay := BuildFortuneGuide(chart, model.Pillar{Gan: "甲", Zhi: "寅"}, 70, "neutral", "neutral", "", nil, "", nil, nil, nil, nil, 80)
	waterDay := BuildFortuneGuide(chart, model.Pillar{Gan: "壬", Zhi: "子"}, 70, "neutral", "neutral", "", nil, "", nil, nil, nil, nil, 80)
	metalDay := BuildFortuneGuide(chart, model.Pillar{Gan: "庚", Zhi: "申"}, 70, "neutral", "neutral", "", nil, "", nil, nil, nil, nil, 80)

	if woodDay.PrimaryElement != "木" || woodDay.SecondaryElement != "水" {
		t.Fatalf("wood flow day should use wood as primary and water as support: %+v", woodDay)
	}
	if waterDay.PrimaryElement != "水" || waterDay.SecondaryElement != "木" {
		t.Fatalf("water flow day should use water as primary and wood as support: %+v", waterDay)
	}
	if woodDay.PrimaryElement == waterDay.PrimaryElement {
		t.Fatalf("primary element should change with the flow day: wood=%s water=%s", woodDay.PrimaryElement, waterDay.PrimaryElement)
	}
	if len(waterDay.LuckyColors) == 0 || waterDay.LuckyColors[0].Element != waterDay.PrimaryElement {
		t.Fatalf("first lucky color should follow the dynamic primary element: %+v", waterDay.LuckyColors)
	}
	if metalDay.PrimaryElement != "水" || metalDay.AvoidElement != "金" {
		t.Fatalf("disliked metal flow day should be drained by favorable water and remain the avoid element: %+v", metalDay)
	}
}

func TestSelectDailyBlessingElementsControlsDislikedFlow(t *testing.T) {
	primary, secondary, avoid := selectDailyBlessingElements("土", []string{"木", "水"}, []string{"土", "火"}, "金")
	if primary != "木" || secondary != "水" || avoid != "土" {
		t.Fatalf("expected wood to control disliked earth flow, got primary=%s secondary=%s avoid=%s", primary, secondary, avoid)
	}
}

func TestBuildFortuneGuideFallbackStillExplains(t *testing.T) {
	chart := &bazipkg.BaziResult{
		DayPillar: model.Pillar{Gan: "辛", Zhi: "酉"},
		BodyStrength: bazipkg.BodyStrengthResult{
			Like:    nil,
			Dislike: nil,
			Verdict: "平衡",
		},
	}

	guide := BuildFortuneGuide(chart, model.Pillar{Gan: "癸", Zhi: "卯"}, 45, "woSheng", "neutral", "", nil, "", nil, nil, nil, nil, 55)
	if guide == nil {
		t.Fatal("guide should not be nil")
	}
	if guide.PrimaryElement == "" || guide.Strategy == "" || guide.Analysis == "" {
		t.Fatalf("guide should contain usable fallback content: %+v", guide)
	}
	if len(guide.LuckyColors) == 0 || len(guide.LuckyNumbers) == 0 {
		t.Fatalf("guide should include color and number fallbacks: %+v", guide)
	}
	if len(guide.RecommendedActions) == 0 || guide.RecommendedActions[0].Category == "" {
		t.Fatalf("fallback guide should include detailed action metadata: %+v", guide.RecommendedActions)
	}
}

func containsGuideValue(items []model.FortuneGuideItem, value string) bool {
	return findGuideValue(items, value) != nil
}

func findGuideValue(items []model.FortuneGuideItem, value string) *model.FortuneGuideItem {
	for _, item := range items {
		if item.Value == value {
			return &item
		}
	}
	return nil
}
