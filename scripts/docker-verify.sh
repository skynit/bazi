#!/bin/bash
# Docker Compose structure verification
# Usage: bash scripts/docker-verify.sh

set -euo pipefail
FAIL=0

pass() {
  echo "✓ $*"
}

fail() {
  echo "✗ $*"
  FAIL=1
}

check_contains() {
  local pattern="$1"
  local label="$2"
  if grep -q "$pattern" docker-compose.yml; then
    pass "$label"
  else
    fail "$label missing"
  fi
}

single_port() {
  local source="$1"
  local values="$2"
  local count
  count="$(printf '%s\n' "$values" | sed '/^$/d' | wc -l)"
  if [[ "$count" -ne 1 ]]; then
    echo "✗ $source must declare one consistent port (found: ${values:-none})" >&2
    return 1
  fi
  printf '%s' "$values"
}

echo "Checking docker-compose.yml..."
check_contains "backend:" "backend service"
check_contains "frontend:" "frontend service"
check_contains "db:" "db service"
check_contains "healthcheck" "healthchecks"
check_contains "mysql_data:" "volume"
check_contains "networks:" "networks"

compose_port_values="$(sed -nE 's/.*SERVER_PORT:-([0-9]+).*/\1/p' docker-compose.yml | sort -u)"
nginx_port_values="$(sed -nE 's/.*proxy_pass http:\/\/backend:([0-9]+).*/\1/p' vue/nginx.conf | sort -u)"
docker_port_values="$(sed -nE 's/^EXPOSE[[:space:]]+([0-9]+)$/\1/p' src/Dockerfile | sort -u)"

compose_port=""
nginx_port=""
docker_port=""
if ! compose_port="$(single_port "docker-compose.yml SERVER_PORT defaults" "$compose_port_values")"; then FAIL=1; fi
if ! nginx_port="$(single_port "vue/nginx.conf backend upstreams" "$nginx_port_values")"; then FAIL=1; fi
if ! docker_port="$(single_port "src/Dockerfile EXPOSE" "$docker_port_values")"; then FAIL=1; fi

if [[ -n "$compose_port" && "$compose_port" == "$nginx_port" && "$compose_port" == "$docker_port" ]]; then
  pass "backend port contract ($compose_port)"
else
  fail "backend port mismatch: compose=${compose_port:-invalid} nginx=${nginx_port:-invalid} docker=${docker_port:-invalid}"
fi

if [ "$FAIL" -eq 0 ]; then
  echo "All checks passed"
else
  echo "Some checks failed"
  exit 1
fi
