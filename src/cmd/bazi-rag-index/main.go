package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"bazi/internal/service/localrag"
)

func main() {
	sourceDir := flag.String("source", "/home/skynit/mingli_db/md/bazi", "BaZi Markdown source directory")
	indexPath := flag.String("index", "./data/bazi_fts.db", "SQLite FTS5 index path")
	catalogPath := flag.String("catalog", "../research/rag/bazi-source-catalog-v1.json", "audited BaZi source catalog path")
	keepIndex := flag.Bool("keep-index", false, "keep 000.md index pages")
	chunkSize := flag.Int("chunk-size", 900, "maximum chunk size in runes")
	overlap := flag.Int("overlap", 120, "chunk overlap in runes")
	flag.Parse()

	started := time.Now()
	stats, err := localrag.BuildIndex(context.Background(), localrag.IndexConfig{
		SourceDir:   *sourceDir,
		IndexPath:   *indexPath,
		CatalogPath: *catalogPath,
		KeepIndex:   *keepIndex,
		ChunkSize:   *chunkSize,
		Overlap:     *overlap,
	})
	if err != nil {
		log.Fatalf("build local BaZi RAG index failed: %v", err)
	}

	fmt.Printf("Built local BaZi RAG index: documents=%d chunks=%d skipped=%d path=%s elapsed=%s\n",
		stats.Documents, stats.Chunks, stats.Skipped, *indexPath, time.Since(started).Round(time.Millisecond))
}
