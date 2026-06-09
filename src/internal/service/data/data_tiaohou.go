package data

// TiaohouRule holds tiaohou (调候) rules from 《穷通宝鉴》.
// Stem: the heavenly stem (甲乙丙丁戊己庚辛壬癸)
// Month: month branch (寅卯辰巳午未申酉戌亥子丑)
// Climate: sub-climate within the month (暖/热/寒/冷/燥/湿 or season phase)
// XiShen: favorable adjustment deity (调候用神)
// JiShen: unfavorable elements to avoid (忌神)
// Reason: classical explanation from the text.
type TiaohouRule struct {
	Stem   string
	Month  string
	Climate string
	XiShen string
	JiShen string
	Reason string
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
	// 经典依据：穷通宝鉴 卯月甲木："二月甲木，庚金为主，丁火佐之。庚丁两透，科甲定然。"
	TiaohouData[0][1] = []TiaohouRule{
		{XiShen: "庚", JiShen: "癸", Reason: "木旺宜庚丁，庚金劈甲引丁，配合有力"},
		{XiShen: "丁", JiShen: "癸", Reason: "丁火温木，木旺无庚则用丁"},
	}
	// 辰月 (month index 2)
	// 经典依据：穷通宝鉴 辰月甲木："三月甲木，先取庚金，次用壬水。"
	TiaohouData[0][2] = []TiaohouRule{
		{XiShen: "庚", JiShen: "乙", Reason: "湿土培木，庚金疏土引木气"},
		{XiShen: "壬", JiShen: "丁", Reason: "辰月土湿，先取庚金次用壬水"},
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
		{XiShen: "丁", JiShen: "己", Reason: "先癸后丁庚金次之，丁火佐之调候"},
	}
	// 未月 (month index 5)
	// 经典依据：穷通宝鉴 未月甲木："六月甲木，癸水为主，丁火庚金佐之。大暑后先丁后庚。"
	TiaohouData[0][5] = []TiaohouRule{
		{XiShen: "癸", JiShen: "丙", Reason: "未月燥土司令，癸水滋木为先"},
		{XiShen: "丁", JiShen: "戊", Reason: "大暑后先丁后庚，丁火调候为用"},
	}
	// 申月 (month index 6)
	// 经典依据：穷通宝鉴 申月甲木："七月甲木，丁火为主，庚金次之。非庚不能造甲，非丁不能煅庚。"
	TiaohouData[0][6] = []TiaohouRule{
		{XiShen: "丁", JiShen: "癸", Reason: "申月金旺，甲木退气，丁火制金护木，庚为必需用神"},
		{XiShen: "庚", JiShen: "乙", Reason: "庚金劈甲引丁，丁庚两透为贵"},
	}
	// 酉月 (month index 7)
	TiaohouData[0][7] = []TiaohouRule{
		{XiShen: "丁", JiShen: "辛", Reason: "酉月辛金当权，丁火煅金煅甲为用"},
		{XiShen: "庚", JiShen: "乙", Reason: "庚金劈甲引丁，配合成局"},
	}
	// 戌月 (month index 8)
	// 经典依据：穷通宝鉴 戌月甲木："九月甲木，独爱丁火，壬癸滋扶。"
	TiaohouData[0][8] = []TiaohouRule{
		{XiShen: "丁", JiShen: "乙", Reason: "戌月土燥，独爱丁火温木，壬癸滋扶"},
		{XiShen: "庚", JiShen: "辛", Reason: "庚金疏土，配合丁火煅金"},
	}
	// 亥月 (month index 9)
	// 经典依据：穷通宝鉴 亥月甲木："十月甲木，以庚为君，以丁为佐。"
	TiaohouData[0][9] = []TiaohouRule{
		{XiShen: "庚", JiShen: "壬", Reason: "亥月水旺木漂，以庚为君以丁为佐"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火暖木，亥月寒水凛冽，丙为调候要神"},
	}
	// 子月 (month index 10)
	// 经典依据：穷通宝鉴 子月甲木："十一月甲木，丁先庚后，丁火必不可少。"
	TiaohouData[0][10] = []TiaohouRule{
		{XiShen: "丁", JiShen: "壬", Reason: "子月寒木冻结，丁先庚后，丁火必不可少"},
		{XiShen: "戊", JiShen: "癸", Reason: "戊土制水培根，丙戊并用贵显"},
	}
	// 丑月 (month index 11)
	// 经典依据：穷通宝鉴 丑月甲木："十二月甲木，先用庚劈甲，方引丁火。"
	TiaohouData[0][11] = []TiaohouRule{
		{XiShen: "庚", JiShen: "癸", Reason: "丑月湿寒交加，先用庚劈甲方引丁火"},
		{XiShen: "丁", JiShen: "壬", Reason: "丁火温木之寒，配合癸水润木"},
	}

	// ========== 乙木 (stem index 1) ==========
	TiaohouData[1][0] = []TiaohouRule{ // 寅月
		{XiShen: "丙", JiShen: "癸", Reason: "寒木萌芽，寅月余寒，丙火温之癸水不宜多"},
	}
	TiaohouData[1][1] = []TiaohouRule{ // 卯月
		{XiShen: "丙", JiShen: "癸", Reason: "木旺之月，丙火泄秀调候，癸水不宜混"},
	}
	// 经典依据：穷通宝鉴 辰月乙木："三月乙木，先癸后丙。乙木阴柔，不能用庚。"
	TiaohouData[1][2] = []TiaohouRule{ // 辰月
		{XiShen: "癸", JiShen: "乙", Reason: "辰月湿土，先癸后丙，乙木阴柔不能用庚"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火暖土，辰土寒湿用丙护乙"},
	}
	TiaohouData[1][3] = []TiaohouRule{ // 巳月
		{XiShen: "癸", JiShen: "丙", Reason: "夏月火旺，癸水滋乙润燥为先"},
	}
	TiaohouData[1][4] = []TiaohouRule{ // 午月
		{XiShen: "癸", JiShen: "丙", Reason: "午月丁火司权，癸水调候滋乙"},
	}
	TiaohouData[1][5] = []TiaohouRule{ // 未月
		{XiShen: "癸", JiShen: "丙", Reason: "未月燥土，癸水滋乙润木为先"},
	}
	TiaohouData[1][6] = []TiaohouRule{ // 申月
		{XiShen: "癸", JiShen: "庚", Reason: "申月金旺水生，癸水滋乙金见则吉"},
		{XiShen: "丙", JiShen: "丁", Reason: "丙火制金温木，乙木赖以生存"},
	}
	TiaohouData[1][7] = []TiaohouRule{ // 酉月
		{XiShen: "癸", JiShen: "辛", Reason: "酉月辛金当令，癸水滋乙泄金"},
		{XiShen: "丙", JiShen: "丁", Reason: "丙火温木，辛金无丙煅炼无威"},
	}
	// 经典依据：穷通宝鉴 戌月乙木："九月乙木，必赖癸水滋养。"
	TiaohouData[1][8] = []TiaohouRule{ // 戌月
		{XiShen: "癸", JiShen: "丙", Reason: "戌月土旺，必赖癸水滋养乙木"},
	}
	TiaohouData[1][9] = []TiaohouRule{ // 亥月
		{XiShen: "丙", JiShen: "癸", Reason: "亥月水冷木寒，丙火调候暖乙为首"},
	}
	TiaohouData[1][10] = []TiaohouRule{ // 子月
		{XiShen: "丙", JiShen: "壬", Reason: "子月寒冰冻结，丙火解冻温乙木之根"},
	}
	TiaohouData[1][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "癸", Reason: "丑月寒湿交加，丙火暖乙木为第一要义"},
	}

	// ========== 丙火 (stem index 2) ==========
	TiaohouData[2][0] = []TiaohouRule{ // 寅月
		{XiShen: "壬", JiShen: "戊", Reason: "寅月丙火入死地，壬水为用，生甲引丙为相"},
	}
	TiaohouData[2][1] = []TiaohouRule{ // 卯月
		{XiShen: "壬", JiShen: "癸", Reason: "卯月木旺火相，壬水为调候用神，泄木之秀"},
		{XiShen: "庚", JiShen: "戊", Reason: "庚金劈甲引丁，壬庚两透为贵"},
	}
	// 经典依据：穷通宝鉴 辰月丙火："三月丙火，壬水为用，取甲为辅。"
	TiaohouData[2][2] = []TiaohouRule{ // 辰月
		{XiShen: "壬", JiShen: "戊", Reason: "辰月湿土晦火，壬水充日元之不足"},
		{XiShen: "甲", JiShen: "戊", Reason: "取甲为辅，甲木生火疏土"},
	}
	TiaohouData[2][3] = []TiaohouRule{ // 巳月
		{XiShen: "壬", JiShen: "戊", Reason: "巳月火旺极热，壬水制火煅金为急"},
	}
	// 经典依据：穷通宝鉴 午月丙火："五月丙火，壬水通根亥子制火为先，忌戊土晦光。"
	TiaohouData[2][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "戊", Reason: "午月丁火司权，壬水通根亥子制火为先，忌戊土晦光"},
		{XiShen: "庚", JiShen: "戊", Reason: "庚金生壬水，制火煅金两相宜"},
	}
	TiaohouData[2][5] = []TiaohouRule{ // 未月
		{XiShen: "壬", JiShen: "己", Reason: "未月火气余烬，壬水制火煅金为用"},
		{XiShen: "癸", JiShen: "丙", Reason: "癸水为副调候，己土混则减力"},
	}
	TiaohouData[2][6] = []TiaohouRule{ // 申月
		{XiShen: "壬", JiShen: "戊", Reason: "申月金水相生，壬水当令，丙火以壬为用"},
		{XiShen: "庚", JiShen: "丁", Reason: "庚金发壬水之源，丙火赖以照耀"},
	}
	TiaohouData[2][7] = []TiaohouRule{ // 酉月
		{XiShen: "壬", JiShen: "辛", Reason: "酉月辛金当令，壬水洗金滋丙火"},
		{XiShen: "庚", JiShen: "丁", Reason: "庚金生壬水，丙火以庚壬为配"},
	}
	TiaohouData[2][8] = []TiaohouRule{ // 戌月
		{XiShen: "壬", JiShen: "戊", Reason: "戌月火气渐退，壬水充实三夏之功用"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金生壬水，酉戌月辛金为水源"},
	}
	TiaohouData[2][9] = []TiaohouRule{ // 亥月
		{XiShen: "壬", JiShen: "戊", Reason: "亥月水临官，壬水当令，丙火绝地用壬为相"},
		{XiShen: "庚", JiShen: "丙", Reason: "庚金生壬水为用，丙火赖之而明"},
	}
	// 经典依据：穷通宝鉴 子月丙火："十一月丙火，壬水为主，配甲木化煞生丙。"
	TiaohouData[2][10] = []TiaohouRule{ // 子月
		{XiShen: "壬", JiShen: "癸", Reason: "子月水旺火死，壬水解冻温木为先"},
		{XiShen: "甲", JiShen: "壬", Reason: "配甲木化煞生丙，甲木化水生火"},
	}
	TiaohouData[2][11] = []TiaohouRule{ // 丑月
		{XiShen: "壬", JiShen: "己", Reason: "丑月寒湿，壬水解冻温丙火为急"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火暖局，丑月寒气重，丙为调候要神"},
	}

	// ========== 丁火 (stem index 3) ==========
	TiaohouData[3][0] = []TiaohouRule{ // 寅月
		{XiShen: "甲", JiShen: "庚", Reason: "寅月木火相生，甲木生丁引丁为用"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火陪衬，丁火赖丙而旺"},
	}
	TiaohouData[3][1] = []TiaohouRule{ // 卯月
		{XiShen: "甲", JiShen: "庚", Reason: "卯月木旺，丁火入相，甲木生丁为用神"},
		{XiShen: "壬", JiShen: "癸", Reason: "壬水洗金滋丁，泄木之秀"},
	}
	TiaohouData[3][2] = []TiaohouRule{ // 辰月
		{XiShen: "甲", JiShen: "庚", Reason: "辰月土旺泄火，甲木生丁疏土为用"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水充壬，泄土扶丁"},
	}
	TiaohouData[3][3] = []TiaohouRule{ // 巳月
		{XiShen: "甲", JiShen: "庚", Reason: "巳月火势炎蒸，甲木生丁制庚煅金"},
	}
	TiaohouData[3][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "癸", Reason: "午月丁火司权，壬水制丁煅庚为急"},
		{XiShen: "甲", JiShen: "丙", Reason: "甲木生丁，午月木性枯槁，用甲引丁"},
	}
	TiaohouData[3][5] = []TiaohouRule{ // 未月
		{XiShen: "甲", JiShen: "丙", Reason: "未月火气余烬，甲木生丁煅辛为用"},
		{XiShen: "壬", JiShen: "己", Reason: "壬水润木滋丁，己土混水则减力"},
	}
	TiaohouData[3][6] = []TiaohouRule{ // 申月
		{XiShen: "甲", JiShen: "庚", Reason: "申月金水相生，甲木生丁为用神首要"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水洗金滋丁，泄金之肃"},
	}
	TiaohouData[3][7] = []TiaohouRule{ // 酉月
		{XiShen: "甲", JiShen: "辛", Reason: "酉月辛金当令，丁火煅辛需甲木生助"},
		{XiShen: "壬", JiShen: "癸", Reason: "壬水洗金滋丁，酉月丁火本柔用壬水"},
	}
	TiaohouData[3][8] = []TiaohouRule{ // 戌月
		{XiShen: "甲", JiShen: "辛", Reason: "戌月土旺金相，丁火以甲木生助为要"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水润木滋丁，泄土之闷"},
	}
	TiaohouData[3][9] = []TiaohouRule{ // 亥月
		{XiShen: "甲", JiShen: "庚", Reason: "亥月水冷火绝，甲木生丁暖局为急"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火陪衬，亥月甲木生丙生丁皆温局"},
	}
	TiaohouData[3][10] = []TiaohouRule{ // 子月
		{XiShen: "甲", JiShen: "庚", Reason: "子月寒水冻木，丁火无根用甲木生助"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火温木化水，子月寒水当令丙丁皆寒"},
	}
	TiaohouData[3][11] = []TiaohouRule{ // 丑月
		{XiShen: "甲", JiShen: "庚", Reason: "丑月寒湿交加，甲木生丁暖局为第一要神"},
		{XiShen: "丙", JiShen: "癸", Reason: "丙火暖木，丑月寒气凛冽用丙丁温局"},
	}

	// ========== 戊土 (stem index 4) ==========
	TiaohouData[4][0] = []TiaohouRule{ // 寅月
		{XiShen: "丙", JiShen: "壬", Reason: "寅月木旺土虚，丙火温土生身为先"},
		{XiShen: "甲", JiShen: "癸", Reason: "甲木疏土，戊土身弱用丙甲相辅"},
	}
	TiaohouData[4][1] = []TiaohouRule{ // 卯月
		{XiShen: "丙", JiShen: "癸", Reason: "卯月木盛土崩，丙火温土扶身为急"},
		{XiShen: "甲", JiShen: "乙", Reason: "甲木疏土，卯月阳木司权甲为首选"},
	}
	TiaohouData[4][2] = []TiaohouRule{ // 辰月
		{XiShen: "丙", JiShen: "癸", Reason: "辰月湿土司令，丙火暖土燥湿为用"},
		{XiShen: "癸", JiShen: "丙", Reason: "辰月用癸润土，恐丙被湿土所晦"},
	}
	TiaohouData[4][3] = []TiaohouRule{ // 巳月
		{XiShen: "丙", JiShen: "壬", Reason: "巳月火旺土焦，丙火温土为先，壬水润土为急"},
		{XiShen: "甲", JiShen: "癸", Reason: "甲木生火疏土，巳月甲木病地宜重用丙火"},
	}
	TiaohouData[4][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "丙", Reason: "午月火极旺土焦，壬水润土为先，丙火为配"},
		{XiShen: "癸", JiShen: "丙", Reason: "癸水润土调候，午月水气绝地用壬癸并"},
	}
	TiaohouData[4][5] = []TiaohouRule{ // 未月
		{XiShen: "癸", JiShen: "丙", Reason: "未月燥土厚重，癸水滋戊润土为先"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金生癸水，润土之功胜于壬水"},
	}
	TiaohouData[4][6] = []TiaohouRule{ // 申月
		{XiShen: "丙", JiShen: "丁", Reason: "申月金水相生，丙火温土生身为急"},
		{XiShen: "丁", JiShen: "壬", Reason: "丁火煅金温土，申月寒凉土性凉用丙丁温"},
	}
	TiaohouData[4][7] = []TiaohouRule{ // 酉月
		{XiShen: "丙", JiShen: "丁", Reason: "酉月金旺土虚，丙火生身煅金为用"},
		{XiShen: "丁", JiShen: "辛", Reason: "丁火煅辛温土，酉月土生辛金所生太过"},
	}
	TiaohouData[4][8] = []TiaohouRule{ // 戌月
		{XiShen: "辛", JiShen: "丁", Reason: "戌月燥土当令，辛金发水源润土为急"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水润土，戌月火气余烬用壬水调和"},
	}
	TiaohouData[4][9] = []TiaohouRule{ // 亥月
		{XiShen: "丙", JiShen: "壬", Reason: "亥月水冷土寒，丙火温土暖局为第一要义"},
		{XiShen: "甲", JiShen: "癸", Reason: "甲木疏土制水，亥月甲木病地赖丙火生"},
	}
	TiaohouData[4][10] = []TiaohouRule{ // 子月
		{XiShen: "丙", JiShen: "壬", Reason: "子月寒水冻结，丙火解冻温土为急"},
		{XiShen: "丁", JiShen: "癸", Reason: "丁火温土，丙丁并用暖局土性方活"},
	}
	TiaohouData[4][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "丁", Reason: "丑月寒湿交加，丙火暖土生身首要"},
		{XiShen: "丁", JiShen: "壬", Reason: "丁火温土，丑月比亥更寒丁火为主"},
	}

	// ========== 己土 (stem index 5) ==========
	TiaohouData[5][0] = []TiaohouRule{ // 寅月
		{XiShen: "丙", JiShen: "甲", Reason: "寅月余寒未退，丙火暖己土为用，甲木忌多"},
		{XiShen: "甲", JiShen: "丙", Reason: "甲木疏土，寅月甲木司权先用丙火温局"},
	}
	TiaohouData[5][1] = []TiaohouRule{ // 卯月
		{XiShen: "丙", JiShen: "乙", Reason: "卯月木旺土虚，丙火温土生身，甲木疏土为用"},
	}
	TiaohouData[5][2] = []TiaohouRule{ // 辰月
		{XiShen: "丙", JiShen: "癸", Reason: "辰月湿土用丙火温燥，癸水润局"},
		{XiShen: "癸", JiShen: "丙", Reason: "辰月土湿，癸水滋己润土，恐丙被湿土所晦"},
	}
	TiaohouData[5][3] = []TiaohouRule{ // 巳月
		{XiShen: "癸", JiShen: "丙", Reason: "巳月火旺土焦，癸水润己土为急"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金生癸水，巳月金绝地用辛发水源"},
	}
	TiaohouData[5][4] = []TiaohouRule{ // 午月
		{XiShen: "癸", JiShen: "丙", Reason: "午月丁火司权，己土被火焦，癸水调候润土"},
	}
	TiaohouData[5][5] = []TiaohouRule{ // 未月
		{XiShen: "癸", JiShen: "丙", Reason: "未月燥土当令，癸水滋己润木为先"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金生癸水润土，未月金水进气用辛发源"},
	}
	TiaohouData[5][6] = []TiaohouRule{ // 申月
		{XiShen: "丙", JiShen: "丁", Reason: "申月金水相生，丙火温土生身首要"},
		{XiShen: "丁", JiShen: "壬", Reason: "丁火煅金温土，申月土性寒凉用丙丁温之"},
	}
	TiaohouData[5][7] = []TiaohouRule{ // 酉月
		{XiShen: "丙", JiShen: "丁", Reason: "酉月金旺土虚，丙火生身煅金为用"},
		{XiShen: "丁", JiShen: "辛", Reason: "丁火煅辛温土，酉月土生太过用丁煅"},
	}
	TiaohouData[5][8] = []TiaohouRule{ // 戌月
		{XiShen: "辛", JiShen: "丁", Reason: "戌月燥土当令，辛金发水源润土为急"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水润土，戌月火气余烬用壬水调和"},
	}
	TiaohouData[5][9] = []TiaohouRule{ // 亥月
		{XiShen: "丙", JiShen: "壬", Reason: "亥月水冷土寒，丙火温土暖局为第一要义"},
		{XiShen: "甲", JiShen: "癸", Reason: "甲木疏土制水，亥月甲木病地赖丙温生"},
	}
	TiaohouData[5][10] = []TiaohouRule{ // 子月
		{XiShen: "丙", JiShen: "壬", Reason: "子月寒水冻结，丙火解冻温土为急"},
		{XiShen: "丁", JiShen: "癸", Reason: "丁火温土，子月寒气凛冽丙丁皆寒"},
	}
	TiaohouData[5][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "丁", Reason: "丑月寒湿交加，丙火暖土生身首要"},
		{XiShen: "丁", JiShen: "壬", Reason: "丁火温土，丑月比亥更寒丁火为主"},
	}

	// ========== 庚金 (stem index 6) ==========
	TiaohouData[6][0] = []TiaohouRule{ // 寅月
		{XiShen: "丁", JiShen: "丙", Reason: "寅月木旺金衰，丁火煅庚暖局为用神首要"},
		{XiShen: "甲", JiShen: "乙", Reason: "甲木生丁煅金，寅月甲木司权丁甲相辅"},
	}
	TiaohouData[6][1] = []TiaohouRule{ // 卯月
		{XiShen: "丁", JiShen: "甲", Reason: "卯月木旺金死，丁火煅庚煅甲为用"},
		{XiShen: "甲", JiShen: "乙", Reason: "甲木生丁，卯月甲木刃地甲丁并用贵显"},
	}
	TiaohouData[6][2] = []TiaohouRule{ // 辰月
		{XiShen: "丁", JiShen: "壬", Reason: "辰月湿土生金，丁火暖局煅金为用"},
		{XiShen: "甲", JiShen: "癸", Reason: "甲木疏土生丁，辰月甲木进气可用"},
	}
	TiaohouData[6][3] = []TiaohouRule{ // 巳月
		{XiShen: "壬", JiShen: "丙", Reason: "巳月火旺金瘟，壬水洗金除瘟为急"},
		{XiShen: "戊", JiShen: "丁", Reason: "戊土生金泄火，巳月火旺金衰用壬戊并"},
	}
	// 经典依据：穷通宝鉴 午月庚金："五月庚金，壬水为主，丁火佐之。丁是配合用神，非忌神。"
	TiaohouData[6][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "己", Reason: "午月火旺金败，壬水制火洗金为第一要义"},
		{XiShen: "丁", JiShen: "己", Reason: "丁火煅金，午月金将熔非丁无以成器"},
	}
	TiaohouData[6][5] = []TiaohouRule{ // 未月
		{XiShen: "壬", JiShen: "丁", Reason: "未月火气余烬，壬水洗金润局"},
		{XiShen: "丁", JiShen: "己", Reason: "丁火煅金，未月金将出头用丁煅煅"},
	}
	TiaohouData[6][6] = []TiaohouRule{ // 申月
		{XiShen: "丁", JiShen: "壬", Reason: "申月金旺土虚，丁火煅金温局为急"},
		{XiShen: "甲", JiShen: "丙", Reason: "甲木生丁煅金，申月甲木绝地用甲丁相辅"},
	}
	TiaohouData[6][7] = []TiaohouRule{ // 酉月
		{XiShen: "丁", JiShen: "辛", Reason: "酉月辛金当令，丁火煅炼庚金为用"},
	}
	TiaohouData[6][8] = []TiaohouRule{ // 戌月
		{XiShen: "丁", JiShen: "辛", Reason: "戌月土旺金相，丁火煅金温局为用"},
		{XiShen: "甲", JiShen: "壬", Reason: "甲木生丁疏土，戌月甲木养于戌土中可用"},
	}
	TiaohouData[6][9] = []TiaohouRule{ // 亥月
		{XiShen: "丁", JiShen: "丙", Reason: "亥月水冷金寒，丁火暖局煅金为急"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火温金，亥月金沉水底用丙火照暖"},
	}
	TiaohouData[6][10] = []TiaohouRule{ // 子月
		{XiShen: "丁", JiShen: "丙", Reason: "子月寒水冻结，丁火暖金煅庚为第一要义"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火温金，子月冰冻之地非丙火无以解冻"},
	}
	TiaohouData[6][11] = []TiaohouRule{ // 丑月
		{XiShen: "丁", JiShen: "丙", Reason: "丑月寒湿交加，丁火暖金煅庚为急"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火温金，丑月己土卑湿用丙火暖局"},
	}

	// ========== 辛金 (stem index 7) ==========
	TiaohouData[7][0] = []TiaohouRule{ // 寅月
		{XiShen: "壬", JiShen: "丙", Reason: "寅月木旺金衰，壬水洗金泄木之秀"},
		{XiShen: "甲", JiShen: "丁", Reason: "甲木生壬疏土，寅月壬水绝地用壬甲相生"},
	}
	TiaohouData[7][1] = []TiaohouRule{ // 卯月
		{XiShen: "壬", JiShen: "乙", Reason: "卯月木旺金死，壬水洗金泄秀"},
		{XiShen: "甲", JiShen: "丁", Reason: "甲木生壬水，卯月甲木司权丁火煅金"},
	}
	TiaohouData[7][2] = []TiaohouRule{ // 辰月
		{XiShen: "壬", JiShen: "戊", Reason: "辰月湿土生金，壬水洗金润土"},
		{XiShen: "甲", JiShen: "庚", Reason: "甲木疏土生壬，辰月甲木余气可用壬甲相生"},
	}
	TiaohouData[7][3] = []TiaohouRule{ // 巳月
		{XiShen: "壬", JiShen: "戊", Reason: "巳月火旺金瘟，壬水洗金除瘟为急"},
		{XiShen: "甲", JiShen: "丙", Reason: "甲木生壬制火，巳月甲木病地用壬甲并"},
	}
	TiaohouData[7][4] = []TiaohouRule{ // 午月
		{XiShen: "壬", JiShen: "己", Reason: "午月火旺金败，壬水制火洗金为先"},
		{XiShen: "癸", JiShen: "丙", Reason: "癸水润金，午月水绝地用壬癸并"},
	}
	TiaohouData[7][5] = []TiaohouRule{ // 未月
		{XiShen: "壬", JiShen: "丁", Reason: "未月火气余烬，壬水洗金润局"},
		{XiShen: "癸", JiShen: "丙", Reason: "癸水滋金，未月金将出头用癸水润"},
	}
	TiaohouData[7][6] = []TiaohouRule{ // 申月
		{XiShen: "壬", JiShen: "戊", Reason: "申月金旺水生，壬水洗金润局为急"},
		{XiShen: "甲", JiShen: "丁", Reason: "甲木生壬水，申月甲木绝地用甲壬相生"},
	}
	TiaohouData[7][7] = []TiaohouRule{ // 酉月
		{XiShen: "壬", JiShen: "癸", Reason: "酉月辛金当令，壬水洗金润局为用"},
		{XiShen: "甲", JiShen: "丁", Reason: "甲木生壬水，酉月辛金自旺用壬甲相生"},
	}
	TiaohouData[7][8] = []TiaohouRule{ // 戌月
		{XiShen: "壬", JiShen: "丁", Reason: "戌月土旺金相，壬水润金洗土"},
		{XiShen: "甲", JiShen: "戊", Reason: "甲木疏土生壬，戌月甲木养于戌可用"},
	}
	TiaohouData[7][9] = []TiaohouRule{ // 亥月
		{XiShen: "壬", JiShen: "丙", Reason: "亥月水冷金寒，壬水洗金温局为急"},
		{XiShen: "丙", JiShen: "丁", Reason: "丙火温金，亥月金沉水底用丙火照暖"},
	}
	TiaohouData[7][10] = []TiaohouRule{ // 子月
		{XiShen: "壬", JiShen: "丙", Reason: "子月寒水冻结，壬水洗金温局为急"},
		{XiShen: "丙", JiShen: "丁", Reason: "丙火温金，子月冰冻之地非丙火不解冻"},
	}
	TiaohouData[7][11] = []TiaohouRule{ // 丑月
		{XiShen: "壬", JiShen: "丙", Reason: "丑月寒湿交加，壬水洗金温局为急"},
		{XiShen: "丙", JiShen: "丁", Reason: "丙火温金，丑月己土卑湿用丙火暖局"},
	}

	// ========== 壬水 (stem index 8) ==========
	TiaohouData[8][0] = []TiaohouRule{ // 寅月
		{XiShen: "戊", JiShen: "丙", Reason: "寅月木旺水缩，戊土制水培木为用"},
		{XiShen: "辛", JiShen: "丁", Reason: "辛金生水，寅月辛金绝地用辛金发源"},
	}
	TiaohouData[8][1] = []TiaohouRule{ // 卯月
		{XiShen: "戊", JiShen: "辛", Reason: "卯月木旺水缩，戊土制水培木为急"},
		{XiShen: "辛", JiShen: "庚", Reason: "辛金生水，卯月甲木死地辛金绝地并用"},
	}
	TiaohouData[8][2] = []TiaohouRule{ // 辰月
		{XiShen: "丙", JiShen: "辛", Reason: "辰月水库，壬水得地，丙火温局暖水为急"},
		{XiShen: "辛", JiShen: "戊", Reason: "辛金生水，辰月辛金进气用丙辛相辅"},
	}
	TiaohouData[8][3] = []TiaohouRule{ // 巳月
		{XiShen: "辛", JiShen: "丙", Reason: "巳月火旺水涸，辛金发水源养壬为先"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水比助，巳月水绝地专赖辛壬发源"},
	}
	TiaohouData[8][4] = []TiaohouRule{ // 午月
		{XiShen: "辛", JiShen: "丙", Reason: "午月火旺水涸，辛金为用神，壬水为副"},
		{XiShen: "壬", JiShen: "癸", Reason: "壬水比助，午月水绝地非辛金发源不可"},
	}
	TiaohouData[8][5] = []TiaohouRule{ // 未月
		{XiShen: "辛", JiShen: "甲", Reason: "未月土旺水弱，辛金发源为急"},
		{XiShen: "壬", JiShen: "癸", Reason: "壬水比助，未月水将绝地专赖辛金发源"},
	}
	TiaohouData[8][6] = []TiaohouRule{ // 申月
		{XiShen: "丁", JiShen: "丙", Reason: "申月金水相生，壬水当令，丁火暖局为急"},
		{XiShen: "丙", JiShen: "丁", Reason: "丙火温金，申月金冷水冷用丁火调候"},
	}
	TiaohouData[8][7] = []TiaohouRule{ // 酉月
		{XiShen: "辛", JiShen: "丁", Reason: "酉月金旺水清，辛金生水为用神首要"},
		{XiShen: "丁", JiShen: "丙", Reason: "丁火温金，酉月辛金当令用丁火配合"},
	}
	TiaohouData[8][8] = []TiaohouRule{ // 戌月
		{XiShen: "辛", JiShen: "丁", Reason: "戌月土旺水弱，辛金发水源为急"},
		{XiShen: "甲", JiShen: "戊", Reason: "甲木疏土制水，戌月土燥用甲木化之"},
	}
	// 经典依据：穷通宝鉴 亥月壬水："十月壬水，丙火为主，戊土佐之。丙是配合用神。"
	TiaohouData[8][9] = []TiaohouRule{ // 亥月
		{XiShen: "戊", JiShen: "壬", Reason: "亥月水旺木漂，戊土制水培木为急，丙是配合用神"},
		{XiShen: "丙", JiShen: "辛", Reason: "丙火温木化水，亥月甲木长生用丙火温局"},
	}
	TiaohouData[8][10] = []TiaohouRule{ // 子月
		{XiShen: "戊", JiShen: "壬", Reason: "子月水旺用事汪洋之势，必用戊土作堤防泛滥，穷通宝鉴：十一月壬水必用戊土"},
		{XiShen: "丙", JiShen: "壬", Reason: "丙火暖局助戊，戊丙齐透富贵极品，仅丙无戊仅得温饱"},
	}
	TiaohouData[8][11] = []TiaohouRule{ // 丑月
		{XiShen: "戊", JiShen: "壬", Reason: "丑月水旺寒湿，戊土堤防止泛滥，大寒前同子月必用戊土"},
		{XiShen: "丙", JiShen: "辛", Reason: "丙火暖局助戊，大寒后丙火为急，戊丙齐透富贵"},
	}

	// ========== 癸水 (stem index 9) ==========
	TiaohouData[9][0] = []TiaohouRule{ // 寅月
		{XiShen: "丙", JiShen: "辛", Reason: "寅月余寒未退，丙火温局暖癸水为急"},
		{XiShen: "辛", JiShen: "壬", Reason: "辛金生水，寅月辛金绝地赖丙火温之"},
	}
	TiaohouData[9][1] = []TiaohouRule{ // 卯月
		{XiShen: "丙", JiShen: "辛", Reason: "卯月木旺水缩，丙火温局化木为急"},
		{XiShen: "辛", JiShen: "壬", Reason: "辛金生水，卯月甲木死地用丙辛温局"},
	}
	TiaohouData[9][2] = []TiaohouRule{ // 辰月
		{XiShen: "丙", JiShen: "辛", Reason: "辰月湿土司令，丙火暖局生身为急"},
		{XiShen: "辛", JiShen: "壬", Reason: "辛金生水，辰月癸水入墓用丙辛温之"},
	}
	TiaohouData[9][3] = []TiaohouRule{ // 巳月
		{XiShen: "辛", JiShen: "丙", Reason: "巳月火旺水涸，辛金发水源养癸为急"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水助癸，巳月水绝地专赖辛金发源"},
	}
	TiaohouData[9][4] = []TiaohouRule{ // 午月
		{XiShen: "辛", JiShen: "丙", Reason: "午月火旺水涸，辛金为用神，壬水为副"},
		{XiShen: "壬", JiShen: "癸", Reason: "壬水比助，午月水绝非辛金发源不可"},
	}
	TiaohouData[9][5] = []TiaohouRule{ // 未月
		{XiShen: "辛", JiShen: "丙", Reason: "未月土旺水弱，辛金发源为急"},
		{XiShen: "庚", JiShen: "丁", Reason: "庚金生水，未月水将绝地赖庚辛发源"},
	}
	TiaohouData[9][6] = []TiaohouRule{ // 申月
		{XiShen: "丁", JiShen: "丙", Reason: "申月金水相生，癸水得地，丁火暖局为急"},
		{XiShen: "丙", JiShen: "辛", Reason: "丙火温金，申月金冷水冷用丁火调候"},
	}
	TiaohouData[9][7] = []TiaohouRule{ // 酉月
		{XiShen: "辛", JiShen: "丁", Reason: "酉月金旺水清，辛金生水为用神首要"},
		{XiShen: "丁", JiShen: "丙", Reason: "丁火温金，酉月辛金当令用丁火配合"},
	}
	TiaohouData[9][8] = []TiaohouRule{ // 戌月
		{XiShen: "辛", JiShen: "丁", Reason: "戌月土旺水弱，辛金发源为急"},
		{XiShen: "壬", JiShen: "戊", Reason: "壬水比助，戌月水将绝地赖辛金发源"},
	}
	TiaohouData[9][9] = []TiaohouRule{ // 亥月
		{XiShen: "丙", JiShen: "辛", Reason: "亥月水冷木漂，丙火温局化木为急"},
		{XiShen: "辛", JiShen: "壬", Reason: "辛金生水，亥月甲木长生用丙火温之"},
	}
	TiaohouData[9][10] = []TiaohouRule{ // 子月
		{XiShen: "丙", JiShen: "丁", Reason: "子月寒水冻结，丙火解冻暖局为急"},
		{XiShen: "丁", JiShen: "辛", Reason: "丁火温局，子月癸水得令用丙丁温之"},
	}
	TiaohouData[9][11] = []TiaohouRule{ // 丑月
		{XiShen: "丙", JiShen: "丁", Reason: "丑月寒湿交加，丙火暖局解冻为急"},
		{XiShen: "丁", JiShen: "辛", Reason: "丁火温局，丑月己土卑湿用丙丁暖水"},
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