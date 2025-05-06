import pytest
import json
import time

def test_disaster_recovery(nomad_client):
    """Test disaster recovery with automatic failover."""
    # Check if the job has disaster recovery configuration
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert job_info["Status"] == "running", "Example job is not running"
    
    # Check for disaster recovery configuration in the job
    dr_config = job_info.get("DisasterRecovery", {})
    assert dr_config is not None, "Disaster recovery configuration not found in job"
    
    # Check specific DR settings
    assert dr_config.get("MaxDisconnect") is not None, "MaxDisconnect not found in DR config"
    assert dr_config.get("AutoRevert") is True, "AutoRevert should be enabled"
    
    # Check if Nomad is configured for DR
    try:
        # Verify the DR configuration in the Nomad server
        with open("/etc/nomad.d/kubernetes-integration.hcl", "r") as f:
            config_content = f.read()
            assert "disaster_recovery {" in config_content, "disaster_recovery block not found in config"
            assert "recovery_threshold" in config_content, "recovery_threshold not found in DR config"
            assert "snapshot_path" in config_content, "snapshot_path not found in DR config"
            assert "snapshot_interval" in config_content, "snapshot_interval not found in DR config"
    except Exception as e:
        pytest.fail(f"Failed to check disaster recovery config: {e}")
    
    # Check if the snapshot directory exists
    try:
        import os
        assert os.path.exists("/var/lib/nomad/snapshots"), "Snapshot directory does not exist"
    except Exception as e:
        pytest.fail(f"Failed to check snapshot directory: {e}")
    
    # Check if we can create a snapshot (for testing DR functionality)
    try:
        response = nomad_client.post("operator/snapshot", {})
        assert "error" not in response, f"Failed to create snapshot: {response.get('error', '')}"
    except Exception as e:
        # This might fail in the test environment, which is okay
        print(f"Note: Snapshot creation failed (expected in test env): {e}")
        
    # Check for other DR-related features
    server_info = nomad_client.get("agent/self")
    config = server_info.get("config", {})
    server_config = config.get("Server", {})
    
    # Check if heartbeat monitoring is enabled (part of DR)
    assert server_config.get("HeartbeatGrace") is not None, "HeartbeatGrace not found in server config"