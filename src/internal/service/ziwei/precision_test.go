package ziwei

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
)

// ZiWeiTestCase 来自 testdata/ziwei_cases.json
type ZiWeiTestCase struct {
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
		Pattern            string            `json:"pattern"`
		FiveBureau         string            `json:"five_bureau"`
		MainStar           string            `json:"main_star"`
		KeyPalaces         map[string]string `json:"key_palaces"`
		OccupationTendency string            `json:"occupation_tendency"`
		LifeTendency       string            `json:"life_tendency"`
		RankTendency       string            `json:"rank_tendency"`
		WealthTendency     string            `json:"wealth_tendency"`
		Tendency           string            `json:"tendency"`
	} `json:"expected"`
}

type ZiWeiTestData struct {
	Version string          `json:"version"`
	Cases   []ZiWeiTestCase `json:"cases"`
}

// ZiWeiAccuracyStats 紫微准确率统计
type ZiWeiAccuracyStats struct {
	Total               int
	Pattern             int
	KeyPalaces          int
	DetectedPatterns    int
}

// TestPrecisionZiWei 运行22+紫微命例的精度测试
func TestPrecisionZiWei(t *testing.T) {
	data, err := loadZiWeiTestData("../testdata/ziwei_cases.json")
	if err != nil {
		t.Fatalf("加载紫微测试数据失败: %v", err)
	}

	if len(data.Cases) < 20 {
		t.Logf("警告: 紫微测试数据不足20个，实际 %d 个", len(data.Cases))
	}

	stats := ZiWeiAccuracyStats{Total: len(data.Cases)}
	detailedResults := []string{}
	svc := NewZiWeiService()

	for _, tc := range data.Cases {
		chart, err := svc.CalculateChart(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, tc.Gender)
		if err != nil {
			t.Logf("[%s] %s: 计算失败: %v", tc.ID, tc.Name, err)
			continue
		}

		// 检测所有格局
		detectedPatterns := svc.DetectLocalPatterns(chart)
		detectedSet := make(map[string]bool)
		for _, p := range detectedPatterns {
			detectedSet[p] = true
		}

		// 校验预期格局是否被检测到
		if tc.Expected.Pattern != "" {
			if detectedSet[tc.Expected.Pattern] {
				stats.Pattern++
			} else {
				// 尝试模糊匹配
				matched := false
				for dp := range detectedSet {
					if strings.Contains(dp, tc.Expected.Pattern) || strings.Contains(tc.Expected.Pattern, dp) {
						matched = true
						break
					}
				}
				if matched {
					stats.Pattern++
				} else {
					detectedStr := strings.Join(detectedPatterns, ", ")
					detailedResults = append(detailedResults,
						fmt.Sprintf("[%s] 格局不符: 期望 %s, 实际 [%s]", tc.ID, tc.Expected.Pattern, detectedStr))
				}
			}
		}

		// 校验关键宫位
		if len(tc.Expected.KeyPalaces) > 0 {
			keyMatched := 0
			for palaceName, expectedStars := range(tc.Expected.KeyPalaces) {
				found := false
				for _, p := range chart.Palaces {
					if p.Name == palaceName {
						allStars := strings.Join(p.MainStars, ",") + "," + strings.Join(p.AuxStars, ",")
						if strings.Contains(allStars, expectedStars) || strings.Contains(expectedStars, allStars) {
							found = true
							break
						}
					}
				}
				if found {
					keyMatched++
				} else {
					detailedResults = append(detailedResults,
						fmt.Sprintf("[%s] 关键宫位不符: %s 期望含 %s", tc.ID, palaceName, expectedStars))
				}
			}
			stats.KeyPalaces += keyMatched
		}

		// 收集检测到的格局
		stats.DetectedPatterns += len(detectedPatterns)
	}

	// 报告
	report := fmt.Sprintf("\n=== 紫微精度测试报告 ===\n")
	report += fmt.Sprintf("总命例数: %d\n", stats.Total)
	report += fmt.Sprintf("格局检测准确率: %d/%d = %.1f%%\n", stats.Pattern, stats.Total, pct(stats.Pattern, stats.Total))
	report += fmt.Sprintf("检测到的总格局数: %d\n", stats.DetectedPatterns)
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
	reportFile, _ := os.Create("/tmp/ziwei_precision_report.txt")
	if reportFile != nil {
		defer reportFile.Close()
		reportFile.WriteString(report)
	}
}

// 辅助函数
func loadZiWeiTestData(path string) (*ZiWeiTestData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data ZiWeiTestData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return nil, err
	}
	return &data, nil
}

func pct(n, total int) float64 {
	if total == 0 {
		return 0
	}
	return float64(n) / float64(total) * 100
}
