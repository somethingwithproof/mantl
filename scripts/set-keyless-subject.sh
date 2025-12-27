#!/usr/bin/env bash
set -euo pipefail

SUBJECT=${1:?"Usage: $0 <keyless-subject> (e.g., https://github.com/ORG/REPO/.github/workflows/ci.yaml@refs/heads/main)"}
POLICY=policies/kyverno/verify-image-keyless.yaml

if [ ! -f "$POLICY" ]; then
  echo "Policy not found: $POLICY" >&2
  exit 1
fi

# Replace the subject line
sed -i.bak "s#^\(\s*subject: \).*#\1\"${SUBJECT}\"#" "$POLICY" && rm "$POLICY".bak

echo "Updated keyless subject in $POLICY to: $SUBJECT"