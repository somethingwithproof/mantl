"""
End-to-end tests for Terraform provider validation.

Tests that all cloud provider blueprints can be planned and validated successfully.
"""
import os
import subprocess
import pytest
from pathlib import Path

# Path to terraform blueprints
BLUEPRINT_DIR = Path(__file__).parent.parent.parent.parent / "infra" / "terraform" / "terraform" / "blueprints"

PROVIDERS = [
    ("aws-eks", "AWS EKS"),
    ("gcp-gke", "GCP GKE"),
    ("azure-aks", "Azure AKS"),
    ("do-doks", "DigitalOcean DOKS"),
    ("linode-lke", "Linode LKE"),
    ("oci-oke", "Oracle OKE"),
    ("ibm-iks", "IBM IKS"),
    ("openstack", "OpenStack"),
]


class TestProviderValidation:
    """Test Terraform provider blueprints for validity."""

    @pytest.mark.parametrize("provider_dir,provider_name", PROVIDERS)
    def test_terraform_init(self, provider_dir, provider_name):
        """Test that terraform init succeeds for each provider."""
        blueprint_path = BLUEPRINT_DIR / provider_dir

        assert blueprint_path.exists(), f"Blueprint directory not found: {blueprint_path}"

        # Run terraform init
        result = subprocess.run(
            ["terraform", "init", "-backend=false"],
            cwd=blueprint_path,
            capture_output=True,
            text=True
        )

        assert result.returncode == 0, (
            f"Terraform init failed for {provider_name}:\n"
            f"STDOUT: {result.stdout}\n"
            f"STDERR: {result.stderr}"
        )

    @pytest.mark.parametrize("provider_dir,provider_name", PROVIDERS)
    def test_terraform_validate(self, provider_dir, provider_name):
        """Test that terraform validate succeeds for each provider."""
        blueprint_path = BLUEPRINT_DIR / provider_dir

        # Initialize first
        subprocess.run(
            ["terraform", "init", "-backend=false"],
            cwd=blueprint_path,
            capture_output=True
        )

        # Run terraform validate
        result = subprocess.run(
            ["terraform", "validate"],
            cwd=blueprint_path,
            capture_output=True,
            text=True
        )

        assert result.returncode == 0, (
            f"Terraform validate failed for {provider_name}:\n"
            f"STDOUT: {result.stdout}\n"
            f"STDERR: {result.stderr}"
        )

    @pytest.mark.parametrize("provider_dir,provider_name", PROVIDERS)
    def test_terraform_fmt_check(self, provider_dir, provider_name):
        """Test that Terraform files are properly formatted."""
        blueprint_path = BLUEPRINT_DIR / provider_dir

        result = subprocess.run(
            ["terraform", "fmt", "-check", "-recursive"],
            cwd=blueprint_path,
            capture_output=True,
            text=True
        )

        assert result.returncode == 0, (
            f"Terraform files not properly formatted for {provider_name}:\n"
            f"Run 'terraform fmt -recursive' in {provider_dir}"
        )

    @pytest.mark.parametrize("provider_dir,provider_name", PROVIDERS)
    def test_required_files_exist(self, provider_dir, provider_name):
        """Test that all required files exist in each blueprint."""
        blueprint_path = BLUEPRINT_DIR / provider_dir

        required_files = ["main.tf", "variables.tf"]

        for file in required_files:
            file_path = blueprint_path / file
            assert file_path.exists(), (
                f"Required file {file} not found in {provider_name} blueprint"
            )

    @pytest.mark.parametrize("provider_dir,provider_name", PROVIDERS)
    def test_kubernetes_version_latest(self, provider_dir, provider_name):
        """Test that Kubernetes version is set to latest (1.31.x)."""
        variables_path = BLUEPRINT_DIR / provider_dir / "variables.tf"

        with open(variables_path, 'r') as f:
            content = f.read()

        # Check that kubernetes_version variable exists and defaults to 1.31.x
        assert 'variable "kubernetes_version"' in content, (
            f"kubernetes_version variable not found in {provider_name}"
        )

        # Check for 1.31 version
        assert "1.31" in content, (
            f"Kubernetes version not set to 1.31.x in {provider_name}"
        )

    @pytest.mark.parametrize("provider_dir,provider_name", PROVIDERS)
    def test_security_features_configured(self, provider_dir, provider_name):
        """Test that security features are configured in each provider."""
        main_tf_path = BLUEPRINT_DIR / provider_dir / "main.tf"

        with open(main_tf_path, 'r') as f:
            content = f.read()

        # Check for network security features (varies by provider)
        security_indicators = [
            "security",  # Security groups/lists
            "firewall",  # Firewall rules
            "private",   # Private networking
        ]

        has_security = any(indicator in content.lower() for indicator in security_indicators)

        assert has_security, (
            f"No security features found in {provider_name} blueprint"
        )


class TestProviderSpecificFeatures:
    """Test provider-specific security features."""

    def test_aws_vpc_flow_logs(self):
        """Test that AWS EKS has VPC Flow Logs configured."""
        main_tf_path = BLUEPRINT_DIR / "aws-eks" / "main.tf"

        with open(main_tf_path, 'r') as f:
            content = f.read()

        assert "aws_flow_log" in content, (
            "VPC Flow Logs not configured for AWS EKS"
        )

    def test_gcp_vpc_flow_logs(self):
        """Test that GCP GKE has VPC Flow Logs configured."""
        main_tf_path = BLUEPRINT_DIR / "gcp-gke" / "main.tf"

        with open(main_tf_path, 'r') as f:
            content = f.read()

        assert "enable_flow_logs" in content, (
            "VPC Flow Logs not configured for GCP GKE"
        )

    def test_azure_network_watcher(self):
        """Test that Azure AKS has Network Watcher Flow Logs."""
        main_tf_path = BLUEPRINT_DIR / "azure-aks" / "main.tf"

        with open(main_tf_path, 'r') as f:
            content = f.read()

        assert "azurerm_network_watcher_flow_log" in content, (
            "Network Watcher Flow Logs not configured for Azure AKS"
        )

    def test_oracle_vault_configured(self):
        """Test that Oracle OKE has OCI Vault configured."""
        main_tf_path = BLUEPRINT_DIR / "oci-oke" / "main.tf"

        with open(main_tf_path, 'r') as f:
            content = f.read()

        assert "oci_kms_vault" in content, (
            "OCI Vault not configured for Oracle OKE"
        )

    def test_ibm_key_protect(self):
        """Test that IBM IKS has Key Protect configured."""
        main_tf_path = BLUEPRINT_DIR / "ibm-iks" / "main.tf"

        with open(main_tf_path, 'r') as f:
            content = f.read()

        assert "ibm_resource_instance" in content and "kms" in content, (
            "IBM Key Protect not configured for IBM IKS"
        )

    def test_all_providers_have_backup_storage(self):
        """Test that all providers have backup storage configured."""
        for provider_dir, provider_name in PROVIDERS:
            main_tf_path = BLUEPRINT_DIR / provider_dir / "main.tf"

            with open(main_tf_path, 'r') as f:
                content = f.read()

            backup_indicators = ["backup", "bucket", "storage"]
            has_backup = any(indicator in content.lower() for indicator in backup_indicators)

            assert has_backup, (
                f"Backup storage not configured for {provider_name}"
            )
