#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SNAPSHOT="${1:-$ROOT/research/rag/snapshots/sanming-ncl-06589-1578-v1}"
OUTPUT="${2:-$SNAPSHOT/book-boundary-audit.json}"
MANIFEST="$SNAPSHOT/snapshot-manifest.json"
THRESHOLD="0.30"

for command in jq sha256sum pdftoppm magick awk; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "required command not found: $command" >&2
    exit 1
  fi
done

if [ ! -f "$MANIFEST" ]; then
  echo "snapshot manifest not found: $MANIFEST" >&2
  exit 1
fi
if [ -e "$OUTPUT" ]; then
  echo "book-boundary audit output already exists: $OUTPUT" >&2
  exit 1
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

manifest_sha256="$(sha256sum "$MANIFEST" | awk '{print $1}')"
if [ "$manifest_sha256" != "a0db6189460aa495122db8809d70a9837a7e92d30f209a11f7839859b3f6c2b3" ]; then
  echo "unexpected snapshot manifest SHA-256: $manifest_sha256" >&2
  exit 1
fi

candidates="$tmp/candidates.jsonl"
: >"$candidates"

scan_part() {
  local part="$1"
  local artifact
  local expected_sha256
  local page_count
  local render_dir="$tmp/part-$part"
  local means="$tmp/part-$part-means.tsv"
  local selected="$tmp/part-$part-selected.json"

  artifact="$(jq -r --argjson part "$part" '.files[] | select(.part == $part) | .local_artifact.path' "$MANIFEST")"
  expected_sha256="$(jq -r --argjson part "$part" '.files[] | select(.part == $part) | .local_artifact.sha256' "$MANIFEST")"
  page_count="$(jq -r --argjson part "$part" '.files[] | select(.part == $part) | .local_artifact.page_count' "$MANIFEST")"
  artifact="$SNAPSHOT/$artifact"

  if [ "$(sha256sum "$artifact" | awk '{print $1}')" != "$expected_sha256" ]; then
    echo "snapshot artifact hash mismatch for part $part" >&2
    exit 1
  fi

  mkdir -p "$render_dir"
  pdftoppm -jpeg -jpegopt quality=50 -r 20 "$artifact" "$render_dir/page"
  if [ "$(find "$render_dir" -maxdepth 1 -name 'page-*.jpg' | wc -l | tr -d ' ')" != "$page_count" ]; then
    echo "rendered page count mismatch for part $part" >&2
    exit 1
  fi

  magick identify -format '%f %[fx:mean]\n' "$render_dir"/page-*.jpg >"$means"
  awk -v threshold="$THRESHOLD" '$2 < threshold' "$means" | while read -r file mean; do
    local digits
    local page
    digits="${file#page-}"
    digits="${digits%.jpg}"
    page="$((10#$digits))"
    jq -n -c --argjson part "$part" --argjson page "$page" --argjson mean "$mean" \
      '{part:$part,physical_page:$page,mean_luminance:$mean,classification:"dark_outer_cover_candidate"}' \
      >>"$candidates"
  done

  jq -s --argjson part "$part" '[.[] | select(.part == $part) | .physical_page]' "$candidates" >"$selected"
  case "$part" in
    1)
      jq -e '. == [1,98,99,190,191,257,258,334,335,425,426,523,524,639,640,759,760,876,877,959,960]' "$selected" >/dev/null
      ;;
    2)
      jq -e '. == [70,71,187]' "$selected" >/dev/null
      ;;
    *)
      echo "unexpected part: $part" >&2
      exit 1
      ;;
  esac
}

scan_part 1
scan_part 2

pdftoppm_version="$(pdftoppm -v 2>&1 | head -n 1)"
imagemagick_version="$(magick -version | head -n 1)"

jq -s \
  --arg manifest_sha256 "$manifest_sha256" \
  --arg pdftoppm_version "$pdftoppm_version" \
  --arg imagemagick_version "$imagemagick_version" \
  --argjson threshold "$THRESHOLD" '
    {
      schema:"sanming_ncl_physical_book_boundary_audit_v1",
      version:"2026-07-17.1",
      status:"physical_book_candidates_not_volume_mapping",
      observed_at:"2026-07-17",
      candidate_id:"sanming-ncl-06589-1578-12vol-scan-v1",
      snapshot_manifest:{path:"snapshot-manifest.json",sha256:$manifest_sha256},
      method:{
        renderer:$pdftoppm_version,
        render_arguments:"-jpeg -jpegopt quality=50 -r 20",
        measurement_tool:$imagemagick_version,
        measurement:"normalized mean luminance",
        candidate_rule:"mean_luminance < threshold",
        threshold:$threshold
      },
      observed_cover_candidates:.,
      physical_book_candidate_ranges:[
        {book_candidate:1,segments:[{part:1,first_page:1,last_page:98}]},
        {book_candidate:2,segments:[{part:1,first_page:99,last_page:190}]},
        {book_candidate:3,segments:[{part:1,first_page:191,last_page:257}]},
        {book_candidate:4,segments:[{part:1,first_page:258,last_page:334}]},
        {book_candidate:5,segments:[{part:1,first_page:335,last_page:425}]},
        {book_candidate:6,segments:[{part:1,first_page:426,last_page:523}]},
        {book_candidate:7,segments:[{part:1,first_page:524,last_page:639}]},
        {book_candidate:8,segments:[{part:1,first_page:640,last_page:759}]},
        {book_candidate:9,segments:[{part:1,first_page:760,last_page:876}]},
        {book_candidate:10,segments:[{part:1,first_page:877,last_page:959}]},
        {book_candidate:11,segments:[{part:1,first_page:960,last_page:1000},{part:2,first_page:1,last_page:70}]},
        {book_candidate:12,segments:[{part:2,first_page:71,last_page:187}]}
      ],
      visual_spot_check:{
        status:"single_operator_non_independent",
        sampled_ranges:[
          {part:1,first_page:96,last_page:101,observation:"text end, inner cover, paired dark outer covers, inner cover, next text"},
          {part:2,first_page:68,last_page:74,observation:"text end, inner cover, paired dark outer covers, inner cover, next text"}
        ],
        independent_review:false
      },
      result:{
        dark_cover_candidate_count:length,
        physical_book_candidate_count:12,
        cross_pdf_book_candidate:11,
        final_book_candidate_start:{part:2,physical_page:71},
        provider_extent_consistent_with_observation:true
      },
      boundaries:{
        cover_detection_is_not_volume_label_reading:true,
        sequence_number_is_a_candidate_not_verified_volume_number:true,
        complete_structure_observed:true,
        complete_primary_text_verified:false,
        volume_mapping_verified:false,
        chapter_page_mapping_verified:false,
        claim_support_reviewed:false,
        runtime_ingestion_allowed:false,
        claim_eligible:false,
        publishable_accuracy:false
      }
    }
  ' "$candidates" >"$OUTPUT"

echo "NCL physical-book boundary audit ready: $OUTPUT"
