package fortune

import (
	"fmt"
	"strings"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
)

const (
	// FortuneEngineVersion identifies the public fortune interpretation engine.
	FortuneEngineVersion = "fortune-engine-2026-07-10"
	// FortuneScorePipelineVersion identifies the scoring stages and weights.
	FortuneScorePipelineVersion = "fortune-score-pipeline-2026-07-10"
)

var stemScoreRules = map[string]struct {
	impact int
	label  string
	detail string
}{
	"same":    {10, "天干比和", "流日天干与日主同类，比劫气势增强。"},
	"shengWo": {18, "流日生扶日主", "流日五行生扶日主，印星助力较明显。"},
	"woSheng": {5, "日主生流日", "日主生出流日，利于表达与输出，但也会消耗精力。"},
	"keWo":    {-18, "流日克制日主", "流日五行克制日主，责任、约束与压力增加。"},
	"woKe":    {8, "日主克流日", "日主能够驾驭流日之气，财星与资源议题更突出。"},
}

var branchScoreRules = map[string]struct {
	impact int
	label  string
	detail string
}{
	"clash":   {-30, "地支六冲", "流日支与日支相冲，变动和对立信号较强。"},
	"harm":    {-15, "地支六害", "流日支与日支相害，需留意隐性摩擦。"},
	"punish":  {-20, "地支相刑", "流日支与日支相刑，规则压力与反复增加。"},
	"break":   {-10, "地支相破", "流日支与日支相破，计划稳定性下降。"},
	"combine": {8, "地支六合", "流日支与日支六合，协同与承接条件增加。"},
	"banHe":   {6, "地支半合", "两支构成含中神的半合，存在相合趋势但尚未成完整三合局。"},
	"gongHe":  {4, "地支拱合", "三合首尾相见而缺中神，仅作拱合趋势处理。"},
	"banHui":  {5, "地支半会", "两支方气相连，但尚未成完整三会局。"},
	"sanHe":   {15, "地支三合", "三支齐备形成完整三合局。"},
	"sanHui":  {20, "地支三会", "三支齐备形成完整三会局。"},
}

func relationScoreStage(stemRel, branchRel, userGan, dayGan string) (int, []model.ScoreEvidence) {
	score := 50
	evidence := make([]model.ScoreEvidence, 0, 3)

	if rule, ok := stemScoreRules[stemRel]; ok {
		score += rule.impact
		evidence = append(evidence, model.ScoreEvidence{
			Code:        "relation.stem." + stemRel,
			Stage:       "relation",
			Category:    "天干生克",
			Label:       rule.label,
			Impact:      rule.impact,
			Description: rule.detail,
			Source:      "《三命通会》十神生克规则",
		})
	}

	if rule, ok := branchScoreRules[branchRel]; ok {
		score += rule.impact
		evidence = append(evidence, model.ScoreEvidence{
			Code:        "relation.branch." + branchRel,
			Stage:       "relation",
			Category:    "地支关系",
			Label:       rule.label,
			Impact:      rule.impact,
			Description: rule.detail,
			Source:      "《三命通会》《协纪辨方书》地支关系规则",
		})
	}

	if userGan != "" && dayGan != "" && isGanHe(userGan, dayGan) {
		score += 12
		evidence = append(evidence, model.ScoreEvidence{
			Code:        "relation.stem.five_combine",
			Stage:       "relation",
			Category:    "天干关系",
			Label:       "天干五合",
			Impact:      12,
			Description: fmt.Sprintf("日主%s与流日天干%s形成五合，气机交融。", userGan, dayGan),
			Source:      "《三命通会》天干五合规则",
		})
	}

	return clampFortuneScore(score), evidence
}

func buildScorePipeline(
	chart *bazipkg.BaziResult,
	stemRel string,
	branchRel string,
	userGan string,
	dayGan string,
	analysis *FortuneAnalysis,
	rikuyo *RikuyoResult,
) model.FortuneScoreBreakdown {
	relationScore, relationEvidence := relationScoreStage(stemRel, branchRel, userGan, dayGan)
	detailScore := relationScore
	finalScore := relationScore

	supporting := make([]model.ScoreEvidence, 0, len(relationEvidence)+9)
	counter := make([]model.ScoreEvidence, 0, len(relationEvidence)+9)
	appendByPolarity := func(item model.ScoreEvidence) {
		if item.Impact > 0 {
			supporting = append(supporting, item)
		} else if item.Impact < 0 {
			counter = append(counter, item)
		}
	}
	for _, item := range relationEvidence {
		appendByPolarity(item)
	}

	if analysis != nil {
		detailScore = analysis.Overall.DetailScore
		finalScore = blendScores(relationScore, detailScore)
		for _, category := range analysis.Categories {
			impact := weightedCategoryImpact(category.Score, category.Weight)
			if impact == 0 {
				continue
			}
			item := model.ScoreEvidence{
				Code:        "detail.category." + categoryEvidenceCode(category.Name),
				Stage:       "detail",
				Category:    category.Name,
				Label:       category.Name + "分项",
				Impact:      impact,
				Description: fmt.Sprintf("%s分项为%d分，权重%d；%s", category.Name, category.Score, category.Weight, strings.TrimSpace(category.Analysis)),
				Source:      "九维运势细项评分规则",
			}
			appendByPolarity(item)
		}
		analysis.Overall.BaseScore = relationScore
		analysis.Overall.Score = finalScore
		analysis.Overall.Stars = scoreToStars(finalScore)
		analysis.Overall.Level = scoreLevel(finalScore)
	}

	return model.FortuneScoreBreakdown{
		PipelineVersion:      FortuneScorePipelineVersion,
		BaseScore:            50,
		RelationScore:        relationScore,
		DetailScore:          detailScore,
		FinalScore:           finalScore,
		EvidenceCompleteness: scoreEvidenceCompleteness(chart, stemRel, branchRel, analysis, rikuyo),
		SupportingEvidence:   supporting,
		CounterEvidence:      counter,
	}
}

func weightedCategoryImpact(score, weight int) int {
	if score == 60 || weight <= 0 {
		return 0
	}
	impact := (score - 60) * weight / 100
	if impact == 0 {
		if score > 60 {
			return 1
		}
		return -1
	}
	return impact
}

func scoreEvidenceCompleteness(chart *bazipkg.BaziResult, stemRel, branchRel string, analysis *FortuneAnalysis, rikuyo *RikuyoResult) int {
	completeness := 0
	if chart != nil && chart.DayPillar.Gan != "" && chart.DayPillar.Zhi != "" {
		completeness += 20
	}
	if chart != nil && chart.YearPillar.Gan != "" && chart.MonthPillar.Gan != "" && chart.HourPillar.Gan != "" {
		completeness += 15
	}
	if chart != nil && strings.TrimSpace(chart.BodyStrength.Verdict) != "" {
		completeness += 15
	}
	if chart != nil {
		like, dislike, _ := getEffectiveFavor(chart)
		if len(like)+len(dislike) > 0 {
			completeness += 10
		}
	}
	if stemRel != "" && stemRel != "unknown" {
		completeness += 10
	}
	if branchRel != "" && branchRel != "unknown" {
		completeness += 10
	}
	if analysis != nil && len(analysis.Categories) == len(dailyCategoryWeights) {
		completeness += 15
	}
	if rikuyo != nil {
		completeness += 5
	}
	if completeness > 100 {
		return 100
	}
	return completeness
}

func categoryEvidenceCode(name string) string {
	replacer := strings.NewReplacer("事业", "career", "财运", "wealth", "感情", "love", "健康", "health", "贵人", "noble", "学业", "study", "投资", "invest", "出行", "travel", "官非", "lawsuit", " ", "_")
	code := replacer.Replace(strings.TrimSpace(name))
	if code == "" {
		return "unknown"
	}
	return code
}

func clampFortuneScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
