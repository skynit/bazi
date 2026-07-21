package ziwei

import "testing"

func TestCalculateChartRejectsInvalidPublicBirthInput(t *testing.T) {
	svc := NewZiWeiService()
	tests := []struct {
		name                        string
		year, month, day, hour, min int
		gender                      string
	}{
		{name: "invalid year", year: 0, month: 1, day: 1, gender: "男"},
		{name: "invalid month", year: 2003, month: 13, day: 1, gender: "男"},
		{name: "invalid day", year: 2003, month: 2, day: 30, gender: "男"},
		{name: "invalid hour", year: 2003, month: 4, day: 15, hour: 24, gender: "男"},
		{name: "invalid minute", year: 2003, month: 4, day: 15, min: 60, gender: "男"},
		{name: "invalid gender", year: 2003, month: 4, day: 15, gender: "未知"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			chart, err := svc.CalculateChart(test.year, test.month, test.day, test.hour, test.min, test.gender)
			if err == nil || chart != nil {
				t.Fatalf("invalid public input returned chart=%+v err=%v", chart, err)
			}
		})
	}
}
