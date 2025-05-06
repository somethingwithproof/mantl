import pytest
import requests
import json
import time

def test_vault_secrets_integration(nomad_client, vault_client):
    """Test Vault integration for secrets management."""
    # Check if Vault is healthy
    assert vault_client.sys.read_health_status(method='GET')['initialized'], "Vault is not initialized"
    
    # Check if the example secret exists
    secret_path = "kv/data/mantl/kubernetes-example"
    secret = vault_client.secrets.kv.v2.read_secret_version(path="mantl/kubernetes-example", mount_point="kv")
    assert secret['data']['data']['api_key'] == "test-api-key", "API key secret is incorrect"
    assert secret['data']['data']['db_password'] == "test-db-password", "DB password secret is incorrect"
    
    # Check if Nomad has access to Vault
    nomad_policies = vault_client.sys.list_policies()['policies']
    assert "nomad-server" in nomad_policies, "Nomad server policy not found in Vault"
    
    # Get allocation ID for the job
    allocations = nomad_client.get("job/mantl-kubernetes-example/allocations")
    assert len(allocations) > 0, "No allocations found for example job"
    alloc_id = allocations[0]["ID"]
    
    # Check if the template was rendered correctly
    time.sleep(5)  # Wait for template to be rendered
    try:
        fs_response = requests.get(
            f"{nomad_client.base_url}/v1/client/fs/cat/alloc/{alloc_id}/secrets/credentials.env",
            headers={"Content-Type": "application/json"}
        )
        fs_response.raise_for_status()
        content = fs_response.text
        assert "EXAMPLE_API_KEY=test-api-key" in content, "API key not found in rendered template"
        assert "EXAMPLE_DB_PASSWORD=test-db-password" in content, "DB password not found in rendered template"
    except Exception as e:
        pytest.fail(f"Failed to verify template contents: {e}")
        
    # Create a new secret to verify Nomad has write access
    try:
        vault_client.secrets.kv.v2.create_or_update_secret(
            path="mantl/nomad-test",
            secret={"test-key": "test-value"},
            mount_point="kv"
        )
        # Verify the secret was created
        secret = vault_client.secrets.kv.v2.read_secret_version(path="mantl/nomad-test", mount_point="kv")
        assert secret['data']['data']['test-key'] == "test-value", "Failed to create test secret"
    except Exception as e:
        pytest.fail(f"Failed to create secret in Vault: {e}")