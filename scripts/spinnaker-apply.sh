#!/usr/bin/env bash
#
# Apply Spinnaker manifests with selected storage overlay
#
# This script deploys Spinnaker to a Kubernetes cluster using Kustomize,
# applying the base configuration and a storage backend overlay.
#
# Usage:
#   spinnaker-apply.sh [overlay]
#   SPIN_OVERLAY=gcs spinnaker-apply.sh
#
# Arguments:
#   overlay  - Storage backend overlay (s3|gcs) (default: s3)
#              Can also be set via SPIN_OVERLAY environment variable
#
# Examples:
#   spinnaker-apply.sh
#   spinnaker-apply.sh s3
#   spinnaker-apply.sh gcs
#   SPIN_OVERLAY=gcs spinnaker-apply.sh
#

set -euo pipefail

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly DEFAULT_OVERLAY="s3"
readonly SPINNAKER_BASE="${ROOT_DIR}/addons/spinnaker/base"
readonly SPINNAKER_OVERLAYS="${ROOT_DIR}/addons/spinnaker/overlays"

# Color output
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_RESET='\033[0m'

# Valid overlay options
readonly VALID_OVERLAYS=("s3" "gcs")

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
Usage: ${0##*/} [overlay]
       SPIN_OVERLAY=<overlay> ${0##*/}

Deploy Spinnaker to Kubernetes with selected storage backend.

Arguments:
  overlay    Storage backend overlay to use (default: s3)
             Valid values: ${VALID_OVERLAYS[*]}
             Can also be set via SPIN_OVERLAY environment variable

The script will:
  1. Validate kubectl is available
  2. Verify Spinnaker base and overlay directories exist
  3. Apply Spinnaker base manifests
  4. Apply selected storage overlay manifests

Examples:
  # Deploy with default S3 storage
  ${0##*/}

  # Deploy with S3 storage (explicit)
  ${0##*/} s3

  # Deploy with GCS storage
  ${0##*/} gcs

  # Deploy with GCS storage using environment variable
  SPIN_OVERLAY=gcs ${0##*/}

Storage Backends:
  s3   - Amazon S3 compatible storage
  gcs  - Google Cloud Storage

Prerequisites:
  - kubectl must be installed and configured
  - Cluster must be accessible
  - Storage backend credentials must be configured

Configuration Files:
  Base:       ${SPINNAKER_BASE}
  S3 Overlay: ${SPINNAKER_OVERLAYS}/s3
  GCS Overlay: ${SPINNAKER_OVERLAYS}/gcs

Notes:
  - Ensure storage credentials are configured before deployment
  - S3 overlay requires AWS credentials or S3-compatible endpoint
  - GCS overlay requires Google Cloud service account

EOF
  exit "${1:-0}"
}

#######################################
# Check if kubectl is installed
#######################################
check_kubectl() {
  if ! command -v kubectl >/dev/null 2>&1; then
    err "kubectl is required but not installed

Install kubectl:
  macOS:   brew install kubectl
  Ubuntu:  sudo apt install kubectl
  Windows: winget install Kubernetes.kubectl

Documentation: https://kubernetes.io/docs/tasks/tools/"
  fi

  info "✓ kubectl found: $(kubectl version --client --short 2>/dev/null || kubectl version --client)"
}

#######################################
# Check cluster connectivity
#######################################
check_cluster_access() {
  if ! kubectl cluster-info >/dev/null 2>&1; then
    warn "Cannot connect to Kubernetes cluster"
    warn "Ensure kubeconfig is properly configured"
    warn "Current context: $(kubectl config current-context 2>/dev/null || echo 'none')"
  else
    info "✓ Cluster accessible: $(kubectl config current-context)"
  fi
}

#######################################
# Validate overlay option
# Arguments:
#   $1 - Overlay name
#######################################
validate_overlay() {
  local overlay="$1"
  local valid=false

  for valid_overlay in "${VALID_OVERLAYS[@]}"; do
    if [[ "$overlay" == "$valid_overlay" ]]; then
      valid=true
      break
    fi
  done

  if [[ "$valid" == "false" ]]; then
    err "Invalid overlay: $overlay (valid: ${VALID_OVERLAYS[*]})"
  fi
}

#######################################
# Validate Spinnaker directories exist
# Arguments:
#   $1 - Overlay name
#######################################
validate_directories() {
  local overlay="$1"

  # Check base directory
  if [[ ! -d "$SPINNAKER_BASE" ]]; then
    err "Spinnaker base directory not found: $SPINNAKER_BASE"
  fi

  if [[ ! -f "$SPINNAKER_BASE/kustomization.yaml" ]]; then
    err "No kustomization.yaml in base directory: $SPINNAKER_BASE"
  fi

  # Check overlay directory
  local overlay_dir="${SPINNAKER_OVERLAYS}/${overlay}"
  if [[ ! -d "$overlay_dir" ]]; then
    err "Overlay directory not found: $overlay_dir"
  fi

  if [[ ! -f "$overlay_dir/kustomization.yaml" ]]; then
    err "No kustomization.yaml in overlay directory: $overlay_dir"
  fi

  info "✓ Spinnaker manifests validated"
}

#######################################
# Apply Spinnaker base manifests
#######################################
apply_base() {
  info "Applying Spinnaker base manifests..."

  if kubectl apply -k "$SPINNAKER_BASE"; then
    info "✓ Base manifests applied"
  else
    err "Failed to apply base manifests"
  fi
}

#######################################
# Apply storage overlay manifests
# Arguments:
#   $1 - Overlay name
#######################################
apply_overlay() {
  local overlay="$1"
  local overlay_dir="${SPINNAKER_OVERLAYS}/${overlay}"

  info "Applying $overlay storage overlay..."

  if kubectl apply -k "$overlay_dir"; then
    info "✓ Overlay manifests applied"
  else
    err "Failed to apply overlay manifests"
  fi
}

#######################################
# Display next steps
# Arguments:
#   $1 - Overlay name
#######################################
show_next_steps() {
  local overlay="$1"

  echo ""
  info "✓ Spinnaker deployment complete!"
  echo ""
  echo "Configuration:"
  echo "  Storage backend: $overlay"
  echo "  Namespace: spinnaker"
  echo ""
  echo "Next steps:"
  echo "  1. Check deployment status:"
  echo "     kubectl get pods -n spinnaker"
  echo ""
  echo "  2. Wait for all pods to be ready:"
  echo "     kubectl wait --for=condition=ready pod -l app.kubernetes.io/part-of=spinnaker -n spinnaker --timeout=300s"
  echo ""
  echo "  3. Access Spinnaker UI:"
  echo "     kubectl port-forward -n spinnaker svc/spin-deck 9000:9000"
  echo "     Open: http://localhost:9000"
  echo ""
  echo "  4. Configure $overlay credentials (if not already done):"

  case "$overlay" in
    s3)
      echo "     kubectl create secret generic s3-credentials -n spinnaker \\"
      echo "       --from-literal=access-key-id=YOUR_ACCESS_KEY \\"
      echo "       --from-literal=secret-access-key=YOUR_SECRET_KEY"
      ;;
    gcs)
      echo "     kubectl create secret generic gcs-credentials -n spinnaker \\"
      echo "       --from-file=credentials.json=PATH_TO_SERVICE_ACCOUNT_JSON"
      ;;
  esac

  echo ""
}

#######################################
# Main function
#######################################
main() {
  # Determine overlay from argument or environment variable
  local overlay="${1:-${SPIN_OVERLAY:-$DEFAULT_OVERLAY}}"

  echo ""
  info "Spinnaker Deployment"
  echo ""

  # Validate prerequisites
  check_kubectl
  check_cluster_access

  # Validate inputs
  validate_overlay "$overlay"
  validate_directories "$overlay"

  # Apply manifests
  echo ""
  apply_base
  echo ""
  apply_overlay "$overlay"

  # Show success and next steps
  show_next_steps "$overlay"
}

# Handle --help flag
if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage 0
fi

main "$@"
