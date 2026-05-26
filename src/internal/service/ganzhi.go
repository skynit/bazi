package service

import (
	"fmt"

	"bazi/internal/model"
)

// ============================================================
// 天干新增数据（已有：Gans, GanElement, ganHe 在 data_gans.go / shensha.go）
// ============================================================

// GanYinYang 天干阴阳
var GanYinYang = map[string]string{
	"甲": "阳", "乙": "阴", "丙": "阳", "丁": "阴",
	"戊": "阳", "己": "阴", "庚": "阳", "辛": "阴",
	"壬": "阳", "癸": "阴",
}

// GanHeHua 天干五合化气
var GanHeHua = map[string]string{
	"甲己": "土", "己甲": "土",
	"乙庚": "金", "庚乙": "金",
	"丙辛": "水", "辛丙": "水",
	"丁壬": "木", "壬丁": "木",
	"戊癸": "火", "癸戊": "火",
}

// GanKe 天干相克（我克者）
var GanKe = map[string]string{
	"甲": "戊", "乙": "己", "丙": "庚", "丁": "辛",
	"戊": "壬", "己": "癸", "庚": "甲", "辛": "乙",
	"壬": "丙", "癸": "丁",
}

// GanSheng 天干相生（我生者）
var GanSheng = map[string]string{
	"甲": "丙", "乙": "丁", "丙": "戊", "丁": "己",
	"戊": "庚", "己": "辛", "庚": "壬", "辛": "癸",
	"壬": "甲", "癸": "乙",
}

// ============================================================
// 地支新增数据（已有：Zhis, ZhiElement, zhiLiuHe, zhiLiuChong 在 data_gans.go / shensha.go）
// ============================================================

// ZhiYinYang 地支阴阳
var ZhiYinYang = map[string]string{
	"子": "阳", "丑": "阴", "寅": "阳", "卯": "阴", "辰": "阳", "巳": "阴",
	"午": "阳", "未": "阴", "申": "阳", "酉": "阴", "戌": "阳", "亥": "阴",
}

// ZhiWuXing 地支五行
var ZhiWuXing = map[string]string{
	"子": "水", "丑": "土", "寅": "木", "卯": "木", "辰": "土", "巳": "火",
	"午": "火", "未": "土", "申": "金", "酉": "金", "戌": "土", "亥": "水",
}

// ZhiSanHe 地支三合局
var ZhiSanHe = map[string][]string{
	"申": {"申", "子", "辰"}, "子": {"申", "子", "辰"}, "辰": {"申", "子", "辰"},
	"亥": {"亥", "卯", "未"}, "卯": {"亥", "卯", "未"}, "未": {"亥", "卯", "未"},
	"寅": {"寅", "午", "戌"}, "午": {"寅", "午", "戌"}, "戌": {"寅", "午", "戌"},
	"巳": {"巳", "酉", "丑"}, "酉": {"巳", "酉", "丑"}, "丑": {"巳", "酉", "丑"},
}

// ZhiSanHui 地支三会
var ZhiSanHui = map[string][]string{
	"寅": {"寅", "卯", "辰"}, "卯": {"寅", "卯", "辰"}, "辰": {"寅", "卯", "辰"},
	"巳": {"巳", "午", "未"}, "午": {"巳", "午", "未"}, "未": {"巳", "午", "未"},
	"申": {"申", "酉", "戌"}, "酉": {"申", "酉", "戌"}, "戌": {"申", "酉", "戌"},
	"亥": {"亥", "子", "丑"}, "子": {"亥", "子", "丑"}, "丑": {"亥", "子", "丑"},
}

// ZhiLiuHai 地支六害
var ZhiLiuHai = map[string]string{
	"子": "未", "未": "子", "丑": "午", "午": "丑",
	"寅": "巳", "巳": "寅", "卯": "辰", "辰": "卯",
	"申": "亥", "亥": "申", "酉": "戌", "戌": "酉",
}

// ZhiXing 地支相刑
var ZhiXing = map[string]string{
	"子": "卯", "卯": "子",
	"寅": "巳", "巳": "申", "申": "寅",
	"丑": "戌", "戌": "未", "未": "丑",
	"辰": "辰", "午": "午", "酉": "酉", "亥": "亥",
}

// ============================================================
// 解释器：天干地支关系命理解读
// ============================================================

// ganHeName 天干五合名称
var ganHeName = map[string]string{
	"甲己": "中正之合", "己甲": "中正之合",
	"乙庚": "仁义之合", "庚乙": "仁义之合",
	"丙辛": "威制之合", "辛丙": "威制之合",
	"丁壬": "淫慝之合", "壬丁": "淫慝之合",
	"戊癸": "无情之合", "癸戊": "无情之合",
}

// pillarMeaning 柱位含义
var pillarMeaning = map[string]string{
	labelYear:  "祖上、童年、大环境",
	labelMonth: "父母、兄弟、青年、事业",
	labelDay:   "自身、配偶、中年",
	labelHour:  "子女、晚年、事业成果",
}

func interpretGanHe(gan1, gan2, label1, label2, hua string) string {
	pair := gan1 + gan2
	name := ganHeName[pair]
	base := fmt.Sprintf("%s合%s（%s）", gan1, gan2, name)
	var body string
	switch {
	case (label1 == labelYear && label2 == labelDay) || (label1 == labelDay && label2 == labelYear):
		body = fmt.Sprintf("【祖上与自身的羁绊】年柱与日柱相合，代表祖辈、家庭背景与你的自身及婚姻有深刻联结。合化为%s，增强了命局的融合与协调。", hua)
	case (label1 == labelMonth && label2 == labelDay) || (label1 == labelDay && label2 == labelMonth):
		body = fmt.Sprintf("【事业与自身的和谐】月柱与日柱相合，代表你的内在自我与外部事业环境高度契合，配偶或合作对象与你的社会追求相辅相成。合化为%s，带来稳定与助力。", hua)
	case (label1 == labelDay && label2 == labelHour) || (label1 == labelHour && label2 == labelDay):
		body = fmt.Sprintf("【自身与晚年的联结】日柱与时柱相合，主自身与子女、晚年归宿关系紧密，家庭生活对你的晚年有深远影响。合化为%s，晚年生活丰足。", hua)
	case (label1 == labelYear && label2 == labelMonth) || (label1 == labelMonth && label2 == labelYear):
		body = fmt.Sprintf("【祖上对事业的牵引】年柱与月柱相合，祖辈、原生家庭对你的青少年时期和事业起步有决定性影响，形成了一种承前启后的良性互动。合化为%s。", hua)
	case (label1 == labelYear && label2 == labelHour) || (label1 == labelHour && label2 == labelYear):
		body = fmt.Sprintf("【祖上与晚景的呼应】年柱与时柱遥合，虽然距离较远，但仍主祖辈福德荫及晚年，或子女与你原生家庭有相似特质。合化为%s，影响较隐晦。", hua)
	case (label1 == labelMonth && label2 == labelHour) || (label1 == labelHour && label2 == labelMonth):
		body = fmt.Sprintf("【事业与晚年的过渡】月柱与时柱相合，青年事业与晚年归宿一脉相承，社会成就可能转化为晚年的安定。合化为%s。", hua)
	default:
		body = fmt.Sprintf("%s与%s天干相合，化气为%s，加强了两柱的联系。", label1, label2, hua)
	}
	return base + "\n" + body
}

func interpretGanKe(ganKe, ganBeiKe, labelKe, labelBeiKe string) string {
	base := fmt.Sprintf("%s克%s", ganKe, ganBeiKe)
	var body string
	switch {
	case (labelKe == labelYear && labelBeiKe == labelDay) || (labelKe == labelDay && labelBeiKe == labelYear):
		body = fmt.Sprintf("【祖上对自身的约束】%s克%s，年柱与日柱相克，代表祖辈家庭或早年环境对你自身发展有一定的限制或磨砺，需通过努力突破固有框架。", ganKe, ganBeiKe)
	case (labelKe == labelMonth && labelBeiKe == labelDay):
		body = "【事业对自身的压力】月柱克日柱，事业、外部环境对个人形成压力，但也可能促使你不断成长，将压力转化为动力。"
	case (labelKe == labelDay && labelBeiKe == labelHour):
		body = "【自身对晚年的主导】日柱克时柱，你对子女、晚景有较强的控制欲或责任感，晚年生活会按你的意愿发展。"
	case (labelKe == labelHour && labelBeiKe == labelYear):
		body = "【晚年与根源的调和】时柱克年柱，子女、晚年的理念或环境可能对原生家庭的影响进行修正，带来新的局面。"
	default:
		body = fmt.Sprintf("%s与%s相克，存在互相制约的关系，需注意相关领域的人际关系与健康。", labelKe, labelBeiKe)
	}
	return base + "\n" + body
}

func interpretGanSheng(shengGan, shouGan, labelSheng, labelShou string) string {
	base := fmt.Sprintf("%s生%s", shengGan, shouGan)
	var body string
	switch {
	case (labelSheng == labelMonth && labelShou == labelDay):
		body = "【事业对自身的强力支撑】月柱生扶日柱，代表父母、事业环境对你自身有极大的帮助和滋养，是你力量的重要来源。"
	case (labelSheng == labelDay && labelShou == labelHour):
		body = "【自身对晚年的付出】日柱生扶时柱，你对子女及晚年生活倾注大量心血，晚年也将因此获得回报和满足感。"
	case (labelSheng == labelYear && labelShou == labelDay):
		body = "【祖上对自身的生扶】年柱生扶日柱，祖辈福荫深厚，对你自身发展有正面影响，根基扎实。"
	case (labelSheng == labelYear && labelShou == labelMonth):
		body = "【祖上对事业的荫庇】年柱生扶月柱，原生家庭对事业起步提供良好基础。"
	case (labelSheng == labelMonth && labelShou == labelHour):
		body = "【事业对晚年的铺垫】月柱生扶时柱，青年时期的奋斗成果将滋养晚年生活。"
	default:
		body = fmt.Sprintf("%s生扶%s，两者相生有情，代表相关宫位间有良好的助益关系。", labelSheng, labelShou)
	}
	return base + "\n" + body
}

func interpretZhiLiuHe(zhi1, zhi2, label1, label2 string) string {
	base := fmt.Sprintf("%s%s合", zhi1, zhi2)
	var body string
	switch {
	case (label1 == labelDay && label2 == labelHour) || (label1 == labelHour && label2 == labelDay):
		body = "【自身与晚年的强连接】日时地支相合，婚姻家庭对晚年生活有决定性的影响，子女缘佳，晚景安稳。"
	case (label1 == labelYear && label2 == labelDay) || (label1 == labelDay && label2 == labelYear):
		body = "【根基与婚姻的深度融合】年日地支相合，祖上与自身、婚姻紧密相连，家族因素对择偶和婚姻质量影响较大。"
	case (label1 == labelYear && label2 == labelMonth) || (label1 == labelMonth && label2 == labelYear):
		body = "【祖上与事业的联结】年月地支相合，原生家庭与你的事业起点有密切关联，父母助力或影响较大。"
	case (label1 == labelMonth && label2 == labelHour) || (label1 == labelHour && label2 == labelMonth):
		body = "【事业与晚年的呼应】月时地支相合，早年事业经历可能在晚年重现或转化，带来安定的结局。"
	default:
		body = fmt.Sprintf("%s与%s地支六合，代表两者关系亲密，相互吸引。", label1, label2)
	}
	return base + "\n" + body
}

func interpretZhiSanHui(zhi1, zhi2, label1, label2 string) string {
	base := fmt.Sprintf("%s%s会", zhi1, zhi2)
	element := ""
	if g, ok := ZhiSanHui[zhi1]; ok {
		switch g[0] {
		case "寅":
			element = "木"
		case "巳":
			element = "火"
		case "申":
			element = "金"
		case "亥":
			element = "水"
		}
	}
	var body string
	switch {
	case (label1 == labelYear && label2 == labelHour) || (label1 == labelHour && label2 == labelYear):
		body = fmt.Sprintf("【祖上与晚景的呼应】年时地支同会%s方，祖辈与子女有相似气质，晚年环境与童年根源产生呼应，加强了%s的厚重特质。", element, element)
	case (label1 == labelDay && label2 == labelHour) || (label1 == labelHour && label2 == labelDay):
		body = fmt.Sprintf("【自身与晚年的共鸣】日时地支同会%s方，家庭与晚年生活深度融合，晚年精神世界富足。", element)
	case (label1 == labelYear && label2 == labelDay) || (label1 == labelDay && label2 == labelYear):
		body = fmt.Sprintf("【根基与自身的统一】年日地支同会%s方，祖辈传承与个人发展一脉相承，自我认同感强。", element)
	default:
		body = fmt.Sprintf("%s与%s地支同会%s方，具有同类相聚的增强效应。", label1, label2, element)
	}
	return base + "\n" + body
}

func interpretZhiGeneric(relType, zhi1, zhi2, label1, label2 string) string {
	base := fmt.Sprintf("%s%s%s", zhi1, zhi2, relType)
	body := fmt.Sprintf("【%s与%s的%s关系】代表%s与%s之间存在矛盾、变动或潜在的不和谐因素，需注意相关领域的沟通与健康。", label1, label2, relType, pillarMeaning[label1], pillarMeaning[label2])
	return base + "\n" + body
}

// ============================================================
// 天干地支综合分析
// ============================================================

// GanRelation 天干关系
type GanRelation struct {
	Pillar1 string `json:"pillar1"`
	Pillar2 string `json:"pillar2"`
	Type    string `json:"type"`
	Detail  string `json:"detail"`
}

// ZhiRelation 地支关系
type ZhiRelation struct {
	Pillar1 string `json:"pillar1"`
	Pillar2 string `json:"pillar2"`
	Type    string `json:"type"`
	Detail  string `json:"detail"`
}

// GanZhiAnalysis 干支综合分析结果
type GanZhiAnalysis struct {
	GanRelations []GanRelation `json:"gan_relations"`
	ZhiRelations []ZhiRelation `json:"zhi_relations"`
}

const (
	labelYear  = "年柱"
	labelMonth = "月柱"
	labelDay   = "日柱"
	labelHour  = "时柱"
)

// CalcGanZhiAnalysis 计算四柱之间的天干和地支关系
func CalcGanZhiAnalysis(year, month, day, hour model.Pillar) GanZhiAnalysis {
	var result GanZhiAnalysis

	pillars := []struct {
		label string
		gan   string
		zhi   string
	}{
		{labelYear, year.Gan, year.Zhi},
		{labelMonth, month.Gan, month.Zhi},
		{labelDay, day.Gan, day.Zhi},
		{labelHour, hour.Gan, hour.Zhi},
	}

	// 6 对两两组合
	for i := 0; i < 4; i++ {
		for j := i + 1; j < 4; j++ {
			// 天干关系
			result.GanRelations = append(result.GanRelations,
				calcGanPairRelations(pillars[i].label, pillars[i].gan,
					pillars[j].label, pillars[j].gan)...)

			// 地支关系
			result.ZhiRelations = append(result.ZhiRelations,
				calcZhiPairRelations(pillars[i].label, pillars[i].zhi,
					pillars[j].label, pillars[j].zhi)...)
		}
	}

	return result
}

func calcGanPairRelations(label1, gan1, label2, gan2 string) []GanRelation {
	var rels []GanRelation

	// 五合
	if ganHe[gan1] == gan2 {
		pair := gan1 + gan2
		hua := GanHeHua[pair]
		detail := interpretGanHe(gan1, gan2, label1, label2, hua)
		rels = append(rels, GanRelation{label1, label2, "五合", detail})
	}

	// 相克
	if GanKe[gan1] == gan2 {
		detail := interpretGanKe(gan1, gan2, label1, label2)
		rels = append(rels, GanRelation{label1, label2, "相克", detail})
	} else if GanKe[gan2] == gan1 {
		detail := interpretGanKe(gan2, gan1, label2, label1)
		rels = append(rels, GanRelation{label1, label2, "相克", detail})
	}

	// 相生
	if GanSheng[gan1] == gan2 {
		detail := interpretGanSheng(gan1, gan2, label1, label2)
		rels = append(rels, GanRelation{label1, label2, "相生", detail})
	} else if GanSheng[gan2] == gan1 {
		detail := interpretGanSheng(gan2, gan1, label2, label1)
		rels = append(rels, GanRelation{label1, label2, "相生", detail})
	}

	return rels
}

func calcZhiPairRelations(label1, zhi1, label2, zhi2 string) []ZhiRelation {
	var rels []ZhiRelation

	// 六冲
	if zhiLiuChong[zhi1] == zhi2 {
		detail := interpretZhiGeneric("冲", zhi1, zhi2, label1, label2)
		rels = append(rels, ZhiRelation{label1, label2, "六冲", detail})
	}

	// 六合
	if zhiLiuHe[zhi1] == zhi2 {
		detail := interpretZhiLiuHe(zhi1, zhi2, label1, label2)
		rels = append(rels, ZhiRelation{label1, label2, "六合", detail})
	}

	// 六害
	if ZhiLiuHai[zhi1] == zhi2 {
		detail := interpretZhiGeneric("害", zhi1, zhi2, label1, label2)
		rels = append(rels, ZhiRelation{label1, label2, "六害", detail})
	}

	// 相刑
	if xingTarget, ok := ZhiXing[zhi1]; ok && xingTarget == zhi2 {
		var xingName string
		switch {
		case (zhi1 == "子" && zhi2 == "卯") || (zhi1 == "卯" && zhi2 == "子"):
			xingName = "无礼之刑"
		case (zhi1 == "寅" && zhi2 == "巳") || (zhi1 == "巳" && zhi2 == "申") || (zhi1 == "申" && zhi2 == "寅"):
			xingName = "恃势之刑"
		case (zhi1 == "丑" && zhi2 == "戌") || (zhi1 == "戌" && zhi2 == "未") || (zhi1 == "未" && zhi2 == "丑"):
			xingName = "无恩之刑"
		default:
			xingName = "自刑"
		}
		detail := interpretZhiGeneric("刑（"+xingName+"）", zhi1, zhi2, label1, label2)
		rels = append(rels, ZhiRelation{label1, label2, "相刑", detail})
	}

	// 三会（两支即视为半会，有会意）
	if group := ZhiSanHui[zhi1]; group != nil {
		for _, z := range group {
			if z == zhi2 {
				detail := interpretZhiSanHui(zhi1, zhi2, label1, label2)
				rels = append(rels, ZhiRelation{label1, label2, "三会", detail})
				break
			}
		}
	}

	return rels
}
