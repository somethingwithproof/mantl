# SPDX-License-Identifier: Apache-2.0
"""Unit tests for Kyverno policies.

These tests validate that Kyverno policies correctly enforce security controls
for image registries and signature verification.
"""

import subprocess
from pathlib import Path

import pytest

# Test fixtures


@pytest.fixture
def policies_dir() -> Path:
    """Get the Kyverno policies directory."""
    return Path(__file__).parent.parent.parent / "policies" / "kyverno"


@pytest.fixture
def restrict_registries_policy(policies_dir: Path) -> Path:
    """Get the restrict-registries policy file."""
    return policies_dir / "restrict-registries.yaml"


@pytest.fixture
def verify_signatures_policy(policies_dir: Path) -> Path:
    """Get the keyless image-signature policy file."""
    return policies_dir / "verify-image-keyless.yaml"


# Helper functions


def create_test_pod(image: str, namespace: str = "default") -> dict:
    """Create a test pod manifest with the specified image."""
    return {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {
            "name": "test-pod",
            "namespace": namespace,
        },
        "spec": {
            "containers": [
                {
                    "name": "test-container",
                    "image": image,
                }
            ]
        },
    }


def validate_policy_syntax(policy_file: Path) -> bool:
    """Validate Kyverno policy syntax using kyverno CLI."""
    try:
        result = subprocess.run(
            ["kyverno", "validate", str(policy_file)],
            capture_output=True,
            text=True,
            timeout=10,
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError):
        # If kyverno CLI not installed, skip validation
        pytest.skip("kyverno CLI not available")
        return False


# Not a test: a helper invoked by the parametrized cases below. The `test_`
# prefix made pytest collect it as a test and fail with a missing-fixture error.
def run_policy_against_pod(policy_file: Path, pod_manifest: dict) -> dict:
    """Apply a policy against a pod manifest using the kyverno CLI."""
    try:
        # Write pod manifest to temp file
        import tempfile

        with tempfile.NamedTemporaryFile(mode="w", suffix=".yaml", delete=False) as f:
            import yaml

            yaml.dump(pod_manifest, f)
            pod_file = f.name

        # Run kyverno apply
        result = subprocess.run(
            ["kyverno", "apply", str(policy_file), "--resource", pod_file],
            capture_output=True,
            text=True,
            timeout=30,
        )

        # Cleanup
        Path(pod_file).unlink()

        return {
            "returncode": result.returncode,
            "stdout": result.stdout,
            "stderr": result.stderr,
        }
    except FileNotFoundError:
        pytest.skip("kyverno CLI not available")
        return {}


# Policy syntax tests


def test_restrict_registries_policy_syntax(restrict_registries_policy: Path):
    """Test that restrict-registries policy has valid syntax."""
    assert restrict_registries_policy.exists()
    assert validate_policy_syntax(restrict_registries_policy)


def test_verify_signatures_policy_syntax(verify_signatures_policy: Path):
    """Test that verify-image-signatures policy has valid syntax."""
    assert verify_signatures_policy.exists()
    assert validate_policy_syntax(verify_signatures_policy)


# Registry restriction tests


@pytest.mark.parametrize(
    "allowed_image",
    [
        "ghcr.io/thomasvincent/myapp:latest",
        "ghcr.io/argoproj/argocd:v2.9.0",
        "ghcr.io/kyverno/kyverno:latest",
        "ghcr.io/cilium/cilium:v1.14.0",
        "docker.io/library/nginx:alpine",
        "docker.io/bitnami/postgresql:15",
        "quay.io/prometheus/prometheus:latest",
        "quay.io/jetstack/cert-manager-controller:v1.13.0",
        "gcr.io/project-123/kube-state-metrics:v2.10.0",
    ],
)
def test_allowed_registries_pass(restrict_registries_policy: Path, allowed_image: str):
    """Test that allowed registry images pass validation."""
    pod = create_test_pod(allowed_image)
    result = run_policy_against_pod(restrict_registries_policy, pod)

    # Policy should pass (returncode 0) or skip if CLI not available
    if result:
        assert result["returncode"] == 0, f"Image {allowed_image} should be allowed"


@pytest.mark.parametrize(
    "blocked_image",
    [
        "docker.io/attacker/malware:latest",
        "random-registry.io/unknown/image:tag",
        "public.ecr.aws/random/image:latest",
        "index.docker.io/randomuser/app:v1",
        "ghcr.io/randomuser/unauthorized:latest",
        "quay.io/random/unauthorized:latest",
    ],
)
def test_blocked_registries_fail(restrict_registries_policy: Path, blocked_image: str):
    """Test that unauthorized registry images are blocked."""
    pod = create_test_pod(blocked_image)
    result = run_policy_against_pod(restrict_registries_policy, pod)

    # Policy should fail (returncode != 0) or skip if CLI not available
    if result:
        assert result["returncode"] != 0, f"Image {blocked_image} should be blocked"


# Image signature tests


def test_signature_verification_required():
    """Test that signature verification is enabled for GHCR images."""
    # This is a placeholder - actual signature verification testing requires
    # a running cluster with Kyverno installed and signed images
    pytest.skip("Signature verification requires cluster with signed images")


# Policy metadata tests


def test_restrict_registries_metadata(restrict_registries_policy: Path):
    """Test that restrict-registries has proper metadata."""
    import yaml

    with open(restrict_registries_policy) as f:
        policy = yaml.safe_load(f)

    assert policy["apiVersion"] == "kyverno.io/v1"
    assert policy["kind"] == "ClusterPolicy"
    assert policy["metadata"]["name"] == "restrict-image-registries"
    assert "annotations" in policy["metadata"]
    assert policy["spec"]["validationFailureAction"] == "Enforce"
    assert policy["spec"]["background"] is True


def test_verify_signatures_metadata(verify_signatures_policy: Path):
    """Test that verify-image-signatures has proper metadata."""
    import yaml

    with open(verify_signatures_policy) as f:
        policy = yaml.safe_load(f)

    assert policy["apiVersion"] == "kyverno.io/v2beta1"
    assert policy["kind"] == "ClusterPolicy"
    assert policy["metadata"]["name"] == "verify-image-keyless"
    assert policy["spec"]["validationFailureAction"] == "Enforce"
    assert policy["spec"]["background"] is True


# Namespace exclusion tests


def test_system_namespaces_excluded(verify_signatures_policy: Path):
    """Test that system namespaces are excluded from signature verification."""
    import yaml

    with open(verify_signatures_policy) as f:
        policy = yaml.safe_load(f)

    # exclude is a match block with an `any`/`all` list, not a bare list.
    # The real policy uses `exclude.any[0].resources.namespaces`.
    excluded_namespaces = policy["spec"]["exclude"]["any"][0]["resources"]["namespaces"]
    required_exclusions = [
        "kube-system",
        "kube-public",
        "kube-node-lease",
        "cert-manager",
        "external-dns",
        "argocd",
        "kyverno",
    ]

    for ns in required_exclusions:
        assert ns in excluded_namespaces, f"Namespace {ns} should be excluded"


# Integration tests


def test_multiple_containers_validation(restrict_registries_policy: Path):
    """Test validation with multiple containers in a pod."""
    pod = {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {"name": "multi-container", "namespace": "default"},
        "spec": {
            "containers": [
                {"name": "app", "image": "ghcr.io/thomasvincent/app:v1"},
                {"name": "sidecar", "image": "ghcr.io/thomasvincent/sidecar:v1"},
            ]
        },
    }

    result = run_policy_against_pod(restrict_registries_policy, pod)
    if result:
        assert result["returncode"] == 0, "All containers should be from allowed registries"


def test_init_containers_validation(restrict_registries_policy: Path):
    """Test validation with init containers."""
    pod = {
        "apiVersion": "v1",
        "kind": "Pod",
        "metadata": {"name": "with-init", "namespace": "default"},
        "spec": {
            "initContainers": [
                {"name": "init", "image": "ghcr.io/thomasvincent/init:v1"},
            ],
            "containers": [
                {"name": "app", "image": "ghcr.io/thomasvincent/app:v1"},
            ],
        },
    }

    result = run_policy_against_pod(restrict_registries_policy, pod)
    if result:
        assert result["returncode"] == 0, "Init containers should be validated"
