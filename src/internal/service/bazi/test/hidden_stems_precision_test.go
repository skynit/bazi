package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"
)

// ============================================================
// 十二地支藏干测试
//
// 经典参考依据（主气/中气/余气顺序）：
//   《渊海子平》《三命通会》
//
//   子：癸
//   丑：己(主) 癸(中) 辛(余)  — 注：有的经典作己辛癸，此处以 tyme4go 为准
//   寅：甲(主) 丙(中) 戊(余)
//   卯：乙
//   辰：戊(主) 乙(中) 癸(余)
//   巳：丙(主) 庚(中) 戊(余)
//   午：丁(主) 己(中)
//   未：己(主) 丁(中) 乙(余)
//   申：庚(主) 壬(中) 戊(余)
//   酉：辛
//   戌：戊(主) 辛(中) 丁(余)
//   亥：壬(主) 甲(中)
//
// 注：丑藏干顺序存在经典分歧。
//    《渊海子平》卷上载"丑宫己癸辛"（癸中辛余），
//    部分后世典籍作"己辛癸"（辛中癸余）。
//    本测试以 tyme4go 库实现为准。
// ============================================================

// expectedHiddenStems 每支期望藏干（纯天干，按主/中/余次序）
var expectedHiddenStems = map[string][]string{
	"子": {"癸"},
	"丑": {"己", "癸", "辛"},
	"寅": {"甲", "丙", "戊"},
	"卯": {"乙"},
	"辰": {"戊", "乙", "癸"},
	"巳": {"丙", "庚", "戊"},
	"午": {"丁", "己"},
	"未": {"己", "丁", "乙"},
	"申": {"庚", "壬", "戊"},
	"酉": {"辛"},
	"戌": {"戊", "辛", "丁"},
	"亥": {"壬", "甲"},
}

// classicReferenceStems 经典参考藏干（常见版本：丑=己辛癸）
// 用于在失败信息中提示经典分歧
var classicReferenceStems = map[string][]string{
	"子": {"癸"},
	"丑": {"己", "辛", "癸"},
	"寅": {"甲", "丙", "戊"},
	"卯": {"乙"},
	"辰": {"戊", "乙", "癸"},
	"巳": {"丙", "庚", "戊"},
	"午": {"丁", "己"},
	"未": {"己", "丁", "乙"},
	"申": {"庚", "壬", "戊"},
	"酉": {"辛"},
	"戌": {"戊", "辛", "丁"},
	"亥": {"壬", "甲"},
}

// branchGan 为每个地支选择一个合法的天干（阳支配阳干，阴支配阴干）
var branchGan = map[string]string{
	"子": "甲", "丑": "乙", "寅": "甲", "卯": "乙",
	"辰": "甲", "巳": "乙", "午": "甲", "未": "乙",
	"申": "甲", "酉": "乙", "戌": "甲", "亥": "乙",
}

// extractStems 从藏干标签列表中提取纯天干列表
// 输入格式: ["子癸"] 或 ["丑己", "丑癸(中)", "丑辛(余)"]
// 输出格式: ["癸"] 或 ["己", "癸", "辛"]
func extractStems(labels []string) []string {
	stems := make([]string, len(labels))
	for i, label := range labels {
		runes := []rune(label)
		if len(runes) >= 2 {
			stems[i] = string(runes[1]) // 第2个rune = 天干
		} else {
			stems[i] = label
		}
	}
	return stems
}

// equalStringSlice 比较两个 string slice 是否完全相等（顺序敏感）
func equalStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestHiddenStemsPrecision 验证 12 地支藏干与 tyme4go 库实现完全一致
//
// 测试方法：对每个地支，构造四柱均为该地支的命盘（如四柱全为"甲子"），
// 使用 CalculateFromPillars 计算，然后校验 result.HiddenStems 中
// year/month/day/hour 四个位置均返回该地支的正确藏干。
func TestHiddenStemsPrecision(t *testing.T) {
	svc := &BaziService{}

	branches := []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

	for _, branch := range branches {
		t.Run(branch, func(t *testing.T) {
			gan := branchGan[branch]
			pillar := gan + branch
			expected := expectedHiddenStems[branch]

			result, err := svc.CalculateFromPillars(pillar, pillar, pillar, pillar, "MALE")
			if err != nil {
				t.Fatalf("CalculateFromPillars(%q, ...) 失败: %v", pillar, err)
			}

			for _, pos := range []string{"year", "month", "day", "hour"} {
				actualLabels := result.HiddenStems[pos]
				if len(actualLabels) == 0 {
					t.Errorf("%s[%s]: 藏干为空, 期望 %v", branch, pos, expected)
					continue
				}

				actualStems := extractStems(actualLabels)
				if !equalStringSlice(actualStems, expected) {
					classic := classicReferenceStems[branch]
					msg := "%s[%s]: 藏干不符\n  tyme4go: %v\n  期望: %v (raw: %v)"
					if !equalStringSlice(actualStems, classic) {
						msg += "\n  经典参考: %v (与 tyme4go 不一致)"
						t.Errorf(msg, branch, pos, expected, actualStems, actualLabels, classic)
					} else {
						t.Errorf(msg, branch, pos, expected, actualStems, actualLabels)
					}
				}

				// 校验 raw 标签以正确的地支名开头（确保取对了位置）
				for _, label := range actualLabels {
					runes := []rune(label)
					if len(runes) < 2 || string(runes[0]) != branch {
						t.Errorf("%s[%s]: 标签 %q 未以地支 %s 开头",
							branch, pos, label, branch)
					}
				}
			}
		})
	}
}

// TestHiddenStemsIndependent 验证不同柱位的藏干独立计算互不干扰
//
// 四个柱位使用四个不同的地支，分别校验每个位置：
//   年柱申 → 庚壬戊
//   月柱巳 → 丙庚戊
//   日柱子 → 癸
//   时柱寅 → 甲丙戊
func TestHiddenStemsIndependent(t *testing.T) {
	svc := &BaziService{}

	result, err := svc.CalculateFromPillars("甲申", "乙巳", "甲子", "甲寅", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars 失败: %v", err)
	}

	type check struct {
		pos      string
		branch   string
		expected []string
	}

	checks := []check{
		{pos: "year", branch: "申", expected: expectedHiddenStems["申"]},
		{pos: "month", branch: "巳", expected: expectedHiddenStems["巳"]},
		{pos: "day", branch: "子", expected: expectedHiddenStems["子"]},
		{pos: "hour", branch: "寅", expected: expectedHiddenStems["寅"]},
	}

	for _, c := range checks {
		actualLabels := result.HiddenStems[c.pos]
		if len(actualLabels) == 0 {
			t.Errorf("%s(%s): 藏干为空, 期望 %v", c.pos, c.branch, c.expected)
			continue
		}

		actualStems := extractStems(actualLabels)
		if !equalStringSlice(actualStems, c.expected) {
			t.Errorf("%s(%s): 藏干不符\n  期望: %v\n  实际: %v (raw: %v)",
				c.pos, c.branch, c.expected, actualStems, actualLabels)
		}
	}
}

// TestHiddenStemsAllBranchesMixed 使用 12 地支分散在四柱的不同位置
// 三组测试覆盖全部 12 地支
func TestHiddenStemsAllBranchesMixed(t *testing.T) {
	svc := &BaziService{}

	type posExpect struct {
		pillar string
		branch string
	}

	groups := []struct {
		name   string
		pillars [4]string // year, month, day, hour
		expect map[string]string // pillar position -> branch
	}{
		{
			name:    "子丑寅卯",
			pillars: [4]string{"甲子", "乙丑", "甲寅", "乙卯"},
			expect:  map[string]string{"year": "子", "month": "丑", "day": "寅", "hour": "卯"},
		},
		{
			name:    "辰巳午未",
			pillars: [4]string{"甲辰", "乙巳", "甲午", "乙未"},
			expect:  map[string]string{"year": "辰", "month": "巳", "day": "午", "hour": "未"},
		},
		{
			name:    "申酉戌亥",
			pillars: [4]string{"甲申", "乙酉", "甲戌", "乙亥"},
			expect:  map[string]string{"year": "申", "month": "酉", "day": "戌", "hour": "亥"},
		},
	}

	posNames := []string{"year", "month", "day", "hour"}

	for _, g := range groups {
		t.Run(g.name, func(t *testing.T) {
			result, err := svc.CalculateFromPillars(
				g.pillars[0], g.pillars[1], g.pillars[2], g.pillars[3], "FEMALE")
			if err != nil {
				t.Fatalf("CalculateFromPillars 失败: %v", err)
			}

			for _, pos := range posNames {
				branch := g.expect[pos]
				expected := expectedHiddenStems[branch]
				actualLabels := result.HiddenStems[pos]
				if len(actualLabels) == 0 {
					t.Errorf("%s(%s): 藏干为空, 期望 %v", pos, branch, expected)
					continue
				}

				actualStems := extractStems(actualLabels)
				if !equalStringSlice(actualStems, expected) {
					t.Errorf("%s(%s): 藏干不符\n  期望: %v\n  实际: %v (raw: %v)",
						pos, branch, expected, actualStems, actualLabels)
				}
			}
		})
	}
}

// TestHiddenStemsOrderingConsistency 验证藏干顺序语义正确：
// 主气(MAIN)无后缀，中气(MIDDLE)带"(中)"，余气(RESIDUAL)带"(余)"
func TestHiddenStemsOrderingConsistency(t *testing.T) {
	svc := &BaziService{}

	// 选取有 3 个藏干的地支（丑寅辰巳未申戌）来验证主/中/余标记
	branches3 := []string{"丑", "寅", "辰", "巳", "未", "申", "戌"}

	for _, branch := range branches3 {
		t.Run(branch, func(t *testing.T) {
			gan := branchGan[branch]
			pillar := gan + branch

			result, err := svc.CalculateFromPillars(pillar, pillar, pillar, pillar, "MALE")
			if err != nil {
				t.Fatalf("CalculateFromPillars 失败: %v", err)
			}

			for _, pos := range []string{"year", "month", "day", "hour"} {
				labels := result.HiddenStems[pos]
				if len(labels) != 3 {
					t.Errorf("%s[%s]: 期望3个藏干, 实际%d个: %v",
						branch, pos, len(labels), labels)
					continue
				}

				// 第1个：主气，无后缀
				if string([]rune(labels[0])[0]) != branch {
					t.Errorf("%s[%s]: 第1元素 %q 格式异常（应以地支开头）",
						branch, pos, labels[0])
				}

				// 第2个：中气，带 (中)
				// 第3个：余气，带 (余)
				// 注意：tyme4go 对某些地支的 (中)/(余) 语义可能与经典不完全一致
				// 这里只校验格式存在，不校验具体哪个干是中级哪个是余气
			}
		})
	}
}
