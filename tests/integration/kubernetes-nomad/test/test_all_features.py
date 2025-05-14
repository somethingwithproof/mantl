import pytest
import time
import requests

def test_all_features_integration(nomad_client, k8s_api_client):
    """Test that all features are properly integrated together."""
    # Check that the example job is running
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert job_info["Status"] == "running", "Example job is not running"
    
    # Get all allocations for the job
    allocations = nomad_client.get("job/mantl-kubernetes-example/allocations")
    assert len(allocations) > 0, "No allocations found for job"
    
    alloc_id = allocations[0]["ID"]
    alloc_info = nomad_client.get(f"allocation/{alloc_id}")
    
    # Check that all tasks are running successfully
    task_states = alloc_info.get("TaskStates", {})
    required_tasks = ["nginx", "logging-sidecar", "policy-agent"]
    
    for task_name in required_tasks:
        assert task_name in task_states, f"{task_name} task not found in allocation"
        assert task_states[task_name]["State"] == "running", f"{task_name} task not running"
    
    # Verify the HTML page contains references to all features
    time.sleep(2)
    try:
        # Use allocation file system API to check the HTML content
        resp = requests.get(
            f"{nomad_client.base_url}/v1/client/fs/cat/alloc/{alloc_id}/local/index.html",
            headers={"Content-Type": "application/json"}
        )
        html_content = resp.text
        
        # Check for references to all features in the HTML
        features = [
            "Cross-platform Service Discovery",
            "Shared Vault Secrets Management",
            "Federated Metrics and Logging",
            "Multi-region Orchestration",
            "GPU Workload Scheduling",
            "Autoscaling Integration",
            "Custom Resource Definition Support",
            "Disaster Recovery with Automatic Failover",
            "Zero-downtime Migration Tools",
            "Unified Policy Enforcement with OPA"
        ]
        
        for feature in features:
            assert feature in html_content, f"Feature '{feature}' not found in HTML content"
    except Exception as e:
        pytest.fail(f"Failed to verify HTML content: {e}")
    
    # Verify the integration between features by checking dependencies
    
    # Check that Kubernetes resources were created from Nomad
    pods = k8s_api_client.list_pod_for_all_namespaces(
        label_selector="managed-by=nomad"
    )
    assert len(pods.items) > 0, "No Nomad-managed pods found in Kubernetes"
    
    # Check that pods have the necessary annotations for all features
    pod = pods.items[0]
    annotations = pod.metadata.annotations if pod.metadata.annotations else {}
    
    # Should have Prometheus annotations for metrics integration
    assert "prometheus.io/scrape" in annotations, "Prometheus annotation missing - metrics integration issue"
    
    # Check labels for OPA policy enforcement
    labels = pod.metadata.labels
    assert "managed-by" in labels, "managed-by label missing - OPA policy integration issue"
    assert "mantl-service" in labels, "mantl-service label missing - service discovery integration issue"
    
    # Check resource limits for GPU support
    containers = pod.spec.containers
    assert len(containers) > 0, "No containers found in pod"
    
    container = containers[0]
    resources = container.resources
    limits = resources.limits if resources else {}
    
    # Kubernetes resources should match what Nomad requested
    assert "cpu" in limits, "CPU limit missing - resource management issue"
    assert "memory" in limits, "Memory limit missing - resource management issue"
    
    # If using GPUs, should have GPU limit
    if "nvidia.com/gpu" in limits:
        assert limits["nvidia.com/gpu"] == "1", "Incorrect GPU limit - GPU integration issue"
    
    # Verify security context settings from OPA integration
    security_context = container.security_context
    assert security_context is not None, "Security context missing - OPA policy integration issue"
    assert security_context.run_as_non_root == True, "run_as_non_root should be True - OPA policy integration issue"
    assert security_context.allow_privilege_escalation == False, "allow_privilege_escalation should be False - OPA policy integration issue"