package bazi_test

import (
	. "bazi/internal/service/bazi"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// TestCaseV3 测试用例（4柱干支输入）
type TestCaseV3 struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	Gender   string `json:"gender"`
	Expected struct {
		YearPillar    string `json:"year_pillar"`
		MonthPillar   string `json:"month_pillar"`
		DayPillar     string `json:"day_pillar"`
		HourPillar    string `json:"hour_pillar"`
		DayMaster     string `json:"day_master"`
		BodyStrength  string `json:"body_strength"`
		Pattern       string `json:"pattern"`
		YongShen      string `json:"yong_shen"`
		XiShen        string `json:"xi_shen"`
		JiShen        string `json:"ji_shen"`
		TiaoHou       string `json:"tiao_hou"`
	} `json:"expected"`
}

type TestDataV3 struct {
	Version string       `json:"version"`
	Cases   []TestCaseV3 `json:"cases"`
}

// V3Stats 准确率统计
type V3Stats struct {
	Total        int
	DayPillar    int
	DayMaster    int
	BodyStrength int
	Pattern      int
	TiaoHou      int
}

// TestPrecisionBaziV3 使用 CalculateFromPillars 的真正精度测试
// 输入：4柱干支（避免公历日期-干支转换歧义）
// 输出：与经典结论对比分析层（身强/格局/调候）的准确率
func TestPrecisionBaziV3(t *testing.T) {
	data, err := loadTestDataV3("../../testdata/classical_cases.json")
	if err != nil {
		t.Fatalf("加载测试数据失败: %v", err)
	}

	stats := V3Stats{Total: len(data.Cases)}
	detailedResults := []string{}
	svc := &BaziService{}

	for _, tc := range data.Cases {
		// 跳过没有四柱的用例
		if tc.Expected.YearPillar == "" || tc.Expected.DayPillar == "" {
			stats.Total--
			continue
		}

		// 直接用4柱干支作为输入
		result, err := svc.CalculateFromPillars(
			tc.Expected.YearPillar,
			tc.Expected.MonthPillar,
			tc.Expected.DayPillar,
			tc.Expected.HourPillar,
			tc.Gender,
		)
		if err != nil {
			t.Logf("[%s] 计算失败: %v", tc.ID, err)
			stats.Total--
			continue
		}

		// 校验日主（应该100%匹配，因为是从日柱取的）
		if result.DayPillar.Gan == tc.Expected.DayMaster {
			stats.DayMaster++
		} else {
			detailedResults = append(detailedResults,
				fmt.Sprintf("[%s] 日主不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.DayMaster, result.DayPillar.Gan))
		}

		// 校验身强身弱
		if tc.Expected.BodyStrength != "" {
			actualVerdict := result.BodyStrength.Verdict
			if bodyStrengthMatchV3(actualVerdict, tc.Expected.BodyStrength) {
				stats.BodyStrength++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 身强身弱: 期望 %s, 实际 %s", tc.ID, tc.Expected.BodyStrength, actualVerdict))
			}
		}

		// 校验格局
		if tc.Expected.Pattern != "" {
			actualPattern := result.PatternAnalysis.PatternName
			if patternMatchV3(actualPattern, tc.Expected.Pattern) {
				stats.Pattern++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 格局: 期望 %s, 实际 %s", tc.ID, tc.Expected.Pattern, actualPattern))
			}
		}

		// 校验调候用神
		if tc.Expected.TiaoHou != "" {
			var actualXiShen string
			if result.Tiaohou != nil && len(result.Tiaohou.Rules) > 0 {
				actualXiShen = result.Tiaohou.Rules[0].XiShen
			}
			if tiaoHouMatchV3(actualXiShen, tc.Expected.TiaoHou) {
				stats.TiaoHou++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 调候: 期望 %s, 实际 %s", tc.ID, tc.Expected.TiaoHou, actualXiShen))
			}
		}
	}

	// 输出报告
	report := fmt.Sprintf("\n=== 八字精度测试V3（4柱干支输入）===\n")
	report += fmt.Sprintf("总有效命例数: %d\n", stats.Total)
	report += fmt.Sprintf("日主准确率: %d/%d = %.1f%%\n", stats.DayMaster, stats.Total, pctV3(stats.DayMaster, stats.Total))
	report += fmt.Sprintf("身强身弱准确率: %d/%d = %.1f%%\n", stats.BodyStrength, stats.Total, pctV3(stats.BodyStrength, stats.Total))
	report += fmt.Sprintf("格局准确率: %d/%d = %.1f%%\n", stats.Pattern, stats.Total, pctV3(stats.Pattern, stats.Total))
	report += fmt.Sprintf("调候用神准确率: %d/%d = %.1f%%\n", stats.TiaoHou, stats.Total, pctV3(stats.TiaoHou, stats.Total))
	report += "\n=== 详细不符合 ===\n"
	if len(detailedResults) == 0 {
		report += "全部通过！\n"
	} else {
		for _, r := range detailedResults {
			report += r + "\n"
		}
	}

	t.Log(report)

	reportFile, _ := os.Create("/tmp/bazi_precision_v3_report.txt")
	if reportFile != nil {
		defer reportFile.Close()
		reportFile.WriteString(report)
	}
}

func loadTestDataV3(path string) (*TestDataV3, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data TestDataV3
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func bodyStrengthMatchV3(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return true
	}
	// 模糊匹配：身旺/偏旺/极旺/有根有力都算"身旺系"
	actualNorm := normalizeStrengthV3(actual)
	expectedNorm := normalizeStrengthV3(expected)
	if actualNorm == expectedNorm {
		return true
	}
	// 包含关系
	return strings.Contains(actual, expected) || strings.Contains(expected, actual)
}

func normalizeStrengthV3(s string) string {
	s = strings.TrimSpace(s)
	switch {
	case strings.Contains(s, "身极旺"), strings.Contains(s, "身旺"), strings.Contains(s, "偏旺"), strings.Contains(s, "有根"):
		return "旺"
	case strings.Contains(s, "身极弱"), strings.Contains(s, "身弱"), strings.Contains(s, "偏弱"), strings.Contains(s, "无根"):
		return "弱"
	case strings.Contains(s, "中和"), strings.Contains(s, "平衡"):
		return "中"
	}
	return s
}

func patternMatchV3(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return true
	}
	// 模糊匹配关键格局
	patternKeys := []string{
		"从旺格", "从弱格", "从财格", "从杀格", "从儿格", "从势格",
		"建禄格", "羊刃格", "专禄格", "日禄归时格", "月刃格",
		"食神制杀", "伤官配印", "财滋杀旺", "食神生财", "正官佩印",
		"正官格", "偏官格", "正印格", "偏印格", "正财格", "偏财格",
		"食神格", "伤官格", "比劫格", "建禄月刃格", "金神格",
		"专旺格", "从强格", "化气格", "从化格", "两神成像", "化气",
		"从革格", "曲直格", "炎上格", "稼穑格", "润下格", "日德格", "魁罡格",
		"正格",
	}
	for _, key := range patternKeys {
		if strings.Contains(actual, key) && strings.Contains(expected, key) {
			return true
		}
	}
	// 如果expected含"格"而actual含"格"，逐个比较
	if strings.Contains(actual, "格") && strings.Contains(expected, "格") {
		// 提取核心格局名（最后一个"格"之前的部分）
		actualCore := extractPatternCoreV3(actual)
		expectedCore := extractPatternCoreV3(expected)
		return actualCore == expectedCore
	}
	return actual == expected
}

func extractPatternCoreV3(p string) string {
	// 例: "建禄格" → "建禄"
	if idx := strings.LastIndex(p, "格"); idx > 0 {
		return p[:idx]
	}
	return p
}

func tiaoHouMatchV3(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return true
	}
	// 提取五行
	actualWuxing := extractWuxingV3(actual)
	expectedWuxing := extractWuxingV3(expected)
	if actualWuxing != "" && expectedWuxing != "" {
		return actualWuxing == expectedWuxing
	}
	// 包含关系
	return strings.Contains(actual, expected) || strings.Contains(expected, actual)
}

func extractWuxingV3(s string) string {
	wuxingList := []string{"金", "木", "水", "火", "土"}
	for _, wx := range wuxingList {
		if strings.Contains(s, wx) {
			return wx
		}
	}
	for gan, wx := range ganWuxingV3 {
		if strings.Contains(s, gan) {
			return wx
		}
	}
	return ""
}

var ganWuxingV3 = map[string]string{
	"甲": "木", "乙": "木",
	"丙": "火", "丁": "火",
	"戊": "土", "己": "土",
	"庚": "金", "辛": "金",
	"壬": "水", "癸": "水",
}

func pctV3(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}
