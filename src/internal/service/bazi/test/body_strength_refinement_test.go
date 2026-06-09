package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"
)

// ============================================================
// A3-06~07 后验修正测试
//
// 验证 calcBodyStrengthV2 中的后验修正逻辑:
//   - 得令不旺: 甲木日主+寅月（得令）+多金(庚申/辛酉)→ 被金克，应 downgrade
//   - 失令不衰: 甲木日主+申月（失令）+多水(壬子/癸亥)→ 被水生扶，应 upgrade
// ============================================================

// TestBodyStrengthRefinement_DeLingBuWang 得令不旺
//
// 甲木日主 + 寅月（得令，lingScore=3.0）+ 多金（庚申/辛酉构成强力克制）
// 期望：后验修正触发，身旺→偏旺 或 偏旺→中和
func TestBodyStrengthRefinement_DeLingBuWang(t *testing.T) {
	svc := &BaziService{}

	// case 1: 甲木日主 + 寅月（得令）+ 庚申年柱 + 辛酉时柱
	// 四柱: 庚申(金) 丙寅(木) 甲午(木火) 辛酉(金)
	// 分析: 甲木生寅月得令，但庚申强金 + 辛酉强金克木，应降级
	t.Run("甲木寅月重金", func(t *testing.T) {
		result, err := svc.CalculateFromPillars("庚申", "丙寅", "甲午", "辛酉", "MALE")
		if err != nil {
			t.Fatalf("CalculateFromPillars 失败: %v", err)
		}

		verdict := result.BodyStrength.Verdict
		t.Logf("甲木寅月重金: 身强判定=%s 总评分=%.3f 得令=%.2f 得地=%.2f 得势=%.2f 得生=%.2f",
			verdict, result.BodyStrength.TotalScore,
			result.BodyStrength.LingScore, result.BodyStrength.DiScore,
			result.BodyStrength.ShiScore, result.BodyStrength.ShengScore)

		// 得令应有高分，但被众多金克制应触发降级
		if result.BodyStrength.LingScore < 2.0 {
			t.Errorf("甲木在寅月得令分应>=2.0, 实际 %.2f", result.BodyStrength.LingScore)
		}
		// 应该不是身旺（被严重克制）
		if verdict == "身旺" {
			t.Errorf("甲木寅月加多金不应为身旺, 期望偏旺/中和/偏弱, 实际 %s", verdict)
		}
	})

	// case 2: 甲木日主 + 寅月 + 三金透干 庚申年 + 庚辰时
	t.Run("甲木寅月三金", func(t *testing.T) {
		result, err := svc.CalculateFromPillars("庚申", "丙寅", "甲午", "庚辰", "MALE")
		if err != nil {
			t.Fatalf("CalculateFromPillars 失败: %v", err)
		}

		verdict := result.BodyStrength.Verdict
		t.Logf("甲木寅月三金: 身强判定=%s 总评分=%.3f", verdict, result.BodyStrength.TotalScore)

		if result.BodyStrength.LingScore < 2.0 {
			t.Errorf("甲木在寅月得令分应>=2.0, 实际 %.2f", result.BodyStrength.LingScore)
		}
		if verdict == "身旺" {
			t.Errorf("甲木寅月加三金不应为身旺, 实际 %s", verdict)
		}
	})
}

// TestBodyStrengthRefinement_ShiLingBuShuai 失令不衰
//
// 甲木日主 + 申月（失令，lingScore≈0.5）+ 多水（壬子/癸亥构成强力生扶）
// 期望：后验修正触发，身弱→偏弱 或 偏弱→中和
func TestBodyStrengthRefinement_ShiLingBuShuai(t *testing.T) {
	svc := &BaziService{}

	// case 1: 甲木日主 + 申月（失令）+ 壬子年柱 + 癸亥时柱
	// 四柱: 壬子(水) 戊申(金水) 甲午(木火) 癸亥(水)
	// 分析: 甲木生申月失令，但壬子癸亥强水印星生扶，应升级
	t.Run("甲木申月重水", func(t *testing.T) {
		result, err := svc.CalculateFromPillars("壬子", "戊申", "甲午", "癸亥", "MALE")
		if err != nil {
			t.Fatalf("CalculateFromPillars 失败: %v", err)
		}

		verdict := result.BodyStrength.Verdict
		t.Logf("甲木申月重水: 身强判定=%s 总评分=%.3f 得令=%.2f 得地=%.2f 得势=%.2f 得生=%.2f",
			verdict, result.BodyStrength.TotalScore,
			result.BodyStrength.LingScore, result.BodyStrength.DiScore,
			result.BodyStrength.ShiScore, result.BodyStrength.ShengScore)

		// 失令
		if result.BodyStrength.LingScore > 1.0 {
			t.Errorf("甲木在申月失令分应<=1.0, 实际 %.2f", result.BodyStrength.LingScore)
		}
		// 不应为身弱（被强力生扶）
		if verdict == "身弱" {
			t.Errorf("甲木申月加多水不应为身弱, 期望偏弱/中和/偏旺, 实际 %s", verdict)
		}
	})

	// case 2: 甲木日主 + 申月 + 双水透干
	t.Run("甲木申月双水", func(t *testing.T) {
		result, err := svc.CalculateFromPillars("辛亥", "丙申", "甲午", "癸巳", "MALE")
		if err != nil {
			t.Fatalf("CalculateFromPillars 失败: %v", err)
		}

		verdict := result.BodyStrength.Verdict
		t.Logf("甲木申月双水: 身强判定=%s 总评分=%.3f 得生=%.2f",
			verdict, result.BodyStrength.TotalScore, result.BodyStrength.ShengScore)

		if result.BodyStrength.LingScore > 1.0 {
			t.Errorf("甲木在申月失令分应<=1.0, 实际 %.2f", result.BodyStrength.LingScore)
		}
		if verdict == "身弱" {
			t.Errorf("甲木申月加双水不应为身弱, 实际 %s", verdict)
		}
	})
}

// TestBodyStrengthRefinement_NormalCases 正常情况不应触发修正
func TestBodyStrengthRefinement_NormalCases(t *testing.T) {
	svc := &BaziService{}

	// 甲木 + 寅月 + 无助亦无克 → 正常得令身旺
	t.Run("甲木寅月正常", func(t *testing.T) {
		result, err := svc.CalculateFromPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
		if err != nil {
			t.Fatalf("CalculateFromPillars 失败: %v", err)
		}
		verdict := result.BodyStrength.Verdict
		t.Logf("甲木寅月正常(壬辰壬寅甲寅庚午): 身强=%s 总分=%.3f",
			verdict, result.BodyStrength.TotalScore)
		if verdict != "身旺" && verdict != "偏旺" {
			t.Errorf("甲木寅月双壬水应身旺/偏旺, 实际 %s", verdict)
		}
	})

	// 甲木 + 申月 + 无助 → 正常失令身弱
	t.Run("甲木申月正常", func(t *testing.T) {
		result, err := svc.CalculateFromPillars("庚申", "甲申", "甲寅", "庚午", "MALE")
		if err != nil {
			t.Fatalf("CalculateFromPillars 失败: %v", err)
		}
		verdict := result.BodyStrength.Verdict
		t.Logf("甲木申月正常(庚申甲申甲寅庚午): 身强=%s 总分=%.3f 得令=%.2f",
			verdict, result.BodyStrength.TotalScore, result.BodyStrength.LingScore)
		// 申月失令 + 无生扶 + 多金 → 身弱或偏弱
		if verdict == "身旺" || verdict == "偏旺" {
			t.Errorf("甲木申月多金应为身弱/偏弱, 实际 %s", verdict)
		}
	})
}
