package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"
)

func TestAnalyzeTiaohou(t *testing.T) {
	tests := []struct {
		stem       string
		month      string
		wantPrimary string
		wantLen    int
	}{
		{"甲", "寅", "丙", 2},
		{"乙", "卯", "丙", 1},
		{"丙", "午", "壬", 2},
		{"丁", "子", "甲", 2},
		{"庚", "申", "丁", 2},
		{"壬", "亥", "戊", 2},
	}

	for _, tt := range tests {
		t.Run(tt.stem+tt.month, func(t *testing.T) {
			result, err := AnalyzeTiaohou(tt.stem, tt.month)
			if err != nil {
				t.Fatalf("AnalyzeTiaohou(%s, %s) error = %v", tt.stem, tt.month, err)
			}
			if result.Primary != tt.wantPrimary {
				t.Errorf("Primary = %v, want %v", result.Primary, tt.wantPrimary)
			}
			if len(result.Rules) != tt.wantLen {
				t.Errorf("len(Rules) = %v, want %v", len(result.Rules), tt.wantLen)
			}
		})
	}
}