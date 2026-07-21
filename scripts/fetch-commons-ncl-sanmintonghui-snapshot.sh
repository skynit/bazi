#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
OUTPUT="${1:-$ROOT/research/rag/snapshots/sanming-ncl-06589-1578-v1}"
API="https://commons.wikimedia.org/w/api.php"
VERSION="2026-07-17.1"
RETRIEVED_AT="2026-07-17"

for command in curl jq sha1sum sha256sum wc pdfinfo; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "required command not found: $command" >&2
    exit 1
  fi
done

if [ -e "$OUTPUT" ]; then
  echo "snapshot output already exists: $OUTPUT" >&2
  exit 1
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/output/artifacts" "$tmp/output/evidence"

api_query() {
  curl --fail --silent --show-error --retry 3 --retry-all-errors --max-time 60 \
    --get "$API" "$@"
}

search_response="$tmp/output/evidence/commons-title-search.json"
api_query \
  --data-urlencode 'action=query' \
  --data-urlencode 'format=json' \
  --data-urlencode 'formatversion=2' \
  --data-urlencode 'list=search' \
  --data-urlencode 'srsearch=intitle:"三命通會"' \
  --data-urlencode 'srnamespace=6|14' \
  --data-urlencode 'srlimit=500' \
  -o "$search_response"

if ! jq -e '
  (.query.searchinfo.totalhits >= 2) and
  ([.query.search[].title] | index("File:NCL-06589 1 三命通會.pdf") != null) and
  ([.query.search[].title] | index("File:NCL-06589 2 三命通會.pdf") != null)
' "$search_response" >/dev/null; then
  echo "Commons title search does not contain both fixed NCL files" >&2
  exit 1
fi

entries="$tmp/entries.jsonl"
: >"$entries"

fetch_part() {
  local part="$1"
  local page_id="$2"
  local revision_id="$3"
  local parent_id="$4"
  local revision_timestamp="$5"
  local revision_size="$6"
  local revision_sha1="$7"
  local title="$8"
  local file_timestamp="$9"
  local remote_size="${10}"
  local page_count="${11}"
  local remote_sha1="${12}"
  local download_url="${13}"
  local response="$tmp/part-$part.json"
  local artifact_name="ncl-06589-part-$part.pdf"
  local evidence_name="commons-file-page-$part-rev$revision_id.wikitext"
  local artifact="$tmp/output/artifacts/$artifact_name"
  local evidence="$tmp/output/evidence/$evidence_name"

  api_query \
    --data-urlencode 'action=query' \
    --data-urlencode 'format=json' \
    --data-urlencode 'formatversion=2' \
    --data-urlencode "revids=$revision_id" \
    --data-urlencode 'prop=revisions|imageinfo|categories' \
    --data-urlencode 'rvprop=ids|timestamp|sha1|size|content|contentmodel' \
    --data-urlencode 'rvslots=main' \
    --data-urlencode 'iiprop=url|size|sha1|mime|timestamp|canonicaltitle|extmetadata' \
    --data-urlencode 'cllimit=max' \
    -o "$response"

  if ! jq -e \
    --arg title "$title" \
    --arg revision_timestamp "$revision_timestamp" \
    --arg revision_sha1 "$revision_sha1" \
    --arg file_timestamp "$file_timestamp" \
    --arg remote_sha1 "$remote_sha1" \
    --arg download_url "$download_url" \
    --argjson page_id "$page_id" \
    --argjson revision_id "$revision_id" \
    --argjson parent_id "$parent_id" \
    --argjson revision_size "$revision_size" \
    --argjson remote_size "$remote_size" \
    --argjson page_count "$page_count" '
      (has("continue") | not) and
      (.query.pages | length) == 1 and
      (.query.pages[0] as $page |
        $page.pageid == $page_id and
        $page.title == $title and
        ($page.revisions | length) == 1 and
        ($page.revisions[0] as $revision |
          $revision.revid == $revision_id and
          $revision.parentid == $parent_id and
          $revision.timestamp == $revision_timestamp and
          $revision.size == $revision_size and
          $revision.sha1 == $revision_sha1 and
          $revision.slots.main.contentmodel == "wikitext" and
          ($revision.slots.main.content | contains("明萬曆戊寅(六年, 1578)刊本")) and
          ($revision.slots.main.content | contains("卷數：十二卷")) and
          ($revision.slots.main.content | contains("數量：12冊")) and
          ($revision.slots.main.content | contains("索書號：306.5 06589")) and
          ($revision.slots.main.content | contains("{{PD-scan|PD-old}}"))
        ) and
        ($page.imageinfo | length) == 1 and
        ($page.imageinfo[0] as $image |
          $image.timestamp == $file_timestamp and
          $image.size == $remote_size and
          $image.pagecount == $page_count and
          $image.sha1 == $remote_sha1 and
          $image.mime == "application/pdf" and
          $image.url == $download_url and
          $image.canonicaltitle == $title and
          $image.extmetadata.License.value == "pd" and
          $image.extmetadata.LicenseShortName.value == "Public domain" and
          $image.extmetadata.UsageTerms.value == "Public domain" and
          $image.extmetadata.AttributionRequired.value == "false" and
          $image.extmetadata.Copyrighted.value == "False"
        ) and
        ([$page.categories[].title] | index("Category:PD-scan (PD-old)") != null) and
        ([$page.categories[].title] | index("Category:Scans from the Rare Books & Special Collections, National Central Library") != null)
      )
    ' "$response" >/dev/null; then
    echo "Commons identity, bibliography, or license mismatch for part $part" >&2
    exit 1
  fi

  jq -j '.query.pages[0].revisions[0].slots.main.content' "$response" >"$evidence"
  curl --fail --location --silent --show-error --retry 5 --retry-all-errors \
    --max-time 900 "$download_url" -o "$artifact"

  local local_size
  local local_sha1
  local local_sha256
  local local_pages
  local evidence_size
  local evidence_sha256
  local_size="$(wc -c <"$artifact" | tr -d ' ')"
  local_sha1="$(sha1sum "$artifact" | awk '{print $1}')"
  local_sha256="$(sha256sum "$artifact" | awk '{print $1}')"
  local_pages="$(pdfinfo "$artifact" | awk -F: '$1 == "Pages" {gsub(/[[:space:]]/, "", $2); print $2}')"
  evidence_size="$(wc -c <"$evidence" | tr -d ' ')"
  evidence_sha256="$(sha256sum "$evidence" | awk '{print $1}')"

  if [ "$local_size" != "$remote_size" ] || [ "$local_sha1" != "$remote_sha1" ] || [ "$local_pages" != "$page_count" ]; then
    echo "downloaded PDF identity mismatch for part $part" >&2
    exit 1
  fi

  jq -n -c \
    --argjson part "$part" \
    --argjson page_id "$page_id" \
    --arg title "$title" \
    --argjson revision_id "$revision_id" \
    --argjson parent_id "$parent_id" \
    --arg revision_timestamp "$revision_timestamp" \
    --argjson revision_size "$revision_size" \
    --arg revision_sha1 "$revision_sha1" \
    --arg source_url "https://commons.wikimedia.org/w/index.php?curid=$page_id&oldid=$revision_id" \
    --arg file_timestamp "$file_timestamp" \
    --argjson remote_size "$remote_size" \
    --argjson page_count "$page_count" \
    --arg remote_sha1 "$remote_sha1" \
    --arg download_url "$download_url" \
    --arg description_url "https://commons.wikimedia.org/wiki/Special:Redirect/page/$page_id" \
    --arg page_locator_pattern "https://commons.wikimedia.org/wiki/Special:Redirect/page/$page_id?page={physical_page}" \
    --arg artifact_path "artifacts/$artifact_name" \
    --arg local_sha1 "$local_sha1" \
    --arg local_sha256 "$local_sha256" \
    --arg evidence_path "evidence/$evidence_name" \
    --argjson evidence_size "$evidence_size" \
    --arg evidence_sha256 "$evidence_sha256" '
      {
        part:$part,
        commons_page_id:$page_id,
        title:$title,
        fixed_file_page_revision:{
          revision_id:$revision_id,
          parent_id:$parent_id,
          timestamp:$revision_timestamp,
          size_bytes:$revision_size,
          sha1:$revision_sha1,
          source_url:$source_url
        },
        remote_file:{
          timestamp:$file_timestamp,
          size_bytes:$remote_size,
          page_count:$page_count,
          sha1:$remote_sha1,
          mime:"application/pdf",
          download_url:$download_url,
          description_url:$description_url,
          physical_page_range:{first:1,last:$page_count},
          page_locator_pattern:$page_locator_pattern
        },
        local_artifact:{
          path:$artifact_path,
          size_bytes:$remote_size,
          page_count:$page_count,
          sha1:$local_sha1,
          sha256:$local_sha256
        },
        license_evidence:{
          path:$evidence_path,
          size_bytes:$evidence_size,
          sha256:$evidence_sha256
        }
      }
    ' >>"$entries"
}

fetch_part \
  1 138125281 1207461339 1138624022 '2026-05-03T01:58:40Z' 5414 \
  a009fba1a2909653d5cc1f55a2e6d022f76d316f \
  'File:NCL-06589 1 三命通會.pdf' '2023-09-26T02:32:46Z' \
  101956385 1000 c222bc54815d8e5cef15338c03b9fc11d540f41a \
  'https://upload.wikimedia.org/wikipedia/commons/f/f4/NCL-06589_1_%E4%B8%89%E5%91%BD%E9%80%9A%E6%9C%83.pdf'

fetch_part \
  2 138043642 1207461333 804277976 '2026-05-03T01:58:39Z' 2208 \
  2c97f0f2ea3cf51ec861585e23b18698bc9e0304 \
  'File:NCL-06589 2 三命通會.pdf' '2023-09-24T12:54:27Z' \
  18967335 187 090ce9d53d9aa37bd7d2680b3b69c16385edbe2d \
  'https://upload.wikimedia.org/wikipedia/commons/f/fd/NCL-06589_2_%E4%B8%89%E5%91%BD%E9%80%9A%E6%9C%83.pdf'

search_sha256="$(sha256sum "$search_response" | awk '{print $1}')"
search_size="$(wc -c <"$search_response" | tr -d ' ')"
search_total="$(jq '.query.searchinfo.totalhits' "$search_response")"
search_returned="$(jq '.query.search | length' "$search_response")"
file_hits="$(jq '[.query.search[] | select(.ns == 6)] | length' "$search_response")"
category_hits="$(jq '[.query.search[] | select(.ns == 14)] | length' "$search_response")"

jq -n \
  --arg version "$VERSION" \
  --arg observed_at "$RETRIEVED_AT" \
  --arg response_sha256 "$search_sha256" \
  --argjson response_size "$search_size" \
  --argjson total_hits "$search_total" \
  --argjson returned_hits "$search_returned" \
  --argjson file_hits "$file_hits" \
  --argjson category_hits "$category_hits" '
    {
      schema:"sanming_complete_scan_discovery_audit_v1",
      version:$version,
      status:"selected_for_research_snapshot_not_runtime_ingestion",
      observed_at:$observed_at,
      work:"三命通會",
      purpose:"Freeze the exact Commons discovery observation and the fail-closed selection of a licensed, bibliographically addressable complete scan candidate.",
      query:{
        provider:"commons.wikimedia.org",
        api:"https://commons.wikimedia.org/w/api.php",
        action:"query",
        list:"search",
        search:"intitle:\"三命通會\"",
        namespaces:[6,14],
        limit:500,
        response_path:"evidence/commons-title-search.json",
        response_size_bytes:$response_size,
        response_sha256:$response_sha256
      },
      observation:{
        total_hits:$total_hits,
        returned_hits:$returned_hits,
        file_hits:$file_hits,
        category_hits:$category_hits,
        selected_titles:[
          "File:NCL-06589 1 三命通會.pdf",
          "File:NCL-06589 2 三命通會.pdf"
        ]
      },
      candidate_matrix:[
        {candidate:"NCL-06589",title_hits:2,edition:"明萬曆戊寅(六年, 1578)刊本",structure:"十二卷十二冊 in two PDFs",license:"PD-scan|PD-old",source_identity:"National Central Library call number 306.5 06589",decision:"selected_for_research_snapshot"},
        {candidate:"Harvard-drs-53262215",title_hits:12,edition:"清雍正十三年候選",structure:"twelve PDFs",license:"Public domain",source_identity:"Harvard IIIF drs:53262215",decision:"retained_as_large_alternative_not_fetched"},
        {candidate:"SSID-13035018-13035029",title_hits:12,edition:"上海江左書林 1909 改良本",structure:"twelve-volume scan set",license:"PD-scan|PD-old",source_identity:"SSID identifiers",decision:"retained_as_later_adapted_alternative_not_fetched"},
        {candidate:"CADAL-060660-and-060564",title_hits:24,edition:"欽定四庫全書本",structure:"two twelve-volume scan sets",license:"Commons PD-scan metadata; archive.org records separately lack rights/licenseurl",source_identity:"CADAL identifiers",decision:"retained_as_siku_alternative_not_fetched"}
      ],
      selection:{
        candidate_id:"sanming-ncl-06589-1578-12vol-scan-v1",
        reasons:[
          "earliest precisely identified edition among the audited complete candidates",
          "institution and call number are explicit",
          "twelve-volume and twelve-book structure are explicit",
          "Commons file pages explicitly apply PD-scan|PD-old",
          "two stable PDF file identities provide physical-page locators"
        ]
      },
      boundaries:{
        provider_metadata_is_not_final_bibliographic_adjudication:true,
        title_level_completeness_is_not_page_level_completeness:true,
        local_snapshot_does_not_imply_runtime_ingestion:true,
        bibliography_adjudicated:false,
        complete_primary_text_verified:false,
        page_mapping_verified:false,
        claim_support_reviewed:false,
        runtime_ingestion_allowed:false,
        claim_eligible:false,
        publishable_accuracy:false
      }
    }
  ' >"$tmp/output/discovery-audit.json"

discovery_sha256="$(sha256sum "$tmp/output/discovery-audit.json" | awk '{print $1}')"
artifact_manifest="$tmp/artifact-manifest.txt"
jq -r '. | [.part,.commons_page_id,.fixed_file_page_revision.revision_id,.remote_file.timestamp,.remote_file.size_bytes,.remote_file.page_count,.remote_file.sha1,.local_artifact.sha256] | @tsv' \
  "$entries" >"$artifact_manifest"
artifact_manifest_sha256="$(sha256sum "$artifact_manifest" | awk '{print $1}')"

jq -s \
  --arg version "$VERSION" \
  --arg retrieved_at "$RETRIEVED_AT" \
  --arg discovery_sha256 "$discovery_sha256" \
  --arg artifact_manifest_sha256 "$artifact_manifest_sha256" '
    {
      schema:"commons_scan_snapshot_v1",
      version:$version,
      status:"research_snapshot_not_runtime_eligible",
      retrieved_at:$retrieved_at,
      candidate_id:"sanming-ncl-06589-1578-12vol-scan-v1",
      provider:"commons.wikimedia.org",
      source_institution:"National Central Library, Republic of China (Taiwan)",
      work:"三命通會",
      author:"萬民英",
      edition:"明萬曆戊寅(六年, 1578)刊本",
      bibliographic_identity:{
        call_number:"306.5 06589",
        alternate_title:"三命會通",
        extent:"十二卷，12冊，線裝",
        classification:"子部-術數類-命相之屬",
        collation:"10行，行22字，雙欄，版心白口，單魚尾",
        source_url:"https://rbook.ncl.edu.tw/NCLSearch"
      },
      license:{
        status:"terms_resolved_public_domain_scan",
        commons_template:"PD-scan|PD-old",
        underlying_work:"PD-old",
        digital_reproduction:"PD-scan",
        usage_terms:"Public domain",
        attribution_required:false,
        evidence_paths:[.[0].license_evidence.path,.[1].license_evidence.path],
        evidence_urls:[
          .[0].fixed_file_page_revision.source_url,
          .[1].fixed_file_page_revision.source_url,
          "https://commons.wikimedia.org/wiki/Template:PD-scan",
          "https://commons.wikimedia.org/wiki/Template:PD-old"
        ]
      },
      discovery_audit:{
        path:"discovery-audit.json",
        sha256:$discovery_sha256
      },
      aggregate:{
        file_count:length,
        total_size_bytes:(map(.local_artifact.size_bytes)|add),
        total_physical_pages:(map(.local_artifact.page_count)|add),
        artifact_manifest_scheme:"part_tab_commons_page_id_tab_revision_id_tab_file_timestamp_tab_size_tab_page_count_tab_remote_sha1_tab_local_sha256_lf_v1",
        artifact_manifest_sha256:$artifact_manifest_sha256
      },
      boundaries:{
        scan_artifact_verified:true,
        stable_remote_identity_verified:true,
        license_terms_resolved:true,
        bibliographic_metadata_frozen:true,
        complete_structure_observed:true,
        local_artifact_frozen:true,
        bibliography_adjudicated:false,
        independent_primary_artifact_verified:false,
        complete_primary_text_verified:false,
        page_mapping_verified:false,
        claim_support_reviewed:false,
        runtime_ingestion_allowed:false,
        claim_eligible:false,
        publishable_accuracy:false
      },
      files:.
    }
  ' "$entries" >"$tmp/output/snapshot-manifest.json"

cat >"$tmp/output/ATTRIBUTION.md" <<'EOF'
# Source notice

This research-only snapshot contains two PDF scans of `三命通會`, National Central Library call number `306.5 06589`, identified on Commons as the 1578 Ming Wanli edition.

- Source institution: National Central Library, Republic of China (Taiwan)
- Distribution source: Wikimedia Commons
- Commons license template: `PD-scan|PD-old`
- Usage terms reported by Commons: Public domain

The snapshot is retained for bibliographic and physical-page adjudication. It is not registered as runtime RAG evidence and does not establish page-to-claim support.
EOF

mv "$tmp/output" "$OUTPUT"
echo "Commons NCL research snapshot ready: $OUTPUT"
