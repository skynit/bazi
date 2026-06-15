package localrag

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"bazi/internal/service/rag"

	_ "github.com/mattn/go-sqlite3"
)

func TestChunkMarkdown(t *testing.T) {
	content := "# 子平真诠\n\n月令为提纲，格局当以月令为主。\n\n调候用神，须看寒暖燥湿。"
	chunks := ChunkMarkdown(content, 24, 6)
	if len(chunks) < 2 {
		t.Fatalf("expected multiple chunks, got %+v", chunks)
	}
	if !strings.Contains(chunks[0], "子平真诠") {
		t.Fatalf("expected title in first chunk: %+v", chunks)
	}
}

func TestExtractSearchTerms(t *testing.T) {
	terms := ExtractSearchTerms("日主=甲木\n月令=寅\n四柱=甲子 乙丑 丙寅 丁卯\n格局=正官格", "pattern")
	joined := strings.Join(terms, "|")
	for _, want := range []string{"月令提纲", "甲木", "寅月", "正官格", "甲子"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing term %q in %v", want, terms)
		}
	}
}

func TestRetrieverFallsBackToLikeWithoutFTS(t *testing.T) {
	indexPath := filepath.Join(t.TempDir(), "bazi_fts.db")
	db, err := sql.Open("sqlite3", indexPath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE chunks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		domain TEXT NOT NULL,
		book TEXT NOT NULL,
		chapter TEXT NOT NULL,
		source_path TEXT NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		is_index INTEGER NOT NULL DEFAULT 0,
		document_id TEXT NOT NULL
	)`)
	if err != nil {
		t.Fatalf("create chunks: %v", err)
	}
	_, err = db.Exec(`INSERT INTO chunks
		(domain, book, chapter, source_path, title, content, is_index, document_id)
		VALUES ('bazi', '子平真诠', '001', 'bazi/子平真诠/001.md', '论月令', '月令提纲，格局当以月令为主。', 0, 'bazi/子平真诠/001.md')`)
	if err != nil {
		t.Fatalf("insert chunk: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close sqlite: %v", err)
	}

	retriever := NewRetriever(Config{Enabled: true, IndexPath: indexPath, MinScore: 0.1, TopK: 4})
	chunks, err := retriever.Retrieve(context.Background(), rag.RetrieveRequest{
		Question: "八字经典依据检索\nfocus=pattern\n重点检索格局、月令提纲、喜忌",
		Focus:    "pattern",
	})
	if err != nil {
		t.Fatalf("Retrieve error: %v", err)
	}
	if len(chunks) != 1 {
		t.Fatalf("expected one chunk, got %+v", chunks)
	}
	if chunks[0].Metadata["book"] != "子平真诠" || chunks[0].Score <= 0 {
		t.Fatalf("unexpected chunk: %+v", chunks[0])
	}
}

func TestRetrieverMissingIndex(t *testing.T) {
	retriever := NewRetriever(Config{Enabled: true, IndexPath: filepath.Join(t.TempDir(), "missing.db")})
	_, err := retriever.Retrieve(context.Background(), rag.RetrieveRequest{Question: "月令提纲", Focus: "overview"})
	if err == nil {
		t.Fatal("expected missing index error")
	}
}
