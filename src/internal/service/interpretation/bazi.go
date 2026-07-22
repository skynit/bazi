package interpretation

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

	ReasonDisabled                   = "disabled"
	ReasonNotConfigured              = "not_configured"
	ReasonTimeout                    = "timeout"
	ReasonEmpty                      = "empty"
	ReasonUpstreamError              = "upstream_error"
	ReasonInvalidResponse            = "invalid_response"
	ReasonRuleOnly                   = "rule_only"
	ReasonCitationMetadataIncomplete = "citation_metadata_incomplete"
	ReasonCitationNotSupporting      = "citation_not_supporting_claim"
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
	resolved, err := bazipkg.ResolveStoredBirth(baziSvc, chart)
	if err != nil {
		return resp, fmt.Errorf("%w: %v", ErrComputeChart, err)
	}
	result := resolved.Result

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
	if !hasClaimEligibleCitation(resp.Citations) {
		resp.Reason = ReasonCitationMetadataIncomplete
		return resp, nil
	}
	evidenceSections := buildEvidenceSections(result, focus, evidence, resp.Citations)
	if !hasSectionCitation(evidenceSections) {
		resp.Reason = ReasonCitationNotSupporting
		return resp, nil
	}
	resp.Sections = evidenceSections
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
		return "格局线索"
	case "tiaohou":
		return "调候参考"
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
	pattern := formatPatternSearchTerms(result.PatternAnalysis)
	tiaohou := ""
	if result.Tiaohou != nil {
		tiaohou = fmt.Sprintf("%s生%s月 表首候选%s 状态未裁决", result.Tiaohou.Stem, result.Tiaohou.Month, result.Tiaohou.TablePrimaryCandidate)
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
		"格局候选=" + pattern,
		"身强评分分段候选=" + result.BodyStrength.ScoreBandCandidate,
		"调候=" + tiaohou,
		"刑冲合害=" + relations,
		"大运=" + dayun,
	}

	switch focus {
	case "pattern":
		parts = append(parts, "重点检索格局规则、月令提纲、候选条件、争议")
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

func formatPatternSearchTerms(analysis bazipkg.PatternAnalysis) string {
	parts := make([]string, 0, len(analysis.Candidates)+len(analysis.MonthCommandEvidence))
	for _, candidate := range analysis.Candidates {
		parts = append(parts, candidate.PatternName+" "+candidate.RuleID)
	}
	for _, evidence := range analysis.MonthCommandEvidence {
		for _, name := range evidence.CandidateNames {
			parts = append(parts, fmt.Sprintf(
				"%s 月令%s中%s透干 %s",
				name,
				evidence.MonthBranch,
				evidence.HiddenStem,
				evidence.RuleID,
			))
		}
	}
	parts = uniqueStrings(parts)
	if len(parts) == 0 {
		return "无可用候选"
	}
	return strings.Join(parts, "；")
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
	parts := make([]string, 0, len(g.GanRelations)+len(g.ZhiRelations))
	for _, rel := range g.GanRelations {
		if rel.Type != "五合" && rel.Type != "天干相冲" {
			continue
		}
		item := fmt.Sprintf("天干%s=%s/%s", rel.Type, strings.Join(uniqueStrings(rel.Stems), ""), strings.Join(rel.Pillars, "-"))
		if rel.Subtype != "" {
			item += "/" + rel.Subtype
		}
		if rel.TargetElement != "" {
			item += "/目标" + rel.TargetElement
		}
		parts = append(parts, item)
	}
	for _, rel := range g.ZhiRelations {
		item := fmt.Sprintf("地支%s=%s/%s", rel.Type, strings.Join(uniqueStrings(rel.Branches), ""), strings.Join(rel.Pillars, "-"))
		if rel.Subtype != "" && rel.Subtype != strings.Join(uniqueStrings(rel.Branches), "") {
			item += "/" + rel.Subtype
		}
		if rel.TargetElement != "" {
			item += "/目标" + rel.TargetElement
		}
		parts = append(parts, item)
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
		quote := firstRunes(cleanContent(chunk.Content), 120)
		citation := model.InterpretationCitation{
			ID:                 i + 1,
			Book:               firstNonEmpty(meta["book"], "未标注典籍"),
			Author:             firstNonEmpty(meta["author"], "unrecorded"),
			Edition:            firstNonEmpty(meta["edition"], "unrecorded"),
			Volume:             meta["volume"],
			Chapter:            firstNonEmpty(meta["chapter"], meta["title"]),
			Page:               meta["page"],
			Locator:            meta["locator"],
			Path:               path,
			ArtifactPath:       meta["artifact_path"],
			ArtifactSHA256:     strings.ToLower(strings.TrimSpace(meta["artifact_sha256"])),
			DocumentSHA256:     strings.ToLower(strings.TrimSpace(meta["document_sha256"])),
			Quote:              quote,
			QuoteSHA256:        citationTextSHA256(quote),
			SourceTier:         firstNonEmpty(meta["source_tier"], "bronze_unverified"),
			VerificationStatus: firstNonEmpty(meta["verification_status"], "source_metadata_missing"),
			ArtifactKind:       firstNonEmpty(meta["artifact_kind"], "unregistered"),
			ProvenanceStatus:   firstNonEmpty(meta["provenance_status"], "source_metadata_missing"),
			IndependenceStatus: firstNonEmpty(meta["independence_status"], "unknown"),
			CoverageStatus:     firstNonEmpty(meta["coverage_status"], "unknown"),
			CatalogSchema:      meta["catalog_schema"],
			CatalogVersion:     meta["catalog_version"],
			CatalogSHA256:      strings.ToLower(strings.TrimSpace(meta["catalog_sha256"])),
			Score:              chunk.Score,
		}
		citation.ClaimEligible = strings.EqualFold(strings.TrimSpace(meta["catalog_claim_eligible"]), "true") &&
			validInterpretationCitation(citation)
		out = append(out, citation)
	}
	return out
}

func validInterpretationCitation(citation model.InterpretationCitation) bool {
	return citation.Book != "" && citation.Book != "未标注典籍" &&
		citation.Author != "" && citation.Author != "unrecorded" &&
		citation.Edition != "" && citation.Edition != "unrecorded" &&
		citation.Path != "" && (citation.Page != "" || citation.Locator != "") &&
		citation.ArtifactPath != "" && validInterpretationSHA256(citation.ArtifactSHA256) &&
		validInterpretationSHA256(citation.DocumentSHA256) && citation.Quote != "" &&
		validInterpretationSHA256(citation.QuoteSHA256) && citation.SourceTier == "classical_text_local" &&
		citation.VerificationStatus == "bibliography_page_mapping_and_support_verified" &&
		validClaimArtifactKind(citation.ArtifactKind) &&
		citation.ProvenanceStatus == "bibliographic_provenance_verified" &&
		citation.IndependenceStatus == "independent_primary_artifact_verified" &&
		citation.CoverageStatus == "complete_primary_text_verified" &&
		citation.CatalogSchema == "bazi_rag_source_catalog_v1" && citation.CatalogVersion != "" &&
		validInterpretationSHA256(citation.CatalogSHA256)
}

func validClaimArtifactKind(value string) bool {
	return value == "published_scan" || value == "publisher_digital_edition"
}

func hasClaimEligibleCitation(citations []model.InterpretationCitation) bool {
	for _, citation := range citations {
		if citation.ClaimEligible {
			return true
		}
	}
	return false
}

func hasSectionCitation(sections []model.InterpretationSection) bool {
	for _, section := range sections {
		if len(section.Citations) > 0 {
			return true
		}
	}
	return false
}

func citationTextSHA256(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func validInterpretationSHA256(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
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
		if citation.ClaimEligible && citation.Path != "" {
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

func evidenceKeywords(result *bazipkg.BaziResult, focus string) []string {
	dayMaster := result.DayPillar.Gan
	dayStem := dayMaster + data.GanElement[dayMaster]
	month := result.MonthPillar.Zhi
	keywords := []string{dayMaster, dayStem, month + "月"}
	switch focus {
	case "pattern":
		keywords = append(keywords, "月令", "格局", "取格", "候选", "争议", result.DayPillar.Gan+result.DayPillar.Zhi)
		for _, candidate := range result.PatternAnalysis.Candidates {
			keywords = append(keywords, candidate.PatternName, candidate.Category, candidate.Source)
		}
		for _, evidence := range result.PatternAnalysis.MonthCommandEvidence {
			keywords = append(keywords, evidence.CandidateNames...)
			keywords = append(keywords, evidence.HiddenStem, evidence.HiddenTenGod, evidence.Source)
		}
		for _, relation := range result.GanZhiAnalysis.GanRelations {
			if relation.Type != "五合" && relation.Type != "天干相冲" {
				continue
			}
			keywords = append(keywords, relation.Type, relation.Subtype, strings.Join(uniqueStrings(relation.Stems), ""))
			if relation.TargetElement != "" {
				keywords = append(keywords, "化"+relation.TargetElement)
			}
		}
		for _, relation := range result.GanZhiAnalysis.ZhiRelations {
			keywords = append(keywords, relation.Type, relation.Subtype, strings.Join(uniqueStrings(relation.Branches), ""))
			if relation.TargetElement != "" {
				keywords = append(keywords, relation.TargetElement+"局")
			}
		}
	case "tiaohou":
		keywords = append(keywords, "调候", "用神", "寒", "暖", "燥", "湿", "丙", "癸", "辰月")
		if result.Tiaohou != nil {
			keywords = append(keywords, result.Tiaohou.TablePrimaryCandidate)
			for _, rule := range result.Tiaohou.Rules {
				keywords = append(keywords, rule.XiShen, rule.JiShen, rule.SourceText)
			}
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
	explanation := []string{fmt.Sprintf(
		"命盘四柱为%s，月支为%s。以下结合月令、透干和干支关系整理格局线索。",
		pillarText(result),
		p.Inputs.MonthBranch,
	)}
	if len(p.Candidates) > 0 {
		explanation = append(explanation, "当前观察到的格局线索："+formatPatternCandidates(p.Candidates)+"。")
	} else {
		explanation = append(explanation, "当前没有观察到可直接列出的格局线索。")
	}
	if len(p.MonthCommandEvidence) > 0 {
		explanation = append(explanation, "月令藏干透出关系："+formatMonthCommandPatternEvidence(p.MonthCommandEvidence)+"。")
	}
	evidenceText := formatEvidencePoints(points)
	if evidenceText != "" {
		explanation = append(explanation, "相关典籍原文："+evidenceText)
	}
	explanation = append(explanation, "这些内容用于理解命盘结构，不等同于已经确定格局、喜忌或现实结果。")
	return strings.Join(explanation, "\n\n")
}

func formatPatternCandidates(candidates []bazipkg.PatternCandidate) string {
	parts := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		item := fmt.Sprintf(
			"%s（%s，来源：%s）",
			candidate.PatternName,
			candidate.Category,
			candidate.Source,
		)
		parts = append(parts, item)
	}
	return strings.Join(parts, "；")
}

func formatMonthCommandPatternEvidence(evidence []bazipkg.MonthCommandPatternEvidence) string {
	parts := make([]string, 0, len(evidence))
	for _, item := range evidence {
		if len(item.CandidateNames) == 0 {
			continue
		}
		exposures := make([]string, 0, len(item.Exposures))
		for _, exposure := range item.Exposures {
			exposures = append(exposures, exposure.Pillar)
		}
		parts = append(parts, fmt.Sprintf(
			"%s（月支%s藏%s，%s，透于%s）",
			strings.Join(item.CandidateNames, "、"),
			item.MonthBranch,
			item.HiddenStem,
			item.HiddenTenGod,
			strings.Join(exposures, "、"),
		))
	}
	return strings.Join(parts, "；")
}

func buildTiaohouEvidenceContent(result *bazipkg.BaziResult, points []evidencePoint) string {
	explanation := []string{"当前未取得完整调候查表证据。"}
	if result.Tiaohou != nil {
		explanation[0] = fmt.Sprintf(
			"按日干%s与月支%s查阅传统调候表，表内首先列出%s作为参考。",
			result.Tiaohou.Stem,
			result.Tiaohou.Month,
			result.Tiaohou.TablePrimaryCandidate,
		)
		depth := result.Tiaohou.DepthEvidence
		if depth.Status == "observed" {
			explanation = append(explanation, fmt.Sprintf(
				"出生时刻位于%s至%s节令区间的%s，约处于该区间的%.1f%%。",
				depth.StartTerm,
				depth.EndTerm,
				depth.Phase,
				depth.Position*100,
			))
		} else {
			explanation = append(explanation, "缺少可定位出生时刻，节令区间深浅不可用。")
		}
	}
	evidenceText := formatEvidencePoints(points)
	if evidenceText != "" {
		explanation = append(explanation, "相关典籍原文："+evidenceText)
	}
	explanation = append(explanation, "调候条目是传统查表参考，不代表唯一用神、现实吉凶或行动建议。")
	return strings.Join(explanation, "\n\n")
}

func buildTenGodEvidenceContent(result *bazipkg.BaziResult, points []evidencePoint) string {
	top := topTenGodRatios(result.TenGodProportion, 4)
	topText := formatTopTenGods(top)
	evidenceText := formatEvidencePoints(points)
	explanation := []string{fmt.Sprintf(
		"按命盘中的非日主透干与四支藏干统计，当前出现次数较高的项目为：%s。",
		topText,
	)}
	if evidenceText != "" {
		explanation = append(explanation, "相关典籍原文："+evidenceText)
	}
	explanation = append(explanation, "出现次数只表示命盘中的分布，不直接代表性格、职业、财富、关系或事件概率。")
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
	candidateCount := len(result.PatternAnalysis.Candidates)
	monthCommandCount := len(result.PatternAnalysis.MonthCommandEvidence)

	switch focus {
	case "pattern":
		return fmt.Sprintf("%s，月支%s与日主%s形成%d项格局线索和%d项月令透干线索。以下内容用于理解结构，不直接确定格局或喜忌。", source, result.MonthPillar.Zhi, dayMaster, candidateCount, monthCommandCount)
	case "tiaohou":
		if result.Tiaohou != nil {
			return fmt.Sprintf("%s，按%s日主与%s月查阅传统调候表，表内首先列出%s作为参考。", source, result.Tiaohou.Stem, result.Tiaohou.Month, result.Tiaohou.TablePrimaryCandidate)
		}
		return fmt.Sprintf("%s，暂未找到日主%s与月令%s对应的调候条目。", source, dayMaster, result.MonthPillar.Zhi)
	case "ten_gods":
		return fmt.Sprintf("%s，重点记录日主%s对应的十神透藏映射。%s", source, dayMaster, buildTenGodContent(result))
	default:
		return fmt.Sprintf("%s，四柱为%s%s、%s%s、%s%s、%s%s；日主为%s，月支为%s。当前整理出%d项格局线索、%d项月令透干线索，五行强弱参考区间为%s。",
			source,
			result.YearPillar.Gan, result.YearPillar.Zhi,
			result.MonthPillar.Gan, result.MonthPillar.Zhi,
			result.DayPillar.Gan, result.DayPillar.Zhi,
			result.HourPillar.Gan, result.HourPillar.Zhi,
			dayMaster,
			result.MonthPillar.Zhi,
			candidateCount,
			monthCommandCount,
			result.BodyStrength.ScoreBandCandidate,
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
			Title:     "格局线索",
			Content:   buildPatternContent(result),
			Citations: ids(3),
		}}
	case "tiaohou":
		return []model.InterpretationSection{{
			Title:     "调候参考",
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
			{Title: "格局线索", Content: buildPatternContent(result), Citations: ids(2)},
			{Title: "调候参考", Content: buildTiaohouContent(result), Citations: ids(4)},
			{Title: "十神结构", Content: buildTenGodContent(result), Citations: ids(5)},
		}
	}
}

func buildPatternContent(result *bazipkg.BaziResult) string {
	p := result.PatternAnalysis
	if len(p.Candidates) == 0 && len(p.MonthCommandEvidence) == 0 {
		return fmt.Sprintf("月支%s与日主%s之间暂未观察到可直接列出的格局线索。", result.MonthPillar.Zhi, result.DayPillar.Gan)
	}
	parts := []string{fmt.Sprintf(
		"命盘四柱为%s，月支为%s。当前观察到%d项特殊结构线索和%d项月令藏干透出线索。",
		pillarText(result),
		p.Inputs.MonthBranch,
		len(p.Candidates),
		len(p.MonthCommandEvidence),
	)}
	if len(p.MonthCommandEvidence) > 0 {
		parts = append(parts, "月令候选为"+formatMonthCommandPatternEvidence(p.MonthCommandEvidence)+"。")
	}
	parts = append(parts, "这些线索用于理解命盘结构，不等同于已经确定格局、喜忌或现实结果。")
	return strings.Join(parts, "")
}

func buildTiaohouContent(result *bazipkg.BaziResult) string {
	if result.Tiaohou == nil {
		return "当前没有找到与日干、月支相对应的调候条目。"
	}
	return fmt.Sprintf(
		"按日干%s与月支%s查阅传统调候表，共找到%d条对应条目，首先列出%s作为参考。调候条目用于观察寒暖燥湿，不代表唯一用神或现实吉凶。",
		result.Tiaohou.Stem,
		result.Tiaohou.Month,
		len(result.Tiaohou.Rules),
		result.Tiaohou.TablePrimaryCandidate,
	)
}

func buildTenGodContent(result *bazipkg.BaziResult) string {
	if result.TenGodAnalysis != nil && result.TenGodAnalysis.Status == "observed" {
		return fmt.Sprintf(
			"按命盘中的非日主透干与四支藏干统计，共记录%d次十神对应关系，其中出现较多的是%s（%.2f%%）。次数只表示命盘中的分布，不直接代表性格、职业、财富、关系或具体事件。",
			result.TenGodAnalysis.TotalOccurrences,
			strings.Join(result.TenGodAnalysis.DominantGods, "、"),
			result.TenGodAnalysis.DominantPercent,
		)
	}
	return "当前没有取得可用于汇总的十神透藏数据。"
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
