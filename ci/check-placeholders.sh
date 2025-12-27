#!/usr/bin/env bash
#
# Check for placeholder values in Kyverno policies
#
# This script scans Kyverno policy files for placeholder values that should
# be replaced with actual configuration before committing.
#
# This is typically used as a pre-commit hook to prevent committing
# placeholder keys and configuration.
#
# Usage:
#   check-placeholders.sh
#
# Exit codes:
#   0 - No placeholders found
#   1 - Placeholders detected or error occurred
#

set -euo pipefail

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

# Placeholder patterns to detect
readonly -a PLACEHOLDER_PATTERNS=(
  'ECR_PLACEHOLDER_KEY'
  'GCR_PLACEHOLDER_KEY'
  'AnPlaceholderKey'
  'YOUR_ORG'
)

# Color output
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_YELLOW='\033[1;33m'
readonly COLOR_RESET='\033[0m'

#######################################
# Print error message
# Arguments:
#   $1 - Error message
#######################################
err() {
  echo -e "${COLOR_RED}$*${COLOR_RESET}" >&2
}

#######################################
# Print success message
# Arguments:
#   $1 - Success message
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
  echo -e "${COLOR_YELLOW}$*${COLOR_RESET}" >&2
}

#######################################
# Display usage information
#######################################
usage() {
  cat <<EOF
Usage: ${0##*/}

Check Kyverno policy files for placeholder values.

This script scans all Kyverno policy files and reports any placeholder
values that should be replaced with actual configuration.

Placeholder patterns checked:
$(printf '  - %s\n' "${PLACEHOLDER_PATTERNS[@]}")

Exit codes:
  0    No placeholders found
  1    Placeholders detected

Examples:
  ${0##*/}
  pre-commit run check-placeholders

EOF
  exit "${1:-0}"
}

#######################################
# Check for placeholders in files
# Returns:
#   0 if no placeholders found
#   1 if placeholders found
#######################################
check_placeholders() {
  local fail=0
  local files

  cd "$ROOT_DIR" || {
    err "ERROR: Cannot change to root directory: $ROOT_DIR"
    return 1
  }

  # Get list of Kyverno policy files
  # Only scan Kyverno policy files to avoid false positives in docs/overlays
  files=$(git ls-files 'policies/kyverno/*.yaml' 2>/dev/null || true)

  if [[ -z "$files" ]]; then
    warn "No Kyverno policy files found in policies/kyverno/"
    return 0
  fi

  # Check each file for placeholders
  for file in $files; do
    if [[ ! -f "$file" ]]; then
      continue
    fi

    for pattern in "${PLACEHOLDER_PATTERNS[@]}"; do
      if grep -q "$pattern" "$file"; then
        err "Placeholder '$pattern' found in $file"
        fail=1
      fi
    done
  done

  return $fail
}

#######################################
# Display remediation instructions
#######################################
show_remediation() {
  cat <<'EOF'

╔══════════════════════════════════════════════════════════════════╗
║           Kyverno Policy Placeholders Detected                   ║
╚══════════════════════════════════════════════════════════════════╝

Fixes:
  1. Replace Cosign public keys:
     ./scripts/set-cosign-key.sh /path/to/cosign.pub

  2. Set keyless subject for GitHub Actions:
     ./scripts/set-keyless-subject.sh \
       "https://github.com/ORG/REPO/.github/workflows/ci.yaml@refs/heads/main"

  3. Replace YOUR_ORG with your actual organization name

After fixing, re-run this check or attempt your commit again.

EOF
}

#######################################
# Main function
#######################################
main() {
  if check_placeholders; then
    info "✓ No policy placeholders found"
    return 0
  else
    show_remediation
    return 1
  fi
}

# Handle --help flag
if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage 0
fi

main "$@"
