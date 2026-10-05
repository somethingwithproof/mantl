#!/usr/bin/env bash
set -euo pipefail

# Validate all Terraform configurations

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "Validating Terraform blueprints..."
echo ""

failed=0
total=0

for dir in infra/terraform/blueprints/*/ infra/terraform/modules/*/; do
    if [[ ! -f "$dir/main.tf" ]]; then
        continue
    fi

    total=$((total + 1))
    echo -e "${YELLOW}Validating $dir${NC}"

    if (cd "$dir" && terraform init -backend=false -input=false -no-color && \
        terraform validate -no-color); then
        echo -e "${GREEN}✓${NC} $dir"
    else
        echo -e "${RED}✗${NC} $dir"
        failed=$((failed + 1))
    fi
done

echo ""
if [[ "$total" -eq 0 ]]; then
    echo "No first-party Terraform modules found" >&2
    exit 1
elif [[ $failed -eq 0 ]]; then
    echo -e "${GREEN}All Terraform configurations valid${NC}"
    exit 0
else
    echo -e "${RED}$failed/$total configurations failed${NC}"
    exit 1
fi
