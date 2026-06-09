package bazi_test

import (
	. "bazi/internal/service/bazi"
	"fmt"
	"os"
	"strings"
	"testing"
)

// =============================================================================
// 大运精度测试
// 验证阳男阴女顺行、阴男阳女逆行
// =============================================================================

type dayunTestCase struct {
	ID, Desc, Gender          string
	Year, Month, Day, Hour, Min int
	ExpectedDir               string // "顺行" or "逆行"
	ExpectedStartAge          int    // 0=不校验
}

var dayunCases = []dayunTestCase{
	// 阳年男命 = 顺行
	{ID: "DY-001", Desc: "甲子阳年男顺行", Gender: "MALE", Year: 1984, Month: 3, Day: 15, Hour: 12, Min: 0, ExpectedDir: "顺行"},
	// 阴年男命 = 逆行
	{ID: "DY-002", Desc: "乙丑阴年男逆行", Gender: "MALE", Year: 1985, Month: 6, Day: 20, Hour: 8, Min: 0, ExpectedDir: "逆行"},
	// 阳年女命 = 逆行
	{ID: "DY-003", Desc: "甲子阳年女逆行", Gender: "FEMALE", Year: 1984, Month: 3, Day: 15, Hour: 12, Min: 0, ExpectedDir: "逆行"},
	// 阴年女命 = 顺行
	{ID: "DY-004", Desc: "乙丑阴年女顺行", Gender: "FEMALE", Year: 1985, Month: 6, Day: 20, Hour: 8, Min: 0, ExpectedDir: "顺行"},
	// 更多测试
	{ID: "DY-005", Desc: "庚申阳年男顺行", Gender: "MALE", Year: 1980, Month: 3, Day: 1, Hour: 10, Min: 0, ExpectedDir: "顺行"},
	{ID: "DY-006", Desc: "癸亥阴年男逆行", Gender: "MALE", Year: 1983, Month: 9, Day: 15, Hour: 14, Min: 0, ExpectedDir: "逆行"},
	{ID: "DY-007", Desc: "丙寅阳年女逆行", Gender: "FEMALE", Year: 1986, Month: 2, Day: 15, Hour: 6, Min: 0, ExpectedDir: "逆行"},
	{ID: "DY-008", Desc: "丁卯阴年女顺行", Gender: "FEMALE", Year: 1987, Month: 5, Day: 5, Hour: 16, Min: 0, ExpectedDir: "顺行"},
	{ID: "DY-009", Desc: "壬戌阳年男顺行", Gender: "MALE", Year: 1982, Month: 11, Day: 25, Hour: 22, Min: 0, ExpectedDir: "顺行"},
	{ID: "DY-010", Desc: "辛酉阴年男逆行", Gender: "MALE", Year: 1981, Month: 8, Day: 8, Hour: 0, Min: 0, ExpectedDir: "逆行"},
}

func TestDayunPrecision(t *testing.T) {
	svc := &BaziService{}
	var pass, fail int
	var fails []string
	var outputs []string

	for _, tc := range dayunCases {
		r, err := svc.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Min, tc.Gender)
		if err != nil {
			fail++
			fails = append(fails, fmt.Sprintf("[%s] 计算失败: %v", tc.ID, err))
			continue
		}

		dir := r.DaYunInfo.Direction
		matched := strings.Contains(dir, tc.ExpectedDir)

		outputs = append(outputs, fmt.Sprintf("[%s] %s: 期望=%s 实际=%s ✓=%v",
			tc.ID, tc.Desc, tc.ExpectedDir, dir, matched))

		if matched {
			pass++
		} else {
			fail++
			fails = append(fails, fmt.Sprintf("[%s] %s: 期望方向含'%s', 实际 '%s'", tc.ID, tc.Desc, tc.ExpectedDir, dir))
		}
	}

	var sb strings.Builder
	sb.WriteString("╔══════════════════════════════════════════════╗\n")
	sb.WriteString("║  大运方向精度测试                         ║\n")
	sb.WriteString("╚══════════════════════════════════════════════╝\n")
	total := pass + fail
	sb.WriteString(fmt.Sprintf("\n断言: %d通过 + %d失败 = %d总 | 准确率: %.1f%%\n", pass, fail, total, pctV3(pass, total)))
	if len(fails) > 0 {
		sb.WriteString("\n失败:\n" + strings.Join(fails, "\n") + "\n")
	} else {
		sb.WriteString("全部通过！\n")
	}
	sb.WriteString("\n详情:\n" + strings.Join(outputs, "\n"))

	t.Log("\n" + sb.String())
	os.WriteFile("/tmp/bazi_dayun_precision_report.txt", []byte(sb.String()), 0644)
}
