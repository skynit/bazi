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
	desc := firstNonEmpty(description, "格局要看月令、透干、根气与制化，不宜只看一个格名。")
	if strings.Contains(pattern, "羊刃") {
		return fmt.Sprintf("这个盘不是柔和一路。%s，%s坐%s，月令%s，日支带刃，气先立起来了。%s", pillars, dayStem, result.DayPillar.Zhi, result.MonthPillar.Zhi, desc)
	}
	if strings.Contains(pattern, "七杀") || strings.Contains(pattern, "偏官") {
		return fmt.Sprintf("这盘先看杀气有没有成器。%s，%s临%s月，%s。七杀不怕见，怕的是无制无化；制得住，是胆识和执行，制不住，就是压力和冲动。", pillars, dayStem, result.MonthPillar.Zhi, desc)
	}
	if strings.Contains(pattern, "正官") {
		return fmt.Sprintf("这个盘要先看规矩和承载。%s，%s生在%s月，%s。官星一路看的是秩序、名分和责任，不是只看有没有官字。", pillars, dayStem, result.MonthPillar.Zhi, desc)
	}
	if isStrong(result.BodyStrength.Verdict) {
		return fmt.Sprintf("这盘底气不薄。%s，日主%s，月令%s，%s。身旺的盘，最怕只加力不疏通，关键是让旺气有用处。", pillars, dayStem, result.MonthPillar.Zhi, desc)
	}
	return fmt.Sprintf("这盘要先看日主能不能接住格局。%s，日主%s，月令%s，%s。身不够时，喜忌不是摆设，先要有根气和帮扶，再谈发挥。", pillars, dayStem, result.MonthPillar.Zhi, desc)
}

func evidenceLead(result *bazipkg.BaziResult, focus string) string {
	switch focus {
	case "pattern":
		if strings.Contains(result.PatternAnalysis.PatternName, "羊刃") {
			return "能借得上的古法，重点都落在“刃要制化”："
		}
		return "翻到典籍里，和这一路相近的说法是："
	case "tiaohou":
		return "调候上可借的依据是："
	case "ten_gods":
		return "十神组合可参的句子是："
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
	explanation := []string{
		fmt.Sprintf("调候这一块，先别急着谈富贵，先看这盘的气候舒不舒服。%s生%s月，月令带湿土之气，日支又是%s，原局容易有“气重、湿滞、事情压住”的味道。当前调候取%s，取的是让局中之气能动起来：%s。", dayStem, result.MonthPillar.Zhi, result.DayPillar.Zhi, primary, firstNonEmpty(reason, "以月令寒暖燥湿判断。")),
	}
	if evidenceText != "" {
		explanation = append(explanation, "典籍里可借来看的线索是："+evidenceText)
	}
	explanation = append(explanation, fmt.Sprintf("但%s不是越多越好。调候是调气候，格局是论成败，两件事要合看：如果局里已经身旺，火土再多会把人推得更固执、更燥；这时木金的约束与疏泄反而更值钱。", primary))
	explanation = append(explanation, "现实里这类盘常见的不是没能力，而是推进方式太重：责任压得住，节奏却容易慢；想得清楚，落地要靠外部规则、期限、交付物来逼出流通。能把目标拆细、把成果拿出来，这盘就活。")
	return strings.Join(explanation, "\n\n")
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

func buildTenGodEvidenceContent(result *bazipkg.BaziResult, points []evidencePoint) string {
	top := topTenGodRatios(result.TenGodProportion, 4)
	topText := formatTopTenGods(top)
	evidenceText := formatEvidencePoints(points)
	explanation := []string{
		fmt.Sprintf("十神这段，不是看哪个百分比最大就完事，要看它在什么位置、能不能成事。此盘较显的是%s；四柱天干十神为%s。", topText, formatTenGodMap(result.TenGods)),
	}
	if evidenceText != "" {
		explanation = append(explanation, "能借用的经典线索是："+evidenceText)
	}
	explanation = append(explanation, fmt.Sprintf("落到人事上，日主%s又%s，比劫重的一面是能扛、敢争、资源意识强；不好的一面是容易把事情抓在自己手里，钱财合作上也要防“分夺”之象，账目、权责、边界最好早说清。", result.DayPillar.Gan+data.GanElement[result.DayPillar.Gan], result.BodyStrength.Verdict))
	explanation = append(explanation, "官杀若成体系，就像给这股力套上缰绳，适合走规则、职位、项目责任；食伤若得用，就把硬气转成技术、表达、产品和输出。最怕的是只剩印比助身：人很能撑，但转化率不高，忙、累、硬顶，最后还觉得别人跟不上。")
	if result.TenGodAnalysis != nil && result.TenGodAnalysis.Summary != "" {
		explanation = append(explanation, "按现有规则再补一句："+result.TenGodAnalysis.Summary)
	}
	return strings.Join(explanation, "\n\n")
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
