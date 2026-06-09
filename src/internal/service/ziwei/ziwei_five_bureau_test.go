package ziwei

import (
	"strings"
	"testing"
)
// ═══════════════════════════════════════════════════════════════════════
// Five Bureau (五行局) Data Structure & Integrity Tests
// ═══════════════════════════════════════════════════════════════════════

// TestFiveBureauDataIntegrity validates the Two-way mapping consistency
// and ensures only valid bureau values (2-6) appear in the NaYin table.
func TestFiveBureauDataIntegrity(t *testing.T) {
	// Valid bureau values (水=2, 木=3, 金=4, 土=5, 火=6)
	validValues := map[int]bool{2: true, 3: true, 4: true, 5: true, 6: true}
	validNames := map[string]bool{"水二局": true, "木三局": true, "金四局": true, "土五局": true, "火六局": true}

	// 1. FiveBureauName maps every valid value to a valid name
	for v := 2; v <= 6; v++ {
		name, ok := FiveBureauName[v]
		if !ok {
			t.Errorf("FiveBureauName[%d] is missing", v)
			continue
		}
		if !validNames[name] {
			t.Errorf("FiveBureauName[%d] = %q, not a valid bureau name", v, name)
		}
	}

	// 2. FiveBureauValue reverse-maps every name back to its value
	for v := 2; v <= 6; v++ {
		name := FiveBureauName[v]
		got, ok := FiveBureauValue[name]
		if !ok {
			t.Errorf("FiveBureauValue[%q] is missing (should be %d)", name, v)
			continue
		}
		if got != v {
			t.Errorf("FiveBureauValue[%q] = %d, want %d", name, got, v)
		}
	}

	// 3. NaYinBureauTable only contains valid bureau values
	for i, v := range NaYinBureauTable {
		if !validValues[v] {
			t.Errorf("NaYinBureauTable[%d] = %d, not a valid bureau value (2-6)", i, v)
		}
	}

	// 4. NaYinBureauTable has exactly 30 entries
	if len(NaYinBureauTable) != 30 {
		t.Errorf("NaYinBureauTable length = %d, want 30", len(NaYinBureauTable))
	}
}

// TestNaYinBureauTable_Full60Ganzhi validates that every possible 干支 combination
// produces the correct bureau according to the 六十甲子纳音 table.
// This covers all 30 干支 pairs (each pair maps to one bureau value).
func TestNaYinBureauTable_Full60Ganzhi(t *testing.T) {
	// Reference: 六十甲子纳音五行局
	// Each row = 2 干支 (stem, branch) pairs sharing the same bureau value.
	// The 60 干支 cycle has 30 such pairs.
	expected := []struct {
		stemA, branchA int // first 干支 of the pair
		stemB, branchB int // second 干支 of the pair
		wantJu         int
		wantName       string
	}{
		// Row 1: 甲子乙丑 → 金
		{0, 0, 1, 1, 4, "金四局"},
		// Row 2: 丙寅丁卯 → 火
		{2, 2, 3, 3, 6, "火六局"},
		// Row 3: 戊辰己巳 → 木
		{4, 4, 5, 5, 3, "木三局"},
		// Row 4: 庚午辛未 → 土
		{6, 6, 7, 7, 5, "土五局"},
		// Row 5: 壬申癸酉 → 金
		{8, 8, 9, 9, 4, "金四局"},
		// Row 6: 甲戌乙亥 → 火
		{0, 10, 1, 11, 6, "火六局"},
		// Row 7: 丙子丁丑 → 水
		{2, 0, 3, 1, 2, "水二局"},
		// Row 8: 戊寅己卯 → 土
		{4, 2, 5, 3, 5, "土五局"},
		// Row 9: 庚辰辛巳 → 金
		{6, 4, 7, 5, 4, "金四局"},
		// Row 10: 壬午癸未 → 木
		{8, 6, 9, 7, 3, "木三局"},
		// Row 11: 甲申乙酉 → 水
		{0, 8, 1, 9, 2, "水二局"},
		// Row 12: 丙戌丁亥 → 土
		{2, 10, 3, 11, 5, "土五局"},
		// Row 13: 戊子己丑 → 火
		{4, 0, 5, 1, 6, "火六局"},
		// Row 14: 庚寅辛卯 → 木
		{6, 2, 7, 3, 3, "木三局"},
		// Row 15: 壬辰癸巳 → 水
		{8, 4, 9, 5, 2, "水二局"},
		// Row 16: 甲午乙未 → 金
		{0, 6, 1, 7, 4, "金四局"},
		// Row 17: 丙申丁酉 → 火
		{2, 8, 3, 9, 6, "火六局"},
		// Row 18: 戊戌己亥 → 木
		{4, 10, 5, 11, 3, "木三局"},
		// Row 19: 庚子辛丑 → 土
		{6, 0, 7, 1, 5, "土五局"},
		// Row 20: 壬寅癸卯 → 金
		{8, 2, 9, 3, 4, "金四局"},
		// Row 21: 甲辰乙巳 → 火
		{0, 4, 1, 5, 6, "火六局"},
		// Row 22: 丙午丁未 → 水
		{2, 6, 3, 7, 2, "水二局"},
		// Row 23: 戊申己酉 → 土
		{4, 8, 5, 9, 5, "土五局"},
		// Row 24: 庚戌辛亥 → 金
		{6, 10, 7, 11, 4, "金四局"},
		// Row 25: 壬子癸丑 → 木
		{8, 0, 9, 1, 3, "木三局"},
		// Row 26: 甲寅乙卯 → 水
		{0, 2, 1, 3, 2, "水二局"},
		// Row 27: 丙辰丁巳 → 土
		{2, 4, 3, 5, 5, "土五局"},
		// Row 28: 戊午己未 → 火
		{4, 6, 5, 7, 6, "火六局"},
		// Row 29: 庚申辛酉 → 木
		{6, 8, 7, 9, 3, "木三局"},
		// Row 30: 壬戌癸亥 → 水
		{8, 10, 9, 11, 2, "水二局"},
	}

	for _, row := range expected {
		// Test first 干支 of the pair
		juA := calcFiveBureau(row.stemA, row.branchA)
		if juA != row.wantJu {
			t.Errorf("%s%s: calcFiveBureau = %d, want %d",
				StemNames[row.stemA], BranchNames[row.branchA], juA, row.wantJu)
		}
		if FiveBureauName[juA] != row.wantName {
			t.Errorf("%s%s: name = %q, want %q",
				StemNames[row.stemA], BranchNames[row.branchA], FiveBureauName[juA], row.wantName)
		}

		// Test second 干支 of the pair (same bureau)
		juB := calcFiveBureau(row.stemB, row.branchB)
		if juB != row.wantJu {
			t.Errorf("%s%s: calcFiveBureau = %d, want %d",
				StemNames[row.stemB], BranchNames[row.branchB], juB, row.wantJu)
		}
		if FiveBureauName[juB] != row.wantName {
			t.Errorf("%s%s: name = %q, want %q",
				StemNames[row.stemB], BranchNames[row.branchB], FiveBureauName[juB], row.wantName)
		}
	}
}

// TestCalcFiveBureau_EdgeCases tests boundary conditions.
// Note: calcFiveBureau internally uses modulo arithmetic via ganzhiPairIndex,
// so "out-of-range" inputs still wrap around and produce valid results.
func TestCalcFiveBureau_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		stem, bran int
		wantJu     int
		wantName   string
	}{
		// First 干支 of the 60-cycle
		{"甲子(cycle start)", 0, 0, 4, "金四局"},
		// Last 干支 of the 60-cycle
		{"癸亥(cycle end)", 9, 11, 2, "水二局"},
		// Mid-cycle point
		{"甲午(half cycle)", 0, 6, 4, "金四局"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ju := calcFiveBureau(tt.stem, tt.bran)
			if ju != tt.wantJu {
				t.Errorf("calcFiveBureau(%d,%d) = %d, want %d", tt.stem, tt.bran, ju, tt.wantJu)
			}
			if FiveBureauName[ju] != tt.wantName {
				t.Errorf("FiveBureauName[%d] = %q, want %q", ju, FiveBureauName[ju], tt.wantName)
			}
		})
	}
}

// TestCalcFiveBureau_InvalidInputs verifies that invalid inputs still produce
// valid bureau values (2-6) thanks to modulo wrapping in ganzhiPairIndex.
func TestCalcFiveBureau_InvalidInputs(t *testing.T) {
	inputs := []struct {
		stem, bran int
	}{
		{-1, 0},
		{0, -1},
		{10, 0},
		{0, 12},
		{-100, 999},
	}

	for _, in := range inputs {
		t.Run("", func(t *testing.T) {
			ju := calcFiveBureau(in.stem, in.bran)
			if ju < 2 || ju > 6 {
				t.Errorf("calcFiveBureau(%d,%d) = %d, not in valid range [2,6]", in.stem, in.bran, ju)
			}
			name := FiveBureauName[ju]
			if name == "" {
				t.Errorf("FiveBureauName[%d] is empty for input (%d,%d)", ju, in.stem, in.bran)
			}
		})
	}
}

// TestGanzhiPairIndex_FullCycle validates that every valid (stem, branch)
// combination maps to the correct pair index (0-29).
func TestGanzhiPairIndex_FullCycle(t *testing.T) {
	// Expected pair indices for the 60 干支 cycle
	// computed manually: 甲子=0, 乙丑=0, 丙寅=1, 丁卯=1, ..., 壬戌=29, 癸亥=29
	expectedPairs := []int{
		0, 0, 1, 1, 2, 2, 3, 3, 4, 4, // 甲子乙丑丙寅丁卯戊辰己巳庚午辛未壬申癸酉
		5, 5, 6, 6, 7, 7, 8, 8, 9, 9, // 甲戌乙亥丙子丁丑戊寅己卯庚辰辛巳壬午癸未
		10, 10, 11, 11, 12, 12, 13, 13, 14, 14, // 甲申乙酉丙戌丁亥戊子己丑庚寅辛卯壬辰癸巳
		15, 15, 16, 16, 17, 17, 18, 18, 19, 19, // 甲午乙未丙申丁酉戊戌己亥庚子辛丑壬寅癸卯
		20, 20, 21, 21, 22, 22, 23, 23, 24, 24, // 甲辰乙巳丙午丁未戊申己酉庚戌辛亥壬子癸丑
		25, 25, 26, 26, 27, 27, 28, 28, 29, 29, // 甲寅乙卯丙辰丁巳戊午己未庚申辛酉壬戌癸亥
	}

	// Generate all 60 干支 combinations
	for ganzhiIdx := 0; ganzhiIdx < 60; ganzhiIdx++ {
		stem := ganzhiIdx % 10
		branch := ganzhiIdx % 12
		pairIdx := ganzhiPairIndex(stem, branch)
		want := expectedPairs[ganzhiIdx]

		if pairIdx != want {
			t.Errorf("ganzhiPairIndex(stem=%d(%s), branch=%d(%s)) = %d, want %d (ganzhi=%d)",
				stem, StemNames[stem], branch, BranchNames[branch], pairIdx, want, ganzhiIdx)
		}
	}
}

// TestGanzhiPairIndex_EachPairConsistent verifies both 干支 in a pair
// produce the same pair index, and all 30 pair indices (0-29) are covered.
func TestGanzhiPairIndex_EachPairConsistent(t *testing.T) {
	pairCounts := make(map[int]int)
	seenPairs := make(map[int]bool)

	for ganzhiIdx := 0; ganzhiIdx < 60; ganzhiIdx++ {
		stem := ganzhiIdx % 10
		branch := ganzhiIdx % 12
		pairIdx := ganzhiPairIndex(stem, branch)

		// Each 干支 in same pair (even+odd ganzhi index) should give same pairIdx
		pairCounts[pairIdx] = pairCounts[pairIdx] + 1
		seenPairs[pairIdx] = true
	}

	// All 30 pairs should be covered
	for i := 0; i < 30; i++ {
		if !seenPairs[i] {
			t.Errorf("pair index %d is not produced by any 干支 combination", i)
		}
	}

	// Each pair should have exactly 2 干支
	for i := 0; i < 30; i++ {
		if pairCounts[i] != 2 {
			t.Errorf("pair index %d has %d 干支, want 2", i, pairCounts[i])
		}
	}
}

// TestFiveBureau_Integration_AdditionalCharts tests FiveBureau in full chart
// calculations with various birth dates. This complements the existing
// TestChart_GuiWeiYear_* which tests 木三局.
func TestFiveBureau_Integration_AdditionalCharts(t *testing.T) {
	svc := NewZiWeiService()

	type testCase struct {
		name   string
		year   int
		month  int
		day    int
		hour   int
		min    int
		gender string
	}

	tests := []testCase{
		{
			name:   "1990-01-01 子时",
			year:   1990,
			month:  1,
			day:    1,
			hour:   23,
			min:    0,
			gender: "男",
		},
		{
			name:   "2000-06-15 午时",
			year:   2000,
			month:  6,
			day:    15,
			hour:   12,
			min:    0,
			gender: "女",
		},
		{
			name:   "1988-08-08 08:00",
			year:   1988,
			month:  8,
			day:    8,
			hour:   8,
			min:    0,
			gender: "男",
		},
		{
			name:   "2024-01-15 15:30",
			year:   2024,
			month:  1,
			day:    15,
			hour:   15,
			min:    30,
			gender: "女",
		},
		{
			name:   "1950-10-01 戌时",
			year:   1950,
			month:  10,
			day:    1,
			hour:   19,
			min:    0,
			gender: "男",
		},
		{
			name:   "1976-07-28 03:42(唐山大地震时间)",
			year:   1976,
			month:  7,
			day:    28,
			hour:   3,
			min:    42,
			gender: "男",
		},
		{
			name:   "2008-05-12 14:28(汶川地震时间)",
			year:   2008,
			month:  5,
			day:    12,
			hour:   14,
			min:    28,
			gender: "女",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(tt.year, tt.month, tt.day, tt.hour, tt.min, tt.gender)
			if err != nil {
				t.Fatalf("CalculateChart failed: %v", err)
			}

			// Check FiveBureau is non-empty
			if chart.FiveBureau == "" {
				t.Error("FiveBureau is empty")
			}

			// Check it ends with "局" (using rune-safe comparison)
			if !strings.HasSuffix(chart.FiveBureau, "局") {
				t.Errorf("FiveBureau = %q, should end with 局", chart.FiveBureau)
			}

			// Check JuValue is in valid range [2,6]
			if chart.JuValue < 2 || chart.JuValue > 6 {
				t.Errorf("JuValue = %d, want 2-6", chart.JuValue)
			}

			// Verify FiveBureau string matches the internal JuValue
			expectedName := FiveBureauName[chart.JuValue]
			if chart.FiveBureau != expectedName {
				t.Errorf("FiveBureau = %q, but JuValue=%d maps to %q",
					chart.FiveBureau, chart.JuValue, expectedName)
			}
		})
	}
}

// TestFiveBureau_Roundtrip verifies that the FiveBureau field on a chart
// roundtrips correctly through the name/value maps.
func TestFiveBureau_Roundtrip(t *testing.T) {
	// For each valid bureau value, verify the roundtrip:
	// value → FiveBureauName → FiveBureauValue → value
	for v := 2; v <= 6; v++ {
		name := FiveBureauName[v]
		back := FiveBureauValue[name]
		if back != v {
			t.Errorf("roundtrip failed: %d → %q → %d", v, name, back)
		}
	}

	// For each valid name, verify reverse roundtrip
	for _, name := range []string{"水二局", "木三局", "金四局", "土五局", "火六局"} {
		v := FiveBureauValue[name]
		back := FiveBureauName[v]
		if back != name {
			t.Errorf("reverse roundtrip failed: %q → %d → %q", name, v, back)
		}
	}
}
