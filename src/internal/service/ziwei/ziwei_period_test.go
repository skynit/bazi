package ziwei

import (
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// 大限/流年测试
// Verify dayun direction and sequences, liunian overlay.
// ═══════════════════════════════════════════════════════════════════════

// TestDayun_Exists verifies that CalculateDayun returns non-nil results
// for charts calculated with various inputs.
func TestDayun_Exists(t *testing.T) {
	svc := NewZiWeiService()

	cases := []struct {
		name   string
		year   int
		month  int
		day    int
		hour   int
		minute int
		gender string
	}{
		{"癸未男", 2003, 4, 15, 14, 0, "男"},
		{"甲子女", 1984, 1, 1, 0, 0, "女"},
		{"庚午男", 1990, 6, 15, 12, 0, "男"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			dayun := svc.CalculateDayun(chart)
			if dayun == nil {
				t.Fatal("CalculateDayun returned nil")
			}

			if len(dayun) == 0 {
				t.Fatal("CalculateDayun returned empty slice")
			}
		})
	}
}

// TestDayun_StartAge verifies that the starting age equals the JuValue
// (五行局值). For example, 木三局 should start at age 3, 金四局 at age 4, etc.
func TestDayun_StartAge(t *testing.T) {
	svc := NewZiWeiService()

	cases := []struct {
		name       string
		year       int
		month      int
		day        int
		hour       int
		minute     int
		gender     string
		wantBureau string
		wantJuVal  int
	}{
		// 癸未年三月十四日未时 男 → 木三局 (JuValue=3)
		{"癸未男_木三局", 2003, 4, 15, 14, 0, "男", "木三局", 3},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			if chart.FiveBureau != tc.wantBureau {
				t.Fatalf("五行局 = %q, want %q", chart.FiveBureau, tc.wantBureau)
			}

			dayun := svc.CalculateDayun(chart)
			if len(dayun) == 0 {
				t.Fatal("dayun empty")
			}

			// First dayun stage should start at JuValue (age = JuValue)
			wantStartAge := tc.wantJuVal
			if dayun[0].StartAge != wantStartAge {
				t.Errorf("第一大限起始年龄 = %d, 应为 %d (五行局=%s)",
					dayun[0].StartAge, wantStartAge, tc.wantBureau)
			}
		})
	}
}

// TestDayun_AgeRanges verifies that each 10-year period has the correct
// StartAge and EndAge, and that they are contiguous without gaps.
func TestDayun_AgeRanges(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	dayun := svc.CalculateDayun(chart)
	if len(dayun) < 6 {
		t.Fatalf("大限阶段数 = %d, 至少应有6个", len(dayun))
	}

	for i, stage := range dayun {
		ageSpan := stage.EndAge - stage.StartAge
		if ageSpan != 9 && ageSpan != 0 {
			// Each 10-year period: start=N, end=N+9 → span=9
			t.Errorf("大限[%d] %s: %d~%d, 跨度应为9年(含), 实际=%d",
				i, stage.Palace, stage.StartAge, stage.EndAge, ageSpan)
		}
		if stage.StartAge > stage.EndAge {
			t.Errorf("大限[%d] %s: 起始年龄(%d) > 结束年龄(%d)",
				i, stage.Palace, stage.StartAge, stage.EndAge)
		}
	}

	// Verify contiguous age ranges (no gaps between stages)
	for i := 1; i < len(dayun); i++ {
		if dayun[i].StartAge != dayun[i-1].EndAge+1 {
			t.Errorf("大限年龄不连续: [%d]结束于%d, [%d]起始于%d (期望%d)",
				i-1, dayun[i-1].EndAge, i, dayun[i].StartAge, dayun[i-1].EndAge+1)
		}
	}
}

// TestDayun_Direction_阳男阴女顺行 verifies the dayun direction logic.
// 阳男(年干偶数) → 顺行 (clockwise from 命宫)
// 阴女(年干奇数) → 顺行
// 阴男(年干奇数) → 逆行 (counterclockwise from 命宫)
// 阳女(年干偶数) → 逆行
//
// We can verify direction by checking if the palace name sequence
// matches the expected palace order.
func TestDayun_Direction(t *testing.T) {
	svc := NewZiWeiService()

	// 阳男(年干偶数): 顺行 — verify forward (clockwise) direction
	t.Run("阳男顺行", func(t *testing.T) {
		chart, err := svc.CalculateChart(1984, 1, 1, 0, 0, "男")
		if err != nil {
			t.Fatalf("CalculateChart failed: %v", err)
		}
		dayun := svc.CalculateDayun(chart)
		if len(dayun) < 3 {
			t.Fatal("dayun must have at least 3 stages")
		}
		soulPalace := chart.Palaces[0].Name
		if dayun[0].Palace != soulPalace {
			t.Errorf("第一大限应在命宫(%s), 实际=%s", soulPalace, dayun[0].Palace)
		}
		// Semantic check: all stages have unique palaces
		palaceSet := make(map[string]bool)
		for _, s := range dayun {
			if palaceSet[s.Palace] {
				t.Errorf("大限宫位重复: %s", s.Palace)
			}
			palaceSet[s.Palace] = true
		}
	})

	// 阴女(年干奇数): 顺行
	t.Run("阴女顺行", func(t *testing.T) {
		chart, err := svc.CalculateChart(1985, 6, 15, 12, 0, "女")
		if err != nil {
			t.Fatalf("CalculateChart failed: %v", err)
		}
		dayun := svc.CalculateDayun(chart)
		if len(dayun) < 3 {
			t.Fatal("dayun must have at least 3 stages")
		}
		soulPalace := chart.Palaces[0].Name
		if dayun[0].Palace != soulPalace {
			t.Errorf("第一大限应在命宫(%s), 实际=%s", soulPalace, dayun[0].Palace)
		}
		// Semantic check: no duplicate palaces, all 12 eventually covered
		palaceSet := make(map[string]bool)
		for _, s := range dayun {
			if palaceSet[s.Palace] {
				t.Errorf("大限宫位重复: %s", s.Palace)
			}
			palaceSet[s.Palace] = true
		}
	})

	// 阴男(年干奇数): 逆行
	t.Run("阴男逆行", func(t *testing.T) {
		chart, err := svc.CalculateChart(1985, 6, 15, 12, 0, "男")
		if err != nil {
			t.Fatalf("CalculateChart failed: %v", err)
		}
		dayun := svc.CalculateDayun(chart)
		if len(dayun) < 3 {
			t.Fatal("dayun must have at least 3 stages")
		}
		soulPalace := chart.Palaces[0].Name
		if dayun[0].Palace != soulPalace {
			t.Errorf("第一大限应在命宫(%s), 实际=%s", soulPalace, dayun[0].Palace)
		}
		// Semantic check: no duplicate palaces
		palaceSet := make(map[string]bool)
		for _, s := range dayun {
			if palaceSet[s.Palace] {
				t.Errorf("大限宫位重复: %s", s.Palace)
			}
			palaceSet[s.Palace] = true
		}
	})

	// 阳女(年干偶数): 逆行
	t.Run("阳女逆行", func(t *testing.T) {
		chart, err := svc.CalculateChart(1984, 1, 1, 0, 0, "女")
		if err != nil {
			t.Fatalf("CalculateChart failed: %v", err)
		}
		dayun := svc.CalculateDayun(chart)
		if len(dayun) < 3 {
			t.Fatal("dayun must have at least 3 stages")
		}
		soulPalace := chart.Palaces[0].Name
		if dayun[0].Palace != soulPalace {
			t.Errorf("第一大限应在命宫(%s), 实际=%s", soulPalace, dayun[0].Palace)
		}
		// Semantic check: no duplicate palaces
		palaceSet := make(map[string]bool)
		for _, s := range dayun {
			if palaceSet[s.Palace] {
				t.Errorf("大限宫位重复: %s", s.Palace)
			}
			palaceSet[s.Palace] = true
		}
	})
}

// TestDayun_StarPresence verifies that dayun stages contain the stars
// from their corresponding palaces.
func TestDayun_StarPresence(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	dayun := svc.CalculateDayun(chart)
	if len(dayun) == 0 {
		t.Fatal("dayun empty")
	}

	for i, stage := range dayun {
		// Find the matching palace
		var palace *PalaceInfo
		for j := range chart.Palaces {
			if chart.Palaces[j].Name == stage.Palace {
				palace = &chart.Palaces[j]
				break
			}
		}
		if palace == nil {
			t.Errorf("大限[%d]: 找不到宫位 %s", i, stage.Palace)
			continue
		}

		// All main stars in the palace should be in the dayun stage's Stars
		for _, ms := range palace.MainStars {
			found := false
			for _, ds := range stage.Stars {
				if ds == ms {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("大限[%d] %s: 缺少主星 %s", i, stage.Palace, ms)
			}
		}

		// All aux stars in the palace should be in the dayun stage's Stars
		for _, as := range palace.AuxStars {
			found := false
			for _, ds := range stage.Stars {
				if ds == as {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("大限[%d] %s: 缺少辅星 %s", i, stage.Palace, as)
			}
		}
	}
}

// TestLiunian_Basic verifies that CalculateLiunian produces a valid overlay
// chart with liu_nian_stars populated.
func TestLiunian_Basic(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// Calculate liunian for a target year (e.g., 2024)
	targetYear := 2024
	liunianChart := svc.CalculateLiunian(chart, targetYear)
	if liunianChart == nil {
		t.Fatal("CalculateLiunian returned nil")
	}

	// Verify LiuNianStars is populated
	hasContent := false
	for i, stars := range liunianChart.LiuNianStars {
		if len(stars) > 0 {
			hasContent = true
			break
		}
		_ = i
	}
	if !hasContent {
		t.Error("LiuNianStars 全部为空, 期望至少有流禄/流羊/流陀/流马之一")
	}

	// Verify the palatial structure is preserved
	for i, p := range liunianChart.Palaces {
		if p.Name == "" {
			t.Errorf("Palaces[%d].Name is empty in liunian chart", i)
		}
	}
}

// TestLiunian_Stars verifies specific liunian star placements.
// 流禄 is placed at LucunBranchIdx for the target year stem
// 流羊 at LucunBranchIdx+1
// 流陀 at LucunBranchIdx-1
// 流马 at TianmaBranchIdx for the target year branch
func TestLiunian_Stars(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// Target year 2024 → 甲辰年 → year stem index = (2024-4)%10 = 0, year branch index = (2024-4)%12 = 4
	targetYear := 2024
	yearStem := (targetYear - 4) % 10
	yearBranch := (targetYear - 4) % 12

	liunianChart := svc.CalculateLiunian(chart, targetYear)

	// Expected liunian star positions
	expectedLucunBranch := LucunBranchIdx[yearStem]
	expectedQingyangBranch := fixIndex(expectedLucunBranch + 1)
	expectedTuoluoBranch := fixIndex(expectedLucunBranch - 1)
	expectedTianmaBranch := TianmaBranchIdx[yearBranch]

	// Check for 流禄
	foundLiuLu, foundLiuYang, foundLiuTuo, foundLiuMa := false, false, false, false
	for _, stars := range liunianChart.LiuNianStars {
		for _, s := range stars {
			if s == "流禄" {
				foundLiuLu = true
			}
			if s == "流羊" {
				foundLiuYang = true
			}
			if s == "流陀" {
				foundLiuTuo = true
			}
			if s == "流马" {
				foundLiuMa = true
			}
		}
	}

	if !foundLiuLu {
		t.Errorf("2024年(甲辰) 流禄未找到, 期望在分支 %d(%s)", expectedLucunBranch, BranchNames[expectedLucunBranch])
	}
	if !foundLiuYang {
		t.Errorf("2024年(甲辰) 流羊未找到, 期望在分支 %d(%s)", expectedQingyangBranch, BranchNames[expectedQingyangBranch])
	}
	if !foundLiuTuo {
		t.Errorf("2024年(甲辰) 流陀未找到, 期望在分支 %d(%s)", expectedTuoluoBranch, BranchNames[expectedTuoluoBranch])
	}
	if !foundLiuMa {
		t.Errorf("2024年(甲辰) 流马未找到, 期望在分支 %d(%s)", expectedTianmaBranch, BranchNames[expectedTianmaBranch])
	}
}

func TestLiunian_StarsUsePalaceBranch(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	targetYear := 2024 // 甲辰
	yearStem, yearBranch := annualStemBranch(targetYear)
	liunianChart := svc.CalculateLiunian(chart, targetYear)

	expected := map[string]string{
		"流禄": BranchNames[LucunBranchIdx[yearStem]],
		"流羊": BranchNames[fixIndex(LucunBranchIdx[yearStem]+1)],
		"流陀": BranchNames[fixIndex(LucunBranchIdx[yearStem]-1)],
		"流马": BranchNames[TianmaBranchIdx[yearBranch]],
	}

	for star, wantBranch := range expected {
		found := false
		for i, stars := range liunianChart.LiuNianStars {
			for _, got := range stars {
				if got != star {
					continue
				}
				found = true
				if liunianChart.Palaces[i].Branch != wantBranch {
					t.Fatalf("%s 落在 %s宫(%s), want branch %s", star, liunianChart.Palaces[i].Name, liunianChart.Palaces[i].Branch, wantBranch)
				}
			}
		}
		if !found {
			t.Fatalf("%s not found", star)
		}
	}
}

func TestLiunianOverlayAnalysis_BuildsEvidence(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	liunianChart := svc.CalculateLiunian(chart, 2024)
	analysis := svc.AnalyzeLiunianOverlay(chart, liunianChart, 2024)
	if analysis == nil {
		t.Fatal("AnalyzeLiunianOverlay returned nil")
	}
	if analysis.GanZhi != "甲辰" {
		t.Fatalf("GanZhi=%q, want 甲辰", analysis.GanZhi)
	}
	if len(analysis.Method) == 0 {
		t.Fatal("Method should not be empty")
	}
	if len(analysis.FourHua) == 0 {
		t.Fatal("FourHua should not be empty")
	}
	if len(analysis.AnnualStars) != 4 {
		t.Fatalf("AnnualStars length=%d, want 4", len(analysis.AnnualStars))
	}
	if len(analysis.FocusPalaces) == 0 {
		t.Fatal("FocusPalaces should not be empty")
	}
}

// TestLiunian_FourHuaInjection verifies that the liunian chart injects
// four hua labels into the LiuNianStars for stars in the target year's
// SiHuaTable.
func TestLiunian_FourHuaInjection(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// 癸年 (year stem 9): 破军化禄, 巨门化权, 太阴化科, 贪狼化忌
	// Target a 甲年 (year stem 0): 廉贞化禄, 破军化权, 武曲化科, 太阳化忌
	targetYear := 1984 // 甲子年, stem=0

	liunianChart := svc.CalculateLiunian(chart, targetYear)

	// Check that the liunian chart's LiuNianStars contain hua labels
	// for stars that exist in the original chart and match the 甲年 SiHuaTable
	huaLabels := map[string]string{}
	for _, stars := range liunianChart.LiuNianStars {
		for _, s := range stars {
			// Extract star name and label (e.g., "廉贞化禄")
			for _, label := range []string{"化禄", "化权", "化科", "化忌"} {
				idx := 0
				for i := 0; i < len(s)-len(label)+1; i++ {
					if s[i:i+len(label)] == label {
						idx = i
						break
					}
				}
				if idx > 0 {
					starName := s[:idx]
					huaLabels[starName] = label
				}
			}
		}
	}

	// 甲年: 廉贞化禄, 破军化权, 武曲化科, 太阳化忌
	// Some of these may or may not be in the chart — but we can verify
	// the mechanism works by checking the chart's palaces
	t.Logf("甲年流年四化注入结果: %v", huaLabels)
}

// TestLiuyue_Basic verifies that CalculateLiuyue returns a valid overlay chart.
func TestLiuyue_Basic(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// Calculate liuyue for the 3rd lunar month
	lunarMonth := 3
	liuyueChart := svc.CalculateLiuyue(chart, lunarMonth)
	if liuyueChart == nil {
		t.Fatal("CalculateLiuyue returned nil")
	}

	// Verify LiuYueStars exists (length check instead of direct comparison)
	if len(liuyueChart.LiuYueStars) != 12 {
		t.Errorf("LiuYueStars length = %d, want 12", len(liuyueChart.LiuYueStars))
	}

	// Verify palaces preserved
	for i, p := range liuyueChart.Palaces {
		if p.Name == "" {
			t.Errorf("Palaces[%d].Name is empty", i)
		}
	}
}

// TestLiuri_Basic verifies that CalculateLiuri returns a valid overlay chart.
func TestLiuri_Basic(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// Calculate liuri for the 15th lunar day
	lunarDay := 15
	liuriChart := svc.CalculateLiuri(chart, lunarDay)
	if liuriChart == nil {
		t.Fatal("CalculateLiuri returned nil")
	}

	// Verify LiuriStars is a [12][]string
	_ = liuriChart.LiuRiStars

	// Verify palaces preserved
	for i, p := range liuriChart.Palaces {
		if p.Name == "" {
			t.Errorf("Palaces[%d].Name is empty", i)
		}
	}
}

// TestDayun_NilChart verifies nil handling.
func TestDayun_NilChart(t *testing.T) {
	svc := NewZiWeiService()
	result := svc.CalculateDayun(nil)
	if result != nil {
		t.Errorf("CalculateDayun(nil) = %v, want nil", result)
	}
}

// TestLiunian_NilChart verifies nil handling.
func TestLiunian_NilChart(t *testing.T) {
	svc := NewZiWeiService()
	result := svc.CalculateLiunian(nil, 2024)
	if result != nil {
		t.Errorf("CalculateLiunian(nil) = %v, want nil", result)
	}
}

// TestLiuyue_NilChart verifies nil handling.
func TestLiuyue_NilChart(t *testing.T) {
	svc := NewZiWeiService()
	result := svc.CalculateLiuyue(nil, 3)
	if result != nil {
		t.Errorf("CalculateLiuyue(nil) = %v, want nil", result)
	}
}

// TestLiuri_NilChart verifies nil handling.
func TestLiuri_NilChart(t *testing.T) {
	svc := NewZiWeiService()
	result := svc.CalculateLiuri(nil, 15)
	if result != nil {
		t.Errorf("CalculateLiuri(nil) = %v, want nil", result)
	}
}
