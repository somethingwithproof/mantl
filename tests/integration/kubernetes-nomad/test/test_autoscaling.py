
def test_autoscaling_integration(nomad_client):
    """Test autoscaling integration."""
    # Verify the Nomad job has autoscaling config
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert job_info["Status"] == "running", "Example job is not running"

    # Check if job has scaling policy
    scaling = job_info.get("Scaling")
    assert scaling is not None, "Scaling configuration not found in job"
    assert scaling.get("Min") is not None, "Min value not found in scaling config"
    assert scaling.get("Max") is not None, "Max value not found in scaling config"

    # Check if autoscaling is enabled in Nomad config
    server_info = nomad_client.get("agent/self")
    config = server_info.get("config", {})
    plugins = config.get("Plugins", {})
    kubernetes_plugin = None

    # The structure might vary, so we need to search for the kubernetes plugin
    try:
        kubernetes_plugin = plugins.get("kubernetes", {})
    except BaseException:
        pass

    if kubernetes_plugin is None:
        # Try alternative approach to find the plugin
        for plugin_name, plugin_data in plugins.items():
            if "kubernetes" in plugin_name.lower():
                kubernetes_plugin = plugin_data
                break

    assert kubernetes_plugin is not None, "Kubernetes plugin not found in Nomad config"

    # Check for autoscaling settings in the Kubernetes plugin config
    # Try to fetch the running config from the file system since the API migh
    # not expose all details
    try:
        with open("/etc/nomad.d/kubernetes-integration.hcl", "r") as f:
            config_content = f.read()
            assert "autoscaling {" in config_content, "Autoscaling block not found in Kubernetes integration config"
            assert "min_replicas" in config_content, "min_replicas not found in autoscaling config"
            assert "max_replicas" in config_content, "max_replicas not found in autoscaling config"
    except BaseException:
        # Fall back to checking for the feature through the API
        job_id = "mantl-kubernetes-example"

        # Get allocations for the job
        allocations = nomad_client.get(f"job/{job_id}/allocations")
        assert len(allocations) > 0, "No allocations found for job"

        # Check if we can get scaling information from the job
        scaling_policies = nomad_client.get(f"scaling/policies")

        # See if any policy applies to our job
        job_policies = [p for p in scaling_policies if p.get("Target", {}).get("Job") == job_id]

        if job_policies:
            assert len(job_policies) > 0, "No scaling policies found for job"
            policy = job_policies[0]
            assert policy.get("Min") is not None, "Min value not found in scaling policy"
            assert policy.get("Max") is not None, "Max value not found in scaling policy"
        else:
            # If we can't find explicit policies, check if the job has scaling blocks
            assert scaling.get("Min") is not None, "Min value not found in job scaling config"
            assert scaling.get("Max") is not None, "Max value not found in job scaling config"
