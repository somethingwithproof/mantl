"""Verify local RC packages and reviewed-plan behavior without publishing."""

import argparse
import hashlib
import json
import platform
import re
import subprocess
import tarfile
import tempfile
from pathlib import Path

from scripts import release_paths

ROOT = Path(__file__).resolve().parents[1]
RC_VERSION = re.compile(
    r"(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)\.(?:0|[1-9][0-9]*)-rc\.[1-9][0-9]*", re.ASCII
)
ARCHIVE_FILES = {
    "mantl",
    "LICENSE",
    "README.md",
    "docs/releases.md",
    "docs/quickstart-platform.md",
    "docs/_static/mantl-mark.svg",
}


def expected_artifacts(version):
    if not RC_VERSION.fullmatch(version):
        raise ValueError("candidate version must be SemVer MAJOR.MINOR.PATCH-rc.N")
    archives = {
        f"mantl_{version}_{os}_{arch}.tar.gz"
        for os in ("linux", "darwin")
        for arch in ("amd64", "arm64")
    }
    packages = {
        f"mantl_{version}_linux_{arch}.{kind}"
        for arch in ("amd64", "arm64")
        for kind in ("deb", "rpm")
    }
    return archives | packages


def verify_checksums(directory, version):
    expected = expected_artifacts(version)
    seen = set()
    manifest = release_paths.artifact_path(directory / "cli-checksums.txt")
    for line in manifest.read_text().splitlines():
        digest, name = line.split(maxsplit=1)
        if name not in expected or name in seen or not re.fullmatch(r"[a-f0-9]{64}", digest):
            raise ValueError("candidate checksum inventory is invalid")
        file = release_paths.artifact_path(directory / name)
        with file.open("rb") as stream:
            if hashlib.file_digest(stream, "sha256").hexdigest() != digest:
                raise ValueError(f"candidate checksum mismatch: {name}")
        seen.add(name)
    if seen != expected:
        raise ValueError("candidate checksum inventory is incomplete")
    return sorted(seen)


def verify_archive(path):
    with tarfile.open(path) as archive:
        entries = archive.getmembers()
        names = {entry.name for entry in entries}
        if len(names) != len(entries) or not ARCHIVE_FILES.issubset(names):
            raise ValueError("candidate archive omits CLI, license, quickstart or logo")
        if any(not entry.isfile() or entry.size > 64 * 1024 * 1024 for entry in entries):
            raise ValueError("candidate archive contains unsupported entries")
        return archive.extractfile("mantl").read()


def verify_native_cli(directory, version):
    system = platform.system().lower()
    architecture = {"x86_64": "amd64", "AMD64": "amd64", "aarch64": "arm64", "arm64": "arm64"}.get(
        platform.machine()
    )
    if system not in {"linux", "darwin"} or architecture is None:
        raise ValueError("candidate CLI smoke check requires a supported native host")
    payload = verify_archive(directory / f"mantl_{version}_{system}_{architecture}.tar.gz")
    with tempfile.TemporaryDirectory(prefix="candidate-smoke-", dir=directory) as temporary:
        work = Path(temporary)
        binary = work / "mantl"
        binary.write_bytes(payload)
        binary.chmod(0o700)
        identity = subprocess.check_output([str(binary), "--version"], text=True).strip()
        commit = subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip()
        if identity != f"mantl version {version} ({commit})":
            raise ValueError("candidate CLI identity does not match this checkout/version")
        spec = ROOT / "examples/mantl-spec.yaml"
        reviewed = work / "reviewed.json"
        generated = work / "generated"
        subprocess.run(
            [
                str(binary),
                "plan",
                str(spec),
                "--output-dir",
                str(generated),
                "--save-plan",
                str(reviewed),
            ],
            cwd=work,
            check=True,
            stdout=subprocess.DEVNULL,
        )
        subprocess.run(
            [
                str(binary),
                "plan",
                str(spec),
                "--dry-run",
                "--compare",
                str(reviewed),
                "--fail-on-change",
            ],
            cwd=work,
            check=True,
            stdout=subprocess.DEVNULL,
        )
        changed = work / "changed.yaml"
        changed.write_text(
            spec.read_text().replace("name: prod-us-east", "name: candidate-changed", 1)
        )
        response = subprocess.run(
            [
                str(binary),
                "plan",
                str(changed),
                "--dry-run",
                "--compare",
                str(reviewed),
                "--fail-on-change",
            ],
            cwd=work,
            capture_output=True,
            text=True,
        )
        if response.returncode == 0 or not response.stdout:
            raise ValueError("candidate read-only change gate accepted changed inputs")
        saved = json.loads(reviewed.read_text())
        if saved["schemaVersion"] != "mantl.io/plan/v1alpha2":
            raise ValueError("candidate plan schema is unexpected")
        return {
            "identity": identity,
            "artifactCount": len(saved["artifacts"]),
            "schemaVersion": saved["schemaVersion"],
            "unchangedComparison": "passed",
            "changedInputGate": "passed",
        }


def verify(version, directory, native_packages=False):
    directory = release_paths.artifact_path(directory)
    artifacts = verify_checksums(directory, version)
    for name in artifacts:
        if name.endswith(".tar.gz"):
            verify_archive(directory / name)
    report = {
        "schemaVersion": "mantl.io/candidate-verification/v1alpha1",
        "version": version,
        "scope": "local-unsigned-snapshot",
        "artifacts": artifacts,
        "cli": verify_native_cli(directory, version),
        "worktreeDirty": bool(subprocess.check_output(["git", "status", "--porcelain"], cwd=ROOT)),
        "sourceDiffSHA256": hashlib.sha256(
            subprocess.check_output(["git", "diff", "--no-ext-diff", "--binary", "HEAD"], cwd=ROOT)
        ).hexdigest(),
        "nativePackages": "not-run",
        "signatureVerification": "requires-hosted-release",
        "provenanceVerification": "requires-hosted-release",
        "cloudAcceptance": "not-run",
    }
    if native_packages:
        subprocess.run(
            ["bash", str(ROOT / "ci/verify-cli-packages.sh"), str(directory)], cwd=ROOT, check=True
        )
        report["nativePackages"] = "passed"
    (directory / "candidate-verification.json").write_text(json.dumps(report, indent=2) + "\n")
    return report


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("version")
    parser.add_argument("--directory", type=Path, required=True)
    parser.add_argument(
        "--native-packages",
        action="store_true",
        help="Run existing isolated Docker DEB/RPM lifecycle checks",
    )
    args = parser.parse_args()
    print(json.dumps(verify(args.version, args.directory, args.native_packages), indent=2))


if __name__ == "__main__":
    main()
