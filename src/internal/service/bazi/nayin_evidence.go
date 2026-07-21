package bazi

import (
	"fmt"

	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

func naYinNameAndElement(gan, zhi string) (string, string, bool) {
	cycle, err := tyme.SixtyCycle{}.FromName(gan + zhi)
	if err != nil {
		return "", "", false
	}

	name := cycle.GetSound().GetName()
	switch name {
	case "沙中金":
		name = "砂中金"
	case "沙中土":
		name = "砂中土"
	case "泉中水":
		name = "井泉水"
	}
	runes := []rune(name)
	if len(runes) == 0 {
		return "", "", false
	}
	element := string(runes[len(runes)-1])
	switch element {
	case "金", "木", "水", "火", "土":
		return name, element, true
	default:
		return "", "", false
	}
}

func calcMingGongGanZhi(yearGan, monthZhi, hourZhi string) (string, error) {
	ganOrder := []rune("甲乙丙丁戊己庚辛壬癸")
	zhiOrder := []rune("子丑寅卯辰巳午未申酉戌亥")
	yearGanIndex := fixedSymbolIndex(ganOrder, yearGan)
	if yearGanIndex < 0 {
		return "", fmt.Errorf("invalid year stem %q", yearGan)
	}
	monthZhiIndex := fixedSymbolIndex(zhiOrder, monthZhi)
	if monthZhiIndex < 0 {
		return "", fmt.Errorf("invalid month branch %q", monthZhi)
	}
	hourZhiIndex := fixedSymbolIndex(zhiOrder, hourZhi)
	if hourZhiIndex < 0 {
		return "", fmt.Errorf("invalid hour branch %q", hourZhi)
	}

	monthNumber := (monthZhiIndex-2+len(zhiOrder))%len(zhiOrder) + 1
	hourNumber := (hourZhiIndex-2+len(zhiOrder))%len(zhiOrder) + 1
	sum := monthNumber + hourNumber
	mingGongMonthNumber := 14 - sum
	if sum >= 14 {
		mingGongMonthNumber = 26 - sum
	}
	mingGongZhiIndex := (mingGongMonthNumber + 1) % len(zhiOrder)
	firstMonthGanIndex := ((yearGanIndex + 1) * 2) % len(ganOrder)
	mingGongGanIndex := (firstMonthGanIndex + mingGongMonthNumber - 1) % len(ganOrder)
	return string(ganOrder[mingGongGanIndex]) + string(zhiOrder[mingGongZhiIndex]), nil
}

func fixedSymbolIndex(order []rune, value string) int {
	runes := []rune(value)
	if len(runes) != 1 {
		return -1
	}
	for index, symbol := range order {
		if symbol == runes[0] {
			return index
		}
	}
	return -1
}

func mingGongShenSha(zhi string) string {
	index := fixedSymbolIndex([]rune("子丑寅卯辰巳午未申酉戌亥"), zhi)
	if index < 0 {
		return ""
	}
	names := [...]string{
		"天贵", "天厄", "天权", "天赦", "天如", "天文",
		"天福", "天驿", "天孤", "天秘", "天艺", "天寿",
	}
	return names[index]
}

func buildMingGongDetail(ganZhi string) data.MingGongDetail {
	detail := data.MingGongDetail{GanZhi: ganZhi}
	runes := []rune(ganZhi)
	if len(runes) != 2 {
		return detail
	}
	detail.Gan = string(runes[0])
	detail.Zhi = string(runes[1])
	name, _, ok := naYinNameAndElement(detail.Gan, detail.Zhi)
	if !ok {
		return detail
	}
	detail.ShenSha = mingGongShenSha(detail.Zhi)
	detail.Nayin = name
	return detail
}

func observeNaYin(gan, zhi string) NaYinInfo {
	name, element, ok := naYinNameAndElement(gan, zhi)
	if !ok {
		return NaYinInfo{
			RuleID: "nayin.sixty-cycle-v1",
			GanZhi: gan + zhi,
			Basis:  "pillar_gan_zhi",
			Status: "unavailable",
		}
	}
	return NaYinInfo{
		RuleID:  "nayin.sixty-cycle-v1",
		GanZhi:  gan + zhi,
		Name:    name,
		Element: element,
		Basis:   "pillar_gan_zhi",
		Status:  "observed",
	}
}

// ValidNaYinEvidence verifies persisted evidence before a saved snapshot is
// allowed to drive API responses.
func ValidNaYinEvidence(evidence NaYinInfo, gan, zhi string) bool {
	want := observeNaYin(gan, zhi)
	return want.Status == "observed" && evidence == want
}
