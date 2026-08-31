package ziwei

import (
	"testing"
	"time"
)

func TestResolvePeriodSolarDate(t *testing.T) {
	now := time.Date(2026, time.August, 31, 16, 0, 0, 0, time.FixedZone("test", 8*60*60))

	year, month, day, err := ResolvePeriodSolarDate(0, 0, 0, now)
	if err != nil || year != 2026 || month != 8 || day != 31 {
		t.Fatalf("default date = %04d-%02d-%02d, err=%v", year, month, day, err)
	}
	if _, _, _, err := ResolvePeriodSolarDate(2026, 8, 0, now); err == nil {
		t.Fatal("partial date was accepted")
	}
	if _, _, _, err := ResolvePeriodSolarDate(2026, 2, 29, now); err == nil {
		t.Fatal("invalid civil date was accepted")
	}
}

func TestNominalAgeAtUsesLunarYear(t *testing.T) {
	chart := &ZiWeiChart{CalculationInput: ZiWeiCalculationInput{Year: 2000, Month: 8, Day: 16}}
	target := time.Date(2022, time.August, 19, 12, 0, 0, 0, time.Local)
	if got := NominalAgeAt(chart, target); got != 23 {
		t.Fatalf("nominal age = %d, want 23", got)
	}
}

func TestDayunDescriptionKeepsSafetyBoundary(t *testing.T) {
	if got := DayunDescription("財帛宮", 23); got != "现金流与资源配置主题的结构位置；不构成财务建议" {
		t.Fatalf("wealth-palace description = %q", got)
	}
}
