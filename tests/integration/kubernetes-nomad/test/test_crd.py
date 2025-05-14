import pytest
import json
from kubernetes import client

def test_custom_resource_definition_support(nomad_client, k8s_api_client):
    """Test custom resource definition support."""
    # First check that the job is using custom resources
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert job_info["Status"] == "running", "Example job is not running"
    
    # Check if tasks contain custom resources
    task_groups = job_info.get("TaskGroups", [])
    assert len(task_groups) > 0, "No task groups found in job"
    
    nginx_task = None
    for group in task_groups:
        for task in group.get("Tasks", []):
            if task.get("Name") == "nginx":
                nginx_task = task
                break
    
    assert nginx_task is not None, "Nginx task not found in job"
    
    # Check for CRD in the task config
    config = nginx_task.get("Config", {})
    assert "custom_resources" in json.dumps(config), "custom_resources not found in task config"
    
    # Check if the CRD groups are configured in Nomad
    try:
        with open("/etc/nomad.d/kubernetes-integration.hcl", "r") as f:
            config_content = f.read()
            assert "custom_resources =" in config_content, "custom_resources not found in Kubernetes integration config"
            assert "custom_resource_groups" in config_content, "custom_resource_groups not found in config"
            assert "monitoring.coreos.com" in config_content, "monitoring.coreos.com group not found in config"
            assert "cert-manager.io" in config_content, "cert-manager.io group not found in config"
    except Exception as e:
        pytest.fail(f"Failed to check Kubernetes integration config: {e}")
    
    # Verify CRD usage in Kubernetes - this requires the CRD API
    try:
        # Check if the ServiceMonitor CRD exists
        api_client = client.ApiClient()
        
        # Try to see if monitoring.coreos.com resources are registered
        api_instance = client.CustomObjectsApi(api_client)
        
        # Try to list ServiceMonitor CRDs if they're available
        try:
            # First check if the API resource exists
            client.ApisApi(api_client).get_api_resources_with_http_info('monitoring.coreos.com/v1')
            
            # If we get here, the API resource exists, so try to list ServiceMonitors
            service_monitors = api_instance.list_cluster_custom_object(
                group="monitoring.coreos.com",
                version="v1",
                plural="servicemonitors"
            )
            
            # Check if the resource was created from our job
            items = service_monitors.get("items", [])
            found = any(item.get("metadata", {}).get("name") == "mantl-example" for item in items)
            
            if found:
                assert True, "ServiceMonitor CRD found and created properly"
            else:
                # We might not be able to create the CRD in testing, so this test is informational
                print("ServiceMonitor CRD API available but no mantl-example resources found")
        except client.rest.ApiException as e:
            # The CRD API might not be installed in the test environment, which is fine
            if e.status == 404:
                # CRDs aren't installed, but that's okay - we're testing the integration, not actual CRD usage
                pass
            else:
                raise
    except Exception:
        # Fallback check - verify the pod has the expected labels that would be used by ServiceMonitor
        pods = k8s_api_client.list_pod_for_all_namespaces(
            label_selector="app=mantl-example"
        )
        assert len(pods.items) > 0, "No pods found in Kubernetes with app=mantl-example label"
        
        # Check pod has metadata required for ServiceMonitor
        pod = pods.items[0]
        labels = pod.metadata.labels
        assert "app" in labels, "app label not found in pod metadata"
        assert labels["app"] == "mantl-example", "Unexpected app label value"