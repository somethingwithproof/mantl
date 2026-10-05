"""Constrain release tooling to this checkout and its disposable dist tree."""

from pathlib import Path

WORKSPACE_ROOT = Path(__file__).resolve().parents[1]


def source_path(value: Path) -> Path:
    resolved = value.resolve()
    if resolved != WORKSPACE_ROOT.resolve():
        raise ValueError("release source must be this checkout")
    return resolved


def artifact_path(value: Path) -> Path:
    root = WORKSPACE_ROOT.resolve()
    candidate = value if value.is_absolute() else root / value
    if ".." in value.parts or candidate.is_symlink():
        raise ValueError("release paths must be canonical and cannot be symlinks")
    resolved = candidate.resolve()
    if not resolved.is_relative_to(root / "dist"):
        raise ValueError("release artifacts must stay inside this checkout's dist directory")
    return resolved
