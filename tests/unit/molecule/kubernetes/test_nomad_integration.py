"""
Tests for Kubernetes-Nomad integration.

This module uses pytest fixtures to eliminate duplication and follow DRY principles.
"""
from __future__ import annotations

import os
from typing import Any, Callable, Dict, List

import pytest
import testinfra.utils.ansible_runner

testinfra_hosts = testinfra.utils.ansible_runner.AnsibleRunner(
    os.environ["MOLECULE_INVENTORY_FILE"]
).get_hosts("all")


# Fixtures for common test preconditions


@pytest.fixture
def ansible_vars(host) -> Dict[str, Any]:
    """Get Ansible variables for the host."""
    return host.ansible.get_variables()


@pytest.fixture
def host_groups(ansible_vars: Dict[str, Any]) -> List[str]:
    """Get the groups this host belongs to."""
    return ansible_vars.get("group_names", [])


@pytest.fixture
def is_control_node(host_groups: List[str]) -> bool:
    """Check if this is a control node."""
    return "control" in host_groups


def skip_if_not_control(is_control_node: bool) -> None:
    """Skip test if not a control node."""
    if not is_control_node:
        pytest.skip("Not a control node")


def skip_if_feature_disabled(ansible_vars: Dict[str, Any], feature: str) -> None:
    """Skip test if a feature is not enabled."""
    if not ansible_vars.get(feature, False):
        pytest.skip(f"{feature} not enabled")


# Test data


KUBERNETES_DIRECTORIES = [
    "/etc/kubernetes",
    "/var/lib/kubernetes",
    "/var/log/kubernetes",
]

INTEGRATION_DIRECTORIES = [
    "/etc/kubernetes/nomad-integration",
    "/var/lib/kubernetes/nomad-integration",
]

INTEGRATION_FILES = [
    "/etc/kubernetes/nomad-integration/k8s-api-server",
    "/etc/kubernetes/nomad-integration/k8s-ca.crt",
    "/etc/kubernetes/nomad-integration/nomad-token",
    "/etc/nomad.d/kubernetes-integration.hcl",
    "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad",
]

RBAC_FILES = [
    "/etc/kubernetes/nomad-integration/nomad-sa.yaml",
    "/etc/kubernetes/nomad-integration/nomad-clusterrole.yaml",
    "/etc/kubernetes/nomad-integration/nomad-clusterrolebinding.yaml",
]


# Test helper functions


def assert_file_exists_and_is_file(host, file_path: str) -> None:
    """Assert that a file exists and is a regular file."""
    file_obj = host.file(file_path)
    assert file_obj.exists, f"{file_path} does not exist"
    assert file_obj.is_file, f"{file_path} is not a file"


def assert_file_contains(host, file_path: str, content: str) -> None:
    """Assert that a file contains specific content."""
    file_obj = host.file(file_path)
    assert file_obj.exists, f"{file_path} does not exist"
    assert file_obj.contains(content), f"{file_path} does not contain '{content}'"


# Directory tests


@pytest.mark.parametrize("path", KUBERNETES_DIRECTORIES)
def test_kubernetes_directories(host, path: str) -> None:
    """Test that Kubernetes directories exist."""
    dir_obj = host.file(path)
    assert dir_obj.exists
    assert dir_obj.is_directory


@pytest.mark.parametrize("path", INTEGRATION_DIRECTORIES)
def test_integration_directories(host, path: str) -> None:
    """Test that Kubernetes-Nomad integration directories exist."""
    dir_obj = host.file(path)
    assert dir_obj.exists
    assert dir_obj.is_directory


# File existence tests


def test_integration_files(host, is_control_node: bool) -> None:
    """Test that integration files exist on control nodes."""
    skip_if_not_control(is_control_node)

    for file_path in INTEGRATION_FILES:
        assert_file_exists_and_is_file(host, file_path)


def test_rbac_files(host, is_control_node: bool) -> None:
    """Test that RBAC files exist."""
    skip_if_not_control(is_control_node)

    for file_path in RBAC_FILES:
        assert_file_exists_and_is_file(host, file_path)


# Configuration content tests


def test_nomad_k8s_config_content(host, is_control_node: bool) -> None:
    """Test the content of the Nomad Kubernetes integration configuration."""
    skip_if_not_control(is_control_node)

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    required_content = [
        'plugin "kubernetes"',
        "host =",
        "service_account_token =",
        "ca_file =",
    ]

    for content in required_content:
        assert_file_contains(host, config_file, content)


def test_example_job_content(host, is_control_node: bool) -> None:
    """Test the content of the example Nomad job for Kubernetes."""
    skip_if_not_control(is_control_node)

    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"
    required_content = [
        'job "mantl-kubernetes-example"',
        'driver = "kubernetes"',
        "image =",
    ]

    for content in required_content:
        assert_file_contains(host, job_file, content)


# Feature-specific tests


def test_common_criteria_settings(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test Common Criteria settings in the integration."""
    skip_if_not_control(is_control_node)
    skip_if_feature_disabled(ansible_vars, "nomad_common_criteria_enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    required_content = [
        "audit {",
        "enabled = true",
        'type = "file"',
        'path = "/var/log/nomad/kubernetes-audit.json"',
    ]

    for content in required_content:
        assert_file_contains(host, config_file, content)


def test_service_discovery(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test cross-platform service discovery configuration."""
    skip_if_not_control(is_control_node)

    # Default to True if not specified
    if not ansible_vars.get("kubernetes_nomad_service_discovery_enabled", True):
        pytest.skip("Service discovery not enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    required_config = [
        'service_discovery "kubernetes"',
        "server_address =",
        "token =",
        "label_selector =",
    ]

    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"
    required_job = [
        "mantl-service=true",
        "Cross-platform Service Discovery",
    ]

    for content in required_config:
        assert_file_contains(host, config_file, content)

    for content in required_job:
        assert_file_contains(host, job_file, content)


def test_vault_integration(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test Vault integration for secrets management."""
    skip_if_not_control(is_control_node)
    skip_if_feature_disabled(ansible_vars, "kubernetes_nomad_vault_integration_enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    required_config = [
        "vault {",
        "enabled = true",
        "address =",
        "token =",
        "kubernetes_auth {",
    ]

    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"
    required_job = [
        "volumeMounts:",
        "name: vault-token",
        "template {",
        "with secret",
        "Shared Vault Secrets Management",
    ]

    for content in required_config:
        assert_file_contains(host, config_file, content)

    for content in required_job:
        assert_file_contains(host, job_file, content)


def test_federated_metrics(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test federated metrics and logging configuration."""
    skip_if_not_control(is_control_node)
    skip_if_feature_disabled(ansible_vars, "kubernetes_nomad_metrics_enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    required_config = [
        "telemetry {",
        "prometheus_metrics = true",
        "publish_allocation_metrics = true",
        "collection_interval =",
    ]

    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"
    required_job = [
        'task "logging-sidecar"',
        "image = \"{{ kubernetes_nomad_fluentd_image",
        "prometheus.yml",
        "Federated Metrics and Logging",
    ]

    for content in required_config:
        assert_file_contains(host, config_file, content)

    for content in required_job:
        assert_file_contains(host, job_file, content)


def test_multi_region(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test multi-region orchestration configuration."""
    skip_if_not_control(is_control_node)
    skip_if_feature_disabled(ansible_vars, "kubernetes_nomad_multi_region_enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"

    for content in ["multi_region {", "enabled = true", "regions =", "strategy ="]:
        assert_file_contains(host, config_file, content)

    for content in ["multiregion {", "Multi-region Orchestration"]:
        assert_file_contains(host, job_file, content)


def test_gpu_scheduling(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test GPU workload scheduling configuration."""
    skip_if_not_control(is_control_node)
    skip_if_feature_disabled(ansible_vars, "kubernetes_nomad_gpu_enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"

    for content in ["gpu_support = true", "gpu_vendor ="]:
        assert_file_contains(host, config_file, content)

    for content in ["device \"{{ kubernetes_nomad_gpu_vendor", "GPU Workload Scheduling"]:
        assert_file_contains(host, job_file, content)


def test_autoscaling(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test autoscaling integration configuration."""
    skip_if_not_control(is_control_node)
    skip_if_feature_disabled(ansible_vars, "kubernetes_nomad_autoscaling_enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"

    for content in ["autoscaling {", "enabled = true", "min_replicas =", "max_replicas ="]:
        assert_file_contains(host, config_file, content)

    for content in ["scaling {", "min     =", "max     =", "Autoscaling Integration"]:
        assert_file_contains(host, job_file, content)


def test_crd_support(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test custom resource definition support configuration."""
    skip_if_not_control(is_control_node)
    skip_if_feature_disabled(ansible_vars, "kubernetes_nomad_crd_enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"

    for content in ["custom_resources =", "custom_resource_groups ="]:
        assert_file_contains(host, config_file, content)

    for content in [
        "custom_resources = [",
        'apiVersion = "monitoring.coreos.com/v1"',
        "Custom Resource Definition Support",
    ]:
        assert_file_contains(host, job_file, content)


def test_disaster_recovery(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test disaster recovery configuration."""
    skip_if_not_control(is_control_node)
    skip_if_feature_disabled(ansible_vars, "kubernetes_nomad_disaster_recovery_enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"

    for content in [
        "disaster_recovery {",
        "enabled = true",
        "recovery_threshold =",
        "snapshot_path =",
    ]:
        assert_file_contains(host, config_file, content)

    for content in ["disaster_recovery {", "Disaster Recovery with Automatic Failover"]:
        assert_file_contains(host, job_file, content)


def test_opa_enforcement(host, is_control_node: bool, ansible_vars: Dict[str, Any]) -> None:
    """Test unified policy enforcement with OPA configuration."""
    skip_if_not_control(is_control_node)
    skip_if_feature_disabled(ansible_vars, "kubernetes_nomad_opa_enabled")

    config_file = "/etc/nomad.d/kubernetes-integration.hcl"
    job_file = "/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad"

    for content in ["policy {", "enabled = true", "opa_url =", "evaluation_paths ="]:
        assert_file_contains(host, config_file, content)

    for content in [
        'task "policy-agent"',
        "openpolicyagent/opa",
        "package kubernetes.admission",
        "Unified Policy Enforcement with OPA",
    ]:
        assert_file_contains(host, job_file, content)
