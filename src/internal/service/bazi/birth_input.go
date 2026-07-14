package bazi

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/6tail/tyme4go/tyme"
)

const (
	BirthNormalizationVersion = "birth-normalization-2026-07-10"
	EngineVersion             = "bazi-engine-2026-07-10"
	DefaultBirthTimezone      = "Asia/Shanghai"
)

// BirthInput contains the factual birth information supplied by a user.
type BirthInput struct {
	Year             int
	Month            int
	Day              int
	Hour             int
	Minute           int
	CalendarType     string
	LunarLeapMonth   bool
	Gender           string
	BirthPlace       string
	Timezone         string
	Longitude        *float64
	UseTrueSolarTime bool
	TimeUncertain    bool
}

// BirthValidation is returned to the user before a chart is persisted.
// It deliberately contains calendar and BaZi facts only, not a Zi Wei chart.
type BirthValidation struct {
	NormalizationVersion       string   `json:"normalization_version"`
	InputCalendar              string   `json:"input_calendar"`
	OriginalDateTime           string   `json:"original_date_time"`
	ConvertedSolarDateTime     string   `json:"converted_solar_date_time"`
	CalculationDateTime        string   `json:"calculation_date_time"`
	LunarDate                  string   `json:"lunar_date"`
	CurrentSolarTerm           string   `json:"current_solar_term"`
	CurrentSolarTermStartedAt  string   `json:"current_solar_term_started_at"`
	BirthPlace                 string   `json:"birth_place,omitempty"`
	Timezone                   string   `json:"timezone"`
	UTCDateTime                string   `json:"utc_date_time"`
	Longitude                  *float64 `json:"longitude,omitempty"`
	TrueSolarTimeApplied       bool     `json:"true_solar_time_applied"`
	TrueSolarAdjustmentMinutes int      `json:"true_solar_adjustment_minutes"`
	TimeUncertain              bool     `json:"time_uncertain"`
	Notices                    []string `json:"notices"`
}

// NormalizedBirth is the deterministic input used by the BaZi engine.
type NormalizedBirth struct {
	Year       int             `json:"year"`
	Month      int             `json:"month"`
	Day        int             `json:"day"`
	Hour       int             `json:"hour"`
	Minute     int             `json:"minute"`
	Gender     string          `json:"gender"`
	Validation BirthValidation `json:"validation"`
}

// NormalizeBirthInput converts lunar input to solar dates, validates timezone
// facts, and optionally applies apparent solar time before chart calculation.
func NormalizeBirthInput(input BirthInput) (*NormalizedBirth, error) {
	calendarType := strings.ToUpper(strings.TrimSpace(input.CalendarType))
	if calendarType == "" {
		calendarType = "SOLAR"
	}
	if calendarType != "SOLAR" && calendarType != "LUNAR" {
		return nil, fmt.Errorf("calendar_type must be SOLAR or LUNAR")
	}

	gender := strings.ToUpper(strings.TrimSpace(input.Gender))
	if gender != "MALE" && gender != "FEMALE" {
		return nil, fmt.Errorf("gender must be MALE or FEMALE")
	}
	if input.Hour < 0 || input.Hour > 23 || input.Minute < 0 || input.Minute > 59 {
		return nil, fmt.Errorf("birth time is out of range")
	}

	timezone := strings.TrimSpace(input.Timezone)
	if timezone == "" {
		timezone = DefaultBirthTimezone
	}
	location, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid timezone %q", timezone)
	}

	solarYear, solarMonth, solarDay := input.Year, input.Month, input.Day
	if calendarType == "LUNAR" {
		lunarMonth := input.Month
		if input.LunarLeapMonth {
			lunarMonth = -lunarMonth
		}
		lunarDay, lunarErr := tyme.LunarDay{}.FromYmd(input.Year, lunarMonth, input.Day)
		if lunarErr != nil {
			return nil, fmt.Errorf("invalid lunar date: %w", lunarErr)
		}
		converted := lunarDay.GetSolarDay()
		solarYear, solarMonth, solarDay = converted.GetYear(), converted.GetMonth(), converted.GetDay()
	} else {
		if _, solarErr := (tyme.SolarDay{}).FromYmd(input.Year, input.Month, input.Day); solarErr != nil {
			return nil, fmt.Errorf("invalid solar date: %w", solarErr)
		}
		if input.LunarLeapMonth {
			return nil, fmt.Errorf("lunar_leap_month is only valid for lunar dates")
		}
	}

	localCivil := time.Date(solarYear, time.Month(solarMonth), solarDay, input.Hour, input.Minute, 0, 0, location)
	// time.Date normalizes nonexistent local times during DST changes. Reject them
	// instead of silently calculating a different birth time.
	ly, lm, ld := localCivil.Date()
	lh, lmin, _ := localCivil.Clock()
	if ly != solarYear || int(lm) != solarMonth || ld != solarDay || lh != input.Hour || lmin != input.Minute {
		return nil, fmt.Errorf("birth time does not exist in timezone %s due to a clock transition", timezone)
	}

	calculationTime := localCivil
	adjustmentMinutes := 0
	notices := make([]string, 0, 3)
	if input.UseTrueSolarTime {
		if input.Longitude == nil || math.IsNaN(*input.Longitude) || math.IsInf(*input.Longitude, 0) || *input.Longitude < -180 || *input.Longitude > 180 {
			return nil, fmt.Errorf("a longitude between -180 and 180 is required for true solar time")
		}
		calculationTime = apparentSolarTime(localCivil, *input.Longitude)
		adjustmentMinutes = int(math.Round(calculationTime.Sub(localCivil).Minutes()))
		notices = append(notices, "已按出生地经度和当日均时差换算真太阳时；经度误差会直接影响校正结果。")
	}
	if input.TimeUncertain {
		notices = append(notices, "出生时间标记为不确定；若接近时辰或日期边界，建议分别保存相邻候选命盘进行核对。")
	}
	if calculationTime.Day() != localCivil.Day() || calculationTime.Month() != localCivil.Month() || calculationTime.Year() != localCivil.Year() {
		notices = append(notices, "时间校正跨越了日期边界，请重点核对最终采用时间。")
	}

	cy, cm, cd := calculationTime.Date()
	ch, cmin, _ := calculationTime.Clock()
	st, err := tyme.SolarTime{}.FromYmdHms(cy, int(cm), cd, ch, cmin, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid normalized birth time: %w", err)
	}
	term := st.GetTerm()
	lunarDate := st.GetSolarDay().GetLunarDay().String()
	longitude := cloneFloat(input.Longitude)
	originalDate := fmt.Sprintf("%04d-%02d-%02d %02d:%02d", input.Year, input.Month, input.Day, input.Hour, input.Minute)
	if calendarType == "LUNAR" && input.LunarLeapMonth {
		originalDate += "（闰月）"
	}

	return &NormalizedBirth{
		Year:   cy,
		Month:  int(cm),
		Day:    cd,
		Hour:   ch,
		Minute: cmin,
		Gender: gender,
		Validation: BirthValidation{
			NormalizationVersion:       BirthNormalizationVersion,
			InputCalendar:              calendarType,
			OriginalDateTime:           originalDate,
			ConvertedSolarDateTime:     localCivil.Format("2006-01-02 15:04"),
			CalculationDateTime:        calculationTime.Format("2006-01-02 15:04"),
			LunarDate:                  lunarDate,
			CurrentSolarTerm:           term.GetName(),
			CurrentSolarTermStartedAt:  term.GetJulianDay().GetSolarTime().String(),
			BirthPlace:                 strings.TrimSpace(input.BirthPlace),
			Timezone:                   timezone,
			UTCDateTime:                localCivil.UTC().Format(time.RFC3339),
			Longitude:                  longitude,
			TrueSolarTimeApplied:       input.UseTrueSolarTime,
			TrueSolarAdjustmentMinutes: adjustmentMinutes,
			TimeUncertain:              input.TimeUncertain,
			Notices:                    notices,
		},
	}, nil
}

// apparentSolarTime converts an instant to local apparent solar time at the
// supplied longitude. The equation-of-time approximation is adequate to within
// roughly one minute for birth-time boundary checking.
func apparentSolarTime(localCivil time.Time, longitude float64) time.Time {
	utc := localCivil.UTC()
	day := float64(utc.YearDay())
	b := 2 * math.Pi * (day - 81) / 365
	equationOfTimeMinutes := 9.87*math.Sin(2*b) - 7.53*math.Cos(b) - 1.5*math.Sin(b)
	solarOffsetSeconds := longitude*4*60 + equationOfTimeMinutes*60
	solarClock := utc.Add(time.Duration(math.Round(solarOffsetSeconds)) * time.Second)
	return time.Date(solarClock.Year(), solarClock.Month(), solarClock.Day(), solarClock.Hour(), solarClock.Minute(), solarClock.Second(), 0, localCivil.Location())
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
