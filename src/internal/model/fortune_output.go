package model

type YiJiItem struct {
	Activity string `json:"activity"`
	Reason   string `json:"reason"`
}

type LuckyInfo struct {
	Color           string   `json:"color"`
	Number          string   `json:"number"`
	Direction       string   `json:"direction"`
	ZodiacClash     string   `json:"zodiac_clash"`
	AuspiciousHours []string `json:"auspicious_hours"`
}

type ShengKeAnalysis struct {
	DayStemRelation   string `json:"day_stem_relation"`
	DayBranchRelation string `json:"day_branch_relation"`
	Summary           string `json:"summary"`
}
