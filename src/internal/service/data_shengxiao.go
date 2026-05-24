package service

// ShengXiao holds zodiac animal data.
// Index order matches zhi order: 子=0→鼠, 丑=1→牛, ..., 亥=11→猪.
var ShengXiao = [12]string{
	"鼠", "牛", "虎", "兔", "龙", "蛇",
	"马", "羊", "猴", "鸡", "狗", "猪",
}

//ShengXiaoToZhi is the reverse lookup: zodiac animal → zhi index.
var ShengXiaoToZhi = map[string]int{
	"鼠": 0, "牛": 1, "虎": 2, "兔": 3,
	"龙": 4, "蛇": 5, "马": 6, "羊": 7,
	"猴": 8, "鸡": 9, "狗": 10, "猪": 11,
}

// ZhiAttributes holds the relation attributes of a zhi.
type ZhiAttributes struct {
	合    []string // 六合
	六    []string // 六和
	会    []string // 三会
	冲    []string // 六冲
	刑    []string // 三刑
	被刑  []string // 被刑
	害    []string // 六害
	破    []string // 相破
}

// ZhiAttrMap holds zhi attribute data.
var ZhiAttrMap = map[string]ZhiAttributes{
	"子": {
		合:   []string{"申", "辰"},
		六:   []string{"丑"},
		会:   []string{"申", "辰"},
		冲:   []string{"午"},
		刑:   []string{"卯"},
		被刑: []string{"卯"},
		害:   []string{"未"},
		破:   []string{"酉"},
	},
	"丑": {
		合:   []string{"巳", "酉"},
		六:   []string{"子"},
		会:   []string{"巳", "酉"},
		冲:   []string{"未"},
		刑:   []string{"戌"},
		被刑: []string{"戌", "酉"},
		害:   []string{"午"},
		破:   []string{"寅"},
	},
	"寅": {
		合:   []string{"亥"},
		六:   []string{"午"},
		会:   []string{"亥", "午"},
		冲:   []string{"申"},
		刑:   []string{"巳"},
		被刑: []string{"巳", "亥"},
		害:   []string{"申"},
		破:   []string{"亥"},
	},
	"卯": {
		合:   []string{"未"},
		六:   []string{"子"},
		会:   []string{"未", "子"},
		冲:   []string{"酉"},
		刑:   []string{"子"},
		被刑: []string{"子"},
		害:   []string{"辰"},
		破:   []string{"子"},
	},
	"辰": {
		合:   []string{"子", "申"},
		六:   []string{"酉"},
		会:   []string{"子", "申"},
		冲:   []string{"戌"},
		刑:   []string{"戌"},
		被刑: []string{"戌"},
		害:   []string{"卯"},
		破:   []string{"丑"},
	},
	"巳": {
		合:   []string{"酉", "丑"},
		六:   []string{"亥"},
		会:   []string{"酉", "丑"},
		冲:   []string{"亥"},
		刑:   []string{"寅"},
		被刑: []string{"寅", "申"},
		害:   []string{"寅"},
		破:   []string{"申"},
	},
	"午": {
		合:   []string{"未"},
		六:   []string{"寅"},
		会:   []string{"未", "寅"},
		冲:   []string{"子"},
		刑:   []string{"戌"},
		被刑: []string{"戌", "未"},
		害:   []string{"丑"},
		破:   []string{"亥"},
	},
	"未": {
		合:   []string{"午", "亥"},
		六:   []string{"丑"},
		会:   []string{"午", "亥"},
		冲:   []string{"丑"},
		刑:   []string{"丑"},
		被刑: []string{"丑"},
		害:   []string{"子"},
		破:   []string{"辰"},
	},
	"申": {
		合:   []string{"巳", "子"},
		六:   []string{"辰"},
		会:   []string{"巳", "子"},
		冲:   []string{"寅"},
		刑:   []string{"亥"},
		被刑: []string{"亥", "寅"},
		害:   []string{"亥"},
		破:   []string{"巳"},
	},
	"酉": {
		合:   []string{"辰", "巳"},
		六:   []string{"卯"},
		会:   []string{"辰", "巳"},
		冲:   []string{"卯"},
		刑:   []string{"戌"},
		被刑: []string{"戌", "辰"},
		害:   []string{"戌"},
		破:   []string{"子"},
	},
	"戌": {
		合:   []string{"寅", "午"},
		六:   []string{"卯"},
		会:   []string{"寅", "午"},
		冲:   []string{"辰"},
		刑:   []string{"丑"},
		被刑: []string{"丑", "午"},
		害:   []string{"酉"},
		破:   []string{"卯"},
	},
	"亥": {
		合:   []string{"寅"},
		六:   []string{"巳"},
		会:   []string{"寅", "巳"},
		冲:   []string{"巳"},
		刑:   []string{"申"},
		被刑: []string{"申", "寅"},
		害:   []string{"申"},
		破:   []string{"酉"},
	},
}