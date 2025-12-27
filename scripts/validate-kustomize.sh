#!/usr/bin/env bash
set -euo pipefail

# Validate all Kustomize builds

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

echo "Validating Kustomize builds..."
echo ""

failed=0
total=0

find . -name kustomization.yaml -not -path "*/node_modules/*" -exec dirname {} \; | sort | while read -r dir; do
    total=$((total + 1))
    if kustomize build "$dir" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $dir"
    else
        echo -e "${RED}✗${NC} $dir"
        failed=$((failed + 1))
    fi
done

echo ""
if [ $failed -eq 0 ]; then
    echo -e "${GREEN}All Kustomize builds valid${NC}"
    exit 0
else
    echo -e "${RED}$failed/$total builds failed${NC}"
    exit 1
fi
