package fortune

import (
	"fmt"
	"sort"
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
	actions := buildGuideActions(primary, secondary, stemRel, branchRel, yi, dayMaster, score, auspiciousHours, rikuyo)
	cautions := buildGuideCautions(avoidElem, stemRel, branchRel, ji, score, rikuyo)
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

func buildGuideActions(primary, secondary, stemRel, branchRel string, yi []model.YiJiItem, dayMaster string, score int, hours []string, rikuyo *RikuyoResult) []model.FortuneGuideItem {
	timing := guideActionTiming(hours)
	out := []model.FortuneGuideItem{
		guideItem("补气", guideElementObjects[primary], primary,
			fmt.Sprintf("%s今日优先借%s气，动作宜小而持续，不宜只做象征。", dayMaster, primary),
			"五行补气", 96, guideActionIntensity(score), timing,
			guideElementActionMethod(primary), "有效喜用", fmt.Sprintf("补%s气，稳住今日主轴", primary)),
	}
	if secondary != "" && secondary != primary {
		out = append(out, guideItem("辅佐", guideElementObjects[secondary], secondary,
			fmt.Sprintf("%s为辅助喜用，适合用小动作补足主轴，不必大面积引动。", secondary),
			"五行辅助", 78, "中", timing,
			guideElementActionMethod(secondary), "辅助喜用", fmt.Sprintf("补%s气，增强执行余地", secondary)))
	}

	for i, act := range guideTenGodActions[stemRel] {
		out = append(out, guideItem("行动", act, "",
			stemRelationGuideReason(stemRel, act),
			"十神行动", 90-i*5, guideActionIntensity(score), timing,
			guideActionMethod(act, stemRel), stemRelLabel(stemRel, "", ""), stemRelationImpact(stemRel)))
	}

	for i, item := range yi {
		out = append(out, guideItem("宜事", item.Activity, "",
			item.Reason,
			"黄历宜事", 72-i*4, guideActionIntensity(score), timing,
			guideYiMethod(item.Activity), "黄历宜事", "顺势推进当天可成之事"))
	}

	if branchRel == "combine" || branchRel == "sanHe" || branchRel == "sanHui" {
		out = append(out, guideItem("人事", "主动联络", "",
			"今日地支有合会之象，人际协作比单打独斗更顺。",
			"地支合会", 86, "中", timing,
			"优先联系一个能推动结果的人，把请求说具体，并约定下一步时间。", branchRelLabel(branchRel), "借合会之气促成协作"))
	}
	if rikuyo != nil && rikuyo.StageFavorable && rikuyo.TwelveStage != "" {
		out = append(out, guideItem("节奏", "顺势推进", "",
			valueOrDefault(rikuyo.StageDesc, "十二长生状态较顺，可把关键事项放到精神最稳的时段。"),
			"日课节奏", 82, "中", timing,
			fmt.Sprintf("以%s的节奏安排任务：先做可见成果，再处理杂项。", rikuyo.TwelveStage), "十二长生", "把好节奏转成可交付结果"))
	}

	return normalizeGuideItems(out, 6)
}

func buildGuideCautions(avoidElem, stemRel, branchRel string, ji []model.YiJiItem, score int, rikuyo *RikuyoResult) []model.FortuneGuideItem {
	out := []model.FortuneGuideItem{
		guideItem("忌神", "少引动"+avoidElem, avoidElem,
			fmt.Sprintf("%s为今日需避的过量气，相关颜色、方位和冲动决策都宜减量。", avoidElem),
			"五行避忌", 96, guideCautionIntensity(score, branchRel), "全天减量",
			guideAvoidElementMethod(avoidElem), "忌神", fmt.Sprintf("减%s气，降低消耗和冲突", avoidElem)),
	}
	switch branchRel {
	case "clash":
		out = append(out, guideItem("冲动", "避免硬碰硬", "",
			"日支六冲，事情容易冲散，先缓一拍再定。",
			"地支风险", 92, "高", "沟通前先缓十分钟",
			"遇到反对先复述对方重点，再决定是否推进；今天不适合当场拍板。", "六冲", "降低冲突和返工"))
	case "harm":
		out = append(out, guideItem("暗耗", "少做口头承诺", "",
			"六害偏暗，容易有误解或后续成本。",
			"地支风险", 88, "中高", "承诺前留书面记录",
			"涉及钱、时间、人情的事先写清边界，避免只凭默契。", "六害", "减少误解和隐性成本"))
	case "punish":
		out = append(out, guideItem("相刑", "谨慎合同流程", "",
			"相刑主规矩、处罚与内耗，文书细节要复核。",
			"地支风险", 88, "中高", "签署前复核",
			"合同、审批、付款、交接都多看一遍，必要时请第三人校对。", "相刑", "减少规则成本"))
	case "break":
		out = append(out, guideItem("相破", "不拆旧局", "",
			"六破不利强行破局，先修补后推进。",
			"地支风险", 86, "中", "变更前先补漏",
			"先处理旧问题和遗留沟通，不在情绪上头时推倒重来。", "六破", "避免关系或计划破损"))
	}
	if stemRel == "keWo" {
		out = append(out, guideItem("压力", "不硬扛", "",
			"官杀压身，宜按规则处理，不宜情绪化对抗。",
			"十神风险", 84, "中高", "遇压先走流程",
			"用事实、节点、责任人回应压力，少用情绪和个人判断硬顶。", stemRelLabel(stemRel, "", ""), "降低压力外溢"))
	}
	if rikuyo != nil && !rikuyo.StageFavorable && rikuyo.TwelveStage != "" {
		out = append(out, guideItem("节奏", "避免强攻", "",
			valueOrDefault(rikuyo.StageDesc, "十二长生状态偏弱，今日不宜强行提速。"),
			"日课节奏", 80, "中", "关键事延后半拍",
			fmt.Sprintf("%s之日先守节奏，先做复核、收尾、准备，不急着开新局。", rikuyo.TwelveStage), "十二长生", "减少体力与决策透支"))
	}
	for i, item := range ji {
		out = append(out, guideItem("忌事", item.Activity, "",
			item.Reason,
			"黄历忌事", 72-i*4, guideCautionIntensity(score, branchRel), "全天避开",
			guideJiMethod(item.Activity), "黄历忌事", "避开当天阻力较大的事项"))
	}
	return normalizeGuideItems(out, 6)
}

func guideItem(label, value, element, reason, category string, priority int, intensity, timing, method, source, impact string) model.FortuneGuideItem {
	return model.FortuneGuideItem{
		Label:     label,
		Value:     valueOrDefault(value, label),
		Element:   element,
		Reason:    reason,
		Category:  category,
		Priority:  priority,
		Intensity: intensity,
		Timing:    timing,
		Method:    method,
		Source:    source,
		Impact:    impact,
	}
}

func normalizeGuideItems(items []model.FortuneGuideItem, limit int) []model.FortuneGuideItem {
	seen := map[string]bool{}
	out := make([]model.FortuneGuideItem, 0, len(items))
	for _, item := range items {
		key := strings.TrimSpace(item.Value)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, item)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Priority > out[j].Priority
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func guideActionTiming(hours []string) string {
	if len(hours) == 0 {
		return "上午定方向，午后推进"
	}
	limit := len(hours)
	if limit > 2 {
		limit = 2
	}
	return "优先" + strings.Join(hours[:limit], "、")
}

func guideActionIntensity(score int) string {
	switch {
	case score >= 80:
		return "强"
	case score >= 60:
		return "中高"
	case score >= 40:
		return "中"
	default:
		return "轻"
	}
}

func guideCautionIntensity(score int, branchRel string) string {
	if branchRel == "clash" || score < 40 {
		return "高"
	}
	if branchRel == "harm" || branchRel == "punish" || score < 60 {
		return "中高"
	}
	return "中"
}

func guideElementActionMethod(elem string) string {
	switch elem {
	case "木":
		return "把任务拆成三步，从最容易生长的一步开始；可用绿植、木质文具或步行启动状态。"
	case "火":
		return "先处理表达、展示、汇报类事项；用暖光或红紫小物提气，但控制节奏。"
	case "土":
		return "先整理桌面、文件和待办，把散乱事项放回固定位置，再推进正事。"
	case "金":
		return "用清单、计时器或规则表定边界，先确认标准再行动。"
	case "水":
		return "先复盘和记录，再沟通；用书面信息减少误差，保持流动但不拖延。"
	default:
		return "用一个可执行的小动作承接今日喜用，先稳住状态再推进。"
	}
}

func guideAvoidElementMethod(elem string) string {
	switch elem {
	case "木":
		return "少临时扩张、少开新分支，避免把简单事项越做越散。"
	case "火":
		return "少争辩、少赶场、少冲动曝光，情绪上来时先降温。"
	case "土":
		return "不要把问题都堆到自己身上，避免拖住进度和消化力。"
	case "金":
		return "少苛责、少硬切关系，规则要用来减负，不要变成对抗。"
	case "水":
		return "少拖延、少含糊承诺，重要信息不要只停留在脑中。"
	default:
		return "减少与忌神相关的颜色、方位和行为强度，先避损再求进。"
	}
}

func guideActionMethod(action, stemRel string) string {
	switch action {
	case "会友":
		return "约一个能带来信息或资源的人，话题聚焦在一个具体问题。"
	case "合作":
		return "先确认分工和边界，再推进共同事项，避免只讲情面。"
	case "复盘资源":
		return "盘点手头人脉、预算、工具，把可复用资源列出来。"
	case "学习":
		return "选择一份资料深读，做出摘要或下一步清单。"
	case "请教":
		return "带着具体问题请教，少问泛泛方向，多问下一步。"
	case "整理文书":
		return "整理合同、证件、资料夹，先补缺漏再提交。"
	case "创作":
		return "先输出草稿，不追求一次完美，用可见成果带动气势。"
	case "表达":
		return "把观点说清楚，控制情绪浓度，留下书面确认。"
	case "交付成果":
		return "优先完成一件可验收的小成果，再处理低优先级事项。"
	case "守规矩":
		return "按流程、标准、节点推进，越有压力越要留凭证。"
	case "汇报":
		return "先讲结论，再讲风险和下一步，让对方容易接住。"
	case "处理责任":
		return "把责任拆成可执行节点，逐项关闭，不要口头硬扛。"
	case "交易":
		return "先核对价格、交付和付款节点，小额试探优于一次押注。"
	case "纳财":
		return "适合收款、对账、谈资源，但要保留凭证。"
	case "预算规划":
		return "把预算分成必要、可选、延后三类，先控风险。"
	default:
		return stemRelationGuideReason(stemRel, action)
	}
}

func guideYiMethod(activity string) string {
	switch activity {
	case "出行", "移徙":
		return "先定路线和备选方案，把关键出发时段放在吉时附近。"
	case "会友", "宴饮":
		return "以轻量社交为宜，带着一个清楚目的去连接人。"
	case "开市", "交易", "纳财":
		return "先核对金额、交付和票据，再推进财务事项。"
	case "入学", "祈福", "祭祀":
		return "适合做定心、学习、立愿类动作，时间不必长但要专注。"
	case "求医", "针灸":
		return "适合预约、问诊、调养，带齐既往记录和问题清单。"
	default:
		return "把此事放在当天高能时段，小范围推进，完成一个可确认结果。"
	}
}

func guideJiMethod(activity string) string {
	switch activity {
	case "动土", "破土", "掘井", "开渠":
		return "涉及土地、结构、水土变动的事先延后，必要时只做检查不动工。"
	case "词讼":
		return "争议事项先收集证据和时间线，不急着正面对抗。"
	case "远行", "乘船", "渡水":
		return "长途、水路和不熟路线要准备备选方案，能改期则改期。"
	case "安葬", "行丧":
		return "肃穆事项宜从简从稳，避免临时加流程。"
	case "安门", "作灶", "上梁", "造屋":
		return "家宅结构相关事项先复核尺寸、方位和人员安排，避免仓促开工。"
	default:
		return "此事项阻力较高，能延后则延后，不能延后就降低规模并复核细节。"
	}
}

func stemRelationImpact(stemRel string) string {
	switch stemRel {
	case "same":
		return "整合资源，借同类之力"
	case "shengWo":
		return "引入支持，提升稳定度"
	case "woSheng":
		return "输出成果，释放表达力"
	case "keWo":
		return "承接责任，降低压力"
	case "woKe":
		return "整理财务，推进资源"
	default:
		return "顺应今日天干气势"
	}
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
