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

// ScoreEvidence records one deterministic rule contribution in the fortune
// scoring pipeline. Impact is the signed point adjustment inside its stage;
// it is explanatory evidence rather than an event probability.
type ScoreEvidence struct {
	Code        string `json:"code"`
	Stage       string `json:"stage"`
	Category    string `json:"category"`
	Label       string `json:"label"`
	Impact      int    `json:"impact"`
	Description string `json:"description"`
	Source      string `json:"source"`
}

// FortuneScoreBreakdown is the single scoring contract shared by daily,
// weekly, and monthly fortune results.
type FortuneScoreBreakdown struct {
	PipelineVersion      string          `json:"pipeline_version"`
	BaseScore            int             `json:"base_score"`
	RelationScore        int             `json:"relation_score"`
	DetailScore          int             `json:"detail_score"`
	FinalScore           int             `json:"final_score"`
	EvidenceCompleteness int             `json:"evidence_completeness"`
	SupportingEvidence   []ScoreEvidence `json:"supporting_evidence"`
	CounterEvidence      []ScoreEvidence `json:"counter_evidence"`
}

// RuleMeta describes the deterministic rule tables used for a calculation.
type RuleMeta struct {
	RuleVersion  string                 `json:"rule_version"`
	School       string                 `json:"school"`
	Tables       []RuleTableMeta        `json:"tables"`
	BodyStrength BodyStrengthRuleConfig `json:"body_strength,omitempty"`
}

type RuleTableMeta struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	School      string `json:"school"`
	Source      string `json:"source"`
	Description string `json:"description"`
	Count       int    `json:"count,omitempty"`
}

type BodyStrengthRuleConfig struct {
	Weights              BodyStrengthWeights              `json:"weights"`
	Normalizers          BodyStrengthNormalizers          `json:"normalizers"`
	AdjustmentThresholds BodyStrengthAdjustmentThresholds `json:"adjustment_thresholds"`
}

type BodyStrengthWeights struct {
	Ling  float64 `json:"ling"`
	Di    float64 `json:"di"`
	Shi   float64 `json:"shi"`
	Sheng float64 `json:"sheng"`
	Bonus float64 `json:"bonus"`
}

type BodyStrengthNormalizers struct {
	Ling                float64 `json:"ling"`
	Di                  float64 `json:"di"`
	ShiSigmoidDivisor   float64 `json:"shi_sigmoid_divisor"`
	ShengSigmoidDivisor float64 `json:"sheng_sigmoid_divisor"`
}

type BodyStrengthAdjustmentThresholds struct {
	DeLingRestrictRatio float64 `json:"deling_restrict_ratio"`
	DeLingMultiplier    float64 `json:"deling_multiplier"`
	ShiLingSupportForce float64 `json:"shiling_support_force"`
	ShiLingBlendSelf    float64 `json:"shiling_blend_self"`
	ShiLingBlendNeutral float64 `json:"shiling_blend_neutral"`
}

// FortuneAnalysis is the detailed daily fortune reading returned by the API.
type FortuneAnalysis struct {
	SolarDate   string          `json:"solar_date"`
	LunarDate   string          `json:"lunar_date"`
	UserBazi    string          `json:"user_bazi"`
	TodayGanZhi string          `json:"today_gan_zhi"`
	TodayElem   string          `json:"today_element"`
	Overall     OverallAnalysis `json:"overall"`
	Categories  []CategoryScore `json:"categories"`
	Hourly      []HourlyFortune `json:"hourly"`
	LuckyGuide  LuckyGuide      `json:"lucky_guide"`
}

type OverallAnalysis struct {
	Score       int    `json:"score"`
	BaseScore   int    `json:"base_score"`
	DetailScore int    `json:"detail_score"`
	Stars       string `json:"stars"`
	Level       string `json:"level"`
	Summary     string `json:"summary"`
	KeyTip      string `json:"key_tip"`
}

type CategoryScore struct {
	Name     string   `json:"name"`
	Score    int      `json:"score"`
	Weight   int      `json:"weight"`
	Stars    string   `json:"stars"`
	Level    string   `json:"level"`
	Trend    string   `json:"trend"`
	Keywords []string `json:"keywords"`
	Analysis string   `json:"analysis"`
	Advice   string   `json:"advice"`
}

type HourlyFortune struct {
	Shichen    string `json:"shichen"`
	TimeRange  string `json:"time_range"`
	Mood       string `json:"mood"`
	Suggestion string `json:"suggestion"`
}

type LuckyGuide struct {
	Colors           string   `json:"colors"`
	Numbers          string   `json:"numbers"`
	Actions          string   `json:"actions"`
	AvoidDir         string   `json:"avoid_dir"`
	FaceDir          string   `json:"face_dir"`
	Outfit           string   `json:"outfit"`
	FavorableElems   []string `json:"favorable_elems"`
	UnfavorableElems []string `json:"unfavorable_elems"`
}

// HiddenStemGod is a query-day hidden-stem ten-god evaluation.
type HiddenStemGod struct {
	Stem      string `json:"stem"`
	Type      string `json:"type"`
	Element   string `json:"element"`
	TenGod    string `json:"ten_god"`
	Favorable bool   `json:"favorable"`
}

type StemRelation struct {
	Type        string `json:"type"`
	Target      string `json:"target"`
	Detail      string `json:"detail"`
	IsFavorable bool   `json:"is_favorable"`
	Note        string `json:"note"`
}

type BranchRelation struct {
	Type        string `json:"type"`
	Target      string `json:"target"`
	Detail      string `json:"detail"`
	IsFavorable bool   `json:"is_favorable"`
}

// ShenShaActivation describes a shen-sha activated by the query day/layer.
type ShenShaActivation struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Category    string `json:"category,omitempty"`
	Polarity    string `json:"polarity,omitempty"`
	Priority    int    `json:"priority,omitempty"`
	Description string `json:"description"`
	Activation  string `json:"activation"`
}

type DaYunInfluence struct {
	CurrentPillar string `json:"current_pillar"`
	StartAge      int    `json:"start_age"`
	EndAge        int    `json:"end_age"`
	TenGod        string `json:"ten_god"`
	Favorable     bool   `json:"favorable"`
	Relation      string `json:"relation"`
	Score         int    `json:"score"`
	Description   string `json:"description"`
}

type LiuNianInfluence struct {
	YearPillar     string `json:"year_pillar"`
	TenGod         string `json:"ten_god"`
	Favorable      bool   `json:"favorable"`
	Relation       string `json:"relation"`
	TaiSuiRelation string `json:"tai_sui_relation"`
	Score          int    `json:"score"`
	Description    string `json:"description"`
}

type AdvanceRetreat struct {
	Phase       string `json:"phase"`
	PhaseDesc   string `json:"phase_desc"`
	Element     string `json:"element"`
	Score       int    `json:"score"`
	Description string `json:"description"`
}

type YongShenImpact struct {
	TiaoHouElement  string   `json:"tiao_hou_element"`
	TiaoHouHit      bool     `json:"tiao_hou_hit"`
	TongGuanElement string   `json:"tong_guan_element"`
	TongGuanHit     bool     `json:"tong_guan_hit"`
	FuYiElements    []string `json:"fu_yi_elements"`
	FuYiHit         bool     `json:"fu_yi_hit"`
	Score           int      `json:"score"`
	Description     string   `json:"description"`
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
	ShenShaDetails   []ShenShaActivation    `json:"shen_sha_details,omitempty"`
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
