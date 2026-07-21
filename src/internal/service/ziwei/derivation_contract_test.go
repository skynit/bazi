package ziwei

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestDerivedChartContractBindsBaseAndQuery(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}

	liunian := svc.CalculateLiunian(base, 2025)
	if liunian == nil {
		t.Fatal("CalculateLiunian returned nil")
	}
	if liunian.ContentHash != "" {
		t.Fatalf("derived chart inherited natal content_hash %q", liunian.ContentHash)
	}
	if liunian.DerivationType != "liunian" || liunian.DerivationInput == nil {
		t.Fatalf("missing liunian derivation metadata: %+v", liunian)
	}
	wantInput, err := buildZiWeiDerivationInput("liunian", 2025, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if *liunian.DerivationInput != wantInput {
		t.Fatalf("derivation input = %+v, want %+v", *liunian.DerivationInput, wantInput)
	}
	if liunian.BaseContentHash != base.ContentHash {
		t.Fatalf("base hash = %q, want natal hash %q", liunian.BaseContentHash, base.ContentHash)
	}
	if !ValidDerivedChartContract(liunian) || !DerivedChartMatchesBase(liunian, base) {
		t.Fatal("fresh liunian chart must satisfy the complete derivation contract")
	}
	if svc.ChartMatchesProfile(liunian, "") {
		t.Fatal("derived chart must not be accepted as a cacheable natal chart")
	}
}

func TestDerivedChartContractIsDeterministicAndQuerySensitive(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(1984, 2, 15, 8, 0, "女")
	if err != nil {
		t.Fatal(err)
	}

	first := svc.CalculateLiuriForDate(base, 2026, 7, 15)
	second := svc.CalculateLiuriForDate(base, 2026, 7, 15)
	nextDay := svc.CalculateLiuriForDate(base, 2026, 7, 16)
	if first == nil || second == nil || nextDay == nil {
		t.Fatal("valid liuri query returned nil")
	}
	firstJSON, err := json.Marshal(first)
	if err != nil {
		t.Fatal(err)
	}
	secondJSON, err := json.Marshal(second)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(firstJSON, secondJSON) {
		t.Fatal("same base chart and liuri query did not produce byte-identical JSON")
	}
	if first.DerivationFingerprint == nextDay.DerivationFingerprint || first.DerivedContentHash == nextDay.DerivedContentHash {
		t.Fatal("changing the target date must change both derivation and content hashes")
	}

	changedBase := *base
	changedBase.Palaces[0].Name = "篡改宫"
	if DerivedChartMatchesBase(first, &changedBase) {
		t.Fatal("derived chart must reject a changed parent payload")
	}
	first.LiuRiStars[0] = append(first.LiuRiStars[0], "篡改星")
	if ValidDerivedChartContract(first) {
		t.Fatal("derived chart must reject changed public transit content")
	}
}

func TestDerivedChartContractCoversMonthAndRejectsInvalidDates(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}

	june := svc.CalculateLiuyueForDate(base, 2026, 6, 15)
	july := svc.CalculateLiuyueForDate(base, 2026, 7, 15)
	if june == nil || july == nil || june.DerivationInput == nil {
		t.Fatal("valid liuyue query returned nil or missing input")
	}
	if june.DerivationInput.Year != 2026 || june.DerivationInput.Month != 6 || june.DerivationInput.Day != 15 ||
		june.DerivationInput.Basis != "target_solar_date_resolved_to_lunar_month" ||
		june.DerivationInput.BoundaryPolicy != ZiWeiHoroscopeBoundaryNormal || june.DerivationInput.PeriodGanZhi == "" {
		t.Fatalf("unexpected liuyue input: %+v", *june.DerivationInput)
	}
	if june.DerivationFingerprint == july.DerivationFingerprint || june.DerivedContentHash == july.DerivedContentHash {
		t.Fatal("changing the target month must change both derivation and content hashes")
	}
	if !ValidDerivedChartContract(june) || !DerivedChartMatchesBase(june, base) {
		t.Fatal("fresh liuyue chart must satisfy the complete derivation contract")
	}

	for name, got := range map[string]*ZiWeiChart{
		"zero year":        svc.CalculateLiunian(base, 0),
		"month 13":         svc.CalculateLiuyueForDate(base, 2026, 13, 1),
		"nonexistent date": svc.CalculateLiuriForDate(base, 2026, 2, 29),
	} {
		if got != nil {
			t.Errorf("%s returned a chart instead of rejecting the invalid query", name)
		}
	}
}

func TestDerivedChartContractRejectsRehashedInvalidProjection(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		derived *ZiWeiChart
		mutate  func(*ZiWeiChart)
	}{
		{
			name:    "liunian star",
			derived: svc.CalculateLiunian(base, 2026),
			mutate: func(chart *ZiWeiChart) {
				chart.LiuNianStars[0] = append(chart.LiuNianStars[0], "篡改流曜")
			},
		},
		{
			name:    "liuyue four hua",
			derived: svc.CalculateLiuyueForDate(base, 2026, 7, 15),
			mutate: func(chart *ZiWeiChart) {
				chart.LiuYueFourHua[0] = append(chart.LiuYueFourHua[0], "篡改四化")
			},
		},
		{
			name:    "liuri palace",
			derived: svc.CalculateLiuriForDate(base, 2026, 7, 15),
			mutate: func(chart *ZiWeiChart) {
				chart.LiuRiPalaces[0] = "篡改流日宫"
			},
		},
		{
			name:    "embedded natal structure",
			derived: svc.CalculateLiunian(base, 2026),
			mutate: func(chart *ZiWeiChart) {
				palaceIdx, starIdx := findPublishedStar(chart, "紫微")
				chart.Palaces[palaceIdx].Stars[starIdx].Brightness = "篡改本命亮度"
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.derived == nil {
				t.Fatal("valid derivation returned nil")
			}
			chart := roundTripProjectionFixture(t, tc.derived)
			tc.mutate(chart)
			restampZiWeiDerivedContentHash(t, chart)
			if !validDerivedContentHashOnly(chart) {
				t.Fatal("fixture must have a valid recomputed derived content hash")
			}
			if ValidDerivedChartContract(chart) || DerivedChartMatchesBase(chart, base) {
				t.Fatal("rehashed invalid derived projection must be rejected")
			}
		})
	}

	staleQuery := roundTripProjectionFixture(t, svc.CalculateLiuriForDate(base, 2026, 7, 15))
	nextInput, err := buildZiWeiDerivationInput("liuri", 2026, 7, 16)
	if err != nil {
		t.Fatal(err)
	}
	staleQuery.DerivationInput = &nextInput
	staleQuery.DerivationFingerprint = ziweiDerivationFingerprint("liuri", nextInput)
	restampZiWeiDerivedContentHash(t, staleQuery)
	if ValidDerivedChartContract(staleQuery) {
		t.Fatal("valid next-day query metadata with stale projection must be rejected")
	}
}

func TestTransitCalculationRejectsRehashedInvalidNatalStructure(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	invalid := roundTripProjectionFixture(t, base)
	movePublishedMajorStar(invalid)
	restampZiWeiChartContentHash(t, invalid)
	if svc.CalculateLiunian(invalid, 2026) != nil ||
		svc.CalculateLiuyueForDate(invalid, 2026, 7, 15) != nil ||
		svc.CalculateLiuriForDate(invalid, 2026, 7, 15) != nil {
		t.Fatal("transit calculations must reject a rehashed invalid natal chart")
	}
}

func restampZiWeiDerivedContentHash(t *testing.T, chart *ZiWeiChart) {
	t.Helper()
	hash, err := chartContentHash(chart)
	if err != nil {
		t.Fatalf("hash derived chart: %v", err)
	}
	chart.DerivedContentHash = hash
}

func validDerivedContentHashOnly(chart *ZiWeiChart) bool {
	want, err := chartContentHash(chart)
	return err == nil && chart.DerivedContentHash == want
}
