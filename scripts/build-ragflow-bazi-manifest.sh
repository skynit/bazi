#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-/home/skynit/mingli_db/md/bazi}"
KEEP_INDEX="${KEEP_INDEX:-0}"

if [ ! -d "$ROOT" ]; then
  echo "source root not found: $ROOT" >&2
  exit 1
fi

find "$ROOT" -type f \( -name '*.md' -o -name '*.MD' \) | sort | while read -r file; do
  rel="${file#"$ROOT"/}"
  if [[ "$rel" == *.obsidian/* ]] || [[ "$rel" == .obsidian/* ]]; then
    continue
  fi

  book="${rel%%/*}"
  chapter="${rel#*/}"
  title="${chapter%.md}"
  source_path="bazi/${book}/${chapter}"
  is_index=false
  if [[ "$chapter" == "000.md" ]]; then
    if [ "$KEEP_INDEX" != "1" ]; then
      continue
    fi
    is_index=true
  fi

  printf '%s\n' "$(python3 - <<'PY' "$book" "$chapter" "$source_path" "$title" "$is_index"
import json, sys
book, chapter, source_path, title, is_index = sys.argv[1:]
print(json.dumps({
    "domain": "bazi",
    "book": book,
    "chapter": chapter[:-3] if chapter.endswith(".md") else chapter,
    "source_path": source_path,
    "title": title,
    "is_index": is_index == "true",
}, ensure_ascii=False))
PY
)"
done
