package bazi

// MonthSeasonEvidence is a factual classification derived only from the
// month-pillar branch. It does not contain strength, fortune, or life outcomes.
type MonthSeasonEvidence struct {
	RuleID           string `json:"rule_id"`
	MonthBranch      string `json:"month_branch"`
	TraditionalMonth int    `json:"traditional_month"`
	Season           string `json:"season"`
	Basis            string `json:"basis"`
	Status           string `json:"status"`
}

func observeMonthSeason(monthBranch string) MonthSeasonEvidence {
	traditionalMonth := map[string]int{
		"寅": 1, "卯": 2, "辰": 3,
		"巳": 4, "午": 5, "未": 6,
		"申": 7, "酉": 8, "戌": 9,
		"亥": 10, "子": 11, "丑": 12,
	}[monthBranch]
	season := map[string]string{
		"寅": "春", "卯": "春", "辰": "春",
		"巳": "夏", "午": "夏", "未": "夏",
		"申": "秋", "酉": "秋", "戌": "秋",
		"亥": "冬", "子": "冬", "丑": "冬",
	}[monthBranch]
	status := "observed"
	if traditionalMonth == 0 || season == "" {
		status = "unavailable"
	}
	return MonthSeasonEvidence{
		RuleID:           "month-season.branch-order-v1",
		MonthBranch:      monthBranch,
		TraditionalMonth: traditionalMonth,
		Season:           season,
		Basis:            "month_pillar_branch",
		Status:           status,
	}
}

// ValidMonthSeasonEvidence verifies persisted evidence before a saved snapshot
// is allowed to drive API responses.
func ValidMonthSeasonEvidence(evidence MonthSeasonEvidence, monthBranch string) bool {
	want := observeMonthSeason(monthBranch)
	return want.Status == "observed" && evidence == want
}
