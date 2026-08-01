package fortune

import (
	"fmt"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
)

const (
	// FortuneEngineVersion identifies the public fortune interpretation engine.
	FortuneEngineVersion = "fortune-engine-2026-07-19.17"
	// FortuneScorePipelineVersion identifies the scoring stages and weights.
	FortuneScorePipelineVersion = "fortune-score-pipeline-2026-07-15.3"
)

var stemScoreRules = map[string]struct {
	impact int
	label  string
	detail string
}{
	"same":    {10, "天干比和", "流日天干与日主天干构成同五行关系。"},
	"shengWo": {18, "流日生扶", "今日天干生助日主。"},
	"woSheng": {5, "日主生流日", "日主生助今日天干。"},
	"keWo":    {-18, "流日克日主", "今日天干克制日主。"},
	"woKe":    {8, "日主克流日", "日主克制今日天干。"},
}

var branchScoreRules = map[string]struct {
	impact int
	label  string
	detail string
}{
	"clash":   {-30, "地支六冲", "流日支与日支命中六冲结构。"},
	"harm":    {-15, "地支六害", "流日支与日支命中六害结构。"},
	"punish":  {-20, "地支相刑", "流日支与日支命中相刑结构。"},
	"break":   {-10, "地支相破", "流日支与日支命中相破结构。"},
	"combine": {8, "地支六合", "流日支与日支命中六合结构。"},
	"banHe":   {6, "地支半合", "两支构成含中神的半合结构，未形成完整三合局。"},
	"gongHe":  {4, "地支拱合", "三合首尾相见而缺中神，仅记录拱合结构。"},
	"banHui":  {5, "地支半会", "两支方气相连，未形成完整三会局。"},
	"sanHe":   {15, "地支三合", "三支齐备，命中完整三合结构。"},
	"sanHui":  {20, "地支三会", "三支齐备，命中完整三会结构。"},
}

func newScoreEvidence(code, category, label, description, source string, impact int) model.ScoreEvidence {
	return model.ScoreEvidence{
		Code:                 code,
		Stage:                "relation",
		Category:             category,
		Label:                label,
		Impact:               impact,
		Description:          description,
		Source:               source,
		EvidenceBasis:        "empirical",
		ValidationStatus:     "not_validated",
		InterpretationStatus: "not_adjudicated",
		IsOutcomeConclusion:  false,
	}
}

func relationScoreStage(stemRel, branchRel, userGan, dayGan string) (int, []model.ScoreEvidence) {
	score := 50
	evidence := make([]model.ScoreEvidence, 0, 3)

	if rule, ok := stemScoreRules[stemRel]; ok {
		score += rule.impact
		evidence = append(evidence, newScoreEvidence(
			"relation.stem."+stemRel,
			"天干生克",
			rule.label,
			rule.detail,
			"本地启发式权重；结构取法参考《三命通会》十神生克规则",
			rule.impact,
		))
	}

	if rule, ok := branchScoreRules[branchRel]; ok {
		score += rule.impact
		evidence = append(evidence, newScoreEvidence(
			"relation.branch."+branchRel,
			"地支关系",
			rule.label,
			rule.detail,
			"本地启发式权重；结构取法参考《三命通会》《协纪辨方书》地支关系规则",
			rule.impact,
		))
	}

	if userGan != "" && dayGan != "" && isGanHe(userGan, dayGan) {
		score += 12
		evidence = append(evidence, newScoreEvidence(
			"relation.stem.five_combine",
			"天干关系",
			"天干五合",
			fmt.Sprintf("日主%s与今日天干%s命中五合结构；这里只记录五合，是否成化不在当前规则判断范围内。", userGan, dayGan),
			"本地启发式权重；结构取法参考《三命通会》天干五合规则",
			12,
		))
	}

	return clampFortuneScore(score), evidence
}

func buildScorePipeline(
	chart *bazipkg.BaziResult,
	stemRel string,
	branchRel string,
	userGan string,
	dayGan string,
) model.FortuneScoreBreakdown {
	relationScore, relationEvidence := relationScoreStage(stemRel, branchRel, userGan, dayGan)

	supporting := make([]model.ScoreEvidence, 0, len(relationEvidence))
	counter := make([]model.ScoreEvidence, 0, len(relationEvidence))
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

	return model.FortuneScoreBreakdown{
		PipelineVersion:      FortuneScorePipelineVersion,
		ScoreKind:            "structural_relation_index",
		EvidenceBasis:        "empirical",
		ValidationStatus:     "not_validated",
		InterpretationStatus: "not_adjudicated",
		IsOutcomeProbability: false,
		BaseScore:            50,
		RelationScore:        relationScore,
		FinalScore:           relationScore,
		EvidenceCompleteness: scoreEvidenceCompleteness(chart, stemRel, branchRel, userGan, dayGan),
		SupportingEvidence:   supporting,
		CounterEvidence:      counter,
	}
}

func scoreEvidenceCompleteness(chart *bazipkg.BaziResult, stemRel, branchRel, userGan, dayGan string) int {
	completeness := 0
	if chart != nil && chart.DayPillar.Gan != "" && chart.DayPillar.Zhi != "" {
		completeness += 40
	}
	if stemRel != "" && stemRel != "unknown" {
		completeness += 25
	}
	if branchRel != "" && branchRel != "unknown" {
		completeness += 25
	}
	if userGan != "" && dayGan != "" {
		completeness += 10
	}
	if completeness > 100 {
		return 100
	}
	return completeness
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
