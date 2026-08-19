# Mantl Development Guide

This guide provides detailed information for developers working on the Mantl 2026 platform.

## Quick Start

```bash
# Install the platform locally
make install-dev

# Or use the interactive wizard
make wizard

# Or use the CLI
./bin/mantl init
```

## Development Environment

### Requirements

**Core Tools:**
- **Docker** or **Podman** - Container runtime
- **kind** - Local Kubernetes clusters
- **kubectl** 1.28+ - Kubernetes CLI
- **kustomize** 5.0+ - Kubernetes configuration management
- **helm** 3.0+ - Package manager for Kubernetes

**Infrastructure:**
- **Terraform** 1.5+ (or **OpenTofu**) - Infrastructure as Code
- **GitHub CLI (gh)** - For PRs and releases

**Development:**
- **Python 3.11+** - For tooling and scripts
- **Node.js 18+** - For some development tools
- **Git** - Version control

**Optional but Recommended:**
- **conftest** - Policy testing
- **kyverno-cli** - Policy validation
- **kubeconform** - Kubernetes manifest validation
- **cosign** - Container signing
- **syft** - SBOM generation
- **trivy** - Security scanning

### Installation (macOS)

```bash
# Install all prerequisites
make install-prereqs

# Or manually with Homebrew
brew install kubectl kind kustomize terraform gh helm docker
```

### Installation (Linux)

```bash
# Ubuntu/Debian
sudo apt-get update
sudo apt-get install -y docker.io kubectl

# Install kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Install kubectl, kustomize, helm, terraform via package manager or binaries
```

## Setting Up Your Development Environment

### 1. Clone and Setup

```bash
# Fork the repository on GitHub first
git clone https://github.com/YOUR_USERNAME/mantl.git
cd mantl

# Add upstream remote
git remote add upstream https://github.com/thomasvincent/mantl.git

# Create a feature branch
git checkout -b feat/your-feature-name
```

### 2. Install Python Dependencies

```bash
# Create virtual environment (recommended)
python3 -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt
pip install -r requirements-test.txt

# Install development tools
pip install black ruff pytest pytest-cov yamllint
```

### 3. Install Pre-commit Hooks

```bash
# Install pre-commit
pip install pre-commit

# Install git hooks
pre-commit install

# Test hooks (optional)
pre-commit run --all-files
```

The pre-commit hooks automatically check your code before each commit for:
- Python formatting (Black)
- Python linting (Ruff)
- YAML linting (yamllint)
- Terraform formatting
- Shell script linting (shellcheck)
- Commit message format (Conventional Commits)

## Development Workflow

### 1. Make Your Changes

Edit code, manifests, or documentation as needed.

### 2. Format and Lint

```bash
# Format code automatically
make format

# Or manually
black .
ruff check --fix .
terraform fmt -recursive

# Lint (check without fixing)
make lint
```

### 3. Run Tests

```bash
# Run all tests
make test

# Run specific tests
pytest tests/unit -v
pytest tests/integration -v

# With coverage
pytest --cov=. --cov-report=html
```

### 4. Validate Manifests

```bash
# Validate Kustomize builds
make kustomize-validate

# Validate Terraform
make terraform-validate

# Validate Kyverno policies
kyverno validate policies/kyverno/*.yaml

# Test policies against manifests
kyverno apply policies/kyverno -r platform/ --audit-warn
```

### 5. Test Locally with kind

```bash
# Create local cluster and deploy platform
make install-dev

# Check status
make status
kubectl get applications -n argocd

# View dashboards
make dashboards

# Clean up
make clean
```

### 6. Commit Your Changes

```bash
# Stage changes
git add .

# Commit with conventional commit message
git commit -m "feat: add new feature description"

# Conventional commit types:
# feat:     New feature
# fix:      Bug fix
# docs:     Documentation changes
# chore:    Maintenance tasks
# refactor: Code refactoring
# perf:     Performance improvements
# test:     Adding tests
# build:    Build system changes
# ci:       CI configuration changes
```

### 7. Push and Create PR

```bash
# Push to your fork
git push origin feat/your-feature-name

# Create pull request using GitHub CLI
gh pr create --fill

# Or manually on GitHub
```

## Code Quality Standards

### Python Code

- **Follow PEP 8** - Python style guide
- **Use Black** for formatting (120 character line length)
- **Use Ruff** for fast linting
- **Type hints** are encouraged but not required
- **Docstrings** for public functions and classes

Example:
```python
def deploy_application(name: str, namespace: str = "default") -> bool:
    """Deploy an application to Kubernetes.

    Args:
        name: Name of the application
        namespace: Target namespace (default: "default")

    Returns:
        True if deployment succeeded, False otherwise
    """
    # Implementation
    pass
```

### Kubernetes Manifests

- **Use Kustomize** for configuration management
- **Follow best practices**:
  - Include resource limits and requests
  - Use security contexts (non-root, read-only filesystem)
  - Include health checks (liveness + readiness)
  - Proper labels: `app.kubernetes.io/*`
  - Use namespaces (never default for platform components)

Example:
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
  labels:
    app.kubernetes.io/name: example
    app.kubernetes.io/component: backend
    app.kubernetes.io/part-of: mantl
    app.kubernetes.io/managed-by: argocd
spec:
  replicas: 2
  selector:
    matchLabels:
      app.kubernetes.io/name: example
  template:
    metadata:
      labels:
        app.kubernetes.io/name: example
    spec:
      serviceAccountName: example
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      containers:
        - name: app
          image: ghcr.io/example/app:1.0.0
          ports:
            - containerPort: 8080
              name: http
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: [ALL]
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
          readinessProbe:
            httpGet:
              path: /ready
              port: http
```

### Terraform Code

- **Use terraform fmt** - Format all `.tf` files
- **Use modules** - Create reusable modules for common patterns
- **Version constraints** - Pin provider versions
- **Variables** - Use descriptive names and include descriptions
- **Outputs** - Document all outputs

Example:
```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

variable "cluster_name" {
  description = "Name of the EKS cluster"
  type        = string
}

variable "kubernetes_version" {
  description = "Kubernetes version"
  type        = string
  default     = "1.31"
}

resource "aws_eks_cluster" "main" {
  name     = var.cluster_name
  role_arn = aws_iam_role.cluster.arn
  version  = var.kubernetes_version

  # Configuration
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = aws_eks_cluster.main.endpoint
}
```

### Shell Scripts

- **Use shellcheck** - Lint shell scripts
- **Use set -euo pipefail** - Fail fast
- **Use functions** - Organize code into functions
- **Use quotes** - Quote variables to prevent word splitting
- **Document** - Add comments for complex logic

Example:
```bash
#!/usr/bin/env bash
set -euo pipefail

# Colors
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly NC='\033[0m'

log_info() {
    echo -e "${GREEN}ℹ${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

main() {
    local cluster_name="${1:-mantl-dev}"

    log_info "Creating cluster: $cluster_name"

    if kind create cluster --name "$cluster_name"; then
        log_info "Cluster created successfully"
    else
        log_error "Failed to create cluster"
        return 1
    fi
}

main "$@"
```

## Testing

### Types of Tests

1. **Unit Tests** - Test individual functions
2. **Integration Tests** - Test component interactions
3. **E2E Tests** - Test complete workflows
4. **Policy Tests** - Test Kyverno policies
5. **Validation Tests** - Test manifest validity

### Running Tests

```bash
# All tests
make test

# Unit tests only
pytest tests/unit -v

# Integration tests
pytest tests/integration -v

# Policy tests
kyverno test policies/kyverno

# E2E tests (requires cluster)
./scripts/test-platform.sh
```

### Writing Tests

```python
# tests/test_deploy.py
import pytest
from mantl.deploy import deploy_application

def test_deploy_application():
    """Test application deployment."""
    result = deploy_application("test-app", "default")
    assert result is True

def test_deploy_invalid_namespace():
    """Test deployment with invalid namespace."""
    with pytest.raises(ValueError):
        deploy_application("test-app", "")
```

## Documentation

### Documentation Structure

- `README.md` - Main documentation
- `DEVELOPMENT.md` - This file
- `CONTRIBUTING.md` - Contribution guidelines
- `docs/` - Detailed documentation
  - `ara/` - Architecture Decision Records
  - `quickstart-platform.md` - Quick start guide
  - `security.md` - Security documentation
  - `*.md` - Feature-specific docs

### Writing Documentation

- **Use Markdown** for new documentation
- **Include examples** - Code snippets and commands
- **Keep it current** - Update when features change
- **Link references** - Link to related docs
- **Use clear language** - Write for your audience

### Building Documentation

```bash
# View locally
open docs/index.md  # Or use your favorite Markdown viewer

# Check for broken links
find docs -name "*.md" -exec markdown-link-check {} \;
```

## Debugging

### Common Issues

**1. Kind cluster won't start**
```bash
# Check Docker is running
docker ps

# Delete and recreate
kind delete cluster --name mantl-dev
make install-dev
```

**2. ArgoCD applications stuck syncing**
```bash
# Check application status
kubectl describe application <app-name> -n argocd

# Force sync
kubectl patch application <app-name> -n argocd \
  --type merge -p '{"operation":{"sync":{}}}'

# View logs
kubectl logs -n argocd -l app.kubernetes.io/name=argocd-application-controller
```

**3. Tests failing locally**
```bash
# Ensure virtual environment is activated
source venv/bin/activate

# Reinstall dependencies
pip install -r requirements.txt -r requirements-test.txt

# Clear pytest cache
rm -rf .pytest_cache
pytest --cache-clear
```

### Debugging Tools

```bash
# CLI tool for debugging
./bin/mantl health        # Check component health
./bin/mantl status        # Platform status
./bin/mantl logs <app>    # View logs

# Makefile targets
make status                   # Platform status
make logs                     # ArgoCD logs
make dashboards               # Open dashboards

# kubectl commands
kubectl get all -A            # All resources
kubectl describe pod <pod>    # Pod details
kubectl logs <pod> -f         # Tail logs
kubectl exec -it <pod> -- sh  # Shell into pod
```

## Release Process

Releases are automated via release-please:

1. **Make changes** and commit with conventional commit messages
2. **Merge to main** - release-please opens a release PR
3. **Review release PR** - Check changelog and version bump
4. **Merge release PR** - Creates tag and GitHub release
5. **Automated workflow** - Signs artifacts and publishes

Manual tagging (if needed):
```bash
# Tag version
make tag VERSION=v1.2.3

# Create GitHub release
make release VERSION=v1.2.3
```

## CI/CD Pipeline

Our GitHub Actions CI pipeline runs on every PR:

1. **Linting** - Black, Ruff, yamllint, shellcheck
2. **Testing** - pytest unit tests
3. **Validation** - Terraform, Kustomize, Kyverno
4. **Security** - Trivy scanning, SBOM generation
5. **E2E** - kind cluster deployment test

Required checks (must pass):
- All linters
- All tests
- Terraform validation
- Kustomize builds
- Security scans

See `.github/workflows/ci.yml` for details.

## Getting Help

- **Documentation**: Check `docs/` directory
- **Issues**: Open an issue on GitHub
- **CLI Help**: `./bin/mantl help`
- **Makefile**: `make help`

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [Kyverno Documentation](https://kyverno.io/docs/)
- [Terraform Documentation](https://www.terraform.io/docs/)
- [Kind Documentation](https://kind.sigs.k8s.io/)
- [CNCF Landscape](https://landscape.cncf.io/)

---

**Happy coding!** 🚀
