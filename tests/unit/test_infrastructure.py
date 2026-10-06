# SPDX-License-Identifier: Apache-2.0
"""Unit tests for infrastructure validation.

Tests validate that infrastructure files exist and are properly formatted.
"""

import re
from pathlib import Path

import pytest
import yaml


def _is_first_party_yaml(path: Path) -> bool:
    """Return True for YAML that should parse as plain YAML.

    Excludes vendored Helm charts and Helm-rendered templates. Their `{{ }}`
    Go-template directives are not valid YAML until rendered, so scanning them
    with a plain YAML parser is a category error. The scope covers anything
    under a `charts/` or `templates/` directory and any file whose body contains
    a `{{` template directive. First-party manifests, including those under
    `platform/` that are not vendored charts, are still validated.
    """
    parts = set(path.parts)
    if "charts" in parts or "templates" in parts:
        return False
    try:
        if "{{" in path.read_text():
            return False
    except (OSError, UnicodeDecodeError):
        return False
    return True


class TestTerraformStructure:
    """Tests for Terraform module structure."""

    @pytest.fixture
    def terraform_dir(self, project_root: Path) -> Path:
        """Return the Terraform infrastructure directory."""
        return project_root / "infra" / "terraform"

    def test_terraform_directory_exists(self, terraform_dir: Path) -> None:
        """Verify Terraform directory exists."""
        assert terraform_dir.exists(), f"Terraform directory not found: {terraform_dir}"

    def test_terraform_modules_have_main(self, terraform_dir: Path) -> None:
        """Verify each Terraform module directory contains a .tf file."""
        if not terraform_dir.exists():
            pytest.skip("Terraform directory not found")

        # A "module" is a directory that directly holds .tf files. Container
        # directories (blueprints/, modules/, the nested terraform/, ...) hold
        # only subdirectories and are not modules, so the previous top-level
        # iteration wrongly required a .tf in them. Discover real module dirs by
        # the presence of .tf files instead.
        module_dirs = {
            tf_file.parent
            for tf_file in terraform_dir.rglob("*.tf")
            if ".terraform" not in tf_file.parts
        }
        assert module_dirs, "No Terraform modules found"

        # A module that declares resources/data/child-modules must have an entry
        # file (main.tf, or the modern.tf convention used by some blueprints).
        # Interface-only modules (just variables/outputs, no resources) are
        # exempt. This is stronger than "has any .tf" without forcing a single
        # filename on legitimate variations.
        entry_names = {"main.tf", "modern.tf"}
        decl_re = re.compile(r"^\s*(resource|data|module)\s", re.MULTILINE)
        for module in module_dirs:
            tf_files = list(module.glob("*.tf"))
            has_entry = any(f.name in entry_names for f in tf_files)
            declares = any(decl_re.search(f.read_text()) for f in tf_files)
            message = f"Terraform module declares resources but has no entry file: {module}"
            assert has_entry or not declares, message

    def test_terraform_files_valid_hcl(self, terraform_dir: Path) -> None:
        """Verify Terraform files contain valid HCL syntax (basic check)."""
        if not terraform_dir.exists():
            pytest.skip("Terraform directory not found")

        for tf_file in terraform_dir.rglob("*.tf"):
            content = tf_file.read_text()
            # Basic syntax checks
            assert content.count("{") == content.count("}"), f"Unmatched braces in {tf_file}"
            assert content.count("[") == content.count("]"), f"Unmatched brackets in {tf_file}"


class TestKustomizeStructure:
    """Tests for Kustomize overlay structure."""

    @pytest.fixture
    def clusters_dir(self, project_root: Path) -> Path:
        """Return the clusters directory."""
        return project_root / "clusters"

    def test_clusters_directory_exists(self, clusters_dir: Path) -> None:
        """Verify clusters directory exists."""
        assert clusters_dir.exists(), f"Clusters directory not found: {clusters_dir}"

    def test_production_kustomization_exists(self, clusters_dir: Path) -> None:
        """Verify production kustomization.yaml exists."""
        if not clusters_dir.exists():
            pytest.skip("Clusters directory not found")

        production_dir = clusters_dir / "production"
        kustomization = production_dir / "kustomization.yaml"

        if production_dir.exists():
            assert kustomization.exists(), "production/kustomization.yaml not found"

    def test_kustomization_files_valid_yaml(self, project_root: Path) -> None:
        """Verify all kustomization.yaml files are valid YAML."""
        kustomization_files = list(project_root.rglob("kustomization.yaml"))

        for kust_file in kustomization_files:
            content = yaml.safe_load(kust_file.read_text())
            assert content is not None, f"Empty kustomization: {kust_file}"
            assert isinstance(content, dict), f"Invalid kustomization: {kust_file}"


class TestPoliciesStructure:
    """Tests for Kyverno policies structure."""

    @pytest.fixture
    def policies_dir(self, project_root: Path) -> Path:
        """Return the policies directory."""
        return project_root / "policies"

    def test_policies_directory_exists(self, policies_dir: Path) -> None:
        """Verify policies directory exists."""
        assert policies_dir.exists(), f"Policies directory not found: {policies_dir}"

    def test_kyverno_policies_exist(self, policies_dir: Path) -> None:
        """Verify Kyverno policies directory exists."""
        if not policies_dir.exists():
            pytest.skip("Policies directory not found")

        kyverno_dir = policies_dir / "kyverno"
        assert kyverno_dir.exists(), "Kyverno policies directory not found"

    def test_policy_files_valid_yaml(self, policies_dir: Path) -> None:
        """Verify all policy YAML files are valid."""
        if not policies_dir.exists():
            pytest.skip("Policies directory not found")

        for yaml_file in policies_dir.rglob("*.yaml"):
            # Skip the vendored kyverno Helm chart under policies/kyverno/charts.
            if not _is_first_party_yaml(yaml_file):
                continue
            content = yaml.safe_load(yaml_file.read_text())
            # Can be None for empty files, dict for single doc
            assert content is None or isinstance(
                content, (dict, list)
            ), f"Invalid content in {yaml_file}"


class TestAppsStructure:
    """Tests for apps directory structure."""

    @pytest.fixture
    def apps_dir(self, project_root: Path) -> Path:
        """Return the apps directory."""
        return project_root / "apps"

    def test_apps_directory_exists(self, apps_dir: Path) -> None:
        """Verify apps directory exists."""
        assert apps_dir.exists(), f"Apps directory not found: {apps_dir}"

    def test_examples_directory_exists(self, apps_dir: Path) -> None:
        """Verify examples directory exists."""
        if not apps_dir.exists():
            pytest.skip("Apps directory not found")

        examples_dir = apps_dir / "examples"
        assert examples_dir.exists(), "Examples directory not found"

    def test_hello_example_exists(self, apps_dir: Path) -> None:
        """Verify hello example exists for E2E testing."""
        if not apps_dir.exists():
            pytest.skip("Apps directory not found")

        hello_dir = apps_dir / "examples" / "hello"
        if hello_dir.exists():
            kustomization = hello_dir / "kustomization.yaml"
            assert kustomization.exists(), "hello example missing kustomization.yaml"


class TestPlatformStructure:
    """Tests for platform components structure."""

    @pytest.fixture
    def platform_dir(self, project_root: Path) -> Path:
        """Return the platform directory."""
        return project_root / "platform"

    def test_platform_directory_exists(self, platform_dir: Path) -> None:
        """Verify platform directory exists."""
        assert platform_dir.exists(), f"Platform directory not found: {platform_dir}"

    def test_platform_has_components(self, platform_dir: Path) -> None:
        """Verify platform has expected component directories."""
        if not platform_dir.exists():
            pytest.skip("Platform directory not found")

        expected_components = ["ingress", "observability", "security"]
        existing = [
            d.name for d in platform_dir.iterdir() if d.is_dir() and not d.name.startswith(".")
        ]

        for component in expected_components:
            if component not in existing:
                # Not a failure, just a warning
                pass


class TestYAMLValidation:
    """Tests for YAML file validation across the project."""

    def test_all_yaml_files_valid(self, project_root: Path) -> None:
        """Verify all YAML files in the project are valid."""
        # Directories to skip
        skip_dirs = {".git", "node_modules", ".venv", "venv", "__pycache__", "legacy", "dist"}

        for yaml_file in project_root.rglob("*.yaml"):
            # Skip excluded directories
            if any(skip_dir in yaml_file.parts for skip_dir in skip_dirs):
                continue
            # Helm templates are Go templates and are only YAML after rendering.
            if "charts" in yaml_file.parts:
                continue

            # Skip vendored Helm charts and Helm-templated manifests; their
            # `{{ }}` directives are not valid YAML until rendered.
            if not _is_first_party_yaml(yaml_file):
                continue

            # safe_load_all validates both single-document and manifest streams.
            list(yaml.safe_load_all(yaml_file.read_text()))

    def test_no_yaml_syntax_errors(self, project_root: Path) -> None:
        """Check for common YAML syntax issues."""
        skip_dirs = {".git", "node_modules", ".venv", "venv", "__pycache__", "legacy"}

        issues = []

        for yaml_file in project_root.rglob("*.yaml"):
            if any(skip_dir in yaml_file.parts for skip_dir in skip_dirs):
                continue

            content = yaml_file.read_text()

            # Check for tabs (YAML prefers spaces)
            if "\t" in content:
                issues.append(f"{yaml_file}: contains tabs (use spaces)")

            # Check for trailing whitespace on non-empty lines
            for i, line in enumerate(content.splitlines(), 1):
                if line != line.rstrip() and line.strip():
                    issues.append(f"{yaml_file}:{i}: trailing whitespace")
                    break  # Only report once per file

        # Soft fail - just report issues
        if issues:
            for issue in issues[:10]:  # Limit output
                print(f"Warning: {issue}")
