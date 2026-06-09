package bazi

import (
	"fmt"
	"strings"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

// isHuaQiWang 化神旺地映射表（三合局月令，《三命通会》标准）。
var isHuaQiWang = map[string][]string{
	"木": {"亥", "卯", "未"},
	"火": {"寅", "午", "戌"},
	"土": {"辰", "戌", "丑", "未"},
	"金": {"巳", "酉", "丑"},
	"水": {"申", "子", "辰"},
}

type PatternAnalysis struct {
	PatternName         string   `json:"pattern_name"`
	PatternType         string   `json:"pattern_type"`
	Description         string   `json:"description"`
	FavorableElements   []string `json:"favorable_elements"`
	UnfavorableElements []string `json:"unfavorable_elements"`
	SubType             string   `json:"sub_type,omitempty"`
}

// AnalyzePatternExtended 八字格局检测主入口。
// 检测顺序（高优先级先匹配）：
//  1. 专旺格（曲直/炎上/稼穑/从革/润下）
//  2. 从格（从财/从杀/从儿/从势/从弱）
//  3. 两神成像格
//  4. 化气格 / 从化格
//  5. 月令建禄 / 月刃 / 专禄 / 日禄归时 / 羊刃（柱支格）
//  6. 八格（正官/七杀/正印/偏印/正财/偏财/食神/伤官）
//  7. 复合格局（食神制杀/伤官配印/财滋杀旺/食神生财/正官佩印）
//  8. 三奇格 / 魁罡格 / 金神格 / 日德格
//  9. 正格（兜底）
func AnalyzePatternExtended(pillars []model.Pillar, monthZhi string, scores map[string]int, bodyStrength BodyStrengthResult) PatternAnalysis {
	if len(pillars) < 4 {
		return buildNormalPattern("", scores, bodyStrength)
	}

	dayGan := pillars[2].Gan
	dayZhi := pillars[2].Zhi
	hourZhi := pillars[3].Zhi

	// 1. 专旺格（最高优先级：若整局专旺，则不进入化气/八格判定）
	if pat := checkCongQiangGe(pillars, monthZhi, scores); pat != nil {
		return *pat
	}

	// 2. 从格
	if pat := checkCongCaiGe(pillars, scores); pat != nil {
		return *pat
	}
	if pat := checkQiMingCongShaGe(pillars, monthZhi, scores); pat != nil {
		return *pat
	}
	if pat := checkCongShiGe(pillars, scores); pat != nil {
		return *pat
	}
	if pat := checkCongErGe(pillars, scores); pat != nil {
		return *pat
	}
	if pat := checkCongRuoGe(pillars, scores); pat != nil {
		return *pat
	}

	// 3. 化气格 / 从化格（要求日主不旺，否则按八格论）
	if pat := checkHuaQiGe(pillars, monthZhi, scores); pat != nil {
		return *pat
	}
	if pat := checkCongHuaGe(pillars, monthZhi, scores); pat != nil {
		return *pat
	}

	// 4. 金神格（乙酉/己巳/癸酉日柱特殊格局，经典优先级高于八格）
	// 经典依据：《渊海子平》"乙酉、己巳、癸酉三日为金神"
	if pat := checkJinShenGe(dayGan, dayZhi, pillars); pat != nil {
		return *pat
	}

	// 5. 月令建禄/月刃；日支专禄/羊刃；时支日禄归时
	if pat := checkJianLuYueRen(pillars, monthZhi); pat != nil {
		return *pat
	}
	if pat := checkZhuanLuGe(dayGan, dayZhi, monthZhi); pat != nil {
		return *pat
	}
	if pat := checkRiLuGuiShiGe(dayGan, hourZhi, pillars); pat != nil {
		return *pat
	}
	if pat := checkYangRenGe(dayGan, dayZhi, monthZhi); pat != nil {
		return *pat
	}

	// 6. 八格（月令本气透干优先，透干次之，本气兜底）
	// 经典依据：《子平真诠》"月令提纲为定格之本"
	// 改进1：月令本气透干时优先取八格
	// 改进2：月令本气不透时，透干取格但月令十神同五行透干优先
	if pat := checkBaGe(pillars, monthZhi); pat != nil {
		return *pat
	}

	// 7. 复合格局（食神制杀/伤官配印/财滋杀旺/食神生财/正官佩印）
	// 优先级低于八格：月令本气可成八格时不应被复合格局截胡
	// 改进：复合格局需月令本气十神与格名匹配
	if pat := checkCompoundGe(pillars, dayGan, monthZhi); pat != nil {
		return *pat
	}

	// 8. 两神成像（八格无法定局时再判）
	if pat := checkLiangShenChengXiang(scores); pat != nil {
		return *pat
	}

	// 9. 其他特殊日柱（魁罡/三奇/日德）
	if pat := checkKuiGangGe(dayGan, dayZhi); pat != nil {
		return *pat
	}
	if pat := checkJinShenGe(dayGan, dayZhi, pillars); pat != nil {
		return *pat
	}
	if pat := checkSanQiGe(pillars); pat != nil {
		return *pat
	}
	if pat := checkRiDeGe(dayGan, dayZhi); pat != nil {
		return *pat
	}

	// 9. 正格（兜底）
	return buildNormalPattern(dayGan, scores, bodyStrength)
}

func checkHuaQiGe(pillars []model.Pillar, monthZhi string, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	huaQiCandidates := []int{1, 3} // 月、时干合
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	// 化气格要求日主不旺：日主五行得分占比 < 30%
	if float64(scores[dayElem])/float64(total) >= 0.3 {
		return nil
	}

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

		// 其余三柱天干（排除日干和当前合方）若 data.GanElement == keHua，破格
		keHua := keWo(huaQi)
		for j, pp := range pillars {
			if j == 2 || j == i {
				continue
			}
			if data.GanElement[pp.Gan] == keHua {
				return nil
			}
		}

		// 地支克化神比例 > 30% 破格
		if float64(scores[keHua])/float64(total) > 0.3 {
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

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	// 月令必须是同我（比劫）或生我（印星）
	yueElem := data.ZhiElement[monthZhi]
	if yueElem != dayElem && yueElem != shengWo(dayElem) {
		return nil
	}

	// 克破五行（官杀）天干不透（除日干外）
	keElem := keWo(dayElem)
	for i := 0; i < 4; i++ {
		if i == 2 {
			continue
		}
		if data.GanElement[pillars[i].Gan] == keElem {
			return nil
		}
	}
	// 地支藏干（含本气、中气、余气）不可有克破五行，否则破格
	for i := 0; i < 4; i++ {
		for _, elem := range ZhiAllElements[pillars[i].Zhi] {
			if elem == keElem {
				return nil
			}
		}
	}
	// 地支克破力量占比 < 10%
	if total > 0 && float64(scores[keElem])/float64(total) > 0.1 {
		return nil
	}

	// 生扶力量（日主+印星） > 60%
	supportScore := scores[dayElem] + scores[shengWo(dayElem)]
	if float64(supportScore)/float64(total) <= 0.6 {
		return nil
	}
	// 日主自身 > 30%
	if float64(scores[dayElem])/float64(total) < 0.3 {
		return nil
	}

	geName := ""
	switch {
	case dayElem == "木" && inStrings(monthZhi, "寅", "卯", "辰", "亥", "未"):
		geName = "曲直格"
	case dayElem == "火" && inStrings(monthZhi, "巳", "午", "未", "寅", "戌"):
		geName = "炎上格"
	case dayElem == "土" && inStrings(monthZhi, "辰", "戌", "丑", "未"):
		geName = "稼穑格"
	case dayElem == "金" && inStrings(monthZhi, "申", "酉", "戌", "巳", "丑"):
		geName = "从革格"
	case dayElem == "水" && inStrings(monthZhi, "亥", "子", "丑", "申", "辰"):
		geName = "润下格"
	}
	if geName == "" {
		return nil
	}

	return &PatternAnalysis{
		PatternName:         fmt.Sprintf("%s（从强格）", geName),
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日主%s，生于%s月，得月令，全局生扶专旺，不见克破，成%s。喜生扶，忌克破。", dayGan, monthZhi, geName),
		FavorableElements:   []string{shengWo(dayElem), dayElem},
		UnfavorableElements: []string{keElem, woKe(dayElem)},
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
		if float64(scores[elem])/float64(total) >= 0.15 {
			majorElems = append(majorElems, elem)
		}
	}
	if len(majorElems) != 2 {
		return nil
	}

	score1 := scores[majorElems[0]]
	score2 := scores[majorElems[1]]
	if absInt(score1-score2) > total/4 {
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
		FavorableElements:   []string{tongGuan(majorElems[0], majorElems[1])},
		UnfavorableElements: []string{},
		SubType:             relation,
	}
}

func checkKuiGangGe(gan, zhi string) *PatternAnalysis {
	// 经典依据：三命通会"庚辰庚戌壬辰戊戌为魁罡"
	dayCol := gan + zhi
	isShuiTu := dayCol == "庚辰" || dayCol == "壬辰"
	isTuJin := dayCol == "戊戌" || dayCol == "庚戌"
	if !isShuiTu && !isTuJin {
		return nil
	}

	dayElem := data.GanElement[gan]
	if isShuiTu {
		// 水土魁罡（庚辰、壬辰）：喜印星护身
		return &PatternAnalysis{
			PatternName:         "魁罡格（水土）",
			PatternType:         "特殊格局",
			Description:         fmt.Sprintf("日柱%s为水土魁罡，性格刚毅果断，喜印星护身、身旺。忌财星坏印、官杀攻身。", dayCol),
			FavorableElements:   []string{dayElem, shengWo(dayElem)},
			UnfavorableElements: []string{woKe(dayElem), keWo(dayElem)},
		}
	}
	// 土金魁罡（戊戌、庚戌）：喜财星滋养
	return &PatternAnalysis{
		PatternName:         "魁罡格（土金）",
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日柱%s为土金魁罡，性格刚毅果断，喜财星滋养。忌印星过旺、比劫争财。", dayCol),
		FavorableElements:   []string{dayElem, woKe(dayElem)},
		UnfavorableElements: []string{shengWo(dayElem), keWo(dayElem)},
	}
}

func checkRiDeGe(gan, zhi string) *PatternAnalysis {
	// 经典依据：《三命通会》日德五日"甲寅、丙午、戊辰、庚辰、壬戌"
	if map[string]bool{"甲寅": true, "丙午": true, "戊辰": true, "庚辰": true, "壬戌": true}[gan+zhi] {
		dayElem := data.GanElement[gan]
		return &PatternAnalysis{
			PatternName:         "日德格",
			PatternType:         "特殊格局",
			Description: fmt.Sprintf(
				"日柱%s%s为日德格，《三命通会》载日德五日：甲寅、丙午、戊辰、庚辰、壬戌。"+
					"性格慈善宽厚，福分深厚。"+
					"喜身旺、印绶生扶、食神泄秀；"+
					"忌刑冲破害、空亡、魁罡混杂。", gan, zhi),
			FavorableElements:   []string{dayElem, shengWo(dayElem), woSheng(dayElem)},
			UnfavorableElements: []string{keWo(dayElem), woKe(dayElem)},
		}
	}
	return nil
}

func checkJianLuYueRen(pillars []model.Pillar, monthZhi string) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	if dayElem == "" {
		return nil
	}

	if monthZhi == luShenZhi[dayGan] {
		// 建禄格：《子平真诠》"建禄格用财须身旺方能任财"
		// 喜财官食伤（我克、克我、我生），忌印比（生我、同我）
		return &PatternAnalysis{
			PatternName:         "建禄格",
			PatternType:         "正格",
			Description:         fmt.Sprintf("日主%s禄在月支%s，为建禄格。《子平真诠》云：建禄格用财须身旺方能任财。喜财官食伤，忌印比夺财。", dayGan, monthZhi),
			FavorableElements:   []string{woKe(dayElem), keWo(dayElem), woSheng(dayElem)},
			UnfavorableElements: []string{shengWo(dayElem), dayElem},
		}
	}
	if monthZhi == yangRenZhi(dayGan) {
		// 月刃格：《子平真诠》"月刃用官须透出有根"
		// 喜官杀制刃、食伤泄刃，忌印比助刃、财星生刃
		return &PatternAnalysis{
			PatternName:         "月刃格",
			PatternType:         "正格",
			Description:         fmt.Sprintf("日主%s羊刃在月支%s，为月刃格。《子平真诠》云：月刃用官须透出有根。喜官杀制刃、食伤泄秀，忌印比助刃。", dayGan, monthZhi),
			FavorableElements:   []string{keWo(dayElem), woSheng(dayElem)},
			UnfavorableElements: []string{shengWo(dayElem), dayElem},
		}
	}
	return nil
}

// checkZhuanLuGe 专禄格：日支为日主之禄（且月支不为日主之禄，区别于建禄）。
// 经典依据：《渊海子平》"日禄归时专禄格，皆同此格断"。
func checkZhuanLuGe(dayGan, dayZhi, monthZhi string) *PatternAnalysis {
	if dayZhi != luShenZhi[dayGan] {
		return nil
	}
	// 若月支也是禄，已被建禄格匹配，此处不重复
	if monthZhi == luShenZhi[dayGan] {
		return nil
	}
	dayElem := data.GanElement[dayGan]
	return &PatternAnalysis{
		PatternName:         "专禄格",
		PatternType:         "正格",
		Description:         fmt.Sprintf("日主%s专禄在日支%s，为专禄格。日支得禄根深蒂固，喜食伤泄秀、财官流通，忌印比重叠。", dayGan, dayZhi),
		FavorableElements:   []string{woSheng(dayElem), woKe(dayElem), keWo(dayElem)},
		UnfavorableElements: []string{shengWo(dayElem), dayElem},
	}
}

// checkRiLuGuiShiGe 日禄归时格：时支为日主之禄，且柱中不见正官透干。
// 经典依据：《子平真诠》"日禄归时格，柱中不见官星，则专取此格"。
func checkRiLuGuiShiGe(dayGan, hourZhi string, pillars []model.Pillar) *PatternAnalysis {
	if hourZhi != luShenZhi[dayGan] {
		return nil
	}
	// 柱中天干透出正官则不成立（官星破禄）
	for i, p := range pillars {
		if i == 2 {
			continue
		}
		if ClassifyTenGod(p.Gan, dayGan, false) == "正官" {
			return nil
		}
	}
	dayElem := data.GanElement[dayGan]
	return &PatternAnalysis{
		PatternName:         "日禄归时格",
		PatternType:         "正格",
		Description:         fmt.Sprintf("日主%s禄在时支%s，且天干不见正官破禄，为日禄归时格。喜食伤生财、财官流通，忌印比劫夺。", dayGan, hourZhi),
		FavorableElements:   []string{woSheng(dayElem), woKe(dayElem), keWo(dayElem)},
		UnfavorableElements: []string{shengWo(dayElem), dayElem},
	}
}

// checkYangRenGe 羊刃格：日支为日主之羊刃（区别于月刃）。
func checkYangRenGe(dayGan, dayZhi, monthZhi string) *PatternAnalysis {
	if dayZhi != yangRenZhi(dayGan) {
		return nil
	}
	if monthZhi == yangRenZhi(dayGan) {
		return nil // 月刃已优先匹配
	}
	dayElem := data.GanElement[dayGan]
	return &PatternAnalysis{
		PatternName:         "羊刃格",
		PatternType:         "正格",
		Description:         fmt.Sprintf("日主%s羊刃在日支%s，为羊刃格。喜官杀制刃、食伤泄秀，忌印比助刃。", dayGan, dayZhi),
		FavorableElements:   []string{keWo(dayElem), woSheng(dayElem)},
		UnfavorableElements: []string{shengWo(dayElem), dayElem},
	}
}

// checkJinShenGe 金神格：日柱为乙酉/己巳/癸酉
// 经典依据：《渊海子平》"金神者，乙酉、己巳、癸酉三日是也"
// 金神格喜火炼，忌水金过旺，火乡必显
func checkJinShenGe(dayGan, dayZhi string, pillars []model.Pillar) *PatternAnalysis {
	jinShenDays := map[string]bool{
		"乙酉": true,
		"己巳": true,
		"癸酉": true,
	}
	dayPillar := dayGan + dayZhi
	if !jinShenDays[dayPillar] {
		return nil
	}
	dayElem := data.GanElement[dayGan]
	return &PatternAnalysis{
		PatternName:         "金神格",
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日柱%s为金神格，《渊海子平》载金神入火乡必显。喜火炼金、食伤泄秀，忌水土过旺。", dayPillar),
		FavorableElements:   []string{woSheng(dayElem)},
		UnfavorableElements: []string{shengWo(dayElem), keWo(dayElem)},
	}
}

func checkSanQiGe(pillars []model.Pillar) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}
	if hasSequence(pillars, "甲", "戊", "庚") || hasSequence(pillars, "庚", "戊", "甲") {
		return &PatternAnalysis{PatternName: "天三奇格", PatternType: "特殊格局", Description: "天干甲戊庚顺排，成天三奇，主贵气非凡。喜顺行无破，忌冲克截断。", FavorableElements: []string{"木", "土", "金"}, UnfavorableElements: []string{"火", "水"}}
	}
	if hasSequence(pillars, "乙", "丙", "丁") || hasSequence(pillars, "丁", "丙", "乙") {
		return &PatternAnalysis{PatternName: "地三奇格", PatternType: "特殊格局", Description: "天干乙丙丁顺排，成地三奇，主才华出众。喜顺行无破，忌冲克截断。", FavorableElements: []string{"木", "火"}, UnfavorableElements: []string{"金", "水"}}
	}
	if hasSequence(pillars, "壬", "癸", "辛") || hasSequence(pillars, "辛", "癸", "壬") {
		return &PatternAnalysis{PatternName: "人三奇格", PatternType: "特殊格局", Description: "天干壬癸辛顺排，成人三奇，主智谋超群。喜顺行无破，忌冲克截断。", FavorableElements: []string{"水", "金"}, UnfavorableElements: []string{"土", "火"}}
	}
	return nil
}

func checkCongRuoGe(pillars []model.Pillar, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	// 经典依据：滴天髓"弱极者扶之扶之徒劳而无功"
	// 生扶力量（日主 + 印星）占比 < 15%
	supportScore := scores[dayElem] + scores[shengWo(dayElem)]
	if float64(supportScore)/float64(total) >= 0.15 {
		return nil
	}

	// 克泄耗必须透干或当令
	restrictElems := []string{keWo(dayElem), woSheng(dayElem), woKe(dayElem)}
	hasRestrict := false
	// 月令是克泄耗吗？
	yueElem := data.ZhiElement[pillars[1].Zhi]
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
			if inStrings(data.GanElement[pillars[i].Gan], restrictElems...) {
				hasRestrict = true
				break
			}
		}
	}
	if !hasRestrict {
		return nil
	}

	// 日主在地支无根（检查所有藏干）
	for _, p := range pillars {
		for _, elem := range ZhiAllElements[p.Zhi] {
			if elem == dayElem {
				return nil
			}
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

// checkCongHuaGe 从化格判断（《三命通会》专旺从化格局）。
// 化气成格后以化神为新日主，喜生扶化神及化神所生，忌克破化神。
func checkCongHuaGe(pillars []model.Pillar, monthZhi string, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	huaPairs := map[string]string{
		"甲": "己", "己": "甲",
		"乙": "庚", "庚": "乙",
		"丙": "辛", "辛": "丙",
		"丁": "壬", "壬": "丁",
		"戊": "癸", "癸": "戊",
	}

	huaTarget := huaPairs[dayGan]
	if huaTarget == "" {
		return nil
	}

	// 检查是否成化：月干或时干有合化之干
	hasHuaGan := false
	huaIdx := -1
	for _, i := range []int{1, 3} {
		if pillars[i].Gan == huaTarget {
			hasHuaGan = true
			huaIdx = i
			break
		}
	}
	if !hasHuaGan {
		return nil
	}

	huaElem := GanHeHua[dayGan+huaTarget]
	if huaElem == "" {
		return nil
	}

	// 月支为化神旺地
	if !inStrings(monthZhi, isHuaQiWang[huaElem]...) {
		return nil
	}

	// 克化神之字不透干
	keHua := keWo(huaElem)
	for i, p := range pillars {
		if i == 2 || i == huaIdx {
			continue
		}
		if data.GanElement[p.Gan] == keHua {
			return nil
		}
	}

	// 日主不旺条件：日主五行得分占比 < 25%（从化要求日主无力，否则为正化而非从化）
	if total > 0 && float64(scores[dayElem])/float64(total) >= 0.25 {
		return nil
	}

	return &PatternAnalysis{
		PatternName:         fmt.Sprintf("从化格（%s）", huaElem),
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日干%s从%s而化，月令%s旺地，化气成格。喜生扶化神及化神所生（印比食伤），忌克破化神（官杀财）。", dayGan, huaTarget, monthZhi),
		FavorableElements:   []string{shengWo(huaElem), huaElem, woSheng(huaElem)},
		UnfavorableElements: []string{keWo(huaElem), woKe(huaElem)},
		SubType:             huaElem,
	}
}

// checkCongCaiGe 从财格判断。
// 日主极弱，财星当令或透干。
func checkCongCaiGe(pillars []model.Pillar, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	// 日主极弱（生扶 < 10%）
	supportScore := scores[dayElem] + scores[shengWo(dayElem)]
	if float64(supportScore)/float64(total) >= 0.1 {
		return nil
	}

	// 日主在日支无根：日支藏干含日主五行则有根，不从
	dayZhi := pillars[2].Zhi
	for _, elem := range ZhiAllElements[dayZhi] {
		if elem == dayElem {
			return nil
		}
	}

	// 财星当令或天透
	caiElem := woKe(dayElem)
	yueElem := data.ZhiElement[pillars[1].Zhi]
	hasCai := yueElem == caiElem
	if !hasCai {
		for i, p := range pillars {
			if i == 2 {
				continue
			}
			if data.GanElement[p.Gan] == caiElem {
				hasCai = true
				break
			}
		}
	}
	if !hasCai {
		return nil
	}

	return &PatternAnalysis{
		PatternName:         "从财格",
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日主%s极弱，财星当令或透干，从财。喜食伤生财、官杀克身为用，忌印比扶身。", dayGan),
		FavorableElements:   []string{woSheng(dayElem), caiElem, keWo(dayElem)},
		UnfavorableElements: []string{dayElem, shengWo(dayElem)},
		SubType:             "财",
	}
}

// checkCongShiGe 从势格判断。
// 日主极弱，满局财官杀（财与官杀皆强），日主从之。
// 区别于从财格（仅财强）和从杀格（仅官杀强）。
func checkCongShiGe(pillars []model.Pillar, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	// 日主极弱（生扶 < 10%）
	supportScore := scores[dayElem] + scores[shengWo(dayElem)]
	if float64(supportScore)/float64(total) >= 0.1 {
		return nil
	}

	// 日主在日支无根：日支藏干含日主五行则有根，不从
	dayZhi := pillars[2].Zhi
	for _, elem := range ZhiAllElements[dayZhi] {
		if elem == dayElem {
			return nil
		}
	}

	caiElem := woKe(dayElem)
	shaElem := keWo(dayElem)

	// 财与官杀必须同时当令或透干
	hasCai := false
	hasSha := false
	yueElem := data.ZhiElement[pillars[1].Zhi]
	if yueElem == caiElem {
		hasCai = true
	}
	if yueElem == shaElem {
		hasSha = true
	}
	for i, p := range pillars {
		if i == 2 {
			continue
		}
		elem := data.GanElement[p.Gan]
		if elem == caiElem {
			hasCai = true
		}
		if elem == shaElem {
			hasSha = true
		}
	}
	if !hasCai || !hasSha {
		return nil
	}

	// 财+官杀合计须 > 60%
	dominantScore := scores[caiElem] + scores[shaElem]
	if float64(dominantScore)/float64(total) < 0.6 {
		return nil
	}

	return &PatternAnalysis{
		PatternName:         "从势格",
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日主%s极弱，满局财官杀，从势。喜食伤生财、财官顺势，忌印比扶身。", dayGan),
		FavorableElements:   []string{woSheng(dayElem), caiElem, shaElem},
		UnfavorableElements: []string{dayElem, shengWo(dayElem)},
		SubType:             "财官",
	}
}

// checkQiMingCongShaGe 弃命从煞格判断。
// 日主极弱，官杀当令。
func checkQiMingCongShaGe(pillars []model.Pillar, monthZhi string, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	// 生扶力量 < 10%
	supportScore := scores[dayElem] + scores[shengWo(dayElem)]
	if float64(supportScore)/float64(total) >= 0.1 {
		return nil
	}

	// 日主在日支无根：日支藏干含日主五行则有根，不从
	dayZhi := pillars[2].Zhi
	for _, elem := range ZhiAllElements[dayZhi] {
		if elem == dayElem {
			return nil
		}
	}

	// 官杀当令
	shaElem := keWo(dayElem)
	yueElem := data.ZhiElement[monthZhi]
	hasSha := yueElem == shaElem
	if !hasSha {
		for i, p := range pillars {
			if i == 2 {
				continue
			}
			if data.GanElement[p.Gan] == shaElem {
				hasSha = true
				break
			}
		}
	}
	if !hasSha {
		return nil
	}

	return &PatternAnalysis{
		PatternName:         "弃命从煞格",
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日主%s极弱，官杀当令，从杀。喜七杀/官星克身，食伤制杀为忌，忌印比扶身。", dayGan),
		FavorableElements:   []string{woKe(dayElem), shaElem},
		UnfavorableElements: []string{dayElem, shengWo(dayElem), woSheng(dayElem)},
		SubType:             "杀",
	}
}

// checkCongErGe 从儿格判断（格局总论"从儿格日主无气满局食伤"）
// 经典依据：滴天髓"从儿不论身强弱，只要吾儿又见儿"
// 日主极弱，食伤当令或透干，满局食伤生财，日主从之。
func checkCongErGe(pillars []model.Pillar, scores map[string]int) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	total := totalScore(scores)
	if dayElem == "" || total == 0 {
		return nil
	}

	// 日主极弱（生扶<15%）
	supportScore := scores[dayElem] + scores[shengWo(dayElem)]
	if float64(supportScore)/float64(total) >= 0.15 {
		return nil
	}

	// 食伤五行
	shiShangElem := woSheng(dayElem)
	// 财星五行
	caiElem := woKe(dayElem)

	// 食伤当令或透干
	yueElem := data.ZhiElement[pillars[1].Zhi]
	hasShiShang := yueElem == shiShangElem
	if !hasShiShang {
		for i, p := range pillars {
			if i == 2 {
				continue
			}
			if data.GanElement[p.Gan] == shiShangElem {
				hasShiShang = true
				break
			}
		}
	}
	if !hasShiShang {
		return nil
	}

	// 食伤+财星合计占比须>60%（满局食伤生财）
	dominantScore := scores[shiShangElem] + scores[caiElem]
	if float64(dominantScore)/float64(total) < 0.6 {
		return nil
	}

	// 日主在地支无根
	for _, p := range pillars {
		for _, elem := range ZhiAllElements[p.Zhi] {
			if elem == dayElem {
				return nil
			}
		}
	}

	return &PatternAnalysis{
		PatternName:         "从儿格",
		PatternType:         "特殊格局",
		Description:         fmt.Sprintf("日主%s极弱，满局食伤生财，从儿。喜食伤财星，忌印比扶身。", dayGan),
		FavorableElements:   []string{shiShangElem, caiElem},
		UnfavorableElements: []string{dayElem, shengWo(dayElem)},
		SubType:             "食伤",
	}
}

// checkBaGe 八格判定。
// 经典依据：《子平真诠》"月令提纲为定格之本，月支本气透干者即取以为格"。
// 规则：
//   0. 月令本气透干 → 本气取格（最优先）
//   1. 月令本气不透 → 月令同五行透干取格（如月令正财不透，偏财透→偏财格）
//   2. 月令十神无透 → 其他柱透干按官杀>食伤>财>印取格
//   3. 无透干 → 月令本气非比劫时取格（兜底）
func checkBaGe(pillars []model.Pillar, monthZhi string) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}

	dayGan := pillars[2].Gan
	dayElem := data.GanElement[dayGan]
	if dayElem == "" {
		return nil
	}

	// 0. 月令本气透干优先
	benQiGan := monthZhiBenQi(monthZhi)
	benQiTG := ClassifyTenGod(benQiGan, dayGan, false)
	if benQiTG != "" && benQiTG != "比肩" && benQiTG != "劫财" {
		for i, p := range pillars {
			if i == 2 {
				continue
			}
			if p.Gan == benQiGan {
				patternName, _ := patternFromTenGod(benQiTG)
				if patternName != "" {
					return buildBaGeResult(patternName, dayGan, monthZhi, benQiTG, dayElem, true)
				}
			}
		}
	}

	// 2. 月令本气不透但有偏正互换透干 → 互换透干取格
	// 允许：正财↔偏财、正印↔偏印、食神↔伤官
	// 不允许：正官↔七杀（官杀虽同类但取格不同，不可越权）
	benQiElem := data.GanElement[benQiGan]
	if benQiElem != "" && benQiTG != "" && benQiTG != "比肩" && benQiTG != "劫财" && benQiTG != "正官" && benQiTG != "七杀" {
		benQiCategory := tenGodCategory(benQiTG)
		for i, p := range pillars {
			if i == 2 {
				continue
			}
			stemTG := ClassifyTenGod(p.Gan, dayGan, i == 2)
			stemCategory := tenGodCategory(stemTG)
			if stemTG != "" && stemCategory == benQiCategory && stemTG != benQiTG {
				patternName, _ := patternFromTenGod(stemTG)
				if patternName != "" {
					return buildBaGeResult(patternName, dayGan, monthZhi, stemTG, dayElem, true)
				}
			}
		}
	}

	// 3. 月令本气不透也无同五行透干 → 月令本气非比劫时直接取格
	// 经典依据：《子平真诠》"月支本气虽不透干，亦可取以为格"
	if benQiTG != "" && benQiTG != "比肩" && benQiTG != "劫财" {
		patternName, _ := patternFromTenGod(benQiTG)
		if patternName != "" {
			return buildBaGeResult(patternName, dayGan, monthZhi, benQiTG, dayElem, false)
		}
	}

	// 4. 月令本气为比劫或无本气 → 其他柱透干按优先级取格
	tgPriority := []string{"正官", "七杀", "食神", "伤官", "正财", "偏财", "正印", "偏印"}
	stemPositions := []int{1, 3, 0} // 月、时、年
	for _, tg := range tgPriority {
		for _, idx := range stemPositions {
			if idx >= len(pillars) {
				continue
			}
			stemTG := ClassifyTenGod(pillars[idx].Gan, dayGan, idx == 2)
			if stemTG != tg {
				continue
			}
			patternName, _ := patternFromTenGod(tg)
			if patternName != "" {
				return buildBaGeResult(patternName, dayGan, monthZhi, tg, dayElem, true)
			}
		}
	}

	// 2. 月支本气取格（本气非比劫）
	mainStem := mainQi(monthZhi)
	if mainStem == "" {
		return nil
	}
	mainTenGod := ClassifyTenGod(stemFromQi(mainStem), dayGan, false)
	if mainTenGod == "" || mainTenGod == "比肩" || mainTenGod == "劫财" {
		return nil
	}
	patternName, _ := patternFromTenGod(mainTenGod)
	if patternName == "" {
		return nil
	}
	return buildBaGeResult(patternName, dayGan, monthZhi, mainTenGod, dayElem, false)
}

func buildBaGeResult(name, dayGan, monthZhi, tg, dayElem string, hasTougan bool) *PatternAnalysis {
	return &PatternAnalysis{
		PatternName:         name,
		PatternType:         "正格",
		Description:         fmt.Sprintf("日主%s，月令%s%s%s，成%s。", dayGan, monthZhi, ternaryStr(hasTougan, "透", "本气"), tg, name),
		FavorableElements:   geFavorable(dayElem, tg, hasTougan),
		UnfavorableElements: geUnfavorable(dayElem, tg),
		SubType:             strings.TrimSuffix(tg, "格"),
	}
}

func ternaryStr(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

// stemFromQi 从地支本气五行反推可能的本气天干（用于十神分类）。
func stemFromQi(qi string) string {
	main := map[string]string{
		"水": "壬", "木": "甲", "土": "戊", "火": "丙", "金": "庚",
	}
	return main[qi]
}

// patternFromTenGod 根据月令十神返回八格名称。
func patternFromTenGod(tg string) (string, string) {
	switch tg {
	case "正官":
		return "正官格", "正格"
	case "七杀":
		return "偏官格（七杀格）", "正格"
	case "正印":
		return "正印格", "正格"
	case "偏印":
		return "偏印格", "正格"
	case "正财":
		return "正财格", "正格"
	case "偏财":
		return "偏财格", "正格"
	case "食神":
		return "食神格", "正格"
	case "伤官":
		return "伤官格", "正格"
	}
	return "", ""
}

// geFavorable 八格的喜用神。
func geFavorable(dayElem, mainTG string, hasTougan bool) []string {
	switch mainTG {
	case "正官", "七杀":
		// 官杀格：喜印星化官、忌财星坏印
		return []string{shengWo(dayElem), dayElem}
	case "正印", "偏印":
		return []string{shengWo(dayElem), dayElem, woSheng(dayElem)}
	case "正财", "偏财":
		return []string{woKe(dayElem), woSheng(dayElem)}
	case "食神", "伤官":
		return []string{woSheng(dayElem), woKe(dayElem)}
	}
	return nil
}

// geUnfavorable 八格的忌神。
func geUnfavorable(dayElem, mainTG string) []string {
	switch mainTG {
	case "正官", "七杀":
		return []string{woKe(dayElem)}
	case "正印", "偏印":
		return []string{woKe(dayElem), keWo(dayElem)}
	case "正财", "偏财":
		return []string{keWo(dayElem), dayElem}
	case "食神", "伤官":
		return []string{keWo(dayElem)}
	}
	return nil
}

// monthZhiBenQi 返回月支本气天干
// 经典依据：寅本气甲、卯本气乙、辰本气戊、巳本气丙、午本气丁、未本气己、申本气庚、酉本气辛、戌本气戊、亥本气壬、子本气癸、丑本气己
var monthZhiBenQiMap = map[string]string{
	"寅": "甲", "卯": "乙", "辰": "戊", "巳": "丙", "午": "丁", "未": "己",
	"申": "庚", "酉": "辛", "戌": "戊", "亥": "壬", "子": "癸", "丑": "己",
}

func monthZhiBenQi(zhi string) string {
	if gan, ok := monthZhiBenQiMap[zhi]; ok {
		return gan
	}
	return ""
}

// tenGodCategory 返回十神类别：官/杀→官杀, 财→财, 印→印, 食/伤→食伤
func tenGodCategory(tg string) string {
	switch tg {
	case "正官", "七杀":
		return "官杀"
	case "正财", "偏财":
		return "财"
	case "正印", "偏印":
		return "印"
	case "食神", "伤官":
		return "食伤"
	default:
		return ""
	}
}

// checkCompoundGe 复合格局判定。
// 经典依据：《子平真诠》"相神有力方能成格"，且月令本气必须为格名所涉十神。
// 检测：食神制杀 / 伤官配印 / 财滋杀旺 / 食神生财 / 正官佩印。
func checkCompoundGe(pillars []model.Pillar, dayGan, monthZhi string) *PatternAnalysis {
	if len(pillars) < 4 {
		return nil
	}
	dayElem := data.GanElement[dayGan]
	if dayElem == "" {
		return nil
	}

	// 月令本气十神：关键约束，复合格局需月令十神与格名匹配
	monthBenQi := monthZhiBenQi(monthZhi)
	monthTenGod := ClassifyTenGod(monthBenQi, dayGan, false)

	// 计算各十神所在柱（基于藏干+天干）
	pairs := samePillarPairs(pillars, dayGan)

	// 食神制杀：食神与七杀在同柱或紧贴，且七杀必须透干，月令为食神或七杀
	shaTouGan := false
	for i, p := range pillars {
		if i == 2 {
			continue
		}
		if ClassifyTenGod(p.Gan, dayGan, false) == "七杀" {
			shaTouGan = true
			break
		}
	}
	if shaTouGan &&
		(hasPair(pairs, "食神", "七杀") || hasAdjacentPair(pillars, dayGan, "食神", "七杀")) &&
		!hasPair(pairs, "正印", "七杀") &&
		(monthTenGod == "食神" || monthTenGod == "七杀") {
		return buildCompound("食神制杀格", dayGan, dayElem,
			[]string{woSheng(dayElem)}, []string{dayElem, shengWo(dayElem), woKe(dayElem)})
	}
	// 伤官配印：伤官与印星（正印或偏印）同柱或紧贴，月令为伤官或印星
	if (hasPair(pairs, "伤官", "正印") || hasPair(pairs, "伤官", "偏印") ||
		hasAdjacentPair(pillars, dayGan, "伤官", "正印") || hasAdjacentPair(pillars, dayGan, "伤官", "偏印")) &&
		(monthTenGod == "伤官" || monthTenGod == "正印" || monthTenGod == "偏印") {
		return buildCompound("伤官配印格", dayGan, dayElem,
			[]string{shengWo(dayElem), dayElem}, []string{woKe(dayElem), keWo(dayElem)})
	}
	// 财滋杀旺：财星与七杀在同柱（财生杀），月令为财星或七杀
	if (hasPair(pairs, "正财", "七杀") || hasPair(pairs, "偏财", "七杀")) &&
		(monthTenGod == "正财" || monthTenGod == "偏财" || monthTenGod == "七杀") {
		return buildCompound("财滋杀旺格", dayGan, dayElem,
			[]string{woKe(dayElem), keWo(dayElem)}, []string{shengWo(dayElem), dayElem})
	}
	// 食神生财：食神与财星在同柱或紧贴，月令为食神或财星
	if (hasPair(pairs, "食神", "正财") || hasPair(pairs, "食神", "偏财") ||
		hasAdjacentPair(pillars, dayGan, "食神", "正财") || hasAdjacentPair(pillars, dayGan, "食神", "偏财")) &&
		(monthTenGod == "食神" || monthTenGod == "正财" || monthTenGod == "偏财") {
		return buildCompound("食神生财格", dayGan, dayElem,
			[]string{woSheng(dayElem), woKe(dayElem)}, []string{keWo(dayElem), dayElem})
	}
	// 正官佩印：正官与印星（正印或偏印）同柱或紧贴，月令为正官或印星
	if (hasPair(pairs, "正官", "正印") || hasPair(pairs, "正官", "偏印") ||
		hasAdjacentPair(pillars, dayGan, "正官", "正印") || hasAdjacentPair(pillars, dayGan, "正官", "偏印")) &&
		(monthTenGod == "正官" || monthTenGod == "正印" || monthTenGod == "偏印") {
		return buildCompound("正官佩印格", dayGan, dayElem,
			[]string{shengWo(dayElem), dayElem}, []string{woKe(dayElem), keWo(dayElem)})
	}
	return nil
}

func buildCompound(name, dayGan, dayElem string, favorable, unfavorable []string) *PatternAnalysis {
	return &PatternAnalysis{
		PatternName:         name,
		PatternType:         "正格",
		Description:         fmt.Sprintf("日主%s，成%s。", dayGan, name),
		FavorableElements:   favorable,
		UnfavorableElements: unfavorable,
	}
}

// samePillarPairs 列出四柱中各柱天干与地支本气的十神配对。
func samePillarPairs(pillars []model.Pillar, dayGan string) []pair {
	var pairs []pair
	for i, p := range pillars {
		ganTG := ClassifyTenGod(p.Gan, dayGan, i == 2)
		zhiMainStem := mainQi(p.Zhi)
		if zhiMainStem == "" {
			continue
		}
		zhiTG := ClassifyTenGod(stemFromQi(zhiMainStem), dayGan, false)
		if ganTG != "" && zhiTG != "" && ganTG != "日主" && zhiTG != "日主" {
			pairs = append(pairs, pair{ganTG, zhiTG})
		}
	}
	return pairs
}

type pair struct {
	a, b string
}

func hasPair(pairs []pair, a, b string) bool {
	for _, p := range pairs {
		if (p.a == a && p.b == b) || (p.a == b && p.b == a) {
			return true
		}
	}
	return false
}

// hasAdjacentPair 检查紧贴柱（年月、月日、日时）之间是否有指定十神配对。
// 紧贴：相邻两柱的天干十神对，或天干与地支本气对。
func hasAdjacentPair(pillars []model.Pillar, dayGan, tgA, tgB string) bool {
	idxPairs := [][2]int{{0, 1}, {1, 2}, {2, 3}}
	for _, ip := range idxPairs {
		a, b := pillars[ip[0]], pillars[ip[1]]
		tgs := []string{
			ClassifyTenGod(a.Gan, dayGan, ip[0] == 2),
			ClassifyTenGod(b.Gan, dayGan, ip[1] == 2),
			ClassifyTenGod(stemFromQi(mainQi(a.Zhi)), dayGan, false),
			ClassifyTenGod(stemFromQi(mainQi(b.Zhi)), dayGan, false),
		}
		for _, x := range tgs {
			for _, y := range tgs {
				if (x == tgA && y == tgB) || (x == tgB && y == tgA) {
					return true
				}
			}
		}
	}
	return false
}
