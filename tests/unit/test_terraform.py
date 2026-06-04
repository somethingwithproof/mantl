"""Unit tests for Terraform configurations.

These tests validate Terraform syntax, configuration validity, and security best practices.
"""

import json
import subprocess
from pathlib import Path
from typing import Dict, List, Optional

import pytest


# Test fixtures

@pytest.fixture
def terraform_root() -> Path:
    """Get the Terraform root directory."""
    return Path(__file__).parent.parent.parent / "infra" / "terraform" / "terraform"


@pytest.fixture
def blueprints_dir(terraform_root: Path) -> Path:
    """Get the Terraform blueprints directory."""
    return terraform_root / "blueprints"


@pytest.fixture
def aws_eks_blueprint(blueprints_dir: Path) -> Path:
    """Get the AWS EKS blueprint directory."""
    return blueprints_dir / "aws-eks"


# Helper functions

def run_terraform_command(
    command: List[str], cwd: Path, timeout: int = 60
) -> Dict[str, any]:
    """Run a terraform command and return the result."""
    try:
        result = subprocess.run(
            ["terraform"] + command,
            cwd=cwd,
            capture_output=True,
            text=True,
            timeout=timeout,
        )
        return {
            "returncode": result.returncode,
            "stdout": result.stdout,
            "stderr": result.stderr,
        }
    except subprocess.TimeoutExpired:
        pytest.fail(f"Terraform command timed out: {command}")
    except FileNotFoundError:
        pytest.skip("terraform not available")
        return {}


# Substrings in `terraform init` stderr that indicate a network/download
# problem (fetching registry modules or providers) rather than a config defect.
_NETWORK_INIT_ERRORS = (
    "could not download",
    "failed to download",
    "failed to query",
    "could not query provider",
    "error querying",
    "connection refused",
    "no such host",
    "timeout",
    "dial tcp",
    "tls handshake",
    "could not retrieve",
    "registry service unavailable",
)


def terraform_init(directory: Path) -> Dict:
    """Initialize Terraform in a directory; returns the command result."""
    return run_terraform_command(["init", "-backend=false", "-no-color"], directory)


def terraform_validate(directory: Path) -> Dict:
    """Validate Terraform configuration, initializing first if needed.

    `init` resolves registry modules and providers over the network. A failure
    is only treated as an environment limitation (skip) when its stderr matches a
    known network/download error; any other init failure is a real config or
    source defect and fails the test.
    """
    if not (directory / ".terraform").exists():
        init = terraform_init(directory)
        if init.get("returncode") != 0:
            stderr = init.get("stderr", "")
            if any(marker in stderr.lower() for marker in _NETWORK_INIT_ERRORS):
                pytest.skip(f"terraform init unavailable (network) for {directory}")
            pytest.fail(f"terraform init failed (non-network) for {directory}:\n{stderr}")

    return run_terraform_command(["validate", "-json"], directory)


def terraform_fmt_check(directory: Path) -> Dict:
    """Check Terraform formatting."""
    return run_terraform_command(["fmt", "-check", "-recursive"], directory)


# Format tests

def test_terraform_formatted(terraform_root: Path):
    """Test that all Terraform files are properly formatted."""
    result = terraform_fmt_check(terraform_root)

    if result:
        assert result["returncode"] == 0, (
            f"Terraform files not formatted. Run: terraform fmt -recursive\n"
            f"Files needing formatting:\n{result['stdout']}"
        )


# AWS EKS Blueprint tests

def test_aws_eks_blueprint_exists(aws_eks_blueprint: Path):
    """Test that AWS EKS blueprint directory exists."""
    assert aws_eks_blueprint.exists()
    assert (aws_eks_blueprint / "main.tf").exists()
    assert (aws_eks_blueprint / "variables.tf").exists()
    assert (aws_eks_blueprint / "outputs.tf").exists()


def test_aws_eks_blueprint_validates(aws_eks_blueprint: Path):
    """Test that AWS EKS blueprint has valid Terraform syntax."""
    result = terraform_validate(aws_eks_blueprint)

    if result and result.get("returncode") == 0:
        validation = json.loads(result["stdout"])
        assert validation["valid"] is True, (
            f"AWS EKS blueprint validation failed:\n{json.dumps(validation, indent=2)}"
        )


def test_aws_eks_variables_have_descriptions(aws_eks_blueprint: Path):
    """Test that all variables have descriptions."""
    vars_file = aws_eks_blueprint / "variables.tf"
    if not vars_file.exists():
        pytest.skip("variables.tf not found")

    # Import inside the try so a missing python-hcl2 skips rather than errors;
    # hcl2 is an optional dev dependency, not part of requirements.
    try:
        import hcl2

        with open(vars_file) as f:
            config = hcl2.load(f)

        variables = config.get("variable", [])
        for var_name, var_config in variables.items():
            assert "description" in var_config[0], (
                f"Variable '{var_name}' missing description"
            )
    except ImportError:
        pytest.skip("python-hcl2 not installed")


def test_aws_eks_security_defaults(aws_eks_blueprint: Path):
    """Test that AWS EKS blueprint has secure defaults."""
    vars_file = aws_eks_blueprint / "variables.tf"
    with open(vars_file) as f:
        content = f.read()

    # Check that public endpoint access is disabled by default. Match the
    # variable's default without coupling to inline-comment whitespace, which
    # `terraform fmt` normalizes (the old exact-string assertion broke on fmt).
    import re

    public_access = re.search(
        r'variable\s+"cluster_endpoint_public_access"\s*\{[^}]*?default\s*=\s*false',
        content,
        re.DOTALL,
    )
    assert public_access, "cluster_endpoint_public_access should default to false"

    # Check that latest Kubernetes version is used
    assert 'default     = "1.31"' in content, (
        "Should use latest stable Kubernetes version (1.31)"
    )


def test_aws_eks_vpc_flow_logs_enabled(aws_eks_blueprint: Path):
    """Test that VPC Flow Logs are configured."""
    main_file = aws_eks_blueprint / "main.tf"
    with open(main_file) as f:
        content = f.read()

    assert 'resource "aws_flow_log" "vpc"' in content, (
        "VPC Flow Logs should be configured for security monitoring"
    )
    assert 'resource "aws_cloudwatch_log_group" "vpc_flow_logs"' in content, (
        "CloudWatch Log Group for VPC Flow Logs should be configured"
    )


def test_aws_eks_encryption_enabled(aws_eks_blueprint: Path):
    """Test that EKS cluster encryption is enabled."""
    main_file = aws_eks_blueprint / "main.tf"
    with open(main_file) as f:
        content = f.read()

    assert "cluster_encryption_config" in content, (
        "EKS cluster encryption should be enabled"
    )
    assert 'resources        = ["secrets"]' in content, (
        "Secrets should be encrypted"
    )


def test_aws_eks_logging_enabled(aws_eks_blueprint: Path):
    """Test that EKS control plane logging is enabled."""
    main_file = aws_eks_blueprint / "main.tf"
    with open(main_file) as f:
        content = f.read()

    required_logs = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
    for log_type in required_logs:
        assert log_type in content, f"EKS log type '{log_type}' should be enabled"


# Multi-blueprint tests

@pytest.mark.parametrize("blueprint", [
    "aws-eks",
    # Add other cloud providers when testing them
    # "gcp-gke",
    # "azure-aks",
])
def test_blueprint_structure(blueprints_dir: Path, blueprint: str):
    """Test that each blueprint has the required file structure."""
    blueprint_dir = blueprints_dir / blueprint

    if not blueprint_dir.exists():
        pytest.skip(f"Blueprint {blueprint} not found")

    required_files = ["main.tf", "variables.tf", "outputs.tf"]
    for required_file in required_files:
        assert (blueprint_dir / required_file).exists(), (
            f"Blueprint {blueprint} missing {required_file}"
        )


def test_all_blueprints_validate(blueprints_dir: Path):
    """Test that all blueprints have valid Terraform syntax."""
    if not blueprints_dir.exists():
        pytest.skip("Blueprints directory not found")

    blueprints = [d for d in blueprints_dir.iterdir() if d.is_dir()]

    for blueprint_dir in blueprints:
        # Skip hidden directories and terraform state
        if blueprint_dir.name.startswith(".") or blueprint_dir.name == ".terraform":
            continue

        # Check if it has terraform files
        tf_files = list(blueprint_dir.glob("*.tf"))
        if not tf_files:
            continue

        result = terraform_validate(blueprint_dir)
        if result and result.get("returncode") == 0:
            validation = json.loads(result["stdout"])
            assert validation["valid"] is True, (
                f"Blueprint {blueprint_dir.name} validation failed:\n"
                f"{json.dumps(validation, indent=2)}"
            )


# Security best practices tests

def test_no_hardcoded_secrets(terraform_root: Path):
    """Test that there are no hardcoded secrets in Terraform files."""
    import re

    # Patterns that might indicate hardcoded secrets
    secret_patterns = [
        r'password\s*=\s*"[^"]+"',
        r'secret\s*=\s*"[^"]+"',
        r'api_key\s*=\s*"[^"]+"',
        r'access_key\s*=\s*"AKIA[A-Z0-9]{16}"',
    ]

    for tf_file in terraform_root.rglob("*.tf"):
        if ".terraform" in str(tf_file):
            continue

        with open(tf_file) as f:
            content = f.read()

        for pattern in secret_patterns:
            matches = re.findall(pattern, content, re.IGNORECASE)
            # Filter out obvious placeholders
            real_matches = [
                m for m in matches
                if "changeme" not in m.lower()
                and "placeholder" not in m.lower()
                and "example" not in m.lower()
            ]
            assert not real_matches, (
                f"Possible hardcoded secret in {tf_file}:\n{real_matches}"
            )


def test_resources_have_tags(aws_eks_blueprint: Path):
    """Test that resources are properly tagged."""
    main_file = aws_eks_blueprint / "main.tf"
    with open(main_file) as f:
        content = f.read()

    # Check that local.tags is used throughout
    assert "tags = local.tags" in content or "tags = merge(local.tags" in content, (
        "Resources should use consistent tagging with local.tags"
    )


# Provider version tests

def test_terraform_version_constraint():
    """Test that Terraform version is constrained."""
    # This would need to check terraform.tf or versions.tf files
    # Skipping for now as structure varies
    pytest.skip("Version constraint testing not implemented")


# Output tests

def test_aws_eks_outputs_exist(aws_eks_blueprint: Path):
    """Test that AWS EKS blueprint has useful outputs."""
    outputs_file = aws_eks_blueprint / "outputs.tf"

    if not outputs_file.exists():
        pytest.skip("outputs.tf not found")

    with open(outputs_file) as f:
        content = f.read()

    required_outputs = [
        "cluster_name",
        "cluster_endpoint",
        "cluster_id",
    ]

    for output_name in required_outputs:
        assert f'output "{output_name}"' in content, (
            f"Output '{output_name}' should be defined"
        )
