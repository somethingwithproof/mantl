#!/usr/bin/env bash
set -euo pipefail

if [ $# -lt 2 ]; then
  echo "Usage: $0 <env: production|staging|dev> <provider: aws|gke|gke-wi|aks|aks-wi|digitalocean|linode> [domain]" >&2
  exit 1
fi
ENV=$1
PROV=$2
DOMAIN=${3:-}

OVERLAY_DIR="clusters/${ENV}/overlays/${PROV}"
if [ ! -d "$OVERLAY_DIR" ]; then
  echo "Overlay not found: $OVERLAY_DIR" >&2
  exit 1
fi

echo "Overlay selected: $OVERLAY_DIR"
echo "Apply with: kubectl apply -k $OVERLAY_DIR"

# Optionally patch domain for gateway sample and example hello
if [ -n "$DOMAIN" ]; then
  echo "Setting domain to: $DOMAIN"
  # Gateway sample custom wrapper (production only)
  if [ "$ENV" = "production" ] && [ -f clusters/production/gateway-sample-custom/kustomization.yaml ]; then
    if grep -q "domain=" clusters/production/gateway-sample-custom/kustomization.yaml; then
      sed -i.bak "s/domain=.*/domain=${DOMAIN}/" clusters/production/gateway-sample-custom/kustomization.yaml && rm clusters/production/gateway-sample-custom/kustomization.yaml.bak
      echo "Updated gateway-sample-custom domain"
    fi
  fi
  # Example hello ingress domain
  if [ -f applications/examples/hello/kustomization.yaml ]; then
    if grep -q "domain=" applications/examples/hello/kustomization.yaml; then
      sed -i.bak "s/domain=.*/domain=${DOMAIN}/" applications/examples/hello/kustomization.yaml && rm applications/examples/hello/kustomization.yaml.bak
      echo "Updated example hello domain"
    fi
  fi
fi
