#!/usr/bin/env bash
set -euo pipefail

# Mantl Platform Setup Wizard
# Interactive configuration and deployment

# Colors. BLUE and MAGENTA round out the palette but are not used yet.
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
# shellcheck disable=SC2034
BLUE='\033[0;34m'
CYAN='\033[0;36m'
# shellcheck disable=SC2034
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

usage() {
    cat <<'EOF'
Usage: setup-wizard.sh [--help]

Interactive wizard that configures and deploys the Mantl platform. Run with no
arguments to start the guided setup; it prompts for environment, cloud provider,
profile, cluster name, and components.

Options:
  -h, --help    Show this help and exit.
EOF
}

case "${1:-}" in
    -h|--help)
        usage
        exit 0
        ;;
esac

# Configuration
ENVIRONMENT=""
CLOUD_PROVIDER=""
PROFILE=""
CLUSTER_NAME=""
ENABLE_OBSERVABILITY=true
ENABLE_SECURITY=true
ENABLE_EXAMPLES=true
CUSTOM_DOMAIN=""

show_banner() {
    clear
    echo -e "${CYAN}"
    cat << "EOF"
    __  ___            __  __
   /  |/  /___ _____  / /_/ /
  / /|_/ / __ `/ __ \/ __/ /
 / /  / / /_/ / / / / /_/ /
/_/  /_/\__,_/_/ /_/\__/_/

Platform Setup Wizard
EOF
    echo -e "${NC}"
    echo ""
}

prompt_choice() {
    local prompt="$1"
    shift
    local options=("$@")
    local choice

    echo -e "${YELLOW}$prompt${NC}"
    for i in "${!options[@]}"; do
        echo "  $((i+1)). ${options[$i]}"
    done
    echo ""

    while true; do
        read -r -p "Enter choice [1-${#options[@]}]: " choice
        if [[ "$choice" =~ ^[0-9]+$ ]] && [ "$choice" -ge 1 ] && [ "$choice" -le "${#options[@]}" ]; then
            echo "${options[$((choice-1))]}"
            return 0
        else
            echo -e "${RED}Invalid choice. Please try again.${NC}"
        fi
    done
}

prompt_yes_no() {
    local prompt="$1"
    local default="${2:-y}"

    while true; do
        if [ "$default" = "y" ]; then
            read -p "$prompt [Y/n]: " -n 1 -r
        else
            read -p "$prompt [y/N]: " -n 1 -r
        fi
        echo

        if [[ -z "$REPLY" ]]; then
            [[ "$default" = "y" ]] && return 0 || return 1
        elif [[ $REPLY =~ ^[Yy]$ ]]; then
            return 0
        elif [[ $REPLY =~ ^[Nn]$ ]]; then
            return 1
        else
            echo -e "${RED}Please answer y or n.${NC}"
        fi
    done
}

prompt_text() {
    local prompt="$1"
    local default="$2"
    local value

    if [ -n "$default" ]; then
        read -r -p "$prompt [$default]: " value
        echo "${value:-$default}"
    else
        while true; do
            read -r -p "$prompt: " value
            if [ -n "$value" ]; then
                echo "$value"
                return 0
            else
                echo -e "${RED}This field is required.${NC}"
            fi
        done
    fi
}

step_environment() {
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN} Step 1: Choose Environment${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    echo "Select the target environment:"
    echo ""

    ENVIRONMENT=$(prompt_choice "Which environment?" \
        "dev - Local development (kind cluster)" \
        "staging - Pre-production testing" \
        "production - Production workloads")

    ENVIRONMENT=$(echo "$ENVIRONMENT" | cut -d' ' -f1)

    echo ""
    echo -e "${GREEN}✓${NC} Environment: ${CYAN}$ENVIRONMENT${NC}"
    echo ""
    sleep 1
}

step_cloud() {
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN} Step 2: Choose Cloud Provider${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""

    if [ "$ENVIRONMENT" = "dev" ]; then
        echo "For dev environment, we recommend local kind cluster."
        echo ""
        if prompt_yes_no "Use local kind cluster?"; then
            CLOUD_PROVIDER="local"
        else
            CLOUD_PROVIDER=$(prompt_choice "Select cloud provider:" \
                "aws - Amazon Web Services (EKS)" \
                "gcp - Google Cloud Platform (GKE)" \
                "azure - Microsoft Azure (AKS)" \
                "do - DigitalOcean (DOKS)" \
                "linode - Linode (LKE)")
            CLOUD_PROVIDER=$(echo "$CLOUD_PROVIDER" | cut -d' ' -f1)
        fi
    else
        CLOUD_PROVIDER=$(prompt_choice "Select cloud provider:" \
            "aws - Amazon Web Services (EKS)" \
            "gcp - Google Cloud Platform (GKE)" \
            "azure - Microsoft Azure (AKS)" \
            "do - DigitalOcean (DOKS)" \
            "linode - Linode (LKE)" \
            "oci - Oracle Cloud (OKE)" \
            "ibm - IBM Cloud (IKS)")
        CLOUD_PROVIDER=$(echo "$CLOUD_PROVIDER" | cut -d' ' -f1)
    fi

    echo ""
    echo -e "${GREEN}✓${NC} Cloud Provider: ${CYAN}$CLOUD_PROVIDER${NC}"
    echo ""
    sleep 1
}

step_profile() {
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN} Step 3: Choose Resource Profile${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    echo "Resource profiles determine component sizing:"
    echo ""

    PROFILE=$(prompt_choice "Select profile:" \
        "small - Minimal (2 CPU, 4GB RAM, dev/testing)" \
        "medium - Standard (4 CPU, 8GB RAM, staging)" \
        "full - Production (8+ CPU, 16GB RAM, all features)")

    PROFILE=$(echo "$PROFILE" | cut -d' ' -f1)

    echo ""
    echo -e "${GREEN}✓${NC} Profile: ${CYAN}$PROFILE${NC}"
    echo ""
    sleep 1
}

step_cluster_name() {
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN} Step 4: Cluster Configuration${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""

    local default_name="mantl-${ENVIRONMENT}"
    CLUSTER_NAME=$(prompt_text "Enter cluster name" "$default_name")

    if [ "$CLOUD_PROVIDER" != "local" ]; then
        echo ""
        echo -e "${YELLOW}Note:${NC} For cloud deployments, you'll need to:"
        echo "  1. Configure cloud credentials"
        echo "  2. Run Terraform to create the cluster"
        echo "  3. Configure kubectl context"
        echo ""
        echo "Terraform blueprint: ${CYAN}terraform/blueprints/${CLOUD_PROVIDER}-*${NC}"
    fi

    echo ""
    echo -e "${GREEN}✓${NC} Cluster Name: ${CYAN}$CLUSTER_NAME${NC}"
    echo ""
    sleep 1
}

step_components() {
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN} Step 5: Platform Components${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""

    echo "Recommended components will be installed:"
    echo "  • ArgoCD (GitOps)"
    echo "  • cert-manager (TLS automation)"
    echo "  • Kyverno (Policy enforcement)"
    echo ""

    if prompt_yes_no "Enable full observability stack? (Prometheus, Grafana, Tempo, Loki)"; then
        ENABLE_OBSERVABILITY=true
    else
        ENABLE_OBSERVABILITY=false
    fi

    echo ""

    if prompt_yes_no "Enable security components? (Falco, External Secrets)"; then
        ENABLE_SECURITY=true
    else
        ENABLE_SECURITY=false
    fi

    echo ""

    if prompt_yes_no "Deploy example applications?"; then
        ENABLE_EXAMPLES=true
    else
        ENABLE_EXAMPLES=false
    fi

    echo ""
    echo -e "${GREEN}✓${NC} Components configured"
    echo ""
    sleep 1
}

step_domain() {
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN} Step 6: Domain Configuration (Optional)${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""

    if prompt_yes_no "Configure custom domain for ingress?" "n"; then
        CUSTOM_DOMAIN=$(prompt_text "Enter domain (e.g., example.com)" "")
        echo ""
        echo -e "${GREEN}✓${NC} Domain: ${CYAN}$CUSTOM_DOMAIN${NC}"
    else
        echo -e "${YELLOW}⚠${NC} Using default (localhost / cluster IPs)"
    fi

    echo ""
    sleep 1
}

show_summary() {
    clear
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo -e "${CYAN} Configuration Summary${NC}"
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
    echo -e "${YELLOW}Environment:${NC}      $ENVIRONMENT"
    echo -e "${YELLOW}Cloud Provider:${NC}   $CLOUD_PROVIDER"
    echo -e "${YELLOW}Profile:${NC}          $PROFILE"
    echo -e "${YELLOW}Cluster Name:${NC}     $CLUSTER_NAME"
    echo -e "${YELLOW}Observability:${NC}    $([ "$ENABLE_OBSERVABILITY" = true ] && echo "Enabled" || echo "Disabled")"
    echo -e "${YELLOW}Security:${NC}         $([ "$ENABLE_SECURITY" = true ] && echo "Enabled" || echo "Disabled")"
    echo -e "${YELLOW}Examples:${NC}         $([ "$ENABLE_EXAMPLES" = true ] && echo "Enabled" || echo "Disabled")"
    if [ -n "$CUSTOM_DOMAIN" ]; then
        echo -e "${YELLOW}Custom Domain:${NC}    $CUSTOM_DOMAIN"
    fi
    echo ""
    echo -e "${CYAN}═══════════════════════════════════════════${NC}"
    echo ""
}

generate_config() {
    cat > /tmp/mantl-config.env <<EOF
# Mantl Platform Configuration
# Generated: $(date)

MANTL_ENVIRONMENT=$ENVIRONMENT
MANTL_CLOUD=$CLOUD_PROVIDER
MANTL_PROFILE=$PROFILE
MANTL_CLUSTER_NAME=$CLUSTER_NAME
MANTL_OBSERVABILITY=$ENABLE_OBSERVABILITY
MANTL_SECURITY=$ENABLE_SECURITY
MANTL_EXAMPLES=$ENABLE_EXAMPLES
MANTL_DOMAIN=${CUSTOM_DOMAIN:-}
EOF

    echo -e "${GREEN}✓${NC} Configuration saved to: ${CYAN}/tmp/mantl-config.env${NC}"
}

install_platform() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo -e "${GREEN} Installing Platform${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════${NC}"
    echo ""

    local bootstrap_args=(
        --environment "$ENVIRONMENT"
        --cloud "$CLOUD_PROVIDER"
        --profile "$PROFILE"
        --cluster-name "$CLUSTER_NAME"
    )

    if [ "$ENABLE_EXAMPLES" = false ]; then
        bootstrap_args+=(--skip-examples)
    fi

    echo -e "${CYAN}Running:${NC} ./scripts/bootstrap-platform.sh ${bootstrap_args[*]}"
    echo ""

    ./scripts/bootstrap-platform.sh "${bootstrap_args[@]}"
}

main() {
    show_banner

    echo "Welcome to the Mantl Platform Setup Wizard!"
    echo ""
    echo "This wizard will guide you through configuring and"
    echo "deploying your Kubernetes platform."
    echo ""
    read -r -p "Press Enter to continue..."

    step_environment
    step_cloud
    step_profile
    step_cluster_name
    step_components
    step_domain

    show_summary

    if prompt_yes_no "Proceed with installation?"; then
        generate_config
        install_platform
    else
        echo ""
        echo -e "${YELLOW}Installation cancelled.${NC}"
        echo "Configuration saved to /tmp/mantl-config.env"
        echo ""
        echo "To install later, run:"
        echo -e "  ${CYAN}make install-$ENVIRONMENT${NC}"
        echo ""
    fi
}

# Run wizard
main
