#!/usr/bin/env bash
#
# Trigger Spinnaker pipeline via webhook
#
# This script triggers a Spinnaker pipeline deployment by sending a webhook
# request with Docker image artifacts and deployment parameters.
#
# Usage:
#   spinnaker-trigger.sh
#
# Environment Variables:
#   GATE_URL     - Spinnaker Gate URL (default: https://spinnaker.example.com)
#   WEBHOOK      - Webhook name (default: app-deploy)
#   IMAGE        - Docker image tag to deploy (required)
#   ENVIRONMENT  - Deployment environment (default: staging)
#
# Examples:
#   IMAGE=myapp:v1.2.3 spinnaker-trigger.sh
#   IMAGE=myapp:v1.2.3 ENVIRONMENT=production spinnaker-trigger.sh
#   GATE_URL=https://spin.mycompany.com IMAGE=myapp:latest spinnaker-trigger.sh
#

set -euo pipefail

# Constants
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/../../.." && pwd)"

# Defaults
readonly DEFAULT_GATE_URL="https://spinnaker.example.com"
readonly DEFAULT_WEBHOOK="app-deploy"
readonly DEFAULT_ENVIRONMENT="staging"

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
Usage: ${0##*/}

Trigger Spinnaker pipeline deployment via webhook.

Environment Variables:
  GATE_URL      Spinnaker Gate API URL
                Default: $DEFAULT_GATE_URL

  WEBHOOK       Webhook identifier configured in Spinnaker
                Default: $DEFAULT_WEBHOOK

  IMAGE         Docker image tag to deploy (REQUIRED)
                Format: registry/image:tag
                Example: myregistry.io/myapp:v1.2.3

  ENVIRONMENT   Target deployment environment
                Default: $DEFAULT_ENVIRONMENT
                Common values: staging, production, dev

The script will:
  1. Validate all required environment variables
  2. Validate the image tag format
  3. Construct webhook payload with artifacts and parameters
  4. Send POST request to Spinnaker Gate API
  5. Report success or failure

Examples:
  # Deploy specific image to staging (default)
  IMAGE=myapp:v1.2.3 ${0##*/}

  # Deploy to production environment
  IMAGE=myapp:v1.2.3 ENVIRONMENT=production ${0##*/}

  # Use custom Spinnaker URL and webhook
  GATE_URL=https://spin.mycompany.com \\
  WEBHOOK=my-webhook \\
  IMAGE=myregistry.io/myapp:v1.2.3 \\
  ${0##*/}

  # Deploy latest tag to dev environment
  IMAGE=myapp:latest ENVIRONMENT=dev ${0##*/}

Webhook Payload:
  The script sends a JSON payload with:
  - artifacts: Docker image artifact information
  - parameters: Deployment environment and other parameters

Prerequisites:
  - Spinnaker must be deployed and accessible
  - Webhook must be configured in Spinnaker pipeline
  - Network access to Spinnaker Gate API
  - curl must be installed

Spinnaker Configuration:
  Configure webhook in your Spinnaker pipeline:
  1. Add a webhook trigger to your pipeline
  2. Set the webhook name (matches WEBHOOK variable)
  3. Configure expected artifacts (docker/image type)
  4. Set up parameters (deployEnv, etc.)

Troubleshooting:
  - 404 error: Webhook name not found in Spinnaker
  - 401/403 error: Authentication/authorization required
  - Connection refused: Check GATE_URL and network access
  - Timeout: Check Spinnaker Gate service health

EOF
  exit "${1:-0}"
}

#######################################
# Check if curl is installed
#######################################
check_curl() {
  if ! command -v curl >/dev/null 2>&1; then
    err "curl is required but not installed

Install curl:
  macOS:   brew install curl
  Ubuntu:  sudo apt install curl
  Windows: winget install curl.curl"
  fi
}

#######################################
# Validate required environment variables
#######################################
validate_required_vars() {
  if [[ -z "${IMAGE:-}" ]]; then
    err "IMAGE environment variable is required

Example:
  IMAGE=myregistry.io/myapp:v1.2.3 $0

Set IMAGE to the Docker image tag you want to deploy."
  fi
}

#######################################
# Validate image format
# Arguments:
#   $1 - Image tag
#######################################
validate_image_format() {
  local image="$1"

  # Basic validation: should contain at least one character
  if [[ -z "$image" ]]; then
    err "Image tag cannot be empty"
  fi

  # Check for suspicious characters
  if [[ "$image" =~ [[:space:]] ]]; then
    err "Image tag cannot contain whitespace: $image"
  fi

  # Warn if no tag specified
  if [[ ! "$image" =~ : ]]; then
    warn "Image has no tag specified: $image (will use 'latest')"
  fi

  # Validate format looks reasonable
  if [[ ! "$image" =~ ^[a-zA-Z0-9._/-]+(:([a-zA-Z0-9._-]+))?$ ]]; then
    warn "Image format may be invalid: $image"
    warn "Expected format: [registry/]image[:tag]"
  fi
}

#######################################
# Validate URL format
# Arguments:
#   $1 - URL
#######################################
validate_url() {
  local url="$1"

  if [[ ! "$url" =~ ^https?:// ]]; then
    err "Invalid URL format: $url

URL must start with http:// or https://
Example: https://spinnaker.example.com"
  fi
}

#######################################
# Create webhook payload
# Arguments:
#   $1 - Image tag
#   $2 - Environment
# Returns:
#   JSON payload string
#######################################
create_payload() {
  local image="$1"
  local environment="$2"

  # Use jq if available for safer JSON construction
  if command -v jq >/dev/null 2>&1; then
    jq -n \
      --arg image "$image" \
      --arg env "$environment" \
      '{
        "artifacts": [
          {
            "type": "docker/image",
            "name": $image
          }
        ],
        "parameters": {
          "deployEnv": $env
        }
      }'
  else
    # Fallback to manual JSON construction
    cat <<EOF
{
  "artifacts": [
    {
      "type": "docker/image",
      "name": "${image}"
    }
  ],
  "parameters": {
    "deployEnv": "${environment}"
  }
}
EOF
  fi
}

#######################################
# Send webhook request
# Arguments:
#   $1 - Gate URL
#   $2 - Webhook name
#   $3 - Payload
#######################################
send_webhook() {
  local gate_url="$1"
  local webhook="$2"
  local payload="$3"
  local webhook_url="${gate_url}/webhooks/webhook/${webhook}"

  info "Sending webhook request..."
  echo ""
  echo "URL: $webhook_url"
  echo "Payload:"
  echo "$payload" | (command -v jq >/dev/null 2>&1 && jq . || cat)
  echo ""

  # Send request with error handling
  local http_code
  local response

  response=$(curl -sSf \
    -w "\n%{http_code}" \
    -X POST \
    -H "Content-Type: application/json" \
    -d "$payload" \
    "$webhook_url" 2>&1) || {

    # Extract HTTP code if available
    http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | head -n-1)

    err "Webhook request failed

HTTP Code: $http_code
Response: $body

Possible causes:
  1. Webhook '$webhook' not configured in Spinnaker
  2. Gate URL incorrect: $gate_url
  3. Network connectivity issues
  4. Authentication required
  5. Spinnaker Gate service unavailable

Check Spinnaker configuration and logs."
  }

  # Extract HTTP code and body
  http_code=$(echo "$response" | tail -n1)
  local body=$(echo "$response" | head -n-1)

  # Validate successful response
  if [[ "$http_code" =~ ^2[0-9]{2}$ ]]; then
    info "✓ Webhook triggered successfully (HTTP $http_code)"
    if [[ -n "$body" ]]; then
      echo ""
      echo "Response:"
      echo "$body" | (command -v jq >/dev/null 2>&1 && jq . || cat)
    fi
  else
    err "Unexpected HTTP response code: $http_code

Response: $body"
  fi
}

#######################################
# Main function
#######################################
main() {
  # Load configuration from environment
  local gate_url="${GATE_URL:-$DEFAULT_GATE_URL}"
  local webhook="${WEBHOOK:-$DEFAULT_WEBHOOK}"
  local image="${IMAGE:-}"
  local environment="${ENVIRONMENT:-$DEFAULT_ENVIRONMENT}"

  echo ""
  info "Spinnaker Webhook Trigger"
  echo ""

  # Validate prerequisites
  check_curl

  # Validate required variables
  validate_required_vars

  # Validate inputs
  validate_url "$gate_url"
  validate_image_format "$image"

  # Display configuration
  echo "Configuration:"
  echo "  Gate URL:    $gate_url"
  echo "  Webhook:     $webhook"
  echo "  Image:       $image"
  echo "  Environment: $environment"
  echo ""

  # Create payload
  local payload
  payload=$(create_payload "$image" "$environment")

  # Send webhook request
  send_webhook "$gate_url" "$webhook" "$payload"

  # Success message
  echo ""
  info "✓ Pipeline trigger complete!"
  echo ""
  echo "Next steps:"
  echo "  1. Monitor pipeline execution in Spinnaker UI:"
  echo "     ${gate_url}"
  echo ""
  echo "  2. Check deployment status:"
  echo "     kubectl get pods -n ${environment}"
  echo ""
}

# Handle --help flag
if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage 0
fi

main "$@"
