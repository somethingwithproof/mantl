#!/usr/bin/env bash
#
# Set and configure Kustomize overlay for environment and provider
#
# This script selects the appropriate Kustomize overlay for a given environment
# and cloud provider combination, and optionally patches domain settings.
#
# Usage:
#   set-overlay.sh <env> <provider> [domain]
#
# Arguments:
#   env       - Environment (production|staging|dev)
#   provider  - Cloud provider (aws|gke|gke-wi|aks|aks-wi|digitalocean|linode)
#   domain    - Optional domain for ingress/gateway configuration
#
# Examples:
#   set-overlay.sh production aws
#   set-overlay.sh staging gke example.com
#   set-overlay.sh dev aks-wi test.example.com
#

set -euo pipefail

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Color output
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_RESET='\033[0m'

# Valid environments and providers
readonly VALID_ENVIRONMENTS=("production" "staging" "dev")
readonly VALID_PROVIDERS=("aws" "gke" "gke-wi" "aks" "aks-wi" "digitalocean" "linode")

#######################################
# Print error message and exit
# Arguments:
#   $1 - Error message
#######################################
err() {
  echo -e "${COLOR_RED}ERROR: $*${COLOR_RESET}" >&2
  exit 1
}

#######################################
# Print info message
# Arguments:
#   $1 - Info message
#######################################
info() {
  echo -e "${COLOR_GREEN}$*${COLOR_RESET}"
}

#######################################
# Print warning message
# Arguments:
#   $1 - Warning message
#######################################
warn() {
  echo -e "${COLOR_YELLOW}WARNING: $*${COLOR_RESET}" >&2
}

#######################################
# Display usage information
#######################################
usage() {
  cat <<EOF
Usage: ${0##*/} <env> <provider> [domain]

Select and configure Kustomize overlay for environment and cloud provider.

Arguments:
  env       Environment to deploy to
            Valid values: ${VALID_ENVIRONMENTS[*]}

  provider  Cloud provider for the deployment
            Valid values: ${VALID_PROVIDERS[*]}

  domain    Optional domain name for ingress/gateway configuration
            If provided, updates gateway and application domain settings

The script will:
  1. Validate the environment and provider combination
  2. Verify the overlay directory exists
  3. Display the overlay path and kubectl apply command
  4. Optionally patch domain settings if provided

Examples:
  # Select production AWS overlay
  ${0##*/} production aws

  # Select staging GKE overlay with custom domain
  ${0##*/} staging gke example.com

  # Select dev AKS with workload identity and domain
  ${0##*/} dev aks-wi test.example.com

Provider Notes:
  *-wi  Workload Identity variants (GKE, AKS)

Overlay Structure:
  Overlays are located at: clusters/<env>/overlays/<provider>

EOF
  exit "${1:-0}"
}

#######################################
# Validate environment argument
# Arguments:
#   $1 - Environment name
#######################################
validate_environment() {
  local env="$1"
  local valid=false

  for valid_env in "${VALID_ENVIRONMENTS[@]}"; do
    if [[ "$env" == "$valid_env" ]]; then
      valid=true
      break
    fi
  done

  if [[ "$valid" == "false" ]]; then
    err "Invalid environment: $env (valid: ${VALID_ENVIRONMENTS[*]})"
  fi
}

#######################################
# Validate provider argument
# Arguments:
#   $1 - Provider name
#######################################
validate_provider() {
  local provider="$1"
  local valid=false

  for valid_provider in "${VALID_PROVIDERS[@]}"; do
    if [[ "$provider" == "$valid_provider" ]]; then
      valid=true
      break
    fi
  done

  if [[ "$valid" == "false" ]]; then
    err "Invalid provider: $provider (valid: ${VALID_PROVIDERS[*]})"
  fi
}

#######################################
# Validate overlay directory exists
# Arguments:
#   $1 - Overlay directory path
#######################################
validate_overlay_dir() {
  local overlay_dir="$1"

  if [[ ! -d "$overlay_dir" ]]; then
    err "Overlay directory not found: $overlay_dir"
  fi

  if [[ ! -f "$overlay_dir/kustomization.yaml" ]]; then
    warn "No kustomization.yaml found in $overlay_dir"
  fi
}

#######################################
# Update domain in kustomization file
# Arguments:
#   $1 - Kustomization file path
#   $2 - Domain name
#   $3 - Component name (for logging)
#######################################
update_domain() {
  local kust_file="$1"
  local domain="$2"
  local component="$3"

  if [[ ! -f "$kust_file" ]]; then
    warn "Kustomization file not found: $kust_file"
    return 1
  fi

  if ! grep -q "domain=" "$kust_file"; then
    warn "No domain placeholder found in $kust_file"
    return 1
  fi

  # Create backup
  cp "$kust_file" "${kust_file}.bak"

  # Update domain using sed (portable across macOS and Linux)
  if sed -i.tmp "s/domain=.*/domain=${domain}/" "$kust_file" 2>/dev/null; then
    rm -f "${kust_file}.tmp" "${kust_file}.bak"
    info "✓ Updated $component domain to: $domain"
  else
    # Restore backup on failure
    mv "${kust_file}.bak" "$kust_file"
    warn "Failed to update domain in $kust_file"
    return 1
  fi
}

#######################################
# Patch domain settings if domain provided
# Arguments:
#   $1 - Environment
#   $2 - Domain name
#######################################
patch_domain_settings() {
  local env="$1"
  local domain="$2"

  info "Setting domain to: $domain"

  # Gateway sample custom wrapper (production only)
  if [[ "$env" == "production" ]]; then
    local gateway_kust="${ROOT_DIR}/clusters/production/gateway-sample-custom/kustomization.yaml"
    if [[ -f "$gateway_kust" ]]; then
      update_domain "$gateway_kust" "$domain" "gateway-sample-custom" || true
    fi
  fi

  # Example hello ingress domain
  local hello_kust="${ROOT_DIR}/apps/examples/hello/kustomization.yaml"
  if [[ -f "$hello_kust" ]]; then
    update_domain "$hello_kust" "$domain" "example hello" || true
  fi
}

#######################################
# Main function
#######################################
main() {
  # Parse arguments
  if [[ $# -lt 2 ]]; then
    usage 1
  fi

  local env="$1"
  local provider="$2"
  local domain="${3:-}"

  # Validate inputs
  validate_environment "$env"
  validate_provider "$provider"

  # Construct overlay path
  local overlay_dir="${ROOT_DIR}/clusters/${env}/overlays/${provider}"

  # Validate overlay exists
  validate_overlay_dir "$overlay_dir"

  # Display overlay information
  echo ""
  info "Overlay selected: $overlay_dir"
  echo ""
  echo "Apply with:"
  echo "  kubectl apply -k $overlay_dir"
  echo ""

  # Patch domain settings if provided
  if [[ -n "$domain" ]]; then
    echo ""
    patch_domain_settings "$env" "$domain"
    echo ""
  fi

  # Success message
  info "✓ Overlay configuration complete!"

  # Display next steps
  echo ""
  echo "Next steps:"
  echo "  1. Review the overlay: ls -la $overlay_dir"
  echo "  2. Preview changes: kubectl diff -k $overlay_dir"
  echo "  3. Apply to cluster: kubectl apply -k $overlay_dir"
  echo ""
}

# Handle --help flag
if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage 0
fi

main "$@"
