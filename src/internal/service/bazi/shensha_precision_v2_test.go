package bazi

import (
	"fmt"
	"strings"
	"testing"
)

// =============================================================================
// 神煞精度测试 v2 — 修正所有柱位预期
// 对照依据：《渊海子平》《三命通会》
// =============================================================================

type shenShaAssert struct {
	Pillar string // "day"/"year"/"month"/"hour"/"global"
	God    string // 神煞名称（包含匹配）
}

type shenShaV2Case struct {
	ID, Desc, Source                string
	YearP, MonP, DayP, HouP, Gender string
	Asserts                         []shenShaAssert
	NegAsserts                      []shenShaAssert // 应不包含
}

var shenShaV2Cases = []shenShaV2Case{
	// ---- 天乙贵人 ----
	{ID: "SS-001", Desc: "甲日天乙贵人—甲寅日寅非丑未→全柱无贵人",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts:    []shenShaAssert{}, // 无柱有丑未→全无
		NegAsserts: []shenShaAssert{{"day", "天乙贵人"}, {"year", "天乙贵人"}},
		Source:     "三命通会PDF第97-98页"},
	{ID: "SS-002", Desc: "乙日天乙贵人在子申—年支子→年柱有贵人",
		YearP: "丙子", MonP: "己亥", DayP: "乙丑", HouP: "壬午", Gender: "MALE",
		Asserts: []shenShaAssert{{"year", "天乙贵人"}}, Source: "三命通会PDF第97-98页"},
	{ID: "SS-004", Desc: "庚日天乙贵人丑未—庚戌日支戌非丑未→日无贵人",
		YearP: "庚申", MonP: "乙酉", DayP: "庚戌", HouP: "庚辰", Gender: "MALE",
		Asserts:    []shenShaAssert{},
		NegAsserts: []shenShaAssert{{"day", "天乙贵人"}},
		Source:     "三命通会PDF第97-98页"},

	// ---- 禄神 ----
	{ID: "SS-010", Desc: "甲禄在寅—甲寅日支寅→日柱有禄神",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "禄神"}}, Source: "渊海子平"},
	{ID: "SS-011", Desc: "丙禄在巳—丙午日支午非巳→日无禄神",
		YearP: "丁未", MonP: "乙未", DayP: "丙午", HouP: "丁未", Gender: "FEMALE",
		Asserts:    []shenShaAssert{},
		NegAsserts: []shenShaAssert{{"day", "禄神"}},
		Source:     "渊海子平"},
	{ID: "SS-012", Desc: "癸禄在子—月柱有子→月柱有禄神",
		YearP: "癸酉", MonP: "甲子", DayP: "癸亥", HouP: "辛酉", Gender: "MALE",
		Asserts: []shenShaAssert{{"month", "禄神"}}, Source: "渊海子平"},

	// ---- 羊刃 ----
	{ID: "SS-020", Desc: "甲刃在卯—甲寅日支寅非卯→日无羊刃",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts:    []shenShaAssert{},
		NegAsserts: []shenShaAssert{{"day", "羊刃"}},
		Source:     "渊海子平"},
	{ID: "SS-021", Desc: "壬刃在子—壬戌日支戌非子→日无羊刃",
		YearP: "己酉", MonP: "乙丑", DayP: "壬戌", HouP: "庚子", Gender: "MALE",
		Asserts:    []shenShaAssert{},
		NegAsserts: []shenShaAssert{{"day", "羊刃"}},
		Source:     "渊海子平"},

	// ---- 三合神煞 ----
	{ID: "SS-032", Desc: "寅午戌华盖在戌—年支戌→年柱有华盖",
		YearP: "庚戌", MonP: "丙戌", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts:    []shenShaAssert{{"year", "华盖"}, {"month", "华盖"}},
		NegAsserts: []shenShaAssert{{"day", "华盖"}},
		Source:     "渊海子平"},
	{ID: "SS-033", Desc: "巳酉丑劫煞在寅—全局无寅支→无劫煞",
		YearP: "癸酉", MonP: "甲子", DayP: "癸亥", HouP: "辛酉", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "劫煞"}, {"global", "劫煞"}},
		Source:     "渊海子平"},

	// ---- 魁罡 ----
	{ID: "SS-050", Desc: "壬辰日→魁罡",
		YearP: "壬辰", MonP: "壬寅", DayP: "壬辰", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "魁罡"}}, Source: "渊海子平"},
	{ID: "SS-051", Desc: "甲寅日→非魁罡",
		YearP: "壬辰", MonP: "壬寅", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"day", "魁罡"}},
		Source:     "渊海子平"},

	// ---- 孤辰 ----
	{ID: "SS-060", Desc: "年酉属申酉戌,孤辰在亥—日支亥→日柱有孤辰",
		YearP: "癸酉", MonP: "甲子", DayP: "癸亥", HouP: "辛酉", Gender: "MALE",
		Asserts: []shenShaAssert{{"day", "孤辰"}}, Source: "三命通会PDF第118页；渊海子平PDF第632、744页"},

	// ---- 天德/月德 ----
	{ID: "SS-040", Desc: "寅月天德在丁—全局无丁→月柱无天德",
		YearP: "乙丑", MonP: "戊寅", DayP: "甲子", HouP: "丙寅", Gender: "MALE",
		NegAsserts: []shenShaAssert{{"month", "天德"}},
		Source:     "渊海子平"},
	{ID: "SS-042", Desc: "戌月天德在丙—丙戌月柱有天德，日干非丙则无月德",
		YearP: "庚戌", MonP: "丙戌", DayP: "甲寅", HouP: "庚午", Gender: "MALE",
		Asserts: []shenShaAssert{{"month", "天德"}},
		NegAsserts: []shenShaAssert{
			{"year", "月德"}, {"month", "月德"}, {"day", "月德"}, {"hour", "月德"},
		}, Source: "渊海子平"},
}

func TestShenShaPrecisionV2(t *testing.T) {
	svc := &BaziService{}
	var pass, fail int
	var fails []string
	var outputs []string

	for _, tc := range shenShaV2Cases {
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
		allOk := true
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
				allOk = false
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
				allOk = false
			}
		}

		if allOk {
			_ = allOk
		}
	}

	// Report
	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════╗\n")
	sb.WriteString("║  神煞规则断言V2（已修正柱位预期）         ║\n")
	sb.WriteString("╚══════════════════════════════════════════════╝\n")
	total := pass + fail
	sb.WriteString(fmt.Sprintf("\n规则断言: %d通过 + %d失败 = %d总\n", pass, fail, total))
	sb.WriteString(fmt.Sprintf("用例: %d\n\n", len(shenShaV2Cases)))
	if len(fails) > 0 {
		sb.WriteString("失败:\n" + strings.Join(fails, "\n") + "\n")
	} else {
		sb.WriteString("全部通过！\n")
	}
	sb.WriteString("\n实际输出:\n" + strings.Join(outputs, "\n"))

	t.Log("\n" + sb.String())

	if fail > 0 {
		t.Errorf("神煞测试有 %d 个断言失败", fail)
	}
}
