#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 2 ]; then
  echo "Usage: $0 <env: production|staging|dev> <provider: aws|gke|gke-wi|aks|aks-wi|digitalocean|linode> [domain]" >&2
  exit 1
fi
ENV=$1
PROV=$2
DOMAIN=${3:-}

ROOT=$(cd "$(dirname "$0")/.." && pwd)
OVERLAY_DIR="$ROOT/clusters/${ENV}/overlays/${PROV}"
if [ ! -d "$OVERLAY_DIR" ]; then
  echo "Overlay not found: $OVERLAY_DIR" >&2
  exit 1
fi

# Optionally set domain via helper
if [ -n "$DOMAIN" ]; then
  "$ROOT/scripts/set-overlay.sh" "$ENV" "$PROV" "$DOMAIN"
fi

echo "Applying base app set for $ENV overlay: $PROV"
kubectl apply -k "$OVERLAY_DIR"

echo "Done. Monitor ArgoCD sync and health: kubectl -n argocd get applications"
