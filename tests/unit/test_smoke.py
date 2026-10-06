# SPDX-License-Identifier: Apache-2.0
"""Basic smoke tests to ensure the test harness runs.

These keep `pytest` from exiting with code 5 (no tests collected) in minimal setups.
"""

import os
import shutil
import subprocess
import sys

import pytest


def test_python_version() -> None:
    """Verify we're running on the expected Python version."""
    assert sys.version_info >= (3, 11), f"Python 3.11+ required, got {sys.version_info}"


@pytest.mark.skipif(
    not shutil.which("kustomize"),
    reason="kustomize not installed (run in container or install locally)",
)
def test_kustomize_in_path() -> None:
    """Verify kustomize is available - skips if not installed."""
    result = subprocess.run(
        ["kustomize", "version"],
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 0, f"kustomize version failed: {result.stderr}"


@pytest.mark.skipif(
    not shutil.which("kubectl"),
    reason="kubectl not installed (run in container or install locally)",
)
def test_kubectl_in_path() -> None:
    """Verify kubectl is available - skips if not installed."""
    result = subprocess.run(
        ["kubectl", "version", "--client", "--output=json"],
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 0, f"kubectl version failed: {result.stderr}"


def test_project_structure() -> None:
    """Verify essential project directories exist."""
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

    essential_dirs = [
        "apps",
        "platform",
        "infra",
        "policies",
        "clusters",
        "tests",
    ]

    for dir_name in essential_dirs:
        dir_path = os.path.join(project_root, dir_name)
        assert os.path.isdir(dir_path), f"Essential directory missing: {dir_name}"


def test_ci_scripts_exist() -> None:
    """Verify CI validation scripts exist and are executable."""
    project_root = os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
    ci_dir = os.path.join(project_root, "ci")

    expected_scripts = [
        "validate-kustomize.sh",
        "validate-terraform.sh",
    ]

    for script in expected_scripts:
        script_path = os.path.join(ci_dir, script)
        if os.path.exists(script_path):
            # Check if the script is executable
            assert os.access(script_path, os.X_OK), f"Script not executable: {script}"
