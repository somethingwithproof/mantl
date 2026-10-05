"""
Test suite for CiliumNetworkPolicy validation.

This module validates network policies for:
- Valid YAML structure
- Required fields
- Security best practices
- Label consistency
"""

import re

import pytest
import yaml


@pytest.fixture
def network_policies_dir(project_root):
    """Get the network policies directory path."""
    return project_root / "platform" / "network-policies"


@pytest.fixture
def network_policy_files(network_policies_dir):
    """Get all network policy YAML files."""
    if not network_policies_dir.exists():
        pytest.skip("network-policies directory does not exist")
    files = list(network_policies_dir.glob("*-network-policy.yaml"))
    if not files:
        pytest.skip("No network policy files found")
    return files


@pytest.fixture
def loaded_policies(network_policy_files):
    """Load all network policies as YAML documents."""
    policies = []
    for file in network_policy_files:
        content = file.read_text()
        docs = list(yaml.safe_load_all(content))
        for doc in docs:
            if doc:  # Skip empty documents
                doc["_source_file"] = file.name
                policies.append(doc)
    return policies


class TestNetworkPolicyStructure:
    """Test network policy YAML structure."""

    @pytest.mark.network_policies
    @pytest.mark.yaml_validation
    def test_policies_are_valid_yaml(self, network_policy_files):
        """Verify all network policy files are valid YAML."""
        for file in network_policy_files:
            content = file.read_text()
            list(yaml.safe_load_all(content))

    @pytest.mark.network_policies
    @pytest.mark.yaml_validation
    def test_policies_have_required_fields(self, loaded_policies):
        """Verify policies have required Kubernetes fields."""
        required_fields = ["apiVersion", "kind", "metadata", "spec"]

        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            for field in required_fields:
                assert field in policy, f"{source}: Missing required field '{field}'"

    @pytest.mark.network_policies
    @pytest.mark.yaml_validation
    def test_policies_are_cilium_network_policies(self, loaded_policies):
        """Verify policies use CiliumNetworkPolicy kind."""
        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            assert (
                policy.get("apiVersion") == "cilium.io/v2"
            ), f"{source}: Expected apiVersion 'cilium.io/v2'"
            assert (
                policy.get("kind") == "CiliumNetworkPolicy"
            ), f"{source}: Expected kind 'CiliumNetworkPolicy'"


class TestNetworkPolicyMetadata:
    """Test network policy metadata."""

    @pytest.mark.network_policies
    def test_policies_have_names(self, loaded_policies):
        """Verify all policies have names."""
        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            metadata = policy.get("metadata", {})
            assert "name" in metadata, f"{source}: Missing policy name"
            assert metadata["name"], f"{source}: Policy name cannot be empty"

    @pytest.mark.network_policies
    def test_policies_have_namespaces(self, loaded_policies):
        """Verify all policies specify namespaces."""
        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            metadata = policy.get("metadata", {})
            assert (
                "namespace" in metadata
            ), f"{source}: Missing namespace - network policies should be namespace-scoped"

    @pytest.mark.network_policies
    def test_policies_have_labels(self, loaded_policies):
        """Verify all policies have labels for identification."""
        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            metadata = policy.get("metadata", {})
            labels = metadata.get("labels", {})
            assert labels, f"{source}: Should have labels for identification"

    @pytest.mark.network_policies
    def test_policies_have_part_of_label(self, loaded_policies):
        """Verify policies have app.kubernetes.io/part-of label."""
        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            metadata = policy.get("metadata", {})
            labels = metadata.get("labels", {})
            assert (
                "app.kubernetes.io/part-of" in labels
            ), f"{source}: Missing 'app.kubernetes.io/part-of' label"


class TestNetworkPolicySpec:
    """Test network policy specifications."""

    @pytest.mark.network_policies
    def test_policies_have_endpoint_selector(self, loaded_policies):
        """Verify all policies have endpoint selector."""
        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            spec = policy.get("spec", {})
            assert (
                "endpointSelector" in spec
            ), f"{source}: Missing endpointSelector - required to target pods"

    @pytest.mark.network_policies
    def test_policies_have_description(self, loaded_policies):
        """Verify all policies have descriptions."""
        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            name = policy.get("metadata", {}).get("name", "unknown")
            spec = policy.get("spec", {})
            assert (
                "description" in spec
            ), f"{source}/{name}: Should have description explaining policy purpose"

    @pytest.mark.network_policies
    @pytest.mark.security
    def test_egress_policies_allow_dns(self, loaded_policies):
        """Verify egress policies allow DNS resolution."""
        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            name = policy.get("metadata", {}).get("name", "unknown")
            spec = policy.get("spec", {})
            egress = spec.get("egress", [])

            if not egress:
                continue  # No egress rules, skip

            # Check if any egress rule allows DNS (port 53)
            allows_dns = False
            for rule in egress:
                to_ports = rule.get("toPorts", [])
                for port_spec in to_ports:
                    ports = port_spec.get("ports", [])
                    for port in ports:
                        if port.get("port") == "53":
                            allows_dns = True
                            break

            assert (
                allows_dns
            ), f"{source}/{name}: Egress policy should allow DNS (port 53) for name resolution"


class TestNetworkPolicyConsistency:
    """Test network policy consistency across files."""

    @pytest.mark.network_policies
    def test_policy_names_are_unique(self, loaded_policies):
        """Verify all policy names are unique within their namespace."""
        seen = {}
        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            name = policy.get("metadata", {}).get("name", "")
            namespace = policy.get("metadata", {}).get("namespace", "default")
            key = f"{namespace}/{name}"

            if key in seen:
                pytest.fail(
                    f"Duplicate policy name '{name}' in namespace '{namespace}' "
                    f"found in {source} and {seen[key]}"
                )
            seen[key] = source

    @pytest.mark.network_policies
    def test_label_selectors_are_valid(self, loaded_policies):
        """Verify label selectors use valid patterns."""
        invalid_chars = re.compile(r"[^a-zA-Z0-9._/-]")

        for policy in loaded_policies:
            source = policy.get("_source_file", "unknown")
            name = policy.get("metadata", {}).get("name", "unknown")
            spec = policy.get("spec", {})

            # Check endpoint selector
            endpoint_selector = spec.get("endpointSelector", {})
            match_labels = endpoint_selector.get("matchLabels", {})

            for key, value in match_labels.items():
                if invalid_chars.search(str(key)) or invalid_chars.search(str(value)):
                    pytest.fail(
                        f"{source}/{name}: Invalid characters in label selector " f"'{key}={value}'"
                    )


class TestNetworkPolicyKustomization:
    """Test network policy kustomization setup."""

    @pytest.mark.network_policies
    def test_kustomization_exists(self, network_policies_dir):
        """Verify kustomization.yaml exists."""
        kustomization = network_policies_dir / "kustomization.yaml"
        assert kustomization.exists(), "network-policies directory should have kustomization.yaml"

    @pytest.mark.network_policies
    def test_kustomization_includes_all_policies(self, network_policies_dir, network_policy_files):
        """Verify kustomization.yaml includes all policy files."""
        kustomization = network_policies_dir / "kustomization.yaml"
        if not kustomization.exists():
            pytest.skip("kustomization.yaml not found")

        content = kustomization.read_text()
        kustomization_data = yaml.safe_load(content)
        resources = kustomization_data.get("resources", [])

        policy_filenames = {f.name for f in network_policy_files}

        for filename in policy_filenames:
            assert (
                filename in resources
            ), f"kustomization.yaml should include '{filename}' in resources"
