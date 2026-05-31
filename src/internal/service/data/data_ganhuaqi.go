package data

// JinJinFuWeiEntry describes a 进神/退神/交神/伏神 day's properties.
// Source: 《三命通会》Chapter 31 "论进交退伏神" (阎东叟).
type JinJinFuWeiEntry struct {
	StemBranch string // 干支，如 "甲子"
	Type       string // 进神/交神/退神/伏神
	Fortune    string // 吉凶断语
}

// JinJinFuWeiKnowledge 进神退神交神伏神知识库。
// Key: stem+branch, value: entry.
var JinJinFuWeiKnowledge = map[string]JinJinFuWeiEntry{
	// 甲己组
	"甲子": {StemBranch: "甲子", Type: "进神", Fortune: "发迹亨快，仕途顺达事业亨通"},
	"丙子": {StemBranch: "丙子", Type: "交神", Fortune: "庶事不谐，诸事摩擦难以协调"},
	"丁丑": {StemBranch: "丁丑", Type: "退神", Fortune: "官资降黜，地位下降声望受损"},
	"戊寅": {StemBranch: "戊寅", Type: "伏神", Fortune: "所作滞留，谋事拖延成果难现"},
	"己卯": {StemBranch: "己卯", Type: "进神", Fortune: "发迹亨快，仕途顺达事业亨通"},
	"辛卯": {StemBranch: "辛卯", Type: "交神", Fortune: "庶事不谐，诸事摩擦难以协调"},
	"壬辰": {StemBranch: "壬辰", Type: "退神", Fortune: "官资降黜，地位下降声望受损"},
	"癸巳": {StemBranch: "癸巳", Type: "伏神", Fortune: "所作滞留，谋事拖延成果难现"},
	"甲午": {StemBranch: "甲午", Type: "进神", Fortune: "发迹亨快，仕途顺达事业亨通"},
	"丙午": {StemBranch: "丙午", Type: "交神", Fortune: "庶事不谐，诸事摩擦难以协调"},
	"丁未": {StemBranch: "丁未", Type: "退神", Fortune: "官资降黜，地位下降声望受损"},
	"戊申": {StemBranch: "戊申", Type: "伏神", Fortune: "所作滞留，谋事拖延成果难现"},
	"己酉": {StemBranch: "己酉", Type: "进神", Fortune: "发迹亨快，仕途顺达事业亨通"},
	"辛酉": {StemBranch: "辛酉", Type: "交神", Fortune: "庶事不谐，诸事摩擦难以协调"},
	"壬戌": {StemBranch: "壬戌", Type: "退神", Fortune: "官资降黜，地位下降声望受损"},
	"癸亥": {StemBranch: "癸亥", Type: "伏神", Fortune: "所作滞留，谋事拖延成果难现"},
}

// GanHuaQiEntry 描述十干化气合的条件与喜忌。
// Source: 《三命通会》Chapter 32 "论十干化气".
type GanHuaQiEntry struct {
	Stem1     string   // 天干1，如 "甲"
	Stem2     string   // 天干2，如 "己"
	HuaQi     string   // 化气五行，如 "土"
	Deity     string   // 德统龙之神，如 "戊德统龙"
	TianShi   string   // 天时之气，如 "钧天土气"
	Conditions []string // 化气条件
	MonthRange string  // 主要化气月令
	Favor     []string // 喜用
	Taboo     []string // 忌见
	LuckyHour string   // 最佳时辰
}

// GanHuaQiKnowledge 十干化气知识库。
var GanHuaQiKnowledge = map[string]GanHuaQiEntry{
	"甲己": {
		Stem1:       "甲",
		Stem2:       "己",
		HuaQi:       "土",
		Deity:       "甲德统龙",
		TianShi:     "钧天土气",
		Conditions:  []string{"辰戌丑未月大吉", "午月亦可化（有戊字间之则不化，名妒合）"},
		MonthRange:   "辰戌丑未月（其次午月）",
		Favor:       []string{"木为官（亥卯未为官）", "戊癸气为福"},
		Taboo:       []string{"忌见丁壬日时"},
		LuckyHour:    "戊辰时生，四季月土成象",
	},
	"乙庚": {
		Stem1:       "乙",
		Stem2:       "庚",
		HuaQi:       "金",
		Deity:       "庚德统龙",
		TianShi:     "颢天金气",
		Conditions:  []string{"巳酉丑月大吉", "七月亦可化（有甲字间之则不化，名妒合）"},
		MonthRange:   "巳酉丑月（其次七月）",
		Favor:       []string{"火为官（喜丙丁己午甲己为福）"},
		Taboo:       []string{"忌见戊癸日时"},
		LuckyHour:    "",
	},
	"丙辛": {
		Stem1:       "丙",
		Stem2:       "辛",
		HuaQi:       "水",
		Deity:       "壬德统龙",
		TianShi:     "玄天水气",
		Conditions:  []string{"申子辰月大吉", "十月亦可化（柱有丁字不化，名妒合）"},
		MonthRange:   "申子辰月（其次十月）",
		Favor:       []string{"土为官（得辰戌丑未为官）", "乙庚为福"},
		Taboo:       []string{"忌见甲己日时"},
		LuckyHour:    "",
	},
	"丁壬": {
		Stem1:       "丁",
		Stem2:       "壬",
		HuaQi:       "木",
		Deity:       "甲德统龙",
		TianShi:     "苍天木气",
		Conditions:  []string{"亥卯未月大吉", "正月亦可化（柱有丙字不化，名妒合）"},
		MonthRange:   "亥卯未月（其次正月）",
		Favor:       []string{"庚辛申酉为官", "丙辛为福"},
		Taboo:       []string{"忌见乙庚日时"},
		LuckyHour:    "",
	},
	"戊癸": {
		Stem1:       "戊",
		Stem2:       "癸",
		HuaQi:       "火",
		Deity:       "丙德统龙",
		TianShi:     "炎天火气",
		Conditions:  []string{"寅午戌月大吉", "四月亦可化（柱有己字不化，名妒合）"},
		MonthRange:   "寅午戌月（其次四月）",
		Favor:       []string{"壬癸亥子为官", "丁壬为福"},
		Taboo:       []string{"忌见丙辛日时"},
		LuckyHour:    "",
	},
}

// GetJinJinFuWei returns the JinJinFuWei entry for a given stem+branch.
func GetJinJinFuWei(gz string) (JinJinFuWeiEntry, bool) {
	e, ok := JinJinFuWeiKnowledge[gz]
	return e, ok
}

// GetGanHuaQi returns the 化气 entry for a given stem pair (shorter first).
func GetGanHuaQi(s1, s2 string) (GanHuaQiEntry, bool) {
	key := s1 + s2
	if key == "" {
		return GanHuaQiEntry{}, false
	}
	e, ok := GanHuaQiKnowledge[key]
	if ok {
		return e, true
	}
	// try reversed
	key2 := s2 + s1
	e, ok = GanHuaQiKnowledge[key2]
	return e, ok
}