#!/usr/bin/env bash
set -euo pipefail

# Bootstrap Mantl Platform
# Usage: ./bootstrap-platform.sh [OPTIONS]

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="dev"
CLOUD_PROVIDER="local"
PROFILE="small"
CLUSTER_NAME="mantl-dev"
SKIP_EXAMPLES=false
VERBOSE=false

# Functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
}

show_help() {
    cat << EOF
Mantl Platform Bootstrap Script

Usage: $0 [OPTIONS]

Options:
    -e, --environment ENV       Environment (dev|staging|production) [default: dev]
    -c, --cloud PROVIDER        Cloud provider (local|aws|gcp|azure|do|linode) [default: local]
    -p, --profile PROFILE       Resource profile (small|medium|full) [default: small]
    -n, --cluster-name NAME     Cluster name [default: mantl-dev]
    --skip-examples             Skip deploying example applications
    -v, --verbose               Verbose output
    -h, --help                  Show this help message

Profiles:
    small   - 2 CPU, 4GB RAM (local development)
    medium  - 4 CPU, 8GB RAM (staging)
    full    - All components, production-ready

Examples:
    # Install dev environment locally
    $0 --environment dev --cloud local --profile small

    # Install staging on AWS
    $0 --environment staging --cloud aws --profile medium

    # Install production with all components
    $0 --environment production --profile full

EOF
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    local missing_tools=()

    for tool in kubectl kind kustomize docker; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done

    if [ ${#missing_tools[@]} -ne 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        log_info "Install them with: make install-prereqs (macOS) or manually"
        exit 1
    fi

    log_success "All prerequisites installed"
}

create_kind_cluster() {
    log_info "Creating kind cluster: $CLUSTER_NAME..."

    # Determine node configuration based on profile
    local worker_nodes=1
    local cpu_limit="2"
    local memory_limit="4g"

    case $PROFILE in
        medium)
            worker_nodes=2
            cpu_limit="4"
            memory_limit="8g"
            ;;
        full)
            worker_nodes=3
            cpu_limit="8"
            memory_limit="16g"
            ;;
    esac

    # Check if cluster already exists
    if kind get clusters 2>/dev/null | grep -q "^${CLUSTER_NAME}$"; then
        log_warn "Cluster $CLUSTER_NAME already exists"
        read -p "Delete and recreate? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            kind delete cluster --name "$CLUSTER_NAME"
        else
            log_info "Using existing cluster"
            return 0
        fi
    fi

    # Create kind config
    cat > /tmp/kind-config.yaml <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: ${CLUSTER_NAME}
nodes:
  - role: control-plane
    kubeadmConfigPatches:
      - |
        kind: InitConfiguration
        nodeRegistration:
          kubeletExtraArgs:
            node-labels: "ingress-ready=true"
    extraPortMappings:
      - containerPort: 80
        hostPort: 80
        protocol: TCP
      - containerPort: 443
        hostPort: 443
        protocol: TCP
EOF

    # Add worker nodes
    for ((i=1; i<=worker_nodes; i++)); do
        cat >> /tmp/kind-config.yaml <<EOF
  - role: worker
EOF
    done

    kind create cluster --config /tmp/kind-config.yaml --wait 5m

    log_success "Kind cluster created"
}

install_argocd() {
    log_info "Installing ArgoCD..."

    kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -

    kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

    log_info "Waiting for ArgoCD to be ready..."
    kubectl wait --for=condition=available --timeout=300s \
        deployment/argocd-server -n argocd

    # Get admin password
    local admin_password
    admin_password=$(kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d)

    log_success "ArgoCD installed"
    log_info "ArgoCD Admin Password: ${CYAN}${admin_password}${NC}"
    log_info "Access ArgoCD: kubectl port-forward svc/argocd-server -n argocd 8080:443"
}

deploy_platform() {
    log_info "Deploying platform for environment: $ENVIRONMENT..."

    # Deploy app-of-apps
    if [ -f "clusters/${ENVIRONMENT}/app-of-apps.yaml" ]; then
        kubectl apply -f "clusters/${ENVIRONMENT}/app-of-apps.yaml"
        log_success "Platform app-of-apps deployed"
    else
        # Fallback to kustomization
        kubectl apply -k "clusters/${ENVIRONMENT}"
        log_success "Platform components deployed"
    fi

    log_info "Waiting for platform to sync..."
    sleep 10

    # Wait for critical apps
    log_info "Waiting for cert-manager..."
    kubectl wait --for=condition=available --timeout=300s \
        deployment/cert-manager -n cert-manager 2>/dev/null || log_warn "cert-manager not ready yet"

    log_success "Platform deployed"
}

deploy_examples() {
    if [ "$SKIP_EXAMPLES" = true ]; then
        log_info "Skipping example applications"
        return 0
    fi

    log_info "Deploying example applications..."

    kubectl apply -k applications/examples/hello 2>/dev/null || log_warn "hello app failed"

    log_success "Example applications deployed"
}

show_access_info() {
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
    echo -e "${GREEN}  Mantl Platform Installed Successfully! 🎉${NC}"
    echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
    echo ""
    echo -e "${CYAN}Cluster:${NC} $CLUSTER_NAME"
    echo -e "${CYAN}Environment:${NC} $ENVIRONMENT"
    echo -e "${CYAN}Profile:${NC} $PROFILE"
    echo ""
    echo -e "${YELLOW}Next Steps:${NC}"
    echo ""
    echo "  1. Access ArgoCD UI:"
    echo -e "     ${CYAN}kubectl port-forward svc/argocd-server -n argocd 8080:443${NC}"
    echo -e "     Open: ${BLUE}https://localhost:8080${NC}"
    echo ""
    echo "  2. Access Grafana:"
    echo -e "     ${CYAN}kubectl port-forward svc/prometheus-grafana -n monitoring 3000:80${NC}"
    echo -e "     Open: ${BLUE}http://localhost:3000${NC}"
    echo ""
    echo "  3. Check platform status:"
    echo -e "     ${CYAN}make status${NC}"
    echo -e "     ${CYAN}mantl status${NC} (if CLI installed)"
    echo ""
    echo "  4. Deploy example apps:"
    echo -e "     ${CYAN}make deploy-examples${NC}"
    echo ""
    echo -e "${YELLOW}Useful Commands:${NC}"
    echo -e "  ${CYAN}make dashboards${NC}     - Open all dashboards"
    echo -e "  ${CYAN}make logs${NC}            - Tail ArgoCD logs"
    echo -e "  ${CYAN}mantl dashboard${NC}      - Quick dashboard access"
    echo -e "  ${CYAN}mantl deploy <app>${NC}   - Deploy an application"
    echo ""
    echo -e "${GREEN}═══════════════════════════════════════════════════════${NC}"
    echo ""
}

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -c|--cloud)
            CLOUD_PROVIDER="$2"
            shift 2
            ;;
        -p|--profile)
            PROFILE="$2"
            shift 2
            ;;
        -n|--cluster-name)
            CLUSTER_NAME="$2"
            shift 2
            ;;
        --skip-examples)
            SKIP_EXAMPLES=true
            shift
            ;;
        -v|--verbose)
            VERBOSE=true
            set -x
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_help
            exit 1
            ;;
    esac
done

# Main execution
main() {
    echo -e "${CYAN}"
    cat << "EOF"
    __  ___            __  __
   /  |/  /___ _____  / /_/ /
  / /|_/ / __ `/ __ \/ __/ /
 / /  / / /_/ / / / / /_/ /
/_/  /_/\__,_/_/ /_/\__/_/

Platform Bootstrap
EOF
    echo -e "${NC}"

    log_info "Environment: $ENVIRONMENT"
    log_info "Cloud: $CLOUD_PROVIDER"
    log_info "Profile: $PROFILE"
    echo ""

    check_prerequisites

    if [ "$CLOUD_PROVIDER" = "local" ]; then
        create_kind_cluster
    else
        log_warn "Cloud provider '$CLOUD_PROVIDER' detected - ensure cluster is created via Terraform"
        log_info "Run: terraform -chdir=terraform/blueprints/${CLOUD_PROVIDER}-* apply"
        read -p "Press Enter when cluster is ready..."
    fi

    install_argocd
    deploy_platform
    deploy_examples

    show_access_info
}

# Run main function
main
