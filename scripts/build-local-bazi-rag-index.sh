#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SOURCE_DIR="${1:-${LOCAL_RAG_SOURCE_DIR:-/home/skynit/mingli_db/md/bazi}}"
INDEX_PATH="${2:-${LOCAL_RAG_INDEX_PATH:-$ROOT/data/bazi_fts.db}}"

mkdir -p "$(dirname "$INDEX_PATH")"

(
  cd "$ROOT/src"
  go run -tags sqlite_fts5 ./cmd/bazi-rag-index \
    -source "$SOURCE_DIR" \
    -index "$INDEX_PATH"
)

echo "Local BaZi RAG index ready: $INDEX_PATH"
