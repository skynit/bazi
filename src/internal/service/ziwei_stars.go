// Package service provides ZiWei Dou Shu calculation and interpretation services.
//
// This file holds all data constants for the ZiWei system — no logic functions.
// Data is derived from iztro library source code and classical Chinese ZiWei texts.
//
// Constants include:
//   - SI_HUA_TABLE: 四化表 (transformation star lookup by dynasty meter)
//   - STAR_BRIGHTNESS: 14 main star brightness levels (庙旺落陷)
//   - STAR_BRIGHTNESS_AUX: 12 auxiliary/unlucky star brightness levels
//   - LUCUN_TABLE: 禄存表 (LuCun star placement)
//   - TIANMA_TABLE: 天马表 (TianMa star placement)
//   - PALACE_NAMES: 12 palace names in order
//   - ADJECTIVE_STAR_PLACEMENTS: placement rules for reactive/secondary stars
//   - 12神 offset tables: 长生12, 博士12, 将前12, 岁建12
package service

// SI_HUA_DYNASTY_TABLE maps dynasty meter (年数) to transformation star indices.
// Index: 0=紫微, 1=天机, 2=太阳, 3=武曲, 4=天同, 5=廉贞, 6=天府, 7=太阴, 8=贪狼, 9=天梁, 10=天机, 11=破军
// From iztro star/siHua.ts and classical ZiWei texts.
var SI_HUA_DYNASTY_TABLE = map[string][]string{
	"水二局": {"天机", "天梁", "天同", "廉贞", "天府", "武曲", "太阳", "天相", "天机", "天梁", "天同", "破军"},
	"木三局": {"天梁", "天机", "天同", "廉贞", "天府", "武曲", "太阳", "天相", "天梁", "天机", "天同", "破军"},
	"金四局": {"天机", "天同", "天梁", "廉贞", "天府", "武曲", "太阳", "天相", "天机", "天同", "天梁", "破军"},
	"土五局": {"天同", "天机", "天梁", "廉贞", "天府", "武曲", "太阳", "天相", "天同", "天机", "天梁", "破军"},
	"火六局": {"天梁", "天同", "天机", "廉贞", "天府", "武曲", "太阳", "天相", "天梁", "天同", "天机", "破军"},
}

// STAR_BRIGHTNESS_MAIN maps main star name to 7-level brightness array.
// Levels: 0=陷(陷), 1=弱(弱), 2=不得地(不得地), 3=得地(得地), 4=中立(中), 5=旺(旺), 6=庙(庙)
// Based on iztro data/starBrightness.ts and classical texts.
// Note: only 14 main stars use this table.
var STAR_BRIGHTNESS_MAIN = map[string][7]int{
	"紫微": {0, 0, 0, 0, 3, 5, 6},
	"天机": {0, 0, 0, 0, 3, 5, 6},
	"太阳": {0, 0, 0, 0, 3, 5, 6},
	"武曲": {0, 0, 0, 0, 3, 4, 6},
	"天同": {0, 0, 0, 0, 3, 5, 6},
	"廉贞": {0, 0, 0, 0, 3, 4, 6},
	"天府": {0, 0, 0, 0, 3, 4, 6},
	"太阴": {0, 0, 0, 0, 3, 5, 6},
	"贪狼": {0, 0, 0, 0, 3, 4, 6},
	"天梁": {0, 0, 0, 0, 3, 5, 6},
	"天相": {0, 0, 0, 0, 3, 4, 6},
	"七杀": {0, 0, 0, 0, 3, 4, 6},
	"破军": {0, 0, 0, 0, 3, 4, 6},
	"巨门": {0, 0, 0, 0, 3, 4, 6},
}

// STAR_BRIGHTNESS_AUX_STARS maps auxiliary/unlucky star name to 7-level brightness array.
// Based on iztro data/starBrightness.ts.
var STAR_BRIGHTNESS_AUX_STARS = map[string][7]int{
	"左辅":    {0, 0, 0, 0, 3, 4, 6},
	"右弼":    {0, 0, 0, 0, 3, 4, 6},
	"天魁":    {0, 0, 0, 0, 3, 5, 6},
	"天钺":    {0, 0, 0, 0, 3, 5, 6},
	"文昌":    {0, 0, 0, 0, 3, 4, 6},
	"文曲":    {0, 0, 0, 0, 3, 4, 6},
	"禄存":    {0, 0, 0, 0, 3, 4, 6},
	"天马":    {0, 0, 0, 0, 3, 4, 6},
	"陀罗":    {0, 1, 2, 0, 0, 0, 0},
	"擎羊":    {0, 1, 2, 0, 0, 0, 0},
	"火星":    {0, 1, 2, 0, 0, 0, 0},
	"铃星":    {0, 1, 2, 0, 0, 0, 0},
}

// LUCUN_FIVE_ELEMENT_TABLE maps five element class (管局) to LuCun branch index.
// From iztro star/lucun.ts and classical ZiWei derivation.
var LUCUN_FIVE_ELEMENT_TABLE = map[int]int{
	2: 9,  // 水二局 -> 申(9)
	3: 11, // 木三局 -> 亥(11)
	4: 5,  // 金四局 -> 巳(5)
	5: 9,  // 土五局 -> 申(9)
	6: 0,  // 火六局 -> 寅(0)
}

// TIANMA_FIVE_ELEMENT_TABLE maps five element class (管局) to TianMa branch index.
// From iztro star/tianma.ts.
var TIANMA_FIVE_ELEMENT_TABLE = map[int]int{
	2: 3,  // 水二局 -> 卯(3)
	3: 5,  // 木三局 -> 巳(5)
	4: 7,  // 金四局 -> 午(7)
	5: 11, // 土五局 -> 亥(11)
	6: 9,  // 火六局 -> 酉(9)
}

// ZIWEI_PALACE_NAMES are the 12 ZiWei palace names in order (starting from 命宫).
// Index 0=命宫, 1=兄弟, 2=夫妻, 3=子女, 4=财帛, 5=疾厄, 6=迁移, 7=交友, 8=事业, 9=田宅, 10=福不全, 11=父母
var ZIWEI_PALACE_NAMES = []string{
	"命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄",
	"迁移", "交友", "事业", "田宅", "福不全", "父母",
}

// AdjectiveStarPlacement describes how to place one adjective (secondary/reactive) star.
type AdjectiveStarPlacement struct {
	Name      string
	BasedOn   string // "year" | "month" | "day" | "hour"
	StartFunc func(chart *ZiWeiChart) int // returns starting palace index
	Direction int                              // +1 or -1
	StepFunc  func(chart *ZiWeiChart) int // returns step count
}

// ADJECTIVE_STAR_PLACEMENTS maps star name to placement rule.
// Derived from iztro src/star/adjectiveStar.ts and classical ZiWei placement rules.
//
// Placement formulas:
//   - 红鸾: start=1(卯宫), direction=-1, steps=yearBranchIndex → target=(1 - yearBranch + 12) % 12
//   - 天喜: target=(红鸾+6) % 12 (对宫)
//   - 咸池: based on year branch (see XIANCHI_TABLE)
//   - 华盖: based on year branch (see HUAGAI_TABLE)
//   - 天姚: start=1(丑宫), direction=+1, steps=monthIndex
//   - 天刑: start=9(酉宫), direction=+1, steps=monthIndex
var ADJECTIVE_STAR_PLACEMENTS = map[string]AdjectiveStarPlacement{
	"红鸾": {
		Name:    "红鸾",
		BasedOn: "year",
		StartFunc: func(c *ZiWeiChart) int {
			return 1 // 卯宫
		},
		Direction: -1,
		StepFunc: func(c *ZiWeiChart) int {
			return int(c.engineChart.YearPillar.Branch) % 12
		},
	},
	"天喜": {
		Name:    "天喜",
		BasedOn: "year",
		StartFunc: func(c *ZiWeiChart) int {
			// 天喜 is always opposite (对宫, +6) of 红鸾
			yearBranch := int(c.engineChart.YearPillar.Branch) % 12
			// 红鸾: (1 - yearBranch + 12) % 12
			hongluanTarget := (1 - yearBranch + 12) % 12
			return (hongluanTarget + 6) % 12
		},
		Direction: +1,
		StepFunc: func(c *ZiWeiChart) int {
			return 6 // always opposite of 红鸾
		},
	},
	"天姚": {
		Name:    "天姚",
		BasedOn: "month",
		StartFunc: func(c *ZiWeiChart) int {
			return 1 // 丑宫
		},
		Direction: +1,
		StepFunc: func(c *ZiWeiChart) int {
			return int(c.engineChart.MonthPillar.Branch) % 12
		},
	},
	"天刑": {
		Name:    "天刑",
		BasedOn: "month",
		StartFunc: func(c *ZiWeiChart) int {
			return 9 // 酉宫
		},
		Direction: +1,
		StepFunc: func(c *ZiWeiChart) int {
			return int(c.engineChart.MonthPillar.Branch) % 12
		},
	},
}

// XIANCHI_TABLE maps year branch index to 咸池 palace index.
// From iztro decorativeStar.ts getXianchi.
// 0=子,1=丑,2=寅,3=卯,4=辰,5=巳,6=午,7=未,8=申,9=酉,10=戌,11=亥
var XIANCHI_TABLE = map[int]int{
	0:  9,  // 子 -> 酉(9)
	1:  3,  // 丑 -> 卯(3)
	2:  6,  // 寅 -> 午(6)
	3:  0,  // 卯 -> 子(0)
	4:  6,  // 辰 -> 午(6)
	5:  9,  // 巳 -> 酉(9)
	6:  0,  // 午 -> 子(0)
	7:  3,  // 未 -> 卯(3)
	8:  6,  // 申 -> 午(6)
	9:  9,  // 酉 -> 酉(9)
	10: 0,  // 戌 -> 子(0)
	11: 3,  // 亥 -> 卯(3)
}

// HUAGAI_TABLE maps year branch index to 华盖 palace index.
// From iztro decorativeStar.ts getHuagai.
var HUAGAI_TABLE = map[int]int{
	0:  5,  // 子 -> 巳(5)
	1:  11, // 丑 -> 亥(11)
	2:  8,  // 寅 -> 申(8)
	3:  2,  // 卯 -> 丑(2)
	4:  8,  // 辰 -> 申(8)
	5:  2,  // 巳 -> 丑(2)
	6:  11, // 午 -> 亥(11)
	7:  5,  // 未 -> 巳(5)
	8:  2,  // 申 -> 丑(2)
	9:  8,  // 酉 -> 申(8)
	10: 11, // 戌 -> 亥(11)
	11: 5,  // 亥 -> 巳(5)
}

// POUSUI_TABLE maps year branch index to 破碎 (PoSui) palace index.
// 3-cycle: 巳(5)/丑(1)/酉(9) based on yearBranchIndex % 3
// From iztro decorativeStar.ts getPosui.
var POUSUI_TABLE = map[int]int{
	0: 5,  // 子年 -> 巳(5)  (0%3=0 -> 巳)
	1: 1,  // 丑年 -> 丑(1)  (1%3=1 -> 丑)
	2: 9,  // 寅年 -> 酉(9)  (2%3=2 -> 酉)
	3: 5,  // 卯年 -> 巳(5)  (0%3=0 -> 巳)
	4: 1,  // 辰年 -> 丑(1)  (1%3=1 -> 丑)
	5: 9,  // 巳年 -> 酉(9)  (2%3=2 -> 酉)
	6: 5,  // 午年 -> 巳(5)
	7: 1,  // 未年 -> 丑(1)
	8: 9,  // 申年 -> 酉(9)
	9: 5,  // 酉年 -> 巳(5)
	10: 1, // 戌年 -> 丑(1)
	11: 9, // 亥年 -> 酉(9)
}

// FEILIAN_TABLE maps year branch index to 飞廉 palace index.
// From iztro decorativeStar.ts getFeilian.
var FEILIAN_TABLE = map[int]int{
	0:  10, // 子 -> 戌(10)
	1:  9,  // 丑 -> 酉(9)
	2:  11, // 寅 -> 亥(11)
	3:  9,  // 卯 -> 酉(9)
	4:  11, // 辰 -> 亥(11)
	5:  10, // 巳 -> 戌(10)
	6:  10, // 午 -> 戌(10)
	7:  9,  // 未 -> 酉(9)
	8:  11, // 申 -> 亥(11)
	9:  9,  // 酉 -> 酉(9)
	10: 11, // 戌 -> 亥(11)
	11: 10, // 亥 -> 戌(10)
}

// YINSHIA_TABLE maps month index (0-based) to 阴煞 palace index.
// From iztro decorativeStar.ts getYinsha.
var YINSHIA_TABLE = map[int]int{
	0:  3,  // 正月(寅) -> 卯(3)
	1:  11, // 二月(卯) -> 亥(11)
	2:  8,  // 三月(辰) -> 申(8)
	3:  5,  // 四月(巳) -> 午(5)
	4:  2,  // 五月(午) -> 丑(2)
	5:  0,  // 六月(未) -> 子(0)
	6:  3,  // 七月(申) -> 卯(3)
	7:  11, // 八月(酉) -> 亥(11)
	8:  8,  // 九月(戌) -> 申(8)
	9:  5,  // 十月(亥) -> 午(5)
	10: 2,  // 十一月(子) -> 丑(2)
	11: 0,  // 十二月(丑) -> 子(0)
}

// CHANGSHENG_12 names in order (index 0=长生, 11=养).
// From iztro decorativeStar.ts getChangsheng12.
var CHANGSHENG_12 = []string{
	"长生", "沐浴", "冠带", "临官", "帝旺", "衰",
	"病", "死", "墓", "绝", "胎", "养",
}

// CHANGSHENG_START_TABLE maps five element class to starting branch index for 长生12.
// From iztro decorativeStar.ts getchangsheng12.
// 水二局→9(申), 木三局→11(亥), 金四局→5(巳), 土五局→9(申), 火六局→0(寅)
var CHANGSHENG_START_TABLE = map[int]int{
	2: 9,  // 水二局
	3: 11, // 木三局
	4: 5,  // 金四局
	5: 9,  // 土五局
	6: 0,  // 火六局
}

// BOSHI_12 names in order (index 0=博士, 11=官府).
// From iztro decorativeStar.ts getBoshi12.
var BOSHI_12 = []string{
	"博士", "力士", "青龙", "小耗", "将军", "奏书",
	"飞廉", "喜神", "病符", "大耗", "伏兵", "官府",
}

// JIANG_QIAN_12 names in order (index 0=将星, 11=亡神).
// From iztro decorativeStar.ts getJiangqian12.
var JIANG_QIAN_12 = []string{
	"将星", "攀鞍", "岁驿", "息神", "华盖", "劫煞",
	"灾煞", "天煞", "指背", "咸池", "月煞", "亡神",
}

// JIANG_QIAN_START_TABLE maps year branch group to starting branch index for 将前12.
// From iztro decorativeStar.ts getJiangqian12.
// 寅午戌→5(午), 申子辰→0(子), 巳酉丑→9(酉), 亥卯未→1(卯)
var JIANG_QIAN_START_TABLE = map[string]int{
	"寅午戌": 5,
	"申子辰": 0,
	"巳酉丑": 9,
	"亥卯未": 1,
}

// SUI_QIAN_12 names in order (index 0=岁建, 11=病符).
// Default order used when scope-specific override not needed.
// From iztro decorativeStar.ts getSuiqian12.
var SUI_QIAN_12 = []string{
	"岁建", "晦气", "丧门", "贯索", "官符", "小耗",
	"大耗", "龙德", "白虎", "天德", "吊客", "病符",
}

// LIU_YAO_STARS is the list of 10 流耀 stars used for 流年/流月/流日.
// Same star list applies to all three time scopes.
// From iztro star/liuyao.ts.
var LIU_YAO_STARS = []string{
	"天魁", "天钺", "文昌", "文曲", "禄存",
	"擎羊", "陀罗", "天马", "红鸾", "天喜",
}

// TIANQIN_TABLE maps year branch index to 天琴 palace index.
// From iztro decorativeStar.ts getTianqin (if applicable).
// var TIANQIN_TABLE = map[int]int{}

// JINGDIAN_TABLE maps year branch index to 经典 (classical) index.
// Not needed for adjective stars.
