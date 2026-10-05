#!/usr/bin/env python3
"""Publish content versions once; reject changed bytes at an existing tag."""

import argparse
import hashlib
import json
import re
import subprocess
import tempfile
from pathlib import Path

try:
    from scripts import release_paths
except ModuleNotFoundError as error:
    if error.name != "scripts":
        raise
    import release_paths


def publish(repository: str, version: str, archive: Path) -> str:
    archive = release_paths.artifact_path(archive)
    if not archive.is_file():
        raise ValueError("release archive must be a regular file")
    if not re.fullmatch(r"\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?", version, flags=re.ASCII):
        raise ValueError("invalid content version")
    if not re.fullmatch(r"ghcr\.io/[a-z0-9._/-]+", repository):
        raise ValueError("expected a GHCR repository")
    reference = f"{repository}:{version}"
    result = subprocess.run(["oras", "resolve", reference], capture_output=True, text=True)
    if result.returncode == 0:
        digest = result.stdout.strip()
        with tempfile.TemporaryDirectory(prefix="mantl-existing-controls-") as directory:
            subprocess.run(
                ["oras", "pull", f"{repository}@{digest}", "--output", directory], check=True
            )
            previous = Path(directory) / archive.name
            if (
                not previous.is_file()
                or hashlib.sha256(previous.read_bytes()).digest()
                != hashlib.sha256(archive.read_bytes()).digest()
            ):
                raise ValueError("published content changed; bump compliance/bundle-version.txt")
    else:
        # Authentication and transport failures must never authorize overwriting.
        if not any(
            marker in result.stderr.lower()
            for marker in ["not found", "manifest_unknown", "name_unknown"]
        ):
            raise RuntimeError("could not resolve content tag; publication refused")
        subprocess.run(
            [
                "oras",
                "push",
                reference,
                f"{archive.name}:application/vnd.oci.image.layer.v1.tar+gzip",
                "--artifact-type",
                "application/vnd.mantl.controls.v1",
            ],
            cwd=archive.parent,
            check=True,
        )
        digest = subprocess.check_output(["oras", "resolve", reference], text=True).strip()
    if not re.fullmatch(r"sha256:[a-f0-9]{64}", digest):
        raise ValueError("registry returned an invalid digest")
    pinned = f"{repository}@{digest}"
    subprocess.run(["cosign", "sign", "--yes", pinned], check=True)
    return pinned


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument("--repository", required=True)
    parser.add_argument("--version", required=True)
    parser.add_argument("--archive", type=Path, required=True)
    parser.add_argument("--identity", type=Path, required=True)
    args = parser.parse_args()
    release_paths.artifact_path(args.identity).write_text(
        json.dumps({"ociReference": publish(args.repository, args.version, args.archive.resolve())})
        + "\n"
    )
