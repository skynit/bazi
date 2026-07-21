package data

// TiaohouRule holds tiaohou (调候) rules from 《穷通宝鉴》.
// Stem: the heavenly stem (甲乙丙丁戊己庚辛壬癸)
// Month: month branch (寅卯辰巳午未申酉戌亥子丑)
// Climate: sub-climate within the month (暖/热/寒/冷/燥/湿 or season phase)
// XiShen: favorable adjustment deity (调候用神)
// JiShen: unfavorable elements to avoid (忌神)
// Reason: classical explanation from the text.
type TiaohouRule struct {
	Stem    string `json:"stem,omitempty"`
	Month   string `json:"month,omitempty"`
	Climate string `json:"climate,omitempty"`
	XiShen  string `json:"xi_shen"`
	JiShen  string `json:"ji_shen"`
	Reason  string `json:"reason"`
}

// TiaohouData is the complete set of tiaohou rules indexed by stem and month.
// Access via GetTiaohou(stem, month).
var TiaohouData [10][12][]TiaohouRule

func init() {
	// Load all tiaohou rules from 穷通宝鉴.
	// Structure: stem (0-9) x month branch (0-11, 寅=0 子=10)
	// Each entry may have multiple climate variants (初/中/末 or 寒/暖).

	// ========== 甲木 (stem index 0) ==========
	// 寅月 (month index 0)
	// 经典依据：穷通宝鉴 寅月甲木："正月甲木，丙火为主，癸水佐之。丙癸两透，富贵双全。"
	TiaohouData[0][0] = []TiaohouRule{
		{XiShen: "丙", JiShen: "庚", Reason: "寒木无权，寅月余寒未退，喜丙火温木，无丙则木气不活"},
		{XiShen: "癸", JiShen: "庚", Reason: "水盛则木浮，寅月木嫩，过癸则寒"},
	}
	// 卯月 (month index 1)
	// 经典依据：《穷通宝鉴》PDF第11页：正二月甲木“有庚戊者上命，如有丁透”；无戊则庚金力薄。
	TiaohouData[0][1] = []TiaohouRule{
		{XiShen: "庚", JiShen: "癸", Reason: "二月阳刃当令，以庚金制刃裁木为主"},
		{XiShen: "戊", JiShen: "甲", Reason: "庚金在春休囚，须戊土资煞生庚，原文明言用庚不离戊"},
		{XiShen: "丁", JiShen: "癸", Reason: "庚戊得用后，丁火制煞泄秀为上等配合候选"},
	}
	// 辰月 (month index 2)
	// 经典依据：《穷通宝鉴》PDF第16页：三月甲木木气相竭，先庚裁木，次壬泄庚润木。
	TiaohouData[0][2] = []TiaohouRule{
		{XiShen: "庚", JiShen: "乙", Reason: "暮春木气相竭，先用庚金裁抑老木使之成材"},
		{XiShen: "壬", JiShen: "丁", Reason: "阳盛木渴，次用壬水泄庚润木，使枝叶繁茂"},
	}
	// 巳月 (month index 3)
	// 经典依据：穷通宝鉴 巳月甲木："四月甲木，癸水为主，丁火佐之。"
	TiaohouData[0][3] = []TiaohouRule{
		{XiShen: "癸", JiShen: "戊", Reason: "夏木性枯，巳月火旺，先癸水润叶"},
		{XiShen: "丁", JiShen: "戊", Reason: "先癸后丁，丁火佐之调候"},
	}
	// 午月 (month index 4)
	// 经典依据：穷通宝鉴 午月甲木："五月甲木，癸水为主，丁火庚金次之。"
	TiaohouData[0][4] = []TiaohouRule{
		{XiShen: "癸", JiShen: "丙", Reason: "午月丁火司权，癸水调候为急"},
		{XiShen: "丁", JiShen: "己", Reason: "五月先癸后丁，丁火佐癸调候"},
		{XiShen: "庚", JiShen: "丙", Reason: "庚金次之，用以生癸水并配合丁火"},
	}
	// 未月 (month index 5)
	// 经典依据：穷通宝鉴 未月甲木："六月三伏生寒，丁火退气。先丁后庚，无癸亦可。"
	TiaohouData[0][5] = []TiaohouRule{
		{XiShen: "丁", JiShen: "戊", Reason: "六月先丁，三伏生寒而丁火退气，取丁引木"},
		{XiShen: "庚", JiShen: "乙", Reason: "六月后取庚，庚丁相制使甲木成器"},
		{XiShen: "癸", JiShen: "丙", Reason: "六月仍属炎燥，癸水为正格候选；原文明言无癸亦可"},
	}
	// 申月 (month index 6)
	// 经典依据：穷通宝鉴 申月甲木："七月甲木，丁火为主，庚金次之。非庚不能造甲，非丁不能煅庚。"
	TiaohouData[0][6] = []TiaohouRule{
		{XiShen: "丁", JiShen: "癸", Reason: "申月金旺，甲木退气，丁火制金护木，庚为必需用神"},
		{XiShen: "庚", JiShen: "乙", Reason: "庚金劈甲引丁，丁庚两透为贵"},
	}
	// 酉月 (month index 7)
	// 经典依据：穷通宝鉴 酉月甲木："八月甲木，木囚金旺。丁火为先，次用丙火，庚金再次。"
	TiaohouData[0][7] = []TiaohouRule{
		{XiShen: "丁", JiShen: "癸", Reason: "八月木囚金旺，丁火为先以煅金成器"},
		{XiShen: "丙", JiShen: "辛", Reason: "无丁次用丙火，取其调候暖木"},
		{XiShen: "庚", JiShen: "乙", Reason: "庚金再次，配合丙丁裁木成材"},
	}
	// 戌月 (month index 8)
	// 经典依据：穷通宝鉴 戌月甲木："九月甲木，独爱丁火，壬癸滋扶。"
	TiaohouData[0][8] = []TiaohouRule{
		{XiShen: "丁", JiShen: "乙", Reason: "戌月土燥，独爱丁火温木，壬癸滋扶"},
		{XiShen: "壬", JiShen: "戊", Reason: "九月甲木凋零，壬水滋扶并润燥土"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水与丁火配合，滋扶凋残甲木"},
	}
	// 亥月 (month index 9)
	// 经典依据：穷通宝鉴 亥月甲木："十月甲木，以庚为君，以丁为佐。"
	TiaohouData[0][9] = []TiaohouRule{
		{XiShen: "庚", JiShen: "壬", Reason: "亥月水旺木漂，以庚为君以丁为佐"},
		{XiShen: "丁", JiShen: "壬", Reason: "丁火为佐，与庚相制使甲木成器"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火次之，用于初冬暖木调候"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水泛木时以戊土为制，甲多破戊则失其用"},
	}
	// 子月 (month index 10)
	// 经典依据：穷通宝鉴 子月甲木："十一月甲木，丁先庚后，丁火必不可少。"
	TiaohouData[0][10] = []TiaohouRule{
		{XiShen: "丁", JiShen: "壬", Reason: "子月寒木冻结，丁先庚后，丁火必不可少"},
		{XiShen: "庚", JiShen: "癸", Reason: "十一月丁先庚后，庚金劈甲引丁"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火佐之，调和严寒气候"},
		{XiShen: "戊", JiShen: "癸", Reason: "壬癸透出时取戊土制水护火"},
	}
	// 丑月 (month index 11)
	// 经典依据：穷通宝鉴 丑月甲木："十二月甲木，先用庚劈甲，方引丁火。"
	TiaohouData[0][11] = []TiaohouRule{
		{XiShen: "庚", JiShen: "癸", Reason: "丑月湿寒交加，先用庚劈甲方引丁火"},
		{XiShen: "丁", JiShen: "壬", Reason: "丁火温木之寒，配合癸水润木"},
		{XiShen: "丙", JiShen: "辛", Reason: "丙火解冻，三冬甲木虽用庚丁亦不能无丙"},
	}

	// ========== 乙木 (stem index 1) ==========
	// 经典依据：穷通宝鉴 正月乙木："非丙不长，虽有癸润，故先用丙，癸次之。"
	TiaohouData[1][0] = []TiaohouRule{ // 寅月
		{XiShen: "丙", JiShen: "癸", Reason: "寒木萌芽，寅月余寒，丙火温之癸水不宜多"},
		{XiShen: "癸", JiShen: "己", Reason: "癸水次之，以雨露滋养乙木根基"},
	}
	// 经典依据：《穷通宝鉴》PDF第53页：二月乙木以丙为君、癸为臣，丙癸各得其用。
	TiaohouData[1][1] = []TiaohouRule{ // 卯月
		{XiShen: "丙", JiShen: "癸", Reason: "二月阳气渐升，以丙为君，泄木之秀并与癸水相济"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸为臣，润泽卯月旺木，与丙火相济"},
	}
	// 经典依据：《穷通宝鉴》PDF第57页：三月乙木“阳气愈炽，先癸后丙”；阳盛以癸滋之，木盛以丙泄之。
	TiaohouData[1][2] = []TiaohouRule{ // 辰月
		{XiShen: "癸", JiShen: "乙", Reason: "暮春阳气愈炽，先用癸水滋润阴柔乙木"},
		{XiShen: "丙", JiShen: "癸", Reason: "乙木春深而盛，次用丙火泄木之秀"},
	}
	// 经典依据：穷通宝鉴 四月乙木："专用癸水，丙火酌用，虽以庚辛佐癸，须辛透为清。"
	TiaohouData[1][3] = []TiaohouRule{ // 巳月
		{XiShen: "癸", JiShen: "丙", Reason: "夏月火旺，癸水滋乙润燥为先"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火酌用；金水偏多时取丙调候"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金生癸水且不合乙木，原文以辛透为清"},
		{XiShen: "庚", JiShen: "戊", Reason: "庚金可佐癸发源，但与乙相合，次于辛金"},
	}
	// 经典依据：穷通宝鉴 五月乙木：上半月先癸次丙，下半月丙癸齐用。
	TiaohouData[1][4] = []TiaohouRule{ // 午月
		{XiShen: "癸", JiShen: "丙", Reason: "午月丁火司权，癸水调候滋乙"},
		{XiShen: "丙", JiShen: "癸", Reason: "夏至后或柱多金水，丙火与癸水并用"},
	}
	// 经典依据：《穷通宝鉴》PDF第66、67页：六月乙木专用癸水；柱多金水时丙火为尊，但用丙不能离癸；庚辛次之以佐癸。
	TiaohouData[1][5] = []TiaohouRule{ // 未月
		{XiShen: "癸", JiShen: "戊", Reason: "六月乙木以癸水润根为通用主候选，原文明言所重在癸"},
		{XiShen: "丙", JiShen: "壬", Reason: "柱多金水时丙火可优先；静态表不自动裁决该命局条件，平常列为配合候选"},
		{XiShen: "庚", JiShen: "丁", Reason: "庚金次用于生助癸水，使夏月润木之水有源"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金同为癸水之佐，按原文列于庚金之后"},
	}
	// 经典依据：穷通宝鉴 七月乙木："喜己土为用，有丙癸透干……即不见丙癸，己土决不可少。"
	TiaohouData[1][6] = []TiaohouRule{ // 申月
		{XiShen: "己", JiShen: "戊", Reason: "申月庚金乘令，以己土培根挫金，原文明言决不可少"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火制金并暖乙，须与己土配合"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水滋乙化金，与丙己共同构成候选"},
	}
	// 经典依据：穷通宝鉴 八月乙木：白露后专用癸，秋分后先丙后癸；金局又取丁制。
	TiaohouData[1][7] = []TiaohouRule{ // 酉月
		{XiShen: "癸", JiShen: "辛", Reason: "酉月辛金当令，癸水滋乙泄金"},
		{XiShen: "丙", JiShen: "癸", Reason: "秋分后寒木向阳，先丙后癸"},
		{XiShen: "丁", JiShen: "辛", Reason: "支成金局时取暗藏丁火制金护木"},
	}
	// 经典依据：穷通宝鉴 戌月乙木："九月乙木，必赖癸水滋养。"
	TiaohouData[1][8] = []TiaohouRule{ // 戌月
		{XiShen: "癸", JiShen: "丙", Reason: "戌月土旺，必赖癸水滋养乙木"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金发癸水之源，癸辛并见方能持续滋木"},
	}
	// 经典依据：《穷通宝鉴》PDF第81、82页：十月乙木丙为尊；水旺用戊，戊多用甲，甲多用庚，但始终兼用丙火。
	TiaohouData[1][9] = []TiaohouRule{ // 亥月
		{XiShen: "丙", JiShen: "癸", Reason: "亥月水冷木寒，丙火调候暖乙为首"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水过多使乙木漂浮，以戊土制水为救"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊土过多时改取甲木制戊，使乙木有所依附"},
		{XiShen: "庚", JiShen: "丁", Reason: "甲木过多时再取庚金裁甲，仍须兼用丙火暖木"},
	}
	// 经典依据：穷通宝鉴 十一月乙木：专用丙火；壬癸透干时须戊土制水。
	TiaohouData[1][10] = []TiaohouRule{ // 子月
		{XiShen: "丙", JiShen: "壬", Reason: "子月寒冰冻结，丙火解冻温乙木之根"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬癸透干时取戊土制水，保护丙火"},
	}
	// 经典依据：穷通宝鉴 十二月乙木："木寒宜丙，有寒谷回春之象。"
	TiaohouData[1][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "癸", Reason: "丑月寒湿交加，丙火暖乙木为第一要义"},
	}

	// ========== 丙火 (stem index 2) ==========
	// 经典依据：《穷通宝鉴》PDF第90、91页：正月丙火取壬为尊、庚金为佐；壬多用戊，戊多以甲救；火局无壬姑用癸。
	TiaohouData[2][0] = []TiaohouRule{ // 寅月
		{XiShen: "壬", JiShen: "丁", Reason: "正月丙火渐炎，取壬水辅映阳光为尊"},
		{XiShen: "庚", JiShen: "辛", Reason: "寅月壬水临绝，以庚金发水源为佐"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水过多、杀重身轻时取戊土制壬"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊土成片晦火塞水时取甲木制戊为救"},
		{XiShen: "癸", JiShen: "己", Reason: "支成火局而无壬水时姑用癸水，层级仍低于壬水"},
	}
	// 经典依据：穷通宝鉴 二月丙火专用壬水，庚辛生壬；壬多用戊，戊多用甲。
	TiaohouData[2][1] = []TiaohouRule{ // 卯月
		{XiShen: "壬", JiShen: "丁", Reason: "二月阳气舒升，专用壬水辅映丙火"},
		{XiShen: "庚", JiShen: "戊", Reason: "庚金生助壬水，使壬水有源"},
		{XiShen: "辛", JiShen: "丙", Reason: "辛金同为生壬之佐，次列为配合候选"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水成势时以戊土制水"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊土过多克壬时取甲木制戊为救"},
	}
	// 经典依据：《穷通宝鉴》PDF第99页："三月火气渐炎，专用壬水，或支成土局，取甲为辅，壬不可离。"
	TiaohouData[2][2] = []TiaohouRule{ // 辰月
		{XiShen: "壬", JiShen: "戊", Reason: "三月火气渐炎，专用壬水辅映阳光；壬为主候选，不作补足弱火解释"},
		{XiShen: "甲", JiShen: "戊", Reason: "取甲为辅，甲木生火疏土"},
		{XiShen: "庚", JiShen: "丁", Reason: "无甲时以庚金泄土生壬，作为退一步候选"},
	}
	// 经典依据：《穷通宝鉴》PDF第103、104页：四月丙火专用壬水，以庚发源；无壬时癸水姑用；壬多或水局用戊制水。
	TiaohouData[2][3] = []TiaohouRule{ // 巳月
		{XiShen: "壬", JiShen: "戊", Reason: "巳月火旺极热，壬水制火煅金为急"},
		{XiShen: "庚", JiShen: "戊", Reason: "壬水至巳绝地，以庚金发水源"},
		{XiShen: "癸", JiShen: "己", Reason: "无壬时姑用癸水，仍须庚金生助"},
		{XiShen: "戊", JiShen: "甲", Reason: "四柱壬水过多，或支成水局又见一二壬透时，取戊土制水为救"},
	}
	// 经典依据：《穷通宝鉴》PDF第107、112页：五月以壬庚为主；丁多时兼看癸水。
	TiaohouData[2][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "戊", Reason: "午月丁火司权，壬水通根亥子制火为先，忌戊土晦光"},
		{XiShen: "庚", JiShen: "戊", Reason: "庚金生壬水，制火煅金两相宜"},
		{XiShen: "癸", JiShen: "丙", Reason: "五月丁火过多时兼看癸水；仅为条件候选，次于壬庚"},
	}
	// 经典依据：《穷通宝鉴》PDF第111、112页："六月用壬，但借庚金为佐。"
	TiaohouData[2][5] = []TiaohouRule{ // 未月
		{XiShen: "壬", JiShen: "己", Reason: "未月火气余烬，壬水制火煅金为用"},
		{XiShen: "庚", JiShen: "丁", Reason: "六月己土泄火混水，必须借庚金生壬为佐"},
	}
	// 经典依据：《穷通宝鉴》PDF第115页：七月丙火专壬次戊；壬多用戊时不能无甲，以甲破戊并生扶丙火。
	TiaohouData[2][6] = []TiaohouRule{ // 申月
		{XiShen: "壬", JiShen: "己", Reason: "七月日近西山，专取壬水辅映太阳余光"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水过多时次取戊土制水，须防制过"},
		{XiShen: "甲", JiShen: "庚", Reason: "用戊制壬时须甲木破戊生丙，避免戊土反而晦火"},
	}
	// 经典依据：穷通宝鉴 八月丙火仍用壬水；无壬时癸亦可用。
	TiaohouData[2][7] = []TiaohouRule{ // 酉月
		{XiShen: "壬", JiShen: "戊", Reason: "八月丙火日近黄昏，仍用壬水辅映"},
		{XiShen: "癸", JiShen: "己", Reason: "无壬水时癸亦可用，但配合次于壬"},
	}
	// 经典依据：穷通宝鉴 九月丙火："必须先用甲木，次取壬水。"
	TiaohouData[2][8] = []TiaohouRule{ // 戌月
		{XiShen: "甲", JiShen: "庚", Reason: "九月燥土晦光，先用甲木制土生丙"},
		{XiShen: "壬", JiShen: "戊", Reason: "次取壬水辅映丙火，并滋养甲木"},
		{XiShen: "癸", JiShen: "己", Reason: "无壬时癸水可润甲，为异途候选"},
	}
	// 经典依据：《穷通宝鉴》PDF第124、125页：十月丙火以甲、戊、庚配合；壬多有甲无戊时用己混壬；火旺用壬。
	TiaohouData[2][9] = []TiaohouRule{ // 亥月
		{XiShen: "甲", JiShen: "庚", Reason: "十月丙火失令，先取甲木生火并化壬"},
		{XiShen: "戊", JiShen: "甲", Reason: "水旺时以戊土制水培木"},
		{XiShen: "庚", JiShen: "丁", Reason: "木旺时取庚，亦可泄戊生壬"},
		{XiShen: "己", JiShen: "甲", Reason: "壬多有甲而无戊时以己土混壬，缓水培木再生丙火"},
		{XiShen: "壬", JiShen: "己", Reason: "原局火旺时仍取壬水辅映"},
	}
	// 经典依据：穷通宝鉴 子月丙火："十一月丙火，壬水为主，配甲木化煞生丙。"
	TiaohouData[2][10] = []TiaohouRule{ // 子月
		{XiShen: "壬", JiShen: "丁", Reason: "十一月丙火弱中复强，以壬水辅映为最"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水乘旺时以戊土为佐，壬戊酌用"},
		{XiShen: "甲", JiShen: "庚", Reason: "用壬用戊皆不可少甲木生助丙火"},
		{XiShen: "癸", JiShen: "己", Reason: "无壬时癸水亦可用，但不及壬水显达"},
	}
	// 经典依据：穷通宝鉴 十二月丙火："喜壬为用，己土司令，土多又不可少甲。"
	TiaohouData[2][11] = []TiaohouRule{ // 丑月
		{XiShen: "壬", JiShen: "己", Reason: "十二月丙火气进二阳，仍喜壬水辅映取清"},
		{XiShen: "甲", JiShen: "庚", Reason: "己土司令而晦火浊壬，土多不可少甲木"},
	}

	// ========== 丁火 (stem index 3) ==========
	// 经典依据：《穷通宝鉴》PDF第133页：正月丁火先庚次壬；庚及壬癸并见、化木不成时以己土制水。
	TiaohouData[3][0] = []TiaohouRule{ // 寅月
		{XiShen: "庚", JiShen: "乙", Reason: "正月甲木当权，先用庚金劈甲引丁"},
		{XiShen: "壬", JiShen: "戊", Reason: "次取壬水调和木火，使气不偏炎"},
		{XiShen: "己", JiShen: "甲", Reason: "庚金及壬癸并见、丁壬化木不成时，以己土制水为条件救应"},
	}
	// 经典依据：《穷通宝鉴》PDF第137页：二月丁火湿乙伤丁，先庚后甲；癸水成势时以戊土制煞。
	TiaohouData[3][1] = []TiaohouRule{ // 卯月
		{XiShen: "庚", JiShen: "乙", Reason: "二月湿乙伤丁，先用庚金去乙"},
		{XiShen: "甲", JiShen: "己", Reason: "后用甲木引丁，庚甲并用而不相碍"},
		{XiShen: "戊", JiShen: "癸", Reason: "乙木偏少而癸水成势时，以戊土制煞为条件救应"},
	}
	// 经典依据：《穷通宝鉴》PDF第139、140页：三月丁火先甲后庚；支成水局且壬透时以戊己制煞。
	TiaohouData[3][2] = []TiaohouRule{ // 辰月
		{XiShen: "甲", JiShen: "庚", Reason: "辰月土旺泄火，甲木生丁疏土为用"},
		{XiShen: "庚", JiShen: "丁", Reason: "次取庚金劈甲引丁，并泄旺土"},
		{XiShen: "戊", JiShen: "甲", Reason: "支成水局又见壬水透干时，以戊土制煞为救"},
		{XiShen: "己", JiShen: "甲", Reason: "水局壬透时己土与戊土同透制煞，作为配套候选"},
	}
	// 经典依据：《穷通宝鉴》PDF第143、144页：取庚劈甲；火势炎上时以壬解炎，无壬用癸；有庚无甲而戊透时用戊生财。
	TiaohouData[3][3] = []TiaohouRule{ // 巳月
		{XiShen: "庚", JiShen: "乙", Reason: "四月丁火乘旺，必须用庚劈甲，甲多更以庚为先"},
		{XiShen: "甲", JiShen: "己", Reason: "保留一甲引丁，配合庚金形成木火通明"},
		{XiShen: "壬", JiShen: "戊", Reason: "火势炎上或丙火夺光时取壬水解炎制丙"},
		{XiShen: "癸", JiShen: "戊", Reason: "无壬水时以癸水解炎制丙，作为退一步候选"},
		{XiShen: "戊", JiShen: "甲", Reason: "有庚无甲而戊土透干时，取戊土形成伤官生财的条件变格"},
	}
	// 经典依据：《穷通宝鉴》PDF第145、146页：五月丁火以壬癸解炎，壬水须庚发源；无火局而水透时须甲庚并用。
	TiaohouData[3][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "癸", Reason: "午月丁火司权，壬水制丁煅庚为急"},
		{XiShen: "庚", JiShen: "戊", Reason: "壬水临绝，取庚金发水源"},
		{XiShen: "癸", JiShen: "戊", Reason: "无壬或支无壬水时，可取一癸独杀解炎"},
		{XiShen: "甲", JiShen: "己", Reason: "干支不成火局而有水透干时，须甲木引化并配庚劈甲"},
	}
	// 经典依据：穷通宝鉴 六月丁火："专取甲木，壬水次之"，且无庚不妙。
	TiaohouData[3][5] = []TiaohouRule{ // 未月
		{XiShen: "甲", JiShen: "丙", Reason: "未月火气余烬，甲木生丁煅辛为用"},
		{XiShen: "壬", JiShen: "己", Reason: "壬水润木滋丁，己土混水则减力"},
		{XiShen: "庚", JiShen: "乙", Reason: "庚金劈甲引丁并使壬水不浊，原文明言无庚不妙"},
	}
	// 经典依据：《穷通宝鉴》PDF第152页：七月丁火甲庚丙并用；无甲可用乙但不离丙；水重用戊，庚多用壬泄金。
	TiaohouData[3][6] = []TiaohouRule{ // 申月
		{XiShen: "甲", JiShen: "庚", Reason: "申月金水相生，甲木生丁为用神首要"},
		{XiShen: "庚", JiShen: "丁", Reason: "庚金劈甲，令甲木成为引火之物"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火暖金晒甲，秋月不以丙夺丁光为常规忌"},
		{XiShen: "乙", JiShen: "辛", Reason: "七月无甲木时可用乙木枯草引灯，但必须配丙火晒燥"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬癸过多时取戊土制水为救"},
		{XiShen: "壬", JiShen: "戊", Reason: "庚金成势而水不足时取壬水泄庚，避免金多无水"},
	}
	// 经典依据：《穷通宝鉴》PDF第152页：八月甲丙庚皆用；无甲可用乙，但乙木不离丙火晒燥。
	TiaohouData[3][7] = []TiaohouRule{ // 酉月
		{XiShen: "甲", JiShen: "辛", Reason: "酉月辛金当令，丁火煅辛需甲木生助"},
		{XiShen: "丙", JiShen: "癸", Reason: "秋月以丙火暖金晒甲，辅助丁火"},
		{XiShen: "庚", JiShen: "乙", Reason: "八月甲丙之后配庚，庚劈甲以引丁"},
		{XiShen: "乙", JiShen: "辛", Reason: "八月无甲木时以乙木枯草引灯，须配丙火晒燥方能取用"},
	}
	// 经典依据：穷通宝鉴 九月丁火端用甲庚。
	TiaohouData[3][8] = []TiaohouRule{ // 戌月
		{XiShen: "甲", JiShen: "辛", Reason: "戌月土旺金相，丁火以甲木生助为要"},
		{XiShen: "庚", JiShen: "乙", Reason: "九月用甲时庚不可少，庚劈甲引丁"},
	}
	// 经典依据：穷通宝鉴 三冬丁火：甲木为尊，庚金佐之，戊癸权宜酌用。
	TiaohouData[3][9] = []TiaohouRule{ // 亥月
		{XiShen: "甲", JiShen: "庚", Reason: "亥月水冷火绝，甲木生丁暖局为急"},
		{XiShen: "庚", JiShen: "己", Reason: "庚金劈甲引丁，为三冬甲木之佐"},
		{XiShen: "癸", JiShen: "戊", Reason: "丙火夺丁时酌用癸水制丙"},
		{XiShen: "戊", JiShen: "甲", Reason: "两壬争合丁火时酌用戊土破壬"},
	}
	TiaohouData[3][10] = []TiaohouRule{ // 子月
		{XiShen: "甲", JiShen: "庚", Reason: "子月寒水冻木，丁火无根用甲木生助"},
		{XiShen: "庚", JiShen: "己", Reason: "庚金劈甲引丁，为甲木不可缺的佐神"},
		{XiShen: "戊", JiShen: "甲", Reason: "水多癸旺或两壬争合时酌用戊土"},
		{XiShen: "癸", JiShen: "戊", Reason: "四柱丙丁火多时酌用癸水制火"},
	}
	TiaohouData[3][11] = []TiaohouRule{ // 丑月
		{XiShen: "甲", JiShen: "庚", Reason: "丑月寒湿交加，甲木生丁暖局为第一要神"},
		{XiShen: "庚", JiShen: "己", Reason: "庚金劈甲引丁，为三冬丁火佐神"},
		{XiShen: "戊", JiShen: "甲", Reason: "水多癸旺时酌用戊土制水"},
		{XiShen: "癸", JiShen: "戊", Reason: "丙丁火多时酌用癸水调节"},
	}

	// ========== 戊土 (stem index 4) ==========
	// 经典依据：穷通宝鉴 三春戊土总论：正二月先丙后甲，癸又次之。
	TiaohouData[4][0] = []TiaohouRule{ // 寅月
		{XiShen: "丙", JiShen: "壬", Reason: "寅月余寒未退，先取丙火照暖戊土"},
		{XiShen: "甲", JiShen: "庚", Reason: "次取甲木疏劈厚土，使戊土灵动"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水再次，以雨露滋润春土"},
	}
	TiaohouData[4][1] = []TiaohouRule{ // 卯月
		{XiShen: "丙", JiShen: "癸", Reason: "卯月春寒未尽，先用丙火照暖"},
		{XiShen: "甲", JiShen: "庚", Reason: "次取甲木疏土，乙木不能代甲疏劈"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水再次，仲春阳壮时润泽戊土"},
	}
	// 经典依据：穷通宝鉴 三月戊土：甲木为先，丙癸为佐。
	TiaohouData[4][2] = []TiaohouRule{ // 辰月
		{XiShen: "甲", JiShen: "庚", Reason: "三月戊土司令，先用甲木疏劈旺土"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火为佐，配合甲木化煞生土"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水为佐，滋养甲木并润泽旺土"},
	}
	// 经典依据：穷通宝鉴 四月戊土：先用甲疏劈，次取丙癸为佐。
	TiaohouData[4][3] = []TiaohouRule{ // 巳月
		{XiShen: "甲", JiShen: "庚", Reason: "四月戊土厚实，先用甲木疏劈"},
		{XiShen: "丙", JiShen: "壬", Reason: "水湿偏重时以丙火为佐，化煞暖土"},
		{XiShen: "癸", JiShen: "戊", Reason: "火土炎燥时以癸水为佐，润土护甲"},
	}
	// 经典依据：穷通宝鉴 五月戊土：先看壬水，次取甲木，丙火酌用，癸水力微。
	TiaohouData[4][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "戊", Reason: "五月火炎土燥，先用壬水解炎润土"},
		{XiShen: "甲", JiShen: "庚", Reason: "得壬水滋养后次取甲木疏土"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火仅酌情采用，不列在壬甲之前"},
		{XiShen: "癸", JiShen: "戊", Reason: "无壬时癸水可作较弱的润燥候选"},
	}
	// 经典依据：穷通宝鉴 六月戊土：先看癸水，次用丙火、甲木；无癸有壬亦可用。
	TiaohouData[4][5] = []TiaohouRule{ // 未月
		{XiShen: "癸", JiShen: "戊", Reason: "六月土性干枯，先取癸水如雨露润土"},
		{XiShen: "丙", JiShen: "壬", Reason: "得癸润泽后，丙火作为阳和候选"},
		{XiShen: "甲", JiShen: "庚", Reason: "得癸润泽后，甲木作为疏劈候选"},
		{XiShen: "壬", JiShen: "戊", Reason: "无癸时壬水亦可灌溉干土"},
	}
	// 经典依据：穷通宝鉴 七月戊土：先丙后癸，甲木次之。
	TiaohouData[4][6] = []TiaohouRule{ // 申月
		{XiShen: "丙", JiShen: "壬", Reason: "七月火气衰退寒气渐出，先用丙火照暖"},
		{XiShen: "癸", JiShen: "戊", Reason: "阳气充足时后取癸水滋润"},
		{XiShen: "甲", JiShen: "庚", Reason: "土多塞滞时再取甲木疏土"},
	}
	// 经典依据：穷通宝鉴 八月戊土：先丙后癸，不必木疏。
	TiaohouData[4][7] = []TiaohouRule{ // 酉月
		{XiShen: "丙", JiShen: "壬", Reason: "八月金泄身寒，先用丙火照暖戊土"},
		{XiShen: "癸", JiShen: "戊", Reason: "次取癸水滋润燥土，使金气流通"},
	}
	// 经典依据：穷通宝鉴 九月戊土：先看甲木，次取癸水；见金则先癸后丙。
	TiaohouData[4][8] = []TiaohouRule{ // 戌月
		{XiShen: "甲", JiShen: "庚", Reason: "九月戊土当权，先用甲木疏劈旺土"},
		{XiShen: "癸", JiShen: "戊", Reason: "次取癸水滋甲润土，须防戊癸合绊"},
		{XiShen: "丙", JiShen: "壬", Reason: "原局见金时，癸水之后取丙火暖土配合"},
	}
	// 经典依据：穷通宝鉴 十月戊土：先用甲木，次取丙火；见庚用丁，见壬用戊。
	TiaohouData[4][9] = []TiaohouRule{ // 亥月
		{XiShen: "甲", JiShen: "庚", Reason: "十月戊土先用甲木疏土，使厚土灵动"},
		{XiShen: "丙", JiShen: "壬", Reason: "次取丙火温暖冬土，使甲木得以生发"},
		{XiShen: "丁", JiShen: "庚", Reason: "庚金破甲时以丁火制庚为救"},
		{XiShen: "戊", JiShen: "壬", Reason: "壬水伤丙时以戊土制壬为救"},
	}
	// 经典依据：《穷通宝鉴》PDF第190页：十一、十二月丙火为尊、甲木为佐。
	TiaohouData[4][10] = []TiaohouRule{ // 子月
		{XiShen: "丙", JiShen: "壬", Reason: "子月寒水冻结，丙火解冻温土为急"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木为佐，疏土并辅助生发"},
	}
	// 同页“运值二阳”专指十二月：原局一派丙火、弱中变强时，才取壬水为条件候选。
	TiaohouData[4][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "壬", Reason: "丑月严寒土冻，丙火为尊且不可缺"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木为佐，疏土以取贵"},
		{XiShen: "壬", JiShen: "戊", Reason: "十二月原局一派丙火、弱中变强时取壬水调节"},
	}

	// ========== 己土 (stem index 5) ==========
	// 经典依据：穷通宝鉴 正月己土：丙暖为尊，癸为佐；壬多时以戊作堤。
	TiaohouData[5][0] = []TiaohouRule{ // 寅月
		{XiShen: "丙", JiShen: "壬", Reason: "正月田园犹冻，专以丙火解冻照暖"},
		{XiShen: "癸", JiShen: "戊", Reason: "丙火得用后，以癸水润土为佐"},
		{XiShen: "戊", JiShen: "壬", Reason: "壬水浸没田园时取戊土筑堤为救"},
	}
	// 经典依据：穷通宝鉴 二月己土：先取甲木疏之，次取癸水润之，丙火配合。
	TiaohouData[5][1] = []TiaohouRule{ // 卯月
		{XiShen: "甲", JiShen: "庚", Reason: "二月田园未展，先取甲木疏土且须防甲己合绊"},
		{XiShen: "癸", JiShen: "戊", Reason: "次取癸水滋润田园"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火作为暖土配合，已非二月首用"},
		{XiShen: "壬", JiShen: "戊", Reason: "无癸时壬水可生甲，但配合层级较低"},
	}
	// 经典依据：穷通宝鉴 三月己土：先丙后癸，土暖而润，随后用甲疏土。
	TiaohouData[5][2] = []TiaohouRule{ // 辰月
		{XiShen: "丙", JiShen: "壬", Reason: "三月己土先用丙火暖土"},
		{XiShen: "癸", JiShen: "戊", Reason: "次取癸水润土，与丙火配合"},
		{XiShen: "甲", JiShen: "庚", Reason: "土既暖润后，再用甲木疏土"},
	}
	// 经典依据：穷通宝鉴 三夏己土：癸水为要，次用丙火，辛金发癸之源；无癸以壬替代。
	TiaohouData[5][3] = []TiaohouRule{ // 巳月
		{XiShen: "癸", JiShen: "戊", Reason: "四月火炎土燥，专以癸水润泽为先"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火次用，与癸水相济而不相碍"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金发癸水之源，防止滴水熬干"},
		{XiShen: "壬", JiShen: "戊", Reason: "无癸时壬水可作灌溉候选，但层级较低"},
		{XiShen: "庚", JiShen: "丁", Reason: "无辛时庚金亦可作为水源"},
	}
	TiaohouData[5][4] = []TiaohouRule{ // 午月
		{XiShen: "癸", JiShen: "戊", Reason: "五月火旺田园易旱，专以癸水润泽为先"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火次用，与癸水构成暖润配合"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金生癸，使绝地之水有源"},
		{XiShen: "壬", JiShen: "戊", Reason: "无癸时壬水可作灌溉候选，但不及癸水"},
		{XiShen: "庚", JiShen: "丁", Reason: "无辛时庚金亦可作为水源"},
	}
	TiaohouData[5][5] = []TiaohouRule{ // 未月
		{XiShen: "癸", JiShen: "戊", Reason: "六月土燥，仍以癸水润泽田园为先"},
		{XiShen: "丙", JiShen: "壬", Reason: "六月金水进气，兼用丙火保持阳和"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金生癸，使润土之水有源"},
		{XiShen: "壬", JiShen: "戊", Reason: "无癸时壬水可作灌溉候选"},
		{XiShen: "庚", JiShen: "丁", Reason: "无辛时庚金亦可作为水源"},
	}
	// 经典依据：穷通宝鉴 三秋己土：先癸后丙，取辛辅癸。
	TiaohouData[5][6] = []TiaohouRule{ // 申月
		{XiShen: "癸", JiShen: "戊", Reason: "七月金泄土气，先用癸水泄金润土"},
		{XiShen: "丙", JiShen: "壬", Reason: "次取丙火温土并制旺金"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金为癸水之源，作为配合候选"},
		{XiShen: "壬", JiShen: "戊", Reason: "无癸时壬水可替代泄金，但层级较低"},
		{XiShen: "丁", JiShen: "癸", Reason: "支成金局时丁火可制金补土"},
	}
	TiaohouData[5][7] = []TiaohouRule{ // 酉月
		{XiShen: "癸", JiShen: "戊", Reason: "八月金局泄土，先用癸水流通旺金"},
		{XiShen: "丙", JiShen: "壬", Reason: "次取丙火温土并补己土元神"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金为癸水之源，作为配合候选"},
		{XiShen: "壬", JiShen: "戊", Reason: "无癸时壬水可替代泄金，但层级较低"},
		{XiShen: "丁", JiShen: "癸", Reason: "支成金局时丁火可制金补土"},
	}
	// 经典依据：《穷通宝鉴》PDF第208页：三秋己土“癸先丙后”；九月仅在支成四库、土重时另取甲疏土。
	TiaohouData[5][8] = []TiaohouRule{ // 戌月
		{XiShen: "癸", JiShen: "戊", Reason: "三秋己土金旺泄气，先取癸水润土并流通金气"},
		{XiShen: "丙", JiShen: "壬", Reason: "癸水之后取丙火温土并补己土元神"},
		{XiShen: "甲", JiShen: "庚", Reason: "九月支成四库、土重时另取甲木疏土，须防甲己合绊"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金发癸水之源"},
		{XiShen: "壬", JiShen: "戊", Reason: "金气偏重时壬水可辅助泄金"},
		{XiShen: "丁", JiShen: "癸", Reason: "支成金局时丁火可制金补土"},
	}
	// 经典依据：穷通宝鉴 三冬己土：丙火为尊，甲木参用；无丙用丁，壬旺以戊为救。
	TiaohouData[5][9] = []TiaohouRule{ // 亥月
		{XiShen: "丙", JiShen: "癸", Reason: "十月湿泥寒冻，先用丙火解冻暖土"},
		{XiShen: "甲", JiShen: "庚", Reason: "土旺时甲木可参用疏土，但不能代替丙火"},
		{XiShen: "戊", JiShen: "壬", Reason: "初冬壬水当旺时以戊土制水为救"},
		{XiShen: "丁", JiShen: "壬", Reason: "无丙时丁火可退而求其次，须甲木配合"},
	}
	TiaohouData[5][10] = []TiaohouRule{ // 子月
		{XiShen: "丙", JiShen: "癸", Reason: "十一月湿泥寒冻，丙火为尊且不可缺"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木可参用疏土，并辅助丙火"},
		{XiShen: "丁", JiShen: "壬", Reason: "无丙时丁火可退而求其次，须甲木配合"},
		{XiShen: "戊", JiShen: "壬", Reason: "壬水过多浸田时以戊土制水为救"},
	}
	TiaohouData[5][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "癸", Reason: "十二月湿泥寒冻，丙火为尊且不可缺"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木可参用疏土，并辅助丙火"},
		{XiShen: "丁", JiShen: "壬", Reason: "无丙时丁火可退而求其次，须甲木配合"},
		{XiShen: "戊", JiShen: "壬", Reason: "壬水过多浸田时以戊土制水为救"},
	}

	// ========== 庚金 (stem index 6) ==========
	// 经典依据：穷通宝鉴 正月庚金：先用丙暖，甲木疏土，丁火次之。
	TiaohouData[6][0] = []TiaohouRule{ // 寅月
		{XiShen: "丙", JiShen: "癸", Reason: "正月金寒未除，先用丙火照暖庚金"},
		{XiShen: "甲", JiShen: "庚", Reason: "次取甲木疏泄厚土，防止土重埋金"},
		{XiShen: "丁", JiShen: "壬", Reason: "丁火可煅炼庚金，但次于丙甲"},
		{XiShen: "壬", JiShen: "戊", Reason: "支成火局时取壬水制火救金"},
		{XiShen: "癸", JiShen: "己", Reason: "火局无壬时癸水可作较弱救应"},
	}
	// 经典依据：穷通宝鉴 二月庚金：专用丁火，次用甲木，又须庚金劈甲；无丁用丙。
	TiaohouData[6][1] = []TiaohouRule{ // 卯月
		{XiShen: "丁", JiShen: "壬", Reason: "二月庚金暗强，专用丁火煅炼"},
		{XiShen: "甲", JiShen: "乙", Reason: "次用甲木引丁，使炉火有源"},
		{XiShen: "庚", JiShen: "乙", Reason: "再取庚金劈甲引丁，并扶助春金"},
		{XiShen: "丙", JiShen: "癸", Reason: "无丁时姑用丙火，格局层级较低"},
	}
	// 经典依据：穷通宝鉴 三月庚金：先用甲木疏土，次用丁火煅金，无丁姑用丙。
	TiaohouData[6][2] = []TiaohouRule{ // 辰月
		{XiShen: "甲", JiShen: "庚", Reason: "三月戊土司令，先用甲木疏土出金"},
		{XiShen: "丁", JiShen: "壬", Reason: "次用丁火煅炼顽金成器"},
		{XiShen: "丙", JiShen: "癸", Reason: "无丁时姑用丙火，较适合异途取用"},
		{XiShen: "癸", JiShen: "戊", Reason: "支成火局时癸水可制伏局中火势"},
		{XiShen: "壬", JiShen: "戊", Reason: "丙丁透而火局过旺时取壬水为救"},
	}
	// 经典依据：穷通宝鉴 四月庚金：先用壬水，次取戊土，丙火佐之；无壬用癸。
	TiaohouData[6][3] = []TiaohouRule{ // 巳月
		{XiShen: "壬", JiShen: "戊", Reason: "四月火旺金弱，先用壬水制火润土"},
		{XiShen: "戊", JiShen: "甲", Reason: "无水或金弱时次取戊土晦火存金"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火作为配合候选，须按金水强弱酌用"},
		{XiShen: "癸", JiShen: "己", Reason: "无壬时癸水可替代制火润土"},
		{XiShen: "丁", JiShen: "壬", Reason: "支成金局、庚金转强时改取丁火煅炼"},
	}
	// 经典依据：穷通宝鉴 五月庚金：专用壬水，癸又次之。
	TiaohouData[6][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "戊", Reason: "五月丁火旺烈，专用壬水破火润土存金"},
		{XiShen: "癸", JiShen: "己", Reason: "癸水次之，力量不及壬水"},
	}
	// 经典依据：穷通宝鉴 六月庚金：专用丁火，甲木为佐。
	TiaohouData[6][5] = []TiaohouRule{ // 未月
		{XiShen: "丁", JiShen: "癸", Reason: "六月三伏生寒而金顽，专用丁火煅炼"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木为佐，引丁并在土厚时疏土"},
	}
	// 经典依据：穷通宝鉴 七月庚金：专用丁火，次取甲木引丁；无丁姑用丙。
	TiaohouData[6][6] = []TiaohouRule{ // 申月
		{XiShen: "丁", JiShen: "壬", Reason: "七月庚金刚锐，专用丁火煅炼成器"},
		{XiShen: "甲", JiShen: "庚", Reason: "次取甲木引丁，使炉火持续"},
		{XiShen: "丙", JiShen: "癸", Reason: "无丁时姑用丙火调候，层级较低"},
	}
	// 经典依据：穷通宝鉴 八月庚金：仍用丁甲，丙不可少。
	TiaohouData[6][7] = []TiaohouRule{ // 酉月
		{XiShen: "丁", JiShen: "壬", Reason: "八月庚金刚锐未退，先用丁火煅炼"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木引丁并生助退气之火"},
		{XiShen: "丙", JiShen: "癸", Reason: "深秋寒气渐重，丙火不可少"},
	}
	// 经典依据：穷通宝鉴 九月庚金：先用甲疏，后用壬洗；水局时丙火调候。
	TiaohouData[6][8] = []TiaohouRule{ // 戌月
		{XiShen: "甲", JiShen: "庚", Reason: "九月土厚埋金，先用甲木疏土"},
		{XiShen: "壬", JiShen: "己", Reason: "后用壬水洗金，须防戊己混浊制水"},
		{XiShen: "丙", JiShen: "癸", Reason: "支成水局时取丙火调候暖金"},
		{XiShen: "丁", JiShen: "壬", Reason: "土不透而金旺时仍可用丁火煅炼"},
	}
	// 经典依据：穷通宝鉴 十月庚金：丁火为主，丙甲为佐。
	TiaohouData[6][9] = []TiaohouRule{ // 亥月
		{XiShen: "丁", JiShen: "壬", Reason: "十月庚金水冷性寒，以丁火煅炼为主"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火为佐，用于解寒暖金"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木为佐，引丁并泄水生火"},
		{XiShen: "己", JiShen: "壬", Reason: "支见亥子水旺时以己土制水存火"},
		{XiShen: "戊", JiShen: "壬", Reason: "水局过旺时戊土亦可制水救火"},
	}
	// 经典依据：穷通宝鉴 十一月庚金：仍取丁甲，次取丙火照暖。
	TiaohouData[6][10] = []TiaohouRule{ // 子月
		{XiShen: "丁", JiShen: "壬", Reason: "十一月仍取丁火煅炼庚金"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木引丁，使寒月炉火有源"},
		{XiShen: "丙", JiShen: "癸", Reason: "次取丙火照暖，解除严寒"},
	}
	// 经典依据：穷通宝鉴 十二月庚金：先丙解冻，次丁炼金，甲亦不可少。
	TiaohouData[6][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "癸", Reason: "十二月湿泥冰冻，先用丙火解冻"},
		{XiShen: "丁", JiShen: "壬", Reason: "次取丁火煅炼庚金成器"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木引丁，亦为冬月不可少的佐神"},
	}

	// ========== 辛金 (stem index 7) ==========
	// 经典依据：穷通宝鉴 正月辛金：先己后壬，庚佐，丙火酌用。
	TiaohouData[7][0] = []TiaohouRule{ // 寅月
		{XiShen: "己", JiShen: "甲", Reason: "正月辛金失令，以己土滋养固本为先"},
		{XiShen: "壬", JiShen: "戊", Reason: "次用壬水淘洗，使辛金显露光泽"},
		{XiShen: "庚", JiShen: "甲", Reason: "甲木破己时以庚金制甲救己"},
		{XiShen: "丙", JiShen: "癸", Reason: "支成水局或金水偏寒时酌用丙火照暖"},
	}
	// 经典依据：穷通宝鉴 二月辛金：壬水为尊，土透时以甲为救；水木过旺酌取庚丙。
	TiaohouData[7][1] = []TiaohouRule{ // 卯月
		{XiShen: "壬", JiShen: "戊", Reason: "二月阳和，专取壬水淘洗辛金"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊己透干埋金浊壬时，以甲木制土为救"},
		{XiShen: "庚", JiShen: "乙", Reason: "木局泄壬或辛金过弱时，以庚金助辛制木"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水重重淘洗过度时酌取戊土制水"},
		{XiShen: "丙", JiShen: "癸", Reason: "金水偏寒时取丙火调候，须与壬水配合"},
	}
	// 经典依据：穷通宝鉴 三月辛金：先壬后甲。
	TiaohouData[7][2] = []TiaohouRule{ // 辰月
		{XiShen: "壬", JiShen: "戊", Reason: "三月戊土司令，先用壬水泄秀洗金"},
		{XiShen: "甲", JiShen: "庚", Reason: "次用甲木制戊疏土，防止埋金塞水"},
	}
	// 经典依据：穷通宝鉴 四月辛金：喜壬水洗淘，庚金助身，甲木去病；无壬可用癸。
	TiaohouData[7][3] = []TiaohouRule{ // 巳月
		{XiShen: "壬", JiShen: "戊", Reason: "四月火燥金脆，先用壬水制火润土洗金"},
		{XiShen: "庚", JiShen: "丁", Reason: "庚金助身并发壬水之源"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊土透干制壬时，以甲木去病"},
		{XiShen: "癸", JiShen: "己", Reason: "无壬时癸水可作较弱的润洗候选"},
	}
	// 经典依据：穷通宝鉴 五月辛金：壬癸己三者并用；水土过多时以甲调节。
	TiaohouData[7][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "戊", Reason: "五月丁火司权，先用壬水破火润土"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水与壬水并用，补充雨露润泽"},
		{XiShen: "己", JiShen: "甲", Reason: "己土为湿泥，用于生扶阴柔辛金"},
		{XiShen: "庚", JiShen: "丁", Reason: "癸水力弱时以庚金发水源"},
		{XiShen: "甲", JiShen: "庚", Reason: "水土过多时取甲木制土泄水生火"},
	}
	// 经典依据：穷通宝鉴 六月辛金：专用壬水，次取庚金佐之；一壬一己亦可配合。
	TiaohouData[7][5] = []TiaohouRule{ // 未月
		{XiShen: "壬", JiShen: "戊", Reason: "六月己土当权，专用壬水冲刷污土显金"},
		{XiShen: "庚", JiShen: "甲", Reason: "次取庚金扶助辛金并佐壬水"},
		{XiShen: "己", JiShen: "甲", Reason: "仅有未中一己时可与壬水配成湿泥"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊己透干埋金塞水时，甲木可按位置救应"},
	}
	// 经典依据：穷通宝鉴 七月辛金：壬水为尊，甲戊酌用，癸不可代壬。
	TiaohouData[7][6] = []TiaohouRule{ // 申月
		{XiShen: "壬", JiShen: "戊", Reason: "七月辛金不旺自旺，以壬水泄秀为尊"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊土透干阻壬时以甲木制戊为救"},
		{XiShen: "戊", JiShen: "甲", Reason: "金水成势、水多泛滥时酌取戊土制水"},
	}
	// 经典依据：穷通宝鉴 八月辛金：专用壬水，次取甲木破土，财多时庚金为救。
	TiaohouData[7][7] = []TiaohouRule{ // 酉月
		{XiShen: "壬", JiShen: "戊", Reason: "八月辛金旺极，专用壬水淘洗泄秀"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊己透干埋金塞水时，次取甲木破土"},
		{XiShen: "庚", JiShen: "丁", Reason: "甲木过多泄壬时，以庚金制甲发水"},
	}
	// 经典依据：穷通宝鉴 九月辛金：先壬后甲，丙火酌用；无壬时癸可润金。
	TiaohouData[7][8] = []TiaohouRule{ // 戌月
		{XiShen: "壬", JiShen: "戊", Reason: "九月土燥金旺，先用壬水洗土泄金"},
		{XiShen: "甲", JiShen: "庚", Reason: "次取甲木疏土，防止戊土埋金制壬"},
		{XiShen: "丙", JiShen: "癸", Reason: "霜降后或化水条件成立时酌用丙火"},
		{XiShen: "癸", JiShen: "戊", Reason: "无壬时癸水可润土清金，但力量较弱"},
		{XiShen: "庚", JiShen: "丁", Reason: "甲戌月壬水受土时，庚金可化土生水"},
	}
	// 经典依据：穷通宝鉴 十月辛金：先壬后丙，水多时戊土收束，土重时甲木疏解。
	TiaohouData[7][9] = []TiaohouRule{ // 亥月
		{XiShen: "壬", JiShen: "戊", Reason: "十月辛金先用壬水淘洗，取金白水清"},
		{XiShen: "丙", JiShen: "癸", Reason: "次用丙火暖壬温辛"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水过多时以戊土制水聚流"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊己过重埋金塞水时以甲木破土"},
		{XiShen: "己", JiShen: "甲", Reason: "原局需先暖后生时，己土可酌情参用"},
	}
	// 经典依据：穷通宝鉴 十一月辛金：丙火调候为先，壬水淘洗；水多以戊甲配合。
	TiaohouData[7][10] = []TiaohouRule{ // 子月
		{XiShen: "丙", JiShen: "癸", Reason: "十一月层冰冷坚，先用丙火温金暖水"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水用于淘洗辛金，须防过多泄金"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水过多或支成水局时以戊土制水"},
		{XiShen: "甲", JiShen: "庚", Reason: "壬多用戊时，以甲木配合疏土生火"},
	}
	// 经典依据：穷通宝鉴 十二月辛金：先丙后壬，戊己次之；金局以丁甲配合。
	TiaohouData[7][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "癸", Reason: "十二月寒冻至极，先用丙火解冻"},
		{XiShen: "壬", JiShen: "戊", Reason: "次取壬水淘洗辛金"},
		{XiShen: "戊", JiShen: "甲", Reason: "水多时戊土作为次级收束候选"},
		{XiShen: "己", JiShen: "甲", Reason: "火土偏重时己土可与癸水同宫配合"},
		{XiShen: "丁", JiShen: "壬", Reason: "支成金局时取丁火煅炼"},
		{XiShen: "甲", JiShen: "庚", Reason: "金局用丁时须甲木引丁"},
	}

	// ========== 壬水 (stem index 8) ==========
	// 经典依据：穷通宝鉴 正月壬水：先庚，次丙，又次戊。
	TiaohouData[8][0] = []TiaohouRule{ // 寅月
		{XiShen: "庚", JiShen: "丁", Reason: "正月壬水气衰散漫，先用庚金发源"},
		{XiShen: "丙", JiShen: "癸", Reason: "次取丙火除寒，使春水温暖"},
		{XiShen: "戊", JiShen: "甲", Reason: "比劫较多、水势复旺时再取戊土止流"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊土过多时以甲木制煞疏土"},
	}
	// 经典依据：穷通宝鉴 二月壬水：先戊次辛，又次庚金。
	TiaohouData[8][1] = []TiaohouRule{ // 卯月
		{XiShen: "戊", JiShen: "甲", Reason: "二月春水散漫，先用戊土筑堤止流"},
		{XiShen: "辛", JiShen: "丁", Reason: "次取辛金发水源，构成煞印相生"},
		{XiShen: "庚", JiShen: "丁", Reason: "无辛时庚金再次，亦可发源制木"},
		{XiShen: "壬", JiShen: "戊", Reason: "木火过旺泄水时酌取壬水比助"},
	}
	// 经典依据：穷通宝鉴 三月壬水：先用甲疏季土，次取庚发水源。
	TiaohouData[8][2] = []TiaohouRule{ // 辰月
		{XiShen: "甲", JiShen: "庚", Reason: "三月戊土司权，先用甲木疏土通水"},
		{XiShen: "庚", JiShen: "丁", Reason: "次取庚金发水源，须避免与甲相碍"},
		{XiShen: "癸", JiShen: "戊", Reason: "甲木偏弱时癸水可滋甲并助壬"},
		{XiShen: "丙", JiShen: "癸", Reason: "申子辰会水且金多时取丙火制金调候"},
	}
	// 经典依据：穷通宝鉴 四月壬水：专取壬水比助，次取辛金，庚金为佐。
	TiaohouData[8][3] = []TiaohouRule{ // 巳月
		{XiShen: "壬", JiShen: "丁", Reason: "四月丙火司权、水弱极，专取壬水比助"},
		{XiShen: "辛", JiShen: "丙", Reason: "次取辛金发源，须防丙辛合绊"},
		{XiShen: "庚", JiShen: "丁", Reason: "辛金被合时以庚金为佐发源"},
		{XiShen: "癸", JiShen: "戊", Reason: "丁火合壬时可改取癸水助身"},
		{XiShen: "甲", JiShen: "庚", Reason: "癸辛并用而戊土合癸时以甲木为救"},
		{XiShen: "戊", JiShen: "甲", Reason: "金水重叠、壬水转强时取戊土止流"},
	}
	// 经典依据：穷通宝鉴 五月壬水：取庚为用，癸为佐，辛壬亦可参用。
	TiaohouData[8][4] = []TiaohouRule{ // 午月
		{XiShen: "庚", JiShen: "丁", Reason: "五月丁火旺而壬水弱，先用庚金蓄源"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水为佐，用于制丁并扶助壬水"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金位置合宜、不受丁伤时可参用"},
		{XiShen: "壬", JiShen: "丁", Reason: "火旺财多时壬水比肩可参用扶身"},
	}
	// 经典依据：穷通宝鉴 六月壬水：先辛后甲，次取癸水；庚金再次。
	TiaohouData[8][5] = []TiaohouRule{ // 未月
		{XiShen: "辛", JiShen: "丁", Reason: "六月己土当权，先用辛金蓄水源"},
		{XiShen: "甲", JiShen: "庚", Reason: "次取甲木劈土，防己土混浊壬水"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水再次，用于滋甲并助壬"},
		{XiShen: "庚", JiShen: "丁", Reason: "无辛或木局泄水时，庚金可作为次级发源候选"},
		{XiShen: "壬", JiShen: "丁", Reason: "火土较重、壬水无根时取比肩扶身"},
	}
	// 经典依据：穷通宝鉴 七月壬水：专用戊土，丁火为佐。
	TiaohouData[8][6] = []TiaohouRule{ // 申月
		{XiShen: "戊", JiShen: "甲", Reason: "七月壬水源远流长，专用辰戌戊土筑堤"},
		{XiShen: "丁", JiShen: "癸", Reason: "丁火为佐，用于制庚并生助戊土"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊土过多而身强时以甲木制煞"},
		{XiShen: "庚", JiShen: "丁", Reason: "甲木过多泄水时取庚金制木扶身"},
	}
	// 经典依据：穷通宝鉴 八月壬水：专用甲木；壬多水旺时改取戊土。
	TiaohouData[8][7] = []TiaohouRule{ // 酉月
		{XiShen: "甲", JiShen: "庚", Reason: "八月金白水清，专用甲木泄壬制土"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬多且支有申亥、水势奔放时改用戊土"},
	}
	// 经典依据：《穷通宝鉴》PDF第317页：九月壬水专用甲木，次看丙火，戊土酌用；庚破甲时以丁为救。
	TiaohouData[8][8] = []TiaohouRule{ // 戌月
		{XiShen: "甲", JiShen: "庚", Reason: "九月戊土司令，专用甲木制煞疏土"},
		{XiShen: "丙", JiShen: "癸", Reason: "次看丙火，用于生戊并调和水土"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水较多时酌用戊土制水"},
		{XiShen: "丁", JiShen: "癸", Reason: "庚金破甲时以丁火制庚为救"},
	}
	// 经典依据：穷通宝鉴 十月壬水：戊土为尊，丙火为佐，庚金次之。
	TiaohouData[8][9] = []TiaohouRule{ // 亥月
		{XiShen: "戊", JiShen: "甲", Reason: "十月壬水司权至旺，以戊土筑堤为尊"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火为佐，暖水并生助戊土"},
		{XiShen: "庚", JiShen: "丁", Reason: "甲木破戊或木局泄水时以庚金为救"},
	}
	// 经典依据：穷通宝鉴 十一月壬水：先戊后丙，二者并用。
	TiaohouData[8][10] = []TiaohouRule{ // 子月
		{XiShen: "戊", JiShen: "甲", Reason: "十一月阳刃帮身，先取戊土止水"},
		{XiShen: "丙", JiShen: "癸", Reason: "次用丙火解寒并生助戊土"},
	}
	// 经典依据：穷通宝鉴 十二月壬水：先取丙火，丁甲为佐；水旺酌用戊土。
	TiaohouData[8][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "癸", Reason: "十二月寒冻，先用丙火解冻且贯穿全月"},
		{XiShen: "丁", JiShen: "癸", Reason: "丁火为佐，丙辛合绊或金局时可辅助解寒"},
		{XiShen: "甲", JiShen: "庚", Reason: "己土当权浊壬时以甲木疏土为佐"},
		{XiShen: "戊", JiShen: "甲", Reason: "比劫多、水势复旺时酌用戊土止流"},
	}

	// ========== 癸水 (stem index 9) ==========
	// 经典依据：穷通宝鉴 正月癸水：辛金为主，庚金次之，丙火照暖；火局以壬为救。
	TiaohouData[9][0] = []TiaohouRule{ // 寅月
		{XiShen: "辛", JiShen: "丁", Reason: "正月癸水旺极转衰，先用辛金发水源"},
		{XiShen: "庚", JiShen: "丁", Reason: "无辛时庚金次之，亦可生扶癸水"},
		{XiShen: "丙", JiShen: "癸", Reason: "金水有源后取丙火照暖，调和阴阳"},
		{XiShen: "壬", JiShen: "戊", Reason: "支成火局、辛金受伤时以壬水为救"},
	}
	// 经典依据：穷通宝鉴 二月癸水：专用庚金，辛金次之。
	TiaohouData[9][1] = []TiaohouRule{ // 卯月
		{XiShen: "庚", JiShen: "丁", Reason: "二月乙木司令泄水，专用庚金破木发源"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金次之，与庚并用可防乙庚合绊"},
		{XiShen: "己", JiShen: "甲", Reason: "庚辛过多、癸水转强时以己土配合制水"},
		{XiShen: "丁", JiShen: "壬", Reason: "庚辛过多时丁火可制金并生己土"},
	}
	// 经典依据：穷通宝鉴 三月癸水：丙辛不可缺，辛甲酌用；下半月庚金亦可蓄源。
	TiaohouData[9][2] = []TiaohouRule{ // 辰月
		{XiShen: "丙", JiShen: "癸", Reason: "三月癸水先保留丙火调候，清明后尤为首用"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金发源，与丙火构成阴阳暖润配合"},
		{XiShen: "甲", JiShen: "庚", Reason: "戊土透干或四库土重时以甲木为救"},
		{XiShen: "庚", JiShen: "丁", Reason: "土气转旺后庚金可蓄水源，不自动按日期裁决"},
		{XiShen: "壬", JiShen: "戊", Reason: "丙火伤辛时以壬水制丙护金"},
	}
	// 经典依据：穷通宝鉴 四月癸水：专用辛金为尊，无辛用庚，印劫须配合。
	TiaohouData[9][3] = []TiaohouRule{ // 巳月
		{XiShen: "辛", JiShen: "丁", Reason: "四月火土两旺、癸水临绝，专用辛金为尊"},
		{XiShen: "庚", JiShen: "丁", Reason: "无辛时以庚金生水，但层级较低"},
		{XiShen: "壬", JiShen: "戊", Reason: "火旺伤金时以壬水制火护印"},
		{XiShen: "癸", JiShen: "戊", Reason: "丁火伤辛时以癸水制丁护金"},
	}
	// 经典依据：穷通宝鉴 五月癸水：庚辛壬癸参酌并用，印劫缺一不可。
	TiaohouData[9][4] = []TiaohouRule{ // 午月
		{XiShen: "庚", JiShen: "丁", Reason: "五月癸水至弱，先列庚金作为生身印星"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金同为生身之本，与庚金参用"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水比劫制火护金，使印星能够生水"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水比肩与金印配合扶身"},
	}
	// 经典依据：穷通宝鉴 六月癸水：专用庚辛，金水并见为美，比劫为佐。
	TiaohouData[9][5] = []TiaohouRule{ // 未月
		{XiShen: "庚", JiShen: "丁", Reason: "六月癸水不能从煞，专用庚金发源"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金与庚金并用，按金气强弱参酌"},
		{XiShen: "壬", JiShen: "戊", Reason: "金弱火旺时以壬水比劫护印扶身"},
		{XiShen: "癸", JiShen: "戊", Reason: "上半月金弱时癸水比肩作为佐助"},
	}
	// 经典依据：穷通宝鉴 七月癸水：庚金过旺，必取丁火为用，甲木为佐。
	TiaohouData[9][6] = []TiaohouRule{ // 申月
		{XiShen: "丁", JiShen: "壬", Reason: "七月庚金司令刚锐，专取丁火损印为用"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木为佐，引丁并形成伤官生财"},
	}
	// 经典依据：穷通宝鉴 八月癸水：取辛金为用，丙火佐之。
	TiaohouData[9][7] = []TiaohouRule{ // 酉月
		{XiShen: "辛", JiShen: "丁", Reason: "八月金白水清，取辛金生癸为用"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火为佐，形成水暖金温"},
		{XiShen: "丁", JiShen: "壬", Reason: "金水偏多而丁火得根时可用财星调节"},
	}

	// 九月以后按穷通宝鉴章首及总结顺序校正。
	TiaohouData[9][8] = []TiaohouRule{ // 戌月
		{XiShen: "辛", JiShen: "丁", Reason: "九月癸水失令无根，专用辛金发源"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木为佐，用于制伏当权戊土"},
		{XiShen: "癸", JiShen: "戊", Reason: "癸水比肩滋甲，使枯木能够制土"},
		{XiShen: "庚", JiShen: "丁", Reason: "无辛时庚金可作较低层级替代"},
		{XiShen: "壬", JiShen: "戊", Reason: "无癸时壬水可滋甲助身"},
	}
	TiaohouData[9][9] = []TiaohouRule{ // 亥月
		{XiShen: "庚", JiShen: "丁", Reason: "十月癸水旺中有弱，先用庚金生扶"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金同为印星，与庚金并用为妙"},
		{XiShen: "戊", JiShen: "甲", Reason: "一派壬水形成冬水汪洋时取戊土止流"},
		{XiShen: "丁", JiShen: "壬", Reason: "庚辛过多造成金多水涩时取丁火损印"},
		{XiShen: "丙", JiShen: "癸", Reason: "金水偏寒且丙火有根时可作调候候选"},
	}
	TiaohouData[9][10] = []TiaohouRule{ // 子月
		{XiShen: "丙", JiShen: "癸", Reason: "十一月冰冻，专用丙火解冻暖水"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金滋扶癸水，但必须与丙火配合"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木为佐，用于生助丙火"},
	}
	TiaohouData[9][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "癸", Reason: "十二月癸水落地成冰，专用丙火解冻"},
		{XiShen: "壬", JiShen: "戊", Reason: "丙火得用后，壬水可辅阳光显水气"},
		{XiShen: "戊", JiShen: "甲", Reason: "壬水或水局泛滥时以戊土止流"},
		{XiShen: "丁", JiShen: "壬", Reason: "辛金合丙时以丁火制辛救丙"},
		{XiShen: "辛", JiShen: "丁", Reason: "火局财旺、癸水转弱时取辛金生水"},
		{XiShen: "庚", JiShen: "丁", Reason: "火局财旺时庚金亦可扶助癸水"},
	}
}

// GetTiaohou returns tiaohou rules for a given stem and month branch.
// Returns nil if no rules are defined.
func GetTiaohou(stem string, month string) []TiaohouRule {
	stemIdx, ok := ganToIndex[stem]
	if !ok {
		return nil
	}
	monthIdx, ok := zhiToIndex[month]
	if !ok {
		return nil
	}
	return TiaohouData[stemIdx][monthIdx]
}

// GetPrimaryTiaohou returns the primary (first) tiaohou rule for a stem/month pair.
// Returns nil if no rule exists.
func GetPrimaryTiaohou(stem string, month string) *TiaohouRule {
	rules := GetTiaohou(stem, month)
	if len(rules) == 0 {
		return nil
	}
	return &rules[0]
}

// ganToIndex maps stem character to array index 0-9.
var ganToIndex = map[string]int{
	"甲": 0, "乙": 1, "丙": 2, "丁": 3, "戊": 4,
	"己": 5, "庚": 6, "辛": 7, "壬": 8, "癸": 9,
}

// zhiToIndex maps branch character to array index 0-11.
// 寅=0, 卯=1, 辰=2, 巳=3, 午=4, 未=5, 申=6, 酉=7, 戌=8, 亥=9, 子=10, 丑=11
var zhiToIndex = map[string]int{
	"寅": 0, "卯": 1, "辰": 2, "巳": 3, "午": 4,
	"未": 5, "申": 6, "酉": 7, "戌": 8, "亥": 9,
	"子": 10, "丑": 11,
}
