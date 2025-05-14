
def test_gpu_support(nomad_client, k8s_api_client):
    """Test GPU workload scheduling support."""
    # Verify the Nomad job has GPU requirements
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert job_info["Status"] == "running", "Example job is not running"

    # Check for GPU resources in the job spec
    task_groups = job_info.get("TaskGroups", [])
    assert len(task_groups) > 0, "No task groups found in job"

    nginx_task = None
    for group in task_groups:
        for task in group.get("Tasks", []):
            if task.get("Name") == "nginx":
                nginx_task = task
                break

    assert nginx_task is not None, "Nginx task not found in job"

    # Check for GPU resource in task
    resources = nginx_task.get("Resources", {})
    devices = resources.get("Devices", [])
    assert len(devices) > 0, "No devices found in task resources"

    gpu_device = None
    for device in devices:
        if "nvidia/gpu" in device.get("Name", ""):
            gpu_device = device
            break

    assert gpu_device is not None, "No GPU device found in task resources"
    assert gpu_device.get("Count") > 0, "GPU count is not greater than 0"

    # Check for GPU resources in the Kubernetes pod spec
    config = nginx_task.get("Config", {})
    pod_spec = config.get("pod_spec", "")
    assert "nvidia.com/gpu" in pod_spec, "GPU resource not found in Kubernetes pod spec"

    # Check for GPU resources in Kubernetes
    pods = k8s_api_client.list_pod_for_all_namespaces(
        label_selector="app=mantl-example"
    )
    assert len(pods.items) > 0, "No pods found in Kubernetes with app=mantl-example label"

    # Check pod resource limits for GPU
    pod = pods.items[0]
    containers = pod.spec.containers
    assert len(containers) > 0, "No containers found in pod"

    container = containers[0]
    limits = container.resources.limits
    assert "nvidia.com/gpu" in limits, "GPU limit not found in container resources"
    assert limits["nvidia.com/gpu"] == "1", "Unexpected GPU limit value"
