package data

import (
	"fmt"
	"bazi/internal/model"
)

// TianYanSuiEntry 滴天髓核心教义条目。
// 数据来源：《滴天髓阐微》原文解读与现代应用。
type TianYanSuiEntry struct {
	Category    string // 分类：强弱/格局/流通/调候/神煞
	Principle   string // 核心原理
	Judgment     string // 判断方法
	Description string // 详细描述
	FavorElements []string // 喜用五行
	TabooElements []string // 忌讳五行
	Fortune      string // 命运特征
}

// TianYanSuiKnowledge 滴天髓核心知识库。
var TianYanSuiKnowledge = map[string]TianYanSuiEntry{
	"配合如清": {
		Category:    "格局",
		Principle:   "天干地支配合得当，如清水般清澈",
		Judgment:    "喜忌分明，格局纯粹，无混杂混乱",
		Description: "配合如清是指四柱天干地支之间配合得当，五行不混杂，清纯而不浊。",
		FavorElements: []string{},
		TabooElements: []string{"混杂"},
		Fortune:      "配合清纯者贵，混杂者贱",
	},
	"甲木参天": {
		Category:    "调候",
		Principle:   "甲木参天，脱胎要火",
		Judgment:    "大树参天需要火来脱胎换骨，戊土来培根",
		Description: "甲木如大树，要成材需要火来雕琢（脱胎），土来培养。",
		FavorElements: []string{"火", "土"},
		TabooElements: []string{"金"},
		Fortune:      "甲木参天者贵，需火土配合",
	},
	"火明土秀": {
		Category:    "流通",
		Principle:   "火光明亮，土气秀美",
		Judgment:    "火土相生，光明磊落",
		Description: "火明土秀，火光明亮照耀，土气秀美养物。",
		FavorElements: []string{"火", "土"},
		TabooElements: []string{"水"},
		Fortune:      "火明土秀者，光明磊落，文章秀美",
	},
	"金水澄清": {
		Category:    "格局",
		Principle:   "金水清澄，不混浊",
		Judgment:    "金白水清，聪明秀气",
		Description: "金水清澄，不混杂。金白水清，聪明秀气之象。",
		FavorElements: []string{"金", "水"},
		TabooElements: []string{"土", "火"},
		Fortune:      "金水澄清者，聪明秀气，文才出众",
	},
	"水空则流": {
		Category:    "强弱",
		Principle:   "水空（地支无根）则流散",
		Judgment:    "水空不能聚物，流散不聚",
		Description: "水空指地支无根（截路空亡等），水气流散不能聚物。",
		FavorElements: []string{"金"},
		TabooElements: []string{"空亡"},
		Fortune:      "水空则流，财难聚，奔波劳碌",
	},
	"木火通明": {
		Category:    "流通",
		Principle:   "木火相生，文明之象",
		Judgment:    "木火通明，才智显露，文章秀美",
		Description: "木火相生，文明之象。才智显露，文章秀美。",
		FavorElements: []string{"木", "火"},
		TabooElements: []string{"水", "金"},
		Fortune:      "木火通明者，才华出众，文采飞扬",
	},
	"金木交关": {
		Category:    "流通",
		Principle:   "金木相克相成",
		Judgment:    "金木交关，关键在于通关",
		Description: "金木相克，需要有通关之神（水）来调和。",
		FavorElements: []string{"水"},
		TabooElements: []string{"无通关"},
		Fortune:      "金木交关，有通关者贵，无通关者贱",
	},
	"土金毓秀": {
		Category:    "流通",
		Principle:   "土金相生，毓秀之美",
		Judgment:    "土金相生，厚德载物，秀外慧中",
		Description: "土金相生，毓秀之美。厚德载物，秀外慧中。",
		FavorElements: []string{"土", "金"},
		TabooElements: []string{"木", "火"},
		Fortune:      "土金毓秀者，厚德载物，贵气内敛",
	},
	"水火既济": {
		Category:    "流通",
		Principle:   "水火相济，坎离相交",
		Judgment:    "水火既济，阴阳平衡",
		Description: "水火相济，坎离相交。既济卦象，阴阳平衡之象。",
		FavorElements: []string{"水", "火"},
		TabooElements: []string{"过偏"},
		Fortune:      "水火既济者，阴阳平衡，功成名就",
	},
	"燥湿调和": {
		Category:    "调候",
		Principle:   "燥土需水润，湿寒需火温",
		Judgment:    "燥湿调和为贵",
		Description: "燥土（夏季土）需水来润泽，湿寒（冬季）需火来温暖。",
		FavorElements: []string{"水", "火"},
		TabooElements: []string{"过燥", "过湿"},
		Fortune:      "燥湿调和者，运气通达",
	},
	"水木相生": {
		Category:    "流通",
		Principle:   "水木相生，生命力强",
		Judgment:    "水生木，生命力旺盛，创业立业",
		Description: "水为木之母，木为水之子，相生有情。",
		FavorElements: []string{"水", "木"},
		TabooElements: []string{"金", "火过旺"},
		Fortune:      "水木相生者，生命力强，宜创业发展",
	},
	"木土相制": {
		Category:    "流通",
		Principle:   "木克土，制化得当",
		Judgment:    "木克土需火通关，无火则凶",
		Description: "木克土为制服，忌土重无火通关。",
		FavorElements: []string{"火", "木"},
		TabooElements: []string{"土过旺", "无火"},
		Fortune:      "木土相制，有通关者贵",
	},
	"火金相克": {
		Category:    "流通",
		Principle:   "火克金，需要水通关",
		Judgment:    "火金相克，水通关则调和",
		Description: "火金交战，需水来通关调和。",
		FavorElements: []string{"水"},
		TabooElements: []string{"无通关"},
		Fortune:      "火金相克，有通关者贵",
	},
	"土水相激": {
		Category:    "流通",
		Principle:   "土克水，需木通关",
		Judgment:    "土水相激，木通关则吉",
		Description: "土克水为相战，需木来通关。",
		FavorElements: []string{"木"},
		TabooElements: []string{"无木"},
		Fortune:      "土水相激，有通关者吉",
	},
	"相生流通": {
		Category:    "流通",
		Principle:   "木->火->土->金->水->木 循环相生",
		Judgment:    "流通无阻，大吉之象",
		Description: "五行相生形成循环流通，源远流长。",
		FavorElements: []string{"木", "火", "土", "金", "水"},
		TabooElements: []string{"阻滞"},
		Fortune:      "流通顺畅者，运势恒通",
	},
	"循环流通": {
		Category:    "流通",
		Principle:   "五行循环相生，周而复始",
		Judgment:    "循环流通，如水之流",
		Description: "五行形成完整循环，生生不息。",
		FavorElements: []string{"木", "火", "土", "金", "水"},
		TabooElements: []string{"某行过旺", "某行过弱"},
		Fortune:      "循环流通者，命运曲折但坚韧",
	},
	"寒暖燥湿": {
		Category:    "调候",
		Principle:   "冬寒需火暖，夏燥需水润",
		Judgment:    "寒暖燥湿调和为贵",
		Description: "冬季水寒无火则寒，夏季火燥无水则枯。",
		FavorElements: []string{"火", "水"},
		TabooElements: []string{"过寒", "过燥"},
		Fortune:      "寒暖调和者，运势平稳",
	},
	"金生水旺": {
		Category:    "流通",
		Principle:   "金水相生，智慧流通",
		Judgment:    "金水流通，聪明睿智",
		Description: "金能生水，水为智慧，金水相生则文采风流。",
		FavorElements: []string{"金", "水"},
		TabooElements: []string{"土重", "火旺"},
		Fortune:      "金生水旺者，聪明多智",
	},
	"火土同旺": {
		Category:    "流通",
		Principle:   "火土相生，光明磊落",
		Judgment:    "火土同旺，诚实忠厚",
		Description: "火生土，土得火照，光明之象。",
		FavorElements: []string{"火", "土"},
		TabooElements: []string{"水", "木"},
		Fortune:      "火土同旺者，光明磊落，诚信可靠",
	},
}

// GetTianYanSuiEntry 返回滴天髓知识条目。
func GetTianYanSuiEntry(key string) (TianYanSuiEntry, bool) {
	e, ok := TianYanSuiKnowledge[key]
	return e, ok
}

// BalanceIndex 五行平衡指数判断。
type BalanceIndex struct {
	Element     string // 五行
	Score       int    // 综合得分
	NormalRange string // 正常范围
	Verdict     string // 判断：过旺/偏旺/中和/偏弱/过弱
}

// GetBalanceIndex 返回五行平衡判断。
func GetBalanceIndex(scores map[string]int) []BalanceIndex {
	var result []BalanceIndex
	elements := []string{"木", "火", "土", "金", "水"}
	normalRange := map[string][2]int{
		"木": {15, 35},
		"火": {15, 35},
		"土": {15, 35},
		"金": {15, 35},
		"水": {15, 35},
	}

	for _, elem := range elements {
		score := scores[elem]
		range_ := normalRange[elem]

		var verdict string
		if score < range_[0]-5 {
			verdict = "过弱"
		} else if score < range_[0] {
			verdict = "偏弱"
		} else if score > range_[1]+5 {
			verdict = "过旺"
		} else if score > range_[1] {
			verdict = "偏旺"
		} else {
			verdict = "中和"
		}

		result = append(result, BalanceIndex{
			Element:     elem,
			Score:       score,
			NormalRange: fmt.Sprintf("%d-%d", range_[0], range_[1]),
			Verdict:     verdict,
		})
	}
	return result
}

// FlowAnalysisResult 流通分析结果。
type FlowAnalysisResult struct {
	HasFlow    bool     // 是否有明显流通
	FlowType   string   // 流通类型：相生流通/相克流通/无流通
	FlowPath   []string // 流通路径
	IsBalanced bool     // 是否平衡
	Advice     string   // 建议
}

// AnalyzeWuXingFlow 分析五行流通情况。
func AnalyzeWuXingFlow(scores map[string]int) FlowAnalysisResult {
	order := []string{"木", "火", "土", "金", "水"}

	var flowPath []string
	var hasFlow bool

	for i := 0; i < len(order); i++ {
		curr := order[i]
		next := order[(i+1)%len(order)]
		if scores[curr] >= 10 && scores[next] >= 10 {
			flowPath = append(flowPath, fmt.Sprintf("%s生%s", curr, next))
			hasFlow = true
		}
	}

	flowType := "无流通"
	if len(flowPath) >= 3 {
		flowType = "相生流通"
	} else if hasFlow {
		flowType = "部分流通"
	}

	maxScore := 0
	minScore := int(^uint(0) >> 1)
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
		if s < minScore {
			minScore = s
		}
	}
	isBalanced := (maxScore - minScore) <= 15

	advice := "五行流通一般"
	if flowType == "相生流通" && isBalanced {
		advice = "五行流通顺畅，阴阳平衡，大吉之象"
	} else if flowType == "相生流通" {
		advice = "流通较好，但有些五行偏旺偏弱，需调和"
	} else if isBalanced {
		advice = "五行虽无明显流通，但相对平衡"
	} else {
		advice = "五行偏枯，需注意填补不足"
	}

	return FlowAnalysisResult{
		HasFlow:    hasFlow,
		FlowType:   flowType,
		FlowPath:   flowPath,
		IsBalanced: isBalanced,
		Advice:     advice,
	}
}

// GanZhiFlowMap 干支流通知识库。
var GanZhiFlowMap = map[string][]string{
	"甲木": {"丁火", "丙火", "癸水", "壬水", "戊土", "己土", "庚金", "辛金"},
	"乙木": {"丙火", "丁火", "癸水", "壬水", "戊土", "己土", "庚金", "辛金"},
	"丙火": {"甲木", "乙木", "壬水", "癸水", "戊土", "己土", "庚金", "辛金"},
	"丁火": {"甲木", "乙木", "壬水", "癸水", "戊土", "己土", "庚金", "辛金"},
	"戊土": {"丙火", "丁火", "庚金", "辛金", "癸水", "壬水", "甲木", "乙木"},
	"己土": {"丙火", "丁火", "庚金", "辛金", "癸水", "壬水", "甲木", "乙木"},
	"庚金": {"戊土", "己土", "壬水", "癸水", "甲木", "乙木", "丙火", "丁火"},
	"辛金": {"戊土", "己土", "壬水", "癸水", "甲木", "乙木", "丙火", "丁火"},
	"壬水": {"庚金", "辛金", "丙火", "丁火", "戊土", "己土", "甲木", "乙木"},
	"癸水": {"庚金", "辛金", "丙火", "丁火", "戊土", "己土", "甲木", "乙木"},
}

// WuXingFlowAnalysis 五行流通分析结果。
type WuXingFlowAnalysis struct {
	DayElement     string   `json:"day_element"`      // 日主五行
	FlowPaths      []string `json:"flow_paths"`       // 流通路径描述
	FlowType       string   `json:"flow_type"`        // 流通类型
	IsSmooth       bool     `json:"is_smooth"`        // 是否顺畅
	BlockedElement string   `json:"blocked_element"`   // 被阻滞的五行（空则无）
	BalanceVerdict string   `json:"balance_verdict"`  // 平衡判断
	Advice         string   `json:"advice"`           // 调候建议
}

// TongGuanAnalysis 通关用神分析。
type TongGuanAnalysis struct {
	HasTongGuan     bool     `json:"has_tong_guan"`      // 是否有通关
	TongGuanElement string   `json:"tong_guan_element"`   // 通关五行
	Weight          float64  `json:"weight"`             // 通关权重（0.0-1.0）
	Description     string   `json:"description"`        // 说明
}

// MissingElementAnalysis 缺失五行分析。
type MissingElementAnalysis struct {
	MissingElements []string         `json:"missing_elements"` // 缺失五行
	RemedyElements  []string         `json:"remedy_elements"`  // 补救五行（生缺失五行的）
	Severity        string           `json:"severity"`         // 缺失程度：轻微/中等/严重
	RemedyAdvice    []RemedyAdviceItem `json:"remedy_advice"`  // 补救建议
}

// RemedyAdviceItem 补救建议条目。
type RemedyAdviceItem struct {
	Element string `json:"element"` // 补救五行
	Direction string `json:"direction"` // 吉利方位
	Color   string `json:"color"`   // 吉利颜色
	Industry string `json:"industry"` // 适合行业
	Daily   string `json:"daily"`   // 日常建议
}

// DaYunFlowItem 大运流通影响。
type DaYunFlowItem struct {
	StartAge   int    `json:"start_age"`
	Pillar     string `json:"pillar"`
	FlowChange string `json:"flow_change"` // 增强/减弱/不变
	Impact     string `json:"impact"`      // 对喜忌的影响
}

// AnalyzeWuXingFlowV2 增强版五行流通分析。
func AnalyzeWuXingFlowV2(scores map[string]int, dayElem string) WuXingFlowAnalysis {
	order := []string{"木", "火", "土", "金", "水"}

	var flowPaths []string
	flowCount := 0
	var blockedElem string

	// 检测相生流通链
	for i := 0; i < len(order); i++ {
		curr := order[i]
		next := order[(i+1)%len(order)]
		if scores[curr] >= 10 && scores[next] >= 10 {
			flowPaths = append(flowPaths, fmt.Sprintf("%s生%s", curr, next))
			flowCount++
		} else if scores[curr] >= 10 && scores[next] < 10 {
			// 此五行旺但下一五行弱，提示可能有阻滞
			if blockedElem == "" {
				blockedElem = next
			}
		}
	}

	// 检测特殊流通格局
	flowType := detectFlowPattern(scores, dayElem)

	// 判断是否顺畅
	isSmooth := flowCount >= 3 && blockedElem == ""

	// 平衡判断
	maxScore, minScore := 0, 999
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
		if s < minScore {
			minScore = s
		}
	}
	balanceVerdict := "平衡"
	if maxScore-minScore > 20 {
		balanceVerdict = "偏枯"
	} else if maxScore-minScore > 15 {
		balanceVerdict = "欠平衡"
	}

	// 生成建议
	advice := generateFlowAdvice(scores, dayElem, isSmooth)

	return WuXingFlowAnalysis{
		DayElement:     dayElem,
		FlowPaths:      flowPaths,
		FlowType:       flowType,
		IsSmooth:       isSmooth,
		BlockedElement: blockedElem,
		BalanceVerdict: balanceVerdict,
		Advice:         advice,
	}
}

// detectFlowPattern 检测特殊流通格局。
func detectFlowPattern(scores map[string]int, dayElem string) string {
	// 木火通明：木火均旺
	if scores["木"] >= 10 && scores["火"] >= 10 {
		return "木火通明"
	}
	// 金水澄清：金水均旺
	if scores["金"] >= 10 && scores["水"] >= 10 {
		return "金水澄清"
	}
	// 土金毓秀：土金均旺
	if scores["土"] >= 10 && scores["金"] >= 10 {
		return "土金毓秀"
	}
	// 水火既济：水火均旺
	if scores["水"] >= 10 && scores["火"] >= 10 {
		return "水火既济"
	}
	// 火土相生：火土均旺
	if scores["火"] >= 10 && scores["土"] >= 10 {
		return "火明土秀"
	}
	// 金木交关：金木均旺
	if scores["金"] >= 10 && scores["木"] >= 10 {
		return "金木交关"
	}

	// 计算流通数量
	order := []string{"木", "火", "土", "金", "水"}
	flowCount := 0
	for i := 0; i < len(order); i++ {
		curr := order[i]
		next := order[(i+1)%len(order)]
		if scores[curr] >= 10 && scores[next] >= 10 {
			flowCount++
		}
	}

	if flowCount >= 3 {
		return "相生流通"
	}
	if isCircularFlow(scores) {
		return "循环流通"
	}
	return "普通流通"
}

// isCircularFlow 检测是否循环流通（木->火->土->金->水->木）
func isCircularFlow(scores map[string]int) bool {
	order := []string{"木", "火", "土", "金", "水"}
	count := 0
	for i := 0; i < len(order); i++ {
		curr := order[i]
		next := order[(i+1)%len(order)]
		if scores[curr] >= 8 && scores[next] >= 8 {
			count++
		}
	}
	return count >= 4
}

func generateFlowAdvice(scores map[string]int, dayElem string, isSmooth bool) string {
	if isSmooth {
		return "五行流通顺畅，阴阳平衡，大吉之象"
	}

	// 找出过旺和过弱的五行
	var strongElems, weakElems []string
	for elem, score := range scores {
		if score > 30 {
			strongElems = append(strongElems, elem)
		} else if score < 10 {
			weakElems = append(weakElems, elem)
		}
	}

	advice := "五行"
	if len(strongElems) > 0 {
		advice += fmt.Sprintf("过旺(%s)，", strongElems[0])
	}
	if len(weakElems) > 0 {
		advice += fmt.Sprintf("过弱(%s)", weakElems[0])
	}
	advice += "，需调和"

	return advice
}

// FindTongGuan 查找通关用神。
// monthZhi 为月支，用于判断节气深浅对通关效果的影响。
func FindTongGuan(pillars []model.Pillar, dayElem string, monthZhi string) TongGuanAnalysis {
	// 通关规则：金木交关需水通关，火金交战需水，木土相战需火
	tongGuanMap := map[string]map[string]string{
		"金": {"木": "水"},
		"木": {"金": "水", "土": "火"},
		"火": {"金": "水"},
		"土": {"木": "火", "水": "土"},
		"水": {"火": "土", "土": "火"},
	}

	// 检查各柱是否有相克需要通关
	elemCount := map[string]int{"木": 0, "火": 0, "土": 0, "金": 0, "水": 0}
	for _, p := range pillars {
		elemCount[GanElement[p.Gan]]++
		elemCount[ZhiElement[p.Zhi]]++
	}

	// 检测金木交关
	if elemCount["金"] > 0 && elemCount["木"] > 0 {
		if mapHasKey(tongGuanMap, "金", "木") {
			weight := calcTongGuanWeight("水", monthZhi)
			desc := "金木交关，需水通关，流通得以调和"
			if weight < 0.8 {
				desc += fmt.Sprintf("（冬月金寒水冷，通关效果减弱，权重%.1f）", weight)
			}
			return TongGuanAnalysis{
				HasTongGuan:     true,
				TongGuanElement: "水",
				Weight:          weight,
				Description:     desc,
			}
		}
	}

	// 检测火金交战
	if elemCount["火"] > 0 && elemCount["金"] > 0 {
		weight := calcTongGuanWeight("水", monthZhi)
		desc := "火金交战，需水通关，流通得以调和"
		if weight < 0.8 {
			desc += fmt.Sprintf("（冬月金寒水冷，通关效果减弱，权重%.1f）", weight)
		}
		return TongGuanAnalysis{
			HasTongGuan:     true,
			TongGuanElement: "水",
			Weight:          weight,
			Description:     desc,
		}
	}

	// 检测木土相战
	if elemCount["木"] > 0 && elemCount["土"] > 0 {
		weight := calcTongGuanWeight("火", monthZhi)
		desc := "木土相战，需火通关，流通得以调和"
		if weight < 0.8 {
			desc += fmt.Sprintf("（冬月火弱，通关效果减弱，权重%.1f）", weight)
		}
		return TongGuanAnalysis{
			HasTongGuan:     true,
			TongGuanElement: "火",
			Weight:          weight,
			Description:     desc,
		}
	}

	return TongGuanAnalysis{HasTongGuan: false, Weight: 1.0}
}

// calcTongGuanWeight 根据月支判断节气深浅对通关权重的影响。
// 经典依据：《滴天髓》"天道有寒暖，地道有燥湿"
// 冬月（亥子丑）水旺金寒，火弱则通关效果减弱；
// 夏月（巳午未）火旺水弱，水通关效果减弱。
func calcTongGuanWeight(tongGuanElem string, monthZhi string) float64 {
	// 月支所属季节
	seasonMap := map[string]string{
		"寅": "春", "卯": "春", "辰": "春",
		"巳": "夏", "午": "夏", "未": "夏",
		"申": "秋", "酉": "秋", "戌": "秋",
		"亥": "冬", "子": "冬", "丑": "冬",
	}
	season := seasonMap[monthZhi]
	if season == "" {
		return 1.0
	}

	// 通关五行在各季节的权重
	// 冬月水旺但火弱，火通关效果减弱；夏月火旺但水弱，水通关效果减弱
	weightMap := map[string]map[string]float64{
		"水": {"春": 0.9, "夏": 0.7, "秋": 1.0, "冬": 1.0}, // 夏月火旺水弱，水通关减弱
		"火": {"春": 1.0, "夏": 1.0, "秋": 0.9, "冬": 0.7}, // 冬月水旺火弱，火通关减弱
		"木": {"春": 1.0, "夏": 0.9, "秋": 0.7, "冬": 0.9}, // 秋月金旺克木，木通关减弱
		"金": {"春": 0.7, "夏": 0.9, "秋": 1.0, "冬": 0.9}, // 春月木旺克金，金通关减弱
		"土": {"春": 0.9, "夏": 1.0, "秋": 0.9, "冬": 0.7}, // 冬月水旺土弱，土通关减弱
	}

	if w, ok := weightMap[tongGuanElem][season]; ok {
		return w
	}
	return 1.0
}

func mapHasKey(m map[string]map[string]string, key1, key2 string) bool {
	if m[key1] == nil {
		return false
	}
	_, ok := m[key1][key2]
	return ok
}

// FindMissingElements 查找缺失五行。
func FindMissingElements(scores map[string]int) MissingElementAnalysis {
	var missing []string
	normalMin := 5 // 低于此值视为缺失

	for elem, score := range scores {
		if score < normalMin {
			missing = append(missing, elem)
		}
	}

	// 补救五行：生缺失五行的
	remedyMap := map[string]string{
		"木": "水",
		"火": "木",
		"土": "火",
		"金": "土",
		"水": "金",
	}

	var remedy []string
	seen := make(map[string]bool)
	for _, m := range missing {
		r := remedyMap[m]
		if r != "" && !seen[r] {
			remedy = append(remedy, r)
			seen[r] = true
		}
	}

	// 严重程度
	severity := "轻微"
	missingCount := len(missing)
	if missingCount >= 3 {
		severity = "严重"
	} else if missingCount >= 2 {
		severity = "中等"
	}

	// 生成补救建议
	advice := generateRemedyAdvice(missing)

	return MissingElementAnalysis{
		MissingElements: missing,
		RemedyElements:  remedy,
		Severity:        severity,
		RemedyAdvice:    advice,
	}
}

// generateRemedyAdvice 根据缺失五行生成生活化补救建议。
// 经典依据：《协纪辨方书》五行方位、颜色、行业对应。
func generateRemedyAdvice(missing []string) []RemedyAdviceItem {
	// 五行对应的生活建议
	adviceMap := map[string]RemedyAdviceItem{
		"木": {
			Element:   "木",
			Direction: "东方",
			Color:     "绿色、青色",
			Industry:  "教育、文化、出版、园艺、中医",
			Daily:     "多接触绿植，佩戴木质饰品，穿绿色衣物",
		},
		"火": {
			Element:   "火",
			Direction: "南方",
			Color:     "红色、紫色、橙色",
			Industry:  "传媒、娱乐、餐饮、能源、电子",
			Daily:     "多晒太阳，佩戴红色饰品，穿暖色衣物",
		},
		"土": {
			Element:   "土",
			Direction: "中央",
			Color:     "黄色、棕色、米色",
			Industry:  "房地产、建筑、农业、珠宝、陶瓷",
			Daily:     "多接触大地，佩戴玉石饰品，穿黄色衣物",
		},
		"金": {
			Element:   "金",
			Direction: "西方",
			Color:     "白色、银色、金色",
			Industry:  "金融、法律、机械、IT硬件、汽车",
			Daily:     "佩戴金属饰品，穿白色衣物，多接触金属器物",
		},
		"水": {
			Element:   "水",
			Direction: "北方",
			Color:     "黑色、蓝色、灰色",
			Industry:  "物流、旅游、渔业、水利、航运",
			Daily:     "多接触水，佩戴黑曜石等深色饰品，穿深色衣物",
		},
	}

	var result []RemedyAdviceItem
	for _, elem := range missing {
		if advice, ok := adviceMap[elem]; ok {
			result = append(result, advice)
		}
	}
	return result
}

// CalcDaYunFlow 计算大运流通影响。
func CalcDaYunFlow(dayGan string, scores map[string]int, dayunPillars []model.Pillar, startAge int) []DaYunFlowItem {
	var result []DaYunFlowItem
	dayElem := GanElement[dayGan]

	for i, p := range dayunPillars {
		ganElem := GanElement[p.Gan]
		zhiElem := ZhiElement[p.Zhi]

		change := "不变"
		impact := "无明显影响"

		// 判断该大运的五行对日主的影响（天干优先，地支辅助）
		elem := ganElem
		if elem == "" {
			elem = zhiElem
		}

		switch elem {
		case ShengWo(dayElem):
			change = "增强"
			impact = fmt.Sprintf("%s大运，%s气（印星）增强，对日主有生扶之力", p.Gan+p.Zhi, elem)
		case dayElem:
			change = "增强"
			impact = fmt.Sprintf("%s大运，%s气（比劫）增强，日主得助但防竞争", p.Gan+p.Zhi, elem)
		case WoSheng(dayElem):
			change = "泄秀"
			impact = fmt.Sprintf("%s大运，%s气（食伤）泄秀，才华得以发挥", p.Gan+p.Zhi, elem)
		case WoKe(dayElem):
			change = "增强"
			impact = fmt.Sprintf("%s大运，%s财星旺盛，利于求财", p.Gan+p.Zhi, elem)
		case KeWo(dayElem):
			change = "减弱"
			impact = fmt.Sprintf("%s大运，%s气（官杀）当令，日主压力增大", p.Gan+p.Zhi, elem)
		}

		result = append(result, DaYunFlowItem{
			StartAge:   startAge + i*10,
			Pillar:     p.Gan + p.Zhi,
			FlowChange: change,
			Impact:     impact,
		})
	}

	return result
}

// BuildFlowPatternDesc 生成流通格局描述。
func BuildFlowPatternDesc(flow WuXingFlowAnalysis, tg TongGuanAnalysis, missing MissingElementAnalysis) string {
	desc := ""

	// 流通类型
	if flow.FlowType != "" && flow.FlowType != "普通流通" {
		desc += fmt.Sprintf("【%s】", flow.FlowType)
	}

	// 通关情况
	if tg.HasTongGuan {
		desc += fmt.Sprintf("有%s通关，", tg.TongGuanElement)
	}

	// 平衡情况
	desc += fmt.Sprintf("%s。", flow.BalanceVerdict)

	// 缺失五行
	if len(missing.MissingElements) > 0 {
		desc += fmt.Sprintf("缺失%s，需注意补足。", joinElems(missing.MissingElements))
	}

	return desc
}

func joinElems(elems []string) string {
	result := ""
	for i, e := range elems {
		if i > 0 {
			result += "、"
		}
		result += e
	}
	return result
}

// ShengWo 返回生我者（印星五行）。导出版本，供其他包调用。
func ShengWo(elem string) string {
	return map[string]string{"木": "水", "火": "木", "土": "火", "金": "土", "水": "金"}[elem]
}

// WoSheng 返回我生者（食伤五行）。导出版本，供其他包调用。
func WoSheng(elem string) string {
	return map[string]string{"木": "火", "火": "土", "土": "金", "金": "水", "水": "木"}[elem]
}

// WoKe 返回我克者（财五行）。导出版本，供其他包调用。
func WoKe(elem string) string {
	return map[string]string{"木": "土", "火": "金", "土": "水", "金": "木", "水": "火"}[elem]
}

// KeWo 返回克我者（官杀五行）。导出版本，供其他包调用。
func KeWo(elem string) string {
	return map[string]string{"木": "金", "火": "水", "土": "木", "金": "火", "水": "土"}[elem]
}