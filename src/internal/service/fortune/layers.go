package fortune

import (
	"strings"
	"time"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/data"
)

// PeriodLayerCalculator calculates one fortune-period layer.
type PeriodLayerCalculator interface {
	Key() string
	Calculate(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer
}

type daYunLayerCalculator struct{}
type liuNianLayerCalculator struct{}
type liuYueLayerCalculator struct{}
type xiaoYunLayerCalculator struct{}

func (daYunLayerCalculator) Key() string   { return "dayun" }
func (liuNianLayerCalculator) Key() string { return "liunian" }
func (liuYueLayerCalculator) Key() string  { return "liuyue" }
func (xiaoYunLayerCalculator) Key() string { return "xiaoyun" }

func (daYunLayerCalculator) Calculate(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	return buildDaYunLayer(bazi, queryDate, birthYear)
}

func (liuNianLayerCalculator) Calculate(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	return buildLiuNianLayer(bazi, queryDate)
}

func (liuYueLayerCalculator) Calculate(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	return buildLiuYueLayer(bazi, queryDate, birthYear)
}

func (xiaoYunLayerCalculator) Calculate(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	return buildXiaoYunLayer(bazi, queryDate, birthYear)
}

var defaultPeriodLayerCalculators = []PeriodLayerCalculator{
	daYunLayerCalculator{},
	liuNianLayerCalculator{},
	liuYueLayerCalculator{},
	xiaoYunLayerCalculator{},
}

// BuildFortuneLayers returns 大运、流年、流月、小运 as explicit layers.
func BuildFortuneLayers(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayerSet {
	return BuildFortuneLayersWithCalculators(bazi, queryDate, birthYear, defaultPeriodLayerCalculators)
}

func BuildFortuneLayersWithCalculators(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int, calculators []PeriodLayerCalculator) model.FortuneLayerSet {
	layers := model.FortuneLayerSet{
		RuleVersion:         bazipkg.RuleVersion,
		School:              bazipkg.RuleSchool,
		InterLayerRelations: []model.FortuneLayerRelation{},
	}
	for _, calculator := range calculators {
		layer := calculator.Calculate(bazi, queryDate, birthYear)
		switch calculator.Key() {
		case "dayun":
			layers.DaYun = layer
		case "liunian":
			layers.LiuNian = layer
		case "liuyue":
			layers.LiuYue = layer
		case "xiaoyun":
			layers.XiaoYun = layer
		}
	}
	layers.InterLayerRelations = coreInterLayerRelations(layers)
	return layers
}

// coreInterLayerRelations records the deterministic 大运-流年-流月 chain.
// It does not assign favorable/unfavorable meaning or change the score.
func coreInterLayerRelations(layers model.FortuneLayerSet) []model.FortuneLayerRelation {
	relations := make([]model.FortuneLayerRelation, 0, 6)
	for _, pair := range []struct {
		source model.FortuneLayer
		target model.FortuneLayer
	}{
		{source: layers.LiuNian, target: layers.DaYun},
		{source: layers.LiuYue, target: layers.LiuNian},
		{source: layers.LiuYue, target: layers.DaYun},
	} {
		if pair.source.Status != "observed" || pair.target.Status != "observed" {
			continue
		}
		stemRelations := layerStemRelations(
			pair.source.Name+"天干", pair.source.Gan,
			pair.target.Name+"天干", pair.target.Gan,
		)
		for i := range stemRelations {
			stemRelations[i].Basis = "period_layer_stem_pair"
		}
		relations = append(relations, stemRelations...)
		branchRelations := layerBranchRelations(
			pair.source.Name+"地支", pair.source.Zhi,
			pair.target.Name+"地支", pair.target.Zhi,
		)
		for i := range branchRelations {
			branchRelations[i].Basis = "period_layer_branch_pair"
		}
		relations = append(relations, branchRelations...)
	}
	return relations
}

func buildDaYunLayer(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	influence := calcDaYunInfluence(bazi, queryDate, birthYear)
	layer := baseLayer("dayun", "大运", influence.CurrentPillar, bazi, queryDate)
	layer.StartAge = influence.StartAge
	layer.EndAge = influence.EndAge
	layer.StartAt = influence.StartAt
	layer.EndAtExclusive = influence.EndAtExclusive
	layer.Age = queryDate.Year() - birthYear
	layer.Basis = influence.SelectionBasis
	layer.Status = influence.Status
	return layer
}

func buildLiuNianLayer(bazi *bazipkg.BaziResult, queryDate time.Time) model.FortuneLayer {
	pillar := getYearGanZhi(queryDate.Year(), int(queryDate.Month()), queryDate.Day())
	layer := baseLayer("liunian", "流年", pillar, bazi, queryDate)
	influence := calcLiuNianInfluence(bazi, queryDate)
	layer.Year = queryDate.Year()
	layer.Basis = influence.SelectionBasis
	layer.Status = influence.Status
	return layer
}

func buildLiuYueLayer(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	pillar := ""
	if ec, err := getDayEightChar(queryDate.Year(), int(queryDate.Month()), queryDate.Day()); err == nil {
		pillar = ec.GetMonth().GetName()
	}
	layer := baseLayer("liuyue", "流月", pillar, bazi, queryDate)
	layer.Year = queryDate.Year()
	layer.Month = int(queryDate.Month())
	layer.Age = queryDate.Year() - birthYear
	layer.Basis = "query_date_month_pillar_at_solar_term_boundary"
	return layer
}

func buildXiaoYunLayer(bazi *bazipkg.BaziResult, queryDate time.Time, birthYear int) model.FortuneLayer {
	nominalAge := queryDate.Year() - birthYear + 1
	offset := nominalAge
	if strings.Contains(bazi.DaYunInfo.Direction, "逆") {
		offset = -nominalAge
	}
	pillar := offsetPillar(bazi.HourPillar, offset)
	layer := baseLayer("xiaoyun", "小运", pillar, bazi, queryDate)
	layer.RuleID = "fortune.layer.xiaoyun-v3"
	layer.Age = nominalAge
	layer.Year = queryDate.Year()
	layer.Basis = "tyme4go_fortune_hour_pillar_direction_and_nominal_age"
	return layer
}

func baseLayer(key, name, pillar string, bazi *bazipkg.BaziResult, queryDate time.Time) model.FortuneLayer {
	gan, zhi := splitPillar(pillar)
	layer := model.FortuneLayer{
		RuleID:               "fortune.layer." + key + "-v2",
		Key:                  key,
		Name:                 name,
		Pillar:               pillar,
		Gan:                  gan,
		Zhi:                  zhi,
		Relations:            layerRelations(gan, zhi, bazi, queryDate),
		ShenShaDetails:       layerShenShaDetails(gan, zhi, bazi),
		Basis:                "period_pillar_and_natal_chart",
		Status:               "unavailable",
		InterpretationStatus: "not_adjudicated",
	}
	if gan != "" {
		layer.TenGod = observeTenGod(bazi.DayPillar.Gan, gan)
	}
	if gan != "" && zhi != "" {
		layer.Status = "observed"
	}
	return layer
}

func splitPillar(pillar string) (string, string) {
	runes := []rune(strings.TrimSpace(pillar))
	if len(runes) < 2 {
		return "", ""
	}
	return string(runes[0]), string(runes[1])
}

func layerRelations(gan, zhi string, bazi *bazipkg.BaziResult, queryDate time.Time) []model.FortuneLayerRelation {
	var rels []model.FortuneLayerRelation
	if gan != "" {
		rels = append(rels, layerStemRelations("周期天干", gan, "日干", bazi.DayPillar.Gan)...)
		if todayGan := queryDateGan(queryDate); todayGan != "" {
			rels = append(rels, layerStemRelations("周期天干", gan, "查询日干", todayGan)...)
		}
	}
	if zhi != "" {
		targets := []struct {
			name string
			zhi  string
		}{
			{"年支", bazi.YearPillar.Zhi},
			{"月支", bazi.MonthPillar.Zhi},
			{"日支", bazi.DayPillar.Zhi},
			{"时支", bazi.HourPillar.Zhi},
		}
		if today, err := getDayPillar(queryDate.Year(), int(queryDate.Month()), queryDate.Day()); err == nil {
			targets = append([]struct {
				name string
				zhi  string
			}{{"查询日支", today.Zhi}}, targets...)
		}
		for _, target := range targets {
			rels = append(rels, layerBranchRelations("周期地支", zhi, target.name, target.zhi)...)
		}
	}
	return rels
}

func layerStemRelations(source, sourceStem, target, targetStem string) []model.FortuneLayerRelation {
	if data.GanIndex(sourceStem) < 0 || data.GanIndex(targetStem) < 0 {
		return []model.FortuneLayerRelation{}
	}
	types := make([]string, 0, 2)
	if isGanHe(sourceStem, targetStem) {
		types = append(types, "five_combine")
	}
	if stemClashMap[sourceStem] == targetStem {
		types = append(types, "clash")
	}
	if elementRelation := stemRelation(targetStem, sourceStem); elementRelation != "unknown" {
		types = append(types, elementRelation)
	}
	relations := make([]model.FortuneLayerRelation, 0, len(types))
	for _, relationType := range types {
		relationName := stemRelLabel(relationType, targetStem, sourceStem)
		switch relationType {
		case "five_combine":
			relationName = "天干五合"
		case "clash":
			relationName = "天干相冲"
		}
		relations = append(relations, model.FortuneLayerRelation{
			RuleID:               "fortune.layer-relation.stem-v3." + relationType,
			Source:               source,
			SourceValue:          sourceStem,
			Target:               target,
			TargetValue:          targetStem,
			Type:                 relationType,
			Name:                 relationName,
			Basis:                "period_stem_and_target_stem_all_structures",
			Status:               "observed",
			InterpretationStatus: "not_adjudicated",
		})
	}
	return relations
}

func layerBranchRelations(source, sourceBranch, target, targetBranch string) []model.FortuneLayerRelation {
	if data.ZhiIndex(sourceBranch) < 0 || data.ZhiIndex(targetBranch) < 0 {
		return []model.FortuneLayerRelation{}
	}
	types := make([]string, 0, 4)
	if sourceBranch == targetBranch {
		types = append(types, "same")
		if strings.Contains("辰午酉亥", sourceBranch) {
			types = append(types, "punish")
		}
	} else {
		if clashPairs[sourceBranch] == targetBranch {
			types = append(types, "clash")
		}
		if punishPairs[sourceBranch+targetBranch] {
			types = append(types, "punish")
		}
		if harmPairs[sourceBranch] == targetBranch {
			types = append(types, "harm")
		}
		if breakPairs[sourceBranch] == targetBranch {
			types = append(types, "break")
		}
		if combinePairs[sourceBranch] == targetBranch {
			types = append(types, "combine")
		}
		if partial := partialSanHeRelation(sourceBranch, targetBranch); partial != "" {
			types = append(types, partial)
		}
		if isInSameGroup(sanHuiGroups, sourceBranch, targetBranch) {
			types = append(types, "banHui")
		}
	}
	relations := make([]model.FortuneLayerRelation, 0, len(types))
	for _, relationType := range types {
		relationName := branchRelLabel(relationType)
		if relationType == "same" {
			relationName = "同支"
		}
		relations = append(relations, model.FortuneLayerRelation{
			RuleID:               "fortune.layer-relation.branch-v3." + relationType,
			Source:               source,
			SourceValue:          sourceBranch,
			Target:               target,
			TargetValue:          targetBranch,
			Type:                 relationType,
			Name:                 relationName,
			Basis:                "period_branch_and_target_branch_all_structures",
			Status:               "observed",
			InterpretationStatus: "not_adjudicated",
		})
	}
	return relations
}

func layerShenShaDetails(gan, zhi string, bazi *bazipkg.BaziResult) []model.ShenShaActivation {
	if gan == "" || zhi == "" {
		return nil
	}
	return calcShenShaActivation(gan, zhi, bazi)
}

func offsetPillar(p model.Pillar, offset int) string {
	start := sixtyIndex(p)
	if start < 0 {
		return p.Gan + p.Zhi
	}
	idx := (start + offset) % 60
	if idx < 0 {
		idx += 60
	}
	return data.Gans[idx%10] + data.Zhis[idx%12]
}

func sixtyIndex(p model.Pillar) int {
	for i := 0; i < 60; i++ {
		if data.Gans[i%10] == p.Gan && data.Zhis[i%12] == p.Zhi {
			return i
		}
	}
	return -1
}
