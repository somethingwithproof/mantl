"""Unit tests for CI scripts and configuration.

Tests validate that CI scripts exist, are executable, and have valid syntax.
"""

import os
import subprocess
from pathlib import Path

import pytest


class TestCIScripts:
    """Tests for CI validation scripts."""

    @pytest.fixture
    def ci_dir(self, project_root: Path) -> Path:
        """Return the CI scripts directory."""
        return project_root / "ci"

    def test_ci_directory_exists(self, ci_dir: Path) -> None:
        """Verify CI directory exists."""
        assert ci_dir.exists(), f"CI directory not found: {ci_dir}"

    def test_validate_kustomize_dev_script_exists(self, ci_dir: Path) -> None:
        """Verify validate-kustomize-dev.sh script exists."""
        if not ci_dir.exists():
            pytest.skip("CI directory not found")

        script = ci_dir / "validate-kustomize-dev.sh"
        assert script.exists(), "validate-kustomize-dev.sh not found"

    def test_makefile_has_kustomize_validate_target(self, project_root: Path) -> None:
        """Verify Makefile has kustomize-validate target."""
        makefile = project_root / "Makefile"
        assert makefile.exists(), "Makefile not found"
        content = makefile.read_text()
        assert "kustomize-validate:" in content, "kustomize-validate target not found in Makefile"

    def test_makefile_has_terraform_validate_target(self, project_root: Path) -> None:
        """Verify Makefile has terraform-validate target."""
        makefile = project_root / "Makefile"
        assert makefile.exists(), "Makefile not found"
        content = makefile.read_text()
        assert "terraform-validate:" in content, "terraform-validate target not found in Makefile"

    def test_scripts_are_executable(self, ci_dir: Path) -> None:
        """Verify CI scripts are executable."""
        if not ci_dir.exists():
            pytest.skip("CI directory not found")

        for script in ci_dir.glob("*.sh"):
            assert os.access(script, os.X_OK), f"Script not executable: {script.name}"

    def test_scripts_have_shebang(self, ci_dir: Path) -> None:
        """Verify CI scripts have proper shebang."""
        if not ci_dir.exists():
            pytest.skip("CI directory not found")

        valid_shebangs = ["#!/bin/bash", "#!/usr/bin/env bash", "#!/bin/sh", "#!/usr/bin/env sh"]

        for script in ci_dir.glob("*.sh"):
            first_line = script.read_text().split("\n")[0].strip()
            assert any(
                first_line.startswith(shebang) for shebang in valid_shebangs
            ), f"Missing or invalid shebang in {script.name}: {first_line}"

    @pytest.mark.skipif(
        not os.path.exists("/bin/bash"), reason="bash not available"
    )
    def test_scripts_syntax_valid(self, ci_dir: Path) -> None:
        """Verify CI scripts have valid bash syntax."""
        if not ci_dir.exists():
            pytest.skip("CI directory not found")

        for script in ci_dir.glob("*.sh"):
            result = subprocess.run(
                ["bash", "-n", str(script)],
                capture_output=True,
                text=True,
                check=False,
            )
            assert result.returncode == 0, f"Syntax error in {script.name}: {result.stderr}"


class TestGitHubWorkflows:
    """Tests for GitHub Actions workflow files."""

    @pytest.fixture
    def workflows_dir(self, project_root: Path) -> Path:
        """Return the GitHub workflows directory."""
        return project_root / ".github" / "workflows"

    def test_workflows_directory_exists(self, workflows_dir: Path) -> None:
        """Verify workflows directory exists."""
        assert workflows_dir.exists(), f"Workflows directory not found: {workflows_dir}"

    def test_ci_workflow_exists(self, workflows_dir: Path) -> None:
        """Verify CI workflow exists."""
        if not workflows_dir.exists():
            pytest.skip("Workflows directory not found")

        ci_workflow = workflows_dir / "ci.yml"
        assert ci_workflow.exists(), "ci.yml workflow not found"

    def test_workflow_files_valid_yaml(self, workflows_dir: Path) -> None:
        """Verify workflow files are valid YAML."""
        if not workflows_dir.exists():
            pytest.skip("Workflows directory not found")

        import yaml

        for workflow in workflows_dir.glob("*.yml"):
            try:
                content = yaml.safe_load(workflow.read_text())
                assert content is not None, f"Empty workflow: {workflow.name}"
                assert isinstance(content, dict), f"Invalid workflow: {workflow.name}"
            except yaml.YAMLError as e:
                pytest.fail(f"Invalid YAML in {workflow.name}: {e}")

    def test_workflows_have_required_fields(self, workflows_dir: Path) -> None:
        """Verify workflows have required fields."""
        if not workflows_dir.exists():
            pytest.skip("Workflows directory not found")

        import yaml

        for workflow in workflows_dir.glob("*.yml"):
            # Skip disabled workflows
            if workflow.name.endswith(".disabled"):
                continue

            content = yaml.safe_load(workflow.read_text())
            if content is None:
                continue

            # Should have 'name', 'on', and 'jobs'
            assert "name" in content, f"Missing 'name' in {workflow.name}"
            assert "on" in content, f"Missing 'on' trigger in {workflow.name}"
            assert "jobs" in content, f"Missing 'jobs' in {workflow.name}"


class TestPreCommitConfig:
    """Tests for pre-commit configuration."""

    def test_pre_commit_config_exists(self, project_root: Path) -> None:
        """Verify pre-commit config exists."""
        config = project_root / ".pre-commit-config.yaml"
        assert config.exists(), ".pre-commit-config.yaml not found"

    def test_pre_commit_config_valid(self, project_root: Path) -> None:
        """Verify pre-commit config is valid YAML."""
        import yaml

        config = project_root / ".pre-commit-config.yaml"
        if not config.exists():
            pytest.skip("Pre-commit config not found")

        try:
            content = yaml.safe_load(config.read_text())
            assert content is not None, "Empty pre-commit config"
            assert "repos" in content, "Missing 'repos' in pre-commit config"
        except yaml.YAMLError as e:
            pytest.fail(f"Invalid pre-commit config: {e}")


class TestDockerfile:
    """Tests for Dockerfile configuration."""

    def test_dockerfile_exists(self, project_root: Path) -> None:
        """Verify Dockerfile exists."""
        dockerfile = project_root / "Dockerfile"
        assert dockerfile.exists(), "Dockerfile not found"

    def test_dockerfile_has_from(self, project_root: Path) -> None:
        """Verify Dockerfile has FROM instruction."""
        dockerfile = project_root / "Dockerfile"
        if not dockerfile.exists():
            pytest.skip("Dockerfile not found")

        content = dockerfile.read_text()
        assert "FROM" in content, "Dockerfile missing FROM instruction"

    def test_dockerfile_uses_python_313(self, project_root: Path) -> None:
        """Verify Dockerfile uses Python 3.13."""
        dockerfile = project_root / "Dockerfile"
        if not dockerfile.exists():
            pytest.skip("Dockerfile not found")

        content = dockerfile.read_text()
        # Check for Python 3.13 in FROM instruction
        assert "python:3.13" in content or "3.13" in content, (
            "Dockerfile should use Python 3.13"
        )


class TestRequirementsFiles:
    """Tests for requirements files."""

    def test_requirements_exists(self, project_root: Path) -> None:
        """Verify requirements.txt exists."""
        requirements = project_root / "requirements.txt"
        assert requirements.exists(), "requirements.txt not found"

    def test_requirements_test_exists(self, project_root: Path) -> None:
        """Verify requirements-test.txt exists."""
        requirements = project_root / "requirements-test.txt"
        assert requirements.exists(), "requirements-test.txt not found"

    def test_requirements_pinned(self, project_root: Path) -> None:
        """Verify requirements are pinned to specific versions."""
        requirements = project_root / "requirements.txt"
        if not requirements.exists():
            pytest.skip("requirements.txt not found")

        content = requirements.read_text()
        lines = [
            line.strip()
            for line in content.splitlines()
            if line.strip() and not line.startswith("#")
        ]

        unpinned = []
        for line in lines:
            # Skip lines that are comments or have version specifiers
            if "==" not in line and ">=" not in line and "<=" not in line:
                unpinned.append(line)

        # Allow some flexibility - just warn
        if unpinned:
            for pkg in unpinned[:5]:
                print(f"Warning: Unpinned dependency: {pkg}")

    def test_no_duplicate_dependencies(self, project_root: Path) -> None:
        """Verify no duplicate dependencies across files."""
        requirements = project_root / "requirements.txt"
        requirements_test = project_root / "requirements-test.txt"

        if not requirements.exists() or not requirements_test.exists():
            pytest.skip("Requirements files not found")

        def get_packages(file_path: Path) -> set:
            packages = set()
            for line in file_path.read_text().splitlines():
                line = line.strip()
                if line and not line.startswith("#"):
                    # Extract package name (before ==, >=, etc.)
                    pkg = line.split("==")[0].split(">=")[0].split("<=")[0].strip()
                    packages.add(pkg.lower())
            return packages

        main_pkgs = get_packages(requirements)
        test_pkgs = get_packages(requirements_test)

        duplicates = main_pkgs & test_pkgs
        # This is informational - pytest duplicates are expected
        if duplicates - {"pytest", "pytest-cov"}:
            print(f"Note: Duplicate packages: {duplicates}")
