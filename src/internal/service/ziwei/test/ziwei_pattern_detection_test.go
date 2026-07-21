package ziwei_test

import (
	. "bazi/internal/service/ziwei"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// B6: 格局检测测试
//
// Constructs charts with specific star placements to trigger patterns:
//   - 紫府同宫: 紫微+天府 in 寅/申 life palace
//   - 杀破狼: 七杀/破军/贪狼全部进入命宫三方四正
//   - 日月拱照: 古籍固定的太阳、太阴与命宫组合
//   - Negative assertions: charts that should NOT match
// ═══════════════════════════════════════════════════════════════════════

// makeChart creates a minimal ZiWeiChart for pattern testing.
// Only the public Stars projection is populated; other fields may be zeroed.
func makeChart(starsByPalace [12][]string) *ZiWeiChart {
	chart := &ZiWeiChart{}
	for i := 0; i < 12; i++ {
		chart.Palaces[i] = PalaceInfo{
			Name:   ZIWEI_PALACE_NAMES[i],
			Branch: BranchNames[i],
			Stars:  publishedStarOutputs(starsByPalace[i], nil),
		}
	}
	return chart
}

// makeChartWithAux creates a chart with both main and aux stars set.
func makeChartWithAux(mainStars, auxStars [12][]string) *ZiWeiChart {
	chart := &ZiWeiChart{}
	for i := 0; i < 12; i++ {
		chart.Palaces[i] = PalaceInfo{
			Name:   ZIWEI_PALACE_NAMES[i],
			Branch: BranchNames[i],
			Stars:  publishedStarOutputs(mainStars[i], auxStars[i]),
		}
	}
	return chart
}

// ──────────────────── 紫府同宫 ────────────────────

func TestPattern_ZiFuTongGong(t *testing.T) {
	t.Run("紫府同在寅宫命中", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"紫微", "天府"}
		chart := makeChart(stars)
		chart.Palaces[0].Branch = "寅"
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "紫府同宫") {
			t.Errorf("期望检测到[紫府同宫], 实际=%v", result)
		}
	})

	t.Run("紫府同在非命宫不发布整盘格局", func(t *testing.T) {
		stars := [12][]string{}
		stars[2] = []string{"紫微", "天府"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if containsPattern(result, "紫府同宫") {
			t.Errorf("紫府不在命宫不得发布[紫府同宫], 实际=%v", result)
		}
	})

	t.Run("紫府同坐子宫命宫不匹配", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"紫微", "天府"}
		result := DetectLocalPatterns(makeChart(stars))
		if containsPattern(result, "紫府同宫") {
			t.Errorf("紫府同宫格只取寅申命宫, 实际=%v", result)
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
	t.Run("七杀破军贪狼齐会命宫三方", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"七杀"}
		stars[4] = []string{"破军"}
		stars[8] = []string{"贪狼"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "杀破狼格") {
			t.Errorf("期望检测到三星齐会命宫三方的[杀破狼格], 实际=%v", result)
		}
	})

	for _, tc := range []struct {
		name  string
		stars [12][]string
	}{
		{name: "七杀破军两星同宫不足", stars: [12][]string{0: {"七杀", "破军"}}},
		{name: "七杀贪狼两星同宫不足", stars: [12][]string{2: {"七杀", "贪狼"}}},
		{name: "破军贪狼两星同宫不足", stars: [12][]string{5: {"破军", "贪狼"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result := DetectLocalPatterns(makeChart(tc.stars))
			if containsPattern(result, "杀破狼格") {
				t.Errorf("两星不得误报杀破狼格, 实际=%v", result)
			}
		})
	}

	t.Run("三星齐全但不在命宫三方", func(t *testing.T) {
		stars := [12][]string{}
		stars[1] = []string{"七杀"}
		stars[5] = []string{"破军"}
		stars[9] = []string{"贪狼"}
		result := DetectLocalPatterns(makeChart(stars))
		if containsPattern(result, "杀破狼格") {
			t.Errorf("非命宫三方不得误报杀破狼格, 实际=%v", result)
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

func TestPattern_JiYueTongLiang(t *testing.T) {
	t.Run("四星齐会命宫三方", func(t *testing.T) {
		stars := [12][]string{}
		stars[2] = []string{"天机"}
		stars[6] = []string{"天同"}
		stars[8] = []string{"太阴"}
		stars[10] = []string{"天梁"}
		chart := makeChart(stars)
		chart.Palaces[0].Name = ""
		chart.Palaces[2].Name = "命宫"
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "机月同梁格") {
			t.Errorf("期望检测到四星齐会命宫三方的[机月同梁格], 实际=%v", result)
		}
	})

	t.Run("命宫三方只有三星", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"天机"}
		stars[4] = []string{"太阴"}
		stars[8] = []string{"天梁"}
		result := DetectLocalPatterns(makeChart(stars))
		if containsPattern(result, "机月同梁格") {
			t.Errorf("三星不得误报机月同梁格, 实际=%v", result)
		}
	})

	t.Run("四星齐全但不在命宫三方", func(t *testing.T) {
		stars := [12][]string{}
		stars[1] = []string{"天机"}
		stars[5] = []string{"太阴"}
		stars[7] = []string{"天同"}
		stars[9] = []string{"天梁"}
		result := DetectLocalPatterns(makeChart(stars))
		if containsPattern(result, "机月同梁格") {
			t.Errorf("非命宫三方不得误报机月同梁格, 实际=%v", result)
		}
	})
}

func TestPattern_FuXiangChaoYuan(t *testing.T) {
	t.Run("天府守事业天相守财帛", func(t *testing.T) {
		stars := [12][]string{}
		stars[8] = []string{"天府"}
		stars[4] = []string{"天相"}
		result := DetectLocalPatterns(makeChart(stars))
		if !containsPattern(result, "府相朝垣") {
			t.Errorf("期望检测到天府守事业、天相守财帛的[府相朝垣], 实际=%v", result)
		}
	})

	t.Run("天府坐命天相守财帛", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"天府"}
		stars[4] = []string{"天相"}
		result := DetectLocalPatterns(makeChart(stars))
		if containsPattern(result, "府相朝垣") {
			t.Errorf("天府坐命不符合天府守事业的府相朝垣结构, 实际=%v", result)
		}
	})

	t.Run("府相宫位对调不成朝垣", func(t *testing.T) {
		stars := [12][]string{}
		stars[4] = []string{"天府"}
		stars[8] = []string{"天相"}
		result := DetectLocalPatterns(makeChart(stars))
		if containsPattern(result, "府相朝垣") {
			t.Errorf("府相宫位对调后不得误报府相朝垣, 实际=%v", result)
		}
	})

	t.Run("两星都不在命宫三方", func(t *testing.T) {
		stars := [12][]string{}
		stars[1] = []string{"天府"}
		stars[5] = []string{"天相"}
		result := DetectLocalPatterns(makeChart(stars))
		if containsPattern(result, "府相朝垣") {
			t.Errorf("府相都在命宫三方之外不得误报, 实际=%v", result)
		}
	})
}

func TestPattern_ChangQuTongHui(t *testing.T) {
	t.Run("昌曲同会命宫三方", func(t *testing.T) {
		mainStars := [12][]string{}
		auxStars := [12][]string{}
		auxStars[0] = []string{"文昌"}
		auxStars[4] = []string{"文曲"}
		result := DetectLocalPatterns(makeChartWithAux(mainStars, auxStars))
		if !containsPattern(result, "昌曲同会") {
			t.Errorf("期望检测到昌曲同会命宫三方, 实际=%v", result)
		}
	})

	t.Run("昌曲同坐命宫", func(t *testing.T) {
		mainStars := [12][]string{}
		auxStars := [12][]string{}
		auxStars[0] = []string{"文昌", "文曲"}
		result := DetectLocalPatterns(makeChartWithAux(mainStars, auxStars))
		if !containsPattern(result, "昌曲同会") {
			t.Errorf("昌曲坐命应属于昌曲同会结构, 实际=%v", result)
		}
	})

	t.Run("命宫三方只有文昌", func(t *testing.T) {
		mainStars := [12][]string{}
		auxStars := [12][]string{}
		auxStars[0] = []string{"文昌"}
		result := DetectLocalPatterns(makeChartWithAux(mainStars, auxStars))
		if containsPattern(result, "昌曲同会") {
			t.Errorf("单星不得误报昌曲同会, 实际=%v", result)
		}
	})

	t.Run("昌曲同会其他宫组三方", func(t *testing.T) {
		mainStars := [12][]string{}
		auxStars := [12][]string{}
		auxStars[1] = []string{"文昌"}
		auxStars[5] = []string{"文曲"}
		result := DetectLocalPatterns(makeChartWithAux(mainStars, auxStars))
		if containsPattern(result, "昌曲同会") {
			t.Errorf("非命宫三方不得误报昌曲同会, 实际=%v", result)
		}
	})
}

func TestPattern_QiShaChaoDou(t *testing.T) {
	t.Run("七杀坐子命对宫武府", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"七杀"}
		stars[6] = []string{"武曲", "天府"}
		result := DetectLocalPatterns(makeChart(stars))
		if !containsPattern(result, "七杀朝斗") {
			t.Errorf("期望检测到七杀坐子命、对宫武府的七杀朝斗, 实际=%v", result)
		}
	})

	t.Run("七杀不坐命", func(t *testing.T) {
		stars := [12][]string{}
		stars[1] = []string{"七杀"}
		stars[7] = []string{"武曲", "天府"}
		result := DetectLocalPatterns(makeChart(stars))
		if containsPattern(result, "七杀朝斗") {
			t.Errorf("七杀不坐命不得误报七杀朝斗, 实际=%v", result)
		}
	})

	t.Run("子命对宫武府缺天府", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"七杀"}
		stars[6] = []string{"武曲"}
		result := DetectLocalPatterns(makeChart(stars))
		if containsPattern(result, "七杀朝斗") {
			t.Errorf("对宫武府不全不得误报七杀朝斗, 实际=%v", result)
		}
	})
}

// ──────────────────── 日月拱照 ────────────────────

func TestPattern_RiYueGongZhao(t *testing.T) {
	t.Run("日巳月酉命丑", func(t *testing.T) {
		stars := [12][]string{}
		stars[5] = []string{"太阳"}
		stars[9] = []string{"太阴"}
		chart := makeChart(stars)
		chart.Palaces[0].Name = ""
		chart.Palaces[1].Name = "命宫"
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "日月拱照") {
			t.Errorf("期望检测到[日月拱照](日巳月酉命丑), 实际=%v", result)
		}
	})

	t.Run("日月同未命丑", func(t *testing.T) {
		stars := [12][]string{}
		stars[7] = []string{"太阳", "太阴"}
		chart := makeChart(stars)
		chart.Palaces[0].Name = ""
		chart.Palaces[1].Name = "命宫"
		result := DetectLocalPatterns(chart)
		if !containsPattern(result, "日月拱照") {
			t.Errorf("期望检测到[日月拱照](日月同未命丑), 实际=%v", result)
		}
	})

	t.Run("任意日月三合_不匹配", func(t *testing.T) {
		stars := [12][]string{}
		stars[0] = []string{"太阳"}
		stars[4] = []string{"太阴"}
		chart := makeChart(stars)
		result := DetectLocalPatterns(chart)
		if containsPattern(result, "日月拱照") {
			t.Errorf("任意日月三合不应检测到日月拱照")
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
	if containsPattern(patterns, "廉贞破军同宫") {
		t.Errorf("2003-04-15 14:00 不得把无固定来源的廉贞破军星对发布成格局: %v", patterns)
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
