# Testing Guide

This document describes the testing strategy, tools, and practices for the Mantl platform.

## Test Categories

### Unit Tests (`tests/unit/`)

Unit tests verify individual components in isolation without external dependencies.

```bash
# Run unit tests
make test-unit

# Run with coverage
pytest -v --cov=. --cov-report=html tests/unit/
```

**Test Files:**
- `test_smoke.py` - Basic smoke tests for environment validation
- `test_infrastructure.py` - Infrastructure file validation tests
- `test_ci_scripts.py` - CI/CD script and workflow validation

### Integration Tests (`tests/integration/`)

Integration tests verify component interactions, often requiring external services.

```bash
# Run integration tests (requires INVENTORY or MOLECULE_INVENTORY_FILE)
make test-integration

# With explicit inventory
INVENTORY=/path/to/inventory pytest -v tests/integration/
```

### E2E Tests (`tests/e2e/`)

End-to-end tests validate complete workflows in a real Kubernetes environment.

```bash
# E2E tests run in CI via Kind clusters
# See .github/workflows/ci.yml for E2E job configurations
```

## Running Tests

### Quick Start

```bash
# Install test dependencies
pip install -r requirements.txt -r requirements-test.txt

# Run all tests
make test

# Run with coverage report
make test-coverage
```

### Watch Mode (TDD)

```bash
# Install pytest-watch
pip install pytest-watch

# Run tests in watch mode
make test-watch
```

## Coverage Requirements

- **Minimum threshold:** 70% (configured in `pyproject.toml`)
- **Target threshold:** 80%
- **Coverage reports:** HTML reports generated in `htmlcov/`

### Checking Coverage

```bash
# Generate coverage report
pytest --cov=. --cov-report=html --cov-report=term-missing tests/

# View HTML report
open htmlcov/index.html
```

## Test Configuration

### pytest Configuration (`pyproject.toml`)

```toml
[tool.pytest.ini_options]
python_files = ["test_*.py"]
python_functions = ["test_*"]
addopts = "-v --maxfail=5 --tb=short"
testpaths = ["tests"]
markers = [
    "slow: marks tests as slow",
    "integration: marks tests as integration tests",
    "e2e: marks tests as end-to-end tests",
    "requires_docker: marks tests that require Docker",
    "requires_kubernetes: marks tests that require a Kubernetes cluster",
]
```

### Coverage Configuration

```toml
[tool.coverage.run]
source = ["."]
branch = true
omit = ["*/tests/*", "*/__pycache__/*", "*/site-packages/*"]

[tool.coverage.report]
fail_under = 70
show_missing = true
```

## Test Markers

Use pytest markers to categorize and selectively run tests:

```bash
# Skip slow tests
pytest -m "not slow" tests/

# Run only integration tests
pytest -m integration tests/

# Run tests that don't require Docker
pytest -m "not requires_docker" tests/
```

## Fixtures

Common fixtures are defined in `tests/conftest.py`:

| Fixture | Scope | Description |
|---------|-------|-------------|
| `project_root` | session | Path to project root directory |
| `tests_dir` | session | Path to tests directory |
| `has_kustomize` | session | Boolean: is kustomize installed |
| `has_kubectl` | session | Boolean: is kubectl installed |
| `has_docker` | session | Boolean: is Docker running |
| `temp_dir` | function | Temporary directory for test artifacts |
| `sample_deployment_yaml` | function | Sample K8s Deployment manifest |

## CI/CD Testing

Tests run automatically in GitHub Actions:

1. **Lint Job** - Python (black, ruff), YAML (yamllint)
2. **Type Check Job** - mypy static analysis
3. **Test Job** - pytest with coverage, Codecov upload
4. **Container Tests** - Tests in Docker container environment
5. **Kind Smoke Test** - Kubernetes cluster verification
6. **E2E Tests** - Full deployment tests

### Local CI Simulation

```bash
# Run all validations locally
make audit

# Format code before committing
make format

# Run security scans
make security-scan
```

## Writing New Tests

### Test File Naming

- Unit tests: `tests/unit/test_<module>.py`
- Integration tests: `tests/integration/test_<feature>_integration.py`
- E2E tests: `tests/e2e/<scenario>/`

### Test Function Naming

```python
def test_<function_name>_<scenario>() -> None:
    """Test description explaining what is being tested."""
    # Arrange
    ...
    # Act
    ...
    # Assert
    ...
```

### Example Test

```python
import pytest
from pathlib import Path

class TestMyFeature:
    """Tests for MyFeature functionality."""

    @pytest.fixture
    def my_fixture(self, project_root: Path) -> str:
        """Set up test data."""
        return project_root / "my" / "path"

    def test_feature_works(self, my_fixture: str) -> None:
        """Verify feature produces expected output."""
        result = my_function(my_fixture)
        assert result == expected_value

    @pytest.mark.slow
    def test_feature_performance(self) -> None:
        """Verify feature meets performance requirements."""
        # Long-running test
        ...

    @pytest.mark.requires_docker
    def test_feature_with_docker(self) -> None:
        """Verify feature works with Docker."""
        # Requires Docker daemon
        ...
```

## Troubleshooting

### Tests Fail Locally but Pass in CI

- Check Python version matches CI (`python --version` should be 3.13+)
- Ensure all dependencies installed: `pip install -r requirements-test.txt`
- Some tests skip if tools aren't installed (kustomize, kubectl)

### Coverage Too Low

- Add tests for untested code paths
- Check coverage report for missing lines: `open htmlcov/index.html`
- Focus on critical business logic first

### Slow Tests

- Mark slow tests with `@pytest.mark.slow`
- Run fast tests during development: `pytest -m "not slow"`
- Consider mocking external dependencies

## Resources

- [pytest Documentation](https://docs.pytest.org/)
- [pytest-cov Documentation](https://pytest-cov.readthedocs.io/)
- [Testing Best Practices](https://docs.python-guide.org/writing/tests/)
