#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

FAIL=0

# Only scan Kyverno policy files to avoid false positives in docs/overlays
FILES=$(git ls-files 'policies/kyverno/*.yaml' 2>/dev/null || true)

if [ -z "$FILES" ]; then
  echo "No Kyverno policy files found."
  exit 0
fi

# Patterns to flag
PATTERNS=(
  'ECR_PLACEHOLDER_KEY'
  'GCR_PLACEHOLDER_KEY'
  'AnPlaceholderKey'
  'YOUR_ORG'
)

for f in $FILES; do
  for p in "${PATTERNS[@]}"; do
    if grep -q "$p" "$f"; then
      echo "Placeholder '$p' found in $f"
      FAIL=1
    fi
  done
done

if [ $FAIL -ne 0 ]; then
  cat <<'EOF'
Kyverno policy placeholders detected.
Fixes:
  - Replace publicKeys placeholders with your Cosign keys (scripts/set-cosign-key.sh)
  - Set keyless subject: scripts/set-keyless-subject.sh "https://github.com/ORG/REPO/.github/workflows/ci.yaml@refs/heads/main"
Then re-run commit/CI.
EOF
  exit 1
fi

echo "No policy placeholders found."
