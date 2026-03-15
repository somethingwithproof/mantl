# Mantl Test Suite

Comprehensive test suite for the mantl Kubernetes platform.

## Test Structure

```
tests/
├── unit/              # Unit tests (fast, no external dependencies)
│   ├── test_smoke.py              # Basic smoke tests and tool availability
│   ├── test_kyverno_policies.py   # Kyverno policy validation
│   └── test_terraform.py          # Terraform configuration tests
├── integration/       # Integration tests (require infrastructure)
│   └── molecule/
│       └── kubernetes/
│           └── test_nomad_integration.py
└── e2e/              # End-to-end tests (require full environment)
```

## Running Tests

### Prerequisites

Install test dependencies:
```bash
pip install -r requirements.txt -r requirements-test.txt
```

Or using `uv` (recommended):
```bash
uv pip install -r requirements.txt -r requirements-test.txt
```

Install optional tools for full test coverage:
- **kustomize**: For Kubernetes manifest validation
- **terraform**: For Terraform configuration validation
- **kyverno CLI**: For Kyverno policy testing
- **kubectl**: For Kubernetes API validation

### Run All Tests

```bash
pytest
```

### Run Specific Test Categories

```bash
# Unit tests only (fast)
pytest tests/unit/

# Integration tests
pytest tests/integration/

# Specific test file
pytest tests/unit/test_kyverno_policies.py

# Specific test function
pytest tests/unit/test_terraform.py::test_aws_eks_blueprint_validates
```

### Run with Coverage

```bash
pytest --cov=. --cov-report=html
# Open htmlcov/index.html to view coverage report
```

### Run Tests in Parallel

```bash
pytest -n auto  # Uses all available CPU cores
```

### Run Tests with Markers

```bash
# Only tests that don't require cluster
pytest -m "not requires_cluster"

# Only Terraform tests
pytest -m requires_terraform

# Skip slow tests
pytest -m "not slow"
```

## Test Categories

### Unit Tests

**Purpose**: Fast tests that don't require external dependencies

**What they test**:
- Configuration syntax validation
- Policy structure and metadata
- Terraform configuration validity
- File structure and organization
- Tool availability checks

**Run time**: < 30 seconds

### Integration Tests

**Purpose**: Tests that require infrastructure components

**What they test**:
- Kubernetes-Nomad integration
- Cross-platform service discovery
- Policy enforcement in real cluster
- Multi-region orchestration

**Run time**: 1-5 minutes

**Requirements**:
- Running Kubernetes cluster (optional)
- Test infrastructure provisioned (optional)

### End-to-End Tests

**Purpose**: Full workflow validation

**What they test**:
- Complete deployment workflows
- Platform upgrades
- Disaster recovery scenarios
- Security compliance end-to-end

**Run time**: 10+ minutes

**Requirements**:
- Full test environment
- Cloud provider credentials (optional)

## Writing Tests

### Test Naming Convention

- Test files: `test_*.py`
- Test functions: `test_*`
- Test classes: `Test*`

### Using Fixtures

```python
import pytest

@pytest.fixture
def example_config():
    return {"key": "value"}

def test_something(example_config):
    assert example_config["key"] == "value"
```

### Parametrized Tests

```python
@pytest.mark.parametrize("input,expected", [
    ("ghcr.io/org/app:v1", True),
    ("unknown.io/app:v1", False),
])
def test_registry_validation(input, expected):
    result = validate_registry(input)
    assert result == expected
```

### Skipping Tests

```python
# Skip if tool not available
@pytest.mark.skipif(not shutil.which("kustomize"), reason="kustomize not installed")
def test_kustomize_build():
    ...

# Skip with conditional
def test_requires_cluster():
    if not cluster_available():
        pytest.skip("Kubernetes cluster not available")
    ...
```

### Test Markers

Mark tests with categories:

```python
@pytest.mark.unit
def test_fast_operation():
    ...

@pytest.mark.integration
@pytest.mark.requires_cluster
def test_deployment():
    ...

@pytest.mark.slow
def test_long_running():
    ...
```

## Continuous Integration

Tests run automatically on:
- Push to `main` branch
- Pull requests
- Manual workflow dispatch

See `.github/workflows/ci.yml` for CI configuration.

### CI Test Matrix

- Python 3.12
- Ubuntu latest
- All test categories (unit, integration with mocks)

## Test Coverage Goals

Current target: **60%+ coverage**

Priority areas for coverage:
1. ✅ Kyverno policies - security critical
2. ✅ Terraform configurations - infrastructure critical
3. ✅ Kustomize manifests - deployment critical
4. ⚠️  Compliance module - needs implementation
5. ⚠️  Monitoring configs - needs expansion

## Troubleshooting

### Tests Skipped

**Problem**: Many tests show as "skipped"

**Solution**: Install optional dependencies:
```bash
# macOS
brew install kustomize terraform kubectl kyverno

# Linux (example)
# Install from official sources or package manager
```

### Terraform Tests Fail

**Problem**: Terraform validation errors

**Solution**:
1. Ensure Terraform is installed and in PATH
2. Run `terraform fmt -recursive` to format files
3. Check for missing required files

### Kyverno Tests Fail

**Problem**: Kyverno CLI not found

**Solution**:
```bash
# Install Kyverno CLI
curl -LO https://github.com/kyverno/kyverno/releases/download/v1.12.1/kyverno-cli_v1.12.1_linux_x86_64.tar.gz
tar -xzf kyverno-cli_v1.12.1_linux_x86_64.tar.gz
sudo mv kyverno /usr/local/bin/
```

### Integration Tests Fail

**Problem**: Missing MOLECULE_INVENTORY_FILE

**Solution**: Integration tests are optional and require Molecule setup. They will skip if environment not configured.

## Contributing

When adding new features:

1. **Write tests first** (TDD approach recommended)
2. **Ensure tests pass** before committing
3. **Maintain coverage** at or above 60%
4. **Document test requirements** if new tools needed
5. **Use appropriate markers** for categorization

## Resources

- [pytest documentation](https://docs.pytest.org/)
- [pytest-cov documentation](https://pytest-cov.readthedocs.io/)
- [Kyverno CLI](https://kyverno.io/docs/kyverno-cli/)
- [Terraform testing best practices](https://www.terraform.io/docs/language/modules/testing-experiment.html)
