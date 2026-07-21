#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INPUT_PATH="${1:-${ROOT_DIR}/src/internal/service/testdata/ziwei_iztro_silver_inputs.json}"
OUTPUT_PATH="${2:-${ROOT_DIR}/src/internal/service/testdata/ziwei_iztro_silver.json}"
REPOSITORY="$(jq -r '.source.repository' "${INPUT_PATH}")"
COMMIT="$(jq -r '.source.commit' "${INPUT_PATH}")"
EXPECTED_STARS_SHA256="$(jq -er '.source.stars_sha256 | select(test("^[0-9a-f]{64}$"))' "${INPUT_PATH}")"
EXPECTED_STEMS_SHA256="$(jq -er '.source.heavenly_stems_sha256 | select(test("^[0-9a-f]{64}$"))' "${INPUT_PATH}")"
EXPECTED_LUNAR_MONTH_SHA256="$(jq -er '.source.lunar_month_index_sha256 | select(test("^[0-9a-f]{64}$"))' "${INPUT_PATH}")"
EXPECTED_MONTHLY_STAR_SHA256="$(jq -er '.source.monthly_star_location_sha256 | select(test("^[0-9a-f]{64}$"))' "${INPUT_PATH}")"
EXPECTED_ADJECTIVE_STARS_SHA256="$(jq -er '.source.adjective_stars_sha256 | select(test("^[0-9a-f]{64}$"))' "${INPUT_PATH}")"
EXPECTED_PERIOD_SHA256="$(jq -er '.source.functional_astrolabe_sha256 | select(test("^[0-9a-f]{64}$"))' "${INPUT_PATH}")"
EXPECTED_TRANSIT_SHA256="$(jq -er '.source.horoscope_stars_sha256 | select(test("^[0-9a-f]{64}$"))' "${INPUT_PATH}")"
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "${TMP_DIR}"' EXIT

git -C "${TMP_DIR}" init --quiet
git -C "${TMP_DIR}" remote add origin "${REPOSITORY}"
git -C "${TMP_DIR}" fetch --quiet --depth 1 origin "${COMMIT}"
git -C "${TMP_DIR}" checkout --quiet --detach FETCH_HEAD

if [[ "$(git -C "${TMP_DIR}" rev-parse HEAD)" != "${COMMIT}" ]]; then
  echo "iztro commit mismatch" >&2
  exit 1
fi

export STARS_SHA256="$(sha256sum "${TMP_DIR}/src/data/stars.ts" | awk '{print $1}')"
export STEMS_SHA256="$(sha256sum "${TMP_DIR}/src/data/heavenlyStems.ts" | awk '{print $1}')"
export LUNAR_MONTH_SHA256="$(sha256sum "${TMP_DIR}/src/utils/index.ts" | awk '{print $1}')"
export MONTHLY_STAR_SHA256="$(sha256sum "${TMP_DIR}/src/star/location.ts" | awk '{print $1}')"
export ADJECTIVE_STARS_SHA256="$(sha256sum "${TMP_DIR}/src/star/adjectiveStar.ts" | awk '{print $1}')"
export PERIOD_SHA256="$(sha256sum "${TMP_DIR}/src/astro/FunctionalAstrolabe.ts" | awk '{print $1}')"
export TRANSIT_SHA256="$(sha256sum "${TMP_DIR}/src/star/horoscopeStar.ts" | awk '{print $1}')"
if [[ "${STARS_SHA256}" != "${EXPECTED_STARS_SHA256}" ]]; then
  echo "iztro stars.ts SHA-256 mismatch: got ${STARS_SHA256}, want ${EXPECTED_STARS_SHA256}" >&2
  exit 1
fi
if [[ "${STEMS_SHA256}" != "${EXPECTED_STEMS_SHA256}" ]]; then
  echo "iztro heavenlyStems.ts SHA-256 mismatch: got ${STEMS_SHA256}, want ${EXPECTED_STEMS_SHA256}" >&2
  exit 1
fi
if [[ "${LUNAR_MONTH_SHA256}" != "${EXPECTED_LUNAR_MONTH_SHA256}" ]]; then
  echo "iztro utils/index.ts SHA-256 mismatch: got ${LUNAR_MONTH_SHA256}, want ${EXPECTED_LUNAR_MONTH_SHA256}" >&2
  exit 1
fi
if [[ "${MONTHLY_STAR_SHA256}" != "${EXPECTED_MONTHLY_STAR_SHA256}" ]]; then
  echo "iztro star/location.ts SHA-256 mismatch: got ${MONTHLY_STAR_SHA256}, want ${EXPECTED_MONTHLY_STAR_SHA256}" >&2
  exit 1
fi
if [[ "${ADJECTIVE_STARS_SHA256}" != "${EXPECTED_ADJECTIVE_STARS_SHA256}" ]]; then
  echo "iztro star/adjectiveStar.ts SHA-256 mismatch: got ${ADJECTIVE_STARS_SHA256}, want ${EXPECTED_ADJECTIVE_STARS_SHA256}" >&2
  exit 1
fi
if [[ "${PERIOD_SHA256}" != "${EXPECTED_PERIOD_SHA256}" ]]; then
  echo "iztro FunctionalAstrolabe.ts SHA-256 mismatch: got ${PERIOD_SHA256}, want ${EXPECTED_PERIOD_SHA256}" >&2
  exit 1
fi
if [[ "${TRANSIT_SHA256}" != "${EXPECTED_TRANSIT_SHA256}" ]]; then
  echo "iztro horoscopeStar.ts SHA-256 mismatch: got ${TRANSIT_SHA256}, want ${EXPECTED_TRANSIT_SHA256}" >&2
  exit 1
fi

corepack yarn --cwd "${TMP_DIR}" install --frozen-lockfile --non-interactive

export INPUT_PATH OUTPUT_PATH TMP_DIR COMMIT

mkdir -p "$(dirname "${OUTPUT_PATH}")"
node <<'NODE'
const fs = require('node:fs');
const path = require('node:path');
const { astro } = require(path.join(process.env.TMP_DIR, 'lib'));

const manifest = JSON.parse(fs.readFileSync(process.env.INPUT_PATH, 'utf8'));
const normalizePalace = (name) => name === '仆役' ? '交友' : name === '官禄' ? '事业' : name;
const timeIndex = (hour) => hour === 23 ? 0 : Math.floor((hour + 1) / 2) % 12;
const stars = (items) => items
  .map((star) => ({
    name: star.name,
    brightness: star.brightness || '',
    mutagen: star.mutagen || '',
  }))
  .sort((left, right) => left.name < right.name ? -1 : left.name > right.name ? 1 : 0);

const cases = manifest.cases.map((input) => {
  const chart = astro.bySolar(input.date, timeIndex(input.hour), input.gender, true, 'zh-CN');
  const palaces = {};
  for (const palace of chart.palaces) {
    palaces[palace.earthlyBranch] = {
      name: normalizePalace(palace.name),
      heavenly_stem: palace.heavenlyStem,
      main_stars: stars(palace.majorStars),
      aux_stars: stars(palace.minorStars),
      adjective_stars: palace.adjectiveStars.map((star) => star.name).sort(),
      changsheng_12: palace.changsheng12,
      boshi_12: palace.boshi12,
      jiang_qian_12: palace.jiangqian12,
      sui_qian_12: palace.suiqian12,
      dayun: { start_age: palace.decadal.range[0], end_age: palace.decadal.range[1] },
    };
  }
  return {
    id: input.id,
    input,
    expected: {
      soul_palace: chart.earthlyBranchOfSoulPalace,
      body_palace: chart.earthlyBranchOfBodyPalace,
      life_master: chart.soul,
      body_master: chart.body,
      five_bureau: chart.fiveElementsClass,
      palaces,
    },
  };
});

const periodCases = manifest.period_cases.map((input) => {
  const chart = astro.bySolar(input.birth_date, timeIndex(input.birth_hour), input.gender, true, 'zh-CN');
  const period = chart.horoscope(input.target);
  const branches = ['寅', '卯', '辰', '巳', '午', '未', '申', '酉', '戌', '亥', '子', '丑'];
  const normalizePeriod = (item) => ({
    index: item.index,
    heavenly_stem: item.heavenlyStem,
    earthly_branch: item.earthlyBranch,
    mutagen: item.mutagen,
    palace_names_by_branch: Object.fromEntries(branches.map((branch, index) => [branch, normalizePalace(item.palaceNames[index])])),
    stars_by_branch: Object.fromEntries(branches.map((branch, index) => [branch, item.stars[index].map((star) => star.name)])),
  });
  return {
    id: input.id,
    input,
    expected: {
      solar_date: period.solarDate,
      lunar_date: period.lunarDate,
      yearly: normalizePeriod(period.yearly),
      monthly: normalizePeriod(period.monthly),
      daily: normalizePeriod(period.daily),
    },
  };
});

const output = {
  version: manifest.version,
  description: manifest.description,
  profile_id: manifest.profile_id,
  metadata: {
    tier: 'silver',
    purpose: 'external_chart_and_period_differential',
    review_status: 'cross_checked_not_gold',
    publishable_accuracy: false,
    generator: 'scripts/generate-ziwei-iztro-silver.sh',
  },
  sources: [
    {
      rule_id: 'ziwei.star-brightness.iztro-v1',
      repository: manifest.source.repository,
      commit: process.env.COMMIT,
      path: 'src/data/stars.ts',
      sha256: process.env.STARS_SHA256,
      license: manifest.source.license,
    },
    {
      rule_id: 'ziwei.sihua.ten-stem.iztro-v1',
      repository: manifest.source.repository,
      commit: process.env.COMMIT,
      path: 'src/data/heavenlyStems.ts',
      sha256: process.env.STEMS_SHA256,
      license: manifest.source.license,
    },
    {
      rule_id: 'ziwei.leap-month.normalization.iztro-v1',
      repository: manifest.source.repository,
      commit: process.env.COMMIT,
      path: 'src/utils/index.ts',
      sha256: process.env.LUNAR_MONTH_SHA256,
      license: manifest.source.license,
    },
    {
      rule_id: 'ziwei.monthly-stars.iztro-v1',
      repository: manifest.source.repository,
      commit: process.env.COMMIT,
      path: 'src/star/location.ts',
      sha256: process.env.MONTHLY_STAR_SHA256,
      license: manifest.source.license,
    },
    {
      rule_id: 'ziwei.adjective-stars.iztro-v1',
      repository: manifest.source.repository,
      commit: process.env.COMMIT,
      path: 'src/star/adjectiveStar.ts',
      sha256: process.env.ADJECTIVE_STARS_SHA256,
      license: manifest.source.license,
    },
    {
      rule_id: 'ziwei.period-chronology.iztro-normal-v1',
      repository: manifest.source.repository,
      commit: process.env.COMMIT,
      path: 'src/astro/FunctionalAstrolabe.ts',
      sha256: process.env.PERIOD_SHA256,
      license: manifest.source.license,
    },
    {
      rule_id: 'ziwei.transit-stars.iztro-v1',
      repository: manifest.source.repository,
      commit: process.env.COMMIT,
      path: 'src/star/horoscopeStar.ts',
      sha256: process.env.TRANSIT_SHA256,
      license: manifest.source.license,
    },
  ],
  cases,
  period_cases: periodCases,
};

fs.writeFileSync(process.env.OUTPUT_PATH, `${JSON.stringify(output, null, 2)}\n`);
NODE

jq empty "${OUTPUT_PATH}"
echo "generated ${OUTPUT_PATH}"
