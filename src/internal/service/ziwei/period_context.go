package ziwei

import (
	"fmt"

	"github.com/6tail/tyme4go/tyme"
)

const (
	ZiWeiHoroscopeBoundaryNormal = "iztro_normal_lunar_boundaries_fix_leap_day_15"
)

// LunarYearLabelForSolarDate resolves the Profile's normal horoscope year
// boundary (lunar new year) for a target civil date.
func LunarYearLabelForSolarDate(year, month, day int) (int, error) {
	if year < 1 || year > 9999 {
		return 0, fmt.Errorf("year must be between 1 and 9999")
	}
	solarTime, err := tyme.SolarTime{}.FromYmdHms(year, month, day, 12, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("invalid target solar date: %w", err)
	}
	return solarTime.GetLunarHour().GetLunarDay().GetYear(), nil
}

func buildZiWeiDerivationInput(derivationType string, year, month, day int) (ZiWeiDerivationInput, error) {
	if year < 1 || year > 9999 {
		return ZiWeiDerivationInput{}, fmt.Errorf("year must be between 1 and 9999")
	}
	switch derivationType {
	case "liunian":
		lunarYear, err := tyme.LunarYear{}.FromYear(year)
		if err != nil {
			return ZiWeiDerivationInput{}, fmt.Errorf("invalid lunar year label: %w", err)
		}
		return ZiWeiDerivationInput{
			CalendarType:   "LUNAR_YEAR",
			Year:           year,
			Basis:          "target_lunar_year_label",
			BoundaryPolicy: ZiWeiHoroscopeBoundaryNormal,
			ResolvedLunarDate: ZiWeiResolvedLunarDate{
				Year: year,
			},
			PeriodGanZhi: lunarYear.GetSixtyCycle().GetName(),
		}, nil
	case "liuyue", "liuri":
		solarTime, err := tyme.SolarTime{}.FromYmdHms(year, month, day, 12, 0, 0)
		if err != nil {
			return ZiWeiDerivationInput{}, fmt.Errorf("invalid target solar date: %w", err)
		}
		lunarDay := solarTime.GetLunarHour().GetLunarDay()
		lunarMonth := lunarDay.GetLunarMonth()
		monthNumber := lunarMonth.GetMonthWithLeap()
		if monthNumber < 0 {
			monthNumber = -monthNumber
		}
		input := ZiWeiDerivationInput{
			CalendarType:   "SOLAR",
			Year:           year,
			Month:          month,
			Day:            day,
			BoundaryPolicy: ZiWeiHoroscopeBoundaryNormal,
			ResolvedLunarDate: ZiWeiResolvedLunarDate{
				Year:        lunarDay.GetYear(),
				Month:       monthNumber,
				Day:         lunarDay.GetDay(),
				IsLeapMonth: lunarMonth.IsLeap(),
			},
		}
		if derivationType == "liuyue" {
			input.Basis = "target_solar_date_resolved_to_lunar_month"
			monthCycle := lunarMonth.GetSixtyCycle()
			if lunarMonth.IsLeap() && lunarDay.GetDay() > 15 {
				monthCycle = monthCycle.Next(1)
			}
			input.PeriodGanZhi = monthCycle.GetName()
		} else {
			input.Basis = "target_solar_date_resolved_to_lunar_day"
			input.PeriodGanZhi = lunarDay.GetSixtyCycle().GetName()
		}
		return input, nil
	default:
		return ZiWeiDerivationInput{}, fmt.Errorf("unknown ziwei derivation type %q", derivationType)
	}
}

func derivationStemBranch(input ZiWeiDerivationInput) (int, int, bool) {
	stem, branch, err := parseGanZhiName(input.PeriodGanZhi)
	return stem, branch, err == nil
}

func chartDerivationStemBranch(chart *ZiWeiChart, derivationType string) (int, int, bool) {
	if chart == nil || chart.DerivationType != derivationType || chart.DerivationInput == nil ||
		!validZiWeiDerivationInput(derivationType, *chart.DerivationInput) {
		return 0, 0, false
	}
	return derivationStemBranch(*chart.DerivationInput)
}

func chartDerivationForQuery(chart *ZiWeiChart, derivationType string, year, month, day int) (int, int, bool) {
	want, err := buildZiWeiDerivationInput(derivationType, year, month, day)
	if err != nil || chart == nil || chart.DerivationInput == nil || chart.DerivationType != derivationType ||
		*chart.DerivationInput != want {
		return 0, 0, false
	}
	return derivationStemBranch(want)
}
