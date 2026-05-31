package data

// XingQingEntry 描述十神与性情的对应关系。
// 数据来源：《三命通会》第162章 "论性情相貌"。
type XingQingEntry struct {
	God       string // 十神名称
	Positive  string // 正面性情
	Negative  string // 负面性情
	Advice    string // 修身建议
}

// XingQingByTenGod 十神性情对照表。
var XingQingByTenGod = map[string]XingQingEntry{
	"正官": {God: "正官", Positive: "正直守信、循规蹈矩、责任心强", Negative: "过于保守、墨守成规、缺乏变通", Advice: "适当放权，学会在规则中寻找灵活性"},
	"七杀": {God: "七杀", Positive: "权威果断、执行力强、勇往直前", Negative: "暴躁凶狠、冲动行事、控制欲强", Advice: "克制冲动，以制化为贵，食神制杀为上"},
	"正印": {God: "正印", Positive: "仁慈宽容、好学深思、乐于助人", Negative: "依赖性强、懒散迟钝、缺乏主见", Advice: "培养独立能力，行动力要跟上思考力"},
	"偏印": {God: "偏印", Positive: "聪明独特、有才艺、思维敏捷", Negative: "孤僻怪异、不合群、多疑善变", Advice: "融入群体，接纳不同意见"},
	"正财": {God: "正财", Positive: "勤俭务实、诚信可靠、踏实稳重", Negative: "小气吝啬、重利轻义、缺乏浪漫", Advice: "财富之外，也要关注情感和人际关系"},
	"偏财": {God: "偏财", Positive: "慷慨大方、善投资、人际关系好", Negative: "挥霍无度、投机心重、不安定", Advice: "控制消费冲动，建立稳健理财习惯"},
	"食神": {God: "食神", Positive: "温和乐观、有口福、享受生活", Negative: "贪图享乐、懒散随意、缺乏斗志", Advice: "享受之余保持进取心"},
	"伤官": {God: "伤官", Positive: "聪明伶俐、创意丰富、善于表达", Negative: "傲慢刻薄、叛逆不羁、得罪人", Advice: "柔和表达，尊重他人，以德服人"},
	"比肩": {God: "比肩", Positive: "独立自主、讲义气、坚韧不拔", Negative: "固执好胜、争强斗胜、不善合作", Advice: "学会合作共赢，适度放下竞争心态"},
	"劫财": {God: "劫财", Positive: "合作能力强、行动迅速、适应力强", Negative: "争夺强占、冲动鲁莽、分财破耗", Advice: "控制冲动，注意与人合作的边界"},
}

// WuxingAppearance 五行相貌特征。
// 数据来源：《三命通会》第162章。
type WuxingAppearance struct {
	Element    string // 五行
	BodyType   string // 体态
	FaceColor  string // 面色
	Feature    string // 特征
}

// WuxingAppearanceMap 五行相貌表。
var WuxingAppearanceMap = map[string]WuxingAppearance{
	"木": {Element: "木", BodyType: "身材修长、清瘦", FaceColor: "面色青", Feature: "手指细长、关节明显"},
	"火": {Element: "火", BodyType: "上身较长、体态偏瘦", FaceColor: "面色红", Feature: "发质偏干、目光有神"},
	"土": {Element: "土", BodyType: "体型壮实、敦厚稳重", FaceColor: "面色黄", Feature: "唇厚鼻大、体格结实"},
	"金": {Element: "金", BodyType: "皮肤白皙、面部方正", FaceColor: "面色白", Feature: "声音清亮、骨格清秀"},
	"水": {Element: "水", BodyType: "体型丰满、圆润", FaceColor: "面色黑", Feature: "眼大瞳黑、柔韧有致"},
}

// WuxingHealth 五行与脏腑疾病对应关系。
// 数据来源：《三命通会》第163章 "论疾病先知五脏六腑所属"。
type WuxingHealth struct {
	Element string   // 五行
	Organs  []string // 对应脏腑
	Excess  string   // 过旺之病
	Deficit string   // 过弱之病
}

// WuxingHealthMap 五行疾病对照表。
var WuxingHealthMap = map[string]WuxingHealth{
	"木": {Element: "木", Organs: []string{"肝", "胆"}, Excess: "肝胆疾病、筋骨酸痛、偏头痛、易怒烦躁", Deficit: "肝血不足、视力减退、韧带松弛"},
	"火": {Element: "火", Organs: []string{"心", "小肠"}, Excess: "心血管病、失眠多梦、口舌生疮、高血压", Deficit: "心悸气短、血液循环不良、畏寒"},
	"土": {Element: "土", Organs: []string{"脾", "胃"}, Excess: "消化不良、腹胀肥胖、糖尿病倾向", Deficit: "脾胃虚弱、食欲不振、营养不良"},
	"金": {Element: "金", Organs: []string{"肺", "大肠"}, Excess: "呼吸系统病、便秘、皮肤过敏、鼻塞", Deficit: "肺气不足、易感风寒、皮肤干燥"},
	"水": {Element: "水", Organs: []string{"肾", "膀胱"}, Excess: "泌尿系统病、水肿、骨质疏松、耳鸣", Deficit: "肾气不足、腰膝酸软、精力不济"},
}

// GetXingQingByGod 根据十神名称获取性情描述。
func GetXingQingByGod(god string) (XingQingEntry, bool) {
	e, ok := XingQingByTenGod[god]
	return e, ok
}

// GetWuxingAppearance 根据五行获取相貌特征。
func GetWuxingAppearance(element string) (WuxingAppearance, bool) {
	a, ok := WuxingAppearanceMap[element]
	return a, ok
}

// GetWuxingHealth 根据五行获取疾病信息。
func GetWuxingHealth(element string) (WuxingHealth, bool) {
	h, ok := WuxingHealthMap[element]
	return h, ok
}
