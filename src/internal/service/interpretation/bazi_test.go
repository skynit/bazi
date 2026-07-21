package interpretation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"bazi/internal/model"
	bazipkg "bazi/internal/service/bazi"
	"bazi/internal/service/rag"
)

type mockChartStore struct {
	chart *model.BirthChart
}

func (m *mockChartStore) FindByIDForUser(id uint, userID uint) (*model.BirthChart, error) {
	if m.chart != nil && m.chart.ID == id && m.chart.UserID == userID {
		return m.chart, nil
	}
	return nil, nil
}

type fakeRetriever struct {
	chunks []rag.RetrievedChunk
	err    error
}

type recordingRetriever struct {
	request rag.RetrieveRequest
	chunks  []rag.RetrievedChunk
}

func (r *recordingRetriever) Retrieve(_ context.Context, req rag.RetrieveRequest) ([]rag.RetrievedChunk, error) {
	r.request = req
	return r.chunks, nil
}

func (f fakeRetriever) Retrieve(ctx context.Context, req rag.RetrieveRequest) ([]rag.RetrievedChunk, error) {
	return f.chunks, f.err
}

func makeChart() *model.BirthChart {
	chart := &model.BirthChart{
		UserID:       1,
		BirthYear:    1990,
		BirthMonth:   6,
		BirthDay:     15,
		BirthHour:    8,
		BirthMin:     0,
		CalendarType: model.CalendarSolar,
		Gender:       model.GenderMale,
	}
	chart.ID = 1
	return chart
}

func eligibleCitationMetadata(book, chapter, sourcePath string) map[string]string {
	return map[string]string{
		"book":                   book,
		"author":                 "审阅作者",
		"edition":                "审阅版次",
		"chapter":                chapter,
		"locator":                "chapter:" + chapter,
		"source_path":            sourcePath,
		"artifact_path":          "library/reviewed.pdf",
		"artifact_sha256":        strings.Repeat("a", 64),
		"document_sha256":        strings.Repeat("b", 64),
		"source_tier":            "classical_text_local",
		"verification_status":    "bibliography_page_mapping_and_support_verified",
		"artifact_kind":          "published_scan",
		"provenance_status":      "bibliographic_provenance_verified",
		"independence_status":    "independent_primary_artifact_verified",
		"coverage_status":        "complete_primary_text_verified",
		"catalog_claim_eligible": "true",
		"catalog_schema":         "bazi_rag_source_catalog_v1",
		"catalog_version":        "reviewed-test-v1",
		"catalog_sha256":         strings.Repeat("c", 64),
	}
}

func TestBuildQueryIncludesCoreFields(t *testing.T) {
	result, err := (&bazipkg.BaziService{}).Calculate(1990, 6, 15, 8, 0, model.GenderMale)
	if err != nil {
		t.Fatalf("calculate error: %v", err)
	}
	q := BuildQueryForTest(result, "pattern")
	for _, want := range []string{"四柱=", "日主=", "月令=", "十神=", "格局候选=", "调候=", "大运=", "重点检索格局"} {
		if !stringsContains(q, want) {
			t.Fatalf("query missing %q: %s", want, q)
		}
	}
}

func TestInterpretBaziUsesPersistedNormalizedLunarBirth(t *testing.T) {
	baziSvc := &bazipkg.BaziService{}
	normalized, err := bazipkg.NormalizeBirthInput(bazipkg.BirthInput{
		Year: 2020, Month: 12, Day: 25, Hour: 8,
		CalendarType: model.CalendarLunar, Gender: model.GenderFemale, Timezone: bazipkg.DefaultBirthTimezone,
	})
	if err != nil {
		t.Fatalf("normalize lunar birth: %v", err)
	}
	normalizedJSON, err := json.Marshal(normalized)
	if err != nil {
		t.Fatalf("marshal normalized birth: %v", err)
	}
	chart := &model.BirthChart{
		UserID: 1, BirthYear: 2020, BirthMonth: 12, BirthDay: 25, BirthHour: 8,
		CalendarType: model.CalendarLunar, Gender: model.GenderFemale, Timezone: bazipkg.DefaultBirthTimezone,
		NormalizedBirth: normalizedJSON,
	}
	chart.ID = 1

	expected, err := baziSvc.CalculateNormalizedBirth(normalized)
	if err != nil {
		t.Fatalf("calculate normalized birth: %v", err)
	}
	rawSolar, err := baziSvc.Calculate(2020, 12, 25, 8, 0, model.GenderFemale)
	if err != nil {
		t.Fatalf("calculate same-number solar birth: %v", err)
	}
	expectedPillars := interpretationPillars(expected)
	rawSolarPillars := interpretationPillars(rawSolar)
	if expectedPillars == rawSolarPillars {
		t.Fatal("test input does not distinguish lunar conversion from the same-number solar date")
	}

	retriever := &recordingRetriever{chunks: []rag.RetrievedChunk{{
		ID: "p", Content: "月令为提纲，取格须核对透干。", Score: 0.9,
		Metadata: eligibleCitationMetadata("子平真诠", "001", "bazi/子平真诠/001.md"),
	}}}
	svc := &Service{Charts: &mockChartStore{chart: chart}, Bazi: baziSvc, Retriever: retriever, MinScore: 0.35, TopK: 8}
	resp, err := svc.InterpretBazi(context.Background(), Request{ChartID: 1, UserID: 1, Focus: "pattern"})
	if err != nil || resp.Status != StatusOK {
		t.Fatalf("interpret normalized lunar birth: err=%v resp=%+v", err, resp)
	}
	if !strings.Contains(retriever.request.Question, "四柱="+expectedPillars) {
		t.Fatalf("interpretation query is not bound to normalized lunar pillars: %s", retriever.request.Question)
	}
	if strings.Contains(retriever.request.Question, "四柱="+rawSolarPillars) {
		t.Fatalf("interpretation query used raw lunar fields as solar date: %s", retriever.request.Question)
	}
}

func interpretationPillars(result *bazipkg.BaziResult) string {
	return result.YearPillar.Gan + result.YearPillar.Zhi + " " +
		result.MonthPillar.Gan + result.MonthPillar.Zhi + " " +
		result.DayPillar.Gan + result.DayPillar.Zhi + " " +
		result.HourPillar.Gan + result.HourPillar.Zhi
}

func TestBuildQueryIncludesMonthCommandPatternCandidate(t *testing.T) {
	result, err := (&bazipkg.BaziService{}).CalculateFromPillars("癸未", "乙卯", "丙午", "辛卯", model.GenderMale)
	if err != nil {
		t.Fatalf("calculate from pillars: %v", err)
	}
	q := BuildQueryForTest(result, "pattern")
	for _, want := range []string{"正印格", "月令卯中乙透干", "bazi.pattern.month-command-exposure.v1"} {
		if !strings.Contains(q, want) {
			t.Fatalf("query missing month-command term %q: %s", want, q)
		}
	}
}

func TestBuildQueryKeepsConcreteBranchRelationTerms(t *testing.T) {
	result, err := (&bazipkg.BaziService{}).CalculateFromPillars("甲申", "丙子", "戊辰", "庚申", model.GenderMale)
	if err != nil {
		t.Fatalf("calculate from pillars: %v", err)
	}
	q := BuildQueryForTest(result, "pattern")
	for _, want := range []string{"地支三合局=申子辰", "申子辰", "目标水"} {
		if !strings.Contains(q, want) {
			t.Fatalf("query missing concrete relation term %q: %s", want, q)
		}
	}
	keywords := evidenceKeywords(result, "pattern")
	for _, want := range []string{"三合局", "申子辰", "水局"} {
		if !containsString(keywords, want) {
			t.Fatalf("pattern evidence keywords missing %q: %v", want, keywords)
		}
	}
}

func TestBuildPatternContentUsesMonthCommandEvidence(t *testing.T) {
	result, err := (&bazipkg.BaziService{}).CalculateFromPillars("癸未", "乙卯", "丙午", "辛卯", model.GenderMale)
	if err != nil {
		t.Fatalf("calculate from pillars: %v", err)
	}
	content := buildPatternContent(result)
	for _, want := range []string{"1项月令藏干透出候选", "正印格", "月支卯藏乙", "透于月干", "不表示优先级、成格结论或唯一格局"} {
		if !strings.Contains(content, want) {
			t.Fatalf("pattern content missing %q: %s", want, content)
		}
	}
	if strings.Contains(content, "未取得完整格局候选证据") {
		t.Fatalf("month-command evidence was discarded: %s", content)
	}
}

func TestInterpretBaziPatternReturnsCandidateEvidenceOnly(t *testing.T) {
	svc := &Service{
		Charts: &mockChartStore{chart: makeChart()},
		Bazi:   &bazipkg.BaziService{},
		Retriever: fakeRetriever{chunks: []rag.RetrievedChunk{
			{ID: "p", Content: "月令为提纲，取格须核对透干与根气。", Score: 0.9, Metadata: eligibleCitationMetadata("子平真诠", "001", "bazi/子平真诠/001.md")},
		}},
		MinScore: 0.35,
		TopK:     8,
	}
	resp, err := svc.InterpretBazi(context.Background(), Request{ChartID: 1, UserID: 1, Focus: "pattern"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != StatusOK || len(resp.Sections) != 1 || resp.Sections[0].Title != "格局规则候选" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	content := resp.Sections[0].Content
	for _, want := range []string{
		"classical_structural_detectors_v45",
		"古籍直接结构检测器",
		"not_validated",
		"not_adjudicated",
		"不决定喜忌或现实结果",
		"仅作为传统规则引用",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("pattern evidence missing %q: %s", want, content)
		}
	}
	for _, forbidden := range []string{
		"性情上，多半",
		"职位、作品、项目成果",
		"喜财印相扶",
		"压力就是权柄",
		"财来财去",
		"格局的好处才显",
	} {
		if strings.Contains(content, forbidden) || strings.Contains(resp.Summary, forbidden) {
			t.Fatalf("pattern interpretation leaked outcome claim %q: summary=%s content=%s", forbidden, resp.Summary, content)
		}
	}
}

func TestInterpretBaziFallbackDisabled(t *testing.T) {
	svc := &Service{Charts: &mockChartStore{chart: makeChart()}}
	resp, err := svc.InterpretBazi(context.Background(), Request{ChartID: 1, UserID: 1, Focus: "overview"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != StatusFallback || resp.Reason != ReasonDisabled {
		t.Fatalf("unexpected resp: %+v", resp)
	}
	if len(resp.Sections) == 0 || resp.Summary == "" {
		t.Fatalf("expected rule-only content: %+v", resp)
	}
	if len(resp.Sections) != 3 || resp.Sections[1].Title != "调候查表证据" {
		t.Fatalf("fallback mislabeled raw Tiaohou candidates: %+v", resp.Sections)
	}
}

func TestInterpretBaziFailsClosedWhenCitationMetadataIsIncomplete(t *testing.T) {
	svc := &Service{
		Charts: &mockChartStore{chart: makeChart()},
		Bazi:   &bazipkg.BaziService{},
		Retriever: fakeRetriever{chunks: []rag.RetrievedChunk{{
			ID: "legacy", Content: "月令为提纲，取格须看成败。", Score: 0.9,
			Metadata: map[string]string{"book": "子平真诠", "chapter": "001", "source_path": "bazi/子平真诠/001.md"},
		}}},
		MinScore: 0.35,
		TopK:     8,
	}
	resp, err := svc.InterpretBazi(context.Background(), Request{ChartID: 1, UserID: 1, Focus: "pattern"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Status != StatusFallback || resp.Reason != ReasonCitationMetadataIncomplete || len(resp.Citations) != 1 ||
		resp.Citations[0].ClaimEligible || len(resp.Citations[0].QuoteSHA256) != 64 {
		t.Fatalf("incomplete citation did not fail closed: %+v", resp)
	}
}

func TestInterpretBaziOKFiltersAndCitations(t *testing.T) {
	svc := &Service{
		Charts: &mockChartStore{chart: makeChart()},
		Bazi:   &bazipkg.BaziService{},
		Retriever: fakeRetriever{chunks: []rag.RetrievedChunk{
			{ID: "1", Content: "目录页内容", Score: 0.9, Metadata: map[string]string{"is_index": "true"}},
			{ID: "2", Content: "十神透藏，须看月令与强弱。有效摘录", Score: 0.9, Metadata: eligibleCitationMetadata("子平真诠", "001", "bazi/子平真诠/001.md")},
			{ID: "2", Content: "有效摘录重复", Score: 0.9, Metadata: eligibleCitationMetadata("子平真诠", "001", "bazi/子平真诠/001.md")},
			{ID: "3", Content: "低分", Score: 0.1, Metadata: eligibleCitationMetadata("子平真诠", "002", "bazi/子平真诠/002.md")},
		}},
		MinScore: 0.35,
		TopK:     8,
	}
	resp, err := svc.InterpretBazi(context.Background(), Request{ChartID: 1, UserID: 1, Focus: "ten_gods"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != StatusOK {
		t.Fatalf("expected ok, got %+v", resp)
	}
	if len(resp.Citations) != 1 || resp.Citations[0].Book != "子平真诠" {
		t.Fatalf("unexpected citations: %+v", resp.Citations)
	}
	if len(resp.Sections) == 0 || resp.Sections[0].Title != "十神结构" {
		t.Fatalf("unexpected sections: %+v", resp.Sections)
	}
	content := resp.Sections[0].Content
	if !strings.Contains(content, "十神透藏") ||
		!strings.Contains(content, "[1]") ||
		!strings.Contains(content, "等权计次") ||
		!strings.Contains(content, "不代表性格、职业、财富、关系或事件概率") {
		t.Fatalf("expected evidence in section content: %s", resp.Sections[0].Content)
	}
	for _, forbidden := range []string{"事业有成", "财运", "婚姻", "适合投资", "行动建议"} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("ten-god evidence content leaked outcome claim %q: %s", forbidden, content)
		}
	}
}

func TestInterpretBaziOverviewUsesSectionSpecificEvidence(t *testing.T) {
	svc := &Service{
		Charts: &mockChartStore{chart: makeChart()},
		Bazi:   &bazipkg.BaziService{},
		Retriever: fakeRetriever{chunks: []rag.RetrievedChunk{
			{ID: "p", Content: "月令为提纲，取格须看成败。", Score: 0.9, Metadata: eligibleCitationMetadata("子平真诠", "001", "bazi/子平真诠/001.md")},
			{ID: "t", Content: "调候用神，须看寒暖燥湿。", Score: 0.85, Metadata: eligibleCitationMetadata("穷通宝鉴", "003", "bazi/穷通宝鉴/003.md")},
			{ID: "g", Content: "十神组合，先辨透干藏干。", Score: 0.8, Metadata: eligibleCitationMetadata("三命通会", "063", "bazi/三命通会/063.md")},
		}},
		MinScore: 0.35,
		TopK:     8,
	}
	resp, err := svc.InterpretBazi(context.Background(), Request{ChartID: 1, UserID: 1, Focus: "overview"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Status != StatusOK {
		t.Fatalf("expected ok, got %+v", resp)
	}
	if len(resp.Sections) != 3 {
		t.Fatalf("expected 3 sections, got %+v", resp.Sections)
	}
	if resp.Sections[1].Title != "调候查表证据" {
		t.Fatalf("Tiaohou section title implies an adjudicated conclusion: %+v", resp.Sections[1])
	}
	if !strings.Contains(resp.Sections[0].Content, "月令为提纲") ||
		!strings.Contains(resp.Sections[1].Content, "调候用神") ||
		!strings.Contains(resp.Sections[2].Content, "十神组合") ||
		!containsAny(resp.Sections[0].Content, "制杀", "成败", "格局") ||
		!strings.Contains(resp.Sections[1].Content, "表首候选") ||
		!strings.Contains(resp.Sections[1].Content, "未裁决") ||
		!containsAny(resp.Sections[2].Content, "食伤", "财官食伤", "十神") {
		t.Fatalf("expected evidence-rich sections: %+v", resp.Sections)
	}
	for _, forbidden := range []string{"现实取象", "用在做事", "该立边界", "更容易成", "财气有路"} {
		if strings.Contains(resp.Sections[1].Content, forbidden) {
			t.Fatalf("Tiaohou evidence leaked action or outcome claim %q: %s", forbidden, resp.Sections[1].Content)
		}
	}
	for _, forbidden := range []string{"性情上，多半", "职位、作品、项目成果", "压力就是权柄", "财来财去"} {
		if strings.Contains(resp.Sections[0].Content, forbidden) {
			t.Fatalf("pattern evidence leaked outcome claim %q: %s", forbidden, resp.Sections[0].Content)
		}
	}
}

func TestInterpretBaziChartNotFound(t *testing.T) {
	svc := &Service{Charts: &mockChartStore{chart: makeChart()}}
	_, err := svc.InterpretBazi(context.Background(), Request{ChartID: 2, UserID: 1})
	if err == nil {
		t.Fatal("expected error")
	}
}

func stringsContains(s, sub string) bool {
	return strings.Index(s, sub) >= 0
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
