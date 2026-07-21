package bazi

import (
	"encoding/json"
	"fmt"
	"strings"

	"bazi/internal/model"
)

const (
	StoredBirthSourceNormalized = "normalized_birth"
	StoredBirthSourceRaw        = "normalized_raw_birth"
)

type StoredBirthResolution struct {
	Normalized *NormalizedBirth
	Result     *BaziResult
	Source     string
}

// ResolveStoredBirth reconstructs the exact normalized input selected when a
// chart was saved. It never treats raw lunar fields as a solar date.
func ResolveStoredBirth(service *BaziService, chart *model.BirthChart) (*StoredBirthResolution, error) {
	if service == nil {
		return nil, fmt.Errorf("bazi service is not available")
	}
	if chart == nil {
		return nil, fmt.Errorf("birth chart is nil")
	}

	if normalized, ok := decodeStoredNormalizedBirth(chart.NormalizedBirth); ok {
		result, err := service.CalculateNormalizedBirth(normalized)
		if err == nil {
			return &StoredBirthResolution{Normalized: normalized, Result: result, Source: StoredBirthSourceNormalized}, nil
		}
	}

	gender, ok := normalizeStoredGender(chart.Gender)
	if !ok {
		return nil, fmt.Errorf("stored birth chart has invalid gender %q", chart.Gender)
	}
	input := BirthInput{
		Year:               chart.BirthYear,
		Month:              chart.BirthMonth,
		Day:                chart.BirthDay,
		Hour:               chart.BirthHour,
		Minute:             chart.BirthMin,
		Second:             chart.BirthSec,
		CalendarType:       chart.CalendarType,
		LunarLeapMonth:     chart.LunarLeapMonth,
		Gender:             gender,
		ZiHourPolicy:       chart.ZiHourPolicy,
		BirthPlace:         chart.BirthPlace,
		Timezone:           chart.Timezone,
		UTCOffsetSeconds:   chart.BirthUTCOffsetSeconds,
		Longitude:          chart.Longitude,
		UseTrueSolarTime:   chart.UseTrueSolarTime,
		TimeUncertain:      chart.TimeUncertain,
		UncertaintySeconds: chart.UncertaintySeconds,
	}
	candidates, err := CalculateBirthCandidates(service, input)
	if err != nil {
		return nil, fmt.Errorf("failed to normalize stored birth input: %w", err)
	}
	selectedID := strings.TrimSpace(chart.SelectedCandidateID)
	if selectedID == "" {
		if candidates.RequiresCandidateSelection {
			return nil, fmt.Errorf("stored birth chart crosses a four-pillar boundary but has no selected candidate ID")
		}
		candidate := candidates.Candidates[0]
		return &StoredBirthResolution{Normalized: candidate.Normalized, Result: candidate.Result, Source: StoredBirthSourceRaw}, nil
	}
	for _, candidate := range candidates.Candidates {
		if candidate.CandidateID == selectedID {
			return &StoredBirthResolution{Normalized: candidate.Normalized, Result: candidate.Result, Source: StoredBirthSourceRaw}, nil
		}
	}
	return nil, fmt.Errorf("stored candidate ID does not match the reconstructed birth-time interval")
}

func decodeStoredNormalizedBirth(raw []byte) (*NormalizedBirth, bool) {
	if len(raw) == 0 || !json.Valid(raw) {
		return nil, false
	}
	var normalized NormalizedBirth
	if err := json.Unmarshal(raw, &normalized); err != nil || normalized.Year <= 0 || normalized.Month <= 0 || normalized.Day <= 0 {
		return nil, false
	}
	return &normalized, true
}

func normalizeStoredGender(value string) (string, bool) {
	value = strings.TrimSpace(value)
	switch {
	case value == "男" || strings.EqualFold(value, "male") || strings.EqualFold(value, "m"):
		return model.GenderMale, true
	case value == "女" || strings.EqualFold(value, "female") || strings.EqualFold(value, "f"):
		return model.GenderFemale, true
	default:
		return "", false
	}
}
