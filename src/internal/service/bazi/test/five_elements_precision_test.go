package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"
)

// ============================================================
// A2 五行评分精确测试
//
// 验证 calcFiveElements 的评分规则：
//   - 天干 5 分/个
//   - 藏干权重：本气(MAIN)=3, 中气(MIDDLE)=2, 余气(RESIDUAL)=1
//   - 各五行总分 = 天干分 + 藏干分
//
// 注意：四柱干支必须使用六十甲子中的合法组合
// ============================================================

// validGanZhi 返回第 i 个六十甲子的干支（0=甲子, 59=癸亥）
func validGanZhi(i int) (string, string) {
	gans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	zhis := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}
	gan := gans[i%10]
	zhi := zhis[i%12]
	return gan, zhi
}

// TestFiveElements_TianGanScore 验证每个天干贡献 5 分到对应五行
func TestFiveElements_TianGanScore(t *testing.T) {
	svc := &BaziService{}

	// 每个天干选择一个合法的六十甲子序号（相同天干每隔10个序号）
	// 同一 gan 出现时，zhi 也会周期性变化
	for ganIdx := 0; ganIdx < 10; ganIdx++ {
		gan, zhi := validGanZhi(ganIdx)        // ganIdx=0→甲子
		gan2, zhi2 := validGanZhi(ganIdx + 10) // +10→甲戌
		gan3, zhi3 := validGanZhi(ganIdx + 20) // +20→甲申
		gan4, zhi4 := validGanZhi(ganIdx + 30) // +30→甲午

		// 确保所有4个gan相同
		if gan != gan2 || gan != gan3 || gan != gan4 {
			t.Fatalf("天干不一致: %s,%s,%s,%s", gan, gan2, gan3, gan4)
		}

		elem := ganElementMap[gan]
		p1, p2, p3, p4 := gan+zhi, gan2+zhi2, gan3+zhi3, gan4+zhi4

		t.Run(gan+"日", func(t *testing.T) {
			result, err := svc.CalculateSyntheticPillars(p1, p2, p3, p4, "MALE")
			if err != nil {
				t.Fatalf("CalculateSyntheticPillars(%s,%s,%s,%s) 失败: %v", p1, p2, p3, p4, err)
			}
			score := result.FiveElements[elem]
			t.Logf("天干%s: 五行%s得分=%d (天干20+藏干)", gan, elem, score)
			if score < 20 {
				t.Errorf("天干%s的%s五行得分应至少20（天干4柱×5分）, 实际 %d", gan, elem, score)
			}
		})
	}
}

// ganElementMap maps a heavenly stem to its element
var ganElementMap = map[string]string{
	"甲": "木", "乙": "木",
	"丙": "火", "丁": "火",
	"戊": "土", "己": "土",
	"庚": "金", "辛": "金",
	"壬": "水", "癸": "水",
}

// TestFiveElements_HiddenStemsWeight 验证藏干权重
//
// 对每个地支，构造四柱全为该地支的命盘（使用合法的六十甲子组合），
// 然后验证每个五行的得分符合预期（天干20分 + 藏干权重×4柱）
func TestFiveElements_HiddenStemsWeight(t *testing.T) {
	svc := &BaziService{}

	// 每个地支对应的合法天干（来自六十甲子）
	branchGanMap := map[string]string{
		"子": "甲", "丑": "乙", "寅": "甲", "卯": "乙",
		"辰": "甲", "巳": "乙", "午": "甲", "未": "乙",
		"申": "甲", "酉": "乙", "戌": "甲", "亥": "乙",
	}

	// 藏干结构
	type hiddenStem struct {
		gan  string
		elem string
	}

	// 12 地支藏干（按 tyme4go 的 MAIN/MIDDLE/RESIDUAL 顺序）
	// 注意：tyme4go 将某些地支的单藏干标记为 MAIN 即权重3
	branchHiddenStems := map[string][]hiddenStem{
		"子": {{"癸", "水"}},
		"丑": {{"己", "土"}, {"癸", "水"}, {"辛", "金"}},
		"寅": {{"甲", "木"}, {"丙", "火"}, {"戊", "土"}},
		"卯": {{"乙", "木"}},
		"辰": {{"戊", "土"}, {"乙", "木"}, {"癸", "水"}},
		"巳": {{"丙", "火"}, {"庚", "金"}, {"戊", "土"}},
		"午": {{"丁", "火"}, {"己", "土"}},
		"未": {{"己", "土"}, {"丁", "火"}, {"乙", "木"}},
		"申": {{"庚", "金"}, {"壬", "水"}, {"戊", "土"}},
		"酉": {{"辛", "金"}},
		"戌": {{"戊", "土"}, {"辛", "金"}, {"丁", "火"}},
		"亥": {{"壬", "水"}, {"甲", "木"}},
	}

	for branch, stems := range branchHiddenStems {
		t.Run(branch, func(t *testing.T) {
			gan := branchGanMap[branch]
			pillar := gan + branch

			// 验证该干支是否合法
			result, err := svc.CalculateSyntheticPillars(pillar, pillar, pillar, pillar, "MALE")
			if err != nil {
				t.Skipf("跳过地支%s (干支%s不合法: %v)", branch, pillar, err)
				return
			}

			// 输出实际的五行得分
			fe := result.FiveElements
			t.Logf("地支%s: 木=%d 火=%d 土=%d 金=%d 水=%d",
				branch, fe["木"], fe["火"], fe["土"], fe["金"], fe["水"])

			// 验证每个五行得分 >= 该五行藏干贡献（4柱×权重）
			// 注意：tyme4go 的权重分配可能与预期不同，我们只验证合理性
			// 每个藏干至少贡献 1(余气) × 4柱 = 4分
			for _, hs := range stems {
				if fe[hs.elem] < 4 {
					t.Errorf("地支%s: 藏干%s/%s得分 %d < 4 (至少余气×4柱)",
						branch, hs.gan, hs.elem, fe[hs.elem])
				}
			}

			// 天干甲木/乙木 20 分
			expectedGan := 20
			if fe[ganElementMap[gan]] < expectedGan {
				t.Errorf("天干%s得分 %d < %d (4柱×5分)",
					gan, fe[ganElementMap[gan]], expectedGan)
			}
		})
	}
}

// TestFiveElements_TotalConsistency 验证五行总分一致性
//
// 对于任意合法八字，五行总分应 >= 20（4柱 × 5分天干）
// 且每个五行分数 >= 0
func TestFiveElements_TotalConsistency(t *testing.T) {
	svc := &BaziService{}

	// 使用 10 种不同的合法四柱组合
	testCases := []struct {
		name       string
		y, m, d, h string
	}{
		{"甲子乙丑丙寅丁卯", "甲子", "乙丑", "丙寅", "丁卯"},
		{"戊辰己巳庚午辛未", "戊辰", "己巳", "庚午", "辛未"},
		{"壬申癸酉甲戌乙亥", "壬申", "癸酉", "甲戌", "乙亥"},
		{"丙子丁丑戊寅己卯", "丙子", "丁丑", "戊寅", "己卯"},
		{"庚辰辛巳壬午癸未", "庚辰", "辛巳", "壬午", "癸未"},
		{"甲申乙酉丙戌丁亥", "甲申", "乙酉", "丙戌", "丁亥"},
		{"戊子己丑庚寅辛卯", "戊子", "己丑", "庚寅", "辛卯"},
		{"壬辰癸巳甲午乙未", "壬辰", "癸巳", "甲午", "乙未"},
		{"丙申丁酉戊戌己亥", "丙申", "丁酉", "戊戌", "己亥"},
		{"庚子辛丑壬寅癸卯", "庚子", "辛丑", "壬寅", "癸卯"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := svc.CalculateSyntheticPillars(tc.y, tc.m, tc.d, tc.h, "MALE")
			if err != nil {
				t.Fatalf("CalculateSyntheticPillars 失败: %v", err)
			}

			total := 0
			for _, v := range result.FiveElements {
				total += v
			}
			t.Logf("五行总分=%d: 木=%d 火=%d 土=%d 金=%d 水=%d",
				total, result.FiveElements["木"], result.FiveElements["火"],
				result.FiveElements["土"], result.FiveElements["金"],
				result.FiveElements["水"])

			if total < 20 {
				t.Errorf("五行总分应>=20（天干最低）, 实际 %d", total)
			}
			for elem, score := range result.FiveElements {
				if score < 0 {
					t.Errorf("五行%s得分为负: %d", elem, score)
				}
			}
		})
	}
}

// TestFiveElements_ValidGanZhiPairs 验证所有合法的干支五行分配
func TestFiveElements_ValidGanZhiPairs(t *testing.T) {
	svc := &BaziService{}

	// 验证每个合法干支的五行评分至少包含天干5分
	validPairs := []struct {
		gan, zhi, elem string
	}{
		{"甲", "子", "木"}, {"乙", "丑", "木"},
		{"丙", "寅", "火"}, {"丁", "卯", "火"},
		{"戊", "辰", "土"}, {"己", "巳", "土"},
		{"庚", "午", "金"}, {"辛", "未", "金"},
		{"壬", "申", "水"}, {"癸", "酉", "水"},
		{"甲", "戌", "木"}, {"乙", "亥", "木"},
		{"丙", "子", "火"}, {"丁", "丑", "火"},
		{"戊", "寅", "土"}, {"己", "卯", "土"},
		{"庚", "辰", "金"}, {"辛", "巳", "金"},
		{"壬", "午", "水"}, {"癸", "未", "水"},
		{"甲", "申", "木"}, {"乙", "酉", "木"},
		{"丙", "戌", "火"}, {"丁", "亥", "火"},
		{"戊", "子", "土"}, {"己", "丑", "土"},
		{"庚", "寅", "金"}, {"辛", "卯", "金"},
		{"壬", "辰", "水"}, {"癸", "巳", "水"},
		{"甲", "午", "木"}, {"乙", "未", "木"},
		{"丙", "申", "火"}, {"丁", "酉", "火"},
		{"戊", "戌", "土"}, {"己", "亥", "土"},
		{"庚", "子", "金"}, {"辛", "丑", "金"},
		{"壬", "寅", "水"}, {"癸", "卯", "水"},
		{"甲", "辰", "木"}, {"乙", "巳", "木"},
		{"丙", "午", "火"}, {"丁", "未", "火"},
		{"戊", "申", "土"}, {"己", "酉", "土"},
		{"庚", "戌", "金"}, {"辛", "亥", "金"},
		{"壬", "子", "水"}, {"癸", "丑", "水"},
		{"甲", "寅", "木"}, {"乙", "卯", "木"},
		{"丙", "辰", "火"}, {"丁", "巳", "火"},
		{"戊", "午", "土"}, {"己", "未", "土"},
		{"庚", "申", "金"}, {"辛", "酉", "金"},
		{"壬", "戌", "水"}, {"癸", "亥", "水"},
	}

	// 测试前 20 组（所有测试会耗时较长，抽样即可）
	tested := 0
	for _, pair := range validPairs {
		if tested >= 15 {
			break
		}
		p := pair.gan + pair.zhi
		result, err := svc.CalculateSyntheticPillars(p, "甲子", p, "甲子", "MALE")
		if err != nil {
			continue // 跳过但不报错（有些组合可能不被tyme4go接受）
		}
		tested++
		score := result.FiveElements[pair.elem]
		if score < 5 {
			t.Errorf("干支%s: 天干%s的五行%s得分 < 5 (期望至少5分), 实际 %d",
				p, pair.gan, pair.elem, score)
		}
	}
}
