#!/usr/bin/env python3
# SPDX-License-Identifier: Apache-2.0
"""Assemble release installation assets without machine-specific state."""

import argparse
import gzip
import io
import json
import re
import subprocess
import tarfile
from pathlib import Path

try:
    from scripts import release_paths
except ModuleNotFoundError as error:
    if error.name != "scripts":
        raise
    import release_paths


def assemble(source: Path, dest: Path, tag: str, image: str) -> None:
    source = release_paths.source_path(source)
    dest = release_paths.artifact_path(dest)
    if not re.fullmatch(r"v\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?", tag, flags=re.ASCII):
        raise ValueError("release tag must be SemVer")
    if not re.fullmatch(r"ghcr\.io/[a-z0-9._/-]+@sha256:[a-f0-9]{64}", image):
        raise ValueError("operator image must be a GHCR digest reference")
    dest.mkdir(parents=True, exist_ok=True)
    render_installations(source, dest, image)
    write_platform_archive(source, dest / f"mantl-platform_{tag}.tar.gz")
    (dest / "release-identity.json").write_text(
        json.dumps(
            {
                "tag": tag,
                "operatorImage": image,
                "sourceCommit": subprocess.check_output(
                    ["git", "rev-parse", "HEAD"], cwd=source, text=True
                ).strip(),
            }
        )
        + "\n"
    )


def render_installations(source: Path, dest: Path, image: str) -> None:
    rendered = subprocess.check_output(
        ["kubectl", "kustomize", str(source / "deploy/operator/base")]
    )
    rendered = rendered.replace(b"mantl-compliance-operator:development", image.encode())
    if image.encode() not in rendered:
        raise ValueError("operator image replacement failed")
    (dest / "operator-install.yaml").write_bytes(rendered)
    fleet = subprocess.check_output(["kubectl", "kustomize", str(source / "deploy/fleet/base")])
    fleet = fleet.replace(b"mantl-compliance-operator:development", image.encode())
    (dest / "fleet-install.yaml").write_bytes(fleet)


PLATFORM_PREFIXES = (
    "infra/terraform",
    "deploy",
    "templates/backstage",
    "docs/adr",
    "compliance/frameworks",
    "compliance/integration",
    "compliance/remediation.yaml",
)
PRIVATE_SUFFIXES = (".tfstate", ".tfstate.backup", ".tfvars", ".tfvars.json", ".pem", ".key")


def platform_file(source: Path, name: str) -> tuple[Path, str]:
    relative = Path(name)
    if relative.is_absolute() or ".." in relative.parts:
        raise ValueError("platform content paths must be relative and canonical")
    file = source / relative
    if file.is_symlink() or not file.resolve().is_relative_to(source):
        raise ValueError("platform bundle cannot contain symlinks or external files")
    return file, relative.as_posix()


def platform_files(source: Path):
    tracked = subprocess.check_output(["git", "ls-files", "-z"], cwd=source).decode().split("\0")
    for name in sorted(set(tracked)):
        if not any(name == prefix or name.startswith(prefix + "/") for prefix in PLATFORM_PREFIXES):
            continue
        relative = Path(name)
        if any(part in {".terraform", ".mantl"} for part in relative.parts):
            continue
        if name.endswith(PRIVATE_SUFFIXES):
            continue
        file, archive_name = platform_file(source, name)
        if file.is_file():
            yield file, archive_name


def write_platform_archive(source: Path, destination: Path) -> None:
    with (
        destination.open("wb") as output,
        gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed,
        tarfile.open(fileobj=compressed, mode="w") as archive,
    ):
        for file, name in platform_files(source):
            data = file.read_bytes()
            info = tarfile.TarInfo(name)
            info.size, info.mode = len(data), 0o644
            archive.addfile(info, io.BytesIO(data))


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, default=Path("."))
    parser.add_argument("--dist", type=Path, default=Path("dist/release"))
    parser.add_argument("--tag", required=True)
    parser.add_argument("--image", required=True)
    args = parser.parse_args()
    assemble(args.source.resolve(), args.dist, args.tag, args.image)
