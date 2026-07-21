package ziwei

import "fmt"

var templateBrightnessLevels = []string{"廟", "旺", "得", "利", "平", "陷", "不"}

// These templates describe computable chart structure only. Outcome claims
// remain unadjudicated until independently reviewed Gold cases exist.
func buildMainStarTemplates() map[string]map[string]string {
	stars := []string{
		"紫微", "天府", "天機", "太陽", "武曲", "天同", "廉貞",
		"貪狼", "巨門", "天相", "天梁", "七殺", "破軍", "太陰",
	}
	result := make(map[string]map[string]string, len(stars))
	for _, star := range stars {
		levels := make(map[string]string, len(templateBrightnessLevels))
		for _, level := range templateBrightnessLevels {
			levels[level] = fmt.Sprintf("%s亮度為%s；此處只記錄傳統星曜強弱結構，不推導個體職業、財務或人生結果。", star, level)
		}
		result[star] = levels
	}
	return result
}

func buildAuxStarTemplates() map[string]string {
	stars := []string{
		"左輔", "右弼", "文昌", "文曲", "天魁", "天鉞",
		"擎羊", "陀羅", "火星", "鈴星", "地空", "地劫", "祿存", "天馬",
	}
	result := make(map[string]string, len(stars))
	for _, star := range stars {
		result[star] = fmt.Sprintf("%s落入本宮；此處只記錄星曜位置，不推導個體職業、財務或人生結果。", star)
	}
	return result
}

func buildFourHuaTemplates() map[string]map[string]string {
	stars := []string{
		"紫微", "天機", "太陽", "武曲", "天同", "廉貞", "天府",
		"太陰", "貪狼", "巨門", "天相", "天梁", "七殺", "破軍",
	}
	themes := map[string]string{
		"化祿": "資源與承接主題",
		"化權": "責任與執行主題",
		"化科": "名聲、學習與憑證主題",
		"化忌": "阻滯、執著與代價主題",
	}
	result := make(map[string]map[string]string, len(themes))
	for hua, theme := range themes {
		starTemplates := make(map[string]string, len(stars))
		for _, star := range stars {
			starTemplates[star] = fmt.Sprintf("%s%s，傳統上記錄%s；具體事件與人生結果未裁決。", star, hua, theme)
		}
		result[hua] = starTemplates
	}
	return result
}

func buildPatternTemplates() map[string]string {
	names := []string{
		"紫府同宮", "殺破狼格", "機月同梁格",
		"紫武廉府", "府相朝垣", "日月拱照", "日月反背", "祿馬交馳",
		"火貪格", "鈴貪格",
		"空宮", "日月並明",
		"極向離明", "石中隱玉", "文桂文華", "祿馬佩印",
	}
	result := make(map[string]string, len(names))
	for _, name := range names {
		result[name] = fmt.Sprintf("檢測到%s的傳統星曜組合；只記錄組合條件，具體事件與人生結果未裁決。", name)
	}
	return result
}

func buildPalaceContext() map[string]string {
	return map[string]string{
		"命宮":  "命宮：性格與自我定位主題。",
		"兄弟宮": "兄弟宮：同輩、協作與資源分配主題。",
		"夫妻宮": "夫妻宮：親密關係、承諾與協商主題。",
		"子女宮": "子女宮：子女、下屬與創造輸出主題。",
		"財帛宮": "財帛宮：現金流與資源配置主題；不構成財務建議。",
		"疾厄宮": "疾厄宮：只展示星曜、四化與三方四正結構，不推導個體身體狀態。",
		"遷移宮": "遷移宮：外部環境、出行與社會形象主題。",
		"僕役宮": "僕役宮：朋友、團隊與合作對象主題。",
		"官祿宮": "官祿宮：職業、責任與組織角色主題；不構成職業建議。",
		"田宅宮": "田宅宮：家庭、居住與不動產主題；不構成交易建議。",
		"福德宮": "福德宮：精神生活、興趣與內在節奏主題。",
		"父母宮": "父母宮：長輩、制度與支持來源主題。",
	}
}
