import time

import pytest
import requests


def test_opa_policy_enforcement(nomad_client):
    """Test unified policy enforcement with OPA."""
    # First check that the OPA task is running
    job_info = nomad_client.get("job/mantl-kubernetes-example")
    assert job_info["Status"] == "running", "Example job is not running"

    # Check for OPA task in the job
    task_groups = job_info.get("TaskGroups", [])
    assert len(task_groups) > 0, "No task groups found in job"

    # Find the policy-agent task
    policy_agent_task = None
    for group in task_groups:
        for task in group.get("Tasks", []):
            if task.get("Name") == "policy-agent":
                policy_agent_task = task
                break

    assert policy_agent_task is not None, "Policy agent task not found in job"

    # Verify the OPA task configuration
    config = policy_agent_task.get("Config", {})
    assert config.get("image", "").startswith("openpolicyagent/opa"), "Unexpected OPA image"

    # Check allocations to make sure OPA is running
    allocations = nomad_client.get("job/mantl-kubernetes-example/allocations")
    assert len(allocations) > 0, "No allocations found for job"

    alloc_id = allocations[0]["ID"]
    alloc_info = nomad_client.get(f"allocation/{alloc_id}")

    # Check if the policy-agent task is running
    task_states = alloc_info.get("TaskStates", {})
    assert "policy-agent" in task_states, "Policy agent task not found in allocation"
    assert task_states["policy-agent"]["State"] == "running", "Policy agent task not running"

    # Check if Rego policy file was created
    time.sleep(2)
    try:
        # Use allocation file system API to check for policy file
        resp = requests.get(
            f"{nomad_client.base_url}/v1/client/fs/stat/alloc/{alloc_id}/local/policies/kubernetes.rego",
            headers={"Content-Type": "application/json"}
        )
        assert resp.status_code == 200, "Rego policy file not found"
    except Exception:
        # Alternative check - look for the template that creates the policy
        for task in task_states:
            if task == "policy-agent":
                # Check task templates
                templates = policy_agent_task.get("Templates", [])
                assert any("package kubernetes.admission" in template.get("EmbeddedTmpl", "")
                           for template in templates), "Kubernetes admission policy not found in templates"

    # Check OPA configuration in Nomad
    try:
        with open("/etc/nomad.d/kubernetes-integration.hcl", "r") as f:
            config_content = f.read()
            assert "policy {" in config_content, "policy block not found in Kubernetes integration config"
            assert "opa_url" in config_content, "opa_url not found in policy config"
            assert "evaluation_paths" in config_content, "evaluation_paths not found in policy config"
            assert "kubernetes/admission" in config_content, "kubernetes/admission path not found in policy config"
    except Exception as e:
        pytest.fail(f"Failed to check OPA configuration: {e}")
