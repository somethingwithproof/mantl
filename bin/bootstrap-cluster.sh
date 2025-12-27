#!/usr/bin/env bash
#
# Bootstrap a Mantl cluster with ArgoCD applications
#
# This script applies the base ArgoCD application set for a given environment
# and cloud provider overlay.
#
# Usage:
#   bootstrap-cluster.sh <environment> <provider> [domain]
#
# Arguments:
#   environment  - Deployment environment (production|staging|dev)
#   provider     - Cloud provider (aws|gke|gke-wi|aks|aks-wi|digitalocean|linode)
#   domain       - Optional domain name to configure
#
# Examples:
#   bootstrap-cluster.sh production aws
#   bootstrap-cluster.sh staging gke example.com
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
# Print warning message
# Arguments:
#   $1 - Warning message
#######################################
warn() {
  echo -e "${COLOR_YELLOW}WARNING: $*${COLOR_RESET}" >&2
}

#######################################
# Print info message
# Arguments:
#   $1 - Info message
#######################################
info() {
  echo -e "${COLOR_GREEN}INFO: $*${COLOR_RESET}"
}

#######################################
# Display usage information
#######################################
usage() {
  cat <<EOF
Usage: ${0##*/} <environment> <provider> [domain]

Bootstrap a Mantl cluster with ArgoCD applications.

Arguments:
  environment    Deployment environment (production|staging|dev)
  provider       Cloud provider overlay
                 (aws|gke|gke-wi|aks|aks-wi|digitalocean|linode)
  domain         Optional domain name to configure

Examples:
  ${0##*/} production aws
  ${0##*/} staging gke example.com

Environment Variables:
  KUBECONFIG     Path to kubeconfig file (default: ~/.kube/config)

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
  case "$env" in
    production|staging|dev)
      return 0
      ;;
    *)
      err "Invalid environment: $env. Must be one of: production, staging, dev"
      ;;
  esac
}

#######################################
# Validate provider argument
# Arguments:
#   $1 - Provider name
#######################################
validate_provider() {
  local provider="$1"
  case "$provider" in
    aws|gke|gke-wi|aks|aks-wi|digitalocean|linode)
      return 0
      ;;
    *)
      err "Invalid provider: $provider. Must be one of: aws, gke, gke-wi, aks, aks-wi, digitalocean, linode"
      ;;
  esac
}

#######################################
# Main function
#######################################
main() {
  # Parse arguments
  if [[ $# -lt 2 ]]; then
    usage 1
  fi

  local environment="$1"
  local provider="$2"
  local domain="${3:-}"

  # Validate inputs
  validate_environment "$environment"
  validate_provider "$provider"

  # Check kubectl
  if ! command -v kubectl >/dev/null 2>&1; then
    err "kubectl is required but not found. Please install kubectl."
  fi

  # Verify overlay directory exists
  local overlay_dir="${ROOT_DIR}/clusters/${environment}/overlays/${provider}"
  if [[ ! -d "$overlay_dir" ]]; then
    err "Overlay directory not found: $overlay_dir"
  fi

  # Optionally configure domain
  if [[ -n "$domain" ]]; then
    info "Configuring domain: $domain"
    if [[ -x "${ROOT_DIR}/scripts/set-overlay.sh" ]]; then
      "${ROOT_DIR}/scripts/set-overlay.sh" "$environment" "$provider" "$domain"
    else
      warn "set-overlay.sh not found or not executable, skipping domain configuration"
    fi
  fi

  # Apply overlay
  info "Applying ArgoCD applications for $environment/$provider"
  if kubectl apply -k "$overlay_dir"; then
    info "Successfully applied overlay: $overlay_dir"
    echo ""
    info "Monitor ArgoCD sync and health:"
    echo "  kubectl -n argocd get applications"
    echo "  kubectl -n argocd get applications -o wide"
  else
    err "Failed to apply overlay: $overlay_dir"
  fi
}

# Handle --help flag
if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage 0
fi

main "$@"
