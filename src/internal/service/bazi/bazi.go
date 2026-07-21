package bazi

import (
	"fmt"
	"math"

	"bazi/internal/model"
	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

// TenGodRatio holds one ten-god's count and percentage in a birth chart.
type TenGodRatio struct {
	Name    string  `json:"name"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

// tenGodNames is the fixed ordering for proportion output.
var tenGodNames = [10]string{
	"比肩", "劫财", "食神", "伤官",
	"正财", "偏财", "正官", "七杀",
	"正印", "偏印",
}

// BaziService calculates BaZi (八字) birth charts using tyme4go.
type BaziService struct{}

// BaziResult holds the complete BaZi calculation output.
type BaziResult struct {
	RuleVersion          string                      `json:"rule_version"`
	CalendarVersion      string                      `json:"calendar_version"`
	School               string                      `json:"school"`
	ZiHourPolicy         string                      `json:"zi_hour_policy,omitempty"`
	RuleMeta             RuleMeta                    `json:"rule_meta"`
	YearPillar           model.Pillar                `json:"year_pillar"`
	MonthPillar          model.Pillar                `json:"month_pillar"`
	DayPillar            model.Pillar                `json:"day_pillar"`
	HourPillar           model.Pillar                `json:"hour_pillar"`
	FiveElements         map[string]int              `json:"five_elements"`
	ElementDetail        []ElementStrength           `json:"element_detail"`
	BodyStrength         BodyStrengthResult          `json:"body_strength"`
	TenGods              map[string]string           `json:"ten_gods"`
	NaYin                map[string]NaYinInfo        `json:"na_yin"`
	HiddenStems          map[string][]string         `json:"hidden_stems"`
	DaYunInfo            DaYunInfo                   `json:"da_yun_info"`
	GanZhiAnalysis       GanZhiAnalysis              `json:"gan_zhi_analysis"`
	PatternAnalysis      PatternAnalysis             `json:"pattern_analysis"`
	MingGong             data.MingGongDetail         `json:"ming_gong"`
	PillarDetails        []PillarDetail              `json:"pillar_details"`
	DayShenSha           []string                    `json:"day_shen_sha"`
	DayShenShaDetails    []ShenShaMeta               `json:"day_shen_sha_details"`
	ShenShaByPillar      []PillarShenSha             `json:"shen_sha_by_pillar"`
	GlobalShenSha        []string                    `json:"global_shen_sha"`
	GlobalShenShaDetails []ShenShaMeta               `json:"global_shen_sha_details"`
	MonthSeason          MonthSeasonEvidence         `json:"month_season"`
	TenGodProportion     []TenGodRatio               `json:"ten_god_proportion"`
	TenGodAnalysis       *TenGodAnalysis             `json:"ten_god_analysis"`
	Tiaohou              *TiaohouResult              `json:"tiaohou"`          // 《穷通宝鉴》调候查表证据
	MissingElements      data.MissingElementAnalysis `json:"missing_elements"` // 缺失五行
}

// ElementStrength holds the strength breakdown for one element.
type ElementStrength struct {
	Element     string   `json:"element"`
	TianGan     int      `json:"tian_gan"`
	ZhiCangGan  int      `json:"zhi_cang_gan"`
	Total       int      `json:"total"`
	CangGanList []string `json:"cang_gan_list"` // e.g. ["未中乙", "辰中乙"]
}

// BodyStrengthResult exposes a local weighted-score observation. Its discrete
// band is a threshold candidate, not an adjudicated strength conclusion.
type BodyStrengthResult struct {
	RuleID               string                    `json:"rule_id"`
	SchemaVersion        string                    `json:"schema_version"`
	RuleVersion          string                    `json:"rule_version"`
	School               string                    `json:"school"`
	ScoringProfile       string                    `json:"scoring_profile"`
	YueLingRuleID        string                    `json:"yue_ling_rule_id"`
	YueLingProfile       string                    `json:"yue_ling_profile"`
	YueLingTableSHA256   string                    `json:"yue_ling_table_sha256"`
	Inputs               BodyStrengthInputSnapshot `json:"inputs"`
	ScoreBandCandidate   string                    `json:"score_band_candidate"`
	BandSelectionBasis   string                    `json:"band_selection_basis"`
	BandRules            []BodyStrengthBandRule    `json:"band_rules"`
	TotalScore           float64                   `json:"total_score"`
	LingScore            float64                   `json:"ling_score"`
	DiScore              float64                   `json:"di_score"`
	ShiScore             float64                   `json:"shi_score"`
	ShengScore           float64                   `json:"sheng_score"`
	LuBonus              float64                   `json:"lu_bonus"`
	Components           []BodyStrengthComponent   `json:"components"`
	Evidence             []BodyStrengthEvidence    `json:"evidence"`
	Adjustments          []BodyStrengthAdjustment  `json:"adjustments"`
	Status               string                    `json:"status"`
	ValidationStatus     string                    `json:"validation_status"`
	InterpretationStatus string                    `json:"interpretation_status"`
	IsStrengthConclusion bool                      `json:"is_strength_conclusion"`
	Limitations          []string                  `json:"limitations"`
}

type BodyStrengthInputSnapshot struct {
	Pillars     []string `json:"pillars"`
	DayStem     string   `json:"day_stem"`
	DayElement  string   `json:"day_element"`
	MonthBranch string   `json:"month_branch"`
}

type BodyStrengthBandRule struct {
	Candidate string  `json:"candidate"`
	Operator  string  `json:"operator"`
	Threshold float64 `json:"threshold,omitempty"`
}

// BodyStrengthComponent is one weighted part of the strength score.
type BodyStrengthComponent struct {
	RuleID           string  `json:"rule_id"`
	Key              string  `json:"key"`
	Name             string  `json:"name"`
	RawScore         float64 `json:"raw_score"`
	NormalizedScore  float64 `json:"normalized_score"`
	Weight           float64 `json:"weight"`
	WeightedScore    float64 `json:"weighted_score"`
	Basis            string  `json:"basis"`
	Status           string  `json:"status"`
	ValidationStatus string  `json:"validation_status"`
	Description      string  `json:"description"`
}

// BodyStrengthEvidence is a computable fact that affected the verdict.
type BodyStrengthEvidence struct {
	RuleID               string  `json:"rule_id"`
	Component            string  `json:"component"`
	Polarity             string  `json:"polarity"`
	Source               string  `json:"source"`
	Item                 string  `json:"item"`
	Score                float64 `json:"score"`
	Basis                string  `json:"basis"`
	Status               string  `json:"status"`
	InterpretationStatus string  `json:"interpretation_status"`
	Reason               string  `json:"reason"`
}

// BodyStrengthAdjustment records a posterior score correction.
type BodyStrengthAdjustment struct {
	RuleID           string  `json:"rule_id"`
	Name             string  `json:"name"`
	Before           float64 `json:"before"`
	After            float64 `json:"after"`
	Reason           string  `json:"reason"`
	Basis            string  `json:"basis"`
	Status           string  `json:"status"`
	ValidationStatus string  `json:"validation_status"`
	Description      string  `json:"description"`
}

// DaYunInfo describes the major fortune cycle (大运).
type DaYunInfo struct {
	Calculated            bool            `json:"calculated"`
	StartAge              int             `json:"start_age"`
	StartAgeDetail        DaYunStartAge   `json:"start_age_detail"`
	StartAt               string          `json:"start_at,omitempty"`
	Direction             string          `json:"direction"`
	DirectionBasis        string          `json:"direction_basis,omitempty"`
	CalculationProfile    string          `json:"calculation_profile,omitempty"`
	Provider              string          `json:"provider,omitempty"`
	TimeBasis             string          `json:"time_basis,omitempty"`
	SolarTermReferenceAt  string          `json:"solar_term_reference_at,omitempty"`
	SolarTermTimezone     string          `json:"solar_term_timezone,omitempty"`
	AgeConversionRule     string          `json:"age_conversion_rule,omitempty"`
	BoundaryRule          string          `json:"boundary_rule,omitempty"`
	PreviousJie           *DaYunSolarTerm `json:"previous_jie,omitempty"`
	NextJie               *DaYunSolarTerm `json:"next_jie,omitempty"`
	ReferenceJie          *DaYunSolarTerm `json:"reference_jie,omitempty"`
	ReferenceDeltaSeconds int             `json:"reference_delta_seconds"`
	Pillars               []model.Pillar  `json:"pillars"`
}

// DaYunStartAge keeps the full traditional conversion result. StartAge on
// DaYunInfo remains the whole-year component used by downstream fortune code.
type DaYunStartAge struct {
	Years   int `json:"years"`
	Months  int `json:"months"`
	Days    int `json:"days"`
	Hours   int `json:"hours"`
	Minutes int `json:"minutes"`
}

// DaYunSolarTerm records one solar-month boundary in China Standard Time.
// DeltaSeconds is signed against the same physical birth instant: past/current
// boundaries are <= 0 and future boundaries are > 0.
type DaYunSolarTerm struct {
	Name         string `json:"name"`
	At           string `json:"at"`
	DeltaSeconds int    `json:"delta_seconds"`
}

// PillarShenSha groups shen-sha items for a single pillar with metadata.
type PillarShenSha struct {
	Pillar  string        `json:"pillar"`
	Label   string        `json:"label"`
	Gan     string        `json:"gan"`
	Zhi     string        `json:"zhi"`
	Items   []string      `json:"items"`
	Details []ShenShaMeta `json:"details"`
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

// PillarDetail holds enriched per-pillar data.
type PillarDetail struct {
	Stem      string    `json:"stem"`
	Branch    string    `json:"branch"`
	ShengXiao string    `json:"sheng_xiao"`
	Empties   [2]string `json:"empties"`
	Nayin     NaYinInfo `json:"nayin"`
}

// Calculate computes a full BaZi chart.
func (s *BaziService) Calculate(year, month, day, hour, minute int, gender string) (*BaziResult, error) {
	return s.CalculateAt(year, month, day, hour, minute, 0, gender)
}

// CalculateAt computes a full BaZi chart while preserving the factual birth
// second for solar-term boundaries and date-level da-yun calculation.
func (s *BaziService) CalculateAt(year, month, day, hour, minute, second int, gender string) (*BaziResult, error) {
	return s.CalculateAtWithPolicy(year, month, day, hour, minute, second, gender, DefaultZiHourPolicy)
}

// CalculateAtWithPolicy computes a chart using an explicit late-Zi day-boundary
// convention without mutating tyme4go's process-global EightCharProvider.
func (s *BaziService) CalculateAtWithPolicy(year, month, day, hour, minute, second int, gender, ziHourPolicy string) (*BaziResult, error) {
	st, err := tyme.SolarTime{}.FromYmdHms(year, month, day, hour, minute, second)
	if err != nil {
		return nil, fmt.Errorf("invalid birth time: %w", err)
	}
	return s.calculateWithTimeBases(st, st, gender, ziHourPolicy)
}

// CalculateNormalizedBirth keeps local day/hour conventions separate from the
// globally instantaneous solar-term year/month boundary.
func (s *BaziService) CalculateNormalizedBirth(birth *NormalizedBirth) (*BaziResult, error) {
	if birth == nil {
		return nil, fmt.Errorf("normalized birth is required")
	}
	calculationTime, err := tyme.SolarTime{}.FromYmdHms(
		birth.Year, birth.Month, birth.Day, birth.Hour, birth.Minute, birth.Second,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid normalized calculation time: %w", err)
	}
	reference := birth.SolarTermReference
	termReferenceTime, err := tyme.SolarTime{}.FromYmdHms(
		reference.Year, reference.Month, reference.Day, reference.Hour, reference.Minute, reference.Second,
	)
	if err != nil {
		return nil, fmt.Errorf("invalid normalized solar-term reference time: %w", err)
	}
	if birth.Validation.CalendarEngineVersion != CalendarEngineVersion ||
		birth.Validation.NormalizationVersion != BirthNormalizationVersion {
		return nil, fmt.Errorf("normalized birth calendar version is stale")
	}
	return s.calculateWithTimeBases(calculationTime, termReferenceTime, birth.Gender, birth.ZiHourPolicy)
}

func (s *BaziService) calculateWithTimeBases(calculationTime, termReferenceTime *tyme.SolarTime, gender, ziHourPolicy string) (*BaziResult, error) {
	tymeGender, err := toTymeGender(gender)
	if err != nil {
		return nil, err
	}

	ziHourPolicy, err = NormalizeZiHourPolicy(ziHourPolicy)
	if err != nil {
		return nil, err
	}
	localEightChar, err := eightCharWithZiHourPolicy(calculationTime, ziHourPolicy)
	if err != nil {
		return nil, err
	}
	termEightChar, err := eightCharWithZiHourPolicy(termReferenceTime, DefaultZiHourPolicy)
	if err != nil {
		return nil, err
	}
	ec := tyme.EightChar{}.FromSixtyCycle(
		termEightChar.GetYear(), termEightChar.GetMonth(), localEightChar.GetDay(), localEightChar.GetHour(),
	)

	result := &BaziResult{
		RuleVersion:     RuleVersion,
		CalendarVersion: CalendarEngineVersion,
		School:          RuleSchool,
		ZiHourPolicy:    ziHourPolicy,
		RuleMeta:        DefaultRuleMeta(),
	}

	// --- four pillars ---
	result.YearPillar = pillarFromSixtyCycle(ec.GetYear())
	result.MonthPillar = pillarFromSixtyCycle(ec.GetMonth())
	result.DayPillar = pillarFromSixtyCycle(ec.GetDay())
	result.HourPillar = pillarFromSixtyCycle(ec.GetHour())
	result.MonthSeason = observeMonthSeason(result.MonthPillar.Zhi)

	// MingGong from year gan, month zhi, hour zhi (《渊海子平》古法)
	mingGongGanZhi, err := calcMingGongGanZhi(result.YearPillar.Gan, result.MonthPillar.Zhi, result.HourPillar.Zhi)
	if err != nil {
		return nil, fmt.Errorf("计算命宫失败: %w", err)
	}
	result.MingGong = buildMingGongDetail(mingGongGanZhi)
	// --- five elements scores ---
	result.FiveElements = calcFiveElements(&ec)
	result.ElementDetail = calcElementDetail(&ec)
	result.BodyStrength = calcBodyStrength(&ec)

	// --- ten gods ---
	result.TenGods = calcTenGods(&ec)

	// --- na yin ---
	result.NaYin = calcNaYin(&ec)

	// --- hidden stems ---
	result.HiddenStems = calcHiddenStems(&ec)

	// --- da yun ---
	result.DaYunInfo = calcDaYunWithReference(calculationTime, termReferenceTime, ec, tymeGender)

	// --- gan/zhi analysis ---
	result.GanZhiAnalysis, err = CalcGanZhiAnalysis(
		result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar,
	)
	if err != nil {
		return nil, fmt.Errorf("计算干支关系失败: %w", err)
	}
	result.PatternAnalysis = AnalyzePatternExtended(
		[]model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar},
		result.MonthPillar.Zhi,
	)

	// --- enrich pillar details ---
	result.TenGodProportion = calcTenGodProportion(&ec, result.DayPillar.Gan)
	result.TenGodAnalysis = ObserveTenGodDistribution(result.TenGodProportion)
	if err := s.enrichPillarDetails(result, gender); err != nil {
		return nil, fmt.Errorf("计算柱位神煞失败: %w", err)
	}

	// --- enrich from 《穷通宝鉴》tiaohou analysis ---
	tiaohouResult, _ := AnalyzeTiaohouForPillarsAt(
		result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar, *termReferenceTime,
	)
	result.Tiaohou = tiaohouResult

	// --- retain raw five-element presence facts ---
	result.MissingElements = data.FindMissingElements(result.FiveElements)

	return result, nil
}

func (s *BaziService) enrichPillarDetails(result *BaziResult, gender string) error {
	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}
	result.PillarDetails = buildPillarDetails(pillars)

	shenSha, err := CalcShenShaByPillars(ShenShaPillars{
		Year:   result.YearPillar,
		Month:  result.MonthPillar,
		Day:    result.DayPillar,
		Hour:   result.HourPillar,
		Gender: gender,
	})
	if err != nil {
		return err
	}
	result.DayShenSha = shenSha.Day
	result.DayShenShaDetails = BuildShenShaDetails(shenSha.Day)
	result.ShenShaByPillar = buildPillarShenSha(result, shenSha)
	result.GlobalShenSha = shenSha.Global
	result.GlobalShenShaDetails = BuildShenShaDetails(shenSha.Global)

	return nil
}

func buildPillarDetails(pillars []model.Pillar) []PillarDetail {
	details := make([]PillarDetail, 0, len(pillars))
	for _, p := range pillars {
		detail := PillarDetail{
			Stem:   p.Gan,
			Branch: p.Zhi,
			Nayin:  observeNaYin(p.Gan, p.Zhi),
		}
		cycle, err := tyme.SixtyCycle{}.FromName(p.Gan + p.Zhi)
		if err == nil {
			detail.ShengXiao = cycle.GetEarthBranch().GetZodiac().GetName()
			empties := cycle.GetExtraEarthBranches()
			detail.Empties = [2]string{empties[0].GetName(), empties[1].GetName()}
		}
		details = append(details, detail)
	}
	return details
}

func buildPillarShenSha(result *BaziResult, calc ShenShaCalcResult) []PillarShenSha {
	pillars := []struct {
		pillar string
		label  string
		gan    string
		zhi    string
		items  []string
	}{
		{"year", "年柱", result.YearPillar.Gan, result.YearPillar.Zhi, calc.Year},
		{"month", "月柱", result.MonthPillar.Gan, result.MonthPillar.Zhi, calc.Month},
		{"day", "日柱", result.DayPillar.Gan, result.DayPillar.Zhi, calc.Day},
		{"hour", "时柱", result.HourPillar.Gan, result.HourPillar.Zhi, calc.Hour},
	}

	output := make([]PillarShenSha, 0, len(pillars))
	for _, p := range pillars {
		output = append(output, PillarShenSha{
			Pillar:  p.pillar,
			Label:   p.label,
			Gan:     p.gan,
			Zhi:     p.zhi,
			Items:   p.items,
			Details: BuildShenShaDetails(p.items),
		})
	}
	return output
}

// --- helpers ---------------------------------------------------------------

func pillarFromSixtyCycle(sc tyme.SixtyCycle) model.Pillar {
	return model.Pillar{
		Gan: sc.GetHeavenStem().GetName(),
		Zhi: sc.GetEarthBranch().GetName(),
	}
}

func calcFiveElements(ec *tyme.EightChar) map[string]int {
	scores := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}

	pillars := [](func() tyme.SixtyCycle){
		ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour,
	}
	for _, fn := range pillars {
		sc := fn()
		// heavenly stem: 5 points
		elem := sc.GetHeavenStem().GetElement().GetName()
		scores[elem] += 5

		// earthly branch hidden stems: main=3, middle=2, residual=1
		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			weight := 1
			switch hhs.GetType() {
			case tyme.MAIN:
				weight = 3
			case tyme.MIDDLE:
				weight = 2
			}
			elem := hhs.GetHeavenStem().GetElement().GetName()
			scores[elem] += weight
		}
	}
	return scores
}

func calcElementDetail(ec *tyme.EightChar) []ElementStrength {
	elements := []string{"木", "火", "土", "金", "水"}
	tianGan := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}
	zhiCangGan := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}
	cangGanMap := map[string][]string{"木": {}, "火": {}, "土": {}, "金": {}, "水": {}}

	pillars := [](func() tyme.SixtyCycle){
		ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour,
	}
	for _, fn := range pillars {
		sc := fn()
		elem := sc.GetHeavenStem().GetElement().GetName()
		tianGan[elem] += 5

		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			weight := 1
			label := sc.GetEarthBranch().GetName() + hhs.GetHeavenStem().GetName()
			if hhs.GetType() == tyme.MAIN {
				weight = 3
			} else if hhs.GetType() == tyme.MIDDLE {
				weight = 2
				label += "(中)"
			} else if hhs.GetType() == tyme.RESIDUAL {
				label += "(余)"
			}
			elem := hhs.GetHeavenStem().GetElement().GetName()
			zhiCangGan[elem] += weight
			cangGanMap[elem] = append(cangGanMap[elem], label)
		}
	}

	var result []ElementStrength
	for _, e := range elements {
		result = append(result, ElementStrength{
			Element:     e,
			TianGan:     tianGan[e],
			ZhiCangGan:  zhiCangGan[e],
			Total:       tianGan[e] + zhiCangGan[e],
			CangGanList: cangGanMap[e],
		})
	}
	return result
}

func calcBodyStrength(ec *tyme.EightChar) BodyStrengthResult {
	return calcBodyStrengthV2(ec)
}

func calcTenGods(ec *tyme.EightChar) map[string]string {
	dayStem := ec.GetDay().GetHeavenStem()
	return map[string]string{
		"year":  dayStem.GetTenStar(ec.GetYear().GetHeavenStem()).GetName(),
		"month": dayStem.GetTenStar(ec.GetMonth().GetHeavenStem()).GetName(),
		"day":   "日主",
		"hour":  dayStem.GetTenStar(ec.GetHour().GetHeavenStem()).GetName(),
	}
}

func calcNaYin(ec *tyme.EightChar) map[string]NaYinInfo {
	pillars := []struct {
		key string
		fn  func() tyme.SixtyCycle
	}{
		{"year", ec.GetYear}, {"month", ec.GetMonth},
		{"day", ec.GetDay}, {"hour", ec.GetHour},
	}
	result := make(map[string]NaYinInfo, 4)
	for _, p := range pillars {
		cycle := p.fn()
		result[p.key] = observeNaYin(
			cycle.GetHeavenStem().GetName(),
			cycle.GetEarthBranch().GetName(),
		)
	}
	return result
}

func calcHiddenStems(ec *tyme.EightChar) map[string][]string {
	result := make(map[string][]string, 4)
	pillars := map[string]func() tyme.SixtyCycle{
		"year": ec.GetYear, "month": ec.GetMonth,
		"day": ec.GetDay, "hour": ec.GetHour,
	}
	for name, fn := range pillars {
		sc := fn()
		var stems []string
		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			label := sc.GetEarthBranch().GetName() + hhs.GetHeavenStem().GetName()
			switch hhs.GetType() {
			case tyme.MAIN:
				// keep as-is
			case tyme.MIDDLE:
				label += "(中)"
			case tyme.RESIDUAL:
				label += "(余)"
			}
			stems = append(stems, label)
		}
		result[name] = stems
	}
	return result
}

func calcDaYun(st *tyme.SolarTime, gender tyme.Gender) DaYunInfo {
	eightChar := st.GetLunarHour().GetEightChar()
	return calcDaYunWithReference(st, st, eightChar, gender)
}

func calcDaYunWithReference(calculationBirth, termReferenceBirth *tyme.SolarTime, eightChar tyme.EightChar, gender tyme.Gender) DaYunInfo {
	yearStem := eightChar.GetYear().GetHeavenStem()
	yang := tyme.YANG == yearStem.GetYinYang()
	man := tyme.MAN == gender
	forward := (yang && man) || (!yang && !man)
	dir := "逆行"
	if forward {
		dir = "顺行"
	}
	previousJie, nextJie := surroundingJie(*termReferenceBirth)
	referenceJie := previousJie
	if forward {
		referenceJie = nextJie
	}
	childLimit := (tyme.DefaultChildLimitProvider{}).GetInfo(*termReferenceBirth, referenceJie)
	previousEvidence := daYunTermEvidence(previousJie, *termReferenceBirth)
	nextEvidence := daYunTermEvidence(nextJie, *termReferenceBirth)
	referenceEvidence := previousEvidence
	if forward {
		referenceEvidence = nextEvidence
	}
	referenceDeltaSeconds := referenceEvidence.DeltaSeconds
	if referenceDeltaSeconds < 0 {
		referenceDeltaSeconds = -referenceDeltaSeconds
	}
	profile, conversionRule := defaultChildLimitProfile()
	yinYang := yearStem.GetYinYang().GetName()
	genderName := gender.GetName()
	directionClass := yinYang + genderName

	daYun := DaYunInfo{
		Calculated: true,
		StartAge:   childLimit.GetYearCount(),
		StartAgeDetail: DaYunStartAge{
			Years:   childLimit.GetYearCount(),
			Months:  childLimit.GetMonthCount(),
			Days:    childLimit.GetDayCount(),
			Hours:   childLimit.GetHourCount(),
			Minutes: childLimit.GetMinuteCount(),
		},
		StartAt: formatSolarTime(addChildLimitToBirth(
			*calculationBirth,
			childLimit.GetYearCount(), childLimit.GetMonthCount(), childLimit.GetDayCount(),
			childLimit.GetHourCount(), childLimit.GetMinuteCount(),
		)),
		Direction:             dir,
		DirectionBasis:        fmt.Sprintf("年干%s属%s，%s命；按‘阳男阴女顺、阴男阳女逆’判为%s%s。", yearStem.GetName(), yinYang, genderName, directionClass, dir),
		CalculationProfile:    profile,
		Provider:              "tyme.DefaultChildLimitProvider",
		TimeBasis:             "日时柱与起运落点使用归一化当地/真太阳钟表时间；年、月柱及节令秒差使用同一出生 UTC 瞬间换算的中国标准时。",
		SolarTermReferenceAt:  formatSolarTime(*termReferenceBirth),
		SolarTermTimezone:     "UTC+08:00",
		AgeConversionRule:     conversionRule,
		BoundaryRule:          "节令边界按秒判定：出生时刻等于节令交接时刻时归入新节令；顺行取下一节令，逆行取出生时刻已进入的当前节令。",
		PreviousJie:           &previousEvidence,
		NextJie:               &nextEvidence,
		ReferenceJie:          &referenceEvidence,
		ReferenceDeltaSeconds: referenceDeltaSeconds,
	}

	for i := 0; i < 8; i++ {
		offset := i + 1
		if !forward {
			offset = -offset
		}
		sx := eightChar.GetMonth().Next(offset)
		daYun.Pillars = append(daYun.Pillars, model.Pillar{
			Gan: sx.GetHeavenStem().GetName(),
			Zhi: sx.GetEarthBranch().GetName(),
		})
	}
	return daYun
}

func addChildLimitToBirth(birth tyme.SolarTime, years, months, days, hours, minutes int) tyme.SolarTime {
	day := birth.GetDay() + days
	hour := birth.GetHour() + hours
	minute := birth.GetMinute() + minutes
	minuteCarry := minute / 60
	minute %= 60
	hour += minuteCarry
	day += hour / 24
	hour %= 24

	baseMonth, _ := tyme.SolarMonth{}.FromYm(birth.GetYear()+years, birth.GetMonth())
	month := baseMonth.Next(months)
	for day > month.GetDayCount() {
		day -= month.GetDayCount()
		month = month.Next(1)
	}
	result, _ := tyme.SolarTime{}.FromYmdHms(
		month.GetYear(), month.GetMonth(), day, hour, minute, birth.GetSecond(),
	)
	return *result
}

func surroundingJie(st tyme.SolarTime) (tyme.SolarTerm, tyme.SolarTerm) {
	previous := st.GetTerm()
	if !previous.IsJie() {
		previous = previous.Next(-1)
	}
	return previous, previous.Next(2)
}

func daYunTermEvidence(term tyme.SolarTerm, birth tyme.SolarTime) DaYunSolarTerm {
	termTime := term.GetJulianDay().GetSolarTime()
	return DaYunSolarTerm{
		Name:         term.GetName(),
		At:           formatSolarTime(termTime),
		DeltaSeconds: termTime.Subtract(birth),
	}
}

func formatSolarTime(st tyme.SolarTime) string {
	return fmt.Sprintf("%04d-%02d-%02dT%02d:%02d:%02d", st.GetYear(), st.GetMonth(), st.GetDay(), st.GetHour(), st.GetMinute(), st.GetSecond())
}

func defaultChildLimitProfile() (string, string) {
	return "tyme-default-seconds-v1", "按节令时刻差逐秒换算：259200秒=1年，21600秒=1月，720秒=1日，30秒=1时，1秒=2分。"
}

// ganInfo holds element and yang flag for a stem name.
type ganInfo struct {
	elem string
	yang bool
}

// GanInfoOf returns element and yang flag for a stem name.
func GanInfoOf(name string) ganInfo {
	gans := []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
	elems := []string{"木", "木", "火", "火", "土", "土", "金", "金", "水", "水"}
	yangs := []bool{true, false, true, false, true, false, true, false, true, false}
	for i, g := range gans {
		if g == name {
			return ganInfo{elems[i], yangs[i]}
		}
	}
	return ganInfo{"", false}
}

// dayGanElement returns the element name for the day stem.
func dayGanElement(dayGan string) string {
	return GanInfoOf(dayGan).elem
}

// ClassifyTenGod classifies a stem (stemName) against the day stem (dayGan),
// returning the ten-god name string.
// Only the VISIBLE day pillar stem itself is treated as 日主 (self) and excluded.
// Hidden stems that happen to share the same name as the day pillar stem
// are classified normally (比肩/劫财 when same element).
func ClassifyTenGod(stemName string, dayGan string, isDayPillarStem bool) string {
	stem := GanInfoOf(stemName)
	day := GanInfoOf(dayGan)
	if stem.elem == "" || day.elem == "" {
		return ""
	}

	// Only the visible day pillar stem is 日主 — hidden stems use normal classification
	if isDayPillarStem && stemName == dayGan {
		return "日主"
	}

	if stem.elem == day.elem {
		// Same element → 比肩 (same yang) or 劫财 (diff yang)
		if stem.yang == day.yang {
			return "比肩"
		}
		return "劫财"
	}
	// Stem element is "child" of day element (泄) → target: 食神/伤官
	// Element cycle: 木→火→土→金→水→木 (each is "child" of previous)
	child := map[string]string{
		"木": "火", "火": "土", "土": "金", "金": "水", "水": "木",
	}
	if child[day.elem] == stem.elem {
		if stem.yang == day.yang {
			return "食神"
		}
		return "伤官"
	}
	// Stem element is "parent" of day element (生) → 正印/偏印
	parent := map[string]string{
		"火": "木", "土": "火", "金": "土", "水": "金", "木": "水",
	}
	if parent[day.elem] == stem.elem {
		if stem.yang == day.yang {
			return "偏印"
		}
		return "正印"
	}
	// Stem element "controls" day element (克) → 正官/七杀
	ke := map[string]string{
		"木": "金", "火": "水", "土": "木", "金": "火", "水": "土",
	}
	if ke[day.elem] == stem.elem {
		if stem.yang == day.yang {
			return "七杀"
		}
		return "正官"
	}
	// Stem element is "controlled" by day element (耗) → 正财/偏财
	hao := map[string]string{
		"金": "木", "水": "火", "木": "土", "火": "金", "土": "水",
	}
	if hao[day.elem] == stem.elem {
		if stem.yang == day.yang {
			return "偏财"
		}
		return "正财"
	}
	return ""
}

// calcTenGodProportion computes the ten-god proportion using simple counting.
// It counts: year/month/hour stems (3) + all hidden stems from 4 branches (via tyme library).
// Returns 10 TenGodRatio in fixed order.
func calcTenGodProportion(ec *tyme.EightChar, dayGan string) []TenGodRatio {
	counts := make(map[string]int)

	// 3 visible stems (not day) — day stem is the visible one so isDayPillarStem=true only for day
	for _, fn := range [](struct {
		fn  func() tyme.SixtyCycle
		day bool
	}){
		{ec.GetYear, false},
		{ec.GetMonth, false},
		{ec.GetHour, false},
	} {
		sc := fn.fn()
		stemName := sc.GetHeavenStem().GetName()
		god := ClassifyTenGod(stemName, dayGan, fn.day)
		if god != "" && god != "日主" {
			counts[god]++
		}
	}

	// Hidden stems from 4 branches — all are from branch (not visible day stem) so isDayPillarStem=false
	for _, fn := range [](func() tyme.SixtyCycle){ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour} {
		sc := fn()
		for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
			stemName := hhs.GetHeavenStem().GetName()
			god := ClassifyTenGod(stemName, dayGan, false)
			if god != "" && god != "日主" {
				counts[god]++
			}
		}
	}

	total := 0
	for _, c := range counts {
		total += c
	}

	var result []TenGodRatio
	for _, name := range tenGodNames {
		c := counts[name]
		pct := 0.0
		if total > 0 {
			pct = math.Round(float64(c)*10000/float64(total)) / 100 // 2 decimal places
		}
		result = append(result, TenGodRatio{Name: name, Count: c, Percent: pct})
	}
	return result
}

func toTymeGender(gender string) (tyme.Gender, error) {
	switch gender {
	case "MALE":
		return tyme.MAN, nil
	case "FEMALE":
		return tyme.WOMAN, nil
	default:
		return 0, fmt.Errorf("invalid gender %q: must be MALE or FEMALE", gender)
	}
}
