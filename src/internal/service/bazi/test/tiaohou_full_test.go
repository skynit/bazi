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
//  2. the table-first candidate matches the first source-table rule
//  3. every result satisfies the auditable evidence contract
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

				if result.TablePrimaryCandidate == "" {
					t.Error("table primary candidate is empty")
				}

				rules := data.GetTiaohou(stem, month)
				if len(rules) == 0 {
					t.Error("GetTiaohou returned empty rules (data integrity issue)")
					return
				}

				if result.TablePrimaryCandidate != rules[0].XiShen {
					t.Errorf("table candidate = %q, want first XiShen %q", result.TablePrimaryCandidate, rules[0].XiShen)
				}
				if !ValidTiaohouEvidence(result, stem, month) {
					t.Errorf("invalid Tiaohou evidence: %+v", result)
				}
			})
		}
	}
}
