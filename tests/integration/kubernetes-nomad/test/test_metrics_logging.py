import time

import pytest
import requests


def test_metrics_integration(nomad_client, k8s_api_client):
    """Test federated metrics and logging integration."""
    # Check if the metrics are exposed in Nomad
    try:
        metrics_response = requests.get(f"{nomad_client.base_url}/v1/metrics")
        metrics_response.raise_for_status()
        metrics = metrics_response.json()

        # Check for expected Nomad metrics
        assert "nomad" in metrics, "Nomad metrics not found"
        assert "runtime" in metrics, "Runtime metrics not found"
        assert metrics["Gauges"] is not None, "Gauge metrics not found"

        # Check for Prometheus-formatted metrics
        prometheus_response = requests.get(f"{nomad_client.base_url}/v1/metrics?format=prometheus")
        prometheus_response.raise_for_status()
        prometheus_metrics = prometheus_response.tex

        assert "nomad_client_allocated_cpu" in prometheus_metrics, "CPU metrics not found"
        assert "nomad_client_allocated_memory" in prometheus_metrics, "Memory metrics not found"
    except Exception as e:
        pytest.fail(f"Failed to get metrics from Nomad: {e}")

    # Check that the metrics service is running
    services = k8s_api_client.list_service_for_all_namespaces(
        label_selector="app=mantl-example"
    )
    assert len(services.items) > 0, "No services found in Kubernetes with app=mantl-example label"

    # Check for annotations related to Prometheus metrics
    pods = k8s_api_client.list_pod_for_all_namespaces(
        label_selector="app=mantl-example"
    )
    assert len(pods.items) > 0, "No pods found in Kubernetes with app=mantl-example label"

    # Check pod annotations for Prometheus scraping
    pod = pods.items[0]
    annotations = pod.metadata.annotations if pod.metadata.annotations else {}
    assert "prometheus.io/scrape" in annotations, "Prometheus scrape annotation not found"
    assert annotations.get("prometheus.io/scrape") == "true", "Prometheus scrape not enabled"
    assert "prometheus.io/port" in annotations, "Prometheus port annotation not found"

    # Verify that the logging sidecar is running
    allocations = nomad_client.get("job/mantl-kubernetes-example/allocations")
    assert len(allocations) > 0, "No allocations found for example job"
    alloc_id = allocations[0]["ID"]

    alloc_details = nomad_client.get(f"allocation/{alloc_id}")
    tasks = alloc_details.get("TaskStates", {})
    assert "logging-sidecar" in tasks, "Logging sidecar task not found"
    assert tasks["logging-sidecar"]["State"] == "running", "Logging sidecar not running"

    # Check that the fluentd config was created
    time.sleep(2)
    try:
        fs_response = requests.get(
            f"{nomad_client.base_url}/v1/client/fs/stat/alloc/{alloc_id}/local/fluentd.conf",
            headers={"Content-Type": "application/json"}
        )
        assert fs_response.status_code == 200, "Fluentd config file not found"
    except Exception as e:
        pytest.fail(f"Failed to verify fluentd config: {e}")
