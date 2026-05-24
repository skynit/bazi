package model

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

type ChartResponse struct {
	YearPillar    Pillar                   `json:"year_pillar"`
	MonthPillar   Pillar                   `json:"month_pillar"`
	DayPillar     Pillar                   `json:"day_pillar"`
	HourPillar    Pillar                   `json:"hour_pillar"`
	FiveElements  map[string]int            `json:"five_elements"`
	NaYin         map[string]NaYinInfo      `json:"na_yin"`
	MingGong      string                   `json:"ming_gong"`
	RiZhuDesc     string                   `json:"ri_zhu_desc"`
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
	// almanac day fields
	SolarDate       string         `json:"solar_date"`
	LunarDate       string         `json:"lunar_date"`
	DayGanZhi       string         `json:"day_gan_zhi"`
	WeekDay         string         `json:"week_day"`
	ShengXiao       string         `json:"sheng_xiao"`
	YiJi            string         `json:"yi_ji"`
	JiShen          string         `json:"ji_shen"`
	XiongShen       string         `json:"xiong_shen"`
	ChongSha        string         `json:"chong_sha"`
	TaiShen         string         `json:"tai_shen"`
	WuXing          string         `json:"wu_xing"`
	PengZu          string         `json:"peng_zu"`
	Gua             string         `json:"gua"`
	JieQi           string         `json:"jie_qi"`
	ElementImages   []ElementImage `json:"element_images"`
	Score           int            `json:"score"`
	LuckyColor      string         `json:"lucky_color"`
	LuckyNumber     int            `json:"lucky_number"`
	WealthDir       string         `json:"wealth_direction"`
	ClashZodiac     string         `json:"clash_zodiac"`
	AuspiciousHours []string       `json:"auspicious_hours"`
	Analysis        interface{}    `json:"analysis"`
	YiItems         []string       `json:"yi"`
	JiItems         []string       `json:"ji"`
	TodayElements   map[string]int `json:"today_elements"`
	TiaoHou         string         `json:"tiao_hou"`
}

type WeeklyFortuneRequest struct {
	ChartID   uint   `json:"chart_id"`
	StartDate string `json:"start_date"`
}

type WeeklyFortuneResponse struct {
	DailyFortunes []FortuneResponse `json:"daily_fortunes"`
	WeeklyScore   int               `json:"weekly_score"`
	ElementTrend  string            `json:"element_trend"`
}

type MonthlyFortuneRequest struct {
	ChartID uint `json:"chart_id"`
	Year    int  `json:"year"`
	Month   int  `json:"month"`
}

type MonthlyFortuneResponse struct {
	DailyFortunes []FortuneResponse `json:"daily_fortunes"`
	WeeklyScore   int               `json:"weekly_score"`
	ElementTrend  string            `json:"element_trend"`
}

// --- AI Fortune Stub ---

type AIFortuneStubResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// --- Generic DTOs ---

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
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
