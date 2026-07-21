package bazi

import (
	"fmt"
	"strings"

	"github.com/6tail/tyme4go/tyme"
)

const (
	ZiHourLateZiNextDay = "late_zi_next_day"
	ZiHourLateZiSameDay = "late_zi_same_day"
	DefaultZiHourPolicy = ZiHourLateZiNextDay
)

// NormalizeZiHourPolicy resolves an empty value to the versioned default and
// rejects unknown school conventions instead of silently changing day pillars.
func NormalizeZiHourPolicy(value string) (string, error) {
	policy := strings.ToLower(strings.TrimSpace(value))
	if policy == "" {
		policy = DefaultZiHourPolicy
	}
	switch policy {
	case ZiHourLateZiNextDay, ZiHourLateZiSameDay:
		return policy, nil
	default:
		return "", fmt.Errorf("zi_hour_policy must be %s or %s", ZiHourLateZiNextDay, ZiHourLateZiSameDay)
	}
}

func eightCharWithZiHourPolicy(solarTime *tyme.SolarTime, policy string) (tyme.EightChar, error) {
	policy, err := NormalizeZiHourPolicy(policy)
	if err != nil {
		return tyme.EightChar{}, err
	}
	lunarHour := solarTime.GetLunarHour()
	if policy == ZiHourLateZiSameDay {
		return (tyme.LunarSect2EightCharProvider{}).GetEightChar(lunarHour), nil
	}
	return (tyme.DefaultEightCharProvider{}).GetEightChar(lunarHour), nil
}
