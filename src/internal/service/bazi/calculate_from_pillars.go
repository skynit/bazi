package bazi

import (
	"fmt"
	"strings"

	"bazi/internal/model"
	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

// CalculateFromPillars 直接从四柱干支计算八字命局
// yearGanZhi, monthGanZhi, dayGanZhi, hourGanZhi 形如 "甲子"
// 适用于测试场景：经典命例的干支已知，但公历日期不一定准确
// 不计算大运（依赖公历起运年龄），重点是分析层（身强/格局/调候/神煞/十神）的计算
func (s *BaziService) CalculateFromPillars(yearGanZhi, monthGanZhi, dayGanZhi, hourGanZhi, gender string) (*BaziResult, error) {
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

	// 命宫
	mingGongGanZhi, err := data.CalcMingGong(result.YearPillar.Gan, result.MonthPillar.Zhi, result.HourPillar.Zhi)
	if err != nil {
		return nil, fmt.Errorf("计算命宫失败: %w", err)
	}
	result.MingGong = data.BuildMingGongDetail(mingGongGanZhi)

	// 日柱描述
	riZhuKey := result.DayPillar.Gan + "日" + result.DayPillar.Zhi
	result.RiZhuDesc = data.SiZiSummaries[riZhuKey]

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
		StartAge:  0,
		Direction: "未计算（需要准确公历日期）",
		Pillars:   []model.Pillar{},
	}

	// 冲合刑
	result.ClashHarmony = calcClashHarmony(&ec)

	// 干支分析 / 格局
	result.GanZhiAnalysis = CalcGanZhiAnalysis(
		result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar,
	)
	result.PatternAnalysis = AnalyzePatternExtended(
		[]model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar},
		result.MonthPillar.Zhi,
		result.FiveElements,
		result.BodyStrength,
	)

	// 柱详情 / 神煞 / 调候文本
	dayElem := data.GanElement[result.DayPillar.Gan]
	result.TenGodProportion = calcTenGodProportion(&ec, result.DayPillar.Gan)
	pillarTenGods := calcPillarTenGods(&ec, result.DayPillar.Gan)
	analyzer := &TenGodAnalyzer{}
	result.TenGodAnalysis = analyzer.AnalyzeTenGod(result.TenGodProportion, dayElem, result.BodyStrength, pillarTenGods, gender)

	// 从月支反推公历月份，供季节/调候文本查询（无具体日期，仅取对应月份）
	birthMonth := zhiToMonth(result.MonthPillar.Zhi)
	s.enrichPillarDetails(result, birthMonth, gender)

	enrichRiZhuText(result)

	// 三命通会/穷通宝鉴等知识层
	enrichWuxingSeason(result, birthMonth)
	enrichJiaZiDetail(result)
	enrichHealthNote(result)

	// 调候用神
	tiaohouResult, _ := AnalyzeTiaohou(result.DayPillar.Gan, result.MonthPillar.Zhi)
	result.Tiaohou = tiaohouResult

	// 五行流通
	pillars := []model.Pillar{result.YearPillar, result.MonthPillar, result.DayPillar, result.HourPillar}
	result.WuXingFlow = data.AnalyzeWuXingFlowV2(result.FiveElements, dayElem)
	result.TongGuan = data.FindTongGuan(pillars, dayElem, result.MonthPillar.Zhi)
	result.MissingElements = data.FindMissingElements(result.FiveElements)
	result.FlowPatternDesc = data.BuildFlowPatternDesc(result.WuXingFlow, result.TongGuan, result.MissingElements)

	// 大运流通（无大运则跳过）
	if len(result.DaYunInfo.Pillars) > 0 {
		result.DaYunFlow = data.CalcDaYunFlow(result.DayPillar.Gan, result.FiveElements, result.DaYunInfo.Pillars, result.DaYunInfo.StartAge)
	}

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
	if _, ok := data.GanElement[gan]; !ok {
		return model.Pillar{}, fmt.Errorf("%s pillar has invalid Gan character: %q", name, gan)
	}
	if data.ZhiIndex(zhi) < 0 {
		return model.Pillar{}, fmt.Errorf("%s pillar has invalid Zhi character: %q", name, zhi)
	}
	return model.Pillar{Gan: gan, Zhi: zhi}, nil
}

// zhiToMonth 由月支反推公历月份（寅=1月, 卯=2月, ..., 丑=12月）。
// 仅用于从四柱已知但公历日期不确定的场景（如经典命例核对）。
func zhiToMonth(zhi string) int {
	idx := data.ZhiIndex(zhi)
	if idx < 0 {
		return 0
	}
	return ((idx - 2 + 12) % 12) + 1
}
