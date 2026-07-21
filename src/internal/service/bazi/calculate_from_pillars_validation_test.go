package bazi

import (
	"strings"
	"testing"

	"github.com/6tail/tyme4go/tyme"
)

func TestCalculateFromPillarsRejectsImpossibleCrossPillarCombinations(t *testing.T) {
	service := &BaziService{}
	tests := []struct {
		name                   string
		year, month, day, hour string
		want                   string
	}{
		{
			name: "month stem violates five-tiger derivation",
			year: "甲子", month: "甲寅", day: "甲子", hour: "甲子",
			want: "expected 丙寅 by five-tiger month derivation",
		},
		{
			name: "hour stem violates five-rat derivation",
			year: "甲子", month: "丙寅", day: "甲子", hour: "戊子",
			want: "expected 甲子 for same-day/early-Zi or 丙子 for late-Zi-same-day",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := service.CalculateFromPillars(tc.year, tc.month, tc.day, tc.hour, "MALE")
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want substring %q", err, tc.want)
			}
		})
	}
}

func TestValidatePillarLinkageExhaustsFiveTigerAndFiveRatCycles(t *testing.T) {
	fixedDay := tyme.SixtyCycle{}.FromIndex(0)
	fixedHour := tyme.SixtyCycle{}.FromIndex(0)
	fixedYear := tyme.SixtyCycle{}.FromIndex(0)
	fixedMonth := tyme.SixtyCycle{}.FromIndex(2)

	for yearIndex := 0; yearIndex < 60; yearIndex++ {
		year := tyme.SixtyCycle{}.FromIndex(yearIndex)
		for branchIndex := 0; branchIndex < 12; branchIndex++ {
			branch := tyme.EarthBranch{}.FromIndex(branchIndex)
			monthOffset := branch.Next(-2).GetIndex()
			stem := tyme.HeavenStem{}.FromIndex((year.GetHeavenStem().GetIndex()+1)*2 + monthOffset)
			month, err := tyme.SixtyCycle{}.FromName(stem.GetName() + branch.GetName())
			if err != nil {
				t.Fatalf("construct month for %s/%s: %v", year.GetName(), branch.GetName(), err)
			}
			if err := validatePillarLinkage(year, *month, fixedDay, fixedHour); err != nil {
				t.Fatalf("valid year/month linkage rejected: %s %s: %v", year.GetName(), month.GetName(), err)
			}
		}
	}

	for dayIndex := 0; dayIndex < 60; dayIndex++ {
		day := tyme.SixtyCycle{}.FromIndex(dayIndex)
		for branchIndex := 0; branchIndex < 12; branchIndex++ {
			branch := tyme.EarthBranch{}.FromIndex(branchIndex)
			stem := tyme.HeavenStem{}.FromIndex(day.GetHeavenStem().GetIndex()%5*2 + branchIndex)
			hour, err := tyme.SixtyCycle{}.FromName(stem.GetName() + branch.GetName())
			if err != nil {
				t.Fatalf("construct hour for %s/%s: %v", day.GetName(), branch.GetName(), err)
			}
			if err := validatePillarLinkage(fixedYear, fixedMonth, day, *hour); err != nil {
				t.Fatalf("valid day/hour linkage rejected: %s %s: %v", day.GetName(), hour.GetName(), err)
			}
		}

		lateZiStem := fiveRatHourStem(day.Next(1).GetHeavenStem(), 0)
		lateZiHour, err := tyme.SixtyCycle{}.FromName(lateZiStem.GetName() + "子")
		if err != nil {
			t.Fatalf("construct late-Zi hour for %s: %v", day.GetName(), err)
		}
		if err := validatePillarLinkage(fixedYear, fixedMonth, day, *lateZiHour); err != nil {
			t.Fatalf("valid late-Zi-same-day linkage rejected: %s %s: %v", day.GetName(), lateZiHour.GetName(), err)
		}
	}
}

func TestCalculateFromPillarsAcceptsBothZiHourSchools(t *testing.T) {
	service := &BaziService{}
	for _, hour := range []string{"甲子", "丙子"} {
		if _, err := service.CalculateFromPillars("甲子", "丙寅", "甲子", hour, "MALE"); err != nil {
			t.Errorf("甲子日%s should be a valid explicit Zi-hour school: %v", hour, err)
		}
	}
}

func TestCalculateFromPillarsKeepsCanonicalChart(t *testing.T) {
	result, err := (&BaziService{}).CalculateFromPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
	if err != nil {
		t.Fatal(err)
	}
	if result.YearPillar.Gan+result.YearPillar.Zhi != "壬辰" ||
		result.MonthPillar.Gan+result.MonthPillar.Zhi != "壬寅" ||
		result.DayPillar.Gan+result.DayPillar.Zhi != "甲寅" ||
		result.HourPillar.Gan+result.HourPillar.Zhi != "庚午" {
		t.Fatalf("canonical chart changed: %+v %+v %+v %+v",
			result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar)
	}
}
