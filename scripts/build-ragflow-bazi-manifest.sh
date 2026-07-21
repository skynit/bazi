#!/usr/bin/env bash
set -euo pipefail

ROOT="${1:-/home/skynit/mingli_db/md/bazi}"
KEEP_INDEX="${KEEP_INDEX:-0}"
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CATALOG="${RAG_SOURCE_CATALOG_PATH:-$REPO_ROOT/research/rag/bazi-source-catalog-v1.json}"
EXPECTED_CATALOG_SHA256="${RAG_SOURCE_CATALOG_SHA256:-5a012c68833eaa1163a175f579833cca2337de2d7b22eeb2ec9ba396038059d5}"

if [ ! -d "$ROOT" ]; then
  echo "source root not found: $ROOT" >&2
  exit 1
fi
if ! jq -e '
  .schema == "bazi_rag_source_catalog_v1" and
  .version == "2026-07-17.3" and
  .default_policy == {
    source_tier:"bronze_unverified",
    verification_status:"source_catalog_missing",
    artifact_kind:"unregistered",
    provenance_status:"source_catalog_missing",
    independence_status:"unknown",
    coverage_status:"unknown",
    claim_eligible:false
  } and
  (.sources | length == 4) and
  all(.sources[];
    .claim_eligible == false and
    .source_tier == "classical_text_local" and
    (.artifact_path | startswith("library/") and (contains("..") | not)) and
    (.artifact_sha256 | test("^[0-9a-f]{64}$")) and
    (.artifact_kind == "legacy_text_pdf" or .artifact_kind == "chromium_web_export") and
    (.provenance_status | type == "string" and length > 0) and
    (.independence_status | type == "string" and length > 0)
    and (.coverage_status | type == "string" and length > 0)
  )
' "$CATALOG" >/dev/null; then
  echo "invalid source catalog: $CATALOG" >&2
  exit 1
fi

catalog_schema="$(jq -r '.schema' "$CATALOG")"
catalog_version="$(jq -r '.version' "$CATALOG")"
catalog_sha256="$(sha256sum "$CATALOG" | awk '{print $1}')"
if [ "$catalog_sha256" != "$EXPECTED_CATALOG_SHA256" ]; then
  echo "source catalog hash mismatch: $CATALOG" >&2
  exit 1
fi
default_source_tier="$(jq -r '.default_policy.source_tier' "$CATALOG")"
default_verification_status="$(jq -r '.default_policy.verification_status' "$CATALOG")"
default_artifact_kind="$(jq -r '.default_policy.artifact_kind' "$CATALOG")"
default_provenance_status="$(jq -r '.default_policy.provenance_status' "$CATALOG")"
default_independence_status="$(jq -r '.default_policy.independence_status' "$CATALOG")"
default_coverage_status="$(jq -r '.default_policy.coverage_status' "$CATALOG")"
while IFS=$'\t' read -r book artifact_path expected_sha256; do
  actual_sha256="$(sha256sum "$REPO_ROOT/$artifact_path" | awk '{print $1}')"
  if [ "$actual_sha256" != "$expected_sha256" ]; then
    echo "source catalog artifact hash mismatch: $book ($artifact_path)" >&2
    exit 1
  fi
done < <(jq -r '.sources[] | [.book,.artifact_path,.artifact_sha256] | @tsv' "$CATALOG")

find "$ROOT" -type f \( -name '*.md' -o -name '*.MD' \) | sort | while read -r file; do
  rel="${file#"$ROOT"/}"
  if [[ "$rel" == *.obsidian/* ]] || [[ "$rel" == .obsidian/* ]]; then
    continue
  fi

  book="${rel%%/*}"
  chapter="${rel#*/}"
  title="${chapter%.md}"
  source_path="bazi/${book}/${chapter}"
  chapter_id="${chapter%.md}"
  document_sha256="$(sha256sum "$file" | awk '{print $1}')"
  is_index=false
  if [[ "$chapter" == "000.md" ]]; then
    if [ "$KEEP_INDEX" != "1" ]; then
      continue
    fi
    is_index=true
  fi

  source="$(jq -c --arg root "$book" '.sources[] | select(.markdown_root == $root)' "$CATALOG" | head -n 1)"
  author="unrecorded"
  edition="unrecorded"
  artifact_path=""
  artifact_sha256=""
  source_tier="$default_source_tier"
  verification_status="$default_verification_status"
  artifact_kind="$default_artifact_kind"
  provenance_status="$default_provenance_status"
  independence_status="$default_independence_status"
  coverage_status="$default_coverage_status"
  catalog_claim_eligible=false
  if [ -n "$source" ]; then
    author="$(jq -r '.author' <<<"$source")"
    edition="$(jq -r '.edition' <<<"$source")"
    artifact_path="$(jq -r '.artifact_path' <<<"$source")"
    artifact_sha256="$(jq -r '.artifact_sha256' <<<"$source")"
    source_tier="$(jq -r '.source_tier' <<<"$source")"
    verification_status="$(jq -r '.verification_status' <<<"$source")"
    artifact_kind="$(jq -r '.artifact_kind' <<<"$source")"
    provenance_status="$(jq -r '.provenance_status' <<<"$source")"
    independence_status="$(jq -r '.independence_status' <<<"$source")"
    coverage_status="$(jq -r '.coverage_status' <<<"$source")"
    catalog_claim_eligible="$(jq -r '.claim_eligible' <<<"$source")"
  fi

  jq -cn \
    --arg book "$book" \
    --arg author "$author" \
    --arg edition "$edition" \
    --arg chapter "$chapter_id" \
    --arg locator "chapter:$chapter_id" \
    --arg source_path "$source_path" \
    --arg title "$title" \
    --arg artifact_path "$artifact_path" \
    --arg artifact_sha256 "$artifact_sha256" \
    --arg document_sha256 "$document_sha256" \
    --arg source_tier "$source_tier" \
    --arg verification_status "$verification_status" \
    --arg artifact_kind "$artifact_kind" \
    --arg provenance_status "$provenance_status" \
    --arg independence_status "$independence_status" \
    --arg coverage_status "$coverage_status" \
    --arg catalog_schema "$catalog_schema" \
    --arg catalog_version "$catalog_version" \
    --arg catalog_sha256 "$catalog_sha256" \
    --argjson catalog_claim_eligible "$catalog_claim_eligible" \
    --argjson is_index "$is_index" \
    '{domain:"bazi",book:$book,author:$author,edition:$edition,volume:"",chapter:$chapter,page:"",locator:$locator,source_path:$source_path,title:$title,artifact_path:$artifact_path,artifact_sha256:$artifact_sha256,document_sha256:$document_sha256,source_tier:$source_tier,verification_status:$verification_status,artifact_kind:$artifact_kind,provenance_status:$provenance_status,independence_status:$independence_status,coverage_status:$coverage_status,catalog_claim_eligible:$catalog_claim_eligible,catalog_schema:$catalog_schema,catalog_version:$catalog_version,catalog_sha256:$catalog_sha256,is_index:$is_index}'
done
