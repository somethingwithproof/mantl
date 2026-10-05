#!/usr/bin/env python3
"""Assemble release installation assets without machine-specific state."""

import argparse
import gzip
import io
import json
import re
import subprocess
import tarfile
from pathlib import Path


def assemble(source: Path, dest: Path, tag: str, image: str) -> None:
    if not re.fullmatch(r"v[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?", tag):
        raise ValueError("release tag must be SemVer")
    if not re.fullmatch(r"ghcr\.io/[a-z0-9._/-]+@sha256:[a-f0-9]{64}", image):
        raise ValueError("operator image must be a GHCR digest reference")
    dest.mkdir(parents=True, exist_ok=True)
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
    tracked = set(
        subprocess.check_output(["git", "ls-files", "-z"], cwd=source)
        .decode()
        .strip("\0")
        .split("\0")
    )
    with (
        (dest / f"mantl-platform_{tag}.tar.gz").open("wb") as output,
        gzip.GzipFile(fileobj=output, mode="wb", mtime=0, filename="") as compressed,
        tarfile.open(fileobj=compressed, mode="w") as archive,
    ):
        for prefix in [
            "infra/terraform",
            "deploy",
            "templates/backstage",
            "docs/adr",
            "compliance/frameworks",
            "compliance/integration",
            "compliance/remediation.yaml",
        ]:
            root = source / prefix
            files = root.rglob("*") if root.is_dir() else [root]
            for file in sorted(files):
                if file.relative_to(source).as_posix() not in tracked:
                    continue
                if any(part in {".terraform", ".mantl"} for part in file.parts):
                    continue
                if file.is_symlink():
                    raise ValueError("platform bundle cannot contain symlinks")
                if not file.is_file():
                    continue
                if file.name.endswith(
                    (".tfstate", ".tfstate.backup", ".tfvars", ".tfvars.json", ".pem", ".key")
                ):
                    continue
                data = file.read_bytes()
                info = tarfile.TarInfo(file.relative_to(source).as_posix())
                info.size, info.mode = len(data), 0o644
                archive.addfile(info, io.BytesIO(data))
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


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--source", type=Path, default=Path("."))
    parser.add_argument("--dist", type=Path, default=Path("dist/release"))
    parser.add_argument("--tag", required=True)
    parser.add_argument("--image", required=True)
    args = parser.parse_args()
    assemble(args.source.resolve(), args.dist, args.tag, args.image)
