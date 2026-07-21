package bazi

import (
	"fmt"
	"strings"

	"bazi/internal/model"
	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

// CalculateFromPillars analyzes a factual four-pillar chart whose calendar date
// is unknown. Besides validating every sixty-cycle pair, it rejects impossible
// year/month and day/hour stem combinations. Zi hour accepts both explicit
// late-Zi schools because pillars alone cannot distinguish 23:xx from 00:xx.
func (s *BaziService) CalculateFromPillars(yearGanZhi, monthGanZhi, dayGanZhi, hourGanZhi, gender string) (*BaziResult, error) {
	return s.calculateFromPillars(yearGanZhi, monthGanZhi, dayGanZhi, hourGanZhi, gender, true)
}

// CalculateSyntheticPillars is restricted to isolated rule fixtures that need
// arbitrary, individually valid sixty-cycle pillars. It must not be used for a
// factual birth chart because the four pillars may not coexist in real time.
func (s *BaziService) CalculateSyntheticPillars(yearGanZhi, monthGanZhi, dayGanZhi, hourGanZhi, gender string) (*BaziResult, error) {
	return s.calculateFromPillars(yearGanZhi, monthGanZhi, dayGanZhi, hourGanZhi, gender, false)
}

func (s *BaziService) calculateFromPillars(yearGanZhi, monthGanZhi, dayGanZhi, hourGanZhi, gender string, validateLinkage bool) (*BaziResult, error) {
	gender = strings.ToUpper(strings.TrimSpace(gender))
	if gender != "MALE" && gender != "FEMALE" {
		return nil, fmt.Errorf("invalid gender %q: must be MALE or FEMALE", gender)
	}

	yearP, err := parsePillarString(yearGanZhi, "year")
	if err != nil {
		return nil, err
	}
	monthP, err := parsePillarString(monthGanZhi, "month")
	if err != nil {
		return nil, err
	}
	dayP, err := parsePillarString(dayGanZhi, "day")
	if err != nil {
		return nil, err
	}
	hourP, err := parsePillarString(hourGanZhi, "hour")
	if err != nil {
		return nil, err
	}

	// 构造虚拟 tyme.EightChar 复用现有计算函数
	yearSC, err := tyme.SixtyCycle{}.FromName(yearP.Gan + yearP.Zhi)
	if err != nil {
		return nil, fmt.Errorf("invalid year pillar %q: %w", yearGanZhi, err)
	}
	monthSC, err := tyme.SixtyCycle{}.FromName(monthP.Gan + monthP.Zhi)
	if err != nil {
		return nil, fmt.Errorf("invalid month pillar %q: %w", monthGanZhi, err)
	}
	daySC, err := tyme.SixtyCycle{}.FromName(dayP.Gan + dayP.Zhi)
	if err != nil {
		return nil, fmt.Errorf("invalid day pillar %q: %w", dayGanZhi, err)
	}
	hourSC, err := tyme.SixtyCycle{}.FromName(hourP.Gan + hourP.Zhi)
	if err != nil {
		return nil, fmt.Errorf("invalid hour pillar %q: %w", hourGanZhi, err)
	}
	if validateLinkage {
		if err := validatePillarLinkage(*yearSC, *monthSC, *daySC, *hourSC); err != nil {
			return nil, err
		}
	}
	ec := tyme.EightChar{}.FromSixtyCycle(*yearSC, *monthSC, *daySC, *hourSC)

	result := &BaziResult{
		RuleVersion: RuleVersion,
		School:      RuleSchool,
		RuleMeta:    DefaultRuleMeta(),
		YearPillar:  yearP,
		MonthPillar: monthP,
		DayPillar:   dayP,
		HourPillar:  hourP,
	}
	result.MonthSeason = observeMonthSeason(result.MonthPillar.Zhi)

	// 命宫
	mingGongGanZhi, err := calcMingGongGanZhi(result.YearPillar.Gan, result.MonthPillar.Zhi, result.HourPillar.Zhi)
	if err != nil {
		return nil, fmt.Errorf("计算命宫失败: %w", err)
	}
	result.MingGong = buildMingGongDetail(mingGongGanZhi)

	// 五行/身强
	result.FiveElements = calcFiveElements(&ec)
	result.ElementDetail = calcElementDetail(&ec)
	result.BodyStrength = calcBodyStrength(&ec)

	// 十神/纳音/藏干
	result.TenGods = calcTenGods(&ec)
	result.NaYin = calcNaYin(&ec)
	result.HiddenStems = calcHiddenStems(&ec)

	// 大运：因无公历日期无法起运，给默认值
	result.DaYunInfo = DaYunInfo{
		StartAge:           0,
		Direction:          "未计算（需要准确公历日期）",
		CalculationProfile: "unavailable-without-birth-date",
		TimeBasis:          "仅有四柱，缺少可用于节令时差计算的准确公历出生时刻。",
		Pillars:            []model.Pillar{},
	}

	// 干支分析 / 格局
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

	// 柱详情 / 神煞 / 调候文本
	result.TenGodProportion = calcTenGodProportion(&ec, result.DayPillar.Gan)
	result.TenGodAnalysis = ObserveTenGodDistribution(result.TenGodProportion)

	if err := s.enrichPillarDetails(result, gender); err != nil {
		return nil, fmt.Errorf("计算柱位神煞失败: %w", err)
	}

	// 调候用神
	tiaohouResult, _ := AnalyzeTiaohouForPillars(
		result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar,
	)
	result.Tiaohou = tiaohouResult

	// 原始五行分布
	result.MissingElements = data.FindMissingElements(result.FiveElements)

	return result, nil
}

// parsePillarString 解析形如 "甲子" 的干支字符串为 model.Pillar。
// name 仅为错误信息中的位置标识（"year"/"month"/"day"/"hour"）。
func parsePillarString(s, name string) (model.Pillar, error) {
	s = strings.TrimSpace(s)
	runes := []rune(s)
	if len(runes) != 2 {
		return model.Pillar{}, fmt.Errorf("%s pillar must be 2 characters (Gan+Zhi), got %d: %q", name, len(runes), s)
	}
	gan := string(runes[0])
	zhi := string(runes[1])
	if GanInfoOf(gan).elem == "" {
		return model.Pillar{}, fmt.Errorf("%s pillar has invalid Gan character: %q", name, gan)
	}
	if data.ZhiIndex(zhi) < 0 {
		return model.Pillar{}, fmt.Errorf("%s pillar has invalid Zhi character: %q", name, zhi)
	}
	return model.Pillar{Gan: gan, Zhi: zhi}, nil
}

// validatePillarLinkage rejects individually valid sixty-cycle pillars that
// cannot belong to one chart. Month stems follow the year stem (五虎遁), and
// hour stems follow the day stem (五鼠遁).
func validatePillarLinkage(year, month, day, hour tyme.SixtyCycle) error {
	monthOffset := month.GetEarthBranch().Next(-2).GetIndex()
	expectedMonthStem := tyme.HeavenStem{}.FromIndex(
		(year.GetHeavenStem().GetIndex()+1)*2 + monthOffset,
	)
	if month.GetHeavenStem().GetName() != expectedMonthStem.GetName() {
		return fmt.Errorf(
			"month pillar %q is inconsistent with year pillar %q: expected %s%s by five-tiger month derivation",
			month.GetName(), year.GetName(), expectedMonthStem.GetName(), month.GetEarthBranch().GetName(),
		)
	}

	hourBranchIndex := hour.GetEarthBranch().GetIndex()
	expectedHourStem := fiveRatHourStem(day.GetHeavenStem(), hourBranchIndex)
	hourStem := hour.GetHeavenStem().GetName()
	if hourStem == expectedHourStem.GetName() {
		return nil
	}
	if hourBranchIndex == 0 {
		lateZiHourStem := fiveRatHourStem(day.Next(1).GetHeavenStem(), hourBranchIndex)
		if hourStem == lateZiHourStem.GetName() {
			return nil
		}
		return fmt.Errorf(
			"hour pillar %q is inconsistent with day pillar %q: expected %s子 for same-day/early-Zi or %s子 for late-Zi-same-day",
			hour.GetName(), day.GetName(), expectedHourStem.GetName(), lateZiHourStem.GetName(),
		)
	}
	return fmt.Errorf(
		"hour pillar %q is inconsistent with day pillar %q: expected %s%s by five-rat hour derivation",
		hour.GetName(), day.GetName(), expectedHourStem.GetName(), hour.GetEarthBranch().GetName(),
	)
}

func fiveRatHourStem(dayStem tyme.HeavenStem, hourBranchIndex int) tyme.HeavenStem {
	return tyme.HeavenStem{}.FromIndex(dayStem.GetIndex()%5*2 + hourBranchIndex)
}
