#!/bin/bash
# Docker Compose structure verification
# Usage: bash scripts/docker-verify.sh

set -euo pipefail
FAIL=0

echo "Checking docker-compose.yml..."
grep -q "backend:" docker-compose.yml && echo "✓ backend service" || { echo "✗ backend service missing"; FAIL=1; }
grep -q "frontend:" docker-compose.yml && echo "✓ frontend service" || { echo "✗ frontend service missing"; FAIL=1; }
grep -q "db:" docker-compose.yml && echo "✓ db service" || { echo "✗ db service missing"; FAIL=1; }
grep -q "healthcheck" docker-compose.yml && echo "✓ healthchecks" || { echo "✗ healthchecks missing"; FAIL=1; }
grep -q "mysql_data:" docker-compose.yml && echo "✓ volume" || { echo "✗ volume missing"; FAIL=1; }
grep -q "networks:" docker-compose.yml && echo "✓ networks" || { echo "✗ networks missing"; FAIL=1; }

if [ "$FAIL" -eq 0 ]; then
  echo "All checks passed"
else
  echo "Some checks failed"
  exit 1
fi
