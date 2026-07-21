package bazi

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/6tail/tyme4go/tyme"
)

const (
	BirthNormalizationVersion      = "birth-normalization-2026-07-18.2"
	CalendarEngineVersion          = "bazi-calendar-2026-07-18.1"
	EngineVersion                  = "bazi-engine-2026-07-17.27"
	DefaultBirthTimezone           = "Asia/Shanghai"
	MaxBirthUncertaintySeconds     = 24 * 60 * 60
	trueSolarAlgorithm             = "usno-approx-solar-coordinates-j2000"
	trueSolarSource                = "USNO Astronomical Applications Department: Approximate Solar Coordinates"
	trueSolarValidatedFrom         = 1800
	trueSolarValidatedThrough      = 2200
	trueSolarCoordinateUncertainty = 4
	trueSolarUTCUT1Allowance       = 1
	trueSolarRoundingAllowance     = 1
	trueSolarUncertainty           = trueSolarCoordinateUncertainty + trueSolarUTCUT1Allowance + trueSolarRoundingAllowance
)

var chinaStandardTime = time.FixedZone("CST", 8*60*60)

// BirthInput contains the factual birth information supplied by a user.
type BirthInput struct {
	Year               int
	Month              int
	Day                int
	Hour               int
	Minute             int
	Second             int
	CalendarType       string
	LunarLeapMonth     bool
	Gender             string
	ZiHourPolicy       string
	BirthPlace         string
	Timezone           string
	UTCOffsetSeconds   *int
	Longitude          *float64
	UseTrueSolarTime   bool
	TimeUncertain      bool
	UncertaintySeconds int
}

// BirthValidation is returned to the user before a chart is persisted.
// It deliberately contains calendar and BaZi facts only, not a Zi Wei chart.
type BirthValidation struct {
	NormalizationVersion          string   `json:"normalization_version"`
	CalendarEngineVersion         string   `json:"calendar_engine_version"`
	InputCalendar                 string   `json:"input_calendar"`
	OriginalDateTime              string   `json:"original_date_time"`
	ConvertedSolarDateTime        string   `json:"converted_solar_date_time"`
	CalculationDateTime           string   `json:"calculation_date_time"`
	SolarTermReferenceDateTime    string   `json:"solar_term_reference_date_time"`
	SolarTermReferenceTimezone    string   `json:"solar_term_reference_timezone"`
	SolarTermTimeBasis            string   `json:"solar_term_time_basis"`
	LunarDate                     string   `json:"lunar_date"`
	CurrentSolarTerm              string   `json:"current_solar_term"`
	CurrentSolarTermStartedAt     string   `json:"current_solar_term_started_at"`
	BirthPlace                    string   `json:"birth_place,omitempty"`
	Timezone                      string   `json:"timezone"`
	UTCDateTime                   string   `json:"utc_date_time"`
	LocalTimeAmbiguous            bool     `json:"local_time_ambiguous"`
	PossibleUTCOffsetSeconds      []int    `json:"possible_utc_offset_seconds"`
	Longitude                     *float64 `json:"longitude,omitempty"`
	TrueSolarTimeApplied          bool     `json:"true_solar_time_applied"`
	TrueSolarAdjustmentMinutes    int      `json:"true_solar_adjustment_minutes"`
	TimezoneOffsetSeconds         int      `json:"timezone_offset_seconds"`
	MeanSolarAdjustmentSeconds    int      `json:"mean_solar_adjustment_seconds"`
	EquationOfTimeSeconds         int      `json:"equation_of_time_seconds"`
	TrueSolarAdjustmentSeconds    int      `json:"true_solar_adjustment_seconds"`
	TrueSolarAlgorithm            string   `json:"true_solar_algorithm,omitempty"`
	TrueSolarSource               string   `json:"true_solar_source,omitempty"`
	TrueSolarWithinValidatedRange bool     `json:"true_solar_within_validated_range"`
	TrueSolarUncertaintySeconds   int      `json:"true_solar_uncertainty_seconds"`
	TimeUncertain                 bool     `json:"time_uncertain"`
	ZiHourPolicy                  string   `json:"zi_hour_policy"`
	UncertaintySeconds            int      `json:"uncertainty_seconds"`
	Notices                       []string `json:"notices"`
}

// SolarTermReference is the same physical birth instant expressed in China
// Standard Time, the civil-time basis used by tyme4go's solar-term table.
type SolarTermReference struct {
	Year   int `json:"year"`
	Month  int `json:"month"`
	Day    int `json:"day"`
	Hour   int `json:"hour"`
	Minute int `json:"minute"`
	Second int `json:"second"`
}

// NormalizedBirth is the deterministic input used by the BaZi engine.
type NormalizedBirth struct {
	Year               int                `json:"year"`
	Month              int                `json:"month"`
	Day                int                `json:"day"`
	Hour               int                `json:"hour"`
	Minute             int                `json:"minute"`
	Second             int                `json:"second"`
	Gender             string             `json:"gender"`
	ZiHourPolicy       string             `json:"zi_hour_policy"`
	SolarTermReference SolarTermReference `json:"solar_term_reference"`
	Validation         BirthValidation    `json:"validation"`
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
	ziHourPolicy, err := NormalizeZiHourPolicy(input.ZiHourPolicy)
	if err != nil {
		return nil, err
	}
	if input.Hour < 0 || input.Hour > 23 || input.Minute < 0 || input.Minute > 59 || input.Second < 0 || input.Second > 59 {
		return nil, fmt.Errorf("birth time is out of range")
	}
	if input.UncertaintySeconds < 0 || input.UncertaintySeconds > MaxBirthUncertaintySeconds {
		return nil, fmt.Errorf("uncertainty_seconds must be between 0 and %d", MaxBirthUncertaintySeconds)
	}
	if input.TimeUncertain != (input.UncertaintySeconds > 0) {
		return nil, fmt.Errorf("time_uncertain must equal whether uncertainty_seconds is greater than zero")
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

	localCandidates := matchingLocalTimes(location, solarYear, solarMonth, solarDay, input.Hour, input.Minute, input.Second)
	if len(localCandidates) == 0 {
		return nil, fmt.Errorf("birth time does not exist in timezone %s due to a clock transition", timezone)
	}
	possibleOffsets := make([]int, 0, len(localCandidates))
	for _, candidate := range localCandidates {
		_, offset := candidate.Zone()
		possibleOffsets = append(possibleOffsets, offset)
	}
	localCivil := localCandidates[0]
	if input.UTCOffsetSeconds == nil && len(localCandidates) > 1 {
		return nil, fmt.Errorf("birth time is ambiguous in timezone %s; birth_utc_offset_seconds must be one of %v", timezone, possibleOffsets)
	}
	if input.UTCOffsetSeconds != nil {
		matched := false
		for _, candidate := range localCandidates {
			_, offset := candidate.Zone()
			if offset == *input.UTCOffsetSeconds {
				localCivil = candidate
				matched = true
				break
			}
		}
		if !matched {
			return nil, fmt.Errorf("birth_utc_offset_seconds must be one of %v for the supplied local time", possibleOffsets)
		}
	}
	termReferenceTime := localCivil.UTC().In(chinaStandardTime)

	calculationTime := localCivil
	adjustmentMinutes := 0
	_, timezoneOffsetSeconds := localCivil.Zone()
	meanSolarAdjustmentSeconds := 0
	equationSeconds := 0
	totalAdjustmentSeconds := 0
	trueSolarWithinRange := false
	trueSolarUncertaintySeconds := 0
	algorithm := ""
	source := ""
	notices := make([]string, 0, 4)
	if input.UseTrueSolarTime {
		if input.Longitude == nil || math.IsNaN(*input.Longitude) || math.IsInf(*input.Longitude, 0) || *input.Longitude < -180 || *input.Longitude > 180 {
			return nil, fmt.Errorf("a longitude between -180 and 180 is required for true solar time")
		}
		meanSolarAdjustmentSeconds = int(math.Round(*input.Longitude*240)) - timezoneOffsetSeconds
		equationSeconds = equationOfTimeSeconds(localCivil)
		totalAdjustmentSeconds = meanSolarAdjustmentSeconds + equationSeconds
		calculationTime = addCivilClockSeconds(localCivil, totalAdjustmentSeconds)
		adjustmentMinutes = int(math.Round(float64(totalAdjustmentSeconds) / 60))
		algorithm = trueSolarAlgorithm
		source = trueSolarSource
		utcYear := localCivil.UTC().Year()
		trueSolarWithinRange = utcYear >= trueSolarValidatedFrom && utcYear <= trueSolarValidatedThrough
		if trueSolarWithinRange {
			trueSolarUncertaintySeconds = trueSolarUncertainty
			notices = append(notices, "真太阳时候选区间按6秒工程量扩展：USNO约1角分坐标精度折算4秒，另计UTC/UT1差异1秒和秒级取整1秒；这不是严格天文误差上界。")
		} else {
			notices = append(notices, "真太阳时公式的公开精度说明适用于 1800–2200 年；当前日期超出该范围，算法误差未知。")
		}
		notices = append(notices, "已按出生瞬间时区偏移、出生地经度和 USNO J2000 太阳坐标近似公式换算真太阳时；经度误差会直接影响校正结果。")
	}
	if input.TimeUncertain {
		notices = append(notices, fmt.Sprintf("出生时间按中心时刻前后各 %d 秒评估；若跨越四柱边界，必须从预览结果中选择候选命盘。", input.UncertaintySeconds))
	}
	if calculationTime.Day() != localCivil.Day() || calculationTime.Month() != localCivil.Month() || calculationTime.Year() != localCivil.Year() {
		notices = append(notices, "时间校正跨越了日期边界，请重点核对最终采用时间。")
	}

	cy, cm, cd := calculationTime.Date()
	ch, cmin, csec := calculationTime.Clock()
	st, err := tyme.SolarTime{}.FromYmdHms(cy, int(cm), cd, ch, cmin, csec)
	if err != nil {
		return nil, fmt.Errorf("invalid normalized birth time: %w", err)
	}
	ty, tm, td := termReferenceTime.Date()
	th, tmin, tsec := termReferenceTime.Clock()
	termST, err := tyme.SolarTime{}.FromYmdHms(ty, int(tm), td, th, tmin, tsec)
	if err != nil {
		return nil, fmt.Errorf("invalid solar-term reference time: %w", err)
	}
	term := termST.GetTerm()
	lunarDate := st.GetSolarDay().GetLunarDay().String()
	longitude := cloneFloat(input.Longitude)
	originalDate := fmt.Sprintf("%04d-%02d-%02d %02d:%02d:%02d", input.Year, input.Month, input.Day, input.Hour, input.Minute, input.Second)
	if calendarType == "LUNAR" && input.LunarLeapMonth {
		originalDate += "（闰月）"
	}

	return &NormalizedBirth{
		Year:         cy,
		Month:        int(cm),
		Day:          cd,
		Hour:         ch,
		Minute:       cmin,
		Second:       csec,
		Gender:       gender,
		ZiHourPolicy: ziHourPolicy,
		SolarTermReference: SolarTermReference{
			Year: ty, Month: int(tm), Day: td, Hour: th, Minute: tmin, Second: tsec,
		},
		Validation: BirthValidation{
			NormalizationVersion:          BirthNormalizationVersion,
			CalendarEngineVersion:         CalendarEngineVersion,
			InputCalendar:                 calendarType,
			OriginalDateTime:              originalDate,
			ConvertedSolarDateTime:        localCivil.Format("2006-01-02 15:04:05"),
			CalculationDateTime:           calculationTime.Format("2006-01-02 15:04:05"),
			SolarTermReferenceDateTime:    termReferenceTime.Format("2006-01-02 15:04:05"),
			SolarTermReferenceTimezone:    "UTC+08:00",
			SolarTermTimeBasis:            "birth_utc_instant_expressed_in_china_standard_time_for_tyme4go_solar_terms",
			LunarDate:                     lunarDate,
			CurrentSolarTerm:              term.GetName(),
			CurrentSolarTermStartedAt:     term.GetJulianDay().GetSolarTime().String(),
			BirthPlace:                    strings.TrimSpace(input.BirthPlace),
			Timezone:                      timezone,
			UTCDateTime:                   localCivil.UTC().Format(time.RFC3339),
			LocalTimeAmbiguous:            len(localCandidates) > 1,
			PossibleUTCOffsetSeconds:      possibleOffsets,
			Longitude:                     longitude,
			TrueSolarTimeApplied:          input.UseTrueSolarTime,
			TrueSolarAdjustmentMinutes:    adjustmentMinutes,
			TimezoneOffsetSeconds:         timezoneOffsetSeconds,
			MeanSolarAdjustmentSeconds:    meanSolarAdjustmentSeconds,
			EquationOfTimeSeconds:         equationSeconds,
			TrueSolarAdjustmentSeconds:    totalAdjustmentSeconds,
			TrueSolarAlgorithm:            algorithm,
			TrueSolarSource:               source,
			TrueSolarWithinValidatedRange: trueSolarWithinRange,
			TrueSolarUncertaintySeconds:   trueSolarUncertaintySeconds,
			TimeUncertain:                 input.TimeUncertain,
			ZiHourPolicy:                  ziHourPolicy,
			UncertaintySeconds:            input.UncertaintySeconds,
			Notices:                       notices,
		},
	}, nil
}

// matchingLocalTimes resolves a wall clock against every UTC offset active
// near that date. Repeated DST clocks return two instants; skipped clocks return
// none, avoiding time.Date's silent normalization or arbitrary selection.
func matchingLocalTimes(location *time.Location, year, month, day, hour, minute, second int) []time.Time {
	guess := time.Date(year, time.Month(month), day, hour, minute, second, 0, location)
	offsetSet := make(map[int]struct{})
	for delta := -48 * time.Hour; delta <= 48*time.Hour; delta += 30 * time.Minute {
		_, offset := guess.Add(delta).Zone()
		offsetSet[offset] = struct{}{}
	}

	wallUTC := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	candidates := make([]time.Time, 0, len(offsetSet))
	seen := make(map[int64]struct{})
	for offset := range offsetSet {
		candidate := wallUTC.Add(-time.Duration(offset) * time.Second).In(location)
		cy, cm, cd := candidate.Date()
		ch, cmin, csec := candidate.Clock()
		_, actualOffset := candidate.Zone()
		if cy != year || int(cm) != month || cd != day || ch != hour || cmin != minute || csec != second || actualOffset != offset {
			continue
		}
		if _, ok := seen[candidate.Unix()]; ok {
			continue
		}
		seen[candidate.Unix()] = struct{}{}
		candidates = append(candidates, candidate)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Before(candidates[j]) })
	return candidates
}

// equationOfTimeSeconds implements the USNO Approximate Solar Coordinates
// J2000 formula. The signed angle between the mean Sun and apparent right
// ascension is converted at 240 seconds of solar time per degree.
func equationOfTimeSeconds(at time.Time) int {
	utc := at.UTC()
	julianDate := float64(utc.Unix())/86400 + float64(utc.Nanosecond())/(86400*1e9) + 2440587.5
	daysSinceJ2000 := julianDate - 2451545.0

	meanAnomaly := degreesToRadians(normalizeDegrees(357.529 + 0.98560028*daysSinceJ2000))
	meanLongitude := normalizeDegrees(280.459 + 0.98564736*daysSinceJ2000)
	apparentLongitude := degreesToRadians(normalizeDegrees(
		meanLongitude + 1.915*math.Sin(meanAnomaly) + 0.020*math.Sin(2*meanAnomaly),
	))
	obliquity := degreesToRadians(23.439 - 0.00000036*daysSinceJ2000)
	rightAscension := normalizeDegrees(math.Atan2(
		math.Cos(obliquity)*math.Sin(apparentLongitude),
		math.Cos(apparentLongitude),
	) * 180 / math.Pi)

	return int(math.Round(normalizeSignedDegrees(meanLongitude-rightAscension) * 240))
}

// addCivilClockSeconds adjusts clock fields without allowing a nearby daylight
// saving transition to introduce an unrelated one-hour jump.
func addCivilClockSeconds(localCivil time.Time, seconds int) time.Time {
	zoneName, zoneOffset := localCivil.Zone()
	fixedZone := time.FixedZone(zoneName, zoneOffset)
	clock := time.Date(
		localCivil.Year(), localCivil.Month(), localCivil.Day(),
		localCivil.Hour(), localCivil.Minute(), localCivil.Second(), 0, fixedZone,
	)
	return clock.Add(time.Duration(seconds) * time.Second)
}

func normalizeDegrees(value float64) float64 {
	value = math.Mod(value, 360)
	if value < 0 {
		value += 360
	}
	return value
}

func normalizeSignedDegrees(value float64) float64 {
	return normalizeDegrees(value+180) - 180
}

func degreesToRadians(value float64) float64 {
	return value * math.Pi / 180
}

func cloneFloat(value *float64) *float64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
