#!/usr/bin/env bash
# SPDX-License-Identifier: Apache-2.0
#
# Create a new service from template
#
# This script scaffolds a new microservice by copying the base template
# and updating the service name in the manifests.
#
# Usage:
#   create-service.sh <service-name>
#
# Arguments:
#   service-name  - Name of the new service (lowercase, alphanumeric, hyphens)
#
# Examples:
#   create-service.sh user-api
#   create-service.sh payment-service
#

set -euo pipefail

# Constants
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly SCRIPT_DIR
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly ROOT_DIR
readonly TEMPLATE_DIR="${ROOT_DIR}/applications/templates/web-service/base"

# Color output
readonly COLOR_RED='\033[0;31m'
readonly COLOR_GREEN='\033[0;32m'
readonly COLOR_BLUE='\033[0;34m'
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
# Print step message
# Arguments:
#   $1 - Step message
#######################################
step() {
  echo -e "${COLOR_BLUE}==>${COLOR_RESET} $*"
}

#######################################
# Display usage information
#######################################
usage() {
  cat <<EOF
Usage: ${0##*/} <service-name>

Create a new microservice from template.

Arguments:
  service-name    Name of the new service
                  (lowercase, alphanumeric, hyphens only)

Examples:
  ${0##*/} user-api
  ${0##*/} payment-service

The script will:
  1. Copy the template from applications/templates/web-service/base
  2. Create new service directory at applications/services/<service-name>
  3. Update service name in deployment.yaml and service.yaml
  4. Provide next steps for customization

EOF
  exit "${1:-0}"
}

#######################################
# Validate service name
# Arguments:
#   $1 - Service name
#######################################
validate_service_name() {
  local name="$1"

  # Check if name is empty
  if [[ -z "$name" ]]; then
    err "Service name cannot be empty"
  fi

  # Check if name matches pattern (lowercase, alphanumeric, hyphens)
  if [[ ! "$name" =~ ^[a-z0-9-]+$ ]]; then
    err "Service name must contain only lowercase letters, numbers, and hyphens"
  fi

  # Check if name starts/ends with hyphen
  if [[ "$name" =~ ^- ]] || [[ "$name" =~ -$ ]]; then
    err "Service name cannot start or end with a hyphen"
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

  local service_name="$1"
  local dest_dir="${ROOT_DIR}/applications/services/${service_name}"

  # Validate service name
  validate_service_name "$service_name"

  # Check if template exists
  if [[ ! -d "$TEMPLATE_DIR" ]]; then
    err "Template directory not found: $TEMPLATE_DIR"
  fi

  # Check if destination already exists
  if [[ -e "$dest_dir" ]]; then
    err "Destination already exists: $dest_dir"
  fi

  # Create service directory
  step "Creating service directory: $dest_dir"
  mkdir -p "$(dirname "$dest_dir")"

  # Copy template
  step "Copying template files"
  cp -R "$TEMPLATE_DIR" "$dest_dir"

  # Update service name in files
  step "Updating service name in manifests"

  # Update deployment.yaml
  if [[ -f "${dest_dir}/deployment.yaml" ]]; then
    sed -i.bak "s/name: app/name: ${service_name}/g" "${dest_dir}/deployment.yaml"
    rm "${dest_dir}/deployment.yaml.bak"
  fi

  # Update service.yaml
  if [[ -f "${dest_dir}/service.yaml" ]]; then
    sed -i.bak "s/name: app/name: ${service_name}/g" "${dest_dir}/service.yaml"
    rm "${dest_dir}/service.yaml.bak"
  fi

  # Success message
  info "✓ Successfully created service: $service_name"
  echo ""

  # Next steps
  cat <<EOF
${COLOR_BLUE}Next steps:${COLOR_RESET}

  1. Update the container image in deployment.yaml:
     ${dest_dir}/deployment.yaml

  2. Update the domain/ingress in the service manifests:
     ${dest_dir}/service.yaml

  3. Customize resource limits and replicas as needed

  4. Deploy the service:
     kubectl apply -k ${dest_dir}

  5. Verify deployment:
     kubectl get pods -l app=${service_name}
     kubectl get svc -l app=${service_name}

EOF
}

# Handle --help flag
if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage 0
fi

main "$@"
