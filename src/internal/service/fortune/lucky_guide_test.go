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
	)

	if guide == nil {
		t.Fatal("guide should not be nil")
	}
	if guide.PrimaryElement != "木" || guide.SecondaryElement != "金" || guide.AvoidElement != "火" {
		t.Fatalf("unexpected elements: %+v", guide)
	}
	if guide.Confidence < 70 {
		t.Fatalf("expected guide confidence >= 70, got %d", guide.Confidence)
	}
	if !strings.Contains(guide.Analysis, "流日甲子") || !strings.Contains(guide.Analysis, "六冲") {
		t.Fatalf("analysis should mention flow day and clash: %s", guide.Analysis)
	}
	if len(guide.Cautions) < 2 || !containsGuideValue(guide.Cautions, "避免硬碰硬") {
		t.Fatalf("expected clash caution, got %+v", guide.Cautions)
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

	guide := BuildFortuneGuide(chart, model.Pillar{Gan: "癸", Zhi: "卯"}, 45, "woSheng", "neutral", "", nil, "", nil, nil, nil, nil)
	if guide == nil {
		t.Fatal("guide should not be nil")
	}
	if guide.PrimaryElement == "" || guide.Strategy == "" || guide.Analysis == "" {
		t.Fatalf("guide should contain usable fallback content: %+v", guide)
	}
	if len(guide.LuckyColors) == 0 || len(guide.LuckyNumbers) == 0 {
		t.Fatalf("guide should include color and number fallbacks: %+v", guide)
	}
}

func containsGuideValue(items []model.FortuneGuideItem, value string) bool {
	for _, item := range items {
		if item.Value == value {
			return true
		}
	}
	return false
}
