"""Reject ambiguous or invalid package inputs before invoking Docker."""

import os
import subprocess

import pytest


@pytest.mark.parametrize("problem", ["missing", "duplicate", "symlink", "architecture"])
def test_invalid_package_inputs_do_not_start_containers(project_root, tmp_path, problem):
    artifacts = tmp_path / "artifacts with spaces"
    artifacts.mkdir()
    for extension in ["tar.gz", "deb", "rpm"]:
        (artifacts / f"mantl_0.3.1_linux_arm64.{extension}").write_bytes(b"unused")
    if problem == "missing":
        (artifacts / "mantl_0.3.1_linux_arm64.rpm").unlink()
    elif problem == "duplicate":
        (artifacts / "mantl_0.3.2_linux_arm64.deb").write_bytes(b"unused")
    elif problem == "symlink":
        archive = artifacts / "mantl_0.3.1_linux_arm64.tar.gz"
        archive.unlink()
        outside = tmp_path / "outside.tar.gz"
        outside.write_bytes(b"unused")
        archive.symlink_to(outside)
    docker_called = tmp_path / "docker-called"
    tools = tmp_path / "tools"
    tools.mkdir()
    docker = tools / "docker"
    docker.write_text('#!/bin/sh\ntouch "$DOCKER_CALLED"\nexit 91\n')
    docker.chmod(0o755)
    result = subprocess.run(
        [
            "bash",
            str(project_root / "ci/verify-cli-packages.sh"),
            str(artifacts),
            "riscv64" if problem == "architecture" else "arm64",
        ],
        env={
            **os.environ,
            "PATH": f"{tools}:{os.environ['PATH']}",
            "DOCKER_CALLED": str(docker_called),
        },
        capture_output=True,
        text=True,
        check=False,
    )
    assert result.returncode == 2, result.stderr
    assert not docker_called.exists()
