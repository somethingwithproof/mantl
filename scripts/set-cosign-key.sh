#!/usr/bin/env bash
set -euo pipefail

KEY_FILE=${1:?"Usage: $0 <path-to-cosign.pub>"}
POLICY=policies/kyverno/verify-image-signatures.yaml

if [ ! -f "$KEY_FILE" ]; then
  echo "Key file not found: $KEY_FILE" >&2
  exit 1
fi

# Indent key 22 spaces to align under 'publicKeys: |'
INDENTED=$(sed 's/^/                      /' "$KEY_FILE")

# Replace the block between the placeholder markers (BEGIN/END lines)
awk -v repl="$INDENTED" '
  BEGIN {inblock=0}
  /publicKeys: \|/ {print; inblock=1; print repl; getline; while ($0 !~ /END PUBLIC KEY/) { getline } ; getline; next }
  {print}
' "$POLICY" > "$POLICY.tmp"

mv "$POLICY.tmp" "$POLICY"

echo "Updated $POLICY with key from $KEY_FILE"
