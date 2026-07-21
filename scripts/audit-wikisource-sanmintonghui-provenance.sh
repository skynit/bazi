#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SNAPSHOT="${RAG_WIKISOURCE_SNAPSHOT:-$REPO_ROOT/research/rag/snapshots/sanming-siku-wikisource-12vol-v1}"
OUTPUT="${1:-$REPO_ROOT/research/rag/provenance/sanming-siku-wikisource-provenance-v1}"
API="https://zh.wikisource.org/w/api.php"
USER_AGENT="bazi-research-provenance/1.0 (local audit; no runtime ingestion)"
SNAPSHOT_MANIFEST_SHA256="38a03d0ee048a097620de3765f2793e9f0a5383a3238748fcd583ac79bb26974"
COMPARISON_SHA256="ce7ee55bd4c6b68c488f3326b9b3725860b5ebea0c679ee12d156f330a180b7f"

for command in curl jq rg sha256sum wc; do
  if ! command -v "$command" >/dev/null; then
    echo "required command not found: $command" >&2
    exit 1
  fi
done

if [ "$(sha256sum "$SNAPSHOT/snapshot-manifest.json" | awk '{print $1}')" != "$SNAPSHOT_MANIFEST_SHA256" ]; then
  echo "snapshot manifest hash mismatch: $SNAPSHOT" >&2
  exit 1
fi
if [ "$(sha256sum "$REPO_ROOT/research/rag/sanming-wikisource-markdown-comparison-v1.json" | awk '{print $1}')" != "$COMPARISON_SHA256" ]; then
  echo "cross-source comparison hash mismatch" >&2
  exit 1
fi
if [ -e "$OUTPUT" ]; then
  echo "provenance output already exists: $OUTPUT" >&2
  exit 1
fi

parent="$(dirname "$OUTPUT")"
mkdir -p "$parent"
tmp="$(mktemp -d "$parent/.sanming-wikisource-provenance.XXXXXX")"
cleanup() {
  rm -rf -- "$tmp"
}
trap cleanup EXIT
mkdir -p "$tmp/evidence"
fixed_entries="$tmp/fixed-evidence.jsonl"
history_entries="$tmp/revision-history.jsonl"
volume_entries="$tmp/volume-observations.jsonl"

api_query() {
  curl -fsSL --get "$API" \
    --user-agent "$USER_AGENT" \
    --retry 5 \
    --retry-all-errors \
    --retry-delay 2 \
    --retry-max-time 120 \
    "$@"
}

fetch_fixed_evidence() {
  local role="$1"
  local page_id="$2"
  local revision_id="$3"
  local title="$4"
  local timestamp="$5"
  local user="$6"
  local remote_size="$7"
  local remote_sha1="$8"
  local file_name="$9"
  local response="$tmp/fixed-$revision_id.json"

  api_query \
    --data-urlencode 'action=query' \
    --data-urlencode 'format=json' \
    --data-urlencode 'formatversion=2' \
    --data-urlencode "revids=$revision_id" \
    --data-urlencode 'prop=revisions' \
    --data-urlencode 'rvprop=ids|timestamp|user|userid|comment|sha1|size|content|contentmodel|flags|tags' \
    --data-urlencode 'rvslots=main' \
    -o "$response"

  if ! jq -e \
    --arg title "$title" \
    --arg user "$user" \
    --arg timestamp "$timestamp" \
    --arg sha1 "$remote_sha1" \
    --argjson page_id "$page_id" \
    --argjson revision_id "$revision_id" \
    --argjson size "$remote_size" '
      (.query.pages | length) == 1 and
      (.query.pages[0] as $page |
        $page.pageid == $page_id and
        $page.title == $title and
        ($page.revisions | length) == 1 and
        ($page.revisions[0] as $revision |
          $revision.revid == $revision_id and
          $revision.timestamp == $timestamp and
          $revision.user == $user and
          $revision.size == $size and
          $revision.sha1 == $sha1 and
          $revision.slots.main.contentmodel == "wikitext"
        )
      )
    ' "$response" >/dev/null; then
    echo "fixed evidence identity mismatch: revision $revision_id" >&2
    exit 1
  fi

  jq -j '.query.pages[0].revisions[0].slots.main.content' "$response" >"$tmp/evidence/$file_name"
  local local_size
  local local_sha256
  local_size="$(wc -c <"$tmp/evidence/$file_name" | tr -d ' ')"
  local_sha256="$(sha256sum "$tmp/evidence/$file_name" | awk '{print $1}')"
  if [ "$local_size" != "$remote_size" ]; then
    echo "fixed evidence size mismatch: revision $revision_id" >&2
    exit 1
  fi

  jq -c \
    --arg role "$role" \
    --arg title "$title" \
    --arg page_id "$page_id" \
    --argjson revision_id "$revision_id" \
    --arg timestamp "$timestamp" \
    --arg user "$user" \
    --argjson remote_size "$remote_size" \
    --arg remote_sha1 "$remote_sha1" \
    --arg source_url "https://zh.wikisource.org/w/index.php?curid=$page_id&oldid=$revision_id" \
    --arg artifact_path "evidence/$file_name" \
    --argjson local_size "$local_size" \
    --arg local_sha256 "$local_sha256" '
      .query.pages[0].revisions[0] |
      {
        role:$role,
        title:$title,
        page_id:$page_id,
        revision_id:$revision_id,
        parent_id:.parentid,
        timestamp:$timestamp,
        user:$user,
        user_id:.userid,
        comment:.comment,
        remote_size:$remote_size,
        remote_sha1:$remote_sha1,
        source_url:$source_url,
        artifact_path:$artifact_path,
        local_size:$local_size,
        local_sha256:$local_sha256
      }
    ' "$response" >>"$fixed_entries"
}

fetch_history() {
  local role="$1"
  local volume="$2"
  local page_id="$3"
  local revision_id="$4"
  local title="$5"
  local remote_size="$6"
  local remote_sha1="$7"
  local response="$tmp/history-$page_id.json"

  api_query \
    --data-urlencode 'action=query' \
    --data-urlencode 'format=json' \
    --data-urlencode 'formatversion=2' \
    --data-urlencode "pageids=$page_id" \
    --data-urlencode 'prop=revisions' \
    --data-urlencode 'rvlimit=max' \
    --data-urlencode 'rvdir=newer' \
    --data-urlencode "rvendid=$revision_id" \
    --data-urlencode 'rvprop=ids|timestamp|user|userid|comment|sha1|size|flags|tags' \
    -o "$response"

  if ! jq -e \
    --arg title "$title" \
    --arg sha1 "$remote_sha1" \
    --argjson page_id "$page_id" \
    --argjson revision_id "$revision_id" \
    --argjson size "$remote_size" '
      (has("continue") | not) and
      (.query.pages | length) == 1 and
      (.query.pages[0] as $page |
        $page.pageid == $page_id and
        $page.title == $title and
        ($page.revisions | length) > 0 and
        ($page.revisions[-1].revid == $revision_id) and
        ($page.revisions[-1].size == $size) and
        ($page.revisions[-1].sha1 == $sha1)
      )
    ' "$response" >/dev/null; then
    echo "revision history identity mismatch: page $page_id" >&2
    exit 1
  fi

  jq -c \
    --arg role "$role" \
    --argjson volume "$volume" \
    --arg page_id "$page_id" \
    --argjson revision_id "$revision_id" '
      .query.pages[0] |
      {
        role:$role,
        volume:$volume,
        page_id:$page_id,
        title:.title,
        fixed_revision_id:$revision_id,
        revision_count:(.revisions | length),
        contributors:([.revisions[].user] | unique),
        comments:([.revisions[].comment] | unique),
        content_hash_changed_since_first:(.revisions[0].sha1 != .revisions[-1].sha1),
        revisions:[.revisions[] | {
          revision_id:.revid,
          parent_id:.parentid,
          minor:(.minor // false),
          user:.user,
          user_id:.userid,
          timestamp:.timestamp,
          size:.size,
          sha1:.sha1,
          comment:.comment,
          tags:.tags
        }]
      }
    ' "$response" >>"$history_entries"
}

count_matches() {
  local pattern="$1"
  shift
  (rg -o "$pattern" "$@" || true) | wc -l | tr -d ' '
}

require_count() {
  local pattern="$1"
  local expected="$2"
  local file="$3"
  local actual
  actual="$(count_matches "$pattern" "$file")"
  if [ "$actual" != "$expected" ]; then
    echo "fixed evidence marker count for $pattern = $actual, want $expected" >&2
    exit 1
  fi
}

fetch_fixed_evidence \
  work_page 260853 657391 '三命通會 (四庫全書本)' \
  '2016-10-15T11:19:12Z' '維基小霸王' 2492 \
  '0fd68ff5e57220d21be67716a6294fdc6b0ddc63' work-page-rev657391.wikitext
fetch_fixed_evidence \
  import_guideline_active_at_bulk_import 205610 513112 'Wikisource:四庫全書原文' \
  '2016-09-21T12:28:20Z' '維基小霸王' 1522 \
  '27341a89cb4b7235381862c09317be4694ef4e1b' import-guideline-rev513112.wikitext
fetch_fixed_evidence \
  import_plan_active_at_bulk_import 204146 534259 'User:維基小霸王/录入四库全书计划' \
  '2016-10-03T13:37:37Z' '維基小霸王' 4564 \
  'b6c7b33884b7f09db30d367334e8707aa574743b' import-plan-rev534259.wikitext

work_evidence="$tmp/evidence/work-page-rev657391.wikitext"
guideline_evidence="$tmp/evidence/import-guideline-rev513112.wikitext"
plan_evidence="$tmp/evidence/import-plan-rev534259.wikitext"
require_count '請根據四庫全書掃描版校對本頁' 1 "$work_evidence"
require_count '導入的內容只包含文本，請根據掃描版本加入圖片' 1 "$guideline_evidence"
require_count '導入的內容只包含文本，請根據掃描版本加入表格' 1 "$guideline_evidence"
require_count '總共有3951個字符' 1 "$guideline_evidence"
require_count '网上看到了一个包含四库全书书籍的文本压缩包' 1 "$plan_evidence"
require_count '只有文本，沒有圖片和表格' 1 "$plan_evidence"
require_count '图片、表格可以用IA的扫描版添加' 1 "$plan_evidence"
require_count 'https://www.w3.org/International/articles/css3-text/' 1 "$plan_evidence"
require_count 'https://www.w3.org/TR/jlreq/#inline_cutting_note' 1 "$plan_evidence"
require_count 'https?://' 2 "$plan_evidence"
require_count 'archive\.org|identifier/[A-Za-z0-9._-]+' 0 "$plan_evidence"

fetch_history work_page 0 260853 657391 '三命通會 (四庫全書本)' 2492 \
  '0fd68ff5e57220d21be67716a6294fdc6b0ddc63'
jq -c '.volumes[]' "$SNAPSHOT/snapshot-manifest.json" |
while IFS= read -r volume; do
  fetch_history \
    volume \
    "$(jq -r '.volume' <<<"$volume")" \
    "$(jq -r '.page_id' <<<"$volume")" \
    "$(jq -r '.revision_id' <<<"$volume")" \
    "$(jq -r '.title' <<<"$volume")" \
    "$(jq -r '.remote_size' <<<"$volume")" \
    "$(jq -r '.remote_sha1' <<<"$volume")"
done

jq -s '{schema:"wikisource_fixed_revision_history_v1",version:"2026-07-17.1",status:"history_through_frozen_revision",pages:.}' \
  "$history_entries" >"$tmp/revision-history.json"

for volume in $(seq 1 12); do
  file_name="$(printf 'volume-%02d.wikitext' "$volume")"
  file="$SNAPSHOT/volumes/$file_name"
  jq -cn \
    --argjson volume "$volume" \
    --arg artifact_path "research/rag/snapshots/sanming-siku-wikisource-12vol-v1/volumes/$file_name" \
    --arg artifact_sha256 "$(sha256sum "$file" | awk '{print $1}')" \
    --argjson scan_notice_count "$(count_matches '請根據四庫全書掃描版校對本頁' "$file")" \
    --argjson skchar_count "$(count_matches '\{\{SKchar\|[0-9]+\}\}' "$file")" \
    --argjson direct_file_links "$(count_matches '\[\[(File|Image|文件|檔案):' "$file")" \
    --argjson direct_page_links "$(count_matches '\[\[(Page|頁):' "$file")" \
    --argjson direct_external_urls "$(count_matches 'https?://' "$file")" '
      {
        volume:$volume,
        artifact_path:$artifact_path,
        artifact_sha256:$artifact_sha256,
        scan_notice_count:$scan_notice_count,
        skchar_count:$skchar_count,
        direct_file_links:$direct_file_links,
        direct_page_links:$direct_page_links,
        direct_external_urls:$direct_external_urls
      }
    ' >>"$volume_entries"
done

history_sha256="$(sha256sum "$tmp/revision-history.json" | awk '{print $1}')"
history_pages="$(jq '.pages | length' "$tmp/revision-history.json")"
history_revisions="$(jq '[.pages[].revision_count] | add' "$tmp/revision-history.json")"
history_changed="$(jq '[.pages[] | select(.content_hash_changed_since_first)] | length' "$tmp/revision-history.json")"
fixed_importer="$(jq '[.pages[] | select(.revisions[-1].user == "維基小霸王")] | length' "$tmp/revision-history.json")"
fixed_wmr="$(jq '[.pages[] | select(.revisions[-1].user == "Wmr-bot")] | length' "$tmp/revision-history.json")"
fixed_crowley="$(jq '[.pages[] | select(.revisions[-1].user == "CrowleyBot")] | length' "$tmp/revision-history.json")"
first_importer="$(jq '[.pages[] | select(.revisions[0].user == "維基小霸王" and .revisions[0].user_id == 20996)] | length' "$tmp/revision-history.json")"
if [ "$history_pages" != 13 ] || [ "$history_revisions" != 61 ] || [ "$history_changed" != 8 ] || \
   [ "$first_importer" != 13 ] || [ "$fixed_importer" != 5 ] || [ "$fixed_wmr" != 6 ] || [ "$fixed_crowley" != 2 ]; then
  echo "unexpected fixed revision history summary" >&2
  exit 1
fi

jq -n \
  --slurpfile fixed "$fixed_entries" \
  --slurpfile volumes "$volume_entries" \
  --arg snapshot_manifest_path "research/rag/snapshots/sanming-siku-wikisource-12vol-v1/snapshot-manifest.json" \
  --arg snapshot_manifest_sha256 "$SNAPSHOT_MANIFEST_SHA256" \
  --arg comparison_path "research/rag/sanming-wikisource-markdown-comparison-v1.json" \
  --arg comparison_sha256 "$COMPARISON_SHA256" \
  --arg history_path "revision-history.json" \
  --arg history_sha256 "$history_sha256" \
  --argjson history_pages "$history_pages" \
  --argjson history_revisions "$history_revisions" \
  --argjson history_changed "$history_changed" \
  --argjson fixed_importer "$fixed_importer" \
  --argjson fixed_wmr "$fixed_wmr" \
  --argjson fixed_crowley "$fixed_crowley" \
  --argjson scan_notice_volumes "$(jq -s '[.[].scan_notice_count | select(. == 1)] | length' "$volume_entries")" \
  --argjson skchar_total "$(jq -s '[.[].skchar_count] | add' "$volume_entries")" \
  --argjson direct_file_links "$(jq -s '[.[].direct_file_links] | add' "$volume_entries")" \
  --argjson direct_page_links "$(jq -s '[.[].direct_page_links] | add' "$volume_entries")" \
  --argjson direct_external_urls "$(jq -s '[.[].direct_external_urls] | add' "$volume_entries")" '
  {
    schema:"sanming_wikisource_transcription_provenance_audit_v1",
    version:"2026-07-17.1",
    status:"bulk_imported_transcription_upstream_archive_unidentified",
    observed_at:"2026-07-17",
    candidate_id:"sanming-siku-wikisource-12vol-v1",
    purpose:"Determine whether the frozen Wikisource transcription can be promoted to an independent, scan-addressable primary artifact.",
    inputs:{
      snapshot_manifest_path:$snapshot_manifest_path,
      snapshot_manifest_sha256:$snapshot_manifest_sha256,
      cross_source_comparison_path:$comparison_path,
      cross_source_comparison_sha256:$comparison_sha256,
      revision_history_path:$history_path,
      revision_history_sha256:$history_sha256
    },
    method:{
      fixed_revision_api:"MediaWiki action=query revids plus revisions content",
      history_api:"MediaWiki action=query single page rvdir=newer rvendid=frozen_revision",
      local_marker_scan:"ripgrep exact patterns over frozen raw wikitext",
      excludes_current_template_render:true,
      excludes_unlicensed_scan_downloads:true
    },
    license:{
      underlying_work:"PD-old",
      digital_contributions:"CC BY-SA 4.0 and GFDL",
      license_url:"https://zh.wikisource.org/wiki/Wikisource:版权信息",
      attribution_required:true
    },
    fixed_evidence:$fixed,
    history_summary:{
      pages:$history_pages,
      revisions:$history_revisions,
      first_revision_importer:"維基小霸王",
      first_revision_importer_user_id:20996,
      pages_with_content_hash_changed_since_first:$history_changed,
      fixed_revision_by_importer:$fixed_importer,
      fixed_revision_by_wmr_bot:$fixed_wmr,
      fixed_revision_by_crowley_bot:$fixed_crowley,
      fixed_comment_classes:["bulk_import","template_repair","single_text_correction"],
      fixed_comment_examples:["导入1个版本","repair templates","辛西→辛酉（[[Special:permalink/2082153|本批详情]]）"]
    },
    import_provenance_observation:{
      upstream_description:"网上看到的一个包含四库全书书籍的文本压缩包",
      upstream_url:"",
      upstream_identifier:"",
      upstream_hash:"",
      scan_edition_identifier:"",
      library_catalog_identifier:"",
      plan_direct_external_urls:[
        "https://www.w3.org/International/articles/css3-text/",
        "https://www.w3.org/TR/jlreq/#inline_cutting_note"
      ],
      plan_upstream_archive_url_count:0,
      plan_states_only_text_without_images_or_tables:true,
      plan_states_ia_scans_are_future_manual_inputs:true,
      guideline_states_images_and_tables_require_future_scan_comparison:true,
      guideline_states_3951_rare_characters_require_manual_identification:true
    },
    frozen_volume_observations:$volumes,
    aggregate_content_markers:{
      volumes_with_scan_comparison_notice:$scan_notice_volumes,
      unresolved_skchar_placeholders:$skchar_total,
      direct_file_links:$direct_file_links,
      direct_page_namespace_links:$direct_page_links,
      direct_external_urls:$direct_external_urls
    },
    decision:{
      artifact_kind:"bulk_imported_plain_text_transcription",
      provenance_status:"upstream_text_compressed_archive_unidentified",
      independence_status:"not_adjudicated_against_local_markdown",
      scan_identity_status:"no_scan_identity_or_page_links_in_frozen_pages",
      scan_proofreading_status:"not_scan_proofread",
      coverage_status:"complete_twelve_volume_text_structure_observed",
      promotion_result:"blocked"
    },
    boundaries:{
      fixed_revision_identity_verified:true,
      complete_text_structure_observed:true,
      bibliographic_provenance_verified:false,
      independent_primary_artifact_verified:false,
      scan_artifact_verified:false,
      page_mapping_verified:false,
      claim_support_reviewed:false,
      runtime_ingestion_allowed:false,
      claim_eligible:false,
      publishable_accuracy:false
    },
    limitations:[
      "A versioned transcription is not a published scan or publisher digital edition.",
      "The import plan names an online compressed text archive but records no URL, identifier, hash, institution, or edition.",
      "The import guideline says images, tables, dates, links, and rare characters still require manual work against scans.",
      "Textual overlap with the local Markdown corpus does not establish independence, page identity, or claim support.",
      "Internet Archive scan candidates remain excluded because their provider rights and license metadata are unresolved."
    ]
  }
' >"$tmp/provenance-audit.json"

rm -f "$fixed_entries" "$history_entries" "$volume_entries" "$tmp"/fixed-*.json "$tmp"/history-*.json
mv "$tmp" "$OUTPUT"
trap - EXIT
echo "Wikisource transcription provenance audit ready: $OUTPUT"
