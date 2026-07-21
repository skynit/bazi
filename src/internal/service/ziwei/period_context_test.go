package ziwei

import "testing"

func TestPeriodLayersShareCanonicalGanZhi(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2000, 8, 16, 3, 0, "女")
	if err != nil {
		t.Fatal(err)
	}
	liuyue := svc.CalculateLiuyueForDate(base, 2023, 8, 19)
	liuri := svc.CalculateLiuriForDate(base, 2023, 8, 19)
	if liuyue == nil || liuri == nil {
		t.Fatal("valid period date returned nil")
	}
	if liuyue.DerivationInput.PeriodGanZhi != "庚申" || liuri.DerivationInput.PeriodGanZhi != "己酉" {
		t.Fatalf("canonical period gan-zhi = month:%q day:%q", liuyue.DerivationInput.PeriodGanZhi, liuri.DerivationInput.PeriodGanZhi)
	}

	interpreter := NewPeriodInterpreterFromChart(base)
	if interpreter == nil {
		t.Fatal("valid natal chart did not create a period interpreter")
	}
	monthText := interpreter.AnalyzeLiuyue(liuyue, 2023, 8, 19)
	dayText := interpreter.AnalyzeLiuri(liuri, 2023, 8, 19)
	monthAnalysis := BuildLiuyueAnalysis(base, liuyue, 2023, 8, 19)
	dayAnalysis := BuildLiuriAnalysis(base, liuri, 2023, 8, 19)
	if monthText == nil || dayText == nil || monthAnalysis == nil || dayAnalysis == nil {
		t.Fatal("canonical period chart was rejected by an analysis layer")
	}
	if monthText.GanZhi != "庚申" || monthAnalysis.GanZhi != "庚申" ||
		dayText.GanZhi != "己酉" || dayAnalysis.GanZhi != "己酉" {
		t.Fatalf("period layers diverged: month=%q/%q day=%q/%q", monthText.GanZhi, monthAnalysis.GanZhi, dayText.GanZhi, dayAnalysis.GanZhi)
	}
	if interpreter.AnalyzeLiuyue(liuyue, 2023, 8, 20) != nil || BuildLiuriAnalysis(base, liuri, 2023, 8, 20) != nil {
		t.Fatal("analysis layer accepted query parameters that do not match the derived chart contract")
	}
}

func TestPeriodContextUsesIztroLeapMonthHalfBoundary(t *testing.T) {
	before, err := buildZiWeiDerivationInput("liuyue", 2020, 6, 6)
	if err != nil {
		t.Fatal(err)
	}
	after, err := buildZiWeiDerivationInput("liuyue", 2020, 6, 7)
	if err != nil {
		t.Fatal(err)
	}
	if before.ResolvedLunarDate != (ZiWeiResolvedLunarDate{Year: 2020, Month: 4, Day: 15, IsLeapMonth: true}) ||
		after.ResolvedLunarDate != (ZiWeiResolvedLunarDate{Year: 2020, Month: 4, Day: 16, IsLeapMonth: true}) {
		t.Fatalf("unexpected leap-month dates: before=%+v after=%+v", before.ResolvedLunarDate, after.ResolvedLunarDate)
	}
	if before.PeriodGanZhi != "辛巳" || after.PeriodGanZhi != "壬午" {
		t.Fatalf("leap-month half boundary = %q -> %q, want 辛巳 -> 壬午", before.PeriodGanZhi, after.PeriodGanZhi)
	}
	if before.BoundaryPolicy != ZiWeiHoroscopeBoundaryNormal || after.BoundaryPolicy != ZiWeiHoroscopeBoundaryNormal {
		t.Fatal("leap-month boundary policy is not recorded in the derivation input")
	}
}

func TestLiunianInputIsLunarYearLabel(t *testing.T) {
	input, err := buildZiWeiDerivationInput("liunian", 2025, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if input.CalendarType != "LUNAR_YEAR" || input.Basis != "target_lunar_year_label" ||
		input.ResolvedLunarDate.Year != 2025 || input.PeriodGanZhi != "乙巳" {
		t.Fatalf("unexpected liunian input: %+v", input)
	}
}

func TestLunarYearLabelUsesLunarNewYearBoundary(t *testing.T) {
	before, err := LunarYearLabelForSolarDate(2024, 2, 9)
	if err != nil {
		t.Fatal(err)
	}
	after, err := LunarYearLabelForSolarDate(2024, 2, 10)
	if err != nil {
		t.Fatal(err)
	}
	if before != 2023 || after != 2024 {
		t.Fatalf("lunar-year boundary = %d -> %d, want 2023 -> 2024", before, after)
	}
}

func TestPeriodSummaryRejectsSolarYearAsPreNewYearLiunianLabel(t *testing.T) {
	svc := NewZiWeiService()
	base, err := svc.CalculateChart(2000, 8, 16, 3, 0, "男")
	if err != nil {
		t.Fatal(err)
	}
	wrongLiunian := svc.CalculateLiunian(base, 2024)
	liuyue := svc.CalculateLiuyueForDate(base, 2024, 2, 9)
	liuri := svc.CalculateLiuriForDate(base, 2024, 2, 9)
	if wrongLiunian == nil || liuyue == nil || liuri == nil {
		t.Fatal("valid period inputs returned nil")
	}
	interpreter := NewPeriodInterpreterFromChart(base)
	if interpreter == nil {
		t.Fatal("valid natal chart did not create a period interpreter")
	}
	if got := interpreter.SummarizeAll(wrongLiunian, liuyue, liuri, 2024, 2, 9); got != nil {
		t.Fatalf("summary accepted solar-year liunian before lunar new year: %+v", got)
	}

	correctLiunian := svc.CalculateLiunian(base, 2023)
	if got := interpreter.SummarizeAll(correctLiunian, liuyue, liuri, 2024, 2, 9); got == nil || got.Liunian.GanZhi != "癸卯" {
		t.Fatalf("summary rejected correct lunar-year context: %+v", got)
	}
}
