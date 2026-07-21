package model

import "encoding/json"

const (
	CalendarSolar = "SOLAR"
	CalendarLunar = "LUNAR"
	CalendarBazi  = "BAZI"
)

const (
	GenderMale   = "MALE"
	GenderFemale = "FEMALE"
)

// --- Auth DTOs ---

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// --- Chart DTOs ---

type ChartRequest struct {
	BirthYear             int      `json:"birth_year"`
	BirthMonth            int      `json:"birth_month"`
	BirthDay              int      `json:"birth_day"`
	BirthHour             int      `json:"birth_hour"`
	BirthMin              int      `json:"birth_min"`
	BirthSec              int      `json:"birth_sec"`
	CalendarType          string   `json:"calendar_type"`
	LunarLeapMonth        bool     `json:"lunar_leap_month"`
	Gender                string   `json:"gender"`
	ZiHourPolicy          string   `json:"zi_hour_policy"`
	Name                  string   `json:"name"`
	BirthPlace            string   `json:"birth_place"`
	Timezone              string   `json:"timezone"`
	BirthUTCOffsetSeconds *int     `json:"birth_utc_offset_seconds"`
	Longitude             *float64 `json:"longitude"`
	UseTrueSolarTime      bool     `json:"use_true_solar_time"`
	TimeUncertain         bool     `json:"time_uncertain"`
	UncertaintySeconds    int      `json:"uncertainty_seconds"`
	CandidateID           string   `json:"candidate_id"`
}

type Pillar struct {
	Gan string `json:"gan"`
	Zhi string `json:"zhi"`
}

// NaYinInfo is factual evidence for a sixty-cycle na-yin mapping.
type NaYinInfo struct {
	RuleID  string `json:"rule_id"`
	GanZhi  string `json:"gan_zhi"`
	Name    string `json:"name"`
	Element string `json:"element"`
	Basis   string `json:"basis"`
	Status  string `json:"status"`
}

type MingGongDTO struct {
	GanZhi  string `json:"gan_zhi"`
	Gan     string `json:"gan"`
	Zhi     string `json:"zhi"`
	ShenSha string `json:"shen_sha"`
	Nayin   string `json:"nayin"`
}

type ChartResponse struct {
	YearPillar   Pillar               `json:"year_pillar"`
	MonthPillar  Pillar               `json:"month_pillar"`
	DayPillar    Pillar               `json:"day_pillar"`
	HourPillar   Pillar               `json:"hour_pillar"`
	FiveElements map[string]int       `json:"five_elements"`
	NaYin        map[string]NaYinInfo `json:"na_yin"`
	MingGong     MingGongDTO          `json:"ming_gong"`
}

// --- Fortune DTOs ---

type FortuneRequest struct {
	ChartID   uint   `json:"chart_id"`
	QueryDate string `json:"query_date"`
}

type ElementImage struct {
	Element     string `json:"element"`
	ImageURL    string `json:"image_url"`
	Description string `json:"description"`
}

type FortuneResponse struct {
	EngineVersion        string   `json:"engine_version"`
	BaziEngineVersion    string   `json:"bazi_engine_version"`
	BaziResolutionSource string   `json:"bazi_resolution_source"`
	RuleVersion          string   `json:"rule_version,omitempty"`
	School               string   `json:"school,omitempty"`
	RuleMeta             RuleMeta `json:"rule_meta,omitempty"`
	// almanac day fields
	SolarDate            string                `json:"solar_date"`
	LunarDate            string                `json:"lunar_date"`
	DayGanZhi            string                `json:"day_gan_zhi"`
	WeekDay              string                `json:"week_day"`
	ShengXiao            string                `json:"sheng_xiao"`
	JiShen               string                `json:"ji_shen"`
	XiongShen            string                `json:"xiong_shen"`
	ChongSha             string                `json:"chong_sha"`
	TaiShen              string                `json:"tai_shen"`
	WuXing               string                `json:"wu_xing"`
	PengZu               string                `json:"peng_zu"`
	Gua                  string                `json:"gua"`
	JieQi                string                `json:"jie_qi"`
	ElementImages        []ElementImage        `json:"element_images"`
	Score                int                   `json:"score"`
	ScoreBreakdown       FortuneScoreBreakdown `json:"score_breakdown"`
	EvidenceCompleteness int                   `json:"evidence_completeness"`
	SupportingEvidence   []ScoreEvidence       `json:"supporting_evidence"`
	CounterEvidence      []ScoreEvidence       `json:"counter_evidence"`
	ClashZodiac          string                `json:"clash_zodiac"`
	TodayElements        map[string]int        `json:"today_elements"`
	SeasonElement        SeasonElementEvidence `json:"season_element"`
	ShengKeAnalysis      ShengKeAnalysis       `json:"sheng_ke_analysis"`

	// 日课推算结果
	TenGod           TenGodEvidence              `json:"ten_god"`
	TwelveStage      TwelveStageEvidence         `json:"twelve_stage"`
	JianChu          TraditionalCalendarEvidence `json:"jian_chu"`
	HuangDao         TraditionalCalendarEvidence `json:"huang_dao"`
	HiddenStems      []HiddenStemGod             `json:"hidden_stems"`
	StemRelations    []StemRelation              `json:"stem_relations"`
	BranchRelations  []BranchRelation            `json:"branch_relations"`
	ActivatedShenSha []ShenShaActivation         `json:"activated_shen_sha"`
	SeasonalState    SeasonalStateEvidence       `json:"seasonal_state"`
	FortuneLayers    FortuneLayerSet             `json:"fortune_layers"`
}

type WeeklyFortuneRequest struct {
	ChartID   uint   `json:"chart_id"`
	StartDate string `json:"start_date"`
}

type WeeklyFortuneResponse struct {
	DailyFortunes           []FortuneResponse `json:"daily_fortunes"`
	StructuralRelationIndex int               `json:"structural_relation_index"`
	ElementTrend            string            `json:"element_trend"`
	Summary                 FortuneSummary    `json:"summary"`
}

type MonthlyFortuneRequest struct {
	ChartID uint `json:"chart_id"`
	Year    int  `json:"year"`
	Month   int  `json:"month"`
}

type MonthlyFortuneResponse struct {
	DailyFortunes           []FortuneResponse `json:"daily_fortunes"`
	StructuralRelationIndex int               `json:"structural_relation_index"`
	ElementTrend            string            `json:"element_trend"`
	Summary                 FortuneSummary    `json:"summary"`
}

// FortuneSummary aggregates descriptive structural-index statistics across a
// multi-day window. These fields do not express outcome quality or probability.
type FortuneSummary struct {
	HighestIndexDay        string             `json:"highest_index_day"`
	LowestIndexDay         string             `json:"lowest_index_day"`
	HighestIndex           int                `json:"highest_index"`
	LowestIndex            int                `json:"lowest_index"`
	ElementDistribution    map[string]float64 `json:"element_distribution"`
	DominantElement        string             `json:"dominant_element"`
	DominantTenGod         string             `json:"dominant_ten_god"`
	AverageIndex           float64            `json:"average_index"`
	IndexStandardDeviation float64            `json:"index_standard_deviation"`
}

// --- AI Fortune Stub ---

type AIFortuneStubResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// --- Classical Interpretation DTOs ---

type BaziInterpretationRequest struct {
	ChartID uint   `json:"chart_id"`
	Focus   string `json:"focus"`
}

type InterpretationSection struct {
	Title     string `json:"title"`
	Content   string `json:"content"`
	Citations []int  `json:"citations"`
}

type InterpretationCitation struct {
	ID                 int     `json:"id"`
	Book               string  `json:"book"`
	Author             string  `json:"author"`
	Edition            string  `json:"edition"`
	Volume             string  `json:"volume"`
	Chapter            string  `json:"chapter"`
	Page               string  `json:"page"`
	Locator            string  `json:"locator"`
	Path               string  `json:"path"`
	ArtifactPath       string  `json:"artifact_path"`
	ArtifactSHA256     string  `json:"artifact_sha256"`
	DocumentSHA256     string  `json:"document_sha256"`
	Quote              string  `json:"quote"`
	QuoteSHA256        string  `json:"quote_sha256"`
	SourceTier         string  `json:"source_tier"`
	VerificationStatus string  `json:"verification_status"`
	ArtifactKind       string  `json:"artifact_kind"`
	ProvenanceStatus   string  `json:"provenance_status"`
	IndependenceStatus string  `json:"independence_status"`
	CoverageStatus     string  `json:"coverage_status"`
	CatalogSchema      string  `json:"catalog_schema"`
	CatalogVersion     string  `json:"catalog_version"`
	CatalogSHA256      string  `json:"catalog_sha256"`
	ClaimEligible      bool    `json:"claim_eligible"`
	Score              float64 `json:"score"`
}

type BaziInterpretationResponse struct {
	Status    string                   `json:"status"`
	Reason    string                   `json:"reason"`
	ChartID   uint                     `json:"chart_id"`
	Focus     string                   `json:"focus"`
	Summary   string                   `json:"summary"`
	Sections  []InterpretationSection  `json:"sections"`
	Citations []InterpretationCitation `json:"citations"`
}

// --- Feedback DTOs ---

type FeedbackRequest struct {
	ChartID         uint     `json:"chart_id"`
	TargetType      string   `json:"target_type"`
	TargetID        string   `json:"target_id"`
	Rating          string   `json:"rating"`
	Tags            []string `json:"tags"`
	Comment         string   `json:"comment"`
	ConsentResearch bool     `json:"consent_research"`
	ConsentTraining bool     `json:"consent_training"`
}

type FeedbackResponse struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type FeedbackSummaryItem struct {
	TargetType    string `json:"target_type"`
	TargetID      string `json:"target_id"`
	Rating        string `json:"rating"`
	EngineVersion string `json:"engine_version"`
	RuleVersion   string `json:"rule_version"`
	Count         int64  `json:"count"`
}

type FeedbackSummaryResponse struct {
	ChartID          uint                  `json:"chart_id"`
	Total            int64                 `json:"total"`
	ResearchEligible int64                 `json:"research_eligible"`
	Scope            string                `json:"scope"`
	Items            []FeedbackSummaryItem `json:"items"`
}

// --- Generic DTOs ---

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ChartSummaryResponse struct {
	ID                    uint     `json:"id"`
	Name                  string   `json:"name"`
	Gender                string   `json:"gender"`
	ZiHourPolicy          string   `json:"zi_hour_policy"`
	BirthYear             int      `json:"birth_year"`
	BirthMonth            int      `json:"birth_month"`
	BirthDay              int      `json:"birth_day"`
	BirthHour             int      `json:"birth_hour"`
	BirthMin              int      `json:"birth_min"`
	BirthSec              int      `json:"birth_sec"`
	CalendarType          string   `json:"calendar_type"`
	LunarLeapMonth        bool     `json:"lunar_leap_month"`
	BirthPlace            string   `json:"birth_place,omitempty"`
	Timezone              string   `json:"timezone,omitempty"`
	BirthUTCOffsetSeconds *int     `json:"birth_utc_offset_seconds,omitempty"`
	Longitude             *float64 `json:"longitude,omitempty"`
	UseTrueSolarTime      bool     `json:"use_true_solar_time"`
	TimeUncertain         bool     `json:"time_uncertain"`
	UncertaintySeconds    int      `json:"uncertainty_seconds"`
	SelectedCandidateID   string   `json:"selected_candidate_id,omitempty"`
	EngineVersion         string   `json:"engine_version,omitempty"`
	StoredRuleVersion     string   `json:"stored_rule_version,omitempty"`
	CreatedAt             string   `json:"created_at"`
	UpdatedAt             string   `json:"updated_at"`
}

type ChartDetailResponse struct {
	ChartSummaryResponse
	BirthValidation      json.RawMessage `json:"birth_validation"`
	RuleVersion          string          `json:"rule_version,omitempty"`
	School               string          `json:"school,omitempty"`
	RuleMeta             RuleMeta        `json:"rule_meta,omitempty"`
	YearPillar           json.RawMessage `json:"year_pillar"`
	MonthPillar          json.RawMessage `json:"month_pillar"`
	DayPillar            json.RawMessage `json:"day_pillar"`
	HourPillar           json.RawMessage `json:"hour_pillar"`
	FiveElements         json.RawMessage `json:"five_elements"`
	ElementDetail        json.RawMessage `json:"element_detail"`
	BodyStrength         json.RawMessage `json:"body_strength"`
	TenGods              json.RawMessage `json:"ten_gods"`
	NaYin                json.RawMessage `json:"na_yin"`
	HiddenStems          json.RawMessage `json:"hidden_stems"`
	DaYunStart           json.RawMessage `json:"da_yun_start"`
	DaYun                json.RawMessage `json:"da_yun"`
	GanZhiAnalysis       json.RawMessage `json:"gan_zhi_analysis"`
	PatternAnalysis      json.RawMessage `json:"pattern_analysis"`
	MingGong             json.RawMessage `json:"ming_gong"`
	PillarDetails        json.RawMessage `json:"pillar_details"`
	Tiaohou              json.RawMessage `json:"tiaohou"`
	GlobalShenSha        json.RawMessage `json:"global_shen_sha"`
	GlobalShenShaDetails json.RawMessage `json:"global_shen_sha_details"`
	DayShenSha           json.RawMessage `json:"day_shen_sha"`
	DayShenShaDetails    json.RawMessage `json:"day_shen_sha_details"`
	MonthSeason          json.RawMessage `json:"month_season"`
	ShenShaByPillar      json.RawMessage `json:"shen_sha_by_pillar"`
	TenGodProportion     json.RawMessage `json:"ten_god_proportion"`
	TenGodAnalysis       json.RawMessage `json:"ten_god_analysis"`
	MissingElements      json.RawMessage `json:"missing_elements"`
	ZiWeiResult          json.RawMessage `json:"ziwei_result"`
	ZiWeiComputed        bool            `json:"ziwei_computed"`
}

type ChartListResponse struct {
	Charts   []ChartSummaryResponse `json:"charts"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type HistoryResponse struct {
	ID        uint   `json:"id"`
	ChartID   uint   `json:"chart_id"`
	ChartName string `json:"chart_name"`
	QueryDate string `json:"query_date"`
	DayGanZhi string `json:"day_gan_zhi"`
	Summary   string `json:"summary"`
	CreatedAt string `json:"created_at"`
}
