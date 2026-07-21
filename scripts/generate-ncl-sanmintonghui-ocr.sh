#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SNAPSHOT="${SNAPSHOT:-$ROOT/research/rag/snapshots/sanming-ncl-06589-1578-v1}"
OUTPUT="${1:-$ROOT/research/rag/ocr/sanming-ncl-06589-rapidocr-v1}"
PYTHON="${RAPIDOCR_PYTHON:-/tmp/bazi-rapidocr-venv/bin/python}"
WORK="${OUTPUT}.work"
RENDERED="$WORK/rendered"
PARTIAL="$WORK/page-ocr.jsonl.partial"
FINAL="$WORK/page-ocr.jsonl"

for command in jq sha256sum wc pdftoppm; do
  if ! command -v "$command" >/dev/null 2>&1; then
    echo "required command not found: $command" >&2
    exit 1
  fi
done
if [ ! -x "$PYTHON" ]; then
  echo "RapidOCR Python is not executable: $PYTHON" >&2
  exit 1
fi
if [ -e "$OUTPUT" ]; then
  echo "OCR output already exists: $OUTPUT" >&2
  exit 1
fi

manifest="$SNAPSHOT/snapshot-manifest.json"
volume_labels="$SNAPSHOT/volume-label-observation.json"
manifest_sha256="$(sha256sum "$manifest" | awk '{print $1}')"
volume_labels_sha256="$(sha256sum "$volume_labels" | awk '{print $1}')"
if [ "$manifest_sha256" != "a0db6189460aa495122db8809d70a9837a7e92d30f209a11f7839859b3f6c2b3" ]; then
  echo "unexpected scan snapshot manifest SHA-256: $manifest_sha256" >&2
  exit 1
fi
if [ "$volume_labels_sha256" != "8dfe0d08345b2ded9b58b394e44576e72a8609d46208ac8d4908c05bb873ca31" ]; then
  echo "unexpected volume-label observation SHA-256: $volume_labels_sha256" >&2
  exit 1
fi

readarray -t runtime_identity < <("$PYTHON" - <<'PY'
from importlib.metadata import version
from pathlib import Path
import rapidocr_onnxruntime

root = Path(rapidocr_onnxruntime.__file__).resolve().parent
print(version("rapidocr-onnxruntime"))
print(version("onnxruntime"))
print(version("opencv-python"))
print(root / "config.yaml")
print(root / "models" / "ch_PP-OCRv4_det_infer.onnx")
print(root / "models" / "ch_PP-OCRv4_rec_infer.onnx")
print(root / "models" / "ch_ppocr_mobile_v2.0_cls_infer.onnx")
PY
)
if [ "${runtime_identity[0]}" != "1.4.4" ] || [ "${runtime_identity[1]}" != "1.27.0" ] || [ "${runtime_identity[2]}" != "5.0.0.93" ]; then
  echo "unexpected OCR runtime versions: ${runtime_identity[*]:0:3}" >&2
  exit 1
fi

config_path="${runtime_identity[3]}"
det_model="${runtime_identity[4]}"
rec_model="${runtime_identity[5]}"
cls_model="${runtime_identity[6]}"
config_sha256="$(sha256sum "$config_path" | awk '{print $1}')"
det_sha256="$(sha256sum "$det_model" | awk '{print $1}')"
rec_sha256="$(sha256sum "$rec_model" | awk '{print $1}')"
cls_sha256="$(sha256sum "$cls_model" | awk '{print $1}')"
if [ "$config_sha256" != "bf94a1da4cba828e67b1d61e27cee14d9e7da27c9f272e04048a17e41ae97332" ] ||
   [ "$det_sha256" != "d2a7720d45a54257208b1e13e36a8479894cb74155a5efe29462512d42f49da9" ] ||
   [ "$rec_sha256" != "48fc40f24f6d2a207a2b1091d3437eb3cc3eb6b676dc3ef9c37384005483683b" ] ||
   [ "$cls_sha256" != "e47acedf663230f8863ff1ab0e64dd2d82b838fceb5957146dab185a89d6215c" ]; then
  echo "RapidOCR config or model identity mismatch" >&2
  exit 1
fi

pdftoppm_version="$(pdftoppm -v 2>&1 | head -n 1)"
if [ "$pdftoppm_version" != "pdftoppm version 26.05.0" ]; then
  echo "unexpected pdftoppm version: $pdftoppm_version" >&2
  exit 1
fi

mkdir -p "$RENDERED"
part1_pdf="$SNAPSHOT/$(jq -r '.files[] | select(.part == 1) | .local_artifact.path' "$manifest")"
part2_pdf="$SNAPSHOT/$(jq -r '.files[] | select(.part == 2) | .local_artifact.path' "$manifest")"
if [ "$(sha256sum "$part1_pdf" | awk '{print $1}')" != "3de5c45efb1919965afae00a3c97121054aaabd0f38f9fd5ab8f0a28bb8e36dd" ] ||
   [ "$(sha256sum "$part2_pdf" | awk '{print $1}')" != "3a309831d5b0f0396ac8c89c0176734c560ceaafb03e71b1457f45662de16eb8" ]; then
  echo "scan PDF identity mismatch" >&2
  exit 1
fi

rendered_count="$(find "$RENDERED" -maxdepth 1 -type f -name 'part-*-page-*.jpg' | wc -l | tr -d ' ')"
if [ "$rendered_count" = "0" ]; then
  pdftoppm -jpeg -jpegopt quality=90 -r 72 "$part1_pdf" "$RENDERED/part-1-page"
  pdftoppm -jpeg -jpegopt quality=90 -r 72 "$part2_pdf" "$RENDERED/part-2-page"
  rendered_count="$(find "$RENDERED" -maxdepth 1 -type f -name 'part-*-page-*.jpg' | wc -l | tr -d ' ')"
fi
if [ "$rendered_count" != "1187" ]; then
  echo "rendered scan page count = $rendered_count, want 1187" >&2
  exit 1
fi

export OCR_RENDERED="$RENDERED"
export OCR_PARTIAL="$PARTIAL"
"$PYTHON" - <<'PY'
from concurrent.futures import ProcessPoolExecutor, as_completed
from pathlib import Path
import hashlib
import json
import os
import re

rendered = Path(os.environ["OCR_RENDERED"])
partial = Path(os.environ["OCR_PARTIAL"])
name_pattern = re.compile(r"^part-(1|2)-page-0*([0-9]+)\.jpg$")
engine = None


def init_engine():
    global engine
    from rapidocr_onnxruntime import RapidOCR

    engine = RapidOCR(
        intra_op_num_threads=2,
        inter_op_num_threads=1,
        rec_batch_num=6,
    )


def page_identity(path):
    match = name_pattern.match(Path(path).name)
    if match is None:
        raise ValueError(f"unexpected rendered page name: {path}")
    return int(match.group(1)), int(match.group(2))


def process_page(path):
    from PIL import Image

    part, page = page_identity(path)
    result, _ = engine(path, use_cls=False)
    lines = []
    if result:
        for box, text, score in result:
            lines.append(
                {
                    "box": [[int(round(point[0])), int(round(point[1]))] for point in box],
                    "text": text,
                    "score_basis_points": int(round(float(score) * 10000)),
                }
            )
    with Image.open(path) as image:
        width, height = image.size
    with open(path, "rb") as handle:
        render_sha256 = hashlib.sha256(handle.read()).hexdigest()
    return {
        "part": part,
        "physical_page": page,
        "render": {
            "width_pixels": width,
            "height_pixels": height,
            "sha256": render_sha256,
        },
        "lines": lines,
    }


completed = {}
if partial.exists():
    with partial.open("r", encoding="utf-8") as handle:
        for line_number, line in enumerate(handle, 1):
            if not line.strip():
                continue
            record = json.loads(line)
            key = (record["part"], record["physical_page"])
            if key in completed:
                raise ValueError(f"duplicate partial OCR page {key} at line {line_number}")
            completed[key] = record

paths = []
for path in sorted(rendered.glob("part-*-page-*.jpg")):
    key = page_identity(path)
    if key not in completed:
        paths.append(str(path))

print(f"RapidOCR resume: completed={len(completed)} remaining={len(paths)}", flush=True)
partial.parent.mkdir(parents=True, exist_ok=True)
with partial.open("a", encoding="utf-8") as output:
    if paths:
        with ProcessPoolExecutor(max_workers=4, initializer=init_engine) as pool:
            futures = [pool.submit(process_page, path) for path in paths]
            for index, future in enumerate(as_completed(futures), 1):
                record = future.result()
                key = (record["part"], record["physical_page"])
                if key in completed:
                    raise ValueError(f"duplicate OCR page returned: {key}")
                completed[key] = record
                output.write(json.dumps(record, ensure_ascii=False, separators=(",", ":")) + "\n")
                output.flush()
                if index % 25 == 0 or index == len(paths):
                    print(f"RapidOCR progress: {len(completed)}/1187", flush=True)

expected = [(1, page) for page in range(1, 1001)] + [(2, page) for page in range(1, 188)]
if sorted(completed) != expected:
    raise ValueError("OCR page identities are incomplete or non-contiguous")

final_path = partial.parent / "page-ocr.jsonl"
with final_path.open("w", encoding="utf-8") as output:
    for key in expected:
        output.write(json.dumps(completed[key], ensure_ascii=False, separators=(",", ":")) + "\n")
PY

ocr_size="$(wc -c <"$FINAL" | tr -d ' ')"
ocr_sha256="$(sha256sum "$FINAL" | awk '{print $1}')"
summary="$(jq -s '
  {
    page_count:length,
    part_1_pages:([.[]|select(.part==1)]|length),
    part_2_pages:([.[]|select(.part==2)]|length),
    empty_pages:([.[]|select((.lines|length)==0)]|length),
    recognized_lines:([.[].lines|length]|add),
    score_basis_points_sum:([.[].lines[].score_basis_points]|add),
    minimum_line_score_basis_points:([.[].lines[].score_basis_points]|min),
    maximum_line_score_basis_points:([.[].lines[].score_basis_points]|max)
  }
' "$FINAL")"

jq -n \
  --arg manifest_sha256 "$manifest_sha256" \
  --arg volume_labels_sha256 "$volume_labels_sha256" \
  --arg pdftoppm_version "$pdftoppm_version" \
  --arg config_sha256 "$config_sha256" \
  --arg det_sha256 "$det_sha256" \
  --arg rec_sha256 "$rec_sha256" \
  --arg cls_sha256 "$cls_sha256" \
  --argjson ocr_size "$ocr_size" \
  --arg ocr_sha256 "$ocr_sha256" \
  --argjson summary "$summary" '
    {
      schema:"sanming_ncl_page_ocr_snapshot_v1",
      version:"2026-07-17.1",
      status:"machine_ocr_silver_not_page_mapping_gold",
      generated_at:"2026-07-17",
      candidate_id:"sanming-ncl-06589-1578-12vol-scan-v1",
      purpose:"Freeze a complete page-addressable OCR observation for candidate mapping while preserving scan, transcription, and independent-review uncertainty.",
      inputs:{
        scan_snapshot_manifest:{path:"../../snapshots/sanming-ncl-06589-1578-v1/snapshot-manifest.json",sha256:$manifest_sha256},
        volume_label_observation:{path:"../../snapshots/sanming-ncl-06589-1578-v1/volume-label-observation.json",sha256:$volume_labels_sha256}
      },
      rendering:{
        tool:$pdftoppm_version,
        arguments:"-jpeg -jpegopt quality=90 -r 72",
        color_mode:"source RGB JPEG",
        page_identity:"part plus one-based physical PDF page"
      },
      ocr:{
        package:"rapidocr-onnxruntime",
        package_version:"1.4.4",
        package_license:"Apache-2.0",
        onnxruntime_version:"1.27.0",
        opencv_python_version:"5.0.0.93",
        use_detection:true,
        use_classification:false,
        use_recognition:true,
        worker_processes:4,
        intra_op_threads_per_session:2,
        inter_op_threads_per_session:1,
        recognition_batch_size:6,
        text_score_threshold:0.5,
        config_sha256:$config_sha256,
        models:{
          detection:{name:"ch_PP-OCRv4_det_infer.onnx",sha256:$det_sha256},
          recognition:{name:"ch_PP-OCRv4_rec_infer.onnx",sha256:$rec_sha256},
          classification_unused:{name:"ch_ppocr_mobile_v2.0_cls_infer.onnx",sha256:$cls_sha256}
        }
      },
      artifact:{
        path:"page-ocr.jsonl",
        size_bytes:$ocr_size,
        sha256:$ocr_sha256,
        record_scheme:"part_physical_page_render_identity_and_ordered_line_boxes_text_score_v1"
      },
      summary:$summary,
      boundaries:{
        complete_page_set_observed:true,
        machine_ocr_only:true,
        ocr_text_is_not_diplomatic_transcription:true,
        ocr_score_is_not_mapping_confidence:true,
        independent_review_complete:false,
        complete_primary_text_verified:false,
        volume_mapping_verified:false,
        chapter_page_mapping_verified:false,
        claim_support_reviewed:false,
        runtime_ingestion_allowed:false,
        claim_eligible:false,
        publishable_accuracy:false
      }
    }
  ' >"$WORK/ocr-manifest.json"

rm -f "$PARTIAL"
rm -rf "$RENDERED"
mv "$WORK" "$OUTPUT"
echo "NCL page OCR snapshot ready: $OUTPUT"
