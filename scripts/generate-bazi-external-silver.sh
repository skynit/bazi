#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INPUT_PATH="${1:-${ROOT_DIR}/src/internal/service/testdata/bazi_external_silver_inputs.json}"
OUTPUT_PATH="${2:-${ROOT_DIR}/src/internal/service/testdata/bazi_external_silver.json}"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

clone_pinned() {
  local key="$1"
  local destination="$2"
  local repository commit
  repository="$(jq -er ".sources.${key}.repository" "${INPUT_PATH}")"
  commit="$(jq -er ".sources.${key}.commit | select(test(\"^[0-9a-f]{40}$\"))" "${INPUT_PATH}")"
  git -C "${destination}" init --quiet
  git -C "${destination}" remote add origin "${repository}"
  git -C "${destination}" fetch --quiet --depth 1 origin "${commit}"
  git -C "${destination}" checkout --quiet --detach FETCH_HEAD
  if [[ "$(git -C "${destination}" rev-parse HEAD)" != "${commit}" ]]; then
    echo "${key} commit mismatch" >&2
    exit 1
  fi
}

verify_files() {
  local key="$1"
  local destination="$2"
  while IFS=$'\t' read -r relative expected; do
    local actual
    actual="$(sha256sum "${destination}/${relative}" | awk '{print $1}')"
    if [[ "${actual}" != "${expected}" ]]; then
      echo "${key} ${relative} SHA-256 mismatch: got ${actual}, want ${expected}" >&2
      exit 1
    fi
  done < <(jq -er ".sources.${key}.files | to_entries[] | [.key, .value] | @tsv" "${INPUT_PATH}")
}

LUNAR_DIR="${TMP_DIR}/lunar-javascript"
CNLUNAR_DIR="${TMP_DIR}/cnlunar"
mkdir -p "${LUNAR_DIR}" "${CNLUNAR_DIR}"
clone_pinned lunar_javascript "${LUNAR_DIR}"
clone_pinned cnlunar "${CNLUNAR_DIR}"
verify_files lunar_javascript "${LUNAR_DIR}"
verify_files cnlunar "${CNLUNAR_DIR}"

export INPUT_PATH OUTPUT_PATH LUNAR_DIR CNLUNAR_DIR
mkdir -p "$(dirname "${OUTPUT_PATH}")"
node <<'NODE'
const fs = require('node:fs');
const path = require('node:path');
const { spawnSync } = require('node:child_process');
const { Solar } = require(process.env.LUNAR_DIR);

const manifest = JSON.parse(fs.readFileSync(process.env.INPUT_PATH, 'utf8'));
const python = String.raw`
import datetime
import json
import os
import sys

sys.path.insert(0, os.environ['CNLUNAR_DIR'])
import cnlunar
from cnlunar.config import the60HeavenlyEarth, theHalf60HeavenlyEarth5ElementsList

def nayin(pillar):
    return theHalf60HeavenlyEarth5ElementsList[the60HeavenlyEarth.index(pillar) // 2]

result = {}
for item in json.load(sys.stdin):
    chart = cnlunar.Lunar(
        datetime.datetime(item['year'], item['month'], item['day'], item['hour'], item['minute'], item['second']),
        godType='8char',
        year8Char='beginningOfSpring',
    )
    pillars = {
        'year': chart.year8Char,
        'month': chart.month8Char,
        'day': chart.day8Char,
        'hour': chart.twohour8Char,
    }
    result[item['id']] = {
        'precision': 'day_level_solar_terms',
        'pillars': pillars,
        'nayin': {key: nayin(value) for key, value in pillars.items()},
    }
json.dump(result, sys.stdout, ensure_ascii=False, sort_keys=True)
`;
const cnlunar = spawnSync('python3', ['-c', python], {
  input: JSON.stringify(manifest.cases),
  encoding: 'utf8',
  env: process.env,
});
if (cnlunar.status !== 0) {
  throw new Error(cnlunar.stderr || `cnlunar exited ${cnlunar.status}`);
}
const cnlunarResults = JSON.parse(cnlunar.stdout);

const pillarValues = (pillars) => [pillars.year, pillars.month, pillars.day, pillars.hour];
const cases = manifest.cases.map((input) => {
  const eightChar = Solar.fromYmdHms(
    input.year, input.month, input.day, input.hour, input.minute, input.second,
  ).getLunar().getEightChar();
  const ziHourPolicy = input.zi_hour_policy || 'late_zi_next_day';
  if (!['late_zi_next_day', 'late_zi_same_day'].includes(ziHourPolicy)) {
    throw new Error(`${input.id} has unknown zi_hour_policy ${ziHourPolicy}`);
  }
  eightChar.setSect(ziHourPolicy === 'late_zi_same_day' ? 2 : 1);
  const yun = eightChar.getYun(input.gender === 'MALE' ? 1 : 0, 2);
  const lunarJavascript = {
    precision: 'second_level_solar_terms',
    zi_hour_policy: ziHourPolicy,
    pillars: {
      year: eightChar.getYear(),
      month: eightChar.getMonth(),
      day: eightChar.getDay(),
      hour: eightChar.getTime(),
    },
    ming_gong: eightChar.getMingGong(),
    nayin: {
      year: eightChar.getYearNaYin(),
      month: eightChar.getMonthNaYin(),
      day: eightChar.getDayNaYin(),
      hour: eightChar.getTimeNaYin(),
    },
    dayun: {
      direction: yun.isForward() ? 'forward' : 'reverse',
      pillars: yun.getDaYun(9).slice(1).map((item) => item.getGanZhi()),
    },
  };
  const cnlunarResult = cnlunarResults[input.id];
  const pillarsAgree = JSON.stringify(pillarValues(lunarJavascript.pillars)) ===
    JSON.stringify(pillarValues(cnlunarResult.pillars));
  if (input.scope === 'dual_consensus' && !pillarsAgree) {
    throw new Error(`${input.id} is marked dual_consensus but sources disagree`);
  }
  if (input.scope === 'upstream_disputed' && pillarsAgree) {
    throw new Error(`${input.id} is marked upstream_disputed but sources agree`);
  }
  return {
    id: input.id,
    scope: input.scope,
    term: input.term || '',
    input: {
      year: input.year,
      month: input.month,
      day: input.day,
      hour: input.hour,
      minute: input.minute,
      second: input.second,
      gender: input.gender,
      zi_hour_policy: ziHourPolicy,
    },
    consensus: {
      pillars_agree: pillarsAgree,
      admitted: input.scope === 'dual_consensus' && pillarsAgree,
    },
    lunar_javascript: lunarJavascript,
    cnlunar: cnlunarResult,
  };
});

const sources = Object.entries(manifest.sources).map(([id, source]) => ({
  id,
  repository: source.repository,
  commit: source.commit,
  license: source.license,
  files: source.files,
}));
const output = {
  version: manifest.version,
  description: manifest.description,
  metadata: {
    tier: 'silver',
    purpose: 'bazi_external_structural_differential',
    review_status: 'cross_checked_not_gold',
    publishable_accuracy: false,
    generator: 'scripts/generate-bazi-external-silver.sh',
    dual_consensus_cases: cases.filter((item) => item.scope === 'dual_consensus').length,
    boundary_cases: cases.filter((item) => item.scope === 'lunar_js_jie_boundary').length,
    boundary_groups: new Set(cases.filter((item) => item.scope === 'lunar_js_jie_boundary').map((item) => `${item.term}/${item.input.year}`)).size,
    zi_policy_cases: cases.filter((item) => item.scope === 'lunar_js_zi_boundary').length,
    disputed_cases: cases.filter((item) => item.scope === 'upstream_disputed').length,
    limitations: [
      'cnlunar solar terms have day precision and are not used to adjudicate second-level boundaries',
      'lunar-javascript and production tyme4go share an author, so boundary checks are correlated Silver evidence',
      'late-Zi sect checks use pinned lunar-javascript only and are correlated Silver evidence, not independent Gold',
      'traditional interpretation accuracy and dayun start-age school choice are outside this fixture',
      'nayin aliases 沙/砂 and 泉中水/井泉水 are normalized explicitly before comparison',
    ],
  },
  sources,
  cases,
};
fs.writeFileSync(process.env.OUTPUT_PATH, `${JSON.stringify(output, null, 2)}\n`);
NODE

jq empty "${OUTPUT_PATH}"
echo "generated ${OUTPUT_PATH}"
