package bazi

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

const uncertaintyProbeSpanSeconds = 60 * 60

// BirthUncertainty describes the evaluated clock interval and the four-pillar
// boundaries found inside it. Ranges are inclusive to one-second precision.
type BirthUncertainty struct {
	Seconds                     int                   `json:"seconds"`
	AlgorithmUncertaintySeconds int                   `json:"algorithm_uncertainty_seconds"`
	EffectiveSeconds            int                   `json:"effective_seconds"`
	InputRangeStart             string                `json:"input_range_start"`
	InputRangeEnd               string                `json:"input_range_end"`
	EvaluationRangeStart        string                `json:"evaluation_range_start"`
	EvaluationRangeEnd          string                `json:"evaluation_range_end"`
	CalculationRangeStart       string                `json:"calculation_range_start"`
	CalculationRangeEnd         string                `json:"calculation_range_end"`
	CrossedBoundaries           []UncertaintyBoundary `json:"crossed_boundaries"`
}

type UncertaintyBoundary struct {
	Type string `json:"type"`
	At   string `json:"at"`
	Name string `json:"name,omitempty"`
}

// BirthChartCandidate is one maximal continuous input-time range whose four
// pillars are identical. DaYun start time is represented as a range because it
// changes continuously inside a candidate and must not create one chart per second.
type BirthChartCandidate struct {
	CandidateID           string
	InputRangeStart       string
	InputRangeEnd         string
	CalculationRangeStart string
	CalculationRangeEnd   string
	RepresentativeTime    string
	Normalized            *NormalizedBirth
	Result                *BaziResult
	DaYunStartAtMin       string
	DaYunStartAtMax       string
	offsetStart           int
	offsetEnd             int
}

type BirthCandidateSet struct {
	Center                     *NormalizedBirth
	CenterResult               *BaziResult
	Uncertainty                BirthUncertainty
	Candidates                 []BirthChartCandidate
	StableFields               []string
	UnstableFields             []string
	RequiresCandidateSelection bool
}

type candidatePoint struct {
	offset     int
	inputTime  time.Time
	normalized *NormalizedBirth
	result     *BaziResult
}

type candidateRange struct {
	start int
	end   int
	key   string
}

// CalculateBirthCandidates normalizes both ends of an uncertain input interval
// independently, then partitions it by exact second into continuous four-pillar
// states. The true-solar nominal error is conservatively added to the evaluated
// interval when that algorithm is active and within its documented range.
func CalculateBirthCandidates(service *BaziService, input BirthInput) (*BirthCandidateSet, error) {
	if service == nil {
		return nil, fmt.Errorf("bazi service is not available")
	}
	center, err := NormalizeBirthInput(input)
	if err != nil {
		return nil, err
	}
	centerResult, err := service.CalculateNormalizedBirth(center)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate center chart: %w", err)
	}

	location, err := time.LoadLocation(center.Validation.Timezone)
	if err != nil {
		return nil, fmt.Errorf("invalid normalized timezone %q", center.Validation.Timezone)
	}
	baseInstant, err := time.Parse(time.RFC3339, center.Validation.UTCDateTime)
	if err != nil {
		return nil, fmt.Errorf("invalid normalized UTC datetime: %w", err)
	}
	baseTime := baseInstant.In(location)

	algorithmUncertainty := 0
	if center.Validation.TrueSolarTimeApplied && center.Validation.TrueSolarWithinValidatedRange {
		algorithmUncertainty = center.Validation.TrueSolarUncertaintySeconds
	}
	effectiveSeconds := input.UncertaintySeconds + algorithmUncertainty
	cache := make(map[int]*candidatePoint)
	pointAt := func(offset int) (*candidatePoint, error) {
		if point, ok := cache[offset]; ok {
			return point, nil
		}
		at := baseTime.Add(time.Duration(offset) * time.Second)
		sample := input
		sample.Year, sample.Month, sample.Day = at.Year(), int(at.Month()), at.Day()
		sample.Hour, sample.Minute, sample.Second = at.Clock()
		sample.CalendarType = "SOLAR"
		sample.LunarLeapMonth = false
		_, atOffset := at.Zone()
		sample.UTCOffsetSeconds = &atOffset
		normalized, normalizeErr := NormalizeBirthInput(sample)
		if normalizeErr != nil {
			return nil, fmt.Errorf("failed to normalize uncertainty offset %+d seconds: %w", offset, normalizeErr)
		}
		result, calculateErr := service.CalculateNormalizedBirth(normalized)
		if calculateErr != nil {
			return nil, fmt.Errorf("failed to calculate uncertainty offset %+d seconds: %w", offset, calculateErr)
		}
		point := &candidatePoint{offset: offset, inputTime: at, normalized: normalized, result: result}
		cache[offset] = point
		return point, nil
	}

	startOffset, endOffset := -effectiveSeconds, effectiveSeconds
	ranges, err := partitionCandidateRanges(pointAt, startOffset, endOffset)
	if err != nil {
		return nil, err
	}
	candidates := make([]BirthChartCandidate, 0, len(ranges))
	for _, interval := range ranges {
		start, startErr := pointAt(interval.start)
		if startErr != nil {
			return nil, startErr
		}
		end, endErr := pointAt(interval.end)
		if endErr != nil {
			return nil, endErr
		}
		representative, representativeErr := pointAt(interval.start + (interval.end-interval.start)/2)
		if representativeErr != nil {
			return nil, representativeErr
		}
		representative.normalized.Validation.InputCalendar = center.Validation.InputCalendar
		representative.normalized.Validation.OriginalDateTime = center.Validation.OriginalDateTime
		startAtMin, startAtMax := orderedStrings(start.result.DaYunInfo.StartAt, end.result.DaYunInfo.StartAt)
		candidate := BirthChartCandidate{
			InputRangeStart:       formatUncertaintyTime(start.inputTime),
			InputRangeEnd:         formatUncertaintyTime(end.inputTime),
			CalculationRangeStart: start.normalized.Validation.CalculationDateTime,
			CalculationRangeEnd:   end.normalized.Validation.CalculationDateTime,
			RepresentativeTime:    representative.normalized.Validation.CalculationDateTime,
			Normalized:            representative.normalized,
			Result:                representative.result,
			DaYunStartAtMin:       startAtMin,
			DaYunStartAtMax:       startAtMax,
			offsetStart:           interval.start,
			offsetEnd:             interval.end,
		}
		candidate.CandidateID = candidateID(input, candidate)
		candidates = append(candidates, candidate)
	}

	inputStart := baseTime.Add(-time.Duration(input.UncertaintySeconds) * time.Second)
	inputEnd := baseTime.Add(time.Duration(input.UncertaintySeconds) * time.Second)
	evaluationStart, evaluationEnd := baseTime.Add(time.Duration(startOffset)*time.Second), baseTime.Add(time.Duration(endOffset)*time.Second)
	first, firstErr := pointAt(startOffset)
	if firstErr != nil {
		return nil, firstErr
	}
	last, lastErr := pointAt(endOffset)
	if lastErr != nil {
		return nil, lastErr
	}
	stable, unstable := candidateFieldStability(candidates)
	return &BirthCandidateSet{
		Center:       center,
		CenterResult: centerResult,
		Uncertainty: BirthUncertainty{
			Seconds:                     input.UncertaintySeconds,
			AlgorithmUncertaintySeconds: algorithmUncertainty,
			EffectiveSeconds:            effectiveSeconds,
			InputRangeStart:             formatUncertaintyTime(inputStart),
			InputRangeEnd:               formatUncertaintyTime(inputEnd),
			EvaluationRangeStart:        formatUncertaintyTime(evaluationStart),
			EvaluationRangeEnd:          formatUncertaintyTime(evaluationEnd),
			CalculationRangeStart:       first.normalized.Validation.CalculationDateTime,
			CalculationRangeEnd:         last.normalized.Validation.CalculationDateTime,
			CrossedBoundaries:           candidateBoundaries(candidates),
		},
		Candidates:                 candidates,
		StableFields:               stable,
		UnstableFields:             unstable,
		RequiresCandidateSelection: len(candidates) > 1,
	}, nil
}

func partitionCandidateRanges(pointAt func(int) (*candidatePoint, error), start, end int) ([]candidateRange, error) {
	var visit func(int, int) ([]candidateRange, error)
	visit = func(lo, hi int) ([]candidateRange, error) {
		left, err := pointAt(lo)
		if err != nil {
			return nil, err
		}
		right, err := pointAt(hi)
		if err != nil {
			return nil, err
		}
		leftKey, rightKey := fourPillarKey(left.result), fourPillarKey(right.result)
		_, leftUTCOffset := left.inputTime.Zone()
		_, rightUTCOffset := right.inputTime.Zone()
		if leftKey == rightKey && leftUTCOffset == rightUTCOffset && hi-lo <= uncertaintyProbeSpanSeconds {
			return []candidateRange{{start: lo, end: hi, key: leftKey}}, nil
		}
		if lo == hi {
			return []candidateRange{{start: lo, end: hi, key: leftKey}}, nil
		}
		mid := lo + (hi-lo)/2
		leftRanges, err := visit(lo, mid)
		if err != nil {
			return nil, err
		}
		rightRanges, err := visit(mid+1, hi)
		if err != nil {
			return nil, err
		}
		return mergeCandidateRanges(append(leftRanges, rightRanges...)), nil
	}
	return visit(start, end)
}

func mergeCandidateRanges(ranges []candidateRange) []candidateRange {
	merged := make([]candidateRange, 0, len(ranges))
	for _, current := range ranges {
		if len(merged) > 0 && merged[len(merged)-1].key == current.key && merged[len(merged)-1].end+1 == current.start {
			merged[len(merged)-1].end = current.end
			continue
		}
		merged = append(merged, current)
	}
	return merged
}

func fourPillarKey(result *BaziResult) string {
	return strings.Join([]string{
		result.YearPillar.Gan, result.YearPillar.Zhi,
		result.MonthPillar.Gan, result.MonthPillar.Zhi,
		result.DayPillar.Gan, result.DayPillar.Zhi,
		result.HourPillar.Gan, result.HourPillar.Zhi,
	}, "|")
}

func candidateID(input BirthInput, candidate BirthChartCandidate) string {
	longitude := ""
	if input.Longitude != nil {
		longitude = fmt.Sprintf("%.8f", *input.Longitude)
	}
	utcOffset := ""
	if input.UTCOffsetSeconds != nil {
		utcOffset = fmt.Sprintf("%d", *input.UTCOffsetSeconds)
	}
	ziHourPolicy, _ := NormalizeZiHourPolicy(input.ZiHourPolicy)
	identity := fmt.Sprintf("%s|%s|%s|%04d-%02d-%02dT%02d:%02d:%02d|%s|%t|%s|%s|%s|%s|%s|%t|%d|%s|%s|%s|%s|%s",
		BirthNormalizationVersion, CalendarEngineVersion, EngineVersion,
		input.Year, input.Month, input.Day, input.Hour, input.Minute, input.Second,
		strings.ToUpper(strings.TrimSpace(input.CalendarType)), input.LunarLeapMonth,
		strings.ToUpper(strings.TrimSpace(input.Gender)), ziHourPolicy, strings.TrimSpace(input.Timezone), utcOffset, longitude,
		input.UseTrueSolarTime, input.UncertaintySeconds, candidate.InputRangeStart, candidate.InputRangeEnd,
		candidate.CalculationRangeStart, candidate.CalculationRangeEnd, fourPillarKey(candidate.Result))
	sum := sha256.Sum256([]byte(identity))
	return hex.EncodeToString(sum[:16])
}

func candidateFieldStability(candidates []BirthChartCandidate) ([]string, []string) {
	fields := []struct {
		name  string
		value func(*BaziResult) string
	}{
		{"year_pillar", func(r *BaziResult) string { return r.YearPillar.Gan + r.YearPillar.Zhi }},
		{"month_pillar", func(r *BaziResult) string { return r.MonthPillar.Gan + r.MonthPillar.Zhi }},
		{"day_pillar", func(r *BaziResult) string { return r.DayPillar.Gan + r.DayPillar.Zhi }},
		{"hour_pillar", func(r *BaziResult) string { return r.HourPillar.Gan + r.HourPillar.Zhi }},
	}
	stable, unstable := make([]string, 0, 4), make([]string, 0, 4)
	for _, field := range fields {
		isStable := len(candidates) > 0
		if isStable {
			first := field.value(candidates[0].Result)
			for _, candidate := range candidates[1:] {
				if field.value(candidate.Result) != first {
					isStable = false
					break
				}
			}
		}
		if isStable {
			stable = append(stable, field.name)
		} else {
			unstable = append(unstable, field.name)
		}
	}
	return stable, unstable
}

func candidateBoundaries(candidates []BirthChartCandidate) []UncertaintyBoundary {
	boundaries := make([]UncertaintyBoundary, 0, len(candidates)-1)
	for i := 1; i < len(candidates); i++ {
		previous, current := candidates[i-1], candidates[i]
		typeName, name := "hour_branch", "时辰交界"
		if previous.Result.YearPillar != current.Result.YearPillar {
			typeName, name = "solar_term", "立春"
		} else if previous.Result.MonthPillar != current.Result.MonthPillar {
			typeName, name = "solar_term", current.Normalized.Validation.CurrentSolarTerm
		} else if previous.Result.DayPillar != current.Result.DayPillar {
			if current.Normalized.ZiHourPolicy == ZiHourLateZiNextDay {
				typeName, name = "zi_hour_day_boundary", "子初换日"
			} else {
				typeName, name = "civil_day", "午夜换日"
			}
		}
		boundaries = append(boundaries, UncertaintyBoundary{Type: typeName, At: current.CalculationRangeStart, Name: name})
	}
	return boundaries
}

func orderedStrings(left, right string) (string, string) {
	values := []string{left, right}
	sort.Strings(values)
	return values[0], values[1]
}

func formatUncertaintyTime(value time.Time) string {
	return value.Format("2006-01-02 15:04:05 -07:00")
}
