package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"

	"bazi/internal/service/data"
)

// tiaohouMonthOrder defines the 12 month branches in the order used by
// the tiaohou data table (寅=0, 卯=1, 辰=2, 巳=3, 午=4, 未=5, 申=6, 酉=7, 戌=8, 亥=9, 子=10, 丑=11).
var tiaohouMonthOrder = [12]string{
	"寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑",
}

// TestTiaohouFull_All120 verifies all 120 调候 combinations (10 day stems × 12 month branches).
// For each pair, it checks:
//  1. AnalyzeTiaohou returns a non-nil result (no error)
//  2. Primary is non-empty
//  3. For pairs with exactly 1 rule: Primary matches the rule's XiShen
//  4. For pairs with multiple rules: Primary is one of the XiShen values
func TestTiaohouFull_All120(t *testing.T) {
	stems := data.Gans // 甲,乙,丙,丁,戊,己,庚,辛,壬,癸

	for _, stem := range stems {
		for _, month := range tiaohouMonthOrder {
			// Use descriptive subtest name
			subtestName := stem + "+" + month
			t.Run(subtestName, func(t *testing.T) {
				// Step 1: Call AnalyzeTiaohou
				result, err := AnalyzeTiaohou(stem, month)
				if err != nil {
					t.Fatalf("AnalyzeTiaohou(%s, %s) returned error: %v", stem, month, err)
				}
				if result == nil {
					t.Fatal("AnalyzeTiaohou returned nil result without error")
				}

				// Step 2: Primary must be non-empty
				if result.Primary == "" {
					t.Error("Primary is empty (expected at least one 调候用神)")
				}

				// Step 3-4: Validate Primary against the raw data rules
				rules := data.GetTiaohou(stem, month)
				if len(rules) == 0 {
					t.Error("GetTiaohou returned empty rules (data integrity issue)")
					return
				}

				// Collect all XiShen from rules
				xishenSet := make(map[string]bool, len(rules))
				for _, r := range rules {
					if r.XiShen != "" {
						xishenSet[r.XiShen] = true
					}
				}

				if len(rules) == 1 {
					// Single-rule entry: Primary must exactly match XiShen
					expected := rules[0].XiShen
					if result.Primary != expected {
						t.Errorf("单条规则: Primary = %q, 期望 XiShen = %q", result.Primary, expected)
					}
				} else {
					// Multi-rule entry: Primary must be one of the XiShen values
					if !xishenSet[result.Primary] {
						t.Errorf("多条规则: Primary = %q 不在 XiShen 集合 %v 中", result.Primary, keysOfSet(xishenSet))
					}
				}
			})
		}
	}
}

// keysOfSet extracts the keys of a string set into a slice.
func keysOfSet(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
