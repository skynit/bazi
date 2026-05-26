package service

import (
	"testing"

	"bazi/internal/model"
)

func TestAnalyzePatternExtendedPrioritizesHuaQiGe(t *testing.T) {
	// 丁壬合化木 + 寅月旺地（真化气经典盘）
	pillars := []model.Pillar{
		{Gan: "丁", Zhi: "卯"},
		{Gan: "壬", Zhi: "寅"},
		{Gan: "丁", Zhi: "卯"},
		{Gan: "甲", Zhi: "辰"},
	}
	scores := map[string]int{"木": 35, "火": 10, "土": 15, "金": 0, "水": 40}

	got := analyzePatternExtended(pillars, "寅", scores, BodyStrengthResult{Verdict: "身旺", Like: []string{"木", "水"}, Dislike: []string{"金", "火"}})
	if got.PatternName != "化气格（木）" {
		t.Fatalf("PatternName = %q, want 化气格（木）", got.PatternName)
	}
	if got.PatternType != "特殊格局" {
		t.Fatalf("PatternType = %q, want 特殊格局", got.PatternType)
	}
	if got.SubType != "木" {
		t.Fatalf("SubType = %q, want 木", got.SubType)
	}
}

func TestCheckHuaQiGeRequiresMonthlySupport(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "己", Zhi: "酉"},
		{Gan: "庚", Zhi: "申"},
		{Gan: "甲", Zhi: "子"},
		{Gan: "壬", Zhi: "亥"},
	}
	scores := map[string]int{"木": 12, "火": 0, "土": 20, "金": 40, "水": 30}

	if got := checkHuaQiGe(pillars, "申", scores); got != nil {
		t.Fatalf("checkHuaQiGe returned %+v, want nil when month does not support hua qi", got)
	}
}

func TestCheckCongQiangGe(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "甲", Zhi: "寅"},
		{Gan: "乙", Zhi: "卯"},
		{Gan: "甲", Zhi: "寅"},
		{Gan: "癸", Zhi: "亥"},
	}
	scores := map[string]int{"木": 80, "火": 8, "土": 4, "金": 2, "水": 20}

	got := checkCongQiangGe(pillars, "卯", scores)
	if got == nil {
		t.Fatal("checkCongQiangGe returned nil")
	}
	if got.PatternName != "曲直格（从强格）" {
		t.Fatalf("PatternName = %q, want 曲直格（从强格）", got.PatternName)
	}
}

func TestCheckLiangShenChengXiang(t *testing.T) {
	scores := map[string]int{"木": 45, "火": 55, "土": 0, "金": 0, "水": 0}

	got := checkLiangShenChengXiang(scores)
	if got == nil {
		t.Fatal("checkLiangShenChengXiang returned nil")
	}
	if got.PatternName != "两神成像格（木生火）" {
		t.Fatalf("PatternName = %q, want 两神成像格（木生火）", got.PatternName)
	}
}

func TestCheckKuiGangAndRiDeGe(t *testing.T) {
	if got := checkKuiGangGe("庚", "辰"); got == nil || got.PatternName != "魁罡格" {
		t.Fatalf("checkKuiGangGe = %+v, want 魁罡格", got)
	}
	if got := checkRiDeGe("甲", "寅"); got == nil || got.PatternName != "日德格" {
		t.Fatalf("checkRiDeGe = %+v, want 日德格", got)
	}
}

func TestCheckJianLuYueRen(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "癸", Zhi: "亥"},
		{Gan: "戊", Zhi: "寅"},
		{Gan: "甲", Zhi: "子"},
		{Gan: "乙", Zhi: "丑"},
	}
	if got := checkJianLuYueRen(pillars, "寅"); got == nil || got.PatternName != "建禄格" {
		t.Fatalf("checkJianLuYueRen 建禄 = %+v", got)
	}
	if got := checkJianLuYueRen(pillars, "卯"); got == nil || got.PatternName != "月刃格" {
		t.Fatalf("checkJianLuYueRen 月刃 = %+v", got)
	}
}

func TestCheckSanQiGe(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "甲", Zhi: "子"},
		{Gan: "戊", Zhi: "寅"},
		{Gan: "庚", Zhi: "辰"},
		{Gan: "乙", Zhi: "午"},
	}
	got := checkSanQiGe(pillars)
	if got == nil || got.PatternName != "天三奇格" {
		t.Fatalf("checkSanQiGe = %+v, want 天三奇格", got)
	}
}

func TestCheckCongRuoGe(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "庚", Zhi: "申"},
		{Gan: "辛", Zhi: "酉"},
		{Gan: "甲", Zhi: "午"},
		{Gan: "丁", Zhi: "巳"},
	}
	scores := map[string]int{"木": 2, "火": 35, "土": 20, "金": 45, "水": 0}

	got := checkCongRuoGe(pillars, scores)
	if got == nil {
		t.Fatal("checkCongRuoGe returned nil")
	}
	if got.PatternName != "从弱格" {
		t.Fatalf("PatternName = %q, want 从弱格", got.PatternName)
	}
}

func TestCheckHuaQiGeRejectsKeTouGan(t *testing.T) {
	// 庚克乙（克化神木），破格
	pillars := []model.Pillar{
		{Gan: "丁", Zhi: "卯"},
		{Gan: "壬", Zhi: "寅"},
		{Gan: "丁", Zhi: "卯"},
		{Gan: "庚", Zhi: "辰"},
	}
	scores := map[string]int{"木": 35, "火": 10, "土": 15, "金": 10, "水": 30}

	if got := checkHuaQiGe(pillars, "寅", scores); got != nil {
		t.Fatalf("checkHuaQiGe returned %+v, want nil when 庚克木", got)
	}
}

func TestCheckHuaQiGeRejectsZhiKeOverThirty(t *testing.T) {
	// 水克火（35% > 30%），破格
	pillars := []model.Pillar{
		{Gan: "丁", Zhi: "卯"},
		{Gan: "壬", Zhi: "巳"},
		{Gan: "丁", Zhi: "酉"},
		{Gan: "甲", Zhi: "午"},
	}
	scores := map[string]int{"木": 15, "火": 25, "土": 10, "金": 15, "水": 35}

	if got := checkHuaQiGe(pillars, "巳", scores); got != nil {
		t.Fatalf("checkHuaQiGe returned %+v, want nil when 地支克化神比例 > 30%%", got)
	}
}

func TestCheckHuaQiGeRejectsYearGanHe(t *testing.T) {
	// 丁壬合化木，月支寅木旺地，但只与年干合（非月时干），应拒绝
	pillars := []model.Pillar{
		{Gan: "壬", Zhi: "卯"},
		{Gan: "丙", Zhi: "寅"},
		{Gan: "丁", Zhi: "酉"},
		{Gan: "甲", Zhi: "辰"},
	}
	scores := map[string]int{"木": 35, "火": 10, "土": 15, "金": 5, "水": 35}

	if got := checkHuaQiGe(pillars, "寅", scores); got != nil {
		t.Fatalf("checkHuaQiGe returned %+v, want nil when only year-gan he", got)
	}
}

func TestCheckCongRuoGeRequiresKeXieHao(t *testing.T) {
	// 生扶<10%但月支与天干都没有克泄耗 → nil
	pillars := []model.Pillar{
		{Gan: "庚", Zhi: "申"},
		{Gan: "壬", Zhi: "亥"},
		{Gan: "甲", Zhi: "子"},
		{Gan: "癸", Zhi: "丑"},
	}
	scores := map[string]int{"木": 2, "火": 0, "土": 5, "金": 35, "水": 58}

	if got := checkCongRuoGe(pillars, scores); got != nil {
		t.Fatalf("checkCongRuoGe returned %+v, want nil when no 克泄耗", got)
	}
}

func TestCheckCongRuoGeRejectsWhenDayRootExists(t *testing.T) {
	// 寅中主气含木（甲），日主为甲，有根 → 破格
	pillars := []model.Pillar{
		{Gan: "庚", Zhi: "申"},
		{Gan: "辛", Zhi: "酉"},
		{Gan: "甲", Zhi: "寅"},
		{Gan: "丁", Zhi: "巳"},
	}
	scores := map[string]int{"木": 20, "火": 35, "土": 20, "金": 25, "水": 0}

	if got := checkCongRuoGe(pillars, scores); got != nil {
		t.Fatalf("checkCongRuoGe returned %+v, want nil when 地支主气含日主", got)
	}
}

func TestAnalyzePatternExtendedFallsBackToNormalPattern(t *testing.T) {
	pillars := []model.Pillar{
		{Gan: "甲", Zhi: "子"},
		{Gan: "辛", Zhi: "丑"},
		{Gan: "己", Zhi: "未"},
		{Gan: "庚", Zhi: "午"},
	}
	scores := map[string]int{"木": 20, "火": 18, "土": 16, "金": 14, "水": 12}
	bs := BodyStrengthResult{Verdict: "身旺", Like: []string{"金", "水"}, Dislike: []string{"火", "土"}}

	got := analyzePatternExtended(pillars, "丑", scores, bs)
	if got.PatternName != "正格" {
		t.Fatalf("PatternName = %q, want 正格", got.PatternName)
	}
	if got.PatternType != "正格" {
		t.Fatalf("PatternType = %q, want 正格", got.PatternType)
	}
}

func TestCalculateIncludesPatternAnalysis(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(1990, 1, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}
	if result.PatternAnalysis.PatternName == "" {
		t.Fatal("PatternAnalysis.PatternName is empty")
	}
	if result.PatternAnalysis.PatternType == "" {
		t.Fatal("PatternAnalysis.PatternType is empty")
	}
}
