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
	BirthYear    int    `json:"birth_year"`
	BirthMonth   int    `json:"birth_month"`
	BirthDay     int    `json:"birth_day"`
	BirthHour    int    `json:"birth_hour"`
	BirthMin     int    `json:"birth_min"`
	CalendarType string `json:"calendar_type"`
	Gender       string `json:"gender"`
	Name         string `json:"name"`
}

type Pillar struct {
	Gan string `json:"gan"`
	Zhi string `json:"zhi"`
}

// NaYinInfo is the JSON-serializable na-yin detail.
type NaYinInfo struct {
	Name        string   `json:"name"`
	Element     string   `json:"element"`
	ImageDesc   string   `json:"image_desc"`
	Personality string   `json:"personality"`
	EnergyStage string   `json:"energy_stage"`
	ModernExt   string   `json:"modern_ext"`
	Judgments   []string `json:"judgments"`
}

type MingGongDTO struct {
	GanZhi      string `json:"gan_zhi"`
	Gan         string `json:"gan"`
	Zhi         string `json:"zhi"`
	ShenSha     string `json:"shen_sha"`
	ShenShaDesc string `json:"shen_sha_desc"`
	ZhiDetail   string `json:"zhi_detail"`
	Nayin       string `json:"nayin"`
}

type ChartResponse struct {
	YearPillar   Pillar               `json:"year_pillar"`
	MonthPillar  Pillar               `json:"month_pillar"`
	DayPillar    Pillar               `json:"day_pillar"`
	HourPillar   Pillar               `json:"hour_pillar"`
	FiveElements map[string]int       `json:"five_elements"`
	NaYin        map[string]NaYinInfo `json:"na_yin"`
	MingGong     MingGongDTO          `json:"ming_gong"`
	RiZhuDesc    string               `json:"ri_zhu_desc"`
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

type FortuneGuideItem struct {
	Label   string `json:"label"`
	Value   string `json:"value"`
	Element string `json:"element,omitempty"`
	Reason  string `json:"reason"`
}

type FortuneGuide struct {
	PrecisionLevel     string             `json:"precision_level"`
	Confidence         int                `json:"confidence"`
	PrimaryElement     string             `json:"primary_element"`
	SecondaryElement   string             `json:"secondary_element"`
	AvoidElement       string             `json:"avoid_element"`
	LuckyColors        []FortuneGuideItem `json:"lucky_colors"`
	LuckyNumbers       []FortuneGuideItem `json:"lucky_numbers"`
	FaceDirection      FortuneGuideItem   `json:"face_direction"`
	WealthDirection    FortuneGuideItem   `json:"wealth_direction"`
	AvoidDirection     FortuneGuideItem   `json:"avoid_direction"`
	RecommendedActions []FortuneGuideItem `json:"recommended_actions"`
	Cautions           []FortuneGuideItem `json:"cautions"`
	BestHours          []FortuneGuideItem `json:"best_hours"`
	Analysis           string             `json:"analysis"`
	Strategy           string             `json:"strategy"`
}

type FortuneResponse struct {
	// almanac day fields
	SolarDate           string         `json:"solar_date"`
	LunarDate           string         `json:"lunar_date"`
	DayGanZhi           string         `json:"day_gan_zhi"`
	WeekDay             string         `json:"week_day"`
	ShengXiao           string         `json:"sheng_xiao"`
	YiJi                string         `json:"yi_ji"`
	JiShen              string         `json:"ji_shen"`
	XiongShen           string         `json:"xiong_shen"`
	ChongSha            string         `json:"chong_sha"`
	TaiShen             string         `json:"tai_shen"`
	WuXing              string         `json:"wu_xing"`
	PengZu              string         `json:"peng_zu"`
	Gua                 string         `json:"gua"`
	JieQi               string         `json:"jie_qi"`
	ElementImages       []ElementImage `json:"element_images"`
	Score               int            `json:"score"`
	LuckyColor          string         `json:"lucky_color"`
	LuckyNumber         int            `json:"lucky_number"`
	WealthDir           string         `json:"wealth_direction"`
	ClashZodiac         string         `json:"clash_zodiac"`
	AuspiciousHours     []string       `json:"auspicious_hours"`
	Guide               *FortuneGuide  `json:"guide,omitempty"`
	Analysis            interface{}    `json:"analysis"`
	YiItems             []string       `json:"yi"`
	JiItems             []string       `json:"ji"`
	TodayElements       map[string]int `json:"today_elements"`
	TiaoHou             string         `json:"tiao_hou"`
	SeasonElementAdvice string         `json:"season_element_advice"`
	FlowImpact          string         `json:"flow_impact"`
	ShengKeAnalysis     interface{}    `json:"sheng_ke_analysis"`

	// 日课推算结果
	TodayTenGod      string      `json:"today_ten_god"`
	TenGodFavorable  bool        `json:"ten_god_favorable"`
	TenGodDesc       string      `json:"ten_god_desc"`
	TwelveStage      string      `json:"twelve_stage"`
	StageFavorable   bool        `json:"stage_favorable"`
	StageDesc        string      `json:"stage_desc"`
	StageFlexible    string      `json:"stage_flexible"`
	HiddenStems      interface{} `json:"hidden_stems"`
	StemRelations    interface{} `json:"stem_relations"`
	BranchRelations  interface{} `json:"branch_relations"`
	ActivatedShenSha interface{} `json:"activated_shen_sha"`
	DaYunInfluence   interface{} `json:"dayun_influence"`
	LiuNianInfluence interface{} `json:"liunian_influence"`
	AdvanceRetreat   interface{} `json:"advance_retreat"`
	YongShenImpact   interface{} `json:"yongshen_impact"`
	OverallVerdict   string      `json:"overall_verdict"`
	FavorScore       int         `json:"favor_score"`

	// 格局信息
	PatternName        string   `json:"pattern_name"`
	PatternType        string   `json:"pattern_type"`
	PatternFavorable   []string `json:"pattern_favorable"`
	PatternUnfavorable []string `json:"pattern_unfavorable"`
}

type WeeklyFortuneRequest struct {
	ChartID   uint   `json:"chart_id"`
	StartDate string `json:"start_date"`
}

type WeeklyFortuneResponse struct {
	DailyFortunes []FortuneResponse `json:"daily_fortunes"`
	WeeklyScore   int               `json:"weekly_score"`
	ElementTrend  string            `json:"element_trend"`
	Summary       FortuneSummary    `json:"summary"`
}

type MonthlyFortuneRequest struct {
	ChartID uint `json:"chart_id"`
	Year    int  `json:"year"`
	Month   int  `json:"month"`
}

type MonthlyFortuneResponse struct {
	DailyFortunes []FortuneResponse `json:"daily_fortunes"`
	MonthlyScore  int               `json:"monthly_score"`
	ElementTrend  string            `json:"element_trend"`
	Summary       FortuneSummary    `json:"summary"`
}

// FortuneSummary aggregates statistics across a multi-day window (week / month).
// Scores are on the 0-100 scale produced by FortuneEngine.
type FortuneSummary struct {
	BestDay             string             `json:"best_day"`
	WorstDay            string             `json:"worst_day"`
	BestScore           int                `json:"best_score"`
	WorstScore          int                `json:"worst_score"`
	PeakDays            []string           `json:"peak_days"`
	LowDays             []string           `json:"low_days"`
	ElementDistribution map[string]float64 `json:"element_distribution"`
	DominantElement     string             `json:"dominant_element"`
	DominantTenGod      string             `json:"dominant_ten_god"`
	GoodStreak          int                `json:"good_streak"`
	BadStreak           int                `json:"bad_streak"`
	AverageScore        float64            `json:"average_score"`
	Volatility          float64            `json:"volatility"`
	KeyAdvice           string             `json:"key_advice"`
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
	ID      int     `json:"id"`
	Book    string  `json:"book"`
	Chapter string  `json:"chapter"`
	Path    string  `json:"path"`
	Quote   string  `json:"quote"`
	Score   float64 `json:"score"`
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
	EventYear       int      `json:"event_year"`
	EventCategory   string   `json:"event_category"`
	ConsentResearch bool     `json:"consent_research"`
	ConsentTraining bool     `json:"consent_training"`
}

type FeedbackResponse struct {
	ID     uint   `json:"id"`
	Status string `json:"status"`
}

type FeedbackSummaryItem struct {
	Rating string `json:"rating"`
	Count  int64  `json:"count"`
}

type FeedbackSummaryResponse struct {
	ChartID uint                  `json:"chart_id"`
	Total   int64                 `json:"total"`
	Items   []FeedbackSummaryItem `json:"items"`
}

// --- Generic DTOs ---

type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ChartSummaryResponse struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Gender       string `json:"gender"`
	BirthYear    int    `json:"birth_year"`
	BirthMonth   int    `json:"birth_month"`
	BirthDay     int    `json:"birth_day"`
	BirthHour    int    `json:"birth_hour"`
	BirthMin     int    `json:"birth_min"`
	CalendarType string `json:"calendar_type"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type ChartDetailResponse struct {
	ChartSummaryResponse
	YearPillar    json.RawMessage `json:"year_pillar"`
	MonthPillar   json.RawMessage `json:"month_pillar"`
	DayPillar     json.RawMessage `json:"day_pillar"`
	HourPillar    json.RawMessage `json:"hour_pillar"`
	FiveElements  json.RawMessage `json:"five_elements"`
	ElementDetail json.RawMessage `json:"element_detail"`
	BodyStrength  json.RawMessage `json:"body_strength"`
	TenGods       json.RawMessage `json:"ten_gods"`
	NaYin         json.RawMessage `json:"na_yin"`
	DaYunStart    json.RawMessage `json:"da_yun_start"`
	DaYun         json.RawMessage `json:"da_yun"`
	ZiWeiResult   json.RawMessage `json:"ziwei_result"`
	ZiWeiComputed bool            `json:"ziwei_computed"`
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
