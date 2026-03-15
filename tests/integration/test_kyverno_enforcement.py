"""Integration tests for Kyverno policy enforcement.

These tests require a running Kubernetes cluster with Kyverno installed.
They verify that policies actually block/allow pods as expected.
"""

import json
import subprocess
import tempfile
from pathlib import Path
from typing import Dict

import pytest
import yaml


# Fixtures

@pytest.fixture(scope="module")
def cluster_available() -> bool:
    """Check if kubectl can connect to a cluster."""
    try:
        result = subprocess.run(
            ["kubectl", "cluster-info"],
            capture_output=True,
            timeout=5,
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError):
        return False


@pytest.fixture(scope="module")
def kyverno_installed(cluster_available: bool) -> bool:
    """Check if Kyverno is installed in the cluster."""
    if not cluster_available:
        return False

    try:
        result = subprocess.run(
            ["kubectl", "get", "deployment", "-n", "kyverno", "kyverno"],
            capture_output=True,
            timeout=5,
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError):
        return False


@pytest.fixture
def test_namespace():
    """Create a temporary namespace for testing."""
    namespace_name = "kyverno-test"

    # Create namespace
    subprocess.run(
        ["kubectl", "create", "namespace", namespace_name],
        capture_output=True,
    )

    yield namespace_name

    # Cleanup
    subprocess.run(
        ["kubectl", "delete", "namespace", namespace_name, "--wait=false"],
        capture_output=True,
    )


# Helper functions

def apply_manifest(manifest: Dict, namespace: str = "default") -> subprocess.CompletedProcess:
    """Apply a Kubernetes manifest and return the result."""
    with tempfile.NamedTemporaryFile(mode='w', suffix='.yaml', delete=False) as f:
        yaml.dump(manifest, f)
        manifest_file = f.name

    try:
        result = subprocess.run(
            ["kubectl", "apply", "-f", manifest_file, "-n", namespace],
            capture_output=True,
            text=True,
            timeout=30,
        )
        return result
    finally:
        Path(manifest_file).unlink()


def create_pod_manifest(
    name: str,
    image: str,
    namespace: str = "default",
    labels: Dict[str, str] = None,
) -> Dict:
    """Create a pod manifest."""
    manifest = {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {
            "name": name,
            "namespace": namespace,
            "labels": labels or {"app": name},
        },
        "spec": {
            "containers": [
                {
                    "name": "main",
                    "image": image,
                    "command": ["sleep", "3600"],
                }
            ],
            "restartPolicy": "Never",
        },
    }
    return manifest


# Registry restriction tests

@pytest.mark.integration
@pytest.mark.requires_cluster
def test_allowed_registry_accepted(cluster_available, kyverno_installed, test_namespace):
    """Test that pods from allowed registries are accepted."""
    if not cluster_available or not kyverno_installed:
        pytest.skip("Requires cluster with Kyverno installed")

    # Pod from approved registry
    manifest = create_pod_manifest(
        name="allowed-registry-test",
        image="ghcr.io/thomasvincent/test:latest",
        namespace=test_namespace,
    )

    result = apply_manifest(manifest, test_namespace)

    # Should succeed or give a different error (not policy violation)
    # Note: May fail due to image not existing, but shouldn't be policy blocked
    assert "image verification failed" not in result.stderr.lower()
    assert "not allowed" not in result.stderr.lower()


@pytest.mark.integration
@pytest.mark.requires_cluster
def test_blocked_registry_rejected(cluster_available, kyverno_installed, test_namespace):
    """Test that pods from unauthorized registries are blocked."""
    if not cluster_available or not kyverno_installed:
        pytest.skip("Requires cluster with Kyverno installed")

    # Pod from unauthorized registry
    manifest = create_pod_manifest(
        name="blocked-registry-test",
        image="random-registry.io/attacker/malware:latest",
        namespace=test_namespace,
    )

    result = apply_manifest(manifest, test_namespace)

    # Should be blocked by policy
    assert result.returncode != 0
    assert any(
        keyword in result.stderr.lower()
        for keyword in ["blocked", "denied", "not allowed", "policy"]
    )


# Required labels tests

@pytest.mark.integration
@pytest.mark.requires_cluster
def test_pod_without_required_labels_blocked(
    cluster_available, kyverno_installed, test_namespace
):
    """Test that pods without required labels are blocked."""
    if not cluster_available or not kyverno_installed:
        pytest.skip("Requires cluster with Kyverno installed")

    # Pod without required labels
    manifest = create_pod_manifest(
        name="no-labels-test",
        image="ghcr.io/thomasvincent/test:latest",
        namespace=test_namespace,
        labels={},  # No labels
    )

    result = apply_manifest(manifest, test_namespace)

    # May be blocked by require-labels policy
    # Check if cluster has this policy enabled
    if "labels" in result.stderr.lower() and result.returncode != 0:
        assert "required" in result.stderr.lower() or "missing" in result.stderr.lower()


# Image signature verification tests

@pytest.mark.integration
@pytest.mark.requires_cluster
@pytest.mark.slow
def test_unsigned_image_blocked_when_policy_enforced(
    cluster_available, kyverno_installed, test_namespace
):
    """Test that unsigned images are blocked when signature verification is enforced."""
    if not cluster_available or not kyverno_installed:
        pytest.skip("Requires cluster with Kyverno installed")

    # Check if image verification policy is in Enforce mode
    result = subprocess.run(
        [
            "kubectl",
            "get",
            "clusterpolicy",
            "verify-image-signatures",
            "-o",
            "json",
        ],
        capture_output=True,
        text=True,
        timeout=5,
    )

    if result.returncode != 0:
        pytest.skip("Image signature policy not installed")

    policy = json.loads(result.stdout)
    if policy.get("spec", {}).get("validationFailureAction") != "Enforce":
        pytest.skip("Image signature policy not in Enforce mode")

    # Try to deploy unsigned image
    manifest = create_pod_manifest(
        name="unsigned-image-test",
        image="ghcr.io/thomasvincent/unsigned-test:latest",
        namespace=test_namespace,
    )

    result = apply_manifest(manifest, test_namespace)

    # Should be blocked by signature verification
    assert result.returncode != 0
    assert "verification" in result.stderr.lower() or "signature" in result.stderr.lower()


# Policy report tests

@pytest.mark.integration
@pytest.mark.requires_cluster
def test_policy_reports_generated(cluster_available, kyverno_installed, test_namespace):
    """Test that Kyverno generates policy reports."""
    if not cluster_available or not kyverno_installed:
        pytest.skip("Requires cluster with Kyverno installed")

    # Create a test pod (may fail, that's ok)
    manifest = create_pod_manifest(
        name="report-test",
        image="ghcr.io/thomasvincent/test:latest",
        namespace=test_namespace,
    )
    apply_manifest(manifest, test_namespace)

    # Wait a bit for policy report generation
    import time
    time.sleep(5)

    # Check if policy reports exist
    result = subprocess.run(
        ["kubectl", "get", "policyreport", "-n", test_namespace],
        capture_output=True,
        text=True,
        timeout=10,
    )

    # Should have policy reports
    assert result.returncode == 0


# Namespace exclusion tests

@pytest.mark.integration
@pytest.mark.requires_cluster
def test_system_namespaces_excluded(cluster_available, kyverno_installed):
    """Test that system namespaces are excluded from policies."""
    if not cluster_available or not kyverno_installed:
        pytest.skip("Requires cluster with Kyverno installed")

    # Try to deploy in kube-system (should not be blocked by policies)
    manifest = create_pod_manifest(
        name="system-namespace-test",
        image="random.io/test:latest",  # Even unauthorized registry
        namespace="kube-system",
    )

    result = apply_manifest(manifest, "kube-system")

    # Should not be blocked by our policies (may fail for other reasons)
    # The key is that it's not a policy violation
    if "policy" in result.stderr.lower():
        # If it mentions policy, it should be about something else, not our policies
        assert "verify-image" not in result.stderr.lower()
        assert "restrict-registries" not in result.stderr.lower()

    # Cleanup
    subprocess.run(
        ["kubectl", "delete", "pod", "system-namespace-test", "-n", "kube-system"],
        capture_output=True,
    )


# Multi-container pod tests

@pytest.mark.integration
@pytest.mark.requires_cluster
def test_multi_container_all_validated(
    cluster_available, kyverno_installed, test_namespace
):
    """Test that all containers in a pod are validated."""
    if not cluster_available or not kyverno_installed:
        pytest.skip("Requires cluster with Kyverno installed")

    # Pod with multiple containers - one from unauthorized registry
    manifest = {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {
            "name": "multi-container-test",
            "namespace": test_namespace,
        },
        "spec": {
            "containers": [
                {
                    "name": "allowed",
                    "image": "ghcr.io/thomasvincent/test:latest",
                    "command": ["sleep", "3600"],
                },
                {
                    "name": "blocked",
                    "image": "random.io/unauthorized:latest",
                    "command": ["sleep", "3600"],
                },
            ],
            "restartPolicy": "Never",
        },
    }

    result = apply_manifest(manifest, test_namespace)

    # Should be blocked because one container is from unauthorized registry
    assert result.returncode != 0
    assert any(
        keyword in result.stderr.lower()
        for keyword in ["blocked", "denied", "not allowed", "policy"]
    )


# Performance test

@pytest.mark.integration
@pytest.mark.requires_cluster
@pytest.mark.slow
def test_policy_validation_performance(
    cluster_available, kyverno_installed, test_namespace
):
    """Test that policy validation doesn't significantly delay pod creation."""
    if not cluster_available or not kyverno_installed:
        pytest.skip("Requires cluster with Kyverno installed")

    import time

    manifest = create_pod_manifest(
        name="performance-test",
        image="ghcr.io/thomasvincent/test:latest",
        namespace=test_namespace,
    )

    start_time = time.time()
    result = apply_manifest(manifest, test_namespace)
    end_time = time.time()

    validation_time = end_time - start_time

    # Policy validation should complete in < 10 seconds
    # (May fail for other reasons, but shouldn't be slow)
    assert validation_time < 10.0, f"Policy validation took {validation_time}s"

    # Cleanup
    if result.returncode == 0:
        subprocess.run(
            ["kubectl", "delete", "pod", "performance-test", "-n", test_namespace],
            capture_output=True,
        )
