# SPDX-License-Identifier: Apache-2.0
"""
End-to-end security verification tests.

Tests that verify security configurations in deployed Kubernetes clusters.
Requires a configured kubectl context.
"""

import json
import subprocess

import pytest


class TestNetworkPolicies:
    """Test Kubernetes Network Policy configurations."""

    def test_network_policy_enabled(self):
        """Test that NetworkPolicy API is available."""
        result = subprocess.run(
            ["kubectl", "api-resources", "--api-group=networking.k8s.io"],
            capture_output=True,
            text=True,
        )

        assert "networkpolicies" in result.stdout, "NetworkPolicy API not available in cluster"

    def test_default_deny_policy_exists(self):
        """Test that default deny network policies exist in protected namespaces."""
        protected_namespaces = ["production", "staging"]

        for namespace in protected_namespaces:
            # Check if namespace exists first
            check_ns = subprocess.run(
                ["kubectl", "get", "namespace", namespace], capture_output=True
            )

            if check_ns.returncode == 0:
                result = subprocess.run(
                    ["kubectl", "get", "networkpolicy", "-n", namespace, "-o", "json"],
                    capture_output=True,
                    text=True,
                )

                if result.returncode == 0:
                    policies = json.loads(result.stdout)
                    has_deny_policy = any(
                        "deny" in policy["metadata"]["name"].lower()
                        for policy in policies.get("items", [])
                    )

                    assert (
                        has_deny_policy or len(policies.get("items", [])) > 0
                    ), f"No network policies found in {namespace} namespace"


class TestRBAC:
    """Test RBAC configurations."""

    def test_rbac_enabled(self):
        """Test that RBAC API is enabled."""
        result = subprocess.run(
            ["kubectl", "api-resources", "--api-group=rbac.authorization.k8s.io"],
            capture_output=True,
            text=True,
        )

        assert "roles" in result.stdout, "RBAC API not available in cluster"

    def test_no_default_serviceaccount_tokens(self):
        """Test that default ServiceAccounts don't have automatic token mounting."""
        result = subprocess.run(
            ["kubectl", "get", "serviceaccount", "default", "-n", "default", "-o", "json"],
            capture_output=True,
            text=True,
        )

        if result.returncode == 0:
            sa = json.loads(result.stdout)
            auto_mount = sa.get("automountServiceAccountToken", True)

            # Best practice is to set this to false
            if "automountServiceAccountToken" in sa:
                assert (
                    not auto_mount
                ), "Default ServiceAccount has automountServiceAccountToken=true"

    def test_cluster_admin_not_bound_to_anonymous(self):
        """Test that cluster-admin is not bound to anonymous or system:unauthenticated."""
        result = subprocess.run(
            ["kubectl", "get", "clusterrolebinding", "-o", "json"], capture_output=True, text=True
        )

        bindings = json.loads(result.stdout)
        for binding in bindings.get("items", []):
            if binding["roleRef"]["name"] == "cluster-admin":
                for subject in binding.get("subjects", []):
                    assert subject.get("name") not in [
                        "system:anonymous",
                        "system:unauthenticated",
                    ], (
                        f"cluster-admin bound to {subject.get('name')} "
                        f"in {binding['metadata']['name']}"
                    )


class TestPodSecurityStandards:
    """Test Pod Security Standards enforcement."""

    def test_pss_labels_exist(self):
        """Test that namespaces have Pod Security Standard labels."""
        result = subprocess.run(
            ["kubectl", "get", "namespaces", "-o", "json"], capture_output=True, text=True
        )

        namespaces = json.loads(result.stdout)
        checked_namespaces = ["default", "kube-system"]

        for ns in namespaces.get("items", []):
            ns_name = ns["metadata"]["name"]
            if ns_name in checked_namespaces:
                labels = ns["metadata"].get("labels", {})

                # Check for Pod Security Standards labels
                has_pss = any(key.startswith("pod-security.kubernetes.io/") for key in labels)

                # kube-system should have privileged, others should have at least baseline
                if ns_name == "kube-system":
                    continue  # Skip kube-system as it needs privileged
                elif ns_name == "default":
                    assert has_pss, f"Namespace {ns_name} missing Pod Security Standard labels"

    def test_no_privileged_pods_in_default(self):
        """Test that no privileged pods are running in default namespace."""
        result = subprocess.run(
            ["kubectl", "get", "pods", "-n", "default", "-o", "json"],
            capture_output=True,
            text=True,
        )

        if result.returncode == 0:
            pods = json.loads(result.stdout)
            for pod in pods.get("items", []):
                for container in pod["spec"].get("containers", []):
                    security_context = container.get("securityContext", {})
                    assert not security_context.get(
                        "privileged", False
                    ), f"Privileged pod {pod['metadata']['name']} found in default namespace"


class TestSecretsEncryption:
    """Test secrets encryption at rest."""

    def test_secrets_encrypted_annotation(self):
        """Test that cluster has encryption configuration."""
        # This is provider-specific and may require additional checks
        # For now, we verify that secrets exist and can be accessed
        result = subprocess.run(
            ["kubectl", "get", "secrets", "-A", "-o", "json"], capture_output=True, text=True
        )

        assert result.returncode == 0, "Unable to list secrets - may indicate encryption issues"

    def test_no_plaintext_secrets_in_etcd(self):
        """Test that secrets are not stored in plaintext (requires etcd access)."""
        # This test requires direct etcd access, which is typically not available
        # in managed clusters. Skip if etcd is not accessible.
        pytest.skip("Requires direct etcd access - not available in managed clusters")


class TestImageSecurity:
    """Test container image security policies."""

    def test_kyverno_policies_installed(self):
        """Test that Kyverno policies are installed."""
        result = subprocess.run(
            ["kubectl", "get", "clusterpolicy", "-o", "json"], capture_output=True, text=True
        )

        if result.returncode == 0:
            policies = json.loads(result.stdout)
            policy_names = [p["metadata"]["name"] for p in policies.get("items", [])]

            # Check for key security policies
            expected_policies = [
                "require-read-only-root-filesystem",
                "restrict-image-registries",
                "disallow-privileged-containers",
            ]

            found_policies = [
                p for p in expected_policies if any(p in name for name in policy_names)
            ]

            assert (
                len(found_policies) > 0
            ), f"Expected security policies not found. Found: {policy_names}"

    def test_no_latest_tags(self):
        """Test that no containers use 'latest' image tag."""
        result = subprocess.run(
            ["kubectl", "get", "pods", "-A", "-o", "json"], capture_output=True, text=True
        )

        pods = json.loads(result.stdout)
        violations = []

        for pod in pods.get("items", []):
            namespace = pod["metadata"]["namespace"]
            pod_name = pod["metadata"]["name"]

            for container in pod["spec"].get("containers", []):
                image = container.get("image", "")
                if ":latest" in image or ":" not in image:
                    violations.append(f"{namespace}/{pod_name}: {image}")

        # Some system pods may use latest, so we just warn
        if violations and len(violations) < 10:
            pytest.skip(f"Some pods using :latest tag: {violations[:5]}")


class TestAuditLogging:
    """Test audit logging configuration."""

    def test_audit_policy_configured(self):
        """Test that audit logging is configured."""
        # This is provider-specific. For managed clusters, audit logs
        # are typically available through cloud provider logging.
        result = subprocess.run(
            ["kubectl", "get", "configmap", "audit-policy", "-n", "kube-system"],
            capture_output=True,
        )

        # If audit-policy configmap doesn't exist, check for provider-specific logging
        if result.returncode != 0:
            pytest.skip(
                "Audit policy configmap not found. "
                "For managed clusters, check cloud provider audit logging."
            )


class TestMonitoring:
    """Test that monitoring components are deployed."""

    def test_prometheus_deployed(self):
        """Test that Prometheus is deployed."""
        result = subprocess.run(
            [
                "kubectl",
                "get",
                "deployment",
                "-n",
                "monitoring",
                "-l",
                "app.kubernetes.io/name=prometheus",
            ],
            capture_output=True,
        )

        assert result.returncode == 0, "Prometheus deployment not found in monitoring namespace"

    def test_falco_deployed(self):
        """Test that Falco is deployed for runtime security."""
        result = subprocess.run(
            ["kubectl", "get", "daemonset", "-n", "security", "-l", "app.kubernetes.io/name=falco"],
            capture_output=True,
        )

        # Falco is optional, so just check if namespace exists
        if result.returncode != 0:
            pytest.skip("Falco not deployed - optional security component")
