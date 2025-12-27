#!/usr/bin/env bash
#
# Set Cosign public key in Kyverno image verification policy
#
# This script updates the Kyverno image verification policy with a Cosign
# public key for verifying container image signatures.
#
# Usage:
#   set-cosign-key.sh <path-to-cosign.pub>
#
# Arguments:
#   path-to-cosign.pub  - Path to Cosign public key file
#
# Examples:
#   set-cosign-key.sh ~/.cosign/cosign.pub
#   set-cosign-key.sh /tmp/mykey.pub
#

set -euo pipefail

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly POLICY_FILE="${ROOT_DIR}/policies/kyverno/verify-image-signatures.yaml"

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
Usage: ${0##*/} <path-to-cosign.pub>

Update Kyverno image verification policy with Cosign public key.

Arguments:
  path-to-cosign.pub    Path to Cosign public key file (PEM format)

The script will:
  1. Validate the public key file exists and is readable
  2. Backup the existing policy file
  3. Update the policy with the new public key
  4. Verify the update was successful

Examples:
  ${0##*/} ~/.cosign/cosign.pub
  ${0##*/} /tmp/mykey.pub

Cosign Key Generation:
  To generate a new key pair:
    cosign generate-key-pair

Policy File:
  ${POLICY_FILE}

EOF
  exit "${1:-0}"
}

#######################################
# Validate key file
# Arguments:
#   $1 - Path to key file
#######################################
validate_key_file() {
  local key_file="$1"

  # Check if file exists
  if [[ ! -f "$key_file" ]]; then
    err "Key file not found: $key_file"
  fi

  # Check if file is readable
  if [[ ! -r "$key_file" ]]; then
    err "Key file is not readable: $key_file"
  fi

  # Basic validation of PEM format
  if ! grep -q "BEGIN PUBLIC KEY" "$key_file"; then
    warn "Key file does not appear to be in PEM format"
    warn "Expected format: BEGIN PUBLIC KEY...END PUBLIC KEY"
  fi
}

#######################################
# Update policy file with key
# Arguments:
#   $1 - Path to key file
#######################################
update_policy() {
  local key_file="$1"
  local tmp_file="${POLICY_FILE}.tmp"
  local backup_file="${POLICY_FILE}.backup.$(date +%Y%m%d_%H%M%S)"

  # Check if policy file exists
  if [[ ! -f "$POLICY_FILE" ]]; then
    err "Policy file not found: $POLICY_FILE"
  fi

  # Create backup
  info "Creating backup: $backup_file"
  cp "$POLICY_FILE" "$backup_file"

  # Indent key content (22 spaces to align under 'publicKeys: |')
  local indented_key
  indented_key=$(sed 's/^/                      /' "$key_file")

  # Replace the block between the placeholder markers
  awk -v repl="$indented_key" '
    BEGIN {inblock=0}
    /publicKeys: \|/ {
      print
      inblock=1
      print repl
      getline
      while ($0 !~ /END PUBLIC KEY/) {
        getline
      }
      getline
      next
    }
    {print}
  ' "$POLICY_FILE" > "$tmp_file"

  # Validate the temporary file was created
  if [[ ! -s "$tmp_file" ]]; then
    rm -f "$tmp_file"
    err "Failed to create updated policy file"
  fi

  # Move temporary file to final location
  mv "$tmp_file" "$POLICY_FILE"

  info "✓ Updated $POLICY_FILE with key from $key_file"
  info "✓ Backup saved to: $backup_file"
}

#######################################
# Main function
#######################################
main() {
  # Parse arguments
  if [[ $# -ne 1 ]]; then
    usage 1
  fi

  local key_file="$1"

  # Validate inputs
  validate_key_file "$key_file"

  # Update policy
  update_policy "$key_file"

  # Success message
  echo ""
  info "Cosign key successfully configured!"
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
