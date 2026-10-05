"""Resolve complete, hash-locked Python dependencies with the pinned uv runtime."""

import argparse
import subprocess
import tempfile
from pathlib import Path

from scripts import release_paths

ROOT = Path(__file__).resolve().parents[1]
TARGETS = (
    "requirements",
    "requirements-test",
    "examples/batch-processing-job/src/requirements",
    "examples/event-driven-arch/publisher/requirements",
    "examples/event-driven-arch/subscriber/requirements",
    "examples/ml-inference-service/src/requirements",
    "examples/ecommerce-microservices/src/products-api/requirements",
)


def compile_lock(root, target, output, constraint=None):
    """Retain existing transitive pins until a dependency is explicitly updated."""
    root = release_paths.source_path(root)
    if target not in TARGETS:
        raise ValueError("dependency input must be a known requirements manifest")
    output = release_paths.artifact_path(output)
    if constraint is not None:
        constraint = release_paths.artifact_path(constraint)
    existing = root / f"{target}.txt"
    output.write_text(existing.read_text(encoding="utf-8"), encoding="utf-8")
    command = [
        "mise",
        "exec",
        "--",
        "uv",
        "pip",
        "compile",
        "--no-config",
        "--default-index",
        "https://pypi.org/simple",
        "--python-version",
        "3.12",
        "--python-platform",
        "x86_64-unknown-linux-gnu",
        "--generate-hashes",
        "--only-binary",
        ":all:",
        "--no-header",
        "--no-annotate",
        str(root / f"{target}.in"),
        "--output-file",
        str(output),
    ]
    if constraint is not None:
        command.extend(["--constraint", str(constraint)])
    subprocess.run(command, cwd=root, check=True, stdout=subprocess.DEVNULL)
    origin = (
        "requirements-test.in with root requirements as constraints"
        if target == "requirements-test"
        else "requirements.in with mise and uv"
    )
    output.write_text(
        f"# Generated from {origin}; edit the input and regenerate.\n"
        + output.read_text(encoding="utf-8"),
        encoding="utf-8",
    )


def regenerate(root, check):
    """Check without changing the checkout; generate all locks before replacing any."""
    root = release_paths.source_path(root)
    (root / "dist").mkdir(exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="python-locks-", dir=root / "dist") as temporary:
        directory = Path(temporary)
        compiled = {}
        for number, target in enumerate(TARGETS):
            output = directory / f"{number}.txt"
            constraint = compiled.get("requirements") if target == "requirements-test" else None
            compile_lock(root, target, output, constraint)
            compiled[target] = output
        changed = [
            target
            for target, output in compiled.items()
            if output.read_bytes() != (root / f"{target}.txt").read_bytes()
        ]
        if check:
            for target in changed:
                print(f"Dependency lock is stale: {target}.txt")
            return not changed
        for target in changed:
            (root / f"{target}.txt").write_bytes(compiled[target].read_bytes())
        return True


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true", help="Report drift without modifying files")
    arguments = parser.parse_args()
    return 0 if regenerate(ROOT, arguments.check) else 1


if __name__ == "__main__":
    raise SystemExit(main())
