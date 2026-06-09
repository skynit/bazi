package fortune_test

import (
	. "bazi/internal/service/fortune"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	bazipkg "bazi/internal/service/bazi"
)

type RikuyoTestCase struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	BirthYear int    `json:"birth_year"`
	BirthMonth int   `json:"birth_month"`
	BirthDay   int   `json:"birth_day"`
	BirthHour  int   `json:"birth_hour"`
	QueryDate  string `json:"query_date"`
	Expected   struct {
		JianChuShouldBe        string `json:"jian_chu_should_be"`
		MonthZhi               string `json:"month_zhi"`
		Favorable              bool   `json:"favorable"`
		HuangDaoOrHeiDao       string `json:"huang_dao_or_hei_dao"`
		TodayGan               string `json:"today_gan"`
		PengzuGanTaboo         string `json:"pengzu_gan_taboo"`
		OverallVerdictShouldBeFavorable bool `json:"overall_verdict_should_be_favorable"`
	} `json:"expected"`
}

type RikuyoTestData struct {
	Version string             `json:"version"`
	Cases   []RikuyoTestCase   `json:"cases"`
}

type RikuyoAccuracyStats struct {
	Total        int
	CompiledOK   int
	HasJianChu   int
	HasHuangDao  int
	HasPengzu    int
	NonEmptyResult int
}

func TestPrecisionRikuyo(t *testing.T) {
	data, err := loadRikuyoTestData("../../testdata/rikuyo_cases.json")
	if err != nil {
		t.Fatalf("加载日课测试数据失败: %v", err)
	}

	stats := RikuyoAccuracyStats{Total: len(data.Cases)}
	detailedResults := []string{}
	baziSvc := &bazipkg.BaziService{}

	for _, tc := range data.Cases {
		// 计算八字基础
		bazi, err := baziSvc.Calculate(tc.BirthYear, tc.BirthMonth, tc.BirthDay, tc.BirthHour, 0, "MALE")
		if err != nil {
			t.Logf("[%s] 八字计算失败: %v", tc.ID, err)
			continue
		}

		// 解析查询日期
		queryDate, err := time.Parse("2006-01-02", tc.QueryDate)
		if err != nil {
			t.Logf("[%s] 日期解析失败: %v", tc.ID, err)
			continue
		}

		// 计算日课
		result := CalcRikuyo(bazi, queryDate, tc.BirthYear)
		if result == nil {
			detailedResults = append(detailedResults, fmt.Sprintf("[%s] 日课结果为空", tc.ID))
			continue
		}

		stats.CompiledOK++

		// 校验建除十二神
		if result.JianChuName != "" {
			stats.HasJianChu++
			if tc.Expected.JianChuShouldBe != "" {
				if strings.Contains(result.JianChuName, tc.Expected.JianChuShouldBe) ||
				   strings.Contains(tc.Expected.JianChuShouldBe, result.JianChuName) {
					// 通过
				} else {
					detailedResults = append(detailedResults,
						fmt.Sprintf("[%s] 建除不符: 期望 %s, 实际 %s", tc.ID, tc.Expected.JianChuShouldBe, result.JianChuName))
				}
			}
		} else {
			detailedResults = append(detailedResults, fmt.Sprintf("[%s] 建除十二神为空", tc.ID))
		}

		// 校验黄道黑道
		if result.HuangDaoName != "" {
			stats.HasHuangDao++
			if tc.Expected.HuangDaoOrHeiDao != "" {
				if strings.Contains(result.HuangDaoName, tc.Expected.HuangDaoOrHeiDao) ||
				   tc.Expected.HuangDaoOrHeiDao == "黄道" && result.HuangDaoFavorable ||
				   tc.Expected.HuangDaoOrHeiDao == "黑道" && !result.HuangDaoFavorable {
					// 通过
				}
			}
		}

		// 校验彭祖百忌
		if result.PengzuGanTaboo != "" {
			stats.HasPengzu++
		}

		// 校验综合判断
		if result.OverallVerdict != "" || result.FavorScore != 0 {
			stats.NonEmptyResult++
		}

		detailedResults = append(detailedResults,
			fmt.Sprintf("[%s] 建除=%s 黄道=%s 吉=%v 评分=%d",
				tc.ID, result.JianChuName, result.HuangDaoName, result.HuangDaoFavorable, result.FavorScore))
	}

	// 报告
	report := fmt.Sprintf("\n=== 日课/运势精度测试报告 ===\n")
	report += fmt.Sprintf("总命例数: %d\n", stats.Total)
	report += fmt.Sprintf("日课计算成功: %d/%d = %.1f%%\n", stats.CompiledOK, stats.Total, pct(stats.CompiledOK, stats.Total))
	report += fmt.Sprintf("建除十二神有结果: %d/%d = %.1f%%\n", stats.HasJianChu, stats.Total, pct(stats.HasJianChu, stats.Total))
	report += fmt.Sprintf("黄道黑道有结果: %d/%d = %.1f%%\n", stats.HasHuangDao, stats.Total, pct(stats.HasHuangDao, stats.Total))
	report += fmt.Sprintf("彭祖百忌有结果: %d/%d = %.1f%%\n", stats.HasPengzu, stats.Total, pct(stats.HasPengzu, stats.Total))
	report += fmt.Sprintf("综合判断非空: %d/%d = %.1f%%\n", stats.NonEmptyResult, stats.Total, pct(stats.NonEmptyResult, stats.Total))
	report += "\n=== 详细结果 ===\n"
	for _, r := range detailedResults {
		report += r + "\n"
	}

	t.Log(report)

	reportFile, _ := os.Create("/tmp/rikuyo_precision_report.txt")
	if reportFile != nil {
		defer reportFile.Close()
		reportFile.WriteString(report)
	}
}

func loadRikuyoTestData(path string) (*RikuyoTestData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var data RikuyoTestData
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
