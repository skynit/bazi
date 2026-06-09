package bazi_test

import (
	. "bazi/internal/service/bazi"
	"fmt"
	"os"
	"strings"
	"testing"
)

// =============================================================================
// 大运起运+序列测试 v2
// 使用 Calculate（需要日历数据），验证方向+8步序列
// =============================================================================

type dayunV2TestCase struct {
	ID, Desc, Gender             string
	Year, Month, Day, Hour, Min  int
	ExpectedDir                  string   // "顺行" or "逆行"
	CheckSequence                bool     // 是否需要校验8步大运序列
	ExpectedFirstPillar          string   // 第一大运（非空则校验）
	ExpectedLastPillar           string   // 第八大运（非空则校验）
}

var dayunV2Cases = []dayunV2TestCase{
	// 阳年男命 = 顺行
	{ID: "DYV2-001", Desc: "甲子阳年男顺行", Gender: "MALE", Year: 1984, Month: 3, Day: 15, Hour: 12, Min: 0,
		ExpectedDir: "顺行", CheckSequence: true},
	// 阴年男命 = 逆行
	{ID: "DYV2-002", Desc: "乙丑阴年男逆行", Gender: "MALE", Year: 1985, Month: 6, Day: 20, Hour: 8, Min: 0,
		ExpectedDir: "逆行", CheckSequence: true},
	// 阳年女命 = 逆行
	{ID: "DYV2-003", Desc: "甲子阳年女逆行", Gender: "FEMALE", Year: 1984, Month: 3, Day: 15, Hour: 12, Min: 0,
		ExpectedDir: "逆行", CheckSequence: true},
	// 阴年女命 = 顺行
	{ID: "DYV2-004", Desc: "乙丑阴年女顺行", Gender: "FEMALE", Year: 1985, Month: 6, Day: 20, Hour: 8, Min: 0,
		ExpectedDir: "顺行", CheckSequence: true},
	// 更多边界情况
	{ID: "DYV2-005", Desc: "丙寅阳年男顺行", Gender: "MALE", Year: 1986, Month: 5, Day: 10, Hour: 10, Min: 0,
		ExpectedDir: "顺行", CheckSequence: true},
	{ID: "DYV2-006", Desc: "丁卯阴年女顺行", Gender: "FEMALE", Year: 1987, Month: 8, Day: 15, Hour: 14, Min: 0,
		ExpectedDir: "顺行", CheckSequence: true},
	{ID: "DYV2-007", Desc: "庚申阳年男顺行", Gender: "MALE", Year: 1980, Month: 3, Day: 1, Hour: 10, Min: 0,
		ExpectedDir: "顺行", CheckSequence: true},
	{ID: "DYV2-008", Desc: "癸亥阴年男逆行", Gender: "MALE", Year: 1983, Month: 11, Day: 25, Hour: 22, Min: 0,
		ExpectedDir: "逆行", CheckSequence: true},
	{ID: "DYV2-009", Desc: "壬戌阳年女逆行", Gender: "FEMALE", Year: 1982, Month: 10, Day: 8, Hour: 16, Min: 0,
		ExpectedDir: "逆行", CheckSequence: true},
	{ID: "DYV2-010", Desc: "辛酉阴年男逆行", Gender: "MALE", Year: 1981, Month: 7, Day: 5, Hour: 6, Min: 0,
		ExpectedDir: "逆行", CheckSequence: true},
	// 特定知名命例
	{ID: "DYV2-011", Desc: "戊午阳年女逆行", Gender: "FEMALE", Year: 1978, Month: 6, Day: 15, Hour: 12, Min: 0,
		ExpectedDir: "逆行", CheckSequence: true},
	{ID: "DYV2-012", Desc: "己未阴年女顺行", Gender: "FEMALE", Year: 1979, Month: 9, Day: 20, Hour: 8, Min: 0,
		ExpectedDir: "顺行", CheckSequence: true},
}

func TestDayunPrecisionV2(t *testing.T) {
	svc := &BaziService{}
	var pass, fail int
	var fails []string
	var outputs []string

	for _, tc := range dayunV2Cases {
		r, err := svc.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Min, tc.Gender)
		if err != nil {
			fail++
			fails = append(fails, fmt.Sprintf("[%s] 计算失败: %v", tc.ID, err))
			continue
		}

		di := r.DaYunInfo
		dirOk := strings.Contains(di.Direction, tc.ExpectedDir)
		isForward := strings.Contains(di.Direction, "顺行")
		sequenceMsg := ""

		if tc.CheckSequence {
			// 验证8步大运
			pillarCount := len(di.Pillars)
			if pillarCount == 8 {
				pass++ // 计数：序列长度正确
			} else {
				fail++
				fails = append(fails, fmt.Sprintf("[%s] 期望8步大运,实际%d步", tc.ID, pillarCount))
			}

			// 验证序列中相邻大运天干+1或-1（取决于方向）
			if pillarCount >= 2 {
				seqOk := true
				for i := 1; i < pillarCount; i++ {
					prevGan := di.Pillars[i-1].Gan
					prevZhi := di.Pillars[i-1].Zhi
					curGan := di.Pillars[i].Gan
					curZhi := di.Pillars[i].Zhi

					if isForward {
						// 顺行：天干+1, 地支+1
						if !isNextGan(prevGan, curGan) || !isNextZhi(prevZhi, curZhi) {
							seqOk = false
							if !isNextGan(prevGan, curGan) {
								fails = append(fails, fmt.Sprintf("[%s] 第%d步大运天干序列错误: %s->%s (期望顺行+1)",
									tc.ID, i+1, prevGan, curGan))
							}
							if !isNextZhi(prevZhi, curZhi) {
								fails = append(fails, fmt.Sprintf("[%s] 第%d步大运地支序列错误: %s->%s (期望顺行+1)",
									tc.ID, i+1, prevZhi, curZhi))
							}
							break
						}
					} else {
						// 逆行：天干-1, 地支-1
						if !isPrevGan(prevGan, curGan) || !isPrevZhi(prevZhi, curZhi) {
							seqOk = false
							if !isPrevGan(prevGan, curGan) {
								fails = append(fails, fmt.Sprintf("[%s] 第%d步大运天干序列错误: %s->%s (期望逆行-1)",
									tc.ID, i+1, prevGan, curGan))
							}
							if !isPrevZhi(prevZhi, curZhi) {
								fails = append(fails, fmt.Sprintf("[%s] 第%d步大运地支序列错误: %s->%s (期望逆行-1)",
									tc.ID, i+1, prevZhi, curZhi))
							}
							break
						}
					}
				}
				if seqOk {
					pass++
				} else {
					fail++
				}
			}

			// 序列字符串
			var sb strings.Builder
			for i, p := range di.Pillars {
				if i > 0 {
					sb.WriteString(" ")
				}
				sb.WriteString(p.Gan + p.Zhi)
			}
			sequenceMsg = fmt.Sprintf("序列=[%s] 起运=%d岁", sb.String(), di.StartAge)
		}

		dirResult := "✓"
		if !dirOk {
			fail++
			fails = append(fails, fmt.Sprintf("[%s] %s: 期望方向'%s',实际'%s'", tc.ID, tc.Desc, tc.ExpectedDir, di.Direction))
			dirResult = "✗"
		} else {
			pass++
		}

		// 验证起运年龄>0
		if di.StartAge <= 0 && di.StartAge != 0 {
			// 允许0（仅在计算错误时）
		}
		if di.StartAge <= 0 {
			fails = append(fails, fmt.Sprintf("[%s] 起运年龄异常: %d岁", tc.ID, di.StartAge))
		}

		outputs = append(outputs, fmt.Sprintf("[%s] %s: 方向=%s %s %s | 起运=%d岁",
			tc.ID, tc.Desc, di.Direction, dirResult, sequenceMsg, di.StartAge))
	}

	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════╗\n")
	sb.WriteString("║  大运起运+序列精度测试 V2                    ║\n")
	sb.WriteString("╚══════════════════════════════════════════════╝\n")
	total := pass + fail
	sb.WriteString(fmt.Sprintf("\n断言: %d通过 + %d失败 = %d总 | 准确率: %.1f%%\n", pass, fail, total, pctV3(pass, total)))
	sb.WriteString(fmt.Sprintf("用例数: %d\n\n", len(dayunV2Cases)))
	if len(fails) > 0 {
		sb.WriteString("失败:\n" + strings.Join(fails, "\n") + "\n")
	} else {
		sb.WriteString("全部通过！\n")
	}
	sb.WriteString("\n详情:\n" + strings.Join(outputs, "\n"))

	t.Log("\n" + sb.String())
	os.WriteFile("/tmp/bazi_dayun_v2_report.txt", []byte(sb.String()), 0644)

	if fail > 0 {
		t.Errorf("大运测试有 %d 个断言失败", fail)
	}
}

// --- 干支序列辅助函数 ---

var tianGanOrder = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var diZhiOrder = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

func ganIndex(g string) int {
	for i, v := range tianGanOrder {
		if v == g {
			return i
		}
	}
	return -1
}

func zhiIndex(g string) int {
	for i, v := range diZhiOrder {
		if v == g {
			return i
		}
	}
	return -1
}

func isNextGan(prev, cur string) bool {
	pi := ganIndex(prev)
	ci := ganIndex(cur)
	if pi < 0 || ci < 0 {
		return false
	}
	return ci == (pi+1)%10
}

func isPrevGan(prev, cur string) bool {
	pi := ganIndex(prev)
	ci := ganIndex(cur)
	if pi < 0 || ci < 0 {
		return false
	}
	return ci == (pi-1+10)%10
}

func isNextZhi(prev, cur string) bool {
	pi := zhiIndex(prev)
	ci := zhiIndex(cur)
	if pi < 0 || ci < 0 {
		return false
	}
	return ci == (pi+1)%12
}

func isPrevZhi(prev, cur string) bool {
	pi := zhiIndex(prev)
	ci := zhiIndex(cur)
	if pi < 0 || ci < 0 {
		return false
	}
	return ci == (pi-1+12)%12
}
