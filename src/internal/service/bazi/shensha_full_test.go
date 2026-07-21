package bazi

import (
	"fmt"
	"strings"
	"testing"
)

// =============================================================================
// 神煞全量测试
// 覆盖所有20+神煞类型，使用 CalculateFromPillars
// =============================================================================

// shenShaFullCase 全量神煞测试用例
type shenShaFullCase struct {
	ID, Desc, Source                string
	YearP, MonP, DayP, HouP, Gender string
	Asserts                         []shenShaAssert // 正向断言（应包含）
	NegAsserts                      []shenShaAssert // 反向断言（应不包含）
}

var shenShaFullCases = []shenShaFullCase{
	// ========== 天乙贵人 ==========
	{ID: "SSF-001", Desc: "乙日天乙贵人在子申—年支子→年柱有贵人",
		YearP: "丙子", MonP: "己亥", DayP: "乙丑", HouP: "壬午", Gender: "MALE",
		Asserts: []shenShaAssert{{"year", "天乙贵人"}}, Source: "三命通会PDF第97-98页"},
	{ID: "SSF-002", Desc: "甲日天乙贵人丑未—时支未→时柱有天乙贵人",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "辛未", Gender: "MALE",
		Asserts: []shenShaAssert{{"hour", "天乙贵人"}}, NegAsserts: []shenShaAssert{{"hour", "时贵"}}, Source: "三命通会PDF第97-98页"},
	{ID: "SSF-003", Desc: "辛日天乙贵人寅午—时支寅→时柱有天乙贵人",
		YearP: "丙申", MonP: "甲午", DayP: "辛酉", HouP: "庚寅", Gender: "FEMALE",
		Asserts: []shenShaAssert{{"hour", "天乙贵人"}}, NegAsserts: []shenShaAssert{{"hour", "时贵"}}, Source: "三命通会PDF第97-98页"},
	{ID: "SSF-004", Desc: "壬日天乙贵人卯巳→四柱无卯巳→全无",
		YearP: "壬辰", MonP: "壬寅", DayP: "壬戌", HouP: "庚子", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "天乙贵人"}, {"year", "天乙贵人"}, {"month", "天乙贵人"}, {"hour", "天乙贵人"}},
		Source:     "三命通会PDF第97-98页"},

	// ========== 太极贵人 ==========
	{ID: "SSF-010", Desc: "甲年太极贵人在子午—日支午→日柱有太极贵人",
		YearP: "甲辰", MonP: "壬寅", DayP: "甲午", HouP: "庚辰", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "太极贵人"}}, Source: "渊海子平PDF第67页"},

	// ========== 禄神 ==========
	{ID: "SSF-020", Desc: "甲禄在寅—甲寅日支寅→日柱有禄神",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "禄神"}}, Source: "渊海子平"},
	{ID: "SSF-021", Desc: "庚禄在申—年支申→年柱有禄神",
		YearP: "庚申", MonP: "乙酉", DayP: "庚戌", HouP: "庚辰", Gender: "MALE",
		Asserts: []shenShaAssert{{"year", "禄神"}}, Source: "渊海子平"},
	{ID: "SSF-022", Desc: "癸禄在子—月柱子与时柱子均输出禄神",
		YearP: "癸酉", MonP: "甲子", DayP: "癸亥", HouP: "壬子", Gender: "MALE",
		Asserts: []shenShaAssert{{"month", "禄神"}, {"hour", "禄神"}}, Source: "渊海子平"},
	{ID: "SSF-023", Desc: "丙禄在巳—丙午日午非巳→日无;时支巳→时柱有",
		YearP: "甲辰", MonP: "戊辰", DayP: "丙午", HouP: "癸巳", Gender: "FEMALE",
		Asserts:    []shenShaAssert{{"hour", "禄神"}},
		NegAsserts: []shenShaAssert{{"day", "禄神"}}, Source: "渊海子平"},

	// ========== 羊刃(阳干) ==========
	{ID: "SSF-030", Desc: "甲刃在卯—年支卯→年柱有羊刃",
		YearP: "癸卯", MonP: "乙丑", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts:    []shenShaAssert{{"year", "羊刃"}},
		NegAsserts: []shenShaAssert{{"day", "羊刃"}}, Source: "渊海子平"},
	{ID: "SSF-031", Desc: "壬刃在子—时柱子输出羊刃",
		YearP: "己酉", MonP: "乙丑", DayP: "壬戌", HouP: "庚子", Gender: "MALE",
		Asserts:    []shenShaAssert{{"hour", "羊刃"}},
		NegAsserts: []shenShaAssert{{"day", "羊刃"}}, Source: "渊海子平"},
	{ID: "SSF-032", Desc: "丙刃在午—丙午日支午→日柱有羊刃",
		YearP: "丁未", MonP: "乙未", DayP: "丙午", HouP: "丁未", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "羊刃"}}, Source: "渊海子平"},

	// ========== 金舆 ==========
	{ID: "SSF-040", Desc: "甲日金舆在辰—日支寅非辰→日无",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "金舆"}}, Source: "三命通会PDF第91页、渊海子平PDF第93页"},
	{ID: "SSF-041", Desc: "乙日金舆在巳—乙巳日→日柱有金舆(日干乙,日支巳)",
		YearP: "癸酉", MonP: "甲子", DayP: "乙巳", HouP: "辛酉", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "金舆"}}, Source: "三命通会PDF第91页、渊海子平PDF第93页"},

	// ========== 文昌贵人（争议表停用） ==========
	{ID: "SSF-050", Desc: "旧甲日文昌在巳口径已停用",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "文昌贵人"}}, Source: "三命通会PDF第105页负向审计"},
	{ID: "SSF-051", Desc: "旧乙日文昌在午口径已停用",
		YearP: "乙巳", MonP: "丙戌", DayP: "乙酉", HouP: "甲午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"hour", "文昌贵人"}}, Source: "三命通会PDF第105页负向审计"},

	// ========== 福星贵人（争议简表停用） ==========
	{ID: "SSF-060", Desc: "旧甲日福星在寅单支口径已停用",
		YearP: "壬辰", MonP: "丙寅", DayP: "甲午", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"month", "福星贵人"}}, Source: "三命通会PDF第227页、渊海子平PDF第77页负向审计"},

	// ========== 魁罡 ==========
	{ID: "SSF-070", Desc: "壬辰日→魁罡",
		YearP: "壬辰", MonP: "壬寅", DayP: "壬辰", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "魁罡"}}, Source: "渊海子平"},
	{ID: "SSF-071", Desc: "庚戌日→魁罡",
		YearP: "庚戌", MonP: "丙戌", DayP: "庚戌", HouP: "庚辰", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "魁罡"}}, Source: "渊海子平"},
	{ID: "SSF-072", Desc: "戊戌日→魁罡",
		YearP: "庚戌", MonP: "丙戌", DayP: "戊戌", HouP: "庚辰", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "魁罡"}}, Source: "渊海子平"},

	// ========== 华盖 ==========
	{ID: "SSF-080", Desc: "寅午戌华盖在戌—年支戌→年柱有华盖",
		YearP: "庚戌", MonP: "丙戌", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts:    []shenShaAssert{{"year", "华盖"}, {"month", "华盖"}},
		NegAsserts: []shenShaAssert{{"day", "华盖"}}, Source: "渊海子平"},
	{ID: "SSF-081", Desc: "申子辰华盖在辰—年支辰→年柱有华盖",
		YearP: "壬辰", MonP: "壬寅", DayP: "壬戌", HouP: "庚子", Gender: "MALE",
		Asserts: []shenShaAssert{{"year", "华盖"}}, Source: "渊海子平"},

	// ========== 驿马 ==========
	{ID: "SSF-090", Desc: "申子辰驿马在寅—全局无寅→无驿马",
		YearP: "壬申", MonP: "庚子", DayP: "壬辰", HouP: "庚子", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "驿马"}, {"year", "驿马"}}, Source: "三命通会PDF第92页"},
	{ID: "SSF-091", Desc: "亥卯未驿马在巳—年支巳→年柱有驿马",
		YearP: "乙巳", MonP: "戊寅", DayP: "己亥", HouP: "丙寅", Gender: "FEMALE",
		Asserts: []shenShaAssert{{"year", "驿马"}}, Source: "三命通会PDF第92页"},

	// ========== 咸池 ==========
	{ID: "SSF-100", Desc: "亥卯未咸池在子—全局无子→无咸池",
		YearP: "乙亥", MonP: "戊寅", DayP: "己亥", HouP: "丙寅", Gender: "FEMALE",
		NegAsserts: []shenShaAssert{{"day", "咸池"}, {"year", "咸池"}}, Source: "三命通会PDF第81页"},
	{ID: "SSF-101", Desc: "巳酉丑咸池在午—年日支酉、月支午→月柱有咸池",
		YearP: "癸酉", MonP: "戊午", DayP: "癸酉", HouP: "辛酉", Gender: "MALE",
		Asserts: []shenShaAssert{{"month", "咸池"}}, Source: "三命通会PDF第81页"},

	// ========== 将星 ==========
	{ID: "SSF-110", Desc: "寅午戌将星在午—日支午→日柱有将星",
		YearP: "庚戌", MonP: "丙戌", DayP: "甲午", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "将星"}}, Source: "三命通会PDF第80页"},

	// ========== 劫煞 ==========
	{ID: "SSF-120", Desc: "亥卯未劫煞在申—全局无申→无劫煞",
		YearP: "乙亥", MonP: "戊寅", DayP: "己亥", HouP: "丙寅", Gender: "FEMALE",
		NegAsserts: []shenShaAssert{{"day", "劫煞"}}, Source: "渊海子平"},

	// ========== 灾煞 ==========
	{ID: "SSF-130", Desc: "申子辰灾煞在午—四柱无午→无灾煞",
		YearP: "壬申", MonP: "庚子", DayP: "壬辰", HouP: "庚子", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "灾煞"}}, Source: "渊海子平"},

	// ========== 亡神 ==========
	{ID: "SSF-140", Desc: "寅午戌亡神在巳—全局无巳→无亡神",
		YearP: "庚戌", MonP: "癸未", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"month", "亡神"}}, Source: "渊海子平"},

	// ========== 孤辰寡宿 ==========
	{ID: "SSF-150", Desc: "申酉戌孤辰在亥—日支亥→日柱有孤辰",
		YearP: "癸酉", MonP: "甲子", DayP: "癸亥", HouP: "辛酉", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "孤辰"}}, Source: "三命通会PDF第118页；渊海子平PDF第632、744页"},
	{ID: "SSF-151", Desc: "寅卯辰寡宿在丑—月支丑→月柱有寡宿",
		YearP: "甲寅", MonP: "乙丑", DayP: "甲子", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"month", "寡宿"}}, Source: "三命通会PDF第118页；渊海子平PDF第632、744页"},

	// ========== 天德/月德 ==========
	{ID: "SSF-160", Desc: "戌月天德在丙—丙戌月柱有天德，日干非丙则无月德",
		YearP: "庚戌", MonP: "丙戌", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"month", "天德"}},
		NegAsserts: []shenShaAssert{
			{"year", "月德"}, {"month", "月德"}, {"day", "月德"}, {"hour", "月德"},
		}, Source: "渊海子平"},
	{ID: "SSF-161", Desc: "寅月天德在丁—全局无丁→无天德",
		YearP: "乙丑", MonP: "戊寅", DayP: "甲子", HouP: "丙寅", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"month", "天德"}}, Source: "渊海子平"},

	// ========== 学堂 ==========
	{ID: "SSF-170", Desc: "生年纳音水见正学堂甲申时柱",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "甲申", Gender: "MALE",
		Asserts: []shenShaAssert{{"hour", "学堂"}}, Source: "三命通会PDF第105-106页"},

	// ========== 词馆 ==========
	{ID: "SSF-171", Desc: "生年纳音水见正词馆癸亥月柱",
		YearP: "壬辰", MonP: "癸亥", DayP: "甲午", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"month", "词馆"}}, Source: "三命通会PDF第105-106页"},

	// ========== 天厨贵人 ==========
	{ID: "SSF-172", Desc: "甲日天厨贵人在巳—全局无巳→无",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "天厨贵人"}}, Source: "渊海子平PDF第76页"},

	// ========== 国印贵人 ==========
	{ID: "SSF-173", Desc: "壬年国印在未—全局无未→无",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "国印贵人"}}, Source: "渊海子平PDF第733页"},

	// ========== 天官贵人 ==========
	{ID: "SSF-174", Desc: "壬年天官贵人在寅—月柱寅命中",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲午", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"month", "天官贵人"}}, Source: "渊海子平PDF第65页"},

	// ========== 未定位红鸾天喜年支表 ==========
	{ID: "SSF-180", Desc: "旧年支酉红鸾在午表无固定来源→不发布",
		YearP: "癸酉", MonP: "甲子", DayP: "癸亥", HouP: "戊午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"hour", "红鸾"}}, Source: "固定五份本地PDF负向检索"},
	{ID: "SSF-181", Desc: "旧年支酉天喜在子表无固定来源→不发布",
		YearP: "癸酉", MonP: "甲子", DayP: "癸亥", HouP: "辛酉", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"month", "天喜"}}, Source: "固定五份本地PDF负向检索"},

	// ========== 空亡 ==========
	{ID: "SSF-190", Desc: "甲子旬空戌亥—年支戌→年柱有空亡",
		YearP: "甲戌", MonP: "丙子", DayP: "甲子", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"year", "空亡"}}, Source: "渊海子平"},
	{ID: "SSF-191", Desc: "甲寅旬空子丑—年支子→年柱有空亡",
		YearP: "甲子", MonP: "丙寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"year", "空亡"}}, Source: "渊海子平"},

	// ========== 天赦 ==========
	{ID: "SSF-200", Desc: "春戊寅日→天赦（寅月戊寅日）",
		YearP: "甲辰", MonP: "丙寅", DayP: "戊寅", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"global", "天赦"}}, Source: "渊海子平"},

	// ========== 四废 ==========
	{ID: "SSF-210", Desc: "春庚申日→四废（春季庚申日）",
		YearP: "甲辰", MonP: "丁卯", DayP: "庚申", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"global", "四废"}}, Source: "渊海子平"},

	// ========== 三奇 ==========
	{ID: "SSF-220", Desc: "甲戊庚相邻顺布→三奇",
		YearP: "甲辰", MonP: "戊辰", DayP: "庚申", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"global", "三奇"}}, Source: "三命通会"},
	{ID: "SSF-221", Desc: "乙丙丁相邻顺布→三奇",
		YearP: "乙巳", MonP: "丙戌", DayP: "丁亥", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"global", "三奇"}}, Source: "三命通会"},

	// ========== 天罗地网 ==========
	{ID: "SSF-230", Desc: "年纳音火(丙寅炉中火)+日时戌亥成对→天罗",
		YearP: "丙寅", MonP: "甲午", DayP: "甲戌", HouP: "乙亥", Gender: "MALE",
		Asserts: []shenShaAssert{{"global", "天罗"}}, Source: "渊海子平"},
	{ID: "SSF-231", Desc: "年纳音水土+日时辰巳成对→地网",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲辰", HouP: "己巳", Gender: "MALE",
		Asserts: []shenShaAssert{{"global", "地网"}}, Source: "渊海子平"},

	// ========== 无来源拱命快捷规则反例 ==========
	{ID: "SSF-240", Desc: "四柱仅有辰寅不构成龙虎拱命",
		YearP: "甲辰", MonP: "丙寅", DayP: "戊子", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"global", "龙虎拱命"}}, Source: "无可定位来源，旧快捷规则已删除"},

	{ID: "SSF-241", Desc: "四柱仅有酉巳不构成凤凰拱命",
		YearP: "乙酉", MonP: "辛巳", DayP: "甲子", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"global", "凤凰拱命"}}, Source: "无可定位来源，旧快捷规则已删除"},

	// ========== 红艳煞 ==========
	{ID: "SSF-250", Desc: "甲日红艳在午—日支午→日柱有红艳煞",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲午", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "红艳煞"}}, Source: "三命通会PDF第125页"},

	// ========== 阴差阳错 ==========
	{ID: "SSF-260", Desc: "丙午日→阴差阳错",
		YearP: "丁未", MonP: "乙未", DayP: "丙午", HouP: "丁未", Gender: "FEMALE",
		Asserts: []shenShaAssert{{"day", "阴差阳错"}}, Source: "渊海子平"},

	// ========== 孤鸾煞 ==========
	{ID: "SSF-270", Desc: "甲寅日→孤鸾煞",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "孤鸾煞"}}, Source: "渊海子平"},

	// ========== 十恶大败 ==========
	{ID: "SSF-280", Desc: "甲辰日→十恶大败",
		YearP: "甲辰", MonP: "丙寅", DayP: "甲辰", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "十恶大败"}}, Source: "渊海子平"},

	// ========== 未定位来源的十灵日反例 ==========
	{ID: "SSF-290", Desc: "甲辰日不发布无来源十灵日",
		YearP: "甲辰", MonP: "丙寅", DayP: "甲辰", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "十灵日"}}, Source: "五份本地资料未定位"},

	// ========== 日德 ==========
	{ID: "SSF-300", Desc: "甲寅日→日德",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "日德"}}, Source: "渊海子平"},

	// ========== 八专 ==========
	{ID: "SSF-310", Desc: "甲寅日→八专",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "八专"}}, Source: "渊海子平"},

	// ========== 金神 ==========
	{ID: "SSF-320", Desc: "癸酉时→金神",
		YearP: "癸酉", MonP: "甲子", DayP: "甲子", HouP: "癸酉", Gender: "MALE",
		Asserts: []shenShaAssert{{"hour", "金神"}}, Source: "渊海子平PDF第221页"},

	// ========== 未定位来源的六秀日反例 ==========
	{ID: "SSF-330", Desc: "丙午日不发布无来源六秀日",
		YearP: "丁未", MonP: "乙未", DayP: "丙午", HouP: "丁未", Gender: "FEMALE",
		NegAsserts: []shenShaAssert{{"day", "六秀日"}}, Source: "五份本地资料未定位"},

	// ========== 福德秀气 ==========
	{ID: "SSF-340", Desc: "乙酉日→福德秀气",
		YearP: "乙酉", MonP: "丙戌", DayP: "乙酉", HouP: "庚辰", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "福德秀气"}}, Source: "渊海子平"},

	// ========== 元辰(性别相关) ==========
	{ID: "SSF-350", Desc: "阴男阳女取六冲支逆行一辰→年支酉冲卯取寅→年柱无元辰",
		YearP: "癸酉", MonP: "甲子", DayP: "癸亥", HouP: "辛酉", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"year", "元辰"}}, Source: "三命通会PDF第114页"},

	// ========== 天医（月支前一位表无固定来源，失败关闭） ==========
	{ID: "SSF-360", Desc: "旧寅月丑支表不得冒名发布天医",
		YearP: "乙丑", MonP: "戊寅", DayP: "甲子", HouP: "丙寅", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"year", "天医"}}, Source: "固定资料只定位不同择方/合婚或年神概念"},

	// ========== 流霞（固定资料未定位，失败关闭） ==========
	{ID: "SSF-370", Desc: "旧甲年酉支流霞表无固定可审计来源→不发布",
		YearP: "甲申", MonP: "丙寅", DayP: "甲子", HouP: "癸酉", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"hour", "流霞"}}, Source: "五份固定本地资料未定位"},

	// ========== 未裁决隔角煞 ==========
	{ID: "SSF-380", Desc: "旧年支酉隔角在戌表与古籍日时/孤寡语境不符→不发布",
		YearP: "癸酉", MonP: "甲子", DayP: "癸亥", HouP: "辛酉", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"year", "隔角煞"}}, Source: "三命通会PDF第118页；渊海子平PDF第635、730页异规则"},

	// ========== 岁驾（旧年干禄位别名无来源，失败关闭） ==========
	{ID: "SSF-390", Desc: "旧甲年禄支寅不得冒名发布岁驾",
		YearP: "甲辰", MonP: "丙寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "岁驾"}}, Source: "渊海子平PDF第672页为太岁禄马条件，非当前年干禄位别名"},

	// ========== 童子煞（五份固定资料未定位，失败关闭） ==========
	{ID: "SSF-400", Desc: "旧巳月辰时童子煞表无固定可审计来源→不发布",
		YearP: "乙巳", MonP: "辛巳", DayP: "甲子", HouP: "戊辰", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"hour", "童子煞"}}, Source: "五份固定本地资料未定位"},
}

func TestShenShaFullCoverage(t *testing.T) {
	svc := &BaziService{}
	var pass, fail int
	var fails []string
	var outputs []string

	// 收集所有出现的神煞名称（去重）
	allShenShaNames := make(map[string]bool)
	for _, tc := range shenShaFullCases {
		for _, a := range tc.Asserts {
			allShenShaNames[a.God] = true
		}
		for _, a := range tc.NegAsserts {
			allShenShaNames[a.God] = true
		}
	}

	for _, tc := range shenShaFullCases {
		r, err := svc.CalculateSyntheticPillars(tc.YearP, tc.MonP, tc.DayP, tc.HouP, tc.Gender)
		if err != nil {
			fail++
			fails = append(fails, fmt.Sprintf("[%s] 计算失败: %v", tc.ID, err))
			continue
		}

		// 构建柱位→神煞列表映射
		items := map[string][]string{}
		for _, pss := range r.ShenShaByPillar {
			items[pss.Pillar] = uniqueStrings(pss.Items)
		}
		items["global"] = uniqueStrings(r.GlobalShenSha)

		// 记录实际输出
		outputs = append(outputs, fmt.Sprintf("[%s] %s: 日%v 年%v 月%v 时%v 全局%v",
			tc.ID, tc.Desc,
			items["day"], items["year"], items["month"], items["hour"], items["global"]))

		// 验证正向断言
		for _, a := range tc.Asserts {
			list := items[a.Pillar]
			found := false
			for _, item := range list {
				if strings.Contains(item, a.God) {
					found = true
					break
				}
			}
			if found {
				pass++
			} else {
				fail++
				fails = append(fails, fmt.Sprintf("[%s] %s柱应含'%s',实际%v", tc.ID, a.Pillar, a.God, list))
			}
		}

		// 验证反向断言
		for _, a := range tc.NegAsserts {
			list := items[a.Pillar]
			found := false
			for _, item := range list {
				if strings.Contains(item, a.God) {
					found = true
					break
				}
			}
			if !found {
				pass++
			} else {
				fail++
				fails = append(fails, fmt.Sprintf("[%s] %s柱不应含'%s',实际有%v", tc.ID, a.Pillar, a.God, list))
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║          神煞全量测试                                         ║\n")
	sb.WriteString("║ 对照来源: 《渊海子平》《三命通会》                              ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════════╝\n\n")

	total := pass + fail
	sb.WriteString(fmt.Sprintf("规则断言: %d通过 + %d失败 = %d总\n", pass, fail, total))
	sb.WriteString(fmt.Sprintf("用例数: %d | 覆盖神煞类型: %d\n\n", len(shenShaFullCases), len(allShenShaNames)))

	if len(fails) > 0 {
		sb.WriteString("失败:\n" + strings.Join(fails, "\n") + "\n")
	} else {
		sb.WriteString("全部通过！\n")
	}
	sb.WriteString("\n详情:\n" + strings.Join(outputs, "\n"))

	t.Log("\n" + sb.String())

	if fail > 0 {
		t.Errorf("神煞全量测试有 %d 个断言失败", fail)
	}
}
