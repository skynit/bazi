package bazi

import (
	"fmt"

	"bazi/internal/service/data"
)

// TiaohouResult holds the tiaohou analysis result for a BaZi chart.
type TiaohouResult struct {
	Stem    string             `json:"stem"`
	Month   string             `json:"month"`
	Rules   []data.TiaohouRule `json:"rules"`
	Primary string             `json:"primary_god"`
	Reasons []string           `json:"reasons"`
	Summary string             `json:"summary"`
}

// AnalyzeTiaohou performs tiaohou (调候) analysis for a given day stem and month branch.
// Returns analysis result based on 《穷通宝鉴》rules.
func AnalyzeTiaohou(dayStem, monthBranch string) (*TiaohouResult, error) {
	rules := data.GetTiaohou(dayStem, monthBranch)
	if len(rules) == 0 {
		return nil, fmt.Errorf("no tiaohou rules for stem=%s month=%s", dayStem, monthBranch)
	}

	primary := rules[0].XiShen
	reasons := make([]string, len(rules))
	for i, r := range rules {
		reasons[i] = r.Reason
	}

	summary := buildTiaohouSummary(dayStem, monthBranch, primary, rules)

	return &TiaohouResult{
		Stem:    dayStem,
		Month:   monthBranch,
		Rules:   rules,
		Primary: primary,
		Reasons: reasons,
		Summary: summary,
	}, nil
}

// buildTiaohouSummary constructs a human-readable summary.
func buildTiaohouSummary(stem, month, primary string, rules []data.TiaohouRule) string {
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

	return fmt.Sprintf("%s生%s月，调候用神为%s（五行属%s）。%s",
		stemNames[stem], monthNames[month], primary, elemNames[primaryElem], rules[0].Reason)
}