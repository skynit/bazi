package ziwei

import (
	"fmt"
	"sort"
	"strings"
)

// LiunianOverlayAnalysis explains how an annual overlay was derived and which
// palaces are most affected.
type LiunianOverlayAnalysis struct {
	Year         int                  `json:"year"`
	GanZhi       string               `json:"gan_zhi"`
	Stem         string               `json:"stem"`
	Branch       string               `json:"branch"`
	ShiShen      string               `json:"shi_shen,omitempty"`
	Score        int                  `json:"score"`
	Tone         string               `json:"tone"`
	KeyTips      string               `json:"key_tips"`
	Summary      string               `json:"summary"`
	Method       []OverlayMethodStep  `json:"method"`
	FourHua      []OverlayTrigger     `json:"four_hua"`
	AnnualStars  []OverlayTrigger     `json:"annual_stars"`
	FocusPalaces []OverlayFocusPalace `json:"focus_palaces"`
}

type OverlayMethodStep struct {
	Label   string `json:"label"`
	Value   string `json:"value"`
	Meaning string `json:"meaning"`
}

type OverlayTrigger struct {
	Type     string `json:"type"`
	Star     string `json:"star,omitempty"`
	Palace   string `json:"palace"`
	Branch   string `json:"branch"`
	Meaning  string `json:"meaning"`
	Polarity string `json:"polarity"`
}

type OverlayFocusPalace struct {
	Palace    string           `json:"palace"`
	Branch    string           `json:"branch"`
	Score     int              `json:"score"`
	Triggers  []OverlayTrigger `json:"triggers"`
	MainStars []string         `json:"main_stars"`
	Advice    string           `json:"advice"`
}

// AnalyzeLiunianOverlay combines annual star placement, annual four-hua and
// period interpretation into a structured explanation for the frontend.
func (s *ZiWeiService) AnalyzeLiunianOverlay(base *ZiWeiChart, liunian *ZiWeiChart, year int) *LiunianOverlayAnalysis {
	if base == nil {
		return nil
	}
	chart := base
	if liunian != nil {
		chart = liunian
	}

	stemIdx, branchIdx := annualStemBranch(year)
	stem := StemNames[stemIdx]
	branch := BranchNames[branchIdx]
	ganZhi := stem + branch

	score := 60
	tone := "依据流年干支、四化与流禄羊陀马定位，重点观察被触发宫位。"
	keyTips := "先看年度四化落宫，再看流禄、流羊、流陀、流马是否叠加到同一宫位。"
	shiShen := ""
	if birth := base.GetBirthData(); birth != nil {
		interp := NewPeriodInterpreter(birth)
		if result := interp.AnalyzeLiunian(chart, year); result != nil {
			score = result.Score
			tone = result.OverallTone
			keyTips = result.KeyTips
			shiShen = result.ShiShen
			ganZhi = result.GanZhi
		}
	}

	fourHua := buildLiunianFourHuaTriggers(chart, stemIdx)
	annualStars := buildAnnualStarTriggers(chart, stemIdx, branchIdx)
	allTriggers := append(append([]OverlayTrigger{}, fourHua...), annualStars...)
	focus := buildOverlayFocusPalaces(chart, allTriggers)

	method := []OverlayMethodStep{
		{
			Label:   "流年干支",
			Value:   fmt.Sprintf("%d年 %s", year, ganZhi),
			Meaning: "用目标年份推得流年天干地支，作为本年叠盘的时间层。",
		},
		{
			Label:   "天干四化",
			Value:   describeFourHuaByStem(stemIdx),
			Meaning: "流年天干决定化禄、化权、化科、化忌，落到含对应星曜的本命宫位。",
		},
		{
			Label:   "禄羊陀马",
			Value:   describeAnnualStarPositions(chart, stemIdx, branchIdx),
			Meaning: "流禄看资源，流羊看竞争压力，流陀看拖延阻滞，流马看移动变化。",
		},
		{
			Label:   "三方四正",
			Value:   "悬停宫位可查看本宫、对宫、三合宫关系",
			Meaning: "年度触发宫位不孤立判断，需要连同对宫和三合宫一起读。",
		},
	}

	return &LiunianOverlayAnalysis{
		Year:         year,
		GanZhi:       ganZhi,
		Stem:         stem,
		Branch:       branch,
		ShiShen:      shiShen,
		Score:        score,
		Tone:         tone,
		KeyTips:      keyTips,
		Summary:      buildOverlaySummary(year, ganZhi, focus),
		Method:       method,
		FourHua:      fourHua,
		AnnualStars:  annualStars,
		FocusPalaces: focus,
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
			Type:     label,
			Star:     star,
			Palace:   palace.Name,
			Branch:   palace.Branch,
			Meaning:  fourHuaMeaning(label, star, palace.Name),
			Polarity: fourHuaPolarity(label),
		})
	}
	return triggers
}

func buildAnnualStarTriggers(chart *ZiWeiChart, stemIdx, branchIdx int) []OverlayTrigger {
	positions := []struct {
		label    string
		branch   int
		polarity string
	}{
		{"流禄", LucunBranchIdx[stemIdx], "good"},
		{"流羊", fixIndex(LucunBranchIdx[stemIdx] + 1), "watch"},
		{"流陀", fixIndex(LucunBranchIdx[stemIdx] - 1), "watch"},
		{"流马", TianmaBranchIdx[branchIdx], "movement"},
	}

	triggers := make([]OverlayTrigger, 0, len(positions))
	for _, pos := range positions {
		palace := findPalaceByBranch(chart, pos.branch)
		if palace == nil {
			continue
		}
		triggers = append(triggers, OverlayTrigger{
			Type:     pos.label,
			Palace:   palace.Name,
			Branch:   palace.Branch,
			Meaning:  annualStarMeaning(pos.label, palace.Name),
			Polarity: pos.polarity,
		})
	}
	return triggers
}

func buildOverlayFocusPalaces(chart *ZiWeiChart, triggers []OverlayTrigger) []OverlayFocusPalace {
	type group struct {
		palace   *PalaceInfo
		triggers []OverlayTrigger
		score    int
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
		g.score += triggerWeight(trigger)
	}

	focus := make([]OverlayFocusPalace, 0, len(groups))
	for _, g := range groups {
		if g.palace == nil {
			continue
		}
		focus = append(focus, OverlayFocusPalace{
			Palace:    g.palace.Name,
			Branch:    g.palace.Branch,
			Score:     g.score,
			Triggers:  cloneOverlayTriggers(g.triggers),
			MainStars: palaceMainStarNames(*g.palace),
			Advice:    focusAdvice(g.palace.Name, g.triggers),
		})
	}

	sort.SliceStable(focus, func(i, j int) bool {
		if focus[i].Score == focus[j].Score {
			return focus[i].Palace < focus[j].Palace
		}
		return focus[i].Score > focus[j].Score
	})
	if len(focus) > 5 {
		focus = focus[:5]
	}
	return focus
}

func triggerWeight(trigger OverlayTrigger) int {
	switch trigger.Type {
	case "化忌":
		return 4
	case "流羊", "流陀":
		return 3
	case "化禄", "流禄":
		return 3
	case "化权", "化科", "流马":
		return 2
	default:
		return 1
	}
}

func focusAdvice(palace string, triggers []OverlayTrigger) string {
	hasWatch := false
	hasGood := false
	hasMove := false
	for _, trigger := range triggers {
		switch trigger.Polarity {
		case "watch":
			hasWatch = true
		case "good":
			hasGood = true
		case "movement":
			hasMove = true
		}
		if trigger.Type == "化忌" {
			hasWatch = true
		}
	}
	switch {
	case hasWatch && hasGood:
		return fmt.Sprintf("%s同时有助力和压力，适合先定边界，再把资源投入到可验证的事项。", palace)
	case hasWatch:
		return fmt.Sprintf("%s有压力或阻滞信号，宜做风险控制，避免情绪化决策。", palace)
	case hasGood && hasMove:
		return fmt.Sprintf("%s有资源与变动并行，适合主动争取机会，同时预留调整空间。", palace)
	case hasGood:
		return fmt.Sprintf("%s有资源、名望或助力信号，可把握窗口推进重点事项。", palace)
	case hasMove:
		return fmt.Sprintf("%s有移动变化信号，适合处理出行、迁动、转换与对外联络。", palace)
	default:
		return fmt.Sprintf("%s被流年触发，需结合本命星曜和三方四正一起判断。", palace)
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
	return fmt.Sprintf("%d年%s叠盘优先观察%s；这些宫位承接了流年四化或流禄羊陀马。", year, ganZhi, strings.Join(names, "、"))
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
		return fmt.Sprintf("%s化禄落%s，主资源、机会或收益倾向增加。", star, palace)
	case "化权":
		return fmt.Sprintf("%s化权落%s，主责任、执行力和竞争压力上升。", star, palace)
	case "化科":
		return fmt.Sprintf("%s化科落%s，主名誉、学习、文书和缓冲修复。", star, palace)
	case "化忌":
		return fmt.Sprintf("%s化忌落%s，主卡点、牵挂、延误或过度消耗。", star, palace)
	default:
		return fmt.Sprintf("%s%s落%s。", star, label, palace)
	}
}

func fourHuaPolarity(label string) string {
	switch label {
	case "化禄", "化科":
		return "good"
	case "化忌":
		return "watch"
	case "化权":
		return "neutral"
	default:
		return "neutral"
	}
}

func annualStarMeaning(label, palace string) string {
	switch label {
	case "流禄":
		return fmt.Sprintf("流禄落%s，代表本年资源、财气或人情助力的入口。", palace)
	case "流羊":
		return fmt.Sprintf("流羊落%s，代表竞争、冲突、急躁和硬碰硬的压力点。", palace)
	case "流陀":
		return fmt.Sprintf("流陀落%s，代表拖延、反复、阻滞和难以快速收尾的事项。", palace)
	case "流马":
		return fmt.Sprintf("流马落%s，代表迁动、出行、岗位转换或对外奔波。", palace)
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
	if len(palace.MainStars) > 0 || len(palace.AuxStars) > 0 {
		names := append([]string{}, palace.MainStars...)
		names = append(names, palace.AuxStars...)
		return overlayUniqueStrings(names)
	}
	names := make([]string, 0, len(palace.Stars))
	for _, star := range palace.Stars {
		names = append(names, star.Name)
	}
	return overlayUniqueStrings(names)
}

func palaceMainStarNames(palace PalaceInfo) []string {
	if len(palace.MainStars) > 0 {
		return overlayUniqueStrings(palace.MainStars)
	}
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
