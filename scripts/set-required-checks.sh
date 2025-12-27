#!/usr/bin/env bash
#
# Set required status checks for GitHub branch protection
#
# This script configures GitHub branch protection rules including required
# status checks, admin enforcement, and pull request review requirements.
#
# Usage:
#   set-required-checks.sh [branch]
#
# Arguments:
#   branch  - Branch name to protect (default: main)
#
# Examples:
#   set-required-checks.sh
#   set-required-checks.sh main
#   set-required-checks.sh develop
#

set -euo pipefail

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly DEFAULT_BRANCH="main"

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
Usage: ${0##*/} [branch]

Configure GitHub branch protection rules with required status checks.

Arguments:
  branch    Branch name to protect (default: main)

The script will:
  1. Verify GitHub CLI is installed and authenticated
  2. Detect the current repository
  3. Configure branch protection rules with:
     - Required status checks (CI workflow jobs)
     - Strict status check enforcement
     - Admin enforcement
     - Required pull request reviews
     - Stale review dismissal

Examples:
  # Protect main branch with default settings
  ${0##*/}

  # Protect specific branch
  ${0##*/} develop

  # Protect production branch
  ${0##*/} production

Required Status Checks:
  The following CI jobs must pass before merging:
    - CI / policy-tests
    - CI / platform-health
    - CI / kyverno-validate
    - CI / kind-smoke
    - CI / e2e-hello
    - CI / gateway-smoke
    - CI / security-scan
    - CI / sbom
    - CI / terraform

Prerequisites:
  - GitHub CLI (gh) must be installed
  - Must be authenticated with admin permissions
  - Must be run from within the repository

Installation:
  Install GitHub CLI: https://cli.github.com/

Authentication:
  gh auth login

EOF
  exit "${1:-0}"
}

#######################################
# Check if GitHub CLI is installed
#######################################
check_gh_cli() {
  if ! command -v gh >/dev/null 2>&1; then
    err "GitHub CLI (gh) is required but not installed

Install from: https://cli.github.com/

macOS:   brew install gh
Ubuntu:  sudo apt install gh
Windows: winget install GitHub.cli"
  fi

  info "✓ GitHub CLI found: $(gh --version | head -n1)"
}

#######################################
# Check if jq is installed
#######################################
check_jq() {
  if ! command -v jq >/dev/null 2>&1; then
    err "jq is required but not installed

Install jq:
  macOS:   brew install jq
  Ubuntu:  sudo apt install jq
  Windows: winget install jqlang.jq"
  fi
}

#######################################
# Get repository information
# Returns:
#   Sets OWNER and REPO variables
#######################################
get_repo_info() {
  local slug_json

  # Attempt to get repo info via gh CLI
  if ! slug_json=$(gh repo view --json owner,name 2>/dev/null); then
    err "Unable to detect repository via GitHub CLI

Possible causes:
  1. Not authenticated: Run 'gh auth login'
  2. Not in a git repository
  3. No GitHub remote configured
  4. Insufficient permissions

Current directory: $(pwd)
Git remote: $(git remote -v 2>/dev/null || echo 'none')"
  fi

  # Validate JSON output
  if [[ -z "$slug_json" ]]; then
    err "Received empty response from GitHub API"
  fi

  # Parse owner and repo
  OWNER=$(echo "$slug_json" | jq -r .owner.login)
  REPO=$(echo "$slug_json" | jq -r .name)

  # Validate parsed values
  if [[ -z "$OWNER" || "$OWNER" == "null" ]]; then
    err "Failed to parse repository owner from GitHub response"
  fi

  if [[ -z "$REPO" || "$REPO" == "null" ]]; then
    err "Failed to parse repository name from GitHub response"
  fi

  info "✓ Repository detected: $OWNER/$REPO"
}

#######################################
# Validate branch name
# Arguments:
#   $1 - Branch name
#######################################
validate_branch() {
  local branch="$1"

  # Basic validation
  if [[ -z "$branch" ]]; then
    err "Branch name cannot be empty"
  fi

  # Check for invalid characters
  if [[ "$branch" =~ [[:space:]] ]]; then
    err "Branch name cannot contain whitespace: $branch"
  fi

  # Warn if branch doesn't exist locally
  if git rev-parse --verify "$branch" >/dev/null 2>&1; then
    info "✓ Branch exists locally: $branch"
  else
    warn "Branch '$branch' not found locally (this is OK if it exists remotely)"
  fi
}

#######################################
# Create branch protection payload
# Returns:
#   Branch protection JSON payload
#######################################
create_protection_payload() {
  cat <<'EOF'
{
  "required_status_checks": {
    "strict": true,
    "contexts": [
      "CI / policy-tests",
      "CI / platform-health",
      "CI / kyverno-validate",
      "CI / kind-smoke",
      "CI / e2e-hello",
      "CI / gateway-smoke",
      "CI / security-scan",
      "CI / sbom",
      "CI / terraform"
    ]
  },
  "enforce_admins": true,
  "required_pull_request_reviews": {
    "dismiss_stale_reviews": true
  },
  "restrictions": null
}
EOF
}

#######################################
# Apply branch protection
# Arguments:
#   $1 - Owner
#   $2 - Repository
#   $3 - Branch name
#######################################
apply_branch_protection() {
  local owner="$1"
  local repo="$2"
  local branch="$3"
  local payload

  payload=$(create_protection_payload)

  info "Setting required checks on $owner/$repo branch: $branch"
  echo ""
  echo "Protection settings:"
  echo "$payload" | jq .
  echo ""

  # Apply protection rules via GitHub API
  if echo "$payload" | gh api -X PUT "repos/$owner/$repo/branches/$branch/protection" --input - >/dev/null 2>&1; then
    info "✓ Branch protection rules applied successfully"
  else
    err "Failed to apply branch protection rules

Possible causes:
  1. Insufficient permissions (requires admin access)
  2. Branch does not exist on remote
  3. API rate limit exceeded
  4. Network connectivity issues

Check your permissions: gh auth status"
  fi
}

#######################################
# Main function
#######################################
main() {
  # Parse arguments
  local branch="${1:-$DEFAULT_BRANCH}"

  echo ""
  info "GitHub Branch Protection Configuration"
  echo ""

  # Validate prerequisites
  check_gh_cli
  check_jq

  # Get repository information
  get_repo_info

  # Validate branch
  validate_branch "$branch"

  # Apply branch protection
  echo ""
  apply_branch_protection "$OWNER" "$REPO" "$branch"

  # Success message
  echo ""
  info "✓ Branch protection configured successfully!"
  echo ""
  echo "Review settings at:"
  echo "  https://github.com/$OWNER/$REPO/settings/branch_protection_rules"
  echo ""
  echo "Branch protection rules:"
  echo "  ✓ Required status checks enabled"
  echo "  ✓ Strict mode enabled (requires up-to-date branches)"
  echo "  ✓ Admin enforcement enabled"
  echo "  ✓ Pull request reviews required"
  echo "  ✓ Stale review dismissal enabled"
  echo ""
}

# Handle --help flag
if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage 0
fi

main "$@"
