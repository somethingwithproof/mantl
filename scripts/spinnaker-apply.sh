#!/usr/bin/env bash
set -euo pipefail

# Apply Spinnaker base and selected overlay
# Usage:
#   SPIN_OVERLAY=s3 ./scripts/spinnaker-apply.sh

OVERLAY=${SPIN_OVERLAY:-s3}
ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

if ! command -v kubectl >/dev/null 2>&1; then
  echo "kubectl is required" >&2
  exit 1
fi

kubectl apply -k "$ROOT_DIR/addons/spinnaker/base"

case "$OVERLAY" in
  s3)
    kubectl apply -k "$ROOT_DIR/addons/spinnaker/overlays/s3"
    ;;
  gcs)
    kubectl apply -k "$ROOT_DIR/addons/spinnaker/overlays/gcs"
    ;;
  *)
    echo "Unknown overlay: $OVERLAY (supported: s3, gcs)" >&2
    exit 2
    ;;
 esac

echo "Spinnaker manifests applied (overlay=$OVERLAY)."
