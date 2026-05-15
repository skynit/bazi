package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ParsedBirth holds the parsed birth date and gender.
type ParsedBirth struct {
	Year, Month, Day, Hour, Minute int
	Gender                         string // "MALE" or "FEMALE"
}

// ParseRequest carries the raw input for parsing.
type ParseRequest struct {
	Input        string `json:"input"`         // e.g. "1990-01-15 08:00", "农历 ...", "己巳 丁丑 庚辰 庚辰"
	CalendarType string `json:"calendar_type"` // "SOLAR","LUNAR","BAZI" or "" for auto-detect
	Gender       string `json:"gender"`        // "MALE","FEMALE"
}

// InputParser parses multi-format birth input.
type InputParser struct{}

// Chinese gan-zhi character sets.
const (
	stemsZh   = "甲乙丙丁戊己庚辛壬癸"
	branchesZh = "子丑寅卯辰巳午未申酉戌亥"
)

var (
	ganZhiChars = stemsZh + branchesZh

	// Solar: "YYYY-MM-DD HH:MM"
	solarRe1 = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})\s+(\d{2}):(\d{2})$`)
	// Solar: "YYYYMMDD HHMM"
	solarRe2 = regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})\s+(\d{2})(\d{2})$`)

	// Bazi: 4 pillars of stem+branch pairs, e.g. "己巳 丁丑 庚辰 庚辰"
	baziRe = regexp.MustCompile(`^([` + ganZhiChars + `]{2})\s+([` + ganZhiChars + `]{2})\s+([` + ganZhiChars + `]{2})\s+([` + ganZhiChars + `]{2})$`)

	// Lunar: contains any CJK character
	lunarRe = regexp.MustCompile(`\p{Han}`)

	// Lunar numeric extraction: digits before 年月日
	lunarYearRe  = regexp.MustCompile(`(\d+)\s*年`)
	lunarMonthRe = regexp.MustCompile(`(\d+)\s*月`)
	lunarDayRe   = regexp.MustCompile(`(\d+)\s*日`)
	lunarHourRe  = regexp.MustCompile(`(\d+)\s*时`)
)

// Parse auto-detects or uses the given calendar type to parse the input.
func (p *InputParser) Parse(req ParseRequest) (*ParsedBirth, error) {
	if req.Gender != "" {
		g := strings.ToUpper(req.Gender)
		if g != "MALE" && g != "FEMALE" {
			return nil, fmt.Errorf("invalid gender %q: must be MALE or FEMALE", req.Gender)
		}
	}

	ct := strings.ToUpper(strings.TrimSpace(req.CalendarType))
	if ct == "" {
		var err error
		ct, err = detectCalendarType(req.Input)
		if err != nil {
			return nil, err
		}
	}

	switch ct {
	case "SOLAR":
		return parseSolar(req)
	case "LUNAR":
		return parseLunar(req)
	case "BAZI":
		return parseBazi(req)
	default:
		return nil, fmt.Errorf("unsupported calendar type: %s", req.CalendarType)
	}
}

// detectCalendarType guesses the format from the input string.
func detectCalendarType(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty input")
	}

	if solarRe1.MatchString(input) || solarRe2.MatchString(input) {
		return "SOLAR", nil
	}
	if baziRe.MatchString(input) {
		return "BAZI", nil
	}
	if lunarRe.MatchString(input) {
		return "LUNAR", nil
	}

	return "", fmt.Errorf("unable to detect calendar type from input: %s", input)
}

// parseSolar handles "YYYY-MM-DD HH:MM" and "YYYYMMDD HHMM".
func parseSolar(req ParseRequest) (*ParsedBirth, error) {
	matches := solarRe1.FindStringSubmatch(req.Input)
	if matches == nil {
		matches = solarRe2.FindStringSubmatch(req.Input)
		if matches == nil {
			return nil, fmt.Errorf("invalid solar date format: %s", req.Input)
		}
	}

	year, _ := strconv.Atoi(matches[1])
	month, _ := strconv.Atoi(matches[2])
	day, _ := strconv.Atoi(matches[3])
	hour, _ := strconv.Atoi(matches[4])
	minute, _ := strconv.Atoi(matches[5])

	if month < 1 || month > 12 || day < 1 || day > 31 || hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return nil, fmt.Errorf("solar date out of range: %s", req.Input)
	}

	return &ParsedBirth{
		Year:   year,
		Month:  month,
		Day:    day,
		Hour:   hour,
		Minute: minute,
		Gender: strings.ToUpper(req.Gender),
	}, nil
}

// parseLunar handles Chinese-character lunar date input.
func parseLunar(req ParseRequest) (*ParsedBirth, error) {
	input := strings.TrimSpace(req.Input)

	// Extract lunar month name lookup
	lunarMonths := map[string]int{
		"正月": 1, "一月": 1,
		"二月": 2, "三月": 3, "四月": 4, "五月": 5, "六月": 6,
		"七月": 7, "八月": 8, "九月": 9, "十月": 10,
		"十一月": 11, "十二月": 12, "冬月": 11, "腊月": 12,
	}

	year := 0
	month := 0
	day := 0
	hour := 0
	minute := 0

	// Try to extract year
	if m := lunarYearRe.FindStringSubmatch(input); m != nil {
		year, _ = strconv.Atoi(m[1])
	}

	// Try to extract month (numeric or Chinese name)
	monthFound := false
	if m := lunarMonthRe.FindStringSubmatch(input); m != nil {
		month, _ = strconv.Atoi(m[1])
		monthFound = true
	} else {
		for name, val := range lunarMonths {
			if strings.Contains(input, name) {
				month = val
				monthFound = true
				break
			}
		}
	}

	// Try to extract day
	if m := lunarDayRe.FindStringSubmatch(input); m != nil {
		day, _ = strconv.Atoi(m[1])
	}

	// Try to extract hour
	if m := lunarHourRe.FindStringSubmatch(input); m != nil {
		hour, _ = strconv.Atoi(m[1])
	}

	if year == 0 || !monthFound || day == 0 {
		return nil, fmt.Errorf("unable to parse lunar date from: %s", req.Input)
	}

	return &ParsedBirth{
		Year:   year,
		Month:  month,
		Day:    day,
		Hour:   hour,
		Minute: minute,
		Gender: strings.ToUpper(req.Gender),
	}, nil
}

// parseBazi handles 4-pillar gan-zhi input (e.g. "己巳 丁丑 庚辰 庚辰").
// Bazi format does not contain calendar date info → returns zero date with gender from request.
func parseBazi(req ParseRequest) (*ParsedBirth, error) {
	if !baziRe.MatchString(req.Input) {
		return nil, fmt.Errorf("invalid bazi format: %s", req.Input)
	}

	return &ParsedBirth{
		Year:   0,
		Month:  0,
		Day:    0,
		Hour:   0,
		Minute: 0,
		Gender: strings.ToUpper(req.Gender),
	}, nil
}
