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

// FortuneLayerSet exposes period influences in a stable, machine-readable
// structure for deterministic analysis and future AI interpretation.
type FortuneLayerSet struct {
	RuleVersion string       `json:"rule_version"`
	School      string       `json:"school"`
	DaYun       FortuneLayer `json:"dayun"`
	LiuNian     FortuneLayer `json:"liunian"`
	LiuYue      FortuneLayer `json:"liuyue"`
	XiaoYun     FortuneLayer `json:"xiaoyun"`
}

// FortuneLayer is one luck-period layer: major fortune, year, month, or small fortune.
type FortuneLayer struct {
	Key              string                 `json:"key"`
	Name             string                 `json:"name"`
	Pillar           string                 `json:"pillar"`
	Gan              string                 `json:"gan"`
	Zhi              string                 `json:"zhi"`
	StartAge         int                    `json:"start_age,omitempty"`
	EndAge           int                    `json:"end_age,omitempty"`
	Age              int                    `json:"age,omitempty"`
	Year             int                    `json:"year,omitempty"`
	Month            int                    `json:"month,omitempty"`
	TenGod           string                 `json:"ten_god"`
	Favorable        bool                   `json:"favorable"`
	Score            int                    `json:"score"`
	Relations        []FortuneLayerRelation `json:"relations"`
	ActivatedShenSha []string               `json:"activated_shen_sha"`
	ElementChange    map[string]int         `json:"element_change"`
	Description      string                 `json:"description"`
	Evidence         []string               `json:"evidence"`
}

// FortuneLayerRelation describes how a layer interacts with the natal chart
// or the query day.
type FortuneLayerRelation struct {
	Target string `json:"target"`
	Type   string `json:"type"`
	Detail string `json:"detail"`
	Score  int    `json:"score"`
}
