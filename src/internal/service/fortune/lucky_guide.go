package fortune

import (
	"fmt"
	"strings"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/data"
)

var guideElementColors = map[string][]string{
	"木": {"绿色系", "青色系"},
	"火": {"红色系", "紫色系"},
	"土": {"黄色系", "棕色系"},
	"金": {"白色系", "金色系"},
	"水": {"黑色系", "蓝色系"},
}

var guideElementNumbers = map[string][]string{
	"木": {"3", "8"},
	"火": {"2", "7"},
	"土": {"5", "0"},
	"金": {"4", "9"},
	"水": {"1", "6"},
}

var guideElementDirections = map[string]string{
	"木": "东方",
	"火": "南方",
	"土": "中宫",
	"金": "西方",
	"水": "北方",
}

var guideElementAvoidDirections = map[string]string{
	"木": "西方",
	"火": "北方",
	"土": "东方",
	"金": "南方",
	"水": "中宫",
}

var guideElementObjects = map[string]string{
	"木": "绿植、木质文具、清晨步行",
	"火": "暖光、红紫配饰、公开表达",
	"土": "陶瓷、黄色便签、整理桌面",
	"金": "金属笔、白金色配饰、清单工具",
	"水": "水杯、蓝黑色配饰、复盘记录",
}

var guideTenGodActions = map[string][]string{
	"same":    {"会友", "合作", "复盘资源"},
	"shengWo": {"学习", "请教", "整理文书"},
	"woSheng": {"创作", "表达", "交付成果"},
	"keWo":    {"守规矩", "汇报", "处理责任"},
	"woKe":    {"交易", "纳财", "预算规划"},
}

// BuildFortuneGuide creates a structured, explainable guide for daily fortune.
func BuildFortuneGuide(
	chart *bazipkg.BaziResult,
	dayPillar model.Pillar,
	score int,
	stemRel string,
	branchRel string,
	luckyColor string,
	luckyNumbers []int,
	wealthDir string,
	auspiciousHours []string,
	yi []model.YiJiItem,
	ji []model.YiJiItem,
	rikuyo *RikuyoResult,
) *model.FortuneGuide {
	if chart == nil {
		return nil
	}

	like, dislike, special := getEffectiveFavor(chart)
	primary := firstElement(like, data.GanElement[dayPillar.Gan], data.GanElement[chart.DayPillar.Gan])
	secondary := secondElement(like, primary, data.GanElement[chart.DayPillar.Gan])
	avoidElem := firstElement(dislike, elementOvercomes[primary])
	todayElem := data.GanElement[dayPillar.Gan]
	dayMaster := chart.DayPillar.Gan + data.GanElement[chart.DayPillar.Gan]

	confidence := guideConfidence(score, special, rikuyo, len(like), branchRel)
	precision := "standard"
	if special || (rikuyo != nil && (rikuyo.YongShenImpact.Score != 0 || rikuyo.FavorScore != 0)) {
		precision = "pattern-aware"
	}

	colors := buildColorItems(primary, secondary, luckyColor, special)
	numbers := buildNumberItems(primary, secondary, luckyNumbers)
	actions := buildGuideActions(primary, stemRel, branchRel, yi, dayMaster)
	cautions := buildGuideCautions(avoidElem, stemRel, branchRel, ji)
	bestHours := buildBestHourItems(auspiciousHours, rikuyo)

	faceDir := model.FortuneGuideItem{
		Label:   "朝向",
		Value:   valueOrDefault(guideElementDirections[primary], "顺手方位"),
		Element: primary,
		Reason:  fmt.Sprintf("按今日有效喜用取%s，工作、沟通、学习时优先借%s气。", primary, primary),
	}
	wealthItem := model.FortuneGuideItem{
		Label:   "财位",
		Value:   valueOrDefault(wealthDir, guideElementDirections[secondary]),
		Element: secondary,
		Reason:  fmt.Sprintf("财位沿用流日天干财方，再以%s作辅助，适合放置账本、计划表或当天要推进的资料。", secondary),
	}
	avoidDir := model.FortuneGuideItem{
		Label:   "避开",
		Value:   valueOrDefault(guideElementAvoidDirections[avoidElem], "嘈杂方位"),
		Element: avoidElem,
		Reason:  fmt.Sprintf("忌神取%s，方位上少主动引动，重大沟通和决策不宜从此方起势。", avoidElem),
	}

	analysis := guideAnalysis(dayMaster, dayPillar, todayElem, primary, secondary, avoidElem, stemRel, branchRel, score, special, rikuyo)
	strategy := guideStrategy(score, stemRel, branchRel, primary)

	return &model.FortuneGuide{
		PrecisionLevel:     precision,
		Confidence:         confidence,
		PrimaryElement:     primary,
		SecondaryElement:   secondary,
		AvoidElement:       avoidElem,
		LuckyColors:        colors,
		LuckyNumbers:       numbers,
		FaceDirection:      faceDir,
		WealthDirection:    wealthItem,
		AvoidDirection:     avoidDir,
		RecommendedActions: actions,
		Cautions:           cautions,
		BestHours:          bestHours,
		Analysis:           analysis,
		Strategy:           strategy,
	}
}

func firstElement(groups ...interface{}) string {
	for _, group := range groups {
		switch v := group.(type) {
		case []string:
			for _, item := range v {
				item = strings.TrimSpace(item)
				if guideElementDirections[item] != "" {
					return item
				}
			}
		case string:
			v = strings.TrimSpace(v)
			if guideElementDirections[v] != "" {
				return v
			}
		}
	}
	return "土"
}

func secondElement(like []string, primary, fallback string) string {
	for _, item := range like {
		item = strings.TrimSpace(item)
		if item != "" && item != primary && guideElementDirections[item] != "" {
			return item
		}
	}
	if fallback != "" && fallback != primary && guideElementDirections[fallback] != "" {
		return fallback
	}
	return primary
}

func buildColorItems(primary, secondary, fallback string, special bool) []model.FortuneGuideItem {
	out := []model.FortuneGuideItem{}
	add := func(label, elem, reason string) {
		colors := guideElementColors[elem]
		if len(colors) == 0 {
			return
		}
		out = append(out, model.FortuneGuideItem{Label: label, Value: strings.Join(colors, "、"), Element: elem, Reason: reason})
	}
	source := "扶抑喜用"
	if special {
		source = "格局喜用"
	}
	add("主色", primary, fmt.Sprintf("主色取%s，来自%s，今天优先用在上衣、桌面或随身物。", primary, source))
	if secondary != "" && secondary != primary {
		add("辅色", secondary, fmt.Sprintf("辅色取%s，用来补足主色，适合小面积点缀。", secondary))
	}
	if len(out) == 0 && fallback != "" {
		out = append(out, model.FortuneGuideItem{Label: "参考色", Value: fallback, Reason: "未取到完整喜用颜色时沿用基础幸运色。"})
	}
	return out
}

func buildNumberItems(primary, secondary string, fallback []int) []model.FortuneGuideItem {
	out := []model.FortuneGuideItem{}
	add := func(label, elem string) {
		nums := guideElementNumbers[elem]
		if len(nums) == 0 {
			return
		}
		out = append(out, model.FortuneGuideItem{
			Label:   label,
			Value:   strings.Join(nums, "、"),
			Element: elem,
			Reason:  fmt.Sprintf("%s数取%s，适合用于提醒时间、座位、优先级或小额尾数。", elem, strings.Join(nums, "、")),
		})
	}
	add("主数", primary)
	if secondary != "" && secondary != primary {
		add("辅数", secondary)
	}
	if len(out) == 0 && len(fallback) > 0 {
		parts := make([]string, 0, len(fallback))
		for _, n := range fallback {
			parts = append(parts, fmt.Sprintf("%d", n))
		}
		out = append(out, model.FortuneGuideItem{Label: "参考数", Value: strings.Join(parts, "、"), Reason: "沿用基础幸运数字。"})
	}
	return out
}

func buildGuideActions(primary, stemRel, branchRel string, yi []model.YiJiItem, dayMaster string) []model.FortuneGuideItem {
	out := []model.FortuneGuideItem{{
		Label:   "补气",
		Value:   guideElementObjects[primary],
		Element: primary,
		Reason:  fmt.Sprintf("%s今日优先借%s气，动作宜小而持续，不宜只做象征。", dayMaster, primary),
	}}
	for _, act := range guideTenGodActions[stemRel] {
		out = append(out, model.FortuneGuideItem{
			Label:  "行动",
			Value:  act,
			Reason: stemRelationGuideReason(stemRel, act),
		})
		if len(out) >= 3 {
			break
		}
	}
	for _, item := range yi {
		if len(out) >= 4 {
			break
		}
		if hasGuideValue(out, item.Activity) {
			continue
		}
		out = append(out, model.FortuneGuideItem{Label: "宜事", Value: item.Activity, Reason: item.Reason})
	}
	if branchRel == "combine" || branchRel == "sanHe" || branchRel == "sanHui" {
		out = append(out, model.FortuneGuideItem{Label: "人事", Value: "主动联络", Reason: "今日地支有合会之象，人际协作比单打独斗更顺。"})
	}
	return out
}

func buildGuideCautions(avoidElem, stemRel, branchRel string, ji []model.YiJiItem) []model.FortuneGuideItem {
	out := []model.FortuneGuideItem{{
		Label:   "忌神",
		Value:   "少引动" + avoidElem,
		Element: avoidElem,
		Reason:  fmt.Sprintf("%s为今日需避的过量气，相关颜色、方位和冲动决策都宜减量。", avoidElem),
	}}
	switch branchRel {
	case "clash":
		out = append(out, model.FortuneGuideItem{Label: "冲动", Value: "避免硬碰硬", Reason: "日支六冲，事情容易冲散，先缓一拍再定。"})
	case "harm":
		out = append(out, model.FortuneGuideItem{Label: "暗耗", Value: "少做口头承诺", Reason: "六害偏暗，容易有误解或后续成本。"})
	case "punish":
		out = append(out, model.FortuneGuideItem{Label: "相刑", Value: "谨慎合同流程", Reason: "相刑主规矩、处罚与内耗，文书细节要复核。"})
	case "break":
		out = append(out, model.FortuneGuideItem{Label: "相破", Value: "不拆旧局", Reason: "六破不利强行破局，先修补后推进。"})
	}
	if stemRel == "keWo" {
		out = append(out, model.FortuneGuideItem{Label: "压力", Value: "不硬扛", Reason: "官杀压身，宜按规则处理，不宜情绪化对抗。"})
	}
	for _, item := range ji {
		if len(out) >= 4 {
			break
		}
		if hasGuideValue(out, item.Activity) {
			continue
		}
		out = append(out, model.FortuneGuideItem{Label: "忌事", Value: item.Activity, Reason: item.Reason})
	}
	return out
}

func buildBestHourItems(hours []string, rikuyo *RikuyoResult) []model.FortuneGuideItem {
	limit := len(hours)
	if limit > 3 {
		limit = 3
	}
	out := make([]model.FortuneGuideItem, 0, limit+1)
	for i := 0; i < limit; i++ {
		out = append(out, model.FortuneGuideItem{Label: "吉时", Value: hours[i], Reason: "按流日地支取吉时，适合安排需要结果的动作。"})
	}
	if rikuyo != nil && rikuyo.TwelveStage != "" {
		reason := "十二长生状态可作为当天精力节奏参考。"
		if rikuyo.StageDesc != "" {
			reason = rikuyo.StageDesc
		}
		out = append(out, model.FortuneGuideItem{Label: "节奏", Value: rikuyo.TwelveStage, Reason: reason})
	}
	return out
}

func guideAnalysis(dayMaster string, dayPillar model.Pillar, todayElem, primary, secondary, avoidElem, stemRel, branchRel string, score int, special bool, rikuyo *RikuyoResult) string {
	source := "身旺身弱喜忌"
	if special {
		source = "格局喜忌"
	}
	parts := []string{
		fmt.Sprintf("此日以%s为日主参考，流日%s%s，天干属%s。开运主轴取%s，是按%s而来；辅以%s，避%s。", dayMaster, dayPillar.Gan, dayPillar.Zhi, todayElem, primary, source, secondary, avoidElem),
	}
	if stemRel != "" {
		parts = append(parts, fmt.Sprintf("天干关系为%s，适合的行动要顺着这层十神气势来取。", stemRelLabel(stemRel, "", "")))
	}
	if branchRel != "" && branchRel != "neutral" {
		parts = append(parts, fmt.Sprintf("地支见%s，因此开运不只看颜色数字，还要看当天人事节奏和避忌。", branchRelLabel(branchRel)))
	}
	if rikuyo != nil && rikuyo.OverallVerdict != "" {
		parts = append(parts, "日课提示："+rikuyo.OverallVerdict)
	}
	if score < 45 {
		parts = append(parts, "分数偏低时，开运以减损为先，少冒进、多复核。")
	} else if score >= 75 {
		parts = append(parts, "分数较高时，开运以承接为主，把好时段用于推进关键事项。")
	}
	return strings.Join(parts, " ")
}

func guideStrategy(score int, stemRel, branchRel, primary string) string {
	switch {
	case score >= 80:
		return fmt.Sprintf("今日宜主动出击：用%s色系和对应吉时做开场，把重要沟通、提交、签约放在精神最稳的时段。", primary)
	case score >= 60:
		return fmt.Sprintf("今日宜稳中推进：先用%s气定方向，再挑一到两件最重要的事完成。", primary)
	case score >= 40:
		return fmt.Sprintf("今日宜守中求进：%s可作辅助，重点是避开冲动承诺和临时改计划。", primary)
	default:
		return fmt.Sprintf("今日以避险为先：%s只作轻量补气，重大决定延后，先处理收尾和复盘。", primary)
	}
}

func guideConfidence(score int, special bool, rikuyo *RikuyoResult, likeCount int, branchRel string) int {
	conf := 52
	if likeCount > 0 {
		conf += 10
	}
	if special {
		conf += 8
	}
	if rikuyo != nil {
		conf += 10
		if rikuyo.YongShenImpact.Score > 0 {
			conf += 5
		}
		if rikuyo.FavorScore >= 70 {
			conf += 5
		}
	}
	if branchRel != "neutral" && branchRel != "" {
		conf += 5
	}
	if score >= 75 || score <= 40 {
		conf += 5
	}
	if conf > 95 {
		return 95
	}
	if conf < 30 {
		return 30
	}
	return conf
}

func stemRelationGuideReason(stemRel, act string) string {
	switch stemRel {
	case "same":
		return "比劫当值，适合把资源和伙伴关系理顺，但边界要先说清。"
	case "shengWo":
		return "印星生扶，适合学习、请教、整理资料，让外部助力进来。"
	case "woSheng":
		return "食伤泄秀，适合表达、创作、交付，把想法变成结果。"
	case "keWo":
		return "官杀当令，适合按流程担责，不宜硬碰硬。"
	case "woKe":
		return "财星被我所用，适合预算、交易、谈资源，但忌贪快。"
	default:
		return "按今日气势选取可执行事项。"
	}
}

func hasGuideValue(items []model.FortuneGuideItem, value string) bool {
	for _, item := range items {
		if item.Value == value {
			return true
		}
	}
	return false
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}
