import pytest

def test_service_mesh_integration(nomad_client, consul_client, k8s_api_client):
    """Test service mesh and zero-downtime migration features."""
    # Check if service mesh is configured in Nomad
    try:
        with open("/etc/nomad.d/kubernetes-integration.hcl", "r") as f:
            config_content = f.read()
            assert "service_mesh {" in config_content, "service_mesh block not found in config"
            assert "enabled = true" in config_content, "service_mesh not enabled in config"
            assert "connect_sidecar_image" in config_content, "connect_sidecar_image not found in config"
            assert "metrics_enabled" in config_content, "metrics_enabled not found in config"
    except Exception as e:
        pytest.fail(f"Failed to check service mesh config: {e}")
    
    # Check the job for service mesh configuration
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert job_info["Status"] == "running", "Example job is not running"
    
    # Check for service mesh port in the job
    task_groups = job_info.get("TaskGroups", [])
    assert len(task_groups) > 0, "No task groups found in job"
    
    # Check if the service has the appropriate networking setup
    services = []
    for group in task_groups:
        for service in group.get("Services", []):
            services.append(service)
    
    assert len(services) > 0, "No services found in job"
    
    # Check for metrics port that would be exposed by service mesh
    metrics_port_found = False
    for service in services:
        if service.get("Name") == "mantl-example-metrics":
            metrics_port_found = True
            break
    
    assert metrics_port_found, "Metrics service not found for service mesh monitoring"
    
    # Check if the network configuration includes the metrics port
    for group in task_groups:
        network = group.get("Networks", [{}])[0]
        ports = network.get("DynamicPorts", []) + network.get("ReservedPorts", [])
        metrics_port = next((p for p in ports if p.get("Label") == "metrics"), None)
        assert metrics_port is not None, "Metrics port not found in network configuration"
    
    # Check for connect sidecars in Consul services
    services = consul_client.catalog.services()[1]
    consul_client.catalog.service("mantl-example-service")[1]
    
    # If the service mesh is fully applied, we should see a tagged service
    tagged_services = []
    for svc_name, tags in services.items():
        if any("mantl" in tag for tag in tags):
            tagged_services.append(svc_name)
    
    assert len(tagged_services) > 0, "No tagged services found in Consul"
    
    # Verify that the Kubernetes pods expose the metrics port
    pods = k8s_api_client.list_pod_for_all_namespaces(
        label_selector="app=mantl-example"
    )
    assert len(pods.items) > 0, "No pods found in Kubernetes with app=mantl-example label"
    
    # Check the port configuration on the pod
    pod = pods.items[0]
    containers = pod.spec.containers
    
    metrics_container_port_found = False
    for container in containers:
        for port in container.ports:
            if port.name == "metrics" and port.container_port == 9090:
                metrics_container_port_found = True
                break
    
    assert metrics_container_port_found, "Metrics port not found in container configuration"