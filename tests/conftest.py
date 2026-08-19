"""Shared pytest fixtures and configuration for Mantl tests.

This module provides:
- Common fixtures for file paths, temp directories, and mock data
- Pytest hooks for test setup/teardown
- Custom markers registration
"""

import os
import shutil
import subprocess
import tempfile
from collections.abc import Generator
from pathlib import Path
from typing import Any

import pytest

# ============================================================================
# Path Fixtures
# ============================================================================


@pytest.fixture(scope="session")
def project_root() -> Path:
    """Return the project root directory."""
    return Path(__file__).parent.parent


@pytest.fixture(scope="session")
def tests_dir(project_root: Path) -> Path:
    """Return the tests directory."""
    return project_root / "tests"


@pytest.fixture(scope="session")
def ci_dir(project_root: Path) -> Path:
    """Return the CI scripts directory."""
    return project_root / "ci"


@pytest.fixture(scope="session")
def infra_dir(project_root: Path) -> Path:
    """Return the infrastructure directory."""
    return project_root / "infra"


@pytest.fixture(scope="session")
def apps_dir(project_root: Path) -> Path:
    """Return the apps directory."""
    return project_root / "apps"


@pytest.fixture(scope="session")
def platform_dir(project_root: Path) -> Path:
    """Return the platform directory."""
    return project_root / "platform"


@pytest.fixture(scope="session")
def policies_dir(project_root: Path) -> Path:
    """Return the policies directory."""
    return project_root / "policies"


@pytest.fixture(scope="session")
def clusters_dir(project_root: Path) -> Path:
    """Return the clusters directory."""
    return project_root / "clusters"


# ============================================================================
# Tool Availability Fixtures
# ============================================================================


@pytest.fixture(scope="session")
def has_kustomize() -> bool:
    """Check if kustomize is available."""
    return shutil.which("kustomize") is not None


@pytest.fixture(scope="session")
def has_kubectl() -> bool:
    """Check if kubectl is available."""
    return shutil.which("kubectl") is not None


@pytest.fixture(scope="session")
def has_terraform() -> bool:
    """Check if terraform or opentofu is available."""
    return shutil.which("terraform") is not None or shutil.which("tofu") is not None


@pytest.fixture(scope="session")
def has_helm() -> bool:
    """Check if helm is available."""
    return shutil.which("helm") is not None


@pytest.fixture(scope="session")
def has_docker() -> bool:
    """Check if docker is available and running."""
    if not shutil.which("docker"):
        return False
    try:
        result = subprocess.run(
            ["docker", "info"],
            capture_output=True,
            text=True,
            timeout=10,
            check=False,
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError):
        return False


# ============================================================================
# Temporary Directory Fixtures
# ============================================================================


@pytest.fixture
def temp_dir() -> Generator[Path, None, None]:
    """Create a temporary directory for test artifacts."""
    tmpdir = tempfile.mkdtemp(prefix="mantl_test_")
    yield Path(tmpdir)
    shutil.rmtree(tmpdir, ignore_errors=True)


@pytest.fixture
def temp_kustomize_dir(temp_dir: Path) -> Path:
    """Create a temporary directory with a basic kustomization.yaml."""
    kustomization = temp_dir / "kustomization.yaml"
    kustomization.write_text(
        """apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources: []
"""
    )
    return temp_dir


# ============================================================================
# Mock Data Fixtures
# ============================================================================


@pytest.fixture
def sample_deployment_yaml() -> str:
    """Return sample Kubernetes deployment YAML."""
    # Use pinned image version for reproducibility and security
    return """apiVersion: apps/v1
kind: Deployment
metadata:
  name: test-app
  labels:
    app: test
spec:
  replicas: 1
  selector:
    matchLabels:
      app: test
  template:
    metadata:
      labels:
        app: test
    spec:
      containers:
      - name: app
        image: nginx:1.27.3-alpine
        ports:
        - containerPort: 80
        securityContext:
          runAsNonRoot: true
          readOnlyRootFilesystem: true
          allowPrivilegeEscalation: false
"""


@pytest.fixture
def sample_service_yaml() -> str:
    """Return sample Kubernetes service YAML."""
    return """apiVersion: v1
kind: Service
metadata:
  name: test-app
spec:
  selector:
    app: test
  ports:
  - port: 80
    targetPort: 80
"""


@pytest.fixture
def sample_configmap_yaml() -> str:
    """Return sample Kubernetes ConfigMap YAML."""
    return """apiVersion: v1
kind: ConfigMap
metadata:
  name: test-config
data:
  config.yaml: |
    key: value
    nested:
      key: value
"""


# ============================================================================
# Environment Fixtures
# ============================================================================


@pytest.fixture
def clean_env() -> Generator[dict[str, str], None, None]:
    """Provide a clean environment without sensitive variables."""
    original_env = os.environ.copy()
    sensitive_vars = [
        "AWS_ACCESS_KEY_ID",
        "AWS_SECRET_ACCESS_KEY",
        "CLOUDFLARE_API_TOKEN",
        "GITHUB_TOKEN",
        "KUBECONFIG",
    ]
    for var in sensitive_vars:
        os.environ.pop(var, None)
    yield os.environ.copy()
    os.environ.clear()
    os.environ.update(original_env)


# ============================================================================
# Pytest Hooks
# ============================================================================


def pytest_configure(config: Any) -> None:
    """Configure pytest with custom markers."""
    config.addinivalue_line("markers", "slow: marks tests as slow")
    config.addinivalue_line("markers", "integration: marks tests as integration tests")
    config.addinivalue_line("markers", "e2e: marks tests as end-to-end tests")
    config.addinivalue_line("markers", "requires_docker: marks tests that require Docker")
    config.addinivalue_line(
        "markers", "requires_kubernetes: marks tests that require a Kubernetes cluster"
    )
    config.addinivalue_line(
        "markers", "requires_cluster: marks tests that require a reachable Kubernetes cluster"
    )


def pytest_collection_modifyitems(config: Any, items: list[Any]) -> None:
    """Modify test collection to skip tests based on available tools."""
    skip_docker = pytest.mark.skip(reason="Docker not available")
    skip_kubernetes = pytest.mark.skip(reason="Kubernetes cluster not available")

    has_docker_available = shutil.which("docker") is not None
    has_kubectl_available = shutil.which("kubectl") is not None
    has_cluster_available = False
    if has_kubectl_available:
        try:
            result = subprocess.run(
                ["kubectl", "cluster-info"],
                capture_output=True,
                text=True,
                timeout=10,
                check=False,
            )
            has_cluster_available = result.returncode == 0
        except (subprocess.TimeoutExpired, OSError):
            has_cluster_available = False

    for item in items:
        if "requires_docker" in item.keywords and not has_docker_available:
            item.add_marker(skip_docker)
        if (
            "requires_kubernetes" in item.keywords or "requires_cluster" in item.keywords
        ) and not has_cluster_available:
            item.add_marker(skip_kubernetes)


# ============================================================================
# Helper Functions
# ============================================================================


def run_command(
    cmd: list[str],
    cwd: Path | None = None,
    timeout: int = 60,
    check: bool = True,
) -> subprocess.CompletedProcess[str]:
    """Run a command and return the result.

    Args:
        cmd: Command to run as list of strings
        cwd: Working directory for the command
        timeout: Timeout in seconds
        check: Whether to raise on non-zero exit code

    Returns:
        CompletedProcess with stdout and stderr

    Raises:
        subprocess.CalledProcessError: If check=True and command fails
        subprocess.TimeoutExpired: If command times out
    """
    return subprocess.run(
        cmd,
        cwd=cwd,
        capture_output=True,
        text=True,
        timeout=timeout,
        check=check,
    )
