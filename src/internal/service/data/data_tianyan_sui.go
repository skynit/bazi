package data

// TianYanSuiEntry 滴天髓核心教义条目。
// 数据来源：《滴天髓阐微》原文解读与现代应用。
type TianYanSuiEntry struct {
	Category      string   // 分类：强弱/格局/流通/调候/神煞
	Principle     string   // 核心原理
	Judgment      string   // 判断方法
	Description   string   // 详细描述
	FavorElements []string // 喜用五行
	TabooElements []string // 忌讳五行
	Fortune       string   // 命运特征
}

// TianYanSuiKnowledge 滴天髓核心知识库。
var TianYanSuiKnowledge = map[string]TianYanSuiEntry{
	"配合如清": {
		Category:      "格局",
		Principle:     "天干地支配合得当，如清水般清澈",
		Judgment:      "喜忌分明，格局纯粹，无混杂混乱",
		Description:   "配合如清是指四柱天干地支之间配合得当，五行不混杂，清纯而不浊。",
		FavorElements: []string{},
		TabooElements: []string{"混杂"},
		Fortune:       "配合清纯者贵，混杂者贱",
	},
	"甲木参天": {
		Category:      "调候",
		Principle:     "甲木参天，脱胎要火",
		Judgment:      "大树参天需要火来脱胎换骨，戊土来培根",
		Description:   "甲木如大树，要成材需要火来雕琢（脱胎），土来培养。",
		FavorElements: []string{"火", "土"},
		TabooElements: []string{"金"},
		Fortune:       "甲木参天者贵，需火土配合",
	},
	"火明土秀": {
		Category:      "流通",
		Principle:     "火光明亮，土气秀美",
		Judgment:      "火土相生，光明磊落",
		Description:   "火明土秀，火光明亮照耀，土气秀美养物。",
		FavorElements: []string{"火", "土"},
		TabooElements: []string{"水"},
		Fortune:       "火明土秀者，光明磊落，文章秀美",
	},
	"金水澄清": {
		Category:      "格局",
		Principle:     "金水清澄，不混浊",
		Judgment:      "金白水清，聪明秀气",
		Description:   "金水清澄，不混杂。金白水清，聪明秀气之象。",
		FavorElements: []string{"金", "水"},
		TabooElements: []string{"土", "火"},
		Fortune:       "金水澄清者，聪明秀气，文才出众",
	},
	"水空则流": {
		Category:      "强弱",
		Principle:     "水空（地支无根）则流散",
		Judgment:      "水空不能聚物，流散不聚",
		Description:   "水空指地支无根（截路空亡等），水气流散不能聚物。",
		FavorElements: []string{"金"},
		TabooElements: []string{"空亡"},
		Fortune:       "水空则流，财难聚，奔波劳碌",
	},
	"木火通明": {
		Category:      "流通",
		Principle:     "木火相生，文明之象",
		Judgment:      "木火通明，才智显露，文章秀美",
		Description:   "木火相生，文明之象。才智显露，文章秀美。",
		FavorElements: []string{"木", "火"},
		TabooElements: []string{"水", "金"},
		Fortune:       "木火通明者，才华出众，文采飞扬",
	},
	"金木交关": {
		Category:      "流通",
		Principle:     "金木相克相成",
		Judgment:      "金木交关，关键在于通关",
		Description:   "金木相克，需要有通关之神（水）来调和。",
		FavorElements: []string{"水"},
		TabooElements: []string{"无通关"},
		Fortune:       "金木交关，有通关者贵，无通关者贱",
	},
	"土金毓秀": {
		Category:      "流通",
		Principle:     "土金相生，毓秀之美",
		Judgment:      "土金相生，厚德载物，秀外慧中",
		Description:   "土金相生，毓秀之美。厚德载物，秀外慧中。",
		FavorElements: []string{"土", "金"},
		TabooElements: []string{"木", "火"},
		Fortune:       "土金毓秀者，厚德载物，贵气内敛",
	},
	"水火既济": {
		Category:      "流通",
		Principle:     "水火相济，坎离相交",
		Judgment:      "水火既济，阴阳平衡",
		Description:   "水火相济，坎离相交。既济卦象，阴阳平衡之象。",
		FavorElements: []string{"水", "火"},
		TabooElements: []string{"过偏"},
		Fortune:       "水火既济者，阴阳平衡，功成名就",
	},
	"燥湿调和": {
		Category:      "调候",
		Principle:     "燥土需水润，湿寒需火温",
		Judgment:      "燥湿调和为贵",
		Description:   "燥土（夏季土）需水来润泽，湿寒（冬季）需火来温暖。",
		FavorElements: []string{"水", "火"},
		TabooElements: []string{"过燥", "过湿"},
		Fortune:       "燥湿调和者，运气通达",
	},
	"水木相生": {
		Category:      "流通",
		Principle:     "水木相生，生命力强",
		Judgment:      "水生木，生命力旺盛，创业立业",
		Description:   "水为木之母，木为水之子，相生有情。",
		FavorElements: []string{"水", "木"},
		TabooElements: []string{"金", "火过旺"},
		Fortune:       "水木相生者，生命力强，宜创业发展",
	},
	"木土相制": {
		Category:      "流通",
		Principle:     "木克土，制化得当",
		Judgment:      "木克土需火通关，无火则凶",
		Description:   "木克土为制服，忌土重无火通关。",
		FavorElements: []string{"火", "木"},
		TabooElements: []string{"土过旺", "无火"},
		Fortune:       "木土相制，有通关者贵",
	},
	"火金相克": {
		Category:      "流通",
		Principle:     "火克金，需要水通关",
		Judgment:      "火金相克，水通关则调和",
		Description:   "火金交战，需水来通关调和。",
		FavorElements: []string{"水"},
		TabooElements: []string{"无通关"},
		Fortune:       "火金相克，有通关者贵",
	},
	"土水相激": {
		Category:      "流通",
		Principle:     "土克水，需木通关",
		Judgment:      "土水相激，木通关则吉",
		Description:   "土克水为相战，需木来通关。",
		FavorElements: []string{"木"},
		TabooElements: []string{"无木"},
		Fortune:       "土水相激，有通关者吉",
	},
	"相生流通": {
		Category:      "流通",
		Principle:     "木->火->土->金->水->木 循环相生",
		Judgment:      "流通无阻，大吉之象",
		Description:   "五行相生形成循环流通，源远流长。",
		FavorElements: []string{"木", "火", "土", "金", "水"},
		TabooElements: []string{"阻滞"},
		Fortune:       "流通顺畅者，运势恒通",
	},
	"循环流通": {
		Category:      "流通",
		Principle:     "五行循环相生，周而复始",
		Judgment:      "循环流通，如水之流",
		Description:   "五行形成完整循环，生生不息。",
		FavorElements: []string{"木", "火", "土", "金", "水"},
		TabooElements: []string{"某行过旺", "某行过弱"},
		Fortune:       "循环流通者，命运曲折但坚韧",
	},
	"寒暖燥湿": {
		Category:      "调候",
		Principle:     "冬寒需火暖，夏燥需水润",
		Judgment:      "寒暖燥湿调和为贵",
		Description:   "冬季水寒无火则寒，夏季火燥无水则枯。",
		FavorElements: []string{"火", "水"},
		TabooElements: []string{"过寒", "过燥"},
		Fortune:       "寒暖调和者，运势平稳",
	},
	"金生水旺": {
		Category:      "流通",
		Principle:     "金水相生，智慧流通",
		Judgment:      "金水流通，聪明睿智",
		Description:   "金能生水，水为智慧，金水相生则文采风流。",
		FavorElements: []string{"金", "水"},
		TabooElements: []string{"土重", "火旺"},
		Fortune:       "金生水旺者，聪明多智",
	},
	"火土同旺": {
		Category:      "流通",
		Principle:     "火土相生，光明磊落",
		Judgment:      "火土同旺，诚实忠厚",
		Description:   "火生土，土得火照，光明之象。",
		FavorElements: []string{"火", "土"},
		TabooElements: []string{"水", "木"},
		Fortune:       "火土同旺者，光明磊落，诚信可靠",
	},
}

// GetTianYanSuiEntry 返回滴天髓知识条目。
func GetTianYanSuiEntry(key string) (TianYanSuiEntry, bool) {
	e, ok := TianYanSuiKnowledge[key]
	return e, ok
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

// MissingElementAnalysis records the raw five-element score distribution.
// It deliberately does not adjudicate favorable elements or remedies.
type MissingElementAnalysis struct {
	Status               string         `json:"status"`
	RuleID               string         `json:"rule_id"`
	MissingElements      []string       `json:"missing_elements"` // score == 0
	WeakElements         []string       `json:"weak_elements"`    // 0 < score < 5
	Scores               map[string]int `json:"scores"`
	MissingCount         int            `json:"missing_count"`
	IsYongshenConclusion bool           `json:"is_yongshen_conclusion"`
	RemedyStatus         string         `json:"remedy_status"`
	Note                 string         `json:"note"`
}

// FindMissingElements observes missing and low-scoring elements.
// scores 是 calcFiveElements 的输出：天干+5、地支藏干本气+3/中气+2/余气+1。
// 判断规则：
//   - score == 0 → 真正缺失（柱中全无）
//   - 0 < score < 5 → 偏弱（仅藏于地支中气/余气，天干不透）
//   - score >= 5 → 正常
//
// 该函数只输出原始计分事实，不推导喜用神、补救五行或生活建议。
func FindMissingElements(scores map[string]int) MissingElementAnalysis {
	elements := []string{"木", "火", "土", "金", "水"}
	missing := make([]string, 0, len(elements))
	weak := make([]string, 0, len(elements))
	normalizedScores := make(map[string]int, len(elements))

	for _, elem := range elements {
		score := scores[elem]
		normalizedScores[elem] = score
		if score == 0 {
			missing = append(missing, elem)
		} else if score < 5 {
			weak = append(weak, elem)
		}
	}

	return MissingElementAnalysis{
		Status:               "observed",
		RuleID:               "wuxing.raw-score-presence-v1",
		MissingElements:      missing,
		WeakElements:         weak,
		Scores:               normalizedScores,
		MissingCount:         len(missing),
		IsYongshenConclusion: false,
		RemedyStatus:         "not_adjudicated",
		Note:                 "原始五行计分中的缺失或偏低不等于喜用神，也不自动代表需要补入相应五行。",
	}
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
