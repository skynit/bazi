package bazi

import (
	"fmt"
	"strings"

	"bazi/internal/model"
	"bazi/internal/service/data"
)

type ShenShaPillars struct {
	Year   model.Pillar
	Month  model.Pillar
	Day    model.Pillar
	Hour   model.Pillar
	Gender string // "MALE" or "FEMALE"
}

type ShenShaCalcResult struct {
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

type nayinSchoolHallRule struct {
	XueTang string
	CiGuan  string
}

var nayinSchoolHallRules = map[string]nayinSchoolHallRule{
	"金": {XueTang: "辛巳", CiGuan: "壬申"},
	"木": {XueTang: "己亥", CiGuan: "庚寅"},
	"水": {XueTang: "甲申", CiGuan: "癸亥"},
	"火": {XueTang: "丙寅", CiGuan: "乙巳"},
	"土": {XueTang: "戊申", CiGuan: "丁亥"},
}

var yearGanShenShaRules = map[string][]shenShaRule{
	"甲": {{"太极贵人", "子午"}, {"国印贵人", "戌"}, {"天官贵人", "未"}},
	"乙": {{"太极贵人", "子午"}, {"国印贵人", "亥"}, {"天官贵人", "辰"}},
	"丙": {{"太极贵人", "卯酉"}, {"国印贵人", "丑"}, {"天官贵人", "巳"}},
	"丁": {{"太极贵人", "卯酉"}, {"国印贵人", "寅"}, {"天官贵人", "酉"}},
	"戊": {{"太极贵人", "辰戌丑未"}, {"国印贵人", "丑"}, {"天官贵人", "戌"}},
	"己": {{"太极贵人", "辰戌丑未"}, {"国印贵人", "寅"}, {"天官贵人", "卯"}},
	"庚": {{"太极贵人", "寅亥"}, {"国印贵人", "辰"}, {"天官贵人", "亥"}},
	"辛": {{"太极贵人", "寅亥"}, {"国印贵人", "巳"}, {"天官贵人", "申"}},
	"壬": {{"太极贵人", "巳申"}, {"国印贵人", "未"}, {"天官贵人", "寅"}},
	"癸": {{"太极贵人", "巳申"}, {"国印贵人", "申"}, {"天官贵人", "午"}},
}

// 年支神煞：丧门(i+2) 吊客(i+10) 病符(i+11) 死符(i+5) 官符(i+4)
// 破宅煞=命后一辰(i+1)
// 参考：《三命通会》《渊海子平》已定位原文。
var yearZhiShenShaRules = map[string][]shenShaRule{
	"子": {{"丧门", "寅"}, {"吊客", "戌"}, {"病符", "亥"}, {"死符", "巳"}, {"官符", "辰"}, {"破宅煞", "丑"}},
	"丑": {{"丧门", "卯"}, {"吊客", "亥"}, {"病符", "子"}, {"死符", "午"}, {"官符", "巳"}, {"破宅煞", "寅"}},
	"寅": {{"丧门", "辰"}, {"吊客", "子"}, {"病符", "丑"}, {"死符", "未"}, {"官符", "午"}, {"破宅煞", "卯"}},
	"卯": {{"丧门", "巳"}, {"吊客", "丑"}, {"病符", "寅"}, {"死符", "申"}, {"官符", "未"}, {"破宅煞", "辰"}},
	"辰": {{"丧门", "午"}, {"吊客", "寅"}, {"病符", "卯"}, {"死符", "酉"}, {"官符", "申"}, {"破宅煞", "巳"}},
	"巳": {{"丧门", "未"}, {"吊客", "卯"}, {"病符", "辰"}, {"死符", "戌"}, {"官符", "酉"}, {"破宅煞", "午"}},
	"午": {{"丧门", "申"}, {"吊客", "辰"}, {"病符", "巳"}, {"死符", "亥"}, {"官符", "戌"}, {"破宅煞", "未"}},
	"未": {{"丧门", "酉"}, {"吊客", "巳"}, {"病符", "午"}, {"死符", "子"}, {"官符", "亥"}, {"破宅煞", "申"}},
	"申": {{"丧门", "戌"}, {"吊客", "午"}, {"病符", "未"}, {"死符", "丑"}, {"官符", "子"}, {"破宅煞", "酉"}},
	"酉": {{"丧门", "亥"}, {"吊客", "未"}, {"病符", "申"}, {"死符", "寅"}, {"官符", "丑"}, {"破宅煞", "戌"}},
	"戌": {{"丧门", "子"}, {"吊客", "申"}, {"病符", "酉"}, {"死符", "卯"}, {"官符", "寅"}, {"破宅煞", "亥"}},
	"亥": {{"丧门", "丑"}, {"吊客", "酉"}, {"病符", "戌"}, {"死符", "辰"}, {"官符", "卯"}, {"破宅煞", "子"}},
}

// 月支神煞（按《三命通会》起法）
// 月厌：正月(寅)起戌，逐月逆行；月煞按四组三合月取固定支。
// 天赦已由 isTianShe() 统一按季节判断，不入此表
var monthZhiShenShaRules = map[string][]shenShaRule{
	"子": {{"月煞", "未"}, {"月厌", "子"}},
	"丑": {{"月煞", "辰"}, {"月厌", "亥"}},
	"寅": {{"月煞", "丑"}, {"月厌", "戌"}},
	"卯": {{"月煞", "戌"}, {"月厌", "酉"}},
	"辰": {{"月煞", "未"}, {"月厌", "申"}},
	"巳": {{"月煞", "辰"}, {"月厌", "未"}},
	"午": {{"月煞", "丑"}, {"月厌", "午"}},
	"未": {{"月煞", "戌"}, {"月厌", "巳"}},
	"申": {{"月煞", "未"}, {"月厌", "辰"}},
	"酉": {{"月煞", "辰"}, {"月厌", "卯"}},
	"戌": {{"月煞", "丑"}, {"月厌", "寅"}},
	"亥": {{"月煞", "戌"}, {"月厌", "丑"}},
}

// 天厨贵人按《渊海子平》十干查支表；当前 Profile 只取日干为主键。
// 天乙贵人当前 Profile 只以日干为主键；昼夜贵和互换贵另行裁决。
var dayGanShenShaRules = map[string][]shenShaRule{
	"甲": {{"羊刃", "卯"}, {"金舆", "辰"}, {"红艳煞", "午"}, {"天厨贵人", "巳"}, {"天乙贵人", "丑未"}},
	// 当前子平 Profile 采用五阳干有刃、五阴干无刃的口径。
	"乙": {{"金舆", "巳"}, {"红艳煞", "午"}, {"天厨贵人", "午"}, {"天乙贵人", "子申"}},
	"丙": {{"羊刃", "午"}, {"金舆", "未"}, {"红艳煞", "寅"}, {"天厨贵人", "巳"}, {"天乙贵人", "亥酉"}},
	"丁": {{"金舆", "申"}, {"红艳煞", "未"}, {"天厨贵人", "午"}, {"天乙贵人", "亥酉"}},
	"戊": {{"羊刃", "午"}, {"金舆", "未"}, {"红艳煞", "子"}, {"天厨贵人", "申"}, {"天乙贵人", "丑未"}},
	"己": {{"金舆", "申"}, {"红艳煞", "辰"}, {"天厨贵人", "酉"}, {"天乙贵人", "子申"}},
	"庚": {{"羊刃", "酉"}, {"金舆", "戌"}, {"红艳煞", "戌"}, {"天厨贵人", "亥"}, {"天乙贵人", "丑未"}},
	"辛": {{"金舆", "亥"}, {"红艳煞", "酉"}, {"天厨贵人", "子"}, {"天乙贵人", "寅午"}},
	"壬": {{"羊刃", "子"}, {"金舆", "丑"}, {"红艳煞", "巳"}, {"天厨贵人", "寅"}, {"天乙贵人", "卯巳"}},
	"癸": {{"金舆", "寅"}, {"红艳煞", "申"}, {"天厨贵人", "卯"}, {"天乙贵人", "卯巳"}},
}

var sanHeShenShaRules = map[string]sanHeShenSha{
	"寅": {"午", "戌", "申", "亥", "子", "巳", "卯"}, "午": {"午", "戌", "申", "亥", "子", "巳", "卯"}, "戌": {"午", "戌", "申", "亥", "子", "巳", "卯"},
	"巳": {"酉", "丑", "亥", "寅", "卯", "申", "午"}, "酉": {"酉", "丑", "亥", "寅", "卯", "申", "午"}, "丑": {"酉", "丑", "亥", "寅", "卯", "申", "午"},
	"申": {"子", "辰", "寅", "巳", "午", "亥", "酉"}, "子": {"子", "辰", "寅", "巳", "午", "亥", "酉"}, "辰": {"子", "辰", "寅", "巳", "午", "亥", "酉"},
	"亥": {"卯", "未", "巳", "申", "酉", "寅", "子"}, "卯": {"卯", "未", "巳", "申", "酉", "寅", "子"}, "未": {"卯", "未", "巳", "申", "酉", "寅", "子"},
}

var specialDayShenShaRules = map[string][]string{
	"戊子": {"九丑日"}, "戊午": {"九丑日", "孤鸾煞"},
	"壬子": {"九丑日", "孤鸾煞"}, "壬午": {"九丑日"}, "乙卯": {"九丑日", "八专"}, "乙酉": {"福德秀气"},
	"辛卯": {"九丑日", "阴差阳错"}, "辛酉": {"九丑日", "阴差阳错", "八专", "福德秀气"}, "己卯": {"九丑日"}, "己酉": {"九丑日", "福德秀气"},
	"甲辰": {"十恶大败"}, "乙巳": {"十恶大败", "孤鸾煞", "福德秀气"}, "丙申": {"十恶大败"}, "丁亥": {"十恶大败"},
	"戊戌": {"十恶大败", "魁罡", "八专"}, "庚辰": {"十恶大败", "魁罡", "日德"}, "辛巳": {"十恶大败", "福德秀气"}, "壬申": {"十恶大败"}, "癸亥": {"十恶大败", "阴差阳错"},
	"壬辰": {"魁罡", "阴差阳错"}, "庚戌": {"魁罡"}, "甲寅": {"孤鸾煞", "八专", "日德"}, "丙午": {"阴差阳错", "孤鸾煞"},
	"丁巳": {"孤鸾煞", "福德秀气"}, "辛亥": {"孤鸾煞"}, "戊申": {"阴差阳错", "孤鸾煞"}, "丙子": {"阴差阳错"},
	"丁丑": {"阴差阳错", "福德秀气"}, "戊寅": {"阴差阳错"}, "癸巳": {"阴差阳错", "福德秀气"}, "丁未": {"阴差阳错", "八专"}, "壬戌": {"阴差阳错", "日德"},
	"己未": {"八专"}, "庚申": {"八专"}, "癸丑": {"八专", "福德秀气"}, "乙丑": {"十恶大败", "福德秀气"},
	"丙辰": {"日德"}, "戊辰": {"日德"},
	"丁酉": {"福德秀气"}, "己巳": {"福德秀气"}, "己丑": {"福德秀气"}, "辛丑": {"福德秀气"}, "癸酉": {"福德秀气"},
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

// --- 三合局追加：六厄（死位） ---
var sanHeLiuE = map[string]string{
	"寅": "酉", "午": "酉", "戌": "酉",
	"巳": "子", "酉": "子", "丑": "子",
	"申": "卯", "子": "卯", "辰": "卯",
	"亥": "午", "卯": "午", "未": "午",
}

// 《三命通会》天刑煞：只以生年支查时柱天干。
var tianXingHourGanByYearZhi = map[string]string{
	"子": "乙", "丑": "乙", "寅": "庚", "卯": "辛", "辰": "辛", "巳": "壬",
	"午": "癸", "未": "癸", "申": "丙", "酉": "丁", "戌": "丁", "亥": "戊",
}

// 《三命通会》雷霆煞：正七月子、二八月寅、三九月辰、
// 四十月午、五十一月申、六十二月戌。
var leiTingShaByMonthZhi = map[string]string{
	"寅": "子", "申": "子", "卯": "寅", "酉": "寅", "辰": "辰", "戌": "辰",
	"巳": "午", "亥": "午", "午": "申", "子": "申", "未": "戌", "丑": "戌",
}

// --- 孤辰寡宿：按年支/日支所在三会方查找（非三合局） ---
// 经典依据：渊海子平/三命通会，孤辰寡宿取法：
// 孤辰：亥子丑→寅，寅卯辰→巳，巳午未→申，申酉戌→亥
// 寡宿：亥子丑→戌，寅卯辰→丑，巳午未→辰，申酉戌→未
var guChenBySanHui = map[string]string{
	"亥": "寅", "子": "寅", "丑": "寅",
	"寅": "巳", "卯": "巳", "辰": "巳",
	"巳": "申", "午": "申", "未": "申",
	"申": "亥", "酉": "亥", "戌": "亥",
}
var guSuBySanHui = map[string]string{
	"亥": "戌", "子": "戌", "丑": "戌",
	"寅": "丑", "卯": "丑", "辰": "丑",
	"巳": "辰", "午": "辰", "未": "辰",
	"申": "未", "酉": "未", "戌": "未",
}

// --- 飞刃：羊刃对冲（阳干有飞刃，阴干无） ---
var feiRenByGan = map[string]string{
	"甲": "酉", // 羊刃卯→冲酉
	"丙": "子", // 羊刃午→冲子
	"戊": "子", // 羊刃午→冲子
	"庚": "卯", // 羊刃酉→冲卯
	"壬": "午", // 羊刃子→冲午
}

var zhiLiuHe = map[string]string{
	"子": "丑", "丑": "子", "寅": "亥", "卯": "戌", "辰": "酉", "巳": "申",
	"午": "未", "未": "午", "申": "巳", "酉": "辰", "戌": "卯", "亥": "寅",
}
var zhiLiuChong = map[string]string{
	"子": "午", "丑": "未", "寅": "申", "卯": "酉", "辰": "戌", "巳": "亥",
	"午": "子", "未": "丑", "申": "寅", "酉": "卯", "戌": "辰", "亥": "巳",
}

func CalcShenShaByPillars(p ShenShaPillars) (ShenShaCalcResult, error) {
	if err := validateShenShaPillars(p); err != nil {
		return ShenShaCalcResult{}, err
	}
	var res ShenShaCalcResult
	allPillars := []model.Pillar{p.Year, p.Month, p.Day, p.Hour}
	allBranches := []string{p.Year.Zhi, p.Month.Zhi, p.Day.Zhi, p.Hour.Zhi}

	// --- 年干/年支/三合局规则：分柱匹配写入 ---
	addRulesToBucketByExactPillar(&res, yearGanShenShaRules[p.Year.Gan], allPillars)
	addRulesToBucketByExactPillar(&res, yearZhiShenShaRules[p.Year.Zhi], allPillars)
	addSanHeRulesToBucketByExactPillar(&res, sanHeShenShaRules[p.Year.Zhi], allPillars)
	addYearZhiExtra(p, allBranches, &res)
	addTianXingSha(p, &res)

	// --- 月柱 ---
	addMonthGanRules(p, &res)
	addRulesToBucketByExactPillar(&res, monthZhiShenShaRules[p.Month.Zhi], allPillars)

	// --- 日柱 ---
	if luBranch, ok := luBranchForStem(p.Day.Gan); ok {
		addRulesToBucketByExactPillar(&res, []shenShaRule{{Name: "禄神", Target: luBranch}}, allPillars)
	}
	addRulesToBucketByExactPillar(&res, dayGanShenShaRules[p.Day.Gan], allPillars)
	addNayinSchoolHall(p, allPillars, &res)
	addSanHeRulesToBucketByExactPillar(&res, sanHeShenShaRules[p.Day.Zhi], allPillars)
	for _, name := range specialDayShenShaRules[p.Day.Gan+p.Day.Zhi] {
		appendShenSha(&res.Day, name, "")
	}
	addDayExtra(p, allBranches, &res)
	addJiaoLu(p, &res)

	// --- 孤辰寡宿（只以生年支三会方为主键） ---
	if guChenZhi := guChenBySanHui[p.Year.Zhi]; guChenZhi != "" {
		appendShenShaByTargetBranch(&res, allBranches, "孤辰", guChenZhi)
	}
	if guSuZhi := guSuBySanHui[p.Year.Zhi]; guSuZhi != "" {
		appendShenShaByTargetBranch(&res, allBranches, "寡宿", guSuZhi)
	}

	// --- 空亡 ---
	addKongWang(p, &res)

	// --- 时柱 ---
	addHourShenSha(p, &res)

	// --- 全局 ---
	addGlobalShenSha(p, &res)

	deduplicateShenShaHits(&res)
	return res, nil
}

func addNayinSchoolHall(p ShenShaPillars, pillars []model.Pillar, res *ShenShaCalcResult) {
	_, yearElement, valid := naYinNameAndElement(p.Year.Gan, p.Year.Zhi)
	if !valid {
		return
	}
	rule, ok := nayinSchoolHallRules[yearElement]
	if !ok {
		return
	}
	addRulesToBucketByExactPillar(res, []shenShaRule{
		{Name: "学堂", Target: rule.XueTang},
		{Name: "词馆", Target: rule.CiGuan},
	}, pillars)
}

func validateShenShaPillars(p ShenShaPillars) error {
	if p.Gender != model.GenderMale && p.Gender != model.GenderFemale {
		return fmt.Errorf("invalid shen-sha gender %q: must be MALE or FEMALE", p.Gender)
	}
	for _, item := range []struct {
		name   string
		pillar model.Pillar
	}{
		{"year", p.Year}, {"month", p.Month}, {"day", p.Day}, {"hour", p.Hour},
	} {
		if sixtyCycleIndex(item.pillar.Gan, item.pillar.Zhi) < 0 {
			return fmt.Errorf("invalid shen-sha %s pillar %q: must be one of the sixty-cycle pairs", item.name, item.pillar.Gan+item.pillar.Zhi)
		}
	}
	return nil
}

func addRulesToBucketByExactPillar(res *ShenShaCalcResult, rules []shenShaRule, pillars []model.Pillar) {
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

func appendShenShaByTargetBranch(res *ShenShaCalcResult, branches []string, name, target string) {
	if target == "" {
		return
	}
	for i, branch := range branches {
		if branch != target {
			continue
		}
		appendShenShaToPillarIndex(res, i, name, target)
	}
}

func appendShenShaToPillarIndex(res *ShenShaCalcResult, index int, name, target string) {
	switch index {
	case 0:
		appendShenSha(&res.Year, name, target)
	case 1:
		appendShenSha(&res.Month, name, target)
	case 2:
		appendShenSha(&res.Day, name, target)
	case 3:
		appendShenSha(&res.Hour, name, target)
	}
}

func addSanHeRulesToBucketByExactPillar(res *ShenShaCalcResult, rule sanHeShenSha, pillars []model.Pillar) {
	if rule.Jiang == "" {
		return
	}
	rules := []shenShaRule{
		{"将星", rule.Jiang}, {"华盖", rule.HuaGai}, {"驿马", rule.YiMa}, {"灾煞", rule.ZaiSha},
		{"劫煞", rule.JieSha}, {"亡神", rule.WangShen}, {"咸池", rule.XianChi},
	}
	addRulesToBucketByExactPillar(res, rules, pillars)
}

func addMonthGanRules(p ShenShaPillars, res *ShenShaCalcResult) {
	monthIndex := data.ZhiIndex(p.Month.Zhi)
	tianDe := data.MonthShenMap[data.TianDe][monthIndex]
	yueDe := data.MonthShenMap[data.YueDe][monthIndex]
	pillars := []model.Pillar{p.Year, p.Month, p.Day, p.Hour}
	addSingleGanZhiTargetToPillars(res, pillars, "天德贵人", tianDe)
	addSingleGanZhiTargetToPillars(res, pillars, "天德合", monthDeHeTarget(tianDe))

	// 《渊海子平》第71页限定月德贵人“亦须在日上见之”。当前
	// Profile 对第73页月德合沿用同一日柱限定，不扩展到年、月、时干。
	if p.Day.Gan == yueDe {
		appendShenSha(&res.Day, "月德贵人", yueDe)
	}
	if yueDeHe := ganHe[yueDe]; p.Day.Gan == yueDeHe {
		appendShenSha(&res.Day, "月德合", yueDeHe)
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
		if role := deXiuRole(p.Month.Zhi, pair.gan); role != "" {
			appendShenSha(pair.pillar, "德秀贵人", role+"/"+pair.gan)
		}
	}
}

func monthDeHeTarget(target string) string {
	if data.GanIndex(target) >= 0 {
		return ganHe[target]
	}
	if data.ZhiIndex(target) >= 0 {
		return zhiLiuHe[target]
	}
	return ""
}

func addSingleGanZhiTargetToPillars(res *ShenShaCalcResult, pillars []model.Pillar, name, target string) {
	for i, pillar := range pillars {
		if exactSingleGanZhiTargetMatchesPillar(target, pillar) {
			appendShenShaToPillarIndex(res, i, name, target)
		}
	}
}

func exactSingleGanZhiTargetMatchesPillar(target string, pillar model.Pillar) bool {
	if data.GanIndex(target) >= 0 {
		return pillar.Gan == target
	}
	if data.ZhiIndex(target) >= 0 {
		return pillar.Zhi == target
	}
	return false
}

func addKongWang(p ShenShaPillars, res *ShenShaCalcResult) {
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

func addHourShenSha(p ShenShaPillars, res *ShenShaCalcResult) {
	// 《三命通会》八专条分别以“日上”“时上”论命中；其余固定日表仍只取日柱。
	if specialDayRuleContains(p.Hour.Gan+p.Hour.Zhi, "八专") {
		appendShenSha(&res.Hour, "八专", "")
	}
	if isJinShenHourPillar(p.Hour) {
		appendShenSha(&res.Hour, "金神", p.Hour.Gan+p.Hour.Zhi)
	}
	for _, zhi := range jieLuKongWangByDayGan[p.Day.Gan] {
		if zhi == p.Hour.Zhi {
			appendShenSha(&res.Hour, "截路空亡", zhi)
		}
	}
}

func specialDayRuleContains(pillar, name string) bool {
	for _, candidate := range specialDayShenShaRules[pillar] {
		if candidate == name {
			return true
		}
	}
	return false
}

func addGlobalShenSha(p ShenShaPillars, res *ShenShaCalcResult) {
	addExtraGlobalRules(p, res)
	if seq := classicalSanQiSequence([]string{p.Year.Gan, p.Month.Gan, p.Day.Gan, p.Hour.Gan}); seq != "" {
		appendShenSha(&res.Global, "三奇", seq)
	}
	dayPillar := p.Day.Gan + p.Day.Zhi
	if isTianShe(p.Month.Zhi, dayPillar) {
		appendShenSha(&res.Global, "天赦", dayPillar)
	}
	if isSiFei(p.Month.Zhi, dayPillar) {
		appendShenSha(&res.Global, "四废", dayPillar)
	}
	_, yearElement, _ := naYinNameAndElement(p.Year.Gan, p.Year.Zhi)
	branches := []string{p.Year.Zhi, p.Month.Zhi, p.Day.Zhi, p.Hour.Zhi}
	if yearElement == "火" {
		if hasBranchPair(branches, "戌", "亥") {
			appendShenSha(&res.Global, "天罗", "戌亥")
		}
	}
	if yearElement == "水" || yearElement == "土" {
		if hasBranchPair(branches, "辰", "巳") {
			appendShenSha(&res.Global, "地网", "辰巳")
		}
	}
	// 拱禄/拱贵
	addGongLuGongGui(p, res)
}

func classicalSanQiSequence(gans []string) string {
	profile := sanQiSemanticProfile()
	if len(gans) != profile.PillarCount {
		return ""
	}
	for _, start := range profile.WindowStarts {
		end := start + profile.WindowSize
		if start < 0 || end > len(gans) {
			return ""
		}
		seq := strings.Join(gans[start:end], "")
		if patternStringProfileContains(profile.StemWindows, seq) {
			return seq
		}
	}
	return ""
}

func isJinShenHourPillar(pillar model.Pillar) bool {
	return patternStringProfileContains(jinShenSemanticProfile().Pillars, pillar.Gan+pillar.Zhi)
}

func addGongLuGongGui(p ShenShaPillars, res *ShenShaCalcResult) {
	rule, ok := gongLuGongGuiRule(p.Day.Gan+p.Day.Zhi, p.Hour.Gan+p.Hour.Zhi)
	if !ok {
		return
	}
	branches := []string{p.Year.Zhi, p.Month.Zhi, p.Day.Zhi, p.Hour.Zhi}
	if branchInList(rule.Target, branches) {
		return
	}
	appendShenSha(&res.Global, rule.Name, rule.Target)
}

func gongLuGongGuiRule(dayPillar, hourPillar string) (shenShaRule, bool) {
	switch dayPillar + "/" + hourPillar {
	case "癸亥/癸丑", "癸丑/癸亥":
		return shenShaRule{Name: "拱禄", Target: "子"}, true
	case "丁巳/丁未", "己未/己巳":
		return shenShaRule{Name: "拱禄", Target: "午"}, true
	case "戊辰/戊午":
		return shenShaRule{Name: "拱禄", Target: "巳"}, true
	case "甲申/甲戌":
		return shenShaRule{Name: "拱贵", Target: "酉"}, true
	case "乙未/乙酉":
		return shenShaRule{Name: "拱贵", Target: "申"}, true
	case "甲寅/甲子":
		return shenShaRule{Name: "拱贵", Target: "丑"}, true
	case "戊申/戊午":
		return shenShaRule{Name: "拱贵", Target: "未"}, true
	case "辛丑/辛卯":
		return shenShaRule{Name: "拱贵", Target: "寅"}, true
	default:
		return shenShaRule{}, false
	}
}

// --- 全局追加：挂剑煞、雷霆煞 ---
func addExtraGlobalRules(p ShenShaPillars, res *ShenShaCalcResult) {
	branches := []string{p.Year.Zhi, p.Month.Zhi, p.Day.Zhi, p.Hour.Zhi}

	if evidence, ok := guaJianEvidence(branches); ok {
		appendShenSha(&res.Global, "挂剑煞", evidence)
	}
	if target := leiTingShaByMonthZhi[p.Month.Zhi]; target != "" {
		appendShenShaByTargetBranch(res, branches, "雷霆煞", target)
	}
}

func guaJianEvidence(branches []string) (string, bool) {
	if len(branches) != 4 {
		return "", false
	}
	counts := make(map[string]int, 4)
	for _, branch := range branches {
		counts[branch]++
	}
	if counts["巳"] == 1 && counts["酉"] == 1 && counts["丑"] == 1 && counts["申"] == 1 {
		return "巳酉丑申", true
	}
	if counts["巳"] >= 1 && counts["酉"] >= 1 && counts["丑"] >= 1 &&
		counts["巳"]+counts["酉"]+counts["丑"] == 4 {
		return "巳酉丑重", true
	}
	return "", false
}

func hasBranchPair(branches []string, first, second string) bool {
	return first != second && branchInList(first, branches) && branchInList(second, branches)
}

// --- 年支额外规则 ---
func addYearZhiExtra(p ShenShaPillars, branches []string, res *ShenShaCalcResult) {
	if t := sanHeLiuE[p.Year.Zhi]; t != "" {
		appendShenShaByTargetBranch(res, branches, "六厄", t)
	}
	// 攀鞍：驿马前一辰。
	if t := panAnByYearZhi(p.Year.Zhi); t != "" {
		appendShenShaByTargetBranch(res, branches, "攀鞍", t)
	}
	// 性别规则：元辰、勾煞、绞煞
	addGenderBasedShenSha(p, branches, res)
}

func panAnByYearZhi(yearZhi string) string {
	yiMa := sanHeShenShaRules[yearZhi].YiMa
	yiMaIndex := data.ZhiIndex(yiMa)
	if yiMaIndex < 0 {
		return ""
	}
	return data.Zhis[(yiMaIndex-1+len(data.Zhis))%len(data.Zhis)]
}

func addTianXingSha(p ShenShaPillars, res *ShenShaCalcResult) {
	if target := tianXingHourGanByYearZhi[p.Year.Zhi]; target != "" && target == p.Hour.Gan {
		appendShenSha(&res.Hour, "天刑煞", target)
	}
}

// --- 日柱额外规则 ---
func addDayExtra(p ShenShaPillars, branches []string, res *ShenShaCalcResult) {
	// 飞刃：羊刃对冲（阳干有飞刃，阴干无）
	if feiRen := feiRenByGan[p.Day.Gan]; feiRen != "" {
		appendShenShaByTargetBranch(res, branches, "飞刃", feiRen)
	}

	luZhi, _ := luBranchForStem(p.Day.Gan)
	if luZhi == "" {
		return
	}
	// 暗禄：四柱不见明禄，见禄位六合支。
	if !branchInList(luZhi, branches) {
		if heZhi := zhiLiuHe[luZhi]; heZhi != "" {
			appendShenShaByTargetBranch(res, branches, "暗禄", heZhi)
		}
	}
	// 四大空亡：按日柱所属旬查缺失五行，并定位到对应纳音柱。
	if empty := siDaKongWangElement(p.Day.Gan, p.Day.Zhi); empty != "" {
		for i, pillar := range []model.Pillar{p.Year, p.Month, p.Day, p.Hour} {
			_, element, _ := naYinNameAndElement(pillar.Gan, pillar.Zhi)
			if element == empty {
				appendShenShaToPillarIndex(res, i, "四大空亡", empty)
			}
		}
	}
}

func addJiaoLu(p ShenShaPillars, res *ShenShaCalcResult) {
	pillars := []model.Pillar{p.Year, p.Month, p.Day, p.Hour}
	for i := 0; i < len(pillars); i++ {
		for j := i + 1; j < len(pillars); j++ {
			left, right := pillars[i], pillars[j]
			leftLu, _ := luBranchForStem(left.Gan)
			rightLu, _ := luBranchForStem(right.Gan)
			if left.Gan == "" || right.Gan == "" || left.Gan == right.Gan || leftLu == "" || rightLu == "" {
				continue
			}
			if left.Zhi == rightLu && right.Zhi == leftLu {
				evidence := left.Gan + left.Zhi + "/" + right.Gan + right.Zhi
				appendShenShaToPillarIndex(res, i, "交禄", evidence)
				appendShenShaToPillarIndex(res, j, "交禄", evidence)
			}
		}
	}
}

func siDaKongWangElement(dayGan, dayZhi string) string {
	dayIndex := sixtyCycleIndex(dayGan, dayZhi)
	if dayIndex < 0 {
		return ""
	}
	switch dayIndex / 10 {
	case 0, 3: // 甲子、甲午旬
		return "水"
	case 2, 5: // 甲申、甲寅旬
		return "金"
	default: // 甲戌、甲辰旬五行俱全
		return ""
	}
}

// --- 性别规则：元辰、勾煞、绞煞 ---
func addGenderBasedShenSha(p ShenShaPillars, branches []string, res *ShenShaCalcResult) {
	yuanChenTarget, gouTarget, jiaoTarget := genderBasedShenShaTargets(p.Year.Gan, p.Year.Zhi, p.Gender)
	appendShenShaByTargetBranch(res, branches, "元辰", yuanChenTarget)
	appendShenShaByTargetBranch(res, branches, "勾煞", gouTarget)
	appendShenShaByTargetBranch(res, branches, "绞煞", jiaoTarget)
}

func genderBasedShenShaTargets(yearGan, yearZhi, gender string) (yuanChen, gou, jiao string) {
	chongZhi := zhiLiuChong[yearZhi]
	if chongZhi == "" {
		return "", "", ""
	}
	chongIdx := data.ZhiIndex(chongZhi)
	yearIdx := data.ZhiIndex(yearZhi)
	if chongIdx < 0 || yearIdx < 0 {
		return "", "", ""
	}

	isYang := yearGan == "甲" || yearGan == "丙" || yearGan == "戊" || yearGan == "庚" || yearGan == "壬"
	isMale := gender == "MALE"
	yangNanYinNv := (isYang && isMale) || (!isYang && !isMale) // 阳男阴女

	if yangNanYinNv {
		return data.Zhis[(chongIdx+1)%12], data.Zhis[(yearIdx+3)%12], data.Zhis[(yearIdx-3+12)%12]
	}
	return data.Zhis[(chongIdx-1+12)%12], data.Zhis[(yearIdx-3+12)%12], data.Zhis[(yearIdx+3)%12]
}

func branchInList(target string, branches []string) bool {
	for _, b := range branches {
		if b == target {
			return true
		}
	}
	return false
}

func anyBranchHasElement(elem string, branches []string) bool {
	for _, b := range branches {
		if data.ZhiElement[b] == elem {
			return true
		}
	}
	return false
}

var suppressedHighRiskShenSha = map[string]struct{}{
	"死符": {},
}

func appendShenSha(items *[]string, name, target string) {
	if _, suppressed := suppressedHighRiskShenSha[name]; suppressed {
		return
	}
	if target != "" {
		*items = append(*items, name+"："+target)
		return
	}
	*items = append(*items, name)
}

func matchGanZhiTarget(target string, p model.Pillar) bool {
	if target == "" {
		return false
	}
	runes := []rune(target)
	if len(runes) == 2 && data.GanIndex(string(runes[0])) >= 0 && data.ZhiIndex(string(runes[1])) >= 0 {
		return target == p.Gan+p.Zhi
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
		if data.Gans[i%10] == gan && data.Zhis[i%12] == zhi {
			return i
		}
	}
	return -1
}

type deXiuRule struct {
	De  string
	Xiu string
}

func deXiuRuleForMonth(monthZhi string) (deXiuRule, bool) {
	switch monthZhi {
	case "寅", "午", "戌":
		return deXiuRule{De: "丙丁", Xiu: "戊癸"}, true
	case "申", "子", "辰":
		return deXiuRule{De: "壬癸戊己", Xiu: "丙辛甲己"}, true
	case "亥", "卯", "未":
		return deXiuRule{De: "甲乙", Xiu: "丁壬"}, true
	case "巳", "酉", "丑":
		return deXiuRule{De: "庚辛", Xiu: "乙庚"}, true
	default:
		return deXiuRule{}, false
	}
}

func deXiuRole(monthZhi, gan string) string {
	rule, ok := deXiuRuleForMonth(monthZhi)
	if !ok || data.GanIndex(gan) < 0 {
		return ""
	}
	isDe := targetContainsGan(rule.De, gan)
	isXiu := targetContainsGan(rule.Xiu, gan)
	switch {
	case isDe && isXiu:
		return "德秀"
	case isDe:
		return "德"
	case isXiu:
		return "秀"
	default:
		return ""
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
		return dayPillar == "庚申"
	case "巳", "午", "未":
		return dayPillar == "壬子"
	case "申", "酉", "戌":
		return dayPillar == "甲寅"
	case "亥", "子", "丑":
		return dayPillar == "丙午"
	default:
		return false
	}
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

func deduplicateShenShaHits(res *ShenShaCalcResult) {
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
