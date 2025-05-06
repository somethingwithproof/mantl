import pytest
import requests

def test_multi_region_orchestration(nomad_client):
    """Test multi-region orchestration configuration."""
    # Check server info to see if regions are configured
    try:
        server_info = nomad_client.get("agent/self")
        # Check if the config contains multi-region settings
        config = server_info.get("config", {})
        assert "Region" in config, "Region not found in Nomad configuration"
        region = config.get("Region")
        assert region == "global", f"Unexpected region: {region}"
        
        # Check if server has multi-region config
        server_config = config.get("Server", {})
        assert server_config.get("EnableMultiRegion", False), "Multi-region is not enabled"
    except Exception as e:
        pytest.fail(f"Failed to get server info: {e}")
    
    # Check job config for multi-region settings
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert "Multiregion" in job_info, "Multiregion config not found in job"
    
    multiregion = job_info.get("Multiregion", {})
    assert multiregion.get("Strategy") is not None, "Strategy not found in multiregion config"
    
    regions = multiregion.get("Regions", [])
    assert len(regions) > 0, "No regions found in multiregion config"
    assert any(r.get("Name") == "global" for r in regions), "Global region not found in job config"
    
    # Check actual job deployments
    try:
        regions_response = requests.get(
            f"{nomad_client.base_url}/v1/regions",
            headers={"Content-Type": "application/json"}
        )
        regions_response.raise_for_status()
        regions = regions_response.json()
        assert "global" in regions, "Global region not found in regions list"
    except Exception as e:
        pytest.fail(f"Failed to get regions: {e}")
    
    # Check disaster recovery configuration
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert "DisasterRecovery" in job_info, "Disaster recovery config not found in job"
    
    dr_config = job_info.get("DisasterRecovery", {})
    assert dr_config.get("MaxDisconnect") is not None, "MaxDisconnect not found in DR config"
    assert dr_config.get("AutoRevert") is True, "AutoRevert not enabled in DR config"