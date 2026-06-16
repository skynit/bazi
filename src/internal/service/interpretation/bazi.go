package interpretation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/data"
	"bazi/internal/service/rag"
)

const (
	StatusOK       = "ok"
	StatusFallback = "fallback"

	ReasonDisabled        = "disabled"
	ReasonNotConfigured   = "not_configured"
	ReasonTimeout         = "timeout"
	ReasonEmpty           = "empty"
	ReasonUpstreamError   = "upstream_error"
	ReasonInvalidResponse = "invalid_response"
	ReasonRuleOnly        = "rule_only"
)

var (
	ErrChartIDRequired = errors.New("chart_id is required")
	ErrChartStore      = errors.New("chart store is required")
	ErrChartNotFound   = errors.New("chart not found")
	ErrComputeChart    = errors.New("failed to compute chart")
)

type ChartStore interface {
	FindByIDForUser(id uint, userID uint) (*model.BirthChart, error)
}

type Service struct {
	Charts    ChartStore
	Bazi      *bazipkg.BaziService
	Retriever rag.Retriever
	MinScore  float64
	TopK      int
}

type Request struct {
	ChartID uint
	UserID  uint
	Focus   string
}

type QueryParts struct {
	Question string
}

type sectionChunks struct {
	Focus  string
	Title  string
	Chunks []rag.RetrievedChunk
}

type evidencePoint struct {
	Text       string
	CitationID int
}

func (s *Service) InterpretBazi(ctx context.Context, req Request) (model.BaziInterpretationResponse, error) {
	focus := normalizeFocus(req.Focus)
	resp := model.BaziInterpretationResponse{
		Status:    StatusFallback,
		Reason:    ReasonRuleOnly,
		ChartID:   req.ChartID,
		Focus:     focus,
		Sections:  []model.InterpretationSection{},
		Citations: []model.InterpretationCitation{},
	}

	if req.ChartID == 0 {
		return resp, ErrChartIDRequired
	}
	if s == nil || s.Charts == nil {
		return resp, ErrChartStore
	}

	chart, err := s.Charts.FindByIDForUser(req.ChartID, req.UserID)
	if err != nil || chart == nil {
		return resp, ErrChartNotFound
	}

	baziSvc := s.Bazi
	if baziSvc == nil {
		baziSvc = &bazipkg.BaziService{}
	}
	result, err := baziSvc.Calculate(
		chart.BirthYear,
		chart.BirthMonth,
		chart.BirthDay,
		chart.BirthHour,
		chart.BirthMin,
		normalizeGender(chart.Gender),
	)
	if err != nil {
		return resp, fmt.Errorf("%w: %v", ErrComputeChart, err)
	}

	resp.Summary = buildSummary(result, focus, false)
	resp.Sections = buildSections(result, focus, nil)

	if s.Retriever == nil {
		resp.Reason = ReasonDisabled
		return resp, nil
	}

	evidence, reason := s.retrieveSectionChunks(ctx, result, focus)
	if reason != "" {
		resp.Reason = reason
		return resp, nil
	}
	if len(evidence) == 0 {
		resp.Reason = ReasonEmpty
		return resp, nil
	}

	merged := mergeSectionChunks(evidence, s.topK())
	if len(merged) == 0 {
		resp.Reason = ReasonEmpty
		return resp, nil
	}

	resp.Citations = buildCitations(merged)
	resp.Sections = buildEvidenceSections(result, focus, evidence, resp.Citations)
	resp.Summary = buildSummary(result, focus, true)
	resp.Status = StatusOK
	resp.Reason = ""
	return resp, nil
}

func (s *Service) retrieveSectionChunks(ctx context.Context, result *bazipkg.BaziResult, focus string) ([]sectionChunks, string) {
	focuses := sectionFocuses(focus)
	out := make([]sectionChunks, 0, len(focuses))
	firstReason := ""
	for _, sectionFocus := range focuses {
		chunks, err := s.Retriever.Retrieve(ctx, rag.RetrieveRequest{
			Question: buildQuery(result, sectionFocus),
			Focus:    sectionFocus,
		})
		if err != nil {
			if firstReason == "" {
				firstReason = mapRetrieverError(err)
			}
			continue
		}
		filtered := filterChunks(chunks, s.minScore(), sectionTopK(s.topK(), focus))
		out = append(out, sectionChunks{
			Focus:  sectionFocus,
			Title:  sectionTitle(sectionFocus),
			Chunks: filtered,
		})
	}
	if hasAnyChunks(out) {
		return out, ""
	}
	if firstReason != "" {
		return nil, firstReason
	}
	return out, ""
}

func sectionFocuses(focus string) []string {
	if focus == "overview" {
		return []string{"pattern", "tiaohou", "ten_gods"}
	}
	return []string{focus}
}

func sectionTitle(focus string) string {
	switch focus {
	case "pattern":
		return "格局与月令"
	case "tiaohou":
		return "调候与平衡"
	case "ten_gods":
		return "十神结构"
	default:
		return "综合解读"
	}
}

func sectionTopK(topK int, focus string) int {
	if topK <= 0 {
		topK = 8
	}
	if focus == "overview" && topK < 4 {
		return 4
	}
	return topK
}

func hasAnyChunks(sections []sectionChunks) bool {
	for _, section := range sections {
		if len(section.Chunks) > 0 {
			return true
		}
	}
	return false
}

func (s *Service) minScore() float64 {
	if s == nil || s.MinScore <= 0 {
		return 0.35
	}
	return s.MinScore
}

func (s *Service) topK() int {
	if s == nil || s.TopK <= 0 {
		return 8
	}
	return s.TopK
}

func normalizeFocus(focus string) string {
	switch strings.TrimSpace(focus) {
	case "", "overview":
		return "overview"
	case "pattern", "tiaohou", "ten_gods":
		return focus
	default:
		return "overview"
	}
}

func normalizeGender(g string) string {
	s := strings.TrimSpace(g)
	switch {
	case s == "男" || strings.EqualFold(s, "male") || strings.EqualFold(s, "m"):
		return model.GenderMale
	case s == "女" || strings.EqualFold(s, "female") || strings.EqualFold(s, "f"):
		return model.GenderFemale
	default:
		return model.GenderMale
	}
}

func mapRetrieverError(err error) string {
	switch {
	case errors.Is(err, rag.ErrDisabled):
		return ReasonDisabled
	case errors.Is(err, rag.ErrNotConfigured):
		return ReasonNotConfigured
	case errors.Is(err, rag.ErrTimeout):
		return ReasonTimeout
	case errors.Is(err, rag.ErrInvalidResponse):
		return ReasonInvalidResponse
	default:
		return ReasonUpstreamError
	}
}

func buildQuery(result *bazipkg.BaziResult, focus string) string {
	pillars := fmt.Sprintf("%s%s %s%s %s%s %s%s",
		result.YearPillar.Gan, result.YearPillar.Zhi,
		result.MonthPillar.Gan, result.MonthPillar.Zhi,
		result.DayPillar.Gan, result.DayPillar.Zhi,
		result.HourPillar.Gan, result.HourPillar.Zhi,
	)
	dayMaster := result.DayPillar.Gan
	dayElement := data.GanElement[dayMaster]
	tenGods := formatTenGodMap(result.TenGods)
	pattern := result.PatternAnalysis.PatternName
	if pattern == "" {
		pattern = "正格"
	}
	tiaohou := ""
	if result.Tiaohou != nil {
		tiaohou = fmt.Sprintf("%s生%s月 调候用神%s %s", result.Tiaohou.Stem, result.Tiaohou.Month, result.Tiaohou.Primary, result.Tiaohou.Summary)
	}
	dayun := formatDayun(result.DaYunInfo)
	relations := formatRelations(result.GanZhiAnalysis)

	parts := []string{
		"八字经典依据检索",
		"focus=" + focus,
		"四柱=" + pillars,
		"日主=" + dayMaster + dayElement,
		"月令=" + result.MonthPillar.Zhi,
		"十神=" + tenGods,
		"格局=" + pattern + " " + result.PatternAnalysis.PatternType,
		"身旺=" + result.BodyStrength.Verdict,
		"调候=" + tiaohou,
		"刑冲合害=" + relations,
		"大运=" + dayun,
	}

	switch focus {
	case "pattern":
		parts = append(parts, "重点检索格局、月令提纲、成格破格、喜忌")
	case "tiaohou":
		parts = append(parts, "重点检索调候、寒暖燥湿、用神取法")
	case "ten_gods":
		parts = append(parts, "重点检索十神、透干藏干、十神组合")
	default:
		parts = append(parts, "重点检索综合断法、月令、格局、调候、十神")
	}
	return strings.Join(parts, "\n")
}

func BuildQueryForTest(result *bazipkg.BaziResult, focus string) string {
	return buildQuery(result, normalizeFocus(focus))
}

func FilterChunksForTest(chunks []rag.RetrievedChunk, minScore float64, topK int) []rag.RetrievedChunk {
	return filterChunks(chunks, minScore, topK)
}

func formatTenGodMap(m map[string]string) string {
	order := []string{"year", "month", "day", "hour"}
	labels := map[string]string{"year": "年", "month": "月", "day": "日", "hour": "时"}
	parts := make([]string, 0, len(m))
	for _, k := range order {
		if v := strings.TrimSpace(m[k]); v != "" {
			parts = append(parts, labels[k]+":"+v)
		}
	}
	return strings.Join(parts, " ")
}

func formatDayun(d bazipkg.DaYunInfo) string {
	if len(d.Pillars) == 0 {
		return ""
	}
	limit := len(d.Pillars)
	if limit > 4 {
		limit = 4
	}
	parts := make([]string, 0, limit+1)
	parts = append(parts, fmt.Sprintf("%d岁%s", d.StartAge, d.Direction))
	for i := 0; i < limit; i++ {
		parts = append(parts, d.Pillars[i].Gan+d.Pillars[i].Zhi)
	}
	return strings.Join(parts, " ")
}

func formatRelations(g bazipkg.GanZhiAnalysis) string {
	parts := []string{}
	for _, rel := range g.GanRelations {
		parts = append(parts, rel.Type+":"+rel.Pillar1+"-"+rel.Pillar2)
	}
	for _, rel := range g.ZhiRelations {
		parts = append(parts, rel.Type+":"+rel.Pillar1+"-"+rel.Pillar2)
	}
	return strings.Join(parts, " ")
}

func filterChunks(chunks []rag.RetrievedChunk, minScore float64, topK int) []rag.RetrievedChunk {
	seen := map[string]bool{}
	out := make([]rag.RetrievedChunk, 0, len(chunks))
	for _, chunk := range chunks {
		content := cleanContent(chunk.Content)
		if content == "" || chunk.Score < minScore || isIndexChunk(chunk, content) {
			continue
		}
		key := citationDedupeKey(chunk, content)
		if key == "" {
			key = chunk.ID
		}
		if key == "" {
			key = chunk.DocumentID + ":" + firstRunes(content, 80)
		}
		if seen[key] {
			continue
		}
		seen[key] = true
		chunk.Content = content
		out = append(out, chunk)
		if len(out) >= topK {
			break
		}
	}
	return out
}

func citationDedupeKey(chunk rag.RetrievedChunk, content string) string {
	path := firstNonEmpty(chunk.Metadata["source_path"], chunk.Metadata["path"], chunk.DocumentID)
	if path == "" {
		return ""
	}
	return path
}

func isIndexChunk(chunk rag.RetrievedChunk, content string) bool {
	if strings.EqualFold(chunk.Metadata["is_index"], "true") {
		return true
	}
	path := strings.TrimSpace(chunk.Metadata["source_path"])
	if path == "" {
		path = strings.TrimSpace(chunk.Metadata["path"])
	}
	if strings.HasSuffix(path, "/000.md") || strings.HasSuffix(path, "\\000.md") {
		return true
	}
	trimmed := strings.TrimSpace(content)
	if strings.Count(trimmed, "\n") > 10 && strings.Count(trimmed, "](") > 5 {
		return true
	}
	return false
}

func buildCitations(chunks []rag.RetrievedChunk) []model.InterpretationCitation {
	out := make([]model.InterpretationCitation, 0, len(chunks))
	for i, chunk := range chunks {
		meta := chunk.Metadata
		path := firstNonEmpty(meta["source_path"], meta["path"])
		out = append(out, model.InterpretationCitation{
			ID:      i + 1,
			Book:    firstNonEmpty(meta["book"], "未标注典籍"),
			Chapter: firstNonEmpty(meta["chapter"], meta["title"]),
			Path:    path,
			Quote:   firstRunes(cleanContent(chunk.Content), 120),
			Score:   chunk.Score,
		})
	}
	return out
}

func mergeSectionChunks(sections []sectionChunks, topK int) []rag.RetrievedChunk {
	if topK <= 0 {
		topK = 8
	}
	seen := map[string]bool{}
	out := []rag.RetrievedChunk{}
	for _, section := range sections {
		for _, chunk := range section.Chunks {
			key := citationDedupeKey(chunk, chunk.Content)
			if key == "" {
				key = chunk.ID
			}
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, chunk)
			if len(out) >= topK {
				return out
			}
		}
	}
	return out
}

func buildEvidenceSections(result *bazipkg.BaziResult, focus string, evidence []sectionChunks, citations []model.InterpretationCitation) []model.InterpretationSection {
	citationIDs := citationIDsByPath(citations)
	byFocus := map[string]sectionChunks{}
	for _, section := range evidence {
		byFocus[section.Focus] = section
	}

	build := func(sectionFocus string) model.InterpretationSection {
		sectionEvidence := byFocus[sectionFocus]
		points := extractEvidencePoints(sectionEvidence.Chunks, citationIDs, evidenceKeywords(result, sectionFocus), result, 3)
		ids := pointCitationIDs(points)
		if len(ids) == 0 {
			ids = firstCitationIDs(citations, 2)
		}
		return model.InterpretationSection{
			Title:     sectionTitle(sectionFocus),
			Content:   buildEvidenceContent(result, sectionFocus, points),
			Citations: ids,
		}
	}

	if focus != "overview" {
		return []model.InterpretationSection{build(focus)}
	}
	return []model.InterpretationSection{
		build("pattern"),
		build("tiaohou"),
		build("ten_gods"),
	}
}

func citationIDsByPath(citations []model.InterpretationCitation) map[string]int {
	out := map[string]int{}
	for _, citation := range citations {
		if citation.Path != "" {
			out[citation.Path] = citation.ID
		}
	}
	return out
}

func extractEvidencePoints(chunks []rag.RetrievedChunk, citationIDs map[string]int, keywords []string, result *bazipkg.BaziResult, limit int) []evidencePoint {
	if limit <= 0 {
		limit = 3
	}
	candidates := []evidencePoint{}
	seen := map[string]bool{}
	for _, chunk := range chunks {
		path := firstNonEmpty(chunk.Metadata["source_path"], chunk.Metadata["path"])
		citationID := citationIDs[path]
		if citationID == 0 {
			continue
		}
		for _, sentence := range splitEvidenceSentences(chunk.Content) {
			if !sentenceMatches(sentence, keywords) {
				continue
			}
			if isOtherDayMasterSentence(sentence, result) {
				continue
			}
			text := firstRunes(sentence, 90)
			if text == "" || seen[text] {
				continue
			}
			seen[text] = true
			candidates = append(candidates, evidencePoint{Text: text, CitationID: citationID})
		}
	}
	sortEvidencePoints(candidates, keywords)
	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates
}

func isOtherDayMasterSentence(sentence string, result *bazipkg.BaziResult) bool {
	if result == nil {
		return false
	}
	dayGan := result.DayPillar.Gan
	patterns := []string{
		"甲", "乙", "丙", "丁", "戊", "己", "庚", "辛", "壬", "癸",
	}
	for _, gan := range patterns {
		if gan == dayGan {
			continue
		}
		if strings.Contains(sentence, gan+"生"+result.MonthPillar.Zhi) ||
			strings.Contains(sentence, gan+"日") ||
			strings.Contains(sentence, gan+"水") ||
			strings.Contains(sentence, gan+"木") ||
			strings.Contains(sentence, gan+"火") ||
			strings.Contains(sentence, gan+"土") ||
			strings.Contains(sentence, gan+"金") {
			return true
		}
	}
	return false
}

func sortEvidencePoints(points []evidencePoint, keywords []string) {
	for i := 0; i < len(points); i++ {
		for j := i + 1; j < len(points); j++ {
			if evidencePointScore(points[j], keywords) > evidencePointScore(points[i], keywords) {
				points[i], points[j] = points[j], points[i]
			}
		}
	}
}

func evidencePointScore(point evidencePoint, keywords []string) int {
	score := 0
	for i, keyword := range keywords {
		if keyword == "" || !strings.Contains(point.Text, keyword) {
			continue
		}
		score += 10
		if i < 10 {
			score += 10 - i
		}
		if utf8.RuneCountInString(keyword) >= 3 {
			score += 4
		}
	}
	return score
}

func splitEvidenceSentences(content string) []string {
	content = cleanContent(content)
	replacer := strings.NewReplacer("。", "。\n", "；", "；\n", "！", "！\n", "？", "？\n")
	content = replacer.Replace(content)
	lines := strings.Split(content, "\n")
	out := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(strings.Trim(line, "-*># "))
		if line == "" || strings.HasPrefix(line, "书籍：") || strings.HasPrefix(line, "原文") || strings.HasPrefix(line, "白话译文") {
			continue
		}
		if utf8.RuneCountInString(line) < 8 {
			continue
		}
		out = append(out, line)
	}
	return out
}

func sentenceMatches(sentence string, keywords []string) bool {
	if len(keywords) == 0 {
		return true
	}
	for _, keyword := range keywords {
		keyword = strings.TrimSpace(keyword)
		if keyword != "" && strings.Contains(sentence, keyword) {
			return true
		}
	}
	return false
}

func pointCitationIDs(points []evidencePoint) []int {
	seen := map[int]bool{}
	out := []int{}
	for _, point := range points {
		if point.CitationID == 0 || seen[point.CitationID] {
			continue
		}
		seen[point.CitationID] = true
		out = append(out, point.CitationID)
	}
	return out
}

func firstCitationIDs(citations []model.InterpretationCitation, max int) []int {
	out := []int{}
	for _, citation := range citations {
		out = append(out, citation.ID)
		if len(out) >= max {
			break
		}
	}
	return out
}

func evidenceKeywords(result *bazipkg.BaziResult, focus string) []string {
	dayMaster := result.DayPillar.Gan
	dayStem := dayMaster + data.GanElement[dayMaster]
	pattern := firstNonEmpty(result.PatternAnalysis.PatternName, result.PatternAnalysis.PatternType)
	month := result.MonthPillar.Zhi
	keywords := []string{dayMaster, dayStem, month + "月"}
	switch focus {
	case "pattern":
		keywords = append(keywords, pattern, "月令", "格局", "取格", "羊刃", "阳刃", "日刃", "刃神", "七煞相制", "官", "杀", "食伤", "喜", "忌", result.DayPillar.Gan+result.DayPillar.Zhi)
	case "tiaohou":
		keywords = append(keywords, "调候", "用神", "寒", "暖", "燥", "湿", "丙", "癸", "辰月")
		if result.Tiaohou != nil {
			keywords = append(keywords, result.Tiaohou.Primary)
			keywords = append(keywords, result.Tiaohou.Reasons...)
		}
	case "ten_gods":
		keywords = append(keywords, "十神", "劫财", "比肩", "正官", "七杀", "食神", "伤官", "印", "财")
		for _, ratio := range topTenGodRatios(result.TenGodProportion, 3) {
			keywords = append(keywords, ratio.Name)
		}
	default:
		keywords = append(keywords, "月令", "格局", "调候", "十神")
	}
	return uniqueStrings(keywords)
}

func buildEvidenceContent(result *bazipkg.BaziResult, focus string, points []evidencePoint) string {
	switch focus {
	case "pattern":
		return buildPatternEvidenceContent(result, points)
	case "tiaohou":
		return buildTiaohouEvidenceContent(result, points)
	case "ten_gods":
		return buildTenGodEvidenceContent(result, points)
	default:
		return buildPatternEvidenceContent(result, points)
	}
}

func buildPatternEvidenceContent(result *bazipkg.BaziResult, points []evidencePoint) string {
	p := result.PatternAnalysis
	pattern := firstNonEmpty(p.PatternName, p.PatternType, "当前格局")
	dayStem := result.DayPillar.Gan + data.GanElement[result.DayPillar.Gan]
	pillars := pillarText(result)
	evidenceText := formatEvidencePoints(points)
	explanation := []string{}
	explanation = append(explanation, patternOpening(result, pattern, dayStem, pillars, p.Description))
	if evidenceText != "" {
		explanation = append(explanation, evidenceLead(result, "pattern")+evidenceText)
	}
	dayElem := data.GanElement[result.DayPillar.Gan]
	explanation = append(explanation, patternUsefulness(result, pattern, dayElem, p))
	explanation = append(explanation, patternConclusion(result, pattern))
	if strings.Contains(pattern, "羊刃") {
		explanation = append(explanation, yangRenStyleAdvice(result))
	}
	return strings.Join(explanation, "\n\n")
}

func patternOpening(result *bazipkg.BaziResult, pattern, dayStem, pillars, description string) string {
	desc := sentenceBody(firstNonEmpty(description, "格局要看月令、透干、根气与制化，不宜只看一个格名。"))
	if strings.Contains(pattern, "羊刃") {
		switch chartStyleIndex(result, "pattern-yangren", 3) {
		case 0:
			return fmt.Sprintf("这个盘不是柔和一路。%s，%s坐%s，月令%s，日支带刃，气先立起来了。%s。", pillars, dayStem, result.DayPillar.Zhi, result.MonthPillar.Zhi, desc)
		case 1:
			return fmt.Sprintf("这局先别急着论吉凶，先看一个“刃”字怎么安放。%s，%s坐%s，月令在%s，气势不虚。%s。", pillars, dayStem, result.DayPillar.Zhi, result.MonthPillar.Zhi, desc)
		default:
			return fmt.Sprintf("此造的关窍不在格名好不好听，而在刚气能不能成器。%s，日主%s，日支%s，月令%s，刃气已经见形。%s。", pillars, dayStem, result.DayPillar.Zhi, result.MonthPillar.Zhi, desc)
		}
	}
	if strings.Contains(pattern, "七杀") || strings.Contains(pattern, "偏官") {
		if chartStyleIndex(result, "pattern-kill", 2) == 0 {
			return fmt.Sprintf("这盘先看杀气有没有成器。%s，%s临%s月，%s。七杀不怕见，怕的是无制无化；制得住，是胆识和执行，制不住，就是压力和冲动。", pillars, dayStem, result.MonthPillar.Zhi, desc)
		}
		return fmt.Sprintf("此局见杀，不能只当压力看。%s，%s生%s月，%s。杀有制化，反作担当与权柄；杀无去处，才成逼身之患。", pillars, dayStem, result.MonthPillar.Zhi, desc)
	}
	if strings.Contains(pattern, "正官") {
		if chartStyleIndex(result, "pattern-official", 2) == 0 {
			return fmt.Sprintf("这个盘要先看规矩和承载。%s，%s生在%s月，%s。官星一路看的是秩序、名分和责任，不是只看有没有官字。", pillars, dayStem, result.MonthPillar.Zhi, desc)
		}
		return fmt.Sprintf("此局官星的味道，要从月令和承载力上看。%s，日主%s，月令%s，%s。官不是一句“有职位”就完了，关键是清、稳、能否为我所任。", pillars, dayStem, result.MonthPillar.Zhi, desc)
	}
	if isStrong(result.BodyStrength.Verdict) {
		if chartStyleIndex(result, "pattern-strong", 2) == 0 {
			return fmt.Sprintf("这盘底气不薄。%s，日主%s，月令%s，%s。身旺的盘，最怕只加力不疏通，关键是让旺气有用处。", pillars, dayStem, result.MonthPillar.Zhi, desc)
		}
		return fmt.Sprintf("这个命局先看“气从哪里来，又往哪里去”。%s，日主%s，月令%s，%s。既然日主有力，后面就不再问够不够强，而问能不能泄、能不能制、能不能成事。", pillars, dayStem, result.MonthPillar.Zhi, desc)
	}
	if chartStyleIndex(result, "pattern-weak", 2) == 0 {
		return fmt.Sprintf("这盘要先看日主能不能接住格局。%s，日主%s，月令%s，%s。身不够时，喜忌不是摆设，先要有根气和帮扶，再谈发挥。", pillars, dayStem, result.MonthPillar.Zhi, desc)
	}
	return fmt.Sprintf("此局不能一上来就谈财官名利，要先问日主有没有承载。%s，%s生%s月，%s。根气接得住，格局才有用；接不住，好处也容易变成压力。", pillars, dayStem, result.MonthPillar.Zhi, desc)
}

func evidenceLead(result *bazipkg.BaziResult, focus string) string {
	switch focus {
	case "pattern":
		if strings.Contains(result.PatternAnalysis.PatternName, "羊刃") {
			return pickText(result, "pattern-evidence", []string{
				"能借得上的古法，重点都落在“刃要制化”：",
				"古书讲刃，多半不是叫人怕它，而是看谁来驾驭它：",
				"这一段可借的经典意思，核心都在制刃、化刃：",
			})
		}
		return pickText(result, "pattern-evidence", []string{
			"翻到典籍里，和这一路相近的说法是：",
			"古法能用上的地方，在这几句里：",
			"这里不硬套条文，只取和盘面贴得上的几句：",
		})
	case "tiaohou":
		return pickText(result, "tiaohou-evidence", []string{
			"调候上可借的依据是：",
			"讲寒暖燥湿，典籍里这几句可以参看：",
			"这一段先取能落到气候上的条文：",
		})
	case "ten_gods":
		return pickText(result, "tengod-evidence", []string{
			"十神组合可参的句子是：",
			"看人事取象，下面几句比单看比例更有用：",
			"这一段借经典，不是借名词，是借判断顺序：",
		})
	default:
		return "可参考的依据是："
	}
}

func patternUsefulness(result *bazipkg.BaziResult, pattern, dayElem string, p bazipkg.PatternAnalysis) string {
	strong := isStrong(result.BodyStrength.Verdict)
	if strings.Contains(pattern, "羊刃") {
		return fmt.Sprintf("所以这盘不能一味说旺就是好。%s，喜%s，忌%s；%s已经有劲，得让%s来立边界，让%s来出成果，再用%s去引财气。没有制化时，人会很能扛，但也容易扛成硬碰硬。", result.BodyStrength.Verdict, joinOrNone(p.FavorableElements), joinOrNone(p.UnfavorableElements), dayElem, roleElementText(dayElem, "官杀"), roleElementText(dayElem, "食伤"), roleElementText(dayElem, "财星"))
	}
	if strong {
		return fmt.Sprintf("此盘%s，喜%s，忌%s。身旺不缺劲，缺的是方向和出口；%s来能约束，%s来能疏泄，%s来才有可经营的财。", result.BodyStrength.Verdict, joinOrNone(p.FavorableElements), joinOrNone(p.UnfavorableElements), roleElementText(dayElem, "官杀"), roleElementText(dayElem, "食伤"), roleElementText(dayElem, "财星"))
	}
	return fmt.Sprintf("此盘%s，喜%s，忌%s。身偏弱时，先看有没有印比扶住，再看财官食伤能不能为我所用；用得太急，反而成压力。", result.BodyStrength.Verdict, joinOrNone(p.FavorableElements), joinOrNone(p.UnfavorableElements))
}

func patternConclusion(result *bazipkg.BaziResult, pattern string) string {
	if strings.Contains(pattern, "羊刃") {
		return "这一格最怕“有刃无制”：人有锋芒，但锋芒若无人收束，就容易变成急躁、争执和财来财去。若行运见官杀、食伤得力，反而能把这股冲劲变成职位、作品、项目成果。"
	}
	if strings.Contains(pattern, "七杀") || strings.Contains(pattern, "偏官") {
		return "这一路重在制杀化杀。能制，压力就是权柄；不能制，机会也会带着风险。看后运时，宜看印、食伤、合制是否接得上。"
	}
	if strings.Contains(pattern, "正官") {
		return "正官一路讲清正和稳定。最怕伤官冲破、官杀混杂；最喜财印相扶，做事有章法，名分和责任自然能立起来。"
	}
	if isStrong(result.BodyStrength.Verdict) {
		return "总的说，这盘不是怕没力，而是怕力气堆在局里不流动。能疏、能制、能转化，层次就出来。"
	}
	return "总的说，这盘先求承载，再求发挥。根气稳了，喜用接上，格局的好处才显。"
}

func yangRenStyleAdvice(result *bazipkg.BaziResult) string {
	if result.TenGods["hour"] == "劫财" || strings.Contains(formatTopTenGods(topTenGodRatios(result.TenGodProportion, 2)), "劫财") {
		return "性情上，多半不喜欢被人牵着走，自己拿主意的劲很强。这个劲用在创业、项目攻坚、技术突破上是好事；用在人际和钱财合作上，就要提前立规矩，否则容易出现“我出了力、别人分了财”的不平。"
	}
	return "日支见刃的人，遇事不太愿意退。这个性子用好了，是敢担当、能拍板；用偏了，就是急、硬、容易和人顶。后运最喜有官杀立边界，或食伤把刚气变成可见结果。"
}

func buildTiaohouEvidenceContent(result *bazipkg.BaziResult, points []evidencePoint) string {
	dayStem := result.DayPillar.Gan + data.GanElement[result.DayPillar.Gan]
	primary := "未定"
	reason := "当前未取得完整调候条目"
	if result.Tiaohou != nil {
		primary = result.Tiaohou.Primary
		reason = strings.Join(result.Tiaohou.Reasons, "；")
		if reason == "" {
			reason = result.Tiaohou.Summary
		}
	}
	evidenceText := formatEvidencePoints(points)
	explanation := []string{tiaohouOpening(result, dayStem, primary, reason)}
	if evidenceText != "" {
		explanation = append(explanation, evidenceLead(result, "tiaohou")+evidenceText)
	}
	explanation = append(explanation, tiaohouUsefulness(result, primary))
	explanation = append(explanation, tiaohouPracticeAdvice(result, primary))
	return strings.Join(explanation, "\n\n")
}

func tiaohouOpening(result *bazipkg.BaziResult, dayStem, primary, reason string) string {
	month := result.MonthPillar.Zhi
	climate := branchClimate(month)
	reason = firstNonEmpty(reason, "以月令寒暖燥湿判断。")
	switch chartStyleIndex(result, "tiaohou-open", 3) {
	case 0:
		return fmt.Sprintf("调候先看气候，不急着断富贵。%s生%s月，%s，日支又见%s，局里那股气要先分清冷暖燥湿。当前取%s，是为了把局中滞住的气调开：%s。", dayStem, month, climate, result.DayPillar.Zhi, primary, reason)
	case 1:
		return fmt.Sprintf("格局像骨架，调候像这副骨架所处的天气。%s在%s月，月令气象是%s；若气候不舒，喜神来了也未必好用。此处首取%s，重点不是堆五行，而是让盘面能运转：%s。", dayStem, month, climate, primary, reason)
	default:
		return fmt.Sprintf("这一盘调候要看得细一点。%s临%s月，%s，不是单说缺什么补什么。取%s，取的是温凉燥湿的平衡，也是让格局能落地的条件：%s。", dayStem, month, climate, primary, reason)
	}
}

func tiaohouUsefulness(result *bazipkg.BaziResult, primary string) string {
	primaryElem := firstStemElement(primary)
	strong := isStrong(result.BodyStrength.Verdict)
	switch primaryElem {
	case "火":
		if strong {
			return fmt.Sprintf("%s在这里像炉火，不是越猛越好。它要暖土、化湿、提精神；但此盘%s，火土若再过，性子会更急，判断会更硬，所以还要有木来立规矩、金来开出口。", primary, result.BodyStrength.Verdict)
		}
		return fmt.Sprintf("%s在这里重在温养。日主若承载不足，先要把寒湿化开，再看印比根气是否接得住；火来得有情，是醒局，不是燥局。", primary)
	case "水":
		return fmt.Sprintf("%s在这里不是泛滥之水，而是润燥、通关、让财气有路。水若有源有去处，事情能流动；若只见水而无堤岸，也容易多想、多拖、多反复。", primary)
	case "木":
		return fmt.Sprintf("%s在这里带生发和约束两层意思。对%s日主来说，木为官杀，来得清，可以把局中的力气收成责任、名分和规则；来得杂，则压力也重。", primary, result.DayPillar.Gan+data.GanElement[result.DayPillar.Gan])
	case "金":
		return fmt.Sprintf("%s在这里重在开口泄秀。土气厚时，金能把闷住的力气化成技术、表达、成果；但金弱无根，只是想法多，未必真能成器。", primary)
	default:
		return fmt.Sprintf("%s不是单独拿来补的字。调候要和格局同看：一边看气候是否舒展，一边看喜忌是否接得上，二者合了，条文才算用到盘里。", primary)
	}
}

func tiaohouPracticeAdvice(result *bazipkg.BaziResult, primary string) string {
	if isStrong(result.BodyStrength.Verdict) {
		return pickText(result, "tiaohou-advice-strong", []string{
			"现实取象上，这类盘常见的问题不是没能力，而是推进方式偏重：想压住场面，也容易把自己压得太紧。把目标拆细、期限定死、成果交出来，局里的气就活。",
			"用在做事上，要少靠硬顶，多靠流程和交付。该立边界时立边界，该交成果时交成果；这样调候的意思才不是纸上谈兵。",
			fmt.Sprintf("所以%s只能算调气的钥匙，不能替代全局喜忌。身旺的盘尤其要防再添火土，越补越重；能有木金来制化疏泄，反而更见层次。", primary),
		})
	}
	return pickText(result, "tiaohou-advice-weak", []string{
		"现实里要先把节奏养稳。身弱或承载不足时，过早追财官，容易看着机会多，实际压力也多；先稳根气，再谈发挥。",
		"这类盘不宜一下子把目标拉太满。先让状态、资源、节奏顺起来，再借喜用发力，反而更容易成。",
		fmt.Sprintf("%s若得地，是把气候调顺；但人事上还要看帮扶是否到位。根气稳，条文里的好处才接得住。", primary),
	})
}

func roleElementText(dayElem, role string) string {
	switch role {
	case "官杀":
		return elementThatControls(dayElem) + "官杀"
	case "食伤":
		return elementProducedBy(dayElem) + "食伤"
	case "财星":
		return elementControlledBy(dayElem) + "财星"
	default:
		return role
	}
}

func elementThatControls(elem string) string {
	switch elem {
	case "木":
		return "金"
	case "火":
		return "水"
	case "土":
		return "木"
	case "金":
		return "火"
	case "水":
		return "土"
	default:
		return ""
	}
}

func elementProducedBy(elem string) string {
	switch elem {
	case "木":
		return "火"
	case "火":
		return "土"
	case "土":
		return "金"
	case "金":
		return "水"
	case "水":
		return "木"
	default:
		return ""
	}
}

func elementControlledBy(elem string) string {
	switch elem {
	case "木":
		return "土"
	case "火":
		return "金"
	case "土":
		return "水"
	case "金":
		return "木"
	case "水":
		return "火"
	default:
		return ""
	}
}

func isStrong(verdict string) bool {
	verdict = strings.TrimSpace(verdict)
	if verdict == "" || strings.Contains(verdict, "弱") || strings.Contains(verdict, "衰") {
		return false
	}
	return strings.Contains(verdict, "旺") || strings.Contains(verdict, "强")
}

func chartStyleIndex(result *bazipkg.BaziResult, salt string, size int) int {
	if size <= 1 || result == nil {
		return 0
	}
	key := pillarText(result) + "|" + result.BodyStrength.Verdict + "|" + result.PatternAnalysis.PatternName + "|" + salt
	sum := 0
	for i, r := range key {
		sum += int(r) * (i + 1)
	}
	if sum < 0 {
		sum = -sum
	}
	return sum % size
}

func pickText(result *bazipkg.BaziResult, salt string, options []string) string {
	if len(options) == 0 {
		return ""
	}
	return options[chartStyleIndex(result, salt, len(options))]
}

func branchClimate(branch string) string {
	switch branch {
	case "寅":
		return "初春木气发动，余寒未尽"
	case "卯":
		return "仲春木旺，生发之气足"
	case "辰":
		return "季春湿土，木余而土湿"
	case "巳":
		return "初夏火起，燥热渐生"
	case "午":
		return "仲夏火旺，炎热最盛"
	case "未":
		return "季夏土燥夹暑"
	case "申":
		return "初秋金气初起，暑湿未尽"
	case "酉":
		return "仲秋金旺，肃杀偏燥"
	case "戌":
		return "季秋燥土，火余入墓"
	case "亥":
		return "初冬水旺，寒气渐重"
	case "子":
		return "仲冬水旺，寒冷最重"
	case "丑":
		return "季冬湿寒之土"
	default:
		return "月令气候需合全局细看"
	}
}

func firstStemElement(stems string) string {
	for _, r := range stems {
		if elem := data.GanElement[string(r)]; elem != "" {
			return elem
		}
	}
	return ""
}

func buildTenGodEvidenceContent(result *bazipkg.BaziResult, points []evidencePoint) string {
	top := topTenGodRatios(result.TenGodProportion, 4)
	topText := formatTopTenGods(top)
	evidenceText := formatEvidencePoints(points)
	dominant := dominantTenGod(top)
	explanation := []string{tenGodOpening(result, topText, dominant)}
	if evidenceText != "" {
		explanation = append(explanation, evidenceLead(result, "ten_gods")+evidenceText)
	}
	explanation = append(explanation, tenGodUsefulness(result, dominant))
	explanation = append(explanation, tenGodFlowAdvice(result, dominant))
	if result.TenGodAnalysis != nil && result.TenGodAnalysis.Summary != "" {
		explanation = append(explanation, pickText(result, "tengod-summary", []string{
			"规则层面的提醒也留一条：" + result.TenGodAnalysis.Summary,
			"再合程序的十神分析看：" + result.TenGodAnalysis.Summary,
			"这和当前十神规则的结论能对上：" + result.TenGodAnalysis.Summary,
		}))
	}
	return strings.Join(explanation, "\n\n")
}

func tenGodOpening(result *bazipkg.BaziResult, topText, dominant string) string {
	stem := result.DayPillar.Gan + data.GanElement[result.DayPillar.Gan]
	switch tenGodGroup(dominant) {
	case "peer":
		return pickText(result, "tengod-open-peer", []string{
			fmt.Sprintf("十神先看谁在局里最有声音。此盘较显的是%s，日主%s又见%s；这不是简单说“朋友多”或“竞争多”，而是自我、资源、分夺和担当都要一起看。四柱天干十神为%s。", topText, stem, result.BodyStrength.Verdict, formatTenGodMap(result.TenGods)),
			fmt.Sprintf("这盘十神的气，先从比劫一路入手。%s最露，日主%s，说明命主做事不太愿意完全借别人手，凡事想自己掌控。四柱天干十神为%s。", topText, stem, formatTenGodMap(result.TenGods)),
			fmt.Sprintf("若只看比例，会说%s；但师傅看盘还要问它落在哪里、是不是帮身过头。日主%s，身势为%s，四柱天干十神为%s。", topText, stem, result.BodyStrength.Verdict, formatTenGodMap(result.TenGods)),
		})
	case "output":
		return pickText(result, "tengod-open-output", []string{
			fmt.Sprintf("这盘十神要看输出之气。较显的是%s，日主%s，才华、表达、技术、作品都在这一层里看。四柱天干十神为%s。", topText, stem, formatTenGodMap(result.TenGods)),
			fmt.Sprintf("食伤一路明显时，不能只说聪明，要看能不能变成成果。此盘较显%s，日主%s，四柱天干十神为%s。", topText, stem, formatTenGodMap(result.TenGods)),
		})
	case "wealth":
		return pickText(result, "tengod-open-wealth", []string{
			fmt.Sprintf("财星明显的盘，先看财有没有源、日主接不接得住。此盘较显%s，日主%s，四柱天干十神为%s。", topText, stem, formatTenGodMap(result.TenGods)),
			fmt.Sprintf("十神这里要看资源和经营。%s较显，日主%s，财不是只代表钱，也代表可调动的人事与机会。四柱天干十神为%s。", topText, stem, formatTenGodMap(result.TenGods)),
		})
	case "authority":
		return pickText(result, "tengod-open-authority", []string{
			fmt.Sprintf("官杀显时，先看压力能不能化成权责。此盘较显%s，日主%s，四柱天干十神为%s。", topText, stem, formatTenGodMap(result.TenGods)),
			fmt.Sprintf("这段看职位、规则、约束与承担。%s较显，日主%s，若能成体系，就是责任；失衡时就是压力。四柱天干十神为%s。", topText, stem, formatTenGodMap(result.TenGods)),
		})
	case "seal":
		return pickText(result, "tengod-open-seal", []string{
			fmt.Sprintf("印星明显，要看学识、凭借、保护，也要防想得多、动得慢。此盘较显%s，日主%s，四柱天干十神为%s。", topText, stem, formatTenGodMap(result.TenGods)),
			fmt.Sprintf("这盘十神先看印的承托。%s较显，日主%s，印能生身，也能让人依赖旧经验。四柱天干十神为%s。", topText, stem, formatTenGodMap(result.TenGods)),
		})
	default:
		return fmt.Sprintf("十神这段，不是看哪个百分比最大就完事，要看它在什么位置、能不能成事。此盘较显的是%s；四柱天干十神为%s。", topText, formatTenGodMap(result.TenGods))
	}
}

func tenGodUsefulness(result *bazipkg.BaziResult, dominant string) string {
	stem := result.DayPillar.Gan + data.GanElement[result.DayPillar.Gan]
	switch tenGodGroup(dominant) {
	case "peer":
		if isStrong(result.BodyStrength.Verdict) {
			return fmt.Sprintf("落到人事上，%s又%s，比劫重的一面是能扛、敢争、资源意识强；不好的一面是容易把事情抓在自己手里，钱财合作上也要防“分夺”之象。账目、权责、边界先说清，比临时讲情面更稳。", stem, result.BodyStrength.Verdict)
		}
		return fmt.Sprintf("比劫若为帮身，先看是不是帮得上。%s若根气不足，有同类来扶是好事；但扶过头，也会变成争执、合伙不清和资源消耗。", stem)
	case "output":
		return fmt.Sprintf("食伤的好处，是把命局里的力气讲出来、做出来、交付出来。对%s而言，输出若清，适合技术、内容、产品、方案；输出若杂，就容易嘴快、心急、和规则顶着来。", stem)
	case "wealth":
		return fmt.Sprintf("财星重，不等于财一定厚。还要看%s接不接得住、财有没有源、有没有官印护住。接得住，是经营和资源；接不住，就是机会多、消耗也多。", stem)
	case "authority":
		return fmt.Sprintf("官杀不是单纯的管束，它也代表职位、责任、规则和风险。%s若能任官杀，压力会变成身份；若任不住，就容易被制度、上级、项目节点推着走。", stem)
	case "seal":
		return "印星重，长处在学习、资格、贵人和系统性；短处是容易停在想、等、准备。若原局已经偏旺，印再多就不宜恋旧法，要借食伤把东西拿出来。"
	default:
		return fmt.Sprintf("落到人事上，要把%s和%s合看。十神只是角色，月令和身强身弱才决定这些角色是帮我，还是耗我、压我。", stem, result.BodyStrength.Verdict)
	}
}

func tenGodFlowAdvice(result *bazipkg.BaziResult, dominant string) string {
	dayElem := data.GanElement[result.DayPillar.Gan]
	if isStrong(result.BodyStrength.Verdict) {
		return pickText(result, "tengod-flow-strong", []string{
			fmt.Sprintf("%s若成体系，就像给这股力套上规矩；%s若得用，就把硬气转成技术、表达、产品和输出；%s有路，才谈钱财流通。最怕只剩印比助身，人很能撑，但转化率不高。", roleElementText(dayElem, "官杀"), roleElementText(dayElem, "食伤"), roleElementText(dayElem, "财星")),
			fmt.Sprintf("这个盘后面看运，最要紧是看%s、%s有没有接力。来得清，就是职位、作品、项目成果；来得混，就会变成忙、累、硬顶，钱财还容易被人事牵走。", roleElementText(dayElem, "官杀"), roleElementText(dayElem, "食伤")),
			fmt.Sprintf("所以断十神不能只说性格，要落到用法：%s立边界，%s给出口，%s管流通。三者接上，原局的旺气才不只是脾气，而能成事。", roleElementText(dayElem, "官杀"), roleElementText(dayElem, "食伤"), roleElementText(dayElem, "财星")),
		})
	}
	return pickText(result, "tengod-flow-weak", []string{
		"若日主承载不足，十神越热闹，越要先分清哪些是助力，哪些是压力。先稳印比根气，再谈财官食伤的发挥。",
		"这类盘看运，不怕机会少，怕机会来得太急。帮身运先把底盘稳住，再遇财官食伤，反而容易接得住。",
		fmt.Sprintf("十神要回到日主能不能任事。%s若先稳住，财官食伤才有用；身弱而急追外物，多半先见压力。", result.DayPillar.Gan+data.GanElement[result.DayPillar.Gan]),
	})
}

func dominantTenGod(ratios []bazipkg.TenGodRatio) string {
	if len(ratios) == 0 {
		return ""
	}
	return ratios[0].Name
}

func tenGodGroup(god string) string {
	switch god {
	case "比肩", "劫财":
		return "peer"
	case "食神", "伤官":
		return "output"
	case "正财", "偏财":
		return "wealth"
	case "正官", "七杀":
		return "authority"
	case "正印", "偏印":
		return "seal"
	default:
		return ""
	}
}

func formatEvidencePoints(points []evidencePoint) string {
	parts := []string{}
	for _, point := range points {
		if point.Text == "" || point.CitationID == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("“%s”[%d]", point.Text, point.CitationID))
	}
	return strings.Join(parts, "；")
}

func pillarText(result *bazipkg.BaziResult) string {
	return fmt.Sprintf("%s%s、%s%s、%s%s、%s%s",
		result.YearPillar.Gan, result.YearPillar.Zhi,
		result.MonthPillar.Gan, result.MonthPillar.Zhi,
		result.DayPillar.Gan, result.DayPillar.Zhi,
		result.HourPillar.Gan, result.HourPillar.Zhi,
	)
}

func joinOrNone(values []string) string {
	if len(values) == 0 {
		return "未明"
	}
	return strings.Join(values, "、")
}

func topTenGodRatios(ratios []bazipkg.TenGodRatio, limit int) []bazipkg.TenGodRatio {
	out := append([]bazipkg.TenGodRatio(nil), ratios...)
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Percent > out[i].Percent {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func formatTopTenGods(ratios []bazipkg.TenGodRatio) string {
	if len(ratios) == 0 {
		return "未取得十神比例"
	}
	parts := make([]string, 0, len(ratios))
	for _, ratio := range ratios {
		if ratio.Percent <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s%.1f%%", ratio.Name, ratio.Percent))
	}
	if len(parts) == 0 {
		return "未取得十神比例"
	}
	return strings.Join(parts, "、")
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}

func buildSummary(result *bazipkg.BaziResult, focus string, withCitations bool) string {
	source := "基于当前命盘规则"
	if withCitations {
		source = "结合当前命盘规则与检索到的经典条文"
	}
	dayMaster := result.DayPillar.Gan + data.GanElement[result.DayPillar.Gan]
	pattern := result.PatternAnalysis.PatternName
	if pattern == "" {
		pattern = result.PatternAnalysis.PatternType
	}
	if pattern == "" {
		pattern = "正格"
	}

	switch focus {
	case "pattern":
		return fmt.Sprintf("%s，重点看月令%s与%s。此盘日主为%s，格局取向为%s，喜忌仍以命局全局平衡为准。", source, result.MonthPillar.Zhi, pattern, dayMaster, pattern)
	case "tiaohou":
		if result.Tiaohou != nil {
			return fmt.Sprintf("%s，重点看%s生%s月的寒暖燥湿。当前调候首取%s，需与格局喜忌合参。", source, result.Tiaohou.Stem, result.Tiaohou.Month, result.Tiaohou.Primary)
		}
		return fmt.Sprintf("%s，重点看日主%s与月令%s的寒暖燥湿，当前未取得完整调候规则。", source, dayMaster, result.MonthPillar.Zhi)
	case "ten_gods":
		return fmt.Sprintf("%s，重点看十神透藏与强弱。此盘日主%s，十神分布为%s，宜结合月令和身旺结论判断。", source, dayMaster, formatTenGodMap(result.TenGods))
	default:
		return fmt.Sprintf("%s，此盘四柱为%s%s、%s%s、%s%s、%s%s；日主%s，月令%s，格局为%s，身旺结论为%s。",
			source,
			result.YearPillar.Gan, result.YearPillar.Zhi,
			result.MonthPillar.Gan, result.MonthPillar.Zhi,
			result.DayPillar.Gan, result.DayPillar.Zhi,
			result.HourPillar.Gan, result.HourPillar.Zhi,
			dayMaster,
			result.MonthPillar.Zhi,
			pattern,
			result.BodyStrength.Verdict,
		)
	}
}

func buildSections(result *bazipkg.BaziResult, focus string, citations []model.InterpretationCitation) []model.InterpretationSection {
	citationIDs := make([]int, 0, len(citations))
	for _, c := range citations {
		citationIDs = append(citationIDs, c.ID)
	}
	ids := func(max int) []int {
		if len(citationIDs) <= max {
			return citationIDs
		}
		return citationIDs[:max]
	}

	switch focus {
	case "pattern":
		return []model.InterpretationSection{{
			Title:     "格局与月令",
			Content:   buildPatternContent(result),
			Citations: ids(3),
		}}
	case "tiaohou":
		return []model.InterpretationSection{{
			Title:     "调候用神",
			Content:   buildTiaohouContent(result),
			Citations: ids(3),
		}}
	case "ten_gods":
		return []model.InterpretationSection{{
			Title:     "十神透藏",
			Content:   buildTenGodContent(result),
			Citations: ids(3),
		}}
	default:
		return []model.InterpretationSection{
			{Title: "格局与月令", Content: buildPatternContent(result), Citations: ids(2)},
			{Title: "调候与平衡", Content: buildTiaohouContent(result), Citations: ids(4)},
			{Title: "十神结构", Content: buildTenGodContent(result), Citations: ids(5)},
		}
	}
}

func buildPatternContent(result *bazipkg.BaziResult) string {
	p := result.PatternAnalysis
	if p.PatternName == "" && p.Description == "" {
		return fmt.Sprintf("月令为%s，日主为%s，当前按普通格局与身旺喜忌综合判断。", result.MonthPillar.Zhi, result.DayPillar.Gan)
	}
	return fmt.Sprintf("%s属%s。%s 喜%s，忌%s。",
		firstNonEmpty(p.PatternName, "当前格局"),
		firstNonEmpty(p.PatternType, "格局判断"),
		firstNonEmpty(p.Description, "需结合月令、透干与全局五行流通判断。"),
		strings.Join(p.FavorableElements, "、"),
		strings.Join(p.UnfavorableElements, "、"),
	)
}

func buildTiaohouContent(result *bazipkg.BaziResult) string {
	if result.Tiaohou == nil {
		return firstNonEmpty(result.DayStemTiaoHou, "当前未取得调候条目，先以月令寒暖燥湿与五行平衡合参。")
	}
	return result.Tiaohou.Summary
}

func buildTenGodContent(result *bazipkg.BaziResult) string {
	if result.TenGodAnalysis != nil && result.TenGodAnalysis.Summary != "" {
		return result.TenGodAnalysis.Summary
	}
	return fmt.Sprintf("十神分布：%s。当前日主%s，需结合月令%s与身旺结论%s合参。",
		formatTenGodMap(result.TenGods),
		result.DayPillar.Gan,
		result.MonthPillar.Zhi,
		result.BodyStrength.Verdict,
	)
}

func cleanContent(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}
	return strings.Join(cleaned, "\n")
}

func firstRunes(s string, max int) string {
	s = strings.TrimSpace(s)
	if max <= 0 || utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return strings.TrimSpace(string(runes[:max])) + "..."
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func sentenceBody(s string) string {
	return strings.TrimRight(strings.TrimSpace(s), "。.!！?？；;")
}
