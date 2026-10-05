"""Dependency regeneration must detect drift and leave failed updates untouched."""

import importlib.util
import subprocess
from pathlib import Path

import pytest


@pytest.fixture
def locks(tmp_path, monkeypatch):
    spec = importlib.util.spec_from_file_location(
        "python_locks", Path(__file__).parents[2] / "ci/lock_python_dependencies.py"
    )
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    monkeypatch.setattr(module, "TARGETS", ("requirements", "requirements-test"))
    monkeypatch.setattr(module.release_paths, "WORKSPACE_ROOT", tmp_path)
    for target in module.TARGETS:
        (tmp_path / f"{target}.txt").write_text("old\n")
        (tmp_path / f"{target}.in").write_text("example==1.0\n")
    return module


def test_check_reports_drift_without_changing_checkout(locks, tmp_path, monkeypatch, capsys):
    def compile_fixture(_root, target, output, constraint):
        if target == "requirements-test":
            assert constraint.read_text() == "resolved root\n"
        output.write_text("resolved root\n" if constraint is None else "resolved test\n")

    monkeypatch.setattr(locks, "compile_lock", compile_fixture)
    assert not locks.regenerate(tmp_path, check=True)
    assert "requirements-test.txt" in capsys.readouterr().out
    assert (tmp_path / "requirements.txt").read_text() == "old\n"
    assert locks.regenerate(tmp_path, check=False)
    assert (tmp_path / "requirements-test.txt").read_text() == "resolved test\n"
    assert locks.regenerate(tmp_path, check=True)


def test_resolution_failure_preserves_all_existing_locks(locks, tmp_path, monkeypatch):
    def compile_fixture(_root, target, output, _constraint):
        if target == "requirements-test":
            raise subprocess.CalledProcessError(1, ["uv", "pip", "compile"])
        output.write_text("new\n")

    monkeypatch.setattr(locks, "compile_lock", compile_fixture)
    with pytest.raises(subprocess.CalledProcessError):
        locks.regenerate(tmp_path, check=False)
    for target in locks.TARGETS:
        assert (tmp_path / f"{target}.txt").read_text() == "old\n"


def test_compilation_retains_pins_and_constrains_test_dependencies(locks, tmp_path, monkeypatch):
    directory = tmp_path / "dist"
    directory.mkdir()
    output = directory / "generated.txt"
    constraint = directory / "root-lock.txt"
    constraint.write_text("example==1.0\n")

    def run(command, **options):
        assert command[:4] == ["mise", "exec", "--", "uv"]
        assert "--generate-hashes" in command
        assert command[-2:] == ["--constraint", str(constraint)]
        assert output.read_text() == "old\n"
        assert options["check"]
        output.write_text("example==1.0 \\\n    --hash=sha256:" + "a" * 64 + "\n")

    monkeypatch.setattr(locks.subprocess, "run", run)
    locks.compile_lock(tmp_path, "requirements-test", output, constraint)
    assert output.read_text().startswith("# Generated from requirements-test.in")
    assert "--hash=sha256:" in output.read_text()


def test_compilation_rejects_external_outputs_before_running(locks, tmp_path, monkeypatch):
    def unexpected(*_args, **_kwargs):
        pytest.fail("invalid paths must not start dependency resolution")

    monkeypatch.setattr(locks.subprocess, "run", unexpected)
    with pytest.raises(ValueError, match="inside this checkout"):
        locks.compile_lock(tmp_path, "requirements", tmp_path / "outside.txt")
    with pytest.raises(ValueError, match="known requirements"):
        locks.compile_lock(tmp_path, "../outside", tmp_path / "dist/output.txt")
    assert not (tmp_path / "outside.txt").exists()
