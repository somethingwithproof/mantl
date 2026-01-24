"""
Pytest configuration and shared fixtures for end-to-end tests.

This file is automatically loaded by pytest and provides:
- Global fixtures available to all tests
- Pytest hooks for test execution
- Custom markers for test categorization
"""
import subprocess
import pytest


def pytest_configure(config):
    """Register custom markers."""
    config.addinivalue_line(
        "markers", "slow: marks tests as slow (deselect with '-m \"not slow\"')"
    )
    config.addinivalue_line(
        "markers", "requires_cluster: marks tests that require a running Kubernetes cluster"
    )
    config.addinivalue_line(
        "markers", "destructive: marks tests that may modify cluster state"
    )
    config.addinivalue_line(
        "markers", "terraform: marks tests related to Terraform validation"
    )
    config.addinivalue_line(
        "markers", "security: marks tests related to security verification"
    )
    config.addinivalue_line(
        "markers", "platform: marks tests related to platform components"
    )
    config.addinivalue_line(
        "markers", "dr: marks tests related to disaster recovery"
    )
    config.addinivalue_line(
        "markers", "performance: marks tests related to performance"
    )


@pytest.fixture(scope="session")
def kubectl_configured():
    """Verify kubectl is configured and can access a cluster."""
    result = subprocess.run(
        ["kubectl", "cluster-info"],
        capture_output=True,
        timeout=10
    )

    if result.returncode != 0:
        pytest.skip("kubectl not configured or cluster not accessible")

    return True


@pytest.fixture(scope="session")
def cluster_version(kubectl_configured):
    """Get the Kubernetes cluster version."""
    result = subprocess.run(
        ["kubectl", "version", "-o", "json"],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        import json
        version_info = json.loads(result.stdout)
        return version_info.get("serverVersion", {})

    return None


@pytest.fixture
def unique_namespace(kubectl_configured):
    """Create a unique test namespace and clean up after test."""
    import time

    namespace = f"test-{int(time.time())}"

    # Create namespace
    subprocess.run(
        ["kubectl", "create", "namespace", namespace],
        capture_output=True
    )

    yield namespace

    # Cleanup
    subprocess.run(
        ["kubectl", "delete", "namespace", namespace, "--force", "--grace-period=0"],
        capture_output=True,
        timeout=60
    )


def pytest_collection_modifyitems(config, items):
    """Automatically mark tests based on their location."""
    for item in items:
        # Mark tests based on directory
        if "terraform" in str(item.fspath):
            item.add_marker(pytest.mark.terraform)
        elif "security" in str(item.fspath):
            item.add_marker(pytest.mark.security)
            item.add_marker(pytest.mark.requires_cluster)
        elif "platform" in str(item.fspath):
            item.add_marker(pytest.mark.platform)
            item.add_marker(pytest.mark.requires_cluster)
        elif "dr" in str(item.fspath):
            item.add_marker(pytest.mark.dr)
            item.add_marker(pytest.mark.requires_cluster)
            item.add_marker(pytest.mark.slow)
        elif "performance" in str(item.fspath):
            item.add_marker(pytest.mark.performance)
            item.add_marker(pytest.mark.requires_cluster)
            item.add_marker(pytest.mark.slow)


def pytest_report_header(config):
    """Add custom information to pytest report header."""
    result = subprocess.run(
        ["kubectl", "version", "--short"],
        capture_output=True,
        text=True
    )

    if result.returncode == 0:
        return f"Kubernetes Cluster:\n{result.stdout}"

    return "Kubernetes Cluster: Not accessible"
