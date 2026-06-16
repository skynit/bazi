package bazi

import (
	"fmt"
	"math"
	"strings"

	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

var elementIdx = map[string]int{"木": 0, "火": 1, "土": 2, "金": 3, "水": 4}

var tianGanMap = map[string]struct {
	WuXing string
}{
	"甲": {"木"}, "乙": {"木"},
	"丙": {"火"}, "丁": {"火"},
	"戊": {"土"}, "己": {"土"},
	"庚": {"金"}, "辛": {"金"},
	"壬": {"水"}, "癸": {"水"},
}

// yueLingMatrix: rows = day element (木火土金水), cols = month branch element
// 经典依据：日主强弱判断"囚：日主克他力量较弱；死：日主被克力量最弱"
// 旺(3) 同我, 相(2) 我生, 休(1) 生我, 囚(0.5) 我克, 死(0) 克我
var yueLingMatrix = [5][5]float64{
	// 木   火   土   金   水   ← 月支五行
	{3, 2, 0, 0.5, 1}, // 木日主: 旺(木) 相(火) 死(土) 囚(金) 休(水)
	{1, 3, 2, 0, 0.5}, // 火日主: 休(木) 旺(火) 相(土) 死(金) 囚(水)
	{0.5, 1, 3, 2, 0}, // 土日主: 囚(木) 休(火) 旺(土) 相(金) 死(水)
	{0, 0.5, 1, 3, 2}, // 金日主: 死(木) 囚(火) 休(土) 旺(金) 相(水)
	{2, 0, 0.5, 1, 3}, // 水日主: 相(木) 死(火) 囚(土) 休(金) 旺(水)
}

func getYueLingScore(dayElem string, monthBranchElem string) float64 {
	di := elementIdx[dayElem]
	mi := elementIdx[monthBranchElem]
	return yueLingMatrix[di][mi]
}

// isSupport returns true if gan's element supports (比劫/印星) the day master.
func isSupport(gan string, dayElem string) bool {
	tg, ok := tianGanMap[gan]
	if !ok {
		return false
	}
	// 比劫：同五行
	if tg.WuXing == dayElem {
		return true
	}
	// 印星：生我者（木生火→火日主，壬癸生木→木日主...）
	// 生我者：木生火、火生土、土生金、金生水、水生木
	supporter := map[string]string{
		"木": "火", "火": "土", "土": "金", "金": "水", "水": "木",
	}
	if supporter[tg.WuXing] == dayElem {
		return true
	}
	return false
}

// isSameElement returns true if gan's element is the same as day master (比劫 only).
// 经典依据：滴天髓"通根者如甲木见寅卯"，通根仅指同五行
func isSameElement(gan string, dayElem string) bool {
	tg, ok := tianGanMap[gan]
	if !ok {
		return false
	}
	return tg.WuXing == dayElem
}

// isYinStar returns true if gan's element is印星 (生我者).
func isYinStar(gan string, dayElem string) bool {
	tg, ok := tianGanMap[gan]
	if !ok {
		return false
	}
	supporter := map[string]string{
		"木": "火", "火": "土", "土": "金", "金": "水", "水": "木",
	}
	return supporter[tg.WuXing] == dayElem
}

// isRestrict returns true if gan's element restricts (克泄耗) the day master.
func isRestrict(gan string, dayElem string) bool {
	tg, ok := tianGanMap[gan]
	if !ok {
		return false
	}
	if tg.WuXing == dayElem {
		return false // 同五行已由 isSupport 处理
	}
	// 克我: gan克day
	// 我生(泄): day生gan
	// 我克(耗): day克gan
	ke := map[string]string{"木": "土", "火": "金", "土": "水", "金": "木", "水": "火"}
	sheng := map[string]string{"木": "火", "火": "土", "土": "金", "金": "水", "水": "木"}
	return ke[tg.WuXing] == dayElem || // 克我
		sheng[dayElem] == tg.WuXing || // 我生(泄)
		ke[dayElem] == tg.WuXing // 我克(耗)
}

// zangGanWeight returns the藏干 weight for a given earth branch position.
func zangGanWeight(hsType tyme.HideHeavenStemType) float64 {
	switch hsType {
	case tyme.MAIN:
		return 0.6
	case tyme.MIDDLE:
		return 0.3
	case tyme.RESIDUAL:
		return 0.1
	}
	return 0.0
}

// changShengMap 十二长生阶段查询表（阳干顺行）
// 经典依据：滴天髓"长生帝旺，得气最厚；冠带临官，得气次之"
var changShengMap = map[string]map[string]string{
	"木": {"亥": "长生", "子": "沐浴", "丑": "冠带", "寅": "临官", "卯": "帝旺", "辰": "衰", "巳": "病", "午": "死", "未": "墓", "申": "绝", "酉": "胎", "戌": "养"},
	"火": {"寅": "长生", "卯": "沐浴", "辰": "冠带", "巳": "临官", "午": "帝旺", "未": "衰", "申": "病", "酉": "死", "戌": "墓", "亥": "绝", "子": "胎", "丑": "养"},
	"土": {"寅": "长生", "卯": "沐浴", "辰": "冠带", "巳": "临官", "午": "帝旺", "未": "衰", "申": "病", "酉": "死", "戌": "墓", "亥": "绝", "子": "胎", "丑": "养"},
	"金": {"巳": "长生", "午": "沐浴", "未": "冠带", "申": "临官", "酉": "帝旺", "戌": "衰", "亥": "病", "子": "死", "丑": "墓", "寅": "绝", "卯": "胎", "辰": "养"},
	"水": {"申": "长生", "酉": "沐浴", "戌": "冠带", "亥": "临官", "子": "帝旺", "丑": "衰", "寅": "病", "卯": "死", "辰": "墓", "巳": "绝", "午": "胎", "未": "养"},
}

// changShengWeight returns the twelve-stage weight for a given day element and branch.
// 长生/帝旺/临官 → 1.5, 沐浴/冠带/衰/墓 → 1.0, 胎/养/病/死 → 0.5, 绝 → 0
func changShengWeight(dayElem, branch string) float64 {
	stages, ok := changShengMap[dayElem]
	if !ok {
		return 1.0
	}
	stage, ok := stages[branch]
	if !ok {
		return 1.0
	}
	switch stage {
	case "长生", "帝旺", "临官":
		return 1.5
	case "沐浴", "冠带", "衰", "墓":
		return 1.0
	case "胎", "养", "病", "死":
		return 0.5
	case "绝":
		return 0.0
	}
	return 1.0
}

func hideStemTypeLabel(t tyme.HideHeavenStemType) string {
	switch t {
	case tyme.MAIN:
		return "本气"
	case tyme.MIDDLE:
		return "中气"
	case tyme.RESIDUAL:
		return "余气"
	default:
		return "藏干"
	}
}

func normalizedBodyScore(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return math.Round(v*10000) / 10000
}

func verdictForBodyStrength(score float64) string {
	switch {
	case score > 0.7:
		return "身旺"
	case score > 0.5:
		return "偏旺"
	case score > 0.4:
		return "中和"
	case score > 0.3:
		return "偏弱"
	default:
		return "身弱"
	}
}

func bodyStrengthSummary(dayGan, dayElem, verdict string, totalScore float64, like, dislike []string) string {
	return fmt.Sprintf(
		"%s日主属%s，综合得分%.3f，判为%s。喜%s，忌%s。",
		dayGan,
		dayElem,
		totalScore,
		verdict,
		strings.Join(like, "、"),
		strings.Join(dislike, "、"),
	)
}

func calcBodyStrengthV2(ec *tyme.EightChar) BodyStrengthResult {
	dayStem := ec.GetDay().GetHeavenStem()
	dayElem := dayStem.GetElement().GetName()

	monthBranch := ec.GetMonth().GetEarthBranch()
	monthElem := monthBranch.GetElement().GetName()

	rules := bodyStrengthRuleConfig()
	weights := rules.Weights
	norms := rules.Normalizers
	thresholds := rules.AdjustmentThresholds

	// 1. 得令
	lingScore := getYueLingScore(dayElem, monthElem)
	evidence := []BodyStrengthEvidence{
		{
			Component: "ling",
			Polarity:  "support",
			Source:    "月令",
			Item:      monthBranch.GetName(),
			Score:     lingScore,
			Reason:    fmt.Sprintf("日主%s遇%s月，月令五行为%s，得令原始分%.2f。", dayElem, monthBranch.GetName(), monthElem, lingScore),
		},
	}

	pillars := [](func() tyme.SixtyCycle){ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour}

	// 2. 得地：四柱地支藏干中仅统计同五行（通根），印星归入得势
	// 经典依据：滴天髓"通根者如甲木见寅卯"，通根仅指同五行，印星归入得势
	// 改进：藏干透天干时加权+20%（透干则力量彰显）
	allStems := []string{}
	for _, fn := range pillars {
		allStems = append(allStems, fn().GetHeavenStem().GetName())
	}
	isTougan := func(ganName string) bool {
		for _, s := range allStems {
			if s == ganName {
				return true
			}
		}
		return false
	}

	diScore := 0.0
	for _, fn := range pillars {
		branch := fn().GetEarthBranch()
		branchName := branch.GetName()
		csW := changShengWeight(dayElem, branchName)
		for _, hhs := range branch.GetHideHeavenStems() {
			hName := hhs.GetHeavenStem().GetName()
			if isSameElement(hName, dayElem) {
				w := zangGanWeight(hhs.GetType()) * 1.5 * csW
				if isTougan(hName) {
					w *= 1.2
				}
				diScore += w
				evidence = append(evidence, BodyStrengthEvidence{
					Component: "di",
					Polarity:  "support",
					Source:    branchName,
					Item:      hName,
					Score:     w,
					Reason:    fmt.Sprintf("%s支%s为%s，属%s同气通根，长生权重%.1f。", branchName, hName, hideStemTypeLabel(hhs.GetType()), dayElem, csW),
				})
			}
		}
	}

	// 专禄/建禄/阳刃加成
	// 经典依据：渊海子平"甲木得寅为禄，乙木得卯为禄"，专禄建禄力量殊胜
	// 改进：专禄/建禄/阳刃加成直接加到总分上，不被归一化稀释
	dayBranchName := ec.GetDay().GetEarthBranch().GetName()
	monthBranchName := ec.GetMonth().GetEarthBranch().GetName()
	dayGanName := dayStem.GetName()

	luBonus := 0.0
	luMap := map[string]string{"甲": "寅", "乙": "卯", "丙": "巳", "丁": "午", "戊": "巳", "己": "午", "庚": "申", "辛": "酉", "壬": "亥", "癸": "子"}
	if luZhi, ok := luMap[dayGanName]; ok {
		if dayBranchName == luZhi {
			luBonus += 0.08 // 专禄：日支为禄，直接加成
			evidence = append(evidence, BodyStrengthEvidence{Component: "bonus", Polarity: "support", Source: "日支", Item: dayBranchName, Score: 0.08, Reason: "日支坐禄，专禄加成。"})
		}
		if monthBranchName == luZhi {
			luBonus += 0.06 // 建禄：月支为禄
			evidence = append(evidence, BodyStrengthEvidence{Component: "bonus", Polarity: "support", Source: "月支", Item: monthBranchName, Score: 0.06, Reason: "月支建禄，月令得禄加成。"})
		}
	}

	renMap := map[string]string{"甲": "卯", "丙": "午", "戊": "午", "庚": "酉", "壬": "子"}
	if renZhi, ok := renMap[dayGanName]; ok {
		if dayBranchName == renZhi {
			luBonus += 0.07 // 阳刃：日支为刃
			evidence = append(evidence, BodyStrengthEvidence{Component: "bonus", Polarity: "support", Source: "日支", Item: dayBranchName, Score: 0.07, Reason: "日支坐阳刃，加成日主根气。"})
		}
		if monthBranchName == renZhi {
			luBonus += 0.05 // 月刃：月支为刃
			evidence = append(evidence, BodyStrengthEvidence{Component: "bonus", Polarity: "support", Source: "月支", Item: monthBranchName, Score: 0.05, Reason: "月支见阳刃，加成月令根气。"})
		}
	}

	// 3. 得势：天干（年月时三干）+ 日支藏干（×1.5，坐下有根最贴身）
	// 经典依据：渊海子平"天干有比劫帮身有印绶生身"，力量有层级差异
	// 经典依据：滴天髓"坐下有根最贴身"，日支藏干给予更高权重
	supportWeight := 0.0
	restrictWeight := 0.0
	for i, fn := range pillars {
		gan := fn().GetHeavenStem().GetName()
		if i == 2 { // 跳过日干本身
			continue
		}
		tg, ok := tianGanMap[gan]
		if !ok {
			continue
		}
		if tg.WuXing == dayElem {
			// 比肩 1.0，劫财 0.8（阴阳异，助力稍弱）
			score := 0.8
			godName := "劫财"
			if GanInfoOf(gan).yang == GanInfoOf(dayStem.GetName()).yang {
				score = 1.0
				godName = "比肩"
			}
			supportWeight += score
			evidence = append(evidence, BodyStrengthEvidence{
				Component: "shi",
				Polarity:  "support",
				Source:    []string{"年干", "月干", "日干", "时干"}[i],
				Item:      gan,
				Score:     score,
				Reason:    fmt.Sprintf("%s透出%s，属%s，帮扶日主。", gan, godName, dayElem),
			})
		} else if isRestrict(gan, dayElem) {
			// 官杀 -1.2，食伤 -0.8，财 -0.6
			godName := ClassifyTenGod(gan, dayStem.GetName(), false)
			score := 1.0
			switch godName {
			case "正官", "七杀":
				score = 1.2
			case "食神", "伤官":
				score = 0.8
			case "正财", "偏财":
				score = 0.6
			default:
				score = 1.0
			}
			restrictWeight += score
			evidence = append(evidence, BodyStrengthEvidence{
				Component: "shi",
				Polarity:  "restrict",
				Source:    []string{"年干", "月干", "日干", "时干"}[i],
				Item:      gan,
				Score:     -score,
				Reason:    fmt.Sprintf("%s透出%s，属克泄耗力量。", gan, godName),
			})
		}
	}
	// 日支藏干：坐下有根最贴身，权重×1.5
	dayBranch := ec.GetDay().GetEarthBranch()
	for _, hhs := range dayBranch.GetHideHeavenStems() {
		hiddenGan := hhs.GetHeavenStem().GetName()
		w := zangGanWeight(hhs.GetType()) * 1.5
		tg, ok := tianGanMap[hiddenGan]
		if !ok {
			continue
		}
		if tg.WuXing == dayElem {
			score := w * 0.8
			godName := "劫财"
			if GanInfoOf(hiddenGan).yang == GanInfoOf(dayStem.GetName()).yang {
				score = w
				godName = "比肩"
			} else {
				score = w * 0.8
			}
			supportWeight += score
			evidence = append(evidence, BodyStrengthEvidence{
				Component: "shi",
				Polarity:  "support",
				Source:    "日支藏干",
				Item:      hiddenGan,
				Score:     score,
				Reason:    fmt.Sprintf("日支藏%s为%s，坐下贴身帮扶。", hiddenGan, godName),
			})
		} else if isRestrict(hiddenGan, dayElem) {
			godName := ClassifyTenGod(hiddenGan, dayStem.GetName(), false)
			score := w
			switch godName {
			case "正官", "七杀":
				score = w * 1.2
			case "食神", "伤官":
				score = w * 0.8
			case "正财", "偏财":
				score = w * 0.6
			default:
				score = w * 1.0
			}
			restrictWeight += score
			evidence = append(evidence, BodyStrengthEvidence{
				Component: "shi",
				Polarity:  "restrict",
				Source:    "日支藏干",
				Item:      hiddenGan,
				Score:     -score,
				Reason:    fmt.Sprintf("日支藏%s为%s，贴身形成克泄耗。", hiddenGan, godName),
			})
		}
	}
	shiScore := supportWeight - restrictWeight

	// 4. 得生：地支藏干中印星归入此处（与通根区分）
	shengScore := 0.0
	// 天干印星
	for i, fn := range pillars {
		if i == 2 {
			continue
		}
		tg := fn().GetHeavenStem()
		if isYinStar(tg.GetName(), dayElem) {
			shengScore += 1.0
			evidence = append(evidence, BodyStrengthEvidence{
				Component: "sheng",
				Polarity:  "support",
				Source:    []string{"年干", "月干", "日干", "时干"}[i],
				Item:      tg.GetName(),
				Score:     1.0,
				Reason:    fmt.Sprintf("%s透出印星，生扶日主%s。", tg.GetName(), dayStem.GetName()),
			})
		}
	}
	// 地支藏干印星
	for _, fn := range pillars {
		for _, hhs := range fn().GetEarthBranch().GetHideHeavenStems() {
			if isYinStar(hhs.GetHeavenStem().GetName(), dayElem) {
				score := zangGanWeight(hhs.GetType())
				shengScore += score
				branchName := fn().GetEarthBranch().GetName()
				stemName := hhs.GetHeavenStem().GetName()
				evidence = append(evidence, BodyStrengthEvidence{
					Component: "sheng",
					Polarity:  "support",
					Source:    branchName,
					Item:      stemName,
					Score:     score,
					Reason:    fmt.Sprintf("%s中藏%s为印星，%s生扶日主。", branchName, stemName, hideStemTypeLabel(hhs.GetType())),
				})
			}
		}
	}

	// 总分：归一化后按经典权重加权
	// 经典依据：日主强弱判断"得令40% 得地30% 得势20% 得气10%"
	normLing := lingScore / norms.Ling
	normDi := diScore / norms.Di
	normShi := 1.0 / (1.0 + math.Exp(-shiScore/norms.ShiSigmoidDivisor))
	normSheng := 1.0 / (1.0 + math.Exp(-shengScore/norms.ShengSigmoidDivisor))
	totalScore := normLing*weights.Ling + normDi*weights.Di + normShi*weights.Shi + normSheng*weights.Sheng + luBonus
	components := []BodyStrengthComponent{
		{Key: "ling", Name: "得令", RawScore: lingScore, NormalizedScore: normalizedBodyScore(normLing), Weight: weights.Ling, WeightedScore: normLing * weights.Ling, Description: "月令为提纲，按日主五行与月支五行旺相休囚死取分。"},
		{Key: "di", Name: "得地", RawScore: diScore, NormalizedScore: normalizedBodyScore(normDi), Weight: weights.Di, WeightedScore: normDi * weights.Di, Description: "四支藏干中同五行通根，结合本中余气、长生权重和透干加成。"},
		{Key: "shi", Name: "得势", RawScore: shiScore, NormalizedScore: normalizedBodyScore(normShi), Weight: weights.Shi, WeightedScore: normShi * weights.Shi, Description: "天干与日支藏干的比劫印助减去官杀食伤财的克泄耗。"},
		{Key: "sheng", Name: "得生", RawScore: shengScore, NormalizedScore: normalizedBodyScore(normSheng), Weight: weights.Sheng, WeightedScore: normSheng * weights.Sheng, Description: "天干和地支藏干中的印星生扶力量。"},
		{Key: "bonus", Name: "禄刃加成", RawScore: luBonus, NormalizedScore: normalizedBodyScore(luBonus), Weight: weights.Bonus, WeightedScore: luBonus, Description: "专禄、建禄、阳刃、月刃直接加成。"},
	}

	var verdict string
	var like, dislike []string
	// 判定阈值：以 0.5 为中和中心，向两侧阶梯划分
	verdict = verdictForBodyStrength(totalScore)

	// 5. 后验修正："得令不旺失令不衰"
	// 经典依据：滴天髓"春木虽强金太重而木亦危"
	// 如果得令五行被克它的五行总力超过自身力量的60%，则降低其旺度20%
	restrainingMap := map[string]string{
		"木": "金", "火": "水", "土": "木", "金": "火", "水": "土",
	}
	supportingMap := map[string]string{
		"木": "水", "火": "木", "土": "火", "金": "土", "水": "金",
	}
	var adjustments []BodyStrengthAdjustment
	if lingScore >= 2.0 { // 得令
		reElem := restrainingMap[dayElem]
		restrainingForce := 0.0
		selfForce := lingScore + diScore + shengScore
		for _, fn := range pillars {
			sc := fn()
			stElem := sc.GetHeavenStem().GetElement().GetName()
			if stElem == reElem {
				restrainingForce += 5.0
			}
			for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
				if hhs.GetHeavenStem().GetElement().GetName() == reElem {
					restrainingForce += zangGanWeight(hhs.GetType()) * 3.0
				}
			}
		}
		// 阈值从0.6降至0.4：克我力量占比超40%即触发降级
		if selfForce > 0 && restrainingForce/selfForce > thresholds.DeLingRestrictRatio {
			before := totalScore
			totalScore *= thresholds.DeLingMultiplier
			adjustments = append(adjustments, BodyStrengthAdjustment{
				Name:        "得令不旺",
				Before:      before,
				After:       totalScore,
				Reason:      fmt.Sprintf("克我%s力量%.2f / 自身生扶%.2f > 40%%。", reElem, restrainingForce, selfForce),
				Description: "虽得月令，但克制力量过重，降低旺度一级。",
			})
			if verdict == "身旺" {
				verdict = "偏旺"
			} else if verdict == "偏旺" {
				verdict = "中和"
			}
		}
	} else if lingScore <= 0.5 { // 失令
		spElem := supportingMap[dayElem]
		supportingForce := 0.0
		for _, fn := range pillars {
			sc := fn()
			stElem := sc.GetHeavenStem().GetElement().GetName()
			if stElem == spElem || stElem == dayElem {
				supportingForce += 5.0
			}
			for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
				hElem := hhs.GetHeavenStem().GetElement().GetName()
				if hElem == spElem || hElem == dayElem {
					supportingForce += zangGanWeight(hhs.GetType()) * 3.0
				}
			}
		}
		// 阈值从8.0降至5.0：生扶力量达5.0即触发升级
		if supportingForce >= thresholds.ShiLingSupportForce {
			before := totalScore
			totalScore = totalScore*thresholds.ShiLingBlendSelf + 0.5*thresholds.ShiLingBlendNeutral
			adjustments = append(adjustments, BodyStrengthAdjustment{
				Name:        "失令不衰",
				Before:      before,
				After:       totalScore,
				Reason:      fmt.Sprintf("生扶%s及同气%s力量合计%.2f，达到修正阈值。", spElem, dayElem, supportingForce),
				Description: "虽失月令，但命局另有生扶根气，提升衰弱判断。",
			})
			if verdict == "身弱" {
				verdict = "偏弱"
			} else if verdict == "偏弱" {
				verdict = "中和"
			}
		}
	}

	// 根据日主五行动态计算喜忌
	// 身旺: 喜克泄耗(克我+我生+我克), 忌生扶(生我+同我)
	// 身弱: 喜生扶(生我+同我), 忌克泄耗(克我+我生+我克)
	elemRelation := map[string]map[string]string{
		"木": {"同我": "木", "生我": "水", "我生": "火", "克我": "金", "我克": "土"},
		"火": {"同我": "火", "生我": "木", "我生": "土", "克我": "水", "我克": "金"},
		"土": {"同我": "土", "生我": "火", "我生": "金", "克我": "木", "我克": "水"},
		"金": {"同我": "金", "生我": "土", "我生": "水", "克我": "火", "我克": "木"},
		"水": {"同我": "水", "生我": "金", "我生": "木", "克我": "土", "我克": "火"},
	}
	rel := elemRelation[dayElem]
	sameElem := rel["同我"]
	supportElem := rel["生我"]
	drainElem := rel["我生"]
	controlElem := rel["克我"]
	wealthElem := rel["我克"]
	if verdict == "身旺" || verdict == "偏旺" {
		like = []string{controlElem, drainElem, wealthElem}
		dislike = []string{supportElem, sameElem}
	} else if verdict == "身弱" || verdict == "偏弱" {
		like = []string{supportElem, sameElem}
		dislike = []string{controlElem, drainElem, wealthElem}
	} else {
		// 中和：五行相对平衡，喜通关调候
		// 经典依据：穷通宝鉴调候用神，在中和格局中优先考虑
		tiaoHouRules := data.GetTiaohou(dayStem.GetName(), monthBranch.GetName())
		tiaoHouElem := ""
		if len(tiaoHouRules) > 0 {
			tiaoHouElem = data.GanElement[tiaoHouRules[0].XiShen]
		}
		like = []string{drainElem, wealthElem}
		if tiaoHouElem != "" && tiaoHouElem != drainElem && tiaoHouElem != wealthElem {
			like = append(like, tiaoHouElem)
		}
		dislike = []string{controlElem}
	}

	return BodyStrengthResult{
		RuleVersion: RuleVersion,
		School:      RuleSchool,
		Verdict:     verdict,
		Like:        like,
		Dislike:     dislike,
		TotalScore:  totalScore,
		LingScore:   lingScore,
		DiScore:     diScore,
		ShiScore:    shiScore,
		ShengScore:  shengScore,
		LuBonus:     luBonus,
		Components:  components,
		Evidence:    evidence,
		Adjustments: adjustments,
		Summary:     bodyStrengthSummary(dayStem.GetName(), dayElem, verdict, totalScore, like, dislike),
	}
}
