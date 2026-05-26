package service

import (
	"fmt"
	"strings"
)
type TenGodAnalysis struct {
	DominantGod     string        `json:"dominant_god"`
	DominantPercent  float64       `json:"dominant_percent"`
	Personality      string        `json:"personality"`
	Interpersonal    string        `json:"interpersonal"`
	CareerFortune    string        `json:"career_fortune"`
	EmotionRelation  string        `json:"emotion_relation"`
	HealthNote       string        `json:"health_note"`
	Taboo            string        `json:"taboo"`
	GodRelations     []GodRelation `json:"god_relations"`
	Summary          string        `json:"summary"`
}

// GodRelation describes one ten god's meaning and advice in this chart.
type GodRelation struct {
	God     string `json:"god"`
	Percent string `json:"percent"`
	Meaning string `json:"meaning"`
	Advice  string `json:"advice"`
}

// TenGodAnalyzer generates TenGodAnalysis from TenGodProportion data.
type TenGodAnalyzer struct{}

// AnalyzeTenGod generates a comprehensive TenGodAnalysis from proportion data.
func (a *TenGodAnalyzer) AnalyzeTenGod(proportions []TenGodRatio) *TenGodAnalysis {
	// Build lookup map
	m := make(map[string]TenGodRatio)
	for _, p := range proportions {
		m[p.Name] = p
	}

	// Find dominant god
	var dominant string
	var dominantPct float64
	for _, p := range proportions {
		if p.Percent > dominantPct {
			dominantPct = p.Percent
			dominant = p.Name
		}
	}

	// Build per-god relation details
	godRelations := a.buildGodRelations(proportions, m)

	// Personality summary based on dominant god + key combinations
	personality := a.buildPersonality(m, dominant)
	interpersonal := a.buildInterpersonal(m)
	careerFortune := a.buildCareer(m)
	emotionRelation := a.buildEmotion(m)
	healthNote := a.buildHealth(m)
	taboo := a.buildTaboo(m, dominant)
	summary := a.buildSummary(m, dominant, dominantPct)

	return &TenGodAnalysis{
		DominantGod:    dominant,
		DominantPercent: dominantPct,
		Personality:     personality,
		Interpersonal:   interpersonal,
		CareerFortune:   careerFortune,
		EmotionRelation: emotionRelation,
		HealthNote:      healthNote,
		Taboo:           taboo,
		GodRelations:    godRelations,
		Summary:         summary,
	}
}

func (a *TenGodAnalyzer) buildGodRelations(proportions []TenGodRatio, m map[string]TenGodRatio) []GodRelation {
	var relations []GodRelation
	for _, p := range proportions {
		if p.Percent == 0 {
			continue
		}
		rel := GodRelation{
			God:     p.Name,
			Percent: formatPercent(p.Percent),
			Meaning: tenGodMeaning(p.Name, p.Percent),
			Advice:  tenGodAdvice(p.Name, p.Percent),
		}
		relations = append(relations, rel)
	}
	return relations
}

// formatPercent returns a clean percentage string.
func formatPercent(p float64) string {
	return fmt.Sprintf("%.1f%%", p)
}

// tenGodMeaning gives the meaning of this ten god in the chart.
func tenGodMeaning(god string, pct float64) string {
	switch god {
	case "比肩":
		if pct >= 20 {
			return "比肩旺者，主独立自信、意志坚定、善于竞争。做事果断，不喜依赖他人，内心有一股不服输的劲。"
		}
		return "比肩适度，主独立自主、善于合作。在团队中能保持自我，不随波逐流。"
	case "劫财":
		if pct >= 20 {
			return "劫财旺者，主勇于冒险、讲义气、社交能力强。敢于拼搏，常有出其不意的行动。"
		}
		return "劫财适度，主热情大方、善于交际。遇到困难时能获得他人帮助。"
	case "食神":
		if pct >= 20 {
			return "食神旺者，主乐观开朗、表达力强、创造力丰富。善于享受生活，财运与福气较好。"
		}
		return "食神适度，主温和友善、享受生活。在艺术或技艺方面有一定天赋。"
	case "伤官":
		if pct >= 20 {
			return "伤官旺者，主才华横溢、思维敏捷、不服输。敢于表现，创业能力强，但易冲动。"
		}
		return "伤官适度，主才华出众、表达力强。在创意行业或技术领域有发展潜力。"
	case "正财":
		if pct >= 20 {
			return "正财旺者，主务实节俭、理财稳重。财富通过正当努力积累，生活安稳。"
		}
		return "正财适度，主踏实努力、财运稳定。适合稳定工作或固定投资。"
	case "偏财":
		if pct >= 20 {
			return "偏财旺者，主大方豪爽、社交广、投机心强。善于理财但需注意风险。"
		}
		return "偏财适度，主社交能力强、有偏财运。适合合伙或投资类事业。"
	case "正官":
		if pct >= 20 {
			return "正官旺者，主正直有责任感、守规矩、事业心强。仕途发展顺利，名声较好。"
		}
		return "正官适度，主为人正派、事业稳定。在管理或体制内工作有发展。"
	case "七杀":
		if pct >= 20 {
			return "七杀旺者，主果断强势、魄力大、压力亦大。创业能力强，但需注意健康。"
		}
		return "七杀适度，主有魄力、敢于挑战。在竞争环境中能脱颖而出。"
	case "正印":
		if pct >= 20 {
			return "正印旺者，主善良稳重、学识渊博、贵人运旺。学业或学术方面有天赋。"
		}
		return "正印适度，主温和善良、有学识。适合教育或学术研究类工作。"
	case "偏印":
		if pct >= 20 {
			return "偏印旺者，主悟性高、敏感细腻、悟道心强。学术或技术研究有独特见解。"
		}
		return "偏印适度，主悟性高、善于思考。在专业技术领域有发展潜力。"
	}
	return ""
}

// tenGodAdvice gives advice based on the ten god's proportion.
func tenGodAdvice(god string, pct float64) string {
	switch god {
	case "比肩":
		if pct >= 20 {
			return "扬长避短：发挥独立自信的优势，但需学会借力，避免过度固执而树敌。"
		}
		return "保持独立精神的同时注重团队合作，适当听取他人意见。"
	case "劫财":
		if pct >= 20 {
			return "注意控制冲动和冒险倾向，理财时需谨慎评估风险，避免合伙纠纷。"
		}
		return "发挥社交能力同时注重信用，财务往来要有凭据。"
	case "食神":
		if pct >= 20 {
			return "发挥创意和表达天赋，但需注意健康管理，避免过度享受导致健康问题。"
		}
		return "善用乐观开朗的性格，但需在重要事项上多加思考。"
	case "伤官":
		if pct >= 20 {
			return "才华需以合适方式展现，收敛冲动和叛逆情绪，创业需谨慎评估。"
		}
		return "发挥才华同时学会妥协，感情上需多沟通避免争执。"
	case "正财":
		if pct >= 20 {
			return "财运稳定，宜稳扎稳打，但需开阔视野，适度参与投资以增值。"
		}
		return "继续保持务实节俭作风，可适当提升理财能力。"
	case "偏财":
		if pct >= 20 {
			return "投资理财需分散风险，不要过于自信或投机，合伙生意要明确权责。"
		}
		return "发挥社交能力的同时注意控制消费欲望，理性投资。"
	case "正官":
		if pct >= 20 {
			return "事业发展顺利，但需注意不要过于追求权力而失去本心，健康也要关注。"
		}
		return "发挥正直品质的同时学会灵活变通，平衡事业与家庭。"
	case "七杀":
		if pct >= 20 {
			return "注意心血管健康，学会放松减压，做重大决策前多听听他人意见。"
		}
		return "发挥魄力的同时注意方法和节奏，避免过度竞争。"
	case "正印":
		if pct >= 20 {
			return "学业事业均能得到贵人相助，但需主动出击，不要过于依赖他人。"
		}
		return "发挥学识优势的同时注重实践，保持善良但要有原则。"
	case "偏印":
		if pct >= 20 {
			return "悟性高但容易想太多而焦虑，建议培养兴趣爱好放松心神，避免精神过耗。"
		}
		return "发挥独特思考能力的同时多与他人交流，避免过于封闭。"
	}
	return ""
}

func (a *TenGodAnalyzer) buildPersonality(m map[string]TenGodRatio, dominant string) string {
	bijiao := m["比肩"].Percent
	jiecai := m["劫财"].Percent
	shishen := m["食神"].Percent
	shangguan := m["伤官"].Percent
	zhengcai := m["正财"].Percent
	piancai := m["偏财"].Percent
	zhengguan := m["正官"].Percent
	qisha := m["七杀"].Percent
	zhengyin := m["正印"].Percent
	pianyin := m["偏印"].Percent

	// Build personality description
	var traits []string

	// 印星
	if zhengyin >= 15 {
		traits = append(traits, "善良稳重、学识渊博")
	} else if pianyin >= 15 {
		traits = append(traits, "悟性高、善于思考")
	} else if zhengyin+pianyin >= 20 {
		traits = append(traits, "学识与悟性兼备")
	}

	// 官星
	if zhengguan >= 15 {
		traits = append(traits, "正直有责任感、守规矩")
	} else if qisha >= 15 {
		traits = append(traits, "果断强势、有魄力")
	} else if zhengguan+qisha >= 20 {
		traits = append(traits, "正直且有魄力")
	}

	// 财星
	if zhengcai >= 15 {
		traits = append(traits, "务实节俭、理财稳重")
	} else if piancai >= 15 {
		traits = append(traits, "大方豪爽、社交能力强")
	} else if zhengcai+piancai >= 20 {
		traits = append(traits, "理财与社交能力兼具")
	}

	// 比劫
	if bijiao >= 15 && jiecai >= 15 {
		traits = append(traits, "独立自信、意志坚定")
	} else if jiecai >= 20 {
		traits = append(traits, "勇于冒险、讲义气")
	} else if bijiao >= 15 {
		traits = append(traits, "独立自主、善于合作")
	}

	// 食伤
	if shishen >= 15 {
		traits = append(traits, "乐观开朗、表达力强")
	} else if shangguan >= 15 {
		traits = append(traits, "才华横溢、思维敏捷")
	} else if shishen+shangguan >= 20 {
		traits = append(traits, "创意与表达能力兼备")
	}

	if len(traits) == 0 {
		traits = append(traits, "各方面特质平衡，运势稳定")
	}

	return "你的性格主要表现为：" + strings.Join(traits, "，") + "。"
}

func (a *TenGodAnalyzer) buildInterpersonal(m map[string]TenGodRatio) string {
	bijiao := m["比肩"].Percent
	jiecai := m["劫财"].Percent
	shishen := m["食神"].Percent
	shangguan := m["伤官"].Percent
	zhengcai := m["正财"].Percent
	piancai := m["偏财"].Percent
	zhengguan := m["正官"].Percent
	qisha := m["七杀"].Percent
	zhengyin := m["正印"].Percent
	pianyin := m["偏印"].Percent

	var relations []string

	// 印星
	if zhengyin >= 15 || pianyin >= 15 {
		relations = append(relations, "容易得到长辈或贵人的帮助")
	}

	// 官星
	if zhengguan >= 15 {
		relations = append(relations, "人际关系正派，在体制内易获信任")
	} else if qisha >= 15 {
		relations = append(relations, "社交果断干脆，但也要注意给他人留有余地")
	}

	// 财星
	if zhengcai+piancai >= 30 {
		relations = append(relations, "社交能力强，人脉广泛，合伙机会多")
	} else if zhengcai >= 15 {
		relations = append(relations, "人际关系务实，重信用")
	}

	// 比劫
	if bijiao+jiecai >= 40 {
		relations = append(relations, "竞争意识强，需学会合作分享，避免树敌")
	} else if jiecai >= 20 {
		relations = append(relations, "讲义气，朋友较多，但也要注意财务往来")
	} else if bijiao >= 15 {
		relations = append(relations, "独立自信，能吸引志同道合的伙伴")
	}

	// 食伤
	if shishen+shangguan >= 25 {
		relations = append(relations, "善于表达，人际交往中能说会道，易获好感")
	}

	return strings.Join(relations, "。") + "。"
}

func (a *TenGodAnalyzer) buildCareer(m map[string]TenGodRatio) string {
	zhengyin := m["正印"].Percent
	pianyin := m["偏印"].Percent
	zhengguan := m["正官"].Percent
	qisha := m["七杀"].Percent
	zhengcai := m["正财"].Percent
	piancai := m["偏财"].Percent
	shishen := m["食神"].Percent
	shangguan := m["伤官"].Percent
	bijiao := m["比肩"].Percent
	jiecai := m["劫财"].Percent

	var career []string

	// 印星
	if zhengyin >= 15 {
		career = append(career, "适合学术、教育、文化、管理类工作")
	} else if pianyin >= 15 {
		career = append(career, "适合技术研究、专业技能型工作")
	} else if zhengyin+pianyin >= 25 {
		career = append(career, "学以致用型，适合需要专业知识的领域")
	}

	// 官星
	if zhengguan >= 20 {
		career = append(career, "仕途发展顺利，适合管理或行政类岗位")
	} else if qisha >= 20 {
		career = append(career, "魄力强，适合创业开拓或高压工作")
	} else if zhengguan+qisha >= 25 {
		career = append(career, "管理与开拓能力兼备，适合管理层或创业")
	}

	// 财星
	if zhengcai >= 20 {
		career = append(career, "财运稳定，适合固定收入的工作或稳健理财")
	} else if piancai >= 20 {
		career = append(career, "偏财运佳，适合投资、合伙或业务拓展类工作")
	} else if zhengcai+piancai >= 25 {
		career = append(career, "正偏财兼具，财务管理与投资能力都较强")
	}

	// 食伤
	if shishen >= 15 {
		career = append(career, "财运与福气较好，适合技艺或服务业")
	} else if shangguan >= 15 {
		career = append(career, "创业能力强，适合创意或技术创业")
	} else if shishen+shangguan >= 25 {
		career = append(career, "创造与表达能力突出，适合创意产业")
	}

	// 比劫
	if bijiao+jiecai >= 40 {
		career = append(career, "竞争激烈，需加强合作或寻求合伙")
	} else if bijiao >= 15 {
		career = append(career, "独立自主，适合自由职业或技术专精方向")
	} else if jiecai >= 20 {
		career = append(career, "敢于拼搏，合伙创业可发挥优势")
	}

	return strings.Join(career, "。") + "。"
}

func (a *TenGodAnalyzer) buildEmotion(m map[string]TenGodRatio) string {
	zhengguan := m["正官"].Percent
	qisha := m["七杀"].Percent
	zhengcai := m["正财"].Percent
	piancai := m["偏财"].Percent
	shishen := m["食神"].Percent
	shangguan := m["伤官"].Percent
	zhengyin := m["正印"].Percent
	pianyin := m["偏印"].Percent
	bijiao := m["比肩"].Percent
	jiecai := m["劫财"].Percent

	var emotion []string

	// 官星
	if zhengguan >= 15 {
		emotion = append(emotion, "传统观念强，重视婚姻稳定，姻缘幸福")
	} else if qisha >= 15 {
		emotion = append(emotion, "感情果断干脆，可能聚少离多或感情路有挑战")
	} else if zhengguan+qisha >= 20 {
		emotion = append(emotion, "感情上有责任感，但也要学会包容")
	}

	// 财星
	if zhengcai >= 20 {
		emotion = append(emotion, "重视物质基础，感情务实，倾向于传统婚姻")
	} else if piancai >= 20 {
		emotion = append(emotion, "社交广泛，异性缘分较多，可能有异地姻缘")
	} else if piancai >= 15 {
		emotion = append(emotion, "感情细腻，异地或社交姻缘机会较多")
	}

	// 食伤
	if shishen >= 15 {
		emotion = append(emotion, "感情甜蜜，生活愉快，但需注意健康管理")
	} else if shangguan >= 15 {
		emotion = append(emotion, "感情上多波折，才华吸引异性，需学会处理感情")
	} else if shishen+shangguan >= 20 {
		emotion = append(emotion, "感情丰富多彩，表达能力强，异性缘佳")
	}

	// 印星
	if zhengyin >= 15 {
		emotion = append(emotion, "婚姻幸福平稳，易得配偶帮助")
	} else if pianyin >= 15 {
		emotion = append(emotion, "感情细腻敏感，注重精神交流")
	}

	// 比劫
	if bijiao+jiecai >= 35 {
		emotion = append(emotion, "感情上竞争意识强，需注意处理竞争关系")
	} else if jiecai >= 20 {
		emotion = append(emotion, "讲义气，感情中敢打敢拼，但也要学会妥协")
	} else if bijiao >= 15 {
		emotion = append(emotion, "独立自主，感情上有主见，倾向于晚婚")
	}

	return strings.Join(emotion, "。") + "。"
}

func (a *TenGodAnalyzer) buildHealth(m map[string]TenGodRatio) string {
	jiecai := m["劫财"].Percent
	shishen := m["食神"].Percent
	shangguan := m["伤官"].Percent
	zhengcai := m["正财"].Percent
	piancai := m["偏财"].Percent
	qisha := m["七杀"].Percent
	zhengguan := m["正官"].Percent
	zhengyin := m["正印"].Percent
	pianyin := m["偏印"].Percent
	bijiao := m["比肩"].Percent

	var health []string

	if pianyin >= 15 || zhengyin >= 15 {
		health = append(health, "肺部")
	}

	if piancai >= 20 || zhengcai >= 20 || jiecai >= 20 {
		health = append(health, "肝胆")
	}

	if piancai >= 15 || zhengcai >= 15 {
		health = append(health, "消化系统")
	}

	if qisha >= 15 {
		health = append(health, "心血管")
	}

	if zhengguan >= 15 {
		health = append(health, "肾/泌尿系统")
	}

	if shishen >= 15 {
		health = append(health, "消化系统")
	}

	if shangguan >= 15 {
		health = append(health, "神经系统")
	}

	if jiecai >= 25 {
		health = append(health, "肾")
	}

	if bijiao >= 20 {
		health = append(health, "肝胆")
	}

	if len(health) == 0 {
		return "健康运势总体平稳，无明显倾向，注意日常保养即可。"
	}

	unique := removeDuplicates(health)
	return "需要注意的身体部位：" + strings.Join(unique, "、") + "。建议定期体检，保持良好生活习惯。"
}

func (a *TenGodAnalyzer) buildTaboo(m map[string]TenGodRatio, dominant string) string {
	var taboos []string

	switch dominant {
	case "劫财":
		taboos = append(taboos, "避免冲动投资和赌博", "财务往来要有凭证，合伙需签合同", "注意控制脾气，避免口角")
	case "七杀":
		taboos = append(taboos, "避免过度压力和熬夜", "心脑血管问题不容忽视", "重大决策切勿草率")
	case "伤官":
		taboos = append(taboos, "避免言语过激伤害他人", "感情上不要过于强势", "创业前要做好充分准备")
	case "偏印":
		taboos = append(taboos, "避免思虑过度和精神内耗", "不要封闭自己，多与人交流", "培养兴趣爱好放松身心")
	case "偏财":
		taboos = append(taboos, "避免贪心和投机型投资", "合伙生意要明确权责", "财务决策要冷静")
	case "比肩":
		taboos = append(taboos, "避免过度固执不听劝", "要学会借力和合作", "竞争中勿伤和气")
	case "正官":
		taboos = append(taboos, "避免过于追求权力", "不要过于保守而错失机会", "健康与家庭需关注")
	case "食神":
		taboos = append(taboos, "避免过度享受和暴饮暴食", "健康管理不容忽视", "工作娱乐要平衡")
	case "正财":
		taboos = append(taboos, "避免过于保守而不敢投资", "要开阔视野，学习新知", "理财要多元化")
	case "正印":
		taboos = append(taboos, "避免过于依赖他人", "要主动出击争取机会", "实践中检验学识")
	}

	return strings.Join(taboos, "。") + "。"
}

func (a *TenGodAnalyzer) buildSummary(m map[string]TenGodRatio, dominant string, dominantPct float64) string {
	bijiao := m["比肩"].Percent
	jiecai := m["劫财"].Percent
	zhengguan := m["正官"].Percent
	qisha := m["七杀"].Percent
	zhengyin := m["正印"].Percent
	pianyin := m["偏印"].Percent

	// Check for special combinations
	var combo []string

	if jiecai >= 20 && zhengguan >= 15 && zhengyin >= 15 {
		combo = append(combo, "劫财+正官+正印 组合，各方面运势相对平衡")
	}

	if bijiao+jiecai >= 40 {
		combo = append(combo, "比劫过旺，竞争意识强，需学会合作")
	}

	if zhengguan >= 15 && qisha >= 15 {
		combo = append(combo, "正官七杀同现，权力欲强，压力亦大")
	}

	if zhengyin >= 15 && pianyin >= 15 {
		combo = append(combo, "正偏印同现，学识与悟性兼备，适合学术研究")
	}

	comboStr := ""
	if len(combo) > 0 {
		comboStr = strings.Join(combo, "，") + "。"
	}

	return fmt.Sprintf("你的命盘中最强的十神是「%s」，占比%.1f%%。%s综合来看，你的运势较为均衡，只要顺势而为，假以时日必有所成。",
		dominant, dominantPct, comboStr)
}

func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	var result []string
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}