#!/usr/bin/env bash
set -euo pipefail

BRANCH=${1:-main}

# Requires GitHub CLI auth with admin permissions on the repo
if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is required. Install from https://cli.github.com/" >&2
  exit 1
fi

slug_json=$(gh repo view --json owner,name 2>/dev/null || true)
if [ -z "$slug_json" ]; then
  echo "Unable to detect repo via gh. Ensure you are in the repo and authenticated." >&2
  exit 1
fi
OWNER=$(echo "$slug_json" | jq -r .owner.login)
REPO=$(echo "$slug_json" | jq -r .name)

# Status contexts correspond to "<workflow name> / <job id>"
# Our workflow name is "CI".
read -r -d '' PAYLOAD <<'JSON'
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
JSON

echo "Setting required checks on $OWNER/$REPO branch $BRANCH..."
echo "$PAYLOAD" | gh api -X PUT "repos/$OWNER/$REPO/branches/$BRANCH/protection" --input -

echo "Done. Review branch protection at: https://github.com/$OWNER/$REPO/settings/branches"
