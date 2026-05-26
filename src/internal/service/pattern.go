package service

import (
	"fmt"
	"strings"

	"bazi/internal/model"
)

// isHuaQiWang 化神旺地映射表（严派标准）。
var isHuaQiWang = map[string][]string{
	"木": {"寅", "卯", "辰"},
	"火": {"巳", "午", "未"},
	"土": {"辰", "戌", "丑", "未"},
	"金": {"申", "酉", "戌"},
	"水": {"亥", "子", "丑"},
}

type PatternAnalysis struct {
	PatternName         string   `json:"pattern_name"`
	PatternType         string   `json:"pattern_type"`
	Description         string   `json:"description"`
	FavorableElements   []string `json:"favorable_elements"`
	UnfavorableElements []string `json:"unfavorable_elements"`
	SubType             string   `json:"sub_type,omitempty"`
}

func analyzePatternExtended(pillars []model.Pillar, monthZhi string, scores map[string]int, bodyStrength BodyStrengthResult) PatternAnalysis {
	if len(pillars) < 4 {
		return buildNormalPattern("", scores, bodyStrength)
	}

	dayGan := pillars[2].Gan
	dayZhi := pillars[2].Zhi

	if pat := checkHuaQiGe(pillars, monthZhi, scores); pat != nil {
		return *pat
	}
	if pat := checkCongQiangGe(pillars, monthZhi, scores); pat != nil {
		return *pat
	}
	if pat := checkLiangShenChengXiang(scores); pat != nil {
		return *pat
	}
	if pat := checkKuiGangGe(dayGan, dayZhi); pat != nil {
		return *pat
	}
	if pat := checkRiDeGe(dayGan, dayZhi); pat != nil {
		return *pat
	}
	if pat := checkJianLuYueRen(pillars, monthZhi); pat != nil {
		return *pat
	}
	if pat := checkSanQiGe(pillars); pat != nil {
		return *pat
	}
	if pat := checkCongRuoGe(pillars, scores); pat != nil {
		return *pat
	}

	return buildNormalPattern(dayGan, scores, bodyStrength)
}

func checkHuaQiGe(pillars []model.Pillar, monthZhi string, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	huaQiCandidates := []int{1, 3} // 月、时干合
	total := totalScore(scores)

	for _, i := range huaQiCandidates {
		p := pillars[i]
		if ganHe[dayGan] != p.Gan {
			continue
		}

		huaQi := GanHeHua[dayGan+p.Gan]
		if huaQi == "" {
			continue
		}

		// 月支必须为化神旺地
		if !inStrings(monthZhi, isHuaQiWang[huaQi]...) {
			continue
		}

		// 其余三柱天干（排除日干和当前合方）若 GanElement == keHua，破格
		keHua := keWo(huaQi)
		for j, pp := range pillars {
			if j == 2 || j == i {
				continue
			}
			if GanElement[pp.Gan] == keHua {
				return nil
			}
		}

		// 地支克化神比例 > 30% 破格
		if total > 0 && float64(scores[keHua])/float64(total) > 0.3 {
			continue
		}

		return &PatternAnalysis{
			PatternName:         fmt.Sprintf("化气格（%s）", huaQi),
			PatternType:         "特殊格局",
			Description:         fmt.Sprintf("日干%s与%s干%s合化%s，月令%s旺地，天干无克破，成真化气格。喜生扶化神及化神所生，忌克破化神。", dayGan, pLabel(i), p.Gan, huaQi, monthZhi),
			FavorableElements:   favorHuaQi(huaQi),
			UnfavorableElements: []string{keHua},
			SubType:             huaQi,
		}
	}
	return nil
}

func checkCongQiangGe(pillars []model.Pillar, monthZhi string, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayElem := GanElement[pillars[2].Gan]
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	// 月令必须是同我（比劫）或生我（印星）
	yueElem := ZhiElement[monthZhi]
	if yueElem != dayElem && yueElem != shengWo(dayElem) {
		return nil
	}

	// 克破五行（官杀）天干不透（除日干外）
	keElem := keWo(dayElem)
	for i := 0; i < 4; i++ {
		if i == 2 {
			continue
		}
		if GanElement[pillars[i].Gan] == keElem {
			return nil
		}
	}
	// 地支克破力量占比 < 10%
	if total > 0 && float64(scores[keElem])/float64(total) > 0.1 {
		return nil
	}

	// 生扶力量（日主+印星） > 2/3
	supportScore := scores[dayElem] + scores[shengWo(dayElem)]
	if float64(supportScore)/float64(total) <= 2.0/3.0 {
		return nil
	}
	// 日主自身 > 30%
	if float64(scores[dayElem])/float64(total) < 0.3 {
		return nil
	}

	geName := ""
	switch {
	case dayElem == "木" && inStrings(monthZhi, "寅", "卯", "辰"):
		geName = "曲直格"
	case dayElem == "火" && inStrings(monthZhi, "巳", "午", "未"):
		geName = "炎上格"
	case dayElem == "土" && inStrings(monthZhi, "辰", "戌", "丑", "未"):
		geName = "稼穑格"
	case dayElem == "金" && inStrings(monthZhi, "申", "酉", "戌"):
		geName = "从革格"
	case dayElem == "水" && inStrings(monthZhi, "亥", "子", "丑"):
		geName = "润下格"
	}
	if geName == "" {
		return nil
	}

	return &PatternAnalysis{
		PatternName:         fmt.Sprintf("%s（从强格）", geName),
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日主%s，生于%s月，得月令，全局生扶专旺，不见克破，成%s。喜生扶，忌克破。", pillars[2].Gan, monthZhi, geName),
		FavorableElements:   []string{shengWo(dayElem), dayElem},
		UnfavorableElements: []string{keElem},
		SubType:             dayElem,
	}
}

func checkLiangShenChengXiang(scores map[string]int) *PatternAnalysis {
	total := totalScore(scores)
	if total == 0 {
		return nil
	}

	var majorElems []string
	for _, elem := range []string{"木", "火", "土", "金", "水"} {
		if float64(scores[elem])/float64(total) >= 0.05 {
			majorElems = append(majorElems, elem)
		}
	}
	if len(majorElems) != 2 {
		return nil
	}

	score1 := scores[majorElems[0]]
	score2 := scores[majorElems[1]]
	if absInt(score1-score2) > total/5 {
		return nil
	}

	relation := ""
	if shengWo(majorElems[1]) == majorElems[0] {
		relation = fmt.Sprintf("%s生%s", majorElems[0], majorElems[1])
	} else if shengWo(majorElems[0]) == majorElems[1] {
		relation = fmt.Sprintf("%s生%s", majorElems[1], majorElems[0])
	} else if keWuXing(majorElems[0]) == majorElems[1] {
		relation = fmt.Sprintf("%s克%s", majorElems[0], majorElems[1])
	} else if keWuXing(majorElems[1]) == majorElems[0] {
		relation = fmt.Sprintf("%s克%s", majorElems[1], majorElems[0])
	} else {
		return nil
	}

	return &PatternAnalysis{
		PatternName:         fmt.Sprintf("两神成像格（%s）", relation),
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("全局仅有%s、%s两种五行，力量均衡，构成两神成像格。喜行通关五行，忌破坏平衡。", majorElems[0], majorElems[1]),
		FavorableElements:   []string{majorElems[0], majorElems[1]},
		UnfavorableElements: []string{},
		SubType:             relation,
	}
}

func checkKuiGangGe(gan, zhi string) *PatternAnalysis {
	if map[string]bool{"庚辰": true, "壬辰": true, "戊戌": true, "庚戌": true}[gan+zhi] {
		return &PatternAnalysis{
			PatternName:         "魁罡格",
			PatternType:         "特殊格局",
			Description:         fmt.Sprintf("日柱%s%s为魁罡日，性格刚毅果断，宜有规矩约束。喜身旺，忌财官。", gan, zhi),
			FavorableElements:   []string{},
			UnfavorableElements: []string{},
		}
	}
	return nil
}

func checkRiDeGe(gan, zhi string) *PatternAnalysis {
	if map[string]bool{"甲寅": true, "丙辰": true, "戊辰": true, "庚辰": true, "壬戌": true}[gan+zhi] {
		return &PatternAnalysis{
			PatternName:         "日德格",
			PatternType:         "特殊格局",
			Description:         fmt.Sprintf("日柱%s%s为日德，性格慈善，福分深厚。", gan, zhi),
			FavorableElements:   []string{},
			UnfavorableElements: []string{},
		}
	}
	return nil
}

func checkJianLuYueRen(pillars []model.Pillar, monthZhi string) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := GanElement[dayGan]
	if dayElem == "" {
		return nil
	}

	if monthZhi == luShenZhi[dayGan] {
		// 建禄喜财（我克）、官（克我）
		return &PatternAnalysis{
			PatternName:         "建禄格",
			PatternType:         "正格",
			Description:         fmt.Sprintf("日主%s禄在月支%s，为建禄格。喜财官，忌劫财。", dayGan, monthZhi),
			FavorableElements:   []string{woKe(dayElem), keWo(dayElem)},
			UnfavorableElements: []string{dayElem},
		}
	}
	if monthZhi == yangRenZhi(dayGan) {
		// 月刃喜官杀（克我）
		return &PatternAnalysis{
			PatternName:         "月刃格",
			PatternType:         "正格",
			Description:         fmt.Sprintf("日主%s羊刃在月支%s，为月刃格。喜官杀制刃。", dayGan, monthZhi),
			FavorableElements:   []string{keWo(dayElem)},
			UnfavorableElements: []string{dayElem},
		}
	}
	return nil
}

func checkSanQiGe(pillars []model.Pillar) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}
	if hasSequence(pillars, "甲", "戊", "庚") || hasSequence(pillars, "庚", "戊", "甲") {
		return &PatternAnalysis{PatternName: "天三奇格", PatternType: "特殊格局", Description: "天干甲戊庚顺排，成天三奇，主贵气非凡。"}
	}
	if hasSequence(pillars, "乙", "丙", "丁") || hasSequence(pillars, "丁", "丙", "乙") {
		return &PatternAnalysis{PatternName: "地三奇格", PatternType: "特殊格局", Description: "天干乙丙丁顺排，成地三奇，主才华出众。"}
	}
	if hasSequence(pillars, "壬", "癸", "辛") || hasSequence(pillars, "辛", "癸", "壬") {
		return &PatternAnalysis{PatternName: "人三奇格", PatternType: "特殊格局", Description: "天干壬癸辛顺排，成人三奇，主智谋超群。"}
	}
	return nil
}

func checkCongRuoGe(pillars []model.Pillar, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := GanElement[dayGan]
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	// 生扶力量（日主 + 印星）占比 < 10%
	supportScore := scores[dayElem] + scores[shengWo(dayElem)]
	if float64(supportScore)/float64(total) >= 0.1 {
		return nil
	}

	// 克泄耗必须透干或当令
	restrictElems := []string{keWo(dayElem), woSheng(dayElem), woKe(dayElem)}
	hasRestrict := false
	// 月令是克泄耗吗？
	yueElem := ZhiElement[pillars[1].Zhi]
	for _, e := range restrictElems {
		if yueElem == e {
			hasRestrict = true
			break
		}
	}
	if !hasRestrict {
		// 天干有克泄耗透出吗（除日干外）
		for i := 0; i < 4; i++ {
			if i == 2 {
				continue
			}
			if inStrings(GanElement[pillars[i].Gan], restrictElems...) {
				hasRestrict = true
				break
			}
		}
	}
	if !hasRestrict {
		return nil
	}

	// 日主在地支无本气根
	for _, p := range pillars {
		if mainQi(p.Zhi) == dayElem {
			return nil
		}
	}

	like, dislike := computeFavorByDayElem(dayElem, true)
	return &PatternAnalysis{
		PatternName:         "从弱格",
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日主%s极弱，满局克泄耗，从弱。喜%s，忌%s。", dayGan, strings.Join(like, ""), strings.Join(dislike, "")),
		FavorableElements:   like,
		UnfavorableElements: dislike,
	}
}

func buildNormalPattern(dayGan string, _ map[string]int, bs BodyStrengthResult) PatternAnalysis {
	return PatternAnalysis{
		PatternName:         "正格",
		PatternType:         "正格",
		Description:         fmt.Sprintf("日主%s，%s。喜%s，忌%s。", dayGan, bs.Verdict, strings.Join(bs.Like, ""), strings.Join(bs.Dislike, "")),
		FavorableElements:   bs.Like,
		UnfavorableElements: bs.Dislike,
	}
}

func pLabel(i int) string {
	labels := []string{"年", "月", "日", "时"}
	if i >= 0 && i < len(labels) {
		return labels[i]
	}
	return ""
}



func yangRenZhi(gan string) string {
	return map[string]string{"甲": "卯", "丙": "午", "戊": "午", "庚": "酉", "壬": "子"}[gan]
}

func hasSequence(pillars []model.Pillar, a, b, c string) bool {
	gans := []string{pillars[0].Gan, pillars[1].Gan, pillars[2].Gan, pillars[3].Gan}
	for i := 0; i < len(gans)-2; i++ {
		if gans[i] == a && gans[i+1] == b && gans[i+2] == c {
			return true
		}
	}
	return false
}
