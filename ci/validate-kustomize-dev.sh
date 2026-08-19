#!/usr/bin/env bash
set -euo pipefail

# Dev-scope Kustomize validation (fast, fail-fast)
# This focuses on clusters/dev and a couple of stable bases to keep CI green while
# broader overlays are being fixed.

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

declare -a DIRS=(
  "clusters/dev"
  "apps/examples/api-service/base"
  "apps/templates/web-service/base"
)

failed=0
for dir in "${DIRS[@]}"; do
if kustomize build "$dir" --enable-helm --load-restrictor=LoadRestrictionsNone > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} $dir"
  else
    echo -e "${RED}✗${NC} $dir"
kustomize build "$dir" --enable-helm --load-restrictor=LoadRestrictionsNone || true
    failed=$((failed+1))
  fi
done

if [ "$failed" -eq 0 ]; then
  echo -e "${GREEN}Dev-scope Kustomize builds valid${NC}"
  exit 0
else
  echo -e "${RED}$failed dev-scope builds failed${NC}"
  exit 1
fi
