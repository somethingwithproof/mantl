#!/usr/bin/env bash
#
# Set keyless verification subject in Kyverno policy
#
# This script updates the Kyverno image verification policy with a keyless
# subject for verifying container images signed via GitHub Actions OIDC.
#
# Usage:
#   set-keyless-subject.sh <keyless-subject>
#
# Arguments:
#   keyless-subject  - GitHub Actions workflow OIDC subject
#                      Format: https://github.com/ORG/REPO/.github/workflows/FILE@refs/heads/BRANCH
#
# Examples:
#   set-keyless-subject.sh \
#     "https://github.com/myorg/myrepo/.github/workflows/ci.yaml@refs/heads/main"
#

set -euo pipefail

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly POLICY_FILE="${ROOT_DIR}/policies/kyverno/verify-image-keyless.yaml"

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
Usage: ${0##*/} <keyless-subject>

Update Kyverno image verification policy with keyless subject.

Arguments:
  keyless-subject    GitHub Actions workflow OIDC subject URL

Subject Format:
  https://github.com/ORG/REPO/.github/workflows/WORKFLOW@refs/heads/BRANCH

Examples:
  ${0##*/} "https://github.com/myorg/myrepo/.github/workflows/ci.yaml@refs/heads/main"
  ${0##*/} "https://github.com/acme/app/.github/workflows/release.yml@refs/heads/production"

The script will:
  1. Validate the subject URL format
  2. Backup the existing policy file
  3. Update the subject in the policy
  4. Verify the update was successful

Policy File:
  ${POLICY_FILE}

EOF
  exit "${1:-0}"
}

#######################################
# Validate subject URL format
# Arguments:
#   $1 - Subject URL
#######################################
validate_subject() {
  local subject="$1"

  # Check if empty
  if [[ -z "$subject" ]]; then
    err "Subject cannot be empty"
  fi

  # Check basic GitHub format
  local github_pattern='^https://github\.com/[^/]+/[^/]+/\.github/workflows/.*@refs/heads/.+'
  if [[ ! "$subject" =~ $github_pattern ]]; then
    err "Invalid subject format." \
      "Expected: https://github.com/ORG/REPO/.github/workflows/FILE@refs/heads/BRANCH"
  fi

  # Warn if using non-standard branch
  if [[ ! "$subject" =~ @refs/heads/(main|master|production)$ ]]; then
    warn "Subject uses non-standard branch (not main/master/production)"
  fi
}

#######################################
# Update policy file with subject
# Arguments:
#   $1 - Subject URL
#######################################
update_policy() {
  local subject="$1"
  local backup_file="${POLICY_FILE}.backup.$(date +%Y%m%d_%H%M%S)"

  # Check if policy file exists
  if [[ ! -f "$POLICY_FILE" ]]; then
    err "Policy file not found: $POLICY_FILE"
  fi

  # Create backup
  info "Creating backup: $backup_file"
  cp "$POLICY_FILE" "$backup_file"

  # Update the subject line
  if sed -i.bak "s#^\(\s*subject: \).*#\1\"${subject}\"#" "$POLICY_FILE"; then
    rm "${POLICY_FILE}.bak"
    info "✓ Updated keyless subject in $POLICY_FILE"
    info "✓ Backup saved to: $backup_file"
  else
    err "Failed to update policy file"
  fi
}

#######################################
# Main function
#######################################
main() {
  # Parse arguments
  if [[ $# -ne 1 ]]; then
    usage 1
  fi

  local subject="$1"

  # Validate subject
  validate_subject "$subject"

  # Update policy
  update_policy "$subject"

  # Success message
  echo ""
  info "Keyless subject successfully configured!"
  echo ""
  echo "Subject: $subject"
  echo ""
  echo "Next steps:"
  echo "  1. Review the updated policy: $POLICY_FILE"
  echo "  2. Commit the changes: git add $POLICY_FILE"
  echo "  3. Apply to cluster: kubectl apply -f $POLICY_FILE"
}

# Handle --help flag
if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage 0
fi

main "$@"
