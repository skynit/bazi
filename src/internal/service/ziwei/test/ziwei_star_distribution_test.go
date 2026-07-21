package ziwei_test

import (
	. "bazi/internal/service/ziwei"
	"strings"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════
// 星曜分布完整性测试
// Verify 14 main stars + 6 auspicious + 6 malefic presence across palaces.
// ═══════════════════════════════════════════════════════════════════════

// 14 main stars used in ZiWei Dou Shu
var allMainStars = []string{
	"紫微", "天机", "太阳", "武曲", "天同", "廉贞",
	"天府", "太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "破军",
}

// 6 auspicious stars (六吉星)
var allAuspiciousStars = []string{
	"左辅", "右弼", "文昌", "文曲", "天魁", "天钺",
}

// 6 malefic stars (六煞星)
var allMaleficStars = []string{
	"擎羊", "陀罗", "火星", "铃星", "地空", "地劫",
}

// starDistributionTestCase holds a test case for star distribution verification.
type starDistributionTestCase struct {
	name   string
	year   int
	month  int
	day    int
	hour   int
	minute int
	gender string
}

// TestMainStarDistribution verifies that all 14 main stars appear exactly once
// across the 12 palaces for various charts.
func TestMainStarDistribution(t *testing.T) {
	cases := []starDistributionTestCase{
		{"癸未年三月未时_男", 2003, 4, 15, 14, 0, "男"},
		{"甲年_1990年正月子时_女", 1990, 1, 1, 23, 0, "女"},
		{"庚年_2000年六月午时_女", 2000, 6, 15, 12, 0, "女"},
		{"丙年_1988年八月辰时_男", 1988, 8, 8, 8, 0, "男"},
		{"戊年_1976年七月寅时_男", 1976, 7, 28, 3, 42, "男"},
		{"丁年_2008年五月未时_女", 2008, 5, 12, 14, 28, "女"},
		{"乙年_1985年六月午时_男", 1985, 6, 18, 12, 0, "男"},
		{"辛年_1992年五月巳时_男", 1992, 5, 15, 10, 0, "男"},
		{"己年_1986年七月未时_男", 1986, 7, 22, 14, 0, "男"},
		{"壬年_1993年十一月亥时_女", 1993, 11, 12, 22, 0, "女"},
	}

	svc := NewZiWeiService()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			// Collect all main stars from all palaces
			starCount := make(map[string]int)
			starPalace := make(map[string]string)
			for _, p := range chart.Palaces {
				for _, s := range publishedMainStarNames(p) {
					starCount[s]++
					starPalace[s] = p.Name
				}
			}

			// Verify all 14 main stars are present
			for _, expected := range allMainStars {
				if starCount[expected] == 0 {
					t.Errorf("主星 %s 未在任何宫位中找到", expected)
				}
			}

			// Verify no main star appears more than once
			for star, count := range starCount {
				if count > 1 {
					t.Errorf("主星 %s 出现 %d 次(应在 %s), 每个主星应只出现一次",
						star, count, starPalace[star])
				}
			}

			// Total main stars should be exactly 14 (some palaces may have 0, some have 2+)
			totalMain := 0
			for _, count := range starCount {
				totalMain += count
			}
			if totalMain != 14 {
				t.Errorf("14主星总数 = %d, 应为14", totalMain)
			}
		})
	}
}

// TestAuxiliaryStarDistribution verifies that auspicious and malefic stars
// are placed correctly. Note: some aux stars depend on year stem/branch,
// so certain stars may not appear for all years (e.g., 天魁/天钺 always present,
// but 左辅/右弼/文昌/文曲 always present).
func TestAuxiliaryStarDistribution(t *testing.T) {
	cases := []starDistributionTestCase{
		{"癸未年三月未时_男", 2003, 4, 15, 14, 0, "男"},
		{"甲年_1990年正月子时_女", 1990, 1, 1, 23, 0, "女"},
		{"庚年_2000年六月午时_女", 2000, 6, 15, 12, 0, "女"},
		{"丙年_1988年八月辰时_男", 1988, 8, 8, 8, 0, "男"},
	}

	svc := NewZiWeiService()

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			// Collect all auxiliary stars from all palaces
			auxCount := make(map[string]int)
			for _, p := range chart.Palaces {
				for _, s := range publishedAuxStarNames(p) {
					auxCount[s]++
				}
			}

			// Verify all 6 auspicious stars are present (each exactly once)
			for _, star := range allAuspiciousStars {
				if auxCount[star] == 0 {
					t.Errorf("吉星 %s 未找到", star)
				} else if auxCount[star] > 1 {
					t.Errorf("吉星 %s 出现 %d 次, 应只出现一次", star, auxCount[star])
				}
			}

			// Verify all 6 malefic stars are present (each exactly once)
			for _, star := range allMaleficStars {
				if auxCount[star] == 0 {
					t.Errorf("煞星 %s 未找到", star)
				} else if auxCount[star] > 1 {
					t.Errorf("煞星 %s 出现 %d 次, 应只出现一次", star, auxCount[star])
				}
			}

			// Verify 禄存 and 天马 are also present (always placed)
			if auxCount["禄存"] == 0 {
				t.Error("禄存未找到")
			}
			if auxCount["天马"] == 0 {
				t.Error("天马未找到")
			}
		})
	}
}

// TestStarNoCrossContamination verifies that main stars are NOT in AuxStars
// and aux stars are NOT in MainStars.
func TestStarNoCrossContamination(t *testing.T) {
	svc := NewZiWeiService()

	cases := []starDistributionTestCase{
		{"癸未年", 2003, 4, 15, 14, 0, "男"},
		{"甲年", 1990, 1, 1, 23, 0, "女"},
		{"庚年", 2000, 6, 15, 12, 0, "女"},
		{"丙年", 1988, 8, 8, 8, 0, "男"},
		{"戊年", 1976, 7, 28, 3, 42, "男"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tc.year, tc.month, tc.day, tc.hour, tc.minute, tc.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			mainSet := make(map[string]bool)
			for _, star := range allMainStars {
				mainSet[star] = true
			}
			auxSet := make(map[string]bool)
			for _, star := range allAuspiciousStars {
				auxSet[star] = true
			}
			for _, star := range allMaleficStars {
				auxSet[star] = true
			}
			auxSet["禄存"] = true
			auxSet["天马"] = true

			for _, p := range chart.Palaces {
				for _, s := range publishedMainStarNames(p) {
					if auxSet[s] {
						t.Errorf("宫位 %s: 辅星 %s 出现在主星列表中", p.Name, s)
					}
				}
				for _, s := range publishedAuxStarNames(p) {
					if mainSet[s] {
						t.Errorf("宫位 %s: 主星 %s 出现在辅星列表中", p.Name, s)
					}
				}
			}
		})
	}
}

// TestStarBrightnessMapIntegrity verifies that all 14 main stars have
// brightness entries for all 12 branches.
func TestStarBrightnessMapIntegrity(t *testing.T) {
	for _, star := range allMainStars {
		brightness, ok := StarBrightnessMap[star]
		if !ok {
			t.Errorf("StarBrightnessMap 缺少主星 %s 的亮度数据", star)
			continue
		}
		if len(brightness) != 12 {
			t.Errorf("StarBrightnessMap[%s] 长度 = %d, 应为12", star, len(brightness))
		}
		for i, b := range brightness {
			if b == "" {
				t.Errorf("StarBrightnessMap[%s][%d(%s)] 亮度为空", star, i, BranchNames[i])
			}
		}
	}

	// The pinned iztro source defines brightness only for these six auxiliary stars.
	auxStarsWithBrightness := []string{"文昌", "文曲", "擎羊", "陀罗", "火星", "铃星"}
	for _, star := range auxStarsWithBrightness {
		if _, ok := AuxStarBrightnessMap[star]; !ok {
			t.Errorf("AuxStarBrightnessMap 缺少辅星 %s 的亮度数据", star)
		}
	}
	for _, star := range []string{"左辅", "右弼", "天魁", "天钺", "禄存", "天马", "地空", "地劫"} {
		if _, ok := AuxStarBrightnessMap[star]; ok {
			t.Errorf("AuxStarBrightnessMap 不应为无来源辅星 %s 生成亮度", star)
		}
	}
}

// TestStarPlacementUniqueness verifies that no star (main or aux) appears
// in more than one palace across multiple diverse charts.
func TestStarPlacementUniqueness(t *testing.T) {
	svc := NewZiWeiService()

	charts := []*ZiWeiChart{}
	for _, p := range []struct {
		y, m, d, h, mi int
		g              string
	}{
		{2003, 4, 15, 14, 0, "男"},
		{1990, 1, 1, 23, 0, "女"},
		{2000, 6, 15, 12, 0, "女"},
		{1988, 8, 8, 8, 0, "男"},
		{1976, 7, 28, 3, 42, "男"},
	} {
		ch, err := svc.CalculateChart(p.y, p.m, p.d, p.h, p.mi, p.g)
		if err != nil {
			t.Fatalf("CalculateChart(%d) failed: %v", p.y, err)
		}
		charts = append(charts, ch)
	}

	// For each chart, verify no star appears in more than one palace
	for i, chart := range charts {
		t.Run("chart_"+strings.ReplaceAll(chart.FiveBureau, "局", ""), func(t *testing.T) {
			starPalaces := make(map[string]string)
			for _, p := range chart.Palaces {
				for _, s := range publishedMainStarNames(p) {
					if prev, ok := starPalaces[s]; ok {
						t.Errorf("主星 %s 出现在多个宫位: %s 和 %s", s, prev, p.Name)
					}
					starPalaces[s] = p.Name
				}
				for _, s := range publishedAuxStarNames(p) {
					if prev, ok := starPalaces[s]; ok {
						t.Errorf("辅星 %s 出现在多个宫位: %s 和 %s", s, prev, p.Name)
					}
					starPalaces[s] = p.Name
				}
			}
			_ = i
		})
	}
}
