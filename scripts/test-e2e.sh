#!/bin/bash
# E2E validation script — tests full user flow via curl
# Usage: bash scripts/test-e2e.sh [BASE_URL]
# Default BASE_URL is http://localhost:8080

set -euo pipefail
BASE="${1:-http://localhost:8080}"
PASS=0
FAIL=0

green() { echo -e "\033[32m✓ $1\033[0m"; PASS=$((PASS+1)); }
red()   { echo -e "\033[31m✗ $1\033[0m"; FAIL=$((FAIL+1)); }

echo "=== E2E Test — Base: $BASE ==="

# 1. Health check
echo -n "[1] Health check ... "
if curl -sf "$BASE/health" | grep -q '"ok"'; then
  green "health OK"
else
  red "health FAIL"
fi

# 2. Register user
echo -n "[2] Register user ... "
REG=$(curl -sf -X POST "$BASE/api/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"username":"e2etest","email":"e2e@test.com","password":"Test1234"}')
TOKEN=$(echo "$REG" | grep -o '"token":"[^"]*"' | cut -d'"' -f4 || true)
if [ -n "$TOKEN" ]; then
  green "register OK, token=$TOKEN"
else
  # Maybe already registered — try login
  LOGIN=$(curl -sf -X POST "$BASE/api/auth/login" \
    -H 'Content-Type: application/json' \
    -d '{"username":"e2etest","password":"Test1234"}')
  TOKEN=$(echo "$LOGIN" | grep -o '"token":"[^"]*"' | cut -d'"' -f4 || true)
  if [ -n "$TOKEN" ]; then
    green "login OK (already registered), token=$TOKEN"
  else
    red "register/login FAIL — response: $REG"
  fi
fi

# 3. Create chart
echo -n "[3] Create chart ... "
CHART=$(curl -sf -X POST "$BASE/api/chart" \
  -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"birth_year":1990,"birth_month":1,"birth_day":15,"birth_hour":8,"birth_min":0,"calendar_type":"SOLAR","gender":"MALE","name":"Test"}')
CHART_ID=$(echo "$CHART" | grep -o '"id":[0-9]*' | head -1 | cut -d: -f2 || true)
if [ -n "$CHART_ID" ]; then
  green "chart created, id=$CHART_ID"
else
  red "chart creation FAIL — response: $CHART"
fi

# 4. Get chart by id
if [ -n "${CHART_ID:-}" ]; then
  echo -n "[4] Get chart ... "
  if curl -sf "$BASE/api/charts/$CHART_ID" -H "Authorization: Bearer $TOKEN" | grep -q '"id"'; then
    green "get chart OK"
  else
    red "get chart FAIL"
  fi
fi

# 5. Calculate fortune
if [ -n "${CHART_ID:-}" ]; then
  echo -n "[5] Calculate fortune ... "
  FORTUNE=$(curl -sf -X POST "$BASE/api/fortune" \
    -H "Authorization: Bearer $TOKEN" \
    -H 'Content-Type: application/json' \
    -d "{\"chart_id\":$CHART_ID,\"query_date\":\"$(date +%Y-%m-%d)\"}")
  if echo "$FORTUNE" | grep -q '"day_gan_zhi"'; then
    green "fortune OK"
  else
    red "fortune FAIL — response: $FORTUNE"
  fi
fi

# 6. List charts (history)
echo -n "[6] List charts ... "
if curl -sf "$BASE/api/charts" -H "Authorization: Bearer $TOKEN" | grep -q '"charts"'; then
  green "history list OK"
else
  red "history list FAIL"
fi

# Summary
echo ""
echo "=== E2E Results: $PASS passed, $FAIL failed ==="
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
