package data

// NvMingEntry 描述女命八法的判断标准。
// 数据来源：《三命通会》第164-180章 "论女命"。
type NvMingEntry struct {
	Name        string // 名称（纯/和/清/贵/浊/滥/娼/淫 等）
	Category    string // 类别：吉/凶/中性
	Description string // 格局特征描述
	Judgment    string // 综合判断
}

// NvMingEightMethods 女命八法知识库。
var NvMingEightMethods = map[string]NvMingEntry{
	// ── 四吉 ──
	"纯": {
		Name: "纯", Category: "吉",
		Description: "格局纯粹，官星一位得位无破。四柱无官杀混杂、无伤官见官，五行中和，财官印食各得其所。",
		Judgment:    "旺夫相子，一生平顺，家庭和睦。如官星得时得地，可享荣华。",
	},
	"和": {
		Name: "和", Category: "吉",
		Description: "五行和顺，恬静温柔。日主旺衰适中，官杀有制化，不冲不克，格局安静平稳。",
		Judgment:    "性情温和，处世圆融，夫妻和谐，子女顺遂。一生安宁少灾。",
	},
	"清": {
		Name: "清", Category: "吉",
		Description: "格局清高，端庄娴雅。官星清透有力，印绶护身，无浊气干扰。四柱不杂七杀伤官。",
		Judgment:    "品格高尚，受人敬重。配偶有才德，自身气质出众。",
	},
	"贵": {
		Name: "贵", Category: "吉",
		Description: "格局贵重，官星有力得位。财官印三者协调，日主旺相得用。禄马贵人汇聚。",
		Judgment:    "夫贵子荣，自身也有地位成就。可享富贵荣华。",
	},

	// ── 四凶 ──
	"浊": {
		Name: "浊", Category: "凶",
		Description: "格局混杂，用神不清。官杀混杂、食伤见官、财多身弱、五行偏枯。",
		Judgment:    "命运多舛，婚姻不顺，容易陷入是非纠纷。需修身养性化解浊气。",
	},
	"滥": {
		Name: "滥", Category: "凶",
		Description: "官杀混杂过重，或伤官见官而无制。日主弱逢官杀围攻，或伤官过旺克制官星。",
		Judgment:    "婚姻坎坷，感情纠葛多，易遇不良姻缘。需制化官杀或印制伤官化解。",
	},
	"娼": {
		Name: "娼", Category: "凶",
		Description: "伤官旺而无制，或日主太过旺相而合多。伤官透干无印绶化解，或比劫重重争合。",
		Judgment:    "品性轻浮，容易陷入不正当关系。改善需印绶制伤官、官星得位。",
	},
	"淫": {
		Name: "淫", Category: "凶",
		Description: "日主旺而合多，或桃花咸池泛滥。天干多合、地支桃花重见，用神被合去向背。",
		Judgment:    "感情混乱，婚姻难以专一。需官星有力约束、合多逢冲化解。",
	},

	// ── 专论 ──
	"旺夫克子": {
		Name: "旺夫克子", Category: "中性",
		Description: "官星有力旺夫，但食伤受克不利于息。官星得位而印绶过旺克食伤。",
		Judgment: "夫运亨通，但子女缘薄。可通过后天修养弥补子女运。",
	},
	"旺子伤夫": {
		Name: "旺子伤夫", Category: "中性",
		Description: "食伤有力旺子，但伤官克官不利夫星。伤官旺而无印制，官星受克。",
		Judgment: "子女有出息，但夫妻关系需用心经营。",
	},
	"伤夫克子": {
		Name: "伤夫克子", Category: "凶",
		Description: "既伤官星又克食伤，伤官与印绶俱旺交战，或日主过旺官食皆不得力。",
		Judgment: "婚姻和子女两难兼顾，需大运化解或后天修德。",
	},
	"安静守分": {
		Name: "安静守分", Category: "吉",
		Description: "格局平稳无冲无战，日主安静，财官印各安其位，不旺不弱。",
		Judgment: "一生平稳，无大起大落，家庭安宁，福寿可期。",
	},
	"横夭少年": {
		Name: "横夭少年", Category: "凶",
		Description: "命局凶煞汇聚，日主无根无气，或从格破格反克。空亡劫煞汇聚，五行偏枯至极。",
		Judgment: "早年多灾多难，需特别注意健康和意外防范。",
	},
	"福寿两备": {
		Name: "福寿两备", Category: "吉",
		Description: "五行中和，财官印食各得其所，日主旺相得用，大运配合得宜。",
		Judgment: "福禄寿三全，一生安康富足，晚年尤佳。",
	},
	"正偏自处": {
		Name: "正偏自处", Category: "中性",
		Description: "正偏财官各有其位，需辨别正偏之分。正官正财为主，偏官偏财为辅。",
		Judgment: "需自主判断人生选择，正途偏途各有所得。",
	},
	"招嫁不定": {
		Name: "招嫁不定", Category: "中性",
		Description: "官杀并见或财杀混杂，日主左右为难。合多冲多，婚姻信息杂乱。",
		Judgment: "婚姻选择困难，需待大运澄清官星去留。",
	},
}

// XiaoErEntry 小儿命理论断。
// 数据来源：《三命通会》第181章 "论小儿"。
type XiaoErEntry struct {
	Condition string // 条件
	Judgment  string // 断语
}

// XiaoErJudgments 小儿命理论断表。
var XiaoErJudgments = []XiaoErEntry{
	{Condition: "日主旺相无冲克", Judgment: "易养安康，聪明伶俐"},
	{Condition: "财官太重克身", Judgment: "体弱多病，需特别注意健康"},
	{Condition: "关煞汇聚（百日关、短命关等）", Judgment: "幼年磨难多，需父母多加看护"},
	{Condition: "印绶护身有力", Judgment: "得长辈疼爱，学业顺利"},
	{Condition: "食神旺而无制", Judgment: "贪玩好动，需引导教育"},
	{Condition: "日主无根气弱", Judgment: "先天体质偏弱，后天需调养"},
}

// SixRelatives 六亲对应十神表。
// 数据来源：《三命通会》第182章 "论六亲"。
type SixRelative struct {
	Relation string // 六亲关系
	God      string // 对应十神
	Note     string // 说明
}

// SixRelativesTable 六亲对应表。
var SixRelativesTable = []SixRelative{
	{Relation: "父", God: "偏财", Note: "阳男阴女偏财为父，阴男阳女正财为父"},
	{Relation: "母", God: "正印", Note: "正印为母，偏印为继母或寄母"},
	{Relation: "兄弟", God: "比肩/劫财", Note: "比肩为同性兄弟姐妹，劫财为异性兄弟姐妹"},
	{Relation: "妻", God: "正财", Note: "正财为妻，偏财为妾或临时伴侣"},
	{Relation: "子", God: "食神/伤官", Note: "食神为子、伤官为女，七杀为子之说亦有"},
	{Relation: "祖父", God: "偏印", Note: "偏印亦可代表祖辈"},
}

// GetNvMingMethod 根据女命八法名称获取详细描述。
func GetNvMingMethod(name string) (NvMingEntry, bool) {
	e, ok := NvMingEightMethods[name]
	return e, ok
}

// GetSixRelative 根据六亲关系获取对应十神。
func GetSixRelative(relation string) (SixRelative, bool) {
	for _, r := range SixRelativesTable {
		if r.Relation == relation {
			return r, true
		}
	}
	return SixRelative{}, false
}
