# SPDX-License-Identifier: Apache-2.0
"""Check tracked SPDX declarations without changing source or upstream files."""

import subprocess
import sys
import tomllib
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
TAG = "SPDX-License-Identifier:"
OPENINGS = ("//", "#", "/*", "<!--", "..", "{{/*")
ENDINGS = ("*/ -}}", "*/}}", "-->", "*/")


def header_identifier(path):
    """Read only a leading declaration, preserving binary and symlink targets."""
    if path.is_symlink():
        return None
    with path.open("rb") as stream:
        prefix = stream.read(4096).decode("utf-8", errors="replace")
    for line in prefix.splitlines()[:20]:
        line = line.strip()
        for opening in OPENINGS:
            if line.startswith(opening):
                line = line.removeprefix(opening).strip()
                break
        if line.startswith(TAG):
            identifier = line.removeprefix(TAG).strip()
            for ending in ENDINGS:
                identifier = identifier.removesuffix(ending).strip()
            return identifier
    return None


def matches(name, pattern):
    """Support exact paths and explicit directory/** annotation scopes."""
    if pattern.endswith("/**"):
        return name.startswith(pattern[:-2])
    if "*" in pattern or "?" in pattern:
        raise ValueError("license annotations require exact paths or directory/**")
    return name == pattern


def annotations(root):
    """Validate the supported REUSE metadata subset before resolving files."""
    data = tomllib.loads((root / "REUSE.toml").read_text())
    if data.get("version") != 1:
        raise ValueError("unsupported REUSE metadata version")
    entries = data.get("annotations", [])
    for entry in entries:
        if entry.get("precedence", "closest") != "closest":
            raise ValueError("license metadata must preserve closest upstream notices")
        patterns = entry["path"]
        patterns = [patterns] if isinstance(patterns, str) else patterns
        if not patterns or any(not isinstance(pattern, str) for pattern in patterns):
            raise ValueError("license annotation paths must be strings")
        for pattern in patterns:
            if pattern.startswith("/") or ".." in pattern.split("/"):
                raise ValueError("license annotation paths must stay in the checkout")
            matches("", pattern)
        entry["path"] = patterns
    return entries


def identifier(root, name, entries):
    """Prefer companion/header declarations; the last fallback table wins."""
    file = root / name
    companion = Path(str(file) + ".license")
    if companion.is_file():
        return header_identifier(companion)
    direct = header_identifier(file)
    if direct is not None:
        return direct
    if name.startswith("LICENSES/") and file.suffix == ".txt":
        return file.stem
    fallback = None
    for entry in entries:
        if any(matches(name, pattern) for pattern in entry["path"]):
            fallback = entry.get("SPDX-License-Identifier")
    return fallback


def check(root, names):
    """Return paths missing a declared ID and its accompanying license text."""
    entries = annotations(root)
    known = {file.stem for file in (root / "LICENSES").glob("*.txt")}
    failures = []
    for name in names:
        value = identifier(root, name, entries)
        if not isinstance(value, str) or value not in known:
            failures.append(name)
    return failures


def main():
    names = subprocess.check_output(["git", "ls-files", "-z"], cwd=ROOT).decode().split("\0")
    names = [name for name in names if name]
    failures = check(ROOT, names)
    if failures:
        for name in failures:
            print(f"Missing/unsupported SPDX declaration or license text: {name}", file=sys.stderr)
        return 1
    print(f"Checked SPDX declarations and license texts for {len(names)} tracked files")
    return 0


if __name__ == "__main__":
    sys.exit(main())
