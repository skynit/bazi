#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SNAPSHOT="${1:-$ROOT/research/rag/snapshots/sanming-ncl-06589-1578-v1}"
OUTPUT="${2:-$SNAPSHOT/volume-label-observation.json}"
EVIDENCE_DIR="$SNAPSHOT/evidence/volume-labels"
MANIFEST="$SNAPSHOT/snapshot-manifest.json"
BOUNDARY_AUDIT="$SNAPSHOT/book-boundary-audit.json"

for command in jq sha256sum wc pdftoppm magick; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "required command not found: $command" >&2
    exit 1
  fi
done

if [ -e "$OUTPUT" ]; then
  echo "volume-label observation output already exists: $OUTPUT" >&2
  exit 1
fi
if [ -e "$EVIDENCE_DIR" ]; then
  echo "volume-label evidence directory already exists: $EVIDENCE_DIR" >&2
  exit 1
fi

manifest_sha256="$(sha256sum "$MANIFEST" | awk '{print $1}')"
boundary_sha256="$(sha256sum "$BOUNDARY_AUDIT" | awk '{print $1}')"
if [ "$manifest_sha256" != "a0db6189460aa495122db8809d70a9837a7e92d30f209a11f7839859b3f6c2b3" ]; then
  echo "unexpected snapshot manifest SHA-256: $manifest_sha256" >&2
  exit 1
fi
if [ "$boundary_sha256" != "bab12b9be839ce8205cc22705529cc6c9fcaab2168241338bba3aff48642f8e2" ]; then
  echo "unexpected physical-book boundary audit SHA-256: $boundary_sha256" >&2
  exit 1
fi

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
mkdir -p "$tmp/evidence"
entries="$tmp/entries.jsonl"
: >"$entries"

observe_volume() {
  local volume="$1"
  local numeral="$2"
  local part="$3"
  local physical_page="$4"
  local title="三命通會卷之$numeral"
  local artifact_path
  local artifact_sha256
  local commons_page_id
  local page_count
  local book_segments
  local evidence_name
  local evidence_path
  local output_prefix
  local evidence_size
  local evidence_sha256
  local dimensions
  local width
  local height

  artifact_path="$(jq -r --argjson part "$part" '.files[] | select(.part == $part) | .local_artifact.path' "$MANIFEST")"
  artifact_sha256="$(jq -r --argjson part "$part" '.files[] | select(.part == $part) | .local_artifact.sha256' "$MANIFEST")"
  commons_page_id="$(jq -r --argjson part "$part" '.files[] | select(.part == $part) | .commons_page_id' "$MANIFEST")"
  page_count="$(jq -r --argjson part "$part" '.files[] | select(.part == $part) | .local_artifact.page_count' "$MANIFEST")"
  book_segments="$(jq -c --argjson volume "$volume" '.physical_book_candidate_ranges[] | select(.book_candidate == $volume) | .segments' "$BOUNDARY_AUDIT")"
  artifact_path="$SNAPSHOT/$artifact_path"

  if [ "$(sha256sum "$artifact_path" | awk '{print $1}')" != "$artifact_sha256" ]; then
    echo "snapshot artifact hash mismatch for volume $volume" >&2
    exit 1
  fi
  if [ "$physical_page" -lt 1 ] || [ "$physical_page" -gt "$page_count" ]; then
    echo "volume-label page outside PDF for volume $volume" >&2
    exit 1
  fi
  if ! jq -e --argjson part "$part" --argjson page "$physical_page" '
    any(.[]; .part == $part and .first_page <= $page and .last_page >= $page)
  ' <<<"$book_segments" >/dev/null; then
    echo "volume-label page outside physical-book candidate $volume" >&2
    exit 1
  fi

  evidence_name="volume-$(printf '%02d' "$volume")-part-$part-page-$(printf '%04d' "$physical_page").jpg"
  evidence_path="$tmp/evidence/$evidence_name"
  output_prefix="${evidence_path%.jpg}"
  pdftoppm -f "$physical_page" -singlefile -jpeg -jpegopt quality=90 -r 144 \
    "$artifact_path" "$output_prefix"

  evidence_size="$(wc -c <"$evidence_path" | tr -d ' ')"
  evidence_sha256="$(sha256sum "$evidence_path" | awk '{print $1}')"
  dimensions="$(magick identify -format '%w %h' "$evidence_path")"
  read -r width height <<<"$dimensions"

  jq -n -c \
    --argjson volume "$volume" \
    --arg numeral "$numeral" \
    --arg title "$title" \
    --argjson part "$part" \
    --argjson physical_page "$physical_page" \
    --argjson book_segments "$book_segments" \
    --arg source_locator "https://commons.wikimedia.org/wiki/Special:Redirect/page/$commons_page_id?page=$physical_page" \
    --arg evidence_path "evidence/volume-labels/$evidence_name" \
    --argjson evidence_size "$evidence_size" \
    --arg evidence_sha256 "$evidence_sha256" \
    --argjson width "$width" \
    --argjson height "$height" '
      {
        book_candidate:$volume,
        volume_candidate:$volume,
        printed_volume_numeral:$numeral,
        transcribed_heading:$title,
        transcription_status:"single_operator_visual_reading",
        source:{
          part:$part,
          physical_page:$physical_page,
          physical_book_segments:$book_segments,
          commons_page_locator:$source_locator
        },
        evidence:{
          path:$evidence_path,
          size_bytes:$evidence_size,
          sha256:$evidence_sha256,
          width_pixels:$width,
          height_pixels:$height,
          mime:"image/jpeg"
        }
      }
    ' >>"$entries"
}

observe_volume 1  一   1 7
observe_volume 2  二   1 101
observe_volume 3  三   1 193
observe_volume 4  四   1 260
observe_volume 5  五   1 337
observe_volume 6  六   1 428
observe_volume 7  七   1 526
observe_volume 8  八   1 642
observe_volume 9  九   1 762
observe_volume 10 十   1 879
observe_volume 11 十一 1 962
observe_volume 12 十二 2 73

mapping_rows="$tmp/mapping-manifest.txt"
jq -r '. | [.volume_candidate,.book_candidate,.source.part,.source.physical_page,.transcribed_heading,.evidence.sha256] | @tsv' \
  "$entries" >"$mapping_rows"
mapping_manifest_sha256="$(sha256sum "$mapping_rows" | awk '{print $1}')"
pdftoppm_version="$(pdftoppm -v 2>&1 | head -n 1)"

jq -s \
  --arg manifest_sha256 "$manifest_sha256" \
  --arg boundary_sha256 "$boundary_sha256" \
  --arg mapping_manifest_sha256 "$mapping_manifest_sha256" \
  --arg pdftoppm_version "$pdftoppm_version" '
    {
      schema:"sanming_ncl_volume_label_observation_v1",
      version:"2026-07-17.1",
      status:"single_operator_volume_mapping_candidates_not_gold",
      observed_at:"2026-07-17",
      candidate_id:"sanming-ncl-06589-1578-12vol-scan-v1",
      purpose:"Freeze page-addressable visual evidence and single-operator transcriptions for the twelve printed volume headings without promoting them to independently reviewed page-mapping Gold.",
      inputs:{
        snapshot_manifest:{path:"snapshot-manifest.json",sha256:$manifest_sha256},
        physical_book_boundary_audit:{path:"book-boundary-audit.json",sha256:$boundary_sha256}
      },
      method:{
        evidence_renderer:$pdftoppm_version,
        render_arguments:"-f physical_page -singlefile -jpeg -jpegopt quality=90 -r 144",
        reading_order:"vertical_columns_right_to_left",
        selection_rule:"first observed printed heading matching 三命通會卷之{Chinese volume numeral} inside each physical-book candidate",
        transcription_scope:"printed_volume_heading_only",
        mapping_manifest_scheme:"volume_tab_book_candidate_tab_part_tab_physical_page_tab_transcribed_heading_tab_evidence_sha256_lf_v1",
        mapping_manifest_sha256:$mapping_manifest_sha256
      },
      review:{
        operator_count:1,
        independent_reviewer_count:0,
        adjudicator_count:0,
        status:"single_operator_non_independent",
        disagreements:[],
        gold_eligible:false
      },
      observations:.,
      result:{
        physical_book_candidate_count:12,
        printed_volume_heading_count:length,
        unique_volume_candidate_count:([.[].volume_candidate] | unique | length),
        sequential_volume_candidates:([.[].volume_candidate] == [1,2,3,4,5,6,7,8,9,10,11,12]),
        book_to_volume_candidate_alignment:([.[] | .book_candidate == .volume_candidate] | all),
        volume_one_front_matter_before_heading:{part:1,first_page:3,last_page:6},
        direct_printed_heading_observed_for_all_books:true
      },
      boundaries:{
        printed_volume_labels_observed:true,
        single_operator_mapping_complete:true,
        provider_extent_consistent:true,
        independent_review_complete:false,
        bibliography_adjudicated:false,
        independent_primary_artifact_verified:false,
        complete_primary_text_verified:false,
        volume_mapping_verified:false,
        chapter_page_mapping_verified:false,
        claim_support_reviewed:false,
        runtime_ingestion_allowed:false,
        claim_eligible:false,
        publishable_accuracy:false
      }
    }
  ' "$entries" >"$tmp/volume-label-observation.json"

mkdir -p "$(dirname "$EVIDENCE_DIR")"
mv "$tmp/evidence" "$EVIDENCE_DIR"
mv "$tmp/volume-label-observation.json" "$OUTPUT"
echo "NCL volume-label observation ready: $OUTPUT"
