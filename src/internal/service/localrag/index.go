package localrag

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultChunkSize = 900
	defaultOverlap   = 120
)

type IndexConfig struct {
	SourceDir string
	IndexPath string
	KeepIndex bool
	ChunkSize int
	Overlap   int
}

type BuildStats struct {
	Documents int
	Chunks    int
	Skipped   int
}

func BuildIndex(ctx context.Context, cfg IndexConfig) (BuildStats, error) {
	var stats BuildStats
	sourceDir := strings.TrimSpace(cfg.SourceDir)
	indexPath := strings.TrimSpace(cfg.IndexPath)
	if sourceDir == "" {
		return stats, errors.New("source dir is required")
	}
	if indexPath == "" {
		return stats, errors.New("index path is required")
	}
	if _, err := os.Stat(sourceDir); err != nil {
		return stats, fmt.Errorf("source dir unavailable: %w", err)
	}

	chunkSize := cfg.ChunkSize
	if chunkSize <= 0 {
		chunkSize = defaultChunkSize
	}
	overlap := cfg.Overlap
	if overlap < 0 {
		overlap = 0
	}
	if overlap == 0 {
		overlap = defaultOverlap
	}
	if overlap >= chunkSize {
		overlap = chunkSize / 5
	}

	if err := os.MkdirAll(filepath.Dir(indexPath), 0o755); err != nil {
		return stats, err
	}
	tmpPath := indexPath + ".tmp"
	_ = os.Remove(tmpPath)
	defer os.Remove(tmpPath)

	db, err := sql.Open("sqlite3", tmpPath)
	if err != nil {
		return stats, err
	}
	defer db.Close()

	if err := createSchema(ctx, db); err != nil {
		return stats, err
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return stats, err
	}
	defer tx.Rollback()

	err = filepath.WalkDir(sourceDir, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			if d.Name() == ".obsidian" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.EqualFold(filepath.Ext(d.Name()), ".md") {
			stats.Skipped++
			return nil
		}

		meta, err := metadataForPath(sourceDir, path)
		if err != nil {
			stats.Skipped++
			return nil
		}
		if meta.IsIndex && !cfg.KeepIndex {
			stats.Skipped++
			return nil
		}

		contentBytes, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		content := strings.TrimSpace(string(contentBytes))
		if content == "" {
			stats.Skipped++
			return nil
		}
		meta.Title = firstMarkdownTitle(content, meta.Title)
		chunks := ChunkMarkdown(content, chunkSize, overlap)
		if len(chunks) == 0 {
			stats.Skipped++
			return nil
		}

		stats.Documents++
		for _, chunk := range chunks {
			if err := insertChunk(ctx, tx, meta, chunk); err != nil {
				return err
			}
			stats.Chunks++
		}
		return nil
	})
	if err != nil {
		return stats, err
	}
	if err := tx.Commit(); err != nil {
		return stats, err
	}
	if err := db.Close(); err != nil {
		return stats, err
	}
	if err := os.Rename(tmpPath, indexPath); err != nil {
		return stats, err
	}
	return stats, nil
}

type documentMetadata struct {
	Domain     string
	Book       string
	Chapter    string
	SourcePath string
	Title      string
	IsIndex    bool
}

func createSchema(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT NOT NULL,
			book TEXT NOT NULL,
			chapter TEXT NOT NULL,
			source_path TEXT NOT NULL,
			title TEXT NOT NULL,
			content TEXT NOT NULL,
			is_index INTEGER NOT NULL DEFAULT 0,
			document_id TEXT NOT NULL
		)`,
		`CREATE INDEX idx_chunks_source_path ON chunks(source_path)`,
		`CREATE VIRTUAL TABLE chunks_fts USING fts5(content, book, chapter, title, tokenize='trigram')`,
	}
	for _, stmt := range stmts {
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			if strings.Contains(err.Error(), "no such module: fts5") {
				return fmt.Errorf("sqlite fts5 is not enabled; build with -tags sqlite_fts5: %w", err)
			}
			return err
		}
	}
	return nil
}

func insertChunk(ctx context.Context, tx *sql.Tx, meta documentMetadata, content string) error {
	isIndex := 0
	if meta.IsIndex {
		isIndex = 1
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO chunks
		(domain, book, chapter, source_path, title, content, is_index, document_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		meta.Domain, meta.Book, meta.Chapter, meta.SourcePath, meta.Title, content, isIndex, meta.SourcePath)
	if err != nil {
		return err
	}
	rowID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO chunks_fts(rowid, content, book, chapter, title) VALUES (?, ?, ?, ?, ?)`,
		rowID, content, meta.Book, meta.Chapter, meta.Title)
	return err
}

func metadataForPath(root, path string) (documentMetadata, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return documentMetadata{}, err
	}
	rel = filepath.ToSlash(rel)
	parts := strings.Split(rel, "/")
	if len(parts) == 0 || parts[0] == "." || parts[0] == "" {
		return documentMetadata{}, fmt.Errorf("invalid source path: %s", path)
	}
	book := parts[0]
	fileName := filepath.Base(rel)
	chapter := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	return documentMetadata{
		Domain:     "bazi",
		Book:       book,
		Chapter:    chapter,
		SourcePath: "bazi/" + rel,
		Title:      chapter,
		IsIndex:    strings.EqualFold(fileName, "000.md"),
	}, nil
}

func ChunkMarkdown(content string, maxRunes int, overlap int) []string {
	content = normalizeMarkdown(content)
	if content == "" {
		return nil
	}
	if maxRunes <= 0 {
		maxRunes = defaultChunkSize
	}
	if overlap < 0 {
		overlap = 0
	}
	if overlap >= maxRunes {
		overlap = maxRunes / 5
	}

	units := splitMarkdownUnits(content)
	chunks := make([]string, 0, len(units))
	var current string
	flush := func() {
		current = strings.TrimSpace(current)
		if current != "" {
			chunks = append(chunks, current)
		}
	}

	for _, unit := range units {
		if utf8.RuneCountInString(unit) > maxRunes {
			flush()
			current = ""
			for _, piece := range splitLongText(unit, maxRunes, overlap) {
				if strings.TrimSpace(piece) != "" {
					chunks = append(chunks, strings.TrimSpace(piece))
				}
			}
			continue
		}
		next := strings.TrimSpace(unit)
		if next == "" {
			continue
		}
		if current == "" {
			current = next
			continue
		}
		candidate := current + "\n\n" + next
		if utf8.RuneCountInString(candidate) > maxRunes {
			flush()
			current = tailRunes(current, overlap)
			if current != "" {
				current += "\n\n" + next
			} else {
				current = next
			}
			continue
		}
		current = candidate
	}
	flush()
	return chunks
}

func normalizeMarkdown(content string) string {
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")
	cleaned := make([]string, 0, len(lines))
	blank := false
	for _, line := range lines {
		line = strings.TrimRight(line, " \t")
		if strings.TrimSpace(line) == "" {
			if !blank {
				cleaned = append(cleaned, "")
			}
			blank = true
			continue
		}
		cleaned = append(cleaned, line)
		blank = false
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

func splitMarkdownUnits(content string) []string {
	blocks := strings.Split(content, "\n\n")
	units := make([]string, 0, len(blocks))
	for _, block := range blocks {
		block = strings.TrimSpace(block)
		if block != "" {
			units = append(units, block)
		}
	}
	return units
}

func splitLongText(s string, maxRunes int, overlap int) []string {
	runes := []rune(s)
	out := []string{}
	step := maxRunes - overlap
	if step <= 0 {
		step = maxRunes
	}
	for start := 0; start < len(runes); start += step {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}
		out = append(out, string(runes[start:end]))
		if end == len(runes) {
			break
		}
	}
	return out
}

func tailRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	runes := []rune(strings.TrimSpace(s))
	if len(runes) <= n {
		return string(runes)
	}
	return strings.TrimSpace(string(runes[len(runes)-n:]))
}

func firstMarkdownTitle(content, fallback string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			title := strings.TrimSpace(strings.TrimLeft(line, "#"))
			if title != "" {
				return title
			}
		}
	}
	return fallback
}
