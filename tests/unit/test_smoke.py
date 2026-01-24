"""Basic smoke tests to ensure the test harness runs.
These keep `pytest` from exiting with code 5 (no tests collected) in minimal setups.
"""

import shutil
import subprocess
from pathlib import Path

import pytest


def test_pytest_runs() -> None:
    """Basic test to ensure pytest is working."""
    assert True


def test_kustomize_in_path() -> None:
    """Test that kustomize is available in PATH."""
    # kustomize is needed by CI validators
    assert shutil.which("kustomize"), "kustomize not found in PATH"


def test_kubectl_in_path() -> None:
    """Test that kubectl is available in PATH."""
    if not shutil.which("kubectl"):
        pytest.skip("kubectl not found in PATH (optional)")


def test_kyverno_cli_in_path() -> None:
    """Test that kyverno CLI is available in PATH."""
    if not shutil.which("kyverno"):
        pytest.skip("kyverno CLI not found in PATH (optional)")


def test_terraform_in_path() -> None:
    """Test that terraform is available in PATH."""
    if not shutil.which("terraform"):
        pytest.skip("terraform not found in PATH (optional)")


# Kustomize validation tests

@pytest.fixture
def k8s_dir() -> Path:
    """Get the kubernetes directory."""
    return Path(__file__).parent.parent.parent / "kubernetes"


def test_kustomize_directories_exist(k8s_dir: Path):
    """Test that kubernetes directories exist."""
    if not k8s_dir.exists():
        pytest.skip("kubernetes directory not found")

    expected_dirs = ["base", "overlays"]
    for dir_name in expected_dirs:
        dir_path = k8s_dir / dir_name
        if dir_path.exists():
            # At least one expected directory exists
            return

    pytest.skip("No kustomize directories found")


def test_kustomize_build_base(k8s_dir: Path):
    """Test that kustomize can build the base configuration."""
    if not shutil.which("kustomize"):
        pytest.skip("kustomize not available")

    base_dir = k8s_dir / "base"
    if not base_dir.exists():
        pytest.skip("base directory not found")

    if not (base_dir / "kustomization.yaml").exists():
        pytest.skip("kustomization.yaml not found in base")

    try:
        result = subprocess.run(
            ["kustomize", "build", str(base_dir)],
            capture_output=True,
            text=True,
            timeout=30,
        )
        assert result.returncode == 0, (
            f"kustomize build failed for base:\n{result.stderr}"
        )
    except subprocess.TimeoutExpired:
        pytest.fail("kustomize build timed out")


@pytest.mark.parametrize("overlay", ["dev", "staging", "production"])
def test_kustomize_build_overlays(k8s_dir: Path, overlay: str):
    """Test that kustomize can build overlay configurations."""
    if not shutil.which("kustomize"):
        pytest.skip("kustomize not available")

    overlay_dir = k8s_dir / "overlays" / overlay
    if not overlay_dir.exists():
        pytest.skip(f"{overlay} overlay not found")

    if not (overlay_dir / "kustomization.yaml").exists():
        pytest.skip(f"kustomization.yaml not found in {overlay}")

    try:
        result = subprocess.run(
            ["kustomize", "build", str(overlay_dir)],
            capture_output=True,
            text=True,
            timeout=30,
        )
        assert result.returncode == 0, (
            f"kustomize build failed for {overlay}:\n{result.stderr}"
        )
    except subprocess.TimeoutExpired:
        pytest.fail(f"kustomize build timed out for {overlay}")


# Repository structure tests

def test_required_files_exist():
    """Test that required repository files exist."""
    repo_root = Path(__file__).parent.parent.parent
    required_files = [
        "README.md",
        "requirements.txt",
        ".gitignore",
    ]

    for file_name in required_files:
        file_path = repo_root / file_name
        assert file_path.exists(), f"Required file {file_name} not found"


def test_ci_files_exist():
    """Test that CI configuration files exist."""
    repo_root = Path(__file__).parent.parent.parent
    ci_file = repo_root / ".github" / "workflows" / "ci.yml"

    if not ci_file.exists():
        pytest.skip("CI workflow file not found")
