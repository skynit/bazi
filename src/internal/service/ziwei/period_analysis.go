package ziwei

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

const (
	PeriodRuleVersion          = "ziwei-period-2026-07-18.16"
	PeriodSchool               = "紫微斗数-大限流年叠盘-v1"
	periodMixedEvidenceBasis   = "mixed_deterministic_projection_and_unadjudicated_traditional_labels"
	periodPlacementBasis       = "deterministic_rule_projection"
	periodInterpretationBasis  = "traditional_rule_labels"
	periodInterpretationStatus = "not_adjudicated"
)

type PeriodAnalysis struct {
	RuleVersion          string                `json:"rule_version"`
	School               string                `json:"school"`
	Layer                string                `json:"layer"`
	Title                string                `json:"title"`
	TimeLabel            string                `json:"time_label"`
	GanZhi               string                `json:"gan_zhi,omitempty"`
	Summary              string                `json:"summary"`
	Method               []string              `json:"method"`
	Highlights           []PeriodHighlight     `json:"highlights"`
	FocusPalaces         []PeriodPalaceFocus   `json:"focus_palaces"`
	Evidence             []PeriodEvidence      `json:"evidence"`
	CrossLayerRelations  []PeriodLayerRelation `json:"cross_layer_relations"`
	ReviewNotes          []string              `json:"review_notes"`
	Limitations          []string              `json:"limitations"`
	DayunStages          []DayunStageAnalysis  `json:"dayun_stages,omitempty"`
	DayunContext         *DayunStageAnalysis   `json:"dayun_context,omitempty"`
	EvidenceBasis        string                `json:"evidence_basis"`
	PlacementBasis       string                `json:"placement_basis"`
	InterpretationBasis  string                `json:"interpretation_basis"`
	InterpretationStatus string                `json:"interpretation_status"`
	ValidationStatus     string                `json:"validation_status"`
	IsOutcomeConclusion  bool                  `json:"is_outcome_conclusion"`
}

type PeriodEvidenceSemantics struct {
	PlacementBasis       string `json:"placement_basis"`
	InterpretationBasis  string `json:"interpretation_basis"`
	InterpretationStatus string `json:"interpretation_status"`
	IsOutcomeConclusion  bool   `json:"is_outcome_conclusion"`
}

type DayunStageAnalysis struct {
	PeriodEvidenceSemantics
	StartAge      int      `json:"start_age"`
	EndAge        int      `json:"end_age"`
	Palace        string   `json:"palace"`
	Branch        string   `json:"branch"`
	HeavenlyStem  string   `json:"heavenly_stem"`
	EarthlyBranch string   `json:"earthly_branch"`
	GanZhi        string   `json:"gan_zhi"`
	MainStars     []string `json:"main_stars"`
	AuxStars      []string `json:"aux_stars"`
	PeriodStars   []string `json:"period_stars"`
	FourHua       []string `json:"four_hua"`
	Sanfang       []string `json:"sanfang"`
	Summary       string   `json:"summary"`
	ReviewNotes   []string `json:"review_notes"`
	Current       bool     `json:"current"`
	NominalAge    int      `json:"nominal_age,omitempty"`
	AgeBasis      string   `json:"age_basis,omitempty"`
}

type PeriodHighlight struct {
	PeriodEvidenceSemantics
	Label string `json:"label"`
	Value string `json:"value"`
	Note  string `json:"note"`
}

type PeriodPalaceFocus struct {
	PeriodEvidenceSemantics
	Palace       string   `json:"palace"`
	PeriodPalace string   `json:"period_palace"`
	Branch       string   `json:"branch"`
	MainStars    []string `json:"main_stars"`
	AuxStars     []string `json:"aux_stars"`
	PeriodStars  []string `json:"period_stars"`
	FourHua      []string `json:"four_hua"`
	Sanfang      []string `json:"sanfang"`
	Reason       string   `json:"reason"`
	ReviewNote   string   `json:"review_note"`
}

type PeriodEvidence struct {
	PeriodEvidenceSemantics
	Type  string `json:"type"`
	Label string `json:"label"`
	Value string `json:"value"`
	Basis string `json:"basis"`
}

// PeriodLayerRelation records deterministic branch interaction between
// nested transit layers. It does not assign strength, favorability, or events.
type PeriodLayerRelation struct {
	SourceLayer          string `json:"source_layer"`
	SourceGanZhi         string `json:"source_gan_zhi"`
	SourceBranch         string `json:"source_branch"`
	TargetLayer          string `json:"target_layer"`
	TargetGanZhi         string `json:"target_gan_zhi"`
	TargetBranch         string `json:"target_branch"`
	Relation             string `json:"relation"`
	Subtype              string `json:"subtype,omitempty"`
	RuleID               string `json:"rule_id"`
	StructuralStatus     string `json:"structural_status"`
	TransformationStatus string `json:"transformation_status"`
	TargetElement        string `json:"target_element,omitempty"`
	EvidenceBasis        string `json:"evidence_basis"`
	InterpretationStatus string `json:"interpretation_status"`
	IsOutcomeConclusion  bool   `json:"is_outcome_conclusion"`
}

func BuildDayunAnalysis(chart *ZiWeiChart, stages Dayun, currentAge int) *PeriodAnalysis {
	if !chartMatchesDeclaredProfile(chart) {
		return nil
	}
	expectedStages := calcDayunFromChart(chart)
	out := basePeriodAnalysis("dayun", "大限分析", "十年阶段")
	out.Method = []string{
		"以五行局数作为起限年龄。",
		"按阳男阴女顺行、阴男阳女逆行，从命宫布十二个十年阶段。",
		"以大限宫干支另布大限运曜，并由大限天干引动该限四化。",
	}

	if len(stages) == 0 {
		stages = expectedStages
	} else if !reflect.DeepEqual(stages, expectedStages) {
		return nil
	}

	for _, stage := range stages {
		idx := palaceIndexByName(chart, stage.Palace)
		if idx < 0 {
			continue
		}
		p := chart.Palaces[idx]
		stem, stemOK := StemIndex[stage.HeavenlyStem]
		branch, branchOK := BranchIndex[stage.EarthlyBranch]
		if !stemOK || !branchOK {
			continue
		}
		periodStars := buildTransitStarDistribution(chart, stem, branch, "dayun")
		periodFourHua := buildTransitFourHua(chart, stem)
		item := DayunStageAnalysis{
			PeriodEvidenceSemantics: periodEvidenceSemantics(),
			StartAge:                stage.StartAge,
			EndAge:                  stage.EndAge,
			Palace:                  p.Name,
			Branch:                  p.Branch,
			HeavenlyStem:            stage.HeavenlyStem,
			EarthlyBranch:           stage.EarthlyBranch,
			GanZhi:                  stage.GanZhi,
			MainStars:               palaceMainStars(p),
			AuxStars:                palaceAuxStars(p),
			PeriodStars:             flattenPeriodStars(periodStars),
			FourHua:                 flattenPeriodStars(periodFourHua),
			Sanfang:                 cloneStrings(sanfangNames(chart, idx)),
			Current:                 currentAge >= stage.StartAge && currentAge <= stage.EndAge,
		}
		item.Summary = buildStageSummary(item)
		item.ReviewNotes = stageReviewNotes(item)
		out.DayunStages = append(out.DayunStages, item)
	}

	out.Summary = "大限按十年阶段展示限宫、限干支、运曜与大限四化；大限四化与本命四化分开计算。"

	current := currentDayunStage(out.DayunStages)
	if current != nil {
		out.GanZhi = current.GanZhi
		out.Highlights = append(out.Highlights,
			newPeriodHighlight("当前大限", fmt.Sprintf("%d-%d岁", current.StartAge, current.EndAge), current.Palace+"为当前年龄对应的十年阶段宫位"),
			newPeriodHighlight("大限干支", current.GanZhi, "以限宫宫干支引动本限四化与运曜"),
			newPeriodHighlight("大限四化", strings.Join(current.FourHua, "、"), "由大限天干引动，不与本命四化混用"),
		)
		stem := StemIndex[current.HeavenlyStem]
		branch := BranchIndex[current.EarthlyBranch]
		periodStars := buildTransitStarDistribution(chart, stem, branch, "dayun")
		periodFourHua := buildTransitFourHua(chart, stem)
		periodPalaces, _ := buildDayunPalaceNames(chart, branch)
		out.FocusPalaces = periodPalaceFocus(chart, chart, periodStars, periodFourHua, periodPalaces)
	}
	out.Evidence = append(out.Evidence,
		newPeriodEvidence("start_age", "起限", dayunStartAgeLabel(chart), "五行局数决定第一步大限起点"),
		newPeriodEvidence("direction", "行运方向", dayunDirectionLabel(chart), "决定十二宫顺排或逆排"),
	)
	if current != nil {
		out.Evidence = append(out.Evidence,
			newPeriodEvidence("dayun_ganzhi", "大限干支", current.GanZhi, "取当前大限宫位的宫干支"),
			newPeriodEvidence("dayun_four_hua", "大限四化", strings.Join(current.FourHua, "、"), "由当前大限天干引动并按本命星曜落宫"),
			newPeriodEvidence("dayun_stars", "大限运曜", strings.Join(current.PeriodStars, "、"), "由当前大限干支安运魁钺昌曲禄羊陀马鸾喜"),
		)
	}
	out.ReviewNotes = []string{
		"先核对当前大限宫位与限干支，再叠加流年触发宫位。",
		"大限四化只使用大限天干计算，本命四化保留在本命盘层。",
	}
	out.Limitations = []string{
		"大限规则未进入独立 Gold 裁决，不能作为现实决策依据。",
		"煞曜与化忌只作为传统结构标签，不证明现实风险。",
	}
	return out
}

func BuildLiunianAnalysis(base, period *ZiWeiChart, year int) *PeriodAnalysis {
	interpreter, stem, branch, ok := prepareTransitAnalysis(base, period, "liunian", year, 0, 0)
	if !ok {
		return nil
	}
	interp := interpreter.AnalyzeLiunian(period, year)
	if interp == nil {
		return nil
	}
	out := buildTransitAnalysis(base, period, transitConfig{
		Layer:         "liunian",
		Title:         "流年分析",
		TimeLabel:     fmt.Sprintf("%d年", year),
		GanZhi:        stemName(stem) + branchName(branch),
		Stem:          stem,
		Branch:        branch,
		PeriodStars:   period.LiuNianStars,
		PeriodFourHua: period.LiuNianFourHua,
		PeriodPalaces: period.LiuNianPalaces,
		Interpreter:   interp,
		Method: []string{
			"按目标年份取流年干支。",
			"用流年天干引动四化，并安流禄、流羊、流陀、流马。",
			"逐项展示触发宫位、命局刑冲合、十神与宫位星曜结构。",
		},
	})
	attachDayunContext(out, base, period.DerivationInput.ResolvedLunarDate.Year, "liunian", stemName(stem)+branchName(branch), branch)
	return out
}

func BuildLiuyueAnalysis(base, period *ZiWeiChart, year, month, day int) *PeriodAnalysis {
	interpreter, stem, branch, ok := prepareTransitAnalysis(base, period, "liuyue", year, month, day)
	if !ok {
		return nil
	}
	interp := interpreter.AnalyzeLiuyue(period, year, month, day)
	if interp == nil {
		return nil
	}
	out := buildTransitAnalysis(base, period, transitConfig{
		Layer:         "liuyue",
		Title:         "流月分析",
		TimeLabel:     fmt.Sprintf("%d年%d月%d日所在农历月", year, month, day),
		GanZhi:        stemName(stem) + branchName(branch),
		Stem:          stem,
		Branch:        branch,
		PeriodStars:   period.LiuYueStars,
		PeriodFourHua: period.LiuYueFourHua,
		PeriodPalaces: period.LiuYuePalaces,
		Interpreter:   interp,
		Method: []string{
			"将目标公历日期转换为农历日期，以农历初一作为流月分界。",
			"用流月天干引动当月四化。",
			"逐项展示当月触发宫位、流年层和命局关系。",
		},
	})
	out.CrossLayerRelations = periodCrossLayerRelations("liuyue", *period.DerivationInput)
	attachDayunContext(out, base, period.DerivationInput.ResolvedLunarDate.Year, "liuyue", stemName(stem)+branchName(branch), branch)
	return out
}

func BuildLiuriAnalysis(base, period *ZiWeiChart, year, month, day int) *PeriodAnalysis {
	interpreter, stem, branch, ok := prepareTransitAnalysis(base, period, "liuri", year, month, day)
	if !ok {
		return nil
	}
	interp := interpreter.AnalyzeLiuri(period, year, month, day)
	if interp == nil {
		return nil
	}
	out := buildTransitAnalysis(base, period, transitConfig{
		Layer:         "liuri",
		Title:         "流日分析",
		TimeLabel:     fmt.Sprintf("%d年%d月%d日", year, month, day),
		GanZhi:        stemName(stem) + branchName(branch),
		Stem:          stem,
		Branch:        branch,
		PeriodStars:   period.LiuRiStars,
		PeriodFourHua: period.LiuRiFourHua,
		PeriodPalaces: period.LiuRiPalaces,
		Interpreter:   interp,
		Method: []string{
			"按日期推流日干支。",
			"用流日天干引动日内四化。",
			"逐项展示当天触发宫位与十二时辰干支、十神、地支关系。",
		},
	})
	out.CrossLayerRelations = periodCrossLayerRelations("liuri", *period.DerivationInput)
	attachDayunContext(out, base, period.DerivationInput.ResolvedLunarDate.Year, "liuri", stemName(stem)+branchName(branch), branch)
	return out
}

func prepareTransitAnalysis(base, period *ZiWeiChart, layer string, year, month, day int) (*PeriodInterpreter, int, int, bool) {
	interpreter := NewPeriodInterpreterFromChart(base)
	if interpreter == nil || period == nil || !DerivedChartMatchesBase(period, base) {
		return nil, 0, 0, false
	}
	stem, branch, ok := chartDerivationForQuery(period, layer, year, month, day)
	if !ok {
		return nil, 0, 0, false
	}
	return interpreter, stem, branch, true
}

func dayunContextForLunarYear(chart *ZiWeiChart, lunarYear int) *DayunStageAnalysis {
	birth, ok := birthDataFromPublishedChart(chart)
	if !ok {
		return nil
	}
	nominalAge := lunarYear - birth.LunarYear + 1
	dayun := BuildDayunAnalysis(chart, nil, nominalAge)
	if dayun == nil {
		return nil
	}
	current := currentDayunStage(dayun.DayunStages)
	if current == nil {
		return nil
	}
	context := *current
	context.NominalAge = nominalAge
	context.AgeBasis = "target_lunar_year_minus_birth_lunar_year_plus_one"
	return &context
}

func attachDayunContext(out *PeriodAnalysis, chart *ZiWeiChart, lunarYear int, sourceLayer, sourceGanZhi string, sourceBranch int) {
	if out == nil {
		return
	}
	context := dayunContextForLunarYear(chart, lunarYear)
	if context == nil {
		return
	}
	out.DayunContext = context
	out.Highlights = append(out.Highlights,
		newPeriodHighlight("所处大限", fmt.Sprintf("%s %d-%d岁", context.GanZhi, context.StartAge, context.EndAge), context.Palace+"为目标农历年对应的大限宫位"),
	)
	out.Evidence = append(out.Evidence,
		newPeriodEvidence("dayun_context", "大限上下文", fmt.Sprintf("虚岁%d · %s", context.NominalAge, context.GanZhi), "先按目标农历年确定大限，再叠加本层触发"),
	)
	out.CrossLayerRelations = append(out.CrossLayerRelations, dayunCrossLayerRelations(sourceLayer, sourceGanZhi, sourceBranch, context)...)
}

func dayunCrossLayerRelations(sourceLayer, sourceGanZhi string, sourceBranch int, context *DayunStageAnalysis) []PeriodLayerRelation {
	if context == nil {
		return []PeriodLayerRelation{}
	}
	targetBranch, ok := BranchIndex[context.EarthlyBranch]
	if !ok {
		return []PeriodLayerRelation{}
	}
	rules := periodPairRelationRules(sourceBranch, targetBranch)
	out := make([]PeriodLayerRelation, 0, len(rules))
	for _, rule := range rules {
		out = append(out, PeriodLayerRelation{
			SourceLayer: sourceLayer, SourceGanZhi: sourceGanZhi, SourceBranch: branchName(sourceBranch),
			TargetLayer: "dayun", TargetGanZhi: context.GanZhi, TargetBranch: context.EarthlyBranch,
			Relation: rule.relation, Subtype: rule.subtype, RuleID: rule.ruleID,
			StructuralStatus: rule.structuralStatus, TransformationStatus: rule.transformationStatus,
			TargetElement: rule.targetElement, EvidenceBasis: periodPlacementBasis,
			InterpretationStatus: periodInterpretationStatus, IsOutcomeConclusion: false,
		})
	}
	return out
}

type transitConfig struct {
	Layer         string
	Title         string
	TimeLabel     string
	GanZhi        string
	Stem          int
	Branch        int
	PeriodStars   [12][]string
	PeriodFourHua [12][]string
	PeriodPalaces [12]string
	Interpreter   any
	Method        []string
}

func buildTransitAnalysis(base, period *ZiWeiChart, cfg transitConfig) *PeriodAnalysis {
	out := basePeriodAnalysis(cfg.Layer, cfg.Title, cfg.TimeLabel)
	out.GanZhi = cfg.GanZhi
	out.Method = cfg.Method

	periodStarNames := flattenPeriodStars(cfg.PeriodStars)
	periodFourHuaNames := flattenPeriodStars(cfg.PeriodFourHua)
	triggered := periodPalaceFocus(base, period, cfg.PeriodStars, cfg.PeriodFourHua, cfg.PeriodPalaces)
	out.FocusPalaces = triggered
	out.Summary = transitSummary(cfg.Layer, cfg.TimeLabel, cfg.GanZhi, triggered, cfg.Interpreter)
	out.Highlights = transitHighlights(cfg, triggered, periodStarNames, periodFourHuaNames)
	out.Evidence = transitEvidence(base, cfg, triggered, periodStarNames, periodFourHuaNames)
	out.ReviewNotes = transitReviewNotes(cfg.Layer, triggered, cfg.Interpreter)
	out.Limitations = transitLimitations(cfg.Layer, triggered, cfg.Interpreter)
	return out
}

func basePeriodAnalysis(layer, title, timeLabel string) *PeriodAnalysis {
	return &PeriodAnalysis{
		RuleVersion:          PeriodRuleVersion,
		School:               PeriodSchool,
		Layer:                layer,
		Title:                title,
		TimeLabel:            timeLabel,
		Method:               []string{},
		Highlights:           []PeriodHighlight{},
		FocusPalaces:         []PeriodPalaceFocus{},
		Evidence:             []PeriodEvidence{},
		CrossLayerRelations:  []PeriodLayerRelation{},
		ReviewNotes:          []string{},
		Limitations:          []string{},
		EvidenceBasis:        periodMixedEvidenceBasis,
		PlacementBasis:       periodPlacementBasis,
		InterpretationBasis:  periodInterpretationBasis,
		InterpretationStatus: periodInterpretationStatus,
		ValidationStatus:     periodInterpretationStatus,
		IsOutcomeConclusion:  false,
	}
}

func periodCrossLayerRelations(layer string, source ZiWeiDerivationInput) []PeriodLayerRelation {
	targets := make([]struct {
		layer string
		input ZiWeiDerivationInput
	}, 0, 2)

	yearInput, yearErr := buildZiWeiDerivationInput("liunian", source.ResolvedLunarDate.Year, 0, 0)
	switch layer {
	case "liuyue":
		if yearErr == nil {
			targets = append(targets, struct {
				layer string
				input ZiWeiDerivationInput
			}{layer: "liunian", input: yearInput})
		}
	case "liuri":
		monthInput, monthErr := buildZiWeiDerivationInput("liuyue", source.Year, source.Month, source.Day)
		if monthErr == nil {
			targets = append(targets, struct {
				layer string
				input ZiWeiDerivationInput
			}{layer: "liuyue", input: monthInput})
		}
		if yearErr == nil {
			targets = append(targets, struct {
				layer string
				input ZiWeiDerivationInput
			}{layer: "liunian", input: yearInput})
		}
	default:
		return []PeriodLayerRelation{}
	}

	_, sourceBranch, ok := derivationStemBranch(source)
	if !ok {
		return []PeriodLayerRelation{}
	}
	out := make([]PeriodLayerRelation, 0, 4)
	for _, target := range targets {
		_, targetBranch, valid := derivationStemBranch(target.input)
		if !valid {
			continue
		}
		for _, rule := range periodPairRelationRules(sourceBranch, targetBranch) {
			out = append(out, PeriodLayerRelation{
				SourceLayer: layer, SourceGanZhi: source.PeriodGanZhi, SourceBranch: branchName(sourceBranch),
				TargetLayer: target.layer, TargetGanZhi: target.input.PeriodGanZhi, TargetBranch: branchName(targetBranch),
				Relation: rule.relation, Subtype: rule.subtype, RuleID: rule.ruleID,
				StructuralStatus: rule.structuralStatus, TransformationStatus: rule.transformationStatus,
				TargetElement: rule.targetElement, EvidenceBasis: periodPlacementBasis,
				InterpretationStatus: periodInterpretationStatus, IsOutcomeConclusion: false,
			})
		}
	}
	return out
}

func periodPalaceFocus(base, period *ZiWeiChart, periodStars, periodFourHua [12][]string, periodPalaces [12]string) []PeriodPalaceFocus {
	out := []PeriodPalaceFocus{}
	for i := range period.Palaces {
		p := period.Palaces[i]
		stars := append([]string(nil), periodStars[i]...)
		fourHua := append([]string(nil), periodFourHua[i]...)
		if len(stars) == 0 && len(fourHua) == 0 {
			continue
		}
		focus := PeriodPalaceFocus{
			PeriodEvidenceSemantics: periodEvidenceSemantics(),
			Palace:                  p.Name,
			PeriodPalace:            periodPalaces[i],
			Branch:                  p.Branch,
			MainStars:               palaceMainStars(p),
			AuxStars:                palaceAuxStars(p),
			PeriodStars:             stars,
			FourHua:                 fourHua,
			Sanfang:                 cloneStrings(sanfangNames(base, i)),
			Reason:                  palaceFocusReason(p, periodPalaces[i], stars, fourHua),
			ReviewNote:              palaceFocusReviewNote(p, stars),
		}
		out = append(out, focus)
	}
	return out
}

func transitSummary(layer, timeLabel, ganZhi string, focus []PeriodPalaceFocus, interpreter any) string {
	focusText := "无流曜或本层四化落点"
	if len(focus) > 0 {
		names := make([]string, 0, len(focus))
		for _, item := range focus {
			names = append(names, item.Palace)
		}
		focusText = strings.Join(names, "、")
	}
	prefix := fmt.Sprintf("%s%s%s。触发宫位：%s。", timeLabel, optionalGanZhi(ganZhi), layerVerb(layer), focusText)
	switch v := interpreter.(type) {
	case *LiunianResult:
		if v != nil {
			return prefix + v.StructuralSummary
		}
	case *LiuyueResult:
		if v != nil {
			return prefix + v.StructuralSummary
		}
	case *LiuriResult:
		if v != nil {
			return prefix + v.StructuralSummary
		}
	}
	return prefix
}

func transitHighlights(cfg transitConfig, focus []PeriodPalaceFocus, periodStars, periodFourHua []string) []PeriodHighlight {
	items := []PeriodHighlight{
		newPeriodHighlight("干支", cfg.GanZhi, fmt.Sprintf("天干%s引动四化，地支%s参与命局关系", stemName(cfg.Stem), branchName(cfg.Branch))),
	}
	if len(periodStars) > 0 {
		items = append(items, newPeriodHighlight("流曜", strings.Join(periodStars[:min(4, len(periodStars))], "、"), "优先看这些星曜落入的宫位"))
	}
	if len(periodFourHua) > 0 {
		items = append(items, newPeriodHighlight("本层四化", strings.Join(periodFourHua, "、"), "由本层天干引动，不与本命四化混用"))
	}
	if len(focus) > 0 {
		names := make([]string, 0, len(focus))
		for _, item := range focus {
			names = append(names, item.Palace)
		}
		items = append(items, newPeriodHighlight("触发宫位", strings.Join(names, "、"), "按派生盘十二宫顺序列出，不做强弱排序"))
	}
	return items
}

func transitEvidence(base *ZiWeiChart, cfg transitConfig, focus []PeriodPalaceFocus, periodStars, periodFourHua []string) []PeriodEvidence {
	var out []PeriodEvidence
	out = append(out,
		newPeriodEvidence("ganzhi", "周期干支", cfg.GanZhi, "决定十神、五行和本层四化"),
		newPeriodEvidence("relation", "命局关系", relationForBranch(base, cfg.Branch), "记录刑冲合伏吟结构，不推导事件结果"),
	)
	if len(periodStars) > 0 {
		out = append(out, newPeriodEvidence("period_star", "本层流曜", strings.Join(periodStars, "、"), "记录流曜落宫位置，不推导事件结果"))
	}
	if len(periodFourHua) > 0 {
		out = append(out, newPeriodEvidence("period_four_hua", "本层四化", strings.Join(periodFourHua, "、"), "记录周期天干引动的四化，不替代本命四化"))
	}
	for _, item := range focus {
		value := item.Branch
		if item.PeriodPalace != "" {
			value += "·本层" + item.PeriodPalace
		}
		out = append(out, newPeriodEvidence("focus_palace", item.Palace, value, item.Reason))
	}
	return out
}

func periodEvidenceSemantics() PeriodEvidenceSemantics {
	return PeriodEvidenceSemantics{
		PlacementBasis:       periodPlacementBasis,
		InterpretationBasis:  periodInterpretationBasis,
		InterpretationStatus: periodInterpretationStatus,
		IsOutcomeConclusion:  false,
	}
}

func newPeriodHighlight(label, value, note string) PeriodHighlight {
	return PeriodHighlight{
		PeriodEvidenceSemantics: periodEvidenceSemantics(),
		Label:                   label,
		Value:                   value,
		Note:                    note,
	}
}

func newPeriodEvidence(evidenceType, label, value, basis string) PeriodEvidence {
	return PeriodEvidence{
		PeriodEvidenceSemantics: periodEvidenceSemantics(),
		Type:                    evidenceType,
		Label:                   label,
		Value:                   value,
		Basis:                   basis,
	}
}

func transitReviewNotes(layer string, focus []PeriodPalaceFocus, interpreter any) []string {
	out := []string{"触发宫位按派生盘十二宫顺序展示，不换算为吉凶、概率或强弱分数。"}
	if len(focus) > 0 {
		top := focus[0]
		out = append(out, fmt.Sprintf("%s：%s", top.Palace, top.ReviewNote))
	}
	switch v := interpreter.(type) {
	case *LiunianResult:
		if v != nil && v.ReviewNote != "" {
			out = append(out, v.ReviewNote)
		}
	case *LiuyueResult:
		if v != nil && v.StructuralSummary != "" {
			out = append(out, v.StructuralSummary)
		}
	case *LiuriResult:
		if v != nil && v.StructuralSummary != "" {
			out = append(out, v.StructuralSummary)
		}
	}
	if layer == "liuri" {
		out = append(out, "十二时辰只展示干支、十神和地支关系，不构成吉凶或行动建议。")
	}
	return periodUniqueStrings(out)
}

func transitLimitations(layer string, focus []PeriodPalaceFocus, interpreter any) []string {
	out := []string{"周期解释未进入独立 Gold 裁决，不能作为职业、财务或人生决定依据。"}
	for _, item := range focus {
		if containsTransitConstraint(item.PeriodStars) || containsHuaJi(item.FourHua) {
			out = append(out, fmt.Sprintf("%s见传统约束类标签，但不证明现实阻滞或损失。", item.Palace))
		}
	}
	switch v := interpreter.(type) {
	case *LiunianResult:
		if v != nil && strings.Contains(v.RelationToMing, "冲") {
			out = append(out, "流年见冲只证明地支关系结构，不证明年度变动。")
		}
	case *LiuyueResult:
		if v != nil && strings.Contains(v.RelationToMing, "刑") {
			out = append(out, "流月见刑只证明地支关系结构，不证明沟通或手续结果。")
		}
	case *LiuriResult:
		if v != nil && strings.Contains(v.RelationToMing, "伏吟") {
			out = append(out, "流日伏吟只证明地支关系结构，不推导个体状态或决策时点。")
		}
	}
	return periodUniqueStrings(out)
}

func containsTransitConstraint(stars []string) bool {
	for _, star := range stars {
		if strings.HasSuffix(star, "羊") || strings.HasSuffix(star, "陀") {
			return true
		}
	}
	return false
}

func buildStageSummary(item DayunStageAnalysis) string {
	return fmt.Sprintf("%s大限落%s%s，对应%s；主星%s，大限四化%s。", item.GanZhi, item.Palace, formatBranch(item.Branch), palaceFocus(item.Palace), joinOrNone(item.MainStars), joinOrNone(item.FourHua))
}

func stageReviewNotes(item DayunStageAnalysis) []string {
	out := []string{fmt.Sprintf("%s仅作为该十年阶段的宫位结构索引，不推导人生结果。", item.Palace)}
	if containsHuaJi(item.FourHua) {
		out = append(out, "本阶段见传统约束类标签，但不证明现实风险或结果。")
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

func palaceFocusReason(p PalaceInfo, periodPalace string, periodStars, periodFourHua []string) string {
	label := p.Name
	if periodPalace != "" {
		label = fmt.Sprintf("%s（本层%s）", p.Name, periodPalace)
	}
	if len(periodStars) > 0 {
		return fmt.Sprintf("%s见%s，被本层流曜直接触发。", label, strings.Join(periodStars, "、"))
	}
	if len(periodFourHua) > 0 {
		return fmt.Sprintf("%s本层四化为%s，记录为本层结构信号。", label, strings.Join(periodFourHua, "、"))
	}
	return fmt.Sprintf("%s记录本命主星%s与辅星%s。", label, joinOrNone(palaceMainStars(p)), joinOrNone(palaceAuxStars(p)))
}

func palaceFocusReviewNote(p PalaceInfo, periodStars []string) string {
	focus := palaceFocus(p.Name)
	if hasTransitStarSuffix(periodStars, "马") {
		return fmt.Sprintf("%s见传统移动类标签；不据此推导现实变动或行动方案。", focus)
	}
	return fmt.Sprintf("%s只作为本层结构主题，不据此推导现实结果。", focus)
}

func hasTransitStarSuffix(stars []string, suffix string) bool {
	for _, star := range stars {
		if strings.HasSuffix(star, suffix) {
			return true
		}
	}
	return false
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
	interpreter := NewPeriodInterpreterFromChart(chart)
	if interpreter == nil {
		return "未恢复命局参数"
	}
	return interpreter.describeRelation(branch)
}

func dayunDirectionLabel(chart *ZiWeiChart) string {
	params, ok := dayunParametersFromChart(chart)
	if !ok {
		return "未知"
	}
	forward := isForwardByYearStem(params.yearStem, params.gender)
	if forward {
		return "顺行"
	}
	return "逆行"
}

func dayunStartAgeLabel(chart *ZiWeiChart) string {
	params, ok := dayunParametersFromChart(chart)
	if !ok {
		return "未知"
	}
	return fmt.Sprintf("%d岁", params.juValue)
}

func layerVerb(layer string) string {
	switch layer {
	case "liunian":
		return "记录年度结构"
	case "liuyue":
		return "记录月度结构"
	case "liuri":
		return "记录流日结构"
	default:
		return "记录周期结构"
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

func fixMod(v, mod int) int {
	out := v % mod
	if out < 0 {
		out += mod
	}
	return out
}
