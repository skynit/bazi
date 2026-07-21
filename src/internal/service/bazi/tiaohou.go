package bazi

import (
	"fmt"
	"reflect"
	"sort"

	"bazi/internal/model"
	"bazi/internal/service/data"
	"github.com/6tail/tyme4go/tyme"
)

const (
	tiaohouRuleID              = "bazi.tiaohou.table-candidates-v3"
	tiaohouJiShenUnadjudicated = "未裁决"
)

// TiaohouResult records table candidates and solar-term depth evidence without
// adjudicating a unique useful god or generating real-world advice.
type TiaohouResult struct {
	RuleID                string                `json:"rule_id"`
	Stem                  string                `json:"stem"`
	Month                 string                `json:"month"`
	Rules                 []TiaohouRuleEvidence `json:"rules"`
	TablePrimaryCandidate string                `json:"table_primary_candidate"`
	SelectionBasis        string                `json:"selection_basis"`
	DepthAffectsSelection bool                  `json:"depth_affects_selection"`
	DepthEvidence         TiaohouDepthEvidence  `json:"depth_evidence"`
	ChartEvidence         TiaohouChartEvidence  `json:"chart_evidence"`
	MatchedConditions     []TiaohouConditionHit `json:"matched_conditions"`
	ChartCandidates       []string              `json:"chart_candidates"`
	ChartSelectionBasis   string                `json:"chart_selection_basis"`
	Status                string                `json:"status"`
	ValidationStatus      string                `json:"validation_status"`
	InterpretationStatus  string                `json:"interpretation_status"`
	Limitations           []string              `json:"limitations"`
}

// TiaohouChartEvidence contains only deterministic four-pillar facts used by
// the small reviewed condition set below. A complete branch structure is not
// treated as proof of transformation.
type TiaohouChartEvidence struct {
	Status                   string                   `json:"status"`
	Basis                    string                   `json:"basis"`
	VisibleStems             []string                 `json:"visible_stems"`
	Branches                 []string                 `json:"branches"`
	CompleteBranchStructures []TiaohouBranchStructure `json:"complete_branch_structures"`
}

type TiaohouBranchStructure struct {
	RuleID               string   `json:"rule_id"`
	Type                 string   `json:"type"`
	Branches             []string `json:"branches"`
	TargetElement        string   `json:"target_element"`
	TransformationStatus string   `json:"transformation_status"`
}

// TiaohouConditionHit records a reviewed source condition that actually
// matches this chart. Candidate order follows the cited source; it is not a
// unique useful-god adjudication.
type TiaohouConditionHit struct {
	RuleID               string   `json:"rule_id"`
	Candidates           []string `json:"candidates"`
	Condition            string   `json:"condition"`
	Evidence             []string `json:"evidence"`
	Source               string   `json:"source"`
	SourceText           string   `json:"source_text"`
	Status               string   `json:"status"`
	ValidationStatus     string   `json:"validation_status"`
	InterpretationStatus string   `json:"interpretation_status"`
}

type tiaohouChartRule struct {
	ruleID               string
	dayStem              string
	monthBranch          string
	requiredStructureID  string
	requiredVisibleStems []string
	candidates           []string
	condition            string
	source               string
	sourceText           string
}

// This intentionally small registry contains only conditions whose chart
// inputs are explicit and already supported by deterministic relation facts.
// Ambiguous phrases such as “水多” and “土重” stay unadjudicated.
var tiaohouChartRules = []tiaohouChartRule{
	{
		ruleID:  "bazi.tiaohou.chart.yi-you-metal-frame",
		dayStem: "乙", monthBranch: "酉",
		requiredStructureID: "branch.sanhe.metal",
		candidates:          []string{"丁"},
		condition:           "四支具备巳酉丑三合金局完整结构",
		source:              "《穷通宝鉴·八月乙木》",
		sourceText:          "或支成金局，宜暗藏丁制。",
	},
	{
		ruleID:  "bazi.tiaohou.chart.bing-si-water-frame",
		dayStem: "丙", monthBranch: "巳",
		requiredStructureID:  "branch.sanhe.water",
		requiredVisibleStems: []string{"壬"},
		candidates:           []string{"戊"},
		condition:            "四支具备申子辰三合水局完整结构，且壬水透干",
		source:               "《穷通宝鉴·四月丙火》",
		sourceText:           "或支成水局，又一二壬透，无一戊制。",
	},
	{
		ruleID:  "bazi.tiaohou.chart.ding-chen-water-frame-ren-visible",
		dayStem: "丁", monthBranch: "辰",
		requiredStructureID:  "branch.sanhe.water",
		requiredVisibleStems: []string{"壬"},
		candidates:           []string{"戊", "己"},
		condition:            "四支具备申子辰三合水局完整结构，且壬水透干",
		source:               "《穷通宝鉴·三月丁火》",
		sourceText:           "或支成水局，加以壬透……或得戊己两透。",
	},
}

type TiaohouRuleEvidence struct {
	RuleID               string `json:"rule_id"`
	XiShen               string `json:"xi_shen"`
	JiShen               string `json:"ji_shen"`
	JiShenStatus         string `json:"ji_shen_status"`
	SourceText           string `json:"source_text"`
	Basis                string `json:"basis"`
	Status               string `json:"status"`
	ValidationStatus     string `json:"validation_status"`
	InterpretationStatus string `json:"interpretation_status"`
}

// TiaohouDepthEvidence records where the factual birth instant falls between
// two solar-month Jie boundaries. It is evidence only until source rules carry
// independently reviewed phase applicability labels.
type TiaohouDepthEvidence struct {
	RuleID                 string                       `json:"rule_id"`
	Status                 string                       `json:"status"`
	Basis                  string                       `json:"basis"`
	Phase                  string                       `json:"phase,omitempty"`
	StartTerm              string                       `json:"start_term,omitempty"`
	StartAt                string                       `json:"start_at,omitempty"`
	EndTerm                string                       `json:"end_term,omitempty"`
	EndAt                  string                       `json:"end_at,omitempty"`
	ElapsedSeconds         int                          `json:"elapsed_seconds,omitempty"`
	IntervalSeconds        int                          `json:"interval_seconds,omitempty"`
	Position               float64                      `json:"position,omitempty"`
	MonthCommandCandidates []MonthCommandDepthCandidate `json:"month_command_candidates"`
	Note                   string                       `json:"note"`
	InterpretationStatus   string                       `json:"interpretation_status"`
}

// AnalyzeTiaohou performs tiaohou (调候) analysis for a given day stem and month branch.
// Returns analysis result based on 《穷通宝鉴》rules.
func AnalyzeTiaohou(dayStem, monthBranch string) (*TiaohouResult, error) {
	return analyzeTiaohou(dayStem, monthBranch, TiaohouDepthEvidence{
		RuleID: "bazi.tiaohou.solar-term-depth-v1",
		Status: "unavailable", Basis: "birth_instant_unavailable",
		MonthCommandCandidates: []MonthCommandDepthCandidate{},
		Note:                   "未提供可定位的出生时刻，不推断月令区间深浅。",
		InterpretationStatus:   "not_adjudicated",
	})
}

// AnalyzeTiaohouAt derives month depth from exact solar-term boundaries.
func AnalyzeTiaohouAt(dayStem, monthBranch string, birth tyme.SolarTime) (*TiaohouResult, error) {
	previousJie, nextJie := surroundingJie(birth)
	start := previousJie.GetJulianDay().GetSolarTime()
	end := nextJie.GetJulianDay().GetSolarTime()
	interval := end.Subtract(start)
	elapsed := birth.Subtract(start)
	if elapsed < 0 {
		elapsed = 0
	}
	if elapsed > interval {
		elapsed = interval
	}
	position := 0.0
	if interval > 0 {
		position = float64(elapsed) / float64(interval)
	}
	phase := "前段"
	if elapsed*3 >= interval*2 {
		phase = "后段"
	} else if elapsed*3 >= interval {
		phase = "中段"
	}
	return analyzeTiaohou(dayStem, monthBranch, TiaohouDepthEvidence{
		RuleID: "bazi.tiaohou.solar-term-depth-v1",
		Status: "observed", Basis: "solar_term_jie_interval", Phase: phase,
		StartTerm: previousJie.GetName(), StartAt: formatSolarTime(start),
		EndTerm: nextJie.GetName(), EndAt: formatSolarTime(end),
		ElapsedSeconds: elapsed, IntervalSeconds: interval, Position: position,
		MonthCommandCandidates: observeEarthMonthCommandCandidates(monthBranch, elapsed),
		Note:                   "区间位置仅作寒暖燥湿证据；当前规则表没有经 Gold 复核的分段适用标签，不据此改选主用。",
		InterpretationStatus:   "not_adjudicated",
	})
}

// AnalyzeTiaohouForPillars evaluates the reviewed chart-level conditions when
// the factual birth instant is unavailable.
func AnalyzeTiaohouForPillars(year, month, day, hour model.Pillar) (*TiaohouResult, error) {
	result, err := AnalyzeTiaohou(day.Gan, month.Zhi)
	if err != nil {
		return nil, err
	}
	return applyTiaohouChartContext(result, []model.Pillar{year, month, day, hour})
}

// AnalyzeTiaohouForPillarsAt combines exact Jie depth evidence with reviewed
// four-pillar conditions.
func AnalyzeTiaohouForPillarsAt(year, month, day, hour model.Pillar, birth tyme.SolarTime) (*TiaohouResult, error) {
	result, err := AnalyzeTiaohouAt(day.Gan, month.Zhi, birth)
	if err != nil {
		return nil, err
	}
	return applyTiaohouChartContext(result, []model.Pillar{year, month, day, hour})
}

func analyzeTiaohou(dayStem, monthBranch string, depth TiaohouDepthEvidence) (*TiaohouResult, error) {
	rules := data.GetTiaohou(dayStem, monthBranch)
	if len(rules) == 0 {
		return nil, fmt.Errorf("no tiaohou rules for stem=%s month=%s", dayStem, monthBranch)
	}

	evidenceRules := buildTiaohouRuleEvidence(dayStem, monthBranch, rules)
	limitations := []string{
		"table order is not an independently adjudicated unique selection",
		"only explicitly structured four-pillar conditions can become chart matches; remaining conditional table text stays source evidence",
		"chart condition matches do not adjudicate a unique useful god",
		"legacy row-level JiShen values are withheld because some contradict XiShen candidates in the same source sequence",
		"solar-term depth does not change candidate order",
		"table candidates do not imply favorable real-world outcomes",
	}
	if len(depth.MonthCommandCandidates) > 0 {
		limitations = append(limitations, "earth-month day-command profiles remain parallel evidence and do not alter body-strength or tiaohou selection")
	}

	return &TiaohouResult{
		RuleID:                tiaohouRuleID,
		Stem:                  dayStem,
		Month:                 monthBranch,
		Rules:                 evidenceRules,
		TablePrimaryCandidate: evidenceRules[0].XiShen,
		SelectionBasis:        "first_table_entry_candidate",
		DepthAffectsSelection: false,
		DepthEvidence:         depth,
		ChartEvidence: TiaohouChartEvidence{
			Status: "unavailable", Basis: "four_pillars_unavailable",
			VisibleStems: []string{}, Branches: []string{},
			CompleteBranchStructures: []TiaohouBranchStructure{},
		},
		MatchedConditions:    []TiaohouConditionHit{},
		ChartCandidates:      []string{},
		ChartSelectionBasis:  "chart_unavailable",
		Status:               "observed",
		ValidationStatus:     "not_validated",
		InterpretationStatus: "not_adjudicated",
		Limitations:          limitations,
	}, nil
}

func applyTiaohouChartContext(result *TiaohouResult, pillars []model.Pillar) (*TiaohouResult, error) {
	evidence, relations, err := buildTiaohouChartEvidence(pillars)
	if err != nil {
		return nil, err
	}
	hits, candidates := matchTiaohouChartRules(result.Stem, result.Month, evidence, relations)
	result.ChartEvidence = evidence
	result.MatchedConditions = hits
	result.ChartCandidates = candidates
	result.ChartSelectionBasis = "no_reviewed_chart_condition_match"
	if len(hits) > 0 {
		result.ChartSelectionBasis = "reviewed_four_pillar_condition_match"
	}
	return result, nil
}

func buildTiaohouChartEvidence(pillars []model.Pillar) (TiaohouChartEvidence, map[string]ZhiRelation, error) {
	if len(pillars) != 4 {
		return TiaohouChartEvidence{}, nil, fmt.Errorf("tiaohou chart requires four pillars, got %d", len(pillars))
	}
	analysis, err := CalcGanZhiAnalysis(pillars[0], pillars[1], pillars[2], pillars[3])
	if err != nil {
		return TiaohouChartEvidence{}, nil, err
	}
	evidence := TiaohouChartEvidence{
		Status: "observed", Basis: "four_pillars_and_complete_branch_structures",
		VisibleStems: make([]string, 0, 4), Branches: make([]string, 0, 4),
		CompleteBranchStructures: []TiaohouBranchStructure{},
	}
	for _, pillar := range pillars {
		evidence.VisibleStems = append(evidence.VisibleStems, pillar.Gan)
		evidence.Branches = append(evidence.Branches, pillar.Zhi)
	}
	relations := make(map[string]ZhiRelation)
	for _, relation := range analysis.ZhiRelations {
		if relation.Type != "三合局" || relation.StructureStatus != "complete_structure" {
			continue
		}
		relations[relation.RuleID] = relation
		evidence.CompleteBranchStructures = append(evidence.CompleteBranchStructures, TiaohouBranchStructure{
			RuleID: relation.RuleID, Type: relation.Type,
			Branches:             append([]string(nil), relation.Branches...),
			TargetElement:        relation.TargetElement,
			TransformationStatus: relation.TransformationStatus,
		})
	}
	sort.Slice(evidence.CompleteBranchStructures, func(i, j int) bool {
		return evidence.CompleteBranchStructures[i].RuleID < evidence.CompleteBranchStructures[j].RuleID
	})
	return evidence, relations, nil
}

func matchTiaohouChartRules(stem, month string, evidence TiaohouChartEvidence, relations map[string]ZhiRelation) ([]TiaohouConditionHit, []string) {
	hits := make([]TiaohouConditionHit, 0, 2)
	candidates := make([]string, 0, 2)
	for _, rule := range tiaohouChartRules {
		if rule.dayStem != stem || rule.monthBranch != month {
			continue
		}
		relation, ok := relations[rule.requiredStructureID]
		if !ok {
			continue
		}
		matched := true
		for _, required := range rule.requiredVisibleStems {
			if !tiaohouContainsString(evidence.VisibleStems, required) {
				matched = false
				break
			}
		}
		if !matched {
			continue
		}
		hitEvidence := []string{
			fmt.Sprintf("%s：%s三支齐备", relation.RuleID, relation.Subtype),
			"结构状态=" + relation.StructureStatus + "；成化状态=" + relation.TransformationStatus,
		}
		for _, required := range rule.requiredVisibleStems {
			hitEvidence = append(hitEvidence, "四柱天干见"+required)
		}
		hits = append(hits, TiaohouConditionHit{
			RuleID: rule.ruleID, Candidates: append([]string(nil), rule.candidates...),
			Condition: rule.condition, Evidence: hitEvidence,
			Source: rule.source, SourceText: rule.sourceText,
			Status: "matched", ValidationStatus: "classical_source_reviewed",
			InterpretationStatus: "not_adjudicated",
		})
		for _, candidate := range rule.candidates {
			if !tiaohouContainsString(candidates, candidate) {
				candidates = append(candidates, candidate)
			}
		}
	}
	return hits, candidates
}

func tiaohouContainsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func buildTiaohouRuleEvidence(stem, month string, rules []data.TiaohouRule) []TiaohouRuleEvidence {
	out := make([]TiaohouRuleEvidence, 0, len(rules))
	for index, rule := range rules {
		out = append(out, TiaohouRuleEvidence{
			RuleID:               fmt.Sprintf("bazi.tiaohou.table.%s.%s.%d", stem, month, index+1),
			XiShen:               rule.XiShen,
			JiShen:               tiaohouJiShenUnadjudicated,
			JiShenStatus:         "not_adjudicated",
			SourceText:           rule.Reason,
			Basis:                "day_stem_month_branch_table",
			Status:               "observed",
			ValidationStatus:     "not_validated",
			InterpretationStatus: "not_adjudicated",
		})
	}
	return out
}

func ValidTiaohouEvidence(result *TiaohouResult, stem, month string) bool {
	if result == nil || result.RuleID != tiaohouRuleID ||
		result.Stem != stem || result.Month != month ||
		result.SelectionBasis != "first_table_entry_candidate" ||
		result.DepthAffectsSelection || result.Status != "observed" ||
		result.ValidationStatus != "not_validated" ||
		result.InterpretationStatus != "not_adjudicated" {
		return false
	}
	expectedRules := buildTiaohouRuleEvidence(stem, month, data.GetTiaohou(stem, month))
	if len(expectedRules) == 0 || !reflect.DeepEqual(result.Rules, expectedRules) ||
		result.TablePrimaryCandidate != expectedRules[0].XiShen {
		return false
	}
	for _, rule := range result.Rules {
		if rule.JiShen != tiaohouJiShenUnadjudicated || rule.JiShenStatus != "not_adjudicated" {
			return false
		}
	}
	if !validTiaohouChartEvidenceShape(result) {
		return false
	}
	depth := result.DepthEvidence
	if depth.RuleID != "bazi.tiaohou.solar-term-depth-v1" ||
		depth.InterpretationStatus != "not_adjudicated" {
		return false
	}
	switch depth.Status {
	case "unavailable":
		return depth.Basis == "birth_instant_unavailable" && len(depth.MonthCommandCandidates) == 0
	case "observed":
		return depth.Basis == "solar_term_jie_interval" &&
			depth.StartTerm != "" && depth.StartAt != "" &&
			depth.EndTerm != "" && depth.EndAt != "" &&
			depth.IntervalSeconds > 0 && depth.ElapsedSeconds >= 0 &&
			depth.ElapsedSeconds <= depth.IntervalSeconds &&
			depth.Position >= 0 && depth.Position <= 1 &&
			(depth.Phase == "前段" || depth.Phase == "中段" || depth.Phase == "后段") &&
			reflect.DeepEqual(depth.MonthCommandCandidates, observeEarthMonthCommandCandidates(month, depth.ElapsedSeconds))
	default:
		return false
	}
}

// ValidTiaohouEvidenceForPillars additionally verifies that chart-condition
// matches are reproducible from the four factual pillars.
func ValidTiaohouEvidenceForPillars(result *TiaohouResult, pillars []model.Pillar) bool {
	if len(pillars) != 4 || !ValidTiaohouEvidence(result, pillars[2].Gan, pillars[1].Zhi) {
		return false
	}
	evidence, relations, err := buildTiaohouChartEvidence(pillars)
	if err != nil {
		return false
	}
	hits, candidates := matchTiaohouChartRules(pillars[2].Gan, pillars[1].Zhi, evidence, relations)
	basis := "no_reviewed_chart_condition_match"
	if len(hits) > 0 {
		basis = "reviewed_four_pillar_condition_match"
	}
	return reflect.DeepEqual(result.ChartEvidence, evidence) &&
		reflect.DeepEqual(result.MatchedConditions, hits) &&
		reflect.DeepEqual(result.ChartCandidates, candidates) &&
		result.ChartSelectionBasis == basis
}

func validTiaohouChartEvidenceShape(result *TiaohouResult) bool {
	switch result.ChartEvidence.Status {
	case "unavailable":
		return result.ChartEvidence.Basis == "four_pillars_unavailable" &&
			len(result.ChartEvidence.VisibleStems) == 0 && len(result.ChartEvidence.Branches) == 0 &&
			len(result.ChartEvidence.CompleteBranchStructures) == 0 &&
			len(result.MatchedConditions) == 0 && len(result.ChartCandidates) == 0 &&
			result.ChartSelectionBasis == "chart_unavailable"
	case "observed":
		if result.ChartEvidence.Basis != "four_pillars_and_complete_branch_structures" ||
			len(result.ChartEvidence.VisibleStems) != 4 || len(result.ChartEvidence.Branches) != 4 {
			return false
		}
		if len(result.MatchedConditions) == 0 {
			return len(result.ChartCandidates) == 0 && result.ChartSelectionBasis == "no_reviewed_chart_condition_match"
		}
		if len(result.ChartCandidates) == 0 || result.ChartSelectionBasis != "reviewed_four_pillar_condition_match" {
			return false
		}
		for _, hit := range result.MatchedConditions {
			if hit.RuleID == "" || len(hit.Candidates) == 0 || hit.Condition == "" || len(hit.Evidence) == 0 ||
				hit.Source == "" || hit.SourceText == "" || hit.Status != "matched" ||
				hit.ValidationStatus != "classical_source_reviewed" || hit.InterpretationStatus != "not_adjudicated" {
				return false
			}
		}
		return true
	default:
		return false
	}
}
