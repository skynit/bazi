package bazi_test

import (
	"testing"

	. "bazi/internal/service/bazi"
)

type dayunV2TestCase struct {
	ID, Desc, Gender            string
	Year, Month, Day, Hour, Min int
	ExpectedDir                 string
}

var dayunV2Cases = []dayunV2TestCase{
	{ID: "DYV2-001", Desc: "甲子阳年男顺行", Gender: "MALE", Year: 1984, Month: 3, Day: 15, Hour: 12, ExpectedDir: "顺行"},
	{ID: "DYV2-002", Desc: "乙丑阴年男逆行", Gender: "MALE", Year: 1985, Month: 6, Day: 20, Hour: 8, ExpectedDir: "逆行"},
	{ID: "DYV2-003", Desc: "甲子阳年女逆行", Gender: "FEMALE", Year: 1984, Month: 3, Day: 15, Hour: 12, ExpectedDir: "逆行"},
	{ID: "DYV2-004", Desc: "乙丑阴年女顺行", Gender: "FEMALE", Year: 1985, Month: 6, Day: 20, Hour: 8, ExpectedDir: "顺行"},
	{ID: "DYV2-005", Desc: "丙寅阳年男顺行", Gender: "MALE", Year: 1986, Month: 5, Day: 10, Hour: 10, ExpectedDir: "顺行"},
	{ID: "DYV2-006", Desc: "丁卯阴年女顺行", Gender: "FEMALE", Year: 1987, Month: 8, Day: 15, Hour: 14, ExpectedDir: "顺行"},
	{ID: "DYV2-007", Desc: "庚申阳年男顺行", Gender: "MALE", Year: 1980, Month: 3, Day: 1, Hour: 10, ExpectedDir: "顺行"},
	{ID: "DYV2-008", Desc: "癸亥阴年男逆行", Gender: "MALE", Year: 1983, Month: 11, Day: 25, Hour: 22, ExpectedDir: "逆行"},
	{ID: "DYV2-009", Desc: "壬戌阳年女逆行", Gender: "FEMALE", Year: 1982, Month: 10, Day: 8, Hour: 16, ExpectedDir: "逆行"},
	{ID: "DYV2-010", Desc: "辛酉阴年男逆行", Gender: "MALE", Year: 1981, Month: 7, Day: 5, Hour: 6, ExpectedDir: "逆行"},
	{ID: "DYV2-011", Desc: "戊午阳年女逆行", Gender: "FEMALE", Year: 1978, Month: 6, Day: 15, Hour: 12, ExpectedDir: "逆行"},
	{ID: "DYV2-012", Desc: "己未阴年女顺行", Gender: "FEMALE", Year: 1979, Month: 9, Day: 20, Hour: 8, ExpectedDir: "顺行"},
}

func TestDayunSequenceInvariants(t *testing.T) {
	service := &BaziService{}
	for _, tc := range dayunV2Cases {
		t.Run(tc.ID, func(t *testing.T) {
			result, err := service.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Min, tc.Gender)
			if err != nil {
				t.Fatal(err)
			}
			dayun := result.DaYunInfo
			if dayun.Direction != tc.ExpectedDir {
				t.Fatalf("%s direction = %q, want %q", tc.Desc, dayun.Direction, tc.ExpectedDir)
			}
			if len(dayun.Pillars) != 8 {
				t.Fatalf("pillar count = %d, want 8", len(dayun.Pillars))
			}
			forward := dayun.Direction == "顺行"
			for i := 1; i < len(dayun.Pillars); i++ {
				previous, current := dayun.Pillars[i-1], dayun.Pillars[i]
				valid := isPrevGan(previous.Gan, current.Gan) && isPrevZhi(previous.Zhi, current.Zhi)
				if forward {
					valid = isNextGan(previous.Gan, current.Gan) && isNextZhi(previous.Zhi, current.Zhi)
				}
				if !valid {
					t.Fatalf("invalid %s sequence at %s -> %s", dayun.Direction, pillarStr(previous), pillarStr(current))
				}
			}
			if !dayun.Calculated || dayun.StartAt == "" || dayun.ReferenceJie == nil {
				t.Fatalf("missing date-level Dayun evidence: %+v", dayun)
			}
		})
	}
}

var tianGanOrder = []string{"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸"}
var diZhiOrder = []string{"子", "丑", "寅", "卯", "辰", "巳", "午", "未", "申", "酉", "戌", "亥"}

func ganIndex(value string) int { return stringIndex(tianGanOrder, value) }
func zhiIndex(value string) int { return stringIndex(diZhiOrder, value) }

func stringIndex(values []string, target string) int {
	for i, value := range values {
		if value == target {
			return i
		}
	}
	return -1
}

func isNextGan(previous, current string) bool {
	return cycleStep(ganIndex(previous), ganIndex(current), 10, 1)
}
func isPrevGan(previous, current string) bool {
	return cycleStep(ganIndex(previous), ganIndex(current), 10, -1)
}
func isNextZhi(previous, current string) bool {
	return cycleStep(zhiIndex(previous), zhiIndex(current), 12, 1)
}
func isPrevZhi(previous, current string) bool {
	return cycleStep(zhiIndex(previous), zhiIndex(current), 12, -1)
}

func cycleStep(previous, current, size, step int) bool {
	return previous >= 0 && current >= 0 && current == (previous+step+size)%size
}
