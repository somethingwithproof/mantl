# SPDX-License-Identifier: Apache-2.0
"""Local candidate validation rejects mismatched bytes and incomplete user assets."""

import hashlib
import io
import json
import subprocess
import sys
import tarfile

import pytest

from ci import verify_release_candidate as candidate


def candidate_directory(tmp_path, monkeypatch):
    monkeypatch.setattr(candidate.release_paths, "WORKSPACE_ROOT", tmp_path)
    directory = tmp_path / "dist" / "candidate"
    directory.mkdir(parents=True)
    return directory


def test_candidate_checksums_require_exact_complete_inventory(tmp_path, monkeypatch):
    directory = candidate_directory(tmp_path, monkeypatch)
    version = "0.5.0-rc.1"
    lines = []
    for name in sorted(candidate.expected_artifacts(version)):
        data = name.encode()
        (directory / name).write_bytes(data)
        lines.append(f"{hashlib.sha256(data).hexdigest()}  {name}\n")
    manifest = directory / "cli-checksums.txt"
    manifest.write_text("".join(lines))
    assert len(candidate.verify_checksums(directory, version)) == 8
    manifest.write_text("".join(lines[:-1]))
    with pytest.raises(ValueError, match="incomplete"):
        candidate.verify_checksums(directory, version)
    manifest.write_text("".join(lines + [lines[0]]))
    with pytest.raises(ValueError, match="invalid"):
        candidate.verify_checksums(directory, version)
    manifest.write_text("".join(lines))
    (directory / sorted(candidate.expected_artifacts(version))[0]).write_bytes(b"changed")
    with pytest.raises(ValueError, match="mismatch"):
        candidate.verify_checksums(directory, version)


def test_candidate_archive_needs_working_quickstart_and_logo(tmp_path):
    path = tmp_path / "candidate.tar.gz"
    with tarfile.open(path, "w:gz") as archive:
        for name in candidate.ARCHIVE_FILES - {"docs/_static/mantl-mark.svg"}:
            data = b"fixture"
            info = tarfile.TarInfo(name)
            info.size = len(data)
            archive.addfile(info, io.BytesIO(data))
    with pytest.raises(ValueError, match="omits"):
        candidate.verify_archive(path)


@pytest.mark.parametrize(
    "version", ["0.5.0", "v0.5.0-rc.1", "00.5.0-rc.1", "0.5.0-rc.0", "../0.5.0-rc.1"]
)
def test_candidate_version_is_explicit_prerelease(version):
    with pytest.raises(ValueError, match="SemVer"):
        candidate.expected_artifacts(version)


def complete_candidate(
    tmp_path, monkeypatch, *, identity=None, schema="mantl.io/plan/v1alpha2", accept_change=False
):
    directory = candidate_directory(tmp_path, monkeypatch)
    monkeypatch.setattr(candidate.platform, "system", lambda: "Linux")
    monkeypatch.setattr(candidate.platform, "machine", lambda: "aarch64")
    version = "0.5.0-rc.1"
    commit = subprocess.check_output(
        ["git", "rev-parse", "HEAD"], cwd=candidate.ROOT, text=True
    ).strip()
    identity = identity or f"mantl version {version} ({commit})"
    binary = (
        f"#!{sys.executable}\n"
        "import json, sys\nfrom pathlib import Path\n"
        f"if '--version' in sys.argv:\n    print({identity!r})\n    sys.exit(0)\n"
        "if '--save-plan' in sys.argv:\n"
        "    target = Path(sys.argv[sys.argv.index('--save-plan') + 1])\n"
        f"    target.write_text(json.dumps({{'schemaVersion': {schema!r}, 'artifacts': []}}))\n"
        "    sys.exit(0)\n"
        "if Path(sys.argv[2]).name == 'changed.yaml':\n"
        "    print('fixture comparison: input changed')\n"
        f"    sys.exit({0 if accept_change else 1})\n"
        "sys.exit(0)\n"
    ).encode()
    checksums = []
    for name in sorted(candidate.expected_artifacts(version)):
        file = directory / name
        if name.endswith(".tar.gz"):
            with tarfile.open(file, "w:gz") as archive:
                for entry in sorted(candidate.ARCHIVE_FILES):
                    payload = binary if entry == "mantl" else b"fixture documentation"
                    info = tarfile.TarInfo(entry)
                    info.size = len(payload)
                    archive.addfile(info, io.BytesIO(payload))
        else:
            file.write_bytes(b"fixture package")
        checksums.append(f"{hashlib.sha256(file.read_bytes()).hexdigest()}  {name}\n")
    (directory / "cli-checksums.txt").write_text("".join(checksums))
    return directory


def test_verifier_runs_cli_and_records_explicit_unsigned_scope(tmp_path, monkeypatch):
    directory = complete_candidate(tmp_path, monkeypatch)
    report = candidate.verify("0.5.0-rc.1", directory)
    assert report["scope"] == "local-unsigned-snapshot"
    assert report["cli"]["unchangedComparison"] == "passed"
    assert report["cli"]["changedInputGate"] == "passed"
    assert report["signatureVerification"] == "requires-hosted-release"
    assert report["provenanceVerification"] == "requires-hosted-release"
    assert report["cloudAcceptance"] == "not-run"
    assert report["nativePackages"] == "not-run"
    assert json.loads((directory / "candidate-verification.json").read_text()) == report


@pytest.mark.parametrize(
    "identity", ["mantl version 0.4.0 (wrong)", "mantl version 0.5.0-rc.1 (wrong)"]
)
def test_verifier_rejects_wrong_cli_identity(tmp_path, monkeypatch, identity):
    directory = complete_candidate(tmp_path, monkeypatch, identity=identity)
    with pytest.raises(ValueError, match="identity"):
        candidate.verify("0.5.0-rc.1", directory)
    assert not (directory / "candidate-verification.json").exists()


def test_verifier_rejects_cli_that_accepts_changed_input(tmp_path, monkeypatch):
    directory = complete_candidate(tmp_path, monkeypatch, accept_change=True)
    with pytest.raises(ValueError, match="change gate"):
        candidate.verify("0.5.0-rc.1", directory)
    assert not (directory / "candidate-verification.json").exists()


def test_verifier_rejects_old_plan_schema(tmp_path, monkeypatch):
    directory = complete_candidate(tmp_path, monkeypatch, schema="mantl.io/plan/v1alpha1")
    with pytest.raises(ValueError, match="schema"):
        candidate.verify("0.5.0-rc.1", directory)


@pytest.mark.parametrize("succeeds", [True, False])
def test_native_package_failure_cannot_record_success(tmp_path, monkeypatch, succeeds):
    directory = complete_candidate(tmp_path, monkeypatch)
    original_run = subprocess.run
    calls = []

    def run(args, **kwargs):
        if args[0] != "bash":
            return original_run(args, **kwargs)
        calls.append((args, kwargs))
        if not succeeds:
            raise subprocess.CalledProcessError(1, args)
        return subprocess.CompletedProcess(args, 0)

    monkeypatch.setattr(candidate.subprocess, "run", run)
    if succeeds:
        report = candidate.verify("0.5.0-rc.1", directory, native_packages=True)
        assert report["nativePackages"] == "passed"
    else:
        with pytest.raises(subprocess.CalledProcessError):
            candidate.verify("0.5.0-rc.1", directory, native_packages=True)
        assert not (directory / "candidate-verification.json").exists()
    assert len(calls) == 1
    assert calls[0][0] == [
        "bash",
        str(candidate.ROOT / "ci/verify-cli-packages.sh"),
        str(directory),
    ]
    assert calls[0][1]["check"] is True


@pytest.mark.parametrize(("system", "machine"), [("Windows", "AMD64"), ("Linux", "i386")])
def test_native_cli_requires_supported_host(tmp_path, monkeypatch, system, machine):
    monkeypatch.setattr(candidate.platform, "system", lambda: system)
    monkeypatch.setattr(candidate.platform, "machine", lambda: machine)
    with pytest.raises(ValueError, match="supported native host"):
        candidate.verify_native_cli(tmp_path, "0.5.0-rc.1")


def test_verifier_rejects_symlinked_report_before_running_cli(tmp_path, monkeypatch):
    directory = complete_candidate(tmp_path, monkeypatch)
    outside = tmp_path / "untouched.json"
    outside.write_text("fixture must stay unchanged")
    (directory / "candidate-verification.json").symlink_to(outside)
    monkeypatch.setattr(
        candidate,
        "verify_native_cli",
        lambda *args: pytest.fail("CLI ran before output validation"),
    )
    with pytest.raises(ValueError, match="symlinks"):
        candidate.verify("0.5.0-rc.1", directory)
    assert outside.read_text() == "fixture must stay unchanged"


@pytest.mark.parametrize("phase", ["identity", "plan", "packages"])
def test_verifier_bounds_commands_and_rejects_timeouts(tmp_path, monkeypatch, phase):
    directory = complete_candidate(tmp_path, monkeypatch)
    original_run = subprocess.run
    calls = []

    def run(args, **kwargs):
        calls.append((args, kwargs))
        target = (
            (phase == "identity" and args[-1] == "--version")
            or (phase == "plan" and "--save-plan" in args)
            or (phase == "packages" and args[0] == "bash")
        )
        if target:
            raise subprocess.TimeoutExpired(args, kwargs["timeout"])
        return original_run(args, **kwargs)

    monkeypatch.setattr(candidate.subprocess, "run", run)
    with pytest.raises(ValueError, match="timed out after"):
        candidate.verify("0.5.0-rc.1", directory, native_packages=phase == "packages")
    assert not (directory / "candidate-verification.json").exists()
    assert calls
    for args, kwargs in calls:
        assert kwargs["timeout"] == (
            candidate.PACKAGE_TIMEOUT_SECONDS
            if args[0] == "bash"
            else candidate.CLI_TIMEOUT_SECONDS
        )


@pytest.mark.parametrize("escape", ["outside", "traversal", "symlink"])
def test_verifier_rejects_untrusted_artifact_directory_before_execution(
    tmp_path, monkeypatch, escape
):
    directory = candidate_directory(tmp_path, monkeypatch)
    outside = tmp_path / "outside"
    outside.mkdir()
    if escape == "outside":
        directory = outside
    elif escape == "traversal":
        directory = directory / ".." / "candidate"
    else:
        alias = tmp_path / "dist" / "alias"
        alias.symlink_to(outside, target_is_directory=True)
        directory = alias
    monkeypatch.setattr(
        candidate,
        "verify_checksums",
        lambda *args: pytest.fail("read artifacts before path validation"),
    )
    with pytest.raises(ValueError, match="release paths|release artifacts"):
        candidate.verify("0.5.0-rc.1", directory)
    assert not (outside / "candidate-verification.json").exists()


def test_report_destination_substitution_cannot_overwrite_outside_file(tmp_path, monkeypatch):
    directory = complete_candidate(tmp_path, monkeypatch)
    outside = tmp_path / "untouched.json"
    outside.write_text("fixture must stay unchanged")
    original = candidate.verify_native_cli

    def substitute(*args):
        result = original(*args)
        (directory / "candidate-verification.json").symlink_to(outside)
        return result

    monkeypatch.setattr(candidate, "verify_native_cli", substitute)
    with pytest.raises(ValueError, match="regular file"):
        candidate.verify("0.5.0-rc.1", directory)
    assert outside.read_text() == "fixture must stay unchanged"
    assert not list(directory.glob(".candidate-report-*.tmp"))


def test_report_directory_substitution_does_not_redirect_output(tmp_path, monkeypatch):
    directory = complete_candidate(tmp_path, monkeypatch)
    relocated = directory.with_name("original-candidate")
    substitute = tmp_path / "substitute"
    substitute.mkdir()
    original = candidate.inspect_candidate

    def replace_directory(*args):
        report = original(*args)
        directory.rename(relocated)
        directory.symlink_to(substitute, target_is_directory=True)
        return report

    monkeypatch.setattr(candidate, "inspect_candidate", replace_directory)
    candidate.verify("0.5.0-rc.1", directory)
    assert (relocated / "candidate-verification.json").is_file()
    assert not list(substitute.iterdir())


def test_successful_verification_replaces_existing_regular_report(tmp_path, monkeypatch):
    directory = complete_candidate(tmp_path, monkeypatch)
    target = directory / "candidate-verification.json"
    target.write_text("old report")
    report = candidate.verify("0.5.0-rc.1", directory)
    assert json.loads(target.read_text()) == report
    assert not list(directory.glob(".candidate-report-*.tmp"))
