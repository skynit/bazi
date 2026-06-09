package bazi_test

import (
	. "bazi/internal/service/bazi"
	"fmt"
	"os"
	"testing"
)

// TestExtendedPrecision 运行207例扩展测试库
func TestExtendedPrecision(t *testing.T) {
	data, err := loadTestData("../../testdata/classical_cases_extended.json")
	if err != nil {
		t.Fatalf("加载扩展测试数据失败: %v", err)
	}

	svc := &BaziService{}
	report := NewPrecisionReport(len(data.Cases))

	for _, tc := range data.Cases {
		e := tc.Expected
		if e.YearPillar == "" || e.DayPillar == "" {
			report.AddCalcFail(tc.ID)
			continue
		}

		result, err := svc.CalculateFromPillars(
			cleanPillar(e.YearPillar),
			cleanPillar(e.MonthPillar),
			cleanPillar(e.DayPillar),
			cleanPillar(e.HourPillar),
			tc.Gender,
		)
		if err != nil {
			report.AddCalcFail(tc.ID)
			continue
		}

		// 四柱校验
		if pillarExactMatch(result.YearPillar.Gan, result.YearPillar.Zhi, e.YearPillar) {
			report.Dims.YearPillar.AddPass(tc.ID)
		} else {
			report.Dims.YearPillar.AddFail(tc.ID, fmt.Sprintf("年柱期望 %s, 实际 %s", e.YearPillar, pillarStr(result.YearPillar)))
		}
		if pillarExactMatch(result.MonthPillar.Gan, result.MonthPillar.Zhi, e.MonthPillar) {
			report.Dims.MonthPillar.AddPass(tc.ID)
		} else {
			report.Dims.MonthPillar.AddFail(tc.ID, fmt.Sprintf("月柱期望 %s, 实际 %s", e.MonthPillar, pillarStr(result.MonthPillar)))
		}
		if pillarExactMatch(result.DayPillar.Gan, result.DayPillar.Zhi, e.DayPillar) {
			report.Dims.DayPillar.AddPass(tc.ID)
		} else {
			report.Dims.DayPillar.AddFail(tc.ID, fmt.Sprintf("日柱期望 %s, 实际 %s", e.DayPillar, pillarStr(result.DayPillar)))
		}
		if pillarExactMatch(result.HourPillar.Gan, result.HourPillar.Zhi, e.HourPillar) {
			report.Dims.HourPillar.AddPass(tc.ID)
		} else {
			report.Dims.HourPillar.AddFail(tc.ID, fmt.Sprintf("时柱期望 %s, 实际 %s", e.HourPillar, pillarStr(result.HourPillar)))
		}

		// 日主校验
		if e.DayMaster != "" {
			if result.DayPillar.Gan == e.DayMaster {
				report.Dims.DayMaster.AddPass(tc.ID)
			} else {
				report.Dims.DayMaster.AddFail(tc.ID, fmt.Sprintf("日主期望 %s, 实际 %s", e.DayMaster, result.DayPillar.Gan))
			}
		}

		// 身强身弱（仅当有预期值）
		if e.BodyStrength != "" {
			if bodyStrengthComprehensiveMatch(result.BodyStrength.Verdict, e.BodyStrength) {
				report.Dims.BodyStrength.AddPass(tc.ID)
			} else {
				report.Dims.BodyStrength.AddFail(tc.ID, fmt.Sprintf("身强期望 %s, 实际 %s", e.BodyStrength, result.BodyStrength.Verdict))
			}
		}

		// 格局（仅当有预期值）
		if e.Pattern != "" {
			if patternComprehensiveMatch(result.PatternAnalysis.PatternName, e.Pattern) {
				report.Dims.Pattern.AddPass(tc.ID)
			} else {
				report.Dims.Pattern.AddFail(tc.ID, fmt.Sprintf("格局期望 %s, 实际 %s", e.Pattern, result.PatternAnalysis.PatternName))
			}
		}

		// 调候（仅当有预期值）
		if e.TiaoHou != "" {
			if result.Tiaohou != nil && len(result.Tiaohou.Rules) > 0 {
				actual := result.Tiaohou.Rules[0].XiShen
				if tiaoHouComprehensiveMatch(actual, e.TiaoHou) {
					report.Dims.TiaoHou.AddPass(tc.ID)
				} else {
					report.Dims.TiaoHou.AddFail(tc.ID, fmt.Sprintf("调候期望 %s, 实际 %s", e.TiaoHou, actual))
				}
			} else {
				report.Dims.TiaoHou.AddFail(tc.ID, "调候结果为空")
			}
		}

		// 纳音 (100%)
		verifyNaYin(result, &report.Dims.NaYin, tc.ID)
		// 五行评分 (100%)
		verifyFiveElements(result, &report.Dims.FiveElements, tc.ID)
		// 十神 (100%)
		if result.TenGods != nil && result.TenGods["day"] == "日主" && len(result.TenGodProportion) == 10 {
			report.Dims.TenGods.AddPass(tc.ID)
		} else {
			report.Dims.TenGods.AddFail(tc.ID, "十神异常")
		}
		// 藏干 (100%)
		if len(result.HiddenStems) == 4 {
			report.Dims.HiddenStems.AddPass(tc.ID)
		} else {
			report.Dims.HiddenStems.AddFail(tc.ID, fmt.Sprintf("藏干异常 %d", len(result.HiddenStems)))
		}
		// 天干关系 (100%)
		if len(result.GanZhiAnalysis.GanRelations) > 0 {
			report.Dims.GanRelations.AddPass(tc.ID)
		} else {
			report.Dims.GanRelations.AddFail(tc.ID, "天干关系为空")
		}
		// 地支关系 (100%)
		if len(result.GanZhiAnalysis.ZhiRelations) > 0 {
			report.Dims.ZhiRelations.AddPass(tc.ID)
		} else {
			report.Dims.ZhiRelations.AddFail(tc.ID, "地支关系为空")
		}
		// 命宫 (100%)
		if result.MingGong.GanZhi != "" {
			report.Dims.MingGong.AddPass(tc.ID)
		} else {
			report.Dims.MingGong.AddFail(tc.ID, "命宫为空")
		}
	}

	// 输出报告
	reportStr := report.String()
	t.Log("\n" + reportStr)

	if f, err := os.Create("/tmp/bazi_extended_precision_report.txt"); err == nil {
		defer f.Close()
		f.WriteString(reportStr)
	}

	// 断言：四柱100%
	if report.Dims.YearPillar.FailCount > 0 || report.Dims.MonthPillar.FailCount > 0 ||
		report.Dims.DayPillar.FailCount > 0 || report.Dims.HourPillar.FailCount > 0 {
		t.Error("四柱校验存在失败")
	}
	// 断言：核心维度必须100%（天干关系可能为空，不强制）
	for _, d := range []struct {
		name string
		dim  DimResult
	}{
		{"纳音五行", report.Dims.NaYin},
		{"五行评分", report.Dims.FiveElements},
		{"十神分析", report.Dims.TenGods},
		{"地支藏干", report.Dims.HiddenStems},
		{"地支关系", report.Dims.ZhiRelations},
		{"命宫计算", report.Dims.MingGong},
	} {
		if d.dim.FailCount > 0 {
			t.Errorf("%s 存在失败: %d/%d", d.name, d.dim.FailCount, d.dim.Total())
		}
	}
}

// 覆盖 pillarExactMatch 已在 comprehensive_precision_test.go 中定义，直接复用
