package ziwei_test

import (
	. "bazi/internal/service/ziwei"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// B6: 格局检测测试
//
// Constructs charts with specific star placements to trigger patterns:
//   - 紫府同宫: 紫微+天府 in same palace
//   - 杀破狼: 七杀/破军/贪狼 (≥2) in same palace or trine
//   - 日月同宫 (日月拱照): 太阳 ↔ 太阴 in trine/opposition
//   - Negative assertions: charts that should NOT match
// ═══════════════════════════════════════════════════════════════════════

// makeChart creates a minimal ZiWeiChart for pattern testing.
// Only Palaces[i].MainStars is populated; other fields may be zeroed.
func makeChart(starsByPalace [12][]string) *ZiWeiChart {
	chart := &ZiWeiChart{}
	for i := 0; i < 12; i++ {
		chart.Palaces[i] = PalaceInfo{
			Name:      ZIWEI_PALACE_NAMES[i],
			Branch:    BranchNames[i],
			MainStars: starsByPalace[i],
		}
	}
	return chart
}

// makeChartWithAux creates a chart with both main and aux stars set.
func makeChartWithAux(mainStars, auxStars [12][]string) *ZiWeiChart {
	chart := &ZiWeiChart{}
	for i := 0; i < 12; i++ {
		chart.Palaces[i] = PalaceInfo{
			Name:      ZIWEI_PALACE_NAMES[i],
			Branch:    BranchNames[i],
			MainStars: mainStars[i],
			AuxStars:  auxStars[i],
		}
	}
	return chart
}

// ──────────────────── 紫府同宫 ────────────────────

func TestPattern_ZiFuTongGong(t *testing.T) {
	t.Run("紫府同在子宫命中", func(t *testing.T) {
		// 紫微 and 天府 in 子(0) = 命宫
		stars := [12][]string{}
		stars[0] = []string{"紫微", "天府"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "紫府同宫") {
			t.Errorf("期望检测到[紫府同宫], 实际=%v", result)
		}
	})

	t.Run("紫府同在午宫财帛", func(t *testing.T) {
		// 紫微 and 天府 in 午(6) = 财帛宫(index 4)
		stars := [12][]string{}
		stars[6] = []string{"紫微", "天府"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "紫府同宫") {
			t.Errorf("期望检测到[紫府同宫], 实际=%v", result)
		}
	})

	t.Run("紫府分开不同宫_不匹配", func(t *testing.T) {
		// 紫微在子(0), 天府在午(6) — 不在同一宫
		stars := [12][]string{}
		stars[0] = []string{"紫微"}
		stars[6] = []string{"天府"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if containsPattern(result, "紫府同宫") {
			t.Errorf("不应检测到[紫府同宫], 实际=%v", result)
		}
	})

	t.Run("只有紫微没有天府_不匹配", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"紫微"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if containsPattern(result, "紫府同宫") {
			t.Errorf("不应检测到[紫府同宫], 实际=%v", result)
		}
		// Still should NOT match 杀破狼
		if containsPattern(result, "杀破狼格") {
			t.Errorf("不应检测到杀破狼格(只有紫微)")
		}
	})
}

// ──────────────────── 杀破狼格 ────────────────────

func TestPattern_ShaPoLang(t *testing.T) {
	t.Run("七杀破军同宫", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"七杀", "破军"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "杀破狼格") {
			t.Errorf("期望检测到[杀破狼格], 实际=%v", result)
		}
	})

	t.Run("七杀贪狼同宫", func(t *testing.T) {
		stars := [12][]string{}
		stars[2] = []string{"七杀", "贪狼"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "杀破狼格") {
			t.Errorf("期望检测到[杀破狼格], 实际=%v", result)
		}
	})

	t.Run("破军贪狼同宫", func(t *testing.T) {
		stars := [12][]string{}
		stars[5] = []string{"破军", "贪狼"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "杀破狼格") {
			t.Errorf("期望检测到[杀破狼格], 实际=%v", result)
		}
	})

	t.Run("七杀破军贪狼三合四正", func(t *testing.T) {
		// 七杀在子(0), 破军在辰(4) — trine relationship
		stars := [12][]string{}
		stars[0] = []string{"七杀"}
		stars[4] = []string{"破军"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "杀破狼格") {
			t.Errorf("期望检测到[杀破狼格](七杀在子,破军在辰为四正关系), 实际=%v", result)
		}
	})

	t.Run("只有七杀_不匹配", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"七杀"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if containsPattern(result, "杀破狼格") {
			t.Errorf("不应检测到杀破狼格(只有七杀)")
		}
	})

	t.Run("无关星曜_不匹配", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"紫微", "天府"}
		stars[1] = []string{"太阳", "太阴"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if containsPattern(result, "杀破狼格") {
			t.Errorf("不应检测到杀破狼格(紫府日月)")
		}
	})
}

// ──────────────────── 日月拱照 ────────────────────

func TestPattern_RiYueGongZhao(t *testing.T) {
	t.Run("太阳在午_太阴在戌(对宫)", func(t *testing.T) {
		// 太阳在午(6), 太阴在子(0) — opposite palace
		stars := [12][]string{}
		stars[6] = []string{"太阳"}
		stars[0] = []string{"太阴"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "日月拱照") {
			t.Errorf("期望检测到[日月拱照](太阳午-太阴子对宫), 实际=%v", result)
		}
	})

	t.Run("太阳在子_太阴在辰(三合)", func(t *testing.T) {
		// 太阳在子(0), 太阴在辰(4) — trine
		stars := [12][]string{}
		stars[0] = []string{"太阳"}
		stars[4] = []string{"太阴"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "日月拱照") {
			t.Errorf("期望检测到[日月拱照](太阳子-太阴辰三合), 实际=%v", result)
		}
	})

	t.Run("太阳太阴同宫_不匹配日月拱照", func(t *testing.T) {
		// 同宫不是拱照
		stars := [12][]string{}
		stars[0] = []string{"太阳", "太阴"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if containsPattern(result, "日月拱照") {
			t.Errorf("不应检测到日月拱照(太阳太阴同宫)")
		}
	})

	t.Run("只有太阳没有太阴_不匹配", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"太阳"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if containsPattern(result, "日月拱照") {
			t.Errorf("不应检测到日月拱照(只有太阳)")
		}
	})
}

// ──────────────────── End-to-End Pattern Detection ────────────────────

// TestPatterns_EndToEnd verifies patterns through full CalculateChart.
func TestPatterns_EndToEnd(t *testing.T) {
	svc := NewZiWeiService()

	// Test multiple charts and verify DetectLocalPatterns works
	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// Just verify the service method works
	patterns := svc.DetectLocalPatterns(chart)
	if patterns == nil {
		t.Fatal("DetectLocalPatterns returned nil")
	}

	t.Logf("检测到的格局: %v", patterns)

	// Also verify the standalone function
	patterns2 := DetectLocalPatterns(chart)
	if len(patterns) != len(patterns2) {
		t.Errorf("DetectLocalPatterns via service (%d) != standalone (%d)",
			len(patterns), len(patterns2))
	}

	if containsPattern(patterns, "紫府同宫") {
		t.Errorf("2003-04-15 14:00 紫微在巳、天府在亥，不应误判为紫府同宫: %v", patterns)
	}
	if !containsPattern(patterns, "廉贞破军同宫") {
		t.Errorf("2003-04-15 14:00 命宫廉贞破军同宫，应检测到廉贞破军同宫: %v", patterns)
	}
}

// ──────────────────── Negative Assertion Test ────────────────────

func TestPattern_NegativeAssertions(t *testing.T) {
	t.Run("空宫_不匹配任何特殊格局", func(t *testing.T) {
		stars := [12][]string{}
		// Put some innocuous stars that don't form the tested patterns
		stars[0] = []string{"天机"}
		stars[3] = []string{"太阴"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)

		// These patterns should NOT appear
		negativePatterns := []string{"紫府同宫", "杀破狼格", "日月拱照"}
		for _, p := range negativePatterns {
			if containsPattern(result, p) {
				t.Errorf("不应检测到[%s] (配置: 天机子,太阴卯)", p)
			}
		}
	})

	t.Run("单星分布_不触发多星格局", func(t *testing.T) {
		stars := [12][]string{}
		// Use stars that won't trigger the OR-logic checkers
		stars[4] = []string{"天相"}
		stars[8] = []string{"天梁"}
		chart := makeChart(stars)

		result := DetectLocalPatterns(chart)
		if containsPattern(result, "紫府同宫") {
			t.Errorf("不应检测到紫府同宫(无紫微天府)")
		}
		if containsPattern(result, "杀破狼格") {
			t.Errorf("不应检测到杀破狼格(无七杀破军贪狼)")
		}
	})
}

// ──────────────────── Helpers ────────────────────

func containsPattern(patterns []string, target string) bool {
	for _, p := range patterns {
		if p == target {
			return true
		}
	}
	return false
}
