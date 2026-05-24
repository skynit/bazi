package service

// GanIndex returns the index (0-9) of a Gan stem.
// Panics if gan is not a valid stem.
func GanIndex(gan string) int {
	switch gan {
	case "甲":
		return 0
	case "乙":
		return 1
	case "丙":
		return 2
	case "丁":
		return 3
	case "戊":
		return 4
	case "己":
		return 5
	case "庚":
		return 6
	case "辛":
		return 7
	case "壬":
		return 8
	case "癸":
		return 9
	default:
		panic("invalid gan: " + gan)
	}
}

// ZhiIndex returns the index (0-11) of a Zhi branch.
// Panics if zhi is not a valid branch.
func ZhiIndex(zhi string) int {
	switch zhi {
	case "子":
		return 0
	case "丑":
		return 1
	case "寅":
		return 2
	case "卯":
		return 3
	case "辰":
		return 4
	case "巳":
		return 5
	case "午":
		return 6
	case "未":
		return 7
	case "申":
		return 8
	case "酉":
		return 9
	case "戌":
		return 10
	case "亥":
		return 11
	default:
		panic("invalid zhi: " + zhi)
	}
}

// Gans are the 10 heavenly stems in order.
var Gans = [10]string{
	"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸",
}

// Zhis are the 12 earthly branches in order.
var Zhis = [12]string{
	"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥",
}

// GanElement maps stem character to its wuxing element string.
var GanElement = map[string]string{
	"甲": "木", "乙": "木", "丙": "火", "丁": "火", "戊": "土",
	"己": "土", "庚": "金", "辛": "金", "壬": "水", "癸": "水",
}

// ZhiElement maps branch character to its wuxing element string.
var ZhiElement = map[string]string{
	"子": "水", "丑": "土", "寅": "木", "卯": "木",
	"辰": "土", "巳": "火", "午": "火", "未": "土",
	"申": "金", "酉": "金", "戌": "土", "亥": "水",
}

// NaYinEntry holds the full na-yin knowledge for one na-yin type.
type NaYinEntry struct {
	Name          string   // e.g. "杨柳木"
	StemBranches  []string // e.g. {"壬午", "癸未"}
	Element       string   // "木"
	ImageDesc     string   // 取象释义
	Personality   string   // 性格/命运隐喻
	EnergyStage   string   // 能量阶段 "死墓"
	ModernExt     string   // 现代延伸联想
	Judgments     []string // 特质/断语
}

// NaYin is the na-yin (纳音) classification indexed by gan index (0-9) and zhi index (0-11).
// Empty string for combinations not defined.
var Nayin [10][12]string

// NaYinMap maps na-yin name (e.g. "杨柳木") to its full entry.
var NaYinMap map[string]NaYinEntry

// Empties is the "空亡" (emptiness) pair indexed by gan index (0-9) and zhi index (0-11).
// Both strings empty when no entry.
var Empties [10][12][2]string

// MingGong is the "命宫" interpretation indexed by zhi index (0-11).
var MingGong [12]string

// RiZhuDesc holds "日主坐命" descriptions keyed by stem+branch (e.g. "甲子").
var RiZhuDesc map[string]string

func init() {
	// Nayin
	nayinData := [][2]string{
		{"甲", "子"}, {"乙", "丑"}, {"壬", "寅"}, {"癸", "卯"},
		{"庚", "辰"}, {"辛", "巳"}, {"甲", "午"}, {"乙", "未"},
		{"壬", "申"}, {"癸", "酉"}, {"庚", "戌"}, {"辛", "亥"},
		{"戊", "子"}, {"己", "丑"}, {"丙", "寅"}, {"丁", "卯"},
		{"甲", "辰"}, {"乙", "巳"}, {"戊", "午"}, {"己", "未"},
		{"丙", "申"}, {"丁", "酉"}, {"甲", "戌"}, {"乙", "亥"},
		{"壬", "子"}, {"癸", "丑"}, {"庚", "寅"}, {"辛", "卯"},
		{"戊", "辰"}, {"己", "巳"}, {"壬", "午"}, {"癸", "未"},
		{"庚", "申"}, {"辛", "酉"}, {"戊", "戌"}, {"己", "亥"},
		{"庚", "子"}, {"辛", "丑"}, {"戊", "寅"}, {"己", "卯"},
		{"丙", "辰"}, {"丁", "巳"}, {"庚", "午"}, {"辛", "未"},
		{"戊", "申"}, {"己", "酉"}, {"丙", "戌"}, {"丁", "亥"},
		{"丙", "子"}, {"丁", "丑"}, {"甲", "寅"}, {"乙", "卯"},
		{"壬", "辰"}, {"癸", "巳"}, {"丙", "午"}, {"丁", "未"},
		{"甲", "申"}, {"乙", "酉"}, {"壬", "戌"}, {"癸", "亥"},
	}
	nayinValues := []string{
		"海中金", "海中金", "金泊金", "金泊金",
		"白蜡金", "白蜡金", "砂中金", "砂中金",
		"剑锋金", "剑锋金", "钗钏金", "钗钏金",
		"霹雳火", "霹雳火", "炉中火", "炉中火",
		"覆灯火", "覆灯火", "天上火", "天上火",
		"山下火", "山下火", "山头火", "山头火",
		"桑柘木", "桑柘木", "松柏木", "松柏木",
		"大林木", "大林木", "杨柳木", "杨柳木",
		"石榴木", "石榴木", "平地木", "平地木",
		"壁上土", "壁上土", "城头土", "城头土",
		"砂中土", "砂中土", "路旁土", "路旁土",
		"大驿土", "大驿土", "屋上土", "屋上土",
		"涧下水", "涧下水", "大溪水", "大溪水",
		"长流水", "长流水", "天河水", "天河水",
		"井泉水", "井泉水", "大海水", "大海水",
	}
	for i, pair := range nayinData {
		Nayin[GanIndex(pair[0])][ZhiIndex(pair[1])] = nayinValues[i]
	}

	// NaYinMap: build full-entry map from knowledge base
	NaYinMap = make(map[string]NaYinEntry)
	for _, entry := range NaYinKnowledge {
		NaYinMap[entry.Name] = entry
	}

	// Empties
	emptiesData := [][2]string{
		{"甲", "子"}, {"乙", "丑"},
		{"丙", "寅"}, {"丁", "卯"},
		{"戊", "辰"}, {"己", "巳"},
		{"庚", "午"}, {"辛", "未"},
		{"壬", "申"}, {"癸", "酉"},

		{"甲", "戌"}, {"乙", "亥"},
		{"丙", "子"}, {"丁", "丑"},
		{"戊", "寅"}, {"己", "卯"},
		{"庚", "辰"}, {"辛", "巳"},
		{"壬", "午"}, {"癸", "未"},

		{"甲", "申"}, {"乙", "酉"},
		{"丙", "戌"}, {"丁", "亥"},
		{"戊", "子"}, {"己", "丑"},
		{"庚", "寅"}, {"辛", "卯"},
		{"壬", "辰"}, {"癸", "巳"},

		{"甲", "午"}, {"乙", "未"},
		{"丙", "申"}, {"丁", "酉"},
		{"戊", "戌"}, {"己", "亥"},
		{"庚", "子"}, {"辛", "丑"},
		{"壬", "寅"}, {"癸", "卯"},

		{"甲", "辰"}, {"乙", "巳"},
		{"丙", "午"}, {"丁", "未"},
		{"戊", "申"}, {"己", "酉"},
		{"庚", "戌"}, {"辛", "亥"},
		{"壬", "子"}, {"癸", "丑"},

		{"甲", "寅"}, {"乙", "卯"},
		{"丙", "辰"}, {"丁", "巳"},
		{"戊", "午"}, {"己", "未"},
		{"庚", "申"}, {"辛", "酉"},
		{"壬", "戌"}, {"癸", "亥"},
	}
	emptiesValues := [][2]string{
		{"戌", "亥"}, {"戌", "亥"},
		{"戌", "亥"}, {"戌", "亥"},
		{"戌", "亥"}, {"戌", "亥"},
		{"戌", "亥"}, {"戌", "亥"},
		{"戌", "亥"}, {"戌", "亥"},

		{"申", "酉"}, {"申", "酉"},
		{"申", "酉"}, {"申", "酉"},
		{"申", "酉"}, {"申", "酉"},
		{"申", "酉"}, {"申", "酉"},
		{"申", "酉"}, {"申", "酉"},

		{"午", "未"}, {"午", "未"},
		{"午", "未"}, {"午", "未"},
		{"午", "未"}, {"午", "未"},
		{"午", "未"}, {"午", "未"},
		{"午", "未"}, {"午", "未"},

		{"辰", "巳"}, {"辰", "巳"},
		{"辰", "巳"}, {"辰", "巳"},
		{"辰", "巳"}, {"辰", "巳"},
		{"辰", "巳"}, {"辰", "巳"},
		{"辰", "巳"}, {"辰", "巳"},

		{"寅", "卯"}, {"寅", "卯"},
		{"寅", "卯"}, {"寅", "卯"},
		{"寅", "卯"}, {"寅", "卯"},
		{"寅", "卯"}, {"寅", "卯"},
		{"寅", "卯"}, {"寅", "卯"},

		{"子", "丑"}, {"子", "丑"},
		{"子", "丑"}, {"子", "丑"},
		{"子", "丑"}, {"子", "丑"},
		{"子", "丑"}, {"子", "丑"},
		{"子", "丑"}, {"子", "丑"},
	}
	for i, pair := range emptiesData {
		Empties[GanIndex(pair[0])][ZhiIndex(pair[1])] = emptiesValues[i]
	}

	// MingGong
	MingGong = [12]string{
		"天贵星、志气不凡、富裕清吉。",
		"天厄星、先难后吉、离祖劳心、晚年吉。",
		"天权星、聪明大器、中年有权柄。",
		"天赦星、慷慨疏财、得权时须谦逊。",
		"天如星、事多翻覆、机谋多能。",
		"天文星、文章振发、女命有好夫。",
		"天福星、荣华吉命。",
		"天驿星、一生劳碌、离祖始安。",
		"天孤星、不宜早婚、女命妨夫。",
		"天秘星、性情刚直、时有是非。",
		"天艺星、心性平和、艺道有名。",
		"天寿星、心慈明悟、克己助人。",
	}

	// RiZhuDesc
	RiZhuDesc = map[string]string{
		"甲子": "虽坐沐浴，若四往有禄，看印，冬生不作妻败",
		"乙丑": "身坐财官，有乙庚合最吉",
		"丙寅": "金绝水死，财官俱背，但丙火长生食神独旺，主有寿，己亥、辛卯、癸巳时贵",
		"丁卯": "财官俱背，须合气、禄、火扶",
		"戊辰": "壬庚入墓，乙木自坐财官",
		"己巳": "水绝木病，丙寅时贵",
		"庚午": "庚金坐死但午上自坐官、印，虽败不困",
		"辛未": "身旺，丙申时贵",
		"壬申": "水辰生位，聪明秀丽",
		"癸酉": "财官无气，要用旺者吉",
		"甲戌": "身坐旺官，临火库，心怀慈善，丙寅时贵",
		"乙亥": "日坐木局，丙壬、壬午、甲申时贵",
		"丙子": "身坐财，癸巳时",
		"丁丑": "金库荣丰，见辛亥时贵",
		"戊寅": "甲木当局，官杀者吉",
		"己卯": "身坐杀地，须身杀力停者吉",
		"庚辰": "魁罡，忌于刑冲",
		"辛巳": "金局坐死不妨，戊子时贵",
		"壬午": "财官双美，伶俐有谋，壬寅时贵",
		"癸未": "身坐杀位须身力二停",
		"甲申": "坐绝，四柱俱绝者吉",
		"乙酉": "坐杀四乙酉或有化杀则吉，辛巳时，为化气金局贵",
		"丙戌": "夏生则财官无气",
		"丁亥": "日贵，壬寅时，乙巳时皆贵",
		"戊子": "自坐财，乙卯时，丁巳时贵",
		"己丑": "有财无官，丙寅时贵",
		"庚寅": "坐绝反主吉昌",
		"辛卯": "财衰无妨，见戊子时贵",
		"壬辰": "魁罡、不喜冲刑，遇建禄反卑",
		"癸巳": "财官双美最吉祥，丁已时贵",
		"甲午": "夏生大吉",
		"乙未": "逢财伤官格",
		"丙申": "身坐财，庚寅时贵，癸巳时亦吉",
		"丁酉": "临财学精，壬寅时贵",
		"戊戌": "魁罡，忌冲刑",
		"己亥": "自坐财官得高名，丙寅时贵",
		"庚子": "有丁火则吉",
		"辛丑": "食神荣昌，主寿",
		"壬寅": "水火既济，见壬寅时大吉",
		"癸卯": "日贵，衰神旺吉",
		"甲辰": "身坐财库水气，性善良，丙寅时吉",
	}
}
