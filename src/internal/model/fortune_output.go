package model

type ShengKeAnalysis struct {
	DayStemRelation   string `json:"day_stem_relation"`
	DayBranchRelation string `json:"day_branch_relation"`
	Summary           string `json:"summary"`
}

// ScoreEvidence records one deterministic rule contribution in the fortune
// scoring pipeline. Impact is the signed point adjustment inside its stage;
// it is explanatory evidence rather than an event probability.
type ScoreEvidence struct {
	Code                 string `json:"code"`
	Stage                string `json:"stage"`
	Category             string `json:"category"`
	Label                string `json:"label"`
	Impact               int    `json:"impact"`
	Description          string `json:"description"`
	Source               string `json:"source"`
	EvidenceBasis        string `json:"evidence_basis"`
	ValidationStatus     string `json:"validation_status"`
	InterpretationStatus string `json:"interpretation_status"`
	IsOutcomeConclusion  bool   `json:"is_outcome_conclusion"`
}

// FortuneScoreBreakdown is the single scoring contract shared by daily,
// weekly, and monthly fortune results.
type FortuneScoreBreakdown struct {
	PipelineVersion      string          `json:"pipeline_version"`
	ScoreKind            string          `json:"score_kind"`
	EvidenceBasis        string          `json:"evidence_basis"`
	ValidationStatus     string          `json:"validation_status"`
	InterpretationStatus string          `json:"interpretation_status"`
	IsOutcomeProbability bool            `json:"is_outcome_probability"`
	BaseScore            int             `json:"base_score"`
	RelationScore        int             `json:"relation_score"`
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
	Key         string           `json:"key"`
	Name        string           `json:"name"`
	Version     string           `json:"version"`
	School      string           `json:"school"`
	Source      string           `json:"source"`
	Sources     []RuleSourceMeta `json:"sources,omitempty"`
	Description string           `json:"description"`
	Count       int              `json:"count,omitempty"`
}

type RuleSourceMeta struct {
	ID               string            `json:"id"`
	Repository       string            `json:"repository"`
	Commit           string            `json:"commit"`
	Files            map[string]string `json:"files"`
	License          string            `json:"license"`
	SourceTier       string            `json:"source_tier"`
	ValidationStatus string            `json:"validation_status"`
}

type BodyStrengthRuleConfig struct {
	Weights              BodyStrengthWeights               `json:"weights"`
	Normalizers          BodyStrengthNormalizers           `json:"normalizers"`
	AdjustmentThresholds BodyStrengthAdjustmentThresholds  `json:"adjustment_thresholds"`
	YueLing              YueLingRuleConfig                 `json:"yue_ling"`
	Root                 BodyStrengthRootRuleConfig        `json:"root"`
	Bonus                BodyStrengthBonusRuleConfig       `json:"bonus"`
	Influence            BodyStrengthInfluenceRuleConfig   `json:"influence"`
	AdjustmentForce      BodyStrengthAdjustmentForceConfig `json:"adjustment_force"`
}

// YueLingRuleConfig publishes the exact 5x12 seasonal-state table used by the
// local body-strength profile. Arrays make the table order explicit and keep
// RuleMeta copies independent without nested slice aliasing.
type YueLingRuleConfig struct {
	RuleID           string               `json:"rule_id"`
	Profile          string               `json:"profile"`
	HashBasis        string               `json:"hash_basis"`
	DayElementOrder  [5]string            `json:"day_element_order"`
	MonthBranchOrder [12]string           `json:"month_branch_order"`
	Scores           [5][12]float64       `json:"scores"`
	ScoreStates      [5]YueLingScoreState `json:"score_states"`
	TableSHA256      string               `json:"table_sha256"`
	EarthMonthPolicy string               `json:"earth_month_policy"`
	ValidationStatus string               `json:"validation_status"`
}

type YueLingScoreState struct {
	State string  `json:"state"`
	Score float64 `json:"score"`
}

type BodyStrengthRootRuleConfig struct {
	RuleID           string                      `json:"rule_id"`
	Profile          string                      `json:"profile"`
	HideStemWeights  BodyStrengthHideStemWeights `json:"hide_stem_weights"`
	TerrainWeights   BodyStrengthTerrainWeights  `json:"terrain_weights"`
	RootMultiplier   float64                     `json:"root_multiplier"`
	TouGanMultiplier float64                     `json:"tou_gan_multiplier"`
	TouGanScope      string                      `json:"tou_gan_scope"`
	ValidationStatus string                      `json:"validation_status"`
}

type BodyStrengthHideStemWeights struct {
	Main     float64 `json:"main"`
	Middle   float64 `json:"middle"`
	Residual float64 `json:"residual"`
}

type BodyStrengthTerrainWeights struct {
	ChangSheng float64 `json:"chang_sheng"`
	MuYu       float64 `json:"mu_yu"`
	GuanDai    float64 `json:"guan_dai"`
	LinGuan    float64 `json:"lin_guan"`
	DiWang     float64 `json:"di_wang"`
	Shuai      float64 `json:"shuai"`
	Bing       float64 `json:"bing"`
	Si         float64 `json:"si"`
	Mu         float64 `json:"mu"`
	Jue        float64 `json:"jue"`
	Tai        float64 `json:"tai"`
	Yang       float64 `json:"yang"`
}

type BodyStrengthBonusRuleConfig struct {
	RuleID             string                  `json:"rule_id"`
	Profile            string                  `json:"profile"`
	HashBasis          string                  `json:"hash_basis"`
	DayStemOrder       [10]string              `json:"day_stem_order"`
	LuBranches         [10]string              `json:"lu_branches"`
	YangRenStemOrder   [5]string               `json:"yang_ren_stem_order"`
	YangRenBranches    [5]string               `json:"yang_ren_branches"`
	Scores             BodyStrengthBonusScores `json:"scores"`
	YinStemBladePolicy string                  `json:"yin_stem_blade_policy"`
	TableSHA256        string                  `json:"table_sha256"`
	ValidationStatus   string                  `json:"validation_status"`
}

type BodyStrengthBonusScores struct {
	DayLu        float64 `json:"day_lu"`
	MonthLu      float64 `json:"month_lu"`
	DayYangRen   float64 `json:"day_yang_ren"`
	MonthYangRen float64 `json:"month_yang_ren"`
}

type BodyStrengthInfluenceRuleConfig struct {
	RuleID                     string  `json:"rule_id"`
	Profile                    string  `json:"profile"`
	VisibleStemScope           string  `json:"visible_stem_scope"`
	SamePolarityPeerWeight     float64 `json:"same_polarity_peer_weight"`
	OppositePolarityPeerWeight float64 `json:"opposite_polarity_peer_weight"`
	OfficerKillerWeight        float64 `json:"officer_killer_weight"`
	OutputWeight               float64 `json:"output_weight"`
	WealthWeight               float64 `json:"wealth_weight"`
	HiddenBranchScope          string  `json:"hidden_branch_scope"`
	HiddenBranchMultiplier     float64 `json:"hidden_branch_multiplier"`
	SameElementRootOwnership   string  `json:"same_element_root_ownership"`
	SealOwnership              string  `json:"seal_ownership"`
	ValidationStatus           string  `json:"validation_status"`
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
	ShiFormula          string  `json:"shi_formula"`
	ShengFormula        string  `json:"sheng_formula"`
}

type BodyStrengthAdjustmentThresholds struct {
	ShiLingSupportForce float64 `json:"shiling_support_force"`
	ShiLingBlendSelf    float64 `json:"shiling_blend_self"`
	ShiLingBlendNeutral float64 `json:"shiling_blend_neutral"`
}

type BodyStrengthAdjustmentForceConfig struct {
	RuleID                 string  `json:"rule_id"`
	Profile                string  `json:"profile"`
	StemForce              float64 `json:"stem_force"`
	HiddenStemMultiplier   float64 `json:"hidden_stem_multiplier"`
	HiddenStemWeightSource string  `json:"hidden_stem_weight_source"`
	ShiLingSupportBasis    string  `json:"shi_ling_support_basis"`
	NeutralTarget          float64 `json:"neutral_target"`
	ValidationStatus       string  `json:"validation_status"`
}

// HiddenStemGod records one hidden stem lookup for the query-day branch.
type HiddenStemGod struct {
	RuleID               string `json:"rule_id"`
	QueryBranch          string `json:"query_branch"`
	ReferenceStem        string `json:"reference_stem"`
	Stem                 string `json:"stem"`
	Type                 string `json:"type"`
	Element              string `json:"element"`
	TenGod               string `json:"ten_god"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
}

type StemRelation struct {
	RuleID               string `json:"rule_id"`
	QueryStem            string `json:"query_stem"`
	TargetPillar         string `json:"target_pillar"`
	TargetStem           string `json:"target_stem"`
	Type                 string `json:"type"`
	Name                 string `json:"name"`
	CombinedElement      string `json:"combined_element,omitempty"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	TransformationStatus string `json:"transformation_status"`
	InterpretationStatus string `json:"interpretation_status"`
}

type BranchRelation struct {
	RuleID               string `json:"rule_id"`
	QueryBranch          string `json:"query_branch"`
	TargetPillar         string `json:"target_pillar"`
	TargetBranch         string `json:"target_branch"`
	Type                 string `json:"type"`
	Name                 string `json:"name"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	TransformationStatus string `json:"transformation_status"`
	InterpretationStatus string `json:"interpretation_status"`
}

// ShenShaActivation describes a shen-sha activated by the query day/layer.
type ShenShaActivation struct {
	Name                 string `json:"name"`
	RuleID               string `json:"rule_id"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
	Activation           string `json:"activation"`
}

type TwelveStageEvidence struct {
	RuleID               string `json:"rule_id"`
	ReferenceStem        string `json:"reference_stem"`
	QueryBranch          string `json:"query_branch"`
	Name                 string `json:"name"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
}

type TenGodEvidence struct {
	RuleID               string `json:"rule_id"`
	ReferenceStem        string `json:"reference_stem"`
	QueryStem            string `json:"query_stem"`
	Name                 string `json:"name"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
}

type SeasonElementEvidence struct {
	RuleID               string `json:"rule_id"`
	ReferenceStem        string `json:"reference_stem"`
	ReferenceElement     string `json:"reference_element"`
	QueryMonthBranch     string `json:"query_month_branch"`
	Season               string `json:"season"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
}

type TraditionalCalendarEvidence struct {
	RuleID               string `json:"rule_id"`
	MonthBranch          string `json:"month_branch"`
	QueryBranch          string `json:"query_branch"`
	Name                 string `json:"name"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
}

type DaYunInfluence struct {
	CurrentPillar        string `json:"current_pillar"`
	Active               bool   `json:"active"`
	Index                int    `json:"index"`
	StartAt              string `json:"start_at,omitempty"`
	EndAtExclusive       string `json:"end_at_exclusive,omitempty"`
	StartAge             int    `json:"start_age"`
	EndAge               int    `json:"end_age"`
	TenGod               string `json:"ten_god"`
	SelectionBasis       string `json:"selection_basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
}

type LiuNianInfluence struct {
	YearPillar           string `json:"year_pillar"`
	TenGod               string `json:"ten_god"`
	SelectionBasis       string `json:"selection_basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
}

type SeasonalStateEvidence struct {
	RuleID               string `json:"rule_id"`
	QueryStem            string `json:"query_stem"`
	QueryElement         string `json:"query_element"`
	QueryMonthBranch     string `json:"query_month_branch"`
	Season               string `json:"season"`
	State                string `json:"state"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
}

// FortuneLayerSet exposes period influences in a stable, machine-readable
// structure for deterministic analysis and future AI interpretation.
type FortuneLayerSet struct {
	RuleVersion         string                 `json:"rule_version"`
	School              string                 `json:"school"`
	DaYun               FortuneLayer           `json:"dayun"`
	LiuNian             FortuneLayer           `json:"liunian"`
	LiuYue              FortuneLayer           `json:"liuyue"`
	XiaoYun             FortuneLayer           `json:"xiaoyun"`
	InterLayerRelations []FortuneLayerRelation `json:"inter_layer_relations"`
}

// FortuneLayer is one luck-period layer: major fortune, year, month, or small fortune.
type FortuneLayer struct {
	RuleID               string                 `json:"rule_id"`
	Key                  string                 `json:"key"`
	Name                 string                 `json:"name"`
	Pillar               string                 `json:"pillar"`
	Gan                  string                 `json:"gan"`
	Zhi                  string                 `json:"zhi"`
	StartAge             int                    `json:"start_age,omitempty"`
	EndAge               int                    `json:"end_age,omitempty"`
	StartAt              string                 `json:"start_at,omitempty"`
	EndAtExclusive       string                 `json:"end_at_exclusive,omitempty"`
	Age                  int                    `json:"age,omitempty"`
	Year                 int                    `json:"year,omitempty"`
	Month                int                    `json:"month,omitempty"`
	TenGod               TenGodEvidence         `json:"ten_god"`
	Relations            []FortuneLayerRelation `json:"relations"`
	ShenShaDetails       []ShenShaActivation    `json:"shen_sha_details"`
	Basis                string                 `json:"basis"`
	Status               string                 `json:"status"`
	InterpretationStatus string                 `json:"interpretation_status"`
}

// FortuneLayerRelation describes how a layer interacts with the natal chart,
// query day, or another period layer.
type FortuneLayerRelation struct {
	RuleID               string `json:"rule_id"`
	Source               string `json:"source"`
	SourceValue          string `json:"source_value"`
	Target               string `json:"target"`
	TargetValue          string `json:"target_value"`
	Type                 string `json:"type"`
	Name                 string `json:"name"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	InterpretationStatus string `json:"interpretation_status"`
}
