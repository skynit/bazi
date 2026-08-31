package ziwei

import (
	"fmt"
	"time"
)

// ResolvePeriodSolarDate validates a complete civil date. An entirely omitted
// date resolves to now, while partial dates are rejected.
func ResolvePeriodSolarDate(year, month, day int, now time.Time) (int, int, int, error) {
	if year == 0 && month == 0 && day == 0 {
		return now.Year(), int(now.Month()), now.Day(), nil
	}
	if year == 0 || month == 0 || day == 0 {
		return 0, 0, 0, fmt.Errorf("year, month, and day must be provided together")
	}
	if year < 1 || year > 9999 || month < 1 || month > 12 || day < 1 || day > 31 {
		return 0, 0, 0, fmt.Errorf("invalid solar date")
	}
	date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	if date.Year() != year || int(date.Month()) != month || date.Day() != day {
		return 0, 0, 0, fmt.Errorf("invalid solar date")
	}
	return year, month, day, nil
}

func CurrentLunarYearLabel(now time.Time) int {
	year, err := LunarYearLabelForSolarDate(now.Year(), int(now.Month()), now.Day())
	if err != nil {
		return now.Year()
	}
	return year
}

func NominalAgeAt(chart *ZiWeiChart, target time.Time) int {
	if chart == nil {
		return 0
	}
	birthLunarYear, err := LunarYearLabelForSolarDate(
		chart.CalculationInput.Year,
		chart.CalculationInput.Month,
		chart.CalculationInput.Day,
	)
	if err != nil {
		return 0
	}
	targetLunarYear, err := LunarYearLabelForSolarDate(target.Year(), int(target.Month()), target.Day())
	if err != nil || targetLunarYear < birthLunarYear {
		return 0
	}
	return targetLunarYear - birthLunarYear + 1
}

func DayunDescription(palace string, startAge int) string {
	descriptions := map[string]string{
		"命宮":  "命宫主题在此大限的结构位置",
		"兄弟宮": "同辈、协作与资源分配主题的结构位置",
		"夫妻宮": "亲密关系、承诺与协商主题的结构位置",
		"子女宮": "子女、下属与创造输出主题的结构位置",
		"財帛宮": "现金流与资源配置主题的结构位置；不构成财务建议",
		"疾厄宮": "传统疾厄宫主题被触发；仅展示宫位与星曜结构，不作个体身体状态推断",
		"遷移宮": "外部环境、出行与社会形象主题的结构位置",
		"僕役宮": "朋友、团队与合作对象主题的结构位置",
		"官祿宮": "职业、责任与组织角色主题的结构位置；不构成职业建议",
		"田宅宮": "家庭、居住与不动产主题的结构位置；不构成交易建议",
		"福德宮": "精神生活、兴趣与内在节奏主题的结构位置",
		"父母宮": "长辈、制度与支持来源主题的结构位置",
	}
	if description, ok := descriptions[palace]; ok {
		return description
	}
	return fmt.Sprintf("%s%s-%d岁大限", palace, palace, startAge)
}
