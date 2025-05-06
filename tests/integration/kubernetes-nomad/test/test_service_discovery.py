import pytest
import time
import requests

def test_service_discovery(nomad_client, consul_client, k8s_api_client):
    """Test cross-platform service discovery between Kubernetes and Nomad."""
    # Verify the example job is running
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert job_info["Status"] == "running", "Example job is not running"
    
    # Check that the service is registered in Consul
    services = consul_client.catalog.services()[1]
    assert "mantl-example-service" in services, "Service not found in Consul"
    
    # Check that service has the right tags
    service_info = consul_client.catalog.service("mantl-example-service")[1]
    assert len(service_info) > 0, "No service instances found"
    
    service_tags = service_info[0]["ServiceTags"]
    assert "mantl-service=true" in service_tags, "Service doesn't have mantl-service tag"
    
    # Check that the service is available in Kubernetes
    services = k8s_api_client.list_service_for_all_namespaces(
        label_selector="mantl-service=true"
    )
    assert len(services.items) > 0, "No services found in Kubernetes with mantl-service label"
    
    # Check service metadata
    service_meta = consul_client.catalog.service("mantl-example-service")[1][0]["ServiceMeta"]
    assert service_meta.get("version") == "latest", "Service metadata is incorrect"
    assert service_meta.get("service_type") == "web", "Service metadata is incorrect"
    
    # Verify connectivity to the service
    time.sleep(5)  # Allow time for endpoint to become available
    service_addr = service_info[0]["ServiceAddress"]
    service_port = service_info[0]["ServicePort"]
    
    try:
        response = requests.get(f"http://{service_addr}:{service_port}", timeout=5)
        assert response.status_code == 200, "Service is not responding"
    except Exception as e:
        pytest.fail(f"Failed to connect to service: {e}")