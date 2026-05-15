package service

import (
	"testing"
)

func TestCalculateMale1990(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(1990, 1, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if result.YearPillar.Gan == "" || result.YearPillar.Zhi == "" {
		t.Errorf("YearPillar empty: %+v", result.YearPillar)
	}
	if result.MonthPillar.Gan == "" || result.MonthPillar.Zhi == "" {
		t.Errorf("MonthPillar empty: %+v", result.MonthPillar)
	}
	if result.DayPillar.Gan == "" || result.DayPillar.Zhi == "" {
		t.Errorf("DayPillar empty: %+v", result.DayPillar)
	}
	if result.HourPillar.Gan == "" || result.HourPillar.Zhi == "" {
		t.Errorf("HourPillar empty: %+v", result.HourPillar)
	}

	if len(result.FiveElements) != 5 {
		t.Errorf("FiveElements has %d keys, want 5: %v", len(result.FiveElements), result.FiveElements)
	}
	for _, elem := range []string{"木", "火", "土", "金", "水"} {
		if _, ok := result.FiveElements[elem]; !ok {
			t.Errorf("FiveElements missing key %q", elem)
		}
	}

	t.Logf("Pillars: Y=%s%s M=%s%s D=%s%s H=%s%s",
		result.YearPillar.Gan, result.YearPillar.Zhi,
		result.MonthPillar.Gan, result.MonthPillar.Zhi,
		result.DayPillar.Gan, result.DayPillar.Zhi,
		result.HourPillar.Gan, result.HourPillar.Zhi)
	t.Logf("FiveElements: %v", result.FiveElements)
	t.Logf("TenGods: %v", result.TenGods)
}

func TestCalculateFemale2000(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.Calculate(2000, 6, 1, 12, 0, "FEMALE")
	if err != nil {
		t.Fatalf("Calculate failed: %v", err)
	}

	if result.DaYunInfo.StartAge <= 0 {
		t.Errorf("DaYun StartAge is %d, expected > 0", result.DaYunInfo.StartAge)
	}
	if len(result.DaYunInfo.Pillars) != 8 {
		t.Errorf("DaYun has %d pillars, want 8", len(result.DaYunInfo.Pillars))
	}

	t.Logf("DaYun: direction=%s startAge=%d pillars=%v",
		result.DaYunInfo.Direction, result.DaYunInfo.StartAge, result.DaYunInfo.Pillars)
	t.Logf("NaYin: %v", result.NaYin)
	t.Logf("HiddenStems: %v", result.HiddenStems)
}

func TestDaYunGenderDiff(t *testing.T) {
	svc := &BaziService{}
	resultM, err := svc.Calculate(1990, 1, 15, 8, 0, "MALE")
	if err != nil {
		t.Fatalf("Calculate MALE failed: %v", err)
	}
	resultF, err := svc.Calculate(1990, 1, 15, 8, 0, "FEMALE")
	if err != nil {
		t.Fatalf("Calculate FEMALE failed: %v", err)
	}

	if resultM.DaYunInfo.Direction == resultF.DaYunInfo.Direction {
		t.Errorf("Expected different DaYun direction for MALE vs FEMALE, both are %q",
			resultM.DaYunInfo.Direction)
	}

	t.Logf("MALE: direction=%s startAge=%d firstPillar=%s%s",
		resultM.DaYunInfo.Direction, resultM.DaYunInfo.StartAge,
		resultM.DaYunInfo.Pillars[0].Gan, resultM.DaYunInfo.Pillars[0].Zhi)
	t.Logf("FEMALE: direction=%s startAge=%d firstPillar=%s%s",
		resultF.DaYunInfo.Direction, resultF.DaYunInfo.StartAge,
		resultF.DaYunInfo.Pillars[0].Gan, resultF.DaYunInfo.Pillars[0].Zhi)
}
