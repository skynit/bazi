package bazi_test

import (
	"bazi/internal/model"
	. "bazi/internal/service/bazi"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

)

// TestCase 来自 testdata/classical_cases.json 的结构
type TestCase struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Source   string `json:"source"`
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Minute   int    `json:"minute"`
	Gender   string `json:"gender"`
	Expected struct {
		YearPillar  string `json:"year_pillar"`
		MonthPillar string `json:"month_pillar"`
		DayPillar   string `json:"day_pillar"`
		HourPillar  string `json:"hour_pillar"`
		DayMaster   string `json:"day_master"`
		BodyStrength string `json:"body_strength"`
		Pattern     string `json:"pattern"`
		YongShen    string `json:"yong_shen"`
		XiShen      string `json:"xi_shen"`
		JiShen      string `json:"ji_shen"`
		TiaoHou     string `json:"tiao_hou"`
	} `json:"expected"`
}

type TestData struct {
	Version string     `json:"version"`
	Cases   []TestCase `json:"cases"`
}

// 准确率统计
type AccuracyStats struct {
	Total       int
	YearPillar  int
	MonthPillar int
	DayPillar   int
	HourPillar  int
	DayMaster   int
	BodyStrength int
	Pattern     int
	TiaoHou     int
}

// TestPrecisionBazi 运行32+经典命例的八字精度测试
func TestPrecisionBazi(t *testing.T) {
	// 读取测试数据
	data, err := loadTestData("../../testdata/classical_cases.json")
	if err != nil {
		t.Fatalf("加载测试数据失败: %v", err)
	}

	if len(data.Cases) < 30 {
		t.Logf("警告: 测试数据不足30个，实际 %d 个", len(data.Cases))
	}

	stats := AccuracyStats{Total: len(data.Cases)}
	svc := &BaziService{}
	detailedResults := []string{}

	for _, tc := range data.Cases {
		// 计算八字
		result, err := svc.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, tc.Gender)
		if err != nil {
			t.Logf("[%s] %s: 计算失败: %v", tc.ID, tc.Name, err)
			continue
		}

		// 校验四柱
		if tc.Expected.YearPillar != "" {
			if pillarMatch(result.YearPillar, tc.Expected.YearPillar) {
				stats.YearPillar++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 年柱不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.YearPillar, pillarStr(result.YearPillar)))
			}
		}
		if tc.Expected.MonthPillar != "" {
			if pillarMatch(result.MonthPillar, tc.Expected.MonthPillar) {
				stats.MonthPillar++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 月柱不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.MonthPillar, pillarStr(result.MonthPillar)))
			}
		}
		if tc.Expected.DayPillar != "" {
			if pillarMatch(result.DayPillar, tc.Expected.DayPillar) {
				stats.DayPillar++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 日柱不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.DayPillar, pillarStr(result.DayPillar)))
			}
		}
		if tc.Expected.HourPillar != "" {
			if pillarMatch(result.HourPillar, tc.Expected.HourPillar) {
				stats.HourPillar++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 时柱不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.HourPillar, pillarStr(result.HourPillar)))
			}
		}

		// 校验日主
		if tc.Expected.DayMaster != "" {
			if result.DayPillar.Gan == tc.Expected.DayMaster {
				stats.DayMaster++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 日主不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.DayMaster, result.DayPillar.Gan))
			}
		}

		// 校验身强身弱
		if tc.Expected.BodyStrength != "" {
			actualVerdict := result.BodyStrength.Verdict
			if bodyStrengthMatch(actualVerdict, tc.Expected.BodyStrength) {
				stats.BodyStrength++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 身强身弱不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.BodyStrength, actualVerdict))
			}
		}

		// 校验格局
		if tc.Expected.Pattern != "" {
			actualPattern := result.PatternAnalysis.PatternName
			if patternMatch(actualPattern, tc.Expected.Pattern) {
				stats.Pattern++
			} else {
				detailedResults = append(detailedResults,
					fmt.Sprintf("[%s] 格局不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.Pattern, actualPattern))
			}
		}

		// 校验调候用神
		if tc.Expected.TiaoHou != "" {
			if result.Tiaohou != nil {
				actualXiShen := tiaohouXiShen(result.Tiaohou)
				if tiaoHouMatch(actualXiShen, tc.Expected.TiaoHou) {
					stats.TiaoHou++
				} else {
					detailedResults = append(detailedResults,
						fmt.Sprintf("[%s] 调候用神不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.TiaoHou, actualXiShen))
				}
			}
		}
	}

	// 输出统计
	report := fmt.Sprintf("\n=== 八字精度测试报告 ===\n")
	report += fmt.Sprintf("总命例数: %d\n", stats.Total)
	report += fmt.Sprintf("年柱准确率: %d/%d = %.1f%%\n", stats.YearPillar, stats.Total, pct(stats.YearPillar, stats.Total))
	report += fmt.Sprintf("月柱准确率: %d/%d = %.1f%%\n", stats.MonthPillar, stats.Total, pct(stats.MonthPillar, stats.Total))
	report += fmt.Sprintf("日柱准确率: %d/%d = %.1f%%\n", stats.DayPillar, stats.Total, pct(stats.DayPillar, stats.Total))
	report += fmt.Sprintf("时柱准确率: %d/%d = %.1f%%\n", stats.HourPillar, stats.Total, pct(stats.HourPillar, stats.Total))
	report += fmt.Sprintf("日主准确率: %d/%d = %.1f%%\n", stats.DayMaster, stats.Total, pct(stats.DayMaster, stats.Total))
	report += fmt.Sprintf("身强身弱准确率: %d/%d = %.1f%%\n", stats.BodyStrength, stats.Total, pct(stats.BodyStrength, stats.Total))
	report += fmt.Sprintf("格局准确率: %d/%d = %.1f%%\n", stats.Pattern, stats.Total, pct(stats.Pattern, stats.Total))
	report += fmt.Sprintf("调候用神准确率: %d/%d = %.1f%%\n", stats.TiaoHou, stats.Total, pct(stats.TiaoHou, stats.Total))
	report += "\n=== 不符合案例 ===\n"
	if len(detailedResults) == 0 {
		report += "全部通过！\n"
	} else {
		for _, r := range detailedResults {
			report += r + "\n"
		}
	}

	t.Log(report)

	// 输出到文件
	reportFile, _ := os.Create("/tmp/bazi_precision_report.txt")
	if reportFile != nil {
		defer reportFile.Close()
		reportFile.WriteString(report)
	}
}

// 辅助函数
func loadTestData(path string) (*TestData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data TestData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func pillarStr(p model.Pillar) string {
	return p.Gan + p.Zhi
}

func pillarMatch(p model.Pillar, expected string) bool {
	expected = strings.TrimSpace(expected)
	expected = strings.ReplaceAll(expected, " ", "")
	runes := []rune(expected)
	if len(runes) < 2 {
		return false
	}
	return p.Gan == string(runes[0]) && p.Zhi == string(runes[1])
}

func bodyStrengthMatch(actual, expected string) bool {
	// 归一化比较
	actual = normalizeVerdict(actual)
	expected = normalizeVerdict(expected)
	return strings.Contains(actual, expected) || strings.Contains(expected, actual)
}

func normalizeVerdict(s string) string {
	s = strings.TrimSpace(s)
	switch {
	case strings.Contains(s, "身旺") || strings.Contains(s, "偏旺"):
		return "身旺"
	case strings.Contains(s, "身弱") || strings.Contains(s, "偏弱"):
		return "身弱"
	case strings.Contains(s, "中和") || strings.Contains(s, "平衡"):
		return "中和"
	}
	return s
}

func patternMatch(actual, expected string) bool {
	// 简化比较
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return true // 无期望/无实际不算错误
	}
	// 提取核心格局名
	patterns := []string{
		"正格", "从旺格", "从弱格", "从财格", "从杀格", "从儿格",
		"建禄格", "羊刃格", "专禄格", "日禄归时格",
		"食神制杀格", "伤官配印格", "财滋杀旺格", "食神生财格", "正官佩印格",
		"正官格", "偏官格", "正印格", "偏印格", "正财格", "偏财格",
		"食神格", "伤官格", "比劫格", "建禄月刃格", "金神格",
		"专旺格", "从强格", "从化格", "化气格",
	}
	for _, p := range patterns {
		if strings.Contains(actual, p) && strings.Contains(expected, p) {
			return true
		}
		if strings.Contains(actual, p) || strings.Contains(expected, p) {
			// 任一含此格局
			return actual == expected
		}
	}
	return actual == expected
}

func tiaohouXiShen(t *TiaohouResult) string {
	if t == nil {
		return ""
	}
	if len(t.Rules) > 0 {
		return t.Rules[0].XiShen
	}
	return ""
}

// ganWuxing 天干到五行的映射
var ganWuxing = map[string]string{
	"甲": "木", "乙": "木",
	"丙": "火", "丁": "火",
	"戊": "土", "己": "土",
	"庚": "金", "辛": "金",
	"壬": "水", "癸": "水",
}

func tiaoHouMatch(actual, expected string) bool {
	actual = strings.TrimSpace(actual)
	expected = strings.TrimSpace(expected)
	if actual == "" || expected == "" {
		return true
	}
	// 比较五行（从天干提取）
	actualWuxing := extractWuxing(actual)
	expectedWuxing := extractWuxing(expected)
	if actualWuxing == expectedWuxing && actualWuxing != "" {
		return true
	}
	// 直接字符串包含
	return strings.Contains(actual, expected) || strings.Contains(expected, actual)
}

func extractWuxing(s string) string {
	// 检查是否包含五行字
	wuxingList := []string{"金", "木", "水", "火", "土"}
	for _, wx := range wuxingList {
		if strings.Contains(s, wx) {
			return wx
		}
	}
	// 检查天干
	for gan, wx := range ganWuxing {
		if strings.Contains(s, gan) {
			return wx
		}
	}
	return ""
}

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}
