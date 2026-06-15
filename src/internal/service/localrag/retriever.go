package localrag

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"bazi/internal/service/rag"

	_ "github.com/mattn/go-sqlite3"
)

type Config struct {
	Enabled   bool
	IndexPath string
	MinScore  float64
	TopK      int
}

type Retriever struct {
	cfg Config
}

func NewRetriever(cfg Config) *Retriever {
	return &Retriever{cfg: cfg}
}

func (r *Retriever) Retrieve(ctx context.Context, req rag.RetrieveRequest) ([]rag.RetrievedChunk, error) {
	if r == nil || !r.cfg.Enabled {
		return nil, rag.ErrDisabled
	}
	if strings.TrimSpace(r.cfg.IndexPath) == "" {
		return nil, rag.ErrNotConfigured
	}
	if _, err := os.Stat(r.cfg.IndexPath); err != nil {
		return nil, fmt.Errorf("%w: %v", rag.ErrNotConfigured, err)
	}

	db, err := sql.Open("sqlite3", r.cfg.IndexPath+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", rag.ErrUpstream, err)
	}
	defer db.Close()

	topK := r.cfg.TopK
	if topK <= 0 {
		topK = 8
	}
	minScore := r.cfg.MinScore
	if minScore <= 0 {
		minScore = 0.35
	}

	terms := ExtractSearchTerms(req.Question, req.Focus)
	if len(terms) == 0 {
		return nil, nil
	}

	chunks, err := r.retrieveFTS(ctx, db, terms, topK)
	if err != nil {
		if strings.Contains(err.Error(), "no such module: fts5") || strings.Contains(err.Error(), "no such table: chunks_fts") {
			chunks, err = r.retrieveLike(ctx, db, terms, topK)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("%w: %v", rag.ErrUpstream, err)
	}
	if likeChunks, likeErr := r.retrieveLike(ctx, db, terms, topK); likeErr == nil {
		chunks = mergeRetrievedChunks(chunks, likeChunks)
	}
	chunks = rerankChunks(chunks, terms)
	chunks = normalizeScores(chunks, minScore)
	if len(chunks) > topK {
		chunks = chunks[:topK]
	}
	return chunks, nil
}

func (r *Retriever) retrieveFTS(ctx context.Context, db *sql.DB, terms []string, topK int) ([]rag.RetrievedChunk, error) {
	query := buildFTSQuery(terms)
	if query == "" {
		return nil, nil
	}
	rows, err := db.QueryContext(ctx, `
		SELECT c.id, c.content, c.book, c.chapter, c.source_path, c.title, c.is_index, c.document_id, bm25(chunks_fts) AS rank
		FROM chunks_fts
		JOIN chunks c ON c.id = chunks_fts.rowid
		WHERE chunks_fts MATCH ? AND c.domain = 'bazi'
		ORDER BY rank
		LIMIT ?`, query, topK*4)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []rag.RetrievedChunk{}
	for rows.Next() {
		chunk, rank, err := scanChunkWithRank(rows)
		if err != nil {
			return nil, err
		}
		_ = rank
		chunk.Score = scoreByRankPosition(len(out))
		out = append(out, chunk)
	}
	return out, rows.Err()
}

func scoreByRankPosition(index int) float64 {
	score := 1 - float64(index)*0.05
	if score < 0.05 {
		return 0.05
	}
	return score
}

func (r *Retriever) retrieveLike(ctx context.Context, db *sql.DB, terms []string, topK int) ([]rag.RetrievedChunk, error) {
	likeTerms := terms
	if len(likeTerms) > 8 {
		likeTerms = likeTerms[:8]
	}
	clauses := make([]string, 0, len(likeTerms))
	args := make([]interface{}, 0, len(likeTerms)+1)
	for _, term := range likeTerms {
		clauses = append(clauses, "c.content LIKE ?")
		args = append(args, "%"+term+"%")
	}
	args = append(args, topK*8)
	rows, err := db.QueryContext(ctx, `
		SELECT c.id, c.content, c.book, c.chapter, c.source_path, c.title, c.is_index, c.document_id
		FROM chunks c
		WHERE c.domain = 'bazi' AND (`+strings.Join(clauses, " OR ")+`)
		LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []rag.RetrievedChunk{}
	for rows.Next() {
		chunk, err := scanChunk(rows)
		if err != nil {
			return nil, err
		}
		chunk.Score = scoreByTermHits(chunk.Content, terms)
		out = append(out, chunk)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Score > out[j].Score
	})
	if len(out) > topK*4 {
		out = out[:topK*4]
	}
	return out, nil
}

func mergeRetrievedChunks(groups ...[]rag.RetrievedChunk) []rag.RetrievedChunk {
	seen := map[string]bool{}
	out := []rag.RetrievedChunk{}
	for _, group := range groups {
		for _, chunk := range group {
			key := chunk.ID
			if key == "" {
				key = chunk.Metadata["source_path"] + ":" + firstContentRunes(chunk.Content, 40)
			}
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, chunk)
		}
	}
	return out
}

func rerankChunks(chunks []rag.RetrievedChunk, terms []string) []rag.RetrievedChunk {
	for i := range chunks {
		hitScore := scoreByTermHits(chunks[i].Content, terms)
		if hitScore > 0 {
			chunks[i].Score = hitScore
		}
	}
	sort.SliceStable(chunks, func(i, j int) bool {
		return chunks[i].Score > chunks[j].Score
	})
	return chunks
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanChunkWithRank(rows rowScanner) (rag.RetrievedChunk, float64, error) {
	var (
		id         int64
		content    string
		book       string
		chapter    string
		sourcePath string
		title      string
		isIndex    int
		documentID string
		rank       float64
	)
	err := rows.Scan(&id, &content, &book, &chapter, &sourcePath, &title, &isIndex, &documentID, &rank)
	if err != nil {
		return rag.RetrievedChunk{}, 0, err
	}
	return buildChunk(id, content, book, chapter, sourcePath, title, isIndex, documentID), rank, nil
}

func scanChunk(rows rowScanner) (rag.RetrievedChunk, error) {
	var (
		id         int64
		content    string
		book       string
		chapter    string
		sourcePath string
		title      string
		isIndex    int
		documentID string
	)
	err := rows.Scan(&id, &content, &book, &chapter, &sourcePath, &title, &isIndex, &documentID)
	if err != nil {
		return rag.RetrievedChunk{}, err
	}
	return buildChunk(id, content, book, chapter, sourcePath, title, isIndex, documentID), nil
}

func buildChunk(id int64, content, book, chapter, sourcePath, title string, isIndex int, documentID string) rag.RetrievedChunk {
	return rag.RetrievedChunk{
		ID:         fmt.Sprintf("local:%d", id),
		Content:    content,
		DocumentID: documentID,
		Metadata: map[string]string{
			"domain":      "bazi",
			"book":        book,
			"chapter":     chapter,
			"source_path": sourcePath,
			"title":       title,
			"is_index":    fmt.Sprintf("%t", isIndex != 0),
		},
	}
}

func normalizeScores(chunks []rag.RetrievedChunk, minScore float64) []rag.RetrievedChunk {
	if len(chunks) == 0 {
		return chunks
	}
	maxScore := 0.0
	for _, chunk := range chunks {
		if chunk.Score > maxScore {
			maxScore = chunk.Score
		}
	}
	if maxScore <= 0 {
		return chunks
	}
	out := make([]rag.RetrievedChunk, 0, len(chunks))
	for _, chunk := range chunks {
		chunk.Score = chunk.Score / maxScore
		if chunk.Score < minScore {
			continue
		}
		out = append(out, chunk)
	}
	return out
}

func scoreByTermHits(content string, terms []string) float64 {
	score := 0.0
	for i, term := range terms {
		if strings.Contains(content, term) {
			runeCount := utf8.RuneCountInString(term)
			weight := 1.0 + float64(minInt(runeCount, 8))/2
			if i < 8 {
				weight += float64(8-i) * 0.75
			}
			if runeCount == 2 {
				weight += 1.5
			}
			score += weight
		}
	}
	return score
}

func buildFTSQuery(terms []string) string {
	parts := make([]string, 0, len(terms))
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" || utf8.RuneCountInString(term) < 3 {
			continue
		}
		parts = append(parts, quoteFTSTerm(term))
		if len(parts) >= 16 {
			break
		}
	}
	return strings.Join(parts, " OR ")
}

func quoteFTSTerm(term string) string {
	return `"` + strings.ReplaceAll(term, `"`, `""`) + `"`
}

var ganZhiSeqRe = regexp.MustCompile(`[甲乙丙丁戊己庚辛壬癸][子丑寅卯辰巳午未申酉戌亥]`)

func ExtractSearchTerms(question, focus string) []string {
	seen := map[string]bool{}
	out := []string{}
	add := func(s string) {
		s = strings.TrimSpace(strings.Trim(s, "，。；：、,. ;:=[]()（）"))
		if s == "" || seen[s] || utf8.RuneCountInString(s) < 2 {
			return
		}
		seen[s] = true
		out = append(out, s)
	}

	fields := map[string]string{}
	for _, line := range strings.Split(question, "\n") {
		key, val, ok := strings.Cut(line, "=")
		if ok {
			fields[strings.TrimSpace(key)] = strings.TrimSpace(val)
		}
	}
	if val := fields["格局"]; val != "" {
		for _, field := range strings.Fields(val) {
			add(field)
			add(field + "成败")
			add(field + "喜忌")
		}
	}
	if val := fields["日主"]; val != "" {
		add(val)
		if utf8.RuneCountInString(val) >= 2 {
			add(firstNRuneString(val, 2))
		}
	}
	if val := fields["月令"]; val != "" {
		add(val + "月")
		add("生" + val + "月")
	}
	for _, gz := range ganZhiSeqRe.FindAllString(question, -1) {
		add(gz)
	}
	if strings.Contains(fields["格局"], "羊刃") || strings.Contains(question, "羊刃") {
		for _, term := range []string{"羊刃", "阳刃", "日刃", "刃神", "七煞相制", "官杀制刃", "食伤泄秀", "羊刃格", "戊午"} {
			add(term)
		}
	}

	switch focus {
	case "pattern":
		for _, term := range []string{"月令提纲", "格局成败", "成格破格", "取格", "喜忌"} {
			add(term)
		}
	case "tiaohou":
		for _, term := range []string{"调候用神", "寒暖燥湿", "穷通宝鉴", "用神取法"} {
			add(term)
		}
	case "ten_gods":
		for _, term := range []string{"十神", "透干藏干", "财官印食", "官杀", "食伤"} {
			add(term)
		}
	default:
		for _, term := range []string{"月令提纲", "格局", "调候用神", "十神", "喜忌"} {
			add(term)
		}
	}

	for key, val := range fields {
		switch key {
		case "调候", "十神", "刑冲合害", "大运":
			for _, token := range strings.Fields(val) {
				if utf8.RuneCountInString(token) >= 2 {
					add(token)
				}
			}
		}
	}

	for _, phrase := range extractChinesePhrases(question) {
		add(phrase)
		if len(out) >= 32 {
			break
		}
	}
	return out
}

func firstContentRunes(s string, max int) string {
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= max {
		return string(runes)
	}
	return string(runes[:max])
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func extractChinesePhrases(s string) []string {
	var phrases []string
	var current []rune
	flush := func() {
		if len(current) >= 3 && len(current) <= 12 {
			phrases = append(phrases, string(current))
		}
		current = current[:0]
	}
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			current = append(current, r)
			if len(current) >= 12 {
				flush()
			}
			continue
		}
		flush()
	}
	flush()
	return phrases
}

func firstNRuneString(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n])
}
