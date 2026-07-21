package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"
)

// ============================================================
// A12 命宫精确测试
//
// 验证 CalculateSyntheticPillars 的命宫计算结果：
//   - GanZhi 非空
//   - Nayin 非空
//   - ShenSha 非空
//   - 10+ 种不同干支组合
// ============================================================

// TestMingGongPrecision_Basic 验证多种组合的命宫基本完整性
//
// 覆盖不同的年干、月支、时支组合，验证：
//  1. GanZhi 非空且格式正确（2字符）
//  2. Gan 非空
//  3. Zhi 非空
//  4. Nayin 非空
//  5. ShenSha 非空
func TestMingGongPrecision_Basic(t *testing.T) {
	svc := &BaziService{}

	testCases := []struct {
		name                       string
		yearP, monthP, dayP, hourP string
	}{
		// 10种以上不同组合
		{"壬辰壬寅甲寅庚午-经典007", "壬辰", "壬寅", "甲寅", "庚午"},
		{"戊辰己未癸未辛酉-经典014", "戊辰", "己未", "癸未", "辛酉"},
		{"辛亥辛丑辛酉戊子-经典010", "辛亥", "辛丑", "辛酉", "戊子"},
		{"甲子甲戌丁卯戊申-甲子年", "甲子", "甲戌", "丁卯", "戊申"},
		{"乙丑丙戌丙子庚寅-乙丑年", "乙丑", "丙戌", "丙子", "庚寅"},
		{"丙寅丁酉戊午壬子-丙寅年", "丙寅", "丁酉", "戊午", "壬子"},
		{"丁卯戊申己巳乙亥-丁卯年", "丁卯", "戊申", "己巳", "乙亥"},
		{"戊辰乙卯庚辰丙子-戊辰年", "戊辰", "乙卯", "庚辰", "丙子"},
		{"己巳庚戌壬午癸卯-己巳年", "己巳", "庚戌", "壬午", "癸卯"},
		{"庚午辛丑甲申乙丑-庚午年", "庚午", "辛丑", "甲申", "乙丑"},
		{"辛未辛亥丙戌丁酉-辛未年", "辛未", "辛亥", "丙戌", "丁酉"},
		{"壬申壬子戊子己未-壬申年", "壬申", "壬子", "戊子", "己未"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := svc.CalculateSyntheticPillars(tc.yearP, tc.monthP, tc.dayP, tc.hourP, "MALE")
			if err != nil {
				t.Fatalf("CalculateSyntheticPillars 失败: %v", err)
			}

			mg := result.MingGong
			t.Logf("命宫: GanZhi=%s Gan=%s Zhi=%s Nayin=%s ShenSha=%s",
				mg.GanZhi, mg.Gan, mg.Zhi, mg.Nayin, mg.ShenSha)

			// GanZhi 非空且格式正确
			if mg.GanZhi == "" {
				t.Error("MingGong.GanZhi 应非空")
			} else {
				runes := []rune(mg.GanZhi)
				if len(runes) != 2 {
					t.Errorf("MingGong.GanZhi 格式异常: %q (len=%d)", mg.GanZhi, len(runes))
				}
			}

			// Gan 非空
			if mg.Gan == "" {
				t.Error("MingGong.Gan 应非空")
			}

			// Zhi 非空
			if mg.Zhi == "" {
				t.Error("MingGong.Zhi 应非空")
			}

			// Nayin 非空
			if mg.Nayin == "" {
				t.Error("MingGong.Nayin 应非空")
			}

			// ShenSha 非空
			if mg.ShenSha == "" {
				t.Error("MingGong.ShenSha 应非空")
			}

		})
	}
}

// TestMingGongPrecision_NayinVariety 验证不同命宫纳音不同
func TestMingGongPrecision_NayinVariety(t *testing.T) {
	svc := &BaziService{}

	// 选取不同的干支组合，命宫纳音应该不同（至少不全相同）
	nayins := make(map[string]bool)
	cases := []struct {
		name       string
		y, m, d, h string
	}{
		{"甲子甲戌丁卯戊申", "甲子", "甲戌", "丁卯", "戊申"},
		{"乙丑丙戌丙子庚寅", "乙丑", "丙戌", "丙子", "庚寅"},
		{"丙寅丁酉戊午壬子", "丙寅", "丁酉", "戊午", "壬子"},
		{"丁卯戊申己巳乙亥", "丁卯", "戊申", "己巳", "乙亥"},
		{"戊辰乙卯庚辰丙子", "戊辰", "乙卯", "庚辰", "丙子"},
		{"庚午辛丑甲申乙丑", "庚午", "辛丑", "甲申", "乙丑"},
		{"壬申壬子戊子己未", "壬申", "壬子", "戊子", "己未"},
	}

	for _, tc := range cases {
		result, err := svc.CalculateSyntheticPillars(tc.y, tc.m, tc.d, tc.h, "MALE")
		if err != nil {
			t.Fatalf("%s: CalculateSyntheticPillars 失败: %v", tc.name, err)
		}
		nayins[result.MingGong.Nayin] = true
		t.Logf("%s: 命宫=%s 纳音=%s", tc.name, result.MingGong.GanZhi, result.MingGong.Nayin)
	}

	// 至少要有不同的纳音
	if len(nayins) < 3 {
		t.Errorf("命宫纳音多样性不足, 仅有 %d 种: %v", len(nayins), nayins)
	}
}

// TestMingGongPrecision_GenderIndependence 验证命宫与性别无关
//
// 命宫计算只依赖于年干、月支、时支，与性别无关
func TestMingGongPrecision_GenderIndependence(t *testing.T) {
	svc := &BaziService{}

	r1, err1 := svc.CalculateSyntheticPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
	r2, err2 := svc.CalculateSyntheticPillars("壬辰", "壬寅", "甲寅", "庚午", "FEMALE")

	if err1 != nil || err2 != nil {
		t.Fatalf("CalculateSyntheticPillars 失败: %v / %v", err1, err2)
	}

	if r1.MingGong.GanZhi != r2.MingGong.GanZhi {
		t.Errorf("命宫应与性别无关: MALE=%s FEMALE=%s",
			r1.MingGong.GanZhi, r2.MingGong.GanZhi)
	}
	if r1.MingGong.Nayin != r2.MingGong.Nayin {
		t.Errorf("命宫纳音应与性别无关: MALE=%s FEMALE=%s",
			r1.MingGong.Nayin, r2.MingGong.Nayin)
	}
	if r1.MingGong.ShenSha != r2.MingGong.ShenSha {
		t.Errorf("命宫神煞应与性别无关: MALE=%s FEMALE=%s",
			r1.MingGong.ShenSha, r2.MingGong.ShenSha)
	}
}

// TestMingGongPrecision_ShenShaConsistency 验证命宫神煞非空
//
// 命宫神煞与地支固定对应，所有合法命宫都应有神煞
func TestMingGongPrecision_ShenShaConsistency(t *testing.T) {
	svc := &BaziService{}

	// 使用经典命例的组合计算命宫，验证其神煞非空
	cases := []struct {
		name       string
		y, m, d, h string
	}{
		{"壬辰壬寅甲寅庚午", "壬辰", "壬寅", "甲寅", "庚午"},
		{"戊辰己未癸未辛酉", "戊辰", "己未", "癸未", "辛酉"},
		{"辛亥辛丑辛酉戊子", "辛亥", "辛丑", "辛酉", "戊子"},
		{"庚申丙寅甲午辛酉", "庚申", "丙寅", "甲午", "辛酉"},
		{"壬子戊申甲午癸亥", "壬子", "戊申", "甲午", "癸亥"},
		{"庚申甲申甲寅庚午", "庚申", "甲申", "甲寅", "庚午"},
		{"甲子甲戌丁卯戊申", "甲子", "甲戌", "丁卯", "戊申"},
		{"丙寅丁酉戊午壬子", "丙寅", "丁酉", "戊午", "壬子"},
		{"戊辰乙卯庚辰丙子", "戊辰", "乙卯", "庚辰", "丙子"},
		{"己巳庚戌壬午癸卯", "己巳", "庚戌", "壬午", "癸卯"},
	}

	uniqueShenSha := make(map[string]bool)
	for _, tc := range cases {
		result, err := svc.CalculateSyntheticPillars(tc.y, tc.m, tc.d, tc.h, "MALE")
		if err != nil {
			t.Fatalf("%s: CalculateSyntheticPillars 失败: %v", tc.name, err)
		}
		mg := result.MingGong
		if mg.ShenSha == "" {
			t.Errorf("%s: 命宫神煞为空 (命宫=%s)", tc.name, mg.GanZhi)
		} else {
			uniqueShenSha[mg.ShenSha] = true
		}
		t.Logf("%s: 命宫=%s 地支=%s 神煞=%s", tc.name, mg.GanZhi, mg.Zhi, mg.ShenSha)
	}

	// 至少覆盖 3 种不同神煞
	if len(uniqueShenSha) < 3 {
		t.Errorf("命宫神煞多样性不足, 仅有 %d 种: %v", len(uniqueShenSha), uniqueShenSha)
	}
}
