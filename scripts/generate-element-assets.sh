#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEFAULT_JOBS_FILE="$ROOT_DIR/vue/public/element-assets/generation-jobs.jsonl"
DEFAULT_OUT_DIR="$ROOT_DIR/vue/public/element-assets"
JOBS_FILE="${JOBS_FILE:-$DEFAULT_JOBS_FILE}"
OUT_DIR="${OUT_DIR:-$DEFAULT_OUT_DIR}"
MANIFEST_FILE="${MANIFEST_FILE:-$OUT_DIR/manifest.json}"
BACKEND_MANIFEST_FILE="${BACKEND_MANIFEST_FILE:-$ROOT_DIR/src/internal/service/elementasset/defaults.json}"
BOTCF_BASE_URL="${BOTCF_BASE_URL:-https://botcf.com/v1}"
CONCURRENCY="${CONCURRENCY:-2}"
MAX_ATTEMPTS="${MAX_ATTEMPTS:-3}"
FORCE="${FORCE:-0}"

if [[ -f "${BOTCF_ENV_FILE:-$HOME/.config/botcf/imagegen.env}" ]]; then
  # shellcheck disable=SC1090
  source "${BOTCF_ENV_FILE:-$HOME/.config/botcf/imagegen.env}"
fi

API_KEY="${BOTCF_API_KEY:-${OPENAI_API_KEY:-}}"
if [[ -z "$API_KEY" ]]; then
  echo "请先设置 BOTCF_API_KEY，或写入 ~/.config/botcf/imagegen.env" >&2
  exit 1
fi
if [[ ! -f "$JOBS_FILE" ]]; then
  echo "找不到生成任务清单: $JOBS_FILE" >&2
  exit 1
fi
if ! [[ "$CONCURRENCY" =~ ^[1-9][0-9]*$ && "$MAX_ATTEMPTS" =~ ^[1-9][0-9]*$ ]]; then
  echo "CONCURRENCY 和 MAX_ATTEMPTS 必须是正整数" >&2
  exit 1
fi
for command in curl jq python3; do
  command -v "$command" >/dev/null || { echo "缺少命令: $command" >&2; exit 1; }
done
python3 - <<'PY' >/dev/null
from PIL import Image
PY

mkdir -p "$OUT_DIR"
WORK_DIR="$(mktemp -d "${TMPDIR:-/tmp}/botcf-element-assets.XXXXXX")"
trap 'rm -rf "$WORK_DIR"' EXIT

mapfile -t JOBS < <(grep -vE '^[[:space:]]*(#|$)' "$JOBS_FILE")
TOTAL="${#JOBS[@]}"
if (( TOTAL == 0 )); then
  echo "任务清单为空: $JOBS_FILE" >&2
  exit 1
fi

build_payload() {
  local job_file="$1" payload_file="$2"
  python3 - "$job_file" "$payload_file" <<'PY'
import json, sys
job = json.load(open(sys.argv[1], encoding="utf-8"))
labels = {
    "prompt": "主题", "use_case": "用途", "style": "风格", "composition": "构图",
    "lighting": "光线", "palette": "色彩", "constraints": "必须满足", "negative": "避免",
}
parts = [f"{labels[key]}：{job[key]}" for key in labels if job.get(key)]
payload = {
    "model": job.get("model", "gpt-image-2"),
    "size": job.get("size", "1024x1024"),
    "n": 1,
    "prompt": "\n".join(parts),
}
with open(sys.argv[2], "w", encoding="utf-8") as fh:
    json.dump(payload, fh, ensure_ascii=False)
PY
}

save_image() {
  local source_file="$1" output_file="$2" size="$3" quality="$4"
  python3 - "$source_file" "$output_file" "$size" "$quality" <<'PY'
import sys
from pathlib import Path
from PIL import Image, ImageOps

source, output, size, quality = sys.argv[1:]
width, height = map(int, size.lower().split("x", 1))
output_path = Path(output)
output_path.parent.mkdir(parents=True, exist_ok=True)
with Image.open(source) as image:
    image = ImageOps.exif_transpose(image).convert("RGB")
    if image.size != (width, height):
        image = ImageOps.fit(image, (width, height), method=Image.Resampling.LANCZOS)
    temp_path = output_path.with_name(output_path.name + ".part")
    fmt = output_path.suffix.lstrip(".").upper()
    if fmt == "JPG":
        fmt = "JPEG"
    options = {"quality": int(quality), "optimize": True}
    if fmt == "WEBP":
        options["method"] = 6
    image.save(temp_path, format=fmt, **options)
    temp_path.replace(output_path)
PY
}

generate_one() {
  local index="$1" job_json="$2"
  local job_dir="$WORK_DIR/job-$index"
  mkdir -p "$job_dir"
  printf '%s\n' "$job_json" > "$job_dir/job.json"

  local relative_out size quality output_file
  relative_out="$(jq -er '.out' "$job_dir/job.json")"
  size="$(jq -er '.size // "1024x1024"' "$job_dir/job.json")"
  quality="$(jq -er '.output_compression // 82' "$job_dir/job.json")"
  output_file="$OUT_DIR/$relative_out"

  if [[ -s "$output_file" && "$FORCE" != "1" ]]; then
    echo "[任务 $index/$TOTAL] 已存在，跳过: $relative_out"
    return 0
  fi

  build_payload "$job_dir/job.json" "$job_dir/payload.json"
  local attempt image_url
  for ((attempt = 1; attempt <= MAX_ATTEMPTS; attempt++)); do
    echo "[任务 $index/$TOTAL] 生成中（尝试 $attempt/$MAX_ATTEMPTS）: $relative_out"
    if curl --fail-with-body --silent --show-error --location \
      --connect-timeout 20 --max-time 360 \
      "$BOTCF_BASE_URL/images/generations" \
      -H "Authorization: Bearer $API_KEY" \
      -H 'Content-Type: application/json' \
      --data-binary "@$job_dir/payload.json" \
      -o "$job_dir/response.json"; then
      image_url="$(jq -r '.data[0].url // empty' "$job_dir/response.json" 2>/dev/null || true)"
      if [[ -n "$image_url" ]] && curl --fail --silent --show-error --location \
        --connect-timeout 20 --max-time 180 --retry 2 \
        "$image_url" -o "$job_dir/source-image"; then
        if save_image "$job_dir/source-image" "$output_file" "$size" "$quality"; then
          echo "[任务 $index/$TOTAL] 完成: $relative_out"
          return 0
        fi
      fi
    fi
    if (( attempt < MAX_ATTEMPTS )); then
      sleep $((attempt * 3))
    fi
  done

  echo "[任务 $index/$TOTAL] 失败: $relative_out" >&2
  if [[ -f "$job_dir/response.json" ]]; then
    jq -c '{error, message, code, data_count: (.data | length?)}' "$job_dir/response.json" >&2 2>/dev/null || true
  fi
  return 1
}

pids=()
for i in "${!JOBS[@]}"; do
  generate_one "$((i + 1))" "${JOBS[$i]}" &
  pids+=("$!")
  if (( ${#pids[@]} >= CONCURRENCY )); then
    wait "${pids[0]}" || true
    pids=("${pids[@]:1}")
  fi
done
for pid in "${pids[@]}"; do
  wait "$pid" || true
done

missing=0
for job_json in "${JOBS[@]}"; do
  relative_out="$(jq -r '.out' <<<"$job_json")"
  if [[ ! -s "$OUT_DIR/$relative_out" ]]; then
    echo "缺少输出: $relative_out" >&2
    missing=$((missing + 1))
  fi
done
if (( missing > 0 )); then
  echo "生成未完成，共缺少 $missing/$TOTAL 张图片。再次运行脚本会自动续传。" >&2
  exit 1
fi

python3 - "$JOBS_FILE" "$OUT_DIR" "$MANIFEST_FILE" <<'PY'
import json, sys
from pathlib import Path
from PIL import Image

jobs_file, out_dir, manifest_file = map(Path, sys.argv[1:])
elements = {"wood": "木", "fire": "火", "earth": "土", "metal": "金", "water": "水"}
colors = {
    "wood": ("#315f49", "#22c59e"), "fire": ("#7d313c", "#fb7185"),
    "earth": ("#80663a", "#f2bd4d"), "metal": ("#66717e", "#cbd5e1"),
    "water": ("#245d74", "#22d3ee"),
}
assets = []
order = 0
for raw in jobs_file.read_text(encoding="utf-8").splitlines():
    raw = raw.strip()
    if not raw or raw.startswith("#"):
        continue
    job = json.loads(raw)
    rel = Path(job["out"])
    path = out_dir / rel
    with Image.open(path) as image:
        width, height = image.size
    slug = rel.parent.name
    ratio = width / height
    if 0.9 <= ratio <= 1.1:
        orientation, scene = "square", "object"
    elif ratio >= 2.1:
        orientation, scene = "panorama", "hero"
    elif ratio > 1:
        orientation, scene = "landscape", "hero"
    else:
        orientation, scene = "portrait", "general"
    dominant, accent = colors[slug]
    key = rel.stem
    order += 1
    assets.append({
        "key": key, "name": job["prompt"][:32], "element": elements[slug],
        "url": "/element-assets/" + rel.as_posix(), "scene": scene,
        "orientation": orientation, "tone": "balanced", "style": "chinese",
        "season": "all", "time_period": "all", "description": job["prompt"],
        "alt_text": job["prompt"], "dominant_color": dominant, "accent_color": accent,
        "focal_x": 0.5, "focal_y": 0.5, "width": width, "height": height,
        "weight": 100, "sort_order": order, "status": "active",
    })
manifest_file.parent.mkdir(parents=True, exist_ok=True)
manifest_file.write_text(json.dumps({"version": 1, "assets": assets}, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(f"已写入素材清单: {manifest_file}（{len(assets)} 项）")
PY

if [[ "$JOBS_FILE" == "$DEFAULT_JOBS_FILE" && "$OUT_DIR" == "$DEFAULT_OUT_DIR" ]]; then
  mkdir -p "$(dirname "$BACKEND_MANIFEST_FILE")"
  cp "$MANIFEST_FILE" "$BACKEND_MANIFEST_FILE"
  echo "已同步后端内置素材清单: $BACKEND_MANIFEST_FILE"
fi

echo "全部完成，共生成 $TOTAL 张五行图片。"
