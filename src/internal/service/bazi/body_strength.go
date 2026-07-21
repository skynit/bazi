package bazi

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/6tail/tyme4go/tyme"
)

const (
	BodyStrengthSchemaVersion               = "body-strength-evidence-2026-07-18.3"
	bodyStrengthRuleID                      = "bazi.body-strength-score-candidate-v3"
	bodyStrengthScoringProfile              = "local_fuyi_weighted_score_v3"
	yueLingRuleID                           = "bazi.body-strength.yue-ling-seasonal-state.v1"
	yueLingProfile                          = "local_ziping_yue_ling_5x12_v1"
	yueLingHashBasis                        = "utf8_compact_json_order_and_scores_v1"
	yueLingTableSHA256                      = "sha256:76d21a2761256db42976144f236056b080a32e367bc50cbb9b927eb04d1c26a6"
	bodyStrengthRootRuleID                  = "bazi.body-strength.root-evidence.v1"
	bodyStrengthRootProfile                 = "local_root_terrain_tou_gan_v1"
	bodyStrengthBonusRuleID                 = "bazi.body-strength.lu-yang-ren-bonus.v1"
	bodyStrengthBonusProfile                = "local_lu_yang_ren_bonus_v1"
	bodyStrengthBonusHashBasis              = "utf8_compact_json_bonus_table_v1"
	bodyStrengthBonusSHA256                 = "sha256:01838d3e24dd23da4c00579676bc7928352364cd1a14f90691b7ea38d10fe737"
	bodyStrengthInfluenceRuleID             = "bazi.body-strength.influence-evidence.v2"
	bodyStrengthInfluenceProfile            = "local_visible_stem_four_branch_influence_v2"
	bodyStrengthSignedNormalizationFormula  = "centered_logistic_v1"
	bodyStrengthSupportNormalizationFormula = "zero_origin_logistic_v1"
	bodyStrengthAdjustmentForceRuleID       = "bazi.body-strength.adjustment-force.v1"
	bodyStrengthAdjustmentForceProfile      = "local_posterior_force_v1"
	bodyStrengthCompleteGroupRuleID         = "bazi.body-strength.adjustment.complete-same-element-branch-group.v1"
)

var yueLingDayElementOrder = [5]string{"木", "火", "土", "金", "水"}

var yueLingMonthBranchOrder = [12]string{
	"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥",
}

var bodyStrengthYangRenStemOrder = [5]string{"甲", "丙", "戊", "庚", "壬"}

var bodyStrengthYangRenBranches = [5]string{"卯", "午", "午", "酉", "子"}

var elementIdx = map[string]int{"木": 0, "火": 1, "土": 2, "金": 3, "水": 4}

var monthBranchIdx = map[string]int{
	"子": 0, "丑": 1, "寅": 2, "卯": 3,
	"辰": 4, "巳": 5, "午": 6, "未": 7,
	"申": 8, "酉": 9, "戌": 10, "亥": 11,
}

var tianGanMap = map[string]struct {
	WuXing string
}{
	"甲": {"木"}, "乙": {"木"},
	"丙": {"火"}, "丁": {"火"},
	"戊": {"土"}, "己": {"土"},
	"庚": {"金"}, "辛": {"金"},
	"壬": {"水"}, "癸": {"水"},
}

// yueLingMatrix is indexed by day-master element and the twelve month
// branches in 子丑寅卯辰巳午未申酉戌亥 order. It is intentionally explicit:
// 旺相休囚死 describes the day master's seasonal state, not the direction in
// which the day master generates or controls the month's element.
//
// Scores: 旺=3, 相=2, 休=1, 囚=0.5, 死=0.
// The four earth branches use the current profile's simplified 土旺口径;
// month-command depth can be introduced later as a separate profile rule.
var yueLingMatrix = [5][12]float64{
	// 子   丑   寅   卯   辰   巳   午   未   申   酉   戌   亥
	{2, 0.5, 3, 3, 0.5, 1, 1, 0.5, 0, 0, 0.5, 2}, // 木
	{0, 1, 2, 2, 1, 3, 3, 1, 0.5, 0.5, 1, 0},     // 火
	{0.5, 3, 0, 0, 3, 2, 2, 3, 1, 1, 3, 0.5},     // 土
	{1, 2, 0.5, 0.5, 2, 0, 0, 2, 3, 3, 2, 1},     // 金
	{3, 0, 1, 1, 0, 0.5, 0.5, 0, 2, 2, 0, 3},     // 水
}

type yueLingMatrixHashPayload struct {
	DayElementOrder  [5]string      `json:"day_element_order"`
	MonthBranchOrder [12]string     `json:"month_branch_order"`
	Scores           [5][12]float64 `json:"scores"`
}

func yueLingMatrixSHA256() string {
	payload, err := json.Marshal(yueLingMatrixHashPayload{
		DayElementOrder:  yueLingDayElementOrder,
		MonthBranchOrder: yueLingMonthBranchOrder,
		Scores:           yueLingMatrix,
	})
	if err != nil {
		panic(fmt.Sprintf("marshal yue-ling matrix: %v", err))
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(payload))
}

func verifiedYueLingMatrixSHA256() string {
	if got := yueLingMatrixSHA256(); got != yueLingTableSHA256 {
		panic(fmt.Sprintf("yue-ling matrix hash = %s, want pinned %s", got, yueLingTableSHA256))
	}
	return yueLingTableSHA256
}

type bodyStrengthBonusHashPayload struct {
	DayStemOrder       [10]string              `json:"day_stem_order"`
	LuBranches         [10]string              `json:"lu_branches"`
	YangRenStemOrder   [5]string               `json:"yang_ren_stem_order"`
	YangRenBranches    [5]string               `json:"yang_ren_branches"`
	Scores             BodyStrengthBonusScores `json:"scores"`
	YinStemBladePolicy string                  `json:"yin_stem_blade_policy"`
}

func bodyStrengthBonusTableSHA256(config BodyStrengthBonusRuleConfig) string {
	payload, err := json.Marshal(bodyStrengthBonusHashPayload{
		DayStemOrder:       config.DayStemOrder,
		LuBranches:         config.LuBranches,
		YangRenStemOrder:   config.YangRenStemOrder,
		YangRenBranches:    config.YangRenBranches,
		Scores:             config.Scores,
		YinStemBladePolicy: config.YinStemBladePolicy,
	})
	if err != nil {
		panic(fmt.Sprintf("marshal body-strength bonus table: %v", err))
	}
	return fmt.Sprintf("sha256:%x", sha256.Sum256(payload))
}

func verifiedBodyStrengthBonusSHA256(config BodyStrengthBonusRuleConfig) string {
	if got := bodyStrengthBonusTableSHA256(config); got != bodyStrengthBonusSHA256 {
		panic(fmt.Sprintf("body-strength bonus table hash = %s, want pinned %s", got, bodyStrengthBonusSHA256))
	}
	return bodyStrengthBonusSHA256
}

func getYueLingScore(dayElem, monthBranch string) float64 {
	di, dayOK := elementIdx[dayElem]
	mi, monthOK := monthBranchIdx[monthBranch]
	if !dayOK || !monthOK {
		return 0
	}
	return yueLingMatrix[di][mi]
}

func yueLingState(score float64) string {
	switch score {
	case 3:
		return "旺"
	case 2:
		return "相"
	case 1:
		return "休"
	case 0.5:
		return "囚"
	default:
		return "死"
	}
}

func yueLingEvidenceReason(dayElement, monthBranch string, score float64) string {
	reason := fmt.Sprintf("%s日主遇%s月，季令状态为%s，得令原始分%.2f。", dayElement, monthBranch, yueLingState(score), score)
	if inStrings(monthBranch, "丑", "辰", "未", "戌") {
		reason += " 四库月含余气、中气与土本气，当前Profile未按节气深浅分日司令，整月状态仅为简化候选。"
	}
	return reason
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

// zangGanWeight returns the configured藏干 weight for an earth-branch position.
func zangGanWeight(hsType tyme.HideHeavenStemType, weights BodyStrengthHideStemWeights) float64 {
	switch hsType {
	case tyme.MAIN:
		return weights.Main
	case tyme.MIDDLE:
		return weights.Middle
	case tyme.RESIDUAL:
		return weights.Residual
	}
	return 0.0
}

// changShengWeight returns the twelve-stage weight for a day stem and branch.
// HeavenStem.GetTerrain applies the selected profile's 阳顺阴逆 table; reducing
// the input to five elements would incorrectly give every yin stem a yang path.
// 长生/帝旺/临官 → 1.5, 沐浴/冠带/衰/墓 → 1.0, 胎/养/病/死 → 0.5, 绝 → 0
func changShengWeight(dayStem tyme.HeavenStem, branch tyme.EarthBranch, weights BodyStrengthTerrainWeights) (string, float64) {
	stage := dayStem.GetTerrain(branch).GetName()
	switch stage {
	case "长生":
		return stage, weights.ChangSheng
	case "沐浴":
		return stage, weights.MuYu
	case "冠带":
		return stage, weights.GuanDai
	case "临官":
		return stage, weights.LinGuan
	case "帝旺":
		return stage, weights.DiWang
	case "衰":
		return stage, weights.Shuai
	case "病":
		return stage, weights.Bing
	case "死":
		return stage, weights.Si
	case "墓":
		return stage, weights.Mu
	case "绝":
		return stage, weights.Jue
	case "胎":
		return stage, weights.Tai
	case "养":
		return stage, weights.Yang
	}
	return stage, 1.0
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

func branchForStem(stem string, stems []string, branches []string) (string, bool) {
	for i, candidate := range stems {
		if candidate == stem && i < len(branches) {
			return branches[i], true
		}
	}
	return "", false
}

func calculateBodyStrengthBonus(config BodyStrengthBonusRuleConfig, dayStem, dayBranch, monthBranch string) (float64, []BodyStrengthEvidence) {
	total := 0.0
	evidence := make([]BodyStrengthEvidence, 0, 4)
	appendBonus := func(ruleSuffix, source, item, reason string, score float64) {
		total += score
		evidence = append(evidence, BodyStrengthEvidence{
			RuleID:    config.RuleID + "." + ruleSuffix,
			Component: "bonus",
			Polarity:  "support",
			Source:    source,
			Item:      item,
			Score:     score,
			Reason:    reason,
		})
	}

	if luBranch, ok := branchForStem(dayStem, config.DayStemOrder[:], config.LuBranches[:]); ok {
		if dayBranch == luBranch {
			appendBonus("day-lu", "日支", dayBranch, "日支坐禄，按本地禄刃 Profile 加成。", config.Scores.DayLu)
		}
		if monthBranch == luBranch {
			appendBonus("month-lu", "月支", monthBranch, "月支建禄，按本地禄刃 Profile 加成。", config.Scores.MonthLu)
		}
	}
	if renBranch, ok := branchForStem(dayStem, config.YangRenStemOrder[:], config.YangRenBranches[:]); ok {
		if dayBranch == renBranch {
			appendBonus("day-yang-ren", "日支", dayBranch, "日支坐阳刃，按本地禄刃 Profile 加成。", config.Scores.DayYangRen)
		}
		if monthBranch == renBranch {
			appendBonus("month-yang-ren", "月支", monthBranch, "月支见阳刃，按本地禄刃 Profile 加成。", config.Scores.MonthYangRen)
		}
	}
	return total, evidence
}

type bodyStrengthInfluenceContribution struct {
	TenGod   string
	Polarity string
	Owner    string
	Score    float64
}

func calculateBodyStrengthInfluence(config BodyStrengthInfluenceRuleConfig, dayStem, candidateStem string, base float64) (bodyStrengthInfluenceContribution, bool) {
	if _, ok := tianGanMap[dayStem]; !ok {
		return bodyStrengthInfluenceContribution{}, false
	}
	if _, ok := tianGanMap[candidateStem]; !ok {
		return bodyStrengthInfluenceContribution{}, false
	}
	contribution := bodyStrengthInfluenceContribution{TenGod: ClassifyTenGod(candidateStem, dayStem, false)}
	switch contribution.TenGod {
	case "比肩":
		contribution.Polarity = "support"
		contribution.Owner = "shi"
		contribution.Score = base * config.SamePolarityPeerWeight
	case "劫财":
		contribution.Polarity = "support"
		contribution.Owner = "shi"
		contribution.Score = base * config.OppositePolarityPeerWeight
	case "正官", "七杀":
		contribution.Polarity = "restrict"
		contribution.Owner = "shi"
		contribution.Score = -base * config.OfficerKillerWeight
	case "食神", "伤官":
		contribution.Polarity = "restrict"
		contribution.Owner = "shi"
		contribution.Score = -base * config.OutputWeight
	case "正财", "偏财":
		contribution.Polarity = "restrict"
		contribution.Owner = "shi"
		contribution.Score = -base * config.WealthWeight
	case "正印", "偏印":
		contribution.Polarity = "support"
		contribution.Owner = config.SealOwnership
		return contribution, false
	default:
		return bodyStrengthInfluenceContribution{}, false
	}
	return contribution, true
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

func normalizeBodyStrengthScore(formula string, score, divisor float64) float64 {
	if divisor <= 0 {
		panic("body-strength normalization divisor must be positive")
	}
	centered := 1.0 / (1.0 + math.Exp(-score/divisor))
	switch formula {
	case bodyStrengthSignedNormalizationFormula:
		return centered
	case bodyStrengthSupportNormalizationFormula:
		return normalizedBodyScore(2*centered - 1)
	default:
		panic(fmt.Sprintf("unknown body-strength normalization formula %q", formula))
	}
}

func scoreBandCandidateForBodyStrength(score float64) string {
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

func bodyStrengthBandRules() []BodyStrengthBandRule {
	return []BodyStrengthBandRule{
		{Candidate: "身旺", Operator: "gt", Threshold: 0.7},
		{Candidate: "偏旺", Operator: "gt", Threshold: 0.5},
		{Candidate: "中和", Operator: "gt", Threshold: 0.4},
		{Candidate: "偏弱", Operator: "gt", Threshold: 0.3},
		{Candidate: "身弱", Operator: "otherwise"},
	}
}

func bodyStrengthInputSnapshot(ec *tyme.EightChar) BodyStrengthInputSnapshot {
	cycles := []tyme.SixtyCycle{ec.GetYear(), ec.GetMonth(), ec.GetDay(), ec.GetHour()}
	pillars := make([]string, 0, len(cycles))
	for _, cycle := range cycles {
		pillars = append(pillars, cycle.GetName())
	}
	dayStem := ec.GetDay().GetHeavenStem()
	return BodyStrengthInputSnapshot{
		Pillars:     pillars,
		DayStem:     dayStem.GetName(),
		DayElement:  dayStem.GetElement().GetName(),
		MonthBranch: ec.GetMonth().GetEarthBranch().GetName(),
	}
}

func completeSameElementBranchGroups(ec *tyme.EightChar, dayElement string) []ZhiRelation {
	cycles := []tyme.SixtyCycle{ec.GetYear(), ec.GetMonth(), ec.GetDay(), ec.GetHour()}
	keys := [...]string{"year", "month", "day", "hour"}
	labels := [...]string{labelYear, labelMonth, labelDay, labelHour}
	pillars := make([]branchRelationPillar, 0, len(cycles))
	for i, cycle := range cycles {
		pillars = append(pillars, branchRelationPillar{
			key: keys[i], label: labels[i], branch: cycle.GetEarthBranch().GetName(),
		})
	}

	groups := make([]ZhiRelation, 0, 1)
	for _, relation := range buildZhiRelationGraph(pillars) {
		if relation.StructureStatus == "complete_structure" && relation.TargetElement == dayElement &&
			(relation.Type == "三合局" || relation.Type == "三会局") {
			groups = append(groups, relation)
		}
	}
	return groups
}

func calcBodyStrengthV2(ec *tyme.EightChar) BodyStrengthResult {
	dayStem := ec.GetDay().GetHeavenStem()
	dayElem := dayStem.GetElement().GetName()

	monthBranch := ec.GetMonth().GetEarthBranch()
	monthBranchName := monthBranch.GetName()

	rules := bodyStrengthRuleConfig()
	weights := rules.Weights
	norms := rules.Normalizers
	thresholds := rules.AdjustmentThresholds
	rootRules := rules.Root
	bonusRules := rules.Bonus
	influenceRules := rules.Influence
	adjustmentForce := rules.AdjustmentForce

	// 1. 得令
	lingScore := getYueLingScore(dayElem, monthBranchName)
	evidence := []BodyStrengthEvidence{
		{
			Component: "ling",
			Polarity:  "support",
			Source:    "月令",
			Item:      monthBranch.GetName(),
			Score:     lingScore,
			Reason:    yueLingEvidenceReason(dayElem, monthBranchName, lingScore),
		},
	}

	pillars := [](func() tyme.SixtyCycle){ec.GetYear, ec.GetMonth, ec.GetDay, ec.GetHour}
	completeSupportingGroups := completeSameElementBranchGroups(ec, dayElem)
	for _, group := range completeSupportingGroups {
		evidence = append(evidence, BodyStrengthEvidence{
			RuleID:    bodyStrengthCompleteGroupRuleID + ".observed",
			Component: "structure",
			Polarity:  "support",
			Source:    group.Type,
			Item:      group.Subtype,
			Score:     0,
			Reason: fmt.Sprintf(
				"%s三支齐备，形成目标五行为%s的%s完整结构，与%s日主同气；只据此确认结构扶身，不宣称已经成化。",
				group.Subtype, group.TargetElement, group.Type, dayElem,
			),
		})
	}

	// 2. 得地：四柱地支藏干中仅统计同五行（通根），印星归入得生
	// 经典依据：滴天髓"通根者如甲木见寅卯"，通根仅指同五行
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
		csStage, csW := changShengWeight(dayStem, branch, rootRules.TerrainWeights)
		for _, hhs := range branch.GetHideHeavenStems() {
			hName := hhs.GetHeavenStem().GetName()
			if isSameElement(hName, dayElem) {
				w := zangGanWeight(hhs.GetType(), rootRules.HideStemWeights) * rootRules.RootMultiplier * csW
				touGan := isTougan(hName)
				if touGan {
					w *= rootRules.TouGanMultiplier
				}
				reason := fmt.Sprintf("%s支%s为%s，属%s同气通根；%s日干临%s为%s，长生权重%.1f。", branchName, hName, hideStemTypeLabel(hhs.GetType()), dayElem, dayStem.GetName(), branchName, csStage, csW)
				if touGan {
					reason += fmt.Sprintf(" 四柱天干见同干，按%s乘数%.2f计透干加成。", rootRules.TouGanScope, rootRules.TouGanMultiplier)
				}
				diScore += w
				evidence = append(evidence, BodyStrengthEvidence{
					RuleID:    rootRules.RuleID + ".lookup",
					Component: "di",
					Polarity:  "support",
					Source:    branchName,
					Item:      hName,
					Score:     w,
					Reason:    reason,
				})
			}
		}
	}

	// 专禄/建禄/阳刃加成
	// 经典依据：渊海子平"甲木得寅为禄，乙木得卯为禄"，专禄建禄力量殊胜
	// 改进：专禄/建禄/阳刃加成直接加到总分上，不被归一化稀释
	dayBranchName := ec.GetDay().GetEarthBranch().GetName()
	dayGanName := dayStem.GetName()

	luBonus, bonusEvidence := calculateBodyStrengthBonus(bonusRules, dayGanName, dayBranchName, monthBranchName)
	evidence = append(evidence, bonusEvidence...)

	// 3. 得势：年月时三干的比劫帮扶与克泄耗，加四支藏干的克泄耗。
	// 四支同气藏干已经完整归入得地，不能在这里再次作为比劫计分。
	// 经典依据：渊海子平"天干有比劫帮身有印绶生身"，力量有层级差异
	supportWeight := 0.0
	restrictWeight := 0.0
	for i, fn := range pillars {
		gan := fn().GetHeavenStem().GetName()
		if i == 2 { // 跳过日干本身
			continue
		}
		contribution, ok := calculateBodyStrengthInfluence(influenceRules, dayStem.GetName(), gan, 1)
		if !ok {
			continue
		}
		if contribution.Polarity == "support" {
			supportWeight += contribution.Score
		} else {
			restrictWeight += -contribution.Score
		}
		evidence = append(evidence, BodyStrengthEvidence{
			RuleID:    influenceRules.RuleID + ".visible-stem",
			Component: "shi",
			Polarity:  contribution.Polarity,
			Source:    []string{"年干", "月干", "日干", "时干"}[i],
			Item:      gan,
			Score:     contribution.Score,
			Reason:    fmt.Sprintf("%s透出%s，按%s Profile 计入得势%s力量%.2f。", gan, contribution.TenGod, influenceRules.Profile, contribution.Polarity, contribution.Score),
		})
	}
	// 四支藏干仅保留克泄耗；同气通根归得地，印星生扶归得生。
	branchLabels := [...]string{"年支藏干", "月支藏干", "日支藏干", "时支藏干"}
	for i, fn := range pillars {
		branch := fn().GetEarthBranch()
		for _, hhs := range branch.GetHideHeavenStems() {
			hiddenGan := hhs.GetHeavenStem().GetName()
			base := zangGanWeight(hhs.GetType(), rootRules.HideStemWeights) * influenceRules.HiddenBranchMultiplier
			contribution, ok := calculateBodyStrengthInfluence(influenceRules, dayStem.GetName(), hiddenGan, base)
			if !ok || contribution.Polarity != "restrict" {
				continue
			}
			restrictWeight += -contribution.Score
			evidence = append(evidence, BodyStrengthEvidence{
				RuleID:    influenceRules.RuleID + ".hidden-branch",
				Component: "shi",
				Polarity:  contribution.Polarity,
				Source:    branchLabels[i],
				Item:      hiddenGan,
				Score:     contribution.Score,
				Reason:    fmt.Sprintf("%s支%s藏%s为%s，按%s范围计克泄耗%.2f。", []string{"年", "月", "日", "时"}[i], branch.GetName(), hiddenGan, contribution.TenGod, influenceRules.HiddenBranchScope, contribution.Score),
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
				score := zangGanWeight(hhs.GetType(), rootRules.HideStemWeights)
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
	normShi := normalizeBodyStrengthScore(norms.ShiFormula, shiScore, norms.ShiSigmoidDivisor)
	normSheng := normalizeBodyStrengthScore(norms.ShengFormula, shengScore, norms.ShengSigmoidDivisor)
	totalScore := normLing*weights.Ling + normDi*weights.Di + normShi*weights.Shi + normSheng*weights.Sheng + luBonus
	components := []BodyStrengthComponent{
		{Key: "ling", Name: "得令", RawScore: lingScore, NormalizedScore: normalizedBodyScore(normLing), Weight: weights.Ling, WeightedScore: normLing * weights.Ling, Description: "月令为提纲，按日主五行与月支五行旺相休囚死取分。"},
		{Key: "di", Name: "得地", RawScore: diScore, NormalizedScore: normalizedBodyScore(normDi), Weight: weights.Di, WeightedScore: normDi * weights.Di, Description: "四支藏干中同五行通根，结合本中余气、十日干阳顺阴逆长生权重和透干加成。"},
		{Key: "shi", Name: "得势", RawScore: shiScore, NormalizedScore: normalizedBodyScore(normShi), Weight: weights.Shi, WeightedScore: normShi * weights.Shi, Description: fmt.Sprintf("年月时天干的比劫帮扶与克泄耗，加四支藏干的克泄耗；按%s归一化。", norms.ShiFormula)},
		{Key: "sheng", Name: "得生", RawScore: shengScore, NormalizedScore: normalizedBodyScore(normSheng), Weight: weights.Sheng, WeightedScore: normSheng * weights.Sheng, Description: fmt.Sprintf("天干和地支藏干中的印星生扶力量；按%s归一化，零证据映射为零。", norms.ShengFormula)},
		{Key: "bonus", Name: "禄刃加成", RawScore: luBonus, NormalizedScore: normalizedBodyScore(luBonus), Weight: weights.Bonus, WeightedScore: luBonus, Description: "专禄、建禄、阳刃、月刃直接加成。"},
	}

	// 5. 后验修正：失令不衰。克泄耗已经完整进入得势，不再以另一套
	// 不可比的原始力值重复执行“得令不旺”整体折减。
	supportingMap := map[string]string{
		"木": "水", "火": "木", "土": "火", "金": "土", "水": "金",
	}
	adjustments := make([]BodyStrengthAdjustment, 0, 2)
	if lingScore <= 0.5 { // 失令
		spElem := supportingMap[dayElem]
		supportingForce := 0.0
		for i, fn := range pillars {
			sc := fn()
			stElem := sc.GetHeavenStem().GetElement().GetName()
			if i != 2 && (stElem == spElem || stElem == dayElem) {
				supportingForce += adjustmentForce.StemForce
			}
			for _, hhs := range sc.GetEarthBranch().GetHideHeavenStems() {
				hElem := hhs.GetHeavenStem().GetElement().GetName()
				if hElem == spElem || hElem == dayElem {
					supportingForce += zangGanWeight(hhs.GetType(), rootRules.HideStemWeights) * adjustmentForce.HiddenStemMultiplier
				}
			}
		}
		// 生扶力量达到当前 Profile 阈值时触发修正。
		if supportingForce >= thresholds.ShiLingSupportForce {
			before := totalScore
			totalScore = totalScore*thresholds.ShiLingBlendSelf + adjustmentForce.NeutralTarget*thresholds.ShiLingBlendNeutral
			adjustments = append(adjustments, BodyStrengthAdjustment{
				Name:        "失令不衰",
				Before:      before,
				After:       totalScore,
				Reason:      fmt.Sprintf("生扶%s及同气%s力量合计%.2f，达到修正阈值%.2f。", spElem, dayElem, supportingForce, thresholds.ShiLingSupportForce),
				Description: "虽失月令但命局另有生扶根气；按修正后总分重新选择分段候选。",
			})
		}
	}

	// 完整同气三合、三会是独立于逐支藏干的整体结构。《滴天髓阐微》
	// 对水、木、火、金命例均明确以“会局帮身”“不当弱论”裁决。这里
	// 不追加经验权重，也不宣称成化；仅当加权结果仍落在弱侧时，复用
	// 当前 Profile 的中和目标作为下限，避免逐支计分遗漏整体结构。
	if len(completeSupportingGroups) > 0 && totalScore < adjustmentForce.NeutralTarget {
		before := totalScore
		totalScore = adjustmentForce.NeutralTarget
		groupNames := make([]string, 0, len(completeSupportingGroups))
		for _, group := range completeSupportingGroups {
			groupNames = append(groupNames, group.Subtype+group.Type)
		}
		adjustments = append(adjustments, BodyStrengthAdjustment{
			Name:        "同气成局不作弱论",
			Before:      before,
			After:       totalScore,
			Reason:      fmt.Sprintf("%s三支齐备且目标五行与%s日主相同；复用中和基准%.2f，不另加主观权重。", strings.Join(groupNames, "、"), dayElem, adjustmentForce.NeutralTarget),
			Basis:       "classical_complete_same_element_branch_group_neutral_floor_v1",
			Description: "完整同气三合或三会提供整体扶身结构；只阻止结果落入弱侧，不据此单独判断成化、格局或用神。",
		})
	}

	// The public band rules must always classify the final adjusted score. A
	// manual one-step upgrade/downgrade can disagree with these thresholds.
	scoreBandCandidate := scoreBandCandidateForBodyStrength(totalScore)

	for i := range components {
		components[i].RuleID = "bazi.body-strength.component." + components[i].Key
		if components[i].Key == "ling" {
			components[i].RuleID = yueLingRuleID
		} else if components[i].Key == "di" {
			components[i].RuleID = rootRules.RuleID
		} else if components[i].Key == "bonus" {
			components[i].RuleID = bonusRules.RuleID
		} else if components[i].Key == "shi" {
			components[i].RuleID = influenceRules.RuleID
		}
		components[i].Basis = "local_weighted_component_profile"
		components[i].Status = "observed"
		components[i].ValidationStatus = "not_validated"
	}
	for i := range evidence {
		if evidence[i].RuleID == "" {
			evidence[i].RuleID = fmt.Sprintf("bazi.body-strength.evidence.%s.%d", evidence[i].Component, i+1)
		}
		if evidence[i].Component == "ling" {
			evidence[i].RuleID = yueLingRuleID + ".lookup"
		}
		evidence[i].Basis = "local_component_scoring_rule"
		evidence[i].Status = "observed"
		evidence[i].InterpretationStatus = "not_adjudicated"
	}
	for i := range adjustments {
		switch adjustments[i].Name {
		case "失令不衰":
			adjustments[i].RuleID = "bazi.body-strength.adjustment.shi-ling-bu-shuai.v1"
		case "同气成局不作弱论":
			adjustments[i].RuleID = bodyStrengthCompleteGroupRuleID
		default:
			adjustments[i].RuleID = fmt.Sprintf("bazi.body-strength.adjustment.%d", i+1)
		}
		if adjustments[i].Basis == "" {
			adjustments[i].Basis = adjustmentForce.Profile
		}
		adjustments[i].Status = "observed"
		adjustments[i].ValidationStatus = "not_validated"
	}

	limitations := []string{
		"component weights and normalizers are local profile parameters without Gold calibration",
		"score-band thresholds and posterior adjustments are not learned from Train Gold",
		"officer-killer restriction is counted once in the influence component without a second whole-score multiplier",
		"the complete same-element branch-group floor is supported by four classical element cases but still awaits expert Gold validation",
		"the score-band candidate does not determine favorable elements or real-world outcomes",
	}
	if inStrings(monthBranchName, "丑", "辰", "未", "戌") {
		limitations = append(limitations, "earth-month seasonal scoring is an unsegmented whole-month candidate; classical day-command profiles differ and are not adjudicated")
	}

	return BodyStrengthResult{
		RuleID:               bodyStrengthRuleID,
		SchemaVersion:        BodyStrengthSchemaVersion,
		RuleVersion:          RuleVersion,
		School:               RuleSchool,
		ScoringProfile:       bodyStrengthScoringProfile,
		YueLingRuleID:        rules.YueLing.RuleID,
		YueLingProfile:       rules.YueLing.Profile,
		YueLingTableSHA256:   rules.YueLing.TableSHA256,
		Inputs:               bodyStrengthInputSnapshot(ec),
		ScoreBandCandidate:   scoreBandCandidate,
		BandSelectionBasis:   "ordered_fixed_local_thresholds_then_posterior_adjustments",
		BandRules:            bodyStrengthBandRules(),
		TotalScore:           totalScore,
		LingScore:            lingScore,
		DiScore:              diScore,
		ShiScore:             shiScore,
		ShengScore:           shengScore,
		LuBonus:              luBonus,
		Components:           components,
		Evidence:             evidence,
		Adjustments:          adjustments,
		Status:               "observed",
		ValidationStatus:     "not_validated",
		InterpretationStatus: "not_adjudicated",
		IsStrengthConclusion: false,
		Limitations:          limitations,
	}
}
