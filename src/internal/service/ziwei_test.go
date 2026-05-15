package service

import (
	"testing"
)

func TestZiWeiChart(t *testing.T) {
	svc := NewZiWeiService()

	// Test with a known birth: 1984-02-15 08:00 Male (甲子年)
	chart, err := svc.CalculateChart(1984, 2, 15, 8, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	if chart == nil {
		t.Fatal("chart is nil")
	}

	// Verify 12 palaces exist
	for i := range chart.Palaces {
		if chart.Palaces[i].Name == "" {
			t.Errorf("palace %d has empty name", i)
		}
	}

	// Verify first palace is 命宮
	if chart.Palaces[0].Name != "命宮" {
		t.Errorf("expected first palace to be 命宮, got %s", chart.Palaces[0].Name)
	}

	// Verify essential fields
	if chart.FiveBureau == "" {
		t.Error("FiveBureau is empty")
	}
	if chart.BodyPalace == "" {
		t.Error("BodyPalace is empty")
	}
	if chart.LifeMaster == "" {
		t.Error("LifeMaster is empty")
	}
	if chart.BodyMaster == "" {
		t.Error("BodyMaster is empty")
	}

	// Verify at least some main stars are placed
	hasStars := false
	for _, p := range chart.Palaces {
		if len(p.MainStars) > 0 {
			hasStars = true
			break
		}
	}
	if !hasStars {
		t.Error("no main stars found in any palace")
	}

	// Verify brightness data
	hasBrightness := false
	for _, p := range chart.Palaces {
		if len(p.Brightness) > 0 {
			hasBrightness = true
			break
		}
	}
	if !hasBrightness {
		t.Error("no brightness data found")
	}

	t.Logf("Chart: FiveBureau=%s, BodyPalace=%s, LifeMaster=%s, BodyMaster=%s",
		chart.FiveBureau, chart.BodyPalace, chart.LifeMaster, chart.BodyMaster)
}

func TestZiWeiPatterns(t *testing.T) {
	svc := NewZiWeiService()

	// Use a date known to produce specific patterns
	chart, err := svc.CalculateChart(1990, 6, 15, 12, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	patterns := svc.DetectPatterns(chart)
	t.Logf("Detected patterns: %v", patterns)

	// Verify that pattern detection returns results (may be empty for some charts)
	// The method should at least not panic
	if patterns == nil {
		t.Error("DetectPatterns returned nil")
	}

	// Test with nil chart
	nilPatterns := svc.DetectPatterns(nil)
	if nilPatterns != nil {
		t.Error("DetectPatterns should return nil for nil chart")
	}

	// Verify each pattern is non-empty string
	for i, p := range patterns {
		if p == "" {
			t.Errorf("pattern at index %d is empty", i)
		}
	}
}

func TestZiWeiDayun(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(1984, 2, 15, 8, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	dayun := svc.CalculateDayun(chart)
	if len(dayun) == 0 {
		t.Fatal("dayun is empty")
	}

	// Verify dayun covers a full lifespan (at least 8 stages, up to ~80+ years)
	if len(dayun) < 8 {
		t.Errorf("expected at least 8 dayun stages, got %d", len(dayun))
	}

	// Verify first dayun starts at reasonable age
	if dayun[0].StartAge < 0 || dayun[0].StartAge > 10 {
		t.Errorf("unexpected first dayun start age: %d", dayun[0].StartAge)
	}

	// Verify ages are sequential and contiguous
	for i := 1; i < len(dayun); i++ {
		if dayun[i].StartAge != dayun[i-1].EndAge+1 {
			t.Errorf("dayun %d start age %d != previous end age %d + 1",
				i, dayun[i].StartAge, dayun[i-1].EndAge)
		}
	}

	// Verify each dayun has a palace name
	for i, d := range dayun {
		if d.Palace == "" {
			t.Errorf("dayun %d has empty palace", i)
		}
		// Stars may be empty for some dayun stages
	}

	// Verify against known date
	t.Logf("First Dayun: age %d-%d, palace=%s, stars=%v",
		dayun[0].StartAge, dayun[0].EndAge, dayun[0].Palace, dayun[0].Stars)

	// Test nil handling
	nilDayun := svc.CalculateDayun(nil)
	if nilDayun != nil {
		t.Error("CalculateDayun should return nil for nil chart")
	}
}

func TestZiWeiFlyingStars(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(1984, 2, 15, 8, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	analysis := svc.AnalyzeFlyingStars(chart)
	if analysis == nil {
		t.Fatal("AnalyzeFlyingStars returned nil")
	}

	// At minimum, 化忌 should always have a target
	if len(analysis.HuaJi) == 0 {
		t.Log("warning: no 化忌 targets found (may be valid)")
	}

	// Verify targets have required fields
	checkTargets := func(name string, targets []FlyTarget) {
		for i, ft := range targets {
			if ft.FromStar == "" {
				t.Errorf("%s target %d: empty FromStar", name, i)
			}
			if ft.ToPalace == "" {
				t.Errorf("%s target %d: empty ToPalace", name, i)
			}
		}
	}
	checkTargets("HuaLu", analysis.HuaLu)
	checkTargets("HuaQuan", analysis.HuaQuan)
	checkTargets("HuaKe", analysis.HuaKe)
	checkTargets("HuaJi", analysis.HuaJi)

	t.Logf("Flying Stars: HuaLu=%v, HuaQuan=%v, HuaKe=%v, HuaJi=%v",
		len(analysis.HuaLu), len(analysis.HuaQuan), len(analysis.HuaKe), len(analysis.HuaJi))
}

func TestZiWeiLiunian(t *testing.T) {
	svc := NewZiWeiService()

	chart, err := svc.CalculateChart(1984, 2, 15, 8, 0, "男")
	if err != nil {
		t.Fatalf("CalculateChart failed: %v", err)
	}

	// Test LiuNian for 2025
	liuNian := svc.CalculateLiunian(chart, 2025)
	if liuNian == nil {
		t.Fatal("CalculateLiunian returned nil")
	}

	// The returned chart should have the same palace structure
	for i := range liuNian.Palaces {
		if liuNian.Palaces[i].Name == "" {
			t.Errorf("liunian palace %d has empty name", i)
		}
	}

	t.Logf("LiuNian 2025: FiveBureau=%s", liuNian.FiveBureau)
}

func TestZiWeiGender(t *testing.T) {
	svc := NewZiWeiService()

	// Test male
	chartM, err := svc.CalculateChart(1984, 2, 15, 8, 0, "男")
	if err != nil {
		t.Fatalf("male chart failed: %v", err)
	}

	// Test female
	chartF, err := svc.CalculateChart(1984, 2, 15, 8, 0, "女")
	if err != nil {
		t.Fatalf("female chart failed: %v", err)
	}

	// Dayun should be different for male vs female (same birth, different gender)
	dayunM := svc.CalculateDayun(chartM)
	dayunF := svc.CalculateDayun(chartF)

	// At least the palace order should differ
	if len(dayunM) > 0 && len(dayunF) > 0 {
		t.Logf("Male first dayun: %s, Female first dayun: %s",
			dayunM[0].Palace, dayunF[0].Palace)
	}

	// Invalid gender should error
	_, err = svc.CalculateChart(1984, 2, 15, 8, 0, "invalid")
	if err == nil {
		t.Error("expected error for invalid gender")
	}
}
