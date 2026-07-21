package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"
)

// expectedTenGods maps dayGan -> stemName -> expected ten-god (isDayPillarStem=false).
// This is the authoritative reference table verified against traditional BaZi theory.
var expectedTenGods = map[string]map[string]string{
	// 甲日主
	"甲": {
		"甲": "比肩", "乙": "劫财", "丙": "食神", "丁": "伤官",
		"戊": "偏财", "己": "正财", "庚": "七杀", "辛": "正官",
		"壬": "偏印", "癸": "正印",
	},
	// 乙日主
	"乙": {
		"甲": "劫财", "乙": "比肩", "丙": "伤官", "丁": "食神",
		"戊": "正财", "己": "偏财", "庚": "正官", "辛": "七杀",
		"壬": "正印", "癸": "偏印",
	},
	// 丙日主
	"丙": {
		"甲": "偏印", "乙": "正印", "丙": "比肩", "丁": "劫财",
		"戊": "食神", "己": "伤官", "庚": "偏财", "辛": "正财",
		"壬": "七杀", "癸": "正官",
	},
	// 丁日主
	"丁": {
		"甲": "正印", "乙": "偏印", "丙": "劫财", "丁": "比肩",
		"戊": "伤官", "己": "食神", "庚": "正财", "辛": "偏财",
		"壬": "正官", "癸": "七杀",
	},
	// 戊日主
	"戊": {
		"甲": "七杀", "乙": "正官", "丙": "偏印", "丁": "正印",
		"戊": "比肩", "己": "劫财", "庚": "食神", "辛": "伤官",
		"壬": "偏财", "癸": "正财",
	},
	// 己日主
	"己": {
		"甲": "正官", "乙": "七杀", "丙": "正印", "丁": "偏印",
		"戊": "劫财", "己": "比肩", "庚": "伤官", "辛": "食神",
		"壬": "正财", "癸": "偏财",
	},
	// 庚日主
	"庚": {
		"甲": "偏财", "乙": "正财", "丙": "七杀", "丁": "正官",
		"戊": "偏印", "己": "正印", "庚": "比肩", "辛": "劫财",
		"壬": "食神", "癸": "伤官",
	},
	// 辛日主
	"辛": {
		"甲": "正财", "乙": "偏财", "丙": "正官", "丁": "七杀",
		"戊": "正印", "己": "偏印", "庚": "劫财", "辛": "比肩",
		"壬": "伤官", "癸": "食神",
	},
	// 壬日主
	"壬": {
		"甲": "食神", "乙": "伤官", "丙": "偏财", "丁": "正财",
		"戊": "七杀", "己": "正官", "庚": "偏印", "辛": "正印",
		"壬": "比肩", "癸": "劫财",
	},
	// 癸日主
	"癸": {
		"甲": "伤官", "乙": "食神", "丙": "正财", "丁": "偏财",
		"戊": "正官", "己": "七杀", "庚": "正印", "辛": "偏印",
		"壬": "劫财", "癸": "比肩",
	},
}

// allStems is the complete set of 10 heavenly stems in standard order.
var allStems = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}

// TestClassifyTenGod_AllCombinations_WithFalse asserts that ClassifyTenGod returns
// the correct ten-god for every one of the 100 (stemName × dayGan) combinations
// when isDayPillarStem=false (the normal case for non-day-pillar stems and hidden stems).
func TestClassifyTenGod_AllCombinations_WithFalse(t *testing.T) {
	for _, dayGan := range allStems {
		dayTable, ok := expectedTenGods[dayGan]
		if !ok {
			t.Fatalf("missing expected table for dayGan=%s", dayGan)
		}

		t.Run(dayGan, func(t *testing.T) {
			for _, stemName := range allStems {
				expected, ok := dayTable[stemName]
				if !ok {
					t.Fatalf("missing expected value for stemName=%s under dayGan=%s", stemName, dayGan)
				}

				got := ClassifyTenGod(stemName, dayGan, false)
				if got != expected {
					t.Errorf(
						"ClassifyTenGod(stemName=%q, dayGan=%q, false) = %q; want %q",
						stemName, dayGan, got, expected,
					)
				}
			}
		})
	}
}

// TestClassifyTenGod_AllCombinations_WithTrue asserts that ClassifyTenGod returns
// "日主" when isDayPillarStem=true AND stemName == dayGan, and returns normal
// classification for all other combinations (matching the isDayPillarStem=false table).
func TestClassifyTenGod_AllCombinations_WithTrue(t *testing.T) {
	for _, dayGan := range allStems {
		dayTable, ok := expectedTenGods[dayGan]
		if !ok {
			t.Fatalf("missing expected table for dayGan=%s", dayGan)
		}

		t.Run(dayGan, func(t *testing.T) {
			for _, stemName := range allStems {
				if stemName == dayGan {
					// When it IS the day pillar stem and matches dayGan → 日主
					got := ClassifyTenGod(stemName, dayGan, true)
					if got != "日主" {
						t.Errorf(
							"ClassifyTenGod(stemName=%q, dayGan=%q, true) = %q; want %q",
							stemName, dayGan, got, "日主",
						)
					}
				} else {
					// When isDayPillarStem=true but stemName != dayGan → normal classification
					expected, ok := dayTable[stemName]
					if !ok {
						t.Fatalf("missing expected value for stemName=%s under dayGan=%s", stemName, dayGan)
					}
					got := ClassifyTenGod(stemName, dayGan, true)
					if got != expected {
						t.Errorf(
							"ClassifyTenGod(stemName=%q, dayGan=%q, true) = %q; want %q",
							stemName, dayGan, got, expected,
						)
					}
				}
			}
		})
	}
}

// TestClassifyTenGod_EdgeCases tests edge cases: empty string input and invalid stem names.
func TestClassifyTenGod_EdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		stemName       string
		dayGan         string
		isDayPillar    bool
		expected       string
		expectNonEmpty bool
	}{
		// Empty stem name with valid dayGan
		{name: "empty stem", stemName: "", dayGan: "甲", isDayPillar: false, expected: ""},
		// Empty dayGan with valid stem
		{name: "empty dayGan", stemName: "甲", dayGan: "", isDayPillar: false, expected: ""},
		// Both empty — invalid inputs fail closed instead of fabricating a peer relation
		{name: "both empty", stemName: "", dayGan: "", isDayPillar: false, expected: ""},
		// Invalid stem name (only stemName invalid, dayGan valid)
		{name: "invalid stem", stemName: "X", dayGan: "甲", isDayPillar: false, expected: ""},
		// Invalid dayGan (stemName valid, dayGan invalid)
		{name: "invalid dayGan", stemName: "甲", dayGan: "Y", isDayPillar: false, expected: ""},
		// Both invalid — unknown stems cannot form a ten-god relation
		{name: "both invalid", stemName: "Z", dayGan: "W", isDayPillar: false, expected: ""},
		// isDayPillarStem=true with mismatch (same as non-day-pillar)
		{name: "true with diff stem", stemName: "乙", dayGan: "甲", isDayPillar: true, expected: "劫财"},
		// isDayPillarStem=true with match → 日主
		{name: "true with same stem", stemName: "甲", dayGan: "甲", isDayPillar: true, expected: "日主"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyTenGod(tt.stemName, tt.dayGan, tt.isDayPillar)
			if got != tt.expected {
				t.Errorf(
					"ClassifyTenGod(stemName=%q, dayGan=%q, %v) = %q; want %q",
					tt.stemName, tt.dayGan, tt.isDayPillar, got, tt.expected,
				)
			}
		})
	}
}
