package bazi

// ShenShaMeta describes display and ranking metadata for one shen-sha item.
type ShenShaMeta struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Polarity    string `json:"polarity"`
	Priority    int    `json:"priority"`
	Source      string `json:"source"`
	Description string `json:"description"`
}

var shenShaMetaCatalog = map[string]ShenShaMeta{
	"天乙贵人": {Name: "天乙贵人", Category: "贵人", Polarity: "吉", Priority: 1, Source: "日干/年干贵人位", Description: "主贵人扶助、遇难得解。"},
	"太极贵人": {Name: "太极贵人", Category: "贵人", Polarity: "吉", Priority: 2, Source: "年干/日干取支", Description: "主聪慧、悟性、宗教玄学缘。"},
	"文昌贵人": {Name: "文昌贵人", Category: "才学", Polarity: "吉", Priority: 1, Source: "日干取支", Description: "主学习、文书、表达与考试。"},
	"福星贵人": {Name: "福星贵人", Category: "福德", Polarity: "吉", Priority: 2, Source: "日干取支", Description: "主福气、照拂与顺遂。"},
	"国印贵人": {Name: "国印贵人", Category: "权印", Polarity: "吉", Priority: 2, Source: "日干取支", Description: "主印信、资质、名誉与制度资源。"},
	"天德贵人": {Name: "天德贵人", Category: "福德", Polarity: "吉", Priority: 1, Source: "月令取干支", Description: "主化解、德荫、逢凶减轻。"},
	"月德贵人": {Name: "月德贵人", Category: "福德", Polarity: "吉", Priority: 1, Source: "月令取干", Description: "主和气、解厄、贵人照应。"},
	"禄神":   {Name: "禄神", Category: "财禄", Polarity: "吉", Priority: 1, Source: "日干禄位", Description: "主衣禄、职位、资源与稳定收益。"},
	"金舆":   {Name: "金舆", Category: "财禄", Polarity: "吉", Priority: 2, Source: "日干取支", Description: "主享受、车舆、伴侣助力。"},
	"学堂":   {Name: "学堂", Category: "才学", Polarity: "吉", Priority: 2, Source: "日干取支", Description: "主学习能力、传承与文凭。"},
	"词馆":   {Name: "词馆", Category: "才学", Polarity: "吉", Priority: 2, Source: "日干取支", Description: "主文辞、表达、写作与声名。"},
	"天厨食禄": {Name: "天厨食禄", Category: "财禄", Polarity: "吉", Priority: 3, Source: "日干取支", Description: "主饮食、口福、资源供给。"},
	"驿马":   {Name: "驿马", Category: "动象", Polarity: "中", Priority: 2, Source: "三合局取冲位", Description: "主移动、出差、变动、消息。"},
	"华盖":   {Name: "华盖", Category: "才艺", Polarity: "中", Priority: 2, Source: "三合局墓位", Description: "主才艺、孤高、宗教玄学缘。"},
	"桃花":   {Name: "桃花", Category: "人缘", Polarity: "中", Priority: 2, Source: "三合局咸池", Description: "主人缘、审美、感情与社交吸引。"},
	"咸池":   {Name: "咸池", Category: "人缘", Polarity: "中", Priority: 2, Source: "三合局桃花位", Description: "主人缘桃花，也需防情感扰动。"},
	"红鸾":   {Name: "红鸾", Category: "婚恋", Polarity: "吉", Priority: 2, Source: "年支取位", Description: "主婚恋喜庆、人际和合。"},
	"天喜":   {Name: "天喜", Category: "婚恋", Polarity: "吉", Priority: 2, Source: "年支取位", Description: "主喜庆、婚恋、社交顺意。"},
	"羊刃":   {Name: "羊刃", Category: "刚烈", Polarity: "凶", Priority: 1, Source: "阳干刃位", Description: "主刚猛、竞争、冲动，宜有制化。"},
	"飞刃":   {Name: "飞刃", Category: "刚烈", Polarity: "凶", Priority: 2, Source: "羊刃对冲", Description: "主急切、伤损、冲突风险。"},
	"劫煞":   {Name: "劫煞", Category: "风险", Polarity: "凶", Priority: 1, Source: "三合局取位", Description: "主损耗、争夺、突发阻碍。"},
	"灾煞":   {Name: "灾煞", Category: "风险", Polarity: "凶", Priority: 1, Source: "年支/三合局取位", Description: "主灾扰、意外、计划受阻。"},
	"岁破":   {Name: "岁破", Category: "风险", Polarity: "凶", Priority: 1, Source: "年支对冲", Description: "主冲破、变动、外部压力。"},
	"大耗":   {Name: "大耗", Category: "耗损", Polarity: "凶", Priority: 1, Source: "年支取位", Description: "主破财、耗费、资源流失。"},
	"小耗":   {Name: "小耗", Category: "耗损", Polarity: "凶", Priority: 2, Source: "年支取位", Description: "主小额耗费、精力消耗。"},
	"白虎":   {Name: "白虎", Category: "风险", Polarity: "凶", Priority: 1, Source: "年支取位", Description: "主伤病、口舌、冲突。"},
	"官符":   {Name: "官符", Category: "是非", Polarity: "凶", Priority: 1, Source: "年支取位", Description: "主文书、规则压力、官非是非。"},
	"病符":   {Name: "病符", Category: "健康", Polarity: "凶", Priority: 2, Source: "年支取位", Description: "主健康波动、疲劳与病气。"},
	"死符":   {Name: "死符", Category: "风险", Polarity: "凶", Priority: 2, Source: "年支取位", Description: "主停滞、低迷、谨慎守成。"},
	"丧门":   {Name: "丧门", Category: "风险", Polarity: "凶", Priority: 1, Source: "年支取位", Description: "主忧烦、家宅不宁、低落气象。"},
	"吊客":   {Name: "吊客", Category: "风险", Polarity: "凶", Priority: 1, Source: "年支取位", Description: "主奔波、吊问、情绪耗损。"},
	"孤辰":   {Name: "孤辰", Category: "孤寡", Polarity: "凶", Priority: 2, Source: "三会方取位", Description: "主独立、孤高、关系疏离。"},
	"寡宿":   {Name: "寡宿", Category: "孤寡", Polarity: "凶", Priority: 2, Source: "三会方取位", Description: "主情感孤独、婚恋迟滞。"},
	"空亡":   {Name: "空亡", Category: "空转", Polarity: "中", Priority: 2, Source: "旬空", Description: "主落空、延迟、需看冲合填实。"},
	"四大空亡": {Name: "四大空亡", Category: "空转", Polarity: "凶", Priority: 1, Source: "旬空纳音", Description: "主某类五行力量虚浮或落空。"},
	"魁罡":   {Name: "魁罡", Category: "刚烈", Polarity: "中", Priority: 2, Source: "日柱特格", Description: "主刚正、决断，也忌过刚。"},
	"红艳煞":  {Name: "红艳煞", Category: "人缘", Polarity: "中", Priority: 3, Source: "日干取支", Description: "主魅力、桃色、人际吸引。"},
	"流霞":   {Name: "流霞", Category: "风险", Polarity: "凶", Priority: 3, Source: "年干取支", Description: "主血光、意外、情绪波动。"},
	"血刃":   {Name: "血刃", Category: "健康", Polarity: "凶", Priority: 2, Source: "年干取支", Description: "主刀伤血光，宜谨慎运动与器械。"},
	"血忌":   {Name: "血忌", Category: "健康", Polarity: "凶", Priority: 2, Source: "年干取支", Description: "主健康禁忌、血气不稳。"},
	"元辰":   {Name: "元辰", Category: "风险", Polarity: "凶", Priority: 2, Source: "年支与性别取位", Description: "主反复、阻隔、不顺心。"},
	"勾绞煞":  {Name: "勾绞煞", Category: "是非", Polarity: "凶", Priority: 2, Source: "年支与性别取位", Description: "主纠缠、口舌、文书牵连。"},
}

// LookupShenShaMeta returns catalog metadata with a neutral fallback.
func LookupShenShaMeta(name string) ShenShaMeta {
	if meta, ok := shenShaMetaCatalog[name]; ok {
		return meta
	}
	return ShenShaMeta{
		Name:        name,
		Category:    "提示",
		Polarity:    "中",
		Priority:    3,
		Source:      "内置神煞规则",
		Description: "按当前神煞规则触发，需结合原局旺衰、喜忌与刑冲合害综合判断。",
	}
}

func LookupShenShaMetaFromItem(item string) ShenShaMeta {
	return LookupShenShaMeta(shenShaName(item))
}

func BuildShenShaDetails(items []string) []ShenShaMeta {
	details := make([]ShenShaMeta, 0, len(items))
	for _, item := range items {
		if item == "" {
			continue
		}
		details = append(details, LookupShenShaMetaFromItem(item))
	}
	return details
}
