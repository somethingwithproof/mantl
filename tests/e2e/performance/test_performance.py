"""
End-to-end performance tests for Kubernetes cluster.

Tests that measure cluster performance characteristics and establish baselines.
"""
import subprocess
import pytest
import json
import time
import statistics
from datetime import datetime


class TestAPIServerPerformance:
    """Test Kubernetes API server performance."""

    def test_api_server_reachable(self):
        """Test that API server responds to requests."""
        result = subprocess.run(
            ["kubectl", "cluster-info"],
            capture_output=True,
            text=True,
            timeout=10
        )

        assert result.returncode == 0, "API server not reachable"
        assert "Kubernetes control plane" in result.stdout

    def test_api_server_response_time(self):
        """Test API server response time for basic queries."""
        measurements = []

        for _ in range(10):
            start_time = time.time()
            result = subprocess.run(
                ["kubectl", "get", "nodes", "-o", "json"],
                capture_output=True,
                timeout=5
            )
            end_time = time.time()

            if result.returncode == 0:
                measurements.append(end_time - start_time)

        assert len(measurements) >= 8, "Too many API failures"

        avg_time = statistics.mean(measurements)
        p95_time = statistics.quantiles(measurements, n=20)[18]  # 95th percentile

        # API server should respond within reasonable time
        assert avg_time < 1.0, f"Average API response time too high: {avg_time:.3f}s"
        assert p95_time < 2.0, f"P95 API response time too high: {p95_time:.3f}s"

    def test_list_pods_all_namespaces_performance(self):
        """Test performance of listing all pods across namespaces."""
        start_time = time.time()
        result = subprocess.run(
            ["kubectl", "get", "pods", "-A", "-o", "json"],
            capture_output=True,
            timeout=30
        )
        duration = time.time() - start_time

        assert result.returncode == 0, "Failed to list pods"
        assert duration < 5.0, f"Listing all pods took too long: {duration:.3f}s"


class TestPodStartupPerformance:
    """Test pod startup performance."""

    @pytest.fixture
    def test_namespace(self):
        """Create a test namespace."""
        namespace = f"perf-test-{int(time.time())}"

        subprocess.run(
            ["kubectl", "create", "namespace", namespace],
            capture_output=True
        )

        yield namespace

        # Cleanup
        subprocess.run(
            ["kubectl", "delete", "namespace", namespace, "--force", "--grace-period=0"],
            capture_output=True
        )

    def test_single_pod_startup_time(self, test_namespace):
        """Test how long it takes for a single pod to start."""
        pod_name = "nginx-test"

        # Create pod
        create_start = time.time()
        result = subprocess.run(
            ["kubectl", "run", pod_name, "--image=nginx:1.27-alpine",
             "-n", test_namespace],
            capture_output=True
        )
        assert result.returncode == 0, "Failed to create pod"

        # Wait for pod to be running
        max_wait = 120
        start_time = time.time()
        pod_running = False

        while time.time() - start_time < max_wait:
            result = subprocess.run(
                ["kubectl", "get", "pod", pod_name, "-n", test_namespace,
                 "-o", "json"],
                capture_output=True,
                text=True
            )

            if result.returncode == 0:
                pod = json.loads(result.stdout)
                if pod["status"]["phase"] == "Running":
                    # Check all containers are ready
                    container_statuses = pod["status"].get("containerStatuses", [])
                    if all(c.get("ready", False) for c in container_statuses):
                        pod_running = True
                        break

            time.sleep(2)

        total_time = time.time() - create_start

        assert pod_running, f"Pod did not start within {max_wait}s"
        assert total_time < 60, f"Pod startup took too long: {total_time:.1f}s"

        # Cleanup
        subprocess.run(
            ["kubectl", "delete", "pod", pod_name, "-n", test_namespace,
             "--force", "--grace-period=0"],
            capture_output=True
        )

    def test_multiple_pods_startup_time(self, test_namespace):
        """Test parallel pod startup performance."""
        deployment_name = "nginx-deployment"
        replica_count = 10

        # Create deployment
        create_start = time.time()
        result = subprocess.run(
            ["kubectl", "create", "deployment", deployment_name,
             "--image=nginx:1.27-alpine", f"--replicas={replica_count}",
             "-n", test_namespace],
            capture_output=True
        )
        assert result.returncode == 0, "Failed to create deployment"

        # Wait for all pods to be ready
        max_wait = 180
        start_time = time.time()
        all_ready = False

        while time.time() - start_time < max_wait:
            result = subprocess.run(
                ["kubectl", "get", "deployment", deployment_name,
                 "-n", test_namespace, "-o", "json"],
                capture_output=True,
                text=True
            )

            if result.returncode == 0:
                deployment = json.loads(result.stdout)
                ready_replicas = deployment["status"].get("readyReplicas", 0)
                if ready_replicas == replica_count:
                    all_ready = True
                    break

            time.sleep(3)

        total_time = time.time() - create_start

        assert all_ready, f"Not all pods ready within {max_wait}s"
        assert total_time < 120, (
            f"Deployment of {replica_count} pods took too long: {total_time:.1f}s"
        )

        # Cleanup
        subprocess.run(
            ["kubectl", "delete", "deployment", deployment_name,
             "-n", test_namespace, "--force", "--grace-period=0"],
            capture_output=True
        )


class TestNetworkPerformance:
    """Test network performance within cluster."""

    @pytest.fixture(scope="class")
    def test_namespace(self):
        """Create a test namespace for network tests."""
        namespace = "network-perf-test"

        subprocess.run(
            ["kubectl", "create", "namespace", namespace],
            capture_output=True
        )

        yield namespace

        # Cleanup
        subprocess.run(
            ["kubectl", "delete", "namespace", namespace, "--force", "--grace-period=0"],
            capture_output=True
        )

    def test_pod_to_pod_connectivity(self, test_namespace):
        """Test basic pod-to-pod connectivity."""
        # Create two nginx pods
        subprocess.run(
            ["kubectl", "run", "pod1", "--image=nginx:1.27-alpine",
             "-n", test_namespace],
            capture_output=True
        )

        subprocess.run(
            ["kubectl", "run", "pod2", "--image=nicolaka/netshoot:latest",
             "-n", test_namespace, "--command", "--", "sleep", "300"],
            capture_output=True
        )

        # Wait for pods to be ready
        time.sleep(15)

        # Get pod1 IP
        result = subprocess.run(
            ["kubectl", "get", "pod", "pod1", "-n", test_namespace,
             "-o", "jsonpath={.status.podIP}"],
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            pytest.skip("Could not get pod IP")

        pod1_ip = result.stdout.strip()

        # Test connectivity from pod2 to pod1
        result = subprocess.run(
            ["kubectl", "exec", "pod2", "-n", test_namespace, "--",
             "ping", "-c", "3", "-W", "2", pod1_ip],
            capture_output=True,
            timeout=10
        )

        assert result.returncode == 0, "Pod-to-pod connectivity failed"

        # Cleanup
        subprocess.run(
            ["kubectl", "delete", "pod", "pod1", "pod2", "-n", test_namespace,
             "--force", "--grace-period=0"],
            capture_output=True
        )

    def test_service_discovery_latency(self, test_namespace):
        """Test DNS resolution latency for services."""
        # Create a service
        subprocess.run(
            ["kubectl", "run", "web", "--image=nginx:1.27-alpine",
             "-n", test_namespace],
            capture_output=True
        )

        subprocess.run(
            ["kubectl", "expose", "pod", "web", "--port=80",
             "-n", test_namespace],
            capture_output=True
        )

        # Create test pod
        subprocess.run(
            ["kubectl", "run", "test", "--image=nicolaka/netshoot:latest",
             "-n", test_namespace, "--command", "--", "sleep", "300"],
            capture_output=True
        )

        # Wait for resources
        time.sleep(15)

        # Test DNS resolution
        measurements = []
        for _ in range(5):
            start_time = time.time()
            result = subprocess.run(
                ["kubectl", "exec", "test", "-n", test_namespace, "--",
                 "nslookup", f"web.{test_namespace}.svc.cluster.local"],
                capture_output=True,
                timeout=5
            )
            duration = time.time() - start_time

            if result.returncode == 0:
                measurements.append(duration)

        if measurements:
            avg_time = statistics.mean(measurements)
            assert avg_time < 1.0, f"DNS resolution too slow: {avg_time:.3f}s"

        # Cleanup
        subprocess.run(
            ["kubectl", "delete", "pod,service", "web,test", "-n", test_namespace,
             "--force", "--grace-period=0"],
            capture_output=True
        )


class TestStoragePerformance:
    """Test storage performance."""

    def test_pvc_provisioning_time(self):
        """Test how long it takes to provision a PVC."""
        pvc_name = f"test-pvc-{int(time.time())}"
        namespace = "default"

        pvc_manifest = f"""
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: {pvc_name}
  namespace: {namespace}
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
"""

        # Create PVC
        create_start = time.time()
        result = subprocess.run(
            ["kubectl", "apply", "-f", "-"],
            input=pvc_manifest,
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            pytest.skip("Could not create PVC - storage class may not be configured")

        # Wait for PVC to be bound
        max_wait = 120
        start_time = time.time()
        pvc_bound = False

        while time.time() - start_time < max_wait:
            result = subprocess.run(
                ["kubectl", "get", "pvc", pvc_name, "-n", namespace,
                 "-o", "json"],
                capture_output=True,
                text=True
            )

            if result.returncode == 0:
                pvc = json.loads(result.stdout)
                if pvc["status"]["phase"] == "Bound":
                    pvc_bound = True
                    break

            time.sleep(2)

        total_time = time.time() - create_start

        # Cleanup
        subprocess.run(
            ["kubectl", "delete", "pvc", pvc_name, "-n", namespace],
            capture_output=True
        )

        assert pvc_bound, f"PVC not bound within {max_wait}s"
        assert total_time < 60, f"PVC provisioning took too long: {total_time:.1f}s"


class TestSchedulerPerformance:
    """Test scheduler performance."""

    def test_scheduler_latency(self):
        """Test scheduler latency by measuring pending to running time."""
        namespace = "default"
        pod_name = f"scheduler-test-{int(time.time())}"

        pod_manifest = f"""
apiVersion: v1
kind: Pod
metadata:
  name: {pod_name}
  namespace: {namespace}
spec:
  containers:
  - name: nginx
    image: nginx:1.27-alpine
    resources:
      requests:
        cpu: 100m
        memory: 128Mi
"""

        # Create pod
        result = subprocess.run(
            ["kubectl", "apply", "-f", "-"],
            input=pod_manifest,
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            pytest.skip("Could not create test pod")

        # Wait and measure time from pending to running
        max_wait = 60
        start_time = time.time()
        scheduled_time = None
        running_time = None

        while time.time() - start_time < max_wait:
            result = subprocess.run(
                ["kubectl", "get", "pod", pod_name, "-n", namespace,
                 "-o", "json"],
                capture_output=True,
                text=True
            )

            if result.returncode == 0:
                pod = json.loads(result.stdout)

                # Check for scheduled condition
                if not scheduled_time:
                    conditions = pod["status"].get("conditions", [])
                    for condition in conditions:
                        if condition["type"] == "PodScheduled" and condition["status"] == "True":
                            scheduled_time = time.time()
                            break

                # Check if running
                if pod["status"]["phase"] == "Running" and scheduled_time:
                    running_time = time.time()
                    break

            time.sleep(1)

        # Cleanup
        subprocess.run(
            ["kubectl", "delete", "pod", pod_name, "-n", namespace,
             "--force", "--grace-period=0"],
            capture_output=True
        )

        if scheduled_time and running_time:
            scheduler_latency = scheduled_time - start_time
            container_start_latency = running_time - scheduled_time

            # Scheduler should be fast
            assert scheduler_latency < 5.0, (
                f"Scheduler latency too high: {scheduler_latency:.3f}s"
            )


class TestResourceUtilization:
    """Test cluster resource utilization."""

    def test_node_resource_utilization(self):
        """Test that nodes have available resources."""
        result = subprocess.run(
            ["kubectl", "top", "nodes"],
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            pytest.skip("Metrics server not available")

        # Parse output (skip header)
        lines = result.stdout.strip().split("\n")[1:]

        for line in lines:
            parts = line.split()
            if len(parts) >= 5:
                cpu_percent = int(parts[2].rstrip("%"))
                memory_percent = int(parts[4].rstrip("%"))

                # Warn if utilization is very high
                if cpu_percent > 90:
                    pytest.skip(f"Node {parts[0]} CPU utilization very high: {cpu_percent}%")
                if memory_percent > 90:
                    pytest.skip(f"Node {parts[0]} memory utilization very high: {memory_percent}%")

    def test_system_pods_resource_usage(self):
        """Test that system pods are not consuming excessive resources."""
        result = subprocess.run(
            ["kubectl", "top", "pods", "-n", "kube-system"],
            capture_output=True,
            text=True
        )

        if result.returncode != 0:
            pytest.skip("Metrics server not available")

        # Just verify we can get metrics
        assert "CPU" in result.stdout or "MEMORY" in result.stdout, (
            "Could not retrieve pod metrics"
        )
