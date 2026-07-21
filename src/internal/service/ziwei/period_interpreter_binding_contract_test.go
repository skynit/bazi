package ziwei

import "testing"

func TestPeriodInterpreterRequiresExactNatalContentHashBinding(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2003, 4, 15, 14, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	liunian := svc.CalculateLiunian(base, 2026)
	liuyue := svc.CalculateLiuyueForDate(base, 2026, 7, 15)
	liuri := svc.CalculateLiuriForDate(base, 2026, 7, 15)
	if liunian == nil || liuyue == nil || liuri == nil {
		t.Fatal("valid derived fixtures were not constructed")
	}
	birth := mustPublishedBirthData(t, base)

	for name, interpreter := range map[string]*PeriodInterpreter{
		"missing hash": {birthData: birth},
		"wrong hash":   {birthData: birth, baseContentHash: "wrong-parent-hash"},
	} {
		t.Run(name, func(t *testing.T) {
			if interpreter.AnalyzeLiunian(liunian, 2026) != nil ||
				interpreter.AnalyzeLiuyue(liuyue, 2026, 7, 15) != nil ||
				interpreter.AnalyzeLiuri(liuri, 2026, 7, 15) != nil {
				t.Fatal("unbound interpreter accepted a derived chart")
			}
		})
	}

	bound := NewPeriodInterpreterFromChart(base)
	if bound == nil || bound.AnalyzeLiunian(liunian, 2026) == nil ||
		bound.AnalyzeLiuyue(liuyue, 2026, 7, 15) == nil ||
		bound.AnalyzeLiuri(liuri, 2026, 7, 15) == nil {
		t.Fatal("authenticated chart did not create a usable bound interpreter")
	}
}
