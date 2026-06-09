package fortune_test

import (
	. "bazi/internal/service/fortune"
	"testing"
	"time"

	bazipkg "bazi/internal/service/bazi"
)

// ── 周运聚合测试 ──────────────────────────────────────────

func TestCalculateWeekly_Basic(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}

	// 标准八字
	bazi, err := baziSvc.Calculate(1990, 6, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatalf("计算八字失败: %v", err)
	}

	weekStart := time.Date(2025, 1, 6, 12, 0, 0, 0, time.UTC) // Monday
	result := engine.CalculateWeekly(bazi, weekStart, 1990)

	if result == nil {
		t.Fatal("CalculateWeekly 返回 nil")
	}

	// 验证周开始日期
	if result.WeekStart != "2025-01-06" {
		t.Errorf("WeekStart = %s, want 2025-01-06", result.WeekStart)
	}

	// 验证 7 天运势
	if len(result.DailyFortunes) != 7 {
		t.Fatalf("DailyFortunes 应有 7 天, got %d", len(result.DailyFortunes))
	}

	// 验证周评分在 [0, 100] 范围内
	if result.WeeklyScore < 0 || result.WeeklyScore > 100 {
		t.Errorf("WeeklyScore = %d, 不在 [0,100] 范围内", result.WeeklyScore)
	}

	// 验证汇总非空
	if result.OverallSummary == "" {
		t.Error("OverallSummary 为空")
	}

	// 验证每日评分
	for i, df := range result.DailyFortunes {
		if df.Score < 0 || df.Score > 100 {
			t.Errorf("DailyFortunes[%d].Score = %d, 超出 [0,100]", i, df.Score)
		}
		if df.Date == "" {
			t.Errorf("DailyFortunes[%d].Date 为空", i)
		}
	}

	// 验证趋势数据
	if len(result.ElementTrend) != 7 {
		t.Fatalf("ElementTrend 应有 7 项, got %d", len(result.ElementTrend))
	}

	// 验证周评分是 7 天平均分
	sum := 0
	for _, df := range result.DailyFortunes {
		sum += df.Score
	}
	avg := sum / 7
	if result.WeeklyScore != avg {
		t.Errorf("WeeklyScore = %d, 期望平均分 = %d", result.WeeklyScore, avg)
	}
}

func TestCalculateWeekly_ScoreRange(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}

	// 不同八字测试
	charts := []struct {
		y, m, d, h int
		startDate  string
	}{
		{1990, 6, 15, 8, "2025-01-06"},
		{1985, 3, 20, 14, "2025-06-01"},
		{2000, 1, 1, 0, "2025-12-01"},
	}

	for _, c := range charts {
		bazi, err := baziSvc.Calculate(c.y, c.m, c.d, c.h, 0, "MALE")
		if err != nil {
			t.Fatalf("计算八字失败 (%d-%d-%d): %v", c.y, c.m, c.d, err)
		}
		start, _ := time.Parse("2006-01-02", c.startDate)
		result := engine.CalculateWeekly(bazi, start, c.y)

		if result.WeeklyScore < 0 || result.WeeklyScore > 100 {
			t.Errorf("八字 %d-%d-%d 周评分=%d 超出 [0,100]", c.y, c.m, c.d, result.WeeklyScore)
		}
	}
}

func TestCalculateWeekly_DateContinuity(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}
	bazi, _ := baziSvc.Calculate(1990, 6, 15, 8, 0, "MALE")

	weekStart := time.Date(2025, 1, 6, 12, 0, 0, 0, time.UTC)
	result := engine.CalculateWeekly(bazi, weekStart, 1990)

	expectedDates := []string{
		"2025-01-06", "2025-01-07", "2025-01-08",
		"2025-01-09", "2025-01-10", "2025-01-11", "2025-01-12",
	}

	for i, expected := range expectedDates {
		if result.DailyFortunes[i].Date != expected {
			t.Errorf("Day %d: date = %s, want %s", i, result.DailyFortunes[i].Date, expected)
		}
	}
}

// ── 月运聚合测试 ──────────────────────────────────────────

func TestCalculateMonthly_Basic(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}

	bazi, err := baziSvc.Calculate(1990, 6, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatalf("计算八字失败: %v", err)
	}

	result := engine.CalculateMonthly(bazi, 2025, 1, 1990)

	if result == nil {
		t.Fatal("CalculateMonthly 返回 nil")
	}

	// 验证年月
	if result.Year != 2025 {
		t.Errorf("Year = %d, want 2025", result.Year)
	}
	if result.Month != 1 {
		t.Errorf("Month = %d, want 1", result.Month)
	}

	// 验证天数 (1月 = 31天)
	if len(result.DailyFortunes) != 31 {
		t.Fatalf("DailyFortunes 应有 31 天, got %d", len(result.DailyFortunes))
	}

	// 验证月评分在 [0, 100] 范围内
	if result.MonthlyScore < 0 || result.MonthlyScore > 100 {
		t.Errorf("MonthlyScore = %d, 不在 [0,100] 范围内", result.MonthlyScore)
	}

	// 验证汇总非空
	if result.OverallSummary == "" {
		t.Error("OverallSummary 为空")
	}

	// 验证每日评分和日期
	for i, df := range result.DailyFortunes {
		if df.Score < 0 || df.Score > 100 {
			t.Errorf("DailyFortunes[%d].Score = %d, 超出 [0,100]", i, df.Score)
		}
		if df.Date == "" {
			t.Errorf("DailyFortunes[%d].Date 为空", i)
		}
	}

	// 验证趋势数据
	if len(result.ElementTrend) != 31 {
		t.Fatalf("ElementTrend 应有 31 项, got %d", len(result.ElementTrend))
	}

	// 验证月评分是平均分
	sum := 0
	for _, df := range result.DailyFortunes {
		sum += df.Score
	}
	avg := sum / 31
	if result.MonthlyScore != avg {
		t.Errorf("MonthlyScore = %d, 期望平均分 = %d", result.MonthlyScore, avg)
	}
}

func TestCalculateMonthly_DateCorrectness(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}
	bazi, _ := baziSvc.Calculate(1990, 6, 15, 8, 0, "MALE")

	// 6月 = 30天
	result := engine.CalculateMonthly(bazi, 2025, 6, 1990)

	if len(result.DailyFortunes) != 30 {
		t.Errorf("6月应有 30 天, got %d", len(result.DailyFortunes))
	}
	if result.Year != 2025 || result.Month != 6 {
		t.Errorf("Year/Month = %d/%d, want 2025/6", result.Year, result.Month)
	}

	// 2月 = 28天 (2025 非闰年)
	result2 := engine.CalculateMonthly(bazi, 2025, 2, 1990)
	if len(result2.DailyFortunes) != 28 {
		t.Errorf("2月(非闰年)应有 28 天, got %d", len(result2.DailyFortunes))
	}

	// 2024年2月 = 29天 (闰年)
	result3 := engine.CalculateMonthly(bazi, 2024, 2, 1990)
	if len(result3.DailyFortunes) != 29 {
		t.Errorf("2月(闰年)应有 29 天, got %d", len(result3.DailyFortunes))
	}
}

func TestCalculateMonthly_ScoreRange(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}

	charts := []struct {
		y, m, d, h int
		queryY, queryM int
	}{
		{1990, 6, 15, 8, 2025, 1},
		{1985, 3, 20, 14, 2024, 6},
		{2000, 1, 1, 0, 2025, 12},
	}

	for _, c := range charts {
		bazi, err := baziSvc.Calculate(c.y, c.m, c.d, c.h, 0, "MALE")
		if err != nil {
			t.Fatalf("计算八字失败 (%d-%d-%d): %v", c.y, c.m, c.d, err)
		}
		result := engine.CalculateMonthly(bazi, c.queryY, c.queryM, c.y)
		if result.MonthlyScore < 0 || result.MonthlyScore > 100 {
			t.Errorf("八字 %d-%d-%d 月评分=%d 超出 [0,100]", c.y, c.m, c.d, result.MonthlyScore)
		}
	}
}

// TestCalculateWeekly_EdgeDates 边界日期测试（跨年、跨月）
func TestCalculateWeekly_EdgeDates(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}
	bazi, _ := baziSvc.Calculate(1990, 6, 15, 8, 0, "MALE")

	// 跨年周
	weekStart := time.Date(2025, 12, 29, 12, 0, 0, 0, time.UTC)
	result := engine.CalculateWeekly(bazi, weekStart, 1990)
	if result.WeekStart != "2025-12-29" {
		t.Errorf("WeekStart = %s, want 2025-12-29", result.WeekStart)
	}
	if len(result.DailyFortunes) != 7 {
		t.Errorf("应有 7 天, got %d", len(result.DailyFortunes))
	}
	// 验证日期序列包含跨年
	dates := []string{}
	for _, df := range result.DailyFortunes {
		dates = append(dates, df.Date)
	}
	t.Logf("跨年周日期: %v", dates)

	// 跨月周 (从月末到月初)
	weekStart2 := time.Date(2025, 1, 27, 12, 0, 0, 0, time.UTC)
	result2 := engine.CalculateWeekly(bazi, weekStart2, 1990)
	if len(result2.DailyFortunes) != 7 {
		t.Errorf("应有 7 天, got %d", len(result2.DailyFortunes))
	}
}
