# SPDX-License-Identifier: Apache-2.0
"""Terraform validation must fail when any owned module is invalid."""

import os
import subprocess

import pytest


@pytest.mark.parametrize("invalid", [False, True])
def test_validation_retains_failures_across_modules(project_root, tmp_path, invalid):
    for name in ["blueprints/one", "modules/two"]:
        module = tmp_path / "infra/terraform" / name
        module.mkdir(parents=True)
        (module / "main.tf").write_text("# fixture")
    if invalid:
        (tmp_path / "infra/terraform/blueprints/one/invalid").touch()
    tools = tmp_path / "tools"
    tools.mkdir()
    terraform = tools / "terraform"
    terraform.write_text(
        '#!/usr/bin/env bash\nif [ "$1" = validate ] && [ -f invalid ]; then exit 1; fi\n'
    )
    terraform.chmod(0o755)
    result = subprocess.run(
        ["bash", str(project_root / "ci/validate-terraform.sh")],
        cwd=tmp_path,
        env={**os.environ, "PATH": str(tools) + os.pathsep + os.environ["PATH"]},
        capture_output=True,
        text=True,
        check=False,
    )
    assert (result.returncode != 0) is invalid


def test_missing_modules_are_not_reported_as_valid(project_root, tmp_path):
    result = subprocess.run(
        ["bash", str(project_root / "ci/validate-terraform.sh")],
        cwd=tmp_path,
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode != 0
    assert "No first-party Terraform modules" in result.stderr
