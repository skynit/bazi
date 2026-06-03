package bazi

import (
	"fmt"

	"bazi/internal/service/data"
)

// TiaohouResult holds the tiaohou analysis result for a BaZi chart.
type TiaohouResult struct {
	Stem      string             `json:"stem"`
	Month     string             `json:"month"`
	Rules     []data.TiaohouRule `json:"rules"`
	Primary   string             `json:"primary_god"`
	DepthHint string             `json:"depth_hint,omitempty"`
	Reasons   []string           `json:"reasons"`
	Summary   string             `json:"summary"`
}

// AnalyzeTiaohou performs tiaohou (调候) analysis for a given day stem and month branch.
// Returns analysis result based on 《穷通宝鉴》rules.
func AnalyzeTiaohou(dayStem, monthBranch string) (*TiaohouResult, error) {
	return AnalyzeTiaohouWithDepth(dayStem, monthBranch, "")
}

// AnalyzeTiaohouWithDepth performs tiaohou (调候) analysis considering month depth (月令深浅).
// dayBranch helps determine if the day is in the early/mid/late part of the lunar month.
// Pass empty string for dayBranch to skip depth consideration.
func AnalyzeTiaohouWithDepth(dayStem, monthBranch, dayBranch string) (*TiaohouResult, error) {
	rules := data.GetTiaohou(dayStem, monthBranch)
	if len(rules) == 0 {
		return nil, fmt.Errorf("no tiaohou rules for stem=%s month=%s", dayStem, monthBranch)
	}

	primary, depthHint := pickPrimaryWithDepth(dayStem, monthBranch, dayBranch, rules)
	reasons := make([]string, len(rules))
	for i, r := range rules {
		reasons[i] = r.Reason
	}

	summary := buildTiaohouSummary(dayStem, monthBranch, primary, depthHint, rules)

	return &TiaohouResult{
		Stem:      dayStem,
		Month:     monthBranch,
		Rules:     rules,
		Primary:   primary,
		DepthHint: depthHint,
		Reasons:   reasons,
		Summary:   summary,
	}, nil
}

// pickPrimaryWithDepth selects the primary 调候用神 considering the depth within the month.
// 穷通宝鉴原文中常区分"上旬/中旬/下旬"，例如甲木生寅月：
//   - 上旬（立春后）余寒重，首用丙火温之
//   - 中旬（雨水后）寒退，丙癸并用
//   - 下旬（惊蛰后）木旺，可用庚金劈甲
// 由于本函数无法获得具体节气和日数，采用日支相对月支的位置作为深度启发。
func pickPrimaryWithDepth(dayStem, monthBranch, dayBranch string, rules []data.TiaohouRule) (string, string) {
	if dayBranch == "" || len(rules) <= 1 {
		return rules[0].XiShen, ""
	}

	monthIdx, ok1 := zhiToDepthIdx[monthBranch]
	dayIdx, ok2 := zhiToDepthIdx[dayBranch]
	if !ok1 || !ok2 {
		return rules[0].XiShen, ""
	}

	// Calculate distance in the 12-branch cycle, handling wrap-around
	// Forward distance (day after month in cycle)
	fwd := (dayIdx - monthIdx + 12) % 12
	// Backward distance (day before month in cycle)
	bwd := (monthIdx - dayIdx + 12) % 12

	depthHint := ""
	selected := rules[0].XiShen

	switch {
	case fwd >= 1 && fwd <= 2:
		// Day branch is 1-2 positions after month branch in 12-cycle
		// This roughly indicates the late part of the month (月内偏后)
		depthHint = "月末"
		if len(rules) > 1 {
			// For late month, prefer the secondary 调候 which often handles
			// the stronger month qi (e.g., 庚金劈甲 for 甲+寅 late month)
			selected = rules[1].XiShen
		}
	case bwd >= 1 && bwd <= 2:
		// Day branch is 1-2 positions before month branch
		// This roughly indicates the early part of the month (月内偏前)
		depthHint = "月初"
		// For early month, keep primary (余寒重, need main warming element)
	case fwd >= 3 && fwd <= 5:
		// Day branch is mid-month forward
		depthHint = "月中"
	case bwd >= 3 && bwd <= 5:
		// Day branch is mid-month backward (previous month)
		depthHint = "月末上月"
	default:
		// Far from month branch, ambiguous
		depthHint = ""
	}

	_ = dayStem // dayStem reserved for future stem-aware depth rules
	return selected, depthHint
}

// zhiToDepthIdx maps 12 branches to a 0-11 index for depth comparison.
// Used to estimate month depth (月令深浅) from the day branch's position.
var zhiToDepthIdx = map[string]int{
	"寅": 0, "卯": 1, "辰": 2, "巳": 3, "午": 4, "未": 5,
	"申": 6, "酉": 7, "戌": 8, "亥": 9, "子": 10, "丑": 11,
}

// buildTiaohouSummary constructs a human-readable summary.
func buildTiaohouSummary(stem, month, primary, depthHint string, rules []data.TiaohouRule) string {
	monthNames := map[string]string{
		"寅": "正月", "卯": "二月", "辰": "三月",
		"巳": "四月", "午": "五月", "未": "六月",
		"申": "七月", "酉": "八月", "戌": "九月",
		"亥": "十月", "子": "十一月", "丑": "十二月",
	}

	stemNames := map[string]string{
		"甲": "甲木", "乙": "乙木", "丙": "丙火", "丁": "丁火",
		"戊": "戊土", "己": "己土", "庚": "庚金", "辛": "辛金",
		"壬": "壬水", "癸": "癸水",
	}

	elemNames := map[string]string{
		"木": "木", "火": "火", "土": "土", "金": "金", "水": "水",
	}

	primaryElem := data.GanElement[primary]
	if primaryElem == "" {
		primaryElem = primary // fallback for non-gan chars
	}

	depthSuffix := ""
	if depthHint != "" {
		depthSuffix = fmt.Sprintf("（%s偏% s）", depthHint, depthLabel(depthHint))
	}

	return fmt.Sprintf("%s生%s月，调候用神为%s（五行属%s）%s。%s",
		stemNames[stem], monthNames[month], primary, elemNames[primaryElem], depthSuffix, rules[0].Reason)
}

func depthLabel(hint string) string {
	switch hint {
	case "月初":
		return "前"
	case "月末":
		return "后"
	case "月中":
		return "中"
	case "月末上月":
		return "末"
	default:
		return ""
	}
}