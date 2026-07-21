package bazi

import "math"

const secondsPerDay = 24 * 60 * 60

// MonthCommandDepthCandidate records which stem commands the current part of
// an earth month under one classical day-allocation profile. Profiles remain
// parallel evidence because their segment lengths and middle stems differ.
type MonthCommandDepthCandidate struct {
	RuleID               string  `json:"rule_id"`
	ProfileID            string  `json:"profile_id"`
	Source               string  `json:"source"`
	Sequence             string  `json:"sequence"`
	CommandingStem       string  `json:"commanding_stem"`
	Segment              string  `json:"segment"`
	SegmentStartDay      int     `json:"segment_start_day"`
	SegmentEndDay        int     `json:"segment_end_day,omitempty"`
	PositionDay          float64 `json:"position_day"`
	Basis                string  `json:"basis"`
	Status               string  `json:"status"`
	InterpretationStatus string  `json:"interpretation_status"`
}

type monthCommandSegment struct {
	stem string
	days int
	name string
}

type monthCommandProfile struct {
	id       string
	source   string
	sequence map[string]string
	segments map[string][3]monthCommandSegment
}

func earthMonthCommandProfiles() []monthCommandProfile {
	return []monthCommandProfile{
		{
			id:     "sanming_tonghui_renyuan_sishi_7_5_remainder",
			source: "《三命通会·论人元司事》",
			sequence: map[string]string{
				"辰": "乙7日、壬5日、余日戊", "未": "丁7日、甲5日、余日己",
				"戌": "辛7日、丙5日、余日戊", "丑": "癸7日、庚5日、余日己",
			},
			segments: map[string][3]monthCommandSegment{
				"辰": {{"乙", 7, "余气段"}, {"壬", 5, "墓库段"}, {"戊", 0, "本气段"}},
				"未": {{"丁", 7, "余气段"}, {"甲", 5, "墓库段"}, {"己", 0, "本气段"}},
				"戌": {{"辛", 7, "余气段"}, {"丙", 5, "墓库段"}, {"戊", 0, "本气段"}},
				"丑": {{"癸", 7, "余气段"}, {"庚", 5, "墓库段"}, {"己", 0, "本气段"}},
			},
		},
		{
			id:     "yuanhai_ziping_jieqige_9_3_remainder",
			source: "《渊海子平·又论节气歌》",
			sequence: map[string]string{
				"辰": "乙9日、癸3日、余日戊", "未": "丁9日、乙3日、余日己",
				"戌": "辛9日、丁3日、余日戊", "丑": "癸9日、辛3日、余日己",
			},
			segments: map[string][3]monthCommandSegment{
				"辰": {{"乙", 9, "余气段"}, {"癸", 3, "中气段"}, {"戊", 0, "本气段"}},
				"未": {{"丁", 9, "余气段"}, {"乙", 3, "中气段"}, {"己", 0, "本气段"}},
				"戌": {{"辛", 9, "余气段"}, {"丁", 3, "中气段"}, {"戊", 0, "本气段"}},
				"丑": {{"癸", 9, "余气段"}, {"辛", 3, "中气段"}, {"己", 0, "本气段"}},
			},
		},
	}
}

func observeEarthMonthCommandCandidates(monthBranch string, elapsedSeconds int) []MonthCommandDepthCandidate {
	result := make([]MonthCommandDepthCandidate, 0, 2)
	if elapsedSeconds < 0 || !inStrings(monthBranch, "丑", "辰", "未", "戌") {
		return result
	}

	elapsedDays := float64(elapsedSeconds) / secondsPerDay
	positionDay := math.Round((elapsedDays+1)*100) / 100
	for _, profile := range earthMonthCommandProfiles() {
		segments, ok := profile.segments[monthBranch]
		if !ok {
			continue
		}
		startDay := 1
		for _, segment := range segments {
			endDay := 0
			if segment.days > 0 {
				endDay = startDay + segment.days - 1
			}
			if segment.days == 0 || elapsedDays < float64(endDay) {
				result = append(result, MonthCommandDepthCandidate{
					RuleID:               "bazi.month-command.day-profile." + profile.id,
					ProfileID:            profile.id,
					Source:               profile.source,
					Sequence:             profile.sequence[monthBranch],
					CommandingStem:       segment.stem,
					Segment:              segment.name,
					SegmentStartDay:      startDay,
					SegmentEndDay:        endDay,
					PositionDay:          positionDay,
					Basis:                "elapsed_since_month_jie",
					Status:               "observed",
					InterpretationStatus: "not_adjudicated",
				})
				break
			}
			startDay = endDay + 1
		}
	}
	return result
}
