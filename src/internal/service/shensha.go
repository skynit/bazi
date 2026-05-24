package service

import (
	"strings"

	"bazi/internal/model"
)

type shenShaPillars struct {
	Year   model.Pillar
	Month  model.Pillar
	Day    model.Pillar
	Hour   model.Pillar
	Gender string // "MALE" or "FEMALE"
}

type shenShaCalcResult struct {
	Year   []string
	Month  []string
	Day    []string
	Hour   []string
	Global []string
}

type shenShaRule struct {
	Name   string
	Target string
}

type sanHeShenSha struct {
	Jiang    string
	HuaGai   string
	YiMa     string
	JieSha   string
	ZaiSha   string
	WangShen string
	XianChi  string
}

var yearGanShenShaRules = map[string][]shenShaRule{
	"甲": {{"天乙贵人", "丑未"}, {"太极贵人", "子午"}, {"流霞", "酉"}, {"血刃", "卯"}, {"血忌", "丑"}},
	"乙": {{"天乙贵人", "子申"}, {"太极贵人", "子午"}, {"流霞", "戌"}, {"血刃", "辰"}, {"血忌", "未"}},
	"丙": {{"天乙贵人", "亥酉"}, {"太极贵人", "卯酉"}, {"流霞", "未"}, {"血刃", "午"}, {"血忌", "寅"}},
	"丁": {{"天乙贵人", "亥酉"}, {"太极贵人", "卯酉"}, {"流霞", "申"}, {"血刃", "未"}, {"血忌", "申"}},
	"戊": {{"天乙贵人", "丑未"}, {"太极贵人", "辰戌"}, {"流霞", "巳"}, {"血刃", "午"}, {"血忌", "卯"}},
	"己": {{"天乙贵人", "子申"}, {"太极贵人", "辰戌"}, {"流霞", "午"}, {"血刃", "未"}, {"血忌", "酉"}},
	"庚": {{"天乙贵人", "丑未"}, {"太极贵人", "寅申"}, {"流霞", "辰"}, {"血刃", "酉"}, {"血忌", "辰"}},
	"辛": {{"天乙贵人", "寅午"}, {"太极贵人", "寅申"}, {"流霞", "卯"}, {"血刃", "戌"}, {"血忌", "戌"}},
	"壬": {{"天乙贵人", "卯巳"}, {"太极贵人", "巳亥"}, {"流霞", "亥"}, {"血刃", "子"}, {"血忌", "巳"}},
	"癸": {{"天乙贵人", "卯巳"}, {"太极贵人", "巳亥"}, {"流霞", "寅"}, {"血刃", "丑"}, {"血忌", "亥"}},
}

// 年支神煞：按十二神顺排（太岁→太阳→丧门→太阴→官符→死符→岁破→龙德→白虎→福德→吊客→病符）
// 丧门(i+2) 吊客(i+10) 病符(i+11) 死符(i+5) 官符(i+4) 大耗=岁破(i+6) 小耗=病符前一位(i+10) 宅煞(i+1) 白虎(i+8) 飞廉(i+5) 的煞(i+11)
// 参考：《三命通会》《协纪辨方》
var yearZhiShenShaRules = map[string][]shenShaRule{
	"子": {{"丧门", "寅"}, {"吊客", "戌"}, {"病符", "亥"}, {"死符", "巳"}, {"官符", "辰"}, {"大耗", "午"}, {"小耗", "戌"}, {"宅煞", "丑"}, {"白虎", "申"}, {"飞廉", "巳"}, {"的煞", "亥"}},
	"丑": {{"丧门", "卯"}, {"吊客", "亥"}, {"病符", "子"}, {"死符", "午"}, {"官符", "巳"}, {"大耗", "未"}, {"小耗", "亥"}, {"宅煞", "寅"}, {"白虎", "酉"}, {"飞廉", "午"}, {"的煞", "子"}},
	"寅": {{"丧门", "辰"}, {"吊客", "子"}, {"病符", "丑"}, {"死符", "未"}, {"官符", "午"}, {"大耗", "申"}, {"小耗", "子"}, {"宅煞", "卯"}, {"白虎", "戌"}, {"飞廉", "未"}, {"的煞", "丑"}},
	"卯": {{"丧门", "巳"}, {"吊客", "丑"}, {"病符", "寅"}, {"死符", "申"}, {"官符", "未"}, {"大耗", "酉"}, {"小耗", "丑"}, {"宅煞", "辰"}, {"白虎", "亥"}, {"飞廉", "申"}, {"的煞", "寅"}},
	"辰": {{"丧门", "午"}, {"吊客", "寅"}, {"病符", "卯"}, {"死符", "酉"}, {"官符", "申"}, {"大耗", "戌"}, {"小耗", "寅"}, {"宅煞", "巳"}, {"白虎", "子"}, {"飞廉", "酉"}, {"的煞", "卯"}},
	"巳": {{"丧门", "未"}, {"吊客", "卯"}, {"病符", "辰"}, {"死符", "戌"}, {"官符", "酉"}, {"大耗", "亥"}, {"小耗", "卯"}, {"宅煞", "午"}, {"白虎", "丑"}, {"飞廉", "戌"}, {"的煞", "辰"}},
	"午": {{"丧门", "申"}, {"吊客", "辰"}, {"病符", "巳"}, {"死符", "亥"}, {"官符", "戌"}, {"大耗", "子"}, {"小耗", "辰"}, {"宅煞", "未"}, {"白虎", "寅"}, {"飞廉", "亥"}, {"的煞", "巳"}},
	"未": {{"丧门", "酉"}, {"吊客", "巳"}, {"病符", "午"}, {"死符", "子"}, {"官符", "亥"}, {"大耗", "丑"}, {"小耗", "巳"}, {"宅煞", "申"}, {"白虎", "卯"}, {"飞廉", "子"}, {"的煞", "午"}},
	"申": {{"丧门", "戌"}, {"吊客", "午"}, {"病符", "未"}, {"死符", "丑"}, {"官符", "子"}, {"大耗", "寅"}, {"小耗", "午"}, {"宅煞", "酉"}, {"白虎", "辰"}, {"飞廉", "丑"}, {"的煞", "未"}},
	"酉": {{"丧门", "亥"}, {"吊客", "未"}, {"病符", "申"}, {"死符", "寅"}, {"官符", "丑"}, {"大耗", "卯"}, {"小耗", "未"}, {"宅煞", "戌"}, {"白虎", "巳"}, {"飞廉", "寅"}, {"的煞", "申"}},
	"戌": {{"丧门", "子"}, {"吊客", "申"}, {"病符", "酉"}, {"死符", "卯"}, {"官符", "寅"}, {"大耗", "辰"}, {"小耗", "申"}, {"宅煞", "亥"}, {"白虎", "午"}, {"飞廉", "卯"}, {"的煞", "酉"}},
	"亥": {{"丧门", "丑"}, {"吊客", "酉"}, {"病符", "戌"}, {"死符", "辰"}, {"官符", "卯"}, {"大耗", "巳"}, {"小耗", "酉"}, {"宅煞", "子"}, {"白虎", "未"}, {"飞廉", "辰"}, {"的煞", "戌"}},
}

// 月支神煞（按《三命通会》起法）
// 月刑：寅刑巳 卯刑子 辰刑辰 巳刑申 午刑午 未刑丑 申刑寅 酉刑酉 戌刑未 亥刑亥 子刑卯 丑刑戌
// 月煞：正月(寅)起戌，逆行四季
// 天赦已由 isTianShe() 统一按季节判断，不入此表
var monthZhiShenShaRules = map[string][]shenShaRule{
	"子": {{"天刑", "亥"}, {"飞廉", "寅"}, {"天火", "午"}, {"天贼", "午"}, {"月煞", "子"}, {"月刑", "卯"}, {"月害", "未"}, {"月厌", "子"}, {"大时", "酉"}, {"兵禁", "申"}, {"天吏", "卯"}, {"致死", "午"}},
	"丑": {{"天刑", "子"}, {"飞廉", "酉"}, {"天火", "酉"}, {"天贼", "巳"}, {"月煞", "亥"}, {"月刑", "戌"}, {"月害", "午"}, {"月厌", "亥"}, {"大时", "午"}, {"兵禁", "亥"}, {"天吏", "子"}, {"致死", "巳"}},
	"寅": {{"天刑", "丑"}, {"飞廉", "戌"}, {"天火", "子"}, {"天贼", "辰"}, {"月煞", "戌"}, {"月刑", "巳"}, {"月害", "巳"}, {"月厌", "戌"}, {"大时", "卯"}, {"兵禁", "寅"}, {"天吏", "酉"}, {"致死", "辰"}},
	"卯": {{"天刑", "寅"}, {"飞廉", "巳"}, {"天火", "卯"}, {"天贼", "卯"}, {"月煞", "酉"}, {"月刑", "子"}, {"月害", "辰"}, {"月厌", "酉"}, {"大时", "子"}, {"兵禁", "巳"}, {"天吏", "午"}, {"致死", "卯"}},
	"辰": {{"天刑", "卯"}, {"飞廉", "午"}, {"天火", "午"}, {"天贼", "寅"}, {"月煞", "申"}, {"月刑", "辰"}, {"月害", "卯"}, {"月厌", "申"}, {"大时", "酉"}, {"兵禁", "申"}, {"天吏", "卯"}, {"致死", "寅"}},
	"巳": {{"天刑", "辰"}, {"飞廉", "未"}, {"天火", "酉"}, {"天贼", "丑"}, {"月煞", "未"}, {"月刑", "申"}, {"月害", "寅"}, {"月厌", "未"}, {"大时", "午"}, {"兵禁", "亥"}, {"天吏", "子"}, {"致死", "丑"}},
	"午": {{"天刑", "巳"}, {"飞廉", "申"}, {"天火", "子"}, {"天贼", "子"}, {"月煞", "午"}, {"月刑", "午"}, {"月害", "丑"}, {"月厌", "午"}, {"大时", "卯"}, {"兵禁", "寅"}, {"天吏", "卯"}, {"致死", "子"}},
	"未": {{"天刑", "午"}, {"飞廉", "酉"}, {"天火", "卯"}, {"天贼", "亥"}, {"月煞", "巳"}, {"月刑", "丑"}, {"月害", "子"}, {"月厌", "巳"}, {"大时", "子"}, {"兵禁", "巳"}, {"天吏", "午"}, {"致死", "亥"}},
	"申": {{"天刑", "未"}, {"飞廉", "戌"}, {"天火", "午"}, {"天贼", "戌"}, {"月煞", "辰"}, {"月刑", "寅"}, {"月害", "亥"}, {"月厌", "辰"}, {"大时", "酉"}, {"兵禁", "申"}, {"天吏", "酉"}, {"致死", "戌"}},
	"酉": {{"天刑", "申"}, {"飞廉", "亥"}, {"天火", "酉"}, {"天贼", "酉"}, {"月煞", "卯"}, {"月刑", "酉"}, {"月害", "戌"}, {"月厌", "卯"}, {"大时", "午"}, {"兵禁", "亥"}, {"天吏", "子"}, {"致死", "酉"}},
	"戌": {{"天刑", "酉"}, {"飞廉", "子"}, {"天火", "子"}, {"天贼", "申"}, {"月煞", "寅"}, {"月刑", "未"}, {"月害", "酉"}, {"月厌", "寅"}, {"大时", "卯"}, {"兵禁", "寅"}, {"天吏", "卯"}, {"致死", "申"}},
	"亥": {{"天刑", "戌"}, {"飞廉", "丑"}, {"天火", "卯"}, {"天贼", "未"}, {"月煞", "丑"}, {"月刑", "亥"}, {"月害", "申"}, {"月厌", "丑"}, {"大时", "子"}, {"兵禁", "巳"}, {"天吏", "午"}, {"致死", "未"}},
}

// 日干神煞：福星贵人按经典口诀（甲寅乙丑丙子丁酉戊申己未庚午辛巳壬辰癸卯）
// 天乙贵人保留供时贵等查用（年干日干同源）
var dayGanShenShaRules = map[string][]shenShaRule{
	"甲": {{"禄神", "寅"}, {"羊刃", "卯"}, {"金舆", "辰"}, {"红艳煞", "午"}, {"文昌贵人", "巳"}, {"学堂", "亥"}, {"词馆", "寅"}, {"天厨食禄", "巳"}, {"福星贵人", "寅"}, {"国印贵人", "戌"}, {"天官", "未"}, {"天乙贵人", "丑未"}, {"太极贵人", "子午"}},
	"乙": {{"禄神", "卯"}, {"羊刃", "寅"}, {"金舆", "巳"}, {"红艳煞", "午"}, {"文昌贵人", "午"}, {"学堂", "午"}, {"词馆", "卯"}, {"天厨食禄", "午"}, {"福星贵人", "丑"}, {"国印贵人", "亥"}, {"天官", "辰"}, {"天乙贵人", "子申"}, {"太极贵人", "子午"}},
	"丙": {{"禄神", "巳"}, {"羊刃", "午"}, {"金舆", "未"}, {"红艳煞", "寅"}, {"文昌贵人", "申"}, {"学堂", "寅"}, {"词馆", "巳"}, {"天厨食禄", "巳"}, {"福星贵人", "子"}, {"国印贵人", "丑"}, {"天官", "巳"}, {"天乙贵人", "亥酉"}, {"太极贵人", "卯酉"}},
	"丁": {{"禄神", "午"}, {"羊刃", "巳"}, {"金舆", "申"}, {"红艳煞", "未"}, {"文昌贵人", "酉"}, {"学堂", "酉"}, {"词馆", "午"}, {"天厨食禄", "午"}, {"福星贵人", "酉"}, {"国印贵人", "寅"}, {"天官", "酉"}, {"天乙贵人", "亥酉"}, {"太极贵人", "卯酉"}},
	"戊": {{"禄神", "巳"}, {"羊刃", "午"}, {"金舆", "未"}, {"红艳煞", "辰"}, {"文昌贵人", "申"}, {"学堂", "申"}, {"词馆", "巳"}, {"天厨食禄", "申"}, {"福星贵人", "申"}, {"国印贵人", "丑"}, {"天官", "戌"}, {"天乙贵人", "丑未"}, {"太极贵人", "辰戌"}},
	"己": {{"禄神", "午"}, {"羊刃", "巳"}, {"金舆", "申"}, {"红艳煞", "辰"}, {"文昌贵人", "酉"}, {"学堂", "酉"}, {"词馆", "午"}, {"天厨食禄", "酉"}, {"福星贵人", "未"}, {"国印贵人", "寅"}, {"天官", "卯"}, {"天乙贵人", "子申"}, {"太极贵人", "辰戌"}},
	"庚": {{"禄神", "申"}, {"羊刃", "酉"}, {"金舆", "戌"}, {"红艳煞", "戌"}, {"文昌贵人", "亥"}, {"学堂", "巳"}, {"词馆", "申"}, {"天厨食禄", "亥"}, {"福星贵人", "午"}, {"国印贵人", "辰"}, {"天官", "亥"}, {"天乙贵人", "丑未"}, {"太极贵人", "寅申"}},
	"辛": {{"禄神", "酉"}, {"羊刃", "申"}, {"金舆", "亥"}, {"红艳煞", "酉"}, {"文昌贵人", "子"}, {"学堂", "子"}, {"词馆", "酉"}, {"天厨食禄", "子"}, {"福星贵人", "巳"}, {"国印贵人", "巳"}, {"天官", "申"}, {"天乙贵人", "寅午"}, {"太极贵人", "寅申"}},
	"壬": {{"禄神", "亥"}, {"羊刃", "子"}, {"金舆", "丑"}, {"红艳煞", "子"}, {"文昌贵人", "寅"}, {"学堂", "申"}, {"词馆", "亥"}, {"天厨食禄", "寅"}, {"福星贵人", "辰"}, {"国印贵人", "未"}, {"天官", "丑"}, {"天乙贵人", "卯巳"}, {"太极贵人", "巳亥"}},
	"癸": {{"禄神", "子"}, {"羊刃", "亥"}, {"金舆", "寅"}, {"红艳煞", "申"}, {"文昌贵人", "卯"}, {"学堂", "卯"}, {"词馆", "子"}, {"天厨食禄", "卯"}, {"福星贵人", "卯"}, {"国印贵人", "申"}, {"天官", "子"}, {"天乙贵人", "卯巳"}, {"太极贵人", "巳亥"}},
}

var sanHeShenShaRules = map[string]sanHeShenSha{
	"寅": {"午", "戌", "申", "亥", "子", "巳", "卯"}, "午": {"午", "戌", "申", "亥", "子", "巳", "卯"}, "戌": {"午", "戌", "申", "亥", "子", "巳", "卯"},
	"巳": {"酉", "丑", "亥", "寅", "卯", "申", "午"}, "酉": {"酉", "丑", "亥", "寅", "卯", "申", "午"}, "丑": {"酉", "丑", "亥", "寅", "卯", "申", "午"},
	"申": {"子", "辰", "寅", "巳", "午", "亥", "酉"}, "子": {"子", "辰", "寅", "巳", "午", "亥", "酉"}, "辰": {"子", "辰", "寅", "巳", "午", "亥", "酉"},
	"亥": {"卯", "未", "巳", "申", "酉", "寅", "子"}, "卯": {"卯", "未", "巳", "申", "酉", "寅", "子"}, "未": {"卯", "未", "巳", "申", "酉", "寅", "子"},
}

var specialDayShenShaRules = map[string][]string{
	"戊子": {"九丑日", "阴差阳错", "六秀日"}, "戊午": {"九丑日", "孤鸾煞", "阴差阳错", "六秀日", "羊刃", "十灵日"},
	"壬子": {"九丑日", "孤鸾煞", "羊刃"}, "壬午": {"九丑日"}, "乙卯": {"九丑日", "八专", "福德秀气"}, "乙酉": {"九丑日", "福德秀气"},
	"辛卯": {"九丑日", "阴差阳错"}, "辛酉": {"九丑日", "阴差阳错", "八专", "六秀日"}, "己卯": {"九丑日"}, "己酉": {"九丑日"},
	"甲辰": {"十恶大败", "十灵日"}, "乙巳": {"十恶大败", "孤鸾煞", "福德秀气"}, "丙申": {"十恶大败"}, "丁亥": {"十恶大败"},
	"戊戌": {"十恶大败", "魁罡", "八专"}, "庚辰": {"十恶大败", "魁罡", "日德"}, "辛巳": {"十恶大败"}, "壬申": {"十恶大败"}, "癸亥": {"十恶大败", "阴差阳错"},
	"壬辰": {"魁罡", "阴差阳错"}, "庚戌": {"魁罡", "八专", "十灵日"}, "甲寅": {"孤鸾煞", "八专", "日德"}, "丙午": {"阴差阳错", "孤鸾煞", "六秀日", "羊刃"},
	"丁巳": {"孤鸾煞"}, "辛亥": {"孤鸾煞", "十灵日"}, "戊申": {"阴差阳错", "孤鸾煞"}, "丙子": {"阴差阳错", "六秀日"},
	"丁丑": {"阴差阳错", "六秀日"}, "戊寅": {"阴差阳错"}, "癸巳": {"阴差阳错"}, "丁未": {"阴差阳错", "八专", "六秀日"}, "壬戌": {"阴差阳错", "日德"},
	"己未": {"八专", "六秀日", "羊刃"}, "庚申": {"八专", "词馆"}, "癸丑": {"八专", "羊刃"}, "乙丑": {"金神", "福德秀气"}, "己巳": {"金神"}, "癸酉": {"金神"},
	"丙辰": {"日德", "十灵日"}, "戊辰": {"日德"}, "己丑": {"六秀日"}, "乙亥": {"十灵日"}, "丁酉": {"十灵日"}, "庚寅": {"十灵日"}, "壬寅": {"十灵日"}, "癸未": {"十灵日"},
}

// --- 截路空亡：按日干查 ---
var jieLuKongWangByDayGan = map[string][]string{
	"甲": {"申", "酉"}, "己": {"申", "酉"},
	"乙": {"午", "未"}, "庚": {"午", "未"},
	"丙": {"辰", "巳"}, "辛": {"辰", "巳"},
	"丁": {"寅", "卯"}, "壬": {"寅", "卯"},
	"戊": {"子", "丑"}, "癸": {"子", "丑"},
}

var ganHe = map[string]string{"甲": "己", "己": "甲", "乙": "庚", "庚": "乙", "丙": "辛", "辛": "丙", "丁": "壬", "壬": "丁", "戊": "癸", "癸": "戊"}

// --- 年支追加规则：红鸾、天喜 ---
var hongLuan = map[string]string{
	"子": "卯", "丑": "寅", "寅": "丑", "卯": "子", "辰": "亥", "巳": "戌",
	"午": "酉", "未": "申", "申": "未", "酉": "午", "戌": "巳", "亥": "辰",
}
var tianXi = map[string]string{
	"子": "酉", "丑": "申", "寅": "未", "卯": "午", "辰": "巳", "巳": "辰",
	"午": "卯", "未": "寅", "申": "丑", "酉": "子", "戌": "亥", "亥": "戌",
}

// --- 三合局追加：六厄（死位）、墓煞（墓位） ---
var sanHeLiuE = map[string]string{
	"寅": "酉", "午": "酉", "戌": "酉",
	"巳": "子", "酉": "子", "丑": "子",
	"申": "卯", "子": "卯", "辰": "卯",
	"亥": "午", "卯": "午", "未": "午",
}
var sanHeMuSha = map[string]string{
	"寅": "戌", "午": "戌", "戌": "戌",
	"巳": "丑", "酉": "丑", "丑": "丑",
	"申": "辰", "子": "辰", "辰": "辰",
	"亥": "未", "卯": "未", "未": "未",
}

// --- 隔角煞 ---
var geJiaoPair = map[string]string{
	"丑": "寅", "寅": "丑", "卯": "辰", "辰": "卯", "巳": "午", "午": "巳",
	"未": "申", "申": "未", "酉": "戌", "戌": "酉", "亥": "子", "子": "亥",
}

// --- 年支追加：自缢煞（与岁破同，即年支冲） ---
var yearZhiChong = map[string]string{
	"子": "午", "丑": "未", "寅": "申", "卯": "酉", "辰": "戌", "巳": "亥",
	"午": "子", "未": "丑", "申": "寅", "酉": "卯", "戌": "辰", "亥": "巳",
}

// --- 月支追加：小时、天杀、大败 ---
var monthXiaoShi = map[string]string{
	"子": "卯", "丑": "子", "寅": "酉", "卯": "午", "辰": "卯", "巳": "子",
	"午": "酉", "未": "午", "申": "卯", "酉": "子", "戌": "酉", "亥": "午",
}
var monthTianSha = map[string]string{
	"子": "丑", "丑": "戌", "寅": "未", "卯": "辰", "辰": "丑", "巳": "戌",
	"午": "未", "未": "辰", "申": "丑", "酉": "戌", "戌": "未", "亥": "辰",
}
var monthDaBai = map[string]string{
	"子": "巳", "丑": "午", "寅": "酉", "卯": "子", "辰": "酉", "巳": "午",
	"午": "酉", "未": "子", "申": "卯", "酉": "午", "戌": "卯", "亥": "子",
}

// --- 四大空亡 ---
var siDaKongWang = map[string][]string{
	"甲子": {"水", "火"}, "甲午": {"水", "火"},
	"甲申": {"金", "木"}, "甲寅": {"金", "木"},
}

// --- 孤辰寡宿：按年支/日支所在三会方查找（非三合局） ---
var guChenBySanHui = map[string]string{
	"寅": "巳", "卯": "巳", "辰": "巳",
	"巳": "申", "午": "申", "未": "申",
	"申": "亥", "酉": "亥", "戌": "亥",
	"亥": "寅", "子": "寅", "丑": "寅",
}
var guSuBySanHui = map[string]string{
	"寅": "丑", "卯": "丑", "辰": "丑",
	"巳": "辰", "午": "辰", "未": "辰",
	"申": "未", "酉": "未", "戌": "未",
	"亥": "戌", "子": "戌", "丑": "戌",
}

// --- 禄神映射（复用 dayGanShenShaRules 中的禄神值做快速查表） ---
var luShenZhi = map[string]string{
	"甲": "寅", "乙": "卯", "丙": "巳", "丁": "午", "戊": "巳",
	"己": "午", "庚": "申", "辛": "酉", "壬": "亥", "癸": "子",
}
var zhiLiuHe = map[string]string{
	"子": "丑", "丑": "子", "寅": "亥", "卯": "戌", "辰": "酉", "巳": "申",
	"午": "未", "未": "午", "申": "巳", "酉": "辰", "戌": "卯", "亥": "寅",
}
var zhiLiuChong = map[string]string{
	"子": "午", "丑": "未", "寅": "申", "卯": "酉", "辰": "戌", "巳": "亥",
	"午": "子", "未": "丑", "申": "寅", "酉": "卯", "戌": "辰", "亥": "巳",
}

func calcDayShenShaOnly(dayGan, dayZhi string) []string {
	var result []string
	for _, rule := range dayGanShenShaRules[dayGan] {
		appendShenSha(&result, rule.Name, rule.Target)
	}
	if rule := sanHeShenShaRules[dayZhi]; rule.Jiang != "" {
		for _, item := range []shenShaRule{
			{"将星", rule.Jiang}, {"华盖", rule.HuaGai}, {"驿马", rule.YiMa}, {"劫煞", rule.JieSha}, {"灾煞", rule.ZaiSha},
			{"亡神", rule.WangShen}, {"桃花", rule.XianChi},
		} {
			appendShenSha(&result, item.Name, item.Target)
		}
	}
	for _, name := range specialDayShenShaRules[dayGan+dayZhi] {
		appendShenSha(&result, name, "")
	}
	return uniqueStrings(result)
}

func calcShenShaByPillars(p shenShaPillars) shenShaCalcResult {
	var res shenShaCalcResult
	allPillars := []model.Pillar{p.Year, p.Month, p.Day, p.Hour}
	allBranches := []string{p.Year.Zhi, p.Month.Zhi, p.Day.Zhi, p.Hour.Zhi}

	// --- 年干/年支/三合局规则：分柱匹配写入 ---
	addRulesToBucketByExactPillar(&res, yearGanShenShaRules[p.Year.Gan], allPillars)
	addRulesToBucketByExactPillar(&res, yearZhiShenShaRules[p.Year.Zhi], allPillars)
	addSanHeRulesToBucketByExactPillar(&res, sanHeShenShaRules[p.Year.Zhi], allPillars, false)
	addYearZhiExtra(p, allBranches, &res)
	addYearGanExtra(p, allBranches, &res)

	// --- 月柱 ---
	addMonthGanRules(p, &res)
	addRulesToBucketByExactPillar(&res, monthZhiShenShaRules[p.Month.Zhi], allPillars)
	addMonthZhiExtra(p, allBranches, &res)

	// --- 月空（按月柱旬空） ---
	addYueKong(p, &res)

	// --- 日柱 ---
	addRulesToBucketByExactPillar(&res, dayGanShenShaRules[p.Day.Gan], allPillars)
	addSanHeRulesToBucketByExactPillar(&res, sanHeShenShaRules[p.Day.Zhi], allPillars, true)
	for _, name := range specialDayShenShaRules[p.Day.Gan+p.Day.Zhi] {
		appendShenSha(&res.Day, name, "")
	}
	addDayExtra(p, allBranches, &res)

	// --- 孤辰寡宿（年支三会方） ---
	if guChenZhi := guChenBySanHui[p.Year.Zhi]; guChenZhi != "" {
		for i, zhi := range allBranches {
			if zhi == guChenZhi {
				switch i {
				case 0: appendShenSha(&res.Year, "孤辰", guChenZhi)
				case 1: appendShenSha(&res.Month, "孤辰", guChenZhi)
				case 2: appendShenSha(&res.Day, "孤辰", guChenZhi)
				case 3: appendShenSha(&res.Hour, "孤辰", guChenZhi)
				}
			}
		}
	}
	if guSuZhi := guSuBySanHui[p.Year.Zhi]; guSuZhi != "" {
		for i, zhi := range allBranches {
			if zhi == guSuZhi {
				switch i {
				case 0: appendShenSha(&res.Year, "寡宿", guSuZhi)
				case 1: appendShenSha(&res.Month, "寡宿", guSuZhi)
				case 2: appendShenSha(&res.Day, "寡宿", guSuZhi)
				case 3: appendShenSha(&res.Hour, "寡宿", guSuZhi)
				}
			}
		}
	}

	// --- 空亡 ---
	addKongWang(p, &res)

	// --- 时柱 ---
	addHourShenSha(p, &res)

	// --- 全局 ---
	addGlobalShenSha(p, &res)

	applyShenShaPriority(&res)
	return res
}

func addRulesToBucketByExactPillar(res *shenShaCalcResult, rules []shenShaRule, pillars []model.Pillar) {
	for _, rule := range rules {
		for i, pillar := range pillars {
			if matchGanZhiTarget(rule.Target, pillar) {
				switch i {
				case 0:
					appendShenSha(&res.Year, rule.Name, rule.Target)
				case 1:
					appendShenSha(&res.Month, rule.Name, rule.Target)
				case 2:
					appendShenSha(&res.Day, rule.Name, rule.Target)
				case 3:
					appendShenSha(&res.Hour, rule.Name, rule.Target)
				}
			}
		}
	}
}

func addSanHeRulesToBucketByExactPillar(res *shenShaCalcResult, rule sanHeShenSha, pillars []model.Pillar, includeDayDerived bool) {
	if rule.Jiang == "" {
		return
	}
	rules := []shenShaRule{
		{"将星", rule.Jiang}, {"华盖", rule.HuaGai}, {"驿马", rule.YiMa},
	}
	if includeDayDerived {
		rules = append(rules,
			shenShaRule{"劫煞", rule.JieSha}, shenShaRule{"灾煞", rule.ZaiSha}, shenShaRule{"亡神", rule.WangShen},
			shenShaRule{"桃花", rule.XianChi},
		)
	}
	addRulesToBucketByExactPillar(res, rules, pillars)
}

func addMonthGanRules(p shenShaPillars, res *shenShaCalcResult) {
	monthIndex := ZhiIndex(p.Month.Zhi)
	tianDe := MonthShenMap[TianDe][monthIndex]
	yueDe := MonthShenMap[YueDe][monthIndex]
	for _, target := range []struct {
		name string
		gan  string
	}{
		{"天德贵人", tianDe}, {"月德贵人", yueDe}, {"天德合", ganHe[tianDe]}, {"月德合", ganHe[yueDe]},
	} {
		if target.gan == "" {
			continue
		}
		if p.Month.Gan == target.gan {
			appendShenSha(&res.Month, target.name, target.gan)
		}
		if p.Day.Gan == target.gan {
			appendShenSha(&res.Day, target.name, target.gan)
		}
		if p.Hour.Gan == target.gan {
			appendShenSha(&res.Hour, target.name, target.gan)
		}
	}
	for _, pair := range []struct {
		pillar *[]string
		gan    string
	}{
		{&res.Year, p.Year.Gan},
		{&res.Month, p.Month.Gan},
		{&res.Day, p.Day.Gan},
		{&res.Hour, p.Hour.Gan},
	} {
		if hasDeXiu(p.Month.Zhi, pair.gan) {
			appendShenSha(pair.pillar, "德秀贵人", pair.gan)
		}
	}
}

func addYueKong(p shenShaPillars, res *shenShaCalcResult) {
	for _, zhi := range getKongWangZhi(p.Month.Gan, p.Month.Zhi) {
		if p.Year.Zhi == zhi {
			appendShenSha(&res.Year, "月空", zhi)
		}
		if p.Month.Zhi == zhi {
			appendShenSha(&res.Month, "月空", zhi)
		}
		if p.Day.Zhi == zhi {
			appendShenSha(&res.Day, "月空", zhi)
		}
		if p.Hour.Zhi == zhi {
			appendShenSha(&res.Hour, "月空", zhi)
		}
	}
}

func addKongWang(p shenShaPillars, res *shenShaCalcResult) {
	for _, zhi := range getKongWangZhi(p.Day.Gan, p.Day.Zhi) {
		if p.Year.Zhi == zhi {
			appendShenSha(&res.Year, "空亡", zhi)
		}
		if p.Month.Zhi == zhi {
			appendShenSha(&res.Month, "空亡", zhi)
		}
		if p.Hour.Zhi == zhi {
			appendShenSha(&res.Hour, "空亡", zhi)
		}
	}
}

func addHourShenSha(p shenShaPillars, res *shenShaCalcResult) {
	// 时桃花：日支的桃花（咸池）落在时支
	if rule := sanHeShenShaRules[p.Day.Zhi]; rule.XianChi == p.Hour.Zhi {
		appendShenSha(&res.Hour, "时桃花", p.Hour.Zhi)
	}
	for _, rule := range dayGanShenShaRules[p.Day.Gan] {
		if !targetContainsZhi(rule.Target, p.Hour.Zhi) {
			continue
		}
		switch rule.Name {
		case "天乙贵人":
			appendShenSha(&res.Hour, "时贵", p.Hour.Zhi)
		case "羊刃":
			appendShenSha(&res.Hour, "时刃", p.Hour.Zhi)
		case "禄神":
			appendShenSha(&res.Hour, "时禄", p.Hour.Zhi)
		}
	}
	if rule := sanHeShenShaRules[p.Day.Zhi]; rule.YiMa == p.Hour.Zhi {
		appendShenSha(&res.Hour, "时马", p.Hour.Zhi)
	}
	if rule := sanHeShenShaRules[p.Day.Zhi]; rule.JieSha == p.Hour.Zhi || rule.WangShen == p.Hour.Zhi {
		appendShenSha(&res.Hour, "时煞", p.Hour.Zhi)
	}
	if isTongZiSha(p.Month.Zhi, p.Hour.Zhi) {
		appendShenSha(&res.Hour, "童子煞", p.Hour.Zhi)
	}
	for _, zhi := range jieLuKongWangByDayGan[p.Day.Gan] {
		if zhi == p.Hour.Zhi {
			appendShenSha(&res.Hour, "截路空亡", zhi)
		}
	}
}

func addGlobalShenSha(p shenShaPillars, res *shenShaCalcResult) {
	addExtraGlobalRules(p, res)
	for _, seq := range []string{p.Year.Gan + p.Month.Gan + p.Day.Gan, p.Month.Gan + p.Day.Gan + p.Hour.Gan} {
		switch seq {
		case "甲戊庚":
			appendShenSha(&res.Global, "天三奇", seq)
		case "乙丙丁":
			appendShenSha(&res.Global, "地三奇", seq)
		case "壬癸辛":
			appendShenSha(&res.Global, "人三奇", seq)
		}
	}
	dayPillar := p.Day.Gan + p.Day.Zhi
	if isTianShe(p.Month.Zhi, dayPillar) {
		appendShenSha(&res.Global, "天赦", dayPillar)
	}
	if isSiFei(p.Month.Zhi, dayPillar) {
		appendShenSha(&res.Global, "四废", dayPillar)
	}
	yearNayin := Nayin[GanIndex(p.Year.Gan)][ZhiIndex(p.Year.Zhi)]
	if entry := NaYinMap[yearNayin]; entry.Element == "火" {
		if p.Day.Zhi == "戌" || p.Day.Zhi == "亥" || p.Hour.Zhi == "戌" || p.Hour.Zhi == "亥" {
			appendShenSha(&res.Global, "天罗", "戌亥")
		}
	}
	if entry := NaYinMap[yearNayin]; entry.Element == "水" || entry.Element == "土" {
		if p.Day.Zhi == "辰" || p.Day.Zhi == "巳" || p.Hour.Zhi == "辰" || p.Hour.Zhi == "巳" {
			appendShenSha(&res.Global, "地网", "辰巳")
		}
	}
	if hasSpecialDay(p.Day.Gan+p.Day.Zhi, "金神") && (GanElement[p.Hour.Gan] == "火" || ZhiElement[p.Hour.Zhi] == "火") {
		appendShenSha(&res.Global, "金神成格", p.Day.Gan+p.Day.Zhi)
	}
	// 拱禄/拱贵
	addGongLuGongGui(p, res)
	// 龙虎拱命
	addLongHuGongMing(p, res)
	// 凤凰拱命
	addFengHuangGongMing(p, res)
	// 紫微星
	addZiWeiXing(p, res)
	// 龙德
	addLongDe(p, res)
	// 天元暗禄、飞禄
	addGlobalLuDerived(p, res)
}

func addGongLuGongGui(p shenShaPillars, res *shenShaCalcResult) {
	dIdx := ZhiIndex(p.Day.Zhi)
	hIdx := ZhiIndex(p.Hour.Zhi)
	if dIdx < 0 || hIdx < 0 {
		return
	}
	diff := hIdx - dIdx
	if diff == 2 || diff == -10 {
		middle := Zhis[(dIdx+1)%12]
		if middle == luShenZhi[p.Day.Gan] {
			appendShenSha(&res.Global, "拱禄", middle)
		}
		// 拱贵：中间地支是天乙贵人
		for _, rule := range dayGanShenShaRules[p.Day.Gan] {
			if rule.Name == "天乙贵人" && targetContainsZhi(rule.Target, middle) {
				appendShenSha(&res.Global, "拱贵", middle)
				break
			}
		}
	}
}

// --- 全局追加：水溺煞、天火煞、挂剑煞、雷霆煞 ---
func addExtraGlobalRules(p shenShaPillars, res *shenShaCalcResult) {
	branches := []string{p.Year.Zhi, p.Month.Zhi, p.Day.Zhi, p.Hour.Zhi}

	// 天火煞：年支和日支同在三合局，且天干有丙丁
	if sameSanHeGroup(p.Year.Zhi, p.Day.Zhi) {
		anyGan := []string{p.Year.Gan, p.Month.Gan, p.Day.Gan, p.Hour.Gan}
		for _, g := range anyGan {
			if g == "丙" || g == "丁" {
				appendShenSha(&res.Global, "天火煞", p.Year.Zhi+p.Day.Zhi)
				break
			}
		}
	}
	// 挂剑煞：四柱地支全含巳酉丑申
	needed := []string{"巳", "酉", "丑", "申"}
	if coversAll(needed, branches) {
		appendShenSha(&res.Global, "挂剑煞", "巳酉丑申")
	}
	// 雷霆煞：寅卯辰月（正月二月三月）见子午卯酉
	if p.Month.Zhi == "寅" || p.Month.Zhi == "卯" || p.Month.Zhi == "辰" {
		for _, b := range branches {
			if b == "子" || b == "午" || b == "卯" || b == "酉" {
				appendShenSha(&res.Global, "雷霆煞", b)
				break
			}
		}
	}
}

func sameSanHeGroup(a, b string) bool {
	rA := sanHeShenShaRules[a]
	rB := sanHeShenShaRules[b]
	return rA.Jiang != "" && rA.Jiang == rB.Jiang
}

func coversAll(needed, branches []string) bool {
	for _, n := range needed {
		if !branchInList(n, branches) {
			return false
		}
	}
	return true
}

func addLongHuGongMing(p shenShaPillars, res *shenShaCalcResult) {
	branches := []string{p.Year.Zhi, p.Month.Zhi, p.Day.Zhi, p.Hour.Zhi}
	hasChen := branchInList("辰", branches)
	hasYin := branchInList("寅", branches)
	if hasChen && hasYin {
		appendShenSha(&res.Global, "龙虎拱命", "辰寅")
	}
}

func addFengHuangGongMing(p shenShaPillars, res *shenShaCalcResult) {
	branches := []string{p.Year.Zhi, p.Month.Zhi, p.Day.Zhi, p.Hour.Zhi}
	hasYou := branchInList("酉", branches)
	hasSi := branchInList("巳", branches)
	if hasYou && hasSi {
		appendShenSha(&res.Global, "凤凰拱命", "酉巳")
	}
}

func addZiWeiXing(p shenShaPillars, res *shenShaCalcResult) {
	rule := sanHeShenShaRules[p.Year.Zhi]
	if rule.Jiang != "" && p.Day.Zhi == rule.Jiang {
		appendShenSha(&res.Global, "紫微星", p.Day.Zhi)
	}
}

func addLongDe(p shenShaPillars, res *shenShaCalcResult) {
	yRule := sanHeShenShaRules[p.Year.Zhi]
	dRule := sanHeShenShaRules[p.Day.Zhi]
	if yRule.Jiang == dRule.Jiang && yRule.Jiang != "" {
		monthTianDe := MonthShenMap[TianDe][ZhiIndex(p.Month.Zhi)]
		if monthTianDe != "" {
			appendShenSha(&res.Global, "龙德", p.Month.Zhi)
		}
	}
}

func addGlobalLuDerived(p shenShaPillars, res *shenShaCalcResult) {
	branches := []string{p.Year.Zhi, p.Month.Zhi, p.Day.Zhi, p.Hour.Zhi}
	luZhi := luShenZhi[p.Day.Gan]
	if luZhi == "" {
		return
	}
	if heZhi := zhiLiuHe[luZhi]; heZhi != "" && branchInList(heZhi, branches) {
		appendShenSha(&res.Global, "天元暗禄", heZhi)
	}
	if chongZhi := zhiLiuChong[luZhi]; chongZhi != "" && branchInList(chongZhi, branches) {
		appendShenSha(&res.Global, "飞禄", chongZhi)
	}
	// 交禄（全局）
	for _, pair := range []string{p.Year.Gan, p.Month.Gan, p.Hour.Gan} {
		if ganHe[p.Day.Gan] == pair && pair != "" {
			partnerLu := luShenZhi[pair]
			if partnerLu != "" && branchInList(partnerLu, branches) {
				appendShenSha(&res.Global, "交禄", partnerLu)
			}
			break
		}
	}
}

// --- 年支额外规则 ---
func addYearZhiExtra(p shenShaPillars, branches []string, res *shenShaCalcResult) {
	checkAndAppend := func(name, target string) {
		if branchInList(target, branches) {
			appendShenSha(&res.Year, name, target)
		}
	}
	if t := hongLuan[p.Year.Zhi]; t != "" {
		checkAndAppend("红鸾", t)
	}
	if t := tianXi[p.Year.Zhi]; t != "" {
		checkAndAppend("天喜", t)
	}
	if t := sanHeLiuE[p.Year.Zhi]; t != "" {
		checkAndAppend("六厄", t)
	}
	if t := sanHeMuSha[p.Year.Zhi]; t != "" {
		checkAndAppend("墓煞", t)
	}
	if t := yearZhiChong[p.Year.Zhi]; t != "" {
		checkAndAppend("自缢煞", t)
	}
	if t := geJiaoPair[p.Year.Zhi]; t != "" && branchInList(t, branches) {
		appendShenSha(&res.Year, "隔角煞", t)
	}
	// 孤寡煞 = 孤辰+寡宿同时命中
	if guChenZhi := guChenBySanHui[p.Year.Zhi]; guChenZhi != "" {
		if guSuZhi := guSuBySanHui[p.Year.Zhi]; guSuZhi != "" {
			hasGu := branchInList(guChenZhi, branches)
			hasGua := branchInList(guSuZhi, branches)
			if hasGu && hasGua {
				appendShenSha(&res.Year, "孤寡煞", guChenZhi+guSuZhi)
			}
		}
	}
	// 性别规则：元辰、勾绞煞、暴败煞
	addGenderBasedShenSha(p, branches, res)
}

// --- 年干额外规则 ---
func addYearGanExtra(p shenShaPillars, branches []string, res *shenShaCalcResult) {
	// 金锁煞：甲己年见申, 乙庚年见未...
	if t := jinSuoShaTarget(p.Year.Gan); t != "" && branchInList(t, branches) {
		appendShenSha(&res.Year, "金锁煞", t)
	}
	// 天刑煞：年干年支组合定时支
	if t := tianXingShaBranch(p.Year.Gan, p.Year.Zhi); t != "" && t == p.Hour.Zhi {
		appendShenSha(&res.Year, "天刑煞", t)
	}
}

// --- 月支额外规则 ---
func addMonthZhiExtra(p shenShaPillars, branches []string, res *shenShaCalcResult) {
	// 小时：月支查，若时支匹配则加到时柱
	if t := monthXiaoShi[p.Month.Zhi]; t != "" && t == p.Hour.Zhi {
		appendShenSha(&res.Hour, "小时", t)
	}
	if t := monthTianSha[p.Month.Zhi]; t != "" && branchInList(t, branches) {
		appendShenSha(&res.Month, "天杀", t)
	}
	if t := monthDaBai[p.Month.Zhi]; t != "" && branchInList(t, branches) {
		appendShenSha(&res.Month, "大败", t)
	}
}

// --- 日柱额外规则 ---
func addDayExtra(p shenShaPillars, branches []string, res *shenShaCalcResult) {
	luZhi := luShenZhi[p.Day.Gan]
	if luZhi == "" {
		return
	}
	if luZhi == p.Month.Zhi {
		appendShenSha(&res.Day, "建禄", luZhi)
	}
	if heZhi := zhiLiuHe[luZhi]; heZhi != "" && branchInList(heZhi, branches) {
		appendShenSha(&res.Day, "暗禄", heZhi)
		appendShenSha(&res.Day, "天元暗禄", heZhi)
	}
	if chongZhi := zhiLiuChong[luZhi]; chongZhi != "" && branchInList(chongZhi, branches) {
		appendShenSha(&res.Day, "飞禄", chongZhi)
	}
	// 交禄：日干与年/月/时干合且合干之禄在地支
	for _, pair := range []struct{ gan, pillar string }{
		{p.Year.Gan, "年"}, {p.Month.Gan, "月"}, {p.Hour.Gan, "时"},
	} {
		if ganHe[p.Day.Gan] == pair.gan && pair.gan != "" {
			partnerLu := luShenZhi[pair.gan]
			if partnerLu != "" && branchInList(partnerLu, branches) {
				appendShenSha(&res.Day, "交禄", partnerLu)
			}
			break
		}
	}
	// 四大空亡
	if empty := siDaKongWang[p.Day.Gan+p.Day.Zhi]; empty != nil {
		for _, elem := range empty {
			if anyBranchHasElement(elem, branches) {
				appendShenSha(&res.Day, "四大空亡", elem)
			}
		}
	}
}

// --- 性别规则：元辰、勾绞煞、暴败煞 ---
func addGenderBasedShenSha(p shenShaPillars, branches []string, res *shenShaCalcResult) {
	chongZhi := yearZhiChong[p.Year.Zhi]
	if chongZhi == "" {
		return
	}
	chongIdx := ZhiIndex(chongZhi)
	front := Zhis[(chongIdx-1+12)%12] // 冲支前一位
	back := Zhis[(chongIdx+1)%12]     // 冲支后一位

	isYang := p.Year.Gan == "甲" || p.Year.Gan == "丙" || p.Year.Gan == "戊" || p.Year.Gan == "庚" || p.Year.Gan == "壬"
	isMale := p.Gender == "MALE"
	yangNanYinNv := (isYang && isMale) || (!isYang && !isMale) // 阳男阴女

	var yuanChenTarget, gouTarget, jiaoTarget string
	if yangNanYinNv {
		yuanChenTarget = front // 冲前一位
		gouTarget = front
		jiaoTarget = back
	} else {
		yuanChenTarget = back // 冲后一位
		gouTarget = back
		jiaoTarget = front
	}
	if branchInList(yuanChenTarget, branches) {
		appendShenSha(&res.Year, "元辰", yuanChenTarget)
	}
	hasGou := branchInList(gouTarget, branches)
	hasJiao := branchInList(jiaoTarget, branches)
	if hasGou || hasJiao {
		label := gouTarget
		if hasJiao {
			label = jiaoTarget
		}
		appendShenSha(&res.Year, "勾绞煞", label)
	}
	if hasGou || hasJiao {
		appendShenSha(&res.Year, "暴败煞", gouTarget+jiaoTarget)
	}
}

func branchInList(target string, branches []string) bool {
	for _, b := range branches {
		if b == target {
			return true
		}
	}
	return false
}

func jinSuoShaTarget(yearGan string) string {
	switch yearGan {
	case "甲", "己":
		return "申"
	case "乙", "庚":
		return "未"
	case "丙", "辛":
		return "辰"
	case "丁", "壬":
		return "卯"
	case "戊", "癸":
		return "寅"
	default:
		return ""
	}
}

func tianXingShaBranch(gan, zhi string) string {
	// 子丑人戌 etc.
	switch zhi {
	case "子", "丑":
		return "戌"
	case "寅", "卯":
		return "巳"
	case "辰", "巳":
		return "辰"
	case "午", "未":
		return "午"
	case "申", "酉":
		return "未"
	case "戌", "亥":
		return "亥"
	default:
		return ""
	}
}

func anyBranchHasElement(elem string, branches []string) bool {
	for _, b := range branches {
		if ZhiElement[b] == elem {
			return true
		}
	}
	return false
}

func appendShenSha(items *[]string, name, target string) {
	desc := ShenInfoMap[name]
	if target != "" && desc != "" {
		*items = append(*items, name+"："+target+"｜"+desc)
		return
	}
	if target != "" {
		*items = append(*items, name+"："+target)
		return
	}
	if desc != "" {
		*items = append(*items, name+"："+desc)
		return
	}
	*items = append(*items, name)
}

func matchGanZhiTarget(target string, p model.Pillar) bool {
	if target == "" {
		return false
	}
	if len([]rune(target)) == 2 && target == p.Gan+p.Zhi {
		return true
	}
	return targetContainsGan(target, p.Gan) || targetContainsZhi(target, p.Zhi)
}

func targetContainsGan(target, gan string) bool {
	for _, r := range []rune(target) {
		if string(r) == gan {
			return true
		}
	}
	return false
}

func targetContainsZhi(target, zhi string) bool {
	for _, r := range []rune(target) {
		if string(r) == zhi {
			return true
		}
	}
	return false
}

func getKongWangZhi(dayGan, dayZhi string) []string {
	dayIndex := sixtyCycleIndex(dayGan, dayZhi)
	if dayIndex < 0 {
		return nil
	}
	switch dayIndex / 10 {
	case 0:
		return []string{"戌", "亥"}
	case 1:
		return []string{"申", "酉"}
	case 2:
		return []string{"午", "未"}
	case 3:
		return []string{"辰", "巳"}
	case 4:
		return []string{"寅", "卯"}
	default:
		return []string{"子", "丑"}
	}
}

func sixtyCycleIndex(gan, zhi string) int {
	for i := 0; i < 60; i++ {
		if Gans[i%10] == gan && Zhis[i%12] == zhi {
			return i
		}
	}
	return -1
}

func hasDeXiu(monthZhi, gan string) bool {
	switch monthZhi {
	case "寅", "午", "戌":
		return gan == "丙" || gan == "丁" || gan == "戊" || gan == "癸"
	case "申", "子", "辰":
		return gan == "壬" || gan == "癸" || gan == "戊" || gan == "己"
	case "亥", "卯", "未":
		return gan == "甲" || gan == "乙" || gan == "己" || gan == "庚"
	case "巳", "酉", "丑":
		return gan == "庚" || gan == "辛" || gan == "乙" || gan == "丙"
	default:
		return false
	}
}

func isTongZiSha(monthZhi, hourZhi string) bool {
	switch monthZhi {
	case "寅", "卯", "辰", "申", "酉", "戌":
		return hourZhi == "寅" || hourZhi == "子"
	case "亥", "子", "丑":
		return hourZhi == "卯" || hourZhi == "未"
	case "巳", "午", "未":
		return hourZhi == "辰" || hourZhi == "戌"
	default:
		return false
	}
}

func isTianShe(monthZhi, dayPillar string) bool {
	switch monthZhi {
	case "寅", "卯", "辰":
		return dayPillar == "戊寅"
	case "巳", "午", "未":
		return dayPillar == "甲午"
	case "申", "酉", "戌":
		return dayPillar == "戊申"
	case "亥", "子", "丑":
		return dayPillar == "甲子"
	default:
		return false
	}
}

func isSiFei(monthZhi, dayPillar string) bool {
	switch monthZhi {
	case "寅", "卯", "辰":
		return dayPillar == "庚申" || dayPillar == "辛酉"
	case "巳", "午", "未":
		return dayPillar == "壬子" || dayPillar == "癸亥"
	case "申", "酉", "戌":
		return dayPillar == "甲寅" || dayPillar == "乙卯"
	case "亥", "子", "丑":
		return dayPillar == "丙午" || dayPillar == "丁巳"
	default:
		return false
	}
}

func hasSpecialDay(dayPillar, name string) bool {
	for _, value := range specialDayShenShaRules[dayPillar] {
		if value == name {
			return true
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func applyShenShaPriority(res *shenShaCalcResult) {
	res.Day = uniqueShenShaByNameLocal(res.Day)
	res.Year = uniqueShenShaByNameLocal(res.Year)
	res.Month = uniqueShenShaByNameLocal(res.Month)
	res.Hour = uniqueShenShaByNameLocal(res.Hour)
	res.Global = uniqueShenShaByNameLocal(res.Global)
}

func uniqueShenShaByNameLocal(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		name := shenShaName(value)
		if seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, value)
	}
	return result
}

func shenShaName(value string) string {
	if idx := strings.Index(value, "："); idx >= 0 {
		return value[:idx]
	}
	if idx := strings.Index(value, "｜"); idx >= 0 {
		return value[:idx]
	}
	return value
}
