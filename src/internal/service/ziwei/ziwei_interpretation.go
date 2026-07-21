package ziwei

import (
	"fmt"
	"sort"
	"strings"

	"bazi/internal/service/bazi"
)

const (
	periodHourStemRuleID     = "ziwei.period.hour-stem.five-rat-v1"
	periodHourBoundaryPolicy = "traditional_two_hour_branch_slots_no_civil_date_assignment"
)

// ──────────────────── Period Interpretation Service ────────────────────

// PeriodInterpreter projects 流年/流月/流日 data into deterministic GanZhi,
// ShiShen and branch-relation evidence. It does not assign favorable or
// unfavorable scores.
type PeriodInterpreter struct {
	birthData       *BirthData
	baseContentHash string
}

// NewPeriodInterpreterFromChart restores period interpretation context from
// the authenticated public natal-chart contract. It returns nil for derived,
// incomplete, or tampered chart payloads.
func NewPeriodInterpreterFromChart(chart *ZiWeiChart) *PeriodInterpreter {
	if !chartMatchesDeclaredProfile(chart) {
		return nil
	}
	birth, ok := birthDataFromPublishedChart(chart)
	if !ok {
		return nil
	}
	return &PeriodInterpreter{birthData: birth, baseContentHash: chart.ContentHash}
}

// ──────────────────── Result types ────────────────────

type LiunianResult struct {
	Year                int                    `json:"year"`
	GanZhi              string                 `json:"gan_zhi"`
	GanZhiDesc          string                 `json:"gan_zhi_desc"`
	ShiShen             string                 `json:"shi_shen"`
	RelationToMing      string                 `json:"relation_to_ming"`
	RelationEvidence    []PeriodBranchRelation `json:"relation_evidence"`
	StructuralSummary   string                 `json:"structural_summary"`
	ReviewNote          string                 `json:"review_note"`
	EvidenceBasis       string                 `json:"evidence_basis"`
	ValidationStatus    string                 `json:"validation_status"`
	IsOutcomeConclusion bool                   `json:"is_outcome_conclusion"`
}

type LiuyueResult struct {
	Year                int                    `json:"year"`
	Month               int                    `json:"month"`
	GanZhi              string                 `json:"gan_zhi"`
	GanZhiDesc          string                 `json:"gan_zhi_desc"`
	ShiShen             string                 `json:"shi_shen"`
	RelationToMing      string                 `json:"relation_to_ming"`
	RelationEvidence    []PeriodBranchRelation `json:"relation_evidence"`
	StructuralSummary   string                 `json:"structural_summary"`
	EvidenceBasis       string                 `json:"evidence_basis"`
	ValidationStatus    string                 `json:"validation_status"`
	IsOutcomeConclusion bool                   `json:"is_outcome_conclusion"`
}

type LiuriResult struct {
	Year                int                    `json:"year"`
	Month               int                    `json:"month"`
	Day                 int                    `json:"day"`
	GanZhi              string                 `json:"gan_zhi"`
	GanZhiDesc          string                 `json:"gan_zhi_desc"`
	ShiShen             string                 `json:"shi_shen"`
	RelationToMing      string                 `json:"relation_to_ming"`
	RelationEvidence    []PeriodBranchRelation `json:"relation_evidence"`
	HourlyAnalysis      []HourBlock            `json:"hourly_analysis"`
	StructuralSummary   string                 `json:"structural_summary"`
	EvidenceBasis       string                 `json:"evidence_basis"`
	ValidationStatus    string                 `json:"validation_status"`
	IsOutcomeConclusion bool                   `json:"is_outcome_conclusion"`
}

type HourBlock struct {
	Stem                     string                 `json:"stem"`
	Branch                   string                 `json:"branch"`
	StemBranch               string                 `json:"stem_branch"`
	IntervalStartHour        int                    `json:"interval_start_hour"`
	IntervalEndHourExclusive int                    `json:"interval_end_hour_exclusive"`
	IntervalLabel            string                 `json:"interval_label"`
	CrossesMidnight          bool                   `json:"crosses_midnight"`
	DayStemBasis             string                 `json:"day_stem_basis"`
	BoundaryPolicy           string                 `json:"boundary_policy"`
	RuleID                   string                 `json:"rule_id"`
	ShiShen                  string                 `json:"shi_shen"`
	RelationToMing           string                 `json:"relation_to_ming"`
	RelationEvidence         []PeriodBranchRelation `json:"relation_evidence"`
	StructuralSummary        string                 `json:"structural_summary"`
	EvidenceBasis            string                 `json:"evidence_basis"`
	ValidationStatus         string                 `json:"validation_status"`
	IsOutcomeConclusion      bool                   `json:"is_outcome_conclusion"`
}

type PeriodBranchRelation struct {
	PeriodBranch         string `json:"period_branch"`
	NatalPillar          string `json:"natal_pillar"`
	NatalBranch          string `json:"natal_branch"`
	Relation             string `json:"relation"`
	Subtype              string `json:"subtype,omitempty"`
	RuleID               string `json:"rule_id"`
	StructuralStatus     string `json:"structural_status"`
	TransformationStatus string `json:"transformation_status"`
	TargetElement        string `json:"target_element,omitempty"`
	EvidenceBasis        string `json:"evidence_basis"`
	InterpretationStatus string `json:"interpretation_status"`
	IsOutcomeConclusion  bool   `json:"is_outcome_conclusion"`
}

// PeriodSummary holds the summary of all three layers.
type PeriodSummary struct {
	Liunian             LiunianSummaryItem `json:"liunian"`
	Liuyue              LiuyueSummaryItem  `json:"liuyue"`
	Liuri               LiuriSummaryItem   `json:"liuri"`
	ReviewNotes         PeriodReviewNotes  `json:"review_notes"`
	EvidenceBasis       string             `json:"evidence_basis"`
	ValidationStatus    string             `json:"validation_status"`
	IsOutcomeConclusion bool               `json:"is_outcome_conclusion"`
}

type LiunianSummaryItem struct {
	GanZhi            string                 `json:"gan_zhi"`
	ShiShen           string                 `json:"shi_shen"`
	Relation          string                 `json:"relation"`
	RelationEvidence  []PeriodBranchRelation `json:"relation_evidence"`
	StructuralSummary string                 `json:"structural_summary"`
}

type LiuyueSummaryItem struct {
	GanZhi            string                 `json:"gan_zhi"`
	ShiShen           string                 `json:"shi_shen"`
	Relation          string                 `json:"relation"`
	RelationEvidence  []PeriodBranchRelation `json:"relation_evidence"`
	StructuralSummary string                 `json:"structural_summary"`
}

type LiuriSummaryItem struct {
	GanZhi            string                 `json:"gan_zhi"`
	ShiShen           string                 `json:"shi_shen"`
	Relation          string                 `json:"relation"`
	RelationEvidence  []PeriodBranchRelation `json:"relation_evidence"`
	StructuralSummary string                 `json:"structural_summary"`
}

type PeriodReviewNotes struct {
	Liunian []string `json:"liunian"`
	Liuyue  []string `json:"liuyue"`
	Liuri   []string `json:"liuri"`
}

// stemName returns the Chinese name of a stem by index.
func stemName(stem int) string {
	if stem >= 0 && stem < len(StemNames) {
		return StemNames[stem]
	}
	return ""
}

// branchName returns the Chinese name of a branch by index.
func branchName(branch int) string {
	if branch >= 0 && branch < len(BranchNames) {
		return BranchNames[branch]
	}
	return ""
}

// wuXingStem returns the five-element name for a stem.
func wuXingStem(stem int) string {
	switch stem {
	case 0, 1:
		return "木"
	case 2, 3:
		return "火"
	case 4, 5:
		return "土"
	case 6, 7:
		return "金"
	case 8, 9:
		return "水"
	}
	return ""
}

// wuXingBranch returns the five-element name for a branch.
func wuXingBranch(branch int) string {
	switch branch {
	case 0, 11:
		return "水"
	case 2, 3:
		return "木"
	case 5, 6:
		return "火"
	case 1, 4, 7, 10:
		return "土"
	case 8, 9:
		return "金"
	}
	return ""
}

// stemWuXingIdx returns the 5-element index for stem (0=木, 1=火, 2=土, 3=金, 4=水).
func stemWuXingIdx(stem int) int {
	switch stem {
	case 0, 1:
		return 0 // 木
	case 2, 3:
		return 1 // 火
	case 4, 5:
		return 2 // 土
	case 6, 7:
		return 3 // 金
	case 8, 9:
		return 4 // 水
	}
	return 0
}

// getShiShen classifies a visible period stem against the natal day stem.
// Index validation stays here because the shared BaZi classifier accepts names.
func getShiShen(stem, dayStem int) (string, bool) {
	if stem < 0 || stem >= len(StemNames) || dayStem < 0 || dayStem >= len(StemNames) {
		return "", false
	}
	name := bazi.ClassifyTenGod(StemNames[stem], StemNames[dayStem], false)
	return name, name != ""
}

type periodPairRelationRule struct {
	relation             string
	subtype              string
	ruleID               string
	structuralStatus     string
	transformationStatus string
	targetElement        string
	priority             int
}

func relationTypes(b1, b2 int) []string {
	rules := periodPairRelationRules(b1, b2)
	relations := make([]string, 0, len(rules))
	for _, rule := range rules {
		relations = append(relations, periodRelationLabel(rule))
	}
	return relations
}

func periodPairRelationRules(b1, b2 int) []periodPairRelationRule {
	if b1 < 0 || b1 >= len(BranchNames) || b2 < 0 || b2 >= len(BranchNames) {
		return []periodPairRelationRule{}
	}
	pair := canonicalPeriodBranchPair(b1, b2)
	rules := make([]periodPairRelationRule, 0, 4)
	if b1 == b2 {
		rules = append(rules, periodPairRelationRule{
			relation: "伏吟", ruleID: "ziwei.period.branch.fuyin." + branchName(b1) + "-v1",
			structuralStatus: "observed", transformationStatus: "not_applicable", priority: 50,
		})
		if isSelfPunishmentBranch(b1) {
			rules = append(rules, periodPairRelationRule{
				relation: "相刑", subtype: "自刑", ruleID: "ziwei.period.branch.punish." + pair + "-v1",
				structuralStatus: "observed", transformationStatus: "not_applicable", priority: 90,
			})
		}
		return sortPeriodPairRelationRules(rules)
	}

	if isPeriodBranchPair(b1, b2, [][2]int{{0, 6}, {1, 7}, {2, 8}, {3, 9}, {4, 10}, {5, 11}}) {
		rules = append(rules, periodPairRelationRule{
			relation: "六冲", ruleID: "ziwei.period.branch.clash." + pair + "-v1",
			structuralStatus: "observed", transformationStatus: "not_applicable", priority: 100,
		})
	}
	if isPeriodBranchPair(b1, b2, [][2]int{{0, 1}, {2, 11}, {3, 10}, {4, 9}, {5, 8}, {6, 7}}) {
		rules = append(rules, periodPairRelationRule{
			relation: "六合", ruleID: "ziwei.period.branch.liuhe." + pair + "-v1",
			structuralStatus: "complete", transformationStatus: "unadjudicated",
			targetElement: periodLiuHeTargetElement(pair), priority: 60,
		})
	}
	if isPeriodBranchPair(b1, b2, [][2]int{{0, 7}, {1, 6}, {2, 5}, {3, 4}, {8, 11}, {9, 10}}) {
		rules = append(rules, periodPairRelationRule{
			relation: "六害", ruleID: "ziwei.period.branch.harm." + pair + "-v1",
			structuralStatus: "observed", transformationStatus: "not_applicable", priority: 80,
		})
	}
	if subtype := periodPairPunishmentSubtype(b1, b2); subtype != "" {
		rules = append(rules, periodPairRelationRule{
			relation: "相刑", subtype: subtype, ruleID: "ziwei.period.branch.punish." + pair + "-v1",
			structuralStatus: "observed", transformationStatus: "not_applicable", priority: 90,
		})
	}
	if isPeriodBranchPair(b1, b2, [][2]int{{0, 9}, {1, 4}, {2, 11}, {3, 6}, {5, 8}, {7, 10}}) {
		rules = append(rules, periodPairRelationRule{
			relation: "六破", ruleID: "ziwei.period.branch.break." + pair + "-v1",
			structuralStatus: "observed", transformationStatus: "not_applicable", priority: 70,
		})
	}
	return sortPeriodPairRelationRules(rules)
}

func sortPeriodPairRelationRules(rules []periodPairRelationRule) []periodPairRelationRule {
	sort.SliceStable(rules, func(i, j int) bool {
		if rules[i].priority != rules[j].priority {
			return rules[i].priority > rules[j].priority
		}
		return rules[i].ruleID < rules[j].ruleID
	})
	return rules
}

func periodRelationLabel(rule periodPairRelationRule) string {
	if rule.subtype == "" {
		return rule.relation
	}
	return fmt.Sprintf("%s（%s）", rule.relation, rule.subtype)
}

func canonicalPeriodBranchPair(b1, b2 int) string {
	if b1 <= b2 {
		return branchName(b1) + branchName(b2)
	}
	return branchName(b2) + branchName(b1)
}

func isPeriodBranchPair(b1, b2 int, pairs [][2]int) bool {
	for _, pair := range pairs {
		if (b1 == pair[0] && b2 == pair[1]) || (b1 == pair[1] && b2 == pair[0]) {
			return true
		}
	}
	return false
}

func isSelfPunishmentBranch(branch int) bool {
	switch branch {
	case 4, 6, 9, 11:
		return true
	default:
		return false
	}
}

func periodPairPunishmentSubtype(b1, b2 int) string {
	switch {
	case isPeriodBranchPair(b1, b2, [][2]int{{0, 3}}):
		return "无礼之刑"
	case pairInPeriodBranchGroup(b1, b2, []int{2, 5, 8}):
		return "无恩之刑"
	case pairInPeriodBranchGroup(b1, b2, []int{1, 10, 7}):
		return "恃势之刑"
	default:
		return ""
	}
}

func pairInPeriodBranchGroup(b1, b2 int, group []int) bool {
	if b1 == b2 {
		return false
	}
	found1, found2 := false, false
	for _, branch := range group {
		found1 = found1 || branch == b1
		found2 = found2 || branch == b2
	}
	return found1 && found2
}

func periodLiuHeTargetElement(pair string) string {
	return map[string]string{
		"子丑": "土", "寅亥": "木", "卯戌": "火",
		"辰酉": "金", "巳申": "水", "午未": "土",
	}[pair]
}

// describeRelation creates a Chinese description of how period branch relates to birth branches.
func (s *PeriodInterpreter) describeRelation(periodBranch int) string {
	evidence := s.relationEvidence(periodBranch)
	if evidence == nil {
		return "命局参数无效"
	}
	if len(evidence) == 0 {
		return "与命局无特殊刑冲合关系"
	}

	seen := map[string]bool{}
	unique := []string{}
	for _, item := range evidence {
		if !seen[item.Relation] {
			seen[item.Relation] = true
			unique = append(unique, item.Relation)
		}
	}
	return strings.Join(unique, "、")
}

func (s *PeriodInterpreter) relationEvidence(periodBranch int) []PeriodBranchRelation {
	if s == nil || s.birthData == nil {
		return nil
	}
	keyBranches := []struct {
		pillar string
		branch int
	}{
		{pillar: "year", branch: s.birthData.YearBranch},
		{pillar: "month", branch: s.birthData.MonthPillarBranch},
		{pillar: "day", branch: s.birthData.DayBranch},
		{pillar: "hour", branch: s.birthData.HourBranch},
	}

	out := make([]PeriodBranchRelation, 0, len(keyBranches))
	for _, natal := range keyBranches {
		for _, rule := range periodPairRelationRules(periodBranch, natal.branch) {
			out = append(out, PeriodBranchRelation{
				PeriodBranch:         branchName(periodBranch),
				NatalPillar:          natal.pillar,
				NatalBranch:          branchName(natal.branch),
				Relation:             rule.relation,
				Subtype:              rule.subtype,
				RuleID:               rule.ruleID,
				StructuralStatus:     rule.structuralStatus,
				TransformationStatus: rule.transformationStatus,
				TargetElement:        rule.targetElement,
				EvidenceBasis:        "deterministic_rule_projection",
				InterpretationStatus: "not_adjudicated",
				IsOutcomeConclusion:  false,
			})
		}
	}
	return out
}

// ──────────────────── Analysis methods ────────────────────

// AnalyzeLiunian produces a full interpretation for the given 流年.
func (s *PeriodInterpreter) AnalyzeLiunian(chart *ZiWeiChart, year int) *LiunianResult {
	if !s.acceptsDerivedChart(chart) {
		return nil
	}
	liunianStem, liunianBranch, ok := chartDerivationForQuery(chart, "liunian", year, 0, 0)
	if !ok {
		return nil
	}

	ganZhi := stemName(liunianStem) + branchName(liunianBranch)
	shiShen, ok := getShiShen(liunianStem, s.birthData.DayStem)
	if !ok {
		return nil
	}
	rel := s.describeRelation(liunianBranch)
	relations := s.relationEvidence(liunianBranch)
	structuralSummary := fmt.Sprintf("流年%s透出%s，地支与命局关系记录为%s；仅陈述规则结构。", ganZhi, shiShen, rel)
	tips := "本结果仅用于核对干支、十神与刑冲合结构；职业、学业和财务决定应依据现实证据。"

	return &LiunianResult{
		Year:                year,
		GanZhi:              ganZhi,
		GanZhiDesc:          fmt.Sprintf("%s（%s）+ %s（%s）", stemName(liunianStem), shiShen, branchName(liunianBranch), wuXingBranch(liunianBranch)),
		ShiShen:             shiShen,
		RelationToMing:      rel,
		RelationEvidence:    relations,
		StructuralSummary:   structuralSummary,
		ReviewNote:          tips,
		EvidenceBasis:       "deterministic_rule_projection",
		ValidationStatus:    "not_adjudicated",
		IsOutcomeConclusion: false,
	}
}

// AnalyzeLiuyue produces a full interpretation for the given 流月.
func (s *PeriodInterpreter) AnalyzeLiuyue(chart *ZiWeiChart, year, month, day int) *LiuyueResult {
	if !s.acceptsDerivedChart(chart) {
		return nil
	}
	liuyueStem, liuyueBranch, ok := chartDerivationForQuery(chart, "liuyue", year, month, day)
	if !ok {
		return nil
	}
	dayStem := s.birthData.DayStem

	ganZhi := stemName(liuyueStem) + branchName(liuyueBranch)
	shiShen, ok := getShiShen(liuyueStem, dayStem)
	if !ok {
		return nil
	}
	rel := s.describeRelation(liuyueBranch)
	relations := s.relationEvidence(liuyueBranch)
	structuralSummary := fmt.Sprintf("流月%s透出%s，地支与命局关系记录为%s；不推导现实事件结果。", ganZhi, shiShen, rel)

	return &LiuyueResult{
		Year:                year,
		Month:               month,
		GanZhi:              ganZhi,
		GanZhiDesc:          fmt.Sprintf("%s（%s）+ %s（%s）", stemName(liuyueStem), shiShen, branchName(liuyueBranch), wuXingBranch(liuyueBranch)),
		ShiShen:             shiShen,
		RelationToMing:      rel,
		RelationEvidence:    relations,
		StructuralSummary:   structuralSummary,
		EvidenceBasis:       "deterministic_rule_projection",
		ValidationStatus:    "not_adjudicated",
		IsOutcomeConclusion: false,
	}
}

// AnalyzeLiuri produces a full interpretation for the given 流日.
func (s *PeriodInterpreter) AnalyzeLiuri(chart *ZiWeiChart, year, month, day int) *LiuriResult {
	if !s.acceptsDerivedChart(chart) {
		return nil
	}
	liuriStem, liuriBranch, ok := chartDerivationForQuery(chart, "liuri", year, month, day)
	if !ok {
		return nil
	}
	dayStem := s.birthData.DayStem

	ganZhi := stemName(liuriStem) + branchName(liuriBranch)
	shiShen, ok := getShiShen(liuriStem, dayStem)
	if !ok {
		return nil
	}
	rel := s.describeRelation(liuriBranch)
	relations := s.relationEvidence(liuriBranch)

	hourly := make([]HourBlock, 12)

	for i := 0; i < 12; i++ {
		hourBranch := i
		hourStemIdx, ok := fiveRatHourStem(liuriStem, hourBranch)
		if !ok {
			return nil
		}
		startHour, endHour, intervalLabel, crossesMidnight, ok := traditionalHourInterval(hourBranch)
		if !ok {
			return nil
		}
		hourGanZhi := stemName(hourStemIdx) + branchName(hourBranch)
		hourShiShen, ok := getShiShen(hourStemIdx, dayStem)
		if !ok {
			return nil
		}
		hourRelation := s.describeRelation(hourBranch)
		hourRelations := s.relationEvidence(hourBranch)

		hourly[i] = HourBlock{
			Stem:                     stemName(hourStemIdx),
			Branch:                   branchName(hourBranch),
			StemBranch:               hourGanZhi,
			IntervalStartHour:        startHour,
			IntervalEndHourExclusive: endHour,
			IntervalLabel:            intervalLabel,
			CrossesMidnight:          crossesMidnight,
			DayStemBasis:             "period_derivation_day_stem",
			BoundaryPolicy:           periodHourBoundaryPolicy,
			RuleID:                   periodHourStemRuleID,
			ShiShen:                  hourShiShen,
			RelationToMing:           hourRelation,
			RelationEvidence:         hourRelations,
			StructuralSummary:        fmt.Sprintf("%s时（%s）为%s，透出%s；地支与命局关系记录为%s。", branchName(hourBranch), intervalLabel, hourGanZhi, hourShiShen, hourRelation),
			EvidenceBasis:            "deterministic_rule_projection",
			ValidationStatus:         "not_adjudicated",
			IsOutcomeConclusion:      false,
		}
	}

	summary := fmt.Sprintf("流日%s透出%s，地支与命局关系记录为%s；十二时辰按传统两小时地支区间展示，子时跨午夜但不在此合同内裁决民用日期归属。", ganZhi, shiShen, rel)

	return &LiuriResult{
		Year:                year,
		Month:               month,
		Day:                 day,
		GanZhi:              ganZhi,
		GanZhiDesc:          fmt.Sprintf("%s（%s）+ %s（%s）", stemName(liuriStem), shiShen, branchName(liuriBranch), wuXingBranch(liuriBranch)),
		ShiShen:             shiShen,
		RelationToMing:      rel,
		RelationEvidence:    relations,
		HourlyAnalysis:      hourly,
		StructuralSummary:   summary,
		EvidenceBasis:       "deterministic_rule_projection",
		ValidationStatus:    "not_adjudicated",
		IsOutcomeConclusion: false,
	}
}

func fiveRatHourStem(dayStem, hourBranch int) (int, bool) {
	if dayStem < 0 || dayStem >= len(StemNames) || hourBranch < 0 || hourBranch >= len(BranchNames) {
		return 0, false
	}
	return ((dayStem%5)*2 + hourBranch) % len(StemNames), true
}

func traditionalHourInterval(hourBranch int) (startHour, endHourExclusive int, label string, crossesMidnight, ok bool) {
	if hourBranch < 0 || hourBranch >= len(BranchNames) {
		return 0, 0, "", false, false
	}
	startHour = hourBranch*2 - 1
	if startHour < 0 {
		startHour = 23
	}
	endHourExclusive = (hourBranch*2 + 1) % 24
	endDisplayHour := (endHourExclusive + 23) % 24
	label = fmt.Sprintf("%02d:00-%02d:59", startHour, endDisplayHour)
	return startHour, endHourExclusive, label, hourBranch == 0, true
}

func (s *PeriodInterpreter) acceptsDerivedChart(chart *ZiWeiChart) bool {
	if s == nil || s.birthData == nil || s.baseContentHash == "" || !ValidDerivedChartContract(chart) {
		return false
	}
	return chart.BaseContentHash == s.baseContentHash
}

// SummarizeAll produces a summary combining all three period layers.
func (s *PeriodInterpreter) SummarizeAll(liunian, liuyue, liuri *ZiWeiChart, year, month, day int) *PeriodSummary {
	if s == nil || s.birthData == nil || liunian == nil || liunian.DerivationInput == nil {
		return nil
	}
	lunarYearLabel, err := LunarYearLabelForSolarDate(year, month, day)
	if err != nil || liunian.DerivationInput.Year != lunarYearLabel {
		return nil
	}
	ln := s.AnalyzeLiunian(liunian, liunian.DerivationInput.Year)
	ly := s.AnalyzeLiuyue(liuyue, year, month, day)
	lr := s.AnalyzeLiuri(liuri, year, month, day)
	if ln == nil || ly == nil || lr == nil {
		return nil
	}

	var reviewNotes PeriodReviewNotes
	reviewNotes.Liunian = []string{"仅核对流年干支、十神和刑冲合结构", "重要决定以现实证据和专业意见为准"}
	reviewNotes.Liuyue = []string{"仅核对流月干支、十神和刑冲合结构", "周期结构不得作为财务或职业决策依据"}
	reviewNotes.Liuri = []string{"仅核对流日与时辰规则结构", "时辰结构不代表吉凶或行动建议"}

	return &PeriodSummary{
		Liunian: LiunianSummaryItem{
			GanZhi:            ln.GanZhi,
			ShiShen:           ln.ShiShen,
			Relation:          ln.RelationToMing,
			RelationEvidence:  append([]PeriodBranchRelation{}, ln.RelationEvidence...),
			StructuralSummary: ln.StructuralSummary,
		},
		Liuyue: LiuyueSummaryItem{
			GanZhi:            ly.GanZhi,
			ShiShen:           ly.ShiShen,
			Relation:          ly.RelationToMing,
			RelationEvidence:  append([]PeriodBranchRelation{}, ly.RelationEvidence...),
			StructuralSummary: ly.StructuralSummary,
		},
		Liuri: LiuriSummaryItem{
			GanZhi:            lr.GanZhi,
			ShiShen:           lr.ShiShen,
			Relation:          lr.RelationToMing,
			RelationEvidence:  append([]PeriodBranchRelation{}, lr.RelationEvidence...),
			StructuralSummary: lr.StructuralSummary,
		},
		ReviewNotes:         reviewNotes,
		EvidenceBasis:       "deterministic_rule_projection",
		ValidationStatus:    "not_adjudicated",
		IsOutcomeConclusion: false,
	}
}
