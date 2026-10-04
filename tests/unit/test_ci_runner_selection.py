"""Exercise the actual CI selection and aggregate scripts without cloud access."""

import os
import subprocess
from pathlib import Path

import pytest
import yaml


@pytest.fixture
def workflow(project_root: Path) -> dict:
    return yaml.safe_load((project_root / ".github/workflows/ci.yml").read_text())


@pytest.mark.parametrize(
    ("is_pr", "do_configured", "pat_configured", "expected"),
    [
        ("false", "", "", "false"),
        ("false", "configured", "", "false"),
        ("false", "", "configured", "false"),
        ("false", "configured", "configured", "true"),
        ("true", "configured", "configured", "false"),
        ("true", "", "", "false"),
    ],
)
def test_runner_selection(
    workflow: dict,
    tmp_path: Path,
    is_pr: str,
    do_configured: str,
    pat_configured: str,
    expected: str,
) -> None:
    output = tmp_path / "output"
    summary = tmp_path / "summary"
    script = workflow["jobs"]["runner-config"]["steps"][0]["run"]
    result = subprocess.run(
        ["bash", "-euo", "pipefail", "-c", script],
        env={
            **os.environ,
            "IS_PULL_REQUEST": is_pr,
            "DO_TOKEN": do_configured,
            "GH_RUNNER_PAT": pat_configured,
            "GITHUB_OUTPUT": str(output),
            "GITHUB_STEP_SUMMARY": str(summary),
        },
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 0, result.stderr
    assert output.read_text().strip() == f"cloud={expected}"
    assert "configured" not in result.stdout


@pytest.mark.parametrize(
    ("config", "cloud", "start", "build", "stop", "passes"),
    [
        ("success", "false", "skipped", "success", "skipped", True),
        ("success", "true", "success", "success", "success", True),
        ("failure", "false", "skipped", "success", "skipped", False),
        ("success", "", "skipped", "success", "skipped", False),
        ("success", "false", "skipped", "failure", "skipped", False),
        ("success", "false", "skipped", "skipped", "skipped", False),
        ("success", "true", "failure", "skipped", "success", False),
        ("success", "true", "success", "failure", "success", False),
        ("success", "true", "success", "success", "failure", False),
        ("success", "true", "success", "success", "skipped", False),
    ],
)
def test_ci_requires_success_on_selected_backend(
    workflow: dict,
    config: str,
    cloud: str,
    start: str,
    build: str,
    stop: str,
    passes: bool,
) -> None:
    script = workflow["jobs"]["ci"]["steps"][0]["run"]
    result = subprocess.run(
        ["bash", "-euo", "pipefail", "-c", script],
        env={
            **os.environ,
            "CONFIG_RESULT": config,
            "CLOUD_RUNNER": cloud,
            "START_RESULT": start,
            "BUILD_RESULT": build,
            "STOP_RESULT": stop,
        },
        capture_output=True,
        text=True,
        check=False,
    )
    assert (result.returncode == 0) is passes
