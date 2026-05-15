package service

import (
	"fmt"
	"strings"
)

// PalaceReading holds interpreted text for a single palace.
type PalaceReading struct {
	PalaceName   string `json:"palace_name"`
	MainStarDesc string `json:"main_star_desc"`
	AuxStarsDesc string `json:"aux_stars_desc"`
	FourHuaDesc  string `json:"four_hua_desc"`
	TrineOppDesc string `json:"trine_opp_desc"`
	PatternDesc  string `json:"pattern_desc"`
	OverallDesc  string `json:"overall_desc"`
}

// ZiWeiInterpreter generates palace interpretation text using rule-based templates.
type ZiWeiInterpreter struct {
	mainStarTemplates map[string]map[string]string
	auxStarTemplates  map[string]string
	fourHuaTemplates  map[string]map[string]string
	patternTemplates  map[string]string
	palaceContext     map[string]string
}

// NewZiWeiInterpreter creates a fully-initialized ZiWeiInterpreter
// with hardcoded template data for all stars, brightness levels,
// four transformations, and patterns.
func NewZiWeiInterpreter() *ZiWeiInterpreter {
	return &ZiWeiInterpreter{
		mainStarTemplates: buildMainStarTemplates(),
		auxStarTemplates:  buildAuxStarTemplates(),
		fourHuaTemplates:  buildFourHuaTemplates(),
		patternTemplates:  buildPatternTemplates(),
		palaceContext:     buildPalaceContext(),
	}
}

// InterpretPalace generates a PalaceReading for the given palace within the chart.
// It combines main star analysis, auxiliary star effects, four-hua transformations,
// trine-opp position analysis, pattern matching, and an overall synthesis.
func (i *ZiWeiInterpreter) InterpretPalace(palace *PalaceInfo, chart *ZiWeiChart) *PalaceReading {
	if palace == nil || chart == nil {
		return &PalaceReading{PalaceName: "未知"}
	}

	reading := &PalaceReading{PalaceName: palace.Name}
	palaceIdx := i.findPalaceIndex(palace.Name, chart)

	reading.MainStarDesc = i.buildMainStarDesc(palace)
	reading.AuxStarsDesc = i.buildAuxStarsDesc(palace)
	reading.FourHuaDesc = i.buildFourHuaDesc(palace)
	reading.TrineOppDesc = i.buildTrineOppDesc(palaceIdx, chart)
	reading.PatternDesc = i.buildPatternDesc(palace, palaceIdx, chart)
	reading.OverallDesc = i.buildOverallDesc(reading, palace)

	return reading
}

// findPalaceIndex returns the index of a palace by name in the chart.
func (i *ZiWeiInterpreter) findPalaceIndex(name string, chart *ZiWeiChart) int {
	for idx := range chart.Palaces {
		if chart.Palaces[idx].Name == name {
			return idx
		}
	}
	return -1
}

// buildMainStarDesc generates text describing the main stars with their brightness.
func (i *ZiWeiInterpreter) buildMainStarDesc(palace *PalaceInfo) string {
	if len(palace.MainStars) == 0 {
		return "此宮無主星，須藉對宮及三方四正之星曜來補足。此格局稱為「空宮」，命主在此宮領域較無主見，容易受外在環境影響，但也因此具有較強的適應力和可塑性。"
	}

	var parts []string
	for _, star := range palace.MainStars {
		brightness := palace.Brightness[star]
		if brightness == "" {
			brightness = "平"
		}

		template := ""
		if starTemplates, ok := i.mainStarTemplates[star]; ok {
			if t, ok := starTemplates[brightness]; ok {
				template = t
			} else {
				template = starTemplates["平"]
			}
		}

		if template != "" {
			parts = append(parts, fmt.Sprintf("【%s・%s】%s", star, brightness, template))
		} else {
			parts = append(parts, fmt.Sprintf("【%s・%s】此星在此亮度下無詳細解讀記錄。", star, brightness))
		}
	}

	return strings.Join(parts, "\n\n")
}

// buildAuxStarsDesc generates text describing the influence of auxiliary stars.
func (i *ZiWeiInterpreter) buildAuxStarsDesc(palace *PalaceInfo) string {
	if len(palace.AuxStars) == 0 {
		return "此宮無輔星影響，命主在此領域較少受外緣因素干擾，表現較為純粹直接。"
	}

	var parts []string
	for _, star := range palace.AuxStars {
		if desc, ok := i.auxStarTemplates[star]; ok {
			parts = append(parts, fmt.Sprintf("・%s：%s", star, desc))
		} else {
			parts = append(parts, fmt.Sprintf("・%s：此輔星暫無詳細解讀記錄。", star))
		}
	}
	return strings.Join(parts, "\n")
}

// buildFourHuaDesc generates text describing the four-hua transformations.
func (i *ZiWeiInterpreter) buildFourHuaDesc(palace *PalaceInfo) string {
	if len(palace.FourHua) == 0 {
		return "此宮無四化影響，命主在此領域的發展主要依靠自身努力和本命星曜的力量，較少受到化祿化權化科化忌的影響。"
	}

	var parts []string
	for _, hua := range palace.FourHua {
		star, huaType := parseFourHua(hua)
		if star == "" || huaType == "" {
			parts = append(parts, fmt.Sprintf("・%s：無法解析此四化組合。", hua))
			continue
		}

		desc := ""
		if huaMap, ok := i.fourHuaTemplates[huaType]; ok {
			if t, ok := huaMap[star]; ok {
				desc = t
			}
		}

		prefix := palace.Name + "宮"
		if desc != "" {
			parts = append(parts, fmt.Sprintf("・%s在%s：%s", hua, prefix, desc))
		} else {
			parts = append(parts, fmt.Sprintf("・%s在%s：此四化組合暫無詳細解讀記錄。", hua, prefix))
		}
	}
	return strings.Join(parts, "\n")
}

// parseFourHua splits a four-hua string like "廉貞化祿" into star name and hua type.
func parseFourHua(hua string) (star, huaType string) {
	idx := strings.Index(hua, "化")
	if idx < 0 {
		return "", ""
	}
	return hua[:idx], hua[idx:]
}

// buildTrineOppDesc generates text describing the 三方四正 influences.
func (i *ZiWeiInterpreter) buildTrineOppDesc(palaceIdx int, chart *ZiWeiChart) string {
	if palaceIdx < 0 || palaceIdx >= 12 {
		return "無法分析三方四正。"
	}

	opposite := (palaceIdx + 6) % 12
	trine1 := (palaceIdx + 4) % 12
	trine2 := (palaceIdx + 8) % 12

	var parts []string

	oppPalace := chart.Palaces[opposite]
	if len(oppPalace.MainStars) > 0 {
		parts = append(parts, fmt.Sprintf("對宮【%s】：有%s等星曜，對本宮形成直接影響。",
			oppPalace.Name, strings.Join(oppPalace.MainStars, "、")))
	} else {
		parts = append(parts, fmt.Sprintf("對宮【%s】：無主星，對本宮的直接影響較弱。",
			oppPalace.Name))
	}

	for _, tri := range []int{trine1, trine2} {
		triPalace := chart.Palaces[tri]
		if len(triPalace.MainStars) > 0 {
			parts = append(parts, fmt.Sprintf("三合宮【%s】：有%s等星曜，與本宮形成和諧共振。",
				triPalace.Name, strings.Join(triPalace.MainStars, "、")))
		} else {
			parts = append(parts, fmt.Sprintf("三合宮【%s】：無主星，對本宮的助力較為薄弱。",
				triPalace.Name))
		}
	}

	parts = append(parts, "三方四正之氣流通暢順與否，關係到本宮領域的整體發展。以上各宮星曜的強弱，共同構成了本宮的外在環境和發展基礎。")

	return strings.Join(parts, "\n")
}

// buildPatternDesc generates pattern-related descriptions matching this palace.
func (i *ZiWeiInterpreter) buildPatternDesc(palace *PalaceInfo, palaceIdx int, chart *ZiWeiChart) string {
	relevant := i.filterRelevantPatterns(palace, palaceIdx, chart)

	if len(relevant) == 0 {
		return "此宮暫無匹配之格局標註。"
	}

	var parts []string
	for _, name := range relevant {
		if desc, ok := i.patternTemplates[name]; ok {
			parts = append(parts, fmt.Sprintf("【%s】%s", name, desc))
		} else if name != "" {
			parts = append(parts, fmt.Sprintf("【%s】此格局暫無詳細解讀記錄。", name))
		}
	}
	return strings.Join(parts, "\n\n")
}

// filterRelevantPatterns returns pattern names relevant to this specific palace.
func (i *ZiWeiInterpreter) filterRelevantPatterns(palace *PalaceInfo, palaceIdx int, chart *ZiWeiChart) []string {
	var relevant []string

	isMingPalace := palaceIdx == 0
	isEmptyPalace := len(palace.MainStars) == 0

	for _, patternName := range chart.Patterns {
		switch {
		case isMingPalace && patternName == "擎羊入命":
			for _, s := range palace.AuxStars {
				if s == "擎羊" {
					relevant = append(relevant, patternName)
					break
				}
			}
		case isMingPalace && patternName == "陀羅入命":
			for _, s := range palace.AuxStars {
				if s == "陀羅" {
					relevant = append(relevant, patternName)
					break
				}
			}
		case isEmptyPalace && patternName == "空宮":
			relevant = append(relevant, patternName)

		case patternName == "火貪格":
			if i.hasAuxStar(palace, "火星") || i.hasAuxStar(palace, "鈴星") {
				for _, s := range palace.MainStars {
					if s == "貪狼" {
						relevant = append(relevant, patternName)
						break
					}
				}
			}
		case patternName == "鈴貪格":
			if i.hasAuxStar(palace, "鈴星") {
				for _, s := range palace.MainStars {
					if s == "貪狼" {
						relevant = append(relevant, patternName)
						break
					}
				}
			}

		case patternName == "祿馬交馳":
			if i.hasAuxStar(palace, "祿存") && i.hasAuxStar(palace, "天馬") {
				relevant = append(relevant, patternName)
			}

		default:
			relevant = append(relevant, patternName)
		}
	}

	return relevant
}

// hasAuxStar checks if a palace contains a specific auxiliary star.
func (i *ZiWeiInterpreter) hasAuxStar(palace *PalaceInfo, target string) bool {
	for _, s := range palace.AuxStars {
		if s == target {
			return true
		}
	}
	return false
}

// buildOverallDesc synthesizes all reading fields into a comprehensive summary.
func (i *ZiWeiInterpreter) buildOverallDesc(reading *PalaceReading, palace *PalaceInfo) string {
	context, ok := i.palaceContext[palace.Name]
	if !ok {
		context = fmt.Sprintf("此為%s領域的紫微斗數分析。", palace.Name)
	}

	var parts []string
	parts = append(parts, context)

	starCount := len(palace.MainStars)
	auxCount := len(palace.AuxStars)
	huaCount := len(palace.FourHua)

	parts = append(parts, fmt.Sprintf("綜合來看，本宮共有%d顆主星、%d顆輔星、%d項四化影響。",
		starCount, auxCount, huaCount))

	if starCount == 0 {
		parts = append(parts, "由於命宮無主星，此領域的發展需仰賴三方四正的星曜補足，以及輔星和四化的加持。建議命主在此領域多藉助外部資源和貴人力量。")
	} else if starCount >= 2 {
		parts = append(parts, "多星匯聚此宮，意味著命主在此領域具有多重面向的特質和能力，但也可能因各種力量的交錯而產生矛盾。建議整合各種力量，找到最適合自己的發展方向。")
	} else {
		parts = append(parts, "單星坐守，特質純粹分明，命主在此領域的發展方向較為明確。建議圍繞此星的核心特質來規劃人生。")
	}

	if huaCount > 0 {
		parts = append(parts, "四化的影響為此宮增添了吉凶變化，須根據化祿化權化科化忌的不同組合來動態調整策略。")
	}

	return strings.Join(parts, " ")
}
