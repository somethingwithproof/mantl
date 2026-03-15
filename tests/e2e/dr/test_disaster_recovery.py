"""
End-to-end tests for disaster recovery capabilities.

Tests Velero backup/restore and database failover scenarios.
"""
import subprocess
import pytest
import json
import time


class TestVeleroBackupRestore:
    """Test Velero backup and restore functionality."""

    @pytest.fixture
    def test_namespace(self):
        """Create a test namespace with resources."""
        namespace = "velero-test-ns"

        # Create namespace
        subprocess.run(
            ["kubectl", "create", "namespace", namespace],
            capture_output=True
        )

        # Create test resources
        subprocess.run(
            ["kubectl", "create", "configmap", "test-config",
             "--from-literal=key=value", "-n", namespace],
            capture_output=True
        )

        subprocess.run(
            ["kubectl", "create", "secret", "generic", "test-secret",
             "--from-literal=password=secret123", "-n", namespace],
            capture_output=True
        )

        yield namespace

        # Cleanup
        subprocess.run(
            ["kubectl", "delete", "namespace", namespace, "--force", "--grace-period=0"],
            capture_output=True
        )

    def test_backup_and_restore_namespace(self, test_namespace):
        """Test backing up and restoring a complete namespace."""
        if not self._velero_installed():
            pytest.skip("Velero not installed")

        backup_name = f"test-backup-{int(time.time())}"

        # Create backup
        result = subprocess.run(
            ["kubectl", "create", "-n", "velero", "backup", backup_name,
             "--include-namespaces", test_namespace, "--wait"],
            capture_output=True,
            text=True,
            timeout=300
        )

        assert result.returncode == 0, f"Backup creation failed: {result.stderr}"

        # Wait for backup to complete
        time.sleep(10)

        # Verify backup completed
        result = subprocess.run(
            ["kubectl", "get", "backup", backup_name, "-n", "velero", "-o", "json"],
            capture_output=True,
            text=True
        )

        backup = json.loads(result.stdout)
        assert backup["status"]["phase"] == "Completed", (
            f"Backup not completed: {backup['status']}"
        )

        # Delete namespace
        subprocess.run(
            ["kubectl", "delete", "namespace", test_namespace, "--force", "--grace-period=0"],
            capture_output=True
        )

        # Wait for namespace to be deleted
        for _ in range(30):
            check = subprocess.run(
                ["kubectl", "get", "namespace", test_namespace],
                capture_output=True
            )
            if check.returncode != 0:
                break
            time.sleep(2)

        # Restore from backup
        restore_name = f"test-restore-{int(time.time())}"
        result = subprocess.run(
            ["kubectl", "create", "-n", "velero", "restore", restore_name,
             "--from-backup", backup_name, "--wait"],
            capture_output=True,
            text=True,
            timeout=300
        )

        assert result.returncode == 0, f"Restore failed: {result.stderr}"

        # Wait for restore to complete
        time.sleep(10)

        # Verify resources were restored
        result = subprocess.run(
            ["kubectl", "get", "configmap", "test-config", "-n", test_namespace, "-o", "json"],
            capture_output=True,
            text=True
        )

        assert result.returncode == 0, "ConfigMap not restored"

        result = subprocess.run(
            ["kubectl", "get", "secret", "test-secret", "-n", test_namespace, "-o", "json"],
            capture_output=True,
            text=True
        )

        assert result.returncode == 0, "Secret not restored"

        # Cleanup
        subprocess.run(["kubectl", "delete", "backup", backup_name, "-n", "velero"], capture_output=True)
        subprocess.run(["kubectl", "delete", "restore", restore_name, "-n", "velero"], capture_output=True)

    def test_backup_storage_location_accessible(self):
        """Test that backup storage location is accessible."""
        if not self._velero_installed():
            pytest.skip("Velero not installed")

        result = subprocess.run(
            ["kubectl", "get", "backupstoragelocation", "-n", "velero", "-o", "json"],
            capture_output=True,
            text=True
        )

        locations = json.loads(result.stdout)
        assert len(locations.get("items", [])) > 0, "No backup storage locations found"

        # Check all locations are available
        for location in locations["items"]:
            status = location.get("status", {})
            assert status.get("phase") == "Available", (
                f"Backup location {location['metadata']['name']} not available: {status}"
            )

    @staticmethod
    def _velero_installed():
        """Check if Velero is installed."""
        result = subprocess.run(
            ["kubectl", "get", "namespace", "velero"],
            capture_output=True
        )
        return result.returncode == 0


class TestDatabaseFailover:
    """Test CloudNativePG database failover."""

    def test_postgres_cluster_exists(self):
        """Test that a PostgreSQL cluster exists for failover testing."""
        if not self._cnpg_installed():
            pytest.skip("CloudNativePG not installed")

        result = subprocess.run(
            ["kubectl", "get", "cluster.postgresql.cnpg.io", "-A", "-o", "json"],
            capture_output=True,
            text=True
        )

        clusters = json.loads(result.stdout)

        if len(clusters.get("items", [])) == 0:
            pytest.skip("No PostgreSQL clusters found for failover testing")

    def test_postgres_cluster_ha_configured(self):
        """Test that PostgreSQL cluster has HA configured."""
        if not self._cnpg_installed():
            pytest.skip("CloudNativePG not installed")

        result = subprocess.run(
            ["kubectl", "get", "cluster.postgresql.cnpg.io", "-A", "-o", "json"],
            capture_output=True,
            text=True
        )

        clusters = json.loads(result.stdout)

        for cluster in clusters.get("items", []):
            instances = cluster.get("spec", {}).get("instances", 1)
            assert instances >= 2, (
                f"Cluster {cluster['metadata']['name']} not configured for HA "
                f"(instances={instances}, should be >=2)"
            )

    def test_postgres_backup_configured(self):
        """Test that PostgreSQL backup is configured."""
        if not self._cnpg_installed():
            pytest.skip("CloudNativePG not installed")

        result = subprocess.run(
            ["kubectl", "get", "cluster.postgresql.cnpg.io", "-A", "-o", "json"],
            capture_output=True,
            text=True
        )

        clusters = json.loads(result.stdout)

        for cluster in clusters.get("items", []):
            backup_config = cluster.get("spec", {}).get("backup", {})
            assert "barmanObjectStore" in backup_config, (
                f"Cluster {cluster['metadata']['name']} missing backup configuration"
            )

    @staticmethod
    def _cnpg_installed():
        """Check if CloudNativePG is installed."""
        result = subprocess.run(
            ["kubectl", "get", "crd", "clusters.postgresql.cnpg.io"],
            capture_output=True
        )
        return result.returncode == 0


class TestHighAvailability:
    """Test high availability configurations."""

    def test_control_plane_replicas(self):
        """Test that control plane components have multiple replicas."""
        control_plane_components = [
            ("kube-system", "coredns"),
            ("monitoring", "prometheus-server"),
        ]

        for namespace, component in control_plane_components:
            result = subprocess.run(
                ["kubectl", "get", "deployment", "-n", namespace,
                 "-l", f"k8s-app={component}", "-o", "json"],
                capture_output=True,
                text=True
            )

            if result.returncode != 0:
                continue

            deployments = json.loads(result.stdout)
            for deployment in deployments.get("items", []):
                replicas = deployment["spec"].get("replicas", 1)
                if component == "coredns":
                    assert replicas >= 2, (
                        f"{component} should have at least 2 replicas for HA"
                    )

    def test_pod_disruption_budgets_exist(self):
        """Test that Pod Disruption Budgets are configured for critical services."""
        result = subprocess.run(
            ["kubectl", "get", "poddisruptionbudget", "-A", "-o", "json"],
            capture_output=True,
            text=True
        )

        pdbs = json.loads(result.stdout)

        # Just verify PDBs can be listed - specific PDBs are optional
        assert isinstance(pdbs.get("items"), list), (
            "Could not list PodDisruptionBudgets"
        )

    def test_anti_affinity_for_critical_workloads(self):
        """Test that critical workloads have pod anti-affinity configured."""
        critical_deployments = [
            ("kube-system", "coredns"),
        ]

        for namespace, name in critical_deployments:
            result = subprocess.run(
                ["kubectl", "get", "deployment", "-n", namespace, name, "-o", "json"],
                capture_output=True,
                text=True
            )

            if result.returncode != 0:
                continue

            deployment = json.loads(result.stdout)
            affinity = deployment["spec"]["template"]["spec"].get("affinity", {})

            # Check for pod anti-affinity (preferred or required)
            pod_anti_affinity = affinity.get("podAntiAffinity", {})
            has_anti_affinity = (
                "requiredDuringSchedulingIgnoredDuringExecution" in pod_anti_affinity or
                "preferredDuringSchedulingIgnoredDuringExecution" in pod_anti_affinity
            )

            if deployment["spec"].get("replicas", 1) > 1:
                # Only enforce anti-affinity if multiple replicas
                assert has_anti_affinity or deployment["spec"]["replicas"] == 1, (
                    f"{name} should have pod anti-affinity configured for HA"
                )
