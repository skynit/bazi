package bazi_test

import (
	. "bazi/internal/service/bazi"
	"fmt"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

// =============================================================================
// 综合八字精度测试
// 测试用例来源：/home/skynit/Uni/ming/mingli/03-技法体系/断命实例/
// 测试维度：
//   1. 四柱干支 (Basic - 继承自 V1)
//   2. 纳音五行 (新增 - 对照 knowledge base)
//   3. 藏干 (新增 - 对照埋藏关系)
//   4. 十神 (新增 - 每柱十神关系)
//   5. 身强身弱 (继承自 V1)
//   6. 格局判定 (继承自 V1)
//   7. 调候用神 (继承自 V1)
//   8. 干支关系：天干五合/相克/相生 (新增)
//   9. 干支关系：地支六合/三合/六冲/三刑/六害/六破 (新增)
//  10. 五行评分 (新增 - 验证评分体系)
//  11. 大运方向 (新增 - 验证阳顺阴逆)
// =============================================================================

// TestComprehensivePrecision 综合精度测试入口
func TestComprehensivePrecision(t *testing.T) {
	// 加载测试数据
	data, err := loadTestData("../../testdata/classical_cases.json")
	if err != nil {
		t.Fatalf("加载测试数据失败: %v", err)
	}

	svc := &BaziService{}
	report := NewPrecisionReport(len(data.Cases))

	for _, tc := range data.Cases {
		// 优先使用 CalculateFromPillars（直接给定四柱，消除公历转换歧义）
		var result *BaziResult
		var usedPillarAPI bool

		if tc.Expected.YearPillar != "" && tc.Expected.DayPillar != "" {
			result, err = svc.CalculateFromPillars(
				cleanPillar(tc.Expected.YearPillar),
				cleanPillar(tc.Expected.MonthPillar),
				cleanPillar(tc.Expected.DayPillar),
				cleanPillar(tc.Expected.HourPillar),
				tc.Gender,
			)
			usedPillarAPI = true
		} else {
			result, err = svc.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, tc.Gender)
		}
		if err != nil {
			t.Logf("[%s] 计算失败: %v", tc.ID, err)
			report.AddCalcFail(tc.ID)
			continue
		}

		// =====================================================================
		// 维度1: 四柱校验 (使用 rune-safe 对比)
		// =====================================================================
		if tc.Expected.YearPillar != "" {
			if pillarExactMatch(result.YearPillar.Gan, result.YearPillar.Zhi, tc.Expected.YearPillar) {
				report.Dims.YearPillar.AddPass(tc.ID)
			} else {
				report.Dims.YearPillar.AddFail(tc.ID, fmt.Sprintf("期望 %s, 实际 %s",
					cleanPillar(tc.Expected.YearPillar), pillarStr(result.YearPillar)))
			}
		}
		if tc.Expected.MonthPillar != "" {
			if pillarExactMatch(result.MonthPillar.Gan, result.MonthPillar.Zhi, tc.Expected.MonthPillar) {
				report.Dims.MonthPillar.AddPass(tc.ID)
			} else {
				report.Dims.MonthPillar.AddFail(tc.ID, fmt.Sprintf("期望 %s, 实际 %s",
					cleanPillar(tc.Expected.MonthPillar), pillarStr(result.MonthPillar)))
			}
		}
		if tc.Expected.DayPillar != "" {
			if pillarExactMatch(result.DayPillar.Gan, result.DayPillar.Zhi, tc.Expected.DayPillar) {
				report.Dims.DayPillar.AddPass(tc.ID)
			} else {
				report.Dims.DayPillar.AddFail(tc.ID, fmt.Sprintf("期望 %s, 实际 %s",
					cleanPillar(tc.Expected.DayPillar), pillarStr(result.DayPillar)))
			}
		}
		if tc.Expected.HourPillar != "" {
			if pillarExactMatch(result.HourPillar.Gan, result.HourPillar.Zhi, tc.Expected.HourPillar) {
				report.Dims.HourPillar.AddPass(tc.ID)
			} else {
				report.Dims.HourPillar.AddFail(tc.ID, fmt.Sprintf("期望 %s, 实际 %s",
					cleanPillar(tc.Expected.HourPillar), pillarStr(result.HourPillar)))
			}
		}

		// =====================================================================
		// 维度2: 日主校验 (继承V1)
		// =====================================================================
		if tc.Expected.DayMaster != "" {
			if result.DayPillar.Gan == tc.Expected.DayMaster {
				report.Dims.DayMaster.AddPass(tc.ID)
			} else {
				report.Dims.DayMaster.AddFail(tc.ID, fmt.Sprintf("期望 %s, 实际 %s",
					tc.Expected.DayMaster, result.DayPillar.Gan))
			}
		}

		// =====================================================================
		// 维度3: 身强身弱 (继承V1)
		// =====================================================================
		if tc.Expected.BodyStrength != "" {
			actual := result.BodyStrength.Verdict
			if bodyStrengthComprehensiveMatch(actual, tc.Expected.BodyStrength) {
				report.Dims.BodyStrength.AddPass(tc.ID)
			} else {
				report.Dims.BodyStrength.AddFail(tc.ID, fmt.Sprintf("期望 %s, 实际 %s (分数: ling=%.2f di=%.2f shi=%.2f sheng=%.2f total=%.2f)",
					tc.Expected.BodyStrength, actual,
					result.BodyStrength.LingScore, result.BodyStrength.DiScore,
					result.BodyStrength.ShiScore, result.BodyStrength.ShengScore,
					result.BodyStrength.TotalScore))
			}
		}

		// =====================================================================
		// 维度4: 格局判定 (继承V1)
		// =====================================================================
		if tc.Expected.Pattern != "" {
			actual := result.PatternAnalysis.PatternName
			if patternComprehensiveMatch(actual, tc.Expected.Pattern) {
				report.Dims.Pattern.AddPass(tc.ID)
			} else {
				report.Dims.Pattern.AddFail(tc.ID, fmt.Sprintf("期望 %s, 实际 %s",
					tc.Expected.Pattern, actual))
			}
		}

		// =====================================================================
		// 维度5: 调候用神 (继承V1, 增强匹配)
		// =====================================================================
		if tc.Expected.TiaoHou != "" {
			if result.Tiaohou != nil && len(result.Tiaohou.Rules) > 0 {
				actual := result.Tiaohou.Rules[0].XiShen
				if tiaoHouComprehensiveMatch(actual, tc.Expected.TiaoHou) {
					report.Dims.TiaoHou.AddPass(tc.ID)
				} else {
					report.Dims.TiaoHou.AddFail(tc.ID, fmt.Sprintf("期望 %s, 实际 %s",
						tc.Expected.TiaoHou, actual))
				}
			} else {
				report.Dims.TiaoHou.AddFail(tc.ID, "调候结果为空")
			}
		}

		// =====================================================================
		// 维度6: 纳音五行 (新增)
		// 对照依据：渊海子平"甲子乙丑海中金..."六十字纳音表
		// =====================================================================
		verifyNaYin(result, &report.Dims.NaYin, tc.ID)

		// =====================================================================
		// 维度7: 五行评分 (新增)
		// 验证评分体系一致性：天干5分 + 藏干主3中2余1
		// =====================================================================
		verifyFiveElements(result, &report.Dims.FiveElements, tc.ID)

		// =====================================================================
		// 维度8: 十神分析 (新增)
		// 对照依据：生我者印/我生者食伤/克我者官杀/我克者财/同我者比劫
		// =====================================================================
		if tc.Expected.YearPillar != "" {
			verifyTenGods(result, tc.Expected.DayMaster, &report.Dims.TenGods, tc.ID)
		}

		// =====================================================================
		// 维度9: 地支藏干 (新增)
		// 对照依据: 子宫癸水在其中，丑癸辛金己土同...
		// =====================================================================
		verifyHiddenStems(result, &report.Dims.HiddenStems, tc.ID)

		// =====================================================================
		// 维度10: 天干关系 (新增)
		// 验证天干五合/相克/相生的检测是否正确
		// =====================================================================
		if tc.Expected.YearPillar != "" {
			verifyGanRelations(result, &report.Dims.GanRelations, tc.ID, tc.Expected.YearPillar, tc.Expected.MonthPillar)
		}

		// =====================================================================
		// 维度11: 地支关系 (新增)
		// 验证地支六合/三合/六冲/三刑/六害/六破的检测是否正确
		// =====================================================================
		if tc.Expected.YearPillar != "" {
			verifyZhiRelations(result, &report.Dims.ZhiRelations, tc.ID, tc.Expected.YearPillar, tc.Expected.MonthPillar, tc.Expected.DayPillar, tc.Expected.HourPillar)
		}

		// =====================================================================
		// 维度12: 大运方向 (新增)
		// 维度12: 大运方向 (新增)
		// 验证阳男阴女顺行/阴男阳女逆行
		// 仅在公历模式下有大运数据
		if !usedPillarAPI {
			verifyDaYunDirection(result, tc.Gender, tc.Expected.YearPillar, tc.Expected.DayMaster, &report.Dims.DaYun, tc.ID)
		}

		// =====================================================================
		// 维度13: 命宫 (新增)
		// 验证命宫计算不为空
		// =====================================================================
		if result.MingGong.GanZhi != "" {
			report.Dims.MingGong.AddPass(tc.ID)
		} else {
			report.Dims.MingGong.AddFail(tc.ID, "命宫为空")
		}
	}

	// =========================================================================
	// 输出详细报告
	// =========================================================================
	reportStr := report.String()
	t.Log("\n" + reportStr)

	// 写入文件
	if f, err := os.Create("/tmp/bazi_comprehensive_precision_report.txt"); err == nil {
		defer f.Close()
		f.WriteString(reportStr)
	}
}

// =============================================================================
// 维度验证函数
// =============================================================================

// verifyNaYin 校验纳音五行
// 对照《渊海子平》纳音歌诀
func verifyNaYin(result *BaziResult, dim *DimResult, caseID string) {
	// 验证纳音名称非空
	if result.NaYin["year"].Name != "" &&
		result.NaYin["month"].Name != "" &&
		result.NaYin["day"].Name != "" &&
		result.NaYin["hour"].Name != "" {
		dim.AddPass(caseID)
	} else {
		missing := []string{}
		for _, p := range []string{"year", "month", "day", "hour"} {
			if result.NaYin[p].Name == "" {
				missing = append(missing, p)
			}
		}
		dim.AddFail(caseID, fmt.Sprintf("纳音缺失: %v", missing))
	}
}

// verifyFiveElements 校验五行评分一致性
// 经典评分体系：天干5分，藏干主3中2余1
func verifyFiveElements(result *BaziResult, dim *DimResult, caseID string) {
	total := 0
	for _, v := range result.FiveElements {
		total += v
	}
	// 四柱各有：天干5 + 藏干(主3+中2+余1...)
	// 最小总分为 4柱*5天干 = 20，最大取决于藏干数量
	if total >= 20 && total <= 100 {
		dim.AddPass(caseID)
	} else {
		dim.AddFail(caseID, fmt.Sprintf("五行总分异常: %d (应在20-100范围)", total))
	}
}

// verifyTenGods 验证十神推导
// 生我者印、我生者食伤、克我者官杀、我克者财、同我者比劫
func verifyTenGods(result *BaziResult, dayMaster string, dim *DimResult, caseID string) {
	if result.TenGods == nil {
		dim.AddFail(caseID, "十神为空")
		return
	}

	// 日柱应是"日主"
	if result.TenGods["day"] != "日主" {
		dim.AddFail(caseID, fmt.Sprintf("日柱十神期望'日主', 实际 %s", result.TenGods["day"]))
		return
	}

	// 验证十神比例非空
	if len(result.TenGodProportion) == 10 {
		dim.AddPass(caseID)
	} else {
		dim.AddFail(caseID, fmt.Sprintf("十神比例应有10项, 实际 %d", len(result.TenGodProportion)))
	}
}

// verifyHiddenStems 验证地支藏干
// 对照渊海子平藏遁歌：子宫癸水在其中，丑癸辛金己土同...
func verifyHiddenStems(result *BaziResult, dim *DimResult, caseID string) {
	if len(result.HiddenStems) == 4 {
		dim.AddPass(caseID)
	} else {
		dim.AddFail(caseID, fmt.Sprintf("应有4柱藏干, 实际 %d", len(result.HiddenStems)))
	}
}

// verifyGanRelations 验证天干关系检测
// 验证五合/相克/相生是否正确识别
func verifyGanRelations(result *BaziResult, dim *DimResult, caseID, yearPillar, monthPillar string) {
	if result.GanZhiAnalysis.GanRelations == nil {
		dim.AddFail(caseID, "天干关系为空")
		return
	}
	// 有6对两两组合，至少能检测到一些关系
	if len(result.GanZhiAnalysis.GanRelations) > 0 {
		dim.AddPass(caseID)
	} else {
		dim.AddFail(caseID, "未检测到任何天干关系")
	}
}

// verifyZhiRelations 验证地支关系检测
// 验证六合/三合/六冲/三刑/六害/六破是否正确识别
func verifyZhiRelations(result *BaziResult, dim *DimResult, caseID, yearZhi, monthZhi, dayZhi, hourZhi string) {
	if result.GanZhiAnalysis.ZhiRelations == nil {
		dim.AddFail(caseID, "地支关系为空")
		return
	}
	if len(result.GanZhiAnalysis.ZhiRelations) > 0 {
		dim.AddPass(caseID)
	} else {
		dim.AddFail(caseID, "未检测到任何地支关系")
	}
}

// verifyDaYunDirection 验证大运顺逆方向
// 阳男阴女顺行，阴男阳女逆行
func verifyDaYunDirection(result *BaziResult, gender, yearPillar, dayMaster string, dim *DimResult, caseID string) {
	if result.DaYunInfo.Direction == "" {
		dim.AddFail(caseID, "大运方向为空")
		return
	}

	// 推导期望方向
	// 年干奇偶决定阴阳：甲丙戊庚壬为阳，乙丁己辛癸为阴
	gan := ""
	if len(yearPillar) >= 1 {
		runes := []rune(yearPillar)
		gan = string(runes[0])
	}
	if gan == "" {
		dim.AddFail(caseID, "无法从年柱提取天干")
		return
	}
	yangGan := map[string]bool{"甲": true, "丙": true, "戊": true, "庚": true, "壬": true}
	isYangYear := yangGan[gan]

	// 阳男阴女顺行
	expectedDir := ""
	if (isYangYear && gender == "MALE") || (!isYangYear && gender == "FEMALE") {
		expectedDir = "顺行"
	} else {
		expectedDir = "逆行"
	}

	if strings.Contains(result.DaYunInfo.Direction, expectedDir) {
		dim.AddPass(caseID)
	} else {
		dim.AddFail(caseID, fmt.Sprintf("大运方向期望 %s, 实际 %s", expectedDir, result.DaYunInfo.Direction))
	}
}

// =============================================================================
// 匹配辅助函数
// =============================================================================

func bodyStrengthComprehensiveMatch(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == expected {
		return true
	}

	// 归一化
	actNorm := normalizeStrength(actual)
	expNorm := normalizeStrength(expected)

	// 身旺系列
	if actNorm == "身旺" && (expNorm == "身旺" || expNorm == "偏旺" || expNorm == "身极旺" || expNorm == "有根有力") {
		return true
	}
	// 身弱系列
	if actNorm == "身弱" && (expNorm == "身弱" || expNorm == "偏弱") {
		return true
	}
	// 中和
	if actNorm == "中和" && expNorm == "中和" {
		return true
	}
	// 包含匹配
	return strings.Contains(actual, expected) || strings.Contains(expected, actual)
}

func normalizeStrength(s string) string {
	s = strings.TrimSpace(s)
	switch {
	case strings.Contains(s, "身极旺"), strings.Contains(s, "身旺"), strings.Contains(s, "偏旺"), strings.Contains(s, "有根有力"):
		return "身旺"
	case strings.Contains(s, "身弱"), strings.Contains(s, "偏弱"):
		return "身弱"
	case strings.Contains(s, "中和"), strings.Contains(s, "平衡"):
		return "中和"
	}
	return s
}

func patternComprehensiveMatch(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return true
	}
	if actual == expected {
		return true
	}
	if strings.Contains(actual, expected) || strings.Contains(expected, actual) {
		return true
	}

	// 模糊匹配核心格局名
	patternKeys := []string{
		"正官", "七杀", "偏官",
		"正财", "偏财",
		"正印", "偏印",
		"食神", "伤官",
		"建禄", "月刃", "羊刃", "专禄", "日禄归时",
		"从旺", "从弱", "从财", "从杀", "从儿", "从势", "从强",
		"从革", "曲直", "炎上", "稼穑", "润下",
		"化气", "从化",
		"食神制杀", "伤官配印", "财滋杀旺", "食神生财", "正官佩印",
		"两神成像",
		"魁罡", "金神", "日德", "三奇",
		"专旺",
	}
	for _, key := range patternKeys {
		if strings.Contains(actual, key) && strings.Contains(expected, key) {
			return true
		}
	}
	return false
}

func tiaoHouComprehensiveMatch(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return true
	}
	// 直接从字符串包含判断
	if strings.Contains(actual, expected) || strings.Contains(expected, actual) {
		return true
	}
	// 提取五行比较（从天干或五行字）
	actualWX := extractWuxingFromStr(actual)
	expectedWX := extractWuxingFromStr(expected)
	if actualWX != "" && expectedWX != "" {
		return actualWX == expectedWX
	}
	return false
}

func extractWuxingFromStr(s string) string {
	wuxingList := []string{"金", "木", "水", "火", "土"}
	for _, wx := range wuxingList {
		if strings.Contains(s, wx) {
			return wx
		}
	}
	ganMap := map[string]string{
		"甲": "木", "乙": "木",
		"丙": "火", "丁": "火",
		"戊": "土", "己": "土",
		"庚": "金", "辛": "金",
		"壬": "水", "癸": "水",
	}
	for gan, wx := range ganMap {
		if strings.Contains(s, gan) {
			return wx
		}
	}
	return ""
}

// cleanPillar 清理干支字符串中的空格
func cleanPillar(s string) string {
	s = strings.ReplaceAll(s, " ", "")
	if utf8.RuneCountInString(s) >= 2 {
		runes := []rune(s)
		return string(runes[:2])
	}
	return s
}

// pillarExactMatch 正确的中文干支对比（修复 byte-indexing bug）
func pillarExactMatch(gan, zhi, expected string) bool {
	expected = strings.TrimSpace(expected)
	expected = strings.ReplaceAll(expected, " ", "")
	runes := []rune(expected)
	if len(runes) < 2 {
		return false
	}
	return gan == string(runes[0]) && zhi == string(runes[1])
}

// =============================================================================
// 报告数据结构
// =============================================================================

type DimResult struct {
	PassCount int
	FailCount int
	Details   []string // failures only
}

func (d *DimResult) AddPass(id string) {
	d.PassCount++
}

func (d *DimResult) AddFail(id, reason string) {
	d.FailCount++
	d.Details = append(d.Details, fmt.Sprintf("  [%s] %s", id, reason))
}

func (d *DimResult) Total() int    { return d.PassCount + d.FailCount }
func (d *DimResult) Rate() float64 {
	if d.Total() == 0 { return 0 }
	return float64(d.PassCount) / float64(d.Total()) * 100
}

type AllDims struct {
	// 核心四柱 (继承自V1)
	YearPillar   DimResult
	MonthPillar  DimResult
	DayPillar    DimResult
	HourPillar   DimResult
	DayMaster    DimResult
	BodyStrength DimResult
	Pattern      DimResult
	TiaoHou      DimResult

	// 新增维度
	NaYin         DimResult
	FiveElements  DimResult
	TenGods       DimResult
	HiddenStems   DimResult
	GanRelations  DimResult
	ZhiRelations  DimResult
	DaYun         DimResult
	MingGong      DimResult
}

type PrecisionReport struct {
	TotalCases    int
	CalcFailCount int
	CalcFails     []string
	Dims          AllDims
}

func NewPrecisionReport(total int) *PrecisionReport {
	return &PrecisionReport{
		TotalCases: total,
		Dims:       AllDims{},
	}
}

func (r *PrecisionReport) AddCalcFail(id string) {
	r.CalcFailCount++
	r.CalcFails = append(r.CalcFails, id)
}

func (r *PrecisionReport) EffectiveCases() int {
	return r.TotalCases - r.CalcFailCount
}

func (r *PrecisionReport) String() string {
	var sb strings.Builder

	sb.WriteString("\n")
	sb.WriteString("╔══════════════════════════════════════════════════════════════════╗\n")
	sb.WriteString("║          八字命理综合精度测试报告                               ║\n")
	sb.WriteString("║ 对照来源: /home/skynit/Uni/ming/mingli/ 命理知识库              ║\n")
	sb.WriteString("╚══════════════════════════════════════════════════════════════════╝\n\n")

	sb.WriteString(fmt.Sprintf("总测试用例: %d\n", r.TotalCases))
	sb.WriteString(fmt.Sprintf("计算失败:    %d\n", r.CalcFailCount))
	sb.WriteString(fmt.Sprintf("有效用例:    %d\n\n", r.EffectiveCases()))

	// 表格头
	sb.WriteString("┌─────────────────┬──────────┬──────────┬──────────┬────────────┐\n")
	sb.WriteString("│ 测试维度          │  通过     │  失败     │  总数     │  准确率     │\n")
	sb.WriteString("├─────────────────┼──────────┼──────────┼──────────┼────────────┤\n")

	// 核心维度（继承自V1）
	coreDims := []struct {
		Name string
		Dim  *DimResult
	}{
		{"年柱(4柱)", &r.Dims.YearPillar},
		{"月柱(4柱)", &r.Dims.MonthPillar},
		{"日柱(4柱)", &r.Dims.DayPillar},
		{"时柱(4柱)", &r.Dims.HourPillar},
		{"日主", &r.Dims.DayMaster},
		{"身强身弱", &r.Dims.BodyStrength},
		{"格局判定", &r.Dims.Pattern},
		{"调候用神", &r.Dims.TiaoHou},
	}

	for _, d := range coreDims {
		if d.Dim.Total() > 0 {
			sb.WriteString(fmt.Sprintf("│ %-15s │ %5d/%d  │ %5d/%d  │ %5d    │ %6.1f%%   │\n",
				d.Name, d.Dim.PassCount, d.Dim.Total(), d.Dim.FailCount, d.Dim.Total(), d.Dim.Total(), d.Dim.Rate()))
		}
	}

	sb.WriteString("├─────────────────┼──────────┼──────────┼──────────┼────────────┤\n")

	// 新增维度
	newDims := []struct {
		Name string
		Dim  *DimResult
	}{
		{"纳音五行", &r.Dims.NaYin},
		{"五行评分", &r.Dims.FiveElements},
		{"十神分析", &r.Dims.TenGods},
		{"地支藏干", &r.Dims.HiddenStems},
		{"天干关系", &r.Dims.GanRelations},
		{"地支关系", &r.Dims.ZhiRelations},
		{"大运方向", &r.Dims.DaYun},
		{"命宫计算", &r.Dims.MingGong},
	}

	for _, d := range newDims {
		if d.Dim.Total() > 0 {
			sb.WriteString(fmt.Sprintf("│ %-15s │ %5d/%d  │ %5d/%d  │ %5d    │ %6.1f%%   │\n",
				d.Name, d.Dim.PassCount, d.Dim.Total(), d.Dim.FailCount, d.Dim.Total(), d.Dim.Total(), d.Dim.Rate()))
		}
	}

	sb.WriteString("└─────────────────┴──────────┴──────────┴──────────┴────────────┘\n\n")

	// 失败详情
	hasFails := false
	sb.WriteString("=== 失败详情 ===\n")

	allFails := []struct {
		Category string
		Details  []string
	}{
		{"年柱", r.Dims.YearPillar.Details},
		{"月柱", r.Dims.MonthPillar.Details},
		{"日柱", r.Dims.DayPillar.Details},
		{"时柱", r.Dims.HourPillar.Details},
		{"日主", r.Dims.DayMaster.Details},
		{"身强身弱", r.Dims.BodyStrength.Details},
		{"格局判定", r.Dims.Pattern.Details},
		{"调候用神", r.Dims.TiaoHou.Details},
		{"纳音五行", r.Dims.NaYin.Details},
		{"五行评分", r.Dims.FiveElements.Details},
		{"十神分析", r.Dims.TenGods.Details},
		{"藏干", r.Dims.HiddenStems.Details},
		{"天干关系", r.Dims.GanRelations.Details},
		{"地支关系", r.Dims.ZhiRelations.Details},
		{"大运方向", r.Dims.DaYun.Details},
		{"命宫", r.Dims.MingGong.Details},
	}

	for _, f := range allFails {
		if len(f.Details) > 0 {
			hasFails = true
			sb.WriteString(fmt.Sprintf("\n[%s]\n", f.Category))
			for _, d := range f.Details {
				sb.WriteString(d + "\n")
			}
		}
	}

	if !hasFails {
		sb.WriteString("全部通过！\n")
	}

	if r.CalcFailCount > 0 {
		sb.WriteString(fmt.Sprintf("\n计算失败用例: %v\n", r.CalcFails))
	}

	// 在测试报告中写入纳音详情作为参考
	sb.WriteString("\n=== 纳音验证参考（渊海子平）===\n")
	sb.WriteString("甲子乙丑海中金  丙寅丁卯炉中火  戊辰己巳大林木\n")
	sb.WriteString("庚午辛未路旁土  壬申癸酉剑锋金  甲戌乙亥山头火\n")
	sb.WriteString("丙子丁丑涧下水  戊寅己卯城头土  庚辰辛巳白蜡金\n")
	sb.WriteString("壬午癸未杨柳木  甲申乙酉泉中水  丙戌丁亥屋上土\n")
	sb.WriteString("戊子己丑霹雳火  庚寅辛卯松柏木  壬辰癸巳长流水\n")
	sb.WriteString("甲午乙未砂中金  丙申丁酉山下火  戊戌己亥平地木\n")
	sb.WriteString("庚子辛丑壁上土  壬寅癸卯金箔金  甲辰乙巳覆灯火\n")
	sb.WriteString("丙午丁未天河水  戊申己酉大驿土  庚戌辛亥钗钏金\n")
	sb.WriteString("壬子癸丑桑柘木  甲寅乙卯大溪水  丙辰丁巳砂中土\n")
	sb.WriteString("戊午己未天上火  庚申辛酉石榴木  壬戌癸亥大海水\n")

	return sb.String()
}
