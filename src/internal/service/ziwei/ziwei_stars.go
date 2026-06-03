// Package service provides ZiWei Dou Shu calculation and interpretation services.
//
// This file holds all data constants for the ZiWei system — no logic functions.
// Data is derived from iztro library source code and classical Chinese ZiWei texts.
//
// Constants include:
//   - STAR_BRIGHTNESS: 14 main star brightness levels (庙旺落陷)
//   - STAR_BRIGHTNESS_AUX: 12 auxiliary/unlucky star brightness levels
//   - LUCUN_TABLE: 禄存表 (LuCun star placement)
//   - TIANMA_TABLE: 天马表 (TianMa star placement)
//   - PALACE_NAMES: 12 palace names in order
//   - ADJECTIVE_STAR_PLACEMENTS: placement rules for reactive/secondary stars
//   - 12神 offset tables: 长生12, 博士12, 将前12, 岁建12
package ziwei

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
	"左辅": {0, 0, 0, 0, 3, 4, 6},
	"右弼": {0, 0, 0, 0, 3, 4, 6},
	"天魁": {0, 0, 0, 0, 3, 5, 6},
	"天钺": {0, 0, 0, 0, 3, 5, 6},
	"文昌": {0, 0, 0, 0, 3, 4, 6},
	"文曲": {0, 0, 0, 0, 3, 4, 6},
	"禄存": {0, 0, 0, 0, 3, 4, 6},
	"天马": {0, 0, 0, 0, 3, 4, 6},
	"陀罗": {0, 1, 2, 0, 0, 0, 0},
	"擎羊": {0, 1, 2, 0, 0, 0, 0},
	"火星": {0, 1, 2, 0, 0, 0, 0},
	"铃星": {0, 1, 2, 0, 0, 0, 0},
}

// ZIWEI_PALACE_NAMES is now defined in ziwei_data.go

// AdjectiveStarPlacement describes how to place one adjective (secondary/reactive) star.
type AdjectiveStarPlacement struct {
	Name      string
	BasedOn   string                      // "year" | "month" | "day" | "hour"
	StartFunc func(chart *ZiWeiChart) int // returns starting palace index
	Direction int                         // +1 or -1
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
			return c.YearBranch % 12
		},
	},
	"天喜": {
		Name:    "天喜",
		BasedOn: "year",
		StartFunc: func(c *ZiWeiChart) int {
			yearBranch := c.YearBranch % 12
			hongluanTarget := (1 - yearBranch + 12) % 12
			return (hongluanTarget + 6) % 12
		},
		Direction: +1,
		StepFunc: func(c *ZiWeiChart) int {
			return 6
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
			return c.LunarMonth % 12
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
			return c.LunarMonth % 12
		},
	},
}

// XIANCHI_TABLE maps year branch index to 咸池 palace index.
// Classical rule: 申子辰见酉, 寅午戌见卯, 巳酉丑见午, 亥卯未见子
// 0=子,1=丑,2=寅,3=卯,4=辰,5=巳,6=午,7=未,8=申,9=酉,10=戌,11=亥
var XIANCHI_TABLE = map[int]int{
	0:  9, // 子年 → 酉
	1:  6, // 丑年 → 午
	2:  3, // 寅年 → 卯
	3:  0, // 卯年 → 子
	4:  9, // 辰年 → 酉
	5:  6, // 巳年 → 午
	6:  3, // 午年 → 卯
	7:  0, // 未年 → 子
	8:  9, // 申年 → 酉
	9:  6, // 酉年 → 午
	10: 3, // 戌年 → 卯
	11: 0, // 亥年 → 子
}

// HUAGAI_TABLE maps year branch index to 华盖 palace index.
// Classical rule: 申子辰见辰, 寅午戌见戌, 巳酉丑见丑, 亥卯未见未
var HUAGAI_TABLE = map[int]int{
	0:  4,  // 子年 → 辰
	1:  1,  // 丑年 → 丑
	2:  10, // 寅年 → 戌
	3:  7,  // 卯年 → 未
	4:  4,  // 辰年 → 辰
	5:  1,  // 巳年 → 丑
	6:  10, // 午年 → 戌
	7:  7,  // 未年 → 未
	8:  4,  // 申年 → 辰
	9:  1,  // 酉年 → 丑
	10: 10, // 戌年 → 戌
	11: 7,  // 亥年 → 未
}

// POUSUI_TABLE maps year branch index to 破碎 (PoSui) palace index.
// Classical rule: 申子辰在辰, 寅午戌在戌, 巳酉丑在丑, 亥卯未在未
var POUSUI_TABLE = map[int]int{
	0:  4,  // 子年 → 辰
	1:  1,  // 丑年 → 丑
	2:  10, // 寅年 → 戌
	3:  7,  // 卯年 → 未
	4:  4,  // 辰年 → 辰
	5:  1,  // 巳年 → 丑
	6:  10, // 午年 → 戌
	7:  7,  // 未年 → 未
	8:  4,  // 申年 → 辰
	9:  1,  // 酉年 → 丑
	10: 10, // 戌年 → 戌
	11: 7,  // 亥年 → 未
}

// FEILIAN_TABLE maps year branch index to 飞廉 palace index.
// Classical rule: 申子辰在寅, 寅午戌在申, 巳酉丑在亥, 亥卯未在巳
var FEILIAN_TABLE = map[int]int{
	0:  2,  // 子年 → 寅
	1:  11, // 丑年 → 亥
	2:  8,  // 寅年 → 申
	3:  5,  // 卯年 → 巳
	4:  2,  // 辰年 → 寅
	5:  11, // 巳年 → 亥
	6:  8,  // 午年 → 申
	7:  5,  // 未年 → 巳
	8:  2,  // 申年 → 寅
	9:  11, // 酉年 → 亥
	10: 8,  // 戌年 → 申
	11: 5,  // 亥年 → 巳
}

// YINSHA_TABLE maps month index (0-based) to 阴煞 palace index.
// Classical rule: 寅月在卯, 卯月在辰, ..., 丑月在寅 (正月起卯逆数)
var YINSHA_TABLE = map[int]int{
	0:  3,  // 正月(寅) → 卯
	1:  4,  // 二月(卯) → 辰
	2:  5,  // 三月(辰) → 巳
	3:  6,  // 四月(巳) → 午
	4:  7,  // 五月(午) → 未
	5:  8,  // 六月(未) → 申
	6:  9,  // 七月(申) → 酉
	7:  10, // 八月(酉) → 戌
	8:  11, // 九月(戌) → 亥
	9:  0,  // 十月(亥) → 子
	10: 1,  // 十一月(子) → 丑
	11: 2,  // 十二月(丑) → 寅
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
