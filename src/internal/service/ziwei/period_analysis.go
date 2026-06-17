package ziwei

import (
	"fmt"
	"sort"
	"strings"
)

const (
	PeriodRuleVersion = "ziwei-period-2026-06-16"
	PeriodSchool      = "紫微斗数-大限流年叠盘-v1"
)

type PeriodAnalysis struct {
	RuleVersion     string               `json:"rule_version"`
	School          string               `json:"school"`
	Layer           string               `json:"layer"`
	Title           string               `json:"title"`
	TimeLabel       string               `json:"time_label"`
	GanZhi          string               `json:"gan_zhi,omitempty"`
	Score           int                  `json:"score"`
	Tone            string               `json:"tone"`
	Summary         string               `json:"summary"`
	Method          []string             `json:"method"`
	Highlights      []PeriodHighlight    `json:"highlights"`
	FocusPalaces    []PeriodPalaceFocus  `json:"focus_palaces"`
	Evidence        []PeriodEvidence     `json:"evidence"`
	Recommendations []string             `json:"recommendations"`
	Risks           []string             `json:"risks"`
	DayunStages     []DayunStageAnalysis `json:"dayun_stages,omitempty"`
}

type DayunStageAnalysis struct {
	StartAge  int      `json:"start_age"`
	EndAge    int      `json:"end_age"`
	Palace    string   `json:"palace"`
	Branch    string   `json:"branch"`
	Score     int      `json:"score"`
	Tone      string   `json:"tone"`
	MainStars []string `json:"main_stars"`
	AuxStars  []string `json:"aux_stars"`
	FourHua   []string `json:"four_hua"`
	Sanfang   []string `json:"sanfang"`
	Summary   string   `json:"summary"`
	Advice    []string `json:"advice"`
	Current   bool     `json:"current"`
}

type PeriodHighlight struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type PeriodPalaceFocus struct {
	Palace      string   `json:"palace"`
	Branch      string   `json:"branch"`
	Score       int      `json:"score"`
	Level       string   `json:"level"`
	MainStars   []string `json:"main_stars"`
	AuxStars    []string `json:"aux_stars"`
	PeriodStars []string `json:"period_stars"`
	FourHua     []string `json:"four_hua"`
	Sanfang     []string `json:"sanfang"`
	Reason      string   `json:"reason"`
	Suggestion  string   `json:"suggestion"`
}

type PeriodEvidence struct {
	Type   string `json:"type"`
	Label  string `json:"label"`
	Value  string `json:"value"`
	Impact string `json:"impact"`
}

func BuildDayunAnalysis(chart *ZiWeiChart, stages Dayun, currentAge int) *PeriodAnalysis {
	if chart == nil {
		return nil
	}
	out := basePeriodAnalysis("dayun", "大限分析", "十年阶段")
	out.Method = []string{
		"以五行局数作为起限年龄。",
		"按阳男阴女顺行、阴男阳女逆行，从命宫布十二个十年阶段。",
		"每个阶段以落入宫位、主星亮度、辅煞、四化与三方四正综合评分。",
	}

	if len(stages) == 0 {
		stages = calcDayunFromChart(chart)
	}

	total := 0
	for _, stage := range stages {
		idx := palaceIndexByName(chart, stage.Palace)
		if idx < 0 {
			continue
		}
		p := chart.Palaces[idx]
		score, evidence := scorePalaceForPeriod(p, nil)
		item := DayunStageAnalysis{
			StartAge:  stage.StartAge,
			EndAge:    stage.EndAge,
			Palace:    p.Name,
			Branch:    p.Branch,
			Score:     score,
			Tone:      toneForScore(score),
			MainStars: palaceMainStars(p),
			AuxStars:  palaceAuxStars(p),
			FourHua:   cloneStrings(p.FourHua),
			Sanfang:   cloneStrings(sanfangNames(chart, idx)),
			Current:   currentAge >= stage.StartAge && currentAge <= stage.EndAge,
		}
		item.Summary = buildStageSummary(item, evidence)
		item.Advice = stageAdvice(item, evidence)
		out.DayunStages = append(out.DayunStages, item)
		total += score
	}

	if len(out.DayunStages) > 0 {
		out.Score = total / len(out.DayunStages)
	} else {
		out.Score = 60
	}
	out.Tone = toneForScore(out.Score)
	out.Summary = fmt.Sprintf("大限用来判断人生十年阶段的底色。当前可重点看标记为“当前”的阶段，再叠加流年、流月、流日触发。整体盘面阶段均值为%d分，属%s。", out.Score, out.Tone)

	current := currentDayunStage(out.DayunStages)
	if current != nil {
		out.Highlights = append(out.Highlights,
			PeriodHighlight{Label: "当前大限", Value: fmt.Sprintf("%d-%d岁", current.StartAge, current.EndAge), Note: current.Palace + "为这十年的主题宫"},
			PeriodHighlight{Label: "主调", Value: current.Tone, Note: current.Summary},
		)
		out.FocusPalaces = append(out.FocusPalaces, PeriodPalaceFocus{
			Palace:     current.Palace,
			Branch:     current.Branch,
			Score:      current.Score,
			Level:      current.Tone,
			MainStars:  cloneStrings(current.MainStars),
			AuxStars:   cloneStrings(current.AuxStars),
			FourHua:    cloneStrings(current.FourHua),
			Sanfang:    cloneStrings(current.Sanfang),
			Reason:     "当前年龄落入此大限，后续流年/月/日都以它作为底色。",
			Suggestion: firstOrFallback(current.Advice, "先按本宫主题设定十年方向，再看短周期触发。"),
		})
	}
	out.Evidence = append(out.Evidence,
		PeriodEvidence{Type: "start_age", Label: "起限", Value: fmt.Sprintf("%d岁", chart.JuValue), Impact: "五行局数决定第一步大限起点"},
		PeriodEvidence{Type: "direction", Label: "行运方向", Value: dayunDirectionLabel(chart), Impact: "决定十二宫顺排或逆排"},
	)
	out.Recommendations = []string{
		"先看当前大限宫位定十年主题，再看流年触发哪些宫位。",
		"大限分数高时宜布局长期目标；分数低时优先修复该宫位主题。",
	}
	out.Risks = []string{
		"大限只定底色，不能替代流年和流月触发。",
		"若当前大限见煞曜或化忌，重要决策应拆成更短周期验证。",
	}
	return out
}

func BuildLiunianAnalysis(base, period *ZiWeiChart, year int) *PeriodAnalysis {
	if base == nil || period == nil {
		return nil
	}
	stem := fixMod(year-4, 10)
	branch := fixMod(year-4, 12)
	interp := NewPeriodInterpreter(base.GetBirthData()).AnalyzeLiunian(period, year)
	return buildTransitAnalysis(base, period, transitConfig{
		Layer:       "liunian",
		Title:       "流年分析",
		TimeLabel:   fmt.Sprintf("%d年", year),
		GanZhi:      stemName(stem) + branchName(branch),
		Stem:        stem,
		Branch:      branch,
		PeriodStars: period.LiuNianStars,
		Interpreter: interp,
		Method: []string{
			"按目标年份取流年干支。",
			"用流年天干引动四化，并安流禄、流羊、流陀、流马。",
			"以触发宫位、命局刑冲合、十神与宫位星曜综合给出年度主题。",
		},
	})
}

func BuildLiuyueAnalysis(base, period *ZiWeiChart, year, month int) *PeriodAnalysis {
	if base == nil || period == nil {
		return nil
	}
	stem, branch := targetMonthStemBranch(year, month)
	interp := NewPeriodInterpreter(base.GetBirthData()).AnalyzeLiuyue(period, year, month)
	return buildTransitAnalysis(base, period, transitConfig{
		Layer:       "liuyue",
		Title:       "流月分析",
		TimeLabel:   fmt.Sprintf("%d年%d月", year, month),
		GanZhi:      stemName(stem) + branchName(branch),
		Stem:        stem,
		Branch:      branch,
		PeriodStars: period.LiuYueStars,
		Interpreter: interp,
		Method: []string{
			"按年份天干与月份推流月天干，按月份推流月地支。",
			"用流月天干引动当月四化。",
			"以当月触发宫位、流年底色和命局关系判断本月主轴。",
		},
	})
}

func BuildLiuriAnalysis(base, period *ZiWeiChart, year, month, day int) *PeriodAnalysis {
	if base == nil || period == nil {
		return nil
	}
	stem, branch := targetDayStemBranch(year, month, day)
	interp := NewPeriodInterpreter(base.GetBirthData()).AnalyzeLiuri(period, year, month, day)
	return buildTransitAnalysis(base, period, transitConfig{
		Layer:       "liuri",
		Title:       "流日分析",
		TimeLabel:   fmt.Sprintf("%d年%d月%d日", year, month, day),
		GanZhi:      stemName(stem) + branchName(branch),
		Stem:        stem,
		Branch:      branch,
		PeriodStars: period.LiuRiStars,
		Interpreter: interp,
		Method: []string{
			"按日期推流日干支。",
			"用流日天干引动日内四化。",
			"以当天触发宫位、时辰评分和短期情绪/健康提示判断行动窗口。",
		},
	})
}

type transitConfig struct {
	Layer       string
	Title       string
	TimeLabel   string
	GanZhi      string
	Stem        int
	Branch      int
	PeriodStars [12][]string
	Interpreter any
	Method      []string
}

func buildTransitAnalysis(base, period *ZiWeiChart, cfg transitConfig) *PeriodAnalysis {
	out := basePeriodAnalysis(cfg.Layer, cfg.Title, cfg.TimeLabel)
	out.GanZhi = cfg.GanZhi
	out.Method = cfg.Method

	periodStarNames := flattenPeriodStars(cfg.PeriodStars)
	triggered := periodPalaceFocus(base, period, cfg.PeriodStars)
	out.FocusPalaces = triggered
	out.Score = scoreFromFocusAndInterpreter(triggered, cfg.Interpreter)
	out.Tone = toneForScore(out.Score)
	out.Summary = transitSummary(cfg.Layer, cfg.TimeLabel, cfg.GanZhi, out.Score, out.Tone, triggered, cfg.Interpreter)
	out.Highlights = transitHighlights(cfg, triggered, periodStarNames)
	out.Evidence = transitEvidence(base, cfg, triggered, periodStarNames)
	out.Recommendations = transitRecommendations(cfg.Layer, out.Score, triggered, cfg.Interpreter)
	out.Risks = transitRisks(cfg.Layer, out.Score, triggered, cfg.Interpreter)
	return out
}

func basePeriodAnalysis(layer, title, timeLabel string) *PeriodAnalysis {
	return &PeriodAnalysis{
		RuleVersion:     PeriodRuleVersion,
		School:          PeriodSchool,
		Layer:           layer,
		Title:           title,
		TimeLabel:       timeLabel,
		Score:           60,
		Tone:            "平稳",
		Method:          []string{},
		Highlights:      []PeriodHighlight{},
		FocusPalaces:    []PeriodPalaceFocus{},
		Evidence:        []PeriodEvidence{},
		Recommendations: []string{},
		Risks:           []string{},
	}
}

func periodPalaceFocus(base, period *ZiWeiChart, periodStars [12][]string) []PeriodPalaceFocus {
	out := []PeriodPalaceFocus{}
	for i := range period.Palaces {
		p := period.Palaces[i]
		stars := append([]string(nil), periodStars[i]...)
		if len(stars) == 0 && len(p.FourHua) == 0 {
			continue
		}
		score, evidence := scorePalaceForPeriod(p, stars)
		focus := PeriodPalaceFocus{
			Palace:      p.Name,
			Branch:      p.Branch,
			Score:       score,
			Level:       toneForScore(score),
			MainStars:   palaceMainStars(p),
			AuxStars:    palaceAuxStars(p),
			PeriodStars: stars,
			FourHua:     cloneStrings(p.FourHua),
			Sanfang:     cloneStrings(sanfangNames(base, i)),
			Reason:      palaceFocusReason(p, stars, evidence),
			Suggestion:  palaceFocusSuggestion(p, score, stars),
		}
		out = append(out, focus)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score == out[j].Score {
			return len(out[i].PeriodStars)+len(out[i].FourHua) > len(out[j].PeriodStars)+len(out[j].FourHua)
		}
		return out[i].Score > out[j].Score
	})
	if len(out) > 6 {
		return out[:6]
	}
	return out
}

func scorePalaceForPeriod(p PalaceInfo, periodStars []string) (int, []string) {
	score := 55
	var evidence []string
	mainStars := palaceMainStars(p)
	if len(mainStars) == 0 {
		score -= 4
		evidence = append(evidence, "空宫")
	} else {
		score += min(len(mainStars)*5, 12)
		evidence = append(evidence, "主星")
	}
	for _, star := range mainStars {
		switch palaceStarBrightness(p, star) {
		case "庙", "旺":
			score += 6
			evidence = append(evidence, star+"亮")
		case "陷", "不":
			score -= 6
			evidence = append(evidence, star+"弱")
		}
	}
	for _, star := range palaceAuxStars(p) {
		switch auxEvidenceType(star) {
		case "soft_star":
			score += 4
			evidence = append(evidence, star+"助")
		case "tough_star":
			score -= 6
			evidence = append(evidence, star+"压")
		default:
			score += 1
		}
	}
	for _, hua := range p.FourHua {
		switch {
		case strings.Contains(hua, "化禄"), strings.Contains(hua, "化权"), strings.Contains(hua, "化科"):
			score += 5
			evidence = append(evidence, hua)
		case strings.Contains(hua, "化忌"):
			score -= 8
			evidence = append(evidence, hua)
		}
	}
	for _, star := range periodStars {
		switch {
		case strings.Contains(star, "化禄"), star == "流禄":
			score += 8
			evidence = append(evidence, star)
		case strings.Contains(star, "化权"), strings.Contains(star, "化科"), star == "流马":
			score += 5
			evidence = append(evidence, star)
		case strings.Contains(star, "化忌"), star == "流羊", star == "流陀":
			score -= 8
			evidence = append(evidence, star)
		default:
			score += 2
			evidence = append(evidence, star)
		}
	}
	return clampScore(score), evidence
}

func scoreFromFocusAndInterpreter(focus []PeriodPalaceFocus, interpreter any) int {
	score := 60
	if len(focus) > 0 {
		total := 0
		for _, item := range focus {
			total += item.Score
		}
		score = total / len(focus)
	}
	switch v := interpreter.(type) {
	case *LiunianResult:
		if v != nil {
			score = (score + v.Score) / 2
		}
	case *LiuyueResult:
		if v != nil {
			score = (score + v.Score) / 2
		}
	case *LiuriResult:
		if v != nil {
			score = (score + v.Score) / 2
		}
	}
	return clampScore(score)
}

func transitSummary(layer, timeLabel, ganZhi string, score int, tone string, focus []PeriodPalaceFocus, interpreter any) string {
	focusText := "暂无明显流曜落点"
	if len(focus) > 0 {
		names := make([]string, 0, min(3, len(focus)))
		for _, item := range focus[:min(3, len(focus))] {
			names = append(names, item.Palace)
		}
		focusText = strings.Join(names, "、")
	}
	prefix := fmt.Sprintf("%s%s%s，综合%d分，属%s。重点宫位：%s。", timeLabel, optionalGanZhi(ganZhi), layerVerb(layer), score, tone, focusText)
	switch v := interpreter.(type) {
	case *LiunianResult:
		if v != nil {
			return prefix + v.OverallTone
		}
	case *LiuyueResult:
		if v != nil {
			return prefix + v.Effect
		}
	case *LiuriResult:
		if v != nil {
			return prefix + v.Summary
		}
	}
	return prefix
}

func transitHighlights(cfg transitConfig, focus []PeriodPalaceFocus, periodStars []string) []PeriodHighlight {
	items := []PeriodHighlight{
		{Label: "干支", Value: cfg.GanZhi, Note: fmt.Sprintf("天干%s引动四化，地支%s参与命局关系", stemName(cfg.Stem), branchName(cfg.Branch))},
	}
	if len(periodStars) > 0 {
		items = append(items, PeriodHighlight{Label: "流曜", Value: strings.Join(periodStars[:min(4, len(periodStars))], "、"), Note: "优先看这些星曜落入的宫位"})
	}
	if len(focus) > 0 {
		items = append(items, PeriodHighlight{Label: "首要宫位", Value: focus[0].Palace, Note: focus[0].Reason})
	}
	return items
}

func transitEvidence(base *ZiWeiChart, cfg transitConfig, focus []PeriodPalaceFocus, periodStars []string) []PeriodEvidence {
	var out []PeriodEvidence
	out = append(out,
		PeriodEvidence{Type: "ganzhi", Label: "周期干支", Value: cfg.GanZhi, Impact: "决定十神、五行和本层四化"},
		PeriodEvidence{Type: "relation", Label: "命局关系", Value: relationForBranch(base, cfg.Branch), Impact: "刑冲合伏吟会改变事件稳定度"},
	)
	if len(periodStars) > 0 {
		out = append(out, PeriodEvidence{Type: "period_star", Label: "本层流曜", Value: strings.Join(periodStars, "、"), Impact: "流曜落宫是本层事件触发点"})
	}
	for _, item := range focus[:min(3, len(focus))] {
		out = append(out, PeriodEvidence{Type: "focus_palace", Label: item.Palace, Value: fmt.Sprintf("%d分", item.Score), Impact: item.Reason})
	}
	return out
}

func transitRecommendations(layer string, score int, focus []PeriodPalaceFocus, interpreter any) []string {
	out := []string{}
	if len(focus) > 0 {
		top := focus[0]
		if top.Score >= 72 || top.Score < 55 || containsAny(top.PeriodStars, []string{"流羊", "流陀", "流马"}) || containsHuaJi(top.PeriodStars) || containsHuaJi(top.FourHua) {
			out = append(out, fmt.Sprintf("%s：%s", top.Palace, top.Suggestion))
		}
	}
	switch v := interpreter.(type) {
	case *LiunianResult:
		if v != nil && v.KeyTips != "" {
			out = append(out, v.KeyTips)
		}
	case *LiuyueResult:
		if v != nil && v.Health != "" {
			out = append(out, v.Health)
		}
	case *LiuriResult:
		if v != nil && v.EmotionalState != "" {
			out = append(out, v.EmotionalState)
		}
	}
	if layer == "liuri" {
		out = append(out, "当天只取一到两件关键事执行，时辰分数低的窗口少做决策。")
	}
	return periodUniqueStrings(out)
}

func transitRisks(layer string, score int, focus []PeriodPalaceFocus, interpreter any) []string {
	var out []string
	if score < 55 {
		out = append(out, "分数偏低，先看化忌、流羊、流陀是否落在重点宫。")
	}
	for _, item := range focus {
		if containsAny(item.PeriodStars, []string{"流羊", "流陀"}) || containsHuaJi(item.PeriodStars) || containsHuaJi(item.FourHua) {
			out = append(out, fmt.Sprintf("%s见压力星，相关主题不宜急进。", item.Palace))
		}
	}
	switch v := interpreter.(type) {
	case *LiunianResult:
		if v != nil && strings.Contains(v.RelationToMing, "冲") {
			out = append(out, "流年见冲，年度计划要预留变动成本。")
		}
	case *LiuyueResult:
		if v != nil && strings.Contains(v.RelationToMing, "刑") {
			out = append(out, "流月见刑，本月避免因沟通和手续拖累节奏。")
		}
	case *LiuriResult:
		if v != nil && strings.Contains(v.RelationToMing, "伏吟") {
			out = append(out, "流日伏吟，容易反复纠结，重大决定延后复核。")
		}
	}
	return periodUniqueStrings(out)
}

func buildStageSummary(item DayunStageAnalysis, evidence []string) string {
	stars := joinOrNone(item.MainStars)
	return fmt.Sprintf("%s%s大限主看%s，主星%s，评分%d分。依据：%s。", item.Palace, formatBranch(item.Branch), palaceFocus(item.Palace), stars, item.Score, joinOrNone(evidence))
}

func stageAdvice(item DayunStageAnalysis, evidence []string) []string {
	out := []string{}
	if item.Score >= 72 {
		out = append(out, fmt.Sprintf("%s星曜条件较足，可作为本阶段较有把握的承接点。", item.Palace))
	} else if item.Score < 55 {
		out = append(out, fmt.Sprintf("%s分数偏低，相关事项先看化忌、煞曜与三方四正压力。", item.Palace))
	}
	if containsHuaJi(item.FourHua) || containsAny(evidence, []string{"流羊", "流陀"}) {
		out = append(out, "见忌煞时，先做风险边界和备选方案。")
	}
	return out
}

func currentDayunStage(stages []DayunStageAnalysis) *DayunStageAnalysis {
	for i := range stages {
		if stages[i].Current {
			return &stages[i]
		}
	}
	return nil
}

func palaceFocusReason(p PalaceInfo, periodStars []string, evidence []string) string {
	if len(periodStars) > 0 {
		return fmt.Sprintf("%s见%s，被本层流曜直接触发。", p.Name, strings.Join(periodStars, "、"))
	}
	if len(p.FourHua) > 0 {
		return fmt.Sprintf("%s本宫四化为%s，事件容易在此宫显形。", p.Name, strings.Join(p.FourHua, "、"))
	}
	return fmt.Sprintf("%s星曜条件较明显：%s。", p.Name, joinOrNone(evidence))
}

func palaceFocusSuggestion(p PalaceInfo, score int, periodStars []string) string {
	focus := palaceFocus(p.Name)
	if score >= 72 {
		return fmt.Sprintf("%s主题可主动推进，优先使用已有资源。", focus)
	}
	if score < 55 {
		return fmt.Sprintf("%s主题先控风险，避免临时加码。", focus)
	}
	if containsAny(periodStars, []string{"流马"}) {
		return fmt.Sprintf("%s主题有变动信号，适合移动、沟通、调整方案。", focus)
	}
	return fmt.Sprintf("%s主题宜稳步推进。", focus)
}

func sanfangNames(chart *ZiWeiChart, palaceIdx int) []string {
	if chart == nil || palaceIdx < 0 || palaceIdx >= len(chart.Palaces) {
		return nil
	}
	oppositeIdx, trine1Idx, trine2Idx := chartSanfangIndexes(chart, palaceIdx)
	return []string{chart.Palaces[oppositeIdx].Name, chart.Palaces[trine1Idx].Name, chart.Palaces[trine2Idx].Name}
}

func palaceIndexByName(chart *ZiWeiChart, name string) int {
	for i, p := range chart.Palaces {
		if p.Name == name {
			return i
		}
	}
	return -1
}

func flattenPeriodStars(stars [12][]string) []string {
	seen := map[string]bool{}
	var out []string
	for _, group := range stars {
		for _, star := range group {
			if star == "" || seen[star] {
				continue
			}
			seen[star] = true
			out = append(out, star)
		}
	}
	sort.Strings(out)
	return out
}

func relationForBranch(chart *ZiWeiChart, branch int) string {
	if chart == nil || chart.GetBirthData() == nil {
		return "未恢复命局参数"
	}
	return NewPeriodInterpreter(chart.GetBirthData()).describeRelation(branch)
}

func dayunDirectionLabel(chart *ZiWeiChart) string {
	if chart == nil || chart.GetBirthData() == nil {
		return "未知"
	}
	isMale := chart.GetBirthData().Gender == "男"
	forward := (BranchIsYang(chart.YearBranch) && isMale) || (!BranchIsYang(chart.YearBranch) && !isMale)
	if forward {
		return "顺行"
	}
	return "逆行"
}

func toneForScore(score int) string {
	switch {
	case score >= 78:
		return "强势"
	case score >= 65:
		return "顺势"
	case score >= 50:
		return "平稳"
	default:
		return "承压"
	}
}

func layerVerb(layer string) string {
	switch layer {
	case "liunian":
		return "定年度底色"
	case "liuyue":
		return "定月度主题"
	case "liuri":
		return "定当天行动窗口"
	default:
		return "形成周期影响"
	}
}

func optionalGanZhi(gz string) string {
	if gz == "" {
		return ""
	}
	return "（" + gz + "）"
}

func firstOrFallback(items []string, fallback string) string {
	if len(items) > 0 && items[0] != "" {
		return items[0]
	}
	return fallback
}

func containsHuaJi(items []string) bool {
	for _, item := range items {
		if strings.Contains(item, "化忌") {
			return true
		}
	}
	return false
}

func containsAny(items, needles []string) bool {
	for _, item := range items {
		for _, needle := range needles {
			if strings.Contains(item, needle) {
				return true
			}
		}
	}
	return false
}

func periodUniqueStrings(items []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, item := range items {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func cloneStrings(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	return append([]string{}, items...)
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}

func fixMod(v, mod int) int {
	out := v % mod
	if out < 0 {
		out += mod
	}
	return out
}
