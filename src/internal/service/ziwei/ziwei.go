package ziwei

import (
	"fmt"
	"strings"
	"sync"

	"github.com/6tail/tyme4go/tyme"
)

// ──────────────────── Output types ────────────────────

// StarOutput is the JSON-friendly star output with all metadata.
type StarOutput struct {
	Name       string `json:"name"`
	Type       string `json:"type"`  // major/soft/tough/tianma/lucun
	Scope      string `json:"scope"` // origin
	Brightness string `json:"brightness"`
}

// PalaceInfo represents a single palace in the ZiWei chart.
type PalaceInfo struct {
	Name           string       `json:"name"`
	Branch         string       `json:"branch"`
	HeavenlyStem   string       `json:"heavenly_stem"`
	IsBodyPalace   bool         `json:"is_body_palace"`
	Stars          []StarOutput `json:"stars"`
	FourHua        []string     `json:"four_hua"`
	AdjectiveStars []string     `json:"adjective_stars"`
	Changsheng12   string       `json:"changsheng_12"`
	Boshi12        string       `json:"boshi_12"`
	JiangQian12    string       `json:"jiang_qian_12"`
	SuiQian12      string       `json:"sui_qian_12"`

	MainStars  []string          `json:"-"`
	AuxStars   []string          `json:"-"`
	Brightness map[string]string `json:"-"`
}

// ZiWeiChart is the output representation of a full ZiWei Dou Shu chart.
type ZiWeiChart struct {
	Palaces    [12]PalaceInfo `json:"palaces"`
	BodyPalace string         `json:"body_palace"`
	LifeMaster string         `json:"life_master"`
	BodyMaster string         `json:"body_master"`
	FiveBureau string         `json:"five_bureau"`
	Patterns   []string       `json:"patterns"`
	LunarMonth int            `json:"lunar_month"`

	EarthlyBranchOfSoulPalace string `json:"earthly_branch_of_soul_palace"` // 命宫地支名
	EarthlyBranchOfBodyPalace string `json:"earthly_branch_of_body_palace"` // 身宫地支名

	SoulBranch int `json:"-"` // 命宫地支索引 (0=子...11=亥)
	BodyBranch int `json:"-"` // 身宫地支索引
	YearStem   int `json:"-"` // 年干索引
	YearBranch int `json:"-"` // 年支索引
	SoulStem   int `json:"-"` // 命宫天干索引
	JuValue    int `json:"-"` // 五行局数值 (2/3/4/5/6)

	SanfangSizheng [12]SanfangSizhengResult `json:"san_fang_si_zheng"`

	LiuNianStars [12][]string `json:"liu_nian_stars"`
	LiuYueStars  [12][]string `json:"liu_yue_stars"`
	LiuRiStars   [12][]string `json:"liu_ri_stars"`

	birthData *BirthData `json:"-"`
}

// ──────────────────── Flying Star Analysis ────────────────────

type FlyingStarAnalysis struct {
	HuaLu   []FlyTarget `json:"hua_lu"`
	HuaQuan []FlyTarget `json:"hua_quan"`
	HuaKe   []FlyTarget `json:"hua_ke"`
	HuaJi   []FlyTarget `json:"hua_ji"`
}

type FlyTarget struct {
	FromStar string `json:"from_star"`
	ToPalace string `json:"to_palace"`
	Effect   string `json:"effect"`
}

// Dayun is a list of Da Yun stages.
type Dayun []DayunStage

type DayunStage struct {
	StartAge     int      `json:"start_age"`
	EndAge       int      `json:"end_age"`
	Palace       string   `json:"palace"`
	Stars        []string `json:"stars"`
	LiuNianStars []string `json:"liu_nian_stars,omitempty"`
	LiuYueStars  []string `json:"liu_yue_stars,omitempty"`
}

// ──────────────────── Plugin Architecture ────────────────────

type StarPlugin interface {
	Name() string
	TransformStar(star string, palaceIdx int) (string, bool)
	MutatePalace(palace *PalaceInfo) bool
}

var (
	pluginRegistry = make(map[string]StarPlugin)
	pluginMu       sync.RWMutex
)

func RegisterPlugin(p StarPlugin) {
	pluginMu.Lock()
	defer pluginMu.Unlock()
	pluginRegistry[p.Name()] = p
}

func ApplyPlugins(chart *ZiWeiChart) {
	pluginMu.RLock()
	defer pluginMu.RUnlock()
	for _, plugin := range pluginRegistry {
		for i := range chart.Palaces {
			if newStar, ok := plugin.TransformStar("", i); ok {
				chart.Palaces[i].AuxStars = append(chart.Palaces[i].AuxStars, newStar)
			}
			plugin.MutatePalace(&chart.Palaces[i])
		}
	}
}

// ──────────────────── Service ────────────────────

// ZiWeiService provides ZiWei Dou Shu calculation methods.
// No longer depends on ziwei-zenith engine.
type ZiWeiService struct {
	mu sync.Mutex
}

func NewZiWeiService() *ZiWeiService {
	return &ZiWeiService{}
}

// SetAlgorithm is kept for API compatibility but no longer affects calculation.
func (s *ZiWeiService) SetAlgorithm(alg AlgorithmType) {}

// GetAlgorithm returns the default algorithm type (reserved for future use).
func (s *ZiWeiService) GetAlgorithm() int { return 0 }

// CalculateChart computes a full ZiWei Dou Shu chart from solar birth data.
func (s *ZiWeiService) CalculateChart(year, month, day, hour, minute int, gender string) (*ZiWeiChart, error) {
	birth, err := buildBirthData(year, month, day, hour, minute, gender)
	if err != nil {
		return nil, fmt.Errorf("build birth data: %w", err)
	}

	chart, err := CalculateZiWeiChart(birth)
	if err != nil {
		return nil, fmt.Errorf("calculate chart: %w", err)
	}

	ApplyPlugins(chart)
	return chart, nil
}

// DetectLocalPatterns detects patterns using the local knowledge layer.
func (s *ZiWeiService) DetectLocalPatterns(chart *ZiWeiChart) []string {
	return DetectLocalPatterns(chart)
}

// AnalyzeFlyingStars performs 四化飞星 analysis on the chart.
func (s *ZiWeiService) AnalyzeFlyingStars(chart *ZiWeiChart) *FlyingStarAnalysis {
	if chart == nil {
		return nil
	}
	return buildFlyingStarAnalysisFromChart(chart)
}

// CalculateDayun computes the 大限 (10-year luck periods).
func (s *ZiWeiService) CalculateDayun(chart *ZiWeiChart) Dayun {
	if chart == nil {
		return nil
	}
	return calcDayunFromChart(chart)
}

// CalculateLiunian computes the 流年 overlay for a given year.
func (s *ZiWeiService) CalculateLiunian(chart *ZiWeiChart, targetYear int) *ZiWeiChart {
	if chart == nil {
		return nil
	}
	result := *chart

	liunianStem := (targetYear - 4) % 10
	if liunianStem < 0 {
		liunianStem += 10
	}
	liunianBranch := (targetYear - 4) % 12
	if liunianBranch < 0 {
		liunianBranch += 12
	}

	liunianHua := calcFourHua(liunianStem)

	var liunianStars [12][]string
	for i := 0; i < 12; i++ {
		palaceStars := make([]string, 0)

		for _, star := range result.Palaces[i].MainStars {
			if label, ok := liunianHua[star]; ok {
				palaceStars = append(palaceStars, star+label)
			}
		}
		for _, star := range result.Palaces[i].AuxStars {
			if label, ok := liunianHua[star]; ok {
				palaceStars = append(palaceStars, star+label)
			}
		}

		luZnIdx := LucunBranchIdx[liunianStem]
		qyIdx := fixIndex(luZnIdx + 1)
		tlIdx := fixIndex(luZnIdx - 1)
		tmIdx := TianmaBranchIdx[liunianBranch]

		branchName := BranchNames[fixIndex(liunianBranch)]
		if i == luZnIdx {
			palaceStars = append(palaceStars, "流禄")
		}
		if i == qyIdx {
			palaceStars = append(palaceStars, "流羊")
		}
		if i == tlIdx {
			palaceStars = append(palaceStars, "流陀")
		}
		if i == tmIdx {
			palaceStars = append(palaceStars, "流马")
		}

		_ = branchName
		liunianStars[i] = palaceStars
	}
	result.LiuNianStars = liunianStars

	return &result
}

// CalculateLiuyue computes the 流月 overlay.
func (s *ZiWeiService) CalculateLiuyue(chart *ZiWeiChart, lunarMonth int) *ZiWeiChart {
	if chart == nil {
		return nil
	}
	result := *chart

	yearStem := chart.YearStem
	if chart.birthData != nil {
		yearStem = chart.birthData.YearStem
	}
	liuyueStem := (yearStem*2 + lunarMonth - 1) % 10
	if liuyueStem < 0 {
		liuyueStem += 10
	}
	liuyueHua := calcFourHua(liuyueStem)
	var liuyueStars [12][]string
	for i := 0; i < 12; i++ {
		palaceStars := make([]string, 0)
		for _, star := range result.Palaces[i].MainStars {
			if label, ok := liuyueHua[star]; ok {
				palaceStars = append(palaceStars, star+label)
			}
		}
		liuyueStars[i] = palaceStars
	}
	result.LiuYueStars = liuyueStars

	return &result
}

// CalculateLiuri computes the 流日 overlay.
func (s *ZiWeiService) CalculateLiuri(chart *ZiWeiChart, lunarDay int) *ZiWeiChart {
	if chart == nil {
		return nil
	}
	result := *chart

	dayStem := chart.YearStem
	if chart.birthData != nil {
		dayStem = chart.birthData.DayStem
	}
	liuriStem := (dayStem*2 + lunarDay - 1) % 10
	if liuriStem < 0 {
		liuriStem += 10
	}
	liuriHua := calcFourHua(liuriStem)
	var liuriStars [12][]string
	for i := 0; i < 12; i++ {
		palaceStars := make([]string, 0)
		for _, star := range result.Palaces[i].MainStars {
			if label, ok := liuriHua[star]; ok {
				palaceStars = append(palaceStars, star+label)
			}
		}
		liuriStars[i] = palaceStars
	}
	result.LiuRiStars = liuriStars

	return &result
}

// AnalyzeSihuaChain performs full sihua chain analysis.
func (s *ZiWeiService) AnalyzeSihuaChain(chart *ZiWeiChart) *SihuaChainResult {
	return AnalyzeSihuaChain(chart)
}

// GetPalaceReading returns a template-based reading for a specific palace.
func (s *ZiWeiService) GetPalaceReading(chart *ZiWeiChart, palaceIdx int) *PalaceReading {
	return GetPalaceReading(chart, palaceIdx)
}

// AnalyzeHeming performs compatibility analysis between two charts.
func (s *ZiWeiService) AnalyzeHeming(chartA, chartB *ZiWeiChart) *HemingResult {
	return AnalyzeHeming(chartA, chartB)
}

// AnalyzeSelfMutagen detects self-mutagen occurrences.
func (s *ZiWeiService) AnalyzeSelfMutagen(chart *ZiWeiChart) []SelfMutagenResult {
	return DetectSelfMutagens(chart)
}

// ──────────────────── Birth Data ────────────────────

// buildBirthData converts solar date parameters to BirthData for the new engine.
func buildBirthData(year, month, day, hour, minute int, gender string) (*BirthData, error) {
	var g string
	switch strings.TrimSpace(strings.ToUpper(gender)) {
	case "男", "MALE", "M":
		g = "男"
	case "女", "FEMALE", "F":
		g = "女"
	default:
		return nil, fmt.Errorf("gender must be 男/女 or male/female or MALE/FEMALE")
	}

	// Use tyme4go for lunar calendar conversion
	st, err := tyme.SolarTime{}.FromYmdHms(year, month, day, hour, minute, 0)
	if err != nil {
		return nil, fmt.Errorf("invalid solar date: %w", err)
	}
	lunarHour := st.GetLunarHour()
	lunarDayObj := lunarHour.GetLunarDay()
	lunarMonthObj := lunarDayObj.GetLunarMonth()
	ec := lunarHour.GetEightChar()

	yearPillarName := lunarMonthObj.GetLunarYear().GetSixtyCycle().GetName()
	monthPillarName := ec.GetMonth().GetName()
	dayPillarName := ec.GetDay().GetName()
	hourPillarName := ec.GetHour().GetName()

	yearStem := stemFromRune([]rune(yearPillarName)[0])
	yearBranch := branchFromRune([]rune(yearPillarName)[1])
	monthBranch := branchFromRune([]rune(monthPillarName)[1])

	// Determine the hour branch from the input hour (not from the pillar)
	// 子时 starts at 23:00, so hour 23-1 maps to 子(0), hour 1-3 maps to 丑(1), etc.
	hourBranch := ((hour + 1) / 2) % 12

	lunarMonthWithLeap := lunarMonthObj.GetMonthWithLeap()
	isLeapMonth := lunarMonthObj.IsLeap()
	lunarMonthInt := lunarMonthWithLeap
	if lunarMonthInt < 0 {
		lunarMonthInt = -lunarMonthInt
	}

	// Handle 晚子时: if hour is 23 (11pm), the day advances
	lunarDay := lunarHour.GetDay()

	return &BirthData{
		SolarYear:         year,
		SolarMonth:        month,
		SolarDay:          day,
		Hour:              hour,
		Minute:            minute,
		Gender:            g,
		LunarYear:         lunarDayObj.GetYear(),
		LunarMonth:        lunarMonthInt,
		LunarDay:          lunarDay,
		YearStem:          yearStem,
		YearBranch:        yearBranch,
		MonthPillarBranch: monthBranch,
		DayStem:           stemFromRune([]rune(dayPillarName)[0]),
		DayBranch:         branchFromRune([]rune(dayPillarName)[1]),
		HourBranch:        hourBranch,
		HourStem:          stemFromRune([]rune(hourPillarName)[0]),
		IsLeapMonth:       isLeapMonth,
	}, nil
}

// stemFromRune converts a heavenly stem rune to its index.
func stemFromRune(r rune) int {
	s := string(r)
	for i, n := range StemNames {
		if n == s {
			return i
		}
	}
	return 0
}

// branchFromRune converts an earthly branch rune to its index.
func branchFromRune(r rune) int {
	s := string(r)
	for i, n := range BranchNames {
		if n == s {
			return i
		}
	}
	return 0
}

// ──────────────────── Flying Star Analysis (new engine) ────────────────────

func buildFlyingStarAnalysisFromChart(chart *ZiWeiChart) *FlyingStarAnalysis {
	if chart == nil {
		return nil
	}

	analysis := &FlyingStarAnalysis{}
	hua := SiHuaTable[chart.YearStem]

	// Build star→palace mapping
	starToPalace := make(map[string]string)
	for _, p := range chart.Palaces {
		for _, star := range p.MainStars {
			starToPalace[star] = p.Name
		}
		for _, star := range p.AuxStars {
			starToPalace[star] = p.Name
		}
	}

	palaceEffects := map[string]string{
		"命宫": "直接影响个人运势与性格", "兄弟": "影响兄弟姐妹关系与助力",
		"夫妻": "影响婚姻感情与配偶关系", "子女": "影响子女缘分与下属关系",
		"财帛": "影响财运与金钱进出", "疾厄": "影响身体健康状况",
		"迁移": "影响外出运程与社会形象", "交友": "影响朋友与部属关系",
		"事业": "影响事业运程与工作成就", "田宅": "影响房产运程与家庭环境",
		"福德": "影响精神享受与内心世界", "父母": "影响父母缘分与长辈助力",
	}

	addTarget := func(starName, huaType string, target *[]FlyTarget) {
		palace, found := starToPalace[starName]
		if !found {
			return
		}
		effect := ""
		if d, ok := palaceEffects[palace]; ok {
			effect = fmt.Sprintf("%s%s飞入%s，%s", starName, huaType, palace, d)
		} else {
			effect = fmt.Sprintf("%s%s飞入%s", starName, huaType, palace)
		}
		*target = append(*target, FlyTarget{FromStar: starName, ToPalace: palace, Effect: effect})
	}

	addTarget(hua[0], "化禄", &analysis.HuaLu)
	addTarget(hua[1], "化权", &analysis.HuaQuan)
	addTarget(hua[2], "化科", &analysis.HuaKe)
	addTarget(hua[3], "化忌", &analysis.HuaJi)

	return analysis
}

// ──────────────────── Dayun (大限) ────────────────────

func calcDayunFromChart(chart *ZiWeiChart) Dayun {
	if chart == nil {
		return nil
	}

	gender := ""
	if chart.birthData != nil {
		gender = chart.birthData.Gender
	}

	isMale := gender == "男" || gender == "MALE" || gender == "M"
	forward := (BranchIsYang(chart.YearBranch) && isMale) || (!BranchIsYang(chart.YearBranch) && !isMale)

	// 大限起运年龄: 五行局值
	juValue := chart.JuValue
	startAge := juValue

	// 大限宫位从命宫(或其相邻宫)开始
	var result Dayun
	soulBranch := chart.SoulBranch
	dir := 1
	if !forward {
		dir = -1
	}

	for i := range chart.Palaces {
		branch := fixIndex(soulBranch + i*dir)
		palaceIdx := -1
		for j, p := range chart.Palaces {
			if p.Branch == BranchNames[branch] {
				palaceIdx = j
				break
			}
		}
		if palaceIdx < 0 {
			palaceIdx = i
		}

		stage := DayunStage{
			StartAge: startAge + i*10,
			EndAge:   startAge + (i+1)*10 - 1,
			Palace:   chart.Palaces[palaceIdx].Name,
		}
		for _, star := range chart.Palaces[palaceIdx].MainStars {
			stage.Stars = append(stage.Stars, star)
		}
		for _, star := range chart.Palaces[palaceIdx].AuxStars {
			stage.Stars = append(stage.Stars, star)
		}
		result = append(result, stage)
	}
	return result
}

// ──────────────────── Compatibility aliases ────────────────────

// AlgorithmType is kept for API compatibility but no longer used.
type AlgorithmType int

const (
	AlgorithmFullBook AlgorithmType = iota
	AlgorithmZhongZhou
)

// NewZiWeiServiceWithAlgorithm creates a ZiWeiService (algorithm param is ignored).
func NewZiWeiServiceWithAlgorithm(_ AlgorithmType) *ZiWeiService {
	return NewZiWeiService()
}

// GetBirthData returns the birth data used for chart calculation.
func (c *ZiWeiChart) GetBirthData() *BirthData {
	return c.birthData
}
