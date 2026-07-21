package ziwei

import (
	"fmt"
	"strings"
)

// LiunianOverlayAnalysis explains how an annual overlay was derived and which
// palaces are most affected.
type LiunianOverlayAnalysis struct {
	Year                 int                    `json:"year"`
	GanZhi               string                 `json:"gan_zhi"`
	Stem                 string                 `json:"stem"`
	Branch               string                 `json:"branch"`
	ShiShen              string                 `json:"shi_shen,omitempty"`
	RelationToMing       string                 `json:"relation_to_ming"`
	RelationEvidence     []PeriodBranchRelation `json:"relation_evidence"`
	ReviewNote           string                 `json:"review_note"`
	Summary              string                 `json:"summary"`
	Method               []OverlayMethodStep    `json:"method"`
	FourHua              []OverlayTrigger       `json:"four_hua"`
	AnnualStars          []OverlayTrigger       `json:"annual_stars"`
	FocusPalaces         []OverlayFocusPalace   `json:"focus_palaces"`
	DayunContext         *DayunStageAnalysis    `json:"dayun_context,omitempty"`
	EvidenceBasis        string                 `json:"evidence_basis"`
	PlacementBasis       string                 `json:"placement_basis"`
	InterpretationBasis  string                 `json:"interpretation_basis"`
	InterpretationStatus string                 `json:"interpretation_status"`
	ValidationStatus     string                 `json:"validation_status"`
	IsOutcomeConclusion  bool                   `json:"is_outcome_conclusion"`
}

type OverlayMethodStep struct {
	PeriodEvidenceSemantics
	Label   string `json:"label"`
	Value   string `json:"value"`
	Meaning string `json:"meaning"`
}

type OverlayTrigger struct {
	PeriodEvidenceSemantics
	Type     string `json:"type"`
	Star     string `json:"star,omitempty"`
	Palace   string `json:"palace"`
	Branch   string `json:"branch"`
	Meaning  string `json:"meaning"`
	Polarity string `json:"polarity"`
}

type OverlayFocusPalace struct {
	PeriodEvidenceSemantics
	Palace     string           `json:"palace"`
	Branch     string           `json:"branch"`
	Triggers   []OverlayTrigger `json:"triggers"`
	MainStars  []string         `json:"main_stars"`
	ReviewNote string           `json:"review_note"`
}

// AnalyzeLiunianOverlay combines annual star placement, annual four-hua and
// period interpretation into a structured explanation for the frontend.
func (s *ZiWeiService) AnalyzeLiunianOverlay(base *ZiWeiChart, liunian *ZiWeiChart, year int) *LiunianOverlayAnalysis {
	interpreter := NewPeriodInterpreterFromChart(base)
	if interpreter == nil {
		return nil
	}
	if liunian == nil {
		liunian = s.CalculateLiunian(base, year)
	}
	if liunian == nil || !DerivedChartMatchesBase(liunian, base) {
		return nil
	}
	chart := liunian

	stemIdx, branchIdx := annualStemBranch(year)
	stem := StemNames[stemIdx]
	branch := BranchNames[branchIdx]
	ganZhi := stem + branch

	result := interpreter.AnalyzeLiunian(chart, year)
	if result == nil {
		return nil
	}
	shiShen := result.ShiShen
	ganZhi = result.GanZhi

	fourHua := buildLiunianFourHuaTriggers(chart, stemIdx)
	annualStars := buildAnnualStarTriggers(chart, stemIdx, branchIdx)
	allTriggers := append(append([]OverlayTrigger{}, fourHua...), annualStars...)
	focus := buildOverlayFocusPalaces(chart, allTriggers)

	method := []OverlayMethodStep{
		{
			PeriodEvidenceSemantics: periodEvidenceSemantics(),
			Label:                   "流年干支",
			Value:                   fmt.Sprintf("%d年 %s", year, ganZhi),
			Meaning:                 "用目标年份推得流年天干地支，作为本年叠盘的时间层。",
		},
		{
			PeriodEvidenceSemantics: periodEvidenceSemantics(),
			Label:                   "天干四化",
			Value:                   describeFourHuaByStem(stemIdx),
			Meaning:                 "流年天干决定化禄、化权、化科、化忌，落到含对应星曜的本命宫位。",
		},
		{
			PeriodEvidenceSemantics: periodEvidenceSemantics(),
			Label:                   "禄羊陀马",
			Value:                   describeAnnualStarPositions(chart, stemIdx, branchIdx),
			Meaning:                 "记录流禄、流羊、流陀、流马的落宫及传统分类，不据此推导现实结果。",
		},
		{
			PeriodEvidenceSemantics: periodEvidenceSemantics(),
			Label:                   "三方四正",
			Value:                   "悬停宫位可查看本宫、对宫、三合宫关系",
			Meaning:                 "年度触发宫位连同对宫和三合宫作为完整结构展示。",
		},
	}

	return &LiunianOverlayAnalysis{
		Year:                 year,
		GanZhi:               ganZhi,
		Stem:                 stem,
		Branch:               branch,
		ShiShen:              shiShen,
		RelationToMing:       result.RelationToMing,
		RelationEvidence:     append([]PeriodBranchRelation{}, result.RelationEvidence...),
		ReviewNote:           result.ReviewNote,
		Summary:              buildOverlaySummary(year, ganZhi, focus),
		Method:               method,
		FourHua:              fourHua,
		AnnualStars:          annualStars,
		FocusPalaces:         focus,
		DayunContext:         dayunContextForLunarYear(base, year),
		EvidenceBasis:        periodMixedEvidenceBasis,
		PlacementBasis:       periodPlacementBasis,
		InterpretationBasis:  periodInterpretationBasis,
		InterpretationStatus: periodInterpretationStatus,
		ValidationStatus:     periodInterpretationStatus,
		IsOutcomeConclusion:  false,
	}
}

func annualStemBranch(year int) (int, int) {
	stem := (year - 4) % 10
	if stem < 0 {
		stem += 10
	}
	branch := (year - 4) % 12
	if branch < 0 {
		branch += 12
	}
	return stem, branch
}

func buildLiunianFourHuaTriggers(chart *ZiWeiChart, stemIdx int) []OverlayTrigger {
	hua := SiHuaTable[stemIdx]
	labels := SiHuaLabels
	triggers := make([]OverlayTrigger, 0, 4)
	for i, star := range hua {
		label := labels[i]
		palace := findPalaceWithStar(chart, star)
		if palace == nil {
			continue
		}
		triggers = append(triggers, OverlayTrigger{
			PeriodEvidenceSemantics: periodEvidenceSemantics(),
			Type:                    label,
			Star:                    star,
			Palace:                  palace.Name,
			Branch:                  palace.Branch,
			Meaning:                 fourHuaMeaning(label, star, palace.Name),
			Polarity:                fourHuaPolarity(label),
		})
	}
	return triggers
}

func buildAnnualStarTriggers(chart *ZiWeiChart, stemIdx, branchIdx int) []OverlayTrigger {
	distribution := buildTransitStarDistribution(chart, stemIdx, branchIdx, "liunian")
	triggers := make([]OverlayTrigger, 0, 11)
	for i, stars := range distribution {
		for _, star := range stars {
			triggers = append(triggers, OverlayTrigger{
				PeriodEvidenceSemantics: periodEvidenceSemantics(),
				Type:                    star,
				Palace:                  chart.Palaces[i].Name,
				Branch:                  chart.Palaces[i].Branch,
				Meaning:                 annualStarMeaning(star, chart.Palaces[i].Name),
				Polarity:                transitStarPolarity(star),
			})
		}
	}
	return triggers
}

func transitStarPolarity(star string) string {
	switch {
	case strings.HasSuffix(star, "禄"):
		return "resource"
	case strings.HasSuffix(star, "羊"), strings.HasSuffix(star, "陀"):
		return "constraint"
	case strings.HasSuffix(star, "马"):
		return "movement"
	default:
		return "neutral"
	}
}

func buildOverlayFocusPalaces(chart *ZiWeiChart, triggers []OverlayTrigger) []OverlayFocusPalace {
	type group struct {
		palace   *PalaceInfo
		triggers []OverlayTrigger
	}
	groups := map[string]*group{}
	for _, trigger := range triggers {
		key := trigger.Palace + "|" + trigger.Branch
		g, ok := groups[key]
		if !ok {
			palace := findPalaceByNameBranch(chart, trigger.Palace, trigger.Branch)
			g = &group{palace: palace}
			groups[key] = g
		}
		g.triggers = append(g.triggers, trigger)
	}

	focus := make([]OverlayFocusPalace, 0, len(groups))
	for i := range chart.Palaces {
		palace := &chart.Palaces[i]
		g := groups[palace.Name+"|"+palace.Branch]
		if g == nil {
			continue
		}
		if g.palace == nil {
			continue
		}
		focus = append(focus, OverlayFocusPalace{
			PeriodEvidenceSemantics: periodEvidenceSemantics(),
			Palace:                  g.palace.Name,
			Branch:                  g.palace.Branch,
			Triggers:                cloneOverlayTriggers(g.triggers),
			MainStars:               palaceMainStarNames(*g.palace),
			ReviewNote:              focusReviewNote(g.palace.Name, g.triggers),
		})
	}
	return focus
}

func focusReviewNote(palace string, triggers []OverlayTrigger) string {
	hasConstraint := false
	hasResource := false
	hasMove := false
	for _, trigger := range triggers {
		switch trigger.Polarity {
		case "constraint":
			hasConstraint = true
		case "resource":
			hasResource = true
		case "movement":
			hasMove = true
		}
		if trigger.Type == "化忌" {
			hasConstraint = true
		}
	}
	switch {
	case hasConstraint && hasResource:
		return fmt.Sprintf("%s同时记录资源类与约束类传统标签；不据此推导现实机会或风险。", palace)
	case hasConstraint:
		return fmt.Sprintf("%s记录约束类传统标签；不据此推导现实阻滞、损失或个体状态。", palace)
	case hasResource && hasMove:
		return fmt.Sprintf("%s同时记录资源类与移动类传统标签；不据此推导现实机会。", palace)
	case hasResource:
		return fmt.Sprintf("%s记录资源类传统标签；不据此推导现实收益或助力。", palace)
	case hasMove:
		return fmt.Sprintf("%s记录移动类传统标签；不据此给出出行、迁动或职业建议。", palace)
	default:
		return fmt.Sprintf("%s记录流年触发结构，需连同本命星曜和三方四正展示。", palace)
	}
}

func buildOverlaySummary(year int, ganZhi string, focus []OverlayFocusPalace) string {
	if len(focus) == 0 {
		return fmt.Sprintf("%d年%s的叠盘重点在年度干支、四化和流曜本身；当前盘面未形成明显多重触发宫位。", year, ganZhi)
	}
	names := make([]string, 0, len(focus))
	for _, item := range focus {
		names = append(names, item.Palace)
	}
	return fmt.Sprintf("%d年%s叠盘记录的触发宫位为%s；这些宫位承接了流年四化或流禄羊陀马，按命盘十二宫顺序展示。", year, ganZhi, strings.Join(names, "、"))
}

func describeFourHuaByStem(stemIdx int) string {
	hua := SiHuaTable[stemIdx]
	parts := make([]string, 0, len(hua))
	for i, star := range hua {
		parts = append(parts, star+SiHuaLabels[i])
	}
	return StemNames[stemIdx] + "干：" + strings.Join(parts, "、")
}

func describeAnnualStarPositions(chart *ZiWeiChart, stemIdx, branchIdx int) string {
	parts := make([]string, 0, 4)
	for _, trigger := range buildAnnualStarTriggers(chart, stemIdx, branchIdx) {
		parts = append(parts, fmt.Sprintf("%s在%s%s", trigger.Type, trigger.Branch, trigger.Palace))
	}
	return strings.Join(parts, "、")
}

func fourHuaMeaning(label, star, palace string) string {
	switch label {
	case "化禄":
		return fmt.Sprintf("%s化禄落%s，记录资源与承接主题；具体结果未裁决。", star, palace)
	case "化权":
		return fmt.Sprintf("%s化权落%s，记录责任与执行主题；具体结果未裁决。", star, palace)
	case "化科":
		return fmt.Sprintf("%s化科落%s，记录名声、学习与凭证主题；具体结果未裁决。", star, palace)
	case "化忌":
		return fmt.Sprintf("%s化忌落%s，记录阻滞、执着与代价主题；具体结果未裁决。", star, palace)
	default:
		return fmt.Sprintf("%s%s落%s。", star, label, palace)
	}
}

func fourHuaPolarity(label string) string {
	switch label {
	case "化禄", "化科":
		return "resource"
	case "化忌":
		return "constraint"
	case "化权":
		return "neutral"
	default:
		return "neutral"
	}
}

func annualStarMeaning(label, palace string) string {
	switch label {
	case "流禄":
		return fmt.Sprintf("流禄落%s，记录传统资源类标签；具体收益或助力未裁决。", palace)
	case "流羊":
		return fmt.Sprintf("流羊落%s，记录传统约束类标签；具体冲突或个体状态未裁决。", palace)
	case "流陀":
		return fmt.Sprintf("流陀落%s，记录传统约束类标签；具体阻滞未裁决。", palace)
	case "流马":
		return fmt.Sprintf("流马落%s，记录传统移动类标签；具体出行或职业变化未裁决。", palace)
	default:
		return fmt.Sprintf("%s落%s。", label, palace)
	}
}

func findPalaceWithStar(chart *ZiWeiChart, star string) *PalaceInfo {
	if chart == nil {
		return nil
	}
	for i := range chart.Palaces {
		if palaceHasStar(chart.Palaces[i], star) {
			return &chart.Palaces[i]
		}
	}
	return nil
}

func findPalaceByBranch(chart *ZiWeiChart, branchIdx int) *PalaceInfo {
	if chart == nil {
		return nil
	}
	branch := BranchNames[fixIndex(branchIdx)]
	for i := range chart.Palaces {
		if chart.Palaces[i].Branch == branch {
			return &chart.Palaces[i]
		}
	}
	return nil
}

func findPalaceByNameBranch(chart *ZiWeiChart, palaceName, branch string) *PalaceInfo {
	if chart == nil {
		return nil
	}
	for i := range chart.Palaces {
		palace := &chart.Palaces[i]
		if palace.Name == palaceName && palace.Branch == branch {
			return palace
		}
	}
	return nil
}

func palaceHasStar(palace PalaceInfo, star string) bool {
	for _, name := range palaceStarNames(palace) {
		if name == star {
			return true
		}
	}
	return false
}

func palaceStarNames(palace PalaceInfo) []string {
	names := make([]string, 0, len(palace.Stars))
	for _, star := range palace.Stars {
		names = append(names, star.Name)
	}
	return overlayUniqueStrings(names)
}

func palaceMainStarNames(palace PalaceInfo) []string {
	names := make([]string, 0)
	for _, star := range palace.Stars {
		if star.Type == "major" {
			names = append(names, star.Name)
		}
	}
	return overlayUniqueStrings(names)
}

func overlayUniqueStrings(input []string) []string {
	seen := make(map[string]bool, len(input))
	out := make([]string, 0, len(input))
	for _, item := range input {
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func cloneOverlayTriggers(items []OverlayTrigger) []OverlayTrigger {
	if len(items) == 0 {
		return []OverlayTrigger{}
	}
	return append([]OverlayTrigger{}, items...)
}
