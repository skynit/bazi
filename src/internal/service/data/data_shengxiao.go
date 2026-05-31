package data

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
	SanHe   []string // 三合
	LiuHe   []string // 六合
	SanHui  []string // 三会
	Chong   []string // 六冲
	Xing    []string // 三刑
	BeiXing []string // 被刑
	Hai     []string // 六害
	Po      []string // 相破
}

// ZhiAttrMap holds zhi attribute data.
var ZhiAttrMap = map[string]ZhiAttributes{
	"子": {
		SanHe:   []string{"申", "辰"},
		LiuHe:   []string{"丑"},
		SanHui:  []string{"亥", "丑"},
		Chong:   []string{"午"},
		Xing:    []string{"卯"},
		BeiXing: []string{"卯"},
		Hai:     []string{"未"},
		Po:      []string{"酉"},
	},
	"丑": {
		SanHe:   []string{"巳", "酉"},
		LiuHe:   []string{"子"},
		SanHui:  []string{"亥", "子"},
		Chong:   []string{"未"},
		Xing:    []string{"戌"},
		BeiXing: []string{"未"},
		Hai:     []string{"午"},
		Po:      []string{"辰"},
	},
	"寅": {
		SanHe:   []string{"午", "戌"},
		LiuHe:   []string{"亥"},
		SanHui:  []string{"卯", "辰"},
		Chong:   []string{"申"},
		Xing:    []string{"巳"},
		BeiXing: []string{"申"},
		Hai:     []string{"巳"},
		Po:      []string{"亥"},
	},
	"卯": {
		SanHe:   []string{"亥", "未"},
		LiuHe:   []string{"戌"},
		SanHui:  []string{"寅", "辰"},
		Chong:   []string{"酉"},
		Xing:    []string{"子"},
		BeiXing: []string{"子"},
		Hai:     []string{"辰"},
		Po:      []string{"午"},
	},
	"辰": {
		SanHe:   []string{"子", "申"},
		LiuHe:   []string{"酉"},
		SanHui:  []string{"寅", "卯"},
		Chong:   []string{"戌"},
		Xing:    []string{"辰"},
		BeiXing: []string{"辰"},
		Hai:     []string{"卯"},
		Po:      []string{"丑"},
	},
	"巳": {
		SanHe:   []string{"酉", "丑"},
		LiuHe:   []string{"申"},
		SanHui:  []string{"午", "未"},
		Chong:   []string{"亥"},
		Xing:    []string{"申"},
		BeiXing: []string{"寅"},
		Hai:     []string{"寅"},
		Po:      []string{"申"},
	},
	"午": {
		SanHe:   []string{"寅", "戌"},
		LiuHe:   []string{"未"},
		SanHui:  []string{"巳", "未"},
		Chong:   []string{"子"},
		Xing:    []string{"午"},
		BeiXing: []string{"午"},
		Hai:     []string{"丑"},
		Po:      []string{"卯"},
	},
	"未": {
		SanHe:   []string{"亥", "卯"},
		LiuHe:   []string{"午"},
		SanHui:  []string{"巳", "午"},
		Chong:   []string{"丑"},
		Xing:    []string{"丑"},
		BeiXing: []string{"戌"},
		Hai:     []string{"子"},
		Po:      []string{"戌"},
	},
	"申": {
		SanHe:   []string{"子", "辰"},
		LiuHe:   []string{"巳"},
		SanHui:  []string{"酉", "戌"},
		Chong:   []string{"寅"},
		Xing:    []string{"寅"},
		BeiXing: []string{"巳"},
		Hai:     []string{"亥"},
		Po:      []string{"巳"},
	},
	"酉": {
		SanHe:   []string{"巳", "丑"},
		LiuHe:   []string{"辰"},
		SanHui:  []string{"申", "戌"},
		Chong:   []string{"卯"},
		Xing:    []string{"酉"},
		BeiXing: []string{"酉"},
		Hai:     []string{"戌"},
		Po:      []string{"子"},
	},
	"戌": {
		SanHe:   []string{"寅", "午"},
		LiuHe:   []string{"卯"},
		SanHui:  []string{"申", "酉"},
		Chong:   []string{"辰"},
		Xing:    []string{"未"},
		BeiXing: []string{"丑"},
		Hai:     []string{"酉"},
		Po:      []string{"未"},
	},
	"亥": {
		SanHe:   []string{"卯", "未"},
		LiuHe:   []string{"寅"},
		SanHui:  []string{"子", "丑"},
		Chong:   []string{"巳"},
		Xing:    []string{"亥"},
		BeiXing: []string{"亥"},
		Hai:     []string{"申"},
		Po:      []string{"寅"},
	},
}
