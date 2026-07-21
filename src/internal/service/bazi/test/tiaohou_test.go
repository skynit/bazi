package bazi_test

import (
	. "bazi/internal/service/bazi"
	"testing"
)

func TestAnalyzeTiaohou(t *testing.T) {
	tests := []struct {
		stem        string
		month       string
		wantPrimary string
		wantLen     int
	}{
		{"甲", "寅", "丙", 2},
		{"乙", "卯", "丙", 2},
		{"丙", "午", "壬", 3},
		{"丁", "子", "甲", 4},
		{"庚", "申", "丁", 3},
		{"壬", "亥", "戊", 3},
	}

	for _, tt := range tests {
		t.Run(tt.stem+tt.month, func(t *testing.T) {
			result, err := AnalyzeTiaohou(tt.stem, tt.month)
			if err != nil {
				t.Fatalf("AnalyzeTiaohou(%s, %s) error = %v", tt.stem, tt.month, err)
			}
			if result.TablePrimaryCandidate != tt.wantPrimary {
				t.Errorf("TablePrimaryCandidate = %v, want %v", result.TablePrimaryCandidate, tt.wantPrimary)
			}
			if len(result.Rules) != tt.wantLen {
				t.Errorf("len(Rules) = %v, want %v", len(result.Rules), tt.wantLen)
			}
		})
	}
}
