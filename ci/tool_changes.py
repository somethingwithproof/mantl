# SPDX-License-Identifier: Apache-2.0
"""Compare verified Git revisions without treating every tool pin as shared code."""

import re
import subprocess
import tomllib


def changed_tools(before, after):
    """Return changed pins, or None when complete validation is required."""
    try:
        old, new = tomllib.loads(before), tomllib.loads(after)
    except (tomllib.TOMLDecodeError, TypeError):
        return None
    old_tools, new_tools = old.pop("tools", {}), new.pop("tools", {})
    if old != new or not isinstance(old_tools, dict) or not isinstance(new_tools, dict):
        return None
    return {
        name
        for name in old_tools.keys() | new_tools.keys()
        if old_tools.get(name) != new_tools.get(name)
    }


def from_git(base, head):
    """Fail conservatively when revisions/configuration cannot be inspected."""
    if not all(isinstance(ref, str) and re.fullmatch(r"[a-f0-9]{40}", ref) for ref in (base, head)):
        return None
    try:
        contents = [
            subprocess.run(
                ["git", "show", f"{ref}:mise.toml"],
                capture_output=True,
                text=True,
                check=True,
                timeout=20,
            ).stdout
            for ref in (base, head)
        ]
    except (subprocess.SubprocessError, OSError, UnicodeError):
        return None
    return changed_tools(*contents)
