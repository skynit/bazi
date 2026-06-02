package bazi

import (
	"fmt"
	"strings"

	"bazi/internal/service/data"
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
// dayElem is the day master's five element (e.g. "木", "火"), used for health mapping.
// ptg provides per-pillar ten god distribution for pillar-dimension analysis.
// gender is "MALE" or "FEMALE", used for gender-specific advice.
func (a *TenGodAnalyzer) AnalyzeTenGod(proportions []TenGodRatio, dayElem string, bodyStrength BodyStrengthResult, ptg PillarTenGods, gender string) *TenGodAnalysis {
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
	isStrong := bodyStrength.Verdict == "身旺"
	godRelations := a.buildGodRelations(proportions, m, isStrong, gender)

	// Personality summary based on dominant god + key combinations
	personality := a.buildPersonality(m, dominant, ptg)
	interpersonal := a.buildInterpersonal(m, ptg)
	careerFortune := a.buildCareer(m, ptg)
	emotionRelation := a.buildEmotion(m, ptg)
	healthNote := a.buildHealth(m, dayElem)
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

func (a *TenGodAnalyzer) buildGodRelations(proportions []TenGodRatio, m map[string]TenGodRatio, isStrong bool, gender string) []GodRelation {
	var relations []GodRelation
	for _, p := range proportions {
		if p.Percent == 0 {
			continue
		}
		rel := GodRelation{
			God:     p.Name,
			Percent: formatPercent(p.Percent),
			Meaning: tenGodMeaning(p.Name, p.Percent, isStrong),
			Advice:  tenGodAdvice(p.Name, p.Percent, gender),
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
// isStrong indicates whether the day master is strong (身旺) or weak (身弱).
// 身旺时：比劫为忌（争夺），食伤为喜（泄秀），财为喜（耗身），官杀为喜（制身），印为忌（生身过旺）
// 身弱时：比劫为喜（帮身），食伤为忌（泄身），财为忌（耗身），官杀为忌（克身），印为喜（生身）
// 经典依据：《子平真诠》《滴天髓》论十神性情
func tenGodMeaning(god string, pct float64, isStrong bool) string {
	// 身旺身弱标签：用于描述十神对命主的利弊方向
	strengthTag := func(godName string) string {
		if isStrong {
			switch godName {
			case "比肩", "劫财":
				return "【忌】身旺逢比劫，争夺之象"
			case "食神", "伤官":
				return "【喜】身旺喜食伤泄秀"
			case "正财", "偏财":
				return "【喜】身旺喜财星耗身"
			case "正官", "七杀":
				return "【喜】身旺喜官杀制身"
			case "正印", "偏印":
				return "【忌】身旺忌印星生扶过度"
			}
		} else {
			switch godName {
			case "比肩", "劫财":
				return "【喜】身弱喜比劫帮身"
			case "食神", "伤官":
				return "【忌】身弱忌食伤泄身"
			case "正财", "偏财":
				return "【忌】身弱忌财星耗身"
			case "正官", "七杀":
				return "【忌】身弱忌官杀克身"
			case "正印", "偏印":
				return "【喜】身弱喜印星生扶"
			}
		}
		return ""
	}

	tag := strengthTag(god)

	switch god {
	case "比肩":
		// 经典依据：《滴天髓》"比肩自旺，兄弟争财"
		if pct >= 35 {
			return tag + "。比肩极旺，主极度独立、固执己见、竞争心极强。须防孤僻、人际关系紧张。"
		}
		if pct >= 25 {
			return tag + "。比肩很旺，主独立自信、意志坚定、善于竞争。做事果断，不喜依赖他人。"
		}
		if pct >= 15 {
			return tag + "。比肩旺者，主独立自信、意志坚定、善于竞争。做事果断，不喜依赖他人，内心有一股不服输的劲。"
		}
		return tag + "。比肩适度，主独立自主、善于合作。在团队中能保持自我，不随波逐流。"
	case "劫财":
		// 经典依据：《子平真诠》"劫财争财，义气用事"
		if pct >= 35 {
			return tag + "。劫财极旺，主极度好胜、争强好斗、易破财。须防冲动投资、合伙纠纷。"
		}
		if pct >= 25 {
			return tag + "。劫财很旺，主勇于冒险、讲义气、社交能力强。敢于拼搏，常有出其不意的行动。"
		}
		if pct >= 15 {
			return tag + "。劫财旺者，主勇于冒险、讲义气、社交能力强。敢于拼搏，常有出其不意的行动。"
		}
		return tag + "。劫财适度，主热情大方、善于交际。遇到困难时能获得他人帮助。"
	case "食神":
		// 经典依据：《滴天髓》"食神有气胜财官"
		if pct >= 35 {
			return tag + "。食神极旺，主过度享受、贪图安逸、口福过盛。须防懒散、健康问题。"
		}
		if pct >= 25 {
			return tag + "。食神很旺，主乐观开朗、表达力强、创造力丰富。善于享受生活，财运与福气较好。"
		}
		if pct >= 15 {
			return tag + "。食神旺者，主乐观开朗、表达力强、创造力丰富。善于享受生活，财运与福气较好。"
		}
		return tag + "。食神适度，主温和友善、享受生活。在艺术或技艺方面有一定天赋。"
	case "伤官":
		// 经典依据：《子平真诠》"伤官见官为祸百端"
		if pct >= 35 {
			return tag + "。伤官极旺，主锋芒毕露、言语尖刻、叛逆心强。易得罪人，感情多波折。"
		}
		if pct >= 25 {
			return tag + "。伤官很旺，主才华横溢、思维敏捷、不服输。敢于表现，创业能力强，但易冲动。"
		}
		if pct >= 15 {
			return tag + "。伤官旺者，主才华横溢、思维敏捷、不服输。敢于表现，创业能力强，但易冲动。"
		}
		return tag + "。伤官适度，主才华出众、表达力强。在创意行业或技术领域有发展潜力。"
	case "正财":
		// 经典依据：《滴天髓》"正财端正，妻贤子孝"
		if pct >= 35 {
			return tag + "。正财极旺，主过度节俭、重利轻义、固守成规。须防吝啬、错失良机。"
		}
		if pct >= 25 {
			return tag + "。正财很旺，主务实节俭、理财稳重。财富通过正当努力积累，生活安稳。"
		}
		if pct >= 15 {
			return tag + "。正财旺者，主务实节俭、理财稳重。财富通过正当努力积累，生活安稳。"
		}
		return tag + "。正财适度，主踏实努力、财运稳定。适合稳定工作或固定投资。"
	case "偏财":
		// 经典依据：《三命通会》"偏财慷慨，不贪不吝"
		if pct >= 35 {
			return tag + "。偏财极旺，主挥金如土、投机心重、易有赌性。须防财务失控、桃色纠纷。"
		}
		if pct >= 25 {
			return tag + "。偏财很旺，主大方豪爽、社交广、投机心强。善于理财但需注意风险。"
		}
		if pct >= 15 {
			return tag + "。偏财旺者，主大方豪爽、社交广、投机心强。善于理财但需注意风险。"
		}
		return tag + "。偏财适度，主社交能力强、有偏财运。适合合伙或投资类事业。"
	case "正官":
		// 经典依据：《子平真诠》"正官清正，贵气天成"
		if pct >= 35 {
			return tag + "。正官极旺，主过于保守、循规蹈矩、压力过大。须防胆小怕事、健康受损。"
		}
		if pct >= 25 {
			return tag + "。正官很旺，主正直有责任感、守规矩、事业心强。仕途发展顺利，名声较好。"
		}
		if pct >= 15 {
			return tag + "。正官旺者，主正直有责任感、守规矩、事业心强。仕途发展顺利，名声较好。"
		}
		return tag + "。正官适度，主为人正派、事业稳定。在管理或体制内工作有发展。"
	case "七杀":
		// 经典依据：《滴天髓》"七杀有制化为权"
		if pct >= 35 {
			return tag + "。七杀极旺，主极端强势、刚愎自用、压力山大。须防意外灾祸、健康危机。"
		}
		if pct >= 25 {
			return tag + "。七杀很旺，主果断强势、魄力大、压力亦大。创业能力强，但需注意健康。"
		}
		if pct >= 15 {
			return tag + "。七杀旺者，主果断强势、魄力大、压力亦大。创业能力强，但需注意健康。"
		}
		return tag + "。七杀适度，主有魄力、敢于挑战。在竞争环境中能脱颖而出。"
	case "正印":
		// 经典依据：《子平真诠》"正印慈祥，聪明智慧"
		if pct >= 35 {
			return tag + "。正印极旺，主过度依赖、优柔寡断、空想不实。须防懒散、脱离实际。"
		}
		if pct >= 25 {
			return tag + "。正印很旺，主善良稳重、学识渊博、贵人运旺。学业或学术方面有天赋。"
		}
		if pct >= 15 {
			return tag + "。正印旺者，主善良稳重、学识渊博、贵人运旺。学业或学术方面有天赋。"
		}
		return tag + "。正印适度，主温和善良、有学识。适合教育或学术研究类工作。"
	case "偏印":
		// 经典依据：《滴天髓》"偏印精明，孤独特异"
		if pct >= 35 {
			return tag + "。偏印极旺，主孤僻多疑、思想偏激、精神紧张。须防心理问题、人际关系疏离。"
		}
		if pct >= 25 {
			return tag + "。偏印很旺，主悟性高、敏感细腻、悟道心强。学术或技术研究有独特见解。"
		}
		if pct >= 15 {
			return tag + "。偏印旺者，主悟性高、敏感细腻、悟道心强。学术或技术研究有独特见解。"
		}
		return tag + "。偏印适度，主悟性高、善于思考。在专业技术领域有发展潜力。"
	}
	return ""
}

// tenGodAdvice gives advice based on the ten god's proportion.
// gender is "MALE" or "FEMALE", used for gender-specific advice.
// 男命：正官为事业星，七杀为压力；正财为妻星，偏财为父星/情人
// 女命：正官为夫星，七杀为情人/偏夫；正财为财运
func tenGodAdvice(god string, pct float64, gender string) string {
	isMale := gender == "MALE"

	switch god {
	case "比肩":
		if pct >= 15 {
			return "扬长避短：发挥独立自信的优势，但需学会借力，避免过度固执而树敌。"
		}
		return "保持独立精神的同时注重团队合作，适当听取他人意见。"
	case "劫财":
		if pct >= 15 {
			return "注意控制冲动和冒险倾向，理财时需谨慎评估风险，避免合伙纠纷。"
		}
		return "发挥社交能力同时注重信用，财务往来要有凭据。"
	case "食神":
		if pct >= 15 {
			return "发挥创意和表达天赋，但需注意健康管理，避免过度享受导致健康问题。"
		}
		return "善用乐观开朗的性格，但需在重要事项上多加思考。"
	case "伤官":
		if pct >= 15 {
			return "才华需以合适方式展现，收敛冲动和叛逆情绪，创业需谨慎评估。"
		}
		return "发挥才华同时学会妥协，感情上需多沟通避免争执。"
	case "正财":
		if isMale {
			// 男命正财为妻星
			if pct >= 15 {
				return "正财旺者妻缘佳，财运稳定，宜稳扎稳打。珍惜伴侣缘分，理财宜稳健。"
			}
			return "继续保持务实节俭作风，用心经营婚姻感情与财务积累。"
		}
		// 女命正财为财运
		if pct >= 15 {
			return "财运稳定，宜稳扎稳打，但需开阔视野，适度参与投资以增值。"
		}
		return "继续保持务实节俭作风，可适当提升理财能力。"
	case "偏财":
		if isMale {
			// 男命偏财为父星/情人
			if pct >= 15 {
				return "偏财旺者社交广泛，异性缘佳，但投资理财需分散风险，避免投机。与父亲缘分较深。"
			}
			return "发挥社交能力的同时注意控制消费欲望，理性投资，珍惜父辈缘。"
		}
		// 女命偏财为偏财运
		if pct >= 15 {
			return "偏财运佳，投资理财需分散风险，不要过于自信或投机，合伙生意要明确权责。"
		}
		return "发挥社交能力的同时注意控制消费欲望，理性投资。"
	case "正官":
		if isMale {
			// 男命正官为事业星
			if pct >= 15 {
				return "正官旺者事业运佳，仕途发展顺利。但需注意不要过于追求权力而失去本心，健康也要关注。"
			}
			return "发挥正直品质的同时学会灵活变通，平衡事业与家庭。"
		}
		// 女命正官为夫星
		if pct >= 15 {
			return "正官旺者夫缘佳，婚姻稳定，另一半正直有责任感。珍惜姻缘，以柔克刚。"
		}
		return "正官适度，感情运势平稳，择偶宜重品行，保持温柔包容。"
	case "七杀":
		if isMale {
			// 男命七杀为压力/竞争
			if pct >= 15 {
				return "七杀旺者压力大，注意心血管健康，学会放松减压，做重大决策前多听听他人意见。"
			}
			return "发挥魄力的同时注意方法和节奏，避免过度竞争。"
		}
		// 女命七杀为情人/偏夫
		if pct >= 15 {
			return "七杀旺者异性缘复杂，感情中易遇强势对象。注意辨别真心，避免卷入多角关系。"
		}
		return "七杀适度，感情中保持独立判断，勿被表象迷惑。"
	case "正印":
		if pct >= 15 {
			return "学业事业均能得到贵人相助，但需主动出击，不要过于依赖他人。"
		}
		return "发挥学识优势的同时注重实践，保持善良但要有原则。"
	case "偏印":
		if pct >= 15 {
			return "悟性高但容易想太多而焦虑，建议培养兴趣爱好放松心神，避免精神过耗。"
		}
		return "发挥独特思考能力的同时多与他人交流，避免过于封闭。"
	}
	return ""
}

func (a *TenGodAnalyzer) buildPersonality(m map[string]TenGodRatio, dominant string, ptg PillarTenGods) string {
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

	// 柱位维度：年柱=祖辈根基与社会形象，月柱=事业与家庭影响，日柱=内在自我，时柱=未来追求
	pillarTraits := a.buildPillarPersonality(ptg)
	traits = append(traits, pillarTraits...)

		// Enrich with SanMingTongHui XingQing data for the dominant god
	if entry, ok := data.XingQingByTenGod[dominant]; ok {
			traits = append(traits, fmt.Sprintf("(%s) %s", entry.God, entry.Positive))
			traits = append(traits, fmt.Sprintf("(%s需注意) %s", entry.God, entry.Negative))
			traits = append(traits, fmt.Sprintf("(%s建议) %s", entry.God, entry.Advice))
	}
	return "你的性格主要表现为：" + strings.Join(traits, "，") + "。"
}

// godSet converts a slice of god names to a set for quick lookup.
func godSet(gods []string) map[string]bool {
	s := make(map[string]bool, len(gods))
	for _, g := range gods {
		s[g] = true
	}
	return s
}

// buildPillarPersonality adds personality traits based on pillar position of ten gods.
func (a *TenGodAnalyzer) buildPillarPersonality(ptg PillarTenGods) []string {
	var traits []string
	yearGods := godSet(ptg.Year)
	monthGods := godSet(ptg.Month)
	dayGods := godSet(ptg.Day)
	hourGods := godSet(ptg.Hour)

	// 年柱印星：祖辈根基深厚，早期教育良好
	if yearGods["正印"] || yearGods["偏印"] {
		traits = append(traits, "祖辈根基深厚，早期受良好教育熏陶")
	}
	// 年柱官杀：社会形象端正或有威严
	if yearGods["正官"] {
		traits = append(traits, "社会形象端正，自幼受规矩约束")
	} else if yearGods["七杀"] {
		traits = append(traits, "早年经历磨砺，锻炼出坚韧性格")
	}
	// 年柱比劫：与同辈竞争从小开始
	if yearGods["比肩"] || yearGods["劫财"] {
		traits = append(traits, "自幼独立意识强，与同辈竞争较多")
	}

	// 月柱财星：务实理财是核心性格
	if monthGods["正财"] || monthGods["偏财"] {
		traits = append(traits, "务实理财为核心性格特质")
	}
	// 月柱食伤：表达与创造是事业驱动力
	if monthGods["食神"] || monthGods["伤官"] {
		traits = append(traits, "表达与创造力是核心驱动力")
	}

	// 日柱比劫：自我意识强烈
	if dayGods["比肩"] || dayGods["劫财"] {
		traits = append(traits, "自我意识强烈，主见分明")
	}

	// 时柱食伤：追求精神满足与自我实现
	if hourGods["食神"] || hourGods["伤官"] {
		traits = append(traits, "追求精神满足与自我实现")
	}
	// 时柱印星：晚年趋向沉稳内省
	if hourGods["正印"] || hourGods["偏印"] {
		traits = append(traits, "晚年趋向沉稳内省")
	}

	return traits
}

func (a *TenGodAnalyzer) buildInterpersonal(m map[string]TenGodRatio, ptg PillarTenGods) string {
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

	// 柱位维度：年柱=长辈，月柱=同事/朋友，时柱=晚辈
	yearGods := godSet(ptg.Year)
	monthGods := godSet(ptg.Month)
	hourGods := godSet(ptg.Hour)

	if yearGods["正印"] || yearGods["偏印"] {
		relations = append(relations, "年柱印星，得长辈荫庇，与师长关系密切")
	}
	if yearGods["正官"] || yearGods["七杀"] {
		relations = append(relations, "年柱官杀，自幼受长辈严格管教，社会关系中有威望")
	}
	if monthGods["比肩"] || monthGods["劫财"] {
		relations = append(relations, "月柱比劫，同事朋友间竞争较多，需注意合作关系")
	}
	if monthGods["食神"] || monthGods["伤官"] {
		relations = append(relations, "月柱食伤，与同事朋友相处活跃，善于社交表达")
	}
	if hourGods["食神"] || hourGods["伤官"] {
		relations = append(relations, "时柱食伤，与子女或晚辈关系融洽，善于引导后辈")
	}
	if hourGods["正官"] || hourGods["七杀"] {
		relations = append(relations, "时柱官杀，对子女或晚辈要求严格，望子成龙心切")
	}

	return strings.Join(relations, "。") + "。"
}

func (a *TenGodAnalyzer) buildCareer(m map[string]TenGodRatio, ptg PillarTenGods) string {
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

	// 柱位维度：月柱=事业主线，时柱=成就收获
	monthGods := godSet(ptg.Month)
	hourGods := godSet(ptg.Hour)

	if monthGods["正官"] {
		career = append(career, "月柱正官，事业主线稳定，适合体制内或管理层发展")
	} else if monthGods["七杀"] {
		career = append(career, "月柱七杀，事业主线偏开拓型，适合创业或高压行业")
	}
	if monthGods["正财"] || monthGods["偏财"] {
		career = append(career, "月柱财星，事业以财务或商业为主线，财运与事业紧密相连")
	}
	if hourGods["食神"] || hourGods["伤官"] {
		career = append(career, "时柱食伤，晚年事业有成就，技艺或创意可转化为长期收益")
	}
	if hourGods["正财"] || hourGods["偏财"] {
		career = append(career, "时柱财星，晚年财运亨通，事业收获丰厚")
	}

	return strings.Join(career, "。") + "。"
}

func (a *TenGodAnalyzer) buildEmotion(m map[string]TenGodRatio, ptg PillarTenGods) string {
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

	// 柱位维度：日柱=配偶关系，时柱=子女关系
	dayGods := godSet(ptg.Day)
	hourGods := godSet(ptg.Hour)

	if dayGods["正官"] || dayGods["正财"] {
		emotion = append(emotion, "日柱正官/正财，配偶关系端正稳定，婚姻基础牢固")
	} else if dayGods["七杀"] || dayGods["偏财"] {
		emotion = append(emotion, "日柱七杀/偏财，配偶关系中有激情与挑战，需相互包容")
	}
	if dayGods["食神"] || dayGods["伤官"] {
		emotion = append(emotion, "日柱食伤，与配偶相处表达丰富，但也需注意言辞分寸")
	}
	if dayGods["正印"] || dayGods["偏印"] {
		emotion = append(emotion, "日柱印星，配偶关系中有包容与支持，精神契合度高")
	}

	if hourGods["正官"] || hourGods["七杀"] {
		emotion = append(emotion, "时柱官杀，子女关系中管教较严，望子成龙心切")
	}
	if hourGods["食神"] || hourGods["伤官"] {
		emotion = append(emotion, "时柱食伤，子女缘分较好，晚年情感生活丰富")
	}
	if hourGods["正印"] || hourGods["偏印"] {
		emotion = append(emotion, "时柱印星，子女孝顺有出息，晚年情感安稳")
	}

	return strings.Join(emotion, "。") + "。"
}

func (a *TenGodAnalyzer) buildHealth(m map[string]TenGodRatio, dayElem string) string {
	// 十神 → 五行映射（基于日主五行）
	godElemMap := map[string]string{
		"比肩": dayElem,
		"劫财": dayElem,
		"食神": woSheng(dayElem),
		"伤官": woSheng(dayElem),
		"正财": woKe(dayElem),
		"偏财": woKe(dayElem),
		"正官": keWo(dayElem),
		"七杀": keWo(dayElem),
		"正印": shengWo(dayElem),
		"偏印": shengWo(dayElem),
	}

	// 按五行聚合十神比例（比肩+劫财→日主五行，食神+伤官→我生五行，以此类推）
	elemPct := make(map[string]float64)
	for god, elem := range godElemMap {
		if ratio, ok := m[god]; ok {
			elemPct[elem] += ratio.Percent
		}
	}

	// 过旺分析：按严重程度分级
	// 15-25%: 轻度, 25-35%: 中度, >35%: 严重
	var excessWarnings []string
	for elem, pct := range elemPct {
		h, ok := data.WuxingHealthMap[elem]
		if !ok || pct < 15 {
			continue
		}
		organs := strings.Join(h.Organs, "、")
		if pct > 35 {
			excessWarnings = append(excessWarnings,
				fmt.Sprintf("%s严重过旺（%.1f%%），%s，注意%s", elem, pct, h.Excess, organs))
		} else if pct > 25 {
			excessWarnings = append(excessWarnings,
				fmt.Sprintf("%s中度过旺（%.1f%%），注意%s问题", elem, pct, organs))
		} else {
			excessWarnings = append(excessWarnings,
				fmt.Sprintf("%s轻度过旺（%.1f%%），留意%s", elem, pct, organs))
		}
	}

	// 过弱分析：<5% 为虚损
	var deficitWarnings []string
	for elem, pct := range elemPct {
		h, ok := data.WuxingHealthMap[elem]
		if !ok || pct >= 5 {
			continue
		}
		organs := strings.Join(h.Organs, "、")
		deficitWarnings = append(deficitWarnings,
			fmt.Sprintf("%s偏弱（%.1f%%），%s，对应%s", elem, pct, h.Deficit, organs))
	}

	// 十神性质相关的健康提示（不仅限于五行→脏腑映射）
	var godWarnings []string
	if m["比肩"].Percent >= 15 || m["劫财"].Percent >= 15 {
		godWarnings = append(godWarnings, "比劫旺者注意外伤、手术及意外伤害")
	}
	if m["食神"].Percent >= 15 {
		godWarnings = append(godWarnings, "食神旺者注意消化系统及代谢问题")
	}
	if m["伤官"].Percent >= 15 {
		godWarnings = append(godWarnings, "伤官旺者注意生殖系统及情绪波动引发的身体问题")
	}
	if m["正官"].Percent >= 15 || m["七杀"].Percent >= 15 {
		godWarnings = append(godWarnings, "官杀旺者注意压力性疾病、高血压及心血管问题")
	}
	if m["偏印"].Percent >= 15 {
		godWarnings = append(godWarnings, "偏印旺者注意心理健康、失眠及神经系统问题")
	}

	// 过弱的十神性质提示
	if m["比肩"].Percent < 5 && m["劫财"].Percent < 5 {
		godWarnings = append(godWarnings, "比劫偏弱，注意自身免疫力和体力不足")
	}
	if m["食神"].Percent < 5 && m["伤官"].Percent < 5 {
		godWarnings = append(godWarnings, "食伤偏弱，注意消化吸收和情志疏导")
	}
	if m["正印"].Percent < 5 && m["偏印"].Percent < 5 {
		godWarnings = append(godWarnings, "印星偏弱，注意休息调养，防思虑过度伤身")
	}

	if len(excessWarnings) == 0 && len(deficitWarnings) == 0 && len(godWarnings) == 0 {
		return "健康运势总体平稳，无明显倾向，注意日常保养即可。"
	}

	var parts []string
	if len(excessWarnings) > 0 {
		unique := removeDuplicates(excessWarnings)
		parts = append(parts, "【过旺预警】"+strings.Join(unique, "；"))
	}
	if len(deficitWarnings) > 0 {
		unique := removeDuplicates(deficitWarnings)
		parts = append(parts, "【虚损预警】"+strings.Join(unique, "；"))
	}
	if len(godWarnings) > 0 {
		parts = append(parts, godWarnings...)
	}
	return strings.Join(parts, "。") + "。建议定期体检，保持良好生活习惯。"
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
		taboos = append(taboos, "避免贪心和投机型投资", "合伙生意要明确权责", "财务决策要冷静", "避免挥霍无度、沉迷投机赌博", "桃色纠纷需警惕，勿因财色伤身败名")
	case "比肩":
		taboos = append(taboos, "避免过度固执不听劝", "要学会借力和合作", "竞争中勿伤和气")
	case "正官":
		taboos = append(taboos, "避免过于追求权力", "不要过于保守而错失机会", "健康与家庭需关注")
	case "食神":
		taboos = append(taboos, "避免过度享受和暴饮暴食", "健康管理不容忽视", "工作娱乐要平衡")
	case "正财":
		taboos = append(taboos, "避免过于保守而不敢投资", "要开阔视野，学习新知", "理财要多元化", "避免重利轻义、因财伤情", "勿因节俭而吝啬，影响人际关系")
	case "正印":
		taboos = append(taboos, "避免过于依赖他人", "要主动出击争取机会", "实践中检验学识")
	}

	return strings.Join(taboos, "。") + "。"
}

func (a *TenGodAnalyzer) buildSummary(m map[string]TenGodRatio, dominant string, dominantPct float64) string {
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

	// Check for special combinations
	// 经典依据：《子平真诠》《滴天髓》论十神组合
	var combo []string

	if jiecai >= 20 && zhengguan >= 15 && zhengyin >= 15 {
		combo = append(combo, "劫财+正官+正印 组合，各方面运势相对平衡")
	}

	if bijiao+jiecai >= 40 {
		combo = append(combo, "比劫过旺，竞争意识强，需学会合作")
	}

	// 经典依据：《子平真诠》"官杀混杂须制化"
	if zhengguan >= 15 && qisha >= 15 {
		combo = append(combo, "官杀混杂，压力大，须以食伤制杀或印化")
	}

	if zhengyin >= 15 && pianyin >= 15 {
		combo = append(combo, "正偏印同现，学识与悟性兼备，适合学术研究")
	}

	// 经典依据：《滴天髓》"食神制杀，英雄独压万人"
	if shishen >= 15 && qisha >= 15 {
		combo = append(combo, "食神制杀，刚柔并济，事业有成之象")
	}

	// 经典依据：《子平真诠》"伤官见官为祸百端"
	if shangguan >= 15 && zhengguan >= 15 {
		combo = append(combo, "伤官见官，易有是非口舌，须以印化解")
	}

	// 经典依据：《三命通会》"财官双美格"
	if zhengcai >= 15 && zhengguan >= 15 {
		combo = append(combo, "财官双美，名利双收之象")
	}

	// 经典依据：《滴天髓》"枭神夺食"
	if pianyin >= 15 && shishen >= 15 {
		combo = append(combo, "枭神夺食，才华受制，须防挫折")
	}

	// 经典依据：《三命通会》"偏财透干，慷慨豪爽"
	if piancai >= 20 && (shishen >= 15 || shangguan >= 15) {
		combo = append(combo, "食伤生偏财，财源广进，投资理财能力强")
	}

	// 经典依据：《子平真诠》"食伤生财，秀气流通"
	if zhengcai >= 15 && (shishen >= 15 || shangguan >= 15) {
		combo = append(combo, "食伤生财，才华转化为财富，收入与能力成正比")
	}

	// 经典依据：《滴天髓》"印化杀，以柔克刚"
	if qisha >= 15 && (zhengyin >= 15 || pianyin >= 15) {
		combo = append(combo, "印化杀，以学识化解压力，遇难呈祥之象")
	}

	// 经典依据：《子平真诠》"财滋弱杀"
	if (zhengcai >= 15 || piancai >= 15) && qisha >= 15 && qisha < 20 {
		combo = append(combo, "财滋杀，财星助长七杀之力，须防压力骤增")
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