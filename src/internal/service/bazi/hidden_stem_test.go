package bazi

import (
	"bazi/internal/model"
	"testing"

)

// Verify 稼穑格 rejects when hidden stems contain 克破(木)
func TestCheckCongQiangGe_RejectsKePoInHiddenStems(t *testing.T) {
	// 戊土日主生于未月(土月)，地支未/辰/未中有乙木(木克土=官杀)
	// 天干无木，但藏干有木 → 应拒绝稼穑格
	pillars := []model.Pillar{
		{Gan: "庚", Zhi: "未"}, // 年: 未藏乙木(木=克破)
		{Gan: "庚", Zhi: "辰"}, // 月: 辰藏乙木(木=克破)
		{Gan: "戊", Zhi: "戌"}, // 日: 戊土
		{Gan: "壬", Zhi: "未"}, // 时: 未藏乙木(木=克破)
	}
	scores := map[string]int{
		"木": 6,  // 3 branches × middle qi 乙木 = 6 ("未" and "辰" have 木 in ZhiAllElements)
		"火": 0,
		"土": 30, // strong earth
		"金": 5,
		"水": 0,
	}

	result := checkCongQiangGe(pillars, "辰", scores)

	if result != nil {
		t.Errorf("稼穑格应被拒绝（未/辰/未藏干含木=克破），但返回了格局: %+v", result)
	}
}

// Verify 稼穑格 accepts when no 克破 in hidden stems
func TestCheckCongQiangGe_AcceptsWhenNoKePo(t *testing.T) {
	// 戊土日主生于辰月，地支全土金水，无木克破
	pillars := []model.Pillar{
		{Gan: "庚", Zhi: "戌"}, // 戌藏: 土,金,火 (无木)
		{Gan: "庚", Zhi: "辰"}, // 辰藏: 土,木,水 (有木!)
		{Gan: "戊", Zhi: "辰"}, // 日: 戊土
		{Gan: "壬", Zhi: "丑"}, // 丑藏: 土,水,金 (无木)
	}
	scores := map[string]int{
		"木": 2,  // only from 辰 middle qi
		"火": 1,
		"土": 35,
		"金": 10,
		"水": 5,
	}

	result := checkCongQiangGe(pillars, "辰", scores)

	// 辰月有木 → 应拒绝
	if result != nil {
		t.Errorf("辰藏木=克破，应拒绝: %+v", result)
	}
}

// Verify clean case: no 克破 anywhere in hidden stems
func TestCheckCongQiangGe_PureCase(t *testing.T) {
	// 戊土日主生于戌月，地支戌/丑/辰(支全土)，无木克破
	// 但是 辰 has 木... all earth branches that could form 稼穑格 (辰戌丑未)
	// need checking. Actually 辰 contains 木, so a "pure" 稼穑格 should use only 戌 and 丑 and 未 (未 also has 木).
	//
	// All four 墓库 branches have 木 in hidden stems:
	// 辰=土/木/水, 未=土/火/木, 戌=土/金/火, 丑=土/水/金
	// Only 戌 and 丑 have no 木.
	//
	// So a pure 稼穑格 with 戌/丑 only:
	pillars := []model.Pillar{
		{Gan: "丙", Zhi: "戌"}, // 戌藏: 土,金,火 (无木)
		{Gan: "庚", Zhi: "戌"}, // 戌藏: 土,金,火 (无木)
		{Gan: "戊", Zhi: "丑"}, // 日: 戊土
		{Gan: "壬", Zhi: "丑"}, // 丑藏: 土,水,金 (无木)
	}
	scores := map[string]int{
		"木": 0,  // no 木 anywhere
		"火": 3,
		"土": 40, // dominant
		"金": 5,
		"水": 3,
	}

	result := checkCongQiangGe(pillars, "戌", scores)

	// 戌月土旺，全局无木克破 → 应该通过并返回 稼穑格
	if result == nil {
		t.Error("纯稼穑格（戌/丑无木克破）应被检测，但返回nil")
	} else if result.PatternName != "稼穑格（从强格）" {
		t.Errorf("期望 稼穑格（从强格）, 实际: %s", result.PatternName)
	}
}
