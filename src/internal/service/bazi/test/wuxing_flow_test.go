package bazi_test

import (
	. "bazi/internal/service/bazi"
	"fmt"
	"os"
	"strings"
	"testing"
)

// =============================================================================
// 五行流通/通关/缺失测试
// 使用 CalculateFromPillars 验证五行流通分析非空且合理
// =============================================================================

type wuxingFlowCase struct {
	ID, Desc, Source                 string
	YearP, MonP, DayP, HouP, Gender string
}

var wuxingFlowCases = []wuxingFlowCase{
	{ID: "WX-001", Desc: "甲日申月—金旺木弱",
		YearP: "壬辰", MonP: "戊申", DayP: "甲午", HouP: "庚午", Gender: "MALE",
		Source: "滴天髓"},
	{ID: "WX-002", Desc: "丙午日午月—火炎土燥",
		YearP: "丁未", MonP: "丙午", DayP: "丙午", HouP: "甲午", Gender: "MALE",
		Source: "滴天髓"},
	{ID: "WX-003", Desc: "壬子日子月—水旺",
		YearP: "壬子", MonP: "壬子", DayP: "壬子", HouP: "庚子", Gender: "MALE",
		Source: "滴天髓"},
	{ID: "WX-004", Desc: "甲午年己巳月戊戌日庚申时—五行较全",
		YearP: "甲午", MonP: "己巳", DayP: "戊戌", HouP: "庚申", Gender: "MALE",
		Source: "滴天髓"},
	{ID: "WX-005", Desc: "庚申日申月—金旺",
		YearP: "庚申", MonP: "甲申", DayP: "庚申", HouP: "庚辰", Gender: "MALE",
		Source: "滴天髓"},
	{ID: "WX-006", Desc: "甲寅日寅月—木旺",
		YearP: "甲寅", MonP: "丙寅", DayP: "甲寅", HouP: "丙寅", Gender: "MALE",
		Source: "滴天髓"},
	{ID: "WX-007", Desc: "己未年癸酉月己未日甲子时",
		YearP: "己未", MonP: "癸酉", DayP: "己未", HouP: "甲子", Gender: "MALE",
		Source: "滴天髓"},
}

func TestWuxingFlowConsistency(t *testing.T) {
	svc := &BaziService{}
	var pass, fail int
	var fails []string
	var outputs []string

	for _, tc := range wuxingFlowCases {
		r, err := svc.CalculateFromPillars(tc.YearP, tc.MonP, tc.DayP, tc.HouP, tc.Gender)
		if err != nil {
			fail++
			fails = append(fails, fmt.Sprintf("[%s] 计算失败: %v", tc.ID, err))
			continue
		}

		casePass := true

		// 1. WuXingFlow非空验证
		wf := r.WuXingFlow
		if wf.FlowType == "" {
			fails = append(fails, fmt.Sprintf("[%s] FlowType为空", tc.ID))
			fail++
			casePass = false
		}
		if wf.Advice == "" {
			fails = append(fails, fmt.Sprintf("[%s] Advice为空", tc.ID))
			fail++
			casePass = false
		}
		// DayElement must not be empty
		if wf.DayElement == "" {
			fails = append(fails, fmt.Sprintf("[%s] DayElement为空", tc.ID))
			fail++
			casePass = false
		}

		// 2. TongGuan非空验证
		tg := r.TongGuan
		if tg.Description == "" && tg.HasTongGuan {
			fails = append(fails, fmt.Sprintf("[%s] TongGuan有通关但Description为空", tc.ID))
			fail++
			casePass = false
		}

		// 3. MissingElements非空验证
		me := r.MissingElements
		if me.Severity == "" {
			fails = append(fails, fmt.Sprintf("[%s] Severity为空", tc.ID))
			fail++
			casePass = false
		}
		// RemedyAdvice只在有缺失或偏弱时非空
		if len(me.MissingElements) > 0 || len(me.WeakElements) > 0 {
			if len(me.RemedyAdvice) == 0 {
				fails = append(fails, fmt.Sprintf("[%s] 有缺失/偏弱但RemedyAdvice为空", tc.ID))
				fail++
				casePass = false
			}
		}

		// 4. FlowPatternDesc非空验证
		if r.FlowPatternDesc == "" {
			fails = append(fails, fmt.Sprintf("[%s] FlowPatternDesc为空", tc.ID))
			fail++
			casePass = false
		}

		if casePass {
			pass++
		}

		// 详细输出
		var flowDir string
		if len(wf.FlowPaths) > 0 {
			flowDir = strings.Join(wf.FlowPaths, ",")
		} else {
			flowDir = "(无)"
		}
		var tgDesc string
		if tg.HasTongGuan {
			tgDesc = fmt.Sprintf("%s(%.0f%%)", tg.TongGuanElement, tg.Weight*100)
		} else {
			tgDesc = "无通关"
		}
		outputs = append(outputs, fmt.Sprintf("[%s] %s: 流通=%s 类型=%s 顺畅=%v 通关=%s 缺失=%v 偏弱=%v FlowDesc=%s",
			tc.ID, tc.Desc,
			flowDir, wf.FlowType, wf.IsSmooth,
			tgDesc, me.MissingElements, me.WeakElements,
			r.FlowPatternDesc))
	}

	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════╗\n")
	sb.WriteString("║  五行流通/通关/缺失精度测试                 ║\n")
	sb.WriteString("╚══════════════════════════════════════════════╝\n")
	total := pass + fail
	sb.WriteString(fmt.Sprintf("\n断言: %d通过 + %d失败 = %d总 | 准确率: %.1f%%\n", pass, fail, total, pctV3(pass, total)))
	sb.WriteString(fmt.Sprintf("用例数: %d\n\n", len(wuxingFlowCases)))
	if len(fails) > 0 {
		sb.WriteString("失败:\n" + strings.Join(fails, "\n") + "\n")
	} else {
		sb.WriteString("全部通过！\n")
	}
	sb.WriteString("\n详情:\n" + strings.Join(outputs, "\n"))

	t.Log("\n" + sb.String())
	os.WriteFile("/tmp/bazi_wuxing_flow_report.txt", []byte(sb.String()), 0644)

	if fail > 0 {
		t.Errorf("五行流通测试有 %d 个断言失败", fail)
	}
}

// TestWuxingFlowNonEmpty 基础非空验证 — 对多种四柱组合检查WuXingFlow基本结构
func TestWuxingFlowNonEmpty(t *testing.T) {
	svc := &BaziService{}
	variations := []struct {
		YP, MP, DP, HP, G string
	}{
		{"甲子", "丙寅", "甲子", "庚午", "MALE"},
		{"乙丑", "丁卯", "乙丑", "辛未", "FEMALE"},
		{"丙寅", "戊辰", "丙寅", "壬申", "MALE"},
		{"丁卯", "己巳", "丁卯", "癸酉", "FEMALE"},
		{"戊辰", "庚午", "戊辰", "甲戌", "MALE"},
		{"己巳", "辛未", "己巳", "乙亥", "FEMALE"},
		{"庚午", "壬申", "庚午", "丙子", "MALE"},
		{"辛未", "癸酉", "辛未", "丁丑", "FEMALE"},
		{"壬申", "甲戌", "壬申", "戊寅", "MALE"},
		{"癸酉", "乙亥", "癸酉", "己卯", "FEMALE"},
	}

	var failCount int
	for _, v := range variations {
		r, err := svc.CalculateFromPillars(v.YP, v.MP, v.DP, v.HP, v.G)
		if err != nil {
			t.Errorf("[%s/%s/%s/%s] 计算失败: %v", v.YP, v.MP, v.DP, v.HP, err)
			failCount++
			continue
		}

		// WuXingFlow必须非空
		if r.WuXingFlow.FlowType == "" {
			t.Errorf("[%s/%s/%s/%s] WuXingFlow.FlowType为空",
				v.YP, v.MP, v.DP, v.HP)
			failCount++
		}
		if r.WuXingFlow.DayElement == "" {
			t.Errorf("[%s/%s/%s/%s] WuXingFlow.DayElement为空",
				v.YP, v.MP, v.DP, v.HP)
			failCount++
		}

		// MissingElements必须非空
		if r.MissingElements.Severity == "" {
			t.Errorf("[%s/%s/%s/%s] MissingElements.Severity为空",
				v.YP, v.MP, v.DP, v.HP)
			failCount++
		}

		// FlowPatternDesc必须非空
		if r.FlowPatternDesc == "" {
			t.Errorf("[%s/%s/%s/%s] FlowPatternDesc为空",
				v.YP, v.MP, v.DP, v.HP)
			failCount++
		}
	}

	if failCount > 0 {
		t.Errorf("五行流通非空验证: %d 个失败", failCount)
	} else {
		t.Logf("五行流通非空验证: 全部%d个通过", len(variations))
	}
}
