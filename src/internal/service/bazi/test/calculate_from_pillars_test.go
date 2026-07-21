package bazi_test

import (
	"bazi/internal/model"
	. "bazi/internal/service/bazi"
	"encoding/json"
	"fmt"
	"os"
	"testing"
)

// FromPillarsTestCase 来自日期候选 fixture，仅用于验证两个计算入口的一致性。
type FromPillarsTestCase struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Year   int    `json:"year"`
	Month  int    `json:"month"`
	Day    int    `json:"day"`
	Hour   int    `json:"hour"`
	Minute int    `json:"minute"`
	Gender string `json:"gender"`
}

// TestCalculateFromPillarsCase001 验证 case_001：四柱"壬辰 壬寅 甲寅 庚午" 身旺或偏旺
// 经典命例源自《滴天髓阐微》：甲木生于寅月（得令），壬水双透（印旺），庚金七杀被印制，
// 整体日主得令得势，本地 BodyStrength 分段候选预期为"身旺"或邻近的"偏旺"。
func TestCalculateFromPillarsCase001(t *testing.T) {
	svc := &BaziService{}
	result, err := svc.CalculateFromPillars("壬辰", "壬寅", "甲寅", "庚午", "MALE")
	if err != nil {
		t.Fatalf("CalculateFromPillars 失败: %v", err)
	}

	// 校验四柱
	if pillarStr(result.YearPillar) != "壬辰" {
		t.Errorf("年柱不符: 期望 壬辰, 实际 %s", pillarStr(result.YearPillar))
	}
	if pillarStr(result.MonthPillar) != "壬寅" {
		t.Errorf("月柱不符: 期望 壬寅, 实际 %s", pillarStr(result.MonthPillar))
	}
	if pillarStr(result.DayPillar) != "甲寅" {
		t.Errorf("日柱不符: 期望 甲寅, 实际 %s", pillarStr(result.DayPillar))
	}
	if pillarStr(result.HourPillar) != "庚午" {
		t.Errorf("时柱不符: 期望 庚午, 实际 %s", pillarStr(result.HourPillar))
	}

	// 校验日主
	if result.DayPillar.Gan != "甲" {
		t.Errorf("日主不符: 期望 甲, 实际 %s", result.DayPillar.Gan)
	}

	// 校验身强：case_001 预期为"身旺"，允许"偏旺"作为算法边界容差
	band := result.BodyStrength.ScoreBandCandidate
	t.Logf("case_001 本地身强分段候选: %s (总评分 %.3f, 得令 %.2f, 得地 %.2f, 得势 %.2f, 得生 %.2f)",
		band, result.BodyStrength.TotalScore,
		result.BodyStrength.LingScore, result.BodyStrength.DiScore,
		result.BodyStrength.ShiScore, result.BodyStrength.ShengScore)
	if band != "身旺" && band != "偏旺" {
		t.Errorf("本地身强分段候选不符: 期望 身旺/偏旺, 实际 %s", band)
	}

	// 校验十神
	if result.TenGods["day"] != "日主" {
		t.Errorf("日柱十神应为 日主, 实际 %s", result.TenGods["day"])
	}

	// 校验纳音
	if result.NaYin["day"].Name == "" {
		t.Error("日柱纳音应非空")
	}

	// 校验命宫
	if result.MingGong.GanZhi == "" {
		t.Error("命宫应非空")
	}

	// 校验大运因无日期应为空 / 默认
	if result.DaYunInfo.StartAge != 0 {
		t.Errorf("无日期大运起运年龄应默认为 0, 实际 %d", result.DaYunInfo.StartAge)
	}
	if len(result.DaYunInfo.Pillars) != 0 {
		t.Errorf("无日期大运柱应为空, 实际 %d 柱", len(result.DaYunInfo.Pillars))
	}

	// 调候结果应非空
	if result.Tiaohou == nil {
		t.Error("调候用神应非空")
	}

	t.Logf("命宫: %s, 调候表首候选: %s, 身强分段候选: %s",
		result.MingGong.GanZhi, result.Tiaohou.TablePrimaryCandidate, result.BodyStrength.ScoreBandCandidate)
}

// TestCalculateFromPillarsInvalidInput 校验入参解析错误
func TestCalculateFromPillarsInvalidInput(t *testing.T) {
	svc := &BaziService{}

	cases := []struct {
		desc                           string
		year, month, day, hour, gender string
	}{
		{"年柱长度不足", "壬", "壬寅", "甲寅", "庚午", "MALE"},
		{"年柱长度超长", "壬辰X", "壬寅", "甲寅", "庚午", "MALE"},
		{"年干非法", "X辰", "壬寅", "甲寅", "庚午", "MALE"},
		{"年支非法", "壬X", "壬寅", "甲寅", "庚午", "MALE"},
		{"月支非法", "壬辰", "壬X", "甲寅", "庚午", "MALE"},
		{"日干非法", "壬辰", "壬寅", "X寅", "庚午", "MALE"},
		{"时支非法", "壬辰", "壬寅", "甲寅", "庚X", "MALE"},
		{"性别非法", "壬辰", "壬寅", "甲寅", "庚午", "OTHER"},
		{"性别小写", "壬辰", "壬寅", "甲寅", "庚午", "male"}, // 小写应被接受
	}

	for _, tc := range cases {
		_, err := svc.CalculateFromPillars(tc.year, tc.month, tc.day, tc.hour, tc.gender)
		if tc.gender == "male" {
			// 小写应当被接受
			if err != nil {
				t.Errorf("[%s] 期望小写性别被接受, 实际错误: %v", tc.desc, err)
			}
			continue
		}
		if err == nil {
			t.Errorf("[%s] 期望错误, 实际未报错", tc.desc)
		}
	}
}

// TestCalculateFromPillarsConsistency 一致性测试：先从日期算出四柱，再从四柱算出 BaziResult，
// 两个入口应得到相同的四柱、身强判定、十神、五行总分等核心分析结果。
// 使用日期先回算干支再传入，避免硬编码"期望干支"与实际公历-干支转换不一致的干扰。
func TestCalculateFromPillarsConsistency(t *testing.T) {
	data, err := loadFromPillarsTestData("../../testdata/bazi_date_gold_candidates.json")
	if err != nil {
		t.Fatalf("加载测试数据失败: %v", err)
	}
	if len(data) == 0 {
		t.Skip("测试数据为空")
	}

	svc := &BaziService{}
	stats := struct {
		total, pillarMatch, bodyMatch, tgMatch, feMatch int
	}{}

	for _, tc := range data {
		dateResult, err := svc.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Minute, tc.Gender)
		if err != nil {
			t.Logf("[%s] %s: 日期计算失败: %v", tc.ID, tc.Name, err)
			continue
		}
		stats.total++

		pillarResult, err := svc.CalculateFromPillars(
			pillarStr(dateResult.YearPillar),
			pillarStr(dateResult.MonthPillar),
			pillarStr(dateResult.DayPillar),
			pillarStr(dateResult.HourPillar),
			tc.Gender,
		)
		if err != nil {
			t.Errorf("[%s] %s: 干支计算失败: %v", tc.ID, tc.Name, err)
			continue
		}

		// 四柱必须完全一致
		ok := true
		for _, pos := range []struct {
			name   string
			date   model.Pillar
			pillar model.Pillar
		}{
			{"年", dateResult.YearPillar, pillarResult.YearPillar},
			{"月", dateResult.MonthPillar, pillarResult.MonthPillar},
			{"日", dateResult.DayPillar, pillarResult.DayPillar},
			{"时", dateResult.HourPillar, pillarResult.HourPillar},
		} {
			if pos.date != pos.pillar {
				t.Errorf("[%s] %s柱不一致: date=%s pillar=%s", tc.ID, pos.name, pillarStr(pos.date), pillarStr(pos.pillar))
				ok = false
			}
		}
		if ok {
			stats.pillarMatch++
		}

		// 本地评分及分段候选应一致
		if dateResult.BodyStrength.ScoreBandCandidate != pillarResult.BodyStrength.ScoreBandCandidate {
			t.Errorf("[%s] 身强分段候选不一致: date=%s pillar=%s", tc.ID, dateResult.BodyStrength.ScoreBandCandidate, pillarResult.BodyStrength.ScoreBandCandidate)
		} else {
			stats.bodyMatch++
		}

		// 十神应一致
		tgOK := true
		for _, k := range []string{"year", "month", "hour"} {
			if dateResult.TenGods[k] != pillarResult.TenGods[k] {
				t.Errorf("[%s] %s柱十神不一致: date=%s pillar=%s", tc.ID, k, dateResult.TenGods[k], pillarResult.TenGods[k])
				tgOK = false
			}
		}
		if tgOK {
			stats.tgMatch++
		}

		// 五行总分应一致
		feOK := true
		for elem, score := range dateResult.FiveElements {
			if pillarResult.FiveElements[elem] != score {
				t.Errorf("[%s] 五行 %s 不一致: date=%d pillar=%d", tc.ID, elem, score, pillarResult.FiveElements[elem])
				feOK = false
			}
		}
		if feOK {
			stats.feMatch++
		}
	}

	t.Logf("一致性统计：总 %d 例 | 四柱一致 %d | 身强一致 %d | 十神一致 %d | 五行一致 %d",
		stats.total, stats.pillarMatch, stats.bodyMatch, stats.tgMatch, stats.feMatch)
	if stats.pillarMatch != stats.total {
		t.Errorf("四柱一致性未达 100%%")
	}
	if stats.bodyMatch != stats.total {
		t.Errorf("身强判定一致性未达 100%%")
	}
	if stats.feMatch != stats.total {
		t.Errorf("五行总分一致性未达 100%%")
	}
}

func loadFromPillarsTestData(path string) ([]FromPillarsTestCase, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取 %s: %w", path, err)
	}
	var wrapper struct {
		Cases []json.RawMessage `json:"cases"`
	}
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, err
	}
	var out []FromPillarsTestCase
	for _, raw := range wrapper.Cases {
		var tc FromPillarsTestCase
		if err := json.Unmarshal(raw, &tc); err != nil {
			return nil, err
		}
		out = append(out, tc)
	}
	return out, nil
}
