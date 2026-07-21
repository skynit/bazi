#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REGISTRY="${RAG_EXTERNAL_SOURCE_REGISTRY:-$REPO_ROOT/research/rag/bazi-external-source-candidates-v1.json}"
CANDIDATE_ID="sanming-siku-wikisource-12vol-v1"
OUTPUT="${1:-$REPO_ROOT/research/rag/snapshots/$CANDIDATE_ID}"
API="https://zh.wikisource.org/w/api.php"
REGISTRY_SHA256="758974aa243b415a23bddbde995a597d82b575afff05000cda536130375a4dd9"

for command in curl jq sha256sum wc; do
  if ! command -v "$command" >/dev/null; then
    echo "required command not found: $command" >&2
    exit 1
  fi
done

if [ "$(sha256sum "$REGISTRY" | awk '{print $1}')" != "$REGISTRY_SHA256" ]; then
  echo "external source registry hash mismatch: $REGISTRY" >&2
  exit 1
fi
if ! jq -e --arg id "$CANDIDATE_ID" '
  .schema == "bazi_external_source_candidate_registry_v1" and
  .status == "candidate_only_no_runtime_ingestion" and
  any(.candidates[];
    .candidate_id == $id and
    .provider == "zh.wikisource.org" and
    .gates.license_terms_resolved == true and
    .gates.runtime_ingestion_allowed == false and
    .gates.claim_eligible == false and
    (.volumes | length == 12) and
    all(.volumes[]; .retrieved == false and .local_sha256 == "")
  )
' "$REGISTRY" >/dev/null; then
  echo "Wikisource candidate is not eligible for a research-only snapshot" >&2
  exit 1
fi
if [ -e "$OUTPUT" ]; then
  echo "snapshot output already exists: $OUTPUT" >&2
  exit 1
fi

parent="$(dirname "$OUTPUT")"
mkdir -p "$parent"
tmp="$(mktemp -d "$parent/.sanming-wikisource-snapshot.XXXXXX")"
cleanup() {
  rm -rf -- "$tmp"
}
trap cleanup EXIT
mkdir -p "$tmp/volumes"
entries="$tmp/volume-manifest.jsonl"

jq -c --arg id "$CANDIDATE_ID" '.candidates[] | select(.candidate_id == $id) | .volumes[]' "$REGISTRY" |
while IFS= read -r expected; do
  volume="$(jq -r '.volume' <<<"$expected")"
  page_id="$(jq -r '.stable_id' <<<"$expected")"
  revision_id="$(jq -r '.stable_revision' <<<"$expected")"
  expected_title="$(jq -r '.title' <<<"$expected")"
  expected_timestamp="$(jq -r '.timestamp' <<<"$expected")"
  expected_size="$(jq -r '.size_bytes' <<<"$expected")"
  expected_sha1="$(jq -r '.remote_sha1' <<<"$expected")"
  expected_source_url="$(jq -r '.source_url' <<<"$expected")"
  file_name="$(printf 'volume-%02d.wikitext' "$volume")"
  response="$tmp/response-$volume.json"

  curl -fsSL --get "$API" \
    --user-agent 'bazi-research-snapshot/1.0 (local reproducibility audit; no runtime ingestion)' \
    --retry 5 \
    --retry-all-errors \
    --retry-delay 2 \
    --retry-max-time 120 \
    --data-urlencode 'action=query' \
    --data-urlencode 'format=json' \
    --data-urlencode 'formatversion=2' \
    --data-urlencode "revids=$revision_id" \
    --data-urlencode 'prop=revisions' \
    --data-urlencode 'rvprop=ids|timestamp|sha1|size|content' \
    --data-urlencode 'rvslots=main' \
    -o "$response"

  if ! jq -e \
    --arg title "$expected_title" \
    --argjson page_id "$page_id" \
    --argjson revision_id "$revision_id" \
    --arg timestamp "$expected_timestamp" \
    --argjson size "$expected_size" \
    --arg sha1 "$expected_sha1" '
      (.query.pages | length) == 1 and
      (.query.pages[0] as $page |
        $page.title == $title and
        $page.pageid == $page_id and
        ($page.revisions | length) == 1 and
        ($page.revisions[0] as $revision |
          $revision.revid == $revision_id and
          $revision.timestamp == $timestamp and
          $revision.size == $size and
          $revision.sha1 == $sha1 and
          $revision.slots.main.contentmodel == "wikitext"
        )
      )
    ' "$response" >/dev/null; then
    echo "Wikisource revision identity mismatch for volume $volume" >&2
    exit 1
  fi

  jq -j '.query.pages[0].revisions[0].slots.main.content' "$response" >"$tmp/volumes/$file_name"
  local_size="$(wc -c <"$tmp/volumes/$file_name" | tr -d ' ')"
  if [ "$local_size" != "$expected_size" ]; then
    echo "Wikisource content size mismatch for volume $volume" >&2
    exit 1
  fi
  local_sha256="$(sha256sum "$tmp/volumes/$file_name" | awk '{print $1}')"
  jq -cn \
    --argjson volume "$volume" \
    --arg title "$expected_title" \
    --arg page_id "$page_id" \
    --argjson revision_id "$revision_id" \
    --arg timestamp "$expected_timestamp" \
    --argjson remote_size "$expected_size" \
    --arg remote_sha1 "$expected_sha1" \
    --arg source_url "$expected_source_url" \
    --arg artifact_path "volumes/$file_name" \
    --argjson local_size "$local_size" \
    --arg local_sha256 "$local_sha256" \
    '{volume:$volume,title:$title,page_id:$page_id,revision_id:$revision_id,timestamp:$timestamp,remote_size:$remote_size,remote_sha1:$remote_sha1,source_url:$source_url,artifact_path:$artifact_path,local_size:$local_size,local_sha256:$local_sha256}' \
    >>"$entries"
done

jq -s \
  --arg registry_path "research/rag/bazi-external-source-candidates-v1.json" \
  --arg registry_sha256 "$REGISTRY_SHA256" \
  --arg candidate_id "$CANDIDATE_ID" \
  '{schema:"wikisource_revision_snapshot_v1",version:"2026-07-17.1",status:"research_snapshot_not_runtime_eligible",retrieved_at:"2026-07-17",registry_path:$registry_path,registry_sha256:$registry_sha256,candidate_id:$candidate_id,provider:"zh.wikisource.org",work:"三命通會",edition:"欽定四庫全書本",license:{underlying_work:"PD-old",digital_contributions:"CC BY-SA 4.0 and GFDL",license_url:"https://zh.wikisource.org/wiki/Wikisource:版权信息",attribution_file:"ATTRIBUTION.md"},boundaries:{raw_wikitext_unmodified:true,local_artifact_frozen:true,independence_verified:false,bibliography_adjudicated:false,page_mapping_verified:false,claim_support_reviewed:false,runtime_ingestion_allowed:false,claim_eligible:false,publishable_accuracy:false},volumes:.}' \
  "$entries" >"$tmp/snapshot-manifest.json"

cat >"$tmp/ATTRIBUTION.md" <<'ATTRIBUTION'
# Attribution

This research-only snapshot contains fixed revisions of `三命通會 (四庫全書本)` from Chinese Wikisource.

- Source: https://zh.wikisource.org/wiki/三命通會_(四庫全書本)
- Underlying work marker: `PD-old`
- Digital contributions: CC BY-SA 4.0 and GFDL
- License information: https://zh.wikisource.org/wiki/Wikisource:版权信息
- Revision identity and per-file source URLs: `snapshot-manifest.json`

The raw wikitext is stored without modification. This snapshot is for internal textual comparison only. It is not registered in the runtime RAG index and is not claim-eligible.
ATTRIBUTION

rm -f "$entries" "$tmp"/response-*.json
mv "$tmp" "$OUTPUT"
trap - EXIT
echo "Wikisource research snapshot ready: $OUTPUT"
