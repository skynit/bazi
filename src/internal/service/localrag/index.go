package localrag

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
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
	SourceDir   string
	IndexPath   string
	CatalogPath string
	KeepIndex   bool
	ChunkSize   int
	Overlap     int
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
	catalog, err := loadSourceCatalog(cfg.CatalogPath)
	if err != nil {
		return stats, fmt.Errorf("source catalog unavailable: %w", err)
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

		meta, err := metadataForPath(sourceDir, path, catalog)
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
		documentSum := sha256.Sum256(contentBytes)
		meta.DocumentSHA256 = hex.EncodeToString(documentSum[:])
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
	Domain               string
	Book                 string
	Author               string
	Edition              string
	Volume               string
	Chapter              string
	Page                 string
	Locator              string
	SourcePath           string
	Title                string
	ArtifactPath         string
	ArtifactSHA256       string
	DocumentSHA256       string
	SourceTier           string
	VerificationStatus   string
	ArtifactKind         string
	ProvenanceStatus     string
	IndependenceStatus   string
	CoverageStatus       string
	CatalogClaimEligible bool
	CatalogSchema        string
	CatalogVersion       string
	CatalogSHA256        string
	IsIndex              bool
}

type sourceCatalog struct {
	Schema        string                     `json:"schema"`
	Version       string                     `json:"version"`
	Description   string                     `json:"description"`
	DefaultPolicy sourceCatalogDefaultPolicy `json:"default_policy"`
	Sources       []sourceCatalogEntry       `json:"sources"`
	byRoot        map[string]sourceCatalogEntry
	sha256        string
}

type sourceCatalogDefaultPolicy struct {
	SourceTier         string `json:"source_tier"`
	VerificationStatus string `json:"verification_status"`
	ArtifactKind       string `json:"artifact_kind"`
	ProvenanceStatus   string `json:"provenance_status"`
	IndependenceStatus string `json:"independence_status"`
	CoverageStatus     string `json:"coverage_status"`
	ClaimEligible      bool   `json:"claim_eligible"`
}

type sourceCatalogEntry struct {
	Book               string   `json:"book"`
	Author             string   `json:"author"`
	Edition            string   `json:"edition"`
	ArtifactPath       string   `json:"artifact_path"`
	ArtifactSHA256     string   `json:"artifact_sha256"`
	MarkdownRoot       string   `json:"markdown_root"`
	SourceTier         string   `json:"source_tier"`
	VerificationStatus string   `json:"verification_status"`
	ArtifactKind       string   `json:"artifact_kind"`
	ProvenanceStatus   string   `json:"provenance_status"`
	IndependenceStatus string   `json:"independence_status"`
	CoverageStatus     string   `json:"coverage_status"`
	PageMappingStatus  string   `json:"page_mapping_status"`
	LicenseScope       string   `json:"license_scope"`
	ClaimEligible      bool     `json:"claim_eligible"`
	Limitations        []string `json:"limitations"`
}

func createSchema(ctx context.Context, db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			domain TEXT NOT NULL,
			book TEXT NOT NULL,
			author TEXT NOT NULL,
			edition TEXT NOT NULL,
			volume TEXT NOT NULL,
			chapter TEXT NOT NULL,
			page TEXT NOT NULL,
			locator TEXT NOT NULL,
			source_path TEXT NOT NULL,
			title TEXT NOT NULL,
			artifact_path TEXT NOT NULL,
			artifact_sha256 TEXT NOT NULL,
			document_sha256 TEXT NOT NULL,
			source_tier TEXT NOT NULL,
			verification_status TEXT NOT NULL,
			artifact_kind TEXT NOT NULL,
			provenance_status TEXT NOT NULL,
			independence_status TEXT NOT NULL,
			coverage_status TEXT NOT NULL,
			catalog_claim_eligible INTEGER NOT NULL DEFAULT 0,
			catalog_schema TEXT NOT NULL,
			catalog_version TEXT NOT NULL,
			catalog_sha256 TEXT NOT NULL,
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
		(domain, book, author, edition, volume, chapter, page, locator, source_path, title,
		 artifact_path, artifact_sha256, document_sha256, source_tier, verification_status,
		 artifact_kind, provenance_status, independence_status, coverage_status, catalog_claim_eligible,
		 catalog_schema, catalog_version, catalog_sha256, content, is_index, document_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		meta.Domain, meta.Book, meta.Author, meta.Edition, meta.Volume, meta.Chapter, meta.Page, meta.Locator,
		meta.SourcePath, meta.Title, meta.ArtifactPath, meta.ArtifactSHA256, meta.DocumentSHA256, meta.SourceTier,
		meta.VerificationStatus, meta.ArtifactKind, meta.ProvenanceStatus, meta.IndependenceStatus, meta.CoverageStatus,
		boolToInt(meta.CatalogClaimEligible), meta.CatalogSchema, meta.CatalogVersion, meta.CatalogSHA256,
		content, isIndex, meta.SourcePath)
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

func metadataForPath(root, path string, catalog sourceCatalog) (documentMetadata, error) {
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
	meta := documentMetadata{
		Domain:               "bazi",
		Book:                 book,
		Author:               "unrecorded",
		Edition:              "unrecorded",
		Chapter:              chapter,
		Locator:              "chapter:" + chapter,
		SourcePath:           "bazi/" + rel,
		Title:                chapter,
		SourceTier:           catalog.DefaultPolicy.SourceTier,
		VerificationStatus:   catalog.DefaultPolicy.VerificationStatus,
		ArtifactKind:         catalog.DefaultPolicy.ArtifactKind,
		ProvenanceStatus:     catalog.DefaultPolicy.ProvenanceStatus,
		IndependenceStatus:   catalog.DefaultPolicy.IndependenceStatus,
		CoverageStatus:       catalog.DefaultPolicy.CoverageStatus,
		CatalogClaimEligible: catalog.DefaultPolicy.ClaimEligible,
		CatalogSchema:        catalog.Schema,
		CatalogVersion:       catalog.Version,
		CatalogSHA256:        catalog.sha256,
		IsIndex:              strings.EqualFold(fileName, "000.md"),
	}
	if source, ok := catalog.byRoot[book]; ok {
		meta.Book = source.Book
		meta.Author = source.Author
		meta.Edition = source.Edition
		meta.ArtifactPath = source.ArtifactPath
		meta.ArtifactSHA256 = source.ArtifactSHA256
		meta.SourceTier = source.SourceTier
		meta.VerificationStatus = source.VerificationStatus
		meta.ArtifactKind = source.ArtifactKind
		meta.ProvenanceStatus = source.ProvenanceStatus
		meta.IndependenceStatus = source.IndependenceStatus
		meta.CoverageStatus = source.CoverageStatus
		meta.CatalogClaimEligible = source.ClaimEligible
	}
	return meta, nil
}

func loadSourceCatalog(path string) (sourceCatalog, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return sourceCatalog{}, errors.New("catalog path is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return sourceCatalog{}, err
	}
	var catalog sourceCatalog
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&catalog); err != nil {
		return sourceCatalog{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return sourceCatalog{}, errors.New("source catalog must contain exactly one JSON document")
	}
	catalogSum := sha256.Sum256(raw)
	catalog.sha256 = hex.EncodeToString(catalogSum[:])
	if catalog.Schema != "bazi_rag_source_catalog_v1" || strings.TrimSpace(catalog.Version) == "" ||
		catalog.DefaultPolicy.SourceTier != "bronze_unverified" ||
		catalog.DefaultPolicy.VerificationStatus != "source_catalog_missing" ||
		catalog.DefaultPolicy.ArtifactKind != "unregistered" ||
		catalog.DefaultPolicy.ProvenanceStatus != "source_catalog_missing" ||
		catalog.DefaultPolicy.IndependenceStatus != "unknown" || catalog.DefaultPolicy.ClaimEligible ||
		catalog.DefaultPolicy.CoverageStatus != "unknown" ||
		len(catalog.Sources) == 0 {
		return sourceCatalog{}, errors.New("invalid source catalog identity or default policy")
	}
	catalog.byRoot = make(map[string]sourceCatalogEntry, len(catalog.Sources))
	repositoryRoot := filepath.Clean(filepath.Join(filepath.Dir(path), "..", ".."))
	for _, source := range catalog.Sources {
		if strings.TrimSpace(source.Book) == "" || strings.TrimSpace(source.MarkdownRoot) == "" ||
			source.Book != source.MarkdownRoot || filepath.Base(source.MarkdownRoot) != source.MarkdownRoot ||
			strings.TrimSpace(source.Author) == "" || strings.TrimSpace(source.Edition) == "" ||
			strings.TrimSpace(source.ArtifactPath) == "" || !validSourceSHA256(source.ArtifactSHA256) ||
			source.SourceTier != "classical_text_local" || source.PageMappingStatus == "" ||
			source.VerificationStatus == "" || source.ArtifactKind == "" || source.ProvenanceStatus == "" ||
			source.IndependenceStatus == "" || source.LicenseScope == "" || source.ClaimEligible || len(source.Limitations) == 0 ||
			source.CoverageStatus == "" ||
			!validCatalogProvenance(source) {
			return sourceCatalog{}, fmt.Errorf("invalid source catalog entry %q", source.Book)
		}
		if _, exists := catalog.byRoot[source.MarkdownRoot]; exists {
			return sourceCatalog{}, fmt.Errorf("duplicate source catalog root %q", source.MarkdownRoot)
		}
		artifactPath := filepath.Clean(source.ArtifactPath)
		if filepath.IsAbs(artifactPath) || artifactPath == "." || artifactPath == ".." ||
			strings.HasPrefix(artifactPath, ".."+string(filepath.Separator)) {
			return sourceCatalog{}, fmt.Errorf("catalog artifact path escapes repository: %s", source.ArtifactPath)
		}
		artifact, err := os.ReadFile(filepath.Join(repositoryRoot, artifactPath))
		if err != nil {
			return sourceCatalog{}, fmt.Errorf("catalog artifact %s: %w", source.ArtifactPath, err)
		}
		sum := sha256.Sum256(artifact)
		if hex.EncodeToString(sum[:]) != source.ArtifactSHA256 {
			return sourceCatalog{}, fmt.Errorf("catalog artifact hash mismatch for %s", source.ArtifactPath)
		}
		catalog.byRoot[source.MarkdownRoot] = source
	}
	return catalog, nil
}

func validCatalogProvenance(source sourceCatalogEntry) bool {
	switch source.ArtifactKind {
	case "legacy_text_pdf":
		return source.ProvenanceStatus == "bibliographic_provenance_unverified" &&
			source.IndependenceStatus == "independence_from_markdown_unverified" &&
			source.CoverageStatus == "partial_pdf_volumes_10_12_missing" &&
			source.VerificationStatus == "artifact_hash_verified_bibliography_and_page_mapping_unavailable"
	case "chromium_web_export":
		return source.ProvenanceStatus == "same_corpus_web_export_detected" &&
			source.IndependenceStatus == "not_independent_from_markdown" &&
			source.CoverageStatus == "same_corpus_export_chapter_structure_observed" &&
			source.VerificationStatus == "artifact_hash_verified_same_corpus_export_not_independent"
	default:
		return false
	}
}

func validSourceSHA256(value string) bool {
	if len(value) != 64 || value != strings.ToLower(value) {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
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
