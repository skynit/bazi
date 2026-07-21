package ziwei

import (
	"fmt"
	"strings"

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
}

// ZiWeiChart is the output representation of a full ZiWei Dou Shu chart.
type ZiWeiChart struct {
	ProfileID               string                `json:"profile_id"`
	EngineVersion           string                `json:"engine_version"`
	RuleVersion             string                `json:"rule_version"`
	RuleSchool              string                `json:"rule_school"`
	RuleSources             []RuleSourceRef       `json:"rule_sources"`
	RuntimeRuleTablesSchema string                `json:"runtime_rule_tables_schema"`
	RuntimeRuleTablesHash   string                `json:"runtime_rule_tables_hash"`
	PluginManifest          []PluginRequirement   `json:"plugin_manifest"`
	PluginManifestHash      string                `json:"plugin_manifest_hash"`
	CalculationInput        ZiWeiCalculationInput `json:"calculation_input"`
	InputFingerprint        string                `json:"input_fingerprint"`
	ContentHash             string                `json:"content_hash,omitempty"`
	DerivationType          string                `json:"derivation_type,omitempty"`
	DerivationInput         *ZiWeiDerivationInput `json:"derivation_input,omitempty"`
	DerivationFingerprint   string                `json:"derivation_fingerprint,omitempty"`
	BaseContentHash         string                `json:"base_content_hash,omitempty"`
	DerivedContentHash      string                `json:"derived_content_hash,omitempty"`
	Palaces                 [12]PalaceInfo        `json:"palaces"`
	BodyPalace              string                `json:"body_palace"`
	LifeMaster              string                `json:"life_master"`
	BodyMaster              string                `json:"body_master"`
	FiveBureau              string                `json:"five_bureau"`
	Patterns                []string              `json:"patterns"`
	LunarMonth              int                   `json:"lunar_month"`

	EarthlyBranchOfSoulPalace string `json:"earthly_branch_of_soul_palace"` // 命宫地支名
	EarthlyBranchOfBodyPalace string `json:"earthly_branch_of_body_palace"` // 身宫地支名

	SanfangSizheng [12]SanfangSizhengResult `json:"san_fang_si_zheng"`

	LiuNianStars   [12][]string `json:"liu_nian_stars"`
	LiuYueStars    [12][]string `json:"liu_yue_stars"`
	LiuRiStars     [12][]string `json:"liu_ri_stars"`
	LiuNianFourHua [12][]string `json:"liu_nian_four_hua"`
	LiuYueFourHua  [12][]string `json:"liu_yue_four_hua"`
	LiuRiFourHua   [12][]string `json:"liu_ri_four_hua"`
	LiuNianPalaces [12]string   `json:"liu_nian_palaces"`
	LiuYuePalaces  [12]string   `json:"liu_yue_palaces"`
	LiuRiPalaces   [12]string   `json:"liu_ri_palaces"`
}

// ──────────────────── Flying Star Analysis ────────────────────

type FlyingStarAnalysis struct {
	SihuaProjectionSemantics
	AnalysisKind string      `json:"analysis_kind"`
	HuaLu        []FlyTarget `json:"hua_lu"`
	HuaQuan      []FlyTarget `json:"hua_quan"`
	HuaKe        []FlyTarget `json:"hua_ke"`
	HuaJi        []FlyTarget `json:"hua_ji"`
}

type FlyTarget struct {
	SihuaProjectionSemantics
	TransformedStar string `json:"transformed_star"`
	HuaType         string `json:"hua_type"`
	TargetPalace    string `json:"target_palace"`
}

// Dayun is a list of Da Yun stages.
type Dayun []DayunStage

type DayunStage struct {
	StartAge      int      `json:"start_age"`
	EndAge        int      `json:"end_age"`
	Palace        string   `json:"palace"`
	HeavenlyStem  string   `json:"heavenly_stem"`
	EarthlyBranch string   `json:"earthly_branch"`
	GanZhi        string   `json:"gan_zhi"`
	Stars         []string `json:"stars"`
	LiuNianStars  []string `json:"liu_nian_stars,omitempty"`
	LiuYueStars   []string `json:"liu_yue_stars,omitempty"`
}

// ──────────────────── Service ────────────────────

// ZiWeiService provides ZiWei Dou Shu calculation methods.
// No longer depends on ziwei-zenith engine.
type ZiWeiService struct {
	profile       CalculationProfile
	pluginCatalog map[string]StarPlugin
}

func NewZiWeiService() *ZiWeiService {
	profile, _ := ResolveProfile(DefaultProfileID)
	return &ZiWeiService{profile: profile, pluginCatalog: map[string]StarPlugin{}}
}

func NewZiWeiServiceWithProfile(profileID string) (*ZiWeiService, error) {
	profile, err := ResolveProfile(profileID)
	if err != nil {
		return nil, err
	}
	return newZiWeiService(profile, nil)
}

// NewZiWeiServiceWithPlugins injects an immutable service-local plugin catalog.
// Only plugins explicitly listed by the selected profile are executed.
func NewZiWeiServiceWithPlugins(profileID string, plugins ...StarPlugin) (*ZiWeiService, error) {
	profile, err := ResolveProfile(profileID)
	if err != nil {
		return nil, err
	}
	return newZiWeiService(profile, plugins)
}

func newZiWeiService(profile CalculationProfile, plugins []StarPlugin) (*ZiWeiService, error) {
	catalog, err := newPluginCatalog(plugins)
	if err != nil {
		return nil, err
	}
	if _, err := buildPluginExecutionPlan(profile, catalog); err != nil {
		return nil, err
	}
	return &ZiWeiService{profile: profile, pluginCatalog: catalog}, nil
}

// CalculateChart computes a full ZiWei Dou Shu chart from solar birth data.
func (s *ZiWeiService) CalculateChart(year, month, day, hour, minute int, gender string) (*ZiWeiChart, error) {
	profile := s.profile
	if profile.ID == "" {
		profile, _ = ResolveProfile(DefaultProfileID)
	}
	return s.calculateChart(profile, year, month, day, hour, minute, gender)
}

// CalculateChartWithProfile computes a chart under an explicitly selected
// profile. Unknown profile IDs fail before any calculation is attempted.
func (s *ZiWeiService) CalculateChartWithProfile(profileID string, year, month, day, hour, minute int, gender string) (*ZiWeiChart, error) {
	profile, err := ResolveProfile(profileID)
	if err != nil {
		return nil, err
	}
	return s.calculateChart(profile, year, month, day, hour, minute, gender)
}

func (s *ZiWeiService) calculateChart(profile CalculationProfile, year, month, day, hour, minute int, gender string) (*ZiWeiChart, error) {
	if err := validateRuntimeRuleTables(profile); err != nil {
		return nil, fmt.Errorf("validate ziwei runtime rule tables: %w", err)
	}
	plan, err := buildPluginExecutionPlan(profile, s.pluginCatalog)
	if err != nil {
		return nil, fmt.Errorf("build ziwei plugin plan: %w", err)
	}
	birth, err := buildBirthData(year, month, day, hour, minute, gender)
	if err != nil {
		return nil, fmt.Errorf("build birth data: %w", err)
	}

	chart, err := calculateZiWeiChart(birth)
	if err != nil {
		return nil, fmt.Errorf("calculate chart: %w", err)
	}

	plan.Apply(chart)
	normalizeUnsupportedStarBrightness(chart)
	stampChartProfile(chart, profile, plan)
	if err := stampChartCacheContract(chart, birth); err != nil {
		return nil, fmt.Errorf("stamp ziwei cache contract: %w", err)
	}
	return chart, nil
}

// ChartMatchesProfile reports whether a cached chart was produced by the
// exact requested engine, rule version, and school profile.
func (s *ZiWeiService) ChartMatchesProfile(chart *ZiWeiChart, profileID string) bool {
	profile, err := ResolveProfile(profileID)
	if err != nil {
		return false
	}
	return chartMatchesProfile(chart, profile)
}

// ChartMatchesInputProfile verifies cache metadata, normalized birth input,
// and the complete serialized chart content without recalculating the chart.
func (s *ZiWeiService) ChartMatchesInputProfile(chart *ZiWeiChart, profileID string, year, month, day, hour, minute int, gender string) bool {
	profile, err := ResolveProfile(profileID)
	if err != nil || !chartMatchesProfile(chart, profile) {
		return false
	}
	birth, err := buildBirthData(year, month, day, hour, minute, gender)
	if err != nil {
		return false
	}
	wantInput := calculationInputFromBirth(birth)
	return chart.CalculationInput == wantInput &&
		chart.InputFingerprint == ziweiInputFingerprint(wantInput)
}

// AttachBirthData authenticates a cached chart and its exact birth input.
// The historical method name is retained for handler callers; no runtime-only
// birth state is attached to the chart.
func (s *ZiWeiService) AttachBirthData(chart *ZiWeiChart, year, month, day, hour, minute int, gender string) error {
	if s == nil || chart == nil {
		return fmt.Errorf("service or chart is nil")
	}
	profile := s.profile
	if profile.ID == "" {
		var err error
		profile, err = ResolveProfile(DefaultProfileID)
		if err != nil {
			return fmt.Errorf("resolve default profile: %w", err)
		}
	}
	if !chartMatchesProfile(chart, profile) {
		return fmt.Errorf("chart does not match service profile")
	}
	birth, err := buildBirthData(year, month, day, hour, minute, gender)
	if err != nil {
		return fmt.Errorf("build birth data: %w", err)
	}
	if calculationInputFromBirth(birth) != chart.CalculationInput {
		return fmt.Errorf("birth input does not match published chart")
	}
	_, ok := BranchIndex[chart.EarthlyBranchOfSoulPalace]
	if !ok {
		return fmt.Errorf("invalid published soul-palace branch %q", chart.EarthlyBranchOfSoulPalace)
	}
	_, ok = BranchIndex[chart.EarthlyBranchOfBodyPalace]
	if !ok {
		return fmt.Errorf("invalid published body-palace branch %q", chart.EarthlyBranchOfBodyPalace)
	}

	return nil
}

// DetectLocalPatterns detects patterns using the local knowledge layer.
func (s *ZiWeiService) DetectLocalPatterns(chart *ZiWeiChart) []string {
	if !s.acceptsPublishedNatalChart(chart) {
		return nil
	}
	return DetectLocalPatterns(chart)
}

// AnalyzeFlyingStars performs 四化飞星 analysis on the chart.
func (s *ZiWeiService) AnalyzeFlyingStars(chart *ZiWeiChart) *FlyingStarAnalysis {
	if !s.acceptsPublishedNatalChart(chart) {
		return nil
	}
	return buildFlyingStarAnalysisFromChart(chart)
}

// CalculateDayun computes the 大限 (10-year luck periods).
func (s *ZiWeiService) CalculateDayun(chart *ZiWeiChart) Dayun {
	if !s.acceptsPublishedNatalChart(chart) {
		return nil
	}
	return calcDayunFromChart(chart)
}

// CalculateLiunian computes the 流年 overlay for a given year.
func (s *ZiWeiService) CalculateLiunian(chart *ZiWeiChart, targetYear int) *ZiWeiChart {
	input, err := buildZiWeiDerivationInput("liunian", targetYear, 0, 0)
	if s == nil || chart == nil || err != nil || !chartMatchesProfile(chart, s.profile) {
		return nil
	}
	palaces, ok := buildTransitPalaceNames(chart, input, "liunian")
	if !ok {
		return nil
	}
	result := *chart

	liunianStem, liunianBranch, ok := derivationStemBranch(input)
	if !ok {
		return nil
	}

	result.LiuNianStars = buildTransitStarDistribution(&result, liunianStem, liunianBranch, "liunian")
	result.LiuNianFourHua = buildTransitFourHua(&result, liunianStem)
	result.LiuNianPalaces = palaces
	if err = stampDerivedChartContract(&result, chart, "liunian", input); err != nil {
		return nil
	}

	return &result
}

// CalculateLiuyueForDate computes the 流月 overlay containing the target
// solar date. The selected Profile uses lunar-month boundaries.
func (s *ZiWeiService) CalculateLiuyueForDate(chart *ZiWeiChart, year, month, day int) *ZiWeiChart {
	input, err := buildZiWeiDerivationInput("liuyue", year, month, day)
	if s == nil || chart == nil || err != nil || !chartMatchesProfile(chart, s.profile) {
		return nil
	}
	palaces, ok := buildTransitPalaceNames(chart, input, "liuyue")
	if !ok {
		return nil
	}
	result := *chart

	liuyueStem, liuyueBranch, ok := derivationStemBranch(input)
	if !ok {
		return nil
	}
	result.LiuYueStars = buildTransitStarDistribution(&result, liuyueStem, liuyueBranch, "liuyue")
	result.LiuYueFourHua = buildTransitFourHua(&result, liuyueStem)
	result.LiuYuePalaces = palaces
	if err = stampDerivedChartContract(&result, chart, "liuyue", input); err != nil {
		return nil
	}

	return &result
}

// CalculateLiuriForDate computes the 流日 overlay for a target solar date.
func (s *ZiWeiService) CalculateLiuriForDate(chart *ZiWeiChart, year, month, day int) *ZiWeiChart {
	input, err := buildZiWeiDerivationInput("liuri", year, month, day)
	if s == nil || chart == nil || err != nil || !chartMatchesProfile(chart, s.profile) {
		return nil
	}
	palaces, ok := buildTransitPalaceNames(chart, input, "liuri")
	if !ok {
		return nil
	}
	result := *chart

	liuriStem, liuriBranch, ok := derivationStemBranch(input)
	if !ok {
		return nil
	}
	result.LiuRiStars = buildTransitStarDistribution(&result, liuriStem, liuriBranch, "liuri")
	result.LiuRiFourHua = buildTransitFourHua(&result, liuriStem)
	result.LiuRiPalaces = palaces
	if err = stampDerivedChartContract(&result, chart, "liuri", input); err != nil {
		return nil
	}

	return &result
}

func normalizeUnsupportedStarBrightness(chart *ZiWeiChart) {
	if chart == nil {
		return
	}
	unsupported := map[string]bool{
		"左辅": true, "右弼": true, "天魁": true, "天钺": true,
		"禄存": true, "天马": true, "地空": true, "地劫": true,
	}
	for i := range chart.Palaces {
		p := &chart.Palaces[i]
		for j := range p.Stars {
			star := &p.Stars[j]
			if !unsupported[star.Name] {
				continue
			}
			star.Brightness = ""
		}
	}
}

// AnalyzeSihuaChain returns direct palace-stem four-hua flight edges. The
// method name is retained for the historical period_type enum.
func (s *ZiWeiService) AnalyzeSihuaChain(chart *ZiWeiChart) *SihuaChainResult {
	if !s.acceptsPublishedNatalChart(chart) {
		return nil
	}
	return analyzeSihuaChain(chart)
}

// GetPalaceReading returns a reading only for a chart authenticated against
// this service's complete published natal-chart profile.
func (s *ZiWeiService) GetPalaceReading(chart *ZiWeiChart, palaceIdx int) *PalaceReading {
	if !s.acceptsPublishedNatalChart(chart) {
		return nil
	}
	return buildPalaceReading(chart, palaceIdx)
}

// BuildQueryView returns a precomputed query surface for an authenticated
// natal or derived chart under this service's profile.
func (s *ZiWeiService) BuildQueryView(chart *ZiWeiChart) *QueryView {
	if !s.acceptsPublishedChart(chart) {
		return nil
	}
	return buildQueryView(chart)
}

// AnalyzeHeming compares structural projections from two authenticated natal
// charts without producing a compatibility result.
func (s *ZiWeiService) AnalyzeHeming(chartA, chartB *ZiWeiChart) *HemingResult {
	if !s.acceptsPublishedNatalChart(chartA) || !s.acceptsPublishedNatalChart(chartB) {
		return nil
	}
	return analyzeHeming(chartA, chartB)
}

// AnalyzeSelfMutagen detects self-mutagen occurrences.
func (s *ZiWeiService) AnalyzeSelfMutagen(chart *ZiWeiChart) []SelfMutagenResult {
	if !s.acceptsPublishedNatalChart(chart) {
		return nil
	}
	return detectSelfMutagens(chart)
}

func (s *ZiWeiService) acceptsPublishedNatalChart(chart *ZiWeiChart) bool {
	return s != nil && chartMatchesProfile(chart, s.profile)
}

func (s *ZiWeiService) acceptsPublishedChart(chart *ZiWeiChart) bool {
	if s == nil || chart == nil {
		return false
	}
	if chart.DerivationType == "" {
		return chartMatchesProfile(chart, s.profile)
	}
	return chart.ProfileID == s.profile.ID && ValidDerivedChartContract(chart)
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

	yearStem, yearBranch, err := parseGanZhiName(yearPillarName)
	if err != nil {
		return nil, fmt.Errorf("invalid lunar year pillar: %w", err)
	}
	_, monthBranch, err := parseGanZhiName(monthPillarName)
	if err != nil {
		return nil, fmt.Errorf("invalid month pillar: %w", err)
	}
	dayStem, dayBranch, err := parseGanZhiName(dayPillarName)
	if err != nil {
		return nil, fmt.Errorf("invalid day pillar: %w", err)
	}
	hourStem, parsedHourBranch, err := parseGanZhiName(hourPillarName)
	if err != nil {
		return nil, fmt.Errorf("invalid hour pillar: %w", err)
	}

	// Determine the hour branch from the input hour (not from the pillar)
	// 子时 starts at 23:00, so hour 23-1 maps to 子(0), hour 1-3 maps to 丑(1), etc.
	hourBranch := ((hour + 1) / 2) % 12
	if parsedHourBranch != hourBranch {
		return nil, fmt.Errorf("hour pillar branch %s does not match input hour branch %s", BranchNames[parsedHourBranch], BranchNames[hourBranch])
	}

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
		DayStem:           dayStem,
		DayBranch:         dayBranch,
		HourBranch:        hourBranch,
		HourStem:          hourStem,
		IsLeapMonth:       isLeapMonth,
	}, nil
}

func parseGanZhiName(name string) (stem, branch int, err error) {
	runes := []rune(strings.TrimSpace(name))
	if len(runes) != 2 {
		return 0, 0, fmt.Errorf("ganzhi must contain exactly two characters, got %q", name)
	}
	stem, ok := StemIndex[string(runes[0])]
	if !ok {
		return 0, 0, fmt.Errorf("unknown heavenly stem %q", string(runes[0]))
	}
	branch, ok = BranchIndex[string(runes[1])]
	if !ok {
		return 0, 0, fmt.Errorf("unknown earthly branch %q", string(runes[1]))
	}
	if stem%2 != branch%2 {
		return 0, 0, fmt.Errorf("stem-branch parity mismatch in %q", name)
	}
	return stem, branch, nil
}

// ──────────────────── Flying Star Analysis (new engine) ────────────────────

func buildFlyingStarAnalysisFromChart(chart *ZiWeiChart) *FlyingStarAnalysis {
	if chart == nil {
		return nil
	}

	analysis := &FlyingStarAnalysis{
		SihuaProjectionSemantics: sihuaProjectionSemantics(),
		AnalysisKind:             "natal_year_stem_four_hua_projection",
	}
	for _, palace := range chart.Palaces {
		for _, transformed := range palace.FourHua {
			starName, huaType, ok := parseFourHuaLabel(transformed)
			if !ok {
				continue
			}

			target := FlyTarget{
				SihuaProjectionSemantics: sihuaProjectionSemantics(),
				TransformedStar:          starName,
				HuaType:                  huaType,
				TargetPalace:             palace.Name,
			}
			switch huaType {
			case "化禄":
				analysis.HuaLu = append(analysis.HuaLu, target)
			case "化权":
				analysis.HuaQuan = append(analysis.HuaQuan, target)
			case "化科":
				analysis.HuaKe = append(analysis.HuaKe, target)
			case "化忌":
				analysis.HuaJi = append(analysis.HuaJi, target)
			}
		}
	}

	return analysis
}

func parseFourHuaLabel(label string) (starName, huaType string, ok bool) {
	label = strings.TrimSpace(label)
	for _, candidate := range SiHuaLabels {
		if strings.HasSuffix(label, candidate) && len(label) > len(candidate) {
			return strings.TrimSpace(strings.TrimSuffix(label, candidate)), candidate, true
		}
	}
	return "", "", false
}

// ──────────────────── Dayun (大限) ────────────────────

type dayunParameters struct {
	yearStem   int
	gender     string
	juValue    int
	soulBranch int
}

func dayunParametersFromChart(chart *ZiWeiChart) (dayunParameters, bool) {
	birth, ok := birthDataFromPublishedChart(chart)
	if !ok {
		return dayunParameters{}, false
	}
	juValue, ok := FiveBureauValue[chart.FiveBureau]
	if !ok {
		return dayunParameters{}, false
	}
	soulBranch, ok := BranchIndex[chart.EarthlyBranchOfSoulPalace]
	if !ok {
		return dayunParameters{}, false
	}
	return dayunParameters{
		yearStem:   birth.YearStem,
		gender:     birth.Gender,
		juValue:    juValue,
		soulBranch: soulBranch,
	}, true
}

func calcDayunFromChart(chart *ZiWeiChart) Dayun {
	params, ok := dayunParametersFromChart(chart)
	if !ok {
		return nil
	}
	forward := isForwardByYearStem(params.yearStem, params.gender)

	// 大限起运年龄: 五行局值
	startAge := params.juValue

	// 大限宫位从命宫(或其相邻宫)开始
	var result Dayun
	soulBranch := params.soulBranch
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
		if palaceIdx < 0 || chart.Palaces[palaceIdx].Name == "" {
			return nil
		}

		stage := DayunStage{
			StartAge:      startAge + i*10,
			EndAge:        startAge + (i+1)*10 - 1,
			Palace:        chart.Palaces[palaceIdx].Name,
			HeavenlyStem:  chart.Palaces[palaceIdx].HeavenlyStem,
			EarthlyBranch: chart.Palaces[palaceIdx].Branch,
			GanZhi:        chart.Palaces[palaceIdx].HeavenlyStem + chart.Palaces[palaceIdx].Branch,
		}
		for _, star := range chart.Palaces[palaceIdx].Stars {
			if star.Name != "" {
				stage.Stars = append(stage.Stars, star.Name)
			}
		}
		result = append(result, stage)
	}
	return result
}
