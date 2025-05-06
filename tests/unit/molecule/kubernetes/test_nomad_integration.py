import os
import pytest
import testinfra.utils.ansible_runner

testinfra_hosts = testinfra.utils.ansible_runner.AnsibleRunner(
    os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('all')


@pytest.mark.parametrize('path', [
    '/etc/kubernetes',
    '/var/lib/kubernetes',
    '/var/log/kubernetes',
])
def test_kubernetes_directories(host, path):
    """Test that Kubernetes directories exist."""
    assert host.file(path).exists
    assert host.file(path).is_directory


@pytest.mark.parametrize('path', [
    '/etc/kubernetes/nomad-integration',
    '/var/lib/kubernetes/nomad-integration',
])
def test_integration_directories(host, path):
    """Test that Kubernetes-Nomad integration directories exist."""
    assert host.file(path).exists
    assert host.file(path).is_directory


def test_integration_files(host):
    """Test that integration files exist on control nodes."""
    # Skip test if not a control node
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    files = [
        '/etc/kubernetes/nomad-integration/k8s-api-server',
        '/etc/kubernetes/nomad-integration/k8s-ca.crt',
        '/etc/kubernetes/nomad-integration/nomad-token',
        '/etc/nomad.d/kubernetes-integration.hcl',
        '/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad'
    ]
    
    for file_path in files:
        assert host.file(file_path).exists
        assert host.file(file_path).is_file


def test_nomad_k8s_config_content(host):
    """Test the content of the Nomad Kubernetes integration configuration."""
    # Skip test if not a control node
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.is_file
    assert config_file.contains('plugin "kubernetes"')
    assert config_file.contains('host =')
    assert config_file.contains('service_account_token =')
    assert config_file.contains('ca_file =')


def test_example_job_content(host):
    """Test the content of the example Nomad job for Kubernetes."""
    # Skip test if not a control node
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.is_file
    assert job_file.contains('job "mantl-kubernetes-example"')
    assert job_file.contains('driver = "kubernetes"')
    assert job_file.contains('image =')


def test_rbac_files(host):
    """Test that RBAC files exist."""
    # Skip test if not a control node
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    rbac_files = [
        '/etc/kubernetes/nomad-integration/nomad-sa.yaml',
        '/etc/kubernetes/nomad-integration/nomad-clusterrole.yaml',
        '/etc/kubernetes/nomad-integration/nomad-clusterrolebinding.yaml'
    ]
    
    for file_path in rbac_files:
        assert host.file(file_path).exists
        assert host.file(file_path).is_file


def test_common_criteria_settings(host):
    """Test Common Criteria settings in the integration."""
    # Skip test if not a control node or if Common Criteria not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('nomad_common_criteria_enabled', False):
        pytest.skip("Common Criteria not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('audit {')
    assert config_file.contains('enabled = true')
    assert config_file.contains('type = "file"')
    assert config_file.contains('path = "/var/log/nomad/kubernetes-audit.json"')


def test_service_discovery(host):
    """Test cross-platform service discovery configuration."""
    # Skip test if not a control node or if service discovery not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('kubernetes_nomad_service_discovery_enabled', True):
        pytest.skip("Service discovery not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('service_discovery "kubernetes"')
    assert config_file.contains('server_address =')
    assert config_file.contains('token =')
    assert config_file.contains('label_selector =')
    
    # Also test in example job
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.contains('mantl-service=true')
    assert job_file.contains('Cross-platform Service Discovery')


def test_vault_integration(host):
    """Test Vault integration for secrets management."""
    # Skip test if not a control node or if Vault integration not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('kubernetes_nomad_vault_integration_enabled', False):
        pytest.skip("Vault integration not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('vault {')
    assert config_file.contains('enabled = true')
    assert config_file.contains('address =')
    assert config_file.contains('token =')
    assert config_file.contains('kubernetes_auth {')
    
    # Also test in example job
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.contains('volumeMounts:')
    assert job_file.contains('name: vault-token')
    assert job_file.contains('template {')
    assert job_file.contains('with secret')
    assert job_file.contains('Shared Vault Secrets Management')


def test_federated_metrics(host):
    """Test federated metrics and logging configuration."""
    # Skip test if not a control node or if metrics not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('kubernetes_nomad_metrics_enabled', False):
        pytest.skip("Metrics not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('telemetry {')
    assert config_file.contains('prometheus_metrics = true')
    assert config_file.contains('publish_allocation_metrics = true')
    assert config_file.contains('collection_interval =')
    
    # Also test in example job
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.contains('task "logging-sidecar"')
    assert job_file.contains('image = "{{ kubernetes_nomad_fluentd_image')
    assert job_file.contains('prometheus.yml')
    assert job_file.contains('Federated Metrics and Logging')


def test_multi_region(host):
    """Test multi-region orchestration configuration."""
    # Skip test if not a control node or if multi-region not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('kubernetes_nomad_multi_region_enabled', False):
        pytest.skip("Multi-region not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('multi_region {')
    assert config_file.contains('enabled = true')
    assert config_file.contains('regions =')
    assert config_file.contains('strategy =')
    
    # Also test in example job
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.contains('multiregion {')
    assert job_file.contains('Multi-region Orchestration')


def test_gpu_scheduling(host):
    """Test GPU workload scheduling configuration."""
    # Skip test if not a control node or if GPU not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('kubernetes_nomad_gpu_enabled', False):
        pytest.skip("GPU scheduling not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('gpu_support = true')
    assert config_file.contains('gpu_vendor =')
    
    # Also test in example job
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.contains('device "{{ kubernetes_nomad_gpu_vendor')
    assert job_file.contains('GPU Workload Scheduling')


def test_autoscaling(host):
    """Test autoscaling integration configuration."""
    # Skip test if not a control node or if autoscaling not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('kubernetes_nomad_autoscaling_enabled', False):
        pytest.skip("Autoscaling not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('autoscaling {')
    assert config_file.contains('enabled = true')
    assert config_file.contains('min_replicas =')
    assert config_file.contains('max_replicas =')
    
    # Also test in example job
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.contains('scaling {')
    assert job_file.contains('min     =')
    assert job_file.contains('max     =')
    assert job_file.contains('Autoscaling Integration')


def test_crd_support(host):
    """Test custom resource definition support configuration."""
    # Skip test if not a control node or if CRD not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('kubernetes_nomad_crd_enabled', False):
        pytest.skip("CRD support not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('custom_resources =')
    assert config_file.contains('custom_resource_groups =')
    
    # Also test in example job
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.contains('custom_resources = [')
    assert job_file.contains('apiVersion = "monitoring.coreos.com/v1"')
    assert job_file.contains('Custom Resource Definition Support')


def test_disaster_recovery(host):
    """Test disaster recovery configuration."""
    # Skip test if not a control node or if disaster recovery not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('kubernetes_nomad_disaster_recovery_enabled', False):
        pytest.skip("Disaster recovery not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('disaster_recovery {')
    assert config_file.contains('enabled = true')
    assert config_file.contains('recovery_threshold =')
    assert config_file.contains('snapshot_path =')
    
    # Also test in example job
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.contains('disaster_recovery {')
    assert job_file.contains('Disaster Recovery with Automatic Failover')


def test_opa_enforcement(host):
    """Test unified policy enforcement with OPA configuration."""
    # Skip test if not a control node or if OPA not enabled
    if 'control' not in host.ansible.get_variables().get('group_names', []):
        pytest.skip("Not a control node")
    
    if not host.ansible.get_variables().get('kubernetes_nomad_opa_enabled', False):
        pytest.skip("OPA policy enforcement not enabled")
    
    config_file = host.file('/etc/nomad.d/kubernetes-integration.hcl')
    assert config_file.exists
    assert config_file.contains('policy {')
    assert config_file.contains('enabled = true')
    assert config_file.contains('opa_url =')
    assert config_file.contains('evaluation_paths =')
    
    # Also test in example job
    job_file = host.file('/var/lib/kubernetes/nomad-integration/kubernetes-example.nomad')
    assert job_file.exists
    assert job_file.contains('task "policy-agent"')
    assert job_file.contains('openpolicyagent/opa')
    assert job_file.contains('package kubernetes.admission')
    assert job_file.contains('Unified Policy Enforcement with OPA')