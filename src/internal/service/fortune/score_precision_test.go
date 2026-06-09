package fortune

import (
	"testing"
	"time"

	bazipkg "bazi/internal/service/bazi"
)

// ── calcScore 评分边界测试 ──────────────────────────────────
// 规范：score ∈ [0, 100]
// 基准分 50，天干关系 +/-18, +/-10, +/-8, +/-5
// 地支关系 -30(冲), -20(刑), -15(害), -10(破), +8(合), +15(三合), +20(三会)
// 天干五合 +12

func TestCalcScore_Default(t *testing.T) {
	// neutral 天干 + neutral 地支 = 50
	got := calcScore("unknown", "neutral", "", "")
	if got != 50 {
		t.Errorf("默认分应为 50, got %d", got)
	}
}

func TestCalcScore_StemRelations_Baseline(t *testing.T) {
	// 各天干关系在 neutral 地支下的基础得/扣分
	tests := []struct {
		stemRel string
		want    int // baseline 50 +/- adjustment
	}{
		{"same", 60},    // 50+10
		{"shengWo", 68}, // 50+18
		{"woSheng", 55}, // 50+5
		{"keWo", 32},    // 50-18
		{"woKe", 58},    // 50+8
	}
	for _, tt := range tests {
		got := calcScore(tt.stemRel, "neutral", "", "")
		if got != tt.want {
			t.Errorf("calcScore(%q, neutral) = %d, want %d", tt.stemRel, got, tt.want)
		}
	}
}

func TestCalcScore_BranchRelations_Baseline(t *testing.T) {
	// 各地支关系在 same 天干下的基础得/扣分
	tests := []struct {
		branchRel string
		want      int
	}{
		{"clash", 30},    // 50-30+10=30
		{"harm", 45},     // 50-15+10=45
		{"punish", 40},   // 50-20+10=40
		{"break", 50},    // 50-10+10=50
		{"combine", 68},  // 50+8+10=68
		{"sanHe", 75},    // 50+15+10=75
		{"sanHui", 80},   // 50+20+10=80
	}
	for _, tt := range tests {
		got := calcScore("same", tt.branchRel, "", "")
		if got != tt.want {
			t.Errorf("calcScore(same, %q) = %d, want %d", tt.branchRel, got, tt.want)
		}
	}
}

// TestCalcScore_Boundary_Max 测试最大可能分数（不超过100）
func TestCalcScore_Boundary_Max(t *testing.T) {
	// sanHui(+20) + shengWo(+18) + 五合(+12) = 50+20+18+12=100
	got := calcScore("shengWo", "sanHui", "甲", "己")
	if got > 100 {
		t.Errorf("分数不应超过100, got %d", got)
	}
	if got != 100 {
		t.Errorf("shengWo+sanHui+五合 应得 100, got %d", got)
	}
}

// TestCalcScore_Boundary_Min 测试最小可能分数（不低于0）
func TestCalcScore_Boundary_Min(t *testing.T) {
	// clash(-30) + keWo(-18) = 50-30-18=2 (>=0)
	got := calcScore("keWo", "clash", "", "")
	if got < 0 {
		t.Errorf("分数不应低于0, got %d", got)
	}
	if got != 2 {
		t.Errorf("keWo+clash 应得 2, got %d", got)
	}

	// 更极端的组合: keWo(-18) + 无五合 + punish(-20) = 12 (still >= 0)
	got2 := calcScore("keWo", "punish", "", "")
	if got2 < 0 {
		t.Errorf("分数不应低于0, got %d", got2)
	}
}

// TestCalcScore_Boundary_Floor 测试地板值（多个扣分叠加）
func TestCalcScore_Boundary_Floor(t *testing.T) {
	// keWo(-18) + clash(-30) = 2, 不会低于0
	got := calcScore("keWo", "clash", "", "")
	if got != 2 {
		t.Errorf("keWo+clash 应 clamp 为 2, got %d", got)
	}
}

// TestCalcScore_Boundary_Ceiling 测试天花板值（多个加分叠加）
func TestCalcScore_Boundary_Ceiling(t *testing.T) {
	// sanHui(+20) + shengWo(+18) = 88, 加五合(+12)=100
	got := calcScore("shengWo", "sanHui", "甲", "己")
	if got != 100 {
		t.Errorf("shengWo+sanHui+五合 应 clamp 为 100, got %d", got)
	}
}

// TestCalcScore_GanHe_Effect 验证天干五合加分
func TestCalcScore_GanHe_Effect(t *testing.T) {
	tests := []struct {
		name   string
		userG  string
		dayG   string
		want   int // 五合+12
	}{
		{"甲己合", "甲", "己", 62}, // 50+12=62
		{"乙庚合", "乙", "庚", 62}, // 50+12=62
		{"丙辛合", "丙", "辛", 62}, // 50+12=62
		{"丁壬合", "丁", "壬", 62}, // 50+12=62
		{"戊癸合", "戊", "癸", 62}, // 50+12=62
		{"己甲合", "己", "甲", 62}, // 50+12=62
	}
	for _, tt := range tests {
		got := calcScore("unknown", "neutral", tt.userG, tt.dayG)
		if got != tt.want {
			t.Errorf("%s: calcScore with five-combine = %d, want %d", tt.name, got, tt.want)
		}
	}
}

// TestCalcScore_ExtremeCombos 极端组合测试
func TestCalcScore_ExtremeCombos(t *testing.T) {
	combo := []struct {
		name     string
		stemRel  string
		branchRel string
		userGan  string
		dayGan   string
		wantMin  int
		wantMax  int
	}{
		{"全凶(keWo+clash)", "keWo", "clash", "", "", 0, 10},
		{"全凶(keWo+punish)", "keWo", "punish", "", "", 0, 15},
		{"全凶(keWo+harm)", "keWo", "harm", "", "", 0, 20},
		{"全吉(shengWo+sanHui+ganHe)", "shengWo", "sanHui", "甲", "己", 95, 100},
		{"全吉(shengWo+sanHe+ganHe)", "shengWo", "sanHe", "甲", "己", 90, 100},
		{"全吉(same+combine)", "same", "combine", "", "", 65, 70},
	}
	for _, c := range combo {
		got := calcScore(c.stemRel, c.branchRel, c.userGan, c.dayGan)
		if got < c.wantMin || got > c.wantMax {
			t.Errorf("[%s] calcScore = %d, want in [%d, %d]", c.name, got, c.wantMin, c.wantMax)
		}
	}
}

// TestCalculateDaily_OutputStructure 验证 DailyFortune 输出结构完整性
func TestCalculateDaily_OutputStructure(t *testing.T) {
	engine := NewFortuneEngine()

	// 使用一个标准八字
	baziSvc := &bazipkg.BaziService{}
	userBazi, err := baziSvc.Calculate(1990, 6, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatalf("计算八字失败: %v", err)
	}

	queryDate := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	result := engine.CalculateDaily(userBazi, queryDate, 1990)

	if result == nil {
		t.Fatal("CalculateDaily 返回 nil")
	}

	// 验证所有必填字段非空
	if result.Date == "" {
		t.Error("Date 为空")
	}
	if result.DayPillar.Gan == "" || result.DayPillar.Zhi == "" {
		t.Error("DayPillar 不完整")
	}
	if result.Score < 0 || result.Score > 100 {
		t.Errorf("Score = %d, 不在 [0,100] 范围内", result.Score)
	}
	if result.LuckyColor == "" {
		t.Error("LuckyColor 为空")
	}
	if len(result.LuckyNumbers) == 0 {
		t.Error("LuckyNumbers 为空")
	}
	if result.WealthDir == "" {
		t.Error("WealthDir 为空")
	}
	if result.ClashZodiac == "" {
		t.Error("ClashZodiac 为空")
	}
	if len(result.AuspiciousHours) == 0 {
		t.Error("AuspiciousHours 为空")
	}
	if len(result.Yi) == 0 {
		t.Error("Yi 为空")
	}
	if len(result.Ji) == 0 {
		t.Error("Ji 为空")
	}
	if result.ShengKe.DayStemRelation == "" {
		t.Error("ShengKe.DayStemRelation 为空")
	}
	if result.ShengKe.Summary == "" {
		t.Error("ShengKe.Summary 为空")
	}
	if len(result.ElementImages) == 0 {
		t.Error("ElementImages 为空")
	}
	if len(result.TodayElements) == 0 {
		t.Error("TodayElements 为空")
	}
	if result.FlowImpact == "" {
		t.Error("FlowImpact 为空")
	}
	if result.Rikuyo == nil {
		t.Error("Rikuyo 为 nil")
	}
}

// TestCalculateDaily_Score_Range 多次运行验证分数边界
func TestCalculateDaily_Score_Range(t *testing.T) {
	engine := NewFortuneEngine()
	baziSvc := &bazipkg.BaziService{}

	// 不同日主在不同日期的运势
	type dateEntry struct {
		date   string
		birthY int
		birthM int
		birthD int
		birthH int
	}

	entries := []dateEntry{
		{birthY: 1990, birthM: 6, birthD: 15, birthH: 8, date: "2025-01-15"},
		{birthY: 1985, birthM: 3, birthD: 20, birthH: 14, date: "2025-06-01"},
		{birthY: 2000, birthM: 1, birthD: 1, birthH: 0, date: "2025-12-25"},
		{birthY: 1995, birthM: 8, birthD: 10, birthH: 10, date: "2025-03-08"},
	}

	for _, e := range entries {
		bazi, err := baziSvc.Calculate(e.birthY, e.birthM, e.birthD, e.birthH, 0, "MALE")
		if err != nil {
			t.Fatalf("计算八字失败 (%d-%d-%d): %v", e.birthY, e.birthM, e.birthD, err)
		}
		qDate, _ := time.Parse("2006-01-02", e.date)
		result := engine.CalculateDaily(bazi, qDate, e.birthY)

		if result.Score < 0 || result.Score > 100 {
			t.Errorf("出生 %d-%d-%d 在 %s 的 score=%d 超出 [0,100]",
				e.birthY, e.birthM, e.birthD, e.date, result.Score)
		}
	}
}
