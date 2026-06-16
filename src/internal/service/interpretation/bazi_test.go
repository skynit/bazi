package interpretation

import (
	"context"
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

func TestBuildQueryIncludesCoreFields(t *testing.T) {
	result, err := (&bazipkg.BaziService{}).Calculate(1990, 6, 15, 8, 0, model.GenderMale)
	if err != nil {
		t.Fatalf("calculate error: %v", err)
	}
	q := BuildQueryForTest(result, "pattern")
	for _, want := range []string{"四柱=", "日主=", "月令=", "十神=", "格局=", "调候=", "大运=", "重点检索格局"} {
		if !stringsContains(q, want) {
			t.Fatalf("query missing %q: %s", want, q)
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
}

func TestInterpretBaziOKFiltersAndCitations(t *testing.T) {
	svc := &Service{
		Charts: &mockChartStore{chart: makeChart()},
		Bazi:   &bazipkg.BaziService{},
		Retriever: fakeRetriever{chunks: []rag.RetrievedChunk{
			{ID: "1", Content: "目录页内容", Score: 0.9, Metadata: map[string]string{"is_index": "true"}},
			{ID: "2", Content: "十神透藏，须看月令与强弱。有效摘录", Score: 0.9, Metadata: map[string]string{"book": "子平真诠", "chapter": "001", "source_path": "bazi/子平真诠/001.md"}},
			{ID: "2", Content: "有效摘录重复", Score: 0.9, Metadata: map[string]string{"book": "子平真诠", "chapter": "001", "source_path": "bazi/子平真诠/001.md"}},
			{ID: "3", Content: "低分", Score: 0.1, Metadata: map[string]string{"book": "子平真诠", "chapter": "002", "source_path": "bazi/子平真诠/002.md"}},
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
		!strings.Contains(content, "四柱天干十神") ||
		!containsAny(content, "食伤", "财官食伤", "月令与强弱") {
		t.Fatalf("expected evidence in section content: %s", resp.Sections[0].Content)
	}
}

func TestInterpretBaziOverviewUsesSectionSpecificEvidence(t *testing.T) {
	svc := &Service{
		Charts: &mockChartStore{chart: makeChart()},
		Bazi:   &bazipkg.BaziService{},
		Retriever: fakeRetriever{chunks: []rag.RetrievedChunk{
			{ID: "p", Content: "月令为提纲，取格须看成败。", Score: 0.9, Metadata: map[string]string{"book": "子平真诠", "chapter": "001", "source_path": "bazi/子平真诠/001.md"}},
			{ID: "t", Content: "调候用神，须看寒暖燥湿。", Score: 0.85, Metadata: map[string]string{"book": "穷通宝鉴", "chapter": "003", "source_path": "bazi/穷通宝鉴/003.md"}},
			{ID: "g", Content: "十神组合，先辨透干藏干。", Score: 0.8, Metadata: map[string]string{"book": "三命通会", "chapter": "063", "source_path": "bazi/三命通会/063.md"}},
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
	if !strings.Contains(resp.Sections[0].Content, "月令为提纲") ||
		!strings.Contains(resp.Sections[1].Content, "调候用神") ||
		!strings.Contains(resp.Sections[2].Content, "十神组合") ||
		!containsAny(resp.Sections[0].Content, "制杀", "成败", "格局") ||
		!containsAny(resp.Sections[1].Content, "寒暖燥湿", "调候", "气候") ||
		!containsAny(resp.Sections[2].Content, "食伤", "财官食伤", "十神") {
		t.Fatalf("expected evidence-rich sections: %+v", resp.Sections)
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

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
