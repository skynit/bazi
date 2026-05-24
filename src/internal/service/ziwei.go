package service

import (
	"errors"
	"fmt"
	"strings"

	"github.com/6tail/tyme4go/tyme"
	"github.com/kaecer68/ziwei-zenith/pkg/basis"
	"github.com/kaecer68/ziwei-zenith/pkg/engine"
)

// ──────────────────── Output types ────────────────────

// PalaceInfo represents a single palace in the ZiWei chart.
type PalaceInfo struct {
	Name           string            `json:"name"`
	MainStars      []string          `json:"main_stars"`
	AuxStars       []string          `json:"aux_stars"`
	Brightness     map[string]string `json:"brightness"`
	FourHua        []string          `json:"four_hua"`
	AdjectiveStars []string          `json:"adjective_stars"`
	Changsheng12   string            `json:"changsheng_12"`
	Boshi12        string            `json:"boshi_12"`
	JiangQian12    string            `json:"jiang_qian_12"`
	SuiQian12      string            `json:"sui_qian_12"`
}

// ZiWeiChart is the output representation of a full ZiWei Dou Shu chart.
type ZiWeiChart struct {
	Palaces    [12]PalaceInfo `json:"palaces"`
	BodyPalace string         `json:"body_palace"`
	LifeMaster string         `json:"life_master"`
	BodyMaster string         `json:"body_master"`
	FiveBureau string         `json:"five_bureau"`
	Patterns   []string       `json:"patterns"`
	LunarMonth int            `json:"lunar_month"` // 1-12, from birth info

	// SanfangSizheng holds precomputed 三方四正 for all 12 palaces
	SanfangSizheng [12]SanfangSizhengResult `json:"san_fang_si_zheng"`

	// 流耀 star overlays per time scope
	LiuNianStars [12][]string `json:"liu_nian_stars"` // 流年星 per palace (10 stars)
	LiuYueStars  [12][]string `json:"liu_yue_stars"`  // 流月星 per palace (10 stars)
	LiuRiStars   [12][]string `json:"liu_ri_stars"`   // 流日星 per palace (10 stars)

	// internal reference for overlay calculations
	engineChart *engine.ZiweiChart `json:"-"`
	birthInfo   basis.BirthInfo    `json:"-"`
}

// BirthInfo returns the birth info used to build this chart.
func (c *ZiWeiChart) BirthInfo() basis.BirthInfo {
	return c.birthInfo
}

// ──────────────────── Flying Star Analysis ────────────────────
type FlyingStarAnalysis struct {
	HuaLu   []FlyTarget `json:"hua_lu"`
	HuaQuan []FlyTarget `json:"hua_quan"`
	HuaKe   []FlyTarget `json:"hua_ke"`
	HuaJi   []FlyTarget `json:"hua_ji"`
}

// FlyTarget describes a single flying star target.
type FlyTarget struct {
	FromStar string `json:"from_star"`
	ToPalace string `json:"to_palace"`
	Effect   string `json:"effect"`
}

// Dayun is a list of Da Yun stages.
type Dayun []DayunStage

// DayunStage represents one 10-year luck period.
type DayunStage struct {
	StartAge     int      `json:"start_age"`
	EndAge       int      `json:"end_age"`
	Palace       string   `json:"palace"`
	Stars        []string `json:"stars"`
	LiuNianStars []string `json:"liu_nian_stars,omitempty"` // 流年星 overlay
	LiuYueStars  []string `json:"liu_yue_stars,omitempty"`  // 流月星 overlay
}

var branchOrder = []string{"寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥", "子", "丑"}

// ──────────────────── Plugin Architecture ────────────────────

// StarPlugin defines an extension point for customizing star placement and transformation.
type StarPlugin interface {
	Name() string
	TransformStar(star string, palaceIdx int) (string, bool)
	MutatePalace(palace *PalaceInfo) bool
}

// pluginRegistry is the global plugin registry.
var pluginRegistry = make(map[string]StarPlugin)

// RegisterPlugin registers a star plugin. Called at startup.
func RegisterPlugin(p StarPlugin) {
	pluginRegistry[p.Name()] = p
}

// ApplyPlugins applies all registered plugins to a chart after it is computed.
func ApplyPlugins(chart *ZiWeiChart) {
	for _, plugin := range pluginRegistry {
		for i := range chart.Palaces {
			if newStar, ok := plugin.TransformStar("", i); ok {
				chart.Palaces[i].AuxStars = append(chart.Palaces[i].AuxStars, newStar)
			}
			plugin.MutatePalace(&chart.Palaces[i])
		}
	}
}

// ──────────────────── Algorithm Types ────────────────────

type AlgorithmType int

const (
    AlgorithmFullBook AlgorithmType = iota // default: 全书派
    AlgorithmZhongZhou                      // 中州派
)

// ──────────────────── Service ────────────────────

// ZiWeiService provides ZiWei Dou Shu calculation methods.
type ZiWeiService struct {
    eng       *engine.ZiweiEngine
    algorithm AlgorithmType
}

// NewZiWeiService creates a new ZiWeiService.
func NewZiWeiService() *ZiWeiService {
    return &ZiWeiService{eng: engine.New(), algorithm: AlgorithmFullBook}
}

// NewZiWeiServiceWithAlgorithm creates a ZiWeiService with a specific algorithm.
func NewZiWeiServiceWithAlgorithm(alg AlgorithmType) *ZiWeiService {
    return &ZiWeiService{eng: engine.New(), algorithm: alg}
}

// SetAlgorithm sets the algorithm type for this service.
func (s *ZiWeiService) SetAlgorithm(alg AlgorithmType) {
    s.algorithm = alg
}

// CalculateChart computes a full ZiWei Dou Shu chart from solar birth data.
func (s *ZiWeiService) CalculateChart(year, month, day, hour, minute int, gender string) (*ZiWeiChart, error) {
	birth, err := buildBirthInfo(year, month, day, hour, minute, gender)
	if err != nil {
		return nil, fmt.Errorf("build birth info: %w", err)
	}

	engChart, err := s.eng.BuildChart(birth)
	if err != nil {
		return nil, fmt.Errorf("build chart: %w", err)
	}

	chart := mapEngineChart(engChart, birth, s)
	ApplyPlugins(chart)
	return chart, nil
}

// DetectPatterns detects fortune patterns (格局) in the chart.
func (s *ZiWeiService) DetectPatterns(chart *ZiWeiChart) []string {
	if chart == nil || chart.engineChart == nil {
		return nil
	}
	patterns := engine.DetectPatterns(chart.engineChart)
	result := make([]string, 0, len(patterns))
	for _, p := range patterns {
		result = append(result, p.Name)
	}
	return result
}

// AnalyzeFlyingStars performs 四化飞星 analysis on the chart.
func (s *ZiWeiService) AnalyzeFlyingStars(chart *ZiWeiChart) *FlyingStarAnalysis {
	if chart == nil || chart.engineChart == nil {
		return nil
	}
	return buildFlyingStarAnalysis(chart.engineChart)
}

// CalculateDayun computes the 大限 (10-year luck periods).
func (s *ZiWeiService) CalculateDayun(chart *ZiWeiChart) Dayun {
	if chart == nil || chart.engineChart == nil {
		return nil
	}
	return mapDayun(chart.engineChart)
}

// CalculateLiunian computes the 流年 overlay for a given year.
func (s *ZiWeiService) CalculateLiunian(chart *ZiWeiChart, targetYear int) *ZiWeiChart {
	if chart == nil || chart.engineChart == nil {
		return nil
	}
	engChart := chart.engineChart
	yearStem, yearBranch := computeYearStemBranch(targetYear)
	liuNian := engine.CalcLiuNian(yearBranch, targetYear)
	liuNian.Stem = yearStem

	// Clone the chart and overlay LiuNian
	clone := *engChart
	clone.LiuNian = liuNian
	result := mapEngineChart(&clone, chart.birthInfo, nil)

	// Compute 流年星耀 (10 stars) for each palace.
	// 流耀星按流年天干四化 + 流年支本位星随年变化，
	// 这里通过 engChart.LiuNianStars 取四化星（已有数据），其余用固定流耀列表补充。
	yearStemIdx := int(yearStem)
	liuyaoList := getLiuYaoByYear(targetYear, yearStemIdx)
	for i := 0; i < 12; i++ {
		palace := basis.Palace(i)
		actualBranch := palaceBranch(engChart, palace)
		// 优先取库计算的四化星
		var stars []string
		if s4, ok := engChart.LiuNianStars[actualBranch]; ok {
			stars = stringifyInterfaces(s4)
		}
		// 追加流耀星（不重复）
		for _, ly := range liuyaoList {
			has := false
			for _, s := range stars {
				if s == ly {
					has = true
					break
				}
			}
			if !has {
				stars = append(stars, ly)
			}
		}
		result.LiuNianStars[i] = stars
	}
	return result
}

// getLiuYaoByYear returns the 流耀星 list for a given year and stem index.
// The 10 流耀 stars include: 天魁, 天钺, 文昌, 文曲, 禄存, 擎羊, 陀罗, 天马, 红鸾, 天喜.
// Some stars change position based on year branch (天马, 红鸾, 天喜).
func getLiuYaoByYear(year, stemIdx int) []string {
	base := []string{"天魁", "天钺", "文昌", "文曲", "禄存", "擎羊", "陀罗", "天马", "红鸾", "天喜"}
	// 天马 moves by year branch (not fully implemented in this simplified version)
	// Return fixed list for now
	_ = stemIdx // stem not used in simplified version
	return base
}

// CalculateLiuyue computes the 流月 overlay for a given lunar month.
func (s *ZiWeiService) CalculateLiuyue(chart *ZiWeiChart, lunarMonth int) *ZiWeiChart {
	if chart == nil || chart.engineChart == nil {
		return nil
	}
	engChart := chart.engineChart

	// Ensure LiuNian is computed first (Liuyue depends on it)
	if engChart.LiuNian.Branch == 0 {
		year := chart.birthInfo.SolarYear
		_, yearBranch := computeYearStemBranch(year)
		liuNian := engine.CalcLiuNian(yearBranch, year)
		engChart.LiuNian = liuNian
	}

	lnBranch := engChart.LiuNian.Branch
	liuYue := engine.CalcLiuYue(lnBranch, chart.birthInfo.LunarMonth, basis.Branch(chart.birthInfo.HourBranch), lunarMonth)

	clone := *engChart
	clone.LiuYue = liuYue
	result := mapEngineChart(&clone, chart.birthInfo, nil)
	result.LiuYueStars = computeLiuYueStars(chart, lunarMonth)
	return result
}

// CalculateLiuri computes the 流日 overlay for a given lunar day.
func (s *ZiWeiService) CalculateLiuri(chart *ZiWeiChart, lunarDay int) *ZiWeiChart {
	if chart == nil || chart.engineChart == nil {
		return nil
	}
	engChart := chart.engineChart

	// Ensure the full chain: LiuNian -> LiuYue -> LiuRi
	if engChart.LiuNian.Branch == 0 {
		year := chart.birthInfo.SolarYear
		_, yearBranch := computeYearStemBranch(year)
		liuNian := engine.CalcLiuNian(yearBranch, year)
		engChart.LiuNian = liuNian
	}
	if engChart.LiuYue == 0 {
		lnBranch := engChart.LiuNian.Branch
		liuYue := engine.CalcLiuYue(lnBranch, chart.birthInfo.LunarMonth, basis.Branch(chart.birthInfo.HourBranch), 1)
		engChart.LiuYue = liuYue
	}

	lyBranch := engChart.LiuYue
	liuRi := engine.CalcLiuRi(lyBranch, lunarDay)

	clone := *engChart
	clone.LiuRi = liuRi
	result := mapEngineChart(&clone, chart.birthInfo, nil)
	result.LiuRiStars = computeLiuRiStars(chart, lunarDay)
	return result
}

// AnalyzeSihuaChain performs full sihua chain analysis using the local knowledge layer.
func (s *ZiWeiService) AnalyzeSihuaChain(chart *ZiWeiChart) *SihuaChainResult {
	return AnalyzeSihuaChain(chart)
}

// DetectLocalPatterns detects patterns using the local knowledge layer (not engine).
func (s *ZiWeiService) DetectLocalPatterns(chart *ZiWeiChart) []string {
	return DetectLocalPatterns(chart)
}

// GetPalaceReading returns a template-based reading for a specific palace.
func (s *ZiWeiService) GetPalaceReading(chart *ZiWeiChart, palaceIdx int) *PalaceReading {
	return GetPalaceReading(chart, palaceIdx)
}

// AnalyzeHeming performs compatibility analysis between two charts.
func (s *ZiWeiService) AnalyzeHeming(chartA, chartB *ZiWeiChart) *HemingResult {
	return AnalyzeHeming(chartA, chartB)
}

// AnalyzeSelfMutagen detects self-mutagen occurrences in a chart.
func (s *ZiWeiService) AnalyzeSelfMutagen(chart *ZiWeiChart) []SelfMutagenResult {
	return DetectSelfMutagens(chart)
}

// ──────────────────── Internal helpers ────────────────────

// buildBirthInfo converts solar date parameters to basis.BirthInfo.
func buildBirthInfo(year, month, day, hour, minute int, gender string) (basis.BirthInfo, error) {
	var sex basis.Sex
	switch strings.TrimSpace(strings.ToUpper(gender)) {
	case "男", "MALE", "M", "m":
		sex = basis.SexMale
	case "女", "FEMALE", "F", "f":
		sex = basis.SexFemale
	default:
		return basis.BirthInfo{}, errors.New("gender must be 男/女 or male/female or MALE/FEMALE")
	}

	st, err := tyme.SolarTime{}.FromYmdHms(year, month, day, hour, minute, 0)
	if err != nil {
		return basis.BirthInfo{}, fmt.Errorf("invalid solar date: %w", err)
	}

	lunarHour := st.GetLunarHour()
	ec := lunarHour.GetEightChar()

	yearPillar := pillarFromName(ec.GetYear().GetName())
	monthPillar := pillarFromName(ec.GetMonth().GetName())
	dayPillar := pillarFromName(ec.GetDay().GetName())
	hourPillar := pillarFromName(ec.GetHour().GetName())

	lunarMonth := lunarHour.GetMonth()

	return basis.BirthInfo{
		SolarYear:   year,
		SolarMonth:  month,
		SolarDay:    day,
		Hour:        hour,
		Minute:      minute,
		Sex:         sex,
		LunarYear:   lunarHour.GetYear(),
		LunarMonth:  lunarMonth,
		LunarDay:    lunarHour.GetDay(),
		HourBranch:  basis.HourBranchFromTime(hour),
		YearPillar:  yearPillar,
		MonthPillar: monthPillar,
		DayPillar:   dayPillar,
		HourPillar:  hourPillar,
	}, nil
}

// pillarFromName converts a sexagenary name like "乙酉" to basis.Pillar.
func pillarFromName(name string) basis.Pillar {
	if len(name) < 2 {
		return basis.Pillar{}
	}
	return basis.Pillar{
		Stem:   stemFromRune([]rune(name)[0]),
		Branch: branchFromRune([]rune(name)[1]),
	}
}

var stemNames = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var branchNames = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

func stemFromRune(r rune) basis.Stem {
	s := string(r)
	for i, n := range stemNames {
		if n == s {
			return basis.Stem(i)
		}
	}
	return basis.Stem(0)
}

func branchFromRune(r rune) basis.Branch {
	s := string(r)
	for i, n := range branchNames {
		if n == s {
			return basis.Branch(i)
		}
	}
	return basis.Branch(0)
}

// computeYearStemBranch returns the stem and branch for a given Gregorian year.
func computeYearStemBranch(year int) (basis.Stem, basis.Branch) {
	idx := (year - 4) % 60
	if idx < 0 {
		idx += 60
	}
	return basis.Stem(idx % 10), basis.Branch(idx % 12)
}

// computeMingGongBranchZhongZhou computes 命宫 branch using 中州派 algorithm.
// 中州派 uses year branch to find 命主: mingZhuIdx = (yearBranch + 1) % 12
func computeMingGongBranchZhongZhou(yearBranch int) string {
	return branchOrder[(yearBranch+1)%12]
}

// mapEngineChart converts the engine's ZiweiChart into our output ZiWeiChart.
func mapEngineChart(ec *engine.ZiweiChart, birth basis.BirthInfo, s *ZiWeiService) *ZiWeiChart {
	if ec == nil {
		return nil
	}

	chart := &ZiWeiChart{
		BodyPalace:  ec.LifePalace.ShenGong.String(),
		LifeMaster:  lifeMaster(birth.YearPillar.Branch),
		BodyMaster:  bodyMaster(birth.YearPillar.Branch),
		FiveBureau:  ec.Wuxing.String(),
		LunarMonth:  birth.LunarMonth,
		engineChart: ec,
		birthInfo:   birth,
	}

	// Build PalaceInfo for each of the 12 palaces
	for i := 0; i < 12; i++ {
		palace := basis.Palace(i)
		branch := palaceBranch(ec, palace)
		chart.Palaces[i] = PalaceInfo{
			Name:       palace.String(),
			MainStars:  starNames(ec.Stars[branch]),
			AuxStars:   auxStarNames(ec, branch),
			Brightness: brightnessMap(ec, branch),
			FourHua:    transformationNames(ec.TransformedStars[branch]),
		}
		// Precompute 三方四正 for this palace
		chart.SanfangSizheng[i] = *GetPalaceSanfang(i)
	}

	// Compute adjective stars across all 12 palaces
	adjStars := ComputeAdjectiveStars(chart)
	for i := 0; i < 12; i++ {
		chart.Palaces[i].AdjectiveStars = adjStars[i]
	}

	// Compute twelve shen systems for all palaces
	twelveShen := ComputeTwelveShen(chart)
	for i := 0; i < 12; i++ {
		chart.Palaces[i].Changsheng12 = twelveShen[i].Changsheng
		chart.Palaces[i].Boshi12 = twelveShen[i].Boshi
		chart.Palaces[i].JiangQian12 = twelveShen[i].Jiangqian
		chart.Palaces[i].SuiQian12 = twelveShen[i].Suiqian
	}

	// Collect pattern names
	for _, p := range ec.Patterns {
		chart.Patterns = append(chart.Patterns, p.Name)
	}

	return chart
}

// computeLiuNianStars computes 流年星 (10 stars) for each palace for a given year.
// The same 10 stars appear as an overlay in each palace.
func computeLiuNianStars(chart *ZiWeiChart, year int) [12][]string {
	if chart == nil {
		return [12][]string{}
	}
	stars := [12][]string{}
	for i := 0; i < 12; i++ {
		stars[i] = LIU_YAO_STARS
	}
	return stars
}

// computeLiuYueStars computes 流月星 (10 stars) for each palace for a given month.
func computeLiuYueStars(chart *ZiWeiChart, month int) [12][]string {
	if chart == nil {
		return [12][]string{}
	}
	stars := [12][]string{}
	for i := 0; i < 12; i++ {
		stars[i] = LIU_YAO_STARS
	}
	return stars
}

// computeLiuRiStars computes 流日星 (10 stars) for each palace for a given day.
func computeLiuRiStars(chart *ZiWeiChart, day int) [12][]string {
	if chart == nil {
		return [12][]string{}
	}
	stars := [12][]string{}
	for i := 0; i < 12; i++ {
		stars[i] = LIU_YAO_STARS
	}
	return stars
}

// palaceBranch returns the Branch associated with a given Palace.
func palaceBranch(ec *engine.ZiweiChart, p basis.Palace) basis.Branch {
	for b, pal := range ec.Palaces {
		if pal == p {
			return b
		}
	}
	return 0
}

// stringifyInterfaces converts []interface{} to []string.
func stringifyInterfaces(v []interface{}) []string {
	if v == nil {
		return nil
	}
	out := make([]string, len(v))
	for i, e := range v {
		if s, ok := e.(string); ok {
			out[i] = s
		} else {
			out[i] = fmt.Sprintf("%v", e)
		}
	}
	return out
}

// starNames converts []basis.Star to []string.
func starNames(stars []basis.Star) []string {
	names := make([]string, len(stars))
	for i, s := range stars {
		names[i] = s.String()
	}
	return names
}

// auxStarNames collects all non-main-star names in a branch.
func auxStarNames(ec *engine.ZiweiChart, b basis.Branch) []string {
	var names []string
	for _, s := range ec.AssistantStars[b] {
		if strer, ok := s.(fmt.Stringer); ok {
			names = append(names, strer.String())
		}
	}
	for _, s := range ec.SecondaryStars[b] {
		if strer, ok := s.(fmt.Stringer); ok {
			names = append(names, strer.String())
		}
	}
	return names
}

// brightnessMap builds the star→brightness mapping for a branch.
func brightnessMap(ec *engine.ZiweiChart, b basis.Branch) map[string]string {
	m := make(map[string]string)
	for _, sb := range ec.StarBrightness {
		if sb.Branch == b {
			m[sb.Star.String()] = sb.Brightness.String()
		}
	}
	return m
}

// transformationNames collects 四化 names from transformed stars.
func transformationNames(stars []interface{}) []string {
	var names []string
	for _, s := range stars {
		if ts, ok := s.(basis.TransformedStar); ok {
			names = append(names, ts.String())
		}
	}
	return names
}

// lifeMaster returns the 命主 star name based on the year branch.
func lifeMaster(yearBranch basis.Branch) string {
	master := []string{
		"贪狼", "巨门", "禄存", "文曲",
		"廉贞", "武曲", "武曲", "破军",
		"武曲", "廉贞", "文曲", "禄存",
	}
	return master[int(yearBranch)%12]
}

// bodyMaster returns the 身主 star name based on the year branch.
func bodyMaster(yearBranch basis.Branch) string {
	master := []string{
		"铃星", "天相", "天梁", "天同",
		"文昌", "天机", "火星", "天相",
		"天梁", "天同", "文昌", "天机",
	}
	return master[int(yearBranch)%12]
}

// mapDayun converts engine DaYun to our output Dayun.
func mapDayun(ec *engine.ZiweiChart) Dayun {
	if ec == nil {
		return nil
	}

	branchOfPalace := make(map[basis.Branch]basis.Palace)
	for b, p := range ec.Palaces {
		branchOfPalace[b] = p
	}

	var result Dayun
	for _, dy := range ec.DaYun {
		stage := DayunStage{
			StartAge: dy.StartAge,
			EndAge:   dy.EndAge,
			Palace:   palaceName(branchOfPalace, dy.Branch),
		}
		// Collect stars in this branch for the dayun stage
		stage.Stars = append(stage.Stars, starNames(ec.Stars[dy.Branch])...)
		for _, s := range ec.AssistantStars[dy.Branch] {
			if strer, ok := s.(fmt.Stringer); ok {
				stage.Stars = append(stage.Stars, strer.String())
			}
		}
		stage.Stars = append(stage.Stars, transformationNames(ec.TransformedStars[dy.Branch])...)
		result = append(result, stage)
	}
	return result
}

// palaceName returns the palace name for a given branch.
func palaceName(branchOfPalace map[basis.Branch]basis.Palace, b basis.Branch) string {
	if p, ok := branchOfPalace[b]; ok {
		return p.String()
	}
	return ""
}

// buildFlyingStarAnalysis builds the 四化飞星 analysis.
func buildFlyingStarAnalysis(ec *engine.ZiweiChart) *FlyingStarAnalysis {
	if ec == nil {
		return nil
	}

	analysis := &FlyingStarAnalysis{}

	yearStem := ec.YearPillar.Stem
	trans, ok := basis.TransformationTable[yearStem]
	if !ok {
		return analysis
	}

	huaLuStar := trans[0]   // 化禄
	huaQuanStar := trans[1] // 化权
	huaKeStar := trans[2]   // 化科
	huaJiStar := trans[3]   // 化忌

	branchOfPalace := make(map[basis.Branch]basis.Palace)
	for b, p := range ec.Palaces {
		branchOfPalace[b] = p
	}

	// Find where each transformed star is located
	starToBranch := make(map[string]basis.Branch)
	for b, stars := range ec.Stars {
		for _, s := range stars {
			starToBranch[s.String()] = b
		}
	}

	// Build targets for each transformation
	addTarget := func(starName string, huaType string, target *[]FlyTarget) {
		b, found := starToBranch[starName]
		if !found {
			return
		}
		palace := branchOfPalace[b]
		effect := flyEffect(huaType, starName, palace)
		*target = append(*target, FlyTarget{
			FromStar: starName,
			ToPalace: palace.String(),
			Effect:   effect,
		})
	}

	addTarget(huaLuStar, "化禄", &analysis.HuaLu)
	addTarget(huaQuanStar, "化权", &analysis.HuaQuan)
	addTarget(huaKeStar, "化科", &analysis.HuaKe)
	addTarget(huaJiStar, "化忌", &analysis.HuaJi)

	return analysis
}

// flyEffect generates a rule-based description for a flying star.
func flyEffect(huaType, star string, palace basis.Palace) string {
	desc := map[string]string{
		"命宮":  "直接影响个人运势与性格",
		"兄弟宮": "影响兄弟姐妹关系与助力",
		"夫妻宮": "影响婚姻感情与配偶关系",
		"子女宮": "影响子女缘分与下属关系",
		"財帛宮": "影响财运与金钱进出",
		"疾厄宮": "影响身体健康状况",
		"遷移宮": "影响外出运程与社会形象",
		"僕役宮": "影响朋友与部属关系",
		"官祿宮": "影响事业运程与工作成就",
		"田宅宮": "影响房产运程与家庭环境",
		"福德宮": "影响精神享受与内心世界",
		"父母宮": "影响父母缘分与长辈助力",
	}
	pName := palace.String()
	if d, ok := desc[pName]; ok {
		return fmt.Sprintf("%s%s飞入%s，%s", star, huaType, pName, d)
	}
	return fmt.Sprintf("%s%s飞入%s", star, huaType, pName)
}
