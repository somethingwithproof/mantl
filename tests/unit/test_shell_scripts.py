# SPDX-License-Identifier: Apache-2.0
"""
Test suite for shell scripts validation.

This module validates shell scripts in the bin/ directory for:
- Syntax correctness (shellcheck)
- Proper error handling (set -e, set -u)
- Function documentation
- Security best practices
"""

import os
import re
import shutil
import subprocess

import pytest


@pytest.fixture
def bin_dir(project_root):
    """Get the bin directory path."""
    return project_root / "bin"


@pytest.fixture
def shell_scripts(bin_dir):
    """Get all shell scripts in bin directory."""
    if not bin_dir.exists():
        pytest.skip("bin directory does not exist")
    scripts = list(bin_dir.glob("*.sh"))
    if not scripts:
        pytest.skip("No shell scripts found in bin/")
    return scripts


class TestShellScriptSyntax:
    """Test shell script syntax and structure."""

    @pytest.mark.shell_scripts
    def test_scripts_have_shebang(self, shell_scripts):
        """Verify all shell scripts have proper shebang."""
        for script in shell_scripts:
            content = script.read_text()
            first_line = content.split("\n")[0] if content else ""
            assert first_line.startswith("#!"), (
                f"{script.name}: Missing shebang line. "
                "Scripts should start with #!/usr/bin/env bash or similar"
            )

    @pytest.mark.shell_scripts
    def test_scripts_use_bash(self, shell_scripts):
        """Verify scripts use bash (not sh for better features)."""
        for script in shell_scripts:
            content = script.read_text()
            first_line = content.split("\n")[0] if content else ""
            assert "bash" in first_line.lower() or "#!/bin/sh" not in first_line, (
                f"{script.name}: Consider using bash for better error handling. "
                "Use #!/usr/bin/env bash"
            )

    @pytest.mark.shell_scripts
    @pytest.mark.skipif(not shutil.which("shellcheck"), reason="shellcheck not installed")
    def test_shellcheck_passes(self, shell_scripts):
        """Run shellcheck on all scripts."""
        for script in shell_scripts:
            result = subprocess.run(
                ["shellcheck", "--severity=error", "-x", "-s", "bash", str(script)],
                capture_output=True,
                text=True,
            )
            assert (
                result.returncode == 0
            ), f"{script.name}: shellcheck found issues:\n{result.stdout}\n{result.stderr}"


class TestShellScriptSafety:
    """Test shell script safety practices."""

    @pytest.mark.shell_scripts
    @pytest.mark.security
    def test_scripts_have_strict_mode(self, shell_scripts):
        """Verify scripts use strict mode (set -e, set -u, set -o pipefail)."""
        strict_patterns = [
            r"set\s+-[euo]+",  # set -euo or variants
            r"set\s+-o\s+errexit",
            r"set\s+-o\s+pipefail",
        ]

        for script in shell_scripts:
            content = script.read_text()

            # Check for at least one strict mode setting
            has_strict = any(re.search(pattern, content) for pattern in strict_patterns)

            # Allow scripts that explicitly document no strict mode
            if "# Intentionally not using strict mode" in content:
                continue

            assert has_strict, (
                f"{script.name}: Should use strict mode. "
                "Add 'set -euo pipefail' near the top of the script"
            )

    @pytest.mark.shell_scripts
    @pytest.mark.security
    def test_no_hardcoded_secrets(self, shell_scripts):
        """Check for potential hardcoded secrets."""
        secret_patterns = [
            r"password\s*=\s*['\"][^$][^'\"]+['\"]",
            r"api_key\s*=\s*['\"][^$][^'\"]+['\"]",
            r"secret\s*=\s*['\"][^$][^'\"]+['\"]",
            r"AWS_SECRET_ACCESS_KEY\s*=\s*['\"][^$][^'\"]+['\"]",
            r"token\s*=\s*['\"][^$][^'\"]+['\"]",
        ]

        for script in shell_scripts:
            content = script.read_text()
            for pattern in secret_patterns:
                matches = re.findall(pattern, content, re.IGNORECASE)
                assert not matches, (
                    f"{script.name}: Potential hardcoded secret found. "
                    "Use environment variables instead"
                )

    @pytest.mark.shell_scripts
    @pytest.mark.security
    def test_no_eval_usage(self, shell_scripts):
        """Check for dangerous eval usage."""
        for script in shell_scripts:
            content = script.read_text()
            # Match eval but not in comments
            lines = content.split("\n")
            for i, line in enumerate(lines, 1):
                stripped = line.strip()
                if stripped.startswith("#"):
                    continue
                if re.search(r"\beval\b", stripped):
                    # Allow if documented as intentional
                    if "# shellcheck disable=SC2086" in content:
                        continue
                    pytest.fail(
                        f"{script.name}:{i}: Avoid using 'eval' - it can be a security risk"
                    )


class TestShellScriptDocumentation:
    """Test shell script documentation."""

    @pytest.mark.shell_scripts
    def test_scripts_have_header_comment(self, shell_scripts):
        """Verify scripts have header documentation."""
        for script in shell_scripts:
            content = script.read_text()
            lines = content.split("\n")

            # Find first non-shebang, non-empty line. A `set -euo pipefail`
            # directive placed right after the shebang is good practice and
            # should not end the search before the descriptive header.
            has_header = False
            for line in lines[1:10]:  # Check first 10 lines after shebang
                stripped = line.strip()
                if stripped.startswith("#") and len(stripped) > 2:
                    has_header = True
                    break
                if stripped.startswith(("set ", "set\t")):
                    continue
                if stripped and not stripped.startswith("#"):
                    break

            assert has_header, f"{script.name}: Should have header documentation explaining purpose"

    @pytest.mark.shell_scripts
    def test_scripts_have_usage_function(self, shell_scripts):
        """Check that scripts with arguments have usage function."""
        for script in shell_scripts:
            content = script.read_text()

            # Check if script takes arguments
            if "$1" in content or "${1" in content or '"$@"' in content:
                has_usage = (
                    "usage()" in content
                    or "usage ()" in content
                    or "function usage" in content
                    or "--help" in content
                )
                assert (
                    has_usage
                ), f"{script.name}: Scripts that take arguments should have a usage function"


class TestShellScriptFunctionality:
    """Test shell script specific functionality."""

    @pytest.mark.shell_scripts
    def test_bootstrap_cluster_validates_inputs(self, bin_dir):
        """Test that bootstrap-cluster.sh validates environment and provider."""
        script = bin_dir / "bootstrap-cluster.sh"
        if not script.exists():
            pytest.skip("bootstrap-cluster.sh not found")

        content = script.read_text()

        # Check for input validation
        assert (
            "validate_environment" in content or "case" in content
        ), "bootstrap-cluster.sh should validate environment input"
        assert (
            "validate_provider" in content or "case" in content
        ), "bootstrap-cluster.sh should validate provider input"

    @pytest.mark.shell_scripts
    def test_bootstrap_cluster_has_rollback(self, bin_dir):
        """Test that bootstrap-cluster.sh supports rollback."""
        script = bin_dir / "bootstrap-cluster.sh"
        if not script.exists():
            pytest.skip("bootstrap-cluster.sh not found")

        content = script.read_text()

        assert (
            "rollback" in content.lower()
        ), "bootstrap-cluster.sh should support rollback functionality"

    @pytest.mark.shell_scripts
    def test_scripts_are_executable(self, shell_scripts):
        """Verify all shell scripts are executable."""
        for script in shell_scripts:
            assert os.access(script, os.X_OK), (
                f"{script.name}: Script should be executable. " f"Run: chmod +x {script}"
            )


class TestShellScriptIntegration:
    """Integration tests for shell scripts (require external tools)."""

    @pytest.mark.shell_scripts
    @pytest.mark.integration
    @pytest.mark.skipif(not shutil.which("bash"), reason="bash not available")
    def test_scripts_have_valid_bash_syntax(self, shell_scripts):
        """Verify bash syntax is valid using bash -n."""
        for script in shell_scripts:
            content = script.read_text()
            if "#!/usr/bin/env bash" in content or "#!/bin/bash" in content:
                result = subprocess.run(["bash", "-n", str(script)], capture_output=True, text=True)
                assert result.returncode == 0, f"{script.name}: Bash syntax error:\n{result.stderr}"

    @pytest.mark.shell_scripts
    @pytest.mark.integration
    def test_bootstrap_cluster_help(self, bin_dir):
        """Test that bootstrap-cluster.sh --help works."""
        script = bin_dir / "bootstrap-cluster.sh"
        if not script.exists():
            pytest.skip("bootstrap-cluster.sh not found")

        result = subprocess.run([str(script), "--help"], capture_output=True, text=True)

        # Should exit with 0 for help
        assert result.returncode == 0, f"--help should exit with 0, got {result.returncode}"
        assert (
            "usage" in result.stdout.lower() or "usage" in result.stderr.lower()
        ), "--help should display usage information"
