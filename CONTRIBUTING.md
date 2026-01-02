# Contributing to Mantl

Thank you for your interest in contributing to Mantl! This guide will help you get started with the contribution process.

## Table of Contents

- [Quick Start for Contributors](#quick-start-for-contributors)
- [Development Environment Setup](#development-environment-setup)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Testing and Validation](#testing-and-validation)
- [Release Process](#release-process)
- [Code of Conduct](#code-of-conduct)

## Quick Start for Contributors

**New contributors**: Follow these steps to make your first contribution:

1. **Fork the repository** on GitHub
2. **Clone your fork**: `git clone https://github.com/YOUR_USERNAME/mantl.git`
3. **Install pre-commit hooks**: `pip install pre-commit && pre-commit install`
4. **Create a feature branch**: `git checkout -b feat/your-feature-name`
5. **Make your changes** and commit using [Conventional Commits](#commit-message-format)
6. **Run tests locally**: `make test && make lint`
7. **Push to your fork**: `git push origin feat/your-feature-name`
8. **Open a Pull Request** on GitHub

**Important**: All CI checks must pass before your PR can be merged.

## Development Environment Setup

If you prefer an all-in-one toolchain, use the provided dev container (Dockerfile) which includes kubectl, kustomize+helm, Terraform, Python, and make.

```bash
# Build once
docker build -t mantl-dev:local .

# Pytest
docker run --rm -v "$PWD":/workspace -w /workspace mantl-dev:local pytest -v tests

# Validators
docker run --rm -v "$PWD":/workspace -w /workspace mantl-dev:local ./ci/validate-kustomize.sh
docker run --rm -v "$PWD":/workspace -w /workspace mantl-dev:local ./ci/validate-terraform.sh
```

Install pre-commit to match CI hooks:

```bash
pipx install pre-commit  # or: python -m pip install pre-commit --user
pre-commit install
pre-commit run -a
```

### Prerequisites

Install the following tools before contributing:

**Required Tools:**
- **Git** - Version control
- **GitHub CLI (gh)** - For creating PRs and releases
- **Docker** or **Podman** - Container runtime
- **kind** - Local Kubernetes clusters (v0.20+)
- **kubectl** - Kubernetes CLI (v1.28+)
- **Terraform** or **OpenTofu** - Infrastructure as Code (v1.5+)
- **Python 3.11+** - For development tools and scripts

**Recommended Tools:**
- **helm** - Kubernetes package manager (v3.0+)
- **kustomize** - Kubernetes configuration management (v5.0+)
- **conftest** - Policy testing
- **kyverno-cli** - Policy validation
- **kubeconform** - Kubernetes manifest validation
- **cosign** - Container signing
- **syft** - SBOM generation
- **trivy** - Security scanning

### Installation

**macOS (Homebrew):**
```bash
brew install gh docker kind kubectl terraform helm kustomize
pip3 install pre-commit black ruff pytest
```

**Linux (Ubuntu/Debian):**
```bash
# Install Docker
sudo apt-get update
sudo apt-get install -y docker.io

# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Install kind
curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64
chmod +x ./kind
sudo mv ./kind /usr/local/bin/kind

# Install Python tools
pip3 install pre-commit black ruff pytest
```

### Repository Setup

1. **Fork and clone the repository:**
   ```bash
   # Fork on GitHub first, then:
   git clone https://github.com/YOUR_USERNAME/mantl.git
   cd mantl
   ```

2. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/thomasvincent/mantl.git
   ```

3. **Install Python dependencies:**
   ```bash
   pip install -r requirements.txt
   pip install -r requirements-test.txt
   ```

4. **Install pre-commit hooks:**
   ```bash
   pre-commit install
   pre-commit run --all-files  # Test the hooks
   ```

## Development Workflow

### Creating a Feature Branch

Always create a new branch for your changes:

```bash
# Update your local main branch
git checkout main
git pull upstream main

# Create a feature branch
git checkout -b feat/your-feature-name
```

**Branch Naming Conventions:**
- `feat/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `chore/` - Maintenance tasks
- `refactor/` - Code refactoring
- `test/` - Test additions or changes

### Making Changes

1. **Write your code** following the [style guidelines](#code-style-guidelines)
2. **Format your code:**
   ```bash
   make format
   # Or manually:
   black .
   ruff check --fix .
   terraform fmt -recursive
   ```

3. **Test your changes:**
   ```bash
   make test
   # Or run specific tests:
   pytest tests/unit -v
   pytest tests/integration -v
   ```

### Commit Message Format

We use [Conventional Commits](https://www.conventionalcommits.org/) for commit messages:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `chore:` - Maintenance tasks
- `refactor:` - Code refactoring
- `perf:` - Performance improvements
- `test:` - Test additions or changes
- `build:` - Build system changes
- `ci:` - CI configuration changes

**Examples:**
```bash
git commit -m "feat: add SOC2 compliance framework"
git commit -m "fix: correct Kyverno policy validation"
git commit -m "docs: update README with quick start guide"
```

The pre-commit hooks will validate your commit messages automatically.

## Pull Request Process

### Before Opening a PR

1. **Ensure all tests pass:**
   ```bash
   make test
   make lint
   make validate
   ```

2. **Update documentation** if you've changed functionality

3. **Rebase on latest main:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

### Opening a PR

1. **Push your branch:**
   ```bash
   git push origin feat/your-feature-name
   ```

2. **Create PR using GitHub CLI:**
   ```bash
   gh pr create --fill
   ```

   Or manually on GitHub.

3. **PR Description Template:**
   ```markdown
   ## Summary
   Brief description of what this PR does.

   ## Motivation
   Why is this change needed?

   ## Changes
   - List of changes made
   - One bullet per significant change

   ## Testing
   How was this tested?

   ## Checklist
   - [ ] Tests pass locally
   - [ ] Documentation updated
   - [ ] Conventional commit messages used
   - [ ] Pre-commit hooks pass
   ```

### PR Review Process

- Maintainers will review your PR
- Address review feedback by pushing new commits
- Once approved, maintainers will merge your PR
- Your contribution will be included in the next release

## Testing and Validation

### Local Testing

**Run all tests:**
```bash
make test
```

**Run specific test suites:**
```bash
# Unit tests
pytest tests/unit -v

# Integration tests
pytest tests/integration -v

# Policy tests
make test-policies
```

### Validation

**Validate Kubernetes manifests:**
```bash
make validate
# Or manually:
kubeconform -strict -summary -ignore-missing-schemas \
  -k8s-version 1.30 -schema-location default <manifest-file>
```

**Validate Kyverno policies:**
```bash
kyverno apply policies/kyverno -r platform/ --audit-warn
kyverno test policies/kyverno/
```

**Validate Terraform:**
```bash
make terraform-validate
# Or manually:
cd terraform/blueprints/aws-eks
terraform init -backend=false
terraform validate
```

### Local Kind Cluster Testing

Test the full platform locally:

```bash
make install-dev
# Or manually:
./scripts/bootstrap-platform.sh --environment dev --cloud local
```

## Code Style Guidelines

### Python

- Follow [PEP 8](https://pep8.org/)
- Use **Black** for formatting (120 character line length)
- Use **Ruff** for linting
- Add type hints where appropriate
- Include docstrings for public functions

### Kubernetes Manifests

- Use **Kustomize** for configuration management
- Include resource limits and requests
- Use security contexts (non-root, read-only filesystem)
- Include health checks (liveness + readiness)
- Use proper labels: `app.kubernetes.io/*`
- Never use the `default` namespace for platform components

### Terraform

- Run `terraform fmt` on all `.tf` files
- Use modules for reusable patterns
- Pin provider versions
- Add descriptions to all variables
- Document all outputs

### Shell Scripts

- Use `shellcheck` for linting
- Use `set -euo pipefail` for safety
- Quote all variables
- Add comments for complex logic
- Use functions to organize code

## Security Policy

### Kyverno Policy Enforcement

Kyverno policies live in `policies/kyverno/`:

1. **Start in Audit mode** - Monitor without blocking
2. **Review PolicyReports** - Check for violations
3. **Switch to Enforce mode** - Block violations

### Image Signature Verification

**Key-based signing:**
```bash
scripts/set-cosign-key.sh path/to/cosign.pub
```

**Keyless signing (OIDC):**
```bash
scripts/set-keyless-subject.sh \
  "https://github.com/ORG/REPO/.github/workflows/ci.yaml@refs/heads/main"
```

See [docs/security.md](docs/security.md) for details.

## Release Process

### Versioning

Mantl follows [Semantic Versioning](https://semver.org/):
- **MAJOR** - Incompatible API changes
- **MINOR** - Backwards-compatible functionality
- **PATCH** - Backwards-compatible bug fixes

### Automated Releases

Releases are automated using **release-please**:

1. **Merge commits to main** using Conventional Commits
2. **release-please opens a release PR** automatically
3. **Review and merge the release PR**
4. **release-please creates a GitHub release** with changelog
5. **CI signs and publishes artifacts** automatically

### Manual Release (Emergency Only)

```bash
# Tag a version
git tag v1.2.3
git push origin v1.2.3

# CI will automatically build, sign, and publish
```

### Artifact Verification

All releases are signed with Cosign:

```bash
cosign verify --certificate-oidc-issuer \
  https://token.actions.githubusercontent.com \
  ghcr.io/thomasvincent/mantl:v1.2.3
```

## CI/CD Pipeline

### CI Checks

Our GitHub Actions CI runs:

1. **Linting** - Black, Ruff, yamllint, shellcheck
2. **Testing** - pytest unit and integration tests
3. **Validation** - Terraform, Kustomize, Kyverno
4. **Security** - Trivy scanning, SBOM generation
5. **E2E** - kind cluster deployment test

**Required checks** are listed in [docs/ci-required-checks.md](docs/ci-required-checks.md).

### Running CI Locally

```bash
# Run all checks
make ci

# Run specific checks
make lint
make test
make validate
```

## Getting Help

- **Documentation**: Check the [docs/](docs/) directory
- **Issues**: Open an issue for bugs or feature requests
- **Discussions**: Use GitHub Discussions for questions
- **Slack**: Join our community Slack (link in README)

## Code of Conduct

All contributors must adhere to our [Code of Conduct](code-of-conduct.md).

We are committed to providing a welcoming and inclusive environment for everyone.

## License

By contributing to Mantl, you agree that your contributions will be licensed under the Apache 2.0 License.

---

**Thank you for contributing to Mantl!** 🚀
