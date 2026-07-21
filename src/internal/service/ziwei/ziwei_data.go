package ziwei

// ziwei_data.go — All data constants for the ZiWei Dou Shu system.
// Derived from iztro (https://github.com/SylarLong/iztro) and classical Chinese ZiWei texts.
// No logic functions in this file — only data tables and constants.

// ──────────── Heavenly Stems (天干) ────────────

// StemIndex maps 天干 name to its index (甲=0...癸=9).
var StemIndex = map[string]int{
	"甲": 0, "乙": 1, "丙": 2, "丁": 3, "戊": 4,
	"己": 5, "庚": 6, "辛": 7, "壬": 8, "癸": 9,
}

// StemNames lists the 10 heavenly stems in order.
var StemNames = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}

// StemYinYang: odd=阳, even=阴
func StemIsYang(idx int) bool { return idx%2 == 0 }

// ──────────── Earthly Branches (地支) ────────────

// BranchIndex maps 地支 name to its index (子=0...亥=11).
var BranchIndex = map[string]int{
	"子": 0, "丑": 1, "寅": 2, "卯": 3, "辰": 4, "巳": 5,
	"午": 6, "未": 7, "申": 8, "酉": 9, "戌": 10, "亥": 11,
}

// BranchNames lists the 12 earthly branches in order.
var BranchNames = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

// BranchYinYang: odd-index=阴, even-index=阳 (子=0阳,丑=1阴...)
func BranchIsYang(idx int) bool { return idx%2 == 0 }

// BranchFiveElement maps branch to its inherent five element.
var BranchFiveElement = map[string]string{
	"子": "水", "丑": "土", "寅": "木", "卯": "木", "辰": "土", "巳": "火",
	"午": "火", "未": "土", "申": "金", "酉": "金", "戌": "土", "亥": "水",
}

// ──────────── Palace Names ────────────

// ZIWEI_PALACE_NAMES lists the 12 palaces in order (starting from 命宫).
var ZIWEI_PALACE_NAMES = []string{
	"命宫", "兄弟", "夫妻", "子女", "财帛", "疾厄",
	"迁移", "交友", "事业", "田宅", "福德", "父母",
}

// ──────────── Five Tiger Rules (五虎遁) ────────────

// TigerRule maps year stem → the stem of the 寅(yin) branch.
// 甲己起丙寅, 乙庚起戊寅, 丙辛起庚寅, 丁壬起壬寅, 戊癸起甲寅
var TigerRule = map[int]int{
	0: 2, // 甲 → 丙(2)
	1: 4, // 乙 → 戊(4)
	2: 6, // 丙 → 庚(6)
	3: 8, // 丁 → 壬(8)
	4: 0, // 戊 → 甲(0)
	5: 2, // 己 → 丙(2)
	6: 4, // 庚 → 戊(4)
	7: 6, // 辛 → 庚(6)
	8: 8, // 壬 → 壬(8)
	9: 0, // 癸 → 甲(0)
}

// GetPalaceStem returns the heavenly stem for a palace at branch index,
// given the year stem index and the branch index of 命宫 (soulBranch).
// The palace's stem is derived from 五虎遁.
func GetPalaceStem(yearStem, branchIdx int) int {
	yinStem := TigerRule[yearStem]
	return (yinStem + (branchIdx-2+12)%12) % 10
}

// ──────────── Five Element Bureau (五行局) ────────────

// FiveBureauName maps bureau value to display name.
var FiveBureauName = map[int]string{
	2: "水二局", 3: "木三局", 4: "金四局", 5: "土五局", 6: "火六局",
}

// FiveBureauValue maps bureau display name to integer value.
var FiveBureauValue = map[string]int{
	"水二局": 2, "木三局": 3, "金四局": 4, "土五局": 5, "火六局": 6,
}

// NaYinBureauTable maps 干支 pair index (0-29) → five element bureau value.
// Each 干支 pair in the 60-cycle is grouped into 30 pairs (甲子乙丑, 丙寅丁卯, ...).
// The pair index is computed from (命宫天干, 命宫地支) via ganzhiPairIndex().
// Bureau values: 金=4, 火=6, 木=3, 土=5, 水=2
var NaYinBureauTable = [30]int{
	4, 6, 3, 5, 4, // 甲子乙丑→金, 丙寅丁卯→火, 戊辰己巳→木, 庚午辛未→土, 壬申癸酉→金
	6, 2, 5, 4, 3, // 甲戌乙亥→火, 丙子丁丑→水, 戊寅己卯→土, 庚辰辛巳→金, 壬午癸未→木
	2, 5, 6, 3, 2, // 甲申乙酉→水, 丙戌丁亥→土, 戊子己丑→火, 庚寅辛卯→木, 壬辰癸巳→水
	4, 6, 3, 5, 4, // 甲午乙未→金, 丙申丁酉→火, 戊戌己亥→木, 庚子辛丑→土, 壬寅癸卯→金
	6, 2, 5, 4, 3, // 甲辰乙巳→火, 丙午丁未→水, 戊申己酉→土, 庚戌辛亥→金, 壬子癸丑→木
	2, 5, 6, 3, 2, // 甲寅乙卯→水, 丙辰丁巳→土, 戊午己未→火, 庚申辛酉→木, 壬戌癸亥→水
}

// ganzhiPairIndex computes the 干支 pair index (0-29) from stem and branch indices.
// Uses Chinese Remainder Theorem: 干支_index = (6*stem - 5*branch + 600) % 60
// Returns pairIndex = 干支_index / 2 (floor division).
func ganzhiPairIndex(stem, branch int) int {
	ganzhi := (6*stem - 5*branch + 600) % 60
	return ganzhi / 2
}

func BranchValue(branchIdx int) int {
	return (branchIdx%6+6)%6 + 1 // deprecated, kept for compatibility
}

// NaYinBureauMap maps sum value → five bureau value.

// ZiweiStarOrder defines the 紫微星系 stars in order from 紫微 position.
// Index 0 = 紫微星本位, each subsequent index is one branch away counterclockwise.
// Empty string "" means no star in that position (skip positions).
// From iztro majorStar.ts: ["紫微", "天机", "", "太阳", "武曲", "天同", "", "", "廉贞"]
var ZiweiStarOrder = []string{"紫微", "天机", "", "太阳", "武曲", "天同", "", "", "廉贞"}

// TianfuStarOrder defines the 天府星系 stars in order from 天府 position.
// Each subsequent index is one branch away clockwise.
// From iztro majorStar.ts: ["天府", "太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "", "", "", "破军"]
var TianfuStarOrder = []string{"天府", "太阴", "贪狼", "巨门", "天相", "天梁", "七杀", "", "", "", "破军"}

// ──────────── Star Brightness Tables ────────────

// StarBrightnessMap maps (star_name, branch_index) -> brightness level string.
// Brightness levels: 庙, 旺, 得, 利, 平, 不, 陷.
// Source: SylarLong/iztro src/data/stars.ts STARS_INFO.
// Iztro stores brightness from 寅, so rows here are converted to BranchNames
// order: 子、丑、寅、卯、辰、巳、午、未、申、酉、戌、亥.
var StarBrightnessMap = map[string][12]string{
	"紫微": {"平", "庙", "旺", "旺", "得", "旺", "庙", "庙", "旺", "旺", "得", "旺"},
	"天机": {"庙", "陷", "得", "旺", "利", "平", "庙", "陷", "得", "旺", "利", "平"},
	"太阳": {"陷", "不", "旺", "庙", "旺", "旺", "旺", "得", "得", "陷", "不", "陷"},
	"武曲": {"旺", "庙", "得", "利", "庙", "平", "旺", "庙", "得", "利", "庙", "平"},
	"天同": {"旺", "不", "利", "平", "平", "庙", "陷", "不", "旺", "平", "平", "庙"},
	"廉贞": {"平", "利", "庙", "平", "利", "陷", "平", "利", "庙", "平", "利", "陷"},
	"天府": {"庙", "庙", "庙", "得", "庙", "得", "旺", "庙", "得", "旺", "庙", "得"},
	"太阴": {"庙", "庙", "旺", "陷", "陷", "陷", "不", "不", "利", "不", "旺", "庙"},
	"贪狼": {"旺", "庙", "平", "利", "庙", "陷", "旺", "庙", "平", "利", "庙", "陷"},
	"巨门": {"旺", "不", "庙", "庙", "陷", "旺", "旺", "不", "庙", "庙", "陷", "旺"},
	"天相": {"庙", "庙", "庙", "陷", "得", "得", "庙", "得", "庙", "陷", "得", "得"},
	"天梁": {"庙", "旺", "庙", "庙", "庙", "陷", "庙", "旺", "陷", "得", "庙", "陷"},
	"七杀": {"旺", "庙", "庙", "旺", "庙", "平", "旺", "庙", "庙", "庙", "庙", "平"},
	"破军": {"庙", "旺", "得", "陷", "旺", "平", "庙", "旺", "得", "陷", "旺", "平"},
}

// AuxStarBrightnessMap contains only auxiliary stars for which the pinned
// iztro STARS_INFO source defines a 12-branch brightness table.
var AuxStarBrightnessMap = map[string][12]string{
	"文昌": {"得", "庙", "陷", "利", "得", "庙", "陷", "利", "得", "庙", "陷", "利"},
	"文曲": {"得", "庙", "平", "旺", "得", "庙", "陷", "旺", "得", "庙", "陷", "旺"},
	"擎羊": {"陷", "庙", "", "陷", "庙", "", "陷", "庙", "", "陷", "庙", ""},
	"陀罗": {"", "庙", "陷", "", "庙", "陷", "", "庙", "陷", "", "庙", "陷"},
	"火星": {"陷", "得", "庙", "利", "陷", "得", "庙", "利", "陷", "得", "庙", "利"},
	"铃星": {"陷", "得", "庙", "利", "陷", "得", "庙", "利", "陷", "得", "庙", "利"},
}

// ──────────── Four Transformations Table (四化表) ────────────

// SiHuaTable maps year stem index → [化禄, 化权, 化科, 化忌].
// Source: SiHuaSourceRepo@SiHuaSourceCommit, SiHuaSourcePath (MIT).
// This is the only authoritative four-hua table used by chart, chain, flying
// star, and period consumers in this profile.
var SiHuaTable = [10][4]string{
	{"廉贞", "破军", "武曲", "太阳"}, // 甲
	{"天机", "天梁", "紫微", "太阴"}, // 乙
	{"天同", "天机", "文昌", "廉贞"}, // 丙
	{"太阴", "天同", "天机", "巨门"}, // 丁
	{"贪狼", "太阴", "右弼", "天机"}, // 戊
	{"武曲", "贪狼", "天梁", "文曲"}, // 己
	{"太阳", "武曲", "太阴", "天同"}, // 庚
	{"巨门", "太阳", "文曲", "文昌"}, // 辛
	{"天梁", "紫微", "左辅", "武曲"}, // 壬
	{"破军", "巨门", "太阴", "贪狼"}, // 癸
}

// SiHuaLabels are the four transformation labels.
var SiHuaLabels = [4]string{"化禄", "化权", "化科", "化忌"}

// ──────────── Auxiliary Star Placement Tables ────────────

// ZuofuTable: 左辅 from 辰(4), +%lunarMonth
func ZuofuIndex(lunarMonth int) int {
	return fixIndex(4 + lunarMonth - 1)
}

// YouyiTable: 右弼 from 戌(10), -%lunarMonth
func YoubiIndex(lunarMonth int) int {
	return fixIndex(10 - (lunarMonth - 1))
}

// WenchangTable: 文昌 from 戌(10), -%timeIndex
func WenchangIndex(timeIndex int) int {
	return fixIndex(10 - timeIndex)
}

// WenquTable: 文曲 from 辰(4), +%timeIndex
func WenquIndex(timeIndex int) int {
	return fixIndex(4 + timeIndex)
}

// KuiYueTable maps year stem → {天魁branch, 天钺branch}.
// 甲戊庚: 魁=丑(1), 钺=未(7); 乙己: 魁=子(0), 钺=申(8); etc.
var KuiYueTable = [10][2]int{
	{1, 7},  // 甲: 魁=丑(1), 钺=未(7)
	{0, 8},  // 乙: 魁=子(0), 钺=申(8)
	{11, 9}, // 丙: 魁=亥(11), 钺=酉(9)
	{11, 9}, // 丁: 魁=亥(11), 钺=酉(9)
	{1, 7},  // 戊: 魁=丑(1), 钺=未(7) (同甲)
	{0, 8},  // 己: 魁=子(0), 钺=申(8) (同乙)
	{1, 7},  // 庚: 魁=丑(1), 钺=未(7)
	{6, 2},  // 辛: 魁=午(6), 钺=寅(2)
	{3, 5},  // 壬: 魁=卯(3), 钺=巳(5)
	{3, 5},  // 癸: 魁=卯(3), 钺=巳(5)
}

// iztro location.ts getLuYangTuoMaIndex:
// 甲→寅(2), 乙→卯(3), 丙→巳(5), 丁→午(6), 戊→巳(5), 己→午(6),
// 庚→申(8), 辛→酉(9), 壬→亥(11), 癸→子(0)
var LucunBranchIdx = [10]int{
	2,  // 甲
	3,  // 乙
	5,  // 丙
	6,  // 丁
	5,  // 戊
	6,  // 己
	8,  // 庚
	9,  // 辛
	11, // 壬
	0,  // 癸
}

// LuCunTable is kept for compatibility with older tests/helpers.
var LuCunTable = LucunBranchIdx

// QingyangIndex = 禄存+1, TuoluoIndex = 禄存-1 (mod 12)
func QingyangIndex(yearStem int) int { return fixIndex(LucunBranchIdx[yearStem] + 1) }
func TuoluoIndex(yearStem int) int   { return fixIndex(LucunBranchIdx[yearStem] - 1) }

// TianmaTable: 天马 from year branch three-combination group.
// 申子辰→寅(2), 寅午戌→申(8), 巳酉丑→亥(11), 亥卯未→巳(5)
var TianmaBranchIdx = [12]int{
	2,  // 子→寅
	11, // 丑→亥
	8,  // 寅→申
	5,  // 卯→巳
	2,  // 辰→寅
	11, // 巳→亥
	8,  // 午→申
	5,  // 未→巳
	2,  // 申→寅
	11, // 酉→亥
	8,  // 戌→申
	5,  // 亥→巳
}

// DiKongDiJieTable: 地空 from 亥(11)逆数, 地劫 from 亥(11)顺数
func DiKongIndex(timeIndex int) int { return fixIndex(11 - timeIndex) }
func DiJieIndex(timeIndex int) int  { return fixIndex(11 + timeIndex) }

// HuolingIndex returns (火星branch, 铃星branch) for a given year branch and time index.
func HuolingIndex(yearBranch, timeIndex int) (huoIdx, lingIdx int) {
	timeIndex = fixIndex(timeIndex)
	switch yearBranch {
	case 2, 6, 10: // 寅午戌
		return fixIndex(1 + timeIndex), fixIndex(3 + timeIndex)
	case 8, 0, 4: // 申子辰
		return fixIndex(2 + timeIndex), fixIndex(10 + timeIndex)
	case 5, 9, 1: // 巳酉丑
		return fixIndex(3 + timeIndex), fixIndex(10 + timeIndex)
	case 11, 3, 7: // 亥卯未
		return fixIndex(9 + timeIndex), fixIndex(10 + timeIndex)
	default:
		return 0, 0
	}
}

// ──────────── Adjective Star Tables ────────────

// HongLuanTable: 红鸾 starts from 卯(3), reverses by year branch.
// 口诀: 子年红鸾在卯, 逆数年支 → 红鸾 = fixIndex(3 - yearBranch)
func HongLuanIndex(yearBranch int) int { return fixIndex(3 - yearBranch) }
func TianXiIndex(yearBranch int) int   { return fixIndex(HongLuanIndex(yearBranch) + 6) }

// TianYaoTable: 天姚 from 丑(1), with 正月 at 丑.
func TianYaoIndex(lunarMonth int) int { return fixIndex(lunarMonth) }

// TianXingTable: 天刑 from 酉(9), with 正月 at 酉.
func TianXingIndex(lunarMonth int) int { return fixIndex(8 + lunarMonth) }

// Monthly adjective-star branches, indexed by the effective lunar month - 1.
var YueJieBranchByMonth = [12]int{8, 8, 10, 10, 0, 0, 2, 2, 4, 4, 6, 6}
var TianYueBranchByMonth = [12]int{10, 5, 4, 2, 7, 3, 11, 7, 2, 6, 10, 2}
var TianWuBranchByMonth = [12]int{5, 8, 2, 11, 5, 8, 2, 11, 5, 8, 2, 11}

// XianChiBranch maps year branch → 咸池 branch.
// 申子辰→酉(9), 寅午戌→卯(3), 巳酉丑→午(6), 亥卯未→子(0)
var XianChiBranch = [12]int{9, 6, 3, 0, 9, 6, 3, 0, 9, 6, 3, 0}

// HuaGaiBranch maps year branch → 华盖 branch.
// 申子辰→辰(4), 寅午戌→戌(10), 巳酉丑→丑(1), 亥卯未→未(7)
var HuaGaiBranch = [12]int{4, 1, 10, 7, 4, 1, 10, 7, 4, 1, 10, 7}

// GuChenBranch and GuaSuBranch map year branch → 孤辰/寡宿 branch.
var GuChenBranch = [12]int{2, 2, 5, 5, 5, 8, 8, 8, 11, 11, 11, 2}
var GuaSuBranch = [12]int{10, 10, 1, 1, 1, 4, 4, 4, 7, 7, 7, 10}

// PoSuiBranch maps year branch → 破碎 branch.
var PoSuiBranch = [12]int{5, 1, 9, 5, 1, 9, 5, 1, 9, 5, 1, 9}

// FeiLianBranch maps year branch → 蜚廉 branch.
var FeiLianBranch = [12]int{8, 9, 10, 5, 6, 7, 2, 3, 4, 11, 0, 1}

// YinShaBranch maps month index → 阴煞 branch.
var YinShaBranch = [12]int{2, 0, 10, 8, 6, 4, 2, 0, 10, 8, 6, 4}

// Irregular year-stem adjective-star branches.
var TianChuBranchByStem = [10]int{5, 6, 0, 5, 6, 8, 2, 6, 9, 11}
var TianGuanBranchByStem = [10]int{7, 4, 5, 2, 3, 9, 11, 9, 10, 6}
var TianFuAdjectiveBranchByStem = [10]int{9, 8, 0, 11, 3, 2, 6, 5, 6, 5}
var JieLuBranchByStem = [10]int{8, 6, 4, 2, 0, 8, 6, 4, 2, 0}
var KongWangBranchByStem = [10]int{9, 7, 5, 3, 1, 9, 7, 5, 3, 1}

// NianJieBranch maps year branch → 年解 branch.
var NianJieBranch = [12]int{10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0, 11}

// ──────────── Life Master / Body Master (命主/身主) ────────────

// LifeMasterTable maps 命宫地支 to 命主 star.
// 子→贪狼, 丑→巨门, 寅→禄存, 卯→文曲, 辰→廉贞, 巳→武曲,
// 午→破军, 未→武曲, 申→廉贞, 酉→文曲, 戌→禄存, 亥→巨门
var LifeMasterTable = [12]string{
	"贪狼", "巨门", "禄存", "文曲", "廉贞", "武曲",
	"破军", "武曲", "廉贞", "文曲", "禄存", "巨门",
}

// BodyMasterTable maps year branch to 身主 star.
// 子午→火星, 丑未→天相, 寅申→天梁, 卯酉→天同, 辰戌→文昌, 巳亥→天机
var BodyMasterTable = [12]string{
	"火星", "天相", "天梁", "天同", "文昌", "天机",
	"火星", "天相", "天梁", "天同", "文昌", "天机",
}

// ──────────── Twelve Shen Tables ────────────

// ChangshengStartBranch maps five bureau value → starting branch index.
// 水二局→申(8), 木三局→亥(11), 金四局→巳(5), 土五局→申(8), 火六局→寅(2)
var ChangshengStartBranch = map[int]int{
	2: 8,  // 水二局 → 申
	3: 11, // 木三局 → 亥
	4: 5,  // 金四局 → 巳
	5: 8,  // 土五局 → 申 (iztro: same as water)
	6: 2,  // 火六局 → 寅
}

// ChangSheng12 names.
var ChangSheng12 = [12]string{
	"长生", "沐浴", "冠带", "临官", "帝旺", "衰",
	"病", "死", "墓", "绝", "胎", "养",
}

// BoShi12 names.
var BoShi12 = [12]string{
	"博士", "力士", "青龙", "小耗", "将军", "奏书",
	"飞廉", "喜神", "病符", "大耗", "伏兵", "官府",
}

// SuiQian12 names (岁前12神).
var SuiQian12 = [12]string{
	"岁建", "晦气", "丧门", "贯索", "官符", "小耗",
	"大耗", "龙德", "白虎", "天德", "吊客", "病符",
}

// JiangQian12 names (将前12神).
var JiangQian12 = [12]string{
	"将星", "攀鞍", "岁驿", "息神", "华盖", "劫煞",
	"灾煞", "天煞", "指背", "咸池", "月煞", "亡神",
}

// JiangQianStartBranch maps year branch group → 将星 starting branch.
// 寅午戌→午(6), 申子辰→子(0), 巳酉丑→酉(9), 亥卯未→卯(3)
var JiangQianStartBranch = map[string]int{
	"寅午戌": 6,
	"申子辰": 0,
	"巳酉丑": 9,
	"亥卯未": 3,
}

// ──────────── Sanfang Sizheng (三方四正) ────────────

// SanfangSizheng returns the 三方四正 palace indices for a given palace.
// 三方: the palace itself + two palaces 4 steps away (clockwise).
// 四正: add the opposite palace (6 steps away).
func SanfangSizheng(palaceIdx int) []int {
	return []int{
		palaceIdx,
		(palaceIdx + 4) % 12,
		(palaceIdx + 8) % 12,
		(palaceIdx + 6) % 12,
	}
}

// ──────────── Helper ────────────

func fixIndex(idx int) int {
	idx = idx % 12
	if idx < 0 {
		idx += 12
	}
	return idx
}

func fixIndexN(idx, n int) int {
	idx = idx % n
	if idx < 0 {
		idx += n
	}
	return idx
}
