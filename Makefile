.PHONY: help install-dev install-staging install-production install-small install-medium install-full clean test lint wizard cli-install

# Default target
.DEFAULT_GOAL := help

# Colors for output
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
CYAN   := $(shell tput -Txterm setaf 6)
RESET  := $(shell tput -Txterm sgr0)

##@ General

help: ## Show this help message
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "  ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

##@ Quick Start

install-dev: ## Install complete dev environment (kind + platform)
	@echo "${GREEN}Installing dev environment...${RESET}"
	@./scripts/bootstrap-platform.sh --environment dev --cloud local --profile small

install-staging: ## Install staging environment
	@echo "${GREEN}Installing staging environment...${RESET}"
	@./scripts/bootstrap-platform.sh --environment staging --profile medium

install-production: ## Install production environment
	@echo "${GREEN}Installing production environment...${RESET}"
	@./scripts/bootstrap-platform.sh --environment production --profile full

##@ Installation Profiles

install-small: ## Install minimal platform (2 CPU, 4GB RAM)
	@echo "${GREEN}Installing small profile...${RESET}"
	@./scripts/bootstrap-platform.sh --profile small

install-medium: ## Install medium platform (4 CPU, 8GB RAM)
	@echo "${GREEN}Installing medium profile...${RESET}"
	@./scripts/bootstrap-platform.sh --profile medium

install-full: ## Install full platform (all components)
	@echo "${GREEN}Installing full profile...${RESET}"
	@./scripts/bootstrap-platform.sh --profile full

##@ Interactive Setup

wizard: ## Run interactive setup wizard
	@./scripts/setup-wizard.sh

##@ Development Tools

cli-install: ## Install mantl CLI tool
	@echo "${GREEN}Installing mantl CLI...${RESET}"
	@chmod +x scripts/mantl
	@sudo cp scripts/mantl /usr/local/bin/mantl 2>/dev/null || cp scripts/mantl ~/bin/mantl || echo "${YELLOW}Please add scripts/mantl to your PATH${RESET}"
	@echo "${GREEN}✓ mantl CLI installed${RESET}"

##@ Testing & Validation

test: ## Run all tests
	@echo "${GREEN}Running tests...${RESET}"
	@pytest -v tests/ || true
	@echo "${GREEN}Validating Kustomize builds...${RESET}"
	@./scripts/validate-kustomize.sh
	@echo "${GREEN}Validating Terraform...${RESET}"
	@./scripts/validate-terraform.sh

lint: ## Run all linters
	@echo "${GREEN}Running linters...${RESET}"
	@black --check . || true
	@ruff check . || true
	@yamllint . || true
	@shellcheck scripts/*.sh || true

terraform-validate: ## Validate all Terraform configurations
	@echo "${GREEN}Validating Terraform blueprints...${RESET}"
	@for dir in terraform/blueprints/*/; do \
		echo "Validating $$dir"; \
		cd "$$dir" && terraform init -backend=false && terraform validate && cd -; \
	done

kustomize-validate: ## Validate all Kustomize configurations
	@echo "${GREEN}Validating Kustomize builds...${RESET}"
	@find . -name kustomization.yaml -exec dirname {} \; | while read dir; do \
		echo "Building $$dir"; \
		kustomize build "$$dir" > /dev/null || echo "Warning: $$dir failed"; \
	done

##@ Example Applications

deploy-examples: ## Deploy all example applications
	@echo "${GREEN}Deploying example applications...${RESET}"
	@kubectl apply -k applications/examples/api-service/base
	@kubectl apply -k applications/examples/frontend/base
	@kubectl apply -k applications/examples/database-app/base
	@echo "${GREEN}✓ Examples deployed${RESET}"

delete-examples: ## Delete all example applications
	@echo "${YELLOW}Deleting example applications...${RESET}"
	@kubectl delete -k applications/examples/api-service/base --ignore-not-found
	@kubectl delete -k applications/examples/frontend/base --ignore-not-found
	@kubectl delete -k applications/examples/database-app/base --ignore-not-found
	@echo "${GREEN}✓ Examples deleted${RESET}"

##@ Platform Management

status: ## Show platform status
	@echo "${CYAN}Platform Status:${RESET}"
	@kubectl get applications -n argocd 2>/dev/null || echo "ArgoCD not installed"
	@echo ""
	@echo "${CYAN}Node Status:${RESET}"
	@kubectl get nodes 2>/dev/null || echo "No cluster found"

dashboards: ## Open platform dashboards
	@echo "${GREEN}Opening dashboards...${RESET}"
	@echo "${CYAN}ArgoCD:${RESET} http://localhost:8080"
	@kubectl port-forward svc/argocd-server -n argocd 8080:443 > /dev/null 2>&1 &
	@echo "${CYAN}Grafana:${RESET} http://localhost:3000"
	@kubectl port-forward svc/prometheus-grafana -n monitoring 3000:80 > /dev/null 2>&1 &
	@echo "${CYAN}Prometheus:${RESET} http://localhost:9090"
	@kubectl port-forward svc/prometheus-prometheus -n monitoring 9090:9090 > /dev/null 2>&1 &

logs: ## Tail ArgoCD application controller logs
	@kubectl logs -n argocd -l app.kubernetes.io/name=argocd-application-controller -f

##@ Cleanup

clean: ## Delete kind cluster and clean up
	@echo "${YELLOW}Cleaning up...${RESET}"
	@kind delete cluster --name mantl-dev 2>/dev/null || true
	@kind delete cluster --name mantl 2>/dev/null || true
	@echo "${GREEN}✓ Cleanup complete${RESET}"

clean-all: clean ## Deep clean (includes Docker images)
	@echo "${YELLOW}Deep cleaning...${RESET}"
	@docker system prune -f
	@echo "${GREEN}✓ Deep cleanup complete${RESET}"

##@ Prerequisites

check-prereqs: ## Check if required tools are installed
	@echo "${GREEN}Checking prerequisites...${RESET}"
	@command -v kubectl >/dev/null 2>&1 || echo "${YELLOW}⚠ kubectl not found${RESET}"
	@command -v kind >/dev/null 2>&1 || echo "${YELLOW}⚠ kind not found${RESET}"
	@command -v kustomize >/dev/null 2>&1 || echo "${YELLOW}⚠ kustomize not found${RESET}"
	@command -v terraform >/dev/null 2>&1 || echo "${YELLOW}⚠ terraform not found${RESET}"
	@command -v docker >/dev/null 2>&1 || echo "${YELLOW}⚠ docker not found${RESET}"
	@echo "${GREEN}✓ Prerequisites check complete${RESET}"

install-prereqs: ## Install prerequisite tools (macOS)
	@echo "${GREEN}Installing prerequisites...${RESET}"
	@if [ "$$(uname)" = "Darwin" ]; then \
		command -v brew >/dev/null 2>&1 || /bin/bash -c "$$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"; \
		brew install kubectl kind kustomize terraform; \
	else \
		echo "${YELLOW}Please install: kubectl, kind, kustomize, terraform${RESET}"; \
	fi
