package fortune_test

import (
	. "bazi/internal/service/fortune"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	bazipkg "bazi/internal/service/bazi"
)

type RikuyoTestCase struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Source     string `json:"source"`
	BirthYear  int    `json:"birth_year"`
	BirthMonth int    `json:"birth_month"`
	BirthDay   int    `json:"birth_day"`
	BirthHour  int    `json:"birth_hour"`
	QueryDate  string `json:"query_date"`
}

type RikuyoTestData struct {
	Version string           `json:"version"`
	Cases   []RikuyoTestCase `json:"cases"`
}

type RikuyoEvidenceStats struct {
	Total         int
	Calculated    int
	TwelveStageOK int
	JianChuOK     int
	HuangDaoOK    int
}

func TestRikuyoEvidenceCompleteness(t *testing.T) {
	data, err := loadRikuyoTestData("../../testdata/rikuyo_cases.json")
	if err != nil {
		t.Fatalf("加载日课测试数据失败: %v", err)
	}

	stats := RikuyoEvidenceStats{Total: len(data.Cases)}
	details := make([]string, 0, len(data.Cases))
	baziSvc := &bazipkg.BaziService{}

	for _, tc := range data.Cases {
		chart, err := baziSvc.Calculate(tc.BirthYear, tc.BirthMonth, tc.BirthDay, tc.BirthHour, 0, "MALE")
		if err != nil {
			t.Errorf("[%s] 八字计算失败: %v", tc.ID, err)
			continue
		}
		queryDate, err := time.Parse("2006-01-02", tc.QueryDate)
		if err != nil {
			t.Errorf("[%s] 日期解析失败: %v", tc.ID, err)
			continue
		}

		result := CalcRikuyo(chart, queryDate)
		if result == nil {
			t.Errorf("[%s] 日课结果为空", tc.ID)
			continue
		}
		stats.Calculated++

		if result.TwelveStage.RuleID == "rikuyo.twelve-stage-v1" &&
			result.TwelveStage.ReferenceStem == chart.DayPillar.Gan &&
			result.TwelveStage.QueryBranch != "" &&
			result.TwelveStage.Name != "" &&
			isObservedUnadjudicated(result.TwelveStage.Status, result.TwelveStage.InterpretationStatus) {
			stats.TwelveStageOK++
		} else {
			t.Errorf("[%s] 十二长生证据不完整: %+v", tc.ID, result.TwelveStage)
		}

		if result.JianChu.RuleID == "rikuyo.jianchu-month-branch-v1" &&
			result.JianChu.MonthBranch != "" &&
			result.JianChu.QueryBranch == result.TwelveStage.QueryBranch &&
			result.JianChu.Name != "" &&
			isObservedUnadjudicated(result.JianChu.Status, result.JianChu.InterpretationStatus) {
			stats.JianChuOK++
		} else {
			t.Errorf("[%s] 建除证据不完整: %+v", tc.ID, result.JianChu)
		}

		if result.HuangDao.RuleID == "rikuyo.twelve-star.tyme4go-v2" &&
			result.HuangDao.MonthBranch == result.JianChu.MonthBranch &&
			result.HuangDao.QueryBranch == result.TwelveStage.QueryBranch &&
			result.HuangDao.Name != "" &&
			isObservedUnadjudicated(result.HuangDao.Status, result.HuangDao.InterpretationStatus) {
			stats.HuangDaoOK++
		} else {
			t.Errorf("[%s] 值神证据不完整: %+v", tc.ID, result.HuangDao)
		}

		details = append(details, fmt.Sprintf(
			"[%s] 十二长生=%s 建除=%s 值神=%s 月支=%s 日支=%s",
			tc.ID,
			result.TwelveStage.Name,
			result.JianChu.Name,
			result.HuangDao.Name,
			result.JianChu.MonthBranch,
			result.JianChu.QueryBranch,
		))
	}

	report := fmt.Sprintf("\n=== 日课结构证据完整性报告 ===\n")
	report += "说明: Bronze 夹具未独立裁决，本测试不发布传统解释准确率。\n"
	report += fmt.Sprintf("总命例数: %d\n", stats.Total)
	report += fmt.Sprintf("计算成功: %d/%d = %.1f%%\n", stats.Calculated, stats.Total, pct(stats.Calculated, stats.Total))
	report += fmt.Sprintf("十二长生证据完整: %d/%d = %.1f%%\n", stats.TwelveStageOK, stats.Total, pct(stats.TwelveStageOK, stats.Total))
	report += fmt.Sprintf("建除证据完整: %d/%d = %.1f%%\n", stats.JianChuOK, stats.Total, pct(stats.JianChuOK, stats.Total))
	report += fmt.Sprintf("值神证据完整: %d/%d = %.1f%%\n", stats.HuangDaoOK, stats.Total, pct(stats.HuangDaoOK, stats.Total))
	for _, detail := range details {
		report += detail + "\n"
	}
	t.Log(report)

	if err := os.WriteFile("/tmp/rikuyo_precision_report.txt", []byte(report), 0o600); err != nil {
		t.Logf("写入临时报告失败: %v", err)
	}
}

func isObservedUnadjudicated(status, interpretationStatus string) bool {
	return status == "observed" && interpretationStatus == "not_adjudicated"
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
