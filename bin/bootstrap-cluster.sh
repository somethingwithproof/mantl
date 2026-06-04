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
# Options:
#   --dry-run    - Show what would be applied without making changes
#   --rollback   - Rollback to previous state if available
#   --verbose    - Enable verbose output
#   --help       - Show this help message
#
# Examples:
#   bootstrap-cluster.sh production aws
#   bootstrap-cluster.sh staging gke example.com
#   bootstrap-cluster.sh --dry-run staging gke
#

set -euo pipefail

# Constants
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly ROOT_DIR
readonly LOG_DIR="${ROOT_DIR}/.bootstrap-logs"
readonly BACKUP_DIR="${ROOT_DIR}/.bootstrap-backups"
TIMESTAMP="$(date +%Y%m%d-%H%M%S)"
readonly TIMESTAMP
readonly LOG_FILE="${LOG_DIR}/bootstrap-${TIMESTAMP}.log"

# State
DRY_RUN=false
VERBOSE=false
ROLLBACK=false
BACKUP_FILE=""
APPLIED_RESOURCES=()

# Color output
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_BLUE='\033[0;34m'
readonly COLOR_RESET='\033[0m'

#######################################
# Initialize logging
#######################################
init_logging() {
  mkdir -p "$LOG_DIR" "$BACKUP_DIR"
  exec > >(tee -a "$LOG_FILE") 2>&1
  log "INFO" "Bootstrap started at $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  log "INFO" "Log file: $LOG_FILE"
}

#######################################
# Log message with timestamp and level
# Arguments:
#   $1 - Log level (INFO, WARN, ERROR, DEBUG)
#   $2 - Message
#######################################
log() {
  local level="$1"
  shift
  local message="$*"
  local timestamp
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "[${timestamp}] [${level}] ${message}"
}

#######################################
# Print error message and exit
# Arguments:
#   $1 - Error message
#######################################
err() {
  echo -e "${COLOR_RED}ERROR: $*${COLOR_RESET}" >&2
  log "ERROR" "$*"
  exit 1
}

#######################################
# Print warning message
# Arguments:
#   $1 - Warning message
#######################################
warn() {
  echo -e "${COLOR_YELLOW}WARNING: $*${COLOR_RESET}" >&2
  log "WARN" "$*"
}

#######################################
# Print info message
# Arguments:
#   $1 - Info message
#######################################
info() {
  echo -e "${COLOR_GREEN}INFO: $*${COLOR_RESET}"
  log "INFO" "$*"
}

#######################################
# Print debug message (only in verbose mode)
# Arguments:
#   $1 - Debug message
#######################################
debug() {
  if [[ "$VERBOSE" == "true" ]]; then
    echo -e "${COLOR_BLUE}DEBUG: $*${COLOR_RESET}"
    log "DEBUG" "$*"
  fi
}

#######################################
# Cleanup handler for script exit
#######################################
cleanup() {
  local exit_code=$?
  if [[ $exit_code -ne 0 ]]; then
    warn "Script exited with code $exit_code"
    if [[ -n "$BACKUP_FILE" ]] && [[ -f "$BACKUP_FILE" ]]; then
      warn "A backup is available at: $BACKUP_FILE"
      warn "To rollback, run: $0 --rollback"
    fi
  fi
  log "INFO" "Bootstrap completed at $(date -u +%Y-%m-%dT%H:%M:%SZ) with exit code $exit_code"
}

#######################################
# Error handler for trap
# Arguments:
#   $1 - Line number where error occurred
#######################################
handle_error() {
  local line_number=$1
  log "ERROR" "Error occurred at line $line_number"
  echo -e "${COLOR_RED}Error occurred at line $line_number${COLOR_RESET}" >&2

  if [[ ${#APPLIED_RESOURCES[@]} -gt 0 ]]; then
    warn "The following resources were applied before the error:"
    for resource in "${APPLIED_RESOURCES[@]}"; do
      echo "  - $resource"
    done
    echo ""
    warn "Consider running with --rollback to restore previous state"
  fi
}

# Set up traps
trap cleanup EXIT
trap 'handle_error ${LINENO}' ERR

#######################################
# Display usage information
#######################################
usage() {
  cat <<EOF
Usage: ${0##*/} [options] <environment> <provider> [domain]

Bootstrap a Mantl cluster with ArgoCD applications.

Arguments:
  environment    Deployment environment (production|staging|dev)
  provider       Cloud provider overlay
                 (aws|gke|gke-wi|aks|aks-wi|digitalocean|linode)
  domain         Optional domain name to configure

Options:
  --dry-run      Show what would be applied without making changes
  --rollback     Rollback to previous state if available
  --verbose      Enable verbose output
  -h, --help     Show this help message

Examples:
  ${0##*/} production aws
  ${0##*/} staging gke example.com
  ${0##*/} --dry-run staging gke

Environment Variables:
  KUBECONFIG     Path to kubeconfig file (default: ~/.kube/config)

Logs are stored in: $LOG_DIR
Backups are stored in: $BACKUP_DIR

EOF
  exit "${1:-0}"
}

#######################################
# Check prerequisites
#######################################
check_prerequisites() {
  info "Checking prerequisites..."

  # Check kubectl
  if ! command -v kubectl >/dev/null 2>&1; then
    err "kubectl is required but not found. Please install kubectl."
  fi
  debug "kubectl found: $(command -v kubectl)"

  # Check kustomize
  if ! command -v kustomize >/dev/null 2>&1; then
    warn "kustomize not found. Using kubectl kustomize instead."
  else
    debug "kustomize found: $(command -v kustomize)"
  fi

  # Check cluster connectivity
  if ! kubectl cluster-info >/dev/null 2>&1; then
    err "Cannot connect to Kubernetes cluster. Check your KUBECONFIG."
  fi
  debug "Cluster connectivity verified"

  # Check ArgoCD namespace exists
  if ! kubectl get namespace argocd >/dev/null 2>&1; then
    warn "ArgoCD namespace does not exist. It will be created."
  fi

  info "Prerequisites check passed"
}

#######################################
# Create backup of current state
# Arguments:
#   $1 - Environment name
#   $2 - Provider name
#######################################
create_backup() {
  local environment="$1"
  local provider="$2"

  BACKUP_FILE="${BACKUP_DIR}/backup-${environment}-${provider}-${TIMESTAMP}.yaml"

  info "Creating backup of current state..."

  # Backup ArgoCD applications if they exist
  if kubectl get applications.argoproj.io -n argocd >/dev/null 2>&1; then
    kubectl get applications.argoproj.io -n argocd -o yaml > "$BACKUP_FILE" 2>/dev/null || true
    if [[ ! -f "$BACKUP_FILE" ]] || [[ ! -s "$BACKUP_FILE" ]]; then
      warn "Backup file was not created or is empty: $BACKUP_FILE"
      BACKUP_FILE=""
    else
      debug "Backup created: $BACKUP_FILE"
    fi
  else
    debug "No existing ArgoCD applications to backup"
  fi
}

#######################################
# Rollback to previous state
#######################################
do_rollback() {
  info "Looking for available backups..."

  local latest_backup
  # Backup names are controlled (backup-*.yaml), so ls -t by mtime is safe here.
  # shellcheck disable=SC2012
  latest_backup=$(ls -t "${BACKUP_DIR}"/backup-*.yaml 2>/dev/null | head -1 || true)

  if [[ -z "$latest_backup" ]] || [[ ! -f "$latest_backup" ]]; then
    err "No backup files found in $BACKUP_DIR"
  fi

  info "Found backup: $latest_backup"
  info "Rolling back to previous state..."

  if [[ "$DRY_RUN" == "true" ]]; then
    info "[DRY-RUN] Would apply: $latest_backup"
    kubectl apply -f "$latest_backup" --dry-run=client
  else
    kubectl apply -f "$latest_backup"
    info "Rollback completed successfully"
  fi
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
    aws|gke|gke-wi|aks|aks-wi|digitalocean|linode|local)
      return 0
      ;;
    *)
      err "Invalid provider: $provider." \
        "Must be one of: aws, gke, gke-wi, aks, aks-wi, digitalocean, linode, local"
      ;;
  esac
}

#######################################
# Apply overlay with progress tracking
# Arguments:
#   $1 - Overlay directory
#######################################
apply_overlay() {
  local overlay_dir="$1"

  info "Applying ArgoCD applications from: $overlay_dir"

  if [[ "$DRY_RUN" == "true" ]]; then
    info "[DRY-RUN] Would apply the following resources:"
    kubectl apply -k "$overlay_dir" --dry-run=client
    return 0
  fi

  # Apply and track resources
  local output
  output=$(kubectl apply -k "$overlay_dir" 2>&1)
  echo "$output"

  # Parse applied resources
  while IFS= read -r line; do
    if [[ "$line" =~ (created|configured|unchanged) ]]; then
      APPLIED_RESOURCES+=("$line")
    fi
  done <<< "$output"

  info "Successfully applied ${#APPLIED_RESOURCES[@]} resources"
}

#######################################
# Wait for ArgoCD sync
# Arguments:
#   $1 - Timeout in seconds (default: 300)
#######################################
wait_for_sync() {
  local timeout="${1:-300}"
  local interval=10
  local elapsed=0

  info "Waiting for ArgoCD applications to sync (timeout: ${timeout}s)..."

  while [[ $elapsed -lt $timeout ]]; do
    local not_synced
    local jsonpath='{.items[?(@.status.sync.status!="Synced")].metadata.name}'
    not_synced=$(kubectl get applications.argoproj.io -n argocd \
      -o jsonpath="$jsonpath" 2>/dev/null || true)

    if [[ -z "$not_synced" ]]; then
      info "All ArgoCD applications are synced"
      return 0
    fi

    debug "Waiting for sync: $not_synced"
    sleep $interval
    elapsed=$((elapsed + interval))
  done

  warn "Timeout waiting for ArgoCD sync after ${timeout}s"
  return 1
}

#######################################
# Main function
#######################################
main() {
  local environment=""
  local provider=""
  local domain=""

  # Parse options
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --dry-run)
        DRY_RUN=true
        shift
        ;;
      --rollback)
        ROLLBACK=true
        shift
        ;;
      --verbose)
        VERBOSE=true
        shift
        ;;
      -h|--help)
        usage 0
        ;;
      -*)
        err "Unknown option: $1"
        ;;
      *)
        if [[ -z "$environment" ]]; then
          environment="$1"
        elif [[ -z "$provider" ]]; then
          provider="$1"
        else
          domain="$1"
        fi
        shift
        ;;
    esac
  done

  # Initialize logging
  init_logging

  # Handle rollback
  if [[ "$ROLLBACK" == "true" ]]; then
    do_rollback
    exit 0
  fi

  # Validate required arguments
  if [[ -z "$environment" ]] || [[ -z "$provider" ]]; then
    usage 1
  fi

  # Validate inputs
  validate_environment "$environment"
  validate_provider "$provider"

  # Check prerequisites
  check_prerequisites

  # Verify overlay directory exists
  local overlay_dir="${ROOT_DIR}/clusters/${environment}/overlays/${provider}"
  if [[ ! -d "$overlay_dir" ]]; then
    # Try without overlays subdirectory
    overlay_dir="${ROOT_DIR}/clusters/${environment}"
    if [[ ! -d "$overlay_dir" ]]; then
      err "Overlay directory not found: ${ROOT_DIR}/clusters/${environment}/overlays/${provider}"
    fi
  fi

  debug "Using overlay directory: $overlay_dir"

  # Create backup before making changes
  if [[ "$DRY_RUN" != "true" ]]; then
    create_backup "$environment" "$provider"
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
  apply_overlay "$overlay_dir"

  # Wait for sync if not dry-run
  if [[ "$DRY_RUN" != "true" ]]; then
    wait_for_sync 300 || true
  fi

  echo ""
  info "Bootstrap completed successfully!"
  echo ""
  echo "Next steps:"
  echo "  1. Monitor ArgoCD sync status:"
  echo "     kubectl -n argocd get applications"
  echo ""
  echo "  2. Access ArgoCD UI:"
  echo "     kubectl port-forward svc/argocd-server -n argocd 8080:443"
  echo "     Open: https://localhost:8080"
  echo ""
  echo "  3. Get ArgoCD admin password:"
  echo "     kubectl -n argocd get secret argocd-initial-admin-secret \\"
  echo "       -o jsonpath='{.data.password}' | base64 -d"
}

main "$@"
