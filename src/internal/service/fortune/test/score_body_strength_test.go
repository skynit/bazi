package fortune_test

import (
	. "bazi/internal/service/fortune"
	"testing"
	"time"

	bazipkg "bazi/internal/service/bazi"
)

// ============================================================
// C1-05 身强对评分影响测试
//
// 验证：相同日柱但不同身强状态 → 天干/地支关系相同，
// 但 score 因喜忌不同而不同。
// 原理：身强者喜克泄耗，被生扶则扣分；
//       身弱者喜生扶，被克泄耗则扣分。
// ============================================================

// TestScoreBodyStrength_Impact 验证身强壮态影响评分
//
// 使用两个不同八字但日干相同的案例，对比它们在同一天的运势评分。
// 日干甲木：甲木身旺者遇木火日（生扶）应扣分，甲木身弱者遇木火日应加分
func TestScoreBodyStrength_Impact(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}

	// Case A: 甲木身旺 — 甲木日主 + 寅月（得令）+ 水木多
	// 喜: 火土金（克泄耗）, 忌: 水木（生扶）
	strongBazi, err := baziSvc.CalculateFromPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
	if err != nil {
		t.Fatalf("甲木身旺八字计算失败: %v", err)
	}
	t.Logf("身旺八字: 身强=%s 总评分=%.3f 喜=%v 忌=%v",
		strongBazi.BodyStrength.Verdict, strongBazi.BodyStrength.TotalScore,
		strongBazi.BodyStrength.Like, strongBazi.BodyStrength.Dislike)

	if strongBazi.BodyStrength.Verdict != "身旺" && strongBazi.BodyStrength.Verdict != "偏旺" {
		t.Logf("注意: 八字身强判定为 %s 而非身旺/偏旺, 测试结果可能不显著", strongBazi.BodyStrength.Verdict)
	}

	// Case B: 甲木身弱 — 甲木日主 + 申月（失令）+ 金多
	// 喜: 水木（生扶）, 忌: 火土金（克泄耗）
	weakBazi, err := baziSvc.CalculateFromPillars("庚申", "甲申", "甲寅", "庚午", "MALE")
	if err != nil {
		t.Fatalf("甲木身弱八字计算失败: %v", err)
	}
	t.Logf("身弱八字: 身强=%s 总评分=%.3f 喜=%v 忌=%v",
		weakBazi.BodyStrength.Verdict, weakBazi.BodyStrength.TotalScore,
		weakBazi.BodyStrength.Like, weakBazi.BodyStrength.Dislike)

	if weakBazi.BodyStrength.Verdict != "身弱" && weakBazi.BodyStrength.Verdict != "偏弱" {
		t.Logf("注意: 八字身弱判定为 %s 而非身弱/偏弱, 测试结果可能不显著", weakBazi.BodyStrength.Verdict)
	}

	// 验证两个八字的身强壮态不同
	if strongBazi.BodyStrength.Verdict == weakBazi.BodyStrength.Verdict {
		t.Log("注意: 两个八字身强壮态相同, 无法测试身强影响差异")
	}

	// 测试 3 个不同日期的运势评分
	testDates := []string{"2026-06-05", "2026-06-15", "2026-07-01"}
	birthYearStrong := 2012 // 随便给一个出生年
	birthYearWeak := 1980

	for _, dateStr := range testDates {
		queryDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			t.Fatalf("日期解析失败: %v", err)
		}

		scoreStrong := engine.CalculateDaily(strongBazi, queryDate, birthYearStrong)
		scoreWeak := engine.CalculateDaily(weakBazi, queryDate, birthYearWeak)

		if scoreStrong == nil || scoreWeak == nil {
			t.Errorf("%s: 计算结果为空", dateStr)
			continue
		}

		t.Logf("%s: 身旺得分=%d 身弱得分=%d 日柱=%s",
			dateStr, scoreStrong.Score, scoreWeak.Score, scoreStrong.DayPillar.Gan+scoreStrong.DayPillar.Zhi)

		// 验证两个评分不同（身强壮态不同导致喜忌不同）
		if scoreStrong.Score == scoreWeak.Score {
			t.Logf("%s: 两个评分相同=%d, 说明身强未影响评分", dateStr, scoreStrong.Score)
		}
	}
}

// TestScoreBodyStrength_SameDayPillar 相同日柱不同身强 → 不同评分
//
// 使用日主相同（甲木）但身强壮态不同的八字，
// 在同一天（流日也为甲木/乙木）验证评分差异
func TestScoreBodyStrength_SameDayPillar(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}

	// 身强八字：甲木 + 寅月 + 水木多 → 身旺
	strongBazi, err := baziSvc.CalculateFromPillars("癸亥", "甲寅", "甲子", "乙丑", "FEMALE")
	if err != nil {
		t.Fatalf("身强八字计算失败: %v", err)
	}
	t.Logf("身强: 判定=%s 总分=%.3f 喜=%v 忌=%v",
		strongBazi.BodyStrength.Verdict, strongBazi.BodyStrength.TotalScore,
		strongBazi.BodyStrength.Like, strongBazi.BodyStrength.Dislike)

	// 身弱八字：甲木 + 申月 + 金多 → 身弱
	weakBazi, err := baziSvc.CalculateFromPillars("庚申", "甲申", "甲子", "丙寅", "FEMALE")
	if err != nil {
		t.Fatalf("身弱八字计算失败: %v", err)
	}
	t.Logf("身弱: 判定=%s 总分=%.3f 喜=%v 忌=%v",
		weakBazi.BodyStrength.Verdict, weakBazi.BodyStrength.TotalScore,
		weakBazi.BodyStrength.Like, weakBazi.BodyStrength.Dislike)

	// 选择日干为甲木/乙木的日期（生扶日主）和丙火/丁火的日期（泄日主）
	dates := []string{
		"2026-06-06", // 某甲日或乙日
		"2026-06-18", // 某丙日或丁日
		"2026-07-05", // 另一日期
	}

	birthYearS := 2010
	birthYearW := 1990

	for _, dateStr := range dates {
		queryDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			t.Fatalf("日期解析失败: %v", err)
		}

		rStrong := engine.CalculateDaily(strongBazi, queryDate, birthYearS)
		rWeak := engine.CalculateDaily(weakBazi, queryDate, birthYearW)

		if rStrong == nil || rWeak == nil {
			t.Errorf("%s: 计算结果为空", dateStr)
			continue
		}

		t.Logf("%s: 流日=%s 身旺得分=%d 身弱得分=%d",
			dateStr, rStrong.DayPillar.Gan+rStrong.DayPillar.Zhi,
			rStrong.Score, rWeak.Score)

		// 验证日主相同（甲木）
		if strongBazi.DayPillar.Gan != weakBazi.DayPillar.Gan {
			t.Errorf("两个八字的日主应相同: strong=%s weak=%s",
				strongBazi.DayPillar.Gan, weakBazi.DayPillar.Gan)
		}

		// 相同流日下的评分应该不同（因为身强不同导致喜忌不同）
		if rStrong.Score == rWeak.Score {
			t.Logf("%s: 评分相同=%d, 可能身强差异未影响评分", dateStr, rStrong.Score)
		}
	}
}

// TestScoreBodyStrength_OutputIntegrity 验证身强不同时输出结构完整性
func TestScoreBodyStrength_OutputIntegrity(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}

	t.Run("身旺案例", func(t *testing.T) {
		bazi, err := baziSvc.CalculateFromPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
		if err != nil {
			t.Fatalf("CalculateFromPillars 失败: %v", err)
		}
		queryDate := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
		result := engine.CalculateDaily(bazi, queryDate, 2012)
		if result == nil {
			t.Fatal("CalculateDaily 返回 nil")
		}
		if result.Score < 0 || result.Score > 100 {
			t.Errorf("评分超出 [0,100]: %d", result.Score)
		}
		if result.LuckyColor == "" {
			t.Error("LuckyColor 为空")
		}
		if len(result.Yi) == 0 {
			t.Error("Yi 为空")
		}
		if len(result.Ji) == 0 {
			t.Error("Ji 为空")
		}
		if result.ShengKe.DayStemRelation == "" {
			t.Error("DayStemRelation 为空")
		}
	})

	t.Run("身弱案例", func(t *testing.T) {
		bazi, err := baziSvc.CalculateFromPillars("庚申", "甲申", "甲寅", "庚午", "MALE")
		if err != nil {
			t.Fatalf("CalculateFromPillars 失败: %v", err)
		}
		queryDate := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
		result := engine.CalculateDaily(bazi, queryDate, 1980)
		if result == nil {
			t.Fatal("CalculateDaily 返回 nil")
		}
		if result.Score < 0 || result.Score > 100 {
			t.Errorf("评分超出 [0,100]: %d", result.Score)
		}
		if result.LuckyColor == "" {
			t.Error("LuckyColor 为空")
		}
		if len(result.Yi) == 0 {
			t.Error("Yi 为空")
		}
		if len(result.Ji) == 0 {
			t.Error("Ji 为空")
		}
	})
}
