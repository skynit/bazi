package bazi_test

import (
	"testing"

	. "bazi/internal/service/bazi"
)

type dayunTestCase struct {
	ID, Desc, Gender            string
	Year, Month, Day, Hour, Min int
	ExpectedDir                 string
}

var dayunCases = []dayunTestCase{
	{ID: "DY-001", Desc: "甲子阳年男顺行", Gender: "MALE", Year: 1984, Month: 3, Day: 15, Hour: 12, ExpectedDir: "顺行"},
	{ID: "DY-002", Desc: "乙丑阴年男逆行", Gender: "MALE", Year: 1985, Month: 6, Day: 20, Hour: 8, ExpectedDir: "逆行"},
	{ID: "DY-003", Desc: "甲子阳年女逆行", Gender: "FEMALE", Year: 1984, Month: 3, Day: 15, Hour: 12, ExpectedDir: "逆行"},
	{ID: "DY-004", Desc: "乙丑阴年女顺行", Gender: "FEMALE", Year: 1985, Month: 6, Day: 20, Hour: 8, ExpectedDir: "顺行"},
	{ID: "DY-005", Desc: "庚申阳年男顺行", Gender: "MALE", Year: 1980, Month: 3, Day: 1, Hour: 10, ExpectedDir: "顺行"},
	{ID: "DY-006", Desc: "癸亥阴年男逆行", Gender: "MALE", Year: 1983, Month: 9, Day: 15, Hour: 14, ExpectedDir: "逆行"},
	{ID: "DY-007", Desc: "丙寅阳年女逆行", Gender: "FEMALE", Year: 1986, Month: 2, Day: 15, Hour: 6, ExpectedDir: "逆行"},
	{ID: "DY-008", Desc: "丁卯阴年女顺行", Gender: "FEMALE", Year: 1987, Month: 5, Day: 5, Hour: 16, ExpectedDir: "顺行"},
	{ID: "DY-009", Desc: "壬戌阳年男顺行", Gender: "MALE", Year: 1982, Month: 11, Day: 25, Hour: 22, ExpectedDir: "顺行"},
	{ID: "DY-010", Desc: "辛酉阴年男逆行", Gender: "MALE", Year: 1981, Month: 8, Day: 8, ExpectedDir: "逆行"},
}

func TestDayunDirectionRule(t *testing.T) {
	service := &BaziService{}
	for _, tc := range dayunCases {
		t.Run(tc.ID, func(t *testing.T) {
			result, err := service.Calculate(tc.Year, tc.Month, tc.Day, tc.Hour, tc.Min, tc.Gender)
			if err != nil {
				t.Fatal(err)
			}
			if result.DaYunInfo.Direction != tc.ExpectedDir {
				t.Fatalf("%s direction = %q, want %q", tc.Desc, result.DaYunInfo.Direction, tc.ExpectedDir)
			}
		})
	}
}
