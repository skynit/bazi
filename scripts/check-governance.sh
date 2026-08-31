#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

fail() {
  echo "governance check failed: $*" >&2
  exit 1
}

[[ -f docs/testing.md ]] || fail "docs/testing.md is missing"
if git check-ignore -q docs/testing.md; then
  fail "docs/testing.md is ignored by Git"
fi
grep -Fq '(testing.md)' docs/frontend-ux-remediation-priorities.md ||
  fail "tracked UX guidance must link to docs/testing.md"

node scripts/check-governance.mjs

go_version="$(awk '$1 == "go" { print $2; exit }' src/go.mod)"
docker_go_version="$(sed -nE 's/^FROM golang:([0-9.]+)-alpine AS builder$/\1/p' src/Dockerfile)"
[[ -n "$go_version" ]] || fail "Go version is missing from src/go.mod"
[[ "$docker_go_version" == "$go_version" ]] ||
  fail "src/Dockerfile uses Go ${docker_go_version:-unknown}; src/go.mod requires $go_version"

node <<'NODE'
const fs = require('node:fs')
const packageJson = JSON.parse(fs.readFileSync('vue/package.json', 'utf8'))
for (const name of ['build', 'lint', 'test']) {
  if (!packageJson.scripts?.[name]) {
    console.error(`governance check failed: vue/package.json is missing the ${name} script`)
    process.exit(1)
  }
}
NODE

while IFS= read -r script; do
  if ! sed -n '1,8p' "$script" | grep -Fq 'set -euo pipefail'; then
    fail "$script must enable set -euo pipefail near the entry point"
  fi
done < <(find scripts -maxdepth 1 -type f -name '*.sh' -print | sort)

echo "governance checks passed"
