package ziwei

import "testing"

func TestSanfangProjectionRejectsInvalidChartOrIndex(t *testing.T) {
	chart := calculateProjectionFixture(t)
	if getPalaceSanfang(0) == nil || getChartPalaceSanfang(chart, 0) == nil || getEnhancedSanfang(chart, 0) == nil {
		t.Fatal("valid sanfang input was rejected")
	}
	for _, index := range []int{-1, len(chart.Palaces)} {
		if getPalaceSanfang(index) != nil || getChartPalaceSanfang(chart, index) != nil || getEnhancedSanfang(chart, index) != nil {
			t.Fatalf("invalid palace index %d produced sanfang output", index)
		}
	}
	if getChartPalaceSanfang(nil, 0) != nil || getEnhancedSanfang(nil, 0) != nil {
		t.Fatal("nil chart produced chart-specific sanfang output")
	}
}
