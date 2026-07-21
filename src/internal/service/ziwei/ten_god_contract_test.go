package ziwei

import (
	"testing"

	"bazi/internal/service/bazi"
)

var validPeriodTenGods = map[string]struct{}{
	"比肩": {}, "劫财": {}, "食神": {}, "伤官": {}, "偏财": {},
	"正财": {}, "七杀": {}, "正官": {}, "偏印": {}, "正印": {},
}

func TestPeriodTenGodMatchesAuthoritativeBaziClassifier(t *testing.T) {
	for dayStem, dayName := range StemNames {
		for stem, stemName := range StemNames {
			got, ok := getShiShen(stem, dayStem)
			want := bazi.ClassifyTenGod(stemName, dayName, false)
			if !ok || got != want {
				t.Fatalf("stem=%s day_stem=%s ten god = %q/%t, want %q", stemName, dayName, got, ok, want)
			}
		}
	}
}

func TestPeriodTenGodRejectsInvalidStemIndexes(t *testing.T) {
	for _, input := range [][2]int{{-1, 0}, {len(StemNames), 0}, {0, -1}, {0, len(StemNames)}} {
		if got, ok := getShiShen(input[0], input[1]); ok || got != "" {
			t.Fatalf("invalid stem indexes %d/%d classified as %q/%t", input[0], input[1], got, ok)
		}
	}
}

func TestPeriodLayersOnlyPublishCanonicalTenGods(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	liunian := svc.CalculateLiunian(base, 2026)
	liuyue := svc.CalculateLiuyueForDate(base, 2026, 7, 15)
	liuri := svc.CalculateLiuriForDate(base, 2026, 7, 15)
	interpreter := NewPeriodInterpreterFromChart(base)
	if liunian == nil || liuyue == nil || liuri == nil || interpreter == nil {
		t.Fatal("valid period fixtures were not constructed")
	}

	yearResult := interpreter.AnalyzeLiunian(liunian, 2026)
	monthResult := interpreter.AnalyzeLiuyue(liuyue, 2026, 7, 15)
	dayResult := interpreter.AnalyzeLiuri(liuri, 2026, 7, 15)
	if yearResult == nil || monthResult == nil || dayResult == nil {
		t.Fatal("valid period chart was rejected")
	}
	for layer, tenGod := range map[string]string{
		"liunian": yearResult.ShiShen,
		"liuyue":  monthResult.ShiShen,
		"liuri":   dayResult.ShiShen,
	} {
		if _, ok := validPeriodTenGods[tenGod]; !ok {
			t.Fatalf("%s published invalid ten god %q", layer, tenGod)
		}
	}
	if len(dayResult.HourlyAnalysis) != len(BranchNames) {
		t.Fatalf("hour blocks = %d, want %d", len(dayResult.HourlyAnalysis), len(BranchNames))
	}
	for _, block := range dayResult.HourlyAnalysis {
		if _, ok := validPeriodTenGods[block.ShiShen]; !ok {
			t.Fatalf("hour %s published invalid ten god %q", block.StemBranch, block.ShiShen)
		}
	}
}

func TestPeriodLayersRejectInvalidNatalDayStem(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	liunian := svc.CalculateLiunian(base, 2026)
	liuyue := svc.CalculateLiuyueForDate(base, 2026, 7, 15)
	liuri := svc.CalculateLiuriForDate(base, 2026, 7, 15)
	invalidBirth := *mustPublishedBirthData(t, base)
	invalidBirth.DayStem = -1
	interpreter := &PeriodInterpreter{birthData: &invalidBirth, baseContentHash: base.ContentHash}

	if interpreter.AnalyzeLiunian(liunian, 2026) != nil ||
		interpreter.AnalyzeLiuyue(liuyue, 2026, 7, 15) != nil ||
		interpreter.AnalyzeLiuri(liuri, 2026, 7, 15) != nil {
		t.Fatal("period layer accepted an invalid natal day stem")
	}
}
