package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"
)

// ============================================================
// 后验修正测试
//
// 验证 calcBodyStrengthV2 中的失令不衰修正逻辑:
//   - 甲木日主+申月（失令）+多水(壬子/癸亥)→ 修正连续分
// 最终五档候选必须始终由修正后的总分重新计算；触发修正不保证跨档。
// ============================================================

func expectedBodyStrengthBand(score float64) string {
	switch {
	case score > 0.7:
		return "身旺"
	case score > 0.5:
		return "偏旺"
	case score > 0.4:
		return "中和"
	case score > 0.3:
		return "偏弱"
	default:
		return "身弱"
	}
}

func assertAdjustmentAndFinalBand(t *testing.T, analysis BodyStrengthResult, ruleID string) {
	t.Helper()
	if len(analysis.Adjustments) != 1 || analysis.Adjustments[0].RuleID != ruleID {
		t.Fatalf("修正证据=%+v，期望唯一规则 %s", analysis.Adjustments, ruleID)
	}
	if want := expectedBodyStrengthBand(analysis.TotalScore); analysis.ScoreBandCandidate != want {
		t.Fatalf("修正后总分 %.6f 应映射为 %s，实际 %s", analysis.TotalScore, want, analysis.ScoreBandCandidate)
	}
}

// TestBodyStrengthRefinement_ShiLingBuShuai 失令不衰
//
// 甲木日主 + 申月（失令，lingScore≈0.5）+ 多水（壬子/癸亥构成强力生扶）
// 期望：后验修正触发；是否跨档由修正后的最终总分决定。
func TestBodyStrengthRefinement_ShiLingBuShuai(t *testing.T) {
	svc := &BaziService{}

	// case 1: 甲木日主 + 申月（失令）+ 壬子年柱 + 癸亥时柱
	// 四柱: 壬子(水) 戊申(金水) 甲午(木火) 癸亥(水)
	// 分析: 甲木生申月失令，但壬子癸亥强水印星生扶，应升级
	t.Run("甲木申月重水", func(t *testing.T) {
		result, err := svc.CalculateSyntheticPillars("壬子", "戊申", "甲午", "癸亥", "MALE")
		if err != nil {
			t.Fatalf("CalculateSyntheticPillars 失败: %v", err)
		}

		verdict := result.BodyStrength.ScoreBandCandidate
		t.Logf("甲木申月重水: 身强判定=%s 总评分=%.3f 得令=%.2f 得地=%.2f 得势=%.2f 得生=%.2f",
			verdict, result.BodyStrength.TotalScore,
			result.BodyStrength.LingScore, result.BodyStrength.DiScore,
			result.BodyStrength.ShiScore, result.BodyStrength.ShengScore)

		// 失令
		if result.BodyStrength.LingScore > 1.0 {
			t.Errorf("甲木在申月失令分应<=1.0, 实际 %.2f", result.BodyStrength.LingScore)
		}
		assertAdjustmentAndFinalBand(t, result.BodyStrength, "bazi.body-strength.adjustment.shi-ling-bu-shuai.v1")
	})

	// case 2: 甲木日主 + 申月 + 双水透干
	t.Run("甲木申月双水", func(t *testing.T) {
		result, err := svc.CalculateSyntheticPillars("辛亥", "丙申", "甲午", "癸巳", "MALE")
		if err != nil {
			t.Fatalf("CalculateSyntheticPillars 失败: %v", err)
		}

		verdict := result.BodyStrength.ScoreBandCandidate
		t.Logf("甲木申月双水: 身强判定=%s 总评分=%.3f 得生=%.2f",
			verdict, result.BodyStrength.TotalScore, result.BodyStrength.ShengScore)

		if result.BodyStrength.LingScore > 1.0 {
			t.Errorf("甲木在申月失令分应<=1.0, 实际 %.2f", result.BodyStrength.LingScore)
		}
		assertAdjustmentAndFinalBand(t, result.BodyStrength, "bazi.body-strength.adjustment.shi-ling-bu-shuai.v1")
	})
}

// TestBodyStrengthRefinement_NormalCases 正常情况不应触发修正
func TestBodyStrengthRefinement_NormalCases(t *testing.T) {
	svc := &BaziService{}

	// 甲木 + 寅月 + 无助亦无克 → 正常得令身旺
	t.Run("甲木寅月正常", func(t *testing.T) {
		result, err := svc.CalculateSyntheticPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
		if err != nil {
			t.Fatalf("CalculateSyntheticPillars 失败: %v", err)
		}
		verdict := result.BodyStrength.ScoreBandCandidate
		t.Logf("甲木寅月正常(壬辰壬寅甲寅庚午): 身强=%s 总分=%.3f",
			verdict, result.BodyStrength.TotalScore)
		if verdict != "身旺" && verdict != "偏旺" {
			t.Errorf("甲木寅月双壬水应身旺/偏旺, 实际 %s", verdict)
		}
	})

	// 甲木 + 申月 + 无助 → 正常失令身弱
	t.Run("甲木申月正常", func(t *testing.T) {
		result, err := svc.CalculateSyntheticPillars("庚申", "甲申", "甲寅", "庚午", "MALE")
		if err != nil {
			t.Fatalf("CalculateSyntheticPillars 失败: %v", err)
		}
		verdict := result.BodyStrength.ScoreBandCandidate
		t.Logf("甲木申月正常(庚申甲申甲寅庚午): 身强=%s 总分=%.3f 得令=%.2f",
			verdict, result.BodyStrength.TotalScore, result.BodyStrength.LingScore)
		// 申月失令 + 无生扶 + 多金 → 身弱或偏弱
		if verdict == "身旺" || verdict == "偏旺" {
			t.Errorf("甲木申月多金应为身弱/偏弱, 实际 %s", verdict)
		}
	})
}
